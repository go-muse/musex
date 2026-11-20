package voice_test

import (
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertEvents(t *testing.T, want, got []*event.Event) {
	t.Helper()
	require.NotNil(t, got)
	require.Len(t, got, len(want))
	for i := range want {
		assert.True(t, want[i].Equal(got[i]), "event %d: want %v, got %v", i, want[i], got[i])
	}
}

// КП-27, П-13, ИЛ-7: a tied row becomes one sound; the reverse conversion
// gives the original events.
func TestEventsTiedRow(t *testing.T) {
	t.Parallel()
	events := []*event.Event{
		noteEvent(c4(), duration.Quarter, tie()),
		noteEvent(c4(), duration.Eighth, event.Marks{}),
		noteEvent(d4(), duration.Half, event.Marks{Fermata: true}),
	}
	got, err := voice.FromEvents(events...)
	require.NoError(t, err)
	want := line(snd(c4(), qe()), snd(d4(), h()).MustWithFermata(true))
	assertLine(t, want, got)
	assertLine(t, want, voice.MustFromEvents(events...))
	assertRoles(t, got, "", voice.PlacementUnset, voice.ThematicUnset, voice.ImportanceUnset)

	back, err := got.Events()
	require.NoError(t, err)
	assertEvents(t, events, back)
	assertEvents(t, events, got.MustEvents())
}

// П-13: conversions of every kind of element.
func TestEventsConversion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		events []*event.Event
		want   *voice.Line
	}{
		{"empty", []*event.Event{}, line()},
		{"note and rest", []*event.Event{noteEvent(c4(), duration.Quarter, event.Marks{}), restEvent(duration.Eighth, event.Marks{})},
			line(snd(c4(), q()), rst(e8()))},
		{"values", []*event.Event{valueNoteEvent(c4(), 1, 3, event.Marks{Fermata: true}),
			event.MustNew(event.RestContent(), event.MustValueDuration(frac(1, 5)), event.Marks{Fermata: true})},
			line(snd(c4(), val(1, 3)).MustWithFermata(true), rst(val(1, 5)).MustWithFermata(true))},
		{"row of three with a fermata on the last", []*event.Event{
			noteEvent(c4(), duration.Half, tie()), noteEvent(c4(), duration.Quarter, tie()),
			noteEvent(c4(), duration.Eighth, event.Marks{Fermata: true})},
			line(snd(c4(), voice.MustTiedDuration(duration.MustNewTied(wr(duration.Half, 0), wr(duration.Quarter, 0), wr(duration.Eighth, 0)))).
				MustWithFermata(true))},
		{"repeated note without a tie", []*event.Event{noteEvent(c4(), duration.Quarter, event.Marks{}), noteEvent(c4(), duration.Quarter, event.Marks{})},
			line(snd(c4(), q()), snd(c4(), q()))},
		{"two rows", []*event.Event{noteEvent(c4(), duration.Quarter, tie()), noteEvent(c4(), duration.Quarter, event.Marks{}),
			noteEvent(c4(), duration.Quarter, tie()), noteEvent(c4(), duration.Eighth, event.Marks{})},
			line(snd(c4(), voice.MustTiedDuration(duration.MustNewTied(wr(duration.Quarter, 0), wr(duration.Quarter, 0)))), snd(c4(), qe()))},
		{"dotted values are kept", []*event.Event{noteEvent(e4(), duration.Quarter, event.Marks{}),
			event.MustNew(event.MustNoteContent(e4()), event.MustWrittenDuration(wr(duration.Quarter, 1)), event.Marks{})},
			line(snd(e4(), q()), snd(e4(), dq()))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := voice.FromEvents(tt.events...)
			require.NoError(t, err)
			assertLine(t, tt.want, got)
			back := got.MustEvents()
			assertEvents(t, tt.events, back)
		})
	}
}

// The note of a tied row is the note of its first event: the correction and
// the source of that note.
func TestEventsRowKeepsFirstNote(t *testing.T) {
	t.Parallel()
	first := withCents(spelling.C(), 4, 7)
	got := voice.MustFromEvents(
		noteEvent(first, duration.Quarter, tie()),
		noteEvent(withCents(spelling.C(), 4, 7), duration.Eighth, event.Marks{}),
	)
	n, ok := got.MustElement(0).MustNote()
	require.True(t, ok)
	assert.True(t, first.Equal(n))
	assert.True(t, first.SameSource(n))
}

