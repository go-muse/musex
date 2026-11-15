package measure

import (
	"iter"
	"math"
	"math/big"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
)

// barOf returns the description of bar k of segment s, which starts at
// start.
func barOf(op string, s *segment, k int, start *big.Rat) (Bar, error) {
	if s.shiftErr {
		_, err := toFraction(op, s.shift)
		return Bar{}, err
	}
	startP, err := toPosition(op, start)
	if err != nil {
		return Bar{}, err
	}
	endP, err := toPosition(op, ratAdd(start, s.length))
	if err != nil {
		return Bar{}, err
	}
	kind := KindFull
	switch c := s.length.Cmp(s.size); {
	case c < 0:
		kind = KindIncomplete
	case c > 0:
		kind = KindExtended
	}
	first := k == s.first
	return Bar{
		Number:      k,
		Meter:       s.meter,
		Length:      s.lengthD,
		Start:       startP,
		End:         endP,
		Kind:        kind,
		Pickup:      s.pickup,
		Explicit:    s.explicit,
		Alignment:   s.align,
		Shift:       s.shiftP,
		Marked:      first && s.marked,
		MeterChange: first && s.change,
	}, nil
}

// bar returns the description of bar k, which exists.
func (g Grid) bar(op string, k int) (Bar, error) {
	s := &g.segs[g.segIndex(k)]
	return barOf(op, s, k, ratAdd(s.start, ratMulInt(s.length, k-s.first)))
}

// Bar returns the description of bar bar: its number, effective meter,
// length, start, end, kind, whether it is the pickup, its explicit length and
// alignment, its shift in the meter, its mark, and its meter change.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar; ErrOutsideGrid if g has no such bar; ErrOutOfRange together
// with fraction.ErrOutOfRange if the start, the end, or the shift of the bar
// is outside the range of common/fraction.
//
// Example: in "4/4;pickup:1/4", bar 0 has the length 1/4, the kind
// incomplete, the alignment end, and the shift 3/4.
func (g Grid) Bar(bar int) (Bar, error) {
	const op = "Grid.Bar"
	if err := g.checkBar(op, bar); err != nil {
		return Bar{}, err
	}
	return g.bar(op, bar)
}

// MustBar is like Bar but panics with the error of Bar.
func (g Grid) MustBar(bar int) Bar {
	return must(g.Bar(bar))
}

// BarMeter returns the effective meter of bar bar: the meter of the last mark
// at a bar not after it; for the pickup, the meter of bar 1.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar; ErrOutsideGrid if g has no such bar.
//
// Example: in "4/4;3:3/4", bar 100 has the meter 3/4.
func (g Grid) BarMeter(bar int) (meter.Meter, error) {
	if err := g.checkBar("Grid.BarMeter", bar); err != nil {
		return meter.Meter{}, err
	}
	return g.segs[g.segIndex(bar)].meter, nil
}

// MustBarMeter is like BarMeter but panics with the error of BarMeter.
func (g Grid) MustBarMeter(bar int) meter.Meter {
	return must(g.BarMeter(bar))
}

// BarLength returns the length of bar bar: the pickup length, the explicit
// length, or the size of its meter.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar; ErrOutsideGrid if g has no such bar.
func (g Grid) BarLength(bar int) (duration.Duration, error) {
	if err := g.checkBar("Grid.BarLength", bar); err != nil {
		return duration.Duration{}, err
	}
	return g.segs[g.segIndex(bar)].lengthD, nil
}

// MustBarLength is like BarLength but panics with the error of BarLength.
func (g Grid) MustBarLength(bar int) duration.Duration {
	return must(g.BarLength(bar))
}

// BarStart returns the start of bar bar from the origin: the sum of the
// lengths of all previous bars. The cost does not depend on the bar number.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar; ErrOutsideGrid if g has no such bar; ErrOutOfRange together
// with fraction.ErrOutOfRange if the start is outside the range of
// common/fraction.
//
// Example: New(meter.FourFour()).BarStart(1 << 62) is 4611686018427387903/1.
func (g Grid) BarStart(bar int) (duration.Position, error) {
	const op = "Grid.BarStart"
	if err := g.checkBar(op, bar); err != nil {
		return duration.Position{}, err
	}
	return toPosition(op, g.barStartRat(bar))
}

// MustBarStart is like BarStart but panics with the error of BarStart.
func (g Grid) MustBarStart(bar int) duration.Position {
	return must(g.BarStart(bar))
}

// BarEnd returns the end of bar bar: its start plus its length, the start of
// the next bar.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar; ErrOutsideGrid if g has no such bar; ErrOutOfRange together
// with fraction.ErrOutOfRange if the end is outside the range of
// common/fraction.
func (g Grid) BarEnd(bar int) (duration.Position, error) {
	const op = "Grid.BarEnd"
	if err := g.checkBar(op, bar); err != nil {
		return duration.Position{}, err
	}
	return toPosition(op, g.barEndRat(bar))
}

