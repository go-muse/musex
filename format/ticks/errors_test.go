package ticks_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/format/ticks"

	"github.com/stretchr/testify/assert"
)

// TestSentinelsArePairwiseDistinct checks 2.5 of the business specification
// and 4.1 of the architecture: the ten kinds of failure are distinguishable
// through errors.Is, and no sentinel wraps another.
func TestSentinelsArePairwiseDistinct(t *testing.T) {
	t.Parallel()
	for i, a := range ticksSentinels {
		assert.NotNil(t, a)
		for j, b := range ticksSentinels {
			if i == j {
				assert.ErrorIs(t, a, b)
				continue
			}
			assert.False(t, errors.Is(a, b), "%v wraps %v", a, b)
			assert.NotEqual(t, a.Error(), b.Error())
		}
	}
}

// TestConstants checks the format limits and defaults of 4.1.
func TestConstants(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		got  int
		want int
	}{
		{name: "MaxSMFResolution", got: ticks.MaxSMFResolution, want: 32767},
		{name: "MaxMusicXMLResolution", got: ticks.MaxMusicXMLResolution, want: 16383},
		{name: "MaxDelta", got: ticks.MaxDelta, want: 268435455},
		{name: "MaxMicrosecondsPerQuarter", got: ticks.MaxMicrosecondsPerQuarter, want: 16777215},
		{name: "DefaultMicroseconds", got: ticks.DefaultMicroseconds, want: 500000},
		{name: "MaxByteField", got: ticks.MaxByteField, want: 255},
		{name: "MaxDenominatorExponent", got: ticks.MaxDenominatorExponent, want: 254},
		{name: "ClocksPerQuarter", got: ticks.ClocksPerQuarter, want: 24},
		{name: "ClocksPerBeat", got: ticks.ClocksPerBeat, want: 6},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}
