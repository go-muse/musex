package ticks

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
)

// Fidelity selects whether a meter record that MIDI cannot hold is an error.
// The zero value and values outside the constants are invalid.
type Fidelity uint8

const (
	Strict Fidelity = iota + 1 // symbols, additive numerators, sums and explicit grouping are errors
	Lossy                      // they are written by their size; the loss is reported
)

var fidelityNames = [...]string{"", "strict", "lossy"}

// IsValid reports whether f is Strict or Lossy.
func (f Fidelity) IsValid() bool { return f == Strict || f == Lossy }

// Equal reports whether f and other are the same value.
func (f Fidelity) Equal(other Fidelity) bool { return f == other }

// String returns "strict" or "lossy", or "Fidelity(invalid)".
func (f Fidelity) String() string {
	if !f.IsValid() {
		return "Fidelity(invalid)"
	}
	return fidelityNames[f]
}

// ParseFidelity parses the text of String.
func ParseFidelity(text string) (Fidelity, error) {
	for f := Strict; f <= Lossy; f++ {
		if fidelityNames[f] == text {
			return f, nil
		}
	}
	return 0, grammarError(text)
}

// MustParseFidelity is like ParseFidelity but panics with its error.
func MustParseFidelity(text string) Fidelity { return must(ParseFidelity(text)) }

// Loss names a part of a meter record lost in MIDI. The zero value and values
// outside the constants are invalid.
type Loss uint8

const (
	LostSymbol   Loss = iota + 1 // C or ¢
	LostGrouping                 // additive numerator or explicit grouping
	LostParts                    // a sum of time signatures
)

var lossNames = [...]string{"", "symbol", "grouping", "parts"}

// IsValid reports whether l is one of the three losses.
func (l Loss) IsValid() bool { return l >= LostSymbol && l <= LostParts }

// Equal reports whether l and other are the same value.
func (l Loss) Equal(other Loss) bool { return l == other }

// String returns "symbol", "grouping" or "parts", or "Loss(invalid)".
func (l Loss) String() string {
	if !l.IsValid() {
		return "Loss(invalid)"
	}
	return lossNames[l]
}

// ParseLoss parses the text of String.
func ParseLoss(text string) (Loss, error) {
	for l := LostSymbol; l <= LostParts; l++ {
		if lossNames[l] == text {
			return l, nil
		}
	}
	return 0, grammarError(text)
}

// MustParseLoss is like ParseLoss but panics with its error.
func MustParseLoss(text string) Loss { return must(ParseLoss(text)) }

// Click is the metronome click of a MIDI time signature: an explicit duration
// or the beat of the meter. The zero value is unset. Click is comparable:
// a == b reports the same as a.Equal(b).
type Click struct {
	every  duration.Duration
	onBeat bool
}

// ClickEvery returns a click every duration d; an unset d is
// ErrInvalidArgument.
func ClickEvery(d duration.Duration) (Click, error) {
	if !d.IsSet() {
		return Click{}, invalidError("ClickEvery", "click duration", duration.ErrUnset)
	}
	return Click{every: d}, nil
}

// MustClickEvery is like ClickEvery but panics with its error.
func MustClickEvery(d duration.Duration) Click { return must(ClickEvery(d)) }

// ClickOnBeat returns a click on every beat of the meter.
func ClickOnBeat() Click { return Click{onBeat: true} }

// IsSet reports whether c is set.
func (c Click) IsSet() bool { return c.onBeat || c.every.IsSet() }

// OnBeat reports whether the click is on the beat.
func (c Click) OnBeat() (bool, error) {
	if !c.IsSet() {
		return false, unsetError("Click.OnBeat", "receiver")
	}
	return c.onBeat, nil
}

// MustOnBeat is like OnBeat but panics with its error.
func (c Click) MustOnBeat() bool { return must(c.OnBeat()) }

// Every returns the explicit click duration and true, or false for a click on
// the beat.
func (c Click) Every() (duration.Duration, bool, error) {
	if !c.IsSet() {
		return duration.Duration{}, false, unsetError("Click.Every", "receiver")
	}
	return c.every, !c.onBeat, nil
}

// MustEvery is like Every but panics with its error.
func (c Click) MustEvery() (duration.Duration, bool) { return must2(c.Every()) }

// Equal reports whether c and other are the same click.
func (c Click) Equal(other Click) bool { return c == other }

