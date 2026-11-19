package tendency_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/problemsolving/tendency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserPentatonicSupportsAndFractionalGoal(t *testing.T) {
	t.Parallel()
	m := mode.MustFromCode("pentatonic-major")
	d := definition(m)
	d.Supports = []tendency.SupportRule{{Degree: 1, Value: tendency.SupportValue{Kind: tendency.MainSupport, Ranked: true, Rank: 4}, Bases: basis("first")}, {Degree: 5, Value: tendency.SupportValue{Kind: tendency.RelativeSupport, Ranked: true, Rank: 4}, Bases: basis("fifth")}}
	d.Tendencies = []tendency.TendencyRule{{Degree: 2, Goals: []tendency.GoalRule{goal(tendency.Immediate, 3, 0, false)}, DirectComplete: true, Bases: basis("second")}}
	s := tendency.MustNew(d)
	assert.Equal(t, tendency.MainSupport, s.MustSupport(m, 1).Value.Kind)
	assert.Equal(t, tendency.RelativeSupport, s.MustSupport(m, 5).Value.Kind)
	r := s.MustCompareSupport(m, 1, 5)
	assert.Equal(t, tendency.Known, r.Knowledge)
	assert.Zero(t, r.Order)
	assert.Equal(t, []tendency.RankGroup{{Rank: 4, Degrees: []int{1, 5}}}, s.MustSupports(m).Order)
	assert.Equal(t, tendency.Allowed, s.MustCheckTransition(m, place(2, 0), place(3, 0)).Decision)
	fractional := interval.MustFromDeviation(2, alteration.MustFromFraction(-1, 2))
	m = mode.MustOpenFromDegreeIntervals([]interval.Interval{iv("P1"), fractional})
	d = definition(m)
	d.Tendencies = []tendency.TendencyRule{{Degree: 1, Goals: []tendency.GoalRule{goal(tendency.Immediate, 2, 0, false)}, DirectComplete: true, Bases: basis("fractional")}}
	s = tendency.MustNew(d)
	answer := s.MustTendency(m, place(1, 0))
	assert.Equal(t, tendency.Known, answer.Knowledge)
	require.Len(t, answer.Goals, 1)
	assert.Equal(t, fractional, answer.Goals[0].Interval)
}

func TestTextDataCanBeCopiedWithoutInventingContext(t *testing.T) {
	t.Parallel()
	m := closed("P1 M2 m3 P4 P5 M6 M7", "P8")
	original := built(tendency.AMTextTargets).MustTendency(m, place(6, 0))
	d := definition(m)
	row := tendency.TendencyRule{Degree: 6, Goals: []tendency.GoalRule{}, Bases: basis("copied-text")}
	for _, g := range original.Goals {
		row.Goals = append(row.Goals, tendency.GoalRule{Target: g.Target, Preferred: g.Preferred, Bases: basis("copied-goal")})
	}
	d.Tendencies = []tendency.TendencyRule{row}
	s := tendency.MustNew(d)
	stored, ok := s.Definition()
	require.True(t, ok)
	for _, g := range stored.Tendencies[0].Goals {
		assert.False(t, g.Target.CycleKnown)
		assert.Zero(t, g.Target.Cycle)
		assert.Equal(t, tendency.MotionUnset, g.Direction)
	}
	r := s.MustTendency(m, place(6, 2))
	assert.Equal(t, tendency.Partial, r.Knowledge)
	assert.Equal(t, tendency.MissingRule, r.DirectionKnowledge)
	assert.Equal(t, tendency.MissingRule, r.DirectKnowledge)
	require.Len(t, r.Goals, 2)
	assert.Equal(t, tendency.Ultimate, r.Goals[0].Target.Kind)
	assert.Equal(t, tendency.Named, r.Goals[1].Target.Kind)
	for _, g := range r.Goals {
		assert.False(t, g.Target.CycleKnown)
		assert.False(t, g.Interval.IsValid())
		assert.Equal(t, tendency.MotionUnset, g.Direction)
		for _, reason := range g.Reasons {
			assert.Equal(t, tendency.UserRule, reason.Origin)
		}
	}
	assert.Equal(t, tendency.Undetermined, s.MustCheckTransition(m, place(6, 2), place(1, 3)).Decision)
}

