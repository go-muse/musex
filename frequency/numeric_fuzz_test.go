package frequency_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzRatio(f *testing.F) {
	for _, pair := range frequencyPairs() {
		f.Add(pair.f1, pair.f2)
	}
	f.Add(440.0, 660.0)
	f.Add(440.0, 442.0)
	f.Add(0.3, 0.1)
	f.Fuzz(func(t *testing.T, f1, f2 float64) {
		if !cents.IsValidFrequency(f1) || !cents.IsValidFrequency(f2) {
			t.Skip("outside the valid frequency domain")
		}
		got, err := frequency.Ratio(f1, f2)
		require.NoError(t, err)
		require.NotNil(t, got)
		want := new(big.Rat).Quo(decimalRat(f2), decimalRat(f1))
		assert.Zero(t, got.Cmp(want))
		product := new(big.Rat).Mul(got, decimalRat(f1))
		assert.Zero(t, product.Cmp(decimalRat(f2)))
	})
}

func FuzzBeatPeriod(f *testing.F) {
	for _, pair := range [][2]float64{
		{440, 442},
		{0.3, 0.1},
		{0.07, 0.07},
		{0x1.e4e2b8b522846p+1023, 0x1.e4e2b8b522846p+1023},
		{math.SmallestNonzeroFloat64, 2 * math.SmallestNonzeroFloat64},
		{math.MaxFloat64, math.SmallestNonzeroFloat64},
	} {
		f.Add(pair[0], pair[1])
	}
	f.Fuzz(func(t *testing.T, f1, f2 float64) {
		if !cents.IsValidFrequency(f1) || !cents.IsValidFrequency(f2) {
			t.Skip("outside the valid frequency domain")
		}
		beat, err := frequency.BeatFrequency(f1, f2)
		require.NoError(t, err)
		assert.Equal(t, modelBeatFrequency(f1, f2), beat)
		assert.GreaterOrEqual(t, beat, 0.0)
		reverse, err := frequency.BeatFrequency(f2, f1)
		require.NoError(t, err)
		assert.Equal(t, beat, reverse)

		for _, input := range []float64{f1, f2} {
			period, err := frequency.Period(input)
			want, ok := modelPeriod(input)
			if !ok {
				assert.Zero(t, period)
				requireErrorChain(t, err, frequency.ErrOutOfRange, nil)
				continue
			}
			require.NoError(t, err)
			assert.Equal(t, want, period)
			assert.Positive(t, period)
		}
	})
}

func fuzzIndices(data []byte) ([]int, bool) {
	indices := make([]int, len(data)%8)
	for i := range indices {
		indices[i] = int(data[i]%4 + 1)
	}
	invalid := false
	if len(indices) > 0 && len(data) > 0 && data[0]%17 == 0 {
		indices[len(indices)-1] = 0
		invalid = true
	}
	return indices, invalid
}

func FuzzPartials(f *testing.F) {
	f.Add(110.0, []byte{2, 3, 4})
	f.Add(math.MaxFloat64, []byte{1, 2})
	f.Fuzz(func(t *testing.T, input float64, data []byte) {
		if !cents.IsValidFrequency(input) {
			t.Skip("outside the valid frequency domain")
		}
		indices, invalid := fuzzIndices(data)
		got, err := frequency.Partials(input, indices)
		if len(indices) == 0 || invalid {
			assert.Nil(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidParameter, nil)
			return
		}
		if err != nil {
			assert.Nil(t, got)
			requireErrorChain(t, err, frequency.ErrOutOfRange, cents.ErrOutOfRange)
			return
		}
		require.Len(t, got, len(indices))
		assert.Equal(t, len(indices), cap(got))
		for i, n := range indices {
			want, ok := modelPartial(input, n)
			require.True(t, ok)
			assert.Equal(t, want, got[i])
		}
	})
}

func FuzzSubharmonics(f *testing.F) {
	f.Add(440.0, []byte{1, 2, 3, 4})
	f.Add(cents.MinNormalFrequency, []byte{1, 2})
	f.Fuzz(func(t *testing.T, input float64, data []byte) {
		if !cents.IsValidFrequency(input) {
			t.Skip("outside the valid frequency domain")
		}
		indices, invalid := fuzzIndices(data)
		got, err := frequency.Subharmonics(input, indices)
		if len(indices) == 0 || invalid {
			assert.Nil(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidParameter, nil)
			return
		}
		if err != nil {
			assert.Nil(t, got)
			requireErrorChain(t, err, frequency.ErrOutOfRange, cents.ErrOutOfRange)
			return
		}
		require.Len(t, got, len(indices))
		assert.Equal(t, len(indices), cap(got))
		for i, n := range indices {
			want, ok := modelSubharmonic(input, n)
			require.True(t, ok)
			assert.Equal(t, want, got[i])
		}
	})
}

func FuzzNormalizeOctave(f *testing.F) {
	for _, pair := range [][2]float64{
		{880, 110},
		{440, 110},
		{220, 110},
		{math.Nextafter(220, 110), 110},
		{math.MaxFloat64, math.MaxFloat64},
		{math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64},
	} {
		f.Add(pair[0], pair[1])
	}
	f.Fuzz(func(t *testing.T, input, base float64) {
		if !cents.IsValidFrequency(input) || !cents.IsValidFrequency(base) {
			t.Skip("outside the valid frequency domain")
		}
		want, wantK, ok, candidate := modelNormalizeOctave(input, base)
		got, k, err := frequency.NormalizeOctave(input, base)
		if !ok {
			assert.Zero(t, got)
			assert.Zero(t, k)
			requireErrorChain(t, err, frequency.ErrOutOfRange, cents.ErrOutOfRange)
			return
		}
		require.NoError(t, err)
		assert.Equal(t, want, got)
		assert.Equal(t, wantK, k)
		lower := rational(base)
		upper := new(big.Rat).Mul(lower, big.NewRat(2, 1))
		assert.GreaterOrEqual(t, candidate.Cmp(lower), 0)
		assert.Less(t, candidate.Cmp(upper), 0)
	})
}
