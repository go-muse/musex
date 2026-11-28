package tempomap

import (
	"strings"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
)

// Segment is a positive-length constant or completed gradual segment,
// or the final open constant segment. An unfinished gradual tail is excluded.
type Segment struct {
	change        Change
	start         Address
	startSpeed    fraction.Fraction
	end           Address
	endSpeed      fraction.Fraction
	startPosition duration.Position
	endPosition   duration.Position
	open          bool
	set           bool
}

// IsSet reports whether s is set.
func (s Segment) IsSet() bool { return s.set }

// Change returns the change that starts the segment.
func (s Segment) Change() (Change, error) {
	if !s.IsSet() {
		return Change{}, simpleError(ErrUnset, "Segment.Change", "receiver")
	}
	return s.change, nil
}

// MustChange is like Change but panics with its error.
func (s Segment) MustChange() Change { return must(s.Change()) }

// Start returns the segment start address.
func (s Segment) Start() (Address, error) {
	if !s.IsSet() {
		return Address{}, simpleError(ErrUnset, "Segment.Start", "receiver")
	}
	return s.start, nil
}

// MustStart is like Start but panics with its error.
func (s Segment) MustStart() Address { return must(s.Start()) }

// StartSpeed returns the speed immediately after the starting change.
func (s Segment) StartSpeed() (fraction.Fraction, error) {
	if !s.IsSet() {
		return fraction.Fraction{}, simpleError(ErrUnset, "Segment.StartSpeed", "receiver")
	}
	return s.startSpeed, nil
}

// MustStartSpeed is like StartSpeed but panics with its error.
func (s Segment) MustStartSpeed() fraction.Fraction { return must(s.StartSpeed()) }

// End returns the following change address and true for a bounded segment.
func (s Segment) End() (Address, bool, error) {
	if !s.IsSet() {
		return Address{}, false, simpleError(ErrUnset, "Segment.End", "receiver")
	}
	if s.open {
		return Address{}, false, nil
	}
	return s.end, true, nil
}

// MustEnd is like End but panics with its error.
func (s Segment) MustEnd() (Address, bool) { return must2(s.End()) }

// EndSpeed returns the left-limit speed and true for a bounded segment.
func (s Segment) EndSpeed() (fraction.Fraction, bool, error) {
	if !s.IsSet() {
		return fraction.Fraction{}, false, simpleError(ErrUnset, "Segment.EndSpeed", "receiver")
	}
	if s.open {
		return fraction.Fraction{}, false, nil
	}
	return s.endSpeed, true, nil
}

// MustEndSpeed is like EndSpeed but panics with its error.
func (s Segment) MustEndSpeed() (fraction.Fraction, bool) { return must2(s.EndSpeed()) }

// Seconds returns the musical seconds of a bounded segment, without holds.
func (s Segment) Seconds(places int, mode rounding.Mode) (SecondsResult, bool, error) {
	const op = "Segment.Seconds"
	if !s.IsSet() {
		return SecondsResult{}, false, simpleError(ErrUnset, op, "receiver")
	}
	if err := checkPrecision(op, places, mode); err != nil {
		return SecondsResult{}, false, err
	}
	if s.open {
		return SecondsResult{}, false, nil
	}
	form, approximate := s.secondsForm()
	result, err := secondsResultFromLogForm(op, form, approximate, places, mode)
	if err != nil {
		return SecondsResult{}, false, err
	}
	return result, true, nil
}

// secondsForm returns the exact musical seconds of a bounded segment as a
// symbolic form and whether they are irrational.
func (s Segment) secondsForm() (logForm, bool) {
	start, end := s.startPosition, s.endPosition
	if s.change.kind == ChangeGradual {
		form := gradualSegmentForm(start, end, ratOfFraction(s.startSpeed), ratOfFraction(s.endSpeed))
		return form, !s.startSpeed.Equal(s.endSpeed)
	}
	return constantSegmentForm(start, end, ratOfFraction(s.startSpeed)), false
}

// MustSeconds is like Seconds but panics with its error.
func (s Segment) MustSeconds(places int, mode rounding.Mode) (SecondsResult, bool) {
	return must2(s.Seconds(places, mode))
}

// Equal reports whether s and other are the same segment.
func (s Segment) Equal(other Segment) bool { return s == other }

// String returns a diagnostic representation, not a serialization format.
func (s Segment) String() string {
	if !s.IsSet() {
		return "Segment(unset)"
	}
	var b strings.Builder
	b.WriteString("Segment{start:")
	b.WriteString(s.start.String())
	b.WriteString(",change:")
	b.WriteString(s.change.String())
	b.WriteString(",start-speed:")
	b.WriteString(s.startSpeed.ReducedString())
	if !s.open {
		b.WriteString(",end:")
		b.WriteString(s.end.String())
		b.WriteString(",end-speed:")
		b.WriteString(s.endSpeed.ReducedString())
	} else {
		b.WriteString(",open:true")
	}
	b.WriteByte('}')
	return b.String()
}

