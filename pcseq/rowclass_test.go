package pcseq_test

import (
	"reflect"
	"testing"

	"github.com/go-muse/musex/pcseq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRowClassModelKinds(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		kind pcseq.RowClassKind
		text string
	}{
		{kind: pcseq.TIR, text: "TIR"},
		{kind: pcseq.TIRRotations, text: "TIRRotations"},
		{kind: pcseq.TTOR, text: "TTOR"},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			assert.True(t, tt.kind.IsValid())
			assert.True(t, tt.kind.Equal(tt.kind))
			assert.Zero(t, tt.kind.Compare(tt.kind))
			assert.Equal(t, tt.text, tt.kind.String())
		})
	}
	var invalid pcseq.RowClassKind
	assert.False(t, invalid.IsValid())
	assert.False(t, invalid.Equal(pcseq.RowClassKind(0)))
	assert.Zero(t, invalid.Compare(pcseq.RowClassKind(99)))
	assert.Negative(t, invalid.Compare(pcseq.TIR))
	assert.Equal(t, "RowClassKind(invalid)", invalid.String())
	assert.Negative(t, pcseq.TIR.Compare(pcseq.TIRRotations))
	assert.Negative(t, pcseq.TIRRotations.Compare(pcseq.TTOR))
}

func TestRowClassModelSizesAndMembers(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		row      pcseq.Row
		expected int
	}{
		{name: "Lutyens", row: mustRow(t, 12, 0, 11, 3, 7, 8, 4, 2, 6, 5, 1, 9, 10), expected: 48},
		{name: "Webern op21", row: mustRow(t, 12, 0, 3, 2, 1, 5, 4, 10, 11, 7, 8, 9, 6), expected: 24},
		{name: "chromatic", row: mustRow(t, 12, identityNumbers(12)...), expected: 24},
		{name: "N1", row: mustRow(t, 1, 0), expected: 1},
		{name: "N2", row: mustRow(t, 2, 0, 1), expected: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			class, err := pcseq.NewRowClass(tt.row, pcseq.TIR)
			require.NoError(t, err)
			assert.True(t, class.IsValid())
			assert.Equal(t, tt.expected, class.Size())
			members, err := class.Members()
			require.NoError(t, err)
			assert.Len(t, members, tt.expected)
			unique := make(map[pcseq.Row]struct{}, len(members))
			for i, member := range members {
				unique[member] = struct{}{}
				assert.True(t, class.Contains(member))
				if i > 0 {
					assert.Negative(t, members[i-1].Compare(member))
				}
			}
			assert.Len(t, unique, tt.expected)
			assert.True(t, class.Contains(tt.row))
			assert.Equal(t, tt.row.Modulus(), class.Modulus())
			assert.Equal(t, pcseq.TIR, class.Kind())
		})
	}
}

func TestRowClassModelCanonicalIdentityAndEquivalence(t *testing.T) {
	t.Parallel()
	row := mustRow(t, 12, 3, 7, 11, 1, 5, 0, 2, 10, 6, 4, 8, 9)
	equivalent, err := row.Form(mustLabel(t, pcseq.RI, 8, 12))
	require.NoError(t, err)
	a, err := pcseq.NewRowClass(row, pcseq.TIR)
	require.NoError(t, err)
	b, err := pcseq.NewRowClass(equivalent, pcseq.TIR)
	require.NoError(t, err)
	assert.Equal(t, a.Representative(), b.Representative())
	assert.True(t, a.Equal(b))
	assert.Equal(t, a, b)
	assert.Zero(t, a.Compare(b))
	assert.Equal(t, a.String(), b.String())
	keys := map[pcseq.RowClass]int{a: 1}
	keys[b] = 2
	assert.Len(t, keys, 1)
	assert.True(t, reflect.TypeFor[pcseq.RowClass]().Comparable())

	equal, err := pcseq.AreEquivalentRows(row, equivalent, pcseq.TIR)
	require.NoError(t, err)
	assert.True(t, equal)
	other := mustRow(t, 12, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 10)
	equal, err = pcseq.AreEquivalentRows(row, other, pcseq.TIR)
	require.NoError(t, err)
	assert.False(t, equal)
	assert.False(t, a.Contains(other))
	assert.False(t, a.Contains(pcseq.Row{}))
	assert.False(t, a.Contains(mustRow(t, 19, identityNumbers(19)...)))

	_, err = pcseq.AreEquivalentRows(row, mustRow(t, 19, identityNumbers(19)...), pcseq.TIR)
	assert.ErrorIs(t, err, pcseq.ErrModulusMismatch)
	_, err = pcseq.AreEquivalentRows(pcseq.Row{}, row, pcseq.TIR)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.AreEquivalentRows(row, equivalent, pcseq.RowClassKind(0))
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.NewRowClass(row, pcseq.RowClassKind(0))
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.NewRowClass(pcseq.Row{}, pcseq.TIR)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
}

