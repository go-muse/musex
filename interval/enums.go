package interval

import "strconv"

// Direction is the direction of an interval: up or down.
//
// The zero value is DirectionUnset. Only Up and Down are set directions:
// an operation that needs a direction reports ErrNoDirection for any other
// value, including values such as Direction(9). The numeric values of the
// constants (0 to 2) are stable because reflection-based encoders output them.
type Direction uint8

// The directions of intervals.
const (
	// DirectionUnset is the zero value: no direction is given. No interval
	// has it, and the unset Interval reports it.
	DirectionUnset Direction = iota
	// Up is the direction of an ascending interval. The perfect unison is
	// always ascending.
	Up
	// Down is the direction of a descending interval.
	Down
)

// String returns the name of d: "unset", "up", or "down". For any other value,
// it returns "Direction(n)", where n is the decimal numeric value of d.
func (d Direction) String() string {
	switch d {
	case DirectionUnset:
		return "unset"
	case Up:
		return "up"
	case Down:
		return "down"
	default:
		return "Direction(" + strconv.Itoa(int(d)) + ")"
	}
}

// isSet reports whether d is Up or Down.
func (d Direction) isSet() bool {
	return d == Up || d == Down
}

// Quality is the kind of the quality of an interval, as returned by
// Interval.Quality: perfect, major, minor, augmented, diminished, neutral,
// or another fractional quality. The multiplicity of augmented and diminished
// intervals is reported separately by Interval.Multiplicity.
//
// The zero value is QualityUnset. The numeric values of the constants (0 to 7)
// are stable because reflection-based encoders output them.
type Quality uint8

// The kinds of quality.
const (
	// QualityUnset is the zero value: no quality is given. The unset Interval
	// reports it.
	QualityUnset Quality = iota
	// Perfect is the quality of a perfect interval: deviation 0 in the
	// perfect group.
	Perfect
	// Major is the quality of a major interval: deviation 0 in the
	// major-minor group.
	Major
	// Minor is the quality of a minor interval: deviation -1 in the
	// major-minor group.
	Minor
	// Augmented is the quality of an augmented interval of multiplicity k:
	// a whole deviation k >= 1.
	Augmented
	// Diminished is the quality of a diminished interval of multiplicity k:
	// a whole deviation -k in the perfect group or -1-k in the major-minor
	// group, k >= 1.
	Diminished
	// Neutral is the quality of a neutral interval: deviation exactly -1/2 in
	// the major-minor group.
	Neutral
	// Fractional is the kind of any other fractional quality. New and
	// NewMultiple do not accept it; use FromDeviation.
	Fractional
)

// String returns the name of q: "unset", "perfect", "major", "minor",
// "augmented", "diminished", "neutral", or "fractional". For any other value,
// it returns "Quality(n)", where n is the decimal numeric value of q.
func (q Quality) String() string {
	switch q {
	case QualityUnset:
		return "unset"
	case Perfect:
		return "perfect"
	case Major:
		return "major"
	case Minor:
		return "minor"
	case Augmented:
		return "augmented"
	case Diminished:
		return "diminished"
	case Neutral:
		return "neutral"
	case Fractional:
		return "fractional"
	default:
		return "Quality(" + strconv.Itoa(int(q)) + ")"
	}
}

// Group is the group of the size of an interval, by its simple part: the
// perfect group (simple sizes 1, 4, 5, 8) or the major-minor group (2, 3, 6,
// 7).
//
// The zero value is GroupUnset. The numeric values of the constants (0 to 2)
// are stable because reflection-based encoders output them.
type Group uint8

// The groups of sizes.
const (
	// GroupUnset is the zero value; the unset Interval reports it.
	GroupUnset Group = iota
	// PerfectGroup is the group of unisons, fourths, fifths, and octaves and
	// their compounds: their basic kind is perfect.
	PerfectGroup
	// MajorMinorGroup is the group of seconds, thirds, sixths, and sevenths
	// and their compounds: their basic kind is major.
	MajorMinorGroup
)

// String returns the name of g: "unset", "perfect", or "major-minor". For any
// other value, it returns "Group(n)", where n is the decimal numeric value of
// g.
func (g Group) String() string {
	switch g {
	case GroupUnset:
		return "unset"
	case PerfectGroup:
		return "perfect"
	case MajorMinorGroup:
		return "major-minor"
	default:
		return "Group(" + strconv.Itoa(int(g)) + ")"
	}
}

// Replace selects which pitch of a pair Respellings replaces: the start
// pitch, the end pitch, or both.
//
// The zero value is ReplaceUnset. Respellings reports ErrUnset for any value
// other than ReplaceFrom, ReplaceTo, and ReplaceBoth. The numeric values of
// the constants (0 to 3) are stable because reflection-based encoders output
// them.
type Replace uint8

// The choices of the replaced pitches.
const (
	// ReplaceUnset is the zero value: no choice is given.
	ReplaceUnset Replace = iota
	// ReplaceFrom replaces the start pitch and keeps the end pitch.
	ReplaceFrom
	// ReplaceTo keeps the start pitch and replaces the end pitch.
	ReplaceTo
	// ReplaceBoth replaces both pitches.
	ReplaceBoth
)

// String returns the name of r: "unset", "from", "to", or "both". For any
// other value, it returns "Replace(n)", where n is the decimal numeric value
// of r.
func (r Replace) String() string {
	switch r {
	case ReplaceUnset:
		return "unset"
	case ReplaceFrom:
		return "from"
	case ReplaceTo:
		return "to"
	case ReplaceBoth:
		return "both"
	default:
		return "Replace(" + strconv.Itoa(int(r)) + ")"
	}
}
