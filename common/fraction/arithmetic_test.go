package fraction_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnaryOperations verifies criterion K-11 (Neg, Abs, Reciprocal, Inc, Dec and Must*).
func TestUnaryOperations(t *testing.T) {
	t.Parallel()

	t.Run("Neg", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			num     int
			den     int
			wantNum int
			wantDen int
		}{
			{name: "1/2->-1/2", num: 1, den: 2, wantNum: -1, wantDen: 2},
			{name: "1/-2->1/2", num: 1, den: -2, wantNum: 1, wantDen: 2},
			{name: "2/4->-2/4", num: 2, den: 4, wantNum: -2, wantDen: 4},
			{name: "0/3->0/3", num: 0, den: 3, wantNum: 0, wantDen: 3},
			{name: "0/-5->0/5", num: 0, den: -5, wantNum: 0, wantDen: 5},
			{name: "-2/-4->-2/4", num: -2, den: -4, wantNum: -2, wantDen: 4},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				orig := fraction.MustNew(tc.num, tc.den)
				got, err := orig.Neg()
				require.NoError(t, err)
				assertExactPairInt(t, got, tc.wantNum, tc.wantDen)
				assert.Equal(t, got, orig.MustNeg())

				// Double negation restores numeric value; if orig had den > 0, restores exact pair.
				twice, err := got.Neg()
				require.NoError(t, err)
				assert.True(t, twice.Equal(orig))
				if tc.den > 0 {
					assert.True(t, twice.EqualExact(orig))
				}

				// Input unchanged.
				assertExactPairInt(t, orig, tc.num, tc.den)
			})
		}
	})

	t.Run("Abs", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			num     int
			den     int
			wantNum int
			wantDen int
		}{
			{name: "-3/4->3/4", num: -3, den: 4, wantNum: 3, wantDen: 4},
			{name: "1/-2->1/2", num: 1, den: -2, wantNum: 1, wantDen: 2},
			{name: "-2/-4->2/4", num: -2, den: -4, wantNum: 2, wantDen: 4},
			{name: "0/-5->0/5", num: 0, den: -5, wantNum: 0, wantDen: 5},
			{name: "6/12->6/12", num: 6, den: 12, wantNum: 6, wantDen: 12},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				orig := fraction.MustNew(tc.num, tc.den)
				got, err := orig.Abs()
				require.NoError(t, err)
				assertExactPairInt(t, got, tc.wantNum, tc.wantDen)
				assert.Equal(t, got, orig.MustAbs())
				assert.GreaterOrEqual(t, got.Sign(), 0)
				assertExactPairInt(t, orig, tc.num, tc.den)
			})
		}
	})

	t.Run("Reciprocal", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			num     int
			den     int
			wantNum int
			wantDen int
		}{
			{name: "3/2->2/3", num: 3, den: 2, wantNum: 2, wantDen: 3},
			{name: "6/4->4/6", num: 6, den: 4, wantNum: 4, wantDen: 6},
			{name: "-2/3->-3/2", num: -2, den: 3, wantNum: -3, wantDen: 2},
			{name: "1/-2->-2/1", num: 1, den: -2, wantNum: -2, wantDen: 1},
			{name: "-2/-4->4/2", num: -2, den: -4, wantNum: 4, wantDen: 2},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				orig := fraction.MustNew(tc.num, tc.den)
				got, err := orig.Reciprocal()
				require.NoError(t, err)
				assertExactPairInt(t, got, tc.wantNum, tc.wantDen)
				assert.Equal(t, got, orig.MustReciprocal())

				if tc.num == 6 && tc.den == 4 {
					assertExactPairInt(t, got.Reduce(), 2, 3)
				}

				// Double reciprocal for positive denominator restores exact pair.
				twice, err := got.Reciprocal()
				require.NoError(t, err)
				assert.True(t, twice.Equal(orig))
				if tc.den > 0 {
					assert.True(t, twice.EqualExact(orig))
				}
			})
		}

		for _, z := range []rawPair{{0, 1}, {0, 3}, {0, -5}} {
			f := fraction.MustNew(z.num, z.den)
			_, err := f.Reciprocal()
			require.ErrorIs(t, err, fraction.ErrDivisionByZero)
			assertMustPanicsWithError(t, func() {
				_ = f.MustReciprocal()
			}, fraction.ErrDivisionByZero)
		}
	})

	t.Run("Inc_and_Dec", func(t *testing.T) {
		t.Parallel()

		half := fraction.MustNew(1, 2)
		inc, err := half.Inc()
		require.NoError(t, err)
		assertExactPairInt(t, inc, 3, 2)
		assert.Equal(t, inc, half.MustInc())

		dec, err := half.Dec()
		require.NoError(t, err)
		assertExactPairInt(t, dec, -1, 2)
		assert.Equal(t, dec, half.MustDec())

		unreducedHalf := fraction.MustNew(6, 12)
		assertExactPairInt(t, unreducedHalf.MustInc(), 3, 2)
		assertExactPairInt(t, unreducedHalf.MustDec(), -1, 2)
	})
}