// String returns "beat" or the click duration as a fraction, such as "3/8", or
// "Click(unset)".
func (c Click) String() string {
	switch {
	case !c.IsSet():
		return "Click(unset)"
	case c.onBeat:
		return "beat"
	}
	return ratText(ratOfDuration(c.every))
}

// MarshalText implements encoding.TextMarshaler.
func (c Click) MarshalText() ([]byte, error) {
	if !c.IsSet() {
		return nil, unsetError("Click.MarshalText", "receiver")
	}
	return []byte(c.String()), nil
}

// AppendText implements encoding.TextAppender.
func (c Click) AppendText(dst []byte) ([]byte, error) {
	if !c.IsSet() {
		return dst, unsetError("Click.AppendText", "receiver")
	}
	return append(dst, c.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves c unchanged on
// error.
func (c *Click) UnmarshalText(text []byte) error {
	if c == nil {
		return invalidError("Click.UnmarshalText", "nil receiver")
	}
	v, err := ParseClick(string(text))
	if err != nil {
		return err
	}
	*c = v
	return nil
}

// ParseClick parses the text of String; a zero duration is ErrInvalidArgument.
func ParseClick(text string) (Click, error) {
	if text == "beat" {
		return ClickOnBeat(), nil
	}
	num, den, ok := scanRatio(text)
	if !ok {
		return Click{}, grammarError(text)
	}
	value, err := ratioValue(text, num, den)
	if err != nil {
		return Click{}, err
	}
	if value.Sign() == 0 {
		return Click{}, valueTextError(text, invalidError("ParseClick", "zero click"))
	}
	d, err := durationOf("ParseClick", value)
	if err != nil {
		return Click{}, valueTextError(text, err)
	}
	return Click{every: d}, nil
}

// MustParseClick is like ParseClick but panics with its error.
func MustParseClick(text string) Click { return must(ParseClick(text)) }

// MIDIMeter is the four fields of an SMF Time Signature event: the numerator
// nn, the exponent dd of the denominator 2^dd, the MIDI clocks per click cc and
// bb, the notated thirty-second notes in a MIDI quarter. The zero value is
// unset. MIDIMeter is comparable: a == b reports the same as a.Equal(b).
type MIDIMeter struct {
	nn, dd, cc, bb uint8
}

// checkByteField checks nn, cc or bb.
func checkByteField(op, name string, v int) error {
	switch {
	case v < 0:
		return invalidError(op, "negative "+name)
	case v == 0:
		return newError(ErrZero, op, "zero "+name)
	case v > MaxByteField:
		return newError(ErrFormatLimit, op, name+" above 255")
	}
	return nil
}

// NewMIDIMeter returns the time signature event of the four fields, checked in
// order: a negative field is ErrInvalidArgument; zero nn, cc or bb ErrZero;
// nn, cc or bb above 255 ErrFormatLimit; dd = 255 ErrOutOfRange (2^255 is
// outside common/fraction) and dd above 255 ErrFormatLimit.
func NewMIDIMeter(numerator, denominatorExponent, clocksPerClick, thirtySeconds int) (MIDIMeter, error) {
	const op = "NewMIDIMeter"
	if err := checkByteField(op, "numerator", numerator); err != nil {
		return MIDIMeter{}, err
	}
	switch {
	case denominatorExponent < 0:
		return MIDIMeter{}, invalidError(op, "negative denominator exponent")
	case denominatorExponent > MaxByteField:
		return MIDIMeter{}, newError(ErrFormatLimit, op, "denominator exponent above 255")
	case denominatorExponent > MaxDenominatorExponent:
		return MIDIMeter{}, rangeError(op, "denominator 2^255 outside common/fraction")
	}
	if err := checkByteField(op, "clocks per click", clocksPerClick); err != nil {
		return MIDIMeter{}, err
	}
	if err := checkByteField(op, "thirty-seconds", thirtySeconds); err != nil {
		return MIDIMeter{}, err
	}
	return MIDIMeter{nn: uint8(numerator), dd: uint8(denominatorExponent), cc: uint8(clocksPerClick), bb: uint8(thirtySeconds)}, nil
}

// MustNewMIDIMeter is like NewMIDIMeter but panics with its error.
func MustNewMIDIMeter(numerator, denominatorExponent, clocksPerClick, thirtySeconds int) MIDIMeter {
	return must(NewMIDIMeter(numerator, denominatorExponent, clocksPerClick, thirtySeconds))
}

// DefaultMIDIMeter returns 4, 2, 24, 8: 4/4 with a click every MIDI quarter
// and the standard quarter. SMF defines only 4/4 as the default; cc and bb are
// a product choice.
func DefaultMIDIMeter() MIDIMeter { return MIDIMeter{nn: 4, dd: 2, cc: 24, bb: 8} }

// defaultMeterOf returns the default event 4, 2, 24, bb.
func defaultMeterOf(bb uint8) MIDIMeter { return MIDIMeter{nn: 4, dd: 2, cc: 24, bb: bb} }

// IsSet reports whether t is set.
func (t MIDIMeter) IsSet() bool { return t.nn != 0 }

func (t MIDIMeter) check(op string) error {
	if !t.IsSet() {
		return unsetError(op, "receiver")
	}
	return nil
}

// Numerator returns nn.
func (t MIDIMeter) Numerator() (int, error) {
	if err := t.check("MIDIMeter.Numerator"); err != nil {
		return 0, err
	}
	return int(t.nn), nil
}

// MustNumerator is like Numerator but panics with its error.
func (t MIDIMeter) MustNumerator() int { return must(t.Numerator()) }

// DenominatorExponent returns dd.
func (t MIDIMeter) DenominatorExponent() (int, error) {
	if err := t.check("MIDIMeter.DenominatorExponent"); err != nil {
		return 0, err
	}
	return int(t.dd), nil
}

// MustDenominatorExponent is like DenominatorExponent but panics with its
// error.
func (t MIDIMeter) MustDenominatorExponent() int { return must(t.DenominatorExponent()) }

// ClocksPerClick returns cc.
func (t MIDIMeter) ClocksPerClick() (int, error) {
	if err := t.check("MIDIMeter.ClocksPerClick"); err != nil {
		return 0, err
	}
	return int(t.cc), nil
}

// MustClocksPerClick is like ClocksPerClick but panics with its error.
func (t MIDIMeter) MustClocksPerClick() int { return must(t.ClocksPerClick()) }

// ThirtySeconds returns bb.
func (t MIDIMeter) ThirtySeconds() (int, error) {
	if err := t.check("MIDIMeter.ThirtySeconds"); err != nil {
		return 0, err
	}
	return int(t.bb), nil
}

// MustThirtySeconds is like ThirtySeconds but panics with its error.
func (t MIDIMeter) MustThirtySeconds() int { return must(t.ThirtySeconds()) }

// Quarter returns the MIDI quarter written as bb thirty-second notes.
func (t MIDIMeter) Quarter() (MIDIQuarter, error) {
	if err := t.check("MIDIMeter.Quarter"); err != nil {
		return MIDIQuarter{}, err
	}
	return MIDIQuarter{bb: t.bb}, nil
}

// MustQuarter is like Quarter but panics with its error.
func (t MIDIMeter) MustQuarter() MIDIQuarter { return must(t.Quarter()) }

func (t MIDIMeter) denominator() *big.Rat {
	return new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), uint(t.dd)))
}

