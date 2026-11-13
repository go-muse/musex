package tuning_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNearestControlExamples(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	for _, tc := range []struct {
		name       string
		freq       float64
		wantStep   system.Step
		wantPitch  string
		wantDigits string
	}{
		{name: "A4", freq: 440, wantStep: 0, wantPitch: "A ASPN 4", wantDigits: "0.000000"},
		{name: "A5", freq: 880, wantStep: 12, wantPitch: "A ASPN 5", wantDigits: "0.000000"},
		{name: "445 Hz", freq: 445, wantStep: 0, wantPitch: "A ASPN 4", wantDigits: "19.562175"},
		{name: "262 Hz", freq: 262, wantStep: -9, wantPitch: "C ASPN 4", wantDigits: "2.475946"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			step, deviation, err := std.Nearest(tc.freq)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStep, step)
			d := cents.MustBetween(440, tc.freq)
			h := cents.MustFromInt(100 * int(tc.wantStep))
			requireCentsEqual(t, d.MustSub(h), deviation)
			assert.Equal(t, tc.wantDigits, deviation.Decimal(6, 0))

			note, pitchDeviation, err := std.NearestPitch(tc.freq, system.Sharps())
			require.NoError(t, err)
			assert.Equal(t, tc.wantPitch, note.String())
			requireCentsEqual(t, deviation, pitchDeviation)
			assert.Equal(t, tc.wantStep, std.MustStepOf(note))
		})
	}
}

func TestNearestDefaultAndDirectedModes(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	assert.Equal(t, rounding.HalfDown, tuning.NearestMode)
	tie := cents.MustFromInt(50).MustApplyTo(440)

	step, deviation, err := std.Nearest(tie)
	require.NoError(t, err)
	assert.Equal(t, system.Step(0), step)
	requireCentsEqual(t, cents.MustFromInt(50), deviation)

	step, deviation, err = std.NearestBy(tie, rounding.HalfUp)
	require.NoError(t, err)
	assert.Equal(t, system.Step(1), step)
	requireCentsEqual(t, cents.MustFromInt(-50), deviation)

	for _, tc := range []struct {
		name       string
		freq       float64
		mode       rounding.Mode
		wantStep   system.Step
		wantDigits string
	}{
		{name: "K-12a down", freq: 445, mode: rounding.Down, wantStep: 0, wantDigits: "19.562175"},
		{name: "K-12a up", freq: 445, mode: rounding.Up, wantStep: 1, wantDigits: "-80.437825"},
		{name: "K-12a exact node up", freq: 440, mode: rounding.Up, wantStep: 0, wantDigits: "0.000000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			step, deviation, err := std.NearestBy(tc.freq, tc.mode)
			require.NoError(t, err)
			assert.Equal(t, tc.wantStep, step)
			assert.Equal(t, tc.wantDigits, deviation.Decimal(6, 0))
		})
	}

	note, deviation, err := std.NearestPitchBy(445, rounding.Up, system.Sharps())
	require.NoError(t, err)
	assert.True(t, aspn(spelling.ASharp(), 4).Equal(note))
	assert.Equal(t, "-80.437825", deviation.Decimal(6, 0))
}

