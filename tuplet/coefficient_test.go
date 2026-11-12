package tuplet_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCoefficient covers KP-1, K-4 and architecture 4.6: the coefficient is
// normal/actual with the source pair (normal, actual), the reciprocal is
// (actual, normal), and the reduced coefficient is the reduced pair.
func TestCoefficient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		actual, normal         int
		reducedNum, reducedDen int
	}{
		{actual: 3, normal: 2, reducedNum: 2, reducedDen: 3},
		{actual: 2, normal: 3, reducedNum: 3, reducedDen: 2},
		{actual: 6, normal: 4, reducedNum: 2, reducedDen: 3},
		{actual: 5, normal: 4, reducedNum: 4, reducedDen: 5},
		{actual: 5, normal: 6, reducedNum: 6, reducedDen: 5},
		{actual: 9, normal: 6, reducedNum: 2, reducedDen: 3},
		{actual: 5, normal: 5, reducedNum: 1, reducedDen: 1},
		{actual: 1, normal: 1, reducedNum: 1, reducedDen: 1},
		{actual: math.MaxInt, normal: math.MaxInt, reducedNum: 1, reducedDen: 1},
		{actual: math.MaxInt, normal: math.MaxInt - 1, reducedNum: math.MaxInt - 1, reducedDen: math.MaxInt},
		{actual: 1 << 62, normal: 1 << 61, reducedNum: 1, reducedDen: 2},
		{actual: 1, normal: math.MaxInt, reducedNum: math.MaxInt, reducedDen: 1},
	}

	for _, tt := range tests {
		t.Run(itoa(tt.actual)+":"+itoa(tt.normal), func(t *testing.T) {
			t.Parallel()

			tp := tuplet.MustNew(tt.actual, tt.normal)

			c, err := tp.Coefficient()
			require.NoError(t, err)
			assertPair(t, c, tt.normal, tt.actual)
			assert.Zero(t, ratOf(t, c).Cmp(ratInts(tt.normal, tt.actual)))

			r, err := tp.ReciprocalCoefficient()
			require.NoError(t, err)
			assertPair(t, r, tt.actual, tt.normal)

			reduced, err := tp.ReducedCoefficient()
			require.NoError(t, err)
			assertPair(t, reduced, tt.reducedNum, tt.reducedDen)
			assert.Zero(t, ratOf(t, reduced).Cmp(ratOf(t, c)), "reduced value equals the coefficient")

			product := new(big.Rat).Mul(ratOf(t, c), ratOf(t, r))
			assert.Zero(t, product.Cmp(big.NewRat(1, 1)), "coefficient times reciprocal is one")

			assert.Equal(t, c, tp.MustCoefficient())
			assert.Equal(t, r, tp.MustReciprocalCoefficient())
			assert.Equal(t, reduced, tp.MustReducedCoefficient())
		})
	}
}

// TestCoefficient_KP1Values covers KP-1 by value: 2/3, 3/2, 2/3, 4/5, 6/5.
func TestCoefficient_KP1Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tp       tuplet.Tuplet
		num, den int64
	}{
		{tp: tuplet.MustNew(3, 2), num: 2, den: 3},
		{tp: tuplet.MustNew(2, 3), num: 3, den: 2},
		{tp: tuplet.MustNew(6, 4), num: 2, den: 3},
		{tp: tuplet.MustNew(5, 4), num: 4, den: 5},
		{tp: tuplet.MustNew(5, 6), num: 6, den: 5},
	}

	for _, tt := range tests {
		t.Run(tt.tp.String(), func(t *testing.T) {
			t.Parallel()

			c := tt.tp.MustCoefficient()
			assert.Zero(t, ratOf(t, c).Cmp(big.NewRat(tt.num, tt.den)))
			assert.True(t, c.Equal(frac(int(tt.num), int(tt.den))))
		})
	}
}

// TestCoefficient_Oracle checks the coefficients of sample tuplets against
// big.Rat built from the counts.
func TestCoefficient_Oracle(t *testing.T) {
	t.Parallel()

	for _, tp := range sampleTuplets() {
		actual, normal := tp.MustPair()

		c := tp.MustCoefficient()
		assertPair(t, c, normal, actual)

		r := tp.MustReciprocalCoefficient()
		assertPair(t, r, actual, normal)

		assertReducedValue(t, tp.MustReducedCoefficient(), ratInts(normal, actual))
	}
}
