package score

import (
	"errors"
	"slices"
	"strings"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"
)

// MarkKind distinguishes tempo changes from extensions.
type MarkKind uint8

const (
	TempoMark     MarkKind = 1
	ExtensionMark MarkKind = 2
)

// MapMark retains the complete original tempo or extension record.
type MapMark struct {
	Kind      MarkKind
	Change    tempomap.Change
	Extension tempomap.Extension
}

// MarkMove records all bar bindings and the common destination of a mark.
type MarkMove struct {
	Mark     MapMark
	From, To duration.Position
	Bindings []Binding
}

// PartChange provides independently owned before and after tracks.
type PartChange struct {
	Part          PartID
	Before, After *track.Track
}

// MeterProposal is an atomic candidate; accepting it means using Score.
type MeterProposal struct {
	Score *Score
	Parts []PartChange
	Marks []MarkMove
}

func (m MapMark) position() duration.Position {
	var a tempomap.Address
	if m.Kind == TempoMark {
		a = m.Change.MustAddress()
	} else {
		a = m.Extension.MustAddress()
	}
	p, _ := a.MustPosition()
	return p
}
func (s *Score) marks() []MapMark {
	var out []MapMark
	for _, c := range s.clock.MustChanges() {
		out = append(out, MapMark{Kind: TempoMark, Change: c})
	}
	for _, e := range s.clock.MustExtensions() {
		out = append(out, MapMark{Kind: ExtensionMark, Extension: e})
	}
	slices.SortStableFunc(out, func(a, b MapMark) int {
		if c := a.position().Compare(b.position()); c != 0 {
			return c
		}
		return int(a.Kind) - int(b.Kind)
	})
	return out
}
func relocate(c tempomap.Change, a tempomap.Address) (tempomap.Change, error) {
	var out tempomap.Change
	var err error
	switch c.MustKind() {
	case tempomap.ChangeMain:
		t, _ := c.MustTempo()
		out, err = tempomap.NewMain(a, t)
	case tempomap.ChangeJump:
		t, _ := c.MustTempo()
		out, err = tempomap.NewJump(a, t)
	case tempomap.ChangeGradual:
		l, _ := c.MustLaw()
		out, err = tempomap.NewGradual(a, l)
	case tempomap.ChangeATempo:
		out, err = tempomap.NewATempo(a)
	case tempomap.ChangeTempoI:
		out, err = tempomap.NewTempoI(a)
	case tempomap.ChangeModulation:
		u, d, _ := c.MustModulation()
		out, err = tempomap.NewModulation(a, u, d)
	}
	if err != nil {
		return tempomap.Change{}, err
	}
	return out.WithSlot(c.MustSlot())
}
func mapRecords(cs []tempomap.Change, es []tempomap.Extension, grid ...measure.Grid) (tempomap.Map, error) {
	var records []string
	for _, c := range cs {
		records = append(records, c.String())
	}
	for _, e := range es {
		records = append(records, e.String())
	}
	m, err := tempomap.Parse("tempomap["+strings.Join(records, ";")+"]", grid...)
	return m, classify(err)
}
func bind(id PartID, g measure.Grid, m MapMark) (Binding, error) {
	p := m.position()
	a, err := g.Locate(p)
	b := Binding{Part: id, Address: a}
	if err == nil {
		return b, nil
	}
	end, _, endErr := g.End()
	if m.Kind == ExtensionMark && endErr == nil && p.Equal(end) {
		n, _ := g.MustBarCount()
		b.AfterBar = new(n)
		b.Address = measure.Address{}
		return b, nil
	}
	return b, errors.Join(ErrMapAddress, classify(err))
}
func bindingPosition(g measure.Grid, b Binding) (duration.Position, error) {
	if b.AfterBar != nil {
		return g.BarEnd(*b.AfterBar)
	}
	return g.PositionOf(b.Address)
}
func mapIssue(id PartID, m MapMark, b Binding, err error) Issue {
	return Issue{Cause: errors.Join(ErrMapAddress, classify(err)), Part: id, Position: m.position(), Mark: new(m), Bindings: []Binding{b}}
}

