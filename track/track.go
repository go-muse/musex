package track

import (
	"errors"
	"fmt"
	"slices"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/voice"
)

// VoiceID identifies one voice in a family of immutable track snapshots.
// Its zero value is invalid. Keys may be compared and used in maps.
type VoiceID struct{ key *voiceKey }
type voiceKey struct{ allocated bool }

// Place is a zero-based event index in a voice and bar.
type Place struct {
	Voice      VoiceID
	Bar, Index int
}

// Span is a half-open interval in whole-note positions.
type Span struct{ Start, End duration.Position }

// Entry is an event's original record and computed interval.
type Entry struct {
	Place Place
	Event *event.Event
	Time  Span
}

// BarEdit replaces the complete record of a cell.
type BarEdit struct {
	Voice  VoiceID
	Bar    int
	Events []*event.Event
}

// BarlineEdit changes a sign after a particular bar.
type BarlineEdit struct {
	AfterBar int
	Fermata  bool
}

// Edit describes an atomic replacement. Nil optional fields preserve their values.
type Edit struct {
	Grid     *measure.Grid
	Bars     []BarEdit
	Barlines []BarlineEdit
	Clock    *Clock
}

// BarHold is an independently scheduled delay at a bar's end.
type BarHold struct {
	AfterBar int
	Seconds  tempo.Seconds
}

// Clock holds a measured map and delays attached to particular bar endings.
type Clock struct {
	Map      tempomap.Map
	BarHolds []BarHold
}

// Occurrence identifies a note occurrence, including duplicates in a chord.
type Occurrence struct {
	Place     Place
	NoteIndex int
}

// Sound is one sustained note, with its original attack and tied parts.
type Sound struct {
	Note  note.Note
	Time  Span
	Parts []Occurrence
}

// TimedSound contains the map's exact or bounded second results.
type TimedSound struct {
	Sound      Sound
	Start, End tempomap.SecondsResult
}

// Groups controls notation when dividing events.
type Groups uint8

const (
	KeepGroups Groups = iota
	WithoutGroups
)

// WriteMode controls whether a voice's prior record is retained.
type WriteMode uint8

const (
	Vacant WriteMode = iota
	ReplaceVoice
)

// WriteOptions configures explicit sequence layout.
type WriteOptions struct {
	Groups Groups
	Mode   WriteMode
}

// Change relates original and resulting records to the input elements.
type Change struct {
	From, To []Entry
	Input    []int
}

// MarkKind identifies a moved mark.
type MarkKind uint8

const (
	TempoChange MarkKind = iota + 1
	Extension
	BarlineFermata
)

// MarkMove retains both forms of a mark's attachment and its positions.
type MarkMove struct {
	From, To                 duration.Position
	OldAddress, NewAddress   measure.Address
	OldAfterBar, NewAfterBar *int
	Kind                     MarkKind
	Slot                     tempomap.Slot
}

// Proposal is a complete new snapshot and an independently owned change report.
type Proposal struct {
	Track   *Track
	Changes []Change
	Marks   []MarkMove
}

// LineMode controls whether grace records prevent extraction.
type LineMode uint8

const (
	Exact LineMode = iota
	PrincipalOnly
)

// Extraction is a line and the information that remains in the track.
type Extraction struct {
	Line              *voice.Line
	Gaps              []Span
	Measured, Skipped []Entry
	FermataBars       []int
	Extensions        []duration.Position
}

// Issue locates a failure; unused indexes are -1 and optional values unset.
type Issue struct {
	Cause                 error
	Place                 Place
	NoteIndex, InputIndex int
	Position              duration.Position
	Address               measure.Address
	AfterBar              *int
	Amount                duration.Duration
	BarsNeeded            int
	Kind                  MarkKind
	Slot                  tempomap.Slot
}

// Problems aggregates all independently detected failures in record order.
type Problems struct{ Issues []Issue }

