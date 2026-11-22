package tempomap_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests of this file cover the step methods of architecture 14 of task
// 28e.1 (business specification 4a: О-30, ТК-20, КП-48). Expected values come
// from the contract: examples 14.2, the exact formulas of 4.7 and the
// equivalence with the decimal methods for step = 10^−D.

// stepsRitardandoMap is the map of КП-15 of task 28e.1: quarter = 120 at 0
// with a gradual change from 0 and quarter = 100 as a new main tempo at 1.
// Start(1) = 12·ln(6/5) ≈ 2.18785868152745551454 s.
func stepsRitardandoMap() tempomap.Map {
	return tempomap.MustNew(quarter(120)).
		MustWithChange(tempomap.MustNewGradual(at(0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarter(100)))
}

// stepTieIndex is the expected index of a value exactly between lower and
// lower+1 for a non-negative value.
func stepTieIndex(mode rounding.Mode, lower int, lowerEven bool) int {
	switch mode {
	case rounding.Down, rounding.TowardZero, rounding.HalfDown, rounding.HalfTowardZero:
		return lower
	case rounding.HalfEven:
		if lowerEven {
			return lower
		}
	}
	return lower + 1
}

// stepInnerIndex is the expected index of a value strictly inside
// (lower, lower+1) other than the midpoint; above reports a value above the
// midpoint.
func stepInnerIndex(mode rounding.Mode, lower int, above bool) int {
	switch mode {
	case rounding.Down, rounding.TowardZero:
		return lower
	case rounding.Up, rounding.AwayFromZero:
		return lower + 1
	}
	if above {
		return lower + 1
	}
	return lower
}

// decimalIndex returns value·10^places as an int; value must be a multiple of 10^−places.
func decimalIndex(t *testing.T, value fraction.Fraction, places int) int {
	t.Helper()
	r, ok := value.BigRat()
	require.True(t, ok)
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	r.Mul(r, new(big.Rat).SetInt(scale))
	require.True(t, r.IsInt(), "%s is not on the decimal grid", r.RatString())
	require.True(t, r.Num().IsInt64())
	return int(r.Num().Int64())
}

func powSeconds(places int) tempo.Seconds {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	return tempo.MustSecondsFromFraction(fraction.MustFromBigInt(big.NewInt(1), scale))
}

func powDuration(places int) duration.Duration {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	return duration.MustFromFraction(fraction.MustFromBigInt(big.NewInt(1), scale))
}

// TestStepsKP48Examples checks the examples of architecture 14.2 and КП-48.
func TestStepsKP48Examples(t *testing.T) {
	t.Parallel()
	m := stepsRitardandoMap()
	got, err := m.BetweenSteps(at(0, 1), at(1, 4), sec(1, 1000000), rounding.HalfUp)
	require.NoError(t, err)
	assert.Equal(t, 510715, got)
	assert.Equal(t, 510715, m.MustBetweenSteps(at(0, 1), at(1, 4), sec(1, 1000000), rounding.HalfUp))

	down, err := m.StartSteps(at(1, 1), sec(1, 1000), rounding.Down)
	require.NoError(t, err)
	assert.Equal(t, 2187, down)
	assert.Equal(t, 2188, m.MustStartSteps(at(1, 1), sec(1, 1000), rounding.Up))

	for _, mode := range allRoundingModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			// 510715.373… µs and 2187.8587… ms are irrational.
			assert.Equal(t, stepInnerIndex(mode, 510715, false), m.MustBetweenSteps(at(0, 1), at(1, 4), sec(1, 1000000), mode))
			assert.Equal(t, stepInnerIndex(mode, 2187, true), m.MustStartSteps(at(1, 1), sec(1, 1000), mode))
			// Start(1) = 4 s at quarter = 60; 4/(8/3) = 3/2 is a tie with an odd lower index.
			assert.Equal(t, stepTieIndex(mode, 1, false), tempomap.MustNew(quarter(60)).MustStartSteps(at(1, 1), sec(8, 3), mode))
		})
	}
	tie := tempomap.MustNew(quarter(60))
	assert.Equal(t, 2, tie.MustStartSteps(at(1, 1), sec(8, 3), rounding.HalfUp))
	assert.Equal(t, 1, tie.MustStartSteps(at(1, 1), sec(8, 3), rounding.HalfDown))
	assert.Equal(t, 2, tie.MustStartSteps(at(1, 1), sec(8, 3), rounding.HalfEven))
	assert.Equal(t, 1, tie.MustStartSteps(at(1, 1), sec(8, 3), rounding.Down))
	assert.Equal(t, 2, tie.MustStartSteps(at(1, 1), sec(8, 3), rounding.Up))
}

// TestStepsHaveNoPrecisionLimit checks the guarantee of 14.2: an index with
// 19 digits of a logarithmic value is correctly rounded.
func TestStepsHaveNoPrecisionLimit(t *testing.T) {
	t.Parallel()
	m := stepsRitardandoMap()
	// Start(1)·10^18 = 2187858681527455514.5406…
	step := powSeconds(18)
	assert.Equal(t, 2187858681527455514, m.MustStartSteps(at(1, 1), step, rounding.Down))
	assert.Equal(t, 2187858681527455515, m.MustStartSteps(at(1, 1), step, rounding.Up))
	assert.Equal(t, 2187858681527455515, m.MustStartSteps(at(1, 1), step, rounding.HalfEven))
	assert.Equal(t, 2187858681527455515, m.MustStartSteps(at(1, 1), step, rounding.HalfDown))
}

