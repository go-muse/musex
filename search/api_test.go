package search_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/mode"
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

func TestCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition search.Condition
		valid     bool
		text      string
	}{
		{name: "unset", condition: search.ConditionUnset, text: "unset"},
		{name: "containment", condition: search.Containment, valid: true, text: "containment"},
		{name: "full match", condition: search.FullMatch, valid: true, text: "full-match"},
		{name: "unknown", condition: search.Condition(9), text: "Condition(9)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.valid, tt.condition.IsValid())
			assert.Equal(t, tt.text, tt.condition.String())
		})
	}
}

func TestSearcherConstructionAndEmptyScope(t *testing.T) {
	t.Parallel()
	assert.False(t, (search.Searcher{}).IsValid())
	assert.True(t, search.BuiltIn().IsValid())

	emptyCatalog := template.Catalog{}
	empty := search.MustNew(&emptyCatalog)
	assert.True(t, empty.IsValid())
	emptySpelling, err := empty.BySpelling(notes("C"), search.Containment)
	requireNonNilEmpty(t, emptySpelling, err)
	emptyPosition, err := empty.ByPosition(system.Traditional(), notes("C"), search.FullMatch)
	requireNonNilEmpty(t, emptyPosition, err)
	source := newCountedSource(settings.Standard())
	contextInput := notesIn(source, "C")
	source.resetReads()
	emptyContext, err := empty.ByPositionFromContext(contextInput, search.Containment)
	requireNonNilEmpty(t, emptyContext, err)
	assert.Equal(t, 1, source.readCount())

	repository := mode.MustNewRepository(&emptyCatalog)
	fromRepository := search.MustNewFromRepository(repository)
	assert.True(t, fromRepository.IsValid())
	emptySpelling, err = fromRepository.BySpelling(notes("C"), search.Containment)
	requireNonNilEmpty(t, emptySpelling, err)

	_, err = search.New(nil)
	requireError(t, err, search.ErrUnset)
	requireMustPanic(t, err, func() { search.MustNew(nil) })
	_, err = search.NewFromRepository(mode.Repository{})
	requireError(t, err, search.ErrUnset, mode.ErrUnset)
	requireMustPanic(t, err, func() { search.MustNewFromRepository(mode.Repository{}) })
}

func TestZeroResults(t *testing.T) {
	t.Parallel()
	var spellingMatch search.SpellingMatch
	assert.False(t, spellingMatch.IsValid())
	assert.False(t, spellingMatch.Mode().IsValid())
	assert.False(t, spellingMatch.Tonic().IsValid())
	assert.Nil(t, spellingMatch.InputNames())
	assert.Nil(t, spellingMatch.DegreesFor(sp("C")))

	var positionMatch search.PositionMatch
	assert.False(t, positionMatch.IsValid())
	assert.False(t, positionMatch.Mode().IsValid())
	assert.False(t, positionMatch.Tonic().IsValid())
	assert.False(t, positionMatch.Family().IsValid())
	assert.Nil(t, positionMatch.InputClasses())
	assert.Nil(t, positionMatch.DegreesForClass(pitchclass.MustNew(0, 12)))

	var family search.TonicFamily
	assert.False(t, family.IsValid())
	assert.Nil(t, family.System())
	assert.False(t, family.Class().IsValid())
	l, ok := family.Letter()
	assert.False(t, ok)
	assert.False(t, l.IsValid())
	seq, err := family.Spellings(alteration.Natural())
	requireError(t, err, search.ErrUnset)
	assert.Nil(t, seq)
	requireMustPanic(t, err, func() { family.MustSpellings(alteration.Natural()) })
}

func TestSearchErrorPrecedence(t *testing.T) {
	t.Parallel()
	validNotes := notes("C")
	invalidCondition := search.Condition(99)

	tests := []struct {
		name     string
		call     func(search.Searcher) error
		expected []error
	}{
		{
			name: "spelling rejects zero searcher before empty input",
			call: func(s search.Searcher) error {
				_, err := s.BySpelling(nil, invalidCondition)
				return err
			},
			expected: []error{search.ErrUnset},
		},
		{
			name: "explicit position rejects zero searcher before empty input",
			call: func(s search.Searcher) error {
				_, err := s.ByPosition(nil, nil, invalidCondition)
				return err
			},
			expected: []error{search.ErrUnset},
		},
		{
			name: "context position rejects zero searcher before empty input",
			call: func(s search.Searcher) error {
				_, err := s.ByPositionFromContext(nil, invalidCondition)
				return err
			},
			expected: []error{search.ErrUnset},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireError(t, tt.call(search.Searcher{}), tt.expected...)
		})
	}

	searcher := builtInSearcher(t, "ionian")
	operations := []struct {
		name string
		call func([]note.Note, search.Condition) error
	}{
		{name: "spelling", call: func(ns []note.Note, c search.Condition) error {
			_, err := searcher.BySpelling(ns, c)
			return err
		}},
		{name: "explicit position", call: func(ns []note.Note, c search.Condition) error {
			_, err := searcher.ByPosition(system.Traditional(), ns, c)
			return err
		}},
		{name: "context position", call: func(ns []note.Note, c search.Condition) error {
			_, err := searcher.ByPositionFromContext(ns, c)
			return err
		}},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			t.Run("nil and empty precede note and condition", func(t *testing.T) {
				for _, input := range [][]note.Note{nil, {}} {
					requireError(t, operation.call(input, invalidCondition), search.ErrEmptyInput)
					assert.NotErrorIs(t, operation.call(input, invalidCondition), search.ErrInvalidArgument)
				}
			})
			t.Run("first unset note precedes condition and system", func(t *testing.T) {
				input := []note.Note{validNotes[0], {}, validNotes[0]}
				requireError(t, operation.call(input, invalidCondition), search.ErrUnset, note.ErrUnset)
				assert.NotErrorIs(t, operation.call(input, invalidCondition), search.ErrInvalidArgument)
			})
			t.Run("condition precedes system", func(t *testing.T) {
				requireError(t, operation.call(validNotes, invalidCondition), search.ErrInvalidArgument)
			})
		})
	}

	result, err := searcher.ByPosition(nil, validNotes, search.Containment)
	assert.Nil(t, result)
	requireError(t, err, search.ErrUnset, system.ErrInvalid)
	result, err = searcher.ByPosition(&system.System{}, validNotes, search.Containment)
	assert.Nil(t, result)
	requireError(t, err, search.ErrUnset, system.ErrInvalid)
}

