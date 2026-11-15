package measure

import (
	"math"
	"slices"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
)

// rebuild returns the grid of the changed record, checking the pickup and
// the explicit lengths aligned to the end.
func (g Grid) rebuild(op string, pickup duration.Duration, marks []Mark, lengths []ExplicitLength, count int) (Grid, error) {
	if len(lengths) == 0 {
		lengths = nil
	}
	return build(op, pickup, marks, lengths, count)
}

// WithPickup returns g with a pickup of the length length, set or replaced:
// 0 < length ≤ the size of the meter of bar 1. A pickup of a whole bar is
// allowed and stays bar 0.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument
// together with duration.ErrUnset for an unset length; ErrLongerThanMeter if
// the pickup is longer than the meter of bar 1.
//
// Example: New(meter.ThreeFour()).WithPickup(duration.MustNew(1, 4)) is
// "3/4;pickup:1/4"; its bar 1 starts at 1/4.
func (g Grid) WithPickup(length duration.Duration) (Grid, error) {
	const op = "Grid.WithPickup"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if !length.IsSet() {
		return Grid{}, errUnsetTime(op, "length")
	}
	return g.rebuild(op, length, g.marks, g.lengths, g.count)
}

// MustWithPickup is like WithPickup but panics with the error of WithPickup.
func (g Grid) MustWithPickup(length duration.Duration) Grid {
	return must(g.WithPickup(length))
}

// WithoutPickup returns g without a pickup; a grid without a pickup is
// returned unchanged.
//
// Errors: ErrUnset for the unset grid.
func (g Grid) WithoutPickup() (Grid, error) {
	const op = "Grid.WithoutPickup"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if !g.pickup.IsSet() {
		return g, nil
	}
	return g.rebuild(op, duration.Duration{}, g.marks, g.lengths, g.count)
}

// MustWithoutPickup is like WithoutPickup but panics with the error of
// WithoutPickup.
func (g Grid) MustWithoutPickup() Grid {
	return must(g.WithoutPickup())
}

// WithMark returns g with the mark of the meter m at bar bar ≥ 1, set or
// replaced: m applies from bar bar until the next mark. A mark with the meter
// that already applies is kept as a repeated mark.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument if
// bar is less than 1, then together with meter.ErrUnset for an unset m;
// ErrOutsideGrid if bar is after the last bar of a finite grid;
// ErrLongerThanMeter if the pickup or a bar aligned to the end would become
// longer than its meter.
//
// Example: New(meter.FourFour()).WithMark(3, meter.ThreeFour()) is
// "4/4;3:3/4".
func (g Grid) WithMark(bar int, m meter.Meter) (Grid, error) {
	const op = "Grid.WithMark"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if bar < 1 {
		return Grid{}, errInvalid(op, "the bar is less than 1")
	}
	if !m.IsSet() {
		return Grid{}, errUnsetMeter(op, "meter")
	}
	if err := g.checkExists(op, bar); err != nil {
		return Grid{}, err
	}
	return g.rebuild(op, g.pickup, setMark(g.marks, Mark{Bar: bar, Meter: m}), g.lengths, g.count)
}

// MustWithMark is like WithMark but panics with the error of WithMark.
func (g Grid) MustWithMark(bar int, m meter.Meter) Grid {
	return must(g.WithMark(bar, m))
}

// setMark returns a new slice of marks with mk set or replaced.
func setMark(marks []Mark, mk Mark) []Mark {
	out := slices.Clone(marks)
	if i := findMark(out, mk.Bar); i >= 0 {
		out[i] = mk
		return out
	}
	i := markIndexAt(out, mk.Bar) + 1
	return slices.Insert(out, i, mk)
}

