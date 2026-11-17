package midi_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/midi"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/require"
)

func TestFromNote_OwnSourceUsesOneSnapshot(t *testing.T) {
	t.Parallel()

	src := newCountingSource(settings.Standard())
	original := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.MustFromInt(30), src)
	before := original
	callsBefore, _ := src.State()

	number, err := midi.FromNote(original)
	require.NoError(t, err)
	requireNumberValue(t, 69, number)
	calls, snapshots := src.State()
	require.Equal(t, callsBefore+1, calls)
	newSnapshots := snapshots[callsBefore:]
	require.Len(t, newSnapshots, 1)
	require.True(t, newSnapshots[0].Equal(settings.Standard()))
	require.True(t, original.Equal(before))
	require.True(t, original.SameSource(before))
	require.True(t, original.Cents().Equal(cents.MustFromInt(30)))

	requireNumberValue(t, 69, midi.MustFromNote(original))
	previousCalls := calls
	calls, snapshots = src.State()
	require.Equal(t, previousCalls+1, calls)
	newSnapshots = snapshots[previousCalls:]
	require.Len(t, newSnapshots, 1)
	require.True(t, newSnapshots[0].Equal(settings.Standard()))
}

func TestFromNoteIn_ExplicitSourceDoesNotChangeNote(t *testing.T) {
	t.Parallel()

	t.Run("unbound note", func(t *testing.T) {
		src := newCountingSource(settings.Standard())
		original := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.MustFromInt(-10), nil)
		before := original
		number, err := midi.FromNoteIn(src, original)
		require.NoError(t, err)
		requireNumberValue(t, 69, number)
		calls, _ := src.State()
		require.Equal(t, 1, calls)
		require.True(t, original.Equal(before))
		require.Nil(t, original.Source())
		require.False(t, original.IsBound())
		requireNumberValue(t, 69, midi.MustFromNoteIn(src, original))
		calls, _ = src.State()
		require.Equal(t, 2, calls)
	})

	t.Run("explicit source takes priority over own 19 position source", func(t *testing.T) {
		own := newCountingSource(settingsForSystem(system.Positions19()))
		explicitSource := newCountingSource(settings.Standard())
		original := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.MustFromInt(30), own)
		before := original
		ownCallsBefore, _ := own.State()

		number, err := midi.FromNoteIn(explicitSource, original)
		require.NoError(t, err)
		requireNumberValue(t, 69, number)
		ownCalls, _ := own.State()
		explicitCalls, _ := explicitSource.State()
		require.Equal(t, ownCallsBefore, ownCalls, "FromNoteIn must not read the note own source")
		require.Equal(t, 1, explicitCalls)
		require.True(t, original.Equal(before))
		require.True(t, original.SameSource(before))
		require.True(t, settings.SameSource(original.Source(), own))
		requireNumberValue(t, 69, midi.MustFromNoteIn(explicitSource, original))
		explicitCalls, _ = explicitSource.State()
		require.Equal(t, 2, explicitCalls)

		_, err = midi.FromNote(original)
		assertMidiError(t, err, midi.ErrNotTraditionalC)
	})
}

func TestFromNote_Errors(t *testing.T) {
	t.Parallel()

	validSource := newCountingSource(settings.Standard())
	freeA4 := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.Zero(), nil)
	withoutOctave := note.MustFromSpelling(spelling.A(), cents.Zero(), nil)

	for _, tt := range []struct {
		name       string
		note       note.Note
		midiError  error
		dependency error
	}{
		{name: "unset note", note: note.Note{}, midiError: midi.ErrUnset, dependency: note.ErrUnset},
		{name: "no octave precedes no source", note: withoutOctave, midiError: midi.ErrNoOctave, dependency: note.ErrNoOctave},
		{name: "no own source", note: freeA4, midiError: midi.ErrNoSource, dependency: note.ErrNoSource},
	} {
		t.Run(tt.name, func(t *testing.T) {
			number, err := midi.FromNote(tt.note)
			assertMidiError(t, err, tt.midiError)
			require.ErrorIs(t, err, tt.dependency)
			require.Equal(t, midi.Number{}, number)
			requirePanicError(t, err, func() { midi.MustFromNote(tt.note) })
		})
	}

	number, err := midi.FromNoteIn(validSource, withoutOctave)
	assertMidiError(t, err, midi.ErrNoOctave)
	require.ErrorIs(t, err, note.ErrNoOctave)
	require.Equal(t, midi.Number{}, number)
	requirePanicError(t, err, func() { midi.MustFromNoteIn(validSource, withoutOctave) })

	number, err = midi.FromNoteIn(nil, withoutOctave)
	assertMidiError(t, err, midi.ErrUnset)
	require.NotErrorIs(t, err, note.ErrUnset)
	require.NotErrorIs(t, err, note.ErrNoOctave)
	require.Equal(t, midi.Number{}, number)
	requirePanicError(t, err, func() { midi.MustFromNoteIn(nil, withoutOctave) })
}

