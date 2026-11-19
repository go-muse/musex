package note_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/search"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Architecture 6.2: every one of the 42 old function types is unchanged.
// Assignments to explicit function types reject a defined Notes parameter or
// result even where an ordinary call would still compile.
var (
	_ func([]note.Note) ([]note.Note, error)                  = note.SortedByPosition
	_ func([]note.Note) []note.Note                           = note.MustSortedByPosition
	_ func(settings.Source, []note.Note) ([]note.Note, error) = note.SortedByPositionIn
	_ func(settings.Source, []note.Note) []note.Note          = note.MustSortedByPositionIn
	_ func(*tuning.Tuning, []note.Note) ([]note.Note, error)  = note.SortedByPositionInTuning
	_ func(*tuning.Tuning, []note.Note) []note.Note           = note.MustSortedByPositionInTuning
	_ func([]note.Note) ([]note.Note, error)                  = note.SortedBySound
	_ func([]note.Note) []note.Note                           = note.MustSortedBySound
	_ func(settings.Source, []note.Note) ([]note.Note, error) = note.SortedBySoundIn
	_ func(settings.Source, []note.Note) []note.Note          = note.MustSortedBySoundIn
	_ func(*tuning.Tuning, []note.Note) ([]note.Note, error)  = note.SortedBySoundInTuning
	_ func(*tuning.Tuning, []note.Note) []note.Note           = note.MustSortedBySoundInTuning

	_ func([]note.Note, interval.Interval) (mode.Mode, error)         = mode.ClosedFromNotes
	_ func([]note.Note, interval.Interval) mode.Mode                  = mode.MustClosedFromNotes
	_ func([]note.Note) (mode.Mode, error)                            = mode.FromNotesInOctave
	_ func([]note.Note) mode.Mode                                     = mode.MustFromNotesInOctave
	_ func([]note.Note) (mode.Mode, error)                            = mode.OpenFromNotes
	_ func([]note.Note) mode.Mode                                     = mode.MustOpenFromNotes
	_ func(mode.Mode, note.Note) ([]note.Note, error)                 = mode.Mode.NotesFrom
	_ func(mode.Mode, note.Note) []note.Note                          = mode.Mode.MustNotesFrom
	_ func(mode.Mode, *system.System, note.Note) ([]note.Note, error) = mode.Mode.NotesFromIn
	_ func(mode.Mode, *system.System, note.Note) []note.Note          = mode.Mode.MustNotesFromIn

	_ func(*scale.Scale) []note.Note                                                  = (*scale.Scale).Notes
	_ func(*scale.Scale, []note.Note, int, ...scale.Comparison) (*scale.Scale, error) = (*scale.Scale).ShiftMelody
	_ func(*scale.Scale, []note.Note, int, ...scale.Comparison) *scale.Scale          = (*scale.Scale).MustShiftMelody
	_ func(*scale.Scale, []note.Note, int, ...scale.Comparison) (*scale.Scale, error) = (*scale.Scale).MirrorMelody
	_ func(*scale.Scale, []note.Note, int, ...scale.Comparison) *scale.Scale          = (*scale.Scale).MustMirrorMelody

	_ func(chord.Chord) []note.Note                           = chord.Chord.Notes
	_ func(chord.Chord) []note.Note                           = chord.Chord.SortedNotes
	_ func(chord.Chord) []note.Note                           = chord.Chord.DistinctNotes
	_ func(chord.Chord, chord.Condition) ([]note.Note, error) = chord.Chord.NotesFunc
	_ func(chord.Chord, chord.Condition) []note.Note          = chord.Chord.MustNotesFunc

	_ func(event.Content) ([]note.Note, error) = event.Content.Notes
	_ func(event.Content) []note.Note          = event.Content.MustNotes
	_ func(*event.Event) ([]note.Note, error)  = (*event.Event).TiedNotes
	_ func(*event.Event) []note.Note           = (*event.Event).MustTiedNotes

	_ func(search.Searcher, []note.Note, search.Condition) ([]search.SpellingMatch, error)                 = search.Searcher.BySpelling
	_ func(search.Searcher, []note.Note, search.Condition) []search.SpellingMatch                          = search.Searcher.MustBySpelling
	_ func(search.Searcher, *system.System, []note.Note, search.Condition) ([]search.PositionMatch, error) = search.Searcher.ByPosition
	_ func(search.Searcher, *system.System, []note.Note, search.Condition) []search.PositionMatch          = search.Searcher.MustByPosition
	_ func(search.Searcher, []note.Note, search.Condition) ([]search.PositionMatch, error)                 = search.Searcher.ByPositionFromContext
	_ func(search.Searcher, []note.Note, search.Condition) []search.PositionMatch                          = search.Searcher.MustByPositionFromContext

	// The eighteen old variadic APIs keep their element type as well.
	_ func(...note.Note) (*scale.Scale, error)           = scale.New
	_ func(...note.Note) *scale.Scale                    = scale.MustNew
	_ func(*scale.Scale, ...note.Note) error             = (*scale.Scale).Replace
	_ func(*scale.Scale, ...note.Note) *scale.Scale      = (*scale.Scale).MustReplace
	_ func(*scale.Scale, int, ...note.Note) error        = (*scale.Scale).Insert
	_ func(*scale.Scale, int, ...note.Note) *scale.Scale = (*scale.Scale).MustInsert
	_ func(*scale.Scale, ...note.Note) error             = (*scale.Scale).Append
	_ func(*scale.Scale, ...note.Note) *scale.Scale      = (*scale.Scale).MustAppend
	_ func(...note.Note) (chord.Chord, error)            = chord.New
	_ func(...note.Note) chord.Chord                     = chord.MustNew
	_ func(chord.Chord, ...note.Note) bool               = chord.Chord.ContainsAll
	_ func(chord.Chord, ...note.Note) bool               = chord.Chord.ContainsAny
	_ func(*chord.Chord, ...note.Note) error             = (*chord.Chord).Add
	_ func(*chord.Chord, ...note.Note) *chord.Chord      = (*chord.Chord).MustAdd
	_ func(*chord.Chord, ...note.Note) error             = (*chord.Chord).AddMissing
	_ func(*chord.Chord, ...note.Note) *chord.Chord      = (*chord.Chord).MustAddMissing
	_ func(*chord.Chord, int, ...note.Note) error        = (*chord.Chord).Insert
	_ func(*chord.Chord, int, ...note.Note) *chord.Chord = (*chord.Chord).MustInsert
)

