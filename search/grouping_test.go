package search_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/search"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func spellingGroups(groups [][]search.SpellingMatch) [][]spellingPair {
	result := make([][]spellingPair, len(groups))
	for i, group := range groups {
		result[i] = spellingPairs(group)
	}
	return result
}

func positionGroups(groups [][]search.PositionMatch) [][]positionPair {
	result := make([][]positionPair, len(groups))
	for i, group := range groups {
		result[i] = positionPairs(group)
	}
	return result
}

func TestGroupSpellingByCollectionControlExample8(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian", "aeolian")
	matches := searcher.MustBySpelling(notes("D", "E", "F♯", "A", "B"), search.Containment)
	groups, err := search.GroupSpellingByCollection(matches)
	require.NoError(t, err)
	assert.Equal(t, [][]spellingPair{
		{{"ionian", "D"}, {"aeolian", "B"}},
		{{"ionian", "G"}, {"aeolian", "E"}},
		{{"ionian", "A"}, {"aeolian", "F♯"}},
	}, spellingGroups(groups))
	assertSpellingGroupingOwnership(t, matches, groups)
}

func TestGroupSpellingByTonicAndStructure(t *testing.T) {
	t.Parallel()
	first := builtInEntry(t, "hirajoshi-sachs")
	second := builtInEntry(t, "iwato")
	searcher := searcherFor(t, first, second)
	matches := searcher.MustBySpelling(notes("C", "D♭", "F", "G♭", "B♭"), search.FullMatch)
	require.Len(t, matches, 2)

	byTonic, err := search.GroupSpellingByTonic(matches)
	require.NoError(t, err)
	assert.Equal(t, [][]spellingPair{{{"hirajoshi-sachs", "C"}, {"iwato", "C"}}}, spellingGroups(byTonic))
	byStructure, err := search.GroupSpellingByStructure(matches)
	require.NoError(t, err)
	assert.Equal(t, spellingGroups(byTonic), spellingGroups(byStructure))

	searcher = builtInSearcher(t, "ionian", "aeolian")
	matches = searcher.MustBySpelling(notes("D", "E", "F♯", "A", "B"), search.Containment)
	byTonic, err = search.GroupSpellingByTonic(matches)
	require.NoError(t, err)
	assert.Len(t, byTonic, len(matches))
	byStructure, err = search.GroupSpellingByStructure(matches)
	require.NoError(t, err)
	assert.Equal(t, [][]spellingPair{
		{{"ionian", "D"}, {"ionian", "G"}, {"ionian", "A"}},
		{{"aeolian", "B"}, {"aeolian", "E"}, {"aeolian", "F♯"}},
	}, spellingGroups(byStructure))
}

func TestGroupPositionRepresentations(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian", "aeolian")
	matches := searcher.MustByPosition(system.Traditional(), notes("D", "E", "F♯", "A", "B"), search.Containment)
	require.NotEmpty(t, matches)

	byCollection, err := search.GroupPositionByCollection(matches)
	require.NoError(t, err)
	require.Len(t, byCollection, 3)
	assert.Equal(t, [][]positionPair{
		{{Code: "ionian", Class: 2, Tonic: "D"}, {Code: "aeolian", Class: 11, Tonic: "B"}},
		{{Code: "ionian", Class: 7, Tonic: "G"}, {Code: "aeolian", Class: 4, Tonic: "E"}},
		{{Code: "ionian", Class: 9, Tonic: "A"}, {Code: "aeolian", Class: 6, Tonic: "F♯"}},
	}, positionGroups(byCollection))
	assertPositionGroupingOwnership(t, matches, byCollection)

	byTonic, err := search.GroupPositionByTonic(matches)
	require.NoError(t, err)
	assert.Len(t, byTonic, len(matches))
	byStructure, err := search.GroupPositionByStructure(matches)
	require.NoError(t, err)
	assert.Len(t, byStructure, len(matches))
}