// WithoutMark returns g without the mark at bar bar ≥ 2; other marks are not
// changed. A bar without a mark, including a bar after the last bar of a
// finite grid, gives g unchanged.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument if
// bar is less than 2 (the mark of bar 1 cannot be removed);
// ErrLongerThanMeter if a bar aligned to the end would become longer than its
// meter.
//
// Example: "4/4;3:3/4;5:2/4" without the mark of bar 3 is "4/4;5:2/4".
func (g Grid) WithoutMark(bar int) (Grid, error) {
	const op = "Grid.WithoutMark"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if bar < 2 {
		return Grid{}, errInvalid(op, "the mark of bar 1 cannot be removed")
	}
	i := findMark(g.marks, bar)
	if i < 0 {
		return g, nil
	}
	return g.rebuild(op, g.pickup, slices.Delete(slices.Clone(g.marks), i, i+1), g.lengths, g.count)
}

// MustWithoutMark is like WithoutMark but panics with the error of
// WithoutMark.
func (g Grid) MustWithoutMark(bar int) Grid {
	return must(g.WithoutMark(bar))
}

// WithMetersIn returns g with the bars first…last, first ≥ 1, in the meters of
// the series meters in order and in a cycle: bar first + i gets the meter
// number (i mod j) + 1 of the j meters. Bars outside the range keep their
// meters:
//  1. the marks of the bars of the range are removed;
//  2. a mark is set at bar first and at every next bar of the range whose
//     meter differs by record from that of the previous bar;
//  3. if g has bar last + 1 without a mark and its former meter differs by
//     record from the new meter of bar last, the mark of the former meter is
//     set at bar last + 1.
//
// Explicit lengths are not changed. A series of one meter is allowed on a
// range of any length.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument if
// first is less than 1, then if first is after last, then for an empty
// series, then together with meter.ErrUnset for the first unset meter of the
// series; ErrOutsideGrid if last is after the last bar of a finite grid;
// ErrTooMany for a series of two or more meters on more than MaxAlternation
// bars; ErrLongerThanMeter if the pickup or a bar aligned to the end would
// become longer than its meter.
//
// Example: New(meter.FourFour()).WithMetersIn(1, 4, meter.SixEight(),
// meter.ThreeFour()) is "4/4" with marks [1:6/8 2:3/4 3:6/8 4:3/4 5:4/4].
func (g Grid) WithMetersIn(first, last int, meters ...meter.Meter) (Grid, error) {
	const op = "Grid.WithMetersIn"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if first < 1 {
		return Grid{}, errInvalid(op, "the first bar is less than 1")
	}
	if first > last {
		return Grid{}, errInvalid(op, "the first bar is after the last bar")
	}
	if len(meters) == 0 {
		return Grid{}, errInvalid(op, "empty series of meters")
	}
	for _, m := range meters {
		if !m.IsSet() {
			return Grid{}, errUnsetMeter(op, "meter of the series")
		}
	}
	if err := g.checkExists(op, last); err != nil {
		return Grid{}, err
	}
	j := len(meters)
	if j >= 2 && last-first+1 > MaxAlternation {
		return Grid{}, errSimple(ErrTooMany, op, "a series of meters on more than MaxAlternation bars")
	}
	hasNext := last < g.lastBar()
	var nextMeter meter.Meter
	nextMarked := false
	if hasNext {
		nextMeter = g.marks[markIndexAt(g.marks, last+1)].Meter
		nextMarked = findMark(g.marks, last+1) >= 0
	}
	marks := make([]Mark, 0, len(g.marks)+min(j, last-first+1)+1)
	for _, mk := range g.marks {
		if mk.Bar < first {
			marks = append(marks, mk)
		}
	}
	if j == 1 {
		marks = append(marks, Mark{Bar: first, Meter: meters[0]})
	} else {
		for i := 0; i <= last-first; i++ {
			m := meters[i%j]
			if i == 0 || !m.Equal(meters[(i-1)%j]) {
				marks = append(marks, Mark{Bar: first + i, Meter: m})
			}
		}
	}
	newLast := meters[(last-first)%j]
	if hasNext && !nextMarked && !nextMeter.Equal(newLast) {
		marks = append(marks, Mark{Bar: last + 1, Meter: nextMeter})
	}
	for _, mk := range g.marks {
		if mk.Bar > last {
			marks = append(marks, mk)
		}
	}
	return g.rebuild(op, g.pickup, marks, g.lengths, g.count)
}

