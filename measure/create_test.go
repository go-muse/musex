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

// Group Т-1: creation, the pickup, time signatures, explicit lengths, and
// the number of bars (architecture 4.4; О-1…О-3, О-5, О-6; П-2, П-4, П-5,
// П-8; КП-7 (the grid part), КП-8, КП-30, КП-40, КП-41, КП-53…КП-55,
// КП-60, КП-62, КП-63).

// TestCreateNew covers О-1 and ТВ-6: New gives an open grid without a
// pickup whose only record is the time signature of bar 1, kept as written.
func TestCreateNew(t *testing.T) {
	t.Parallel()

	for _, text := range []string{"4/4", "3/4", "6/8", "C", "C|", "2/4+3/8", "5/8", "7/8[3+2+2]", "(3+2+3)/8", "4/3", "1/4"} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g, err := measure.New(mt(text))
			require.NoError(t, err)

			assert.True(t, g.IsSet())
			assert.Equal(t, record{Marks: []string{"1:" + text}}, recordOf(t, g))

			marks, err := g.Marks()
			require.NoError(t, err)
			require.Len(t, marks, 1)
			assert.Equal(t, measure.Mark{Bar: 1, Meter: mt(text)}, marks[0])

			ok, err := g.HasPickup()
			require.NoError(t, err)
			assert.False(t, ok)

			changes, err := g.Changes()
			require.NoError(t, err)
			assert.Nil(t, changes)

			lengths, err := g.Lengths()
			require.NoError(t, err)
			assert.Nil(t, lengths)

			finite, err := g.IsFinite()
			require.NoError(t, err)
			assert.False(t, finite)

			n, finite, err := g.TotalBars()
			require.NoError(t, err)
			assert.Equal(t, 0, n)
			assert.False(t, finite)

			end, finite, err := g.End()
			require.NoError(t, err)
			assert.False(t, finite)
			assert.False(t, end.IsSet(), "an open grid has no end")

			first, err := g.FirstBar()
			require.NoError(t, err)
			assert.Equal(t, 1, first)
		})
	}
}

// TestCreateUnsetMeter covers КП-39: New with the unset meter fails with Е-2
// together with meter.ErrUnset and returns the unset grid.
func TestCreateUnsetMeter(t *testing.T) {
	t.Parallel()
	defer guard(t)

	g, err := measure.New(meter.Meter{})
	assertErr(t, err, measure.ErrInvalidArgument, meter.ErrUnset)
	assert.False(t, g.IsSet())
}

// TestPickupRead covers О-2 and КП-63: Pickup and HasPickup with and
// without a pickup, and FirstBar.
func TestPickupRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		grid   string
		pickup string // "" without a pickup
		first  int
	}{
		{"3/4 with pickup 1/4", "3/4;pickup:1/4", "1/4", 0},
		{"3/4 without pickup", "3/4", "", 1},
		{"4/4 pickup of a whole bar", "4/4;pickup:1/1", "1/1", 0},
		{"6/8 pickup 1/8", "6/8;pickup:1/8", "1/8", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			p, ok, err := g.Pickup()
			require.NoError(t, err)
			assert.Equal(t, tt.pickup != "", ok)

			if tt.pickup == "" {
				assert.False(t, p.IsSet(), "no pickup gives the unset duration")
			} else {
				assertText(t, tt.pickup, p)
			}

			has, err := g.HasPickup()
			require.NoError(t, err)
			assert.Equal(t, tt.pickup != "", has)

			first, err := g.FirstBar()
			require.NoError(t, err)
			assert.Equal(t, tt.first, first)

			hasZero, err := g.HasBar(0)
			require.NoError(t, err)
			assert.Equal(t, tt.pickup != "", hasZero)
		})
	}
}

