package measure_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-7: changes of a grid by ranges of bars and simplification
// (architecture 4.12, 4.14; О-4, О-7, О-7а, О-19; П-12; КП-31…КП-34,
// КП-36, КП-39…КП-41, КП-43, КП-45, КП-48, КП-67, КП-68).

// editCase is a change of a grid with its expected record or error.
type editCase struct {
	name   string
	grid   string
	edit   func(measure.Grid) (measure.Grid, error)
	want   string // the record in the notation of gridOf, or "" for an error
	err    error
	causes []error
	starts map[int]string // expected starts of bars of the result
	end    string         // expected end of a finite result
}

// runEdits runs edit cases: the result has the expected record, starts,
// and end; a failure returns the unset grid; the receiver is unchanged.
func runEdits(t *testing.T, tests []editCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			got, err := tt.edit(g)

			assertRecord(t, tt.grid, g, "the receiver is unchanged")

			if tt.err != nil {
				assertErr(t, err, tt.err, tt.causes...)
				assert.False(t, got.IsSet(), "the unset grid on error")

				return
			}

			require.NoError(t, err)
			assertRecord(t, tt.want, got)

			for bar, want := range tt.starts {
				s, err := got.BarStart(bar)
				require.NoError(t, err)
				assertText(t, canonicalRat(want), s, "BarStart(%d)", bar)
			}

			if tt.end != "" {
				end, finite, err := got.End()
				require.NoError(t, err)
				assert.True(t, finite)
				assertText(t, canonicalRat(tt.end), end)
			}
		})
	}
}

func metersIn(first, last int, texts ...string) func(measure.Grid) (measure.Grid, error) {
	return func(g measure.Grid) (measure.Grid, error) { return g.WithMetersIn(first, last, mts(texts...)...) }
}

