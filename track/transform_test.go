package track_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/track"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A report's addresses must identify the actual snapshots, independently of
// how the implementation groups its changes into fragments.
func assertReport(t *testing.T, old *track.Track, p track.Proposal) {
	t.Helper()
	require.NotNil(t, p.Track)
	for _, change := range p.Changes {
		for side, entries := range [][]track.Entry{change.From, change.To} {
			snapshot := old
			if side == 1 {
				snapshot = p.Track
			}
			for _, entry := range entries {
				actual := snapshot.MustEntries(entry.Place.Voice, entry.Place.Bar)
				require.GreaterOrEqual(t, entry.Place.Index, 0)
				require.Less(t, entry.Place.Index, len(actual))
				got := actual[entry.Place.Index]
				assertEvent(t, entry.Event, got.Event)
				assert.True(t, entry.Time.Start.Equal(got.Time.Start))
				assert.True(t, entry.Time.End.Equal(got.Time.End))
			}
		}
	}
}

// Compare the complete unions of both sides without specifying how Change
// groups a split or a merge. Each individual side must remain in place order.
func assertChanges(t *testing.T, old *track.Track, p track.Proposal, from, to []track.Entry) {
	t.Helper()
	assertReport(t, old, p)
	ranks := map[track.VoiceID]int{}
	for i, v := range old.MustVoices() {
		ranks[v] = i
	}
	less := func(a, b track.Place) bool {
		if a.Voice != b.Voice {
			return ranks[a.Voice] < ranks[b.Voice]
		}
		if a.Bar != b.Bar {
			return a.Bar < b.Bar
		}
		return a.Index < b.Index
	}
	var previous track.Place
	for i, c := range p.Changes {
		entries := c.From
		if len(entries) == 0 {
			entries = c.To
		}
		require.NotEmpty(t, entries, "empty change has no provenance")
		anchor := entries[0].Place
		if i > 0 {
			assert.False(t, less(anchor, previous), "changes out of place order")
		}
		previous = anchor
	}
	for side, want := range [][]track.Entry{from, to} {
		union := map[track.Place]track.Entry{}
		for _, change := range p.Changes {
			entries := change.From
			if side == 1 {
				entries = change.To
			}
			for i, e := range entries {
				if i > 0 {
					assert.True(t, less(entries[i-1].Place, e.Place), "report side out of order")
				}
				if previous, ok := union[e.Place]; ok {
					assertEvent(t, previous.Event, e.Event)
					assert.True(t, previous.Time.Start.Equal(e.Time.Start))
					assert.True(t, previous.Time.End.Equal(e.Time.End))
				}
				union[e.Place] = e
			}
		}
		require.Len(t, union, len(want), "incomplete or extraneous report entries on side %d", side)
		for _, e := range want {
			got, ok := union[e.Place]
			require.True(t, ok, "missing report place %+v", e.Place)
			assertEvent(t, e.Event, got.Event)
			assert.True(t, e.Time.Start.Equal(got.Time.Start))
			assert.True(t, e.Time.End.Equal(got.Time.End))
		}
	}
}

func assertNoInputs(t *testing.T, p track.Proposal) {
	t.Helper()
	for _, c := range p.Changes {
		assert.Empty(t, c.Input)
	}
}

func assertInputs(t *testing.T, p track.Proposal, sources map[track.Place]int) {
	t.Helper()
	for _, c := range p.Changes {
		want := []int{}
		for _, e := range c.To {
			if input, ok := sources[e.Place]; ok && !slices.Contains(want, input) {
				want = append(want, input)
			}
		}
		slices.Sort(want)
		if len(want) == 0 {
			assert.Empty(t, c.Input)
		} else {
			assert.Equal(t, want, c.Input)
		}
	}
}

