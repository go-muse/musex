package ticks

import (
	"math/big"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
)

func checkPlaces(op string, places int) error {
	if places < 0 || places > tempomap.MaxPlaces {
		return invalidError(op, "decimal places outside 0..1000")
	}
	return nil
}

// positionAddress returns the positional address of tick ticks of r.
func (r Resolution) positionAddress(op string, tick int) (tempomap.Address, error) {
	p, err := positionOf(op, new(big.Rat).Mul(ratInt(tick), r.tick()))
	if err != nil {
		return tempomap.Address{}, err
	}
	return tempomap.MustAt(p), nil
}

func (r Resolution) momentSeconds(op string, m tempomap.Map, tick, places int, mode rounding.Mode, grid []measure.Grid, end bool) (tempomap.SecondsResult, error) {
	if err := r.check(op); err != nil {
		return tempomap.SecondsResult{}, err
	}
	if err := checkMusexMap(op, m); err != nil {
		return tempomap.SecondsResult{}, err
	}
	if err := checkTick(op, tick); err != nil {
		return tempomap.SecondsResult{}, err
	}
	if err := checkPlaces(op, places); err != nil {
		return tempomap.SecondsResult{}, err
	}
	if err := checkMode(op, mode); err != nil {
		return tempomap.SecondsResult{}, err
	}
	pm, err := positionalMap(op, m, grid)
	if err != nil {
		return tempomap.SecondsResult{}, err
	}
	address, err := r.positionAddress(op, tick)
	if err != nil {
		return tempomap.SecondsResult{}, err
	}
	var result tempomap.SecondsResult
	if end {
		result, err = pm.End(address, places, mode)
	} else {
		result, err = pm.Start(address, places, mode)
	}
	if err != nil {
		return tempomap.SecondsResult{}, dependencyError(op, err)
	}
	return result, nil
}

// StartSeconds returns the seconds at the start of tick by the musex tempo map
// m: m.Start at the position of tick, with the decimal places and the mode of
// tempomap. A measure map takes exactly one grid.
func (r Resolution) StartSeconds(m tempomap.Map, tick, places int, mode rounding.Mode, grid ...measure.Grid) (tempomap.SecondsResult, error) {
	return r.momentSeconds("Resolution.StartSeconds", m, tick, places, mode, grid, false)
}

// MustStartSeconds is like StartSeconds but panics with its error.
func (r Resolution) MustStartSeconds(m tempomap.Map, tick, places int, mode rounding.Mode, grid ...measure.Grid) tempomap.SecondsResult {
	return must(r.StartSeconds(m, tick, places, mode, grid...))
}

// EndSeconds returns the seconds at the end of the hold at tick: m.End.
func (r Resolution) EndSeconds(m tempomap.Map, tick, places int, mode rounding.Mode, grid ...measure.Grid) (tempomap.SecondsResult, error) {
	return r.momentSeconds("Resolution.EndSeconds", m, tick, places, mode, grid, true)
}

// MustEndSeconds is like EndSeconds but panics with its error.
func (r Resolution) MustEndSeconds(m tempomap.Map, tick, places int, mode rounding.Mode, grid ...measure.Grid) tempomap.SecondsResult {
	return must(r.EndSeconds(m, tick, places, mode, grid...))
}

// SecondsBetween returns m.Between the positions of from and to; from > to is
// ErrInvalidArgument.
func (r Resolution) SecondsBetween(m tempomap.Map, from, to, places int, mode rounding.Mode, grid ...measure.Grid) (tempomap.SecondsResult, error) {
	const op = "Resolution.SecondsBetween"
	if err := r.check(op); err != nil {
		return tempomap.SecondsResult{}, err
	}
	if err := checkMusexMap(op, m); err != nil {
		return tempomap.SecondsResult{}, err
	}
	if err := checkTick(op, from); err != nil {
		return tempomap.SecondsResult{}, err
	}
	if err := checkTick(op, to); err != nil {
		return tempomap.SecondsResult{}, err
	}
	if err := checkPlaces(op, places); err != nil {
		return tempomap.SecondsResult{}, err
	}
	if err := checkMode(op, mode); err != nil {
		return tempomap.SecondsResult{}, err
	}
	pm, err := positionalMap(op, m, grid)
	if err != nil {
		return tempomap.SecondsResult{}, err
	}
	if from > to {
		return tempomap.SecondsResult{}, invalidError(op, "from follows to")
	}
	fromAddress, err := r.positionAddress(op, from)
	if err != nil {
		return tempomap.SecondsResult{}, err
	}
	toAddress, err := r.positionAddress(op, to)
	if err != nil {
		return tempomap.SecondsResult{}, err
	}
	result, err := pm.Between(fromAddress, toAddress, places, mode)
	if err != nil {
		return tempomap.SecondsResult{}, dependencyError(op, err)
	}
	return result, nil
}