// Client interfaces exercise the preserved method sets, not private storage.
type legacyNotesProvider interface{ Notes() []note.Note }
type legacyContentNotes interface {
	Notes() ([]note.Note, error)
	MustNotes() []note.Note
}
type legacyTiedNotes interface {
	TiedNotes() ([]note.Note, error)
	MustTiedNotes() []note.Note
}

var (
	_ legacyNotesProvider = (*scale.Scale)(nil)
	_ legacyNotesProvider = chord.Chord{}
	_ legacyContentNotes  = event.Content{}
	_ legacyTiedNotes     = (*event.Event)(nil)
)

func TestNotesContainerOwnership(t *testing.T) {
	t.Parallel()
	src := newCounting(std())
	c := mk(t, spelling.C(), 4, ct(17), src)
	d := mkName(t, spelling.DFlat(), ct(-9), nil)
	input := note.Notes{c, d, c}
	want := []note.Note{c, d, c}
	reads := src.reads()
	s, err := scale.New(input...)
	require.NoError(t, err)
	ch, err := chord.New(input...)
	require.NoError(t, err)
	ss, cs := s.MustIter(0, s.Len(), scale.Forward), ch.Values()
	input[0].MustSetCents(ct(33))
	input[1] = c
	for _, got := range [][]note.Note{s.Notes(), ch.Notes()} {
		assert.IsType(t, []note.Note(nil), any(got))
		assertNotesPreserved(t, want, got)
		got[0].MustSetCents(ct(44))
		got[1] = note.Note{}
	}
	assertNotesPreserved(t, want, s.Notes())
	assertNotesPreserved(t, want, ch.Notes())
	for _, tt := range []struct {
		result []note.Note
		want   []note.Note
	}{
		{ch.SortedNotes(), []note.Note{d, c, c}},
		{ch.DistinctNotes(), []note.Note{c, d}},
		{ch.MustNotesFunc(func(note.Note) (bool, error) { return true, nil }), want},
	} {
		assert.IsType(t, []note.Note(nil), any(tt.result))
		assertNotesPreserved(t, tt.want, tt.result)
		tt.result[0] = note.Note{}
		assertNotesPreserved(t, want, ch.Notes())
	}
	scopy, ccopy := *s, ch
	require.NoError(t, scopy.Set(0, d))
	require.NoError(t, ccopy.ReplaceAt(0, d))
	assertNotesPreserved(t, want, s.Notes())
	assertNotesPreserved(t, want, ch.Notes())
	require.NoError(t, s.Clear())
	require.NoError(t, ch.Clear())
	assertNotesPreserved(t, want, slices.Collect(ss))
	assertNotesPreserved(t, want, slices.Collect(cs))
	assertNotesPreserved(t, []note.Note{d, d, c}, scopy.Notes())
	assertNotesPreserved(t, []note.Note{d, d, c}, ccopy.Notes())
	assert.ErrorIs(t, scopy.Append(c, note.Note{}), scale.ErrUnset)
	assert.ErrorIs(t, ccopy.Add(c, note.Note{}), chord.ErrUnset)
	assertNotesPreserved(t, []note.Note{d, d, c}, scopy.Notes())
	assertNotesPreserved(t, []note.Note{d, d, c}, ccopy.Notes())
	assert.Equal(t, reads, src.reads(), "container operations do not consult note sources")
}

