package ticks_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTicksOfDurationsAndPositions checks О-4, О-5, ИТ-1, ТК-1, ТК-2,
// КП-1, КП-4, КП-27, КП-45.
func TestTicksOfDurationsAndPositions(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		r     ticks.Resolution
		d     duration.Duration
		ticks int
	}{
		{name: "KP-1 quarter", r: res(480), d: dur(1, 4), ticks: 480},
		{name: "KP-1 eighth", r: res(480), d: dur(1, 8), ticks: 240},
		{name: "KP-1 triplet eighth", r: res(480), d: dur(1, 12), ticks: 160},
		{name: "KP-1 dotted quarter", r: res(480), d: dur(3, 8), ticks: 720},
		{name: "KP-1 whole", r: res(480), d: dur(1, 1), ticks: 1920},
		{name: "KP-1 quintuplet sixteenth", r: res(480), d: dur(1, 20), ticks: 96},
		{name: "KP-4 MusicXML eighth at 2", r: res(2), d: dur(1, 8), ticks: 1},
		{name: "KP-4 triplet eighth at 6", r: res(6), d: dur(1, 12), ticks: 2},
		{name: "KP-45 half with written half quarter", r: resQ(480, 16), d: dur(1, 2), ticks: 480},
		{name: "KP-45 quarter with written half quarter", r: resQ(480, 16), d: dur(1, 4), ticks: 240},
		{name: "dotted written quarter bb 12", r: resQ(3, 12), d: dur(1, 4), ticks: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n, err := tt.r.Ticks(tt.d)
			require.NoError(t, err)
			assert.Equal(t, tt.ticks, n)
			assert.Equal(t, tt.ticks, tt.r.MustTicks(tt.d))
			u, err := tt.r.Units(tt.d)
			require.NoError(t, err)
			assertFraction(t, fraction.FromInt(tt.ticks), u)
			assertFraction(t, fraction.FromInt(tt.ticks), tt.r.MustUnits(tt.d))

			p := duration.Origin().MustAdd(tt.d)
			pn, err := tt.r.PositionTicks(p)
			require.NoError(t, err)
			assert.Equal(t, tt.ticks, pn)
			assert.Equal(t, tt.ticks, tt.r.MustPositionTicks(p))
			pu, err := tt.r.PositionUnits(p)
			require.NoError(t, err)
			assertFraction(t, fraction.FromInt(tt.ticks), pu)
			assertFraction(t, fraction.FromInt(tt.ticks), tt.r.MustPositionUnits(p))

			back, err := tt.r.Duration(tt.ticks)
			require.NoError(t, err)
			assertEqualValue(t, tt.d, back)
			assertEqualValue(t, tt.d, tt.r.MustDuration(tt.ticks))
			backPosition, err := tt.r.Position(tt.ticks)
			require.NoError(t, err)
			assertEqualValue(t, p, backPosition)
			assertEqualValue(t, p, tt.r.MustPosition(tt.ticks))
		})
	}
}

// TestPositionTicksKP1AndKP27 checks the position cases of КП-1, КП-2, КП-27.
func TestPositionTicksKP1AndKP27(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 1440, res(480).MustPositionTicks(pos(3, 4)))
	assertEqualValue(t, pos(3, 4), res(480).MustPosition(1440))
	assertEqualValue(t, dur(3, 8), res(480).MustDuration(720))
	assertEqualValue(t, dur(1, 8), res(96).MustDuration(48))
	for _, d := range []int{1, 2, 96, 480, 32767, 1 << 40} {
		n, err := res(d).PositionTicks(duration.Origin())
		require.NoError(t, err)
		assert.Zero(t, n)
		p, err := res(d).Position(0)
		require.NoError(t, err)
		assertEqualValue(t, duration.Origin(), p)
		u, err := res(d).PositionUnits(duration.Origin())
		require.NoError(t, err)
		assert.True(t, u.IsZero())
	}
}

