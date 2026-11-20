package ticks

import (
	"math/big"
	"strconv"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
)

// MIDITempo is a Set Tempo value: microseconds per MIDI quarter, 1 through
// 16777215. The zero value is unset. MIDITempo is comparable: a == b reports
// the same as a.Equal(b).
type MIDITempo struct {
	us int32
}

// NewMIDITempo returns the tempo of microseconds per MIDI quarter. Zero is
// ErrZero, a negative value ErrInvalidArgument and a value above 16777215
// ErrFormatLimit.
func NewMIDITempo(microseconds int) (MIDITempo, error) {
	const op = "NewMIDITempo"
	switch {
	case microseconds < 0:
		return MIDITempo{}, invalidError(op, "negative microseconds")
	case microseconds == 0:
		return MIDITempo{}, newError(ErrZero, op, "zero microseconds")
	case microseconds > MaxMicrosecondsPerQuarter:
		return MIDITempo{}, newError(ErrFormatLimit, op, "microseconds above 16777215")
	}
	return MIDITempo{us: int32(microseconds)}, nil
}

// MustNewMIDITempo is like NewMIDITempo but panics with its error.
func MustNewMIDITempo(microseconds int) MIDITempo { return must(NewMIDITempo(microseconds)) }

// DefaultMIDITempo returns 500000 microseconds per MIDI quarter: 120 MIDI
// quarters per minute, the SMF default.
func DefaultMIDITempo() MIDITempo { return MIDITempo{us: DefaultMicroseconds} }

func tempoArgument(op string, t tempo.Tempo) (*big.Rat, error) {
	if !t.IsSet() {
		return nil, invalidError(op, "tempo", tempo.ErrUnset)
	}
	speed, err := t.Speed()
	if err != nil {
		return nil, dependencyError(op, err)
	}
	return ratOfFraction(speed), nil
}

// microsecondsOfSpeed returns 60*10^6*w/s for a speed s in whole notes per
// minute.
func microsecondsOfSpeed(speed, w *big.Rat) *big.Rat {
	value := new(big.Rat).Mul(big.NewRat(60000000, 1), w)
	return value.Quo(value, speed)
}

// MicrosecondsPerQuarter returns the exact microseconds per MIDI quarter of the
// marking t: 60*10^6*w/s for the speed s in whole notes per minute and the
// MIDI quarter w from q (the standard quarter by default).
func MicrosecondsPerQuarter(t tempo.Tempo, q ...MIDIQuarter) (fraction.Fraction, error) {
	const op = "MicrosecondsPerQuarter"
	speed, err := tempoArgument(op, t)
	if err != nil {
		return fraction.Fraction{}, err
	}
	quarter, err := quarterArgument(op, q)
	if err != nil {
		return fraction.Fraction{}, err
	}
	return fractionOf(op, microsecondsOfSpeed(speed, quarter.rat()))
}

// MustMicrosecondsPerQuarter is like MicrosecondsPerQuarter but panics with its
// error.
func MustMicrosecondsPerQuarter(t tempo.Tempo, q ...MIDIQuarter) fraction.Fraction {
	return must(MicrosecondsPerQuarter(t, q...))
}

// midiTempoOfRat rounds exact microseconds once with mode; a result outside
// 1..16777215 is ErrFormatLimit.
func midiTempoOfRat(op string, value *big.Rat, mode rounding.Mode) (MIDITempo, error) {
	n, err := roundInt(op, value, mode)
	if err != nil {
		return MIDITempo{}, err
	}
	if n.Sign() <= 0 || n.Cmp(big.NewInt(MaxMicrosecondsPerQuarter)) > 0 {
		return MIDITempo{}, newError(ErrFormatLimit, op, "tempo outside 1..16777215 microseconds")
	}
	return MIDITempo{us: int32(n.Int64())}, nil
}

