package cents

import (
	"math"
	"math/big"
	"strconv"
)

var gridDen = big.NewInt(GridDenominator)

// shortest returns the shortest decimal representation of x, as written by
// strconv.FormatFloat(x, 'g', -1, 64), as an exact rational. x is finite.
func shortest(x float64) *big.Rat {
	r, _ := new(big.Rat).SetString(strconv.FormatFloat(x, 'g', -1, 64))
	return r
}

// gridRecord returns the record of the multiple of 1/GridDenominator cent
// nearest to x; an exact half is rounded away from zero.
func gridRecord(x *big.Rat) *record {
	rc := new(record)
	rc.r.SetFrac(roundRat(x, gridDen), gridDen)
	return rc
}

// gridLog returns the record of the multiple of 1/GridDenominator cent
// nearest to 1200·log2(r) for a positive rational r. The logarithm is
// rational only if r is a power of 2, and then it is a multiple of 1200;
// otherwise it is irrational and has no exact half.
func gridLog(r *big.Rat) *record {
	n, d := r.Num(), r.Denom()
	if isPowerOfTwo(n) && isPowerOfTwo(d) {
		rc := new(record)
		rc.r.SetInt64(int64(n.BitLen()-d.BitLen()) * 1200)
		return rc
	}
	var t big.Int
	for prec := uint(minPrec); ; prec *= 2 {
		c := centsLog(n, d, prec)
		scale := new(big.Int).Lsh(bigOne, prec)
		lo := roundHalfAway(t.Mul(t.Sub(c, big.NewInt(2)), gridDen), scale)
		hi := roundHalfAway(t.Mul(t.Add(c, big.NewInt(2)), gridDen), scale)
		if lo.Cmp(hi) == 0 {
			rc := new(record)
			rc.r.SetFrac(lo, gridDen)
			return rc
		}
	}
}

// isPowerOfTwo reports whether the positive n is a power of 2.
func isPowerOfTwo(n *big.Int) bool {
	return n.TrailingZeroBits() == uint(n.BitLen()-1)
}

// finite reports whether x is neither NaN nor an infinity.
func finite(x float64) bool {
	return !math.IsNaN(x) && !math.IsInf(x, 0)
}

// FromFloat returns the value nearest to x cents on the grid of
// 1/GridDenominator cent. x is understood as its shortest decimal
// representation (strconv.FormatFloat(x, 'g', -1, 64)), so 7.1 is exactly
// 71/10 cents, a value with at most 6 digits after the point is exact, and
// 5e-7 is exactly half a step of the grid and rounds to 1/1000000 cent. An
// exact half is rounded away from zero, symmetrically for both signs; -0
// gives 0. It returns [ErrInvalidArgument] for NaN and infinities and
// [ErrOutOfRange] if the rounded value is outside the range.
func FromFloat(x float64) (Cents, error) {
	if !finite(x) {
		return Cents{}, errorf(ErrInvalidArgument, "FromFloat(%v): not a finite number", x)
	}
	return finish(gridRecord(shortest(x)), "FromFloat")
}

// MustFromFloat is like [FromFloat] but panics with the error. It is meant
// for literals and values known to be valid.
func MustFromFloat(x float64) Cents {
	return must(FromFloat(x))
}

// FromRatioFloat returns the cents 1200·log2(r) of a frequency ratio given
// as a float64, rounded correctly to the grid of 1/GridDenominator cent. r
// is understood as its shortest decimal representation, as in [FromFloat].
// The result is an ordinary exact value on the grid: FromRatioFloat(1.5) is
// 701.955001 cents, not FromRatio(3, 2). It returns [ErrInvalidArgument] if
// r is not positive, NaN, or an infinity; subnormal positive numbers are
// accepted, and the result is always within the range.
func FromRatioFloat(r float64) (Cents, error) {
	if !finite(r) || r <= 0 {
		return Cents{}, errorf(ErrInvalidArgument, "FromRatioFloat(%v): ratio must be positive and finite", r)
	}
	return finish(gridLog(shortest(r)), "FromRatioFloat")
}

// MustFromRatioFloat is like [FromRatioFloat] but panics with the error. It
// is meant for literals and values known to be valid.
func MustFromRatioFloat(r float64) Cents {
	return must(FromRatioFloat(r))
}

