package meter

import (
	"github.com/go-muse/musex/duration"
)

// The strong beats follow a rule of the product for any grouping: the first
// unit of the first group is strong, the first units of the other groups are
// relatively strong, and the other units are weak. In these names a beat is a
// unit, as in the Russian tradition. A deeper hierarchy is not modeled.

// unitWeights returns the weight of every unit of a meter with grouping g.
func unitWeights(g []int) []Weight {
	n := 0
	for _, k := range g {
		n += k
	}
	ws := make([]Weight, 0, n)
	for i, k := range g {
		first := WeightRelativelyStrong
		if i == 0 {
			first = WeightStrong
		}
		ws = append(ws, first)
		for range k - 1 {
			ws = append(ws, WeightWeak)
		}
	}
	return ws
}

// StrongBeats returns the offset of the strong beat of m, always 0 (the
// start of the first group), and the starts of the second and later groups,
// the relatively strong beats, in order; relativelyStrong is nil for a single
// group. The result slice is new at every call.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// for a meter without a grouping.
//
// Example: 4/4 gives 0 and [1/2]; 12/8 gives 0 and [3/8 3/4 9/8]; 3/4 gives
// 0 and nil.
func (m Meter) StrongBeats() (strong duration.Position, relativelyStrong []duration.Position, err error) {
	const op = "Meter.StrongBeats"
	r, err := m.recordOf(op)
	if err != nil {
		return duration.Position{}, nil, err
	}
	g, err := r.groupingOf(op)
	if err != nil {
		return duration.Position{}, nil, err
	}
	starts, _ := r.elementStarts(g)
	for _, s := range starts[1:] {
		relativelyStrong = append(relativelyStrong, position(s))
	}
	return duration.Origin(), relativelyStrong, nil
}

// MustStrongBeats is like StrongBeats but panics with the error of
// StrongBeats.
func (m Meter) MustStrongBeats() (strong duration.Position, relativelyStrong []duration.Position) {
	return must2(m.StrongBeats())
}

// UnitWeights returns the weight of every unit of m in order: WeightStrong
// for the first unit of the first group, WeightRelativelyStrong for the first
// units of the other groups, and WeightWeak for the others. Its length is
// UnitCount. The result is a new slice.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// for a meter without a grouping.
//
// Example: 6/8 gives [strong weak weak relatively-strong weak weak].
func (m Meter) UnitWeights() ([]Weight, error) {
	const op = "Meter.UnitWeights"
	r, err := m.recordOf(op)
	if err != nil {
		return nil, err
	}
	g, err := r.groupingOf(op)
	if err != nil {
		return nil, err
	}
	return unitWeights(g), nil
}

// MustUnitWeights is like UnitWeights but panics with the error of
// UnitWeights.
func (m Meter) MustUnitWeights() []Weight {
	return must(m.UnitWeights())
}

// WeightAt returns the weight of the unit of m that starts at t, or
// WeightOffUnit if no unit starts at t. The weight is found by comparing t
// with the starts of units, without subtraction, so it never fails with
// ErrOutOfRange.
//
// Errors, in this order: ErrUnset for the unset meter; ErrInvalidArgument
// together with duration.ErrUnset for an unset t; ErrGroupingRequired for a
// meter without a grouping; ErrOutsideMeter if t is not less than the size of
// m.
//
// Example: in 6/8, t = 3/8 gives WeightRelativelyStrong and t = 7/16 gives
// WeightOffUnit.
func (m Meter) WeightAt(t duration.Position) (Weight, error) {
	const op = "Meter.WeightAt"
	r, err := m.recordOf(op)
	if err != nil {
		return 0, err
	}
	if !t.IsSet() {
		return 0, unsetInputError(op, "position")
	}
	g, err := r.groupingOf(op)
	if err != nil {
		return 0, err
	}
	starts, size := r.layout()
	x := bigRat(t)
	if x.Cmp(size) >= 0 {
		return 0, simpleError(ErrOutsideMeter, op, "position not less than the size")
	}
	ws := unitWeights(g)
	for i, s := range starts {
		if s.Cmp(x) == 0 {
			return ws[i], nil
		}
	}
	return WeightOffUnit, nil
}

// MustWeightAt is like WeightAt but panics with the error of WeightAt.
func (m Meter) MustWeightAt(t duration.Position) Weight {
	return must(m.WeightAt(t))
}