// TestEditMetersIn covers О-4, П-12, and КП-31, КП-32, КП-39, КП-41, КП-45:
// bars of a range take the meters of a series in order and in a cycle; the
// bars outside keep their meters.
func TestEditMetersIn(t *testing.T) {
	t.Parallel()

	runEdits(t, []editCase{
		{name: "КП-31 1…4 to 3/4", grid: "4/4", edit: metersIn(1, 4, "3/4"), want: "3/4;5:4/4", starts: map[int]string{5: "3"}},
		{name: "КП-31 7…7 to 5/4", grid: "4/4", edit: metersIn(7, 7, "5/4"), want: "4/4;7:5/4;8:4/4", starts: map[int]string{8: "29/4", 9: "33/4"}},
		{name: "КП-31 3…4 to 4/4", grid: "4/4;3:3/4", edit: metersIn(3, 4, "4/4"), want: "4/4;3:4/4;5:3/4"},
		{name: "КП-31 2…2 to 3/4", grid: "4/4;3:3/4", edit: metersIn(2, 2, "3/4"), want: "4/4;2:3/4;3:3/4"},
		{name: "КП-32 alternation 6/8, 3/4", grid: "4/4", edit: metersIn(1, 4, "6/8", "3/4"), want: "6/8;2:3/4;3:6/8;4:3/4;5:4/4"},
		{name: "КП-32 series 3/4, 3/4, 2/4", grid: "4/4", edit: metersIn(2, 6, "3/4", "3/4", "2/4"), want: "4/4;2:3/4;4:2/4;5:3/4;7:4/4"},
		{name: "marks inside are removed", grid: "4/4;3:3/4;4:2/4;6:C", edit: metersIn(2, 5, "6/8"), want: "4/4;2:6/8;6:C"},
		{name: "same meter after the range", grid: "4/4;5:3/4", edit: metersIn(2, 4, "3/4"), want: "4/4;2:3/4;5:3/4"},
		{name: "no mark needed after the range", grid: "3/4", edit: metersIn(2, 3, "4/4", "3/4"), want: "3/4;2:4/4;3:3/4"},
		{name: "explicit lengths unchanged", grid: "4/4;3:length:1/2", edit: metersIn(2, 4, "3/4"), want: "4/4;2:3/4;5:4/4;3:length:1/2"},
		{name: "finite up to N", grid: "4/4;bars:4", edit: metersIn(3, 4, "3/4"), want: "4/4;3:3/4;bars:4"},
		{name: "one meter on a long range", grid: "4/4", edit: metersIn(1, 1<<40, "3/4"), want: "3/4;1099511627777:4/4"},
		{name: "1024 bars of an alternation", grid: "4/4", edit: metersIn(1, 1024, "3/4", "3/4"), want: "3/4;1025:4/4"},
		{name: "pickup that fits", grid: "4/4;pickup:1/4", edit: metersIn(1, 2, "2/4"), want: "2/4;pickup:1/4;3:4/4"},
		{name: "КП-39 empty series", grid: "4/4", edit: metersIn(1, 2), err: measure.ErrInvalidArgument},
		{name: "КП-39 unset meter in the series", grid: "4/4", edit: func(g measure.Grid) (measure.Grid, error) {
			return g.WithMetersIn(1, 2, mt("3/4"), meter.Meter{})
		}, err: measure.ErrInvalidArgument, causes: []error{meter.ErrUnset}},
		{name: "bar 0", grid: "4/4;pickup:1/4", edit: metersIn(0, 2, "3/4"), err: measure.ErrInvalidArgument},
		{name: "КП-39 3…2", grid: "4/4", edit: metersIn(3, 2, "3/4"), err: measure.ErrInvalidArgument},
		{name: "КП-41 4…5 of N = 4", grid: "4/4;bars:4", edit: metersIn(4, 5, "3/4"), err: measure.ErrOutsideGrid},
		{name: "КП-45 1…1025 alternation", grid: "4/4", edit: metersIn(1, 1025, "3/4", "2/4"), err: measure.ErrTooMany},
		{name: "КП-45 even with equal meters", grid: "4/4", edit: metersIn(1, 1025, "3/4", "3/4"), err: measure.ErrTooMany},
		{name: "pickup longer than the new meter", grid: "4/4;pickup:3/4", edit: metersIn(1, 2, "2/4"), err: measure.ErrLongerThanMeter},
		{name: "to the end longer than the new meter", grid: "4/4;3:length:7/8:end", edit: metersIn(2, 3, "3/4"), err: measure.ErrLongerThanMeter},
	})

	t.Run("order of checks", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4;bars:4")

		_, err := g.WithMetersIn(0, -1)
		assertErr(t, err, measure.ErrInvalidArgument)

		_, err = g.WithMetersIn(2, 1, meter.Meter{})
		assertErr(t, err, measure.ErrInvalidArgument)
		assertNoCause(t, err, meter.ErrUnset)

		_, err = g.WithMetersIn(5, 6, meter.Meter{})
		assertErr(t, err, measure.ErrInvalidArgument, meter.ErrUnset)

		_, err = gridOf(t, "4/4;bars:2000").WithMetersIn(1, 2001, mt("3/4"), mt("2/4"))
		assertErr(t, err, measure.ErrOutsideGrid)

		_, err = gridOf(t, "4/4;pickup:3/4").WithMetersIn(1, 1025, mt("2/4"), mt("3/4"))
		assertErr(t, err, measure.ErrTooMany)
	})
}

// TestEditMetersInSlice covers КП-56: the series of meters is only read
// during the call.
func TestEditMetersInSlice(t *testing.T) {
	t.Parallel()
	defer guard(t)

	series := mts("6/8", "3/4")
	got, err := gridOf(t, "4/4").WithMetersIn(1, 4, series...)
	require.NoError(t, err)

	series[0] = mt("2/4")
	series[1] = mt("5/4")

	assertRecord(t, "6/8;2:3/4;3:6/8;4:3/4;5:4/4", got)
}

