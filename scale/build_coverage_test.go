package scale_test

import (
	"strings"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenErrorMessage(t *testing.T) {
	t.Parallel()

	_, err := scale.FromTokens(settings.Standard(), "C", "X")
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "token 1"))
}

func TestNewMustVariants(t *testing.T) {
	t.Parallel()

	src := settings.Standard()
	assert.Equal(t, 2, scale.MustFromSpellings([]spelling.Spelling{spelling.C(), spelling.D()}, src).Len())
	assert.Panics(t, func() { scale.MustFromSpellings([]spelling.Spelling{{}}, src) })

	assert.Equal(t, 1, scale.MustFromTokens(src, "C").Len())
	assert.Panics(t, func() { scale.MustFromTokens(src, "X") })

	s := scale.MustFromTokens(src, "C4")
	assert.Equal(t, []string{"C4"}, s.MustTokens())
	assert.Equal(t, []string{"C4"}, s.MustTokensIn(src))
	assert.Equal(t, 1, s.MustPitchClassSequence(system.Traditional()).Len())

	assert.Panics(t, func() { scale.MustNew().MustTokensIn(nil) })
	assert.Panics(t, func() {
		var nilScale *scale.Scale
		nilScale.MustPitchClassSequence(system.Traditional())
	})

	from := pitch.MustNew(spelling.C(), octave.FromASPN(4))
	to := pitch.MustNew(spelling.C(), octave.FromASPN(5))
	rule := system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)
	assert.Equal(t, 13, scale.MustChromaticRun(from, to, rule, src).Len())
	assert.Panics(t, func() { scale.MustChromaticRun(from, to, rule, nil) })

	// Unset settings are rejected.
	_, err := scale.ChromaticRun(from, to, rule, &settings.Settings{})
	assert.ErrorIs(t, err, scale.ErrUnset)
}
