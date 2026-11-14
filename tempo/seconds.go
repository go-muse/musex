package tempo

import (
	"math/big"
	"time"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
)

// Seconds is an exact span of time in seconds, at least zero, such as 1/3 of
// a second. Unlike time.Duration it is not rounded to nanoseconds; rounding
// happens only in TimeDuration with a mode chosen by the caller.
//
// The zero value Seconds{} is the unset time; it is not zero, which is a valid
// time returned by [ZeroSeconds]. A set time is stored in reduced form.
// Seconds is comparable with ==, which matches Equal, and can be used as a
// map key.
type Seconds struct {
	v fraction.Fraction // reduced, ≥ 0; Fraction{} for the unset time
}

// ZeroSeconds returns the time zero, a valid value.
func ZeroSeconds() Seconds {
	return Seconds{v: fraction.Zero()}
}

// NewSeconds returns the time num/den seconds in reduced form; the signs are
// taken by value, so NewSeconds(-1, -2) is 1/2. In order of checking, it
// returns an error wrapping ErrInvalidArgument and fraction.ErrZeroDenominator
// if den is zero, and ErrNegativeTime if the value is negative.
func NewSeconds(num, den int) (Seconds, error) {
	const op = "NewSeconds"
	if den == 0 {
		return Seconds{}, zeroDenominatorError(op)
	}
	r := new(big.Rat).SetFrac(big.NewInt(int64(num)), big.NewInt(int64(den)))
	if r.Sign() < 0 {
		return Seconds{}, simpleError(ErrNegativeTime, op, "negative time")
	}
	return toSeconds(op, r)
}

// MustNewSeconds is like NewSeconds but panics with the returned error.
func MustNewSeconds(num, den int) Seconds {
	return must(NewSeconds(num, den))
}

// SecondsFromFraction returns the time f seconds in reduced form. In order of
// checking, it returns an error wrapping ErrInvalidArgument if f is unset and
// ErrNegativeTime if f is negative.
func SecondsFromFraction(f fraction.Fraction) (Seconds, error) {
	return secondsFrom("SecondsFromFraction", f, 1)
}

// MustSecondsFromFraction is like SecondsFromFraction but panics with the
// returned error.
func MustSecondsFromFraction(f fraction.Fraction) Seconds {
	return must(SecondsFromFraction(f))
}

// SecondsFromMinutes returns the time of f minutes, 60·f seconds. In order of
// checking, it returns an error wrapping ErrInvalidArgument if f is unset,
// ErrNegativeTime if f is negative, and ErrOutOfRange together with
// fraction.ErrOutOfRange if the result is outside the range of package
// common/fraction.
func SecondsFromMinutes(f fraction.Fraction) (Seconds, error) {
	return secondsFrom("SecondsFromMinutes", f, 60)
}

// MustSecondsFromMinutes is like SecondsFromMinutes but panics with the
// returned error.
func MustSecondsFromMinutes(f fraction.Fraction) Seconds {
	return must(SecondsFromMinutes(f))
}

// secondsFrom returns the time f·factor seconds.
func secondsFrom(op string, f fraction.Fraction, factor int64) (Seconds, error) {
	if unsetFraction(f) {
		return Seconds{}, invalidError(op, "unset fraction")
	}
	if f.Sign() < 0 {
		return Seconds{}, simpleError(ErrNegativeTime, op, "negative time")
	}
	return toSeconds(op, mul(ratOf(f), ratInt(factor)))
}

// SecondsFromTimeDuration returns the time of d exactly, d/10⁹ seconds. It
// returns an error wrapping ErrNegativeTime if d is negative.
func SecondsFromTimeDuration(d time.Duration) (Seconds, error) {
	const op = "SecondsFromTimeDuration"
	if d < 0 {
		return Seconds{}, simpleError(ErrNegativeTime, op, "negative time")
	}
	return toSeconds(op, new(big.Rat).SetFrac64(int64(d), int64(time.Second)))
}

// MustSecondsFromTimeDuration is like SecondsFromTimeDuration but panics with
// the returned error.
func MustSecondsFromTimeDuration(d time.Duration) Seconds {
	return must(SecondsFromTimeDuration(d))
}

// IsSet reports whether s is set, that is, not the unset Seconds{}.
func (s Seconds) IsSet() bool {
	return s.v.IsValid()
}

// unset returns an error wrapping ErrUnset for the receiver of op.
func (Seconds) unset(op string) error {
	return unsetError(op, "receiver")
}

// Fraction returns the number of seconds as a reduced fraction, 0/1 for zero.
// It returns an error wrapping ErrUnset if s is unset.
func (s Seconds) Fraction() (fraction.Fraction, error) {
	if !s.IsSet() {
		return fraction.Fraction{}, s.unset("Seconds.Fraction")
	}
	return s.v, nil
}

// MustFraction is like Fraction but panics with the returned error.
func (s Seconds) MustFraction() fraction.Fraction {
	return must(s.Fraction())
}