// MustSecondsBetween is like SecondsBetween but panics with its error.
func (r Resolution) MustSecondsBetween(m tempomap.Map, from, to, places int, mode rounding.Mode, grid ...measure.Grid) tempomap.SecondsResult {
	return must(r.SecondsBetween(m, from, to, places, mode, grid...))
}

// positionSteps returns mode(P/step) for the exact position P at seconds of a
// positional map; inside a hold P is the position of the hold.
func positionSteps(op string, pm tempomap.Map, seconds tempo.Seconds, step duration.Duration, mode rounding.Mode) (int, error) {
	k, err := pm.PositionSteps(seconds, step, mode)
	if err != nil {
		return 0, dependencyError(op, err)
	}
	return k, nil
}

// TickAt returns the tick at seconds by the musex tempo map m: the exact
// position rounded once with mode to whole ticks, correctly also on gradual
// segments. Inside a hold the position of the hold is used.
func (r Resolution) TickAt(m tempomap.Map, s tempo.Seconds, mode rounding.Mode, grid ...measure.Grid) (int, error) {
	const op = "Resolution.TickAt"
	if err := r.check(op); err != nil {
		return 0, err
	}
	if err := checkMusexMap(op, m); err != nil {
		return 0, err
	}
	if err := secondsArgument(op, s); err != nil {
		return 0, err
	}
	if err := checkMode(op, mode); err != nil {
		return 0, err
	}
	pm, err := positionalMap(op, m, grid)
	if err != nil {
		return 0, err
	}
	step, err := durationOf(op, r.tick())
	if err != nil {
		return 0, err
	}
	return positionSteps(op, pm, s, step, mode)
}

// MustTickAt is like TickAt but panics with its error.
func (r Resolution) MustTickAt(m tempomap.Map, s tempo.Seconds, mode rounding.Mode, grid ...measure.Grid) int {
	return must(r.TickAt(m, s, mode, grid...))
}

// smpteStep returns the length of one SMPTE tick in seconds.
func (s SMPTE) smpteStep(op string) (tempo.Seconds, error) {
	tps, err := s.ticksPerSecond(op)
	if err != nil {
		return tempo.Seconds{}, err
	}
	return secondsOf(op, tps.Inv(tps))
}

func (s SMPTE) momentTick(op string, m tempomap.Map, p duration.Position, mode rounding.Mode, grid []measure.Grid, end bool) (int, error) {
	if err := s.check(op); err != nil {
		return 0, err
	}
	if err := checkMusexMap(op, m); err != nil {
		return 0, err
	}
	if err := checkPosition(op, p); err != nil {
		return 0, err
	}
	if err := checkMode(op, mode); err != nil {
		return 0, err
	}
	pm, err := positionalMap(op, m, grid)
	if err != nil {
		return 0, err
	}
	step, err := s.smpteStep(op)
	if err != nil {
		return 0, err
	}
	var k int
	if end {
		k, err = pm.EndSteps(tempomap.MustAt(p), step, mode)
	} else {
		k, err = pm.StartSteps(tempomap.MustAt(p), step, mode)
	}
	if err != nil {
		return 0, dependencyError(op, err)
	}
	return k, nil
}

