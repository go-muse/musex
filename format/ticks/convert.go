package ticks

import (
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
)

// unitsOf returns the exact number of units v*D/w of a value in whole notes.
func (r Resolution) unitsOf(value *big.Rat) *big.Rat {
	units := new(big.Rat).Mul(value, ratInt(r.d))
	return units.Quo(units, r.quarter())
}

func checkDuration(op string, d duration.Duration) error {
	if !d.IsSet() {
		return invalidError(op, "duration", duration.ErrUnset)
	}
	return nil
}

func checkPosition(op string, p duration.Position) error {
	if !p.IsSet() {
		return invalidError(op, "position", duration.ErrUnset)
	}
	return nil
}

// wholeTicks returns an integral number of units as int.
func wholeTicks(op string, units *big.Rat) (int, error) {
	if !units.IsInt() {
		return 0, newError(ErrNotRepresentable, op, "not a whole number of ticks")
	}
	return intOf(op, units.Num())
}

// Ticks returns the duration d in ticks: d*D/w. A value that is not a whole
// number of ticks is ErrNotRepresentable; one outside int is ErrOutOfRange.
func (r Resolution) Ticks(d duration.Duration) (int, error) {
	const op = "Resolution.Ticks"
	if err := r.check(op); err != nil {
		return 0, err
	}
	if err := checkDuration(op, d); err != nil {
		return 0, err
	}
	return wholeTicks(op, r.unitsOf(ratOfDuration(d)))
}

// MustTicks is like Ticks but panics with its error.
func (r Resolution) MustTicks(d duration.Duration) int { return must(r.Ticks(d)) }

// PositionTicks returns the position p in ticks, like Ticks.
func (r Resolution) PositionTicks(p duration.Position) (int, error) {
	const op = "Resolution.PositionTicks"
	if err := r.check(op); err != nil {
		return 0, err
	}
	if err := checkPosition(op, p); err != nil {
		return 0, err
	}
	return wholeTicks(op, r.unitsOf(ratOfPosition(p)))
}

// MustPositionTicks is like PositionTicks but panics with its error.
func (r Resolution) MustPositionTicks(p duration.Position) int { return must(r.PositionTicks(p)) }

// Units returns the exact number of units d*D/w, for example the MusicXML
// duration of d in divisions.
func (r Resolution) Units(d duration.Duration) (fraction.Fraction, error) {
	const op = "Resolution.Units"
	if err := r.check(op); err != nil {
		return fraction.Fraction{}, err
	}
	if err := checkDuration(op, d); err != nil {
		return fraction.Fraction{}, err
	}
	return fractionOf(op, r.unitsOf(ratOfDuration(d)))
}

// MustUnits is like Units but panics with its error.
func (r Resolution) MustUnits(d duration.Duration) fraction.Fraction { return must(r.Units(d)) }

// PositionUnits returns the exact number of units of the position p.
func (r Resolution) PositionUnits(p duration.Position) (fraction.Fraction, error) {
	const op = "Resolution.PositionUnits"
	if err := r.check(op); err != nil {
		return fraction.Fraction{}, err
	}
	if err := checkPosition(op, p); err != nil {
		return fraction.Fraction{}, err
	}
	return fractionOf(op, r.unitsOf(ratOfPosition(p)))
}

// MustPositionUnits is like PositionUnits but panics with its error.
func (r Resolution) MustPositionUnits(p duration.Position) fraction.Fraction {
	return must(r.PositionUnits(p))
}

// Duration returns the duration of ticks > 0 ticks: ticks*w/D whole notes.
func (r Resolution) Duration(ticks int) (duration.Duration, error) {
	const op = "Resolution.Duration"
	if err := r.check(op); err != nil {
		return duration.Duration{}, err
	}
	if ticks <= 0 {
		return duration.Duration{}, invalidError(op, "ticks of a duration are not positive")
	}
	return durationOf(op, new(big.Rat).Mul(ratInt(ticks), r.tick()))
}

// MustDuration is like Duration but panics with its error.
func (r Resolution) MustDuration(ticks int) duration.Duration { return must(r.Duration(ticks)) }

// Position returns the position of ticks >= 0 ticks: ticks*w/D whole notes.
func (r Resolution) Position(ticks int) (duration.Position, error) {
	const op = "Resolution.Position"
	if err := r.check(op); err != nil {
		return duration.Position{}, err
	}
	if ticks < 0 {
		return duration.Position{}, invalidError(op, "negative ticks")
	}
	return positionOf(op, new(big.Rat).Mul(ratInt(ticks), r.tick()))
}

