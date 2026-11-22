package track_test

import (
	"errors"
	"math"
	"sync"
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/track"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAndVoiceIdentity(t *testing.T) {
	t.Parallel()
	a, b := piano(), piano()
	p, err := track.New(a, grid(2))
	require.NoError(t, err)
	require.True(t, p.IsSet())
	assert.Same(t, a, p.MustInstrument())
	assert.True(t, grid(2).Equal(p.MustGrid()))
	assert.Len(t, p.MustVoices(), 1)
	v := p.MustVoices()[0]
	assert.NotEqual(t, track.VoiceID{}, v)
	assertPosition(t, 0, 1, p.MustVoiceEnd(v))
	assert.Empty(t, p.MustEntries(v, 1))
	assert.Empty(t, p.MustFermataBars())
	assert.Empty(t, p.MustSounds())
	_, set, err := p.Clock()
	require.NoError(t, err)
	assert.False(t, set)

	other := track.MustNew(a, grid(2))
	assert.True(t, p.Equal(other), "voice keys do not enter record equality")
	assert.False(t, p.Equal(track.MustNew(b, grid(2))), "instrument identity enters equality")
	derived, id, err := p.AddVoice()
	require.NoError(t, err)
	assert.Equal(t, []track.VoiceID{v, id}, derived.MustVoices())
	assert.Len(t, p.MustVoices(), 1)
	sibling, siblingID := p.MustAddVoice()
	assert.NotEqual(t, id, siblingID)
	assert.True(t, derived.Equal(sibling))
	assert.NotEqual(t, v, id)
	assert.Equal(t, 2, map[track.VoiceID]int{v: 1, id: 2}[id])
	for name, bad := range map[string]track.VoiceID{"zero": {}, "foreign": other.MustVoices()[0], "sibling": siblingID} {
		t.Run(name, func(t *testing.T) {
			_, err := derived.Entries(bad, 1)
			assert.ErrorIs(t, err, track.ErrVoice)
		})
	}
	_, err = p.Entries(id, 1)
	assert.ErrorIs(t, err, track.ErrVoice)
	copyOf := *derived
	copyNext := copyOf.MustSplice(id, 1, 0, 0, tone("C4", duration.Quarter, event.Marks{}))
	assert.Empty(t, derived.MustEntries(id, 1))
	assert.Len(t, copyNext.MustEntries(id, 1), 1)
}

func TestUnsetAndConstructorArguments(t *testing.T) {
	t.Parallel()
	for _, p := range []*track.Track{nil, {}} {
		assert.False(t, p.IsSet())
		assert.True(t, p.Equal(nil))
		assert.True(t, p.Equal(&track.Track{}))
		_, err := p.Instrument()
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.Grid()
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.Voices()
		assert.ErrorIs(t, err, track.ErrUnset)
		next, id, err := p.AddVoice()
		assert.Nil(t, next)
		assert.Zero(t, id)
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.Edit(track.Edit{})
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.Splice(track.VoiceID{}, 1, 0, 0)
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.Entries(track.VoiceID{}, 1)
		assert.ErrorIs(t, err, track.ErrUnset)
		_, _, err = p.At(track.VoiceID{}, duration.Origin())
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.VoiceEnd(track.VoiceID{})
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.FermataBars()
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.Sounds()
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.SoundingAt(duration.Origin())
		assert.ErrorIs(t, err, track.ErrUnset)
		_, _, err = p.Clock()
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.PerformanceMap()
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.SoundSeconds(3, rounding.HalfUp)
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.WriteEvents(track.VoiceID{}, address(1, 0, 1), nil, track.WriteOptions{})
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.Rebar(grid(1), track.KeepGroups)
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.Quantize(track.VoiceID{}, 1, 1, length(1, 8), rounding.HalfUp, track.KeepGroups)
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.Split(track.Place{}, length(1, 8), track.KeepGroups)
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.Merge(track.VoiceID{}, 1, 0, 2, track.KeepGroups)
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.PackRests(track.VoiceID{}, 1, 1)
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.UnpackRest(track.Place{})
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.ExtractLine(track.VoiceID{}, track.Exact)
		assert.ErrorIs(t, err, track.ErrUnset)
		_, err = p.WriteLine(track.VoiceID{}, address(1, 0, 1), nil, track.WriteOptions{})
		assert.ErrorIs(t, err, track.ErrUnset)
	}
	for _, tc := range []struct {
		name string
		inst *instrument.Instrument
		g    measure.Grid
	}{
		{"nil instrument", nil, grid(1)}, {"unset instrument", &instrument.Instrument{}, grid(1)},
		{"unset grid", piano(), measure.Grid{}}, {"open grid", piano(), measure.MustNew(meter.MustNew(4, 4))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := track.New(tc.inst, tc.g)
			assert.Nil(t, p)
			require.ErrorIs(t, err, track.ErrInvalidArgument)
			_, ok := errors.AsType[*track.Problems](err)
			assert.False(t, ok)
			assertMustError(t, err, track.ErrInvalidArgument, func() { track.MustNew(tc.inst, tc.g) })
		})
	}
	withoutRange := instrument.MustNew(instrument.Piano().MustWithoutRange())
	_, err := track.New(withoutRange, grid(1))
	assert.NoError(t, err)
}

