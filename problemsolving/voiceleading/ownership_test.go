package voiceleading_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertResult(t *testing.T, want, got any) {
	t.Helper()
	switch w := want.(type) {
	case voiceleading.Vertical:
		assertVertical(t, w, got.(voiceleading.Vertical))
	case voiceleading.Transition:
		assertTransition(t, w, got.(voiceleading.Transition))
	case []voiceleading.Vertical:
		g := got.([]voiceleading.Vertical)
		require.Len(t, g, len(w))
		for i := range w {
			assertVertical(t, w[i], g[i])
		}
	case []voiceleading.Transition:
		g := got.([]voiceleading.Transition)
		require.Len(t, g, len(w))
		for i := range w {
			assertTransition(t, w[i], g[i])
		}
	case []voiceleading.Crossing:
		g := got.([]voiceleading.Crossing)
		require.Len(t, g, len(w))
		for i := range w {
			assert.Equal(t, w[i].Pair, g[i].Pair)
			assertPosition(t, w[i].Start, g[i].Start)
			assertPosition(t, w[i].End, g[i].End)
			assertResult(t, w[i].Verticals, g[i].Verticals)
		}
	case []voiceleading.Overlap:
		g := got.([]voiceleading.Overlap)
		require.Len(t, g, len(w))
		for i := range w {
			assert.Equal(t, w[i].FirstAbovePreviousSecond, g[i].FirstAbovePreviousSecond)
			assert.Equal(t, w[i].SecondBelowPreviousFirst, g[i].SecondBelowPreviousFirst)
			assertTransition(t, w[i].Transition, g[i].Transition)
		}
	case *voiceleading.Assessment:
		g := got.(*voiceleading.Assessment)
		require.NotNil(t, g)
		assert.Equal(t, w.Rule, g.Rule)
		assert.Equal(t, w.Scope, g.Scope)
		assert.Equal(t, w.Applicable, g.Applicable)
		require.Len(t, g.Issues, len(w.Issues))
		require.Len(t, g.Violations, len(w.Violations))
		for i := range w.Issues {
			assert.Equal(t, w.Issues[i].Reason, g.Issues[i].Reason)
			assertVertical(t, w.Issues[i].Vertical, g.Issues[i].Vertical)
		}
		for i := range w.Violations {
			assert.Equal(t, w.Violations[i].Reason, g.Violations[i].Reason)
			assertTransition(t, w.Violations[i].Transition, g.Violations[i].Transition)
		}
	case *system.System:
		g := got.(*system.System)
		if w == nil {
			assert.Nil(t, g)
		} else {
			assert.True(t, w.Equal(g))
		}
	default:
		assert.Equal(t, want, got)
	}
}