// TestStepNumerator verifies criterion K-12.
func TestStepNumerator(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		num     int
		den     int
		delta   int
		wantNum int
		wantDen int
	}{
		{name: "6/12+1->7/12", num: 6, den: 12, delta: 1, wantNum: 7, wantDen: 12},
		{name: "6/12-1->5/12", num: 6, den: 12, delta: -1, wantNum: 5, wantDen: 12},
		{name: "6/12+6->12/12", num: 6, den: 12, delta: 6, wantNum: 12, wantDen: 12},
		{name: "6/12-6->0/12", num: 6, den: 12, delta: -6, wantNum: 0, wantDen: 12},
		{name: "6/12+0->6/12", num: 6, den: 12, delta: 0, wantNum: 6, wantDen: 12},
		{name: "1/-2+2->3/-2", num: 1, den: -2, delta: 2, wantNum: 3, wantDen: -2},
		{name: "1/-2-1->0/-2", num: 1, den: -2, delta: -1, wantNum: 0, wantDen: -2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orig := fraction.MustNew(tc.num, tc.den)
			got, err := orig.StepNumerator(tc.delta)
			require.NoError(t, err)
			assertExactPairInt(t, got, tc.wantNum, tc.wantDen)
			assert.Equal(t, got, orig.MustStepNumerator(tc.delta))
			assertExactPairInt(t, orig, tc.num, tc.den)
		})
	}
}