func TestWriteEventsSplitsAcrossBarsAndReportsInputs(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	q := tone("C4", duration.Quarter, event.Marks{})
	h := tone("D4", duration.Half, event.Marks{Fermata: true})
	input := []*event.Event{q, q, q, h}
	bad, err := p.Splice(v, 1, 0, 0, input...)
	assert.Nil(t, bad)
	issues(t, err, track.ErrOverflow)
	proposal, err := p.WriteEvents(v, address(1, 0, 1), input, track.WriteOptions{})
	require.NoError(t, err)
	assertReport(t, p, proposal)
	quarterFirst := tone("D4", duration.Quarter, event.Marks{Ties: []int{0}})
	quarterLast := tone("D4", duration.Quarter, event.Marks{Fermata: true})
	wantTo := []track.Entry{
		expectedEntry(v, 1, 0, q, 0, 1, 1, 4), expectedEntry(v, 1, 1, q, 1, 4, 1, 2), expectedEntry(v, 1, 2, q, 1, 2, 3, 4),
		expectedEntry(v, 1, 3, quarterFirst, 3, 4, 1, 1), expectedEntry(v, 2, 0, quarterLast, 1, 1, 5, 4),
	}
	assertChanges(t, p, proposal, nil, wantTo)
	assertInputs(t, proposal, map[track.Place]int{wantTo[0].Place: 0, wantTo[1].Place: 1, wantTo[2].Place: 2, wantTo[3].Place: 3, wantTo[4].Place: 3})
	require.NotEmpty(t, proposal.Changes)
	next := proposal.Track
	e1 := next.MustEntries(v, 1)
	e2 := next.MustEntries(v, 2)
	require.Len(t, e1, 4)
	require.Len(t, e2, 1)
	for i := range 3 {
		assert.Same(t, q, e1[i].Event)
	}
	assertSpan(t, e1[3].Time, 3, 4, 1, 1)
	assertSpan(t, e2[0].Time, 1, 1, 5, 4)
	assert.Equal(t, []int{0}, e1[3].Event.MustMarks().Ties)
	assert.False(t, e1[3].Event.MustMarks().Fermata)
	assert.Empty(t, e2[0].Event.MustMarks().Ties)
	assert.True(t, e2[0].Event.MustMarks().Fermata)
	assert.True(t, e1[3].Event.MustDuration().Equal(written(duration.Quarter, 0)))
	assert.True(t, e2[0].Event.MustDuration().Equal(written(duration.Quarter, 0)))
	s := next.MustSounds(v)
	require.Len(t, s, 4)
	assertSpan(t, s[3].Time, 3, 4, 5, 4)
	assert.Len(t, s[3].Parts, 2)
	seen := map[int]int{}
	for _, c := range proposal.Changes {
		assert.Empty(t, c.From)
		for _, i := range c.Input {
			seen[i] += len(c.To)
		}
	}
	for i := range input {
		assert.Positive(t, seen[i], "input %d missing from provenance", i)
	}
	assert.Empty(t, p.MustEntries(v, 1))
	input[3] = nil
	assert.Len(t, next.MustEntries(v, 2), 1)
}

