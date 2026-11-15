package measure

import (
	"math/big"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
)

// Split divides the span of positions [start, start + value) at bar lines,
// the same as SplitBy(start, value, LevelBars). The parts follow each other
// without gaps, the first starts at start, their lengths sum to value, and
// each lies in one bar. Each part has the address and the position of its
// start and its length. Splitting is mechanical: it does not claim a rule of
// notation.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument
// together with duration.ErrUnset for an unset start, then for an unset
// value; ErrOutsideGrid if start + value is after the end of a finite grid;
// ErrTooMany if the span occupies more than MaxParts bars; ErrOutOfRange if
// start + value is after the end of bar math.MaxInt of an open grid; then,
// in the order of parts, ErrTooMany for part MaxParts + 1 or ErrOutOfRange
// together with fraction.ErrOutOfRange for a part whose start, offset, or
// length is outside the range of common/fraction.
//
// Example: in 4/4, the span of 9/4 from 0/1 gives
// [1/1@1:0/1 1/1@2:0/1 1/4@3:0/1].
func (g Grid) Split(start duration.Position, value duration.Duration) ([]Part, error) {
	return g.split("Grid.Split", start, value, LevelBars)
}

// MustSplit is like Split but panics with the error of Split.
func (g Grid) MustSplit(start duration.Position, value duration.Duration) []Part {
	return must(g.Split(start, value))
}

// SplitBy divides the span of positions [start, start + value) at the
// boundaries of level, as Split does at bar lines. LevelGroups, LevelBeats,
// and LevelUnits also split at the starts of the groups, beats, or units of
// the effective meter of every spanned bar: an element at the place s in the
// meter gives the boundary at the offset s − Shift of the bar if it is
// inside the bar. In an extended bar the offset equal to the size of the
// meter is a boundary too, and beyond it LevelUnits continues with steps of
// the last unit of the meter; groups and beats do not continue.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument
// together with duration.ErrUnset for an unset start, then for an unset
// value; ErrInvalidArgument for an invalid level; ErrOutsideGrid if
// start + value is after the end of a finite grid; ErrGroupingRequired
// together with meter.ErrGroupingRequired for LevelGroups or LevelBeats if
// the meter of a spanned bar has no grouping; then the errors of Split from
// ErrTooMany on.
//
// Example: in 4/4, the span of 3/8 from 3/8 by groups gives
// [1/8@1:3/8 1/4@1:1/2].
func (g Grid) SplitBy(start duration.Position, value duration.Duration, level Level) ([]Part, error) {
	return g.split("Grid.SplitBy", start, value, level)
}

// MustSplitBy is like SplitBy but panics with the error of SplitBy.
func (g Grid) MustSplitBy(start duration.Position, value duration.Duration, level Level) []Part {
	return must(g.SplitBy(start, value, level))
}

// split implements Split and SplitBy.
func (g Grid) split(op string, start duration.Position, value duration.Duration, level Level) ([]Part, error) {
	if !g.IsSet() {
		return nil, errUnsetGrid(op)
	}
	if !start.IsSet() {
		return nil, errUnsetTime(op, "start")
	}
	if !value.IsSet() {
		return nil, errUnsetTime(op, "value")
	}
	if !level.IsValid() {
		return nil, errInvalid(op, "invalid level")
	}
	s := ratOfPosition(start)
	e := ratAdd(s, ratOfDuration(value))
	if g.count > 0 && e.Cmp(g.end) > 0 {
		return nil, errSimple(ErrOutsideGrid, op, "the span ends after the end of the grid")
	}
	if level == LevelGroups || level == LevelBeats {
		for i := g.segIndexAt(s); i < len(g.segs) && g.segs[i].start.Cmp(e) < 0; i++ {
			if ok, _ := g.segs[i].meter.HasGrouping(); !ok {
				return nil, errGrouping(op)
			}
		}
	}
	firstBig := g.barBig(s)
	n := new(big.Int).Sub(g.lastBarBig(e), firstBig)
	if n.Cmp(big.NewInt(MaxParts-1)) > 0 {
		return nil, errSimple(ErrTooMany, op, "the span occupies more than MaxParts bars")
	}
	if err := g.checkSpanEnd(op, e); err != nil {
		return nil, err
	}
	b := splitter{op: op, level: level, parts: make([]Part, 0, min(MaxParts, int(n.Int64())+1))}
	first, last := int(firstBig.Int64()), int(g.lastBarBig(e).Int64())
	i := g.segIndex(first)
	barStart := g.barStartRat(first)
	for k := first; ; k++ {
		for k > g.segs[i].last {
			i++
		}
		seg := &g.segs[i]
		barEnd := ratAdd(barStart, seg.length)
		from, to := ratZero, seg.length
		if s.Cmp(barStart) > 0 {
			from = ratSub(s, barStart)
		}
		if e.Cmp(barEnd) < 0 {
			to = ratSub(e, barStart)
		}
		if err := b.bar(i, seg, k, barStart, from, to); err != nil {
			return nil, err
		}
		if k == last {
			break
		}
		barStart = barEnd
	}
	return b.parts, nil
}

