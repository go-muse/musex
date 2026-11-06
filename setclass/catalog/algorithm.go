package catalog

import "cmp"

// Algorithm represents the canonical prime form packing convention.
type Algorithm uint8

const (
	// Rahn selects John Rahn's (1980) prime form packing algorithm (left-packed from the right; default in musex).
	Rahn Algorithm = iota
	// Forte selects Allen Forte's (1973) prime form packing algorithm (packed from the left).
	Forte
)

// IsValid reports whether a is a known prime form algorithm (Rahn or Forte).
func (a Algorithm) IsValid() bool {
	return a <= Forte
}

// Equal reports whether a and other are both valid and represent the same Algorithm.
// If either a or other is invalid, Equal returns false.
func (a Algorithm) Equal(other Algorithm) bool {
	return a.IsValid() && other.IsValid() && a == other
}

// Compare returns -1 if a < other, 0 if a == other, and +1 if a > other (Rahn < Forte).
// Any invalid Algorithm value is ordered before all valid values; two invalid values return 0.
func (a Algorithm) Compare(other Algorithm) int {
	av, bv := a.IsValid(), other.IsValid()
	if !av && !bv {
		return 0
	}
	if !av {
		return -1
	}
	if !bv {
		return 1
	}
	return cmp.Compare(a, other)
}

// String returns "Rahn" for Rahn, "Forte" for Forte, and "Algorithm(invalid)" for any invalid value.
func (a Algorithm) String() string {
	switch a {
	case Rahn:
		return "Rahn"
	case Forte:
		return "Forte"
	default:
		return "Algorithm(invalid)"
	}
}

// Suffix represents the chirality orientation suffix of a directed Tn-type.
type Suffix uint8

const (
	// SuffixNone indicates no orientation suffix (inversionally symmetric set class).
	SuffixNone Suffix = iota
	// SuffixA indicates the canonical "A" orientation (contains the prime form of the parent ForteClass).
	SuffixA
	// SuffixB indicates the inversional "B" orientation (the inversion of the "A" orientation).
	SuffixB
)

// IsValid reports whether s is a known orientation suffix (SuffixNone, SuffixA, or SuffixB).
func (s Suffix) IsValid() bool {
	return s <= SuffixB
}

// Equal reports whether s and other are both valid and represent the same Suffix.
// If either s or other is invalid, Equal returns false.
func (s Suffix) Equal(other Suffix) bool {
	return s.IsValid() && other.IsValid() && s == other
}

// Compare returns -1 if s < other, 0 if s == other, and +1 if s > other (SuffixNone < SuffixA < SuffixB).
// Any invalid Suffix value is ordered before all valid values; two invalid values return 0.
func (s Suffix) Compare(other Suffix) int {
	sv, bv := s.IsValid(), other.IsValid()
	if !sv && !bv {
		return 0
	}
	if !sv {
		return -1
	}
	if !bv {
		return 1
	}
	return cmp.Compare(s, other)
}

// String returns "" for SuffixNone, "A" for SuffixA, "B" for SuffixB, and "Suffix(invalid)" for any invalid value.
func (s Suffix) String() string {
	switch s {
	case SuffixNone:
		return ""
	case SuffixA:
		return "A"
	case SuffixB:
		return "B"
	default:
		return "Suffix(invalid)"
	}
}
