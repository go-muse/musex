package meter_test

import (
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustCase is one Must variant of architecture 4.1: a failing input of the
// base operation and of the Must variant, and a check of equal results on a
// successful input.
type mustCase struct {
	name     string
	failBase func() error
	failMust func()
	pass     func(t *testing.T)
}

// same1 checks that a Must variant returns the result of the base operation.
func same1[T any](t *testing.T, base func() (T, error), must func() T) {
	t.Helper()

	v, err := base()
	require.NoError(t, err)
	assert.Equal(t, v, must())
}

// same2 checks a Must variant with two results.
func same2[T, U any](t *testing.T, base func() (T, U, error), must func() (T, U)) {
	t.Helper()

	v, w, err := base()
	require.NoError(t, err)

	mv, mw := must()
	assert.Equal(t, v, mv)
	assert.Equal(t, w, mw)
}

// mustCases lists all 69 Must variants by the summary of architecture 4.1:
// 7 functions, 8 names of enumerations, Span.MustEnd, and 53 methods of
// Meter.
func mustCases() []mustCase {
	var u meter.Meter

	m := mt("7/8[3+2+2]")
	sum := mt("2/4+3/8")
	p := pos(3, 8)
	tiny := durRat(new(big.Rat).SetFrac(big.NewInt(1), pow2(200)))

	return []mustCase{
		// Functions: 7.
		{"MustNew", func() error { return e(meter.New(0, 4)) }, func() { meter.MustNew(0, 4) },
			func(t *testing.T) {
				same1(t, func() (meter.Meter, error) { return meter.New(6, 8) }, func() meter.Meter { return meter.MustNew(6, 8) })
			}},
		{"MustNewAdditive", func() error { return e(meter.NewAdditive(nil, 8)) }, func() { meter.MustNewAdditive(nil, 8) },
			func(t *testing.T) {
				same1(t, func() (meter.Meter, error) { return meter.NewAdditive([]int{3, 2}, 8) },
					func() meter.Meter { return meter.MustNewAdditive([]int{3, 2}, 8) })
			}},
		{"MustNewSum", func() error { return e(meter.NewSum()) }, func() { meter.MustNewSum() },
			func(t *testing.T) {
				same1(t, func() (meter.Meter, error) { return meter.NewSum(mt("2/4"), mt("3/8")) },
					func() meter.Meter { return meter.MustNewSum(mt("2/4"), mt("3/8")) })
			}},
		{"MustNewSymbol", func() error { return e(meter.NewSymbol(0)) }, func() { meter.MustNewSymbol(0) },
			func(t *testing.T) {
				same1(t, func() (meter.Meter, error) { return meter.NewSymbol(meter.SymbolCut) },
					func() meter.Meter { return meter.MustNewSymbol(meter.SymbolCut) })
			}},
		{"MustFromDuration", func() error { return e(meter.FromDuration(dur(1, 3), 4)) }, func() { meter.MustFromDuration(dur(1, 3), 4) },
			func(t *testing.T) {
				same1(t, func() (meter.Meter, error) { return meter.FromDuration(dur(3, 4), 8) },
					func() meter.Meter { return meter.MustFromDuration(dur(3, 4), 8) })
			}},
		{"MustParse", func() error { return e(meter.Parse("x")) }, func() { meter.MustParse("x") },
			func(t *testing.T) {
				same1(t, func() (meter.Meter, error) { return meter.Parse("5/8[2+3]") }, func() meter.Meter { return meter.MustParse("5/8[2+3]") })
			}},
		{"MustGroupingVariants", func() error { return e(meter.GroupingVariants(0)) }, func() { meter.MustGroupingVariants(0) },
			func(t *testing.T) {
				same1(t, func() ([][]int, error) { return meter.GroupingVariants(8) }, func() [][]int { return meter.MustGroupingVariants(8) })
			}},

		// Names of enumerations: 8.
		{"Symbol.MustName", func() error { return e2(meter.Symbol(0).Name()) }, func() { meter.Symbol(0).MustName() },
			func(t *testing.T) { same2(t, meter.SymbolCut.Name, meter.SymbolCut.MustName) }},
		{"Symbol.MustRussianName", func() error { return e2(meter.Symbol(9).RussianName()) }, func() { meter.Symbol(9).MustRussianName() },
			func(t *testing.T) { same2(t, meter.SymbolCommon.RussianName, meter.SymbolCommon.MustRussianName) }},
		{"Kind.MustName", func() error { return e2(meter.Kind(0).Name()) }, func() { meter.Kind(0).MustName() },
			func(t *testing.T) { same2(t, meter.KindCompound.Name, meter.KindCompound.MustName) }},
		{"Kind.MustRussianName", func() error { return e2(meter.Kind(7).RussianName()) }, func() { meter.Kind(7).MustRussianName() },
			func(t *testing.T) {
				same2(t, meter.KindUndetermined.RussianName, meter.KindUndetermined.MustRussianName)
			}},
		{"RussianKind.MustName", func() error { return e2(meter.RussianKind(0).Name()) }, func() { meter.RussianKind(0).MustName() },
			func(t *testing.T) { same2(t, meter.RussianSimple.Name, meter.RussianSimple.MustName) }},
		{"RussianKind.MustRussianName", func() error { return e2(meter.RussianKind(0).RussianName()) }, func() { meter.RussianKind(0).MustRussianName() },
			func(t *testing.T) {
				same2(t, meter.RussianCompoundMixed.RussianName, meter.RussianCompoundMixed.MustRussianName)
			}},
		{"Weight.MustName", func() error { return e2(meter.Weight(0).Name()) }, func() { meter.Weight(0).MustName() },
			func(t *testing.T) { same2(t, meter.WeightStrong.Name, meter.WeightStrong.MustName) }},
		{"Weight.MustRussianName", func() error { return e2(meter.Weight(5).RussianName()) }, func() { meter.Weight(5).MustRussianName() },
			func(t *testing.T) { same2(t, meter.WeightWeak.RussianName, meter.WeightWeak.MustRussianName) }},

		// Span: 1.
		{"Span.MustEnd", func() error { return e(meter.Span{}.End()) }, func() { meter.Span{}.MustEnd() },
			func(t *testing.T) { s := sp("3/8@1/4"); same1(t, s.End, s.MustEnd) }},

		// Reading the record: 10.
		{"MustForm", func() error { return e(u.Form()) }, func() { u.MustForm() },
			func(t *testing.T) { same1(t, sum.Form, sum.MustForm) }},
		{"MustSymbol", func() error { return e2(u.Symbol()) }, func() { u.MustSymbol() },
			func(t *testing.T) { c := meter.CommonTime(); same2(t, c.Symbol, c.MustSymbol) }},
		{"MustPartCount", func() error { return e(u.PartCount()) }, func() { u.MustPartCount() },
			func(t *testing.T) { same1(t, sum.PartCount, sum.MustPartCount) }},
		{"MustParts", func() error { return e(u.Parts()) }, func() { u.MustParts() },
			func(t *testing.T) { same1(t, sum.Parts, sum.MustParts) }},
		{"MustAddends", func() error { return e(u.Addends()) }, func() { u.MustAddends() },
			func(t *testing.T) { same1(t, sum.Addends, sum.MustAddends) }},
		{"MustNumerator", func() error { return e(u.Numerator()) }, func() { u.MustNumerator() },
			func(t *testing.T) { same1(t, sum.Numerator, sum.MustNumerator) }},
		{"MustDenominator", func() error { return e(sum.Denominator()) }, func() { sum.MustDenominator() },
			func(t *testing.T) { same1(t, m.Denominator, m.MustDenominator) }},
		{"MustHasCommonDenominator", func() error { return e(u.HasCommonDenominator()) }, func() { u.MustHasCommonDenominator() },
			func(t *testing.T) { same1(t, sum.HasCommonDenominator, sum.MustHasCommonDenominator) }},
		{"MustHasPowerOfTwoDenominators", func() error { return e(u.HasPowerOfTwoDenominators()) }, func() { u.MustHasPowerOfTwoDenominators() },
			func(t *testing.T) { same1(t, sum.HasPowerOfTwoDenominators, sum.MustHasPowerOfTwoDenominators) }},
		{"MustDenominatorExponent", func() error { return e(mt("4/3").DenominatorExponent()) }, func() { mt("4/3").MustDenominatorExponent() },
			func(t *testing.T) { same1(t, m.DenominatorExponent, m.MustDenominatorExponent) }},

		// Size: 2.
		{"MustDuration", func() error { return e(u.Duration()) }, func() { u.MustDuration() },
			func(t *testing.T) { same1(t, sum.Duration, sum.MustDuration) }},
		{"MustQuoRem", func() error { _, _, err := meter.FourFour().QuoRem(tiny); return err }, func() { meter.FourFour().MustQuoRem(tiny) },
			func(t *testing.T) {
				same2(t, func() (int, any, error) { q, r, err := sum.QuoRem(dur(1, 4)); return q, r, err },
					func() (int, any) { q, r := sum.MustQuoRem(dur(1, 4)); return q, r })
			}},

		// Units: 6.
		{"MustUnitCount", func() error { return e(u.UnitCount()) }, func() { u.MustUnitCount() },
			func(t *testing.T) { same1(t, sum.UnitCount, sum.MustUnitCount) }},
		{"MustUnit", func() error { return e(sum.Unit()) }, func() { sum.MustUnit() },
			func(t *testing.T) { same1(t, m.Unit, m.MustUnit) }},
		{"MustUnitWritten", func() error { return e(mt("4/3").UnitWritten()) }, func() { mt("4/3").MustUnitWritten() },
			func(t *testing.T) { same1(t, m.UnitWritten, m.MustUnitWritten) }},
		{"MustUnits", func() error { return e(u.Units()) }, func() { u.MustUnits() },
			func(t *testing.T) { same1(t, sum.Units, sum.MustUnits) }},
		{"MustUnitStart", func() error { return e(m.UnitStart(0)) }, func() { m.MustUnitStart(0) },
			func(t *testing.T) {
				same1(t, func() (duration.Position, error) { return sum.UnitStart(4) }, func() duration.Position { return sum.MustUnitStart(4) })
			}},
		{"MustUnitAt", func() error { return e2(m.UnitAt(pos(1, 1))) }, func() { m.MustUnitAt(pos(1, 1)) },
			func(t *testing.T) {
				same2(t, func() (int, duration.Position, error) { return sum.UnitAt(pos(11, 16)) },
					func() (int, duration.Position) { return sum.MustUnitAt(pos(11, 16)) })
			}},

		// Grouping: 11.
		{"MustHasGrouping", func() error { return e(u.HasGrouping()) }, func() { u.MustHasGrouping() },
			func(t *testing.T) { same1(t, m.HasGrouping, m.MustHasGrouping) }},
		{"MustHasExplicitGrouping", func() error { return e(u.HasExplicitGrouping()) }, func() { u.MustHasExplicitGrouping() },
			func(t *testing.T) { same1(t, m.HasExplicitGrouping, m.MustHasExplicitGrouping) }},
		{"MustGrouping", func() error { return e(mt("5/8").Grouping()) }, func() { mt("5/8").MustGrouping() },
			func(t *testing.T) { same1(t, m.Grouping, m.MustGrouping) }},
		{"MustGroupCount", func() error { return e(mt("5/8").GroupCount()) }, func() { mt("5/8").MustGroupCount() },
			func(t *testing.T) { same1(t, m.GroupCount, m.MustGroupCount) }},
		{"MustGroups", func() error { return e(mt("5/8").Groups()) }, func() { mt("5/8").MustGroups() },
			func(t *testing.T) { same1(t, m.Groups, m.MustGroups) }},
		{"MustGroupsWritten", func() error { return e(mt("7/8[5+2]").GroupsWritten()) }, func() { mt("7/8[5+2]").MustGroupsWritten() },
			func(t *testing.T) { same1(t, m.GroupsWritten, m.MustGroupsWritten) }},
		{"MustGroupStart", func() error { return e(m.GroupStart(4)) }, func() { m.MustGroupStart(4) },
			func(t *testing.T) {
				same1(t, func() (duration.Position, error) { return m.GroupStart(3) }, func() duration.Position { return m.MustGroupStart(3) })
			}},
		{"MustGroupAt", func() error { return e2(m.GroupAt(duration.Position{})) }, func() { m.MustGroupAt(duration.Position{}) },
			func(t *testing.T) {
				same2(t, func() (int, duration.Position, error) { return m.GroupAt(p) }, func() (int, duration.Position) { return m.MustGroupAt(p) })
			}},
		{"MustWithGrouping", func() error { return e(mt("5/8").WithGrouping(2, 2)) }, func() { mt("5/8").MustWithGrouping(2, 2) },
			func(t *testing.T) {
				same1(t, func() (meter.Meter, error) { return mt("5/8").WithGrouping(3, 2) },
					func() meter.Meter { return mt("5/8").MustWithGrouping(3, 2) })
			}},
		{"MustResetGrouping", func() error { return e(u.ResetGrouping()) }, func() { u.MustResetGrouping() },
			func(t *testing.T) { same1(t, m.ResetGrouping, m.MustResetGrouping) }},
		{"Meter.MustGroupingVariants", func() error { return e(mt("28/8").GroupingVariants()) }, func() { mt("28/8").MustGroupingVariants() },
			func(t *testing.T) { same1(t, m.GroupingVariants, m.MustGroupingVariants) }},

		// Beats: 5.
		{"MustBeatCount", func() error { return e(mt("5/8").BeatCount()) }, func() { mt("5/8").MustBeatCount() },
			func(t *testing.T) { same1(t, m.BeatCount, m.MustBeatCount) }},
		{"MustBeats", func() error { return e(mt("5/8").Beats()) }, func() { mt("5/8").MustBeats() },
			func(t *testing.T) { same1(t, m.Beats, m.MustBeats) }},
		{"MustBeatsWritten", func() error { return e(mt("4/3").BeatsWritten()) }, func() { mt("4/3").MustBeatsWritten() },
			func(t *testing.T) { same1(t, m.BeatsWritten, m.MustBeatsWritten) }},
		{"MustBeatStart", func() error { return e(meter.ThreeFour().BeatStart(4)) }, func() { meter.ThreeFour().MustBeatStart(4) },
			func(t *testing.T) {
				same1(t, func() (duration.Position, error) { return m.BeatStart(2) }, func() duration.Position { return m.MustBeatStart(2) })
			}},
		{"MustBeatAt", func() error { return e2(mt("5/8").BeatAt(p)) }, func() { mt("5/8").MustBeatAt(p) },
			func(t *testing.T) {
				same2(t, func() (int, duration.Position, error) { return m.BeatAt(p) }, func() (int, duration.Position) { return m.MustBeatAt(p) })
			}},

		// Kinds: 5.
		{"MustKind", func() error { return e(mt("8/8").Kind()) }, func() { mt("8/8").MustKind() },
			func(t *testing.T) { same1(t, m.Kind, m.MustKind) }},
		{"MustRussianKind", func() error { return e(mt("8/8").RussianKind()) }, func() { mt("8/8").MustRussianKind() },
			func(t *testing.T) { same1(t, m.RussianKind, m.MustRussianKind) }},
		{"MustBeatDivision", func() error { return e2(m.BeatDivision()) }, func() { m.MustBeatDivision() },
			func(t *testing.T) { s := meter.SixEight(); same2(t, s.BeatDivision, s.MustBeatDivision) }},
		{"MustKindName", func() error { return e2(mt("8/8").KindName()) }, func() { mt("8/8").MustKindName() },
			func(t *testing.T) { s := meter.SixEight(); same2(t, s.KindName, s.MustKindName) }},
		{"MustKindRussianName", func() error { return e2(u.KindRussianName()) }, func() { u.MustKindRussianName() },
			func(t *testing.T) { s := meter.SixEight(); same2(t, s.KindRussianName, s.MustKindRussianName) }},

		// Accents: 3.
		{"MustStrongBeats", func() error { return e2(mt("5/8").StrongBeats()) }, func() { mt("5/8").MustStrongBeats() },
			func(t *testing.T) { same2(t, m.StrongBeats, m.MustStrongBeats) }},
		{"MustUnitWeights", func() error { return e(mt("5/8").UnitWeights()) }, func() { mt("5/8").MustUnitWeights() },
			func(t *testing.T) { same1(t, m.UnitWeights, m.MustUnitWeights) }},
		{"MustWeightAt", func() error { return e(m.WeightAt(pos(7, 8))) }, func() { m.MustWeightAt(pos(7, 8)) },
			func(t *testing.T) {
				same1(t, func() (meter.Weight, error) { return m.WeightAt(p) }, func() meter.Weight { return m.MustWeightAt(p) })
			}},

		// Changing the record or the unit: 8.
		{"MustToNumeric", func() error { return e(u.ToNumeric()) }, func() { u.MustToNumeric() },
			func(t *testing.T) { c := meter.CommonTime(); same1(t, c.ToNumeric, c.MustToNumeric) }},
		{"MustToSymbol", func() error { return e(meter.ThreeFour().ToSymbol()) }, func() { meter.ThreeFour().MustToSymbol() },
			func(t *testing.T) { f := meter.FourFour(); same1(t, f.ToSymbol, f.MustToSymbol) }},
		{"MustToSimple", func() error { return e(mt("5/8+1/8").ToSimple()) }, func() { mt("5/8+1/8").MustToSimple() },
			func(t *testing.T) { a := mt("(2+3)/8"); same1(t, a.ToSimple, a.MustToSimple) }},
		{"MustToAdditive", func() error { return e(mt("5/8").ToAdditive()) }, func() { mt("5/8").MustToAdditive() },
			func(t *testing.T) { same1(t, m.ToAdditive, m.MustToAdditive) }},
		{"MustToSum", func() error { return e(meter.ThreeFour().ToSum()) }, func() { meter.ThreeFour().MustToSum() },
			func(t *testing.T) { same1(t, m.ToSum, m.MustToSum) }},
		{"MustShorter", func() error { return e(meter.MustNew(1, math.MaxInt).Shorter()) }, func() { meter.MustNew(1, math.MaxInt).MustShorter() },
			func(t *testing.T) { same1(t, sum.Shorter, sum.MustShorter) }},
		{"MustLonger", func() error { return e(mt("3/1").Longer()) }, func() { mt("3/1").MustLonger() },
			func(t *testing.T) { same1(t, sum.Longer, sum.MustLonger) }},
		{"MustWithDenominator", func() error { return e(sum.WithDenominator(16)) }, func() { sum.MustWithDenominator(16) },
			func(t *testing.T) {
				same1(t, func() (meter.Meter, error) { return m.WithDenominator(4) }, func() meter.Meter { return m.MustWithDenominator(4) })
			}},

		// Equality and order: 3.
		{"MustEqualStructure", func() error { return e(m.EqualStructure(u)) }, func() { m.MustEqualStructure(u) },
			func(t *testing.T) {
				same1(t, func() (bool, error) { return m.EqualStructure(sum) }, func() bool { return m.MustEqualStructure(sum) })
			}},
		{"MustEqualDuration", func() error { return e(u.EqualDuration(m)) }, func() { u.MustEqualDuration(m) },
			func(t *testing.T) {
				same1(t, func() (bool, error) { return m.EqualDuration(sum) }, func() bool { return m.MustEqualDuration(sum) })
			}},
		{"MustCompareDuration", func() error { return e(m.CompareDuration(u)) }, func() { m.MustCompareDuration(u) },
			func(t *testing.T) {
				same1(t, func() (int, error) { return m.CompareDuration(sum) }, func() int { return m.MustCompareDuration(sum) })
			}},
	}
}

// TestMust covers О-21, dev.md "Must", and architecture 4.2: all 69 Must
// variants panic on failure with an error value of the same class and text
// as the base operation, and return the same results on success.
func TestMust(t *testing.T) {
	t.Parallel()

	cases := mustCases()
	require.Len(t, cases, 69)

	names := map[string]bool{}
	for _, c := range cases {
		assert.False(t, names[c.name], "duplicate %s", c.name)
		names[c.name] = true
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			baseErr := c.failBase()
			require.Error(t, baseErr, "the input fails")

			panicErr := panicError(t, c.failMust)
			require.Error(t, panicErr)

			assert.Equal(t, baseErr.Error(), panicErr.Error())

			for _, s := range allSentinels() {
				assert.Equal(t, errors.Is(baseErr, s.err), errors.Is(panicErr, s.err), "errors.Is(err, %s)", s.name)
			}

			value, panicked := capturePanic(func() { c.pass(t) })
			assert.False(t, panicked, "no panic on success: %v", value)
		})
	}
}
