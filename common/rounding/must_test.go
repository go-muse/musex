package rounding_test

import (
	"errors"
	"math"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustVariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		call     func() ([]any, error)
		must     func() []any
		failCall func() ([]any, error)
		failMust func() []any
	}{
		{
			name:     "MustParse",
			call:     func() ([]any, error) { v, err := rounding.Parse("half-even"); return []any{v}, err },
			must:     func() []any { return []any{rounding.MustParse("half-even")} },
			failCall: func() ([]any, error) { v, err := rounding.Parse("HALF-EVEN"); return []any{v}, err },
			failMust: func() []any { return []any{rounding.MustParse("HALF-EVEN")} },
		},
		{
			name:     "MustName",
			call:     func() ([]any, error) { v, err := rounding.HalfEven.Name(); return []any{v}, err },
			must:     func() []any { return []any{rounding.HalfEven.MustName()} },
			failCall: func() ([]any, error) { v, err := rounding.Mode(0).Name(); return []any{v}, err },
			failMust: func() []any { return []any{rounding.Mode(0).MustName()} },
		},
		{
			name: "MustChoose",
			call: func() ([]any, error) {
				v, err := rounding.HalfEven.Choose(0, rounding.Positive, false)
				return []any{v}, err
			},
			must: func() []any { return []any{rounding.HalfEven.MustChoose(0, rounding.Positive, false)} },
			failCall: func() ([]any, error) {
				v, err := rounding.Mode(0).Choose(0, rounding.Positive, false)
				return []any{v}, err
			},
			failMust: func() []any { return []any{rounding.Mode(0).MustChoose(0, rounding.Positive, false)} },
		},
		{
			name:     "MustChooseIndex",
			call:     func() ([]any, error) { v, err := rounding.HalfDown.ChooseIndex(0, 0); return []any{v}, err },
			must:     func() []any { return []any{rounding.HalfDown.MustChooseIndex(0, 0)} },
			failCall: func() ([]any, error) { v, err := rounding.Up.ChooseIndex(math.MaxInt, 0); return []any{v}, err },
			failMust: func() []any { return []any{rounding.Up.MustChooseIndex(math.MaxInt, 0)} },
		},
		{
			name:     "MustQuoRem",
			call:     func() ([]any, error) { q, r, err := rounding.HalfEven.QuoRem(7, 2); return []any{q, r}, err },
			must:     func() []any { q, r := rounding.HalfEven.MustQuoRem(7, 2); return []any{q, r} },
			failCall: func() ([]any, error) { q, r, err := rounding.Down.QuoRem(7, 0); return []any{q, r}, err },
			failMust: func() []any { q, r := rounding.Down.MustQuoRem(7, 0); return []any{q, r} },
		},
		{
			name:     "MustIntMultiple",
			call:     func() ([]any, error) { v, err := rounding.HalfDown.IntMultiple(1200, 480); return []any{v}, err },
			must:     func() []any { return []any{rounding.HalfDown.MustIntMultiple(1200, 480)} },
			failCall: func() ([]any, error) { v, err := rounding.Up.IntMultiple(math.MaxInt, 2); return []any{v}, err },
			failMust: func() []any { return []any{rounding.Up.MustIntMultiple(math.MaxInt, 2)} },
		},
		{
			name: "MustQuoRemBig",
			call: func() ([]any, error) {
				q, r, err := rounding.HalfEven.QuoRemBig(bigInt(7), bigInt(2))
				return []any{q, r}, err
			},
			must:     func() []any { q, r := rounding.HalfEven.MustQuoRemBig(bigInt(7), bigInt(2)); return []any{q, r} },
			failCall: func() ([]any, error) { q, r, err := rounding.Down.QuoRemBig(nil, bigInt(2)); return []any{q, r}, err },
			failMust: func() []any { q, r := rounding.Down.MustQuoRemBig(nil, bigInt(2)); return []any{q, r} },
		},
		{
			name:     "MustRatInt",
			call:     func() ([]any, error) { v, err := rounding.HalfUp.RatInt(bigRat(7, 3)); return []any{v}, err },
			must:     func() []any { return []any{rounding.HalfUp.MustRatInt(bigRat(7, 3))} },
			failCall: func() ([]any, error) { v, err := rounding.Down.RatInt(nil); return []any{v}, err },
			failMust: func() []any { return []any{rounding.Down.MustRatInt(nil)} },
		},
		{
			name: "MustRatMultiple",
			call: func() ([]any, error) {
				k, v, err := rounding.HalfEven.RatMultiple(bigRat(5, 8), bigRat(1, 4))
				return []any{k, v}, err
			},
			must: func() []any {
				k, v := rounding.HalfEven.MustRatMultiple(bigRat(5, 8), bigRat(1, 4))
				return []any{k, v}
			},
			failCall: func() ([]any, error) {
				k, v, err := rounding.Down.RatMultiple(bigRat(5, 8), bigRat(0, 1))
				return []any{k, v}, err
			},
			failMust: func() []any { k, v := rounding.Down.MustRatMultiple(bigRat(5, 8), bigRat(0, 1)); return []any{k, v} },
		},
		{
			name: "MustRatDecimal",
			call: func() ([]any, error) { v, err := rounding.HalfEven.RatDecimal(bigRat(1, 8), 2); return []any{v}, err },
			must: func() []any { return []any{rounding.HalfEven.MustRatDecimal(bigRat(1, 8), 2)} },
			failCall: func() ([]any, error) {
				v, err := rounding.HalfEven.RatDecimal(bigRat(1, 8), rounding.MaxPlaces+1)
				return []any{v}, err
			},
			failMust: func() []any { return []any{rounding.HalfEven.MustRatDecimal(bigRat(1, 8), rounding.MaxPlaces+1)} },
		},
		{
			name:     "MustFloat",
			call:     func() ([]any, error) { v, err := rounding.HalfEven.Float(2.5); return []any{v}, err },
			must:     func() []any { return []any{rounding.HalfEven.MustFloat(2.5)} },
			failCall: func() ([]any, error) { v, err := rounding.HalfEven.Float(math.NaN()); return []any{v}, err },
			failMust: func() []any { return []any{rounding.HalfEven.MustFloat(math.NaN())} },
		},
		{
			name:     "MustFloatInt",
			call:     func() ([]any, error) { v, err := rounding.HalfEven.FloatInt(-2.5); return []any{v}, err },
			must:     func() []any { return []any{rounding.HalfEven.MustFloatInt(-2.5)} },
			failCall: func() ([]any, error) { v, err := rounding.Down.FloatInt(9223372036854775807.0); return []any{v}, err },
			failMust: func() []any { return []any{rounding.Down.MustFloatInt(9223372036854775807.0)} },
		},
		{
			name:     "MustFloatMultiple",
			call:     func() ([]any, error) { v, err := rounding.HalfUp.FloatMultiple(0.15, 0.1); return []any{v}, err },
			must:     func() []any { return []any{rounding.HalfUp.MustFloatMultiple(0.15, 0.1)} },
			failCall: func() ([]any, error) { v, err := rounding.HalfUp.FloatMultiple(0.15, 0); return []any{v}, err },
			failMust: func() []any { return []any{rounding.HalfUp.MustFloatMultiple(0.15, 0)} },
		},
		{
			name:     "MustFloatSteps",
			call:     func() ([]any, error) { v, err := rounding.HalfUp.FloatSteps(0.15, 0.1); return []any{v}, err },
			must:     func() []any { return []any{rounding.HalfUp.MustFloatSteps(0.15, 0.1)} },
			failCall: func() ([]any, error) { v, err := rounding.Down.FloatSteps(1e308, 1e-308); return []any{v}, err },
			failMust: func() []any { return []any{rounding.Down.MustFloatSteps(1e308, 1e-308)} },
		},
		{
			name:     "MustFloatPlaces",
			call:     func() ([]any, error) { v, err := rounding.HalfUp.FloatPlaces(0.145, 2); return []any{v}, err },
			must:     func() []any { return []any{rounding.HalfUp.MustFloatPlaces(0.145, 2)} },
			failCall: func() ([]any, error) { v, err := rounding.Up.FloatPlaces(math.MaxFloat64, -308); return []any{v}, err },
			failMust: func() []any { return []any{rounding.Up.MustFloatPlaces(math.MaxFloat64, -308)} },
		},
		{
			name: "MustFloatDecimal",
			call: func() ([]any, error) { v, err := rounding.HalfUp.FloatDecimal(0.145, 2); return []any{v}, err },
			must: func() []any { return []any{rounding.HalfUp.MustFloatDecimal(0.145, 2)} },
			failCall: func() ([]any, error) {
				v, err := rounding.HalfUp.FloatDecimal(0.145, rounding.MaxPlaces+1)
				return []any{v}, err
			},
			failMust: func() []any { return []any{rounding.HalfUp.MustFloatDecimal(0.145, rounding.MaxPlaces+1)} },
		},
		{
			name:     "MustFloatFraction",
			call:     func() ([]any, error) { v, err := rounding.HalfEven.FloatFraction(0.5, 3); return []any{v}, err },
			must:     func() []any { return []any{rounding.HalfEven.MustFloatFraction(0.5, 3)} },
			failCall: func() ([]any, error) { v, err := rounding.HalfEven.FloatFraction(0.5, 0); return []any{v}, err },
			failMust: func() []any { return []any{rounding.HalfEven.MustFloatFraction(0.5, 0)} },
		},
		{
			name:     "MustFloatNumerator",
			call:     func() ([]any, error) { v, err := rounding.HalfEven.FloatNumerator(0.5, 3); return []any{v}, err },
			must:     func() []any { return []any{rounding.HalfEven.MustFloatNumerator(0.5, 3)} },
			failCall: func() ([]any, error) { v, err := rounding.Down.FloatNumerator(1e300, 1); return []any{v}, err },
			failMust: func() []any { return []any{rounding.Down.MustFloatNumerator(1e300, 1)} },
		},
	}
	require.Len(t, tests, 18)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			want, err := tt.call()
			require.NoError(t, err)
			assert.Equal(t, want, tt.must())
			failureResults, failureErr := tt.failCall()
			require.Error(t, failureErr)
			for _, result := range failureResults {
				assert.Zero(t, result)
			}
			func() {
				defer func() {
					recovered := recover()
					require.NotNil(t, recovered)
					panicErr, ok := recovered.(error)
					require.True(t, ok)
					for _, sentinel := range []error{rounding.ErrUnset, rounding.ErrUnknownMode, rounding.ErrInvalidArgument, rounding.ErrOutOfRange} {
						assert.Equal(t, errors.Is(failureErr, sentinel), errors.Is(panicErr, sentinel))
					}
				}()
				tt.failMust()
			}()
		})
	}
}
