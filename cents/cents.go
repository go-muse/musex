package cents

import "math/big"

// Limits of the package.
const (
	// MaxMagnitude is the largest magnitude of a value in cents: every value
	// v satisfies -MaxMagnitude <= v <= MaxMagnitude. It is a technical
	// limit, not a musical one: 9,000,000 cents are 7500 octaves.
	MaxMagnitude = 9_000_000

	// MaxPrime is the largest prime of an exact value (the prime limit):
	// ratios of integers whose prime factors are at most 31 are exact.
	MaxPrime = 31

	// GridDenominator is the denominator of the grid of approximate values:
	// a value created from a float64 is a multiple of 1/GridDenominator cent.
	GridDenominator = 1_000_000

	// MaxDigits is the longest run of digits accepted in a text, leading
	// zeros included.
	MaxDigits = 100

	// MaxFactors is the largest number of factors of a product of powers
	// accepted in a text.
	MaxFactors = 64
)

// Cents is a signed pitch distance in cents: 1200 cents are an octave, and
// 100 cents are an equal-tempered semitone. The same value serves as the
// correction of a note, as a value of a temperament, and as a distance.
//
// A value is exact: its record is a rational number of cents plus rational
// multiples of the sizes of the primes 3 to 31 (see the package
// documentation). Every value has exactly one internal representation, so
// a == b if and only if a.Equal(b), and Cents can be used as a map key.
//
// The zero value is 0 cents, the same as [Zero]. Values are immutable and
// can be copied by assignment and shared between goroutines.
type Cents struct {
	rec string
}

// value returns the record of c.
func (c Cents) value() *record {
	return decode(c.rec)
}

// fromRecord returns the value of the record.
func fromRecord(rc *record) Cents {
	return Cents{rec: rc.encode()}
}

// finish checks the record limits and the range of rc and returns its value.
func finish(rc *record, op string) (Cents, error) {
	if err := rc.check(op); err != nil {
		return Cents{}, err
	}
	return fromRecord(rc), nil
}

// FromInt returns n cents. It returns [ErrOutOfRange] if |n| > MaxMagnitude.
func FromInt(n int) (Cents, error) {
	if n > MaxMagnitude || n < -MaxMagnitude {
		return Cents{}, errorf(ErrOutOfRange, "FromInt(%d): value outside ±%d cents", n, MaxMagnitude)
	}
	rc := new(record)
	rc.r.SetInt64(int64(n))
	return fromRecord(rc), nil
}

// MustFromInt is like [FromInt] but panics with the error. It is meant for
// literals and values known to be valid.
func MustFromInt(n int) Cents {
	return must(FromInt(n))
}

// FromFraction returns num/den cents, exactly. It returns
// [ErrInvalidArgument] if den is 0 (0/0 included) and [ErrOutOfRange] if the
// value is outside the range.
func FromFraction(num, den int) (Cents, error) {
	if den == 0 {
		return Cents{}, errorf(ErrInvalidArgument, "FromFraction(%d, %d): zero denominator", num, den)
	}
	rc := new(record)
	rc.r.SetFrac(big.NewInt(int64(num)), big.NewInt(int64(den)))
	return finish(rc, "FromFraction")
}

// MustFromFraction is like [FromFraction] but panics with the error. It is
// meant for literals and values known to be valid.
func MustFromFraction(num, den int) Cents {
	return must(FromFraction(num, den))
}

// Step returns the k-th step of the equal division of the octave into n
// steps (n-EDO), that is, k·1200/n cents, exactly. The unit "1/n of an
// octave" is Step(1, n). It returns [ErrInvalidArgument] if n < 1 and
// [ErrOutOfRange] if the value is outside the range.
func Step(k, n int) (Cents, error) {
	if n < 1 {
		return Cents{}, errorf(ErrInvalidArgument, "Step(%d, %d): number of steps less than 1", k, n)
	}
	rc := new(record)
	num := new(big.Int).Mul(big.NewInt(int64(k)), big1200)
	rc.r.SetFrac(num, big.NewInt(int64(n)))
	return finish(rc, "Step")
}

// MustStep is like [Step] but panics with the error. It is meant for
// literals and values known to be valid.
func MustStep(k, n int) Cents {
	return must(Step(k, n))
}

