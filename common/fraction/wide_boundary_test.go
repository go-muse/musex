package fraction_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/require"
)

// Word-boundary tests (architecture of task 26.6, "Accuracy tests"). Operands
// are chosen so that the intermediate integers of the contract formulas (the
// table "Intermediate values of operations") fall next to the 64-, 128-, 192-,
// 256- and 512-bit boundaries, and every result is compared with the oracle.

// Factors of 2^64+1 and 2^128+1: their products hit those values exactly.
var (
	wF64a  = big.NewInt(274177)
	wF64b  = big.NewInt(67280421310721)
	wF128a = wDec("59649589127497217")
	wF128b = wDec("5704689200685129054721")
)

// wKnuthU and wKnuthV are an add-back pattern of long division in 64-bit words:
// the first estimate of the quotient digit is one too large.
var (
	wKnuthU = wHex("7fff_ffff_ffff_ffff_8000_0000_0000_0000_0000_0000_0000_0000_0000_0000_0000_0000")
	wKnuthV = wHex("8000_0000_0000_0000_0000_0000_0000_0000_0000_0000_0000_0001")
)

// wBoundaryPool returns operands built from word-boundary values; ordered pairs
// of the pool cover the products and sums named in the accuracy tests.
func wBoundaryPool() []fraction.Fraction {
	p := wPow2
	one := big.NewInt(1)
	maxC := fraction.MaxComponentBig()
	g64 := wOff(p(64), 13)
	g128 := wOff(p(128), 51)
	quoNum := wOff(wMul(wOff(p(64), -1), wOff(p(64), 1)), 1<<62) // (2^64-1)(2^64+1) + 2^62
	return []fraction.Fraction{
		wFrac(big.NewInt(0), one),
		wFrac(big.NewInt(0), wNeg(wOff(p(64), 1))),
		wFrac(big.NewInt(0), p(128)),
		wFrac(one, one),
		wFrac(big.NewInt(-1), one),
		wFrac(wOff(p(64), -1), one),
		wFrac(p(64), one),
		wFrac(wNeg(wOff(p(64), 1)), one),
		wFrac(p(63), big.NewInt(-1)),
		wFrac(one, wOff(p(32), -1)),
		wFrac(one, wOff(p(32), 1)),
		wFrac(big.NewInt(-1), p(32)),
		wFrac(one, wF64a),
		wFrac(big.NewInt(3), wF64b),
		wFrac(big.NewInt(5), p(64)),
		wFrac(big.NewInt(7), wOff(p(64), -1)),
		wFrac(big.NewInt(-11), wOff(p(64), 1)),
		wFrac(wOff(p(63), 1), p(63)),
		wFrac(wOff(p(64), -1), wNeg(wOff(p(63), 5))),
		wFrac(wOff(p(127), 1), wOff(p(64), 1)),
		wFrac(wOff(p(128), -1), wOff(p(127), 3)),
		wFrac(one, wF128a),
		wFrac(big.NewInt(-3), wF128b),
		wFrac(wOff(p(96), 1), p(96)),
		wFrac(big.NewInt(3), p(128)),
		wFrac(wOff(p(192), 1), wOff(p(128), -1)),
		wFrac(quoNum, wOff(p(64), 1)),
		wFrac(wOff(p(200), 7), wOff(p(64), 1)),
		wFrac(maxC, one),
		wFrac(wNeg(maxC), maxC),
		wFrac(maxC, maxC),
		wFrac(maxC, p(254)),
		wFrac(wKnuthU, wKnuthV),
		wFrac(wKnuthV, big.NewInt(3)),
		wFrac(wMul(g64, big.NewInt(3)), wMul(g64, big.NewInt(5))),
		wFrac(wMul(g128, big.NewInt(7)), wNeg(wMul(g128, big.NewInt(11)))),
		wFrac(wMul(p(100), big.NewInt(3)), wMul(p(90), big.NewInt(5))),
		wFrac(wOff(p(130), 1), wMul(wOff(p(130), 1), big.NewInt(3))),
		wFrac(wMul(wOff(p(70), 1), big.NewInt(-5)), wOff(p(70), 1)),
	}
}

// TestWideBoundaryPool runs all unary operations on each pool operand and all
// binary operations on each ordered pair of pool operands.
func TestWideBoundaryPool(t *testing.T) {
	t.Parallel()

	pool := wBoundaryPool()
	for i, x := range pool {
		t.Run(fmt.Sprintf("unary/%d_%s", i, x), func(t *testing.T) {
			t.Parallel()
			checkWideUnary(t, x)
		})
	}
	for i, x := range pool {
		t.Run(fmt.Sprintf("binary/%d", i), func(t *testing.T) {
			t.Parallel()
			for j, y := range pool {
				t.Run(fmt.Sprintf("%d_%s__%d_%s", i, x, j, y), func(t *testing.T) {
					checkWideBinary(t, x, y)
				})
			}
		})
	}
}

