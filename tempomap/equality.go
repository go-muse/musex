package tempomap

import (
	"math/big"
	"sort"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
)

// ToPositions converts a measure map to the same records at exact grid positions.
func (m Map) ToPositions(grid measure.Grid) (Map, error) {
	const op = "Map.ToPositions"
	if !m.IsSet() {
		return Map{}, simpleError(ErrUnset, op, "receiver")
	}
	if !grid.IsSet() {
		return Map{}, dependencyArgumentError(op, measure.ErrUnset)
	}
	if m.addressing != MeasureAddressing {
		return Map{}, simpleError(ErrInvalidArgument, op, "map is not measure-addressed")
	}
	idx, err := m.positionIndex([]measure.Grid{grid})
	if err != nil {
		return Map{}, err
	}
	changes := make([]Change, len(idx.changes))
	for i, current := range idx.changes {
		changes[i] = current.change
		changes[i].at, _ = At(current.position)
	}
	extensions := make([]Extension, len(idx.extensions))
	for i, current := range idx.extensions {
		extensions[i] = current.extension
		extensions[i].at, _ = At(current.position)
	}
	return (Map{addressing: PositionAddressing}).withRecords(op, changes, extensions, nil)
}

// MustToPositions is like ToPositions but panics with its error.
func (m Map) MustToPositions(grid measure.Grid) Map { return must(m.ToPositions(grid)) }

// ToMeasures converts a positional map to valid addresses in grid.
func (m Map) ToMeasures(grid measure.Grid) (Map, error) {
	const op = "Map.ToMeasures"
	if !m.IsSet() {
		return Map{}, simpleError(ErrUnset, op, "receiver")
	}
	if !grid.IsSet() {
		return Map{}, dependencyArgumentError(op, measure.ErrUnset)
	}
	if m.addressing != PositionAddressing {
		return Map{}, simpleError(ErrInvalidArgument, op, "map is not position-addressed")
	}
	idx, err := m.positionIndex(nil)
	if err != nil {
		return Map{}, err
	}
	changes := make([]Change, len(idx.changes))
	for i, current := range idx.changes {
		bar, locateErr := grid.Locate(current.position)
		if locateErr != nil {
			return Map{}, measureOperationError(op, locateErr)
		}
		changes[i] = current.change
		changes[i].at, _ = AtBar(bar)
	}
	extensions := make([]Extension, len(idx.extensions))
	for i, current := range idx.extensions {
		bar, locateErr := grid.Locate(current.position)
		if locateErr != nil {
			return Map{}, measureOperationError(op, locateErr)
		}
		extensions[i] = current.extension
		extensions[i].at, _ = AtBar(bar)
	}
	return (Map{addressing: MeasureAddressing}).withRecords(op, changes, extensions, []measure.Grid{grid})
}

// MustToMeasures is like ToMeasures but panics with its error.
func (m Map) MustToMeasures(grid measure.Grid) Map { return must(m.ToMeasures(grid)) }

func timelineConversion(m Map, grid *measure.Grid) (Map, error) {
	if m.addressing == PositionAddressing {
		if grid != nil {
			return Map{}, simpleError(ErrInvalidArgument, "Map.EqualTimeline", "positional map takes no grid")
		}
		return m, nil
	}
	if grid == nil {
		return Map{}, simpleError(ErrInvalidArgument, "Map.EqualTimeline", "measure map requires a grid")
	}
	return m.ToPositions(*grid)
}

func timelineSpeedsEqual(a, b *mapIndex, position duration.Position) (bool, error) {
	aSpeed, err := speedAtIndex(a, position)
	if err != nil {
		return false, err
	}
	bSpeed, err := speedAtIndex(b, position)
	if err != nil {
		return false, err
	}
	return aSpeed.Equal(bSpeed), nil
}

func uniqueChangePositions(maps ...*mapIndex) []duration.Position {
	var positions []duration.Position
	for _, idx := range maps {
		for _, change := range idx.changes {
			positions = append(positions, change.position)
		}
	}
	sort.Slice(positions, func(i, j int) bool { return positions[i].Compare(positions[j]) < 0 })
	out := positions[:0]
	for _, position := range positions {
		if len(out) == 0 || out[len(out)-1].Compare(position) != 0 {
			out = append(out, position)
		}
	}
	return out
}

