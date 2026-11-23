package scale_test

import (
	"testing"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FromTokens reads a token as a written pitch before a name. In ABC the
// token "C" is both a name and a written pitch; the pitch reading wins, so
// the note keeps its octave.
func TestFromTokensPitchBeforeName(t *testing.T) {
	t.Parallel()

	src := settings.MustNew(tuning.Standard(), notation.MustBuiltIn(notation.ABCID), number.ASPN(), naming.English())

	// Both readings of the token exist in this notation.
	_, err := src.Parse("C")
	require.NoError(t, err)
	_, err = src.ParsePitch("C")
	require.NoError(t, err)

	s, err := scale.FromTokens(src, "C")
	require.NoError(t, err)
	n, err := s.At(0)
	require.NoError(t, err)
	assert.True(t, n.HasOctave(), "the written-pitch reading wins over the name reading")
	assert.True(t, n.Spelling().Equal(spelling.C()))
}
