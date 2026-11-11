package pcseq_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOwnershipSlicesAndInputs(t *testing.T) {
	s := mustSequence(t, 12, 7, 11, 9, 6, 7)
	numbers := s.Numbers()
	numbers[0] = 99
	assert.Equal(t, []int{7, 11, 9, 6, 7}, s.Numbers())
	classes := s.Classes()
	classes[0] = pitchclass.Class{}
	assert.Equal(t, []int{7, 11, 9, 6, 7}, classNumbers(s.Classes()))
	intervals, err := s.Intervals()
	require.NoError(t, err)
	intervals[0] = pitchclass.Class{}
	freshIntervals, err := s.Intervals()
	require.NoError(t, err)
	assert.Equal(t, []int{4, 10, 9, 1}, classNumbers(freshIntervals))

	segments, err := s.DiscreteSegments(1)
	require.NoError(t, err)
	segments[0] = pcseq.Sequence{}
	freshSegments, err := s.DiscreteSegments(1)
	require.NoError(t, err)
	assert.Equal(t, s.Numbers(), freshSegments[0].Numbers())

	row := mustRow(t, 12, 3, 7, 11, 1, 5, 0, 2, 10, 6, 4, 8, 9)
	form, err := row.Form(mustLabel(t, pcseq.P, 0, 12))
	require.NoError(t, err)
	labels, err := row.LabelsOf(form)
	require.NoError(t, err)
	labels[0] = pcseq.FormLabel{}
	freshLabels, err := row.LabelsOf(form)
	require.NoError(t, err)
	assert.Equal(t, "P0", freshLabels[0].String())

	matrix, err := pcseq.NewMatrix(row)
	require.NoError(t, err)
	cells, err := matrix.Cells()
	require.NoError(t, err)
	cells[0][0] = pitchclass.Class{}
	freshCells, err := matrix.Cells()
	require.NoError(t, err)
	assert.Equal(t, 0, freshCells[0][0].Number())

	class, err := pcseq.NewRowClass(row, pcseq.TIR)
	require.NoError(t, err)
	members, err := class.Members()
	require.NoError(t, err)
	members[0] = pcseq.Row{}
	freshMembers, err := class.Members()
	require.NoError(t, err)
	assert.True(t, freshMembers[0].IsValid())

	op24 := mustRow(t, 12, 0, 11, 3, 4, 8, 7, 9, 5, 6, 1, 2, 10)
	groups, err := op24.Derive(3)
	require.NoError(t, err)
	groups[0][0] = pcseq.DerivationOperator{}
	freshGroups, err := op24.Derive(3)
	require.NoError(t, err)
	assert.Equal(t, "T0", freshGroups[0][0].String())

	combForms, err := op24.CombinatorialForms()
	require.NoError(t, err)
	combForms[0] = pcseq.FormLabel{}
	freshCombForms, err := op24.CombinatorialForms()
	require.NoError(t, err)
	assert.Equal(t, "P2", freshCombForms[0].String())

	marsden := mustRow(t, 12, 0, 2, 6, 10, 7, 4, 11, 9, 8, 3, 5, 1)
	rotations, err := marsden.CombinatorialRotations()
	require.NoError(t, err)
	rotations[0].Forms[0] = pcseq.FormLabel{}
	freshRotations, err := marsden.CombinatorialRotations()
	require.NoError(t, err)
	assert.Equal(t, "I3", freshRotations[0].Forms[0].String())

	hex, err := op24.HexachordAnalysis()
	require.NoError(t, err)
	hex.Transpositions[0] = 99
	hex.Inversions[0] = 99
	hex.RetrogradeInversions[0] = 99
	freshHex, err := op24.HexachordAnalysis()
	require.NoError(t, err)
	assert.Equal(t, []int{2, 6, 10}, freshHex.Transpositions)
	assert.Equal(t, []int{1, 5, 9}, freshHex.Inversions)
	assert.Equal(t, []int{3, 7, 11}, freshHex.RetrogradeInversions)
}

func TestOwnershipBigIntegers(t *testing.T) {
	first, err := pcseq.RowCountBig(12)
	require.NoError(t, err)
	first.SetInt64(0)
	second, err := pcseq.RowCountBig(12)
	require.NoError(t, err)
	assert.Equal(t, "479001600", second.String())

	lengths, err := pcseq.CountRowsWithSegmentLengthsBig(12, 6, 6)
	require.NoError(t, err)
	lengths.SetInt64(0)
	freshLengths, err := pcseq.CountRowsWithSegmentLengthsBig(12, 6, 6)
	require.NoError(t, err)
	assert.Equal(t, "518400", freshLengths.String())
}

func TestOwnershipComparableValues(t *testing.T) {
	for _, typ := range []reflect.Type{
		reflect.TypeFor[pcseq.Sequence](),
		reflect.TypeFor[pcseq.Row](),
		reflect.TypeFor[pcseq.FormLabel](),
		reflect.TypeFor[pcseq.Matrix](),
		reflect.TypeFor[pcseq.RowClass](),
		reflect.TypeFor[pcseq.DerivationOperator](),
	} {
		assert.True(t, typ.Comparable(), typ.String())
	}
}

func TestConcurrencyIndependentCalls(t *testing.T) {
	row := mustRow(t, 12, 3, 7, 11, 1, 5, 0, 2, 10, 6, 4, 8, 9)
	class, err := pcseq.NewRowClass(row, pcseq.TIR)
	require.NoError(t, err)
	matrix, err := pcseq.NewMatrix(row)
	require.NoError(t, err)
	label := mustLabel(t, pcseq.RI, 7, 12)
	const workers = 16
	const iterations = 40
	var wg sync.WaitGroup
	results := make(chan error, workers)
	for range workers {
		wg.Go(func() {
			for range iterations {
				transposed, err := row.Transpose(5)
				if err != nil {
					results <- err
					return
				}
				form, err := transposed.Form(label)
				if err != nil {
					results <- err
					return
				}
				if !class.Contains(form) {
					results <- assert.AnError
					return
				}
				if _, err := class.Members(); err != nil {
					results <- err
					return
				}
				cell, err := matrix.Cell(3, 8)
				if err != nil || !cell.IsValid() {
					results <- assert.AnError
					return
				}
				if _, err := row.Sequence().Intervals(); err != nil {
					results <- err
					return
				}
			}
			results <- nil
		})
	}
	wg.Wait()
	close(results)
	for err := range results {
		assert.NoError(t, err)
	}
}
