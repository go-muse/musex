package measure

// Alignment places an incomplete bar inside a full bar of its meter. It
// decides where the bar lies in its meter: the place in the meter and the
// count of units by the place in the meter. Lengths and starts of bars do
// not depend on it.
//
// The zero value and values outside the constants are invalid: IsValid
// reports false, and operations that take an Alignment reject them with
// ErrInvalidArgument. Alignment is comparable with ==, which matches Equal.
type Alignment uint8

const (
	// AlignStart begins the bar with the first unit of the meter, as a bar
	// that starts on the downbeat.
	AlignStart Alignment = iota + 1 // begins with the first unit of the meter
	// AlignEnd ends the bar where a full bar would end, as a pickup: the bar
	// takes the last part of a full bar of its meter. A pickup is always
	// aligned to the end.
	AlignEnd // ends where a full bar would end, as a pickup
)

// IsValid reports whether a is one of the constants AlignStart and AlignEnd.
func (a Alignment) IsValid() bool {
	return a == AlignStart || a == AlignEnd
}

// Equal reports whether a and other are the same value; it is the same as ==
// for all values, including invalid ones.
func (a Alignment) Equal(other Alignment) bool {
	return a == other
}

// String returns "start" or "end", or "Alignment(invalid)" for an invalid
// value.
func (a Alignment) String() string {
	switch a {
	case AlignStart:
		return "start"
	case AlignEnd:
		return "end"
	default:
		return "Alignment(invalid)"
	}
}

// Kind compares the length of a bar with the size of its meter: full, when
// they are equal; incomplete, when the bar is shorter; extended, when it is
// longer.
//
// The zero value and values outside the constants are invalid: IsValid
// reports false. Kind is comparable with ==, which matches Equal.
type Kind uint8

const (
	// KindFull is a bar whose length equals the size of its meter.
	KindFull Kind = iota + 1 // length equals the size of the meter
	// KindIncomplete is a bar shorter than its meter, such as a pickup.
	KindIncomplete // shorter than the meter
	// KindExtended is a bar longer than its meter.
	KindExtended // longer than the meter
)

// IsValid reports whether k is one of the constants KindFull,
// KindIncomplete, and KindExtended.
func (k Kind) IsValid() bool {
	return k >= KindFull && k <= KindExtended
}

// Equal reports whether k and other are the same value; it is the same as ==
// for all values, including invalid ones.
func (k Kind) Equal(other Kind) bool {
	return k == other
}

// String returns "full", "incomplete", or "extended", or "Kind(invalid)" for
// an invalid value.
func (k Kind) String() string {
	switch k {
	case KindFull:
		return "full"
	case KindIncomplete:
		return "incomplete"
	case KindExtended:
		return "extended"
	default:
		return "Kind(invalid)"
	}
}

// Level selects the boundaries at which Grid.SplitBy divides a duration:
// bar lines, and also the starts of groups, of beats, or of units of the
// meter of every spanned bar. A beat is a beat of package meter
// (meter.Meter.Beats), and a unit is a unit of the time signature
// (meter.Meter.Units).
//
// The zero value and values outside the constants are invalid: IsValid
// reports false, and SplitBy rejects them with ErrInvalidArgument. Level is
// comparable with ==, which matches Equal.
type Level uint8

const (
	// LevelBars splits at bar lines only.
	LevelBars Level = iota + 1 // bar lines
	// LevelGroups splits at bar lines and at the starts of groups.
	LevelGroups // and starts of groups
	// LevelBeats splits at bar lines and at the starts of beats.
	LevelBeats // and starts of beats
	// LevelUnits splits at bar lines and at the starts of units of the meter.
	LevelUnits // and starts of units of the meter
)

// IsValid reports whether l is one of the constants LevelBars, LevelGroups,
// LevelBeats, and LevelUnits.
func (l Level) IsValid() bool {
	return l >= LevelBars && l <= LevelUnits
}

// Equal reports whether l and other are the same value; it is the same as ==
// for all values, including invalid ones.
func (l Level) Equal(other Level) bool {
	return l == other
}

// String returns "bars", "groups", "beats", or "units", or "Level(invalid)"
// for an invalid value.
func (l Level) String() string {
	switch l {
	case LevelBars:
		return "bars"
	case LevelGroups:
		return "groups"
	case LevelBeats:
		return "beats"
	case LevelUnits:
		return "units"
	default:
		return "Level(invalid)"
	}
}