// EqualTimeline compares speed functions and fermata holds after conversion to positions.
func (m Map) EqualTimeline(other Map, grid ...measure.Grid) (bool, error) {
	const op = "Map.EqualTimeline"
	if !m.IsSet() {
		return false, simpleError(ErrUnset, op, "receiver")
	}
	if !other.IsSet() {
		return false, simpleError(ErrUnset, op, "argument")
	}
	if len(grid) > 2 {
		return false, simpleError(ErrInvalidArgument, op, "too many grids")
	}
	for _, g := range grid {
		if !g.IsSet() {
			return false, dependencyArgumentError(op, measure.ErrUnset)
		}
	}
	var leftGrid, rightGrid *measure.Grid
	switch len(grid) {
	case 0:
		if m.addressing != PositionAddressing || other.addressing != PositionAddressing {
			return false, simpleError(ErrInvalidArgument, op, "measure maps require grids")
		}
	case 1:
		if m.addressing == PositionAddressing && other.addressing == PositionAddressing {
			return false, simpleError(ErrInvalidArgument, op, "a grid is not used for two positional maps")
		}
		if m.addressing == MeasureAddressing {
			leftGrid = &grid[0]
		}
		if other.addressing == MeasureAddressing {
			rightGrid = &grid[0]
		}
	case 2:
		if m.addressing != MeasureAddressing || other.addressing != MeasureAddressing {
			return false, simpleError(ErrInvalidArgument, op, "two grids require two measure maps")
		}
		leftGrid, rightGrid = &grid[0], &grid[1]
	}
	left, err := timelineConversion(m, leftGrid)
	if err != nil {
		return false, err
	}
	right, err := timelineConversion(other, rightGrid)
	if err != nil {
		return false, err
	}
	unfinished, _ := left.Unfinished()
	if unfinished {
		return false, simpleError(ErrUnfinished, op, "left map has an unfinished gradual tail")
	}
	unfinished, _ = right.Unfinished()
	if unfinished {
		return false, simpleError(ErrUnfinished, op, "right map has an unfinished gradual tail")
	}
	positions := uniqueChangePositions(left.index, right.index)
	for i, position := range positions {
		equal, speedErr := timelineSpeedsEqual(left.index, right.index, position)
		if speedErr != nil || !equal {
			return false, speedErr
		}
		var next duration.Position
		if i+1 < len(positions) {
			next = positions[i+1]
		} else {
			maxPosition, rangeErr := positionFromRat(op, new(big.Rat).SetInt(fraction.MaxComponentBig()))
			if rangeErr != nil {
				return false, rangeErr
			}
			next = maxPosition
			one := ratOfPosition(position)
			one.Add(one, big.NewRat(1, 1))
			if one.Cmp(ratOfPosition(next)) < 0 {
				next, rangeErr = positionFromRat(op, one)
				if rangeErr != nil {
					return false, rangeErr
				}
			}
		}
		midpoint := new(big.Rat).Add(ratOfPosition(position), ratOfPosition(next))
		midpoint.Quo(midpoint, big.NewRat(2, 1))
		if midpoint.Cmp(ratOfPosition(position)) == 0 {
			continue
		}
		mid, rangeErr := positionFromRat(op, midpoint)
		if rangeErr != nil {
			return false, rangeErr
		}
		equal, speedErr = timelineSpeedsEqual(left.index, right.index, mid)
		if speedErr != nil || !equal {
			return false, speedErr
		}
	}
	var extensionPositions []duration.Position
	for _, idx := range []*mapIndex{left.index, right.index} {
		for _, extension := range idx.extensions {
			extensionPositions = append(extensionPositions, extension.position)
		}
	}
	sort.Slice(extensionPositions, func(i, j int) bool { return extensionPositions[i].Compare(extensionPositions[j]) < 0 })
	for i, position := range extensionPositions {
		if i > 0 && extensionPositions[i-1].Compare(position) == 0 {
			continue
		}
		leftSeconds, leftOK := extensionAt(left.index, position)
		rightSeconds, rightOK := extensionAt(right.index, position)
		if !leftOK {
			leftSeconds = tempo.ZeroSeconds()
		}
		if !rightOK {
			rightSeconds = tempo.ZeroSeconds()
		}
		if !leftSeconds.Equal(rightSeconds) {
			return false, nil
		}
	}
	return true, nil
}

// MustEqualTimeline is like EqualTimeline but panics with its error.
func (m Map) MustEqualTimeline(other Map, grid ...measure.Grid) bool {
	return must(m.EqualTimeline(other, grid...))
}
