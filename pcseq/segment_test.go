package pcseq_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/setclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSegmentSingleAndCollections(t *testing.T) {
	t.Parallel()
	op28 := mustSequence(t, 12, 0, 11, 2, 1, 5, 6, 3, 4, 8, 7, 10, 9)

	segment, err := op28.Segment(4, 4)
	require.NoError(t, err)
	assert.Equal(t, []int{5, 6, 3, 4}, segment.Numbers())
	assert.Equal(t, 12, segment.Modulus())
	assert.False(t, segment.IsRow())

	for _, tt := range []struct {
		name   string
		start  int
		length int
	}{
		{name: "zero at start", start: 0, length: 0},
		{name: "zero inside", start: 4, length: 0},
		{name: "zero at end", start: 12, length: 0},
		{name: "full sequence", start: 0, length: 12},
		{name: "last element", start: 11, length: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := op28.Segment(tt.start, tt.length)
			require.NoError(t, err)
			assert.Equal(t, op28.Numbers()[tt.start:tt.start+tt.length], got.Numbers())
		})
	}

	for _, tt := range []struct {
		name   string
		start  int
		length int
	}{
		{name: "negative start", start: -1, length: 1},
		{name: "negative length", start: 0, length: -1},
		{name: "past end", start: 12, length: 1},
		{name: "length above sequence", start: 0, length: 13},
		{name: "sum above sequence", start: 11, length: 2},
		{name: "overflowing sum", start: math.MaxInt, length: math.MaxInt},
		{name: "maximum length", start: 0, length: math.MaxInt},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := op28.Segment(tt.start, tt.length)
			assert.ErrorIs(t, err, pcseq.ErrInvalidSegment)
			assert.Equal(t, pcseq.Sequence{}, got)
		})
	}

	overlapping, err := op28.OverlappingSegments(3)
	require.NoError(t, err)
	require.Len(t, overlapping, 10)
	for i, got := range overlapping {
		assert.Equal(t, op28.Numbers()[i:i+3], got.Numbers())
	}
	full, err := op28.OverlappingSegments(12)
	require.NoError(t, err)
	require.Len(t, full, 1)
	assert.Equal(t, op28, full[0])
	for _, length := range []int{-1, 0, 13, math.MaxInt} {
		parts, err := op28.OverlappingSegments(length)
		assert.ErrorIs(t, err, pcseq.ErrInvalidSegment)
		assert.Nil(t, parts)
	}

	discrete, err := op28.DiscreteSegments(3)
	require.NoError(t, err)
	require.Len(t, discrete, 4)
	for i, got := range discrete {
		assert.Equal(t, op28.Numbers()[3*i:3*i+3], got.Numbers())
	}
	emptyParts, err := mustSequence(t, 12).DiscreteSegments(7)
	require.NoError(t, err)
	assert.NotNil(t, emptyParts)
	assert.Empty(t, emptyParts)
	for _, tt := range []struct {
		name        string
		sequenceLen int
		length      int
	}{
		{name: "zero", sequenceLen: 12, length: 0},
		{name: "negative", sequenceLen: 12, length: -3},
		{name: "not divisor", sequenceLen: 7, length: 3},
		{name: "above length", sequenceLen: 7, length: 8},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := mustSequence(t, 12, identityNumbers(tt.sequenceLen)...)
			parts, err := s.DiscreteSegments(tt.length)
			assert.ErrorIs(t, err, pcseq.ErrInvalidPartition)
			assert.Nil(t, parts)
		})
	}

	_, err = pcseq.Sequence{}.Segment(0, 0)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.Sequence{}.OverlappingSegments(1)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.Sequence{}.DiscreteSegments(1)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
}

