package pitchclass_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-15, C-18, G-4, G-9: Mm(x) = (m*x) mod N for any int m without overflow;
// a non-invertible m merges classes and is not an error.
func TestClassMultiply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		x, m, size int
		want       int
	}{
		{name: "M7(1) at 12", x: 1, m: 7, size: 12, want: 7},
		{name: "M7(2) at 12", x: 2, m: 7, size: 12, want: 2},
		{name: "M5(1) at 12", x: 1, m: 5, size: 12, want: 5},
		{name: "M11(4) at 12", x: 4, m: 11, size: 12, want: 8},
		{name: "M-1(4) at 12", x: 4, m: -1, size: 12, want: 8},
		{name: "M2(6) at 12", x: 6, m: 2, size: 12, want: 0},
		{name: "M2(0) at 12", x: 0, m: 2, size: 12, want: 0},
		{name: "M0(5) at 12", x: 5, m: 0, size: 12, want: 0},
		{name: "MMAX(11) at 12", x: 11, m: maxInt, size: 12, want: 5},
		{name: "MMIN(1) at 12", x: 1, m: minInt, size: 12, want: 4},
		{name: "M2(MAX-1) at MAX", x: maxInt - 1, m: 2, size: maxInt, want: maxInt - 2},
		{name: "M(MAX-1)(MAX-1) at MAX", x: maxInt - 1, m: maxInt - 1, size: maxInt, want: 1},
		{name: "MMIN(1) at MAX", x: 1, m: minInt, size: maxInt, want: maxInt - 1},
		{name: "MMIN(MAX-1) at MAX", x: maxInt - 1, m: minInt, size: maxInt, want: 1},
		{name: "MMAX(MAX-1) at MAX", x: maxInt - 1, m: maxInt, size: maxInt, want: 0},
		{name: "M5(0) at 1", x: 0, m: 5, size: 1, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, mulMod(tc.m, tc.x, tc.size), tc.want, "test table disagrees with the oracle")
			assertClass(t, tc.want, tc.size, newClass(t, tc.x, tc.size).Multiply(tc.m), "Multiply")
		})
	}

	t.Run("M11 is I0 and M1 is the identity at 12", func(t *testing.T) {
		t.Parallel()

		for _, x := range classesOf(t, 12) {
			assert.True(t, x.Multiply(11).Equal(x.Invert(0)), "M11(%v) = I0(%v)", x, x)
			assert.True(t, x.Multiply(1).Equal(x), "M1(%v)", x)
			assertClass(t, 0, 12, x.Multiply(0), fmt.Sprintf("M0(%v)", x))
		}
	})

	t.Run("distinct images at 12", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct{ m, distinct int }{{5, 12}, {7, 12}, {2, 6}, {11, 12}, {3, 4}, {0, 1}} {
			images := map[pitchclass.Class]bool{}
			for _, x := range classesOf(t, 12) {
				images[x.Multiply(tc.m)] = true
			}
			assert.Len(t, images, tc.distinct, "M%d at 12", tc.m)
		}
	})

	for _, size := range []int{1, 2, 12, 19} {
		t.Run(fmt.Sprintf("oracle N=%d", size), func(t *testing.T) {
			t.Parallel()

			for _, x := range classesOf(t, size) {
				for _, m := range append(ints(-30, 30), edgeInts()...) {
					assertClass(t, mulMod(m, x.Number(), size), size, x.Multiply(m), fmt.Sprintf("%v.Multiply(%d)", x, m))
				}
			}
		})
	}

	t.Run("oracle N=MAX", func(t *testing.T) {
		t.Parallel()

		for _, k := range []int{0, 1, 2, maxInt / 2, maxInt - 2, maxInt - 1} {
			x := newClass(t, k, maxInt)
			for _, m := range edgeInts() {
				assertClass(t, mulMod(m, k, maxInt), maxInt, x.Multiply(m), fmt.Sprintf("%v.Multiply(%d)", x, m))
			}
		}
	})
}

