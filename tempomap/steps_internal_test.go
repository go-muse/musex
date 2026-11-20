package tempomap

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests of this file cover the internal seam of architecture 14.4 of task
// 28e.1: makeStepAttempt and roundStepWithLimit behave as makeRoundingAttempt
// and roundWithLimit of 7.2 with the grid k·step instead of k/10^places.

var stepInternalAllModes = []struct {
	name string
	mode rounding.Mode
}{
	{name: "down", mode: rounding.Down},
	{name: "up", mode: rounding.Up},
	{name: "toward zero", mode: rounding.TowardZero},
	{name: "away from zero", mode: rounding.AwayFromZero},
	{name: "half up", mode: rounding.HalfUp},
	{name: "half down", mode: rounding.HalfDown},
	{name: "half away from zero", mode: rounding.HalfAwayFromZero},
	{name: "half toward zero", mode: rounding.HalfTowardZero},
	{name: "half even", mode: rounding.HalfEven},
}

func stepPower(places int) *big.Rat {
	return new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil))
}

// TestMakeStepAttemptUnitStepMidpoints repeats the midpoint table of 7.2 with step = 1.
func TestMakeStepAttemptUnitStepMidpoints(t *testing.T) {
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
				step := big.NewRat(1, 1)
				attempt := makeStepAttempt(64, 256, lower, upper, step, mode.mode)
				assert.True(t, attempt.decided)
				assertInternalBigInt(t, tt.want[i], attempt.lowerK)
				assertInternalBigInt(t, tt.want[i], attempt.upperK)
				assert.Equal(t, big.NewRat(tt.num, tt.den).RatString(), lower.RatString())
				assert.Equal(t, big.NewRat(tt.num, tt.den).RatString(), upper.RatString())
				assert.Equal(t, "1", step.RatString())
			})
		}
	}
}

// TestMakeStepAttemptRecomputesCandidatesAndLimit repeats the scenario of 7.2 with step = 1.
func TestMakeStepAttemptRecomputesCandidatesAndLimit(t *testing.T) {
	t.Parallel()
	one := big.NewRat(1, 1)
	for _, mode := range internalHalfModes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			at64 := makeStepAttempt(64, 256, big.NewRat(499, 1000), big.NewRat(501, 1000), one, mode.mode)
			assertInternalBigInt(t, 0, at64.lowerK)
			assertInternalBigInt(t, 1, at64.upperK)
			assert.False(t, at64.decided)
			assert.Equal(t, uint(2), at64.limit)

			at128 := makeStepAttempt(128, 256, big.NewRat(499, 1000), big.NewRat(501, 1000), one, mode.mode)
			assertInternalBigInt(t, 0, at128.lowerK)
			assertInternalBigInt(t, 1, at128.upperK)
			assert.False(t, at128.decided)
			assert.Equal(t, uint(1), at128.limit)

			aboveAtMax := makeStepAttempt(256, 256, big.NewRat(501, 1000), big.NewRat(501, 1000), one, mode.mode)
			assertInternalBigInt(t, 1, aboveAtMax.lowerK)
			assertInternalBigInt(t, 1, aboveAtMax.upperK)
			assert.True(t, aboveAtMax.decided)
			assert.Zero(t, aboveAtMax.limit)

			belowAtMax := makeStepAttempt(256, 256, big.NewRat(499, 1000), big.NewRat(499, 1000), one, mode.mode)
			assertInternalBigInt(t, 0, belowAtMax.lowerK)
			assertInternalBigInt(t, 0, belowAtMax.upperK)
			assert.True(t, belowAtMax.decided)
			assert.Zero(t, belowAtMax.limit)
		})
	}
	assert.Equal(t, "1", one.RatString())
}

// TestRoundStepWithLimitUnitStep repeats "termination exactly at maxPrecision" of 7.2 with step = 1.
func TestRoundStepWithLimitUnitStep(t *testing.T) {
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
					k, usedPrecision, refinements, ok = roundStepWithLimit(big.NewRat(1, 1), mode.mode, 64, 256, eval)
				})
				assertInternalBigInt(t, tt.want[i], k)
				assert.Equal(t, uint(256), usedPrecision)
				assert.Equal(t, uint(2), refinements)
				assert.True(t, ok)
				assert.Equal(t, []uint{64, 128, 256}, calls)
			})
		}
	}
	for _, mode := range internalHalfModes {
		t.Run("overlapping/"+mode.name, func(t *testing.T) {
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
				k, usedPrecision, refinements, ok = roundStepWithLimit(big.NewRat(1, 1), mode.mode, 64, 256, eval)
			})
			assert.Nil(t, k)
			assert.Equal(t, uint(256), usedPrecision)
			assert.Equal(t, uint(2), refinements)
			assert.False(t, ok)
			assert.Equal(t, []uint{64, 128, 256}, calls)
		})
	}
}

