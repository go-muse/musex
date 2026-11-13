package tuning_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithReferenceNoteOctave(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	changed, err := std.WithReferenceNote(aspn(spelling.A(), 5))
	require.NoError(t, err)
	assert.True(t, changed.Reference().Pitch().Equal(aspn(spelling.A(), 5)))
	assert.Equal(t, 880.0, changed.Reference().Frequency(), "K-27")
	assert.Equal(t, 440.0, std.Reference().Frequency())
	requireCentsEqual(t, cents.Zero(), std.MustLevelDifference(changed))
	assert.Equal(t, 1.0, std.MustLevelRatio(changed))
	assert.True(t, std.MustSoundsSame(changed))
	assert.True(t, changed.MustSoundsSame(std))
}

func TestWithReferenceNoteEqualTempered(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	c5 := aspn(spelling.C(), 5)
	changed := std.MustWithReferenceNote(c5)
	assert.Equal(t, std.MustFrequency(c5), changed.Reference().Frequency())
	requireCentsEqual(t, cents.Zero(), std.MustLevelDifference(changed), "K-15")
	assert.False(t, std.MustSoundsSame(changed), "strict frequencies differ after re-expression")
}

func TestWithReferenceNotePythagorean(t *testing.T) {
	t.Parallel()
	original := newPythagoreanTuning(t, tuning.RootFromClass(pitchclass.MustNew(0, 12)))
	changed := original.MustWithReferenceNote(aspn(spelling.C(), 5))
	expectedReference := original.MustFrequency(aspn(spelling.C(), 5))
	assert.Equal(t, expectedReference, changed.Reference().Frequency())
	assert.Equal(t, "521.4815", strconv.FormatFloat(expectedReference, 'f', 4, 64))

	for _, note := range []pitch.Pitch{
		aspn(spelling.A(), 4), aspn(spelling.C(), 5), aspn(spelling.F(), 5), aspn(spelling.C(), 4),
	} {
		before := original.MustFrequency(note)
		after := changed.MustFrequency(note)
		assert.InDelta(t, before, after, math.Abs(before)*1e-12)
	}

	wantDifference := cents.MustBetween(440, expectedReference).MustSub(cents.MustFromRatio(32, 27))
	requireCentsEqual(t, wantDifference, original.MustLevelDifference(changed), "K-13")
	assert.Equal(t, wantDifference.MustMultiplier(), original.MustLevelRatio(changed))
	assert.False(t, original.MustSoundsSame(changed))
}

func TestLevelDifferencePythagoreanWithEqualTemperedValue(t *testing.T) {
	t.Parallel()
	original := newPythagoreanTuning(t, tuning.RootFromClass(pitchclass.MustNew(0, 12)))
	etC5 := tuning.Standard().MustFrequency(aspn(spelling.C(), 5))
	other := newTestTuningAt(t, system.Traditional(), temperament.PythagoreanEFlatGSharp(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), aspn(spelling.C(), 5), etC5)

	want := cents.MustBetween(440, etC5).MustSub(cents.MustFromRatio(32, 27))
	requireCentsEqual(t, want, original.MustLevelDifference(other), "K-14")
	assert.Equal(t, "5.865003", want.Decimal(6, 0))
	assert.InDelta(t, 1.003394, original.MustLevelRatio(other), 0.0000005)
	assert.False(t, original.MustSoundsSame(other))
}

func TestLevelDifferenceEnharmonicReference(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	bFlat4 := aspn(spelling.BFlat(), 4)
	other := std.MustWithReferenceNote(bFlat4)
	assert.Equal(t, "466.1638", strconv.FormatFloat(other.Reference().Frequency(), 'f', 4, 64))
	requireCentsEqual(t, cents.Zero(), std.MustLevelDifference(other), "K-18")
	assert.False(t, std.MustSoundsSame(other))
}

func TestWithReferenceFrequency(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	changed, err := std.WithReferenceFrequency(435)
	require.NoError(t, err)
	assert.Equal(t, 435.0, changed.MustFrequency(aspn(spelling.A(), 4)))
	assert.Equal(t, 440.0, std.MustFrequency(aspn(spelling.A(), 4)))
	assert.Equal(t, 435.0/440.0, std.MustFrequencyRatio(435))
	assert.Equal(t, 0.988636, math.Round(std.MustFrequencyRatio(435)*1e6)/1e6)
	requireCentsEqual(t, cents.MustBetween(440, 435), std.MustLevelDifference(changed))

	_, err = std.WithReferenceFrequency(0)
	requireTuningError(t, err, tuning.ErrInvalidArgument)
	_, err = std.FrequencyRatio(0)
	requireTuningError(t, err, tuning.ErrInvalidArgument)
}

func TestWithReferenceAndAt(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	ref := tuning.MustNewReference(aspn(spelling.BFlat(), 4), 466.1638)
	changed, err := std.WithReference(ref)
	require.NoError(t, err)
	assert.True(t, ref.Equal(changed.Reference()))
	assert.True(t, std.System().Equal(changed.System()))
	assert.True(t, std.Temperament().Equal(changed.Temperament()))
	assert.Equal(t, std.RootClass(), changed.RootClass())

	changedAt, err := std.WithReferenceAt(aspn(spelling.C(), 5), 523)
	require.NoError(t, err)
	assert.True(t, tuning.C5At523().Equal(changedAt.Reference()))

	_, err = std.WithReference(tuning.Reference{})
	requireTuningError(t, err, tuning.ErrUnset)
	_, err = std.WithReference(tuning.MustNewReference(aspn(spelling.MustParse("C+1/2"), 4), 440))
	requireTuningError(t, err, tuning.ErrInadmissible)
	_, err = std.WithReferenceAt(aspn(spelling.MustParse("C+1/2"), 4), -1)
	requireTuningError(t, err, tuning.ErrInadmissible, "pitch is checked before frequency")
	_, err = std.WithReferenceAt(aspn(spelling.A(), 4), -1)
	requireTuningError(t, err, tuning.ErrInvalidArgument)
}