// Denominator returns the note value 1/2^dd of the denominator.
func (t MIDIMeter) Denominator() (duration.Duration, error) {
	const op = "MIDIMeter.Denominator"
	if err := t.check(op); err != nil {
		return duration.Duration{}, err
	}
	return durationOf(op, t.denominator())
}

// MustDenominator is like Denominator but panics with its error.
func (t MIDIMeter) MustDenominator() duration.Duration { return must(t.Denominator()) }

// Size returns the bar size nn/2^dd.
func (t MIDIMeter) Size() (duration.Duration, error) {
	const op = "MIDIMeter.Size"
	if err := t.check(op); err != nil {
		return duration.Duration{}, err
	}
	return durationOf(op, new(big.Rat).Mul(t.denominator(), big.NewRat(int64(t.nn), 1)))
}

// MustSize is like Size but panics with its error.
func (t MIDIMeter) MustSize() duration.Duration { return must(t.Size()) }

// Click returns the click duration cc*w/24 = cc*bb/768.
func (t MIDIMeter) Click() (duration.Duration, error) {
	const op = "MIDIMeter.Click"
	if err := t.check(op); err != nil {
		return duration.Duration{}, err
	}
	return durationOf(op, big.NewRat(int64(t.cc)*int64(t.bb), 768))
}