// TestPickupSet covers О-2, ИС-3, П-8, КП-40, КП-51, and КП-63: a pickup
// from above zero up to the size of the meter of bar 1 inclusive; a
// replacement; a longer pickup fails with Е-3 and an unset length with Е-2.
func TestPickupSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		grid   string
		pickup duration.Duration
		want   string // the record, or "" for an error
		err    error
		causes []error
	}{
		{"3/4 pickup 1/4", "3/4", dur("1/4"), "3/4;pickup:1/4", nil, nil},
		{"replace 1/4 with 1/2", "3/4;pickup:1/4", dur("1/2"), "3/4;pickup:1/2", nil, nil},
		{"whole bar in 4/4", "4/4", dur("1"), "4/4;pickup:1/1", nil, nil},
		{"whole bar in 3/4", "3/4", dur("3/4"), "3/4;pickup:3/4", nil, nil},
		{"by the meter of bar 1", "4/4;3:2/4", dur("3/4"), "4/4;pickup:3/4;3:2/4", nil, nil},
		{"tiny pickup", "4/4", dur("1/1024"), "4/4;pickup:1/1024", nil, nil},
		{"finite grid keeps its count", "4/4;bars:4", dur("1/4"), "4/4;pickup:1/4;bars:4", nil, nil},
		{"КП-40 5/4 in 4/4", "4/4", dur("5/4"), "", measure.ErrLongerThanMeter, nil},
		{"КП-40 1 in 3/4", "3/4", dur("1"), "", measure.ErrLongerThanMeter, nil},
		{"replacement longer than the meter", "3/4;pickup:1/4", dur("7/8"), "", measure.ErrLongerThanMeter, nil},
		{"КП-39 unset length", "4/4", duration.Duration{}, "", measure.ErrInvalidArgument, []error{duration.ErrUnset}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			got, err := g.WithPickup(tt.pickup)

			if tt.err != nil {
				assertErr(t, err, tt.err, tt.causes...)
				assert.False(t, got.IsSet(), "the unset grid on error")
				assertRecord(t, tt.grid, g, "the receiver is unchanged")

				return
			}

			require.NoError(t, err)
			assertRecord(t, tt.want, got)
			assertRecord(t, tt.grid, g, "the receiver is unchanged")
		})
	}
}

// TestPickupRemove covers О-2, КП-60, and КП-63: WithoutPickup removes the
// pickup; without a pickup it returns the same value without an error.
func TestPickupRemove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid string
		want string
	}{
		{"КП-63 remove", "3/4;pickup:1/4", "3/4"},
		{"whole bar", "4/4;pickup:1/1;3:3/4;bars:4", "4/4;3:3/4;bars:4"},
		{"КП-60 no pickup", "3/4", "3/4"},
		{"КП-60 no pickup, finite", "4/4;5:3/4;bars:6", "4/4;5:3/4;bars:6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			got, err := g.WithoutPickup()
			require.NoError(t, err)
			assertRecord(t, tt.want, got)
			assertRecord(t, tt.grid, g, "the receiver is unchanged")

			_, ok, err := got.Pickup()
			require.NoError(t, err)
			assert.False(t, ok)
		})
	}
}

// TestMarksRead covers О-3, П-4, and КП-30: Marks keeps repeated time
// signatures; Changes and ChangesAt report only changes by record.
func TestMarksRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		grid      string
		changes   []string
		changesAt map[int]bool
		hasMark   map[int]bool
	}{
		{
			"КП-30 repeated 4/4", "4/4;5:4/4", nil,
			map[int]bool{1: false, 2: false, 5: false, 6: false},
			map[int]bool{1: true, 2: false, 5: true, 6: false},
		},
		{
			"КП-30 4/4 then C", "4/4;5:C", []string{"5:C"},
			map[int]bool{1: false, 4: false, 5: true, 6: false},
			map[int]bool{1: true, 5: true},
		},
		{
			"changes and a repeat", "4/4;3:3/4;5:3/4;7:4/4",
			[]string{"3:3/4", "7:4/4"},
			map[int]bool{2: false, 3: true, 4: false, 5: false, 7: true, 8: false},
			map[int]bool{3: true, 4: false, 5: true, 7: true, 8: false},
		},
		{
			"pickup bar 0", "3/4;pickup:1/4;2:2/4",
			[]string{"2:2/4"},
			map[int]bool{0: false, 1: false, 2: true},
			map[int]bool{0: false, 1: true, 2: true},
		},
		{
			"C and ¢ differ", "C;2:C|;3:C|", []string{"2:C|"},
			map[int]bool{2: true, 3: false},
			map[int]bool{2: true, 3: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			assertRecord(t, tt.grid, g)

			changes, err := g.Changes()
			require.NoError(t, err)

			var texts []string
			for _, c := range changes {
				texts = append(texts, c.String())
			}

			assert.Equal(t, tt.changes, texts)

			if tt.changes == nil {
				assert.Nil(t, changes, "no changes give nil")
			}

			for bar, want := range tt.changesAt {
				got, err := g.ChangesAt(bar)
				require.NoError(t, err)
				assert.Equal(t, want, got, "ChangesAt(%d)", bar)
			}

			for bar, want := range tt.hasMark {
				got, err := g.HasMark(bar)
				require.NoError(t, err)
				assert.Equal(t, want, got, "HasMark(%d)", bar)
			}
		})
	}
}

