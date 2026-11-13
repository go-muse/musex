package duration

import (
	"math"
	"math/big"

	"github.com/go-muse/musex/common/fraction"
)

// Duration is an exact positive length of musical time in whole notes,
// without notation: 1/4 is a quarter, 3/8 a dotted quarter, and 1/12 an
// eighth of a triplet.
//
// The zero value Duration{} is the unset duration; it is not zero, which is
// not a duration. A set duration is stored in reduced form, so New(6, 16)
// and New(3, 8) are the same value. Duration is comparable with ==, which
// matches Equal, and can be used as a map key.
type Duration struct {
	v fraction.Fraction
}

// newDuration returns the reduced duration r > 0, or an error wrapping
// ErrOutOfRange if r is outside the range of package common/fraction.
func newDuration(op string, r *big.Rat) (Duration, error) {
	f, ok := fromRat(r)
	if !ok {
		return Duration{}, rangeError(op)
	}
	return Duration{v: f}, nil
}

// New returns the duration num/den in whole notes. The signs of num and den
// are taken by value: New(-1, -4) is 1/4. It returns an error wrapping
// ErrInvalidArgument and fraction.ErrZeroDenominator if den is zero, and
// ErrNotPositive if the value is zero or negative.
func New(num, den int) (Duration, error) {
	const op = "New"
	if den == 0 {
		return Duration{}, zeroDenominatorError(op)
	}
	r := new(big.Rat).SetFrac(big.NewInt(int64(num)), big.NewInt(int64(den)))
	if r.Sign() <= 0 {
		return Duration{}, simpleError(ErrNotPositive, op)
	}
	return newDuration(op, r)
}

// MustNew is like New but panics with the returned error.
func MustNew(num, den int) Duration {
	return must(New(num, den))
}

// FromFraction returns the duration with the value of f in whole notes, in
// reduced form. It returns an error wrapping ErrInvalidArgument if f is unset
// and ErrNotPositive if f is zero or negative.
func FromFraction(f fraction.Fraction) (Duration, error) {
	const op = "FromFraction"
	if !f.IsValid() {
		return Duration{}, invalidError(op, "unset fraction")
	}
	if f.Sign() <= 0 {
		return Duration{}, simpleError(ErrNotPositive, op)
	}
	return Duration{v: f.Reduce()}, nil
}

// MustFromFraction is like FromFraction but panics with the returned error.
func MustFromFraction(f fraction.Fraction) Duration {
	return must(FromFraction(f))
}

// checkList returns an error wrapping ErrInvalidArgument for an empty list
// and ErrUnset for the first unset element.
func checkList(op string, ds []Duration) error {
	if len(ds) == 0 {
		return invalidError(op, "empty list")
	}
	for _, d := range ds {
		if !d.IsSet() {
			return unsetError(op, "list element")
		}
	}
	return nil
}

// Sum returns the exact sum of a non-empty list of durations. It returns an
// error wrapping ErrInvalidArgument for an empty list, ErrUnset for an unset
// element, and ErrOutOfRange if the sum is outside the range of package
// common/fraction. The result does not depend on the order of the list, and
// the list is only read.
func Sum(ds ...Duration) (Duration, error) {
	const op = "Sum"
	if err := checkList(op, ds); err != nil {
		return Duration{}, err
	}
	sum := new(big.Rat)
	for _, d := range ds {
		sum.Add(sum, ratOf(d.v))
	}
	return newDuration(op, sum)
}

// MustSum is like Sum but panics with the returned error.
func MustSum(ds ...Duration) Duration {
	return must(Sum(ds...))
}

// Min returns the smallest duration of a non-empty list. It returns an error
// wrapping ErrInvalidArgument for an empty list and ErrUnset for an unset
// element.
func Min(ds ...Duration) (Duration, error) {
	return extreme("Min", -1, ds)
}

// MustMin is like Min but panics with the returned error.
func MustMin(ds ...Duration) Duration {
	return must(Min(ds...))
}

// Max returns the largest duration of a non-empty list. It returns an error
// wrapping ErrInvalidArgument for an empty list and ErrUnset for an unset
// element.
func Max(ds ...Duration) (Duration, error) {
	return extreme("Max", 1, ds)
}

// MustMax is like Max but panics with the returned error.
func MustMax(ds ...Duration) Duration {
	return must(Max(ds...))
}

