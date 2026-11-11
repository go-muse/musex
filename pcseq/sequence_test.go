package pcseq_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSequenceCoreCreationAndReading(t *testing.T) {
	t.Parallel()

	t.Run("strict construction preserves order and repetitions", func(t *testing.T) {
		t.Parallel()
		s := mustSequence(t, 12, 7, 11, 9, 6, 7)
		assert.True(t, s.IsValid())
		assert.False(t, s.IsEmpty())
		assert.False(t, s.IsRow())
		assert.Equal(t, 12, s.Modulus())
		assert.Equal(t, 5, s.Len())
		assert.Equal(t, []int{7, 11, 9, 6, 7}, s.Numbers())
		assert.Equal(t, []int{7, 11, 9, 6, 7}, classNumbers(s.Classes()))
		at, err := s.At(1)
		require.NoError(t, err)
		assert.Equal(t, mustClass(t, 11, 12), at)
		assert.Equal(t, "<7, 11, 9, 6, 7> mod 12", s.String())
	})

	t.Run("empty is valid and distinct from zero", func(t *testing.T) {
		t.Parallel()
		for _, create := range []struct {
			name string
			call func() (pcseq.Sequence, error)
		}{
			{name: "Empty", call: func() (pcseq.Sequence, error) { return pcseq.Empty(12) }},
			{name: "New without numbers", call: func() (pcseq.Sequence, error) { return pcseq.New(12) }},
			{name: "Reduce without numbers", call: func() (pcseq.Sequence, error) { return pcseq.Reduce(12) }},
			{name: "FromClassesWithModulus", call: func() (pcseq.Sequence, error) {
				return pcseq.FromClassesWithModulus(12)
			}},
		} {
			t.Run(create.name, func(t *testing.T) {
				t.Parallel()
				s, err := create.call()
				require.NoError(t, err)
				assert.True(t, s.IsValid())
				assert.True(t, s.IsEmpty())
				assert.Equal(t, 12, s.Modulus())
				assert.Zero(t, s.Len())
				assert.NotNil(t, s.Numbers())
				assert.Empty(t, s.Numbers())
				assert.NotNil(t, s.Classes())
				assert.Empty(t, s.Classes())
				assert.False(t, s.IsRow())
				assert.NotEqual(t, pcseq.Sequence{}, s)
			})
		}
		var zero pcseq.Sequence
		assert.False(t, zero.IsValid())
		assert.False(t, zero.IsEmpty())
		assert.False(t, zero.IsRow())
		assert.Zero(t, zero.Modulus())
		assert.Zero(t, zero.Len())
		assert.Nil(t, zero.Numbers())
		assert.Nil(t, zero.Classes())
		assert.Equal(t, "Sequence(invalid)", zero.String())
	})

	t.Run("reduction handles negative and extreme integers", func(t *testing.T) {
		t.Parallel()
		s := mustReduceSequence(t, 12, -1, 13, math.MinInt, math.MaxInt)
		assert.Equal(t, []int{11, 1, 4, 7}, s.Numbers())
	})

	t.Run("class constructors infer and check modulus", func(t *testing.T) {
		t.Parallel()
		classes := []pitchclass.Class{mustClass(t, 7, 12), mustClass(t, 11, 12), mustClass(t, 7, 12)}
		s, err := pcseq.FromClasses(classes...)
		require.NoError(t, err)
		assert.Equal(t, []int{7, 11, 7}, s.Numbers())
		classes[0] = mustClass(t, 0, 12)
		assert.Equal(t, []int{7, 11, 7}, s.Numbers(), "constructor must own its result")
		explicit, err := pcseq.FromClassesWithModulus(12, mustClass(t, 3, 12), mustClass(t, 4, 12))
		require.NoError(t, err)
		assert.Equal(t, []int{3, 4}, explicit.Numbers())
	})

	t.Run("constructor errors are classified", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name   string
			call   func() (pcseq.Sequence, error)
			target error
		}{
			{name: "zero modulus", call: func() (pcseq.Sequence, error) { return pcseq.New(0, 0) }, target: pcseq.ErrInvalidModulus},
			{name: "negative modulus", call: func() (pcseq.Sequence, error) { return pcseq.Reduce(-1, 0) }, target: pcseq.ErrInvalidModulus},
			{name: "modulus above maximum", call: func() (pcseq.Sequence, error) { return pcseq.New(65_537) }, target: pcseq.ErrModulusTooLarge},
			{name: "strict high number", call: func() (pcseq.Sequence, error) { return pcseq.New(12, 12) }, target: pcseq.ErrInvalidNumber},
			{name: "strict negative number", call: func() (pcseq.Sequence, error) { return pcseq.New(12, -1) }, target: pcseq.ErrInvalidNumber},
			{name: "classes without modulus", call: func() (pcseq.Sequence, error) { return pcseq.FromClasses() }, target: pcseq.ErrInvalidValue},
			{name: "invalid class", call: func() (pcseq.Sequence, error) {
				return pcseq.FromClasses(pitchclass.Class{})
			}, target: pcseq.ErrInvalidValue},
			{name: "mixed moduli", call: func() (pcseq.Sequence, error) {
				return pcseq.FromClasses(mustClass(t, 0, 12), mustClass(t, 0, 19))
			}, target: pcseq.ErrModulusMismatch},
			{name: "class outside explicit modulus", call: func() (pcseq.Sequence, error) {
				return pcseq.FromClassesWithModulus(12, mustClass(t, 0, 19))
			}, target: pcseq.ErrModulusMismatch},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				s, err := tt.call()
				assert.ErrorIs(t, err, tt.target)
				assert.Equal(t, pcseq.Sequence{}, s)
			})
		}
	})

	t.Run("position errors", func(t *testing.T) {
		t.Parallel()
		s := mustSequence(t, 12, 7)
		for _, index := range []int{-1, 1, math.MaxInt} {
			_, err := s.At(index)
			assert.ErrorIs(t, err, pcseq.ErrInvalidSegment)
		}
		_, err := pcseq.Sequence{}.At(0)
		assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	})
}