func TestFromNoteIn_NilAndUnsetSources(t *testing.T) {
	t.Parallel()

	freeA4 := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.Zero(), nil)
	var nilSource settings.Source
	number, err := midi.FromNoteIn(nilSource, freeA4)
	assertMidiError(t, err, midi.ErrUnset)
	require.NotErrorIs(t, err, note.ErrUnset)
	require.Equal(t, midi.Number{}, number)
	requirePanicError(t, err, func() { midi.MustFromNoteIn(nilSource, freeA4) })

	created, err := midi.MustNew(69).Note(nilSource, system.Sharps(), cents.Zero())
	assertMidiError(t, err, midi.ErrUnset)
	require.NotErrorIs(t, err, note.ErrUnset)
	require.False(t, created.IsValid())
	requirePanicError(t, err, func() { midi.MustNew(69).MustNote(nilSource, system.Sharps(), cents.Zero()) })

	var nilSettings *settings.Settings
	for _, tt := range []struct {
		name   string
		source settings.Source
	}{
		{name: "typed nil settings", source: nilSettings},
		{name: "zero settings", source: &settings.Settings{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			number, err := midi.FromNoteIn(tt.source, freeA4)
			assertMidiError(t, err, midi.ErrUnsetSettings)
			require.ErrorIs(t, err, settings.ErrUnset)
			require.NotErrorIs(t, err, note.ErrUnsetSettings)
			require.Equal(t, midi.Number{}, number)
			requirePanicError(t, err, func() { midi.MustFromNoteIn(tt.source, freeA4) })

			created, err := midi.MustNew(69).Note(tt.source, system.Sharps(), cents.Zero())
			assertMidiError(t, err, midi.ErrUnsetSettings)
			require.ErrorIs(t, err, settings.ErrUnset)
			require.NotErrorIs(t, err, note.ErrUnsetSettings)
			require.False(t, created.IsValid())
			requirePanicError(t, err, func() { midi.MustNew(69).MustNote(tt.source, system.Sharps(), cents.Zero()) })
		})
	}
}

func TestFromNote_OwnSourceReturnsUnsetSettings(t *testing.T) {
	t.Parallel()

	src := newCountingSource(settings.Standard())
	original := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.Zero(), src)
	before := original
	src.Replace(nil)
	number, err := midi.FromNote(original)
	assertMidiError(t, err, midi.ErrUnsetSettings)
	require.ErrorIs(t, err, settings.ErrUnset)
	require.ErrorIs(t, err, note.ErrUnsetSettings)
	require.Equal(t, midi.Number{}, number)
	require.True(t, original.Equal(before))
	require.True(t, original.SameSource(before))
	requirePanicError(t, err, func() { midi.MustFromNote(original) })
}

func TestFromNote_SystemNameAndRangeErrors(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name       string
		system     *system.System
		spelling   spelling.Spelling
		octave     int
		midiError  error
		dependency error
	}{
		{name: "not traditional C", system: system.Positions19(), spelling: spelling.C(), octave: 4, midiError: midi.ErrNotTraditionalC},
		{name: "inadmissible name", system: naturalOnlySystem(), spelling: spelling.CSharp(), octave: 4, midiError: midi.ErrInadmissible, dependency: system.ErrInadmissible},
		{name: "out of range", system: system.Traditional(), spelling: spelling.C(), octave: 10, midiError: midi.ErrOutOfRange},
	} {
		t.Run(tt.name, func(t *testing.T) {
			src := newCountingSource(settingsForSystem(tt.system))
			free := note.MustFromPitch(pitch.MustFromASPN(tt.spelling, tt.octave), cents.Zero(), nil)
			number, err := midi.FromNoteIn(src, free)
			assertMidiError(t, err, tt.midiError)
			if tt.dependency != nil {
				require.ErrorIs(t, err, tt.dependency)
			}
			require.Equal(t, midi.Number{}, number)
			requirePanicError(t, err, func() { midi.MustFromNoteIn(src, free) })
		})
	}
}

