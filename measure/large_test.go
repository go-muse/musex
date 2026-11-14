package measure_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-9, large numbers and limits (ИС-16; П-3; П-15; Т-17; КП-45,
// КП-48, КП-57; architecture 4.5…4.12, КР-MS11). No timing is measured:
// results at such numbers confirm that bars are not enumerated.

// TestLargeStarts covers ИС-16 and КП-57: the start of bar 2⁶² and the
// address of position 2⁶² without enumerating bars.
func TestLargeStarts(t *testing.T) {
	t.Parallel()

	t.Run("КП-57 4/4 start of bar 2⁶²", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		s, err := gridOf(t, "4/4").BarStart(1 << 62)
		require.NoError(t, err)
		assertText(t, "4611686018427387903/1", s)
	})

	t.Run("КП-57 3/4 start of bar 2⁶²", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		s, err := gridOf(t, "3/4").BarStart(1 << 62)
		require.NoError(t, err)
		assertText(t, "13835058055282163709/4", s)
	})

	t.Run("КП-57 address of 2⁶²", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		a, err := gridOf(t, "4/4").Locate(posRat(ratInt(pow2(62))))
		require.NoError(t, err)
		assert.Equal(t, measure.MustNewAddress(4611686018427387905, pos("0")), a)
	})

	t.Run("bar 2⁶² after changes", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4;pickup:1/4;3:3/4;4:length:1/2;9:6/8")
		s, err := g.BarStart(1 << 62)
		require.NoError(t, err)

		// 1/4 + 2 + 3/4 + 1/2 + 4·3/4 + (2⁶² − 9)·3/4.
		want := ratAdd(big.NewRat(13, 2), ratMul(big.NewRat(3, 4), ratSub(ratInt(pow2(62)), big.NewRat(9, 1))))
		assert.Equal(t, 0, ratP(s).Cmp(want))

		a, err := g.Locate(s)
		require.NoError(t, err)
		assert.Equal(t, measure.MustNewAddress(1<<62, pos("0")), a)

		o, err := g.Ordinal(1 << 62)
		require.NoError(t, err)
		assert.Equal(t, 1<<62+1, o)
	})

	t.Run("range and split far away", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4;3:3/4")

		l, err := g.RangeLength(3, 1<<62)
		require.NoError(t, err)
		assert.Equal(t, 0, ratD(l).Cmp(ratMul(big.NewRat(3, 4), ratSub(ratInt(pow2(62)), big.NewRat(2, 1)))))

		runs, err := g.MeterRuns(1, 1<<62)
		require.NoError(t, err)
		assert.Equal(t, []measure.MeterRun{{First: 1, Last: 2, Meter: mt("4/4")}, {First: 3, Last: 1 << 62, Meter: mt("3/4")}}, runs)

		full, err := g.AllFull(1, 1<<62)
		require.NoError(t, err)
		assert.True(t, full)

		s, err := g.BarStart(1 << 62)
		require.NoError(t, err)

		parts, err := g.Split(s, dur("3/2"))
		require.NoError(t, err)
		assert.Equal(t, []string{"3/4@4611686018427387904:0/1", "3/4@4611686018427387905:0/1"}, partTexts(parts))
	})
}

// TestLargeTooMany covers КП-45 and architecture 4.11, 4.12: Е-8 without
// enumerating bars.
func TestLargeTooMany(t *testing.T) {
	t.Parallel()
	defer guard(t)

	_, err := gridOf(t, "4/4").Split(pos("0"), durRat(ratInt(pow2(100))))
	assertErr(t, err, measure.ErrTooMany)

	_, err = gridOf(t, "4/4").SplitBy(pos("0"), durRat(ratInt(pow2(200))), measure.LevelUnits)
	assertErr(t, err, measure.ErrTooMany)

	_, err = gridOf(t, "4/4").WithMetersIn(1, math.MaxInt, mt("3/4"), mt("2/4"))
	assertErr(t, err, measure.ErrTooMany)

	g, err := gridOf(t, "4/4").WithMetersIn(2, math.MaxInt, mt("3/4"))
	require.NoError(t, err)
	assertRecord(t, "4/4;2:3/4", g, "one meter on any range; no bar after math.MaxInt")
}

