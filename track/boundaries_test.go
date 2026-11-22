package track_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModesAreValidatedEvenForEmptyInput(t *testing.T) {
	t.Parallel()
	p, v := fixture(1)
	for _, opts := range []track.WriteOptions{{Groups: track.Groups(2)}, {Mode: track.WriteMode(2)}} {
		out, err := p.WriteEvents(v, address(1, 0, 1), nil, opts)
		assert.Nil(t, out.Track)
		assert.ErrorIs(t, err, track.ErrInvalidArgument)
		out, err = p.WriteLine(v, address(1, 0, 1), voice.MustNew(), opts)
		assert.Nil(t, out.Track)
		assert.ErrorIs(t, err, track.ErrInvalidArgument)
	}
	_, err := p.Rebar(grid(1), track.Groups(2))
	assert.ErrorIs(t, err, track.ErrInvalidArgument)
	_, err = p.ExtractLine(v, track.LineMode(2))
	assert.ErrorIs(t, err, track.ErrInvalidArgument)
	for _, step := range []duration.Duration{{}, length(1, 8)} {
		mode := rounding.HalfUp
		if step.IsSet() {
			mode = rounding.Mode(0)
		}
		_, err = p.Quantize(v, 1, 1, step, mode, track.KeepGroups)
		assert.ErrorIs(t, err, track.ErrInvalidArgument)
	}
	_, err = p.Quantize(v, 1, 1, length(1, 8), rounding.HalfUp, track.Groups(2))
	assert.ErrorIs(t, err, track.ErrInvalidArgument)
	_, err = p.WriteEvents(v, measure.Address{}, nil, track.WriteOptions{})
	assert.Error(t, err)
	_, err = p.WriteLine(v, address(1, 0, 1), nil, track.WriteOptions{})
	assert.Error(t, err)
	// No successful empty result can conceal a bad timing precision/mode.
	p = withClock(p, clockFor(grid(1), 60))
	_, err = p.SoundSeconds(-1, rounding.HalfUp)
	assert.Error(t, err)
	_, err = p.SoundSeconds(3, rounding.Mode(0))
	assert.Error(t, err)
}

