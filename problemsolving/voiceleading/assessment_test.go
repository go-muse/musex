package voiceleading_test

import (
	"testing"

	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOMTFirstSpecies(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		low, high []string
		reasons   []string
	}{
		{"contrary fifth to compound fifth", []string{"C4", "G3"}, []string{"G4", "D5"}, []string{"consecutive-perfect-fifths"}},
		{"stationary fifths", []string{"C4", "C4"}, []string{"G4", "G4"}, []string{"consecutive-perfect-fifths"}},
		{"octave family includes unison and double octave", []string{"C4", "D4", "E4"}, []string{"C4", "D5", "E6"}, []string{"consecutive-perfect-octaves", "consecutive-perfect-octaves"}},
		{"fifth to third permitted by this rule", []string{"C4", "D4"}, []string{"G4", "F4"}, nil},
		{"one vertical", []string{"C4"}, []string{"G4"}, nil},
		{"empty", nil, nil, nil},
		{"rests only", []string{"-", "-"}, []string{"-", "-"}, nil},
		{"rest splits perfects", []string{"C4", "-", "D4"}, []string{"G4", "-", "A4"}, nil},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := analyze(t, tt.low, tt.high)
			assessment, err := a.CheckOMTFirstSpecies(pair)
			require.NoError(t, err)
			require.NotNil(t, assessment)
			assert.Equal(t, voiceleading.OMTFirstSpeciesRule, assessment.Rule)
			assert.Equal(t, "omt-first-species-consecutive-perfects", assessment.Rule)
			assert.Equal(t, "two-part first-species counterpoint, note against note; consecutive perfect families only", assessment.Scope)
			assert.True(t, assessment.Applicable)
			assert.NotNil(t, assessment.Issues)
			assert.Empty(t, assessment.Issues)
			assert.NotNil(t, assessment.Violations)
			require.Len(t, assessment.Violations, len(tt.reasons))
			consecutive := a.MustFindConsecutive(voiceleading.PerfectBoth)
			require.Len(t, consecutive, len(tt.reasons))
			for i, v := range assessment.Violations {
				assert.Equal(t, tt.reasons[i], v.Reason)
				assertTransition(t, consecutive[i], v.Transition)
			}
		})
	}
}

func TestOMTScopeUsesOriginalNoteBoundaries(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		low, high   []voice.Element
		issueStarts []int
	}{
		{"new attack over held note", []voice.Element{sound(t, "C", 4, 1, 2, nil)}, []voice.Element{sound(t, "G", 4, 1, 4, nil), sound(t, "A", 4, 1, 4, nil)}, []int{0, 1}},
		{"entry after rest over held note", []voice.Element{sound(t, "C", 4, 1, 2, nil)}, []voice.Element{rest(1, 4), sound(t, "G", 4, 1, 4, nil)}, []int{1}},
		{"clipped unequal original ends", []voice.Element{sound(t, "C", 4, 1, 2, nil)}, []voice.Element{sound(t, "G", 4, 1, 4, nil)}, []int{0}},
		{"one note ends at rest boundary", []voice.Element{sound(t, "C", 4, 1, 2, nil)}, []voice.Element{sound(t, "G", 4, 1, 4, nil), rest(1, 4)}, []int{0}},
		{"no partial violations", []voice.Element{sound(t, "C", 4, 1, 4, nil), sound(t, "D", 4, 1, 4, nil), sound(t, "E", 4, 1, 2, nil)}, []voice.Element{sound(t, "G", 4, 1, 4, nil), sound(t, "A", 4, 1, 4, nil), sound(t, "B", 4, 1, 4, nil), sound(t, "C", 5, 1, 4, nil)}, []int{2, 3}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := voiceleading.MustAnalyze([]*voice.Line{voice.MustNew(tt.low...), voice.MustNew(tt.high...)}, voiceleading.Options{System: system.Traditional()})
			result := a.MustCheckOMTFirstSpecies(pair)
			assert.False(t, result.Applicable)
			require.Len(t, result.Issues, len(tt.issueStarts))
			assert.NotNil(t, result.Violations)
			assert.Empty(t, result.Violations)
			for i, issue := range result.Issues {
				assert.Equal(t, "asynchronous-note-boundaries", issue.Reason)
				assertPosition(t, pos(tt.issueStarts[i], 4), issue.Vertical.Start)
				assert.True(t, issue.Vertical.Sounding)
			}
			assert.NotEmpty(t, a.MustVerticals(pair))
		})
	}
	// Variable but matching note lengths remain within scope; unrelated third-part rhythm is irrelevant.
	low := voice.MustNew(sound(t, "C", 4, 1, 3, nil), sound(t, "D", 4, 1, 6, nil))
	high := voice.MustNew(sound(t, "G", 4, 1, 3, nil), sound(t, "A", 4, 1, 6, nil))
	a := voiceleading.MustAnalyze([]*voice.Line{low, high, voice.MustNew(rest(1, 7))}, voiceleading.Options{System: system.Traditional()})
	result := a.MustCheckOMTFirstSpecies(pair)
	assert.True(t, result.Applicable)
	assert.Len(t, result.Violations, 1)
}
