package rounding_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzChoose(f *testing.F) {
	f.Add(uint8(1), -1, int8(-1), false)
	f.Add(uint8(9), 0, int8(1), true)
	f.Add(uint8(0), 0, int8(0), false)
	f.Add(uint8(255), math.MaxInt, int8(127), true)
	f.Fuzz(func(t *testing.T, rawMode uint8, comparison int, rawSide int8, lowerEven bool) {
		mode, side := rounding.Mode(rawMode), rounding.Side(rawSide)
		got, err := mode.Choose(comparison, side, lowerEven)
		switch {
		case mode == 0:
			checkError(t, err, rounding.ErrUnset)
			assert.Zero(t, got)
		case mode < rounding.Down || mode > rounding.HalfEven:
			checkError(t, err, rounding.ErrUnknownMode)
			assert.Zero(t, got)
		case side != rounding.Negative && side != rounding.Positive:
			checkError(t, err, rounding.ErrInvalidArgument)
			assert.Zero(t, got)
		default:
			require.NoError(t, err)
			assert.Equal(t, expectedNeighbor(mode, comparison, side, lowerEven), got)
		}
	})
}

func FuzzQuoRem(f *testing.F) {
	f.Add(uint8(1), 7, 3, 8)
	f.Add(uint8(9), math.MaxInt, 2, math.MinInt)
	f.Add(uint8(5), math.MinInt, -1, 0)
	f.Add(uint8(9), -163, -55, 8)
	f.Fuzz(func(t *testing.T, rawMode uint8, n, d, otherN int) {
		mode := rounding.Mode(rawMode)
		q, r, err := mode.QuoRem(n, d)
		switch {
		case mode == 0:
			checkError(t, err, rounding.ErrUnset)
			assert.Zero(t, q)
			assert.Zero(t, r)
			return
		case mode < rounding.Down || mode > rounding.HalfEven:
			checkError(t, err, rounding.ErrUnknownMode)
			assert.Zero(t, q)
			assert.Zero(t, r)
			return
		case d == 0:
			checkError(t, err, rounding.ErrInvalidArgument)
			assert.Zero(t, q)
			assert.Zero(t, r)
			return
		}
		exact := new(big.Rat).SetFrac(bigInt(n), bigInt(d))
		wantQ := roundRat(mode, exact)
		if !intInRange(wantQ) {
			checkError(t, err, rounding.ErrOutOfRange)
			assert.Zero(t, q)
			assert.Zero(t, r)
			return
		}
		require.NoError(t, err)
		assert.Equal(t, int(wantQ.Int64()), q)
		wantR := new(big.Int).Sub(bigInt(n), new(big.Int).Mul(wantQ, bigInt(d)))
		assert.Equal(t, int(wantR.Int64()), r)
		absoluteR := new(big.Int).Abs(bigInt(r))
		assert.Equal(t, -1, absoluteR.Cmp(new(big.Int).Abs(bigInt(d))))

		otherQ, _, otherErr := mode.QuoRem(otherN, d)
		otherWantQ := roundRat(mode, new(big.Rat).SetFrac(bigInt(otherN), bigInt(d)))
		if intInRange(otherWantQ) {
			require.NoError(t, otherErr)
			// A negative divisor reverses the order of exact quotients.
			if (n <= otherN) == (d > 0) {
				assert.LessOrEqual(t, q, otherQ)
			} else {
				assert.GreaterOrEqual(t, q, otherQ)
			}
		} else {
			checkError(t, otherErr, rounding.ErrOutOfRange)
		}

		if n != math.MinInt {
			mirrorQ, _, mirrorErr := mode.Mirror().QuoRem(-n, d)
			if intInRange(new(big.Int).Neg(wantQ)) {
				require.NoError(t, mirrorErr)
				assert.Equal(t, -q, mirrorQ)
			} else {
				checkError(t, mirrorErr, rounding.ErrOutOfRange)
			}
		}
	})
}