// splitter builds the parts of a split.
type splitter struct {
	op    string
	level Level
	parts []Part
	// seg is the index of the segment of offsets, or -1 before the first.
	seg     int
	offsets []*big.Rat // boundaries of the elements inside a bar of seg
	started bool
}

// elementOffsets returns the sorted offsets inside a bar of seg of the starts
// of the elements of the level, without the continuation of an extended bar.
func elementOffsets(seg *segment, level Level) []*big.Rat {
	var spans []meter.Span
	switch level {
	case LevelGroups:
		spans, _ = seg.meter.Groups()
	case LevelBeats:
		spans, _ = seg.meter.Beats()
	case LevelUnits:
		spans, _ = seg.meter.Units()
	default:
		return nil
	}
	out := make([]*big.Rat, 0, len(spans)+1)
	for _, sp := range spans {
		o := ratSub(ratOfPosition(sp.Start), seg.shift)
		if o.Sign() > 0 && o.Cmp(seg.length) < 0 {
			out = append(out, o)
		}
	}
	if seg.length.Cmp(seg.size) > 0 {
		out = append(out, seg.size)
	}
	return out
}

// bar adds the parts of [from, to) of bar k of segment index i, which starts
// at barStart.
func (b *splitter) bar(i int, seg *segment, k int, barStart, from, to *big.Rat) error {
	if !b.started || b.seg != i {
		b.started = true
		b.seg = i
		b.offsets = elementOffsets(seg, b.level)
	}
	cur := from
	for _, o := range b.offsets {
		if o.Cmp(cur) <= 0 {
			continue
		}
		if o.Cmp(to) >= 0 {
			return b.add(k, barStart, cur, to)
		}
		if err := b.add(k, barStart, cur, o); err != nil {
			return err
		}
		cur = o
	}
	if b.level == LevelUnits && seg.length.Cmp(seg.size) > 0 {
		// Continue with steps of the last unit beyond the size of the meter.
		spans, _ := seg.meter.Units()
		step := ratOfDuration(spans[len(spans)-1].Duration)
		j := big.NewInt(1)
		if cur.Cmp(seg.size) >= 0 {
			j = floorRat(ratQuo(ratSub(cur, seg.size), step))
			j.Add(j, big.NewInt(1))
		}
		for {
			o := ratAdd(seg.size, ratMulBigInt(step, j))
			if o.Cmp(to) >= 0 {
				break
			}
			if err := b.add(k, barStart, cur, o); err != nil {
				return err
			}
			cur = o
			j = new(big.Int).Add(j, big.NewInt(1))
		}
	}
	return b.add(k, barStart, cur, to)
}

// add adds the part [from, to) of bar k, which starts at barStart.
func (b *splitter) add(k int, barStart, from, to *big.Rat) error {
	if len(b.parts) == MaxParts {
		return errSimple(ErrTooMany, b.op, "more than MaxParts parts")
	}
	startP, err := toPosition(b.op, ratAdd(barStart, from))
	if err != nil {
		return err
	}
	offset, err := toPosition(b.op, from)
	if err != nil {
		return err
	}
	d, err := toDuration(b.op, ratSub(to, from))
	if err != nil {
		return err
	}
	b.parts = append(b.parts, Part{Address: Address{bar: k, offset: offset}, Start: startP, Duration: d})
	return nil
}
