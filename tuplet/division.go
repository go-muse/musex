package tuplet

import (
	"fmt"
	"math/bits"
)

// Division is the kind of normal division that a normal count belongs to.
// The zero value and every value other than the declared constants are
// invalid. Division values are compared with ==.
type Division uint8

// Supported kinds of normal division.
const (
	// Binary is the division of a plain value into 2^k equal parts, k >= 1:
	// 2, 4, 8, 16, ….
	Binary Division = iota + 1

	// Ternary is the division of a dotted value into 3·2^k equal parts,
	// k >= 0: 3, 6, 12, 24, ….
	Ternary
)

// DivisionOf reports the division whose series contains count: (Binary, true)
// for 2, 4, 8, …, (Ternary, true) for 3, 6, 12, …, and (0, false) otherwise,
// including 1, zero, and negative counts. The two series do not intersect.
func DivisionOf(count int) (Division, bool) {
	switch {
	case isBinaryCount(count):
		return Binary, true
	case isTernaryCount(count):
		return Ternary, true
	default:
		return 0, false
	}
}

// IsValid reports whether d is one of the declared constants Binary and
// Ternary.
func (d Division) IsValid() bool {
	return d == Binary || d == Ternary
}

// String returns "binary" or "ternary" for a valid division and
// "Division(invalid)" otherwise.
func (d Division) String() string {
	switch d {
	case Binary:
		return "binary"
	case Ternary:
		return "ternary"
	default:
		return "Division(invalid)"
	}
}

// isPowerOfTwo reports whether n is a positive power of two (including 1).
func isPowerOfTwo(n int) bool {
	return n > 0 && n&(n-1) == 0
}

// isBinaryCount reports whether n is 2^k for some k >= 1.
func isBinaryCount(n int) bool {
	return n >= 2 && isPowerOfTwo(n)
}

// isTernaryCount reports whether n is 3·2^k for some k >= 0.
func isTernaryCount(n int) bool {
	return n >= 3 && n%3 == 0 && isPowerOfTwo(n/3)
}

// defaultNormal returns the default normal count for actual count n and
// division d: the largest member of the series of d that is less than n, with
// the exception that the ternary default for n = 2 is 3. It reports false when
// n < 2, when n is itself a member of the series of d, or when d is invalid.
func defaultNormal(n int, d Division) (int, bool) {
	switch d {
	case Binary:
		if n < 3 || isBinaryCount(n) {
			return 0, false
		}
		return 1 << (bits.Len(uint(n)) - 1), true
	case Ternary:
		if n == 2 {
			return 3, true
		}
		if n < 4 || isTernaryCount(n) {
			return 0, false
		}
		return 3 << (bits.Len(uint((n-1)/3)) - 1), true
	default:
		return 0, false
	}
}

// Default returns the tuplet actual:m, where m is the default normal count for
// actual and division: the largest member of the division's series that is
// less than actual (2, 4, 8, … for Binary; 3, 6, 12, … for Ternary), except
// that the Ternary default for 2 is 3. It returns an error wrapping
// ErrInvalidCount if actual < 1 and ErrNoDefault if division is unknown, if
// actual is 1, or if actual itself belongs to the division's series.
func Default(actual int, division Division) (Tuplet, error) {
	if actual < 1 {
		return Tuplet{}, fmt.Errorf("%w: actual count %d", ErrInvalidCount, actual)
	}
	if !division.IsValid() {
		return Tuplet{}, fmt.Errorf("%w: unknown division %d", ErrNoDefault, uint8(division))
	}
	normal, ok := defaultNormal(actual, division)
	if !ok {
		return Tuplet{}, fmt.Errorf("%w: actual count %d, %s division", ErrNoDefault, actual, division)
	}
	return Tuplet{actual: actual, normal: normal}, nil
}

// MustDefault is like Default but panics with the returned error.
func MustDefault(actual int, division Division) Tuplet {
	return must(Default(actual, division))
}

// NormalDivision reports the division whose series contains the normal count
// of t, as DivisionOf does: (division, true, nil) or (0, false, nil). It
// returns an error wrapping ErrUnset if t is unset.
func (t Tuplet) NormalDivision() (division Division, ok bool, err error) {
	if !t.IsSet() {
		return 0, false, unsetError("receiver")
	}
	division, ok = DivisionOf(t.normal)
	return division, ok, nil
}

// MustNormalDivision is like NormalDivision but panics with the returned
// error.
func (t Tuplet) MustNormalDivision() (division Division, ok bool) {
	division, ok, err := t.NormalDivision()
	if err != nil {
		panic(err)
	}
	return division, ok
}

// HasDefaultNormalFor reports whether a default normal count exists for the
// actual count of t and division, and equals the normal count of t. A missing
// default yields false rather than an error. It returns an error wrapping
// ErrUnset if t is unset and ErrNoDefault if division is unknown.
func (t Tuplet) HasDefaultNormalFor(division Division) (bool, error) {
	if !t.IsSet() {
		return false, unsetError("receiver")
	}
	if !division.IsValid() {
		return false, fmt.Errorf("%w: unknown division %d", ErrNoDefault, uint8(division))
	}
	normal, ok := defaultNormal(t.actual, division)
	return ok && normal == t.normal, nil
}

// MustHasDefaultNormalFor is like HasDefaultNormalFor but panics with the
// returned error.
func (t Tuplet) MustHasDefaultNormalFor(division Division) bool {
	return must(t.HasDefaultNormalFor(division))
}

// HasDefaultNormal reports whether t is the default tuplet for its actual
// count in Binary or in Ternary division. It returns an error wrapping
// ErrUnset if t is unset.
func (t Tuplet) HasDefaultNormal() (bool, error) {
	if !t.IsSet() {
		return false, unsetError("receiver")
	}
	for _, d := range [...]Division{Binary, Ternary} {
		if normal, ok := defaultNormal(t.actual, d); ok && normal == t.normal {
			return true, nil
		}
	}
	return false, nil
}

// MustHasDefaultNormal is like HasDefaultNormal but panics with the returned
// error.
func (t Tuplet) MustHasDefaultNormal() bool {
	return must(t.HasDefaultNormal())
}
