package tendency_test

import (
	"testing"

	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/problemsolving/tendency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoalSelection(t *testing.T) {
	t.Parallel()
	s := built(tendency.AMNeighborTargets)
	m := ionian()
	source := place(2, 0)
	for _, tc := range []struct {
		name                         string
		filter                       tendency.GoalFilter
		matched, uncertain, excluded int
	}{
		{"nil unrestricted", tendency.GoalFilter{}, 2, 0, 0},
		{"empty targets", tendency.GoalFilter{Targets: []tendency.Place{}}, 0, 0, 2},
		{"empty directions", tendency.GoalFilter{Directions: []tendency.Motion{}}, 0, 0, 2},
		{"ascending only", tendency.GoalFilter{Directions: []tendency.Motion{tendency.Ascending}}, 1, 0, 1},
		{"duplicates", tendency.GoalFilter{Targets: []tendency.Place{place(3, 0), place(3, 0)}}, 1, 0, 1},
		{"support", tendency.GoalFilter{Support: built(tendency.AMSupports)}, 2, 0, 0},
		{"no support task", tendency.GoalFilter{Support: built(tendency.AMMajorRoles)}, 0, 2, 0},
		{"impossible conjunction", tendency.GoalFilter{Targets: []tendency.Place{place(1, 0)}, Directions: []tendency.Motion{tendency.Ascending}}, 0, 0, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := s.MustSelectGoals(m, source, tc.filter)
			assert.Equal(t, s.MustTendency(m, source), r.Original)
			assert.Len(t, r.Matched, tc.matched)
			assert.Len(t, r.Uncertain, tc.uncertain)
			assert.Len(t, r.Excluded, tc.excluded)
			assert.NotNil(t, r.Matched)
			assert.NotNil(t, r.Uncertain)
			assert.NotNil(t, r.Excluded)
			assert.NotNil(t, r.Reasons)
			for _, g := range r.Matched {
				assert.False(t, g.Preferred && g.Target.Degree == 3, "removing preferred I must not promote III")
			}
			for _, e := range r.Excluded {
				checkReasons(t, e.Reasons)
			}
		})
	}
	text := built(tendency.AMTextTargets)
	for _, tc := range []struct {
		filter                       tendency.GoalFilter
		matched, uncertain, excluded int
	}{
		{tendency.GoalFilter{}, 2, 0, 0},
		{tendency.GoalFilter{Targets: []tendency.Place{place(1, 0)}}, 0, 1, 1},
		{tendency.GoalFilter{Targets: []tendency.Place{place(5, 0)}}, 0, 0, 2},
		{tendency.GoalFilter{Directions: []tendency.Motion{tendency.Ascending}}, 0, 2, 0},
		{tendency.GoalFilter{Targets: []tendency.Place{}, Directions: []tendency.Motion{tendency.Ascending}}, 0, 0, 2},
	} {
		r := text.MustSelectGoals(m, source, tc.filter)
		assert.Len(t, r.Matched, tc.matched)
		assert.Len(t, r.Uncertain, tc.uncertain)
		assert.Len(t, r.Excluded, tc.excluded)
	}
	d := definition(m)
	d.Supports = []tendency.SupportRule{{Degree: 1, Value: tendency.SupportValue{Kind: tendency.Unstable}, Bases: basis("unstable")}}
	support := tendency.MustNew(d)
	r := s.MustSelectGoals(m, source, tendency.GoalFilter{Support: support})
	assert.Empty(t, r.Matched)
	assert.Len(t, r.Uncertain, 1)
	assert.Len(t, r.Excluded, 1)
}