func TestUserEmptyPairsAndRelativeCircleOverflow(t *testing.T) {
	t.Parallel()
	d := definition(ionian())
	d.Modal = []tendency.ModalRule{{Source: pair(2, 4), Options: []tendency.ModalOption{}, Bases: basis("empty")}}
	d.Intervals = []tendency.IntervalRule{{Source: iv("m2"), Bases: basis("empty")}}
	s := tendency.MustNew(d)
	r := s.MustResolveModal(ionian(), pair(2, 4))
	assert.Equal(t, tendency.Known, r.Knowledge)
	assert.NotNil(t, r.Allowed)
	assert.Empty(t, r.Allowed)
	assert.Equal(t, tendency.Denied, s.MustCheckModal(ionian(), pair(2, 4), pair(1, 3)).Decision)
	assert.Equal(t, tendency.MissingRule, s.MustResolveModal(ionian(), pair(2, 6)).Knowledge)
	r2 := s.MustResolveInterval(iv("m2"))
	assert.Equal(t, tendency.Known, r2.Knowledge)
	assert.Empty(t, r2.Allowed)
	assert.Equal(t, tendency.Denied, s.MustCheckInterval(iv("m2"), tendency.Motions{First: iv("P1"), Second: iv("P1")}).Decision)
	// Unknown circles are not shifted; a known relative circle is checked before overflow.
	d = definition(closed("P1", "A1"))
	d.Tendencies = []tendency.TendencyRule{{Degree: 1, Goals: []tendency.GoalRule{goal(tendency.Immediate, 1, 1, false)}, DirectComplete: true, Bases: basis("next")}}
	s = tendency.MustNew(d)
	_, err := s.Tendency(d.Mode, place(1, math.MaxInt))
	assert.ErrorIs(t, err, tendency.ErrOutOfRange)
	// An open target may name a degree without a circle, but an exact continuation is invalid.
	d = definition(open("P1 M2"))
	d.Tendencies = []tendency.TendencyRule{{Degree: 1, Goals: []tendency.GoalRule{goal(tendency.Immediate, 2, 1, false)}, DirectComplete: true, Bases: basis("outside")}}
	_, err = tendency.New(d)
	assert.ErrorIs(t, err, tendency.ErrInvalidDefinition)
	assert.ErrorIs(t, err, tendency.ErrInvalidPlace)
}

func TestComparisonUsesConcreteMotionsNotFinalInterval(t *testing.T) {
	t.Parallel()
	d := definition(mode.Mode{})
	stay := tendency.Motions{First: iv("P1"), Second: iv("P1")}
	parallel := tendency.Motions{First: iv("M2"), Second: iv("M2")}
	d.Intervals = []tendency.IntervalRule{{Source: iv("M3"), Options: []tendency.IntervalOption{{Motions: stay, Bases: basis("stay")}}, Bases: basis("row")}}
	a := tendency.MustNew(d)
	d.ID = "user/parallel"
	d.Intervals[0].Options[0].Motions = parallel
	b := tendency.MustNew(d)
	r := tendency.MustCompareIntervalResolutions(iv("M3"), []*tendency.Scheme{b, a})
	require.Len(t, r.Items, 2)
	assert.Equal(t, stay, r.Items[0].Value)
	assert.Equal(t, []int{1}, r.Items[0].Present)
	assert.Equal(t, []int{0}, r.Items[0].Absent)
	assert.Equal(t, parallel, r.Items[1].Value)
	assert.Equal(t, []int{0}, r.Items[1].Present)
	d = definition(ionian())
	d.Roles = []tendency.RoleRule{{Degree: 4, Role: "s", Bases: basis("lowercase")}}
	custom := tendency.MustNew(d)
	roles := tendency.MustCompareRoles(ionian(), 4, []*tendency.Scheme{custom, built(tendency.AMMajorRoles)})
	require.Len(t, roles.Items, 2)
	assert.Equal(t, "S", roles.Items[0].Value)
	assert.Equal(t, "s", roles.Items[1].Value)
	assert.Equal(t, []int{1}, roles.Items[0].Present)
	assert.Equal(t, []int{0}, roles.Items[0].Absent)
}

func ExampleScheme_MustTendency() {
	m := mode.MustFromCode("aeolian")
	r := tendency.MustBuiltin(tendency.SeredaModal).MustTendency(m, tendency.Place{Degree: 7})
	fmt.Println(r.Goals[0].Target.Degree, r.Goals[0].Target.Cycle, r.Goals[0].Target.CycleKnown)
	// Output: 1 1 true
}

func ExampleScheme_MustCheckTransition() {
	m := mode.MustFromCode("aeolian")
	r := tendency.MustBuiltin(tendency.AMTextTargets).MustCheckTransition(m, tendency.Place{Degree: 7}, tendency.Place{Degree: 1, Cycle: 1})
	fmt.Println(r.Decision == tendency.Undetermined)
	// Output: true
}