// TestEditInsert covers О-7, П-12, and КП-33, КП-39, КП-41: inserted bars
// are full with the meter of the bar before; the records after shift.
func TestEditInsert(t *testing.T) {
	t.Parallel()

	insert := func(before, m int) func(measure.Grid) (measure.Grid, error) {
		return func(g measure.Grid) (measure.Grid, error) { return g.InsertBars(before, m) }
	}

	runEdits(t, []editCase{
		{name: "КП-33 2 before 3", grid: "4/4;3:3/4", edit: insert(3, 2), want: "4/4;5:3/4", starts: map[int]string{5: "4"}},
		{name: "КП-33 1 before 1", grid: "4/4;3:3/4", edit: insert(1, 1), want: "4/4;4:3/4"},
		{name: "КП-33 finite with explicit length", grid: "4/4;3:3/4;3:length:1/2;bars:6", edit: insert(3, 2), want: "4/4;5:3/4;5:length:1/2;bars:8", end: "27/4"},
		{name: "meter of the bar before", grid: "4/4;3:3/4;6:2/4", edit: insert(5, 2), want: "4/4;3:3/4;8:2/4", starts: map[int]string{5: "7/2", 7: "5", 8: "23/4"}},
		{name: "before bar 1 with a later change", grid: "3/4;2:2/4", edit: insert(1, 2), want: "3/4;4:2/4"},
		{name: "explicit length before is not copied", grid: "4/4;2:length:1/2", edit: insert(3, 1), want: "4/4;2:length:1/2", starts: map[int]string{3: "3/2", 4: "5/2"}},
		{name: "append to a finite grid", grid: "4/4;bars:4", edit: insert(5, 2), want: "4/4;bars:6", end: "6"},
		{name: "pickup kept", grid: "4/4;pickup:1/4;2:3/4", edit: insert(1, 1), want: "4/4;pickup:1/4;3:3/4"},
		{name: "КП-39 before 0", grid: "4/4", edit: insert(0, 1), err: measure.ErrInvalidArgument},
		{name: "КП-39 before −1", grid: "4/4", edit: insert(-1, 1), err: measure.ErrInvalidArgument},
		{name: "КП-39 insert 0", grid: "4/4", edit: insert(1, 0), err: measure.ErrInvalidArgument},
		{name: "negative count", grid: "4/4", edit: insert(1, -2), err: measure.ErrInvalidArgument},
		{name: "КП-41 before 6 of N = 4", grid: "4/4;bars:4", edit: insert(6, 1), err: measure.ErrOutsideGrid},
		{name: "count before missing bar", grid: "4/4;bars:4", edit: insert(6, 0), err: measure.ErrInvalidArgument},
	})
}

// TestEditDelete covers О-7, П-12, and КП-34, КП-39…КП-41: the records of
// deleted bars are dropped, the later ones shift, and the meter of the bar
// after the range is kept.
func TestEditDelete(t *testing.T) {
	t.Parallel()

	del := func(first, last int) func(measure.Grid) (measure.Grid, error) {
		return func(g measure.Grid) (measure.Grid, error) { return g.DeleteBars(first, last) }
	}

	runEdits(t, []editCase{
		{name: "КП-34 2…3", grid: "4/4;3:3/4;6:2/4;2:length:1/2;7:length:1/4", edit: del(2, 3), want: "4/4;2:3/4;4:2/4;5:length:1/4"},
		{name: "КП-34 1…2", grid: "4/4;3:3/4", edit: del(1, 2), want: "3/4"},
		{name: "КП-34 2…3 marks 2 and 5", grid: "4/4;2:3/4;5:2/4", edit: del(2, 3), want: "4/4;2:3/4;3:2/4"},
		{name: "КП-34 3…4 marks 2 and 5", grid: "4/4;2:3/4;5:2/4", edit: del(3, 4), want: "4/4;2:3/4;3:2/4"},
		{name: "КП-34 finite 5…6", grid: "4/4;3:3/4;3:length:1/2;bars:6", edit: del(5, 6), want: "4/4;3:3/4;3:length:1/2;bars:4", end: "13/4"},
		{name: "first bar keeps its meter", grid: "4/4;2:3/4", edit: del(1, 1), want: "3/4"},
		{name: "repeated meter after the range", grid: "4/4;3:3/4", edit: del(2, 2), want: "4/4;2:3/4", starts: map[int]string{2: "1"}},
		{name: "open, the rest of the grid", grid: "4/4;5:3/4", edit: del(3, math.MaxInt), want: "4/4;3:3/4"},
		{name: "open, after the last mark", grid: "4/4;5:3/4", edit: del(7, math.MaxInt), want: "4/4;5:3/4"},
		{name: "pickup fits the new bar 1", grid: "4/4;pickup:1/4;3:2/4", edit: del(1, 2), want: "2/4;pickup:1/4"},
		{name: "КП-39 delete all of N = 4", grid: "4/4;bars:4", edit: del(1, 4), err: measure.ErrInvalidArgument},
		{name: "bar 0", grid: "4/4;pickup:1/4", edit: del(0, 1), err: measure.ErrInvalidArgument},
		{name: "КП-39 3…2", grid: "4/4", edit: del(3, 2), err: measure.ErrInvalidArgument},
		{name: "КП-41 4…5 of N = 4", grid: "4/4;bars:4", edit: del(4, 5), err: measure.ErrOutsideGrid},
		{name: "outside before all bars", grid: "4/4;bars:4", edit: del(1, 5), err: measure.ErrOutsideGrid},
		{name: "КП-40 pickup longer than the new bar 1", grid: "4/4;pickup:3/4;3:2/4", edit: del(1, 2), err: measure.ErrLongerThanMeter},
	})
}

