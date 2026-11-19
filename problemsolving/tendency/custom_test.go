package tendency_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/problemsolving/tendency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func finiteDefinition() tendency.Definition {
	d := definition(ionian())
	d.Supports = []tendency.SupportRule{{Degree: 1, Value: tendency.SupportValue{Kind: tendency.MainSupport, Ranked: true, Rank: 1}, Bases: basis("main")}, {Degree: 5, Value: tendency.SupportValue{Kind: tendency.RelativeSupport}, Bases: basis("relative")}, {Degree: 2, Value: tendency.SupportValue{Kind: tendency.Unstable}, Bases: basis("unstable")}}
	d.Roles = []tendency.RoleRule{{Degree: 1, Role: "Home", Bases: basis("role")}, {Degree: 2, Role: "", Bases: basis("unlabelled")}}
	d.Tendencies = []tendency.TendencyRule{{Degree: 2, Goals: []tendency.GoalRule{goal(tendency.Immediate, 3, 0, true)}, DirectComplete: true, Bases: basis("second")}, {Degree: 4, Goals: []tendency.GoalRule{}, DirectComplete: true, Bases: basis("empty")}}
	d.Modal = []tendency.ModalRule{{Source: pair(2, 4), Repeat: true, Options: []tendency.ModalOption{{Target: pair(2, 4), Preferred: true, Bases: basis("stay")}}, Bases: basis("pair")}}
	d.Intervals = []tendency.IntervalRule{{Source: iv("M2"), Options: []tendency.IntervalOption{{Motions: tendency.Motions{First: iv("P1"), Second: iv("P1")}, Bases: basis("stay")}}, Bases: basis("interval")}}
	return d
}

func TestFiniteUserRules(t *testing.T) {
	t.Parallel()
	s, err := tendency.New(finiteDefinition())
	require.NoError(t, err)
	assert.False(t, s.Info().Builtin)
	assert.Equal(t, tendency.SupportValue{Kind: tendency.MainSupport, Ranked: true, Rank: 1}, s.MustSupport(ionian(), 1).Value)
	assert.Equal(t, tendency.MissingRule, s.MustCompareSupport(ionian(), 1, 5).Knowledge)
	assert.Equal(t, tendency.MissingRule, s.MustSupport(ionian(), 7).Knowledge)
	assert.Equal(t, tendency.Partial, s.MustSupports(ionian()).Knowledge)
	assert.Equal(t, "Home", s.MustRole(ionian(), 1).Role)
	assert.Equal(t, "", s.MustRole(ionian(), 2).Role)
	assert.Equal(t, tendency.Known, s.MustRole(ionian(), 2).Knowledge)
	assert.Equal(t, tendency.Partial, s.MustRoles(ionian()).Knowledge)
	assert.Equal(t, tendency.Known, s.MustTendency(ionian(), place(4, 0)).Knowledge)
	assert.Empty(t, s.MustTendency(ionian(), place(4, 0)).Goals)
	assert.Equal(t, tendency.MissingRule, s.MustTendency(ionian(), place(6, 0)).Knowledge)
	assert.Equal(t, tendency.Undetermined, s.MustCheckTransition(ionian(), place(6, 0), place(1, 0)).Decision)
	assert.Equal(t, tendency.Denied, s.MustCheckTransition(ionian(), place(4, 0), place(1, 0)).Decision)
	r := s.MustTendency(ionian(), place(2, -3))
	require.Len(t, r.Goals, 1)
	assert.Equal(t, place(3, -3), targetPlaces(r.Goals)[0])
	assert.Equal(t, tendency.Ascending, r.Goals[0].Direction)
	for _, reason := range r.Goals[0].Reasons {
		assert.Equal(t, tendency.UserRule, reason.Origin)
		assert.Contains(t, reason.Rule, "user/")
	}
	assert.Equal(t, tendency.OutsideScope, s.MustSupport(aeolian(), 1).Knowledge)
	// Explicit user data may preserve a dissonance; AM is not silently applied.
	modal := s.MustResolveModal(ionian(), tendency.Pair{First: place(2, -2), Second: place(4, -2)})
	require.Len(t, modal.Allowed, 1)
	assert.Equal(t, modal.Source, modal.Allowed[0].Target)
	assert.Equal(t, iv("m3"), modal.Allowed[0].Final)
	assert.Equal(t, tendency.Known, s.MustResolveInterval(iv("M2")).Knowledge)
	require.Len(t, s.MustResolveInterval(iv("M2")).Allowed, 1)
	assert.Equal(t, tendency.MissingRule, s.MustResolveInterval(iv("m2")).Knowledge)
	assert.Equal(t, tendency.Denied, s.MustCheckInterval(iv("M2"), tendency.Motions{First: iv("m2"), Second: iv("M2")}).Decision)
	assert.Equal(t, tendency.UnsupportedTask, s.MustClassify(iv("M3")).Knowledge)
	// Empty sections are supported but uncovered, unlike missing sections.
	d := definition(ionian())
	d.Supports = []tendency.SupportRule{}
	empty := tendency.MustNew(d)
	assert.Equal(t, tendency.MissingRule, empty.MustSupport(ionian(), 1).Knowledge)
	assert.Equal(t, tendency.UnsupportedTask, empty.MustRole(ionian(), 1).Knowledge)
	analysis := empty.MustSupports(ionian())
	assert.Equal(t, tendency.MissingRule, analysis.Knowledge)
	require.Len(t, analysis.Groups, 3)
	for _, g := range analysis.Groups {
		assert.NotNil(t, g.Degrees)
		assert.Empty(t, g.Degrees)
	}
	// Pentatonic, single-degree, non-octave and fractional structures retain their own rules.
	for _, m := range []mode.Mode{mode.MustFromCode("pentatonic-major"), open("P1"), closed("P1 M3 P5", "P12"), mode.MustOpenFromDegreeIntervals([]interval.Interval{iv("P1"), interval.MustFromDeviation(2, alteration.MustFromFraction(-1, 2))})} {
		d := definition(m)
		d.Roles = []tendency.RoleRule{{Degree: 1, Role: "Anchor", Bases: basis("anchor")}}
		custom := tendency.MustNew(d)
		assert.Equal(t, "Anchor", custom.MustRole(m, 1).Role)
	}
}

