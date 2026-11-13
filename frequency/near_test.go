package frequency_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNear_ControlExamples(t *testing.T) {
	adjacent := math.Nextafter(440, math.Inf(1))
	a, b := 0.1, 0.2
	tests := []struct {
		name    string
		c       frequency.Criterion
		f1, f2  float64
		near    bool
		compare int
	}{
		{"K-1 strict equal", frequency.Strict(), 440, 440, true, 0},
		{"K-2 strict adjacent", frequency.Strict(), 440, adjacent, false, -1},
		{"K-3 strict binary sum", frequency.Strict(), a + b, 0.3, false, 1},
		{"K-4 five cents", centsCriterion(5), 440, 441, true, 0},
		{"K-5 cents boundary", frequency.MustWithinCents(cents.MustParse("3.930158")), 440, 441, true, 0},
		{"K-6a three cents below", centsCriterion(3), 440, 441, false, -1},
		{"K-6b three cents above", centsCriterion(3), 441, 440, false, 1},
		{"K-6v three cents reverse pair", centsCriterion(3), 440, 439, false, 1},
		{"K-7a zero cents adjacent", frequency.MustWithinCents(cents.Zero()), 440, adjacent, true, 0},
		{"K-7b zero cents first", frequency.MustWithinCents(cents.Zero()), 440, 440.0000001, true, 0},
		{"K-7b zero cents second", frequency.MustWithinCents(cents.Zero()), 440.0000001, 440.0000002, true, 0},
		{"K-7b zero cents nontransitive", frequency.MustWithinCents(cents.Zero()), 440, 440.0000002, false, -1},
		{"K-8 hertz first", frequency.MustWithinHertz(1), 440, 441, true, 0},
		{"K-8 hertz second", frequency.MustWithinHertz(1), 441, 442, true, 0},
		{"K-8 hertz nontransitive", frequency.MustWithinHertz(1), 440, 442, false, -1},
		{"K-9 octave boundary", frequency.MustWithinCents(cents.Octave()), 440, 880, true, 0},
		{"K-9 below octave boundary", frequency.MustWithinCents(cents.MustParse("1199.999999")), 440, 880, false, -1},
		{"K-10 half hertz", frequency.MustWithinHertz(0.5), 440, 441, false, -1},
		{"K-11 zero hertz adjacent", frequency.MustWithinHertz(0), 440, adjacent, false, -1},
		{"K-12 exact hertz arithmetic", frequency.MustWithinHertz(1 << 53), 9007199254740994, 1.5, false, 1},
		{"K-13 relative boundary", frequency.MustWithinRelative(0.25), 4, 3, true, 0},
		{"K-14 relative outside", frequency.MustWithinRelative(0.25), 4, 2.9999999999999996, false, 1},
		{"K-15 exact relative arithmetic", frequency.MustWithinRelative(0.082), 797.3, 731.9214, false, 1},
		{"K-16 relative one", frequency.MustWithinRelative(1), 1, 1000, true, 0},
		{"K-17a strict subnormal", frequency.Strict(), math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64, true, 0},
		{"K-17b cents subnormals", frequency.MustWithinCents(cents.Octave()), math.SmallestNonzeroFloat64, 2 * math.SmallestNonzeroFloat64, true, 0},
		{"K-17v hertz subnormals", frequency.MustWithinHertz(math.SmallestNonzeroFloat64), math.SmallestNonzeroFloat64, 2 * math.SmallestNonzeroFloat64, true, 0},
		{"K-18a maximum hertz", frequency.MustWithinHertz(math.MaxFloat64), math.SmallestNonzeroFloat64, math.MaxFloat64, true, 0},
		{"K-18b extreme cents boundary", frequency.MustWithinCents(cents.MustParse("2517579.329578")), math.SmallestNonzeroFloat64, math.MaxFloat64, true, 0},
		{"K-18v below extreme cents", frequency.MustWithinCents(cents.MustParse("2517579.329577")), math.SmallestNonzeroFloat64, math.MaxFloat64, false, -1},
		{"K-19 any", frequency.MustAnyOf(centsCriterion(1), frequency.MustWithinHertz(0.5)), 440, 441, false, -1},
		{"K-20 all", frequency.MustAllOf(centsCriterion(5), frequency.MustWithinHertz(2)), 440, 441, true, 0},
		{"K-21 strict or zero cents", frequency.MustAnyOf(frequency.Strict(), frequency.MustWithinCents(cents.Zero())), 440, adjacent, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			near, err := tt.c.Near(tt.f1, tt.f2)
			require.NoError(t, err)
			assert.Equal(t, tt.near, near)
			assert.Equal(t, tt.near, tt.c.MustNear(tt.f1, tt.f2))
			assert.Equal(t, modelNear(tt.c, tt.f1, tt.f2), near)

			comparison, err := tt.c.Compare(tt.f1, tt.f2)
			require.NoError(t, err)
			assert.Equal(t, tt.compare, comparison)
			assert.Equal(t, tt.compare, tt.c.MustCompare(tt.f1, tt.f2))
			assert.Equal(t, modelCompare(tt.c, tt.f1, tt.f2), comparison)
		})
	}
}