// TestEditSlice covers О-7а, П-12, and КП-39, КП-41, КП-67: a slice is a
// finite grid of the bars of a range with their meters, lengths, and
// alignments.
func TestEditSlice(t *testing.T) {
	t.Parallel()

	slice := func(first, last int) func(measure.Grid) (measure.Grid, error) {
		return func(g measure.Grid) (measure.Grid, error) { return g.Slice(first, last) }
	}

	kp67 := "4/4;pickup:1/4;3:3/4;4:length:1/2"

	runEdits(t, []editCase{
		{name: "КП-67 2…4", grid: kp67, edit: slice(2, 4), want: "4/4;2:3/4;3:length:1/2;bars:3", end: "9/4"},
		{name: "КП-67 0…1", grid: kp67, edit: slice(0, 1), want: "4/4;pickup:1/4;bars:1", end: "5/4"},
		{name: "КП-67 3…5", grid: "4/4;3:3/4", edit: slice(3, 5), want: "3/4;bars:3", end: "9/4"},
		{name: "first bar to the end stays bar 1", grid: "4/4;3:length:1/4:end", edit: slice(3, 4), want: "4/4;1:length:1/4:end;bars:2", end: "5/4"},
		{name: "marks inside move", grid: "4/4;3:3/4;5:2/4;5:length:1/4", edit: slice(4, 6), want: "3/4;2:2/4;2:length:1/4;bars:3", end: "3/2"},
		{name: "repeated mark kept", grid: "4/4;3:4/4", edit: slice(2, 3), want: "4/4;2:4/4;bars:2"},
		{name: "finite whole", grid: "4/4;bars:4", edit: slice(1, 4), want: "4/4;bars:4"},
		{name: "КП-39 0…0", grid: "4/4;pickup:1/4", edit: slice(0, 0), err: measure.ErrInvalidArgument},
		{name: "negative", grid: "4/4", edit: slice(-1, 2), err: measure.ErrInvalidArgument},
		{name: "3…2", grid: "4/4", edit: slice(3, 2), err: measure.ErrInvalidArgument},
		{name: "0…0 without pickup", grid: "4/4", edit: slice(0, 0), err: measure.ErrInvalidArgument},
		{name: "bar 0 without pickup", grid: "4/4", edit: slice(0, 2), err: measure.ErrOutsideGrid},
		{name: "КП-41 3…5 of N = 4", grid: "4/4;bars:4", edit: slice(3, 5), err: measure.ErrOutsideGrid},
	})
}

