package pcseq_test

import (
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sequenceAllocations(t *testing.T, operation func() (pcseq.Sequence, error)) float64 {
	t.Helper()
	var result pcseq.Sequence
	var err error
	allocations := testing.AllocsPerRun(100, func() {
		result, err = operation()
	})
	require.NoError(t, err)
	require.True(t, result.IsValid())
	require.NotEmpty(t, result.Numbers())
	return allocations
}

func measureSequenceAllocations(t *testing.T, name string, operation func() (pcseq.Sequence, error)) {
	t.Helper()
	assert.LessOrEqual(t, sequenceAllocations(t, operation), 2.0, name)
}

func TestAllocationsBasicOperations(t *testing.T) {
	s := mustSequence(t, 12, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	op := mustOperator(t, 5, 2, 12)
	measureSequenceAllocations(t, "New", func() (pcseq.Sequence, error) {
		return pcseq.New(12, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	})
	measureSequenceAllocations(t, "Transpose", func() (pcseq.Sequence, error) { return s.Transpose(2) })
	measureSequenceAllocations(t, "Invert", func() (pcseq.Sequence, error) { return s.Invert(2) })
	measureSequenceAllocations(t, "Multiply", func() (pcseq.Sequence, error) { return s.Multiply(2) })
	measureSequenceAllocations(t, "Apply", func() (pcseq.Sequence, error) { return s.Apply(op) })
	measureSequenceAllocations(t, "Retrograde", func() (pcseq.Sequence, error) { return s.Retrograde() })
	measureSequenceAllocations(t, "RetrogradeInversion", func() (pcseq.Sequence, error) {
		return s.RetrogradeInversion(2)
	})
	measureSequenceAllocations(t, "Rotate", func() (pcseq.Sequence, error) { return s.Rotate(3) })
	measureSequenceAllocations(t, "Segment", func() (pcseq.Sequence, error) { return s.Segment(1, 10) })
}

func TestAllocationsLengthAndModulusIndependence(t *testing.T) {
	for _, length := range []int{1, 12, 120, 1_200} {
		t.Run(bigLengthName(length), func(t *testing.T) {
			numbers := make([]int, length)
			for i := range length {
				numbers[i] = i % 12
			}
			s := mustSequence(t, 12, numbers...)
			op := mustOperator(t, 5, 2, 12)
			measureSequenceAllocations(t, "New", func() (pcseq.Sequence, error) {
				return pcseq.New(12, numbers...)
			})
			measureSequenceAllocations(t, "Transpose", func() (pcseq.Sequence, error) { return s.Transpose(2) })
			measureSequenceAllocations(t, "Invert", func() (pcseq.Sequence, error) { return s.Invert(2) })
			measureSequenceAllocations(t, "Multiply", func() (pcseq.Sequence, error) { return s.Multiply(2) })
			measureSequenceAllocations(t, "Apply", func() (pcseq.Sequence, error) { return s.Apply(op) })
			measureSequenceAllocations(t, "Retrograde", func() (pcseq.Sequence, error) { return s.Retrograde() })
			measureSequenceAllocations(t, "RetrogradeInversion", func() (pcseq.Sequence, error) {
				return s.RetrogradeInversion(2)
			})
			measureSequenceAllocations(t, "Rotate", func() (pcseq.Sequence, error) { return s.Rotate(3) })
			measureSequenceAllocations(t, "Segment", func() (pcseq.Sequence, error) {
				return s.Segment(0, length)
			})
		})
	}

	for _, tt := range []struct {
		name string
		call func(int, []int, pcseq.Sequence, pitchclass.Operator) (pcseq.Sequence, error)
	}{
		{name: "New", call: func(modulus int, numbers []int, _ pcseq.Sequence, _ pitchclass.Operator) (pcseq.Sequence, error) {
			return pcseq.New(modulus, numbers...)
		}},
		{name: "Transpose", call: func(_ int, _ []int, s pcseq.Sequence, _ pitchclass.Operator) (pcseq.Sequence, error) {
			return s.Transpose(2)
		}},
		{name: "Invert", call: func(_ int, _ []int, s pcseq.Sequence, _ pitchclass.Operator) (pcseq.Sequence, error) {
			return s.Invert(2)
		}},
		{name: "Multiply", call: func(_ int, _ []int, s pcseq.Sequence, _ pitchclass.Operator) (pcseq.Sequence, error) {
			return s.Multiply(2)
		}},
		{name: "Apply", call: func(_ int, _ []int, s pcseq.Sequence, op pitchclass.Operator) (pcseq.Sequence, error) {
			return s.Apply(op)
		}},
		{name: "Retrograde", call: func(_ int, _ []int, s pcseq.Sequence, _ pitchclass.Operator) (pcseq.Sequence, error) {
			return s.Retrograde()
		}},
		{name: "RetrogradeInversion", call: func(_ int, _ []int, s pcseq.Sequence, _ pitchclass.Operator) (pcseq.Sequence, error) {
			return s.RetrogradeInversion(2)
		}},
		{name: "Rotate", call: func(_ int, _ []int, s pcseq.Sequence, _ pitchclass.Operator) (pcseq.Sequence, error) {
			return s.Rotate(3)
		}},
		{name: "Segment", call: func(_ int, _ []int, s pcseq.Sequence, _ pitchclass.Operator) (pcseq.Sequence, error) {
			return s.Segment(0, s.Len())
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			numbers := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
			small := mustSequence(t, 12, numbers...)
			large := mustSequence(t, 65_536, numbers...)
			smallOperator := mustOperator(t, 5, 2, 12)
			largeOperator := mustOperator(t, 5, 2, 65_536)
			smallAllocations := sequenceAllocations(t, func() (pcseq.Sequence, error) {
				return tt.call(12, numbers, small, smallOperator)
			})
			largeAllocations := sequenceAllocations(t, func() (pcseq.Sequence, error) {
				return tt.call(65_536, numbers, large, largeOperator)
			})
			assert.LessOrEqual(t, smallAllocations, 2.0)
			assert.LessOrEqual(t, largeAllocations, smallAllocations)
		})
	}
}

func bigLengthName(length int) string {
	switch length {
	case 1:
		return "L1"
	case 12:
		return "L12"
	case 120:
		return "L120"
	case 1_200:
		return "L1200"
	default:
		return "L"
	}
}

func TestThresholdsMatrixCells(t *testing.T) {
	row := mustRow(t, 1_024, identityNumbers(1_024)...)
	matrix, err := pcseq.NewMatrix(row)
	require.NoError(t, err)
	cells, err := matrix.Cells()
	require.NoError(t, err)
	require.Len(t, cells, 1_024)
	count := 0
	for _, row := range cells {
		require.Len(t, row, 1_024)
		count += len(row)
	}
	assert.Equal(t, 1_048_576, count)
}

func TestThresholdsRowClassesAndTTO(t *testing.T) {
	row1025 := mustRow(t, 1_025, identityNumbers(1_025)...)
	class, err := pcseq.NewRowClass(row1025, pcseq.TIR)
	require.NoError(t, err)
	assert.Positive(t, class.Size())
	assert.True(t, class.Contains(row1025))
	members, err := class.Members()
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
	assert.Nil(t, members)
	_, err = pcseq.NewRowClass(row1025, pcseq.TIRRotations)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)

	row257 := mustRow(t, 257, identityNumbers(257)...)
	applied, err := row257.Apply(mustOperator(t, 5, 2, 257))
	require.NoError(t, err)
	assert.Equal(t, 257, applied.Len())
	_, err = pcseq.NewRowClass(row257, pcseq.TTOR)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
}

func TestThresholdsAnalysis(t *testing.T) {
	row := mustRow(t, 1_025, identityNumbers(1_025)...)
	forms, err := row.CombinatorialForms()
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
	assert.Nil(t, forms)
	rotations, err := row.CombinatorialRotations()
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
	assert.Nil(t, rotations)
}
