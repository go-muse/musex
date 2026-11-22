package track

import (
	"errors"
	"slices"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
)

func (t *Track) element(p Place) (Entry, error) {
	es, err := t.Entries(p.Voice, p.Bar)
	if err != nil {
		return Entry{}, err
	}
	if p.Index < 0 || p.Index >= len(es) {
		return Entry{}, problem(ErrPlace, p)
	}
	return es[p.Index], nil
}
func (t *Track) rangeBars(v VoiceID, first, last int) error {
	if err := t.bar(v, first); err != nil {
		return err
	}
	if err := t.bar(v, last); err != nil {
		return err
	}
	if first > last {
		return problem(ErrInvalidArgument, Place{v, first, -1})
	}
	return nil
}
func (t *Track) spliceProposal(v VoiceID, b, index, count int, es []*event.Event) (Proposal, error) {
	before, err := t.Entries(v, b)
	if err != nil {
		return Proposal{}, err
	}
	c, err := t.Splice(v, b, index, count, es...)
	if err != nil {
		return Proposal{}, err
	}
	after, err := c.Entries(v, b)
	if err != nil {
		return Proposal{}, err
	}
	p := Proposal{Track: c, Changes: []Change{{From: slices.Clone(before[index : index+count]), To: slices.Clone(after[index : index+len(es)])}}}
	delta := len(es) - count
	for i := index + count; i < len(before); i++ {
		j := i + delta
		if !sameEntry(before[i], after[j]) {
			p.Changes = append(p.Changes, Change{From: []Entry{before[i]}, To: []Entry{after[j]}})
		}
	}
	return p, nil
}

// Split divides an ordinary event at a strictly interior offset.
func (t *Track) Split(at Place, offset duration.Duration, g Groups) (Proposal, error) {
	en, err := t.element(at)
	if err != nil {
		return Proposal{}, err
	}
	if !validGroups(g) {
		return Proposal{}, ErrInvalidArgument
	}
	if !offset.IsSet() {
		return Proposal{}, problem(errors.Join(ErrInvalidArgument, duration.ErrUnset), at)
	}
	if en.Event.MustMarks().Grace || en.Event.MustDuration().Kind() == event.DurationMeasures {
		return Proposal{}, problem(ErrInvalidArgument, at)
	}
	length := en.Event.MustDuration().MustValue()
	if offset.Compare(length) >= 0 {
		return Proposal{}, problem(ErrInvalidArgument, at)
	}
	tail, err := length.Sub(offset)
	if err != nil {
		return Proposal{}, err
	}
	es, err := lengthsToEvents(en.Event, []duration.Duration{offset, tail}, g, en.Event.Kind() != event.KindRest)
	if err != nil {
		x := issue(err, at)
		x.Amount = offset
		if detail, ok := errors.AsType[*notationPartError](err); ok {
			x.Amount = detail.amount
		}
		return Proposal{}, &Problems{[]Issue{x}}
	}
	return t.spliceProposal(at.Voice, at.Bar, at.Index, 1, es)
}

// Merge joins compatible sustained records, retaining the attack and final signs.
func (t *Track) Merge(v VoiceID, b, index, count int, g Groups) (Proposal, error) {
	p := Place{v, b, index}
	en, err := t.element(p)
	if err != nil {
		return Proposal{}, err
	}
	if !validGroups(g) {
		return Proposal{}, ErrInvalidArgument
	}
	es, _ := t.Entries(v, b)
	if count < 2 || count > len(es)-index {
		return Proposal{}, problem(ErrInvalidArgument, p)
	}
	selected := es[index : index+count]
	d0 := en.Event.MustDuration()
	if d0.Kind() == event.DurationMeasures || en.Event.MustMarks().Grace {
		return Proposal{}, problem(ErrWouldLoseMark, p)
	}
	ns := en.Event.MustContent().MustNotes()
	last := selected[len(selected)-1]
	length, err := en.Time.Start.DurationTo(last.Time.End)
	if err != nil {
		return Proposal{}, err
	}
	for i, entry := range selected {
		e := entry.Event
		m := e.MustMarks()
		d := e.MustDuration()
		fail := false
		if e.Kind() != en.Event.Kind() || d.Kind() != d0.Kind() || m.Grace {
			fail = true
		}
		other := e.MustContent().MustNotes()
		if len(ns) != len(other) {
			fail = true
		} else {
			for k, n := range ns {
				if !n.Equal(other[k]) || !sameSource(n, other[k]) {
					fail = true
				}
			}
		}
		if i > 0 && m.Arpeggio {
			fail = true
		}
		if i < len(selected)-1 {
			if m.Fermata {
				fail = true
			}
			if len(m.Ties) != len(ns) {
				fail = true
			} else {
				for k, ni := range m.Ties {
					if k != ni {
						fail = true
					}
				}
			}
		}
		if d.Kind() == event.DurationWritten && d0.Kind() == event.DurationWritten && g == KeepGroups {
			if !slices.Equal(d.MustWritten().MustChain(), d0.MustWritten().MustChain()) {
				fail = true
			}
		}
		if fail {
			return Proposal{}, problem(ErrWouldLoseMark, entry.Place)
		}
	}
	var d event.Duration
	if d0.Kind() == event.DurationValue {
		d = event.MustValueDuration(length.MustFraction())
	} else {
		w := d0.MustWritten()
		chain := w.MustChain()
		if g == WithoutGroups {
			chain = nil
		}
		w, err = length.ToWritten(chain...)
		if err != nil {
			return Proposal{}, problem(errors.Join(ErrNotation, err), p)
		}
		d = event.MustWrittenDuration(w)
	}
	marks := en.Event.MustMarks()
	marks.Ties = last.Event.MustMarks().Ties
	marks.Fermata = last.Event.MustMarks().Fermata
	e, err := event.New(en.Event.MustContent(), d, marks)
	if err != nil {
		return Proposal{}, err
	}
	return t.spliceProposal(v, b, index, count, []*event.Event{e})
}

