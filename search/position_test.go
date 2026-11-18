package search_test

import (
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/search"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestByPositionControlExample3(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	input := notes("C♯", "F")
	matches, err := searcher.ByPosition(system.Traditional(), input, search.Containment)
	require.NoError(t, err)
	require.Len(t, matches, 3)
	assertPositionOrder(t, matches)
	assert.Equal(t, []int{1, 6, 8}, classSet(matches))
	assert.Equal(t, []positionPair{
		{Code: "ionian", Class: 1, Tonic: "C♯"},
		{Code: "ionian", Class: 6, Tonic: "F♯"},
		{Code: "ionian", Class: 8, Tonic: "G♯"},
	}, positionPairs(matches))
	alternatives := []string{"D♭", "G♭", "A♭"}
	for i, match := range matches {
		assert.Equal(t, []int{1, 5}, classNumbers(match.InputClasses()))
		_, hasLetter := match.Family().Letter()
		assert.False(t, hasLetter)
		assert.True(t, match.Family().System().Equal(system.Traditional()))
		assertPositionMatch(t, match)
		spellings, spellingsErr := match.Family().Spellings(alt(1))
		require.NoError(t, spellingsErr)
		values := collectSpellings(spellings, -1)
		assert.Len(t, values, 2)
		assert.Contains(t, spellingStrings(values), alternatives[i])
		assert.Contains(t, spellingStrings(values), match.Tonic().String())
	}
	full, err := searcher.ByPosition(system.Traditional(), input, search.FullMatch)
	requireNonNilEmpty(t, full, err)
	independentPositionMembership(t, searcher, system.Traditional(), input)
	assert.LessOrEqual(t, len(matches), matches[0].Mode().Len())
}

func TestByPositionNonRegularControlExample4(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	matches, err := searcher.ByPosition(nonRegularSystem(), notes("E", "F♯"), search.Containment)
	require.NoError(t, err)
	assert.Len(t, matches, 29)
	assertPositionOrder(t, matches)
	assert.Len(t, classSet(matches), 11)
	found := false
	for _, match := range matches {
		l, ok := match.Family().Letter()
		assert.True(t, ok)
		assert.Equal(t, 12, match.Family().Class().Modulus())
		if l.Equal(letter.D) && match.Family().Class().Number() == 3 {
			found = true
			assert.Equal(t, "D", match.Tonic().String())
		}
		assertPositionMatch(t, match)
		assert.LessOrEqual(t, len(matches), 7*match.Mode().Len())
	}
	assert.True(t, found, "missing (D,3) family")
	assert.Subset(t, classSet(matches), []int{0, 8, 10})
}

func TestByPositionIrregularControlExample5(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	matches, err := searcher.ByPosition(c3System(), notes("C♭", "G♭"), search.Containment)
	require.NoError(t, err)
	expected := []int{1, 2, 3, 4, 6, 8, 9}
	assert.Equal(t, expected, classSet(matches))
	assertPositionOrder(t, matches)
	assert.NotEqual(t, []int{2, 4, 6, 7, 9, 11}, classSet(matches))
	for _, match := range matches {
		_, ok := match.Family().Letter()
		assert.True(t, ok)
		assertPositionMatch(t, match)
	}
}

func TestPositionFormulaOracles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		interval   string
		input      string
		tonic      string
		tonicClass int
	}{
		{name: "minor second from E", interval: "m2", input: "F", tonic: "E", tonicClass: 4},
		{name: "diminished fifth from B", interval: "d5", input: "F", tonic: "B", tonicClass: 11},
		{name: "perfect fourth from F", interval: "P4", input: "B♭", tonic: "F", tonicClass: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpl := template.MustOpenFromIntervals([]interval.Interval{interval.MustParse(tt.interval)})
			searcher := searcherFor(t, entry("formula", tmpl))
			matches, err := searcher.ByPosition(system.Traditional(), notes(tt.input), search.Containment)
			require.NoError(t, err)
			index := slices.IndexFunc(matches, func(match search.PositionMatch) bool {
				_, ok := match.Family().Letter()
				return !ok && match.Family().Class().Number() == tt.tonicClass && match.Tonic().Equal(sp(tt.tonic))
			})
			if assert.NotEqual(t, -1, index) {
				assertPositionMatch(t, matches[index])
			}
		})
	}
}

