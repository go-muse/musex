package tempo

import (
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
)

// The formula of bars, units, beats per minute, and time is n·r = N·t with
// r = L/u: n bars of length L, each holding r units of duration u, last t
// minutes at N beats per minute. It holds for a constant tempo and equal full
// bars; the package does not check this condition, since it knows no bars.
// Time is given in seconds, t = s/60. All results are exact; n may be
// fractional when the time does not end on a bar line.

// BarUnits returns r = L/u, the number of units of t in a bar of length
// barLen: 4 for a whole-note bar at a quarter unit. In order of checking, it
// returns an error wrapping ErrUnset if t is unset, ErrInvalidArgument if
// barLen is unset, and ErrOutOfRange together with fraction.ErrOutOfRange if
// the result is outside the range of package common/fraction.
func (t Tempo) BarUnits(barLen duration.Duration) (fraction.Fraction, error) {
	const op = "Tempo.BarUnits"
	if !t.IsSet() {
		return fraction.Fraction{}, t.unset(op)
	}
	if !barLen.IsSet() {
		return fraction.Fraction{}, invalidError(op, "unset bar length")
	}
	return toFraction(op, quo(durRat(barLen), unitRat(t.m.unit)))
}

// MustBarUnits is like BarUnits but panics with the returned error.
func (t Tempo) MustBarUnits(barLen duration.Duration) fraction.Fraction {
	return must(t.BarUnits(barLen))
}

// Bars returns n = N·t·u/L, the number of bars of length barLen in s seconds:
// 60 bars of a whole note in 120 seconds at ♩ = 120. In order of checking, it
// returns an error wrapping ErrUnset if t or then s is unset,
// ErrInvalidArgument if barLen is unset, ErrNoSolution if s is zero, and
// ErrOutOfRange together with fraction.ErrOutOfRange if the result is outside
// the range of package common/fraction.
func (t Tempo) Bars(barLen duration.Duration, s Seconds) (fraction.Fraction, error) {
	const op = "Tempo.Bars"
	if err := t.checkTime(op, s); err != nil {
		return fraction.Fraction{}, err
	}
	if !barLen.IsSet() {
		return fraction.Fraction{}, invalidError(op, "unset bar length")
	}
	if s.v.Sign() == 0 {
		return fraction.Fraction{}, simpleError(ErrNoSolution, op, "zero time")
	}
	return toFraction(op, quo(mul(t.m.speed(), minutes(s)), durRat(barLen)))
}

// MustBars is like Bars but panics with the returned error.
func (t Tempo) MustBars(barLen duration.Duration, s Seconds) fraction.Fraction {
	return must(t.Bars(barLen, s))
}

// SecondsOfBars returns s = 60·n·L/(u·N), the time of bars bars of length
// barLen. In order of checking, it returns an error wrapping ErrUnset if t is
// unset, ErrInvalidArgument if barLen or then bars is unset, ErrNoSolution if
// bars is zero or negative, and ErrOutOfRange together with
// fraction.ErrOutOfRange if the result is outside the range of package
// common/fraction.
func (t Tempo) SecondsOfBars(barLen duration.Duration, bars fraction.Fraction) (Seconds, error) {
	const op = "Tempo.SecondsOfBars"
	switch {
	case !t.IsSet():
		return Seconds{}, t.unset(op)
	case !barLen.IsSet():
		return Seconds{}, invalidError(op, "unset bar length")
	case unsetFraction(bars):
		return Seconds{}, invalidError(op, "unset number of bars")
	case bars.Sign() <= 0:
		return Seconds{}, simpleError(ErrNoSolution, op, "number of bars not positive")
	}
	return toSeconds(op, quo(mul(ratInt(60), ratOf(bars), durRat(barLen)), t.m.speed()))
}

// MustSecondsOfBars is like SecondsOfBars but panics with the returned error.
func (t Tempo) MustSecondsOfBars(barLen duration.Duration, bars fraction.Fraction) Seconds {
	return must(t.SecondsOfBars(barLen, bars))
}

