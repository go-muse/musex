package tendency_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/problemsolving/tendency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNeighborTargets(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		m         mode.Mode
		number    int
		targets   []tendency.Place
		preferred int
	}{
		{"major second", ionian(), 2, []tendency.Place{place(1, 0), place(3, 0)}, 1},
		{"minor second K3", aeolian(), 2, []tendency.Place{place(1, 0), place(3, 0)}, 3},
		{"minor fourth K4", aeolian(), 4, []tendency.Place{place(3, 0), place(5, 0)}, 0},
		{"seventh crosses cycle", ionian(), 7, []tendency.Place{place(1, 1)}, 1},
		{"support is empty", ionian(), 1, []tendency.Place{}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := built(tendency.AMNeighborTargets)
			r := s.MustTendency(tc.m, place(tc.number, 0))
			assertAnswer(t, s, r.Answer, tendency.Known)
			assert.Equal(t, tendency.Known, r.DirectKnowledge)
			assert.Equal(t, tendency.Known, r.DirectionKnowledge)
			assert.Equal(t, tendency.Known, r.GoalKnowledge)
			assert.Equal(t, tc.targets, targetPlaces(r.Goals))
			assert.NotNil(t, r.Goals)
			assert.NotNil(t, r.Directions)
			for _, g := range r.Goals {
				assert.Equal(t, tendency.Immediate, g.Target.Kind)
				assert.True(t, g.Target.CycleKnown)
				assert.Equal(t, g.Target.Degree == tc.preferred, g.Preferred)
				assert.Equal(t, g.Interval, tendency.MustBetween(tc.m, r.Source, place(g.Target.Degree, g.Target.Cycle)).Interval)
				checkReasons(t, g.Reasons)
				assert.Equal(t, tendency.Allowed, s.MustCheckTransition(tc.m, r.Source, place(g.Target.Degree, g.Target.Cycle)).Decision)
			}
		})
	}
	s := built(tendency.AMNeighborTargets)
	r := s.MustTendencies(ionian(), -1)
	require.Len(t, r.Rows, 7)
	for i, row := range r.Rows {
		assert.Equal(t, place(i+1, -1), row.Source)
		assert.Equal(t, row, s.MustTendency(ionian(), row.Source))
	}
	assert.Equal(t, place(1, 0), targetPlaces(r.Rows[6].Goals)[0])
}

func TestNeighborTargetsAllSixStructures(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		structure string
		preferred [4]int
	}{
		{"P1 M2 M3 P4 P5 M6 M7", [4]int{1, 3, 5, 1}},
		{"P1 M2 M3 P4 P5 m6 M7", [4]int{1, 3, 5, 1}},
		{"P1 M2 M3 P4 P5 m6 m7", [4]int{1, 3, 5, 1}},
		{"P1 M2 m3 P4 P5 m6 m7", [4]int{3, 0, 5, 1}},
		{"P1 M2 m3 P4 P5 m6 M7", [4]int{3, 0, 5, 1}},
		{"P1 M2 m3 P4 P5 M6 M7", [4]int{3, 0, 5, 1}},
	} {
		t.Run(tc.structure, func(t *testing.T) {
			t.Parallel()
			m := closed(tc.structure, "P8")
			s := built(tendency.AMNeighborTargets)
			want := [][]tendency.Place{{place(1, 0), place(3, 0)}, {place(3, 0), place(5, 0)}, {place(5, 0)}, {place(1, 1)}}
			for i, number := range []int{2, 4, 6, 7} {
				r := s.MustTendency(m, place(number, 0))
				assert.Equal(t, want[i], targetPlaces(r.Goals))
				for _, g := range r.Goals {
					assert.Equal(t, tc.preferred[i] == g.Target.Degree, g.Preferred)
				}
			}
			for _, number := range []int{1, 3, 5} {
				assert.Empty(t, s.MustTendency(m, place(number, 0)).Goals)
			}
		})
	}
}

