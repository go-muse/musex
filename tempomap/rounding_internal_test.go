package tempomap

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var internalHalfModes = []struct {
	name string
	mode rounding.Mode
}{
	{name: "half up", mode: rounding.HalfUp},
	{name: "half down", mode: rounding.HalfDown},
	{name: "half away from zero", mode: rounding.HalfAwayFromZero},
	{name: "half toward zero", mode: rounding.HalfTowardZero},
	{name: "half even", mode: rounding.HalfEven},
}

func assertInternalBigInt(t *testing.T, want int64, got *big.Int) {
	t.Helper()
	require.NotNil(t, got)
	assert.Equal(t, want, got.Int64())
}

func TestMakeRoundingAttemptMidpoints(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		num  int64
		den  int64
		want [5]int64
	}{
		{name: "exact one half", num: 1, den: 2, want: [5]int64{1, 0, 1, 0, 0}},
		{name: "exact three halves", num: 3, den: 2, want: [5]int64{2, 1, 2, 1, 2}},
		{name: "below one half", num: 499, den: 1000, want: [5]int64{0, 0, 0, 0, 0}},
		{name: "above one half", num: 501, den: 1000, want: [5]int64{1, 1, 1, 1, 1}},
		{name: "below three halves", num: 1499, den: 1000, want: [5]int64{1, 1, 1, 1, 1}},
		{name: "above three halves", num: 1501, den: 1000, want: [5]int64{2, 2, 2, 2, 2}},
	} {
		for i, mode := range internalHalfModes {
			t.Run(tt.name+"/"+mode.name, func(t *testing.T) {
				t.Parallel()
				lower := big.NewRat(tt.num, tt.den)
				upper := big.NewRat(tt.num, tt.den)
				lowerBefore := lower.RatString()
				upperBefore := upper.RatString()
				attempt := makeRoundingAttempt(64, 256, lower, upper, 0, mode.mode)
				assert.True(t, attempt.decided)
				assertInternalBigInt(t, tt.want[i], attempt.lowerK)
				assertInternalBigInt(t, tt.want[i], attempt.upperK)
				assert.Equal(t, lowerBefore, lower.RatString())
				assert.Equal(t, upperBefore, upper.RatString())
			})
		}
	}
}

func TestMakeRoundingAttemptRecomputesCandidatesAndLimit(t *testing.T) {
	t.Parallel()
	for _, mode := range internalHalfModes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			at64 := makeRoundingAttempt(64, 256, big.NewRat(499, 1000), big.NewRat(501, 1000), 0, mode.mode)
			assertInternalBigInt(t, 0, at64.lowerK)
			assertInternalBigInt(t, 1, at64.upperK)
			assert.False(t, at64.decided)
			assert.Equal(t, uint(2), at64.limit)

			at128 := makeRoundingAttempt(128, 256, big.NewRat(499, 1000), big.NewRat(501, 1000), 0, mode.mode)
			assertInternalBigInt(t, 0, at128.lowerK)
			assertInternalBigInt(t, 1, at128.upperK)
			assert.False(t, at128.decided)
			assert.Equal(t, uint(1), at128.limit)

			aboveAtMax := makeRoundingAttempt(256, 256, big.NewRat(501, 1000), big.NewRat(501, 1000), 0, mode.mode)
			assertInternalBigInt(t, 1, aboveAtMax.lowerK)
			assertInternalBigInt(t, 1, aboveAtMax.upperK)
			assert.True(t, aboveAtMax.decided)
			assert.Zero(t, aboveAtMax.limit)

			belowAtMax := makeRoundingAttempt(256, 256, big.NewRat(499, 1000), big.NewRat(499, 1000), 0, mode.mode)
			assertInternalBigInt(t, 0, belowAtMax.lowerK)
			assertInternalBigInt(t, 0, belowAtMax.upperK)
			assert.True(t, belowAtMax.decided)
			assert.Zero(t, belowAtMax.limit)
		})
	}
}

func TestRoundWithLimitFinishesAtMaxPrecision(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		num  int64
		den  int64
		want [5]int64
	}{
		{name: "exact midpoint", num: 1, den: 2, want: [5]int64{1, 0, 1, 0, 0}},
		{name: "above midpoint", num: 501, den: 1000, want: [5]int64{1, 1, 1, 1, 1}},
		{name: "below midpoint", num: 499, den: 1000, want: [5]int64{0, 0, 0, 0, 0}},
	} {
		for i, mode := range internalHalfModes {
			t.Run(tt.name+"/"+mode.name, func(t *testing.T) {
				t.Parallel()
				var calls []uint
				eval := func(precision uint) (lower, upper *big.Rat) {
					calls = append(calls, precision)
					if precision < 256 {
						return big.NewRat(499, 1000), big.NewRat(501, 1000)
					}
					return big.NewRat(tt.num, tt.den), big.NewRat(tt.num, tt.den)
				}
				var k *big.Int
				var usedPrecision, refinements uint
				var ok bool
				assert.NotPanics(t, func() {
					k, usedPrecision, refinements, ok = roundWithLimit(0, mode.mode, 64, 256, eval)
				})
				assertInternalBigInt(t, tt.want[i], k)
				assert.Equal(t, uint(256), usedPrecision)
				assert.Equal(t, uint(2), refinements)
				assert.True(t, ok)
				assert.Equal(t, []uint{64, 128, 256}, calls)
			})
		}
	}
}

func TestRoundWithLimitOverlappingIntervalReturnsFalseWithoutPanic(t *testing.T) {
	t.Parallel()
	for _, mode := range internalHalfModes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			var calls []uint
			eval := func(precision uint) (lower, upper *big.Rat) {
				calls = append(calls, precision)
				return big.NewRat(499, 1000), big.NewRat(501, 1000)
			}
			var k *big.Int
			var usedPrecision, refinements uint
			var ok bool
			assert.NotPanics(t, func() {
				k, usedPrecision, refinements, ok = roundWithLimit(0, mode.mode, 64, 256, eval)
			})
			assert.Nil(t, k)
			assert.Equal(t, uint(256), usedPrecision)
			assert.Equal(t, uint(2), refinements)
			assert.False(t, ok)
			assert.Equal(t, []uint{64, 128, 256}, calls)
		})
	}
}
