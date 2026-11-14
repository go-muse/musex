package meter

import (
	"math/big"
	"math/bits"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
)

const (
	// MaxUnits is the largest number of units of a meter.
	MaxUnits = 1024
	// MaxGroupingVariants is the largest number of grouping variants returned.
	MaxGroupingVariants = 1024
)

// Meter is a time signature: its record and its grouping. A meter describes
// one full measure; it does not know the sequence of measures or the tempo.
//
// The zero value Meter{} is the unset meter. Only IsSet, Equal, Compare, and
// String work with it without an error; every other operation returns
// ErrUnset.
//
// Meter is an immutable value: no operation changes its receiver or its
// arguments, except UnmarshalText through its pointer. Values are comparable
// with ==, which is the same as Equal (equality of the record), and are
// usable as map keys. All operations are safe for concurrent use.
type Meter struct {
	// rec is the canonical text of the record, or "" for the unset meter.
	rec string
	// size is the size of the meter, derived from rec.
	size duration.Duration
}

// New returns the simple signature numerator/denominator, such as 3/4. Any
// denominator of at least 1 is accepted, including 1, 3, and other numbers
// that are not powers of two. The grouping is the standard grouping of the
// numerator, if the numerator has one: New(6, 8) is grouped 3+3, and New(5, 8)
// has no grouping.
//
// Errors, in this order: ErrInvalidArgument for a numerator less than 1, then
// for a denominator less than 1; ErrTooMany for a numerator greater than
// MaxUnits.
//
// Example: New(3, 4) is 3/4; New(1025, 8) fails with ErrTooMany.
func New(numerator, denominator int) (Meter, error) {
	const op = "meter.New"
	switch {
	case numerator < 1:
		return Meter{}, invalidError(op, "numerator less than 1")
	case denominator < 1:
		return Meter{}, invalidError(op, "denominator less than 1")
	case numerator > MaxUnits:
		return Meter{}, simpleError(ErrTooMany, op, "more than MaxUnits units")
	}
	return build(op, record{parts: []part{{addends: []int{numerator}, den: denominator}}})
}

// MustNew is like New but panics with the error of New.
func MustNew(numerator, denominator int) Meter {
	return must(New(numerator, denominator))
}

// NewAdditive returns the compound numerator (a₁+a₂+…)/denominator, such as
// (3+2+3)/8. Its groups are its addends. The slice addends is only read
// during the call.
//
// Errors, in this order: ErrInvalidArgument for fewer than two addends
// (including nil), for the first addend less than 1, then for a denominator
// less than 1; ErrTooMany for a sum of addends greater than MaxUnits.
//
// Example: NewAdditive([]int{3, 2, 3}, 8) is (3+2+3)/8.
func NewAdditive(addends []int, denominator int) (Meter, error) {
	const op = "meter.NewAdditive"
	if len(addends) < 2 {
		return Meter{}, invalidError(op, "fewer than two addends")
	}
	for _, a := range addends {
		if a < 1 {
			return Meter{}, invalidError(op, "addend less than 1")
		}
	}
	if denominator < 1 {
		return Meter{}, invalidError(op, "denominator less than 1")
	}
	if !unitsFit(addends) {
		return Meter{}, simpleError(ErrTooMany, op, "more than MaxUnits units")
	}
	p := part{addends: append([]int(nil), addends...), den: denominator}
	return build(op, record{parts: []part{p}})
}

// MustNewAdditive is like NewAdditive but panics with the error of
// NewAdditive.
func MustNewAdditive(addends []int, denominator int) Meter {
	return must(NewAdditive(addends, denominator))
}

// unitsFit reports whether the sum of counts, each at least 1, is at most
// MaxUnits. It never overflows.
func unitsFit(counts []int) bool {
	n := 0
	for _, c := range counts {
		if c > MaxUnits-n {
			return false
		}
		n += c
	}
	return true
}