// TestLargeOutOfRange covers КП-48 and architecture 4.2: Е-10 for values
// outside the range of common/fraction and for numbers past math.MaxInt.
func TestLargeOutOfRange(t *testing.T) {
	t.Parallel()

	huge := durRat(ratInt(pow2(254)))
	hugeGrid := func(t *testing.T, finite bool) measure.Grid {
		steps := []step{
			func(g measure.Grid) (measure.Grid, error) { return g.WithLength(1, huge, measure.AlignStart) },
			func(g measure.Grid) (measure.Grid, error) { return g.WithLength(2, huge, measure.AlignStart) },
		}
		if finite {
			steps = append(steps, withBarCount(2))
		}

		return build(t, mt("4/4"), steps...)
	}

	// tinyEnd is 3/1 with bar 2 of length 1/(2²⁵⁴+1) aligned to the end: the
	// start 3 and the end of bar 2 are in range, the shift 3 − 1/(2²⁵⁴+1) is
	// not.
	tinyEnd := func(t *testing.T) measure.Grid {
		l := durRat(ratFrac(big.NewInt(1), new(big.Int).Add(pow2(254), big.NewInt(1))))

		return build(t, mt("3/1"), func(g measure.Grid) (measure.Grid, error) { return g.WithLength(2, l, measure.AlignEnd) })
	}

	// nearOne is 4/4 with bar 1 of length 2²⁵⁴/(2²⁵⁴+1): at the offset 1/3,
	// Elapsed (2²⁵⁴+1)/(3·2²⁵⁴) and Remaining (2²⁵⁵−1)/(3·2²⁵⁴+3) are outside
	// the range, while the position and the offset are in it.
	nearOne := func(t *testing.T) measure.Grid {
		l := durRat(ratFrac(pow2(254), new(big.Int).Add(pow2(254), big.NewInt(1))))

		return build(t, mt("4/4"), func(g measure.Grid) (measure.Grid, error) { return g.WithLength(1, l, measure.AlignStart) })
	}

	// tinyPos is 1/(2²⁵⁵−1): the count 1 + 4·t is outside the range.
	tinyPos := func() duration.Position { return posRat(ratFrac(big.NewInt(1), fraction.MaxComponentBig())) }

	// countNearOne is 1 + 2⁻²⁵⁴: in 4/4 its offset 2⁻²⁵⁶ is outside the range.
	countNearOne := func() fraction.Fraction { return fracRat(ratAdd(big.NewRat(1, 1), ratFrac(big.NewInt(1), pow2(254)))) }

	tests := []struct {
		name   string
		call   func(t *testing.T) error
		causes []error
	}{
		{"КП-48 start of bar 3", func(t *testing.T) error { return e(hugeGrid(t, false).BarStart(3)) }, []error{fraction.ErrOutOfRange}},
		{"end of bar 2", func(t *testing.T) error { return e(hugeGrid(t, false).BarEnd(2)) }, []error{fraction.ErrOutOfRange}},
		{"Bar 3", func(t *testing.T) error { return e(hugeGrid(t, false).Bar(3)) }, []error{fraction.ErrOutOfRange}},
		{"Bars 1…2", func(t *testing.T) error { return e(hugeGrid(t, false).Bars(1, 2)) }, []error{fraction.ErrOutOfRange}},
		{"End", func(t *testing.T) error { return e2(hugeGrid(t, true).End()) }, []error{fraction.ErrOutOfRange}},
		{"RangeLength 1…2", func(t *testing.T) error { return e(hugeGrid(t, false).RangeLength(1, 2)) }, []error{fraction.ErrOutOfRange}},
		{"PositionOf (3, 0)", func(t *testing.T) error { return e(hugeGrid(t, false).PositionOf(ad(3, "0"))) }, []error{fraction.ErrOutOfRange}},
		{"PartOfBar", func(t *testing.T) error {
			tiny := durRat(ratFrac(big.NewInt(1), pow2(254)))
			g := build(t, mt("4/4"), func(g measure.Grid) (measure.Grid, error) { return g.WithLength(1, tiny, measure.AlignStart) })

			return e(g.PartOfBar(1, durRat(ratInt(pow2(254)))))
		}, []error{fraction.ErrOutOfRange}},
		{"Part.End", func(t *testing.T) error {
			return e(measure.Part{Address: ad(1, "0"), Start: posRat(ratInt(pow2(254))), Duration: huge}.End())
		}, []error{fraction.ErrOutOfRange}},
		{"shift of a bar aligned to the end: Bar", func(t *testing.T) error { return e(tinyEnd(t).Bar(2)) }, []error{fraction.ErrOutOfRange}},
		{"shift of a bar aligned to the end: BarAt", func(t *testing.T) error { return e(tinyEnd(t).BarAt(pos("3"))) }, []error{fraction.ErrOutOfRange}},
		{"shift of a bar aligned to the end: Bars", func(t *testing.T) error { return e(tinyEnd(t).Bars(1, 2)) }, []error{fraction.ErrOutOfRange}},
		{"Elapsed by its value", func(t *testing.T) error { return e(nearOne(t).Elapsed(ad(1, "1/3"))) }, []error{fraction.ErrOutOfRange}},
		{"Remaining by its value", func(t *testing.T) error { return e(nearOne(t).Remaining(pos("1/3"))) }, []error{fraction.ErrOutOfRange}},
		{"TstampAt by its value", func(t *testing.T) error { return e(gridOf(t, "4/4").TstampAt(1, tinyPos())) }, []error{fraction.ErrOutOfRange}},
		{"CountAt by its value", func(t *testing.T) error { return e(gridOf(t, "4/4").CountAt(1, tinyPos())) }, []error{fraction.ErrOutOfRange}},
		{"MeterCountOffset by its value", func(t *testing.T) error { return e(measure.MeterCountOffset(mt("4/4"), countNearOne())) }, []error{fraction.ErrOutOfRange}},
		{"CountOffset by its value", func(t *testing.T) error { return e(gridOf(t, "4/4").CountOffset(1, countNearOne())) }, []error{fraction.ErrOutOfRange}},
		{"КП-48 insert math.MaxInt bars", func(t *testing.T) error { return e(gridOf(t, "4/4;2:3/4").InsertBars(1, math.MaxInt)) }, nil},
		{"insert past math.MaxInt in a finite grid", func(t *testing.T) error { return e(gridOf(t, "4/4;bars:4").InsertBars(5, math.MaxInt-3)) }, nil},
		{"КП-48 add to N = math.MaxInt", func(t *testing.T) error { return e(gridOf(t, "4/4;bars:9223372036854775807").AddBars(1)) }, nil},
		{"TotalBars with pickup and N = math.MaxInt", func(t *testing.T) error {
			return e2(gridOf(t, "4/4;pickup:1/4;bars:9223372036854775807").TotalBars())
		}, nil},
		{"Ordinal of math.MaxInt with pickup", func(t *testing.T) error { return e(gridOf(t, "4/4;pickup:1/4").Ordinal(math.MaxInt)) }, nil},
		{"Append past math.MaxInt", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:9223372036854775807").Append(gridOf(t, "3/4;bars:1")))
		}, nil},
		{"Append moving a mark past math.MaxInt", func(t *testing.T) error {
			return e(gridOf(t, "4/4;bars:2").Append(gridOf(t, "3/4;9223372036854775806:2/4")))
		}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			assertErr(t, tt.call(t), measure.ErrOutOfRange, tt.causes...)
		})
	}

	t.Run("in range", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := hugeGrid(t, false)
		s, err := g.BarStart(2)
		require.NoError(t, err)
		assert.Equal(t, 0, ratP(s).Cmp(ratInt(pow2(254))))

		total, finite, err := gridOf(t, "4/4;bars:9223372036854775807").TotalBars()
		require.NoError(t, err)
		assert.True(t, finite)
		assert.Equal(t, math.MaxInt, total)

		o, err := gridOf(t, "4/4;pickup:1/4").Ordinal(math.MaxInt - 1)
		require.NoError(t, err)
		assert.Equal(t, math.MaxInt, o)
	})
}

