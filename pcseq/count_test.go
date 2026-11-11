package pcseq_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/pcseq"
	"github.com/go-muse/musex/pcset"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRowCountMachineAndBig(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		modulus  int
		expected int
	}{
		{modulus: 1, expected: 1},
		{modulus: 2, expected: 2},
		{modulus: 12, expected: 479_001_600},
		{modulus: 20, expected: 2_432_902_008_176_640_000},
	} {
		t.Run(big.NewInt(int64(tt.modulus)).String(), func(t *testing.T) {
			t.Parallel()
			got, err := pcseq.RowCount(tt.modulus)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
			exact, err := pcseq.RowCountBig(tt.modulus)
			require.NoError(t, err)
			require.NotNil(t, exact)
			assert.Equal(t, big.NewInt(int64(tt.expected)), exact)
		})
	}

	got, err := pcseq.RowCount(21)
	assert.Zero(t, got)
	assert.ErrorIs(t, err, pcseq.ErrOverflow)
	exact, err := pcseq.RowCountBig(21)
	require.NoError(t, err)
	expected, ok := new(big.Int).SetString("51090942171709440000", 10)
	require.True(t, ok)
	assert.Equal(t, expected, exact)
	assert.PanicsWithValue(t, err, func() { pcseq.MustRowCount(21) })

	_, err = pcseq.RowCount(0)
	assert.ErrorIs(t, err, pcseq.ErrInvalidModulus)
	_, err = pcseq.RowCountBig(65_537)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
}

func TestRowClassCount(t *testing.T) {
	t.Parallel()
	expected := map[int]int{1: 1, 2: 1, 3: 1, 4: 3, 5: 8, 6: 38, 7: 192, 8: 1_320, 12: 9_985_920}
	for modulus, want := range expected {
		t.Run(big.NewInt(int64(modulus)).String(), func(t *testing.T) {
			t.Parallel()
			got, err := pcseq.RowClassCount(modulus, pcseq.TIR)
			require.NoError(t, err)
			assert.Equal(t, want, got)
			exact, err := pcseq.RowClassCountBig(modulus, pcseq.TIR)
			require.NoError(t, err)
			assert.Equal(t, big.NewInt(int64(want)), exact)
		})
	}
	for _, tt := range []struct {
		kind pcseq.RowClassKind
		want int
	}{
		{kind: pcseq.TIRRotations, want: 836_017},
		{kind: pcseq.TTOR, want: 4_995_360},
	} {
		got, err := pcseq.RowClassCount(12, tt.kind)
		require.NoError(t, err)
		assert.Equal(t, tt.want, got)
		exact, err := pcseq.RowClassCountBig(12, tt.kind)
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(int64(tt.want)), exact)
		_, err = pcseq.RowClassCount(11, tt.kind)
		assert.ErrorIs(t, err, pcseq.ErrNotModulus12)
		_, err = pcseq.RowClassCountBig(11, tt.kind)
		assert.ErrorIs(t, err, pcseq.ErrNotModulus12)
	}

	_, err := pcseq.RowClassCount(0, pcseq.TIR)
	assert.ErrorIs(t, err, pcseq.ErrInvalidModulus)
	_, err = pcseq.RowClassCount(1_025, pcseq.TIR)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
	_, err = pcseq.RowClassCount(12, pcseq.RowClassKind(0))
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.RowClassCount(23, pcseq.TIR)
	assert.ErrorIs(t, err, pcseq.ErrOverflow)
	exact, err := pcseq.RowClassCountBig(23, pcseq.TIR)
	require.NoError(t, err)
	assert.Positive(t, exact.Sign())
}

func TestCountRowsWithSegmentLengths(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		lengths []int
		want    int
	}{
		{name: "hexachords", lengths: []int{6, 6}, want: 518_400},
		{name: "tetrachords", lengths: []int{4, 4, 4}, want: 13_824},
		{name: "trichords", lengths: []int{3, 3, 3, 3}, want: 1_296},
		{name: "dyads", lengths: []int{2, 2, 2, 2, 2, 2}, want: 64},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pcseq.CountRowsWithSegmentLengths(12, tt.lengths...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			exact, err := pcseq.CountRowsWithSegmentLengthsBig(12, tt.lengths...)
			require.NoError(t, err)
			assert.Equal(t, big.NewInt(int64(tt.want)), exact)
		})
	}

	for _, tt := range []struct {
		name    string
		modulus int
		lengths []int
	}{
		{name: "no lengths", modulus: 12, lengths: nil},
		{name: "zero length", modulus: 12, lengths: []int{6, 0, 6}},
		{name: "negative length", modulus: 12, lengths: []int{6, -1, 7}},
		{name: "wrong sum", modulus: 12, lengths: []int{5, 5}},
		{name: "sum overflow", modulus: 12, lengths: []int{math.MaxInt, math.MaxInt}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pcseq.CountRowsWithSegmentLengths(tt.modulus, tt.lengths...)
			assert.Zero(t, got)
			assert.ErrorIs(t, err, pcseq.ErrInvalidSegmentContent)
			exact, err := pcseq.CountRowsWithSegmentLengthsBig(tt.modulus, tt.lengths...)
			assert.Nil(t, exact)
			assert.ErrorIs(t, err, pcseq.ErrInvalidSegmentContent)
		})
	}
	_, err := pcseq.CountRowsWithSegmentLengths(0, 1)
	assert.ErrorIs(t, err, pcseq.ErrInvalidModulus)
	_, err = pcseq.CountRowsWithSegmentLengths(65_537, 1)
	assert.ErrorIs(t, err, pcseq.ErrModulusTooLarge)
	_, err = pcseq.CountRowsWithSegmentLengths(21, 21)
	assert.ErrorIs(t, err, pcseq.ErrOverflow)
	exact, err := pcseq.CountRowsWithSegmentLengthsBig(21, 21)
	require.NoError(t, err)
	assert.NotNil(t, exact)
}