// TestUnitsAreExact checks П-1, О-5, О-6, ТК-2, КП-4.
func TestUnitsAreExact(t *testing.T) {
	t.Parallel()
	r := res(2)
	u, err := r.Units(dur(1, 12))
	require.NoError(t, err)
	assertFraction(t, frac(2, 3), u)
	n, err := r.Ticks(dur(1, 12))
	requireTicksError(t, err, ticks.ErrNotRepresentable)
	assert.Zero(t, n)
	requireMustPanic(t, err, ticks.ErrNotRepresentable, func() { r.MustTicks(dur(1, 12)) })
	_, err = r.PositionTicks(pos(1, 12))
	requireTicksError(t, err, ticks.ErrNotRepresentable)

	back, err := r.UnitsDuration(frac(2, 3))
	require.NoError(t, err)
	assertEqualValue(t, dur(1, 12), back)
	assertEqualValue(t, dur(1, 12), r.MustUnitsDuration(frac(2, 3)))
	backPosition, err := r.UnitsPosition(frac(2, 3))
	require.NoError(t, err)
	assertEqualValue(t, pos(1, 12), backPosition)
	assertEqualValue(t, pos(1, 12), r.MustUnitsPosition(frac(2, 3)))
	zero, err := r.UnitsPosition(fraction.Zero())
	require.NoError(t, err)
	assertEqualValue(t, duration.Origin(), zero)

	// With w = 1/2 the units are d·D/w.
	h := resQ(3, 16)
	hu, err := h.Units(dur(1, 4))
	require.NoError(t, err)
	assertFraction(t, frac(3, 2), hu)
	assertEqualValue(t, dur(1, 4), h.MustUnitsDuration(frac(3, 2)))
}

// TestConversionErrors checks КП-26 and 4.4 failures.
func TestConversionErrors(t *testing.T) {
	t.Parallel()
	r := res(480)
	huge := duration.MustFromFraction(fraction.FromInt(1 << 62))
	enormous := duration.MustFromFraction(hugeFraction(250))
	for _, tt := range []struct {
		name string
		call func() error
		must func()
		own  error
		deps []error
	}{
		{
			name: "duration of zero ticks",
			call: func() error { _, err := r.Duration(0); return err },
			must: func() { r.MustDuration(0) },
			own:  ticks.ErrInvalidArgument,
		},
		{
			name: "KP-26 duration of negative ticks",
			call: func() error { _, err := r.Duration(-1); return err },
			must: func() { r.MustDuration(-1) },
			own:  ticks.ErrInvalidArgument,
		},
		{
			name: "KP-26 position of negative ticks",
			call: func() error { _, err := r.Position(-1); return err },
			must: func() { r.MustPosition(-1) },
			own:  ticks.ErrInvalidArgument,
		},
		{
			name: "KP-26 unset duration",
			call: func() error { _, err := r.Ticks(duration.Duration{}); return err },
			must: func() { r.MustTicks(duration.Duration{}) },
			own:  ticks.ErrInvalidArgument,
			deps: []error{duration.ErrUnset},
		},
		{
			name: "KP-26 unset position",
			call: func() error { _, err := r.PositionTicks(duration.Position{}); return err },
			must: func() { r.MustPositionTicks(duration.Position{}) },
			own:  ticks.ErrInvalidArgument,
			deps: []error{duration.ErrUnset},
		},
		{
			name: "unset duration units",
			call: func() error { _, err := r.Units(duration.Duration{}); return err },
			must: func() { r.MustUnits(duration.Duration{}) },
			own:  ticks.ErrInvalidArgument,
			deps: []error{duration.ErrUnset},
		},
		{
			name: "unset position units",
			call: func() error { _, err := r.PositionUnits(duration.Position{}); return err },
			must: func() { r.MustPositionUnits(duration.Position{}) },
			own:  ticks.ErrInvalidArgument,
			deps: []error{duration.ErrUnset},
		},
		{
			name: "units of an unset fraction",
			call: func() error { _, err := r.UnitsDuration(fraction.Fraction{}); return err },
			must: func() { r.MustUnitsDuration(fraction.Fraction{}) },
			own:  ticks.ErrInvalidArgument,
			deps: []error{fraction.ErrInvalid},
		},
		{
			name: "zero units are not a duration",
			call: func() error { _, err := r.UnitsDuration(fraction.Zero()); return err },
			must: func() { r.MustUnitsDuration(fraction.Zero()) },
			own:  ticks.ErrInvalidArgument,
		},
		{
			name: "negative units are not a duration",
			call: func() error { _, err := r.UnitsDuration(frac(-1, 2)); return err },
			must: func() { r.MustUnitsDuration(frac(-1, 2)) },
			own:  ticks.ErrInvalidArgument,
		},
		{
			name: "negative units are not a position",
			call: func() error { _, err := r.UnitsPosition(frac(-1, 2)); return err },
			must: func() { r.MustUnitsPosition(frac(-1, 2)) },
			own:  ticks.ErrInvalidArgument,
		},
		{
			name: "position units of an unset fraction",
			call: func() error { _, err := r.UnitsPosition(fraction.Fraction{}); return err },
			must: func() { r.MustUnitsPosition(fraction.Fraction{}) },
			own:  ticks.ErrInvalidArgument,
			deps: []error{fraction.ErrInvalid},
		},
		{
			name: "ticks outside int",
			call: func() error { _, err := r.Ticks(huge); return err },
			must: func() { r.MustTicks(huge) },
			own:  ticks.ErrOutOfRange,
		},
		{
			name: "position ticks outside int",
			call: func() error { _, err := r.PositionTicks(duration.Origin().MustAdd(huge)); return err },
			must: func() { r.MustPositionTicks(duration.Origin().MustAdd(huge)) },
			own:  ticks.ErrOutOfRange,
		},
		{
			name: "units outside fraction",
			call: func() error { _, err := r.Units(enormous); return err },
			must: func() { r.MustUnits(enormous) },
			own:  ticks.ErrOutOfRange,
			deps: []error{fraction.ErrOutOfRange},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.call()
			requireTicksError(t, err, tt.own, tt.deps...)
			requireMustPanic(t, err, tt.own, tt.must)
		})
	}
}

