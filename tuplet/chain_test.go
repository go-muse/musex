package tuplet_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChain_KP5 covers KP-5: the chain coefficients 8/15 and 4/9 and the
// flattened notation 9:4.
func TestChain_KP5(t *testing.T) {
	t.Parallel()

	c, err := tuplet.ChainCoefficient(tuplet.Quintuplet(), tuplet.Triplet())
	require.NoError(t, err)
	assertPair(t, c, 8, 15)

	c, err = tuplet.ChainCoefficient(tuplet.Triplet(), tuplet.Triplet())
	require.NoError(t, err)
	assertPair(t, c, 4, 9)

	flat, err := tuplet.Flatten(tuplet.Triplet(), tuplet.Triplet())
	require.NoError(t, err)
	assertTuplet(t, flat, 9, 4)
}

// TestChain_KP6 covers KP-6: an eighth in the chain (3:2, 3:2) is 1/18.
func TestChain_KP6(t *testing.T) {
	t.Parallel()

	got, err := tuplet.ApplyChain(frac(1, 8), tuplet.Triplet(), tuplet.Triplet())
	require.NoError(t, err)
	assertPair(t, got, 1, 18)
}

// TestNest_KP13 covers KP-13 and G-4: a duplet inside a triplet is 6:6 with
// the coefficient value 1.
func TestNest_KP13(t *testing.T) {
	t.Parallel()

	got, err := tuplet.Triplet().Nest(tuplet.Duplet())
	require.NoError(t, err)
	assertTuplet(t, got, 6, 6)
	assert.True(t, got.MustCoefficient().IsOne())
	assert.Equal(t, got, tuplet.Triplet().MustNest(tuplet.Duplet()))

	kind, err := got.Kind()
	require.NoError(t, err)
	assert.Equal(t, tuplet.Neutral, kind)
}

// TestChain_Empty covers G-3: the empty chain is 1:1 with coefficient 1/1,
// and applying it gives d.Reduce().
func TestChain_Empty(t *testing.T) {
	t.Parallel()

	flat, err := tuplet.Flatten()
	require.NoError(t, err)
	assertTuplet(t, flat, 1, 1)

	c, err := tuplet.ChainCoefficient()
	require.NoError(t, err)
	assertPair(t, c, 1, 1)

	for _, d := range []fraction.Fraction{frac(2, 4), frac(0, 5), frac(-2, 4), frac(1, -2), frac(3, 1)} {
		got, err := tuplet.ApplyChain(d)
		require.NoError(t, err)
		assert.Equal(t, d.Reduce(), got, "ApplyChain(%s)", d)
	}
}

// TestChain_Oracle covers IP-6, K-9 and architecture 4.10 on chains of up to
// three groups: the flattened pair, the unreduced chain coefficient pair,
// application, Nest, and agreement between Flatten and ChainCoefficient.
func TestChain_Oracle(t *testing.T) {
	t.Parallel()

	chains := [][]tuplet.Tuplet{
		{tuplet.Triplet()},
		{tuplet.Sextuplet()},
		{tuplet.Triplet(), tuplet.Duplet()},
		{tuplet.Quintuplet(), tuplet.Triplet()},
		{tuplet.Sextuplet(), tuplet.Triplet(), tuplet.Duplet()},
		{tuplet.MustNew(7, 11), tuplet.MustNew(1, 1), tuplet.MustNew(13, 2)},
		{tuplet.MustNew(1<<31, 3), tuplet.MustNew(1<<31, 5)},
		{tuplet.MustNew(math.MaxInt, 1), tuplet.MustNew(1, math.MaxInt)},
	}

	durations := []fraction.Fraction{frac(1, 8), frac(0, 3), frac(-3, 4), frac(6, 4)}

	for _, chain := range chains {
		t.Run(chainName(chain), func(t *testing.T) {
			t.Parallel()

			nProduct, mProduct := big.NewInt(1), big.NewInt(1)
			for _, tp := range chain {
				a, n := tp.MustPair()
				nProduct.Mul(nProduct, bigOf(a))
				mProduct.Mul(mProduct, bigOf(n))
			}

			c, err := tuplet.ChainCoefficient(chain...)
			require.NoError(t, err)
			assertBigPair(t, c, mProduct, nProduct)

			value := new(big.Rat).SetFrac(mProduct, nProduct)
			for _, d := range durations {
				got, err := tuplet.ApplyChain(d, chain...)
				if assert.NoError(t, err) {
					assertReducedValue(t, got, new(big.Rat).Mul(ratOf(t, d), value))
				}
			}

			fitsInt := nProduct.IsInt64() && mProduct.IsInt64()

			flat, err := tuplet.Flatten(chain...)
			if !fitsInt {
				assertErr(t, err, tuplet.ErrOutOfRange)

				return
			}

			require.NoError(t, err)
			assertTuplet(t, flat, int(nProduct.Int64()), int(mProduct.Int64()))
			assert.Equal(t, c, flat.MustCoefficient(), "Flatten coefficient matches ChainCoefficient pair")

			if len(chain) == 2 {
				nested, err := chain[0].Nest(chain[1])
				require.NoError(t, err)
				assert.Equal(t, flat, nested)
			}
		})
	}
}

