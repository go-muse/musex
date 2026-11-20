package ticks

import (
	"math/big"
	"strconv"

	"github.com/go-muse/musex/duration"
)

// MIDIQuarter is the written length of the MIDI quarter: bb notated
// thirty-second notes (field bb of the SMF Time Signature event), so the MIDI
// quarter is written as w = bb/32 of a whole note. The zero value is unset.
// MIDIQuarter is comparable: a == b reports the same as a.Equal(b).
type MIDIQuarter struct {
	bb uint8
}

// NewMIDIQuarter returns the MIDI quarter written as thirtySeconds notated
// thirty-second notes. Zero is ErrZero, a negative value ErrInvalidArgument and
// a value above 255 ErrFormatLimit.
func NewMIDIQuarter(thirtySeconds int) (MIDIQuarter, error) {
	const op = "NewMIDIQuarter"
	switch {
	case thirtySeconds < 0:
		return MIDIQuarter{}, invalidError(op, "negative thirty-seconds")
	case thirtySeconds == 0:
		return MIDIQuarter{}, newError(ErrZero, op, "zero thirty-seconds")
	case thirtySeconds > MaxByteField:
		return MIDIQuarter{}, newError(ErrFormatLimit, op, "thirty-seconds above 255")
	}
	return MIDIQuarter{bb: uint8(thirtySeconds)}, nil
}

// MustNewMIDIQuarter is like NewMIDIQuarter but panics with its error.
func MustNewMIDIQuarter(thirtySeconds int) MIDIQuarter { return must(NewMIDIQuarter(thirtySeconds)) }

// MIDIQuarterOf returns the MIDI quarter written as the duration written. An
// unset duration is ErrInvalidArgument with duration.ErrUnset; a duration
// that is not a whole number of thirty-second notes is ErrNotRepresentable and
// more than 255 of them ErrFormatLimit.
func MIDIQuarterOf(written duration.Duration) (MIDIQuarter, error) {
	const op = "MIDIQuarterOf"
	if !written.IsSet() {
		return MIDIQuarter{}, invalidError(op, "written duration", duration.ErrUnset)
	}
	n := new(big.Rat).Mul(ratOfDuration(written), big.NewRat(32, 1))
	if !n.IsInt() {
		return MIDIQuarter{}, newError(ErrNotRepresentable, op, "not a whole number of thirty-second notes")
	}
	if n.Num().Cmp(big.NewInt(MaxByteField)) > 0 {
		return MIDIQuarter{}, newError(ErrFormatLimit, op, "thirty-seconds above 255")
	}
	return MIDIQuarter{bb: uint8(n.Num().Int64())}, nil
}

// MustMIDIQuarterOf is like MIDIQuarterOf but panics with its error.
func MustMIDIQuarterOf(written duration.Duration) MIDIQuarter { return must(MIDIQuarterOf(written)) }

// StandardQuarter returns the MIDI quarter written as a quarter: bb = 8.
func StandardQuarter() MIDIQuarter { return MIDIQuarter{bb: 8} }

// IsSet reports whether q is set.
func (q MIDIQuarter) IsSet() bool { return q.bb != 0 }

// ThirtySeconds returns bb, the notated thirty-second notes in the MIDI quarter.
func (q MIDIQuarter) ThirtySeconds() (int, error) {
	if !q.IsSet() {
		return 0, unsetError("MIDIQuarter.ThirtySeconds", "receiver")
	}
	return int(q.bb), nil
}

// MustThirtySeconds is like ThirtySeconds but panics with its error.
func (q MIDIQuarter) MustThirtySeconds() int { return must(q.ThirtySeconds()) }

// Duration returns the written length bb/32 of the MIDI quarter.
func (q MIDIQuarter) Duration() (duration.Duration, error) {
	if !q.IsSet() {
		return duration.Duration{}, unsetError("MIDIQuarter.Duration", "receiver")
	}
	return duration.MustNew(int(q.bb), 32), nil
}

// MustDuration is like Duration but panics with its error.
func (q MIDIQuarter) MustDuration() duration.Duration { return must(q.Duration()) }

