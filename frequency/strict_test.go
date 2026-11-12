package frequency_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrictFunctions_MatchCriterion(t *testing.T) {
	for _, pair := range frequencyPairs() {
		t.Run(pair.name, func(t *testing.T) {
			t.Parallel()
			wantNear := pair.f1 == pair.f2
			got, err := frequency.Equal(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.Equal(t, wantNear, got)
			assert.Equal(t, wantNear, frequency.MustEqual(pair.f1, pair.f2))
			fromCriterion, err := frequency.Strict().Near(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.Equal(t, got, fromCriterion)

			wantComparison := 0
			if pair.f1 < pair.f2 {
				wantComparison = -1
			} else if pair.f1 > pair.f2 {
				wantComparison = 1
			}
			comparison, err := frequency.Compare(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.Equal(t, wantComparison, comparison)
			assert.Equal(t, wantComparison, frequency.MustCompare(pair.f1, pair.f2))
			fromCriterionComparison, err := frequency.Strict().Compare(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.Equal(t, comparison, fromCriterionComparison)
		})
	}
}

func TestStrictFunctions_InvalidFrequency(t *testing.T) {
	invalid := []struct {
		name string
		f    float64
	}{
		{"zero", 0},
		{"negative_zero", math.Copysign(0, -1)},
		{"negative", -1},
		{"nan", math.NaN()},
		{"positive_infinity", math.Inf(1)},
		{"negative_infinity", math.Inf(-1)},
	}
	for _, tt := range invalid {
		t.Run(tt.name+"/first", func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Equal(tt.f, 440)
			assert.False(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
			comparison, err := frequency.Compare(tt.f, 440)
			assert.Zero(t, comparison)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
		t.Run(tt.name+"/second", func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Equal(440, tt.f)
			assert.False(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
			comparison, err := frequency.Compare(440, tt.f)
			assert.Zero(t, comparison)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
	}
}