// BarFor returns the bar in which bars bars last s seconds: r = N·t/n units
// and the bar length L = u·r. At ♩ = 60, 30 bars in 60 seconds give r = 2 and
// L = 1/2. In order of checking, it returns an error wrapping ErrUnset if t
// or then s is unset, ErrInvalidArgument if bars is unset, ErrNoSolution if
// bars is zero or negative and then if s is zero, and ErrOutOfRange together
// with fraction.ErrOutOfRange if r or then L is outside the range of package
// common/fraction.
func (t Tempo) BarFor(bars fraction.Fraction, s Seconds) (r fraction.Fraction, barLen duration.Duration, err error) {
	const op = "Tempo.BarFor"
	if err := t.checkTime(op, s); err != nil {
		return fraction.Fraction{}, duration.Duration{}, err
	}
	switch {
	case unsetFraction(bars):
		return fraction.Fraction{}, duration.Duration{}, invalidError(op, "unset number of bars")
	case bars.Sign() <= 0:
		return fraction.Fraction{}, duration.Duration{}, simpleError(ErrNoSolution, op, "number of bars not positive")
	case s.v.Sign() == 0:
		return fraction.Fraction{}, duration.Duration{}, simpleError(ErrNoSolution, op, "zero time")
	}
	rr := quo(mul(ratOf(t.m.bpm), minutes(s)), ratOf(bars))
	rf, err := toFraction(op, rr)
	if err != nil {
		return fraction.Fraction{}, duration.Duration{}, err
	}
	l, err := toDuration(op, mul(unitRat(t.m.unit), rr))
	if err != nil {
		return fraction.Fraction{}, duration.Duration{}, err
	}
	return rf, l, nil
}

// MustBarFor is like BarFor but panics with the returned error.
func (t Tempo) MustBarFor(bars fraction.Fraction, s Seconds) (r fraction.Fraction, barLen duration.Duration) {
	return must2(t.BarFor(bars, s))
}

// BarsPerMinute returns N/r = N·u/L, the number of bars of length barLen per
// minute: 60 for a bar of 3/4 at ♩ = 180. In order of checking, it returns an
// error wrapping ErrUnset if t is unset, ErrInvalidArgument if barLen is
// unset, and ErrOutOfRange together with fraction.ErrOutOfRange if the result
// is outside the range of package common/fraction.
func (t Tempo) BarsPerMinute(barLen duration.Duration) (fraction.Fraction, error) {
	const op = "Tempo.BarsPerMinute"
	if !t.IsSet() {
		return fraction.Fraction{}, t.unset(op)
	}
	if !barLen.IsSet() {
		return fraction.Fraction{}, invalidError(op, "unset bar length")
	}
	return toFraction(op, quo(t.m.speed(), durRat(barLen)))
}

// MustBarsPerMinute is like BarsPerMinute but panics with the returned error.
func (t Tempo) MustBarsPerMinute(barLen duration.Duration) fraction.Fraction {
	return must(t.BarsPerMinute(barLen))
}

// FromBars returns the tempo with unit u in which bars bars of length barLen
// last s seconds: N = 60·n·L/(u·s). In order of checking, it returns an error
// wrapping ErrUnset if s is unset, ErrInvalidArgument if unit, barLen, or bars
// is unset, ErrNoSolution if bars is zero or negative and then if s is zero,
// and ErrOutOfRange together with fraction.ErrOutOfRange if the result is
// outside the range of package common/fraction.
func FromBars(unit duration.Written, barLen duration.Duration, bars fraction.Fraction, s Seconds) (Tempo, error) {
	const op = "FromBars"
	switch {
	case !s.IsSet():
		return Tempo{}, unsetError(op, "time")
	case !unit.IsSet():
		return Tempo{}, invalidError(op, "unset unit")
	case !barLen.IsSet():
		return Tempo{}, invalidError(op, "unset bar length")
	case unsetFraction(bars):
		return Tempo{}, invalidError(op, "unset number of bars")
	case bars.Sign() <= 0:
		return Tempo{}, simpleError(ErrNoSolution, op, "number of bars not positive")
	case s.v.Sign() == 0:
		return Tempo{}, simpleError(ErrNoSolution, op, "zero time")
	}
	n := mul(ratInt(60), ratOf(bars), durRat(barLen))
	m, err := markOf(op, unit, quo(n, mul(unitRat(unit), ratOf(s.v))))
	return Tempo{m: m}, err
}

// MustFromBars is like FromBars but panics with the returned error.
func MustFromBars(unit duration.Written, barLen duration.Duration, bars fraction.Fraction, s Seconds) Tempo {
	return must(FromBars(unit, barLen, bars, s))
}

// FromBarsPerMinute returns the tempo with unit u at perMinute bars of length
// barLen per minute: N = m·L/u, so 60 bars of 3/4 per minute give ♩ = 180. In
// order of checking, it returns an error wrapping ErrInvalidArgument if unit,
// barLen, or perMinute is unset, ErrNoSolution if perMinute is zero or
// negative, and ErrOutOfRange together with fraction.ErrOutOfRange if the
// result is outside the range of package common/fraction.
func FromBarsPerMinute(unit duration.Written, barLen duration.Duration, perMinute fraction.Fraction) (Tempo, error) {
	const op = "FromBarsPerMinute"
	switch {
	case !unit.IsSet():
		return Tempo{}, invalidError(op, "unset unit")
	case !barLen.IsSet():
		return Tempo{}, invalidError(op, "unset bar length")
	case unsetFraction(perMinute):
		return Tempo{}, invalidError(op, "unset bars per minute")
	case perMinute.Sign() <= 0:
		return Tempo{}, simpleError(ErrNoSolution, op, "bars per minute not positive")
	}
	m, err := markOf(op, unit, quo(mul(ratOf(perMinute), durRat(barLen)), unitRat(unit)))
	return Tempo{m: m}, err
}

