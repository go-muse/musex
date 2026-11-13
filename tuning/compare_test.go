package tuning_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSoundsSameRootChangeAndSymmetry(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	rootA := std.MustWithRoot(tuning.RootFromClass(pitchclass.MustNew(9, 12)))
	assert.False(t, std.Equal(rootA), "different roots are different tuning values")
	assert.True(t, std.MustSoundsSame(rootA), "K-17 and K-28")
	assert.True(t, rootA.MustSoundsSame(std), "G-10 symmetry")

	for _, tc := range []struct {
		name string
		a    *tuning.Tuning
		b    *tuning.Tuning
		want bool
	}{
		{name: "K-27 octave re-expression", a: std, b: std.MustWithReferenceNote(aspn(spelling.A(), 5)), want: true},
		{name: "K-15 C5 re-expression is strict false", a: std, b: std.MustWithReferenceNote(aspn(spelling.C(), 5)), want: false},
		{name: "K-18 altered reference re-expression", a: std, b: std.MustWithReferenceNote(aspn(spelling.BFlat(), 4)), want: false},
		{name: "next float frequency is different", a: std, b: std.MustWithReferenceFrequency(math.Nextafter(440, 441)), want: false},
		{name: "different system", a: std, b: std.MustWithSystem(traditionalSystemWithAlterations(-1, 1, "limited")), want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.a.SoundsSame(tc.b)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			reverse, err := tc.b.SoundsSame(tc.a)
			require.NoError(t, err)
			assert.Equal(t, got, reverse)
		})
	}
}

func TestSoundsSamePythagoreanReexpression(t *testing.T) {
	t.Parallel()
	original := newPythagoreanTuning(t, tuning.RootFromClass(pitchclass.MustNew(0, 12)))
	changed := original.MustWithReferenceNote(aspn(spelling.C(), 5))
	assert.False(t, original.MustSoundsSame(changed), "K-13")
	assert.False(t, changed.MustSoundsSame(original))
}

func TestSoundsSameUnset(t *testing.T) {
	t.Parallel()
	var unset *tuning.Tuning
	for _, tc := range []struct {
		name string
		a    *tuning.Tuning
		b    *tuning.Tuning
	}{
		{name: "nil receiver", a: unset, b: tuning.Standard()},
		{name: "nil argument", a: tuning.Standard(), b: unset},
		{name: "both nil", a: unset, b: unset},
		{name: "zero receiver", a: &tuning.Tuning{}, b: tuning.Standard()},
		{name: "zero argument", a: tuning.Standard(), b: &tuning.Tuning{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := tc.a.SoundsSame(tc.b)
			requireTuningError(t, err, tuning.ErrUnset)
			requireNoDependencyError(t, err)
		})
	}
}

func TestSoundsSameUnrepresentableInOtherTuning(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	root := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	for _, tc := range []struct {
		name  string
		other *tuning.Tuning
	}{
		{
			name:  "reference step does not fit int",
			other: newTestTuningAt(t, system.Traditional(), temperament.Equal12(), root, aspn(spelling.A(), math.MaxInt), 440),
		},
		{
			name:  "height in other tuning exceeds cents range",
			other: newTestTuningAt(t, system.Traditional(), temperament.Equal12(), root, aspn(spelling.A(), 8304), 440),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := std.SoundsSame(tc.other)
			require.NoError(t, err)
			assert.False(t, got, "G-10: unrepresentable values in B mean false, not an error")
		})
	}
}

func TestSoundsSameNonNormalFrequencyInOtherTuning(t *testing.T) {
	t.Parallel()
	root := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	other := newTestTuningAt(t, system.Traditional(), temperament.Equal12(), root, aspn(spelling.A(), 3000), cents.MaxFrequency)
	const step system.Step = -35952

	height, err := other.CentsFromReferenceOfStep(step)
	require.NoError(t, err)
	requireCentsEqual(t, cents.MustFromInt(-3_595_200), height)
	_, err = other.FrequencyOfStep(step)
	requireTuningError(t, err, tuning.ErrOutOfRange)
	assert.ErrorIs(t, err, cents.ErrOutOfRange)

	got, err := tuning.Standard().SoundsSame(other)
	require.NoError(t, err)
	assert.False(t, got, "G-10: non-normal frequency in B means false, not an error")
}

func TestSoundsSameWindowDirection(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		freq float64
	}{
		{name: "downward finds an ordinary mismatch before later complexity", freq: 440},
		{name: "upward reaches complexity on the first step", freq: 0.5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			complexTune := complexTuning(t).MustWithReferenceFrequency(tc.freq)
			simple := newTestTuningAt(t, traditionalSystemWithClasses(3), temperament.MustEDO(3), tuning.RootFromClass(pitchclass.MustNew(0, 3)), aspn(spelling.CSharp(), 4), tc.freq)

			got, err := complexTune.SoundsSame(simple)
			if tc.freq < 1 {
				requireTuningError(t, err, tuning.ErrTooComplex)
				assert.ErrorIs(t, err, cents.ErrTooComplex)
				return
			}
			require.NoError(t, err)
			assert.False(t, got)
		})
	}
}

func TestSoundsSameComplexity(t *testing.T) {
	t.Parallel()
	tune := complexTuning(t)
	_, err := tune.SoundsSame(tune)
	requireTuningError(t, err, tuning.ErrTooComplex)
	assert.ErrorIs(t, err, cents.ErrTooComplex)
	assert.NotErrorIs(t, err, temperament.ErrTooComplex)
}
