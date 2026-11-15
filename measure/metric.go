package measure

import (
	"math/big"
	"sort"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
)

// unitTable is the units of a meter as exact values: their starts and sizes
// in order, and the size of the meter.
type unitTable struct {
	starts, sizes []*big.Rat
	size          *big.Rat
}

// unitsOf returns the unit table of the set meter m.
func unitsOf(m meter.Meter) unitTable {
	spans, _ := m.Units()
	t := unitTable{
		starts: make([]*big.Rat, len(spans)),
		sizes:  make([]*big.Rat, len(spans)),
		size:   meterSize(m),
	}
	for i, s := range spans {
		t.starts[i] = ratOfPosition(s.Start)
		t.sizes[i] = ratOfDuration(s.Duration)
	}
	return t
}

// count returns the unit count at the place u ≥ 0 in the meter: i + the part
// of unit i passed for u inside unit i, n + 1 at the size of the meter, and
// n + 1 + (u − size)/eₙ beyond it.
func (t unitTable) count(u *big.Rat) *big.Rat {
	n := len(t.starts)
	switch c := u.Cmp(t.size); {
	case c == 0:
		return ratInt(n + 1)
	case c > 0:
		return ratAdd(ratInt(n+1), ratQuo(ratSub(u, t.size), t.sizes[n-1]))
	}
	i := sort.Search(n, func(i int) bool { return t.starts[i].Cmp(u) > 0 }) - 1
	return ratAdd(ratInt(i+1), ratQuo(ratSub(u, t.starts[i]), t.sizes[i]))
}

// place returns the place in the meter of the unit count c ≥ 1, the inverse
// of count.
func (t unitTable) place(c *big.Rat) *big.Rat {
	n := len(t.starts)
	i := floorRat(c)
	if i.Cmp(big.NewInt(int64(n))) <= 0 {
		k := int(i.Int64())
		return ratAdd(t.starts[k-1], ratMul(ratSub(c, ratInt(k)), t.sizes[k-1]))
	}
	return ratAdd(t.size, ratMul(ratSub(c, ratInt(n+1)), t.sizes[n-1]))
}

// ratOne is the value 1. It is only read.
var ratOne = big.NewRat(1, 1)

// MetricPlace returns the effective meter of the bar of a and the place of a
// in the meter: the offset from the start of a full bar of the meter, the
// shift of the bar plus the offset of a. Package meter gives the unit, the
// beat, the group, and the metric weight at this place (meter.Meter.UnitAt,
// BeatAt, GroupAt, WeightAt).
//
// Errors, in this order: ErrUnset for the unset grid, then for the unset
// address; ErrOutsideGrid if g has no bar of a; ErrOutsideBar if the offset
// is not less than the length of the bar; ErrUndefined if the place is not
// less than the size of the meter, which happens only in an extended bar;
// ErrOutOfRange together with fraction.ErrOutOfRange if the place is outside
// the range of common/fraction.
//
// Example: in "3/4;pickup:1/4", 0:0/1 is at 1/2 of 3/4 (unit 3, weak) and
// 1:0/1 at 0/1.
func (g Grid) MetricPlace(a Address) (m meter.Meter, u duration.Position, err error) {
	const op = "Grid.MetricPlace"
	if err := g.checkAddress(op, a); err != nil {
		return meter.Meter{}, duration.Position{}, err
	}
	s, offset, err := g.addressInBar(op, a)
	if err != nil {
		return meter.Meter{}, duration.Position{}, err
	}
	place := ratAdd(s.shift, offset)
	if place.Cmp(s.size) >= 0 {
		return meter.Meter{}, duration.Position{}, errSimple(ErrUndefined, op,
			"the place is beyond the size of the meter in an extended bar")
	}
	p, err := toPosition(op, place)
	if err != nil {
		return meter.Meter{}, duration.Position{}, err
	}
	return s.meter, p, nil
}