// TestBinaryArithmetic_Mode1_Reduced verifies criterion K-13 (Add, Sub, Mul, Div, *Int, DenomReduced, Must*).
func TestBinaryArithmetic_Mode1_Reduced(t *testing.T) {
	t.Parallel()

	t.Run("explicit_cases", func(t *testing.T) {
		t.Parallel()

		type opKind int
		const (
			opAdd opKind = iota
			opSub
			opMul
			opDiv
		)

		cases := []struct {
			name    string
			a       rawPair
			b       rawPair
			op      opKind
			wantNum int
			wantDen int
		}{
			{name: "1/3+1/6=1/2", a: rawPair{1, 3}, b: rawPair{1, 6}, op: opAdd, wantNum: 1, wantDen: 2},
			{name: "1/6+1/6=1/3", a: rawPair{1, 6}, b: rawPair{1, 6}, op: opAdd, wantNum: 1, wantDen: 3},
			{name: "1/-2+1/2=0/1", a: rawPair{1, -2}, b: rawPair{1, 2}, op: opAdd, wantNum: 0, wantDen: 1},
			{name: "1/2-1/3=1/6", a: rawPair{1, 2}, b: rawPair{1, 3}, op: opSub, wantNum: 1, wantDen: 6},
			{name: "1/4-1/4=0/1", a: rawPair{1, 4}, b: rawPair{1, 4}, op: opSub, wantNum: 0, wantDen: 1},
			{name: "1/8*2/3=1/12", a: rawPair{1, 8}, b: rawPair{2, 3}, op: opMul, wantNum: 1, wantDen: 12},
			{name: "3/4*4/3=1/1", a: rawPair{3, 4}, b: rawPair{4, 3}, op: opMul, wantNum: 1, wantDen: 1},
			{name: "(1/2)/(3/4)=2/3", a: rawPair{1, 2}, b: rawPair{3, 4}, op: opDiv, wantNum: 2, wantDen: 3},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				fa := fraction.MustNew(tc.a.num, tc.a.den)
				fb := fraction.MustNew(tc.b.num, tc.b.den)

				var got, gotWith, gotMust, gotMustWith fraction.Fraction
				var err, errWith error

				switch tc.op {
				case opAdd:
					got, err = fa.Add(fb)
					gotWith, errWith = fa.AddWith(fb, fraction.DenomReduced)
					gotMust = fa.MustAdd(fb)
					gotMustWith = fa.MustAddWith(fb, fraction.DenomReduced)
				case opSub:
					got, err = fa.Sub(fb)
					gotWith, errWith = fa.SubWith(fb, fraction.DenomReduced)
					gotMust = fa.MustSub(fb)
					gotMustWith = fa.MustSubWith(fb, fraction.DenomReduced)
				case opMul:
					got, err = fa.Mul(fb)
					gotWith, errWith = fa.MulWith(fb, fraction.DenomReduced)
					gotMust = fa.MustMul(fb)
					gotMustWith = fa.MustMulWith(fb, fraction.DenomReduced)
				case opDiv:
					got, err = fa.Div(fb)
					gotWith, errWith = fa.DivWith(fb, fraction.DenomReduced)
					gotMust = fa.MustDiv(fb)
					gotMustWith = fa.MustDivWith(fb, fraction.DenomReduced)
				}

				require.NoError(t, err)
				require.NoError(t, errWith)
				assertExactPairInt(t, got, tc.wantNum, tc.wantDen)
				assert.True(t, got.IsReduced())
				assert.Equal(t, got, gotWith)
				assert.Equal(t, got, gotMust)
				assert.Equal(t, got, gotMustWith)
			})
		}

		// Three triplet eighths: 1/12 + 1/12 + 1/12 = 1/4.
		t12 := fraction.MustNew(1, 12)
		sum3 := t12.MustAdd(t12).MustAdd(t12)
		assertExactPairInt(t, sum3, 1, 4)
	})

	t.Run("int_operations", func(t *testing.T) {
		t.Parallel()

		half := fraction.MustNew(1, 2)
		add1, err := half.AddInt(1)
		require.NoError(t, err)
		assertExactPairInt(t, add1, 3, 2)
		assert.Equal(t, add1, half.MustAddInt(1))

		sub1, err := half.SubInt(1)
		require.NoError(t, err)
		assertExactPairInt(t, sub1, -1, 2)
		assert.Equal(t, sub1, half.MustSubInt(1))

		mul2, err := half.MulInt(2)
		require.NoError(t, err)
		assertExactPairInt(t, mul2, 1, 1)
		assert.Equal(t, mul2, half.MustMulInt(2))

		threeQuarters := fraction.MustNew(3, 4)
		div3, err := threeQuarters.DivInt(3)
		require.NoError(t, err)
		assertExactPairInt(t, div3, 1, 4)
		assert.Equal(t, div3, threeQuarters.MustDivInt(3))
	})

	t.Run("identities_and_oracle_on_S", func(t *testing.T) {
		t.Parallel()

		s := sampleSetS()
		zero := fraction.Zero()
		one := fraction.One()
		minusOne := fraction.MinusOne()

		for _, pa := range s {
			fa := fraction.MustNew(pa.num, pa.den)
			ra := big.NewRat(int64(pa.num), int64(pa.den))

			// Neutral operands (G-1).
			assert.Equal(t, fa.Reduce(), fa.MustAdd(zero))
			assert.Equal(t, fa.Reduce(), fa.MustSub(zero))
			assert.Equal(t, zero, fa.MustSub(fa))
			assert.Equal(t, zero, fa.MustMul(zero))
			assert.Equal(t, fa.Reduce(), fa.MustMul(one))
			assert.Equal(t, fa.MustNeg().Reduce(), fa.MustMul(minusOne))
			assert.Equal(t, fa.Reduce(), fa.MustDiv(one))
			assert.Equal(t, fa.MustNeg().Reduce(), fa.MustDiv(minusOne))

			for _, pb := range s {
				fb := fraction.MustNew(pb.num, pb.den)
				rb := big.NewRat(int64(pb.num), int64(pb.den))

				wantAdd := new(big.Rat).Add(ra, rb)
				gotAdd, err := fa.Add(fb)
				require.NoError(t, err)
				assertExactPairInt(t, gotAdd, int(wantAdd.Num().Int64()), int(wantAdd.Denom().Int64()))

				wantSub := new(big.Rat).Sub(ra, rb)
				gotSub, err := fa.Sub(fb)
				require.NoError(t, err)
				assertExactPairInt(t, gotSub, int(wantSub.Num().Int64()), int(wantSub.Denom().Int64()))

				wantMul := new(big.Rat).Mul(ra, rb)
				gotMul, err := fa.Mul(fb)
				require.NoError(t, err)
				assertExactPairInt(t, gotMul, int(wantMul.Num().Int64()), int(wantMul.Denom().Int64()))

				if pb.num != 0 {
					wantDiv := new(big.Rat).Quo(ra, rb)
					gotDiv, err := fa.Div(fb)
					require.NoError(t, err)
					assertExactPairInt(t, gotDiv, int(wantDiv.Num().Int64()), int(wantDiv.Denom().Int64()))
				}
			}
		}
	})

	t.Run("division_by_zero_errors", func(t *testing.T) {
		t.Parallel()

		a := fraction.MustNew(1, 2)
		for _, zp := range []rawPair{{0, 1}, {0, 3}, {0, -5}} {
			zf := fraction.MustNew(zp.num, zp.den)
			_, err := a.Div(zf)
			require.ErrorIs(t, err, fraction.ErrDivisionByZero)
			assertMustPanicsWithError(t, func() {
				_ = a.MustDiv(zf)
			}, fraction.ErrDivisionByZero)

			_, err = a.DivWith(zf, fraction.DenomReduced)
			require.ErrorIs(t, err, fraction.ErrDivisionByZero)
			assertMustPanicsWithError(t, func() {
				_ = a.MustDivWith(zf, fraction.DenomReduced)
			}, fraction.ErrDivisionByZero)
		}

		_, err := a.DivInt(0)
		require.ErrorIs(t, err, fraction.ErrDivisionByZero)
		assertMustPanicsWithError(t, func() {
			_ = a.MustDivInt(0)
		}, fraction.ErrDivisionByZero)
	})
}

