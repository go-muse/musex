package score_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/score"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pos(n, d int) duration.Position { return duration.MustNewPosition(n, d) }
func dur(n, d int) duration.Duration { return duration.MustNew(n, d) }
func at(n, d int) tempomap.Address   { return tempomap.MustAt(pos(n, d)) }
func bpm(n int) tempo.Tempo          { return tempo.MustNewQuarter(fraction.MustNew(n, 1)) }
func clock60() tempomap.Map          { return tempomap.MustNew(bpm(60)) }
func grid(n, d, count int) measure.Grid {
	return measure.MustNew(meter.MustNew(n, d)).MustWithBarCount(count)
}
func emptyPart(n, d, count int) *track.Track {
	return track.MustNew(instrument.MustNew(instrument.Piano()), grid(n, d, count))
}
func tone(p *track.Track, pitch string, n, d int, marks event.Marks) *event.Event {
	return event.MustNew(event.MustNoteContent(note.MustParsePitch(p.MustInstrument(), pitch)), event.MustValueDuration(fraction.MustNew(n, d)), marks)
}
func rest(count int, fermata bool) *event.Event {
	return event.MustNew(event.RestContent(), event.MustMeasureDuration(count), event.Marks{Fermata: fermata})
}
func write(p *track.Track, bar int, events ...*event.Event) *track.Track {
	return p.MustSplice(p.MustVoices()[0], bar, 0, 0, events...)
}
func assemble(clock tempomap.Map, parts ...*track.Track) (*score.Score, []score.PartID) {
	s := score.MustNew(clock)
	ids := make([]score.PartID, len(parts))
	for i, p := range parts {
		s, ids[i] = s.MustAdd(p)
	}
	return s, ids
}
func assertPos(t *testing.T, want, got duration.Position) {
	t.Helper()
	assert.True(t, want.Equal(got), "position: want %s, got %s", want, got)
}

func assertKey[T comparable](t *testing.T, want, got T) {
	t.Helper()
	assert.True(t, want == got, "keys must have the same identity: want %v, got %v", want, got)
}

func assertKeys[T comparable](t *testing.T, want, got []T) {
	t.Helper()
	require.Len(t, got, len(want))
	for i, key := range want {
		assert.True(t, key == got[i], "key at index %d must have the same identity", i)
	}
}
func assertEvent(t *testing.T, want, got *event.Event) {
	t.Helper()
	require.NotNil(t, got)
	assert.True(t, want.Equal(got), "event record differs")
}
func assertSeconds(t *testing.T, got tempomap.SecondsResult, n, d int) {
	t.Helper()
	assert.True(t, got.MustValue().Equal(tempo.MustNewSeconds(n, d)))
	assert.False(t, got.MustApproximate())
	assert.True(t, got.MustErrorBound().Equal(tempo.ZeroSeconds()))
}
func problems(t *testing.T, err error) *score.Problems {
	t.Helper()
	require.Error(t, err)
	var p *score.Problems
	require.ErrorAs(t, err, &p)
	require.NotEmpty(t, p.Issues)
	return p
}