func TestWriteVacancyReplacementAndExactBoundary(t *testing.T) {
	t.Parallel()
	p, v := fixture(3)
	p, w := p.MustAddVoice()
	q := tone("C4", duration.Quarter, event.Marks{})
	d := tone("D4", duration.Quarter, event.Marks{})
	p = p.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: v, Bar: 2, Events: []*event.Event{q}}, {Voice: v, Bar: 3, Events: []*event.Event{d}}, {Voice: w, Bar: 1, Events: []*event.Event{d}}}})
	proposal, err := p.WriteEvents(v, address(2, 1, 2), []*event.Event{d}, track.WriteOptions{})
	require.NoError(t, err)
	assertReport(t, p, proposal)
	e := proposal.Track.MustEntries(v, 2)
	require.Len(t, e, 3)
	assert.Same(t, q, e[0].Event)
	assert.Equal(t, event.KindRest, e[1].Event.Kind())
	assertSpan(t, e[1].Time, 5, 4, 3, 2)
	assert.Equal(t, event.DurationValue, e[1].Event.MustDuration().Kind())
	assert.Same(t, d, e[2].Event)
	assert.Empty(t, proposal.Track.MustEntries(v, 1))
	assert.Same(t, d, proposal.Track.MustEntries(v, 3)[0].Event)
	materialized := false
	for _, c := range proposal.Changes {
		for _, e := range c.To {
			if e.Event.Kind() == event.KindRest {
				materialized = true
				assert.Empty(t, c.From)
				assert.Empty(t, c.Input)
			}
		}
	}
	assert.True(t, materialized)
	for _, off := range []duration.Position{position(0, 1), position(1, 8)} {
		got, err := p.WriteEvents(v, measure.MustNewAddress(2, off), []*event.Event{q}, track.WriteOptions{})
		noProposal(t, got, err, track.ErrOccupied)
	}
	exact, err := p.WriteEvents(v, address(2, 1, 4), []*event.Event{q, q, q}, track.WriteOptions{})
	require.NoError(t, err)
	assert.Len(t, exact.Track.MustEntries(v, 2), 4)
	assert.Len(t, exact.Track.MustEntries(v, 3), 1)
	replaced, err := p.WriteEvents(v, address(2, 0, 1), []*event.Event{q}, track.WriteOptions{Mode: track.ReplaceVoice})
	require.NoError(t, err)
	assertReport(t, p, replaced)
	assert.Empty(t, replaced.Track.MustEntries(v, 3))
	assert.Len(t, replaced.Track.MustEntries(w, 1), 1)
	removed := false
	for _, c := range replaced.Changes {
		if len(c.From) > 0 && len(c.To) == 0 {
			removed = true
		}
		for _, e := range c.From {
			assert.True(t, v == e.Place.Voice)
		}
		for _, e := range c.To {
			assert.True(t, v == e.Place.Voice)
		}
	}
	assert.True(t, removed)
	unchanged := p.MustWriteEvents(v, address(1, 0, 1), nil, track.WriteOptions{})
	assert.True(t, p.Equal(unchanged.Track))
	cleared := p.MustWriteEvents(v, address(1, 0, 1), nil, track.WriteOptions{Mode: track.ReplaceVoice})
	assertPosition(t, 0, 1, cleared.Track.MustVoiceEnd(v))
	assert.Len(t, cleared.Track.MustEntries(w, 1), 1)
	for name, populated := range map[string]*track.Track{
		"explicit rest":         p.MustSplice(v, 1, 0, 0, rest(written(duration.Whole, 0), false)),
		"multibar continuation": p.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: v, Bar: 1, Events: []*event.Event{measureRest(2, false)}}, {Voice: v, Bar: 2}}}),
	} {
		t.Run(name, func(t *testing.T) {
			bar := 1
			if name == "multibar continuation" {
				bar = 2
			}
			out, err := populated.WriteEvents(v, address(bar, 0, 1), []*event.Event{q}, track.WriteOptions{})
			noProposal(t, out, err, track.ErrOccupied)
		})
	}
}

