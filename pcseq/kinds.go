package pcseq

import "cmp"

// FormKind identifies a labelled row form.
type FormKind uint8

// Row form families in canonical order.
const (
	P FormKind = iota + 1
	I
	R
	RI
)

// IsValid reports whether k is P, I, R, or RI.
func (k FormKind) IsValid() bool { return k >= P && k <= RI }

// Equal reports whether both kinds are valid and equal.
func (k FormKind) Equal(other FormKind) bool { return k.IsValid() && other.IsValid() && k == other }

// Compare orders invalid values first, then P, I, R, and RI.
func (k FormKind) Compare(other FormKind) int {
	if !k.IsValid() || !other.IsValid() {
		switch {
		case !k.IsValid() && !other.IsValid():
			return 0
		case !k.IsValid():
			return -1
		default:
			return 1
		}
	}
	return cmp.Compare(k, other)
}

// String returns the form family name or FormKind(invalid).
func (k FormKind) String() string {
	switch k {
	case P:
		return "P"
	case I:
		return "I"
	case R:
		return "R"
	case RI:
		return "RI"
	default:
		return "FormKind(invalid)"
	}
}

// RowClassKind identifies a row equivalence relation.
type RowClassKind uint8

// Row equivalence relations in canonical order.
const (
	TIR RowClassKind = iota + 1
	TIRRotations
	TTOR
)

// IsValid reports whether k names a supported row equivalence relation.
func (k RowClassKind) IsValid() bool { return k >= TIR && k <= TTOR }

// Equal reports whether both kinds are valid and equal.
func (k RowClassKind) Equal(other RowClassKind) bool {
	return k.IsValid() && other.IsValid() && k == other
}

// Compare orders invalid values first, then TIR, TIRRotations, and TTOR.
func (k RowClassKind) Compare(other RowClassKind) int {
	if !k.IsValid() || !other.IsValid() {
		switch {
		case !k.IsValid() && !other.IsValid():
			return 0
		case !k.IsValid():
			return -1
		default:
			return 1
		}
	}
	return cmp.Compare(k, other)
}

// String returns the relation name or RowClassKind(invalid).
func (k RowClassKind) String() string {
	switch k {
	case TIR:
		return "TIR"
	case TIRRotations:
		return "TIRRotations"
	case TTOR:
		return "TTOR"
	default:
		return "RowClassKind(invalid)"
	}
}

// DerivationFamily identifies an ordered-segment transformation.
type DerivationFamily uint8

// Derivation families in canonical order.
const (
	DerivationT DerivationFamily = iota + 1
	DerivationI
	DerivationTR
	DerivationIR
)

// IsValid reports whether k names a derivation family.
func (k DerivationFamily) IsValid() bool { return k >= DerivationT && k <= DerivationIR }

// Equal reports whether both families are valid and equal.
func (k DerivationFamily) Equal(other DerivationFamily) bool {
	return k.IsValid() && other.IsValid() && k == other
}

// Compare orders invalid values first, then T, I, TR, and IR.
func (k DerivationFamily) Compare(other DerivationFamily) int {
	if !k.IsValid() || !other.IsValid() {
		switch {
		case !k.IsValid() && !other.IsValid():
			return 0
		case !k.IsValid():
			return -1
		default:
			return 1
		}
	}
	return cmp.Compare(k, other)
}

// String returns the derivation family name or DerivationFamily(invalid).
func (k DerivationFamily) String() string {
	switch k {
	case DerivationT:
		return "T"
	case DerivationI:
		return "I"
	case DerivationTR:
		return "TR"
	case DerivationIR:
		return "IR"
	default:
		return "DerivationFamily(invalid)"
	}
}

// AllIntervalCountKind identifies a confirmed all-interval row count.
type AllIntervalCountKind uint8

// Confirmed all-interval counts in canonical order.
const (
	AllIntervalTotal AllIntervalCountKind = iota + 1
	AllIntervalFromZero
	AllIntervalUpToTI
	AllIntervalUpToTIR
	AllIntervalRSymmetric
	AllIntervalRISymmetric
)

// IsValid reports whether k names a supported count.
func (k AllIntervalCountKind) IsValid() bool {
	return k >= AllIntervalTotal && k <= AllIntervalRISymmetric
}
