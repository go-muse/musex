package track_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/track"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleTrack_ExtractLine() {
	piano := instrument.MustNew(instrument.Piano())
	grid := measure.MustNew(meter.MustNew(4, 4)).MustWithBarCount(3)
	part := track.MustNew(piano, grid)
	upper := part.MustVoices()[0]
	quarter := event.MustWrittenDuration(duration.MustNewWritten(duration.Quarter, 0))
	e4 := note.MustParse(piano, "E4")
	e := event.MustNew(event.MustNoteContent(e4), quarter, event.Marks{})
	part = part.MustSplice(upper, 2, 0, 0, e)
	out, err := part.ExtractLine(upper, track.Exact)
	if err != nil {
		panic(err)
	}
	fmt.Println(out.Line.MustLen(), out.Line.MustStartOf(1), out.Line.MustEnd())
	fmt.Println(part.MustInstrument() == piano)
	// Output:
	// 2 1/1 5/4
	// true
}

func ExampleTrack_WriteLine() {
	piano := instrument.MustNew(instrument.Piano())
	part := track.MustNew(piano, measure.MustNew(meter.MustNew(3, 4)).MustWithBarCount(2))
	upper := part.MustVoices()[0]
	c4 := note.MustParse(piano, "C4")
	half := voice.MustWrittenDuration(duration.MustNewWritten(duration.Half, 0))
	line := voice.MustNew(voice.MustNewSound(c4, half))
	start := measure.MustNewAddress(1, duration.MustNewPosition(1, 2))
	proposal, err := part.WriteLine(upper, start, line, track.WriteOptions{})
	if err != nil {
		panic(err)
	}
	part = proposal.Track
	back, err := part.ExtractLine(upper, track.Exact)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(part.MustEntries(upper, 1)), len(part.MustEntries(upper, 2)))
	fmt.Println(back.Line.MustStartOf(1), back.Line.MustEnd())
	// Output:
	// 2 1
	// 1/2 1/1
}

func TestExtractLineGapsMeasuredRestsAndNoTrailingSilence(t *testing.T) {
	t.Parallel()
	p, v := fixture(6)
	q := tone("E4", duration.Quarter, event.Marks{})
	p = p.MustSplice(v, 2, 0, 0, q).MustSplice(v, 4, 0, 0, measureRest(2, true))
	p = p.MustEdit(track.Edit{Barlines: []track.BarlineEdit{{AfterBar: 1, Fermata: true}, {AfterBar: 5, Fermata: true}, {AfterBar: 6, Fermata: true}}})
	c := clockFor(grid(6), 60)
	c = withExtension(c, grid(6), address(2, 1, 4), 1, 1)
	c.BarHolds = []track.BarHold{{AfterBar: 5, Seconds: tempo.MustNewSeconds(1, 1)}, {AfterBar: 6, Seconds: tempo.MustNewSeconds(1, 1)}}
	p = withClock(p, c)
	out, err := p.ExtractLine(v, track.Exact)
	require.NoError(t, err)
	require.NotNil(t, out.Line)
	assert.Equal(t, 4, out.Line.MustLen())
	assertPosition(t, 5, 1, out.Line.MustEnd())
	assert.True(t, out.Line.MustElement(0).IsRest())
	assertLength(t, 1, 1, out.Line.MustElement(0).MustLength())
	assertNote(t, parsed("E4"), out.Line.MustNotes()[0])
	assertPosition(t, 1, 1, out.Line.MustStartOf(1))
	assertPosition(t, 5, 4, out.Line.MustEndOf(1))
	assertLength(t, 7, 4, out.Line.MustElement(2).MustLength())
	assertLength(t, 2, 1, out.Line.MustElement(3).MustLength())
	assert.True(t, out.Line.MustElement(3).MustFermata())
	require.Len(t, out.Gaps, 2)
	assertSpan(t, out.Gaps[0], 0, 1, 1, 1)
	assertSpan(t, out.Gaps[1], 5, 4, 3, 1)
	require.Len(t, out.Measured, 1)
	assert.Equal(t, track.Place{Voice: v, Bar: 4, Index: 0}, out.Measured[0].Place)
	assert.Equal(t, []int{1, 5}, out.FermataBars)
	require.Len(t, out.Extensions, 2)
	assertPosition(t, 5, 4, out.Extensions[0])
	assertPosition(t, 5, 1, out.Extensions[1])
	assert.Empty(t, out.Line.MustName())
	assert.Equal(t, voice.PlacementUnset, out.Line.MustPlacement())
	assert.Equal(t, voice.ThematicUnset, out.Line.MustThematic())
	assert.Equal(t, voice.ImportanceUnset, out.Line.MustImportance())
	assert.Empty(t, out.Skipped)
	out.Gaps[0].End = position(99, 1)
	out.Measured[0].Place.Bar = 99
	out.FermataBars[0] = 99
	out.Extensions[0] = position(99, 1)
	clean := p.MustExtractLine(v, track.Exact)
	assertSpan(t, clean.Gaps[0], 0, 1, 1, 1)
	assert.Equal(t, 4, clean.Measured[0].Place.Bar)
	assert.Equal(t, []int{1, 5}, clean.FermataBars)
	assertPosition(t, 5, 4, clean.Extensions[0])
	// A recorded silence extends the line; an empty tail does not.
	empty, ev := fixture(6)
	eo := empty.MustExtractLine(ev, track.Exact)
	assert.True(t, eo.Line.IsSet())
	assert.True(t, eo.Line.MustIsEmpty())
	assert.Empty(t, eo.Gaps)
	assert.Empty(t, eo.Measured)
	assert.Empty(t, eo.Extensions)
	assert.Empty(t, eo.FermataBars)
}