// TestMarksReadErrors covers architecture 4.2 for HasMark and ChangesAt:
// Е-2 for a negative number, Е-4 for a missing bar.
func TestMarksReadErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid string
		bar  int
		err  error
	}{
		{"negative", "4/4", -1, measure.ErrInvalidArgument},
		{"bar 0 without pickup", "4/4", 0, measure.ErrOutsideGrid},
		{"after N", "4/4;bars:4", 5, measure.ErrOutsideGrid},
		{"min int", "4/4", math.MinInt, measure.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			has, err := g.HasMark(tt.bar)
			assertErr(t, err, tt.err)
			assert.False(t, has)

			changes, err := g.ChangesAt(tt.bar)
			assertErr(t, err, tt.err)
			assert.False(t, changes)
		})
	}
}

// TestMarksWith covers О-3, ИС-12, П-4, П-8, КП-39, КП-40, and КП-41:
// WithMark sets or replaces a time signature from bar k ≥ 1; a repeated one
// is kept; errors in the order of architecture 4.2.
func TestMarksWith(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		grid   string
		bar    int
		meter  meter.Meter
		want   string
		err    error
		causes []error
	}{
		{"set at 3", "4/4", 3, mt("3/4"), "4/4;3:3/4", nil, nil},
		{"replace at 3", "4/4;3:3/4", 3, mt("2/4"), "4/4;3:2/4", nil, nil},
		{"replace bar 1", "4/4;3:3/4", 1, mt("C"), "C;3:3/4", nil, nil},
		{"repeated is kept", "4/4", 5, mt("4/4"), "4/4;5:4/4", nil, nil},
		{"between marks", "4/4;5:3/4", 3, mt("2/4"), "4/4;3:2/4;5:3/4", nil, nil},
		{"bar 1 with a pickup that fits", "4/4;pickup:1/4", 1, mt("2/4"), "2/4;pickup:1/4", nil, nil},
		{"finite at N", "4/4;bars:4", 4, mt("3/4"), "4/4;4:3/4;bars:4", nil, nil},
		{"max int", "4/4", math.MaxInt, mt("3/4"), "4/4;9223372036854775807:3/4", nil, nil},
		{"lengths kept", "4/4;4:length:1/2", 3, mt("3/4"), "4/4;3:3/4;4:length:1/2", nil, nil},
		{"КП-39 bar 0", "4/4", 0, mt("3/4"), "", measure.ErrInvalidArgument, nil},
		{"КП-39 bar −1", "4/4", -1, mt("3/4"), "", measure.ErrInvalidArgument, nil},
		{"КП-39 unset meter", "4/4", 3, meter.Meter{}, "", measure.ErrInvalidArgument, []error{meter.ErrUnset}},
		{"КП-41 after N", "4/4;bars:4", 5, mt("3/4"), "", measure.ErrOutsideGrid, nil},
		{"КП-40 pickup 3/4, bar 1 to 2/4", "4/4;pickup:3/4", 1, mt("2/4"), "", measure.ErrLongerThanMeter, nil},
		{"КП-40 bar 5 length 1 to the end, from bar 4 3/4", "4/4;5:length:1/1:end", 4, mt("3/4"), "", measure.ErrLongerThanMeter, nil},
		{"aligned to the end in a later bar", "4/4;6:length:7/8:end", 2, mt("3/4"), "", measure.ErrLongerThanMeter, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			got, err := g.WithMark(tt.bar, tt.meter)

			if tt.err != nil {
				assertErr(t, err, tt.err, tt.causes...)
				assert.False(t, got.IsSet())
				assertRecord(t, tt.grid, g, "the receiver is unchanged")

				return
			}

			require.NoError(t, err)
			assertRecord(t, tt.want, got)
			assertRecord(t, tt.grid, g, "the receiver is unchanged")
		})
	}
}

