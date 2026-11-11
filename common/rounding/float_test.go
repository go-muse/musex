package rounding_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModeFloatControlExamples(t *testing.T) {
	t.Parallel()
	inputs := []float64{0.5, -0.5, 1.5, -1.5, 2.5, -2.5}
	want := [][]float64{
		{0, -1, 1, -2, 2, -3},
		{1, 0, 2, -1, 3, -2},
		{0, 0, 1, -1, 2, -2},
		{1, -1, 2, -2, 3, -3},
		{1, 0, 2, -1, 3, -2},
		{0, -1, 1, -2, 2, -3},
		{1, -1, 2, -2, 3, -3},
		{0, 0, 1, -1, 2, -2},
		{0, 0, 2, -2, 2, -2},
	}
	for i, mode := range testModes {
		for j, x := range inputs {
			t.Run(fmt.Sprintf("%s/%g", testModeNames[i], x), func(t *testing.T) {
				t.Parallel()
				got, err := mode.Float(x)
				require.NoError(t, err)
				assert.Equal(t, want[i][j], got)
				gotInt, err := mode.FloatInt(x)
				require.NoError(t, err)
				assert.Equal(t, int(want[i][j]), gotInt)
			})
		}
	}
}

func TestModeFloatNegativeFalseTies(t *testing.T) {
	t.Parallel()
	for _, x := range []float64{math.Nextafter(-0.5, 0), -1.4999999999999998, -2.5000000000000004} {
		for i, mode := range testModes {
			t.Run(fmt.Sprintf("%g/%s", x, testModeNames[i]), func(t *testing.T) {
				t.Parallel()
				got, err := mode.Float(x)
				require.NoError(t, err)
				assert.Equal(t, expectedFloat(mode, x), got)
				gotInt, err := mode.FloatInt(x)
				require.NoError(t, err)
				assert.Equal(t, int(expectedFloatInt(mode, x).Int64()), gotInt)
			})
		}
	}
}

func TestModeFloatNodesAndReadingBoundary(t *testing.T) {
	t.Parallel()
	values := []float64{
		0.49999999999999994,
		4503599627370497,
		4503599627370495.5,
		9007199254740991,
		9007199254740992,
		-9007199254740992,
		1152921504606846976,
		9223372036854774784,
	}
	for _, x := range values {
		for i, mode := range testModes {
			t.Run(fmt.Sprintf("%g/%s", x, testModeNames[i]), func(t *testing.T) {
				t.Parallel()
				got, err := mode.Float(x)
				require.NoError(t, err)
				assert.Equal(t, expectedFloat(mode, x), got)
				gotInt, err := mode.FloatInt(x)
				require.NoError(t, err)
				assert.Equal(t, int(expectedFloatInt(mode, x).Int64()), gotInt)
			})
		}
	}
}