func TestNotesScaleMelodyCompatibility(t *testing.T) {
	t.Parallel()
	src := newCounting(std())
	c := mk(t, spelling.C(), 4, ct(17), src)
	d := mk(t, spelling.D(), 4, ct(-9), src)
	e := mk(t, spelling.E(), 4, ct(3), src)
	s := scale.MustNew(c, d, e)
	melody := note.Notes{e, c, e}
	before := slices.Clone([]note.Note(melody))
	reads := src.reads()
	for _, tt := range []struct {
		name string
		run  func() (*scale.Scale, error)
		want []note.Note
	}{
		{"shift", func() (*scale.Scale, error) { return s.ShiftMelody(melody, 1) }, []note.Note{c, d, c}},
		{"Must shift", func() (*scale.Scale, error) { return s.MustShiftMelody(melody, 1), nil }, []note.Note{c, d, c}},
		{"mirror", func() (*scale.Scale, error) { return s.MirrorMelody(melody, 0) }, []note.Note{d, c, d}},
		{"Must mirror", func() (*scale.Scale, error) { return s.MustMirrorMelody(melody, 0), nil }, []note.Note{d, c, d}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.run()
			require.NoError(t, err)
			assertNotesPreserved(t, tt.want, got.Notes())
			require.NoError(t, got.Clear())
			assertNotesPreserved(t, []note.Note{c, d, e}, s.Notes())
			assertNotesPreserved(t, before, melody)
		})
	}
	assert.Equal(t, reads, src.reads(), "default record matching does not consult sources")
}

func TestNotesVariadicBoundaries(t *testing.T) {
	t.Parallel()
	c, d := mkName(t, spelling.C(), ct(0), nil), mkName(t, spelling.D(), ct(0), nil)
	input := note.Notes{c, d}
	s := scale.MustNew(input...)
	require.NoError(t, s.Replace(input...))
	assert.Same(t, s, s.MustReplace(input...))
	require.NoError(t, s.Insert(0, input...))
	assert.Same(t, s, s.MustInsert(0, input...))
	require.NoError(t, s.Append(input...))
	assert.Same(t, s, s.MustAppend(input...))
	assertNotesPreserved(t, []note.Note{c, d, c, d, c, d, c, d, c, d}, s.Notes())
	ch := chord.MustNew(input...)
	assert.True(t, ch.ContainsAll(input...))
	assert.True(t, ch.ContainsAny(input...))
	require.NoError(t, ch.Add(input...))
	assert.Same(t, &ch, ch.MustAdd(input...))
	require.NoError(t, ch.AddMissing(input...))
	assert.Same(t, &ch, ch.MustAddMissing(input...))
	require.NoError(t, ch.Insert(0, input...))
	assert.Same(t, &ch, ch.MustInsert(0, input...))
	assertNotesPreserved(t, []note.Note{c, d, c, d, c, d, c, d, c, d}, ch.Notes())
	for _, empty := range []note.Notes{nil, {}} {
		assert.NotNil(t, scale.MustNew(empty...).Notes())
		assert.NotNil(t, chord.MustNew(empty...).Notes())
	}
}

