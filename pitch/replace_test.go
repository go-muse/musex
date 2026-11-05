package pitch_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithSpellingOctaveAndASPN(t *testing.T) {
	t.Parallel()

	t.Run("WithSpelling_replaces_name_preserving_octave", func(t *testing.T) {
		t.Parallel()

		orig := pa(spelling.CSharp(), 4)
		before := orig

		got, err := orig.WithSpelling(spelling.DFlat())
		require.NoError(t, err)
		assert.Equal(t, pa(spelling.DFlat(), 4), got)
		assert.Equal(t, got, orig.MustWithSpelling(spelling.DFlat()))
		assert.True(t, got.SameOctave(orig))
		assert.Equal(t, before, orig, "receiver must remain unchanged")

		// Fractional shift via Spelling + WithSpelling (K-8).
		c4 := pa(spelling.C(), 4)
		half := alteration.MustFromFraction(1, 2)
		shiftedSpelling := c4.Spelling().MustRaiseByAlteration(half)
		gotFrac, err := c4.WithSpelling(shiftedSpelling)
		require.NoError(t, err)
		assert.Equal(t, pa(cPlusHalf(), 4), gotFrac)
		assert.Equal(t, gotFrac, c4.MustWithSpelling(shiftedSpelling))
		assert.Equal(t, pa(spelling.C(), 4), c4, "receiver must remain unchanged")
	})

	t.Run("WithOctave_and_WithASPN_replace_octave_preserving_spelling", func(t *testing.T) {
		t.Parallel()

		orig := pa(spelling.CSharp(), 4)
		before := orig

		gotOct, err := orig.WithOctave(o(6))
		require.NoError(t, err)
		assert.Equal(t, pa(spelling.CSharp(), 6), gotOct)
		assert.Equal(t, gotOct, orig.MustWithOctave(o(6)))
		assert.True(t, gotOct.SameSpelling(orig))
		assert.Equal(t, before, orig, "receiver must remain unchanged")

		for _, targetASPN := range []int{math.MinInt, -1, 0, 4, 6, math.MaxInt} {
			gotASPN, err := orig.WithASPN(targetASPN)
			require.NoError(t, err)
			assert.Equal(t, pa(spelling.CSharp(), targetASPN), gotASPN)
			assert.Equal(t, gotASPN, orig.MustWithASPN(targetASPN))
			assert.True(t, gotASPN.SameSpelling(orig))
			assert.Equal(t, targetASPN, gotASPN.MustASPN())
			assert.Equal(t, before, orig)
		}
	})

	t.Run("invalid_receiver_or_argument_returns_ErrInvalid", func(t *testing.T) {
		t.Parallel()

		var (
			zeroP pitch.Pitch
			zeroS spelling.Spelling
			zeroO octave.Octave
		)
		validP := pa(spelling.C(), 4)

		t.Run("valid_pitch_with_unset_spelling", func(t *testing.T) {
			t.Parallel()
			got, err := validP.WithSpelling(zeroS)
			requirePitchErrorIs(t, err, pitch.ErrInvalid)
			assert.Equal(t, pitch.Pitch{}, got)
			requireMustPanicsIs(t, func() {
				_ = validP.MustWithSpelling(zeroS)
			}, pitch.ErrInvalid)
		})

		t.Run("valid_pitch_with_invalid_octave", func(t *testing.T) {
			t.Parallel()
			got, err := validP.WithOctave(zeroO)
			requirePitchErrorIs(t, err, pitch.ErrInvalid)
			assert.Equal(t, pitch.Pitch{}, got)
			requireMustPanicsIs(t, func() {
				_ = validP.MustWithOctave(zeroO)
			}, pitch.ErrInvalid)
		})

		t.Run("zero_pitch_WithSpelling", func(t *testing.T) {
			t.Parallel()
			for _, arg := range []spelling.Spelling{spelling.C(), zeroS} {
				got, err := zeroP.WithSpelling(arg)
				requirePitchErrorIs(t, err, pitch.ErrInvalid)
				assert.Equal(t, pitch.Pitch{}, got)
				requireMustPanicsIs(t, func() {
					_ = zeroP.MustWithSpelling(arg)
				}, pitch.ErrInvalid)
			}
		})

		t.Run("zero_pitch_WithOctave", func(t *testing.T) {
			t.Parallel()
			for _, arg := range []octave.Octave{o(4), zeroO} {
				got, err := zeroP.WithOctave(arg)
				requirePitchErrorIs(t, err, pitch.ErrInvalid)
				assert.Equal(t, pitch.Pitch{}, got)
				requireMustPanicsIs(t, func() {
					_ = zeroP.MustWithOctave(arg)
				}, pitch.ErrInvalid)
			}
		})

		t.Run("zero_pitch_WithASPN", func(t *testing.T) {
			t.Parallel()
			for _, aspn := range []int{math.MinInt, 0, 4, math.MaxInt} {
				got, err := zeroP.WithASPN(aspn)
				requirePitchErrorIs(t, err, pitch.ErrInvalid)
				assert.Equal(t, pitch.Pitch{}, got)
				requireMustPanicsIs(t, func() {
					_ = zeroP.MustWithASPN(aspn)
				}, pitch.ErrInvalid)
			}
		})
	})
}

func TestNatural(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   pitch.Pitch
		want pitch.Pitch
	}{
		{
			name: "CSharp_4_to_C_4",
			in:   pa(spelling.CSharp(), 4),
			want: pa(spelling.C(), 4),
		},
		{
			name: "BDoubleFlat_3_to_B_3",
			in:   pa(spelling.BDoubleFlat(), 3),
			want: pa(spelling.B(), 3),
		},
		{
			name: "BSharp_4_preserves_B_and_ASPN_4",
			in:   pa(spelling.BSharp(), 4),
			want: pa(spelling.B(), 4),
		},
		{
			name: "C_plus_half_minus_1_to_C_minus_1",
			in:   pa(cPlusHalf(), -1),
			want: pa(spelling.C(), -1),
		},
		{
			name: "already_natural_C_4",
			in:   pa(spelling.C(), 4),
			want: pa(spelling.C(), 4),
		},
		{
			name: "zero_pitch_returns_zero_pitch",
			in:   pitch.Pitch{},
			want: pitch.Pitch{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			before := tc.in
			got := tc.in.Natural()
			assert.Equal(t, tc.want, got)
			assert.Equal(t, before, tc.in, "receiver must remain unchanged")
		})
	}

	t.Run("all_predefined_spellings_in_ASPN_4_and_boundaries", func(t *testing.T) {
		t.Parallel()

		for _, s := range sampleSpellings() {
			for _, aspn := range []int{math.MinInt, 4, math.MaxInt} {
				p := pa(s, aspn)
				nat := p.Natural()
				msg := fmt.Sprintf("s=%v aspn=%d", s, aspn)
				assert.True(t, nat.IsValid(), msg)
				assert.True(t, nat.IsNatural(), msg)
				assert.True(t, nat.IsInteger(), msg)
				assert.Equal(t, spelling.DirectionNatural, nat.Direction(), msg)
				assert.True(t, nat.SameLetter(p), msg)
				assert.True(t, nat.SameOctave(p), msg)
				assert.Equal(t, nat, nat.Natural(), msg)
			}
		}
	})
}