func TestSequenceCoreComparisonAndMapKeys(t *testing.T) {
	t.Parallel()

	a := mustSequence(t, 12, 7, 11, 9, 6, 7)
	same := mustSequence(t, 12, 7, 11, 9, 6, 7)
	differentOrder := mustSequence(t, 12, 11, 7, 9, 6, 7)
	differentModulus := mustSequence(t, 19, 7, 11, 9, 6, 7)
	shorter := mustSequence(t, 12, 7, 11)
	lexicallyGreater := mustSequence(t, 12, 7, 11, 9, 6, 8)

	assert.True(t, a.Equal(same))
	assert.Equal(t, a, same)
	assert.Zero(t, a.Compare(same))
	assert.False(t, a.Equal(differentOrder))
	assert.False(t, a.Equal(differentModulus))
	assert.Negative(t, a.Compare(differentModulus))
	assert.Positive(t, differentModulus.Compare(a))
	assert.Positive(t, shorter.Compare(a))
	assert.Negative(t, a.Compare(lexicallyGreater))

	keys := map[pcseq.Sequence]int{a: 1}
	keys[same] = 2
	assert.Len(t, keys, 1)
	assert.Equal(t, 2, keys[a])
	assert.True(t, reflect.TypeFor[pcseq.Sequence]().Comparable())

	var zero pcseq.Sequence
	assert.True(t, zero == pcseq.Sequence{})
	assert.False(t, zero.Equal(pcseq.Sequence{}))
	assert.Zero(t, zero.Compare(pcseq.Sequence{}))
	assert.Negative(t, zero.Compare(a))
	assert.Positive(t, a.Compare(zero))
}