// FromRatio returns the cents of the frequency ratio num:den, that is,
// 1200·log2(num/den), exactly. The ratio is reduced first, so 111:74 is
// 3:2. It returns [ErrInvalidArgument] if num or den is not positive and
// [ErrPrimeLimit] if the reduced ratio has a prime factor greater than 31.
// The result is always within the range.
func FromRatio(num, den int) (Cents, error) {
	if num <= 0 || den <= 0 {
		return Cents{}, errorf(ErrInvalidArgument, "FromRatio(%d, %d): ratio terms must be positive", num, den)
	}
	rc, err := ratioRecord(big.NewInt(int64(num)), big.NewInt(int64(den)), "FromRatio")
	if err != nil {
		return Cents{}, err
	}
	return finish(rc, "FromRatio")
}

// MustFromRatio is like [FromRatio] but panics with the error. It is meant
// for literals and values known to be valid.
func MustFromRatio(num, den int) Cents {
	return must(FromRatio(num, den))
}

// Power is a factor of a composition: the prime Base raised to the exponent
// Num/Den. A nil Num means 0, and a nil Den means 1. Base 2 adds 1200·Num/Den
// cents to the rational part; a base 3 to 31 adds Num/Den times the size of
// the prime.
type Power struct {
	Base     int      // 2 or a prime 3...31
	Num, Den *big.Int // exponent Num/Den; nil Num means 0, nil Den means 1
}

// FromComposition returns the value num/den + Σ (Num/Den)·1200·log2(Base)
// cents of a composition: a rational part num/den in cents and powers of the
// primes up to 31. A nil num means 0, and a nil den means 1. A base may
// repeat; its exponents are added. The arguments are neither modified nor
// retained.
//
// The errors are checked in this order. [ErrInvalidArgument]: a zero
// denominator, or a base that is not a prime (less than 2, negative, or
// composite, such as 4 or 9). [ErrPrimeLimit]: a prime base greater than 31.
// A base is checked whatever its exponent, so Power{Base: 37} gives
// ErrPrimeLimit and Power{Base: 4} gives ErrInvalidArgument.
// [ErrTooComplex]: the record of the result is beyond the record limits.
// [ErrOutOfRange]: the value is outside the range.
func FromComposition(num, den *big.Int, powers ...Power) (Cents, error) {
	if den != nil && den.Sign() == 0 {
		return Cents{}, errorf(ErrInvalidArgument, "FromComposition: zero denominator of the rational part")
	}
	for _, p := range powers {
		if p.Den != nil && p.Den.Sign() == 0 {
			return Cents{}, errorf(ErrInvalidArgument, "FromComposition: zero denominator of the exponent of %d", p.Base)
		}
		if !isPrime(p.Base) {
			return Cents{}, errorf(ErrInvalidArgument, "FromComposition: base %d is not a prime", p.Base)
		}
	}
	for _, p := range powers {
		if p.Base > MaxPrime {
			return Cents{}, errorf(ErrPrimeLimit, "FromComposition: prime base %d greater than %d", p.Base, MaxPrime)
		}
	}
	rc := new(record)
	if num != nil {
		d := bigOne
		if den != nil {
			d = den
		}
		rc.r.SetFrac(num, d)
	}
	var e big.Rat
	for _, p := range powers {
		if p.Num == nil {
			continue
		}
		d := bigOne
		if p.Den != nil {
			d = p.Den
		}
		e.SetFrac(p.Num, d)
		if p.Base == 2 {
			e.Mul(&e, rat1200)
			rc.r.Add(&rc.r, &e)
			continue
		}
		i := primeIndex(int64(p.Base))
		rc.e[i].Add(&rc.e[i], &e)
	}
	return finish(rc, "FromComposition")
}

// MustFromComposition is like [FromComposition] but panics with the error.
// It is meant for literals and values known to be valid.
func MustFromComposition(num, den *big.Int, powers ...Power) Cents {
	return must(FromComposition(num, den, powers...))
}

// isPrime reports whether n is a prime.
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	return big.NewInt(int64(n)).ProbablyPrime(0) // exact below 2^64
}

// Add returns c + d. It returns [ErrTooComplex] if the record of the result
// is beyond the record limits and [ErrOutOfRange] if the result is outside
// the range.
func (c Cents) Add(d Cents) (Cents, error) {
	return finish(add(c.value(), d.value()), "Add")
}

