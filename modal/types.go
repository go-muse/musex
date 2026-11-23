package modal

import (
	"strconv"

	"github.com/go-muse/musex/interval"
)

// Category is a relative modal position, not a numeric weight. It classifies
// the quality of an upward interval from a source degree: from degree 1 for a
// relative position, and from any selected degree for a Relation.
type Category uint8

const (
	// CategoryUnset denotes an absent result.
	CategoryUnset Category = iota
	// High denotes a major or augmented interval from the source degree.
	High
	// Neutral denotes a perfect interval from the source degree.
	Neutral
	// Low denotes a minor or diminished interval from the source degree.
	Low
)

// IsValid reports whether c is High, Neutral, or Low.
func (c Category) IsValid() bool { return c >= High && c <= Low }

// String returns the English category name, or Category(n) for an unknown value.
func (c Category) String() string {
	switch c {
	case CategoryUnset:
		return "unset"
	case High:
		return "high"
	case Neutral:
		return "neutral"
	case Low:
		return "low"
	default:
		return "Category(" + strconv.Itoa(int(c)) + ")"
	}
}

// RelativePosition retains the complete specific interval from the first degree.
// The zero value denotes an absent result. Values can be copied and compared.
type RelativePosition struct {
	Number   int
	Category Category
	Interval interval.Interval
}

// RelativeGroups groups degree numbers in ascending order within each category.
// Successful results have non-nil slices, including empty categories. Assignment
// shares slice elements; independent calls and different groups do not.
type RelativeGroups struct {
	High    []int
	Neutral []int
	Low     []int
}

// QualityCounts counts intervals by family, counting each multiplicity once.
type QualityCounts struct {
	Diminished int
	Minor      int
	Perfect    int
	Major      int
	Augmented  int
}

// AbsolutePosition describes one degree in the complete absolute analysis.
// Score is Minor + Diminished - Major - Augmented; larger scores are brighter.
// Rank is the dense group number, starting at 1. The zero value is absent.
type AbsolutePosition struct {
	Number int
	Counts QualityCounts
	Score  int
	Rank   int
}

// AbsoluteGroup contains equal-score degrees in ascending numeric order.
// Numbers belongs to the caller; assignment shares its elements.
type AbsoluteGroup struct {
	Rank    int
	Score   int
	Numbers []int
}

// AbsoluteAnalysis contains positions in degree order and groups from brightest
// to darkest. All slices belong to the caller. Different groups, calls, and the
// input mode share no mutable result data. Assignment shares slice elements.
// The zero value denotes an absent result and contains nil slices.
type AbsoluteAnalysis struct {
	Positions []AbsolutePosition
	Order     []AbsoluteGroup
}

// Relation is the upward relation of a source degree to a target degree: the
// full specific interval from From to To in the cycle given by Cycle, and the
// category of that interval. Both numbers are actual degree numbers 1..Len;
// Cycle 0 is the same cycle and 1 the next one. The interval is compound when
// the span is, and is never reduced to an octave. Relation is a comparable
// value: == compares all fields and agrees with Interval.Equal for results of
// this package, and it can be a map key. The zero value denotes an absent
// result. The package does not accept a Relation as input, retains no mode,
// and promises no consistency of fields changed by the caller.
type Relation struct {
	From     int               // source degree, 1..k
	To       int               // target degree in the mode's own numbering, 1..k
	Cycle    int               // 0: same cycle, 1: next cycle
	Interval interval.Interval // full specific interval upward from From to To
	Category Category          // category of Interval
}

// RelativeChange compares the relative positions of one degree number in two
// modes. IntervalChanged reports that the specific intervals from degree 1
// differ as records (kind, deviation, and direction), so A4 and d5 differ;
// CategoryChanged reports different categories. The flags are computed by
// the package. RelativeChange is comparable; its zero value is an absent
// result with no flags set.
type RelativeChange struct {
	Number          int
	Before          RelativePosition
	After           RelativePosition
	IntervalChanged bool
	CategoryChanged bool
}

// RelativeComparison contains one change per degree number in ascending
// order. It holds a slice, so it is not comparable with ==; assignment shares
// the elements, as usual in Go. The zero value denotes an absent result.
type RelativeComparison struct {
	Degrees []RelativeChange
}

// IntervalChanges returns the numbers of the degrees whose IntervalChanged is
// set, in the order of Degrees. The flags are read, not recomputed. The result
// is a new non-nil slice, empty when nothing changed, including for the zero
// value.
func (c RelativeComparison) IntervalChanges() []int {
	return changedNumbers(c.Degrees, func(d RelativeChange) (int, bool) { return d.Number, d.IntervalChanged })
}

// CategoryChanges returns the numbers of the degrees whose CategoryChanged is
// set, as IntervalChanges does.
func (c RelativeComparison) CategoryChanges() []int {
	return changedNumbers(c.Degrees, func(d RelativeChange) (int, bool) { return d.Number, d.CategoryChanged })
}

// DegreePositions contains the relative and the absolute position of one
// degree in one mode. It is comparable; the zero value is an absent result.
type DegreePositions struct {
	Relative RelativePosition
	Absolute AbsolutePosition
}

// PositionChange compares all positions of one degree number in two modes.
// IntervalChanged and CategoryChanged compare the relative positions as in
// RelativeChange; ScoreChanged and RankChanged compare Absolute.Score and
// Absolute.Rank. The distribution of qualities has no flag of its own:
// compare Absolute.Counts. PositionChange is comparable; its zero value is an
// absent result with no flags set.
type PositionChange struct {
	Number          int
	Before          DegreePositions
	After           DegreePositions
	IntervalChanged bool
	CategoryChanged bool
	ScoreChanged    bool
	RankChanged     bool
}

// PositionComparison contains one change per degree number in ascending
// order. It holds a slice, so it is not comparable with ==; assignment shares
// the elements, as usual in Go. The zero value denotes an absent result.
type PositionComparison struct {
	Degrees []PositionChange
}

// IntervalChanges returns the numbers of the degrees whose IntervalChanged is
// set, in the order of Degrees. The flags are read, not recomputed. The result
// is a new non-nil slice, empty when nothing changed, including for the zero
// value.
func (c PositionComparison) IntervalChanges() []int {
	return changedNumbers(c.Degrees, func(d PositionChange) (int, bool) { return d.Number, d.IntervalChanged })
}

// CategoryChanges returns the numbers of the degrees whose CategoryChanged is
// set, as IntervalChanges does.
func (c PositionComparison) CategoryChanges() []int {
	return changedNumbers(c.Degrees, func(d PositionChange) (int, bool) { return d.Number, d.CategoryChanged })
}

// ScoreChanges returns the numbers of the degrees whose ScoreChanged is set,
// as IntervalChanges does.
func (c PositionComparison) ScoreChanges() []int {
	return changedNumbers(c.Degrees, func(d PositionChange) (int, bool) { return d.Number, d.ScoreChanged })
}

// RankChanges returns the numbers of the degrees whose RankChanged is set, as
// IntervalChanges does.
func (c PositionComparison) RankChanges() []int {
	return changedNumbers(c.Degrees, func(d PositionChange) (int, bool) { return d.Number, d.RankChanged })
}

func changedNumbers[T any](degrees []T, flag func(T) (int, bool)) []int {
	result := make([]int, 0)
	for _, d := range degrees {
		if number, changed := flag(d); changed {
			result = append(result, number)
		}
	}
	return result
}
