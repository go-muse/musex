package frequency_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartial_ControlExamples(t *testing.T) {
	for _, n := range []int{1, 2, 3, 4} {
		t.Run("partial", func(t *testing.T) {
			t.Parallel()
			want, ok := modelPartial(110, n)
			require.True(t, ok)
			got, err := frequency.Partial(110, n)
			require.NoError(t, err)
			assert.Equal(t, want, got)
			assert.Equal(t, want, frequency.MustPartial(110, n))
		})
	}
	got, err := frequency.Partials(110, []int{2, 3, 4})
	require.NoError(t, err)
	assert.Equal(t, []float64{220, 330, 440}, got)
	assert.Equal(t, []float64{220, 330, 440}, frequency.MustPartials(110, []int{2, 3, 4}))
}

func TestPartial_ErrorsAndPrecedence(t *testing.T) {
	for _, invalid := range invalidFrequencyCases() {
		t.Run(invalid.name+" before index", func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Partial(invalid.f, 0)
			assert.Zero(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
	}
	for _, n := range []int{0, -1} {
		got, err := frequency.Partial(110, n)
		assert.Zero(t, got)
		requireErrorChain(t, err, frequency.ErrInvalidParameter, nil)
	}
	for _, tt := range []struct {
		name string
		f    float64
		n    int
	}{
		{"K-44 overflow", math.MaxFloat64, 2},
		{"subnormal fundamental", math.SmallestNonzeroFloat64, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Partial(tt.f, tt.n)
			assert.Zero(t, got)
			requireErrorChain(t, err, frequency.ErrOutOfRange, cents.ErrOutOfRange)
		})
	}
}

func TestPartials_OrderOwnershipAndAtomicity(t *testing.T) {
	indices := []int{3, 1, 2, 4}
	original := append([]int(nil), indices...)
	got, err := frequency.Partials(110, indices)
	require.NoError(t, err)
	require.Len(t, got, len(indices))
	assert.Equal(t, len(indices), cap(got))
	assert.Equal(t, original, indices)
	for i, n := range indices {
		want, ok := modelPartial(110, n)
		require.True(t, ok)
		assert.Equal(t, want, got[i])
	}
	got[0] = 0
	fresh, err := frequency.Partials(110, indices)
	require.NoError(t, err)
	assert.Equal(t, 330.0, fresh[0])
	assert.NotSame(t, &got[0], &fresh[0])

	for _, invalid := range [][]int{nil, {}, {0}, {1, -2}} {
		got, err := frequency.Partials(110, invalid)
		assert.Nil(t, got)
		requireErrorChain(t, err, frequency.ErrInvalidParameter, nil)
	}

	invalidIndices := []int{2, 0}
	originalInvalid := append([]int(nil), invalidIndices...)
	got, err = frequency.Partials(math.MaxFloat64, invalidIndices)
	assert.Nil(t, got)
	requireErrorChain(t, err, frequency.ErrInvalidParameter, nil)
	assert.Equal(t, originalInvalid, invalidIndices)

	got, err = frequency.Partials(math.MaxFloat64, []int{1, 2})
	assert.Nil(t, got)
	requireErrorChain(t, err, frequency.ErrOutOfRange, cents.ErrOutOfRange)
}

func TestPartials_InvalidFrequencyBeforeList(t *testing.T) {
	for _, invalid := range invalidFrequencyCases() {
		t.Run(invalid.name, func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Partials(invalid.f, nil)
			assert.Nil(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
	}
}
