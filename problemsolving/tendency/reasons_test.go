package tendency_test

import (
	"cmp"
	"slices"
	"testing"

	"github.com/go-muse/musex/problemsolving/tendency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The public order is lexicographic over every Reason field, not only Rule.
func compareReasonEvidence(first, second tendency.Reason) int {
	return cmp.Or(
		cmp.Compare(first.Rule, second.Rule),
		cmp.Compare(first.Origin, second.Origin),
		cmp.Compare(first.Reference, second.Reference),
		cmp.Compare(first.Explanation, second.Explanation),
		slices.CompareFunc(first.Facts, second.Facts, func(a, b tendency.Fact) int {
			return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.Value, b.Value))
		}),
	)
}

func checkReasonOrder(t *testing.T, reasons []tendency.Reason) {
	t.Helper()
	expected := slices.Clone(reasons)
	slices.SortFunc(expected, compareReasonEvidence)
	assert.Equal(t, expected, reasons, "Reasons must follow Rule, Origin, Reference, Explanation, then ordered Facts")
}

func TestReasonsCanonicalOrder(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name          string
		query         func(*testing.T) []tendency.Reason
		requiredRules []string
	}{
		{
			name: "classification_missing_row",
			query: func(t *testing.T) []tendency.Reason {
				r, err := built(tendency.AMConsonance).Classify(iv("M10"))
				require.NoError(t, err)
				assert.Equal(t, tendency.MissingRule, r.Knowledge)
				return r.Reasons
			},
			requiredRules: []string{"am-consonance", "coverage.row"},
		},
		{
			name: "denied_neighbor_transition",
			query: func(t *testing.T) []tendency.Reason {
				r, err := built(tendency.AMNeighborTargets).CheckTransition(ionian(), place(7, 0), place(6, 0))
				require.NoError(t, err)
				assert.Equal(t, tendency.Denied, r.Decision)
				return r.Reasons
			},
			requiredRules: []string{"am.neighbor-targets", "filter.direction"},
		},
		{
			name: "aggregate_support_evidence",
			query: func(t *testing.T) []tendency.Reason {
				r, err := built(tendency.AMSupports).Supports(ionian())
				require.NoError(t, err)
				assert.Equal(t, tendency.Known, r.Knowledge)
				return r.Reasons
			},
			requiredRules: []string{"am.supports"},
		},
		{
			name: "same_user_rule_different_evidence",
			query: func(t *testing.T) []tendency.Reason {
				d := definition(ionian())
				d.Roles = []tendency.RoleRule{{Degree: 1, Role: "Anchor", Bases: []tendency.Basis{
					{Rule: "shared", Explanation: "Z explanation", References: []string{"a"}},
					{Rule: "shared", Explanation: "A explanation", References: []string{"z"}},
					{Rule: "shared", Explanation: "B explanation", References: []string{"a"}},
				}}}
				s, err := tendency.New(d)
				require.NoError(t, err)
				r, err := s.Role(d.Mode, 1)
				require.NoError(t, err)
				assert.Equal(t, tendency.Known, r.Knowledge)
				require.GreaterOrEqual(t, len(r.Reasons), 3, "distinct explanations must retain their evidence")
				return r.Reasons
			},
			requiredRules: []string{"user/shared"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reasons := tc.query(t)
			for _, rule := range tc.requiredRules {
				assert.Contains(t, reasonRules(reasons), rule)
			}
			checkReasons(t, reasons)
		})
	}
}