func TestRowCoreCreationAndReading(t *testing.T) {
	t.Parallel()

	mt21 := []int{3, 7, 11, 1, 5, 0, 2, 10, 6, 4, 8, 9}
	r := mustRow(t, 12, mt21...)
	assert.True(t, r.IsValid())
	assert.Equal(t, 12, r.Modulus())
	assert.Equal(t, 12, r.Len())
	assert.Equal(t, mt21, r.Numbers())
	assert.Equal(t, mt21, classNumbers(r.Classes()))
	assert.True(t, r.Sequence().IsRow())
	assert.Equal(t, r.Numbers(), r.Sequence().Numbers())
	at, err := r.At(0)
	require.NoError(t, err)
	assert.Equal(t, mustClass(t, 3, 12), at)

	for _, tt := range []struct {
		name    string
		modulus int
		numbers []int
	}{
		{name: "one", modulus: 1, numbers: []int{0}},
		{name: "two ascending", modulus: 2, numbers: []int{0, 1}},
		{name: "two descending", modulus: 2, numbers: []int{1, 0}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			row := mustRow(t, tt.modulus, tt.numbers...)
			assert.True(t, row.IsValid())
			assert.Equal(t, tt.numbers, row.Numbers())
		})
	}

	reduced, err := pcseq.ReduceRow(12, 15, -5, 23, 1, 5, 0, 2, 10, 6, 4, 8, 9)
	require.NoError(t, err)
	assert.Equal(t, mt21, reduced.Numbers())

	fromClasses, err := pcseq.RowFromClasses(mustClass(t, 0, 2), mustClass(t, 1, 2))
	require.NoError(t, err)
	assert.Equal(t, []int{0, 1}, fromClasses.Numbers())
	fromExplicit, err := pcseq.RowFromClassesWithModulus(2, mustClass(t, 1, 2), mustClass(t, 0, 2))
	require.NoError(t, err)
	assert.Equal(t, []int{1, 0}, fromExplicit.Numbers())

	fromSequence, err := pcseq.RowFromSequence(r.Sequence())
	require.NoError(t, err)
	assert.Equal(t, r, fromSequence)

	for _, tt := range []struct {
		name    string
		numbers []int
	}{
		{name: "wrong length", numbers: []int{0, 1, 2}},
		{name: "repetition and missing class", numbers: []int{7, 11, 9, 6, 7}},
		{name: "empty", numbers: nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			row, err := pcseq.NewRow(12, tt.numbers...)
			assert.ErrorIs(t, err, pcseq.ErrNotRow)
			assert.Equal(t, pcseq.Row{}, row)
			_, err = pcseq.RowFromSequence(mustSequence(t, 12, tt.numbers...))
			assert.ErrorIs(t, err, pcseq.ErrNotRow)
		})
	}

	_, err = pcseq.RowFromSequence(pcseq.Sequence{})
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.NewRow(12, 12)
	assert.ErrorIs(t, err, pcseq.ErrInvalidNumber)
	_, err = pcseq.NewRow(0)
	assert.ErrorIs(t, err, pcseq.ErrInvalidModulus)
	_, err = pcseq.NewRow(65_537)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
	_, err = pcseq.RowFromClasses()
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.RowFromClasses(mustClass(t, 0, 12), mustClass(t, 1, 19))
	assert.ErrorIs(t, err, pcseq.ErrModulusMismatch)

	var zero pcseq.Row
	assert.False(t, zero.IsValid())
	assert.Zero(t, zero.Modulus())
	assert.Zero(t, zero.Len())
	assert.Nil(t, zero.Numbers())
	assert.Nil(t, zero.Classes())
	assert.False(t, zero.Sequence().IsValid())
	assert.Equal(t, "Row(invalid)", zero.String())
	_, err = zero.At(0)
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
}

func TestRowCoreComparisonAndMapKeys(t *testing.T) {
	t.Parallel()
	a := mustRow(t, 3, 1, 0, 2)
	same := mustRow(t, 3, 1, 0, 2)
	other := mustRow(t, 3, 0, 1, 2)
	assert.True(t, a.Equal(same))
	assert.Zero(t, a.Compare(same))
	assert.False(t, a.Equal(other))
	assert.Positive(t, a.Compare(other))
	keys := map[pcseq.Row]string{a: "row"}
	keys[same] = "same"
	assert.Len(t, keys, 1)
	assert.Equal(t, "same", keys[a])
	assert.True(t, reflect.TypeFor[pcseq.Row]().Comparable())
	var zero pcseq.Row
	assert.True(t, zero == pcseq.Row{})
	assert.False(t, zero.Equal(pcseq.Row{}))
	assert.Zero(t, zero.Compare(pcseq.Row{}))
	assert.Negative(t, zero.Compare(a))
}

