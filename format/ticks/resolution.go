package ticks

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/go-muse/musex/duration"
)

// Format limits and defaults.
const (
	MaxSMFResolution          = 32767     // ticks per quarter in an SMF division word
	MaxMusicXMLResolution     = 16383     // MusicXML divisions recommended for SMF compatibility
	MaxDelta                  = 268435455 // 0x0FFFFFFF, the largest SMF delta-time
	MaxMicrosecondsPerQuarter = 16777215  // the largest Set Tempo value
	DefaultMicroseconds       = 500000    // SMF default tempo: 120 MIDI quarters per minute
	MaxByteField              = 255       // nn, cc, bb, ticks per frame
	MaxDenominatorExponent    = 254       // dd whose 2^dd fits common/fraction
	ClocksPerQuarter          = 24        // MIDI clocks per MIDI quarter
	ClocksPerBeat             = 6         // MIDI clocks per MIDI beat (a sixteenth)
)

// Resolution is a metric resolution: D ticks per MIDI quarter and the written
// length w of that quarter, so one tick is w/D of a whole note. MusicXML
// divisions and MEI ppq use the standard quarter (w = 1/4). The zero value is
// unset. Resolution is comparable: a == b reports the same as a.Equal(b).
type Resolution struct {
	d  int
	bb uint8
}

// NewResolution returns perQuarter ticks per MIDI quarter with the standard
// quarter. Zero is ErrZero and a negative value ErrInvalidArgument; the value
// has no upper limit (see FitsSMF and WithinMusicXML).
func NewResolution(perQuarter int) (Resolution, error) {
	const op = "NewResolution"
	switch {
	case perQuarter < 0:
		return Resolution{}, invalidError(op, "negative resolution")
	case perQuarter == 0:
		return Resolution{}, newError(ErrZero, op, "zero resolution")
	}
	return Resolution{d: perQuarter, bb: 8}, nil
}

// MustNewResolution is like NewResolution but panics with its error.
func MustNewResolution(perQuarter int) Resolution { return must(NewResolution(perQuarter)) }

// IsSet reports whether r is set.
func (r Resolution) IsSet() bool { return r.d != 0 }

func (r Resolution) check(op string) error {
	if !r.IsSet() {
		return unsetError(op, "receiver")
	}
	return nil
}

// PerQuarter returns D, the ticks per MIDI quarter.
func (r Resolution) PerQuarter() (int, error) {
	if err := r.check("Resolution.PerQuarter"); err != nil {
		return 0, err
	}
	return r.d, nil
}

// MustPerQuarter is like PerQuarter but panics with its error.
func (r Resolution) MustPerQuarter() int { return must(r.PerQuarter()) }

// Quarter returns the written length of the MIDI quarter.
func (r Resolution) Quarter() (MIDIQuarter, error) {
	if err := r.check("Resolution.Quarter"); err != nil {
		return MIDIQuarter{}, err
	}
	return MIDIQuarter{bb: r.bb}, nil
}

// MustQuarter is like Quarter but panics with its error.
func (r Resolution) MustQuarter() MIDIQuarter { return must(r.Quarter()) }

// WithQuarter returns the same D with the MIDI quarter written as q.
func (r Resolution) WithQuarter(q MIDIQuarter) (Resolution, error) {
	const op = "Resolution.WithQuarter"
	if err := r.check(op); err != nil {
		return Resolution{}, err
	}
	if !q.IsSet() {
		return Resolution{}, unsetError(op, "MIDI quarter")
	}
	return Resolution{d: r.d, bb: q.bb}, nil
}

// MustWithQuarter is like WithQuarter but panics with its error.
func (r Resolution) MustWithQuarter(q MIDIQuarter) Resolution { return must(r.WithQuarter(q)) }

// quarter returns w.
func (r Resolution) quarter() *big.Rat { return big.NewRat(int64(r.bb), 32) }

// tick returns the length w/D of one tick in whole notes.
func (r Resolution) tick() *big.Rat { return new(big.Rat).Quo(r.quarter(), ratInt(r.d)) }

// TickLength returns the length w/D of one tick in whole notes.
func (r Resolution) TickLength() (duration.Duration, error) {
	const op = "Resolution.TickLength"
	if err := r.check(op); err != nil {
		return duration.Duration{}, err
	}
	return durationOf(op, r.tick())
}

// MustTickLength is like TickLength but panics with its error.
func (r Resolution) MustTickLength() duration.Duration { return must(r.TickLength()) }

// FitsSMF reports whether D fits an SMF division word: D <= 32767.
func (r Resolution) FitsSMF() (bool, error) {
	if err := r.check("Resolution.FitsSMF"); err != nil {
		return false, err
	}
	return r.d <= MaxSMFResolution, nil
}

// MustFitsSMF is like FitsSMF but panics with its error.
func (r Resolution) MustFitsSMF() bool { return must(r.FitsSMF()) }

// WithinMusicXML reports whether D is within the MusicXML recommendation:
// D <= 16383.
func (r Resolution) WithinMusicXML() (bool, error) {
	if err := r.check("Resolution.WithinMusicXML"); err != nil {
		return false, err
	}
	return r.d <= MaxMusicXMLResolution, nil
}

// MustWithinMusicXML is like WithinMusicXML but panics with its error.
func (r Resolution) MustWithinMusicXML() bool { return must(r.WithinMusicXML()) }