// K-16, C-19, E-2, E-4: invertibility of m and the inverse multiplier.
func TestInvertibleAndInverseMultiplier(t *testing.T) {
	t.Parallel()

	invertible := []struct {
		name          string
		m, size, want int
	}{
		{name: "1 at 12", m: 1, size: 12, want: 1},
		{name: "5 at 12", m: 5, size: 12, want: 5},
		{name: "7 at 12", m: 7, size: 12, want: 7},
		{name: "11 at 12", m: 11, size: 12, want: 11},
		{name: "-1 at 12", m: -1, size: 12, want: 11},
		{name: "17 at 12", m: 17, size: 12, want: 5},
		{name: "2 at 19", m: 2, size: 19, want: 10},
		{name: "0 at 1", m: 0, size: 1, want: 0},
		{name: "5 at 1", m: 5, size: 1, want: 0},
		{name: "-1 at 1", m: -1, size: 1, want: 0},
		{name: "MIN at 1", m: minInt, size: 1, want: 0},
		{name: "MAX at 1", m: maxInt, size: 1, want: 0},
		{name: "1 at 2", m: 1, size: 2, want: 1},
		{name: "MAX-1 at MAX", m: maxInt - 1, size: maxInt, want: maxInt - 1},
		{name: "2 at MAX", m: 2, size: maxInt, want: 4611686018427387904},
		{name: "MIN at MAX", m: minInt, size: maxInt, want: maxInt - 1},
	}

	for _, tc := range invertible {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.True(t, invertibleOracle(tc.m, tc.size), "test table disagrees with the oracle")
			require.Equal(t, inverseOracle(tc.m, tc.size), tc.want, "test table disagrees with the oracle")

			ok, err := pitchclass.Invertible(tc.m, tc.size)
			require.NoError(t, err)
			assert.True(t, ok)
			assert.True(t, pitchclass.MustInvertible(tc.m, tc.size))

			inv, err := pitchclass.InverseMultiplier(tc.m, tc.size)
			require.NoError(t, err)
			assert.Equal(t, tc.want, inv)
			assert.Equal(t, tc.want, pitchclass.MustInverseMultiplier(tc.m, tc.size))
		})
	}

	notInvertible := []struct {
		name    string
		m, size int
	}{
		{name: "0 at 12", m: 0, size: 12},
		{name: "2 at 12", m: 2, size: 12},
		{name: "3 at 12", m: 3, size: 12},
		{name: "4 at 12", m: 4, size: 12},
		{name: "6 at 12", m: 6, size: 12},
		{name: "8 at 12", m: 8, size: 12},
		{name: "9 at 12", m: 9, size: 12},
		{name: "10 at 12", m: 10, size: 12},
		{name: "12 at 12", m: 12, size: 12},
		{name: "MIN at 12", m: minInt, size: 12},
		{name: "0 at 2", m: 0, size: 2},
		{name: "0 at 19", m: 0, size: 19},
		{name: "7 at MAX", m: 7, size: maxInt},
		{name: "0 at MAX", m: 0, size: maxInt},
		{name: "MAX at MAX", m: maxInt, size: maxInt},
	}

	for _, tc := range notInvertible {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.False(t, invertibleOracle(tc.m, tc.size), "test table disagrees with the oracle")

			ok, err := pitchclass.Invertible(tc.m, tc.size)
			require.NoError(t, err)
			assert.False(t, ok)
			assert.False(t, pitchclass.MustInvertible(tc.m, tc.size))

			inv, err := pitchclass.InverseMultiplier(tc.m, tc.size)
			assertOnlyError(t, err, pitchclass.ErrNotInvertible, "InverseMultiplier")
			assert.Zero(t, inv)
			assertPanicsWithErrorIs(t, pitchclass.ErrNotInvertible, func() { pitchclass.MustInverseMultiplier(tc.m, tc.size) }, "MustInverseMultiplier")
		})
	}

	for _, size := range []int{0, -1, minInt} {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			for _, m := range []int{1, 2, 0, minInt} {
				desc := fmt.Sprintf("(%d, %d)", m, size)

				ok, err := pitchclass.Invertible(m, size)
				assertOnlyError(t, err, pitchclass.ErrInvalidModulus, "Invertible"+desc)
				assert.False(t, ok)

				inv, err := pitchclass.InverseMultiplier(m, size)
				assertOnlyError(t, err, pitchclass.ErrInvalidModulus, "InverseMultiplier"+desc)
				assert.Zero(t, inv)

				assertPanicsWithErrorIs(t, pitchclass.ErrInvalidModulus, func() { pitchclass.MustInvertible(m, size) }, "MustInvertible"+desc)
				assertPanicsWithErrorIs(t, pitchclass.ErrInvalidModulus, func() { pitchclass.MustInverseMultiplier(m, size) }, "MustInverseMultiplier"+desc)
			}
		})
	}
}

