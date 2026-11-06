package cents_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-19: Float64 is the nearest float64 (ties to even).
func TestFloat64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		c    cents.Cents
		want float64
	}{
		// 1200*log2(3/2): Python decimal, 120 digits, nearest float64.
		{"R(3:2)", noErr(cents.FromRatio(3, 2))(t), 701.9550008653874},
		{"<9000000>", noErr(cents.FromInt(9000000))(t), 9e6},
		{"<-9000000>", noErr(cents.FromInt(-9000000))(t), -9e6},
		{"<7.1>", noErr(cents.FromFraction(71, 10))(t), 7.1},
		{"<1/3>", noErr(cents.FromFraction(1, 3))(t), 1.0 / 3.0},
		// Exact ties of rational values (Python fractions): down and up to
		// the even mantissa, symmetric for negative values.
		{"(2^53+1)/2^30", noErr(cents.FromFraction(1<<53+1, 1<<30))(t), 8388608},
		{"(2^53+3)/2^30", noErr(cents.FromFraction(1<<53+3, 1<<30))(t), 0x1.0000000000002p+23},
		{"-(2^53+1)/2^30", noErr(cents.FromFraction(-(1<<53 + 1), 1<<30))(t), -8388608},
		{"-(2^53+3)/2^30", noErr(cents.FromFraction(-(1<<53 + 3), 1<<30))(t), -0x1.0000000000002p+23},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.c.Float64(), "%x", tc.c.Float64())
		})
	}

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			got := a.v.Float64()
			assert.Equal(t, floatOracle(a.c), got, "%s: %x", a.name, got)
			assert.False(t, math.Signbit(got) && got == 0, "no -0")
		}
	})
}

// K-19: the exact fraction of cents of a rational value.
func TestFraction(t *testing.T) {
	t.Parallel()

	type result struct {
		Num, Den int
		OK       bool
	}

	set := setS(t)
	cases := []struct {
		name string
		c    cents.Cents
		want result
	}{
		{"E(1,19)", noErr(cents.Step(1, 19))(t), result{1200, 19, true}},
		{"<7.1>", noErr(cents.FromFraction(71, 10))(t), result{71, 10, true}},
		{"0", cents.Zero(), result{0, 1, true}},
		{"<-13.7>", noErr(cents.FromFraction(-137, 10))(t), result{-137, 10, true}},
		{"<-1/MaxInt>", noErr(cents.FromFraction(-1, maxInt))(t), result{-1, maxInt, true}},
		{"R(3:2)", noErr(cents.FromRatio(3, 2))(t), result{0, 0, false}},
		{"Q", byName(t, set, "Q").v, result{0, 0, false}},
		{"SAV", cents.Savart(), result{0, 0, false}},
		// The denominator 2^63 does not fit in int.
		{"<1/MinInt>", noErr(cents.FromFraction(1, minInt))(t), result{0, 0, false}},
		{"<1/10^30>", noErr(cents.Parse("0.000000000000000000000000000001"))(t), result{0, 0, false}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			num, den, ok := tc.c.Fraction()
			assert.Equal(t, tc.want, result{num, den, ok})
		})
	}

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			num, den, ok := a.v.Fraction()
			want := result{}
			if a.c.rational() && a.c.r.Num().IsInt64() && a.c.r.Denom().IsInt64() {
				want = result{int(a.c.r.Num().Int64()), int(a.c.r.Denom().Int64()), true}
			}

			assert.Equal(t, want, result{num, den, ok}, a.name)
		}
	})
}

// K-19: the exact ratio of integers of a value that is the cents of a ratio.
func TestRatio(t *testing.T) {
	t.Parallel()

	type result struct {
		Num, Den int
		OK       bool
	}

	set := setS(t)
	cases := []struct {
		name string
		c    cents.Cents
		want result
	}{
		{"R(3:2)", noErr(cents.FromRatio(3, 2))(t), result{3, 2, true}},
		{"<1200>", cents.Octave(), result{2, 1, true}},
		{"0", cents.Zero(), result{1, 1, true}},
		{"<-1200>", noErr(cents.FromInt(-1200))(t), result{1, 2, true}},
		{"R(2^62:3^39)", noErr(cents.FromRatio(1<<62, 4052555153018976267))(t), result{1 << 62, 4052555153018976267, true}},
		{"<700>", noErr(cents.FromInt(700))(t), result{0, 0, false}},
		{"Q", byName(t, set, "Q").v, result{0, 0, false}},
		{"MK", byName(t, set, "MK").v, result{0, 0, false}},
		{"<1200>*63", noErr(cents.Octave().Mul(63))(t), result{0, 0, false}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			num, den, ok := tc.c.Ratio()
			assert.Equal(t, tc.want, result{num, den, ok})
		})
	}

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			num, den, ok := a.v.Ratio()
			want := result{}
			if n, d, isRatio := a.c.ratioParts(); isRatio && n.IsInt64() && d.IsInt64() {
				want = result{int(n.Int64()), int(d.Int64()), true}
			}

			assert.Equal(t, want, result{num, den, ok}, a.name)
		}
	})
}

