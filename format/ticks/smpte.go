package ticks

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/tempo"
)

// FrameRate is the SMPTE format of an SMF division word. The zero value and
// values outside the constants are invalid.
type FrameRate uint8

const (
	FrameRate24 FrameRate = iota + 1 // code -24
	FrameRate25                      // code -25
	FrameRate29                      // code -29; seconds need a Reading
	FrameRate30                      // code -30
)

var frameRateCodes = [...]int{0, -24, -25, -29, -30}

func frameRateOfCode(code int) (FrameRate, bool) {
	for i := FrameRate24; i <= FrameRate30; i++ {
		if frameRateCodes[i] == code {
			return i, true
		}
	}
	return 0, false
}

// FrameRateOf returns the frame rate of an SMPTE code: -24, -25, -29 or -30.
// Other codes are ErrInvalidArgument.
func FrameRateOf(code int) (FrameRate, error) {
	rate, ok := frameRateOfCode(code)
	if !ok {
		return 0, invalidError("FrameRateOf", "frame code "+strconv.Itoa(code))
	}
	return rate, nil
}

// MustFrameRateOf is like FrameRateOf but panics with its error.
func MustFrameRateOf(code int) FrameRate { return must(FrameRateOf(code)) }

// IsValid reports whether f is one of the four frame rates.
func (f FrameRate) IsValid() bool { return f >= FrameRate24 && f <= FrameRate30 }

// Equal reports whether f and other are the same value.
func (f FrameRate) Equal(other FrameRate) bool { return f == other }

// Code returns the negative SMPTE code of f.
func (f FrameRate) Code() (int, error) {
	if !f.IsValid() {
		return 0, invalidError("FrameRate.Code", "invalid frame rate")
	}
	return frameRateCodes[f], nil
}

// MustCode is like Code but panics with its error.
func (f FrameRate) MustCode() int { return must(f.Code()) }

// String returns the code, such as "-25", or "FrameRate(invalid)".
func (f FrameRate) String() string {
	if !f.IsValid() {
		return "FrameRate(invalid)"
	}
	return strconv.Itoa(frameRateCodes[f])
}

// ParseFrameRate parses the text of String.
func ParseFrameRate(text string) (FrameRate, error) {
	for f := FrameRate24; f <= FrameRate30; f++ {
		if f.String() == text {
			return f, nil
		}
	}
	return 0, grammarError(text)
}

// MustParseFrameRate is like ParseFrameRate but panics with its error.
func MustParseFrameRate(text string) FrameRate { return must(ParseFrameRate(text)) }

// Reading selects how code -29 is read as frames per second. The zero value
// and values outside the constants are invalid.
type Reading uint8

const (
	ReadingDropFrame Reading = iota + 1 // 30000/1001 (29.97 drop-frame)
	Reading30                           // 30
	Reading29                           // 29
)

var readingNames = [...]string{"", "df", "30", "29"}

// IsValid reports whether r is one of the three readings.
func (r Reading) IsValid() bool { return r >= ReadingDropFrame && r <= Reading29 }

// Equal reports whether r and other are the same value.
func (r Reading) Equal(other Reading) bool { return r == other }

// String returns "df", "30" or "29", or "Reading(invalid)".
func (r Reading) String() string {
	if !r.IsValid() {
		return "Reading(invalid)"
	}
	return readingNames[r]
}

// ParseReading parses the text of String.
func ParseReading(text string) (Reading, error) {
	for r := ReadingDropFrame; r <= Reading29; r++ {
		if readingNames[r] == text {
			return r, nil
		}
	}
	return 0, grammarError(text)
}

// MustParseReading is like ParseReading but panics with its error.
func MustParseReading(text string) Reading { return must(ParseReading(text)) }

func (r Reading) framesPerSecond() *big.Rat {
	switch r {
	case ReadingDropFrame:
		return big.NewRat(30000, 1001)
	case Reading30:
		return big.NewRat(30, 1)
	default:
		return big.NewRat(29, 1)
	}
}

