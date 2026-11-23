package scale_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromSpellingsEdges(t *testing.T) {
	t.Parallel()

	_, err := scale.FromSpellings([]spelling.Spelling{{}}, settings.Standard())
	assert.ErrorIs(t, err, note.ErrUnset)

	s, err := scale.FromSpellings(nil, settings.Standard())
	require.NoError(t, err)
	assert.True(t, s.IsEmpty())
}

func TestFromTokensEdges(t *testing.T) {
	t.Parallel()

	_, err := scale.FromTokens(nil, "C")
	assert.ErrorIs(t, err, scale.ErrUnset)

	s, err := scale.FromTokens(settings.Standard())
	require.NoError(t, err)
	assert.True(t, s.IsEmpty())

	// A written pitch token keeps its octave.
	sp, err := scale.FromTokens(settings.Standard(), "C4")
	require.NoError(t, err)
	n, err := sp.At(0)
	require.NoError(t, err)
	assert.True(t, n.HasOctave())
}

func TestTokensEdges(t *testing.T) {
	t.Parallel()

	var nilScale *scale.Scale
	_, err := nilScale.Tokens()
	assert.ErrorIs(t, err, scale.ErrUnset)
	_, err = nilScale.TokensIn(settings.Standard())
	assert.ErrorIs(t, err, scale.ErrUnset)
	assert.Nil(t, nilScale.Spellings())

	_, err = scale.MustNew().TokensIn(nil)
	assert.ErrorIs(t, err, scale.ErrUnset)

	// An unbound note cannot be formatted in its own source.
	unbound := note.MustFromSpelling(spelling.C(), cents.Cents{}, nil)
	_, err = scale.MustNew(unbound).Tokens()
	assert.Error(t, err)
}

func TestChromaticRunEdges(t *testing.T) {
	t.Parallel()

	src := settings.Standard()
	c4 := pitch.MustNew(spelling.C(), octave.FromASPN(4))
	c5 := pitch.MustNew(spelling.C(), octave.FromASPN(5))
	rule := system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)

	_, err := scale.ChromaticRun(c4, c5, rule, nil)
	assert.ErrorIs(t, err, scale.ErrUnset)

	_, err = scale.ChromaticRun(c4, c5, system.SpellingRule{}, src)
	assert.ErrorIs(t, err, system.ErrInvalid)

	// Equal endpoints give one note.
	one, err := scale.ChromaticRun(c4, c4, rule, src)
	require.NoError(t, err)
	assert.Equal(t, 1, one.Len())

	// Descending order follows the endpoints.
	desc := system.MustChromaticScale(spelling.C(), system.Major, system.Descending)
	down, err := scale.ChromaticRun(c5, c4, desc, src)
	require.NoError(t, err)
	require.Equal(t, 13, down.Len())
	first, err := down.At(0)
	require.NoError(t, err)
	assert.True(t, first.MustPitch().Equal(c5))
}

func TestPitchClassSequenceEdges(t *testing.T) {
	t.Parallel()

	var nilScale *scale.Scale
	_, err := nilScale.PitchClassSequence(system.Traditional())
	assert.ErrorIs(t, err, scale.ErrUnset)

	_, err = scale.MustNew().PitchClassSequence(nil)
	assert.ErrorIs(t, err, system.ErrInvalid)

	empty, err := scale.MustNew().PitchClassSequence(system.Traditional())
	require.NoError(t, err)
	assert.True(t, empty.IsEmpty())
	assert.Equal(t, 12, empty.Modulus())
}
