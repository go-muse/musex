package modal

import (
	"iter"

	"github.com/go-muse/musex/mode"
)

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// MustRelative calls Relative and panics with its error on failure.
func MustRelative(m mode.Mode, number int) RelativePosition { return must(Relative(m, number)) }

// MustRelativePositions calls RelativePositions and panics with its error on failure.
func MustRelativePositions(m mode.Mode) []RelativePosition { return must(RelativePositions(m)) }

// MustGroupRelative calls GroupRelative and panics with its error on failure.
func MustGroupRelative(m mode.Mode) RelativeGroups { return must(GroupRelative(m)) }

// MustAbsolute calls Absolute and panics with its error on failure.
func MustAbsolute(m mode.Mode, number int) AbsolutePosition { return must(Absolute(m, number)) }

// MustAbsolutePositions calls AbsolutePositions and panics with its error on failure.
func MustAbsolutePositions(m mode.Mode) []AbsolutePosition { return must(AbsolutePositions(m)) }

// MustOrder calls Order and panics with its error on failure.
func MustOrder(m mode.Mode) []AbsoluteGroup { return must(Order(m)) }

// MustCompareAbsolute calls CompareAbsolute and panics with its error on failure.
func MustCompareAbsolute(m mode.Mode, first, second int) int {
	return must(CompareAbsolute(m, first, second))
}

// MustAnalyzeAbsolute calls AnalyzeAbsolute and panics with its error on failure.
func MustAnalyzeAbsolute(m mode.Mode) AbsoluteAnalysis { return must(AnalyzeAbsolute(m)) }

// MustRelationTo calls RelationTo and panics with its error on failure.
func MustRelationTo(m mode.Mode, from, to int) Relation { return must(RelationTo(m, from, to)) }

// MustRelationsFrom calls RelationsFrom and panics with its error on failure.
func MustRelationsFrom(m mode.Mode, from int) []Relation { return must(RelationsFrom(m, from)) }

// MustRelationTable calls RelationTable and panics with its error on failure.
func MustRelationTable(m mode.Mode) iter.Seq[[]Relation] { return must(RelationTable(m)) }

// MustCompareRelativePositions calls CompareRelativePositions and panics with
// its error on failure.
func MustCompareRelativePositions(before, after mode.Mode) RelativeComparison {
	return must(CompareRelativePositions(before, after))
}

// MustComparePositions calls ComparePositions and panics with its error on failure.
func MustComparePositions(before, after mode.Mode) PositionComparison {
	return must(ComparePositions(before, after))
}