// wIntermediate names one intermediate integer of the contract formulas.
type wIntermediate uint8

const (
	wSumNum      wIntermediate = iota // |n1'·D2 + n2'·D1| of Add and modes 5 and 7
	wDenProd                          // D1·D2 of Add, Sub, Mul
	wMulNum                           // |n1'·n2'| of Mul
	wQuoDividend                      // |n1|·D2 of QuoRem, DivModFloor, IsMultipleOf
	wQuoDivisor                       // D1·|n2| of the same operations
	wQuotient                         // trunc(|n1|·D2 / (D1·|n2|))
)

func wIntermediateValue(kind wIntermediate, x, y fraction.Fraction) *big.Int {
	n1, d1 := wNorm(x)
	n2, d2 := wNorm(y)
	switch kind {
	case wSumNum:
		return new(big.Int).Abs(new(big.Int).Add(wMul(n1, d2), wMul(n2, d1)))
	case wDenProd:
		return wMul(d1, d2)
	case wMulNum:
		return new(big.Int).Abs(wMul(n1, n2))
	case wQuoDividend:
		return new(big.Int).Abs(wMul(n1, d2))
	case wQuoDivisor:
		return new(big.Int).Abs(wMul(d1, n2))
	default:
		return new(big.Int).Quo(new(big.Int).Abs(wMul(n1, d2)), new(big.Int).Abs(wMul(d1, n2)))
	}
}