func TestNumber_NoteUsesTwoEqualSnapshotsAndPreservesCorrection(t *testing.T) {
	t.Parallel()

	for _, correction := range []cents.Cents{{}, cents.MustFromInt(30), cents.MustFromInt(-10)} {
		t.Run(fmt.Sprintf("cents_%v", correction.Float64()), func(t *testing.T) {
			src := newCountingSource(settings.Standard())
			number := midi.MustNew(69)
			created, err := number.Note(src, system.Sharps(), correction)
			require.NoError(t, err)
			calls, snapshots := src.State()
			require.Equal(t, 2, calls)
			require.Len(t, snapshots, 2)
			require.True(t, snapshots[0].Equal(snapshots[1]), "Note requires two equal settings snapshots")
			require.True(t, created.IsValid())
			require.True(t, created.HasOctave())
			require.Same(t, src, created.Source())
			require.True(t, settings.SameSource(created.Source(), src))
			written, err := created.Pitch()
			require.NoError(t, err)
			require.True(t, written.Equal(pitch.MustFromASPN(spelling.A(), 4)))
			require.True(t, created.Cents().Equal(correction))

			mustCreated := number.MustNote(src, system.Sharps(), correction)
			require.True(t, mustCreated.Equal(created))
			require.Same(t, src, mustCreated.Source())
			calls, _ = src.State()
			require.Equal(t, 4, calls)
		})
	}
}

func TestFromNote_CorrectionAndSoundSettingsDoNotAffectNumber(t *testing.T) {
	t.Parallel()

	for _, correction := range []cents.Cents{cents.Zero(), cents.MustFromInt(30), cents.MustFromInt(-10)} {
		original := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), correction, settings.Standard())
		requireNumberValue(t, 69, midi.MustFromNote(original))
		require.True(t, original.Cents().Equal(correction))
	}

	standardTuning := tuning.Standard()
	for _, tt := range []struct {
		name   string
		tuning *tuning.Tuning
	}{
		{name: "A4 at 415", tuning: standardTuning.MustWithReferenceFrequency(415)},
		{name: "different temperament", tuning: standardTuning.MustWithTemperament(temperament.Pythagorean())},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := settings.Standard().MustWithTuning(tt.tuning)
			original := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.Zero(), source)
			requireNumberValue(t, 69, midi.MustFromNote(original))
		})
	}
}

func TestMidi_ConcurrentReadOnlyOperations(t *testing.T) {
	t.Parallel()

	number := midi.MustNew(60)
	written := pitch.MustFromASPN(spelling.C(), 4)
	boundNote := note.MustFromPitch(written, cents.Zero(), settings.Standard())

	const workers = 8
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if value, err := number.Int(); err != nil || value != 60 {
				errs[i] = fmt.Errorf("Int: value=%d err=%v", value, err)
				return
			}
			shifted, err := number.Shift(9)
			if err != nil {
				errs[i] = err
				return
			}
			if value, err := shifted.Int(); err != nil || value != 69 {
				errs[i] = fmt.Errorf("Shift: value=%d err=%v", value, err)
				return
			}
			o, class, position, err := number.Parts()
			if err != nil || o.MustASPN() != 4 || class.Number() != 0 || position != 48 {
				errs[i] = fmt.Errorf("Parts: octave=%v class=%v position=%d err=%v", o, class, position, err)
				return
			}
			assembled, err := midi.FromParts(o, class)
			if err != nil || !assembled.Equal(number) {
				errs[i] = fmt.Errorf("FromParts: number=%v err=%v", assembled, err)
				return
			}
			fromPitch, err := midi.FromPitch(written, system.Traditional())
			if err != nil || !fromPitch.Equal(number) {
				errs[i] = fmt.Errorf("FromPitch: number=%v err=%v", fromPitch, err)
				return
			}
			reversed, err := number.Pitch(system.Traditional(), system.Sharps())
			if err != nil || !reversed.Equal(written) {
				errs[i] = fmt.Errorf("Pitch: written=%v err=%v", reversed, err)
				return
			}
			fromNote, err := midi.FromNote(boundNote)
			if err != nil || !fromNote.Equal(number) {
				errs[i] = fmt.Errorf("FromNote: number=%v err=%v", fromNote, err)
				return
			}
			count := 0
			for range midi.All() {
				count++
			}
			if count != 128 {
				errs[i] = fmt.Errorf("All yielded %d values", count)
			}
		}()
	}
	wg.Wait()
	for i, err := range errs {
		require.NoError(t, err, "worker %d", i)
	}
	requireNumberValue(t, 60, number)
	require.True(t, written.Equal(pitch.MustFromASPN(spelling.C(), 4)))
}