// TestStepsMatchDecimalMethods checks that for step = 10^−D the step methods
// return Value·10^D of the decimal methods (implementation plan, 4a).
func TestStepsMatchDecimalMethods(t *testing.T) {
	t.Parallel()
	m := stepsRitardandoMap().MustWithExtension(tempomap.MustNewExtension(at(1, 1), sec(1, 3)))
	addresses := []tempomap.Address{at(1, 8), at(1, 2), at(1, 1), at(3, 2)}
	for _, places := range []int{3, 6} {
		for _, mode := range allRoundingModes {
			t.Run(mode.String()+"/"+string(rune('0'+places)), func(t *testing.T) {
				t.Parallel()
				step := powSeconds(places)
				for _, a := range addresses {
					start := m.MustStart(a, places, mode)
					assert.Equal(t, decimalIndex(t, start.MustValue().MustFraction(), places), m.MustStartSteps(a, step, mode), "start %s", a)
					end := m.MustEnd(a, places, mode)
					assert.Equal(t, decimalIndex(t, end.MustValue().MustFraction(), places), m.MustEndSteps(a, step, mode), "end %s", a)
				}
				between := m.MustBetween(at(1, 4), at(3, 2), places, mode)
				assert.Equal(t, decimalIndex(t, between.MustValue().MustFraction(), places), m.MustBetweenSteps(at(1, 4), at(3, 2), step, mode))
				for _, s := range []tempo.Seconds{sec(1, 1), sec(2, 1), sec(9, 4), secondsRat("2.5"), sec(4, 1)} {
					position := m.MustPosition(s, places, mode)
					assert.Equal(t, decimalIndex(t, position.MustValue().MustFraction(), places), m.MustPositionSteps(s, powDuration(places), mode), "position %s", s)
				}
				for _, segment := range m.MustSegments() {
					result, ok := segment.MustSeconds(places, mode)
					k, okSteps := segment.MustSecondsSteps(step, mode)
					assert.Equal(t, ok, okSteps)
					if ok {
						assert.Equal(t, decimalIndex(t, result.MustValue().MustFraction(), places), k)
					} else {
						assert.Zero(t, k)
					}
				}
			})
		}
	}
}

// TestStepsOnExactValues checks that an exact value gives mode.RatMultiple(X, h).
func TestStepsOnExactValues(t *testing.T) {
	t.Parallel()
	hold := newHoldMap() // quarter = 60, hold of 1 s at 1/2: Start(1/2) = 2, End(1/2) = 3.
	for _, mode := range allRoundingModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			want := func(x, step *big.Rat) int {
				k, _, err := mode.RatMultiple(x, step)
				require.NoError(t, err)
				return int(k.Int64())
			}
			assert.Equal(t, want(big.NewRat(2, 1), big.NewRat(3, 7)), hold.MustStartSteps(at(1, 2), sec(3, 7), mode))
			assert.Equal(t, want(big.NewRat(3, 1), big.NewRat(3, 7)), hold.MustEndSteps(at(1, 2), sec(3, 7), mode))
			// Start(1) = 5: the hold at 1/2 is included.
			assert.Equal(t, want(big.NewRat(5, 1), big.NewRat(2, 3)), hold.MustStartSteps(at(1, 1), sec(2, 3), mode))
			// Between(1/2, 1) = 3: the hold at from is included.
			assert.Equal(t, want(big.NewRat(3, 1), big.NewRat(2, 1)), hold.MustBetweenSteps(at(1, 2), at(1, 1), sec(2, 1), mode))
			assert.Equal(t, stepTieIndex(mode, 1, false), hold.MustBetweenSteps(at(1, 2), at(1, 1), sec(2, 1), mode))
			// Inside the hold the position is the position of the hold.
			assert.Equal(t, 4, hold.MustPositionSteps(sec(5, 2), duration.MustNew(1, 8), mode))
			assert.Equal(t, 4, hold.MustPositionSteps(sec(2, 1), duration.MustNew(1, 8), mode))
			assert.Equal(t, 4, hold.MustPositionSteps(sec(3, 1), duration.MustNew(1, 8), mode))
			// 7/2 s is position 5/8 = 2.5 quarter steps.
			assert.Equal(t, stepTieIndex(mode, 2, true), hold.MustPositionSteps(sec(7, 2), duration.MustNew(1, 4), mode))
			// Zero.
			assert.Zero(t, hold.MustStartSteps(at(0, 1), sec(1, 1), mode))
			assert.Zero(t, hold.MustPositionSteps(tempo.ZeroSeconds(), duration.MustNew(1, 4), mode))
			assert.Zero(t, hold.MustBetweenSteps(at(1, 4), at(1, 4), sec(1, 1), mode))
		})
	}
}

