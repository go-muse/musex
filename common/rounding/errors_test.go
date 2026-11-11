package rounding_test

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
)

func TestErrorSentinelsAreDistinct(t *testing.T) {
	t.Parallel()
	sentinels := []error{rounding.ErrUnset, rounding.ErrUnknownMode, rounding.ErrInvalidArgument, rounding.ErrOutOfRange}
	for _, sentinel := range sentinels {
		checkError(t, sentinel, sentinel)
		for _, other := range sentinels {
			assert.Equal(t, sentinel == other, errors.Is(sentinel, other))
		}
	}
}

func TestFloatOperationsRejectNonFiniteInput(t *testing.T) {
	t.Parallel()
	values := []float64{math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, x := range values {
		for _, tt := range []struct {
			name string
			call func() (any, error)
		}{
			{"float", func() (any, error) { return rounding.HalfEven.Float(x) }},
			{"float_int", func() (any, error) { return rounding.HalfEven.FloatInt(x) }},
			{"float_multiple", func() (any, error) { return rounding.HalfEven.FloatMultiple(x, 0.1) }},
			{"float_steps", func() (any, error) { return rounding.HalfEven.FloatSteps(x, 0.1) }},
			{"float_places", func() (any, error) { return rounding.HalfEven.FloatPlaces(x, 2) }},
			{"float_decimal", func() (any, error) { return rounding.HalfEven.FloatDecimal(x, 2) }},
			{"float_fraction", func() (any, error) { return rounding.HalfEven.FloatFraction(x, 3) }},
			{"float_numerator", func() (any, error) { return rounding.HalfEven.FloatNumerator(x, 3) }},
		} {
			t.Run(fmt.Sprintf("%v/%s", x, tt.name), func(t *testing.T) {
				t.Parallel()
				got, err := tt.call()
				checkError(t, err, rounding.ErrInvalidArgument)
				assert.Zero(t, got)
			})
		}
	}
}

func TestFloatMultipleAndStepsRejectInvalidStep(t *testing.T) {
	t.Parallel()
	steps := []float64{0, math.Copysign(0, -1), -0.1, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, step := range steps {
		t.Run(fmt.Sprintf("step_%v", step), func(t *testing.T) {
			t.Parallel()
			got, err := rounding.HalfEven.FloatMultiple(0.15, step)
			checkError(t, err, rounding.ErrInvalidArgument)
			assert.Zero(t, got)
			index, err := rounding.HalfEven.FloatSteps(0.15, step)
			checkError(t, err, rounding.ErrInvalidArgument)
			assert.Zero(t, index)
		})
	}
}

func TestErrorPriority(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		mode rounding.Mode
		want error
	}{{0, rounding.ErrUnset}, {10, rounding.ErrUnknownMode}, {255, rounding.ErrUnknownMode}} {
		t.Run(fmt.Sprintf("mode_%d", tt.mode), func(t *testing.T) {
			t.Parallel()
			got, err := tt.mode.Float(math.NaN())
			checkError(t, err, tt.want)
			assert.Zero(t, got)
			multiple, err := tt.mode.FloatMultiple(math.NaN(), math.NaN())
			checkError(t, err, tt.want)
			assert.Zero(t, multiple)
			decimal, err := tt.mode.FloatDecimal(math.NaN(), rounding.MaxPlaces+1)
			checkError(t, err, tt.want)
			assert.Empty(t, decimal)
		})
	}
}
