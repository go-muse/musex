package frequency_test

import (
	"cmp"
	"errors"
	"math"
	"math/big"
	"strconv"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func centsCriterion(n int) frequency.Criterion {
	return frequency.MustWithinCents(cents.MustFromInt(n))
}

func definedCriteria() []struct {
	name string
	c    frequency.Criterion
} {
	return []struct {
		name string
		c    frequency.Criterion
	}{
		{"strict", frequency.Strict()},
		{"zero_cents", frequency.MustWithinCents(cents.Zero())},
		{"five_cents", centsCriterion(5)},
		{"ratio_cents", frequency.MustWithinCents(cents.MustFromRatio(81, 80))},
		{"power_cents", frequency.MustWithinCents(cents.MustParse("5^(1/4)"))},
		{"zero_hertz", frequency.MustWithinHertz(0)},
		{"one_hertz", frequency.MustWithinHertz(1)},
		{"negative_zero_hertz", frequency.MustWithinHertz(math.Copysign(0, -1))},
		{"zero_relative", frequency.MustWithinRelative(0)},
		{"small_relative", frequency.MustWithinRelative(0.01)},
		{"all", frequency.MustAllOf(centsCriterion(5), frequency.MustWithinHertz(2))},
		{"any_nested", frequency.MustAnyOf(centsCriterion(1), frequency.MustAllOf(frequency.Strict(), frequency.MustWithinHertz(0.5)))},
	}
}

func rational(x float64) *big.Rat {
	return new(big.Rat).SetFloat64(x)
}

func exactDifference(f1, f2 float64) *big.Rat {
	return new(big.Rat).Sub(rational(f1), rational(f2))
}

func modelNear(c frequency.Criterion, f1, f2 float64) bool {
	switch c.Kind() {
	case frequency.KindStrict:
		return f1 == f2
	case frequency.KindCents:
		tolerance, ok := c.CentsTolerance()
		if !ok {
			panic("cents criterion without tolerance")
		}
		return cents.MustBetween(f1, f2).Abs().Compare(tolerance) <= 0
	case frequency.KindHertz:
		tolerance, ok := c.Tolerance()
		if !ok {
			panic("hertz criterion without tolerance")
		}
		difference := exactDifference(f1, f2)
		difference.Abs(difference)
		return difference.Cmp(rational(tolerance)) <= 0
	case frequency.KindRelative:
		tolerance, ok := c.Tolerance()
		if !ok {
			panic("relative criterion without tolerance")
		}
		maximum := rational(f1)
		if f2 > f1 {
			maximum = rational(f2)
		}
		limit := new(big.Rat).Mul(rational(tolerance), maximum)
		difference := exactDifference(f1, f2)
		difference.Abs(difference)
		return difference.Cmp(limit) <= 0
	case frequency.KindAllOf:
		for _, component := range c.Components() {
			if !modelNear(component, f1, f2) {
				return false
			}
		}
		return true
	case frequency.KindAnyOf:
		for _, component := range c.Components() {
			if modelNear(component, f1, f2) {
				return true
			}
		}
		return false
	default:
		panic("modelNear called with an unset criterion")
	}
}

func modelCompare(c frequency.Criterion, f1, f2 float64) int {
	if modelNear(c, f1, f2) {
		return 0
	}
	return cmp.Compare(f1, f2)
}

func requireCriterionEqual(t *testing.T, expected, actual frequency.Criterion) {
	t.Helper()
	require.True(t, expected.Equal(actual), "criteria differ: %q != %q", expected.String(), actual.String())
}

func requireErrorChain(t *testing.T, err error, expectedFrequency, expectedCents error) {
	t.Helper()
	require.Error(t, err)

	for _, target := range []error{
		frequency.ErrInvalidFrequency,
		frequency.ErrInvalidTolerance,
		frequency.ErrUnset,
		frequency.ErrNoComponents,
		frequency.ErrInvalidText,
		frequency.ErrInvalidParameter,
		frequency.ErrOutOfRange,
	} {
		assert.Equal(t, target == expectedFrequency, errors.Is(err, target),
			"unexpected frequency error chain for %v", target)
	}
	for _, target := range []error{
		cents.ErrInvalidArgument,
		cents.ErrInvalidText,
		cents.ErrTooComplex,
		cents.ErrPrimeLimit,
		cents.ErrOutOfRange,
	} {
		assert.Equal(t, target == expectedCents, errors.Is(err, target),
			"unexpected cents error chain for %v", target)
	}
}

func requireAnyFrequencyError(t *testing.T, err error) error {
	t.Helper()
	require.Error(t, err)
	frequencyErrors := []error{
		frequency.ErrInvalidFrequency,
		frequency.ErrInvalidTolerance,
		frequency.ErrUnset,
		frequency.ErrNoComponents,
		frequency.ErrInvalidText,
		frequency.ErrInvalidParameter,
		frequency.ErrOutOfRange,
	}
	var matched error
	for _, target := range frequencyErrors {
		if errors.Is(err, target) {
			require.Nil(t, matched, "error matches more than one frequency sentinel")
			matched = target
		}
	}
	require.NotNil(t, matched, "error has no frequency sentinel")

	centsErrors := []error{
		cents.ErrInvalidArgument,
		cents.ErrInvalidText,
		cents.ErrTooComplex,
		cents.ErrPrimeLimit,
		cents.ErrOutOfRange,
	}
	var centsMatch error
	for _, target := range centsErrors {
		if errors.Is(err, target) {
			require.Nil(t, centsMatch, "error matches more than one cents sentinel")
			centsMatch = target
		}
	}
	switch matched {
	case frequency.ErrInvalidFrequency:
		require.Equal(t, cents.ErrInvalidArgument, centsMatch)
	case frequency.ErrInvalidTolerance:
		if centsMatch != nil {
			require.Contains(t, []error{cents.ErrTooComplex, cents.ErrPrimeLimit, cents.ErrOutOfRange}, centsMatch)
		}
	case frequency.ErrInvalidText:
		if centsMatch != nil {
			require.Equal(t, cents.ErrInvalidText, centsMatch)
		}
	case frequency.ErrInvalidParameter:
		if centsMatch != nil {
			require.Equal(t, cents.ErrInvalidArgument, centsMatch)
		}
	case frequency.ErrOutOfRange:
		if centsMatch != nil {
			require.Equal(t, cents.ErrOutOfRange, centsMatch)
		}
	default:
		require.Nil(t, centsMatch)
	}
	return matched
}

func binarySum() float64 {
	a, b := 0.1, 0.2
	return a + b
}

func frequencyPairs() []struct {
	name   string
	f1, f2 float64
} {
	return []struct {
		name   string
		f1, f2 float64
	}{
		{"equal", 440, 440},
		{"adjacent", 440, math.Nextafter(440, math.Inf(1))},
		{"near", 440, 441},
		{"reverse", 441, 440},
		{"binary_sum", binarySum(), 0.3},
		{"subnormal_equal", math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64},
		{"subnormal_double", math.SmallestNonzeroFloat64, 2 * math.SmallestNonzeroFloat64},
		{"extremes", math.SmallestNonzeroFloat64, math.MaxFloat64},
		{"extremes_reverse", math.MaxFloat64, math.SmallestNonzeroFloat64},
		{"exact_hertz_counterexample", 9007199254740994, 1.5},
		{"exact_relative_counterexample", 797.3, 731.9214},
		{"relative_boundary", 4, 3},
		{"relative_outside", 4, 2.9999999999999996},
		{"zero_cents_nontransitive_a", 440, 440.0000001},
		{"zero_cents_nontransitive_b", 440.0000001, 440.0000002},
		{"zero_cents_nontransitive_c", 440, 440.0000002},
		{"octave", 440, 880},
		{"wide", 1, 1000},
	}
}

func negativeZero() float64 {
	return math.Copysign(0, -1)
}

func invalidFrequencyCases() []struct {
	name string
	f    float64
} {
	return []struct {
		name string
		f    float64
	}{
		{"zero", 0},
		{"negative_zero", negativeZero()},
		{"negative", -1},
		{"nan", math.NaN()},
		{"positive_infinity", math.Inf(1)},
		{"negative_infinity", math.Inf(-1)},
	}
}

func decimalRat(f float64) *big.Rat {
	text := strconv.FormatFloat(f, 'g', -1, 64)
	value, ok := new(big.Rat).SetString(text)
	if !ok {
		panic("cannot parse shortest decimal representation: " + text)
	}
	return value
}

func generatedFrequency(x *big.Rat) (float64, bool) {
	rounded := new(big.Float).SetPrec(53).SetMode(big.ToNearestEven).SetRat(x)
	minimum := new(big.Float).SetPrec(53).SetFloat64(cents.MinNormalFrequency)
	maximum := new(big.Float).SetPrec(53).SetFloat64(cents.MaxFrequency)
	if rounded.Cmp(minimum) < 0 || rounded.Cmp(maximum) > 0 {
		return 0, false
	}
	result, _ := rounded.Float64()
	return result, true
}

func modelPartial(f float64, n int) (float64, bool) {
	x := new(big.Rat).Mul(rational(f), new(big.Rat).SetInt64(int64(n)))
	return generatedFrequency(x)
}

func modelSubharmonic(f float64, n int) (float64, bool) {
	x := new(big.Rat).Quo(rational(f), new(big.Rat).SetInt64(int64(n)))
	return generatedFrequency(x)
}

func modelBeatFrequency(f1, f2 float64) float64 {
	difference := exactDifference(f1, f2)
	difference.Abs(difference)
	result, _ := difference.Float64()
	return result
}

func modelPeriod(f float64) (float64, bool) {
	x := new(big.Rat).Inv(rational(f))
	result, _ := x.Float64()
	return result, !math.IsInf(result, 1)
}

func powerOfTwo(k int) *big.Rat {
	if k >= 0 {
		return new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(k)))
	}
	return new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), uint(-k)))
}

func modelNormalizeOctave(f, base float64) (float64, int, bool, *big.Rat) {
	fractionF, exponentF := math.Frexp(f)
	fractionBase, exponentBase := math.Frexp(base)
	q := exponentF - exponentBase
	if fractionF < fractionBase {
		q--
	}
	k := -q
	candidate := new(big.Rat).Mul(rational(f), powerOfTwo(k))
	representative, ok := generatedFrequency(candidate)
	return representative, k, ok, candidate
}
