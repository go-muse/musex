package tuplet

import (
	"cmp"
	"math/bits"
)

// Equal reports whether t and other have the same notation: equal actual and
// equal normal counts. 3:2 and 6:4 are not equal. The unset tuplet equals
// only the unset tuplet. Equal matches the built-in == operator.
//
// Unlike fraction.Fraction.Equal, which compares numeric values, Equal
// compares the stored pair; use [Tuplet.EqualCoefficient] to compare
// coefficient values.
func (t Tuplet) Equal(other Tuplet) bool {
	return t == other
}

// Compare defines a total order for sorting: by coefficient value
// normal/actual in ascending order, then by actual count, then by normal
// count. The unset tuplet precedes every set tuplet, and two unset tuplets
// compare equal. Compare returns 0 if and only if t == other, so it can be
// used with slices.SortFunc as tuplet.Tuplet.Compare.
func (t Tuplet) Compare(other Tuplet) int {
	switch {
	case !t.IsSet() && !other.IsSet():
		return 0
	case !t.IsSet():
		return -1
	case !other.IsSet():
		return 1
	}
	if c := cmpRatio(t.normal, t.actual, other.normal, other.actual); c != 0 {
		return c
	}
	if c := cmp.Compare(t.actual, other.actual); c != 0 {
		return c
	}
	return cmp.Compare(t.normal, other.normal)
}

// EqualCoefficient reports whether the coefficients of t and other are equal
// by value: 3:2 and 6:4 have equal coefficients. It returns an error wrapping
// ErrUnset if t or other is unset, checking t first.
func (t Tuplet) EqualCoefficient(other Tuplet) (bool, error) {
	c, err := t.CompareCoefficient(other)
	return c == 0 && err == nil, err
}

// MustEqualCoefficient is like EqualCoefficient but panics with the returned
// error.
func (t Tuplet) MustEqualCoefficient(other Tuplet) bool {
	return must(t.EqualCoefficient(other))
}

// CompareCoefficient compares the coefficient values of t and other and
// returns -1, 0, or +1 as the sign of normal1/actual1 - normal2/actual2. It
// returns an error wrapping ErrUnset if t or other is unset, checking t first.
func (t Tuplet) CompareCoefficient(other Tuplet) (int, error) {
	if !t.IsSet() {
		return 0, unsetError("receiver")
	}
	if !other.IsSet() {
		return 0, unsetError("argument")
	}
	return cmpRatio(t.normal, t.actual, other.normal, other.actual), nil
}

// MustCompareCoefficient is like CompareCoefficient but panics with the
// returned error.
func (t Tuplet) MustCompareCoefficient(other Tuplet) int {
	return must(t.CompareCoefficient(other))
}

// cmpRatio compares the ratios m1/n1 and m2/n2 of positive integers exactly by
// comparing the 128-bit cross products m1·n2 and m2·n1.
func cmpRatio(m1, n1, m2, n2 int) int {
	hi1, lo1 := bits.Mul64(uint64(m1), uint64(n2))
	hi2, lo2 := bits.Mul64(uint64(m2), uint64(n1))
	if c := cmp.Compare(hi1, hi2); c != 0 {
		return c
	}
	return cmp.Compare(lo1, lo2)
}
