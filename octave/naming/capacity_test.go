package naming_test

import (
	"math"
	"runtime"
	"strconv"
	"testing"

	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// emptyPairs is the number of pairs with the empty name added to the inputs
// below.
const emptyPairs = 1 << 16

// allocationTolerance is how many more bytes New may allocate for an input
// with emptyPairs empty names than for the same input without them.
const allocationTolerance = 64 << 10

// sinkEqual keeps the results of Equal in the allocation test and the
// benchmark observable, so the calls cannot be removed as dead code.
var sinkEqual bool

// withEmptyNames returns the pair {0: "a"} plus n pairs k: "" for k = 1...n.
func withEmptyNames(n int) map[int]string {
	pairs := make(map[int]string, n+1)
	pairs[0] = "a"
	for k := 1; k <= n; k++ {
		pairs[k] = ""
	}

	return pairs
}

// emptyNamesOnly returns n pairs k: "" for k = 1...n.
func emptyNamesOnly(n int) map[int]string {
	pairs := make(map[int]string, n)
	for k := 1; k <= n; k++ {
		pairs[k] = ""
	}

	return pairs
}

// newAllocatedBytes calls New(input) the given number of times and returns
// the smallest number of bytes allocated by one call (runtime.MemStats
// TotalAlloc). The minimum removes rare background allocations of the
// runtime. Each catalog is kept alive until the second reading.
func newAllocatedBytes(t *testing.T, input map[int]string, repeats int) uint64 {
	t.Helper()

	var before, after runtime.MemStats
	least := uint64(math.MaxUint64)
	for range repeats {
		runtime.ReadMemStats(&before)
		c, err := naming.New(input)
		runtime.ReadMemStats(&after)
		runtime.KeepAlive(c)

		require.NoError(t, err)
		require.NotNil(t, c)
		least = min(least, after.TotalAlloc-before.TotalAlloc)
	}

	return least
}

// Pairs with the empty name in the input of New do not increase the capacity
// of the catalog indexes, which is set by the number of non-empty names. This
// is observed through allocation: New for {0: "a"} plus emptyPairs empty pairs
// may allocate at most allocationTolerance bytes more than New for {0: "a"}
// alone. Both inputs hold only the names "a" and "", so the size of the names
// does not affect the difference.
//
// The test reads the allocation counter of the whole process, so it must not
// run alongside other tests of this package. It deliberately does not call
// t.Parallel and has no parallel subtests (an exception to the preference for
// t.Parallel in docs/rules/dev-test.md): the testing package resumes parallel
// top-level tests only after all sequential ones have finished. The isolation
// holds while every other test of this package calls t.Parallel as its first
// action and no sequential test leaves goroutines running; a new sequential
// test with background work would break it. Examples run after the tests.
// Both inputs are built before the measurement.
func TestNewAllocationIgnoresEmptyNames(t *testing.T) {
	const repeats = 5

	small := withEmptyNames(0)
	large := withEmptyNames(emptyPairs)

	smallBytes := newAllocatedBytes(t, small, repeats)
	largeBytes := newAllocatedBytes(t, large, repeats)

	extra := int64(largeBytes) - int64(smallBytes)
	assert.LessOrEqualf(t, extra, int64(allocationTolerance),
		"New allocates %d bytes for {0: \"a\"} plus %d empty pairs and %d bytes for {0: \"a\"}: %d bytes more",
		largeBytes, emptyPairs, smallBytes, extra)
}

// Equal allocates no memory, also for catalogs created from inputs with many
// empty pairs. Like TestNewAllocationIgnoresEmptyNames, this test is
// sequential: testing.AllocsPerRun reads the allocation counter of the whole
// process and sets GOMAXPROCS to 1 while it runs, so no other test of this
// package may run at the same time (the same isolation condition applies).
func TestEqualDoesNotAllocate(t *testing.T) {
	cases := []struct {
		name        string
		left, right *naming.Catalog
		want        bool
	}{
		{name: "one name, empty pairs vs one name", left: naming.MustNew(withEmptyNames(emptyPairs)), right: naming.MustNew(withEmptyNames(0)), want: true},
		{name: "one name vs one name, empty pairs", left: naming.MustNew(withEmptyNames(0)), right: naming.MustNew(withEmptyNames(emptyPairs)), want: true},
		{name: "empty pairs only vs nil", left: naming.MustNew(emptyNamesOnly(emptyPairs)), right: nil, want: true},
		{name: "one name, empty pairs vs Russian", left: naming.MustNew(withEmptyNames(emptyPairs)), right: naming.Russian(), want: false},
		{name: "Russian vs Russian", left: naming.Russian(), right: naming.Russian(), want: true},
		{name: "Russian vs English", left: naming.Russian(), right: naming.English(), want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, tc.left.Equal(tc.right))

			allocs := testing.AllocsPerRun(100, func() {
				sinkEqual = tc.left.Equal(tc.right)
			})
			assert.Zero(t, allocs, "allocations per Equal call")
		})
	}
}

// A catalog created from {0: "a"} plus emptyPairs empty pairs is the catalog
// {0: "a"}; a catalog created from empty pairs only is a catalog without
// names.
func TestNewWithManyEmptyNames(t *testing.T) {
	t.Parallel()

	t.Run("one name and empty pairs", func(t *testing.T) {
		t.Parallel()

		c, err := naming.New(withEmptyNames(emptyPairs))
		require.NoError(t, err)
		require.NotNil(t, c)

		one := naming.MustNew(map[int]string{0: "a"})
		assert.True(t, c.Equal(one))
		assert.True(t, one.Equal(c))

		assert.Equal(t, "a", c.Name(0))
		assert.Empty(t, c.Name(1))
		assert.Empty(t, c.Name(emptyPairs))

		aspn, err := c.ASPN("a")
		require.NoError(t, err)
		assert.Equal(t, 0, aspn)

		aspn, err = c.ASPN("")
		assert.ErrorIs(t, err, naming.ErrUnknownName)
		assert.Equal(t, 0, aspn)
	})

	t.Run("empty pairs only", func(t *testing.T) {
		t.Parallel()

		c, err := naming.New(emptyNamesOnly(emptyPairs))
		require.NoError(t, err)
		require.NotNil(t, c)

		none := naming.MustNew(nil)
		assert.True(t, c.Equal(none))
		assert.True(t, none.Equal(c))
		assert.True(t, c.Equal(nil))
		assert.Empty(t, c.Name(1))

		_, err = c.ASPN("")
		assert.ErrorIs(t, err, naming.ErrUnknownName)
	})
}

// BenchmarkCatalogEqual measures Equal between the catalog {0: "a"} and a
// catalog created from {0: "a"} plus n empty pairs. It is not a pass/fail
// criterion: go test without -bench only compiles it.
func BenchmarkCatalogEqual(b *testing.B) {
	for _, n := range []int{1, 1 << 10, 1 << 16} {
		b.Run("empty="+strconv.Itoa(n), func(b *testing.B) {
			c := naming.MustNew(withEmptyNames(n))
			other := naming.MustNew(map[int]string{0: "a"})
			require.True(b, c.Equal(other))

			b.ReportAllocs()
			for b.Loop() {
				sinkEqual = c.Equal(other)
			}
		})
	}
}
