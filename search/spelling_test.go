package search_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/search"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBySpellingControlExample1(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	expected := []spellingPair{{"ionian", "D"}, {"ionian", "G"}, {"ionian", "A"}}
	expectedDegrees := map[string]map[string][]int{
		"D": {"D": {1}, "E": {2}, "F♯": {3}, "A": {5}, "B": {6}},
		"G": {"D": {5}, "E": {6}, "F♯": {7}, "A": {2}, "B": {3}},
		"A": {"D": {4}, "E": {5}, "F♯": {6}, "A": {1}, "B": {2}},
	}
	tests := []struct {
		name  string
		input []note.Note
	}{
		{name: "specified order", input: notes("D", "E", "F♯", "A", "B")},
		{name: "reverse order and duplicate", input: notes("B", "A", "F♯", "E", "D", "D")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matches, err := searcher.BySpelling(tt.input, search.Containment)
			require.NoError(t, err)
			assert.Equal(t, expected, spellingPairs(matches))
			assert.Equal(t, []string{"D", "E", "F♯", "A", "B"}, spellingStrings(matches[0].InputNames()))
			for _, match := range matches {
				assertSpellingMatch(t, match)
				for _, name := range match.InputNames() {
					assert.Equal(t, expectedDegrees[match.Tonic().String()][name.String()], match.DegreesFor(name))
				}
			}
			independentSpellingMembership(t, searcher, tt.input, search.Containment)
		})
	}
}

func TestBySpellingControlExample2(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian", "dorian", "phrygian", "lydian", "mixolydian", "aeolian", "locrian")
	input := notes("C", "D", "E", "F", "G", "A", "B")
	matches, err := searcher.BySpelling(input, search.FullMatch)
	require.NoError(t, err)
	assert.Equal(t, []spellingPair{
		{"ionian", "C"},
		{"dorian", "D"},
		{"phrygian", "E"},
		{"lydian", "F"},
		{"mixolydian", "G"},
		{"aeolian", "A"},
		{"locrian", "B"},
	}, spellingPairs(matches))
	for _, match := range matches {
		assertSpellingMatch(t, match)
		assert.True(t, match.Tonic().Equal(match.InputNames()[0]) || slicesContainsSpelling(match.InputNames(), match.Tonic()))
	}
	containment := searcher.MustBySpelling(input, search.Containment)
	assert.Subset(t, spellingPairs(containment), spellingPairs(matches))
	assert.Greater(t, len(containment), len(matches))
}

func slicesContainsSpelling(values []spelling.Spelling, wanted spelling.Spelling) bool {
	for _, value := range values {
		if value.Equal(wanted) {
			return true
		}
	}
	return false
}

func TestBySpellingControlExample3(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	matches, err := searcher.BySpelling(notes("C♯", "F"), search.Containment)
	requireNonNilEmpty(t, matches, err)
	matches, err = searcher.BySpelling(notes("C♯", "F"), search.FullMatch)
	requireNonNilEmpty(t, matches, err)
}

func TestBySpellingOpenModeControlExample6(t *testing.T) {
	t.Parallel()
	tmpl := template.MustOpenFromIntervals([]interval.Interval{interval.MustParse("m3"), interval.MustParse("M2")})
	searcher := searcherFor(t, entry("open-tri", tmpl))
	tests := []struct {
		name      string
		input     []string
		condition search.Condition
		expected  []spellingPair
	}{
		{name: "single contained note", input: []string{"G"}, condition: search.Containment, expected: []spellingPair{{"open-tri", "D"}, {"open-tri", "E"}, {"open-tri", "G"}}},
		{name: "two contained notes", input: []string{"E", "G"}, condition: search.Containment, expected: []spellingPair{{"open-tri", "E"}}},
		{name: "full open collection", input: []string{"E", "G", "A"}, condition: search.FullMatch, expected: []spellingPair{{"open-tri", "E"}}},
		{name: "extra input is not full collection", input: []string{"E", "G"}, condition: search.FullMatch, expected: []spellingPair{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matches, err := searcher.BySpelling(notes(tt.input...), tt.condition)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, spellingPairs(matches))
			for _, match := range matches {
				assert.True(t, match.Mode().IsOpen())
				assertSpellingMatch(t, match)
			}
		})
	}
}

