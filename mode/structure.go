package mode

import (
	"slices"

	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/interval"
)

// Degree returns the numbered degree without a note.
func (m Mode) Degree(number int) (degree.Degree, error) {
	if err := m.checkDegrees(number); err != nil {
		return degree.Degree{}, err
	}
	return m.degrees[number-1], nil
}

// Degrees returns an independent slice of all actual degrees, without notes.
func (m Mode) Degrees() ([]degree.Degree, error) {
	if !m.IsValid() {
		return nil, ErrUnset
	}
	return slices.Clone(m.degrees), nil
}

// DegreeInterval returns the interval from the first degree to number.
// Closed modes also accept positive virtual degree numbers.
func (m Mode) DegreeInterval(number int) (interval.Interval, error) {
	if number > 0 && number <= m.Len() {
		return m.prefixes[number-1], nil
	}
	i, err := m.structure.PlaceInterval(number)
	return i, templateError(err)
}

// AdjacentIntervalAfter returns the interval following number.
func (m Mode) AdjacentIntervalAfter(number int) (interval.Interval, error) {
	i, err := m.structure.AdjacentIntervalAfter(number)
	return i, templateError(err)
}

// BetweenDegrees returns the directed interval from one degree to another.
func (m Mode) BetweenDegrees(from, to int) (interval.Interval, error) {
	i, err := m.structure.BetweenPlaces(from, to)
	return i, templateError(err)
}

// Segment returns the adjacent intervals from from up to, but excluding, to.
func (m Mode) Segment(from, to int) ([]interval.Interval, error) {
	if !m.IsValid() {
		return nil, ErrUnset
	}
	if from < 1 || to <= from || to-from > m.Len() || (m.IsOpen() && to > m.Len()) {
		return nil, ErrInvalidArgument
	}
	result := make([]interval.Interval, to-from)
	for i := range result {
		result[i] = m.structure.MustAdjacentIntervalAfter(from + i)
	}
	return result, nil
}

func (m Mode) checkCircle(numbers ...int) error {
	if err := m.checkDegrees(numbers...); err != nil {
		return err
	}
	return m.checkClosed()
}

// Next moves one degree up in the circle and reports whether it wrapped.
func (m Mode) Next(number int) (degree.Degree, bool, error) {
	if err := m.checkCircle(number); err != nil {
		return degree.Degree{}, false, err
	}
	d, wrapped, err := m.degrees[number-1].Next(m.Len())
	return d, wrapped, degreeError(err)
}

// Previous moves one degree down in the circle and reports whether it wrapped.
func (m Mode) Previous(number int) (degree.Degree, bool, error) {
	if err := m.checkCircle(number); err != nil {
		return degree.Degree{}, false, err
	}
	d, wrapped, err := m.degrees[number-1].Previous(m.Len())
	return d, wrapped, degreeError(err)
}

// Shift moves by a signed number of degrees and returns the signed cycle count.
func (m Mode) Shift(number, distance int) (degree.Degree, int, error) {
	if err := m.checkCircle(number); err != nil {
		return degree.Degree{}, 0, err
	}
	d, cycles, err := m.degrees[number-1].Shift(m.Len(), distance)
	return d, cycles, degreeError(err)
}

// DistanceUp counts the upward transitions between actual degrees.
func (m Mode) DistanceUp(from, to int) (int, error) {
	if err := m.checkCircle(from, to); err != nil {
		return 0, err
	}
	distance, err := m.degrees[from-1].DistanceUp(m.degrees[to-1], m.Len())
	return distance, degreeError(err)
}

// DistanceDown counts the downward transitions between actual degrees.
func (m Mode) DistanceDown(from, to int) (int, error) {
	if err := m.checkCircle(from, to); err != nil {
		return 0, err
	}
	distance, err := m.degrees[from-1].DistanceDown(m.degrees[to-1], m.Len())
	return distance, degreeError(err)
}

// ShortestDistance returns the shortest signed transition count in the circle.
func (m Mode) ShortestDistance(from, to int) (int, error) {
	if err := m.checkCircle(from, to); err != nil {
		return 0, err
	}
	distance, err := m.degrees[from-1].ShortestDistance(m.degrees[to-1], m.Len())
	return distance, degreeError(err)
}
