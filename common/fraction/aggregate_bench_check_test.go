package fraction_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// References of requirement T-3 (architecture of task 26.7): the computations
// the consumers use now, in exactly their form, every element converted by
// BigRat. They are the baselines of the benchmarks and are checked against the
// functions on every benchmark input.

// agRefSum is the form of duration.Sum: acc.Add(acc, x.BigRat()), then FromBigRat.
func agRefSum(xs []fraction.Fraction) (fraction.Fraction, error) {
	acc := new(big.Rat)
	for _, x := range xs {
		r, _ := x.BigRat()
		acc.Add(acc, r)
	}
	return fraction.FromBigRat(acc)
}

// agRefProduct is the form of mul in tempo: acc.Mul(acc, x.BigRat()) from 1, then FromBigRat.
func agRefProduct(xs []fraction.Fraction) (fraction.Fraction, error) {
	acc := big.NewRat(1, 1)
	for _, x := range xs {
		r, _ := x.BigRat()
		acc.Mul(acc, r)
	}
	return fraction.FromBigRat(acc)
}

// agRefLCM is the form of score: ReducedBigPair of both fractions, GCD of the
// numerators, Quo, Mul, GCD of the denominators, FromBigIntReduced.
func agRefLCM(x, y fraction.Fraction) (fraction.Fraction, error) {
	n1, d1, _ := x.ReducedBigPair()
	n2, d2, _ := y.ReducedBigPair()
	g := new(big.Int).GCD(nil, nil, n1, n2)
	num := new(big.Int).Quo(n1, g)
	num.Mul(num, n2)
	den := new(big.Int).GCD(nil, nil, d1, d2)
	return fraction.FromBigIntReduced(num, den)
}

// agRefCommonDenominatorBig is the form of format/ticks: lcm from 1, for every
// den := x.BigRat().Denom(), g := GCD(lcm, den), lcm.Mul(lcm, Quo(den, g)).
func agRefCommonDenominatorBig(xs []fraction.Fraction) *big.Int {
	lcm := big.NewInt(1)
	for _, x := range xs {
		r, _ := x.BigRat()
		den := r.Denom()
		g := new(big.Int).GCD(nil, nil, lcm, den)
		lcm.Mul(lcm, new(big.Int).Quo(den, g))
	}
	return lcm
}

// agRefCommonDenominator is the narrow form: the result converted to int.
func agRefCommonDenominator(xs []fraction.Fraction) (int, bool) {
	lcm := agRefCommonDenominatorBig(xs)
	if !lcm.IsInt64() {
		return 0, false
	}
	return int(lcm.Int64()), true
}

// agCheckElementBits checks the widths of both original components of every element.
func agCheckElementBits(t *testing.T, label, class string, fs ...fraction.Fraction) {
	t.Helper()
	for _, f := range fs {
		num, den, ok := f.BigPair()
		require.Truef(t, ok, "%s: valid element", label)
		if num.Sign() != 0 {
			benchComponentBits(t, class, label+" numerator", num, false)
		}
		benchComponentBits(t, class, label+" denominator", den, true)
	}
}