func TestUnknownNamedAndUltimateRemainUnknown(t *testing.T) {
	t.Parallel()
	for _, kind := range []tendency.TargetKind{tendency.Named, tendency.Ultimate} {
		t.Run(map[tendency.TargetKind]string{tendency.Named: "named", tendency.Ultimate: "ultimate"}[kind], func(t *testing.T) {
			t.Parallel()
			for _, m := range []mode.Mode{ionian(), open("P1 M2 M3")} {
				d := definition(m)
				d.Tendencies = []tendency.TendencyRule{{Degree: 2, Goals: []tendency.GoalRule{named(kind, 1)}, Bases: basis("partial")}}
				s := tendency.MustNew(d)
				def, ok := s.Definition()
				require.True(t, ok)
				assert.Equal(t, d.Tendencies[0].Goals[0].Target, def.Tendencies[0].Goals[0].Target)
				cycle := 0
				if m.IsClosed() {
					cycle = -2
				}
				r := s.MustTendency(m, place(2, cycle))
				require.Len(t, r.Goals, 1)
				g := r.Goals[0]
				assert.Equal(t, tendency.Target{Kind: kind, Degree: 1}, g.Target)
				assert.Equal(t, tendency.MotionUnset, g.Direction)
				assert.False(t, g.Interval.IsValid())
				assert.Equal(t, tendency.MissingRule, r.DirectionKnowledge)
				assert.Equal(t, tendency.MissingRule, r.DirectKnowledge)
				assert.Equal(t, tendency.Known, r.GoalKnowledge)
				assert.Equal(t, tendency.Undetermined, s.MustCheckTransition(m, r.Source, place(1, 0)).Decision)
			}
		})
	}
	// A known Named location supplies geometry without establishing directness.
	d := definition(ionian())
	d.Tendencies = []tendency.TendencyRule{{Degree: 2, Goals: []tendency.GoalRule{goal(tendency.Named, 1, 0, false)}, Bases: basis("named")}}
	s := tendency.MustNew(d)
	r := s.MustTendency(ionian(), place(2, 0))
	assert.Equal(t, iv("M2").MustDown(), r.Goals[0].Interval)
	assert.Equal(t, tendency.Undetermined, s.MustCheckTransition(ionian(), place(2, 0), place(1, 0)).Decision)
	d.Tendencies[0].DirectComplete = true
	s = tendency.MustNew(d)
	assert.Equal(t, tendency.Denied, s.MustCheckTransition(ionian(), place(2, 0), place(1, 0)).Decision)
	// A mandatory direction alone can disprove a move, but cannot prove its goal.
	d.Tendencies = []tendency.TendencyRule{{Degree: 2, Directions: []tendency.Motion{tendency.Ascending}, Bases: basis("direction")}}
	s = tendency.MustNew(d)
	assert.Equal(t, tendency.Denied, s.MustCheckTransition(ionian(), place(2, 0), place(1, 0)).Decision)
	assert.Equal(t, tendency.Undetermined, s.MustCheckTransition(ionian(), place(2, 0), place(3, 0)).Decision)
}