// chainName names a chain for subtests.
func chainName(chain []tuplet.Tuplet) string {
	name := ""
	for i, tp := range chain {
		if i > 0 {
			name += ","
		}

		name += tp.String()
	}

	return name
}

// TestFlatten_Single covers architecture 4.10: Flatten(t) == t.
func TestFlatten_Single(t *testing.T) {
	t.Parallel()

	for _, tp := range sampleTuplets() {
		got, err := tuplet.Flatten(tp)
		if assert.NoError(t, err) {
			assert.Equal(t, tp, got)
		}
	}
}

// permutations returns all orderings of chain.
func permutations(chain []tuplet.Tuplet) [][]tuplet.Tuplet {
	if len(chain) <= 1 {
		return [][]tuplet.Tuplet{append([]tuplet.Tuplet(nil), chain...)}
	}

	var out [][]tuplet.Tuplet

	for i := range chain {
		rest := make([]tuplet.Tuplet, 0, len(chain)-1)
		rest = append(rest, chain[:i]...)
		rest = append(rest, chain[i+1:]...)

		for _, p := range permutations(rest) {
			out = append(out, append([]tuplet.Tuplet{chain[i]}, p...))
		}
	}

	return out
}

// TestChain_OrderIndependence covers K-9 and architecture 4.10: results,
// including the presence and class of an error, do not depend on the order
// of the chain.
func TestChain_OrderIndependence(t *testing.T) {
	t.Parallel()

	chains := []struct {
		name  string
		chain []tuplet.Tuplet
	}{
		{name: "named", chain: []tuplet.Tuplet{tuplet.Quintuplet(), tuplet.Triplet(), tuplet.Duplet()}},
		{name: "flatten overflow", chain: []tuplet.Tuplet{tuplet.MustNew(math.MaxInt, 1), tuplet.MustNew(2, 1), tuplet.MustNew(1, 1)}},
		{name: "normal overflow", chain: []tuplet.Tuplet{tuplet.MustNew(1, 1<<40), tuplet.MustNew(1, 1<<20), tuplet.MustNew(3, 1<<3)}},
		{name: "fits exactly", chain: []tuplet.Tuplet{tuplet.MustNew(1<<30, 1), tuplet.MustNew(1<<30, 3), tuplet.MustNew(1<<2, 5)}},
		{
			name: "coefficient overflow",
			chain: []tuplet.Tuplet{
				tuplet.MustNew(1, math.MaxInt), tuplet.MustNew(1, math.MaxInt), tuplet.MustNew(1, math.MaxInt),
				tuplet.MustNew(1, math.MaxInt), tuplet.MustNew(2, 1<<62),
			},
		},
	}

	for _, tt := range chains {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			wantFlat, wantFlatErr := tuplet.Flatten(tt.chain...)
			wantCoefficient, wantCoefficientErr := tuplet.ChainCoefficient(tt.chain...)
			wantApplied, wantAppliedErr := tuplet.ApplyChain(frac(1, 8), tt.chain...)

			for _, p := range permutations(tt.chain) {
				flat, err := tuplet.Flatten(p...)
				assert.Equal(t, wantFlat, flat, "Flatten(%s)", chainName(p))
				assert.Equal(t, wantFlatErr == nil, err == nil, "Flatten error presence for %s", chainName(p))

				if wantFlatErr != nil {
					assertErr(t, err, tuplet.ErrOutOfRange)
				}

				c, err := tuplet.ChainCoefficient(p...)
				assert.Equal(t, wantCoefficient, c, "ChainCoefficient(%s)", chainName(p))
				assert.Equal(t, wantCoefficientErr == nil, err == nil, "ChainCoefficient error presence for %s", chainName(p))

				if wantCoefficientErr != nil {
					assertErr(t, err, tuplet.ErrOutOfRange, fraction.ErrOutOfRange)
				}

				applied, err := tuplet.ApplyChain(frac(1, 8), p...)
				assert.Equal(t, wantApplied, applied, "ApplyChain(%s)", chainName(p))
				assert.Equal(t, wantAppliedErr == nil, err == nil, "ApplyChain error presence for %s", chainName(p))
			}
		})
	}
}

