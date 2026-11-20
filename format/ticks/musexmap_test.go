package ticks_test

import (
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Expected values of this file on the gradual change of КП-15 (quarter = 120
// slowing linearly to quarter = 100 over [0, 1)) were computed outside Go
// with 80 significant decimal digits (Python decimal) from the exact formulas
// of tempomap 4.7 (ИК-20, ИК-21): the speed in whole notes per minute is
// s(p) = 30 − 5p, so
//
//	B(p) = 12·ln(30/(30 − 5p)) seconds for 0 ≤ p ≤ 1,
//	P(T) = 6·(1 − e^(−T/12)) for 0 ≤ T ≤ 12·ln(6/5),
//	after position 1 the speed is 25 whole notes per minute.
//
// Values used (D = 480, w = 1/4, so a position p is 1920·p ticks; SMPTE −25×40
// has 1000 ticks per second):
//
//	1920·P(1)                 = 921.0883434701961842…
//	1920·P(2)                 = 1768.5305292601258670…
//	1920·(1 + 25·(3 − B(1))/60) = 2569.7130547780355883…
//	1000·B(1/4)               = 510.7153730255511460…
//	1000·B(1/2)               = 1044.1365238755571940…
//	1000·B(1)                 = 2187.8586815274555145…
//	1000·(B(1) + 1/2·60/25)   = 3387.8586815274555145…
//
// None of them is within 10^−2 of a rounding boundary, so the nearest modes
// agree and the directed modes take the lower or upper neighbor.

// TestSecondsByMusexMap checks О-19, ИТ-11 and 4.11.
func TestSecondsByMusexMap(t *testing.T) {
	t.Parallel()
	r := res(480)
	rit := ritardandoMap()
	got, err := r.StartSeconds(rit, 1920, 6, rounding.HalfUp)
	require.NoError(t, err)
	assertEqualValue(t, rit.MustStart(at(1, 1), 6, rounding.HalfUp), got)
	value, err := got.Value()
	require.NoError(t, err)
	assertSeconds(t, sec(2187859, 1000000), value)
	approximate, err := got.Approximate()
	require.NoError(t, err)
	assert.True(t, approximate)
	assertEqualValue(t, got, r.MustStartSeconds(rit, 1920, 6, rounding.HalfUp))
	for _, mode := range allModes {
		for _, tick := range []int{0, 240, 960, 1920, 2880} {
			p := r.MustPosition(tick)
			assertEqualValue(t, rit.MustStart(tempomap.MustAt(p), 9, mode), r.MustStartSeconds(rit, tick, 9, mode))
			assertEqualValue(t, rit.MustEnd(tempomap.MustAt(p), 9, mode), r.MustEndSeconds(rit, tick, 9, mode))
		}
		assertEqualValue(t, rit.MustBetween(at(1, 8), at(3, 2), 7, mode), r.MustSecondsBetween(rit, 240, 2880, 7, mode))
	}

	hold := holdMap(pos(1, 1), sec(1, 1))
	start := r.MustStartSeconds(hold, 1920, 3, rounding.HalfUp)
	assertSeconds(t, sec(4, 1), start.MustValue())
	assert.False(t, start.MustApproximate())
	end, err := r.EndSeconds(hold, 1920, 3, rounding.HalfUp)
	require.NoError(t, err)
	assertSeconds(t, sec(5, 1), end.MustValue())
	assertSeconds(t, sec(5, 1), r.MustEndSeconds(hold, 1920, 3, rounding.HalfUp).MustValue())
	between, err := r.SecondsBetween(hold, 0, 1920, 3, rounding.HalfUp)
	require.NoError(t, err)
	assertSeconds(t, sec(4, 1), between.MustValue())
	assertSeconds(t, sec(2, 1), r.MustSecondsBetween(hold, 1920, 2400, 3, rounding.HalfUp).MustValue())
	assertSeconds(t, tempo.ZeroSeconds(), r.MustSecondsBetween(hold, 960, 960, 3, rounding.HalfUp).MustValue())
	// The written quarter: tick 960 with w = 1/2 is position 1.
	assertSeconds(t, sec(4, 1), resQ(480, 16).MustStartSeconds(hold, 960, 0, rounding.Down).MustValue())

	// A measured map needs exactly one grid.
	grid := grid44()
	measured := tempomap.MustNewMeasured(grid, quarterTempo(60, 1))
	assertSeconds(t, sec(4, 1), r.MustStartSeconds(measured, 1920, 3, rounding.HalfUp, grid).MustValue())
	assertSeconds(t, sec(6, 1), r.MustEndSeconds(measured, 2880, 3, rounding.HalfUp, grid).MustValue())
	assertSeconds(t, sec(2, 1), r.MustSecondsBetween(measured, 1920, 2880, 3, rounding.HalfUp, grid).MustValue())
}

// TestTickAtByMusexMap checks О-19, ИТ-11, ТК-13 through the public API.
func TestTickAtByMusexMap(t *testing.T) {
	t.Parallel()
	r := res(480)
	rit := ritardandoMap()
	for _, mode := range allModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			n, err := r.TickAt(rit, sec(1, 1), mode)
			require.NoError(t, err)
			assert.Equal(t, int(irrationalIndex(mode, 921, false)), n)
			assert.Equal(t, int(irrationalIndex(mode, 1768, true)), r.MustTickAt(rit, sec(2, 1), mode))
			assert.Equal(t, int(irrationalIndex(mode, 2569, true)), r.MustTickAt(rit, sec(3, 1), mode))
			// The value agrees with PositionSteps of tempomap with the tick length.
			assert.Equal(t, rit.MustPositionSteps(sec(1, 1), dur(1, 1920), mode), r.MustTickAt(rit, sec(1, 1), mode))
			// Inside a hold the position is the position of the hold.
			hold := holdMap(pos(1, 1), sec(1, 1))
			assert.Equal(t, 1920, r.MustTickAt(hold, sec(9, 2), mode))
			assert.Equal(t, 1920, r.MustTickAt(hold, sec(5, 1), mode))
			// Exact values: quarter = 60 makes a tick at D = 1 last one second.
			assert.Equal(t, int(tieIndex(mode, 1, false)), res(1).MustTickAt(constantMap(60), sec(3, 2), mode))
			assert.Equal(t, int(tieIndex(mode, 2, true)), res(1).MustTickAt(constantMap(60), sec(5, 2), mode))
			assert.Equal(t, 480, r.MustTickAt(constantMap(60), sec(1, 1), mode))
			assert.Equal(t, 0, r.MustTickAt(rit, tempo.ZeroSeconds(), mode))
		})
	}
	// A measured map with its grid.
	grid := grid44()
	measured := tempomap.MustNewMeasured(grid, quarterTempo(60, 1))
	assert.Equal(t, 2400, r.MustTickAt(measured, sec(5, 1), rounding.Down, grid))
	// The written quarter: 1 s at quarter = 60 is position 1/4, 240 ticks with w = 1/2.
	assert.Equal(t, 240, resQ(480, 16).MustTickAt(constantMap(60), sec(1, 1), rounding.Down))
}