func TestWriteNeedBarsAndRetry(t *testing.T) {
	t.Parallel()
	p, v := fixture(1)
	input := []*event.Event{tone("C4", duration.Half, event.Marks{})}
	bad, err := p.WriteEvents(v, address(1, 3, 4), input, track.WriteOptions{})
	is := noProposal(t, bad, err, track.ErrNeedBars)
	assert.Equal(t, 0, is[0].InputIndex)
	assertPosition(t, 1, 1, is[0].Position)
	assertLength(t, 1, 4, is[0].Amount)
	assert.Empty(t, p.MustEntries(v, 1))
	g := grid(2)
	long := p.MustEdit(track.Edit{Grid: &g})
	out := long.MustWriteEvents(v, address(1, 3, 4), input, track.WriteOptions{})
	assert.Len(t, out.Track.MustEntries(v, 1), 2)
	assert.Len(t, out.Track.MustEntries(v, 2), 1)
	bad, err = p.WriteEvents(v, address(1, 0, 1), []*event.Event{measureRest(3, false)}, track.WriteOptions{})
	is = noProposal(t, bad, err, track.ErrNeedBars)
	assert.Equal(t, 2, is[0].BarsNeeded)
	_, err = p.WriteEvents(v, address(1, 1, 4), []*event.Event{measureRest(1, false)}, track.WriteOptions{})
	assert.ErrorIs(t, err, track.ErrMeasureRest)
	_, err = p.WriteEvents(v, address(2, 0, 1), nil, track.WriteOptions{})
	assert.ErrorIs(t, err, track.ErrPlace)
}

func TestSplitMergePreserveMarksSourcesAndNotation(t *testing.T) {
	t.Parallel()
	p, v := fixture(1)
	c, e := parsed("C4"), parsed("E4")
	original := chordEvent(written(duration.Half, 0), event.Marks{Arpeggio: true, Fermata: true}, c, e)
	tail := tone("G4", duration.Quarter, event.Marks{})
	p = p.MustSplice(v, 1, 0, 0, original, tail)
	out := p.MustSplit(track.Place{Voice: v, Bar: 1, Index: 0}, length(1, 4), track.KeepGroups)
	assertReport(t, p, out)
	wantFrom := []track.Entry{expectedEntry(v, 1, 0, original, 0, 1, 1, 2), expectedEntry(v, 1, 1, tail, 1, 2, 3, 4)}
	wantTo := []track.Entry{
		expectedEntry(v, 1, 0, chordEvent(written(duration.Quarter, 0), event.Marks{Arpeggio: true, Ties: []int{0, 1}}, c, e), 0, 1, 1, 4),
		expectedEntry(v, 1, 1, chordEvent(written(duration.Quarter, 0), event.Marks{Fermata: true}, c, e), 1, 4, 1, 2),
		expectedEntry(v, 1, 2, tail, 1, 2, 3, 4),
	}
	assertChanges(t, p, out, wantFrom, wantTo)
	assertNoInputs(t, out)
	es := out.Track.MustEntries(v, 1)
	require.Len(t, es, 3)
	assert.Equal(t, []int{0, 1}, es[0].Event.MustMarks().Ties)
	assert.True(t, es[0].Event.MustMarks().Arpeggio)
	assert.False(t, es[0].Event.MustMarks().Fermata)
	assert.Empty(t, es[1].Event.MustMarks().Ties)
	assert.False(t, es[1].Event.MustMarks().Arpeggio)
	assert.True(t, es[1].Event.MustMarks().Fermata)
	for _, entry := range es[:2] {
		for i, n := range entry.Event.MustContent().MustNotes() {
			assertNote(t, []note.Note{c, e}[i], n)
		}
	}
	merged := out.Track.MustMerge(v, 1, 0, 2, track.KeepGroups)
	assertReport(t, out.Track, merged)
	assertChanges(t, out.Track, merged, wantTo, wantFrom)
	assertNoInputs(t, merged)
	got := merged.Track.MustEntries(v, 1)
	require.Len(t, got, 2)
	assert.True(t, original.Equal(got[0].Event))
	assert.Len(t, p.MustEntries(v, 1), 2)
	for _, tc := range []struct {
		name        string
		first, last *event.Event
		cause       error
	}{
		{"attack", tone("C4", duration.Quarter, event.Marks{}), tone("C4", duration.Eighth, event.Marks{}), track.ErrWouldLoseMark},
		{"fermata", tone("C4", duration.Quarter, event.Marks{Ties: []int{0}, Fermata: true}), tone("C4", duration.Eighth, event.Marks{}), track.ErrWouldLoseMark},
		{"mixed form", sounded(c, value(1, 4), event.Marks{Ties: []int{0}}), tone("C4", duration.Eighth, event.Marks{}), nil},
		{"rest fermata", rest(written(duration.Quarter, 0), true), rest(written(duration.Quarter, 0), false), track.ErrWouldLoseMark},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, v := fixture(1)
			p = p.MustSplice(v, 1, 0, 0, tc.first, tc.last)
			out, err := p.Merge(v, 1, 0, 2, track.KeepGroups)
			if tc.cause != nil {
				noProposal(t, out, err, tc.cause)
			} else {
				// Mixed forms promise an addressed refusal, not a sentinel
				// or a preference for either of the two conflicting parts.
				require.Error(t, err)
				assert.Nil(t, out.Track)
				assert.Empty(t, out.Changes)
				assert.Empty(t, out.Marks)
				var problems *track.Problems
				require.ErrorAs(t, err, &problems)
				require.NotNil(t, problems)
				require.NotEmpty(t, problems.Issues)
				for _, i := range problems.Issues {
					assert.True(t, i.Place.Voice == v)
					assert.Equal(t, 1, i.Place.Bar)
					assert.GreaterOrEqual(t, i.Place.Index, 0)
					assert.Less(t, i.Place.Index, 2)
				}
				entries := p.MustEntries(v, 1)
				require.Len(t, entries, 2)
				assertEvent(t, tc.first, entries[0].Event)
				assertEvent(t, tc.last, entries[1].Event)
				assertSpan(t, entries[0].Time, 0, 1, 1, 4)
				assertSpan(t, entries[1].Time, 1, 4, 3, 8)
			}
			assert.Len(t, p.MustEntries(v, 1), 2)
		})
	}
	q := tone("C4", duration.Quarter, event.Marks{Ties: []int{0}})
	eighth := tone("C4", duration.Eighth, event.Marks{})
	dotted, v := fixture(1)
	dotted = dotted.MustSplice(v, 1, 0, 0, q, eighth)
	m := dotted.MustMerge(v, 1, 0, 2, track.KeepGroups)
	assert.True(t, m.Track.MustEntries(v, 1)[0].Event.MustDuration().Equal(written(duration.Quarter, 1)))
	rp, rv := fixture(1)
	rp = rp.MustSplice(rv, 1, 0, 0, rest(value(1, 8), false), rest(value(1, 4), false))
	rm := rp.MustMerge(rv, 1, 0, 2, track.KeepGroups)
	assert.True(t, rm.Track.MustEntries(rv, 1)[0].Event.MustDuration().Equal(value(3, 8)))
}