// NewSum returns the sum of signatures p₁+p₂+…, such as 2/4+3/8. Each part
// is a set meter of the form FormSimple or FormAdditive without an explicit
// grouping: a result of New, NewAdditive, or an element of Parts. The units
// and groups of the sum are those of its parts in order; the size is the sum
// of their sizes. If a part has no grouping, the sum has none. The slice
// parts is only read during the call.
//
// Errors, in this order: ErrInvalidArgument for fewer than two parts, then
// for the first invalid part (unset, a symbol, a sum, or with an explicit
// grouping); ErrTooMany for more than MaxUnits units in total; ErrOutOfRange
// together with fraction.ErrOutOfRange if the size or the start of a unit is
// outside the range of common/fraction.
//
// Example: NewSum(MustNew(2, 4), MustNew(3, 8)) is 2/4+3/8 with the size 7/8.
func NewSum(parts ...Meter) (Meter, error) {
	const op = "meter.NewSum"
	if len(parts) < 2 {
		return Meter{}, invalidError(op, "fewer than two parts")
	}
	r := record{parts: make([]part, 0, len(parts))}
	for _, m := range parts {
		if !m.IsSet() {
			return Meter{}, invalidError(op, "unset part")
		}
		pr := decode(m.rec)
		if f := pr.form(); f != FormSimple && f != FormAdditive {
			return Meter{}, invalidError(op, "part is not a simple signature or a compound numerator")
		}
		if pr.explicit != nil {
			return Meter{}, invalidError(op, "part with an explicit grouping")
		}
		r.parts = append(r.parts, pr.parts[0])
	}
	counts := make([]int, len(r.parts))
	for i, p := range r.parts {
		counts[i] = p.count()
	}
	if !unitsFit(counts) {
		return Meter{}, simpleError(ErrTooMany, op, "more than MaxUnits units")
	}
	return build(op, r)
}

// MustNewSum is like NewSum but panics with the error of NewSum.
func MustNewSum(parts ...Meter) Meter {
	return must(NewSum(parts...))
}

// NewSymbol returns the meter written with the symbol s: C with the structure
// of 4/4 and the grouping 2+2, or ¢ with the structure of 2/2 and one group.
//
// Errors: ErrInvalidArgument for an invalid symbol.
//
// Example: NewSymbol(SymbolCut) is ¢ with the text "C|".
func NewSymbol(s Symbol) (Meter, error) {
	const op = "meter.NewSymbol"
	if !s.IsValid() {
		return Meter{}, invalidError(op, "invalid symbol")
	}
	n := s.numerator()
	return build(op, record{symbol: s, parts: []part{{addends: []int{n}, den: n}}})
}

// MustNewSymbol is like NewSymbol but panics with the error of NewSymbol.
func MustNewSymbol(s Symbol) Meter {
	return must(NewSymbol(s))
}

// FromDuration returns the simple signature n/denominator whose size is
// size, where n = size·denominator, with the standard grouping or without a
// grouping.
//
// Errors, in this order: ErrInvalidArgument together with duration.ErrUnset
// for an unset size; ErrInvalidArgument for a denominator less than 1, then
// for a size that is not a multiple of 1/denominator; ErrTooMany if n is
// greater than MaxUnits.
//
// Example: FromDuration(duration.MustNew(3, 4), 8) is 6/8.
func FromDuration(size duration.Duration, denominator int) (Meter, error) {
	const op = "meter.FromDuration"
	if !size.IsSet() {
		return Meter{}, unsetInputError(op, "size")
	}
	if denominator < 1 {
		return Meter{}, invalidError(op, "denominator less than 1")
	}
	r, _ := size.MustFraction().BigRat()
	r.Mul(r, new(big.Rat).SetInt64(int64(denominator)))
	if !r.IsInt() {
		return Meter{}, invalidError(op, "size is not a multiple of the unit")
	}
	n := r.Num()
	if n.Cmp(big.NewInt(MaxUnits)) > 0 {
		return Meter{}, simpleError(ErrTooMany, op, "more than MaxUnits units")
	}
	return build(op, record{parts: []part{{addends: []int{int(n.Int64())}, den: denominator}}})
}