// TestMarksWithout covers О-3, ИС-12, КП-39, КП-40, КП-60, and КП-63:
// WithoutMark removes a time signature of bar k ≥ 2 and keeps the others.
func TestMarksWithout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid string
		bar  int
		want string
		err  error
	}{
		{"КП-63 remove bar 3", "4/4;3:3/4;5:2/4", 3, "4/4;5:2/4", nil},
		{"remove the repeat", "4/4;5:4/4", 5, "4/4", nil},
		{"keeps the later mark", "4/4;5:3/4;9:4/4", 5, "4/4;9:4/4", nil},
		{"КП-60 bar without a mark", "4/4;3:3/4", 4, "4/4;3:3/4", nil},
		{"КП-60 after N", "4/4;bars:4", 6, "4/4;bars:4", nil},
		{"max int without a mark", "4/4", math.MaxInt, "4/4", nil},
		{"pickup that still fits", "4/4;pickup:1/4;2:2/4", 2, "4/4;pickup:1/4", nil},
		{"КП-39 bar 1", "4/4;3:3/4", 1, "", measure.ErrInvalidArgument},
		{"bar 0", "4/4;pickup:1/4", 0, "", measure.ErrInvalidArgument},
		{"negative", "4/4", -1, "", measure.ErrInvalidArgument},
		{"КП-40 to the end, 5/4", "4/4;3:5/4;4:length:5/4:end", 3, "", measure.ErrLongerThanMeter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			got, err := g.WithoutMark(tt.bar)

			if tt.err != nil {
				assertErr(t, err, tt.err)
				assert.False(t, got.IsSet())
				assertRecord(t, tt.grid, g, "the receiver is unchanged")

				return
			}

			require.NoError(t, err)
			assertRecord(t, tt.want, got)
			assertRecord(t, tt.grid, g, "the receiver is unchanged")
		})
	}
}

