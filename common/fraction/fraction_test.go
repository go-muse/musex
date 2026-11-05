package fraction_test

import (
	"errors"
	"math"
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rawPair struct {
	num int
	den int
}

// sampleSetS returns the 19 canonical validation pairs S from the requirements specification.
func sampleSetS() []rawPair {
	return []rawPair{
		{0, 1},
		{0, 3},
		{0, -5},
		{1, 1},
		{-1, 1},
		{1, 2},
		{6, 12},
		{1, -2},
		{-2, -4},
		{1, 3},
		{2, 3},
		{-3, 4},
		{3, 2},
		{6, 4},
		{14, 12},
		{-13, 12},
		{-1, 12},
		{6, 3},
		{-4, 2},
	}
}

func bigTenPow(exp int64) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(exp), nil)
}

func d30Big() *big.Int {
	return bigTenPow(30)
}

func mercatorNumBig() *big.Int {
	return new(big.Int).Exp(big.NewInt(3), big.NewInt(53), nil)
}

func mercatorDenBig() *big.Int {
	return new(big.Int).Exp(big.NewInt(2), big.NewInt(84), nil)
}

func oracleReducedPair(num, den int) (int, int) {
	if num == 0 {
		return 0, 1
	}
	r := big.NewRat(int64(num), int64(den))
	return int(r.Num().Int64()), int(r.Denom().Int64())
}

func assertExactPairInt(t *testing.T, f fraction.Fraction, wantNum, wantDen int) {
	t.Helper()
	require.True(t, f.IsValid(), "expected valid fraction for (%d, %d)", wantNum, wantDen)

	gotNum, ok := f.Numerator()
	require.True(t, ok)
	assert.Equal(t, wantNum, gotNum)

	gotDen, ok := f.Denominator()
	require.True(t, ok)
	assert.Equal(t, wantDen, gotDen)

	pNum, pDen, ok := f.Pair()
	require.True(t, ok)
	assert.Equal(t, wantNum, pNum)
	assert.Equal(t, wantDen, pDen)

	bNum, ok := f.BigNumerator()
	require.True(t, ok)
	assert.Equal(t, 0, bNum.Cmp(big.NewInt(int64(wantNum))))

	bDen, ok := f.BigDenominator()
	require.True(t, ok)
	assert.Equal(t, 0, bDen.Cmp(big.NewInt(int64(wantDen))))

	bpNum, bpDen, ok := f.BigPair()
	require.True(t, ok)
	assert.Equal(t, 0, bpNum.Cmp(big.NewInt(int64(wantNum))))
	assert.Equal(t, 0, bpDen.Cmp(big.NewInt(int64(wantDen))))
}

func assertReducedPairInt(t *testing.T, f fraction.Fraction, wantNum, wantDen int) {
	t.Helper()
	require.True(t, f.IsValid())

	gotNum, ok := f.ReducedNumerator()
	require.True(t, ok)
	assert.Equal(t, wantNum, gotNum)

	gotDen, ok := f.ReducedDenominator()
	require.True(t, ok)
	assert.Equal(t, wantDen, gotDen)

	pNum, pDen, ok := f.ReducedPair()
	require.True(t, ok)
	assert.Equal(t, wantNum, pNum)
	assert.Equal(t, wantDen, pDen)

	bNum, ok := f.ReducedBigNumerator()
	require.True(t, ok)
	assert.Equal(t, 0, bNum.Cmp(big.NewInt(int64(wantNum))))

	bDen, ok := f.ReducedBigDenominator()
	require.True(t, ok)
	assert.Equal(t, 0, bDen.Cmp(big.NewInt(int64(wantDen))))

	bpNum, bpDen, ok := f.ReducedBigPair()
	require.True(t, ok)
	assert.Equal(t, 0, bpNum.Cmp(big.NewInt(int64(wantNum))))
	assert.Equal(t, 0, bpDen.Cmp(big.NewInt(int64(wantDen))))
}

func assertMustPanicsWithError(t *testing.T, fn func(), wantErr error) {
	t.Helper()
	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		fn()
	}()
	require.NotNil(t, recovered, "expected panic with %v", wantErr)
	err, ok := recovered.(error)
	require.True(t, ok, "expected recovered value to implement error, got %T", recovered)
	assert.ErrorIs(t, err, wantErr)
}

// TestReadyValues verifies criterion K-1 (Zero, One, MinusOne).
func TestReadyValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fn       func() fraction.Fraction
		wantNum  int
		wantDen  int
		wantSign int
		isZero   bool
		isOne    bool
		isProper bool
	}{
		{
			name:     "Zero",
			fn:       fraction.Zero,
			wantNum:  0,
			wantDen:  1,
			wantSign: 0,
			isZero:   true,
			isOne:    false,
			isProper: true,
		},
		{
			name:     "One",
			fn:       fraction.One,
			wantNum:  1,
			wantDen:  1,
			wantSign: 1,
			isZero:   false,
			isOne:    true,
			isProper: false,
		},
		{
			name:     "MinusOne",
			fn:       fraction.MinusOne,
			wantNum:  -1,
			wantDen:  1,
			wantSign: -1,
			isZero:   false,
			isOne:    false,
			isProper: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := tc.fn()
			assert.True(t, f.IsValid())
			assert.True(t, f.IsReduced())
			assert.True(t, f.IsInteger())
			assert.Equal(t, tc.isZero, f.IsZero())
			assert.Equal(t, tc.isOne, f.IsOne())
			assert.Equal(t, tc.isProper, f.IsProper())
			assert.Equal(t, tc.wantSign, f.Sign())

			assertExactPairInt(t, f, tc.wantNum, tc.wantDen)
			assertReducedPairInt(t, f, tc.wantNum, tc.wantDen)

			intVal, ok := f.Int()
			require.True(t, ok)
			assert.Equal(t, tc.wantNum, intVal)

			fromI := fraction.FromInt(tc.wantNum)
			assert.True(t, f.Equal(fromI))
			assert.True(t, f.EqualExact(fromI))
			assert.Equal(t, fromI, f)
		})
	}
}