// TestAddressAtSteps checks AddressAtSteps of 14.2.
func TestAddressAtSteps(t *testing.T) {
	t.Parallel()
	grid := measure.MustNew(meter.FourFour())
	m := tempomap.MustNewMeasured(grid, quarter(60))
	// The map of КП-15 in measures: bar 2 is position 1.
	gradual := tempomap.MustNewMeasured(grid, quarter(120)).
		MustWithChange(tempomap.MustNewGradual(tempomap.MustAtBar(measure.MustNewAddress(1, pos(0, 1))), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond), grid).
		MustWithChange(tempomap.MustNewMain(tempomap.MustAtBar(measure.MustNewAddress(2, pos(0, 1))), quarter(100)), grid)
	a, k, err := m.AddressAtSteps(grid, sec(6, 1), duration.MustNew(1, 8), rounding.Down)
	require.NoError(t, err)
	assert.Equal(t, 12, k)
	assert.True(t, a.Equal(measure.MustNewAddress(2, pos(1, 2))), "address %s", a)
	for _, mode := range allRoundingModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			// 13/2 s is position 13/8: 6.5 quarter steps.
			address, index := m.MustAddressAtSteps(grid, sec(13, 2), duration.MustNew(1, 4), mode)
			want := stepTieIndex(mode, 6, true)
			assert.Equal(t, want, index)
			assert.True(t, address.Equal(grid.MustLocate(duration.Origin().MustAdd(duration.MustNew(want, 4)))), "address %s", address)
			// An exact position 13/12 (13/3 s): k = mode(13/12·1000) by
			// RatMultiple, the address is that of k/1000.
			k, _, err := mode.RatMultiple(big.NewRat(13, 12), big.NewRat(1, 1000))
			require.NoError(t, err)
			stepAddress, stepIndex := m.MustAddressAtSteps(grid, sec(13, 3), powDuration(3), mode)
			assert.Equal(t, int(k.Int64()), stepIndex)
			assert.True(t, stepAddress.Equal(grid.MustLocate(duration.Origin().MustAdd(duration.MustNew(int(k.Int64()), 1000)))), "address %s", stepAddress)
			// On a gradual change the decimal result is approximate and
			// agrees with the step result for step = 10^−3 and 10^−6.
			for _, places := range []int{3, 6} {
				decimalAddress, result := gradual.MustAddressAt(grid, sec(1, 1), places, mode)
				require.True(t, result.MustApproximate())
				gradualAddress, gradualIndex := gradual.MustAddressAtSteps(grid, sec(1, 1), powDuration(places), mode)
				assert.Equal(t, decimalIndex(t, result.MustValue().MustFraction(), places), gradualIndex)
				assert.True(t, decimalAddress.Equal(gradualAddress), "address %s, want %s", gradualAddress, decimalAddress)
			}
		})
	}

	finite := measure.MustNew(meter.FourFour()).MustWithBarCount(1)
	fm := tempomap.MustNewMeasured(finite, quarter(60))
	a, k, err = fm.AddressAtSteps(finite, sec(4, 1), duration.MustNew(1, 4), rounding.Down)
	requireTempomapError(t, err, tempomap.ErrInvalidAddress, measure.ErrOutsideGrid)
	assert.Equal(t, measure.Address{}, a)
	assert.Zero(t, k)
	requirePanicError(t, err, tempomap.ErrInvalidAddress, func() {
		fm.MustAddressAtSteps(finite, sec(4, 1), duration.MustNew(1, 4), rounding.Down)
	})
	// Rounding below the end keeps the address inside the grid.
	a, k = fm.MustAddressAtSteps(finite, sec(399, 100), duration.MustNew(1, 4), rounding.Down)
	assert.Equal(t, 3, k)
	assert.True(t, a.Equal(measure.MustNewAddress(1, pos(3, 4))))

	_, _, err = tempomap.MustNew(quarter(60)).AddressAtSteps(grid, sec(1, 1), duration.MustNew(1, 4), rounding.Down)
	requireTempomapError(t, err, tempomap.ErrInvalidArgument)
	_, err = m.PositionSteps(sec(1, 1), duration.MustNew(1, 4), rounding.Down)
	requireTempomapError(t, err, tempomap.ErrInvalidArgument)
}

// TestSegmentSecondsSteps checks Segment.SecondsSteps of 14.1, 14.2.
func TestSegmentSecondsSteps(t *testing.T) {
	t.Parallel()
	segments := stepsRitardandoMap().MustSegments()
	require.Len(t, segments, 2)
	k, ok, err := segments[0].SecondsSteps(sec(1, 1000), rounding.Down)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 2187, k)
	k, ok = segments[0].MustSecondsSteps(sec(1, 1000), rounding.HalfEven)
	assert.True(t, ok)
	assert.Equal(t, 2188, k)

	k, ok, err = segments[1].SecondsSteps(sec(1, 1000), rounding.Down)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Zero(t, k)

	for _, tt := range []struct {
		name string
		seg  tempomap.Segment
		step tempo.Seconds
		mode rounding.Mode
		own  error
		deps []error
	}{
		{name: "open segment checks the step", seg: segments[1], step: tempo.Seconds{}, mode: rounding.Down, own: tempomap.ErrInvalidArgument, deps: []error{tempo.ErrUnset}},
		{name: "zero step", seg: segments[0], step: tempo.ZeroSeconds(), mode: rounding.Down, own: tempomap.ErrInvalidArgument},
		{name: "open segment checks the mode", seg: segments[1], step: sec(1, 1), mode: rounding.Mode(0), own: tempomap.ErrInvalidArgument, deps: []error{rounding.ErrUnset}},
		{name: "unknown mode", seg: segments[0], step: sec(1, 1), mode: rounding.Mode(99), own: tempomap.ErrInvalidArgument, deps: []error{rounding.ErrUnknownMode}},
		{name: "unset segment", seg: tempomap.Segment{}, step: tempo.Seconds{}, mode: rounding.Mode(0), own: tempomap.ErrUnset},
		{name: "index outside int", seg: segments[0], step: powSeconds(19), mode: rounding.Down, own: tempomap.ErrOutOfRange, deps: []error{fraction.ErrNotRepresentable}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			k, ok, err := tt.seg.SecondsSteps(tt.step, tt.mode)
			requireTempomapError(t, err, tt.own, tt.deps...)
			assert.Zero(t, k)
			assert.False(t, ok)
			requirePanicError(t, err, tt.own, func() { tt.seg.MustSecondsSteps(tt.step, tt.mode) })
		})
	}
}