// TestEditAppend covers О-7а, П-12, and КП-43, КП-68: the bars of the other
// grid follow bar N; its pickup becomes bar N + 1 with an explicit length
// aligned to the end.
func TestEditAppend(t *testing.T) {
	t.Parallel()

	// appendText builds the other grid in the goroutine of this test, while
	// the table is built, so that a failure of the build stops this test.
	appendText := func(other string) func(measure.Grid) (measure.Grid, error) {
		o := gridOf(t, other)

		return func(g measure.Grid) (measure.Grid, error) { return g.Append(o) }
	}

	runEdits(t, []editCase{
		{name: "КП-68 pickup", grid: "4/4;bars:2", edit: appendText("3/4;pickup:1/4;bars:2"), want: "4/4;3:3/4;3:length:1/4:end;bars:5", end: "15/4"},
		{name: "КП-68 open other", grid: "4/4;bars:2", edit: appendText("4/4"), want: "4/4;3:4/4"},
		{name: "no pickup", grid: "4/4;bars:2", edit: appendText("3/4;2:2/4;2:length:1/8;bars:3"), want: "4/4;3:3/4;4:2/4;4:length:1/8;bars:5", end: "27/8"},
		{name: "receiver records kept", grid: "4/4;pickup:1/8;2:3/4;2:length:1/2:end;bars:2", edit: appendText("6/8;bars:1"), want: "4/4;pickup:1/8;2:3/4;3:6/8;2:length:1/2:end;bars:3", end: "19/8"},
		{name: "КП-43 open receiver", grid: "4/4", edit: appendText("3/4"), err: measure.ErrUndefined},
	})

	t.Run("КП-68 place in the meter of (3, 0)", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		got, err := gridOf(t, "4/4;bars:2").Append(gridOf(t, "3/4;pickup:1/4;bars:2"))
		require.NoError(t, err)

		m, u, err := got.MetricPlace(ad(3, "0"))
		require.NoError(t, err)
		assertText(t, "3/4", m)
		assertText(t, "1/2", u)

		b, err := got.Bar(3)
		require.NoError(t, err)
		assert.Equal(t, barOf{3, "3/4", "1/4", "2", "9/4", measure.KindIncomplete, false, true, measure.AlignEnd, "1/2", true, true}.value(), b)

		b4, err := got.Bar(4)
		require.NoError(t, err)
		assert.False(t, b4.Marked, "the former bar 1 has no time signature")
	})

	t.Run("slice after append keeps bar 1", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		joined, err := gridOf(t, "4/4;bars:2").Append(gridOf(t, "3/4;pickup:1/4;bars:2"))
		require.NoError(t, err)

		s, err := joined.Slice(3, 5)
		require.NoError(t, err)
		assertRecord(t, "3/4;1:length:1/4:end;bars:3", s)

		_, ok, err := s.Pickup()
		require.NoError(t, err)
		assert.False(t, ok)
	})
}

// TestSimplify covers О-19 and КП-36: Simplify removes repeated time
// signatures and explicit lengths equal to the size of the meter of their
// bar, and keeps the pickup and the number of bars.
func TestSimplify(t *testing.T) {
	t.Parallel()

	simplify := func(g measure.Grid) (measure.Grid, error) { return g.Simplify() }

	runEdits(t, []editCase{
		{name: "КП-36", grid: "4/4;5:4/4;9:C;3:length:1/1;4:length:1/2", edit: simplify, want: "4/4;9:C;4:length:1/2"},
		{name: "length to the end equal to the meter", grid: "4/4;5:length:1/1:end", edit: simplify, want: "4/4"},
		{name: "length equal to a changed meter", grid: "4/4;3:3/4;4:length:3/4;5:length:1/1", edit: simplify, want: "4/4;3:3/4;5:length:1/1"},
		{name: "nothing to remove", grid: "4/4;pickup:1/4;3:3/4;4:length:1/2:end;bars:8", edit: simplify, want: "4/4;pickup:1/4;3:3/4;4:length:1/2:end;bars:8"},
		{name: "several repeats", grid: "3/4;2:3/4;3:3/4;4:2/4;5:2/4", edit: simplify, want: "3/4;4:2/4"},
		{name: "C after 4/4 stays", grid: "4/4;2:C;3:4/4", edit: simplify, want: "4/4;2:C;3:4/4"},
	})
}