// Minutes returns the number of minutes, s/60. In order of checking, it
// returns an error wrapping ErrUnset if s is unset and ErrOutOfRange together
// with fraction.ErrOutOfRange if the result is outside the range of package
// common/fraction.
func (s Seconds) Minutes() (fraction.Fraction, error) {
	const op = "Seconds.Minutes"
	if !s.IsSet() {
		return fraction.Fraction{}, s.unset(op)
	}
	return toFraction(op, quo(ratOf(s.v), ratInt(60)))
}

// MustMinutes is like Minutes but panics with the returned error.
func (s Seconds) MustMinutes() fraction.Fraction {
	return must(s.Minutes())
}

// Equal reports whether s and other are equal times; two unset times are
// equal. Equal matches ==.
func (s Seconds) Equal(other Seconds) bool {
	return s == other
}

// Compare compares s and other by value and returns −1, 0, or +1. The unset
// time precedes every set time, and two unset times compare as 0.
func (s Seconds) Compare(other Seconds) int {
	switch {
	case !s.IsSet() || !other.IsSet():
		return compareSet(s.IsSet(), other.IsSet())
	default:
		return s.v.Compare(other.v)
	}
}

// compareSet orders an unset value before a set one; it is called when at
// least one of the two values is unset.
func compareSet(a, b bool) int {
	switch {
	case a == b:
		return 0
	case a:
		return +1
	default:
		return -1
	}
}

// Add returns s + other. In order of checking, it returns an error wrapping
// ErrUnset if s or then other is unset, and ErrOutOfRange together with
// fraction.ErrOutOfRange if the result is outside the range of package
// common/fraction.
func (s Seconds) Add(other Seconds) (Seconds, error) {
	const op = "Seconds.Add"
	if err := s.checkPair(op, other); err != nil {
		return Seconds{}, err
	}
	return toSeconds(op, new(big.Rat).Add(ratOf(s.v), ratOf(other.v)))
}

// MustAdd is like Add but panics with the returned error.
func (s Seconds) MustAdd(other Seconds) Seconds {
	return must(s.Add(other))
}

// Sub returns s − other. In order of checking, it returns an error wrapping
// ErrUnset if s or then other is unset, ErrNegativeTime if the result is
// negative, and ErrOutOfRange together with fraction.ErrOutOfRange if the
// result is outside the range of package common/fraction.
func (s Seconds) Sub(other Seconds) (Seconds, error) {
	const op = "Seconds.Sub"
	if err := s.checkPair(op, other); err != nil {
		return Seconds{}, err
	}
	r := new(big.Rat).Sub(ratOf(s.v), ratOf(other.v))
	if r.Sign() < 0 {
		return Seconds{}, simpleError(ErrNegativeTime, op, "negative difference")
	}
	return toSeconds(op, r)
}

// MustSub is like Sub but panics with the returned error.
func (s Seconds) MustSub(other Seconds) Seconds {
	return must(s.Sub(other))
}

// checkPair returns the error of an unset receiver or argument.
func (s Seconds) checkPair(op string, other Seconds) error {
	if !s.IsSet() {
		return s.unset(op)
	}
	if !other.IsSet() {
		return unsetError(op, "argument")
	}
	return nil
}

// nanosecond is 10⁹, the number of nanoseconds in a second.
var nanosecond = big.NewRat(int64(time.Second), 1)

// TimeDuration returns s as a time.Duration: s·10⁹ rounded to an integer
// number of nanoseconds by mode. There is no default mode. Since a time is
// not negative, TowardZero gives the same result as Down, AwayFromZero as
// Up, HalfTowardZero as HalfDown, and HalfAwayFromZero as HalfUp; a whole
// number of nanoseconds is unchanged in every mode. The largest result is
// math.MaxInt64 nanoseconds, about 292 years. In order of checking, it
// returns an error wrapping ErrUnset if s is unset, ErrInvalidArgument
// together with rounding.ErrUnset or rounding.ErrUnknownMode if mode is unset
// or unknown, and ErrTimeOverflow if the rounded number of nanoseconds is
// greater than math.MaxInt64.
func (s Seconds) TimeDuration(mode rounding.Mode) (time.Duration, error) {
	const op = "Seconds.TimeDuration"
	if !s.IsSet() {
		return 0, s.unset(op)
	}
	if err := checkMode(op, mode); err != nil {
		return 0, err
	}
	k, err := mode.RatInt(mul(ratOf(s.v), nanosecond))
	if err != nil {
		return 0, modeError(op, mode)
	}
	if !k.IsInt64() {
		return 0, simpleError(ErrTimeOverflow, op, "more than math.MaxInt64 nanoseconds")
	}
	return time.Duration(k.Int64()), nil
}

// MustTimeDuration is like TimeDuration but panics with the returned error.
func (s Seconds) MustTimeDuration(mode rounding.Mode) time.Duration {
	return must(s.TimeDuration(mode))
}
