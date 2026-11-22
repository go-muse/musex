package voiceleading_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrittenMotion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name                            string
		low, high                       []string
		motion                          voiceleading.Motion
		firstSize, secondSize           int
		firstQuality, secondQuality     interval.Quality
		firstDirection, secondDirection interval.Direction
		firstSteps, secondSteps         system.Step
	}{
		{"different qualities same generic size", []string{"C4", "D4"}, []string{"E4", "F4"}, voiceleading.MotionParallel, 2, 2, interval.Major, interval.Minor, interval.Up, interval.Up, 2, 1},
		{"direct seconds and thirds", []string{"C4", "D4"}, []string{"E4", "G4"}, voiceleading.MotionDirect, 2, 3, interval.Major, interval.Minor, interval.Up, interval.Up, 2, 3},
		{"contrary", []string{"C4", "B3"}, []string{"E4", "F4"}, voiceleading.MotionContrary, 2, 2, interval.Minor, interval.Minor, interval.Down, interval.Up, -1, 1},
		{"oblique lower held", []string{"C4", "C4"}, []string{"E4", "F4"}, voiceleading.MotionOblique, 1, 2, interval.Perfect, interval.Minor, interval.Up, interval.Up, 0, 1},
		{"oblique upper held", []string{"C4", "D4"}, []string{"G4", "G4"}, voiceleading.MotionOblique, 2, 1, interval.Major, interval.Perfect, interval.Up, interval.Up, 2, 0},
		{"stationary reattacks", []string{"C4", "C4"}, []string{"E4", "E4"}, voiceleading.MotionStationary, 1, 1, interval.Perfect, interval.Perfect, interval.Up, interval.Up, 0, 0},
		{"parallel descending unequal qualities", []string{"C4", "B3"}, []string{"G4", "F4"}, voiceleading.MotionParallel, 2, 2, interval.Minor, interval.Major, interval.Down, interval.Down, -1, -2},
		{"full generic motion differs", []string{"C4", "D4"}, []string{"G4", "A5"}, voiceleading.MotionDirect, 2, 9, interval.Major, interval.Major, interval.Up, interval.Up, 2, 14},
		{"enharmonic written steps with zero distance", []string{"C#4", "Db4"}, []string{"F#4", "Gb4"}, voiceleading.MotionParallel, 2, 2, interval.Diminished, interval.Diminished, interval.Up, interval.Up, 0, 0},
		{"written direction opposite to chromatic", []string{"C##4", "Db4"}, []string{"G4", "A4"}, voiceleading.MotionParallel, 2, 2, interval.Diminished, interval.Major, interval.Up, interval.Up, -1, 2},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := analyze(t, tt.low, tt.high)
			ts := a.MustTransitions(pair)
			require.Len(t, ts, 1)
			tr := ts[0]
			assert.Equal(t, tt.motion, tr.Motion)
			assertInterval(t, tr.First, tt.firstSize, tt.firstQuality, tt.firstDirection)
			assertInterval(t, tr.Second, tt.secondSize, tt.secondQuality, tt.secondDirection)
			assert.Equal(t, tt.firstSteps, tr.FirstChromatic)
			assert.Equal(t, tt.secondSteps, tr.SecondChromatic)
			assert.Len(t, a.MustFindMotion(tt.motion), 1)
			directs := a.MustFindMotion(voiceleading.MotionDirect)
			if tt.motion.IsDirect() {
				assert.Len(t, directs, 1)
			} else {
				assert.Empty(t, directs)
			}
			if tt.motion != voiceleading.MotionParallel {
				assert.Empty(t, a.MustFindMotion(voiceleading.MotionParallel))
			}
		})
	}
}

