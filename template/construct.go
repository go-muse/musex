package template

import (
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
)

// ClosedFromIntervals creates a closed template from its adjacent intervals.
// The last interval closes the cycle. The input slice is copied.
func ClosedFromIntervals(intervals []interval.Interval) (Template, error) {
	return templateFromIntervals(intervals, false, "ClosedFromIntervals")
}

// MustClosedFromIntervals is ClosedFromIntervals without the error result. It
// panics with the same error.
func MustClosedFromIntervals(intervals []interval.Interval) Template {
	return must(ClosedFromIntervals(intervals))
}

// OpenFromIntervals creates an open template from its adjacent intervals. One
// interval is sufficient and describes two places. The input slice is copied.
func OpenFromIntervals(intervals []interval.Interval) (Template, error) {
	return templateFromIntervals(intervals, true, "OpenFromIntervals")
}

// MustOpenFromIntervals is OpenFromIntervals without the error result. It
// panics with the same error.
func MustOpenFromIntervals(intervals []interval.Interval) Template {
	return must(OpenFromIntervals(intervals))
}

func templateFromIntervals(intervals []interval.Interval, open bool, op string) (Template, error) {
	for _, current := range intervals {
		if !current.IsValid() {
			return Template{}, unsetIntervalError(op)
		}
	}
	if len(intervals) == 0 {
		return Template{}, opError(op, ErrEmptyInput)
	}
	places := len(intervals)
	if open {
		places++
	}
	if places > MaxPlaces {
		return Template{}, opError(op, ErrInvalidArgument)
	}
	for _, current := range intervals {
		if !current.IsAscending() || current.IsPerfectUnison() {
			return Template{}, opError(op, ErrInvalidInterval)
		}
	}

	if err := validateDerivedIntervals(intervals, open, op); err != nil {
		return Template{}, err
	}
	return Template{intervals: append([]interval.Interval{}, intervals...), open: open}, nil
}

// validateDerivedIntervals checks every place interval of an open template and
// only the complete cycle of a closed template. Accumulation uses exact
// intermediate coordinates, and only each required result is materialized.
func validateDerivedIntervals(intervals []interval.Interval, open bool, op string) error {
	total := zeroCoordinates()
	for _, current := range intervals {
		total = addCoordinates(total, coordinatesOf(current))
		if open {
			if _, err := intervalFromCoordinates(total, op); err != nil {
				return err
			}
		}
	}
	if !open {
		_, err := intervalFromCoordinates(total, op)
		return err
	}
	return nil
}

// ClosedFromPlaceIntervals creates a closed template from intervals from the
// first place to each place and the interval of its cycle. The first place
// interval must be the perfect unison. The input slice is copied into the
// resulting adjacent-interval representation.
func ClosedFromPlaceIntervals(places []interval.Interval, cycle interval.Interval) (Template, error) {
	return templateFromPlaceIntervals(places, cycle, false, "ClosedFromPlaceIntervals")
}

// MustClosedFromPlaceIntervals is ClosedFromPlaceIntervals without the error
// result. It panics with the same error.
func MustClosedFromPlaceIntervals(places []interval.Interval, cycle interval.Interval) Template {
	return must(ClosedFromPlaceIntervals(places, cycle))
}

// OpenFromPlaceIntervals creates an open template from intervals from the
// first place to each place. At least two places are required. The input slice
// is copied into the resulting adjacent-interval representation.
func OpenFromPlaceIntervals(places []interval.Interval) (Template, error) {
	return templateFromPlaceIntervals(places, interval.Interval{}, true, "OpenFromPlaceIntervals")
}

// MustOpenFromPlaceIntervals is OpenFromPlaceIntervals without the error
// result. It panics with the same error.
func MustOpenFromPlaceIntervals(places []interval.Interval) Template {
	return must(OpenFromPlaceIntervals(places))
}