func TestModeFloatIntRange(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		mode rounding.Mode
		x    float64
		want int
		err  error
	}{
		{"half_even_negative", rounding.HalfEven, -2.5, -2, nil},
		{"min", rounding.Up, float64(math.MinInt), math.MinInt, nil},
		{"large_down", rounding.Down, 9.3e18, 0, rounding.ErrOutOfRange},
		{"two_to_63", rounding.Down, 9223372036854775807.0, 0, rounding.ErrOutOfRange},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.mode.FloatInt(tt.x)
			if tt.err != nil {
				checkError(t, err, tt.err)
				assert.Zero(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
	for i, mode := range testModes {
		t.Run(fmt.Sprintf("two_to_60/%s", testModeNames[i]), func(t *testing.T) {
			t.Parallel()
			got, err := mode.FloatInt(float64(uint64(1) << 60))
			require.NoError(t, err)
			assert.Equal(t, int(1152921504606846976), got)
			steps, err := mode.FloatSteps(float64(uint64(1)<<60), 1)
			require.NoError(t, err)
			assert.Equal(t, got, steps)
		})
	}
}

func TestModeFloatMultipleAndStepsControlExamples(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		mode     rounding.Mode
		x, step  float64
		want     float64
		wantStep int
	}{
		{rounding.HalfUp, 0.15, 0.1, 0.2, 2},
		{rounding.HalfDown, 0.15, 0.1, 0.1, 1},
		{rounding.HalfEven, 0.15, 0.1, 0.2, 2},
		{rounding.Down, 0.15, 0.1, 0.1, 1},
		{rounding.Up, 0.15, 0.1, 0.2, 2},
		{rounding.HalfUp, -0.15, 0.1, -0.1, -1},
		{rounding.HalfAwayFromZero, -0.15, 0.1, -0.2, -2},
		{rounding.HalfDown, 0.5, 0.3333333333333333, 0.6666666666666666, 2},
	} {
		t.Run(fmt.Sprintf("%s/%g_step_%g", tt.mode, tt.x, tt.step), func(t *testing.T) {
			t.Parallel()
			got, err := tt.mode.FloatMultiple(tt.x, tt.step)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			steps, err := tt.mode.FloatSteps(tt.x, tt.step)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStep, steps)
		})
	}

	t.Run("value_fits_but_step_index_does_not", func(t *testing.T) {
		t.Parallel()
		got, err := rounding.Down.FloatMultiple(1e308, 1e-308)
		require.NoError(t, err)
		assert.Equal(t, 1e308, got)
		steps, err := rounding.Down.FloatSteps(1e308, 1e-308)
		checkError(t, err, rounding.ErrOutOfRange)
		assert.Zero(t, steps)
	})

	t.Run("idempotence_is_not_promised", func(t *testing.T) {
		t.Parallel()
		first, err := rounding.Up.FloatMultiple(1000000000000001.5, 0.3)
		require.NoError(t, err)
		assert.Equal(t, 1000000000000001.75, first)
		second, err := rounding.Up.FloatMultiple(first, 0.3)
		require.NoError(t, err)
		assert.Equal(t, 1000000000000002.0, second)
	})
}

func TestModeFloatMultipleOverflowThreshold(t *testing.T) {
	t.Parallel()
	t.Run("below_threshold_returns_max_float", func(t *testing.T) {
		t.Parallel()
		step := math.Ldexp(5, 968)
		got, err := rounding.Up.FloatMultiple(math.MaxFloat64, step)
		require.NoError(t, err)
		assert.Equal(t, math.MaxFloat64, got)
	})
	t.Run("at_threshold_is_out_of_range", func(t *testing.T) {
		t.Parallel()
		step := math.Ldexp(3, 969)
		got, err := rounding.Up.FloatMultiple(math.MaxFloat64, step)
		checkError(t, err, rounding.ErrOutOfRange)
		assert.Zero(t, got)
	})
}

func TestModeFloatSubnormalInputsAndSteps(t *testing.T) {
	t.Parallel()
	for i, mode := range testModes {
		t.Run(testModeNames[i], func(t *testing.T) {
			t.Parallel()
			got, err := mode.FloatMultiple(5e-324, 5e-324)
			require.NoError(t, err)
			assert.Equal(t, 5e-324, got)
			steps, err := mode.FloatSteps(1e-323, 5e-324)
			require.NoError(t, err)
			assert.Equal(t, 2, steps)
		})
	}
}

func TestModeFloatFractionAndNumerator(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		mode rounding.Mode
		want float64
		num  int
	}{{rounding.HalfDown, 0.3333333333333333, 1}, {rounding.HalfUp, 0.6666666666666666, 2}, {rounding.HalfEven, 0.6666666666666666, 2}} {
		t.Run(tt.mode.String(), func(t *testing.T) {
			t.Parallel()
			got, err := tt.mode.FloatFraction(0.5, 3)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			num, err := tt.mode.FloatNumerator(0.5, 3)
			require.NoError(t, err)
			assert.Equal(t, tt.num, num)
		})
	}
	t.Run("max_float_does_not_overflow", func(t *testing.T) {
		t.Parallel()
		got, err := rounding.Up.FloatFraction(math.MaxFloat64, 7)
		require.NoError(t, err)
		assert.Equal(t, math.MaxFloat64, got)
	})
	t.Run("numerator_out_of_range", func(t *testing.T) {
		t.Parallel()
		got, err := rounding.Down.FloatNumerator(1e300, 1)
		checkError(t, err, rounding.ErrOutOfRange)
		assert.Zero(t, got)
	})
	for _, den := range []int{0, -3} {
		t.Run(fmt.Sprintf("invalid_den_%d", den), func(t *testing.T) {
			t.Parallel()
			got, err := rounding.HalfEven.FloatFraction(0.5, den)
			checkError(t, err, rounding.ErrInvalidArgument)
			assert.Zero(t, got)
			num, err := rounding.HalfEven.FloatNumerator(0.5, den)
			checkError(t, err, rounding.ErrInvalidArgument)
			assert.Zero(t, num)
		})
	}
}