// TestQuantize checks О-7, П-12, ТК-3, КП-17, КП-23 and ИТ-11 on exact values.
func TestQuantize(t *testing.T) {
	t.Parallel()
	r := res(480)
	// KP-17: 1/28 at D = 480 is 480/7 ≈ 68.57 units.
	q, err := r.Quantize(dur(1, 28), rounding.HalfEven)
	require.NoError(t, err)
	assertQuantized(t, q, 69, frac(1, 4480))
	assert.Equal(t, "69+1/4480", q.String())
	assertQuantized(t, r.MustQuantize(dur(1, 28), rounding.HalfUp), 69, frac(1, 4480))
	assertQuantized(t, r.MustQuantize(dur(1, 28), rounding.Down), 68, frac(-1, 3360))
	assert.Equal(t, "68-1/3360", r.MustQuantize(dur(1, 28), rounding.Down).String())
	assertQuantized(t, r.MustQuantizePosition(pos(1, 28), rounding.Down), 68, frac(-1, 3360))

	// Exact value: zero deviation.
	exact := r.MustQuantize(dur(3, 8), rounding.Down)
	assertQuantized(t, exact, 720, fraction.Zero())
	assert.Equal(t, "720+0/1", exact.String())

	// Ties: 1/1280 is 1.5 units and 1/768 is 2.5 units at D = 480.
	for _, mode := range allModes {
		t.Run("tie 1.5 "+mode.String(), func(t *testing.T) {
			t.Parallel()
			got, err := r.Quantize(dur(1, 1280), mode)
			require.NoError(t, err)
			want := tieIndex(mode, 1, false)
			assertQuantized(t, got, int(want), fraction.MustNew(int(want)*2-3, 3840))
			gotPosition, err := r.QuantizePosition(pos(1, 1280), mode)
			require.NoError(t, err)
			assert.True(t, got.Equal(gotPosition))
		})
		t.Run("tie 2.5 "+mode.String(), func(t *testing.T) {
			t.Parallel()
			got, err := r.Quantize(dur(1, 768), mode)
			require.NoError(t, err)
			want := tieIndex(mode, 2, true)
			assertQuantized(t, got, int(want), fraction.MustNew(int(want)*2-5, 3840))
		})
	}

	// П-12, КП-23: a duration that rounds to zero ticks is an error, a position is not.
	_, err = res(1).Quantize(dur(1, 64), rounding.Down)
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { res(1).MustQuantize(dur(1, 64), rounding.Down) })
	zero, err := res(1).QuantizePosition(pos(1, 64), rounding.Down)
	require.NoError(t, err)
	assertQuantized(t, zero, 0, frac(-1, 64))
	up, err := res(1).Quantize(dur(1, 64), rounding.Up)
	require.NoError(t, err)
	assertQuantized(t, up, 1, frac(15, 64))

	// With w = 1/2: 1/3 is 2/3·D units.
	assertQuantized(t, resQ(1, 16).MustQuantize(dur(1, 3), rounding.HalfUp), 1, frac(1, 6))
}

