package tempomap

import (
	"math/big"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
)

func secondsResultFromLogForm(op string, form logForm, approximate bool, places int, mode rounding.Mode) (SecondsResult, error) {
	if !approximate {
		value, err := secondsFromRat(op, form.constant)
		if err != nil {
			return SecondsResult{}, err
		}
		return exactSecondsResult(value), nil
	}
	rounded := roundExpression(logExpression{form: form}, places, mode)
	value, err := secondsFromRat(op, rounded)
	if err != nil {
		return SecondsResult{}, err
	}
	boundRat, _ := resultBound(places, mode.IsDirected())
	bound, err := secondsFromRat(op, boundRat)
	if err != nil {
		return SecondsResult{}, err
	}
	return approximateSecondsResult(value, places, bound), nil
}

func positionResultFromExpression(op string, expression numericExpression, approximate bool, exact *big.Rat, places int, mode rounding.Mode) (PositionResult, error) {
	if !approximate {
		value, err := positionFromRat(op, exact)
		if err != nil {
			return PositionResult{}, err
		}
		return exactPositionResult(value), nil
	}
	rounded := roundExpression(expression, places, mode)
	value, err := positionFromRat(op, rounded)
	if err != nil {
		return PositionResult{}, err
	}
	boundRat, _ := resultBound(places, mode.IsDirected())
	bound, err := positionFromRat(op, boundRat)
	if err != nil {
		return PositionResult{}, err
	}
	return approximatePositionResult(value, places, bound), nil
}

func (m Map) secondsAt(at Address, end bool, places int, mode rounding.Mode, grid []measure.Grid) (SecondsResult, error) {
	op := "Map.Start"
	if end {
		op = "Map.End"
	}
	if !m.IsSet() {
		return SecondsResult{}, simpleError(ErrUnset, op, "receiver")
	}
	if !at.IsSet() {
		return SecondsResult{}, simpleError(ErrUnset, op, "address")
	}
	g, err := m.gridArgument(op, grid)
	if err != nil {
		return SecondsResult{}, err
	}
	if err = checkPrecision(op, places, mode); err != nil {
		return SecondsResult{}, err
	}
	position, err := addressPosition(op, at, m.addressing, g)
	if err != nil {
		return SecondsResult{}, err
	}
	idx, err := m.positionIndex(grid)
	if err != nil {
		return SecondsResult{}, err
	}
	if unfinishedAt(idx, position) {
		return SecondsResult{}, simpleError(ErrUnfinished, op, "unfinished gradual segment")
	}
	form, approximate := startFormAt(idx, position, end)
	return secondsResultFromLogForm(op, form, approximate, places, mode)
}

// Start returns seconds at the beginning of at, excluding its own hold.
func (m Map) Start(at Address, places int, mode rounding.Mode, grid ...measure.Grid) (SecondsResult, error) {
	return m.secondsAt(at, false, places, mode, grid)
}

// MustStart is like Start but panics with its error.
func (m Map) MustStart(at Address, places int, mode rounding.Mode, grid ...measure.Grid) SecondsResult {
	return must(m.Start(at, places, mode, grid...))
}

// End returns seconds at the end of the hold at at.
func (m Map) End(at Address, places int, mode rounding.Mode, grid ...measure.Grid) (SecondsResult, error) {
	return m.secondsAt(at, true, places, mode, grid)
}

// MustEnd is like End but panics with its error.
func (m Map) MustEnd(at Address, places int, mode rounding.Mode, grid ...measure.Grid) SecondsResult {
	return must(m.End(at, places, mode, grid...))
}

func approximateBetween(idx *mapIndex, from, to duration.Position) bool {
	for i, current := range idx.changes {
		if current.change.kind != ChangeGradual || i+1 >= len(idx.changes) || current.startSpeed.Equal(idx.changes[i+1].startSpeed) {
			continue
		}
		start, end := current.position, idx.changes[i+1].position
		left := from
		if start.Compare(left) > 0 {
			left = start
		}
		right := to
		if end.Compare(right) < 0 {
			right = end
		}
		if left.Compare(right) < 0 {
			return true
		}
	}
	return false
}