// TestNew_PreservesPair verifies criterion K-2 (New and MustNew preserve exact pair and reject d == 0).
func TestNew_PreservesPair(t *testing.T) {
	t.Parallel()

	explicitCases := []struct {
		name string
		num  int
		den  int
	}{
		{name: "6/12", num: 6, den: 12},
		{name: "1/-2", num: 1, den: -2},
		{name: "-2/-4", num: -2, den: -4},
		{name: "0/3", num: 0, den: 3},
		{name: "0/-5", num: 0, den: -5},
		{name: "4/2", num: 4, den: 2},
		{name: "MaxInt/MinInt", num: math.MaxInt, den: math.MinInt},
		{name: "MinInt/MaxInt", num: math.MinInt, den: math.MaxInt},
		{name: "MinInt/-1", num: math.MinInt, den: -1},
	}

	for _, tc := range explicitCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := fraction.New(tc.num, tc.den)
			require.NoError(t, err)
			assertExactPairInt(t, f, tc.num, tc.den)

			mf := fraction.MustNew(tc.num, tc.den)
			assert.True(t, f.EqualExact(mf))
			assert.Equal(t, f, mf)
		})
	}

	t.Run("cartesian_-12_to_12", func(t *testing.T) {
		t.Parallel()

		for n := -12; n <= 12; n++ {
			for d := -12; d <= 12; d++ {
				if d == 0 {
					continue
				}
				f, err := fraction.New(n, d)
				require.NoError(t, err, "New(%d, %d)", n, d)
				assertExactPairInt(t, f, n, d)

				mf := fraction.MustNew(n, d)
				assert.Equal(t, f, mf)
				assert.True(t, f.EqualExact(mf))
			}
		}
	})

	t.Run("zero_denominator_error", func(t *testing.T) {
		t.Parallel()

		for _, n := range []int{-12, -1, 0, 1, 6, math.MinInt, math.MaxInt} {
			f, err := fraction.New(n, 0)
			require.ErrorIs(t, err, fraction.ErrZeroDenominator)
			assert.False(t, f.IsValid())

			assertMustPanicsWithError(t, func() {
				_ = fraction.MustNew(n, 0)
			}, fraction.ErrZeroDenominator)
		}
	})
}

// TestNewReduced verifies criterion K-3 (NewReduced and MustNewReduced).
func TestNewReduced(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		num     int
		den     int
		wantNum int
		wantDen int
	}{
		{name: "6/12->1/2", num: 6, den: 12, wantNum: 1, wantDen: 2},
		{name: "1/-2->-1/2", num: 1, den: -2, wantNum: -1, wantDen: 2},
		{name: "-2/-4->1/2", num: -2, den: -4, wantNum: 1, wantDen: 2},
		{name: "0/3->0/1", num: 0, den: 3, wantNum: 0, wantDen: 1},
		{name: "0/-5->0/1", num: 0, den: -5, wantNum: 0, wantDen: 1},
		{name: "4/2->2/1", num: 4, den: 2, wantNum: 2, wantDen: 1},
		{name: "MinInt/MinInt->1/1", num: math.MinInt, den: math.MinInt, wantNum: 1, wantDen: 1},
		{name: "MinInt/2", num: math.MinInt, den: 2, wantNum: math.MinInt / 2, wantDen: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := fraction.NewReduced(tc.num, tc.den)
			require.NoError(t, err)
			assert.True(t, f.IsValid())
			assert.True(t, f.IsReduced())
			assertExactPairInt(t, f, tc.wantNum, tc.wantDen)
			assertReducedPairInt(t, f, tc.wantNum, tc.wantDen)

			mf := fraction.MustNewReduced(tc.num, tc.den)
			assert.Equal(t, f, mf)
			assert.True(t, f.EqualExact(mf))

			raw := fraction.MustNew(tc.num, tc.den)
			assert.True(t, f.EqualExact(raw.Reduce()))
			assert.True(t, f.Equal(raw))
		})
	}

	t.Run("MinInt_over_minus1_exceeds_int64_but_fits_255bit", func(t *testing.T) {
		t.Parallel()

		// -math.MinInt = +2^63, which fits into the 255-bit component range.
		f, err := fraction.NewReduced(math.MinInt, -1)
		require.NoError(t, err)
		assert.True(t, f.IsValid())
		assert.True(t, f.IsReduced())
		assert.True(t, f.IsInteger())

		// +2^63 does not fit into signed 64-bit int.
		_, ok := f.Numerator()
		assert.False(t, ok)
		_, _, ok = f.Pair()
		assert.False(t, ok)
		_, ok = f.Int()
		assert.False(t, ok)

		wantBig := new(big.Int).Lsh(big.NewInt(1), 63)
		bNum, bDen, ok := f.BigPair()
		require.True(t, ok)
		assert.Equal(t, 0, bNum.Cmp(wantBig))
		assert.Equal(t, 0, bDen.Cmp(big.NewInt(1)))
	})

	t.Run("all_S_match_Reduce", func(t *testing.T) {
		t.Parallel()

		for _, p := range sampleSetS() {
			r, err := fraction.NewReduced(p.num, p.den)
			require.NoError(t, err)
			raw := fraction.MustNew(p.num, p.den)
			assert.True(t, r.IsReduced())
			assert.True(t, r.EqualExact(raw.Reduce()))
			assert.Equal(t, raw.Reduce(), r)

			wantN, wantD := oracleReducedPair(p.num, p.den)
			assertExactPairInt(t, r, wantN, wantD)
		}
	})

	t.Run("zero_denominator_error", func(t *testing.T) {
		t.Parallel()

		f, err := fraction.NewReduced(1, 0)
		require.ErrorIs(t, err, fraction.ErrZeroDenominator)
		assert.False(t, f.IsValid())

		assertMustPanicsWithError(t, func() {
			_ = fraction.MustNewReduced(1, 0)
		}, fraction.ErrZeroDenominator)
	})
}