func TestGroupingNilEmptyAndUnset(t *testing.T) {
	t.Parallel()
	spellingTests := []struct {
		name string
		call func([]search.SpellingMatch) ([][]search.SpellingMatch, error)
		must func([]search.SpellingMatch) [][]search.SpellingMatch
	}{
		{name: "spelling collection", call: search.GroupSpellingByCollection, must: search.MustGroupSpellingByCollection},
		{name: "spelling tonic", call: search.GroupSpellingByTonic, must: search.MustGroupSpellingByTonic},
		{name: "spelling structure", call: search.GroupSpellingByStructure, must: search.MustGroupSpellingByStructure},
	}
	for _, tt := range spellingTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			groups, err := tt.call(nil)
			assert.Nil(t, groups)
			requireError(t, err, search.ErrUnset)
			requireMustPanic(t, err, func() { tt.must(nil) })
			groups, err = tt.call([]search.SpellingMatch{})
			requireNonNilEmpty(t, groups, err)
			groups, err = tt.call([]search.SpellingMatch{{}})
			assert.Nil(t, groups)
			requireError(t, err, search.ErrUnset)
		})
	}

	positionTests := []struct {
		name string
		call func([]search.PositionMatch) ([][]search.PositionMatch, error)
		must func([]search.PositionMatch) [][]search.PositionMatch
	}{
		{name: "position collection", call: search.GroupPositionByCollection, must: search.MustGroupPositionByCollection},
		{name: "position tonic", call: search.GroupPositionByTonic, must: search.MustGroupPositionByTonic},
		{name: "position structure", call: search.GroupPositionByStructure, must: search.MustGroupPositionByStructure},
	}
	for _, tt := range positionTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			groups, err := tt.call(nil)
			assert.Nil(t, groups)
			requireError(t, err, search.ErrUnset)
			requireMustPanic(t, err, func() { tt.must(nil) })
			groups, err = tt.call([]search.PositionMatch{})
			requireNonNilEmpty(t, groups, err)
			groups, err = tt.call([]search.PositionMatch{{}})
			assert.Nil(t, groups)
			requireError(t, err, search.ErrUnset)
		})
	}
}

func TestPositionGroupingRejectsDifferentSystems(t *testing.T) {
	t.Parallel()
	searcher := searcherFor(t, onePlaceEntry("one"))
	traditional := searcher.MustByPosition(system.Traditional(), notes("C"), search.FullMatch)
	nr := searcher.MustByPosition(nonRegularSystem(), notes("C"), search.FullMatch)
	require.Len(t, traditional, 1)
	require.Len(t, nr, 1)
	input := []search.PositionMatch{traditional[0], nr[0]}
	calls := []struct {
		name string
		call func([]search.PositionMatch) ([][]search.PositionMatch, error)
		must func([]search.PositionMatch) [][]search.PositionMatch
	}{
		{name: "collection", call: search.GroupPositionByCollection, must: search.MustGroupPositionByCollection},
		{name: "tonic", call: search.GroupPositionByTonic, must: search.MustGroupPositionByTonic},
		{name: "structure", call: search.GroupPositionByStructure, must: search.MustGroupPositionByStructure},
	}
	for _, tt := range calls {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			groups, err := tt.call(input)
			assert.Nil(t, groups)
			requireError(t, err, search.ErrDifferentSystems)
			requireMustPanic(t, err, func() { tt.must(input) })
		})
	}
}

func TestConcurrentGroupings(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian", "aeolian")
	spellingMatches := searcher.MustBySpelling(notes("D", "E", "F♯", "A", "B"), search.Containment)
	positionMatches := searcher.MustByPosition(system.Traditional(), notes("D", "E", "F♯", "A", "B"), search.Containment)
	expectedSpelling, err := search.GroupSpellingByCollection(spellingMatches)
	require.NoError(t, err)
	expectedPosition, err := search.GroupPositionByCollection(positionMatches)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			actualSpelling, spellingErr := search.GroupSpellingByCollection(spellingMatches)
			actualPosition, positionErr := search.GroupPositionByCollection(positionMatches)
			assert.NoError(t, spellingErr)
			assert.NoError(t, positionErr)
			assert.Equal(t, spellingGroups(expectedSpelling), spellingGroups(actualSpelling))
			assert.Equal(t, positionGroups(expectedPosition), positionGroups(actualPosition))
		})
	}
	wg.Wait()
}

func assertSpellingGroupingOwnership(t *testing.T, input []search.SpellingMatch, groups [][]search.SpellingMatch) {
	t.Helper()
	expected := spellingPairs(input)
	require.NotEmpty(t, groups)
	require.NotEmpty(t, groups[0])
	flattened := make([]spellingPair, 0)
	for _, group := range groups {
		flattened = append(flattened, spellingPairs(group)...)
	}
	assert.ElementsMatch(t, expected, flattened)
	assert.Len(t, flattened, len(expected))
	groups[0][0] = search.SpellingMatch{}
	assert.Equal(t, expected, spellingPairs(input))
}

func assertPositionGroupingOwnership(t *testing.T, input []search.PositionMatch, groups [][]search.PositionMatch) {
	t.Helper()
	expected := positionPairs(input)
	require.NotEmpty(t, groups)
	require.NotEmpty(t, groups[0])
	flattened := make([]positionPair, 0)
	for _, group := range groups {
		flattened = append(flattened, positionPairs(group)...)
	}
	assert.ElementsMatch(t, expected, flattened)
	assert.Len(t, flattened, len(expected))
	groups[0][0] = search.PositionMatch{}
	assert.Equal(t, expected, positionPairs(input))
}