// MustWithMetersIn is like WithMetersIn but panics with the error of
// WithMetersIn.
func (g Grid) MustWithMetersIn(first, last int, meters ...meter.Meter) Grid {
	return must(g.WithMetersIn(first, last, meters...))
}

// WithLength returns g with the explicit length length of bar bar ≥ 1 with
// the alignment align, set or replaced: any length is allowed with
// AlignStart, and with AlignEnd a length not longer than the size of the
// meter of the bar. The meter of the bar does not change, and a length equal
// to the size of the meter is kept.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument if
// bar is less than 1, then together with duration.ErrUnset for an unset
// length, then for an invalid align; ErrOutsideGrid if bar is after the last
// bar of a finite grid; ErrLongerThanMeter for AlignEnd with a length longer
// than the meter of the bar.
//
// Example: New(meter.FourFour()).WithLength(3, duration.MustNew(1, 2),
// AlignStart) is "4/4;3:length:1/2"; its bar 4 starts at 5/2.
func (g Grid) WithLength(bar int, length duration.Duration, align Alignment) (Grid, error) {
	const op = "Grid.WithLength"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if bar < 1 {
		return Grid{}, errInvalid(op, "the bar is less than 1")
	}
	if !length.IsSet() {
		return Grid{}, errUnsetTime(op, "length")
	}
	if !align.IsValid() {
		return Grid{}, errInvalid(op, "invalid alignment")
	}
	if err := g.checkExists(op, bar); err != nil {
		return Grid{}, err
	}
	l := ExplicitLength{Bar: bar, Duration: length, Alignment: align}
	lengths := slices.Clone(g.lengths)
	if i := findLength(lengths, bar); i >= 0 {
		lengths[i] = l
	} else {
		i := 0
		for i < len(lengths) && lengths[i].Bar < bar {
			i++
		}
		lengths = slices.Insert(lengths, i, l)
	}
	return g.rebuild(op, g.pickup, g.marks, lengths, g.count)
}

// MustWithLength is like WithLength but panics with the error of WithLength.
func (g Grid) MustWithLength(bar int, length duration.Duration, align Alignment) Grid {
	return must(g.WithLength(bar, length, align))
}

// WithoutLength returns g without the explicit length of bar bar ≥ 1. A bar
// without an explicit length, including a bar after the last bar of a finite
// grid, gives g unchanged.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument if
// bar is less than 1.
func (g Grid) WithoutLength(bar int) (Grid, error) {
	const op = "Grid.WithoutLength"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if bar < 1 {
		return Grid{}, errInvalid(op, "the bar is less than 1")
	}
	i := findLength(g.lengths, bar)
	if i < 0 {
		return g, nil
	}
	return g.rebuild(op, g.pickup, g.marks, slices.Delete(slices.Clone(g.lengths), i, i+1), g.count)
}

// MustWithoutLength is like WithoutLength but panics with the error of
// WithoutLength.
func (g Grid) MustWithoutLength(bar int) Grid {
	return must(g.WithoutLength(bar))
}

// lastRecordBar returns the greatest bar with a mark or an explicit length.
func (g Grid) lastRecordBar() int {
	k := g.marks[len(g.marks)-1].Bar
	if n := len(g.lengths); n > 0 && g.lengths[n-1].Bar > k {
		k = g.lengths[n-1].Bar
	}
	return k
}

