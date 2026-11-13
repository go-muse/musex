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

// spreadNames is the number of names of the catalogs CD and CS of the NF-7
// check.
const spreadNames = 4096

// allAllocationLimit is the most allocations that a call of All with a full
// traversal may make: the sequence function of the package and, without
// inlining, the loop body function and its state in the caller.
const allAllocationLimit = 3

// allBytesTolerance is how many more bytes a call of All with a full
// traversal may allocate for the catalogs CE, CD and CS than for C1. A copy
// of the pairs on every call would take at least 32 KiB for CD and CS.
const allBytesTolerance = 1 << 10

// sinkNumber and sinkName keep the pairs yielded in the allocation tests and
// the benchmark observable, so the traversal cannot be removed as dead code.
var (
	sinkNumber int
	sinkName   string
)

// traverseAll is the single measured function of the NF-7 check: a call of
// All and a full traversal that stores every pair in package variables.
func traverseAll(c *naming.Catalog) {
	for aspn, name := range c.All() {
		sinkNumber, sinkName = aspn, name
	}
}

// denseNames returns n names "d0"... for the ASPN numbers 0...n-1.
func denseNames(n int) map[int]string {
	pairs := make(map[int]string, n)
	for k := range n {
		pairs[k] = "d" + strconv.Itoa(k)
	}

	return pairs
}

// spreadOverInt returns n >= 2 names "s0"... for numbers spread in ascending
// order over the whole range of int (int is 64-bit), from math.MinInt to
// math.MaxInt inclusive.
func spreadOverInt(n int) map[int]string {
	step := uint64(math.MaxUint64) / uint64(n-1)
	pairs := make(map[int]string, n)
	for k := range n - 1 {
		// Flipping the sign bit maps the ascending uint64 values onto
		// ascending int values: 0 becomes math.MinInt.
		aspn := int(int64(uint64(k)*step) ^ math.MinInt)
		pairs[aspn] = "s" + strconv.Itoa(k)
	}
	pairs[math.MaxInt] = "s" + strconv.Itoa(n-1)

	return pairs
}

// allMeasuredCatalog is a catalog of the NF-7 check with its number of named
// octaves.
type allMeasuredCatalog struct {
	name    string
	catalog *naming.Catalog
	names   int
}

// allMeasuredCatalogs builds the catalogs of the NF-7 check before any
// measurement: C1 with one name, CE with one name and emptyPairs pairs with
// the empty name, CD with spreadNames names 0...4095, CS with spreadNames
// names spread over the whole int range, and the Russian catalog.
func allMeasuredCatalogs(t *testing.T) []allMeasuredCatalog {
	t.Helper()

	catalogs := []allMeasuredCatalog{
		{name: "C1", catalog: naming.MustNew(withEmptyNames(0)), names: 1},
		{name: "CE", catalog: naming.MustNew(withEmptyNames(emptyPairs)), names: 1},
		{name: "CD", catalog: naming.MustNew(denseNames(spreadNames)), names: spreadNames},
		{name: "CS", catalog: naming.MustNew(spreadOverInt(spreadNames)), names: spreadNames},
		{name: "Russian", catalog: naming.Russian(), names: 11},
	}

	for _, m := range catalogs {
		pairs := collectPairs(m.catalog.All())
		require.Len(t, pairs, m.names, "%s: pairs yielded before the measurement", m.name)
	}

	cs := collectPairs(catalogs[3].catalog.All())
	require.Equal(t, math.MinInt, cs[0].aspn, "CS: first number")
	require.Equal(t, math.MaxInt, cs[len(cs)-1].aspn, "CS: last number")

	return catalogs
}

// allAllocatedBytes calls traverseAll(c) the given number of times and
// returns the smallest number of bytes allocated by one call
// (runtime.MemStats TotalAlloc). The minimum removes rare background
// allocations of the runtime.
func allAllocatedBytes(c *naming.Catalog, repeats int) uint64 {
	var before, after runtime.MemStats
	least := uint64(math.MaxUint64)
	for range repeats {
		runtime.ReadMemStats(&before)
		traverseAll(c)
		runtime.ReadMemStats(&after)

		least = min(least, after.TotalAlloc-before.TotalAlloc)
	}

	return least
}

// K-27, NF-7 check (a): a call of All with a full traversal makes the same
// number of allocations for all catalogs - one name, one name with many
// empty pairs, 4096 dense names, 4096 names spread over the whole int range
// and the Russian catalog - and at most allAllocationLimit. The cost of All
// thus depends neither on the number of names nor on the spread of the
// numbers nor on the empty pairs given to New.
//
// Like TestNewAllocationIgnoresEmptyNames, this test reads the allocation
// counter of the whole process and is sequential: it does not call
// t.Parallel and has no parallel subtests. The isolation holds while every
// other test of this package calls t.Parallel as its first action and no
// sequential test leaves goroutines running. All catalogs are built before
// the measurement.
func TestCatalogAllAllocations(t *testing.T) {
	const runs = 100

	catalogs := allMeasuredCatalogs(t)

	allocs := make([]float64, len(catalogs))
	for i, m := range catalogs {
		allocs[i] = testing.AllocsPerRun(runs, func() { traverseAll(m.catalog) })
	}

	for i, m := range catalogs {
		assert.LessOrEqualf(t, allocs[i], float64(allAllocationLimit),
			"%s: allocations per call of All with a full traversal", m.name)
		assert.Equalf(t, allocs[0], allocs[i],
			"%s (%d names) and %s (%d names): allocations per call of All with a full traversal",
			m.name, m.names, catalogs[0].name, catalogs[0].names)
	}
}

// K-27, NF-7 check (b): a call of All with a full traversal allocates at most
// allBytesTolerance bytes more for CE, CD and CS than for C1. The bytes of one
// call are the smallest runtime.MemStats TotalAlloc difference of several
// repeats, as in TestNewAllocationIgnoresEmptyNames. This test is sequential
// for the same reason as TestCatalogAllAllocations.
func TestCatalogAllAllocatedBytes(t *testing.T) {
	const repeats = 5

	catalogs := allMeasuredCatalogs(t)

	bytes := make([]uint64, len(catalogs))
	for i, m := range catalogs {
		bytes[i] = allAllocatedBytes(m.catalog, repeats)
	}

	base := catalogs[0]
	require.Equal(t, "C1", base.name)

	for i, m := range catalogs {
		if m.name != "CE" && m.name != "CD" && m.name != "CS" {
			continue
		}

		extra := int64(bytes[i]) - int64(bytes[0])
		assert.LessOrEqualf(t, extra, int64(allBytesTolerance),
			"All with a full traversal allocates %d bytes for %s (%d names) and %d bytes for %s: %d bytes more",
			bytes[i], m.name, m.names, bytes[0], base.name, extra)
	}
}

// BenchmarkCatalogAll measures a call of All with a full traversal for the
// catalogs of the NF-7 check. It is for reference only, not a pass/fail
// criterion: go test without -bench only compiles it.
func BenchmarkCatalogAll(b *testing.B) {
	cases := []struct {
		name  string
		pairs map[int]string
	}{
		{name: "C1", pairs: withEmptyNames(0)},
		{name: "CE", pairs: withEmptyNames(emptyPairs)},
		{name: "CD", pairs: denseNames(spreadNames)},
		{name: "CS", pairs: spreadOverInt(spreadNames)},
		{name: "Russian", pairs: russianOracle()},
	}

	for _, bc := range cases {
		b.Run(bc.name, func(b *testing.B) {
			c := naming.MustNew(bc.pairs)

			b.ReportAllocs()
			for b.Loop() {
				traverseAll(c)
			}
		})
	}
}
