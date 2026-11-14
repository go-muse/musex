package tempo

import (
	"math/big"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo/term"
)

// Range is a tempo range: a written unit and two numbers, the lower first and
// strictly less than the upper, such as ♪ = 132–144. Both numbers belong to
// the range. A range has both bounds; open ranges exist only as guides of
// terms. A range becomes a Tempo only through an explicit choice: the lower
// bound, the upper bound, or the exact middle.
//
// The zero value Range{} is the unset range. Range is comparable with ==,
// which matches Equal, and can be used as a map key.
type Range struct {
	unit  duration.Written
	lower fraction.Fraction // reduced
	upper fraction.Fraction // reduced
}

// NewRange returns the range from lower to upper with the given unit. In order
// of checking, it returns an error wrapping ErrInvalidArgument if unit, lower,
// or upper is unset, ErrNoTempo if lower or then upper is zero or negative,
// and ErrInvalidArgument if lower is not less than upper.
func NewRange(unit duration.Written, lower, upper fraction.Fraction) (Range, error) {
	const op = "NewRange"
	switch {
	case !unit.IsSet():
		return Range{}, invalidError(op, "unset unit")
	case unsetFraction(lower):
		return Range{}, invalidError(op, "unset lower number")
	case unsetFraction(upper):
		return Range{}, invalidError(op, "unset upper number")
	case lower.Sign() <= 0:
		return Range{}, simpleError(ErrNoTempo, op, "lower number not positive")
	case upper.Sign() <= 0:
		return Range{}, simpleError(ErrNoTempo, op, "upper number not positive")
	case lower.Compare(upper) >= 0:
		return Range{}, invalidError(op, "lower number not less than upper number")
	}
	return Range{unit: unit, lower: lower.Reduce(), upper: upper.Reduce()}, nil
}

// MustNewRange is like NewRange but panics with the returned error.
func MustNewRange(unit duration.Written, lower, upper fraction.Fraction) Range {
	return must(NewRange(unit, lower, upper))
}

// RangeOf returns the range of the guide of the term t with the given unit:
// Allegro of the list of the English Wikipedia with a quarter unit gives
// ♩ = 120–156. The numbers of a guide carry no duration; the unit is chosen by
// the caller. In order of checking, it returns an error wrapping
// ErrInvalidArgument if t or then unit is unset, and ErrNoRange if the term
// has no guide, an open bound, or equal bounds.
func RangeOf(t term.Term, unit duration.Written) (Range, error) {
	const op = "RangeOf"
	switch {
	case !t.IsSet():
		return Range{}, invalidError(op, "unset term")
	case !unit.IsSet():
		return Range{}, invalidError(op, "unset unit")
	}
	g, ok, _ := t.Guide()
	if !ok {
		return Range{}, simpleError(ErrNoRange, op, "term without a guide")
	}
	lo, okLo := g.Lower()
	hi, okHi := g.Upper()
	if !okLo || !okHi {
		return Range{}, simpleError(ErrNoRange, op, "guide with an open bound")
	}
	if lo.Compare(hi) == 0 {
		return Range{}, simpleError(ErrNoRange, op, "guide with equal bounds")
	}
	return Range{unit: unit, lower: lo.Reduce(), upper: hi.Reduce()}, nil
}

// MustRangeOf is like RangeOf but panics with the returned error.
func MustRangeOf(t term.Term, unit duration.Written) Range {
	return must(RangeOf(t, unit))
}

// IsSet reports whether r is set, that is, not the unset Range{}.
func (r Range) IsSet() bool {
	return r.unit.IsSet()
}

// unset returns an error wrapping ErrUnset for the receiver of op.
func (Range) unset(op string) error {
	return unsetError(op, "receiver")
}

// Unit returns the written unit of r. It returns an error wrapping ErrUnset if
// r is unset.
func (r Range) Unit() (duration.Written, error) {
	if !r.IsSet() {
		return duration.Written{}, r.unset("Range.Unit")
	}
	return r.unit, nil
}

