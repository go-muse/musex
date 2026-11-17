package event_test

import (
	"reflect"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pitchNote(name string) note.Note {
	return note.MustParsePitch(settings.Standard(), name)
}

func chordContent(notes ...note.Note) event.Content {
	return event.MustChordContent(chord.MustNew(notes...))
}

func assertNote(t *testing.T, want, got note.Note) {
	t.Helper()
	assert.True(t, want.Equal(got), "want %v, got %v", want, got)
	assert.True(t, settings.SameSource(want.Source(), got.Source()), "note source changed")
}

func assertNotes(t *testing.T, want, got []note.Note) {
	t.Helper()
	require.NotNil(t, got)
	require.Len(t, got, len(want))
	for i := range want {
		assertNote(t, want[i], got[i])
	}
}

func TestKinds(t *testing.T) {
	t.Parallel()

	for _, kind := range []event.Kind{event.KindNote, event.KindChord, event.KindRest} {
		assert.True(t, kind.IsValid(), "content kind %d", kind)
	}
	for _, kind := range []event.Kind{event.KindUnset, 4, 255} {
		assert.False(t, kind.IsValid(), "content kind %d", kind)
	}
	for _, kind := range []event.DurationKind{event.DurationWritten, event.DurationValue, event.DurationMeasures} {
		assert.True(t, kind.IsValid(), "duration kind %d", kind)
	}
	for _, kind := range []event.DurationKind{event.DurationUnset, 4, 255} {
		assert.False(t, kind.IsValid(), "duration kind %d", kind)
	}
}

func TestContentOccurrences(t *testing.T) {
	t.Parallel()

	c := pitchNote("C4")
	e := pitchNote("E4")
	noOctave := note.MustParse(settings.Standard(), "C")
	triad := chord.MustNew(c, e, c)
	single := chord.MustNew(c)

	pitched, err := event.NoteContent(c)
	require.NoError(t, err)
	unpitched, err := event.NoteContent(noOctave)
	require.NoError(t, err)
	chordWithRepeat, err := event.ChordContent(triad)
	require.NoError(t, err)
	oneNoteChord, err := event.ChordContent(single)
	require.NoError(t, err)
	assert.True(t, event.MustNoteContent(c).Equal(pitched))
	assert.True(t, event.MustChordContent(triad).Equal(chordWithRepeat))

	cases := []struct {
		name    string
		content event.Content
		kind    event.Kind
		notes   []note.Note
		hasC    bool
	}{
		{"note with octave", pitched, event.KindNote, []note.Note{c}, true},
		{"note without octave", unpitched, event.KindNote, []note.Note{noOctave}, false},
		{"chord with repeat", chordWithRepeat, event.KindChord, []note.Note{c, e, c}, true},
		{"one-note chord", oneNoteChord, event.KindChord, []note.Note{c}, true},
		{"rest", event.RestContent(), event.KindRest, nil, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, tt.content.IsSet())
			assert.Equal(t, tt.kind, tt.content.Kind())
			count, err := tt.content.Len()
			require.NoError(t, err)
			assert.Equal(t, len(tt.notes), count)
			assert.Equal(t, len(tt.notes), tt.content.MustLen())

			notes, err := tt.content.Notes()
			require.NoError(t, err)
			assertNotes(t, tt.notes, notes)
			assertNotes(t, tt.notes, tt.content.MustNotes())

			for index, want := range tt.notes {
				got, err := tt.content.At(index)
				require.NoError(t, err)
				assertNote(t, want, got)
				assertNote(t, want, tt.content.MustAt(index))
			}
			for _, index := range []int{-1, len(tt.notes)} {
				t.Run(indexName(index), func(t *testing.T) {
					assertFailure(t, event.ErrIndex,
						func() (note.Note, error) { return tt.content.At(index) },
						func() note.Note { return tt.content.MustAt(index) })
				})
			}

			hasC, err := tt.content.Contains(c)
			require.NoError(t, err)
			assert.Equal(t, tt.hasC, hasC)
			assert.Equal(t, tt.hasC, tt.content.MustContains(c))

			var unset note.Note
			assertFailure(t, note.ErrUnset,
				func() (bool, error) { return tt.content.Contains(unset) },
				func() bool { return tt.content.MustContains(unset) })
		})
	}
}