// assertCompositionShape checks the form of Composition(): non-nil reduced
// r with den >= 1; exponents of the odd primes in increasing order, without
// base 2 and zero numerators, each reduced with Den >= 1.
func assertCompositionShape(t testing.TB, c cents.Cents, msgAndArgs ...any) {
	t.Helper()

	num, den, ps := c.Composition()
	require.NotNil(t, num, msgAndArgs...)
	require.NotNil(t, den, msgAndArgs...)
	assert.Equal(t, 1, den.Sign(), msgAndArgs...)
	if num.Sign() == 0 {
		assert.Equal(t, 0, den.Cmp(big.NewInt(1)), "0 is 0/1")
	} else {
		assert.Equal(t, 0, new(big.Int).GCD(nil, nil, new(big.Int).Abs(num), den).Cmp(big.NewInt(1)), msgAndArgs...)
	}

	prev := 2
	for _, p := range ps {
		assert.Greater(t, p.Base, prev, msgAndArgs...)
		prev = p.Base

		assert.GreaterOrEqual(t, primeIndex(p.Base), 0, msgAndArgs...)
		require.NotNil(t, p.Num, msgAndArgs...)
		require.NotNil(t, p.Den, msgAndArgs...)
		assert.NotEqual(t, 0, p.Num.Sign(), msgAndArgs...)
		assert.Equal(t, 1, p.Den.Sign(), msgAndArgs...)
		assert.Equal(t, 0, new(big.Int).GCD(nil, nil, new(big.Int).Abs(p.Num), p.Den).Cmp(big.NewInt(1)), msgAndArgs...)
	}
}

// K-19: the exact composition (r; e_p) of any value.
func TestComposition(t *testing.T) {
	t.Parallel()

	set := setS(t)
	cases := []struct {
		name string
		c    cents.Cents
		want comp
	}{
		{"R(3:2)", noErr(cents.FromRatio(3, 2))(t), ratC(-1200, 1).withExp(3, rat(1, 1))},
		{"Q", byName(t, set, "Q").v, zeroComp().withExp(5, rat(1, 4))},
		{"PYTH", cents.PythagoreanComma(), ratC(-22800, 1).withExp(3, rat(12, 1))},
		{"<7.1>", noErr(cents.FromFraction(71, 10))(t), ratC(71, 10)},
		{"SAV", cents.Savart(), ratC(6, 5).withExp(5, rat(1, 1000))},
		{"MK", byName(t, set, "MK").v, ratC(-100800, 1).withExp(3, rat(53, 1))},
		{"0", cents.Zero(), zeroComp()},
		{"R(442:440)", noErr(cents.FromRatio(442, 440))(t),
			ratC(-2400, 1).withExp(5, rat(-1, 1)).withExp(11, rat(-1, 1)).withExp(13, rat(1, 1)).withExp(17, rat(1, 1))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assertComp(t, tc.want, tc.c)
			assertCompositionShape(t, tc.c)
		})
	}

	t.Run("MK has no int ratio but a ratio text", func(t *testing.T) {
		t.Parallel()

		mk := byName(t, set, "MK").v
		_, _, ok := mk.Ratio()
		assert.False(t, ok)

		s, ok := mk.TextAs(cents.KindRatio, 0)
		assert.True(t, ok)
		assert.Equal(t, "19383245667680019896796723:19342813113834066795298816", s)
	})

	t.Run("for all S and the named values", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			assertComp(t, a.c, a.v, a.name)
			assertCompositionShape(t, a.v, a.name)
		}

		for _, nv := range append(cents.NamedValues(), cents.NamedUnits()...) {
			assertCompositionShape(t, nv.Value, nv.Name)
		}
	})

	t.Run("outputs are new", func(t *testing.T) {
		t.Parallel()

		c := cents.PythagoreanComma()
		num, den, ps := c.Composition()
		num.SetInt64(99)
		den.SetInt64(7)
		ps[0].Num.SetInt64(5)
		ps[0].Den.SetInt64(3)
		ps[0] = cents.Power{Base: 31, Num: big.NewInt(1)}

		assertComp(t, ratC(-22800, 1).withExp(3, rat(12, 1)), c, "the value is unchanged")
		assertComp(t, ratC(-22800, 1).withExp(3, rat(12, 1)), cents.PythagoreanComma(), "the next call is unchanged")
	})
}