func FuzzFloatStepsAndMultiple(f *testing.F) {
	f.Add(uint8(5), 0.15, 0.1, 0.25)
	f.Add(uint8(9), -2.5, 0.5, 2.5)
	f.Add(uint8(1), 1e308, 1e-308, 1.0)
	f.Fuzz(func(t *testing.T, rawMode uint8, x, step, otherX float64) {
		mode := rounding.Mode(rawMode)
		if mode < rounding.Down || mode > rounding.HalfEven || math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(step) || math.IsInf(step, 0) || step <= 0 {
			return
		}
		gotSteps, stepsErr := mode.FloatSteps(x, step)
		wantSteps := expectedFloatSteps(mode, x, step)
		if intInRange(wantSteps) {
			require.NoError(t, stepsErr)
			assert.Equal(t, int(wantSteps.Int64()), gotSteps)
		} else {
			checkError(t, stepsErr, rounding.ErrOutOfRange)
			assert.Zero(t, gotSteps)
		}
		gotMultiple, multipleErr := mode.FloatMultiple(x, step)
		wantMultiple, overflow := expectedFloatMultiple(mode, x, step)
		if overflow {
			checkError(t, multipleErr, rounding.ErrOutOfRange)
			assert.Zero(t, gotMultiple)
		} else {
			require.NoError(t, multipleErr)
			assert.Equal(t, wantMultiple, gotMultiple)
		}

		mirrorValue, mirrorErr := mode.Mirror().FloatMultiple(-x, step)
		if overflow {
			checkError(t, mirrorErr, rounding.ErrOutOfRange)
		} else {
			require.NoError(t, mirrorErr)
			assert.Equal(t, -gotMultiple, mirrorValue)
		}
		if math.IsNaN(otherX) || math.IsInf(otherX, 0) {
			return
		}
		otherValue, otherErr := mode.FloatMultiple(otherX, step)
		_, otherOverflow := expectedFloatMultiple(mode, otherX, step)
		if !overflow && !otherOverflow {
			require.NoError(t, otherErr)
			if x <= otherX {
				assert.LessOrEqual(t, gotMultiple, otherValue)
			} else {
				assert.GreaterOrEqual(t, gotMultiple, otherValue)
			}
		}
		down, downErr := rounding.Down.FloatMultiple(x, step)
		up, upErr := rounding.Up.FloatMultiple(x, step)
		if downErr == nil && upErr == nil {
			assert.LessOrEqual(t, down, x)
			assert.GreaterOrEqual(t, up, x)
		}
	})
}

func FuzzFloatPlacesAndDecimal(f *testing.F) {
	f.Add(uint8(5), 0.145, 2, 0.15)
	f.Add(uint8(9), 1250.0, -2, 1200.0)
	f.Add(uint8(6), math.MaxFloat64, rounding.MaxPlaces, 1.0)
	f.Fuzz(func(t *testing.T, rawMode uint8, x float64, places int, otherX float64) {
		mode := rounding.Mode(rawMode)
		if mode < rounding.Down || mode > rounding.HalfEven || math.IsNaN(x) || math.IsInf(x, 0) {
			return
		}
		places = max(-rounding.MaxPlaces, min(rounding.MaxPlaces, places))
		got, err := mode.FloatPlaces(x, places)
		want, overflow := expectedFloatPlaces(mode, x, places)
		if overflow {
			checkError(t, err, rounding.ErrOutOfRange)
			assert.Zero(t, got)
		} else {
			require.NoError(t, err)
			assert.Equal(t, want, got)
		}
		decimal, err := mode.FloatDecimal(x, places)
		require.NoError(t, err)
		assert.Equal(t, expectedFloatDecimal(mode, x, places), decimal)

		mirror, mirrorErr := mode.Mirror().FloatPlaces(-x, places)
		if overflow {
			checkError(t, mirrorErr, rounding.ErrOutOfRange)
		} else {
			require.NoError(t, mirrorErr)
			assert.Equal(t, -got, mirror)
		}
		mirrorDecimal, err := mode.Mirror().FloatDecimal(-x, places)
		require.NoError(t, err)
		wantMirrorDecimal := expectedFloatDecimal(mode.Mirror(), -x, places)
		assert.Equal(t, wantMirrorDecimal, mirrorDecimal)

		if math.IsNaN(otherX) || math.IsInf(otherX, 0) {
			return
		}
		other, otherErr := mode.FloatPlaces(otherX, places)
		_, otherOverflow := expectedFloatPlaces(mode, otherX, places)
		if !overflow && !otherOverflow {
			require.NoError(t, otherErr)
			if x <= otherX {
				assert.LessOrEqual(t, got, other)
			} else {
				assert.GreaterOrEqual(t, got, other)
			}
		}
		down, downErr := rounding.Down.FloatPlaces(x, places)
		up, upErr := rounding.Up.FloatPlaces(x, places)
		if downErr == nil && upErr == nil {
			assert.LessOrEqual(t, down, x)
			assert.GreaterOrEqual(t, up, x)
		}
	})
}