func TestSegmentContentClasses(t *testing.T) {
	t.Parallel()
	segment := mustSequence(t, 12, 0, 1, 4, 5, 8, 9)
	for _, kind := range []setclass.Kind{setclass.Tn, setclass.TnIn, setclass.TTO} {
		t.Run(kind.String(), func(t *testing.T) {
			t.Parallel()
			class, err := segment.ContentClass(kind)
			require.NoError(t, err)
			assert.True(t, class.IsValid())
			assert.Equal(t, kind, class.Kind())
			content, err := segment.Content()
			require.NoError(t, err)
			assert.True(t, class.Contains(content))
		})
	}
	_, err := segment.ContentClass(setclass.Kind(0))
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.Sequence{}.ContentClass(setclass.Tn)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)

	large := mustSequence(t, 257, 0, 1, 2)
	_, err = large.ContentClass(setclass.TTO)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
	_, err = large.ContentClass(setclass.TnIn)
	assert.NoError(t, err)
}

func TestContentInvariances(t *testing.T) {
	t.Parallel()
	op28 := mustSequence(t, 12, 0, 11, 2, 1, 5, 6, 3, 4, 8, 7, 10, 9)
	parts, err := op28.DiscreteSegments(4)
	require.NoError(t, err)
	require.Len(t, parts, 3)

	invariance, err := parts[0].InvariancesTo(parts[1])
	require.NoError(t, err)
	assert.Equal(t, []int{4}, invariance.Transpositions)
	assert.Equal(t, []int{5}, invariance.Inversions)

	transposed, err := op28.Transpose(4)
	require.NoError(t, err)
	transformedParts, err := transposed.DiscreteSegments(4)
	require.NoError(t, err)
	for i := range parts {
		self, err := parts[i].InvariancesTo(parts[i])
		require.NoError(t, err)
		assert.Contains(t, self.Transpositions, 4)
		contentBefore, err := parts[i].Content()
		require.NoError(t, err)
		contentAfter, err := transformedParts[i].Content()
		require.NoError(t, err)
		assert.True(t, contentBefore.Equal(contentAfter))
	}

	unrelated := mustSequence(t, 12, 0, 1, 2, 4)
	none, err := parts[0].InvariancesTo(unrelated)
	require.NoError(t, err)
	assert.NotNil(t, none.Transpositions)
	assert.NotNil(t, none.Inversions)
	assert.Empty(t, none.Transpositions)
	assert.Empty(t, none.Inversions)

	emptyA := mustSequence(t, 12)
	emptyB := mustSequence(t, 12)
	all, err := emptyA.InvariancesTo(emptyB)
	require.NoError(t, err)
	assert.Equal(t, identityNumbers(12), all.Transpositions)
	assert.Equal(t, identityNumbers(12), all.Inversions)

	_, err = parts[0].InvariancesTo(mustSequence(t, 12, 0))
	assert.ErrorIs(t, err, pcseq.ErrInvalidSegment)
	_, err = parts[0].InvariancesTo(mustSequence(t, 19, 0, 1, 2, 3))
	assert.ErrorIs(t, err, pcseq.ErrModulusMismatch)
	_, err = pcseq.Sequence{}.InvariancesTo(parts[0])
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = parts[0].InvariancesTo(pcseq.Sequence{})
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
}

func TestSegmentMustVariants(t *testing.T) {
	t.Parallel()
	s := mustSequence(t, 12, 0, 1, 2)
	_, err := s.Segment(1, 3)
	assert.PanicsWithValue(t, err, func() { s.MustSegment(1, 3) })
	assert.Equal(t, []int{1, 2}, s.MustSegment(1, 2).Numbers())
	_, err = s.OverlappingSegments(0)
	assert.PanicsWithValue(t, err, func() { s.MustOverlappingSegments(0) })
	assert.Len(t, s.MustOverlappingSegments(2), 2)
	_, err = s.DiscreteSegments(2)
	assert.PanicsWithValue(t, err, func() { s.MustDiscreteSegments(2) })
	assert.Len(t, s.MustDiscreteSegments(3), 1)
	_, err = s.ContentClass(setclass.Kind(0))
	assert.PanicsWithValue(t, err, func() { s.MustContentClass(setclass.Kind(0)) })
	_, err = s.InvariancesTo(mustSequence(t, 12, 0))
	assert.PanicsWithValue(t, err, func() { s.MustInvariancesTo(mustSequence(t, 12, 0)) })
}