// WithBarCount returns g as a finite grid with N = n bars: it decreases,
// increases, or sets the number of bars; new bars have the last meter.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument if
// n is less than 1; ErrOutsideGrid if a bar after n has a mark or an
// explicit length (Truncate drops them).
//
// Example: New(meter.FourFour()).WithBarCount(4) is "4/4;bars:4" and ends at
// 4/1.
func (g Grid) WithBarCount(n int) (Grid, error) {
	const op = "Grid.WithBarCount"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if n < 1 {
		return Grid{}, errInvalid(op, "the number of bars is less than 1")
	}
	if g.lastRecordBar() > n {
		return Grid{}, errSimple(ErrOutsideGrid, op, "a bar after the number of bars has a record")
	}
	return g.rebuild(op, g.pickup, g.marks, g.lengths, n)
}

// MustWithBarCount is like WithBarCount but panics with the error of
// WithBarCount.
func (g Grid) MustWithBarCount(n int) Grid {
	return must(g.WithBarCount(n))
}

// Truncate returns g cut to n bars: a finite grid with N = n, also from an
// open grid; the marks and explicit lengths of bars after n are dropped. A
// finite grid with N ≤ n is returned unchanged.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument if
// n is less than 1.
//
// Example: "4/4;6:3/4;7:length:1/2" truncated to 4 is "4/4;bars:4".
func (g Grid) Truncate(n int) (Grid, error) {
	const op = "Grid.Truncate"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if n < 1 {
		return Grid{}, errInvalid(op, "the number of bars is less than 1")
	}
	if g.count > 0 && g.count <= n {
		return g, nil
	}
	var marks []Mark
	for _, mk := range g.marks {
		if mk.Bar <= n {
			marks = append(marks, mk)
		}
	}
	var lengths []ExplicitLength
	for _, l := range g.lengths {
		if l.Bar <= n {
			lengths = append(lengths, l)
		}
	}
	return g.rebuild(op, g.pickup, marks, lengths, n)
}

// MustTruncate is like Truncate but panics with the error of Truncate.
func (g Grid) MustTruncate(n int) Grid {
	return must(g.Truncate(n))
}

// WithoutBarCount returns g as an open grid; an open grid is returned
// unchanged.
//
// Errors: ErrUnset for the unset grid.
func (g Grid) WithoutBarCount() (Grid, error) {
	const op = "Grid.WithoutBarCount"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if g.count == 0 {
		return g, nil
	}
	return g.rebuild(op, g.pickup, g.marks, g.lengths, 0)
}

// MustWithoutBarCount is like WithoutBarCount but panics with the error of
// WithoutBarCount.
func (g Grid) MustWithoutBarCount() Grid {
	return must(g.WithoutBarCount())
}

// AddBars returns the finite grid g with m ≥ 1 bars added at its end, N + m
// bars in all; the new bars have the meter of bar N.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument if
// m is less than 1; ErrUndefined for an open grid, which has no end;
// ErrOutOfRange if N + m is greater than math.MaxInt.
//
// Example: "4/4;bars:4" with 2 bars added is "4/4;bars:6" and ends at 6/1.
func (g Grid) AddBars(m int) (Grid, error) {
	const op = "Grid.AddBars"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if m < 1 {
		return Grid{}, errInvalid(op, "the number of added bars is less than 1")
	}
	if g.count == 0 {
		return Grid{}, errSimple(ErrUndefined, op, "an open grid has no end")
	}
	if g.count > math.MaxInt-m {
		return Grid{}, errSimple(ErrOutOfRange, op, "the number of bars is greater than math.MaxInt")
	}
	return g.rebuild(op, g.pickup, g.marks, g.lengths, g.count+m)
}

// MustAddBars is like AddBars but panics with the error of AddBars.
func (g Grid) MustAddBars(m int) Grid {
	return must(g.AddBars(m))
}

