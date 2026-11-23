package sonance_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/problemsolving/sonance"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDistanceOverflow covers the out-of-range path at the edges of int.
func TestDistanceOverflow(t *testing.T) {
	t.Parallel()

	ns := []note.Note{
		note.MustFromASPN(spelling.C(), math.MaxInt, cents.Zero(), nil),
		note.MustFromASPN(spelling.C(), math.MinInt, cents.Zero(), nil),
	}
	_, err := sonance.ScoreNotes(system.Traditional(), ns, sonance.HuronPairProfile())
	assert.ErrorIs(t, err, sonance.ErrOutOfRange)
	assert.ErrorIs(t, err, system.ErrOutOfRange)
	assert.Panics(t, func() { sonance.MustScoreNotes(system.Traditional(), ns, sonance.HuronPairProfile()) })
}

// edgeOctave is the ASPN octave of the C whose distance from G sharp 4 is
// exactly math.MinInt steps and from G4 math.MinInt+1 steps in the
// traditional system: 12 * (edgeOctave - 4) - 8 == math.MinInt.
const edgeOctave = 4 - 768614336404564650

// panicValue runs f and returns the value it panics with, or nil.
func panicValue(f func()) (v any) {
	defer func() { v = recover() }()
	f()
	return nil
}

// TestDistanceEdges covers the distances math.MinInt, math.MinInt+1 and
// math.MaxInt between two notes. Exactly math.MinInt steps, in either
// listing order, is out of range: an error matching ErrOutOfRange and
// system.ErrOutOfRange, never a runtime panic; MustScoreNotes panics with
// that error. The neighbouring distances fit and score the one pair C-G,
// a fifth or fourth, at Huron's weight 1.240.
func TestDistanceEdges(t *testing.T) {
	t.Parallel()

	low := note.MustFromASPN(spelling.C(), edgeOctave, cents.Zero(), nil)
	gSharp := note.MustFromASPN(spelling.GSharp(), 4, cents.Zero(), nil)
	g := note.MustFromASPN(spelling.G(), 4, cents.Zero(), nil)

	cases := []struct {
		name       string
		notes      []note.Note
		outOfRange bool
	}{
		{"distance math.MinInt", []note.Note{gSharp, low}, true},
		{"distance -math.MinInt", []note.Note{low, gSharp}, true},
		{"distance math.MinInt+1", []note.Note{g, low}, false},
		{"distance math.MaxInt", []note.Note{low, g}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var (
				s   sonance.Score
				err error
			)
			if !assert.NotPanics(t, func() {
				s, err = sonance.ScoreNotes(system.Traditional(), tc.notes, sonance.HuronPairProfile())
			}) {
				return
			}
			if !tc.outOfRange {
				require.NoError(t, err)
				assert.InDelta(t, 1.240, s.Value(), 1e-9)
				return
			}
			assert.ErrorIs(t, err, sonance.ErrOutOfRange)
			assert.ErrorIs(t, err, system.ErrOutOfRange)
			v := panicValue(func() {
				sonance.MustScoreNotes(system.Traditional(), tc.notes, sonance.HuronPairProfile())
			})
			perr, ok := v.(error)
			if assert.True(t, ok, "MustScoreNotes panics with %v, want an error", v) {
				assert.ErrorIs(t, perr, sonance.ErrOutOfRange)
				assert.ErrorIs(t, perr, system.ErrOutOfRange)
			}
		})
	}
}
