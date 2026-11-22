package tempomap

import (
	"fmt"
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
)

// checkSecondsStep validates a step in seconds and the rounding mode, in the
// place of the decimal places of the operations without a step.
func checkSecondsStep(op string, step tempo.Seconds, mode rounding.Mode) (*big.Rat, error) {
	if !step.IsSet() {
		return nil, dependencyArgumentError(op, tempo.ErrUnset)
	}
	h := ratOfSeconds(step)
	if h.Sign() <= 0 {
		return nil, simpleError(ErrInvalidArgument, op, "step is not positive")
	}
	if !mode.IsValid() {
		return nil, modeError(op, mode)
	}
	return h, nil
}

// checkPositionStep validates a step in whole notes and the rounding mode.
func checkPositionStep(op string, step duration.Duration, mode rounding.Mode) (*big.Rat, error) {
	if !step.IsSet() {
		return nil, dependencyArgumentError(op, duration.ErrUnset)
	}
	h := ratOfFraction(step.MustFraction())
	if h.Sign() <= 0 {
		return nil, simpleError(ErrInvalidArgument, op, "step is not positive")
	}
	if !mode.IsValid() {
		return nil, modeError(op, mode)
	}
	return h, nil
}

// stepIndex returns the index k = mode(X/step) of the exact value X, given as
// a rational value or, when exact is nil, as an irrational expression.
func stepIndex(op string, exact *big.Rat, expression numericExpression, step *big.Rat, mode rounding.Mode) (int, error) {
	var k *big.Int
	if exact != nil {
		index, _, err := mode.RatMultiple(exact, step)
		if err != nil {
			panic(err)
		}
		k = index
	} else {
		k = roundExpressionSteps(expression, step, mode, stepInitialPrecision(step, expression.inputBits()))
	}
	if !k.IsInt64() {
		return 0, stepRangeError(op)
	}
	return int(k.Int64()), nil
}

func stepRangeError(op string) error {
	return fmt.Errorf("%w: %s: step index exceeds int: %w", ErrOutOfRange, op, fraction.ErrNotRepresentable)
}

// formStepIndex rounds the seconds of a symbolic form to a multiple of step.
func formStepIndex(op string, form logForm, approximate bool, step *big.Rat, mode rounding.Mode) (int, error) {
	if !approximate {
		return stepIndex(op, form.constant, nil, step, mode)
	}
	return stepIndex(op, nil, logExpression{form: form}, step, mode)
}

// StartSteps returns k = mode(X/step), where X is the exact value of
// Start(at): the start of the moment in seconds, including earlier holds. The
// rounded value is k*step. The result is correctly rounded from the exact
// value for every input, including irrational values and exact ties; there is
// no precision limit. Errors and their order are those of Start, with the step
// checked in the place of the decimal places: an unset step wraps
// tempo.ErrUnset and a zero step is ErrInvalidArgument. An index greater than
// math.MaxInt is ErrOutOfRange together with fraction.ErrNotRepresentable.
func (m Map) StartSteps(at Address, step tempo.Seconds, mode rounding.Mode, grid ...measure.Grid) (int, error) {
	return m.momentSteps("Map.StartSteps", at, false, step, mode, grid)
}

// MustStartSteps is like StartSteps but panics with its error.
func (m Map) MustStartSteps(at Address, step tempo.Seconds, mode rounding.Mode, grid ...measure.Grid) int {
	return must(m.StartSteps(at, step, mode, grid...))
}

// EndSteps is StartSteps for End(at): the end of the hold at the moment.
func (m Map) EndSteps(at Address, step tempo.Seconds, mode rounding.Mode, grid ...measure.Grid) (int, error) {
	return m.momentSteps("Map.EndSteps", at, true, step, mode, grid)
}

// MustEndSteps is like EndSteps but panics with its error.
func (m Map) MustEndSteps(at Address, step tempo.Seconds, mode rounding.Mode, grid ...measure.Grid) int {
	return must(m.EndSteps(at, step, mode, grid...))
}

func (m Map) momentSteps(op string, at Address, end bool, step tempo.Seconds, mode rounding.Mode, grid []measure.Grid) (int, error) {
	var h *big.Rat
	check := func() error {
		var err error
		h, err = checkSecondsStep(op, step, mode)
		return err
	}
	form, approximate, err := m.momentForm(op, at, end, check, grid)
	if err != nil {
		return 0, err
	}
	return formStepIndex(op, form, approximate, h, mode)
}

