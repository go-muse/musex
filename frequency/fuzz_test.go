package frequency_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzParse(f *testing.F) {
	for _, tt := range canonicalTextCases() {
		f.Add(tt.text)
	}
	for _, tt := range parseErrorCases() {
		f.Add(tt.input)
	}
	for _, text := range []string{
		"cents(+5)", "cents(5 cents)", "cents(5¢)", "cents(6:4)",
		"cents(-0)", "cents(0.10)", "cents(9000000)", "hertz(00.50)",
		"hertz(1e5)", "hertz(1e6)", "hertz(3e-324)", "hertz(-0e5)",
		"relative(1e0)", "unset",
	} {
		f.Add(text)
	}
	f.Fuzz(func(t *testing.T, text string) {
		criterion, err := frequency.Parse(text)
		if err != nil {
			assert.False(t, criterion.IsValid())
			matched := requireAnyFrequencyError(t, err)
			require.NotNil(t, matched)
			return
		}
		canonical := criterion.String()
		roundTrip, err := frequency.Parse(canonical)
		require.NoError(t, err)
		requireCriterionEqual(t, criterion, roundTrip)
		assert.Equal(t, canonical, roundTrip.String())
	})
}

func FuzzNear(f *testing.F) {
	for _, pair := range frequencyPairs() {
		f.Add(pair.f1, pair.f2, 0.0)
		f.Add(pair.f1, pair.f2, 0.5)
		f.Add(pair.f1, pair.f2, 1.0)
	}
	f.Add(797.3, 731.9214, 0.082)
	f.Add(float64(9007199254740994), 1.5, float64(1<<53))
	f.Fuzz(func(t *testing.T, f1, f2, tolerance float64) {
		if !cents.IsValidFrequency(f1) || !cents.IsValidFrequency(f2) || math.IsNaN(tolerance) || math.IsInf(tolerance, 0) {
			t.Skip("outside the valid frequency and tolerance domain")
		}
		tolerance = math.Abs(tolerance)
		centsTolerance, err := cents.FromFloat(tolerance)
		if err != nil {
			t.Skip("cents tolerance outside its range")
		}
		criteria := []frequency.Criterion{
			frequency.Strict(),
			frequency.MustWithinHertz(tolerance),
			frequency.MustWithinRelative(tolerance),
			frequency.MustWithinCents(centsTolerance),
		}
		for _, criterion := range criteria {
			near, err := criterion.Near(f1, f2)
			require.NoError(t, err)
			assert.Equal(t, modelNear(criterion, f1, f2), near)
			reverse, err := criterion.Near(f2, f1)
			require.NoError(t, err)
			assert.Equal(t, near, reverse)
			reflexive, err := criterion.Near(f1, f1)
			require.NoError(t, err)
			assert.True(t, reflexive)

			comparison, err := criterion.Compare(f1, f2)
			require.NoError(t, err)
			assert.Equal(t, modelCompare(criterion, f1, f2), comparison)
			reverseComparison, err := criterion.Compare(f2, f1)
			require.NoError(t, err)
			assert.Equal(t, -comparison, reverseComparison)
		}
	})
}

func fuzzCriterion(data []byte) frequency.Criterion {
	if len(data) == 0 {
		return frequency.Strict()
	}
	simple := func(selector byte) frequency.Criterion {
		switch selector % 6 {
		case 0:
			return frequency.Strict()
		case 1:
			return centsCriterion(int(selector))
		case 2:
			return frequency.MustWithinCents(cents.MustFromRatio(81, 80))
		case 3:
			tolerances := []float64{0, 0.1, 0.5, 1, math.SmallestNonzeroFloat64, math.MaxFloat64}
			return frequency.MustWithinHertz(tolerances[int(selector)%len(tolerances)])
		case 4:
			tolerances := []float64{0, 0.001, 0.01, 1, math.SmallestNonzeroFloat64, math.MaxFloat64}
			return frequency.MustWithinRelative(tolerances[int(selector)%len(tolerances)])
		default:
			return frequency.MustWithinCents(cents.MustParse("5^(1/4)"))
		}
	}
	criterion := simple(data[0])
	for _, selector := range data[1:min(len(data), 16)] {
		component := simple(selector)
		if selector%2 == 0 {
			criterion = frequency.MustAllOf(criterion, component)
		} else {
			criterion = frequency.MustAnyOf(criterion, component)
		}
	}
	return criterion
}

func FuzzText(f *testing.F) {
	for _, tt := range canonicalTextCases() {
		f.Add([]byte(tt.text))
	}
	f.Add([]byte{})
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 255})
	f.Fuzz(func(t *testing.T, data []byte) {
		first := fuzzCriterion(data)
		text := first.String()
		parsed, err := frequency.Parse(text)
		require.NoError(t, err)
		requireCriterionEqual(t, first, parsed)

		marshaled, err := first.MarshalText()
		require.NoError(t, err)
		var unmarshaled frequency.Criterion
		require.NoError(t, unmarshaled.UnmarshalText(marshaled))
		requireCriterionEqual(t, first, unmarshaled)

		second := fuzzCriterion(data[len(data)/2:])
		assert.Equal(t, first.String() == second.String(), first.Equal(second))
	})
}