func TestRowClassModelRotationAndTTOGroups(t *testing.T) {
	t.Parallel()
	op24 := mustRow(t, 12, 0, 11, 3, 4, 8, 7, 9, 5, 6, 1, 2, 10)
	tir, err := pcseq.NewRowClass(op24, pcseq.TIR)
	require.NoError(t, err)
	rotations, err := pcseq.NewRowClass(op24, pcseq.TIRRotations)
	require.NoError(t, err)
	assert.Equal(t, 48, tir.Size())
	assert.Equal(t, 24, rotations.Size())
	tto, err := pcseq.NewRowClass(op24, pcseq.TTOR)
	require.NoError(t, err)
	assert.Equal(t, 96, tto.Size())
	assert.True(t, tto.Contains(op24))

	rotation, err := op24.Rotate(6)
	require.NoError(t, err)
	assert.True(t, rotations.Contains(rotation))
	equivalent, err := pcseq.AreEquivalentRows(op24, rotation, pcseq.TIRRotations)
	require.NoError(t, err)
	assert.True(t, equivalent)
	affine, err := op24.Apply(mustOperator(t, 5, 3, 12))
	require.NoError(t, err)
	assert.True(t, tto.Contains(affine))
	equivalent, err = pcseq.AreEquivalentRows(op24, affine, pcseq.TTOR)
	require.NoError(t, err)
	assert.True(t, equivalent)

	tirMembers, err := tir.Members()
	require.NoError(t, err)
	require.Len(t, tirMembers, 48)
	assert.Equal(t, tir.Representative(), tirMembers[0])
	rotationMembers, err := rotations.Members()
	require.NoError(t, err)
	require.Len(t, rotationMembers, 24)
	assert.Equal(t, rotations.Representative(), rotationMembers[0])
	ttoMembers, err := tto.Members()
	require.NoError(t, err)
	require.Len(t, ttoMembers, 96)
	assert.Equal(t, tto.Representative(), ttoMembers[0])

	for _, tt := range []struct {
		name    string
		class   pcseq.RowClass
		members []pcseq.Row
	}{
		{name: "TIR", class: tir, members: tirMembers},
		{name: "TIRRotations", class: rotations, members: rotationMembers},
		{name: "TTOR", class: tto, members: ttoMembers},
	} {
		t.Run(tt.name+" canonical members", func(t *testing.T) {
			t.Parallel()
			unique := make(map[pcseq.Row]struct{}, len(tt.members))
			for i, member := range tt.members {
				unique[member] = struct{}{}
				assert.True(t, tt.class.Contains(member))
				if i > 0 {
					assert.Negative(t, tt.members[i-1].Compare(member))
				}
			}
			assert.Len(t, unique, tt.class.Size())
			assert.Equal(t, tt.members[0], tt.class.Representative())
		})
	}

	rotationsEquivalent, err := pcseq.NewRowClass(rotation, pcseq.TIRRotations)
	require.NoError(t, err)
	assert.Equal(t, rotations.Representative(), rotationsEquivalent.Representative())
	assert.True(t, rotations.Equal(rotationsEquivalent))
	assert.True(t, rotations == rotationsEquivalent)
	ttoEquivalent, err := pcseq.NewRowClass(affine, pcseq.TTOR)
	require.NoError(t, err)
	assert.Equal(t, tto.Representative(), ttoEquivalent.Representative())
	assert.True(t, tto.Equal(ttoEquivalent))
	assert.True(t, tto == ttoEquivalent)

	for _, tt := range []struct {
		name     string
		row      pcseq.Row
		kind     pcseq.RowClassKind
		expected int
	}{
		{name: "N1 TIR", row: mustRow(t, 1, 0), kind: pcseq.TIR, expected: 1},
		{name: "N1 rotations", row: mustRow(t, 1, 0), kind: pcseq.TIRRotations, expected: 1},
		{name: "N1 TTO", row: mustRow(t, 1, 0), kind: pcseq.TTOR, expected: 1},
		{name: "N2 TIR", row: mustRow(t, 2, 0, 1), kind: pcseq.TIR, expected: 2},
		{name: "N2 rotations", row: mustRow(t, 2, 0, 1), kind: pcseq.TIRRotations, expected: 2},
		{name: "N2 TTO", row: mustRow(t, 2, 0, 1), kind: pcseq.TTOR, expected: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			class, err := pcseq.NewRowClass(tt.row, tt.kind)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, class.Size())
			members, err := class.Members()
			require.NoError(t, err)
			assert.Len(t, members, tt.expected)
		})
	}
}

func TestRowClassModelZeroAndString(t *testing.T) {
	t.Parallel()
	class, err := pcseq.NewRowClass(mustRow(t, 2, 0, 1), pcseq.TIR)
	require.NoError(t, err)
	assert.Equal(t, "TIR(<0, 1> mod 2)", class.String())

	var zero pcseq.RowClass
	assert.False(t, zero.IsValid())
	assert.Zero(t, zero.Modulus())
	assert.Equal(t, pcseq.RowClassKind(0), zero.Kind())
	assert.False(t, zero.Representative().IsValid())
	assert.Zero(t, zero.Size())
	assert.False(t, zero.Contains(mustRow(t, 2, 0, 1)))
	assert.Equal(t, "RowClass(invalid)", zero.String())
	assert.False(t, zero.Equal(pcseq.RowClass{}))
	assert.Zero(t, zero.Compare(pcseq.RowClass{}))
	assert.Negative(t, zero.Compare(class))
	_, err = zero.Members()
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
}
