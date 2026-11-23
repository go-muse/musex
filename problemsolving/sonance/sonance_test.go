package sonance_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/problemsolving/sonance"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func n(name spelling.Spelling, aspn int) note.Note {
	return note.MustFromASPN(name, aspn, cents.Zero(), nil)
}

// TestHuronProfile pins the published weights.
func TestHuronProfile(t *testing.T) {
	t.Parallel()

	p := sonance.HuronProfile()
	require.True(t, p.IsValid())
	want := []float64{0, -1.428, -0.582, 0.594, 0.386, 1.240, -0.453}
	assert.Equal(t, want, p.Weights())
	for ic, w := range want {
		assert.Equal(t, w, p.MustWeight(ic))
	}
}

// TestScoreSet covers the interval-class vector scoring of pitch-class
// sets.
func TestScoreSet(t *testing.T) {
	t.Parallel()

	t.Run("major triad", func(t *testing.T) {
		t.Parallel()
		s := sonance.MustScoreSet(pcset.MustNew(12, 0, 4, 7), sonance.HuronProfile())
		assert.InDelta(t, 0.594+0.386+1.240, s.Value(), 1e-9)
		cs := s.Contributions()
		require.Len(t, cs, 7)
		assert.Equal(t, 0, cs[0].Count())
		assert.Equal(t, 0.0, cs[0].Value())
		assert.Equal(t, 3, cs[3].IntervalClass())
		assert.Equal(t, 1, cs[3].Count())
		assert.Equal(t, 0.594, cs[3].Weight())
		assert.InDelta(t, 0.594, cs[3].Value(), 1e-9)
		assert.Equal(t, 1, cs[4].Count())
		assert.Equal(t, 1, cs[5].Count())
		assert.Equal(t, 0, cs[6].Count())
	})

	t.Run("dissonant inventory", func(t *testing.T) {
		t.Parallel()
		s := sonance.MustScoreSet(pcset.MustNew(12, 0, 1, 6), sonance.HuronProfile())
		assert.InDelta(t, -1.428+1.240-0.453, s.Value(), 1e-9)
		assert.Negative(t, s.Value())
	})

	t.Run("empty set", func(t *testing.T) {
		t.Parallel()
		s := sonance.MustScoreSet(pcset.MustNew(12), sonance.HuronProfile())
		assert.Equal(t, 0.0, s.Value())
		for _, c := range s.Contributions() {
			assert.Equal(t, 0, c.Count())
		}
	})

	t.Run("custom profile", func(t *testing.T) {
		t.Parallel()
		p := sonance.MustNewProfile([]float64{1, 2, 3, 4, 5, 6, 7})
		s := sonance.MustScoreSet(pcset.MustNew(12, 0, 4, 7), p)
		assert.InDelta(t, 4+5+6, s.Value(), 1e-9)
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		_, err := sonance.ScoreSet(pcset.Set{}, sonance.HuronProfile())
		assert.ErrorIs(t, err, sonance.ErrUnset)
		_, err = sonance.ScoreSet(pcset.MustNew(12, 0, 4, 7), sonance.Profile{})
		assert.ErrorIs(t, err, sonance.ErrUnset)
		_, err = sonance.ScoreSet(pcset.MustNew(19, 0, 4, 7), sonance.HuronProfile())
		assert.ErrorIs(t, err, sonance.ErrInvalidArgument)
		assert.Panics(t, func() { sonance.MustScoreSet(pcset.Set{}, sonance.HuronProfile()) })
	})
}

