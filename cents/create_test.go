package cents_test

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ctor is a constructor on fixed inputs and its Must variant on the same
// inputs (K-31).
type ctor struct {
	plain func() (cents.Cents, error)
	must  func() cents.Cents
}

// createCase is one call of a constructor with the expected record or error.
type createCase struct {
	name string
	call ctor
	want comp
	err  error
}

func runCreateCases(t *testing.T, cases []createCase) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.name+" "+errName(tc.err), func(t *testing.T) {
			t.Parallel()

			got, err := tc.call.plain()
			assertResult(t, tc.want, tc.err, got, err)
			assertMust(t, got, err, tc.call.must, "the Must variant")
		})
	}
}

// K-2: values from integers and fractions of cents.
func TestFromInt(t *testing.T) {
	t.Parallel()

	for n := -10; n <= 10; n++ {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()

			c, err := cents.FromInt(n)
			require.NoError(t, err)
			assertMust(t, c, err, func() cents.Cents { return cents.MustFromInt(n) }, "MustFromInt")

			num, den, ok := c.Fraction()
			assert.Equal(t, []any{n, 1, true}, []any{num, den, ok})
			assertComp(t, ratC(int64(n), 1), c)
			assertCents(t, noErr(cents.FromFraction(n, 1))(t), c)
		})
	}

	fromInt := func(n int) ctor {
		return ctor{
			plain: func() (cents.Cents, error) { return cents.FromInt(n) },
			must:  func() cents.Cents { return cents.MustFromInt(n) },
		}
	}

	runCreateCases(t, []createCase{
		{name: "max", call: fromInt(9000000), want: ratC(9000000, 1)},
		{name: "min", call: fromInt(-9000000), want: ratC(-9000000, 1)},
		{name: "above max", call: fromInt(9000001), err: cents.ErrOutOfRange},
		{name: "below min", call: fromInt(-9000001), err: cents.ErrOutOfRange},
		{name: "MaxInt", call: fromInt(maxInt), err: cents.ErrOutOfRange},
		{name: "MinInt", call: fromInt(minInt), err: cents.ErrOutOfRange},
	})
}

// K-2 and the edge cases of the architecture for FromFraction.
func TestFromFraction(t *testing.T) {
	t.Parallel()

	fraction := func(num, den int) ctor {
		return ctor{
			plain: func() (cents.Cents, error) { return cents.FromFraction(num, den) },
			must:  func() cents.Cents { return cents.MustFromFraction(num, den) },
		}
	}

	runCreateCases(t, []createCase{
		{name: "2400/2", call: fraction(2400, 2), want: ratC(1200, 1)},
		{name: "1/-2", call: fraction(1, -2), want: ratC(-1, 2)},
		{name: "-2/-4", call: fraction(-2, -4), want: ratC(1, 2)},
		{name: "0/5", call: fraction(0, 5), want: zeroComp()},
		{name: "0/-3", call: fraction(0, -3), want: zeroComp()},
		{name: "1200/19", call: fraction(1200, 19), want: stepC(1, 19)},
		{name: "71/10", call: fraction(71, 10), want: ratC(71, 10)},
		{name: "max", call: fraction(9000000, 1), want: ratC(9000000, 1)},
		{name: "min", call: fraction(-9000000, 1), want: ratC(-9000000, 1)},
		{name: "18000000/2", call: fraction(18000000, 2), want: ratC(9000000, 1)},
		{name: "MinInt/MinInt", call: fraction(minInt, minInt), want: ratC(1, 1)},
		{name: "1/MinInt", call: fraction(1, minInt), want: bigRatC(new(big.Rat).SetFrac(big.NewInt(-1), new(big.Int).Lsh(big.NewInt(1), 63)))},
		{name: "MaxInt/MaxInt", call: fraction(maxInt, maxInt), want: ratC(1, 1)},
		{name: "1/MaxInt", call: fraction(1, maxInt), want: ratC(1, maxInt)},
		{name: "MaxInt/1", call: fraction(maxInt, 1), err: cents.ErrOutOfRange},
		{name: "MinInt/1", call: fraction(minInt, 1), err: cents.ErrOutOfRange},
		{name: "MinInt/-1", call: fraction(minInt, -1), err: cents.ErrOutOfRange},
		{name: "9000001/1", call: fraction(9000001, 1), err: cents.ErrOutOfRange},
		{name: "18000001/2", call: fraction(18000001, 2), err: cents.ErrOutOfRange},
		{name: "1/0", call: fraction(1, 0), err: cents.ErrInvalidArgument},
		{name: "0/0", call: fraction(0, 0), err: cents.ErrInvalidArgument},
		{name: "-5/0", call: fraction(-5, 0), err: cents.ErrInvalidArgument},
		{name: "MinInt/0", call: fraction(minInt, 0), err: cents.ErrInvalidArgument},
		{name: "MaxInt/0", call: fraction(maxInt, 0), err: cents.ErrInvalidArgument},
	})

	t.Run("1200/19 decimal", func(t *testing.T) {
		t.Parallel()

		// K-2: <1200/19> ≈ 63.157895 (6).
		assert.Equal(t, "63.157895", noErr(cents.FromFraction(1200, 19))(t).Decimal(6, 0))
	})
}

