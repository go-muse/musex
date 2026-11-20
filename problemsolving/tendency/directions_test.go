package tendency_test

import (
	"testing"

	"github.com/go-muse/musex/problemsolving/tendency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExplicitEmptyDirectionsRejectTransitions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name               string
		directions         []tendency.Motion
		directionKnowledge tendency.Knowledge
		knowledge          tendency.Knowledge
		decision           tendency.Decision
	}{
		{"explicit_empty", []tendency.Motion{}, tendency.Known, tendency.Partial, tendency.Denied},
		{"nil_unspecified", nil, tendency.MissingRule, tendency.MissingRule, tendency.Undetermined},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := ionian()
			d := definition(m)
			d.Tendencies = []tendency.TendencyRule{{
				Degree:         2,
				Directions:     tc.directions,
				Goals:          nil,
				DirectComplete: false,
				Bases:          basis("directions"),
			}}
			s, err := tendency.New(d)
			require.NoError(t, err)
			source := place(2, 0)
			r, err := s.Tendency(m, source)
			require.NoError(t, err)
			assertAnswer(t, s, r.Answer, tc.knowledge)
			assert.Equal(t, tc.directionKnowledge, r.DirectionKnowledge)
			assert.Equal(t, tendency.MissingRule, r.GoalKnowledge)
			assert.Equal(t, tendency.MissingRule, r.DirectKnowledge)
			assert.Empty(t, r.Directions)
			assert.Empty(t, r.Goals)
			for _, move := range []struct {
				name   string
				target tendency.Place
			}{
				{"ascending_second_to_third", place(3, 0)},
				{"stationary_second_to_second", place(2, 0)},
			} {
				t.Run(move.name, func(t *testing.T) {
					t.Parallel()
					check, err := s.CheckTransition(m, source, move.target)
					require.NoError(t, err)
					assert.Equal(t, tc.decision, check.Decision)
					if tc.directions != nil {
						assert.Equal(t, tendency.Partial, check.Knowledge, "unknown goals must not erase a proven direction violation")
					}
					checkReasons(t, check.Reasons)
				})
			}
		})
	}
}
