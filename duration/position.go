package duration

import (
	"math/big"

	"github.com/go-muse/musex/common/fraction"
)

// Position is an exact offset of musical time from an origin chosen by the
// caller, in whole notes, at least zero. Zero is the origin itself and is a
// valid position, unlike a zero duration.
//
// The zero value Position{} is the unset position; it is not the origin,
// which Origin returns. A set position is stored in reduced form. Position is
// comparable with ==, which matches Equal, and can be used as a map key.
type Position struct {
	v fraction.Fraction
}

// newPosition returns the reduced position r ≥ 0, or an error wrapping
// ErrOutOfRange if r is outside the range of package common/fraction.
func newPosition(op string, r *big.Rat) (Position, error) {
	f, ok := fromRat(r)
	if !ok {
		return Position{}, rangeError(op)
	}
	return Position{v: f}, nil
}

// positionOf returns the value f ≥ 0 of a fraction operation as a position,
// or an error wrapping ErrOutOfRange if the operation failed. The operands of
// every such operation are set and in range, so the only possible failure is
// a result outside the range of package common/fraction.
func positionOf(op string, f fraction.Fraction, err error) (Position, error) {
	if err != nil {
		return Position{}, rangeError(op)
	}
	return Position{v: f}, nil
}

// Origin returns the position zero, the origin chosen by the caller.
func Origin() Position {
	return Position{v: fraction.Zero()}
}

// NewPosition returns the position num/den in whole notes. The signs of num
// and den are taken by value. It returns an error wrapping ErrInvalidArgument
// and fraction.ErrZeroDenominator if den is zero, and ErrNegativePosition if
// the value is negative.
func NewPosition(num, den int) (Position, error) {
	const op = "NewPosition"
	if den == 0 {
		return Position{}, zeroDenominatorError(op)
	}
	r := new(big.Rat).SetFrac(big.NewInt(int64(num)), big.NewInt(int64(den)))
	if r.Sign() < 0 {
		return Position{}, simpleError(ErrNegativePosition, op)
	}
	return newPosition(op, r)
}

// MustNewPosition is like NewPosition but panics with the returned error.
func MustNewPosition(num, den int) Position {
	return must(NewPosition(num, den))
}

// PositionFromFraction returns the position with the value of f in whole
// notes, in reduced form. It returns an error wrapping ErrInvalidArgument if
// f is unset and ErrNegativePosition if f is negative.
func PositionFromFraction(f fraction.Fraction) (Position, error) {
	const op = "PositionFromFraction"
	if !f.IsValid() {
		return Position{}, invalidError(op, "unset fraction")
	}
	if f.Sign() < 0 {
		return Position{}, simpleError(ErrNegativePosition, op)
	}
	return Position{v: f.Reduce()}, nil
}

// MustPositionFromFraction is like PositionFromFraction but panics with the
// returned error.
func MustPositionFromFraction(f fraction.Fraction) Position {
	return must(PositionFromFraction(f))
}

// IsSet reports whether p is a set position.
func (p Position) IsSet() bool {
	return p.v.IsValid()
}

// Fraction returns the value of p as a reduced fraction with a positive
// denominator; for the origin it is 0/1. It returns an error wrapping
// ErrUnset if p is unset.
func (p Position) Fraction() (fraction.Fraction, error) {
	if !p.IsSet() {
		return fraction.Fraction{}, unsetError("Position.Fraction", "receiver")
	}
	return p.v, nil
}

// MustFraction is like Fraction but panics with the returned error.
func (p Position) MustFraction() fraction.Fraction {
	return must(p.Fraction())
}

// Equal reports whether p and other have the same value. It matches ==; the
// unset position is equal only to the unset position.
func (p Position) Equal(other Position) bool {
	return p == other
}

// Compare compares p and other by value and returns −1, 0, or +1. The unset
// position precedes every set position, and two unset positions are equal.
func (p Position) Compare(other Position) int {
	return p.v.Compare(other.v)
}

// withDuration checks that p and d are set.
func (p Position) withDuration(op string, d Duration) error {
	if !p.IsSet() {
		return unsetError(op, "receiver")
	}
	if !d.IsSet() {
		return unsetError(op, "duration")
	}
	return nil
}

// Add returns the position p + d. It returns an error wrapping ErrUnset if p
// or d is unset and ErrOutOfRange if the result is outside the range of
// package common/fraction.
func (p Position) Add(d Duration) (Position, error) {
	const op = "Position.Add"
	if err := p.withDuration(op, d); err != nil {
		return Position{}, err
	}
	sum, err := p.v.Add(d.v)
	return positionOf(op, sum, err)
}

// MustAdd is like Add but panics with the returned error.
func (p Position) MustAdd(d Duration) Position {
	return must(p.Add(d))
}

// Sub returns the position p − d. It returns an error wrapping ErrUnset if p
// or d is unset, ErrNegativePosition if the result is negative, and
// ErrOutOfRange if it is outside the range of package common/fraction.
func (p Position) Sub(d Duration) (Position, error) {
	const op = "Position.Sub"
	if err := p.withDuration(op, d); err != nil {
		return Position{}, err
	}
	// The sign is decided by an exact comparison before the subtraction, so
	// a result that is both negative and outside the range reports
	// ErrNegativePosition.
	if p.v.Compare(d.v) < 0 {
		return Position{}, simpleError(ErrNegativePosition, op)
	}
	diff, err := p.v.Sub(d.v)
	return positionOf(op, diff, err)
}

// MustSub is like Sub but panics with the returned error.
func (p Position) MustSub(d Duration) Position {
	return must(p.Sub(d))
}

// between checks that p and end are set and returns the sign of end − p,
// decided by an exact comparison before any subtraction, so that the sign
// errors of the callers precede the range error of the difference.
func (p Position) between(op string, end Position) (int, error) {
	if !p.IsSet() {
		return 0, unsetError(op, "receiver")
	}
	if !end.IsSet() {
		return 0, unsetError(op, "argument")
	}
	return end.v.Compare(p.v), nil
}

// OffsetTo returns the offset end − p ≥ 0 as a position: from 1/2 to 1/2 it is
// the origin. It returns an error wrapping ErrUnset if p or end is unset,
// ErrNegativePosition if end precedes p, and ErrOutOfRange if the offset is
// outside the range of package common/fraction.
func (p Position) OffsetTo(end Position) (Position, error) {
	const op = "Position.OffsetTo"
	sign, err := p.between(op, end)
	if err != nil {
		return Position{}, err
	}
	if sign < 0 {
		return Position{}, simpleError(ErrNegativePosition, op)
	}
	diff, err := end.v.Sub(p.v)
	return positionOf(op, diff, err)
}

// MustOffsetTo is like OffsetTo but panics with the returned error.
func (p Position) MustOffsetTo(end Position) Position {
	return must(p.OffsetTo(end))
}

// DurationTo returns the duration end − p > 0 from p to end. It returns an
// error wrapping ErrUnset if p or end is unset, ErrNotPositive if end does not
// follow p, and ErrOutOfRange if the duration is outside the range of package
// common/fraction.
func (p Position) DurationTo(end Position) (Duration, error) {
	const op = "Position.DurationTo"
	sign, err := p.between(op, end)
	if err != nil {
		return Duration{}, err
	}
	if sign <= 0 {
		return Duration{}, simpleError(ErrNotPositive, op)
	}
	diff, err := end.v.Sub(p.v)
	return durationOf(op, diff, err)
}

// MustDurationTo is like DurationTo but panics with the returned error.
func (p Position) MustDurationTo(end Position) Duration {
	return must(p.DurationTo(end))
}
