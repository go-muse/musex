package mode

import (
	"fmt"
	"iter"

	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/interval"
)

// IntervalAt describes an interval reached from a degree by moving Distance
// transitions. Degree never carries a note.
type IntervalAt struct {
	Degree   degree.Degree
	Distance int
	Interval interval.Interval
}

func (m Mode) tableCoordinates() []coordinates {
	count := m.Len()
	if m.IsClosed() {
		count *= 2
	}
	at := make([]coordinates, count)
	for i, p := range m.prefixes {
		at[i] = coordinatesOf(p)
	}
	if m.IsClosed() {
		cycle := coordinatesOf(m.cycle)
		for i, p := range at[:m.Len()] {
			at[m.Len()+i] = p.add(cycle)
		}
	}
	return at
}

// scanIntervals visits upward intervals without first materializing them as
// Interval values. A distance of zero or a full cycle is omitted.
func (m Mode) scanIntervals(at []coordinates, visit func(int, int, coordinates) bool) {
	for i := range m.Len() {
		limit := min(i+m.Len(), len(at))
		for j := i + 1; j < limit; j++ {
			if !visit(i, j-i, at[j].sub(at[i])) {
				return
			}
		}
	}
}

func (m Mode) intervalTable(chromaticOnly bool) (iter.Seq[IntervalAt], error) {
	if !m.IsValid() {
		return nil, ErrUnset
	}
	at := m.tableCoordinates()
	include := func(position coordinates) bool {
		if !chromaticOnly {
			return true
		}
		_, chromatic := position.qualities()
		return chromatic
	}

	var rangeErr error
	m.scanIntervals(at, func(i, distance int, position coordinates) bool {
		if include(position) {
			if _, err := position.ascendingInterval(); err != nil {
				rangeErr = fmt.Errorf("degree %d, distance %d: %w", i+1, distance, err)
				return false
			}
		}
		return true
	})
	if rangeErr != nil {
		return nil, rangeErr
	}

	return func(yield func(IntervalAt) bool) {
		m.scanIntervals(at, func(i, distance int, position coordinates) bool {
			if !include(position) {
				return true
			}
			// The same immutable coordinates were validated before the iterator
			// was returned. Conversion cannot fail during an iteration.
			iValue := must(position.ascendingInterval())
			return yield(IntervalAt{Degree: m.degrees[i], Distance: distance, Interval: iValue})
		})
	}, nil
}

// Intervals lists the upward intervals between distinct degrees, in order by
// starting degree and distance. It range-checks the whole table before return.
func (m Mode) Intervals() (iter.Seq[IntervalAt], error) { return m.intervalTable(false) }

// ChromaticIntervals lists the chromatic intervals in the same order as Intervals.
func (m Mode) ChromaticIntervals() (iter.Seq[IntervalAt], error) { return m.intervalTable(true) }

// OnlyDiatonicIntervals reports whether every interval in the table is diatonic.
func (m Mode) OnlyDiatonicIntervals() (bool, error) {
	if !m.IsValid() {
		return false, ErrUnset
	}
	all := true
	m.scanIntervals(m.tableCoordinates(), func(_, _ int, position coordinates) bool {
		all, _ = position.qualities()
		return all
	})
	return all, nil
}

// DegreesForInterval lists each starting degree that yields want at least
// once. A distance of zero or, for a closed mode, a whole cycle is included.
func (m Mode) DegreesForInterval(want interval.Interval) ([]degree.Degree, error) {
	if !m.IsValid() || !want.IsValid() {
		return nil, ErrUnset
	}
	if !want.IsAscending() {
		return nil, ErrInvalidArgument
	}
	at := m.tableCoordinates()
	target := coordinatesOf(want)
	result := make([]degree.Degree, 0)
	for i, d := range m.degrees {
		limit := min(i+m.Len()+1, len(at))
		for j := i; j < limit; j++ {
			if at[j].sub(at[i]).equal(target) {
				result = append(result, d)
				break
			}
		}
	}
	return result, nil
}