func TestPerfectSearches(t *testing.T) {
	t.Parallel()
	// Fixed expectations come from VL-A6 and the accepted musical examples.
	cases := []struct {
		name                                               string
		low, high                                          []string
		perfect                                            voiceleading.PerfectSet
		compound, unisons                                  bool
		parallels, consecutive, anyOrigin, differentFamily int
	}{
		{"fifths", []string{"C4", "D4", "E4"}, []string{"G4", "A4", "C5"}, voiceleading.PerfectFifths, false, false, 1, 1, 1, 0},
		{"octaves", []string{"C4", "D4"}, []string{"C5", "D5"}, voiceleading.PerfectOctaves, false, false, 1, 1, 1, 0},
		{"octaves not fifths", []string{"C4", "D4"}, []string{"C5", "D5"}, voiceleading.PerfectFifths, true, true, 0, 0, 0, 0},
		{"compound fifth excluded", []string{"C4", "D4"}, []string{"G5", "A5"}, voiceleading.PerfectBoth, false, false, 0, 1, 1, 0},
		{"higher compound fifth included", []string{"C3", "D3"}, []string{"G5", "A5"}, voiceleading.PerfectFifths, true, false, 1, 1, 1, 0},
		{"triple octave included", []string{"C3", "D3"}, []string{"C6", "D6"}, voiceleading.PerfectOctaves, true, false, 1, 1, 1, 0},
		{"compound fifth included", []string{"C4", "D4"}, []string{"G5", "A5"}, voiceleading.PerfectFifths, true, false, 1, 1, 1, 0},
		{"double octave excluded", []string{"C3", "D3"}, []string{"C5", "D5"}, voiceleading.PerfectOctaves, false, false, 0, 1, 1, 0},
		{"double octave included", []string{"C3", "D3"}, []string{"C5", "D5"}, voiceleading.PerfectOctaves, true, false, 1, 1, 1, 0},
		{"unisons excluded", []string{"C4", "D4"}, []string{"C4", "D4"}, voiceleading.PerfectBoth, true, false, 0, 1, 1, 0},
		{"unisons independent of compound", []string{"C4", "D4"}, []string{"C4", "D4"}, voiceleading.PerfectOctaves, false, true, 1, 1, 1, 0},
		{"unisons not fifths", []string{"C4", "D4"}, []string{"C4", "D4"}, voiceleading.PerfectFifths, true, true, 0, 0, 0, 0},
		{"contrary fifth to twelfth", []string{"C4", "G3"}, []string{"G4", "D5"}, voiceleading.PerfectBoth, true, true, 0, 1, 0, 0},
		{"direct fifth to twelfth", []string{"C4", "D4"}, []string{"G4", "A5"}, voiceleading.PerfectBoth, true, true, 0, 1, 1, 0},
		{"stationary perfect fifth", []string{"C4", "C4"}, []string{"G4", "G4"}, voiceleading.PerfectBoth, true, true, 0, 1, 0, 0},
		{"diminished destination", []string{"C4", "B3"}, []string{"G4", "F4"}, voiceleading.PerfectBoth, true, true, 0, 0, 0, 0},
		{"enharmonic distance is not perfect fifth", []string{"Ab3", "G3"}, []string{"D#4", "D4"}, voiceleading.PerfectBoth, true, true, 0, 0, 1, 1},
		{"direct from third", []string{"C4", "D4"}, []string{"E4", "A4"}, voiceleading.PerfectFifths, false, false, 0, 0, 1, 1},
		{"same octave family different full size", []string{"C4", "D4"}, []string{"C4", "D5"}, voiceleading.PerfectOctaves, true, true, 0, 1, 1, 0},
		{"descending vertical fifths", []string{"G4", "A4"}, []string{"C4", "D4"}, voiceleading.PerfectFifths, false, false, 1, 1, 1, 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := analyze(t, tt.low, tt.high)
			ps := a.MustFindParallels(voiceleading.ParallelOptions{Perfect: tt.perfect, Compound: tt.compound, Unisons: tt.unisons})
			cs := a.MustFindConsecutive(tt.perfect)
			ds := a.MustFindDirect(tt.perfect, voiceleading.AnyOrigin)
			different := a.MustFindDirect(tt.perfect, voiceleading.DifferentFamily)
			assert.Len(t, ps, tt.parallels)
			assert.Len(t, cs, tt.consecutive)
			assert.Len(t, ds, tt.anyOrigin)
			assert.Len(t, different, tt.differentFamily)
			if len(ps) > 0 {
				assertPosition(t, pos(0, 1), ps[0].From.Start)
				assertPosition(t, pos(1, 4), ps[0].To.Start)
			}
			if len(cs) > 0 && len(ds) > 0 {
				assertTransition(t, cs[0], ds[0])
			}
		})
	}
}