func TestPositionFormulaNaturalDeviationInNR(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "phrygian")
	matches, err := searcher.ByPosition(nonRegularSystem(), notes("E", "F"), search.Containment)
	require.NoError(t, err)
	index := slices.IndexFunc(matches, func(match search.PositionMatch) bool {
		l, ok := match.Family().Letter()
		return ok && l.Equal(letter.E) && match.Family().Class().Number() == 4 && match.Tonic().Equal(sp("E"))
	})
	if assert.NotEqual(t, -1, index, "phrygian / E must survive in NR") {
		assertPositionMatch(t, matches[index])
	}
}

func TestByPositionMetadataControlExample9(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	plain := searcher.MustByPosition(system.Traditional(), notes("D", "E", "F♯", "A", "B"), search.Containment)
	input := []note.Note{
		note.MustFromASPN(sp("D"), 4, cents.MustFromInt(30), nil),
		note.MustFromSpelling(sp("E"), cents.Zero(), nil),
		note.MustFromASPN(sp("F♯"), 5, cents.MustFromInt(-10), nil),
		note.MustFromASPN(sp("A"), 3, cents.Zero(), nil),
		note.MustFromASPN(sp("B"), 7, cents.Zero(), nil),
	}
	withMetadata := searcher.MustByPosition(system.Traditional(), input, search.Containment)
	assert.Equal(t, positionPairs(plain), positionPairs(withMetadata))
}

func TestByPositionRepeatedClassControlExample10(t *testing.T) {
	t.Parallel()
	tmpl := template.MustFromSpellingsInOctave([]spelling.Spelling{sp("C"), sp("C♯"), sp("D♭")})
	searcher := searcherFor(t, entry("repeated-class", tmpl))
	matches, err := searcher.ByPosition(system.Traditional(), notes("C", "C♯", "D♭"), search.FullMatch)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	match := matches[0]
	assert.Equal(t, 0, match.Family().Class().Number())
	assert.Equal(t, []int{0, 1}, classNumbers(match.InputClasses()))
	assert.Equal(t, []int{1}, match.DegreesForClass(pitchclass.MustNew(0, 12)))
	assert.Equal(t, []int{2, 3}, match.DegreesForClass(pitchclass.MustNew(1, 12)))
	assert.Nil(t, match.DegreesForClass(pitchclass.MustNew(2, 12)))
	assert.Nil(t, match.DegreesForClass(pitchclass.MustNew(1, 19)))
	assertPositionMatch(t, match)
}

func TestByPositionRejectsUninterpretableInput(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	input := notes("C", "D+1/2", "C")
	matches, err := searcher.ByPosition(system.Traditional(), input, search.Containment)
	assert.Nil(t, matches)
	requireError(t, err, search.ErrUninterpretableName, system.ErrInadmissible)
	fractionalSource := newCountedSource(settingsWithSystem(onePlaceSystem("fractional-context", 24, 2, system.MustMultiples(2))))
	contextInput := notesIn(fractionalSource, "C", "D+1/2")
	fractionalSource.current.Store(settings.Standard())
	fractionalSource.resetReads()
	matches, err = searcher.ByPositionFromContext(contextInput, search.FullMatch)
	assert.Nil(t, matches)
	requireError(t, err, search.ErrUninterpretableName, system.ErrInadmissible)
	assert.Equal(t, 1, fractionalSource.readCount())
}

func TestByPositionNaturalOnlyControlExample13(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	matches, err := searcher.ByPosition(naturalOnlySystem(), notes("C"), search.Containment)
	require.NoError(t, err)
	assert.Equal(t, []positionPair{{Code: "ionian", Class: 0, Tonic: "C"}}, positionPairs(matches))
	require.Len(t, matches, 1)
	assertPositionMatch(t, matches[0])
}