func TestNotesLegacyEmptyStatesAndTypes(t *testing.T) {
	t.Parallel()
	var absent *scale.Scale
	assert.Nil(t, absent.Notes())
	assert.Nil(t, absent.Clone())
	assert.NotNil(t, new(scale.Scale).Notes())
	assert.NotNil(t, new(scale.Scale).Clone().Notes())
	var unset chord.Chord
	assert.False(t, unset.IsValid())
	empty := chord.MustNew()
	assert.True(t, empty.IsValid())
	for _, ch := range []chord.Chord{unset, empty} {
		for _, got := range [][]note.Note{ch.Notes(), ch.SortedNotes(), ch.DistinctNotes(), ch.MustNotesFunc(func(note.Note) (bool, error) { return true, nil })} {
			assert.IsType(t, []note.Note(nil), any(got))
			assert.NotNil(t, got)
			assert.Empty(t, got)
		}
	}
	for _, run := range []func([]note.Note) ([]note.Note, error){note.SortedByPosition, note.SortedBySound} {
		got, err := run(note.Notes(nil))
		require.NoError(t, err)
		assert.IsType(t, []note.Note(nil), any(got))
		assert.NotNil(t, got)
	}
	// Empty sort results must not retain a writable window into a live list.
	n := mkName(t, spelling.C(), ct(0), nil)
	backing := note.Notes{n}
	for _, method := range notesSortCases() {
		got, err := method.run(backing[:0], std(), tuning.Standard())
		require.NoError(t, err)
		got = append(got, note.Note{})
		assert.Len(t, got, 1)
		assertNotesPreserved(t, []note.Note{n}, backing)
	}
}

func TestNotesModeCompatibility(t *testing.T) {
	t.Parallel()
	src := newCounting(std())
	var input note.Notes
	for _, name := range []spelling.Spelling{spelling.C(), spelling.D(), spelling.E(), spelling.F(), spelling.G(), spelling.A(), spelling.B()} {
		input = append(input, mk(t, name, 4, ct(17), src))
	}
	want := slices.Clone([]note.Note(input))
	reads := src.reads()
	closed, err := mode.ClosedFromNotes(input, interval.MustParse("P8"))
	require.NoError(t, err)
	inOctave, err := mode.FromNotesInOctave(input)
	require.NoError(t, err)
	open, err := mode.OpenFromNotes(input)
	require.NoError(t, err)
	ionian := mode.MustParse("M2 M2 m2 M2 M2 M2 m2")
	assert.True(t, closed.Equal(ionian))
	assert.True(t, inOctave.Equal(ionian))
	assert.True(t, open.Equal(mode.MustParse("M2 M2 m2 M2 M2 M2 |")))
	assert.True(t, mode.MustClosedFromNotes(input, interval.MustParse("P8")).Equal(closed))
	assert.True(t, mode.MustFromNotesInOctave(input).Equal(inOctave))
	assert.True(t, mode.MustOpenFromNotes(input).Equal(open))
	assert.Equal(t, reads, src.reads(), "mode factories read spellings only")
	first := input[0]
	input[1] = note.Note{}
	for _, run := range []func() ([]note.Note, error){
		func() ([]note.Note, error) { return closed.NotesFrom(first) },
		func() ([]note.Note, error) { return closed.NotesFromIn(system.Traditional(), first) },
		func() ([]note.Note, error) { return closed.MustNotesFrom(first), nil },
		func() ([]note.Note, error) { return closed.MustNotesFromIn(system.Traditional(), first), nil },
	} {
		got, err := run()
		require.NoError(t, err)
		assert.IsType(t, []note.Note(nil), any(got))
		assertNotesPreserved(t, want, got)
		got[0].MustSetCents(ct(90))
		assert.True(t, first.Cents().Equal(ct(17)))
	}
	for _, run := range []func() (*scale.Scale, error){
		func() (*scale.Scale, error) { return closed.FromTonic(first) },
		func() (*scale.Scale, error) { return closed.FromTonicIn(system.Traditional(), first) },
	} {
		built, err := run()
		require.NoError(t, err)
		assertNotesPreserved(t, want, built.Notes())
		require.NoError(t, built.Set(0, note.MustFromSpelling(spelling.G(), ct(0), nil)))
		assertNotesPreserved(t, want, closed.MustFromTonic(first).Notes())
	}
}

