package ticks

import (
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
)

func checkGrid(op string, g measure.Grid) error {
	if !g.IsSet() {
		return invalidError(op, "grid", measure.ErrUnset)
	}
	return nil
}

// addressPosition returns the position of the address a in g.
func addressPosition(op string, g measure.Grid, a measure.Address) (duration.Position, error) {
	if !a.IsSet() {
		return duration.Position{}, invalidError(op, "address", measure.ErrUnset)
	}
	p, err := g.PositionOf(a)
	if err != nil {
		return duration.Position{}, dependencyError(op, err)
	}
	return p, nil
}

// AddressTicks returns the tick of the address a in the grid g: the position of
// a in ticks. An address outside the grid is ErrInvalidArgument with the causes
// of measure; a position that is not a whole number of ticks is
// ErrNotRepresentable.
func (r Resolution) AddressTicks(g measure.Grid, a measure.Address) (int, error) {
	const op = "Resolution.AddressTicks"
	if err := r.check(op); err != nil {
		return 0, err
	}
	if err := checkGrid(op, g); err != nil {
		return 0, err
	}
	p, err := addressPosition(op, g, a)
	if err != nil {
		return 0, err
	}
	return wholeTicks(op, r.unitsOf(ratOfPosition(p)))
}

// MustAddressTicks is like AddressTicks but panics with its error.
func (r Resolution) MustAddressTicks(g measure.Grid, a measure.Address) int {
	return must(r.AddressTicks(g, a))
}

// QuantizeAddress rounds the position of the address a in g to whole ticks
// once with mode.
func (r Resolution) QuantizeAddress(g measure.Grid, a measure.Address, mode rounding.Mode) (Quantized, error) {
	const op = "Resolution.QuantizeAddress"
	if err := r.check(op); err != nil {
		return Quantized{}, err
	}
	if err := checkGrid(op, g); err != nil {
		return Quantized{}, err
	}
	if !a.IsSet() {
		return Quantized{}, invalidError(op, "address", measure.ErrUnset)
	}
	if err := checkMode(op, mode); err != nil {
		return Quantized{}, err
	}
	p, err := addressPosition(op, g, a)
	if err != nil {
		return Quantized{}, err
	}
	value := ratOfPosition(p)
	return quantize(op, r.unitsOf(value), value, r.tick(), mode, false)
}

// MustQuantizeAddress is like QuantizeAddress but panics with its error.
func (r Resolution) MustQuantizeAddress(g measure.Grid, a measure.Address, mode rounding.Mode) Quantized {
	return must(r.QuantizeAddress(g, a, mode))
}

// Address returns the address in g of the position of tick >= 0: g.Locate.
func (r Resolution) Address(g measure.Grid, tick int) (measure.Address, error) {
	const op = "Resolution.Address"
	if err := r.check(op); err != nil {
		return measure.Address{}, err
	}
	if err := checkGrid(op, g); err != nil {
		return measure.Address{}, err
	}
	p, err := r.Position(tick)
	if err != nil {
		return measure.Address{}, err
	}
	a, err := g.Locate(p)
	if err != nil {
		return measure.Address{}, dependencyError(op, err)
	}
	return a, nil
}

// MustAddress is like Address but panics with its error.
func (r Resolution) MustAddress(g measure.Grid, tick int) measure.Address {
	return must(r.Address(g, tick))
}

// BarTicks returns the tick of the start of the bar and its length in ticks.
// A bar outside the grid is ErrInvalidArgument with the causes of measure; a
// start or length that is not a whole number of ticks is ErrNotRepresentable.
func (r Resolution) BarTicks(g measure.Grid, bar int) (start, length int, err error) {
	const op = "Resolution.BarTicks"
	if err := r.check(op); err != nil {
		return 0, 0, err
	}
	if err := checkGrid(op, g); err != nil {
		return 0, 0, err
	}
	if bar < 0 {
		return 0, 0, invalidError(op, "negative bar")
	}
	p, err := g.BarStart(bar)
	if err != nil {
		return 0, 0, dependencyError(op, err)
	}
	d, err := g.BarLength(bar)
	if err != nil {
		return 0, 0, dependencyError(op, err)
	}
	if start, err = wholeTicks(op, r.unitsOf(ratOfPosition(p))); err != nil {
		return 0, 0, err
	}
	if length, err = wholeTicks(op, r.unitsOf(ratOfDuration(d))); err != nil {
		return 0, 0, err
	}
	return start, length, nil
}

// MustBarTicks is like BarTicks but panics with its error.
func (r Resolution) MustBarTicks(g measure.Grid, bar int) (start, length int) {
	return must2(r.BarTicks(g, bar))
}
