package event_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChordTiesUseOccurrenceIndices(t *testing.T) {
	t.Parallel()

	c, e, g := pitchNote("C4"), pitchNote("E4"), pitchNote("G4")
	half := writtenDuration(duration.Half)
	fullChord := chordContent(c, e, g)
	ordinary := event.MustNew(fullChord, half, event.Marks{Ties: []int{2}, Fermata: true, Arpeggio: true})
	assertNotes(t, []note.Note{c, e, g}, ordinary.MustContent().MustNotes())
	assertMarks(t, event.Marks{Ties: []int{2}, Fermata: true, Arpeggio: true}, ordinary.MustMarks())
	assertFraction(t, fraction.MustNew(1, 2), ordinary.MustLength())
	assertNotes(t, []note.Note{g}, ordinary.MustTiedNotes())

	withRepeat := chordContent(c, e, c)
	first := event.MustNew(withRepeat, half, event.Marks{Ties: []int{0}})
	last := event.MustNew(withRepeat, half, event.Marks{Ties: []int{2}})
	assert.True(t, first.MustContent().MustContains(c))
	assert.True(t, last.MustContent().MustContains(c))
	assert.False(t, first.Equal(last))

	set := []struct {
		name    string
		input   []int
		indices []int
		tied    []note.Note
	}{
		{"nil ties", nil, nil, nil},
		{"empty ties", []int{}, nil, nil},
		{"only first C", []int{0}, []int{0}, []note.Note{c}},
		{"only second C", []int{2}, []int{2}, []note.Note{c}},
		{"both Cs retained", []int{2, 0, 2}, []int{0, 2}, []note.Note{c, c}},
		{"all occurrences in content order", []int{2, 1, 0, 1}, []int{0, 1, 2}, []note.Note{c, e, c}},
	}
	for _, tt := range set {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := event.MustNew(withRepeat, half, event.Marks{Ties: tt.input})
			assertMarks(t, event.Marks{Ties: tt.indices}, e.MustMarks())
			assert.True(t, e.Equal(event.MustNew(withRepeat, half, event.Marks{Ties: tt.indices})))
			got, err := e.TiedNotes()
			require.NoError(t, err)
			assertNotes(t, tt.tied, got)
			assertNotes(t, tt.tied, e.MustTiedNotes())

			for i := range 3 {
				tied, err := e.IsTied(i)
				require.NoError(t, err)
				assert.Equal(t, slices.Contains(tt.indices, i), tied, "index %d", i)
				assert.Equal(t, slices.Contains(tt.indices, i), e.MustIsTied(i), "Must index %d", i)
			}
		})
	}
}

func TestIsTiedRejectsInvalidIndices(t *testing.T) {
	t.Parallel()

	c, e := pitchNote("C4"), pitchNote("E4")
	d := writtenDuration(duration.Half)
	for _, tt := range []struct {
		name    string
		content event.Content
		invalid []int
	}{
		{"note", event.MustNoteContent(c), []int{-1, 1}},
		{"chord", chordContent(c, e), []int{-1, 2}},
		{"rest", event.RestContent(), []int{-1, 0}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e := event.MustNew(tt.content, d, event.Marks{})
			for _, i := range tt.invalid {
				t.Run(indexName(i), func(t *testing.T) {
					assertFailure(t, event.ErrIndex,
						func() (bool, error) { return e.IsTied(i) },
						func() bool { return e.MustIsTied(i) })
				})
			}
		})
	}

	n := event.MustNew(event.MustNoteContent(c), d, event.Marks{Ties: []int{0, 0}})
	assert.True(t, n.MustIsTied(0))
	assertMarks(t, event.Marks{Ties: []int{0}}, n.MustMarks())
	assertNotes(t, []note.Note{c}, n.MustTiedNotes())
}

