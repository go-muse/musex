package pcseq_test

import (
	"testing"

	"github.com/go-muse/musex/pcseq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDerivationOperatorModel(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		family pcseq.DerivationFamily
		text   string
	}{
		{family: pcseq.DerivationT, text: "T"},
		{family: pcseq.DerivationI, text: "I"},
		{family: pcseq.DerivationTR, text: "TR"},
		{family: pcseq.DerivationIR, text: "IR"},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			assert.True(t, tt.family.IsValid())
			assert.True(t, tt.family.Equal(tt.family))
			assert.Zero(t, tt.family.Compare(tt.family))
			assert.Equal(t, tt.text, tt.family.String())
		})
	}
	var invalid pcseq.DerivationFamily
	assert.False(t, invalid.IsValid())
	assert.False(t, invalid.Equal(pcseq.DerivationFamily(0)))
	assert.Zero(t, invalid.Compare(pcseq.DerivationFamily(99)))
	assert.Negative(t, invalid.Compare(pcseq.DerivationT))
	assert.Equal(t, "DerivationFamily(invalid)", invalid.String())

	for _, tt := range []struct {
		family   pcseq.DerivationFamily
		text     string
		expected []int
	}{
		{family: pcseq.DerivationT, text: "T1", expected: []int{1, 2, 3}},
		{family: pcseq.DerivationI, text: "I1", expected: []int{1, 0, 3}},
		{family: pcseq.DerivationTR, text: "T1R", expected: []int{3, 2, 1}},
		{family: pcseq.DerivationIR, text: "I1R", expected: []int{3, 0, 1}},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			op, err := pcseq.NewDerivationOperator(tt.family, 1, 4)
			require.NoError(t, err)
			assert.True(t, op.IsValid())
			assert.Equal(t, tt.family, op.Family())
			assert.Equal(t, 1, op.Index())
			assert.Equal(t, 4, op.Modulus())
			assert.Equal(t, tt.text, op.String())
			got, err := op.Apply(mustSequence(t, 4, 0, 1, 2))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got.Numbers())
			same, err := pcseq.NewDerivationOperator(tt.family, 1, 4)
			require.NoError(t, err)
			assert.True(t, op.Equal(same))
			assert.Equal(t, op, same)
			assert.Zero(t, op.Compare(same))
		})
	}

	_, err := pcseq.NewDerivationOperator(pcseq.DerivationFamily(0), 0, 12)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.NewDerivationOperator(pcseq.DerivationT, 12, 12)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.NewDerivationOperator(pcseq.DerivationT, 0, 0)
	assert.ErrorIs(t, err, pcseq.ErrInvalidModulus)
	_, err = pcseq.NewDerivationOperator(pcseq.DerivationT, 0, 65_537)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)

	var zero pcseq.DerivationOperator
	assert.False(t, zero.IsValid())
	assert.Equal(t, "DerivationOperator(invalid)", zero.String())
	assert.False(t, zero.Equal(pcseq.DerivationOperator{}))
	assert.Zero(t, zero.Compare(pcseq.DerivationOperator{}))
	_, err = zero.Apply(mustSequence(t, 12))
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	valid, err := pcseq.NewDerivationOperator(pcseq.DerivationT, 0, 12)
	require.NoError(t, err)
	_, err = valid.Apply(pcseq.Sequence{})
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = valid.Apply(mustSequence(t, 19))
	assert.ErrorIs(t, err, pcseq.ErrModulusMismatch)
}

func TestDerivationWebernAndAllAlternatives(t *testing.T) {
	t.Parallel()
	op24 := mustRow(t, 12, 0, 11, 3, 4, 8, 7, 9, 5, 6, 1, 2, 10)
	derived, err := op24.IsDerived(3)
	require.NoError(t, err)
	assert.True(t, derived)
	groups, err := op24.Derive(3)
	require.NoError(t, err)
	assert.Equal(t, []string{"[T0]", "[I7R]", "[T6R]", "[I1]"}, derivationGroupsStrings(groups))
	for _, group := range groups {
		for _, operator := range group {
			applied, err := operator.Apply(mustSequence(t, 12, 0, 11, 3))
			require.NoError(t, err)
			assert.True(t, applied.IsValid())
		}
	}

	row2 := mustRow(t, 2, 0, 1)
	groups, err = row2.Derive(1)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"[T0 I0 T0R I0R]",
		"[T1 I1 T1R I1R]",
	}, derivationGroupsStrings(groups))

	for _, length := range []int{0, 5, 13} {
		groups, err := op24.Derive(length)
		assert.ErrorIs(t, err, pcseq.ErrInvalidPartition)
		assert.Nil(t, groups)
		_, err = op24.IsDerived(length)
		assert.ErrorIs(t, err, pcseq.ErrInvalidPartition)
	}
	_, err = pcseq.Row{}.Derive(1)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.Row{}.IsDerived(1)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
}

func TestDerivationNotDerived(t *testing.T) {
	t.Parallel()
	row := mustRow(t, 6, 0, 1, 2, 3, 5, 4)
	derived, err := row.IsDerived(3)
	require.NoError(t, err)
	assert.False(t, derived)
	groups, err := row.Derive(3)
	assert.ErrorIs(t, err, pcseq.ErrNotDerived)
	assert.Nil(t, groups)
	assert.PanicsWithValue(t, err, func() { row.MustDerive(3) })

	large := mustRow(t, 1_025, identityNumbers(1_025)...)
	_, err = large.Derive(1)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
	_, err = large.IsDerived(1)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
}
