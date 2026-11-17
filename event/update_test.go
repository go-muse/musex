package event_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventValidationOrder(t *testing.T) {
	t.Parallel()

	n := event.MustNoteContent(pitchNote("C4"))
	m := event.MustMeasureDuration(1)

	for _, tt := range []struct {
		name    string
		content event.Content
		d       event.Duration
		marks   event.Marks
		err     error
	}{
		{"content before duration", event.Content{}, event.Duration{}, event.Marks{Ties: []int{-1}}, event.ErrUnsetContent},
		{"duration before signs", n, event.Duration{}, event.Marks{Ties: []int{-1}, Grace: true}, event.ErrUnsetDuration},
		{"measure content before signs", n, m, event.Marks{Ties: []int{-1}, Arpeggio: true, Grace: true}, event.ErrMeasureContent},
		{"measure duration with chord", chordContent(pitchNote("C4")), m, event.Marks{}, event.ErrMeasureContent},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertFailure(t, tt.err,
				func() (*event.Event, error) { return event.New(tt.content, tt.d, tt.marks) },
				func() *event.Event { return event.MustNew(tt.content, tt.d, tt.marks) })
		})
	}
}

func TestUnsetEvent(t *testing.T) {
	t.Parallel()

	var zero event.Event
	valid := event.MustNew(event.RestContent(), valueDuration(1, 3), event.Marks{})
	for _, tt := range []struct {
		name string
		e    *event.Event
	}{
		{"nil receiver", nil},
		{"zero receiver", &zero},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.e
			assert.False(t, e.IsSet())
			assert.Equal(t, event.KindUnset, e.Kind())
			assert.False(t, e.Equal(e))
			assert.False(t, e.Equal(valid))
			assert.False(t, valid.Equal(e))

			t.Run("Content", func(t *testing.T) {
				assertFailure(t, event.ErrUnset, e.Content, e.MustContent)
			})
			t.Run("Duration", func(t *testing.T) {
				assertFailure(t, event.ErrUnset, e.Duration, e.MustDuration)
			})
			t.Run("Marks", func(t *testing.T) {
				assertFailure(t, event.ErrUnset, e.Marks, e.MustMarks)
			})
			t.Run("Length", func(t *testing.T) {
				assertFailure(t, event.ErrUnset, e.Length, e.MustLength)
			})
			t.Run("TiedNotes", func(t *testing.T) {
				assertFailure(t, event.ErrUnset, e.TiedNotes, e.MustTiedNotes)
			})
			t.Run("IsTied", func(t *testing.T) {
				assertFailure(t, event.ErrUnset,
					func() (bool, error) { return e.IsTied(-1) },
					func() bool { return e.MustIsTied(-1) })
			})
			t.Run("WithContent", func(t *testing.T) {
				assertFailure(t, event.ErrUnset,
					func() (*event.Event, error) { return e.WithContent(event.Content{}) },
					func() *event.Event { return e.MustWithContent(event.Content{}) })
			})
			t.Run("WithContentAndMarks", func(t *testing.T) {
				marks := event.Marks{Ties: []int{-1}}
				assertFailure(t, event.ErrUnset,
					func() (*event.Event, error) { return e.WithContentAndMarks(event.Content{}, marks) },
					func() *event.Event { return e.MustWithContentAndMarks(event.Content{}, marks) })
			})
			t.Run("WithDuration", func(t *testing.T) {
				assertFailure(t, event.ErrUnset,
					func() (*event.Event, error) { return e.WithDuration(event.Duration{}) },
					func() *event.Event { return e.MustWithDuration(event.Duration{}) })
			})
			t.Run("WithMarks", func(t *testing.T) {
				marks := event.Marks{Ties: []int{-1}}
				assertFailure(t, event.ErrUnset,
					func() (*event.Event, error) { return e.WithMarks(marks) },
					func() *event.Event { return e.MustWithMarks(marks) })
			})
			t.Run("EqualLength", func(t *testing.T) {
				assertFailure(t, event.ErrUnset,
					func() (bool, error) { return e.EqualLength(valid) },
					func() bool { return e.MustEqualLength(valid) })
			})
			t.Run("CompareLength", func(t *testing.T) {
				assertFailure(t, event.ErrUnset,
					func() (int, error) { return e.CompareLength(valid) },
					func() int { return e.MustCompareLength(valid) })
			})
		})
	}
}