// TestLargeShift covers architecture 4.6, 4.9, and 4.10: a shift outside the
// range of common/fraction for a valid pickup of 1/(2²⁵⁴+1) in 3/4.
func TestLargeShift(t *testing.T) {
	t.Parallel()

	tinyPickup := func(t *testing.T) measure.Grid {
		a := durRat(ratFrac(big.NewInt(1), new(big.Int).Add(pow2(254), big.NewInt(1))))

		return build(t, mt("3/4"), func(g measure.Grid) (measure.Grid, error) { return g.WithPickup(a) })
	}

	tests := []struct {
		name string
		call func(g measure.Grid) error
	}{
		{"Bar 0", func(g measure.Grid) error { return e(g.Bar(0)) }},
		{"BarAt 0", func(g measure.Grid) error { return e(g.BarAt(pos("0"))) }},
		{"Bars 0…1", func(g measure.Grid) error { return e(g.Bars(0, 1)) }},
		{"MetricPlace (0, 0)", func(g measure.Grid) error { return e2(g.MetricPlace(ad(0, "0"))) }},
		{"MetricCountAt 0", func(g measure.Grid) error { return e(g.MetricCountAt(0, pos("0"))) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			assertErr(t, tt.call(tinyPickup(t)), measure.ErrOutOfRange, fraction.ErrOutOfRange)
		})
	}

	t.Run("Complement", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		a := durRat(ratFrac(big.NewInt(1), new(big.Int).Add(pow2(254), big.NewInt(1))))
		closing, ok, err := measure.Complement(mt("3/4"), a)
		assertErr(t, err, measure.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.False(t, ok)
		assert.False(t, closing.IsSet())
	})

	t.Run("other bars are in range", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := tinyPickup(t)

		b, err := g.Bar(1)
		require.NoError(t, err)
		assertText(t, "0/1", b.Shift)

		seq, err := g.Bars(1, 3)
		require.NoError(t, err)
		assert.Len(t, collect(seq), 3)

		a, err := g.Locate(pos("0"))
		require.NoError(t, err)
		assert.Equal(t, ad(0, "0"), a)

		c, err := g.CountAt(0, pos("0"))
		require.NoError(t, err)
		assertFrac(t, "1", c)

		m, u, err := g.MetricPlace(ad(1, "1/4"))
		require.NoError(t, err)
		assertText(t, "3/4", m)
		assertText(t, "1/4", u)
	})

	t.Run("place out of range by its value", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4;2:length:1/3:end")
		off := posRat(ratFrac(big.NewInt(1), pow2(254)))

		_, _, err := g.MetricPlace(measure.MustNewAddress(2, off))
		assertErr(t, err, measure.ErrOutOfRange, fraction.ErrOutOfRange)

		p, err := g.PositionOf(measure.MustNewAddress(2, off))
		require.NoError(t, err, "the position itself is in range")
		assert.Equal(t, 0, ratP(p).Cmp(ratAdd(big.NewRat(1, 1), ratFrac(big.NewInt(1), pow2(254)))))
	})
}

