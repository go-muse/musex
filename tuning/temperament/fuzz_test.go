package temperament_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/tuning/temperament"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FuzzCumulativeOffsetAndIntervals verifies monotonicity, octave periodicity, DirectedInterval identities, and CyclicInterval complementarity across presets and arbitrary steps.
func FuzzCumulativeOffsetAndIntervals(f *testing.F) {
	f.Add(0, 0, 1, 1)
	f.Add(1, -12, 12, -2)
	f.Add(6, -25, 37, 3)
	f.Add(11, -100, 100, -5)
	f.Add(17, 0, 7, 10)
	f.Add(30, -53, 53, 2)

	// Built-in presets plus temperaments with N > 64 (uniform and non-uniform).
	presets := append(temperament.Presets(), largeTemperaments()...)

	f.Fuzz(func(t *testing.T, presetIdx, k1, k2, d int) {
		if len(presets) == 0 {
			t.Skip("no presets available")
		}
		idx := ((presetIdx % len(presets)) + len(presets)) % len(presets)
		tm := presets[idx]
		n := tm.StepsPerOctave()
		require.Positive(t, n)

		// Bound k1, k2, d to avoid int64 overflow while exercising multi-octave negative and positive ranges.
		k1 %= 10000
		k2 %= 10000
		d %= 1000

		c1, err := tm.CumulativeOffset(k1)
		require.NoError(t, err)
		c2, err := tm.CumulativeOffset(k2)
		require.NoError(t, err)

		// Strict monotonicity: k1 < k2 <=> c(k1) < c(k2).
		switch {
		case k1 < k2:
			assert.Negative(t, c1.Compare(c2))
		case k1 > k2:
			assert.Positive(t, c1.Compare(c2))
		default:
			assert.True(t, c1.Equal(c2))
		}

		// Octave periodicity (IV-13): c(k1 + d*N) == c(k1) + 1200*d.
		cShifted, err := tm.CumulativeOffset(k1 + d*n)
		require.NoError(t, err)
		octaveShift, err := cents.FromInt(d * 1200)
		require.NoError(t, err)
		expectedShifted, err := c1.Add(octaveShift)
		require.NoError(t, err)
		assert.True(t, cShifted.Equal(expectedShifted), "c(%d) = %v, want c(%d) + %d*1200 = %v", k1+d*n, cShifted, k1, d, expectedShifted)

		// SplitStep identity: c(k1) == octave*1200 + offset.
		octave, offset, err := tm.SplitStep(k1)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, offset.Compare(cents.Zero()), 0)
		assert.Less(t, offset.Compare(cents.Octave()), 0)
		octCents, err := cents.FromInt(octave * 1200)
		require.NoError(t, err)
		rebuilt, err := octCents.Add(offset)
		require.NoError(t, err)
		assert.True(t, c1.Equal(rebuilt))

		// DirectedInterval identities: I_dir(k1, k2) = c(k2) - c(k1), antisymmetry, and additivity.
		dir12, err := tm.DirectedInterval(k1, k2)
		require.NoError(t, err)
		dir21, err := tm.DirectedInterval(k2, k1)
		require.NoError(t, err)
		assert.True(t, dir12.Equal(dir21.Neg()))

		kMid := (k1 + k2) / 2
		dir1M, err := tm.DirectedInterval(k1, kMid)
		require.NoError(t, err)
		dirM2, err := tm.DirectedInterval(kMid, k2)
		require.NoError(t, err)
		sumDir, err := dir1M.Add(dirM2)
		require.NoError(t, err)
		assert.True(t, sumDir.Equal(dir12), "I(%d,%d) + I(%d,%d) = %v, want I(%d,%d) = %v", k1, kMid, kMid, k2, sumDir, k1, k2, dir12)
		assert.True(t, dir12.Equal(c2.MustSub(c1)), "I(%d,%d) must equal c(%d) - c(%d)", k1, k2, k2, k1)

		// CyclicInterval complementarity for modular steps m1, m2 in [0, N-1].
		m1 := ((k1 % n) + n) % n
		m2 := ((k2 % n) + n) % n
		cyc12, err := tm.CyclicInterval(m1, m2)
		require.NoError(t, err)
		cyc21, err := tm.CyclicInterval(m2, m1)
		require.NoError(t, err)

		if m1 == m2 {
			assert.True(t, cyc12.Equal(cents.Zero()))
			assert.True(t, cyc21.Equal(cents.Zero()))
		} else {
			sumCyc, err := cyc12.Add(cyc21)
			require.NoError(t, err)
			assert.True(t, sumCyc.Equal(cents.Octave()))
		}
	})
}

