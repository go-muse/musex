package rounding

import "fmt"

// Mode selects one of two neighboring grid values. Its zero value is unset;
// the package deliberately has no default mode.
type Mode uint8

const (
	// Down always selects the lower neighbor, toward negative infinity.
	Down Mode = 1 + iota
	// Up always selects the upper neighbor, toward positive infinity.
	Up
	// TowardZero selects the neighbor nearest to the grid's zero.
	TowardZero
	// AwayFromZero selects the neighbor farthest from the grid's zero.
	AwayFromZero
	// HalfUp selects the nearest neighbor and resolves ties toward positive infinity.
	HalfUp
	// HalfDown selects the nearest neighbor and resolves ties toward negative infinity.
	HalfDown
	// HalfAwayFromZero selects the nearest neighbor and resolves ties away from zero.
	HalfAwayFromZero
	// HalfTowardZero selects the nearest neighbor and resolves ties toward zero.
	HalfTowardZero
	// HalfEven selects the nearest neighbor and resolves ties to an even index.
	HalfEven
)

var modeNames = [...]string{
	"unset",
	"down",
	"up",
	"toward-zero",
	"away-from-zero",
	"half-up",
	"half-down",
	"half-away-from-zero",
	"half-toward-zero",
	"half-even",
}

// Modes returns all nine modes in their declaration order. The returned slice
// has length and capacity nine and belongs to the caller.
func Modes() []Mode {
	return []Mode{Down, Up, TowardZero, AwayFromZero, HalfUp, HalfDown, HalfAwayFromZero, HalfTowardZero, HalfEven}
}

// IsValid reports whether m is one of the nine defined modes.
func (m Mode) IsValid() bool {
	return m >= Down && m <= HalfEven
}

// Equal reports whether m and other are the same mode value. It is equivalent
// to == for valid, unset, and unknown values.
func (m Mode) Equal(other Mode) bool {
	return m == other
}

// IsDirected reports whether m always chooses by direction rather than distance.
func (m Mode) IsDirected() bool {
	return m >= Down && m <= AwayFromZero
}

// IsNearest reports whether m chooses the nearest neighbor with a tie rule.
func (m Mode) IsNearest() bool {
	return m >= HalfUp && m <= HalfEven
}

// Mirror returns the mode that gives the negated result for a negated input.
// Down and Up are exchanged, as are HalfUp and HalfDown. Other modes, unset,
// and unknown values are returned unchanged.
func (m Mode) Mirror() Mode {
	switch m {
	case Down:
		return Up
	case Up:
		return Down
	case HalfUp:
		return HalfDown
	case HalfDown:
		return HalfUp
	default:
		return m
	}
}

// String returns the canonical name, "unset" for the zero mode, or a numeric
// diagnostic form for an unknown mode.
func (m Mode) String() string {
	if m < Mode(len(modeNames)) {
		return modeNames[m]
	}
	return fmt.Sprintf("Mode(%d)", uint8(m))
}