func partitionContents(t *testing.T, modulus int, lengths ...int) []pcset.Set {
	t.Helper()
	result := make([]pcset.Set, len(lengths))
	offset := 0
	for i, length := range lengths {
		result[i] = mustSet(t, modulus, identityNumbers(modulus)[offset:offset+length]...)
		offset += length
	}
	return result
}

func TestCountRowsWithSegmentContents(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		lengths []int
		want    int
	}{
		{name: "hexachords", lengths: []int{6, 6}, want: 518_400},
		{name: "tetrachords", lengths: []int{4, 4, 4}, want: 13_824},
		{name: "trichords", lengths: []int{3, 3, 3, 3}, want: 1_296},
		{name: "dyads", lengths: []int{2, 2, 2, 2, 2, 2}, want: 64},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			contents := partitionContents(t, 12, tt.lengths...)
			got, err := pcseq.CountRowsWithSegmentContents(12, contents...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			exact, err := pcseq.CountRowsWithSegmentContentsBig(12, contents...)
			require.NoError(t, err)
			assert.Equal(t, big.NewInt(int64(tt.want)), exact)
		})
	}

	for _, tt := range []struct {
		name     string
		contents []pcset.Set
	}{
		{name: "empty input", contents: nil},
		{name: "overlap", contents: []pcset.Set{
			mustSet(t, 12, 0, 1, 2, 3, 4, 5),
			mustSet(t, 12, 5, 6, 7, 8, 9, 10),
		}},
		{name: "sum and coverage", contents: []pcset.Set{
			mustSet(t, 12, 0, 1, 2, 3, 4),
			mustSet(t, 12, 5, 6, 7, 8, 9),
		}},
		{name: "incomplete coverage", contents: []pcset.Set{
			mustSet(t, 12, 0, 1, 2, 3, 4, 5),
			mustSet(t, 12, 6, 7, 8, 9, 10),
		}},
		{name: "empty segment", contents: []pcset.Set{
			mustSet(t, 12),
			mustSet(t, 12, identityNumbers(12)...),
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pcseq.CountRowsWithSegmentContents(12, tt.contents...)
			assert.Zero(t, got)
			assert.ErrorIs(t, err, pcseq.ErrInvalidSegmentContent)
			exact, err := pcseq.CountRowsWithSegmentContentsBig(12, tt.contents...)
			assert.Nil(t, exact)
			assert.ErrorIs(t, err, pcseq.ErrInvalidSegmentContent)
		})
	}

	_, err := pcseq.CountRowsWithSegmentContents(12, pcset.Set{})
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.CountRowsWithSegmentContents(12,
		mustSet(t, 12, 0, 1, 2, 3, 4, 5),
		mustSet(t, 13, 0, 1, 2, 3, 4, 5),
	)
	assert.ErrorIs(t, err, pcseq.ErrModulusMismatch)
	_, err = pcseq.CountRowsWithSegmentContents(0, mustSet(t, 12, 0))
	assert.ErrorIs(t, err, pcseq.ErrInvalidModulus)
}

func TestAllIntervalCount(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		kind pcseq.AllIntervalCountKind
		want int
	}{
		{kind: pcseq.AllIntervalTotal, want: 46_272},
		{kind: pcseq.AllIntervalFromZero, want: 3_856},
		{kind: pcseq.AllIntervalUpToTI, want: 1_928},
		{kind: pcseq.AllIntervalUpToTIR, want: 1_008},
		{kind: pcseq.AllIntervalRSymmetric, want: 88},
		{kind: pcseq.AllIntervalRISymmetric, want: 0},
	} {
		t.Run(big.NewInt(int64(tt.want)).String(), func(t *testing.T) {
			t.Parallel()
			assert.True(t, tt.kind.IsValid())
			got, err := pcseq.AllIntervalCount(12, tt.kind)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			exact, err := pcseq.AllIntervalCountBig(12, tt.kind)
			require.NoError(t, err)
			assert.Equal(t, big.NewInt(int64(tt.want)), exact)
		})
	}
	assert.False(t, pcseq.AllIntervalCountKind(0).IsValid())
	assert.False(t, pcseq.AllIntervalCountKind(99).IsValid())
	_, err := pcseq.AllIntervalCount(12, pcseq.AllIntervalCountKind(0))
	assert.ErrorIs(t, err, pcseq.ErrInvalidValue)
	_, err = pcseq.AllIntervalCount(11, pcseq.AllIntervalTotal)
	assert.ErrorIs(t, err, pcseq.ErrNotModulus12)
	_, err = pcseq.AllIntervalCountBig(11, pcseq.AllIntervalTotal)
	assert.ErrorIs(t, err, pcseq.ErrNotModulus12)
	_, err = pcseq.AllIntervalCount(0, pcseq.AllIntervalTotal)
	assert.ErrorIs(t, err, pcseq.ErrInvalidModulus)
}
