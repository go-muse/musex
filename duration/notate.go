package duration

import (
	"math/big"

	"github.com/go-muse/musex/tuplet"
)

// MaxParts is the largest number of parts that Notate and NotateWith return.
const MaxParts = 1024

// maxExactDots is the largest maximum number of dots for which NotateWith
// applies its rule directly. A duration v' = v/C of a set duration v under a
// chain with coefficient C has a numerator and a power-of-two denominator
// below 2^510. If v' < 2L for the longest base L, no remainder of the rule has
// more than 513 ones after its leading bit, so every larger maximum gives the
// same result as maxExactDots. If v' ≥ 2L, the first sign is L·(2 − 2^−D)
// and, for D > maxExactDots, the remainder keeps the addition q·L·2^−D below
// the lowest bit of v' that no sign of a base from 1024th can remove, so the
// rule gives no signs.
const maxExactDots = 1020

// sign is a written sign of the rule of NotateWith: a base and its dots.
type sign struct {
	base Base
	dots int
}

// toWritten implements Duration.ToWritten and Tied.ToWritten for a set d.
func (d Duration) toWritten(op string, chain []tuplet.Tuplet) (Written, error) {
	coef, err := checkedCoefficient(op, chain)
	if err != nil {
		return Written{}, err
	}
	v := ratOf(d.v)
	v.Quo(v, coef)
	num, den := v.Num(), v.Denom()
	if !isPow2(den) {
		return Written{}, simpleError(ErrNotNotatable, op)
	}
	// v' = a·2^j with odd a; one sign with k dots requires a = 2^(k+1) − 1,
	// and its base is 2^(j+k).
	shift := num.TrailingZeroBits()
	a := new(big.Int).Rsh(num, shift)
	a.Add(a, big.NewInt(1))
	if !isPow2(a) {
		return Written{}, simpleError(ErrNotNotatable, op)
	}
	k := a.BitLen() - 2
	j := int(shift) - (den.BitLen() - 1)
	b, ok := baseOfExp(j + k)
	if !ok {
		return Written{}, simpleError(ErrNotNotatable, op)
	}
	return signWritten(op, sign{base: b, dots: k}, packChain(chain), coef)
}

// signWritten returns the written duration of s under a packed chain with
// coefficient coef.
func signWritten(op string, s sign, chain string, coef *big.Rat) (Written, error) {
	v, err := writtenValue(op, s.base, s.dots, coef)
	if err != nil {
		return Written{}, err
	}
	return Written{base: s.base, dots: s.dots, chain: chain, value: v}, nil
}

// ToWritten returns the single sign with the duration d under chain, listed
// from the outermost to the innermost tuplet (an empty chain means no
// tuplets): a base b from the series with k ≥ 0 dots such that
// d/C = b·(2 − 1/2^k), where C is the chain coefficient. Such a sign is
// unique: 3/8 gives a dotted quarter, and 1/3 under a triplet gives a half in
// 3:2. The package does not choose a chain itself. In order of checking, it
// returns an error wrapping ErrUnset if d is unset, ErrInvalidArgument
// together with tuplet.ErrUnset if a tuplet is unset, ErrOutOfRange if the
// chain coefficient is outside the range of package common/fraction, and
// ErrNotNotatable if no such sign exists, as for 5/16. The chain is only read.
func (d Duration) ToWritten(chain ...tuplet.Tuplet) (Written, error) {
	const op = "Duration.ToWritten"
	if !d.IsSet() {
		return Written{}, unsetError(op, "receiver")
	}
	return d.toWritten(op, chain)
}

// MustToWritten is like ToWritten but panics with the returned error.
func (d Duration) MustToWritten(chain ...tuplet.Tuplet) Written {
	return must(d.ToWritten(chain...))
}

// Notate is NotateWith(Breve, 0, chain...): the default rule of notation with
// ties by bases without dots from the breve down, so 5/16 gives a quarter and
// a sixteenth.
func (d Duration) Notate(chain ...tuplet.Tuplet) ([]Written, error) {
	return d.notateWith("Duration.Notate", Breve, 0, chain)
}

// MustNotate is like Notate but panics with the returned error.
func (d Duration) MustNotate(chain ...tuplet.Tuplet) []Written {
	return must(d.Notate(chain...))
}