func TestBySpellingNonOctaveCycleControlExample7(t *testing.T) {
	t.Parallel()
	tmpl := template.MustClosedFromSpellings([]spelling.Spelling{sp("C"), sp("D"), sp("E")}, interval.MustParse("P4"))
	searcher := searcherFor(t, entry("quarter-cycle", tmpl))
	matches, err := searcher.BySpelling(notes("D", "E"), search.Containment)
	require.NoError(t, err)
	assert.Equal(t, []spellingPair{{"quarter-cycle", "C"}, {"quarter-cycle", "D"}}, spellingPairs(matches))
	matches, err = searcher.BySpelling(notes("C", "D", "E"), search.FullMatch)
	require.NoError(t, err)
	assert.Equal(t, []spellingPair{{"quarter-cycle", "C"}}, spellingPairs(matches))
	for _, match := range matches {
		assert.True(t, match.Mode().IsClosed())
		assertSpellingMatch(t, match)
	}
}

func TestBySpellingMetadataControlExample9(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	plain := searcher.MustBySpelling(notes("D", "E", "F♯", "A", "B"), search.Containment)
	metadata := []note.Note{
		note.MustFromASPN(sp("D"), 4, cents.MustFromInt(30), nil),
		note.MustFromSpelling(sp("E"), cents.Zero(), nil),
		note.MustFromASPN(sp("F♯"), 5, cents.MustFromInt(-10), nil),
		note.MustFromASPN(sp("A"), 3, cents.Zero(), nil),
		note.MustFromASPN(sp("B"), 7, cents.Zero(), nil),
	}
	withMetadata := searcher.MustBySpelling(metadata, search.Containment)
	assert.Equal(t, spellingPairs(plain), spellingPairs(withMetadata))
	for _, match := range withMetadata {
		assert.False(t, match.Tonic().String() == "")
		assertSpellingMatch(t, match)
	}
}

func TestBySpellingRepeatedClassControlExample10(t *testing.T) {
	t.Parallel()
	tmpl := template.MustFromSpellingsInOctave([]spelling.Spelling{sp("C"), sp("C♯"), sp("D♭")})
	searcher := searcherFor(t, entry("repeated-class", tmpl))
	matches, err := searcher.BySpelling(notes("C", "C♯", "D♭"), search.FullMatch)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, []int{2}, matches[0].DegreesFor(sp("C♯")))
	assert.Equal(t, []int{3}, matches[0].DegreesFor(sp("D♭")))
	assert.Nil(t, matches[0].DegreesFor(sp("E")))
	assertSpellingMatch(t, matches[0])
}

func TestBySpellingOriginAndSynonymsControlExample12(t *testing.T) {
	t.Parallel()
	catalog := template.BuiltIn()
	first, ok := catalog.ByCode("hirajoshi-sachs")
	require.True(t, ok)
	second, ok := catalog.ByCode("iwato")
	require.True(t, ok)
	searcher := searcherFor(t, first, second)
	matches, err := searcher.BySpelling(notes("C", "D♭", "F", "G♭", "B♭"), search.FullMatch)
	require.NoError(t, err)
	assert.Equal(t, []spellingPair{{"hirajoshi-sachs", "C"}, {"iwato", "C"}}, spellingPairs(matches))
	require.Len(t, matches, 2)
	firstOrigin, ok := matches[0].Mode().Origin()
	require.True(t, ok)
	secondOrigin, ok := matches[1].Mode().Origin()
	require.True(t, ok)
	assert.Equal(t, first.Names(), firstOrigin.Names())
	assert.Equal(t, second.Names(), secondOrigin.Names())
	assert.True(t, matches[0].Mode().Equal(matches[1].Mode()))
}

func TestBySpellingCandidateLimitsControlExample13(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	matches, err := searcher.BySpelling(notes("C"), search.Containment)
	require.NoError(t, err)
	assert.Equal(t, []spellingPair{
		{"ionian", "C"},
		{"ionian", "D♭"},
		{"ionian", "E♭"},
		{"ionian", "F"},
		{"ionian", "G"},
		{"ionian", "A♭"},
		{"ionian", "B♭"},
	}, spellingPairs(matches))
	assert.LessOrEqual(t, len(matches), matches[0].Mode().Len())
}