// MustMetricPlace is like MetricPlace but panics with the error of
// MetricPlace.
func (g Grid) MustMetricPlace(a Address) (m meter.Meter, u duration.Position) {
	return must2(g.MetricPlace(a))
}

// barOffsetArg checks the receiver, a bar, and an offset in the bar for the
// counts: ErrUnset; ErrInvalidArgument for a negative bar, then together
// with duration.ErrUnset for an unset offset; ErrOutsideGrid for a missing
// bar.
func (g Grid) barOffsetArg(op string, bar int, t duration.Position) (*segment, *big.Rat, error) {
	if !g.IsSet() {
		return nil, nil, errUnsetGrid(op)
	}
	if bar < 0 {
		return nil, nil, errInvalid(op, "negative bar number")
	}
	if !t.IsSet() {
		return nil, nil, errUnsetTime(op, "offset")
	}
	if err := g.checkExists(op, bar); err != nil {
		return nil, nil, err
	}
	return &g.segs[g.segIndex(bar)], ratOfPosition(t), nil
}

// barFractionArg checks the receiver, a bar, and a count or a tstamp:
// ErrUnset; ErrInvalidArgument for a negative bar, then together with
// fraction.ErrInvalid for an unset fraction; ErrOutsideGrid for a missing
// bar.
func (g Grid) barFractionArg(op, operand string, bar int, c fraction.Fraction) (*segment, *big.Rat, error) {
	if !g.IsSet() {
		return nil, nil, errUnsetGrid(op)
	}
	if bar < 0 {
		return nil, nil, errInvalid(op, "negative bar number")
	}
	if !c.IsValid() {
		return nil, nil, errUnsetFraction(op, operand)
	}
	if err := g.checkExists(op, bar); err != nil {
		return nil, nil, err
	}
	return &g.segs[g.segIndex(bar)], ratOfFraction(c), nil
}

// countAt returns the unit count at the offset t of bar bar, counted from
// the start of the bar or, if metric, by the place in the meter.
func (g Grid) countAt(op string, bar int, t duration.Position, metric bool) (fraction.Fraction, error) {
	s, offset, err := g.barOffsetArg(op, bar, t)
	if err != nil {
		return fraction.Fraction{}, err
	}
	if offset.Cmp(s.length) > 0 {
		return fraction.Fraction{}, errSimple(ErrOutsideBar, op, "the offset is greater than the length of the bar")
	}
	u := offset
	if metric {
		u = ratAdd(s.shift, offset)
	}
	return toFraction(op, unitsOf(s.meter).count(u))
}

// countOffset returns the offset in bar bar of the unit count c, counted from
// the start of the bar or, if metric, by the place in the meter.
func (g Grid) countOffset(op string, bar int, c fraction.Fraction, metric bool) (duration.Position, error) {
	s, cr, err := g.barFractionArg(op, "count", bar, c)
	if err != nil {
		return duration.Position{}, err
	}
	if cr.Cmp(ratOne) < 0 {
		return duration.Position{}, errSimple(ErrOutsideBar, op, "the count is less than 1")
	}
	t := unitsOf(s.meter).place(cr)
	if metric {
		t = ratSub(t, s.shift)
	}
	if t.Sign() < 0 || t.Cmp(s.length) > 0 {
		return duration.Position{}, errSimple(ErrOutsideBar, op, "the count gives an offset outside the bar")
	}
	return toPosition(op, t)
}

// CountAt returns the unit count at the offset t of bar bar, counted from the
// start of the bar: the place u = t, and the count is i + (u − sᵢ)/eᵢ inside
// unit i of the meter (start sᵢ, size eᵢ), n + 1 at the size of the meter (the
// right bar line), and n + 1 + (u − size)/eₙ beyond it in an extended bar. An
// offset equal to the length of the bar is the right bar line. The unit count
// is a rule of the product for any bar and any meter; in a full bar of a meter
// with a common denominator it equals the MEI tstamp (TstampAt), and it is not
// a beat of package meter.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar, then together with duration.ErrUnset for an unset t;
// ErrOutsideGrid if g has no such bar; ErrOutsideBar if t is greater than the
// length of the bar; ErrOutOfRange together with fraction.ErrOutOfRange if
// the count is outside the range of common/fraction.
//
// Example: in 4/4, 1/8 gives 3/2 and 1/1 gives 5/1; in the pickup of
// "4/4;pickup:1/4", 0/1 gives 1/1 (MetricCountAt gives 4/1).
func (g Grid) CountAt(bar int, t duration.Position) (fraction.Fraction, error) {
	return g.countAt("Grid.CountAt", bar, t, false)
}

