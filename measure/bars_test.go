package measure_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/measure"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-2: bars, their meters, lengths, starts, and information, the
// sequence of bars, and the ordinal place (architecture 4.5, 4.6; О-8;
// ИС-1, ИС-2, ИС-5; КП-1, КП-2, КП-5, КП-6, КП-8, КП-9, КП-10, КП-51,
// КП-54, КП-55, КП-58, КП-63, КП-66).

// barValues is the expected meter, length, start, and end of one bar.
type barValues struct {
	bar    int
	meter  string
	length string
	start  string
	end    string
}

// TestBarInfoValues covers ИС-1, ИС-2, ИС-5 and КП-1, КП-2, КП-5, КП-6,
// КП-8, КП-51, КП-54, КП-55, КП-58, КП-63: BarMeter, BarLength, BarStart,
// and BarEnd agree with the fields of Bar.
func TestBarInfoValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid string
		bars []barValues
	}{
		{"КП-1 4/4", "4/4", []barValues{
			{1, "4/4", "1", "0", "1"}, {2, "4/4", "1", "1", "2"}, {5, "4/4", "1", "4", "5"},
		}},
		{"КП-1 3/4", "3/4", []barValues{{4, "3/4", "3/4", "9/4", "3"}}},
		{"КП-1 6/8", "6/8", []barValues{{3, "6/8", "3/4", "3/2", "9/4"}}},
		{"КП-2 3/4 pickup 1/4", "3/4;pickup:1/4", []barValues{
			{0, "3/4", "1/4", "0", "1/4"}, {1, "3/4", "3/4", "1/4", "1"}, {2, "3/4", "3/4", "1", "7/4"},
		}},
		{"КП-2 4/4 pickup 1/4", "4/4;pickup:1/4", []barValues{
			{0, "4/4", "1/4", "0", "1/4"}, {1, "4/4", "1", "1/4", "5/4"}, {2, "4/4", "1", "5/4", "9/4"},
		}},
		{"КП-5 change from bar 3", "4/4;3:3/4", []barValues{
			{2, "4/4", "1", "1", "2"}, {3, "3/4", "3/4", "2", "11/4"}, {4, "3/4", "3/4", "11/4", "7/2"},
			{100, "3/4", "3/4", "299/4", "151/2"},
		}},
		{"КП-6 shorter bar 3", "4/4;3:length:1/2", []barValues{
			{3, "4/4", "1/2", "2", "5/2"}, {4, "4/4", "1", "5/2", "7/2"}, {5, "4/4", "1", "7/2", "9/2"},
		}},
		{"КП-6 longer bar 2", "4/4;2:length:3/2", []barValues{
			{2, "4/4", "3/2", "1", "5/2"}, {3, "4/4", "1", "5/2", "7/2"},
		}},
		{"КП-8 bar 8", "4/4;pickup:1/4;8:length:3/4;bars:8", []barValues{
			{8, "4/4", "3/4", "29/4", "8"},
		}},
		{"КП-51 whole-bar pickup", "4/4;pickup:1/1", []barValues{
			{0, "4/4", "1", "0", "1"}, {1, "4/4", "1", "1", "2"},
		}},
		{"КП-54 equal to the meter", "4/4;3:length:1/1", []barValues{{3, "4/4", "1", "2", "3"}}},
		{"КП-54 after a change to 3/4", "3/4;3:length:1/1", []barValues{
			{3, "3/4", "1", "3/2", "5/2"}, {4, "3/4", "3/4", "5/2", "13/4"},
		}},
		{"КП-55 explicit length kept", "3/4;3:length:1/2", []barValues{
			{3, "3/4", "1/2", "3/2", "2"}, {4, "3/4", "3/4", "2", "11/4"},
		}},
		{"КП-58 4/3", "4/3", []barValues{{1, "4/3", "4/3", "0", "4/3"}, {2, "4/3", "4/3", "4/3", "8/3"}}},
		{"КП-63 pickup replaced by 1/2", "3/4;pickup:1/2", []barValues{
			{1, "3/4", "3/4", "1/2", "5/4"}, {2, "3/4", "3/4", "5/4", "2"},
		}},
		{"КП-63 pickup removed", "3/4", []barValues{{2, "3/4", "3/4", "3/4", "3/2"}}},
		{"КП-63 marks before", "4/4;3:3/4;5:2/4", []barValues{{3, "3/4", "3/4", "2", "11/4"}, {5, "2/4", "1/2", "7/2", "4"}}},
		{"КП-63 mark 3 removed", "4/4;5:2/4", []barValues{{3, "4/4", "1", "2", "3"}, {5, "2/4", "1/2", "4", "9/2"}}},
		{"КП-63 length replaced by 3/2", "4/4;3:length:3/2", []barValues{{4, "4/4", "1", "7/2", "9/2"}}},
		{"КП-63 length removed", "4/4", []barValues{{4, "4/4", "1", "3", "4"}}},
		{"to the end", "4/4;5:length:1/4:end", []barValues{{5, "4/4", "1/4", "4", "17/4"}, {6, "4/4", "1", "17/4", "21/4"}}},
		{"sum of signatures", "2/4+3/8", []barValues{{2, "2/4+3/8", "7/8", "7/8", "7/4"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			for _, want := range tt.bars {
				m, err := g.BarMeter(want.bar)
				require.NoError(t, err)
				assertText(t, mt(want.meter).String(), m, "BarMeter(%d)", want.bar)

				l, err := g.BarLength(want.bar)
				require.NoError(t, err)
				assertText(t, canonicalRat(want.length), l, "BarLength(%d)", want.bar)

				s, err := g.BarStart(want.bar)
				require.NoError(t, err)
				assertText(t, canonicalRat(want.start), s, "BarStart(%d)", want.bar)

				end, err := g.BarEnd(want.bar)
				require.NoError(t, err)
				assertText(t, canonicalRat(want.end), end, "BarEnd(%d)", want.bar)

				b, err := g.Bar(want.bar)
				require.NoError(t, err)
				assert.Equal(t, want.bar, b.Number)
				assert.Equal(t, m, b.Meter)
				assert.Equal(t, l, b.Length)
				assert.Equal(t, s, b.Start)
				assert.Equal(t, end, b.End)
			}
		})
	}
}