// TestStepsErrors checks the failures and their order (14.2, tempomap 4.11).
func TestStepsErrors(t *testing.T) {
	t.Parallel()
	m := stepsRitardandoMap()
	grid := measure.MustNew(meter.FourFour())
	measured := tempomap.MustNewMeasured(grid, quarter(60))
	unfinished := tempomap.MustNew(quarter(60)).MustWithChange(tempomap.MustNewGradual(at(1, 1), tempomap.LinearSpeed))
	constant := tempomap.MustNew(quarter(60))
	for _, tt := range []struct {
		name string
		call func() (int, error)
		must func()
		own  error
		deps []error
		not  []error
	}{
		{
			name: "unset map",
			call: func() (int, error) { return tempomap.Map{}.StartSteps(at(1, 1), sec(1, 1), rounding.Down) },
			must: func() { tempomap.Map{}.MustStartSteps(at(1, 1), sec(1, 1), rounding.Down) },
			own:  tempomap.ErrUnset,
		},
		{
			name: "unset address",
			call: func() (int, error) { return m.EndSteps(tempomap.Address{}, sec(1, 1), rounding.Down) },
			must: func() { m.MustEndSteps(tempomap.Address{}, sec(1, 1), rounding.Down) },
			own:  tempomap.ErrUnset,
		},
		{
			name: "unset step",
			call: func() (int, error) { return m.StartSteps(at(1, 1), tempo.Seconds{}, rounding.Down) },
			must: func() { m.MustStartSteps(at(1, 1), tempo.Seconds{}, rounding.Down) },
			own:  tempomap.ErrInvalidArgument,
			deps: []error{tempo.ErrUnset},
		},
		{
			name: "zero step",
			call: func() (int, error) { return m.EndSteps(at(1, 1), tempo.ZeroSeconds(), rounding.Down) },
			must: func() { m.MustEndSteps(at(1, 1), tempo.ZeroSeconds(), rounding.Down) },
			own:  tempomap.ErrInvalidArgument,
		},
		{
			name: "unset mode",
			call: func() (int, error) { return m.BetweenSteps(at(0, 1), at(1, 1), sec(1, 1), rounding.Mode(0)) },
			must: func() { m.MustBetweenSteps(at(0, 1), at(1, 1), sec(1, 1), rounding.Mode(0)) },
			own:  tempomap.ErrInvalidArgument,
			deps: []error{rounding.ErrUnset},
		},
		{
			name: "unknown mode",
			call: func() (int, error) { return m.StartSteps(at(1, 1), sec(1, 1), rounding.Mode(99)) },
			must: func() { m.MustStartSteps(at(1, 1), sec(1, 1), rounding.Mode(99)) },
			own:  tempomap.ErrInvalidArgument,
			deps: []error{rounding.ErrUnknownMode},
		},
		{
			name: "step is checked before the mode",
			call: func() (int, error) { return m.StartSteps(at(1, 1), tempo.Seconds{}, rounding.Mode(0)) },
			must: func() { m.MustStartSteps(at(1, 1), tempo.Seconds{}, rounding.Mode(0)) },
			own:  tempomap.ErrInvalidArgument,
			deps: []error{tempo.ErrUnset},
			not:  []error{rounding.ErrUnset},
		},
		{
			name: "grid is checked before the step",
			call: func() (int, error) { return m.StartSteps(at(1, 1), tempo.Seconds{}, rounding.Down, grid) },
			must: func() { m.MustStartSteps(at(1, 1), tempo.Seconds{}, rounding.Down, grid) },
			own:  tempomap.ErrInvalidArgument,
			not:  []error{tempo.ErrUnset},
		},
		{
			name: "measured map needs a grid",
			call: func() (int, error) {
				return measured.StartSteps(tempomap.MustAtBar(measure.MustNewAddress(2, pos(0, 1))), sec(1, 1), rounding.Down)
			},
			must: func() {
				measured.MustStartSteps(tempomap.MustAtBar(measure.MustNewAddress(2, pos(0, 1))), sec(1, 1), rounding.Down)
			},
			own: tempomap.ErrInvalidArgument,
		},
		{
			name: "unset grid",
			call: func() (int, error) {
				return measured.StartSteps(tempomap.MustAtBar(measure.MustNewAddress(2, pos(0, 1))), sec(1, 1), rounding.Down, measure.Grid{})
			},
			must: func() {
				measured.MustStartSteps(tempomap.MustAtBar(measure.MustNewAddress(2, pos(0, 1))), sec(1, 1), rounding.Down, measure.Grid{})
			},
			own:  tempomap.ErrInvalidArgument,
			deps: []error{measure.ErrUnset},
		},
		{
			name: "reverse interval",
			call: func() (int, error) { return m.BetweenSteps(at(1, 1), at(0, 1), sec(1, 1), rounding.Down) },
			must: func() { m.MustBetweenSteps(at(1, 1), at(0, 1), sec(1, 1), rounding.Down) },
			own:  tempomap.ErrInvalidArgument,
		},
		{
			name: "unfinished gradual change",
			call: func() (int, error) { return unfinished.StartSteps(at(2, 1), sec(1, 1), rounding.Down) },
			must: func() { unfinished.MustStartSteps(at(2, 1), sec(1, 1), rounding.Down) },
			own:  tempomap.ErrUnfinished,
		},
		{
			name: "unfinished gradual change in position",
			call: func() (int, error) {
				return unfinished.PositionSteps(sec(10, 1), duration.MustNew(1, 4), rounding.Down)
			},
			must: func() { unfinished.MustPositionSteps(sec(10, 1), duration.MustNew(1, 4), rounding.Down) },
			own:  tempomap.ErrUnfinished,
		},
		{
			name: "start index outside int",
			call: func() (int, error) { return constant.StartSteps(at(4, 1), powSeconds(18), rounding.Down) },
			must: func() { constant.MustStartSteps(at(4, 1), powSeconds(18), rounding.Down) },
			own:  tempomap.ErrOutOfRange,
			deps: []error{fraction.ErrNotRepresentable},
		},
		{
			name: "irrational index outside int",
			call: func() (int, error) { return m.StartSteps(at(1, 1), powSeconds(19), rounding.Down) },
			must: func() { m.MustStartSteps(at(1, 1), powSeconds(19), rounding.Down) },
			own:  tempomap.ErrOutOfRange,
			deps: []error{fraction.ErrNotRepresentable},
		},
		{
			name: "unset seconds of position",
			call: func() (int, error) { return m.PositionSteps(tempo.Seconds{}, duration.MustNew(1, 4), rounding.Down) },
			must: func() { m.MustPositionSteps(tempo.Seconds{}, duration.MustNew(1, 4), rounding.Down) },
			own:  tempomap.ErrInvalidArgument,
			deps: []error{tempo.ErrUnset},
		},
		{
			name: "unset step of position",
			call: func() (int, error) { return m.PositionSteps(sec(1, 1), duration.Duration{}, rounding.Down) },
			must: func() { m.MustPositionSteps(sec(1, 1), duration.Duration{}, rounding.Down) },
			own:  tempomap.ErrInvalidArgument,
			deps: []error{duration.ErrUnset},
		},
		{
			name: "position index outside int",
			call: func() (int, error) {
				return constant.PositionSteps(tempo.MustSecondsFromFraction(fraction.FromInt(1<<40)), powDuration(18), rounding.Down)
			},
			must: func() {
				constant.MustPositionSteps(tempo.MustSecondsFromFraction(fraction.FromInt(1<<40)), powDuration(18), rounding.Down)
			},
			own:  tempomap.ErrOutOfRange,
			deps: []error{fraction.ErrNotRepresentable},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			k, err := tt.call()
			requireTempomapError(t, err, tt.own, tt.deps...)
			for _, n := range tt.not {
				assert.NotErrorIs(t, err, n)
			}
			assert.Zero(t, k)
			requirePanicError(t, err, tt.own, tt.must)
		})
	}
	_, _, err := tempomap.Map{}.AddressAtSteps(grid, sec(1, 1), duration.MustNew(1, 4), rounding.Down)
	requireTempomapError(t, err, tempomap.ErrUnset)
	_, _, err = measured.AddressAtSteps(measure.Grid{}, sec(1, 1), duration.MustNew(1, 4), rounding.Down)
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, measure.ErrUnset)
	_, _, err = measured.AddressAtSteps(grid, sec(1, 1), duration.Duration{}, rounding.Down)
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, duration.ErrUnset)
}

