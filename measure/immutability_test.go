package measure_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-9, immutability (architecture 3.2; ИС-12, ИС-15; П-1; Т-16;
// КП-56).

// TestImmutabilityOperations covers КП-56: after each failure of КП-39…КП-48
// and after successful changes, the receiver is equal by record to its copy
// parsed from its text before the call, and its text is the same.
func TestImmutabilityOperations(t *testing.T) {
	t.Parallel()

	type op func(g measure.Grid) error

	ops := []struct {
		name string
		op   op
	}{
		{"WithPickup 5/4", func(g measure.Grid) error { return e(g.WithPickup(dur("5/4"))) }},
		{"WithPickup unset", func(g measure.Grid) error { return e(g.WithPickup(duration.Duration{})) }},
		{"WithPickup 1/8", func(g measure.Grid) error { return e(g.WithPickup(dur("1/8"))) }},
		{"WithoutPickup", func(g measure.Grid) error { return e(g.WithoutPickup()) }},
		{"WithMark 0", func(g measure.Grid) error { return e(g.WithMark(0, mt("3/4"))) }},
		{"WithMark unset", func(g measure.Grid) error { return e(g.WithMark(3, meter.Meter{})) }},
		{"WithMark 1 to 2/4", func(g measure.Grid) error { return e(g.WithMark(1, mt("2/4"))) }},
		{"WithMark 4 to 3/4", func(g measure.Grid) error { return e(g.WithMark(4, mt("3/4"))) }},
		{"WithMark 5 after N", func(g measure.Grid) error { return e(g.WithMark(5, mt("3/4"))) }},
		{"WithoutMark 1", func(g measure.Grid) error { return e(g.WithoutMark(1)) }},
		{"WithoutMark 3", func(g measure.Grid) error { return e(g.WithoutMark(3)) }},
		{"WithMetersIn empty", func(g measure.Grid) error { return e(g.WithMetersIn(1, 2)) }},
		{"WithMetersIn 1…1025", func(g measure.Grid) error { return e(g.WithMetersIn(1, 1025, mt("3/4"), mt("2/4"))) }},
		{"WithMetersIn 1…2", func(g measure.Grid) error { return e(g.WithMetersIn(1, 2, mt("6/8"), mt("2/4"))) }},
		{"WithLength 6", func(g measure.Grid) error { return e(g.WithLength(6, dur("1/2"), measure.AlignStart)) }},
		{"WithLength 5/4 to the end", func(g measure.Grid) error { return e(g.WithLength(2, dur("5/4"), measure.AlignEnd)) }},
		{"WithLength 2", func(g measure.Grid) error { return e(g.WithLength(2, dur("1/2"), measure.AlignStart)) }},
		{"WithoutLength", func(g measure.Grid) error { return e(g.WithoutLength(5)) }},
		{"WithBarCount 0", func(g measure.Grid) error { return e(g.WithBarCount(0)) }},
		{"WithBarCount 2", func(g measure.Grid) error { return e(g.WithBarCount(2)) }},
		{"WithBarCount 9", func(g measure.Grid) error { return e(g.WithBarCount(9)) }},
		{"WithoutBarCount", func(g measure.Grid) error { return e(g.WithoutBarCount()) }},
		{"Truncate 2", func(g measure.Grid) error { return e(g.Truncate(2)) }},
		{"AddBars", func(g measure.Grid) error { return e(g.AddBars(2)) }},
		{"AddBars max int", func(g measure.Grid) error { return e(g.AddBars(math.MaxInt)) }},
		{"InsertBars", func(g measure.Grid) error { return e(g.InsertBars(2, 3)) }},
		{"InsertBars max int", func(g measure.Grid) error { return e(g.InsertBars(1, math.MaxInt)) }},
		{"InsertBars 0", func(g measure.Grid) error { return e(g.InsertBars(0, 1)) }},
		{"DeleteBars 1…4", func(g measure.Grid) error { return e(g.DeleteBars(1, 4)) }},
		{"DeleteBars 1…2", func(g measure.Grid) error { return e(g.DeleteBars(1, 2)) }},
		{"DeleteBars 2…3", func(g measure.Grid) error { return e(g.DeleteBars(2, 3)) }},
		{"Slice 0…0", func(g measure.Grid) error { return e(g.Slice(0, 0)) }},
		{"Slice 1…3", func(g measure.Grid) error { return e(g.Slice(1, 3)) }},
		{"Append", func(g measure.Grid) error { return e(g.Append(gridOf(t, "3/4;pickup:1/4"))) }},
		{"Simplify", func(g measure.Grid) error { return e(g.Simplify()) }},
		{"Split", func(g measure.Grid) error { return e(g.Split(pos("1/8"), dur("2"))) }},
		{"SplitBy groups", func(g measure.Grid) error { return e(g.SplitBy(pos("0"), dur("3"), measure.LevelGroups)) }},
		{"Locate", func(g measure.Grid) error { return e(g.Locate(pos("9"))) }},
		{"PositionOf", func(g measure.Grid) error { return e(g.PositionOf(ad(1, "1"))) }},
		{"Normalize", func(g measure.Grid) error { return e(g.Normalize(ad(1, "7/2"))) }},
		{"MetricPlace", func(g measure.Grid) error { return e2(g.MetricPlace(ad(2, "9/8"))) }},
		{"TstampAt", func(g measure.Grid) error { return e(g.TstampAt(0, pos("0"))) }},
		{"Bars", func(g measure.Grid) error {
			seq, err := g.Bars(0, 4)
			if err == nil {
				collect(seq)
			}

			return err
		}},
		{"UnmarshalText of a copy", func(g measure.Grid) error {
			c := g

			return c.UnmarshalText([]byte("3/4;bars:2"))
		}},
	}

	grids := []string{
		"4/4;pickup:3/4;3:2/4;5:length:1/2:end;bars:5",
		"4/4;pickup:1/4;3:3/4;4:length:1/2;bars:6",
		"4/4;2:length:5/4;3:5/8",
		"4/4;bars:9223372036854775807",
	}

	for _, text := range grids {
		for _, o := range ops {
			t.Run(text+" "+o.name, func(t *testing.T) {
				t.Parallel()
				defer guard(t)

				g := gridOf(t, text)
				before := g.String()
				copyBefore, err := measure.Parse(before)
				require.NoError(t, err)

				_ = o.op(g)

				assert.Equal(t, before, g.String())
				assert.True(t, g.Equal(copyBefore), "the receiver equals itself before the call")
			})
		}
	}
}

// TestImmutabilityArguments covers КП-56: a grid passed as an argument and
// a series of meters are not changed.
func TestImmutabilityArguments(t *testing.T) {
	t.Parallel()
	defer guard(t)

	g := gridOf(t, "4/4;bars:2")
	other := gridOf(t, "3/4;pickup:1/4;3:2/4;bars:4")
	before := other.String()

	_, err := g.Append(other)
	require.NoError(t, err)

	_, err = g.EqualMarkup(other)
	require.NoError(t, err)

	_, err = g.EqualBarLengths(other)
	require.NoError(t, err)

	assert.Equal(t, before, other.String())

	series := mts("6/8", "3/4")
	_, err = gridOf(t, "4/4").WithMetersIn(1, 4, series...)
	require.NoError(t, err)
	assert.Equal(t, mts("6/8", "3/4"), series, "the series is only read")
}

// TestImmutabilityResults covers architecture 3.2 and КП-56: the slices of
// results are new at each call and belong to the caller; changing them does
// not change the grid or the next results.
func TestImmutabilityResults(t *testing.T) {
	t.Parallel()

	g := func(t *testing.T) measure.Grid { return gridOf(t, "4/4;pickup:1/4;3:3/4;5:2/4;4:length:1/2;bars:8") }

	t.Run("Marks", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		grid := g(t)
		first, err := grid.Marks()
		require.NoError(t, err)

		want := append([]measure.Mark(nil), first...)
		first[0] = measure.Mark{Bar: 99, Meter: mt("7/8")}

		again, err := grid.Marks()
		require.NoError(t, err)
		assert.Equal(t, want, again)
	})

	t.Run("Changes", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		grid := g(t)
		first, err := grid.Changes()
		require.NoError(t, err)

		want := append([]measure.Mark(nil), first...)
		first[0] = measure.Mark{}

		again, err := grid.Changes()
		require.NoError(t, err)
		assert.Equal(t, want, again)
	})

	t.Run("Lengths", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		grid := g(t)
		first, err := grid.Lengths()
		require.NoError(t, err)

		want := append([]measure.ExplicitLength(nil), first...)
		first[0].Bar = 77

		again, err := grid.Lengths()
		require.NoError(t, err)
		assert.Equal(t, want, again)

		b, err := grid.BarLength(4)
		require.NoError(t, err)
		assertText(t, "1/2", b)
	})

	t.Run("MeterRuns", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		grid := g(t)
		first, err := grid.MeterRuns(0, 8)
		require.NoError(t, err)

		want := append([]measure.MeterRun(nil), first...)
		first[0].Last = 100

		again, err := grid.MeterRuns(0, 8)
		require.NoError(t, err)
		assert.Equal(t, want, again)
	})

	t.Run("Split and SplitBy", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		grid := g(t)

		first, err := grid.Split(pos("0"), dur("3"))
		require.NoError(t, err)

		want := append([]measure.Part(nil), first...)
		first[0] = measure.Part{}

		again, err := grid.Split(pos("0"), dur("3"))
		require.NoError(t, err)
		assert.Equal(t, want, again)

		by, err := grid.SplitBy(pos("0"), dur("3"), measure.LevelUnits)
		require.NoError(t, err)

		wantBy := append([]measure.Part(nil), by...)
		by[1] = measure.Part{}

		againBy, err := grid.SplitBy(pos("0"), dur("3"), measure.LevelUnits)
		require.NoError(t, err)
		assert.Equal(t, wantBy, againBy)
	})

	t.Run("MarshalText", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		grid := g(t)
		b, err := grid.MarshalText()
		require.NoError(t, err)

		want := string(b)
		b[0] = 'X'

		again, err := grid.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, want, string(again))
		assert.Equal(t, want, grid.String())
	})

	t.Run("AppendText keeps dst", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		grid := g(t)
		dst := make([]byte, 3, 256)
		copy(dst, "abc")

		out, err := grid.AppendText(dst)
		require.NoError(t, err)
		assert.Equal(t, "abc", string(dst[:3]))
		assert.Equal(t, "abc"+grid.String(), string(out))

		a := ad(2, "1/8")
		adst := make([]byte, 2, 64)
		copy(adst, "@@")

		aout, err := a.AppendText(adst)
		require.NoError(t, err)
		assert.Equal(t, "@@", string(adst[:2]))
		assert.Equal(t, "@@2:1/8", string(aout))
	})

	t.Run("Bars after a partial iteration", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		grid := g(t)
		seq, err := grid.Bars(0, 6)
		require.NoError(t, err)

		for b := range seq {
			if b.Number == 2 {
				break
			}
		}

		bars := collect(seq)
		require.Len(t, bars, 7)
		assert.Equal(t, 0, bars[0].Number, "each iteration starts at first")
	})

	t.Run("assignment copies the value", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		grid := g(t)
		c := grid

		changed, err := c.WithMark(6, mt("C"))
		require.NoError(t, err)

		c = changed
		assert.False(t, grid.Equal(c))
		assert.Equal(t, "4/4;pickup:1/4;3:3/4;4:length:1/2;5:2/4;bars:8", grid.String())
	})
}