func TestContentSpecificForms(t *testing.T) {
	t.Parallel()

	c := pitchNote("C4")
	e := pitchNote("E4")
	n := event.MustNoteContent(c)
	triad := chord.MustNew(c, e, c)
	ch := event.MustChordContent(triad)
	rest := event.RestContent()

	gotNote, err := n.Note()
	require.NoError(t, err)
	assertNote(t, c, gotNote)
	assertNote(t, c, n.MustNote())

	gotChord, err := ch.Chord()
	require.NoError(t, err)
	assert.True(t, triad.EqualInOrder(gotChord))
	assert.True(t, triad.EqualInOrder(ch.MustChord()))

	for _, tt := range []struct {
		name    string
		content event.Content
	}{
		{"chord", ch},
		{"rest", rest},
	} {
		t.Run(tt.name+" has no single note", func(t *testing.T) {
			assertFailure(t, event.ErrWrongKind, tt.content.Note, tt.content.MustNote)
		})
	}
	for _, tt := range []struct {
		name    string
		content event.Content
	}{
		{"note", n},
		{"rest", rest},
	} {
		t.Run(tt.name+" has no chord", func(t *testing.T) {
			assertFailure(t, event.ErrWrongKind, tt.content.Chord, tt.content.MustChord)
		})
	}
}

func TestContentConstructorFailures(t *testing.T) {
	t.Parallel()

	t.Run("unset note", func(t *testing.T) {
		var n note.Note
		assertFailure(t, note.ErrUnset,
			func() (event.Content, error) { return event.NoteContent(n) },
			func() event.Content { return event.MustNoteContent(n) })
	})
	chords := []struct {
		name string
		ch   chord.Chord
		err  error
	}{
		{"unset", chord.Chord{}, chord.ErrUnset},
		{"set but empty", chord.MustNew(), event.ErrEmptyChord},
	}
	for _, tt := range chords {
		t.Run(tt.name, func(t *testing.T) {
			assertFailure(t, tt.err,
				func() (event.Content, error) { return event.ChordContent(tt.ch) },
				func() event.Content { return event.MustChordContent(tt.ch) })
		})
	}
}

func TestUnsetContent(t *testing.T) {
	t.Parallel()

	var c event.Content
	assert.False(t, reflect.TypeFor[event.Content]().Comparable())
	assert.False(t, c.IsSet())
	assert.Equal(t, event.KindUnset, c.Kind())
	assert.False(t, c.Equal(c))
	assert.False(t, c.Equal(event.RestContent()))
	assert.False(t, event.RestContent().Equal(c))

	t.Run("Note", func(t *testing.T) {
		assertFailure(t, event.ErrUnsetContent, c.Note, c.MustNote)
	})
	t.Run("Chord", func(t *testing.T) {
		assertFailure(t, event.ErrUnsetContent, c.Chord, c.MustChord)
	})
	t.Run("Len", func(t *testing.T) {
		assertFailure(t, event.ErrUnsetContent, c.Len, c.MustLen)
	})
	t.Run("Notes", func(t *testing.T) {
		assertFailure(t, event.ErrUnsetContent, c.Notes, c.MustNotes)
	})
	t.Run("At before index validation", func(t *testing.T) {
		assertFailure(t, event.ErrUnsetContent,
			func() (note.Note, error) { return c.At(-1) },
			func() note.Note { return c.MustAt(-1) })
	})
	t.Run("Contains before needle validation", func(t *testing.T) {
		assertFailure(t, event.ErrUnsetContent,
			func() (bool, error) { return c.Contains(note.Note{}) },
			func() bool { return c.MustContains(note.Note{}) })
	})
}

func TestContentEqualityAndMembershipUseRecordedNotes(t *testing.T) {
	t.Parallel()

	c := pitchNote("C4")
	e := pitchNote("E4")
	sharp := pitchNote("C♯4")
	adjusted := c
	adjusted.MustSetCents(cents.MustFromInt(25))

	cases := []struct {
		name string
		have note.Note
		want note.Note
		same bool
	}{
		{"different sources", c, pitchNote("C4"), true},
		{"enharmonic spellings", sharp, pitchNote("D♭4"), false},
		{"different cents", c, adjusted, false},
		{"different octaves", c, pitchNote("C5"), false},
		{"missing octave", c, note.MustParse(settings.Standard(), "C"), false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a, b := event.MustNoteContent(tt.have), event.MustNoteContent(tt.want)
			assert.Equal(t, tt.same, a.Equal(b))
			assert.Equal(t, tt.same, b.Equal(a))
			assert.Equal(t, tt.same, a.MustContains(tt.want))
			ch := chordContent(e, tt.have, e)
			assert.Equal(t, tt.same, ch.MustContains(tt.want))
			assert.Equal(t, tt.same, ch.Equal(chordContent(e, tt.want, e)))
		})
	}

	ch := chordContent(c, e, c)
	assert.True(t, ch.Equal(chordContent(c, e, c)))
	assert.False(t, ch.Equal(chordContent(c, c, e)))
	assert.False(t, ch.Equal(chordContent(c, e)))
	assert.False(t, event.MustNoteContent(c).Equal(chordContent(c)))
	assert.True(t, event.RestContent().Equal(event.RestContent()))
}