// TestStepsDoNotChangeInputs checks the immutability rules of tempomap 3.2.
func TestStepsDoNotChangeInputs(t *testing.T) {
	t.Parallel()
	m := stepsRitardandoMap().MustWithExtension(tempomap.MustNewExtension(at(1, 1), sec(1, 1)))
	copied := m
	before := m.String()
	step := sec(1, 1000)
	stepBefore := step.String()
	m.MustStartSteps(at(1, 1), step, rounding.HalfUp)
	m.MustEndSteps(at(1, 1), step, rounding.HalfUp)
	m.MustBetweenSteps(at(0, 1), at(3, 2), step, rounding.HalfUp)
	m.MustPositionSteps(sec(3, 1), duration.MustNew(1, 1920), rounding.HalfUp)
	for _, segment := range m.MustSegments() {
		segment.MustSecondsSteps(step, rounding.HalfUp)
	}
	_, _ = m.StartSteps(at(1, 1), tempo.ZeroSeconds(), rounding.HalfUp)
	assert.Equal(t, before, m.String())
	assert.True(t, m.Equal(copied))
	assert.Equal(t, stepBefore, step.String())
}

// TestStepsConcurrentCalls checks that step methods are safe for concurrent use.
func TestStepsConcurrentCalls(t *testing.T) {
	t.Parallel()
	m := stepsRitardandoMap()
	results := make(chan int, 16)
	for range 16 {
		go func() {
			results <- m.MustStartSteps(at(1, 1), sec(1, 1000), rounding.Down)
		}()
	}
	for range 16 {
		assert.Equal(t, 2187, <-results)
	}
}

// TestWeightedBetweenStepsExact checks 14.6: round the sum once, without
// rounding its terms; repeated, overlapping and unordered terms are valid.
func TestWeightedBetweenStepsExact(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60))
	terms := []tempomap.WeightedInterval{
		{From: at(0, 1), To: at(1, 4), Weight: big.NewRat(1, 3)},
		{From: at(1, 4), To: at(1, 2), Weight: big.NewRat(2, 3)},
	}
	for _, mode := range allRoundingModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			got, err := m.WeightedBetweenSteps(terms, big.NewRat(2, 3), mode)
			require.NoError(t, err)
			assert.Equal(t, stepTieIndex(mode, 1, false), got)
			assert.Equal(t, got, m.MustWeightedBetweenSteps(terms, big.NewRat(2, 3), mode))
			assert.Equal(t, stepTieIndex(mode, 0, true), m.MustWeightedBetweenSteps(terms, big.NewRat(2, 1), mode))
			assert.Equal(t, got, m.MustWeightedBetweenSteps([]tempomap.WeightedInterval{terms[1], terms[0]}, big.NewRat(2, 3), mode))
			// Two half-second contributions sum to an exact integer, even
			// though rounding each would give 0 or 2 in several modes.
			repeated := []tempomap.WeightedInterval{
				{From: at(0, 1), To: at(1, 4), Weight: big.NewRat(1, 2)},
				{From: at(0, 1), To: at(1, 4), Weight: big.NewRat(1, 2)},
			}
			assert.Equal(t, 1, m.MustWeightedBetweenSteps(repeated, big.NewRat(1, 1), mode))
			overlapped := []tempomap.WeightedInterval{
				{From: at(1, 4), To: at(1, 2), Weight: big.NewRat(1, 2)},
				{From: at(0, 1), To: at(1, 2), Weight: big.NewRat(1, 4)},
				{From: at(1, 1), To: at(1, 1), Weight: big.NewRat(13, 1)},
			}
			assert.Equal(t, 1, m.MustWeightedBetweenSteps(overlapped, big.NewRat(1, 1), mode))
			assert.Zero(t, m.MustWeightedBetweenSteps(nil, big.NewRat(1, 1), mode))
			assert.Zero(t, m.MustWeightedBetweenSteps([]tempomap.WeightedInterval{}, big.NewRat(1, 1), mode))
		})
	}
}