// K-4: steps of equal divisions of the octave.
func TestStep(t *testing.T) {
	t.Parallel()

	step := func(k, n int) ctor {
		return ctor{
			plain: func() (cents.Cents, error) { return cents.Step(k, n) },
			must:  func() cents.Cents { return cents.MustStep(k, n) },
		}
	}

	runCreateCases(t, []createCase{
		{name: "E(7,12)", call: step(7, 12), want: ratC(700, 1)},
		{name: "E(0,19)", call: step(0, 19), want: zeroComp()},
		{name: "E(-1,12)", call: step(-1, 12), want: ratC(-100, 1)},
		{name: "E(19,19)", call: step(19, 19), want: ratC(1200, 1)},
		{name: "E(31,31)", call: step(31, 31), want: ratC(1200, 1)},
		{name: "E(1,19)", call: step(1, 19), want: ratC(1200, 19)},
		{name: "E(1,53)", call: step(1, 53), want: ratC(1200, 53)},
		{name: "E(1,1200)", call: step(1, 1200), want: ratC(1, 1)},
		{name: "E(14,12)", call: step(14, 12), want: ratC(1400, 1)},
		{name: "E(1,1)", call: step(1, 1), want: ratC(1200, 1)},
		{name: "E(90000,12)", call: step(90000, 12), want: ratC(9000000, 1)},
		{name: "E(-90000,12)", call: step(-90000, 12), want: ratC(-9000000, 1)},
		{name: "E(MaxInt,MaxInt)", call: step(maxInt, maxInt), want: ratC(1200, 1)},
		{name: "E(1,MaxInt)", call: step(1, maxInt), want: ratC(1200, maxInt)},
		{name: "E(MinInt,MaxInt)", call: step(minInt, maxInt), want: bigRatC(new(big.Rat).SetFrac(
			new(big.Int).Mul(big.NewInt(minInt), big.NewInt(1200)), big.NewInt(maxInt)))},
		{name: "E(7500,1)", call: step(7500, 1), want: ratC(9000000, 1)},
		{name: "E(90001,12)", call: step(90001, 12), err: cents.ErrOutOfRange},
		{name: "E(-90001,12)", call: step(-90001, 12), err: cents.ErrOutOfRange},
		{name: "E(MaxInt,1)", call: step(maxInt, 1), err: cents.ErrOutOfRange},
		{name: "E(MinInt,1)", call: step(minInt, 1), err: cents.ErrOutOfRange},
		{name: "E(7501,1)", call: step(7501, 1), err: cents.ErrOutOfRange},
		{name: "N=0", call: step(1, 0), err: cents.ErrInvalidArgument},
		{name: "N=-12", call: step(1, -12), err: cents.ErrInvalidArgument},
		{name: "N=MinInt", call: step(1, minInt), err: cents.ErrInvalidArgument},
		{name: "k=MaxInt N=0", call: step(maxInt, 0), err: cents.ErrInvalidArgument},
	})

	t.Run("Holdrian comma", func(t *testing.T) {
		t.Parallel()

		assertCents(t, cents.HoldrianComma(), noErr(cents.Step(1, 53))(t))
	})
}

