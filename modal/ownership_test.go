package modal_test

import (
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/modal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueResultsAreComparableIndependentCopies(t *testing.T) {
	t.Parallel()
	m := dorianMode(t)
	rp := modal.MustRelative(m, 6)
	rpCopy := rp
	assert.True(t, rp == rpCopy)
	assert.Equal(t, 1, map[modal.RelativePosition]int{rp: 1}[rpCopy])
	rpCopy.Number = 99
	rpCopy.Category = modal.Low
	rpCopy.Interval = interval.MinorSixth()
	assert.NotEqual(t, rp, rpCopy)
	assert.Equal(t, modal.RelativePosition{Number: 6, Category: modal.High, Interval: interval.MajorSixth()}, rp)
	ap := modal.MustAbsolute(m, 6)
	apCopy := ap
	assert.True(t, ap == apCopy)
	assert.Equal(t, 1, map[modal.AbsolutePosition]int{ap: 1}[apCopy])
	apCopy.Counts.Minor = 99
	apCopy.Score = 99
	apCopy.Rank = 99
	assert.NotEqual(t, ap, apCopy)
	assert.Equal(t, dorianAnalysis().Positions[5], ap)
	counts := ap.Counts
	countsCopy := counts
	assert.True(t, counts == countsCopy)
	assert.Equal(t, 1, map[modal.QualityCounts]int{counts: 1}[countsCopy])
	countsCopy.Perfect = 99
	assert.NotEqual(t, counts, countsCopy)
	assert.Equal(t, 2, counts.Perfect)
	assert.Zero(t, modal.RelativePosition{}.Number)
	assert.False(t, modal.RelativePosition{}.Interval.IsValid())
	assert.Equal(t, modal.CategoryUnset, modal.RelativePosition{}.Category)
	assert.Equal(t, modal.QualityCounts{}, modal.AbsolutePosition{}.Counts)
	assert.Nil(t, modal.RelativeGroups{}.High)
	assert.Nil(t, modal.AbsoluteGroup{}.Numbers)
	assert.Nil(t, modal.AbsoluteAnalysis{}.Positions)
	assert.Nil(t, modal.AbsoluteAnalysis{}.Order)
}

func TestResultsOwnTheirSlices(t *testing.T) {
	t.Parallel()
	m := dorianMode(t)
	before := m.DegreeIntervals()
	relative := modal.MustRelativePositions(m)
	relativeSnapshot := slices.Clone(relative)
	otherRelative := modal.MustRelativePositions(m)
	relative[0].Number = 99
	relative[1].Interval = interval.MinorSecond()
	relative = append(relative, modal.RelativePosition{Number: 99})
	assert.Len(t, relative, 8)
	assert.Equal(t, relativeSnapshot, otherRelative)
	assert.Equal(t, relativeSnapshot, modal.MustRelativePositions(m))

	groups := modal.MustGroupRelative(m)
	otherGroups := modal.MustGroupRelative(m)
	expectedGroups := modal.RelativeGroups{High: []int{2, 6}, Neutral: []int{1, 4, 5}, Low: []int{3, 7}}
	groups.High[0] = 99
	groups.High = append(groups.High, 98)
	assert.Equal(t, []int{1, 4, 5}, groups.Neutral)
	assert.Equal(t, []int{3, 7}, groups.Low)
	groups.Neutral[0] = 97
	groups.Neutral = append(groups.Neutral, 96)
	assert.Equal(t, []int{3, 7}, groups.Low)
	groups.Low[0] = 95
	groups.Low = append(groups.Low, 94)
	assert.Equal(t, expectedGroups, otherGroups)
	assert.Equal(t, expectedGroups, modal.MustGroupRelative(m))

	analysis := modal.MustAnalyzeAbsolute(m)
	otherAnalysis := modal.MustAnalyzeAbsolute(m)
	positions := modal.MustAbsolutePositions(m)
	order := modal.MustOrder(m)
	analysis.Positions[0].Number = 99
	analysis.Positions[0].Counts.Perfect = 99
	analysis.Positions = append(analysis.Positions, modal.AbsolutePosition{Number: 99})
	assert.Len(t, analysis.Positions, 8)
	assert.Equal(t, dorianAnalysis().Order, analysis.Order)
	analysis.Order[0].Numbers[0] = 99
	analysis.Order[0].Numbers = append(analysis.Order[0].Numbers, 98)
	assert.Equal(t, []int{2}, analysis.Order[1].Numbers)
	analysis.Order[0].Rank = 99
	analysis.Order = append(analysis.Order, modal.AbsoluteGroup{Rank: 99, Numbers: []int{99}})
	assert.Len(t, analysis.Order, 8)
	assert.Equal(t, dorianAnalysis(), otherAnalysis)
	assert.Equal(t, dorianAnalysis().Positions, positions)
	assert.Equal(t, dorianAnalysis().Order, order)
	positions[0].Score = 99
	positions = append(positions, modal.AbsolutePosition{Number: 99})
	assert.Len(t, positions, 8)
	order[0].Numbers[0] = 99
	order[0].Numbers = append(order[0].Numbers, 98)
	assert.Equal(t, []int{2}, order[1].Numbers)
	order = append(order, modal.AbsoluteGroup{Rank: 99})
	assert.Len(t, order, 8)
	assert.Equal(t, dorianAnalysis(), modal.MustAnalyzeAbsolute(m))
	assert.Equal(t, before, m.DegreeIntervals())
	assert.Equal(t, 7, m.Len())
	assert.True(t, m.IsClosed())
	assert.True(t, m.IsOctave())
	// Assigned slice containers retain ordinary Go shallow-copy semantics.
	shallow := otherGroups
	shallow.High[0] = 42
	assert.Equal(t, 42, otherGroups.High[0])
	shallowAnalysis := otherAnalysis
	shallowAnalysis.Order[0].Numbers[0] = 42
	assert.Equal(t, 42, otherAnalysis.Order[0].Numbers[0])
	assert.Equal(t, expectedGroups, modal.MustGroupRelative(m))
}

func TestModeMutationDoesNotChangeEarlierResults(t *testing.T) {
	t.Parallel()
	m := dorianMode(t)
	beforeIntervals := m.DegreeIntervals()
	before := make(map[string]any)
	for _, q := range queries() {
		result, err := q.run(m, 6, 5)
		require.NoError(t, err)
		before[q.name] = result
	}
	wantRelative := modal.MustRelativePositions(m)
	wantGroups := modal.MustGroupRelative(m)
	wantAbsolute := dorianAnalysis()
	require.NoError(t, m.SetDegreeInterval(6, interval.MinorSixth()))
	assert.Equal(t, beforeIntervals[5], interval.MajorSixth())
	assert.Equal(t, modal.RelativePosition{Number: 6, Category: modal.High, Interval: interval.MajorSixth()}, before["Relative"])
	assert.Equal(t, wantRelative, before["RelativePositions"])
	assert.Equal(t, wantGroups, before["GroupRelative"])
	assert.Equal(t, wantAbsolute.Positions[5], before["Absolute"])
	assert.Equal(t, wantAbsolute.Positions, before["AbsolutePositions"])
	assert.Equal(t, wantAbsolute.Order, before["Order"])
	assert.Equal(t, wantAbsolute, before["AnalyzeAbsolute"])
	assert.Equal(t, -1, before["CompareAbsolute"])
	newRelative, err := modal.Relative(m, 6)
	require.NoError(t, err)
	assert.Equal(t, modal.Low, newRelative.Category)
	assert.True(t, newRelative.Interval.Equal(interval.MinorSixth()))
	newGroups, err := modal.GroupRelative(m)
	require.NoError(t, err)
	assert.Equal(t, modal.RelativeGroups{High: []int{2}, Neutral: []int{1, 4, 5}, Low: []int{3, 6, 7}}, newGroups)
	newAbsolute, err := modal.AnalyzeAbsolute(m)
	require.NoError(t, err)
	require.Len(t, newAbsolute.Positions, 7)
	assert.Equal(t, 2, newAbsolute.Positions[0].Score)
	assert.Equal(t, -5, newAbsolute.Positions[5].Score)
	assert.NotEqual(t, wantAbsolute, newAbsolute)
	assert.Equal(t, newAbsolute.Positions, modal.MustAbsolutePositions(m))
	assert.Equal(t, newAbsolute.Order, modal.MustOrder(m))
	assert.Equal(t, newAbsolute.Positions[5], modal.MustAbsolute(m, 6))
	assert.Equal(t, 1, modal.MustCompareAbsolute(m, 6, 5))
}

func TestConcurrentReadQueriesHaveIndependentResults(t *testing.T) {
	t.Parallel()
	m := dorianMode(t)
	before := m.DegreeIntervals()
	ops := queries()
	want := make([]any, len(ops))
	for i, q := range ops {
		var err error
		want[i], err = q.run(m, 6, 5)
		require.NoError(t, err)
	}
	type observation struct {
		index  int
		result any
		must   any
		err    error
	}
	const workers = 16
	results := make(chan observation, workers*len(ops))
	start := make(chan struct{})
	var workersDone sync.WaitGroup
	for range workers {
		workersDone.Go(func() {
			<-start
			for i, q := range ops {
				got, err := q.run(m, 6, 5)
				results <- observation{index: i, result: got, must: q.must(m, 6, 5), err: err}
			}
		})
	}
	close(start)
	workersDone.Wait()
	close(results)
	for got := range results {
		require.NoError(t, got.err)
		assert.Equal(t, want[got.index], got.result, ops[got.index].name)
		assert.Equal(t, want[got.index], got.must, ops[got.index].name)
		// Changing a returned slice must not corrupt another worker's data.
		switch value := got.result.(type) {
		case []modal.RelativePosition:
			value[0].Number = 99
		case modal.RelativeGroups:
			value.Neutral[0] = 99
		case []modal.AbsolutePosition:
			value[0].Score = 99
		case []modal.AbsoluteGroup:
			value[0].Numbers[0] = 99
		case modal.AbsoluteAnalysis:
			value.Positions[0].Score = 99
			value.Order[0].Numbers[0] = 99
		}
	}
	assert.Equal(t, before, m.DegreeIntervals())
	assert.Equal(t, dorianAnalysis(), modal.MustAnalyzeAbsolute(m))
}