// TestLengthsWith covers О-5, ИС-4, П-5, П-6, КП-39…КП-41, КП-54, and
// КП-63: WithLength sets or replaces an explicit length of any size; one
// aligned to the end is at most the size of the meter of its bar.
func TestLengthsWith(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		grid   string
		bar    int
		length duration.Duration
		align  measure.Alignment
		want   string
		err    error
		causes []error
	}{
		{"shorter", "4/4", 3, dur("1/2"), measure.AlignStart, "4/4;3:length:1/2", nil, nil},
		{"longer", "4/4", 2, dur("3/2"), measure.AlignStart, "4/4;2:length:3/2", nil, nil},
		{"КП-54 equal to the meter is kept", "4/4", 3, dur("1"), measure.AlignStart, "4/4;3:length:1/1", nil, nil},
		{"to the end", "4/4", 5, dur("1/4"), measure.AlignEnd, "4/4;5:length:1/4:end", nil, nil},
		{"to the end, whole bar", "4/4", 5, dur("1"), measure.AlignEnd, "4/4;5:length:1/1:end", nil, nil},
		{"to the end by the meter of the bar", "4/4;3:5/4", 4, dur("5/4"), measure.AlignEnd, "4/4;3:5/4;4:length:5/4:end", nil, nil},
		{"bar 1", "3/4", 1, dur("1/4"), measure.AlignStart, "3/4;1:length:1/4", nil, nil},
		{"КП-63 replace", "4/4;3:length:1/2", 3, dur("3/2"), measure.AlignStart, "4/4;3:length:3/2", nil, nil},
		{"replace the alignment", "4/4;3:length:1/2", 3, dur("1/2"), measure.AlignEnd, "4/4;3:length:1/2:end", nil, nil},
		{"order of bars", "4/4;5:length:1/4", 2, dur("1/8"), measure.AlignStart, "4/4;2:length:1/8;5:length:1/4", nil, nil},
		{"finite at N", "4/4;bars:4", 4, dur("1/2"), measure.AlignStart, "4/4;4:length:1/2;bars:4", nil, nil},
		{"КП-40 to the end longer than the meter", "4/4", 5, dur("5/4"), measure.AlignEnd, "", measure.ErrLongerThanMeter, nil},
		{"КП-39 bar 0", "4/4", 0, dur("1/2"), measure.AlignStart, "", measure.ErrInvalidArgument, nil},
		{"КП-39 bar −1", "4/4", -1, dur("1/2"), measure.AlignStart, "", measure.ErrInvalidArgument, nil},
		{"КП-39 unset length", "4/4", 3, duration.Duration{}, measure.AlignStart, "", measure.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"alignment 0", "4/4", 3, dur("1/2"), measure.Alignment(0), "", measure.ErrInvalidArgument, nil},
		{"alignment 3", "4/4", 3, dur("1/2"), measure.Alignment(3), "", measure.ErrInvalidArgument, nil},
		{"КП-41 after N", "4/4;bars:4", 6, dur("1/2"), measure.AlignStart, "", measure.ErrOutsideGrid, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			got, err := g.WithLength(tt.bar, tt.length, tt.align)

			if tt.err != nil {
				assertErr(t, err, tt.err, tt.causes...)
				assert.False(t, got.IsSet())
				assertRecord(t, tt.grid, g, "the receiver is unchanged")

				return
			}

			require.NoError(t, err)
			assertRecord(t, tt.want, got)
			assertRecord(t, tt.grid, g, "the receiver is unchanged")
		})
	}
}

// TestLengthsRead covers О-5: Lengths lists the explicit lengths in order of
// bars with their alignment.
func TestLengthsRead(t *testing.T) {
	t.Parallel()
	defer guard(t)

	g := gridOf(t, "4/4;2:length:1/8;5:length:1/4:end;9:length:5/2")

	lengths, err := g.Lengths()
	require.NoError(t, err)
	assert.Equal(t, []measure.ExplicitLength{
		{Bar: 2, Duration: dur("1/8"), Alignment: measure.AlignStart},
		{Bar: 5, Duration: dur("1/4"), Alignment: measure.AlignEnd},
		{Bar: 9, Duration: dur("5/2"), Alignment: measure.AlignStart},
	}, lengths)
}

// TestLengthsWithout covers О-5, КП-39, КП-60, and КП-63: WithoutLength
// removes an explicit length; without one it returns the same value.
func TestLengthsWithout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid string
		bar  int
		want string
		err  error
	}{
		{"КП-63 remove", "4/4;3:length:1/2", 3, "4/4", nil},
		{"keeps the others", "4/4;3:length:1/2;5:length:1/4:end", 5, "4/4;3:length:1/2", nil},
		{"КП-60 bar without a length", "4/4;3:length:1/2", 4, "4/4;3:length:1/2", nil},
		{"КП-60 after N", "4/4;bars:4", 6, "4/4;bars:4", nil},
		{"bar 0", "4/4", 0, "", measure.ErrInvalidArgument},
		{"negative", "4/4", -1, "", measure.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			got, err := g.WithoutLength(tt.bar)

			if tt.err != nil {
				assertErr(t, err, tt.err)
				assert.False(t, got.IsSet())
				assertRecord(t, tt.grid, g, "the receiver is unchanged")

				return
			}

			require.NoError(t, err)
			assertRecord(t, tt.want, got)
			assertRecord(t, tt.grid, g, "the receiver is unchanged")
		})
	}
}

