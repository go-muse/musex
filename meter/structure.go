package meter

import (
	"math/big"
	"slices"

	"github.com/go-muse/musex/duration"
)

// Span is a part of a measure: its start from the beginning of the measure
// and its duration. Units, groups, and beats are returned as spans in order:
// the first starts at 0, each next one starts at the end of the previous one,
// and the last one ends at the size of the meter.
//
// Spans returned by the package have both fields set. Span{} has both fields
// unset; the package never returns it and never takes a Span as input to a
// Meter operation. Span values are comparable with ==, which is the same as
// Equal.
type Span struct {
	// Start is the offset of the start from the beginning of the measure.
	Start duration.Position
	// Duration is the length of the span.
	Duration duration.Duration
}

// Equal reports whether s and other have equal fields. It is the same as
// s == other.
func (s Span) Equal(other Span) bool {
	return s.Start.Equal(other.Start) && s.Duration.Equal(other.Duration)
}

// End returns the end of s, Start + Duration. For spans returned by the
// package it never fails: the end is the start of the next span or the size
// of the meter.
//
// Errors, in this order: ErrInvalidArgument together with duration.ErrUnset
// for an unset Start, then for an unset Duration; ErrOutOfRange together with
// fraction.ErrOutOfRange and duration.ErrOutOfRange if the end is outside the
// range of common/fraction.
//
// Example: the end of 3/8@3/8 is 3/4.
func (s Span) End() (duration.Position, error) {
	const op = "Span.End"
	switch {
	case !s.Start.IsSet():
		return duration.Position{}, unsetInputError(op, "start")
	case !s.Duration.IsSet():
		return duration.Position{}, unsetInputError(op, "duration")
	}
	end, err := s.Start.Add(s.Duration)
	if err != nil {
		return duration.Position{}, causedRangeError(op, err)
	}
	return end, nil
}

// MustEnd is like End but panics with the error of End.
func (s Span) MustEnd() duration.Position {
	return must(s.End())
}

// String returns "<Duration>@<Start>" with the texts of package duration,
// such as "3/8@0/1", and "Duration(unset)@Position(unset)" for Span{}.
func (s Span) String() string {
	return s.Duration.String() + "@" + s.Start.String()
}

// ones returns a grouping of n groups of one unit each.
func ones(n int) []int {
	g := make([]int, n)
	for i := range g {
		g[i] = 1
	}
	return g
}

// spans returns the elements of r formed by the consecutive groups of units
// g, whose sum is the number of units of r. Every group lies within one part,
// so its units have one denominator.
func (r record) spans(g []int) []Span {
	starts, _ := r.layout()
	dens := r.unitDens()
	out := make([]Span, len(g))
	u := 0
	for i, k := range g {
		out[i] = Span{Start: position(starts[u]), Duration: duration.MustNew(k, dens[u])}
		u += k
	}
	return out
}

// written returns the written durations of the elements of r formed by the
// groups of units g. It fails with ErrNotNotatable at the first element in
// order that is not written as a single sign without tuplets.
func (r record) written(op string, g []int) ([]duration.Written, error) {
	dens := r.unitDens()
	out := make([]duration.Written, len(g))
	u := 0
	for i, k := range g {
		w, err := duration.MustNew(k, dens[u]).ToWritten()
		if err != nil {
			return nil, notNotatableError(op, err)
		}
		out[i] = w
		u += k
	}
	return out, nil
}

// groupingOf returns the grouping of r or an ErrGroupingRequired error.
func (r record) groupingOf(op string) ([]int, error) {
	g := r.grouping()
	if g == nil {
		return nil, simpleError(ErrGroupingRequired, op, "meter has no grouping")
	}
	return g, nil
}

// standardSimple reports whether units of denominators dens with grouping g
// form the standard simple structure: all units have one duration, there are
// two, three, or four of them, and g is their standard grouping.
func standardSimple(dens, g []int) bool {
	n := len(dens)
	if n < 2 || n > 4 {
		return false
	}
	for _, d := range dens[1:] {
		if d != dens[0] {
			return false
		}
	}
	return slices.Equal(g, standardGrouping(n))
}

// beatGrouping returns the beats of r as a grouping of its units: one unit
// per beat for the standard simple structure, otherwise the groups. It fails
// with ErrGroupingRequired for a meter without a grouping.
func (r record) beatGrouping(op string) ([]int, error) {
	g, err := r.groupingOf(op)
	if err != nil {
		return nil, err
	}
	if standardSimple(r.unitDens(), g) {
		return ones(r.numerator()), nil
	}
	return g, nil
}

// UnitCount returns the number of units of m: the sum of the numerators of
// its parts, from 1 to MaxUnits. Units are what the numerator counts: six
// eighths in 6/8.
//
// Errors: ErrUnset for the unset meter.
func (m Meter) UnitCount() (int, error) {
	r, err := m.recordOf("Meter.UnitCount")
	if err != nil {
		return 0, err
	}
	return r.numerator(), nil
}