// MustPosition is like Position but panics with its error.
func (r Resolution) MustPosition(ticks int) duration.Position { return must(r.Position(ticks)) }

func unitsArgument(op string, units fraction.Fraction) (*big.Rat, error) {
	value, ok := units.BigRat()
	if !ok {
		return nil, invalidError(op, "units", fraction.ErrInvalid)
	}
	return value, nil
}

// UnitsDuration returns the duration of an exact positive number of units:
// units*w/D whole notes.
func (r Resolution) UnitsDuration(units fraction.Fraction) (duration.Duration, error) {
	const op = "Resolution.UnitsDuration"
	if err := r.check(op); err != nil {
		return duration.Duration{}, err
	}
	value, err := unitsArgument(op, units)
	if err != nil {
		return duration.Duration{}, err
	}
	if value.Sign() <= 0 {
		return duration.Duration{}, invalidError(op, "units of a duration are not positive")
	}
	return durationOf(op, value.Mul(value, r.tick()))
}

// MustUnitsDuration is like UnitsDuration but panics with its error.
func (r Resolution) MustUnitsDuration(units fraction.Fraction) duration.Duration {
	return must(r.UnitsDuration(units))
}

// UnitsPosition returns the position of an exact non-negative number of units.
func (r Resolution) UnitsPosition(units fraction.Fraction) (duration.Position, error) {
	const op = "Resolution.UnitsPosition"
	if err := r.check(op); err != nil {
		return duration.Position{}, err
	}
	value, err := unitsArgument(op, units)
	if err != nil {
		return duration.Position{}, err
	}
	if value.Sign() < 0 {
		return duration.Position{}, invalidError(op, "negative units")
	}
	return positionOf(op, value.Mul(value, r.tick()))
}

// MustUnitsPosition is like UnitsPosition but panics with its error.
func (r Resolution) MustUnitsPosition(units fraction.Fraction) duration.Position {
	return must(r.UnitsPosition(units))
}

// quantize rounds the exact units of value once with mode. A duration that
// rounds to zero ticks is ErrInvalidArgument.
func quantize(op string, units, original, tick *big.Rat, mode rounding.Mode, isDuration bool) (Quantized, error) {
	n, err := roundInt(op, units, mode)
	if err != nil {
		return Quantized{}, err
	}
	if isDuration && n.Sign() == 0 {
		return Quantized{}, invalidError(op, "duration quantized to zero ticks")
	}
	ticks, err := intOf(op, n)
	if err != nil {
		return Quantized{}, err
	}
	deviation := new(big.Rat).Mul(new(big.Rat).SetInt(n), tick)
	deviation.Sub(deviation, original)
	dev, err := fractionOf(op, deviation)
	if err != nil {
		return Quantized{}, err
	}
	return Quantized{ticks: ticks, deviation: dev, set: true}, nil
}

// Quantize rounds the duration d to whole ticks once with mode, from the exact
// value d*D/w, and returns the ticks and the exact deviation in whole notes. A
// duration that rounds to zero ticks is ErrInvalidArgument.
func (r Resolution) Quantize(d duration.Duration, mode rounding.Mode) (Quantized, error) {
	const op = "Resolution.Quantize"
	if err := r.check(op); err != nil {
		return Quantized{}, err
	}
	if err := checkDuration(op, d); err != nil {
		return Quantized{}, err
	}
	if err := checkMode(op, mode); err != nil {
		return Quantized{}, err
	}
	value := ratOfDuration(d)
	return quantize(op, r.unitsOf(value), value, r.tick(), mode, true)
}

// MustQuantize is like Quantize but panics with its error.
func (r Resolution) MustQuantize(d duration.Duration, mode rounding.Mode) Quantized {
	return must(r.Quantize(d, mode))
}

// QuantizePosition rounds the position p to whole ticks; zero ticks are valid.
func (r Resolution) QuantizePosition(p duration.Position, mode rounding.Mode) (Quantized, error) {
	const op = "Resolution.QuantizePosition"
	if err := r.check(op); err != nil {
		return Quantized{}, err
	}
	if err := checkPosition(op, p); err != nil {
		return Quantized{}, err
	}
	if err := checkMode(op, mode); err != nil {
		return Quantized{}, err
	}
	value := ratOfPosition(p)
	return quantize(op, r.unitsOf(value), value, r.tick(), mode, false)
}

