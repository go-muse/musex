// Package score composes immutable tracks on one musical timeline and tempo map.
package score

import (
	"errors"
	"math"
	"slices"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"
)

// PartID identifies one participation. Its zero value is invalid.
type PartID struct{ key *partKey }
type partKey struct{ allocated bool }
type part struct {
	id     PartID
	record *track.Track
	cells  []track.BarEdit
}

// Score is an immutable ordered ensemble. Nil and zero values are unset.
// Copies share private immutable records and retain live instrument sources.
type Score struct {
	set   bool
	clock tempomap.Map
	parts []part
}

// Inclusion reports an explicit removal of a local clock.
type Inclusion struct {
	Score     *Score
	Part      PartID
	Discarded *track.Clock
}

func checkMap(m tempomap.Map) error {
	if !m.IsSet() {
		return errors.Join(ErrInvalidArgument, tempomap.ErrUnset)
	}
	if m.MustAddressing() != tempomap.PositionAddressing {
		return ErrInvalidArgument
	}
	_, err := m.Validate()
	return classify(err)
}

// New creates a set, empty ensemble with a positional clock.
func New(clock tempomap.Map) (*Score, error) {
	if err := checkMap(clock); err != nil {
		return nil, err
	}
	return &Score{set: true, clock: clock}, nil
}

// IsSet reports whether the score has been initialized.
func (s *Score) IsSet() bool { return s != nil && s.set }

// Equal compares the ordered musical records, ignoring participation keys.
func (s *Score) Equal(o *Score) bool {
	if !s.IsSet() || !o.IsSet() {
		return s.IsSet() == o.IsSet()
	}
	if !s.clock.Equal(o.clock) || len(s.parts) != len(o.parts) {
		return false
	}
	for i, p := range s.parts {
		if !p.record.Equal(o.parts[i].record) {
			return false
		}
	}
	return true
}
func (s *Score) index(id PartID) (int, error) {
	if !s.IsSet() {
		return 0, ErrUnset
	}
	for i, p := range s.parts {
		if p.id == id {
			return i, nil
		}
	}
	return 0, ErrPart
}
func (s *Score) derive() *Score { v := *s; v.parts = slices.Clone(s.parts); return &v }

// Parts returns participation keys in score order.
func (s *Score) Parts() ([]PartID, error) {
	if !s.IsSet() {
		return nil, ErrUnset
	}
	out := make([]PartID, len(s.parts))
	for i, p := range s.parts {
		out[i] = p.id
	}
	return out, nil
}

// Map returns the immutable common positional map.
func (s *Score) Map() (tempomap.Map, error) {
	if !s.IsSet() {
		return tempomap.Map{}, ErrUnset
	}
	return s.clock, nil
}

// WithMap derives a score with a validated common map.
func (s *Score) WithMap(m tempomap.Map) (*Score, error) {
	if !s.IsSet() {
		return nil, ErrUnset
	}
	if err := checkMap(m); err != nil {
		return nil, err
	}
	v := s.derive()
	v.clock = m
	return v, nil
}

// End returns the latest finite grid end, or the origin for an empty score.
func (s *Score) End() (duration.Position, error) {
	if !s.IsSet() {
		return duration.Position{}, ErrUnset
	}
	end := duration.Origin()
	for _, p := range s.parts {
		v, _, err := p.record.MustGrid().End()
		if err != nil {
			return duration.Position{}, classify(err)
		}
		if v.Compare(end) > 0 {
			end = v
		}
	}
	return end, nil
}
func cloneEvent(e *event.Event) *event.Event {
	if e == nil {
		return nil
	}
	v := *e
	return &v
}
func cloneEntry(e track.Entry) track.Entry { e.Event = cloneEvent(e.Event); return e }
func cloneCells(cells []track.BarEdit) []track.BarEdit {
	out := make([]track.BarEdit, len(cells))
	for i, c := range cells {
		out[i] = c
		out[i].Events = make([]*event.Event, len(c.Events))
		for j, e := range c.Events {
			out[i].Events[j] = cloneEvent(e)
		}
	}
	return out
}
func detached(t *track.Track, cells []track.BarEdit) (*track.Track, error) {
	v, err := t.Edit(track.Edit{Bars: cloneCells(cells)})
	if err != nil {
		return nil, classify(err)
	}
	c := *v
	return &c, nil
}
func capture(t *track.Track) (*track.Track, []track.BarEdit, error) {
	g := t.MustGrid()
	a, err := g.Locate(duration.Origin())
	if err != nil {
		return nil, nil, classify(err)
	}
	var cells []track.BarEdit
	for _, v := range t.MustVoices() {
		pr, err := t.WriteEvents(v, a, nil, track.WriteOptions{Mode: track.ReplaceVoice})
		if err != nil {
			return nil, nil, classify(err)
		}
		byBar := map[int][]track.Entry{}
		for _, c := range pr.Changes {
			for _, e := range c.From {
				byBar[e.Place.Bar] = append(byBar[e.Place.Bar], e)
			}
		}
		bars := make([]int, 0, len(byBar))
		for b := range byBar {
			bars = append(bars, b)
		}
		slices.Sort(bars)
		for _, b := range bars {
			entries := byBar[b]
			slices.SortFunc(entries, func(a, b track.Entry) int { return a.Place.Index - b.Place.Index })
			es := make([]*event.Event, len(entries))
			for i, e := range entries {
				es[i] = cloneEvent(e.Event)
			}
			cells = append(cells, track.BarEdit{Voice: v, Bar: b, Events: es})
		}
	}
	v, err := t.Edit(track.Edit{Bars: cells})
	if err != nil {
		return nil, nil, classify(err)
	}
	copyTrack := *v
	return &copyTrack, cells, nil
}

