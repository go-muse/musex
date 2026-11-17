package event_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoteAndWrittenCopiesAreIndependent(t *testing.T) {
	t.Parallel()

	wantNote := pitchNote("C4")
	inputNote := wantNote
	written := duration.MustParseWritten("4.[3:2]")
	wantWritten := written
	content := event.MustNoteContent(inputNote)
	timing := event.MustWrittenDuration(written)
	e := event.MustNew(content, timing, event.Marks{Ties: []int{0}})

	inputNote.MustSetASPN(6)
	require.NoError(t, written.UnmarshalText([]byte("8")))
	returnedNote := e.MustContent().MustNote()
	returnedNote.MustSetASPN(5)
	returnedOccurrence := e.MustContent().MustAt(0)
	returnedOccurrence.MustSetASPN(7)
	returnedWritten := e.MustDuration().MustWritten()
	require.NoError(t, returnedWritten.UnmarshalText([]byte("16")))

	assertNote(t, wantNote, content.MustNote())
	assertNote(t, wantNote, e.MustContent().MustNote())
	assertNote(t, wantNote, e.MustContent().MustAt(0))
	assertNotes(t, []note.Note{wantNote}, e.MustTiedNotes())
	assert.True(t, wantWritten.Equal(e.MustDuration().MustWritten()))
	assertFraction(t, fraction.MustNew(1, 4), e.MustLength())
}

func TestChordAndTieCopiesAreIndependent(t *testing.T) {
	t.Parallel()

	c, e, g := pitchNote("C4"), pitchNote("E4"), pitchNote("G4")
	inputChord := chord.MustNew(c, e, g)
	ties := []int{2, 0, 2}
	eventValue := event.MustNew(event.MustChordContent(inputChord), writtenDuration(duration.Half), event.Marks{Ties: ties})
	valueCopy := *eventValue
	pointerCopy := eventValue
	assert.Same(t, eventValue, pointerCopy)
	assert.Equal(t, []int{2, 0, 2}, ties, "New changed the input slice")
	ties[0] = 1
	inputChord.MustClear()

	outputChord := eventValue.MustContent().MustChord()
	chordNote := outputChord.MustAt(0)
	chordNote.MustSetASPN(6)
	outputChord.MustClear()

	notes := pointerCopy.MustContent().MustNotes()
	tied := pointerCopy.MustTiedNotes()
	marks := pointerCopy.MustMarks()
	otherNotes := eventValue.MustContent().MustNotes()
	otherTied := eventValue.MustTiedNotes()
	otherMarks := eventValue.MustMarks()

	require.Len(t, notes, 3)
	require.Len(t, tied, 2)
	require.Len(t, marks.Ties, 2)
	notes[0].MustSetASPN(6)
	notes[1] = g
	tied[0].MustSetASPN(6)
	tied[1] = e
	marks.Ties[0] = 1

	for _, original := range []*event.Event{eventValue, &valueCopy} {
		assertNotes(t, []note.Note{c, e, g}, original.MustContent().MustNotes())
		assertNotes(t, []note.Note{c, g}, original.MustTiedNotes())
		assertMarks(t, event.Marks{Ties: []int{0, 2}}, original.MustMarks())
	}
	assertNotes(t, []note.Note{c, e, g}, otherNotes)
	assertNotes(t, []note.Note{c, g}, otherTied)
	assertMarks(t, event.Marks{Ties: []int{0, 2}}, otherMarks)

	valueCopy = *eventValue.MustWithContentAndMarks(event.RestContent(), event.Marks{})
	assert.Equal(t, event.KindRest, valueCopy.Kind())
	assert.Equal(t, event.KindChord, pointerCopy.Kind())
	assertNotes(t, []note.Note{c, e, g}, pointerCopy.MustContent().MustNotes())
}

func TestWithOperationsDoNotRetainInputTieSlice(t *testing.T) {
	t.Parallel()

	content := chordContent(pitchNote("C4"), pitchNote("G4"), pitchNote("E4"))
	original := event.MustNew(content, writtenDuration(duration.Half), event.Marks{})
	for _, tt := range []struct {
		name string
		with func(event.Marks) (*event.Event, error)
	}{
		{"WithMarks", original.WithMarks},
		{"WithContentAndMarks", func(m event.Marks) (*event.Event, error) {
			return original.WithContentAndMarks(content, m)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ties := []int{2, 0, 2}
			e, err := tt.with(event.Marks{Ties: ties})
			require.NoError(t, err)
			require.NotNil(t, e)
			assert.Equal(t, []int{2, 0, 2}, ties)
			ties[0], ties[1] = 1, 1
			assertMarks(t, event.Marks{Ties: []int{0, 2}}, e.MustMarks())
			assertMarks(t, event.Marks{}, original.MustMarks())
		})
	}
}

type countingSettingsSource struct {
	snapshot *settings.Settings
	reads    atomic.Int64
}

func (s *countingSettingsSource) Settings() *settings.Settings {
	s.reads.Add(1)
	return s.snapshot
}