// MustQuantizePosition is like QuantizePosition but panics with its error.
func (r Resolution) MustQuantizePosition(p duration.Position, mode rounding.Mode) Quantized {
	return must(r.QuantizePosition(p, mode))
}

// rescaled returns the exact ticks of t ticks of r at the resolution to.
func (r Resolution) rescaled(t int, to Resolution) *big.Rat {
	value := new(big.Rat).Mul(ratInt(t), r.tick())
	return value.Quo(value, to.tick())
}

func (r Resolution) rescale(op string, t int, to Resolution, isDuration bool) (int, error) {
	if err := r.check(op); err != nil {
		return 0, err
	}
	if isDuration && t <= 0 {
		return 0, invalidError(op, "ticks of a duration are not positive")
	}
	if t < 0 {
		return 0, invalidError(op, "negative ticks")
	}
	if !to.IsSet() {
		return 0, unsetError(op, "target resolution")
	}
	value := r.rescaled(t, to)
	if !value.IsInt() {
		return 0, newError(ErrNotRepresentable, op, "rescaled ticks are not whole")
	}
	return intOf(op, value.Num())
}

// Rescale returns the duration of t > 0 ticks of r in ticks of to:
// t*(w*D')/(w'*D). A result that is not whole is ErrNotRepresentable.
func (r Resolution) Rescale(ticks int, to Resolution) (int, error) {
	return r.rescale("Resolution.Rescale", ticks, to, true)
}

// MustRescale is like Rescale but panics with its error.
func (r Resolution) MustRescale(ticks int, to Resolution) int { return must(r.Rescale(ticks, to)) }

// RescalePosition is Rescale for a position: t >= 0.
func (r Resolution) RescalePosition(ticks int, to Resolution) (int, error) {
	return r.rescale("Resolution.RescalePosition", ticks, to, false)
}

// MustRescalePosition is like RescalePosition but panics with its error.
func (r Resolution) MustRescalePosition(ticks int, to Resolution) int {
	return must(r.RescalePosition(ticks, to))
}

func (r Resolution) quantizeRescale(op string, t int, to Resolution, mode rounding.Mode, isDuration bool) (Quantized, error) {
	if err := r.check(op); err != nil {
		return Quantized{}, err
	}
	if isDuration && t <= 0 {
		return Quantized{}, invalidError(op, "ticks of a duration are not positive")
	}
	if t < 0 {
		return Quantized{}, invalidError(op, "negative ticks")
	}
	if !to.IsSet() {
		return Quantized{}, unsetError(op, "target resolution")
	}
	if err := checkMode(op, mode); err != nil {
		return Quantized{}, err
	}
	original := new(big.Rat).Mul(ratInt(t), r.tick())
	return quantize(op, r.rescaled(t, to), original, to.tick(), mode, isDuration)
}

// QuantizeRescale rounds the duration of t > 0 ticks of r to whole ticks of to
// once with mode and returns the deviation in whole notes. A result of zero
// ticks is ErrInvalidArgument.
func (r Resolution) QuantizeRescale(ticks int, to Resolution, mode rounding.Mode) (Quantized, error) {
	return r.quantizeRescale("Resolution.QuantizeRescale", ticks, to, mode, true)
}

// MustQuantizeRescale is like QuantizeRescale but panics with its error.
func (r Resolution) MustQuantizeRescale(ticks int, to Resolution, mode rounding.Mode) Quantized {
	return must(r.QuantizeRescale(ticks, to, mode))
}

// QuantizeRescalePosition is QuantizeRescale for a position: t >= 0 and zero
// ticks are valid.
func (r Resolution) QuantizeRescalePosition(ticks int, to Resolution, mode rounding.Mode) (Quantized, error) {
	return r.quantizeRescale("Resolution.QuantizeRescalePosition", ticks, to, mode, false)
}

// MustQuantizeRescalePosition is like QuantizeRescalePosition but panics with
// its error.
func (r Resolution) MustQuantizeRescalePosition(ticks int, to Resolution, mode rounding.Mode) Quantized {
	return must(r.QuantizeRescalePosition(ticks, to, mode))
}