func TestModeFloatPlaces(t *testing.T) {
	t.Parallel()
	got, err := rounding.HalfUp.FloatPlaces(0.145, 2)
	require.NoError(t, err)
	assert.Equal(t, 0.15, got)
	got, err = rounding.Up.FloatPlaces(math.MaxFloat64, -308)
	checkError(t, err, rounding.ErrOutOfRange)
	assert.Zero(t, got)
}

func TestModeFloatPositiveZero(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		call func() (float64, error)
	}{
		{"up_negative_fraction", func() (float64, error) { return rounding.Up.Float(-0.3) }},
		{"toward_zero_negative_fraction", func() (float64, error) { return rounding.TowardZero.Float(-0.3) }},
		{"half_toward_zero_tie", func() (float64, error) { return rounding.HalfTowardZero.Float(-0.5) }},
		{"places_negative_small", func() (float64, error) { return rounding.TowardZero.FloatPlaces(-0.001, 2) }},
		{"places_negative_small_half", func() (float64, error) { return rounding.HalfAwayFromZero.FloatPlaces(-0.001, 2) }},
		{"fraction_negative_small", func() (float64, error) { return rounding.TowardZero.FloatFraction(-0.1, 3) }},
		{"fraction_negative_small_half", func() (float64, error) { return rounding.HalfAwayFromZero.FloatFraction(-0.1, 3) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.call()
			require.NoError(t, err)
			assert.Zero(t, got)
			assert.False(t, math.Signbit(got))
		})
	}
	for i, mode := range testModes {
		t.Run(testModeNames[i], func(t *testing.T) {
			t.Parallel()
			for _, call := range []func() (float64, error){
				func() (float64, error) { return mode.Float(math.Copysign(0, -1)) },
				func() (float64, error) { return mode.FloatMultiple(math.Copysign(0, -1), 0.1) },
				func() (float64, error) { return mode.FloatPlaces(math.Copysign(0, -1), 2) },
				func() (float64, error) { return mode.FloatFraction(math.Copysign(0, -1), 3) },
			} {
				got, err := call()
				require.NoError(t, err)
				assert.Zero(t, got)
				assert.False(t, math.Signbit(got))
			}
		})
	}
}

func TestModeFloatOperationsAgree(t *testing.T) {
	t.Parallel()
	values := []float64{-2.5, -0.15, 0, 0.145, 0.5, 2.675, 1234.5}
	for _, x := range values {
		for i, mode := range testModes {
			t.Run(fmt.Sprintf("%g/%s", x, testModeNames[i]), func(t *testing.T) {
				t.Parallel()
				steps, err := mode.FloatSteps(x, 1)
				require.NoError(t, err)
				integer, err := mode.FloatInt(x)
				require.NoError(t, err)
				assert.Equal(t, integer, steps)
				got, err := mode.Float(x)
				require.NoError(t, err)
				assert.Equal(t, float64(integer), got)
				for _, places := range []int{-2, 0, 2, 400} {
					decimal, err := mode.FloatDecimal(x, places)
					require.NoError(t, err)
					ratDecimal, err := mode.RatDecimal(floatValue(x), places)
					require.NoError(t, err)
					assert.Equal(t, ratDecimal, decimal)
				}
			})
		}
	}
}