func TestMustWrappersPanicWithTheirError(t *testing.T) {
	t.Parallel()
	p, v := fixture(1)
	bad := track.VoiceID{}
	var unset *track.Track
	cases := []struct {
		name     string
		cause    error
		ordinary func() error
		must     func()
	}{
		{"instrument", track.ErrUnset, func() error { _, e := unset.Instrument(); return e }, func() { unset.MustInstrument() }},
		{"grid", track.ErrUnset, func() error { _, e := unset.Grid(); return e }, func() { unset.MustGrid() }},
		{"voices", track.ErrUnset, func() error { _, e := unset.Voices(); return e }, func() { unset.MustVoices() }},
		{"add voice", track.ErrUnset, func() error { _, _, e := unset.AddVoice(); return e }, func() { unset.MustAddVoice() }},
		{"edit", track.ErrInvalidArgument, func() error { _, e := p.Edit(track.Edit{Grid: &measure.Grid{}}); return e }, func() { p.MustEdit(track.Edit{Grid: &measure.Grid{}}) }},
		{"splice", track.ErrPlace, func() error { _, e := p.Splice(v, 1, -1, 0); return e }, func() { p.MustSplice(v, 1, -1, 0) }},
		{"entries", track.ErrVoice, func() error { _, e := p.Entries(bad, 1); return e }, func() { p.MustEntries(bad, 1) }},
		{"at", track.ErrPlace, func() error { _, _, e := p.At(v, position(2, 1)); return e }, func() { p.MustAt(v, position(2, 1)) }},
		{"voice end", track.ErrVoice, func() error { _, e := p.VoiceEnd(bad); return e }, func() { p.MustVoiceEnd(bad) }},
		{"fermata bars", track.ErrUnset, func() error { _, e := unset.FermataBars(); return e }, func() { unset.MustFermataBars() }},
		{"sounds", track.ErrVoice, func() error { _, e := p.Sounds(bad); return e }, func() { p.MustSounds(bad) }},
		{"sounding at", track.ErrPlace, func() error { _, e := p.SoundingAt(position(2, 1)); return e }, func() { p.MustSoundingAt(position(2, 1)) }},
		{"clock", track.ErrUnset, func() error { _, _, e := unset.Clock(); return e }, func() { unset.MustClock() }},
		{"performance map", track.ErrNoClock, func() error { _, e := p.PerformanceMap(); return e }, func() { p.MustPerformanceMap() }},
		{"sound seconds", track.ErrNoClock, func() error { _, e := p.SoundSeconds(3, rounding.HalfUp); return e }, func() { p.MustSoundSeconds(3, rounding.HalfUp) }},
		{"write events", track.ErrVoice, func() error { _, e := p.WriteEvents(bad, address(1, 0, 1), nil, track.WriteOptions{}); return e }, func() { p.MustWriteEvents(bad, address(1, 0, 1), nil, track.WriteOptions{}) }},
		{"rebar", track.ErrInvalidArgument, func() error { _, e := p.Rebar(measure.Grid{}, track.KeepGroups); return e }, func() { p.MustRebar(measure.Grid{}, track.KeepGroups) }},
		{"quantize", track.ErrInvalidArgument, func() error {
			_, e := p.Quantize(v, 1, 1, duration.Duration{}, rounding.HalfUp, track.KeepGroups)
			return e
		}, func() { p.MustQuantize(v, 1, 1, duration.Duration{}, rounding.HalfUp, track.KeepGroups) }},
		{"split", track.ErrPlace, func() error {
			_, e := p.Split(track.Place{Voice: v, Bar: 1, Index: 0}, length(1, 8), track.KeepGroups)
			return e
		}, func() { p.MustSplit(track.Place{Voice: v, Bar: 1, Index: 0}, length(1, 8), track.KeepGroups) }},
		{"merge", track.ErrPlace, func() error { _, e := p.Merge(v, 1, 0, 2, track.KeepGroups); return e }, func() { p.MustMerge(v, 1, 0, 2, track.KeepGroups) }},
		{"pack", track.ErrPlace, func() error { _, e := p.PackRests(v, 0, 1); return e }, func() { p.MustPackRests(v, 0, 1) }},
		{"unpack", track.ErrPlace, func() error { _, e := p.UnpackRest(track.Place{Voice: v, Bar: 1, Index: 0}); return e }, func() { p.MustUnpackRest(track.Place{Voice: v, Bar: 1, Index: 0}) }},
		{"extract", track.ErrVoice, func() error { _, e := p.ExtractLine(bad, track.Exact); return e }, func() { p.MustExtractLine(bad, track.Exact) }},
		{"write line", track.ErrUnset, func() error { _, e := unset.WriteLine(bad, address(1, 0, 1), nil, track.WriteOptions{}); return e }, func() { unset.MustWriteLine(bad, address(1, 0, 1), nil, track.WriteOptions{}) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { assertMustError(t, tc.ordinary(), tc.cause, tc.must) })
	}
}

func TestEqualityIncludesSourcesNotationGridClockAndSigns(t *testing.T) {
	t.Parallel()
	a, b := piano(), piano()
	base := track.MustNew(a, grid(2))
	v := base.MustVoices()[0]
	na, nb := note.MustParsePitch(a, "C4"), note.MustParsePitch(b, "C4")
	p := base.MustSplice(v, 1, 0, 0, sounded(na, written(duration.Quarter, 0), event.Marks{}))
	same := base.MustSplice(v, 1, 0, 0, sounded(na, written(duration.Quarter, 0), event.Marks{}))
	assert.True(t, p.Equal(same))
	for _, other := range []*track.Track{
		base.MustSplice(v, 1, 0, 0, sounded(nb, written(duration.Quarter, 0), event.Marks{})),
		base.MustSplice(v, 1, 0, 0, sounded(na, value(1, 4), event.Marks{})),
		p.MustEdit(track.Edit{Barlines: []track.BarlineEdit{{AfterBar: 1, Fermata: true}}}),
		withClock(p, clockFor(grid(2), 60)),
		p.MustSplice(v, 1, 0, 1, sounded(na, written(duration.Quarter, 0), event.Marks{Fermata: true})),
	} {
		assert.False(t, p.Equal(other))
		assert.False(t, other.Equal(p))
	}
	assert.False(t, p.Equal(nil))
	assert.False(t, base.Equal(p))
}

func TestSplittingNumericChordKeepsSourcesAndEveryMark(t *testing.T) {
	t.Parallel()
	a, b := piano(), piano()
	c, e := note.MustParsePitch(a, "C4"), note.MustParsePitch(b, "E4")
	p, v := fixture(3)
	first := chordEvent(value(3, 2), event.Marks{Arpeggio: true, Ties: []int{1}, Fermata: true}, c, e)
	last := sounded(e, written(duration.Quarter, 0), event.Marks{})
	out, err := p.WriteEvents(v, address(1, 0, 1), []*event.Event{first, last}, track.WriteOptions{})
	require.NoError(t, err)
	assertReport(t, p, out)
	x, y := out.Track.MustEntries(v, 1)[0], out.Track.MustEntries(v, 2)
	require.Len(t, y, 2)
	assert.Equal(t, event.DurationWritten, x.Event.MustDuration().Kind())
	assert.Equal(t, []int{0, 1}, x.Event.MustMarks().Ties)
	assert.True(t, x.Event.MustMarks().Arpeggio)
	assert.False(t, x.Event.MustMarks().Fermata)
	assert.Equal(t, []int{1}, y[0].Event.MustMarks().Ties)
	assert.False(t, y[0].Event.MustMarks().Arpeggio)
	assert.True(t, y[0].Event.MustMarks().Fermata)
	for _, entry := range []track.Entry{x, y[0]} {
		ns := entry.Event.MustContent().MustNotes()
		assertNote(t, c, ns[0])
		assertNote(t, e, ns[1])
	}
	s := out.Track.MustSounds()
	require.Len(t, s, 2)
	assertSpan(t, s[0].Time, 0, 1, 3, 2)
	assertSpan(t, s[1].Time, 0, 1, 7, 4)
	assert.Len(t, s[1].Parts, 3)
}

func TestRestSplittingAndNonSplittingKeepWrittenForm(t *testing.T) {
	t.Parallel()
	for _, d := range []event.Duration{value(1, 2), written(duration.Half, 0)} {
		p, v := fixture(2)
		r := rest(d, true)
		out := p.MustWriteEvents(v, address(1, 3, 4), []*event.Event{r}, track.WriteOptions{})
		es := []track.Entry{out.Track.MustEntries(v, 1)[1], out.Track.MustEntries(v, 2)[0]}
		for i, e := range es {
			assert.Equal(t, event.KindRest, e.Event.Kind())
			assert.Empty(t, e.Event.MustMarks().Ties)
			assert.Equal(t, i == 1, e.Event.MustMarks().Fermata)
			assert.Equal(t, d.Kind(), e.Event.MustDuration().Kind())
		}
		assert.Empty(t, out.Track.MustSounds())
	}
	p, v := fixture(1)
	dotted := sounded(parsed("C4"), written(duration.Quarter, 1), event.Marks{Fermata: true})
	out := p.MustWriteEvents(v, address(1, 0, 1), []*event.Event{dotted}, track.WriteOptions{Groups: track.WithoutGroups})
	assert.Same(t, dotted, out.Track.MustEntries(v, 1)[0].Event)
}

func TestMergeRefusesDifferentSourcesOrAdditionalChordAttack(t *testing.T) {
	t.Parallel()
	a, b := piano(), piano()
	na, nb := note.MustParsePitch(a, "C4"), note.MustParsePitch(b, "C4")
	for _, pair := range [][]*event.Event{
		{sounded(na, written(duration.Quarter, 0), event.Marks{Ties: []int{0}}), sounded(nb, written(duration.Quarter, 0), event.Marks{})},
		{chordEvent(written(duration.Quarter, 0), event.Marks{Ties: []int{0}}, na, parsed("E4")), chordEvent(written(duration.Quarter, 0), event.Marks{}, na, parsed("E4"))},
	} {
		p, v := fixture(1)
		p = p.MustSplice(v, 1, 0, 0, pair...)
		out, err := p.Merge(v, 1, 0, 2, track.KeepGroups)
		assert.Nil(t, out.Track)
		assert.Error(t, err)
		assert.Len(t, p.MustEntries(v, 1), 2)
	}
}

func TestQuantizeGraceFollowsAttackAndContextRestsStayWhole(t *testing.T) {
	t.Parallel()
	p, v := fixture(3)
	r := rest(value(3, 16), false)
	g := tone("D4", duration.Sixteenth, event.Marks{Grace: true})
	n := sounded(parsed("E4"), value(5, 16), event.Marks{})
	p = p.MustSplice(v, 1, 0, 0, r, g, n).MustSplice(v, 2, 0, 0, measureRest(2, false))
	out := p.MustQuantize(v, 1, 2, length(1, 8), rounding.HalfUp, track.KeepGroups)
	e := out.Track.MustEntries(v, 1)
	require.Len(t, e, 3)
	assert.Same(t, g, e[1].Event)
	assertSpan(t, e[1].Time, 1, 4, 1, 4)
	assertSpan(t, e[2].Time, 1, 4, 1, 2)
	assert.Equal(t, 2, out.Track.MustEntries(v, 2)[0].Event.MustDuration().MustMeasureCount())
	assert.Empty(t, out.Track.MustEntries(v, 3))
	assertPosition(t, 3, 1, out.Track.MustVoiceEnd(v))
}

func TestRebarFinalExtensionAndOrdinaryEditAddressBinding(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	p = p.MustSplice(v, 1, 0, 0, tone("C4", duration.Quarter, event.Marks{}))
	c := withExtension(clockFor(grid(2), 60), grid(2), address(2, 0, 1), 1, 1)
	p = withClock(p, c)
	out := p.MustRebar(grid(1), track.KeepGroups)
	oc, ok := out.Track.MustClock()
	require.True(t, ok)
	require.Len(t, oc.BarHolds, 1)
	assert.Equal(t, 1, oc.BarHolds[0].AfterBar)
	assertSeconds(t, 5, 1, out.Track.MustPerformanceMap().MustEnd(tempomap.MustAt(position(1, 1)), 3, rounding.HalfUp))
	// Editing the grid leaves a bar-addressed mark at the same numbered bar.
	g := grid(2).MustWithMark(1, meter.MustNew(3, 4))
	edited := p.MustEdit(track.Edit{Grid: &g})
	ec, _ := edited.MustClock()
	assert.True(t, c.Map.Equal(ec.Map))
	assertSeconds(t, 4, 1, edited.MustPerformanceMap().MustEnd(tempomap.MustAt(position(3, 4)), 3, rounding.HalfUp))
}

func TestPackRestsKeepsClockAndBarlineData(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	p = p.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: v, Bar: 1, Events: []*event.Event{measureRest(1, false)}}, {Voice: v, Bar: 2, Events: []*event.Event{measureRest(1, true)}}}, Barlines: []track.BarlineEdit{{AfterBar: 1, Fermata: true}}})
	c := withExtension(clockFor(grid(2), 60), grid(2), address(2, 0, 1), 1, 1)
	c.BarHolds = []track.BarHold{{AfterBar: 2, Seconds: tempo.MustNewSeconds(2, 1)}}
	p = withClock(p, c)
	out := p.MustPackRests(v, 1, 2)
	oc, _ := out.Track.MustClock()
	assert.True(t, c.Map.Equal(oc.Map))
	assert.Equal(t, c.BarHolds, oc.BarHolds)
	assert.Equal(t, []int{1}, out.Track.MustFermataBars())
	assert.Empty(t, out.Marks)
	back := out.Track.MustUnpackRest(track.Place{Voice: v, Bar: 1, Index: 0})
	assert.True(t, p.Equal(back.Track))
}