// TestFromInt verifies criterion K-4.
func TestFromInt(t *testing.T) {
	t.Parallel()

	inputs := []int{-100, -4, -1, 0, 1, 3, 100, math.MinInt, math.MaxInt}
	for _, n := range inputs {
		f := fraction.FromInt(n)
		assert.True(t, f.IsValid())
		assert.True(t, f.IsReduced())
		assert.True(t, f.IsInteger())
		assertExactPairInt(t, f, n, 1)
		assertReducedPairInt(t, f, n, 1)

		gotInt, ok := f.Int()
		require.True(t, ok)
		assert.Equal(t, n, gotInt)

		gotBigInt, ok := f.BigInt()
		require.True(t, ok)
		assert.Equal(t, 0, gotBigInt.Cmp(big.NewInt(int64(n))))
	}
}

// TestBigIntAndBigRat verifies criterion K-5 (FromBigInt, FromBigIntReduced, FromBigRat, Big* getters, defensive copies).
func TestBigIntAndBigRat(t *testing.T) {
	t.Parallel()

	t.Run("unreduced_10pow30_over_2x10pow30", func(t *testing.T) {
		t.Parallel()

		d30 := d30Big()
		twoD30 := new(big.Int).Mul(big.NewInt(2), d30)

		f, err := fraction.FromBigInt(d30, twoD30)
		require.NoError(t, err)
		mf := fraction.MustFromBigInt(d30, twoD30)
		assert.Equal(t, f, mf)
		assert.True(t, f.IsValid())
		assert.False(t, f.IsReduced())
		assert.True(t, f.Equal(fraction.MustNew(1, 2)))

		// Raw pair does not fit in 64-bit int.
		_, ok := f.Numerator()
		assert.False(t, ok)
		_, ok = f.Denominator()
		assert.False(t, ok)
		_, _, ok = f.Pair()
		assert.False(t, ok)

		// Reduced pair (1, 2) fits in 64-bit int.
		assertReducedPairInt(t, f, 1, 2)

		// BigPair preserves (10^30, 2*10^30).
		bNum, bDen, ok := f.BigPair()
		require.True(t, ok)
		assert.Equal(t, 0, bNum.Cmp(d30))
		assert.Equal(t, 0, bDen.Cmp(twoD30))

		// FromBigIntReduced immediately reduces to (1, 2).
		red, err := fraction.FromBigIntReduced(d30, twoD30)
		require.NoError(t, err)
		mred := fraction.MustFromBigIntReduced(d30, twoD30)
		assert.Equal(t, red, mred)
		assertExactPairInt(t, red, 1, 2)
		assert.True(t, red.EqualExact(f.Reduce()))
	})

	t.Run("mercator_comma_3pow53_over_2pow84", func(t *testing.T) {
		t.Parallel()

		mNum := mercatorNumBig()
		mDen := mercatorDenBig()

		f, err := fraction.FromBigInt(mNum, mDen)
		require.NoError(t, err)
		assert.True(t, f.IsValid())
		assert.True(t, f.IsReduced())
		assert.False(t, f.IsInteger())
		assert.False(t, f.IsProper())

		bNum, bDen, ok := f.BigPair()
		require.True(t, ok)
		assert.Equal(t, 0, bNum.Cmp(mNum))
		assert.Equal(t, 0, bDen.Cmp(mDen))

		rbNum, rbDen, ok := f.ReducedBigPair()
		require.True(t, ok)
		assert.Equal(t, 0, rbNum.Cmp(mNum))
		assert.Equal(t, 0, rbDen.Cmp(mDen))

		rat, ok := f.BigRat()
		require.True(t, ok)
		wantRat := new(big.Rat).SetFrac(mNum, mDen)
		assert.Equal(t, 0, rat.Cmp(wantRat))

		fromRat, err := fraction.FromBigRat(wantRat)
		require.NoError(t, err)
		mFromRat := fraction.MustFromBigRat(wantRat)
		assert.Equal(t, f, fromRat)
		assert.Equal(t, f, mFromRat)

		// Narrow getters return ok == false.
		_, _, ok = f.Pair()
		assert.False(t, ok)
		_, _, ok = f.ReducedPair()
		assert.False(t, ok)
		_, ok = f.Int()
		assert.False(t, ok)
		_, ok = f.BigInt()
		assert.False(t, ok)
	})

	t.Run("from_big_rat_6_over_12", func(t *testing.T) {
		t.Parallel()

		r := big.NewRat(6, 12)
		f, err := fraction.FromBigRat(r)
		require.NoError(t, err)
		assertExactPairInt(t, f, 1, 2)
		assert.True(t, f.IsReduced())
	})

	t.Run("nil_and_zero_error_priority", func(t *testing.T) {
		t.Parallel()

		one := big.NewInt(1)
		zero := big.NewInt(0)

		// Step 1: num == nil -> ErrInvalid (even if den is nil or 0).
		for _, den := range []*big.Int{one, nil, zero} {
			_, err := fraction.FromBigInt(nil, den)
			require.ErrorIs(t, err, fraction.ErrInvalid)
			_, err = fraction.FromBigIntReduced(nil, den)
			require.ErrorIs(t, err, fraction.ErrInvalid)

			assertMustPanicsWithError(t, func() {
				_ = fraction.MustFromBigInt(nil, den)
			}, fraction.ErrInvalid)
			assertMustPanicsWithError(t, func() {
				_ = fraction.MustFromBigIntReduced(nil, den)
			}, fraction.ErrInvalid)
		}

		// Step 2: num != nil and (den == nil || den == 0) -> ErrZeroDenominator.
		for _, den := range []*big.Int{nil, zero} {
			_, err := fraction.FromBigInt(one, den)
			require.ErrorIs(t, err, fraction.ErrZeroDenominator)
			_, err = fraction.FromBigIntReduced(one, den)
			require.ErrorIs(t, err, fraction.ErrZeroDenominator)

			assertMustPanicsWithError(t, func() {
				_ = fraction.MustFromBigInt(one, den)
			}, fraction.ErrZeroDenominator)
			assertMustPanicsWithError(t, func() {
				_ = fraction.MustFromBigIntReduced(one, den)
			}, fraction.ErrZeroDenominator)
		}

		// FromBigRat(nil) -> ErrInvalid.
		_, err := fraction.FromBigRat(nil)
		require.ErrorIs(t, err, fraction.ErrInvalid)
		assertMustPanicsWithError(t, func() {
			_ = fraction.MustFromBigRat(nil)
		}, fraction.ErrInvalid)
	})

	t.Run("defensive_copies_on_inputs_and_outputs", func(t *testing.T) {
		t.Parallel()

		inNum := big.NewInt(6)
		inDen := big.NewInt(12)
		f, err := fraction.FromBigInt(inNum, inDen)
		require.NoError(t, err)

		// Mutate inputs after creation.
		inNum.SetInt64(999)
		inDen.SetInt64(888)
		assertExactPairInt(t, f, 6, 12)

		inRat := big.NewRat(3, 4)
		fr, err := fraction.FromBigRat(inRat)
		require.NoError(t, err)
		inRat.SetInt64(42)
		assertExactPairInt(t, fr, 3, 4)

		// Mutate returned big.Int / big.Rat values.
		bNum, ok := f.BigNumerator()
		require.True(t, ok)
		bNum.SetInt64(-111)

		bDen, ok := f.BigDenominator()
		require.True(t, ok)
		bDen.SetInt64(-222)

		bpNum, bpDen, ok := f.BigPair()
		require.True(t, ok)
		bpNum.SetInt64(-333)
		bpDen.SetInt64(-444)

		rbNum, ok := f.ReducedBigNumerator()
		require.True(t, ok)
		rbNum.SetInt64(-555)

		rbDen, ok := f.ReducedBigDenominator()
		require.True(t, ok)
		rbDen.SetInt64(-666)

		rbpNum, rbpDen, ok := f.ReducedBigPair()
		require.True(t, ok)
		rbpNum.SetInt64(-777)
		rbpDen.SetInt64(-888)

		ratOut, ok := f.BigRat()
		require.True(t, ok)
		ratOut.SetInt64(-999)

		assertExactPairInt(t, f, 6, 12)
		assertReducedPairInt(t, f, 1, 2)

		intFrac := fraction.MustNew(6, 3)
		bigI, ok := intFrac.BigInt()
		require.True(t, ok)
		assert.Equal(t, 0, bigI.Cmp(big.NewInt(2)))
		bigI.SetInt64(777)
		gotInt, ok := intFrac.Int()
		require.True(t, ok)
		assert.Equal(t, 2, gotInt)
	})
}