func TestBySpellingUnrepresentableCandidateControlExample14(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "phrygian")
	input := notes("E")
	input[0] = note.MustFromSpelling(spelling.MustNew(input[0].Letter(), alt(alteration.MaxMagnitude)), cents.Zero(), nil)
	matches, err := searcher.BySpelling(input, search.Containment)
	require.NoError(t, err)
	maximum := alt(alteration.MaxMagnitude)
	expectedTonics := []spelling.Spelling{
		spelling.MustNew(sp("C").Letter(), maximum),
		input[0].Spelling(),
		spelling.MustNew(sp("A").Letter(), maximum),
		spelling.MustNew(sp("B").Letter(), maximum),
	}
	expected := make([]spellingPair, len(expectedTonics))
	for i, tonic := range expectedTonics {
		expected[i] = spellingPair{Code: "phrygian", Tonic: tonic.String()}
	}
	assert.Equal(t, expected, spellingPairs(matches))
	for _, match := range matches {
		assert.LessOrEqual(t, len(matches), match.Mode().Len())
		assertSpellingMatch(t, match)
	}
}

func TestBySpellingInputAndScopeOwnership(t *testing.T) {
	t.Parallel()
	catalog := template.MustNewCatalog(builtInEntry(t, "ionian"))
	searcher := search.MustNew(&catalog)
	repository := mode.MustNewRepository(&catalog)
	fromRepository := search.MustNewFromRepository(repository)
	input := notes("C")
	before := append([]note.Note(nil), input...)
	expectedNotes := append([]note.Note(nil), before...)
	catalog = template.Catalog{}
	repository = mode.Repository{}
	input[0] = note.MustFromSpelling(sp("D"), cents.Zero(), nil)

	for name, current := range map[string]search.Searcher{"catalog": searcher, "repository": fromRepository} {
		t.Run(name, func(t *testing.T) {
			matches, err := current.BySpelling(before, search.Containment)
			require.NoError(t, err)
			assert.NotEmpty(t, matches)
			assert.Equal(t, "C", matches[0].InputNames()[0].String())
			assertNotesUnchanged(t, expectedNotes, before)
		})
	}
}

func builtInEntry(t testing.TB, code string) template.Entry {
	t.Helper()
	entry, ok := template.BuiltIn().ByCode(code)
	require.True(t, ok)
	return entry
}

func TestIndependentModesAndConcurrentSearches(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian", "aeolian")
	input := notes("D", "E", "F♯", "A", "B")
	expected, err := searcher.BySpelling(input, search.Containment)
	require.NoError(t, err)
	require.NotEmpty(t, expected)
	changed := expected[0].Mode()
	require.NoError(t, changed.Rotate(2))
	assert.False(t, changed.Equal(expected[0].Mode()))
	assert.Equal(t, spellingPairs(expected), spellingPairs(searcher.MustBySpelling(input, search.Containment)))

	const workers = 8
	results := make(chan []spellingPair, workers)
	errors := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			matches, searchErr := searcher.BySpelling(input, search.Containment)
			if searchErr != nil {
				errors <- searchErr
				return
			}
			results <- spellingPairs(matches)
		})
	}
	wg.Wait()
	close(results)
	close(errors)
	for searchErr := range errors {
		assert.NoError(t, searchErr)
	}
	for result := range results {
		assert.Equal(t, spellingPairs(expected), result)
	}
}

func TestSpellingResultSliceOwnership(t *testing.T) {
	t.Parallel()
	searcher := searcherFor(t, onePlaceEntry("one"))
	matches := searcher.MustBySpelling(notes("C"), search.FullMatch)
	require.Len(t, matches, 1)
	match := matches[0]
	expectedNames := spellingStrings(match.InputNames())
	returnedNames := match.InputNames()
	returnedNames[0] = sp("D")
	assert.Equal(t, expectedNames, spellingStrings(match.InputNames()))
	expectedDegrees := match.DegreesFor(sp("C"))
	returnedDegrees := match.DegreesFor(sp("C"))
	returnedDegrees[0] = 99
	assert.Equal(t, expectedDegrees, match.DegreesFor(sp("C")))
	assert.Nil(t, match.DegreesFor(sp("D")))
}

func TestBySpellingDoesNotReadSources(t *testing.T) {
	t.Parallel()
	first := newCountedSource(settings.Standard())
	second := newCountedSource(settingsWithSystem(nonRegularSystem()))
	input := append(notesIn(first, "D", "E"), notesIn(second, "F♯", "A", "B")...)
	first.resetReads()
	second.resetReads()
	searcher := builtInSearcher(t, "ionian")
	matches, err := searcher.BySpelling(input, search.Containment)
	require.NoError(t, err)
	assert.NotEmpty(t, matches)
	assert.Equal(t, 0, first.readCount())
	assert.Equal(t, 0, second.readCount())
}
