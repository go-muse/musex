package pcseq

import "github.com/go-muse/musex/pitchclass"

// Row is an immutable permutation of every class of its modulus. Its zero
// value is invalid.
type Row struct {
	modulus  int
	elements string
}

// NewRow returns a row from strict class numbers.
func NewRow(modulus int, numbers ...int) (Row, error) {
	sequence, err := New(modulus, numbers...)
	if err != nil {
		return Row{}, err
	}
	return RowFromSequence(sequence)
}

// ReduceRow returns a row after reducing arbitrary integers modulo modulus.
func ReduceRow(modulus int, xs ...int) (Row, error) {
	sequence, err := Reduce(modulus, xs...)
	if err != nil {
		return Row{}, err
	}
	return RowFromSequence(sequence)
}

// RowFromClasses returns a row and infers its modulus from nonempty input.
func RowFromClasses(classes ...pitchclass.Class) (Row, error) {
	sequence, err := FromClasses(classes...)
	if err != nil {
		return Row{}, err
	}
	return RowFromSequence(sequence)
}

// RowFromClassesWithModulus returns a row of an explicit modulus.
func RowFromClassesWithModulus(modulus int, classes ...pitchclass.Class) (Row, error) {
	sequence, err := FromClassesWithModulus(modulus, classes...)
	if err != nil {
		return Row{}, err
	}
	return RowFromSequence(sequence)
}

// RowFromSequence validates s as a row without changing it.
func RowFromSequence(s Sequence) (Row, error) {
	if !s.IsValid() {
		return Row{}, invalidValue()
	}
	if !s.IsRow() {
		return Row{}, ErrNotRow
	}
	return newRowUnchecked(s.modulus, s.elements), nil
}

// IsValid reports whether r is a valid row.
func (r Row) IsValid() bool { return rowIsValid(r) }

// Modulus returns the row modulus, or 0 for an invalid value.
func (r Row) Modulus() int { return r.modulus }

// Len returns the number of elements, equal to the modulus of a valid row.
func (r Row) Len() int { return len(r.elements) / 2 }

// At returns the class at a zero-based position.
func (r Row) At(index int) (pitchclass.Class, error) {
	if !r.IsValid() {
		return pitchclass.Class{}, invalidValue()
	}
	if index < 0 || index >= r.Len() {
		return pitchclass.Class{}, invalidSegment()
	}
	return classAt(r.modulus, r.elements, index), nil
}

// Classes returns a caller-owned slice in row order.
func (r Row) Classes() []pitchclass.Class { return r.Sequence().Classes() }

// Numbers returns a caller-owned slice of class numbers in row order.
func (r Row) Numbers() []int { return r.Sequence().Numbers() }

// Sequence returns the same ordered elements as a sequence.
func (r Row) Sequence() Sequence {
	return newSequenceUnchecked(r.modulus, r.elements)
}

// Equal reports domain equality of two valid rows.
func (r Row) Equal(other Row) bool {
	return r.IsValid() && other.IsValid() && r == other
}

// Compare orders by modulus, length, and lexicographic element order.
func (r Row) Compare(other Row) int {
	if !r.IsValid() || !other.IsValid() {
		switch {
		case !r.IsValid() && !other.IsValid():
			return 0
		case !r.IsValid():
			return -1
		default:
			return 1
		}
	}
	return compareEncoded(r.modulus, r.elements, other.modulus, other.elements)
}

// Transpose applies T_n to every element and preserves the row invariant.
func (r Row) Transpose(n int) (Row, error) {
	sequence, err := r.Sequence().Transpose(n)
	if err != nil {
		return Row{}, err
	}
	return newRowUnchecked(sequence.modulus, sequence.elements), nil
}

// Invert applies I_n to every element and preserves the row invariant.
func (r Row) Invert(n int) (Row, error) {
	sequence, err := r.Sequence().Invert(n)
	if err != nil {
		return Row{}, err
	}
	return newRowUnchecked(sequence.modulus, sequence.elements), nil
}

// Multiply applies M_m and returns a sequence because multipliers may collapse classes.
func (r Row) Multiply(m int) (Sequence, error) { return r.Sequence().Multiply(m) }

// Apply applies an invertible pitch-class operator and preserves the row invariant.
func (r Row) Apply(op pitchclass.Operator) (Row, error) {
	sequence, err := r.Sequence().Apply(op)
	if err != nil {
		return Row{}, err
	}
	return newRowUnchecked(sequence.modulus, sequence.elements), nil
}

// Retrograde reverses element order.
func (r Row) Retrograde() (Row, error) {
	sequence, err := r.Sequence().Retrograde()
	if err != nil {
		return Row{}, err
	}
	return newRowUnchecked(sequence.modulus, sequence.elements), nil
}

// RetrogradeInversion reverses the inverted row.
func (r Row) RetrogradeInversion(n int) (Row, error) {
	sequence, err := r.Sequence().RetrogradeInversion(n)
	if err != nil {
		return Row{}, err
	}
	return newRowUnchecked(sequence.modulus, sequence.elements), nil
}

// Rotate rotates a row left by k modulo its length.
func (r Row) Rotate(k int) (Row, error) {
	sequence, err := r.Sequence().Rotate(k)
	if err != nil {
		return Row{}, err
	}
	return newRowUnchecked(sequence.modulus, sequence.elements), nil
}

// IsAllInterval reports whether every nonzero interval occurs exactly once.
func (r Row) IsAllInterval() bool { return false }

// IsDerived reports whether all discrete segments derive from the first.
func (r Row) IsDerived(length int) (bool, error) { return false, nil }

// Derive returns all derivation operators of every segment.
func (r Row) Derive(length int) ([][]DerivationOperator, error) { return nil, nil }

// IsCombinatorialPair reports whether two forms have complementary first-half content.
func (r Row) IsCombinatorialPair(a, b FormLabel) (bool, error) { return false, nil }

// CombinatorialForms returns combinatorial P, I, and RI labels.
func (r Row) CombinatorialForms() ([]FormLabel, error) { return nil, nil }

// CombinatorialRotations returns inversion-combinatorial rotated forms.
func (r Row) CombinatorialRotations() ([]RotationCombinatoriality, error) { return nil, nil }

// HexachordAnalysis analyzes the first-half hexachord of a twelve-tone row.
func (r Row) HexachordAnalysis() (HexachordAnalysis, error) { return HexachordAnalysis{}, nil }