// MustAdd is like [Cents.Add] but panics with the error. It is meant for
// literals and values known to be valid.
func (c Cents) MustAdd(d Cents) Cents {
	return must(c.Add(d))
}

// Sub returns c - d. It returns [ErrTooComplex] if the record of the result
// is beyond the record limits and [ErrOutOfRange] if the result is outside
// the range.
func (c Cents) Sub(d Cents) (Cents, error) {
	return finish(sub(c.value(), d.value()), "Sub")
}

// MustSub is like [Cents.Sub] but panics with the error. It is meant for
// literals and values known to be valid.
func (c Cents) MustSub(d Cents) Cents {
	return must(c.Sub(d))
}

// Neg returns -c. The range is symmetric, so there is no error.
func (c Cents) Neg() Cents {
	return fromRecord(neg(c.value()))
}

// Abs returns |c|.
func (c Cents) Abs() Cents {
	if c.Sign() < 0 {
		return c.Neg()
	}
	return c
}

// Mul returns c·k for any int k. It returns [ErrTooComplex] if a prime
// exponent of the result exceeds the record limits and [ErrOutOfRange] if
// the result is outside the range; Zero().Mul(math.MaxInt) is 0.
func (c Cents) Mul(k int) (Cents, error) {
	return finish(scale(c.value(), new(big.Rat).SetInt64(int64(k))), "Mul")
}

// MustMul is like [Cents.Mul] but panics with the error. It is meant for
// literals and values known to be valid.
func (c Cents) MustMul(k int) Cents {
	return must(c.Mul(k))
}

// Div returns c/n, exactly. It returns [ErrInvalidArgument] if n is 0 and
// [ErrTooComplex] if a denominator of the result exceeds the record limits.
// c.Div(-1) is c.Neg().
func (c Cents) Div(n int) (Cents, error) {
	if n == 0 {
		return Cents{}, errorf(ErrInvalidArgument, "Div(0): zero divisor")
	}
	return finish(scale(c.value(), new(big.Rat).SetFrac64(1, int64(n))), "Div")
}

// MustDiv is like [Cents.Div] but panics with the error. It is meant for
// literals and values known to be valid.
func (c Cents) MustDiv(n int) Cents {
	return must(c.Div(n))
}

// Scale returns c·num/den in one exact operation, without an intermediate
// range check. Its main use is the value of an amount of a unit:
// Millioctave().Scale(1750, 3) is 700 cents. c.Mul(k) is c.Scale(k, 1), and
// c.Div(n) is c.Scale(1, n). It returns [ErrInvalidArgument] if den is 0,
// then [ErrTooComplex] and [ErrOutOfRange].
func (c Cents) Scale(num, den int) (Cents, error) {
	if den == 0 {
		return Cents{}, errorf(ErrInvalidArgument, "Scale(%d, %d): zero denominator", num, den)
	}
	k := new(big.Rat).SetFrac(big.NewInt(int64(num)), big.NewInt(int64(den)))
	return finish(scale(c.value(), k), "Scale")
}

// MustScale is like [Cents.Scale] but panics with the error. It is meant for
// literals and values known to be valid.
func (c Cents) MustScale(num, den int) Cents {
	return must(c.Scale(num, den))
}

// ScaleRat returns c·r for any rational r, such as an amount of a unit
// returned by [Quotient.Exact]. A nil r means 0. r is neither modified nor
// retained. It returns [ErrTooComplex] and [ErrOutOfRange].
func (c Cents) ScaleRat(r *big.Rat) (Cents, error) {
	k := new(big.Rat)
	if r != nil {
		k.Set(r)
	}
	return finish(scale(c.value(), k), "ScaleRat")
}

// MustScaleRat is like [Cents.ScaleRat] but panics with the error. It is
// meant for literals and values known to be valid.
func (c Cents) MustScaleRat(r *big.Rat) Cents {
	return must(c.ScaleRat(r))
}

// RoundTo returns the multiple of 1/d cent nearest to c; an exact half is
// rounded away from zero, symmetrically for both signs. The result is
// rational with a denominator that divides d. It returns
// [ErrInvalidArgument] if d < 1.
func (c Cents) RoundTo(d int) (Cents, error) {
	if d < 1 {
		return Cents{}, errorf(ErrInvalidArgument, "RoundTo(%d): denominator less than 1", d)
	}
	bd := big.NewInt(int64(d))
	m := c.value().roundValue(bd, bigOne)
	rc := new(record)
	rc.r.SetFrac(m, bd)
	return fromRecord(rc), nil
}