func TestTextTargetsAllTwelveRows(t *testing.T) {
	t.Parallel()
	// Expectations transcribe the twelve source assertions, not the neighbor algorithm.
	cases := []struct {
		structure string
		number    int
		targets   []int
		preferred int
		ultimate  bool
	}{
		{"P1 M2 M3 P4 P5 M6 M7", 2, []int{1, 3}, 1, false},
		{"P1 M2 M3 P4 P5 M6 M7", 4, []int{3, 5}, 3, false},
		{"P1 M2 M3 P4 P5 M6 M7", 6, []int{5}, 5, false},
		{"P1 M2 M3 P4 P5 M6 M7", 7, []int{1}, 1, false},
		{"P1 M2 M3 P4 P5 m6 M7", 6, []int{5}, 5, false},
		{"P1 M2 M3 P4 P5 m6 m7", 7, []int{1}, 1, false},
		{"P1 M2 m3 P4 P5 m6 m7", 2, []int{1, 3}, 1, false},
		{"P1 M2 m3 P4 P5 m6 m7", 4, []int{3, 5}, 0, false},
		{"P1 M2 m3 P4 P5 m6 m7", 6, []int{5}, 5, false},
		{"P1 M2 m3 P4 P5 m6 m7", 7, []int{1}, 1, false},
		{"P1 M2 m3 P4 P5 m6 M7", 7, []int{1}, 1, false},
		{"P1 M2 m3 P4 P5 M6 M7", 6, []int{1, 5}, 1, true},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			s := built(tendency.AMTextTargets)
			m := closed(tc.structure, "P8")
			r := s.MustTendency(m, place(tc.number, -2))
			assertAnswer(t, s, r.Answer, tendency.Partial)
			assert.Equal(t, tendency.Known, r.GoalKnowledge)
			assert.Equal(t, tendency.MissingRule, r.DirectKnowledge)
			assert.Equal(t, tendency.MissingRule, r.DirectionKnowledge)
			assert.Empty(t, r.Directions)
			require.Len(t, r.Goals, len(tc.targets))
			for j, g := range r.Goals {
				assert.Equal(t, tc.targets[j], g.Target.Degree)
				assert.False(t, g.Target.CycleKnown)
				assert.Zero(t, g.Target.Cycle)
				assert.Equal(t, tendency.MotionUnset, g.Direction)
				assert.False(t, g.Interval.IsValid())
				assert.Equal(t, g.Target.Degree == tc.preferred, g.Preferred)
				kind := tendency.Named
				if tc.ultimate && g.Target.Degree == 1 {
					kind = tendency.Ultimate
					found := false
					for _, reason := range g.Reasons {
						for _, fact := range reason.Facts {
							if fact.Name == "via-degree" && fact.Value == "7" {
								found = true
							}
						}
					}
					assert.True(t, found, "ultimate I must disclose VII")
				}
				assert.Equal(t, kind, g.Target.Kind)
				checkReasons(t, g.Reasons)
			}
			for number := 1; number <= 7; number++ {
				assert.Equal(t, tendency.Undetermined, s.MustCheckTransition(m, r.Source, place(number, 0)).Decision)
			}
		})
	}
	s := built(tendency.AMTextTargets)
	for _, tc := range []struct {
		m mode.Mode
		n int
	}{{ionian(), 1}, {closed("P1 M2 M3 P4 P5 m6 M7", "P8"), 2}, {closed("P1 M2 m3 P4 P5 M6 M7", "P8"), 7}} {
		r := s.MustTendency(tc.m, place(tc.n, 0))
		assert.Equal(t, tendency.MissingRule, r.Knowledge)
		assert.Empty(t, r.Goals)
	}
	r := s.MustTendencies(ionian(), 0)
	assert.Equal(t, tendency.Partial, r.Knowledge)
	require.Len(t, r.Rows, 7)
	assert.Equal(t, tendency.MissingRule, r.Rows[0].Knowledge)
	assert.Equal(t, tendency.Partial, r.Rows[1].Knowledge)
}

func TestSeredaDirectionsAndContext(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		id             tendency.ID
		m              mode.Mode
		source, target tendency.Place
		motion         tendency.Motion
	}{
		{tendency.SeredaTonal, aeolian(), place(7, 0), place(6, 0), tendency.Descending},
		{tendency.SeredaModal, aeolian(), place(7, 0), place(1, 1), tendency.Ascending},
		{tendency.SeredaTonal, ionian(), place(2, -1), place(3, -1), tendency.Ascending},
		{tendency.SeredaModal, ionian(), place(2, -1), place(1, -1), tendency.Descending},
	} {
		s := built(tc.id)
		r := s.MustTendency(tc.m, tc.source)
		assertAnswer(t, s, r.Answer, tendency.Known)
		require.Len(t, r.Goals, 1)
		assert.Equal(t, []tendency.Motion{tc.motion}, r.Directions)
		assert.Equal(t, []tendency.Place{tc.target}, targetPlaces(r.Goals))
		rules := reasonRules(r.Goals[0].Reasons)
		assert.Contains(t, rules, "product.direction-to-neighbor")
		assert.Contains(t, rules, string(tc.id[:6])+"."+string(tc.id[7:])+"-direction")
	}
	for _, id := range []tendency.ID{tendency.SeredaTonal, tendency.SeredaModal} {
		for _, number := range []int{1, 4, 5} {
			r := built(id).MustTendency(ionian(), place(number, 0))
			assert.Equal(t, tendency.MissingContext, r.Knowledge)
			assert.Contains(t, reasonRules(r.Reasons), "context.neutral-pair")
			assert.Empty(t, r.Goals)
		}
	}
	m := open("P1 M2")
	s := built(tendency.SeredaTonal)
	r := s.MustTendency(m, place(2, 0))
	assertAnswer(t, s, r.Answer, tendency.Known)
	assert.Equal(t, []tendency.Motion{tendency.Ascending}, r.Directions)
	assert.Empty(t, r.Goals)
	assert.Equal(t, tendency.Known, r.DirectKnowledge)
	assert.Contains(t, reasonRules(r.Reasons), "boundary.no-neighbor")
	assert.Equal(t, tendency.Denied, s.MustCheckTransition(m, place(2, 0), place(1, 0)).Decision)
	fractional := interval.MustFromDeviation(2, alteration.MustFromFraction(-1, 2))
	m = mode.MustOpenFromDegreeIntervals([]interval.Interval{iv("P1"), fractional})
	r = s.MustTendency(m, place(2, 0))
	assert.Equal(t, tendency.OutsideScope, r.Knowledge)
	assert.Contains(t, reasonRules(r.Reasons), "modal.unsupported-quality")
	assert.NotContains(t, reasonRules(r.Reasons), "context.neutral-pair")
}
