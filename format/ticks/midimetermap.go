package ticks

import (
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
)

// IrregularBars selects how bars whose length differs from their meter (a
// pickup, an explicit length) are exported. The zero value and values outside
// the constants are invalid.
type IrregularBars uint8

const (
	IrregularEvent IrregularBars = iota + 1 // a separate event of the bar length
	IrregularSkip                           // no separate event
)

var irregularNames = [...]string{"", "event", "skip"}

// IsValid reports whether b is IrregularEvent or IrregularSkip.
func (b IrregularBars) IsValid() bool { return b == IrregularEvent || b == IrregularSkip }

// Equal reports whether b and other are the same value.
func (b IrregularBars) Equal(other IrregularBars) bool { return b == other }

// String returns "event" or "skip", or "IrregularBars(invalid)".
func (b IrregularBars) String() string {
	if !b.IsValid() {
		return "IrregularBars(invalid)"
	}
	return irregularNames[b]
}

// ParseIrregularBars parses the text of String.
func ParseIrregularBars(text string) (IrregularBars, error) {
	for b := IrregularEvent; b <= IrregularSkip; b++ {
		if irregularNames[b] == text {
			return b, nil
		}
	}
	return 0, grammarError(text)
}

// MustParseIrregularBars is like ParseIrregularBars but panics with its error.
func MustParseIrregularBars(text string) IrregularBars { return must(ParseIrregularBars(text)) }

// MIDIMeterChange is one Time Signature event at an absolute tick. Its fields
// are checked where it is used: a negative tick is ErrInvalidArgument and an
// unset meter ErrUnset.
type MIDIMeterChange struct {
	Tick  int
	Meter MIDIMeter
}

func (c MIDIMeterChange) validate(op string) error {
	if !c.Meter.IsSet() {
		return unsetError(op, "time signature of a change")
	}
	if c.Tick < 0 {
		return invalidError(op, "negative tick of a time signature change")
	}
	return nil
}

// Equal reports whether c and other have equal fields.
func (c MIDIMeterChange) Equal(other MIDIMeterChange) bool { return c == other }

// String returns the tick, ":" and the event, such as "0:4,2,24,8". It prints
// the fields as they are, including invalid ones.
func (c MIDIMeterChange) String() string { return strconv.Itoa(c.Tick) + ":" + c.Meter.String() }

// MarshalText implements encoding.TextMarshaler; it requires valid fields.
func (c MIDIMeterChange) MarshalText() ([]byte, error) {
	if err := c.validate("MIDIMeterChange.MarshalText"); err != nil {
		return nil, err
	}
	return []byte(c.String()), nil
}

