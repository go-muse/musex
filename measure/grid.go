package measure

import (
	"math"
	"math/big"
	"slices"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
)

const (
	// MaxParts is the largest number of parts returned by Split and SplitBy.
	MaxParts = 1024
	// MaxAlternation is the largest number of bars WithMetersIn fills with
	// a series of two or more meters.
	MaxAlternation = 1024
)

// Grid is a measure grid: bars with their meters and lengths from an origin.
// Grid is not comparable with ==; use Equal, EqualMarkup or EqualBarLengths.
//
// The record of a grid is the time signatures written at bars (marks; bar 1
// always has one), an optional pickup (bar 0), explicit lengths of single
// bars with their alignment, and an optional number of bars N. Without N the
// grid is open: the last meter continues, and the grid has bars 1 to
// math.MaxInt. With N the grid is finite and has bars 1 to N.
//
// A grid is an immutable value: operations that change it return a new grid,
// and on an error they return Grid{} and leave the receiver unchanged. A grid
// is safe for concurrent use. Copying a grid copies the value; its internal
// data are shared but never changed.
//
// The zero value Grid{} is the unset grid. Only IsSet, Equal, and String
// accept it without an error; String returns "Grid(unset)". Every other
// operation returns ErrUnset for an unset receiver or an unset grid argument.
type Grid struct {
	pickup  duration.Duration // unset when the grid has no pickup
	count   int               // N of a finite grid, 0 for an open grid
	marks   []Mark            // sorted by bar, the first at bar 1; nil for the unset grid
	lengths []ExplicitLength  // sorted by bar
	segs    []segment
	end     *big.Rat // end of bar N, or of bar math.MaxInt for an open grid
}

// New returns the open grid of meter m without a pickup: the mark of m is at
// bar 1, and all bars have the meter m.
//
// Errors: ErrInvalidArgument together with meter.ErrUnset for the unset
// meter.
//
// Example: New(meter.ThreeFour()) is "3/4"; its bar 4 starts at 9/4.
func New(m meter.Meter) (Grid, error) {
	const op = "New"
	if !m.IsSet() {
		return Grid{}, errUnsetMeter(op, "meter")
	}
	return build(op, duration.Duration{}, []Mark{{Bar: 1, Meter: m}}, nil, 0)
}

// MustNew is like New but panics with the error of New.
func MustNew(m meter.Meter) Grid {
	return must(New(m))
}

// IsSet reports whether g is not the unset grid Grid{}.
func (g Grid) IsSet() bool {
	return g.marks != nil
}

// Pickup returns the length of the pickup and true, or the unset duration
// and false if g has no pickup. No pickup is a result, not an error.
//
// Errors: ErrUnset for the unset grid.
//
// Example: for "3/4;pickup:1/4" it returns 1/4 and true.
func (g Grid) Pickup() (length duration.Duration, ok bool, err error) {
	if !g.IsSet() {
		return duration.Duration{}, false, errUnsetGrid("Grid.Pickup")
	}
	return g.pickup, g.pickup.IsSet(), nil
}

// MustPickup is like Pickup but panics with the error of Pickup.
func (g Grid) MustPickup() (length duration.Duration, ok bool) {
	return must2(g.Pickup())
}

// HasPickup reports whether g has a pickup.
//
// Errors: ErrUnset for the unset grid.
func (g Grid) HasPickup() (bool, error) {
	if !g.IsSet() {
		return false, errUnsetGrid("Grid.HasPickup")
	}
	return g.pickup.IsSet(), nil
}

// MustHasPickup is like HasPickup but panics with the error of HasPickup.
func (g Grid) MustHasPickup() bool {
	return must(g.HasPickup())
}

// Marks returns all marks of g in the order of bars; the first is at bar 1.
// Repeated marks, with the meter that already applies, are included. The
// result is a new slice.
//
// Errors: ErrUnset for the unset grid.
//
// Example: for "4/4;5:4/4" it returns [1:4/4 5:4/4].
func (g Grid) Marks() ([]Mark, error) {
	if !g.IsSet() {
		return nil, errUnsetGrid("Grid.Marks")
	}
	return slices.Clone(g.marks), nil
}

