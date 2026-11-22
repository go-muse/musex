package track_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The fixture values implement the published contract, not a second track engine.
func position(n, d int) duration.Position   { return duration.MustNewPosition(n, d) }
func length(n, d int) duration.Duration     { return duration.MustNew(n, d) }
func address(bar, n, d int) measure.Address { return measure.MustNewAddress(bar, position(n, d)) }
func grid(bars int) measure.Grid            { return measure.MustNew(meter.MustNew(4, 4)).MustWithBarCount(bars) }
func piano() *instrument.Instrument         { return instrument.MustNew(instrument.Piano()) }
func fixture(bars int) (*track.Track, track.VoiceID) {
	p := track.MustNew(piano(), grid(bars))
	return p, p.MustVoices()[0]
}
func parsed(text string) note.Note { return note.MustParsePitch(settings.Standard(), text) }
func written(base duration.Base, dots int) event.Duration {
	return event.MustWrittenDuration(duration.MustNewWritten(base, dots))
}
func value(n, d int) event.Duration { return event.MustValueDuration(fraction.MustNew(n, d)) }
func sounded(n note.Note, d event.Duration, marks event.Marks) *event.Event {
	return event.MustNew(event.MustNoteContent(n), d, marks)
}
func tone(text string, base duration.Base, marks event.Marks) *event.Event {
	return sounded(parsed(text), written(base, 0), marks)
}
func nameTone(text string, base duration.Base, marks event.Marks) *event.Event {
	return sounded(note.MustParse(settings.Standard(), text), written(base, 0), marks)
}
func chordEvent(d event.Duration, marks event.Marks, ns ...note.Note) *event.Event {
	return event.MustNew(event.MustChordContent(chord.MustNew(ns...)), d, marks)
}
func rest(d event.Duration, fermata bool) *event.Event {
	return event.MustNew(event.RestContent(), d, event.Marks{Fermata: fermata})
}
func measureRest(bars int, fermata bool) *event.Event {
	return rest(event.MustMeasureDuration(bars), fermata)
}
func clockFor(g measure.Grid, bpm int) track.Clock {
	return track.Clock{Map: tempomap.MustNewMeasured(g, tempo.MustNewQuarter(fraction.FromInt(bpm)))}
}
func withClock(p *track.Track, c track.Clock) *track.Track { return p.MustEdit(track.Edit{Clock: &c}) }
func withExtension(c track.Clock, g measure.Grid, a measure.Address, n, d int) track.Clock {
	c.Map = c.Map.MustWithExtension(tempomap.MustNewExtension(tempomap.MustAtBar(a), tempo.MustNewSeconds(n, d)), g)
	return c
}
func assertPosition(t *testing.T, n, d int, got duration.Position) {
	t.Helper()
	assert.True(t, position(n, d).Equal(got), "position: want %d/%d, got %v", n, d, got)
}
func assertLength(t *testing.T, n, d int, got duration.Duration) {
	t.Helper()
	assert.True(t, length(n, d).Equal(got), "length: want %d/%d, got %v", n, d, got)
}
func assertSpan(t *testing.T, got track.Span, sn, sd, en, ed int) {
	t.Helper()
	assertPosition(t, sn, sd, got.Start)
	assertPosition(t, en, ed, got.End)
}
func assertNote(t *testing.T, want, got note.Note) {
	t.Helper()
	assert.True(t, want.Equal(got), "note record: want %v, got %v", want, got)
	if want.Source() == nil {
		assert.Nil(t, got.Source())
		return
	}
	assert.True(t, settings.SameSource(want.Source(), got.Source()), "note source changed")
}
func assertEntry(t *testing.T, got track.Entry, v track.VoiceID, bar, index int, e *event.Event, sn, sd, en, ed int) {
	t.Helper()
	assertPlace(t, track.Place{Voice: v, Bar: bar, Index: index}, got.Place)
	require.NotNil(t, got.Event)
	assertEvent(t, e, got.Event)
	assertSpan(t, got.Time, sn, sd, en, ed)
}

func assertPlace(t *testing.T, want, got track.Place) {
	t.Helper()
	assert.True(t, want.Voice == got.Voice, "voice identity changed")
	assert.Equal(t, want.Bar, got.Bar)
	assert.Equal(t, want.Index, got.Index)
}