func TestPairAndIntervalSelection(t *testing.T) {
	t.Parallel()
	s := built(tendency.AMModalResolution)
	m := ionian()
	source := pair(2, 4)
	r := s.MustSelectModal(m, source, tendency.ModalFilter{})
	assert.Len(t, r.Matched, 3)
	assert.Len(t, r.Uncertain, 1)
	assert.Empty(t, r.Excluded)
	r = s.MustSelectModal(m, source, tendency.ModalFilter{FirstTargets: []tendency.Place{place(3, 0)}})
	require.Len(t, r.Matched, 1)
	assert.False(t, r.Matched[0].Preferred)
	assert.Len(t, r.Uncertain, 1)
	assert.Len(t, r.Excluded, 2)
	r = s.MustSelectModal(m, source, tendency.ModalFilter{Keep: tendency.KeepFirst})
	assert.Empty(t, r.Matched)
	assert.Empty(t, r.Uncertain)
	assert.Len(t, r.Excluded, 4)
	for _, filter := range []tendency.ModalFilter{{FirstTargets: []tendency.Place{}}, {SecondTargets: []tendency.Place{}}, {FirstDirections: []tendency.Motion{}}, {SecondDirections: []tendency.Motion{}}, {Keep: tendency.KeepBoth}} {
		r = s.MustSelectModal(m, source, filter)
		assert.Empty(t, r.Matched)
		assert.Empty(t, r.Uncertain)
		assert.Len(t, r.Excluded, 4)
	}
	r = s.MustSelectModal(m, source, tendency.ModalFilter{Support: built(tendency.AMSupports)})
	assert.Len(t, r.Matched, 3)
	assert.Len(t, r.Uncertain, 1, "positive filters cannot prove the gap candidate")
	intervals := built(tendency.AMIntervalResolution)
	for _, tc := range []struct {
		name   string
		filter tendency.IntervalFilter
		count  int
	}{
		{"all", tendency.IntervalFilter{}, 2}, {"keep first", tendency.IntervalFilter{Keep: tendency.KeepFirst}, 0}, {"keep second", tendency.IntervalFilter{Keep: tendency.KeepSecond}, 2}, {"keep both", tendency.IntervalFilter{Keep: tendency.KeepBoth}, 0}, {"empty first", tendency.IntervalFilter{FirstDirections: []tendency.Motion{}}, 0}, {"empty second", tendency.IntervalFilter{SecondDirections: []tendency.Motion{}}, 0}, {"descending first", tendency.IntervalFilter{FirstDirections: []tendency.Motion{tendency.Descending}}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := intervals.MustSelectInterval(iv("M2"), tc.filter)
			assert.Len(t, r.Matched, tc.count)
			assert.Empty(t, r.Uncertain)
			assert.Len(t, r.Excluded, 2-tc.count)
			assert.Equal(t, intervals.MustResolveInterval(iv("M2")), r.Original)
		})
	}
}

func TestComparisonPresenceAndOrigins(t *testing.T) {
	t.Parallel()
	m := aeolian()
	source := place(2, 0)
	neighbor := built(tendency.AMNeighborTargets)
	text := built(tendency.AMTextTargets)
	r := tendency.MustCompareTendencies(m, source, []*tendency.Scheme{text, neighbor}, tendency.ExactGoal)
	require.Len(t, r.Answers, 2)
	assert.Equal(t, text.Info(), r.Schemes[0])
	assert.Equal(t, neighbor.Info(), r.Schemes[1])
	require.Len(t, r.Items, 4)
	for _, item := range r.Items {
		assert.Len(t, item.Present, 1)
		if item.Value.Target.Kind == tendency.Immediate {
			assert.Equal(t, []int{1}, item.Present)
			assert.Equal(t, []int{0}, item.Unknown)
			assert.Empty(t, item.Absent)
		} else {
			assert.Equal(t, []int{0}, item.Present)
			assert.Equal(t, []int{1}, item.Unknown)
			assert.Empty(t, item.Absent)
		}
	}
	r = tendency.MustCompareTendencies(m, source, []*tendency.Scheme{text, neighbor}, tendency.DegreeGoal)
	require.Len(t, r.Items, 2)
	for _, item := range r.Items {
		assert.Equal(t, []int{0, 1}, item.Present)
		assert.Empty(t, item.Absent)
		assert.Empty(t, item.Unknown)
		assert.Equal(t, tendency.Target{Degree: item.Value.Target.Degree}, item.Value.Target)
		if item.Value.Target.Degree == 1 {
			assert.Equal(t, []int{0}, item.Preferred)
		} else {
			assert.Equal(t, []int{1}, item.Preferred)
		}
	}
	assert.Equal(t, tendency.MissingRule, r.Answers[0].DirectKnowledge)
	assert.Equal(t, tendency.Named, r.Answers[0].Goals[0].Target.Kind)
	// Known empty goals prove absence; unsupported tasks and missing rows do not.
	d := definition(m)
	d.Tendencies = []tendency.TendencyRule{{Degree: 2, Goals: []tendency.GoalRule{}, DirectComplete: true, Bases: basis("empty")}}
	empty := tendency.MustNew(d)
	d.ID = "user/missing"
	d.Tendencies = []tendency.TendencyRule{}
	missing := tendency.MustNew(d)
	r = tendency.MustCompareTendencies(m, source, []*tendency.Scheme{neighbor, empty, missing}, tendency.ExactGoal)
	for _, item := range r.Items {
		assert.Equal(t, []int{0}, item.Present)
		assert.Equal(t, []int{1}, item.Absent)
		assert.Equal(t, []int{2}, item.Unknown)
	}
	supports := tendency.MustCompareSupports(m, 1, []*tendency.Scheme{built(tendency.AMSupports), built(tendency.SeredaTonal)})
	require.Len(t, supports.Items, 1)
	assert.Equal(t, []int{0}, supports.Items[0].Present)
	assert.Equal(t, []int{1}, supports.Items[0].Unknown)
	roles := tendency.MustCompareRoles(m, 2, []*tendency.Scheme{built(tendency.AMMinorRoles)})
	require.Len(t, roles.Items, 1)
	assert.Equal(t, "", roles.Items[0].Value)
	classifications := tendency.MustCompareClassifications(iv("P4"), []*tendency.Scheme{built(tendency.AMConsonance), built(tendency.OMTHarmonic)})
	require.Len(t, classifications.Items, 2)
	for _, item := range classifications.Items {
		assert.Len(t, item.Present, 1)
		assert.Len(t, item.Absent, 1)
		assert.Empty(t, item.Unknown)
	}
	modal := tendency.MustCompareModalResolutions(ionian(), pair(2, 4), []*tendency.Scheme{built(tendency.AMModalResolution)})
	require.Len(t, modal.Items, 4)
	for _, item := range modal.Items {
		if item.Value == pair(3, 3) {
			assert.Empty(t, item.Present)
			assert.Equal(t, []int{0}, item.Unknown)
		} else {
			assert.Equal(t, []int{0}, item.Present)
		}
	}
	inter := tendency.MustCompareIntervalResolutions(iv("A4"), []*tendency.Scheme{built(tendency.AMIntervalResolution)})
	assert.Len(t, inter.Items, 3)
}