func TestTransformSequenceOperations(t *testing.T) {
	t.Parallel()
	s := mustSequence(t, 12, 7, 11, 9, 6, 3, 5, 8)

	tests := []struct {
		name     string
		call     func() (pcseq.Sequence, error)
		expected []int
	}{
		{name: "transpose", call: func() (pcseq.Sequence, error) { return s.Transpose(2) }, expected: []int{9, 1, 11, 8, 5, 7, 10, 10}},
		{name: "negative transpose", call: func() (pcseq.Sequence, error) { return s.Transpose(-1) }, expected: []int{6, 10, 8, 5, 2, 4, 7, 7}},
		{name: "invert two", call: func() (pcseq.Sequence, error) { return s.Invert(2) }, expected: []int{7, 3, 5, 8, 11, 9, 6, 6}},
		{name: "multiply two", call: func() (pcseq.Sequence, error) { return s.Multiply(2) }, expected: []int{2, 10, 6, 0, 6, 10, 4, 4}},
		{name: "apply affine", call: func() (pcseq.Sequence, error) {
			return s.Apply(mustOperator(t, 5, 2, 12))
		}, expected: []int{1, 9, 11, 8, 5, 3, 6, 6}},
		{name: "retrograde", call: func() (pcseq.Sequence, error) { return s.Retrograde() }, expected: []int{8, 5, 3, 6, 9, 11, 7}},
		{name: "retrograde inversion", call: func() (pcseq.Sequence, error) { return s.RetrogradeInversion(2) }, expected: []int{6, 9, 11, 8, 5, 3, 7}},
		{name: "minimum transpose", call: func() (pcseq.Sequence, error) { return s.Transpose(math.MinInt) }, expected: []int{11, 3, 1, 10, 7, 9, 0, 0}},
		{name: "maximum invert", call: func() (pcseq.Sequence, error) { return s.Invert(math.MaxInt) }, expected: []int{0, 8, 10, 1, 4, 2, 11, 11}},
		{name: "minimum multiply", call: func() (pcseq.Sequence, error) { return s.Multiply(math.MinInt) }, expected: []int{4, 8, 0, 0, 0, 8, 8}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.call()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got.Numbers())
			assert.Equal(t, s.Numbers(), mustSequence(t, 12, 7, 11, 9, 6, 3, 5, 8).Numbers())
		})
	}

	_, err := s.Apply(mustOperator(t, 1, 0, 19))
	assert.ErrorIs(t, err, pcseq.ErrModulusMismatch)
	_, err = s.Apply(pitchclass.Operator{})
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)

	empty := mustSequence(t, 12)
	for name, call := range map[string]func() (pcseq.Sequence, error){
		"transpose": func() (pcseq.Sequence, error) { return empty.Transpose(math.MaxInt) },
		"invert":    func() (pcseq.Sequence, error) { return empty.Invert(math.MinInt) },
		"multiply":  func() (pcseq.Sequence, error) { return empty.Multiply(math.MinInt) },
		"apply":     func() (pcseq.Sequence, error) { return empty.Apply(mustOperator(t, 5, 0, 12)) },
		"retrograde": func() (pcseq.Sequence, error) {
			return empty.Retrograde()
		},
		"retrograde inversion": func() (pcseq.Sequence, error) { return empty.RetrogradeInversion(2) },
	} {
		t.Run("empty "+name, func(t *testing.T) {
			t.Parallel()
			got, err := call()
			require.NoError(t, err)
			assert.True(t, got.IsValid())
			assert.True(t, got.IsEmpty())
		})
	}

	for name, call := range map[string]func() (pcseq.Sequence, error){
		"transpose": func() (pcseq.Sequence, error) { return pcseq.Sequence{}.Transpose(1) },
		"invert":    func() (pcseq.Sequence, error) { return pcseq.Sequence{}.Invert(1) },
		"multiply":  func() (pcseq.Sequence, error) { return pcseq.Sequence{}.Multiply(2) },
		"apply":     func() (pcseq.Sequence, error) { return pcseq.Sequence{}.Apply(mustOperator(t, 1, 0, 12)) },
		"retrograde": func() (pcseq.Sequence, error) {
			return pcseq.Sequence{}.Retrograde()
		},
		"retrograde inversion": func() (pcseq.Sequence, error) {
			return pcseq.Sequence{}.RetrogradeInversion(2)
		},
		"rotate": func() (pcseq.Sequence, error) { return pcseq.Sequence{}.Rotate(1) },
	} {
		t.Run("zero "+name, func(t *testing.T) {
			t.Parallel()
			_, err := call()
			assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
		})
	}
}