// TestFiniteRead covers О-6, П-2, КП-8, and КП-53: BarCount, IsFinite,
// TotalBars, and End of open and finite grids.
func TestFiniteRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		grid   string
		count  int // 0 for an open grid
		total  int
		end    string
		finite bool
	}{
		{"open", "4/4;3:3/4", 0, 0, "", false},
		{"open with pickup", "3/4;pickup:1/4", 0, 0, "", false},
		{"КП-8 4/4, N = 4", "4/4;bars:4", 4, 4, "4", true},
		{"КП-8 pickup and explicit length", "4/4;pickup:1/4;8:length:3/4;bars:8", 8, 9, "8", true},
		{"КП-53 one bar", "4/4;bars:1", 1, 1, "1", true},
		{"changes", "4/4;3:3/4;bars:5", 5, 5, "17/4", true},
		{"extended and to the end", "4/4;2:length:3/2;3:length:1/4:end;bars:3", 3, 3, "11/4", true},
		{"max int", "4/4;bars:9223372036854775807", math.MaxInt, math.MaxInt, "9223372036854775807", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			n, finite, err := g.BarCount()
			require.NoError(t, err)
			assert.Equal(t, tt.count, n)
			assert.Equal(t, tt.finite, finite)

			isFinite, err := g.IsFinite()
			require.NoError(t, err)
			assert.Equal(t, tt.finite, isFinite)

			total, finite, err := g.TotalBars()
			require.NoError(t, err)
			assert.Equal(t, tt.total, total)
			assert.Equal(t, tt.finite, finite)

			end, finite, err := g.End()
			require.NoError(t, err)
			assert.Equal(t, tt.finite, finite)

			if tt.finite {
				assertText(t, canonicalRat(tt.end), end)
			} else {
				assert.False(t, end.IsSet())
			}
		})
	}
}

// TestFiniteHasBar covers 2.2, П-2, and П-3: which bars exist in open and
// finite grids, with and without a pickup.
func TestFiniteHasBar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid string
		bars map[int]bool
	}{
		{"open", "4/4", map[int]bool{0: false, 1: true, 2: true, 1 << 40: true, math.MaxInt: true}},
		{"open with pickup", "4/4;pickup:1/4", map[int]bool{0: true, 1: true, math.MaxInt: true}},
		{"finite", "4/4;bars:4", map[int]bool{0: false, 1: true, 4: true, 5: false, math.MaxInt: false}},
		{"finite with pickup", "4/4;pickup:1/4;bars:2", map[int]bool{0: true, 2: true, 3: false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			for bar, want := range tt.bars {
				got, err := g.HasBar(bar)
				require.NoError(t, err)
				assert.Equal(t, want, got, "HasBar(%d)", bar)
			}

			has, err := g.HasBar(-1)
			assertErr(t, err, measure.ErrInvalidArgument)
			assert.False(t, has)
		})
	}
}

// TestFiniteBarCount covers О-6, П-2, КП-39, and КП-41: WithBarCount makes a
// grid finite; a record after the new count fails with Е-4.
func TestFiniteBarCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid string
		n    int
		want string
		err  error
	}{
		{"open to 8", "4/4;3:3/4", 8, "4/4;3:3/4;bars:8", nil},
		{"decrease", "4/4;bars:8", 4, "4/4;bars:4", nil},
		{"increase", "4/4;bars:4", 6, "4/4;bars:6", nil},
		{"at the last record", "4/4;6:3/4;7:length:1/2", 7, "4/4;6:3/4;7:length:1/2;bars:7", nil},
		{"one bar", "4/4", 1, "4/4;bars:1", nil},
		{"with pickup", "4/4;pickup:1/4", 2, "4/4;pickup:1/4;bars:2", nil},
		{"КП-39 N = 0", "4/4", 0, "", measure.ErrInvalidArgument},
		{"negative", "4/4", -3, "", measure.ErrInvalidArgument},
		{"КП-41 mark after N", "4/4;6:3/4", 4, "", measure.ErrOutsideGrid},
		{"length after N", "4/4;7:length:1/2", 6, "", measure.ErrOutsideGrid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			got, err := g.WithBarCount(tt.n)

			if tt.err != nil {
				assertErr(t, err, tt.err)
				assert.False(t, got.IsSet())
				assertRecord(t, tt.grid, g, "the receiver is unchanged")

				return
			}

			require.NoError(t, err)
			assertRecord(t, tt.want, got)
			assertRecord(t, tt.grid, g, "the receiver is unchanged")
		})
	}
}

