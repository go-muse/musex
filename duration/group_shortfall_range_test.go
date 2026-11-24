package duration_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cyclotomicChain returns tuplets n:1, one per divisor d > 1 of k, with n the
// value of the cyclotomic polynomial Phi_d at 2. The product of these n is
// 2^k - 1, so the chain has the coefficient 1/(2^k - 1), while every count
// stays far below math.MaxInt for the k used here (Phi_d(2) < 2^(phi(d)+1)).
func cyclotomicChain(t *testing.T, k int) []tuplet.Tuplet {
	t.Helper()

	values := map[int]*big.Int{}

	var chain []tuplet.Tuplet

	for d := 1; d <= k; d++ {
		if k%d != 0 {
			continue
		}

		v := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(d)), big.NewInt(1))

		for e, phi := range values {
			if d%e == 0 {
				v.Quo(v, phi)
			}
		}

		values[d] = v

		if d == 1 {
			continue
		}

		require.True(t, v.IsInt64(), "Phi_%d(2) fits in int", d)
		chain = append(chain, tp(int(v.Int64()), 1))
	}

	product := big.NewInt(1)
	for _, v := range values {
		product.Mul(product, v)
	}

	want := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(k)), big.NewInt(1))
	require.Zero(t, want.Cmp(product), "the counts multiply to 2^%d - 1", k)

	return chain
}

// A group whose span and contents are in range but whose shortfall is not
// fails with ErrOutOfRange together with fraction.ErrOutOfRange.
//
// The unit is a whole note under a chain with the coefficient
// 1/(2^180 - 1), so the group 1:1 spans 1/(2^180 - 1). Its single content is
// a 1024th with 179 dots, written (2^180 - 1)/2^189, under the same chain
// extended by 1:1: it sounds 1/2^189, less than the span, so NewGroup accepts
// it. The shortfall 1/(2^180 - 1) - 1/2^189 has the reduced denominator
// (2^180 - 1)·2^189, above the largest component 2^255 - 1 of package
// common/fraction.
func TestGroupShortfallOutOfRange(t *testing.T) {
	t.Parallel()

	const k = 180

	chain := cyclotomicChain(t, k)
	unit := wr(duration.Whole, 0, chain...)
	neutral := tp(1, 1)
	content := wcon(duration.ThousandTwentyFourth, k-1, append(append([]tuplet.Tuplet{}, chain...), neutral)...)

	g, err := duration.NewGroup(neutral, unit, content)
	require.NoError(t, err)

	m := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), k), big.NewInt(1))
	span := new(big.Rat).SetFrac(big.NewInt(1), m)
	sounding := new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), k+9))

	require.Zero(t, span.Cmp(ratOf(noErr(g.Span()))), "the span is 1/(2^180 - 1)")
	require.Zero(t, sounding.Cmp(ratOfDuration(noErr(content.MustWritten().Duration()))),
		"the content sounds 1/2^189")

	shortfall := new(big.Rat).Sub(span, sounding)
	require.Greater(t, shortfall.Denom().Cmp(fraction.MaxComponentBig()), 0,
		"the exact shortfall has a denominator outside the range of common/fraction")

	_, err = g.Shortfall()
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.NotErrorIs(t, err, duration.ErrUnset)
}