// TestComponentsAndReduce verifies criterion K-6.
func TestComponentsAndReduce(t *testing.T) {
	t.Parallel()

	t.Run("6_over_12_and_immutability", func(t *testing.T) {
		t.Parallel()

		f := fraction.MustNew(6, 12)
		assertExactPairInt(t, f, 6, 12)
		assertReducedPairInt(t, f, 1, 2)

		_, ok := f.Int()
		assert.False(t, ok)
		_, ok = f.BigInt()
		assert.False(t, ok)

		red := f.Reduce()
		assertExactPairInt(t, red, 1, 2)
		assertReducedPairInt(t, red, 1, 2)
		assert.True(t, red.IsReduced())

		// Original instance is unchanged.
		assertExactPairInt(t, f, 6, 12)
	})

	t.Run("1_over_minus2", func(t *testing.T) {
		t.Parallel()

		f := fraction.MustNew(1, -2)
		assertExactPairInt(t, f, 1, -2)
		assertReducedPairInt(t, f, -1, 2)
		assertExactPairInt(t, f.Reduce(), -1, 2)
	})

	t.Run("integer_fractions_0_over_3_and_6_over_3_and_minus4_over_2", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			num     int
			den     int
			wantInt int
		}{
			{0, 3, 0},
			{0, -5, 0},
			{6, 3, 2},
			{-4, 2, -2},
		}

		for _, tc := range cases {
			f := fraction.MustNew(tc.num, tc.den)
			assertExactPairInt(t, f, tc.num, tc.den)
			assertReducedPairInt(t, f, tc.wantInt, 1)

			got, ok := f.Int()
			require.True(t, ok)
			assert.Equal(t, tc.wantInt, got)

			gotBig, ok := f.BigInt()
			require.True(t, ok)
			assert.Equal(t, 0, gotBig.Cmp(big.NewInt(int64(tc.wantInt))))
		}
	})

	t.Run("large_integer_outside_int64", func(t *testing.T) {
		t.Parallel()

		mNum := mercatorNumBig()
		f := fraction.MustFromBigInt(mNum, big.NewInt(1))
		assert.True(t, f.IsInteger())

		_, ok := f.Int()
		assert.False(t, ok)

		gotBig, ok := f.BigInt()
		require.True(t, ok)
		assert.Equal(t, 0, gotBig.Cmp(mNum))
	})
}

