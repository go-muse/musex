package duration_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

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

// mustCases lists all 73 Must variants by the summary of architecture 4.1.
func mustCases() []mustCase {
	q := wr(duration.Quarter, 1, tuplet.Triplet())
	h := wr(duration.Half, 0)
	hq := tie(h, wr(duration.Quarter, 0))
	d := dur(5, 16)
	p := pos(7, 24)
	step := frac(1, 8)
	var (
		ud duration.Duration
		uw duration.Written
		ut duration.Tied
		up duration.Position
	)

	return []mustCase{
		// Base and the functions of bases: 10.
		{"MustBaseOf", func() error { return e(duration.BaseOf(dur(3, 8))) }, func() { duration.MustBaseOf(dur(3, 8)) },
			func(t *testing.T) { assert.Equal(t, noErr(duration.BaseOf(dur(1, 4))), duration.MustBaseOf(dur(1, 4))) }},
		{"MustParseBase", func() error { return e(duration.ParseBase("x")) }, func() { duration.MustParseBase("x") },
			func(t *testing.T) {
				assert.Equal(t, noErr(duration.ParseBase("crotchet")), duration.MustParseBase("crotchet"))
			}},
		{"Base.MustDuration", func() error { return e(duration.Base(0).Duration()) }, func() { duration.Base(0).MustDuration() },
			func(t *testing.T) { assert.Equal(t, noErr(duration.Longa.Duration()), duration.Longa.MustDuration()) }},
		{"Base.MustWritten", func() error { return e(duration.Base(15).Written()) }, func() { duration.Base(15).MustWritten() },
			func(t *testing.T) { assert.Equal(t, noErr(duration.Longa.Written()), duration.Longa.MustWritten()) }},
		{"Base.MustLonger", func() error { return e(duration.Maxima.Longer()) }, func() { duration.Maxima.MustLonger() },
			func(t *testing.T) { assert.Equal(t, noErr(duration.Half.Longer()), duration.Half.MustLonger()) }},
		{"Base.MustShorter", func() error { return e(duration.ThousandTwentyFourth.Shorter()) }, func() { duration.ThousandTwentyFourth.MustShorter() },
			func(t *testing.T) { assert.Equal(t, noErr(duration.Half.Shorter()), duration.Half.MustShorter()) }},
		{"Base.MustShift", func() error { return e(duration.Quarter.Shift(9)) }, func() { duration.Quarter.MustShift(9) },
			func(t *testing.T) { assert.Equal(t, noErr(duration.Quarter.Shift(-2)), duration.Quarter.MustShift(-2)) }},
		{"Base.MustName", func() error { return e2(duration.Base(0).Name()) }, func() { duration.Base(0).MustName() },
			func(t *testing.T) { checkNames(t, duration.Quarter.Name, duration.Quarter.MustName) }},
		{"Base.MustBritishName", func() error { return e2(duration.Base(0).BritishName()) }, func() { duration.Base(0).MustBritishName() },
			func(t *testing.T) {
				checkNames(t, duration.ThirtySecond.BritishName, duration.ThirtySecond.MustBritishName)
			}},
		{"Base.MustRussianName", func() error { return e2(duration.Base(0).RussianName()) }, func() { duration.Base(0).MustRussianName() },
			func(t *testing.T) { checkNames(t, duration.Breve.RussianName, duration.Breve.MustRussianName) }},

		// Duration and its functions: 19.
		{"MustNew", func() error { return e(duration.New(0, 1)) }, func() { duration.MustNew(0, 1) },
			func(t *testing.T) { assert.Equal(t, noErr(duration.New(6, 16)), duration.MustNew(6, 16)) }},
		{"MustFromFraction", func() error { return e(duration.FromFraction(fraction.Fraction{})) }, func() { duration.MustFromFraction(fraction.Fraction{}) },
			func(t *testing.T) {
				assert.Equal(t, noErr(duration.FromFraction(frac(3, 8))), duration.MustFromFraction(frac(3, 8)))
			}},
		{"MustSum", func() error { return e(duration.Sum()) }, func() { duration.MustSum() },
			func(t *testing.T) { assert.Equal(t, noErr(duration.Sum(d, d)), duration.MustSum(d, d)) }},
		{"MustMin", func() error { return e(duration.Min(ud)) }, func() { duration.MustMin(ud) },
			func(t *testing.T) { assert.Equal(t, noErr(duration.Min(d, dur(1, 2))), duration.MustMin(d, dur(1, 2))) }},
		{"MustMax", func() error { return e(duration.Max()) }, func() { duration.MustMax() },
			func(t *testing.T) { assert.Equal(t, noErr(duration.Max(d, dur(1, 2))), duration.MustMax(d, dur(1, 2))) }},
		{"MustParse", func() error { return e(duration.Parse("0/1")) }, func() { duration.MustParse("0/1") },
			func(t *testing.T) { assert.Equal(t, noErr(duration.Parse("6/16")), duration.MustParse("6/16")) }},
		{"Duration.MustFraction", func() error { return e(ud.Fraction()) }, func() { ud.MustFraction() },
			func(t *testing.T) { assert.True(t, noErr(d.Fraction()).EqualExact(d.MustFraction())) }},
		{"Duration.MustAdd", func() error { return e(d.Add(ud)) }, func() { d.MustAdd(ud) },
			func(t *testing.T) { assert.Equal(t, noErr(d.Add(d)), d.MustAdd(d)) }},
		{"Duration.MustSub", func() error { return e(d.Sub(d)) }, func() { d.MustSub(d) },
			func(t *testing.T) { assert.Equal(t, noErr(d.Sub(dur(1, 16))), d.MustSub(dur(1, 16))) }},
		{"Duration.MustMul", func() error { return e(d.Mul(fraction.Zero())) }, func() { d.MustMul(fraction.Zero()) },
			func(t *testing.T) { assert.Equal(t, noErr(d.Mul(frac(2, 3))), d.MustMul(frac(2, 3))) }},
		{"Duration.MustMulInt", func() error { return e(d.MulInt(0)) }, func() { d.MustMulInt(0) },
			func(t *testing.T) { assert.Equal(t, noErr(d.MulInt(3)), d.MustMulInt(3)) }},
		{"Duration.MustDiv", func() error { return e(d.Div(frac(-1, 2))) }, func() { d.MustDiv(frac(-1, 2)) },
			func(t *testing.T) { assert.Equal(t, noErr(d.Div(frac(2, 3))), d.MustDiv(frac(2, 3))) }},
		{"Duration.MustDivInt", func() error { return e(d.DivInt(-1)) }, func() { d.MustDivInt(-1) },
			func(t *testing.T) { assert.Equal(t, noErr(d.DivInt(3)), d.MustDivInt(3)) }},
		{"Duration.MustRatio", func() error { return e(d.Ratio(ud)) }, func() { d.MustRatio(ud) },
			func(t *testing.T) { assert.True(t, noErr(d.Ratio(dur(1, 8))).EqualExact(d.MustRatio(dur(1, 8)))) }},
		{"Duration.MustQuoRem", func() error { return e2(ud.QuoRem(d)) }, func() { ud.MustQuoRem(d) },
			func(t *testing.T) {
				quo, rem, err := dur(1, 1).QuoRem(dur(3, 8))
				require.NoError(t, err)

				mustQuo, mustRem := dur(1, 1).MustQuoRem(dur(3, 8))
				assert.Equal(t, quo, mustQuo)
				assert.True(t, rem.EqualExact(mustRem))
			}},
		{"Duration.MustToWritten", func() error { return e(d.ToWritten()) }, func() { d.MustToWritten() },
			func(t *testing.T) { assert.Equal(t, noErr(dur(3, 8).ToWritten()), dur(3, 8).MustToWritten()) }},
		{"Duration.MustNotate", func() error { return e(dur(1, 3).Notate()) }, func() { dur(1, 3).MustNotate() },
			func(t *testing.T) { assert.Equal(t, noErr(d.Notate()), d.MustNotate()) }},
		{"Duration.MustNotateWith", func() error { return e(d.NotateWith(duration.Breve, -1)) }, func() { d.MustNotateWith(duration.Breve, -1) },
			func(t *testing.T) {
				assert.Equal(t, noErr(d.NotateWith(duration.Whole, 1)), d.MustNotateWith(duration.Whole, 1))
			}},
		{"Duration.MustQuantize", func() error { return e(dur(1, 32).Quantize(step, rounding.HalfEven)) },
			func() { dur(1, 32).MustQuantize(step, rounding.HalfEven) },
			func(t *testing.T) {
				assert.Equal(t, noErr(d.Quantize(step, rounding.HalfUp)), d.MustQuantize(step, rounding.HalfUp))
			}},

		// Written: 24.
		{"MustNewWritten", func() error { return e(duration.NewWritten(duration.Quarter, -1)) }, func() { duration.MustNewWritten(duration.Quarter, -1) },
			func(t *testing.T) {
				assert.Equal(t, noErr(duration.NewWritten(duration.Eighth, 1, tp(6, 4))), duration.MustNewWritten(duration.Eighth, 1, tp(6, 4)))
			}},
		{"MustParseWritten", func() error { return e(duration.ParseWritten("3")) }, func() { duration.MustParseWritten("3") },
			func(t *testing.T) {
				assert.Equal(t, noErr(duration.ParseWritten("4.[3:2]")), duration.MustParseWritten("4.[3:2]"))
			}},
		{"Written.MustBase", func() error { return e(uw.Base()) }, func() { uw.MustBase() },
			func(t *testing.T) { assert.Equal(t, noErr(q.Base()), q.MustBase()) }},
		{"Written.MustDots", func() error { return e(uw.Dots()) }, func() { uw.MustDots() },
			func(t *testing.T) { assert.Equal(t, noErr(q.Dots()), q.MustDots()) }},
		{"Written.MustChain", func() error { return e(uw.Chain()) }, func() { uw.MustChain() },
			func(t *testing.T) { assert.Equal(t, noErr(q.Chain()), q.MustChain()) }},
		{"Written.MustCoefficient", func() error { return e(uw.Coefficient()) }, func() { uw.MustCoefficient() },
			func(t *testing.T) { assert.True(t, noErr(q.Coefficient()).EqualExact(q.MustCoefficient())) }},
		{"Written.MustDuration", func() error { return e(uw.Duration()) }, func() { uw.MustDuration() },
			func(t *testing.T) { assert.Equal(t, noErr(q.Duration()), q.MustDuration()) }},
		{"Written.MustHasDots", func() error { return e(uw.HasDots()) }, func() { uw.MustHasDots() },
			func(t *testing.T) { assert.Equal(t, noErr(q.HasDots()), q.MustHasDots()) }},
		{"Written.MustHasTuplets", func() error { return e(uw.HasTuplets()) }, func() { uw.MustHasTuplets() },
			func(t *testing.T) { assert.Equal(t, noErr(q.HasTuplets()), q.MustHasTuplets()) }},
		{"Written.MustIsBasic", func() error { return e(uw.IsBasic()) }, func() { uw.MustIsBasic() },
			func(t *testing.T) { assert.Equal(t, noErr(h.IsBasic()), h.MustIsBasic()) }},
		{"Written.MustEqualDuration", func() error { return e(q.EqualDuration(uw)) }, func() { q.MustEqualDuration(uw) },
			func(t *testing.T) { assert.Equal(t, noErr(q.EqualDuration(h)), q.MustEqualDuration(h)) }},
		{"Written.MustCompareDuration", func() error { return e(uw.CompareDuration(q)) }, func() { uw.MustCompareDuration(q) },
			func(t *testing.T) { assert.Equal(t, noErr(q.CompareDuration(h)), q.MustCompareDuration(h)) }},
		{"Written.MustWithBase", func() error { return e(q.WithBase(0)) }, func() { q.MustWithBase(0) },
			func(t *testing.T) { assert.Equal(t, noErr(q.WithBase(duration.Half)), q.MustWithBase(duration.Half)) }},
		{"Written.MustShiftBase", func() error { return e(q.ShiftBase(10)) }, func() { q.MustShiftBase(10) },
			func(t *testing.T) { assert.Equal(t, noErr(q.ShiftBase(1)), q.MustShiftBase(1)) }},
		{"Written.MustWithDots", func() error { return e(q.WithDots(-1)) }, func() { q.MustWithDots(-1) },
			func(t *testing.T) { assert.Equal(t, noErr(q.WithDots(2)), q.MustWithDots(2)) }},
		{"Written.MustAddDot", func() error { return e(uw.AddDot()) }, func() { uw.MustAddDot() },
			func(t *testing.T) { assert.Equal(t, noErr(q.AddDot()), q.MustAddDot()) }},
		{"Written.MustRemoveDot", func() error { return e(h.RemoveDot()) }, func() { h.MustRemoveDot() },
			func(t *testing.T) { assert.Equal(t, noErr(q.RemoveDot()), q.MustRemoveDot()) }},
		{"Written.MustWithChain", func() error { return e(q.WithChain(tuplet.Tuplet{})) }, func() { q.MustWithChain(tuplet.Tuplet{}) },
			func(t *testing.T) { assert.Equal(t, noErr(q.WithChain(tp(6, 4))), q.MustWithChain(tp(6, 4))) }},
		{"Written.MustAddOuterTuplet", func() error { return e(q.AddOuterTuplet(tuplet.Tuplet{})) }, func() { q.MustAddOuterTuplet(tuplet.Tuplet{}) },
			func(t *testing.T) { assert.Equal(t, noErr(q.AddOuterTuplet(tp(5, 4))), q.MustAddOuterTuplet(tp(5, 4))) }},
		{"Written.MustAddInnerTuplet", func() error { return e(q.AddInnerTuplet(tuplet.Tuplet{})) }, func() { q.MustAddInnerTuplet(tuplet.Tuplet{}) },
			func(t *testing.T) { assert.Equal(t, noErr(q.AddInnerTuplet(tp(5, 4))), q.MustAddInnerTuplet(tp(5, 4))) }},
		{"Written.MustRemoveOuterTuplet", func() error { return e(h.RemoveOuterTuplet()) }, func() { h.MustRemoveOuterTuplet() },
			func(t *testing.T) { assert.Equal(t, noErr(q.RemoveOuterTuplet()), q.MustRemoveOuterTuplet()) }},
		{"Written.MustRemoveInnerTuplet", func() error { return e(h.RemoveInnerTuplet()) }, func() { h.MustRemoveInnerTuplet() },
			func(t *testing.T) { assert.Equal(t, noErr(q.RemoveInnerTuplet()), q.MustRemoveInnerTuplet()) }},
		{"Written.MustRemoveTuplets", func() error { return e(uw.RemoveTuplets()) }, func() { uw.MustRemoveTuplets() },
			func(t *testing.T) { assert.Equal(t, noErr(q.RemoveTuplets()), q.MustRemoveTuplets()) }},
		{"Written.MustSubdivide", func() error { return e2(wr(duration.Quarter, 2).Subdivide()) }, func() { wr(duration.Quarter, 2).MustSubdivide() },
			func(t *testing.T) {
				count, part, err := q.Subdivide()
				require.NoError(t, err)

				mustCount, mustPart := q.MustSubdivide()
				assert.Equal(t, count, mustCount)
				assert.Equal(t, part, mustPart)
			}},

		// Tied: 10.
		{"MustNewTied", func() error { return e(duration.NewTied(h)) }, func() { duration.MustNewTied(h) },
			func(t *testing.T) { assert.Equal(t, noErr(duration.NewTied(h, q)), duration.MustNewTied(h, q)) }},
		{"MustParseTied", func() error { return e(duration.ParseTied("4")) }, func() { duration.MustParseTied("4") },
			func(t *testing.T) { assert.Equal(t, noErr(duration.ParseTied("2~4")), duration.MustParseTied("2~4")) }},
		{"Tied.MustLen", func() error { return e(ut.Len()) }, func() { ut.MustLen() },
			func(t *testing.T) { assert.Equal(t, noErr(hq.Len()), hq.MustLen()) }},
		{"Tied.MustParts", func() error { return e(ut.Parts()) }, func() { ut.MustParts() },
			func(t *testing.T) { assert.Equal(t, noErr(hq.Parts()), hq.MustParts()) }},
		{"Tied.MustDuration", func() error { return e(ut.Duration()) }, func() { ut.MustDuration() },
			func(t *testing.T) { assert.Equal(t, noErr(hq.Duration()), hq.MustDuration()) }},
		{"Tied.MustEqualDuration", func() error { return e(hq.EqualDuration(ut)) }, func() { hq.MustEqualDuration(ut) },
			func(t *testing.T) { assert.Equal(t, noErr(hq.EqualDuration(hq)), hq.MustEqualDuration(hq)) }},
		{"Tied.MustCompareDuration", func() error { return e(ut.CompareDuration(hq)) }, func() { ut.MustCompareDuration(hq) },
			func(t *testing.T) { assert.Equal(t, noErr(hq.CompareDuration(hq)), hq.MustCompareDuration(hq)) }},
		{"Tied.MustAppend", func() error { return e(hq.Append(uw)) }, func() { hq.MustAppend(uw) },
			func(t *testing.T) { assert.Equal(t, noErr(hq.Append(q)), hq.MustAppend(q)) }},
		{"Tied.MustJoin", func() error { return e(hq.Join(ut)) }, func() { hq.MustJoin(ut) },
			func(t *testing.T) { assert.Equal(t, noErr(hq.Join(hq)), hq.MustJoin(hq)) }},
		{"Tied.MustToWritten", func() error { return e(tie(h, wr(duration.Eighth, 0)).ToWritten()) },
			func() { tie(h, wr(duration.Eighth, 0)).MustToWritten() },
			func(t *testing.T) { assert.Equal(t, noErr(hq.ToWritten()), hq.MustToWritten()) }},

		// Position and its functions: 10.
		{"MustNewPosition", func() error { return e(duration.NewPosition(-1, 2)) }, func() { duration.MustNewPosition(-1, 2) },
			func(t *testing.T) { assert.Equal(t, noErr(duration.NewPosition(0, 3)), duration.MustNewPosition(0, 3)) }},
		{"MustPositionFromFraction", func() error { return e(duration.PositionFromFraction(fraction.Fraction{})) },
			func() { duration.MustPositionFromFraction(fraction.Fraction{}) },
			func(t *testing.T) {
				assert.Equal(t, noErr(duration.PositionFromFraction(frac(6, 16))), duration.MustPositionFromFraction(frac(6, 16)))
			}},
		{"MustParsePosition", func() error { return e(duration.ParsePosition("-1/2")) }, func() { duration.MustParsePosition("-1/2") },
			func(t *testing.T) {
				assert.Equal(t, noErr(duration.ParsePosition("0/7")), duration.MustParsePosition("0/7"))
			}},
		{"MustQuantizeSpan", func() error { return e2(duration.QuantizeSpan(pos(1, 10), dur(1, 40), step, rounding.HalfEven)) },
			func() { duration.MustQuantizeSpan(pos(1, 10), dur(1, 40), step, rounding.HalfEven) },
			func(t *testing.T) {
				start, length, err := duration.QuantizeSpan(pos(1, 10), dur(1, 5), step, rounding.HalfEven)
				require.NoError(t, err)

				mustStart, mustLength := duration.MustQuantizeSpan(pos(1, 10), dur(1, 5), step, rounding.HalfEven)
				assert.Equal(t, start, mustStart)
				assert.Equal(t, length, mustLength)
			}},
		{"Position.MustFraction", func() error { return e(up.Fraction()) }, func() { up.MustFraction() },
			func(t *testing.T) { assert.True(t, noErr(p.Fraction()).EqualExact(p.MustFraction())) }},
		{"Position.MustAdd", func() error { return e(p.Add(ud)) }, func() { p.MustAdd(ud) },
			func(t *testing.T) { assert.Equal(t, noErr(p.Add(d)), p.MustAdd(d)) }},
		{"Position.MustSub", func() error { return e(p.Sub(dur(1, 2))) }, func() { p.MustSub(dur(1, 2)) },
			func(t *testing.T) { assert.Equal(t, noErr(p.Sub(dur(1, 24))), p.MustSub(dur(1, 24))) }},
		{"Position.MustOffsetTo", func() error { return e(p.OffsetTo(duration.Origin())) }, func() { p.MustOffsetTo(duration.Origin()) },
			func(t *testing.T) {
				assert.Equal(t, noErr(duration.Origin().OffsetTo(p)), duration.Origin().MustOffsetTo(p))
			}},
		{"Position.MustDurationTo", func() error { return e(p.DurationTo(p)) }, func() { p.MustDurationTo(p) },
			func(t *testing.T) {
				assert.Equal(t, noErr(duration.Origin().DurationTo(p)), duration.Origin().MustDurationTo(p))
			}},
		{"Position.MustQuantize", func() error { return e(p.Quantize(step, rounding.Mode(0))) }, func() { p.MustQuantize(step, rounding.Mode(0)) },
			func(t *testing.T) {
				assert.Equal(t, noErr(p.Quantize(step, rounding.Up)), p.MustQuantize(step, rounding.Up))
			}},
	}
}

// checkNames compares a name lookup with its Must variant.
func checkNames(t *testing.T, get func() (string, bool, error), must func() (string, bool)) {
	t.Helper()

	name, ok, err := get()
	require.NoError(t, err)

	mustName, mustOK := must()
	assert.Equal(t, name, mustName)
	assert.Equal(t, ok, mustOK)
}

// TestMust covers O-17, dev.md "Must", and architecture 4.2: all 73 Must
// variants panic on failure with an error value of the same class and text
// as the base operation, and return the same results on success.
func TestMust(t *testing.T) {
	t.Parallel()

	cases := mustCases()
	require.Len(t, cases, 73)

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

			for _, s := range append(slicesClone(durationSentinels), dependencySentinels...) {
				assert.Equal(t, errors.Is(baseErr, s.err), errors.Is(panicErr, s.err), "errors.Is(err, %s)", s.name)
			}

			value, panicked := capturePanic(func() { c.pass(t) })
			assert.False(t, panicked, "no panic on success: %v", value)
		})
	}
}

// slicesClone copies a list of named sentinels.
func slicesClone(s []namedError) []namedError { return append([]namedError(nil), s...) }
