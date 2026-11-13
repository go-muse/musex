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

func TestStepOf(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	for _, tc := range []struct {
		name string
		note pitch.Pitch
		want system.Step
	}{
		{name: "reference A4", note: aspn(spelling.A(), 4), want: 0},
		{name: "C5", note: aspn(spelling.C(), 5), want: 3},
		{name: "B3", note: aspn(spelling.B(), 3), want: -10},
		{name: "written B-sharp3 is positional octave 4", note: aspn(spelling.BSharp(), 4), want: -9},
		{name: "C4", note: aspn(spelling.C(), 4), want: -9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := std.StepOf(tc.note)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("enharmonic written C-flat4", func(t *testing.T) {
		t.Parallel()
		cFlat4 := aspn(spelling.CFlat(), 3)
		b3 := aspn(spelling.B(), 3)
		stepB3, err := std.StepOf(b3)
		require.NoError(t, err)
		stepCFlat4, err := std.StepOf(cFlat4)
		require.NoError(t, err)
		assert.Equal(t, stepB3, stepCFlat4)
		assert.Equal(t, std.MustFrequency(b3), std.MustFrequency(cFlat4))
		requireCentsEqual(t, std.MustCentsFromReference(b3), std.MustCentsFromReference(cFlat4))
	})
}

func TestPitchOfStep(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	for _, tc := range []struct {
		name string
		step system.Step
		rule system.SpellingRule
		want pitch.Pitch
	}{
		{name: "sharps", step: 1, rule: system.Sharps(), want: aspn(spelling.ASharp(), 4)},
		{name: "flats", step: 1, rule: system.Flats(), want: aspn(spelling.BFlat(), 4)},
		{name: "negative step", step: -9, rule: system.Sharps(), want: aspn(spelling.C(), 4)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := std.PitchOfStep(tc.step, tc.rule)
			require.NoError(t, err)
			assert.True(t, tc.want.Equal(got))
		})
	}

	for _, k := range []system.Step{-1000, -25, -12, -1, 0, 1, 12, 25, 1000} {
		t.Run(strconv.Itoa(int(k)), func(t *testing.T) {
			t.Parallel()
			got := std.MustPitchOfStep(k, system.Sharps())
			assert.Equal(t, k, std.MustStepOf(got))
		})
	}
}

func TestOneClassSystem(t *testing.T) {
	t.Parallel()
	sys := traditionalSystemWithClasses(1)
	tune := newTestTuningAt(t, sys, temperament.MustEDO(1), tuning.RootFromClass(pitchclass.MustNew(0, 1)), aspn(spelling.C(), 4), 440)

	assert.Equal(t, system.Step(1), tune.MustStepOf(aspn(spelling.C(), 5)))
	assert.Equal(t, system.Step(-1), tune.MustStepOf(aspn(spelling.C(), 3)))
	for _, tc := range []struct {
		step system.Step
		want float64
	}{
		{step: -2, want: 110},
		{step: -1, want: 220},
		{step: 0, want: 440},
		{step: 1, want: 880},
		{step: 2, want: 1760},
	} {
		assert.Equal(t, tc.want, tune.MustFrequencyOfStep(tc.step))
		requireCentsEqual(t, cents.MustFromInt(1200*int(tc.step)), tune.MustCentsFromReferenceOfStep(tc.step))
	}
}

func TestCentsFromReferenceAndBetween(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	a4 := aspn(spelling.A(), 4)
	c5 := aspn(spelling.C(), 5)
	b3 := aspn(spelling.B(), 3)

	for _, tc := range []struct {
		name string
		note pitch.Pitch
		step system.Step
		want cents.Cents
	}{
		{name: "reference", note: a4, step: 0, want: cents.Zero()},
		{name: "C5", note: c5, step: 3, want: cents.MustFromInt(300)},
		{name: "B3", note: b3, step: -10, want: cents.MustFromInt(-1000)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			requireCentsEqual(t, tc.want, std.MustCentsFromReference(tc.note))
			requireCentsEqual(t, tc.want, std.MustCentsFromReferenceOfStep(tc.step))
		})
	}

	requireCentsEqual(t, cents.MustFromInt(300), std.MustCentsBetween(a4, c5), "K-7")
	requireCentsEqual(t, cents.MustFromInt(-300), std.MustCentsBetween(c5, a4))
	requireCentsEqual(t, cents.MustFromInt(300), std.MustCentsBetweenSteps(0, 3))
	requireCentsEqual(t, cents.MustFromInt(-300), std.MustCentsBetweenSteps(3, 0))

	pyth := newPythagoreanTuning(t, tuning.RootFromClass(pitchclass.MustNew(0, 12)))
	requireCentsEqual(t, cents.MustFromRatio(32, 27), pyth.MustCentsBetween(a4, c5), "K-8")
	assert.Equal(t, "294.134997", pyth.MustCentsBetween(a4, c5).Decimal(6, 0))
	requireCentsEqual(t, cents.MustFromRatio(32, 27), pyth.MustCentsFromReference(c5))

	changedReference := pyth.MustWithReferenceFrequency(435)
	requireCentsEqual(t, pyth.MustCentsBetween(a4, c5), changedReference.MustCentsBetween(a4, c5), "distance does not depend on the reference")
}