// extreme returns the element of ds that compares as sign to all others.
func extreme(op string, sign int, ds []Duration) (Duration, error) {
	if err := checkList(op, ds); err != nil {
		return Duration{}, err
	}
	best := ds[0]
	for _, d := range ds[1:] {
		if d.Compare(best) == sign {
			best = d
		}
	}
	return best, nil
}

// IsSet reports whether d is a set duration.
func (d Duration) IsSet() bool {
	return d.v.IsValid()
}

// Fraction returns the value of d as a reduced fraction with a positive
// denominator: New(6, 16).Fraction() has the stored pair (3, 8). It returns
// an error wrapping ErrUnset if d is unset.
func (d Duration) Fraction() (fraction.Fraction, error) {
	if !d.IsSet() {
		return fraction.Fraction{}, unsetError("Duration.Fraction", "receiver")
	}
	return d.v, nil
}

// MustFraction is like Fraction but panics with the returned error.
func (d Duration) MustFraction() fraction.Fraction {
	return must(d.Fraction())
}

// Equal reports whether d and other have the same value. It matches ==; the
// unset duration is equal only to the unset duration.
func (d Duration) Equal(other Duration) bool {
	return d == other
}

// Compare compares d and other by value and returns −1, 0, or +1. The unset
// duration precedes every set duration, and two unset durations are equal.
// It suits slices.SortFunc(ds, Duration.Compare).
func (d Duration) Compare(other Duration) int {
	return d.v.Compare(other.v)
}

// binary checks that d and other are set and returns their values.
func (d Duration) binary(op string, other Duration) (x, y *big.Rat, err error) {
	if !d.IsSet() {
		return nil, nil, unsetError(op, "receiver")
	}
	if !other.IsSet() {
		return nil, nil, unsetError(op, "argument")
	}
	return ratOf(d.v), ratOf(other.v), nil
}

// Add returns the exact sum d + other. It returns an error wrapping ErrUnset
// if d or other is unset and ErrOutOfRange if the sum is outside the range of
// package common/fraction.
func (d Duration) Add(other Duration) (Duration, error) {
	const op = "Duration.Add"
	x, y, err := d.binary(op, other)
	if err != nil {
		return Duration{}, err
	}
	return newDuration(op, x.Add(x, y))
}

// MustAdd is like Add but panics with the returned error.
func (d Duration) MustAdd(other Duration) Duration {
	return must(d.Add(other))
}

// Sub returns the exact difference d − other. It returns an error wrapping
// ErrUnset if d or other is unset, ErrNotPositive if the difference is zero
// or negative, and ErrOutOfRange if it is outside the range of package
// common/fraction.
func (d Duration) Sub(other Duration) (Duration, error) {
	const op = "Duration.Sub"
	x, y, err := d.binary(op, other)
	if err != nil {
		return Duration{}, err
	}
	x.Sub(x, y)
	if x.Sign() <= 0 {
		return Duration{}, simpleError(ErrNotPositive, op)
	}
	return newDuration(op, x)
}

// MustSub is like Sub but panics with the returned error.
func (d Duration) MustSub(other Duration) Duration {
	return must(d.Sub(other))
}

// positiveFactor checks that d is set and f is a set positive fraction.
func (d Duration) positiveFactor(op string, f fraction.Fraction) error {
	if !d.IsSet() {
		return unsetError(op, "receiver")
	}
	if !f.IsValid() {
		return invalidError(op, "unset fraction")
	}
	if f.Sign() <= 0 {
		return invalidError(op, "fraction is not positive")
	}
	return nil
}

// positiveInt checks that d is set and n ≥ 1.
func (d Duration) positiveInt(op string, n int) error {
	if !d.IsSet() {
		return unsetError(op, "receiver")
	}
	if n < 1 {
		return invalidError(op, "integer less than 1")
	}
	return nil
}

// Mul returns the exact product d·f. It returns an error wrapping ErrUnset if
// d is unset, ErrInvalidArgument if f is unset, zero, or negative, and
// ErrOutOfRange if the product is outside the range of package
// common/fraction.
func (d Duration) Mul(f fraction.Fraction) (Duration, error) {
	const op = "Duration.Mul"
	if err := d.positiveFactor(op, f); err != nil {
		return Duration{}, err
	}
	x := ratOf(d.v)
	return newDuration(op, x.Mul(x, ratOf(f)))
}

// MustMul is like Mul but panics with the returned error.
func (d Duration) MustMul(f fraction.Fraction) Duration {
	return must(d.Mul(f))
}

