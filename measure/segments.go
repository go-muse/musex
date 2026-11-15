package measure

import (
	"math"
	"math/big"
	"slices"
	"sort"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
)

// segment is a run of bars first…last with the same meter, length, and
// shift. The pickup, a bar with an explicit length, and a bar with a mark
// start their own segments, so only the first bar of a segment can have a
// mark, a meter change, or an explicit length.
type segment struct {
	first, last int
	meter       meter.Meter
	size        *big.Rat // size of the meter
	length      *big.Rat // length of every bar of the segment
	lengthD     duration.Duration
	shift       *big.Rat // offset of the bar start in a full bar of the meter
	shiftP      duration.Position
	shiftErr    bool     // the shift is outside the range of common/fraction
	start       *big.Rat // start of bar first
	align       Alignment
	pickup      bool
	explicit    bool
	marked      bool // bar first has a mark
	change      bool // the meter changes by record at bar first
}

// lastBar returns the number of the last bar of the grid: N of a finite grid
// or math.MaxInt of an open grid.
func (g Grid) lastBar() int {
	if g.count > 0 {
		return g.count
	}
	return math.MaxInt
}

// markIndexAt returns the index of the last mark at a bar not after k; k is
// at least 0, and the mark of bar 1 is used for the pickup.
func markIndexAt(marks []Mark, k int) int {
	i := sort.Search(len(marks), func(i int) bool { return marks[i].Bar > k })
	if i == 0 {
		return 0
	}
	return i - 1
}

// findMark returns the index of the mark at bar k, or -1.
func findMark(marks []Mark, k int) int {
	i := sort.Search(len(marks), func(i int) bool { return marks[i].Bar >= k })
	if i < len(marks) && marks[i].Bar == k {
		return i
	}
	return -1
}

// findLength returns the index of the explicit length of bar k, or -1.
func findLength(lengths []ExplicitLength, k int) int {
	i := sort.Search(len(lengths), func(i int) bool { return lengths[i].Bar >= k })
	if i < len(lengths) && lengths[i].Bar == k {
		return i
	}
	return -1
}

// build returns the grid of the record pickup, marks, lengths, and count,
// with its segments. marks are sorted by bar, the first at bar 1; lengths are
// sorted by bar; records after bar count of a finite grid are not given.
//
// It returns ErrLongerThanMeter for the first violation of the record: the
// pickup longer than the size of the meter of bar 1, then explicit lengths
// aligned to the end longer than the size of the meter of their bar, in the
// order of bars.
func build(op string, pickup duration.Duration, marks []Mark, lengths []ExplicitLength, count int) (Grid, error) {
	if pickup.IsSet() && ratOfDuration(pickup).Cmp(meterSize(marks[0].Meter)) > 0 {
		return Grid{}, errSimple(ErrLongerThanMeter, op, "the pickup is longer than the meter of bar 1")
	}
	for _, l := range lengths {
		if l.Alignment != AlignEnd {
			continue
		}
		m := marks[markIndexAt(marks, l.Bar)].Meter
		if ratOfDuration(l.Duration).Cmp(meterSize(m)) > 0 {
			return Grid{}, errSimple(ErrLongerThanMeter, op,
				"an explicit length aligned to the end is longer than the meter of its bar")
		}
	}
	g := Grid{pickup: pickup, marks: marks, lengths: lengths, count: count}
	g.segs, g.end = g.segments()
	return g, nil
}

// segments computes the segments of the record of g and the end of its last
// bar: bar N of a finite grid or bar math.MaxInt of an open grid.
func (g Grid) segments() ([]segment, *big.Rat) {
	last := g.lastBar()
	segs := make([]segment, 0, 2*(len(g.marks)+len(g.lengths))+2)
	pos := new(big.Rat)
	if g.pickup.IsSet() {
		m := g.marks[0].Meter
		size := meterSize(m)
		length := ratOfDuration(g.pickup)
		segs = append(segs, newSegment(segment{
			first: 0, last: 0, meter: m, size: size, length: length, lengthD: g.pickup,
			shift: ratSub(size, length), start: pos, align: AlignEnd, pickup: true,
		}))
		pos = ratAdd(pos, length)
	}
	// Boundaries: bars of marks and of explicit lengths, and the bars after
	// explicit lengths.
	bounds := make([]int, 0, len(g.marks)+2*len(g.lengths))
	for _, mk := range g.marks {
		bounds = append(bounds, mk.Bar)
	}
	for _, l := range g.lengths {
		bounds = append(bounds, l.Bar)
		if l.Bar < last {
			bounds = append(bounds, l.Bar+1)
		}
	}
	slices.Sort(bounds)
	bounds = slices.Compact(bounds)
	mi, li := 0, 0
	cur := g.marks[0].Meter
	for bi, first := range bounds {
		segLast := last
		if bi+1 < len(bounds) {
			segLast = bounds[bi+1] - 1
		}
		marked, change := false, false
		if mi < len(g.marks) && g.marks[mi].Bar == first {
			marked = true
			change = mi > 0 && !g.marks[mi].Meter.Equal(cur)
			cur = g.marks[mi].Meter
			mi++
		}
		size := meterSize(cur)
		s := segment{
			first: first, last: segLast, meter: cur, size: size, start: pos,
			align: AlignStart, marked: marked, change: change,
		}
		if li < len(g.lengths) && g.lengths[li].Bar == first {
			l := g.lengths[li]
			li++
			s.length = ratOfDuration(l.Duration)
			s.lengthD = l.Duration
			s.explicit = true
			s.align = l.Alignment
			if l.Alignment == AlignEnd {
				s.shift = ratSub(size, s.length)
			}
		} else {
			s.length = size
			d, _ := cur.Duration()
			s.lengthD = d
		}
		if s.shift == nil {
			s.shift = new(big.Rat)
		}
		segs = append(segs, newSegment(s))
		n := new(big.Int).Sub(big.NewInt(int64(segLast)), big.NewInt(int64(first)))
		n.Add(n, big.NewInt(1))
		pos = ratAdd(pos, ratMulBigInt(s.length, n))
	}
	return segs, pos
}

