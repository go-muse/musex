package meter

import (
	"math"
	"slices"
)

// Changing the record keeps the structure: the result of every To method has
// the same units in order and the same grouping as its receiver, so
// EqualStructure of the two is true. If the meter is already written in the
// requested form, it is returned unchanged without further checks.

// numeric returns r written with digits: a symbol becomes the simple
// signature of its structure with the same explicit grouping.
func (r record) numeric() record {
	r.symbol = 0
	r.parts = slices.Clone(r.parts)
	return r
}

// simpleRecord returns the simple signature n/d with the grouping g, or
// without a grouping if g is nil.
func simpleRecord(n, d int, g []int) record {
	r := record{parts: []part{{addends: []int{n}, den: d}}}
	if g != nil {
		r = r.withGrouping(g)
	}
	return r
}

// ToNumeric returns m written with digits: C gives 4/4, ¢ gives 2/2, and
// C[3+1] gives 4/4[3+1]. Every other form is returned unchanged.
//
// Errors: ErrUnset for the unset meter.
func (m Meter) ToNumeric() (Meter, error) {
	const op = "Meter.ToNumeric"
	r, err := m.recordOf(op)
	if err != nil {
		return Meter{}, err
	}
	if r.form() != FormSymbol {
		return m, nil
	}
	return build(op, r.numeric())
}

// MustToNumeric is like ToNumeric but panics with the error of ToNumeric.
func (m Meter) MustToNumeric() Meter {
	return must(m.ToNumeric())
}

// ToSymbol returns m written with a symbol: 4/4 gives C and 2/2 gives ¢, with
// the explicit grouping kept (4/4[3+1] gives C[3+1]).
//
// Errors, in this order: ErrUnset for the unset meter; ErrUndefined for a
// meter that is neither a symbol nor the simple signature 4/4 or 2/2.
func (m Meter) ToSymbol() (Meter, error) {
	const op = "Meter.ToSymbol"
	r, err := m.recordOf(op)
	if err != nil {
		return Meter{}, err
	}
	if r.form() == FormSymbol {
		return m, nil
	}
	if r.form() == FormSimple {
		p := r.parts[0]
		for _, s := range [...]Symbol{SymbolCommon, SymbolCut} {
			if p.addends[0] == s.numerator() && p.den == s.numerator() {
				r.symbol = s
				return build(op, r)
			}
		}
	}
	return Meter{}, simpleError(ErrUndefined, op, "no symbol has this structure")
}

// MustToSymbol is like ToSymbol but panics with the error of ToSymbol.
func (m Meter) MustToSymbol() Meter {
	return must(m.ToSymbol())
}

// ToSimple returns m written as a simple signature: a symbol as ToNumeric
// does; a compound numerator or a sum with one denominator as the simple
// signature of the same number of units with the explicit grouping equal to
// its grouping ((2+3)/8 gives 5/8[2+3]; (3+3)/8 gives 6/8). A sum with one
// denominator and without a grouping gives a simple signature without a
// grouping (5/8+2/8 gives 7/8).
//
// Errors, in this order: ErrUnset for the unset meter;
// ErrNoCommonDenominator for a sum with different denominators; ErrUndefined
// for a sum without a grouping whose number of units has a standard
// grouping, such as 5/8+1/8, which a simple signature always has.
func (m Meter) ToSimple() (Meter, error) {
	const op = "Meter.ToSimple"
	r, err := m.recordOf(op)
	if err != nil {
		return Meter{}, err
	}
	switch r.form() {
	case FormSimple:
		return m, nil
	case FormSymbol:
		return build(op, r.numeric())
	}
	d, ok := r.commonDen()
	if !ok {
		return Meter{}, simpleError(ErrNoCommonDenominator, op, "parts have different denominators")
	}
	n := r.numerator()
	g := r.grouping()
	if g == nil && standardGrouping(n) != nil {
		return Meter{}, simpleError(ErrUndefined, op, "a simple signature of this numerator always has a grouping")
	}
	return build(op, simpleRecord(n, d, g))
}

// MustToSimple is like ToSimple but panics with the error of ToSimple.
func (m Meter) MustToSimple() Meter {
	return must(m.ToSimple())
}

// ToAdditive returns m written as a compound numerator: its groups become the
// addends over the common denominator. 6/8 gives (3+3)/8, C gives (2+2)/4,
// and 2/8+3/8 gives (2+3)/8.
//
// Errors, in this order: ErrUnset for the unset meter;
// ErrNoCommonDenominator for a sum with different denominators;
// ErrGroupingRequired for a meter without a grouping; ErrUndefined for a
// meter of one group, such as 3/4 or ¢.
func (m Meter) ToAdditive() (Meter, error) {
	const op = "Meter.ToAdditive"
	r, err := m.recordOf(op)
	if err != nil {
		return Meter{}, err
	}
	if r.form() == FormAdditive {
		return m, nil
	}
	d, ok := r.commonDen()
	if !ok {
		return Meter{}, simpleError(ErrNoCommonDenominator, op, "parts have different denominators")
	}
	g, err := r.groupingOf(op)
	if err != nil {
		return Meter{}, err
	}
	if len(g) == 1 {
		return Meter{}, simpleError(ErrUndefined, op, "a compound numerator needs at least two groups")
	}
	return build(op, record{parts: []part{{addends: g, den: d}}})
}

