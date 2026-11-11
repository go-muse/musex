package pcseq_test

import (
	"testing"

	"github.com/go-muse/musex/pcseq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatrixMoveableZeroModel(t *testing.T) {
	t.Parallel()
	source := mustRow(t, 3, 1, 0, 2)
	matrix, err := pcseq.NewMatrix(source)
	require.NoError(t, err)
	assert.True(t, matrix.IsValid())
	assert.Equal(t, 3, matrix.Modulus())
	assert.Equal(t, source, matrix.Source())
	assert.Equal(t, "Matrix(<1, 0, 2> mod 3)", matrix.String())

	expectedRows := [][]int{{0, 2, 1}, {1, 0, 2}, {2, 1, 0}}
	expectedColumns := [][]int{{0, 1, 2}, {2, 0, 1}, {1, 2, 0}}
	expectedRowLabels := []string{"P0", "P1", "P2"}
	expectedColumnLabels := []string{"I0", "I2", "I1"}
	for i := range 3 {
		row, err := matrix.Row(i)
		require.NoError(t, err)
		assert.Equal(t, expectedRows[i], row.Numbers())
		rowLabel, err := matrix.RowLabel(i)
		require.NoError(t, err)
		assert.Equal(t, expectedRowLabels[i], rowLabel.String())
		fromLabel, err := source.Form(rowLabel)
		require.NoError(t, err)
		assert.Equal(t, fromLabel, row)

		column, err := matrix.Column(i)
		require.NoError(t, err)
		assert.Equal(t, expectedColumns[i], column.Numbers())
		columnLabel, err := matrix.ColumnLabel(i)
		require.NoError(t, err)
		assert.Equal(t, expectedColumnLabels[i], columnLabel.String())
		fromLabel, err = source.Form(columnLabel)
		require.NoError(t, err)
		assert.Equal(t, fromLabel, column)

		for j := range 3 {
			cell, err := matrix.Cell(i, j)
			require.NoError(t, err)
			assert.Equal(t, expectedRows[i][j], cell.Number())
		}
	}

	retrograde, err := matrix.RetrogradeRow(0)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 0}, retrograde.Numbers())
	assert.Equal(t, mustRow(t, 3, 1, 2, 0), retrograde)
	ri, err := matrix.RetrogradeInversionColumn(0)
	require.NoError(t, err)
	assert.Equal(t, []int{2, 1, 0}, ri.Numbers())

	cells, err := matrix.Cells()
	require.NoError(t, err)
	require.Len(t, cells, 3)
	for i := range cells {
		assert.Equal(t, expectedRows[i], classNumbers(cells[i]))
	}
}

func TestMatrixLutyensFormulaAndLabels(t *testing.T) {
	t.Parallel()
	source := mustRow(t, 12, 0, 11, 3, 7, 8, 4, 2, 6, 5, 1, 9, 10)
	matrix, err := pcseq.NewMatrix(source)
	require.NoError(t, err)
	cells, err := matrix.Cells()
	require.NoError(t, err)
	require.Len(t, cells, 12)
	p := source.Numbers()
	cellCount := 0
	for i := range 12 {
		require.Len(t, cells[i], 12)
		for j := range 12 {
			cell, err := matrix.Cell(i, j)
			require.NoError(t, err)
			expected := (p[j] - p[i] + 12) % 12
			assert.Equal(t, expected, cell.Number())
			assert.Equal(t, expected, cells[i][j].Number())
			cellCount++
		}
	}
	assert.Equal(t, 144, cellCount)
	assert.Equal(t, []int{0, 11, 3}, classNumbers(cells[0])[:3])
	assert.Equal(t, []int{1, 0, 4}, classNumbers(cells[1])[:3])
	assert.Equal(t, []int{9, 8, 0}, classNumbers(cells[2])[:3])

	for i, expected := range []string{"P0", "P1", "P9", "P5"} {
		label, err := matrix.RowLabel(i)
		require.NoError(t, err)
		assert.Equal(t, expected, label.String())
	}
	for j, expected := range []string{"I0", "I11", "I3", "I7"} {
		label, err := matrix.ColumnLabel(j)
		require.NoError(t, err)
		assert.Equal(t, expected, label.String())
	}
}

func TestMatrixSingleAndErrors(t *testing.T) {
	t.Parallel()
	one, err := pcseq.NewMatrix(mustRow(t, 1, 0))
	require.NoError(t, err)
	cell, err := one.Cell(0, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, cell.Number())
	rowLabel, err := one.RowLabel(0)
	require.NoError(t, err)
	assert.Equal(t, "P0", rowLabel.String())
	columnLabel, err := one.ColumnLabel(0)
	require.NoError(t, err)
	assert.Equal(t, "I0", columnLabel.String())

	valid, err := pcseq.NewMatrix(mustRow(t, 3, 0, 1, 2))
	require.NoError(t, err)
	for _, call := range []func() error{
		func() error { _, err := valid.Cell(-1, 0); return err },
		func() error { _, err := valid.Cell(0, 3); return err },
		func() error { _, err := valid.Row(3); return err },
		func() error { _, err := valid.Column(-1); return err },
		func() error { _, err := valid.RowLabel(4); return err },
		func() error { _, err := valid.ColumnLabel(4); return err },
		func() error { _, err := valid.RetrogradeRow(4); return err },
		func() error { _, err := valid.RetrogradeInversionColumn(4); return err },
	} {
		assert.ErrorIs(t, call(), pcseq.ErrInvalidSegment)
	}

	_, err = pcseq.NewMatrix(pcseq.Row{})
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	var zero pcseq.Matrix
	assert.False(t, zero.IsValid())
	assert.Zero(t, zero.Modulus())
	assert.False(t, zero.Source().IsValid())
	assert.Equal(t, "Matrix(invalid)", zero.String())
	_, err = zero.Cell(0, 0)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = zero.Cells()
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)

	large := mustRow(t, 1_025, identityNumbers(1_025)...)
	_, err = pcseq.NewMatrix(large)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
}
