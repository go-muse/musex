package pcseq

import (
	"cmp"
	"slices"

	"github.com/go-muse/musex/pitchclass"
)

// RowClass is an immutable row orbit identified by its relation and canonical
// lexicographically minimal representative.
type RowClass struct {
	kind           RowClassKind
	representative Row
}

func rowClassLimit(kind RowClassKind) int {
	switch kind {
	case TIRRotations:
		return MaxRotationClassModulus
	case TTOR:
		return MaxTTOClassModulus
	default:
		return MaxModulus
	}
}

func enumerateTIR(row Row, yield func(Row) bool) {
	for _, kind := range []FormKind{P, I, R, RI} {
		for n := range row.modulus {
			if !yield(formUnchecked(row, FormLabel{kind: kind, index: n, modulus: row.modulus})) {
				return
			}
		}
	}
}

func minimalRotation(row Row) Row {
	n := row.Len()
	if n <= 1 {
		return row
	}
	i, j, k := 0, 1, 0
	for i < n && j < n && k < n {
		a := numberAt(row.elements, (i+k)%n)
		b := numberAt(row.elements, (j+k)%n)
		if a == b {
			k++
			continue
		}
		if a > b {
			i += k + 1
			if i == j {
				i++
			}
		} else {
			j += k + 1
			if i == j {
				j++
			}
		}
		k = 0
	}
	return row.MustRotate(min(i, j))
}

func canonicalRow(source Row, kind RowClassKind) Row {
	representative := source
	switch kind {
	case TIR:
		enumerateTIR(source, func(candidate Row) bool {
			if candidate.Compare(representative) < 0 {
				representative = candidate
			}
			return true
		})
	case TIRRotations:
		enumerateTIR(source, func(candidate Row) bool {
			candidate = minimalRotation(candidate)
			if candidate.Compare(representative) < 0 {
				representative = candidate
			}
			return true
		})
	case TTOR:
		for operator := range pitchclass.MustOperators(source.modulus) {
			candidate := source.MustApply(operator)
			if candidate.Compare(representative) < 0 {
				representative = candidate
			}
			candidate = candidate.MustRetrograde()
			if candidate.Compare(representative) < 0 {
				representative = candidate
			}
		}
	}
	return representative
}

// NewRowClass returns the orbit of source under kind.
func NewRowClass(source Row, kind RowClassKind) (RowClass, error) {
	if !source.IsValid() {
		return RowClass{}, invalidValue()
	}
	if !kind.IsValid() {
		return RowClass{}, invalidValue()
	}
	if err := checkModulusLimit(source.modulus, rowClassLimit(kind)); err != nil {
		return RowClass{}, err
	}
	return RowClass{kind: kind, representative: canonicalRow(source, kind)}, nil
}

// AreEquivalentRows reports whether a and b belong to one orbit of kind.
func AreEquivalentRows(a, b Row, kind RowClassKind) (bool, error) {
	if !a.IsValid() || !b.IsValid() {
		return false, invalidValue()
	}
	if a.modulus != b.modulus {
		return false, modulusMismatch(a.modulus, b.modulus)
	}
	class, err := NewRowClass(a, kind)
	if err != nil {
		return false, err
	}
	return class.Contains(b), nil
}

// IsValid reports whether c is a valid row class.
func (c RowClass) IsValid() bool {
	return c.kind.IsValid() && c.representative.IsValid() &&
		c.representative.modulus <= rowClassLimit(c.kind)
}

// Modulus returns the class modulus, or 0 for an invalid value.
func (c RowClass) Modulus() int { return c.representative.modulus }

// Kind returns the row equivalence relation.
func (c RowClass) Kind() RowClassKind { return c.kind }

// Representative returns the canonical minimal row.
func (c RowClass) Representative() Row { return c.representative }