// FuzzRotateGroupLaws verifies group homomorphism, periodicity, and step-size sum preservation under Rotate.
func FuzzRotateGroupLaws(f *testing.F) {
	f.Add(0, 0, 0, 1)
	f.Add(5, 2, 7, -1)
	f.Add(12, -3, 15, 4)
	f.Add(17, -11, -5, -3)
	f.Add(25, 1, 11, 2)

	// Built-in presets plus temperaments with N > 64 (uniform and non-uniform).
	presets := append(temperament.Presets(), largeTemperaments()...)

	f.Fuzz(func(t *testing.T, presetIdx, a, b, d int) {
		if len(presets) == 0 {
			t.Skip("no presets available")
		}
		idx := ((presetIdx % len(presets)) + len(presets)) % len(presets)
		tm := presets[idx]
		n := tm.StepsPerOctave()
		require.Positive(t, n)

		a %= 10000
		b %= 10000
		d %= 1000

		rotA, err := tm.Rotate(a)
		require.NoError(t, err)
		rotAB, err := rotA.Rotate(b)
		require.NoError(t, err)
		rotSum, err := tm.Rotate(a + b)
		require.NoError(t, err)

		// Group law: Rotate(a).Rotate(b) == Rotate(a + b) exactly (P-12 item 3, K-13).
		assert.True(t, rotAB.Equal(rotSum), "Rotate(%d).Rotate(%d) != Rotate(%d)", a, b, a+b)

		// Periodicity: Rotate(a + d*N) == Rotate(a).
		rotPeriodic, err := tm.Rotate(a + d*n)
		require.NoError(t, err)
		assert.True(t, rotA.Equal(rotPeriodic))

		// Identity: Rotate(0) == tm (offsets match exactly).
		rotZero, err := tm.Rotate(0)
		require.NoError(t, err)
		assert.True(t, tm.Equal(rotZero))

		// StepSizes sum is preserved and equals 1200 cents.
		steps, err := rotA.StepSizes()
		require.NoError(t, err)
		require.Len(t, steps, n)
		sum := cents.Zero()
		for _, s := range steps {
			assert.Positive(t, s.Sign())
			sum, err = sum.Add(s)
			require.NoError(t, err)
		}
		assert.True(t, sum.Equal(cents.Octave()), "sum of step sizes = %v, want exactly 1200 cents", sum)
	})
}

// FuzzNearestStep verifies optimality and tie-breaking of NearestStep and NearestCumulativeStep across arbitrary cent inputs.
func FuzzNearestStep(f *testing.F) {
	f.Add(0, 0.0)
	f.Add(0, 50.0)
	f.Add(0, 1150.0)
	f.Add(6, 386.313714)
	f.Add(12, -1200.0)
	f.Add(17, 701.955001)
	f.Add(22, 2450.123456)

	// Built-in presets plus temperaments with N > 64 (uniform and non-uniform).
	presets := append(temperament.Presets(), largeTemperaments()...)

	f.Fuzz(func(t *testing.T, presetIdx int, centsVal float64) {
		if math.IsNaN(centsVal) || math.IsInf(centsVal, 0) || math.Abs(centsVal) > 1e6 {
			t.Skip("out-of-range float64 input")
		}
		if len(presets) == 0 {
			t.Skip("no presets available")
		}
		idx := ((presetIdx % len(presets)) + len(presets)) % len(presets)
		tm := presets[idx]
		n := tm.StepsPerOctave()
		require.Positive(t, n)

		x, err := cents.FromFloat(centsVal)
		require.NoError(t, err)

		mStar, delta, err := tm.NearestStep(x)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, mStar, 0)
		assert.Less(t, mStar, n)

		// Compare against all modular steps in [0, N-1] on the circle [0, 1200).
		_, xRed := x.SplitOctaves()
		bestDist := delta.Abs()
		offsets := tm.Offsets()
		for m, tmOffset := range offsets {
			diff, err := xRed.Sub(tmOffset)
			require.NoError(t, err)
			// Circular distance in [-600, +600].
			_, rawDiff := diff.SplitOctaves()
			dist := rawDiff
			negCandidate, err := rawDiff.Sub(cents.Octave())
			require.NoError(t, err)
			if negCandidate.Abs().Compare(dist.Abs()) < 0 {
				dist = negCandidate
			}
			assert.LessOrEqual(t, bestDist.Compare(dist.Abs()), 0, "step %d has smaller distance than chosen step %d", m, mStar)
		}

		// Verify consistency between NearestCumulativeStep and NearestStep.
		kStar, cumDelta, err := tm.NearestCumulativeStep(x)
		require.NoError(t, err)
		assert.True(t, delta.Equal(cumDelta))
		modStep := ((kStar % n) + n) % n
		assert.Equal(t, mStar, modStep)
		_, splitOffset, err := tm.SplitStep(kStar)
		require.NoError(t, err)
		assert.True(t, splitOffset.Equal(tm.MustOffset(mStar)))

		ckStar, err := tm.CumulativeOffset(kStar)
		require.NoError(t, err)
		reconstructed, err := ckStar.Add(cumDelta)
		require.NoError(t, err)
		assert.True(t, reconstructed.Equal(x), "c(%d) + delta = %v, want x = %v", kStar, reconstructed, x)
	})
}
