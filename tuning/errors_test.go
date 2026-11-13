package tuning_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/frequency"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
)

func dependencySentinels() []error {
	return []error{
		cents.ErrOutOfRange,
		cents.ErrTooComplex,
		cents.ErrPrimeLimit,
		cents.ErrInvalidArgument,
		cents.ErrInvalidText,
		system.ErrInvalid,
		system.ErrDefinition,
		system.ErrClassCount,
		system.ErrOffsets,
		system.ErrDescending,
		system.ErrChromatic,
		system.ErrNoNatural,
		system.ErrFractionalShift,
		system.ErrRule,
		system.ErrModulusMismatch,
		system.ErrLimit,
		system.ErrNotApplicable,
		system.ErrInadmissible,
		system.ErrNoSpelling,
		system.ErrBadSpelling,
		system.ErrOutOfRange,
		temperament.ErrInvalidArgument,
		temperament.ErrOutOfRange,
		temperament.ErrPrimeLimit,
		temperament.ErrTooComplex,
		rounding.ErrUnset,
		rounding.ErrUnknownMode,
		rounding.ErrInvalidArgument,
		rounding.ErrOutOfRange,
		frequency.ErrInvalidFrequency,
		frequency.ErrInvalidTolerance,
		frequency.ErrUnset,
		frequency.ErrNoComponents,
		frequency.ErrInvalidText,
		frequency.ErrInvalidParameter,
		frequency.ErrOutOfRange,
	}
}

func TestErrorsWithoutDependencySentinels(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		call func() error
		want error
	}{
		{name: "E-1", call: func() error { _, err := (*tuning.Tuning)(nil).StepOf(aspn(spelling.A(), 4)); return err }, want: tuning.ErrUnset},
		{name: "E-2", call: func() error { _, err := tuning.NewReference(aspn(spelling.A(), 4), 0); return err }, want: tuning.ErrInvalidArgument},
		{
			name: "E-3",
			call: func() error {
				_, err := tuning.New(system.Positions19(), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 19)), tuning.A4At440())
				return err
			},
			want: tuning.ErrClassCount,
		},
		{
			name: "E-4",
			call: func() error {
				_, err := tuning.NewAt(system.Traditional(), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), aspn(spelling.MustParse("C+1/2"), 4), 440)
				return err
			},
			want: tuning.ErrInadmissible,
		},
		{name: "E-5", call: func() error { _, err := tuning.A4At440().CentsTo(tuning.C5At523()); return err }, want: tuning.ErrIncomparable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.call()
			requireTuningError(t, err, tc.want)
			requireNoDependencyError(t, err)
		})
	}
}

func TestOutOfRangeDependencyChains(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()

	_, err := std.CentsFromReferenceOfStep(90001)
	requireTuningError(t, err, tuning.ErrOutOfRange)
	assert.ErrorIs(t, err, cents.ErrOutOfRange)
	assert.NotErrorIs(t, err, system.ErrOutOfRange)

	_, err = std.StepOf(aspn(spelling.A(), math.MaxInt))
	requireTuningError(t, err, tuning.ErrOutOfRange)
	assert.ErrorIs(t, err, system.ErrOutOfRange)
	assert.NotErrorIs(t, err, cents.ErrOutOfRange)

	_, err = std.FrequencyOfStep(13000)
	requireTuningError(t, err, tuning.ErrOutOfRange)
	assert.ErrorIs(t, err, cents.ErrOutOfRange)
	assert.NotErrorIs(t, err, system.ErrOutOfRange)
}

func TestSpellingDependencyChains(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		call func() error
		want error
	}{
		{
			name: "not applicable",
			call: func() error {
				tune := newTestTuning(t, system.Positions19(), temperament.Equal19(), tuning.RootFromClass(pitchclass.MustNew(0, 19)), tuning.A4At440())
				_, err := tune.PitchOfStep(1, system.MustChromaticScale(spelling.C(), system.Major, system.Ascending))
				return err
			},
			want: system.ErrNotApplicable,
		},
		{
			name: "inadmissible tonic",
			call: func() error {
				tune := newTestTuning(t, traditionalSystemWithAlterations(-1, 1, "limited"), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), tuning.A4At440())
				rule := system.MustChromaticScale(spelling.CDoubleSharp(), system.Major, system.Ascending)
				_, err := tune.PitchOfStep(0, rule)
				return err
			},
			want: system.ErrInadmissible,
		},
		{
			name: "no spelling",
			call: func() error {
				tune := newTestTuning(t, traditionalSystemWithAlterations(0, 0, "naturals only"), temperament.Equal12(), tuning.RootFromClass(pitchclass.MustNew(0, 12)), tuning.A4At440())
				rule := system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)
				_, err := tune.PitchOfStep(1, rule)
				return err
			},
			want: system.ErrNoSpelling,
		},
		{
			name: "bad custom spelling",
			call: func() error {
				rule := system.MustCustomSpelling(func(*system.System, pitchclass.Class) spelling.Spelling {
					return spelling.C()
				})
				_, err := tuning.Standard().PitchOfStep(1, rule)
				return err
			},
			want: system.ErrBadSpelling,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.call()
			requireTuningError(t, err, tuning.ErrSpelling)
			assert.ErrorIs(t, err, tc.want)
			assert.NotErrorIs(t, err, cents.ErrOutOfRange)
		})
	}
}