// Between returns Start(to)-Start(from), including the hold at from and excluding to's.
func (m Map) Between(from, to Address, places int, mode rounding.Mode, grid ...measure.Grid) (SecondsResult, error) {
	const op = "Map.Between"
	if !m.IsSet() {
		return SecondsResult{}, simpleError(ErrUnset, op, "receiver")
	}
	if !from.IsSet() {
		return SecondsResult{}, simpleError(ErrUnset, op, "from address")
	}
	if !to.IsSet() {
		return SecondsResult{}, simpleError(ErrUnset, op, "to address")
	}
	g, err := m.gridArgument(op, grid)
	if err != nil {
		return SecondsResult{}, err
	}
	if err = checkPrecision(op, places, mode); err != nil {
		return SecondsResult{}, err
	}
	fromPosition, err := addressPosition(op, from, m.addressing, g)
	if err != nil {
		return SecondsResult{}, err
	}
	toPosition, err := addressPosition(op, to, m.addressing, g)
	if err != nil {
		return SecondsResult{}, err
	}
	if fromPosition.Compare(toPosition) > 0 {
		return SecondsResult{}, simpleError(ErrInvalidArgument, op, "from follows to")
	}
	idx, err := m.positionIndex(grid)
	if err != nil {
		return SecondsResult{}, err
	}
	if unfinishedAt(idx, toPosition) {
		return SecondsResult{}, simpleError(ErrUnfinished, op, "unfinished gradual segment")
	}
	if fromPosition.Compare(toPosition) == 0 {
		return exactSecondsResult(tempo.ZeroSeconds()), nil
	}
	fromForm, _ := startFormAt(idx, fromPosition, false)
	toForm, _ := startFormAt(idx, toPosition, false)
	form := logFormSub(toForm, fromForm)
	return secondsResultFromLogForm(op, form, approximateBetween(idx, fromPosition, toPosition), places, mode)
}

// MustBetween is like Between but panics with its error.
func (m Map) MustBetween(from, to Address, places int, mode rounding.Mode, grid ...measure.Grid) SecondsResult {
	return must(m.Between(from, to, places, mode, grid...))
}

func positionFromSeconds(idx *mapIndex, seconds tempo.Seconds, places int, mode rounding.Mode) (PositionResult, error) {
	const op = "tempomap.positionFromSeconds"
	time := ratOfSeconds(seconds)
	if time.Sign() == 0 {
		return exactPositionResult(duration.Origin()), nil
	}
	lo, hi := 0, len(idx.events)
	for lo < hi {
		middle := int(uint(lo+hi) / 2)
		if compareLogForm(idx.events[middle].start, time) > 0 {
			hi = middle
		} else {
			lo = middle + 1
		}
	}
	event := idx.events[lo-1]
	if compareLogForm(event.start, time) == 0 {
		return exactPositionResult(event.position), nil
	}
	if event.hasExtension && compareLogForm(event.end, time) >= 0 {
		return exactPositionResult(event.position), nil
	}
	changeIndex := changeAtPosition(idx, event.position)
	current := idx.changes[changeIndex]
	if changeIndex+1 == len(idx.changes) && current.change.kind == ChangeGradual {
		return PositionResult{}, simpleError(ErrUnfinished, op, "unfinished gradual segment")
	}
	startSeconds, _ := startFormAt(idx, current.position, false)
	holdsAfterStart := new(big.Rat).Sub(extensionSumAt(idx, event.position, true), extensionSumAt(idx, current.position, true))
	residual := logFormSub(rationalLogForm(time), logFormAddRat(startSeconds, holdsAfterStart))
	approximate := current.approximateBefore
	if current.change.kind == ChangeGradual && !current.startSpeed.Equal(idx.changes[changeIndex+1].startSpeed) {
		next := idx.changes[changeIndex+1]
		length := new(big.Rat).Sub(ratOfPosition(next.position), ratOfPosition(current.position))
		delta := new(big.Rat).Sub(ratOfFraction(next.startSpeed), ratOfFraction(current.startSpeed))
		x := logFormScale(residual, new(big.Rat).Quo(delta, new(big.Rat).Mul(big.NewRat(60, 1), length)))
		b := new(big.Rat).Mul(ratOfFraction(current.startSpeed), length)
		b.Quo(b, delta)
		ratio := new(big.Rat).Quo(ratOfFraction(next.startSpeed), ratOfFraction(current.startSpeed))
		minExp := new(big.Rat).Set(big.NewRat(1, 1))
		maxExp := new(big.Rat).Set(ratio)
		if ratio.Cmp(big.NewRat(1, 1)) < 0 {
			minExp, maxExp = maxExp, minExp
		}
		expression := expForm{a: ratOfPosition(current.position), b: b, x: x, minExp: minExp, maxExp: maxExp}
		return positionResultFromExpression(op, expression, true, nil, places, mode)
	}
	speedOver60 := new(big.Rat).Quo(ratOfFraction(current.startSpeed), big.NewRat(60, 1))
	position := logFormAddRat(logFormScale(residual, speedOver60), ratOfPosition(current.position))
	return positionResultFromExpression(op, logExpression{form: position}, approximate, position.constant, places, mode)
}