func TestRecordSpliceAndHalfOpenLookup(t *testing.T) {
	t.Parallel()
	p, v := fixture(3)
	r := rest(written(duration.Half, 0), false)
	c := tone("C4", duration.Quarter, event.Marks{})
	d := tone("D4", duration.Quarter, event.Marks{})
	p = p.MustSplice(v, 1, 0, 0, r, c)
	entries := p.MustEntries(v, 1)
	require.Len(t, entries, 2)
	assertEntry(t, entries[0], v, 1, 0, r, 0, 1, 1, 2)
	assertEntry(t, entries[1], v, 1, 1, c, 1, 2, 3, 4)
	assert.Same(t, c, entries[1].Event)
	p2 := p.MustSplice(v, 1, 1, 0, d)
	p3 := p2.MustSplice(v, 1, 0, 1)
	assert.Len(t, p.MustEntries(v, 1), 2)
	assert.Len(t, p2.MustEntries(v, 1), 3)
	assertEntry(t, p3.MustEntries(v, 1)[0], v, 1, 0, d, 0, 1, 1, 4)
	for _, tc := range []struct {
		name  string
		p     duration.Position
		found bool
		index int
	}{
		{"rest", position(1, 4), true, 0}, {"next begins", position(1, 2), true, 1},
		{"empty tail", position(3, 4), false, 0}, {"empty next bar", position(1, 1), false, 0}, {"end", position(3, 1), false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, ok, err := p.At(v, tc.p)
			require.NoError(t, err)
			assert.Equal(t, tc.found, ok)
			if ok {
				assert.Equal(t, tc.index, e.Place.Index)
			} else {
				assert.Zero(t, e)
			}
		})
	}
	assertPosition(t, 3, 4, p.MustVoiceEnd(v))
	for _, bad := range []duration.Position{{}, position(13, 4)} {
		_, _, err := p.At(v, bad)
		require.Error(t, err)
	}
	_, _, err := p.At(v, position(13, 4))
	is := issues(t, err, track.ErrPlace)
	assert.True(t, is[0].Position.Equal(position(13, 4)))
}

