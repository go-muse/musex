package measure_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Group Т-6: splitting a duration (architecture 4.11; О-17; ИС-14; П-10;
// КП-26…КП-29, КП-41, КП-44, КП-45, КП-52, КП-53, КП-65).

// assertParts checks the properties of a split of КП-61 and architecture
// 4.11: the parts follow each other from start without gaps, their sum is
// value, each part lies in one bar, and the position of its address is its
// start.
func assertParts(t *testing.T, g measure.Grid, start duration.Position, value duration.Duration, parts []measure.Part) {
	t.Helper()

	if !assert.NotEmpty(t, parts) {
		return
	}

	at := start
	sum := new(big.Rat)

	for i, p := range parts {
		assert.Equal(t, at, p.Start, "part %d starts where the previous ends", i)

		pp, err := g.PositionOf(p.Address)
		if assert.NoError(t, err, "part %d", i) {
			assert.Equal(t, p.Start, pp, "part %d: the position of the address is the start", i)
		}

		end, err := p.End()
		require.NoError(t, err)

		first, last, err := g.BarsSpanned(p.Start, p.Duration)
		if assert.NoError(t, err, "part %d", i) {
			assert.Equal(t, first, last, "part %d lies in one bar", i)
		}

		r, ok := p.Duration.MustFraction().BigRat()
		require.True(t, ok)
		sum.Add(sum, r)

		at = end
	}

	vr, ok := value.MustFraction().BigRat()
	require.True(t, ok)
	assert.Equal(t, 0, sum.Cmp(vr), "the sum of the parts is the value: %s, want %s", sum, vr)
}

// splitCase is a split by a level and its parts as "duration@bar:offset".
type splitCase struct {
	name  string
	grid  string
	start string
	value string
	level measure.Level
	want  []string
}

// TestSplitBars covers О-17, ИС-14, and КП-26, КП-27, КП-29, КП-52, КП-53,
// КП-65: a split at the bar lines; Split equals SplitBy with LevelBars.
func TestSplitBars(t *testing.T) {
	t.Parallel()

	tests := []splitCase{
		{"КП-26 3/4 1/2 and 1/2", "3/4", "1/2", "1/2", measure.LevelBars, []string{"1/4@1:1/2", "1/4@2:0/1"}},
		{"КП-26 4/4 3/4 and 1", "4/4", "3/4", "1", measure.LevelBars, []string{"1/4@1:3/4", "3/4@2:0/1"}},
		{"КП-26 4/4 0 and 9/4", "4/4", "0", "9/4", measure.LevelBars, []string{"1/1@1:0/1", "1/1@2:0/1", "1/4@3:0/1"}},
		{"КП-26 4/4 3/4 and 1/3", "4/4", "3/4", "1/3", measure.LevelBars, []string{"1/4@1:3/4", "1/12@2:0/1"}},
		{"КП-27 change 7/4 and 1/2", "4/4;3:3/4", "7/4", "1/2", measure.LevelBars, []string{"1/4@2:3/4", "1/4@3:0/1"}},
		{"КП-27 change 1 and 5/2", "4/4;3:3/4", "1", "5/2", measure.LevelBars, []string{"1/1@2:0/1", "3/4@3:0/1", "3/4@4:0/1"}},
		{"КП-27 pickup 0 and 5/4", "3/4;pickup:1/4", "0", "5/4", measure.LevelBars, []string{"1/4@0:0/1", "3/4@1:0/1", "1/4@2:0/1"}},
		{"КП-29 extended by bars", "4/4;2:length:5/4", "7/4", "1/2", measure.LevelBars, []string{"1/2@2:3/4"}},
		{"КП-52 one whole bar", "4/4", "0", "1", measure.LevelBars, []string{"1/1@1:0/1"}},
		{"КП-53 N = 1", "4/4;bars:1", "0", "1", measure.LevelBars, []string{"1/1@1:0/1"}},
		{"КП-65 extended by bars", "4/4;2:length:3/2", "7/4", "3/4", measure.LevelBars, []string{"3/4@2:3/4"}},
		{"explicit length", "4/4;3:length:1/2", "15/8", "1", measure.LevelBars, []string{"1/8@2:7/8", "1/2@3:0/1", "3/8@4:0/1"}},
		{"to the end of a finite grid", "4/4;bars:4", "7/2", "1/2", measure.LevelBars, []string{"1/2@4:1/2"}},
		{"inside one bar", "4/4", "1/8", "1/4", measure.LevelBars, []string{"1/4@1:1/8"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			parts, err := g.Split(pos(tt.start), dur(tt.value))
			require.NoError(t, err)
			assert.Equal(t, tt.want, partTexts(parts))
			assertParts(t, g, pos(tt.start), dur(tt.value), parts)

			by, err := g.SplitBy(pos(tt.start), dur(tt.value), measure.LevelBars)
			require.NoError(t, err)
			assert.Equal(t, parts, by, "Split is SplitBy with LevelBars")
		})
	}

	t.Run("КП-27 starts 1, 2, 11/4", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		parts, err := gridOf(t, "4/4;3:3/4").Split(pos("1"), dur("5/2"))
		require.NoError(t, err)
		require.Len(t, parts, 3)
		assertText(t, "1/1", parts[0].Start)
		assertText(t, "2/1", parts[1].Start)
		assertText(t, "11/4", parts[2].Start)
		assert.Equal(t, measure.Part{Address: ad(3, "0"), Start: pos("2"), Duration: dur("3/4")}, parts[1])
	})

	t.Run("КП-52 1024 parts", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		g := gridOf(t, "1/4")
		parts, err := g.Split(pos("0"), dur("256"))
		require.NoError(t, err)
		require.Len(t, parts, measure.MaxParts)
		assert.Equal(t, measure.Part{Address: ad(1024, "0"), Start: pos("1023/4"), Duration: dur("1/4")}, parts[1023])
		assertParts(t, g, pos("0"), dur("256"), parts)
	})
}