// MustCountAt is like CountAt but panics with the error of CountAt.
func (g Grid) MustCountAt(bar int, t duration.Position) fraction.Fraction {
	return must(g.CountAt(bar, t))
}

// MetricCountAt returns the unit count at the offset t of bar bar by the
// place in the meter: the place u = Shift + t, so the pickup and bars aligned
// to the end count as the end of a full bar. Otherwise it is CountAt; the
// two differ only for the pickup and bars aligned to the end.
//
// Errors are those of CountAt, in the same order.
//
// Example: in "4/4;pickup:1/4", bar 0 gives 4/1 at 0/1, 9/2 at 1/8, and 5/1
// at 1/4.
func (g Grid) MetricCountAt(bar int, t duration.Position) (fraction.Fraction, error) {
	return g.countAt("Grid.MetricCountAt", bar, t, true)
}

// MustMetricCountAt is like MetricCountAt but panics with the error of
// MetricCountAt.
func (g Grid) MustMetricCountAt(bar int, t duration.Position) fraction.Fraction {
	return must(g.MetricCountAt(bar, t))
}

// CountOffset returns the offset t in bar bar of the unit count c counted
// from the start of the bar, the inverse of CountAt.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar, then together with fraction.ErrInvalid for an unset c;
// ErrOutsideGrid if g has no such bar; ErrOutsideBar if c is less than 1 or
// gives an offset outside [0, length of the bar]; ErrOutOfRange together
// with fraction.ErrOutOfRange if the offset is outside the range of
// common/fraction.
//
// Example: in 4/4, 3/2 gives 1/8 and 5/1 gives 1/1.
func (g Grid) CountOffset(bar int, c fraction.Fraction) (duration.Position, error) {
	return g.countOffset("Grid.CountOffset", bar, c, false)
}

// MustCountOffset is like CountOffset but panics with the error of
// CountOffset.
func (g Grid) MustCountOffset(bar int, c fraction.Fraction) duration.Position {
	return must(g.CountOffset(bar, c))
}

// MetricCountOffset returns the offset t in bar bar of the unit count c by
// the place in the meter, t = u − Shift, the inverse of MetricCountAt.
//
// Errors are those of CountOffset, in the same order.
//
// Example: in "4/4;pickup:1/4", bar 0 gives 0/1 for 4/1 and 1/4 for 5/1.
func (g Grid) MetricCountOffset(bar int, c fraction.Fraction) (duration.Position, error) {
	return g.countOffset("Grid.MetricCountOffset", bar, c, true)
}

// MustMetricCountOffset is like MetricCountOffset but panics with the error
// of MetricCountOffset.
func (g Grid) MustMetricCountOffset(bar int, c fraction.Fraction) duration.Position {
	return must(g.MetricCountOffset(bar, c))
}

// tstampBar checks that bar s is full and that its meter has a common
// denominator, and returns the denominator: ErrUndefined, then ErrUndefined
// together with meter.ErrNoCommonDenominator.
func tstampBar(op string, s *segment) (*big.Rat, error) {
	if s.length.Cmp(s.size) != 0 {
		return nil, errSimple(ErrUndefined, op, "MEI tstamp needs a full bar")
	}
	d, err := s.meter.Denominator()
	if err != nil {
		return nil, errNoCommonDenominator(op)
	}
	return ratInt(d), nil
}

