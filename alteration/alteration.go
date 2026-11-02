package alteration

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
)

// The guaranteed range of alteration values.
//
// A value p/q in lowest terms, with q > 0, is in range if |p/q| <=
// MaxMagnitude and q <= MaxDenominator. Every [Alteration] value is in range;
// operations whose exact result is outside it return an error that matches
// [ErrOutOfRange].
//
// Both constants are untyped, so they can be used in int arithmetic and
// comparisons, for example FromInt(MaxMagnitude) or
// FromFraction(1, MaxDenominator).
const (
	// MaxMagnitude is the largest absolute value, in chromatic units, of a
	// value in range: 2^24, that is, 16777216.
	MaxMagnitude = 1 << 24

	// MaxDenominator is the largest denominator, in lowest terms, of a value
	// in range: 2^16, that is, 65536. The smallest positive value with this
	// denominator is FromFraction(1, MaxDenominator).
	MaxDenominator = 1 << 16
)

// ErrZeroDenominator reports a fraction with a zero denominator.
//
// [FromFraction] returns errors that match ErrZeroDenominator with
// [errors.Is]; the text of the error is not part of the API.
var ErrZeroDenominator = errors.New("alteration: zero denominator")

// ErrOutOfRange reports that the exact value of a result is outside the
// guaranteed range set by [MaxMagnitude] and [MaxDenominator].
//
// [FromInt], [FromFraction], [Alteration.Raise], [Alteration.Lower],
// [Alteration.RaiseBy], [Alteration.LowerBy], [Alteration.Add],
// [Alteration.Sub], and [Alteration.Mul] return errors that match
// ErrOutOfRange with [errors.Is]. Such an error never matches
// [ErrZeroDenominator]. The text of the error is not part of the API.
var ErrOutOfRange = errors.New("alteration: value out of range")

// Range limits as int64 for internal arithmetic.
const (
	maxMagnitude   int64 = MaxMagnitude
	maxDenominator int64 = MaxDenominator
)

// Alteration is an alteration value: an exact rational number of chromatic
// units in the guaranteed range, such as 0 for a natural, +1 for a sharp, or
// +1/2 for a quarter tone up.
//
// The zero value is the natural (0) and equals [Natural]. Every value of the
// type is a valid alteration value; values are created by [Natural],
// [Sharp], [Flat], [DoubleSharp], [DoubleFlat], [FromInt], [FromFraction],
// their Must variants, and the results of operations.
//
// Each number has exactly one representation, so == and [Alteration.Equal]
// agree for all values. Alteration is a comparable value type of fixed size:
// it can be copied freely, and a copy shares no data with the original. It
// can be used as a map key or in a switch, for example case
// alteration.Sharp():. Methods never modify the receiver or the arguments.
//
// Alteration has no exported fields and no serialization methods: encoding/json
// outputs it as {}, and other encoders may return an error. Use
// [Alteration.Fraction] with [FromFraction] for storage and exchange;
// [Alteration.String] is for output only.
type Alteration struct {
	// num is the numerator in lowest terms; its sign is the sign of the value.
	num int64
	// denMinus1 is the denominator in lowest terms minus 1, so that the zero
	// struct means 0/1.
	denMinus1 int64
}

// Natural returns the natural: the value 0. It equals the zero value of
// [Alteration].
func Natural() Alteration {
	return Alteration{}
}

// Sharp returns the sharp: the value +1.
func Sharp() Alteration {
	return Alteration{num: 1}
}

// Flat returns the flat: the value -1.
func Flat() Alteration {
	return Alteration{num: -1}
}

// DoubleSharp returns the double sharp: the value +2.
func DoubleSharp() Alteration {
	return Alteration{num: 2}
}

// DoubleFlat returns the double flat: the value -2.
func DoubleFlat() Alteration {
	return Alteration{num: -2}
}

// FromInt returns the whole alteration value n, in chromatic units.
//
// If |n| > [MaxMagnitude], FromInt returns the zero value and an error that
// matches [ErrOutOfRange]. Any int is accepted, including [math.MinInt].
func FromInt(n int) (Alteration, error) {
	a, ok := normalize(n < 0, absInt(n), 1)
	if !ok {
		return Alteration{}, fmt.Errorf("%w: %d", ErrOutOfRange, n)
	}

	return a, nil
}

// MustFromInt is like [FromInt] but panics if n is out of range. The panic
// value is the error returned by FromInt, which matches [ErrOutOfRange].
//
// It is intended for literals and values known to be in range.
func MustFromInt(n int) Alteration {
	return must(FromInt(n))
}

