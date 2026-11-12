package tuplet_test

import (
	"errors"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FuzzParse covers K-11 and architecture 4.12 on arbitrary text: no panic,
// the round trip of accepted input, only E-4 or E-5 on rejection, and a
// bounded error text that quotes the excerpt.
func FuzzParse(f *testing.F) {
	for _, seed := range []string{
		"3:2", "6:4", "1:1", "9223372036854775807:1", "9223372036854775808:1", "", " 3:2", "03:2",
		"3:", ":2", "3:2:1", "Tuplet(unset)", "triplet", "\xff", strings.Repeat("9", 70), "3:" + strings.Repeat("9", 70),
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, text string) {
		var (
			got tuplet.Tuplet
			err error
		)

		if !assertNoPanic(t, func() { got, err = tuplet.Parse(text) }) {
			return
		}

		if err == nil {
			require.True(t, got.IsSet())
			assert.Equal(t, text, got.String(), "accepted input is canonical")

			reparsed, err := tuplet.Parse(got.String())
			require.NoError(t, err)
			assert.Equal(t, got, reparsed)

			return
		}

		assert.Equal(t, tuplet.Tuplet{}, got)

		if assert.Error(t, err) {
			isText := errors.Is(err, tuplet.ErrInvalidText)
			isRange := errors.Is(err, tuplet.ErrOutOfRange)
			assert.True(t, isText != isRange, "exactly one of E-4 and E-5")
		}

		checkErrorText(t, err, text, "fuzz")
	})
}

// FuzzNew covers K-1, K-4, KP-16 and architecture 4.6, 4.7 on arbitrary
// counts: creation errors, coefficients by the big.Rat oracle, round trips
// through the coefficients, and agreement of Compare and CompareCoefficient
// with the oracle.
func FuzzNew(f *testing.F) {
	f.Add(3, 2, 6, 4)
	f.Add(1, 1, 2, 3)
	f.Add(math.MaxInt, math.MaxInt-1, math.MaxInt-1, math.MaxInt)
	f.Add(0, 2, 3, -1)
	f.Add(math.MinInt, 1, 1, math.MaxInt)

	f.Fuzz(func(t *testing.T, a1, m1, a2, m2 int) {
		t1, err1 := tuplet.New(a1, m1)
		t2, err2 := tuplet.New(a2, m2)

		if a1 < 1 || m1 < 1 {
			assertErr(t, err1, tuplet.ErrInvalidCount)
			assert.Equal(t, tuplet.Tuplet{}, t1)

			return
		}

		if a2 < 1 || m2 < 1 {
			assertErr(t, err2, tuplet.ErrInvalidCount)
			assert.Equal(t, tuplet.Tuplet{}, t2)

			return
		}

		require.NoError(t, err1)
		require.NoError(t, err2)

		c := t1.MustCoefficient()
		assertPair(t, c, m1, a1)
		assertPair(t, t1.MustReciprocalCoefficient(), a1, m1)
		assertReducedValue(t, t1.MustReducedCoefficient(), ratInts(m1, a1))

		fromCoefficient, err := tuplet.FromCoefficient(c)
		require.NoError(t, err)
		assert.Equal(t, t1, fromCoefficient)

		fromReciprocal, err := tuplet.FromReciprocalCoefficient(t1.MustReciprocalCoefficient())
		require.NoError(t, err)
		assert.Equal(t, t1, fromReciprocal)

		want := ratInts(m1, a1).Cmp(ratInts(m2, a2))

		cmp, err := t1.CompareCoefficient(t2)
		require.NoError(t, err)
		assert.Equal(t, want, cmp)

		equal, err := t1.EqualCoefficient(t2)
		require.NoError(t, err)
		assert.Equal(t, want == 0, equal)

		assert.Equal(t, oracleCompare([2]int{a1, m1}, [2]int{a2, m2}, true, true), sign(t1.Compare(t2)))
		assert.Equal(t, t1 == t2, t1.Equal(t2))
	})
}

// FuzzDefault covers K-3 and rule P-5 on arbitrary counts: the default is a
// member of the series below actual with no member in between, except 2:3 for
// Ternary; there is no default for actual 1 or a member of the series.
func FuzzDefault(f *testing.F) {
	f.Add(3, false)
	f.Add(2, true)
	f.Add(7, true)
	f.Add(1, false)
	f.Add(math.MaxInt, true)
	f.Add(1<<62, false)
	f.Add(0, false)

	f.Fuzz(func(t *testing.T, actual int, ternary bool) {
		division := tuplet.Binary
		if ternary {
			division = tuplet.Ternary
		}

		got, err := tuplet.Default(actual, division)

		if actual < 1 {
			assertErr(t, err, tuplet.ErrInvalidCount)

			return
		}

		if actual == 1 || isSeriesMember(actual, division) {
			assertErr(t, err, tuplet.ErrNoDefault)
			assert.Equal(t, tuplet.Tuplet{}, got)

			return
		}

		require.NoError(t, err)

		gotActual, normal := got.MustPair()
		assert.Equal(t, actual, gotActual)

		if division == tuplet.Ternary && actual == 2 {
			assert.Equal(t, 3, normal)

			return
		}

		assert.True(t, isSeriesMember(normal, division), "%d is a member of the series", normal)
		assert.Less(t, normal, actual)

		if normal <= math.MaxInt/2 {
			assert.GreaterOrEqual(t, 2*normal, actual, "no member of the series between %d and %d", normal, actual)
		}
	})
}

// FuzzChain covers K-9 and architecture 4.10 on chains of two and three
// groups: Flatten and ChainCoefficient agree with the big integer products
// and do not depend on the order of the chain.
func FuzzChain(f *testing.F) {
	f.Add(3, 2, 5, 4, 2, 3, true)
	f.Add(math.MaxInt, 1, 2, 1, 1, 1, false)
	f.Add(1<<31, 3, 1<<31, 5, 7, 11, true)

	f.Fuzz(func(t *testing.T, a1, m1, a2, m2, a3, m3 int, three bool) {
		pairs := [][2]int{{a1, m1}, {a2, m2}}
		if three {
			pairs = append(pairs, [2]int{a3, m3})
		}

		chain := make([]tuplet.Tuplet, 0, len(pairs))
		actualProduct, normalProduct := big.NewInt(1), big.NewInt(1)

		for _, p := range pairs {
			if p[0] < 1 || p[1] < 1 {
				t.Skip("counts below 1 are outside the chain contract")
			}

			chain = append(chain, tuplet.MustNew(p[0], p[1]))
			actualProduct.Mul(actualProduct, bigOf(p[0]))
			normalProduct.Mul(normalProduct, bigOf(p[1]))
		}

		c, err := tuplet.ChainCoefficient(chain...)
		require.NoError(t, err, "three groups always fit the range of common/fraction")
		assertBigPair(t, c, normalProduct, actualProduct)

		flat, flatErr := tuplet.Flatten(chain...)
		if actualProduct.IsInt64() && normalProduct.IsInt64() {
			require.NoError(t, flatErr)
			assertTuplet(t, flat, int(actualProduct.Int64()), int(normalProduct.Int64()))
			assert.Equal(t, c, flat.MustCoefficient())
		} else {
			assertErr(t, flatErr, tuplet.ErrOutOfRange)
		}

		for _, p := range permutations(chain) {
			pc, err := tuplet.ChainCoefficient(p...)
			assert.NoError(t, err)
			assert.Equal(t, c, pc)

			pf, err := tuplet.Flatten(p...)
			assert.Equal(t, flat, pf)
			assert.Equal(t, flatErr == nil, err == nil)
		}
	})
}