func (c RowClass) distinctSize() int {
	unique := make(map[Row]struct{})
	switch c.kind {
	case TIRRotations:
		enumerateTIR(c.representative, func(candidate Row) bool {
			unique[minimalRotation(candidate)] = struct{}{}
			return true
		})
	case TTOR:
		for operator := range pitchclass.MustOperators(c.representative.modulus) {
			candidate := c.representative.MustApply(operator)
			unique[candidate] = struct{}{}
			unique[candidate.MustRetrograde()] = struct{}{}
		}
	}
	return len(unique)
}

// Size returns the number of distinct forms.
func (c RowClass) Size() int {
	if !c.IsValid() {
		return 0
	}
	n := c.representative.modulus
	if n <= 2 {
		return n
	}
	if c.kind != TIR {
		return c.distinctSize()
	}
	retrograde := c.representative.MustRetrograde()
	labels := c.representative.MustLabelsOf(retrograde)
	for _, label := range labels {
		if label.kind == P || label.kind == I {
			return 2 * n
		}
	}
	return 4 * n
}

// Contains reports whether row belongs to c.
func (c RowClass) Contains(row Row) bool {
	if !c.IsValid() || !row.IsValid() || c.representative.modulus != row.modulus {
		return false
	}
	switch c.kind {
	case TIR:
		labels := c.representative.MustLabelsOf(row)
		return len(labels) > 0
	case TIRRotations:
		minimal := minimalRotation(row)
		found := false
		enumerateTIR(c.representative, func(candidate Row) bool {
			if minimalRotation(candidate) == minimal {
				found = true
				return false
			}
			return true
		})
		return found
	case TTOR:
		for operator := range pitchclass.MustOperators(c.representative.modulus) {
			candidate := c.representative.MustApply(operator)
			if candidate == row || candidate.MustRetrograde() == row {
				return true
			}
		}
	}
	return false
}

// Members returns all unique forms in canonical row order.
func (c RowClass) Members() ([]Row, error) {
	if !c.IsValid() {
		return nil, invalidValue()
	}
	if err := checkModulusLimit(c.representative.modulus, MaxMaterializedClassModulus); err != nil {
		return nil, err
	}
	unique := make(map[Row]struct{})
	switch c.kind {
	case TIR:
		enumerateTIR(c.representative, func(candidate Row) bool {
			unique[candidate] = struct{}{}
			return true
		})
	case TIRRotations:
		enumerateTIR(c.representative, func(candidate Row) bool {
			if c.representative.modulus > 2 {
				candidate = minimalRotation(candidate)
			}
			unique[candidate] = struct{}{}
			return true
		})
	case TTOR:
		for operator := range pitchclass.MustOperators(c.representative.modulus) {
			candidate := c.representative.MustApply(operator)
			unique[candidate] = struct{}{}
			unique[candidate.MustRetrograde()] = struct{}{}
		}
	}
	members := make([]Row, 0, len(unique))
	for member := range unique {
		members = append(members, member)
	}
	slices.SortFunc(members, Row.Compare)
	return members, nil
}

// Equal reports domain equality of two valid classes.
func (c RowClass) Equal(other RowClass) bool {
	return c.IsValid() && other.IsValid() && c == other
}

// Compare orders by modulus, kind, and representative.
func (c RowClass) Compare(other RowClass) int {
	if !c.IsValid() || !other.IsValid() {
		switch {
		case !c.IsValid() && !other.IsValid():
			return 0
		case !c.IsValid():
			return -1
		default:
			return 1
		}
	}
	if result := cmp.Compare(c.representative.modulus, other.representative.modulus); result != 0 {
		return result
	}
	if result := c.kind.Compare(other.kind); result != 0 {
		return result
	}
	return c.representative.Compare(other.representative)
}

// String returns Kind(representative), or RowClass(invalid).
func (c RowClass) String() string {
	if !c.IsValid() {
		return "RowClass(invalid)"
	}
	return c.kind.String() + "(" + c.representative.String() + ")"
}
