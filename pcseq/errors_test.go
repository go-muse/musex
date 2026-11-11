package pcseq_test

import (
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSequenceCoreErrorPrecedence(t *testing.T) {
	t.Parallel()
	_, err := pcseq.NewRow(0, 12)
	assert.ErrorIs(t, err, pcseq.ErrInvalidModulus)
	_, err = pcseq.NewRow(65_537, 0)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
	_, err = pcseq.NewFormLabel(pcseq.FormKind(0), 12, 0)
	assert.ErrorIs(t, err, pcseq.ErrInvalidModulus)
	_, err = pcseq.NewFormLabel(pcseq.FormKind(0), 12, 65_537)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
	_, err = pcseq.Sequence{}.Apply(pitchclass.Operator{})
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
}

func TestDerivationErrorPrecedence(t *testing.T) {
	t.Parallel()
	row := mustRow(t, 1_025, identityNumbers(1_025)...)
	_, err := row.Derive(0)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
	_, err = row.IsDerived(0)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
}

func TestCountRowsErrorPrecedenceAndOverflow(t *testing.T) {
	t.Parallel()
	universe, err := pcset.Universe(21)
	require.NoError(t, err)
	got, err := pcseq.CountRowsWithSegmentContents(21, universe)
	assert.Zero(t, got)
	assert.ErrorIs(t, err, pcseq.ErrOverflow)
	exact, err := pcseq.CountRowsWithSegmentContentsBig(21, universe)
	require.NoError(t, err)
	assert.Equal(t, "51090942171709440000", exact.String())
}

func TestParseErrorPrecedence(t *testing.T) {
	t.Parallel()
	_, err := pcseq.ParseNotation("not a sequence", pitchclass.TE, 0)
	assert.ErrorIs(t, err, pcseq.ErrInvalidModulus)
	_, err = pcseq.ParseNotation("not a sequence", pitchclass.TE, 65_537)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
	_, err = pcseq.ParseNotation("not a sequence", pitchclass.Style(99), 12)
	assert.ErrorIs(t, err, pcseq.ErrNotation)
}