func TestPickupSparseAndContextualRests(t *testing.T) {
	t.Parallel()
	g := measure.MustNew(meter.MustNew(3, 4)).MustWithPickup(length(1, 4)).MustWithBarCount(2)
	p := track.MustNew(piano(), g)
	v := p.MustVoices()[0]
	d, e := tone("D4", duration.Quarter, event.Marks{}), tone("E4", duration.Quarter, event.Marks{})
	p = p.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: v, Bar: 0, Events: []*event.Event{d}}, {Voice: v, Bar: 1, Events: []*event.Event{e}}}})
	assertEntry(t, p.MustEntries(v, 1)[0], v, 1, 0, e, 1, 4, 1, 2)
	at, ok := p.MustAt(v, position(1, 4))
	require.True(t, ok)
	assert.Equal(t, 1, at.Place.Bar)
	p = withClock(p, clockFor(g, 60))
	timed := p.MustSoundSeconds(3, rounding.HalfUp)
	require.Len(t, timed, 2)
	assertSeconds(t, 1, 1, timed[1].Start)

	sparse, sv := fixture(25)
	c := tone("C4", duration.Quarter, event.Marks{})
	sparse = sparse.MustSplice(sv, 25, 0, 0, c)
	assert.Empty(t, sparse.MustEntries(sv, 24))
	assertPosition(t, 97, 4, sparse.MustVoiceEnd(sv))
	assertSpan(t, sparse.MustSounds()[0].Time, 24, 1, 97, 4)
	sparse = withClock(sparse, clockFor(grid(25), 120))
	ts := sparse.MustSoundSeconds(3, rounding.HalfUp)
	require.Len(t, ts, 1)
	assertSeconds(t, 48, 1, ts[0].Start)
	assertSeconds(t, 97, 2, ts[0].End)

	rp, rv := fixture(4)
	r := measureRest(4, true)
	rp = rp.MustSplice(rv, 1, 0, 0, r)
	for bar := 2; bar <= 4; bar++ {
		assert.Empty(t, rp.MustEntries(rv, bar))
	}
	got, ok := rp.MustAt(rv, position(7, 2))
	require.True(t, ok)
	assertEntry(t, got, rv, 1, 0, r, 0, 1, 4, 1)
	assertPosition(t, 4, 1, rp.MustVoiceEnd(rv))
	short := grid(2).MustWithLength(1, length(1, 2), measure.AlignEnd)
	sp := track.MustNew(piano(), short)
	sv = sp.MustVoices()[0]
	sp = sp.MustSplice(sv, 1, 0, 0, measureRest(1, false))
	assertSpan(t, sp.MustEntries(sv, 1)[0].Time, 0, 1, 1, 2)
}

func TestAtomicEditsOverflowAndTruncation(t *testing.T) {
	t.Parallel()
	p, v := fixture(4)
	p, w := p.MustAddVoice()
	q := tone("C4", duration.Quarter, event.Marks{})
	whole := tone("D3", duration.Whole, event.Marks{})
	over := []*event.Event{q, q, q, tone("C4", duration.Half, event.Marks{})}
	next, err := p.Edit(track.Edit{Bars: []track.BarEdit{{Voice: w, Bar: 2, Events: over}, {Voice: v, Bar: 1, Events: over}}})
	assert.Nil(t, next)
	all := issues(t, err, track.ErrOverflow)
	require.Len(t, all, 2)
	assert.Equal(t, v, all[0].Place.Voice)
	assert.Equal(t, 1, all[0].Place.Bar)
	assert.Equal(t, w, all[1].Place.Voice)
	assert.Equal(t, 2, all[1].Place.Bar)
	for _, i := range all {
		assertLength(t, 1, 4, i.Amount)
	}
	assert.Empty(t, p.MustEntries(v, 1))
	assert.Empty(t, p.MustEntries(w, 2))
	p = p.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: v, Bar: 1, Events: []*event.Event{whole}}, {Voice: w, Bar: 1, Events: []*event.Event{whole}}}})
	assert.Len(t, p.MustSounds(), 2)
	long := p.MustSplice(v, 4, 0, 0, q).MustEdit(track.Edit{Barlines: []track.BarlineEdit{{AfterBar: 4, Fermata: true}}})
	short := grid(2)
	next, err = long.Edit(track.Edit{Grid: &short})
	assert.Nil(t, next)
	issues(t, err, track.ErrLostRecord)
	cleared := long.MustEdit(track.Edit{Grid: &short, Bars: []track.BarEdit{{Voice: v, Bar: 4}}, Barlines: []track.BarlineEdit{{AfterBar: 4, Fermata: false}}})
	assert.True(t, short.Equal(cleared.MustGrid()))
	assert.Empty(t, cleared.MustFermataBars())
	assert.Len(t, long.MustEntries(v, 4), 1)
	assert.Equal(t, []int{4}, long.MustFermataBars())

	for name, edit := range map[string]track.Edit{
		"duplicate cell":    {Bars: []track.BarEdit{{Voice: v, Bar: 1}, {Voice: v, Bar: 1}}},
		"duplicate barline": {Barlines: []track.BarlineEdit{{AfterBar: 1}, {AfterBar: 1}}},
		"unset grid":        {Grid: &measure.Grid{}}, "unset event": {Bars: []track.BarEdit{{Voice: v, Bar: 2, Events: []*event.Event{nil}}}},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := p.Edit(edit)
			assert.Nil(t, got)
			assert.Error(t, err)
			assert.Len(t, p.MustEntries(v, 1), 1)
		})
	}
}

