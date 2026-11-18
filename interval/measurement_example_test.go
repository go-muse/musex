package interval_test

import (
	"fmt"
	"iter"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleChromatic() {
	sys := system.Traditional()
	from := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), centsOf(0), nil)
	to := note.MustFromPitch(pitch.MustFromASPN(spelling.EFlat(), 4), centsOf(0), nil)
	shift := interval.MustChromaticBetweenNotesIn(sys, from, to)
	fmt.Println(shift, shift.Steps())
	d4 := note.MustFromPitch(pitch.MustFromASPN(spelling.D(), 4), centsOf(0), nil)
	moved := shift.MustTransposeNote(d4, system.Flats())
	fmt.Println(moved.Spelling())
	// Output:
	// +3 3
	// F
}

func ExampleGeneric_Resolve() {
	shift := interval.MustNewChromatic(system.Traditional(), 3)
	fmt.Println(interval.MustNewGeneric(2).MustResolve(shift, spelling.C(), interval.Up))
	fmt.Println(interval.MustNewGeneric(3).MustResolve(shift, spelling.C(), interval.Up))
	// Output:
	// A2
	// m3
}

func TestMeasurementIteratorsConcurrent(t *testing.T) {
	t.Parallel()
	sys := trad()
	g := interval.MustNewGeneric(3)
	p := interval.Deviations(ai(-1), alt(-1, 2), ai(0))
	abstract, err := g.Variants(interval.Up, p)
	require.NoError(t, err)
	pitches, err := g.FromPitch(qt24(), wp(qt24(), spelling.C(), 4), interval.Up, p)
	require.NoError(t, err)
	seq, err := interval.ForShift(interval.MustNewChromatic(sys, 6), spelling.C(), interval.Up, interval.MustThrough(2))
	require.NoError(t, err)
	abstractResults := make([][]interval.Interval, 16)
	pitchResults := make([][]interval.PitchVariant, 16)
	shiftResults := make([][]interval.Interval, 16)
	var wg sync.WaitGroup
	for worker := range abstractResults {
		wg.Go(func() {
			abstractResults[worker] = slices.Collect(abstract)
			pitchResults[worker] = slices.Collect(pitches)
			shiftResults[worker] = slices.Collect(seq)
		})
	}
	wg.Wait()
	for worker := range abstractResults {
		assert.Equal(t, ivs("m3", "n3", "M3"), abstractResults[worker])
		require.Len(t, pitchResults[worker], 3)
		for index, v := range pitchResults[worker] {
			assert.True(t, ivs("m3", "n3", "M3")[index].Equal(v.Interval))
		}
		assert.Equal(t, ivs("AA3", "A4", "d5", "dd6"), shiftResults[worker])
	}
	// Separate iterators also remain independent when traversals are interleaved.
	next, stop := iter.Pull(seq)
	defer stop()
	first, ok := next()
	require.True(t, ok)
	assert.True(t, iv("AA3").Equal(first))
	assert.Equal(t, ivs("AA3", "A4", "d5", "dd6"), slices.Collect(seq))
	second, ok := next()
	require.True(t, ok)
	assert.True(t, iv("A4").Equal(second))
}