// MustBarEnd is like BarEnd but panics with the error of BarEnd.
func (g Grid) MustBarEnd(bar int) duration.Position {
	return must(g.BarEnd(bar))
}

// checkRange checks the receiver and a range of bars first…last: ErrUnset;
// ErrInvalidArgument for a negative first bar or first after last;
// ErrOutsideGrid for a missing first bar, then a missing last bar.
func (g Grid) checkRange(op string, first, last int) error {
	if !g.IsSet() {
		return errUnsetGrid(op)
	}
	if first < 0 {
		return errInvalid(op, "negative bar number")
	}
	if first > last {
		return errInvalid(op, "the first bar is after the last bar")
	}
	if err := g.checkExists(op, first); err != nil {
		return err
	}
	return g.checkExists(op, last)
}

// Bars returns the sequence of descriptions of bars first…last in the order
// of bars, as Bar returns them, one by one without building a list. The
// sequence can be stopped early, iterated again, and iterated concurrently;
// each iteration starts at first and gives the same values. All failures are
// reported by Bars itself, before iteration; iteration has no errors. The
// cost of the call and of each step does not depend on bar numbers.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative first bar, then for first after last; ErrOutsideGrid if g has
// no bar first, then no bar last; ErrOutOfRange together with
// fraction.ErrOutOfRange if the end of bar last or the shift of a bar of the
// range is outside the range of common/fraction.
//
// Example: the bars 1…1<<62 of an open 4/4 grid can be stopped after three
// bars that start at 0/1, 1/1, and 2/1.
func (g Grid) Bars(first, last int) (iter.Seq[Bar], error) {
	const op = "Grid.Bars"
	if err := g.checkRange(op, first, last); err != nil {
		return nil, err
	}
	fi, li := g.segIndex(first), g.segIndex(last)
	for i := fi; i <= li; i++ {
		if g.segs[i].shiftErr {
			_, err := toFraction(op, g.segs[i].shift)
			return nil, err
		}
		if i > fi && !inRange(g.segs[i].start) {
			_, err := toFraction(op, g.segs[i].start)
			return nil, err
		}
	}
	firstStart := g.barStartRat(first)
	if _, err := toFraction(op, firstStart); err != nil {
		return nil, err
	}
	if _, err := toFraction(op, g.barEndRat(last)); err != nil {
		return nil, err
	}
	return func(yield func(Bar) bool) {
		i := fi
		start := firstStart
		for k := first; ; k++ {
			for k > g.segs[i].last {
				i++
			}
			s := &g.segs[i]
			b, err := barOf(op, s, k, start)
			if err != nil || !yield(b) || k == last {
				return
			}
			start = ratAdd(start, s.length)
		}
	}, nil
}

// MustBars is like Bars but panics with the error of Bars.
func (g Grid) MustBars(first, last int) iter.Seq[Bar] {
	return must(g.Bars(first, last))
}

// Ordinal returns the ordinal place of bar bar in the sequence of bars,
// counted from 1 with the pickup first: bar + 1 with a pickup, bar
// otherwise.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar; ErrOutsideGrid if g has no such bar; ErrOutOfRange if the
// place is greater than math.MaxInt.
//
// Example: with a pickup, bar 0 is at place 1 and bar 3 at place 4.
func (g Grid) Ordinal(bar int) (int, error) {
	const op = "Grid.Ordinal"
	if err := g.checkBar(op, bar); err != nil {
		return 0, err
	}
	if !g.pickup.IsSet() {
		return bar, nil
	}
	if bar == math.MaxInt {
		return 0, errSimple(ErrOutOfRange, op, "the place is greater than math.MaxInt")
	}
	return bar + 1, nil
}

// MustOrdinal is like Ordinal but panics with the error of Ordinal.
func (g Grid) MustOrdinal(bar int) int {
	return must(g.Ordinal(bar))
}

// BarAtOrdinal returns the number of the bar at the ordinal place n, the
// inverse of Ordinal.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument if
// n is less than 1; ErrOutsideGrid if n is greater than the number of bars
// together with the pickup.
//
// Example: with a pickup, place 1 is bar 0 and place 4 is bar 3.
func (g Grid) BarAtOrdinal(n int) (int, error) {
	const op = "Grid.BarAtOrdinal"
	if !g.IsSet() {
		return 0, errUnsetGrid(op)
	}
	if n < 1 {
		return 0, errInvalid(op, "the place is less than 1")
	}
	k := n
	if g.pickup.IsSet() {
		k = n - 1
	}
	if !g.hasBar(k) {
		return 0, errSimple(ErrOutsideGrid, op, "the place is greater than the number of bars")
	}
	return k, nil
}

// MustBarAtOrdinal is like BarAtOrdinal but panics with the error of
// BarAtOrdinal.
func (g Grid) MustBarAtOrdinal(n int) int {
	return must(g.BarAtOrdinal(n))
}