func TestDefinitionCanonicalizationAndOwnership(t *testing.T) {
	t.Parallel()
	d := finiteDefinition()
	s := tendency.MustNew(d)
	d.Supports[0].Bases[0].References[0] = "changed"
	d.Tendencies[0].Goals[0].Target.Degree = 1
	require.NoError(t, d.Mode.RebuildFromCode("aeolian"))
	r := s.MustTendency(ionian(), place(2, 0))
	require.Len(t, r.Goals, 1)
	assert.Equal(t, 3, r.Goals[0].Target.Degree)
	out, ok := s.Definition()
	require.True(t, ok)
	out.Supports[0].Bases[0].References[0] = "mutated"
	out.Tendencies[0].Goals[0].Target.Degree = 1
	out.Modal[0].Options[0].Bases[0].Rule = "mutated"
	assert.True(t, s.Equal(tendency.MustNew(finiteDefinition())))
	assert.False(t, s.Equal(nil))
	other := finiteDefinition()
	other.Name = "Other"
	assert.False(t, s.Equal(tendency.MustNew(other)))
	other = finiteDefinition()
	other.Supports[0], other.Supports[2] = other.Supports[2], other.Supports[0]
	other.Roles = append(other.Roles, other.Roles[0])
	other.Supports = append(other.Supports, other.Supports[0])
	other.Tendencies = append(other.Tendencies, other.Tendencies[0])
	other.Intervals = append(other.Intervals, other.Intervals[0])
	other.Modal = append(other.Modal, other.Modal[0])
	assert.True(t, s.Equal(tendency.MustNew(other)))
	info := s.Info()
	info.Tasks[0] = tendency.TaskUnset
	assert.NotEqual(t, info, s.Info())
	r.Goals[0].Reasons[0].Explanation = "mutated"
	r.Reasons[0].Explanation = "mutated"
	assert.NotEqual(t, r, s.MustTendency(ionian(), place(2, 0)))
	// Duplicates merge evidence, with distinct origins and sorted unique references.
	d = definition(ionian())
	g := goal(tendency.Immediate, 3, 0, false)
	g.Bases = append(g.Bases, g.Bases[0])
	g.Bases[0].References = []string{"z", "a", "z"}
	d.Tendencies = []tendency.TendencyRule{{Degree: 2, DirectComplete: true, Goals: []tendency.GoalRule{g, g}, Bases: basis("row")}}
	s = tendency.MustNew(d)
	out, _ = s.Definition()
	require.Len(t, out.Tendencies[0].Goals, 1)
	assert.Equal(t, []string{"a", "z"}, out.Tendencies[0].Goals[0].Bases[0].References)
	// Repeating modal keys are independent of their initial common cycle.
	d = finiteDefinition()
	d.Modal[0].Source = tendency.Pair{First: place(2, 2), Second: place(4, 2)}
	d.Modal[0].Options[0].Target = d.Modal[0].Source
	assert.True(t, tendency.MustNew(finiteDefinition()).Equal(tendency.MustNew(d)))
}

