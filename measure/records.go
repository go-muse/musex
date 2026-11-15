package measure

import (
	"strconv"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
)

// Mark is a time signature written at a bar: the meter from that bar on,
// until the next mark. Grid.Marks and Grid.Changes return marks in the order
// of bars; in their results Bar is at least 1 and Meter is set.
//
// Mark{} has both fields unset; operations never take a Mark as input. Mark
// values are comparable with ==, which is the same as Equal.
type Mark struct {
	// Bar is the number of the bar where the time signature is written.
	Bar int
	// Meter is the time signature.
	Meter meter.Meter
}

// Equal reports whether m and other have the same bar and the same meter by
// record (meter.Meter.Equal); it is the same as ==.
func (m Mark) Equal(other Mark) bool {
	return m == other
}

// String returns "<Bar>:<Meter>", such as "3:3/4".
func (m Mark) String() string {
	return strconv.Itoa(m.Bar) + ":" + m.Meter.String()
}

// ExplicitLength is the length of one bar set apart from its meter, with the
// alignment of the bar in a full bar of its meter. Grid.Lengths returns
// explicit lengths in the order of bars; in its result Bar is at least 1,
// Duration is set, and Alignment is valid.
//
// ExplicitLength{} has all fields unset; operations never take an
// ExplicitLength as input. ExplicitLength values are comparable with ==,
// which is the same as Equal.
type ExplicitLength struct {
	// Bar is the number of the bar.
	Bar int
	// Duration is the length of the bar.
	Duration duration.Duration
	// Alignment places the bar in a full bar of its meter.
	Alignment Alignment
}

// Equal reports whether l and other have the same bar, length, and
// alignment; it is the same as ==.
func (l ExplicitLength) Equal(other ExplicitLength) bool {
	return l == other
}

// String returns "<Bar>:length:<Duration>", followed by ":end" for
// AlignEnd: "4:length:1/2", "5:length:1/4:end". An invalid alignment is
// written as ":" and the text of the alignment.
func (l ExplicitLength) String() string {
	s := strconv.Itoa(l.Bar) + ":length:" + l.Duration.String()
	switch l.Alignment {
	case AlignStart:
	case AlignEnd:
		s += ":end"
	default:
		s += ":" + l.Alignment.String()
	}
	return s
}

// Bar describes one bar of a grid: its number, meter, length, start and
// end, kind, and record. Grid.Bar, Grid.BarAt, and Grid.Bars return it with
// all fields set.
//
// Bar{} has all fields unset; operations never take a Bar as input. Bar
// values are comparable with ==, which is the same as Equal.
type Bar struct {
	// Number is the number of the bar; 0 for the pickup.
	Number int // 0 for the pickup
	// Meter is the effective meter of the bar: the meter of the last mark at
	// a bar not after it; for the pickup, the meter of bar 1.
	Meter meter.Meter // effective meter
	// Length is the length of the bar: the pickup length, the explicit
	// length, or the size of the meter.
	Length duration.Duration // length of the bar
	// Start is the start of the bar from the origin of the grid.
	Start duration.Position // start from the origin of the grid
	// End is Start + Length: the start of the next bar.
	End duration.Position // Start + Length
	// Kind compares Length with the size of Meter.
	Kind Kind
	// Pickup reports whether the bar is the pickup.
	Pickup bool // the bar is the pickup
	// Explicit reports whether the bar has an explicit length; it is false
	// for the pickup.
	Explicit bool // the bar has an explicit length
	// Alignment is the alignment of the explicit length; AlignEnd for the
	// pickup; AlignStart for other bars.
	Alignment Alignment
	// Shift is the offset of the bar start in a full bar of its meter: the
	// size of the meter minus the length for the pickup and for a bar aligned
	// to the end, and 0 otherwise.
	Shift duration.Position // offset of the bar start in a full bar of its meter
	// Marked reports whether a time signature is written at the bar; it is
	// always true for bar 1 and false for the pickup.
	Marked bool // a time signature is written at the bar
	// MeterChange reports whether the meter of the bar differs by record from
	// the meter of the previous bar; it is false for the pickup and bar 1.
	MeterChange bool // the meter differs by record from the previous bar
}

// Equal reports whether all fields of b and other are equal; it is the same
// as ==.
func (b Bar) Equal(other Bar) bool {
	return b == other
}

// MeterRun is a run of consecutive bars First…Last without a meter change,
// all of meter Meter. Grid.MeterRuns returns runs with First ≤ Last and a set
// meter.
//
// MeterRun{} has all fields unset; operations never take a MeterRun as input.
// MeterRun values are comparable with ==, which is the same as Equal.
type MeterRun struct {
	// First is the number of the first bar of the run.
	First int
	// Last is the number of the last bar of the run.
	Last int
	// Meter is the meter of every bar of the run.
	Meter meter.Meter
}

// Equal reports whether r and other have the same bars and the same meter by
// record; it is the same as ==.
func (r MeterRun) Equal(other MeterRun) bool {
	return r == other
}

// String returns "<First>-<Last>:<Meter>", such as "3-5:3/4".
func (r MeterRun) String() string {
	return strconv.Itoa(r.First) + "-" + strconv.Itoa(r.Last) + ":" + r.Meter.String()
}

// Part is a piece of a split duration inside one bar, or inside one group,
// beat, or unit of the meter of a bar: the address of its start, the position
// of its start, and its length. Grid.Split and Grid.SplitBy return parts with
// all fields set; the position of the address in the grid equals Start.
//
// Part{} has all fields unset; only Part.End takes a Part assembled by the
// caller. Part values are comparable with ==, which is the same as Equal.
type Part struct {
	// Address is the bar and the offset of the start of the part.
	Address Address
	// Start is the position of the start of the part.
	Start duration.Position
	// Duration is the length of the part.
	Duration duration.Duration
}

// Equal reports whether p and other have equal fields; it is the same as ==.
func (p Part) Equal(other Part) bool {
	return p == other
}

// End returns Start + Duration, the position of the end of the part.
//
// Errors, in this order: ErrInvalidArgument together with duration.ErrUnset
// if Start is unset, then if Duration is unset; ErrOutOfRange together with
// fraction.ErrOutOfRange if the sum is outside the range of common/fraction.
//
// Example: the part 1/4@1:3/4 starting at the position 3/4 ends at 1/1.
func (p Part) End() (duration.Position, error) {
	const op = "Part.End"
	if !p.Start.IsSet() {
		return duration.Position{}, errUnsetTime(op, "start")
	}
	if !p.Duration.IsSet() {
		return duration.Position{}, errUnsetTime(op, "duration")
	}
	return toPosition(op, ratAdd(ratOfPosition(p.Start), ratOfDuration(p.Duration)))
}

// MustEnd is like End but panics with the error of End.
func (p Part) MustEnd() duration.Position {
	return must(p.End())
}

// String returns "<Duration>@<Address>", such as "1/4@2:0/1".
func (p Part) String() string {
	return p.Duration.String() + "@" + p.Address.String()
}