func TestTupletSplitChoiceAndUnnotatablePart(t *testing.T) {
	t.Parallel()
	p, v := fixture(1)
	triplet := duration.MustNewWritten(duration.Quarter, 0, tuplet.Triplet())
	e := sounded(parsed("C4"), event.MustWrittenDuration(triplet), event.Marks{})
	p = p.MustSplice(v, 1, 0, 0, e)
	at := track.Place{Voice: v, Bar: 1, Index: 0}
	keep := p.MustSplit(at, length(1, 12), track.KeepGroups)
	es := keep.Track.MustEntries(v, 1)
	require.Len(t, es, 2)
	for _, entry := range es {
		w := entry.Event.MustDuration().MustWritten()
		assert.True(t, w.Equal(duration.MustNewWritten(duration.Eighth, 0, tuplet.Triplet())))
	}
	out, err := p.Split(at, length(1, 12), track.WithoutGroups)
	is := noProposal(t, out, err, track.ErrNotation)
	assertLength(t, 1, 12, is[0].Amount)
	merged := keep.Track.MustMerge(v, 1, 0, 2, track.KeepGroups)
	assert.True(t, e.Equal(merged.Track.MustEntries(v, 1)[0].Event))
	// A numeric input is not forced into notation unless splitting requires it.
	numeric := sounded(parsed("D4"), value(1, 7), event.Marks{})
	np, nv := fixture(1)
	np = np.MustWriteEvents(nv, address(1, 0, 1), []*event.Event{numeric}, track.WriteOptions{}).Track
	assert.True(t, numeric.Equal(np.MustEntries(nv, 1)[0].Event))
}

