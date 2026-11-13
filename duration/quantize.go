package duration

import (
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
)

// checkGrid checks the rounding mode and then the grid step and returns the
// step value.
func checkGrid(op string, step fraction.Fraction, mode rounding.Mode) (*big.Rat, error) {
	if !mode.IsValid() {
		return nil, modeError(op, mode)
	}
	if !step.IsValid() {
		return nil, invalidError(op, "unset grid step")
	}
	if step.Sign() <= 0 {
		return nil, invalidError(op, "grid step is not positive")
	}
	return ratOf(step), nil
}

// quantize returns the node of the grid k·q selected for x ≥ 0 by mode, as
// rounding.Mode.RatMultiple does.
func quantize(op string, x, q *big.Rat, mode rounding.Mode) (*big.Rat, error) {
	_, v, err := mode.RatMultiple(x, q)
	if err != nil {
		// Unreachable: x and q are set and q > 0 after checkGrid.
		return nil, invalidError(op, "grid")
	}
	return v, nil
}

// Quantize returns the node of the grid k·step (k an integer, the grid starting
// at zero) that mode selects for d, as rounding.Mode.RatMultiple of package
// common/rounding does; a duration on a node is unchanged in every mode. Since
// durations are positive, TowardZero equals Down, AwayFromZero equals Up,
// HalfTowardZero equals HalfDown, and HalfAwayFromZero equals HalfUp. There is
// no default mode. In order of checking, it returns an error wrapping
// ErrUnset if d is unset, ErrInvalidArgument if mode is unset or unknown
// (together with rounding.ErrUnset or rounding.ErrUnknownMode) or if step is
// unset, zero, or negative, ErrQuantizedToZero if the selected node is zero,
// and ErrOutOfRange if it is outside the range of package common/fraction.
func (d Duration) Quantize(step fraction.Fraction, mode rounding.Mode) (Duration, error) {
	const op = "Duration.Quantize"
	if !d.IsSet() {
		return Duration{}, unsetError(op, "receiver")
	}
	q, err := checkGrid(op, step, mode)
	if err != nil {
		return Duration{}, err
	}
	v, err := quantize(op, ratOf(d.v), q, mode)
	if err != nil {
		return Duration{}, err
	}
	if v.Sign() == 0 {
		return Duration{}, simpleError(ErrQuantizedToZero, op)
	}
	return newDuration(op, v)
}

// MustQuantize is like Quantize but panics with the returned error.
func (d Duration) MustQuantize(step fraction.Fraction, mode rounding.Mode) Duration {
	return must(d.Quantize(step, mode))
}

// Quantize returns the node of the grid k·step (k an integer, the grid starting
// at the origin) that mode selects for p, as rounding.Mode.RatMultiple of
// package common/rounding does; the origin is a valid result. A grid shifted
// by a start position is obtained by subtracting the start first. In order of
// checking, it returns an error wrapping ErrUnset if p is unset,
// ErrInvalidArgument if mode is unset or unknown (together with
// rounding.ErrUnset or rounding.ErrUnknownMode) or if step is unset, zero, or
// negative, and ErrOutOfRange if the node is outside the range of package
// common/fraction.
func (p Position) Quantize(step fraction.Fraction, mode rounding.Mode) (Position, error) {
	const op = "Position.Quantize"
	if !p.IsSet() {
		return Position{}, unsetError(op, "receiver")
	}
	q, err := checkGrid(op, step, mode)
	if err != nil {
		return Position{}, err
	}
	v, err := quantize(op, ratOf(p.v), q, mode)
	if err != nil {
		return Position{}, err
	}
	return newPosition(op, v)
}

// MustQuantize is like Quantize but panics with the returned error.
func (p Position) MustQuantize(step fraction.Fraction, mode rounding.Mode) Position {
	return must(p.Quantize(step, mode))
}

// QuantizeSpan quantizes the span from start lasting length: it quantizes
// start and the end start + length with the same grid and mode, as
// Position.Quantize does, and returns the quantized start and the duration
// from it to the quantized end. So the end does not accumulate the rounding of
// the length. The exact sum start + length is not itself checked against the
// range. In order of checking, it returns an error wrapping ErrUnset if start
// or length is unset, ErrInvalidArgument if mode is unset or unknown or if
// step is unset, zero, or negative, ErrQuantizedToZero if the quantized start
// and end coincide, and ErrOutOfRange if the quantized start, end, or
// duration is outside the range of package common/fraction.
func QuantizeSpan(start Position, length Duration, step fraction.Fraction, mode rounding.Mode) (Position, Duration, error) {
	const op = "QuantizeSpan"
	if !start.IsSet() {
		return Position{}, Duration{}, unsetError(op, "start")
	}
	if !length.IsSet() {
		return Position{}, Duration{}, unsetError(op, "length")
	}
	q, err := checkGrid(op, step, mode)
	if err != nil {
		return Position{}, Duration{}, err
	}
	x := ratOf(start.v)
	end := new(big.Rat).Add(x, ratOf(length.v))
	s, err := quantize(op, x, q, mode)
	if err != nil {
		return Position{}, Duration{}, err
	}
	e, err := quantize(op, end, q, mode)
	if err != nil {
		return Position{}, Duration{}, err
	}
	if s.Cmp(e) == 0 {
		return Position{}, Duration{}, simpleError(ErrQuantizedToZero, op)
	}
	p, err := newPosition(op, s)
	if err != nil {
		return Position{}, Duration{}, err
	}
	if _, ok := fromRat(e); !ok {
		return Position{}, Duration{}, rangeError(op)
	}
	d, err := newDuration(op, e.Sub(e, s))
	if err != nil {
		return Position{}, Duration{}, err
	}
	return p, d, nil
}

// MustQuantizeSpan is like QuantizeSpan but panics with the returned error.
func MustQuantizeSpan(start Position, length Duration, step fraction.Fraction, mode rounding.Mode) (Position, Duration) {
	return must2(QuantizeSpan(start, length, step, mode))
}
