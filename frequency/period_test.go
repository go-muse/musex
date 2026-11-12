package frequency_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeriod_ControlExamples(t *testing.T) {
	for _, tt := range []struct {
		name string
		f    float64
	}{
		{"K-50 concert pitch", 440},
		{"K-50 subnormal period", math.MaxFloat64},
		{"K-53 binary reciprocal", 0.07},
		{"K-54 direct rounding", 0x1.e4e2b8b522846p+1023},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			want, ok := modelPeriod(tt.f)
			require.True(t, ok)
			got, err := frequency.Period(tt.f)
			require.NoError(t, err)
			assert.Equal(t, want, got)
			assert.Equal(t, want, frequency.MustPeriod(tt.f))
			assert.Positive(t, got)
		})
	}
}

func TestPeriod_DirectRoundingRegression(t *testing.T) {
	const input = 0x1.e4e2b8b522846p+1023
	const want = 0x0.43942db6a69f1p-1022
	const wrongWithIntermediateRound53 = 0x0.43942db6a69f0p-1022
	got, err := frequency.Period(input)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.NotEqual(t, wrongWithIntermediateRound53, got)
	model, ok := modelPeriod(input)
	require.True(t, ok)
	assert.Equal(t, want, model)
}

func TestPeriod_OverflowAndInvalidFrequency(t *testing.T) {
	got, err := frequency.Period(math.SmallestNonzeroFloat64)
	assert.Zero(t, got)
	requireErrorChain(t, err, frequency.ErrOutOfRange, nil)

	for _, invalid := range invalidFrequencyCases() {
		t.Run(invalid.name, func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Period(invalid.f)
			assert.Zero(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
	}
}