func TestWithDurationRetainsContentAndMarks(t *testing.T) {
	t.Parallel()

	c, g := pitchNote("C4"), pitchNote("G4")
	ch := chordContent(c, g)
	initial := writtenDuration(duration.Quarter)
	m := event.Marks{Ties: []int{0, 1}, Arpeggio: true, Fermata: true, Grace: true}
	original := event.MustNew(ch, initial, m)
	d := writtenDuration(duration.Half)

	next, err := original.WithDuration(d)
	require.NoError(t, err)
	require.NotNil(t, next)
	assert.NotSame(t, original, next)
	assert.True(t, ch.Equal(next.MustContent()))
	assert.True(t, d.Equal(next.MustDuration()))
	assertMarks(t, m, next.MustMarks())
	assertFraction(t, fraction.Zero(), next.MustLength())
	assertFraction(t, fraction.MustNew(1, 2), next.MustDuration().MustValue().MustFraction())
	assert.NotSame(t, original, original.MustWithDuration(d))
	assert.True(t, next.Equal(original.MustWithDuration(d)))

	for _, tt := range []struct {
		name string
		d    event.Duration
		err  error
	}{
		{"unset duration", event.Duration{}, event.ErrUnsetDuration},
		{"grace needs written duration", valueDuration(1, 2), event.ErrIncompatibleMark},
		{"measure content before grace incompatibility", event.MustMeasureDuration(1), event.ErrMeasureContent},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertFailure(t, tt.err,
				func() (*event.Event, error) { return original.WithDuration(tt.d) },
				func() *event.Event { return original.MustWithDuration(tt.d) })
			assert.True(t, initial.Equal(original.MustDuration()))
			assertMarks(t, m, original.MustMarks())
		})
	}
	regularMarks := event.Marks{Ties: []int{1}, Arpeggio: true, Fermata: true}
	regular := original.MustWithMarks(regularMarks)
	changed := regular.MustWithDuration(valueDuration(1, 3))
	assertMarks(t, regularMarks, changed.MustMarks())
	assertFraction(t, fraction.MustNew(1, 3), changed.MustLength())
	assertFraction(t, fraction.MustNew(1, 4), regular.MustLength())
	assert.True(t, ch.Equal(original.MustContent()))
}

func TestWithContentPreservesTiedOccurrences(t *testing.T) {
	t.Parallel()

	c, e, g, a := pitchNote("C4"), pitchNote("E4"), pitchNote("G4"), pitchNote("A4")
	reboundG := pitchNote("G4")
	quarter := writtenDuration(duration.Quarter)
	marks := event.Marks{Ties: []int{2}, Fermata: true, Arpeggio: true}
	ch := chordContent(c, e, g)
	original := event.MustNew(ch, quarter, marks)

	replaced := chordContent(a, c, reboundG, e)
	got, err := original.WithContent(replaced)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.NotSame(t, original, got)
	assertNotes(t, []note.Note{a, c, reboundG, e}, got.MustContent().MustNotes())
	assertNotes(t, []note.Note{reboundG}, got.MustTiedNotes())
	assert.True(t, quarter.Equal(got.MustDuration()))
	assertMarks(t, marks, got.MustMarks())
	assert.NotSame(t, original, original.MustWithContent(replaced))
	assert.True(t, got.Equal(original.MustWithContent(replaced)))

	adjusted := g
	adjusted.MustSetCents(cents.MustFromInt(25))
	for _, tt := range []struct {
		name    string
		content event.Content
		err     error
	}{
		{"unset content first", event.Content{}, event.ErrUnsetContent},
		{"tied index disappears", chordContent(c, g), event.ErrTiedNoteChanged},
		{"tied note shifted", chordContent(g, e, c), event.ErrTiedNoteChanged},
		{"tied note changes cents", chordContent(c, e, adjusted), event.ErrTiedNoteChanged},
		{"tie preservation precedes incompatibility", event.RestContent(), event.ErrTiedNoteChanged},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertFailure(t, tt.err,
				func() (*event.Event, error) { return original.WithContent(tt.content) },
				func() *event.Event { return original.MustWithContent(tt.content) })
			assertNotes(t, []note.Note{c, e, g}, original.MustContent().MustNotes())
			assertMarks(t, marks, original.MustMarks())
		})
	}

	n := event.MustNoteContent(c)
	single := chordContent(c)
	tied := event.Marks{Ties: []int{0}, Fermata: true}
	lone := event.MustNew(n, quarter, tied)
	asChord := lone.MustWithContent(single)
	assert.Equal(t, event.KindChord, asChord.Kind())
	assertMarks(t, tied, asChord.MustMarks())
	asNote := asChord.MustWithContent(n)
	assert.Equal(t, event.KindNote, asNote.Kind())
	assertMarks(t, tied, asNote.MustMarks())

	arpeggiated := asChord.MustWithMarks(event.Marks{Ties: []int{0}, Arpeggio: true})
	assertFailure(t, event.ErrIncompatibleMark,
		func() (*event.Event, error) { return arpeggiated.WithContent(n) },
		func() *event.Event { return arpeggiated.MustWithContent(n) })

	measureRest := event.MustNew(event.RestContent(), event.MustMeasureDuration(1), event.Marks{})
	assertFailure(t, event.ErrMeasureContent,
		func() (*event.Event, error) { return measureRest.WithContent(n) },
		func() *event.Event { return measureRest.MustWithContent(n) })
}