func (p *Problems) Error() string {
	if p == nil || len(p.Issues) == 0 {
		return "track: problems"
	}
	return fmt.Sprintf("track: %d issue(s): %v", len(p.Issues), p.Issues[0].Cause)
}
func (p *Problems) Unwrap() []error {
	if p == nil {
		return nil
	}
	out := make([]error, len(p.Issues))
	for i, x := range p.Issues {
		out[i] = x.Cause
	}
	return out
}

var (
	ErrUnset           = errors.New("track: unset track")
	ErrVoice           = errors.New("track: unknown voice")
	ErrInvalidArgument = errors.New("track: invalid argument")
	ErrPlace           = errors.New("track: invalid place")
	ErrOverflow        = errors.New("track: bar overflow")
	ErrLostRecord      = errors.New("track: lost record")
	ErrOccupied        = errors.New("track: occupied record")
	ErrNeedBars        = errors.New("track: additional bars required")
	ErrMeasureRest     = errors.New("track: invalid measured rest")
	ErrTieEnd          = errors.New("track: tie at end")
	ErrTieRest         = errors.New("track: tie to rest")
	ErrTieGap          = errors.New("track: tie across gap")
	ErrTiePitch        = errors.New("track: tie pitch differs")
	ErrTieMultiplicity = errors.New("track: insufficient tied occurrences")
	ErrGraceTarget     = errors.New("track: missing grace target")
	ErrWouldLoseMark   = errors.New("track: would lose mark")
	ErrNotation        = errors.New("track: unrepresentable notation")
	ErrClockAddress    = errors.New("track: invalid clock address")
	ErrNoClock         = errors.New("track: no clock")
	ErrCollapsed       = errors.New("track: collapsed interval")
	ErrCrossBar        = errors.New("track: quantization crosses barline")
	ErrHoldConflict    = errors.New("track: conflicting hold")
)

func issue(cause error, p Place) Issue {
	return Issue{Cause: cause, Place: p, NoteIndex: -1, InputIndex: -1}
}
func problem(cause error, p Place) error { return &Problems{Issues: []Issue{issue(cause, p)}} }
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
func must2[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}
	return a, b
}

type record struct {
	id      VoiceID
	bars    map[int][]*event.Event
	entries []Entry
}

// Track is an immutable record. Nil and zero values are unset. Copies share
// immutable records; methods may read and derive snapshots concurrently.
// The assigned live instrument is intentionally borrowed, never copied.
type Track struct {
	inst     *instrument.Instrument
	grid     measure.Grid
	records  []record
	fermatas map[int]bool
	clock    Clock
}

