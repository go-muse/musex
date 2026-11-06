package cents_test

import (
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/require"
)

// Sinks keep the results of the measured calls; concrete types, not
// interfaces: storing a struct in an interface would allocate by itself.
var (
	sinkBool  bool
	sinkCents cents.Cents
)

// TestEqualDoesNotAllocate checks that Equal, == and copying do not allocate
// (architecture "Type Cents", "Size and memory"; N-4).
//
// The test does not call t.Parallel: testing.AllocsPerRun counts the
// allocations of the whole process, so it runs in the sequential phase of
// the package tests.
func TestEqualDoesNotAllocate(t *testing.T) {
	set := setS(t)

	for _, a := range set {
		for _, b := range set {
			x, y := a.v, b.v

			allocs := testing.AllocsPerRun(100, func() { sinkBool = x.Equal(y) })
			require.Zero(t, allocs, "%s.Equal(%s)", a.name, b.name)

			allocs = testing.AllocsPerRun(100, func() { sinkBool = x == y })
			require.Zero(t, allocs, "%s == %s", a.name, b.name)
		}

		x := a.v
		allocs := testing.AllocsPerRun(100, func() {
			c := x
			sinkCents = c
		})
		require.Zero(t, allocs, "copy of %s", a.name)
	}
}