// TestBinaryArithmetic_Mode2_UnreducedLCM verifies criterion K-14.
func TestBinaryArithmetic_Mode2_UnreducedLCM(t *testing.T) {
	t.Parallel()

	f := func(n, d int) fraction.Fraction { return fraction.MustNew(n, d) }

	assertExactPairInt(t, f(1, 6).MustAddWith(f(1, 6), fraction.DenomUnreducedLCM), 2, 6)
	assertExactPairInt(t, f(1, 4).MustAddWith(f(1, 6), fraction.DenomUnreducedLCM), 5, 12)
	assertExactPairInt(t, f(1, 4).MustAddWith(f(1, 12), fraction.DenomUnreducedLCM), 4, 12)
	assertExactPairInt(t, f(1, 4).MustSubWith(f(1, 4), fraction.DenomUnreducedLCM), 0, 4)
	assertExactPairInt(t, f(1, -6).MustAddWith(f(2, 6), fraction.DenomUnreducedLCM), 1, 6)
	assertExactPairInt(t, f(2, 3).MustMulWith(f(3, 4), fraction.DenomUnreducedLCM), 6, 12)
	assertExactPairInt(t, f(1, 2).MustDivWith(f(3, 4), fraction.DenomUnreducedLCM), 4, 6)
	assertExactPairInt(t, f(1, 2).MustDivWith(f(-3, 4), fraction.DenomUnreducedLCM), -4, 6)
	assertExactPairInt(t, f(1, -2).MustDivWith(f(3, -4), fraction.DenomUnreducedLCM), 4, 6)

	_, err := f(1, 2).DivWith(f(0, 3), fraction.DenomUnreducedLCM)
	require.ErrorIs(t, err, fraction.ErrDivisionByZero)
	assertMustPanicsWithError(t, func() {
		_ = f(1, 2).MustDivWith(f(0, 3), fraction.DenomUnreducedLCM)
	}, fraction.ErrDivisionByZero)
}

// TestBinaryArithmetic_Mode3_UnreducedProduct verifies criterion K-15.
func TestBinaryArithmetic_Mode3_UnreducedProduct(t *testing.T) {
	t.Parallel()

	f := func(n, d int) fraction.Fraction { return fraction.MustNew(n, d) }

	assertExactPairInt(t, f(1, 6).MustAddWith(f(1, 6), fraction.DenomUnreducedProduct), 12, 36)
	assertExactPairInt(t, f(1, 4).MustAddWith(f(1, 6), fraction.DenomUnreducedProduct), 10, 24)
	assertExactPairInt(t, f(3, 4).MustSubWith(f(1, 4), fraction.DenomUnreducedProduct), 8, 16)
	assertExactPairInt(t, f(1, -4).MustAddWith(f(1, 6), fraction.DenomUnreducedProduct), -2, 24)
	assertExactPairInt(t, f(2, 3).MustMulWith(f(3, 4), fraction.DenomUnreducedProduct), 6, 12)
	assertExactPairInt(t, f(1, 2).MustDivWith(f(3, 4), fraction.DenomUnreducedProduct), 4, 6)

	_, err := f(1, 2).DivWith(f(0, 3), fraction.DenomUnreducedProduct)
	require.ErrorIs(t, err, fraction.ErrDivisionByZero)
	assertMustPanicsWithError(t, func() {
		_ = f(1, 2).MustDivWith(f(0, 3), fraction.DenomUnreducedProduct)
	}, fraction.ErrDivisionByZero)
}

// TestBinaryArithmetic_Mode4_SameUnit verifies criterion K-16.
func TestBinaryArithmetic_Mode4_SameUnit(t *testing.T) {
	t.Parallel()

	f := func(n, d int) fraction.Fraction { return fraction.MustNew(n, d) }

	assertExactPairInt(t, f(1, 6).MustAddWith(f(1, 6), fraction.DenomSameUnit), 2, 6)
	assertExactPairInt(t, f(5, 12).MustSubWith(f(1, 12), fraction.DenomSameUnit), 4, 12)
	assertExactPairInt(t, f(1, -6).MustAddWith(f(2, 6), fraction.DenomSameUnit), 1, 6)
	assertExactPairInt(t, f(2, 4).MustMulWith(f(2, 4), fraction.DenomSameUnit), 1, 4)
	assertExactPairInt(t, f(2, 4).MustDivWith(f(4, 4), fraction.DenomSameUnit), 2, 4)

	// Different |d1| != |d2| -> ErrNotRepresentable.
	_, err := f(1, 4).AddWith(f(1, 8), fraction.DenomSameUnit)
	require.ErrorIs(t, err, fraction.ErrNotRepresentable)
	_, err = f(1, 4).SubWith(f(1, 8), fraction.DenomSameUnit)
	require.ErrorIs(t, err, fraction.ErrNotRepresentable)
	_, err = f(1, 4).MulWith(f(1, 8), fraction.DenomSameUnit)
	require.ErrorIs(t, err, fraction.ErrNotRepresentable)
	_, err = f(1, 4).DivWith(f(1, 8), fraction.DenomSameUnit)
	require.ErrorIs(t, err, fraction.ErrNotRepresentable)

	// Same |d1| == |d2| = 4, but product/quotient not representable in fourths.
	_, err = f(1, 4).MulWith(f(1, 4), fraction.DenomSameUnit)
	require.ErrorIs(t, err, fraction.ErrNotRepresentable)
	_, err = f(1, 4).DivWith(f(3, 4), fraction.DenomSameUnit)
	require.ErrorIs(t, err, fraction.ErrNotRepresentable)

	// Division by zero precedes D1 != D2 check (Step 2 before Step 3).
	_, err = f(1, 4).DivWith(f(0, 8), fraction.DenomSameUnit)
	require.ErrorIs(t, err, fraction.ErrDivisionByZero)
}

