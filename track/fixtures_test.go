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
func parsed(text string) note.Note { return note.MustParse(settings.Standard(), text) }
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
	assert.True(t, settings.SameSource(want.Source(), got.Source()), "note source changed")
}
func assertEntry(t *testing.T, got track.Entry, v track.VoiceID, bar, index int, e *event.Event, sn, sd, en, ed int) {
	t.Helper()
	assert.Equal(t, track.Place{Voice: v, Bar: bar, Index: index}, got.Place)
	require.NotNil(t, got.Event)
	assert.True(t, e.Equal(got.Event))
	assertSpan(t, got.Time, sn, sd, en, ed)
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