// TestWithDenominator verifies criterion K-7 (WithDenominator, WithBigDenominator, Must*).
func TestWithDenominator(t *testing.T) {
	t.Parallel()

	successCases := []struct {
		name      string
		num       int
		den       int
		targetDen int
		wantNum   int
		wantDen   int
	}{
		{name: "1/3_to_12", num: 1, den: 3, targetDen: 12, wantNum: 4, wantDen: 12},
		{name: "1/4_to_12", num: 1, den: 4, targetDen: 12, wantNum: 3, wantDen: 12},
		{name: "6/12_to_2", num: 6, den: 12, targetDen: 2, wantNum: 1, wantDen: 2},
		{name: "6/12_to_6", num: 6, den: 12, targetDen: 6, wantNum: 3, wantDen: 6},
		{name: "0/3_to_5", num: 0, den: 3, targetDen: 5, wantNum: 0, wantDen: 5},
		{name: "1/-2_to_4", num: 1, den: -2, targetDen: 4, wantNum: -2, wantDen: 4},
	}

	for _, tc := range successCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orig := fraction.MustNew(tc.num, tc.den)
			got, err := orig.WithDenominator(tc.targetDen)
			require.NoError(t, err)
			assertExactPairInt(t, got, tc.wantNum, tc.wantDen)
			assert.True(t, got.Equal(orig))

			mgot := orig.MustWithDenominator(tc.targetDen)
			assert.Equal(t, got, mgot)

			targetBig := big.NewInt(int64(tc.targetDen))
			bgot, err := orig.WithBigDenominator(targetBig)
			require.NoError(t, err)
			assert.Equal(t, got, bgot)
			assert.Equal(t, got, orig.MustWithBigDenominator(targetBig))

			// Defensive copy of targetBig and immutability of orig.
			targetBig.SetInt64(999)
			assertExactPairInt(t, bgot, tc.wantNum, tc.wantDen)
			assertExactPairInt(t, orig, tc.num, tc.den)
		})
	}

	t.Run("not_representable_error", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			num       int
			den       int
			targetDen int
		}{
			{1, 3, 10},
			{1, 4, 6},
			{1, 12, 4},
		}

		for _, tc := range cases {
			f := fraction.MustNew(tc.num, tc.den)
			_, err := f.WithDenominator(tc.targetDen)
			require.ErrorIs(t, err, fraction.ErrNotRepresentable)

			_, err = f.WithBigDenominator(big.NewInt(int64(tc.targetDen)))
			require.ErrorIs(t, err, fraction.ErrNotRepresentable)

			assertMustPanicsWithError(t, func() {
				_ = f.MustWithDenominator(tc.targetDen)
			}, fraction.ErrNotRepresentable)

			assertMustPanicsWithError(t, func() {
				_ = f.MustWithBigDenominator(big.NewInt(int64(tc.targetDen)))
			}, fraction.ErrNotRepresentable)
		}
	})

	t.Run("non_positive_denominator_error", func(t *testing.T) {
		t.Parallel()

		f := fraction.MustNew(1, 2)
		for _, d := range []int{0, -1, -4, math.MinInt} {
			_, err := f.WithDenominator(d)
			require.ErrorIs(t, err, fraction.ErrZeroDenominator)

			_, err = f.WithBigDenominator(big.NewInt(int64(d)))
			require.ErrorIs(t, err, fraction.ErrZeroDenominator)

			assertMustPanicsWithError(t, func() {
				_ = f.MustWithDenominator(d)
			}, fraction.ErrZeroDenominator)
		}

		_, err := f.WithBigDenominator(nil)
		require.ErrorIs(t, err, fraction.ErrZeroDenominator)
		assertMustPanicsWithError(t, func() {
			_ = f.MustWithBigDenominator(nil)
		}, fraction.ErrZeroDenominator)
	})
}

