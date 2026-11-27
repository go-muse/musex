package fraction

import (
	"errors"
	"math"
	"math/big"
	"math/bits"
)

// GCD returns the greatest common divisor of f and other: the largest positive
// fraction g for which f/g and other/g are both integers. For the reduced
// forms f = p/q and other = r/s it equals gcd(|p|, |r|) / lcm(q, s) in
// canonical reduced form, so the sign of the operands does not matter,
// gcd(0, x) = |x|, and gcd(0, 0) = 0/1.
//
// GCD returns [ErrInvalid] if f or other is invalid and [ErrOutOfRange] if the
// denominator lcm(q, s) exceeds [MaxComponent].
func (f Fraction) GCD(other Fraction) (Fraction, error) {
	if !f.IsValid() || !other.IsValid() {
		return Fraction{}, ErrInvalid
	}
	a, b := f.Reduce(), other.Reduce()
	if a.num.isZero() && b.num.isZero() {
		return Zero(), nil
	}
	den, ok := lcmU256(a.den, b.den)
	if !ok {
		return Fraction{}, ErrOutOfRange
	}
	return makeFraction(u256GCD(a.num, b.num), den, false, false), nil
}

// MustGCD is like [Fraction.GCD] but panics on error.
func (f Fraction) MustGCD(other Fraction) Fraction {
	res, err := f.GCD(other)
	if err != nil {
		panic(err)
	}
	return res
}

// LCM returns the least common multiple of f and other: the smallest positive
// fraction that is an integer multiple of both. For the reduced forms
// f = p/q and other = r/s it equals lcm(|p|, |r|) / gcd(q, s) in canonical
// reduced form, so the sign of the operands does not matter; if f or other is
// zero, the result is 0/1.
//
// LCM returns [ErrInvalid] if f or other is invalid and [ErrOutOfRange] if the
// numerator lcm(|p|, |r|) exceeds [MaxComponent].
func (f Fraction) LCM(other Fraction) (Fraction, error) {
	if !f.IsValid() || !other.IsValid() {
		return Fraction{}, ErrInvalid
	}
	a, b := f.Reduce(), other.Reduce()
	if a.num.isZero() || b.num.isZero() {
		return Zero(), nil
	}
	num, ok := lcmU256(a.num, b.num)
	if !ok {
		return Fraction{}, ErrOutOfRange
	}
	return makeFraction(num, u256GCD(a.den, b.den), false, false), nil
}

// MustLCM is like [Fraction.LCM] but panics on error.
func (f Fraction) MustLCM(other Fraction) Fraction {
	res, err := f.LCM(other)
	if err != nil {
		panic(err)
	}
	return res
}

// CommonDenominator returns the least positive integer D for which every
// fraction of fs is an integer multiple of 1/D: the least common multiple of
// the reduced denominators (6/12 contributes 2, zero contributes 1). An empty
// fs yields 1. fs is only read.
//
// CommonDenominator returns [ErrInvalid] if any element of fs is invalid and
// [ErrNotRepresentable] if D exceeds math.MaxInt; validity of all elements is
// checked first. [CommonDenominatorBig] has no such limit.
func CommonDenominator(fs ...Fraction) (int, error) {
	if !allValid(fs) {
		return 0, ErrInvalid
	}
	acc := uint64(1)
	for i := range fs {
		d := reducedDen(fs[i])
		if !d.isUint64() {
			// d >= 2^64 > math.MaxInt, and the LCM is at least d.
			return 0, ErrNotRepresentable
		}
		if acc%d[0] == 0 {
			continue
		}
		hi, lo := bits.Mul64(acc/gcdUint64(acc, d[0]), d[0])
		if hi != 0 || lo > math.MaxInt {
			// The LCM of the whole set is not smaller than this prefix LCM.
			return 0, ErrNotRepresentable
		}
		acc = lo
	}
	return int(acc), nil
}

// MustCommonDenominator is like [CommonDenominator] but panics on error.
func MustCommonDenominator(fs ...Fraction) int {
	res, err := CommonDenominator(fs...)
	if err != nil {
		panic(err)
	}
	return res
}

// CommonDenominatorBig is like [CommonDenominator] without an upper limit: it
// returns a new *big.Int with the least common multiple of the reduced
// denominators of fs (1 for an empty fs). It returns [ErrInvalid] if any
// element of fs is invalid. The cost grows with the length of the result; a
// result above 512 bits continues in math/big.
func CommonDenominatorBig(fs ...Fraction) (*big.Int, error) {
	if !allValid(fs) {
		return nil, ErrInvalid
	}
	acc := u512{1}
	for i := range fs {
		d := reducedDen(fs[i])
		d512 := u512FromU256(d)
		g := u512GCD(acc, d512)
		if !g.isOne() {
			q, _ := u512DivMod(d512, g)
			d = q.low256()
		}
		if d.isOne() {
			continue
		}
		next, ok := u512MulU256(acc, d)
		if !ok {
			return commonDenominatorBigTail(acc, fs[i:]), nil
		}
		acc = next
	}
	return acc.toBigInt(false), nil
}

// MustCommonDenominatorBig is like [CommonDenominatorBig] but panics on error.
func MustCommonDenominatorBig(fs ...Fraction) *big.Int {
	res, err := CommonDenominatorBig(fs...)
	if err != nil {
		panic(err)
	}
	return res
}

