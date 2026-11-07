package setclass

import (
	"cmp"
	"fmt"

	"github.com/go-muse/musex/pcset"
)

// Kind represents the canonical equivalence group G over which a SetClass is defined.
// The zero value Kind(0) is invalid (IsValid() == false).
type Kind uint8

const (
	// Tn represents transpositional equivalence under the cyclic group T_n = {T_0, ..., T_{N-1}}.
	Tn Kind = 1
	// TnIn represents transpositional and inversional equivalence under the dihedral group T_n/I_n.
	TnIn Kind = 2
	// TTO represents affine equivalence under the group of twelve-tone operators T_n M_m with gcd(m, N) = 1.
	TTO Kind = 3
)

// Kinds returns the 3 valid equivalence group kinds in canonical order: [Tn, TnIn, TTO].
func Kinds() [3]Kind {
	return [3]Kind{Tn, TnIn, TTO}
}

// ParseKind parses the canonical string representation of a Kind ("Tn", "TnIn", or "TTO").
// Any other string returns the zero Kind and an error wrapping [ErrNotation] and [ErrInvalidKind].
func ParseKind(s string) (Kind, error) {
	switch s {
	case "Tn":
		return Tn, nil
	case "TnIn":
		return TnIn, nil
	case "TTO":
		return TTO, nil
	default:
		return 0, fmt.Errorf("%w (%w): %q", ErrNotation, ErrInvalidKind, s)
	}
}

// MustParseKind is like [ParseKind] but panics on error.
func MustParseKind(s string) Kind {
	k, err := ParseKind(s)
	if err != nil {
		panic(err)
	}
	return k
}

// IsValid reports whether k is a known equivalence group kind (Tn, TnIn, or TTO).
func (k Kind) IsValid() bool {
	return k >= Tn && k <= TTO
}

// Equal reports whether k and other are both valid and represent the same Kind.
// If either k or other is invalid, Equal returns false.
func (k Kind) Equal(other Kind) bool {
	return k.IsValid() && other.IsValid() && k == other
}

// Compare returns -1 if k < other, 0 if k == other, and +1 if k > other (Tn < TnIn < TTO).
// Any invalid Kind value is ordered before all valid values; two invalid values return 0.
func (k Kind) Compare(other Kind) int {
	kv, ov := k.IsValid(), other.IsValid()
	if !kv && !ov {
		return 0
	}
	if !kv {
		return -1
	}
	if !ov {
		return 1
	}
	return cmp.Compare(k, other)
}

// String returns "Tn", "TnIn", "TTO", or "Kind(invalid)" for any invalid value.
func (k Kind) String() string {
	switch k {
	case Tn:
		return "Tn"
	case TnIn:
		return "TnIn"
	case TTO:
		return "TTO"
	default:
		return "Kind(invalid)"
	}
}

// Algorithm represents the canonical prime form packing convention.
// The zero value Algorithm(0) is invalid (IsValid() == false).
type Algorithm uint8

const (
	// Rahn selects John Rahn's (1980) prime form packing algorithm (left-packed from the right; default in musex).
	Rahn Algorithm = 1
	// Forte selects Allen Forte's (1973) prime form packing algorithm (packed from the left).
	Forte Algorithm = 2
)

// Algorithms returns the 2 valid prime form algorithms in canonical order: [Rahn, Forte].
func Algorithms() [2]Algorithm {
	return [2]Algorithm{Rahn, Forte}
}

// ParseAlgorithm parses the canonical string representation of an Algorithm ("Rahn" or "Forte").
// Any other string returns the zero Algorithm and an error wrapping [ErrNotation] and [ErrInvalidAlgorithm].
func ParseAlgorithm(s string) (Algorithm, error) {
	switch s {
	case "Rahn":
		return Rahn, nil
	case "Forte":
		return Forte, nil
	default:
		return 0, fmt.Errorf("%w (%w): %q", ErrNotation, ErrInvalidAlgorithm, s)
	}
}

// MustParseAlgorithm is like [ParseAlgorithm] but panics on error.
func MustParseAlgorithm(s string) Algorithm {
	a, err := ParseAlgorithm(s)
	if err != nil {
		panic(err)
	}
	return a
}

// IsValid reports whether a is a known prime form algorithm (Rahn or Forte).
func (a Algorithm) IsValid() bool {
	return a == Rahn || a == Forte
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

func toPCSetAlgorithm(a Algorithm) pcset.Algorithm {
	if a == Forte {
		return pcset.Forte
	}
	return pcset.Rahn
}
