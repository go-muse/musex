package fraction_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSplitAndQuoRem verifies criterion K-20 (Split*, SplitFloor*, QuoRem*, DivModFloor*).
func TestSplitAndQuoRem(t *testing.T) {
	t.Parallel()

	t.Run("Split_toward_zero", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name       string
			num        int
			den        int
			wantWhole  int
			wantRemNum int
			wantRemDen int
		}{
			{name: "14/12->1_and_2/12", num: 14, den: 12, wantWhole: 1, wantRemNum: 2, wantRemDen: 12},
			{name: "-13/12->-1_and_-1/12", num: -13, den: 12, wantWhole: -1, wantRemNum: -1, wantRemDen: 12},
			{name: "-1/12->0_and_-1/12", num: -1, den: 12, wantWhole: 0, wantRemNum: -1, wantRemDen: 12},
			{name: "6/3->2_and_0/3", num: 6, den: 3, wantWhole: 2, wantRemNum: 0, wantRemDen: 3},
			{name: "1/-2->0_and_-1/2", num: 1, den: -2, wantWhole: 0, wantRemNum: -1, wantRemDen: 2},
			{name: "-13/-12->1_and_1/12", num: -13, den: -12, wantWhole: 1, wantRemNum: 1, wantRemDen: 12},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				f := fraction.MustNew(tc.num, tc.den)
				whole, rem, err := f.Split()
				require.NoError(t, err)
				assert.Equal(t, tc.wantWhole, whole)
				assertExactPairInt(t, rem, tc.wantRemNum, tc.wantRemDen)

				mWhole, mRem := f.MustSplit()
				assert.Equal(t, whole, mWhole)
				assert.Equal(t, rem, mRem)

				bWhole, bRem, err := f.SplitBig()
				require.NoError(t, err)
				assert.Equal(t, 0, bWhole.Cmp(big.NewInt(int64(tc.wantWhole))))
				assert.Equal(t, rem, bRem)

				mbWhole, mbRem := f.MustSplitBig()
				assert.Equal(t, 0, mbWhole.Cmp(big.NewInt(int64(tc.wantWhole))))
				assert.Equal(t, rem, mbRem)

				// Defensive copy check on SplitBig.
				bWhole.SetInt64(999999)
				bWhole2, _, err := f.SplitBig()
				require.NoError(t, err)
				assert.Equal(t, 0, bWhole2.Cmp(big.NewInt(int64(tc.wantWhole))))
			})
		}
	})

	t.Run("SplitFloor_toward_minus_infinity", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name       string
			num        int
			den        int
			wantPeriod int
			wantRemNum int
			wantRemDen int
		}{
			{name: "14/12->1_and_2/12", num: 14, den: 12, wantPeriod: 1, wantRemNum: 2, wantRemDen: 12},
			{name: "-1/12->-1_and_11/12", num: -1, den: 12, wantPeriod: -1, wantRemNum: 11, wantRemDen: 12},
			{name: "-13/12->-2_and_11/12", num: -13, den: 12, wantPeriod: -2, wantRemNum: 11, wantRemDen: 12},
			{name: "-12/12->-1_and_0/12", num: -12, den: 12, wantPeriod: -1, wantRemNum: 0, wantRemDen: 12},
			{name: "0/12->0_and_0/12", num: 0, den: 12, wantPeriod: 0, wantRemNum: 0, wantRemDen: 12},
			{name: "1/-2->-1_and_1/2", num: 1, den: -2, wantPeriod: -1, wantRemNum: 1, wantRemDen: 2},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				f := fraction.MustNew(tc.num, tc.den)
				periods, rem, err := f.SplitFloor()
				require.NoError(t, err)
				assert.Equal(t, tc.wantPeriod, periods)
				assertExactPairInt(t, rem, tc.wantRemNum, tc.wantRemDen)

				mPeriods, mRem := f.MustSplitFloor()
				assert.Equal(t, periods, mPeriods)
				assert.Equal(t, rem, mRem)

				bPeriods, bRem, err := f.SplitFloorBig()
				require.NoError(t, err)
				assert.Equal(t, 0, bPeriods.Cmp(big.NewInt(int64(tc.wantPeriod))))
				assert.Equal(t, rem, bRem)

				mbPeriods, mbRem := f.MustSplitFloorBig()
				assert.Equal(t, 0, mbPeriods.Cmp(big.NewInt(int64(tc.wantPeriod))))
				assert.Equal(t, rem, mbRem)
			})
		}
	})

	t.Run("invariants_across_S", func(t *testing.T) {
		t.Parallel()

		zero := fraction.Zero()
		one := fraction.One()

		for _, p := range sampleSetS() {
			f := fraction.MustNew(p.num, p.den)
			absDen := p.den
			if absDen < 0 {
				absDen = -absDen
			}

			// Split invariants: FromInt(q) + r == f, |r| < 1, sign(r) in {0, sign(f)}, den(r) == |d|.
			q, r, err := f.Split()
			require.NoError(t, err)
			assert.True(t, fraction.FromInt(q).MustAdd(r).Equal(f))
			assert.True(t, r.MustAbs().Less(one))
			if !r.IsZero() {
				assert.Equal(t, f.Sign(), r.Sign())
			}
			_, rDen, ok := r.Pair()
			require.True(t, ok)
			assert.Equal(t, absDen, rDen)

			// SplitFloor invariants: FromInt(qf) + rf == f, 0 <= rf < 1, den(rf) == |d|.
			qf, rf, err := f.SplitFloor()
			require.NoError(t, err)
			assert.True(t, fraction.FromInt(qf).MustAdd(rf).Equal(f))
			assert.True(t, rf.GreaterOrEqual(zero))
			assert.True(t, rf.Less(one))
			_, rfDen, ok := rf.Pair()
			require.True(t, ok)
			assert.Equal(t, absDen, rfDen)
		}
	})

	t.Run("large_quotient_outside_int64", func(t *testing.T) {
		t.Parallel()

		mNum := mercatorNumBig() // 3^53 (odd, > MaxInt)
		negMNum := new(big.Int).Neg(mNum)
		f := fraction.MustFromBigInt(negMNum, big.NewInt(2))

		_, _, err := f.Split()
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		assertMustPanicsWithError(t, func() {
			_, _ = f.MustSplit()
		}, fraction.ErrNotRepresentable)

		_, _, err = f.SplitFloor()
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		assertMustPanicsWithError(t, func() {
			_, _ = f.MustSplitFloor()
		}, fraction.ErrNotRepresentable)

		// SplitBig and SplitFloorBig succeed.
		bWhole, bRem, err := f.SplitBig()
		require.NoError(t, err)
		wantWhole := new(big.Int).Quo(negMNum, big.NewInt(2))
		assert.Equal(t, 0, bWhole.Cmp(wantWhole))
		assertExactPairInt(t, bRem, -1, 2)

		bfPeriods, bfRem, err := f.SplitFloorBig()
		require.NoError(t, err)
		wantFloor := new(big.Int).Sub(wantWhole, big.NewInt(1))
		assert.Equal(t, 0, bfPeriods.Cmp(wantFloor))
		assertExactPairInt(t, bfRem, 1, 2)
	})

	t.Run("QuoRem_and_DivModFloor", func(t *testing.T) {
		t.Parallel()

		f := func(n, d int) fraction.Fraction { return fraction.MustNew(n, d) }

		// QuoRem((-7, 4), (1, 2)) -> (-3, (-1, 4)).
		q, rem, err := f(-7, 4).QuoRem(f(1, 2))
		require.NoError(t, err)
		assert.Equal(t, -3, q)
		assertExactPairInt(t, rem, -1, 4)
		mq, mrem := f(-7, 4).MustQuoRem(f(1, 2))
		assert.Equal(t, q, mq)
		assert.Equal(t, rem, mrem)

		bq, brem, err := f(-7, 4).QuoRemBig(f(1, 2))
		require.NoError(t, err)
		assert.Equal(t, 0, bq.Cmp(big.NewInt(-3)))
		assert.Equal(t, rem, brem)
		mbq, mbrem := f(-7, 4).MustQuoRemBig(f(1, 2))
		assert.Equal(t, 0, mbq.Cmp(big.NewInt(-3)))
		assert.Equal(t, rem, mbrem)

		// QuoRem allows negative step: (7, 8) / (-3, 4) -> q = -1, rem = 1/8.
		qNeg, remNeg, err := f(7, 8).QuoRem(f(-3, 4))
		require.NoError(t, err)
		assert.Equal(t, -1, qNeg)
		assertExactPairInt(t, remNeg, 1, 8)

		// DivModFloor((-1, 12), (1, 4)) -> (-1, (1, 6)).
		pf, remf, err := f(-1, 12).DivModFloor(f(1, 4))
		require.NoError(t, err)
		assert.Equal(t, -1, pf)
		assertExactPairInt(t, remf, 1, 6)
		mpf, mremf := f(-1, 12).MustDivModFloor(f(1, 4))
		assert.Equal(t, pf, mpf)
		assert.Equal(t, remf, mremf)

		bpf, bremf, err := f(-1, 12).DivModFloorBig(f(1, 4))
		require.NoError(t, err)
		assert.Equal(t, 0, bpf.Cmp(big.NewInt(-1)))
		assert.Equal(t, remf, bremf)
		mbpf, mbremf := f(-1, 12).MustDivModFloorBig(f(1, 4))
		assert.Equal(t, 0, mbpf.Cmp(big.NewInt(-1)))
		assert.Equal(t, remf, mbremf)

		// DivModFloor((7, 8), (3, 4)) -> (1, (1, 8)).
		pf2, remf2, err := f(7, 8).DivModFloor(f(3, 4))
		require.NoError(t, err)
		assert.Equal(t, 1, pf2)
		assertExactPairInt(t, remf2, 1, 8)

		// DivModFloor with zero step -> ErrDivisionByZero.
		for _, zStep := range []fraction.Fraction{fraction.Zero(), f(0, 3)} {
			_, _, err := f(1, 2).QuoRem(zStep)
			require.ErrorIs(t, err, fraction.ErrDivisionByZero)
			_, _, err = f(1, 2).QuoRemBig(zStep)
			require.ErrorIs(t, err, fraction.ErrDivisionByZero)
			_, _, err = f(1, 2).DivModFloor(zStep)
			require.ErrorIs(t, err, fraction.ErrDivisionByZero)
			_, _, err = f(1, 2).DivModFloorBig(zStep)
			require.ErrorIs(t, err, fraction.ErrDivisionByZero)

			assertMustPanicsWithError(t, func() {
				_, _ = f(1, 2).MustQuoRem(zStep)
			}, fraction.ErrDivisionByZero)
			assertMustPanicsWithError(t, func() {
				_, _ = f(1, 2).MustDivModFloor(zStep)
			}, fraction.ErrDivisionByZero)
		}

		// DivModFloor with negative step -> ErrZeroDenominator.
		for _, negStep := range []fraction.Fraction{f(-1, 4), f(1, -4)} {
			_, _, err := f(1, 2).DivModFloor(negStep)
			require.ErrorIs(t, err, fraction.ErrZeroDenominator)
			_, _, err = f(1, 2).DivModFloorBig(negStep)
			require.ErrorIs(t, err, fraction.ErrZeroDenominator)
			assertMustPanicsWithError(t, func() {
				_, _ = f(1, 2).MustDivModFloor(negStep)
			}, fraction.ErrZeroDenominator)
			assertMustPanicsWithError(t, func() {
				_, _ = f(1, 2).MustDivModFloorBig(negStep)
			}, fraction.ErrZeroDenominator)
		}

		// Quotient outside int64 -> ErrNotRepresentable in narrow QuoRem / DivModFloor, success in *Big.
		large := fraction.MustFromBigInt(mercatorNumBig(), big.NewInt(1))
		_, _, err = large.QuoRem(f(1, 2))
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		_, _, err = large.DivModFloor(f(1, 2))
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)

		bigQuo, remLarge, err := large.QuoRemBig(f(1, 2))
		require.NoError(t, err)
		assert.Equal(t, 0, bigQuo.Cmp(new(big.Int).Mul(mercatorNumBig(), big.NewInt(2))))
		assertExactPairInt(t, remLarge, 0, 1)
	})
}