// Position returns the position at seconds in a positional map.
func (m Map) Position(seconds tempo.Seconds, places int, mode rounding.Mode) (PositionResult, error) {
	const op = "Map.Position"
	if !m.IsSet() {
		return PositionResult{}, simpleError(ErrUnset, op, "receiver")
	}
	if !seconds.IsSet() {
		return PositionResult{}, dependencyArgumentError(op, tempo.ErrUnset)
	}
	if err := checkPrecision(op, places, mode); err != nil {
		return PositionResult{}, err
	}
	if ratOfSeconds(seconds).Sign() < 0 {
		return PositionResult{}, simpleError(ErrNegativeTime, op, "negative seconds")
	}
	if m.addressing != PositionAddressing {
		return PositionResult{}, simpleError(ErrInvalidArgument, op, "measure map requires AddressAt")
	}
	idx, err := m.positionIndex(nil)
	if err != nil {
		return PositionResult{}, err
	}
	return positionFromSeconds(idx, seconds, places, mode)
}

// MustPosition is like Position but panics with its error.
func (m Map) MustPosition(seconds tempo.Seconds, places int, mode rounding.Mode) PositionResult {
	return must(m.Position(seconds, places, mode))
}

// AddressAt returns the grid address of the rounded position at seconds.
func (m Map) AddressAt(grid measure.Grid, seconds tempo.Seconds, places int, mode rounding.Mode) (measure.Address, PositionResult, error) {
	const op = "Map.AddressAt"
	if !m.IsSet() {
		return measure.Address{}, PositionResult{}, simpleError(ErrUnset, op, "receiver")
	}
	if !grid.IsSet() {
		return measure.Address{}, PositionResult{}, dependencyArgumentError(op, measure.ErrUnset)
	}
	if !seconds.IsSet() {
		return measure.Address{}, PositionResult{}, dependencyArgumentError(op, tempo.ErrUnset)
	}
	if err := checkPrecision(op, places, mode); err != nil {
		return measure.Address{}, PositionResult{}, err
	}
	if ratOfSeconds(seconds).Sign() < 0 {
		return measure.Address{}, PositionResult{}, simpleError(ErrNegativeTime, op, "negative seconds")
	}
	if m.addressing != MeasureAddressing {
		return measure.Address{}, PositionResult{}, simpleError(ErrInvalidArgument, op, "positional map requires Position")
	}
	idx, err := m.positionIndex([]measure.Grid{grid})
	if err != nil {
		return measure.Address{}, PositionResult{}, err
	}
	position, err := positionFromSeconds(idx, seconds, places, mode)
	if err != nil {
		return measure.Address{}, PositionResult{}, err
	}
	address, err := grid.Locate(position.value)
	if err != nil {
		return measure.Address{}, PositionResult{}, measureOperationError(op, err)
	}
	return address, position, nil
}

// MustAddressAt is like AddressAt but panics with its error.
func (m Map) MustAddressAt(grid measure.Grid, seconds tempo.Seconds, places int, mode rounding.Mode) (measure.Address, PositionResult) {
	return must2(m.AddressAt(grid, seconds, places, mode))
}
