package modal

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
)

// Absolute returns a degree's counts, score, and rank in the complete analysis.
// It validates the entire mode, including intervals from other degrees.
// Errors return a zero position.
func Absolute(m mode.Mode, number int) (AbsolutePosition, error) {
	analysis, err := analyzeAbsolute(m, number)
	if err != nil {
		return AbsolutePosition{}, err
	}
	return analysis.Positions[number-1], nil
}

// AbsolutePositions returns positions in ascending degree order, or nil on error.
// It performs the same complete validation as AnalyzeAbsolute.
func AbsolutePositions(m mode.Mode) ([]AbsolutePosition, error) {
	analysis, err := AnalyzeAbsolute(m)
	if err != nil {
		return nil, err
	}
	return analysis.Positions, nil
}

// Order returns equal-score groups from brightest to darkest, or nil on error.
// Each group's Numbers is independently owned and sorted in ascending order.
func Order(m mode.Mode) ([]AbsoluteGroup, error) {
	analysis, err := AnalyzeAbsolute(m)
	if err != nil {
		return nil, err
	}
	return analysis.Order, nil
}

// CompareAbsolute returns -1 when first is brighter, 1 when darker, and 0 when
// scores match. Even self-comparison validates the complete mode. It checks
// first before second, then structure and intervals; errors return 0.
func CompareAbsolute(m mode.Mode, first, second int) (int, error) {
	analysis, err := analyzeAbsolute(m, first, second)
	if err != nil {
		return 0, err
	}
	return cmp.Compare(analysis.Positions[second-1].Score, analysis.Positions[first-1].Score), nil
}

// AnalyzeAbsolute analyzes every upward relation in a closed perfect-octave
// mode. Each row includes its own P1 once and every other degree once, without
// an additional P8. Integral multiplicities count once in their quality family.
// It checks all interval representability before rejecting unsupported quality.
// On error it returns a zero analysis. Time is O(k² + k log k), space O(k).
func AnalyzeAbsolute(m mode.Mode) (AbsoluteAnalysis, error) {
	return analyzeAbsolute(m)
}

func analyzeAbsolute(m mode.Mode, numbers ...int) (AbsoluteAnalysis, error) {
	if err := checkDegrees(m, numbers...); err != nil {
		return AbsoluteAnalysis{}, err
	}
	if err := checkOctaveCycle(m, "absolute analysis"); err != nil {
		return AbsoluteAnalysis{}, err
	}
	// Intervals checks representability of the whole table before returning its
	// iterator. Quality rejection must follow that check, including in address queries.
	intervals, err := m.Intervals()
	if err != nil {
		return AbsoluteAnalysis{}, fmt.Errorf("absolute interval table: %w", err)
	}
	positions := make([]AbsolutePosition, m.Len())
	for index := range positions {
		positions[index].Number = index + 1
		positions[index].Counts.Perfect = 1
	}
	for at := range intervals {
		number := at.Degree.Number()
		q := at.Interval.Quality()
		if _, err := classify(q); err != nil {
			return AbsoluteAnalysis{}, fmt.Errorf("absolute degree %d, distance %d (%s): %w", number, at.Distance, at.Interval, err)
		}
		counts := &positions[number-1].Counts
		switch q {
		case interval.Diminished:
			counts.Diminished++
		case interval.Minor:
			counts.Minor++
		case interval.Perfect:
			counts.Perfect++
		case interval.Major:
			counts.Major++
		case interval.Augmented:
			counts.Augmented++
		}
	}
	ordered := make([]int, len(positions))
	for index := range positions {
		counts := positions[index].Counts
		positions[index].Score = counts.Minor + counts.Diminished - counts.Major - counts.Augmented
		ordered[index] = index
	}
	slices.SortFunc(ordered, func(a, b int) int {
		if byScore := cmp.Compare(positions[b].Score, positions[a].Score); byScore != 0 {
			return byScore
		}
		return cmp.Compare(a, b)
	})
	groups := make([]AbsoluteGroup, 0)
	for _, index := range ordered {
		p := &positions[index]
		if len(groups) == 0 || groups[len(groups)-1].Score != p.Score {
			groups = append(groups, AbsoluteGroup{Rank: len(groups) + 1, Score: p.Score, Numbers: make([]int, 0)})
		}
		group := &groups[len(groups)-1]
		group.Numbers = append(group.Numbers, p.Number)
		p.Rank = group.Rank
	}
	return AbsoluteAnalysis{Positions: positions, Order: groups}, nil
}