// TestBinaryArithmetic_Mode5And6_LargerAndSmallerUnit verifies criterion K-17.
func TestBinaryArithmetic_Mode5And6_LargerAndSmallerUnit(t *testing.T) {
	t.Parallel()

	f := func(n, d int) fraction.Fraction { return fraction.MustNew(n, d) }

	t.Run("DenomLargerUnit_min_denominator", func(t *testing.T) {
		t.Parallel()

		assertExactPairInt(t, f(1, 4).MustAddWith(f(2, 8), fraction.DenomLargerUnit), 2, 4)
		assertExactPairInt(t, f(3, 4).MustSubWith(f(2, 8), fraction.DenomLargerUnit), 2, 4)
		assertExactPairInt(t, f(2, 2).MustMulWith(f(2, 4), fraction.DenomLargerUnit), 1, 2)
		assertExactPairInt(t, f(2, 4).MustDivWith(f(2, 2), fraction.DenomLargerUnit), 1, 2)

		_, err := f(1, 4).AddWith(f(1, 8), fraction.DenomLargerUnit)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		_, err = f(1, 4).AddWith(f(1, 6), fraction.DenomLargerUnit)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		_, err = f(1, 2).MulWith(f(2, 4), fraction.DenomLargerUnit)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		_, err = f(1, 4).DivWith(f(2, 2), fraction.DenomLargerUnit)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)

		assertMustPanicsWithError(t, func() {
			_ = f(1, 4).MustAddWith(f(1, 8), fraction.DenomLargerUnit)
		}, fraction.ErrNotRepresentable)
	})

	t.Run("DenomSmallerUnit_max_denominator", func(t *testing.T) {
		t.Parallel()

		assertExactPairInt(t, f(1, 4).MustAddWith(f(2, 8), fraction.DenomSmallerUnit), 4, 8)
		assertExactPairInt(t, f(1, 4).MustAddWith(f(1, 8), fraction.DenomSmallerUnit), 3, 8)
		assertExactPairInt(t, f(3, 4).MustSubWith(f(1, 8), fraction.DenomSmallerUnit), 5, 8)
		assertExactPairInt(t, f(2, 4).MustMulWith(f(1, 2), fraction.DenomSmallerUnit), 1, 4)
		assertExactPairInt(t, f(1, 4).MustDivWith(f(1, 2), fraction.DenomSmallerUnit), 2, 4)

		_, err := f(1, 4).AddWith(f(1, 6), fraction.DenomSmallerUnit)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		_, err = f(1, 2).MulWith(f(1, 4), fraction.DenomSmallerUnit)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		_, err = f(1, 4).DivWith(f(3, 4), fraction.DenomSmallerUnit)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)

		assertMustPanicsWithError(t, func() {
			_ = f(1, 4).MustAddWith(f(1, 6), fraction.DenomSmallerUnit)
		}, fraction.ErrNotRepresentable)
	})

	t.Run("unknown_DenomMode", func(t *testing.T) {
		t.Parallel()

		unknownMode := fraction.DenomMode(99)
		_, err := f(1, 2).AddWith(f(1, 2), unknownMode)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		_, err = f(1, 2).SubWith(f(1, 2), unknownMode)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		_, err = f(1, 2).MulWith(f(1, 2), unknownMode)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		_, err = f(1, 2).DivWith(f(1, 2), unknownMode)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
	})
}

// TestBinaryArithmetic_Mode7_InDenominator verifies criterion K-18.
func TestBinaryArithmetic_Mode7_InDenominator(t *testing.T) {
	t.Parallel()

	f := func(n, d int) fraction.Fraction { return fraction.MustNew(n, d) }

	assertExactPairInt(t, f(1, 3).MustAddInDenominator(f(1, 4), 12), 7, 12)
	assertExactPairInt(t, f(1, 3).MustAddInDenominator(f(1, 4), 24), 14, 24)
	assertExactPairInt(t, f(1, 6).MustAddInDenominator(f(1, 6), 12), 4, 12)
	assertExactPairInt(t, f(1, 6).MustAddInDenominator(f(1, 6), 3), 1, 3)
	assertExactPairInt(t, f(1, 2).MustSubInDenominator(f(1, 6), 6), 2, 6)
	assertExactPairInt(t, f(1, 8).MustMulInDenominator(f(2, 3), 24), 2, 24)
	assertExactPairInt(t, f(1, 2).MustDivInDenominator(f(3, 4), 12), 8, 12)

	// Not representable in target denominator.
	_, err := f(1, 3).AddInDenominator(f(1, 4), 10)
	require.ErrorIs(t, err, fraction.ErrNotRepresentable)
	_, err = f(1, 2).SubInDenominator(f(1, 3), 10)
	require.ErrorIs(t, err, fraction.ErrNotRepresentable)
	_, err = f(1, 8).MulInDenominator(f(2, 3), 10)
	require.ErrorIs(t, err, fraction.ErrNotRepresentable)
	_, err = f(1, 2).DivInDenominator(f(3, 4), 10)
	require.ErrorIs(t, err, fraction.ErrNotRepresentable)

	// Non-positive target denominator.
	for _, d := range []int{0, -1, -12} {
		_, err := f(1, 3).AddInDenominator(f(1, 4), d)
		require.ErrorIs(t, err, fraction.ErrZeroDenominator)
		_, err = f(1, 3).SubInDenominator(f(1, 4), d)
		require.ErrorIs(t, err, fraction.ErrZeroDenominator)
		_, err = f(1, 3).MulInDenominator(f(1, 4), d)
		require.ErrorIs(t, err, fraction.ErrZeroDenominator)
		_, err = f(1, 3).DivInDenominator(f(1, 4), d)
		require.ErrorIs(t, err, fraction.ErrZeroDenominator)

		assertMustPanicsWithError(t, func() {
			_ = f(1, 3).MustAddInDenominator(f(1, 4), d)
		}, fraction.ErrZeroDenominator)
	}

	// Division by zero in DivInDenominator.
	_, err = f(1, 2).DivInDenominator(f(0, 3), 12)
	require.ErrorIs(t, err, fraction.ErrDivisionByZero)
	assertMustPanicsWithError(t, func() {
		_ = f(1, 2).MustDivInDenominator(f(0, 3), 12)
	}, fraction.ErrDivisionByZero)
}