// Events of a line: a tied sound gives a row with ties on every event but
// the last and the fermata on the last; events are new values.
func TestLineEvents(t *testing.T) {
	t.Parallel()
	l := line(snd(c4(), qe()).MustWithFermata(true), rst(h()).MustWithFermata(true), snd(d4(), val(1, 3)))
	got, err := l.Events()
	require.NoError(t, err)
	want := []*event.Event{
		noteEvent(c4(), duration.Quarter, tie()),
		noteEvent(c4(), duration.Eighth, event.Marks{Fermata: true}),
		restEvent(duration.Half, event.Marks{Fermata: true}),
		valueNoteEvent(d4(), 1, 3, event.Marks{}),
	}
	assertEvents(t, want, got)
	again := l.MustEvents()
	for i := range got {
		assert.NotSame(t, got[i], again[i], "each call gives new events")
	}

	empty, err := line().Events()
	require.NoError(t, err)
	require.NotNil(t, empty)
	assert.Empty(t, empty)
}

// КП-28…КП-30 and the order of the checks of architecture 6.13.
func TestFromEventsFailures(t *testing.T) {
	t.Parallel()
	c, ce := noteEvent(c4(), duration.Quarter, event.Marks{}), noteEvent(c4(), duration.Eighth, event.Marks{})
	tied := noteEvent(c4(), duration.Quarter, tie())
	chordEvent := event.MustNew(event.MustChordContent(chord.MustNew(c4(), e4())), event.MustWrittenDuration(wr(duration.Quarter, 0)), event.Marks{})
	oneNoteChord := event.MustNew(event.MustChordContent(chord.MustNew(c4())), event.MustWrittenDuration(wr(duration.Quarter, 0)), event.Marks{})
	measureRest := event.MustNew(event.RestContent(), event.MustMeasureDuration(1), event.Marks{})
	grace := noteEvent(c4(), duration.Eighth, event.Marks{Grace: true})
	octaveless := noteEvent(noOctave(), duration.Quarter, event.Marks{})
	tests := []struct {
		name   string
		events []*event.Event
		want   want
	}{
		{"КП-28 tie to another note", []*event.Event{tied, noteEvent(d4(), duration.Quarter, event.Marks{})}, failsAt(0, voice.ErrBrokenTie).only()},
		{"КП-28 tie to an enharmonic note", []*event.Event{noteEvent(nt(spelling.CSharp(), 4), duration.Quarter, tie()),
			noteEvent(nt(spelling.DFlat(), 4), duration.Quarter, event.Marks{})}, failsAt(0, voice.ErrBrokenTie).only()},
		{"КП-28 tie to a rest", []*event.Event{tied, restEvent(duration.Quarter, event.Marks{})}, failsAt(0, voice.ErrBrokenTie).only()},
		{"КП-28 tie on the last event", []*event.Event{c, tied}, failsAt(1, voice.ErrBrokenTie).only()},
		{"tie to a note with another correction", []*event.Event{tied, noteEvent(withCents(spelling.C(), 4, 5), duration.Quarter, event.Marks{})},
			failsAt(0, voice.ErrBrokenTie).only()},
		{"tie to a chord", []*event.Event{tied, oneNoteChord}, failsAt(0, voice.ErrBrokenTie).only()},
		{"tie to an unset event is checked first", []*event.Event{tied, nil}, failsAt(0, voice.ErrBrokenTie).only()},
		{"КП-29 tied duration without notation", []*event.Event{valueNoteEvent(c4(), 1, 4, tie()), ce}, failsAt(0, voice.ErrTiedValue).only()},
		{"value duration at the end of a row", []*event.Event{tied, valueNoteEvent(c4(), 1, 8, event.Marks{})}, failsAt(1, voice.ErrTiedValue).only()},
		{"КП-29 fermata inside a tie", []*event.Event{noteEvent(c4(), duration.Quarter, event.Marks{Ties: []int{0}, Fermata: true}), ce},
			failsAt(0, voice.ErrFermataInTie).only()},
		{"Е-14 before Е-15", []*event.Event{valueNoteEvent(c4(), 1, 4, event.Marks{Ties: []int{0}, Fermata: true}), ce},
			failsAt(0, voice.ErrTiedValue).only()},
		{"Е-15 before Е-13", []*event.Event{noteEvent(c4(), duration.Quarter, event.Marks{Ties: []int{0}, Fermata: true}),
			noteEvent(d4(), duration.Quarter, event.Marks{})}, failsAt(0, voice.ErrFermataInTie).only()},
		{"КП-30 chord", []*event.Event{c, chordEvent}, failsAt(1, voice.ErrChord).only()},
		{"КП-30 one-note chord", []*event.Event{oneNoteChord}, failsAt(0, voice.ErrChord).only()},
		{"КП-30 rest in measures", []*event.Event{c, ce, measureRest}, failsAt(2, voice.ErrMeasureDuration).only()},
		{"КП-30 grace note", []*event.Event{grace, c}, failsAt(0, voice.ErrGrace).only()},
		{"КП-30 note without octave", []*event.Event{octaveless}, failsAt(0, voice.ErrNoOctave).only()},
		{"unset event", []*event.Event{c, nil}, failsAt(1, voice.ErrUnset, event.ErrUnset)},
		{"zero event", []*event.Event{{}}, failsAt(0, voice.ErrUnset, event.ErrUnset)},
		{"chord before grace", []*event.Event{event.MustNew(event.MustChordContent(chord.MustNew(c4())), event.MustWrittenDuration(wr(duration.Eighth, 0)),
			event.Marks{Grace: true})}, failsAt(0, voice.ErrChord).only()},
		{"grace before no octave", []*event.Event{noteEvent(noOctave(), duration.Eighth, event.Marks{Grace: true})}, failsAt(0, voice.ErrGrace).only()},
		{"no octave before a broken tie", []*event.Event{noteEvent(noOctave(), duration.Quarter, tie()), c}, failsAt(0, voice.ErrNoOctave).only()},
		{"events are checked in order", []*event.Event{c, octaveless, chordEvent}, failsAt(1, voice.ErrNoOctave).only()},
		{"event k before event k+1", []*event.Event{tied, ce, grace, measureRest}, failsAt(2, voice.ErrGrace).only()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertFailure(t, tt.want,
				func() (*voice.Line, error) { return voice.FromEvents(tt.events...) },
				func() *voice.Line { return voice.MustFromEvents(tt.events...) })
		})
	}
}