func TestWithContentAndMarksRevalidatesTheResultAtomically(t *testing.T) {
	t.Parallel()

	c, e, g := pitchNote("C4"), pitchNote("E4"), pitchNote("G4")
	ch := chordContent(c, e, g)
	d := writtenDuration(duration.Half)
	originalMarks := event.Marks{Ties: []int{2}, Arpeggio: true, Fermata: true}
	original := event.MustNew(ch, d, originalMarks)
	single := event.MustNoteContent(e)
	ties := []int{0, 0}
	newMarks := event.Marks{Ties: ties, Grace: true}

	replaced, err := original.WithContentAndMarks(single, newMarks)
	require.NoError(t, err)
	require.NotNil(t, replaced)
	assert.NotSame(t, original, replaced)
	assert.True(t, single.Equal(replaced.MustContent()))
	assert.True(t, d.Equal(replaced.MustDuration()))
	assert.Equal(t, []int{0, 0}, ties)
	assertMarks(t, event.Marks{Ties: []int{0}, Grace: true}, replaced.MustMarks())
	assertFraction(t, fraction.Zero(), replaced.MustLength())
	assertNotes(t, []note.Note{e}, replaced.MustTiedNotes())

	rest := original.MustWithContentAndMarks(event.RestContent(), event.Marks{})
	require.NotNil(t, rest)
	assert.NotSame(t, original, rest)
	assert.Equal(t, event.KindRest, rest.Kind())
	assert.True(t, d.Equal(rest.MustDuration()))
	assertFraction(t, fraction.MustNew(1, 2), rest.MustLength())
	assertMarks(t, event.Marks{}, rest.MustMarks())

	for _, tt := range []struct {
		name    string
		content event.Content
		marks   event.Marks
		err     error
	}{
		{"unset content before marks", event.Content{}, event.Marks{Ties: []int{-1}}, event.ErrUnsetContent},
		{"reject a new tie on a rest", event.RestContent(), event.Marks{Ties: []int{2}}, event.ErrIncompatibleMark},
		{"reject a new out of range tie", single, event.Marks{Ties: []int{0, 2, 0}}, event.ErrIndex},
	} {
		t.Run(tt.name, func(t *testing.T) {
			originalTies := slices.Clone(tt.marks.Ties)
			assertFailure(t, tt.err,
				func() (*event.Event, error) { return original.WithContentAndMarks(tt.content, tt.marks) },
				func() *event.Event { return original.MustWithContentAndMarks(tt.content, tt.marks) })
			assert.Equal(t, originalTies, tt.marks.Ties)
			assertNotes(t, []note.Note{c, e, g}, original.MustContent().MustNotes())
			assertMarks(t, originalMarks, original.MustMarks())
			assert.True(t, d.Equal(original.MustDuration()))
		})
	}
}

func TestWithOperationsReturnNewPointersForEqualRecords(t *testing.T) {
	t.Parallel()

	content := event.RestContent()
	d := event.MustMeasureDuration(3)
	marks := event.Marks{Fermata: true}
	original := event.MustNew(content, d, marks)
	for _, tt := range []struct {
		name string
		with func() (*event.Event, error)
	}{
		{"WithContent", func() (*event.Event, error) { return original.WithContent(content) }},
		{"WithContentAndMarks", func() (*event.Event, error) { return original.WithContentAndMarks(content, marks) }},
		{"WithDuration", func() (*event.Event, error) { return original.WithDuration(d) }},
		{"WithMarks", func() (*event.Event, error) { return original.WithMarks(marks) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e, err := tt.with()
			require.NoError(t, err)
			require.NotNil(t, e)
			assert.NotSame(t, original, e)
			assert.True(t, original.Equal(e))
		})
	}
}