// Word returns the SMF division word of r: D with bit 15 clear. D above 32767
// is ErrFormatLimit. The MIDI quarter notation is not part of the word.
func (r Resolution) Word() (uint16, error) {
	const op = "Resolution.Word"
	if err := r.check(op); err != nil {
		return 0, err
	}
	if r.d > MaxSMFResolution {
		return 0, newError(ErrFormatLimit, op, "resolution above 32767")
	}
	return uint16(r.d), nil
}

// MustWord is like Word but panics with its error.
func (r Resolution) MustWord() uint16 { return must(r.Word()) }

// DecodeDivision decodes an SMF division word. With bit 15 clear it returns a
// metric resolution with the standard quarter and an unset SMPTE; with bit 15
// set it returns an unset Resolution and an SMPTE resolution without a reading
// of code -29. A zero resolution or zero ticks per frame is ErrZero and an
// unknown frame code ErrInvalidArgument.
func DecodeDivision(word uint16) (Resolution, SMPTE, error) {
	const op = "DecodeDivision"
	if word&0x8000 == 0 {
		d := int(word)
		if d == 0 {
			return Resolution{}, SMPTE{}, newError(ErrZero, op, "zero resolution")
		}
		return Resolution{d: d, bb: 8}, SMPTE{}, nil
	}
	code := int(int8(uint8(word >> 8)))
	rate, ok := frameRateOfCode(code)
	if !ok {
		return Resolution{}, SMPTE{}, invalidError(op, "frame code "+strconv.Itoa(code))
	}
	tpf := int(word & 0xFF)
	if tpf == 0 {
		return Resolution{}, SMPTE{}, newError(ErrZero, op, "zero ticks per frame")
	}
	return Resolution{}, SMPTE{rate: rate, tpf: uint8(tpf)}, nil
}

// MustDecodeDivision is like DecodeDivision but panics with its error.
func MustDecodeDivision(word uint16) (Resolution, SMPTE) { return must2(DecodeDivision(word)) }

// Equal reports whether r and other have the same D and MIDI quarter notation.
func (r Resolution) Equal(other Resolution) bool { return r == other }

// String returns D, followed by "@" and the MIDI quarter when it is not the
// standard quarter, such as "480" or "480@1/2", or "Resolution(unset)".
func (r Resolution) String() string {
	if !r.IsSet() {
		return "Resolution(unset)"
	}
	text := strconv.Itoa(r.d)
	if r.bb != 8 {
		text += "@" + MIDIQuarter{bb: r.bb}.String()
	}
	return text
}

// MarshalText implements encoding.TextMarshaler.
func (r Resolution) MarshalText() ([]byte, error) {
	if err := r.check("Resolution.MarshalText"); err != nil {
		return nil, err
	}
	return []byte(r.String()), nil
}

// AppendText implements encoding.TextAppender.
func (r Resolution) AppendText(dst []byte) ([]byte, error) {
	if err := r.check("Resolution.AppendText"); err != nil {
		return dst, err
	}
	return append(dst, r.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves r unchanged on
// error.
func (r *Resolution) UnmarshalText(text []byte) error {
	if r == nil {
		return invalidError("Resolution.UnmarshalText", "nil receiver")
	}
	v, err := ParseResolution(string(text))
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// ParseResolution parses the text of String. "@1/4" is accepted and gives the
// standard quarter.
func ParseResolution(text string) (Resolution, error) {
	raw, ok := scanResolution(text)
	if !ok {
		return Resolution{}, grammarError(text)
	}
	values, err := raw.numbers(text)
	if err != nil {
		return Resolution{}, err
	}
	r, err := values.resolution()
	if err != nil {
		return Resolution{}, valueTextError(text, err)
	}
	return r, nil
}

// MustParseResolution is like ParseResolution but panics with its error.
func MustParseResolution(text string) Resolution { return must(ParseResolution(text)) }

// rawResolution is the grammatical form of a resolution.
type rawResolution struct {
	d        string
	num, den string
}

type resolutionValues struct {
	d       int
	quarter *big.Rat
}

func scanResolution(text string) (rawResolution, bool) {
	dText, quarterText, hasQuarter := strings.Cut(text, "@")
	if !isUint(dText) {
		return rawResolution{}, false
	}
	raw := rawResolution{d: dText}
	if hasQuarter {
		num, den, ok := scanRatio(quarterText)
		if !ok {
			return rawResolution{}, false
		}
		raw.num, raw.den = num, den
	}
	return raw, true
}

func (raw rawResolution) numbers(text string) (resolutionValues, error) {
	d, err := uintValue(text, raw.d)
	if err != nil {
		return resolutionValues{}, err
	}
	values := resolutionValues{d: d}
	if raw.num != "" {
		values.quarter, err = ratioValue(text, raw.num, raw.den)
		if err != nil {
			return resolutionValues{}, err
		}
	}
	return values, nil
}

func (values resolutionValues) resolution() (Resolution, error) {
	r, err := NewResolution(values.d)
	if err != nil {
		return Resolution{}, err
	}
	if values.quarter != nil {
		q, err := quarterOfRat(values.quarter)
		if err != nil {
			return Resolution{}, err
		}
		r.bb = q.bb
	}
	return r, nil
}
