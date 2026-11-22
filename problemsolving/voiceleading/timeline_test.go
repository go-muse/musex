package voiceleading_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPairTimeline(t *testing.T) {
	t.Parallel()
	tied := voice.MustTiedDuration(duration.MustNewTied(duration.MustNewWritten(duration.Quarter, 0), duration.MustNewWritten(duration.Quarter, 0)))
	held := sound(t, "C", 4, 1, 2, nil).MustWithDuration(tied).MustWithFermata(true)
	low := voice.MustNew(held, sound(t, "D", 4, 1, 3, nil))
	high := voice.MustNew(rest(1, 6), sound(t, "E", 4, 1, 3, nil), sound(t, "F", 4, 1, 4, nil))
	a, err := voiceleading.Analyze([]*voice.Line{low, high}, voiceleading.Options{System: system.Traditional()})
	require.NoError(t, err)
	vs, err := a.Verticals(pair)
	require.NoError(t, err)
	require.Len(t, vs, 3)
	starts := []duration.Position{pos(0, 1), pos(1, 6), pos(1, 2)}
	ends := []duration.Position{pos(1, 6), pos(1, 2), pos(3, 4)}
	for i, v := range vs {
		assert.Equal(t, pair, v.Pair)
		assertPosition(t, starts[i], v.Start)
		assertPosition(t, ends[i], v.End)
	}
	assert.Equal(t, []int{0, 0, 1}, []int{vs[0].First.Index, vs[1].First.Index, vs[2].First.Index})
	assert.Equal(t, []int{0, 1, 2}, []int{vs[0].Second.Index, vs[1].Second.Index, vs[2].Second.Index})
	assert.True(t, held.Equal(vs[1].First.Element))
	assert.True(t, vs[1].First.Element.MustFermata())
	assert.Equal(t, voice.DurationTied, vs[1].First.Element.MustDuration().Kind())
	assertPosition(t, pos(0, 1), vs[1].First.Start)
	assertPosition(t, pos(1, 2), vs[1].First.End)
	assertPosition(t, pos(5, 6), vs[2].First.End) // Original end survives clipping at 3/4.
	assert.False(t, vs[0].Sounding)
	assert.True(t, vs[0].Second.Element.IsRest())
	assert.False(t, vs[0].Specific.IsValid())
	assert.Zero(t, vs[0].Chromatic)
	assertInterval(t, vs[1].Specific, 3, interval.Major, interval.Up)
	assert.Equal(t, system.Step(4), vs[1].Chromatic)
	for _, at := range []duration.Position{pos(1, 6), pos(1, 3), pos(2, 6)} {
		v, ok, err := a.VerticalAt(pair, at)
		require.NoError(t, err)
		require.True(t, ok)
		assertVertical(t, vs[1], v)
	}
	for _, at := range []duration.Position{pos(3, 4), pos(1, 1)} {
		v, ok, err := a.VerticalAt(pair, at)
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Zero(t, v)
	}
	ts := a.MustTransitions(pair)
	require.Len(t, ts, 2)
	assert.Equal(t, voiceleading.MotionUnclassified, ts[0].Motion)
	assert.False(t, ts[0].First.IsValid())
	assert.False(t, ts[0].Second.IsValid())
	assert.Zero(t, ts[0].FirstChromatic)
	assert.Zero(t, ts[0].SecondChromatic)
	assertTransition(t, ts[1], a.MustTransitionAt(pair, pos(1, 6), pos(1, 2)))

	// A third part must not subdivide this pair's timeline.
	b, err := voiceleading.Analyze([]*voice.Line{low, high, voice.MustNew(rest(1, 7), rest(1, 7), rest(1, 7), rest(1, 7), rest(1, 7), rest(1, 7))}, voiceleading.Options{System: system.Traditional()})
	require.NoError(t, err)
	more := b.MustVerticals(pair)
	require.Len(t, more, len(vs))
	for i := range vs {
		assertVertical(t, vs[i], more[i])
	}
}

func TestHeldNoteAndRepeatedAttacks(t *testing.T) {
	t.Parallel()
	low := voice.MustNew(sound(t, "C", 4, 1, 2, nil))
	high := line(t, "E4", "F4")
	a := voiceleading.MustAnalyze([]*voice.Line{low, high}, voiceleading.Options{System: system.Traditional()})
	ts := a.MustTransitions(pair)
	require.Len(t, ts, 1)
	assert.Equal(t, voiceleading.MotionOblique, ts[0].Motion)
	assertInterval(t, ts[0].First, 1, interval.Perfect, interval.Up)
	assertInterval(t, ts[0].Second, 2, interval.Minor, interval.Up)
	repeated := analyze(t, []string{"C4", "C4"}, []string{"G4", "G4"})
	require.Len(t, repeated.MustVerticals(pair), 2)
	assert.Equal(t, voiceleading.MotionStationary, repeated.MustTransitions(pair)[0].Motion)
	assert.Len(t, repeated.MustFindConsecutive(voiceleading.PerfectFifths), 1)
	assert.Empty(t, repeated.MustFindParallels(voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth}))
}

func TestRestsBreakEverySearch(t *testing.T) {
	t.Parallel()
	a := analyze(t, []string{"C4", "-", "-", "D4"}, []string{"G4", "G4", "G4", "A4"})
	vs := a.MustVerticals(pair)
	require.Len(t, vs, 4) // Adjacent rests are distinct original elements.
	ts := a.MustTransitions(pair)
	require.Len(t, ts, 3)
	for _, tr := range ts {
		assert.Equal(t, voiceleading.MotionUnclassified, tr.Motion)
		assert.False(t, tr.First.IsValid())
		assert.False(t, tr.Second.IsValid())
		assert.Zero(t, tr.FirstChromatic)
		assert.Zero(t, tr.SecondChromatic)
	}
	v, ok := a.MustVerticalAt(pair, pos(3, 8))
	assert.True(t, ok)
	assert.True(t, v.First.Element.IsRest())
	assert.Empty(t, a.MustFindMotion(voiceleading.MotionDirect))
	assert.Empty(t, a.MustFindConsecutive(voiceleading.PerfectBoth))
	assert.Empty(t, a.MustFindDirect(voiceleading.PerfectBoth, voiceleading.AnyOrigin))
	assert.Empty(t, a.MustFindParallels(voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth}))
	assert.True(t, a.MustCheckOMTFirstSpecies(pair).Applicable)
	assert.Empty(t, a.MustCheckOMTFirstSpecies(pair).Violations)
}

func TestTransitionRequiresAdjacentExactStarts(t *testing.T) {
	t.Parallel()
	a := analyze(t, []string{"C4", "D4", "E4"}, []string{"G4", "A4", "B4"})
	cases := []struct {
		name     string
		from, to duration.Position
	}{
		{"same", pos(0, 1), pos(0, 1)}, {"reverse", pos(1, 4), pos(0, 1)},
		{"skip", pos(0, 1), pos(1, 2)}, {"inside from", pos(1, 8), pos(1, 4)},
		{"inside to", pos(0, 1), pos(1, 8)}, {"end", pos(1, 2), pos(3, 4)},
		{"beyond", pos(3, 4), pos(1, 1)},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tr, err := a.TransitionAt(pair, tt.from, tt.to)
			assert.ErrorIs(t, err, voiceleading.ErrTransition)
			assert.Zero(t, tr)
		})
	}
	assertTransition(t, a.MustTransitions(pair)[0], a.MustTransitionAt(pair, pos(0, 3), pos(2, 8)))
}
