package naming_test

import (
	"fmt"
	"iter"
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
)

// sharedSequence is a catalog that the workers of TestCatalogAllConcurrentUse
// use at the same time, a sequence taken from it before the workers start and
// the expected pairs. Nothing writes to them after they are created.
type sharedSequence struct {
	name    string
	catalog *naming.Catalog
	seq     iter.Seq2[int, string]
	want    []seqPair
}

// K-26 (NF-2): many goroutines traverse the same sequence, stop it early and
// call All on the same catalog at the same time; all results agree with the
// expected pairs. Data races are detected under go test -race, as in make
// test and CI. Workers record mismatches and panics in their own slot; the
// assertions run only in the test goroutine after all workers are joined.
func TestCatalogAllConcurrentUse(t *testing.T) {
	t.Parallel()

	const iterations = 20

	newShared := func(name string, c *naming.Catalog, want []seqPair) sharedSequence {
		return sharedSequence{name: name, catalog: c, seq: c.All(), want: want}
	}

	shared := []sharedSequence{
		newShared("Russian", naming.Russian(), sortedPairs(russianOracle())),
		newShared("English", naming.English(), sortedPairs(englishOracle())),
		newShared("German", naming.German(), sortedPairs(germanOracle())),
		newShared("custom", naming.MustNew(customPairs()), sortedPairs(customPairs())),
		newShared("extreme numbers", naming.MustNew(extremePairs()), sortedPairs(extremePairs())),
		newShared("nil", nil, []seqPair{}),
	}

	workers := runtime.GOMAXPROCS(0) * 4
	failures := make([][]string, workers)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			// A panic in a worker would stop the whole test binary; record it
			// as a failure of this worker instead.
			defer func() {
				if r := recover(); r != nil {
					failures[w] = append(failures[w], fmt.Sprintf("panic: %v", r))
				}
			}()

			<-start
			for range iterations {
				failures[w] = append(failures[w], checkConcurrentSequences(shared)...)
			}
		})
	}
	close(start)
	wg.Wait()

	for w, f := range failures {
		assert.Empty(t, f, "worker %d", w)
	}
}

// checkConcurrentSequences traverses the shared sequences fully and with an
// early stop, calls All on the shared catalogs and returns a description of
// every mismatch with the expected pairs.
func checkConcurrentSequences(shared []sharedSequence) []string {
	var failures []string
	fail := func(format string, args ...any) {
		failures = append(failures, fmt.Sprintf(format, args...))
	}

	for _, s := range shared {
		if got := collectPairs(s.seq); !slices.Equal(got, s.want) {
			fail("%s: shared sequence yields %v", s.name, got)
		}

		if got := collectPairs(s.catalog.All()); !slices.Equal(got, s.want) {
			fail("%s: new All yields %v", s.name, got)
		}

		var prefix []seqPair
		for aspn, name := range s.seq {
			prefix = append(prefix, seqPair{aspn: aspn, name: name})
			if len(prefix) == 2 {
				break
			}
		}
		if want := s.want[:min(2, len(s.want))]; !slices.Equal(prefix, want) {
			fail("%s: stopped shared sequence yields %v", s.name, prefix)
		}
	}

	return failures
}

// KR-NM16: a goroutine assigns to *c while the test goroutine and readers
// traverse a sequence taken from c before the goroutine started. Each
// traversal yields the pairs c held when All was called, and there is no data
// race under go test -race: the traversal does not read *c. The catalog
// object belongs to this test only. All goroutines are joined before the
// assertions, which run only in the test goroutine.
func TestCatalogAllConcurrentRebind(t *testing.T) {
	t.Parallel()

	const (
		traversals = 50
		readers    = 4
	)

	c := naming.Russian()
	seq := c.All()
	want := sortedPairs(russianOracle())

	start := make(chan struct{})
	var wg sync.WaitGroup

	wg.Go(func() {
		<-start
		for i := range traversals {
			if i%2 == 0 {
				*c = *naming.English()
			} else {
				*c = naming.Catalog{}
			}
		}
		*c = *naming.German()
	})

	readerResults := make([][][]seqPair, readers)
	for r := range readers {
		wg.Go(func() {
			<-start
			for range traversals {
				readerResults[r] = append(readerResults[r], collectPairs(seq))
			}
		})
	}

	close(start)
	own := make([][]seqPair, 0, traversals)
	for range traversals {
		own = append(own, collectPairs(seq))
	}
	wg.Wait()

	for i, got := range own {
		assert.Equal(t, want, got, "traversal %d of the test goroutine", i)
	}
	for r, results := range readerResults {
		for i, got := range results {
			assert.Equal(t, want, got, "traversal %d of reader %d", i, r)
		}
	}

	assert.Equal(t, sortedPairs(germanOracle()), collectPairs(c.All()), "new All after the assignments")
	assert.Equal(t, want, collectPairs(seq), "sequence after the assignments")
}
