package tuplet

import (
	"fmt"
	"math"
	"math/bits"
	"strconv"

	"github.com/go-muse/musex/common/fraction"
)

// Coefficient returns the duration coefficient normal/actual with the stored
// pair (normal, actual), unreduced: 4/6 for 6:4 and 2/3 for 3:2. A written
// duration multiplied by the coefficient is the sounding duration of each
// note of the tuplet. It returns an error wrapping ErrUnset if t is unset.
func (t Tuplet) Coefficient() (fraction.Fraction, error) {
	if !t.IsSet() {
		return fraction.Fraction{}, unsetError("receiver")
	}
	return fraction.MustNew(t.normal, t.actual), nil
}

// MustCoefficient is like Coefficient but panics with the returned error.
func (t Tuplet) MustCoefficient() fraction.Fraction {
	return must(t.Coefficient())
}

// ReciprocalCoefficient returns the reciprocal coefficient actual/normal with
// the stored pair (actual, normal), unreduced: 6/4 for 6:4. It returns an
// error wrapping ErrUnset if t is unset.
func (t Tuplet) ReciprocalCoefficient() (fraction.Fraction, error) {
	if !t.IsSet() {
		return fraction.Fraction{}, unsetError("receiver")
	}
	return fraction.MustNew(t.actual, t.normal), nil
}

// MustReciprocalCoefficient is like ReciprocalCoefficient but panics with the
// returned error.
func (t Tuplet) MustReciprocalCoefficient() fraction.Fraction {
	return must(t.ReciprocalCoefficient())
}

// ReducedCoefficient returns the coefficient normal/actual in canonical
// reduced form with a positive denominator: 2/3 for 6:4. It returns an error
// wrapping ErrUnset if t is unset.
func (t Tuplet) ReducedCoefficient() (fraction.Fraction, error) {
	if !t.IsSet() {
		return fraction.Fraction{}, unsetError("receiver")
	}
	return fraction.MustNewReduced(t.normal, t.actual), nil
}

// MustReducedCoefficient is like ReducedCoefficient but panics with the
// returned error.
func (t Tuplet) MustReducedCoefficient() fraction.Fraction {
	return must(t.ReducedCoefficient())
}

// Reduce returns the tuplet (actual/g):(normal/g), where g is the greatest
// common divisor of the counts: 3:2 for 6:4. The coefficient value is
// unchanged. It returns an error wrapping ErrUnset if t is unset.
func (t Tuplet) Reduce() (Tuplet, error) {
	if !t.IsSet() {
		return Tuplet{}, unsetError("receiver")
	}
	g := gcd(t.actual, t.normal)
	return Tuplet{actual: t.actual / g, normal: t.normal / g}, nil
}

// MustReduce is like Reduce but panics with the returned error.
func (t Tuplet) MustReduce() Tuplet {
	return must(t.Reduce())
}

// Scale returns the tuplet (k·actual):(k·normal): 6:4 for 3:2 and k = 2. The
// coefficient value is unchanged. In order of checking, it returns an error
// wrapping ErrUnset if t is unset, ErrInvalidArgument if k < 1, and
// ErrOutOfRange if a product exceeds math.MaxInt.
func (t Tuplet) Scale(k int) (Tuplet, error) {
	if !t.IsSet() {
		return Tuplet{}, unsetError("receiver")
	}
	if k < 1 {
		return Tuplet{}, fmt.Errorf("%w: scale factor %d is less than 1", ErrInvalidArgument, k)
	}
	actual, okActual := mulInt(t.actual, k)
	normal, okNormal := mulInt(t.normal, k)
	if !okActual || !okNormal {
		return Tuplet{}, fmt.Errorf("%w: %s scaled by %d", ErrOutOfRange, t, k)
	}
	return Tuplet{actual: actual, normal: normal}, nil
}

// MustScale is like Scale but panics with the returned error.
func (t Tuplet) MustScale(k int) Tuplet {
	return must(t.Scale(k))
}

// checkApply returns an error wrapping ErrUnset if t is unset and
// ErrInvalidArgument if d is unset.
func (t Tuplet) checkApply(d fraction.Fraction) error {
	if !t.IsSet() {
		return unsetError("receiver")
	}
	if !d.IsValid() {
		return fmt.Errorf("%w: unset fraction", ErrInvalidArgument)
	}
	return nil
}