// PackRests packs explicit one-bar rests, separately for each meter run.
func (t *Track) PackRests(v VoiceID, first, last int) (Proposal, error) {
	if err := t.rangeBars(v, first, last); err != nil {
		return Proposal{}, err
	}
	runs, err := t.grid.MeterRuns(first, last)
	if err != nil {
		return Proposal{}, err
	}
	var edits []BarEdit
	var groups [][]Entry
	for _, run := range runs {
		var entries []Entry
		for b := run.First; ; b++ {
			es, err := t.Entries(v, b)
			if err != nil {
				return Proposal{}, err
			}
			if len(es) != 1 || es[0].Event.Kind() != event.KindRest || es[0].Event.MustDuration().Kind() != event.DurationMeasures || es[0].Event.MustDuration().MustMeasureCount() != 1 {
				return Proposal{}, problem(ErrMeasureRest, Place{v, b, -1})
			}
			e := es[0]
			if b < run.Last && e.Event.MustMarks().Fermata {
				x := issue(ErrWouldLoseMark, e.Place)
				x.Position = e.Time.End
				return Proposal{}, &Problems{[]Issue{x}}
			}
			entries = append(entries, e)
			if b == run.Last {
				break
			}
		}
		lastEntry := entries[len(entries)-1]
		e, err := event.New(event.RestContent(), event.MustMeasureDuration(run.Last-run.First+1), lastEntry.Event.MustMarks())
		if err != nil {
			return Proposal{}, err
		}
		edits = append(edits, BarEdit{v, run.First, []*event.Event{e}})
		for _, entry := range entries[1:] {
			edits = append(edits, BarEdit{v, entry.Place.Bar, nil})
		}
		groups = append(groups, entries)
	}
	c, err := t.Edit(Edit{Bars: edits})
	if err != nil {
		return Proposal{}, err
	}
	p := Proposal{Track: c}
	for i, run := range runs {
		after, err := c.Entries(v, run.First)
		if err != nil {
			return Proposal{}, err
		}
		if len(groups[i]) == 1 && sameEntry(groups[i][0], after[0]) {
			continue
		}
		p.Changes = append(p.Changes, Change{From: groups[i], To: after})
	}
	return p, nil
}

// UnpackRest restores one explicit rest in each covered bar.
func (t *Track) UnpackRest(at Place) (Proposal, error) {
	en, err := t.element(at)
	if err != nil {
		return Proposal{}, err
	}
	d := en.Event.MustDuration()
	if d.Kind() != event.DurationMeasures {
		return Proposal{}, problem(ErrMeasureRest, at)
	}
	count := d.MustMeasureCount()
	var edits []BarEdit
	for i := range count {
		m := event.Marks{}
		if i == count-1 {
			m = en.Event.MustMarks()
		}
		e, err := event.New(event.RestContent(), event.MustMeasureDuration(1), m)
		if err != nil {
			return Proposal{}, err
		}
		edits = append(edits, BarEdit{at.Voice, at.Bar + i, []*event.Event{e}})
	}
	c, err := t.Edit(Edit{Bars: edits})
	if err != nil {
		return Proposal{}, err
	}
	var to []Entry
	for _, e := range edits {
		es, err := c.Entries(e.Voice, e.Bar)
		if err != nil {
			return Proposal{}, err
		}
		to = append(to, es...)
	}
	return Proposal{Track: c, Changes: []Change{{From: []Entry{en}, To: to}}}, nil
}

