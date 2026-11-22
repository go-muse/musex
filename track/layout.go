package track

import (
	"errors"
	"slices"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/measure"
)

type layoutItem struct {
	e      *event.Event
	start  duration.Position
	length duration.Duration
	from   []Entry
	input  []int
	bar    int
}
type fragment struct {
	e     *event.Event
	start duration.Position
	bar   int
	owner int
}

func validGroups(g Groups) bool { return g == KeepGroups || g == WithoutGroups }
func validOptions(o WriteOptions) bool {
	return validGroups(o.Groups) && (o.Mode == Vacant || o.Mode == ReplaceVoice)
}
func rest(d duration.Duration, m event.Marks) (*event.Event, error) {
	ed, err := event.ValueDuration(d.MustFraction())
	if err != nil {
		return nil, err
	}
	return event.New(event.RestContent(), ed, m)
}

// notationPartError retains the input segment that failed exact notation.
// The segment index is independent of how many signs earlier segments used.
type notationPartError struct {
	index  int
	amount duration.Duration
	cause  error
}

func (e *notationPartError) Error() string { return e.cause.Error() }
func (e *notationPartError) Unwrap() error { return e.cause }

func lengthsToEvents(e *event.Event, ds []duration.Duration, g Groups, force bool) ([]*event.Event, error) {
	ed := e.MustDuration()
	var forms []event.Duration
	for partIndex, d := range ds {
		if ed.Kind() == event.DurationValue && !force {
			forms = append(forms, event.MustValueDuration(d.MustFraction()))
			continue
		}
		w := duration.MustNewWritten(duration.Quarter, 0)
		chain := w.MustChain()
		if ed.Kind() == event.DurationWritten && g == KeepGroups {
			chain = ed.MustWritten().MustChain()
		}
		single, err := d.ToWritten(chain...)
		if err == nil {
			forms = append(forms, event.MustWrittenDuration(single))
			continue
		}
		ws, err := d.Notate(chain...)
		if err != nil {
			return nil, &notationPartError{index: partIndex, amount: d, cause: errors.Join(ErrNotation, err)}
		}
		for _, w := range ws {
			forms = append(forms, event.MustWrittenDuration(w))
		}
	}
	ns := e.MustContent().MustNotes()
	out := make([]*event.Event, 0, len(forms))
	for i, d := range forms {
		m := e.MustMarks()
		if i > 0 {
			m.Arpeggio = false
		}
		if i < len(forms)-1 {
			m.Fermata = false
			m.Ties = nil
			for ni := range ns {
				m.Ties = append(m.Ties, ni)
			}
		}
		x, err := event.New(e.MustContent(), d, m)
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, nil
}
func (t *Track) lengthAt(e *event.Event, b int) (duration.Duration, error) {
	d := e.MustDuration()
	if d.Kind() != event.DurationMeasures {
		return d.Value()
	}
	n := d.MustMeasureCount()
	maxbar, _, _ := t.grid.BarCount()
	if n-1 > maxbar-b {
		return duration.Duration{}, ErrNeedBars
	}
	return t.grid.RangeLength(b, b+n-1)
}
func (t *Track) fragments(v VoiceID, items []layoutItem, g Groups) ([]fragment, error) {
	var out []fragment
	gridEnd, _, _ := t.grid.End()
	for i, it := range items {
		m := it.e.MustMarks()
		p := Place{Voice: v, Index: -1}
		x := issue(ErrNeedBars, p)
		x.Position = it.start
		if len(it.input) > 0 {
			x.InputIndex = it.input[0]
		}
		if len(it.from) > 0 {
			x.Place = it.from[0].Place
		}
		if m.Grace {
			if it.start.Compare(gridEnd) > 0 {
				return nil, &Problems{[]Issue{x}}
			}
			b := it.bar
			if b < 0 {
				if it.start.Equal(gridEnd) {
					b, _, _ = t.grid.BarCount()
				} else {
					a, err := t.grid.Locate(it.start)
					if err != nil {
						return nil, err
					}
					b = a.MustBar()
				}
			}
			out = append(out, fragment{it.e, it.start, b, i})
			continue
		}
		end, err := it.start.Add(it.length)
		if err != nil {
			return nil, err
		}
		if end.Compare(gridEnd) > 0 {
			if it.start.Compare(gridEnd) < 0 {
				x.Amount, err = gridEnd.DurationTo(end)
				x.Position = gridEnd
			} else {
				x.Amount = it.length
			}
			if err != nil {
				return nil, err
			}
			if it.e.MustDuration().Kind() == event.DurationMeasures {
				x.BarsNeeded = it.e.MustDuration().MustMeasureCount()
				n, _, _ := t.grid.BarCount()
				if it.bar >= 0 {
					x.BarsNeeded = max(1, x.BarsNeeded-(n-it.bar+1))
				}
			}
			return nil, &Problems{[]Issue{x}}
		}
		if it.e.MustDuration().Kind() == event.DurationMeasures {
			a, err := t.grid.Locate(it.start)
			if err != nil {
				return nil, err
			}
			if !a.MustOffset().Equal(duration.Origin()) {
				return nil, problem(ErrMeasureRest, Place{v, a.MustBar(), -1})
			}
			count := it.e.MustDuration().MustMeasureCount()
			n, _, _ := t.grid.BarCount()
			if count-1 > n-a.MustBar() {
				x.BarsNeeded = count - 1 - (n - a.MustBar())
				return nil, &Problems{[]Issue{x}}
			}
			last := a.MustBar() + count - 1
			runs, err := t.grid.MeterRuns(a.MustBar(), last)
			if err != nil {
				return nil, err
			}
			if len(runs) != 1 {
				return nil, problem(ErrMeasureRest, Place{v, a.MustBar(), -1})
			}
			out = append(out, fragment{it.e, it.start, a.MustBar(), i})
			continue
		}
		parts, err := t.grid.Split(it.start, it.length)
		if err != nil {
			return nil, err
		}
		if len(parts) == 0 {
			return nil, ErrInvalidArgument
		}
		if len(parts) == 1 {
			out = append(out, fragment{it.e, it.start, parts[0].Address.MustBar(), i})
			continue
		}
		ds := make([]duration.Duration, len(parts))
		for k, p := range parts {
			ds[k] = p.Duration
		}
		es, err := lengthsToEvents(it.e, ds, g, it.e.Kind() != event.KindRest)
		if err != nil {
			x.Cause = err
			failed := 0
			if detail, ok := errors.AsType[*notationPartError](err); ok {
				failed = detail.index
			}
			x.Amount = parts[failed].Duration
			x.Address = parts[failed].Address
			x.Position = parts[failed].Start
			return nil, &Problems{[]Issue{x}}
		}
		pos := it.start
		for _, e := range es {
			a, err := t.grid.Locate(pos)
			if err != nil {
				return nil, err
			}
			out = append(out, fragment{e, pos, a.MustBar(), i})
			pos, err = pos.Add(e.MustDuration().MustValue())
			if err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}
func intervalsOverlap(a, b Span) bool {
	return a.Start.Compare(b.End) < 0 && b.Start.Compare(a.End) < 0
}
func (t *Track) layout(v VoiceID, items []layoutItem, g Groups, old []Entry) (Proposal, error) {
	fs, err := t.fragments(v, items, g)
	if err != nil {
		return Proposal{}, err
	}
	slices.SortStableFunc(fs, func(a, b fragment) int {
		if a.bar != b.bar {
			return compareInt(a.bar, b.bar)
		}
		return a.start.Compare(b.start)
	})
	c := t.clone()
	ri, _ := c.voice(v)
	c.records[ri].bars = map[int][]*event.Event{}
	to := make([][]Entry, len(items))
	var added []Change
	bar := -1
	var pos duration.Position
	for _, f := range fs {
		if f.bar != bar {
			bar = f.bar
			pos, err = c.grid.BarStart(bar)
			if err != nil {
				return Proposal{}, err
			}
		}
		if f.start.Compare(pos) < 0 {
			return Proposal{}, problem(ErrOccupied, Place{v, bar, len(c.records[ri].bars[bar])})
		}
		if f.start.Compare(pos) > 0 {
			d, err := pos.DurationTo(f.start)
			if err != nil {
				return Proposal{}, err
			}
			e, err := rest(d, event.Marks{})
			if err != nil {
				return Proposal{}, err
			}
			p := Place{v, bar, len(c.records[ri].bars[bar])}
			c.records[ri].bars[bar] = append(c.records[ri].bars[bar], e)
			added = append(added, Change{To: []Entry{{p, e, Span{pos, f.start}}}})
			pos = f.start
		}
		p := Place{v, bar, len(c.records[ri].bars[bar])}
		end := f.start
		if !f.e.MustMarks().Grace {
			d, err := c.lengthAt(f.e, bar)
			if err != nil {
				return Proposal{}, err
			}
			end, err = f.start.Add(d)
			if err != nil {
				return Proposal{}, err
			}
			pos = end
		}
		c.records[ri].bars[bar] = append(c.records[ri].bars[bar], f.e)
		to[f.owner] = append(to[f.owner], Entry{p, f.e, Span{f.start, end}})
	}
	if err = c.validate(); err != nil {
		return Proposal{}, err
	}
	proposal := Proposal{Track: c}
	used := map[Place]bool{}
	for i, it := range items {
		for _, e := range it.from {
			used[e.Place] = true
		}
		if len(it.from) == 1 && len(to[i]) == 1 && sameEntry(it.from[0], to[i][0]) && len(it.input) == 0 {
			continue
		}
		proposal.Changes = append(proposal.Changes, Change{From: slices.Clone(it.from), To: slices.Clone(to[i]), Input: slices.Clone(it.input)})
	}
	for _, e := range old {
		if !used[e.Place] {
			proposal.Changes = append(proposal.Changes, Change{From: []Entry{e}})
		}
	}
	proposal.Changes = append(proposal.Changes, added...)
	sortChanges(proposal.Changes)
	return proposal, nil
}
func sameEntry(a, b Entry) bool {
	return a.Place == b.Place && a.Time.Start.Equal(b.Time.Start) && a.Time.End.Equal(b.Time.End) && sameEvent(a.Event, b.Event)
}
func sortChanges(cs []Change) {
	slices.SortStableFunc(cs, func(a, b Change) int {
		key := func(c Change) Entry {
			if len(c.From) > 0 {
				return c.From[0]
			}
			return c.To[0]
		}
		x, y := key(a), key(b)
		if x.Place.Bar != y.Place.Bar {
			return compareInt(x.Place.Bar, y.Place.Bar)
		}
		return compareInt(x.Place.Index, y.Place.Index)
	})
}
func itemsOf(es []Entry) []layoutItem {
	out := make([]layoutItem, 0, len(es))
	for _, e := range es {
		d := duration.Duration{}
		if e.Time.End.Compare(e.Time.Start) > 0 {
			d = e.Time.Start.MustDurationTo(e.Time.End)
		}
		out = append(out, layoutItem{e: e.Event, start: e.Time.Start, length: d, from: []Entry{e}, bar: e.Place.Bar})
	}
	return out
}

// WriteEvents lays out a sequence, preserving the old snapshot on every error.
func (t *Track) WriteEvents(v VoiceID, at measure.Address, events []*event.Event, o WriteOptions) (Proposal, error) {
	ri, err := t.voice(v)
	if err != nil {
		return Proposal{}, err
	}
	if !validOptions(o) {
		return Proposal{}, ErrInvalidArgument
	}
	if !at.IsSet() {
		return Proposal{}, errors.Join(ErrInvalidArgument, measure.ErrUnset)
	}
	if ok, _ := t.grid.Contains(at); !ok {
		x := issue(ErrPlace, Place{v, at.MustBar(), -1})
		x.Address = at
		return Proposal{}, &Problems{[]Issue{x}}
	}
	start, err := t.grid.PositionOf(at)
	if err != nil {
		return Proposal{}, err
	}
	old := t.records[ri].entries
	items := []layoutItem{}
	if o.Mode == Vacant {
		items = itemsOf(old)
	}
	pos := start
	for i, e := range events {
		p := Place{v, at.MustBar(), -1}
		if !e.IsSet() {
			return Proposal{}, problem(errors.Join(ErrInvalidArgument, event.ErrUnset), p)
		}
		it := layoutItem{e: e, start: pos, input: []int{i}, bar: -1}
		if !e.MustMarks().Grace {
			a, err := t.grid.Locate(pos)
			if err != nil {
				x := issue(ErrNeedBars, p)
				x.InputIndex = i
				x.Position = pos
				if e.MustDuration().Kind() != event.DurationMeasures {
					x.Amount = e.MustDuration().MustValue()
				} else {
					x.BarsNeeded = e.MustDuration().MustMeasureCount()
				}
				return Proposal{}, &Problems{[]Issue{x}}
			}
			if e.MustDuration().Kind() == event.DurationMeasures && !a.MustOffset().Equal(duration.Origin()) {
				return Proposal{}, problem(ErrMeasureRest, p)
			}
			it.bar = a.MustBar()
			it.length, err = t.lengthAt(e, it.bar)
			if err != nil {
				x := issue(err, p)
				x.InputIndex = i
				x.Position = pos
				if errors.Is(err, ErrNeedBars) {
					n, _, _ := t.grid.BarCount()
					x.BarsNeeded = e.MustDuration().MustMeasureCount() - (n - it.bar + 1)
				}
				return Proposal{}, &Problems{[]Issue{x}}
			}
			pos, err = pos.Add(it.length)
			if err != nil {
				return Proposal{}, err
			}
			gridEnd, _, err := t.grid.End()
			if err != nil {
				return Proposal{}, err
			}
			if pos.Compare(gridEnd) > 0 {
				missing, err := gridEnd.DurationTo(pos)
				if err != nil {
					return Proposal{}, err
				}
				x := issue(ErrNeedBars, Place{v, it.bar, -1})
				x.InputIndex = i
				x.Position = gridEnd
				x.Amount = missing
				return Proposal{}, &Problems{[]Issue{x}}
			}
		}
		if o.Mode == Vacant {
			span := Span{it.start, pos}
			for _, ex := range old {
				grace := e.MustMarks().Grace || ex.Event.MustMarks().Grace
				if intervalsOverlap(span, ex.Time) || (grace && it.start.Equal(ex.Time.Start)) {
					x := issue(ErrOccupied, ex.Place)
					x.InputIndex = i
					return Proposal{}, &Problems{[]Issue{x}}
				}
			}
		}
		items = append(items, it)
	}
	return t.layout(v, items, o.Groups, old)
}