func TestByPositionOpenAndNonOctaveModes(t *testing.T) {
	t.Parallel()
	openTemplate := template.MustOpenFromIntervals([]interval.Interval{interval.MustParse("m3"), interval.MustParse("M2")})
	nonOctaveTemplate := template.MustClosedFromSpellings([]spelling.Spelling{sp("C"), sp("D"), sp("E")}, interval.MustParse("P4"))
	tests := []struct {
		name      string
		entry     template.Entry
		input     []string
		condition search.Condition
		expected  []positionPair
	}{
		{name: "open containment", entry: entry("open-tri", openTemplate), input: []string{"E", "G"}, condition: search.Containment, expected: []positionPair{{Code: "open-tri", Class: 4, Tonic: "E"}}},
		{name: "open full match", entry: entry("open-tri", openTemplate), input: []string{"E", "G", "A"}, condition: search.FullMatch, expected: []positionPair{{Code: "open-tri", Class: 4, Tonic: "E"}}},
		{name: "non-octave containment", entry: entry("quarter-cycle", nonOctaveTemplate), input: []string{"D", "E"}, condition: search.Containment, expected: []positionPair{{Code: "quarter-cycle", Class: 0, Tonic: "C"}, {Code: "quarter-cycle", Class: 2, Tonic: "D"}}},
		{name: "non-octave full match", entry: entry("quarter-cycle", nonOctaveTemplate), input: []string{"C", "D", "E"}, condition: search.FullMatch, expected: []positionPair{{Code: "quarter-cycle", Class: 0, Tonic: "C"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			searcher := searcherFor(t, tt.entry)
			matches, err := searcher.ByPosition(system.Traditional(), notes(tt.input...), tt.condition)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, positionPairs(matches))
			for _, match := range matches {
				assertPositionMatch(t, match)
			}
		})
	}
}

func TestEnharmonicNamesDependOnSystem(t *testing.T) {
	t.Parallel()
	searcher := searcherFor(t, onePlaceEntry("one"))
	traditional := searcher.MustByPosition(system.Traditional(), notes("C♯", "D♭"), search.Containment)
	require.Len(t, traditional, 1)
	assert.Equal(t, []int{1}, classNumbers(traditional[0].InputClasses()))

	positions19 := system.Positions19()
	sharp := searcher.MustByPosition(positions19, notes("C♯"), search.Containment)
	flat := searcher.MustByPosition(positions19, notes("D♭"), search.Containment)
	require.Len(t, sharp, 1)
	require.Len(t, flat, 1)
	assert.NotEqual(t, sharp[0].Family().Class().Number(), flat[0].Family().Class().Number())
	both, err := searcher.ByPosition(positions19, notes("C♯", "D♭"), search.Containment)
	requireNonNilEmpty(t, both, err)
}

func TestByPositionFromContextReadsEachSourceOnce(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	first := newCountedSource(settings.Standard())
	second := newCountedSource(settings.Standard())
	input := notesIn(first, "D", "E", "D")
	input = append(input, notesIn(second, "F♯", "A", "B")...)
	expectedNotes := append([]note.Note(nil), input...)
	first.resetReads()
	second.resetReads()

	matches, err := searcher.ByPositionFromContext(input, search.Containment)
	require.NoError(t, err)
	assertNotesUnchanged(t, expectedNotes, input)
	explicit := searcher.MustByPosition(system.Traditional(), input, search.Containment)
	assert.Equal(t, positionPairs(explicit), positionPairs(matches))
	assert.Equal(t, 1, first.readCount())
	assert.Equal(t, 1, second.readCount())

	first.current.Store(settingsWithSystem(nonRegularSystem()))
	second.current.Store(settingsWithSystem(c3System()))
	for _, match := range matches {
		assert.True(t, match.Family().System().Equal(system.Traditional()))
		assertPositionMatch(t, match)
		assert.NotEmpty(t, collectSpellings(match.Family().MustSpellings(alt(1)), 1))
	}
	assert.Equal(t, 1, first.readCount())
	assert.Equal(t, 1, second.readCount())
}