// Quantize rounds shared boundaries once, keeping actual barlines fixed.
func (t *Track) Quantize(v VoiceID, first, last int, step duration.Duration, mode rounding.Mode, g Groups) (Proposal, error) {
	if err := t.rangeBars(v, first, last); err != nil {
		return Proposal{}, err
	}
	if !step.IsSet() || !mode.IsValid() || !validGroups(g) {
		return Proposal{}, ErrInvalidArgument
	}
	ri, _ := t.voice(v)
	old := t.records[ri].entries
	items := itemsOf(old)
	cache := map[duration.Position]duration.Position{}
	holds := map[duration.Position]Issue{}
	if t.clock.Map.IsSet() {
		for _, ex := range t.clock.Map.MustExtensions() {
			address, _ := ex.MustAddress().MustBar()
			pos, err := t.grid.PositionOf(address)
			if err != nil {
				return Proposal{}, err
			}
			x := issue(ErrHoldConflict, Place{Index: -1})
			x.Position = pos
			x.Address = address
			x.Kind = Extension
			holds[pos] = x
		}
		for _, h := range t.clock.BarHolds {
			pos, err := t.grid.BarEnd(h.AfterBar)
			if err != nil {
				return Proposal{}, err
			}
			x := issue(ErrHoldConflict, Place{Index: -1})
			x.Position = pos
			x.AfterBar = new(h.AfterBar)
			x.Kind = Extension
			holds[pos] = x
		}
	}
	for i, en := range old {
		if en.Place.Bar < first || en.Place.Bar > last || en.Event.MustDuration().Kind() == event.DurationMeasures {
			continue
		}
		bar, err := t.grid.Bar(en.Place.Bar)
		if err != nil {
			return Proposal{}, err
		}
		round := func(p duration.Position) (duration.Position, error) {
			if p.Equal(bar.Start) || p.Equal(bar.End) {
				return p, nil
			}
			if q, ok := cache[p]; ok {
				return q, nil
			}
			q, err := p.Quantize(step.MustFraction(), mode)
			if err != nil {
				return duration.Position{}, err
			}
			cache[p] = q
			return q, nil
		}
		start, err := round(en.Time.Start)
		if err != nil {
			return Proposal{}, err
		}
		end, err := round(en.Time.End)
		if err != nil {
			return Proposal{}, err
		}
		if start.Compare(bar.Start) < 0 || end.Compare(bar.End) > 0 {
			x := issue(ErrCrossBar, en.Place)
			x.Position = en.Time.End
			return Proposal{}, &Problems{[]Issue{x}}
		}
		if en.Event.MustMarks().Grace {
			items[i].start = start
			continue
		}
		if end.Compare(start) <= 0 {
			x := issue(ErrCollapsed, en.Place)
			x.Position = en.Time.Start
			return Proposal{}, &Problems{[]Issue{x}}
		}
		if hold, ok := holds[en.Time.End]; en.Event.MustMarks().Fermata && ok && !end.Equal(en.Time.End) {
			x := hold
			x.Place = en.Place
			return Proposal{}, &Problems{[]Issue{x}}
		}
		length, err := start.DurationTo(end)
		if err != nil {
			return Proposal{}, err
		}
		items[i].start = start
		items[i].length = length
		if !length.Equal(en.Event.MustDuration().MustValue()) {
			parts, err := lengthsToEvents(en.Event, []duration.Duration{length}, g, false)
			if err != nil {
				return Proposal{}, problem(err, en.Place)
			}
			if len(parts) == 1 {
				items[i].e = parts[0]
			} else {
				// Expand notation while retaining one source relation for the whole group.
				items[i].e = parts[0]
				items[i].length = parts[0].MustDuration().MustValue()
				pos := start
				for k, e := range parts {
					if k == 0 {
						pos, err = pos.Add(items[i].length)
						if err != nil {
							return Proposal{}, err
						}
						continue
					}
					l := e.MustDuration().MustValue()
					items = append(items, layoutItem{e: e, start: pos, length: l, from: []Entry{en}, bar: en.Place.Bar})
					pos, err = pos.Add(l)
					if err != nil {
						return Proposal{}, err
					}
				}
			}
		}
	}
	p, err := t.layout(v, items, g, old)
	if err != nil {
		return Proposal{}, err
	}
	return coalesceChanges(p), nil
}
func coalesceChanges(p Proposal) Proposal {
	seen := map[Place]int{}
	var out []Change
	for _, c := range p.Changes {
		if len(c.From) == 1 {
			key := c.From[0].Place
			if j, ok := seen[key]; ok {
				out[j].To = append(out[j].To, c.To...)
				continue
			}
			seen[key] = len(out)
		}
		out = append(out, c)
	}
	p.Changes = out
	return p
}