func TestLevelDifferenceIncomparable(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	for _, tc := range []struct {
		name  string
		other *tuning.Tuning
	}{
		{name: "system", other: std.MustWithSystem(traditionalSystemWithAlterations(-1, 1, "limited"))},
		{name: "temperament", other: std.MustWithTemperament(temperament.PythagoreanEFlatGSharp())},
		{name: "root", other: std.MustWithRoot(tuning.RootFromClass(pitchclass.MustNew(9, 12)))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := std.LevelDifference(tc.other)
			requireTuningError(t, err, tuning.ErrIncomparable)
			requireNoDependencyError(t, err)
			_, err = std.LevelRatio(tc.other)
			requireTuningError(t, err, tuning.ErrIncomparable)
		})
	}
}

func TestLevelRatioThresholds(t *testing.T) {
	t.Parallel()
	root := tuning.RootFromClass(pitchclass.MustNew(0, 12))

	t.Run("positive multiplier overflows", func(t *testing.T) {
		t.Parallel()
		low := newTestTuningAt(t, system.Traditional(), temperament.Equal12(), root, aspn(spelling.A(), 4), cents.MinNormalFrequency)
		target := cents.MustFromInt(1_228_800)
		highFrequency := target.MustApplyTo(cents.MinNormalFrequency)
		high := low.MustWithReferenceFrequency(highFrequency)
		requireCentsEqual(t, target, low.MustLevelDifference(high))
		_, err := low.LevelRatio(high)
		requireTuningError(t, err, tuning.ErrOutOfRange)
		assert.ErrorIs(t, err, cents.ErrOutOfRange)
	})

	t.Run("negative multiplier is subnormal", func(t *testing.T) {
		t.Parallel()
		std := tuning.Standard()
		target := cents.MustFromInt(-1_226_401)
		otherFrequency := target.MustApplyTo(440)
		other := std.MustWithReferenceFrequency(otherFrequency)
		requireCentsEqual(t, target, std.MustLevelDifference(other))
		_, err := std.LevelRatio(other)
		requireTuningError(t, err, tuning.ErrOutOfRange)
		assert.ErrorIs(t, err, cents.ErrOutOfRange)
	})
}

func TestLevelDifferenceWithoutIntermediateHeightRangeCheck(t *testing.T) {
	t.Parallel()
	root := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	a := newTestTuningAt(t, system.Traditional(), temperament.Equal12(), root, aspn(spelling.A(), 4), math.Ldexp(1, -1000))
	b := newTestTuningAt(t, system.Traditional(), temperament.Equal12(), root, aspn(spelling.A(), 8304), math.Ldexp(1, 700))

	d := cents.MustBetween(a.Reference().Frequency(), b.Reference().Frequency())
	want := d.MustSub(cents.MustFromInt(9_000_000)).MustSub(cents.MustFromInt(960_000))
	got, err := a.LevelDifference(b)
	require.NoError(t, err)
	requireCentsEqual(t, want, got)
	_, err = a.LevelRatio(b)
	requireTuningError(t, err, tuning.ErrOutOfRange)
	assert.ErrorIs(t, err, cents.ErrOutOfRange)
}

func TestLevelDifferenceComplexityPrecedesStepOverflow(t *testing.T) {
	t.Parallel()
	a := complexTuning(t)
	b := a.MustWithReferenceAt(aspn(spelling.CDoubleSharp(), math.MaxInt), 440)

	_, err := a.LevelDifference(b)
	requireTuningError(t, err, tuning.ErrTooComplex)
	assert.ErrorIs(t, err, cents.ErrTooComplex)
	assert.NotErrorIs(t, err, cents.ErrOutOfRange)
	assert.NotErrorIs(t, err, system.ErrOutOfRange)
}

func TestLevelDifferenceStepOverflowUsesCentsCause(t *testing.T) {
	t.Parallel()
	other := newTestTuningAt(t, system.Traditional(), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), aspn(spelling.A(), math.MaxInt), 440)
	_, err := tuning.Standard().LevelDifference(other)
	requireTuningError(t, err, tuning.ErrOutOfRange)
	assert.ErrorIs(t, err, cents.ErrOutOfRange)
	assert.NotErrorIs(t, err, system.ErrOutOfRange)
}

func TestTuningFrequencyRatioRangeErrors(t *testing.T) {
	t.Parallel()
	low := newTestTuningAt(t, system.Traditional(), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), aspn(spelling.A(), 4), cents.MinNormalFrequency)
	_, err := low.FrequencyRatio(cents.MaxFrequency)
	requireTuningError(t, err, tuning.ErrOutOfRange)
	assert.ErrorIs(t, err, cents.ErrOutOfRange)

	high := newTestTuningAt(t, system.Traditional(), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), aspn(spelling.A(), 4), cents.MaxFrequency)
	_, err = high.FrequencyRatio(cents.MinNormalFrequency)
	requireTuningError(t, err, tuning.ErrOutOfRange)
	assert.ErrorIs(t, err, cents.ErrOutOfRange)
}