func FuzzFloatFractionAndNumerator(f *testing.F) {
	f.Add(uint8(6), 0.5, 3, 0.25)
	f.Add(uint8(9), -2.5, 2, 2.5)
	f.Add(uint8(1), 1e300, 1, 1.0)
	f.Fuzz(func(t *testing.T, rawMode uint8, x float64, den int, otherX float64) {
		mode := rounding.Mode(rawMode)
		if mode < rounding.Down || mode > rounding.HalfEven || den <= 0 || math.IsNaN(x) || math.IsInf(x, 0) {
			return
		}
		got, err := mode.FloatFraction(x, den)
		require.NoError(t, err)
		assert.Equal(t, expectedFloatFraction(mode, x, den), got)
		numerator, numeratorErr := mode.FloatNumerator(x, den)
		wantNumerator := expectedFloatNumerator(mode, x, den)
		if intInRange(wantNumerator) {
			require.NoError(t, numeratorErr)
			assert.Equal(t, int(wantNumerator.Int64()), numerator)
		} else {
			checkError(t, numeratorErr, rounding.ErrOutOfRange)
			assert.Zero(t, numerator)
		}

		mirror, mirrorErr := mode.Mirror().FloatFraction(-x, den)
		require.NoError(t, mirrorErr)
		assert.Equal(t, -got, mirror)
		if math.IsNaN(otherX) || math.IsInf(otherX, 0) {
			return
		}
		other, otherErr := mode.FloatFraction(otherX, den)
		require.NoError(t, otherErr)
		if x <= otherX {
			assert.LessOrEqual(t, got, other)
		} else {
			assert.GreaterOrEqual(t, got, other)
		}
		down, downErr := rounding.Down.FloatFraction(x, den)
		up, upErr := rounding.Up.FloatFraction(x, den)
		require.NoError(t, downErr)
		require.NoError(t, upErr)
		assert.LessOrEqual(t, down, x)
		assert.GreaterOrEqual(t, up, x)
	})
}

func FuzzFloatIntAndFloat(f *testing.F) {
	f.Add(uint8(5), 0.5, 1.5)
	f.Add(uint8(9), -2.5, 2.5)
	f.Add(uint8(1), 1e300, 0.0)
	f.Fuzz(func(t *testing.T, rawMode uint8, x, otherX float64) {
		mode := rounding.Mode(rawMode)
		if mode < rounding.Down || mode > rounding.HalfEven || math.IsNaN(x) || math.IsInf(x, 0) {
			return
		}
		got, err := mode.Float(x)
		require.NoError(t, err)
		assert.Equal(t, expectedFloat(mode, x), got)
		integer, integerErr := mode.FloatInt(x)
		wantInteger := expectedFloatInt(mode, x)
		if intInRange(wantInteger) {
			require.NoError(t, integerErr)
			assert.Equal(t, int(wantInteger.Int64()), integer)
			assert.Equal(t, float64(integer), got)
		} else {
			checkError(t, integerErr, rounding.ErrOutOfRange)
			assert.Zero(t, integer)
		}
		mirror, mirrorErr := mode.Mirror().Float(-x)
		require.NoError(t, mirrorErr)
		assert.Equal(t, -got, mirror)
		if math.IsNaN(otherX) || math.IsInf(otherX, 0) {
			return
		}
		other, otherErr := mode.Float(otherX)
		require.NoError(t, otherErr)
		if x <= otherX {
			assert.LessOrEqual(t, got, other)
		} else {
			assert.GreaterOrEqual(t, got, other)
		}
		down, downErr := rounding.Down.Float(x)
		up, upErr := rounding.Up.Float(x)
		require.NoError(t, downErr)
		require.NoError(t, upErr)
		assert.LessOrEqual(t, down, x)
		assert.GreaterOrEqual(t, up, x)
	})
}
