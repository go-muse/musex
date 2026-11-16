package template

import "github.com/go-muse/musex/interval"

// AdjacentIntervalAfter returns the adjacent interval after place. A closed
// template accepts every positive virtual place and uses its cyclic order. An
// open template accepts only places before its last place.
func (t Template) AdjacentIntervalAfter(place int) (interval.Interval, error) {
	if !t.IsValid() {
		return interval.Interval{}, opError("AdjacentIntervalAfter", ErrUnset)
	}
	if place < 1 || (t.open && place >= t.Len()) {
		return interval.Interval{}, opError("AdjacentIntervalAfter", ErrInvalidArgument)
	}
	return t.intervals[(place-1)%len(t.intervals)], nil
}

// MustAdjacentIntervalAfter is AdjacentIntervalAfter without the error result.
// It panics with the same error.
func (t Template) MustAdjacentIntervalAfter(place int) interval.Interval {
	return must(t.AdjacentIntervalAfter(place))
}

// Cycle returns the exact sum of all adjacent intervals of a closed template.
// The second result is false for an open or unset template.
func (t Template) Cycle() (interval.Interval, bool) {
	if !t.IsClosed() {
		return interval.Interval{}, false
	}
	result, err := intervalFromCoordinates(sumCoordinates(t.intervals), "Cycle")
	if err != nil {
		// Construction validates the cycle, and rotation preserves its exact sum.
		return interval.Interval{}, false
	}
	return result, true
}

// IsOctave reports whether t is closed and its cycle is the perfect octave.
func (t Template) IsOctave() bool {
	cycle, ok := t.Cycle()
	return ok && cycle.Equal(interval.PerfectOctave())
}

// PlaceIntervals returns intervals from the first place to each of the Len()
// places, starting with the perfect unison. It returns nil for an unset
// template and never returns a partial result on error.
func (t Template) PlaceIntervals() ([]interval.Interval, error) {
	if !t.IsValid() {
		return nil, opError("PlaceIntervals", ErrUnset)
	}
	result := make([]interval.Interval, t.Len())
	result[0] = interval.PerfectUnison()
	total := zeroCoordinates()
	for i, current := range t.intervals {
		total = addCoordinates(total, coordinatesOf(current))
		if !t.open && i == len(t.intervals)-1 {
			break
		}
		converted, err := intervalFromCoordinates(total, "PlaceIntervals")
		if err != nil {
			return nil, err
		}
		result[i+1] = converted
	}
	return result, nil
}

// MustPlaceIntervals is PlaceIntervals without the error result. It panics
// with the same error.
func (t Template) MustPlaceIntervals() []interval.Interval {
	return must(t.PlaceIntervals())
}

// PlaceInterval returns the interval from the first place to place. A closed
// template accepts every positive virtual place; place Len()+1 is the first
// place raised by one cycle. An open template accepts only its actual places.
func (t Template) PlaceInterval(place int) (interval.Interval, error) {
	coordinates, err := t.placeCoordinates(place, "PlaceInterval")
	if err != nil {
		return interval.Interval{}, err
	}
	return intervalFromCoordinates(coordinates, "PlaceInterval")
}

// MustPlaceInterval is PlaceInterval without the error result. It panics with
// the same error.
func (t Template) MustPlaceInterval(place int) interval.Interval {
	return must(t.PlaceInterval(place))
}

func (t Template) placeCoordinates(place int, op string) (coordinates, error) {
	if !t.IsValid() {
		return coordinates{}, opError(op, ErrUnset)
	}
	if place < 1 || (t.open && place > t.Len()) {
		return coordinates{}, opError(op, ErrInvalidArgument)
	}
	if t.open {
		return prefixCoordinates(t.intervals, place-1), nil
	}
	k := t.Len()
	cycles, remainder := (place-1)/k, (place-1)%k
	result := prefixCoordinates(t.intervals, remainder)
	if cycles != 0 {
		result = addCoordinates(result, multiplyCoordinates(sumCoordinates(t.intervals), cycles))
	}
	return result, nil
}

// BetweenPlaces returns the directed interval from place i to place j. Equal
// places give the perfect unison. The calculation uses exact coordinates and
// materializes only their final difference, not either absolute place interval.
func (t Template) BetweenPlaces(i, j int) (interval.Interval, error) {
	if !t.IsValid() {
		return interval.Interval{}, opError("BetweenPlaces", ErrUnset)
	}
	from, err := t.placeCoordinates(i, "BetweenPlaces")
	if err != nil {
		return interval.Interval{}, err
	}
	to, err := t.placeCoordinates(j, "BetweenPlaces")
	if err != nil {
		return interval.Interval{}, err
	}
	if i == j {
		return interval.PerfectUnison(), nil
	}
	return intervalFromCoordinates(subCoordinates(to, from), "BetweenPlaces")
}

// MustBetweenPlaces is BetweenPlaces without the error result. It panics with
// the same error.
func (t Template) MustBetweenPlaces(i, j int) interval.Interval {
	return must(t.BetweenPlaces(i, j))
}

// MismatchingPlaces returns ascending places whose corresponding intervals are
// not equal. Adjacent compares neighboring intervals. FromFirst compares places
// 2 through Len(), and for closed templates also compares the cycle as place
// Len()+1. Templates must have the same kind and number of places.
func (t Template) MismatchingPlaces(other Template, measure Measure) ([]int, error) {
	if !t.IsValid() || !other.IsValid() {
		return nil, opError("MismatchingPlaces", ErrUnset)
	}
	if !measure.IsValid() {
		return nil, opError("MismatchingPlaces", ErrInvalidArgument)
	}
	if t.open != other.open || t.Len() != other.Len() {
		return nil, opError("MismatchingPlaces", ErrIncomparable)
	}
	result := make([]int, 0)
	if measure == Adjacent {
		for i := range t.intervals {
			if !t.intervals[i].Equal(other.intervals[i]) {
				result = append(result, i+1)
			}
		}
		return result, nil
	}

	left, right := zeroCoordinates(), zeroCoordinates()
	for i := 1; i < t.Len(); i++ {
		left = addCoordinates(left, coordinatesOf(t.intervals[i-1]))
		right = addCoordinates(right, coordinatesOf(other.intervals[i-1]))
		if !equalCoordinates(left, right) {
			result = append(result, i+1)
		}
	}
	if t.IsClosed() {
		left = addCoordinates(left, coordinatesOf(t.intervals[len(t.intervals)-1]))
		right = addCoordinates(right, coordinatesOf(other.intervals[len(other.intervals)-1]))
		if !equalCoordinates(left, right) {
			result = append(result, t.Len()+1)
		}
	}
	return result, nil
}

// MustMismatchingPlaces is MismatchingPlaces without the error result. It
// panics with the same error.
func (t Template) MustMismatchingPlaces(other Template, measure Measure) []int {
	return must(t.MismatchingPlaces(other, measure))
}