// Apply returns d·normal/actual, the sounding duration of a note with written
// duration d inside the tuplet, in canonical reduced form: a triplet maps 1/8
// to 1/12 and a duplet maps 1/8 to 3/16. Any set fraction is accepted,
// including zero and negative values. In order of checking, it returns an
// error wrapping ErrUnset if t is unset, ErrInvalidArgument if d is unset, and
// ErrOutOfRange together with fraction.ErrOutOfRange if the reduced result is
// outside the range of package common/fraction.
func (t Tuplet) Apply(d fraction.Fraction) (fraction.Fraction, error) {
	if err := t.checkApply(d); err != nil {
		return fraction.Fraction{}, err
	}
	r, err := d.Mul(fraction.MustNew(t.normal, t.actual))
	if err != nil {
		return fraction.Fraction{}, fractionRangeError("apply", err)
	}
	return r, nil
}

// MustApply is like Apply but panics with the returned error.
func (t Tuplet) MustApply(d fraction.Fraction) fraction.Fraction {
	return must(t.Apply(d))
}

// Unapply returns d·actual/normal, which undoes Apply, in canonical reduced
// form. Errors are as for Apply.
func (t Tuplet) Unapply(d fraction.Fraction) (fraction.Fraction, error) {
	if err := t.checkApply(d); err != nil {
		return fraction.Fraction{}, err
	}
	r, err := d.Mul(fraction.MustNew(t.actual, t.normal))
	if err != nil {
		return fraction.Fraction{}, fractionRangeError("unapply", err)
	}
	return r, nil
}

// MustUnapply is like Unapply but panics with the returned error.
func (t Tuplet) MustUnapply(d fraction.Fraction) fraction.Fraction {
	return must(t.Unapply(d))
}

// Span returns normal·unit, the total duration of the tuplet written with
// parts of duration unit, in canonical reduced form: 5:4 in eighths spans 1/2.
// Errors are as for Apply, with unit in place of d.
func (t Tuplet) Span(unit fraction.Fraction) (fraction.Fraction, error) {
	if err := t.checkApply(unit); err != nil {
		return fraction.Fraction{}, err
	}
	r, err := unit.MulInt(t.normal)
	if err != nil {
		return fraction.Fraction{}, fractionRangeError("span", err)
	}
	return r, nil
}

// MustSpan is like Span but panics with the returned error.
func (t Tuplet) MustSpan(unit fraction.Fraction) fraction.Fraction {
	return must(t.Span(unit))
}

// checkChain returns an error wrapping ErrUnset if any element of chain is
// unset.
func checkChain(chain []Tuplet) error {
	for i, t := range chain {
		if !t.IsSet() {
			return unsetError("chain element " + strconv.Itoa(i))
		}
	}
	return nil
}

// product returns the tuplet (a.actual·b.actual):(a.normal·b.normal) for set
// tuplets a and b, or an error wrapping ErrOutOfRange if a product exceeds
// math.MaxInt.
func product(a, b Tuplet) (Tuplet, error) {
	actual, okActual := mulInt(a.actual, b.actual)
	normal, okNormal := mulInt(a.normal, b.normal)
	if !okActual || !okNormal {
		return Tuplet{}, fmt.Errorf("%w: flattened counts exceed int", ErrOutOfRange)
	}
	return Tuplet{actual: actual, normal: normal}, nil
}

// Nest returns the flattened notation of t with inner nested inside one of its
// parts: Flatten(t, inner). A duplet 2:3 inside a triplet 3:2 gives 6:6. In
// order of checking, it returns an error wrapping ErrUnset if t or inner is
// unset and ErrOutOfRange if a product of counts exceeds math.MaxInt.
func (t Tuplet) Nest(inner Tuplet) (Tuplet, error) {
	if !t.IsSet() {
		return Tuplet{}, unsetError("receiver")
	}
	if !inner.IsSet() {
		return Tuplet{}, unsetError("argument")
	}
	return product(t, inner)
}

// MustNest is like Nest but panics with the returned error.
func (t Tuplet) MustNest(inner Tuplet) Tuplet {
	return must(t.Nest(inner))
}

