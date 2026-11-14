package measure_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-4: ranges of bars, bars spanned by a range of positions, and
// parts of the length of a bar (architecture 4.8; О-11…О-13; КП-13…КП-16,
// КП-30, КП-41, КП-42).

// TestRangeLength covers О-11 and КП-13: the length of a range of bars is
// the end of its last bar minus the start of its first bar.
func TestRangeLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		grid        string
		first, last int
		want        string
	}{
		{"КП-13 4/4 1…3", "4/4", 1, 3, "3"},
		{"КП-13 change 1…3", "4/4;3:3/4", 1, 3, "11/4"},
		{"КП-13 change 3…5", "4/4;3:3/4", 3, 5, "9/4"},
		{"one bar", "4/4;3:3/4", 3, 3, "3/4"},
		{"with pickup", "3/4;pickup:1/4", 0, 2, "7/4"},
		{"pickup only", "3/4;pickup:1/4", 0, 0, "1/4"},
		{"explicit lengths", "4/4;2:length:3/2;3:length:1/4:end", 1, 4, "15/4"},
		{"finite up to N", "4/4;bars:4", 1, 4, "4"},
		{"far", "4/4", 1 << 40, 1<<40 + 9, "10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			l, err := g.RangeLength(tt.first, tt.last)
			require.NoError(t, err)
			assertText(t, canonicalRat(tt.want), l)

			start, err := g.BarStart(tt.first)
			require.NoError(t, err)

			end, err := g.BarEnd(tt.last)
			require.NoError(t, err)
			assert.Equal(t, start.MustDurationTo(end), l, "КП-61: the end minus the start")
		})
	}

	t.Run("КП-13 start and end of 3…5", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "4/4;3:3/4")

		start, err := g.BarStart(3)
		require.NoError(t, err)
		assertText(t, "2/1", start)

		end, err := g.BarEnd(5)
		require.NoError(t, err)
		assertText(t, "17/4", end)
	})
}