func TestNear_InvalidFrequency(t *testing.T) {
	criteria := []struct {
		name string
		c    frequency.Criterion
	}{
		{"strict", frequency.Strict()},
		{"cents", centsCriterion(5)},
		{"hertz", frequency.MustWithinHertz(1)},
		{"relative", frequency.MustWithinRelative(0.01)},
		{"any", frequency.MustAnyOf(frequency.Strict())},
	}
	invalid := []struct {
		name string
		f    float64
	}{
		{"zero", 0},
		{"negative_zero", math.Copysign(0, -1)},
		{"negative", -1},
		{"nan", math.NaN()},
		{"positive_infinity", math.Inf(1)},
		{"negative_infinity", math.Inf(-1)},
	}
	for _, criterion := range criteria {
		for _, invalidFrequency := range invalid {
			t.Run(criterion.name+"/first/"+invalidFrequency.name, func(t *testing.T) {
				t.Parallel()
				near, err := criterion.c.Near(invalidFrequency.f, 440)
				assert.False(t, near)
				requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
				comparison, err := criterion.c.Compare(invalidFrequency.f, 440)
				assert.Zero(t, comparison)
				requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
			})
			t.Run(criterion.name+"/second/"+invalidFrequency.name, func(t *testing.T) {
				t.Parallel()
				near, err := criterion.c.Near(440, invalidFrequency.f)
				assert.False(t, near)
				requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
				comparison, err := criterion.c.Compare(440, invalidFrequency.f)
				assert.Zero(t, comparison)
				requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
			})
		}
	}
}

func TestNear_UnsetAndErrorPrecedence(t *testing.T) {
	var unset frequency.Criterion
	tests := []struct {
		name    string
		c       frequency.Criterion
		f1, f2  float64
		wantErr error
	}{
		{"K-25 valid frequencies", unset, 440, 440, frequency.ErrUnset},
		{"K-25 unset precedes invalid frequency", unset, 440, math.NaN(), frequency.ErrUnset},
		{"K-25 unset precedes both invalid", unset, math.NaN(), 0, frequency.ErrUnset},
		{"K-26 first invalid", centsCriterion(5), math.NaN(), 0, frequency.ErrInvalidFrequency},
		{"first invalid before second", frequency.Strict(), 0, math.NaN(), frequency.ErrInvalidFrequency},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			near, err := tt.c.Near(tt.f1, tt.f2)
			assert.False(t, near)
			expectedCents := error(nil)
			if tt.wantErr == frequency.ErrInvalidFrequency {
				expectedCents = cents.ErrInvalidArgument
			}
			requireErrorChain(t, err, tt.wantErr, expectedCents)
			comparison, err := tt.c.Compare(tt.f1, tt.f2)
			assert.Zero(t, comparison)
			requireErrorChain(t, err, tt.wantErr, expectedCents)
		})
	}
}

func TestNear_Properties(t *testing.T) {
	criteria := definedCriteria()
	pairs := frequencyPairs()
	for _, criterion := range criteria {
		for _, pair := range pairs {
			t.Run(criterion.name+"/"+pair.name, func(t *testing.T) {
				t.Parallel()
				near, err := criterion.c.Near(pair.f1, pair.f2)
				require.NoError(t, err)
				reverse, err := criterion.c.Near(pair.f2, pair.f1)
				require.NoError(t, err)
				assert.Equal(t, modelNear(criterion.c, pair.f1, pair.f2), near)
				assert.Equal(t, near, reverse, "Near must be symmetric")

				comparison, err := criterion.c.Compare(pair.f1, pair.f2)
				require.NoError(t, err)
				reverseComparison, err := criterion.c.Compare(pair.f2, pair.f1)
				require.NoError(t, err)
				assert.Equal(t, modelCompare(criterion.c, pair.f1, pair.f2), comparison)
				assert.Equal(t, -comparison, reverseComparison, "Compare must be antisymmetric")
				assert.Equal(t, near, comparison == 0)

				reflexive, err := criterion.c.Near(pair.f1, pair.f1)
				require.NoError(t, err)
				assert.True(t, reflexive)
			})
		}
	}
}

