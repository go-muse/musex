package tuning_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzFrequencyOfStepProperties(f *testing.F) {
	for _, k := range []int{-600, -12, -1, 0, 1, 12, 600} {
		f.Add(k)
	}
	std := tuning.Standard()
	f.Fuzz(func(t *testing.T, k int) {
		if k < -600 || k > 600 {
			t.Skip("keep both frequencies normal")
		}
		base := std.MustFrequencyOfStep(system.Step(k))
		octave := std.MustFrequencyOfStep(system.Step(k + std.Classes()))
		assert.Equal(t, 2*base, octave, "f(k+N) must be exactly twice f(k)")
		shifted := std.MustShiftedFrequencyOfStep(system.Step(k), cents.Zero())
		assert.Equal(t, math.Float64bits(base), math.Float64bits(shifted))
	})
}

func FuzzCentsBetweenSteps(f *testing.F) {
	for _, pair := range [][2]int{{-1000, 1000}, {-2, 3}, {0, 0}, {1, -1}, {999, -999}} {
		f.Add(pair[0], pair[1])
	}
	std := tuning.Standard()
	f.Fuzz(func(t *testing.T, a, b int) {
		if a < -1000 || a > 1000 || b < -1000 || b > 1000 {
			t.Skip("keep both heights in the guaranteed cents range")
		}
		ha := std.MustCentsFromReferenceOfStep(system.Step(a))
		hb := std.MustCentsFromReferenceOfStep(system.Step(b))
		got := std.MustCentsBetweenSteps(system.Step(a), system.Step(b))
		want := hb.MustSub(ha)
		assert.True(t, want.Equal(got), "CentsBetweenSteps(%d, %d) = %v, want %v", a, b, got, want)
	})
}

func FuzzNearestIsAdjacentWithBoundedDeviation(f *testing.F) {
	for _, d := range []int{-1200, -150, -50, 0, 50, 150, 1200} {
		f.Add(d)
	}
	std := tuning.Standard()
	f.Fuzz(func(t *testing.T, input int) {
		if input < -2400 || input > 2400 {
			t.Skip("bounded seed domain")
		}
		freq := cents.MustFromInt(input).MustApplyTo(440)
		d := cents.MustBetween(440, freq)
		step, deviation, err := std.Nearest(freq)
		require.NoError(t, err)

		lower := 0
		found := false
		for k := -30; k < 30; k++ {
			h := cents.MustFromInt(100 * k)
			next := cents.MustFromInt(100 * (k + 1))
			if h.Compare(d) <= 0 && d.Compare(next) < 0 {
				lower = k
				found = true
				break
			}
		}
		require.True(t, found)
		assert.Contains(t, []system.Step{system.Step(lower), system.Step(lower + 1)}, step)
		wantDeviation := d.MustSub(cents.MustFromInt(100 * int(step)))
		assert.True(t, wantDeviation.Equal(deviation))
		assert.LessOrEqual(t, deviation.Abs().Compare(cents.MustFromInt(100)), 0)
	})
}

func FuzzSoundsSameSymmetry(f *testing.F) {
	for _, shift := range []int{-12000, -100, 0, 100, 12000} {
		f.Add(shift)
	}
	std := tuning.Standard()
	f.Fuzz(func(t *testing.T, shift int) {
		if shift < -12000 || shift > 12000 {
			t.Skip("bounded frequency domain")
		}
		other := std.MustWithReferenceFrequency(cents.MustFromInt(shift).MustApplyTo(440))
		if shift%2 != 0 {
			other = other.MustWithRoot(tuning.RootFromClass(pitchclass.MustNew(9, 12)))
		}
		forward, err := std.SoundsSame(other)
		require.NoError(t, err)
		reverse, err := other.SoundsSame(std)
		require.NoError(t, err)
		assert.Equal(t, forward, reverse)
	})
}

func FuzzPitchOfStepRoundTrip(f *testing.F) {
	for _, k := range []int{-100000, -12, -1, 0, 1, 12, 100000} {
		f.Add(k)
	}
	std := tuning.Standard()
	f.Fuzz(func(t *testing.T, k int) {
		if k < -100000 || k > 100000 {
			t.Skip("bounded octave domain")
		}
		note := std.MustPitchOfStep(system.Step(k), system.Sharps())
		assert.Equal(t, system.Step(k), std.MustStepOf(note))
	})
}