// MustFromBarsPerMinute is like FromBarsPerMinute but panics with the returned
// error.
func MustFromBarsPerMinute(unit duration.Written, barLen duration.Duration, perMinute fraction.Fraction) Tempo {
	return must(FromBarsPerMinute(unit, barLen, perMinute))
}

// BarUnits returns r = L/u, the number of units of duration unit in a bar of
// length barLen. In order of checking, it returns an error wrapping
// ErrInvalidArgument if barLen or then unit is unset, and ErrOutOfRange
// together with fraction.ErrOutOfRange if the result is outside the range of
// package common/fraction.
func BarUnits(barLen, unit duration.Duration) (fraction.Fraction, error) {
	const op = "BarUnits"
	switch {
	case !barLen.IsSet():
		return fraction.Fraction{}, invalidError(op, "unset bar length")
	case !unit.IsSet():
		return fraction.Fraction{}, invalidError(op, "unset unit")
	}
	return toFraction(op, quo(durRat(barLen), durRat(unit)))
}

// MustBarUnits is like BarUnits but panics with the returned error.
func MustBarUnits(barLen, unit duration.Duration) fraction.Fraction {
	return must(BarUnits(barLen, unit))
}

// BarLength returns L = u·r, the length of a bar of r units of duration unit:
// 1/2 for two quarters. In order of checking, it returns an error wrapping
// ErrInvalidArgument if unit or then r is unset, ErrNoSolution if r is zero or
// negative, and ErrOutOfRange together with fraction.ErrOutOfRange if the
// result is outside the range of package common/fraction.
func BarLength(unit duration.Duration, r fraction.Fraction) (duration.Duration, error) {
	const op = "BarLength"
	if err := checkLengthArgs(op, unit, "unit", r); err != nil {
		return duration.Duration{}, err
	}
	return toDuration(op, mul(durRat(unit), ratOf(r)))
}

// MustBarLength is like BarLength but panics with the returned error.
func MustBarLength(unit duration.Duration, r fraction.Fraction) duration.Duration {
	return must(BarLength(unit, r))
}

// UnitLength returns u = L/r, the duration of a unit when a bar of length
// barLen holds r units: 1/4 for a bar of 1/2 and two units. In order of
// checking, it returns an error wrapping ErrInvalidArgument if barLen or then
// r is unset, ErrNoSolution if r is zero or negative, and ErrOutOfRange
// together with fraction.ErrOutOfRange if the result is outside the range of
// package common/fraction.
func UnitLength(barLen duration.Duration, r fraction.Fraction) (duration.Duration, error) {
	const op = "UnitLength"
	if err := checkLengthArgs(op, barLen, "bar length", r); err != nil {
		return duration.Duration{}, err
	}
	return toDuration(op, quo(durRat(barLen), ratOf(r)))
}

// MustUnitLength is like UnitLength but panics with the returned error.
func MustUnitLength(barLen duration.Duration, r fraction.Fraction) duration.Duration {
	return must(UnitLength(barLen, r))
}

// checkLengthArgs checks a length and a number of units per bar in the order
// of BarLength and UnitLength.
func checkLengthArgs(op string, d duration.Duration, name string, r fraction.Fraction) error {
	switch {
	case !d.IsSet():
		return invalidError(op, "unset "+name)
	case unsetFraction(r):
		return invalidError(op, "unset number of units per bar")
	case r.Sign() <= 0:
		return simpleError(ErrNoSolution, op, "number of units per bar not positive")
	}
	return nil
}

// checkNumbers checks fractions of the formula in order: first that each is
// set, then that each is positive.
func checkNumbers(op string, names [2]string, fs [2]fraction.Fraction) error {
	for i, f := range fs {
		if unsetFraction(f) {
			return invalidError(op, "unset "+names[i])
		}
	}
	for i, f := range fs {
		if f.Sign() <= 0 {
			return simpleError(ErrNoSolution, op, names[i]+" not positive")
		}
	}
	return nil
}

