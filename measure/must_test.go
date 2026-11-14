package measure_test

import (
	"errors"
	"iter"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-9, Must variants (architecture 4.2, "Must"; О-23; dev.md,
// "Must"): each of the 74 variants returns the results of its base
// operation on success and panics with the error of the base operation on
// failure.

// mustCase is one Must variant: a failing call of the base operation and of
// the Must variant, and a check of equal results on a successful input.
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

// sameGrid checks a Must variant that returns a grid: equal by record.
func sameGrid(t *testing.T, base func() (measure.Grid, error), must func() measure.Grid) {
	t.Helper()

	v, err := base()
	require.NoError(t, err)

	m := must()
	assert.True(t, v.IsSet())
	assert.True(t, v.Equal(m), "%s and %s", v, m)
}

// sameSeq checks a Must variant that returns a sequence of bars.
func sameSeq(t *testing.T, base func() (iter.Seq[measure.Bar], error), must func() iter.Seq[measure.Bar]) {
	t.Helper()

	v, err := base()
	require.NoError(t, err)
	assert.Equal(t, collect(v), collect(must()))
}

// mustCases lists all 74 Must variants of architecture 4.1: 9 functions,
// 2 methods of Address, Part.MustEnd, and 62 methods of Grid.
func mustCases(t *testing.T) []mustCase {
	t.Helper()

	var (
		u  measure.Grid
		ua measure.Address
	)

	g := gridOf(t, "4/4;pickup:1/4;3:3/4;4:length:1/2;bars:8")
	a := ad(2, "1/8")
	p := pos("9/8")
	three := gridOf(t, "3/4")
	part := measure.Part{Address: ad(1, "0"), Start: pos("1/4"), Duration: dur("1/8")}

	return []mustCase{
		// Functions: 9.
		{"MustNew", func() error { return e(measure.New(meter.Meter{})) }, func() { measure.MustNew(meter.Meter{}) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return measure.New(mt("6/8")) }, func() measure.Grid { return measure.MustNew(mt("6/8")) })
			}},
		{"MustParse", func() error { return e(measure.Parse("x")) }, func() { measure.MustParse("x") },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return measure.Parse("4/4;3:3/4") }, func() measure.Grid { return measure.MustParse("4/4;3:3/4") })
			}},
		{"MustNewAddress", func() error { return e(measure.NewAddress(-1, pos("0"))) }, func() { measure.MustNewAddress(-1, pos("0")) },
			func(t *testing.T) {
				same1(t, func() (measure.Address, error) { return measure.NewAddress(2, pos("1/8")) },
					func() measure.Address { return measure.MustNewAddress(2, pos("1/8")) })
			}},
		{"MustParseAddress", func() error { return e(measure.ParseAddress("x")) }, func() { measure.MustParseAddress("x") },
			func(t *testing.T) {
				same1(t, func() (measure.Address, error) { return measure.ParseAddress("2:1/8") },
					func() measure.Address { return measure.MustParseAddress("2:1/8") })
			}},
		{"MustComplement", func() error { return e2(measure.Complement(mt("3/4"), dur("1"))) }, func() { measure.MustComplement(mt("3/4"), dur("1")) },
			func(t *testing.T) {
				same2(t, func() (duration.Duration, bool, error) { return measure.Complement(mt("4/4"), dur("1/4")) },
					func() (duration.Duration, bool) { return measure.MustComplement(mt("4/4"), dur("1/4")) })
			}},
		{"MustMeterCountAt", func() error { return e(measure.MeterCountAt(mt("4/4"), pos("9/8"))) }, func() { measure.MustMeterCountAt(mt("4/4"), pos("9/8")) },
			func(t *testing.T) {
				same1(t, func() (fraction.Fraction, error) { return measure.MeterCountAt(mt("4/4"), pos("1/8")) },
					func() fraction.Fraction { return measure.MustMeterCountAt(mt("4/4"), pos("1/8")) })
			}},
		{"MustMeterCountOffset", func() error { return e(measure.MeterCountOffset(mt("4/4"), frac("6"))) }, func() { measure.MustMeterCountOffset(mt("4/4"), frac("6")) },
			func(t *testing.T) {
				same1(t, func() (duration.Position, error) { return measure.MeterCountOffset(mt("6/8"), frac("4")) },
					func() duration.Position { return measure.MustMeterCountOffset(mt("6/8"), frac("4")) })
			}},
		{"MustMeterTstampAt", func() error { return e(measure.MeterTstampAt(mt("2/4+3/8"), pos("0"))) }, func() { measure.MustMeterTstampAt(mt("2/4+3/8"), pos("0")) },
			func(t *testing.T) {
				same1(t, func() (fraction.Fraction, error) { return measure.MeterTstampAt(mt("6/8"), pos("3/8")) },
					func() fraction.Fraction { return measure.MustMeterTstampAt(mt("6/8"), pos("3/8")) })
			}},
		{"MustMeterTstampOffset", func() error { return e(measure.MeterTstampOffset(mt("4/4"), frac("0"))) }, func() { measure.MustMeterTstampOffset(mt("4/4"), frac("0")) },
			func(t *testing.T) {
				same1(t, func() (duration.Position, error) { return measure.MeterTstampOffset(mt("6/8"), frac("4")) },
					func() duration.Position { return measure.MustMeterTstampOffset(mt("6/8"), frac("4")) })
			}},

		// Address: 2.
		{"Address.MustBar", func() error { return e(ua.Bar()) }, func() { ua.MustBar() },
			func(t *testing.T) { same1(t, a.Bar, a.MustBar) }},
		{"Address.MustOffset", func() error { return e(ua.Offset()) }, func() { ua.MustOffset() },
			func(t *testing.T) { same1(t, a.Offset, a.MustOffset) }},

		// Part: 1.
		{"Part.MustEnd", func() error { return e(measure.Part{}.End()) }, func() { measure.Part{}.MustEnd() },
			func(t *testing.T) { same1(t, part.End, part.MustEnd) }},

		// Reading the record: 13.
		{"MustPickup", func() error { return e2(u.Pickup()) }, func() { u.MustPickup() },
			func(t *testing.T) { same2(t, g.Pickup, g.MustPickup) }},
		{"MustHasPickup", func() error { return e(u.HasPickup()) }, func() { u.MustHasPickup() },
			func(t *testing.T) { same1(t, g.HasPickup, g.MustHasPickup) }},
		{"MustMarks", func() error { return e(u.Marks()) }, func() { u.MustMarks() },
			func(t *testing.T) { same1(t, g.Marks, g.MustMarks) }},
		{"MustHasMark", func() error { return e(g.HasMark(-1)) }, func() { g.MustHasMark(-1) },
			func(t *testing.T) {
				same1(t, func() (bool, error) { return g.HasMark(3) }, func() bool { return g.MustHasMark(3) })
			}},
		{"MustChanges", func() error { return e(u.Changes()) }, func() { u.MustChanges() },
			func(t *testing.T) { same1(t, g.Changes, g.MustChanges) }},
		{"MustChangesAt", func() error { return e(g.ChangesAt(9)) }, func() { g.MustChangesAt(9) },
			func(t *testing.T) {
				same1(t, func() (bool, error) { return g.ChangesAt(3) }, func() bool { return g.MustChangesAt(3) })
			}},
		{"MustLengths", func() error { return e(u.Lengths()) }, func() { u.MustLengths() },
			func(t *testing.T) { same1(t, g.Lengths, g.MustLengths) }},
		{"MustBarCount", func() error { return e2(u.BarCount()) }, func() { u.MustBarCount() },
			func(t *testing.T) { same2(t, g.BarCount, g.MustBarCount) }},
		{"MustIsFinite", func() error { return e(u.IsFinite()) }, func() { u.MustIsFinite() },
			func(t *testing.T) { same1(t, g.IsFinite, g.MustIsFinite) }},
		{"MustTotalBars", func() error { return e2(u.TotalBars()) }, func() { u.MustTotalBars() },
			func(t *testing.T) { same2(t, g.TotalBars, g.MustTotalBars) }},
		{"MustEnd", func() error { return e2(u.End()) }, func() { u.MustEnd() },
			func(t *testing.T) { same2(t, g.End, g.MustEnd) }},
		{"MustFirstBar", func() error { return e(u.FirstBar()) }, func() { u.MustFirstBar() },
			func(t *testing.T) { same1(t, g.FirstBar, g.MustFirstBar) }},
		{"MustHasBar", func() error { return e(g.HasBar(-1)) }, func() { g.MustHasBar(-1) },
			func(t *testing.T) {
				same1(t, func() (bool, error) { return g.HasBar(9) }, func() bool { return g.MustHasBar(9) })
			}},

		// Bars: 8.
		{"MustBar", func() error { return e(g.Bar(9)) }, func() { g.MustBar(9) },
			func(t *testing.T) {
				same1(t, func() (measure.Bar, error) { return g.Bar(3) }, func() measure.Bar { return g.MustBar(3) })
			}},
		{"MustBarMeter", func() error { return e(g.BarMeter(-1)) }, func() { g.MustBarMeter(-1) },
			func(t *testing.T) {
				same1(t, func() (meter.Meter, error) { return g.BarMeter(3) }, func() meter.Meter { return g.MustBarMeter(3) })
			}},
		{"MustBarLength", func() error { return e(g.BarLength(9)) }, func() { g.MustBarLength(9) },
			func(t *testing.T) {
				same1(t, func() (duration.Duration, error) { return g.BarLength(4) }, func() duration.Duration { return g.MustBarLength(4) })
			}},
		{"MustBarStart", func() error { return e(g.BarStart(9)) }, func() { g.MustBarStart(9) },
			func(t *testing.T) {
				same1(t, func() (duration.Position, error) { return g.BarStart(5) }, func() duration.Position { return g.MustBarStart(5) })
			}},
		{"MustBarEnd", func() error { return e(g.BarEnd(9)) }, func() { g.MustBarEnd(9) },
			func(t *testing.T) {
				same1(t, func() (duration.Position, error) { return g.BarEnd(5) }, func() duration.Position { return g.MustBarEnd(5) })
			}},
		{"MustBars", func() error { return e(g.Bars(3, 2)) }, func() { g.MustBars(3, 2) },
			func(t *testing.T) {
				sameSeq(t, func() (iter.Seq[measure.Bar], error) { return g.Bars(0, 5) }, func() iter.Seq[measure.Bar] { return g.MustBars(0, 5) })
			}},
		{"MustOrdinal", func() error { return e(g.Ordinal(9)) }, func() { g.MustOrdinal(9) },
			func(t *testing.T) {
				same1(t, func() (int, error) { return g.Ordinal(3) }, func() int { return g.MustOrdinal(3) })
			}},
		{"MustBarAtOrdinal", func() error { return e(g.BarAtOrdinal(0)) }, func() { g.MustBarAtOrdinal(0) },
			func(t *testing.T) {
				same1(t, func() (int, error) { return g.BarAtOrdinal(4) }, func() int { return g.MustBarAtOrdinal(4) })
			}},

		// Positions and addresses: 7.
		{"MustLocate", func() error { return e(g.Locate(pos("9"))) }, func() { g.MustLocate(pos("9")) },
			func(t *testing.T) {
				same1(t, func() (measure.Address, error) { return g.Locate(p) }, func() measure.Address { return g.MustLocate(p) })
			}},
		{"MustBarAt", func() error { return e(g.BarAt(pos("9"))) }, func() { g.MustBarAt(pos("9")) },
			func(t *testing.T) {
				same1(t, func() (measure.Bar, error) { return g.BarAt(p) }, func() measure.Bar { return g.MustBarAt(p) })
			}},
		{"MustRemaining", func() error { return e(g.Remaining(duration.Position{})) }, func() { g.MustRemaining(duration.Position{}) },
			func(t *testing.T) {
				same1(t, func() (duration.Duration, error) { return g.Remaining(p) }, func() duration.Duration { return g.MustRemaining(p) })
			}},
		{"MustOnBarline", func() error { return e(g.OnBarline(pos("9"))) }, func() { g.MustOnBarline(pos("9")) },
			func(t *testing.T) {
				same1(t, func() (bool, error) { return g.OnBarline(pos("5/4")) }, func() bool { return g.MustOnBarline(pos("5/4")) })
			}},
		{"MustPositionOf", func() error { return e(g.PositionOf(ad(1, "1"))) }, func() { g.MustPositionOf(ad(1, "1")) },
			func(t *testing.T) {
				same1(t, func() (duration.Position, error) { return g.PositionOf(a) }, func() duration.Position { return g.MustPositionOf(a) })
			}},
		{"MustContains", func() error { return e(g.Contains(ua)) }, func() { g.MustContains(ua) },
			func(t *testing.T) {
				same1(t, func() (bool, error) { return g.Contains(a) }, func() bool { return g.MustContains(a) })
			}},
		{"MustNormalize", func() error { return e(g.Normalize(ad(9, "0"))) }, func() { g.MustNormalize(ad(9, "0")) },
			func(t *testing.T) {
				same1(t, func() (measure.Address, error) { return g.Normalize(ad(1, "2")) }, func() measure.Address { return g.MustNormalize(ad(1, "2")) })
			}},

		// Ranges of bars: 5.
		{"MustRangeLength", func() error { return e(g.RangeLength(3, 9)) }, func() { g.MustRangeLength(3, 9) },
			func(t *testing.T) {
				same1(t, func() (duration.Duration, error) { return g.RangeLength(1, 4) }, func() duration.Duration { return g.MustRangeLength(1, 4) })
			}},
		{"MustChangesWithin", func() error { return e(g.ChangesWithin(-1, 2)) }, func() { g.MustChangesWithin(-1, 2) },
			func(t *testing.T) {
				same1(t, func() (bool, error) { return g.ChangesWithin(1, 4) }, func() bool { return g.MustChangesWithin(1, 4) })
			}},
		{"MustAllFull", func() error { return e(g.AllFull(4, 3)) }, func() { g.MustAllFull(4, 3) },
			func(t *testing.T) {
				same1(t, func() (bool, error) { return g.AllFull(1, 4) }, func() bool { return g.MustAllFull(1, 4) })
			}},
		{"MustMeterRuns", func() error { return e(g.MeterRuns(1, 9)) }, func() { g.MustMeterRuns(1, 9) },
			func(t *testing.T) {
				same1(t, func() ([]measure.MeterRun, error) { return g.MeterRuns(0, 8) }, func() []measure.MeterRun { return g.MustMeterRuns(0, 8) })
			}},
		{"MustBarsSpanned", func() error { return e2(g.BarsSpanned(pos("6"), dur("1"))) }, func() { g.MustBarsSpanned(pos("6"), dur("1")) },
			func(t *testing.T) {
				same2(t, func() (int, int, error) { return g.BarsSpanned(pos("1"), dur("3")) },
					func() (int, int) { return g.MustBarsSpanned(pos("1"), dur("3")) })
			}},

		// Parts of a bar: 2.
		{"MustElapsed", func() error { return e(g.Elapsed(ua)) }, func() { g.MustElapsed(ua) },
			func(t *testing.T) {
				same1(t, func() (fraction.Fraction, error) { return g.Elapsed(a) }, func() fraction.Fraction { return g.MustElapsed(a) })
			}},
		{"MustPartOfBar", func() error { return e(g.PartOfBar(3, duration.Duration{})) }, func() { g.MustPartOfBar(3, duration.Duration{}) },
			func(t *testing.T) {
				same1(t, func() (fraction.Fraction, error) { return g.PartOfBar(3, dur("1/2")) }, func() fraction.Fraction { return g.MustPartOfBar(3, dur("1/2")) })
			}},

		// Place in the meter, counts, and MEI tstamp: 7.
		{"MustMetricPlace", func() error { return e2(g.MetricPlace(ad(1, "1"))) }, func() { g.MustMetricPlace(ad(1, "1")) },
			func(t *testing.T) {
				same2(t, func() (meter.Meter, duration.Position, error) { return g.MetricPlace(ad(0, "0")) },
					func() (meter.Meter, duration.Position) { return g.MustMetricPlace(ad(0, "0")) })
			}},
		{"MustCountAt", func() error { return e(g.CountAt(1, pos("9/8"))) }, func() { g.MustCountAt(1, pos("9/8")) },
			func(t *testing.T) {
				same1(t, func() (fraction.Fraction, error) { return g.CountAt(0, pos("1/8")) }, func() fraction.Fraction { return g.MustCountAt(0, pos("1/8")) })
			}},
		{"MustMetricCountAt", func() error { return e(g.MetricCountAt(0, pos("1/2"))) }, func() { g.MustMetricCountAt(0, pos("1/2")) },
			func(t *testing.T) {
				same1(t, func() (fraction.Fraction, error) { return g.MetricCountAt(0, pos("1/8")) },
					func() fraction.Fraction { return g.MustMetricCountAt(0, pos("1/8")) })
			}},
		{"MustCountOffset", func() error { return e(g.CountOffset(1, frac("0"))) }, func() { g.MustCountOffset(1, frac("0")) },
			func(t *testing.T) {
				same1(t, func() (duration.Position, error) { return g.CountOffset(1, frac("3/2")) },
					func() duration.Position { return g.MustCountOffset(1, frac("3/2")) })
			}},
		{"MustMetricCountOffset", func() error { return e(g.MetricCountOffset(0, frac("3"))) }, func() { g.MustMetricCountOffset(0, frac("3")) },
			func(t *testing.T) {
				same1(t, func() (duration.Position, error) { return g.MetricCountOffset(0, frac("4")) },
					func() duration.Position { return g.MustMetricCountOffset(0, frac("4")) })
			}},
		{"MustTstampAt", func() error { return e(g.TstampAt(0, pos("0"))) }, func() { g.MustTstampAt(0, pos("0")) },
			func(t *testing.T) {
				same1(t, func() (fraction.Fraction, error) { return g.TstampAt(1, pos("1/8")) }, func() fraction.Fraction { return g.MustTstampAt(1, pos("1/8")) })
			}},
		{"MustTstampOffset", func() error { return e(g.TstampOffset(4, frac("1"))) }, func() { g.MustTstampOffset(4, frac("1")) },
			func(t *testing.T) {
				same1(t, func() (duration.Position, error) { return g.TstampOffset(1, frac("3/2")) },
					func() duration.Position { return g.MustTstampOffset(1, frac("3/2")) })
			}},

		// Splitting: 2.
		{"MustSplit", func() error { return e(g.Split(pos("6"), dur("1"))) }, func() { g.MustSplit(pos("6"), dur("1")) },
			func(t *testing.T) {
				same1(t, func() ([]measure.Part, error) { return g.Split(pos("1/8"), dur("3")) },
					func() []measure.Part { return g.MustSplit(pos("1/8"), dur("3")) })
			}},
		{"MustSplitBy", func() error { return e(g.SplitBy(pos("0"), dur("1"), 0)) }, func() { g.MustSplitBy(pos("0"), dur("1"), 0) },
			func(t *testing.T) {
				same1(t, func() ([]measure.Part, error) { return g.SplitBy(pos("1/8"), dur("3"), measure.LevelBeats) },
					func() []measure.Part { return g.MustSplitBy(pos("1/8"), dur("3"), measure.LevelBeats) })
			}},

		// Changing the grid: 16.
		{"MustWithPickup", func() error { return e(g.WithPickup(dur("5/4"))) }, func() { g.MustWithPickup(dur("5/4")) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.WithPickup(dur("1/8")) }, func() measure.Grid { return g.MustWithPickup(dur("1/8")) })
			}},
		{"MustWithoutPickup", func() error { return e(u.WithoutPickup()) }, func() { u.MustWithoutPickup() },
			func(t *testing.T) { sameGrid(t, g.WithoutPickup, g.MustWithoutPickup) }},
		{"MustWithMark", func() error { return e(g.WithMark(0, mt("3/4"))) }, func() { g.MustWithMark(0, mt("3/4")) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.WithMark(5, mt("2/4")) }, func() measure.Grid { return g.MustWithMark(5, mt("2/4")) })
			}},
		{"MustWithoutMark", func() error { return e(g.WithoutMark(1)) }, func() { g.MustWithoutMark(1) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.WithoutMark(3) }, func() measure.Grid { return g.MustWithoutMark(3) })
			}},
		{"MustWithMetersIn", func() error { return e(g.WithMetersIn(1, 2)) }, func() { g.MustWithMetersIn(1, 2) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.WithMetersIn(5, 6, mt("2/4")) },
					func() measure.Grid { return g.MustWithMetersIn(5, 6, mt("2/4")) })
			}},
		{"MustWithLength", func() error { return e(g.WithLength(9, dur("1/2"), measure.AlignStart)) }, func() { g.MustWithLength(9, dur("1/2"), measure.AlignStart) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.WithLength(6, dur("1/4"), measure.AlignEnd) },
					func() measure.Grid { return g.MustWithLength(6, dur("1/4"), measure.AlignEnd) })
			}},
		{"MustWithoutLength", func() error { return e(g.WithoutLength(0)) }, func() { g.MustWithoutLength(0) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.WithoutLength(4) }, func() measure.Grid { return g.MustWithoutLength(4) })
			}},
		{"MustWithBarCount", func() error { return e(g.WithBarCount(0)) }, func() { g.MustWithBarCount(0) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.WithBarCount(10) }, func() measure.Grid { return g.MustWithBarCount(10) })
			}},
		{"MustWithoutBarCount", func() error { return e(u.WithoutBarCount()) }, func() { u.MustWithoutBarCount() },
			func(t *testing.T) { sameGrid(t, g.WithoutBarCount, g.MustWithoutBarCount) }},
		{"MustTruncate", func() error { return e(g.Truncate(0)) }, func() { g.MustTruncate(0) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.Truncate(4) }, func() measure.Grid { return g.MustTruncate(4) })
			}},
		{"MustAddBars", func() error { return e(three.AddBars(1)) }, func() { three.MustAddBars(1) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.AddBars(2) }, func() measure.Grid { return g.MustAddBars(2) })
			}},
		{"MustInsertBars", func() error { return e(g.InsertBars(10, 1)) }, func() { g.MustInsertBars(10, 1) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.InsertBars(2, 1) }, func() measure.Grid { return g.MustInsertBars(2, 1) })
			}},
		{"MustDeleteBars", func() error { return e(g.DeleteBars(1, 8)) }, func() { g.MustDeleteBars(1, 8) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.DeleteBars(2, 2) }, func() measure.Grid { return g.MustDeleteBars(2, 2) })
			}},
		{"MustSlice", func() error { return e(g.Slice(0, 0)) }, func() { g.MustSlice(0, 0) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.Slice(0, 3) }, func() measure.Grid { return g.MustSlice(0, 3) })
			}},
		{"MustAppend", func() error { return e(three.Append(g)) }, func() { three.MustAppend(g) },
			func(t *testing.T) {
				sameGrid(t, func() (measure.Grid, error) { return g.Append(three) }, func() measure.Grid { return g.MustAppend(three) })
			}},
		{"MustSimplify", func() error { return e(u.Simplify()) }, func() { u.MustSimplify() },
			func(t *testing.T) { sameGrid(t, g.Simplify, g.MustSimplify) }},

		// Equality: 2.
		{"MustEqualMarkup", func() error { return e(g.EqualMarkup(u)) }, func() { g.MustEqualMarkup(u) },
			func(t *testing.T) {
				same1(t, func() (bool, error) { return g.EqualMarkup(three) }, func() bool { return g.MustEqualMarkup(three) })
			}},
		{"MustEqualBarLengths", func() error { return e(u.EqualBarLengths(g)) }, func() { u.MustEqualBarLengths(g) },
			func(t *testing.T) {
				same1(t, func() (bool, error) { return g.EqualBarLengths(g) }, func() bool { return g.MustEqualBarLengths(g) })
			}},
	}
}

// TestMust covers О-23: each Must variant panics with the error of its base
// operation, an error value with the same message and sentinels, and
// returns the same results on success without a panic.
func TestMust(t *testing.T) {
	t.Parallel()
	defer guard(t)

	cases := mustCases(t)
	require.Len(t, cases, 74)

	names := map[string]bool{}
	for _, c := range cases {
		assert.False(t, names[c.name], "duplicate %s", c.name)
		names[c.name] = true
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

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
