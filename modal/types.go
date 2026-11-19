package modal

import (
	"strconv"

	"github.com/go-muse/musex/interval"
)

// Category is a relative modal position, not a numeric weight.
type Category uint8

const (
	// CategoryUnset denotes an absent result.
	CategoryUnset Category = iota
	// High denotes a major or augmented interval from the first degree.
	High
	// Neutral denotes a perfect interval from the first degree.
	Neutral
	// Low denotes a minor or diminished interval from the first degree.
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