func TestEveryOperationOnUnsetTuningReturnsUnset(t *testing.T) {
	t.Parallel()
	validRef := tuning.A4At440()
	validRoot := tuning.RootFromClass(pitchclass.MustNew(0, 12))
	validNote := aspn(spelling.A(), 4)
	for _, tc := range []struct {
		name string
		call func(*tuning.Tuning) error
	}{
		{name: "StepOf", call: func(v *tuning.Tuning) error { _, err := v.StepOf(validNote); return err }},
		{name: "PitchOfStep", call: func(v *tuning.Tuning) error { _, err := v.PitchOfStep(0, system.Sharps()); return err }},
		{name: "CentsFromReference", call: func(v *tuning.Tuning) error { _, err := v.CentsFromReference(validNote); return err }},
		{name: "CentsFromReferenceOfStep", call: func(v *tuning.Tuning) error { _, err := v.CentsFromReferenceOfStep(0); return err }},
		{name: "CentsBetween", call: func(v *tuning.Tuning) error { _, err := v.CentsBetween(validNote, validNote); return err }},
		{name: "CentsBetweenSteps", call: func(v *tuning.Tuning) error { _, err := v.CentsBetweenSteps(0, 1); return err }},
		{name: "Frequency", call: func(v *tuning.Tuning) error { _, err := v.Frequency(validNote); return err }},
		{name: "FrequencyOfStep", call: func(v *tuning.Tuning) error { _, err := v.FrequencyOfStep(0); return err }},
		{name: "ShiftedFrequency", call: func(v *tuning.Tuning) error { _, err := v.ShiftedFrequency(validNote, cents.Zero()); return err }},
		{name: "ShiftedFrequencyOfStep", call: func(v *tuning.Tuning) error { _, err := v.ShiftedFrequencyOfStep(0, cents.Zero()); return err }},
		{name: "Nearest", call: func(v *tuning.Tuning) error { _, _, err := v.Nearest(440); return err }},
		{name: "NearestBy", call: func(v *tuning.Tuning) error { _, _, err := v.NearestBy(440, rounding.HalfDown); return err }},
		{name: "NearestPitch", call: func(v *tuning.Tuning) error { _, _, err := v.NearestPitch(440, system.Sharps()); return err }},
		{name: "NearestPitchBy", call: func(v *tuning.Tuning) error {
			_, _, err := v.NearestPitchBy(440, rounding.HalfDown, system.Sharps())
			return err
		}},
		{name: "WithReferenceNote", call: func(v *tuning.Tuning) error { _, err := v.WithReferenceNote(validNote); return err }},
		{name: "WithReferenceFrequency", call: func(v *tuning.Tuning) error { _, err := v.WithReferenceFrequency(440); return err }},
		{name: "FrequencyRatio", call: func(v *tuning.Tuning) error { _, err := v.FrequencyRatio(440); return err }},
		{name: "WithReference", call: func(v *tuning.Tuning) error { _, err := v.WithReference(validRef); return err }},
		{name: "WithReferenceAt", call: func(v *tuning.Tuning) error { _, err := v.WithReferenceAt(validNote, 440); return err }},
		{name: "LevelDifference", call: func(v *tuning.Tuning) error { _, err := v.LevelDifference(tuning.Standard()); return err }},
		{name: "LevelRatio", call: func(v *tuning.Tuning) error { _, err := v.LevelRatio(tuning.Standard()); return err }},
		{name: "WithRoot", call: func(v *tuning.Tuning) error { _, err := v.WithRoot(validRoot); return err }},
		{name: "WithTemperament", call: func(v *tuning.Tuning) error { _, err := v.WithTemperament(temperament.Equal12()); return err }},
		{name: "WithSystem", call: func(v *tuning.Tuning) error { _, err := v.WithSystem(system.Traditional()); return err }},
		{name: "ChangedClasses", call: func(v *tuning.Tuning) error { _, err := v.ChangedClasses(validRoot); return err }},
		{name: "SoundsSame", call: func(v *tuning.Tuning) error { _, err := v.SoundsSame(tuning.Standard()); return err }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, value := range []*tuning.Tuning{nil, {}} {
				err := tc.call(value)
				requireTuningError(t, err, tuning.ErrUnset)
				requireNoDependencyError(t, err)
			}
		})
	}
}

func TestInvalidPitchIsNotDependencyError(t *testing.T) {
	t.Parallel()
	_, err := tuning.Standard().Frequency(pitch.Pitch{})
	requireTuningError(t, err, tuning.ErrUnset)
	assert.NotErrorIs(t, err, pitch.ErrInvalid)
}