// TestWideIntermediateBoundaries checks, for each case, that the named
// intermediate value is exactly the boundary value, then runs the operations.
func TestWideIntermediateBoundaries(t *testing.T) {
	t.Parallel()

	p := wPow2
	one := big.NewInt(1)
	maxC := fraction.MaxComponentBig()
	cases := []struct {
		name string
		x, y fraction.Fraction
		kind wIntermediate
		want *big.Int
	}{
		{"sum_2^64-1", wFrac(p(63), one), wFrac(wOff(p(63), -1), one), wSumNum, wOff(p(64), -1)},
		{"sum_2^64", wFrac(p(63), one), wFrac(p(63), one), wSumNum, p(64)},
		{"sum_2^64+1", wFrac(p(63), one), wFrac(wOff(p(63), 1), one), wSumNum, wOff(p(64), 1)},
		{"sum_cancels_to_zero", wFrac(p(62), big.NewInt(3)), wFrac(p(62), big.NewInt(-3)), wSumNum, big.NewInt(0)},
		{"sum_3·2^64_reduced_by_3", wFrac(p(63), big.NewInt(3)), wFrac(p(63), big.NewInt(3)), wSumNum, wMul(big.NewInt(3), p(64))},
		{"den_2^64-1", wFrac(one, wOff(p(32), -1)), wFrac(big.NewInt(-1), wOff(p(32), 1)), wDenProd, wOff(p(64), -1)},
		{"den_2^64", wFrac(big.NewInt(3), p(32)), wFrac(big.NewInt(5), wNeg(p(32))), wDenProd, p(64)},
		{"den_2^64+1", wFrac(one, wF64a), wFrac(one, wF64b), wDenProd, wOff(p(64), 1)},
		{"den_2^127", wFrac(big.NewInt(3), p(63)), wFrac(big.NewInt(5), p(64)), wDenProd, p(127)},
		{"den_2^128-1", wFrac(big.NewInt(7), wOff(p(64), -1)), wFrac(big.NewInt(5), wOff(p(64), 1)), wDenProd, wOff(p(128), -1)},
		{"den_2^128", wFrac(big.NewInt(3), p(64)), wFrac(big.NewInt(-1), p(64)), wDenProd, p(128)},
		{"den_2^128+1", wFrac(one, wF128a), wFrac(big.NewInt(2), wF128b), wDenProd, wOff(p(128), 1)},
		{"den_2^192", wFrac(wOff(p(96), 1), p(96)), wFrac(one, wNeg(p(96))), wDenProd, p(192)},
		{"den_2^256", wFrac(big.NewInt(3), p(128)), wFrac(big.NewInt(5), p(128)), wDenProd, p(256)},
		{"sum_2^511", wFrac(maxC, maxC), wFrac(maxC, maxC), wSumNum, wMul(big.NewInt(2), maxC, maxC)},
		{"mul_2^64", wFrac(p(32), big.NewInt(3)), wFrac(p(32), big.NewInt(5)), wMulNum, p(64)},
		{"mul_2^128-1", wFrac(wOff(p(64), -1), big.NewInt(3)), wFrac(wOff(p(64), 1), big.NewInt(7)), wMulNum, wOff(p(128), -1)},
		{"mul_2^128+1", wFrac(wF128a, big.NewInt(9)), wFrac(wF128b, big.NewInt(-11)), wMulNum, wOff(p(128), 1)},
		{"mul_2^256", wFrac(p(128), big.NewInt(3)), wFrac(p(128), wOff(p(64), 1)), wMulNum, p(256)},
		{"mul_2^510", wFrac(maxC, big.NewInt(3)), wFrac(wNeg(maxC), maxC), wMulNum, wMul(maxC, maxC)},
		// Divisor with the top bit of a word set, longer dividends.
		{"divisor_2^64-1", wFrac(wOff(p(200), 7), one), wFrac(wOff(p(64), -1), one), wQuoDivisor, wOff(p(64), -1)},
		{"divisor_2^63+12345", wFrac(wOff(p(130), 3), one), wFrac(wOff(p(63), 12345), one), wQuoDivisor, wOff(p(63), 12345)},
		{"divisor_2^128-1", wFrac(wOff(p(250), 1), one), wFrac(wOff(p(128), -1), one), wQuoDivisor, wOff(p(128), -1)},
		{"divisor_2^127+3", wFrac(wOff(p(190), 5), one), wFrac(wOff(p(127), 3), one), wQuoDivisor, wOff(p(127), 3)},
		{"divisor_one_word_dividend_four_words", wFrac(wOff(maxC, -2), one), wFrac(big.NewInt(1000003), one), wQuoDivisor, big.NewInt(1000003)},
		{"divisor_two_words_dividend_eight_words", wFrac(maxC, one), wFrac(wOff(p(100), 1), maxC), wQuoDivisor, wOff(p(100), 1)},
		// Quotient next to 2^64-1 with a divisor just above 2^64.
		{
			"quotient_2^64-1", wFrac(wOff(wMul(wOff(p(64), -1), wOff(p(64), 1)), 1<<62), one),
			wFrac(wOff(p(64), 1), one), wQuotient, wOff(p(64), -1),
		},
		{
			"quotient_2^64-1_divisor_2^64+3", wFrac(wOff(wMul(wOff(p(64), -1), wOff(p(64), 3)), 1<<62), one),
			wFrac(wOff(p(64), 3), big.NewInt(-1)), wQuotient, wOff(p(64), -1),
		},
		// Dividend below and equal to the divisor.
		{"dividend_below_divisor", wFrac(wOff(p(128), -2), one), wFrac(wOff(p(128), -1), one), wQuotient, big.NewInt(0)},
		{"dividend_equals_divisor", wFrac(wOff(p(128), -1), big.NewInt(-1)), wFrac(wOff(p(128), -1), one), wQuotient, one},
		// Add-back case of long division.
		{"add_back", wFrac(wKnuthU, one), wFrac(wKnuthV, one), wQuotient, new(big.Int).Quo(wKnuthU, wKnuthV)},
		{"add_back_dividend", wFrac(wKnuthU, big.NewInt(5)), wFrac(wKnuthV, big.NewInt(-5)), wQuoDividend, wMul(wKnuthU, big.NewInt(5))},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, 0, wIntermediateValue(c.kind, c.x, c.y).Cmp(c.want), "the case must hit its boundary")
			checkWideBinary(t, c.x, c.y)
			checkWideBinary(t, c.y, c.x)
			checkWideUnary(t, c.x)
			checkWideUnary(t, c.y)
		})
	}
}