// MustFromDuration is like FromDuration but panics with the error of
// FromDuration.
func MustFromDuration(size duration.Duration, denominator int) Meter {
	return must(FromDuration(size, denominator))
}

// Ready-made meters. Each function returns the same value at every call.
var (
	commonTime  = MustNewSymbol(SymbolCommon)
	cutTime     = MustNewSymbol(SymbolCut)
	twoFour     = MustNew(2, 4)
	threeFour   = MustNew(3, 4)
	fourFour    = MustNew(4, 4)
	twoTwo      = MustNew(2, 2)
	threeEight  = MustNew(3, 8)
	sixEight    = MustNew(6, 8)
	nineEight   = MustNew(9, 8)
	twelveEight = MustNew(12, 8)
	sixFour     = MustNew(6, 4)
)

// CommonTime returns C, equal to MustNewSymbol(SymbolCommon).
func CommonTime() Meter { return commonTime }

// CutTime returns ¢, equal to MustNewSymbol(SymbolCut). Its text is "C|".
func CutTime() Meter { return cutTime }

// TwoFour returns 2/4, equal to MustNew(2, 4).
func TwoFour() Meter { return twoFour }

// ThreeFour returns 3/4, equal to MustNew(3, 4).
func ThreeFour() Meter { return threeFour }

// FourFour returns 4/4, equal to MustNew(4, 4). It differs from CommonTime
// in the record and has the same structure.
func FourFour() Meter { return fourFour }

// TwoTwo returns 2/2, equal to MustNew(2, 2). It differs from CutTime in the
// record and has the same structure.
func TwoTwo() Meter { return twoTwo }

// ThreeEight returns 3/8, equal to MustNew(3, 8).
func ThreeEight() Meter { return threeEight }

// SixEight returns 6/8, equal to MustNew(6, 8).
func SixEight() Meter { return sixEight }

// NineEight returns 9/8, equal to MustNew(9, 8).
func NineEight() Meter { return nineEight }

// TwelveEight returns 12/8, equal to MustNew(12, 8).
func TwelveEight() Meter { return twelveEight }

// SixFour returns 6/4, equal to MustNew(6, 4).
func SixFour() Meter { return sixFour }

// IsSet reports whether m is a set meter, that is, not Meter{}.
func (m Meter) IsSet() bool {
	return m.rec != ""
}

// recordOf returns the record of m or an ErrUnset error naming op.
func (m Meter) recordOf(op string) (record, error) {
	if !m.IsSet() {
		return record{}, unsetError(op, "receiver")
	}
	return decode(m.rec), nil
}

// Form returns the form of the record of m: FormSimple for 5/8[2+3],
// FormSymbol for C, FormAdditive for (3+2+3)/8, FormSum for 2/4+3/8.
//
// Errors: ErrUnset for the unset meter.
func (m Meter) Form() (Form, error) {
	r, err := m.recordOf("Meter.Form")
	if err != nil {
		return 0, err
	}
	return r.form(), nil
}

// MustForm is like Form but panics with the error of Form.
func (m Meter) MustForm() Form {
	return must(m.Form())
}

// Symbol returns the symbol of m and true if m is written with a symbol, and
// (0, false) without an error otherwise.
//
// Errors: ErrUnset for the unset meter.
//
// Example: CommonTime().Symbol() is (SymbolCommon, true, nil).
func (m Meter) Symbol() (s Symbol, ok bool, err error) {
	r, err := m.recordOf("Meter.Symbol")
	if err != nil {
		return 0, false, err
	}
	return r.symbol, r.symbol != 0, nil
}

// MustSymbol is like Symbol but panics with the error of Symbol.
func (m Meter) MustSymbol() (s Symbol, ok bool) {
	return must2(m.Symbol())
}

// PartCount returns the number of parts of the record of m: the number of
// signatures of a sum and 1 for every other form.
//
// Errors: ErrUnset for the unset meter.
func (m Meter) PartCount() (int, error) {
	r, err := m.recordOf("Meter.PartCount")
	if err != nil {
		return 0, err
	}
	return len(r.parts), nil
}