func assertOccurrences(t *testing.T, want, got []track.Occurrence) {
	t.Helper()
	require.Len(t, got, len(want))
	for i, o := range want {
		assertPlace(t, o.Place, got[i].Place)
		assert.Equal(t, o.NoteIndex, got[i].NoteIndex)
	}
}

func assertVoices(t *testing.T, want, got []track.VoiceID) {
	t.Helper()
	require.Len(t, got, len(want))
	for i, v := range want {
		assert.True(t, v == got[i], "voice identity changed at index %d", i)
	}
}

func assertEvent(t *testing.T, want, got *event.Event) {
	t.Helper()
	require.NotNil(t, want)
	require.NotNil(t, got)
	assert.True(t, want.Equal(got))
	wn, gn := want.MustContent().MustNotes(), got.MustContent().MustNotes()
	require.Len(t, gn, len(wn))
	for i, n := range wn {
		assertNote(t, n, gn[i])
	}
}

func expectedEntry(v track.VoiceID, bar, index int, e *event.Event, sn, sd, en, ed int) track.Entry {
	return track.Entry{Place: track.Place{Voice: v, Bar: bar, Index: index}, Event: e, Time: track.Span{Start: position(sn, sd), End: position(en, ed)}}
}

// Check the recovered error and preserve the addressed details of the paired
// ordinary operation. A panic string or a runtime error cannot satisfy this.
func assertMustError(t *testing.T, ordinary, cause error, call func()) {
	t.Helper()
	require.ErrorIs(t, ordinary, cause)
	var caught any
	func() { defer func() { caught = recover() }(); call() }()
	got, ok := caught.(error)
	require.True(t, ok, "Must panic must carry error, got %T", caught)
	require.ErrorIs(t, got, cause)
	wantProblems, wantOK := errors.AsType[*track.Problems](ordinary)
	gotProblems, gotOK := errors.AsType[*track.Problems](got)
	require.Equal(t, wantOK, gotOK, "Must lost addressed error details")
	if !wantOK {
		return
	}
	require.Len(t, gotProblems.Issues, len(wantProblems.Issues))
	for i, want := range wantProblems.Issues {
		actual := gotProblems.Issues[i]
		assert.ErrorIs(t, actual.Cause, cause)
		assertPlace(t, want.Place, actual.Place)
		assert.Equal(t, want.NoteIndex, actual.NoteIndex)
		assert.Equal(t, want.InputIndex, actual.InputIndex)
		assert.Equal(t, want.AfterBar, actual.AfterBar)
		assert.Equal(t, want.BarsNeeded, actual.BarsNeeded)
		assert.Equal(t, want.Kind, actual.Kind)
		assert.Equal(t, want.Slot, actual.Slot)
		assert.Equal(t, want.Position.IsSet(), actual.Position.IsSet())
		if want.Position.IsSet() {
			assert.True(t, want.Position.Equal(actual.Position))
		}
		assert.Equal(t, want.Amount.IsSet(), actual.Amount.IsSet())
		if want.Amount.IsSet() {
			assert.True(t, want.Amount.Equal(actual.Amount))
		}
		assert.Equal(t, want.Address.IsSet(), actual.Address.IsSet())
		if want.Address.IsSet() {
			assert.True(t, want.Address.Equal(actual.Address))
		}
	}
}
func issues(t *testing.T, err, cause error) []track.Issue {
	t.Helper()
	require.ErrorIs(t, err, cause)
	p, ok := errors.AsType[*track.Problems](err)
	require.True(t, ok, "addressed failure must expose Problems")
	require.NotNil(t, p)
	require.NotEmpty(t, p.Issues)
	return p.Issues
}
func noProposal(t *testing.T, p track.Proposal, err, cause error) []track.Issue {
	t.Helper()
	assert.Nil(t, p.Track)
	assert.Empty(t, p.Changes)
	assert.Empty(t, p.Marks)
	return issues(t, err, cause)
}
func assertSeconds(t *testing.T, n, d int, got tempomap.SecondsResult) {
	t.Helper()
	assert.True(t, tempo.MustNewSeconds(n, d).Equal(got.MustValue()), "seconds: got %v", got)
	assert.False(t, got.MustApproximate())
	assert.True(t, tempo.ZeroSeconds().Equal(got.MustErrorBound()))
}