// ExtractWithMap copies the complete common map, including a terminal hold.
func (s *Score) ExtractWithMap(id PartID) (*track.Track, error) {
	i, err := s.index(id)
	if err != nil {
		return nil, err
	}
	p := s.parts[i]
	g := p.record.MustGrid()
	var cs []tempomap.Change
	var es []tempomap.Extension
	var holds []track.BarHold
	var issues []Issue
	for _, m := range s.marks() {
		b, err := bind(id, g, m)
		if err != nil {
			issues = append(issues, mapIssue(id, m, b, err))
			continue
		}
		if b.AfterBar != nil {
			holds = append(holds, track.BarHold{AfterBar: *b.AfterBar, Seconds: m.Extension.MustSeconds()})
			continue
		}
		a := tempomap.MustAtBar(b.Address)
		if m.Kind == TempoMark {
			c, err := relocate(m.Change, a)
			if err != nil {
				return nil, classify(err)
			}
			cs = append(cs, c)
		} else {
			es = append(es, tempomap.MustNewExtension(a, m.Extension.MustSeconds()))
		}
	}
	if len(issues) > 0 {
		return nil, &Problems{issues}
	}
	m, err := mapRecords(cs, es, g)
	if err != nil {
		return nil, err
	}
	t, err := detached(p.record, p.cells)
	if err != nil {
		return nil, err
	}
	out, err := t.Edit(track.Edit{Clock: &track.Clock{Map: m, BarHolds: holds}})
	if err != nil {
		return nil, classify(err)
	}
	return out, nil
}

// ChangeMeter proposes a strict bar-preserving change for every participation.
func (s *Score) ChangeMeter(first, last int, value meter.Meter) (MeterProposal, error) {
	if !s.IsSet() {
		return MeterProposal{}, ErrUnset
	}
	if first < 1 || last < first || len(s.parts) == 0 {
		return MeterProposal{}, ErrInvalidArgument
	}
	if !value.IsSet() {
		return MeterProposal{}, errors.Join(ErrInvalidArgument, meter.ErrUnset)
	}
	out := s.derive()
	grids := make([]measure.Grid, len(s.parts))
	var issues []Issue
	for i, p := range s.parts {
		g, err := p.record.MustGrid().WithMetersIn(first, last, value)
		if err != nil {
			issues = append(issues, trackIssues(p.id, err)...)
			continue
		}
		grids[i] = g
		t, err := p.record.Edit(track.Edit{Grid: &g})
		if err != nil {
			issues = append(issues, trackIssues(p.id, err)...)
			continue
		}
		out.parts[i].record = t
	}
	var moves []MarkMove
	var cs []tempomap.Change
	var es []tempomap.Extension
	for _, m := range s.marks() {
		move := MarkMove{Mark: m, From: m.position()}
		valid := true
		for i, p := range s.parts {
			if !grids[i].IsSet() {
				valid = false
				continue
			}
			b, err := bind(p.id, p.record.MustGrid(), m)
			if err != nil {
				issues = append(issues, mapIssue(p.id, m, b, err))
				valid = false
				continue
			}
			pos, err := bindingPosition(grids[i], b)
			if err != nil {
				issues = append(issues, mapIssue(p.id, m, b, err))
				valid = false
				continue
			}
			b.Position = pos
			move.Bindings = append(move.Bindings, b)
		}
		if !valid {
			continue
		}
		move.To = move.Bindings[0].Position
		conflict := false
		for _, b := range move.Bindings[1:] {
			if !b.Position.Equal(move.To) {
				conflict = true
			}
		}
		if conflict {
			issues = append(issues, Issue{Cause: ErrMapConflict, Position: move.From, Mark: new(m), Bindings: move.Bindings})
			continue
		}
		moves = append(moves, move)
		a := tempomap.MustAt(move.To)
		if m.Kind == TempoMark {
			c, err := relocate(m.Change, a)
			if err != nil {
				issues = append(issues, Issue{Cause: classify(err), Mark: new(m)})
				continue
			}
			cs = append(cs, c)
		} else {
			es = append(es, tempomap.MustNewExtension(a, m.Extension.MustSeconds()))
		}
	}
	if len(issues) > 0 {
		return MeterProposal{}, &Problems{issues}
	}
	// Distinct original records must not silently replace each other after moving.
	for i, a := range moves {
		for _, b := range moves[i+1:] {
			same := a.Mark.Kind == b.Mark.Kind && a.To.Equal(b.To)
			if same && a.Mark.Kind == TempoMark {
				same = a.Mark.Change.MustSlot() == b.Mark.Change.MustSlot()
			}
			if same {
				issues = append(issues, Issue{Cause: tempomap.ErrDuplicateAddress, Mark: new(a.Mark), Position: a.From}, Issue{Cause: tempomap.ErrDuplicateAddress, Mark: new(b.Mark), Position: b.From})
			}
		}
	}
	if len(issues) > 0 {
		return MeterProposal{}, &Problems{issues}
	}
	m, err := mapRecords(cs, es)
	if err != nil {
		return MeterProposal{}, &Problems{Issues: []Issue{{Cause: err}}}
	}
	out.clock = m
	proposal := MeterProposal{Score: out, Marks: moves}
	for i, p := range s.parts {
		before, err := detached(p.record, p.cells)
		if err != nil {
			return MeterProposal{}, err
		}
		after, err := detached(out.parts[i].record, p.cells)
		if err != nil {
			return MeterProposal{}, err
		}
		proposal.Parts = append(proposal.Parts, PartChange{p.id, before, after})
	}
	return proposal, nil
}
