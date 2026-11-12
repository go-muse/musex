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

func TestNormalizeOctave_ControlExamples(t *testing.T) {
	for _, tt := range []struct {
		name string
		f    float64
		base float64
		k    int
	}{
		{"K-42 880", 880, 110, -3},
		{"K-42 440", 440, 110, -2},
		{"K-42 220", 220, 110, -1},
		{"additional 220", 220, 110, -1},
		{"K-46 just below upper boundary", math.Nextafter(220, 110), 110, 0},
		{"K-46 maximum", math.MaxFloat64, math.MaxFloat64, 0},
		{"lower boundary", 110, 110, 0},
		{"below lower boundary", math.Nextafter(110, 0), 110, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			want, wantK, ok, candidate := modelNormalizeOctave(tt.f, tt.base)
			require.True(t, ok)
			assert.Equal(t, tt.k, wantK)
			got, k, err := frequency.NormalizeOctave(tt.f, tt.base)
			require.NoError(t, err)
			assert.Equal(t, want, got)
			assert.Equal(t, wantK, k)
			assert.True(t, cents.IsNormalFrequency(got))

			lower := rational(tt.base)
			upper := new(big.Rat).Mul(lower, big.NewRat(2, 1))
			assert.GreaterOrEqual(t, candidate.Cmp(lower), 0)
			assert.Less(t, candidate.Cmp(upper), 0)
			for otherK := -4; otherK <= 4; otherK++ {
				other := new(big.Rat).Mul(rational(tt.f), powerOfTwo(otherK))
				if other.Cmp(lower) >= 0 && other.Cmp(upper) < 0 {
					assert.Equal(t, k, otherK, "octave representative must be unique")
				}
			}

			mustGot, mustK := frequency.MustNormalizeOctave(tt.f, tt.base)
			assert.Equal(t, want, mustGot)
			assert.Equal(t, wantK, mustK)
		})
	}
}

func TestNormalizeOctave_NoNormalRepresentative(t *testing.T) {
	got, k, err := frequency.NormalizeOctave(math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64)
	assert.Zero(t, got)
	assert.Zero(t, k)
	requireErrorChain(t, err, frequency.ErrOutOfRange, cents.ErrOutOfRange)
}

func TestNormalizeOctave_InvalidFrequencyOrder(t *testing.T) {
	for _, invalid := range invalidFrequencyCases() {
		t.Run(invalid.name+"/frequency", func(t *testing.T) {
			t.Parallel()
			got, k, err := frequency.NormalizeOctave(invalid.f, math.NaN())
			assert.Zero(t, got)
			assert.Zero(t, k)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
		t.Run(invalid.name+"/base", func(t *testing.T) {
			t.Parallel()
			got, k, err := frequency.NormalizeOctave(440, invalid.f)
			assert.Zero(t, got)
			assert.Zero(t, k)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
	}
}