// IsStandard reports whether the MIDI quarter is written as a quarter (bb = 8).
func (q MIDIQuarter) IsStandard() (bool, error) {
	if !q.IsSet() {
		return false, unsetError("MIDIQuarter.IsStandard", "receiver")
	}
	return q.bb == 8, nil
}

// MustIsStandard is like IsStandard but panics with its error.
func (q MIDIQuarter) MustIsStandard() bool { return must(q.IsStandard()) }

// Equal reports whether q and other are the same notation.
func (q MIDIQuarter) Equal(other MIDIQuarter) bool { return q == other }

// String returns the written length as a reduced fraction, such as "1/4", or
// "MIDIQuarter(unset)".
func (q MIDIQuarter) String() string {
	if !q.IsSet() {
		return "MIDIQuarter(unset)"
	}
	return big.NewRat(int64(q.bb), 32).String()
}

// MarshalText implements encoding.TextMarshaler.
func (q MIDIQuarter) MarshalText() ([]byte, error) {
	if !q.IsSet() {
		return nil, unsetError("MIDIQuarter.MarshalText", "receiver")
	}
	return []byte(q.String()), nil
}

// AppendText implements encoding.TextAppender.
func (q MIDIQuarter) AppendText(dst []byte) ([]byte, error) {
	if !q.IsSet() {
		return dst, unsetError("MIDIQuarter.AppendText", "receiver")
	}
	return append(dst, q.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves q unchanged on
// error.
func (q *MIDIQuarter) UnmarshalText(text []byte) error {
	if q == nil {
		return invalidError("MIDIQuarter.UnmarshalText", "nil receiver")
	}
	v, err := ParseMIDIQuarter(string(text))
	if err != nil {
		return err
	}
	*q = v
	return nil
}

// ParseMIDIQuarter parses the text of String, a fraction n/d of a whole note.
// Unreduced fractions are accepted.
func ParseMIDIQuarter(text string) (MIDIQuarter, error) {
	num, den, ok := scanRatio(text)
	if !ok {
		return MIDIQuarter{}, grammarError(text)
	}
	value, err := ratioValue(text, num, den)
	if err != nil {
		return MIDIQuarter{}, err
	}
	q, err := quarterOfRat(value)
	if err != nil {
		return MIDIQuarter{}, valueTextError(text, err)
	}
	return q, nil
}

// MustParseMIDIQuarter is like ParseMIDIQuarter but panics with its error.
func MustParseMIDIQuarter(text string) MIDIQuarter { return must(ParseMIDIQuarter(text)) }

func quarterOfRat(value *big.Rat) (MIDIQuarter, error) {
	const op = "ParseMIDIQuarter"
	if value.Sign() == 0 {
		return MIDIQuarter{}, newError(ErrZero, op, "zero MIDI quarter")
	}
	n := new(big.Rat).Mul(value, big.NewRat(32, 1))
	if !n.IsInt() {
		return MIDIQuarter{}, newError(ErrNotRepresentable, op, "not a whole number of thirty-second notes")
	}
	if n.Num().Cmp(big.NewInt(MaxByteField)) > 0 {
		return MIDIQuarter{}, newError(ErrFormatLimit, op, "thirty-seconds above 255")
	}
	return MIDIQuarter{bb: uint8(n.Num().Int64())}, nil
}

// rat returns w = bb/32.
func (q MIDIQuarter) rat() *big.Rat { return big.NewRat(int64(q.bb), 32) }

// quarterArgument resolves the variadic MIDI quarter argument: none is the
// standard quarter, one must be set, more is ErrInvalidArgument.
func quarterArgument(op string, q []MIDIQuarter) (MIDIQuarter, error) {
	switch len(q) {
	case 0:
		return StandardQuarter(), nil
	case 1:
		if !q[0].IsSet() {
			return MIDIQuarter{}, unsetError(op, "MIDI quarter")
		}
		return q[0], nil
	default:
		return MIDIQuarter{}, invalidError(op, strconv.Itoa(len(q))+" MIDI quarter arguments")
	}
}