// TestAggregateBenchmarkInputs checks the inputs of the task 26.7 benchmarks:
// the classes and sizes present, component widths, nontrivial reductions, the
// size of the common factor of the pairs, prefixes in range on the path
// without the exact continuation and out of range on the overflow path,
// success of every operation and agreement with the oracle and the references.
func TestAggregateBenchmarkInputs(t *testing.T) {
	t.Parallel()
	in := agInputs()

	t.Run("coverage", func(t *testing.T) {
		t.Parallel()
		count := func(sets []agSetInput) map[string]int {
			m := make(map[string]int)
			for _, s := range sets {
				m[s.class]++
			}
			return m
		}
		for _, class := range agClasses {
			assert.Equal(t, len(agSizes), count(in.sum)[class], "Sum sets of %s", class)
			assert.Equal(t, len(agSizes), count(in.commonDen)[class], "CommonDenominator sets of %s", class)
		}
		assert.Len(t, in.product, len(agClasses)*len(agSizes)-1, "Product sets")
		assert.Equal(t, []string{"alt/64"}, in.skippedProduct)
		assert.Len(t, in.pairs, len(agClasses))
		// Overflow sets: impossible for n = 2 and where n elements cannot hold
		// the out-of-range prefix and the cancellation.
		assert.Equal(t, []string{"w31/2", "w31/8", "alt/2", "alt/8", "w63/2", "w128/2", "w255/2"}, in.skippedSumOF)
		assert.Equal(t, []string{"w31/2", "w31/8", "alt/2", "w63/2", "w128/2", "w255/2"}, in.skippedProdOF)
	})

	t.Run("Sum", func(t *testing.T) {
		t.Parallel()
		for _, s := range in.sum {
			label := "Sum " + s.name()
			require.Len(t, s.fs, s.n, label)
			agCheckElementBits(t, label, s.class, s.fs...)
			d := big.NewInt(1)
			for _, f := range s.fs {
				num, den, _ := f.BigPair()
				benchNontrivialGCD(t, label, num, den)
				p, q := agReduced(f)
				assert.LessOrEqual(t, p.BitLen(), 31, "%s: reduced numerator", label)
				assert.LessOrEqual(t, q.BitLen(), 31, "%s: reduced denominator", label)
				d = agLCMInt(d, q)
			}
			assert.LessOrEqual(t, d.BitLen(), 62, "%s: D ≤ 2^62", label)
			assert.True(t, agPrefixesInRange(s.fs, false), "%s: every reduced prefix sum in range", label)
			got, err := fraction.Sum(s.fs...)
			wCheckFrac(t, label, got, err, agWantSum(s.fs))
			ref, err := agRefSum(s.fs)
			require.NoError(t, err, label)
			assert.True(t, ref == got, "%s: reference %s, Sum %s", label, ref, got)
		}
	})

	t.Run("Product", func(t *testing.T) {
		t.Parallel()
		for _, s := range in.product {
			label := "Product " + s.name()
			require.Len(t, s.fs, s.n, label)
			agCheckElementBits(t, label, s.class, s.fs...)
			assert.True(t, agPrefixesInRange(s.fs, true), "%s: every reduced prefix product in range", label)
			// Every step of the fast path reduces: NOD of the products of the
			// reduced partial result and the next element is above 1.
			for i := 1; i < len(s.fs); i++ {
				acc := agExactProduct(s.fs[:i])
				num, den, _ := s.fs[i].BigPair()
				g := wGCD(wMul(acc.Num(), num), wMul(acc.Denom(), den))
				assert.Equal(t, 1, g.Cmp(big.NewInt(1)), "%s: step %d must reduce", label, i)
			}
			got, err := fraction.Product(s.fs...)
			wCheckFrac(t, label, got, err, agWantProduct(s.fs))
			ref, err := agRefProduct(s.fs)
			require.NoError(t, err, label)
			assert.True(t, ref == got, "%s: reference %s, Product %s", label, ref, got)
		}
	})

	t.Run("overflow", func(t *testing.T) {
		t.Parallel()
		for _, group := range []struct {
			product bool
			sets    []agSetInput
		}{{false, in.sumOverflow}, {true, in.prodOverflow}} {
			for _, s := range group.sets {
				label := s.name()
				require.Len(t, s.fs, s.n, label)
				assert.False(t, agPrefixesInRange(s.fs, group.product), "%s: a prefix must be out of range", label)
				var got, ref fraction.Fraction
				var err, refErr error
				if group.product {
					got, err = fraction.Product(s.fs...)
					wCheckFrac(t, "Product "+label, got, err, agWantProduct(s.fs))
					ref, refErr = agRefProduct(s.fs)
				} else {
					got, err = fraction.Sum(s.fs...)
					wCheckFrac(t, "Sum "+label, got, err, agWantSum(s.fs))
					ref, refErr = agRefSum(s.fs)
				}
				require.NoError(t, refErr, label)
				assert.True(t, ref == got, "%s: reference %s, result %s", label, ref, got)
			}
		}
	})

	t.Run("CommonDenominator", func(t *testing.T) {
		t.Parallel()
		for _, s := range in.commonDen {
			label := "CommonDenominator " + s.name()
			require.Len(t, s.fs, s.n, label)
			agCheckElementBits(t, label, s.class, s.fs...)
			for _, f := range s.fs {
				num, den, _ := f.BigPair()
				benchNontrivialGCD(t, label, num, den)
				_, q := agReduced(f)
				assert.LessOrEqual(t, q.BitLen(), 62, "%s: reduced denominator ≤ 2^62", label)
			}
			want, err := agWantCommonDenominator(s.fs)
			require.NoError(t, err, "%s: D ≤ math.MaxInt", label)
			assert.Greater(t, want, 1, label)
			ref, ok := agRefCommonDenominator(s.fs)
			require.True(t, ok, label)
			assert.Equal(t, want, ref, "%s: reference", label)
			got, err := fraction.CommonDenominator(s.fs...)
			if assert.NoError(t, err, label) {
				assert.Equal(t, want, got, label)
			}
			gotBig, err := fraction.CommonDenominatorBig(s.fs...)
			if assert.NoError(t, err, label) {
				wCheckBigInt(t, label+" big", gotBig, agRefCommonDenominatorBig(s.fs))
			}
		}
	})

	t.Run("GCD and LCM", func(t *testing.T) {
		t.Parallel()
		for _, p := range in.pairs {
			label := "pair " + p.class
			agCheckElementBits(t, label, p.class, p.x, p.y)
			for _, f := range []fraction.Fraction{p.x, p.y} {
				num, den, _ := f.BigPair()
				g := benchNontrivialGCD(t, label, num, den)
				if bits, ok := benchGBits[p.class]; ok {
					assert.InDelta(t, bits, g.BitLen(), 2, "%s: common factor near a quarter of the width", label)
				}
			}
			px, qx := agReduced(p.x)
			py, qy := agReduced(p.y)
			benchNontrivialGCD(t, label+" numerators", px, py)
			benchNontrivialGCD(t, label+" denominators", qx, qy)
			got, err := p.x.GCD(p.y)
			wCheckFrac(t, label+" GCD", got, err, agWantGCD(p.x, p.y))
			got, err = p.x.LCM(p.y)
			wCheckFrac(t, label+" LCM", got, err, agWantLCM(p.x, p.y))
			ref, err := agRefLCM(p.x, p.y)
			require.NoError(t, err, label)
			assert.True(t, ref == got, "%s: reference %s, LCM %s", label, ref, got)
		}
	})

	t.Run("BigRatTo", func(t *testing.T) {
		t.Parallel()
		var classes []string
		for _, bi := range agBigRatToInputs() {
			classes = append(classes, bi.class)
			num, den, _ := bi.x.BigPair()
			benchNontrivialGCD(t, "BigRatTo "+bi.class, num, den)
			agCheckBigRatTo(t, "BigRatTo "+bi.class, bi.x, nil)
			agCheckBigRatTo(t, "BigRatTo "+bi.class, bi.x, new(big.Rat))
		}
		assert.Equal(t, []string{"w31", "w63", "w128", "w255"}, classes)
	})
}