// TestPow verifies criterion K-19.
func TestPow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		num     int
		den     int
		exp     int
		wantNum int
		wantDen int
	}{
		{name: "(3/2)^4=81/16", num: 3, den: 2, exp: 4, wantNum: 81, wantDen: 16},
		{name: "(3/2)^12=531441/4096", num: 3, den: 2, exp: 12, wantNum: 531441, wantDen: 4096},
		{name: "(3/2)^0=1/1", num: 3, den: 2, exp: 0, wantNum: 1, wantDen: 1},
		{name: "(0/3)^0=1/1", num: 0, den: 3, exp: 0, wantNum: 1, wantDen: 1},
		{name: "(0/3)^5=0/1", num: 0, den: 3, exp: 5, wantNum: 0, wantDen: 1},
		{name: "(2/3)^-2=9/4", num: 2, den: 3, exp: -2, wantNum: 9, wantDen: 4},
		{name: "(-2/3)^3=-8/27", num: -2, den: 3, exp: 3, wantNum: -8, wantDen: 27},
		{name: "(-2/3)^-3=-27/8", num: -2, den: 3, exp: -3, wantNum: -27, wantDen: 8},
		{name: "(6/12)^1=1/2", num: 6, den: 12, exp: 1, wantNum: 1, wantDen: 2},
		{name: "(6/4)^2=9/4", num: 6, den: 4, exp: 2, wantNum: 9, wantDen: 4},
		{name: "(-1/1)^MaxInt=-1/1", num: -1, den: 1, exp: math.MaxInt, wantNum: -1, wantDen: 1},
		{name: "(-1/1)^(MaxInt-1)=1/1", num: -1, den: 1, exp: math.MaxInt - 1, wantNum: 1, wantDen: 1},
		{name: "(1/1)^MinInt=1/1", num: 1, den: 1, exp: math.MinInt, wantNum: 1, wantDen: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := fraction.MustNew(tc.num, tc.den)
			got, err := f.Pow(tc.exp)
			require.NoError(t, err)
			assertExactPairInt(t, got, tc.wantNum, tc.wantDen)
			assert.True(t, got.IsReduced())
			assert.Equal(t, got, f.MustPow(tc.exp))
		})
	}

	t.Run("mercator_comma_via_Pow", func(t *testing.T) {
		t.Parallel()

		p3 := fraction.FromInt(3).MustPow(53)
		p2 := fraction.FromInt(2).MustPow(84)
		comma, err := p3.Div(p2)
		require.NoError(t, err)

		bNum, bDen, ok := comma.BigPair()
		require.True(t, ok)
		assert.Equal(t, 0, bNum.Cmp(mercatorNumBig()))
		assert.Equal(t, 0, bDen.Cmp(mercatorDenBig()))
	})

	t.Run("zero_to_negative_power_error", func(t *testing.T) {
		t.Parallel()

		for _, zp := range []rawPair{{0, 1}, {0, 3}, {0, -5}} {
			zf := fraction.MustNew(zp.num, zp.den)
			for _, exp := range []int{-1, -2, -100, math.MinInt} {
				_, err := zf.Pow(exp)
				require.ErrorIs(t, err, fraction.ErrDivisionByZero)
				assertMustPanicsWithError(t, func() {
					_ = zf.MustPow(exp)
				}, fraction.ErrDivisionByZero)
			}
		}
	})

	t.Run("overflow_and_huge_exponent_early_cutoff", func(t *testing.T) {
		t.Parallel()

		two := fraction.FromInt(2)
		// 2^254 fits into 255-bit MaxComponent; 2^255 exceeds 2^255 - 1.
		pow254, err := two.Pow(254)
		require.NoError(t, err)
		assert.True(t, pow254.IsValid())

		_, err = two.Pow(255)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = two.Pow(-255)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = two.Pow(math.MaxInt)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = two.Pow(math.MinInt)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)

		assertMustPanicsWithError(t, func() {
			_ = two.MustPow(255)
		}, fraction.ErrOutOfRange)
	})
}

