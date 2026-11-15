package measure

import (
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
)

// Exact intermediate values are *big.Rat and *big.Int. Values stored in a
// grid are never written after the grid is built: every computation writes
// only into new values, so stored values are safe for concurrent reading.

// ratZero is the value 0. It is only read.
var ratZero = new(big.Rat)

// ratOfFraction returns a new *big.Rat equal to the valid fraction f.
func ratOfFraction(f fraction.Fraction) *big.Rat {
	r, _ := f.BigRat()
	return r
}

// ratOfDuration returns a new *big.Rat equal to the set duration d.
func ratOfDuration(d duration.Duration) *big.Rat {
	return ratOfFraction(d.MustFraction())
}

// ratOfPosition returns a new *big.Rat equal to the set position p.
func ratOfPosition(p duration.Position) *big.Rat {
	return ratOfFraction(p.MustFraction())
}

// meterSize returns a new *big.Rat equal to the size of the set meter m.
func meterSize(m meter.Meter) *big.Rat {
	return ratOfDuration(m.MustDuration())
}

// ratInt returns a new *big.Rat equal to n.
func ratInt(n int) *big.Rat {
	return new(big.Rat).SetInt64(int64(n))
}

// ratAdd returns a new *big.Rat equal to a + b.
func ratAdd(a, b *big.Rat) *big.Rat {
	return new(big.Rat).Add(a, b)
}

// ratSub returns a new *big.Rat equal to a - b.
func ratSub(a, b *big.Rat) *big.Rat {
	return new(big.Rat).Sub(a, b)
}

// ratMul returns a new *big.Rat equal to a * b.
func ratMul(a, b *big.Rat) *big.Rat {
	return new(big.Rat).Mul(a, b)
}

// ratQuo returns a new *big.Rat equal to a / b; b is not zero.
func ratQuo(a, b *big.Rat) *big.Rat {
	return new(big.Rat).Quo(a, b)
}

// ratMulInt returns a new *big.Rat equal to a * n.
func ratMulInt(a *big.Rat, n int) *big.Rat {
	return new(big.Rat).Mul(a, ratInt(n))
}

// ratMulBigInt returns a new *big.Rat equal to a * n.
func ratMulBigInt(a *big.Rat, n *big.Int) *big.Rat {
	return new(big.Rat).Mul(a, new(big.Rat).SetInt(n))
}

// floorRat returns a new *big.Int equal to the floor of the non-negative r.
func floorRat(r *big.Rat) *big.Int {
	return new(big.Int).Quo(r.Num(), r.Denom())
}

// ceilRat returns a new *big.Int equal to the ceiling of the non-negative r.
func ceilRat(r *big.Rat) *big.Int {
	q, m := new(big.Int).QuoRem(r.Num(), r.Denom(), new(big.Int))
	if m.Sign() != 0 {
		q.Add(q, big.NewInt(1))
	}
	return q
}

// toFraction converts r to a reduced fraction. It returns ErrOutOfRange
// together with fraction.ErrOutOfRange if r is outside the range of
// common/fraction.
func toFraction(op string, r *big.Rat) (fraction.Fraction, error) {
	f, err := fraction.FromBigRat(r)
	if err != nil {
		return fraction.Fraction{}, errCaused(ErrOutOfRange, op, err)
	}
	return f, nil
}

// toPosition converts the non-negative r to a position. It returns
// ErrOutOfRange together with fraction.ErrOutOfRange if r is outside the
// range of common/fraction.
func toPosition(op string, r *big.Rat) (duration.Position, error) {
	f, err := toFraction(op, r)
	if err != nil {
		return duration.Position{}, err
	}
	// f is not negative, so the conversion cannot fail.
	return duration.MustPositionFromFraction(f), nil
}

// toDuration converts the positive r to a duration. It returns
// ErrOutOfRange together with fraction.ErrOutOfRange if r is outside the
// range of common/fraction.
func toDuration(op string, r *big.Rat) (duration.Duration, error) {
	f, err := toFraction(op, r)
	if err != nil {
		return duration.Duration{}, err
	}
	// f is positive, so the conversion cannot fail.
	return duration.MustFromFraction(f), nil
}