// TestRangeChanges covers О-11, П-4, КП-13, and КП-30: a change of the
// meter from bar k counts only for first < k ≤ last; a repeated time
// signature is not a change.
func TestRangeChanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		grid        string
		first, last int
		want        bool
	}{
		{"КП-13 1…2", "4/4;3:3/4", 1, 2, false},
		{"КП-13 1…3", "4/4;3:3/4", 1, 3, true},
		{"КП-13 3…5", "4/4;3:3/4", 3, 5, false},
		{"КП-30 repeated 1…8", "4/4;5:4/4", 1, 8, false},
		{"КП-30 C 1…8", "4/4;5:C", 1, 8, true},
		{"single bar at the change", "4/4;3:3/4", 3, 3, false},
		{"pickup 0…1", "4/4;pickup:1/4", 0, 1, false},
		{"pickup 0…2 with change at 2", "4/4;pickup:1/4;2:3/4", 0, 2, true},
		{"explicit length is not a change", "4/4;3:length:1/2", 1, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := gridOf(t, tt.grid).ChangesWithin(tt.first, tt.last)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestRangeAllFull covers О-11 and КП-14: whether all bars of a range are
// full.
func TestRangeAllFull(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		grid        string
		first, last int
		want        bool
	}{
		{"КП-14 1…2", "4/4;3:length:1/2", 1, 2, true},
		{"КП-14 1…3", "4/4;3:length:1/2", 1, 3, false},
		{"КП-14 pickup 0…1", "3/4;pickup:1/4", 0, 1, false},
		{"КП-14 1…2 after pickup", "3/4;pickup:1/4", 1, 2, true},
		{"whole-bar pickup", "4/4;pickup:1/1", 0, 2, true},
		{"explicit equal to the meter", "4/4;3:length:1/1", 1, 5, true},
		{"extended", "4/4;2:length:3/2", 1, 3, false},
		{"after the incomplete bar", "4/4;3:length:1/2", 4, 9, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := gridOf(t, tt.grid).AllFull(tt.first, tt.last)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestRangeMeterRuns covers О-11 and КП-13: the runs without a change of
// the meter cover the range in order.
func TestRangeMeterRuns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		grid        string
		first, last int
		want        []string
	}{
		{"КП-13 1…5", "4/4;3:3/4", 1, 5, []string{"1-2:4/4", "3-5:3/4"}},
		{"inside one run", "4/4;3:3/4", 3, 5, []string{"3-5:3/4"}},
		{"repeated mark is no change", "4/4;5:4/4", 1, 8, []string{"1-8:4/4"}},
		{"pickup", "4/4;pickup:1/4;3:3/4", 0, 4, []string{"0-2:4/4", "3-4:3/4"}},
		{"several", "4/4;2:3/4;4:C;5:4/4", 1, 6, []string{"1-1:4/4", "2-3:3/4", "4-4:C", "5-6:4/4"}},
		{"one bar", "4/4", 7, 7, []string{"7-7:4/4"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			runs, err := gridOf(t, tt.grid).MeterRuns(tt.first, tt.last)
			require.NoError(t, err)

			texts := make([]string, len(runs))
			for i, r := range runs {
				texts[i] = r.String()
			}

			assert.Equal(t, tt.want, texts)
		})
	}

	t.Run("fields", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		runs, err := gridOf(t, "4/4;3:3/4").MeterRuns(1, 5)
		require.NoError(t, err)
		assert.Equal(t, []measure.MeterRun{
			{First: 1, Last: 2, Meter: mt("4/4")},
			{First: 3, Last: 5, Meter: mt("3/4")},
		}, runs)
	})
}

// TestRangeErrors covers architecture 4.2 for the range operations and
// КП-39, КП-41: Е-2 for first < 0, then for first > last; Е-4 for a missing
// first bar, then for a missing last bar.
func TestRangeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		grid        string
		first, last int
		err         error
	}{
		{"first < 0", "4/4", -1, 3, measure.ErrInvalidArgument},
		{"КП-39 3…2", "4/4", 3, 2, measure.ErrInvalidArgument},
		{"first > last with missing bars", "4/4;bars:4", 7, 6, measure.ErrInvalidArgument},
		{"bar 0 without pickup", "4/4", 0, 2, measure.ErrOutsideGrid},
		{"КП-41 3…5 of N = 4", "4/4;bars:4", 3, 5, measure.ErrOutsideGrid},
		{"first after N", "4/4;bars:4", 5, 5, measure.ErrOutsideGrid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			l, err := g.RangeLength(tt.first, tt.last)
			assertErr(t, err, tt.err)
			assert.False(t, l.IsSet())

			c, err := g.ChangesWithin(tt.first, tt.last)
			assertErr(t, err, tt.err)
			assert.False(t, c)

			f, err := g.AllFull(tt.first, tt.last)
			assertErr(t, err, tt.err)
			assert.False(t, f)

			runs, err := g.MeterRuns(tt.first, tt.last)
			assertErr(t, err, tt.err)
			assert.Nil(t, runs)
		})
	}
}

// TestRangeBarsSpanned covers О-12 and КП-15: the first and last bars that
// a range of positions takes; an end on a bar line does not take the next
// bar.
func TestRangeBarsSpanned(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		grid        string
		start       string
		value       string
		first, last int
	}{
		{"КП-15 3/4 and 1", "4/4", "3/4", "1", 1, 2},
		{"КП-15 0 and 1", "4/4", "0", "1", 1, 1},
		{"КП-15 0 and 5/4", "4/4", "0", "5/4", 1, 2},
		{"КП-15 pickup 0 and 1/4", "3/4;pickup:1/4", "0", "1/4", 0, 0},
		{"КП-15 pickup 0 and 1", "3/4;pickup:1/4", "0", "1", 0, 1},
		{"to the end of a finite grid", "4/4;bars:4", "3", "1", 4, 4},
		{"across a change", "4/4;3:3/4", "3/2", "2", 2, 4},
		{"tiny inside", "4/4", "1/2", "1/1024", 1, 1},
		{"far", "4/4", "1099511627776", "3", 1<<40 + 1, 1<<40 + 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			first, last, err := gridOf(t, tt.grid).BarsSpanned(pos(tt.start), dur(tt.value))
			require.NoError(t, err)
			assert.Equal(t, tt.first, first)
			assert.Equal(t, tt.last, last)
		})
	}
}