// newSegment completes s with the converted shift.
func newSegment(s segment) segment {
	p, err := toPosition("", s.shift)
	s.shiftP = p
	s.shiftErr = err != nil
	return s
}

// segIndex returns the index of the segment that contains bar k; bar k exists.
func (g Grid) segIndex(k int) int {
	i := sort.Search(len(g.segs), func(i int) bool { return g.segs[i].first > k })
	return i - 1
}

// segIndexAt returns the index of the last segment whose start is not after
// p.
func (g Grid) segIndexAt(p *big.Rat) int {
	// The first segment starts at 0, which is not after p.
	return sort.Search(len(g.segs), func(i int) bool { return g.segs[i].start.Cmp(p) > 0 }) - 1
}

// segIndexBefore returns the index of the last segment whose start is before
// p; p is positive.
func (g Grid) segIndexBefore(p *big.Rat) int {
	// The first segment starts at 0, which is before p.
	return sort.Search(len(g.segs), func(i int) bool { return g.segs[i].start.Cmp(p) >= 0 }) - 1
}

// hasBar reports whether the grid has bar k.
func (g Grid) hasBar(k int) bool {
	switch {
	case k < 0:
		return false
	case k == 0:
		return g.pickup.IsSet()
	default:
		return g.count == 0 || k <= g.count
	}
}

// barStartRat returns the start of bar k, which exists.
func (g Grid) barStartRat(k int) *big.Rat {
	s := &g.segs[g.segIndex(k)]
	return ratAdd(s.start, ratMulInt(s.length, k-s.first))
}

// barEndRat returns the end of bar k, which exists.
func (g Grid) barEndRat(k int) *big.Rat {
	s := &g.segs[g.segIndex(k)]
	return ratAdd(s.start, ratMulInt(s.length, k-s.first+1))
}

// barBig returns the number of the bar that contains the non-negative
// position p, with the last segment continued without end; the result can
// exceed math.MaxInt.
func (g Grid) barBig(p *big.Rat) *big.Int {
	s := &g.segs[g.segIndexAt(p)]
	q := floorRat(ratQuo(ratSub(p, s.start), s.length))
	return q.Add(q, big.NewInt(int64(s.first)))
}

// lastBarBig returns the number of the bar that contains the moment just
// before the positive position e, with the last segment continued without
// end: a span that ends at e on a bar line does not occupy the next bar.
func (g Grid) lastBarBig(e *big.Rat) *big.Int {
	s := &g.segs[g.segIndexBefore(e)]
	q := ceilRat(ratQuo(ratSub(e, s.start), s.length))
	q.Add(q, big.NewInt(int64(s.first)-1))
	return q
}

// locateRat returns the segment index, the bar, and the offset of the
// position p, which is before the end of the last bar of the grid.
func (g Grid) locateRat(p *big.Rat) (int, int, *big.Rat) {
	i := g.segIndexAt(p)
	s := &g.segs[i]
	rel := ratSub(p, s.start)
	q := floorRat(ratQuo(rel, s.length))
	n := int(q.Int64())
	offset := ratSub(rel, ratMulBigInt(s.length, q))
	return i, s.first + n, offset
}

// checkPosition reports the failure for a position p at or after the end of
// the grid: ErrOutsideGrid for a finite grid when p is not before its end (or
// after its end if atEnd is allowed), ErrOutOfRange for an open grid when p
// is not before the end of bar math.MaxInt.
func (g Grid) checkPosition(op string, p *big.Rat, allowEnd bool) error {
	c := p.Cmp(g.end)
	if g.count > 0 {
		if c > 0 || (c == 0 && !allowEnd) {
			return errSimple(ErrOutsideGrid, op, "the position is not before the end of the grid")
		}
		return nil
	}
	if c >= 0 {
		return errSimple(ErrOutOfRange, op, "the position is not before the end of bar math.MaxInt")
	}
	return nil
}

// checkSpanEnd reports the failure for the end e of a span of positions:
// ErrOutsideGrid if e is after the end of a finite grid, ErrOutOfRange if e
// is after the end of bar math.MaxInt of an open grid. An end exactly at the
// end is allowed.
func (g Grid) checkSpanEnd(op string, e *big.Rat) error {
	if e.Cmp(g.end) <= 0 {
		return nil
	}
	if g.count > 0 {
		return errSimple(ErrOutsideGrid, op, "the span ends after the end of the grid")
	}
	return errSimple(ErrOutOfRange, op, "the span ends after the end of bar math.MaxInt")
}