// TestNest_Overflow covers E-5 for Flatten and Nest: a product above
// math.MaxInt in either count.
func TestNest_Overflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		outer, inner tuplet.Tuplet
	}{
		{name: "actual", outer: tuplet.MustNew(math.MaxInt, 1), inner: tuplet.MustNew(2, 1)},
		{name: "normal", outer: tuplet.MustNew(1, math.MaxInt), inner: tuplet.MustNew(1, 2)},
		{name: "just above max", outer: tuplet.MustNew(1<<32, 1), inner: tuplet.MustNew(1<<31, 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.outer.Nest(tt.inner)
			assertErr(t, err, tuplet.ErrOutOfRange)
			assert.Equal(t, tuplet.Tuplet{}, got)

			got, err = tt.inner.Nest(tt.outer)
			assertErr(t, err, tuplet.ErrOutOfRange)
			assert.Equal(t, tuplet.Tuplet{}, got)

			got, err = tuplet.Flatten(tt.outer, tt.inner)
			assertErr(t, err, tuplet.ErrOutOfRange)
			assert.Equal(t, tuplet.Tuplet{}, got)
		})
	}

	got, err := tuplet.MustNew(1<<31, 1).Nest(tuplet.MustNew(1<<31, 1))
	require.NoError(t, err)
	assertTuplet(t, got, 1<<62, 1)
}

