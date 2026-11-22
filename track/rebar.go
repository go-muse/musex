package track

import (
	"errors"
	"strings"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempomap"
)

// Rebar preserves absolute musical positions, rather than bar numbers.
func (t *Track) Rebar(grid measure.Grid, g Groups) (Proposal, error) {
	if !t.IsSet() {
		return Proposal{}, ErrUnset
	}
	if !validGroups(g) {
		return Proposal{}, ErrInvalidArgument
	}
	if err := checkGrid(grid); err != nil {
		return Proposal{}, err
	}
	c := t.clone()
	c.grid = grid
	c.fermatas = map[int]bool{}
	c.clock = Clock{}
	for i := range c.records {
		c.records[i].bars = map[int][]*event.Event{}
		c.records[i].entries = nil
	}
	p := Proposal{}
	for _, r := range t.records {
		items := itemsOf(r.entries)
		for i := range items {
			it := &items[i]
			it.bar = -1
			if it.e.MustDuration().Kind() != event.DurationMeasures {
				continue
			}
			// A contextual sign stays one record when its endpoints coincide with
			// complete bars of a single meter run. No list of per-bar pieces is needed.
			startAddress, startErr := grid.Locate(it.start)
			finish, err := it.start.Add(it.length)
			if err != nil {
				return Proposal{}, err
			}
			last, lastErr := barEndingAt(grid, finish)
			same := startErr == nil && lastErr == nil && startAddress.MustOffset().Equal(duration.Origin()) && last >= startAddress.MustBar()
			if same {
				first := startAddress.MustBar()
				runs, err := grid.MeterRuns(first, last)
				if err != nil {
					return Proposal{}, err
				}
				same = len(runs) == 1
				if same {
					e, err := event.New(event.RestContent(), event.MustMeasureDuration(last-first+1), it.e.MustMarks())
					if err != nil {
						return Proposal{}, err
					}
					it.e = e
				}
			}
			if !same {
				e, err := rest(it.length, it.e.MustMarks())
				if err != nil {
					return Proposal{}, err
				}
				it.e = e
			}
		}
		next, err := c.layout(r.id, items, g, r.entries)
		if err != nil {
			return Proposal{}, err
		}
		c = next.Track
		p.Changes = append(p.Changes, next.Changes...)
	}
	for _, b := range t.MustFermataBars() {
		old, err := t.grid.BarEnd(b)
		if err != nil {
			return Proposal{}, err
		}
		nb, err := barEndingAt(grid, old)
		if err != nil {
			x := issue(errors.Join(ErrWouldLoseMark, err), Place{Bar: b, Index: -1})
			x.Position = old
			x.AfterBar = new(b)
			x.Kind = BarlineFermata
			return Proposal{}, &Problems{[]Issue{x}}
		}
		c.fermatas[nb] = true
		if b != nb {
			p.Marks = append(p.Marks, MarkMove{From: old, To: old, OldAfterBar: new(b), NewAfterBar: new(nb), Kind: BarlineFermata})
		}
	}
	clock, moves, err := t.rebarClock(grid)
	if err != nil {
		return Proposal{}, err
	}
	c.clock = clock
	p.Marks = append(p.Marks, moves...)
	if err = c.validate(); err != nil {
		return Proposal{}, err
	}
	p.Track = c
	return p, nil
}
func barEndingAt(g measure.Grid, p duration.Position) (int, error) {
	end, _, err := g.End()
	if err != nil {
		return 0, err
	}
	if p.Equal(end) {
		n, _, err := g.BarCount()
		return n, err
	}
	a, err := g.Locate(p)
	if err != nil {
		return 0, err
	}
	if !a.MustOffset().Equal(duration.Origin()) {
		return 0, ErrPlace
	}
	b := a.MustBar() - 1
	if ok, _ := g.HasBar(b); !ok {
		return 0, ErrPlace
	}
	return b, nil
}
func moveChange(c tempomap.Change, a tempomap.Address) (tempomap.Change, error) {
	var x tempomap.Change
	var err error
	switch c.MustKind() {
	case tempomap.ChangeMain:
		val, _ := c.MustTempo()
		x, err = tempomap.NewMain(a, val)
	case tempomap.ChangeJump:
		val, _ := c.MustTempo()
		x, err = tempomap.NewJump(a, val)
	case tempomap.ChangeGradual:
		law, _ := c.MustLaw()
		x, err = tempomap.NewGradual(a, law)
	case tempomap.ChangeATempo:
		x, err = tempomap.NewATempo(a)
	case tempomap.ChangeTempoI:
		x, err = tempomap.NewTempoI(a)
	case tempomap.ChangeModulation:
		unit, old, _ := c.MustModulation()
		x, err = tempomap.NewModulation(a, unit, old)
	default:
		return tempomap.Change{}, ErrInvalidArgument
	}
	if err != nil {
		return tempomap.Change{}, err
	}
	return x.WithSlot(c.MustSlot())
}
func (t *Track) rebarClock(g measure.Grid) (Clock, []MarkMove, error) {
	if !t.clock.Map.IsSet() {
		return Clock{}, nil, nil
	}
	posmap, err := t.clock.Map.ToPositions(t.grid)
	if err != nil {
		return Clock{}, nil, err
	}
	out := Clock{}
	var moves []MarkMove
	var text []string
	oldChanges := t.clock.Map.MustChanges()
	for i, ch := range posmap.MustChanges() {
		pos, _, _ := ch.MustAddress().Position()
		a, err := g.Locate(pos)
		old, _ := oldChanges[i].MustAddress().MustBar()
		if err != nil {
			x := issue(errors.Join(ErrClockAddress, err), Place{Bar: old.MustBar(), Index: -1})
			x.Address = old
			x.Position = pos
			x.Kind = TempoChange
			x.Slot = ch.MustSlot()
			return Clock{}, nil, &Problems{[]Issue{x}}
		}
		newCh, err := moveChange(ch, tempomap.MustAtBar(a))
		if err != nil {
			return Clock{}, nil, err
		}
		text = append(text, newCh.String())
		if !a.Equal(old) {
			moves = append(moves, MarkMove{From: pos, To: pos, OldAddress: old, NewAddress: a, Kind: TempoChange, Slot: ch.MustSlot()})
		}
	}
	end, _, err := g.End()
	if err != nil {
		return Clock{}, nil, err
	}
	type extensionRecord struct {
		p                duration.Position
		old              measure.Address
		after            *int
		secondsExtension tempomap.Extension
	}
	var exs []extensionRecord
	oldExt := t.clock.Map.MustExtensions()
	for i, ex := range posmap.MustExtensions() {
		pos, _, _ := ex.MustAddress().Position()
		a, _ := oldExt[i].MustAddress().MustBar()
		exs = append(exs, extensionRecord{p: pos, old: a, secondsExtension: ex})
	}
	for _, h := range t.clock.BarHolds {
		pos, err := t.grid.BarEnd(h.AfterBar)
		if err != nil {
			return Clock{}, nil, err
		}
		exs = append(exs, extensionRecord{p: pos, after: new(h.AfterBar), secondsExtension: tempomap.MustNewExtension(tempomap.MustAt(pos), h.Seconds)})
	}
	for _, ex := range exs {
		move := MarkMove{From: ex.p, To: ex.p, OldAddress: ex.old, OldAfterBar: ex.after, Kind: Extension}
		if ex.p.Equal(end) {
			b, _, _ := g.BarCount()
			out.BarHolds = append(out.BarHolds, BarHold{b, ex.secondsExtension.MustSeconds()})
			move.NewAfterBar = new(b)
		} else {
			a, err := g.Locate(ex.p)
			if err != nil {
				x := issue(errors.Join(ErrClockAddress, err), Place{Index: -1})
				x.Position = ex.p
				x.Address = ex.old
				x.AfterBar = ex.after
				x.Kind = Extension
				return Clock{}, nil, &Problems{[]Issue{x}}
			}
			text = append(text, tempomap.MustNewExtension(tempomap.MustAtBar(a), ex.secondsExtension.MustSeconds()).String())
			move.NewAddress = a
		}
		same := move.OldAddress.Equal(move.NewAddress) && ((move.OldAfterBar == nil && move.NewAfterBar == nil) || (move.OldAfterBar != nil && move.NewAfterBar != nil && *move.OldAfterBar == *move.NewAfterBar))
		if !same {
			moves = append(moves, move)
		}
	}
	out.Map, err = tempomap.Parse("tempomap["+strings.Join(text, ";")+"]", g)
	if err != nil {
		return Clock{}, nil, errors.Join(ErrClockAddress, err)
	}
	return out, moves, nil
}