func TestNear_MonotonicToleranceAndZeroEquivalence(t *testing.T) {
	pairs := frequencyPairs()
	t.Run("cents", func(t *testing.T) {
		small := centsCriterion(1)
		large := centsCriterion(5)
		for _, pair := range pairs {
			smallNear, err := small.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			largeNear, err := large.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.False(t, smallNear && !largeNear, pair.name)
		}
	})
	t.Run("hertz", func(t *testing.T) {
		small := frequency.MustWithinHertz(0.5)
		large := frequency.MustWithinHertz(1)
		for _, pair := range pairs {
			smallNear, err := small.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			largeNear, err := large.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.False(t, smallNear && !largeNear, pair.name)
		}
	})
	t.Run("relative", func(t *testing.T) {
		small := frequency.MustWithinRelative(0.001)
		large := frequency.MustWithinRelative(0.01)
		for _, pair := range pairs {
			smallNear, err := small.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			largeNear, err := large.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.False(t, smallNear && !largeNear, pair.name)
		}
	})
	t.Run("zero equals strict", func(t *testing.T) {
		strict := frequency.Strict()
		hertz := frequency.MustWithinHertz(0)
		relative := frequency.MustWithinRelative(0)
		for _, pair := range pairs {
			want, err := strict.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			gotHertz, err := hertz.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			gotRelative, err := relative.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.Equal(t, want, gotHertz, pair.name)
			assert.Equal(t, want, gotRelative, pair.name)
		}
	})
}

func TestNear_CompositeOrderIndependence(t *testing.T) {
	components := []frequency.Criterion{
		centsCriterion(5),
		frequency.MustWithinHertz(2),
		frequency.MustWithinRelative(0.01),
		frequency.MustAnyOf(frequency.Strict(), frequency.MustWithinCents(cents.Zero())),
	}
	permutations := [][]int{
		{0, 1, 2, 3}, {0, 3, 2, 1}, {1, 0, 3, 2}, {1, 2, 3, 0},
		{2, 1, 0, 3}, {2, 3, 1, 0}, {3, 0, 1, 2}, {3, 2, 0, 1},
		{0, 2, 1, 3}, {1, 3, 0, 2}, {2, 0, 3, 1}, {3, 1, 2, 0},
		{0, 3, 1, 2}, {1, 0, 2, 3}, {2, 1, 3, 0}, {3, 2, 1, 0},
		{0, 1, 3, 2}, {1, 2, 0, 3}, {2, 3, 0, 1}, {3, 0, 2, 1},
		{0, 2, 3, 1}, {1, 3, 2, 0}, {2, 0, 1, 3}, {3, 1, 0, 2},
	}
	for _, permutation := range permutations {
		ordered := make([]frequency.Criterion, len(permutation))
		for i, index := range permutation {
			ordered[i] = components[index]
		}
		all := frequency.MustAllOf(ordered...)
		any := frequency.MustAnyOf(ordered...)
		for _, pair := range frequencyPairs() {
			gotAll, err := all.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.Equal(t, modelNear(all, pair.f1, pair.f2), gotAll)
			gotAny, err := any.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.Equal(t, modelNear(any, pair.f1, pair.f2), gotAny)

			baseAll := frequency.MustAllOf(components...)
			baseAny := frequency.MustAnyOf(components...)
			wantAll, err := baseAll.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			wantAny, err := baseAny.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.Equal(t, wantAll, gotAll)
			assert.Equal(t, wantAny, gotAny)

			compareAll, err := all.Compare(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.Equal(t, modelCompare(all, pair.f1, pair.f2), compareAll)
			assert.Equal(t, modelCompare(all, pair.f1, pair.f2), all.MustCompare(pair.f1, pair.f2))
			compareAny, err := any.Compare(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.Equal(t, modelCompare(any, pair.f1, pair.f2), compareAny)
			assert.Equal(t, modelCompare(any, pair.f1, pair.f2), any.MustCompare(pair.f1, pair.f2))

			wantCompareAll, err := baseAll.Compare(pair.f1, pair.f2)
			require.NoError(t, err)
			wantCompareAny, err := baseAny.Compare(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.Equal(t, wantCompareAll, compareAll)
			assert.Equal(t, wantCompareAll, all.MustCompare(pair.f1, pair.f2))
			assert.Equal(t, wantCompareAny, compareAny)
			assert.Equal(t, wantCompareAny, any.MustCompare(pair.f1, pair.f2))
		}
	}
}

func TestNear_RelativeAtLeastOne(t *testing.T) {
	for _, tolerance := range []float64{1, 1.5, math.MaxFloat64} {
		criterion := frequency.MustWithinRelative(tolerance)
		for _, pair := range frequencyPairs() {
			near, err := criterion.Near(pair.f1, pair.f2)
			require.NoError(t, err)
			assert.True(t, near)
			assert.Zero(t, criterion.MustCompare(pair.f1, pair.f2))
		}
	}
}
