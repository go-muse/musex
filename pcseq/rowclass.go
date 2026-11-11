package pcseq

// RowClass is an immutable row orbit identified by its relation and canonical
// lexicographically minimal representative.
type RowClass struct {
	kind           RowClassKind
	representative Row
}

// NewRowClass returns the orbit of source under kind.
func NewRowClass(source Row, kind RowClassKind) (RowClass, error) { return RowClass{}, nil }

// AreEquivalentRows reports whether a and b belong to one orbit of kind.
func AreEquivalentRows(a, b Row, kind RowClassKind) (bool, error) { return false, nil }

// IsValid reports whether c is a valid row class.
func (c RowClass) IsValid() bool { return false }

// Modulus returns the class modulus, or 0 for an invalid value.
func (c RowClass) Modulus() int { return 0 }

// Kind returns the row equivalence relation.
func (c RowClass) Kind() RowClassKind { return 0 }

// Representative returns the canonical minimal row.
func (c RowClass) Representative() Row { return Row{} }

// Size returns the number of distinct forms.
func (c RowClass) Size() int { return 0 }

// Contains reports whether row belongs to c.
func (c RowClass) Contains(row Row) bool { return false }

// Members returns all unique forms in canonical row order.
func (c RowClass) Members() ([]Row, error) { return nil, nil }

// Equal reports domain equality of two valid classes.
func (c RowClass) Equal(other RowClass) bool { return false }

// Compare orders by modulus, kind, and representative.
func (c RowClass) Compare(other RowClass) int { return 0 }

// String returns Kind(representative), or RowClass(invalid).
func (c RowClass) String() string { return "" }