// MulInt returns the exact product d·n. It returns an error wrapping
// ErrUnset if d is unset, ErrInvalidArgument if n < 1, and ErrOutOfRange if
// the product is outside the range of package common/fraction.
func (d Duration) MulInt(n int) (Duration, error) {
	const op = "Duration.MulInt"
	if err := d.positiveInt(op, n); err != nil {
		return Duration{}, err
	}
	x := ratOf(d.v)
	return newDuration(op, x.Mul(x, new(big.Rat).SetInt64(int64(n))))
}

// MustMulInt is like MulInt but panics with the returned error.
func (d Duration) MustMulInt(n int) Duration {
	return must(d.MulInt(n))
}

// Div returns the exact quotient d/f. It returns an error wrapping ErrUnset if
// d is unset, ErrInvalidArgument if f is unset, zero, or negative, and
// ErrOutOfRange if the quotient is outside the range of package
// common/fraction.
func (d Duration) Div(f fraction.Fraction) (Duration, error) {
	const op = "Duration.Div"
	if err := d.positiveFactor(op, f); err != nil {
		return Duration{}, err
	}
	x := ratOf(d.v)
	return newDuration(op, x.Quo(x, ratOf(f)))
}

// MustDiv is like Div but panics with the returned error.
func (d Duration) MustDiv(f fraction.Fraction) Duration {
	return must(d.Div(f))
}

// DivInt returns the exact quotient d/n. It returns an error wrapping
// ErrUnset if d is unset, ErrInvalidArgument if n < 1, and ErrOutOfRange if
// the quotient is outside the range of package common/fraction.
func (d Duration) DivInt(n int) (Duration, error) {
	const op = "Duration.DivInt"
	if err := d.positiveInt(op, n); err != nil {
		return Duration{}, err
	}
	x := ratOf(d.v)
	return newDuration(op, x.Quo(x, new(big.Rat).SetInt64(int64(n))))
}

// MustDivInt is like DivInt but panics with the returned error.
func (d Duration) MustDivInt(n int) Duration {
	return must(d.DivInt(n))
}

// Ratio returns the exact ratio d/other as a reduced fraction, which is a
// number rather than a duration: 3/8 : 1/8 is 3. It returns an error
// wrapping ErrUnset if d or other is unset and ErrOutOfRange if the ratio is
// outside the range of package common/fraction.
func (d Duration) Ratio(other Duration) (fraction.Fraction, error) {
	const op = "Duration.Ratio"
	x, y, err := d.binary(op, other)
	if err != nil {
		return fraction.Fraction{}, err
	}
	f, ok := fromRat(x.Quo(x, y))
	if !ok {
		return fraction.Fraction{}, rangeError(op)
	}
	return f, nil
}

// MustRatio is like Ratio but panics with the returned error.
func (d Duration) MustRatio(other Duration) fraction.Fraction {
	return must(d.Ratio(other))
}

// QuoRem returns how many times other fits in d entirely, quo = ⌊d/other⌋ ≥ 0,
// and the remainder rem = d − quo·other as a reduced fraction with
// 0 ≤ rem < other: 1.QuoRem(3/8) gives 2 and 1/4. It returns an error
// wrapping ErrUnset if d or other is unset; ErrOutOfRange together with
// fraction.ErrNotRepresentable if quo is greater than math.MaxInt; and then
// ErrOutOfRange together with fraction.ErrOutOfRange if rem is outside the
// range of package common/fraction, which is possible even for durations
// within the range.
func (d Duration) QuoRem(other Duration) (quo int, rem fraction.Fraction, err error) {
	const op = "Duration.QuoRem"
	x, y, err := d.binary(op, other)
	if err != nil {
		return 0, fraction.Fraction{}, err
	}
	q := floorRat(new(big.Rat).Quo(x, y))
	if !q.IsInt64() || q.Int64() > math.MaxInt {
		return 0, fraction.Fraction{}, wrapNotRepresentable(op)
	}
	y.Mul(y, new(big.Rat).SetInt(q))
	f, ok := fromRat(x.Sub(x, y))
	if !ok {
		return 0, fraction.Fraction{}, rangeError(op)
	}
	return int(q.Int64()), f, nil
}

// MustQuoRem is like QuoRem but panics with the returned error.
func (d Duration) MustQuoRem(other Duration) (quo int, rem fraction.Fraction) {
	return must2(d.QuoRem(other))
}