// TestBarInfoFields covers О-8 and КП-9, КП-51: all fields of Bar for bars
// 0 to 5 of the grid of КП-9, and a pickup of a whole bar.
func TestBarInfoFields(t *testing.T) {
	t.Parallel()

	kp9 := "4/4;pickup:1/4;3:3/4;4:length:1/2"

	tests := []struct {
		name string
		grid string
		want barOf
	}{
		{"КП-9 bar 0", kp9, barOf{0, "4/4", "1/4", "0", "1/4", measure.KindIncomplete, true, false, measure.AlignEnd, "3/4", false, false}},
		{"КП-9 bar 1", kp9, barOf{1, "4/4", "1", "1/4", "5/4", measure.KindFull, false, false, measure.AlignStart, "0", true, false}},
		{"КП-9 bar 2", kp9, barOf{2, "4/4", "1", "5/4", "9/4", measure.KindFull, false, false, measure.AlignStart, "0", false, false}},
		{"КП-9 bar 3", kp9, barOf{3, "3/4", "3/4", "9/4", "3", measure.KindFull, false, false, measure.AlignStart, "0", true, true}},
		{"КП-9 bar 4", kp9, barOf{4, "3/4", "1/2", "3", "7/2", measure.KindIncomplete, false, true, measure.AlignStart, "0", false, false}},
		{"КП-9 bar 5", kp9, barOf{5, "3/4", "3/4", "7/2", "17/4", measure.KindFull, false, false, measure.AlignStart, "0", false, false}},
		{"КП-51 whole-bar pickup", "4/4;pickup:1/1", barOf{0, "4/4", "1", "0", "1", measure.KindFull, true, false, measure.AlignEnd, "0", false, false}},
		{"КП-6 extended", "4/4;2:length:3/2", barOf{2, "4/4", "3/2", "1", "5/2", measure.KindExtended, false, true, measure.AlignStart, "0", false, false}},
		{"КП-54 explicit equal to the meter", "4/4;3:length:1/1", barOf{3, "4/4", "1", "2", "3", measure.KindFull, false, true, measure.AlignStart, "0", false, false}},
		{"КП-54 becomes extended", "3/4;3:length:1/1", barOf{3, "3/4", "1", "3/2", "5/2", measure.KindExtended, false, true, measure.AlignStart, "0", false, false}},
		{"to the end", "4/4;5:length:1/4:end", barOf{5, "4/4", "1/4", "4", "17/4", measure.KindIncomplete, false, true, measure.AlignEnd, "3/4", false, false}},
		{"to the end, whole", "4/4;5:length:1/1:end", barOf{5, "4/4", "1", "4", "5", measure.KindFull, false, true, measure.AlignEnd, "0", false, false}},
		{"repeated mark", "4/4;5:4/4", barOf{5, "4/4", "1", "4", "5", measure.KindFull, false, false, measure.AlignStart, "0", true, false}},
		{"change to C", "4/4;5:C", barOf{5, "C", "1", "4", "5", measure.KindFull, false, false, measure.AlignStart, "0", true, true}},
		{"after a change", "4/4;5:C", barOf{6, "C", "1", "5", "6", measure.KindFull, false, false, measure.AlignStart, "0", false, false}},
		{"pickup with change from bar 1", "6/8;pickup:1/8", barOf{0, "6/8", "1/8", "0", "1/8", measure.KindIncomplete, true, false, measure.AlignEnd, "5/8", false, false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			b, err := g.Bar(tt.want.number)
			require.NoError(t, err)
			assert.Equal(t, tt.want.value(), b)
			assert.True(t, b.Equal(tt.want.value()))
		})
	}
}

