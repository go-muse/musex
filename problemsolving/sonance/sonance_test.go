package sonance_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
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

// restricted admits only the alterations -1, 0 and +1.
func restricted() *system.System {
	return system.MustNew(system.Definition{
		Name:        "restricted",
		Classes:     12,
		Boundary:    letter.C,
		Offsets:     [letter.Count]int{0, 2, 4, 5, 7, 9, 11},
		Chromatic:   1,
		Alterations: system.MustIntegerRange(-1, 1),
	})
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
	assert.True(t, p.Equal(sonance.HuronProfile()))
	assert.False(t, p.Equal(sonance.MustNewProfile([]float64{1, 1, 1, 1, 1, 1, 1})))

	pair := sonance.HuronPairProfile()
	require.True(t, pair.IsValid())
	assert.Equal(t, []float64{0, -1.428, -0.582, 0.594, 0.386, 1.240, -0.453, 1.240, 0.386, 0.594, -0.582, -1.428}, pair.Weights())
	assert.True(t, pair.Equal(sonance.HuronPairProfile()))
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
		assert.Equal(t, 3, cs[3].Interval())
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

// TestScoreNotes covers actual pitch pairs, including unisons, octaves,
// doubled pitch classes and the fourth-versus-fifth distinction.
func TestScoreNotes(t *testing.T) {
	t.Parallel()

	t.Run("voiced major triad", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)}
		s := sonance.MustScoreNotes(system.Traditional(), ns, sonance.HuronPairProfile())
		assert.InDelta(t, 0.594+0.386+1.240, s.Value(), 1e-9)
		require.Len(t, s.Contributions(), 12)
	})

	t.Run("doubled bass counts unison and extra pairs", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)}
		s := sonance.MustScoreNotes(system.Traditional(), ns, sonance.HuronPairProfile())
		assert.InDelta(t, 2*0.386+2*1.240+0.594, s.Value(), 1e-9)
		cs := s.Contributions()
		assert.Equal(t, 1, cs[0].Count())
		assert.Equal(t, 2, cs[4].Count())
		assert.Equal(t, 2, cs[7].Count())
	})

	t.Run("octave pair is distance zero", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.C(), 5)}
		weights := make([]float64, 12)
		weights[0] = 1
		p := sonance.MustNewPairProfile(weights)
		s := sonance.MustScoreNotes(system.Traditional(), ns, p)
		assert.InDelta(t, 1, s.Value(), 1e-9)
		assert.Equal(t, 1, s.Contributions()[0].Count())
	})

	t.Run("a profile distinguishes fourths from fifths", func(t *testing.T) {
		t.Parallel()
		weights := make([]float64, 12)
		weights[5] = 1 // fourth
		weights[7] = 2 // fifth
		p := sonance.MustNewPairProfile(weights)
		fourth := sonance.MustScoreNotes(system.Traditional(), []note.Note{n(spelling.C(), 4), n(spelling.F(), 4)}, p)
		fifth := sonance.MustScoreNotes(system.Traditional(), []note.Note{n(spelling.C(), 4), n(spelling.G(), 4)}, p)
		assert.InDelta(t, 1, fourth.Value(), 1e-9)
		assert.InDelta(t, 2, fifth.Value(), 1e-9)
		assert.Equal(t, 1, fourth.Contributions()[5].Count())
		assert.Equal(t, 1, fifth.Contributions()[7].Count())
		assert.NotEqual(t, fourth.Value(), fifth.Value())
	})

	t.Run("enharmonic tritone", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.GFlat(), 4)}
		s := sonance.MustScoreNotes(system.Traditional(), ns, sonance.HuronPairProfile())
		assert.InDelta(t, -0.453, s.Value(), 1e-9)
	})

	t.Run("empty and single", func(t *testing.T) {
		t.Parallel()
		for _, ns := range [][]note.Note{nil, {n(spelling.C(), 4)}} {
			s := sonance.MustScoreNotes(system.Traditional(), ns, sonance.HuronPairProfile())
			assert.Equal(t, 0.0, s.Value())
			require.Len(t, s.Contributions(), 12)
			for _, c := range s.Contributions() {
				assert.Equal(t, 0, c.Count())
			}
		}
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.E(), 4)}
		_, err := sonance.ScoreNotes(nil, ns, sonance.HuronPairProfile())
		assert.ErrorIs(t, err, sonance.ErrUnset)
		_, err = sonance.ScoreNotes(system.Traditional(), ns, sonance.PairProfile{})
		assert.ErrorIs(t, err, sonance.ErrUnset)
		_, err = sonance.ScoreNotes(system.Positions19(), ns, sonance.HuronPairProfile())
		assert.ErrorIs(t, err, sonance.ErrUnsupportedSystem)
		_, err = sonance.ScoreNotes(system.Traditional(), []note.Note{{}}, sonance.HuronPairProfile())
		assert.ErrorIs(t, err, sonance.ErrUnset)
		_, err = sonance.ScoreNotes(system.Traditional(), []note.Note{note.MustFromSpelling(spelling.C(), cents.Zero(), nil)}, sonance.HuronPairProfile())
		assert.ErrorIs(t, err, sonance.ErrNoOctave)
		_, err = sonance.ScoreNotes(restricted(), []note.Note{n(spelling.CDoubleSharp(), 4)}, sonance.HuronPairProfile())
		assert.ErrorIs(t, err, sonance.ErrInadmissible)
		assert.ErrorIs(t, err, system.ErrInadmissible)
		assert.Panics(t, func() { sonance.MustScoreNotes(nil, nil, sonance.HuronPairProfile()) })
	})
}

// TestProfiles covers construction and access errors of both profiles.
func TestProfiles(t *testing.T) {
	t.Parallel()

	_, err := sonance.NewProfile([]float64{1, 2, 3})
	assert.ErrorIs(t, err, sonance.ErrInvalidArgument)
	_, err = sonance.NewProfile([]float64{0, 0, 0, 0, 0, 0, math.NaN()})
	assert.ErrorIs(t, err, sonance.ErrInvalidArgument)
	_, err = sonance.NewProfile([]float64{0, 0, 0, 0, 0, 0, math.Inf(1)})
	assert.ErrorIs(t, err, sonance.ErrInvalidArgument)
	_, err = sonance.NewPairProfile([]float64{1, 2, 3})
	assert.ErrorIs(t, err, sonance.ErrInvalidArgument)
	_, err = sonance.NewPairProfile(make([]float64, 12))
	require.NoError(t, err)
	bad := make([]float64, 12)
	bad[11] = math.NaN()
	_, err = sonance.NewPairProfile(bad)
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

	var zeroPair sonance.PairProfile
	assert.False(t, zeroPair.IsValid())
	assert.Nil(t, zeroPair.Weights())
	_, err = zeroPair.Weight(0)
	assert.ErrorIs(t, err, sonance.ErrUnset)
	_, err = sonance.HuronPairProfile().Weight(12)
	assert.ErrorIs(t, err, sonance.ErrInvalidArgument)
	_, err = sonance.HuronPairProfile().Weight(-1)
	assert.ErrorIs(t, err, sonance.ErrInvalidArgument)

	p := sonance.MustNewProfile([]float64{0, 1, 2, 3, 4, 5, 6})
	w := p.Weights()
	w[0] = 99
	assert.Equal(t, 0.0, p.MustWeight(0), "the profile keeps its own copy")
	pp := sonance.MustNewPairProfile(make([]float64, 12))
	pw := pp.Weights()
	pw[0] = 99
	assert.Equal(t, 0.0, pp.MustWeight(0), "the pair profile keeps its own copy")
}
