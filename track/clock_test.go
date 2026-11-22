package track_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoClockAndSignsDoNotCreateTime(t *testing.T) {
	t.Parallel()
	p, v := fixture(1)
	_, err := p.PerformanceMap()
	assert.ErrorIs(t, err, track.ErrNoClock)
	_, err = p.SoundSeconds(3, rounding.HalfUp)
	assert.ErrorIs(t, err, track.ErrNoClock)
	p = p.MustSplice(v, 1, 0, 0, tone("C4", duration.Whole, event.Marks{Fermata: true})).MustEdit(track.Edit{Barlines: []track.BarlineEdit{{AfterBar: 1, Fermata: true}}})
	p = withClock(p, clockFor(grid(1), 60))
	timed := p.MustSoundSeconds(3, rounding.HalfUp)
	require.Len(t, timed, 1)
	assertSeconds(t, 0, 1, timed[0].Start)
	assertSeconds(t, 4, 1, timed[0].End)
	_, ok := p.MustClock()
	assert.True(t, ok)
	removed := p.MustEdit(track.Edit{Clock: &track.Clock{}})
	_, ok = removed.MustClock()
	assert.False(t, ok)
	_, err = removed.SoundSeconds(3, rounding.HalfUp)
	assert.ErrorIs(t, err, track.ErrNoClock)
	assert.Equal(t, []int{1}, removed.MustFermataBars())
}

func TestSoundSecondsRejectsExcessPrecisionForEmptyTrack(t *testing.T) {
	t.Parallel()
	p, _ := fixture(1)
	p = withClock(p, clockFor(grid(1), 60))
	result, err := p.SoundSeconds(tempomap.MaxPlaces+1, rounding.HalfUp)
	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestSharedHoldUsesEndsForAttacksAndReleases(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	p, w := p.MustAddVoice()
	p, cross := p.MustAddVoice()
	h := tone("C4", duration.Half, event.Marks{Fermata: true})
	next := tone("D4", duration.Half, event.Marks{})
	p = p.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: v, Bar: 1, Events: []*event.Event{h, next}}, {Voice: w, Bar: 1, Events: []*event.Event{h, next}}, {Voice: cross, Bar: 1, Events: []*event.Event{tone("E3", duration.Whole, event.Marks{})}}}})
	c := withExtension(clockFor(grid(2), 60), grid(2), address(1, 1, 2), 1, 1)
	p = withClock(p, c)
	ts := p.MustSoundSeconds(3, rounding.HalfUp)
	require.Len(t, ts, 5)
	// The first three attacks are simultaneous, then the two following notes.
	assertSeconds(t, 3, 1, ts[0].End)
	assertSeconds(t, 3, 1, ts[1].End)
	assertSeconds(t, 5, 1, ts[2].End)
	assertSeconds(t, 3, 1, ts[3].Start)
	assertSeconds(t, 3, 1, ts[4].Start)
	m := p.MustPerformanceMap()
	assert.Equal(t, tempomap.PositionAddressing, m.MustAddressing())
	a := tempomap.MustAt(position(1, 2))
	assertSeconds(t, 2, 1, m.MustStart(a, 3, rounding.HalfUp))
	assertSeconds(t, 3, 1, m.MustEnd(a, 3, rounding.HalfUp))
	for _, sec := range []tempo.Seconds{tempo.MustNewSeconds(5, 2), tempo.MustNewSeconds(3, 1)} {
		r := m.MustPosition(sec, 3, rounding.HalfUp)
		assertPosition(t, 1, 2, r.MustValue())
		assert.False(t, r.MustApproximate())
	}
	// Without signs, the same explicit hold has the same timing.
	plain := p.MustSplice(v, 1, 0, 1, tone("C4", duration.Half, event.Marks{})).MustSplice(w, 1, 0, 1, tone("C4", duration.Half, event.Marks{}))
	assertSeconds(t, 3, 1, plain.MustSoundSeconds(3, rounding.HalfUp)[0].End)
}