func TestTransformRowOperations(t *testing.T) {
	t.Parallel()
	r := mustRow(t, 12, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)

	transposed, err := r.Transpose(1)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 0}, transposed.Numbers())
	inverted, err := r.Invert(0)
	require.NoError(t, err)
	assert.Equal(t, []int{0, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}, inverted.Numbers())
	applied, err := r.Apply(mustOperator(t, 5, 0, 12))
	require.NoError(t, err)
	assert.Equal(t, []int{0, 5, 10, 3, 8, 1, 6, 11, 4, 9, 2, 7}, applied.Numbers())
	retro, err := r.Retrograde()
	require.NoError(t, err)
	assert.Equal(t, []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}, retro.Numbers())
	ri, err := r.RetrogradeInversion(0)
	require.NoError(t, err)
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, ri.Numbers())

	multiplied, err := r.Multiply(2)
	require.NoError(t, err)
	assert.Equal(t, []int{0, 2, 4, 6, 8, 10, 0, 2, 4, 6, 8, 10}, multiplied.Numbers())
	assert.False(t, multiplied.IsRow())
	_, err = pcseq.RowFromSequence(multiplied)
	assert.ErrorIs(t, err, pcseq.ErrNotRow)

	invertible, err := r.Multiply(5)
	require.NoError(t, err)
	assert.True(t, invertible.IsRow())
	_, err = pcseq.RowFromSequence(invertible)
	assert.NoError(t, err)
}

func TestTransformRotationAlgebra(t *testing.T) {
	t.Parallel()
	s := mustSequence(t, 12, 0, 2, 6, 10, 7, 4, 11, 9, 8, 3, 5, 1)

	rotated, err := s.Rotate(1)
	require.NoError(t, err)
	assert.Equal(t, []int{2, 6, 10, 7, 4, 11, 9, 8, 3, 5, 1, 0}, rotated.Numbers())
	identity, err := s.Rotate(12)
	require.NoError(t, err)
	assert.Equal(t, s, identity)
	negative, err := s.Rotate(-1)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 0, 2, 6, 10, 7, 4, 11, 9, 8, 3, 5}, negative.Numbers())
	minimum, err := s.Rotate(math.MinInt)
	require.NoError(t, err)
	expectedMinimum, err := s.Rotate(4)
	require.NoError(t, err)
	assert.Equal(t, expectedMinimum, minimum)
	maximum, err := s.Rotate(math.MaxInt)
	require.NoError(t, err)
	expectedMaximum, err := s.Rotate(7)
	require.NoError(t, err)
	assert.Equal(t, expectedMaximum, maximum)

	for _, tt := range []struct{ j, k int }{{2, 5}, {-3, 7}, {math.MinInt, math.MaxInt}} {
		left, err := s.Rotate(tt.j)
		require.NoError(t, err)
		left, err = left.Rotate(tt.k)
		require.NoError(t, err)
		right, err := s.Rotate(tt.j + tt.k)
		if tt.j == math.MinInt {
			right, err = s.Rotate(-1)
		}
		require.NoError(t, err)
		assert.Equal(t, right, left)
	}

	rr, err := s.Retrograde()
	require.NoError(t, err)
	rr, err = rr.Retrograde()
	require.NoError(t, err)
	assert.Equal(t, s, rr)

	for _, k := range []int{1, 5, -2, math.MinInt, math.MaxInt} {
		rOfRotation, err := s.Rotate(k)
		require.NoError(t, err)
		rOfRotation, err = rOfRotation.Retrograde()
		require.NoError(t, err)
		rOfRotation, err = rOfRotation.Apply(mustOperator(t, 5, 2, 12))
		require.NoError(t, err)

		applied, err := s.Apply(mustOperator(t, 5, 2, 12))
		require.NoError(t, err)
		applied, err = applied.Retrograde()
		require.NoError(t, err)
		applied, err = applied.Rotate(-k)
		if k == math.MinInt {
			applied, err = applied.Rotate(math.MinInt)
		}
		require.NoError(t, err)
		assert.Equal(t, applied, rOfRotation)
	}

	_, err = mustSequence(t, 12).Rotate(0)
	assert.ErrorIs(t, err, pcseq.ErrInvalidSegment)
	one := mustSequence(t, 12, 7)
	oneRotated, err := one.Rotate(math.MinInt)
	require.NoError(t, err)
	assert.Equal(t, one, oneRotated)
}