// State returns the exact current and main tempo outside a gradual segment.
func (m Map) State(at Address, grid ...measure.Grid) (State, error) {
	const op = "Map.State"
	if !m.IsSet() {
		return State{}, simpleError(ErrUnset, op, "receiver")
	}
	if !at.IsSet() {
		return State{}, simpleError(ErrUnset, op, "address")
	}
	g, err := m.gridArgument(op, grid)
	if err != nil {
		return State{}, err
	}
	position, err := addressPosition(op, at, m.addressing, g)
	if err != nil {
		return State{}, err
	}
	idx, err := m.positionIndex(grid)
	if err != nil {
		return State{}, err
	}
	if unfinishedAt(idx, position) {
		return State{}, simpleError(ErrUnfinished, op, "unfinished gradual segment")
	}
	current := idx.changes[changeAtPosition(idx, position)]
	if current.change.kind == ChangeGradual {
		return State{}, simpleError(ErrInvalidArgument, op, "no single exact tempo inside a gradual segment")
	}
	return current.after, nil
}

// MustState is like State but panics with its error.
func (m Map) MustState(at Address, grid ...measure.Grid) State { return must(m.State(at, grid...)) }

// Speed returns the exact whole-notes-per-minute speed at at.
func (m Map) Speed(at Address, grid ...measure.Grid) (fraction.Fraction, error) {
	const op = "Map.Speed"
	if !m.IsSet() {
		return fraction.Fraction{}, simpleError(ErrUnset, op, "receiver")
	}
	if !at.IsSet() {
		return fraction.Fraction{}, simpleError(ErrUnset, op, "address")
	}
	g, err := m.gridArgument(op, grid)
	if err != nil {
		return fraction.Fraction{}, err
	}
	position, err := addressPosition(op, at, m.addressing, g)
	if err != nil {
		return fraction.Fraction{}, err
	}
	idx, err := m.positionIndex(grid)
	if err != nil {
		return fraction.Fraction{}, err
	}
	if unfinishedAt(idx, position) {
		return fraction.Fraction{}, simpleError(ErrUnfinished, op, "unfinished gradual segment")
	}
	return speedAtIndex(idx, position)
}

// MustSpeed is like Speed but panics with its error.
func (m Map) MustSpeed(at Address, grid ...measure.Grid) fraction.Fraction {
	return must(m.Speed(at, grid...))
}

// TempoAt returns an exact marking equivalent to the speed in the requested unit.
func (m Map) TempoAt(at Address, unit duration.Written, grid ...measure.Grid) (tempo.Tempo, error) {
	const op = "Map.TempoAt"
	if !m.IsSet() {
		return tempo.Tempo{}, simpleError(ErrUnset, op, "receiver")
	}
	if !at.IsSet() {
		return tempo.Tempo{}, simpleError(ErrUnset, op, "address")
	}
	if !unit.IsSet() {
		return tempo.Tempo{}, dependencyArgumentError(op, duration.ErrUnset)
	}
	speed, err := m.Speed(at, grid...)
	if err != nil {
		return tempo.Tempo{}, err
	}
	unitDuration, err := unit.Duration()
	if err != nil {
		return tempo.Tempo{}, dependencyArgumentError(op, err)
	}
	// Both values are positive and in range, so the quotient fails only if
	// it is outside the range.
	bpmFraction, err := speed.Div(unitDuration.MustFraction())
	if err != nil {
		return tempo.Tempo{}, rangeError(op)
	}
	mark, err := tempo.New(unit, bpmFraction)
	if err != nil {
		return tempo.Tempo{}, tempoRangeError(op, err)
	}
	return mark, nil
}

// MustTempoAt is like TempoAt but panics with its error.
func (m Map) MustTempoAt(at Address, unit duration.Written, grid ...measure.Grid) tempo.Tempo {
	return must(m.TempoAt(at, unit, grid...))
}

// Segments returns positive-length constant and completed gradual segments and
// the final open constant segment.
func (m Map) Segments(grid ...measure.Grid) ([]Segment, error) {
	const op = "Map.Segments"
	if !m.IsSet() {
		return nil, simpleError(ErrUnset, op, "receiver")
	}
	if _, err := m.gridArgument(op, grid); err != nil {
		return nil, err
	}
	idx, err := m.positionIndex(grid)
	if err != nil {
		return nil, err
	}
	segments := make([]Segment, 0, len(idx.changes))
	for i, current := range idx.changes {
		if i+1 == len(idx.changes) {
			if current.change.kind != ChangeGradual {
				segments = append(segments, Segment{change: current.change, start: current.change.at, startSpeed: current.startSpeed, startPosition: current.position, open: true, set: true})
			}
			continue
		}
		next := idx.changes[i+1]
		if next.position.Compare(current.position) <= 0 {
			continue
		}
		endSpeed := current.startSpeed
		if current.change.kind == ChangeGradual {
			endSpeed = next.startSpeed
		}
		segments = append(segments, Segment{change: current.change, start: current.change.at, startSpeed: current.startSpeed, end: next.change.at, endSpeed: endSpeed, startPosition: current.position, endPosition: next.position, set: true})
	}
	return segments, nil
}

// MustSegments is like Segments but panics with its error.
func (m Map) MustSegments(grid ...measure.Grid) []Segment { return must(m.Segments(grid...)) }