// TestBarInfoErrors covers architecture 4.2 for the bar operations and
// КП-41: Е-2 for a negative number, Е-4 for a missing bar.
func TestBarInfoErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid string
		bar  int
		err  error
	}{
		{"negative", "4/4", -1, measure.ErrInvalidArgument},
		{"min int", "4/4;pickup:1/4", math.MinInt, measure.ErrInvalidArgument},
		{"КП-41 bar 0 without pickup", "4/4", 0, measure.ErrOutsideGrid},
		{"КП-41 bar 5 of N = 4", "4/4;bars:4", 5, measure.ErrOutsideGrid},
		{"max int of a finite grid", "4/4;bars:4", math.MaxInt, measure.ErrOutsideGrid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			b, err := g.Bar(tt.bar)
			assertErr(t, err, tt.err)
			assert.Equal(t, measure.Bar{}, b)

			m, err := g.BarMeter(tt.bar)
			assertErr(t, err, tt.err)
			assert.False(t, m.IsSet())

			l, err := g.BarLength(tt.bar)
			assertErr(t, err, tt.err)
			assert.False(t, l.IsSet())

			s, err := g.BarStart(tt.bar)
			assertErr(t, err, tt.err)
			assert.False(t, s.IsSet())

			end, err := g.BarEnd(tt.bar)
			assertErr(t, err, tt.err)
			assert.False(t, end.IsSet())

			o, err := g.Ordinal(tt.bar)
			assertErr(t, err, tt.err)
			assert.Zero(t, o)
		})
	}
}

// TestBarInfoLastBar covers П-2: in a finite grid the last bar N exists, and
// its end is the end of the grid.
func TestBarInfoLastBar(t *testing.T) {
	t.Parallel()
	defer guard(t)

	g := gridOf(t, "4/4;3:3/4;bars:4")

	end, err := g.BarEnd(4)
	require.NoError(t, err)

	gridEnd, finite, err := g.End()
	require.NoError(t, err)
	require.True(t, finite)
	assert.Equal(t, gridEnd, end)
	assertText(t, "7/2", end)
}

// TestBarsIter covers О-8, КП-9, and КП-66: Bars yields the information of
// each bar of a range in order, equal to Bar, and can be iterated again.
func TestBarsIter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		grid        string
		first, last int
	}{
		{"КП-66 bars 0…5 of КП-9", "4/4;pickup:1/4;3:3/4;4:length:1/2", 0, 5},
		{"КП-66 one bar 3…3", "4/4;pickup:1/4;3:3/4;4:length:1/2", 3, 3},
		{"from bar 1", "4/4;3:3/4;5:length:1/4:end", 1, 7},
		{"finite up to N", "4/4;3:3/4;bars:4", 2, 4},
		{"pickup only", "3/4;pickup:1/4", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			var want []measure.Bar

			for k := tt.first; k <= tt.last; k++ {
				b, err := g.Bar(k)
				require.NoError(t, err)

				want = append(want, b)
			}

			seq, err := g.Bars(tt.first, tt.last)
			require.NoError(t, err)
			require.NotNil(t, seq)

			assert.Equal(t, want, collect(seq))
			assert.Equal(t, want, collect(seq), "a second iteration gives the same bars")
		})
	}

	t.Run("КП-66 values of КП-9", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4;pickup:1/4;3:3/4;4:length:1/2")
		seq, err := g.Bars(0, 5)
		require.NoError(t, err)

		bars := collect(seq)
		require.Len(t, bars, 6)

		starts := []string{"0/1", "1/4", "5/4", "9/4", "3/1", "7/2"}
		for i, b := range bars {
			assert.Equal(t, i, b.Number)
			assertText(t, starts[i], b.Start)
		}

		assert.Equal(t, barOf{4, "3/4", "1/2", "3", "7/2", measure.KindIncomplete, false, true, measure.AlignStart, "0", false, false}.value(), bars[4])
	})
}

