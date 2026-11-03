package spelling

import "strconv"

// Direction is the direction of the alteration value of a name, as returned
// by Spelling.Direction: lowered, natural, or raised, or unset for the unset
// name.
//
// The zero value is DirectionUnset. Direction is a comparable value type that
// can be used in a switch or as a map key. The numeric values of the constants
// (0 to 3) are stable because reflection-based encoders such as encoding/json
// output them; they are not a sign such as -1, 0, or +1. Use Direction.String
// or the constants for output and comparisons.
type Direction uint8

// The directions of names.
const (
	// DirectionUnset is the direction of the unset name. No valid name has
	// it. It is the zero value of Direction.
	DirectionUnset Direction = iota

	// DirectionLowered is the direction of a name whose alteration value is
	// below 0, such as D flat.
	DirectionLowered

	// DirectionNatural is the direction of a natural name, whose alteration
	// value is 0, such as C.
	DirectionNatural

	// DirectionRaised is the direction of a name whose alteration value is
	// above 0, such as C sharp.
	DirectionRaised
)

// String returns the name of d: "unset", "lowered", "natural", or "raised".
// For any other value, it returns "Direction(n)", where n is the decimal
// numeric value of d, for example "Direction(9)".
func (d Direction) String() string {
	switch d {
	case DirectionUnset:
		return "unset"
	case DirectionLowered:
		return "lowered"
	case DirectionNatural:
		return "natural"
	case DirectionRaised:
		return "raised"
	default:
		return "Direction(" + strconv.Itoa(int(d)) + ")"
	}
}
