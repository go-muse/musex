package modal

import (
	"fmt"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
)

// Relative classifies an actual degree using its full interval from degree 1.
// It checks only the requested degree, accepts open and non-octave modes, and
// returns a zero position on error. Virtual degree numbers are not accepted.
func Relative(m mode.Mode, number int) (RelativePosition, error) {
	if err := checkDegrees(m, number); err != nil {
		return RelativePosition{}, err
	}
	i, err := m.DegreeInterval(number)
	if err != nil {
		return RelativePosition{}, fmt.Errorf("relative degree %d: %w", number, err)
	}
	return relativePosition(number, i)
}

func relativePosition(number int, i interval.Interval) (RelativePosition, error) {
	c, err := classify(i.Quality())
	if err != nil {
		return RelativePosition{}, fmt.Errorf("relative degree %d (%s): %w", number, i, err)
	}
	return RelativePosition{Number: number, Category: c, Interval: i}, nil
}

// RelativePositions returns all positions in ascending degree order. It checks
// qualities in that order and returns nil on error, without a partial result.
func RelativePositions(m mode.Mode) ([]RelativePosition, error) {
	if err := checkDegrees(m); err != nil {
		return nil, err
	}
	positions := make([]RelativePosition, m.Len())
	for index, i := range m.DegreeIntervals() {
		p, err := relativePosition(index+1, i)
		if err != nil {
			return nil, err
		}
		positions[index] = p
	}
	return positions, nil
}

// GroupRelative returns the complete relative analysis grouped by category.
// Each group is independently owned, ordered, and non-nil, even when empty.
// Errors return a zero RelativeGroups with nil slices.
func GroupRelative(m mode.Mode) (RelativeGroups, error) {
	positions, err := RelativePositions(m)
	if err != nil {
		return RelativeGroups{}, err
	}
	groups := RelativeGroups{High: make([]int, 0), Neutral: make([]int, 0), Low: make([]int, 0)}
	for _, p := range positions {
		switch p.Category {
		case High:
			groups.High = append(groups.High, p.Number)
		case Neutral:
			groups.Neutral = append(groups.Neutral, p.Number)
		case Low:
			groups.Low = append(groups.Low, p.Number)
		}
	}
	return groups, nil
}
