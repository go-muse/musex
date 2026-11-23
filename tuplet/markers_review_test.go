package tuplet_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FormatABCMarker chooses the documented form: the short form when the
// normal count is the ABC default and the event count equals the actual
// count, the two-count form when only the event count differs.
func TestABCFormatChoice(t *testing.T) {
	t.Parallel()

	cases := []struct {
		m    tuplet.ABCMarker
		div  tuplet.Division
		want string
	}{
		{tuplet.ABCMarker{Tuplet: tuplet.Triplet(), Events: 3}, tuplet.Binary, "(3"},
		{tuplet.ABCMarker{Tuplet: tuplet.Triplet(), Events: 2}, tuplet.Binary, "(3:2:2"},
		{tuplet.ABCMarker{Tuplet: tuplet.MustNew(5, 2), Events: 5}, tuplet.Binary, "(5"},
		{tuplet.ABCMarker{Tuplet: tuplet.MustNew(5, 4), Events: 5}, tuplet.Binary, "(5:4"},
		{tuplet.ABCMarker{Tuplet: tuplet.MustNew(5, 3), Events: 5}, tuplet.Ternary, "(5"},
		{tuplet.ABCMarker{Tuplet: tuplet.MustNew(5, 2), Events: 5}, tuplet.Ternary, "(5:2"},
		{tuplet.ABCMarker{Tuplet: tuplet.Triplet(), Events: 3}, tuplet.Ternary, "(3"},
	}

	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tuplet.MustFormatABCMarker(tc.m, tc.div))
			back, err := tuplet.ParseABCMarker(tc.want, tc.div)
			require.NoError(t, err)
			assert.Equal(t, tc.m, back)
		})
	}
}

// LilyPond integer multipliers are read and written in the integer form.
func TestLilyPondIntegerMultiplier(t *testing.T) {
	t.Parallel()

	c, err := tuplet.ParseLilyPondMultiplier("*4")
	require.NoError(t, err)
	assert.Equal(t, fraction.MustNew(4, 1), c)
	assert.Equal(t, "*4", tuplet.MustFormatLilyPondMultiplier(c))

	c2, err := tuplet.ParseLilyPondMultiplier("*4/1")
	require.NoError(t, err)
	assert.Equal(t, c, c2)
	assert.Equal(t, "*4", tuplet.MustFormatLilyPondMultiplier(c2))
}
