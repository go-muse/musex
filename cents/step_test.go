package cents_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nearestStepOracle returns k, the integer nearest to c*n/1200 (half away
// from zero), and the record of the deviation c - k*1200/n.
func nearestStepOracle(c comp, n int64) (*big.Int, comp) {
	k := nearestScaled(c, rat(n, 1200))
	step := new(big.Rat).SetFrac(new(big.Int).Mul(k, big.NewInt(1200)), big.NewInt(n))

	return k, c.sub(bigRatC(step))
}

// assertMustNearest checks MustNearestStep on the inputs of a NearestStep
// call with the results k, dev and err (K-31).
func assertMustNearest(t testing.TB, c cents.Cents, n, k int, dev cents.Cents, err error, msgAndArgs ...any) bool {
	t.Helper()

	return assertMust(t, mustNearest{k, dev}, err, func() mustNearest {
		k, dev := c.MustNearestStep(n)

		return mustNearest{k, dev}
	}, msgAndArgs...)
}

// K-42, G-12: the nearest step of N-EDO and the deviation.
func TestNearestStep(t *testing.T) {
	t.Parallel()

	r32 := noErr(cents.FromRatio(3, 2))(t)

	cases := []struct {
		name string
		c    cents.Cents
		n    int
		k    int
		dev  comp
		dec  string // Decimal(6) of the deviation; empty when not checked
	}{
		// Python decimal, 120 digits: 1200*log2(3/2) - 700.
		{"R(3:2) in 12", r32, 12, 7, ratio(3, 2).sub(ratC(700, 1)), "1.955001"},
		// Python decimal, 120 digits: 1200*log2(3/2) - 31*1200/53.
		{"R(3:2) in 53", r32, 53, 31, ratio(3, 2).sub(ratC(31*1200, 53)), "0.068208"},
		// Python decimal, 120 digits: 1200*log2(5/4) - 400.
		{"R(5:4) in 12", noErr(cents.FromRatio(5, 4))(t), 12, 4, ratio(5, 4).sub(ratC(400, 1)), "-13.686286"},
		{"R(442:440) in 12", noErr(cents.FromRatio(442, 440))(t), 12, 0, ratio(442, 440), ""},
		{"<50> in 12", noErr(cents.FromInt(50))(t), 12, 1, ratC(-50, 1), ""},
		{"<-50> in 12", noErr(cents.FromInt(-50))(t), 12, -1, ratC(50, 1), ""},
		{"<-650> in 12", noErr(cents.FromInt(-650))(t), 12, -7, ratC(50, 1), ""},
		{"0 in 19", cents.Zero(), 19, 0, zeroComp(), ""},
		{"<1200> in 1", cents.Octave(), 1, 1, zeroComp(), ""},
		{"HOLDER in 53", cents.HoldrianComma(), 53, 1, zeroComp(), ""},
		{"<9000000> in 12", noErr(cents.FromInt(9000000))(t), 12, 90000, zeroComp(), ""},
		{"<-9000000> in 12", noErr(cents.FromInt(-9000000))(t), 12, -90000, zeroComp(), ""},
		{"<1200> in MaxInt", cents.Octave(), maxInt, maxInt, zeroComp(), ""},
		{"<-1200> in MaxInt", noErr(cents.FromInt(-1200))(t), maxInt, -maxInt, zeroComp(), ""},
		{"<600> in 1", noErr(cents.FromInt(600))(t), 1, 1, ratC(-600, 1), ""},
		{"<-600> in 1", noErr(cents.FromInt(-600))(t), 1, -1, ratC(600, 1), ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			k, dev, err := tc.c.NearestStep(tc.n)
			require.NoError(t, err)
			assert.Equal(t, tc.k, k)
			assertComp(t, tc.dev, dev)
			assertMustNearest(t, tc.c, tc.n, k, dev, err, "MustNearestStep")

			if tc.dec != "" {
				assert.Equal(t, tc.dec, dev.Decimal(6, 0))
			}
		})
	}

	errorCases := []struct {
		name string
		c    cents.Cents
		n    int
		err  error
	}{
		{"N = 0", r32, 0, cents.ErrInvalidArgument},
		{"N = -12", r32, -12, cents.ErrInvalidArgument},
		{"N = MinInt", r32, minInt, cents.ErrInvalidArgument},
		{"step number beyond int", noErr(cents.FromInt(9000000))(t), maxInt, cents.ErrOutOfRange},
		{"negative step number beyond int", noErr(cents.FromInt(-9000000))(t), maxInt, cents.ErrOutOfRange},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			k, dev, err := tc.c.NearestStep(tc.n)
			assertOnlyError(t, err, tc.err)
			assert.Zero(t, k)
			assert.True(t, dev == cents.Cents{})
			assertPanicsWithErrorIs(t, tc.err, func() { tc.c.MustNearestStep(tc.n) }, "MustNearestStep")
		})
	}

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			for _, n := range []int{1, 12, 19, 53, 1200} {
				k, dev, err := a.v.NearestStep(n)
				require.NoError(t, err, "%s in %d", a.name, n)
				assertMustNearest(t, a.v, n, k, dev, err, "MustNearestStep: %s in %d", a.name, n)

				wantK, wantDev := nearestStepOracle(a.c, int64(n))
				require.True(t, wantK.IsInt64())
				assert.Equal(t, int(wantK.Int64()), k, "%s in %d", a.name, n)
				assertComp(t, wantDev, dev, "%s in %d", a.name, n)

				assertCents(t, a.v, noErr(noErr(cents.Step(k, n))(t, "E(%d, %d)", k, n).Add(dev))(t, "E(k, N) + deviation for %s in %d", a.name, n),
					"E(k, N) + deviation = %s", a.name)

				limit := noErr(cents.FromFraction(600, n))(t, "600/%d", n)
				assert.LessOrEqual(t, dev.Abs().Compare(limit), 0, "|deviation| <= 600/N for %s in %d", a.name, n)
			}
		}
	})

	// The half: exactly between two steps, away from zero, symmetric.
	t.Run("half", func(t *testing.T) {
		t.Parallel()

		for _, n := range []int{1, 12, 19, 53, 1200} {
			half := noErr(cents.FromFraction(600, n))(t)
			k, dev, err := half.NearestStep(n)
			require.NoError(t, err)
			assertMustNearest(t, half, n, k, dev, err, "MustNearestStep: +half in %d", n)
			assert.Equal(t, 1, k, "+half in %d", n)
			assertCents(t, half.Neg(), dev)

			k, dev, err = half.Neg().NearestStep(n)
			require.NoError(t, err)
			assert.Equal(t, -1, k, "-half in %d", n)
			assertCents(t, half, dev)
		}
	})

	// NF-5 (d): the deviation of the values of K-30 and S for N <= 10^6
	// (including the prime 999983) is within the record limit.
	t.Run("minimum without too complex", func(t *testing.T) {
		t.Parallel()

		values := append(setS(t), k30Values(t)...)
		for _, a := range values {
			for _, n := range []int{12, 999983, 1000000} {
				k, dev, err := a.v.NearestStep(n)
				require.NoError(t, err, "%s in %d", a.name, n)
				assertMustNearest(t, a.v, n, k, dev, err, "MustNearestStep: %s in %d", a.name, n)

				wantK, wantDev := nearestStepOracle(a.c, int64(n))
				assert.Equal(t, int(wantK.Int64()), k, "%s in %d", a.name, n)
				assertComp(t, wantDev, dev, "%s in %d", a.name, n)
			}
		}
	})

	// The deviation beyond the record limit (V-1, architecture "Arithmetic"):
	// the reduced denominator of its rational part is above 10^30. The oracle
	// is exact rational arithmetic (Python fractions and big.Rat here).
	t.Run("too complex deviation", func(t *testing.T) {
		t.Parallel()

		micro := noErr(cents.FromFraction(1, 1000000))(t)
		e := noErr(cents.Step(1, 999983))(t)
		small := noErr(micro.Add(e))(t)
		near700 := noErr(cents.Parse("700.000000000000000000000000000001"))(t)
		big8 := noErr(noErr(noErr(cents.FromInt(8999998))(t).Add(micro))(t).Add(e))(t)

		smallC := ratC(1, 1000000).add(stepC(1, 999983))
		cases := []struct {
			name string
			c    cents.Cents
			rec  comp
			n    int
		}{
			// k = 9231214980209 fits in int; the denominator of the deviation
			// is 9223215239530149275811281000000 > 10^30.
			{"<0.000001> + E(1,999983) in MaxInt", small, smallC, maxInt},
			{"<700 + 10^-30> in 999983", near700, ratC(700, 1).add(bigRatC(new(big.Rat).SetFrac(big.NewInt(1), pow10(30)))), 999983},
			// Both: k is beyond int and the denominator is beyond the limit;
			// ErrTooComplex comes before ErrOutOfRange.
			{"<8999998.000001> + E(1,999983) in MaxInt", big8, ratC(8999998, 1).add(smallC), maxInt},
		}

		for _, tc := range cases {
			k, dev, err := tc.c.NearestStep(tc.n)
			assertOnlyError(t, err, cents.ErrTooComplex, tc.name)
			assertPanicsWithErrorIs(t, cents.ErrTooComplex, func() { tc.c.MustNearestStep(tc.n) }, "MustNearestStep: %s", tc.name)
			assert.Zero(t, k, tc.name)
			assert.True(t, dev == cents.Cents{}, tc.name)

			assertComp(t, tc.rec, tc.c, tc.name)

			wantK, wantDev := nearestStepOracle(tc.rec, int64(tc.n))
			assert.True(t, wantDev.tooComplex(), "oracle: %s, k = %s", tc.name, wantK)
		}

		for _, n := range []int{12, 999983, 1000000} {
			_, _, err := small.NearestStep(n)
			require.NoError(t, err, "<0.000001> + E(1,999983) in %d", n)
		}
	})
}