func TestFinalAndInitialHoldsAndBarIdentity(t *testing.T) {
	t.Parallel()
	p, v := fixture(1)
	p = p.MustSplice(v, 1, 0, 0, tone("C4", duration.Whole, event.Marks{}))
	c := clockFor(grid(1), 60)
	c.BarHolds = []track.BarHold{{AfterBar: 1, Seconds: tempo.MustNewSeconds(1, 1)}}
	p = withClock(p, c)
	m := p.MustPerformanceMap()
	assertSeconds(t, 4, 1, m.MustStart(tempomap.MustAt(position(1, 1)), 3, rounding.HalfUp))
	assertSeconds(t, 5, 1, m.MustEnd(tempomap.MustAt(position(1, 1)), 3, rounding.HalfUp))
	assertSeconds(t, 5, 1, p.MustSoundSeconds(3, rounding.HalfUp)[0].End)
	for _, sec := range []tempo.Seconds{tempo.MustNewSeconds(9, 2), tempo.MustNewSeconds(5, 1)} {
		assertPosition(t, 1, 1, m.MustPosition(sec, 3, rounding.HalfUp).MustValue())
	}
	end, finite := p.MustGrid().MustEnd()
	assert.True(t, finite)
	assertPosition(t, 1, 1, end)
	g := grid(2)
	long := p.MustEdit(track.Edit{Grid: &g})
	lc, _ := long.MustClock()
	assert.Equal(t, 1, lc.BarHolds[0].AfterBar)
	assertSeconds(t, 5, 1, long.MustPerformanceMap().MustEnd(tempomap.MustAt(position(1, 1)), 3, rounding.HalfUp))
	assertSeconds(t, 9, 1, long.MustPerformanceMap().MustEnd(tempomap.MustAt(position(2, 1)), 3, rounding.HalfUp))
	_, found := p.MustAt(v, position(1, 1))
	assert.False(t, found)
	assertSeconds(t, 9, 1, m.MustEnd(tempomap.MustAt(position(2, 1)), 3, rounding.HalfUp))
	initial := withExtension(clockFor(grid(1), 60), grid(1), address(1, 0, 1), 2, 1)
	assertSeconds(t, 2, 1, withClock(p, initial).MustSoundSeconds(3, rounding.HalfUp)[0].Start)
}

func TestClockValidationAndAtomicGridChanges(t *testing.T) {
	t.Parallel()
	p, _ := fixture(2)
	for _, c := range []track.Clock{
		{Map: tempomap.MustNew(tempo.MustNewQuarter(fraction.FromInt(60)))},
		{BarHolds: []track.BarHold{{AfterBar: 1, Seconds: tempo.MustNewSeconds(1, 1)}}},
		{Map: clockFor(grid(2), 60).Map, BarHolds: []track.BarHold{{AfterBar: 1}}},
		{Map: clockFor(grid(2), 60).Map, BarHolds: []track.BarHold{{AfterBar: 3, Seconds: tempo.ZeroSeconds()}}},
		{Map: clockFor(grid(2), 60).Map, BarHolds: []track.BarHold{{AfterBar: 1, Seconds: tempo.ZeroSeconds()}, {AfterBar: 1, Seconds: tempo.ZeroSeconds()}}},
	} {
		next, err := p.Edit(track.Edit{Clock: &c})
		assert.Nil(t, next)
		assert.Error(t, err)
		_, set := p.MustClock()
		assert.False(t, set)
	}
	c := withExtension(clockFor(grid(2), 60), grid(2), address(2, 0, 1), 1, 1)
	c.BarHolds = []track.BarHold{{AfterBar: 1, Seconds: tempo.MustNewSeconds(1, 1)}}
	next, err := p.Edit(track.Edit{Clock: &c})
	assert.Nil(t, next)
	issues(t, err, track.ErrHoldConflict)
	c = withExtension(clockFor(grid(2), 60), grid(2), address(1, 7, 8), 1, 1)
	p = withClock(p, c)
	g := grid(2).MustWithMark(1, meter.MustNew(3, 4))
	next, err = p.Edit(track.Edit{Grid: &g})
	assert.Nil(t, next)
	is := issues(t, err, track.ErrClockAddress)
	assert.True(t, is[0].Address.Equal(address(1, 7, 8)))
	assert.Equal(t, track.Extension, is[0].Kind)
	assert.True(t, grid(2).Equal(p.MustGrid()))
	assert.Len(t, p.MustPerformanceMap().MustExtensions(), 1)
	// Removing the obsolete clock in the same edit makes the grid change valid.
	changed := p.MustEdit(track.Edit{Grid: &g, Clock: &track.Clock{}})
	assert.True(t, g.Equal(changed.MustGrid()))
	base, _ := fixture(2)
	base = withClock(base, clockFor(grid(2), 60))
	pickup := grid(2).MustWithPickup(length(1, 4))
	_, err = base.Edit(track.Edit{Grid: &pickup})
	assert.ErrorIs(t, err, track.ErrClockAddress)
	// Holds attached to a removed bar require explicit replacement.
	c = clockFor(grid(2), 60)
	c.BarHolds = []track.BarHold{{AfterBar: 2, Seconds: tempo.MustNewSeconds(1, 1)}}
	base = withClock(base, c)
	short := grid(1)
	_, err = base.Edit(track.Edit{Grid: &short})
	is = issues(t, err, track.ErrClockAddress)
	require.NotNil(t, is[0].AfterBar)
	assert.Equal(t, 2, *is[0].AfterBar)
	base = base.MustEdit(track.Edit{Grid: &short, Clock: &track.Clock{}})
	assert.True(t, short.Equal(base.MustGrid()))
}