// TestWideSingleValueBoundaries covers the single-fraction operations at the
// boundaries of |n| ÷ D and of the reduction NOD(|n|, D).
func TestWideSingleValueBoundaries(t *testing.T) {
	t.Parallel()

	p := wPow2
	one := big.NewInt(1)
	maxC := fraction.MaxComponentBig()
	g64 := wOff(p(64), 13)
	g128 := wOff(p(128), 51)
	cases := []struct {
		name string
		num  *big.Int
		den  *big.Int
		gcd  *big.Int // the expected NOD(|n|, |d|), checked before running the operations
	}{
		{"gcd_1", wOff(p(200), 1), wOff(p(127), 3), nil},
		{"gcd_power_of_two", wMul(p(100), big.NewInt(3)), wMul(p(90), big.NewInt(-5)), p(90)},
		{"gcd_power_of_two_2^64", wMul(p(64), big.NewInt(7)), wMul(p(64), big.NewInt(9)), p(64)},
		{"gcd_large_2^64", wMul(g64, wOff(p(100), 1)), wMul(g64, wOff(p(90), 3)), nil},
		{"gcd_large_2^128", wMul(g128, wOff(p(120), 1)), wNeg(wMul(g128, wOff(p(110), 3))), nil},
		{"gcd_equals_numerator", wOff(p(130), 1), wMul(wOff(p(130), 1), big.NewInt(3)), wOff(p(130), 1)},
		{"gcd_equals_denominator", wMul(wOff(p(70), 1), big.NewInt(-5)), wOff(p(70), 1), wOff(p(70), 1)},
		{"gcd_equals_both", wOff(p(200), 3), wOff(p(200), 3), wOff(p(200), 3)},
		{"zero_numerator", big.NewInt(0), wNeg(wOff(p(192), 1)), wOff(p(192), 1)},
		{"zero_numerator_two_words", big.NewInt(0), wOff(p(127), 9), wOff(p(127), 9)},
		{"max_components", maxC, maxC, maxC},
		{"max_numerator", wNeg(maxC), one, one},
		{"max_denominator", one, wNeg(maxC), one},
		{"n_2^64-1_d_2^63", wOff(p(64), -1), p(63), one},
		{"n_2^64+1_d_2^64-1", wOff(p(64), 1), wOff(p(64), -1), one},
		{"n_2^128+1_d_2^64+1", wOff(p(128), 1), wOff(p(64), 1), one},
		{"n_2^128-1_d_2^64+1", wOff(p(128), -1), wOff(p(64), 1), wOff(p(64), 1)},
		{"n_2^192_d_2^128-1", p(192), wOff(p(128), -1), one},
		{"n_quotient_2^64-1", wOff(wMul(wOff(p(64), -1), wOff(p(64), 1)), 5), wOff(p(64), 1), nil},
		{"n_add_back", wKnuthU, wNeg(wKnuthV), nil},
		{"n_below_d", wOff(p(128), -2), wOff(p(128), -1), nil},
		{"n_equals_d_negative", wOff(p(128), -1), wNeg(wOff(p(128), -1)), wOff(p(128), -1)},
		{"half_exact", wOff(p(200), 1), wMul(big.NewInt(2), wOff(p(200), 1)), wOff(p(200), 1)},
		{"odd_half", wMul(big.NewInt(-3), wOff(p(150), 1)), wMul(big.NewInt(2), wOff(p(150), 1)), wOff(p(150), 1)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if c.gcd != nil {
				require.Equal(t, 0, wGCD(c.num, c.den).Cmp(c.gcd), "the case must have its NOD")
			}
			checkWideUnary(t, wFrac(c.num, c.den))
		})
	}
}

// TestWideReducingConstructors covers NewReduced, FromBigIntReduced and
// FromBigRat at the range boundary, including inputs beyond MaxComponent
// that reduce into the range.
func TestWideReducingConstructors(t *testing.T) {
	t.Parallel()

	p := wPow2
	maxC := fraction.MaxComponentBig()
	g := wOff(p(200), 1)
	cases := []struct {
		name     string
		num, den *big.Int
	}{
		{"beyond_range_reduces_in", wMul(g, wOff(p(200), 3)), wNeg(wMul(g, wOff(p(190), 7)))},
		{"beyond_range_stays_out", wOff(maxC, 2), big.NewInt(3)},
		{"max_over_max", maxC, maxC},
		{"twice_max_over_two", wMul(big.NewInt(2), maxC), big.NewInt(-2)},
		{"twice_max_plus_two_over_two", wOff(wMul(big.NewInt(2), maxC), 2), big.NewInt(2)},
		{"zero_over_huge", big.NewInt(0), wMul(maxC, maxC)},
		{"two_words_gcd", wMul(wOff(p(64), 13), wOff(p(64), -1)), wMul(wOff(p(64), 13), wOff(p(63), 1))},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			want := wantFromBigIntReduced(c.num, c.den)
			got, err := fraction.FromBigIntReduced(c.num, c.den)
			wCheckFrac(t, "FromBigIntReduced", got, err, want)
			wCheckMust(t, "MustFromBigIntReduced", func() fraction.Fraction {
				return fraction.MustFromBigIntReduced(c.num, c.den)
			}, want)
			got, err = fraction.FromBigRat(new(big.Rat).SetFrac(c.num, c.den))
			wCheckFrac(t, "FromBigRat", got, err, want)
		})
	}

	ints := []struct {
		name     string
		num, den int
	}{
		{"min_int_over_min_int", -1 << 63, -1 << 63},
		{"min_int_over_minus_two", -1 << 63, -2},
		{"max_int_over_min_int", 1<<63 - 1, -1 << 63},
		{"two_powers", 1 << 62, 1 << 40},
		{"zero_over_negative", 0, -7},
	}
	for _, c := range ints {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			want := wantFromBigIntReduced(big.NewInt(int64(c.num)), big.NewInt(int64(c.den)))
			got, err := fraction.NewReduced(c.num, c.den)
			wCheckFrac(t, "NewReduced", got, err, want)
		})
	}
}
