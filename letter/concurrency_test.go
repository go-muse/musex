package letter_test

import (
	"fmt"
	"math"
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/letter"

	"github.com/stretchr/testify/assert"
)

// K-13: all operations called concurrently from many goroutines give the same
// results as K-1...K-6. Data races are detected when the test runs under
// go test -race, as in CI. Workers report mismatches through their own slot;
// assertions run only in the test goroutine.
func TestConcurrentUse(t *testing.T) {
	t.Parallel()

	const iterations = 20

	workers := runtime.GOMAXPROCS(0) * 4
	failures := make([][]string, workers)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			<-start
			for range iterations {
				failures[w] = append(failures[w], checkOperations()...)
			}
		})
	}
	close(start)
	wg.Wait()

	for w, f := range failures {
		assert.Empty(t, f, "worker %d", w)
	}
}

// checkOperations calls every public operation for all letters and pairs,
// compares the results with the oracle and changes the returned slices.
// It returns a description of every mismatch.
func checkOperations() []string {
	var failures []string
	fail := func(format string, args ...any) {
		failures = append(failures, fmt.Sprintf(format, args...))
	}

	want := oracle()
	names := oracleNames()

	all := letter.All()
	if !slices.Equal(want[:], all) {
		fail("All() = %v", all)
	}
	clear(all)

	for i, a := range want {
		if !a.IsValid() {
			fail("%s.IsValid() = false", names[i])
		}
		if got := a.Ordinal(); got != i {
			fail("%s.Ordinal() = %d", names[i], got)
		}
		if got := a.String(); got != names[i] {
			fail("%s.String() = %q", names[i], got)
		}
		if got, err := letter.FromOrdinal(i); err != nil || got != a {
			fail("FromOrdinal(%d) = %v, %v", i, got, err)
		}
		if got := letter.MustFromOrdinal(i); got != a {
			fail("MustFromOrdinal(%d) = %v", i, got)
		}

		cycle := a.Cycle()
		for k := range oracleCount {
			if len(cycle) != oracleCount || cycle[k] != expectedShift(i, k) {
				fail("%s.Cycle() = %v", names[i], cycle)

				break
			}
		}
		for k := range cycle {
			cycle[k] = letter.Letter(0)
		}

		for n := -15; n <= 15; n++ {
			if got := a.Shift(n); got != expectedShift(i, n) {
				fail("%s.Shift(%d) = %v", names[i], n, got)
			}
		}
		if got := a.Shift(math.MaxInt); got != expectedShift(i, math.MaxInt) {
			fail("%s.Shift(MaxInt) = %v", names[i], got)
		}
		if got := a.Shift(math.MinInt); got != expectedShift(i, math.MinInt) {
			fail("%s.Shift(MinInt) = %v", names[i], got)
		}

		for j, b := range want {
			if got := a.DistanceUp(b); got != mod7(j-i) {
				fail("%s.DistanceUp(%s) = %d", names[i], names[j], got)
			}
			if got := a.DistanceDown(b); got != mod7(i-j) {
				fail("%s.DistanceDown(%s) = %d", names[i], names[j], got)
			}
			if got := a.Equal(b); got != (i == j) {
				fail("%s.Equal(%s) = %t", names[i], names[j], got)
			}
		}
	}

	return failures
}