func TestExtractLineKeepsAdjacentExplicitRestsSeparate(t *testing.T) {
	t.Parallel()
	p, v := fixture(3)
	p = p.MustSplice(v, 1, 0, 0, rest(written(duration.Quarter, 0), false), rest(value(1, 4), true)).MustSplice(v, 2, 0, 0, measureRest(1, false))
	out := p.MustExtractLine(v, track.Exact)
	assert.Equal(t, 4, out.Line.MustLen())
	require.Len(t, out.Gaps, 1)
	assertSpan(t, out.Gaps[0], 1, 2, 1, 1)
	assert.True(t, out.Line.MustElement(1).MustFermata())
	assertPosition(t, 2, 1, out.Line.MustEnd())
}

func TestExtractionFailuresRetainOriginalAddresses(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		events []*event.Event
		cause  error
	}{
		{"one note chord", []*event.Event{chordEvent(written(duration.Quarter, 0), event.Marks{}, parsed("C4"))}, voice.ErrChord},
		{"chord", []*event.Event{chordEvent(written(duration.Quarter, 0), event.Marks{}, parsed("C4"), parsed("E4"))}, voice.ErrChord},
		{"no octave", []*event.Event{tone("C", duration.Quarter, event.Marks{})}, voice.ErrNoOctave},
		{"grace", []*event.Event{tone("C4", duration.Eighth, event.Marks{Grace: true}), tone("D4", duration.Quarter, event.Marks{})}, voice.ErrGrace},
		{"intermediate fermata", []*event.Event{tone("C4", duration.Quarter, event.Marks{Ties: []int{0}, Fermata: true}), tone("C4", duration.Quarter, event.Marks{})}, voice.ErrFermataInTie},
		{"numeric tie", []*event.Event{sounded(parsed("C4"), value(1, 4), event.Marks{Ties: []int{0}}), tone("C4", duration.Quarter, event.Marks{})}, voice.ErrTiedValue},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, v := fixture(3)
			p = p.MustSplice(v, 2, 0, 0, tc.events...)
			out, err := p.ExtractLine(v, track.Exact)
			assert.Nil(t, out.Line)
			assert.Empty(t, out.Gaps)
			assert.Empty(t, out.Measured)
			assert.Empty(t, out.Skipped)
			is := issues(t, err, tc.cause)
			assert.Equal(t, v, is[0].Place.Voice)
			assert.Equal(t, 2, is[0].Place.Bar)
			assert.Equal(t, 0, is[0].Place.Index)
			if tc.name != "grace" {
				out, err = p.ExtractLine(v, track.PrincipalOnly)
				assert.Nil(t, out.Line)
				assert.ErrorIs(t, err, tc.cause)
			}
			assert.Len(t, p.MustEntries(v, 2), len(tc.events))
		})
	}
}