// MIDITempoOf returns the Set Tempo value of the marking t: the exact
// MicrosecondsPerQuarter rounded once with mode. A result outside 1..16777215,
// including zero after rounding, is ErrFormatLimit.
func MIDITempoOf(t tempo.Tempo, mode rounding.Mode, q ...MIDIQuarter) (MIDITempo, error) {
	const op = "MIDITempoOf"
	speed, err := tempoArgument(op, t)
	if err != nil {
		return MIDITempo{}, err
	}
	if err := checkMode(op, mode); err != nil {
		return MIDITempo{}, err
	}
	quarter, err := quarterArgument(op, q)
	if err != nil {
		return MIDITempo{}, err
	}
	return midiTempoOfRat(op, microsecondsOfSpeed(speed, quarter.rat()), mode)
}

// MustMIDITempoOf is like MIDITempoOf but panics with its error.
func MustMIDITempoOf(t tempo.Tempo, mode rounding.Mode, q ...MIDIQuarter) MIDITempo {
	return must(MIDITempoOf(t, mode, q...))
}

// IsSet reports whether m is set.
func (m MIDITempo) IsSet() bool { return m.us != 0 }

func (m MIDITempo) check(op string) error {
	if !m.IsSet() {
		return unsetError(op, "receiver")
	}
	return nil
}

// Microseconds returns the microseconds per MIDI quarter.
func (m MIDITempo) Microseconds() (int, error) {
	if err := m.check("MIDITempo.Microseconds"); err != nil {
		return 0, err
	}
	return int(m.us), nil
}

// MustMicroseconds is like Microseconds but panics with its error.
func (m MIDITempo) MustMicroseconds() int { return must(m.Microseconds()) }

func unitArgument(op string, unit duration.Written) (*big.Rat, error) {
	if !unit.IsSet() {
		return nil, invalidError(op, "unit", duration.ErrUnset)
	}
	d, err := unit.Duration()
	if err != nil {
		return nil, dependencyError(op, err)
	}
	return ratOfDuration(d), nil
}

// tempoOfMarking returns the marking unit = bpm.
func tempoOfMarking(op string, unit duration.Written, bpm *big.Rat) (tempo.Tempo, error) {
	f, err := fractionOf(op, bpm)
	if err != nil {
		return tempo.Tempo{}, err
	}
	t, err := tempo.New(unit, f)
	if err != nil {
		return tempo.Tempo{}, dependencyError(op, err)
	}
	return t, nil
}

// markingOf returns the exact marking unit = N of microseconds per MIDI quarter
// with the MIDI quarter w: N = 60*10^6*w/(u*m).
func markingOf(op string, us int, unit duration.Written, u, w *big.Rat) (tempo.Tempo, error) {
	bpm := new(big.Rat).Mul(big.NewRat(60000000, 1), w)
	bpm.Quo(bpm, new(big.Rat).Mul(u, ratInt(us)))
	return tempoOfMarking(op, unit, bpm)
}

// Tempo returns the exact marking unit = N of m: N = 60*10^6*w/(u*m) for the
// length u of unit and the MIDI quarter w from q.
func (m MIDITempo) Tempo(unit duration.Written, q ...MIDIQuarter) (tempo.Tempo, error) {
	const op = "MIDITempo.Tempo"
	if err := m.check(op); err != nil {
		return tempo.Tempo{}, err
	}
	u, err := unitArgument(op, unit)
	if err != nil {
		return tempo.Tempo{}, err
	}
	quarter, err := quarterArgument(op, q)
	if err != nil {
		return tempo.Tempo{}, err
	}
	return markingOf(op, int(m.us), unit, u, quarter.rat())
}

// MustTempo is like Tempo but panics with its error.
func (m MIDITempo) MustTempo(unit duration.Written, q ...MIDIQuarter) tempo.Tempo {
	return must(m.Tempo(unit, q...))
}

func quarterWritten() duration.Written { return duration.MustNewWritten(duration.Quarter, 0) }

// QuarterTempo returns the exact marking of m with the quarter as unit.
func (m MIDITempo) QuarterTempo(q ...MIDIQuarter) (tempo.Tempo, error) {
	const op = "MIDITempo.QuarterTempo"
	if err := m.check(op); err != nil {
		return tempo.Tempo{}, err
	}
	quarter, err := quarterArgument(op, q)
	if err != nil {
		return tempo.Tempo{}, err
	}
	return markingOf(op, int(m.us), quarterWritten(), big.NewRat(1, 4), quarter.rat())
}