// TestWeightedBetweenStepsHolds checks Start(to)-Start(from): a hold at
// from is included, one at to is excluded, and a zero interval is zero.
func TestWeightedBetweenStepsHolds(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60))
	for _, p := range []tempomap.Address{at(0, 1), at(1, 4), at(1, 2)} {
		m = m.MustWithExtension(tempomap.MustNewExtension(p, sec(1, 1)))
	}
	for _, tt := range []struct {
		name     string
		from, to tempomap.Address
		want     int
	}{
		{name: "left boundary", from: at(0, 1), to: at(1, 4), want: 2},
		{name: "right boundary", from: at(1, 8), to: at(1, 4), want: 1},
		{name: "interior and left holds", from: at(0, 1), to: at(1, 2), want: 4},
		{name: "zero interval on a hold", from: at(1, 4), to: at(1, 4)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Weight two, step two: the numerical index equals elapsed seconds.
			// The half-second case instead uses step one to get an integral index.
			step := big.NewRat(2, 1)
			if tt.name == "right boundary" {
				step = big.NewRat(1, 1)
			}
			got, err := m.WeightedBetweenSteps([]tempomap.WeightedInterval{{From: tt.from, To: tt.to, Weight: big.NewRat(2, 1)}}, step, rounding.HalfEven)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestWeightedBetweenStepsMixedExactSum uses an independent 85-digit
// Decimal calculation: X = 4*ln(6/5)+2/5 = 1.1292862271758185048468721...
func TestWeightedBetweenStepsMixedExactSum(t *testing.T) {
	t.Parallel()
	m := stepsRitardandoMap()
	terms := []tempomap.WeightedInterval{
		{From: at(0, 1), To: at(1, 1), Weight: big.NewRat(1, 3)},
		{From: at(1, 1), To: at(5, 4), Weight: big.NewRat(2, 3)},
	}
	for _, mode := range allRoundingModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			got, err := m.WeightedBetweenSteps(terms, big.NewRat(1, 1000000000000000000), mode)
			require.NoError(t, err)
			assert.Equal(t, stepInnerIndex(mode, 1129286227175818504, true), got)
			assert.Equal(t, got, m.MustWeightedBetweenSteps([]tempomap.WeightedInterval{terms[1], terms[0]}, big.NewRat(1, 1000000000000000000), mode))
		})
	}
}

func TestWeightedBetweenStepsUnboundedRationals(t *testing.T) {
	t.Parallel()
	huge := new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 300))
	tiny := new(big.Rat).Inv(huge)
	for _, scale := range []*big.Rat{huge, tiny} {
		terms := []tempomap.WeightedInterval{{From: at(0, 1), To: at(1, 4), Weight: scale}}
		m := tempomap.MustNew(quarter(60))
		assert.Equal(t, 1, m.MustWeightedBetweenSteps(terms, scale, rounding.HalfEven))
		// Both the sum and the individual factors may exceed fraction;
		// only the final integer index is bounded by int.
		terms = append(terms, terms[0])
		assert.Equal(t, 2, m.MustWeightedBetweenSteps(terms, scale, rounding.Down))
		terms[0].To, terms[1].To = at(1, 1), at(0, 1)
		assert.Equal(t, 2, stepsRitardandoMap().MustWeightedBetweenSteps(terms, scale, rounding.HalfEven))
	}
}