// AppendText implements encoding.TextAppender; it requires valid fields.
func (c MIDIMeterChange) AppendText(dst []byte) ([]byte, error) {
	if err := c.validate("MIDIMeterChange.AppendText"); err != nil {
		return dst, err
	}
	return append(dst, c.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves c unchanged on
// error.
func (c *MIDIMeterChange) UnmarshalText(text []byte) error {
	if c == nil {
		return invalidError("MIDIMeterChange.UnmarshalText", "nil receiver")
	}
	v, err := ParseMIDIMeterChange(string(text))
	if err != nil {
		return err
	}
	*c = v
	return nil
}

type rawMeterChange struct {
	tick   string
	fields []string
}

func scanMeterChange(text string) (rawMeterChange, bool) {
	tick, rest, ok := strings.Cut(text, ":")
	if !ok || !isUint(tick) {
		return rawMeterChange{}, false
	}
	fields, ok := scanMIDIMeter(rest)
	if !ok {
		return rawMeterChange{}, false
	}
	return rawMeterChange{tick: tick, fields: fields}, true
}

type meterChangeNumbers struct {
	tick   int
	fields [4]int
}

func (raw rawMeterChange) numbers(text string) (meterChangeNumbers, error) {
	tick, err := uintValue(text, raw.tick)
	if err != nil {
		return meterChangeNumbers{}, err
	}
	fields, err := midiMeterNumbers(text, raw.fields)
	if err != nil {
		return meterChangeNumbers{}, err
	}
	return meterChangeNumbers{tick: tick, fields: fields}, nil
}

func (n meterChangeNumbers) change() (MIDIMeterChange, error) {
	t, err := NewMIDIMeter(n.fields[0], n.fields[1], n.fields[2], n.fields[3])
	if err != nil {
		return MIDIMeterChange{}, err
	}
	return MIDIMeterChange{Tick: n.tick, Meter: t}, nil
}

// ParseMIDIMeterChange parses the text of String.
func ParseMIDIMeterChange(text string) (MIDIMeterChange, error) {
	raw, ok := scanMeterChange(text)
	if !ok {
		return MIDIMeterChange{}, grammarError(text)
	}
	n, err := raw.numbers(text)
	if err != nil {
		return MIDIMeterChange{}, err
	}
	c, err := n.change()
	if err != nil {
		return MIDIMeterChange{}, valueTextError(text, err)
	}
	return c, nil
}

// MustParseMIDIMeterChange is like ParseMIDIMeterChange but panics with its
// error.
func MustParseMIDIMeterChange(text string) MIDIMeterChange { return must(ParseMIDIMeterChange(text)) }

// BarLoss reports what a lossy export lost at a bar. Its fields are checked
// where it is used: a negative bar or an invalid loss is ErrInvalidArgument.
// BarLoss is not comparable with ==; use Equal.
type BarLoss struct {
	Bar    int
	Losses []Loss
}

func (l BarLoss) validate(op string) error {
	if l.Bar < 0 {
		return invalidError(op, "negative bar")
	}
	for _, loss := range l.Losses {
		if !loss.IsValid() {
			return invalidError(op, "invalid loss")
		}
	}
	return nil
}

// Equal reports whether l and other have the same bar and losses in order; a
// nil and an empty list are equal.
func (l BarLoss) Equal(other BarLoss) bool {
	return l.Bar == other.Bar && slices.Equal(l.Losses, other.Losses)
}

// String returns the bar, ":" and the losses separated by ",", such as
// "3:symbol,grouping" or "3:". It prints the fields as they are.
func (l BarLoss) String() string {
	var b strings.Builder
	b.WriteString(strconv.Itoa(l.Bar))
	b.WriteByte(':')
	for i, loss := range l.Losses {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(loss.String())
	}
	return b.String()
}

// MarshalText implements encoding.TextMarshaler; it requires valid fields.
func (l BarLoss) MarshalText() ([]byte, error) {
	if err := l.validate("BarLoss.MarshalText"); err != nil {
		return nil, err
	}
	return []byte(l.String()), nil
}

// AppendText implements encoding.TextAppender; it requires valid fields.
func (l BarLoss) AppendText(dst []byte) ([]byte, error) {
	if err := l.validate("BarLoss.AppendText"); err != nil {
		return dst, err
	}
	return append(dst, l.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves l unchanged on
// error.
func (l *BarLoss) UnmarshalText(text []byte) error {
	if l == nil {
		return invalidError("BarLoss.UnmarshalText", "nil receiver")
	}
	v, err := ParseBarLoss(string(text))
	if err != nil {
		return err
	}
	*l = v
	return nil
}

// ParseBarLoss parses the text of String; losses are kept as written, in any
// order and with repetitions.
func ParseBarLoss(text string) (BarLoss, error) {
	barText, rest, ok := strings.Cut(text, ":")
	if !ok || !isUint(barText) {
		return BarLoss{}, grammarError(text)
	}
	var losses []Loss
	for _, item := range splitList(rest, ",") {
		loss, err := ParseLoss(item)
		if err != nil {
			return BarLoss{}, grammarError(text)
		}
		losses = append(losses, loss)
	}
	bar, err := uintValue(text, barText)
	if err != nil {
		return BarLoss{}, err
	}
	return BarLoss{Bar: bar, Losses: losses}, nil
}

// MustParseBarLoss is like ParseBarLoss but panics with its error.
func MustParseBarLoss(text string) BarLoss { return must(ParseBarLoss(text)) }

// MIDIMeterMap is an immutable list of Time Signature events of a resolution
// in non-decreasing tick order. Several events may share a tick; the last one
// is in effect. The zero value is unset. MIDIMeterMap is not comparable with
// ==; use Equal.
type MIDIMeterMap struct {
	r       Resolution
	changes []MIDIMeterChange
}

// NewMIDIMeterMap returns the map of r with the changes in the given order. An
// unset resolution or event is ErrUnset, a negative tick ErrInvalidArgument
// and decreasing ticks ErrOrder.
func NewMIDIMeterMap(r Resolution, changes ...MIDIMeterChange) (MIDIMeterMap, error) {
	const op = "NewMIDIMeterMap"
	if !r.IsSet() {
		return MIDIMeterMap{}, unsetError(op, "resolution")
	}
	for _, c := range changes {
		if err := c.validate(op); err != nil {
			return MIDIMeterMap{}, err
		}
	}
	for i := 1; i < len(changes); i++ {
		if changes[i].Tick < changes[i-1].Tick {
			return MIDIMeterMap{}, newError(ErrOrder, op, "time signature changes out of tick order")
		}
	}
	var stored []MIDIMeterChange
	if len(changes) != 0 {
		stored = slices.Clone(changes)
	}
	return MIDIMeterMap{r: r, changes: stored}, nil
}

// MustNewMIDIMeterMap is like NewMIDIMeterMap but panics with its error.
func MustNewMIDIMeterMap(r Resolution, changes ...MIDIMeterChange) MIDIMeterMap {
	return must(NewMIDIMeterMap(r, changes...))
}

// IsSet reports whether m is set.
func (m MIDIMeterMap) IsSet() bool { return m.r.IsSet() }

func (m MIDIMeterMap) check(op string) error {
	if !m.IsSet() {
		return unsetError(op, "receiver")
	}
	return nil
}

// Resolution returns the resolution of the map.
func (m MIDIMeterMap) Resolution() (Resolution, error) {
	if err := m.check("MIDIMeterMap.Resolution"); err != nil {
		return Resolution{}, err
	}
	return m.r, nil
}

// MustResolution is like Resolution but panics with its error.
func (m MIDIMeterMap) MustResolution() Resolution { return must(m.Resolution()) }

// Changes returns all events in order as a new slice, or nil.
func (m MIDIMeterMap) Changes() ([]MIDIMeterChange, error) {
	if err := m.check("MIDIMeterMap.Changes"); err != nil {
		return nil, err
	}
	if len(m.changes) == 0 {
		return nil, nil
	}
	return slices.Clone(m.changes), nil
}

// MustChanges is like Changes but panics with its error.
func (m MIDIMeterMap) MustChanges() []MIDIMeterChange { return must(m.Changes()) }

// ChangesAt returns the events at tick in order, or nil.
func (m MIDIMeterMap) ChangesAt(tick int) ([]MIDIMeter, error) {
	const op = "MIDIMeterMap.ChangesAt"
	if err := m.check(op); err != nil {
		return nil, err
	}
	if err := checkTick(op, tick); err != nil {
		return nil, err
	}
	var out []MIDIMeter
	for _, c := range m.changes {
		if c.Tick == tick {
			out = append(out, c.Meter)
		}
	}
	return out, nil
}

// MustChangesAt is like ChangesAt but panics with its error.
func (m MIDIMeterMap) MustChangesAt(tick int) []MIDIMeter { return must(m.ChangesAt(tick)) }

// WithChange returns the map with c after all events at its tick.
func (m MIDIMeterMap) WithChange(c MIDIMeterChange) (MIDIMeterMap, error) {
	const op = "MIDIMeterMap.WithChange"
	if err := m.check(op); err != nil {
		return MIDIMeterMap{}, err
	}
	if err := c.validate(op); err != nil {
		return MIDIMeterMap{}, err
	}
	i := insertionIndex(func(i int) int { return m.changes[i].Tick }, len(m.changes), c.Tick)
	m.changes = slices.Insert(slices.Clone(m.changes), i, c)
	return m, nil
}

// MustWithChange is like WithChange but panics with its error.
func (m MIDIMeterMap) MustWithChange(c MIDIMeterChange) MIDIMeterMap { return must(m.WithChange(c)) }

// ReplaceChange returns the map with the index-th event (from 0) at tick
// replaced by t. A missing event is ErrInvalidArgument.
func (m MIDIMeterMap) ReplaceChange(tick, index int, t MIDIMeter) (MIDIMeterMap, error) {
	const op = "MIDIMeterMap.ReplaceChange"
	if err := m.check(op); err != nil {
		return MIDIMeterMap{}, err
	}
	if err := checkTick(op, tick); err != nil {
		return MIDIMeterMap{}, err
	}
	if index < 0 {
		return MIDIMeterMap{}, invalidError(op, "negative index")
	}
	if !t.IsSet() {
		return MIDIMeterMap{}, unsetError(op, "time signature")
	}
	i := nthAt(func(i int) int { return m.changes[i].Tick }, len(m.changes), tick, index)
	if i < 0 {
		return MIDIMeterMap{}, invalidError(op, "no such change")
	}
	m.changes = slices.Clone(m.changes)
	m.changes[i].Meter = t
	return m, nil
}

// MustReplaceChange is like ReplaceChange but panics with its error.
func (m MIDIMeterMap) MustReplaceChange(tick, index int, t MIDIMeter) MIDIMeterMap {
	return must(m.ReplaceChange(tick, index, t))
}

// WithoutChange returns the map without the index-th event (from 0) at tick. A
// missing event gives the same map without error.
func (m MIDIMeterMap) WithoutChange(tick, index int) (MIDIMeterMap, error) {
	const op = "MIDIMeterMap.WithoutChange"
	if err := m.check(op); err != nil {
		return MIDIMeterMap{}, err
	}
	if err := checkTick(op, tick); err != nil {
		return MIDIMeterMap{}, err
	}
	if index < 0 {
		return MIDIMeterMap{}, invalidError(op, "negative index")
	}
	i := nthAt(func(i int) int { return m.changes[i].Tick }, len(m.changes), tick, index)
	if i < 0 {
		return m, nil
	}
	changes := slices.Delete(slices.Clone(m.changes), i, i+1)
	if len(changes) == 0 {
		changes = nil
	}
	m.changes = changes
	return m, nil
}

// MustWithoutChange is like WithoutChange but panics with its error.
func (m MIDIMeterMap) MustWithoutChange(tick, index int) MIDIMeterMap {
	return must(m.WithoutChange(tick, index))
}

// MeterAt returns the event in effect at tick: the last event at a tick <=
// tick, or before the first event the default 4, 2, 24, bb with the bb of the
// MIDI quarter of the map.
func (m MIDIMeterMap) MeterAt(tick int) (MIDIMeter, error) {
	const op = "MIDIMeterMap.MeterAt"
	if err := m.check(op); err != nil {
		return MIDIMeter{}, err
	}
	if err := checkTick(op, tick); err != nil {
		return MIDIMeter{}, err
	}
	current := defaultMeterOf(m.r.bb)
	for _, c := range m.changes {
		if c.Tick > tick {
			break
		}
		current = c.Meter
	}
	return current, nil
}

// MustMeterAt is like MeterAt but panics with its error.
func (m MIDIMeterMap) MustMeterAt(tick int) MIDIMeter { return must(m.MeterAt(tick)) }

// Grid returns the open measure grid of the events without a pickup. Every
// event must have the bb of the MIDI quarter of the map, otherwise it is
// ErrInvalidArgument. Without events, and before a first event not at tick 0,
// the meter is 4/4. At a tick the last event is in effect. An event inside a
// bar ends it: the bar gets an explicit length up to the event, aligned at its
// start, and the next bar starts with the new meter. Every event in effect
// sets its meter.
func (m MIDIMeterMap) Grid() (measure.Grid, error) {
	const op = "MIDIMeterMap.Grid"
	if err := m.check(op); err != nil {
		return measure.Grid{}, err
	}
	for _, c := range m.changes {
		if c.Meter.bb != m.r.bb {
			return measure.Grid{}, invalidError(op, "time signature with a different MIDI quarter notation")
		}
	}
	type effective struct {
		tick  int
		meter meter.Meter
	}
	var events []effective
	for i, c := range m.changes {
		if i+1 < len(m.changes) && m.changes[i+1].Tick == c.Tick {
			continue
		}
		mt, err := c.Meter.Meter()
		if err != nil {
			return measure.Grid{}, err
		}
		events = append(events, effective{tick: c.Tick, meter: mt})
	}
	initial := meter.FourFour()
	if len(events) != 0 && events[0].tick == 0 {
		initial = events[0].meter
		events = events[1:]
	}
	g, err := measure.New(initial)
	if err != nil {
		return measure.Grid{}, dependencyError(op, err)
	}
	bar := 1
	barStart := new(big.Rat)
	size := ratOfDuration(initial.MustDuration())
	for _, e := range events {
		p := new(big.Rat).Mul(ratInt(e.tick), m.r.tick())
		offset := new(big.Rat).Sub(p, barStart)
		bars := new(big.Rat).Quo(offset, size)
		whole := new(big.Int).Quo(bars.Num(), bars.Denom())
		n := bar + int(whole.Int64())
		within := new(big.Rat).Sub(offset, new(big.Rat).Mul(new(big.Rat).SetInt(whole), size))
		if within.Sign() != 0 {
			length, err := durationOf(op, within)
			if err != nil {
				return measure.Grid{}, err
			}
			if g, err = g.WithLength(n, length, measure.AlignStart); err != nil {
				return measure.Grid{}, dependencyError(op, err)
			}
			n++
		}
		if g, err = g.WithMark(n, e.meter); err != nil {
			return measure.Grid{}, dependencyError(op, err)
		}
		bar, barStart = n, p
		size = ratOfDuration(e.meter.MustDuration())
	}
	return g, nil
}

// MustGrid is like Grid but panics with its error.
func (m MIDIMeterMap) MustGrid() measure.Grid { return must(m.Grid()) }

// Rescale returns the map with every tick rescaled exactly to the resolution
// to. A different MIDI quarter notation is ErrInvalidArgument and a tick that
// is not whole ErrNotRepresentable.
func (m MIDIMeterMap) Rescale(to Resolution) (MIDIMeterMap, error) {
	const op = "MIDIMeterMap.Rescale"
	if err := m.check(op); err != nil {
		return MIDIMeterMap{}, err
	}
	if err := rescaleTarget(op, m.r, to); err != nil {
		return MIDIMeterMap{}, err
	}
	return m.rescale(op, to, 0)
}

// MustRescale is like Rescale but panics with its error.
func (m MIDIMeterMap) MustRescale(to Resolution) MIDIMeterMap { return must(m.Rescale(to)) }

// QuantizeRescale returns the map with every tick rescaled to the resolution to
// and rounded with mode; the order is kept.
func (m MIDIMeterMap) QuantizeRescale(to Resolution, mode rounding.Mode) (MIDIMeterMap, error) {
	const op = "MIDIMeterMap.QuantizeRescale"
	if err := m.check(op); err != nil {
		return MIDIMeterMap{}, err
	}
	if err := rescaleTarget(op, m.r, to); err != nil {
		return MIDIMeterMap{}, err
	}
	if err := checkMode(op, mode); err != nil {
		return MIDIMeterMap{}, err
	}
	return m.rescale(op, to, mode)
}

// MustQuantizeRescale is like QuantizeRescale but panics with its error.
func (m MIDIMeterMap) MustQuantizeRescale(to Resolution, mode rounding.Mode) MIDIMeterMap {
	return must(m.QuantizeRescale(to, mode))
}

func (m MIDIMeterMap) rescale(op string, to Resolution, mode rounding.Mode) (MIDIMeterMap, error) {
	var changes []MIDIMeterChange
	if len(m.changes) != 0 {
		changes = make([]MIDIMeterChange, len(m.changes))
	}
	for i, c := range m.changes {
		tick, err := rescaleTick(op, m.r, to, c.Tick, mode)
		if err != nil {
			return MIDIMeterMap{}, err
		}
		changes[i] = MIDIMeterChange{Tick: tick, Meter: c.Meter}
	}
	return MIDIMeterMap{r: to, changes: changes}, nil
}

// Equal reports whether m and other have the same resolution and events in
// order.
func (m MIDIMeterMap) Equal(other MIDIMeterMap) bool {
	return m.r == other.r && slices.Equal(m.changes, other.changes)
}

// String returns "midimeter[" resolution, the events "tick:nn,dd,cc,bb"
// preceded by ";" and "]", such as "midimeter[480;0:4,2,24,8;3840:3,2,24,8]",
// or "MIDIMeterMap(unset)".
func (m MIDIMeterMap) String() string {
	if !m.IsSet() {
		return "MIDIMeterMap(unset)"
	}
	var b strings.Builder
	b.WriteString("midimeter[")
	b.WriteString(m.r.String())
	for _, c := range m.changes {
		b.WriteByte(';')
		b.WriteString(c.String())
	}
	b.WriteByte(']')
	return b.String()
}

// MarshalText implements encoding.TextMarshaler.
func (m MIDIMeterMap) MarshalText() ([]byte, error) {
	if err := m.check("MIDIMeterMap.MarshalText"); err != nil {
		return nil, err
	}
	return []byte(m.String()), nil
}

// AppendText implements encoding.TextAppender.
func (m MIDIMeterMap) AppendText(dst []byte) ([]byte, error) {
	if err := m.check("MIDIMeterMap.AppendText"); err != nil {
		return dst, err
	}
	return append(dst, m.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves m unchanged on
// error.
func (m *MIDIMeterMap) UnmarshalText(text []byte) error {
	if m == nil {
		return invalidError("MIDIMeterMap.UnmarshalText", "nil receiver")
	}
	v, err := ParseMIDIMeterMap(string(text))
	if err != nil {
		return err
	}
	*m = v
	return nil
}

// ParseMIDIMeterMap parses the text of String.
func ParseMIDIMeterMap(text string) (MIDIMeterMap, error) {
	body, ok := strings.CutPrefix(text, "midimeter[")
	if !ok {
		return MIDIMeterMap{}, grammarError(text)
	}
	body, ok = strings.CutSuffix(body, "]")
	if !ok {
		return MIDIMeterMap{}, grammarError(text)
	}
	items := strings.Split(body, ";")
	rawR, ok := scanResolution(items[0])
	if !ok {
		return MIDIMeterMap{}, grammarError(text)
	}
	raws := make([]rawMeterChange, 0, len(items)-1)
	for _, item := range items[1:] {
		raw, ok := scanMeterChange(item)
		if !ok {
			return MIDIMeterMap{}, grammarError(text)
		}
		raws = append(raws, raw)
	}
	rValues, err := rawR.numbers(text)
	if err != nil {
		return MIDIMeterMap{}, err
	}
	nums := make([]meterChangeNumbers, len(raws))
	for i, raw := range raws {
		if nums[i], err = raw.numbers(text); err != nil {
			return MIDIMeterMap{}, err
		}
	}
	r, err := rValues.resolution()
	if err != nil {
		return MIDIMeterMap{}, valueTextError(text, err)
	}
	changes := make([]MIDIMeterChange, len(nums))
	for i, n := range nums {
		if changes[i], err = n.change(); err != nil {
			return MIDIMeterMap{}, valueTextError(text, err)
		}
	}
	m, err := NewMIDIMeterMap(r, changes...)
	if err != nil {
		return MIDIMeterMap{}, valueTextError(text, err)
	}
	return m, nil
}

// MustParseMIDIMeterMap is like ParseMIDIMeterMap but panics with its error.
func MustParseMIDIMeterMap(text string) MIDIMeterMap { return must(ParseMIDIMeterMap(text)) }

// desiredEvent returns the event that the bar of g wants, with its losses.
func desiredEvent(op string, g measure.Grid, bar int, r Resolution, click Click, f Fidelity, irregular IrregularBars) (MIDIMeter, []Loss, error) {
	m, err := g.BarMeter(bar)
	if err != nil {
		return MIDIMeter{}, nil, dependencyError(op, err)
	}
	event, losses, err := MIDIMeterOf(m, click, MIDIQuarter{bb: r.bb}, f)
	if err != nil {
		return MIDIMeter{}, nil, err
	}
	length, err := g.BarLength(bar)
	if err != nil {
		return MIDIMeter{}, nil, dependencyError(op, err)
	}
	if irregular == IrregularSkip || length.Equal(m.MustDuration()) {
		return event, losses, nil
	}
	l := ratOfDuration(length)
	denominator := new(big.Int).Lsh(big.NewInt(1), uint(event.dd))
	if l.Denom().Cmp(denominator) > 0 {
		denominator = l.Denom()
	}
	dd, ok := exponentOf(denominator)
	if !ok {
		return MIDIMeter{}, nil, notMIDI(op, "bar length without a power-of-two denominator")
	}
	nn := new(big.Rat).Mul(l, new(big.Rat).SetInt(denominator))
	if nn.Num().Cmp(big.NewInt(MaxByteField)) > 0 {
		return MIDIMeter{}, nil, notMIDI(op, "numerator of the bar length above 255")
	}
	event.nn, event.dd = uint8(nn.Num().Int64()), uint8(dd)
	return event, losses, nil
}

// lastInterestingBar returns the last bar with a written meter or an explicit
// length, plus one more bar if the grid has it.
func lastInterestingBar(op string, g measure.Grid) (int, error) {
	last := g.MustFirstBar()
	for _, mark := range g.MustMarks() {
		last = max(last, mark.Bar)
	}
	for _, l := range g.MustLengths() {
		last = max(last, l.Bar)
	}
	has, err := g.HasBar(last + 1)
	if err != nil {
		return 0, dependencyError(op, err)
	}
	if has {
		last++
	}
	return last, nil
}

// MIDIMeterMapOf converts the grid g to Time Signature events of the resolution
// r with the click and the fidelity, and returns the losses of a lossy export
// by bar (nil without losses). Bars are visited from the first bar to the last
// bar with a written meter or an explicit length, and one more bar if the grid
// has it. A bar whose length equals its meter, and any bar with IrregularSkip,
// wants the event of its meter; another bar with IrregularEvent wants an event
// of its length with the larger of the denominators of its meter and its
// length. An event is placed at the start of the first bar and of every bar
// whose wanted event differs from the last placed one. A bar that MIDI cannot
// express is ErrMeterNotMIDI and a start that is not a whole number of ticks
// ErrNotRepresentable.
func MIDIMeterMapOf(g measure.Grid, r Resolution, click Click, f Fidelity, irregular IrregularBars) (MIDIMeterMap, []BarLoss, error) {
	const op = "MIDIMeterMapOf"
	if err := checkGrid(op, g); err != nil {
		return MIDIMeterMap{}, nil, err
	}
	if !r.IsSet() {
		return MIDIMeterMap{}, nil, unsetError(op, "resolution")
	}
	if !click.IsSet() {
		return MIDIMeterMap{}, nil, unsetError(op, "click")
	}
	if !f.IsValid() {
		return MIDIMeterMap{}, nil, invalidError(op, "invalid fidelity")
	}
	if !irregular.IsValid() {
		return MIDIMeterMap{}, nil, invalidError(op, "invalid irregular bars")
	}
	last, err := lastInterestingBar(op, g)
	if err != nil {
		return MIDIMeterMap{}, nil, err
	}
	var changes []MIDIMeterChange
	var losses []BarLoss
	for bar := g.MustFirstBar(); bar <= last; bar++ {
		event, lost, err := desiredEvent(op, g, bar, r, click, f, irregular)
		if err != nil {
			return MIDIMeterMap{}, nil, err
		}
		if len(changes) != 0 && changes[len(changes)-1].Meter == event {
			continue
		}
		start, err := g.BarStart(bar)
		if err != nil {
			return MIDIMeterMap{}, nil, dependencyError(op, err)
		}
		tick, err := wholeTicks(op, r.unitsOf(ratOfPosition(start)))
		if err != nil {
			return MIDIMeterMap{}, nil, err
		}
		changes = append(changes, MIDIMeterChange{Tick: tick, Meter: event})
		if len(lost) != 0 {
			losses = append(losses, BarLoss{Bar: bar, Losses: lost})
		}
	}
	return MIDIMeterMap{r: r, changes: changes}, losses, nil
}

// MustMIDIMeterMapOf is like MIDIMeterMapOf but panics with its error.
func MustMIDIMeterMapOf(g measure.Grid, r Resolution, click Click, f Fidelity, irregular IrregularBars) (MIDIMeterMap, []BarLoss) {
	return must2(MIDIMeterMapOf(g, r, click, f, irregular))
}