// MustToAdditive is like ToAdditive but panics with the error of ToAdditive.
func (m Meter) MustToAdditive() Meter {
	return must(m.ToAdditive())
}

// ToSum returns m written as a sum of signatures: a group of k units of 1/d
// becomes the part k/d. 6/8 gives 3/8+3/8, C gives 2/4+2/4, C[3+1] gives
// 3/4+1/4, and (2+3)/8 gives 2/8+3/8.
//
// Errors, in this order: ErrUnset for the unset meter; ErrGroupingRequired
// for a meter without a grouping; ErrUndefined for a meter of one group, then
// for a group of more than three units, such as in (4+4)/8, because the part
// 4/8 would be grouped 2+2.
func (m Meter) ToSum() (Meter, error) {
	const op = "Meter.ToSum"
	r, err := m.recordOf(op)
	if err != nil {
		return Meter{}, err
	}
	if r.form() == FormSum {
		return m, nil
	}
	g, err := r.groupingOf(op)
	if err != nil {
		return Meter{}, err
	}
	if len(g) == 1 {
		return Meter{}, simpleError(ErrUndefined, op, "a sum needs at least two groups")
	}
	if slices.Max(g) > 3 {
		return Meter{}, simpleError(ErrUndefined, op, "a group of more than three units")
	}
	dens := r.unitDens()
	parts := make([]part, len(g))
	u := 0
	for i, k := range g {
		parts[i] = part{addends: []int{k}, den: dens[u]}
		u += k
	}
	return build(op, record{parts: parts})
}

// MustToSum is like ToSum but panics with the error of ToSum.
func (m Meter) MustToSum() Meter {
	return must(m.ToSum())
}

// withDens returns r written with digits and with the denominator of every
// part replaced by f of it. The addends and the explicit grouping are kept.
func (r record) withDens(f func(int) int) record {
	r = r.numeric()
	for i, p := range r.parts {
		r.parts[i] = part{addends: p.addends, den: f(p.den)}
	}
	return r
}

// Shorter returns m with units half as long: every denominator is doubled.
// The addends and the explicit grouping are kept; a symbol is written with
// digits. 3/4 gives 3/8, C gives 4/8, 2/4+3/8 gives 2/8+3/16.
//
// Errors, in this order: ErrUnset for the unset meter; ErrOutOfRange without
// other sentinels if a new denominator is greater than math.MaxInt;
// ErrOutOfRange together with fraction.ErrOutOfRange if the size or the
// start of a unit of a sum is outside the range of common/fraction.
func (m Meter) Shorter() (Meter, error) {
	const op = "Meter.Shorter"
	r, err := m.recordOf(op)
	if err != nil {
		return Meter{}, err
	}
	for _, p := range r.parts {
		if p.den > math.MaxInt/2 {
			return Meter{}, simpleError(ErrOutOfRange, op, "denominator greater than math.MaxInt")
		}
	}
	return build(op, r.withDens(func(d int) int { return d * 2 }))
}

// MustShorter is like Shorter but panics with the error of Shorter.
func (m Meter) MustShorter() Meter {
	return must(m.Shorter())
}

// Longer returns m with units twice as long: every denominator is halved.
// The addends and the explicit grouping are kept; a symbol is written with
// digits. 3/4 gives 3/2, 5/8[2+3] gives 5/4[2+3].
//
// Errors, in this order: ErrUnset for the unset meter; ErrUndefined if a
// denominator is odd, such as for 3/1 and 4/3; ErrOutOfRange together with
// fraction.ErrOutOfRange if the size or the start of a unit of a sum is
// outside the range of common/fraction.
func (m Meter) Longer() (Meter, error) {
	const op = "Meter.Longer"
	r, err := m.recordOf(op)
	if err != nil {
		return Meter{}, err
	}
	for _, p := range r.parts {
		if p.den%2 != 0 {
			return Meter{}, simpleError(ErrUndefined, op, "odd denominator")
		}
	}
	return build(op, r.withDens(func(d int) int { return d / 2 }))
}

// MustLonger is like Longer but panics with the error of Longer.
func (m Meter) MustLonger() Meter {
	return must(m.Longer())
}

// WithDenominator returns m with the common denominator replaced by
// denominator: the same numerator, addends, parts, and grouping. A symbol is
// written with digits even for its own denominator: C with 4 gives 4/4. 6/8
// with 4 gives 6/4.
//
// Errors, in this order: ErrUnset for the unset meter; ErrInvalidArgument for
// a denominator less than 1; ErrNoCommonDenominator for a sum with different
// denominators.
func (m Meter) WithDenominator(denominator int) (Meter, error) {
	const op = "Meter.WithDenominator"
	r, err := m.recordOf(op)
	if err != nil {
		return Meter{}, err
	}
	if denominator < 1 {
		return Meter{}, invalidError(op, "denominator less than 1")
	}
	if _, ok := r.commonDen(); !ok {
		return Meter{}, simpleError(ErrNoCommonDenominator, op, "parts have different denominators")
	}
	return build(op, r.withDens(func(int) int { return denominator }))
}

// MustWithDenominator is like WithDenominator but panics with the error of
// WithDenominator.
func (m Meter) MustWithDenominator(denominator int) Meter {
	return must(m.WithDenominator(denominator))
}
