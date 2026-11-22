package score_test

import (
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/score"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/track"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-2: discarding an equal local clock is still an explicit operation.
func TestExplicitClockDiscard(t *testing.T) {
	t.Parallel()
	for _, speed := range []int{60, 90} {
		t.Run(bpm(speed).String(), func(t *testing.T) {
			p := emptyPart(4, 4, 1)
			local := track.Clock{Map: tempomap.MustNewMeasured(p.MustGrid(), bpm(speed)), BarHolds: []track.BarHold{{AfterBar: 1, Seconds: tempo.MustNewSeconds(2, 1)}}}
			p = p.MustEdit(track.Edit{Clock: &local})
			s := score.MustNew(clock60())
			out, id, err := s.Add(p)
			require.ErrorIs(t, err, score.ErrLocalClock)
			assert.Nil(t, out)
			assert.Zero(t, id)
			assert.Zero(t, problems(t, err).Issues[0].Part)
			accepted, err := s.AddWithoutClock(p)
			require.NoError(t, err)
			require.NotNil(t, accepted.Discarded)
			assert.True(t, accepted.Discarded.Map.Equal(local.Map))
			assert.Equal(t, local.BarHolds, accepted.Discarded.BarHolds)
			_, hasClock, e := accepted.Score.MustPart(accepted.Part).Clock()
			require.NoError(t, e)
			assert.False(t, hasClock)
			original, ok, e := p.Clock()
			require.NoError(t, e)
			assert.True(t, ok)
			assert.True(t, original.Map.Equal(local.Map))
			out, err = accepted.Score.Replace(accepted.Part, p)
			require.ErrorIs(t, err, score.ErrLocalClock)
			assert.Nil(t, out)
			assert.Equal(t, accepted.Part, problems(t, err).Issues[0].Part)
			replacement, err := accepted.Score.ReplaceWithoutClock(accepted.Part, p)
			require.NoError(t, err)
			assert.Equal(t, accepted.Part, replacement.Part)
			require.NotNil(t, replacement.Discarded)
			assert.True(t, replacement.Discarded.Map.Equal(local.Map))
			replacement.Discarded.BarHolds[0].AfterBar = 999
			preserved, ok, e := p.Clock()
			require.NoError(t, e)
			require.True(t, ok)
			assert.Equal(t, 1, preserved.BarHolds[0].AfterBar)
		})
	}
	p := emptyPart(4, 4, 1)
	inc := score.MustNew(clock60()).MustAddWithoutClock(p)
	assert.Nil(t, inc.Discarded)
	assert.Nil(t, inc.Score.MustReplaceWithoutClock(inc.Part, p).Discarded)
}

// K-8: the full record, including slots and modulation parameters, survives.
func TestExtractionPreservesEveryMapRecord(t *testing.T) {
	t.Parallel()
	clock := clock60().MustWithChange(tempomap.MustNewGradual(at(0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 2), bpm(120))).
		MustWithChange(tempomap.MustNewJump(at(3, 4), bpm(90))).
		MustWithChange(tempomap.MustNewATempo(at(1, 1))).
		MustWithChange(tempomap.MustNewModulation(at(5, 4), duration.MustNewWritten(duration.Quarter, 1), dur(1, 4))).
		MustWithChange(tempomap.MustNewJump(at(3, 2), bpm(80))).
		MustWithChange(tempomap.MustNewTempoI(at(7, 4))).
		MustWithExtension(tempomap.MustNewExtension(at(1, 4), tempo.MustNewSeconds(0, 1))).
		MustWithExtension(tempomap.MustNewExtension(at(1, 1), tempo.MustNewSeconds(1, 2)))
	for _, pickup := range []bool{false, true} {
		t.Run(map[bool]string{false: "different grid", true: "pickup"}[pickup], func(t *testing.T) {
			p := emptyPart(3, 4, 4)
			if pickup {
				g := p.MustGrid().MustWithPickup(dur(1, 4))
				p = p.MustEdit(track.Edit{Grid: &g})
			}
			p = write(p, 1, tone(p, "C4", 1, 4, event.Marks{}))
			s, ids := assemble(clock, p)
			plain := s.MustPart(ids[0])
			_, err := plain.SoundSeconds(3, rounding.HalfUp)
			require.ErrorIs(t, err, track.ErrNoClock)
			full, err := s.ExtractWithMap(ids[0])
			require.NoError(t, err)
			assert.True(t, full.MustPerformanceMap().Equal(clock))
			assert.True(t, full.MustGrid().Equal(p.MustGrid()))
			assert.Equal(t, p.MustVoices(), full.MustVoices())
			local, exists, err := full.Clock()
			require.NoError(t, err)
			assert.True(t, exists)
			assert.Equal(t, tempomap.MeasureAddressing, local.Map.MustAddressing())
			assert.Len(t, local.Map.MustChanges(), 8)
			edited := full.MustSplice(full.MustVoices()[0], 1, 0, 1, tone(p, "D4", 1, 4, event.Marks{}))
			assert.False(t, edited.Equal(full))
			assert.True(t, s.MustPart(ids[0]).Equal(p))
		})
	}
}

// K-17: terminal holds do not require a synthetic next bar or a fermata sign.
func TestExtractionTerminalHold(t *testing.T) {
	t.Parallel()
	for _, seconds := range []int{0, 1} {
		t.Run(tempo.MustNewSeconds(seconds, 1).String(), func(t *testing.T) {
			p := emptyPart(4, 4, 1)
			p = write(p, 1, tone(p, "C4", 1, 1, event.Marks{}))
			clock := clock60().MustWithExtension(tempomap.MustNewExtension(at(1, 1), tempo.MustNewSeconds(seconds, 1)))
			s, ids := assemble(clock, p)
			full := s.MustExtractWithMap(ids[0])
			local, ok, err := full.Clock()
			require.NoError(t, err)
			require.True(t, ok)
			require.Len(t, local.BarHolds, 1)
			assert.Equal(t, 1, local.BarHolds[0].AfterBar)
			assert.True(t, local.BarHolds[0].Seconds.Equal(tempo.MustNewSeconds(seconds, 1)))
			end, finite := full.MustGrid().MustEnd()
			assert.True(t, finite)
			assertPos(t, pos(1, 1), end)
			assert.True(t, full.MustPerformanceMap().Equal(clock))
			assert.Empty(t, full.MustFermataBars())
			assert.False(t, full.MustEntries(full.MustVoices()[0], 1)[0].Event.MustMarks().Fermata)
			assertSeconds(t, full.MustSoundSeconds(3, rounding.HalfUp)[0].End, 4+seconds, 1)
		})
	}
}

// K-18..K-20: reject loss, but preserve an intentionally unfinished tail.
func TestExtractionAddressFailuresAndUnfinishedTail(t *testing.T) {
	t.Parallel()
	p := emptyPart(4, 4, 1)
	p = write(p, 1, tone(p, "C4", 1, 1, event.Marks{}))
	for _, tc := range []struct {
		name  string
		clock tempomap.Map
		n, d  int
		kind  score.MarkKind
	}{
		{"equal tempo at end", clock60().MustWithChange(tempomap.MustNewMain(at(1, 1), bpm(60))), 1, 1, score.TempoMark},
		{"tempo after end", clock60().MustWithChange(tempomap.MustNewMain(at(2, 1), bpm(120))), 2, 1, score.TempoMark},
		{"zero hold after end", clock60().MustWithExtension(tempomap.MustNewExtension(at(3, 2), tempo.ZeroSeconds())), 3, 2, score.ExtensionMark},
		{"gradual endpoint beyond short grid", clock60().MustWithChange(tempomap.MustNewGradual(at(0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).MustWithChange(tempomap.MustNewMain(at(2, 1), bpm(120))), 2, 1, score.TempoMark},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, ids := assemble(tc.clock, p, emptyPart(4, 4, 3))
			got, err := s.ExtractWithMap(ids[0])
			require.ErrorIs(t, err, score.ErrMapAddress)
			assert.Nil(t, got)
			i := problems(t, err).Issues[0]
			assert.Equal(t, ids[0], i.Part)
			assertPos(t, pos(tc.n, tc.d), i.Position)
			require.NotNil(t, i.Mark)
			assert.Equal(t, tc.kind, i.Mark.Kind)
			assert.True(t, s.MustPart(ids[0]).Equal(p))
			assert.True(t, s.MustMap().Equal(tc.clock))
			longer := p.MustGrid().MustWithBarCount(3)
			updated := s.MustReplace(ids[0], p.MustEdit(track.Edit{Grid: &longer}))
			assert.True(t, updated.MustExtractWithMap(ids[0]).MustPerformanceMap().Equal(tc.clock))
		})
	}
	unfinished := clock60().MustWithChange(tempomap.MustNewGradual(at(1, 2), tempomap.LinearSpeed))
	s, ids := assemble(unfinished, p)
	full := s.MustExtractWithMap(ids[0])
	assert.True(t, full.MustPerformanceMap().Equal(unfinished))
	got, err := full.SoundSeconds(3, rounding.HalfUp)
	require.ErrorIs(t, err, tempomap.ErrUnfinished)
	assert.Nil(t, got)
	// Independent invalid marks are all reported in source order.
	bad := clock60().MustWithChange(tempomap.MustNewMain(at(1, 1), bpm(60))).MustWithExtension(tempomap.MustNewExtension(at(3, 2), tempo.ZeroSeconds())).MustWithChange(tempomap.MustNewMain(at(2, 1), bpm(90)))
	s, ids = assemble(bad, p)
	_, err = s.ExtractWithMap(ids[0])
	issues := problems(t, err).Issues
	require.Len(t, issues, 3)
	for i, want := range []duration.Position{pos(1, 1), pos(3, 2), pos(2, 1)} {
		assertPos(t, want, issues[i].Position)
	}
}

func TestBarViewMultibarRest(t *testing.T) {
	t.Parallel()
	p := emptyPart(4, 4, 3)
	p, v2 := p.MustAddVoice()
	v := p.MustVoices()[0]
	p = write(p, 1, rest(3, true))
	b := emptyPart(4, 4, 3)
	b = write(b, 2, tone(b, "C4", 1, 4, event.Marks{}))
	s, ids := assemble(clock60(), p, b)
	for bar := 1; bar <= 3; bar++ {
		view, err := s.BarView(ids[0], v, bar)
		require.NoError(t, err)
		require.Len(t, view, 1)
		assert.True(t, view[0].Expanded)
		assert.Equal(t, bar, view[0].Bar)
		assert.Equal(t, 1, view[0].Event.MustDuration().MustMeasureCount())
		assert.Equal(t, bar == 3, view[0].Event.MustMarks().Fermata)
		assert.Equal(t, 1, view[0].Source.Place.Bar)
		assert.Equal(t, 3, view[0].Source.Event.MustDuration().MustMeasureCount())
		assertPos(t, pos(0, 1), view[0].Source.Time.Start)
		assertPos(t, pos(3, 1), view[0].Source.Time.End)
		assertPos(t, pos(bar-1, 1), view[0].Time.Start)
		assertPos(t, pos(bar, 1), view[0].Time.End)
		assert.Empty(t, s.MustBarView(ids[0], v2, bar))
	}
	assert.True(t, s.MustAt(pos(1, 1))[1].Voices[0].Found)
	assert.Len(t, s.MustPart(ids[0]).MustEntries(v, 1), 1)
	assert.Empty(t, s.MustPart(ids[0]).MustEntries(v, 2))
	p = emptyPart(4, 4, 1)
	p = write(p, 1, rest(1, false))
	s, ids = assemble(clock60(), p)
	one := s.MustBarView(ids[0], p.MustVoices()[0], 1)
	require.Len(t, one, 1)
	assert.False(t, one[0].Expanded)
	assert.Equal(t, one[0].Source.Time, one[0].Time)
	// Unknown bar retains Track's place classification.
	got, err := s.BarView(ids[0], p.MustVoices()[0], 2)
	require.ErrorIs(t, err, track.ErrPlace)
	assert.Nil(t, got)
}
