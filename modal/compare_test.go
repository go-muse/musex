package modal_test

import (
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/modal"
	"github.com/go-muse/musex/mode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The expected comparisons below come from business-spec 21.2, MOD-8,
// example 4, and the comparison rows of the table "additional checks of the
// new queries". They are not derived from the code under test.

// side describes one side of one degree in a comparison.
type side struct {
	interval string
	category modal.Category
	score    int
	rank     int
}

type degreeChange struct {
	before, after side
}

func checkRelativeSide(t *testing.T, number int, want side, got modal.RelativePosition) {
	t.Helper()
	assert.Equal(t, number, got.Number, "relative number")
	assert.True(t, interval.MustParse(want.interval).Equal(got.Interval), "degree %d: interval %v, want %s", number, got.Interval, want.interval)
	assert.Equal(t, want.category, got.Category, "degree %d: category", number)
}

func checkRelativeComparison(t *testing.T, want []degreeChange, got modal.RelativeComparison) {
	t.Helper()
	require.Len(t, got.Degrees, len(want))
	for i, w := range want {
		d := got.Degrees[i]
		assert.Equal(t, i+1, d.Number)
		checkRelativeSide(t, i+1, w.before, d.Before)
		checkRelativeSide(t, i+1, w.after, d.After)
		assert.Equal(t, w.before.interval != w.after.interval, d.IntervalChanged, "degree %d: IntervalChanged", i+1)
		assert.Equal(t, w.before.category != w.after.category, d.CategoryChanged, "degree %d: CategoryChanged", i+1)
	}
}

func checkPositionComparison(t *testing.T, want []degreeChange, got modal.PositionComparison) {
	t.Helper()
	require.Len(t, got.Degrees, len(want))
	for i, w := range want {
		d := got.Degrees[i]
		assert.Equal(t, i+1, d.Number)
		checkRelativeSide(t, i+1, w.before, d.Before.Relative)
		checkRelativeSide(t, i+1, w.after, d.After.Relative)
		assert.Equal(t, i+1, d.Before.Absolute.Number)
		assert.Equal(t, i+1, d.After.Absolute.Number)
		assert.Equal(t, w.before.score, d.Before.Absolute.Score, "degree %d: score before", i+1)
		assert.Equal(t, w.after.score, d.After.Absolute.Score, "degree %d: score after", i+1)
		assert.Equal(t, w.before.rank, d.Before.Absolute.Rank, "degree %d: rank before", i+1)
		assert.Equal(t, w.after.rank, d.After.Absolute.Rank, "degree %d: rank after", i+1)
		assert.Equal(t, w.before.interval != w.after.interval, d.IntervalChanged, "degree %d: IntervalChanged", i+1)
		assert.Equal(t, w.before.category != w.after.category, d.CategoryChanged, "degree %d: CategoryChanged", i+1)
		assert.Equal(t, w.before.score != w.after.score, d.ScoreChanged, "degree %d: ScoreChanged", i+1)
		assert.Equal(t, w.before.rank != w.after.rank, d.RankChanged, "degree %d: RankChanged", i+1)
	}
}

// checkSidesMatchSeparateQueries checks that both sides equal the separate
// accepted queries for each mode.
func checkSidesMatchSeparateQueries(t *testing.T, before, after mode.Mode, got modal.PositionComparison) {
	t.Helper()
	relBefore, relAfter := modal.MustRelativePositions(before), modal.MustRelativePositions(after)
	absBefore, absAfter := modal.MustAnalyzeAbsolute(before).Positions, modal.MustAnalyzeAbsolute(after).Positions
	require.Len(t, got.Degrees, len(relBefore))
	for i, d := range got.Degrees {
		assert.Equal(t, relBefore[i], d.Before.Relative)
		assert.Equal(t, relAfter[i], d.After.Relative)
		assert.Equal(t, absBefore[i], d.Before.Absolute)
		assert.Equal(t, absAfter[i], d.After.Absolute)
	}
	relative := modal.MustCompareRelativePositions(before, after)
	require.Len(t, relative.Degrees, len(got.Degrees))
	for i, d := range got.Degrees {
		assert.Equal(t, modal.RelativeChange{
			Number: d.Number, Before: d.Before.Relative, After: d.After.Relative,
			IntervalChanged: d.IntervalChanged, CategoryChanged: d.CategoryChanged,
		}, relative.Degrees[i], "the relative comparison is the relative part of the full one")
	}
}

func naturalMinor(t *testing.T) mode.Mode {
	t.Helper()
	return closedMode(t, "P1 M2 m3 P4 P5 m6 m7", "P8")
}

func harmonicMinor(t *testing.T) mode.Mode {
	t.Helper()
	return closedMode(t, "P1 M2 m3 P4 P5 m6 M7", "P8")
}

func TestComparePositionsExample4(t *testing.T) {
	t.Parallel()
	const high, neutral, low = modal.High, modal.Neutral, modal.Low
	before, after := naturalMinor(t), harmonicMinor(t)
	want := []degreeChange{
		{side{"P1", neutral, 2, 3}, side{"P1", neutral, 0, 4}},
		{side{"M2", high, 5, 1}, side{"M2", high, 3, 2}},
		{side{"m3", low, -4, 6}, side{"m3", low, -5, 6}},
		{side{"P4", neutral, 0, 4}, side{"P4", neutral, -1, 5}},
		{side{"P5", neutral, 4, 2}, side{"P5", neutral, 2, 3}},
		{side{"m6", low, -5, 7}, side{"m6", low, -5, 6}},
		{side{"m7", low, -2, 5}, side{"M7", high, 6, 1}},
	}
	got, err := modal.ComparePositions(before, after)
	require.NoError(t, err)
	checkPositionComparison(t, want, got)
	checkSidesMatchSeparateQueries(t, before, after, got)
	assert.Equal(t, []int{7}, got.IntervalChanges())
	assert.Equal(t, []int{7}, got.CategoryChanges())
	assert.Equal(t, []int{1, 2, 3, 4, 5, 7}, got.ScoreChanges())
	assert.Equal(t, []int{1, 2, 4, 5, 6, 7}, got.RankChanges())
	counts := func(d, m, p, maj, a int) modal.QualityCounts {
		return modal.QualityCounts{Diminished: d, Minor: m, Perfect: p, Major: maj, Augmented: a}
	}
	assert.Equal(t, counts(0, 3, 3, 1, 0), got.Degrees[0].Before.Absolute.Counts)
	assert.Equal(t, counts(0, 2, 3, 2, 0), got.Degrees[0].After.Absolute.Counts)
	assert.Equal(t, counts(1, 4, 2, 0, 0), got.Degrees[1].Before.Absolute.Counts)
	assert.Equal(t, counts(1, 3, 2, 1, 0), got.Degrees[1].After.Absolute.Counts)
	assert.Equal(t, counts(0, 0, 2, 4, 1), got.Degrees[5].Before.Absolute.Counts)
	assert.Equal(t, counts(0, 0, 2, 3, 2), got.Degrees[5].After.Absolute.Counts)
	assert.Equal(t, got, modal.MustComparePositions(before, after))

	relative, err := modal.CompareRelativePositions(before, after)
	require.NoError(t, err)
	checkRelativeComparison(t, want, relative)
	assert.Equal(t, []int{7}, relative.IntervalChanges())
	assert.Equal(t, []int{7}, relative.CategoryChanges())
	assert.Equal(t, relative, modal.MustCompareRelativePositions(before, after))

	t.Run("swapped arguments swap the sides", func(t *testing.T) {
		t.Parallel()
		swapped := modal.MustComparePositions(after, before)
		require.Len(t, swapped.Degrees, len(got.Degrees))
		for i, d := range got.Degrees {
			assert.Equal(t, modal.PositionChange{
				Number: d.Number, Before: d.After, After: d.Before,
				IntervalChanged: d.IntervalChanged, CategoryChanged: d.CategoryChanged,
				ScoreChanged: d.ScoreChanged, RankChanged: d.RankChanged,
			}, swapped.Degrees[i])
		}
		swappedRelative := modal.MustCompareRelativePositions(after, before)
		require.Len(t, swappedRelative.Degrees, len(relative.Degrees))
		for i, d := range relative.Degrees {
			assert.Equal(t, modal.RelativeChange{
				Number: d.Number, Before: d.After, After: d.Before,
				IntervalChanged: d.IntervalChanged, CategoryChanged: d.CategoryChanged,
			}, swappedRelative.Degrees[i])
		}
	})
}

func TestComparisonOfEqualModes(t *testing.T) {
	t.Parallel()
	dorian := dorianMode(t)
	other := dorianMode(t)
	full := modal.MustComparePositions(dorian, other)
	relative := modal.MustCompareRelativePositions(dorian, other)
	require.Len(t, full.Degrees, 7)
	require.Len(t, relative.Degrees, 7)
	for i := range full.Degrees {
		f, r := full.Degrees[i], relative.Degrees[i]
		assert.Equal(t, i+1, f.Number)
		assert.Equal(t, f.Before, f.After, "degree %d", i+1)
		assert.False(t, f.IntervalChanged || f.CategoryChanged || f.ScoreChanged || f.RankChanged, "degree %d", i+1)
		assert.Equal(t, r.Before, r.After, "degree %d", i+1)
		assert.False(t, r.IntervalChanged || r.CategoryChanged, "degree %d", i+1)
	}
	checkSidesMatchSeparateQueries(t, dorian, other, full)
	// Example 2 of modal.md: degree IV of Dorian has score -2 and rank 5.
	assert.Equal(t, -2, full.Degrees[3].Before.Absolute.Score)
	assert.Equal(t, 5, full.Degrees[3].Before.Absolute.Rank)
	for name, changes := range map[string][]int{
		"IntervalChanges": full.IntervalChanges(), "CategoryChanges": full.CategoryChanges(),
		"ScoreChanges": full.ScoreChanges(), "RankChanges": full.RankChanges(),
		"relative IntervalChanges": relative.IntervalChanges(), "relative CategoryChanges": relative.CategoryChanges(),
	} {
		assert.NotNil(t, changes, name)
		assert.Empty(t, changes, name)
	}
}

func TestComparisonOfDifferentLetterSizes(t *testing.T) {
	t.Parallel()
	const high, neutral, low = modal.High, modal.Neutral, modal.Low
	before := pentatonicMode(t)
	after := closedMode(t, "P1 m3 P4 P5 m7", "P8")
	want := []degreeChange{
		{side{"P1", neutral, -3, 5}, side{"P1", neutral, 2, 2}},
		{side{"M2", high, 0, 3}, side{"m3", low, -3, 5}},
		{side{"M3", high, 3, 1}, side{"P4", neutral, 0, 3}},
		{side{"P5", neutral, -2, 4}, side{"P5", neutral, 3, 1}},
		{side{"M6", high, 2, 2}, side{"m7", low, -2, 4}},
	}
	got := modal.MustComparePositions(before, after)
	checkPositionComparison(t, want, got)
	checkSidesMatchSeparateQueries(t, before, after, got)
	assert.Equal(t, []int{2, 3, 5}, got.IntervalChanges())
	assert.Equal(t, []int{2, 3, 5}, got.CategoryChanges())
	assert.Equal(t, []int{1, 2, 3, 4, 5}, got.ScoreChanges())
	assert.Equal(t, []int{1, 2, 3, 4, 5}, got.RankChanges())
}

func TestRelativeComparisonOfOpenAndNonOctaveModes(t *testing.T) {
	t.Parallel()
	const high, neutral, low = modal.High, modal.Neutral, modal.Low
	before := openMode(t, "P1 m3 P4")
	after := closedMode(t, "P1 M2 M3", "P4")
	got, err := modal.CompareRelativePositions(before, after)
	require.NoError(t, err)
	checkRelativeComparison(t, []degreeChange{
		{side{interval: "P1", category: neutral}, side{interval: "P1", category: neutral}},
		{side{interval: "m3", category: low}, side{interval: "M2", category: high}},
		{side{interval: "P4", category: neutral}, side{interval: "M3", category: high}},
	}, got)
	assert.Equal(t, []int{2, 3}, got.IntervalChanges())
	assert.Equal(t, []int{2, 3}, got.CategoryChanges())
	relBefore, relAfter := modal.MustRelativePositions(before), modal.MustRelativePositions(after)
	for i, d := range got.Degrees {
		assert.Equal(t, relBefore[i], d.Before)
		assert.Equal(t, relAfter[i], d.After)
	}
}

// TestIntervalChangeComparesWrittenIntervals checks that the interval flag
// compares specific intervals, not positional sizes: A4 and d5 differ, and
// M3 and A2 differ although both are high.
func TestIntervalChangeComparesWrittenIntervals(t *testing.T) {
	t.Parallel()
	const high, low = modal.High, modal.Low
	augmented := closedMode(t, "P1 A4", "P8")
	diminished := closedMode(t, "P1 d5", "P8")
	full := modal.MustComparePositions(augmented, diminished)
	checkPositionComparison(t, []degreeChange{
		{side{"P1", modal.Neutral, -1, 2}, side{"P1", modal.Neutral, 1, 1}},
		{side{"A4", high, 1, 1}, side{"d5", low, -1, 2}},
	}, full)
	checkSidesMatchSeparateQueries(t, augmented, diminished, full)
	assert.Equal(t, []int{2}, full.IntervalChanges())

	third := closedMode(t, "P1 M3", "P8")
	second := closedMode(t, "P1 A2", "P8")
	relative := modal.MustCompareRelativePositions(third, second)
	require.Len(t, relative.Degrees, 2)
	assert.True(t, relative.Degrees[1].IntervalChanged, "M3 and A2 are different intervals")
	assert.False(t, relative.Degrees[1].CategoryChanged, "M3 and A2 are both high")
	assert.Equal(t, []int{2}, relative.IntervalChanges())
	assert.Equal(t, []int{}, relative.CategoryChanges())
	assert.Equal(t, high, relative.Degrees[1].After.Category)
}

func TestComparisonFailures(t *testing.T) {
	t.Parallel()
	ionian := ionianMode(t)
	dorian := dorianMode(t)
	pentatonic := pentatonicMode(t)
	open := openMode(t, "P1 m3 P4")
	fourth := closedMode(t, "P1 M2 M3", "P4")
	neutralOpen := openMode(t, "P1 M2 n3")
	control := controlMode(t)
	plain := closedMode(t, "P1 M2 M3 P5", "P8")
	neutralOctave := closedMode(t, "P1 M2 n3 P5", "P8")
	unset := []error{modal.ErrUnset, mode.ErrUnset}
	structure := []error{modal.ErrUnsupportedStructure}
	incomparable := []error{modal.ErrIncomparable}
	quality := []error{modal.ErrUnsupportedQuality}
	for _, tt := range []struct {
		name          string
		before, after mode.Mode
		full          []error
		relative      []error
	}{
		{"unset before", mode.Mode{}, dorian, unset, unset},
		{"unset after", dorian, mode.Mode{}, unset, unset},
		{"both unset", mode.Mode{}, mode.Mode{}, unset, unset},
		{"unset before an open after", mode.Mode{}, open, unset, unset},
		{"unset after an open before", open, mode.Mode{}, unset, unset},
		{"unset after an incomparable before", pentatonic, mode.Mode{}, unset, unset},
		{"different numbers of degrees", ionian, pentatonic, incomparable, incomparable},
		{"different numbers of degrees, swapped", pentatonic, ionian, incomparable, incomparable},
		{"structure before comparability, open after", ionian, open, structure, incomparable},
		{"structure before comparability, open before", open, ionian, structure, incomparable},
		{"structure of the P4 cycle before", fourth, ionian, structure, incomparable},
		{"structure of the P4 cycle after", ionian, fourth, structure, incomparable},
		{"open and P4 cycle", open, fourth, structure, nil},
		{"comparability before qualities", neutralOpen, ionian, structure, incomparable},
		{"unsupported quality after", open, neutralOpen, structure, quality},
		{"unsupported quality before", neutralOpen, open, structure, quality},
		{"control before: representability, relative quality", control, plain, representable(), quality},
		{"control after: representability, relative quality", plain, control, representable(), quality},
		{"before is checked first: quality of before", neutralOctave, control, quality, quality},
		{"before is checked first: representability of before", control, neutralOctave, representable(), quality},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			full, err := modal.ComparePositions(tt.before, tt.after)
			checkModalError(t, err, tt.full...)
			assert.Equal(t, modal.PositionComparison{}, full, "no partial result on failure")
			assert.Nil(t, full.Degrees)
			checkMustPanic(t, err, func() { modal.MustComparePositions(tt.before, tt.after) })
			if slices.Contains(tt.full, mode.ErrOutOfRange) {
				assert.NotErrorIs(t, err, modal.ErrUnsupportedQuality, "representability is checked before qualities")
			}
			if tt.full != nil && tt.full[0] == modal.ErrIncomparable {
				assert.NotErrorIs(t, err, mode.ErrIncomparable, "modal.ErrIncomparable is its own sentinel")
			}

			relative, err := modal.CompareRelativePositions(tt.before, tt.after)
			if tt.relative == nil {
				require.NoError(t, err)
				return
			}
			checkModalError(t, err, tt.relative...)
			assert.Equal(t, modal.RelativeComparison{}, relative, "no partial result on failure")
			assert.Nil(t, relative.Degrees)
			checkMustPanic(t, err, func() { modal.MustCompareRelativePositions(tt.before, tt.after) })
			if tt.relative[0] == modal.ErrIncomparable {
				assert.NotErrorIs(t, err, mode.ErrIncomparable, "modal.ErrIncomparable is its own sentinel")
			}
		})
	}
}

