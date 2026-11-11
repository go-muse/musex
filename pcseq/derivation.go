package pcseq

// DerivationOperator applies T_n, I_n, T_nR, or I_nR to an ordered segment.
type DerivationOperator struct {
	family  DerivationFamily
	index   int
	modulus int
}

// NewDerivationOperator returns an operator with index in 0..modulus-1.
func NewDerivationOperator(family DerivationFamily, index, modulus int) (DerivationOperator, error) {
	return DerivationOperator{}, nil
}

// IsValid reports whether o is a valid derivation operator.
func (o DerivationOperator) IsValid() bool { return false }

// Family returns the derivation family.
func (o DerivationOperator) Family() DerivationFamily { return 0 }

// Index returns the operator index, or 0 for an invalid value.
func (o DerivationOperator) Index() int { return 0 }

// Modulus returns the operator modulus, or 0 for an invalid value.
func (o DerivationOperator) Modulus() int { return 0 }

// Equal reports domain equality of two valid operators.
func (o DerivationOperator) Equal(other DerivationOperator) bool { return false }

// Compare orders by modulus, family, and index.
func (o DerivationOperator) Compare(other DerivationOperator) int { return 0 }

// Apply transforms a sequence of the same modulus.
func (o DerivationOperator) Apply(s Sequence) (Sequence, error) { return Sequence{}, nil }

// String returns T<n>, I<n>, T<n>R, I<n>R, or DerivationOperator(invalid).
func (o DerivationOperator) String() string { return "" }