// TestSplitLevels covers О-17, ИС-14, П-10, and КП-28, КП-29, КП-65: the
// boundaries of groups, beats, and units of the meter of each bar, shifted
// by the place in the meter; in an extended bar units continue after the
// size of the meter, groups and beats do not.
func TestSplitLevels(t *testing.T) {
	t.Parallel()

	tests := []splitCase{
		{"КП-28 4/4 groups 3/8 and 3/8", "4/4", "3/8", "3/8", measure.LevelGroups, []string{"1/8@1:3/8", "1/4@1:1/2"}},
		{"КП-28 4/4 groups 3/4 and 1", "4/4", "3/4", "1", measure.LevelGroups, []string{"1/4@1:3/4", "1/2@2:0/1", "1/4@2:1/2"}},
		{"КП-28 6/8 beats 1/4 and 1/4", "6/8", "1/4", "1/4", measure.LevelBeats, []string{"1/8@1:1/4", "1/8@1:3/8"}},
		{"КП-28 6/8 groups 1/8 and 1/2", "6/8", "1/8", "1/2", measure.LevelGroups, []string{"1/4@1:1/8", "1/4@1:3/8"}},
		{"КП-28 3/4 units", "3/4", "1/8", "1/2", measure.LevelUnits, []string{"1/8@1:1/8", "1/4@1:1/4", "1/8@1:1/2"}},
		{"КП-28 3/4 beats", "3/4", "1/8", "1/2", measure.LevelBeats, []string{"1/8@1:1/8", "1/4@1:1/4", "1/8@1:1/2"}},
		{"КП-28 3/4 groups", "3/4", "1/8", "1/2", measure.LevelGroups, []string{"1/2@1:1/8"}},
		{"КП-28 4/4 beats 1/8 and 3/4", "4/4", "1/8", "3/4", measure.LevelBeats, []string{"1/8@1:1/8", "1/4@1:1/4", "1/4@1:1/2", "1/8@1:3/4"}},
		{"КП-28 5/8[2+3] groups", "5/8[2+3]", "0", "5/8", measure.LevelGroups, []string{"1/4@1:0/1", "3/8@1:1/4"}},
		{"КП-29 4/4 pickup 3/8 units", "4/4;pickup:3/8", "0", "3/8", measure.LevelUnits, []string{"1/8@0:0/1", "1/4@0:1/8"}},
		{"КП-29 6/8 pickup 1/2 groups", "6/8;pickup:1/2", "0", "1/2", measure.LevelGroups, []string{"1/8@0:0/1", "3/8@0:1/8"}},
		{"КП-29 extended units", "4/4;2:length:5/4", "7/4", "1/2", measure.LevelUnits, []string{"1/4@2:3/4", "1/4@2:1/1"}},
		{"КП-65 units", "4/4;2:length:3/2", "7/4", "3/4", measure.LevelUnits, []string{"1/4@2:3/4", "1/4@2:1/1", "1/4@2:5/4"}},
		{"КП-65 beats", "4/4;2:length:3/2", "7/4", "3/4", measure.LevelBeats, []string{"1/4@2:3/4", "1/2@2:1/1"}},
		{"КП-65 groups", "4/4;2:length:3/2", "7/4", "3/4", measure.LevelGroups, []string{"1/4@2:3/4", "1/2@2:1/1"}},
		{"to the end, units", "4/4;3:length:1/4:end", "15/8", "3/8", measure.LevelUnits, []string{"1/8@2:7/8", "1/4@3:0/1"}},
		{"to the end, aligned like a pickup", "4/4;3:length:3/8:end", "2", "3/8", measure.LevelUnits, []string{"1/8@3:0/1", "1/4@3:1/8"}},
		{"from the start, shorter", "4/4;3:length:3/8", "2", "3/8", measure.LevelUnits, []string{"1/4@3:0/1", "1/8@3:1/4"}},
		{"sum by units", "2/4+3/8", "3/8", "1/4", measure.LevelUnits, []string{"1/8@1:3/8", "1/8@1:1/2"}},
		{"across bars by groups", "6/8", "1/2", "1/2", measure.LevelGroups, []string{"1/4@1:1/2", "1/4@2:0/1"}},
		{"5/8 by units, no grouping needed", "5/8", "0", "3/8", measure.LevelUnits, []string{"1/8@1:0/1", "1/8@1:1/8", "1/8@1:1/4"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)

			parts, err := g.SplitBy(pos(tt.start), dur(tt.value), tt.level)
			require.NoError(t, err)
			assert.Equal(t, tt.want, partTexts(parts))
			assertParts(t, g, pos(tt.start), dur(tt.value), parts)
		})
	}
}