func New(inst *instrument.Instrument, g measure.Grid) (*Track, error) {
	if !inst.IsValid() {
		return nil, fmt.Errorf("%w: %w", ErrInvalidArgument, instrument.ErrUnset)
	}
	if err := checkGrid(g); err != nil {
		return nil, err
	}
	return &Track{inst: inst, grid: g, records: []record{{id: VoiceID{&voiceKey{allocated: true}}, bars: map[int][]*event.Event{}}}, fermatas: map[int]bool{}}, nil
}
func checkGrid(g measure.Grid) error {
	if !g.IsSet() {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, measure.ErrUnset)
	}
	n, finite, err := g.BarCount()
	if err != nil {
		return err
	}
	if !finite || n < 1 {
		return ErrInvalidArgument
	}
	_, _, err = g.End()
	return err
}
func (t *Track) IsSet() bool {
	return t != nil && t.inst != nil && t.grid.IsSet() && len(t.records) > 0
}
func (t *Track) Instrument() (*instrument.Instrument, error) {
	if !t.IsSet() {
		return nil, ErrUnset
	}
	return t.inst, nil
}
func (t *Track) Grid() (measure.Grid, error) {
	if !t.IsSet() {
		return measure.Grid{}, ErrUnset
	}
	return t.grid, nil
}
func (t *Track) Voices() ([]VoiceID, error) {
	if !t.IsSet() {
		return nil, ErrUnset
	}
	out := make([]VoiceID, len(t.records))
	for i, r := range t.records {
		out[i] = r.id
	}
	return out, nil
}
func (t *Track) voice(v VoiceID) (int, error) {
	if !t.IsSet() {
		return -1, ErrUnset
	}
	if v.key == nil || !v.key.allocated {
		return -1, ErrVoice
	}
	for i, r := range t.records {
		if v == r.id {
			return i, nil
		}
	}
	return -1, ErrVoice
}
func (t *Track) clone() *Track {
	c := *t
	c.records = slices.Clone(t.records)
	for i, r := range t.records {
		c.records[i].bars = make(map[int][]*event.Event, len(r.bars))
		for b, es := range r.bars {
			c.records[i].bars[b] = slices.Clone(es)
		}
		c.records[i].entries = slices.Clone(r.entries)
	}
	c.fermatas = make(map[int]bool, len(t.fermatas))
	for b, x := range t.fermatas {
		c.fermatas[b] = x
	}
	c.clock.BarHolds = slices.Clone(t.clock.BarHolds)
	return &c
}
func (t *Track) AddVoice() (*Track, VoiceID, error) {
	if !t.IsSet() {
		return nil, VoiceID{}, ErrUnset
	}
	c := t.clone()
	v := VoiceID{&voiceKey{allocated: true}}
	c.records = append(c.records, record{id: v, bars: map[int][]*event.Event{}})
	return c, v, nil
}
func sameEvent(a, b *event.Event) bool {
	if !a.Equal(b) {
		return false
	}
	an := a.MustContent().MustNotes()
	bn := b.MustContent().MustNotes()
	for i, n := range an {
		if !sameSource(n, bn[i]) {
			return false
		}
	}
	return true
}
func (t *Track) Equal(o *Track) bool {
	if !t.IsSet() || !o.IsSet() {
		return !t.IsSet() && !o.IsSet()
	}
	if !t.inst.Same(o.inst) || !t.grid.Equal(o.grid) || len(t.records) != len(o.records) || len(t.fermatas) != len(o.fermatas) || !t.clock.Map.Equal(o.clock.Map) || !slices.Equal(t.clock.BarHolds, o.clock.BarHolds) {
		return false
	}
	for b := range t.fermatas {
		if !o.fermatas[b] {
			return false
		}
	}
	for i, r := range t.records {
		q := o.records[i]
		if len(r.bars) != len(q.bars) {
			return false
		}
		for b, es := range r.bars {
			if !slices.EqualFunc(es, q.bars[b], sameEvent) {
				return false
			}
		}
	}
	return true
}
func (t *Track) bar(v VoiceID, b int) error {
	if _, err := t.voice(v); err != nil {
		return err
	}
	ok, err := t.grid.HasBar(b)
	if err != nil || !ok {
		return problem(ErrPlace, Place{v, b, -1})
	}
	return nil
}
func (t *Track) Entries(v VoiceID, b int) ([]Entry, error) {
	if err := t.bar(v, b); err != nil {
		return nil, err
	}
	i, _ := t.voice(v)
	var out []Entry
	for _, e := range t.records[i].entries {
		if e.Place.Bar == b {
			out = append(out, e)
		}
	}
	return out, nil
}
func (t *Track) position(p duration.Position, v VoiceID) error {
	end, _, err := t.grid.End()
	if err != nil {
		return err
	}
	if !p.IsSet() {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, duration.ErrUnset)
	}
	if p.Compare(end) > 0 {
		x := issue(ErrPlace, Place{Voice: v, Index: -1})
		x.Position = p
		return &Problems{[]Issue{x}}
	}
	return nil
}
func (t *Track) At(v VoiceID, p duration.Position) (Entry, bool, error) {
	i, err := t.voice(v)
	if err != nil {
		return Entry{}, false, err
	}
	if err = t.position(p, v); err != nil {
		return Entry{}, false, err
	}
	for _, e := range t.records[i].entries {
		if p.Compare(e.Time.Start) >= 0 && p.Compare(e.Time.End) < 0 {
			return e, true, nil
		}
	}
	return Entry{}, false, nil
}
func (t *Track) VoiceEnd(v VoiceID) (duration.Position, error) {
	i, err := t.voice(v)
	if err != nil {
		return duration.Position{}, err
	}
	end := duration.Origin()
	for _, e := range t.records[i].entries {
		if e.Time.End.Compare(end) > 0 {
			end = e.Time.End
		}
	}
	return end, nil
}
func (t *Track) FermataBars() ([]int, error) {
	if !t.IsSet() {
		return nil, ErrUnset
	}
	out := make([]int, 0, len(t.fermatas))
	for b := range t.fermatas {
		out = append(out, b)
	}
	slices.Sort(out)
	return out, nil
}
func (t *Track) Splice(v VoiceID, b, index, remove int, add ...*event.Event) (*Track, error) {
	if err := t.bar(v, b); err != nil {
		return nil, err
	}
	i, _ := t.voice(v)
	es := t.records[i].bars[b]
	p := Place{v, b, index}
	if index < 0 || index > len(es) {
		return nil, problem(ErrPlace, p)
	}
	if remove < 0 || remove > len(es)-index {
		return nil, problem(ErrInvalidArgument, p)
	}
	out := append(slices.Clone(es[:index]), add...)
	out = append(out, es[index+remove:]...)
	return t.Edit(Edit{Bars: []BarEdit{{v, b, out}}})
}
func (t *Track) Edit(e Edit) (*Track, error) {
	if !t.IsSet() {
		return nil, ErrUnset
	}
	c := t.clone()
	if e.Grid != nil {
		if err := checkGrid(*e.Grid); err != nil {
			return nil, err
		}
		c.grid = *e.Grid
	}
	seen := map[struct {
		v VoiceID
		b int
	}]bool{}
	for _, x := range e.Bars {
		if _, err := t.voice(x.Voice); err != nil {
			return nil, err
		}
		k := struct {
			v VoiceID
			b int
		}{x.Voice, x.Bar}
		if seen[k] {
			return nil, problem(ErrInvalidArgument, Place{x.Voice, x.Bar, -1})
		}
		seen[k] = true
		if err := c.bar(x.Voice, x.Bar); err != nil {
			if len(x.Events) > 0 || t.bar(x.Voice, x.Bar) != nil {
				return nil, err
			}
		}
		i, _ := c.voice(x.Voice)
		if len(x.Events) == 0 {
			delete(c.records[i].bars, x.Bar)
		} else {
			c.records[i].bars[x.Bar] = slices.Clone(x.Events)
		}
	}
	seenB := map[int]bool{}
	for _, x := range e.Barlines {
		p := Place{Bar: x.AfterBar, Index: -1}
		if seenB[x.AfterBar] {
			return nil, problem(ErrInvalidArgument, p)
		}
		seenB[x.AfterBar] = true
		if ok, _ := c.grid.HasBar(x.AfterBar); !ok {
			if x.Fermata || !t.fermatas[x.AfterBar] {
				return nil, problem(ErrPlace, p)
			}
		}
		if x.Fermata {
			c.fermatas[x.AfterBar] = true
		} else {
			delete(c.fermatas, x.AfterBar)
		}
	}
	if e.Clock != nil {
		c.clock = *e.Clock
		c.clock.BarHolds = slices.Clone(e.Clock.BarHolds)
		slices.SortFunc(c.clock.BarHolds, func(a, b BarHold) int { return compareInt(a.AfterBar, b.AfterBar) })
	}
	if err := c.validate(t); err != nil {
		return nil, err
	}
	return c, nil
}

func sameSource(a, b note.Note) bool {
	return (a.Source() == nil && b.Source() == nil) || a.SameSource(b)
}