// TestFiniteTruncate covers О-6, КП-39, КП-60, and КП-62: Truncate drops the
// records after n; a finite grid with N ≤ n is unchanged.
func TestFiniteTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid string
		n    int
		want string
		err  error
	}{
		{"КП-62", "4/4;6:3/4;7:length:1/2", 4, "4/4;bars:4", nil},
		{"open keeps earlier records", "4/4;3:3/4;3:length:1/2;6:2/4", 5, "4/4;3:3/4;3:length:1/2;bars:5", nil},
		{"finite decrease", "4/4;3:3/4;bars:8", 2, "4/4;bars:2", nil},
		{"КП-60 N = 4 to 6", "4/4;bars:4", 6, "4/4;bars:4", nil},
		{"N equal", "4/4;bars:4", 4, "4/4;bars:4", nil},
		{"keeps the pickup", "4/4;pickup:1/4;5:3/4", 1, "4/4;pickup:1/4;bars:1", nil},
		{"0", "4/4", 0, "", measure.ErrInvalidArgument},
		{"negative", "4/4;bars:4", -1, "", measure.ErrInvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			got, err := g.Truncate(tt.n)

			if tt.err != nil {
				assertErr(t, err, tt.err)
				assert.False(t, got.IsSet())
				assertRecord(t, tt.grid, g, "the receiver is unchanged")

				return
			}

			require.NoError(t, err)
			assertRecord(t, tt.want, got)
			assertRecord(t, tt.grid, g, "the receiver is unchanged")
		})
	}
}

// TestFiniteOpen covers О-6, КП-60, and КП-62: WithoutBarCount makes a grid
// open; an open grid is unchanged.
func TestFiniteOpen(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct{ grid, want string }{
		{"4/4;bars:4", "4/4"},
		{"4/4;pickup:1/4;3:3/4;4:length:1/2;bars:8", "4/4;pickup:1/4;3:3/4;4:length:1/2"},
		{"4/4", "4/4"},
	} {
		t.Run(tt.grid, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			got, err := g.WithoutBarCount()
			require.NoError(t, err)
			assertRecord(t, tt.want, got)
			assertRecord(t, tt.grid, g, "the receiver is unchanged")
		})
	}

	t.Run("КП-62 truncate then open", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4;6:3/4;7:length:1/2")
		cut, err := g.Truncate(4)
		require.NoError(t, err)

		open, err := cut.WithoutBarCount()
		require.NoError(t, err)
		assertRecord(t, "4/4", open)
	})
}

// TestFiniteAddBars covers О-6, КП-8, КП-39, and КП-43: AddBars appends bars
// to a finite grid with the meter of bar N; an open grid fails with Е-6.
func TestFiniteAddBars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		grid string
		m    int
		want string
		end  string
		err  error
	}{
		{"КП-8 add 2", "4/4;bars:4", 2, "4/4;bars:6", "6", nil},
		{"meter of bar N", "4/4;3:3/4;bars:4", 2, "4/4;3:3/4;bars:6", "5", nil},
		{"after an explicit length", "4/4;4:length:1/2;bars:4", 1, "4/4;4:length:1/2;bars:5", "9/2", nil},
		{"КП-39 add 0", "4/4;bars:4", 0, "", "", measure.ErrInvalidArgument},
		{"negative", "4/4;bars:4", -2, "", "", measure.ErrInvalidArgument},
		{"КП-43 open", "4/4", 2, "", "", measure.ErrUndefined},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			got, err := g.AddBars(tt.m)

			if tt.err != nil {
				assertErr(t, err, tt.err)
				assert.False(t, got.IsSet())
				assertRecord(t, tt.grid, g, "the receiver is unchanged")

				return
			}

			require.NoError(t, err)
			assertRecord(t, tt.want, got)

			end, finite, err := got.End()
			require.NoError(t, err)
			assert.True(t, finite)
			assertText(t, canonicalRat(tt.end), end)
		})
	}
}
