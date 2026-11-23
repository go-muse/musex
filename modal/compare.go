package modal

import (
	"fmt"

	"github.com/go-muse/musex/mode"
)

// CompareRelativePositions compares the relative positions of two modes degree
// by degree. Degree j of before is matched with degree j of after; degree 1 of
// both is the common origin. Equal numbers of degrees suffice: the kinds,
// cycles, and letter sizes of the intervals need not match, so open modes and
// modes with any cycle are accepted. Degrees has one change per number
// 1..Len() in ascending order; Before and After equal the elements of
// RelativePositions of each mode, and the flags compare them. Comparing a mode
// with an equal mode sets no flag; swapping the arguments swaps Before and
// After and keeps the flags.
//
// Errors, in this order: ErrUnset together with mode.ErrUnset for an unset
// before, then after; ErrIncomparable for different numbers of degrees; the
// error RelativePositions gives for before, then for after (only
// ErrUnsupportedQuality is possible there). On error the result is the zero
// value with nil Degrees, without a partial result. The dependency error is
// preserved. Time and memory are O(Len()).
func CompareRelativePositions(before, after mode.Mode) (RelativeComparison, error) {
	if err := checkComparison(before, after, false); err != nil {
		return RelativeComparison{}, err
	}
	first, err := RelativePositions(before)
	if err != nil {
		return RelativeComparison{}, fmt.Errorf("compare before: %w", err)
	}
	second, err := RelativePositions(after)
	if err != nil {
		return RelativeComparison{}, fmt.Errorf("compare after: %w", err)
	}
	degrees := make([]RelativeChange, len(first))
	for index := range degrees {
		b, a := first[index], second[index]
		degrees[index] = RelativeChange{
			Number:          index + 1,
			Before:          b,
			After:           a,
			IntervalChanged: !b.Interval.Equal(a.Interval),
			CategoryChanged: b.Category != a.Category,
		}
	}
	return RelativeComparison{Degrees: degrees}, nil
}

// ComparePositions compares the relative and absolute positions of two closed
// perfect-octave modes degree by degree, matched by number as in
// CompareRelativePositions. Before.Absolute and After.Absolute equal the
// positions of AnalyzeAbsolute of each mode; each rank is counted within its
// own mode, and equal scores share a rank. The four flags compare the two
// sides of each degree.
//
// Errors, in this order: ErrUnset together with mode.ErrUnset for an unset
// before, then after; ErrUnsupportedStructure when before, then after, is not
// closed with a perfect-octave cycle; ErrIncomparable for different numbers of
// degrees; the error AnalyzeAbsolute gives for before, then for after, with
// representability of the whole table before qualities. An unrepresentable
// interval matches mode.ErrOutOfRange and interval.ErrOutOfRange; other causes
// are not promised. Structure is checked before comparability, as in mode. On
// error the result is the zero value with nil Degrees. Time is
// O(Len()² + Len() log Len()) for each mode, memory O(Len()).
func ComparePositions(before, after mode.Mode) (PositionComparison, error) {
	if err := checkComparison(before, after, true); err != nil {
		return PositionComparison{}, err
	}
	first, err := AnalyzeAbsolute(before)
	if err != nil {
		return PositionComparison{}, fmt.Errorf("compare before: %w", err)
	}
	second, err := AnalyzeAbsolute(after)
	if err != nil {
		return PositionComparison{}, fmt.Errorf("compare after: %w", err)
	}
	// Absolute analysis has accepted every interval from degree 1, so the
	// relative positions of both modes are available.
	relativeBefore, err := RelativePositions(before)
	if err != nil {
		return PositionComparison{}, fmt.Errorf("compare before: %w", err)
	}
	relativeAfter, err := RelativePositions(after)
	if err != nil {
		return PositionComparison{}, fmt.Errorf("compare after: %w", err)
	}
	degrees := make([]PositionChange, len(relativeBefore))
	for index := range degrees {
		b := DegreePositions{Relative: relativeBefore[index], Absolute: first.Positions[index]}
		a := DegreePositions{Relative: relativeAfter[index], Absolute: second.Positions[index]}
		degrees[index] = PositionChange{
			Number:          index + 1,
			Before:          b,
			After:           a,
			IntervalChanged: !b.Relative.Interval.Equal(a.Relative.Interval),
			CategoryChanged: b.Relative.Category != a.Relative.Category,
			ScoreChanged:    b.Absolute.Score != a.Absolute.Score,
			RankChanged:     b.Absolute.Rank != a.Absolute.Rank,
		}
	}
	return PositionComparison{Degrees: degrees}, nil
}

// checkComparison checks both modes for being set, then, for a full
// comparison, their structure, then the numbers of degrees.
func checkComparison(before, after mode.Mode, full bool) error {
	if err := checkDegrees(before); err != nil {
		return fmt.Errorf("compare before: %w", err)
	}
	if err := checkDegrees(after); err != nil {
		return fmt.Errorf("compare after: %w", err)
	}
	if full {
		if err := checkOctaveCycle(before, "full comparison of before"); err != nil {
			return err
		}
		if err := checkOctaveCycle(after, "full comparison of after"); err != nil {
			return err
		}
	}
	if before.Len() != after.Len() {
		return fmt.Errorf("%w: %d and %d degrees", ErrIncomparable, before.Len(), after.Len())
	}
	return nil
}
