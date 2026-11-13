package frequency

import "strconv"

// Kind identifies the form of a Criterion. The zero Kind is the kind of an
// unset criterion; values outside the declared constants are unknown kinds.
type Kind uint8

const (
	KindStrict   Kind = 1 + iota // Exact float64 equality without a hidden tolerance.
	KindCents                    // A tolerance measured in cents.
	KindHertz                    // An absolute tolerance measured in hertz.
	KindRelative                 // A tolerance relative to the greater frequency.
	KindAllOf                    // All components must report closeness.
	KindAnyOf                    // At least one component must report closeness.
)

// Kinds returns all six valid kinds in declaration order. The returned slice
// has length and capacity six and belongs to the caller.
func Kinds() []Kind {
	return []Kind{KindStrict, KindCents, KindHertz, KindRelative, KindAllOf, KindAnyOf}
}

// IsValid reports whether k is one of the six declared kinds.
func (k Kind) IsValid() bool {
	return k >= KindStrict && k <= KindAnyOf
}

// Equal reports whether k and other are the same kind. It is equivalent to ==.
func (k Kind) Equal(other Kind) bool {
	return k == other
}

// HasTolerance reports whether criteria of kind k carry a tolerance. A cents
// tolerance is read with Criterion.CentsTolerance; the other two kinds use
// Criterion.Tolerance.
func (k Kind) HasTolerance() bool {
	return k == KindCents || k == KindHertz || k == KindRelative
}

// IsComposite reports whether criteria of kind k contain components.
func (k Kind) IsComposite() bool {
	return k == KindAllOf || k == KindAnyOf
}

// String returns "strict", "cents", "hertz", "relative", "all", or "any"
// for a declared kind, "unset" for zero, and "Kind(n)" for an unknown kind.
func (k Kind) String() string {
	switch k {
	case 0:
		return "unset"
	case KindStrict:
		return "strict"
	case KindCents:
		return "cents"
	case KindHertz:
		return "hertz"
	case KindRelative:
		return "relative"
	case KindAllOf:
		return "all"
	case KindAnyOf:
		return "any"
	default:
		return "Kind(" + strconv.FormatUint(uint64(k), 10) + ")"
	}
}
