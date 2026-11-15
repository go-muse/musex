package measure

import (
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
)

// RangeLength returns the length of the bars first…last: the sum of their
// lengths, BarEnd(last) − BarStart(first).
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative first bar, then for first after last; ErrOutsideGrid if g has
// no bar first, then no bar last; ErrOutOfRange together with
// fraction.ErrOutOfRange if the length is outside the range of
// common/fraction.
//
// Example: in "4/4;3:3/4", the bars 1…3 last 11/4.
func (g Grid) RangeLength(first, last int) (duration.Duration, error) {
	const op = "Grid.RangeLength"
	if err := g.checkRange(op, first, last); err != nil {
		return duration.Duration{}, err
	}
	return toDuration(op, ratSub(g.barEndRat(last), g.barStartRat(first)))
}

// MustRangeLength is like RangeLength but panics with the error of
// RangeLength.
func (g Grid) MustRangeLength(first, last int) duration.Duration {
	return must(g.RangeLength(first, last))
}

// ChangesWithin reports whether the meter changes at a bar k with
// first < k ≤ last. A change at bar first itself is not counted, and a
// repeated mark is not a change.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative first bar, then for first after last; ErrOutsideGrid if g has
// no bar first, then no bar last.
//
// Example: in "4/4;3:3/4", the bars 1…3 have a change and 3…5 have none.
func (g Grid) ChangesWithin(first, last int) (bool, error) {
	if err := g.checkRange("Grid.ChangesWithin", first, last); err != nil {
		return false, err
	}
	for i := 1; i < len(g.marks); i++ {
		k := g.marks[i].Bar
		if k > first && k <= last && !g.marks[i].Meter.Equal(g.marks[i-1].Meter) {
			return true, nil
		}
	}
	return false, nil
}

// MustChangesWithin is like ChangesWithin but panics with the error of
// ChangesWithin.
func (g Grid) MustChangesWithin(first, last int) bool {
	return must(g.ChangesWithin(first, last))
}

// AllFull reports whether every bar of first…last is full: its length equals
// the size of its meter.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative first bar, then for first after last; ErrOutsideGrid if g has
// no bar first, then no bar last.
//
// Example: in "4/4;3:length:1/2", the bars 1…2 are full and 1…3 are not.
func (g Grid) AllFull(first, last int) (bool, error) {
	if err := g.checkRange("Grid.AllFull", first, last); err != nil {
		return false, err
	}
	for i := g.segIndex(first); i <= g.segIndex(last); i++ {
		if g.segs[i].length.Cmp(g.segs[i].size) != 0 {
			return false, nil
		}
	}
	return true, nil
}

// MustAllFull is like AllFull but panics with the error of AllFull.
func (g Grid) MustAllFull(first, last int) bool {
	return must(g.AllFull(first, last))
}

// MeterRuns returns the runs of bars without a meter change that cover
// first…last in order: a new run starts at every bar k, first < k ≤ last,
// where the meter changes. A run from the pickup starts at 0. The result is
// a new slice.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative first bar, then for first after last; ErrOutsideGrid if g has
// no bar first, then no bar last.
//
// Example: in "4/4;3:3/4", the bars 1…5 give [1-2:4/4 3-5:3/4].
func (g Grid) MeterRuns(first, last int) ([]MeterRun, error) {
	if err := g.checkRange("Grid.MeterRuns", first, last); err != nil {
		return nil, err
	}
	cur := MeterRun{First: first, Meter: g.marks[markIndexAt(g.marks, first)].Meter}
	var out []MeterRun
	for i := 1; i < len(g.marks); i++ {
		mk := g.marks[i]
		if mk.Bar <= first || mk.Bar > last || mk.Meter.Equal(g.marks[i-1].Meter) {
			continue
		}
		cur.Last = mk.Bar - 1
		out = append(out, cur)
		cur = MeterRun{First: mk.Bar, Meter: mk.Meter}
	}
	cur.Last = last
	return append(out, cur), nil
}

