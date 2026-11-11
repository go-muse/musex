package pcseq

import (
	"cmp"
	"strconv"
)

// DerivationOperator applies T_n, I_n, T_nR, or I_nR to an ordered segment.
type DerivationOperator struct {
	family  DerivationFamily
	index   int
	modulus int
}

// NewDerivationOperator returns an operator with index in 0..modulus-1.
func NewDerivationOperator(family DerivationFamily, index, modulus int) (DerivationOperator, error) {
	if err := checkModulus(modulus); err != nil {
		return DerivationOperator{}, err
	}
	if !family.IsValid() || index < 0 || index >= modulus {
		return DerivationOperator{}, invalidValue()
	}
	return DerivationOperator{family: family, index: index, modulus: modulus}, nil
}

// IsValid reports whether o is a valid derivation operator.
func (o DerivationOperator) IsValid() bool {
	return o.family.IsValid() && o.modulus >= 1 && o.modulus <= MaxModulus &&
		o.index >= 0 && o.index < o.modulus
}

// Family returns the derivation family.
func (o DerivationOperator) Family() DerivationFamily { return o.family }

// Index returns the operator index, or 0 for an invalid value.
func (o DerivationOperator) Index() int { return o.index }

// Modulus returns the operator modulus, or 0 for an invalid value.
func (o DerivationOperator) Modulus() int { return o.modulus }

// Equal reports domain equality of two valid operators.
func (o DerivationOperator) Equal(other DerivationOperator) bool {
	return o.IsValid() && other.IsValid() && o == other
}

// Compare orders by modulus, family, and index.
func (o DerivationOperator) Compare(other DerivationOperator) int {
	if !o.IsValid() || !other.IsValid() {
		switch {
		case !o.IsValid() && !other.IsValid():
			return 0
		case !o.IsValid():
			return -1
		default:
			return 1
		}
	}
	if result := cmp.Compare(o.modulus, other.modulus); result != 0 {
		return result
	}
	if result := o.family.Compare(other.family); result != 0 {
		return result
	}
	return cmp.Compare(o.index, other.index)
}

// Apply transforms a sequence of the same modulus.
func (o DerivationOperator) Apply(s Sequence) (Sequence, error) {
	if !o.IsValid() || !s.IsValid() {
		return Sequence{}, invalidValue()
	}
	if o.modulus != s.modulus {
		return Sequence{}, modulusMismatch(o.modulus, s.modulus)
	}
	switch o.family {
	case DerivationT:
		return s.Transpose(o.index)
	case DerivationI:
		return s.Invert(o.index)
	case DerivationTR:
		return s.MustRetrograde().Transpose(o.index)
	case DerivationIR:
		return s.MustRetrograde().Invert(o.index)
	default:
		return Sequence{}, invalidValue()
	}
}

// String returns T<n>, I<n>, T<n>R, I<n>R, or DerivationOperator(invalid).
func (o DerivationOperator) String() string {
	if !o.IsValid() {
		return "DerivationOperator(invalid)"
	}
	index := strconv.Itoa(o.index)
	switch o.family {
	case DerivationT:
		return "T" + index
	case DerivationI:
		return "I" + index
	case DerivationTR:
		return "T" + index + "R"
	case DerivationIR:
		return "I" + index + "R"
	default:
		return "DerivationOperator(invalid)"
	}
}

// IsDerived reports whether all discrete segments derive from the first.
func (r Row) IsDerived(length int) (bool, error) {
	_, err := r.Derive(length)
	if err == ErrNotDerived {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Derive returns all derivation operators of every segment.
func (r Row) Derive(length int) ([][]DerivationOperator, error) {
	if !r.IsValid() {
		return nil, invalidValue()
	}
	if err := checkModulusLimit(r.modulus, MaxAnalysisModulus); err != nil {
		return nil, err
	}
	segments, err := r.Sequence().DiscreteSegments(length)
	if err != nil {
		return nil, err
	}
	generator := segments[0]
	result := make([][]DerivationOperator, len(segments))
	for i, segment := range segments {
		result[i] = make([]DerivationOperator, 0)
		for _, family := range []DerivationFamily{DerivationT, DerivationI, DerivationTR, DerivationIR} {
			for n := range r.modulus {
				operator := DerivationOperator{family: family, index: n, modulus: r.modulus}
				if operator.MustApply(generator) == segment {
					result[i] = append(result[i], operator)
				}
			}
		}
		if len(result[i]) == 0 {
			return nil, ErrNotDerived
		}
	}
	return result, nil
}