// MustPartCount is like PartCount but panics with the error of PartCount.
func (m Meter) MustPartCount() int {
	return must(m.PartCount())
}

// Parts returns the parts of the record of m as meters without an explicit
// grouping: a simple signature itself without its explicit grouping, 4/4 for
// C and 2/2 for ¢, a compound numerator itself, and the parts of a sum in
// order. NewSum(m.MustParts()...) equals m for every sum. The result is a new
// slice.
//
// Errors: ErrUnset for the unset meter.
//
// Example: the parts of 2/4+3/8 are [2/4 3/8].
func (m Meter) Parts() ([]Meter, error) {
	r, err := m.recordOf("Meter.Parts")
	if err != nil {
		return nil, err
	}
	ms := make([]Meter, len(r.parts))
	for i, p := range r.parts {
		ms[i], _ = build("Meter.Parts", record{parts: []part{p}})
	}
	return ms, nil
}

// MustParts is like Parts but panics with the error of Parts.
func (m Meter) MustParts() []Meter {
	return must(m.Parts())
}

// Addends returns the numbers above the line in order: the numerator of a
// simple signature or a symbol, the addends of a compound numerator, and the
// addends of all parts of a sum. The result is a new slice.
//
// Errors: ErrUnset for the unset meter.
//
// Example: the addends of 2/4+3/8 are [2 3]; of (3+2+3)/8, [3 2 3].
func (m Meter) Addends() ([]int, error) {
	r, err := m.recordOf("Meter.Addends")
	if err != nil {
		return nil, err
	}
	var as []int
	for _, p := range r.parts {
		as = append(as, p.addends...)
	}
	return as, nil
}

// MustAddends is like Addends but panics with the error of Addends.
func (m Meter) MustAddends() []int {
	return must(m.Addends())
}

// Numerator returns the sum of the addends of m, which always equals
// UnitCount.
//
// Errors: ErrUnset for the unset meter.
//
// Example: the numerator of 2/4+3/8 is 5; of C, 4.
func (m Meter) Numerator() (int, error) {
	r, err := m.recordOf("Meter.Numerator")
	if err != nil {
		return 0, err
	}
	return r.numerator(), nil
}

// MustNumerator is like Numerator but panics with the error of Numerator.
func (m Meter) MustNumerator() int {
	return must(m.Numerator())
}

// Denominator returns the denominator shared by all parts of m: 4 for C, 2
// for ¢, 8 for 2/8+3/8.
//
// Errors, in this order: ErrUnset for the unset meter;
// ErrNoCommonDenominator for a sum with different denominators, such as
// 2/4+3/8.
func (m Meter) Denominator() (int, error) {
	const op = "Meter.Denominator"
	r, err := m.recordOf(op)
	if err != nil {
		return 0, err
	}
	d, ok := r.commonDen()
	if !ok {
		return 0, simpleError(ErrNoCommonDenominator, op, "parts have different denominators")
	}
	return d, nil
}

// MustDenominator is like Denominator but panics with the error of
// Denominator.
func (m Meter) MustDenominator() int {
	return must(m.Denominator())
}

// HasCommonDenominator reports whether all parts of m have the same
// denominator. It is false only for a sum with different denominators.
//
// Errors: ErrUnset for the unset meter.
func (m Meter) HasCommonDenominator() (bool, error) {
	r, err := m.recordOf("Meter.HasCommonDenominator")
	if err != nil {
		return false, err
	}
	_, ok := r.commonDen()
	return ok, nil
}

// MustHasCommonDenominator is like HasCommonDenominator but panics with the
// error of HasCommonDenominator.
func (m Meter) MustHasCommonDenominator() bool {
	return must(m.HasCommonDenominator())
}