func TestUnfinishedAndGradualClockResults(t *testing.T) {
	t.Parallel()
	g := grid(2)
	p, v := fixture(2)
	p = p.MustSplice(v, 1, 0, 0, tone("C4", duration.Whole, event.Marks{}))
	c := clockFor(g, 60)
	gradual := tempomap.MustNewGradual(tempomap.MustAtBar(address(1, 0, 1)), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)
	c.Map = c.Map.MustWithChange(gradual, g)
	unfinished := withClock(p, c)
	_, err := unfinished.SoundSeconds(3, rounding.HalfUp)
	assert.ErrorIs(t, err, tempomap.ErrUnfinished)
	c.Map = c.Map.MustWithChange(tempomap.MustNewMain(tempomap.MustAtBar(address(2, 0, 1)), tempo.MustNewQuarter(fraction.FromInt(120))), g)
	complete := withClock(p, c)
	ts := complete.MustSoundSeconds(3, rounding.HalfUp)
	require.Len(t, ts, 1)
	assert.True(t, ts[0].End.MustApproximate())
	assert.True(t, tempo.MustNewSeconds(2773, 1000).Equal(ts[0].End.MustValue()))
	assert.True(t, tempo.MustNewSeconds(1, 2000).Equal(ts[0].End.MustErrorBound()))
	places, set := ts[0].End.MustPlaces()
	assert.True(t, set)
	assert.Equal(t, 3, places)
}

func TestQuantizeHoldConflictAndIndependentClockMarks(t *testing.T) {
	t.Parallel()
	for _, fermata := range []bool{true, false} {
		p, v := fixture(1)
		p = p.MustSplice(v, 1, 0, 0, sounded(parsed("C4"), written(duration.Eighth, 1), event.Marks{Fermata: fermata}))
		c := withExtension(clockFor(grid(1), 60), grid(1), address(1, 3, 16), 1, 1)
		p = withClock(p, c)
		out, err := p.Quantize(v, 1, 1, length(1, 8), rounding.HalfUp, track.KeepGroups)
		if fermata {
			is := noProposal(t, out, err, track.ErrHoldConflict)
			assertPosition(t, 3, 16, is[0].Position)
			assert.True(t, is[0].Address.Equal(address(1, 3, 16)))
			continue
		}
		require.NoError(t, err)
		assertPosition(t, 1, 4, out.Track.MustVoiceEnd(v))
		oc, _ := out.Track.MustClock()
		assert.True(t, c.Map.Equal(oc.Map))
	}
}

func TestRebarMovesClockAddressesAndBarlineMarks(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	p = p.MustSplice(v, 1, 0, 0, tone("C4", duration.Whole, event.Marks{})).MustEdit(track.Edit{Barlines: []track.BarlineEdit{{AfterBar: 1, Fermata: true}}})
	c := clockFor(grid(2), 60)
	c.Map = c.Map.MustWithChange(tempomap.MustNewJump(tempomap.MustAtBar(address(2, 0, 1)), tempo.MustNewQuarter(fraction.FromInt(120))), grid(2))
	c.BarHolds = []track.BarHold{{AfterBar: 1, Seconds: tempo.MustNewSeconds(1, 1)}}
	p = withClock(p, c)
	g := measure.MustNew(meter.MustNew(2, 4)).MustWithBarCount(4)
	out := p.MustRebar(g, track.KeepGroups)
	assertReport(t, p, out)
	assert.Equal(t, []int{2}, out.Track.MustFermataBars())
	assertSeconds(t, 5, 1, out.Track.MustSoundSeconds(3, rounding.HalfUp)[0].End)
	kinds := map[track.MarkKind]bool{}
	for _, m := range out.Marks {
		kinds[m.Kind] = true
		assertPosition(t, 1, 1, m.From)
		assertPosition(t, 1, 1, m.To)
		if m.Kind == track.BarlineFermata {
			require.NotNil(t, m.OldAfterBar)
			require.NotNil(t, m.NewAfterBar)
			assert.Equal(t, 1, *m.OldAfterBar)
			assert.Equal(t, 2, *m.NewAfterBar)
		}
		if m.Kind == track.TempoChange {
			assert.True(t, m.OldAddress.Equal(address(2, 0, 1)))
			assert.True(t, m.NewAddress.Equal(address(3, 0, 1)))
			assert.Equal(t, tempomap.SlotSingle, m.Slot)
		}
	}
	assert.True(t, kinds[track.TempoChange])
	assert.True(t, kinds[track.Extension])
	assert.True(t, kinds[track.BarlineFermata])
	// A barline sign cannot become an intra-bar mark.
	incompatible := measure.MustNew(meter.MustNew(3, 4)).MustWithBarCount(3)
	bad, err := p.Rebar(incompatible, track.KeepGroups)
	noProposal(t, bad, err, track.ErrWouldLoseMark)
	assert.Equal(t, []int{1}, p.MustFermataBars())
	// A report does not own the snapshot's bar-number pointers.
	for i := range out.Marks {
		if out.Marks[i].NewAfterBar != nil {
			*out.Marks[i].NewAfterBar = 99
		}
	}
	again := p.MustRebar(g, track.KeepGroups)
	assert.Equal(t, []int{2}, again.Track.MustFermataBars())
}