func assertTiedSource(t *testing.T, e *event.Event, want *countingSettingsSource) {
	t.Helper()
	tied := e.MustTiedNotes()
	require.Len(t, tied, 1)
	assert.Same(t, want, tied[0].Source())
}

func TestEventReadsAndEditsDoNotConsultSettings(t *testing.T) {
	t.Parallel()

	first := &countingSettingsSource{snapshot: settings.Standard()}
	second := &countingSettingsSource{snapshot: settings.Standard()}
	n := note.MustParsePitch(first, "C4")
	same := note.MustParsePitch(second, "C4")
	e := note.MustParsePitch(first, "E4")
	ch := chord.MustNew(n, e)
	reboundCh := chord.MustNew(same, e)
	first.reads.Store(0)
	second.reads.Store(0)

	noteContent := event.MustNoteContent(n)
	otherContent := event.MustNoteContent(same)
	builtChord := event.MustChordContent(ch)
	d := writtenDuration(duration.Quarter)
	m := event.Marks{Ties: []int{0}}
	pitched := event.MustNew(noteContent, d, m)
	chordEvent := event.MustNew(builtChord, d, m)

	assert.True(t, noteContent.Equal(otherContent))
	assert.True(t, pitched.Equal(event.MustNew(otherContent, d, m)))
	assert.True(t, noteContent.MustContains(same))
	assert.True(t, builtChord.MustContains(same))
	assert.True(t, builtChord.Equal(event.MustChordContent(reboundCh)))
	assert.Same(t, first, noteContent.MustNote().Source())
	assert.Same(t, first, builtChord.MustChord().MustAt(0).Source())
	assert.Same(t, first, chordEvent.MustContent().MustAt(0).Source())
	assertNotes(t, []note.Note{n, e}, chordEvent.MustContent().MustNotes())
	assertNotes(t, []note.Note{n}, chordEvent.MustTiedNotes())
	assert.True(t, chordEvent.MustIsTied(0))
	assertMarks(t, m, chordEvent.MustMarks())
	assertFraction(t, fraction.MustNew(1, 4), chordEvent.MustLength())
	assert.True(t, chordEvent.MustEqualLength(pitched))
	assert.Zero(t, chordEvent.MustCompareLength(pitched))
	assertTiedSource(t, pitched.MustWithContent(otherContent), second)
	assertTiedSource(t, chordEvent.MustWithContent(event.MustChordContent(reboundCh)), second)
	assertTiedSource(t, chordEvent.MustWithContentAndMarks(otherContent, m), second)
	assertTiedSource(t, chordEvent.MustWithDuration(valueDuration(1, 2)), first)
	assert.Same(t, first, chordEvent.MustWithMarks(event.Marks{}).MustContent().MustAt(0).Source())

	assert.Zero(t, first.reads.Load())
	assert.Zero(t, second.reads.Load())
}

func TestConcurrentReadsAndUpdatesUseIndependentCopies(t *testing.T) {
	t.Parallel()

	c, e, g := pitchNote("C4"), pitchNote("E4"), pitchNote("G4")
	content := chordContent(c, e, g)
	original := event.MustNew(content, writtenDuration(duration.Half), event.Marks{Ties: []int{0, 2}})
	const workers = 8
	var results [workers]struct {
		length     fraction.Fraction
		updated    *event.Event
		err        error
		panicValue any
	}
	var ready, done sync.WaitGroup
	ready.Add(workers)
	start := make(chan struct{})
	for i := range workers {
		done.Go(func() {
			defer func() { results[i].panicValue = recover() }()
			ready.Done()
			<-start

			// Changing these copies in each worker must not change the shared event.
			output := original.MustContent().MustChord()
			output.MustClear()
			notes := original.MustContent().MustNotes()
			notes[0] = e
			tied := original.MustTiedNotes()
			tied[0] = g
			marks := original.MustMarks()
			if len(marks.Ties) != 2 {
				return
			}
			marks.Ties[0] = i % 3
			results[i].length = original.MustLength()
			results[i].updated, results[i].err = original.WithMarks(marks)
		})
	}
	ready.Wait()
	close(start)
	done.Wait()

	for i, got := range results {
		require.Nil(t, got.panicValue, "worker %d panicked", i)
		require.NoError(t, got.err, "worker %d", i)
		require.NotNil(t, got.updated, "worker %d did not produce an updated event", i)
		assert.NotSame(t, original, got.updated)
		assertFraction(t, fraction.MustNew(1, 2), got.length)
		want := []int{2}
		if index := i % 3; index < 2 {
			want = []int{index, 2}
		}
		assertMarks(t, event.Marks{Ties: want}, got.updated.MustMarks())
		assertNotes(t, []note.Note{c, e, g}, got.updated.MustContent().MustNotes())
	}
	assertMarks(t, event.Marks{Ties: []int{0, 2}}, original.MustMarks())
	assertNotes(t, []note.Note{c, e, g}, original.MustContent().MustNotes())
	assertNotes(t, []note.Note{c, g}, original.MustTiedNotes())
}