// MustUnit is like Unit but panics with the returned error.
func (r Range) MustUnit() duration.Written {
	return must(r.Unit())
}

// UnitDuration returns the duration of the unit of r. It returns an error
// wrapping ErrUnset if r is unset.
func (r Range) UnitDuration() (duration.Duration, error) {
	if !r.IsSet() {
		return duration.Duration{}, r.unset("Range.UnitDuration")
	}
	d, _ := r.unit.Duration()
	return d, nil
}

// MustUnitDuration is like UnitDuration but panics with the returned error.
func (r Range) MustUnitDuration() duration.Duration {
	return must(r.UnitDuration())
}

// Lower returns the lower number of r. It returns an error wrapping ErrUnset
// if r is unset.
func (r Range) Lower() (fraction.Fraction, error) {
	if !r.IsSet() {
		return fraction.Fraction{}, r.unset("Range.Lower")
	}
	return r.lower, nil
}

// MustLower is like Lower but panics with the returned error.
func (r Range) MustLower() fraction.Fraction {
	return must(r.Lower())
}

// Upper returns the upper number of r. It returns an error wrapping ErrUnset
// if r is unset.
func (r Range) Upper() (fraction.Fraction, error) {
	if !r.IsSet() {
		return fraction.Fraction{}, r.unset("Range.Upper")
	}
	return r.upper, nil
}

// MustUpper is like Upper but panics with the returned error.
func (r Range) MustUpper() fraction.Fraction {
	return must(r.Upper())
}

// LowerTempo returns the exact marking of the lower bound: ♪ = 132 for
// ♪ = 132–144. It returns an error wrapping ErrUnset if r is unset.
func (r Range) LowerTempo() (Tempo, error) {
	if !r.IsSet() {
		return Tempo{}, r.unset("Range.LowerTempo")
	}
	return Tempo{m: mark{unit: r.unit, bpm: r.lower}}, nil
}

// MustLowerTempo is like LowerTempo but panics with the returned error.
func (r Range) MustLowerTempo() Tempo {
	return must(r.LowerTempo())
}

// UpperTempo returns the exact marking of the upper bound. It returns an error
// wrapping ErrUnset if r is unset.
func (r Range) UpperTempo() (Tempo, error) {
	if !r.IsSet() {
		return Tempo{}, r.unset("Range.UpperTempo")
	}
	return Tempo{m: mark{unit: r.unit, bpm: r.upper}}, nil
}

// MustUpperTempo is like UpperTempo but panics with the returned error.
func (r Range) MustUpperTempo() Tempo {
	return must(r.UpperTempo())
}

// MidTempo returns the exact marking of the middle, (lower + upper)/2, without
// rounding: ♪ = 138 for ♪ = 132–144 and ♩ = 251/2 for ♩ = 120–131. A rounded
// middle is obtained with Tempo.Quantize. In order of checking, it returns an
// error wrapping ErrUnset if r is unset and ErrOutOfRange together with
// fraction.ErrOutOfRange if the result is outside the range of package
// common/fraction.
func (r Range) MidTempo() (Tempo, error) {
	const op = "Range.MidTempo"
	if !r.IsSet() {
		return Tempo{}, r.unset(op)
	}
	mid := new(big.Rat).Add(ratOf(r.lower), ratOf(r.upper))
	mid.Quo(mid, ratInt(2))
	m, err := markOf(op, r.unit, mid)
	return Tempo{m: m}, err
}

// MustMidTempo is like MidTempo but panics with the returned error.
func (r Range) MustMidTempo() Tempo {
	return must(r.MidTempo())
}

// Equal reports whether r and other are the same record: equal units and
// equal numbers. Two unset ranges are equal. Equal matches ==.
func (r Range) Equal(other Range) bool {
	return r == other
}

// speeds returns the speeds of the bounds of the set range r.
func (r Range) speeds() (lower, upper *big.Rat) {
	u := unitRat(r.unit)
	return mul(u, ratOf(r.lower)), mul(u, ratOf(r.upper))
}