func TestIndicesAndRangeDiagnostics(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	p = p.MustSplice(v, 1, 0, 0, tone("C4", duration.Quarter, event.Marks{}))
	for _, tc := range []struct {
		name          string
		index, remove int
		cause         error
	}{
		{"negative index", -1, -1, track.ErrPlace}, {"past index", 2, 0, track.ErrPlace},
		{"negative remove", 0, -1, track.ErrInvalidArgument}, {"past tail", 1, 1, track.ErrInvalidArgument},
		{"integer remove", 1, math.MaxInt, track.ErrInvalidArgument}, {"integer index", math.MaxInt, 1, track.ErrPlace},
	} {
		t.Run(tc.name, func(t *testing.T) {
			next, err := p.Splice(v, 1, tc.index, tc.remove)
			assert.Nil(t, next)
			is := issues(t, err, tc.cause)
			assert.Equal(t, track.Place{Voice: v, Bar: 1, Index: tc.index}, is[0].Place)
			assert.Equal(t, -1, is[0].NoteIndex)
			assert.Equal(t, -1, is[0].InputIndex)
		})
	}
	for _, tc := range []struct {
		first, last int
		bad         int
		cause       error
	}{{0, 9, 0, track.ErrPlace}, {1, 9, 9, track.ErrPlace}, {2, 1, 2, track.ErrInvalidArgument}} {
		_, err := p.Quantize(v, tc.first, tc.last, length(1, 8), rounding.HalfUp, track.KeepGroups)
		is := issues(t, err, tc.cause)
		assert.Equal(t, tc.bad, is[0].Place.Bar)
		assert.Equal(t, -1, is[0].Place.Index)
		_, err = p.PackRests(v, tc.first, tc.last)
		is = issues(t, err, tc.cause)
		assert.Equal(t, tc.bad, is[0].Place.Bar)
	}
	for _, bar := range []int{0, -1, 3, math.MaxInt} {
		_, err := p.Entries(v, bar)
		issues(t, err, track.ErrPlace)
	}
	for _, index := range []int{-1, 1, math.MaxInt} {
		at := track.Place{Voice: v, Bar: 1, Index: index}
		_, err := p.Split(at, length(1, 8), track.KeepGroups)
		issues(t, err, track.ErrPlace)
		_, err = p.UnpackRest(at)
		issues(t, err, track.ErrPlace)
		_, err = p.Merge(v, 1, index, 2, track.KeepGroups)
		issues(t, err, track.ErrPlace)
	}
	for _, count := range []int{-1, 0, 1, 2, math.MaxInt} {
		_, err := p.Merge(v, 1, 0, count, track.KeepGroups)
		issues(t, err, track.ErrInvalidArgument)
	}
	for _, off := range []duration.Duration{{}, length(1, 4), length(1, 2)} {
		_, err := p.Split(track.Place{Voice: v, Bar: 1, Index: 0}, off, track.KeepGroups)
		issues(t, err, track.ErrInvalidArgument)
	}
}

