package ticks

import (
	"errors"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
)

// MIDITempoChange is one Set Tempo event at an absolute tick. Its fields are
// checked where it is used: a negative tick is ErrInvalidArgument and an unset
// tempo ErrUnset.
type MIDITempoChange struct {
	Tick  int
	Tempo MIDITempo
}

func (c MIDITempoChange) validate(op string) error {
	if !c.Tempo.IsSet() {
		return unsetError(op, "tempo of a change")
	}
	if c.Tick < 0 {
		return invalidError(op, "negative tick of a tempo change")
	}
	return nil
}

// Equal reports whether c and other have equal fields.
func (c MIDITempoChange) Equal(other MIDITempoChange) bool { return c == other }

// String returns the tick, ":" and the tempo, such as "192:1000000". It prints
// the fields as they are, including invalid ones.
func (c MIDITempoChange) String() string { return strconv.Itoa(c.Tick) + ":" + c.Tempo.String() }

// MarshalText implements encoding.TextMarshaler; it requires valid fields.
func (c MIDITempoChange) MarshalText() ([]byte, error) {
	if err := c.validate("MIDITempoChange.MarshalText"); err != nil {
		return nil, err
	}
	return []byte(c.String()), nil
}

// AppendText implements encoding.TextAppender; it requires valid fields.
func (c MIDITempoChange) AppendText(dst []byte) ([]byte, error) {
	if err := c.validate("MIDITempoChange.AppendText"); err != nil {
		return dst, err
	}
	return append(dst, c.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves c unchanged on
// error.
func (c *MIDITempoChange) UnmarshalText(text []byte) error {
	if c == nil {
		return invalidError("MIDITempoChange.UnmarshalText", "nil receiver")
	}
	v, err := ParseMIDITempoChange(string(text))
	if err != nil {
		return err
	}
	*c = v
	return nil
}

// ParseMIDITempoChange parses the text of String.
func ParseMIDITempoChange(text string) (MIDITempoChange, error) {
	tickText, tempoText, ok := strings.Cut(text, ":")
	if !ok || !isUint(tickText) || !isUint(tempoText) {
		return MIDITempoChange{}, grammarError(text)
	}
	tick, err := uintValue(text, tickText)
	if err != nil {
		return MIDITempoChange{}, err
	}
	us, err := uintValue(text, tempoText)
	if err != nil {
		return MIDITempoChange{}, err
	}
	t, err := NewMIDITempo(us)
	if err != nil {
		return MIDITempoChange{}, valueTextError(text, err)
	}
	return MIDITempoChange{Tick: tick, Tempo: t}, nil
}

// MustParseMIDITempoChange is like ParseMIDITempoChange but panics with its
// error.
func MustParseMIDITempoChange(text string) MIDITempoChange { return must(ParseMIDITempoChange(text)) }

// MIDITempoMap is an immutable MIDI tempo map: a resolution, an initial tempo
// in effect before the first change (500000 by default) and Set Tempo events
// in non-decreasing tick order. Several changes may share a tick; the last one
// is in effect. The zero value is unset. MIDITempoMap is not comparable with
// ==; use Equal.
type MIDITempoMap struct {
	r       Resolution
	initial MIDITempo
	changes []MIDITempoChange
}

// NewMIDITempoMap returns the map of r with the default initial tempo and the
// changes in the given order. An unset resolution or tempo is ErrUnset, a
// negative tick ErrInvalidArgument and decreasing ticks ErrOrder.
func NewMIDITempoMap(r Resolution, changes ...MIDITempoChange) (MIDITempoMap, error) {
	const op = "NewMIDITempoMap"
	if !r.IsSet() {
		return MIDITempoMap{}, unsetError(op, "resolution")
	}
	for _, c := range changes {
		if err := c.validate(op); err != nil {
			return MIDITempoMap{}, err
		}
	}
	for i := 1; i < len(changes); i++ {
		if changes[i].Tick < changes[i-1].Tick {
			return MIDITempoMap{}, newError(ErrOrder, op, "tempo changes out of tick order")
		}
	}
	var stored []MIDITempoChange
	if len(changes) != 0 {
		stored = slices.Clone(changes)
	}
	return MIDITempoMap{r: r, initial: DefaultMIDITempo(), changes: stored}, nil
}

// MustNewMIDITempoMap is like NewMIDITempoMap but panics with its error.
func MustNewMIDITempoMap(r Resolution, changes ...MIDITempoChange) MIDITempoMap {
	return must(NewMIDITempoMap(r, changes...))
}

// IsSet reports whether m is set.
func (m MIDITempoMap) IsSet() bool { return m.r.IsSet() }

func (m MIDITempoMap) check(op string) error {
	if !m.IsSet() {
		return unsetError(op, "receiver")
	}
	return nil
}

// Resolution returns the resolution of the map.
func (m MIDITempoMap) Resolution() (Resolution, error) {
	if err := m.check("MIDITempoMap.Resolution"); err != nil {
		return Resolution{}, err
	}
	return m.r, nil
}

// MustResolution is like Resolution but panics with its error.
func (m MIDITempoMap) MustResolution() Resolution { return must(m.Resolution()) }

// Initial returns the tempo in effect before the first change.
func (m MIDITempoMap) Initial() (MIDITempo, error) {
	if err := m.check("MIDITempoMap.Initial"); err != nil {
		return MIDITempo{}, err
	}
	return m.initial, nil
}

// MustInitial is like Initial but panics with its error.
func (m MIDITempoMap) MustInitial() MIDITempo { return must(m.Initial()) }

// Changes returns all changes in order as a new slice, or nil.
func (m MIDITempoMap) Changes() ([]MIDITempoChange, error) {
	if err := m.check("MIDITempoMap.Changes"); err != nil {
		return nil, err
	}
	if len(m.changes) == 0 {
		return nil, nil
	}
	return slices.Clone(m.changes), nil
}

// MustChanges is like Changes but panics with its error.
func (m MIDITempoMap) MustChanges() []MIDITempoChange { return must(m.Changes()) }

func checkTick(op string, tick int) error {
	if tick < 0 {
		return invalidError(op, "negative tick")
	}
	return nil
}

// ChangesAt returns the tempos of the changes at tick in order, or nil.
func (m MIDITempoMap) ChangesAt(tick int) ([]MIDITempo, error) {
	const op = "MIDITempoMap.ChangesAt"
	if err := m.check(op); err != nil {
		return nil, err
	}
	if err := checkTick(op, tick); err != nil {
		return nil, err
	}
	var out []MIDITempo
	for _, c := range m.changes {
		if c.Tick == tick {
			out = append(out, c.Tempo)
		}
	}
	return out, nil
}

// MustChangesAt is like ChangesAt but panics with its error.
func (m MIDITempoMap) MustChangesAt(tick int) []MIDITempo { return must(m.ChangesAt(tick)) }

// WithInitial returns the map with the initial tempo t.
func (m MIDITempoMap) WithInitial(t MIDITempo) (MIDITempoMap, error) {
	const op = "MIDITempoMap.WithInitial"
	if err := m.check(op); err != nil {
		return MIDITempoMap{}, err
	}
	if !t.IsSet() {
		return MIDITempoMap{}, unsetError(op, "initial tempo")
	}
	m.initial = t
	return m, nil
}

// MustWithInitial is like WithInitial but panics with its error.
func (m MIDITempoMap) MustWithInitial(t MIDITempo) MIDITempoMap { return must(m.WithInitial(t)) }

// insertionIndex returns the index after all changes at ticks <= tick.
func insertionIndex(ticks func(int) int, n, tick int) int {
	i := 0
	for i < n && ticks(i) <= tick {
		i++
	}
	return i
}

// nthAt returns the slice index of the index-th change at tick, or -1.
func nthAt(ticks func(int) int, n, tick, index int) int {
	seen := 0
	for i := range n {
		if ticks(i) == tick {
			if seen == index {
				return i
			}
			seen++
		}
	}
	return -1
}

// WithChange returns the map with c after all changes at its tick.
func (m MIDITempoMap) WithChange(c MIDITempoChange) (MIDITempoMap, error) {
	const op = "MIDITempoMap.WithChange"
	if err := m.check(op); err != nil {
		return MIDITempoMap{}, err
	}
	if err := c.validate(op); err != nil {
		return MIDITempoMap{}, err
	}
	i := insertionIndex(func(i int) int { return m.changes[i].Tick }, len(m.changes), c.Tick)
	m.changes = slices.Insert(slices.Clone(m.changes), i, c)
	return m, nil
}

// MustWithChange is like WithChange but panics with its error.
func (m MIDITempoMap) MustWithChange(c MIDITempoChange) MIDITempoMap { return must(m.WithChange(c)) }

// ReplaceChange returns the map with the tempo of the index-th change (from 0)
// at tick replaced by t. A missing change is ErrInvalidArgument.
func (m MIDITempoMap) ReplaceChange(tick, index int, t MIDITempo) (MIDITempoMap, error) {
	const op = "MIDITempoMap.ReplaceChange"
	if err := m.check(op); err != nil {
		return MIDITempoMap{}, err
	}
	if err := checkTick(op, tick); err != nil {
		return MIDITempoMap{}, err
	}
	if index < 0 {
		return MIDITempoMap{}, invalidError(op, "negative index")
	}
	if !t.IsSet() {
		return MIDITempoMap{}, unsetError(op, "tempo")
	}
	i := nthAt(func(i int) int { return m.changes[i].Tick }, len(m.changes), tick, index)
	if i < 0 {
		return MIDITempoMap{}, invalidError(op, "no such change")
	}
	m.changes = slices.Clone(m.changes)
	m.changes[i].Tempo = t
	return m, nil
}

// MustReplaceChange is like ReplaceChange but panics with its error.
func (m MIDITempoMap) MustReplaceChange(tick, index int, t MIDITempo) MIDITempoMap {
	return must(m.ReplaceChange(tick, index, t))
}

// WithoutChange returns the map without the index-th change (from 0) at tick.
// A missing change gives the same map without error.
func (m MIDITempoMap) WithoutChange(tick, index int) (MIDITempoMap, error) {
	const op = "MIDITempoMap.WithoutChange"
	if err := m.check(op); err != nil {
		return MIDITempoMap{}, err
	}
	if err := checkTick(op, tick); err != nil {
		return MIDITempoMap{}, err
	}
	if index < 0 {
		return MIDITempoMap{}, invalidError(op, "negative index")
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
func (m MIDITempoMap) MustWithoutChange(tick, index int) MIDITempoMap {
	return must(m.WithoutChange(tick, index))
}

// tempoAtRat returns the tempo in effect at a possibly fractional tick.
func (m MIDITempoMap) tempoAtRat(tick *big.Rat) MIDITempo {
	current := m.initial
	for _, c := range m.changes {
		if ratInt(c.Tick).Cmp(tick) > 0 {
			break
		}
		current = c.Tempo
	}
	return current
}

// TempoAt returns the tempo in effect at tick: the last change at a tick <=
// tick, or the initial tempo.
func (m MIDITempoMap) TempoAt(tick int) (MIDITempo, error) {
	const op = "MIDITempoMap.TempoAt"
	if err := m.check(op); err != nil {
		return MIDITempo{}, err
	}
	if err := checkTick(op, tick); err != nil {
		return MIDITempo{}, err
	}
	return m.tempoAtRat(ratInt(tick)), nil
}

// MustTempoAt is like TempoAt but panics with its error.
func (m MIDITempoMap) MustTempoAt(tick int) MIDITempo { return must(m.TempoAt(tick)) }

// scale returns D*10^6: elapsed tick-microseconds divided by it are seconds.
func (m MIDITempoMap) scale() *big.Rat {
	return new(big.Rat).Mul(ratInt(m.r.d), big.NewRat(1000000, 1))
}

// secondsAtRat returns the exact seconds of a non-negative fractional tick.
func (m MIDITempoMap) secondsAtRat(tick *big.Rat) *big.Rat {
	acc := new(big.Rat)
	previous := new(big.Rat)
	current := m.initial
	for _, c := range m.changes {
		at := ratInt(c.Tick)
		if at.Cmp(tick) > 0 {
			break
		}
		acc.Add(acc, new(big.Rat).Mul(new(big.Rat).Sub(at, previous), ratInt(int(current.us))))
		previous = at
		current = c.Tempo
	}
	acc.Add(acc, new(big.Rat).Mul(new(big.Rat).Sub(tick, previous), ratInt(int(current.us))))
	return acc.Quo(acc, m.scale())
}

// ticksAtRat returns the exact fractional tick of non-negative seconds.
func (m MIDITempoMap) ticksAtRat(seconds *big.Rat) *big.Rat {
	target := new(big.Rat).Mul(seconds, m.scale())
	acc := new(big.Rat)
	previous := new(big.Rat)
	current := m.initial
	for _, c := range m.changes {
		at := ratInt(c.Tick)
		end := new(big.Rat).Add(acc, new(big.Rat).Mul(new(big.Rat).Sub(at, previous), ratInt(int(current.us))))
		if target.Cmp(end) <= 0 {
			break
		}
		acc = end
		previous = at
		current = c.Tempo
	}
	offset := new(big.Rat).Sub(target, acc)
	offset.Quo(offset, ratInt(int(current.us)))
	return offset.Add(offset, previous)
}

// Seconds returns the exact seconds at tick >= 0: the sum of dt/D*m/10^6 over
// the pieces of constant tempo.
func (m MIDITempoMap) Seconds(tick int) (tempo.Seconds, error) {
	const op = "MIDITempoMap.Seconds"
	if err := m.check(op); err != nil {
		return tempo.Seconds{}, err
	}
	if err := checkTick(op, tick); err != nil {
		return tempo.Seconds{}, err
	}
	return secondsOf(op, m.secondsAtRat(ratInt(tick)))
}

// MustSeconds is like Seconds but panics with its error.
func (m MIDITempoMap) MustSeconds(tick int) tempo.Seconds { return must(m.Seconds(tick)) }

// SecondsBetween returns Seconds(to) - Seconds(from); from > to is
// ErrInvalidArgument.
func (m MIDITempoMap) SecondsBetween(from, to int) (tempo.Seconds, error) {
	const op = "MIDITempoMap.SecondsBetween"
	if err := m.check(op); err != nil {
		return tempo.Seconds{}, err
	}
	if err := checkTick(op, from); err != nil {
		return tempo.Seconds{}, err
	}
	if err := checkTick(op, to); err != nil {
		return tempo.Seconds{}, err
	}
	if from > to {
		return tempo.Seconds{}, invalidError(op, "from follows to")
	}
	value := new(big.Rat).Sub(m.secondsAtRat(ratInt(to)), m.secondsAtRat(ratInt(from)))
	return secondsOf(op, value)
}

// MustSecondsBetween is like SecondsBetween but panics with its error.
func (m MIDITempoMap) MustSecondsBetween(from, to int) tempo.Seconds {
	return must(m.SecondsBetween(from, to))
}

// TicksAt returns the exact fractional tick at seconds.
func (m MIDITempoMap) TicksAt(s tempo.Seconds) (fraction.Fraction, error) {
	const op = "MIDITempoMap.TicksAt"
	if err := m.check(op); err != nil {
		return fraction.Fraction{}, err
	}
	if err := secondsArgument(op, s); err != nil {
		return fraction.Fraction{}, err
	}
	return fractionOf(op, m.ticksAtRat(ratOfSeconds(s)))
}

// MustTicksAt is like TicksAt but panics with its error.
func (m MIDITempoMap) MustTicksAt(s tempo.Seconds) fraction.Fraction { return must(m.TicksAt(s)) }

// TickAt returns the tick at seconds rounded once with mode.
func (m MIDITempoMap) TickAt(s tempo.Seconds, mode rounding.Mode) (int, error) {
	const op = "MIDITempoMap.TickAt"
	if err := m.check(op); err != nil {
		return 0, err
	}
	if err := secondsArgument(op, s); err != nil {
		return 0, err
	}
	if err := checkMode(op, mode); err != nil {
		return 0, err
	}
	n, err := roundInt(op, m.ticksAtRat(ratOfSeconds(s)), mode)
	if err != nil {
		return 0, err
	}
	return intOf(op, n)
}

// MustTickAt is like TickAt but panics with its error.
func (m MIDITempoMap) MustTickAt(s tempo.Seconds, mode rounding.Mode) int {
	return must(m.TickAt(s, mode))
}

// PositionSeconds returns the exact seconds of the position p, at the
// fractional tick p*D/w.
func (m MIDITempoMap) PositionSeconds(p duration.Position) (tempo.Seconds, error) {
	const op = "MIDITempoMap.PositionSeconds"
	if err := m.check(op); err != nil {
		return tempo.Seconds{}, err
	}
	if err := checkPosition(op, p); err != nil {
		return tempo.Seconds{}, err
	}
	return secondsOf(op, m.secondsAtRat(m.r.unitsOf(ratOfPosition(p))))
}

// MustPositionSeconds is like PositionSeconds but panics with its error.
func (m MIDITempoMap) MustPositionSeconds(p duration.Position) tempo.Seconds {
	return must(m.PositionSeconds(p))
}

// PositionAt returns the exact position at seconds.
func (m MIDITempoMap) PositionAt(s tempo.Seconds) (duration.Position, error) {
	const op = "MIDITempoMap.PositionAt"
	if err := m.check(op); err != nil {
		return duration.Position{}, err
	}
	if err := secondsArgument(op, s); err != nil {
		return duration.Position{}, err
	}
	ticks := m.ticksAtRat(ratOfSeconds(s))
	return positionOf(op, ticks.Mul(ticks, m.r.tick()))
}

// MustPositionAt is like PositionAt but panics with its error.
func (m MIDITempoMap) MustPositionAt(s tempo.Seconds) duration.Position {
	return must(m.PositionAt(s))
}

func smpteArgument(op string, s SMPTE) error {
	if !s.IsSet() {
		return unsetError(op, "SMPTE resolution")
	}
	return nil
}

// SMPTETick returns the SMPTE tick of the metric tick: the exact seconds of
// tick times the SMPTE ticks per second, rounded once with mode. Code -29
// without a reading is ErrInvalidArgument.
func (m MIDITempoMap) SMPTETick(s SMPTE, tick int, mode rounding.Mode) (int, error) {
	const op = "MIDITempoMap.SMPTETick"
	if err := m.check(op); err != nil {
		return 0, err
	}
	if err := smpteArgument(op, s); err != nil {
		return 0, err
	}
	if err := checkTick(op, tick); err != nil {
		return 0, err
	}
	if err := checkMode(op, mode); err != nil {
		return 0, err
	}
	tps, err := s.ticksPerSecond(op)
	if err != nil {
		return 0, err
	}
	n, err := roundInt(op, tps.Mul(tps, m.secondsAtRat(ratInt(tick))), mode)
	if err != nil {
		return 0, err
	}
	return intOf(op, n)
}

// MustSMPTETick is like SMPTETick but panics with its error.
func (m MIDITempoMap) MustSMPTETick(s SMPTE, tick int, mode rounding.Mode) int {
	return must(m.SMPTETick(s, tick, mode))
}

// TickFromSMPTE returns the metric tick of the SMPTE tick smpteTick rounded once
// with mode from the exact value.
func (m MIDITempoMap) TickFromSMPTE(s SMPTE, smpteTick int, mode rounding.Mode) (int, error) {
	const op = "MIDITempoMap.TickFromSMPTE"
	if err := m.check(op); err != nil {
		return 0, err
	}
	if err := smpteArgument(op, s); err != nil {
		return 0, err
	}
	if err := checkTick(op, smpteTick); err != nil {
		return 0, err
	}
	if err := checkMode(op, mode); err != nil {
		return 0, err
	}
	tps, err := s.ticksPerSecond(op)
	if err != nil {
		return 0, err
	}
	n, err := roundInt(op, m.ticksAtRat(new(big.Rat).Quo(ratInt(smpteTick), tps)), mode)
	if err != nil {
		return 0, err
	}
	return intOf(op, n)
}

// MustTickFromSMPTE is like TickFromSMPTE but panics with its error.
func (m MIDITempoMap) MustTickFromSMPTE(s SMPTE, smpteTick int, mode rounding.Mode) int {
	return must(m.TickFromSMPTE(s, smpteTick, mode))
}

// rescaleTarget checks the target resolution of a rescaled map.
func rescaleTarget(op string, from, to Resolution) error {
	if !to.IsSet() {
		return unsetError(op, "target resolution")
	}
	if to.bb != from.bb {
		return invalidError(op, "different MIDI quarter notation")
	}
	return nil
}

// rescaleTicks returns the ticks of r at the resolution to; with a nil mode the
// result must be whole.
func rescaleTick(op string, from, to Resolution, tick int, mode rounding.Mode) (int, error) {
	value := from.rescaled(tick, to)
	if mode == 0 {
		if !value.IsInt() {
			return 0, newError(ErrNotRepresentable, op, "rescaled tick is not whole")
		}
		return intOf(op, value.Num())
	}
	n, err := roundInt(op, value, mode)
	if err != nil {
		return 0, err
	}
	return intOf(op, n)
}

// Rescale returns the map with every tick rescaled exactly to the resolution
// to. A different MIDI quarter notation is ErrInvalidArgument and a tick that
// is not whole ErrNotRepresentable.
func (m MIDITempoMap) Rescale(to Resolution) (MIDITempoMap, error) {
	const op = "MIDITempoMap.Rescale"
	if err := m.check(op); err != nil {
		return MIDITempoMap{}, err
	}
	if err := rescaleTarget(op, m.r, to); err != nil {
		return MIDITempoMap{}, err
	}
	return m.rescale(op, to, 0)
}

// MustRescale is like Rescale but panics with its error.
func (m MIDITempoMap) MustRescale(to Resolution) MIDITempoMap { return must(m.Rescale(to)) }

// QuantizeRescale returns the map with every tick rescaled to the resolution to
// and rounded with mode. The order is kept; ticks that coincide give several
// changes at one tick.
func (m MIDITempoMap) QuantizeRescale(to Resolution, mode rounding.Mode) (MIDITempoMap, error) {
	const op = "MIDITempoMap.QuantizeRescale"
	if err := m.check(op); err != nil {
		return MIDITempoMap{}, err
	}
	if err := rescaleTarget(op, m.r, to); err != nil {
		return MIDITempoMap{}, err
	}
	if err := checkMode(op, mode); err != nil {
		return MIDITempoMap{}, err
	}
	return m.rescale(op, to, mode)
}

// MustQuantizeRescale is like QuantizeRescale but panics with its error.
func (m MIDITempoMap) MustQuantizeRescale(to Resolution, mode rounding.Mode) MIDITempoMap {
	return must(m.QuantizeRescale(to, mode))
}

func (m MIDITempoMap) rescale(op string, to Resolution, mode rounding.Mode) (MIDITempoMap, error) {
	var changes []MIDITempoChange
	if len(m.changes) != 0 {
		changes = make([]MIDITempoChange, len(m.changes))
	}
	for i, c := range m.changes {
		tick, err := rescaleTick(op, m.r, to, c.Tick, mode)
		if err != nil {
			return MIDITempoMap{}, err
		}
		changes[i] = MIDITempoChange{Tick: tick, Tempo: c.Tempo}
	}
	return MIDITempoMap{r: to, initial: m.initial, changes: changes}, nil
}

// TempoMap returns the positional musex tempo map of m: a new main tempo at
// position 0 equal to the tempo in effect at tick 0, and a new main tempo at
// every later tick with changes, by the change in effect there. Markings use
// unit and the MIDI quarter of the map.
func (m MIDITempoMap) TempoMap(unit duration.Written) (tempomap.Map, error) {
	const op = "MIDITempoMap.TempoMap"
	if err := m.check(op); err != nil {
		return tempomap.Map{}, err
	}
	u, err := unitArgument(op, unit)
	if err != nil {
		return tempomap.Map{}, err
	}
	return m.tempoMap(op, unit, u)
}

func (m MIDITempoMap) tempoMap(op string, unit duration.Written, u *big.Rat) (tempomap.Map, error) {
	w := m.r.quarter()
	initial, err := markingOf(op, int(m.tempoAtRat(new(big.Rat)).us), unit, u, w)
	if err != nil {
		return tempomap.Map{}, err
	}
	result, err := tempomap.New(initial)
	if err != nil {
		return tempomap.Map{}, dependencyError(op, err)
	}
	for i, c := range m.changes {
		if c.Tick == 0 || (i+1 < len(m.changes) && m.changes[i+1].Tick == c.Tick) {
			continue
		}
		marking, err := markingOf(op, int(c.Tempo.us), unit, u, w)
		if err != nil {
			return tempomap.Map{}, err
		}
		position, err := positionOf(op, new(big.Rat).Mul(ratInt(c.Tick), m.r.tick()))
		if err != nil {
			return tempomap.Map{}, err
		}
		change, err := tempomap.NewMain(tempomap.MustAt(position), marking)
		if err != nil {
			return tempomap.Map{}, dependencyError(op, err)
		}
		if result, err = result.WithChange(change); err != nil {
			return tempomap.Map{}, dependencyError(op, err)
		}
	}
	return result, nil
}

// MustTempoMap is like TempoMap but panics with its error.
func (m MIDITempoMap) MustTempoMap(unit duration.Written) tempomap.Map { return must(m.TempoMap(unit)) }

// MeasuredTempoMap returns TempoMap converted to a measure map of the grid g. A
// position outside the grid is ErrNotConvertible with the causes of tempomap
// and measure.
func (m MIDITempoMap) MeasuredTempoMap(g measure.Grid, unit duration.Written) (tempomap.Map, error) {
	const op = "MIDITempoMap.MeasuredTempoMap"
	if err := m.check(op); err != nil {
		return tempomap.Map{}, err
	}
	if !g.IsSet() {
		return tempomap.Map{}, invalidError(op, "grid", measure.ErrUnset)
	}
	u, err := unitArgument(op, unit)
	if err != nil {
		return tempomap.Map{}, err
	}
	positional, err := m.tempoMap(op, unit, u)
	if err != nil {
		return tempomap.Map{}, err
	}
	measured, err := positional.ToMeasures(g)
	if err != nil {
		if errors.Is(err, tempomap.ErrInvalidAddress) {
			return tempomap.Map{}, newError(ErrNotConvertible, op, "position outside the grid", err)
		}
		return tempomap.Map{}, dependencyError(op, err)
	}
	return measured, nil
}

// MustMeasuredTempoMap is like MeasuredTempoMap but panics with its error.
func (m MIDITempoMap) MustMeasuredTempoMap(g measure.Grid, unit duration.Written) tempomap.Map {
	return must(m.MeasuredTempoMap(g, unit))
}

// Equal reports whether m and other have the same record: resolution, initial
// tempo and changes in order.
func (m MIDITempoMap) Equal(other MIDITempoMap) bool {
	return m.r == other.r && m.initial == other.initial && slices.Equal(m.changes, other.changes)
}

// slope returns m/w of the tempo in effect at a position: proportional to the
// seconds per whole note.
func (m MIDITempoMap) slope(position *big.Rat) *big.Rat {
	t := m.tempoAtRat(m.r.unitsOf(position))
	return new(big.Rat).Quo(ratInt(int(t.us)), m.r.quarter())
}

// EqualTimeline reports whether m and other give equal seconds at every
// position p >= 0. Resolutions may differ.
func (m MIDITempoMap) EqualTimeline(other MIDITempoMap) (bool, error) {
	const op = "MIDITempoMap.EqualTimeline"
	if err := m.check(op); err != nil {
		return false, err
	}
	if !other.IsSet() {
		return false, unsetError(op, "other map")
	}
	points := []*big.Rat{new(big.Rat)}
	for _, mm := range []MIDITempoMap{m, other} {
		for _, c := range mm.changes {
			points = append(points, new(big.Rat).Mul(ratInt(c.Tick), mm.r.tick()))
		}
	}
	for _, p := range points {
		if m.slope(p).Cmp(other.slope(p)) != 0 {
			return false, nil
		}
	}
	return true, nil
}

// MustEqualTimeline is like EqualTimeline but panics with its error.
func (m MIDITempoMap) MustEqualTimeline(other MIDITempoMap) bool { return must(m.EqualTimeline(other)) }

// String returns "miditempo[" resolution ";" initial tempo, the changes
// "tick:tempo" separated by ";" and "]", such as
// "miditempo[96;500000;0:500000;192:1000000]", or "MIDITempoMap(unset)".
func (m MIDITempoMap) String() string {
	if !m.IsSet() {
		return "MIDITempoMap(unset)"
	}
	var b strings.Builder
	b.WriteString("miditempo[")
	b.WriteString(m.r.String())
	b.WriteByte(';')
	b.WriteString(m.initial.String())
	for _, c := range m.changes {
		b.WriteByte(';')
		b.WriteString(c.String())
	}
	b.WriteByte(']')
	return b.String()
}

// MarshalText implements encoding.TextMarshaler.
func (m MIDITempoMap) MarshalText() ([]byte, error) {
	if err := m.check("MIDITempoMap.MarshalText"); err != nil {
		return nil, err
	}
	return []byte(m.String()), nil
}

// AppendText implements encoding.TextAppender.
func (m MIDITempoMap) AppendText(dst []byte) ([]byte, error) {
	if err := m.check("MIDITempoMap.AppendText"); err != nil {
		return dst, err
	}
	return append(dst, m.String()...), nil
}

// UnmarshalText implements encoding.TextUnmarshaler; it leaves m unchanged on
// error.
func (m *MIDITempoMap) UnmarshalText(text []byte) error {
	if m == nil {
		return invalidError("MIDITempoMap.UnmarshalText", "nil receiver")
	}
	v, err := ParseMIDITempoMap(string(text))
	if err != nil {
		return err
	}
	*m = v
	return nil
}

// ParseMIDITempoMap parses the text of String.
func ParseMIDITempoMap(text string) (MIDITempoMap, error) {
	body, ok := strings.CutPrefix(text, "miditempo[")
	if !ok {
		return MIDITempoMap{}, grammarError(text)
	}
	body, ok = strings.CutSuffix(body, "]")
	if !ok {
		return MIDITempoMap{}, grammarError(text)
	}
	items := strings.Split(body, ";")
	if len(items) < 2 {
		return MIDITempoMap{}, grammarError(text)
	}
	rawR, ok := scanResolution(items[0])
	if !ok || !isUint(items[1]) {
		return MIDITempoMap{}, grammarError(text)
	}
	type rawChange struct{ tick, us string }
	raws := make([]rawChange, 0, len(items)-2)
	for _, item := range items[2:] {
		tick, us, found := strings.Cut(item, ":")
		if !found || !isUint(tick) || !isUint(us) {
			return MIDITempoMap{}, grammarError(text)
		}
		raws = append(raws, rawChange{tick: tick, us: us})
	}
	rValues, err := rawR.numbers(text)
	if err != nil {
		return MIDITempoMap{}, err
	}
	initialUS, err := uintValue(text, items[1])
	if err != nil {
		return MIDITempoMap{}, err
	}
	type numbers struct{ tick, us int }
	nums := make([]numbers, len(raws))
	for i, raw := range raws {
		if nums[i].tick, err = uintValue(text, raw.tick); err != nil {
			return MIDITempoMap{}, err
		}
		if nums[i].us, err = uintValue(text, raw.us); err != nil {
			return MIDITempoMap{}, err
		}
	}
	r, err := rValues.resolution()
	if err != nil {
		return MIDITempoMap{}, valueTextError(text, err)
	}
	initial, err := NewMIDITempo(initialUS)
	if err != nil {
		return MIDITempoMap{}, valueTextError(text, err)
	}
	changes := make([]MIDITempoChange, len(nums))
	for i, n := range nums {
		t, err := NewMIDITempo(n.us)
		if err != nil {
			return MIDITempoMap{}, valueTextError(text, err)
		}
		changes[i] = MIDITempoChange{Tick: n.tick, Tempo: t}
	}
	m, err := NewMIDITempoMap(r, changes...)
	if err != nil {
		return MIDITempoMap{}, valueTextError(text, err)
	}
	m.initial = initial
	return m, nil
}

// MustParseMIDITempoMap is like ParseMIDITempoMap but panics with its error.
func MustParseMIDITempoMap(text string) MIDITempoMap { return must(ParseMIDITempoMap(text)) }
