package naming_test

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
)

// sharedCatalogs are the catalogs that all workers of TestConcurrentUse use at
// the same time. Nothing writes to them or to input after they are created.
type sharedCatalogs struct {
	russian, english, german, custom, none *naming.Catalog
	input                                  map[int]string
}

// K-15: all functions and methods called concurrently from many goroutines on
// shared catalogs give the results of K-1...K-7. Data races are detected when
// the test runs under go test -race, as in make test and CI. Workers record
// mismatches and panics in their own slot; assertions run only in the test
// goroutine after all workers finish. No worker writes to a shared catalog or
// to the shared input map (that would be a race by contract).
func TestConcurrentUse(t *testing.T) {
	t.Parallel()

	const iterations = 20

	input := customPairs()
	shared := sharedCatalogs{
		russian: naming.Russian(),
		english: naming.English(),
		german:  naming.German(),
		custom:  naming.MustNew(input),
		none:    nil,
		input:   input,
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
				failures[w] = append(failures[w], checkConcurrentOperations(shared)...)
			}
		})
	}
	close(start)
	wg.Wait()

	for w, f := range failures {
		assert.Empty(t, f, "worker %d", w)
	}
}

// checkConcurrentOperations calls every public function and method on the
// shared catalogs, compares the results with the oracles and returns a
// description of every mismatch.
func checkConcurrentOperations(shared sharedCatalogs) []string {
	var failures []string
	fail := func(format string, args ...any) {
		failures = append(failures, fmt.Sprintf(format, args...))
	}

	cases := []struct {
		name    string
		catalog *naming.Catalog
		fresh   func() *naming.Catalog
		oracle  map[int]string
	}{
		{name: "Russian", catalog: shared.russian, fresh: naming.Russian, oracle: russianOracle()},
		{name: "English", catalog: shared.english, fresh: naming.English, oracle: englishOracle()},
		{name: "German", catalog: shared.german, fresh: naming.German, oracle: germanOracle()},
		{name: "custom", catalog: shared.custom, fresh: func() *naming.Catalog { return naming.MustNew(shared.input) }, oracle: customPairs()},
		{name: "nil", catalog: shared.none, fresh: func() *naming.Catalog { return nil }, oracle: map[int]string{}},
	}

	for _, tc := range cases {
		c := tc.catalog

		for n := -20; n <= 20; n++ {
			if got := c.Name(n); got != tc.oracle[n] {
				fail("%s.Name(%d) = %q", tc.name, n, got)
			}
		}

		for want, name := range tc.oracle {
			if got, err := c.ASPN(name); err != nil || got != want {
				fail("%s.ASPN(%q) = %d, %v", tc.name, name, got, err)
			}
			if got := c.MustASPN(name); got != want {
				fail("%s.MustASPN(%q) = %d", tc.name, name, got)
			}
		}

		if got, err := c.ASPN("unknown"); got != 0 || !errors.Is(err, naming.ErrUnknownName) {
			fail("%s.ASPN(unknown) = %d, %v", tc.name, got, err)
		}

		fresh := tc.fresh()
		if !c.Equal(fresh) || !fresh.Equal(c) {
			fail("%s: shared and new catalogs are not equal", tc.name)
		}

		for _, other := range cases {
			if want := other.name == tc.name; c.Equal(other.catalog) != want {
				fail("%s.Equal(%s) = %t", tc.name, other.name, !want)
			}
		}
	}

	created, err := naming.New(shared.input)
	if err != nil || created == nil || !created.Equal(shared.custom) {
		fail("New(shared input) = %v, %v", created, err)
	}

	return failures
}
