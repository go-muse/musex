package frequency_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBeatFrequency_ControlExamples(t *testing.T) {
	adjacent := math.Nextafter(440, math.Inf(1))
	for _, tt := range []struct {
		name   string
		f1, f2 float64
		want   float64
	}{
		{"K-49 upward", 440, 442, 2},
		{"K-49 reverse", 442, 440, 2},
		{"K-49 unison", 440, 440, 0},
		{"K-49 adjacent", 440, adjacent, modelBeatFrequency(440, adjacent)},
		{"subnormal difference", math.SmallestNonzeroFloat64, 2 * math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64},
		{"K-52 exact binary difference", 0.3, 0.1, 0.19999999999999998},
		{"extreme difference", math.SmallestNonzeroFloat64, math.MaxFloat64, modelBeatFrequency(math.SmallestNonzeroFloat64, math.MaxFloat64)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := frequency.BeatFrequency(tt.f1, tt.f2)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, modelBeatFrequency(tt.f1, tt.f2), got)
			assert.Equal(t, tt.want, frequency.MustBeatFrequency(tt.f1, tt.f2))
			assert.GreaterOrEqual(t, got, 0.0)
			reverse, err := frequency.BeatFrequency(tt.f2, tt.f1)
			require.NoError(t, err)
			assert.Equal(t, got, reverse)
		})
	}
}

func TestBeatFrequency_InvalidFrequency(t *testing.T) {
	for _, invalid := range invalidFrequencyCases() {
		t.Run(invalid.name+"/first", func(t *testing.T) {
			t.Parallel()
			got, err := frequency.BeatFrequency(invalid.f, 440)
			assert.Zero(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
		t.Run(invalid.name+"/second", func(t *testing.T) {
			t.Parallel()
			got, err := frequency.BeatFrequency(440, invalid.f)
			assert.Zero(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
	}
}