func TestMeterEditResolvesRestButRejectsWrittenOverflow(t *testing.T) {
	t.Parallel()
	g := grid(2).MustWithMark(1, meter.MustNew(3, 4))
	p, v := fixture(2)
	half := tone("C4", duration.Half, event.Marks{})
	p = p.MustSplice(v, 1, 0, 0, half, half)
	next, err := p.Edit(track.Edit{Grid: &g})
	assert.Nil(t, next)
	is := issues(t, err, track.ErrOverflow)
	assert.True(t, v == is[0].Place.Voice)
	assert.Equal(t, 1, is[0].Place.Bar)
	assertLength(t, 1, 4, is[0].Amount)
	assert.True(t, grid(2).Equal(p.MustGrid()))

	p = p.MustSplice(v, 1, 0, 2, measureRest(1, false)).MustSplice(v, 2, 0, 0, tone("D4", duration.Quarter, event.Marks{}))
	next = p.MustEdit(track.Edit{Grid: &g})
	assertSpan(t, next.MustEntries(v, 1)[0].Time, 0, 1, 3, 4)
	assertSpan(t, next.MustEntries(v, 2)[0].Time, 3, 4, 1, 1)
	assertSpan(t, p.MustEntries(v, 2)[0].Time, 1, 1, 5, 4)
}

