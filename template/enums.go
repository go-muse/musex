package template

import "fmt"

// Kind identifies whether a template repeats after a cycle or ends after its
// last place.
type Kind uint8

const (
	// KindUnset is the kind of the zero, unset Template.
	KindUnset Kind = iota
	// Closed identifies a template whose last adjacent interval returns to its
	// first place in the next cycle.
	Closed
	// Open identifies a template without a closing interval or cycle.
	Open
)

// IsValid reports whether k is Closed or Open.
func (k Kind) IsValid() bool {
	return k == Closed || k == Open
}

// String returns "unset", "closed", or "open" for a known kind.
func (k Kind) String() string {
	switch k {
	case KindUnset:
		return "unset"
	case Closed:
		return "closed"
	case Open:
		return "open"
	default:
		return fmt.Sprintf("Kind(%d)", uint8(k))
	}
}

// Measure selects which intervals are compared by MismatchingPlaces and
// Catalog.Related.
type Measure uint8

const (
	// MeasureUnset is not a valid measure for an operation.
	MeasureUnset Measure = iota
	// FromFirst compares intervals from the first place to each later place,
	// including the cycle of a closed template as place Len()+1.
	FromFirst
	// Adjacent compares intervals between neighboring places, including the
	// closing interval of a closed template.
	Adjacent
)

// IsValid reports whether m is FromFirst or Adjacent.
func (m Measure) IsValid() bool {
	return m == FromFirst || m == Adjacent
}

// String returns "unset", "from-first", or "adjacent" for a known measure.
func (m Measure) String() string {
	switch m {
	case MeasureUnset:
		return "unset"
	case FromFirst:
		return "from-first"
	case Adjacent:
		return "adjacent"
	default:
		return fmt.Sprintf("Measure(%d)", uint8(m))
	}
}