// MustUnitCount is like UnitCount but panics with the error of UnitCount.
func (m Meter) MustUnitCount() int {
	return must(m.UnitCount())
}

// Unit returns the unit of m, 1/denominator: 1/4 for 3/4, 1/8 for 6/8, 1/3
// for 4/3.
//
// Errors, in this order: ErrUnset for the unset meter;
// ErrNoCommonDenominator for a sum with different denominators.
func (m Meter) Unit() (duration.Duration, error) {
	const op = "Meter.Unit"
	r, err := m.recordOf(op)
	if err != nil {
		return duration.Duration{}, err
	}
	d, ok := r.commonDen()
	if !ok {
		return duration.Duration{}, simpleError(ErrNoCommonDenominator, op, "parts have different denominators")
	}
	return duration.MustNew(1, d), nil
}

// MustUnit is like Unit but panics with the error of Unit.
func (m Meter) MustUnit() duration.Duration {
	return must(m.Unit())
}

// UnitWritten returns the unit of m as a written duration: a base without
// dots, such as a quarter for 3/4 and a 1024th for 6/1024. A unit is written
// only for a denominator that is a power of two from 1 to 1024.
//
// Errors, in this order: ErrUnset for the unset meter;
// ErrNoCommonDenominator for a sum with different denominators;
// ErrNotNotatable together with duration.ErrNotNotatable if the unit is not
// written as a single sign, such as for 4/3 and 3/2048.
func (m Meter) UnitWritten() (duration.Written, error) {
	const op = "Meter.UnitWritten"
	u, err := m.Unit()
	if err != nil {
		return duration.Written{}, err
	}
	w, err := u.ToWritten()
	if err != nil {
		return duration.Written{}, notNotatableError(op, err)
	}
	return w, nil
}

// MustUnitWritten is like UnitWritten but panics with the error of
// UnitWritten.
func (m Meter) MustUnitWritten() duration.Written {
	return must(m.UnitWritten())
}

// Units returns the units of m in order; they are defined without a grouping
// too. A part with numerator n and denominator d gives n units of 1/d: 6/8
// gives six units of 1/8 starting at 0, 1/8, …, 5/8; 2/4+3/8 gives 1/4@0,
// 1/4@1/4, 1/8@1/2, 1/8@5/8, 1/8@3/4. The result is a new slice.
//
// Errors: ErrUnset for the unset meter.
func (m Meter) Units() ([]Span, error) {
	r, err := m.recordOf("Meter.Units")
	if err != nil {
		return nil, err
	}
	return r.spans(ones(r.numerator())), nil
}

// MustUnits is like Units but panics with the error of Units.
func (m Meter) MustUnits() []Span {
	return must(m.Units())
}

// HasGrouping reports whether m has a grouping. A simple signature without a
// standard grouping and without an explicit grouping, such as 5/8, has none,
// and so does a sum with such a part.
//
// Errors: ErrUnset for the unset meter.
func (m Meter) HasGrouping() (bool, error) {
	r, err := m.recordOf("Meter.HasGrouping")
	if err != nil {
		return false, err
	}
	return r.grouping() != nil, nil
}

// MustHasGrouping is like HasGrouping but panics with the error of
// HasGrouping.
func (m Meter) MustHasGrouping() bool {
	return must(m.HasGrouping())
}

// HasExplicitGrouping reports whether m has an explicit grouping: true only
// for a simple signature or a symbol with a grouping that differs from the
// standard one, such as 5/8[2+3]. The groups of a compound numerator and of a
// sum are part of the record, not an explicit grouping.
//
// Errors: ErrUnset for the unset meter.
func (m Meter) HasExplicitGrouping() (bool, error) {
	r, err := m.recordOf("Meter.HasExplicitGrouping")
	if err != nil {
		return false, err
	}
	return r.explicit != nil, nil
}

// MustHasExplicitGrouping is like HasExplicitGrouping but panics with the
// error of HasExplicitGrouping.
func (m Meter) MustHasExplicitGrouping() bool {
	return must(m.HasExplicitGrouping())
}

// Grouping returns the sizes of the groups of m in units, not in beats: [3 2
// 3] for (3+2+3)/8, [2 3] for 2/4+3/8, [3 3] for 6/8. The result is a new
// slice.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// for a meter without a grouping.
func (m Meter) Grouping() ([]int, error) {
	const op = "Meter.Grouping"
	r, err := m.recordOf(op)
	if err != nil {
		return nil, err
	}
	return r.groupingOf(op)
}

// MustGrouping is like Grouping but panics with the error of Grouping.
func (m Meter) MustGrouping() []int {
	return must(m.Grouping())
}

