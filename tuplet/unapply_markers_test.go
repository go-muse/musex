package tuplet_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UnapplyChain undoes a complete chain of tuplet coefficients.
func TestUnapplyChainAcceptance(t *testing.T) {
	t.Parallel()

	got, err := tuplet.UnapplyChain(fraction.MustNewReduced(1, 15), tuplet.Quintuplet(), tuplet.Triplet())
	require.NoError(t, err)
	assert.Equal(t, fraction.MustNewReduced(1, 8), got)

	// An empty chain preserves the numerical value.
	got, err = tuplet.UnapplyChain(fraction.MustNew(6, 16))
	require.NoError(t, err)
	assert.Equal(t, fraction.MustNewReduced(3, 8), got)

	// UnapplyChain undoes ApplyChain.
	d := fraction.MustNewReduced(1, 8)
	chain := []tuplet.Tuplet{tuplet.Quintuplet(), tuplet.Triplet()}
	applied, err := tuplet.ApplyChain(d, chain...)
	require.NoError(t, err)
	back, err := tuplet.UnapplyChain(applied, chain...)
	require.NoError(t, err)
	assert.Equal(t, d, back)
}

// ABC markers use the ABC 2.1 defaults in the meter context.
func TestABCAcceptance(t *testing.T) {
	t.Parallel()

	cases := []struct {
		text     string
		division tuplet.Division
		want     tuplet.Tuplet
		events   int
	}{
		{"(5", tuplet.Binary, tuplet.MustNew(5, 2), 5},
		{"(5", tuplet.Ternary, tuplet.MustNew(5, 3), 5},
		{"(3", tuplet.Binary, tuplet.Triplet(), 3},
		{"(2", tuplet.Binary, tuplet.Duplet(), 2},
		{"(6", tuplet.Binary, tuplet.MustNew(6, 2), 6},
		{"(3:2:2", tuplet.Binary, tuplet.Triplet(), 2},
		{"(3::2", tuplet.Binary, tuplet.Triplet(), 2},
		{"(7:4", tuplet.Ternary, tuplet.MustNew(7, 4), 7},
	}

	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()

			m, err := tuplet.ParseABCMarker(tc.text, tc.division)
			require.NoError(t, err)
			assert.Equal(t, tc.want, m.Tuplet)
			assert.Equal(t, tc.events, m.Events)

			// Parsing a formatted marker returns the same marker in the same
			// context.
			text, err := tuplet.FormatABCMarker(m, tc.division)
			require.NoError(t, err)
			back, err := tuplet.ParseABCMarker(text, tc.division)
			require.NoError(t, err)
			assert.Equal(t, m, back)
		})
	}

	_, err := tuplet.ParseABCMarker("(10", tuplet.Binary)
	assert.ErrorIs(t, err, tuplet.ErrNoDefault)
}

// LilyPond group markers and duration multipliers stay separate.
func TestLilyPondAcceptance(t *testing.T) {
	t.Parallel()

	got, err := tuplet.ParseLilyPondTuplet(`\tuplet 3/2`)
	require.NoError(t, err)
	assert.Equal(t, tuplet.Triplet(), got)

	text, err := tuplet.FormatLilyPondTuplet(got)
	require.NoError(t, err)
	assert.Equal(t, `\tuplet 3/2`, text)

	coef, err := tuplet.ParseLilyPondMultiplier("*2/3")
	require.NoError(t, err)
	assert.Equal(t, fraction.MustNew(2, 3), coef)

	text, err = tuplet.FormatLilyPondMultiplier(coef)
	require.NoError(t, err)
	assert.Equal(t, "*2/3", text)

	// The optional span duration is rejected.
	_, err = tuplet.ParseLilyPondTuplet(`\tuplet 3/2 4`)
	assert.ErrorIs(t, err, tuplet.ErrInvalidText)
}