// checkSolve checks the arguments of SolveBars, SolveBarUnits, and SolveBPM:
// the time is set, the fractions are set and positive, and the time is not
// zero.
func checkSolve(op string, names [2]string, fs [2]fraction.Fraction, s Seconds) error {
	if !s.IsSet() {
		return unsetError(op, "time")
	}
	if err := checkNumbers(op, names, fs); err != nil {
		return err
	}
	if s.v.Sign() == 0 {
		return simpleError(ErrNoSolution, op, "zero time")
	}
	return nil
}

// SolveBars returns n = N·t/r, the number of bars of r units in s seconds at
// bpm beats per minute. In order of checking, it returns an error wrapping
// ErrUnset if s is unset, ErrInvalidArgument if r or then bpm is unset,
// ErrNoSolution if r or then bpm is zero or negative and then if s is zero,
// and ErrOutOfRange together with fraction.ErrOutOfRange if the result is
// outside the range of package common/fraction. Here bpm is a number, not a
// marking, so a number that is not positive means no solution.
func SolveBars(r, bpm fraction.Fraction, s Seconds) (fraction.Fraction, error) {
	const op = "SolveBars"
	if err := checkSolve(op, [2]string{"number of units per bar", "beats per minute"}, [2]fraction.Fraction{r, bpm}, s); err != nil {
		return fraction.Fraction{}, err
	}
	return toFraction(op, quo(mul(ratOf(bpm), minutes(s)), ratOf(r)))
}

// MustSolveBars is like SolveBars but panics with the returned error.
func MustSolveBars(r, bpm fraction.Fraction, s Seconds) fraction.Fraction {
	return must(SolveBars(r, bpm, s))
}

// SolveBarUnits returns r = N·t/n, the number of units per bar when bars bars
// last s seconds at bpm beats per minute: 2 for 30 bars in 60 seconds at 60.
// The order of checking is that of SolveBars, with bars and bpm.
func SolveBarUnits(bars, bpm fraction.Fraction, s Seconds) (fraction.Fraction, error) {
	const op = "SolveBarUnits"
	if err := checkSolve(op, [2]string{"number of bars", "beats per minute"}, [2]fraction.Fraction{bars, bpm}, s); err != nil {
		return fraction.Fraction{}, err
	}
	return toFraction(op, quo(mul(ratOf(bpm), minutes(s)), ratOf(bars)))
}

// MustSolveBarUnits is like SolveBarUnits but panics with the returned error.
func MustSolveBarUnits(bars, bpm fraction.Fraction, s Seconds) fraction.Fraction {
	return must(SolveBarUnits(bars, bpm, s))
}

// SolveBPM returns N = n·r/t, the beats per minute at which bars bars of r
// units last s seconds. The order of checking is that of SolveBars, with bars
// and r.
func SolveBPM(bars, r fraction.Fraction, s Seconds) (fraction.Fraction, error) {
	const op = "SolveBPM"
	if err := checkSolve(op, [2]string{"number of bars", "number of units per bar"}, [2]fraction.Fraction{bars, r}, s); err != nil {
		return fraction.Fraction{}, err
	}
	return toFraction(op, quo(mul(ratOf(bars), ratOf(r)), minutes(s)))
}

// MustSolveBPM is like SolveBPM but panics with the returned error.
func MustSolveBPM(bars, r fraction.Fraction, s Seconds) fraction.Fraction {
	return must(SolveBPM(bars, r, s))
}

// SolveSeconds returns s = 60·n·r/N, the time of bars bars of r units at bpm
// beats per minute. In order of checking, it returns an error wrapping
// ErrInvalidArgument if bars, r, or bpm is unset, ErrNoSolution if bars, r, or
// bpm is zero or negative, and ErrOutOfRange together with
// fraction.ErrOutOfRange if the result is outside the range of package
// common/fraction.
func SolveSeconds(bars, r, bpm fraction.Fraction) (Seconds, error) {
	const op = "SolveSeconds"
	names := [3]string{"number of bars", "number of units per bar", "beats per minute"}
	fs := [3]fraction.Fraction{bars, r, bpm}
	for i, f := range fs {
		if unsetFraction(f) {
			return Seconds{}, invalidError(op, "unset "+names[i])
		}
	}
	for i, f := range fs {
		if f.Sign() <= 0 {
			return Seconds{}, simpleError(ErrNoSolution, op, names[i]+" not positive")
		}
	}
	return toSeconds(op, quo(mul(ratInt(60), ratOf(bars), ratOf(r)), ratOf(bpm)))
}

// MustSolveSeconds is like SolveSeconds but panics with the returned error.
func MustSolveSeconds(bars, r, bpm fraction.Fraction) Seconds {
	return must(SolveSeconds(bars, r, bpm))
}