func TestInvalidDefinitions(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		mutate func(*tendency.Definition)
		cause  error
	}{
		{"identity", func(d *tendency.Definition) { d.ID = "user/" }, nil},
		{"builtin identity", func(d *tendency.Definition) { d.ID = tendency.AMSupports }, nil},
		{"blank name", func(d *tendency.Definition) { d.Name = "  " }, nil},
		{"blank description", func(d *tendency.Definition) { d.Description = "\t" }, nil},
		{"no tasks", func(d *tendency.Definition) { *d = definition(ionian()) }, nil},
		{"mode missing", func(d *tendency.Definition) { d.Mode = mode.Mode{} }, nil},
		{"bad degree", func(d *tendency.Definition) { d.Supports[0].Degree = 8 }, tendency.ErrInvalidPlace},
		{"unset support", func(d *tendency.Definition) { d.Supports[0].Value.Kind = tendency.SupportUnset }, nil},
		{"invalid support enum", func(d *tendency.Definition) { d.Supports[0].Value.Kind = tendency.SupportKind(200) }, nil},
		{"invalid target enum", func(d *tendency.Definition) { d.Tendencies[0].Goals[0].Target.Kind = tendency.TargetKind(200) }, nil},
		{"invalid direction enum", func(d *tendency.Definition) { d.Tendencies[0].Directions = []tendency.Motion{tendency.Motion(200)} }, nil},
		{"unstable rank", func(d *tendency.Definition) { d.Supports[2].Value.Ranked = true; d.Supports[2].Value.Rank = 1 }, nil},
		{"unranked rank", func(d *tendency.Definition) { d.Supports[1].Value.Rank = 1 }, nil},
		{"nonpositive rank", func(d *tendency.Definition) { d.Supports[0].Value.Rank = 0 }, nil},
		{"role conflict", func(d *tendency.Definition) { r := d.Roles[0]; r.Role = "other"; d.Roles = append(d.Roles, r) }, nil},
		{"rank conflict", func(d *tendency.Definition) { r := d.Supports[0]; r.Value.Rank = 2; d.Supports = append(d.Supports, r) }, nil},
		{"missing basis", func(d *tendency.Definition) { d.Roles[0].Bases = nil }, nil},
		{"blank basis", func(d *tendency.Definition) { d.Roles[0].Bases[0].Rule = " " }, nil},
		{"blank explanation", func(d *tendency.Definition) { d.Roles[0].Bases[0].Explanation = " " }, nil},
		{"immediate unknown cycle", func(d *tendency.Definition) { d.Tendencies[0].Goals[0].Target.CycleKnown = false }, nil},
		{"unknown cycle nonzero", func(d *tendency.Definition) {
			d.Tendencies[0].Goals = []tendency.GoalRule{named(tendency.Named, 1)}
			d.Tendencies[0].Goals[0].Target.Cycle = 1
			d.Tendencies[0].DirectComplete = false
		}, nil},
		{"unknown cycle direction", func(d *tendency.Definition) {
			d.Tendencies[0].Goals = []tendency.GoalRule{named(tendency.Ultimate, 1)}
			d.Tendencies[0].Goals[0].Direction = tendency.Ascending
			d.Tendencies[0].DirectComplete = false
		}, nil},
		{"named number invalid", func(d *tendency.Definition) {
			d.Tendencies[0].Goals = []tendency.GoalRule{named(tendency.Named, 8)}
			d.Tendencies[0].DirectComplete = false
		}, tendency.ErrInvalidPlace},
		{"complete nil goals", func(d *tendency.Definition) { d.Tendencies[0].Goals = nil }, nil},
		{"wrong goal direction", func(d *tendency.Definition) { d.Tendencies[0].Goals[0].Direction = tendency.Descending }, nil},
		{"inconsistent directions", func(d *tendency.Definition) { d.Tendencies[0].Directions = []tendency.Motion{tendency.Descending} }, nil},
		{"goal preferred conflict", func(d *tendency.Definition) {
			g := d.Tendencies[0].Goals[0]
			g.Preferred = false
			d.Tendencies[0].Goals = append(d.Tendencies[0].Goals, g)
		}, nil},
		{"two preferences", func(d *tendency.Definition) {
			d.Tendencies[0].Goals = append(d.Tendencies[0].Goals, goal(tendency.Immediate, 1, 0, true))
		}, nil},
		{"row conflict", func(d *tendency.Definition) {
			r := d.Tendencies[0]
			r.DirectComplete = false
			d.Tendencies = append(d.Tendencies, r)
		}, nil},
		{"crossing modal", func(d *tendency.Definition) { d.Modal[0].Options[0].Target = pair(5, 1) }, nil},
		{"repeat overlap", func(d *tendency.Definition) { r := d.Modal[0]; r.Repeat = false; d.Modal = append(d.Modal, r) }, nil},
		{"open repetition", func(d *tendency.Definition) { d.Mode = open("P1 M2 M3 P4 P5 M6 M7") }, nil},
		{"modal coordinate overflow", func(d *tendency.Definition) {
			d.Mode = closed("P1", "A1")
			d.Supports = nil
			d.Roles = nil
			d.Tendencies = nil
			d.Modal = []tendency.ModalRule{{Source: tendency.Pair{First: place(1, math.MinInt), Second: place(1, 0)}, Repeat: true, Bases: basis("wide")}}
		}, tendency.ErrOutOfRange},
		{"interval unset", func(d *tendency.Definition) { d.Intervals[0].Source = interval.Interval{} }, tendency.ErrInvalidInterval},
		{"motion unset", func(d *tendency.Definition) { d.Intervals[0].Options[0].Motions.First = interval.Interval{} }, tendency.ErrInvalidInterval},
		{"interval crossing", func(d *tendency.Definition) { d.Intervals[0].Options[0].Motions.First = iv("P8") }, nil},
		{"interval option conflict", func(d *tendency.Definition) {
			o := d.Intervals[0].Options[0]
			o.AllowCrossing = true
			d.Intervals[0].Options = append(d.Intervals[0].Options, o)
		}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := finiteDefinition()
			tc.mutate(&d)
			s, err := tendency.New(d)
			require.Error(t, err)
			assert.Nil(t, s)
			assert.ErrorIs(t, err, tendency.ErrInvalidDefinition)
			if tc.cause != nil {
				assert.ErrorIs(t, err, tc.cause)
			}
		})
	}
}