// K-16: for N in 1...60 and m in -30...30 invertibility and the inverse
// agree with the oracle: (m * m^-1) mod N = 1 mod N.
func TestInverseMultiplierOracle(t *testing.T) {
	t.Parallel()

	for size := 1; size <= 60; size++ {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			for _, m := range append(ints(-30, 30), edgeInts()...) {
				desc := fmt.Sprintf("(%d, %d)", m, size)
				want := invertibleOracle(m, size)

				ok, err := pitchclass.Invertible(m, size)
				require.NoError(t, err, "Invertible"+desc)
				assert.Equal(t, want, ok, "Invertible"+desc)

				inv, err := pitchclass.InverseMultiplier(m, size)
				if !want {
					assertOnlyError(t, err, pitchclass.ErrNotInvertible, "InverseMultiplier"+desc)

					continue
				}

				require.NoError(t, err, "InverseMultiplier"+desc)
				assert.Equal(t, inverseOracle(m, size), inv, "InverseMultiplier"+desc)
				assert.True(t, inv >= 0 && inv < size, "InverseMultiplier%s = %d in 0...N-1", desc, inv)
				assert.Equal(t, mod(1, size), mulMod(m, inv, size), "m * m^-1 for %s", desc)
			}
		})
	}
}

// K-17, C-20: the invertible multipliers in ascending order and phi(N).
func TestInvertibleMultipliersAndTotient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		size int
		want []int
	}{
		{size: 12, want: []int{1, 5, 7, 11}},
		{size: 24, want: []int{1, 5, 7, 11, 13, 17, 19, 23}},
		{size: 19, want: ints(1, 18)},
		{size: 31, want: ints(1, 30)},
		{size: 1, want: []int{0}},
		{size: 2, want: []int{1}},
		{size: 4, want: []int{1, 3}},
		{size: 7, want: ints(1, 6)},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("N=%d", tc.size), func(t *testing.T) {
			t.Parallel()

			require.Equal(t, invertibleList(tc.size), tc.want, "test table disagrees with the oracle")

			seq, err := pitchclass.InvertibleMultipliers(tc.size)
			require.NoError(t, err)
			assertSeq(t, seq, tc.want, "InvertibleMultipliers")
			assertSeq(t, pitchclass.MustInvertibleMultipliers(tc.size), tc.want, "MustInvertibleMultipliers")

			phi, err := pitchclass.Totient(tc.size)
			require.NoError(t, err)
			assert.Equal(t, len(tc.want), phi)
			assert.Equal(t, len(tc.want), pitchclass.MustTotient(tc.size))
		})
	}

	t.Run("N=MAX first multipliers", func(t *testing.T) {
		t.Parallel()

		seq, err := pitchclass.InvertibleMultipliers(maxInt)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 8}, take(seq, 7))
		assert.Zero(t, callsAfterStop(seq, 7))
	})

	for _, size := range []int{0, -1, minInt} {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			seq, err := pitchclass.InvertibleMultipliers(size)
			assertOnlyError(t, err, pitchclass.ErrInvalidModulus, "InvertibleMultipliers")
			assertEmptySeq(t, seq, "InvertibleMultipliers")

			phi, err := pitchclass.Totient(size)
			assertOnlyError(t, err, pitchclass.ErrInvalidModulus, "Totient")
			assert.Zero(t, phi)

			assertPanicsWithErrorIs(t, pitchclass.ErrInvalidModulus, func() { pitchclass.MustInvertibleMultipliers(size) }, "MustInvertibleMultipliers")
			assertPanicsWithErrorIs(t, pitchclass.ErrInvalidModulus, func() { pitchclass.MustTotient(size) }, "MustTotient")
		})
	}
}

// K-17: for all N <= 2000 Totient agrees with the brute-force oracle; for
// N <= 200 InvertibleMultipliers yields exactly the invertible m.
func TestTotientOracle(t *testing.T) {
	t.Parallel()

	for size := 1; size <= 2000; size++ {
		phi, err := pitchclass.Totient(size)
		require.NoError(t, err, "Totient(%d)", size)
		assert.Equal(t, totientBrute(size), phi, "Totient(%d)", size)
	}

	for size := 1; size <= 200; size++ {
		seq, err := pitchclass.InvertibleMultipliers(size)
		require.NoError(t, err)
		assert.Equal(t, invertibleList(size), take(seq, size+1), "InvertibleMultipliers(%d)", size)
	}
}

// factor is a prime power p^e of a factorization in the test table.
type factor struct {
	p *big.Int
	e int
}

func pow(p, e int) factor {
	return factor{p: bigInt(p), e: e}
}