func TestIncomparableIsADistinctSentinel(t *testing.T) {
	t.Parallel()
	require.NotNil(t, modal.ErrIncomparable)
	assert.NotErrorIs(t, modal.ErrIncomparable, mode.ErrIncomparable)
	assert.NotErrorIs(t, mode.ErrIncomparable, modal.ErrIncomparable)
	for _, other := range []error{modal.ErrUnset, modal.ErrInvalidDegree, modal.ErrUnsupportedStructure, modal.ErrUnsupportedQuality} {
		assert.NotErrorIs(t, modal.ErrIncomparable, other)
		assert.NotErrorIs(t, other, modal.ErrIncomparable)
	}
}

func TestComparisonValuesAndChangeLists(t *testing.T) {
	t.Parallel()
	t.Run("zero containers report no changes", func(t *testing.T) {
		t.Parallel()
		var relative modal.RelativeComparison
		var full modal.PositionComparison
		for name, changes := range map[string][]int{
			"relative IntervalChanges": relative.IntervalChanges(), "relative CategoryChanges": relative.CategoryChanges(),
			"IntervalChanges": full.IntervalChanges(), "CategoryChanges": full.CategoryChanges(),
			"ScoreChanges": full.ScoreChanges(), "RankChanges": full.RankChanges(),
		} {
			assert.NotNil(t, changes, name)
			assert.Empty(t, changes, name)
		}
	})
	t.Run("lists read the flags in the order of Degrees", func(t *testing.T) {
		t.Parallel()
		full := modal.PositionComparison{Degrees: []modal.PositionChange{
			{Number: 3, ScoreChanged: true, CategoryChanged: true},
			{Number: 1, ScoreChanged: true, RankChanged: true},
			{Number: 2, IntervalChanged: true},
		}}
		assert.Equal(t, []int{2}, full.IntervalChanges())
		assert.Equal(t, []int{3}, full.CategoryChanges())
		assert.Equal(t, []int{3, 1}, full.ScoreChanges())
		assert.Equal(t, []int{1}, full.RankChanges())
		relative := modal.RelativeComparison{Degrees: []modal.RelativeChange{
			{Number: 5, CategoryChanged: true}, {Number: 4, IntervalChanged: true, CategoryChanged: true},
		}}
		assert.Equal(t, []int{4}, relative.IntervalChanges())
		assert.Equal(t, []int{5, 4}, relative.CategoryChanges())
	})
	t.Run("every call returns a new list", func(t *testing.T) {
		t.Parallel()
		full := modal.MustComparePositions(naturalMinor(t), harmonicMinor(t))
		scores := full.ScoreChanges()
		scores[0] = 99
		_ = append(scores[:1], 98)
		assert.Equal(t, []int{1, 2, 3, 4, 5, 7}, full.ScoreChanges())
		empty := modal.MustComparePositions(dorianMode(t), dorianMode(t)).RankChanges()
		_ = append(empty, 1)
		relative := modal.MustCompareRelativePositions(naturalMinor(t), harmonicMinor(t))
		categories := relative.CategoryChanges()
		categories[0] = 99
		assert.Equal(t, []int{7}, relative.CategoryChanges())
		assert.True(t, full.Degrees[6].CategoryChanged, "changing a list does not change the flags")
	})
	t.Run("records are comparable independent copies", func(t *testing.T) {
		t.Parallel()
		// The records are comparable: the compiler accepts them as map keys.
		relativeKeys := map[modal.RelativeChange]int{{}: 1}
		positionKeys := map[modal.PositionChange]int{{}: 2}
		degreeKeys := map[modal.DegreePositions]int{{}: 3}
		relationKeys := map[modal.Relation]int{{}: 4}
		assert.Equal(t, 1, relativeKeys[modal.RelativeChange{}])
		assert.Equal(t, 2, positionKeys[modal.PositionChange{}])
		assert.Equal(t, 3, degreeKeys[modal.DegreePositions{}])
		assert.Equal(t, 4, relationKeys[modal.Relation{}])
		var zero modal.PositionChange
		assert.False(t, zero.IntervalChanged || zero.CategoryChanged || zero.ScoreChanged || zero.RankChanged)
		full := modal.MustComparePositions(naturalMinor(t), harmonicMinor(t))
		record := full.Degrees[6]
		copied := record
		copied.After.Absolute.Score = 100
		copied.RankChanged = false
		assert.Equal(t, 6, record.After.Absolute.Score)
		assert.True(t, record.RankChanged)
		assert.Equal(t, 6, full.Degrees[6].After.Absolute.Score)
		assert.True(t, record == full.Degrees[6])
		again := modal.MustComparePositions(naturalMinor(t), harmonicMinor(t))
		full.Degrees[0].Before.Absolute.Score = 100
		assert.Equal(t, 2, again.Degrees[0].Before.Absolute.Score, "results of different calls are independent")
	})
}

func TestComparisonsDoNotChangeTheModes(t *testing.T) {
	t.Parallel()
	before, after := naturalMinor(t), harmonicMinor(t)
	savedBefore, savedAfter := before, after
	_ = modal.MustComparePositions(before, after)
	_ = modal.MustCompareRelativePositions(before, after)
	assert.True(t, savedBefore.Equal(before))
	assert.True(t, savedAfter.Equal(after))
	assert.Equal(t, []int{7}, modal.MustComparePositions(before, after).CategoryChanges())
}

func TestComparisonsConcurrently(t *testing.T) {
	t.Parallel()
	before, after := naturalMinor(t), harmonicMinor(t)
	want := modal.MustComparePositions(before, after)
	wantRelative := modal.MustCompareRelativePositions(before, after)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 20 {
				got, err := modal.ComparePositions(before, after)
				assert.NoError(t, err)
				assert.Equal(t, want, got)
				relative, err := modal.CompareRelativePositions(before, after)
				assert.NoError(t, err)
				assert.Equal(t, wantRelative, relative)
				assert.Equal(t, []int{1, 2, 4, 5, 6, 7}, got.RankChanges())
			}
		})
	}
	wg.Wait()
}
