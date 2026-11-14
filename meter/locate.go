package meter

import (
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
)

// Numbers of units, groups, and beats count from 1; offsets count from 0. An
// element k contains a position t if start_k ≤ t < start_k + duration_k: a
// position exactly at the start of an element belongs to it, and a position
// equal to the size of the meter is outside the meter.

// locate returns the number, counted from 1, of the element of r formed by
// the groups of units g that contains the set position t, and the offset of t
// in it. It fails with ErrOutsideMeter if t is not less than the size and
// with ErrOutOfRange if the offset is outside the range of common/fraction.
func (r record) locate(op string, g []int, t duration.Position) (int, duration.Position, error) {
	starts, size := r.elementStarts(g)
	x := bigRat(t)
	if x.Cmp(size) >= 0 {
		return 0, duration.Position{}, simpleError(ErrOutsideMeter, op, "position not less than the size")
	}
	i := len(starts) - 1
	for starts[i].Cmp(x) > 0 {
		i--
	}
	off, err := fraction.FromBigRat(new(big.Rat).Sub(x, starts[i]))
	if err != nil {
		return 0, duration.Position{}, causedRangeError(op, err)
	}
	return i + 1, duration.MustPositionFromFraction(off), nil
}

// startOf returns the start of the element number n, counted from 1, of r
// formed by the groups of units g. It fails with ErrOutsideMeter for n less
// than 1 or greater than the number of elements.
func (r record) startOf(op string, g []int, n int) (duration.Position, error) {
	if n < 1 || n > len(g) {
		return duration.Position{}, simpleError(ErrOutsideMeter, op, "number outside the meter")
	}
	starts, _ := r.elementStarts(g)
	return position(starts[n-1]), nil
}

// UnitAt returns the number of the unit of m that contains t, counted from
// 1, and the offset of t in it. A unit is found without a grouping too.
//
// Errors, in this order: ErrUnset for the unset meter; ErrInvalidArgument
// together with duration.ErrUnset for an unset t; ErrOutsideMeter if t is not
// less than the size of m; ErrOutOfRange together with fraction.ErrOutOfRange
// if the offset in the unit is outside the range of common/fraction.
//
// Example: in 6/8, t = 7/16 gives unit 4 and the offset 1/16.
func (m Meter) UnitAt(t duration.Position) (n int, offset duration.Position, err error) {
	const op = "Meter.UnitAt"
	r, err := m.recordOf(op)
	if err != nil {
		return 0, duration.Position{}, err
	}
	if !t.IsSet() {
		return 0, duration.Position{}, unsetInputError(op, "position")
	}
	return r.locate(op, ones(r.numerator()), t)
}

// MustUnitAt is like UnitAt but panics with the error of UnitAt.
func (m Meter) MustUnitAt(t duration.Position) (n int, offset duration.Position) {
	return must2(m.UnitAt(t))
}

// UnitStart returns the start of the unit number n of m, counted from 1.
//
// Errors, in this order: ErrUnset for the unset meter; ErrOutsideMeter for n
// less than 1 or greater than UnitCount.
//
// Example: unit 4 of 6/8 starts at 3/8.
func (m Meter) UnitStart(n int) (duration.Position, error) {
	const op = "Meter.UnitStart"
	r, err := m.recordOf(op)
	if err != nil {
		return duration.Position{}, err
	}
	return r.startOf(op, ones(r.numerator()), n)
}

// MustUnitStart is like UnitStart but panics with the error of UnitStart.
func (m Meter) MustUnitStart(n int) duration.Position {
	return must(m.UnitStart(n))
}

// GroupAt returns the number of the group of m that contains t, counted from
// 1, and the offset of t in it.
//
// Errors, in this order: ErrUnset for the unset meter; ErrInvalidArgument
// together with duration.ErrUnset for an unset t; ErrGroupingRequired for a
// meter without a grouping; ErrOutsideMeter if t is not less than the size of
// m; ErrOutOfRange together with fraction.ErrOutOfRange if the offset in the
// group is outside the range of common/fraction.
//
// Example: in 4/4, t = 3/4 gives group 2 and the offset 1/4.
func (m Meter) GroupAt(t duration.Position) (n int, offset duration.Position, err error) {
	const op = "Meter.GroupAt"
	r, err := m.recordOf(op)
	if err != nil {
		return 0, duration.Position{}, err
	}
	if !t.IsSet() {
		return 0, duration.Position{}, unsetInputError(op, "position")
	}
	g, err := r.groupingOf(op)
	if err != nil {
		return 0, duration.Position{}, err
	}
	return r.locate(op, g, t)
}

// MustGroupAt is like GroupAt but panics with the error of GroupAt.
func (m Meter) MustGroupAt(t duration.Position) (n int, offset duration.Position) {
	return must2(m.GroupAt(t))
}

// GroupStart returns the start of the group number n of m, counted from 1.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// for a meter without a grouping; ErrOutsideMeter for n less than 1 or
// greater than GroupCount.
//
// Example: group 3 of (3+2+3)/8 starts at 5/8.
func (m Meter) GroupStart(n int) (duration.Position, error) {
	const op = "Meter.GroupStart"
	r, err := m.recordOf(op)
	if err != nil {
		return duration.Position{}, err
	}
	g, err := r.groupingOf(op)
	if err != nil {
		return duration.Position{}, err
	}
	return r.startOf(op, g, n)
}

// MustGroupStart is like GroupStart but panics with the error of GroupStart.
func (m Meter) MustGroupStart(n int) duration.Position {
	return must(m.GroupStart(n))
}

// BeatAt returns the number of the beat of m that contains t, counted from
// 1, and the offset of t in it. Beats are described at BeatCount.
//
// Errors, in this order: ErrUnset for the unset meter; ErrInvalidArgument
// together with duration.ErrUnset for an unset t; ErrGroupingRequired for a
// meter without a grouping; ErrOutsideMeter if t is not less than the size of
// m; ErrOutOfRange together with fraction.ErrOutOfRange if the offset in the
// beat is outside the range of common/fraction.
//
// Example: in 6/8, t = 1/2 gives beat 2 and the offset 1/8.
func (m Meter) BeatAt(t duration.Position) (n int, offset duration.Position, err error) {
	const op = "Meter.BeatAt"
	r, err := m.recordOf(op)
	if err != nil {
		return 0, duration.Position{}, err
	}
	if !t.IsSet() {
		return 0, duration.Position{}, unsetInputError(op, "position")
	}
	g, err := r.beatGrouping(op)
	if err != nil {
		return 0, duration.Position{}, err
	}
	return r.locate(op, g, t)
}

// MustBeatAt is like BeatAt but panics with the error of BeatAt.
func (m Meter) MustBeatAt(t duration.Position) (n int, offset duration.Position) {
	return must2(m.BeatAt(t))
}

// BeatStart returns the start of the beat number n of m, counted from 1.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// for a meter without a grouping; ErrOutsideMeter for n less than 1 or
// greater than BeatCount.
//
// Example: beat 3 of 4/4 starts at 1/2.
func (m Meter) BeatStart(n int) (duration.Position, error) {
	const op = "Meter.BeatStart"
	r, err := m.recordOf(op)
	if err != nil {
		return duration.Position{}, err
	}
	g, err := r.beatGrouping(op)
	if err != nil {
		return duration.Position{}, err
	}
	return r.startOf(op, g, n)
}

// MustBeatStart is like BeatStart but panics with the error of BeatStart.
func (m Meter) MustBeatStart(n int) duration.Position {
	return must(m.BeatStart(n))
}