// TestDivisibilityAndProgression verifies criterion K-21 (ScaleIntExact, IsMultipleOf, InProgression).
func TestDivisibilityAndProgression(t *testing.T) {
	t.Parallel()

	f := func(n, d int) fraction.Fraction { return fraction.MustNew(n, d) }

	t.Run("ScaleIntExact", func(t *testing.T) {
		t.Parallel()

		// 1/2 * 2 = 1 (exact).
		prod, exact, err := f(1, 2).ScaleIntExact(2)
		require.NoError(t, err)
		assert.True(t, exact)
		assert.Equal(t, 1, prod)
		mProd, mExact := f(1, 2).MustScaleIntExact(2)
		assert.Equal(t, 1, mProd)
		assert.True(t, mExact)

		// 1/2 * 1 = 1/2 (not integer -> 0, false, nil).
		prod, exact, err = f(1, 2).ScaleIntExact(1)
		require.NoError(t, err)
		assert.False(t, exact)
		assert.Equal(t, 0, prod)
		mProd, mExact = f(1, 2).MustScaleIntExact(1)
		assert.Equal(t, 0, mProd)
		assert.False(t, mExact)

		// -3/4 * 8 = -6 (exact).
		prod, exact, err = f(-3, 4).ScaleIntExact(8)
		require.NoError(t, err)
		assert.True(t, exact)
		assert.Equal(t, -6, prod)

		// Unreduced 6/12 * 2 = 1 (exact because reduced denominator is 2).
		prod, exact, err = f(6, 12).ScaleIntExact(2)
		require.NoError(t, err)
		assert.True(t, exact)
		assert.Equal(t, 1, prod)

		// Product is an integer but overflows 64-bit int -> ErrNotRepresentable.
		maxIntFrac := fraction.FromInt(math.MaxInt)
		_, _, err = maxIntFrac.ScaleIntExact(2)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		assertMustPanicsWithError(t, func() {
			_, _ = maxIntFrac.MustScaleIntExact(2)
		}, fraction.ErrNotRepresentable)
	})

	t.Run("IsMultipleOf", func(t *testing.T) {
		t.Parallel()

		mult, err := f(3, 4).IsMultipleOf(f(1, 4))
		require.NoError(t, err)
		assert.True(t, mult)
		assert.True(t, f(3, 4).MustIsMultipleOf(f(1, 4)))

		mult, err = f(-6, 12).IsMultipleOf(f(1, 6))
		require.NoError(t, err)
		assert.True(t, mult)

		mult, err = f(0, 3).IsMultipleOf(f(2, 5))
		require.NoError(t, err)
		assert.True(t, mult)

		mult, err = f(1, 2).IsMultipleOf(f(1, 3))
		require.NoError(t, err)
		assert.False(t, mult)
		assert.False(t, f(1, 2).MustIsMultipleOf(f(1, 3)))

		_, err = f(1, 2).IsMultipleOf(fraction.Zero())
		require.ErrorIs(t, err, fraction.ErrDivisionByZero)
		_, err = f(1, 2).IsMultipleOf(f(0, 3))
		require.ErrorIs(t, err, fraction.ErrDivisionByZero)
		assertMustPanicsWithError(t, func() {
			_ = f(1, 2).MustIsMultipleOf(fraction.Zero())
		}, fraction.ErrDivisionByZero)
	})

	t.Run("InProgression", func(t *testing.T) {
		t.Parallel()

		start := f(1, 4)
		step := f(1, 2)

		// a = 3/4: (3/4 - 1/4) / (1/2) = 1.
		steps, ok, err := f(3, 4).InProgression(start, step)
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, 1, steps)
		mSteps, mOk := f(3, 4).MustInProgression(start, step)
		assert.Equal(t, 1, mSteps)
		assert.True(t, mOk)

		// a = -1/4: (-1/4 - 1/4) / (1/2) = -1.
		steps, ok, err = f(-1, 4).InProgression(start, step)
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, -1, steps)

		// a = 1/4: 0 steps.
		steps, ok, err = f(1, 4).InProgression(start, step)
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, 0, steps)

		// a = 1/2: not in progression -> (0, false, nil).
		steps, ok, err = f(1, 2).InProgression(start, step)
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, 0, steps)

		// a = 1/3: not in progression -> (0, false, nil).
		steps, ok, err = f(1, 3).InProgression(start, step)
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, 0, steps)

		// step == 0 -> ErrDivisionByZero.
		_, _, err = f(3, 4).InProgression(start, fraction.Zero())
		require.ErrorIs(t, err, fraction.ErrDivisionByZero)
		_, _, err = f(3, 4).InProgression(start, f(0, 5))
		require.ErrorIs(t, err, fraction.ErrDivisionByZero)
		assertMustPanicsWithError(t, func() {
			_, _ = f(3, 4).MustInProgression(start, fraction.Zero())
		}, fraction.ErrDivisionByZero)

		// step < 0 -> ErrZeroDenominator.
		_, _, err = f(3, 4).InProgression(start, f(-1, 2))
		require.ErrorIs(t, err, fraction.ErrZeroDenominator)
		assertMustPanicsWithError(t, func() {
			_, _ = f(3, 4).MustInProgression(start, f(-1, 2))
		}, fraction.ErrZeroDenominator)

		// Integer steps count exceeding int -> ErrNotRepresentable.
		large := fraction.MustFromBigInt(mercatorNumBig(), big.NewInt(1))
		_, _, err = large.InProgression(fraction.Zero(), fraction.One())
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
	})
}