func TestCustomCrossingAndExactPairs(t *testing.T) {
	t.Parallel()
	d := definition(ionian())
	d.Modal = []tendency.ModalRule{{Source: pair(5, 2), Options: []tendency.ModalOption{{Target: pair(5, 1), AllowCrossing: true, Bases: basis("cross")}}, Bases: basis("pair")}}
	s := tendency.MustNew(d)
	r := s.MustResolveModal(ionian(), pair(5, 2))
	require.Len(t, r.Allowed, 1)
	assert.Equal(t, pair(5, 1), r.Allowed[0].Target)
	assert.Equal(t, tendency.MissingRule, s.MustResolveModal(ionian(), tendency.Pair{First: place(5, 1), Second: place(2, 1)}).Knowledge)
	d = definition(mode.Mode{})
	d.Intervals = []tendency.IntervalRule{{Source: iv("M2"), Options: []tendency.IntervalOption{{Motions: tendency.Motions{First: iv("P8"), Second: iv("P1")}, AllowCrossing: true, Bases: basis("cross")}}, Bases: basis("row")}}
	s = tendency.MustNew(d)
	r2 := s.MustResolveInterval(iv("M2"))
	require.Len(t, r2.Allowed, 1)
	assert.Equal(t, interval.Down, r2.Allowed[0].Final.Direction())
	assert.Equal(t, tendency.MissingRule, s.MustResolveInterval(iv("M9")).Knowledge)
	// Equal finals do not collapse different concrete motions (TH-K20).
	d.Intervals[0].Options = []tendency.IntervalOption{{Motions: tendency.Motions{First: iv("P1"), Second: iv("P1")}, Bases: basis("stay")}, {Motions: tendency.Motions{First: iv("M2"), Second: iv("M2")}, Bases: basis("parallel")}}
	s = tendency.MustNew(d)
	r2 = s.MustResolveInterval(iv("M2"))
	require.Len(t, r2.Allowed, 2)
	assert.Equal(t, r2.Allowed[0].Final, r2.Allowed[1].Final)
	assert.NotEqual(t, r2.Allowed[0].Motions, r2.Allowed[1].Motions)
}