// InsertBars returns g with m ≥ 1 full bars inserted before bar before ≥ 1:
// the marks and explicit lengths of bars from before on get numbers greater
// by m, the mark of bar 1 stays at bar 1, and the inserted bars have the
// meter of bar before − 1 (before bar 1, the meter of bar 1). A finite grid
// gets N + m bars; inserting before bar N + 1 adds bars at its end.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument if
// before is less than 1, then if m is less than 1; ErrOutsideGrid if before
// is greater than N + 1 of a finite grid; ErrOutOfRange if the number of a
// moved record or N + m is greater than math.MaxInt.
//
// Example: "4/4;3:3/4" with 2 bars inserted before bar 3 is "4/4;5:3/4".
func (g Grid) InsertBars(before, m int) (Grid, error) {
	const op = "Grid.InsertBars"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if before < 1 {
		return Grid{}, errInvalid(op, "the bar is less than 1")
	}
	if m < 1 {
		return Grid{}, errInvalid(op, "the number of inserted bars is less than 1")
	}
	if g.count > 0 && before > g.count+1 {
		return Grid{}, errSimple(ErrOutsideGrid, op, "the bar is after the bar after the last bar")
	}
	tooFar := errSimple(ErrOutOfRange, op, "a bar number is greater than math.MaxInt")
	count := g.count
	if count > 0 {
		if count > math.MaxInt-m {
			return Grid{}, tooFar
		}
		count += m
	}
	marks := slices.Clone(g.marks)
	for i := range marks {
		if marks[i].Bar >= before && marks[i].Bar > 1 {
			if marks[i].Bar > math.MaxInt-m {
				return Grid{}, tooFar
			}
			marks[i].Bar += m
		}
	}
	lengths := slices.Clone(g.lengths)
	for i := range lengths {
		if lengths[i].Bar >= before {
			if lengths[i].Bar > math.MaxInt-m {
				return Grid{}, tooFar
			}
			lengths[i].Bar += m
		}
	}
	return g.rebuild(op, g.pickup, marks, lengths, count)
}

// MustInsertBars is like InsertBars but panics with the error of InsertBars.
func (g Grid) MustInsertBars(before, m int) Grid {
	return must(g.InsertBars(before, m))
}

// DeleteBars returns g without the bars first…last, first ≥ 1: the marks and
// explicit lengths of the deleted bars are dropped, and the next bars get
// numbers smaller by last − first + 1. If the next bar had no mark and its
// meter was set by a mark inside the deleted range, the mark of that meter is
// set at the new bar first. The bars of an open grid continue after bar
// math.MaxInt with its meter, so deleting up to math.MaxInt keeps that meter.
// A finite grid gets fewer bars; the pickup is removed by WithoutPickup.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument if
// first is less than 1, then if first is after last; ErrOutsideGrid if last
// is after the last bar of a finite grid; ErrInvalidArgument for all numbered
// bars of a finite grid; ErrLongerThanMeter if the pickup or a bar aligned to
// the end would become longer than its meter.
//
// Example: "4/4;3:3/4" without the bars 1…2 is "3/4".
func (g Grid) DeleteBars(first, last int) (Grid, error) {
	const op = "Grid.DeleteBars"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if first < 1 {
		return Grid{}, errInvalid(op, "the first bar is less than 1")
	}
	if first > last {
		return Grid{}, errInvalid(op, "the first bar is after the last bar")
	}
	if err := g.checkExists(op, last); err != nil {
		return Grid{}, err
	}
	if g.count > 0 && first == 1 && last == g.count {
		return Grid{}, errInvalid(op, "all numbered bars of a finite grid cannot be deleted")
	}
	d := last - first + 1
	// The bar after the range: the next bar of a finite grid, or the
	// continuation of an open grid after bar math.MaxInt.
	hasNext := g.count == 0 || last < g.count
	var marks []Mark
	for _, mk := range g.marks {
		switch {
		case mk.Bar < first:
			marks = append(marks, mk)
		case mk.Bar > last:
			marks = append(marks, Mark{Bar: mk.Bar - d, Meter: mk.Meter})
		}
	}
	if hasNext {
		nextBar := last
		if last < math.MaxInt {
			nextBar = last + 1
		}
		nextMarked := last < math.MaxInt && findMark(g.marks, last+1) >= 0
		setInside := g.marks[markIndexAt(g.marks, nextBar)].Bar >= first
		if !nextMarked && setInside {
			mk := Mark{Bar: first, Meter: g.marks[markIndexAt(g.marks, nextBar)].Meter}
			i := 0
			for i < len(marks) && marks[i].Bar < first {
				i++
			}
			marks = slices.Insert(marks, i, mk)
		}
	}
	var lengths []ExplicitLength
	for _, l := range g.lengths {
		switch {
		case l.Bar < first:
			lengths = append(lengths, l)
		case l.Bar > last:
			lengths = append(lengths, ExplicitLength{Bar: l.Bar - d, Duration: l.Duration, Alignment: l.Alignment})
		}
	}
	count := g.count
	if count > 0 {
		count -= d
	}
	return g.rebuild(op, g.pickup, marks, lengths, count)
}

