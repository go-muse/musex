package tendency_test

import (
	"strings"
	"testing"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/problemsolving/tendency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func iv(text string) interval.Interval       { return interval.MustParse(text) }
func place(degree, cycle int) tendency.Place { return tendency.Place{Degree: degree, Cycle: cycle} }
func pair(first, second int) tendency.Pair {
	return tendency.Pair{First: place(first, 0), Second: place(second, 0)}
}
func built(id tendency.ID) *tendency.Scheme { return tendency.MustBuiltin(id) }
func ionian() mode.Mode                     { return mode.MustFromCode("ionian") }
func aeolian() mode.Mode                    { return mode.MustFromCode("aeolian") }
func degrees(text string) []interval.Interval {
	out := make([]interval.Interval, 0)
	for _, value := range strings.Fields(text) {
		out = append(out, iv(value))
	}
	return out
}
func closed(text, cycle string) mode.Mode {
	return mode.MustClosedFromDegreeIntervals(degrees(text), iv(cycle))
}
func open(text string) mode.Mode { return mode.MustOpenFromDegreeIntervals(degrees(text)) }
func basis(rule string) []tendency.Basis {
	return []tendency.Basis{{Rule: rule, Explanation: "An explicit user assertion", References: []string{"user documentation"}}}
}
func definition(m mode.Mode) tendency.Definition {
	return tendency.Definition{ID: "user/fixture", Name: "Fixture", Description: "Finite explicit test rules", Mode: m}
}
func goal(kind tendency.TargetKind, degree, cycle int, preferred bool) tendency.GoalRule {
	return tendency.GoalRule{Target: tendency.Target{Kind: kind, Degree: degree, Cycle: cycle, CycleKnown: true}, Preferred: preferred, Bases: basis("goal")}
}
func named(kind tendency.TargetKind, degree int) tendency.GoalRule {
	return tendency.GoalRule{Target: tendency.Target{Kind: kind, Degree: degree}, Bases: basis("named")}
}
func reasonRules(reasons []tendency.Reason) []string {
	out := make([]string, 0, len(reasons))
	for _, r := range reasons {
		out = append(out, r.Rule)
	}
	return out
}
func checkReasons(t *testing.T, reasons []tendency.Reason) {
	t.Helper()
	require.NotEmpty(t, reasons)
	checkReasonOrder(t, reasons)
	for index, r := range reasons {
		assert.NotEmpty(t, r.Rule)
		assert.NotEmpty(t, r.Reference)
		assert.NotEmpty(t, r.Explanation)
		assert.NotEqual(t, tendency.OriginUnset, r.Origin)
		assert.NotNil(t, r.Facts)
		for j := 1; j < len(r.Facts); j++ {
			assert.LessOrEqual(t, r.Facts[j-1].Name, r.Facts[j].Name)
		}
		for j := 0; j < index; j++ {
			assert.NotEqual(t, reasons[j], r, "duplicate reason")
		}
	}
}
func assertAnswer(t *testing.T, s *tendency.Scheme, answer tendency.Answer, knowledge tendency.Knowledge) {
	t.Helper()
	assert.Equal(t, s.Info().ID, answer.Scheme)
	assert.Equal(t, s.Info().Name, answer.Name)
	assert.Equal(t, knowledge, answer.Knowledge)
	checkReasons(t, answer.Reasons)
}
func targetPlaces(goals []tendency.Goal) []tendency.Place {
	out := make([]tendency.Place, 0, len(goals))
	for _, g := range goals {
		out = append(out, place(g.Target.Degree, g.Target.Cycle))
	}
	return out
}
func modalTargets(candidates []tendency.ModalCandidate) []tendency.Pair {
	out := make([]tendency.Pair, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, candidate.Target)
	}
	return out
}
