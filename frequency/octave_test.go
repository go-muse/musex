package frequency_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShiftOctaves_ControlExamples(t *testing.T) {
	for _, tt := range []struct {
		name string
		f    float64
		k    int
		want float64
	}{
		{"K-41 up", 440, 1, 880},
		{"K-41 down", 440, -1, 220},
		{"zero", 440, 0, 440},
		{"two down", 880, -2, 220},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := frequency.ShiftOctaves(tt.f, tt.k)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, frequency.MustShiftOctaves(tt.f, tt.k))
			octaves, err := cents.Octave().Mul(tt.k)
			require.NoError(t, err)
			want, err := octaves.ApplyTo(tt.f)
			require.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}

func TestShiftOctaves_Errors(t *testing.T) {
	for _, invalid := range invalidFrequencyCases() {
		t.Run(invalid.name, func(t *testing.T) {
			t.Parallel()
			got, err := frequency.ShiftOctaves(invalid.f, 1)
			assert.Zero(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
	}
	for _, tt := range []struct {
		name string
		f    float64
		k    int
	}{
		{"K-44 overflow", math.MaxFloat64, 1},
		{"subnormal result", math.SmallestNonzeroFloat64, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := frequency.ShiftOctaves(tt.f, tt.k)
			assert.Zero(t, got)
			requireErrorChain(t, err, frequency.ErrOutOfRange, cents.ErrOutOfRange)
		})
	}
}