// MustDeleteBars is like DeleteBars but panics with the error of DeleteBars.
func (g Grid) MustDeleteBars(first, last int) Grid {
	return must(g.DeleteBars(first, last))
}

// Slice returns the finite grid of the bars first…last, both of which exist
// and last ≥ 1; every bar keeps its meter, length, and alignment. For
// first ≥ 1 bar first + i becomes bar 1 + i, without a pickup, and
// N = last − first + 1; bar 1 of the slice does not become a pickup, even if
// it is aligned to the end. For first = 0 the pickup stays bar 0 and N =
// last. The new bar 1 has the mark of the former meter of bar first (of bar 1
// for first = 0); the marks of the other bars and the explicit lengths move
// with their bars.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative first, then if first is after last, then if last is less than
// 1 (a slice of the pickup alone); ErrOutsideGrid if g has no bar first,
// then no bar last.
//
// Example: "4/4;3:3/4" sliced to 3…5 is "3/4;bars:3".
func (g Grid) Slice(first, last int) (Grid, error) {
	const op = "Grid.Slice"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if first < 0 {
		return Grid{}, errInvalid(op, "negative bar number")
	}
	if first > last {
		return Grid{}, errInvalid(op, "the first bar is after the last bar")
	}
	if last < 1 {
		return Grid{}, errInvalid(op, "a slice of the pickup alone")
	}
	if err := g.checkExists(op, first); err != nil {
		return Grid{}, err
	}
	if err := g.checkExists(op, last); err != nil {
		return Grid{}, err
	}
	pickup := duration.Duration{}
	from := first
	if first == 0 {
		pickup = g.pickup
		from = 1
	}
	off := from - 1
	marks := []Mark{{Bar: 1, Meter: g.marks[markIndexAt(g.marks, from)].Meter}}
	for _, mk := range g.marks {
		if mk.Bar > from && mk.Bar <= last {
			marks = append(marks, Mark{Bar: mk.Bar - off, Meter: mk.Meter})
		}
	}
	var lengths []ExplicitLength
	for _, l := range g.lengths {
		if l.Bar >= from && l.Bar <= last {
			lengths = append(lengths, ExplicitLength{Bar: l.Bar - off, Duration: l.Duration, Alignment: l.Alignment})
		}
	}
	return g.rebuild(op, pickup, marks, lengths, last-off)
}

// MustSlice is like Slice but panics with the error of Slice.
func (g Grid) MustSlice(first, last int) Grid {
	return must(g.Slice(first, last))
}

