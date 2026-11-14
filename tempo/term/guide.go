package term

import (
	"github.com/go-muse/musex/common/fraction"
)

// Guide is the reference range of beats per minute that a list gives for a
// term; one of its bounds may be open. Both bounds belong to the range. The
// numbers are reference data of the list, not a tempo of the term, and carry
// no duration: they are compared with a number of beats as it is.
//
// A guide is created by [Between], [AtLeast], or [AtMost], which store their
// bounds in reduced form. The zero value Guide{} has no bounds and is not a
// valid guide: IsValid reports false, Contains fails, and a row of a list
// cannot carry it. Guide is comparable with ==, which matches Equal, and can
// be used as a map key.
type Guide struct {
	lower fraction.Fraction // reduced; Fraction{} if the lower bound is open
	upper fraction.Fraction // reduced; Fraction{} if the upper bound is open
}

// positiveBound reports whether f is a set fraction greater than zero.
func positiveBound(f fraction.Fraction) bool {
	return f.IsValid() && f.Sign() > 0
}

// Between returns the guide from lower to upper, both included. Equal bounds
// are allowed and name a single number. It returns an error wrapping
// ErrInvalidArgument if a bound is unset or not positive, or if lower is
// greater than upper.
func Between(lower, upper fraction.Fraction) (Guide, error) {
	const op = "Between"
	if !positiveBound(lower) || !positiveBound(upper) {
		return Guide{}, invalidError(op, "bound unset or not positive")
	}
	if lower.Compare(upper) > 0 {
		return Guide{}, invalidError(op, "lower bound greater than upper bound")
	}
	return Guide{lower: lower.Reduce(), upper: upper.Reduce()}, nil
}

// MustBetween is like Between but panics with the returned error.
func MustBetween(lower, upper fraction.Fraction) Guide {
	return must(Between(lower, upper))
}

// AtLeast returns the guide "from lower", with an open upper bound; lower is
// included. It returns an error wrapping ErrInvalidArgument if lower is unset
// or not positive.
func AtLeast(lower fraction.Fraction) (Guide, error) {
	if !positiveBound(lower) {
		return Guide{}, invalidError("AtLeast", "bound unset or not positive")
	}
	return Guide{lower: lower.Reduce()}, nil
}

// MustAtLeast is like AtLeast but panics with the returned error.
func MustAtLeast(lower fraction.Fraction) Guide {
	return must(AtLeast(lower))
}

// AtMost returns the guide "up to upper", with an open lower bound; upper is
// included. It returns an error wrapping ErrInvalidArgument if upper is unset
// or not positive.
func AtMost(upper fraction.Fraction) (Guide, error) {
	if !positiveBound(upper) {
		return Guide{}, invalidError("AtMost", "bound unset or not positive")
	}
	return Guide{upper: upper.Reduce()}, nil
}

// MustAtMost is like AtMost but panics with the returned error.
func MustAtMost(upper fraction.Fraction) Guide {
	return must(AtMost(upper))
}

// IsValid reports whether g has at least one bound. The zero value Guide{} has
// none.
func (g Guide) IsValid() bool {
	return g.lower.IsValid() || g.upper.IsValid()
}

// Lower returns the lower bound and true, or Fraction{} and false if the lower
// bound is open or g has no bounds.
func (g Guide) Lower() (fraction.Fraction, bool) {
	return g.lower, g.lower.IsValid()
}

// Upper returns the upper bound and true, or Fraction{} and false if the upper
// bound is open or g has no bounds.
func (g Guide) Upper() (fraction.Fraction, bool) {
	return g.upper, g.upper.IsValid()
}

// Contains reports whether lower ≤ beats ≤ upper; an open bound does not
// limit the range. In order of checking, it returns an error wrapping
// ErrInvalidArgument if g has no bounds or beats is unset, and ErrNoTempo if
// beats is zero or negative.
func (g Guide) Contains(beats fraction.Fraction) (bool, error) {
	const op = "Guide.Contains"
	if !g.IsValid() {
		return false, invalidError(op, "guide without bounds")
	}
	if !beats.IsValid() {
		return false, invalidError(op, "unset number of beats")
	}
	if beats.Sign() <= 0 {
		return false, noTempoError(op)
	}
	return g.contains(beats), nil
}

// MustContains is like Contains but panics with the returned error.
func (g Guide) MustContains(beats fraction.Fraction) bool {
	return must(g.Contains(beats))
}

// contains reports whether the valid guide g holds the positive number beats.
func (g Guide) contains(beats fraction.Fraction) bool {
	if g.lower.IsValid() && beats.Compare(g.lower) < 0 {
		return false
	}
	return !g.upper.IsValid() || beats.Compare(g.upper) <= 0
}

// Equal reports whether g and other have equal bounds and the same open
// bounds; it matches ==.
func (g Guide) Equal(other Guide) bool {
	return g == other
}

// String returns the guide with its numbers written as integers or reduced
// fractions "n/d": "40-60" for both bounds, "100-100" for equal bounds,
// "<=24" for an open lower bound, ">=200" for an open upper bound, and
// "Guide(none)" for a guide without bounds.
func (g Guide) String() string {
	switch {
	case g.lower.IsValid() && g.upper.IsValid():
		return string(appendNumber(appendNumber(nil, g.lower), g.upper, '-'))
	case g.lower.IsValid():
		return ">=" + string(appendNumber(nil, g.lower))
	case g.upper.IsValid():
		return "<=" + string(appendNumber(nil, g.upper))
	default:
		return "Guide(none)"
	}
}

// appendNumber appends the reduced value of the valid fraction f, preceded by
// sep if one is given: an integer as its digits and any other value as "n/d".
func appendNumber(b []byte, f fraction.Fraction, sep ...byte) []byte {
	b = append(b, sep...)
	num, den, _ := f.ReducedBigPair()
	b = num.Append(b, 10)
	if den.IsInt64() && den.Int64() == 1 {
		return b
	}
	b = append(b, '/')
	return den.Append(b, 10)
}