// TstampAt returns the MEI tstamp at the offset t of bar bar: 1 + t·d, only
// for a full bar of a meter with a common denominator d, 0 ≤ t ≤ the length
// of the bar. Outside this domain, which is the one established for MEI, the
// tstamp is undefined; CountAt gives the unit count of the product for any
// bar.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar, then together with duration.ErrUnset for an unset t;
// ErrOutsideGrid if g has no such bar; ErrUndefined if the bar is not full,
// then together with meter.ErrNoCommonDenominator if the meter has no common
// denominator; ErrOutsideBar if t is greater than the length of the bar;
// ErrOutOfRange together with fraction.ErrOutOfRange if the tstamp is
// outside the range of common/fraction.
//
// Example: in 4/4, 1/8 gives 3/2; in 6/8, 3/8 gives 4/1.
func (g Grid) TstampAt(bar int, t duration.Position) (fraction.Fraction, error) {
	const op = "Grid.TstampAt"
	s, offset, err := g.barOffsetArg(op, bar, t)
	if err != nil {
		return fraction.Fraction{}, err
	}
	d, err := tstampBar(op, s)
	if err != nil {
		return fraction.Fraction{}, err
	}
	if offset.Cmp(s.length) > 0 {
		return fraction.Fraction{}, errSimple(ErrOutsideBar, op, "the offset is greater than the length of the bar")
	}
	return toFraction(op, ratAdd(ratOne, ratMul(offset, d)))
}

// MustTstampAt is like TstampAt but panics with the error of TstampAt.
func (g Grid) MustTstampAt(bar int, t duration.Position) fraction.Fraction {
	return must(g.TstampAt(bar, t))
}

// TstampOffset returns the offset t = (ts − 1)/d in bar bar of the MEI
// tstamp ts, the inverse of TstampAt, in the same domain.
//
// Errors, in this order: ErrUnset for the unset grid; ErrInvalidArgument for
// a negative bar, then together with fraction.ErrInvalid for an unset ts;
// ErrOutsideGrid if g has no such bar; ErrUndefined if the bar is not full,
// then together with meter.ErrNoCommonDenominator if the meter has no common
// denominator; ErrOutsideBar if ts is less than 1 or gives an offset outside
// [0, length of the bar]; ErrOutOfRange together with fraction.ErrOutOfRange
// if the offset is outside the range of common/fraction.
//
// Example: in 4/4, 3/2 gives 1/8.
func (g Grid) TstampOffset(bar int, ts fraction.Fraction) (duration.Position, error) {
	const op = "Grid.TstampOffset"
	s, tr, err := g.barFractionArg(op, "tstamp", bar, ts)
	if err != nil {
		return duration.Position{}, err
	}
	d, err := tstampBar(op, s)
	if err != nil {
		return duration.Position{}, err
	}
	t := ratQuo(ratSub(tr, ratOne), d)
	if t.Sign() < 0 || t.Cmp(s.length) > 0 {
		return duration.Position{}, errSimple(ErrOutsideBar, op, "the tstamp gives an offset outside the bar")
	}
	return toPosition(op, t)
}

// MustTstampOffset is like TstampOffset but panics with the error of
// TstampOffset.
func (g Grid) MustTstampOffset(bar int, ts fraction.Fraction) duration.Position {
	return must(g.TstampOffset(bar, ts))
}

// meterPlaceArgs checks a meter and a place in its full bar: ErrInvalidArgument
// together with meter.ErrUnset, then together with duration.ErrUnset.
func meterPlaceArgs(op string, m meter.Meter, u duration.Position) (*big.Rat, error) {
	if !m.IsSet() {
		return nil, errUnsetMeter(op, "meter")
	}
	if !u.IsSet() {
		return nil, errUnsetTime(op, "place")
	}
	return ratOfPosition(u), nil
}