// MustQuarterTempo is like QuarterTempo but panics with its error.
func (m MIDITempo) MustQuarterTempo(q ...MIDIQuarter) tempo.Tempo { return must(m.QuarterTempo(q...)) }

// Seconds returns the exact seconds of ticks >= 0 ticks of r at the constant
// tempo m: ticks/D*m/10^6.
func (m MIDITempo) Seconds(r Resolution, ticks int) (tempo.Seconds, error) {
	const op = "MIDITempo.Seconds"
	if err := m.check(op); err != nil {
		return tempo.Seconds{}, err
	}
	if !r.IsSet() {
		return tempo.Seconds{}, unsetError(op, "resolution")
	}
	if ticks < 0 {
		return tempo.Seconds{}, invalidError(op, "negative ticks")
	}
	value := new(big.Rat).Mul(ratInt(ticks), ratInt(int(m.us)))
	value.Quo(value, new(big.Rat).Mul(ratInt(r.d), big.NewRat(1000000, 1)))
	return secondsOf(op, value)
}

// MustSeconds is like Seconds but panics with its error.
func (m MIDITempo) MustSeconds(r Resolution, ticks int) tempo.Seconds {
	return must(m.Seconds(r, ticks))
}

// Equal reports whether m and other are the same value.
func (m MIDITempo) Equal(other MIDITempo) bool { return m == other }

// Compare compares microseconds per MIDI quarter: -1, 0 or +1. The unset value
// precedes every set value.
func (m MIDITempo) Compare(other MIDITempo) int {
	switch {
	case m.us < other.us:
		return -1
	case m.us > other.us:
		return 1
	}
	return 0
}

// String returns the microseconds, such as "500000", or "MIDITempo(unset)".
func (m MIDITempo) String() string {
	if !m.IsSet() {
		return "MIDITempo(unset)"
	}
	return strconv.Itoa(int(m.us))
}

// MarshalText implements encoding.TextMarshaler.
func (m MIDITempo) MarshalText() ([]byte, error) {
	if err := m.check("MIDITempo.MarshalText"); err != nil {
		return nil, err
	}
	return []byte(m.String()), nil
}