func TestFrequencyControlExamples(t *testing.T) {
	t.Parallel()
	rootC := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	rootA := tuning.RootFromClass(pitchclass.MustNew(9, 12))
	et435 := tuning.Standard().MustWithReferenceFrequency(435)
	pythC := newPythagoreanTuning(t, rootC)
	pythA := newPythagoreanTuning(t, rootA)
	justC := newTestTuning(t, system.Traditional(), temperament.Just5Limit(), rootC, tuning.A4At440())
	justA := newTestTuning(t, system.Traditional(), temperament.Just5Limit(), rootA, tuning.A4At440())
	scientific := newTestTuning(t, system.Traditional(), temperament.Equal12(), rootC, tuning.C0At16())

	for _, tc := range []struct {
		name    string
		tune    *tuning.Tuning
		note    pitch.Pitch
		rounded string
	}{
		{name: "K-1 A4", tune: tuning.Standard(), note: aspn(spelling.A(), 4), rounded: "440.0000"},
		{name: "K-1 A5", tune: tuning.Standard(), note: aspn(spelling.A(), 5), rounded: "880.0000"},
		{name: "K-1 A3", tune: tuning.Standard(), note: aspn(spelling.A(), 3), rounded: "220.0000"},
		{name: "K-2 C5", tune: tuning.Standard(), note: aspn(spelling.C(), 5), rounded: "523.2511"},
		{name: "K-2 C-sharp5", tune: tuning.Standard(), note: aspn(spelling.CSharp(), 5), rounded: "554.3653"},
		{name: "K-2 C0", tune: tuning.Standard(), note: aspn(spelling.C(), 0), rounded: "16.3516"},
		{name: "K-3 C5 at 435", tune: et435, note: aspn(spelling.C(), 5), rounded: "517.3051"},
		{name: "K-4 Pythagorean C5", tune: pythC, note: aspn(spelling.C(), 5), rounded: "521.4815"},
		{name: "K-4 Pythagorean F5", tune: pythC, note: aspn(spelling.F(), 5), rounded: "695.3086"},
		{name: "K-5 root A F5", tune: pythA, note: aspn(spelling.F(), 5), rounded: "704.7949"},
		{name: "K-5 root A C5", tune: pythA, note: aspn(spelling.C(), 5), rounded: "521.4815"},
		{name: "K-6 just root C C-sharp5", tune: justC, note: aspn(spelling.CSharp(), 5), rounded: "563.2000"},
		{name: "K-6 just root A C-sharp5", tune: justA, note: aspn(spelling.CSharp(), 5), rounded: "550.0000"},
		{name: "K-6 just root C A4", tune: justC, note: aspn(spelling.A(), 4), rounded: "440.0000"},
		{name: "K-6 just root A A4", tune: justA, note: aspn(spelling.A(), 4), rounded: "440.0000"},
		{name: "K-19 scientific C4", tune: scientific, note: aspn(spelling.C(), 4), rounded: "256.0000"},
		{name: "K-19 scientific A4", tune: scientific, note: aspn(spelling.A(), 4), rounded: "430.5390"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := tc.tune.Frequency(tc.note)
			require.NoError(t, err)
			k := tc.tune.MustStepOf(tc.note)
			want := frequencyAt(t, tc.tune.System(), tc.tune.Temperament(), tc.tune.RootClass().Number(), tc.tune.Reference(), k)
			assert.Equal(t, want, got)
			assert.Equal(t, want, tc.tune.MustFrequencyOfStep(k))
			assert.Equal(t, tc.rounded, strconv.FormatFloat(got, 'f', 4, 64))
		})
	}
}

func TestShiftedFrequency(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	c4 := aspn(spelling.C(), 4)
	cSharp4 := aspn(spelling.CSharp(), 4)
	a4 := aspn(spelling.A(), 4)

	shifted := std.MustShiftedFrequency(c4, cents.MustFromInt(100))
	assert.Equal(t, math.Float64bits(std.MustFrequency(cSharp4)), math.Float64bits(shifted), "K-20")
	assert.Equal(t, 880.0, std.MustShiftedFrequency(a4, cents.Octave()))

	for _, k := range []system.Step{-25, -1, 0, 1, 25} {
		plain := std.MustFrequencyOfStep(k)
		zeroShift := std.MustShiftedFrequencyOfStep(k, cents.Zero())
		assert.Equal(t, math.Float64bits(plain), math.Float64bits(zeroShift), "G-6 at step %d", k)
	}
	assert.Equal(t, math.Float64bits(std.MustFrequency(c4)), math.Float64bits(std.MustShiftedFrequency(c4, cents.Zero())))
}