// TestRangeBarsSpannedErrors covers architecture 4.2 and КП-39, КП-41 for
// BarsSpanned: Е-2 for an unset start, then for an unset value; Е-4 for a
// range after the end of a finite grid.
func TestRangeBarsSpannedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		grid    string
		start   duration.Position
		value   duration.Duration
		err     error
		causes  []error
		without []error
	}{
		{"unset start", "4/4", duration.Position{}, dur("1"), measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"КП-39 unset value", "4/4", pos("0"), duration.Duration{}, measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"КП-41 3 and 2 of N = 4", "4/4;bars:4", pos("3"), dur("2"), measure.ErrOutsideGrid, nil, nil},
		{"start at the end", "4/4;bars:4", pos("4"), dur("1/4"), measure.ErrOutsideGrid, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			first, last, err := gridOf(t, tt.grid).BarsSpanned(tt.start, tt.value)
			assertErr(t, err, tt.err, tt.causes...)
			assertNoCause(t, err, tt.without...)
			assert.Zero(t, first)
			assert.Zero(t, last)
		})
	}
}

// TestElapsed covers О-13 and КП-16, КП-42: the part of the length of a bar
// already passed, from 0 up to but not including 1.
func TestElapsed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		grid    string
		address measure.Address
		want    string
		err     error
	}{
		{"КП-16 4/4 (1, 1/4)", "4/4", ad(1, "1/4"), "1/4", nil},
		{"КП-16 3/4 (2, 1/2)", "3/4", ad(2, "1/2"), "2/3", nil},
		{"start", "4/4", ad(3, "0"), "0", nil},
		{"pickup", "3/4;pickup:1/4", ad(0, "1/8"), "1/2", nil},
		{"explicit length", "4/4;3:length:1/2", ad(3, "3/8"), "3/4", nil},
		{"extended", "4/4;2:length:3/2", ad(2, "1"), "2/3", nil},
		{"КП-42 4/4 (1, 1)", "4/4", ad(1, "1"), "", measure.ErrOutsideBar},
		{"bar 0 without pickup", "4/4", ad(0, "0"), "", measure.ErrOutsideGrid},
		{"after N", "4/4;bars:4", ad(5, "0"), "", measure.ErrOutsideGrid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			f, err := gridOf(t, tt.grid).Elapsed(tt.address)

			if tt.err != nil {
				assertErr(t, err, tt.err)
				assert.False(t, f.IsValid())

				return
			}

			require.NoError(t, err)
			assertFrac(t, tt.want, f)
		})
	}
}

// TestElapsedPartOfBar covers О-13 and КП-16: the ratio of a value to the
// length of a bar, greater than 0 and possibly 1 or more.
func TestElapsedPartOfBar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		grid   string
		bar    int
		value  duration.Duration
		want   string
		err    error
		causes []error
	}{
		{"КП-16 3/4 1/4", "3/4", 1, dur("1/4"), "1/3", nil, nil},
		{"КП-16 3/4 1", "3/4", 2, dur("1"), "4/3", nil, nil},
		{"КП-16 explicit 1/2", "4/4;3:length:1/2", 3, dur("1/2"), "1", nil, nil},
		{"pickup", "4/4;pickup:1/4", 0, dur("1/8"), "1/2", nil, nil},
		{"negative bar", "4/4", -1, dur("1/4"), "", measure.ErrInvalidArgument, nil},
		{"unset value", "4/4", 1, duration.Duration{}, "", measure.ErrInvalidArgument, []error{duration.ErrUnset}},
		{"bar 0 without pickup", "4/4", 0, dur("1/4"), "", measure.ErrOutsideGrid, nil},
		{"after N", "4/4;bars:2", 3, dur("1/4"), "", measure.ErrOutsideGrid, nil},
		{"min int", "4/4", math.MinInt, duration.Duration{}, "", measure.ErrInvalidArgument, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			f, err := gridOf(t, tt.grid).PartOfBar(tt.bar, tt.value)

			if tt.err != nil {
				assertErr(t, err, tt.err, tt.causes...)
				assert.False(t, f.IsValid())

				return
			}

			require.NoError(t, err)
			assertFrac(t, tt.want, f)
		})
	}

	t.Run("negative bar before unset value", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		_, err := gridOf(t, "4/4").PartOfBar(-1, duration.Duration{})
		assertErr(t, err, measure.ErrInvalidArgument)
		assertNoCause(t, err, duration.ErrUnset)
	})
}