// MustMarks is like Marks but panics with the error of Marks.
func (g Grid) MustMarks() []Mark {
	return must(g.Marks())
}

// HasMark reports whether a mark is written at bar bar; the pickup never has
// one.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar; ErrOutsideGrid if g has no such bar.
func (g Grid) HasMark(bar int) (bool, error) {
	const op = "Grid.HasMark"
	if err := g.checkBar(op, bar); err != nil {
		return false, err
	}
	return bar > 0 && findMark(g.marks, bar) >= 0, nil
}

// MustHasMark is like HasMark but panics with the error of HasMark.
func (g Grid) MustHasMark(bar int) bool {
	return must(g.HasMark(bar))
}

// Changes returns the meter changes of g in order: the marks at bars k ≥ 2
// whose meter differs by record (meter.Meter.Equal) from the meter of bar
// k − 1. It returns nil if the meter never changes; a repeated mark is not a
// change. The result is a new slice.
//
// Errors: ErrUnset for the unset grid.
//
// Example: "4/4;5:C" has the change [5:C]; "4/4;5:4/4" has none.
func (g Grid) Changes() ([]Mark, error) {
	if !g.IsSet() {
		return nil, errUnsetGrid("Grid.Changes")
	}
	var out []Mark
	for i := 1; i < len(g.marks); i++ {
		if !g.marks[i].Meter.Equal(g.marks[i-1].Meter) {
			out = append(out, g.marks[i])
		}
	}
	return out, nil
}

// MustChanges is like Changes but panics with the error of Changes.
func (g Grid) MustChanges() []Mark {
	return must(g.Changes())
}

// ChangesAt reports whether the meter changes at bar bar: true only for the
// bars of Changes; false for the pickup and bar 1.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar; ErrOutsideGrid if g has no such bar.
func (g Grid) ChangesAt(bar int) (bool, error) {
	const op = "Grid.ChangesAt"
	if err := g.checkBar(op, bar); err != nil {
		return false, err
	}
	return g.changesAt(bar), nil
}

// changesAt reports whether the meter changes at bar k.
func (g Grid) changesAt(k int) bool {
	if k < 2 {
		return false
	}
	i := findMark(g.marks, k)
	return i > 0 && !g.marks[i].Meter.Equal(g.marks[i-1].Meter)
}

// MustChangesAt is like ChangesAt but panics with the error of ChangesAt.
func (g Grid) MustChangesAt(bar int) bool {
	return must(g.ChangesAt(bar))
}

// Lengths returns all explicit lengths of g in the order of bars, or nil if
// there are none. An explicit length equal to the size of the meter is
// included. The result is a new slice.
//
// Errors: ErrUnset for the unset grid.
//
// Example: for "4/4;4:length:1/2;5:length:1/4:end" it returns
// [4:length:1/2 5:length:1/4:end].
func (g Grid) Lengths() ([]ExplicitLength, error) {
	if !g.IsSet() {
		return nil, errUnsetGrid("Grid.Lengths")
	}
	if len(g.lengths) == 0 {
		return nil, nil
	}
	return slices.Clone(g.lengths), nil
}

// MustLengths is like Lengths but panics with the error of Lengths.
func (g Grid) MustLengths() []ExplicitLength {
	return must(g.Lengths())
}

// BarCount returns the number of numbered bars N and true for a finite grid,
// or 0 and false for an open grid. The pickup is not counted.
//
// Errors: ErrUnset for the unset grid.
func (g Grid) BarCount() (n int, finite bool, err error) {
	if !g.IsSet() {
		return 0, false, errUnsetGrid("Grid.BarCount")
	}
	return g.count, g.count > 0, nil
}

// MustBarCount is like BarCount but panics with the error of BarCount.
func (g Grid) MustBarCount() (n int, finite bool) {
	return must2(g.BarCount())
}

// IsFinite reports whether g has a number of bars.
//
// Errors: ErrUnset for the unset grid.
func (g Grid) IsFinite() (bool, error) {
	if !g.IsSet() {
		return false, errUnsetGrid("Grid.IsFinite")
	}
	return g.count > 0, nil
}