func TestNearestAllModesBothSides(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	for _, tc := range []struct {
		name  string
		cents int
		want  map[rounding.Mode]system.Step
	}{
		{
			name:  "positive below midpoint",
			cents: 40,
			want: map[rounding.Mode]system.Step{
				rounding.Down: 0, rounding.Up: 1, rounding.TowardZero: 0, rounding.AwayFromZero: 1,
				rounding.HalfUp: 0, rounding.HalfDown: 0, rounding.HalfAwayFromZero: 0, rounding.HalfTowardZero: 0, rounding.HalfEven: 0,
			},
		},
		{
			name:  "positive above midpoint",
			cents: 60,
			want: map[rounding.Mode]system.Step{
				rounding.Down: 0, rounding.Up: 1, rounding.TowardZero: 0, rounding.AwayFromZero: 1,
				rounding.HalfUp: 1, rounding.HalfDown: 1, rounding.HalfAwayFromZero: 1, rounding.HalfTowardZero: 1, rounding.HalfEven: 1,
			},
		},
		{
			name:  "positive tie with even lower neighbor",
			cents: 50,
			want: map[rounding.Mode]system.Step{
				rounding.Down: 0, rounding.Up: 1, rounding.TowardZero: 0, rounding.AwayFromZero: 1,
				rounding.HalfUp: 1, rounding.HalfDown: 0, rounding.HalfAwayFromZero: 1, rounding.HalfTowardZero: 0, rounding.HalfEven: 0,
			},
		},
		{
			name:  "positive tie with odd lower neighbor",
			cents: 150,
			want: map[rounding.Mode]system.Step{
				rounding.Down: 1, rounding.Up: 2, rounding.TowardZero: 1, rounding.AwayFromZero: 2,
				rounding.HalfUp: 2, rounding.HalfDown: 1, rounding.HalfAwayFromZero: 2, rounding.HalfTowardZero: 1, rounding.HalfEven: 2,
			},
		},
		{
			name:  "negative below midpoint toward lower neighbor",
			cents: -60,
			want: map[rounding.Mode]system.Step{
				rounding.Down: -1, rounding.Up: 0, rounding.TowardZero: 0, rounding.AwayFromZero: -1,
				rounding.HalfUp: -1, rounding.HalfDown: -1, rounding.HalfAwayFromZero: -1, rounding.HalfTowardZero: -1, rounding.HalfEven: -1,
			},
		},
		{
			name:  "negative above midpoint toward upper neighbor",
			cents: -40,
			want: map[rounding.Mode]system.Step{
				rounding.Down: -1, rounding.Up: 0, rounding.TowardZero: 0, rounding.AwayFromZero: -1,
				rounding.HalfUp: 0, rounding.HalfDown: 0, rounding.HalfAwayFromZero: 0, rounding.HalfTowardZero: 0, rounding.HalfEven: 0,
			},
		},
		{
			name:  "negative tie with even upper neighbor",
			cents: -50,
			want: map[rounding.Mode]system.Step{
				rounding.Down: -1, rounding.Up: 0, rounding.TowardZero: 0, rounding.AwayFromZero: -1,
				rounding.HalfUp: 0, rounding.HalfDown: -1, rounding.HalfAwayFromZero: -1, rounding.HalfTowardZero: 0, rounding.HalfEven: 0,
			},
		},
		{
			name:  "negative tie with even lower neighbor",
			cents: -150,
			want: map[rounding.Mode]system.Step{
				rounding.Down: -2, rounding.Up: -1, rounding.TowardZero: -1, rounding.AwayFromZero: -2,
				rounding.HalfUp: -1, rounding.HalfDown: -2, rounding.HalfAwayFromZero: -2, rounding.HalfTowardZero: -1, rounding.HalfEven: -2,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			freq := cents.MustFromInt(tc.cents).MustApplyTo(440)
			d := cents.MustBetween(440, freq)
			for _, mode := range rounding.Modes() {
				t.Run(mode.String(), func(t *testing.T) {
					t.Parallel()
					wantStep := tc.want[mode]
					step, deviation, err := std.NearestBy(freq, mode)
					require.NoError(t, err)
					assert.Equal(t, wantStep, step)
					h := cents.MustFromInt(100 * int(wantStep))
					requireCentsEqual(t, d.MustSub(h), deviation)
				})
			}
		})
	}
}

func TestNearestExactNodeInEveryMode(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	for _, mode := range rounding.Modes() {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			step, deviation, err := std.NearestBy(440, mode)
			require.NoError(t, err)
			assert.Zero(t, step)
			requireCentsEqual(t, cents.Zero(), deviation)
		})
	}
}

func TestNearestSubnormalFrequency(t *testing.T) {
	t.Parallel()
	step, deviation, err := tuning.Standard().Nearest(math.SmallestNonzeroFloat64)
	require.NoError(t, err)
	assert.Equal(t, system.Step(-12993), step)
	assert.Equal(t, "-16.961235", deviation.Decimal(6, 0), "K-26 uses the shortest decimal input representation")
	d := cents.MustBetween(440, math.SmallestNonzeroFloat64)
	requireCentsEqual(t, d.MustSub(cents.MustFromInt(-1299300)), deviation)
}

