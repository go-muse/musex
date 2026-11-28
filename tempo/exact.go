package tempo

import (
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
)

// Intermediate values are exact *big.Rat values. The range of package
// common/fraction is checked only for results, so a failure depends on the
// exact result and not on the order of the calculation. A result that is a
// single operation on values in range, or a product of several values
// computed by fraction.Product, is computed in common/fraction directly: its
// range check also depends only on the exact result.

// Constant factors of the products computed in common/fraction.
var (
	inv60 = fraction.MustNew(1, 60)
	two   = fraction.FromInt(2)
	half  = fraction.MustNew(1, 2)
	four  = fraction.FromInt(4)
)

// ratOf returns a new *big.Rat equal to the valid fraction f.
func ratOf(f fraction.Fraction) *big.Rat {
	r, ok := f.BigRat()
	if !ok {
		return new(big.Rat)
	}
	return r
}

// ratInt returns n as a new *big.Rat.
func ratInt(n int64) *big.Rat {
	return new(big.Rat).SetInt64(n)
}

// durRat returns the value of the set duration d in whole notes.
func durRat(d duration.Duration) *big.Rat {
	f, _ := d.Fraction()
	return ratOf(f)
}

// unitRat returns the duration of the set written duration w in whole notes.
func unitRat(w duration.Written) *big.Rat {
	d, _ := w.Duration()
	return durRat(d)
}

// mul returns the product of the factors as a new *big.Rat.
func mul(factors ...*big.Rat) *big.Rat {
	r := ratInt(1)
	for _, f := range factors {
		r.Mul(r, f)
	}
	return r
}

// quo returns a/b as a new *big.Rat; b is not zero.
func quo(a, b *big.Rat) *big.Rat {
	return new(big.Rat).Quo(a, b)
}

// toFraction returns r as a reduced fraction, or an error wrapping
// ErrOutOfRange and fraction.ErrOutOfRange for op if r is outside the range.
func toFraction(op string, r *big.Rat) (fraction.Fraction, error) {
	f, err := fraction.FromBigRat(r)
	if err != nil {
		return fraction.Fraction{}, rangeError(op)
	}
	return f, nil
}

// fractionResult returns the result f of an operation of package common/fraction
// on set values in range, or an error wrapping ErrOutOfRange and
// fraction.ErrOutOfRange for op if the operation failed: for such operands the
// only possible failure is a result outside the range.
func fractionResult(op string, f fraction.Fraction, err error) (fraction.Fraction, error) {
	if err != nil {
		return fraction.Fraction{}, rangeError(op)
	}
	return f, nil
}

// durationResult returns the positive result f of an operation of package
// common/fraction as a duration, or the error of fractionResult.
func durationResult(op string, f fraction.Fraction, err error) (duration.Duration, error) {
	f, err = fractionResult(op, f, err)
	if err != nil {
		return duration.Duration{}, err
	}
	d, _ := duration.FromFraction(f)
	return d, nil
}

// secondsResult returns the result f ≥ 0 of an operation of package
// common/fraction as a time, or the error of fractionResult.
func secondsResult(op string, f fraction.Fraction, err error) (Seconds, error) {
	f, err = fractionResult(op, f, err)
	if err != nil {
		return Seconds{}, err
	}
	return Seconds{v: f}, nil
}

// durFraction returns the value of the set duration d in whole notes.
func durFraction(d duration.Duration) fraction.Fraction {
	f, _ := d.Fraction()
	return f
}

// toDuration returns the positive value r as a duration, or the error of
// toFraction.
func toDuration(op string, r *big.Rat) (duration.Duration, error) {
	f, err := toFraction(op, r)
	if err != nil {
		return duration.Duration{}, err
	}
	d, _ := duration.FromFraction(f)
	return d, nil
}

// toSeconds returns the value r ≥ 0 as a time, or the error of toFraction.
func toSeconds(op string, r *big.Rat) (Seconds, error) {
	f, err := toFraction(op, r)
	if err != nil {
		return Seconds{}, err
	}
	return Seconds{v: f}, nil
}

// unsetFraction reports whether f is unset; a set fraction is valid.
func unsetFraction(f fraction.Fraction) bool {
	return !f.IsValid()
}
