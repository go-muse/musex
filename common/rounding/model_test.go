package rounding_test

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testModes = []rounding.Mode{
	rounding.Down,
	rounding.Up,
	rounding.TowardZero,
	rounding.AwayFromZero,
	rounding.HalfUp,
	rounding.HalfDown,
	rounding.HalfAwayFromZero,
	rounding.HalfTowardZero,
	rounding.HalfEven,
}

var testModeNames = []string{
	"down",
	"up",
	"toward-zero",
	"away-from-zero",
	"half-up",
	"half-down",
	"half-away-from-zero",
	"half-toward-zero",
	"half-even",
}

func checkError(t *testing.T, err, want error) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, want)
	for _, other := range []error{
		rounding.ErrUnset,
		rounding.ErrUnknownMode,
		rounding.ErrInvalidArgument,
		rounding.ErrOutOfRange,
	} {
		if other != want {
			assert.NotErrorIs(t, err, other)
		}
	}
}

func expectedNeighbor(m rounding.Mode, comparison int, side rounding.Side, lowerEven bool) rounding.Neighbor {
	switch m {
	case rounding.Down:
		return rounding.Lower
	case rounding.Up:
		return rounding.Upper
	case rounding.TowardZero:
		if side == rounding.Positive {
			return rounding.Lower
		}
		return rounding.Upper
	case rounding.AwayFromZero:
		if side == rounding.Positive {
			return rounding.Upper
		}
		return rounding.Lower
	}
	if comparison < 0 {
		return rounding.Lower
	}
	if comparison > 0 {
		return rounding.Upper
	}
	switch m {
	case rounding.HalfUp:
		return rounding.Upper
	case rounding.HalfDown:
		return rounding.Lower
	case rounding.HalfAwayFromZero:
		if side == rounding.Positive {
			return rounding.Upper
		}
		return rounding.Lower
	case rounding.HalfTowardZero:
		if side == rounding.Positive {
			return rounding.Lower
		}
		return rounding.Upper
	case rounding.HalfEven:
		if lowerEven {
			return rounding.Lower
		}
		return rounding.Upper
	default:
		return rounding.Neighbor(0)
	}
}

func pow10(n int) *big.Int {
	if n < 0 {
		n = -n
	}
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

func scaleRat(x *big.Rat, places int) *big.Rat {
	result := new(big.Rat).Set(x)
	if places >= 0 {
		return result.Mul(result, new(big.Rat).SetInt(pow10(places)))
	}
	return result.Quo(result, new(big.Rat).SetInt(pow10(-places)))
}

func unscaleInt(k *big.Int, places int) *big.Rat {
	result := new(big.Rat).SetInt(k)
	if places >= 0 {
		return result.Quo(result, new(big.Rat).SetInt(pow10(places)))
	}
	return result.Mul(result, new(big.Rat).SetInt(pow10(-places)))
}

// roundRat implements the contract's exact arithmetic independently of rounding.Choose.
func roundRat(m rounding.Mode, x *big.Rat) *big.Int {
	if x.IsInt() {
		return new(big.Int).Set(x.Num())
	}
	lower := new(big.Int)
	remainder := new(big.Int)
	lower.QuoRem(x.Num(), x.Denom(), remainder)
	if x.Sign() < 0 {
		lower.Sub(lower, big.NewInt(1))
	}
	fraction := new(big.Rat).Sub(x, new(big.Rat).SetInt(lower))
	twiceFraction := new(big.Rat).Mul(fraction, big.NewRat(2, 1))
	comparison := twiceFraction.Cmp(big.NewRat(1, 1))
	side := rounding.Positive
	if lower.Sign() < 0 {
		side = rounding.Negative
	}
	if expectedNeighbor(m, comparison, side, lower.Bit(0) == 0) == rounding.Lower {
		return lower
	}
	return lower.Add(lower, big.NewInt(1))
}

func decimalFromInt(k *big.Int, places int) string {
	if k.Sign() == 0 {
		if places > 0 {
			return "0." + strings.Repeat("0", places)
		}
		return "0"
	}
	if places <= 0 {
		return new(big.Int).Mul(k, pow10(-places)).String()
	}
	absolute := new(big.Int).Abs(k)
	integer, fraction := new(big.Int), new(big.Int)
	integer.QuoRem(absolute, pow10(places), fraction)
	fractionText := fraction.String()
	result := integer.String() + "." + strings.Repeat("0", places-len(fractionText)) + fractionText
	if k.Sign() < 0 {
		return "-" + result
	}
	return result
}

func expectedDecimal(m rounding.Mode, x *big.Rat, places int) string {
	return decimalFromInt(roundRat(m, scaleRat(x, places)), places)
}

const twoTo53 = 9007199254740992.0

func floatValue(v float64) *big.Rat {
	if math.Abs(v) < twoTo53 {
		text := strconv.FormatFloat(v, 'e', -1, 64)
		result, ok := new(big.Rat).SetString(text)
		if !ok {
			panic("test model: invalid shortest float representation")
		}
		return result
	}
	return new(big.Rat).SetFloat64(v)
}

func nearestFloat(x *big.Rat) (float64, bool) {
	result, _ := x.Float64()
	if math.IsInf(result, 0) {
		return result, true
	}
	if result == 0 {
		return 0, false
	}
	return result, false
}

func expectedFloat(m rounding.Mode, x float64) float64 {
	result, overflow := nearestFloat(new(big.Rat).SetInt(roundRat(m, floatValue(x))))
	if overflow {
		panic("test model: Float cannot overflow")
	}
	return result
}

func expectedFloatInt(m rounding.Mode, x float64) *big.Int {
	return roundRat(m, floatValue(x))
}

func expectedFloatMultiple(m rounding.Mode, x, step float64) (float64, bool) {
	k := roundRat(m, new(big.Rat).Quo(floatValue(x), floatValue(step)))
	return nearestFloat(new(big.Rat).Mul(new(big.Rat).SetInt(k), floatValue(step)))
}

func expectedFloatSteps(m rounding.Mode, x, step float64) *big.Int {
	return roundRat(m, new(big.Rat).Quo(floatValue(x), floatValue(step)))
}

func expectedFloatPlaces(m rounding.Mode, x float64, places int) (float64, bool) {
	k := roundRat(m, scaleRat(floatValue(x), places))
	return nearestFloat(unscaleInt(k, places))
}

func expectedFloatDecimal(m rounding.Mode, x float64, places int) string {
	return expectedDecimal(m, floatValue(x), places)
}

func expectedFloatFraction(m rounding.Mode, x float64, den int) float64 {
	k := roundRat(m, new(big.Rat).Mul(floatValue(x), new(big.Rat).SetInt64(int64(den))))
	result, overflow := nearestFloat(new(big.Rat).Quo(new(big.Rat).SetInt(k), new(big.Rat).SetInt64(int64(den))))
	if overflow {
		panic("test model: FloatFraction cannot overflow")
	}
	return result
}

func expectedFloatNumerator(m rounding.Mode, x float64, den int) *big.Int {
	return roundRat(m, new(big.Rat).Mul(floatValue(x), new(big.Rat).SetInt64(int64(den))))
}

func intInRange(x *big.Int) bool {
	return x.IsInt64()
}

func bigInt(x int) *big.Int {
	return new(big.Int).SetInt64(int64(x))
}

func bigRat(n, d int) *big.Rat {
	return new(big.Rat).SetFrac(bigInt(n), bigInt(d))
}