func TestMeasureRestsAndMeterChange(t *testing.T) {
	t.Parallel()
	p, v := fixture(4)
	p, w := p.MustAddVoice()
	p = p.MustSplice(v, 1, 0, 0, measureRest(4, false)).MustSplice(w, 3, 0, 0, tone("C4", duration.Whole, event.Marks{}))
	g := grid(4).MustWithMark(3, meter.MustNew(3, 4))
	next, err := p.Edit(track.Edit{Grid: &g})
	assert.Nil(t, next)
	issues(t, err, track.ErrMeasureRest)
	// Clear the independently overflowing other voice in the same atomic edit.
	next = p.MustEdit(track.Edit{Grid: &g, Bars: []track.BarEdit{{Voice: v, Bar: 1, Events: []*event.Event{measureRest(2, false)}}, {Voice: v, Bar: 3, Events: []*event.Event{measureRest(2, false)}}, {Voice: w, Bar: 3}}})
	assertPosition(t, 7, 2, next.MustVoiceEnd(v))
	assert.Len(t, p.MustEntries(w, 3), 1)
	_, err = next.Splice(v, 2, 0, 0, tone("D4", duration.Quarter, event.Marks{}))
	assert.Error(t, err)
	short := grid(3)
	_, err = p.Edit(track.Edit{Grid: &short})
	assert.ErrorIs(t, err, track.ErrLostRecord)
}

func TestPackUnpackRestsAndFermataBarrier(t *testing.T) {
	t.Parallel()
	for _, firstFermata := range []bool{true, false} {
		p, v := fixture(2)
		p = p.MustEdit(track.Edit{Bars: []track.BarEdit{{Voice: v, Bar: 1, Events: []*event.Event{measureRest(1, firstFermata)}}, {Voice: v, Bar: 2, Events: []*event.Event{measureRest(1, true)}}}, Barlines: []track.BarlineEdit{{AfterBar: 1, Fermata: true}}})
		out, err := p.PackRests(v, 1, 2)
		if firstFermata {
			noProposal(t, out, err, track.ErrWouldLoseMark)
			assert.Len(t, p.MustEntries(v, 2), 1)
			continue
		}
		require.NoError(t, err)
		assertReport(t, p, out)
		wantFrom := []track.Entry{expectedEntry(v, 1, 0, measureRest(1, false), 0, 1, 1, 1), expectedEntry(v, 2, 0, measureRest(1, true), 1, 1, 2, 1)}
		wantTo := []track.Entry{expectedEntry(v, 1, 0, measureRest(2, true), 0, 1, 2, 1)}
		assertChanges(t, p, out, wantFrom, wantTo)
		assertNoInputs(t, out)
		e := out.Track.MustEntries(v, 1)
		require.Len(t, e, 1)
		assert.Equal(t, 2, e[0].Event.MustDuration().MustMeasureCount())
		assert.True(t, e[0].Event.MustMarks().Fermata)
		assert.Empty(t, out.Track.MustEntries(v, 2))
		assert.Equal(t, []int{1}, out.Track.MustFermataBars())
		back := out.Track.MustUnpackRest(e[0].Place)
		assertReport(t, out.Track, back)
		assertChanges(t, out.Track, back, wantTo, wantFrom)
		assertNoInputs(t, back)
		assert.True(t, p.Equal(back.Track))
	}
	g := grid(4).MustWithMark(3, meter.MustNew(3, 4))
	p := track.MustNew(piano(), g)
	v := p.MustVoices()[0]
	for bar := 1; bar <= 4; bar++ {
		p = p.MustSplice(v, bar, 0, 0, measureRest(1, false))
	}
	out := p.MustPackRests(v, 1, 4)
	assert.Equal(t, 2, out.Track.MustEntries(v, 1)[0].Event.MustDuration().MustMeasureCount())
	assert.Equal(t, 2, out.Track.MustEntries(v, 3)[0].Event.MustDuration().MustMeasureCount())
	one := p.MustPackRests(v, 2, 2)
	assert.True(t, p.Equal(one.Track))
	empty, ev := fixture(2)
	bad, err := empty.PackRests(ev, 1, 2)
	noProposal(t, bad, err, track.ErrMeasureRest)
}