// SMPTE is an SMPTE resolution: a frame rate and ticks per frame; for code -29
// an optional reading of the frames per second. A tick is a fraction of a
// second and does not depend on the tempo. The zero value is unset. SMPTE is
// comparable: a == b reports the same as a.Equal(b).
type SMPTE struct {
	rate    FrameRate
	tpf     uint8
	reading Reading
}

// NewSMPTE returns the SMPTE resolution of rate with ticksPerFrame ticks per
// frame. An invalid rate is ErrInvalidArgument; zero ticks per frame ErrZero,
// a negative value ErrInvalidArgument and a value above 255 ErrFormatLimit.
func NewSMPTE(rate FrameRate, ticksPerFrame int) (SMPTE, error) {
	const op = "NewSMPTE"
	if !rate.IsValid() {
		return SMPTE{}, invalidError(op, "invalid frame rate")
	}
	switch {
	case ticksPerFrame < 0:
		return SMPTE{}, invalidError(op, "negative ticks per frame")
	case ticksPerFrame == 0:
		return SMPTE{}, newError(ErrZero, op, "zero ticks per frame")
	case ticksPerFrame > MaxByteField:
		return SMPTE{}, newError(ErrFormatLimit, op, "ticks per frame above 255")
	}
	return SMPTE{rate: rate, tpf: uint8(ticksPerFrame)}, nil
}

// MustNewSMPTE is like NewSMPTE but panics with its error.
func MustNewSMPTE(rate FrameRate, ticksPerFrame int) SMPTE {
	return must(NewSMPTE(rate, ticksPerFrame))
}

// IsSet reports whether s is set.
func (s SMPTE) IsSet() bool { return s.rate != 0 }

func (s SMPTE) check(op string) error {
	if !s.IsSet() {
		return unsetError(op, "receiver")
	}
	return nil
}

// Rate returns the frame rate.
func (s SMPTE) Rate() (FrameRate, error) {
	if err := s.check("SMPTE.Rate"); err != nil {
		return 0, err
	}
	return s.rate, nil
}

// MustRate is like Rate but panics with its error.
func (s SMPTE) MustRate() FrameRate { return must(s.Rate()) }

// TicksPerFrame returns the ticks per frame.
func (s SMPTE) TicksPerFrame() (int, error) {
	if err := s.check("SMPTE.TicksPerFrame"); err != nil {
		return 0, err
	}
	return int(s.tpf), nil
}

// MustTicksPerFrame is like TicksPerFrame but panics with its error.
func (s SMPTE) MustTicksPerFrame() int { return must(s.TicksPerFrame()) }

// Reading returns the reading of code -29 and true if one is selected.
func (s SMPTE) Reading() (Reading, bool, error) {
	if err := s.check("SMPTE.Reading"); err != nil {
		return 0, false, err
	}
	return s.reading, s.reading != 0, nil
}

// MustReading is like Reading but panics with its error.
func (s SMPTE) MustReading() (Reading, bool) { return must2(s.Reading()) }

// WithReading returns s with the reading r of code -29. Other frame rates and
// invalid readings are ErrInvalidArgument.
func (s SMPTE) WithReading(r Reading) (SMPTE, error) {
	const op = "SMPTE.WithReading"
	if err := s.check(op); err != nil {
		return SMPTE{}, err
	}
	if !r.IsValid() {
		return SMPTE{}, invalidError(op, "invalid reading")
	}
	if s.rate != FrameRate29 {
		return SMPTE{}, invalidError(op, "reading of a frame rate other than -29")
	}
	s.reading = r
	return s, nil
}

// MustWithReading is like WithReading but panics with its error.
func (s SMPTE) MustWithReading(r Reading) SMPTE { return must(s.WithReading(r)) }