func TestQuantizeErrors(t *testing.T) {
	t.Parallel()
	r := res(480)
	for _, tt := range []struct {
		name string
		call func() error
		own  error
		deps []error
		not  []error
	}{
		{
			name: "unset mode",
			call: func() error { _, err := r.Quantize(dur(1, 4), rounding.Mode(0)); return err },
			own:  ticks.ErrInvalidArgument,
			deps: []error{rounding.ErrUnset},
		},
		{
			name: "unknown mode",
			call: func() error { _, err := r.QuantizePosition(pos(1, 4), rounding.Mode(99)); return err },
			own:  ticks.ErrInvalidArgument,
			deps: []error{rounding.ErrUnknownMode},
		},
		{
			name: "duration is checked before the mode",
			call: func() error { _, err := r.Quantize(duration.Duration{}, rounding.Mode(0)); return err },
			own:  ticks.ErrInvalidArgument,
			deps: []error{duration.ErrUnset},
			not:  []error{rounding.ErrUnset},
		},
		{
			name: "receiver is checked first",
			call: func() error { _, err := ticks.Resolution{}.Quantize(duration.Duration{}, rounding.Mode(0)); return err },
			own:  ticks.ErrUnset,
			not:  []error{duration.ErrUnset, rounding.ErrUnset},
		},
		{
			name: "quantized ticks outside int",
			call: func() error {
				_, err := r.Quantize(duration.MustFromFraction(fraction.FromInt(1<<62)), rounding.Down)
				return err
			},
			own: ticks.ErrOutOfRange,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.call()
			requireTicksError(t, err, tt.own, tt.deps...)
			for _, n := range tt.not {
				assert.NotErrorIs(t, err, n)
			}
		})
	}
}