func TestAssignedInstrumentForTwoVoicesOfThePublishedMelody(t *testing.T) {
	t.Parallel()
	a, b := piano(), piano()
	p := track.MustNew(a, grid(2))
	v := p.MustVoices()[0]
	p, w := p.MustAddVoice()
	upper := []*event.Event{}
	for _, text := range []string{"E4", "F4", "G4", "A4"} {
		upper = append(upper, sounded(note.MustParsePitch(a, text), written(duration.Quarter, 0), event.Marks{}))
	}
	c, g := note.MustParsePitch(a, "C3"), note.MustParsePitch(a, "G3")
	lower := chordEvent(written(duration.Whole, 0), event.Marks{Ties: []int{1}}, c, g)
	p = p.MustEdit(track.Edit{Bars: []track.BarEdit{
		{Voice: v, Bar: 1, Events: upper},
		{Voice: w, Bar: 1, Events: []*event.Event{lower}},
		{Voice: w, Bar: 2, Events: []*event.Event{sounded(g, written(duration.Quarter, 0), event.Marks{})}},
	}})
	assert.Same(t, a, p.MustInstrument())
	assert.False(t, a.Same(b))
	assert.True(t, a.Equal(b))
	assert.Len(t, p.MustEntries(v, 1), 4)
	s := p.MustSounds(w)
	require.Len(t, s, 2)
	assertSpan(t, s[0].Time, 0, 1, 1, 1)
	assertSpan(t, s[1].Time, 0, 1, 5, 4)
	assert.Len(t, s[1].Parts, 2)
}

