package voiceleading_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The fixtures of this file serve the operations added by task 33.2. Compact
// note tokens are a letter, repeated sharps (#) or flats (b), and a written
// ASPN octave: "B#3" is the written B sharp 3, which sounds as C4 in the
// traditional system. Unusual spellings such as quarter tones are built
// directly in the tests that need them.

func traditionalSystem() *system.System { return system.Traditional() }

func quarterValue() duration.Duration { return duration.MustNew(1, 4) }

func spellingOf(t testing.TB, text string) spelling.Spelling {
	t.Helper()
	require.NotEmpty(t, text)
	letter, accidentals := text[:1], text[1:]
	machine := letter
	switch {
	case accidentals == "":
	case strings.Trim(accidentals, "#") == "":
		machine += fmt.Sprintf("+%d", len(accidentals))
	case strings.Trim(accidentals, "b") == "":
		machine += fmt.Sprintf("-%d", len(accidentals))
	default:
		machine += accidentals
	}
	s, err := spelling.Parse(machine)
	require.NoError(t, err, "spelling %q", text)
	return s
}

// writtenPitch reads a compact token with a written octave in the traditional
// system.
func writtenPitch(t testing.TB, token string) pitch.Pitch {
	t.Helper()
	i := len(token)
	for i > 0 && token[i-1] >= '0' && token[i-1] <= '9' {
		i--
	}
	require.True(t, i > 0 && i < len(token), "token %q needs a name and an octave", token)
	number, err := strconv.Atoi(token[i:])
	require.NoError(t, err)
	p, err := traditionalSystem().FromWritten(spellingOf(t, token[:i]), octave.FromASPN(number))
	require.NoError(t, err, "token %q", token)
	return p
}

func noteOf(t testing.TB, token string, src settings.Source) note.Note {
	t.Helper()
	n, err := note.FromPitch(writtenPitch(t, token), cents.Zero(), src)
	require.NoError(t, err)
	return n
}

// aspnNote builds a note from a positional ASPN octave, which reaches extreme
// registers that a compact token cannot express.
func aspnNote(t testing.TB, name string, aspn int) note.Note {
	t.Helper()
	p, err := pitch.FromASPN(spelling.MustParse(name), aspn)
	require.NoError(t, err)
	n, err := note.FromPitch(p, cents.Zero(), nil)
	require.NoError(t, err)
	return n
}

func nameOf(t testing.TB, text string) note.Note {
	t.Helper()
	n, err := note.FromSpelling(spellingOf(t, text), cents.Zero(), nil)
	require.NoError(t, err)
	return n
}

func notesOf(t testing.TB, tokens ...string) []note.Note {
	t.Helper()
	ns := make([]note.Note, len(tokens))
	for i, token := range tokens {
		ns[i] = noteOf(t, token, nil)
	}
	return ns
}

func chordOf(t testing.TB, tokens ...string) chord.Chord {
	t.Helper()
	c, err := chord.New(notesOf(t, tokens...)...)
	require.NoError(t, err)
	return c
}

// namesChord is a chord of notes without octaves: a named chord composition.
func namesChord(t testing.TB, texts ...string) chord.Chord {
	t.Helper()
	ns := make([]note.Note, len(texts))
	for i, text := range texts {
		ns[i] = nameOf(t, text)
	}
	c, err := chord.New(ns...)
	require.NoError(t, err)
	return c
}

func sonorityOf(t testing.TB, tokens ...string) *voiceleading.Sonority {
	t.Helper()
	s, err := voiceleading.NewSonority(chordOf(t, tokens...), traditionalSystem())
	require.NoError(t, err)
	require.NotNil(t, s)
	return s
}

func melodyOf(t *testing.T, tokens ...string) *voiceleading.Melody {
	t.Helper()
	m, err := voiceleading.NewMelody(line(t, tokens...), traditionalSystem())
	require.NoError(t, err)
	require.NotNil(t, m)
	return m
}

func analyzeParts(t *testing.T, order voiceleading.Order, sys *system.System, lines ...*voice.Line) *voiceleading.Analysis {
	t.Helper()
	a, err := voiceleading.Analyze(lines, voiceleading.Options{System: sys, Order: order})
	require.NoError(t, err)
	require.NotNil(t, a)
	return a
}