// Part extracts an independently owned track without a clock.
func (s *Score) Part(id PartID) (*track.Track, error) {
	i, err := s.index(id)
	if err != nil {
		return nil, err
	}
	return detached(s.parts[i].record, s.parts[i].cells)
}
func (s *Score) include(id PartID, t *track.Track, replace, discard bool) (Inclusion, error) {
	if !s.IsSet() {
		return Inclusion{}, ErrUnset
	}
	idx := 0
	var err error
	if replace {
		idx, err = s.index(id)
		if err != nil {
			return Inclusion{}, err
		}
	}
	if !t.IsSet() {
		return Inclusion{}, errors.Join(ErrInvalidArgument, track.ErrUnset)
	}
	clock, has, err := t.Clock()
	if err != nil {
		return Inclusion{}, err
	}
	if has && !discard {
		return Inclusion{}, &Problems{Issues: []Issue{{Cause: ErrLocalClock, Part: id}}}
	}
	var removed *track.Clock
	if has {
		removed = &clock
		t, err = t.Edit(track.Edit{Clock: &track.Clock{}})
		if err != nil {
			return Inclusion{}, classify(err)
		}
	}
	if replace {
		if err := checkStarts(id, s.parts[idx].record.MustGrid(), t.MustGrid()); err != nil {
			return Inclusion{}, err
		}
	}
	owned, cells, err := capture(t)
	if err != nil {
		return Inclusion{}, err
	}
	out := s.derive()
	if !replace {
		id = PartID{key: &partKey{allocated: true}}
		out.parts = append(out.parts, part{id, owned, cells})
	} else {
		out.parts[idx] = part{id, owned, cells}
	}
	return Inclusion{out, id, removed}, nil
}

// Add appends a clockless participation with a fresh key.
func (s *Score) Add(t *track.Track) (*Score, PartID, error) {
	v, err := s.include(PartID{}, t, false, false)
	return v.Score, v.Part, err
}

// AddWithoutClock explicitly discards and reports the incoming clock.
func (s *Score) AddWithoutClock(t *track.Track) (Inclusion, error) {
	return s.include(PartID{}, t, false, true)
}

// Replace replaces a participation while preserving its key and common bar starts.
func (s *Score) Replace(id PartID, t *track.Track) (*Score, error) {
	v, err := s.include(id, t, true, false)
	return v.Score, err
}

// ReplaceWithoutClock replaces a participation and explicitly discards its clock.
func (s *Score) ReplaceWithoutClock(id PartID, t *track.Track) (Inclusion, error) {
	return s.include(id, t, true, true)
}

// Remove derives a score without the selected participation.
func (s *Score) Remove(id PartID) (*Score, error) {
	i, err := s.index(id)
	if err != nil {
		return nil, err
	}
	v := s.derive()
	v.parts = slices.Delete(v.parts, i, i+1)
	return v, nil
}

// Reorder requires an exact permutation of all participation keys.
func (s *Score) Reorder(ids ...PartID) (*Score, error) {
	if !s.IsSet() {
		return nil, ErrUnset
	}
	if len(ids) != len(s.parts) {
		return nil, ErrInvalidArgument
	}
	seen := map[PartID]bool{}
	v := s.derive()
	for j, id := range ids {
		i, err := s.index(id)
		if err != nil {
			return nil, err
		}
		if seen[id] {
			return nil, ErrInvalidArgument
		}
		seen[id] = true
		v.parts[j] = s.parts[i]
	}
	return v, nil
}
func gridChecks(g measure.Grid) []int {
	n, _ := g.MustBarCount()
	points := []int{g.MustFirstBar(), 1, n}
	for _, m := range g.MustMarks() {
		points = append(points, m.Bar)
		if m.Bar < math.MaxInt {
			points = append(points, m.Bar+1)
		}
	}
	for _, l := range g.MustLengths() {
		points = append(points, l.Bar)
		if l.Bar < math.MaxInt {
			points = append(points, l.Bar+1)
		}
	}
	slices.Sort(points)
	return slices.Compact(points)
}
func checkStarts(id PartID, a, b measure.Grid) error {
	points := append(gridChecks(a), gridChecks(b)...)
	slices.Sort(points)
	for _, k := range slices.Compact(points) {
		if !a.MustHasBar(k) || !b.MustHasBar(k) {
			continue
		}
		x, err := a.BarStart(k)
		if err != nil {
			return classify(err)
		}
		y, err := b.BarStart(k)
		if err != nil {
			return classify(err)
		}
		if !x.Equal(y) {
			return &Problems{Issues: []Issue{{Cause: ErrGridShift, Part: id, Place: track.Place{Bar: k, Index: -1}, Position: x, Proposed: y}}}
		}
	}
	return nil
}