// TestSMPTEByMusexMap checks О-21, ИТ-11, ТК-12, ТК-13, КП-18.
func TestSMPTEByMusexMap(t *testing.T) {
	t.Parallel()
	s := smpte(-25, 40)
	// КП-18.
	q120 := constantMap(120)
	n, err := s.StartTick(q120, pos(1, 2), rounding.HalfUp)
	require.NoError(t, err)
	assert.Equal(t, 1000, n)
	p, err := s.PositionAt(q120, 1000, 3, rounding.HalfUp)
	require.NoError(t, err)
	assertEqualValue(t, pos(1, 2), p.MustValue())
	assert.False(t, p.MustApproximate())
	assertEqualValue(t, pos(1, 2), s.MustPositionAt(q120, 1000, 3, rounding.HalfUp).MustValue())
	metric, err := s.MetricTick(q120, res(480), 1000, rounding.HalfUp)
	require.NoError(t, err)
	assert.Equal(t, 960, metric)
	assert.Equal(t, 960, s.MustMetricTick(q120, res(480), 1000, rounding.HalfUp))
	assert.Equal(t, 1000, s.MustStartTick(q120, res(480).MustPosition(960), rounding.Down))

	rit := ritardandoMap()
	ritHold := rit.MustWithExtension(tempomap.MustNewExtension(at(1, 1), sec(1, 1)))
	for _, mode := range allModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, int(irrationalIndex(mode, 510, true)), s.MustStartTick(rit, pos(1, 4), mode))
			assert.Equal(t, int(irrationalIndex(mode, 1044, false)), s.MustStartTick(rit, pos(1, 2), mode))
			assert.Equal(t, int(irrationalIndex(mode, 2187, true)), s.MustStartTick(rit, pos(1, 1), mode))
			assert.Equal(t, int(irrationalIndex(mode, 3387, true)), s.MustStartTick(rit, pos(3, 2), mode))
			// End adds the hold of 1 s: 3187.8586…
			end, err := s.EndTick(ritHold, pos(1, 1), mode)
			require.NoError(t, err)
			assert.Equal(t, int(irrationalIndex(mode, 3187, true)), end)
			assert.Equal(t, int(irrationalIndex(mode, 2187, true)), s.MustEndTick(rit, pos(1, 1), mode))
			assert.Equal(t, rit.MustStartSteps(at(1, 1), sec(1, 1000), mode), s.MustStartTick(rit, pos(1, 1), mode))
			// SMPTE tick 1000 is 1 s; 2000 is 2 s.
			assert.Equal(t, int(irrationalIndex(mode, 921, false)), s.MustMetricTick(rit, res(480), 1000, mode))
			assert.Equal(t, int(irrationalIndex(mode, 1768, true)), s.MustMetricTick(rit, res(480), 2000, mode))
			assertEqualValue(t, rit.MustPosition(sec(1, 1), 12, mode), s.MustPositionAt(rit, 1000, 12, mode))
		})
	}
	// −29 needs a reading.
	df := smpte(-29, 80)
	_, err = df.StartTick(q120, pos(1, 2), rounding.Down)
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	assert.Equal(t, 2400, df.MustWithReading(ticks.Reading30).MustStartTick(q120, pos(1, 2), rounding.Down))
	assert.Equal(t, 2397, df.MustWithReading(ticks.ReadingDropFrame).MustStartTick(q120, pos(1, 2), rounding.Down))

	// A measured map with its grid.
	grid := grid44()
	measured := tempomap.MustNewMeasured(grid, quarterTempo(60, 1))
	assert.Equal(t, 4000, s.MustStartTick(measured, pos(1, 1), rounding.Down, grid))
	assert.Equal(t, 1920, s.MustMetricTick(measured, res(480), 4000, rounding.Down, grid))
}