// TestEqualityAndMapKeys verifies criterion K-8 (Equal, EqualExact, ==, map keys).
func TestEqualityAndMapKeys(t *testing.T) {
	t.Parallel()

	t.Run("numeric_equality_classes", func(t *testing.T) {
		t.Parallel()

		classes := [][]rawPair{
			{{6, 12}, {1, 2}, {-2, -4}, {2, 4}, {-1, -2}},
			{{0, 1}, {0, 3}, {0, -5}},
			{{1, -2}, {-1, 2}, {-6, 12}, {6, -12}},
			{{6, 3}, {2, 1}, {-4, -2}},
			{{3, 2}, {6, 4}},
		}

		for ci, classA := range classes {
			for i, pa := range classA {
				fa := fraction.MustNew(pa.num, pa.den)
				for j, pb := range classA {
					fb := fraction.MustNew(pb.num, pb.den)
					assert.True(t, fa.Equal(fb), "%v should numerically equal %v", pa, pb)
					assert.Equal(t, i == j, fa.EqualExact(fb), "EqualExact(%v, %v)", pa, pb)
					assert.Equal(t, i == j, fa == fb, "== (%v, %v)", pa, pb)
					assert.Equal(t, fa.Reduce(), fb.Reduce())
				}

				for cj, classB := range classes {
					if ci == cj {
						continue
					}
					for _, pb := range classB {
						fb := fraction.MustNew(pb.num, pb.den)
						assert.False(t, fa.Equal(fb), "%v should not equal %v", pa, pb)
						assert.False(t, fa.EqualExact(fb))
						assert.NotEqual(t, fa, fb)
					}
				}
			}
		}
	})

	t.Run("map_keys_exact_vs_reduced", func(t *testing.T) {
		t.Parallel()

		f32 := fraction.MustNew(3, 2)
		f64 := fraction.MustNew(6, 4)

		exactMap := make(map[fraction.Fraction]int)
		exactMap[f32] = 10
		exactMap[f64] = 20
		assert.Len(t, exactMap, 2)
		assert.Equal(t, 10, exactMap[f32])
		assert.Equal(t, 20, exactMap[f64])

		reducedMap := make(map[fraction.Fraction]int)
		reducedMap[f32.Reduce()] = 10
		reducedMap[f64.Reduce()] = 20
		assert.Len(t, reducedMap, 1)
		assert.Equal(t, 20, reducedMap[f32])
		assert.Equal(t, 20, reducedMap[f32.Reduce()])
		assert.Equal(t, 20, reducedMap[f64.Reduce()])
	})
}

// TestComparisonAndOrdering verifies criterion K-9 (Compare, CompareExact, Less, LessOrEqual, Greater, GreaterOrEqual).
func TestComparisonAndOrdering(t *testing.T) {
	t.Parallel()

	t.Run("ascending_chain", func(t *testing.T) {
		t.Parallel()

		// Levels in strictly increasing numeric order; pairs within a level are numerically equal.
		levels := [][]rawPair{
			{{-7, 2}},
			{{-4, 2}},
			{{1, -2}, {-1, 2}},
			{{0, 3}, {0, 1}, {0, -5}},
			{{1, 3}},
			{{1, 2}, {6, 12}, {-2, -4}},
			{{2, 3}},
			{{3, 2}, {6, 4}},
			{{6, 3}},
		}

		for i, levelA := range levels {
			for _, pa := range levelA {
				fa := fraction.MustNew(pa.num, pa.den)
				for j, levelB := range levels {
					for _, pb := range levelB {
						fb := fraction.MustNew(pb.num, pb.den)
						wantCmp := 0
						if i < j {
							wantCmp = -1
						} else if i > j {
							wantCmp = 1
						}

						assert.Equal(t, wantCmp, fa.Compare(fb), "Compare(%v, %v)", pa, pb)
						assert.Equal(t, wantCmp == 0, fa.Equal(fb))
						assert.Equal(t, wantCmp < 0, fa.Less(fb))
						assert.Equal(t, wantCmp <= 0, fa.LessOrEqual(fb))
						assert.Equal(t, wantCmp > 0, fa.Greater(fb))
						assert.Equal(t, wantCmp >= 0, fa.GreaterOrEqual(fb))
					}
				}
			}
		}
	})

	t.Run("all_pairs_in_S_match_big_rat", func(t *testing.T) {
		t.Parallel()

		s := sampleSetS()
		for _, pa := range s {
			fa := fraction.MustNew(pa.num, pa.den)
			ra := big.NewRat(int64(pa.num), int64(pa.den))
			for _, pb := range s {
				fb := fraction.MustNew(pb.num, pb.den)
				rb := big.NewRat(int64(pb.num), int64(pb.den))

				want := ra.Cmp(rb)
				assert.Equal(t, want, fa.Compare(fb))
				assert.Equal(t, want == 0, fa.Equal(fb))
				assert.Equal(t, want < 0, fa.Less(fb))
				assert.Equal(t, want <= 0, fa.LessOrEqual(fb))
				assert.Equal(t, want > 0, fa.Greater(fb))
				assert.Equal(t, want >= 0, fa.GreaterOrEqual(fb))
			}
		}
	})

	t.Run("CompareExact_total_order_and_sorting", func(t *testing.T) {
		t.Parallel()

		var z fraction.Fraction
		f6_12 := fraction.MustNew(6, 12)
		f1_m2 := fraction.MustNew(1, -2)
		fm1_2 := fraction.MustNew(-1, 2)
		f1_2 := fraction.MustNew(1, 2)
		f0_3 := fraction.MustNew(0, 3)
		f0_1 := fraction.MustNew(0, 1)

		wantSorted := []fraction.Fraction{
			z,
			f1_m2,
			fm1_2,
			f0_1,
			f0_3,
			f1_2,
			f6_12,
		}

		permutations := [][]fraction.Fraction{
			{f6_12, f1_m2, fm1_2, f1_2, f0_3, f0_1, z},
			{z, f0_1, f0_3, f1_2, f6_12, fm1_2, f1_m2},
			{f1_2, f0_3, z, f6_12, f1_m2, f0_1, fm1_2},
			{fm1_2, f6_12, f0_1, f1_m2, z, f1_2, f0_3},
		}

		for _, perm := range permutations {
			cp := slices.Clone(perm)
			slices.SortFunc(cp, func(a, b fraction.Fraction) int {
				return a.CompareExact(b)
			})
			assert.Equal(t, wantSorted, cp)
		}

		for i, a := range wantSorted {
			for j, b := range wantSorted {
				want := 0
				if i < j {
					want = -1
				} else if i > j {
					want = 1
				}
				assert.Equal(t, want, a.CompareExact(b))
				assert.Equal(t, want == 0, a == b)
			}
		}
	})
}