// TestScoreNotes covers actual pitch pairs, including unisons, octaves and
// doubled pitch classes.
func TestScoreNotes(t *testing.T) {
	t.Parallel()

	t.Run("voiced major triad", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)}
		s := sonance.MustScoreNotes(system.Traditional(), ns, sonance.HuronProfile())
		assert.InDelta(t, 0.594+0.386+1.240, s.Value(), 1e-9)
	})

	t.Run("doubled bass counts unison and extra pairs", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)}
		s := sonance.MustScoreNotes(system.Traditional(), ns, sonance.HuronProfile())
		assert.InDelta(t, 2*0.386+2*1.240+0.594, s.Value(), 1e-9)
		cs := s.Contributions()
		assert.Equal(t, 1, cs[0].Count())
		assert.Equal(t, 2, cs[4].Count())
		assert.Equal(t, 2, cs[5].Count())
	})

	t.Run("octave pair is IC0", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.C(), 5)}
		p := sonance.MustNewProfile([]float64{1, 0, 0, 0, 0, 0, 0})
		s := sonance.MustScoreNotes(system.Traditional(), ns, p)
		assert.InDelta(t, 1, s.Value(), 1e-9)
		assert.Equal(t, 1, s.Contributions()[0].Count())
	})

	t.Run("enharmonic tritone", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.GFlat(), 4)}
		s := sonance.MustScoreNotes(system.Traditional(), ns, sonance.HuronProfile())
		assert.InDelta(t, -0.453, s.Value(), 1e-9)
	})

	t.Run("empty and single", func(t *testing.T) {
		t.Parallel()
		for _, ns := range [][]note.Note{nil, {n(spelling.C(), 4)}} {
			s := sonance.MustScoreNotes(system.Traditional(), ns, sonance.HuronProfile())
			assert.Equal(t, 0.0, s.Value())
			require.Len(t, s.Contributions(), 7)
			for _, c := range s.Contributions() {
				assert.Equal(t, 0, c.Count())
			}
		}
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.E(), 4)}
		_, err := sonance.ScoreNotes(nil, ns, sonance.HuronProfile())
		assert.ErrorIs(t, err, sonance.ErrUnset)
		_, err = sonance.ScoreNotes(system.Traditional(), ns, sonance.Profile{})
		assert.ErrorIs(t, err, sonance.ErrUnset)
		_, err = sonance.ScoreNotes(system.Positions19(), ns, sonance.HuronProfile())
		assert.ErrorIs(t, err, sonance.ErrUnsupportedSystem)
		_, err = sonance.ScoreNotes(system.Traditional(), []note.Note{{}}, sonance.HuronProfile())
		assert.ErrorIs(t, err, sonance.ErrUnset)
		_, err = sonance.ScoreNotes(system.Traditional(), []note.Note{note.MustFromSpelling(spelling.C(), cents.Zero(), nil)}, sonance.HuronProfile())
		assert.ErrorIs(t, err, sonance.ErrNoOctave)
		assert.Panics(t, func() { sonance.MustScoreNotes(nil, nil, sonance.HuronProfile()) })
	})
}

// TestProfile covers construction and access errors.
func TestProfile(t *testing.T) {
	t.Parallel()

	_, err := sonance.NewProfile([]float64{1, 2, 3})
	assert.ErrorIs(t, err, sonance.ErrInvalidArgument)
	_, err = sonance.NewProfile([]float64{0, 0, 0, 0, 0, 0, math.NaN()})
	assert.ErrorIs(t, err, sonance.ErrInvalidArgument)
	_, err = sonance.NewProfile([]float64{0, 0, 0, 0, 0, 0, math.Inf(1)})
	assert.ErrorIs(t, err, sonance.ErrInvalidArgument)

	var zero sonance.Profile
	assert.False(t, zero.IsValid())
	assert.Nil(t, zero.Weights())
	_, err = zero.Weight(0)
	assert.ErrorIs(t, err, sonance.ErrUnset)
	_, err = sonance.HuronProfile().Weight(7)
	assert.ErrorIs(t, err, sonance.ErrInvalidArgument)
	_, err = sonance.HuronProfile().Weight(-1)
	assert.ErrorIs(t, err, sonance.ErrInvalidArgument)

	p := sonance.MustNewProfile([]float64{0, 1, 2, 3, 4, 5, 6})
	w := p.Weights()
	w[0] = 99
	assert.Equal(t, 0.0, p.MustWeight(0), "the profile keeps its own copy")
}