// Flatten returns the flattened notation (n1·n2·…):(m1·m2·…) of a chain of
// tuplets listed from the outermost to the innermost: two triplets give 9:4.
// An empty chain gives 1:1 and a single tuplet gives itself. The result does
// not depend on the order of the chain. In order of checking, it returns an
// error wrapping ErrUnset if any element is unset and ErrOutOfRange if a
// product of counts exceeds math.MaxInt. The chain is only read.
func Flatten(chain ...Tuplet) (Tuplet, error) {
	if err := checkChain(chain); err != nil {
		return Tuplet{}, err
	}
	acc := Tuplet{actual: 1, normal: 1}
	for _, t := range chain {
		var err error
		if acc, err = product(acc, t); err != nil {
			return Tuplet{}, err
		}
	}
	return acc, nil
}

// MustFlatten is like Flatten but panics with the returned error.
func MustFlatten(chain ...Tuplet) Tuplet {
	return must(Flatten(chain...))
}

// ChainCoefficient returns the total coefficient of a chain of tuplets listed
// from the outermost to the innermost, with the unreduced stored pair
// (m1·m2·…, n1·n2·…): a triplet inside a quintuplet gives 8/15. An empty chain
// gives 1/1. For a successful Flatten of the same chain, the result has the
// same stored pair as its Coefficient. The result, including failure, does
// not depend on the order of the chain. In order of checking, it returns an
// error wrapping ErrUnset if any element is unset, and ErrOutOfRange together
// with fraction.ErrOutOfRange if a product exceeds fraction.MaxComponent. Any
// chain of four tuplets is representable. The chain is only read.
func ChainCoefficient(chain ...Tuplet) (fraction.Fraction, error) {
	if err := checkChain(chain); err != nil {
		return fraction.Fraction{}, err
	}
	return chainCoefficient(chain)
}

// chainCoefficient multiplies the coefficients of a chain whose elements are
// all set.
func chainCoefficient(chain []Tuplet) (fraction.Fraction, error) {
	acc := fraction.One()
	for _, t := range chain {
		var err error
		acc, err = acc.MulWith(fraction.MustNew(t.normal, t.actual), fraction.DenomUnreducedProduct)
		if err != nil {
			return fraction.Fraction{}, fractionRangeError("chain coefficient", err)
		}
	}
	return acc, nil
}

// MustChainCoefficient is like ChainCoefficient but panics with the returned
// error.
func MustChainCoefficient(chain ...Tuplet) fraction.Fraction {
	return must(ChainCoefficient(chain...))
}

// ApplyChain returns d multiplied by ChainCoefficient(chain...), in canonical
// reduced form: an eighth 1/8 inside two nested triplets gives 1/18. An empty
// chain gives d in reduced form. In order of checking, it returns an error
// wrapping ErrUnset if any element is unset, ErrInvalidArgument if d is unset,
// the errors of ChainCoefficient, and ErrOutOfRange together with
// fraction.ErrOutOfRange if the reduced result is outside the range of package
// common/fraction. The chain is only read.
func ApplyChain(d fraction.Fraction, chain ...Tuplet) (fraction.Fraction, error) {
	if err := checkChain(chain); err != nil {
		return fraction.Fraction{}, err
	}
	if !d.IsValid() {
		return fraction.Fraction{}, fmt.Errorf("%w: unset fraction", ErrInvalidArgument)
	}
	c, err := chainCoefficient(chain)
	if err != nil {
		return fraction.Fraction{}, err
	}
	r, err := d.Mul(c)
	if err != nil {
		return fraction.Fraction{}, fractionRangeError("apply chain", err)
	}
	return r, nil
}

// MustApplyChain is like ApplyChain but panics with the returned error.
func MustApplyChain(d fraction.Fraction, chain ...Tuplet) fraction.Fraction {
	return must(ApplyChain(d, chain...))
}

// mulInt returns a·b and true if the product of the non-negative integers a
// and b does not exceed math.MaxInt; otherwise it returns 0 and false.
func mulInt(a, b int) (int, bool) {
	hi, lo := bits.Mul64(uint64(a), uint64(b))
	if hi != 0 || lo > math.MaxInt {
		return 0, false
	}
	return int(lo), true
}

// gcd returns the greatest common divisor of the positive integers a and b.
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