func TestMultivoicePairAndSearchOrdering(t *testing.T) {
	t.Parallel()
	a := voiceleading.MustAnalyze([]*voice.Line{line(t, "C3", "D3", "E3"), line(t, "G3", "A3", "B3"), line(t, "C4", "D4", "E4")}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh})
	assert.Equal(t, []voiceleading.Pair{{First: 0, Second: 1}, {First: 0, Second: 2}, {First: 1, Second: 2}}, a.MustPairs())
	found := a.MustFindParallels(voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth})
	require.Len(t, found, 4)
	assert.Equal(t, []voiceleading.Pair{{First: 0, Second: 1}, {First: 0, Second: 1}, {First: 0, Second: 2}, {First: 0, Second: 2}}, []voiceleading.Pair{found[0].From.Pair, found[1].From.Pair, found[2].From.Pair, found[3].From.Pair})
	assertPosition(t, pos(0, 1), found[0].From.Start)
	assertPosition(t, pos(1, 4), found[1].From.Start)
	assertPosition(t, pos(0, 1), found[2].From.Start)
	assertPosition(t, pos(1, 4), found[3].From.Start)
	assert.Equal(t, 5, found[0].From.Specific.Size())
	assert.Equal(t, 8, found[2].From.Specific.Size())
	assert.Len(t, a.MustFindMotion(voiceleading.MotionParallel), 6)
	assert.Len(t, a.MustFindConsecutive(voiceleading.PerfectBoth), 4)
	assert.Len(t, a.MustFindDirect(voiceleading.PerfectBoth, voiceleading.AnyOrigin), 4)
}

func TestCrossingRuns(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		low, high    []string
		starts, ends []int
		lengths      []int
	}{
		{"begins at origin and ends at unison", []string{"F4", "G4", "E4", "F4"}, []string{"D4", "E4", "E4", "D4"}, []int{0, 3}, []int{2, 4}, []int{2, 1}},
		{"rest separates runs", []string{"F4", "-", "F4"}, []string{"D4", "D4", "D4"}, []int{0, 2}, []int{1, 3}, []int{1, 1}},
		{"restored order terminates run", []string{"F4", "C4"}, []string{"D4", "E4"}, []int{0}, []int{1}, []int{1}},
		{"enharmonic equality is not crossing", []string{"C#4"}, []string{"Db4"}, nil, nil, nil},
		{"written above but chromatically below", []string{"Db4"}, []string{"C##4"}, nil, nil, nil},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := analyze(t, tt.low, tt.high)
			xs := a.MustFindCrossings()
			require.Len(t, xs, len(tt.starts))
			vs := a.MustVerticals(pair)
			for i, x := range xs {
				assert.Equal(t, pair, x.Pair)
				assertPosition(t, pos(tt.starts[i], 4), x.Start)
				assertPosition(t, pos(tt.ends[i], 4), x.End)
				require.Len(t, x.Verticals, tt.lengths[i])
				for j, v := range x.Verticals {
					assertVertical(t, vs[tt.starts[i]+j], v)
				}
			}
		})
	}
	// The shorter part clips a run, but not its source element.
	a := voiceleading.MustAnalyze([]*voice.Line{voice.MustNew(sound(t, "F", 4, 1, 1, nil)), line(t, "D4")}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh})
	xs := a.MustFindCrossings()
	require.Len(t, xs, 1)
	assertPosition(t, pos(1, 4), xs[0].End)
	assertPosition(t, pos(1, 1), xs[0].Verticals[0].First.End)
}