// MustIsFinite is like IsFinite but panics with the error of IsFinite.
func (g Grid) MustIsFinite() bool {
	return must(g.IsFinite())
}

// TotalBars returns the number of bars together with the pickup, N or N + 1,
// and true for a finite grid, or 0 and false for an open grid.
//
// Errors, in this order: ErrUnset for the unset grid; ErrOutOfRange if
// N + 1 is greater than math.MaxInt.
//
// Example: "4/4;pickup:1/4;8:length:3/4;bars:8" has 9 bars.
func (g Grid) TotalBars() (n int, finite bool, err error) {
	const op = "Grid.TotalBars"
	if !g.IsSet() {
		return 0, false, errUnsetGrid(op)
	}
	if g.count == 0 {
		return 0, false, nil
	}
	if !g.pickup.IsSet() {
		return g.count, true, nil
	}
	if g.count == math.MaxInt {
		return 0, false, errSimple(ErrOutOfRange, op, "the number of bars is greater than math.MaxInt")
	}
	return g.count + 1, true, nil
}

// MustTotalBars is like TotalBars but panics with the error of TotalBars.
func (g Grid) MustTotalBars() (n int, finite bool) {
	return must2(g.TotalBars())
}

// End returns the end of a finite grid, the end of bar N, and true, or the
// unset position and false for an open grid.
//
// Errors, in this order: ErrUnset for the unset grid; ErrOutOfRange together
// with fraction.ErrOutOfRange if the end is outside the range of
// common/fraction.
//
// Example: "4/4;bars:4" ends at 4/1.
func (g Grid) End() (end duration.Position, finite bool, err error) {
	const op = "Grid.End"
	if !g.IsSet() {
		return duration.Position{}, false, errUnsetGrid(op)
	}
	if g.count == 0 {
		return duration.Position{}, false, nil
	}
	p, err := toPosition(op, g.end)
	if err != nil {
		return duration.Position{}, false, err
	}
	return p, true, nil
}

// MustEnd is like End but panics with the error of End.
func (g Grid) MustEnd() (end duration.Position, finite bool) {
	return must2(g.End())
}

// FirstBar returns the number of the first bar: 0 if g has a pickup, and 1
// otherwise.
//
// Errors: ErrUnset for the unset grid.
func (g Grid) FirstBar() (int, error) {
	if !g.IsSet() {
		return 0, errUnsetGrid("Grid.FirstBar")
	}
	if g.pickup.IsSet() {
		return 0, nil
	}
	return 1, nil
}

// MustFirstBar is like FirstBar but panics with the error of FirstBar.
func (g Grid) MustFirstBar() int {
	return must(g.FirstBar())
}

// HasBar reports whether g has bar bar: bar 0 with a pickup, bars 1 to N of
// a finite grid, and bars 1 to math.MaxInt of an open grid.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar.
func (g Grid) HasBar(bar int) (bool, error) {
	const op = "Grid.HasBar"
	if !g.IsSet() {
		return false, errUnsetGrid(op)
	}
	if bar < 0 {
		return false, errInvalid(op, "negative bar number")
	}
	return g.hasBar(bar), nil
}

// MustHasBar is like HasBar but panics with the error of HasBar.
func (g Grid) MustHasBar(bar int) bool {
	return must(g.HasBar(bar))
}

// checkBar checks the receiver and a bar number that must exist: ErrUnset,
// ErrInvalidArgument for a negative bar, ErrOutsideGrid for a missing bar.
func (g Grid) checkBar(op string, bar int) error {
	if !g.IsSet() {
		return errUnsetGrid(op)
	}
	if bar < 0 {
		return errInvalid(op, "negative bar number")
	}
	return g.checkExists(op, bar)
}

// checkExists returns ErrOutsideGrid if g has no bar k.
func (g Grid) checkExists(op string, k int) error {
	if !g.hasBar(k) {
		if k == 0 {
			return errSimple(ErrOutsideGrid, op, "the grid has no pickup")
		}
		return errSimple(ErrOutsideGrid, op, "the bar is after the last bar of the grid")
	}
	return nil
}