// TestPredicatesAndSign verifies criterion K-10 across all fractions in S and additional cases.
func TestPredicatesAndSign(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		f         fraction.Fraction
		isValid   bool
		isZero    bool
		isOne     bool
		isInteger bool
		isProper  bool
		isReduced bool
		sign      int
	}{
		{name: "invalid_0/0", f: fraction.Fraction{}, isValid: false, isZero: false, isOne: false, isInteger: false, isProper: false, isReduced: false, sign: 0},
		{name: "0/1", f: fraction.MustNew(0, 1), isValid: true, isZero: true, isOne: false, isInteger: true, isProper: true, isReduced: true, sign: 0},
		{name: "0/3", f: fraction.MustNew(0, 3), isValid: true, isZero: true, isOne: false, isInteger: true, isProper: true, isReduced: false, sign: 0},
		{name: "0/-5", f: fraction.MustNew(0, -5), isValid: true, isZero: true, isOne: false, isInteger: true, isProper: true, isReduced: false, sign: 0},
		{name: "1/1", f: fraction.MustNew(1, 1), isValid: true, isZero: false, isOne: true, isInteger: true, isProper: false, isReduced: true, sign: 1},
		{name: "-1/1", f: fraction.MustNew(-1, 1), isValid: true, isZero: false, isOne: false, isInteger: true, isProper: false, isReduced: true, sign: -1},
		{name: "1/2", f: fraction.MustNew(1, 2), isValid: true, isZero: false, isOne: false, isInteger: false, isProper: true, isReduced: true, sign: 1},
		{name: "6/12", f: fraction.MustNew(6, 12), isValid: true, isZero: false, isOne: false, isInteger: false, isProper: true, isReduced: false, sign: 1},
		{name: "1/-2", f: fraction.MustNew(1, -2), isValid: true, isZero: false, isOne: false, isInteger: false, isProper: true, isReduced: false, sign: -1},
		{name: "-2/-4", f: fraction.MustNew(-2, -4), isValid: true, isZero: false, isOne: false, isInteger: false, isProper: true, isReduced: false, sign: 1},
		{name: "1/3", f: fraction.MustNew(1, 3), isValid: true, isZero: false, isOne: false, isInteger: false, isProper: true, isReduced: true, sign: 1},
		{name: "2/3", f: fraction.MustNew(2, 3), isValid: true, isZero: false, isOne: false, isInteger: false, isProper: true, isReduced: true, sign: 1},
		{name: "-3/4", f: fraction.MustNew(-3, 4), isValid: true, isZero: false, isOne: false, isInteger: false, isProper: true, isReduced: true, sign: -1},
		{name: "3/2", f: fraction.MustNew(3, 2), isValid: true, isZero: false, isOne: false, isInteger: false, isProper: false, isReduced: true, sign: 1},
		{name: "6/4", f: fraction.MustNew(6, 4), isValid: true, isZero: false, isOne: false, isInteger: false, isProper: false, isReduced: false, sign: 1},
		{name: "14/12", f: fraction.MustNew(14, 12), isValid: true, isZero: false, isOne: false, isInteger: false, isProper: false, isReduced: false, sign: 1},
		{name: "-13/12", f: fraction.MustNew(-13, 12), isValid: true, isZero: false, isOne: false, isInteger: false, isProper: false, isReduced: true, sign: -1},
		{name: "-1/12", f: fraction.MustNew(-1, 12), isValid: true, isZero: false, isOne: false, isInteger: false, isProper: true, isReduced: true, sign: -1},
		{name: "6/3", f: fraction.MustNew(6, 3), isValid: true, isZero: false, isOne: false, isInteger: true, isProper: false, isReduced: false, sign: 1},
		{name: "-4/2", f: fraction.MustNew(-4, 2), isValid: true, isZero: false, isOne: false, isInteger: true, isProper: false, isReduced: false, sign: -1},
		{name: "3/3", f: fraction.MustNew(3, 3), isValid: true, isZero: false, isOne: true, isInteger: true, isProper: false, isReduced: false, sign: 1},
		{name: "-2/-2", f: fraction.MustNew(-2, -2), isValid: true, isZero: false, isOne: true, isInteger: true, isProper: false, isReduced: false, sign: 1},
		{name: "2/1", f: fraction.MustNew(2, 1), isValid: true, isZero: false, isOne: false, isInteger: true, isProper: false, isReduced: true, sign: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.isValid, tc.f.IsValid(), "IsValid")
			assert.Equal(t, tc.isZero, tc.f.IsZero(), "IsZero")
			assert.Equal(t, tc.isOne, tc.f.IsOne(), "IsOne")
			assert.Equal(t, tc.isInteger, tc.f.IsInteger(), "IsInteger")
			assert.Equal(t, tc.isProper, tc.f.IsProper(), "IsProper")
			assert.Equal(t, tc.isReduced, tc.f.IsReduced(), "IsReduced")
			assert.Equal(t, tc.sign, tc.f.Sign(), "Sign")
		})
	}
}