func TestContentIntervalsAndSet(t *testing.T) {
	t.Parallel()
	s := mustSequence(t, 12, 7, 11, 9, 6, 3, 5, 8)
	intervals, err := s.Intervals()
	require.NoError(t, err)
	assert.Equal(t, []int{4, 10, 9, 9, 2, 3}, classNumbers(intervals))

	transposed, err := s.Transpose(5)
	require.NoError(t, err)
	transposedIntervals, err := transposed.Intervals()
	require.NoError(t, err)
	assert.Equal(t, classNumbers(intervals), classNumbers(transposedIntervals))

	inverted, err := s.Invert(2)
	require.NoError(t, err)
	invertedIntervals, err := inverted.Intervals()
	require.NoError(t, err)
	assert.Equal(t, []int{8, 2, 3, 3, 10, 9}, classNumbers(invertedIntervals))

	retro, err := s.Retrograde()
	require.NoError(t, err)
	retroIntervals, err := retro.Intervals()
	require.NoError(t, err)
	assert.Equal(t, []int{9, 10, 3, 3, 2, 8}, classNumbers(retroIntervals))

	ri, err := s.RetrogradeInversion(2)
	require.NoError(t, err)
	riIntervals, err := ri.Intervals()
	require.NoError(t, err)
	assert.Equal(t, []int{3, 2, 9, 9, 10, 4}, classNumbers(riIntervals))

	content, err := mustSequence(t, 12, 7, 11, 9, 6, 7).Content()
	require.NoError(t, err)
	assert.Equal(t, []int{6, 7, 9, 11}, content.Numbers())
	emptyContent, err := mustSequence(t, 12).Content()
	require.NoError(t, err)
	assert.True(t, emptyContent.IsValid())
	assert.True(t, emptyContent.IsEmpty())
	assert.Equal(t, 12, emptyContent.Modulus())
	assert.NotNil(t, emptyContent.Numbers())

	emptyIntervals, err := mustSequence(t, 12).Intervals()
	require.NoError(t, err)
	assert.NotNil(t, emptyIntervals)
	assert.Empty(t, emptyIntervals)
	oneIntervals, err := mustSequence(t, 12, 7).Intervals()
	require.NoError(t, err)
	assert.Empty(t, oneIntervals)

	_, err = pcseq.Sequence{}.Intervals()
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.Sequence{}.Content()
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
}

func TestSequenceCoreErrorCatalogAndMust(t *testing.T) {
	t.Parallel()
	sentinels := []error{
		pcseq.ErrInvalidModulus,
		pcseq.ErrModulusTooLarge,
		pcseq.ErrInvalidValue,
		pcseq.ErrModulusMismatch,
		pcseq.ErrInvalidNumber,
		pcseq.ErrNotRow,
		pcseq.ErrInvalidSegment,
		pcseq.ErrInvalidPartition,
		pcseq.ErrInvalidFormLabel,
		pcseq.ErrNotDerived,
		pcseq.ErrNotation,
		pcseq.ErrOverflow,
		pcseq.ErrNotModulus12,
		pcseq.ErrInvalidSegmentContent,
	}
	for i, target := range sentinels {
		for j, err := range sentinels {
			if i != j {
				assert.NotErrorIs(t, err, target)
			}
		}
	}

	_, baseErr := pcseq.New(12, 12)
	assert.PanicsWithValue(t, baseErr, func() { pcseq.MustNew(12, 12) })
	assert.Equal(t, []int{0, 1}, pcseq.MustNew(12, 0, 1).Numbers())
	_, baseErr = pcseq.NewRow(12, 0, 0)
	assert.PanicsWithValue(t, baseErr, func() { pcseq.MustNewRow(12, 0, 0) })
	_, baseErr = pcseq.RowFromSequence(mustSequence(t, 12, 0, 0))
	assert.PanicsWithValue(t, baseErr, func() {
		pcseq.MustRowFromSequence(mustSequence(t, 12, 0, 0))
	})
}