// TestAggregate_Allocs checks T-2 on the benchmark inputs: GCD, LCM and
// CommonDenominator without allocations in every class; Sum and Product
// without allocations when every reduced prefix is in range, both for
// Sum(xs...) and for a call with listed arguments; CommonDenominatorBig with
// at most 2 allocations for D ≤ 2^512; BigRatTo without allocations when z is
// reused and with at most 4 for nil z. It is not parallel: AllocsPerRun
// measures process-wide allocation counters.
func TestAggregate_Allocs(t *testing.T) {
	in := agInputs()
	const runs = 100
	for _, p := range in.pairs {
		t.Run("GCD/"+p.class, func(t *testing.T) {
			x, y := p.x, p.y
			_, err := x.GCD(y)
			require.NoError(t, err)
			assert.Zero(t, testing.AllocsPerRun(runs, func() { _, _ = x.GCD(y) }))
		})
		t.Run("LCM/"+p.class, func(t *testing.T) {
			x, y := p.x, p.y
			_, err := x.LCM(y)
			require.NoError(t, err)
			assert.Zero(t, testing.AllocsPerRun(runs, func() { _, _ = x.LCM(y) }))
		})
	}
	for _, s := range in.commonDen {
		t.Run("CommonDenominator/"+s.name(), func(t *testing.T) {
			fs := s.fs
			_, err := fraction.CommonDenominator(fs...)
			require.NoError(t, err)
			assert.Zero(t, testing.AllocsPerRun(runs, func() { _, _ = fraction.CommonDenominator(fs...) }))
		})
		t.Run("CommonDenominatorBig/"+s.name(), func(t *testing.T) {
			fs := s.fs
			d, err := fraction.CommonDenominatorBig(fs...)
			require.NoError(t, err)
			require.LessOrEqual(t, d.BitLen(), 512)
			assert.LessOrEqual(t, testing.AllocsPerRun(runs, func() { _, _ = fraction.CommonDenominatorBig(fs...) }), 2.0)
		})
	}
	sumProduct := []struct {
		name string
		sets []agSetInput
		call func(fs ...fraction.Fraction) (fraction.Fraction, error)
	}{
		{"Sum", in.sum, fraction.Sum},
		{"Product", in.product, fraction.Product},
	}
	for _, op := range sumProduct {
		for _, s := range op.sets {
			t.Run(op.name+"/"+s.name(), func(t *testing.T) {
				fs := s.fs
				_, err := op.call(fs...)
				require.NoError(t, err)
				a, b, c := fs[0], fs[1], fs[len(fs)-1]
				if op.name == "Sum" {
					assert.Zero(t, testing.AllocsPerRun(runs, func() { _, _ = fraction.Sum(fs...) }), "Sum(xs...)")
					if len(fs) > 2 {
						c = fs[2]
						assert.Zero(t, testing.AllocsPerRun(runs, func() { _, _ = fraction.Sum(a, b, c) }), "Sum(a, b, c)")
					} else {
						assert.Zero(t, testing.AllocsPerRun(runs, func() { _, _ = fraction.Sum(a, b) }), "Sum(a, b)")
					}
					return
				}
				assert.Zero(t, testing.AllocsPerRun(runs, func() { _, _ = fraction.Product(fs...) }), "Product(xs...)")
				if len(fs) > 2 {
					c = fs[2]
					assert.Zero(t, testing.AllocsPerRun(runs, func() { _, _ = fraction.Product(a, b, c) }), "Product(a, b, c)")
				} else {
					assert.Zero(t, testing.AllocsPerRun(runs, func() { _, _ = fraction.Product(a, b) }), "Product(a, b)")
				}
			})
		}
	}
	for _, bi := range agBigRatToInputs() {
		t.Run("BigRatTo/"+bi.class, func(t *testing.T) {
			x := bi.x
			z := new(big.Rat)
			_, ok := x.BigRatTo(z)
			require.True(t, ok)
			assert.Zero(t, testing.AllocsPerRun(runs, func() { _, _ = x.BigRatTo(z) }), "reused z")
			assert.LessOrEqual(t, testing.AllocsPerRun(runs, func() { _, _ = x.BigRatTo(nil) }), 4.0, "nil z")
		})
	}
	// The narrow common denominator also stays without allocations at its
	// boundary, where the result is math.MaxInt.
	t.Run("CommonDenominator/MaxInt", func(t *testing.T) {
		fs := []fraction.Fraction{fraction.MustNew(1, 3577), fraction.MustNew(1, math.MaxInt/3577)}
		d, err := fraction.CommonDenominator(fs...)
		require.NoError(t, err)
		require.Equal(t, math.MaxInt, d)
		assert.Zero(t, testing.AllocsPerRun(runs, func() { _, _ = fraction.CommonDenominator(fs...) }))
	})
}