// TestWeightedBetweenStepsErrors checks every validation stage of 14.6,
// including precedence when several invalid inputs occur together.
func TestWeightedBetweenStepsErrors(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60))
	grid := measure.MustNew(meter.FourFour()).MustWithBarCount(1)
	measured := tempomap.MustNewMeasured(grid, quarter(60))
	unfinished := m.MustWithChange(tempomap.MustNewGradual(at(1, 1), tempomap.LinearSpeed))
	one := big.NewRat(1, 1)
	valid := tempomap.WeightedInterval{From: at(0, 1), To: at(1, 4), Weight: one}
	for _, tt := range []struct {
		name      string
		m         tempomap.Map
		terms     []tempomap.WeightedInterval
		step      *big.Rat
		mode      rounding.Mode
		grids     []measure.Grid
		own       error
		deps, not []error
	}{
		{name: "unset map before other inputs", terms: []tempomap.WeightedInterval{{}}, own: tempomap.ErrUnset},
		{name: "all addresses before weights", m: m, terms: []tempomap.WeightedInterval{{From: at(0, 1), To: at(1, 4)}, {}}, own: tempomap.ErrUnset},
		{name: "unset from", m: m, terms: []tempomap.WeightedInterval{{To: at(1, 4), Weight: one}}, step: one, mode: rounding.Down, own: tempomap.ErrUnset},
		{name: "unset to", m: m, terms: []tempomap.WeightedInterval{{From: at(0, 1), Weight: one}}, step: one, mode: rounding.Down, own: tempomap.ErrUnset},
		{name: "nil weight", m: m, terms: []tempomap.WeightedInterval{{From: at(0, 1), To: at(1, 4)}}, step: one, mode: rounding.Down, own: tempomap.ErrInvalidArgument, not: []error{fraction.ErrInvalid, fraction.ErrOutOfRange}},
		{name: "zero weight even on zero interval", m: m, terms: []tempomap.WeightedInterval{{From: at(0, 1), To: at(0, 1), Weight: big.NewRat(0, 1)}}, step: one, mode: rounding.Down, own: tempomap.ErrInvalidArgument},
		{name: "negative weight", m: m, terms: []tempomap.WeightedInterval{{From: at(0, 1), To: at(1, 4), Weight: big.NewRat(-1, 1)}}, step: one, mode: rounding.Down, own: tempomap.ErrInvalidArgument},
		{name: "weight before unset grid", m: measured, terms: []tempomap.WeightedInterval{{From: bar(1, 0, 1), To: bar(1, 1, 4)}}, grids: []measure.Grid{{}}, own: tempomap.ErrInvalidArgument, not: []error{measure.ErrUnset, rounding.ErrUnset}},
		{name: "positional map rejects grid", m: m, terms: []tempomap.WeightedInterval{valid}, step: one, mode: rounding.Down, grids: []measure.Grid{grid}, own: tempomap.ErrInvalidArgument},
		{name: "too many grids", m: measured, step: one, mode: rounding.Down, grids: []measure.Grid{grid, grid}, own: tempomap.ErrInvalidArgument},
		{name: "measured map requires grid even when empty", m: measured, step: one, mode: rounding.Down, own: tempomap.ErrInvalidArgument},
		{name: "grid before mode on empty list", m: measured, step: one, grids: []measure.Grid{{}}, own: tempomap.ErrInvalidArgument, deps: []error{measure.ErrUnset}, not: []error{rounding.ErrUnset}},
		{name: "nil step", m: m, terms: []tempomap.WeightedInterval{valid}, mode: rounding.Down, own: tempomap.ErrInvalidArgument, not: []error{fraction.ErrInvalid}},
		{name: "zero step", m: m, step: big.NewRat(0, 1), mode: rounding.Down, own: tempomap.ErrInvalidArgument},
		{name: "negative step", m: m, step: big.NewRat(-1, 1), mode: rounding.Down, own: tempomap.ErrInvalidArgument},
		{name: "step before mode", m: m, own: tempomap.ErrInvalidArgument, not: []error{rounding.ErrUnset}},
		{name: "unset mode even when empty", m: m, step: one, own: tempomap.ErrInvalidArgument, deps: []error{rounding.ErrUnset}},
		{name: "unknown mode", m: m, terms: []tempomap.WeightedInterval{valid}, step: one, mode: rounding.Mode(99), own: tempomap.ErrInvalidArgument, deps: []error{rounding.ErrUnknownMode}},
		{name: "mode before resolving addresses", m: measured, terms: []tempomap.WeightedInterval{{From: bar(2, 0, 1), To: bar(2, 1, 4), Weight: one}}, step: one, grids: []measure.Grid{grid}, own: tempomap.ErrInvalidArgument, deps: []error{rounding.ErrUnset}, not: []error{measure.ErrOutsideGrid}},
		{name: "mismatched addressing", m: m, terms: []tempomap.WeightedInterval{{From: bar(1, 0, 1), To: bar(1, 1, 4), Weight: one}}, step: one, mode: rounding.Down, own: tempomap.ErrInvalidArgument},
		{name: "from outside grid", m: measured, terms: []tempomap.WeightedInterval{{From: bar(2, 0, 1), To: bar(1, 1, 4), Weight: one}}, step: one, mode: rounding.Down, grids: []measure.Grid{grid}, own: tempomap.ErrInvalidAddress, deps: []error{measure.ErrOutsideGrid}},
		{name: "to outside grid", m: measured, terms: []tempomap.WeightedInterval{{From: bar(1, 0, 1), To: bar(2, 0, 1), Weight: one}}, step: one, mode: rounding.Down, grids: []measure.Grid{grid}, own: tempomap.ErrInvalidAddress, deps: []error{measure.ErrOutsideGrid}},
		{name: "reverse interval", m: m, terms: []tempomap.WeightedInterval{{From: at(1, 4), To: at(0, 1), Weight: one}}, step: one, mode: rounding.Down, own: tempomap.ErrInvalidArgument},
		{name: "unfinished interval", m: unfinished, terms: []tempomap.WeightedInterval{{From: at(0, 1), To: at(2, 1), Weight: one}}, step: one, mode: rounding.Down, own: tempomap.ErrUnfinished},
		{name: "zero interval in unfinished tail", m: unfinished, terms: []tempomap.WeightedInterval{{From: at(2, 1), To: at(2, 1), Weight: one}}, step: one, mode: rounding.Down, own: tempomap.ErrUnfinished},
		{name: "all interval order checks before unfinished", m: unfinished, terms: []tempomap.WeightedInterval{{From: at(0, 1), To: at(2, 1), Weight: one}, {From: at(1, 4), To: at(0, 1), Weight: one}}, step: one, mode: rounding.Down, own: tempomap.ErrInvalidArgument, not: []error{tempomap.ErrUnfinished}},
		{name: "final index exceeds int", m: m, terms: []tempomap.WeightedInterval{{From: at(0, 1), To: at(1, 4), Weight: new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 63))}}, step: one, mode: rounding.Down, own: tempomap.ErrOutOfRange, deps: []error{fraction.ErrNotRepresentable}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.m.WeightedBetweenSteps(tt.terms, tt.step, tt.mode, tt.grids...)
			requireTempomapError(t, err, tt.own, tt.deps...)
			assert.Zero(t, got)
			for _, absent := range tt.not {
				assert.NotErrorIs(t, err, absent)
			}
			requirePanicError(t, err, tt.own, func() { tt.m.MustWeightedBetweenSteps(tt.terms, tt.step, tt.mode, tt.grids...) })
		})
	}
	// Empty terms do not touch an unfinished part of the map.
	assert.Zero(t, unfinished.MustWeightedBetweenSteps(nil, one, rounding.Down))
	assert.Equal(t, 1, unfinished.MustWeightedBetweenSteps([]tempomap.WeightedInterval{valid}, one, rounding.Down))
	assert.Zero(t, measured.MustWeightedBetweenSteps(nil, one, rounding.Down, grid))
	assert.Equal(t, 1, measured.MustWeightedBetweenSteps([]tempomap.WeightedInterval{{From: bar(1, 0, 1), To: bar(1, 1, 4), Weight: one}}, one, rounding.Down, grid))
	// A valid measured address can resolve outside fraction even when its
	// offset and the meter fit individually: their denominators multiply.
	max := fraction.MaxComponentBig()
	offset := duration.MustPositionFromFraction(fraction.MustFromBigInt(new(big.Int).Sub(max, big.NewInt(1)), max))
	wideGrid := measure.MustNew(meter.MustNew(1<<62, (1<<62)-1))
	wideMap := tempomap.MustNewMeasured(wideGrid, quarter(60))
	a := tempomap.MustAtBar(measure.MustNewAddress(2, offset))
	k, err := wideMap.WeightedBetweenSteps([]tempomap.WeightedInterval{{From: a, To: a, Weight: one}}, one, rounding.Down, wideGrid)
	requireTempomapError(t, err, tempomap.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.Zero(t, k)
}