func TestRebarPreservesSoundAndReportsRestMaterialization(t *testing.T) {
	t.Parallel()
	p, v := fixture(2)
	p = p.MustSplice(v, 1, 0, 0, measureRest(1, false)).MustSplice(v, 2, 0, 0, tone("C4", duration.Quarter, event.Marks{}))
	g := measure.MustNew(meter.MustNew(3, 4)).MustWithBarCount(2)
	out := p.MustRebar(g, track.KeepGroups)
	assertReport(t, p, out)
	q := tone("C4", duration.Quarter, event.Marks{})
	assertChanges(t, p, out,
		[]track.Entry{expectedEntry(v, 1, 0, measureRest(1, false), 0, 1, 1, 1), expectedEntry(v, 2, 0, q, 1, 1, 5, 4)},
		[]track.Entry{expectedEntry(v, 1, 0, rest(value(3, 4), false), 0, 1, 3, 4), expectedEntry(v, 2, 0, rest(value(1, 4), false), 3, 4, 1, 1), expectedEntry(v, 2, 1, q, 1, 1, 5, 4)})
	assertNoInputs(t, out)
	e1, e2 := out.Track.MustEntries(v, 1), out.Track.MustEntries(v, 2)
	require.Len(t, e1, 1)
	require.Len(t, e2, 2)
	assert.Equal(t, event.DurationValue, e1[0].Event.MustDuration().Kind())
	assertSpan(t, e1[0].Time, 0, 1, 3, 4)
	assert.Equal(t, event.KindRest, e2[0].Event.Kind())
	assertSpan(t, e2[0].Time, 3, 4, 1, 1)
	assert.Empty(t, e2[0].Event.MustMarks().Ties)
	assertSpan(t, e2[1].Time, 1, 1, 5, 4)
	assertSpan(t, out.Track.MustSounds()[0].Time, 1, 1, 5, 4)
	assert.Equal(t, event.DurationMeasures, p.MustEntries(v, 1)[0].Event.MustDuration().Kind())
	short := measure.MustNew(meter.MustNew(3, 4)).MustWithBarCount(1)
	bad, err := p.Rebar(short, track.KeepGroups)
	assert.Nil(t, bad.Track)
	assert.Error(t, err)
	// Rebar preserves positions; ordinary grid editing preserves bar numbers.
	edited := p.MustEdit(track.Edit{Grid: &g})
	assertSpan(t, edited.MustSounds()[0].Time, 3, 4, 1, 1)
}

