package frequency_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubharmonic_ControlExamples(t *testing.T) {
	for _, n := range []int{1, 2, 3, 4} {
		t.Run("subharmonic", func(t *testing.T) {
			t.Parallel()
			want, ok := modelSubharmonic(440, n)
			require.True(t, ok)
			got, err := frequency.Subharmonic(440, n)
			require.NoError(t, err)
			assert.Equal(t, want, got)
			assert.Equal(t, want, frequency.MustSubharmonic(440, n))
		})
	}
	got, err := frequency.Subharmonics(440, []int{1, 2, 3, 4})
	require.NoError(t, err)
	require.Len(t, got, 4)
	for i, n := range []int{1, 2, 3, 4} {
		want, ok := modelSubharmonic(440, n)
		require.True(t, ok)
		assert.Equal(t, want, got[i])
	}
	assert.Equal(t, got, frequency.MustSubharmonics(440, []int{1, 2, 3, 4}))
}

func TestSubharmonic_ErrorsAndPrecedence(t *testing.T) {
	for _, invalid := range invalidFrequencyCases() {
		t.Run(invalid.name+" before index", func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Subharmonic(invalid.f, 0)
			assert.Zero(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
	}
	for _, n := range []int{0, -1} {
		got, err := frequency.Subharmonic(440, n)
		assert.Zero(t, got)
		requireErrorChain(t, err, frequency.ErrInvalidParameter, nil)
	}
	for _, tt := range []struct {
		name string
		f    float64
		n    int
	}{
		{"subnormal fundamental", math.SmallestNonzeroFloat64, 1},
		{"subnormal result", cents.MinNormalFrequency, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Subharmonic(tt.f, tt.n)
			assert.Zero(t, got)
			requireErrorChain(t, err, frequency.ErrOutOfRange, cents.ErrOutOfRange)
		})
	}
}

func TestSubharmonics_OrderOwnershipAndAtomicity(t *testing.T) {
	indices := []int{4, 1, 3, 2}
	original := append([]int(nil), indices...)
	got, err := frequency.Subharmonics(440, indices)
	require.NoError(t, err)
	require.Len(t, got, len(indices))
	assert.Equal(t, len(indices), cap(got))
	assert.Equal(t, original, indices)
	for i, n := range indices {
		want, ok := modelSubharmonic(440, n)
		require.True(t, ok)
		assert.Equal(t, want, got[i])
	}
	got[0] = 0
	fresh, err := frequency.Subharmonics(440, indices)
	require.NoError(t, err)
	assert.Equal(t, 110.0, fresh[0])
	assert.NotSame(t, &got[0], &fresh[0])

	for _, invalid := range [][]int{nil, {}, {0}, {1, -2}} {
		got, err := frequency.Subharmonics(440, invalid)
		assert.Nil(t, got)
		requireErrorChain(t, err, frequency.ErrInvalidParameter, nil)
	}

	got, err = frequency.Subharmonics(cents.MinNormalFrequency, []int{2, 0})
	assert.Nil(t, got)
	requireErrorChain(t, err, frequency.ErrInvalidParameter, nil)
	got, err = frequency.Subharmonics(cents.MinNormalFrequency, []int{1, 2})
	assert.Nil(t, got)
	requireErrorChain(t, err, frequency.ErrOutOfRange, cents.ErrOutOfRange)
}

func TestSubharmonics_InvalidFrequencyBeforeList(t *testing.T) {
	for _, invalid := range invalidFrequencyCases() {
		t.Run(invalid.name, func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Subharmonics(invalid.f, nil)
			assert.Nil(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
	}
}