// K-5, G-5: values of integer ratios; primes up to 31 are exact, 37 and more
// give ErrPrimeLimit; 64-bit ratios never give ErrOutOfRange.
func TestFromRatio(t *testing.T) {
	t.Parallel()

	fromRatio := func(a, b int) ctor {
		return ctor{
			plain: func() (cents.Cents, error) { return cents.FromRatio(a, b) },
			must:  func() cents.Cents { return cents.MustFromRatio(a, b) },
		}
	}
	pow := func(b, e int64) *big.Int { return new(big.Int).Exp(big.NewInt(b), big.NewInt(e), nil) }
	r62over39, _ := ratioComp(pow(2, 62), pow(3, 39))

	runCreateCases(t, []createCase{
		{name: "3:2", call: fromRatio(3, 2), want: ratio(3, 2)},
		{name: "6:4", call: fromRatio(6, 4), want: ratio(3, 2)},
		{name: "111:74", call: fromRatio(111, 74), want: ratio(3, 2)},
		{name: "37:37", call: fromRatio(37, 37), want: zeroComp()},
		{name: "1:1", call: fromRatio(1, 1), want: zeroComp()},
		{name: "1:2", call: fromRatio(1, 2), want: ratC(-1200, 1)},
		{name: "2:1", call: fromRatio(2, 1), want: ratC(1200, 1)},
		{name: "5:4", call: fromRatio(5, 4), want: ratio(5, 4)},
		{name: "442:440", call: fromRatio(442, 440), want: ratio(221, 220)},
		{name: "31:16", call: fromRatio(31, 16), want: ratio(31, 16)},
		{name: "2^62:1", call: fromRatio(1<<62, 1), want: ratC(74400, 1)},
		{name: "1:2^62", call: fromRatio(1, 1<<62), want: ratC(-74400, 1)},
		{name: "2^62:3^39", call: fromRatio(1<<62, int(pow(3, 39).Int64())), want: r62over39},
		{name: "37:32", call: fromRatio(37, 32), err: cents.ErrPrimeLimit},
		{name: "415:440", call: fromRatio(415, 440), err: cents.ErrPrimeLimit},
		{name: "MaxInt:1", call: fromRatio(maxInt, 1), err: cents.ErrPrimeLimit},
		{name: "0:1", call: fromRatio(0, 1), err: cents.ErrInvalidArgument},
		{name: "-3:2", call: fromRatio(-3, 2), err: cents.ErrInvalidArgument},
		{name: "3:0", call: fromRatio(3, 0), err: cents.ErrInvalidArgument},
		{name: "3:-2", call: fromRatio(3, -2), err: cents.ErrInvalidArgument},
		{name: "MinInt:1", call: fromRatio(minInt, 1), err: cents.ErrInvalidArgument},
		{name: "0:37 invalid before prime", call: fromRatio(0, 37), err: cents.ErrInvalidArgument},
		{name: "-37:32 invalid before prime", call: fromRatio(-37, 32), err: cents.ErrInvalidArgument},
	})

	t.Run("decimal views", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			a, b   int
			digits int
			want   string
		}{
			{3, 2, 6, "701.955001"},
			{3, 2, 15, "701.955000865387418"},
			{5, 4, 6, "386.313714"},
			{442, 440, 6, "7.851415"},
			{31, 16, 6, "1145.035572"},
			{1 << 62, int(pow(3, 39).Int64()), 6, "223.754966"},
		}

		for _, tc := range cases {
			assert.Equal(t, tc.want, noErr(cents.FromRatio(tc.a, tc.b))(t).Decimal(tc.digits, 0), "R(%d:%d)", tc.a, tc.b)
		}
	})

	t.Run("6:4 equals 3:2", func(t *testing.T) {
		t.Parallel()

		assertCents(t, noErr(cents.FromRatio(3, 2))(t), noErr(cents.FromRatio(6, 4))(t))
		assertCents(t, noErr(cents.FromRatio(3, 2))(t), noErr(cents.FromRatio(111, 74))(t))
	})

	// Every pair of 64-bit numbers: the record by factoring, or ErrPrimeLimit
	// when a prime greater than 31 remains after reducing; never out of range.
	t.Run("64-bit pairs", func(t *testing.T) {
		t.Parallel()

		numbers := []int{
			1, 2, 3, 5, 6, 7, 30, 64, 80, 81, 37 * 3, 37 * 2, 999999, 1000003,
			1 << 62, int(pow(3, 39).Int64()), int(pow(31, 12).Int64()), int(pow(29, 12).Int64()),
			3 << 61, maxInt, maxInt - 1, 2 * 3 * 5 * 7 * 11 * 13 * 17 * 19 * 23 * 29 * 31,
		}

		for _, a := range numbers {
			for _, b := range numbers {
				want, ok := ratioComp(big.NewInt(int64(a)), big.NewInt(int64(b)))
				var wantErr error
				if !ok {
					wantErr = cents.ErrPrimeLimit
				}

				got, err := cents.FromRatio(a, b)
				assertResult(t, want, wantErr, got, err, fmt.Sprintf("R(%d:%d)", a, b))
				assertMust(t, got, err, func() cents.Cents { return cents.MustFromRatio(a, b) }, "MustFromRatio(%d, %d)", a, b)
			}
		}
	})
}