// MustMeterRuns is like MeterRuns but panics with the error of MeterRuns.
func (g Grid) MustMeterRuns(first, last int) []MeterRun {
	return must(g.MeterRuns(first, last))
}

// BarsSpanned returns the numbers of the first and the last bars that the
// span of positions [start, start + value) occupies. A span that ends on a
// bar line does not occupy the next bar.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument
// together with duration.ErrUnset for an unset start, then for an unset
// value; ErrOutsideGrid if start + value is after the end of a finite grid;
// ErrOutOfRange if start + value is after the end of bar math.MaxInt of an
// open grid.
//
// Example: in 4/4, the span of 1/1 from 3/4 occupies the bars 1 and 2, and
// the span of 1/1 from 0/1 only bar 1.
func (g Grid) BarsSpanned(start duration.Position, value duration.Duration) (first, last int, err error) {
	const op = "Grid.BarsSpanned"
	if !g.IsSet() {
		return 0, 0, errUnsetGrid(op)
	}
	if !start.IsSet() {
		return 0, 0, errUnsetTime(op, "start")
	}
	if !value.IsSet() {
		return 0, 0, errUnsetTime(op, "value")
	}
	s := ratOfPosition(start)
	e := ratAdd(s, ratOfDuration(value))
	if err := g.checkSpanEnd(op, e); err != nil {
		return 0, 0, err
	}
	_, first, _ = g.locateRat(s)
	return first, int(g.lastBarBig(e).Int64()), nil
}

// MustBarsSpanned is like BarsSpanned but panics with the error of
// BarsSpanned.
func (g Grid) MustBarsSpanned(start duration.Position, value duration.Duration) (first, last int) {
	return must2(g.BarsSpanned(start, value))
}

// Elapsed returns the part of the length of the bar of a that has passed at
// a: the offset divided by the length of the bar, 0 ≤ x < 1.
//
// Errors, in this order: ErrUnset for the unset grid, then for the unset
// address; ErrOutsideGrid if g has no bar of a; ErrOutsideBar if the offset
// is not less than the length of the bar; ErrOutOfRange together with
// fraction.ErrOutOfRange if the result is outside the range of
// common/fraction.
//
// Example: in 3/4, 2:1/2 gives 2/3.
func (g Grid) Elapsed(a Address) (fraction.Fraction, error) {
	const op = "Grid.Elapsed"
	if err := g.checkAddress(op, a); err != nil {
		return fraction.Fraction{}, err
	}
	s, offset, err := g.addressInBar(op, a)
	if err != nil {
		return fraction.Fraction{}, err
	}
	return toFraction(op, ratQuo(offset, s.length))
}

// MustElapsed is like Elapsed but panics with the error of Elapsed.
func (g Grid) MustElapsed(a Address) fraction.Fraction {
	return must(g.Elapsed(a))
}

// PartOfBar returns value divided by the length of bar bar, a positive
// fraction that can be 1 or more.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar, then together with duration.ErrUnset for an unset value;
// ErrOutsideGrid if g has no such bar; ErrOutOfRange together with
// fraction.ErrOutOfRange if the result is outside the range of
// common/fraction.
//
// Example: in 3/4, 1/4 is 1/3 of a bar and 1/1 is 4/3.
func (g Grid) PartOfBar(bar int, value duration.Duration) (fraction.Fraction, error) {
	const op = "Grid.PartOfBar"
	if !g.IsSet() {
		return fraction.Fraction{}, errUnsetGrid(op)
	}
	if bar < 0 {
		return fraction.Fraction{}, errInvalid(op, "negative bar number")
	}
	if !value.IsSet() {
		return fraction.Fraction{}, errUnsetTime(op, "value")
	}
	if err := g.checkExists(op, bar); err != nil {
		return fraction.Fraction{}, err
	}
	s := &g.segs[g.segIndex(bar)]
	return toFraction(op, ratQuo(ratOfDuration(value), s.length))
}

// MustPartOfBar is like PartOfBar but panics with the error of PartOfBar.
func (g Grid) MustPartOfBar(bar int, value duration.Duration) fraction.Fraction {
	return must(g.PartOfBar(bar, value))
}