// TestChainCoefficient_Range covers K-9 and architecture 4.10: the chain
// coefficient succeeds exactly when both products are at most 2**255 - 1;
// any four groups fit; five groups with normal math.MaxInt do not.
func TestChainCoefficient_Range(t *testing.T) {
	t.Parallel()

	maxPow := tuplet.MustNew(1, 1<<62)
	maxInt := tuplet.MustNew(1, math.MaxInt)
	maxBoth := tuplet.MustNew(math.MaxInt, math.MaxInt)
	mi := bigOf(math.MaxInt)
	mi4 := new(big.Int).Exp(mi, big.NewInt(4), nil)

	t.Run("four groups with max normal", func(t *testing.T) {
		t.Parallel()

		c, err := tuplet.ChainCoefficient(maxInt, maxInt, maxInt, maxInt)
		require.NoError(t, err)
		assertBigPair(t, c, mi4, big.NewInt(1))
	})

	t.Run("four groups with max counts", func(t *testing.T) {
		t.Parallel()

		c, err := tuplet.ChainCoefficient(maxBoth, maxBoth, maxBoth, maxBoth)
		require.NoError(t, err)
		assertBigPair(t, c, mi4, mi4)
	})

	t.Run("five groups with max normal", func(t *testing.T) {
		t.Parallel()

		c, err := tuplet.ChainCoefficient(maxInt, maxInt, maxInt, maxInt, maxInt)
		assertErr(t, err, tuplet.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, fraction.Fraction{}, c)
	})

	t.Run("five groups with max actual", func(t *testing.T) {
		t.Parallel()

		maxActual := tuplet.MustNew(math.MaxInt, 1)
		c, err := tuplet.ChainCoefficient(maxActual, maxActual, maxActual, maxActual, maxActual)
		assertErr(t, err, tuplet.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, fraction.Fraction{}, c)
	})

	t.Run("normal product 2**255 - 2**248 fits", func(t *testing.T) {
		t.Parallel()

		c, err := tuplet.ChainCoefficient(maxPow, maxPow, maxPow, maxPow, tuplet.MustNew(1, 127))
		require.NoError(t, err)
		assertBigPair(t, c, new(big.Int).Mul(pow2(248), big.NewInt(127)), big.NewInt(1))
	})

	t.Run("normal product 2**255 does not fit", func(t *testing.T) {
		t.Parallel()

		c, err := tuplet.ChainCoefficient(maxPow, maxPow, maxPow, maxPow, tuplet.MustNew(1, 128))
		assertErr(t, err, tuplet.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, fraction.Fraction{}, c)
	})

	t.Run("actual product 2**255 does not fit", func(t *testing.T) {
		t.Parallel()

		actualPow := tuplet.MustNew(1<<62, 1)
		c, err := tuplet.ChainCoefficient(actualPow, actualPow, actualPow, actualPow, tuplet.MustNew(128, 1))
		assertErr(t, err, tuplet.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, fraction.Fraction{}, c)
	})

	t.Run("ApplyChain propagates the chain error", func(t *testing.T) {
		t.Parallel()

		got, err := tuplet.ApplyChain(frac(1, 8), maxInt, maxInt, maxInt, maxInt, maxInt)
		assertErr(t, err, tuplet.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, fraction.Fraction{}, got)
	})
}

// TestApplyChain_Errors covers E-2, E-5, E-6 and the order E-2 -> E-6 -> E-5.
func TestApplyChain_Errors(t *testing.T) {
	t.Parallel()

	maxInt := tuplet.MustNew(1, math.MaxInt)
	overflowChain := []tuplet.Tuplet{maxInt, maxInt, maxInt, maxInt, maxInt}

	tests := []struct {
		name   string
		d      fraction.Fraction
		chain  []tuplet.Tuplet
		want   error
		causes []error
	}{
		{name: "unset duration", d: fraction.Fraction{}, chain: []tuplet.Tuplet{tuplet.Triplet()}, want: tuplet.ErrInvalidArgument},
		{name: "unset duration with empty chain", d: fraction.Fraction{}, want: tuplet.ErrInvalidArgument},
		{name: "unset element", d: frac(1, 8), chain: []tuplet.Tuplet{tuplet.Triplet(), {}}, want: tuplet.ErrUnset},
		{name: "unset element before unset duration", d: fraction.Fraction{}, chain: []tuplet.Tuplet{{}}, want: tuplet.ErrUnset},
		{name: "unset duration before chain overflow", d: fraction.Fraction{}, chain: overflowChain, want: tuplet.ErrInvalidArgument},
		{
			name: "unset element before chain overflow", d: frac(1, 8),
			chain: append(append([]tuplet.Tuplet(nil), overflowChain...), tuplet.Tuplet{}), want: tuplet.ErrUnset,
		},
		{
			name: "result out of range", d: bigFrac(t, fraction.MaxComponentBig(), big.NewInt(1)),
			chain: []tuplet.Tuplet{tuplet.Duplet()}, want: tuplet.ErrOutOfRange, causes: []error{fraction.ErrOutOfRange},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ApplyChain(tt.d, tt.chain...)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, fraction.Fraction{}, got)
		})
	}
}