func TestByPositionResultOwnership(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	source := newCountedSource(settings.Standard())
	input := []note.Note{note.MustFromASPN(sp("C"), 4, cents.MustFromInt(30), source)}
	expectedNotes := append([]note.Note(nil), input...)
	source.resetReads()
	matches := searcher.MustByPosition(system.Traditional(), input, search.Containment)
	require.NotEmpty(t, matches)
	assertNotesUnchanged(t, expectedNotes, input)
	assert.Equal(t, 0, source.readCount())
	match := matches[0]
	expectedInputs := classNumbers(match.InputClasses())
	inputClasses := match.InputClasses()
	inputClasses[0] = pitchclass.MustNew(11, 12)
	assert.Equal(t, expectedInputs, classNumbers(match.InputClasses()))
	class := match.InputClasses()[0]
	expectedDegrees := match.DegreesForClass(class)
	degrees := match.DegreesForClass(class)
	degrees[0] = 99
	assert.Equal(t, expectedDegrees, match.DegreesForClass(class))

	changed := match.Mode()
	require.NoError(t, changed.Rotate(2))
	assert.False(t, changed.Equal(match.Mode()))
	assert.Equal(t, "C", match.Tonic().String())
}

func TestConcurrentPositionSearches(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian", "aeolian")
	input := notes("D", "E", "F♯", "A", "B")
	expected := searcher.MustByPosition(system.Traditional(), input, search.Containment)
	const workers = 8
	results := make(chan []positionPair, workers)
	errors := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			matches, err := searcher.ByPosition(system.Traditional(), input, search.Containment)
			if err != nil {
				errors <- err
				return
			}
			results <- positionPairs(matches)
		})
	}
	wg.Wait()
	close(results)
	close(errors)
	for err := range errors {
		assert.NoError(t, err)
	}
	for result := range results {
		assert.Equal(t, positionPairs(expected), result)
	}
}

func TestPositionInputValidationPrecedesDuplicates(t *testing.T) {
	t.Parallel()
	searcher := searcherFor(t, onePlaceEntry("one"))
	matches, err := searcher.ByPosition(system.Traditional(), notes("D+1/2", "C", "D+1/2"), search.Containment)
	assert.Nil(t, matches)
	requireError(t, err, search.ErrUninterpretableName, system.ErrInadmissible)
}

func TestPositionResultUsesNoInputAlterationMetadata(t *testing.T) {
	t.Parallel()
	searcher := searcherFor(t, onePlaceEntry("one"))
	matches := searcher.MustByPosition(system.Traditional(), notes("C"), search.FullMatch)
	require.Len(t, matches, 1)
	assert.True(t, matches[0].Tonic().Alteration().Equal(alteration.Natural()))
}

func TestByPositionOrderAndDuplicates(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	plain := searcher.MustByPosition(system.Traditional(), notes("D", "E", "F♯", "A", "B"), search.Containment)
	reordered := searcher.MustByPosition(system.Traditional(), notes("B", "A", "F♯", "E", "D", "D"), search.Containment)
	assert.Equal(t, positionPairs(plain), positionPairs(reordered))
	for i := range plain {
		assert.Equal(t, classNumbers(plain[i].InputClasses()), classNumbers(reordered[i].InputClasses()))
		for _, class := range plain[i].InputClasses() {
			assert.Equal(t, plain[i].DegreesForClass(class), reordered[i].DegreesForClass(class))
		}
	}
}

func TestApplicableSpellingResultsAppearInPositionSearch(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	input := notes("C")
	bySpelling := searcher.MustBySpelling(input, search.Containment)
	byPosition := searcher.MustByPosition(system.Traditional(), input, search.Containment)
	for _, spellingMatch := range bySpelling {
		class, err := system.Traditional().Class(spellingMatch.Tonic())
		require.NoError(t, err)
		l := spellingMatch.Tonic().Letter()
		found := slices.ContainsFunc(byPosition, func(positionMatch search.PositionMatch) bool {
			familyLetter, hasLetter := positionMatch.Family().Letter()
			letterMatches := !hasLetter || familyLetter.Equal(l)
			return positionMatch.Family().Class().Equal(class) && letterMatches
		})
		assert.True(t, found, "missing spelling tonic %s", spellingMatch.Tonic())
	}
}