// boundElement builds a quarter-note sound bound to src. The source is read
// by the note constructor, so tests reset its counter afterwards.
func boundElement(t *testing.T, token string, src settings.Source) voice.Element {
	t.Helper()
	e, err := voice.NewSound(noteOf(t, token, src), voice.MustValueDuration(quarterValue()))
	require.NoError(t, err)
	return e
}

func ivl(t testing.TB, text string) interval.Interval {
	t.Helper()
	i, err := interval.Parse(text)
	require.NoError(t, err, "interval %q", text)
	return i
}

// assertIntervalText compares an interval with its short notation; an empty
// want means the unset interval.
func assertIntervalText(t testing.TB, want string, got interval.Interval, context ...any) {
	t.Helper()
	if want == "" {
		assert.False(t, got.IsValid(), context...)
		return
	}
	if assert.True(t, got.IsValid(), context...) {
		assert.True(t, ivl(t, want).Equal(got), "want %s, got %s %v", want, got, context)
	}
}

func assertNotesAre(t testing.TB, want []string, got []note.Note, context ...any) {
	t.Helper()
	if !assert.Len(t, got, len(want), context...) {
		return
	}
	for i, token := range want {
		p, err := got[i].Pitch()
		if assert.NoError(t, err, context...) {
			assert.True(t, writtenPitch(t, token).Equal(p), "voice %d: want %s, got %s %v", i, token, got[i], context)
		}
	}
}

func assertVoiceInterval(t testing.TB, got voiceleading.VoiceInterval, lower, upper int, iv string, chromatic int, context ...any) {
	t.Helper()
	assert.Equal(t, lower, got.Lower, context...)
	assert.Equal(t, upper, got.Upper, context...)
	assertIntervalText(t, iv, got.Interval, context...)
	assert.Equal(t, system.Step(chromatic), got.Chromatic, context...)
}

var allSentinels = []error{
	voiceleading.ErrUnset, voiceleading.ErrPartCount, voiceleading.ErrPart, voiceleading.ErrOption,
	voiceleading.ErrNoOrder, voiceleading.ErrNoSystem, voiceleading.ErrDifferentSystems, voiceleading.ErrTransition,
	voiceleading.ErrNoOctave, voiceleading.ErrVoiceCount, voiceleading.ErrEmpty, voiceleading.ErrNoVoices,
	voiceleading.ErrSilentPart,
}

// assertSentinel checks that err matches want and no other sentinel of the
// package; a nil want means a cause of a dependency without any sentinel.
func assertSentinel(t testing.TB, err, want error) {
	t.Helper()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "voiceleading: ")
	for _, sentinel := range allSentinels {
		assert.Equal(t, sentinel == want, errors.Is(err, sentinel), "sentinel %v in %v", sentinel, err)
	}
}

func assertChordError(t testing.TB, err, target error, chordIndex, noteIndex int) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, target)
	ce, ok := errors.AsType[*voiceleading.ChordError](err)
	require.True(t, ok, "want a ChordError, got %v", err)
	require.NotNil(t, ce)
	assert.Equal(t, chordIndex, ce.Chord)
	assert.Equal(t, noteIndex, ce.Note)
	assert.ErrorIs(t, ce.Unwrap(), target)
	assert.Contains(t, ce.Error(), "voiceleading: ")
}

func assertNoChordError(t testing.TB, err error) {
	t.Helper()
	_, ok := errors.AsType[*voiceleading.ChordError](err)
	assert.False(t, ok, "unexpected ChordError in %v", err)
}

func assertNoPartError(t testing.TB, err error) {
	t.Helper()
	_, ok := errors.AsType[*voiceleading.PartError](err)
	assert.False(t, ok, "unexpected PartError in %v", err)
}

// assertMustPanics checks that a Must form panics with an error equal in text
// and causes to the error of the ordinary form.
func assertMustPanics(t testing.TB, want error, call func()) {
	t.Helper()
	require.Error(t, want)
	var recovered any
	func() { defer func() { recovered = recover() }(); call() }()
	err, ok := recovered.(error)
	require.True(t, ok, "a Must form must panic with an error, got %T", recovered)
	assert.Equal(t, want.Error(), err.Error())
	for _, sentinel := range allSentinels {
		assert.Equal(t, errors.Is(want, sentinel), errors.Is(err, sentinel), "sentinel %v", sentinel)
	}
}