// TestInProgression_IntermediateBeyondMaxComponent verifies that InProgression
// is decided only by the integer step count k = (f - start) / step, even when
// the difference f - start (or its numerator) exceeds MaxComponent. Below,
// M = MaxComponent = 2^255 - 1, which is odd, divisible by 7, and not
// divisible by 2, 3, or 5.
func TestInProgression_IntermediateBeyondMaxComponent(t *testing.T) {
	t.Parallel()

	// frac returns sign*M/den with the exact stored pair.
	frac := func(sign int64, den int64) fraction.Fraction {
		num := new(big.Int).Mul(fraction.MaxComponentBig(), big.NewInt(sign))
		return fraction.MustFromBigInt(num, big.NewInt(den))
	}
	mMinus1 := fraction.MustFromBigInt(new(big.Int).Sub(fraction.MaxComponentBig(), big.NewInt(1)), big.NewInt(1))

	cases := []struct {
		name      string
		f         fraction.Fraction
		start     fraction.Fraction
		step      fraction.Fraction
		wantSteps int
		wantOK    bool
		wantErr   error
	}{
		// f - start = 2M, 2M / M = 2.
		{name: "positive_count_difference_2M", f: frac(1, 1), start: frac(-1, 1), step: frac(1, 1), wantSteps: 2, wantOK: true},
		// f - start = -2M, -2M / M = -2.
		{name: "negative_count_difference_-2M", f: frac(-1, 1), start: frac(1, 1), step: frac(1, 1), wantSteps: -2, wantOK: true},
		// f - start = M/2 + M/3 = 5M/6 (numerator 5M > M), step M/6, k = 5.
		{name: "fractional_operands_numerator_5M_positive", f: frac(1, 2), start: frac(-1, 3), step: frac(1, 6), wantSteps: 5, wantOK: true},
		// f - start = -M/2 - M/3 = -5M/6, step M/6, k = -5.
		{name: "fractional_operands_numerator_5M_negative", f: frac(-1, 2), start: frac(1, 3), step: frac(1, 6), wantSteps: -5, wantOK: true},
		// Control: equal operands at the limit give 0 steps.
		{name: "equal_operands_at_limit", f: frac(1, 1), start: frac(1, 1), step: frac(1, 1), wantSteps: 0, wantOK: true},
		// 2M / 4 = M/2, not an integer because M is odd.
		{name: "not_integer_2M_over_4", f: frac(1, 1), start: frac(-1, 1), step: fraction.MustNew(4, 1), wantOK: false},
		// 2M / (M-1) = 2 + 2/(M-1), not an integer because M-1 > 2.
		{name: "not_integer_2M_over_M_minus_1", f: frac(1, 1), start: frac(-1, 1), step: mMinus1, wantOK: false},
		// (5M/6) / (M/5) = 25/6, not an integer.
		{name: "not_integer_fractional_operands", f: frac(1, 2), start: frac(-1, 3), step: frac(1, 5), wantOK: false},
		// k = 2M does not fit in int.
		{name: "count_beyond_int_positive", f: frac(1, 1), start: frac(-1, 1), step: fraction.One(), wantErr: fraction.ErrNotRepresentable},
		// k = -2M does not fit in int.
		{name: "count_beyond_int_negative", f: frac(-1, 1), start: frac(1, 1), step: fraction.One(), wantErr: fraction.ErrNotRepresentable},
		// k = (5M/6) / (1/6) = 5M does not fit in int.
		{name: "count_beyond_int_fractional_operands", f: frac(1, 2), start: frac(-1, 3), step: fraction.MustNew(1, 6), wantErr: fraction.ErrNotRepresentable},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			steps, ok, err := tc.f.InProgression(tc.start, tc.step)

			if tc.wantErr != nil {
				assertErrorMatchesOnly(t, err, tc.wantErr)
				assert.False(t, ok)
				assert.Equal(t, 0, steps)

				panicErr := recoverPanicError(t, func() { _, _ = tc.f.MustInProgression(tc.start, tc.step) })
				assertErrorMatchesOnly(t, panicErr, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.NotErrorIs(t, err, fraction.ErrOutOfRange)
			assert.Equal(t, tc.wantOK, ok)
			assert.Equal(t, tc.wantSteps, steps)

			mSteps, mOK := tc.f.MustInProgression(tc.start, tc.step)
			assert.Equal(t, tc.wantSteps, mSteps)
			assert.Equal(t, tc.wantOK, mOK)
		})
	}
}