func TestQuantizeSharedBoundariesTiesAndEmptyTail(t *testing.T) {
	t.Parallel()
	for _, tied := range []bool{false, true} {
		p, v := fixture(2)
		p, w := p.MustAddVoice()
		marks := event.Marks{}
		if tied {
			marks.Ties = []int{0}
		}
		first := sounded(parsed("C4"), written(duration.Eighth, 1), marks)
		second := sounded(parsed("C4"), written(duration.Eighth, 1), event.Marks{})
		if !tied {
			second = rest(value(5, 16), false)
		}
		p = p.MustSplice(v, 1, 0, 0, first, second).MustSplice(w, 1, 0, 0, tone("D3", duration.Whole, event.Marks{}))
		out := p.MustQuantize(v, 1, 1, length(1, 8), rounding.HalfUp, track.KeepGroups)
		assertReport(t, p, out)
		oldEndN, oldEndD := 1, 2
		newSecond := rest(value(1, 4), false)
		if tied {
			oldEndN, oldEndD = 3, 8
			newSecond = tone("C4", duration.Eighth, event.Marks{})
		}
		assertChanges(t, p, out,
			[]track.Entry{expectedEntry(v, 1, 0, first, 0, 1, 3, 16), expectedEntry(v, 1, 1, second, 3, 16, oldEndN, oldEndD)},
			[]track.Entry{expectedEntry(v, 1, 0, tone("C4", duration.Quarter, marks), 0, 1, 1, 4), expectedEntry(v, 1, 1, newSecond, 1, 4, oldEndN, oldEndD)})
		assertNoInputs(t, out)
		es := out.Track.MustEntries(v, 1)
		require.Len(t, es, 2)
		assertSpan(t, es[0].Time, 0, 1, 1, 4)
		if tied {
			assertSpan(t, es[1].Time, 1, 4, 3, 8)
			assert.True(t, es[0].Event.MustDuration().Equal(written(duration.Quarter, 0)))
			assert.True(t, es[1].Event.MustDuration().Equal(written(duration.Eighth, 0)))
			assertSpan(t, out.Track.MustSounds(v)[0].Time, 0, 1, 3, 8)
		} else {
			assertSpan(t, es[1].Time, 1, 4, 1, 2)
		}
		assert.Empty(t, out.Track.MustEntries(v, 2))
		assert.Same(t, p.MustEntries(w, 1)[0].Event, out.Track.MustEntries(w, 1)[0].Event)
		for _, c := range out.Changes {
			for _, e := range append(append([]track.Entry{}, c.From...), c.To...) {
				assert.True(t, v == e.Place.Voice)
			}
		}
		assertSpan(t, p.MustEntries(v, 1)[0].Time, 0, 1, 3, 16)
	}
	p, v := fixture(1)
	p = p.MustSplice(v, 1, 0, 0, sounded(parsed("C4"), written(duration.Eighth, 1), event.Marks{}))
	out := p.MustQuantize(v, 1, 1, length(1, 8), rounding.HalfUp, track.KeepGroups)
	assert.Len(t, out.Track.MustEntries(v, 1), 1)
	assertPosition(t, 1, 4, out.Track.MustVoiceEnd(v))
	_, ok := out.Track.MustAt(v, position(1, 4))
	assert.False(t, ok)
}

func TestQuantizeRefusesCollapseAndCrossBarAndPinsBarEnds(t *testing.T) {
	t.Parallel()
	p, v := fixture(1)
	e := tone("C4", duration.Sixteenth, event.Marks{})
	p = p.MustSplice(v, 1, 0, 0, e)
	out, err := p.Quantize(v, 1, 1, length(1, 4), rounding.HalfUp, track.KeepGroups)
	noProposal(t, out, err, track.ErrCollapsed)
	assert.Same(t, e, p.MustEntries(v, 1)[0].Event)
	g := grid(2).MustWithLength(1, length(23, 64), measure.AlignStart)
	p = track.MustNew(piano(), g)
	v = p.MustVoices()[0]
	e = sounded(parsed("C4"), value(11, 32), event.Marks{})
	p = p.MustSplice(v, 1, 0, 0, e)
	out, err = p.Quantize(v, 1, 1, length(1, 8), rounding.HalfUp, track.KeepGroups)
	is := noProposal(t, out, err, track.ErrCrossBar)
	assertPosition(t, 11, 32, is[0].Position)
	p = p.MustSplice(v, 1, 0, 1, rest(value(23, 64), false))
	for _, step := range []duration.Duration{length(1, 8), length(1, 2)} {
		same := p.MustQuantize(v, 1, 1, step, rounding.HalfUp, track.KeepGroups)
		assert.True(t, p.Equal(same.Track))
	}
}
