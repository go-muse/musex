package rounding

import (
	"fmt"
	"math"
)

// Side is the position of a value relative to the grid's zero.
type Side int8

const (
	// Negative marks a value below the grid's zero.
	Negative Side = -1
	// Positive marks a value above the grid's zero.
	Positive Side = 1
)

// IsValid reports whether s is Negative or Positive.
func (s Side) IsValid() bool {
	return s == Negative || s == Positive
}

// String returns "negative", "positive", or a numeric diagnostic form for an
// invalid side.
func (s Side) String() string {
	switch s {
	case Negative:
		return "negative"
	case Positive:
		return "positive"
	default:
		return fmt.Sprintf("Side(%d)", int8(s))
	}
}

// Neighbor identifies one of the two adjacent grid values. Its zero value is
// returned only together with an error.
type Neighbor uint8

const (
	// Lower identifies the neighbor below the input value.
	Lower Neighbor = 1 + iota
	// Upper identifies the neighbor above the input value.
	Upper
)

// String returns "lower", "upper", or a numeric diagnostic form for an invalid
// neighbor.
func (n Neighbor) String() string {
	switch n {
	case Lower:
		return "lower"
	case Upper:
		return "upper"
	default:
		return fmt.Sprintf("Neighbor(%d)", uint8(n))
	}
}

// Choose selects a neighbor for a value strictly between two grid nodes. cmp
// compares the lower and upper distances: negative means the lower is nearer,
// zero means a tie, and positive means the upper is nearer. side and lowerEven
// describe the caller's grid. Invalid side returns ErrInvalidArgument; mode
// failures precede it.
func (m Mode) Choose(cmp int, side Side, lowerEven bool) (Neighbor, error) {
	if err := validateMode(m); err != nil {
		return 0, err
	}
	if !side.IsValid() {
		return 0, ErrInvalidArgument
	}
	return m.choose(cmp, side, lowerEven), nil
}

// MustChoose calls Choose and panics with its error on failure.
func (m Mode) MustChoose(cmp int, side Side, lowerEven bool) Neighbor {
	n, err := m.Choose(cmp, side, lowerEven)
	if err != nil {
		panic(err)
	}
	return n
}

// ChooseIndex selects lower or lower+1 on the integer index grid. The grid's
// zero is index zero and parity is the index parity. Selecting upper at
// math.MaxInt returns ErrOutOfRange.
func (m Mode) ChooseIndex(lower, cmp int) (int, error) {
	if err := validateMode(m); err != nil {
		return 0, err
	}
	side := Positive
	if lower < 0 {
		side = Negative
	}
	if m.choose(cmp, side, lower%2 == 0) == Upper {
		if lower == math.MaxInt {
			return 0, ErrOutOfRange
		}
		return lower + 1, nil
	}
	return lower, nil
}

// MustChooseIndex calls ChooseIndex and panics with its error on failure.
func (m Mode) MustChooseIndex(lower, cmp int) int {
	k, err := m.ChooseIndex(lower, cmp)
	if err != nil {
		panic(err)
	}
	return k
}

func (m Mode) choose(cmp int, side Side, lowerEven bool) Neighbor {
	switch m {
	case Down:
		return Lower
	case Up:
		return Upper
	case TowardZero:
		if side == Positive {
			return Lower
		}
		return Upper
	case AwayFromZero:
		if side == Positive {
			return Upper
		}
		return Lower
	}

	if cmp < 0 {
		return Lower
	}
	if cmp > 0 {
		return Upper
	}

	switch m {
	case HalfUp:
		return Upper
	case HalfDown:
		return Lower
	case HalfAwayFromZero:
		if side == Positive {
			return Upper
		}
		return Lower
	case HalfTowardZero:
		if side == Positive {
			return Lower
		}
		return Upper
	case HalfEven:
		if lowerEven {
			return Lower
		}
		return Upper
	default:
		panic("rounding: unchecked mode")
	}
}