// HasPowerOfTwoDenominators reports whether every denominator of m is a power
// of two, 1 included. False means an irrational meter, such as 4/3 or
// 2/4+2/3; it is a result, not an error.
//
// Errors: ErrUnset for the unset meter.
func (m Meter) HasPowerOfTwoDenominators() (bool, error) {
	r, err := m.recordOf("Meter.HasPowerOfTwoDenominators")
	if err != nil {
		return false, err
	}
	for _, p := range r.parts {
		if !isPowerOfTwo(p.den) {
			return false, nil
		}
	}
	return true, nil
}

// MustHasPowerOfTwoDenominators is like HasPowerOfTwoDenominators but panics
// with the error of HasPowerOfTwoDenominators.
func (m Meter) MustHasPowerOfTwoDenominators() bool {
	return must(m.HasPowerOfTwoDenominators())
}

// isPowerOfTwo reports whether n, at least 1, is a power of two.
func isPowerOfTwo(n int) bool {
	return n&(n-1) == 0
}

// DenominatorExponent returns k such that the common denominator of m is
// 2^k: 3 for 6/8, 0 for 3/1.
//
// Errors, in this order: ErrUnset for the unset meter;
// ErrNoCommonDenominator for a sum with different denominators; ErrUndefined
// for a denominator that is not a power of two, such as 4/3.
func (m Meter) DenominatorExponent() (int, error) {
	const op = "Meter.DenominatorExponent"
	r, err := m.recordOf(op)
	if err != nil {
		return 0, err
	}
	d, ok := r.commonDen()
	if !ok {
		return 0, simpleError(ErrNoCommonDenominator, op, "parts have different denominators")
	}
	if !isPowerOfTwo(d) {
		return 0, simpleError(ErrUndefined, op, "denominator is not a power of two")
	}
	return bits.TrailingZeros(uint(d)), nil
}

// MustDenominatorExponent is like DenominatorExponent but panics with the
// error of DenominatorExponent.
func (m Meter) MustDenominatorExponent() int {
	return must(m.DenominatorExponent())
}

// Duration returns the size of m: the sum over its parts of numerator over
// denominator. 6/8 and 3/4 have the size 3/4; C and (3+2+3)/8 have the size
// 1; 2/4+3/8 has the size 7/8.
//
// Errors: ErrUnset for the unset meter. The range is checked at creation.
func (m Meter) Duration() (duration.Duration, error) {
	if !m.IsSet() {
		return duration.Duration{}, unsetError("Meter.Duration", "receiver")
	}
	return m.size, nil
}

// MustDuration is like Duration but panics with the error of Duration.
func (m Meter) MustDuration() duration.Duration {
	return must(m.Duration())
}

// QuoRem returns how many times v fits wholly into the size of m and the
// remainder, the same as m.MustDuration().QuoRem(v): quo ≥ 0 and
// 0 ≤ rem < v, with rem reduced.
//
// Errors, in this order: ErrUnset for the unset meter; ErrInvalidArgument
// together with duration.ErrUnset for an unset v; ErrOutOfRange for a
// quotient greater than math.MaxInt (with fraction.ErrNotRepresentable), then
// for a remainder outside the range (with fraction.ErrOutOfRange); both also
// wrap duration.ErrOutOfRange.
//
// Example: 6/8 and 1/8 give (6, 0); 6/8 and 1/2 give (1, 1/4).
func (m Meter) QuoRem(v duration.Duration) (quo int, rem fraction.Fraction, err error) {
	const op = "Meter.QuoRem"
	if !m.IsSet() {
		return 0, fraction.Fraction{}, unsetError(op, "receiver")
	}
	if !v.IsSet() {
		return 0, fraction.Fraction{}, unsetInputError(op, "duration")
	}
	quo, rem, err = m.size.QuoRem(v)
	if err != nil {
		return 0, fraction.Fraction{}, causedRangeError(op, err)
	}
	return quo, rem, nil
}

// MustQuoRem is like QuoRem but panics with the error of QuoRem.
func (m Meter) MustQuoRem(v duration.Duration) (quo int, rem fraction.Fraction) {
	return must2(m.QuoRem(v))
}