// TestInvalidFraction_Part1 verifies criterion K-24 (part 1: predicates, getters, comparisons, Reduce, WithDenominator on zero value).
func TestInvalidFraction_Part1(t *testing.T) {
	t.Parallel()

	var z fraction.Fraction
	zero := fraction.Zero()

	assert.False(t, z.IsValid())
	assert.False(t, z.IsZero())
	assert.False(t, z.IsOne())
	assert.False(t, z.IsInteger())
	assert.False(t, z.IsProper())
	assert.False(t, z.IsReduced())
	assert.Equal(t, 0, z.Sign())

	assert.False(t, z.Equal(z))
	assert.False(t, z.Equal(zero))
	assert.False(t, zero.Equal(z))
	assert.False(t, z.EqualExact(z))
	assert.False(t, z.EqualExact(zero))
	assert.False(t, zero.EqualExact(z))

	assert.True(t, z == fraction.Fraction{})
	assert.False(t, z == zero)

	assert.Equal(t, 0, z.Compare(z))
	assert.Equal(t, -1, z.Compare(zero))
	assert.Equal(t, 1, zero.Compare(z))
	assert.Equal(t, 0, z.CompareExact(z))
	assert.Equal(t, -1, z.CompareExact(zero))
	assert.Equal(t, 1, zero.CompareExact(z))

	assert.False(t, z.Less(z))
	assert.False(t, z.Less(zero))
	assert.False(t, zero.Less(z))
	assert.False(t, z.LessOrEqual(z))
	assert.False(t, z.LessOrEqual(zero))
	assert.False(t, zero.LessOrEqual(z))
	assert.False(t, z.Greater(z))
	assert.False(t, z.Greater(zero))
	assert.False(t, zero.Greater(z))
	assert.False(t, z.GreaterOrEqual(z))
	assert.False(t, z.GreaterOrEqual(zero))
	assert.False(t, zero.GreaterOrEqual(z))

	num, ok := z.Numerator()
	assert.False(t, ok)
	assert.Equal(t, 0, num)

	den, ok := z.Denominator()
	assert.False(t, ok)
	assert.Equal(t, 0, den)

	pNum, pDen, ok := z.Pair()
	assert.False(t, ok)
	assert.Equal(t, 0, pNum)
	assert.Equal(t, 0, pDen)

	rNum, ok := z.ReducedNumerator()
	assert.False(t, ok)
	assert.Equal(t, 0, rNum)

	rDen, ok := z.ReducedDenominator()
	assert.False(t, ok)
	assert.Equal(t, 0, rDen)

	rpNum, rpDen, ok := z.ReducedPair()
	assert.False(t, ok)
	assert.Equal(t, 0, rpNum)
	assert.Equal(t, 0, rpDen)

	iv, ok := z.Int()
	assert.False(t, ok)
	assert.Equal(t, 0, iv)

	bNum, ok := z.BigNumerator()
	assert.False(t, ok)
	assert.Nil(t, bNum)

	bDen, ok := z.BigDenominator()
	assert.False(t, ok)
	assert.Nil(t, bDen)

	bpNum, bpDen, ok := z.BigPair()
	assert.False(t, ok)
	assert.Nil(t, bpNum)
	assert.Nil(t, bpDen)

	rbNum, ok := z.ReducedBigNumerator()
	assert.False(t, ok)
	assert.Nil(t, rbNum)

	rbDen, ok := z.ReducedBigDenominator()
	assert.False(t, ok)
	assert.Nil(t, rbDen)

	rbpNum, rbpDen, ok := z.ReducedBigPair()
	assert.False(t, ok)
	assert.Nil(t, rbpNum)
	assert.Nil(t, rbpDen)

	bi, ok := z.BigInt()
	assert.False(t, ok)
	assert.Nil(t, bi)

	br, ok := z.BigRat()
	assert.False(t, ok)
	assert.Nil(t, br)

	assert.Equal(t, fraction.Fraction{}, z.Reduce())

	_, err := z.WithDenominator(4)
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.WithDenominator(0) // Step 1 (ErrInvalid) precedes Step 2 (ErrZeroDenominator).
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.WithBigDenominator(big.NewInt(4))
	require.ErrorIs(t, err, fraction.ErrInvalid)
	_, err = z.WithBigDenominator(nil)
	require.ErrorIs(t, err, fraction.ErrInvalid)

	assertMustPanicsWithError(t, func() {
		_ = z.MustWithDenominator(4)
	}, fraction.ErrInvalid)
	assertMustPanicsWithError(t, func() {
		_ = z.MustWithBigDenominator(big.NewInt(4))
	}, fraction.ErrInvalid)

	// Confirm errors.Is works directly as expected.
	assert.True(t, errors.Is(err, fraction.ErrInvalid))
}