// TestExplicitRoundingAndFloat64 verifies criterion K-22 (RoundToInt, RoundToBigInt, RoundToDenominator, RoundToBigDenominator, Float64).
func TestExplicitRoundingAndFloat64(t *testing.T) {
	t.Parallel()

	t.Run("RoundToInt_and_RoundToBigInt_matrix", func(t *testing.T) {
		t.Parallel()

		inputs := []rawPair{
			{1, 2},
			{-1, 2},
			{3, 2},
			{-3, 2},
			{5, 2},
			{-5, 2},
			{7, 3},
			{-7, 3},
		}

		modeExpected := []struct {
			name string
			mode fraction.RoundingMode
			want []int
		}{
			{
				name: "RoundHalfAwayFromZero",
				mode: fraction.RoundHalfAwayFromZero,
				want: []int{1, -1, 2, -2, 3, -3, 2, -2},
			},
			{
				name: "RoundHalfToEven",
				mode: fraction.RoundHalfToEven,
				want: []int{0, 0, 2, -2, 2, -2, 2, -2},
			},
			{
				name: "RoundFloor",
				mode: fraction.RoundFloor,
				want: []int{0, -1, 1, -2, 2, -3, 2, -3},
			},
			{
				name: "RoundCeil",
				mode: fraction.RoundCeil,
				want: []int{1, 0, 2, -1, 3, -2, 3, -2},
			},
			{
				name: "RoundTrunc",
				mode: fraction.RoundTrunc,
				want: []int{0, 0, 1, -1, 2, -2, 2, -2},
			},
		}

		for _, mc := range modeExpected {
			t.Run(mc.name, func(t *testing.T) {
				t.Parallel()

				for i, in := range inputs {
					f := fraction.MustNew(in.num, in.den)
					got, err := f.RoundToInt(mc.mode)
					require.NoError(t, err, "%s on %v", mc.name, in)
					assert.Equal(t, mc.want[i], got, "%s on %v", mc.name, in)
					assert.Equal(t, mc.want[i], f.MustRoundToInt(mc.mode))

					gotBig, err := f.RoundToBigInt(mc.mode)
					require.NoError(t, err)
					assert.Equal(t, 0, gotBig.Cmp(big.NewInt(int64(mc.want[i]))))
					assert.Equal(t, 0, f.MustRoundToBigInt(mc.mode).Cmp(big.NewInt(int64(mc.want[i]))))
				}
			})
		}

		// Large fraction outside int64: RoundToInt returns ErrNotRepresentable; RoundToBigInt succeeds.
		large := fraction.MustFromBigInt(mercatorNumBig(), big.NewInt(1))
		_, err := large.RoundToInt(fraction.RoundHalfAwayFromZero)
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		gotBig, err := large.RoundToBigInt(fraction.RoundHalfAwayFromZero)
		require.NoError(t, err)
		assert.Equal(t, 0, gotBig.Cmp(mercatorNumBig()))

		// Unknown RoundingMode -> ErrNotRepresentable.
		_, err = fraction.One().RoundToInt(fraction.RoundingMode(99))
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
		_, err = fraction.One().RoundToBigInt(fraction.RoundingMode(99))
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)
	})

	t.Run("RoundToDenominator_and_RoundToBigDenominator", func(t *testing.T) {
		t.Parallel()

		oneThird := fraction.MustNew(1, 3)
		for _, mode := range []fraction.RoundingMode{
			fraction.RoundHalfAwayFromZero,
			fraction.RoundHalfToEven,
			fraction.RoundFloor,
			fraction.RoundTrunc,
		} {
			got, err := oneThird.RoundToDenominator(10, mode)
			require.NoError(t, err)
			assertExactPairInt(t, got, 3, 10)
			assert.Equal(t, got, oneThird.MustRoundToDenominator(10, mode))

			bgot, err := oneThird.RoundToBigDenominator(big.NewInt(10), mode)
			require.NoError(t, err)
			assert.Equal(t, got, bgot)
			assert.Equal(t, got, oneThird.MustRoundToBigDenominator(big.NewInt(10), mode))
		}

		gotCeil, err := oneThird.RoundToDenominator(10, fraction.RoundCeil)
		require.NoError(t, err)
		assertExactPairInt(t, gotCeil, 4, 10)

		oneSixth := fraction.MustNew(1, 6)
		gotSixth, err := oneSixth.RoundToDenominator(10, fraction.RoundHalfAwayFromZero)
		require.NoError(t, err)
		assertExactPairInt(t, gotSixth, 2, 10)
		assertExactPairInt(t, gotSixth.Reduce(), 1, 5)

		// Non-positive denominator -> ErrZeroDenominator (even if mode is invalid: Step 2 before Step 3).
		for _, d := range []int{0, -1, -10} {
			_, err := oneThird.RoundToDenominator(d, fraction.RoundHalfAwayFromZero)
			require.ErrorIs(t, err, fraction.ErrZeroDenominator)
			_, err = oneThird.RoundToDenominator(d, fraction.RoundingMode(99))
			require.ErrorIs(t, err, fraction.ErrZeroDenominator)
			_, err = oneThird.RoundToBigDenominator(big.NewInt(int64(d)), fraction.RoundHalfAwayFromZero)
			require.ErrorIs(t, err, fraction.ErrZeroDenominator)
		}
		_, err = oneThird.RoundToBigDenominator(nil, fraction.RoundHalfAwayFromZero)
		require.ErrorIs(t, err, fraction.ErrZeroDenominator)

		// Unknown RoundingMode with valid D > 0 -> ErrNotRepresentable.
		_, err = oneThird.RoundToDenominator(10, fraction.RoundingMode(99))
		require.ErrorIs(t, err, fraction.ErrNotRepresentable)

		// Overflow in RoundToDenominator: (MaxComponent, 1) rounded to denominator 2 -> ErrOutOfRange.
		fMax := fraction.MustFromBigInt(fraction.MaxComponentBig(), big.NewInt(1))
		_, err = fMax.RoundToDenominator(2, fraction.RoundHalfAwayFromZero)
		require.ErrorIs(t, err, fraction.ErrOutOfRange)
	})

	t.Run("Float64", func(t *testing.T) {
		t.Parallel()

		val, exact := fraction.MustNew(1, 2).Float64()
		assert.True(t, exact)
		assert.Equal(t, 0.5, val)

		val, exact = fraction.MustNew(-1, 4).Float64()
		assert.True(t, exact)
		assert.Equal(t, -0.25, val)

		val, exact = fraction.MustNew(0, 3).Float64()
		assert.True(t, exact)
		assert.Equal(t, 0.0, val)

		val, exact = fraction.MustNew(1, 3).Float64()
		assert.False(t, exact)
		assert.InDelta(t, 1.0/3.0, val, 1e-16)

		var z fraction.Fraction
		val, exact = z.Float64()
		assert.False(t, exact)
		assert.True(t, math.IsNaN(val))
	})
}