// TestSplitErrors covers architecture 4.2 for Split and SplitBy and КП-39,
// КП-41, КП-44, КП-45, КП-53.
func TestSplitErrors(t *testing.T) {
	t.Parallel()

	huge := durRat(ratInt(pow2(100)))

	tests := []struct {
		name    string
		grid    string
		start   duration.Position
		value   duration.Duration
		level   measure.Level
		err     error
		causes  []error
		without []error
	}{
		{"unset start", "4/4", duration.Position{}, dur("1"), measure.LevelBars, measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"КП-39 unset value", "4/4", pos("0"), duration.Duration{}, measure.LevelBars, measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"level 0", "4/4", pos("0"), dur("1"), measure.Level(0), measure.ErrInvalidArgument, nil, nil},
		{"level 5", "4/4", pos("0"), dur("1"), measure.Level(5), measure.ErrInvalidArgument, nil, nil},
		{"unset value before level", "4/4", pos("0"), duration.Duration{}, measure.Level(0), measure.ErrInvalidArgument, []error{duration.ErrUnset}, nil},
		{"level before outside the grid", "4/4;bars:4", pos("7/2"), dur("1"), measure.Level(9), measure.ErrInvalidArgument, nil, nil},
		{"КП-41 7/2 and 1 of N = 4", "4/4;bars:4", pos("7/2"), dur("1"), measure.LevelBars, measure.ErrOutsideGrid, nil, nil},
		{"КП-53 0 and 5/4 of N = 1", "4/4;bars:1", pos("0"), dur("5/4"), measure.LevelBars, measure.ErrOutsideGrid, nil, nil},
		{"outside before grouping", "5/8;bars:2", pos("0"), dur("2"), measure.LevelGroups, measure.ErrOutsideGrid, nil, nil},
		{"КП-44 5/8 groups", "5/8", pos("0"), dur("5/8"), measure.LevelGroups, measure.ErrGroupingRequired, []error{meter.ErrGroupingRequired}, nil},
		{"КП-44 5/8 beats", "5/8", pos("0"), dur("5/8"), measure.LevelBeats, measure.ErrGroupingRequired, []error{meter.ErrGroupingRequired}, nil},
		{"КП-44 change to 7/8", "4/4;2:7/8", pos("3/4"), dur("1/2"), measure.LevelGroups, measure.ErrGroupingRequired, []error{meter.ErrGroupingRequired}, nil},
		{"grouping before too many", "5/8", pos("0"), dur("1000"), measure.LevelGroups, measure.ErrGroupingRequired, []error{meter.ErrGroupingRequired}, nil},
		{"КП-45 1/4 0 and 1025/4", "1/4", pos("0"), dur("1025/4"), measure.LevelBars, measure.ErrTooMany, nil, nil},
		{"КП-45 4/4 0 and 2¹⁰⁰", "4/4", pos("0"), huge, measure.LevelBars, measure.ErrTooMany, nil, nil},
		{"КП-45 2¹⁰⁰ by units", "4/4", pos("0"), huge, measure.LevelUnits, measure.ErrTooMany, nil, nil},
		{"part 1025 by units", "4/4", pos("0"), dur("257"), measure.LevelUnits, measure.ErrTooMany, nil, nil},
		{"1024 parts by units", "4/4", pos("1/4"), dur("1024/4"), measure.LevelUnits, nil, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			g := gridOf(t, tt.grid)
			parts, err := g.SplitBy(tt.start, tt.value, tt.level)

			if tt.err == nil {
				require.NoError(t, err)
				assert.Len(t, parts, measure.MaxParts)

				return
			}

			assertErr(t, err, tt.err, tt.causes...)
			assertNoCause(t, err, tt.without...)
			assert.Nil(t, parts)

			if tt.level == measure.LevelBars {
				parts, err := g.Split(tt.start, tt.value)
				assertErr(t, err, tt.err, tt.causes...)
				assert.Nil(t, parts)
			}
		})
	}

	t.Run("grouping without enumerating bars", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		_, err := gridOf(t, "5/8").SplitBy(pos("0"), durRat(ratInt(pow2(40))), measure.LevelGroups)
		assertErr(t, err, measure.ErrGroupingRequired, meter.ErrGroupingRequired)
	})
}