// TestRescale checks О-8, ИТ-4, ТК-5, КП-5, КП-23.
func TestRescale(t *testing.T) {
	t.Parallel()
	from := res(480)
	to := res(96)
	for _, tt := range []struct {
		in, out int
	}{
		{in: 120, out: 24},
		{in: 100, out: 20},
		{in: 480, out: 96},
	} {
		n, err := from.Rescale(tt.in, to)
		require.NoError(t, err)
		assert.Equal(t, tt.out, n)
		assert.Equal(t, tt.out, from.MustRescale(tt.in, to))
		p, err := from.RescalePosition(tt.in, to)
		require.NoError(t, err)
		assert.Equal(t, tt.out, p)
		assert.Equal(t, tt.out, from.MustRescalePosition(tt.in, to))
	}
	assert.Equal(t, 0, from.MustRescalePosition(0, to))
	assert.Equal(t, 2400, to.MustRescale(480, from))

	// KP-5: 1 tick at 480 is 1/5 tick at 96.
	_, err := from.Rescale(1, to)
	requireTicksError(t, err, ticks.ErrNotRepresentable)
	requireMustPanic(t, err, ticks.ErrNotRepresentable, func() { from.MustRescale(1, to) })
	_, err = from.RescalePosition(1, to)
	requireTicksError(t, err, ticks.ErrNotRepresentable)

	// Different written quarters: t·(w·D′)/(w′·D).
	assert.Equal(t, 240, from.MustRescale(480, resQ(480, 16)))
	assert.Equal(t, 960, resQ(480, 16).MustRescale(480, from))

	for name, call := range map[string]func() error{
		"zero duration ticks":     func() error { _, err := from.Rescale(0, to); return err },
		"negative duration ticks": func() error { _, err := from.Rescale(-1, to); return err },
		"negative position ticks": func() error { _, err := from.RescalePosition(-1, to); return err },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, call(), ticks.ErrInvalidArgument)
		})
	}
	_, err = from.Rescale(1, ticks.Resolution{})
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = from.Rescale(math.MaxInt, res(960))
	requireTicksError(t, err, ticks.ErrOutOfRange)
}

// TestQuantizeRescale checks О-8 with a mode, П-12, КП-23 and all nine modes on ties.
func TestQuantizeRescale(t *testing.T) {
	t.Parallel()
	// КП-23: 1 tick at 480 → 96 rounds down to 0.
	_, err := res(480).QuantizeRescale(1, res(96), rounding.Down)
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { res(480).MustQuantizeRescale(1, res(96), rounding.Down) })
	zero, err := res(480).QuantizeRescalePosition(1, res(96), rounding.Down)
	require.NoError(t, err)
	assertQuantized(t, zero, 0, frac(-1, 1920))
	assertQuantized(t, res(480).MustQuantizeRescale(100, res(96), rounding.Down), 20, fraction.Zero())
	// 101 ticks at 480 = 20.2 ticks at 96.
	assertQuantized(t, res(480).MustQuantizeRescale(101, res(96), rounding.HalfUp), 20, frac(-1, 1920))
	assertQuantized(t, res(480).MustQuantizeRescalePosition(101, res(96), rounding.Up), 21, frac(4, 1920))

	// 3 ticks at D = 2 are 1.5 ticks at D = 1: deviation n/4 − 3/8.
	for _, mode := range allModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			want := int(tieIndex(mode, 1, false))
			got, err := res(2).QuantizeRescale(3, res(1), mode)
			require.NoError(t, err)
			assertQuantized(t, got, want, fraction.MustNew(2*want-3, 8))
			gotPosition, err := res(2).QuantizeRescalePosition(3, res(1), mode)
			require.NoError(t, err)
			assert.True(t, got.Equal(gotPosition))
		})
	}
	_, err = res(2).QuantizeRescale(3, res(1), rounding.Mode(0))
	requireTicksError(t, err, ticks.ErrInvalidArgument, rounding.ErrUnset)
	_, err = res(2).QuantizeRescale(0, res(1), rounding.Down)
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	_, err = res(2).QuantizeRescalePosition(-1, res(1), rounding.Down)
	requireTicksError(t, err, ticks.ErrInvalidArgument)
}

// TestTicksDurationRoundTrip checks ИТ-1: Ticks(Duration(t)) == t.
func TestTicksDurationRoundTrip(t *testing.T) {
	t.Parallel()
	for _, r := range []ticks.Resolution{res(1), res(7), res(96), res(480), resQ(480, 16), resQ(5, 3)} {
		for _, n := range []int{1, 2, 3, 59, 480, 1919, 1 << 30} {
			d := r.MustDuration(n)
			assert.Equal(t, n, r.MustTicks(d), "%s %d", r, n)
			p := r.MustPosition(n)
			assert.Equal(t, n, r.MustPositionTicks(p), "%s %d", r, n)
		}
	}
}