// totientCase is a modulus with its known factorization; the test checks
// that the factors multiply to N before it trusts phi derived from them.
type totientCase struct {
	name    string
	n       int
	factors []factor
	want    int
}

// hardTotients returns the moduli named in the architecture (Totient,
// "Boundary cases"): large primes, semiprimes near 2^63, a prime square,
// powers of 2 and 3, a strong pseudoprime to small bases and MAX.
func hardTotients() []totientCase {
	return []totientCase{
		{name: "MAX", n: maxInt, factors: []factor{pow(7, 2), pow(73, 1), pow(127, 1), pow(337, 1), pow(92737, 1), pow(649657, 1)}, want: 7713001620195508224},
		{name: "largest prime below 2^63", n: 9223372036854775783, factors: []factor{pow(9223372036854775783, 1)}, want: 9223372036854775782},
		{name: "3037000493 * 3037000453", n: 9223371873002223329, factors: []factor{pow(3037000493, 1), pow(3037000453, 1)}, want: 9223371866928222384},
		{name: "3037000493^2", n: 9223371994482243049, factors: []factor{pow(3037000493, 2)}, want: 9223371991445242556},
		{name: "2^62", n: 1 << 62, factors: []factor{pow(2, 62)}, want: 1 << 61},
		{name: "4", n: 4, factors: []factor{pow(2, 2)}, want: 2},
		{name: "8", n: 8, factors: []factor{pow(2, 3)}, want: 4},
		{name: "strong pseudoprime 3825123056546413051", n: 3825123056546413051, factors: []factor{pow(149491, 1), pow(747451, 1), pow(34233211, 1)}, want: 3825092239639605000},
		{name: "3215031751", n: 3215031751, factors: []factor{pow(151, 1), pow(751, 1), pow(28351, 1)}, want: 3189375000},
		{name: "3^39", n: 4052555153018976267, factors: []factor{pow(3, 39)}, want: 2701703435345984178},
		{name: "prime 3037000507", n: 3037000507, factors: []factor{pow(3037000507, 1)}, want: 3037000506},
		{name: "prime 2^31-1", n: 1<<31 - 1, factors: []factor{pow(1<<31-1, 1)}, want: 1<<31 - 2},
		// Semiprimes with both factors just above 1000 (QC-1): after small
		// factors are removed by trial division, the rho method must split
		// them, including the retry when an attempt fails (architecture,
		// Totient, required part 2; risk R-PC8).
		{name: "1009 * 1049", n: 1058441, factors: []factor{pow(1009, 1), pow(1049, 1)}, want: 1056384},
		{name: "1009 * 1709", n: 1724381, factors: []factor{pow(1009, 1), pow(1709, 1)}, want: 1721664},
		{name: "1009 * 3881", n: 3915929, factors: []factor{pow(1009, 1), pow(3881, 1)}, want: 3911040},
		{name: "1013 * 1109", n: 1123417, factors: []factor{pow(1013, 1), pow(1109, 1)}, want: 1121296},
	}
}

// checkFactorization verifies the test table: the prime powers multiply to
// n and each p is prime. It returns phi(n) computed from the factors.
func checkFactorization(t *testing.T, tc totientCase) int {
	t.Helper()

	product := big.NewInt(1)
	phi := big.NewInt(1)
	one := big.NewInt(1)
	for _, f := range tc.factors {
		require.True(t, f.p.ProbablyPrime(20), "%v is prime", f.p)

		pe := new(big.Int).Exp(f.p, big.NewInt(int64(f.e)), nil)
		product.Mul(product, pe)

		pe1 := new(big.Int).Exp(f.p, big.NewInt(int64(f.e-1)), nil)
		phi.Mul(phi, pe1)
		phi.Mul(phi, new(big.Int).Sub(f.p, one))
	}
	require.Equal(t, 0, product.Cmp(bigInt(tc.n)), "factors of %d", tc.n)
	require.True(t, phi.IsInt64())

	return int(phi.Int64())
}

// K-17, D-6: Totient of the hard moduli is exact.
func TestTotientHardModuli(t *testing.T) {
	t.Parallel()

	for _, tc := range hardTotients() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, checkFactorization(t, tc), tc.want, "test table disagrees with its factorization")

			phi, err := pitchclass.Totient(tc.n)
			require.NoError(t, err)
			assert.Equal(t, tc.want, phi)
			assert.Equal(t, tc.want, pitchclass.MustTotient(tc.n))

			again, err := pitchclass.Totient(tc.n)
			require.NoError(t, err)
			assert.Equal(t, phi, again, "the same result on every call")
		})
	}
}