// FromFraction returns the alteration value num/den, in chromatic units.
//
// The fraction is stored in lowest terms with a positive denominator:
// FromFraction(2, 4), FromFraction(-1, -2), and FromFraction(1, 2) are the
// same value, and FromFraction(1, -2) is -1/2. Any int arguments are
// accepted, including [math.MinInt]; the result is exact.
//
// If den is 0, FromFraction returns the zero value and an error that matches
// [ErrZeroDenominator], whatever num is. Otherwise, if the value in lowest
// terms is out of range, it returns the zero value and an error that matches
// [ErrOutOfRange].
//
// For any value a, FromFraction(a.Fraction()) equals a.
func FromFraction(num, den int) (Alteration, error) {
	if den == 0 {
		return Alteration{}, fmt.Errorf("%w: %d/%d", ErrZeroDenominator, num, den)
	}

	a, ok := normalize((num < 0) != (den < 0), absInt(num), absInt(den))
	if !ok {
		return Alteration{}, fmt.Errorf("%w: %d/%d", ErrOutOfRange, num, den)
	}

	return a, nil
}

// MustFromFraction is like [FromFraction] but panics if den is 0 or the value
// is out of range. The panic value is the error returned by FromFraction,
// which matches [ErrZeroDenominator] or [ErrOutOfRange].
//
// It is intended for literals and values known to be in range, for example
// alteration.MustFromFraction(1, 2) for a quarter tone up.
func MustFromFraction(num, den int) Alteration {
	return must(FromFraction(num, den))
}

// Raise returns a raised by one chromatic unit: a + 1. It equals
// a.RaiseBy(1) and a.Add(Sharp()).
//
// If the result is out of range, Raise returns the zero value and an error
// that matches [ErrOutOfRange]; a is not changed.
func (a Alteration) Raise() (Alteration, error) {
	return a.RaiseBy(1)
}

// MustRaise is like [Alteration.Raise] but panics if the result is out of
// range. The panic value is the error returned by Raise, which matches
// [ErrOutOfRange].
//
// It is intended for values known to be in range.
func (a Alteration) MustRaise() Alteration {
	return must(a.Raise())
}

// Lower returns a lowered by one chromatic unit: a - 1. It equals
// a.LowerBy(1) and a.Sub(Sharp()).
//
// If the result is out of range, Lower returns the zero value and an error
// that matches [ErrOutOfRange]; a is not changed.
func (a Alteration) Lower() (Alteration, error) {
	return a.LowerBy(1)
}

// MustLower is like [Alteration.Lower] but panics if the result is out of
// range. The panic value is the error returned by Lower, which matches
// [ErrOutOfRange].
//
// It is intended for values known to be in range.
func (a Alteration) MustLower() Alteration {
	return must(a.Lower())
}

// RaiseBy returns a raised by n chromatic units: a + n. A negative n lowers
// a. Any int is accepted, including [math.MinInt] and [math.MaxInt]; the
// result is exact.
//
// If the result is out of range, RaiseBy returns the zero value and an error
// that matches [ErrOutOfRange]; a is not changed. For any n other than
// math.MinInt, a.RaiseBy(n) equals a.LowerBy(-n).
func (a Alteration) RaiseBy(n int) (Alteration, error) {
	r, ok := a.shift(n < 0, absInt(n))
	if !ok {
		return Alteration{}, fmt.Errorf("%w: %v raised by %d", ErrOutOfRange, a, n)
	}

	return r, nil
}

// MustRaiseBy is like [Alteration.RaiseBy] but panics if the result is out of
// range. The panic value is the error returned by RaiseBy, which matches
// [ErrOutOfRange].
//
// It is intended for values known to be in range.
func (a Alteration) MustRaiseBy(n int) Alteration {
	return must(a.RaiseBy(n))
}

// LowerBy returns a lowered by n chromatic units: a - n. A negative n raises
// a. Any int is accepted, including [math.MinInt] and [math.MaxInt]; the
// result is exact.
//
// If the result is out of range, LowerBy returns the zero value and an error
// that matches [ErrOutOfRange]; a is not changed.
func (a Alteration) LowerBy(n int) (Alteration, error) {
	r, ok := a.shift(n > 0, absInt(n))
	if !ok {
		return Alteration{}, fmt.Errorf("%w: %v lowered by %d", ErrOutOfRange, a, n)
	}

	return r, nil
}