// BetweenSteps is StartSteps for Between(from, to): the seconds from the start
// of from, including its hold, to the start of to.
func (m Map) BetweenSteps(from, to Address, step tempo.Seconds, mode rounding.Mode, grid ...measure.Grid) (int, error) {
	const op = "Map.BetweenSteps"
	var h *big.Rat
	check := func() error {
		var err error
		h, err = checkSecondsStep(op, step, mode)
		return err
	}
	form, approximate, err := m.betweenForm(op, from, to, check, grid)
	if err != nil {
		return 0, err
	}
	return formStepIndex(op, form, approximate, h, mode)
}

// MustBetweenSteps is like BetweenSteps but panics with its error.
func (m Map) MustBetweenSteps(from, to Address, step tempo.Seconds, mode rounding.Mode, grid ...measure.Grid) int {
	return must(m.BetweenSteps(from, to, step, mode, grid...))
}

// PositionSteps returns k = mode(P/step) for the exact position P of
// Position(seconds) of a position map; seconds inside a hold give the exact
// position of the hold. The rounded position is k*step. Errors and their order
// are those of Position, with the step checked in the place of the decimal
// places: an unset step wraps duration.ErrUnset.
func (m Map) PositionSteps(seconds tempo.Seconds, step duration.Duration, mode rounding.Mode) (int, error) {
	const op = "Map.PositionSteps"
	if !m.IsSet() {
		return 0, simpleError(ErrUnset, op, "receiver")
	}
	if !seconds.IsSet() {
		return 0, dependencyArgumentError(op, tempo.ErrUnset)
	}
	h, err := checkPositionStep(op, step, mode)
	if err != nil {
		return 0, err
	}
	if ratOfSeconds(seconds).Sign() < 0 {
		return 0, simpleError(ErrNegativeTime, op, "negative seconds")
	}
	if m.addressing != PositionAddressing {
		return 0, simpleError(ErrInvalidArgument, op, "measure map requires AddressAtSteps")
	}
	idx, err := m.positionIndex(nil)
	if err != nil {
		return 0, err
	}
	exact, expression, err := positionTarget(idx, seconds)
	if err != nil {
		return 0, err
	}
	return stepIndex(op, exact, expression, h, mode)
}

// MustPositionSteps is like PositionSteps but panics with its error.
func (m Map) MustPositionSteps(seconds tempo.Seconds, step duration.Duration, mode rounding.Mode) int {
	return must(m.PositionSteps(seconds, step, mode))
}

// AddressAtSteps is PositionSteps for a measure map: it returns the address of
// the position k*step in grid, as grid.Locate does, and k. Both results refer
// to the same position. Errors and their order are those of AddressAt, with
// the step checked in the place of the decimal places; on error it returns
// measure.Address{} and 0.
func (m Map) AddressAtSteps(grid measure.Grid, seconds tempo.Seconds, step duration.Duration, mode rounding.Mode) (measure.Address, int, error) {
	const op = "Map.AddressAtSteps"
	if !m.IsSet() {
		return measure.Address{}, 0, simpleError(ErrUnset, op, "receiver")
	}
	if !grid.IsSet() {
		return measure.Address{}, 0, dependencyArgumentError(op, measure.ErrUnset)
	}
	if !seconds.IsSet() {
		return measure.Address{}, 0, dependencyArgumentError(op, tempo.ErrUnset)
	}
	h, err := checkPositionStep(op, step, mode)
	if err != nil {
		return measure.Address{}, 0, err
	}
	if ratOfSeconds(seconds).Sign() < 0 {
		return measure.Address{}, 0, simpleError(ErrNegativeTime, op, "negative seconds")
	}
	if m.addressing != MeasureAddressing {
		return measure.Address{}, 0, simpleError(ErrInvalidArgument, op, "positional map requires PositionSteps")
	}
	idx, err := m.positionIndex([]measure.Grid{grid})
	if err != nil {
		return measure.Address{}, 0, err
	}
	exact, expression, err := positionTarget(idx, seconds)
	if err != nil {
		return measure.Address{}, 0, err
	}
	k, err := stepIndex(op, exact, expression, h, mode)
	if err != nil {
		return measure.Address{}, 0, err
	}
	position, err := positionFromRat(op, stepNode(big.NewInt(int64(k)), h))
	if err != nil {
		return measure.Address{}, 0, err
	}
	address, err := grid.Locate(position)
	if err != nil {
		return measure.Address{}, 0, measureOperationError(op, err)
	}
	return address, k, nil
}

// MustAddressAtSteps is like AddressAtSteps but panics with its error.
func (m Map) MustAddressAtSteps(grid measure.Grid, seconds tempo.Seconds, step duration.Duration, mode rounding.Mode) (measure.Address, int) {
	return must2(m.AddressAtSteps(grid, seconds, step, mode))
}