func TestNearestNonUniformTemperament(t *testing.T) {
	t.Parallel()
	tune := newPythagoreanTuning(t, tuning.RootFromClass(pitchclass.MustNew(0, 12)))
	h0 := tune.MustCentsFromReferenceOfStep(0)
	h1 := tune.MustCentsFromReferenceOfStep(1)
	midpoint := h1.MustDiv(2)
	freq := midpoint.MustApplyTo(440)
	d := cents.MustBetween(440, freq)
	comparison := d.MustSub(h0).Compare(h1.MustSub(d))

	for _, tc := range []struct {
		name string
		mode rounding.Mode
		call func() (system.Step, cents.Cents, error)
	}{
		{name: "default", mode: rounding.HalfDown, call: func() (system.Step, cents.Cents, error) { return tune.Nearest(freq) }},
		{name: "half up", mode: rounding.HalfUp, call: func() (system.Step, cents.Cents, error) { return tune.NearestBy(freq, rounding.HalfUp) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantStep := system.Step(tc.mode.MustChooseIndex(0, comparison))
			step, deviation, err := tc.call()
			require.NoError(t, err)
			assert.Equal(t, wantStep, step)
			wantHeight := h0
			if wantStep == 1 {
				wantHeight = h1
			}
			requireCentsEqual(t, d.MustSub(wantHeight), deviation)
		})
	}
}

func TestNearestNineteenClasses(t *testing.T) {
	t.Parallel()
	tune := newTestTuning(t, system.Positions19(), temperament.Equal19(), tuning.RootFromClass(pitchclass.MustNew(0, 19)), tuning.A4At440())
	firstStep := cents.MustStep(1, 19)
	midpoint := firstStep.MustDiv(2)
	freq := midpoint.MustApplyTo(440)
	d := cents.MustBetween(440, freq)
	comparison := d.Compare(firstStep.MustSub(d))

	for _, tc := range []struct {
		name string
		mode rounding.Mode
		call func() (system.Step, cents.Cents, error)
	}{
		{name: "default", mode: rounding.HalfDown, call: func() (system.Step, cents.Cents, error) { return tune.Nearest(freq) }},
		{name: "half up", mode: rounding.HalfUp, call: func() (system.Step, cents.Cents, error) { return tune.NearestBy(freq, rounding.HalfUp) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantStep := system.Step(tc.mode.MustChooseIndex(0, comparison))
			step, deviation, err := tc.call()
			require.NoError(t, err)
			assert.Equal(t, wantStep, step)
			wantHeight := cents.Zero()
			if wantStep == 1 {
				wantHeight = firstStep
			}
			requireCentsEqual(t, d.MustSub(wantHeight), deviation)
		})
	}
}

func TestNearestInputErrors(t *testing.T) {
	t.Parallel()
	std := tuning.Standard()
	for _, freq := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, _, err := std.Nearest(freq)
		requireTuningError(t, err, tuning.ErrInvalidArgument)
		requireNoDependencyError(t, err)
	}

	_, _, err := std.NearestBy(440, rounding.Mode(0))
	requireTuningError(t, err, tuning.ErrUnset)
	requireNoDependencyError(t, err)
	_, _, err = std.NearestBy(440, rounding.Mode(255))
	requireTuningError(t, err, tuning.ErrInvalidArgument)
	requireNoDependencyError(t, err)
	_, _, err = std.NearestPitch(440, system.SpellingRule{})
	requireTuningError(t, err, tuning.ErrUnset)

	chromatic := system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)
	nontraditional := newTestTuning(t, system.Positions19(), temperament.Equal19(), tuning.RootFromClass(pitchclass.MustNew(0, 19)), tuning.A4At440())
	_, _, err = nontraditional.NearestPitch(440, chromatic)
	requireTuningError(t, err, tuning.ErrSpelling)
	assert.ErrorIs(t, err, system.ErrNotApplicable)
}

func TestNearestComplexityInEveryModeIncludingExactNode(t *testing.T) {
	t.Parallel()
	tune := complexTuning(t)
	for _, mode := range rounding.Modes() {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			_, _, err := tune.NearestBy(440, mode)
			requireTuningError(t, err, tuning.ErrTooComplex)
			assert.ErrorIs(t, err, cents.ErrTooComplex)
			assert.NotErrorIs(t, err, temperament.ErrTooComplex)
			assert.NotErrorIs(t, err, temperament.ErrOutOfRange)
		})
	}
}
