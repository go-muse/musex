package scale_test

import (
	"errors"
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

// FromSpellings and FromTokens preserve all entries, spellings, and order
func TestFromSpellingsAndTokensAcceptance(t *testing.T) {
	t.Parallel()

	src := settings.Standard()
	names := []spelling.Spelling{spelling.C(), spelling.DFlat(), spelling.CSharp(), spelling.C()}

	s, err := scale.FromSpellings(names, src)
	require.NoError(t, err)
	require.Equal(t, 4, s.Len())
	assert.Equal(t, names, s.Spellings())

	first, err := s.At(0)
	require.NoError(t, err)
	assert.False(t, first.HasOctave())

	st, err := scale.FromTokens(src, "C", "D♭", "C♯", "C")
	require.NoError(t, err)
	require.Equal(t, 4, st.Len())
	assert.Equal(t, names, st.Spellings())

	tokens, err := st.Tokens()
	require.NoError(t, err)
	assert.Equal(t, []string{"C", "D♭", "C♯", "C"}, tokens)

	// The failing token is identified by index, reachable with errors.As.
	_, err = scale.FromTokens(src, "C", "X", "D")
	require.Error(t, err)
	assert.ErrorIs(t, err, scale.ErrToken)
	var te *scale.TokenError
	require.True(t, errors.As(err, &te))
	assert.Equal(t, 1, te.Index)
}

// ChromaticRun builds a playable run with the directional spelling rule
func TestChromaticRunAcceptance(t *testing.T) {
	t.Parallel()

	src := settings.Standard()
	from := pitch.MustNew(spelling.C(), octave.FromASPN(4))
	to := pitch.MustNew(spelling.C(), octave.FromASPN(5))
	rule := system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)

	s, err := scale.ChromaticRun(from, to, rule, src)
	require.NoError(t, err)
	require.Equal(t, 13, s.Len())

	first, err := s.At(0)
	require.NoError(t, err)
	last, err := s.At(12)
	require.NoError(t, err)
	firstPitch, err := first.Pitch()
	require.NoError(t, err)
	lastPitch, err := last.Pitch()
	require.NoError(t, err)
	assert.True(t, firstPitch.Equal(from))
	assert.True(t, lastPitch.Equal(to))

	tenth, err := s.At(10)
	require.NoError(t, err)
	assert.True(t, tenth.Spelling().Equal(spelling.BFlat()), "tenth semitone of C major ascending")

	for i := 0; i < s.Len(); i++ {
		n, err := s.At(i)
		require.NoError(t, err)
		assert.True(t, n.HasOctave())
		assert.True(t, settings.SameSource(n.Source(), src))
	}
}

// PitchClassSequence keeps one position per note, with repetitions and
// enharmonic equivalents.
func TestPitchClassSequenceAcceptance(t *testing.T) {
	t.Parallel()

	src := settings.Standard()
	notes := []note.Note{
		note.MustFromPitch(pitch.MustNew(spelling.C(), octave.FromASPN(4)), cents.Cents{}, src),
		note.MustFromPitch(pitch.MustNew(spelling.DFlat(), octave.FromASPN(4)), cents.Cents{}, src),
		note.MustFromPitch(pitch.MustNew(spelling.CSharp(), octave.FromASPN(5)), cents.Cents{}, src),
		note.MustFromPitch(pitch.MustNew(spelling.D(), octave.FromASPN(4)), cents.Cents{}, src),
	}
	s, err := scale.New(notes...)
	require.NoError(t, err)

	seq, err := s.PitchClassSequence(system.Traditional())
	require.NoError(t, err)
	require.Equal(t, 4, seq.Len())
	for i, want := range []int{0, 1, 1, 2} {
		got, err := seq.At(i)
		require.NoError(t, err)
		assert.Equal(t, want, got.Number())
	}
}