// Append returns the grid of the bars of the finite grid g followed by the
// bars of other after bar N of g. Without a pickup, bar j of other becomes
// bar N + j. With a pickup, the pickup becomes bar N + 1 with the explicit
// length of the pickup aligned to the end, and bar j becomes bar N + 1 + j.
// The marks and explicit lengths of other move with their bars; the mark of
// bar 1 of other is set at its first bar in the result. The pickup, marks,
// and explicit lengths of g are kept. A finite other gives a finite result
// with the sum of the numbers of bars (one more for the former pickup); an
// open other gives an open result.
//
// Errors, in this order: ErrUnset for the unset grid, then for an unset
// other; ErrUndefined if g is open; ErrOutOfRange if a bar number or the
// number of bars is greater than math.MaxInt.
//
// Example: "4/4;bars:2" followed by "3/4;pickup:1/4;bars:2" is
// "4/4;3:3/4;3:length:1/4:end;bars:5".
func (g Grid) Append(other Grid) (Grid, error) {
	const op = "Grid.Append"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	if !other.IsSet() {
		return Grid{}, errUnset(op, "grid argument")
	}
	if g.count == 0 {
		return Grid{}, errSimple(ErrUndefined, op, "an open grid has no end")
	}
	tooFar := errSimple(ErrOutOfRange, op, "a bar number is greater than math.MaxInt")
	base := g.count // bar j of other becomes bar base + j
	if other.pickup.IsSet() {
		if base == math.MaxInt {
			return Grid{}, tooFar
		}
		base++
	}
	shift := func(k int) (int, bool) {
		if k > math.MaxInt-base {
			return 0, false
		}
		return k + base, true
	}
	marks := slices.Clone(g.marks)
	lengths := slices.Clone(g.lengths)
	if other.pickup.IsSet() {
		marks = append(marks, Mark{Bar: base, Meter: other.marks[0].Meter})
		lengths = append(lengths, ExplicitLength{Bar: base, Duration: other.pickup, Alignment: AlignEnd})
	}
	for i, mk := range other.marks {
		if i == 0 && other.pickup.IsSet() {
			continue
		}
		k, ok := shift(mk.Bar)
		if !ok {
			return Grid{}, tooFar
		}
		marks = append(marks, Mark{Bar: k, Meter: mk.Meter})
	}
	for _, l := range other.lengths {
		k, ok := shift(l.Bar)
		if !ok {
			return Grid{}, tooFar
		}
		lengths = append(lengths, ExplicitLength{Bar: k, Duration: l.Duration, Alignment: l.Alignment})
	}
	count := 0
	if other.count > 0 {
		k, ok := shift(other.count)
		if !ok {
			return Grid{}, tooFar
		}
		count = k
	}
	return g.rebuild(op, g.pickup, marks, lengths, count)
}

// MustAppend is like Append but panics with the error of Append.
func (g Grid) MustAppend(other Grid) Grid {
	return must(g.Append(other))
}

// Simplify returns the grid equal to g by markup without repeated marks
// (marks at bars k ≥ 2 with the meter of bar k − 1) and without explicit
// lengths equal to the size of the meter of their bar. The pickup and the
// number of bars do not change.
//
// Errors: ErrUnset for the unset grid.
//
// Example: "4/4;3:length:1/1;4:length:1/2;5:4/4;9:C" simplifies to
// "4/4;4:length:1/2;9:C".
func (g Grid) Simplify() (Grid, error) {
	const op = "Grid.Simplify"
	if !g.IsSet() {
		return Grid{}, errUnsetGrid(op)
	}
	marks := []Mark{g.marks[0]}
	for _, mk := range g.marks[1:] {
		if !mk.Meter.Equal(marks[len(marks)-1].Meter) {
			marks = append(marks, mk)
		}
	}
	var lengths []ExplicitLength
	for _, l := range g.lengths {
		m := g.marks[markIndexAt(g.marks, l.Bar)].Meter
		if ratOfDuration(l.Duration).Cmp(meterSize(m)) != 0 {
			lengths = append(lengths, l)
		}
	}
	return g.rebuild(op, g.pickup, marks, lengths, g.count)
}

// MustSimplify is like Simplify but panics with the error of Simplify.
func (g Grid) MustSimplify() Grid {
	return must(g.Simplify())
}
