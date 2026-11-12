package pcseq_test

import (
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/setclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rowPermutations(numbers []int) [][]int {
	if len(numbers) == 0 {
		return [][]int{{}}
	}
	result := make([][]int, 0)
	for i := range numbers {
		rest := make([]int, 0, len(numbers)-1)
		rest = append(rest, numbers[:i]...)
		rest = append(rest, numbers[i+1:]...)
		for _, tail := range rowPermutations(rest) {
			row := make([]int, 0, len(numbers))
			row = append(row, numbers[i])
			row = append(row, tail...)
			result = append(result, row)
		}
	}
	return result
}

func TestRowClassModelExhaustiveSmallRotationCanonical(t *testing.T) {
	for _, modulus := range []int{4, 5, 6} {
		t.Run(bigLengthName(modulus), func(t *testing.T) {
			t.Parallel()
			for _, numbers := range rowPermutations(identityNumbers(modulus)) {
				row := mustRow(t, modulus, numbers...)
				class, err := pcseq.NewRowClass(row, pcseq.TIRRotations)
				require.NoError(t, err)
				members, err := class.Members()
				require.NoError(t, err)
				require.NotEmpty(t, members)
				assert.Equal(t, members[0], class.Representative())
				fromCanonical, err := pcseq.NewRowClass(class.Representative(), pcseq.TIRRotations)
				require.NoError(t, err)
				assert.True(t, class.Equal(fromCanonical))
				assert.True(t, class.Contains(row))
			}
		})
	}
}

func TestCombinatorialPairLabelErrorPositions(t *testing.T) {
	t.Parallel()
	row := mustRow(t, 4, 0, 1, 2, 3)
	for _, tt := range []struct {
		name   string
		a      pcseq.FormLabel
		b      pcseq.FormLabel
		target error
	}{
		{name: "invalid first", a: pcseq.FormLabel{}, b: mustLabel(t, pcseq.P, 0, 4), target: pcseq.ErrInvalidFormLabel},
		{name: "invalid second", a: mustLabel(t, pcseq.P, 0, 4), b: pcseq.FormLabel{}, target: pcseq.ErrInvalidFormLabel},
		{name: "both invalid", a: pcseq.FormLabel{}, b: pcseq.FormLabel{}, target: pcseq.ErrInvalidFormLabel},
		{name: "both from another modulus", a: mustLabel(t, pcseq.P, 0, 5), b: mustLabel(t, pcseq.R, 0, 5), target: pcseq.ErrModulusMismatch},
		{name: "invalid first and second modulus", a: pcseq.FormLabel{}, b: mustLabel(t, pcseq.R, 0, 5), target: pcseq.ErrInvalidFormLabel},
		{name: "first modulus and invalid second", a: mustLabel(t, pcseq.P, 0, 5), b: pcseq.FormLabel{}, target: pcseq.ErrModulusMismatch},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ok, err := row.IsCombinatorialPair(tt.a, tt.b)
			assert.ErrorIs(t, err, tt.target)
			assert.False(t, ok)
		})
	}
}

func TestSequenceCoreContentClassErrorPrecedence(t *testing.T) {
	t.Parallel()
	large := mustSequence(t, 257, 0, 1, 2)
	_, err := large.ContentClass(setclass.Kind(99))
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.Sequence{}.ContentClass(setclass.Kind(99))
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
}

func TestContentInvariancesAdditionalBoundaries(t *testing.T) {
	t.Parallel()
	a := mustSequence(t, 12, 0, 1, 2)
	b := mustSequence(t, 12, 2, 1, 0)
	got, err := a.InvariancesTo(b)
	require.NoError(t, err)
	assert.Equal(t, []int{0}, got.Transpositions)
	assert.Equal(t, []int{2}, got.Inversions)

	_, err = a.InvariancesTo(mustSequence(t, 19, 0))
	assert.ErrorIs(t, err, pcseq.ErrModulusMismatch)
	_, err = a.InvariancesTo(mustSequence(t, 19, 0, 1, 2))
	assert.ErrorIs(t, err, pcseq.ErrModulusMismatch)
}

func TestDerivationApplyErrorsByFamily(t *testing.T) {
	t.Parallel()
	for _, family := range []pcseq.DerivationFamily{
		pcseq.DerivationT,
		pcseq.DerivationI,
		pcseq.DerivationTR,
		pcseq.DerivationIR,
	} {
		t.Run(family.String(), func(t *testing.T) {
			t.Parallel()
			op, err := pcseq.NewDerivationOperator(family, 0, 12)
			require.NoError(t, err)
			_, err = op.Apply(mustSequence(t, 19, 0))
			assert.ErrorIs(t, err, pcseq.ErrModulusMismatch)
			_, err = op.Apply(pcseq.Sequence{})
			assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
		})
	}
}