// MustLowerBy is like [Alteration.LowerBy] but panics if the result is out of
// range. The panic value is the error returned by LowerBy, which matches
// [ErrOutOfRange].
//
// It is intended for values known to be in range.
func (a Alteration) MustLowerBy(n int) Alteration {
	return must(a.LowerBy(n))
}

// Add returns a raised by the alteration value d: a + d. The value d may be
// whole or fractional, positive or negative. For a whole d equal to n,
// a.Add(d) equals a.RaiseBy(n).
//
// The result is exact and in lowest terms. If its magnitude or its
// denominator is out of range, Add returns the zero value and an error that
// matches [ErrOutOfRange]; a and d are not changed.
func (a Alteration) Add(d Alteration) (Alteration, error) {
	r, ok := fromInt64(a.num*d.den()+d.num*a.den(), a.den()*d.den())
	if !ok {
		return Alteration{}, fmt.Errorf("%w: %v + %v", ErrOutOfRange, a, d)
	}

	return r, nil
}

// MustAdd is like [Alteration.Add] but panics if the result is out of range.
// The panic value is the error returned by Add, which matches
// [ErrOutOfRange].
//
// It is intended for values known to be in range.
func (a Alteration) MustAdd(d Alteration) Alteration {
	return must(a.Add(d))
}

// Sub returns a lowered by the alteration value d: a - d. It is also the
// difference between a and d: how much higher a is than d. For example,
// Sharp().Sub(Flat()) is +2.
//
// The result is exact and in lowest terms. If its magnitude or its
// denominator is out of range, Sub returns the zero value and an error that
// matches [ErrOutOfRange]; a and d are not changed.
func (a Alteration) Sub(d Alteration) (Alteration, error) {
	r, ok := fromInt64(a.num*d.den()-d.num*a.den(), a.den()*d.den())
	if !ok {
		return Alteration{}, fmt.Errorf("%w: %v - %v", ErrOutOfRange, a, d)
	}

	return r, nil
}

// MustSub is like [Alteration.Sub] but panics if the result is out of range.
// The panic value is the error returned by Sub, which matches
// [ErrOutOfRange].
//
// It is intended for values known to be in range.
func (a Alteration) MustSub(d Alteration) Alteration {
	return must(a.Sub(d))
}

// Mul returns a multiplied by the whole number k: a * k. Any int is accepted,
// including [math.MinInt] and [math.MaxInt]; the result is exact. For
// example, the natural multiplied by any k is the natural, and +1/2
// multiplied by 2 is +1.
//
// The denominator of the result never exceeds the denominator of a, so the
// result is out of range only by magnitude. In that case Mul returns the zero
// value and an error that matches [ErrOutOfRange]; a is not changed.
func (a Alteration) Mul(k int) (Alteration, error) {
	r, ok := a.mul(k < 0, absInt(k))
	if !ok {
		return Alteration{}, fmt.Errorf("%w: %v * %d", ErrOutOfRange, a, k)
	}

	return r, nil
}

// MustMul is like [Alteration.Mul] but panics if the result is out of range.
// The panic value is the error returned by Mul, which matches
// [ErrOutOfRange].
//
// It is intended for values known to be in range.
func (a Alteration) MustMul(k int) Alteration {
	return must(a.Mul(k))
}

// Equal reports whether a and other are the same number, however they were
// obtained: MustFromFraction(2, 4).Equal(MustFromFraction(1, 2)) is true.
//
// a.Equal(other) equals a == other for all values.
func (a Alteration) Equal(other Alteration) bool {
	return a == other
}

// Compare compares a and other by size. It returns -1 if a < other, 0 if a
// equals other, and +1 if a > other. The comparison is exact for all values.
//
// Compare orders alteration values as numbers, not pitches of notes. It can
// be used with [slices.SortFunc] as alteration.Alteration.Compare.
func (a Alteration) Compare(other Alteration) int {
	// |num| <= 2^40 and den <= 2^16, so the products fit in int64.
	return cmp.Compare(a.num*other.den(), other.num*a.den())
}

// Sign returns the direction of a: +1 if a raises (a > 0), -1 if a lowers
// (a < 0), and 0 if a is the natural.
func (a Alteration) Sign() int {
	return cmp.Compare(a.num, 0)
}

// IsNatural reports whether a is the natural, the value 0. It equals
// a.Equal(Natural()) and a.Sign() == 0.
func (a Alteration) IsNatural() bool {
	return a.num == 0
}

// IsInteger reports whether a is a whole number of chromatic units, that is,
// whether the denominator of a in lowest terms is 1. The natural is whole.
func (a Alteration) IsInteger() bool {
	return a.denMinus1 == 0
}