func TestAnalysisOwnsInputsAndReturnedData(t *testing.T) {
	t.Parallel()
	low, high := line(t, "C4", "D4"), line(t, "G4", "A4")
	inputs := []*voice.Line{low, high}
	opts := voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh}
	a := voiceleading.MustAnalyze(inputs, opts)
	same := voiceleading.MustAnalyze([]*voice.Line{line(t, "C4", "D4"), line(t, "G4", "A4")}, opts)
	inputs[0] = line(t, "F4")
	*high = *line(t, "B4")
	assert.True(t, a.Equal(same))
	copyOfAnalysis := *a
	assert.True(t, copyOfAnalysis.Equal(a))
	pairs := a.MustPairs()
	pairs[0] = voiceleading.Pair{First: 19, Second: 20}
	assert.Equal(t, []voiceleading.Pair{pair}, a.MustPairs())
	verticals := a.MustVerticals(pair)
	verticals[0].Chromatic = 999
	verticals[0].First.Index = 99
	verticals[0].Specific = interval.Interval{}
	verticals[0].Start = duration.Position{}
	assertResult(t, same.MustVerticals(pair), a.MustVerticals(pair))
	v, ok := a.MustVerticalAt(pair, pos(0, 1))
	require.True(t, ok)
	v.First.Element = rest(1, 4)
	assertResult(t, same.MustVerticals(pair), a.MustVerticals(pair))
	n, ok := a.MustVerticals(pair)[0].First.Element.MustNote()
	require.True(t, ok)
	require.NoError(t, n.Unbind())
	require.NoError(t, n.ReplaceWithPitch(voiceNotePitch(t, "F", 5)))
	assert.True(t, a.Equal(same))
	transitions := a.MustTransitions(pair)
	transitions[0].Motion = voiceleading.MotionUnclassified
	transitions[0].To.Second.Index = 88
	assertResult(t, same.MustTransitions(pair), a.MustTransitions(pair))
	for _, get := range []func() []voiceleading.Transition{func() []voiceleading.Transition { return a.MustFindMotion(voiceleading.MotionParallel) }, func() []voiceleading.Transition {
		return a.MustFindParallels(voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth})
	}, func() []voiceleading.Transition { return a.MustFindConsecutive(voiceleading.PerfectBoth) }, func() []voiceleading.Transition {
		return a.MustFindDirect(voiceleading.PerfectBoth, voiceleading.AnyOrigin)
	}} {
		before := get()
		changed := get()
		require.NotEmpty(t, changed)
		changed[0].From.Chromatic = 777
		changed[0].First = interval.Interval{}
		assertResult(t, before, get())
	}
	assessment := a.MustCheckOMTFirstSpecies(pair)
	before := a.MustCheckOMTFirstSpecies(pair)
	assessment.Rule = "changed"
	assessment.Scope = "changed"
	assessment.Applicable = false
	assessment.Violations[0].Reason = "changed"
	assessment.Violations[0].Transition.To.Chromatic = 888
	assertResult(t, before, a.MustCheckOMTFirstSpecies(pair))
	assert.True(t, copyOfAnalysis.Equal(same))
}

func TestNestedResultsAreIndependent(t *testing.T) {
	t.Parallel()
	crossed := analyze(t, []string{"F4", "G4"}, []string{"D4", "E4"})
	before := crossed.MustFindCrossings()
	xs := crossed.MustFindCrossings()
	require.Len(t, xs, 1)
	xs[0].Verticals[0].First.Index = 999
	xs[0].Pair = voiceleading.Pair{}
	xs[0].Start = duration.Position{}
	assertResult(t, before, crossed.MustFindCrossings())
	overlap := analyze(t, []string{"C4", "F4"}, []string{"E4", "G4"})
	os := overlap.MustFindOverlaps()
	require.Len(t, os, 1)
	old := overlap.MustFindOverlaps()
	os[0].FirstAbovePreviousSecond = false
	os[0].Transition.From.Chromatic = 99
	assertResult(t, old, overlap.MustFindOverlaps())
	async := voiceleading.MustAnalyze([]*voice.Line{voice.MustNew(sound(t, "C", 4, 1, 2, nil)), line(t, "G4", "A4")}, voiceleading.Options{System: system.Traditional()})
	expected := async.MustCheckOMTFirstSpecies(pair)
	result := async.MustCheckOMTFirstSpecies(pair)
	require.NotEmpty(t, result.Issues)
	result.Issues[0].Reason = "changed"
	result.Issues[0].Vertical.First.Index = 999
	assertResult(t, expected, async.MustCheckOMTFirstSpecies(pair))
	// DTO assignment shares only the caller-owned slice, as documented.
	caller := crossed.MustFindCrossings()
	assigned := caller[0]
	assigned.Verticals[0].Chromatic = 123
	assert.Equal(t, system.Step(123), caller[0].Verticals[0].Chromatic)
	assertResult(t, before, crossed.MustFindCrossings())
}

func voiceNotePitch(t *testing.T, name string, octave int) (p pitch.Pitch) {
	t.Helper()
	n, ok := sound(t, name, octave, 1, 4, nil).MustNote()
	require.True(t, ok)
	p, err := n.Pitch()
	require.NoError(t, err)
	return p
}