// TestInvalidFraction_Part2_Arithmetic verifies criterion K-24 (part 2: unary and binary arithmetic on zero value).
func TestInvalidFraction_Part2_Arithmetic(t *testing.T) {
	t.Parallel()

	var z fraction.Fraction
	v := fraction.MustNew(1, 2)

	_, err := z.Neg()
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.Abs()
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.Reciprocal()
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.Inc()
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.Dec()
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.StepNumerator(1)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.Pow(0)
	require.ErrorIs(t, err, fraction.ErrInvalid)

	_, err = z.Add(v)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = v.Add(z)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.Sub(v)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = v.Sub(z)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.Mul(v)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = v.Mul(z)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.Div(fraction.Zero()) // Step 1 (ErrInvalid) precedes Step 2 (ErrDivisionByZero).
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = v.Div(z)
	require.ErrorIs(t, err, fraction.ErrInvalid)

	_, err = z.AddInt(1)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.SubInt(1)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.MulInt(2)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.DivInt(0) // Step 1 precedes Step 2.
	require.ErrorIs(t, err, fraction.ErrInvalid)

	_, err = z.AddWith(v, fraction.DenomUnreducedLCM)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = v.AddWith(z, fraction.DenomUnreducedLCM)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.SubWith(v, fraction.DenomUnreducedProduct)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = v.SubWith(z, fraction.DenomUnreducedProduct)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.MulWith(v, fraction.DenomSameUnit)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = v.MulWith(z, fraction.DenomSameUnit)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.DivWith(fraction.Zero(), fraction.DenomLargerUnit)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = v.DivWith(z, fraction.DenomSmallerUnit)
	require.ErrorIs(t, err, fraction.ErrInvalid)

	_, err = z.AddInDenominator(v, 0) // Step 1 precedes Step 2.
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = v.AddInDenominator(z, 0)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.SubInDenominator(v, 4)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.MulInDenominator(v, 4)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.DivInDenominator(fraction.Zero(), 0)
	require.ErrorIs(t, err, fraction.ErrInvalid)
}