// meterFractionArgs checks a meter and a count or a tstamp:
// ErrInvalidArgument together with meter.ErrUnset, then together with
// fraction.ErrInvalid.
func meterFractionArgs(op, operand string, m meter.Meter, c fraction.Fraction) (*big.Rat, error) {
	if !m.IsSet() {
		return nil, errUnsetMeter(op, "meter")
	}
	if !c.IsValid() {
		return nil, errUnsetFraction(op, operand)
	}
	return ratOfFraction(c), nil
}

// MeterCountAt returns the unit count at the place u, 0 ≤ u ≤ size, of a full
// bar of the meter m, as Grid.CountAt counts it: u = size gives n + 1, the
// right bar line.
//
// Errors, in this order: ErrInvalidArgument together with meter.ErrUnset for
// an unset m, then together with duration.ErrUnset for an unset u;
// ErrOutsideBar if u is greater than the size of m; ErrOutOfRange together
// with fraction.ErrOutOfRange if the count is outside the range of
// common/fraction.
//
// Example: in 4/4, 1/8 gives 3/2; in 2/4+3/8, 9/16 gives 7/2.
func MeterCountAt(m meter.Meter, u duration.Position) (fraction.Fraction, error) {
	const op = "MeterCountAt"
	ur, err := meterPlaceArgs(op, m, u)
	if err != nil {
		return fraction.Fraction{}, err
	}
	t := unitsOf(m)
	if ur.Cmp(t.size) > 0 {
		return fraction.Fraction{}, errSimple(ErrOutsideBar, op, "the place is greater than the size of the meter")
	}
	return toFraction(op, t.count(ur))
}

// MustMeterCountAt is like MeterCountAt but panics with the error of
// MeterCountAt.
func MustMeterCountAt(m meter.Meter, u duration.Position) fraction.Fraction {
	return must(MeterCountAt(m, u))
}

// MeterCountOffset returns the place in a full bar of the meter m of the unit
// count c, 1 ≤ c ≤ n + 1, the inverse of MeterCountAt.
//
// Errors, in this order: ErrInvalidArgument together with meter.ErrUnset for
// an unset m, then together with fraction.ErrInvalid for an unset c;
// ErrOutsideBar if c is less than 1 or greater than n + 1; ErrOutOfRange
// together with fraction.ErrOutOfRange if the place is outside the range of
// common/fraction.
//
// Example: in 6/8, 4/1 gives 3/8.
func MeterCountOffset(m meter.Meter, c fraction.Fraction) (duration.Position, error) {
	const op = "MeterCountOffset"
	cr, err := meterFractionArgs(op, "count", m, c)
	if err != nil {
		return duration.Position{}, err
	}
	t := unitsOf(m)
	if cr.Cmp(ratOne) < 0 || cr.Cmp(ratInt(len(t.starts)+1)) > 0 {
		return duration.Position{}, errSimple(ErrOutsideBar, op, "the count is outside the bar")
	}
	return toPosition(op, t.place(cr))
}

// MustMeterCountOffset is like MeterCountOffset but panics with the error of
// MeterCountOffset.
func MustMeterCountOffset(m meter.Meter, c fraction.Fraction) duration.Position {
	return must(MeterCountOffset(m, c))
}

// MeterTstampAt returns the MEI tstamp 1 + u·d at the place u, 0 ≤ u ≤ size,
// of a full bar of the meter m with the common denominator d.
//
// Errors, in this order: ErrInvalidArgument together with meter.ErrUnset for
// an unset m, then together with duration.ErrUnset for an unset u;
// ErrUndefined together with meter.ErrNoCommonDenominator if m has no common
// denominator; ErrOutsideBar if u is greater than the size of m;
// ErrOutOfRange together with fraction.ErrOutOfRange if the tstamp is
// outside the range of common/fraction.
//
// Example: in 6/8, 3/8 gives 4/1.
func MeterTstampAt(m meter.Meter, u duration.Position) (fraction.Fraction, error) {
	const op = "MeterTstampAt"
	ur, err := meterPlaceArgs(op, m, u)
	if err != nil {
		return fraction.Fraction{}, err
	}
	d, err := m.Denominator()
	if err != nil {
		return fraction.Fraction{}, errNoCommonDenominator(op)
	}
	if ur.Cmp(meterSize(m)) > 0 {
		return fraction.Fraction{}, errSimple(ErrOutsideBar, op, "the place is greater than the size of the meter")
	}
	return toFraction(op, ratAdd(ratOne, ratMulInt(ur, d)))
}