// TestBarsIterBreak covers КП-66 and ИС-16: iteration over bars 1…2⁶² of an
// open grid stops after three bars, and a range near math.MaxInt ends.
func TestBarsIterBreak(t *testing.T) {
	t.Parallel()

	t.Run("КП-66 1…2⁶²", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4")
		seq, err := g.Bars(1, 1<<62)
		require.NoError(t, err)

		var starts []string

		for b := range seq {
			starts = append(starts, b.Start.String())
			if len(starts) == 3 {
				break
			}
		}

		assert.Equal(t, []string{"0/1", "1/1", "2/1"}, starts)
	})

	t.Run("up to math.MaxInt", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4")
		seq, err := g.Bars(math.MaxInt-2, math.MaxInt)
		require.NoError(t, err)

		var numbers []int
		for b := range seq {
			numbers = append(numbers, b.Number)
		}

		assert.Equal(t, []int{math.MaxInt - 2, math.MaxInt - 1, math.MaxInt}, numbers)
	})
}

// TestBarsIterErrors covers architecture 4.6: all failures of Bars are
// returned by the call, before iteration, in the order Е-2 (first < 0),
// Е-2 (first > last), Е-4 (first, then last).
func TestBarsIterErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		grid        string
		first, last int
		err         error
	}{
		{"first < 0", "4/4", -1, 2, measure.ErrInvalidArgument},
		{"КП-39 3…2", "4/4", 3, 2, measure.ErrInvalidArgument},
		{"first < 0 and first > last", "4/4", -1, -2, measure.ErrInvalidArgument},
		{"first > last and missing first", "4/4;bars:4", 6, 5, measure.ErrInvalidArgument},
		{"bar 0 without pickup", "4/4", 0, 2, measure.ErrOutsideGrid},
		{"КП-41 last after N", "4/4;bars:4", 3, 5, measure.ErrOutsideGrid},
		{"first after N", "4/4;bars:4", 5, 6, measure.ErrOutsideGrid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			seq, err := g.Bars(tt.first, tt.last)
			assertErr(t, err, tt.err)
			assert.Nil(t, seq)
		})
	}
}

// TestOrdinal covers 2.2 and КП-10: the ordinal place counts from 1 with
// the pickup first, and BarAtOrdinal is its inverse.
func TestOrdinal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		grid    string
		bar     int
		ordinal int
	}{
		{"КП-10 pickup, bar 0", "4/4;pickup:1/4", 0, 1},
		{"КП-10 pickup, bar 3", "4/4;pickup:1/4", 3, 4},
		{"КП-10 no pickup, bar 3", "4/4", 3, 3},
		{"no pickup, bar 1", "4/4", 1, 1},
		{"finite with pickup, bar N", "4/4;pickup:1/4;bars:4", 4, 5},
		{"open, max int", "4/4", math.MaxInt, math.MaxInt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			o, err := g.Ordinal(tt.bar)
			require.NoError(t, err)
			assert.Equal(t, tt.ordinal, o)

			bar, err := g.BarAtOrdinal(tt.ordinal)
			require.NoError(t, err)
			assert.Equal(t, tt.bar, bar)
		})
	}
}

// TestOrdinalErrors covers КП-39, КП-41, and architecture 4.2 for
// BarAtOrdinal: Е-2 for a place below 1, Е-4 for a place after the bars.
func TestOrdinalErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		grid    string
		ordinal int
		err     error
	}{
		{"КП-39 place 0", "4/4", 0, measure.ErrInvalidArgument},
		{"negative", "4/4;pickup:1/4", -1, measure.ErrInvalidArgument},
		{"КП-41 place 5 of N = 4", "4/4;bars:4", 5, measure.ErrOutsideGrid},
		{"place 6 of N = 4 with pickup", "4/4;pickup:1/4;bars:4", 6, measure.ErrOutsideGrid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			bar, err := g.BarAtOrdinal(tt.ordinal)
			assertErr(t, err, tt.err)
			assert.Zero(t, bar)
		})
	}

	t.Run("place 5 with pickup and N = 4", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4;pickup:1/4;bars:4")
		bar, err := g.BarAtOrdinal(5)
		require.NoError(t, err)
		assert.Equal(t, 4, bar)
	})
}