func TestInstrumentAndNoteSourcesRemainIndependent(t *testing.T) {
	t.Parallel()
	a, b := piano(), piano()
	p := track.MustNew(a, grid(2))
	v := p.MustVoices()[0]
	na, nb := note.MustParse(a, "A4"), note.MustParse(b, "A4")
	free := na
	free.MustUnbind()
	input := []*event.Event{sounded(na, written(duration.Quarter, 0), event.Marks{}), sounded(nb, written(duration.Quarter, 0), event.Marks{}), sounded(free, written(duration.Quarter, 0), event.Marks{})}
	p = p.MustSplice(v, 1, 0, 0, input...)
	input[0] = nil
	a.MustRetuneTuning(tuning.Standard().MustWithReferenceFrequency(415))
	assert.Same(t, a, p.MustInstrument())
	entries := p.MustEntries(v, 1)
	require.Len(t, entries, 3)
	gotA := entries[0].Event.MustContent().MustNote()
	gotB := entries[1].Event.MustContent().MustNote()
	gotFree := entries[2].Event.MustContent().MustNote()
	assertNote(t, na, gotA)
	assertNote(t, nb, gotB)
	assertNote(t, free, gotFree)
	assert.InDelta(t, 415, gotA.MustFrequency(), 1e-9)
	assert.InDelta(t, 440, gotB.MustFrequency(), 1e-9)
	_, err := gotFree.Frequency()
	assert.ErrorIs(t, err, note.ErrNoSource)
	assert.InDelta(t, 415, gotFree.MustFrequencyIn(a), 1e-9)
	p = p.MustSplice(v, 2, 0, 0, sounded(note.MustParse(a, "D8"), written(duration.Quarter, 0), event.Marks{}))
	assert.Len(t, p.MustEntries(v, 2), 1)
	gotA.MustSetASPN(5)
	assertNote(t, na, p.MustEntries(v, 1)[0].Event.MustContent().MustNote())
}

func TestConcurrentSnapshotReadingAndIndependentEdits(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	e := tone("C4", duration.Quarter, event.Marks{})
	p = p.MustSplice(v, 1, 0, 0, e)
	type result struct {
		p   *track.Track
		id  track.VoiceID
		err error
	}
	ch := make(chan result, 16)
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			if _, err := p.Sounds(); err != nil {
				ch <- result{err: err}
				return
			}
			if _, _, err := p.At(v, position(1, 8)); err != nil {
				ch <- result{err: err}
				return
			}
			n, id, err := p.AddVoice()
			if err == nil {
				n, err = n.Splice(id, 2, 0, 0, e)
			}
			ch <- result{n, id, err}
		})
	}
	wg.Wait()
	close(ch)
	ids := map[track.VoiceID]bool{}
	for r := range ch {
		require.NoError(t, r.err)
		assert.False(t, ids[r.id])
		ids[r.id] = true
		assert.Len(t, r.p.MustVoices(), 2)
		assert.Len(t, r.p.MustEntries(r.id, 2), 1)
	}
	assert.Len(t, p.MustVoices(), 1)
	assert.Empty(t, p.MustEntries(v, 2))
}

func TestBarHoldInputOwnership(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	c := clockFor(grid(2), 60)
	c.BarHolds = []track.BarHold{{AfterBar: 1, Seconds: tempo.MustNewSeconds(1, 1)}}
	p = withClock(p, c)
	c.BarHolds[0].AfterBar = 99
	got, ok := p.MustClock()
	require.True(t, ok)
	assert.Equal(t, 1, got.BarHolds[0].AfterBar)
	got.BarHolds[0].AfterBar = 99
	again, _ := p.MustClock()
	assert.Equal(t, 1, again.BarHolds[0].AfterBar)
	voices := p.MustVoices()
	voices[0] = track.VoiceID{}
	assert.Equal(t, v, p.MustVoices()[0])
}