// MustRoundTo is like [Cents.RoundTo] but panics with the error. It is meant
// for literals and values known to be valid.
func (c Cents) MustRoundTo(d int) Cents {
	return must(c.RoundTo(d))
}

// SplitOctaves returns octaves and rest with c = 1200·octaves + rest and
// 0 <= rest < 1200 cents, so octaves = floor(c/1200). -100 cents are -1
// octave and 1100 cents.
func (c Cents) SplitOctaves() (octaves int, rest Cents) {
	rc := c.value()
	q := rc.floorValue(big1200)
	shift := new(big.Rat).SetInt(new(big.Int).Mul(q, big1200))
	rc.r.Sub(&rc.r, shift)
	return int(q.Int64()), fromRecord(rc)
}

// NearestStep returns the number k of the step of the n-EDO nearest to c,
// that is, the integer nearest to c·n/1200, with an exact half rounded away
// from zero, and the deviation c - k·1200/n, exactly, with
// |deviation| <= 600/n cents.
//
// It returns [ErrInvalidArgument] if n < 1. It returns [ErrTooComplex] if the
// reduced denominator of the rational part of the deviation exceeds 10^30:
// that denominator divides d·n, where d is the denominator of the rational
// part of c. It returns [ErrOutOfRange] if k does not fit in an int.
func (c Cents) NearestStep(n int) (k int, deviation Cents, err error) {
	if n < 1 {
		return 0, Cents{}, errorf(ErrInvalidArgument, "NearestStep(%d): number of steps less than 1", n)
	}
	rc := c.value()
	bn := big.NewInt(int64(n))
	bk := rc.roundValue(bn, big1200)
	step := new(big.Rat).SetFrac(new(big.Int).Mul(bk, big1200), bn)
	rc.r.Sub(&rc.r, step)
	if !rc.withinLimits() {
		return 0, Cents{}, errorf(ErrTooComplex, "NearestStep(%d): deviation beyond the record limits", n)
	}
	if !bk.IsInt64() {
		return 0, Cents{}, errorf(ErrOutOfRange, "NearestStep(%d): step number does not fit in int", n)
	}
	return int(bk.Int64()), fromRecord(rc), nil
}

// MustNearestStep is like [Cents.NearestStep] but panics with the error. It
// is meant for literals and values known to be valid.
func (c Cents) MustNearestStep(n int) (k int, deviation Cents) {
	k, deviation, err := c.NearestStep(n)
	if err != nil {
		panic(err)
	}
	return k, deviation
}

// Equal reports whether c and d are the same number of cents, however they
// were obtained. It is the same as c == d and does not allocate.
func (c Cents) Equal(d Cents) bool {
	return c.rec == d.rec
}

// Compare returns -1, 0, or +1 if c is less than, equal to, or greater than
// d, comparing the exact values, so it can be used with [slices.SortFunc].
// Compare(d) == 0 if and only if c.Equal(d).
func (c Cents) Compare(d Cents) int {
	if c.rec == d.rec {
		return 0
	}
	return sub(c.value(), d.value()).sign()
}

// Sign returns -1, 0, or +1 by the sign of c.
func (c Cents) Sign() int {
	return c.value().sign()
}

// Near reports whether |c - d| <= tolerance, comparing the exact values; the
// difference is not limited by the range. Near(d, Zero()) is Equal(d). It
// returns [ErrInvalidArgument] if the tolerance is negative.
func (c Cents) Near(d, tolerance Cents) (bool, error) {
	t := tolerance.value()
	if t.sign() < 0 {
		return false, errorf(ErrInvalidArgument, "Near: negative tolerance")
	}
	diff := sub(c.value(), d.value())
	return sub(t, diff).sign() >= 0 && add(t, diff).sign() >= 0, nil
}

// MustNear is like [Cents.Near] but panics with the error. It is meant for
// literals and values known to be valid.
func (c Cents) MustNear(d, tolerance Cents) bool {
	return must(c.Near(d, tolerance))
}

// must returns v or panics with err.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