func TestPrincipalOnlySkipsOnlyGraceAndKeepsWrittenSigns(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	g1 := tone("C4", duration.Sixteenth, event.Marks{Grace: true})
	g2 := tone("D4", duration.Eighth, event.Marks{Grace: true, Ties: []int{0}})
	main := tone("D4", duration.Quarter, event.Marks{Fermata: true})
	p = p.MustSplice(v, 2, 0, 0, g1, g2, main)
	out := p.MustExtractLine(v, track.PrincipalOnly)
	require.Len(t, out.Skipped, 2)
	assert.Same(t, g1, out.Skipped[0].Event)
	assert.Same(t, g2, out.Skipped[1].Event)
	assert.Equal(t, track.Place{Voice: v, Bar: 2, Index: 1}, out.Skipped[1].Place)
	assert.Equal(t, 2, out.Line.MustLen())
	assertNote(t, parsed("D4"), out.Line.MustNotes()[0])
	assert.True(t, out.Line.MustElement(1).MustFermata())
	assertPosition(t, 5, 4, out.Line.MustEnd())
}

func TestCrossBarExtractionAndDifferentSourcesUseFirstLink(t *testing.T) {
	t.Parallel()
	a, b := piano(), piano()
	na, nb := note.MustParse(a, "C4"), note.MustParse(b, "C4")
	p := track.MustNew(a, grid(2))
	v := p.MustVoices()[0]
	first := sounded(na, written(duration.Whole, 0), event.Marks{Ties: []int{0}})
	last := sounded(nb, written(duration.Quarter, 0), event.Marks{Fermata: true})
	p = p.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: v, Bar: 1, Events: []*event.Event{first}}, {Voice: v, Bar: 2, Events: []*event.Event{last}}}})
	out := p.MustExtractLine(v, track.Exact)
	require.Equal(t, 1, out.Line.MustLen())
	assertNote(t, na, out.Line.MustNotes()[0])
	assertPosition(t, 5, 4, out.Line.MustEnd())
	assert.True(t, out.Line.MustElement(0).MustFermata())
	d := out.Line.MustElement(0).MustDuration()
	assert.Equal(t, voice.DurationTied, d.Kind())
	tied, ok := d.MustTied()
	require.True(t, ok)
	parts := tied.MustParts()
	require.Len(t, parts, 2)
	assert.True(t, parts[0].Equal(duration.MustNewWritten(duration.Whole, 0)))
	assert.True(t, parts[1].Equal(duration.MustNewWritten(duration.Quarter, 0)))
	dest := track.MustNew(a, grid(2))
	dv := dest.MustVoices()[0]
	back := dest.MustWriteLine(dv, address(1, 0, 1), out.Line, track.WriteOptions{})
	assertReport(t, dest, back)
	assertNote(t, na, back.Track.MustEntries(dv, 1)[0].Event.MustContent().MustNote())
	assertNote(t, na, back.Track.MustEntries(dv, 2)[0].Event.MustContent().MustNote())
	assertNote(t, na, p.MustEntries(v, 1)[0].Event.MustContent().MustNote())
	assertNote(t, nb, p.MustEntries(v, 2)[0].Event.MustContent().MustNote())
	assert.False(t, p.Equal(back.Track), "source B cannot be recovered from the single line note")
	// Every original link stays readable while the sounding note uses the first.
	assertNote(t, na, p.MustSounds()[0].Note)
}

func TestWriteLineOffsetRoundTripAndElementProvenance(t *testing.T) {
	t.Parallel()
	g := measure.MustNew(meter.MustNew(3, 4)).MustWithBarCount(2)
	p := track.MustNew(piano(), g)
	v := p.MustVoices()[0]
	p, w := p.MustAddVoice()
	low := tone("D3", duration.Half, event.Marks{})
	p = p.MustSplice(w, 1, 0, 0, low)
	n := parsed("C4")
	line := voice.MustNew(voice.MustNewSound(n, voice.MustWrittenDuration(duration.MustNewWritten(duration.Half, 0)))).MustWithName("Theme").MustWithPlacement(voice.Soprano).MustWithThematic(voice.Leading).MustWithImportance(voice.Principal)
	out, err := p.WriteLine(v, address(1, 1, 2), line, track.WriteOptions{})
	require.NoError(t, err)
	assertReport(t, p, out)
	e1, e2 := out.Track.MustEntries(v, 1), out.Track.MustEntries(v, 2)
	require.Len(t, e1, 2)
	require.Len(t, e2, 1)
	assert.Equal(t, event.KindRest, e1[0].Event.Kind())
	assertSpan(t, e1[0].Time, 0, 1, 1, 2)
	assertSpan(t, e1[1].Time, 1, 2, 3, 4)
	assertSpan(t, e2[0].Time, 3, 4, 1, 1)
	assert.Equal(t, []int{0}, e1[1].Event.MustMarks().Ties)
	assert.Empty(t, e2[0].Event.MustMarks().Ties)
	count := 0
	for _, c := range out.Changes {
		if len(c.Input) > 0 {
			assert.Equal(t, []int{0}, c.Input)
			count += len(c.To)
		}
	}
	assert.Equal(t, 2, count)
	back := out.Track.MustExtractLine(v, track.Exact)
	assert.Equal(t, 2, back.Line.MustLen())
	assertPosition(t, 1, 2, back.Line.MustStartOf(1))
	assertPosition(t, 1, 1, back.Line.MustEnd())
	assertNote(t, n, back.Line.MustNotes()[0])
	assert.Empty(t, back.Line.MustName())
	assert.Equal(t, "Theme", line.MustName())
	assert.Equal(t, voice.Soprano, line.MustPlacement())
	assert.Equal(t, voice.Leading, line.MustThematic())
	assert.Equal(t, voice.Principal, line.MustImportance())
	assertPosition(t, 1, 2, line.MustEnd())
	assert.Same(t, low, out.Track.MustEntries(w, 1)[0].Event)
	assert.Empty(t, p.MustEntries(v, 1))
	// Mutating nested report containers cannot alter either snapshot or reruns.
	for i := range out.Changes {
		if len(out.Changes[i].Input) > 0 {
			out.Changes[i].Input[0] = 99
		}
		if len(out.Changes[i].To) > 0 {
			out.Changes[i].To[0].Place.Index = 99
		}
	}
	again := p.MustWriteLine(v, address(1, 1, 2), line, track.WriteOptions{})
	assert.Len(t, again.Track.MustEntries(v, 1), 2)
	for _, c := range again.Changes {
		for _, i := range c.Input {
			assert.Equal(t, 0, i)
		}
	}
}