// MustClick is like Click but panics with its error.
func (t MIDIMeter) MustClick() duration.Duration { return must(t.Click()) }

// Meter returns the simple time signature nn/2^dd with the standard grouping.
// dd above 62 is ErrOutOfRange: the denominator does not fit int.
func (t MIDIMeter) Meter() (meter.Meter, error) {
	const op = "MIDIMeter.Meter"
	if err := t.check(op); err != nil {
		return meter.Meter{}, err
	}
	if t.dd > 62 {
		return meter.Meter{}, intRangeError(op, "denominator outside int")
	}
	m, err := meter.New(int(t.nn), 1<<t.dd)
	if err != nil {
		return meter.Meter{}, dependencyError(op, err)
	}
	return m, nil
}

// MustMeter is like Meter but panics with its error.
func (t MIDIMeter) MustMeter() meter.Meter { return must(t.Meter()) }

// Equal reports whether t and other have equal fields.
func (t MIDIMeter) Equal(other MIDIMeter) bool { return t == other }

// String returns the four fields, such as "6,3,36,8", or "MIDIMeter(unset)".
func (t MIDIMeter) String() string {
	if !t.IsSet() {
		return "MIDIMeter(unset)"
	}
	return strconv.Itoa(int(t.nn)) + "," + strconv.Itoa(int(t.dd)) + "," +
		strconv.Itoa(int(t.cc)) + "," + strconv.Itoa(int(t.bb))
}

// MarshalText implements encoding.TextMarshaler.
func (t MIDIMeter) MarshalText() ([]byte, error) {
	if err := t.check("MIDIMeter.MarshalText"); err != nil {
		return nil, err
	}
	return []byte(t.String()), nil
}

// AppendText implements encoding.TextAppender.
func (t MIDIMeter) AppendText(dst []byte) ([]byte, error) {
	if err := t.check("MIDIMeter.AppendText"); err != nil {
		return dst, err
	}
	return append(dst, t.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves t unchanged on
// error.
func (t *MIDIMeter) UnmarshalText(text []byte) error {
	if t == nil {
		return invalidError("MIDIMeter.UnmarshalText", "nil receiver")
	}
	v, err := ParseMIDIMeter(string(text))
	if err != nil {
		return err
	}
	*t = v
	return nil
}

// scanMIDIMeter checks the grammar uint "," uint "," uint "," uint.
func scanMIDIMeter(text string) ([]string, bool) {
	fields := strings.Split(text, ",")
	if len(fields) != 4 {
		return nil, false
	}
	for _, f := range fields {
		if !isUint(f) {
			return nil, false
		}
	}
	return fields, true
}

func midiMeterNumbers(text string, fields []string) ([4]int, error) {
	var values [4]int
	for i, f := range fields {
		v, err := uintValue(text, f)
		if err != nil {
			return values, err
		}
		values[i] = v
	}
	return values, nil
}

// ParseMIDIMeter parses the text of String.
func ParseMIDIMeter(text string) (MIDIMeter, error) {
	fields, ok := scanMIDIMeter(text)
	if !ok {
		return MIDIMeter{}, grammarError(text)
	}
	v, err := midiMeterNumbers(text, fields)
	if err != nil {
		return MIDIMeter{}, err
	}
	t, err := NewMIDIMeter(v[0], v[1], v[2], v[3])
	if err != nil {
		return MIDIMeter{}, valueTextError(text, err)
	}
	return t, nil
}

// MustParseMIDIMeter is like ParseMIDIMeter but panics with its error.
func MustParseMIDIMeter(text string) MIDIMeter { return must(ParseMIDIMeter(text)) }

func notMIDI(op, detail string, causes ...error) error {
	return newError(ErrMeterNotMIDI, op, detail, causes...)
}

// exponentOf returns log2 of a power of two, or false.
func exponentOf(n *big.Int) (int, bool) {
	if n.Sign() <= 0 || new(big.Int).And(n, new(big.Int).Sub(n, big.NewInt(1))).Sign() != 0 {
		return 0, false
	}
	return n.BitLen() - 1, true
}

// clocksPerClick returns cc for the click of the meter m written with the MIDI
// quarter q.
func clocksPerClick(op string, m meter.Meter, click Click, q MIDIQuarter) (uint8, error) {
	var length *big.Rat
	if click.onBeat {
		beats, err := m.Beats()
		if err != nil {
			return 0, meterError(op, err)
		}
		length = ratOfDuration(beats[0].Duration)
		for _, b := range beats[1:] {
			if ratOfDuration(b.Duration).Cmp(length) != 0 {
				return 0, notMIDI(op, "beats of different lengths need an explicit click")
			}
		}
	} else {
		length = ratOfDuration(click.every)
	}
	cc := new(big.Rat).Mul(length, big.NewRat(ClocksPerQuarter, 1))
	cc.Quo(cc, q.rat())
	if !cc.IsInt() || cc.Sign() <= 0 || cc.Num().Cmp(big.NewInt(MaxByteField)) > 0 {
		return 0, notMIDI(op, "clocks per click not a whole number in 1..255")
	}
	return uint8(cc.Num().Int64()), nil
}

// meterSize returns nn and dd of the size of m: the simple numerator over the
// denominator, or for a sum the size over the largest denominator of the parts.
func meterSize(op string, m meter.Meter) (*big.Int, int, error) {
	parts, err := m.Parts()
	if err != nil {
		return nil, 0, meterError(op, err)
	}
	largest := 0
	for _, p := range parts {
		den := big.NewInt(int64(p.MustDenominator()))
		exp, _ := exponentOf(den)
		largest = max(largest, exp)
	}
	size := ratOfDuration(m.MustDuration())
	nn := new(big.Rat).Mul(size, new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(largest))))
	if !nn.IsInt() {
		return nil, 0, notMIDI(op, "size not a whole number of the largest denominator")
	}
	return nn.Num(), largest, nil
}