func TestCentsRangeWithoutIntermediateOverflow(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	for _, tc := range []struct {
		name string
		step system.Step
		want cents.Cents
	}{
		{name: "upper boundary", step: 90000, want: cents.MustFromInt(9_000_000)},
		{name: "lower boundary", step: -90000, want: cents.MustFromInt(-9_000_000)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			requireCentsEqual(t, tc.want, std.MustCentsFromReferenceOfStep(tc.step))
		})
	}

	for _, k := range []system.Step{90001, -90001} {
		_, err := std.CentsFromReferenceOfStep(k)
		requireTuningError(t, err, tuning.ErrOutOfRange)
		assert.ErrorIs(t, err, cents.ErrOutOfRange)
	}

	requireCentsEqual(t, cents.MustFromInt(100), std.MustCentsBetweenSteps(200000, 200001))
	requireCentsEqual(t, cents.MustFromInt(-100), std.MustCentsBetweenSteps(200001, 200000))

	_, err := std.ShiftedFrequencyOfStep(90001, cents.MustFromInt(-100))
	requireTuningError(t, err, tuning.ErrOutOfRange)
	assert.ErrorIs(t, err, cents.ErrOutOfRange)
}

func TestComplexityPrecedesRange(t *testing.T) {
	t.Parallel()
	tune := complexTuning(t)

	_, err := tune.CentsFromReferenceOfStep(1)
	requireTuningError(t, err, tuning.ErrTooComplex)
	assert.ErrorIs(t, err, cents.ErrTooComplex)
	assert.NotErrorIs(t, err, cents.ErrOutOfRange)

	_, err = tune.CentsFromReferenceOfStep(90001)
	requireTuningError(t, err, tuning.ErrTooComplex)
	assert.ErrorIs(t, err, cents.ErrTooComplex)
	assert.NotErrorIs(t, err, cents.ErrOutOfRange)

	_, err = tune.CentsBetweenSteps(0, 1)
	requireTuningError(t, err, tuning.ErrTooComplex)
	assert.ErrorIs(t, err, cents.ErrTooComplex)
}

func TestFrequencyRangeErrors(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	for _, k := range []system.Step{13000, -13000} {
		_, err := std.FrequencyOfStep(k)
		requireTuningError(t, err, tuning.ErrOutOfRange)
		assert.ErrorIs(t, err, cents.ErrOutOfRange)
		assert.NotErrorIs(t, err, system.ErrOutOfRange)
	}
}

func TestPitchOperationErrors(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()

	_, err := std.StepOf(pitch.Pitch{})
	requireTuningError(t, err, tuning.ErrUnset)
	_, err = std.CentsFromReference(aspn(spelling.MustParse("C+1/2"), 4))
	requireTuningError(t, err, tuning.ErrInadmissible)
	_, err = std.Frequency(aspn(spelling.MustParse("C+1/2"), 4))
	requireTuningError(t, err, tuning.ErrInadmissible)
	_, err = std.CentsBetween(aspn(spelling.MustParse("C+1/2"), 4), pitch.Pitch{})
	requireTuningError(t, err, tuning.ErrInadmissible, "the from pitch is checked first")
	_, err = std.CentsBetween(aspn(spelling.A(), 4), pitch.Pitch{})
	requireTuningError(t, err, tuning.ErrUnset)

	_, err = std.StepOf(aspn(spelling.A(), math.MaxInt))
	requireTuningError(t, err, tuning.ErrOutOfRange)
	assert.ErrorIs(t, err, system.ErrOutOfRange)
	assert.NotErrorIs(t, err, cents.ErrOutOfRange)

	representable, err := std.PitchOfStep(system.Step(math.MaxInt), system.Sharps())
	require.NoError(t, err)
	wantRepresentable := system.Traditional().MustMove(aspn(spelling.A(), 4), system.Step(math.MaxInt), system.Sharps())
	assert.True(t, wantRepresentable.Equal(representable))

	highReference := newTestTuningAt(t, system.Traditional(), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), aspn(spelling.A(), math.MaxInt), 440)
	_, err = highReference.PitchOfStep(3, system.Sharps())
	requireTuningError(t, err, tuning.ErrOutOfRange)
	assert.ErrorIs(t, err, system.ErrOutOfRange)

	_, err = std.PitchOfStep(0, system.SpellingRule{})
	requireTuningError(t, err, tuning.ErrUnset)

	chromatic := system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)
	nontraditional := newTestTuning(t, system.Positions19(), temperament.Equal19(), tuning.RootFromClass(pitchclass.MustNew(0, 19)), tuning.A4At440())
	_, err = nontraditional.PitchOfStep(1, chromatic)
	requireTuningError(t, err, tuning.ErrSpelling)
	assert.ErrorIs(t, err, system.ErrNotApplicable)
}

func TestK23ChangeTemperamentFrequency(t *testing.T) {
	t.Parallel()
	changed := tuning.Standard().MustWithTemperament(temperament.PythagoreanEFlatGSharp())
	assert.Equal(t, 440.0, changed.MustFrequency(aspn(spelling.A(), 4)))
	assert.Equal(t, "521.4815", strconv.FormatFloat(changed.MustFrequency(aspn(spelling.C(), 5)), 'f', 4, 64))
}