// TestLargeLastBar covers КР-MS11 and architecture 4.7, 4.8, 4.11, 4.12: an
// open grid has bars up to math.MaxInt and no positions after the end of
// bar math.MaxInt; a finite grid with N = math.MaxInt ends there.
func TestLargeLastBar(t *testing.T) {
	t.Parallel()

	maxPos := posInt(math.MaxInt)        // end of bar math.MaxInt of open 4/4
	beforeMax := posInt(math.MaxInt - 1) // start of bar math.MaxInt

	t.Run("open: start of bar math.MaxInt", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4")

		a, err := g.Locate(beforeMax)
		require.NoError(t, err)
		assert.Equal(t, measure.MustNewAddress(math.MaxInt, pos("0")), a)

		on, err := g.OnBarline(beforeMax)
		require.NoError(t, err)
		assert.True(t, on)

		r, err := g.Remaining(beforeMax)
		require.NoError(t, err)
		assertText(t, "1/1", r)

		b, err := g.BarAt(posRat(ratAdd(ratP(beforeMax), big.NewRat(1, 2))))
		require.NoError(t, err)
		assert.Equal(t, math.MaxInt, b.Number)

		end, err := g.BarEnd(math.MaxInt)
		require.NoError(t, err)
		assert.Equal(t, maxPos, end)

		first, last, err := g.BarsSpanned(beforeMax, dur("1"))
		require.NoError(t, err)
		assert.Equal(t, math.MaxInt, first)
		assert.Equal(t, math.MaxInt, last)

		parts, err := g.Split(beforeMax, dur("1"))
		require.NoError(t, err)
		assert.Equal(t, []string{"1/1@9223372036854775807:0/1"}, partTexts(parts))

		n, err := g.Normalize(measure.MustNewAddress(math.MaxInt-1, pos("3/2")))
		require.NoError(t, err)
		assert.Equal(t, measure.MustNewAddress(math.MaxInt, pos("1/2")), n)
	})

	t.Run("open: at and after the end of bar math.MaxInt", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4")

		for _, p := range []duration.Position{maxPos, posRat(ratAdd(ratP(maxPos), big.NewRat(1, 2)))} {
			_, err := g.Locate(p)
			assertErr(t, err, measure.ErrOutOfRange)

			_, err = g.BarAt(p)
			assertErr(t, err, measure.ErrOutOfRange)

			_, err = g.Remaining(p)
			assertErr(t, err, measure.ErrOutOfRange)

			_, err = g.OnBarline(p)
			assertErr(t, err, measure.ErrOutOfRange)
		}

		_, _, err := g.BarsSpanned(beforeMax, dur("2"))
		assertErr(t, err, measure.ErrOutOfRange)

		_, err = g.Split(beforeMax, dur("2"))
		assertErr(t, err, measure.ErrOutOfRange)

		_, err = g.SplitBy(posInt(math.MaxInt-300), durInt(301), measure.LevelUnits)
		assertErr(t, err, measure.ErrOutOfRange)

		_, err = g.SplitBy(posInt(math.MaxInt-2000), durInt(2001), measure.LevelBars)
		assertErr(t, err, measure.ErrTooMany)

		_, err = g.Normalize(measure.MustNewAddress(math.MaxInt, pos("1")))
		assertErr(t, err, measure.ErrOutOfRange)
	})

	t.Run("finite N = math.MaxInt", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4;bars:9223372036854775807")

		on, err := g.OnBarline(maxPos)
		require.NoError(t, err)
		assert.True(t, on, "the end of the grid is a bar line")

		_, err = g.Locate(maxPos)
		assertErr(t, err, measure.ErrOutsideGrid)

		_, err = g.BarAt(maxPos)
		assertErr(t, err, measure.ErrOutsideGrid)

		_, err = g.Remaining(maxPos)
		assertErr(t, err, measure.ErrOutsideGrid)

		_, _, err = g.BarsSpanned(beforeMax, dur("2"))
		assertErr(t, err, measure.ErrOutsideGrid)

		end, finite, err := g.End()
		require.NoError(t, err)
		assert.True(t, finite)
		assert.Equal(t, maxPos, end)
	})

	t.Run("DeleteBars up to math.MaxInt", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4;5:3/4")

		got, err := g.DeleteBars(3, math.MaxInt)
		require.NoError(t, err)
		assertRecord(t, "4/4;3:3/4", got)

		got, err = g.DeleteBars(7, math.MaxInt)
		require.NoError(t, err)
		assertRecord(t, "4/4;5:3/4", got)

		got, err = gridOf(t, "4/4;bars:9223372036854775807").DeleteBars(2, math.MaxInt)
		require.NoError(t, err)
		assertRecord(t, "4/4;bars:1", got)
	})

	t.Run("bars near math.MaxInt", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4")

		has, err := g.HasBar(math.MaxInt)
		require.NoError(t, err)
		assert.True(t, has)

		l, err := g.RangeLength(math.MaxInt-1, math.MaxInt)
		require.NoError(t, err)
		assertText(t, "2/1", l)

		c, err := g.ChangesWithin(1, math.MaxInt)
		require.NoError(t, err)
		assert.False(t, c)

		got, err := g.WithMark(math.MaxInt, mt("3/4"))
		require.NoError(t, err)

		b, err := got.Bar(math.MaxInt)
		require.NoError(t, err)
		assertText(t, "3/4", b.Meter)
		assert.True(t, b.MeterChange)
	})
}