// TestInvalidFraction_Part3_DivisionAndRounding verifies criterion K-24 (part 3: division, divisibility, and rounding on zero value).
func TestInvalidFraction_Part3_DivisionAndRounding(t *testing.T) {
	t.Parallel()

	var z fraction.Fraction
	v := fraction.MustNew(1, 2)

	_, _, err := z.Split()
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, _, err = z.SplitBig()
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, _, err = z.SplitFloor()
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, _, err = z.SplitFloorBig()
	require.ErrorIs(t, err, fraction.ErrInvalid)

	_, _, err = z.QuoRem(fraction.Zero()) // Step 1 precedes Step 2.
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, _, err = v.QuoRem(z)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, _, err = z.QuoRemBig(v)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, _, err = v.QuoRemBig(z)
	require.ErrorIs(t, err, fraction.ErrInvalid)

	_, _, err = z.DivModFloor(fraction.Zero())
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, _, err = v.DivModFloor(z)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, _, err = z.DivModFloorBig(v)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, _, err = v.DivModFloorBig(z)
	require.ErrorIs(t, err, fraction.ErrInvalid)

	_, _, err = z.ScaleIntExact(2)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.IsMultipleOf(fraction.Zero())
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = v.IsMultipleOf(z)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, _, err = z.InProgression(v, fraction.Zero())
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, _, err = v.InProgression(z, v)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, _, err = v.InProgression(v, z)
	require.ErrorIs(t, err, fraction.ErrInvalid)

	_, err = z.RoundToInt(fraction.RoundHalfAwayFromZero)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.RoundToBigInt(fraction.RoundHalfAwayFromZero)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.RoundToDenominator(0, fraction.RoundHalfAwayFromZero)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.RoundToBigDenominator(nil, fraction.RoundHalfAwayFromZero)
	require.ErrorIs(t, err, fraction.ErrInvalid)
}