func TestSearchMustWrappers(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	validNotes := notes("C")
	tests := []struct {
		name string
		call func() error
		must func()
	}{
		{
			name: "by spelling",
			call: func() error {
				_, err := searcher.BySpelling(unsetNotesForTest(), search.Containment)
				return err
			},
			must: func() { searcher.MustBySpelling(unsetNotesForTest(), search.Containment) },
		},
		{
			name: "by position",
			call: func() error {
				_, err := searcher.ByPosition(system.Traditional(), validNotes, search.Condition(8))
				return err
			},
			must: func() { searcher.MustByPosition(system.Traditional(), validNotes, search.Condition(8)) },
		},
		{
			name: "by position from context",
			call: func() error {
				_, err := searcher.ByPositionFromContext(validNotes, search.Containment)
				return err
			},
			must: func() { searcher.MustByPositionFromContext(validNotes, search.Containment) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireMustPanic(t, tt.call(), tt.must)
		})
	}
}

func unsetNotesForTest() []note.Note {
	return []note.Note{{}}
}

func TestExplicitPositionDoesNotReadSources(t *testing.T) {
	t.Parallel()
	traditional := newCountedSource(settings.Standard())
	other := newCountedSource(settingsWithSystem(nonRegularSystem()))
	input := append(notesIn(traditional, "C♯"), notesIn(other, "D♭")...)
	traditional.resetReads()
	other.resetReads()

	searcher := builtInSearcher(t, "ionian")
	matches, err := searcher.ByPosition(system.Traditional(), input, search.Containment)
	require.NoError(t, err)
	require.NotEmpty(t, matches)
	assert.Equal(t, 0, traditional.readCount())
	assert.Equal(t, 0, other.readCount())
	assert.Equal(t, []int{1}, classNumbers(matches[0].InputClasses()))
}

func TestContextErrorsAndCauses(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")

	matches, err := searcher.ByPositionFromContext(notes("C"), search.Containment)
	assert.Nil(t, matches)
	requireError(t, err, search.ErrNoSystem, note.ErrNoSource)

	emptySource := newCountedSource(settings.Standard())
	input := notesIn(emptySource, "C")
	emptySource.current.Store(nil)
	emptySource.resetReads()
	matches, err = searcher.ByPositionFromContext(input, search.Containment)
	assert.Nil(t, matches)
	requireError(t, err, search.ErrContext, note.ErrUnsetSettings, settings.ErrUnset)
	assert.Equal(t, 1, emptySource.readCount())

	first := newCountedSource(settings.Standard())
	second := newCountedSource(settingsWithSystem(nonRegularSystem()))
	input = append(notesIn(first, "C"), notesIn(second, "D")...)
	first.resetReads()
	second.resetReads()
	matches, err = searcher.ByPositionFromContext(input, search.Containment)
	assert.Nil(t, matches)
	requireError(t, err, search.ErrDifferentSystems, note.ErrDifferentSystems)
	assert.Equal(t, 1, first.readCount())
	assert.Equal(t, 1, second.readCount())
}

func TestNegativeSpellingsLimit(t *testing.T) {
	t.Parallel()
	searcher := searcherFor(t, onePlaceEntry("one"))
	matches, err := searcher.ByPosition(system.Traditional(), notes("C"), search.Containment)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	seq, err := matches[0].Family().Spellings(alt(-1))
	assert.Nil(t, seq)
	requireError(t, err, search.ErrInvalidArgument, system.ErrLimit)
	requireMustPanic(t, err, func() { matches[0].Family().MustSpellings(alt(-1)) })
}

func TestResultCopiesAreReadable(t *testing.T) {
	t.Parallel()
	searcher := builtInSearcher(t, "ionian")
	spellingMatches := searcher.MustBySpelling(notes("C"), search.Containment)
	positionMatches := searcher.MustByPosition(system.Traditional(), notes("C"), search.Containment)
	require.NotEmpty(t, spellingMatches)
	require.NotEmpty(t, positionMatches)
	spellingCopy := spellingMatches[0]
	positionCopy := positionMatches[0]
	familyCopy := positionCopy.Family()
	assert.Equal(t, spellingPairs(spellingMatches[:1]), spellingPairs([]search.SpellingMatch{spellingCopy}))
	assert.Equal(t, positionPairs(positionMatches[:1]), positionPairs([]search.PositionMatch{positionCopy}))
	assert.True(t, familyCopy.System().Equal(positionCopy.Family().System()))
	assert.Equal(t, []spelling.Spelling{positionCopy.Tonic()}, collectSpellings(familyCopy.MustSpellings(alteration.Natural()), 1))
}