// GroupCount returns the number of groups of m.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// for a meter without a grouping.
//
// Example: 12/8 has 4 groups.
func (m Meter) GroupCount() (int, error) {
	const op = "Meter.GroupCount"
	r, err := m.recordOf(op)
	if err != nil {
		return 0, err
	}
	g, err := r.groupingOf(op)
	if err != nil {
		return 0, err
	}
	return len(g), nil
}

// MustGroupCount is like GroupCount but panics with the error of GroupCount.
func (m Meter) MustGroupCount() int {
	return must(m.GroupCount())
}

// Groups returns the groups of m in order. The start of a group is the sum of
// the durations of the previous groups: (3+2+3)/8 gives 3/8@0, 1/4@3/8,
// 3/8@5/8. The result is a new slice.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// for a meter without a grouping.
func (m Meter) Groups() ([]Span, error) {
	const op = "Meter.Groups"
	r, err := m.recordOf(op)
	if err != nil {
		return nil, err
	}
	g, err := r.groupingOf(op)
	if err != nil {
		return nil, err
	}
	return r.spans(g), nil
}

// MustGroups is like Groups but panics with the error of Groups.
func (m Meter) MustGroups() []Span {
	return must(m.Groups())
}

// GroupsWritten returns the duration of each group of m as a written
// duration, by groups: a quarter and a dotted quarter for 5/8[2+3]. The
// result is a new slice.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// for a meter without a grouping; ErrNotNotatable together with
// duration.ErrNotNotatable for the whole list if a group is not written as a
// single sign, such as the group 5/8 of 7/8[5+2].
func (m Meter) GroupsWritten() ([]duration.Written, error) {
	const op = "Meter.GroupsWritten"
	r, err := m.recordOf(op)
	if err != nil {
		return nil, err
	}
	g, err := r.groupingOf(op)
	if err != nil {
		return nil, err
	}
	return r.written(op, g)
}

// MustGroupsWritten is like GroupsWritten but panics with the error of
// GroupsWritten.
func (m Meter) MustGroupsWritten() []duration.Written {
	return must(m.GroupsWritten())
}

// BeatCount returns the number of beats of m. A beat here is the beat of the
// English-language tradition, not a unit: 6/8 has two beats and six units.
// The beats are the units for the standard simple structure (two, three, or
// four equal units in their standard grouping, such as 3/4, 4/4, C, and
// (2+2)/4) and the groups otherwise.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// for a meter without a grouping, such as 5/8.
func (m Meter) BeatCount() (int, error) {
	const op = "Meter.BeatCount"
	r, err := m.recordOf(op)
	if err != nil {
		return 0, err
	}
	g, err := r.beatGrouping(op)
	if err != nil {
		return 0, err
	}
	return len(g), nil
}

// MustBeatCount is like BeatCount but panics with the error of BeatCount.
func (m Meter) MustBeatCount() int {
	return must(m.BeatCount())
}

// Beats returns the beats of m in order, as described for BeatCount: three
// beats of 1/4 for 3/4, two beats of 3/8 for 6/8, 1/4@0 and 3/8@1/4 for
// 5/8[2+3]. The result is a new slice.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// for a meter without a grouping.
func (m Meter) Beats() ([]Span, error) {
	const op = "Meter.Beats"
	r, err := m.recordOf(op)
	if err != nil {
		return nil, err
	}
	g, err := r.beatGrouping(op)
	if err != nil {
		return nil, err
	}
	return r.spans(g), nil
}

// MustBeats is like Beats but panics with the error of Beats.
func (m Meter) MustBeats() []Span {
	return must(m.Beats())
}

// BeatsWritten returns the duration of each beat of m as a written duration,
// by beats: two dotted quarters for 6/8, a half with two dots for 7/8[7]. A
// beat may be written even when the unit is not: the beats of 6/6 are two
// halves. The result is a new slice.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// for a meter without a grouping; ErrNotNotatable together with
// duration.ErrNotNotatable for the whole list if a beat is not written as a
// single sign, such as for 5/8[5] and 4/3.
func (m Meter) BeatsWritten() ([]duration.Written, error) {
	const op = "Meter.BeatsWritten"
	r, err := m.recordOf(op)
	if err != nil {
		return nil, err
	}
	g, err := r.beatGrouping(op)
	if err != nil {
		return nil, err
	}
	return r.written(op, g)
}

// MustBeatsWritten is like BeatsWritten but panics with the error of
// BeatsWritten.
func (m Meter) MustBeatsWritten() []duration.Written {
	return must(m.BeatsWritten())
}

// elementStarts returns the exact starts of the elements of r formed by the
// groups of units g, and the size of r.
func (r record) elementStarts(g []int) (starts []*big.Rat, size *big.Rat) {
	unitStarts, size := r.layout()
	starts = make([]*big.Rat, len(g))
	u := 0
	for i, k := range g {
		starts[i] = unitStarts[u]
		u += k
	}
	return starts, size
}
