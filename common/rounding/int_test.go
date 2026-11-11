package rounding_test

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModeQuoRemControlQuotients(t *testing.T) {
	t.Parallel()
	inputs := [][2]int{{1, 2}, {-1, 2}, {3, 2}, {-3, 2}, {5, 2}, {-5, 2}, {7, 3}, {-7, 3}, {8, 3}, {-8, 3}}
	quotients := [][]int{
		{0, -1, 1, -2, 2, -3, 2, -3, 2, -3},
		{1, 0, 2, -1, 3, -2, 3, -2, 3, -2},
		{0, 0, 1, -1, 2, -2, 2, -2, 2, -2},
		{1, -1, 2, -2, 3, -3, 3, -3, 3, -3},
		{1, 0, 2, -1, 3, -2, 2, -2, 3, -3},
		{0, -1, 1, -2, 2, -3, 2, -2, 3, -3},
		{1, -1, 2, -2, 3, -3, 2, -2, 3, -3},
		{0, 0, 1, -1, 2, -2, 2, -2, 3, -3},
		{0, 0, 2, -2, 2, -2, 2, -2, 3, -3},
	}
	for i, mode := range testModes {
		for j, input := range inputs {
			t.Run(fmt.Sprintf("%s/%d_over_%d", testModeNames[i], input[0], input[1]), func(t *testing.T) {
				t.Parallel()
				q, r, err := mode.QuoRem(input[0], input[1])
				require.NoError(t, err)
				assert.Equal(t, quotients[i][j], q)
				assert.Equal(t, input[0]-quotients[i][j]*input[1], r)
			})
		}
	}
}

func TestModeQuoRemControlEdges(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mode rounding.Mode
		n, d int
		q, r int
		err  error
	}{
		{"down", rounding.Down, 7, 3, 2, 1, nil},
		{"up", rounding.Up, 7, 3, 3, -2, nil},
		{"negative_divisor", rounding.Down, 7, -3, -3, -2, nil},
		{"min_over_minus_one", rounding.Down, math.MinInt, -1, 0, 0, rounding.ErrOutOfRange},
		{"max_over_one", rounding.Up, math.MaxInt, 1, math.MaxInt, 0, nil},
		{"max_half_even", rounding.HalfEven, math.MaxInt, 2, 1 << 62, -1, nil},
		{"max_half_down", rounding.HalfDown, math.MaxInt, 2, (1 << 62) - 1, 1, nil},
		{"zero_divisor", rounding.Down, 7, 0, 0, 0, rounding.ErrInvalidArgument},
		{"unset_precedes_zero_divisor", 0, 7, 0, 0, 0, rounding.ErrUnset},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			q, r, err := tt.mode.QuoRem(tt.n, tt.d)
			if tt.err != nil {
				checkError(t, err, tt.err)
				assert.Zero(t, q)
				assert.Zero(t, r)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.q, q)
			assert.Equal(t, tt.r, r)
		})
	}
	for _, tt := range []struct {
		name string
		n, d int
		q    int
	}{
		{"positive_node", 6, 3, 2},
		{"negative_node", -6, 3, -2},
		{"zero", 0, 5, 0},
	} {
		for i, mode := range testModes {
			t.Run(fmt.Sprintf("%s/%s", tt.name, testModeNames[i]), func(t *testing.T) {
				t.Parallel()
				q, r, err := mode.QuoRem(tt.n, tt.d)
				require.NoError(t, err)
				assert.Equal(t, tt.q, q)
				assert.Zero(t, r)
			})
		}
	}
}

func TestModeQuoRemExactModel(t *testing.T) {
	t.Parallel()
	ns := []int{0, 1, -1, math.MaxInt, math.MinInt, math.MinInt + 1}
	ds := []int{1, -1, 2, -2, 3, -3, math.MaxInt, math.MinInt, math.MinInt + 1}
	for _, n := range ns {
		for _, d := range ds {
			for i, mode := range testModes {
				t.Run(fmt.Sprintf("%d_over_%d/%s", n, d, testModeNames[i]), func(t *testing.T) {
					t.Parallel()
					q, r, err := mode.QuoRem(n, d)
					exact := new(big.Rat).SetFrac(bigInt(n), bigInt(d))
					wantQ := roundRat(mode, exact)
					if !intInRange(wantQ) {
						checkError(t, err, rounding.ErrOutOfRange)
						assert.Zero(t, q)
						assert.Zero(t, r)
						return
					}
					require.NoError(t, err)
					assert.Equal(t, int(wantQ.Int64()), q)
					wantR := new(big.Int).Sub(bigInt(n), new(big.Int).Mul(wantQ, bigInt(d)))
					assert.Equal(t, int(wantR.Int64()), r)
					absoluteR := new(big.Int).Abs(bigInt(r))
					absoluteD := new(big.Int).Abs(bigInt(d))
					assert.Equal(t, -1, absoluteR.Cmp(absoluteD))
				})
			}
		}
	}
}

func TestModeIntMultipleControlExamples(t *testing.T) {
	t.Parallel()
	modes := []rounding.Mode{rounding.Down, rounding.Up, rounding.HalfUp, rounding.HalfDown, rounding.HalfEven, rounding.HalfAwayFromZero, rounding.HalfTowardZero}
	tests := []struct {
		n, step int
		want    []int
	}{
		{1234, 480, []int{960, 1440, 1440, 1440, 1440, 1440, 1440}},
		{1200, 480, []int{960, 1440, 1440, 960, 960, 1440, 960}},
		{-1200, 480, []int{-1440, -960, -960, -1440, -960, -1440, -960}},
	}
	for _, tt := range tests {
		for i, mode := range modes {
			t.Run(fmt.Sprintf("%d_step_%d/%s", tt.n, tt.step, mode), func(t *testing.T) {
				t.Parallel()
				got, err := mode.IntMultiple(tt.n, tt.step)
				require.NoError(t, err)
				assert.Equal(t, tt.want[i], got)
			})
		}
	}
}

func TestModeIntMultipleEdgesAndErrors(t *testing.T) {
	t.Parallel()
	for _, step := range []int{0, -480, math.MinInt} {
		t.Run(fmt.Sprintf("invalid_step_%d", step), func(t *testing.T) {
			t.Parallel()
			got, err := rounding.HalfEven.IntMultiple(10, step)
			checkError(t, err, rounding.ErrInvalidArgument)
			assert.Zero(t, got)
		})
	}
	for _, tt := range []struct {
		name string
		mode rounding.Mode
		n    int
		step int
	}{
		{"max_up", rounding.Up, math.MaxInt, 2},
		{"min_plus_one_down", rounding.Down, math.MinInt + 1, 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.mode.IntMultiple(tt.n, tt.step)
			checkError(t, err, rounding.ErrOutOfRange)
			assert.Zero(t, got)
		})
	}
	t.Run("max_down", func(t *testing.T) {
		t.Parallel()
		got, err := rounding.Down.IntMultiple(math.MaxInt, 2)
		require.NoError(t, err)
		assert.Equal(t, math.MaxInt-1, got)
	})
	for i, mode := range testModes {
		t.Run(fmt.Sprintf("min_node/%s", testModeNames[i]), func(t *testing.T) {
			t.Parallel()
			got, err := mode.IntMultiple(math.MinInt, 2)
			require.NoError(t, err)
			assert.Equal(t, math.MinInt, got)
		})
	}
}