func TestMidi_ConcurrentSharedSourceWithExternallySynchronizedReplacement(t *testing.T) {
	t.Parallel()

	src := newCountingSource(settings.Standard())
	freeA4 := note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.Zero(), nil)
	number := midi.MustNew(69)
	var operationSync sync.RWMutex

	const workers = 8
	type result struct {
		number midi.Number
		note   note.Note
		err    error
	}
	results := make([]result, workers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			operationSync.RLock()
			defer operationSync.RUnlock()
			if i%2 == 0 {
				results[i].number, results[i].err = midi.FromNoteIn(src, freeA4)
				return
			}
			results[i].note, results[i].err = number.Note(src, system.Sharps(), cents.Zero())
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		operationSync.Lock()
		defer operationSync.Unlock()
		src.Replace(settings.Standard().MustWithTuning(tuning.Standard().MustWithReferenceFrequency(415)))
	}()

	close(start)
	wg.Wait()

	for i, result := range results {
		require.NoError(t, result.err, "worker %d", i)
		if i%2 == 0 {
			requireNumberValue(t, 69, result.number)
			continue
		}
		require.True(t, result.note.IsValid())
		require.Same(t, src, result.note.Source())
		written, err := result.note.Pitch()
		require.NoError(t, err)
		require.True(t, written.Equal(pitch.MustFromASPN(spelling.A(), 4)))
	}
	require.True(t, freeA4.Equal(note.MustFromPitch(pitch.MustFromASPN(spelling.A(), 4), cents.Zero(), nil)))
}

func TestMidi_InputsRemainUnchangedOnSuccessAndFailure(t *testing.T) {
	t.Parallel()

	written := pitch.MustFromASPN(spelling.CSharp(), 4)
	writtenBefore := written
	sys := system.Traditional()
	definitionBefore := sys.Definition()
	src := newCountingSource(settings.Standard())
	bound := note.MustFromPitch(written, cents.MustFromInt(30), src)
	boundBefore := bound

	requireNumberValue(t, 61, midi.MustFromPitch(written, sys))
	_, err := midi.MustNew(61).Pitch(sys, system.SpellingRule{})
	require.Error(t, err)
	requireNumberValue(t, 61, midi.MustFromNote(bound))
	_, err = midi.MustNew(61).Note(src, system.SpellingRule{}, cents.Zero())
	require.Error(t, err)

	require.True(t, written.Equal(writtenBefore))
	require.Equal(t, definitionBefore, sys.Definition())
	require.True(t, bound.Equal(boundBefore))
	require.True(t, bound.SameSource(boundBefore))
	require.True(t, bound.Cents().Equal(cents.MustFromInt(30)))
	require.Same(t, src, bound.Source())
}

func TestMidi_ErrorPrecedence(t *testing.T) {
	t.Parallel()

	var invalid midi.Number
	written, err := invalid.Pitch(nil, system.SpellingRule{})
	assertMidiError(t, err, midi.ErrInvalidNumber)
	require.False(t, written.IsValid())
	require.NotErrorIs(t, err, system.ErrInvalid)

	created, err := invalid.Note(nil, system.SpellingRule{}, cents.Zero())
	assertMidiError(t, err, midi.ErrInvalidNumber)
	require.False(t, created.IsValid())
	require.NotErrorIs(t, err, settings.ErrUnset)

	number, err := midi.FromNoteIn(nil, note.Note{})
	assertMidiError(t, err, midi.ErrUnset)
	require.ErrorIs(t, err, note.ErrUnset)
	require.NotErrorIs(t, err, note.ErrNoOctave)
	require.Equal(t, midi.Number{}, number)
}