func TestNotesEventCompatibility(t *testing.T) {
	t.Parallel()
	src := newCounting(std())
	c, e := mk(t, spelling.C(), 4, ct(17), src), mk(t, spelling.E(), 4, ct(-3), src)
	input := note.Notes{c, e, c}
	ch := chord.MustNew(input...)
	content := event.MustChordContent(ch)
	ev := event.MustNew(content, event.MustValueDuration(fraction.MustNew(1, 4)), event.Marks{Ties: []int{2, 0, 2}})
	reads := src.reads()
	input[0] = note.Note{}
	require.NoError(t, ch.Clear())
	for _, got := range [][]note.Note{content.MustNotes(), ev.MustContent().MustNotes()} {
		assert.IsType(t, []note.Note(nil), any(got))
		assertNotesPreserved(t, []note.Note{c, e, c}, got)
		got[0].MustSetCents(ct(40))
		got[1] = note.Note{}
	}
	tied, err := ev.TiedNotes()
	require.NoError(t, err)
	assert.IsType(t, []note.Note(nil), any(tied))
	assertNotesPreserved(t, []note.Note{c, c}, tied)
	tied[0] = note.Note{}
	assertNotesPreserved(t, []note.Note{c, c}, ev.MustTiedNotes())
	assertNotesPreserved(t, []note.Note{c, e, c}, content.MustNotes())
	assertNotesPreserved(t, []note.Note{c}, event.MustNoteContent(c).MustNotes())
	assert.Equal(t, reads, src.reads())
	rest := event.MustNew(event.RestContent(), event.MustValueDuration(fraction.MustNew(1, 4)), event.Marks{})
	for _, got := range [][]note.Note{rest.MustContent().MustNotes(), rest.MustTiedNotes()} {
		assert.NotNil(t, got)
		assert.Empty(t, got)
	}
	got, err := (event.Content{}).Notes()
	assert.ErrorIs(t, err, event.ErrUnsetContent)
	assert.NotErrorIs(t, err, event.ErrUnset)
	assert.Nil(t, got)
}

func TestNotesSearchCompatibility(t *testing.T) {
	t.Parallel()
	builtIn := template.BuiltIn()
	entry, ok := builtIn.ByCode("ionian")
	require.True(t, ok)
	catalog := template.MustNewCatalog(entry)
	searcher := search.MustNew(&catalog)
	a, b := newCounting(std()), newCounting(std())
	var input note.Notes
	for i, name := range []spelling.Spelling{spelling.C(), spelling.D(), spelling.E(), spelling.F(), spelling.G(), spelling.A(), spelling.B(), spelling.C()} {
		src := a
		if i%2 == 1 {
			src = b
		}
		input = append(input, mkName(t, name, ct(i), src))
	}
	want := slices.Clone([]note.Note(input))
	ar, br := a.reads(), b.reads()
	spelled, err := searcher.BySpelling(input, search.FullMatch)
	require.NoError(t, err)
	require.Len(t, spelled, 1)
	assert.True(t, spelled[0].Tonic().Equal(spelling.C()))
	positions, err := searcher.ByPosition(system.Traditional(), input, search.FullMatch)
	require.NoError(t, err)
	require.Len(t, positions, 1)
	assert.Zero(t, positions[0].Family().Class().Number())
	assert.Equal(t, ar, a.reads())
	assert.Equal(t, br, b.reads())
	contextual, err := searcher.ByPositionFromContext(input, search.FullMatch)
	require.NoError(t, err)
	require.Len(t, contextual, 1)
	assert.Zero(t, contextual[0].Family().Class().Number())
	assert.Equal(t, ar+1, a.reads())
	assert.Equal(t, br+1, b.reads())
	assertNotesPreserved(t, want, input)
	assert.Len(t, searcher.MustBySpelling(input, search.FullMatch), 1)
	assert.Len(t, searcher.MustByPosition(system.Traditional(), input, search.FullMatch), 1)
	assert.Len(t, searcher.MustByPositionFromContext(input, search.FullMatch), 1)
	assertNotesPreserved(t, want, input)
	for _, empty := range []note.Notes{nil, {}} {
		spelled, err := searcher.BySpelling(empty, search.FullMatch)
		assert.ErrorIs(t, err, search.ErrEmptyInput)
		assert.Nil(t, spelled)
		positions, err := searcher.ByPosition(system.Traditional(), empty, search.FullMatch)
		assert.ErrorIs(t, err, search.ErrEmptyInput)
		assert.Nil(t, positions)
		positions, err = searcher.ByPositionFromContext(empty, search.FullMatch)
		assert.ErrorIs(t, err, search.ErrEmptyInput)
		assert.Nil(t, positions)
	}
}