func TestWeightedBetweenStepsOwnershipAndConcurrency(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60))
	before := m.String()
	weight, step := big.NewRat(1, 3), big.NewRat(2, 3)
	term := tempomap.WeightedInterval{From: at(0, 1), To: at(1, 4), Weight: weight}
	terms := []tempomap.WeightedInterval{term, term, term}
	copyOfTerm := term
	assert.True(t, term == copyOfTerm, "copy shares the weight pointer")
	numericallyEqual := term
	numericallyEqual.Weight = big.NewRat(1, 3)
	assert.False(t, term == numericallyEqual, "pointer identity is not numerical weight equality")
	assert.Zero(t, term.Weight.Cmp(numericallyEqual.Weight))

	got := m.MustWeightedBetweenSteps(terms, step, rounding.HalfEven)
	assert.Equal(t, 2, got)
	for _, input := range terms {
		assert.True(t, input.From.Equal(term.From))
		assert.True(t, input.To.Equal(term.To))
		assert.Same(t, weight, input.Weight)
		assert.Equal(t, "1/3", input.Weight.RatString())
	}
	assert.Equal(t, "2/3", step.RatString())
	_, err := m.WeightedBetweenSteps(terms, step, rounding.Mode(0))
	requireTempomapError(t, err, tempomap.ErrInvalidArgument, rounding.ErrUnset)
	assert.Equal(t, "1/3", weight.RatString())
	assert.Equal(t, "2/3", step.RatString())
	assert.Equal(t, before, m.String())

	// The same read-only slice and big.Rats can be shared by callers.
	results := make(chan int, 16)
	for range 16 {
		go func() { results <- m.MustWeightedBetweenSteps(terms, step, rounding.HalfEven) }()
	}
	for range 16 {
		assert.Equal(t, 2, <-results)
	}
	// All workers have joined before the caller mutates its records.
	weight.SetInt64(3)
	step.SetInt64(2)
	terms[0] = tempomap.WeightedInterval{}
	assert.Equal(t, "3", copyOfTerm.Weight.RatString())
	assert.Equal(t, 2, got, "returned index is independent of later input mutation")
	assert.Equal(t, before, m.String())
	assert.Equal(t, 1, m.MustWeightedBetweenSteps([]tempomap.WeightedInterval{numericallyEqual}, big.NewRat(1, 3), rounding.Down))
}

func TestWeightedBetweenStepsIndexBoundary(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarter(60))
	max := int(^uint(0) >> 1)
	weight := new(big.Rat).SetInt64(int64(max))
	terms := []tempomap.WeightedInterval{{From: at(0, 1), To: at(1, 4), Weight: weight}}
	assert.Equal(t, max, m.MustWeightedBetweenSteps(terms, big.NewRat(1, 1), rounding.Up))
	// The exact sum can exceed int before rounding; a final index that
	// still fits is accepted. max+1/2 rounds down to max, up to max+1.
	weight.Add(weight, big.NewRat(1, 2))
	assert.Equal(t, max, m.MustWeightedBetweenSteps(terms, big.NewRat(1, 1), rounding.Down))
	got, err := m.WeightedBetweenSteps(terms, big.NewRat(1, 1), rounding.Up)
	requireTempomapError(t, err, tempomap.ErrOutOfRange, fraction.ErrNotRepresentable)
	assert.Zero(t, got)
}