// NotateWith writes d as signs to be tied, under chain. The allowed signs are
// the bases from longest down to the 1024th with 0 through maxDots dots.
// Starting with the remainder d/C, where C is the chain coefficient, each
// next part is the longest allowed sign not longer than the remainder, so the
// parts go from long to short; every part carries the chain as given. If d is
// one allowed sign, the result is that sign alone. The result is a new slice
// of 1 through MaxParts written durations that belongs to the caller; a result
// of two or more parts makes a Tied with NewTied.
//
// ErrNotNotatable means that this rule gives no notation, not that none
// exists: 1/512 with longest ThousandTwentyFourth and maxDots 1 fails although
// two tied 1024ths make it. In order of checking, it returns an error wrapping
// ErrUnset if d is unset; ErrInvalidArgument if longest is not a base, if
// maxDots < 0, or if a tuplet is unset (together with tuplet.ErrUnset);
// ErrOutOfRange if the chain coefficient is outside the range of package
// common/fraction; ErrNotNotatable if the rule does not finish; ErrTooManyParts
// if it would need more than MaxParts parts, which is detected without
// building them; and ErrOutOfRange if the duration of a part is outside the
// range. The chain is only read.
func (d Duration) NotateWith(longest Base, maxDots int, chain ...tuplet.Tuplet) ([]Written, error) {
	return d.notateWith("Duration.NotateWith", longest, maxDots, chain)
}

// MustNotateWith is like NotateWith but panics with the returned error.
func (d Duration) MustNotateWith(longest Base, maxDots int, chain ...tuplet.Tuplet) []Written {
	return must(d.NotateWith(longest, maxDots, chain...))
}

// notateWith implements Notate and NotateWith.
func (d Duration) notateWith(op string, longest Base, maxDots int, chain []tuplet.Tuplet) ([]Written, error) {
	if !d.IsSet() {
		return nil, unsetError(op, "receiver")
	}
	if !longest.IsValid() {
		return nil, invalidBaseError(op, longest)
	}
	if maxDots < 0 {
		return nil, invalidError(op, "negative maximum number of dots")
	}
	coef, err := checkedCoefficient(op, chain)
	if err != nil {
		return nil, err
	}
	v := ratOf(d.v)
	v.Quo(v, coef)
	first, count, tail, err := notatePlan(op, v, longest, maxDots)
	if err != nil {
		return nil, err
	}
	packed := packChain(chain)
	parts := make([]Written, 0, count+len(tail))
	if count > 0 {
		w, err := signWritten(op, first, packed, coef)
		if err != nil {
			return nil, err
		}
		for range count {
			parts = append(parts, w)
		}
	}
	for _, s := range tail {
		w, err := signWritten(op, s, packed, coef)
		if err != nil {
			return nil, err
		}
		parts = append(parts, w)
	}
	return parts, nil
}

// notatePlan applies the rule of NotateWith to v > 0 without building parts.
// It returns the longest allowed sign first with the number count of its
// repetitions in the first phase, and the signs of the remainder in order. It
// returns an error wrapping ErrNotNotatable if the rule does not finish and
// ErrTooManyParts if it needs more than MaxParts parts.
func notatePlan(op string, v *big.Rat, longest Base, maxDots int) (first sign, count int, tail []sign, err error) {
	if !isPow2(v.Denom()) {
		return sign{}, 0, nil, simpleError(ErrNotNotatable, op)
	}
	eL := longest.exp()
	dots := maxDots
	if dots > maxExactDots {
		if v.Cmp(pow2Rat(eL+1)) >= 0 {
			return sign{}, 0, nil, simpleError(ErrNotNotatable, op)
		}
		dots = maxExactDots
	}
	first = sign{base: longest, dots: dots}
	// First phase: the longest allowed sign while the remainder is not
	// shorter, counted by division.
	maxSign := pow2Rat(eL)
	maxSign.Mul(maxSign, dotFactor(dots))
	q := floorRat(new(big.Rat).Quo(v, maxSign))
	r := new(big.Rat).Sub(v, new(big.Rat).Mul(maxSign, new(big.Rat).SetInt(q)))
	// Tail: each step removes at least the leading bit of the remainder, which
	// starts below 2L and must stay at or above the 1024th.
	for r.Sign() > 0 {
		s, ok := tailSign(r, eL, dots)
		if !ok {
			return sign{}, 0, nil, simpleError(ErrNotNotatable, op)
		}
		tail = append(tail, s)
		value := pow2Rat(s.base.exp())
		r.Sub(r, value.Mul(value, dotFactor(s.dots)))
	}
	if !q.IsInt64() || q.Int64() > MaxParts || int(q.Int64())+len(tail) > MaxParts {
		return sign{}, 0, nil, simpleError(ErrTooManyParts, op)
	}
	return first, int(q.Int64()), tail, nil
}

// tailSign returns the longest allowed sign not longer than the dyadic
// remainder r > 0, for bases from 2^eL down to the 1024th with at most dots
// dots, and reports whether it exists.
func tailSign(r *big.Rat, eL, dots int) (sign, bool) {
	num, den := r.Num(), r.Denom()
	lead := num.BitLen() - den.BitLen() // binary exponent of the leading bit
	p := min(lead, eL)
	b, ok := baseOfExp(p)
	if !ok {
		return sign{}, false
	}
	k := dots
	if p == lead {
		// r/2^p = 1.11…10…: the number of ones right after the leading one.
		m := 0
		for i := num.BitLen() - 2; i >= 0 && num.Bit(i) == 1; i-- {
			m++
		}
		k = min(m, dots)
	}
	return sign{base: b, dots: k}, true
}