func TestNumericLineCrossingBecomesRepresentableWrittenTie(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	line := voice.MustNew(voice.MustNewSound(parsed("E4"), voice.MustValueDuration(length(1, 2))))
	out := p.MustWriteLine(v, address(1, 3, 4), line, track.WriteOptions{})
	assert.Equal(t, event.DurationWritten, out.Track.MustEntries(v, 1)[1].Event.MustDuration().Kind())
	assert.Equal(t, event.DurationWritten, out.Track.MustEntries(v, 2)[0].Event.MustDuration().Kind())
	back := out.Track.MustExtractLine(v, track.Exact)
	assert.Equal(t, 2, back.Line.MustLen())
	assertPosition(t, 3, 4, back.Line.MustStartOf(1))
	assertPosition(t, 5, 4, back.Line.MustEnd())
	assert.Equal(t, voice.DurationTied, back.Line.MustElement(1).MustDuration().Kind())
}

func TestWriteLineReportsElementIndicesAfterExpansion(t *testing.T) {
	t.Parallel()
	g := measure.MustNew(meter.MustNew(3, 4)).MustWithBarCount(2)
	p := track.MustNew(piano(), g)
	v := p.MustVoices()[0]
	a, b := piano(), piano()
	c, d := note.MustParse(a, "C4"), note.MustParse(b, "D4")
	line := voice.MustNew(
		voice.MustNewSound(c, voice.MustWrittenDuration(duration.MustNewWritten(duration.Half, 0))),
		voice.MustNewSound(d, voice.MustWrittenDuration(duration.MustNewWritten(duration.Quarter, 0))),
	)
	out := p.MustWriteLine(v, address(1, 1, 2), line, track.WriteOptions{})
	want := []track.Entry{
		expectedEntry(v, 1, 0, rest(value(1, 2), false), 0, 1, 1, 2),
		expectedEntry(v, 1, 1, sounded(c, written(duration.Quarter, 0), event.Marks{Ties: []int{0}}), 1, 2, 3, 4),
		expectedEntry(v, 2, 0, sounded(c, written(duration.Quarter, 0), event.Marks{}), 3, 4, 1, 1),
		expectedEntry(v, 2, 1, sounded(d, written(duration.Quarter, 0), event.Marks{}), 1, 1, 5, 4),
	}
	assertChanges(t, p, out, nil, want)
	assertInputs(t, out, map[track.Place]int{want[1].Place: 0, want[2].Place: 0, want[3].Place: 1})
	back := out.Track.MustExtractLine(v, track.Exact)
	assert.Equal(t, 3, back.Line.MustLen())
	assertNote(t, c, back.Line.MustNotes()[0])
	assertNote(t, d, back.Line.MustNotes()[1])
	assertPosition(t, 1, 1, back.Line.MustStartOf(2))
	assertPosition(t, 5, 4, back.Line.MustEnd())
	assertPosition(t, 3, 4, line.MustEnd())
}