// Compare is a total order for sorting that returns −1, 0, or +1: by the
// speed of the lower bound, then by the speed of the upper bound, and then by
// unit as duration.Written.Compare orders them. The unset range precedes every
// set one. Compare returns 0 exactly when r == other.
func (r Range) Compare(other Range) int {
	if !r.IsSet() || !other.IsSet() {
		return compareSet(r.IsSet(), other.IsSet())
	}
	al, au := r.speeds()
	bl, bu := other.speeds()
	if c := al.Cmp(bl); c != 0 {
		return c
	}
	if c := au.Cmp(bu); c != 0 {
		return c
	}
	return r.unit.Compare(other.unit)
}

// EqualSpeed reports whether the bounds of r and other have equal speeds. It
// returns an error wrapping ErrUnset if r or then other is unset.
func (r Range) EqualSpeed(other Range) (bool, error) {
	const op = "Range.EqualSpeed"
	if !r.IsSet() {
		return false, r.unset(op)
	}
	if !other.IsSet() {
		return false, unsetError(op, "argument")
	}
	al, au := r.speeds()
	bl, bu := other.speeds()
	return al.Cmp(bl) == 0 && au.Cmp(bu) == 0, nil
}

// MustEqualSpeed is like EqualSpeed but panics with the returned error.
func (r Range) MustEqualSpeed(other Range) bool {
	return must(r.EqualSpeed(other))
}

// WithUnit returns the range of the same speeds with unit u: ♪ = 132–144
// becomes ♩ = 66–72. In order of checking, it returns an error wrapping
// ErrUnset if r is unset, ErrInvalidArgument if unit is unset, and
// ErrOutOfRange together with fraction.ErrOutOfRange if a number of the result
// is outside the range of package common/fraction.
func (r Range) WithUnit(unit duration.Written) (Range, error) {
	const op = "Range.WithUnit"
	if !r.IsSet() {
		return Range{}, r.unset(op)
	}
	lo, err := mark{unit: r.unit, bpm: r.lower}.withUnit(op, unit)
	if err != nil {
		return Range{}, err
	}
	hi, err := mark{unit: r.unit, bpm: r.upper}.withUnit(op, unit)
	if err != nil {
		return Range{}, err
	}
	return Range{unit: unit, lower: lo.bpm, upper: hi.bpm}, nil
}

// MustWithUnit is like WithUnit but panics with the returned error.
func (r Range) MustWithUnit(unit duration.Written) Range {
	return must(r.WithUnit(unit))
}

// Contains reports whether lower ≤ beats ≤ upper; the number carries no
// duration. In order of checking, it returns an error wrapping ErrUnset if r
// is unset, ErrInvalidArgument if beats is unset, and ErrNoTempo if beats is
// zero or negative.
func (r Range) Contains(beats fraction.Fraction) (bool, error) {
	const op = "Range.Contains"
	if !r.IsSet() {
		return false, r.unset(op)
	}
	if err := checkBeats(op, beats); err != nil {
		return false, err
	}
	return beats.Compare(r.lower) >= 0 && beats.Compare(r.upper) <= 0, nil
}

// MustContains is like Contains but panics with the returned error.
func (r Range) MustContains(beats fraction.Fraction) bool {
	return must(r.Contains(beats))
}

// ContainsTempo reports whether the speed of t lies between the speeds of the
// bounds of r, bounds included: ♩ = 69 lies in ♪ = 132–144 and ♩ = 65 does
// not. It returns an error wrapping ErrUnset if r or then t is unset.
func (r Range) ContainsTempo(t Tempo) (bool, error) {
	const op = "Range.ContainsTempo"
	if !r.IsSet() {
		return false, r.unset(op)
	}
	if !t.IsSet() {
		return false, unsetError(op, "argument")
	}
	lo, hi := r.speeds()
	s := t.m.speed()
	return s.Cmp(lo) >= 0 && s.Cmp(hi) <= 0, nil
}

// MustContainsTempo is like ContainsTempo but panics with the returned error.
func (r Range) MustContainsTempo(t Tempo) bool {
	return must(r.ContainsTempo(t))
}