// TestMakeStepAttemptThirdStep checks the step = 1/3 expectations of 14.4.
func TestMakeStepAttemptThirdStep(t *testing.T) {
	t.Parallel()
	third := big.NewRat(1, 3)
	for _, tt := range []struct {
		name string
		x    *big.Rat
		want map[rounding.Mode]int64
	}{
		{
			name: "1/6 is the midpoint of nodes 0 and 1",
			x:    big.NewRat(1, 6),
			want: map[rounding.Mode]int64{
				rounding.HalfUp: 1, rounding.HalfDown: 0, rounding.HalfAwayFromZero: 1, rounding.HalfTowardZero: 0, rounding.HalfEven: 0,
				rounding.Down: 0, rounding.Up: 1, rounding.TowardZero: 0, rounding.AwayFromZero: 1,
			},
		},
		{
			name: "1/2 is the midpoint of nodes 1 and 2",
			x:    big.NewRat(1, 2),
			want: map[rounding.Mode]int64{
				rounding.HalfUp: 2, rounding.HalfDown: 1, rounding.HalfAwayFromZero: 2, rounding.HalfTowardZero: 1, rounding.HalfEven: 2,
				rounding.Down: 1, rounding.Up: 2, rounding.TowardZero: 1, rounding.AwayFromZero: 2,
			},
		},
		{
			name: "1/3 is node 1",
			x:    big.NewRat(1, 3),
			want: map[rounding.Mode]int64{
				rounding.HalfUp: 1, rounding.HalfDown: 1, rounding.HalfAwayFromZero: 1, rounding.HalfTowardZero: 1, rounding.HalfEven: 1,
				rounding.Down: 1, rounding.Up: 1, rounding.TowardZero: 1, rounding.AwayFromZero: 1,
			},
		},
		{
			name: "2/5 is between nodes 1 and 2",
			x:    big.NewRat(2, 5),
			want: map[rounding.Mode]int64{
				rounding.HalfUp: 1, rounding.HalfDown: 1, rounding.HalfAwayFromZero: 1, rounding.HalfTowardZero: 1, rounding.HalfEven: 1,
				rounding.Down: 1, rounding.Up: 2, rounding.TowardZero: 1, rounding.AwayFromZero: 2,
			},
		},
	} {
		for _, mode := range stepInternalAllModes {
			t.Run(tt.name+"/"+mode.name, func(t *testing.T) {
				t.Parallel()
				lower := new(big.Rat).Set(tt.x)
				upper := new(big.Rat).Set(tt.x)
				attempt := makeStepAttempt(64, 256, lower, upper, third, mode.mode)
				assert.True(t, attempt.decided)
				assertInternalBigInt(t, tt.want[mode.mode], attempt.lowerK)
				assertInternalBigInt(t, tt.want[mode.mode], attempt.upperK)
				assert.Equal(t, tt.x.RatString(), lower.RatString())
				assert.Equal(t, tt.x.RatString(), upper.RatString())
			})
		}
	}
	assert.Equal(t, "1/3", third.RatString())
}

// TestStepSeamEquivalentToDecimalSeam checks the equivalence of 14.4:
// step = 10^−D gives the fields and results of the decimal seam of 7.2.
func TestStepSeamEquivalentToDecimalSeam(t *testing.T) {
	t.Parallel()
	inputs := []struct {
		name         string
		lower, upper *big.Rat
	}{
		{name: "interval around one half", lower: big.NewRat(499, 1000), upper: big.NewRat(501, 1000)},
		{name: "point one half", lower: big.NewRat(1, 2), upper: big.NewRat(1, 2)},
		{name: "point 2773/1000", lower: big.NewRat(2773, 1000), upper: big.NewRat(2773, 1000)},
	}
	for _, places := range []int{0, 3} {
		for _, in := range inputs {
			for _, mode := range stepInternalAllModes {
				t.Run(in.name+"/"+mode.name+"/"+string(rune('0'+places)), func(t *testing.T) {
					t.Parallel()
					step := stepPower(places)
					for _, precision := range []uint{64, 128, 256} {
						want := makeRoundingAttempt(precision, 256, in.lower, in.upper, places, mode.mode)
						got := makeStepAttempt(precision, 256, in.lower, in.upper, step, mode.mode)
						require.NotNil(t, want.lowerK)
						require.NotNil(t, got.lowerK)
						assert.Equal(t, 0, want.lowerK.Cmp(got.lowerK), "lowerK")
						assert.Equal(t, 0, want.upperK.Cmp(got.upperK), "upperK")
						assert.Equal(t, want.decided, got.decided)
						assert.Equal(t, want.limit, got.limit)
					}

					makeEval := func() func(uint) (*big.Rat, *big.Rat) {
						return func(precision uint) (*big.Rat, *big.Rat) {
							if precision < 256 {
								return new(big.Rat).Set(in.lower), new(big.Rat).Set(in.upper)
							}
							return big.NewRat(2773, 1000), big.NewRat(2773, 1000)
						}
					}
					wantK, wantUsed, wantRefinements, wantOK := roundWithLimit(places, mode.mode, 64, 256, makeEval())
					gotK, gotUsed, gotRefinements, gotOK := roundStepWithLimit(step, mode.mode, 64, 256, makeEval())
					assert.Equal(t, wantOK, gotOK)
					assert.Equal(t, wantUsed, gotUsed)
					assert.Equal(t, wantRefinements, gotRefinements)
					if wantK == nil {
						assert.Nil(t, gotK)
					} else {
						require.NotNil(t, gotK)
						assert.Equal(t, 0, wantK.Cmp(gotK))
					}
					assert.Equal(t, stepPower(places).RatString(), step.RatString(), "step changed")
				})
			}
		}
	}
	for _, in := range inputs {
		assert.True(t, in.lower.Cmp(in.upper) <= 0)
	}
	assert.Equal(t, "499/1000", inputs[0].lower.RatString())
	assert.Equal(t, "501/1000", inputs[0].upper.RatString())
}
