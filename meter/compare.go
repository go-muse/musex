package meter

import (
	"slices"
	"strings"
)

// Equal reports whether m and other are equal by record: the same form,
// symbol, parts with their addends and denominators in order, and the same
// explicit grouping or its absence. It is the same as m == other; the unset
// meter equals only the unset meter. C and 4/4 differ; so do 5/8 and
// 5/8[2+3]. 6/8[3+3] is the same value as 6/8.
func (m Meter) Equal(other Meter) bool {
	return m == other
}

// Compare returns −1, 0, or +1 and orders meters totally for
// slices.SortFunc: the unset meter comes first; set meters are ordered by
// size and, at an equal size, by their canonical text with strings.Compare.
// Compare returns 0 exactly for equal meters: 2/4 comes before 3/4, and 3/4
// before 6/8.
func (m Meter) Compare(other Meter) int {
	switch {
	case !m.IsSet() && !other.IsSet():
		return 0
	case !m.IsSet():
		return -1
	case !other.IsSet():
		return 1
	}
	if c := m.size.Compare(other.size); c != 0 {
		return c
	}
	return strings.Compare(m.rec, other.rec)
}

// bothSet returns the records of m and other, or an ErrUnset error for the
// receiver and then for the argument.
func (m Meter) bothSet(op string, other Meter) (record, record, error) {
	r, err := m.recordOf(op)
	if err != nil {
		return record{}, record{}, err
	}
	if !other.IsSet() {
		return record{}, record{}, unsetError(op, "argument")
	}
	return r, decode(other.rec), nil
}

// EqualStructure reports whether m and other are equal by structure: the
// same unit durations in order and the same grouping, or no grouping in both.
// C and 4/4 are equal by structure; so are (2+3)/8, 5/8[2+3], and 2/8+3/8.
// Meters equal by record are equal by structure.
//
// Errors: ErrUnset for the unset receiver, then for the unset argument.
func (m Meter) EqualStructure(other Meter) (bool, error) {
	r, o, err := m.bothSet("Meter.EqualStructure", other)
	if err != nil {
		return false, err
	}
	rs, os := r.shapeOf(), o.shapeOf()
	return slices.Equal(rs.dens, os.dens) && (rs.grouping == nil) == (os.grouping == nil) &&
		slices.Equal(rs.grouping, os.grouping), nil
}

// MustEqualStructure is like EqualStructure but panics with the error of
// EqualStructure.
func (m Meter) MustEqualStructure(other Meter) bool {
	return must(m.EqualStructure(other))
}

// EqualDuration reports whether m and other have equal sizes: 3/4 and 6/8
// are equal by size. Meters equal by structure are equal by size.
//
// Errors: ErrUnset for the unset receiver, then for the unset argument.
func (m Meter) EqualDuration(other Meter) (bool, error) {
	if _, _, err := m.bothSet("Meter.EqualDuration", other); err != nil {
		return false, err
	}
	return m.size.Equal(other.size), nil
}

// MustEqualDuration is like EqualDuration but panics with the error of
// EqualDuration.
func (m Meter) MustEqualDuration(other Meter) bool {
	return must(m.EqualDuration(other))
}

// CompareDuration returns the sign of the difference of the sizes of m and
// other: −1, 0, or +1.
//
// Errors: ErrUnset for the unset receiver, then for the unset argument.
func (m Meter) CompareDuration(other Meter) (int, error) {
	if _, _, err := m.bothSet("Meter.CompareDuration", other); err != nil {
		return 0, err
	}
	return m.size.Compare(other.size), nil
}

// MustCompareDuration is like CompareDuration but panics with the error of
// CompareDuration.
func (m Meter) MustCompareDuration(other Meter) int {
	return must(m.CompareDuration(other))
}