func templateFromPlaceIntervals(places []interval.Interval, cycle interval.Interval, open bool, op string) (Template, error) {
	for _, current := range places {
		if !current.IsValid() {
			return Template{}, unsetIntervalError(op)
		}
	}
	if !open && !cycle.IsValid() {
		return Template{}, unsetIntervalError(op)
	}
	if len(places) == 0 || (open && len(places) == 1) {
		return Template{}, opError(op, ErrEmptyInput)
	}
	if len(places) > MaxPlaces {
		return Template{}, opError(op, ErrInvalidArgument)
	}
	if !places[0].Equal(interval.PerfectUnison()) {
		return Template{}, opError(op, ErrInvalidInterval)
	}

	differences := make([]coordinates, 0, len(places))
	for i := 1; i < len(places); i++ {
		difference := subCoordinates(coordinatesOf(places[i]), coordinatesOf(places[i-1]))
		if !isAdjacent(difference) {
			return Template{}, opError(op, ErrInvalidInterval)
		}
		differences = append(differences, difference)
	}
	if !open {
		closing := subCoordinates(coordinatesOf(cycle), coordinatesOf(places[len(places)-1]))
		if !isAdjacent(closing) {
			return Template{}, opError(op, ErrInvalidInterval)
		}
		differences = append(differences, closing)
	}

	adjacent := make([]interval.Interval, len(differences))
	for i, difference := range differences {
		current, err := intervalFromCoordinates(difference, op)
		if err != nil {
			return Template{}, err
		}
		adjacent[i] = current
	}
	return Template{intervals: adjacent, open: open}, nil
}

// ClosedFromSpellings creates a closed template from note names read upward in
// order and an explicit cycle. The cycle need not be an octave. Names are not
// retained, and no musical system is consulted.
func ClosedFromSpellings(names []spelling.Spelling, cycle interval.Interval) (Template, error) {
	return templateFromSpellings(names, cycle, false, "ClosedFromSpellings")
}

// MustClosedFromSpellings is ClosedFromSpellings without the error result. It
// panics with the same error.
func MustClosedFromSpellings(names []spelling.Spelling, cycle interval.Interval) Template {
	return must(ClosedFromSpellings(names, cycle))
}

// FromSpellingsInOctave is exactly ClosedFromSpellings with a perfect octave as
// the cycle.
func FromSpellingsInOctave(names []spelling.Spelling) (Template, error) {
	return ClosedFromSpellings(names, interval.PerfectOctave())
}

// MustFromSpellingsInOctave is FromSpellingsInOctave without the error result.
// It panics with the same error.
func MustFromSpellingsInOctave(names []spelling.Spelling) Template {
	return must(FromSpellingsInOctave(names))
}

// OpenFromSpellings creates an open template from at least two note names read
// upward in order. Names are not retained, and no musical system is consulted.
func OpenFromSpellings(names []spelling.Spelling) (Template, error) {
	return templateFromSpellings(names, interval.Interval{}, true, "OpenFromSpellings")
}

// MustOpenFromSpellings is OpenFromSpellings without the error result. It
// panics with the same error.
func MustOpenFromSpellings(names []spelling.Spelling) Template {
	return must(OpenFromSpellings(names))
}

func templateFromSpellings(names []spelling.Spelling, cycle interval.Interval, open bool, op string) (Template, error) {
	for _, name := range names {
		if !name.IsValid() {
			return Template{}, unsetIntervalError(op)
		}
	}
	if !open && !cycle.IsValid() {
		return Template{}, unsetIntervalError(op)
	}
	if len(names) == 0 || (open && len(names) == 1) {
		return Template{}, opError(op, ErrEmptyInput)
	}
	if len(names) > MaxPlaces {
		return Template{}, opError(op, ErrInvalidArgument)
	}
	for i := 1; i < len(names); i++ {
		if names[i-1].Equal(names[i]) {
			return Template{}, opError(op, ErrInvalidInterval)
		}
	}

	sequence := make([]coordinates, 0, len(names))
	total := zeroCoordinates()
	for i := 1; i < len(names); i++ {
		current := coordinatesBetweenSpellings(names[i-1], names[i])
		if !isAdjacent(current) {
			return Template{}, opError(op, ErrInvalidInterval)
		}
		sequence = append(sequence, current)
		total = addCoordinates(total, current)
	}
	if !open {
		closing := subCoordinates(coordinatesOf(cycle), total)
		if !isAdjacent(closing) {
			return Template{}, opError(op, ErrInvalidInterval)
		}
		sequence = append(sequence, closing)
	}

	adjacent := make([]interval.Interval, len(sequence))
	for i, current := range sequence {
		converted, err := intervalFromCoordinates(current, op)
		if err != nil {
			return Template{}, err
		}
		adjacent[i] = converted
	}
	if err := validateDerivedIntervals(adjacent, open, op); err != nil {
		return Template{}, err
	}
	return Template{intervals: adjacent, open: open}, nil
}