// Sum returns the exact sum of all elements of fs in canonical reduced form
// (positive denominator, 0/1 for zero). An empty fs yields 0/1, and a single
// element yields its reduced form. fs is only read.
//
// The value and the [ErrOutOfRange] error depend only on the exact result, not
// on the order of the elements or on the size of intermediate sums: unlike a
// chain of [Fraction.Add] calls, an intermediate sum outside the range does not
// fail when the final sum is in range.
//
// Sum returns [ErrInvalid] if any element of fs is invalid (checked before any
// computation) and [ErrOutOfRange] if a component of the reduced result exceeds
// [MaxComponent]. It does not allocate while every reduced prefix sum is in
// range; otherwise the computation continues exactly in math/big, which
// allocates and costs more with the total length of the inputs.
func Sum(fs ...Fraction) (Fraction, error) {
	if !allValid(fs) {
		return Fraction{}, ErrInvalid
	}
	if len(fs) == 0 {
		return Zero(), nil
	}
	acc := fs[0].Reduce()
	for i := 1; i < len(fs); i++ {
		next, err := acc.Add(fs[i])
		if err != nil {
			if !errors.Is(err, ErrOutOfRange) {
				return Fraction{}, err
			}
			return sumTail(acc, fs[i:])
		}
		acc = next
	}
	return acc, nil
}

// MustSum is like [Sum] but panics on error.
func MustSum(fs ...Fraction) Fraction {
	res, err := Sum(fs...)
	if err != nil {
		panic(err)
	}
	return res
}

// Product returns the exact product of all elements of fs in canonical reduced
// form (positive denominator, 0/1 for zero). An empty fs yields 1/1, a single
// element yields its reduced form, and any zero element yields 0/1 regardless
// of the other valid elements. fs is only read.
//
// The value and the [ErrOutOfRange] error depend only on the exact result, not
// on the order of the elements or on the size of intermediate products: unlike
// a chain of [Fraction.Mul] calls, an intermediate product outside the range
// does not fail when the final product is in range.
//
// Product returns [ErrInvalid] if any element of fs is invalid (checked before
// any computation) and [ErrOutOfRange] if a component of the reduced result
// exceeds [MaxComponent]. It does not allocate while every reduced prefix
// product is in range; otherwise the computation continues exactly in
// math/big, which allocates and costs more with the total length of the
// inputs.
func Product(fs ...Fraction) (Fraction, error) {
	if !allValid(fs) {
		return Fraction{}, ErrInvalid
	}
	if len(fs) == 0 {
		return One(), nil
	}
	for i := range fs {
		if fs[i].num.isZero() {
			return Zero(), nil
		}
	}
	acc := fs[0].Reduce()
	for i := 1; i < len(fs); i++ {
		next, err := acc.Mul(fs[i])
		if err != nil {
			if !errors.Is(err, ErrOutOfRange) {
				return Fraction{}, err
			}
			return productTail(acc, fs[i:])
		}
		acc = next
	}
	return acc, nil
}

// MustProduct is like [Product] but panics on error.
func MustProduct(fs ...Fraction) Fraction {
	res, err := Product(fs...)
	if err != nil {
		panic(err)
	}
	return res
}

// allValid reports whether every element of fs is valid.
func allValid(fs []Fraction) bool {
	for i := range fs {
		if !fs[i].IsValid() {
			return false
		}
	}
	return true
}

// reducedDen returns the positive denominator of the reduced form of a valid f.
func reducedDen(f Fraction) u256 {
	if f.num.isZero() || f.den.isOne() {
		return u256One
	}
	g := u256GCD(f.num, f.den)
	if g.isOne() {
		return f.den
	}
	q, _ := u256DivMod(f.den, g)
	return q
}

// lcmU256 returns lcm(a, b) for a, b > 0 and true if it does not exceed
// MaxComponent; otherwise it returns an unspecified value and false.
func lcmU256(a, b u256) (u256, bool) {
	g := u256GCD(a, b)
	if !g.isOne() {
		a, _ = u256DivMod(a, g)
	}
	p := u256Mul(a, b)
	if !p.fitsU256Max() {
		return u256{}, false
	}
	return p.low256(), true
}

// commonDenominatorBigTail continues CommonDenominatorBig in math/big from the
// prefix LCM acc and the remaining valid elements rest.
func commonDenominatorBigTail(acc u512, rest []Fraction) *big.Int {
	lcm := acc.toBigInt(false)
	var d, g big.Int
	for i := range rest {
		reducedDen(rest[i]).setBigInt(&d, false)
		g.GCD(nil, nil, lcm, &d)
		d.Quo(&d, &g)
		lcm.Mul(lcm, &d)
	}
	return lcm
}

// sumTail continues Sum in math/big from the valid reduced prefix sum acc and
// the remaining valid elements rest, checking the range only for the result.
func sumTail(acc Fraction, rest []Fraction) (Fraction, error) {
	r, _ := acc.BigRatTo(nil)
	var x big.Rat
	for i := range rest {
		rest[i].BigRatTo(&x)
		r.Add(r, &x)
	}
	return FromBigRat(r)
}

// productTail continues Product in math/big from the valid reduced prefix
// product acc and the remaining valid non-zero elements rest, checking the
// range only for the result.
func productTail(acc Fraction, rest []Fraction) (Fraction, error) {
	r, _ := acc.BigRatTo(nil)
	var x big.Rat
	for i := range rest {
		rest[i].BigRatTo(&x)
		r.Mul(r, &x)
	}
	return FromBigRat(r)
}