// Between returns the distance from the frequency f1 to the frequency f2:
// 1200·log2(f2/f1) cents, rounded correctly to the grid of 1/GridDenominator
// cent, with the frequencies understood as their shortest decimal
// representations. Between(440, 442) is 7.851415 cents, and Between(442,
// 440) is -7.851415 cents. It returns [ErrInvalidArgument] if a frequency is
// not positive, NaN, or an infinity; the result is always within the range.
func Between(f1, f2 float64) (Cents, error) {
	if !finite(f1) || f1 <= 0 || !finite(f2) || f2 <= 0 {
		return Cents{}, errorf(ErrInvalidArgument, "Between(%v, %v): frequencies must be positive and finite", f1, f2)
	}
	r := new(big.Rat).Quo(shortest(f2), shortest(f1))
	return finish(gridLog(r), "Between")
}

// MustBetween is like [Between] but panics with the error. It is meant for
// literals and values known to be valid.
func MustBetween(f1, f2 float64) Cents {
	return must(Between(f1, f2))
}

// Float64 returns the float64 nearest to the exact value of c; a value
// exactly halfway is rounded to the even mantissa, as strconv.ParseFloat
// does. 0 gives +0. The result does not depend on the platform.
func (c Cents) Float64() float64 {
	rc := c.value()
	if rc.isRational() {
		f, _ := rc.r.Float64()
		return f
	}
	for prec := rc.startPrec(); ; prec *= 2 {
		lo, hi := rc.bounds(prec)
		a := round53(lo, -int(prec))
		if a.Cmp(round53(hi, -int(prec))) == 0 {
			f, _ := a.Float64()
			return f
		}
	}
}

// Fraction returns the reduced fraction num/den of cents, with den >= 1, if
// c is rational and both numbers fit in an int; otherwise ok is false, and
// [Cents.Composition] gives the exact record. 0 gives (0, 1, true).
func (c Cents) Fraction() (num, den int, ok bool) {
	rc := c.value()
	if !rc.isRational() || !rc.r.Num().IsInt64() || !rc.r.Denom().IsInt64() {
		return 0, 0, false
	}
	return int(rc.r.Num().Int64()), int(rc.r.Denom().Int64()), true
}

// Ratio returns the reduced frequency ratio num:den, with positive terms, if
// c is the cents of a ratio of integers and both terms fit in an int;
// otherwise ok is false. 0 gives (1, 1, true), and -1200 cents give
// (1, 2, true).
func (c Cents) Ratio() (num, den int, ok bool) {
	n, d, ok := c.value().ratioParts(63)
	if !ok {
		return 0, 0, false
	}
	return int(n.Int64()), int(d.Int64()), true
}

// Composition returns the exact record of c: the rational part num/den in
// cents, reduced, with den >= 1, and the nonzero exponents of the primes 3
// to 31 in increasing order of Base, each a reduced fraction with Num != 0
// and Den >= 1. The prime 2 is part of the rational part and never appears
// in powers; a rational value has no powers. All the numbers and the slice
// are new and belong to the caller. FromComposition(c.Composition()) returns
// a value equal to c.
func (c Cents) Composition() (num, den *big.Int, powers []Power) {
	rc := c.value()
	num = new(big.Int).Set(rc.r.Num())
	den = new(big.Int).Set(rc.r.Denom())
	powers = make([]Power, 0, numPrimes)
	for i := range rc.e {
		e := &rc.e[i]
		if e.Sign() != 0 {
			powers = append(powers, Power{
				Base: int(primes[i]),
				Num:  new(big.Int).Set(e.Num()),
				Den:  new(big.Int).Set(e.Denom()),
			})
		}
	}
	return num, den, powers
}

// Multiplier returns the frequency multiplier 2^(c/1200), correctly rounded
// to a float64 (to nearest, ties to even). It is exact for 0 (1) and for
// multiples of 1200 cents whose result is a normal float64. It returns
// [ErrOutOfRange] if the rounded result is not a normal positive finite
// float64: at least 1228800 cents overflow, and at most -1226401 cents give
// a subnormal result.
func (c Cents) Multiplier() (float64, error) {
	f, err := apply(c.value(), 1)
	if err != nil {
		return 0, errorf(ErrOutOfRange, "Multiplier: %v", err)
	}
	return f, nil
}

// MustMultiplier is like [Cents.Multiplier] but panics with the error. It is
// meant for literals and values known to be valid.
func (c Cents) MustMultiplier() float64 {
	return must(c.Multiplier())
}