func TestEmptyComparisonsValidateInputs(t *testing.T) {
	t.Parallel()
	m := ionian()
	a := tendency.MustCompareSupports(m, 1, nil)
	assert.NotNil(t, a.Schemes)
	assert.NotNil(t, a.Answers)
	assert.NotNil(t, a.Items)
	assert.Empty(t, a.Items)
	b := tendency.MustCompareRoles(m, 1, []*tendency.Scheme{})
	assert.NotNil(t, b.Answers)
	assert.Empty(t, b.Answers)
	c := tendency.MustCompareTendencies(m, place(1, 0), nil, tendency.ExactGoal)
	assert.NotNil(t, c.Items)
	assert.Empty(t, c.Items)
	d := tendency.MustCompareClassifications(iv("P4"), nil)
	assert.NotNil(t, d.Items)
	assert.Empty(t, d.Items)
	e := tendency.MustCompareModalResolutions(m, pair(2, 4), nil)
	assert.NotNil(t, e.Items)
	assert.Empty(t, e.Items)
	f := tendency.MustCompareIntervalResolutions(iv("M2"), nil)
	assert.NotNil(t, f.Items)
	assert.Empty(t, f.Items)
	_, err := tendency.CompareSupports(mode.Mode{}, 1, nil)
	assert.ErrorIs(t, err, tendency.ErrInvalidMode)
	_, err = tendency.CompareTendencies(m, place(1, 0), nil, tendency.GoalMatch(200))
	assert.ErrorIs(t, err, tendency.ErrInvalidArgument)
	for _, schemes := range [][]*tendency.Scheme{{nil}, {new(tendency.Scheme)}, {built(tendency.AMSupports), built(tendency.AMSupports)}} {
		_, err := tendency.CompareSupports(m, 1, schemes)
		require.Error(t, err)
	}
}

func TestSelectionAndComparisonDeepCopies(t *testing.T) {
	t.Parallel()
	s := built(tendency.AMNeighborTargets)
	m := ionian()
	source := place(2, 0)
	r := s.MustSelectGoals(m, source, tendency.GoalFilter{})
	require.NotEmpty(t, r.Matched)
	require.NotEmpty(t, r.Matched[0].Reasons)
	before := r.Original.Goals[0].Reasons[0].Explanation
	r.Matched[0].Reasons[0].Explanation = "changed"
	assert.Equal(t, before, r.Original.Goals[0].Reasons[0].Explanation)
	r.Original.Goals[0].Reasons[0].Facts = append(r.Original.Goals[0].Reasons[0].Facts, tendency.Fact{Name: "test", Value: "changed"})
	assert.NotEqual(t, r.Original, s.MustTendency(m, source))
	c := tendency.MustCompareTendencies(m, source, []*tendency.Scheme{s, built(tendency.AMTextTargets)}, tendency.DegreeGoal)
	before = c.Answers[0].Name
	c.Schemes[0].Name = "changed"
	assert.Equal(t, before, c.Answers[0].Name)
	c.Answers[0].Goals[0].Reasons[0].Explanation = "changed"
	c.Items[0].Present[0] = 99
	assert.NotEqual(t, c, tendency.MustCompareTendencies(m, source, []*tendency.Scheme{s, built(tendency.AMTextTargets)}, tendency.DegreeGoal))
	p := built(tendency.AMModalResolution).MustSelectModal(m, pair(2, 4), tendency.ModalFilter{})
	before = p.Original.Allowed[0].Reasons[0].Explanation
	p.Matched[0].Reasons[0].Explanation = "changed"
	assert.Equal(t, before, p.Original.Allowed[0].Reasons[0].Explanation)
	i := built(tendency.AMIntervalResolution).MustSelectInterval(iv("M2"), tendency.IntervalFilter{})
	before = i.Original.Allowed[0].Reasons[0].Explanation
	i.Matched[0].Reasons[0].Explanation = "changed"
	assert.Equal(t, before, i.Original.Allowed[0].Reasons[0].Explanation)
}