// MustMeterTstampAt is like MeterTstampAt but panics with the error of
// MeterTstampAt.
func MustMeterTstampAt(m meter.Meter, u duration.Position) fraction.Fraction {
	return must(MeterTstampAt(m, u))
}

// MeterTstampOffset returns the place (ts − 1)/d in a full bar of the meter m
// of the MEI tstamp ts, the inverse of MeterTstampAt.
//
// Errors, in this order: ErrInvalidArgument together with meter.ErrUnset for
// an unset m, then together with fraction.ErrInvalid for an unset ts;
// ErrUndefined together with meter.ErrNoCommonDenominator if m has no common
// denominator; ErrOutsideBar if ts is less than 1 or greater than n + 1;
// ErrOutOfRange together with fraction.ErrOutOfRange if the place is outside
// the range of common/fraction.
//
// Example: in 6/8, 4/1 gives 3/8.
func MeterTstampOffset(m meter.Meter, ts fraction.Fraction) (duration.Position, error) {
	const op = "MeterTstampOffset"
	tr, err := meterFractionArgs(op, "tstamp", m, ts)
	if err != nil {
		return duration.Position{}, err
	}
	d, err := m.Denominator()
	if err != nil {
		return duration.Position{}, errNoCommonDenominator(op)
	}
	u := ratQuo(ratSub(tr, ratOne), ratInt(d))
	if u.Sign() < 0 || u.Cmp(meterSize(m)) > 0 {
		return duration.Position{}, errSimple(ErrOutsideBar, op, "the tstamp is outside the bar")
	}
	return toPosition(op, u)
}

// MustMeterTstampOffset is like MeterTstampOffset but panics with the error
// of MeterTstampOffset.
func MustMeterTstampOffset(m meter.Meter, ts fraction.Fraction) duration.Position {
	return must(MeterTstampOffset(m, ts))
}

// Complement returns the length of the closing bar that completes the pickup
// of the length pickup to a full bar of the meter m: the size of m minus the
// pickup, and true. For a pickup equal to the size of m there is no
// incomplete closing bar: the result is the unset duration and false, which
// is a result, not an error.
//
// Errors, in this order: ErrInvalidArgument together with meter.ErrUnset for
// an unset m, then together with duration.ErrUnset for an unset pickup;
// ErrLongerThanMeter if the pickup is longer than the size of m;
// ErrOutOfRange together with fraction.ErrOutOfRange if the result is
// outside the range of common/fraction.
//
// Example: 4/4 and 1/4 give 3/4; 3/4 and 1/8 give 5/8; 4/4 and 1/1 give
// false.
func Complement(m meter.Meter, pickup duration.Duration) (closing duration.Duration, ok bool, err error) {
	const op = "Complement"
	if !m.IsSet() {
		return duration.Duration{}, false, errUnsetMeter(op, "meter")
	}
	if !pickup.IsSet() {
		return duration.Duration{}, false, errUnsetTime(op, "pickup")
	}
	rest := ratSub(meterSize(m), ratOfDuration(pickup))
	switch rest.Sign() {
	case -1:
		return duration.Duration{}, false, errSimple(ErrLongerThanMeter, op, "the pickup is longer than the meter")
	case 0:
		return duration.Duration{}, false, nil
	}
	d, err := toDuration(op, rest)
	if err != nil {
		return duration.Duration{}, false, err
	}
	return d, true, nil
}

// MustComplement is like Complement but panics with the error of Complement.
func MustComplement(m meter.Meter, pickup duration.Duration) (closing duration.Duration, ok bool) {
	return must2(Complement(m, pickup))
}
