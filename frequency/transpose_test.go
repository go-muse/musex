package frequency_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranspose_MatchesCentsComposition(t *testing.T) {
	for _, tt := range []struct {
		name string
		f    float64
		k, n int
	}{
		{"K-37 seven semitones", 440, 7, 12},
		{"down seven semitones", 440, -7, 12},
		{"zero steps", 440, 0, 12},
		{"negative direction", 660, -1, 12},
		{"large normal division", 440, 3, 19},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			step, err := cents.Step(tt.k, tt.n)
			require.NoError(t, err)
			want, err := step.ApplyTo(tt.f)
			require.NoError(t, err)
			got, err := frequency.Transpose(tt.f, tt.k, tt.n)
			require.NoError(t, err)
			assert.Equal(t, want, got)
			assert.Equal(t, want, frequency.MustTranspose(tt.f, tt.k, tt.n))
		})
	}
}

func TestTranspose_ErrorsAndPrecedence(t *testing.T) {
	for _, invalid := range invalidFrequencyCases() {
		t.Run(invalid.name+" before parameter", func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Transpose(invalid.f, 1, 0)
			assert.Zero(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
	}
	for _, n := range []int{0, -1} {
		t.Run("invalid division", func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Transpose(440, 1, n)
			assert.Zero(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidParameter, cents.ErrInvalidArgument)
		})
	}
	for _, tt := range []struct {
		name string
		f    float64
		k, n int
	}{
		{"subnormal result", math.SmallestNonzeroFloat64, 0, 12},
		{"overflow", math.MaxFloat64, 1, 12},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Transpose(tt.f, tt.k, tt.n)
			assert.Zero(t, got)
			requireErrorChain(t, err, frequency.ErrOutOfRange, cents.ErrOutOfRange)
		})
	}
}

func TestTranspose_BoundaryBeforeFloat64(t *testing.T) {
	const n = 12_000_000_000_000_000
	step, err := cents.Step(-1, n)
	require.NoError(t, err)
	_, err = step.ApplyTo(cents.MinNormalFrequency)
	require.ErrorIs(t, err, cents.ErrOutOfRange)

	got, err := frequency.Transpose(cents.MinNormalFrequency, -1, n)
	assert.Zero(t, got)
	requireErrorChain(t, err, frequency.ErrOutOfRange, cents.ErrOutOfRange)
}

func TestCentsBetween_ControlExamples(t *testing.T) {
	for _, tt := range []struct {
		name   string
		f1, f2 float64
		want   cents.Cents
	}{
		{"K-39 upward", 440, 442, cents.MustParse("7.851415")},
		{"K-39 downward", 442, 440, cents.MustParse("-7.851415")},
		{"octave", 440, 880, cents.Octave()},
		{"unison", 440, 440, cents.Zero()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := frequency.CentsBetween(tt.f1, tt.f2)
			require.NoError(t, err)
			assert.True(t, got.Equal(tt.want))
			assert.True(t, frequency.MustCentsBetween(tt.f1, tt.f2).Equal(tt.want))
			direct, err := cents.Between(tt.f1, tt.f2)
			require.NoError(t, err)
			assert.True(t, got.Equal(direct))
		})
	}
}

func TestCentsBetween_InvalidFrequency(t *testing.T) {
	for _, invalid := range invalidFrequencyCases() {
		t.Run(invalid.name+"/first", func(t *testing.T) {
			t.Parallel()
			got, err := frequency.CentsBetween(invalid.f, 440)
			assert.True(t, got.Equal(cents.Cents{}))
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
		t.Run(invalid.name+"/second", func(t *testing.T) {
			t.Parallel()
			got, err := frequency.CentsBetween(440, invalid.f)
			assert.True(t, got.Equal(cents.Cents{}))
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
	}
}