// StartTick returns the SMPTE tick at the start of the position p by the musex
// tempo map m: the exact seconds times the SMPTE ticks per second, rounded once
// with mode. Code -29 without a reading is ErrInvalidArgument.
func (s SMPTE) StartTick(m tempomap.Map, p duration.Position, mode rounding.Mode, grid ...measure.Grid) (int, error) {
	return s.momentTick("SMPTE.StartTick", m, p, mode, grid, false)
}

// MustStartTick is like StartTick but panics with its error.
func (s SMPTE) MustStartTick(m tempomap.Map, p duration.Position, mode rounding.Mode, grid ...measure.Grid) int {
	return must(s.StartTick(m, p, mode, grid...))
}

// EndTick is StartTick at the end of the hold at p.
func (s SMPTE) EndTick(m tempomap.Map, p duration.Position, mode rounding.Mode, grid ...measure.Grid) (int, error) {
	return s.momentTick("SMPTE.EndTick", m, p, mode, grid, true)
}

// MustEndTick is like EndTick but panics with its error.
func (s SMPTE) MustEndTick(m tempomap.Map, p duration.Position, mode rounding.Mode, grid ...measure.Grid) int {
	return must(s.EndTick(m, p, mode, grid...))
}

// PositionAt returns the position at the SMPTE tick by the musex tempo map m:
// m.Position of its exact seconds, with the decimal places and the mode of
// tempomap.
func (s SMPTE) PositionAt(m tempomap.Map, tick, places int, mode rounding.Mode, grid ...measure.Grid) (tempomap.PositionResult, error) {
	const op = "SMPTE.PositionAt"
	if err := s.check(op); err != nil {
		return tempomap.PositionResult{}, err
	}
	if err := checkMusexMap(op, m); err != nil {
		return tempomap.PositionResult{}, err
	}
	if err := checkTick(op, tick); err != nil {
		return tempomap.PositionResult{}, err
	}
	if err := checkPlaces(op, places); err != nil {
		return tempomap.PositionResult{}, err
	}
	if err := checkMode(op, mode); err != nil {
		return tempomap.PositionResult{}, err
	}
	pm, err := positionalMap(op, m, grid)
	if err != nil {
		return tempomap.PositionResult{}, err
	}
	seconds, err := s.Seconds(tick)
	if err != nil {
		return tempomap.PositionResult{}, err
	}
	result, err := pm.Position(seconds, places, mode)
	if err != nil {
		return tempomap.PositionResult{}, dependencyError(op, err)
	}
	return result, nil
}

// MustPositionAt is like PositionAt but panics with its error.
func (s SMPTE) MustPositionAt(m tempomap.Map, tick, places int, mode rounding.Mode, grid ...measure.Grid) tempomap.PositionResult {
	return must(s.PositionAt(m, tick, places, mode, grid...))
}

// MetricTick returns the metric tick of r at the SMPTE tick by the musex tempo
// map m: the exact position of its seconds rounded once with mode.
func (s SMPTE) MetricTick(m tempomap.Map, r Resolution, tick int, mode rounding.Mode, grid ...measure.Grid) (int, error) {
	const op = "SMPTE.MetricTick"
	if err := s.check(op); err != nil {
		return 0, err
	}
	if err := checkMusexMap(op, m); err != nil {
		return 0, err
	}
	if !r.IsSet() {
		return 0, unsetError(op, "resolution")
	}
	if err := checkTick(op, tick); err != nil {
		return 0, err
	}
	if err := checkMode(op, mode); err != nil {
		return 0, err
	}
	pm, err := positionalMap(op, m, grid)
	if err != nil {
		return 0, err
	}
	seconds, err := s.Seconds(tick)
	if err != nil {
		return 0, err
	}
	step, err := durationOf(op, r.tick())
	if err != nil {
		return 0, err
	}
	return positionSteps(op, pm, seconds, step, mode)
}

// MustMetricTick is like MetricTick but panics with its error.
func (s SMPTE) MustMetricTick(m tempomap.Map, r Resolution, tick int, mode rounding.Mode, grid ...measure.Grid) int {
	return must(s.MetricTick(m, r, tick, mode, grid...))
}