// SecondsSteps is StartSteps for the musical seconds B(b)-B(a) of a bounded
// segment, without holds. For the open final segment it returns 0, false, nil
// after the checks; otherwise k, true, nil.
func (s Segment) SecondsSteps(step tempo.Seconds, mode rounding.Mode) (int, bool, error) {
	const op = "Segment.SecondsSteps"
	if !s.IsSet() {
		return 0, false, simpleError(ErrUnset, op, "receiver")
	}
	h, err := checkSecondsStep(op, step, mode)
	if err != nil {
		return 0, false, err
	}
	if s.open {
		return 0, false, nil
	}
	form, approximate := s.secondsForm()
	k, err := formStepIndex(op, form, approximate, h, mode)
	if err != nil {
		return 0, false, err
	}
	return k, true, nil
}

// MustSecondsSteps is like SecondsSteps but panics with its error.
func (s Segment) MustSecondsSteps(step tempo.Seconds, mode rounding.Mode) (int, bool) {
	return must2(s.SecondsSteps(step, mode))
}

// WeightedInterval contributes Weight times the exact Between(From, To).
// Weight must be non-nil and positive. Copying the record shares the weight
// pointer; pointer equality does not mean numerical interval equality.
type WeightedInterval struct {
	From   Address
	To     Address
	Weight *big.Rat
}

// WeightedBetweenSteps rounds the complete sum of weighted exact Between
// intervals once to the positive step in seconds and returns its integer
// index. Intervals may overlap, repeat, or occur in any order. Empty input
// gives zero after validation. Weights and step are not limited by fraction.
// The call neither changes nor retains any input. Concurrent calls are safe
// provided callers do not concurrently mutate the input slice or rationals.
// Invalid weights or step return ErrInvalidArgument. An index outside int
// returns ErrOutOfRange with fraction.ErrNotRepresentable.
func (m Map) WeightedBetweenSteps(terms []WeightedInterval, step *big.Rat, mode rounding.Mode, grid ...measure.Grid) (int, error) {
	const op = "Map.WeightedBetweenSteps"
	if !m.IsSet() {
		return 0, simpleError(ErrUnset, op, "receiver")
	}
	for _, term := range terms {
		if !term.From.IsSet() {
			return 0, simpleError(ErrUnset, op, "from address")
		}
		if !term.To.IsSet() {
			return 0, simpleError(ErrUnset, op, "to address")
		}
	}
	for _, term := range terms {
		if term.Weight == nil || term.Weight.Sign() <= 0 {
			return 0, simpleError(ErrInvalidArgument, op, "weight is not positive")
		}
	}
	g, err := m.gridArgument(op, grid)
	if err != nil {
		return 0, err
	}
	if step == nil || step.Sign() <= 0 {
		return 0, simpleError(ErrInvalidArgument, op, "step is not positive")
	}
	if !mode.IsValid() {
		return 0, modeError(op, mode)
	}
	type bounds struct{ from, to duration.Position }
	positions := make([]bounds, len(terms))
	for i, term := range terms {
		from, err := addressPosition(op, term.From, m.addressing, g)
		if err != nil {
			return 0, err
		}
		to, err := addressPosition(op, term.To, m.addressing, g)
		if err != nil {
			return 0, err
		}
		if from.Compare(to) > 0 {
			return 0, simpleError(ErrInvalidArgument, op, "from follows to")
		}
		positions[i] = bounds{from: from, to: to}
	}
	if len(terms) == 0 {
		return 0, nil
	}
	idx, err := m.positionIndex(grid)
	if err != nil {
		return 0, err
	}
	form := zeroLogForm()
	approximate := false
	for i, p := range positions {
		if unfinishedAt(idx, p.to) {
			return 0, simpleError(ErrUnfinished, op, "unfinished gradual segment")
		}
		if p.from.Compare(p.to) == 0 {
			continue
		}
		from, _ := startFormAt(idx, p.from, false)
		to, _ := startFormAt(idx, p.to, false)
		form = logFormAdd(form, logFormScale(logFormSub(to, from), terms[i].Weight))
		approximate = approximate || approximateBetween(idx, p.from, p.to)
	}
	return formStepIndex(op, form, approximate, step, mode)
}

// MustWeightedBetweenSteps is like WeightedBetweenSteps but panics with its error.
func (m Map) MustWeightedBetweenSteps(terms []WeightedInterval, step *big.Rat, mode rounding.Mode, grid ...measure.Grid) int {
	return must(m.WeightedBetweenSteps(terms, step, mode, grid...))
}