func TestAccumulatedPositionPreservesFractionRangeError(t *testing.T) {
	t.Parallel()
	// common/fraction's public limit is M, not the host integer width.
	// 1/M + 1/(M-1) = (2M-1)/(M(M-1)); both factors are coprime
	// with 2M-1, so the result cannot be reduced into the allowed range.
	m := fraction.MaxComponentBig()
	f1, err := fraction.FromBigInt(big.NewInt(1), m)
	require.NoError(t, err)
	f2, err := fraction.FromBigInt(big.NewInt(1), new(big.Int).Sub(m, big.NewInt(1)))
	require.NoError(t, err)
	d1, err := event.ValueDuration(f1)
	require.NoError(t, err)
	d2, err := event.ValueDuration(f2)
	require.NoError(t, err)
	first := sounded(parsed("C4"), d1, event.Marks{})
	second := sounded(parsed("D4"), d2, event.Marks{})
	p, v := fixture(1)
	p = p.MustSplice(v, 1, 0, 0, first)
	before := p.MustEntries(v, 1)
	assert.True(t, duration.MustPositionFromFraction(f1).Equal(before[0].Time.End))
	next, err := p.Splice(v, 1, 1, 0, second)
	assert.Nil(t, next)
	require.ErrorIs(t, err, fraction.ErrOutOfRange)
	after := p.MustEntries(v, 1)
	require.Len(t, after, 1)
	assertEvent(t, first, after[0].Event)
	assert.True(t, before[0].Time.Start.Equal(after[0].Time.Start))
	assert.True(t, before[0].Time.End.Equal(after[0].Time.End))
	assert.True(t, p.MustVoiceEnd(v).Equal(before[0].Time.End))
	// A failed proposal leaves the original version usable for another edit.
	replaced := p.MustSplice(v, 1, 0, 1, tone("E4", duration.Quarter, event.Marks{}))
	assertPosition(t, 1, 4, replaced.MustVoiceEnd(v))
	assertEvent(t, first, p.MustEntries(v, 1)[0].Event)
}