// framesPerSecond returns the exact frames per second; code -29 without a
// reading is ErrInvalidArgument.
func (s SMPTE) framesPerSecond(op string) (*big.Rat, error) {
	switch s.rate {
	case FrameRate24:
		return big.NewRat(24, 1), nil
	case FrameRate25:
		return big.NewRat(25, 1), nil
	case FrameRate30:
		return big.NewRat(30, 1), nil
	}
	if s.reading == 0 {
		return nil, invalidError(op, "code -29 needs a reading")
	}
	return s.reading.framesPerSecond(), nil
}

// ticksPerSecond returns the exact ticks per second.
func (s SMPTE) ticksPerSecond(op string) (*big.Rat, error) {
	fps, err := s.framesPerSecond(op)
	if err != nil {
		return nil, err
	}
	return fps.Mul(fps, big.NewRat(int64(s.tpf), 1)), nil
}

// FramesPerSecond returns the exact frames per second: 24, 25, 30, or for code
// -29 30000/1001, 30 or 29 by the reading. Code -29 without a reading is
// ErrInvalidArgument.
func (s SMPTE) FramesPerSecond() (fraction.Fraction, error) {
	const op = "SMPTE.FramesPerSecond"
	if err := s.check(op); err != nil {
		return fraction.Fraction{}, err
	}
	fps, err := s.framesPerSecond(op)
	if err != nil {
		return fraction.Fraction{}, err
	}
	return fractionOf(op, fps)
}

// MustFramesPerSecond is like FramesPerSecond but panics with its error.
func (s SMPTE) MustFramesPerSecond() fraction.Fraction { return must(s.FramesPerSecond()) }

// TicksPerSecond returns the exact ticks per second: frames per second times
// ticks per frame.
func (s SMPTE) TicksPerSecond() (fraction.Fraction, error) {
	const op = "SMPTE.TicksPerSecond"
	if err := s.check(op); err != nil {
		return fraction.Fraction{}, err
	}
	tps, err := s.ticksPerSecond(op)
	if err != nil {
		return fraction.Fraction{}, err
	}
	return fractionOf(op, tps)
}

// MustTicksPerSecond is like TicksPerSecond but panics with its error.
func (s SMPTE) MustTicksPerSecond() fraction.Fraction { return must(s.TicksPerSecond()) }

// Word returns the SMF division word of s: bit 15 set, the code in the high
// byte in two's complement and the ticks per frame in the low byte.
func (s SMPTE) Word() (uint16, error) {
	if err := s.check("SMPTE.Word"); err != nil {
		return 0, err
	}
	high := uint16(uint8(int8(frameRateCodes[s.rate])))
	return high<<8 | uint16(s.tpf), nil
}

// MustWord is like Word but panics with its error.
func (s SMPTE) MustWord() uint16 { return must(s.Word()) }

// Seconds returns the exact seconds of ticks >= 0 SMPTE ticks.
func (s SMPTE) Seconds(ticks int) (tempo.Seconds, error) {
	const op = "SMPTE.Seconds"
	if err := s.check(op); err != nil {
		return tempo.Seconds{}, err
	}
	if ticks < 0 {
		return tempo.Seconds{}, invalidError(op, "negative ticks")
	}
	tps, err := s.ticksPerSecond(op)
	if err != nil {
		return tempo.Seconds{}, err
	}
	return secondsOf(op, new(big.Rat).Quo(ratInt(ticks), tps))
}

// MustSeconds is like Seconds but panics with its error.
func (s SMPTE) MustSeconds(ticks int) tempo.Seconds { return must(s.Seconds(ticks)) }

func secondsArgument(op string, seconds tempo.Seconds) error {
	if !seconds.IsSet() {
		return invalidError(op, "seconds", tempo.ErrUnset)
	}
	return nil
}

// TicksIn returns the exact SMPTE ticks of seconds.
func (s SMPTE) TicksIn(seconds tempo.Seconds) (fraction.Fraction, error) {
	const op = "SMPTE.TicksIn"
	if err := s.check(op); err != nil {
		return fraction.Fraction{}, err
	}
	if err := secondsArgument(op, seconds); err != nil {
		return fraction.Fraction{}, err
	}
	tps, err := s.ticksPerSecond(op)
	if err != nil {
		return fraction.Fraction{}, err
	}
	return fractionOf(op, tps.Mul(tps, ratOfSeconds(seconds)))
}

