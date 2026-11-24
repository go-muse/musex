package tuplet_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A positive coefficient is written with positive components whatever the
// signs of its stored pair: two negative components are written without
// their signs, the pair is not reduced, and a denominator of 1 (or -1) gives
// the integer form. ParseLilyPondMultiplier reads the text back to a
// coefficient equal to the original by fraction.Fraction.Equal.
func TestFormatLilyPondMultiplierSigns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    fraction.Fraction
		want string
	}{
		{"positive pair", frac(2, 3), "*2/3"},
		{"negative pair", frac(-2, -3), "*2/3"},
		{"positive unreduced pair", frac(4, 6), "*4/6"},
		{"negative unreduced pair", frac(-4, -6), "*4/6"},
		{"positive integer", frac(4, 1), "*4"},
		{"negative integer pair", frac(-4, -1), "*4"},
		{"negative one", frac(-1, -1), "*1"},
		{"negative pair above one", frac(-7, -4), "*7/4"},
		{"negative pair with the largest int", frac(-math.MaxInt, -1), "*9223372036854775807"},
		{"negative pair with the largest int below", frac(-1, -math.MaxInt), "*1/9223372036854775807"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, 1, tt.c.Sign(), "the coefficient %s is positive", tt.c)

			got, err := tuplet.FormatLilyPondMultiplier(tt.c)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			if assertNoPanic(t, func() { tuplet.MustFormatLilyPondMultiplier(tt.c) }) {
				assert.Equal(t, tt.want, tuplet.MustFormatLilyPondMultiplier(tt.c))
			}

			back, err := tuplet.ParseLilyPondMultiplier(got)
			require.NoError(t, err, "the written text %q parses", got)
			assert.True(t, back.Equal(tt.c), "parsed %s, want a value equal to %s", back, tt.c)
		})
	}
}

// Writing and then reading a multiplier returns an equal coefficient for every
// stored pair with a positive value: both components positive or both
// negative.
func TestFormatLilyPondMultiplierRoundTripSigns(t *testing.T) {
	t.Parallel()

	for _, pair := range [][2]int{{1, 1}, {2, 3}, {3, 2}, {4, 6}, {6, 4}, {4, 1}, {1, 4}, {5, 7}, {12, 8}} {
		for _, s := range []int{1, -1} {
			c := frac(s*pair[0], s*pair[1])

			text, err := tuplet.FormatLilyPondMultiplier(c)
			if !assert.NoError(t, err, "format %s", c) {
				continue
			}

			back, err := tuplet.ParseLilyPondMultiplier(text)
			if !assert.NoError(t, err, "parse %q, the text of %s", text, c) {
				continue
			}

			assert.True(t, back.Equal(c), "%s written as %q parses to %s", c, text, back)
		}
	}
}

// A coefficient that is not positive fails with ErrInvalidCount whatever the
// signs of its pair, and an unset one with ErrInvalidArgument.
func TestFormatLilyPondMultiplierRejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    fraction.Fraction
		want error
	}{
		{"unset", fraction.Fraction{}, tuplet.ErrInvalidArgument},
		{"negative numerator", frac(-2, 3), tuplet.ErrInvalidCount},
		{"negative denominator", frac(2, -3), tuplet.ErrInvalidCount},
		{"negative integer", frac(-4, 1), tuplet.ErrInvalidCount},
		{"zero", frac(0, 3), tuplet.ErrInvalidCount},
		{"zero over a negative denominator", frac(0, -3), tuplet.ErrInvalidCount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tuplet.FormatLilyPondMultiplier(tt.c)
			assertErr(t, err, tt.want)
		})
	}
}