func TestWithMarksSetsAndClearsEverySign(t *testing.T) {
	t.Parallel()

	c := pitchNote("C4")
	d := writtenDuration(duration.Half)
	// A single-note chord retains the chord form and can carry an arpeggio.
	content := chordContent(c)
	original := event.MustNew(content, d, event.Marks{})

	for _, tt := range []struct {
		name   string
		marks  event.Marks
		length fraction.Fraction
	}{
		{"tie", event.Marks{Ties: []int{0}}, fraction.MustNew(1, 2)},
		{"fermata", event.Marks{Fermata: true}, fraction.MustNew(1, 2)},
		{"arpeggio", event.Marks{Arpeggio: true}, fraction.MustNew(1, 2)},
		{"grace", event.Marks{Grace: true}, fraction.Zero()},
		{"all signs", event.Marks{Ties: []int{0}, Fermata: true, Arpeggio: true, Grace: true}, fraction.Zero()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			with, err := original.WithMarks(tt.marks)
			require.NoError(t, err)
			require.NotNil(t, with)
			assert.NotSame(t, original, with)
			assert.Equal(t, event.KindChord, with.Kind())
			assertMarks(t, tt.marks, with.MustMarks())
			assert.True(t, content.Equal(with.MustContent()))
			assert.True(t, d.Equal(with.MustDuration()))
			assertFraction(t, tt.length, with.MustLength())
			assertMarks(t, event.Marks{}, original.MustMarks())

			cleared := with.MustWithMarks(event.Marks{})
			require.NotNil(t, cleared)
			assert.NotSame(t, with, cleared)
			assertMarks(t, event.Marks{}, cleared.MustMarks())
			assertMarks(t, tt.marks, with.MustMarks())
			assert.True(t, original.Equal(cleared))
		})
	}
	all := event.Marks{Ties: []int{0}, Fermata: true, Arpeggio: true, Grace: true}
	replaced := original.MustWithMarks(all).MustWithMarks(event.Marks{Fermata: true})
	assertMarks(t, event.Marks{Fermata: true}, replaced.MustMarks())
}

func TestNewAndWithMarksRejectIncompatibleMarks(t *testing.T) {
	t.Parallel()

	c := pitchNote("C4")
	n, ch, rest := event.MustNoteContent(c), chordContent(c), event.RestContent()
	w, v, m := writtenDuration(duration.Quarter), valueDuration(1, 4), event.MustMeasureDuration(2)
	for _, tt := range []struct {
		name    string
		content event.Content
		d       event.Duration
		marks   event.Marks
		err     error
	}{
		{"ties on a rest", rest, w, event.Marks{Ties: []int{0}}, event.ErrIncompatibleMark},
		{"out of range tie on a rest", rest, m, event.Marks{Ties: []int{-1}}, event.ErrIncompatibleMark},
		{"arpeggio on a note", n, w, event.Marks{Arpeggio: true}, event.ErrIncompatibleMark},
		{"arpeggio on a rest", rest, v, event.Marks{Arpeggio: true}, event.ErrIncompatibleMark},
		{"grace written rest", rest, w, event.Marks{Grace: true}, event.ErrIncompatibleMark},
		{"grace measure rest", rest, m, event.Marks{Grace: true}, event.ErrIncompatibleMark},
		{"grace exact value note", n, v, event.Marks{Grace: true}, event.ErrIncompatibleMark},
		{"grace exact value chord", ch, v, event.Marks{Grace: true}, event.ErrIncompatibleMark},
		{"negative note tie index", n, w, event.Marks{Ties: []int{-1}}, event.ErrIndex},
		{"note tie index at length", n, w, event.Marks{Ties: []int{1}}, event.ErrIndex},
		{"chord ties include an out of range index", ch, w, event.Marks{Ties: []int{0, 1, 0}}, event.ErrIndex},
		{"tie index validation precedes incompatibility", n, v, event.Marks{Ties: []int{1}, Arpeggio: true, Grace: true}, event.ErrIndex},
	} {
		t.Run(tt.name, func(t *testing.T) {
			empty := event.MustNew(tt.content, tt.d, event.Marks{})
			inputTies := slices.Clone(tt.marks.Ties)
			assertFailure(t, tt.err,
				func() (*event.Event, error) { return event.New(tt.content, tt.d, tt.marks) },
				func() *event.Event { return event.MustNew(tt.content, tt.d, tt.marks) })
			assertFailure(t, tt.err,
				func() (*event.Event, error) { return empty.WithMarks(tt.marks) },
				func() *event.Event { return empty.MustWithMarks(tt.marks) })
			assert.Equal(t, inputTies, tt.marks.Ties)
			assertMarks(t, event.Marks{}, empty.MustMarks())
		})
	}
}