// Fraction returns the exact value of a as a fraction num/den in lowest
// terms with den >= 1. The natural is (0, 1), and a whole value n is (n, 1).
//
// The results satisfy 1 <= den <= [MaxDenominator] and
// |num| <= MaxMagnitude*den. FromFraction(a.Fraction()) equals a.
func (a Alteration) Fraction() (num, den int) {
	return int(a.num), int(a.den())
}

// String returns a fixed form of a for output and debugging: "0" for the
// natural; a sign and a number for whole values, such as "+1", "-2", or
// "+16777216"; and a sign, the absolute value of the numerator, a slash, and
// the denominator for other values, such as "+1/2" or "-3/2". The signs are
// the ASCII characters + and -.
//
// The form is not a notation: it uses no accidental symbols, is not
// localized, and has no parser.
func (a Alteration) String() string {
	if a.num == 0 {
		return "0"
	}

	var buf [24]byte

	b := buf[:0]
	if a.num > 0 {
		b = append(b, '+')
	}

	b = strconv.AppendInt(b, a.num, 10)
	if !a.IsInteger() {
		b = append(b, '/')
		b = strconv.AppendInt(b, a.den(), 10)
	}

	return string(b)
}

// den returns the denominator of a in lowest terms.
func (a Alteration) den() int64 {
	return a.denMinus1 + 1
}

// shift returns a + n or a - n, where n is the absolute value absN and neg
// selects subtraction. ok is false if the result is out of range.
func (a Alteration) shift(neg bool, absN uint64) (Alteration, bool) {
	// Since |a| <= MaxMagnitude, a shift by more than 2*MaxMagnitude is
	// always out of range. Checking first keeps the arithmetic below within
	// int64: |num| <= 2^40 and |n*den| <= 2^41.
	if absN > 2*MaxMagnitude {
		return Alteration{}, false
	}

	n := int64(absN)
	if neg {
		n = -n
	}

	return fromInt64(a.num+n*a.den(), a.den())
}

// mul returns a * k, where k has the absolute value absK and neg is true for
// a negative k. ok is false if the result is out of range.
func (a Alteration) mul(neg bool, absK uint64) (Alteration, bool) {
	if a.num == 0 || absK == 0 {
		return Alteration{}, true
	}

	// Reduce k with the denominator first: the result p*k'/q' is then in
	// lowest terms, and q' <= q.
	q := uint64(a.den())
	g := gcd(absK, q)
	absK /= g
	q /= g

	// |p*k'|/q' <= MaxMagnitude, checked by division to avoid overflow.
	absP := absInt64(a.num)
	if absK > uint64(MaxMagnitude)*q/absP {
		return Alteration{}, false
	}

	num := int64(absP * absK)
	if (a.num < 0) != neg {
		num = -num
	}

	return Alteration{num: num, denMinus1: int64(q) - 1}, true
}

// fromInt64 returns num/den in lowest terms, where den > 0. ok is false if
// the value is out of range.
func fromInt64(num, den int64) (Alteration, bool) {
	return normalize(num < 0, absInt64(num), uint64(den))
}

// normalize returns the value absNum/absDen with the sign given by neg,
// reduced to lowest terms, where absDen > 0. ok is false if the reduced value
// is out of range.
func normalize(neg bool, absNum, absDen uint64) (Alteration, bool) {
	g := gcd(absNum, absDen)
	absNum /= g
	absDen /= g

	// absDen <= MaxDenominator is checked first, so MaxMagnitude*absDen
	// does not overflow.
	if absDen > uint64(maxDenominator) || absNum > uint64(maxMagnitude)*absDen {
		return Alteration{}, false
	}

	num := int64(absNum)
	if neg {
		num = -num
	}

	return Alteration{num: num, denMinus1: int64(absDen) - 1}, true
}

// gcd returns the greatest common divisor of a and b; gcd(0, b) is b.
func gcd(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}

	return a
}

// absInt returns |n| as uint64; it is exact for math.MinInt.
func absInt(n int) uint64 {
	return absInt64(int64(n))
}

// absInt64 returns |n| as uint64; it is exact for math.MinInt64.
func absInt64(n int64) uint64 {
	if n < 0 {
		return -uint64(n)
	}

	return uint64(n)
}

// must returns a if err is nil and panics with err otherwise.
func must(a Alteration, err error) Alteration {
	if err != nil {
		panic(err)
	}

	return a
}
