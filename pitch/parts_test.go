package pitch_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
)

func TestPredicates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		p         pitch.Pitch
		isNatural bool
		isInteger bool
		direction spelling.Direction
	}{
		{
			name:      "C_ASPN_4_natural_integer",
			p:         pa(spelling.C(), 4),
			isNatural: true,
			isInteger: true,
			direction: spelling.DirectionNatural,
		},
		{
			name:      "CSharp_ASPN_4_raised_integer",
			p:         pa(spelling.CSharp(), 4),
			isNatural: false,
			isInteger: true,
			direction: spelling.DirectionRaised,
		},
		{
			name:      "DFlat_ASPN_4_lowered_integer",
			p:         pa(spelling.DFlat(), 4),
			isNatural: false,
			isInteger: true,
			direction: spelling.DirectionLowered,
		},
		{
			name:      "C_plus_half_ASPN_4_raised_fractional",
			p:         pa(cPlusHalf(), 4),
			isNatural: false,
			isInteger: false,
			direction: spelling.DirectionRaised,
		},
		{
			name:      "B_minus_seven_halves_ASPN_minus_1_lowered_fractional",
			p:         pa(bMinusSevenHalves(), -1),
			isNatural: false,
			isInteger: false,
			direction: spelling.DirectionLowered,
		},
		{
			name:      "zero_pitch_invalid",
			p:         pitch.Pitch{},
			isNatural: false,
			isInteger: false,
			direction: spelling.DirectionUnset,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.isNatural, tc.p.IsNatural())
			assert.Equal(t, tc.isInteger, tc.p.IsInteger())
			assert.Equal(t, tc.direction, tc.p.Direction())
		})
	}

	t.Run("all_sample_spellings_match_spelling_predicates", func(t *testing.T) {
		t.Parallel()

		for _, s := range sampleSpellings() {
			for _, aspn := range []int{math.MinInt, -1, 0, 4, math.MaxInt} {
				p := pa(s, aspn)
				assert.Equal(t, s.IsNatural(), p.IsNatural(), "s=%v aspn=%d", s, aspn)
				assert.Equal(t, s.IsInteger(), p.IsInteger(), "s=%v aspn=%d", s, aspn)
				assert.Equal(t, s.Direction(), p.Direction(), "s=%v aspn=%d", s, aspn)
			}
		}
	})
}

func TestSameParts(t *testing.T) {
	t.Parallel()

	var zero pitch.Pitch

	cases := []struct {
		name           string
		p              pitch.Pitch
		q              pitch.Pitch
		sameSpelling   bool
		sameOctave     bool
		sameLetter     bool
		sameAlteration bool
	}{
		{
			name:           "1_identical_CSharp_4_and_CSharp_4",
			p:              pa(spelling.CSharp(), 4),
			q:              pa(spelling.CSharp(), 4),
			sameSpelling:   true,
			sameOctave:     true,
			sameLetter:     true,
			sameAlteration: true,
		},
		{
			name:           "2_same_spelling_different_octave_CSharp_4_and_CSharp_5",
			p:              pa(spelling.CSharp(), 4),
			q:              pa(spelling.CSharp(), 5),
			sameSpelling:   true,
			sameOctave:     false,
			sameLetter:     true,
			sameAlteration: true,
		},
		{
			name:           "3_same_octave_and_letter_different_alteration_CSharp_4_and_CFlat_4",
			p:              pa(spelling.CSharp(), 4),
			q:              pa(spelling.CFlat(), 4),
			sameSpelling:   false,
			sameOctave:     true,
			sameLetter:     true,
			sameAlteration: false,
		},
		{
			name:           "4_same_alteration_different_letter_and_octave_CSharp_4_and_DSharp_5",
			p:              pa(spelling.CSharp(), 4),
			q:              pa(spelling.DSharp(), 5),
			sameSpelling:   false,
			sameOctave:     false,
			sameLetter:     false,
			sameAlteration: true,
		},
		{
			name:           "5a_enharmonic_boundary_BSharp_4_and_C_4",
			p:              pa(spelling.BSharp(), 4),
			q:              pa(spelling.C(), 4),
			sameSpelling:   false,
			sameOctave:     true,
			sameLetter:     false,
			sameAlteration: false,
		},
		{
			name:           "5b_enharmonic_CSharp_4_and_DFlat_4",
			p:              pa(spelling.CSharp(), 4),
			q:              pa(spelling.DFlat(), 4),
			sameSpelling:   false,
			sameOctave:     true,
			sameLetter:     false,
			sameAlteration: false,
		},
		{
			name:           "zero_and_zero",
			p:              zero,
			q:              zero,
			sameSpelling:   false,
			sameOctave:     false,
			sameLetter:     false,
			sameAlteration: false,
		},
		{
			name:           "zero_and_valid",
			p:              zero,
			q:              pa(spelling.C(), 4),
			sameSpelling:   false,
			sameOctave:     false,
			sameLetter:     false,
			sameAlteration: false,
		},
		{
			name:           "valid_and_zero",
			p:              pa(spelling.C(), 4),
			q:              zero,
			sameSpelling:   false,
			sameOctave:     false,
			sameLetter:     false,
			sameAlteration: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.sameSpelling, tc.p.SameSpelling(tc.q))
			assert.Equal(t, tc.sameSpelling, tc.q.SameSpelling(tc.p))

			assert.Equal(t, tc.sameOctave, tc.p.SameOctave(tc.q))
			assert.Equal(t, tc.sameOctave, tc.q.SameOctave(tc.p))

			assert.Equal(t, tc.sameLetter, tc.p.SameLetter(tc.q))
			assert.Equal(t, tc.sameLetter, tc.q.SameLetter(tc.p))

			assert.Equal(t, tc.sameAlteration, tc.p.SameAlteration(tc.q))
			assert.Equal(t, tc.sameAlteration, tc.q.SameAlteration(tc.p))
		})
	}

	t.Run("component_equivalence_invariants", func(t *testing.T) {
		t.Parallel()

		set := []pitch.Pitch{
			zero,
			pa(spelling.C(), 4),
			pa(spelling.CSharp(), 4),
			pa(spelling.CSharp(), 5),
			pa(spelling.DFlat(), 4),
			pa(spelling.BSharp(), 4),
			pa(cPlusHalf(), 4),
		}
		for i, p := range set {
			for j, q := range set {
				msg := fmt.Sprintf("i=%d (%v) j=%d (%v)", i, p, j, q)
				assert.Equal(t, p.SameSpelling(q) && p.SameOctave(q), p.Equal(q), msg)
				assert.Equal(t, p.SameLetter(q) && p.SameAlteration(q), p.SameSpelling(q), msg)
			}
		}
	})
}