// TestMusexMapErrors checks the failures of 4.11 and their mapping (4.16).
func TestMusexMapErrors(t *testing.T) {
	t.Parallel()
	r := res(480)
	s := smpte(-25, 40)
	rit := ritardandoMap()
	grid := grid44()
	measured := tempomap.MustNewMeasured(grid, quarterTempo(60, 1))
	for _, tt := range []struct {
		name string
		call func() error
		own  error
		deps []error
		not  []error
	}{
		{name: "unset receiver", call: func() error { _, err := ticks.Resolution{}.StartSeconds(rit, 0, 3, rounding.Down); return err }, own: ticks.ErrUnset},
		{name: "unset map", call: func() error { _, err := r.StartSeconds(tempomap.Map{}, 0, 3, rounding.Down); return err }, own: ticks.ErrInvalidArgument, deps: []error{tempomap.ErrUnset}},
		{name: "negative tick", call: func() error { _, err := r.EndSeconds(rit, -1, 3, rounding.Down); return err }, own: ticks.ErrInvalidArgument},
		{name: "reverse interval", call: func() error { _, err := r.SecondsBetween(rit, 960, 480, 3, rounding.Down); return err }, own: ticks.ErrInvalidArgument},
		{name: "unset mode", call: func() error { _, err := r.StartSeconds(rit, 0, 3, rounding.Mode(0)); return err }, own: ticks.ErrInvalidArgument, deps: []error{rounding.ErrUnset}},
		{name: "places outside tempomap", call: func() error { _, err := r.StartSeconds(rit, 0, -1, rounding.Down); return err }, own: ticks.ErrInvalidArgument},
		{name: "approximation with 77 places", call: func() error { _, err := r.StartSeconds(rit, 1920, 77, rounding.Down); return err }, own: ticks.ErrOutOfRange, deps: []error{tempomap.ErrOutOfRange}},
		{name: "unfinished map", call: func() error { _, err := r.StartSeconds(unfinishedMap(), 3840, 3, rounding.Down); return err }, own: ticks.ErrNotConvertible, deps: []error{tempomap.ErrUnfinished}},
		{name: "positional map with a grid", call: func() error { _, err := r.StartSeconds(rit, 0, 3, rounding.Down, grid); return err }, own: ticks.ErrInvalidArgument},
		{name: "measured map without a grid", call: func() error { _, err := r.StartSeconds(measured, 0, 3, rounding.Down); return err }, own: ticks.ErrInvalidArgument},
		{name: "measured map with two grids", call: func() error { _, err := r.TickAt(measured, sec(1, 1), rounding.Down, grid, grid); return err }, own: ticks.ErrInvalidArgument},
		{name: "unset grid", call: func() error { _, err := r.TickAt(measured, sec(1, 1), rounding.Down, measure.Grid{}); return err }, own: ticks.ErrInvalidArgument, deps: []error{measure.ErrUnset}},
		{name: "unset seconds", call: func() error { _, err := r.TickAt(rit, tempo.Seconds{}, rounding.Down); return err }, own: ticks.ErrInvalidArgument, deps: []error{tempo.ErrUnset}},
		{name: "tick outside int", call: func() error {
			_, err := r.TickAt(constantMap(60), tempo.MustSecondsFromFraction(hugeFraction(62)), rounding.Down)
			return err
		}, own: ticks.ErrOutOfRange},
		{name: "unfinished map tick", call: func() error { _, err := r.TickAt(unfinishedMap(), sec(10, 1), rounding.Down); return err }, own: ticks.ErrNotConvertible, deps: []error{tempomap.ErrUnfinished}},
		{name: "SMPTE unset position", call: func() error { _, err := s.StartTick(rit, duration.Position{}, rounding.Down); return err }, own: ticks.ErrInvalidArgument, deps: []error{duration.ErrUnset}},
		{name: "SMPTE unset map", call: func() error { _, err := s.EndTick(tempomap.Map{}, pos(1, 1), rounding.Down); return err }, own: ticks.ErrInvalidArgument, deps: []error{tempomap.ErrUnset}},
		{name: "SMPTE negative tick", call: func() error { _, err := s.PositionAt(rit, -1, 3, rounding.Down); return err }, own: ticks.ErrInvalidArgument},
		{name: "SMPTE unset resolution", call: func() error { _, err := s.MetricTick(rit, ticks.Resolution{}, 1, rounding.Down); return err }, own: ticks.ErrUnset},
		{name: "SMPTE unset receiver", call: func() error { _, err := ticks.SMPTE{}.StartTick(rit, pos(1, 1), rounding.Down); return err }, own: ticks.ErrUnset},
		{name: "SMPTE tick outside int", call: func() error {
			_, err := s.StartTick(constantMap(60), duration.Origin().MustAdd(duration.MustFromFraction(hugeFraction(60))), rounding.Down)
			return err
		}, own: ticks.ErrOutOfRange},
		{name: "map is checked before the mode", call: func() error { _, err := r.StartSeconds(tempomap.Map{}, 0, 3, rounding.Mode(0)); return err }, own: ticks.ErrInvalidArgument, deps: []error{tempomap.ErrUnset}, not: []error{rounding.ErrUnset}},
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
	_, err := r.StartSeconds(tempomap.Map{}, 0, 3, rounding.Down)
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { r.MustStartSeconds(tempomap.Map{}, 0, 3, rounding.Down) })
	_, err = s.MetricTick(rit, ticks.Resolution{}, 1, rounding.Down)
	requireMustPanic(t, err, ticks.ErrUnset, func() { s.MustMetricTick(rit, ticks.Resolution{}, 1, rounding.Down) })
}