func TestOverlapBoundaries(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name          string
		low, high     []string
		count         int
		first, second bool
	}{
		{"lower enters previous upper space", []string{"C4", "F4"}, []string{"E4", "G4"}, 1, true, false},
		{"upper enters previous lower space", []string{"C4", "B3"}, []string{"E4", "B3"}, 1, false, true},
		{"both cross their previous boundaries", []string{"C4", "F4"}, []string{"E4", "B3"}, 1, true, true},
		{"leaving a unison upward", []string{"C4", "D4"}, []string{"C4", "E4"}, 1, true, false},
		{"leaving a unison downward", []string{"C4", "B3"}, []string{"C4", "A3"}, 1, false, true},
		{"touching previous heights", []string{"C4", "E4"}, []string{"E4", "C4"}, 0, false, false},
		{"already crossed repeat", []string{"F4", "F4"}, []string{"D4", "D4"}, 0, false, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := analyze(t, tt.low, tt.high)
			xs := a.MustFindOverlaps()
			require.Len(t, xs, tt.count)
			if tt.count > 0 {
				assert.Equal(t, tt.first, xs[0].FirstAbovePreviousSecond)
				assert.Equal(t, tt.second, xs[0].SecondBelowPreviousFirst)
				assertTransition(t, a.MustTransitions(pair)[0], xs[0].Transition)
			}
		})
	}
	// Only the outer pair has an overlap; it must be omitted from the full texture.
	lines := []*voice.Line{line(t, "C4", "F4"), line(t, "C6", "C6"), line(t, "E4", "G4")}
	full := voiceleading.MustAnalyze(lines, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh})
	assert.Empty(t, full.MustFindOverlaps())
	assert.Len(t, full.MustFindCrossings(), 1) // All pairs, including the middle/upper pair.
	outer := voiceleading.MustAnalyze([]*voice.Line{lines[0], lines[2]}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh})
	assert.Len(t, outer.MustFindOverlaps(), 1)
}

func TestCrossingsIncludeNonAdjacentParts(t *testing.T) {
	t.Parallel()
	a := voiceleading.MustAnalyze([]*voice.Line{line(t, "C6"), line(t, "E4"), line(t, "G4")}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh})
	xs := a.MustFindCrossings()
	require.Len(t, xs, 2)
	assert.Equal(t, voiceleading.Pair{First: 0, Second: 1}, xs[0].Pair)
	assert.Equal(t, voiceleading.Pair{First: 0, Second: 2}, xs[1].Pair)
	for _, x := range xs {
		assertPosition(t, pos(0, 1), x.Start)
		assertPosition(t, pos(1, 4), x.End)
		assert.Len(t, x.Verticals, 1)
	}
}

func TestBoundedGridAndSearchInvariants(t *testing.T) {
	t.Parallel()
	for _, count := range []int{2, 3, 4} {
		t.Run(fmt.Sprintf("%d parts", count), func(t *testing.T) {
			t.Parallel()
			voices := []*voice.Line{line(t, "C3", "D3", "E3"), line(t, "G3", "A3", "B3"), line(t, "C4", "D4", "E4"), line(t, "G4", "A4", "B4")}
			a := voiceleading.MustAnalyze(voices[:count], voiceleading.Options{System: system.Traditional()})
			pairs := a.MustPairs()
			require.Len(t, pairs, count*(count-1)/2)
			seen := map[voiceleading.Pair]bool{}
			for _, p := range pairs {
				assert.False(t, seen[p])
				seen[p] = true
				vs := a.MustVerticals(p)
				ts := a.MustTransitions(p)
				require.Len(t, vs, 3)
				require.Len(t, ts, 2)
				for i, v := range vs {
					assertPosition(t, pos(i, 4), v.Start)
					assertPosition(t, pos(i+1, 4), v.End)
					got, ok := a.MustVerticalAt(p, pos(2*i+1, 8))
					require.True(t, ok)
					assertVertical(t, v, got)
				}
				for i, tr := range ts {
					assert.Equal(t, voiceleading.MotionParallel, tr.Motion)
					assertTransition(t, tr, a.MustTransitionAt(p, pos(i, 4), pos(i+1, 4)))
				}
			}
			found := a.MustFindMotion(voiceleading.MotionParallel)
			assert.Len(t, found, count*(count-1))
			for i, tr := range found {
				assertTransition(t, a.MustTransitionAt(tr.From.Pair, tr.From.Start, tr.To.Start), tr)
				if i > 0 {
					prev := found[i-1]
					if prev.From.Pair == tr.From.Pair {
						assert.Less(t, prev.From.Start.Compare(tr.From.Start), 0)
					} else {
						assert.True(t, prev.From.Pair.First < tr.From.Pair.First || (prev.From.Pair.First == tr.From.Pair.First && prev.From.Pair.Second < tr.From.Pair.Second))
					}
				}
			}
		})
	}
}