// The overflow of a line built from events has no *voice.ItemError and comes
// after the checks of all events (architecture 6.1, 6.13).
func TestFromEventsOverflow(t *testing.T) {
	t.Parallel()
	invP, _, invQ := overflowDurations()
	first := event.MustNew(event.RestContent(), event.MustValueDuration(invP.MustFraction()), event.Marks{})
	second := event.MustNew(event.RestContent(), event.MustValueDuration(invQ.MustFraction()), event.Marks{})
	assertFailure(t, overflowWant(),
		func() (*voice.Line, error) { return voice.FromEvents(first, second) },
		func() *voice.Line { return voice.MustFromEvents(first, second) })
	assertFailure(t, failsAt(2, voice.ErrUnset, event.ErrUnset),
		func() (*voice.Line, error) { return voice.FromEvents(first, second, nil) },
		func() *voice.Line { return voice.MustFromEvents(first, second, nil) })
}

// ИЛ-7: Events, then FromEvents gives a line equal by the record of its
// elements.
func TestEventsRoundTrip(t *testing.T) {
	t.Parallel()
	for name, l := range invariantLines() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			src := named(l)
			back, err := voice.FromEvents(src.MustEvents()...)
			require.NoError(t, err)
			assert.True(t, src.EqualElements(back), "want %v, got %v", src, back)
			assertRoles(t, back, "", voice.PlacementUnset, voice.ThematicUnset, voice.ImportanceUnset)
		})
	}
}