// AppendText implements encoding.TextAppender.
func (m MIDITempo) AppendText(dst []byte) ([]byte, error) {
	if err := m.check("MIDITempo.AppendText"); err != nil {
		return dst, err
	}
	return append(dst, m.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves m unchanged on
// error.
func (m *MIDITempo) UnmarshalText(text []byte) error {
	if m == nil {
		return invalidError("MIDITempo.UnmarshalText", "nil receiver")
	}
	v, err := ParseMIDITempo(string(text))
	if err != nil {
		return err
	}
	*m = v
	return nil
}

// ParseMIDITempo parses the text of String.
func ParseMIDITempo(text string) (MIDITempo, error) {
	if !isUint(text) {
		return MIDITempo{}, grammarError(text)
	}
	us, err := uintValue(text, text)
	if err != nil {
		return MIDITempo{}, err
	}
	m, err := NewMIDITempo(us)
	if err != nil {
		return MIDITempo{}, valueTextError(text, err)
	}
	return m, nil
}

// MustParseMIDITempo is like ParseMIDITempo but panics with its error.
func MustParseMIDITempo(text string) MIDITempo { return must(ParseMIDITempo(text)) }

// ---- quarters per minute ----

// QuartersPerMinute returns the exact quarters per minute of the marking t:
// the MusicXML sound tempo and the MEI midi.bpm.
func QuartersPerMinute(t tempo.Tempo) (fraction.Fraction, error) {
	const op = "QuartersPerMinute"
	if !t.IsSet() {
		return fraction.Fraction{}, invalidError(op, "tempo", tempo.ErrUnset)
	}
	qpm, err := t.QuartersPerMinute()
	if err != nil {
		return fraction.Fraction{}, dependencyError(op, err)
	}
	return qpm, nil
}

// MustQuartersPerMinute is like QuartersPerMinute but panics with its error.
func MustQuartersPerMinute(t tempo.Tempo) fraction.Fraction { return must(QuartersPerMinute(t)) }

// FormatQuartersPerMinute returns the quarters per minute of t as a decimal
// with exactly places digits after the point, rounded once with mode; places
// outside 0..rounding.MaxPlaces is ErrInvalidArgument.
func FormatQuartersPerMinute(t tempo.Tempo, places int, mode rounding.Mode) (string, error) {
	const op = "FormatQuartersPerMinute"
	if !t.IsSet() {
		return "", invalidError(op, "tempo", tempo.ErrUnset)
	}
	if places < 0 || places > rounding.MaxPlaces {
		return "", invalidError(op, "decimal places outside 0.."+strconv.Itoa(rounding.MaxPlaces))
	}
	if err := checkMode(op, mode); err != nil {
		return "", err
	}
	qpm, err := t.QuartersPerMinute()
	if err != nil {
		return "", dependencyError(op, err)
	}
	text, err := mode.RatDecimal(ratOfFraction(qpm), places)
	if err != nil {
		return "", invalidError(op, "decimal", err)
	}
	return text, nil
}

// MustFormatQuartersPerMinute is like FormatQuartersPerMinute but panics with
// its error.
func MustFormatQuartersPerMinute(t tempo.Tempo, places int, mode rounding.Mode) string {
	return must(FormatQuartersPerMinute(t, places, mode))
}

// ParseQuartersPerMinute parses quarters per minute written as an integer, a
// fraction n/d or a decimal, such as "90", "600/7" or "85.71". Zero is ErrZero.
func ParseQuartersPerMinute(text string) (fraction.Fraction, error) {
	const op = "ParseQuartersPerMinute"
	var value *big.Rat
	var err error
	switch {
	case isUint(text):
		value, err = ratioValue(text, text, "1")
	case isDecimal(text):
		value, err = decimalValue(text, text)
	default:
		num, den, ok := scanRatio(text)
		if !ok {
			return fraction.Fraction{}, grammarError(text)
		}
		value, err = ratioValue(text, num, den)
	}
	if err != nil {
		return fraction.Fraction{}, err
	}
	if value.Sign() == 0 {
		return fraction.Fraction{}, valueTextError(text, newError(ErrZero, op, "zero quarters per minute"))
	}
	f, err := fraction.FromBigRat(value)
	if err != nil {
		return fraction.Fraction{}, numberRangeTextError(text)
	}
	return f, nil
}

// MustParseQuartersPerMinute is like ParseQuartersPerMinute but panics with its
// error.
func MustParseQuartersPerMinute(text string) fraction.Fraction {
	return must(ParseQuartersPerMinute(text))
}

// TempoFromQuartersPerMinute returns the exact marking unit = N of qpm quarters
// per minute: N = qpm/(4*u). Zero is ErrZero, a negative or unset value
// ErrInvalidArgument.
func TempoFromQuartersPerMinute(qpm fraction.Fraction, unit duration.Written) (tempo.Tempo, error) {
	const op = "TempoFromQuartersPerMinute"
	k, ok := qpm.BigRat()
	if !ok {
		return tempo.Tempo{}, invalidError(op, "quarters per minute", fraction.ErrInvalid)
	}
	switch k.Sign() {
	case -1:
		return tempo.Tempo{}, invalidError(op, "negative quarters per minute")
	case 0:
		return tempo.Tempo{}, newError(ErrZero, op, "zero quarters per minute")
	}
	u, err := unitArgument(op, unit)
	if err != nil {
		return tempo.Tempo{}, err
	}
	return tempoOfMarking(op, unit, k.Quo(k, new(big.Rat).Mul(big.NewRat(4, 1), u)))
}

// MustTempoFromQuartersPerMinute is like TempoFromQuartersPerMinute but panics
// with its error.
func MustTempoFromQuartersPerMinute(qpm fraction.Fraction, unit duration.Written) tempo.Tempo {
	return must(TempoFromQuartersPerMinute(qpm, unit))
}