// MustTicksIn is like TicksIn but panics with its error.
func (s SMPTE) MustTicksIn(seconds tempo.Seconds) fraction.Fraction { return must(s.TicksIn(seconds)) }

// TickAt returns the SMPTE tick of seconds rounded once with mode.
func (s SMPTE) TickAt(seconds tempo.Seconds, mode rounding.Mode) (int, error) {
	const op = "SMPTE.TickAt"
	if err := s.check(op); err != nil {
		return 0, err
	}
	if err := secondsArgument(op, seconds); err != nil {
		return 0, err
	}
	if err := checkMode(op, mode); err != nil {
		return 0, err
	}
	tps, err := s.ticksPerSecond(op)
	if err != nil {
		return 0, err
	}
	n, err := roundInt(op, tps.Mul(tps, ratOfSeconds(seconds)), mode)
	if err != nil {
		return 0, err
	}
	return intOf(op, n)
}

// MustTickAt is like TickAt but panics with its error.
func (s SMPTE) MustTickAt(seconds tempo.Seconds, mode rounding.Mode) int {
	return must(s.TickAt(seconds, mode))
}

// Equal reports whether s and other have the same frame rate, ticks per frame
// and reading.
func (s SMPTE) Equal(other SMPTE) bool { return s == other }

// String returns the code, "x" and the ticks per frame, and ":" with the
// reading if one is selected, such as "-25x40" or "-29x80:df", or
// "SMPTE(unset)".
func (s SMPTE) String() string {
	if !s.IsSet() {
		return "SMPTE(unset)"
	}
	text := s.rate.String() + "x" + strconv.Itoa(int(s.tpf))
	if s.reading != 0 {
		text += ":" + s.reading.String()
	}
	return text
}

// MarshalText implements encoding.TextMarshaler.
func (s SMPTE) MarshalText() ([]byte, error) {
	if err := s.check("SMPTE.MarshalText"); err != nil {
		return nil, err
	}
	return []byte(s.String()), nil
}

// AppendText implements encoding.TextAppender.
func (s SMPTE) AppendText(dst []byte) ([]byte, error) {
	if err := s.check("SMPTE.AppendText"); err != nil {
		return dst, err
	}
	return append(dst, s.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves s unchanged on
// error.
func (s *SMPTE) UnmarshalText(text []byte) error {
	if s == nil {
		return invalidError("SMPTE.UnmarshalText", "nil receiver")
	}
	v, err := ParseSMPTE(string(text))
	if err != nil {
		return err
	}
	*s = v
	return nil
}

// ParseSMPTE parses the text of String.
func ParseSMPTE(text string) (SMPTE, error) {
	codeText, rest, ok := strings.Cut(text, "x")
	if !ok {
		return SMPTE{}, grammarError(text)
	}
	rate, err := ParseFrameRate(codeText)
	if err != nil {
		return SMPTE{}, grammarError(text)
	}
	tpfText, readingText, hasReading := strings.Cut(rest, ":")
	if !isUint(tpfText) {
		return SMPTE{}, grammarError(text)
	}
	var reading Reading
	if hasReading {
		if reading, err = ParseReading(readingText); err != nil {
			return SMPTE{}, grammarError(text)
		}
	}
	tpf, err := uintValue(text, tpfText)
	if err != nil {
		return SMPTE{}, err
	}
	s, err := NewSMPTE(rate, tpf)
	if err == nil && hasReading {
		s, err = s.WithReading(reading)
	}
	if err != nil {
		return SMPTE{}, valueTextError(text, err)
	}
	return s, nil
}

// MustParseSMPTE is like ParseSMPTE but panics with its error.
func MustParseSMPTE(text string) SMPTE { return must(ParseSMPTE(text)) }
