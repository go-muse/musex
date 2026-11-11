package rounding_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModeQuoRemBigAndRatIntControlQuotients(t *testing.T) {
	t.Parallel()
	inputs := [][2]int{{1, 2}, {-1, 2}, {3, 2}, {-3, 2}, {5, 2}, {-5, 2}, {7, 3}, {-7, 3}, {8, 3}, {-8, 3}}
	for i, mode := range testModes {
		for _, input := range inputs {
			t.Run(fmt.Sprintf("%s/%d_over_%d", testModeNames[i], input[0], input[1]), func(t *testing.T) {
				t.Parallel()
				n, d := bigInt(input[0]), bigInt(input[1])
				q, r, err := mode.QuoRemBig(n, d)
				require.NoError(t, err)
				wantQ := roundRat(mode, bigRat(input[0], input[1]))
				assert.Zero(t, q.Cmp(wantQ))
				wantR := new(big.Int).Sub(n, new(big.Int).Mul(wantQ, d))
				assert.Zero(t, r.Cmp(wantR))
				gotInt, err := mode.RatInt(bigRat(input[0], input[1]))
				require.NoError(t, err)
				assert.Zero(t, gotInt.Cmp(wantQ))
			})
		}
	}
}

func TestModeQuoRemBigLargeControlExample(t *testing.T) {
	t.Parallel()
	d := new(big.Int).Lsh(big.NewInt(1), 100)
	half := new(big.Int).Lsh(big.NewInt(1), 99)
	n := new(big.Int).Add(d, half)
	nBefore, dBefore := n.String(), d.String()
	for _, tt := range []struct {
		name string
		mode rounding.Mode
		q    int64
		rNeg bool
	}{{"half_even", rounding.HalfEven, 2, true}, {"down", rounding.Down, 1, false}} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			q, r, err := tt.mode.QuoRemBig(n, d)
			require.NoError(t, err)
			assert.Equal(t, tt.q, q.Int64())
			wantR := new(big.Int).Set(half)
			if tt.rNeg {
				wantR.Neg(wantR)
			}
			assert.Zero(t, r.Cmp(wantR))
			assert.Equal(t, nBefore, n.String())
			assert.Equal(t, dBefore, d.String())
		})
	}
}

func TestModeRatOperationsControlExamples(t *testing.T) {
	t.Parallel()
	t.Run("rat_int_up", func(t *testing.T) {
		t.Parallel()
		got, err := rounding.Up.RatInt(bigRat(7, 3))
		require.NoError(t, err)
		assert.Equal(t, int64(3), got.Int64())
	})
	t.Run("rat_int_down_negative", func(t *testing.T) {
		t.Parallel()
		got, err := rounding.Down.RatInt(bigRat(-7, 3))
		require.NoError(t, err)
		assert.Equal(t, int64(-3), got.Int64())
	})
	for _, tt := range []struct {
		name    string
		x, step *big.Rat
		mode    rounding.Mode
		wantK   int64
		wantV   *big.Rat
	}{
		{"seven_thirds_quarters", bigRat(7, 3), bigRat(1, 4), rounding.HalfEven, 9, bigRat(9, 4)},
		{"tie_half_even", bigRat(5, 8), bigRat(1, 4), rounding.HalfEven, 2, bigRat(1, 2)},
		{"tie_half_up", bigRat(5, 8), bigRat(1, 4), rounding.HalfUp, 3, bigRat(3, 4)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			k, v, err := tt.mode.RatMultiple(tt.x, tt.step)
			require.NoError(t, err)
			assert.Equal(t, tt.wantK, k.Int64())
			assert.Zero(t, v.Cmp(tt.wantV))
		})
	}
	t.Run("binary_values_are_not_decimal_tie", func(t *testing.T) {
		t.Parallel()
		x := new(big.Rat).SetFloat64(0.15)
		step := new(big.Rat).SetFloat64(0.1)
		quotient := new(big.Rat).Quo(x, step)
		got, err := rounding.HalfUp.RatInt(quotient)
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.Int64())
		k, v, err := rounding.HalfUp.RatMultiple(x, step)
		require.NoError(t, err)
		assert.Equal(t, int64(1), k.Int64())
		assert.Equal(t, "3602879701896397/36028797018963968", v.RatString())
	})
}

func TestModeBigOperationsZeroValues(t *testing.T) {
	t.Parallel()
	zeroInt := new(big.Int)
	zeroRat := new(big.Rat)
	q, r, err := rounding.HalfEven.QuoRemBig(zeroInt, big.NewInt(3))
	require.NoError(t, err)
	assert.Zero(t, q.Sign())
	assert.Zero(t, r.Sign())
	gotInt, err := rounding.Up.RatInt(zeroRat)
	require.NoError(t, err)
	assert.Zero(t, gotInt.Sign())
	k, v, err := rounding.Down.RatMultiple(zeroRat, bigRat(1, 4))
	require.NoError(t, err)
	assert.Zero(t, k.Sign())
	assert.Zero(t, v.Sign())
}

func TestModeBigOperationsErrors(t *testing.T) {
	t.Parallel()
	t.Run("quo_rem_zero_divisor", func(t *testing.T) {
		t.Parallel()
		q, r, err := rounding.Down.QuoRemBig(big.NewInt(7), new(big.Int))
		checkError(t, err, rounding.ErrInvalidArgument)
		assert.Nil(t, q)
		assert.Nil(t, r)
	})
	t.Run("rat_multiple_zero_step", func(t *testing.T) {
		t.Parallel()
		k, v, err := rounding.Down.RatMultiple(bigRat(1, 2), new(big.Rat))
		checkError(t, err, rounding.ErrInvalidArgument)
		assert.Nil(t, k)
		assert.Nil(t, v)
	})
	t.Run("rat_multiple_negative_step", func(t *testing.T) {
		t.Parallel()
		k, v, err := rounding.Down.RatMultiple(bigRat(1, 2), bigRat(-1, 4))
		checkError(t, err, rounding.ErrInvalidArgument)
		assert.Nil(t, k)
		assert.Nil(t, v)
	})

	validInt := big.NewInt(2)
	validRat := bigRat(1, 2)
	for _, tt := range []struct {
		name string
		call func() error
	}{
		{"quo_rem_nil_n", func() error { _, _, err := rounding.Down.QuoRemBig(nil, validInt); return err }},
		{"quo_rem_nil_d", func() error { _, _, err := rounding.Down.QuoRemBig(validInt, nil); return err }},
		{"rat_int_nil", func() error { _, err := rounding.Down.RatInt(nil); return err }},
		{"rat_multiple_nil_x", func() error { _, _, err := rounding.Down.RatMultiple(nil, validRat); return err }},
		{"rat_multiple_nil_step", func() error { _, _, err := rounding.Down.RatMultiple(validRat, nil); return err }},
		{"rat_decimal_nil", func() error { _, err := rounding.Down.RatDecimal(nil, 2); return err }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkError(t, tt.call(), rounding.ErrUnset)
		})
	}

	for _, tt := range []struct {
		name string
		call func() error
		want error
	}{
		{"unset_precedes_nil_and_zero", func() error { _, _, err := rounding.Mode(0).QuoRemBig(nil, new(big.Int)); return err }, rounding.ErrUnset},
		{"unknown_after_nil", func() error { _, _, err := rounding.Mode(10).QuoRemBig(nil, new(big.Int)); return err }, rounding.ErrUnset},
		{"unknown_precedes_zero", func() error { _, _, err := rounding.Mode(10).QuoRemBig(validInt, new(big.Int)); return err }, rounding.ErrUnknownMode},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkError(t, tt.call(), tt.want)
		})
	}
}