// TestMaxComponentAndIntermediatePrecision verifies criterion K-25.
func TestMaxComponentAndIntermediatePrecision(t *testing.T) {
	t.Parallel()

	t.Run("compile_time_constants_and_MaxComponentBig", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 255, fraction.MaxComponentBits)
		assert.Equal(t, 100, fraction.MaxDigits)

		// Verify untyped compile-time constant MaxComponent = (1 << 255) - 1 via 64-bit word slices.
		const (
			word0 = uint64(fraction.MaxComponent & 0xffffffffffffffff)
			word1 = uint64((fraction.MaxComponent >> 64) & 0xffffffffffffffff)
			word2 = uint64((fraction.MaxComponent >> 128) & 0xffffffffffffffff)
			word3 = uint64(fraction.MaxComponent >> 192)
			word4 = uint64(fraction.MaxComponent >> 255)
		)
		assert.Equal(t, ^uint64(0), word0)
		assert.Equal(t, ^uint64(0), word1)
		assert.Equal(t, ^uint64(0), word2)
		assert.Equal(t, (uint64(1)<<63)-1, word3)
		assert.Equal(t, uint64(0), word4)

		maxBig := fraction.MaxComponentBig()
		wantMaxBig := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(1))
		assert.Equal(t, 0, maxBig.Cmp(wantMaxBig))
		assert.Equal(t, fraction.MaxComponentString, maxBig.String())

		// MaxComponent >= 2^127 - 1.
		minRequired := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(1))
		assert.GreaterOrEqual(t, maxBig.Cmp(minRequired), 0)

		// Defensive copy check on MaxComponentBig().
		maxBig.SetInt64(0)
		assert.Equal(t, 0, fraction.MaxComponentBig().Cmp(wantMaxBig))
	})

	t.Run("boundary_creation_and_out_of_range", func(t *testing.T) {
		t.Parallel()

		maxBig := fraction.MaxComponentBig()
		negMaxBig := new(big.Int).Neg(maxBig)
		oneBig := big.NewInt(1)

		// (MaxComponent, 1), (-MaxComponent, 1), (1, MaxComponent), (MaxComponent, MaxComponent) succeed.
		fMax, err := fraction.FromBigInt(maxBig, oneBig)
		require.NoError(t, err)
		bNum, bDen, ok := fMax.BigPair()
		require.True(t, ok)
		assert.Equal(t, 0, bNum.Cmp(maxBig))
		assert.Equal(t, 0, bDen.Cmp(oneBig))

		fNegMax, err := fraction.FromBigInt(negMaxBig, oneBig)
		require.NoError(t, err)
		bNum, bDen, ok = fNegMax.BigPair()
		require.True(t, ok)
		assert.Equal(t, 0, bNum.Cmp(negMaxBig))
		assert.Equal(t, 0, bDen.Cmp(oneBig))

		fOneOverMax, err := fraction.FromBigInt(oneBig, maxBig)
		require.NoError(t, err)
		bNum, bDen, ok = fOneOverMax.BigPair()
		require.True(t, ok)
		assert.Equal(t, 0, bNum.Cmp(oneBig))
		assert.Equal(t, 0, bDen.Cmp(maxBig))

		fMaxOverMax, err := fraction.FromBigInt(maxBig, maxBig)
		require.NoError(t, err)
		assertExactPairInt(t, fMaxOverMax.Reduce(), 1, 1)

		// Parse with MaxComponentString.
		parsedMax, err := fraction.Parse(fraction.MaxComponentString + "/1")
		require.NoError(t, err)
		assert.True(t, parsedMax.EqualExact(fMax))

		// MaxComponent + 1 -> ErrOutOfRange.
		overBig := new(big.Int).Add(maxBig, oneBig)
		negOverBig := new(big.Int).Neg(overBig)

		_, err = fraction.FromBigInt(overBig, oneBig)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = fraction.FromBigInt(negOverBig, oneBig)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = fraction.FromBigInt(oneBig, overBig)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = fraction.FromBigInt(oneBig, negOverBig)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = fraction.FromBigIntReduced(overBig, oneBig)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = fraction.FromBigRat(new(big.Rat).SetFrac(overBig, oneBig))
		require.ErrorIs(t, err, fraction.ErrOutOfRange)

		// Arithmetic exceeding MaxComponent -> ErrOutOfRange.
		_, err = fMax.Add(fraction.One())
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = fMax.Inc()
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = fNegMax.Sub(fraction.One())
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = fNegMax.Dec()
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = fMax.StepNumerator(1)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = fNegMax.StepNumerator(-1)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = fMax.MulInt(2)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
		_, err = fOneOverMax.DivInt(2)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
	})

	t.Run("no_intermediate_overflow_when_reduced_result_fits", func(t *testing.T) {
		t.Parallel()

		maxBig := fraction.MaxComponentBig()
		d30 := d30Big()
		twoD30 := new(big.Int).Mul(big.NewInt(2), d30)
		fourD30 := new(big.Int).Mul(big.NewInt(4), d30)

		// FromBigIntReduced(MaxComponent, MaxComponent) -> (1, 1).
		r1, err := fraction.FromBigIntReduced(maxBig, maxBig)
		require.NoError(t, err)
		assertExactPairInt(t, r1, 1, 1)

		// FromBigIntReduced(2*10^30, 4*10^30) -> (1, 2).
		r2, err := fraction.FromBigIntReduced(twoD30, fourD30)
		require.NoError(t, err)
		assertExactPairInt(t, r2, 1, 2)

		// Even inputs exceeding MaxComponent (2*10^80 / 4*10^80) reduce to (1, 2) before range check!
		d80 := bigTenPow(80)
		twoD80 := new(big.Int).Mul(big.NewInt(2), d80)
		fourD80 := new(big.Int).Mul(big.NewInt(4), d80)
		r80, err := fraction.FromBigIntReduced(twoD80, fourD80)
		require.NoError(t, err)
		assertExactPairInt(t, r80, 1, 2)

		// Add((1, 2*10^30), (1, 2*10^30)) -> (1, 10^30).
		fHalfD30 := fraction.MustFromBigInt(big.NewInt(1), twoD30)
		sumD30, err := fHalfD30.Add(fHalfD30)
		require.NoError(t, err)
		bNum, bDen, ok := sumD30.BigPair()
		require.True(t, ok)
		assert.Equal(t, 0, bNum.Cmp(big.NewInt(1)))
		assert.Equal(t, 0, bDen.Cmp(d30))

		// (3^53 / 2^84) * (2^84 / 3^53) -> (1, 1).
		mNum := mercatorNumBig()
		mDen := mercatorDenBig()
		fMerc := fraction.MustFromBigInt(mNum, mDen)
		fMercInv := fraction.MustFromBigInt(mDen, mNum)
		prodMerc, err := fMerc.Mul(fMercInv)
		require.NoError(t, err)
		assertExactPairInt(t, prodMerc, 1, 1)

		// (MaxComponent / 2) + (MaxComponent / 2) -> (MaxComponent, 1) without intermediate overflow.
		fHalfMax := fraction.MustFromBigInt(maxBig, big.NewInt(2))
		sumMax, err := fHalfMax.Add(fHalfMax)
		require.NoError(t, err)
		bNum, bDen, ok = sumMax.BigPair()
		require.True(t, ok)
		assert.Equal(t, 0, bNum.Cmp(maxBig))
		assert.Equal(t, 0, bDen.Cmp(big.NewInt(1)))

		// (MaxComponent / 1) * (1 / MaxComponent) -> (1, 1).
		fMax := fraction.MustFromBigInt(maxBig, big.NewInt(1))
		fInvMax := fraction.MustFromBigInt(big.NewInt(1), maxBig)
		prodMax, err := fMax.Mul(fInvMax)
		require.NoError(t, err)
		assertExactPairInt(t, prodMax, 1, 1)
	})
}
