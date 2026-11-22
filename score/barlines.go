package score

import (
	"errors"
	"math/big"
	"slices"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/track"
)

// Boundary identifies the bars adjacent to a real grid boundary.
type Boundary struct {
	Part                 PartID
	PreviousBar, NextBar *int
}

// Barline is a common boundary in score order.
type Barline struct {
	Position   duration.Position
	Boundaries []Boundary
}

func boundaryAt(id PartID, g measure.Grid, p duration.Position) (Boundary, error) {
	end, _, err := g.End()
	if err != nil {
		return Boundary{}, classify(err)
	}
	n, _ := g.MustBarCount()
	if p.Equal(end) {
		return Boundary{Part: id, PreviousBar: new(n)}, nil
	}
	a, err := g.Locate(p)
	if err != nil {
		return Boundary{}, classify(err)
	}
	b := a.MustBar()
	r := Boundary{Part: id, NextBar: new(b)}
	if b > g.MustFirstBar() {
		r.PreviousBar = new(b - 1)
	}
	return r, nil
}
func nextBoundary(g measure.Grid, p duration.Position, strict bool) (duration.Position, bool, error) {
	end, _, err := g.End()
	if err != nil {
		return duration.Position{}, false, classify(err)
	}
	if p.Compare(end) > 0 || (strict && p.Equal(end)) {
		return duration.Position{}, false, nil
	}
	if p.Equal(end) {
		return end, true, nil
	}
	a, err := g.Locate(p)
	if err != nil {
		return duration.Position{}, false, classify(err)
	}
	b, err := g.Bar(a.MustBar())
	if err != nil {
		return duration.Position{}, false, classify(err)
	}
	if !strict && b.Start.Equal(p) {
		return p, true, nil
	}
	return b.End, true, nil
}

// CommonBarlines intersects finite grids over a closed score interval.
func (s *Score) CommonBarlines(from, to duration.Position, ids ...PartID) ([]Barline, error) {
	if !s.IsSet() {
		return nil, ErrUnset
	}
	if !from.IsSet() || !to.IsSet() {
		return nil, errors.Join(ErrInvalidArgument, duration.ErrUnset)
	}
	if from.Compare(to) > 0 || len(ids) < 2 {
		return nil, ErrInvalidArgument
	}
	seen := map[PartID]bool{}
	for _, id := range ids {
		if _, err := s.index(id); err != nil {
			return nil, err
		}
		if seen[id] {
			return nil, ErrInvalidArgument
		}
		seen[id] = true
	}
	if err := s.checkPosition(to); err != nil {
		return nil, err
	}
	var selected []part
	for _, p := range s.parts {
		if seen[p.id] {
			selected = append(selected, p)
		}
	}
	candidate := from
	var out []Barline
	for candidate.Compare(to) <= 0 {
		aligned := true
		for _, p := range selected {
			x, ok, err := nextBoundary(p.record.MustGrid(), candidate, false)
			if err != nil {
				return nil, err
			}
			if !ok {
				return out, nil
			}
			if x.Compare(candidate) > 0 {
				candidate = x
				aligned = false
			}
		}
		if !aligned {
			continue
		}
		line := Barline{Position: candidate}
		for _, p := range selected {
			b, err := boundaryAt(p.id, p.record.MustGrid(), candidate)
			if err != nil {
				return nil, err
			}
			line.Boundaries = append(line.Boundaries, b)
		}
		out = append(out, line)
		x, ok, err := nextBoundary(selected[0].record.MustGrid(), candidate, true)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		candidate = x
	}
	return out, nil
}
func periodLength(p part) (duration.Duration, error) {
	g := p.record.MustGrid()
	fail := func(b int) (duration.Duration, error) {
		return duration.Duration{}, &Problems{Issues: []Issue{{Cause: ErrPeriod, Part: p.id, Place: track.Place{Bar: b, Index: -1}}}}
	}
	if g.MustHasPickup() {
		return fail(0)
	}
	n, _ := g.MustBarCount()
	length := g.MustBarLength(1)
	points := gridChecks(g)
	slices.Sort(points)
	for _, b := range points {
		if b < 1 || b > n {
			continue
		}
		bar, err := g.Bar(b)
		if err != nil {
			return duration.Duration{}, classify(err)
		}
		if !bar.Length.Equal(bar.Meter.MustDuration()) || !bar.Length.Equal(length) {
			return fail(b)
		}
	}
	return length, nil
}

// BarlinePeriod returns the exact mathematical period of constant full bars.
func (s *Score) BarlinePeriod(a, b PartID) (duration.Duration, error) {
	i, err := s.index(a)
	if err != nil {
		return duration.Duration{}, err
	}
	j, err := s.index(b)
	if err != nil {
		return duration.Duration{}, err
	}
	if a == b {
		return duration.Duration{}, ErrInvalidArgument
	}
	x, err := periodLength(s.parts[i])
	if err != nil {
		return duration.Duration{}, err
	}
	y, err := periodLength(s.parts[j])
	if err != nil {
		return duration.Duration{}, err
	}
	an, ad, _ := x.MustFraction().ReducedBigPair()
	bn, bd, _ := y.MustFraction().ReducedBigPair()
	g := new(big.Int).GCD(nil, nil, an, bn)
	num := new(big.Int).Mul(new(big.Int).Quo(an, g), bn)
	den := new(big.Int).GCD(nil, nil, ad, bd)
	f, err := fraction.FromBigIntReduced(num, den)
	if err != nil {
		return duration.Duration{}, classify(err)
	}
	out, err := duration.FromFraction(f)
	if err != nil {
		return duration.Duration{}, classify(err)
	}
	return out, nil
}