func TestEqualSnapshotSemantics(t *testing.T) {
	t.Parallel()
	var nilAnalysis *voiceleading.Analysis
	zero := new(voiceleading.Analysis)
	assert.True(t, nilAnalysis.Equal(nil))
	assert.True(t, nilAnalysis.Equal(zero))
	assert.True(t, zero.Equal(nilAnalysis))
	assert.True(t, zero.Equal(new(voiceleading.Analysis)))
	for _, tt := range []struct {
		name      string
		low, high []string
	}{{"all rests", []string{"-", "-"}, []string{"-", "-"}}, {"unclassified transitions", []string{"C4", "-", "D4"}, []string{"G4", "G4", "A4"}}, {"sound", []string{"C4", "D4"}, []string{"G4", "A4"}}, {"empty", nil, nil}} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := analyze(t, tt.low, tt.high), analyze(t, tt.low, tt.high)
			assert.True(t, a.Equal(a))
			assert.True(t, a.Equal(b))
			assert.True(t, b.Equal(a))
			assert.False(t, a.Equal(nilAnalysis))
			assert.False(t, zero.Equal(a))
		})
	}
	a := analyze(t, []string{"C4", "D4"}, []string{"G4", "A4"})
	for _, tt := range []struct {
		name  string
		lines []*voice.Line
		opts  voiceleading.Options
	}{
		{"names", []*voice.Line{line(t, "C4", "D4").MustWithName("renamed"), line(t, "G4", "A4")}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh}},
		{"part order", []*voice.Line{line(t, "G4", "A4"), line(t, "C4", "D4")}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh}},
		{"nominal order", []*voice.Line{line(t, "C4", "D4"), line(t, "G4", "A4")}, voiceleading.Options{System: system.Traditional()}},
		{"context", []*voice.Line{line(t, "C4", "D4"), line(t, "G4", "A4")}, voiceleading.Options{System: system.Positions19(), Order: voiceleading.LowToHigh}},
		{"present and absent intervals", []*voice.Line{line(t, "C4", "-"), line(t, "G4", "A4")}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b := voiceleading.MustAnalyze(tt.lines, tt.opts)
			assert.False(t, a.Equal(b))
			assert.False(t, b.Equal(a))
		})
	}
	emptyWithout := voiceleading.MustAnalyze([]*voice.Line{line(t), line(t)}, voiceleading.Options{})
	emptyWith := voiceleading.MustAnalyze([]*voice.Line{line(t), line(t)}, voiceleading.Options{System: system.Traditional()})
	assert.False(t, emptyWith.Equal(emptyWithout))
	// Source identity is retained through Element.Equal, while Equal never rereads it.
	src := &testSource{current: settings.Standard()}
	inputs := []*voice.Line{voice.MustNew(sound(t, "C", 4, 1, 4, src)), voice.MustNew(sound(t, "G", 4, 1, 4, src))}
	left := voiceleading.MustAnalyze(inputs, voiceleading.Options{})
	right := voiceleading.MustAnalyze(inputs, voiceleading.Options{})
	src.reset(nil)
	assert.True(t, left.Equal(right))
	assert.Zero(t, src.count())
}

func TestConcurrentReadOperations(t *testing.T) {
	t.Parallel()
	a := analyze(t, []string{"C4", "D4", "F4", "-", "G4"}, []string{"G4", "A4", "D4", "E4", "C4"})
	type observation struct {
		results   [][]reflect.Value
		selfEqual bool
	}
	expected := make([][]reflect.Value, 0)
	for _, op := range operations(a) {
		expected = append(expected, call(op.normal, op.args))
	}
	results := make(chan observation, 16)
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			out := observation{selfEqual: a.Equal(a)}
			for _, op := range operations(a) {
				out.results = append(out.results, call(op.normal, op.args))
			}
			results <- out
		})
	}
	wg.Wait()
	close(results)
	for result := range results {
		assert.True(t, result.selfEqual)
		for i, values := range result.results {
			require.True(t, values[len(values)-1].IsNil())
			for j, v := range values[:len(values)-1] {
				assertResult(t, expected[i][j].Interface(), v.Interface())
			}
		}
	}
}