// MIDIMeterOf returns the time signature event of the meter m with the click
// and the MIDI quarter q, and the lost parts of the record. Denominators that
// are not powers of two are ErrMeterNotMIDI. With Strict a symbol, an additive
// numerator, a sum or an explicit grouping is ErrMeterNotMIDI; with Lossy they
// are written by their size and reported in the order LostSymbol,
// LostGrouping, LostParts (nil without losses). A numerator above 255, a click
// on the beat of a meter without equal beats, and clocks per click that are
// not a whole number in 1..255 are ErrMeterNotMIDI.
func MIDIMeterOf(m meter.Meter, click Click, q MIDIQuarter, f Fidelity) (MIDIMeter, []Loss, error) {
	const op = "MIDIMeterOf"
	if !m.IsSet() {
		return MIDIMeter{}, nil, invalidError(op, "meter", meter.ErrUnset)
	}
	if !click.IsSet() {
		return MIDIMeter{}, nil, unsetError(op, "click")
	}
	if !q.IsSet() {
		return MIDIMeter{}, nil, unsetError(op, "MIDI quarter")
	}
	if !f.IsValid() {
		return MIDIMeter{}, nil, invalidError(op, "invalid fidelity")
	}
	powers, err := m.HasPowerOfTwoDenominators()
	if err != nil {
		return MIDIMeter{}, nil, meterError(op, err)
	}
	if !powers {
		return MIDIMeter{}, nil, notMIDI(op, "denominator not a power of two")
	}
	var losses []Loss
	form := m.MustForm()
	explicit := m.MustHasExplicitGrouping()
	if form == meter.FormSymbol {
		losses = append(losses, LostSymbol)
	}
	if form == meter.FormAdditive || explicit {
		losses = append(losses, LostGrouping)
	}
	if form == meter.FormSum {
		losses = append(losses, LostParts)
	}
	if f == Strict && len(losses) != 0 {
		return MIDIMeter{}, nil, notMIDI(op, "the record is not a simple time signature")
	}
	nn, dd, err := meterSize(op, m)
	if err != nil {
		return MIDIMeter{}, nil, err
	}
	if nn.Cmp(big.NewInt(MaxByteField)) > 0 {
		return MIDIMeter{}, nil, notMIDI(op, "numerator above 255")
	}
	cc, err := clocksPerClick(op, m, click, q)
	if err != nil {
		return MIDIMeter{}, nil, err
	}
	return MIDIMeter{nn: uint8(nn.Int64()), dd: uint8(dd), cc: cc, bb: q.bb}, losses, nil
}

// MustMIDIMeterOf is like MIDIMeterOf but panics with its error.
func MustMIDIMeterOf(m meter.Meter, click Click, q MIDIQuarter, f Fidelity) (MIDIMeter, []Loss) {
	return must2(MIDIMeterOf(m, click, q, f))
}
