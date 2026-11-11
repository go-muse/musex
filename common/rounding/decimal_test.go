package rounding_test

import (
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModeRatDecimalControlExamples(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		x      *big.Rat
		places int
		mode   rounding.Mode
		want   string
	}{
		{"eighth_half_even", bigRat(1, 8), 2, rounding.HalfEven, "0.12"},
		{"eighth_half_away", bigRat(1, 8), 2, rounding.HalfAwayFromZero, "0.13"},
		{"negative_eighth", bigRat(-1, 8), 2, rounding.HalfAwayFromZero, "-0.13"},
		{"two_thirds_half_away", bigRat(2, 3), 4, rounding.HalfAwayFromZero, "0.6667"},
		{"two_thirds_toward_zero", bigRat(2, 3), 4, rounding.TowardZero, "0.6666"},
		{"negative_zero", bigRat(-1, 10000), 2, rounding.HalfAwayFromZero, "0.00"},
		{"hundreds_half_even", bigRat(1250, 1), -2, rounding.HalfEven, "1200"},
		{"hundreds_half_up", bigRat(1250, 1), -2, rounding.HalfUp, "1300"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.mode.RatDecimal(tt.x, tt.places)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestModeFloatDecimalControlExamples(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		x      float64
		places int
		mode   rounding.Mode
		want   string
	}{
		{"decimal_tie_half_up", 0.145, 2, rounding.HalfUp, "0.15"},
		{"decimal_tie_half_even", 0.145, 2, rounding.HalfEven, "0.14"},
		{"decimal_tie_half_down", 0.145, 2, rounding.HalfDown, "0.14"},
		{"decimal_tie_up", 0.145, 2, rounding.Up, "0.15"},
		{"decimal_tie_down", 0.145, 2, rounding.Down, "0.14"},
		{"binary_strconv_difference", 2.675, 2, rounding.HalfAwayFromZero, "2.68"},
		{"negative_small_zero", -0.001, 2, rounding.HalfAwayFromZero, "0.00"},
		{"negative_tie_away", -0.005, 2, rounding.HalfAwayFromZero, "-0.01"},
		{"negative_tie_toward", -0.005, 2, rounding.HalfTowardZero, "0.00"},
		{"hundreds_half_even", 1250, -2, rounding.HalfEven, "1200"},
		{"hundreds_half_up", 1250, -2, rounding.HalfUp, "1300"},
		{"integer_half_even", 1234.5, 0, rounding.HalfEven, "1234"},
		{"subnormal_up", 5e-324, 3, rounding.Up, "0.001"},
		{"max_float_to_hundreds", math.MaxFloat64, -308, rounding.Up, "2" + strings.Repeat("0", 308)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.mode.FloatDecimal(tt.x, tt.places)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
	for _, tt := range []struct {
		name string
		x    float64
		want string
	}{
		{"two_to_60", float64(uint64(1) << 60), "1152921504606846976"},
		{"one_e_23", 1e23, "99999999999999991611392"},
	} {
		for i, mode := range testModes {
			t.Run(fmt.Sprintf("%s/%s", tt.name, testModeNames[i]), func(t *testing.T) {
				t.Parallel()
				got, err := mode.FloatDecimal(tt.x, 0)
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			})
		}
	}
}

func TestModeDecimalPlacesLimits(t *testing.T) {
	t.Parallel()
	for _, places := range []int{-rounding.MaxPlaces, rounding.MaxPlaces} {
		t.Run(fmt.Sprintf("valid_%d", places), func(t *testing.T) {
			t.Parallel()
			ratGot, err := rounding.HalfEven.RatDecimal(bigRat(1, 8), places)
			require.NoError(t, err)
			assert.Equal(t, expectedDecimal(rounding.HalfEven, bigRat(1, 8), places), ratGot)
			floatGot, err := rounding.HalfEven.FloatDecimal(math.MaxFloat64, places)
			require.NoError(t, err)
			assert.Equal(t, expectedFloatDecimal(rounding.HalfEven, math.MaxFloat64, places), floatGot)
			assert.LessOrEqual(t, len(floatGot), 711)
		})
	}
	for _, places := range []int{-rounding.MaxPlaces - 1, rounding.MaxPlaces + 1} {
		t.Run(fmt.Sprintf("invalid_%d", places), func(t *testing.T) {
			t.Parallel()
			ratGot, err := rounding.HalfEven.RatDecimal(bigRat(1, 8), places)
			checkError(t, err, rounding.ErrInvalidArgument)
			assert.Empty(t, ratGot)
			floatGot, err := rounding.HalfEven.FloatDecimal(0.145, places)
			checkError(t, err, rounding.ErrInvalidArgument)
			assert.Empty(t, floatGot)
			placesValue, err := rounding.HalfEven.FloatPlaces(0.145, places)
			checkError(t, err, rounding.ErrInvalidArgument)
			assert.Zero(t, placesValue)
		})
	}
}

func TestModeDecimalLongExactOutput(t *testing.T) {
	t.Parallel()
	for i, mode := range testModes {
		t.Run(testModeNames[i], func(t *testing.T) {
			t.Parallel()
			got, err := mode.FloatDecimal(math.MaxFloat64, rounding.MaxPlaces)
			require.NoError(t, err)
			assert.Equal(t, expectedFloatDecimal(mode, math.MaxFloat64, rounding.MaxPlaces), got)
			assert.LessOrEqual(t, len(got), 711)
		})
	}
}