// ApplyTo returns the frequency freq·2^(c/1200), with a single correct
// rounding of the exact product to a float64 (to nearest, ties to even).
// freq is its exact binary value. ApplyTo(440) is 440 for 0 and 880 for 1200
// cents, exactly. It returns [ErrInvalidArgument] if freq is not positive,
// NaN, or an infinity; a subnormal positive freq is accepted. It returns
// [ErrOutOfRange] if the rounded result is not a normal positive finite
// float64 (overflow, or a result below 2^-1022).
func (c Cents) ApplyTo(freq float64) (float64, error) {
	if !finite(freq) || freq <= 0 {
		return 0, errorf(ErrInvalidArgument, "ApplyTo(%v): frequency must be positive and finite", freq)
	}
	f, err := apply(c.value(), freq)
	if err != nil {
		return 0, errorf(ErrOutOfRange, "ApplyTo(%v): %v", freq, err)
	}
	return f, nil
}

// MustApplyTo is like [Cents.ApplyTo] but panics with the error. It is meant
// for literals and values known to be valid.
func (c Cents) MustApplyTo(freq float64) float64 {
	return must(c.ApplyTo(freq))
}

// errNotNormal is the reason of an apply result that is not a normal float64.
type errNotNormal string

func (e errNotNormal) Error() string { return string(e) }

// apply returns freq·2^(v/1200) rounded to a 53-bit mantissa without a limit
// on the exponent, or an error if the rounded result is not a normal finite
// float64. freq is positive and finite.
func apply(rc *record, freq float64) (float64, error) {
	frac, exp := math.Frexp(freq)
	m := big.NewInt(int64(frac * (1 << 53)))
	exp -= 53
	y := exactApply(rc, m, exp)
	if y == nil {
		y = approxApply(rc, m, exp)
	}
	switch e := y.MantExp(nil); {
	case e < -1021:
		return 0, errNotNormal("result below the normal float64 range")
	case e > 1024:
		return 0, errNotNormal("result overflows float64")
	}
	f, _ := y.Float64()
	return f, nil
}

// maxExactExp is the largest magnitude of a prime exponent for which apply
// computes a rational result exactly. A larger exponent makes the odd part
// of the result either not an integer or greater than 2^54, so the result is
// not a midpoint between two float64 values and the intervals of
// approxApply decide its rounding.
const maxExactExp = 64

// exactApply returns m·2^exp·2^(v/1200) rounded to 53 bits if it is rational
// and small enough to compute exactly, and nil otherwise.
func exactApply(rc *record, m *big.Int, exp int) *big.Float {
	if !rc.r.IsInt() {
		return nil
	}
	var a2, rem big.Int
	a2.QuoRem(rc.r.Num(), big1200, &rem)
	if rem.Sign() != 0 || !a2.IsInt64() {
		return nil
	}
	num := new(big.Int).Set(m)
	den := big.NewInt(1)
	for i := range rc.e {
		e := &rc.e[i]
		if !e.IsInt() || e.Num().CmpAbs(big.NewInt(maxExactExp)) > 0 {
			return nil
		}
		mulPower(num, den, primes[i], e.Num().Int64())
	}
	z := new(big.Float).SetPrec(53).SetMode(big.ToNearestEven)
	z.Quo(new(big.Float).SetInt(num), new(big.Float).SetInt(den))
	return z.SetMantExp(z, exp+int(a2.Int64()))
}

// approxApply returns m·2^exp·2^(v/1200) rounded to 53 bits, using
// intervals of growing precision. The result is not a midpoint between two
// float64 values (see maxExactExp), so the loop ends.
func approxApply(rc *record, m *big.Int, exp int) *big.Float {
	var t big.Int
	for prec := rc.startPrec(); ; prec *= 2 {
		lo, hi := rc.bounds(prec)
		// y = v/1200 lies within [ylo, yhi]/2^prec.
		ylo := lo.Div(lo, big1200)
		yhi := ceilDiv(hi, hi, big1200)
		a := powerBound(ylo, prec, m, exp, -2, &t)
		b := powerBound(yhi, prec, m, exp, 2, &t)
		if a.Cmp(b) == 0 {
			return a
		}
	}
}

// powerBound returns m·2^exp·2^(y/2^prec) rounded to 53 bits, with the power
// of two replaced by a bound: delta = -2 gives a lower bound and delta = 2
// an upper bound, by the error bound of exp2Frac.
func powerBound(y *big.Int, prec uint, m *big.Int, exp, delta int, t *big.Int) *big.Float {
	n := new(big.Int).Rsh(y, prec)
	f := t.Sub(y, new(big.Int).Lsh(n, prec))
	p := exp2Frac(f, prec)
	p.Add(p, big.NewInt(int64(delta)))
	p.Mul(p, m)
	return round53(p, exp+int(n.Int64())-int(prec))
}
