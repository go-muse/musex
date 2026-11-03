package pitchclass_test

import (
	"fmt"
	"iter"
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// group is one of the three operator groups with its expected elements.
type group struct {
	name  string
	list  func(int) (iter.Seq[pitchclass.Operator], error)
	must  func(int) iter.Seq[pitchclass.Operator]
	specs func(int) []opSpec
	count func(int) (int, error)
}

func groups() []group {
	return []group{
		{
			name:  "Transpositions",
			list:  pitchclass.Transpositions,
			must:  pitchclass.MustTranspositions,
			specs: transpositionSpecs,
			count: pitchclass.CountTranspositions,
		},
		{
			name:  "TranspositionsInversions",
			list:  pitchclass.TranspositionsInversions,
			must:  pitchclass.MustTranspositionsInversions,
			specs: tiSpecs,
			count: pitchclass.CountTranspositionsInversions,
		},
		{
			name:  "Operators",
			list:  pitchclass.Operators,
			must:  pitchclass.MustOperators,
			specs: operatorSpecs,
			count: pitchclass.CountOperators,
		},
	}
}

// K-23, C-27, G-7, E-2: the three groups in their constant order, without
// repetitions, closed under composition and inverse.
func TestOperatorGroups(t *testing.T) {
	t.Parallel()

	for _, g := range groups() {
		for _, size := range []int{1, 2, 3, 7, 12, 19} {
			t.Run(fmt.Sprintf("%s N=%d", g.name, size), func(t *testing.T) {
				t.Parallel()

				specs := g.specs(size)
				want := make([]pitchclass.Operator, 0, len(specs))
				for _, s := range specs {
					want = append(want, newOperator(t, s))
				}

				seq, err := g.list(size)
				require.NoError(t, err)
				assertSeq(t, seq, want, g.name)
				assertSeq(t, g.must(size), want, "Must"+g.name)

				got := slices.Collect(seq)
				seen := map[pitchclass.Operator]bool{}
				for _, f := range got {
					assert.False(t, seen[f], "%v repeats", f)
					seen[f] = true
				}

				for _, f := range got {
					assert.True(t, seen[f.Inverse()], "the inverse of %v is in the group", f)
					if size > 12 {
						continue
					}

					for _, h := range got {
						if fh, err := f.Compose(h); err != nil || !seen[fh] {
							assert.Fail(t, "closure", "%v∘%v = %v (error %v) is not in the group", f, h, fh, err)
						}
					}
				}
			})
		}
	}

	t.Run("sizes", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			size       int
			t, ti, all int
		}{
			{size: 12, t: 12, ti: 24, all: 48},
			{size: 2, t: 2, ti: 2, all: 2},
			{size: 1, t: 1, ti: 1, all: 1},
			{size: 19, t: 19, ti: 38, all: 342},
		}

		for _, tc := range tests {
			for i, want := range []int{tc.t, tc.ti, tc.all} {
				g := groups()[i]
				assert.Len(t, slices.Collect(g.must(tc.size)), want, "%s(%d)", g.name, tc.size)
			}
		}
	})

	t.Run("order at 12", func(t *testing.T) {
		t.Parallel()

		ti := specsOf(slices.Collect(pitchclass.MustTranspositionsInversions(12)))
		require.Len(t, ti, 24)
		for n := range 12 {
			assert.Equal(t, transpositionSpec(n, 12), ti[n], "T%d", n)
			assert.Equal(t, inversionSpec(n, 12), ti[12+n], "I%d", n)
		}

		all := specsOf(slices.Collect(pitchclass.MustOperators(12)))
		require.Len(t, all, 48)
		for i, m := range []int{1, 5, 7, 11} {
			for n := range 12 {
				assert.Equal(t, opSpec{m: m, n: n, size: 12}, all[12*i+n], "m = %d, n = %d", m, n)
			}
		}
	})

	for _, g := range groups() {
		t.Run(g.name+" N=MAX first three", func(t *testing.T) {
			t.Parallel()

			seq, err := g.list(maxInt)
			require.NoError(t, err)

			got := take(seq, 3)
			require.Len(t, got, 3)
			for n, f := range got {
				assertOperator(t, transpositionSpec(n, maxInt), f, fmt.Sprintf("element %d", n))
			}
			assert.Zero(t, callsAfterStop(seq, 3))
		})
	}

	for _, g := range groups() {
		for _, size := range []int{0, -1, minInt} {
			t.Run(fmt.Sprintf("%s N=%d", g.name, size), func(t *testing.T) {
				t.Parallel()

				seq, err := g.list(size)
				assertOnlyError(t, err, pitchclass.ErrInvalidModulus, g.name)
				assertEmptySeq(t, seq, g.name)
				assertPanicsWithErrorIs(t, pitchclass.ErrInvalidModulus, func() { g.must(size) }, "Must"+g.name)
			})
		}
	}
}

// bigFromString parses a decimal number of the test table.
func bigFromString(t testing.TB, s string) *big.Int {
	t.Helper()

	v, ok := new(big.Int).SetString(s, 10)
	require.True(t, ok, "test table number %q", s)

	return v
}

// K-24, C-28, E-7: the number of operators of each group, as int with
// ErrOverflow and exactly as *big.Int.
func TestCountOperators(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		size      int
		t         int
		ti, all   string
		tiOverInt bool
		allOver   bool
	}{
		{name: "N=12", size: 12, t: 12, ti: "24", all: "48"},
		{name: "N=24", size: 24, t: 24, ti: "48", all: "192"},
		{name: "N=19", size: 19, t: 19, ti: "38", all: "342"},
		{name: "N=31", size: 31, t: 31, ti: "62", all: "930"},
		{name: "N=2", size: 2, t: 2, ti: "2", all: "2"},
		{name: "N=1", size: 1, t: 1, ti: "1", all: "1"},
		{name: "N=3", size: 3, t: 3, ti: "6", all: "6"},
		{
			name: "N=MAX", size: maxInt, t: maxInt,
			ti: "18446744073709551614", all: "71139883463926830590344547103944736768",
			tiOverInt: true, allOver: true,
		},
		{
			name: "N=MAX/2", size: maxInt / 2, t: maxInt / 2,
			ti: "9223372036854775806", all: "14178431928629715148775868154390424232", allOver: true,
		},
		{
			name: "N=MAX/2+1", size: maxInt/2 + 1, t: maxInt/2 + 1,
			ti: "9223372036854775808", all: "10633823966279326983230456482242756608", tiOverInt: true, allOver: true,
		},
		{name: "N=1009*1049", size: 1058441, t: 1058441, ti: "2116882", all: "1118120137344"},
		{name: "N=1009*1709", size: 1724381, t: 1724381, ti: "3448762", all: "2968804689984"},
		{name: "N=1009*3881", size: 3915929, t: 3915929, ti: "7831858", all: "15315354956160"},
		{name: "N=1013*1109", size: 1123417, t: 1123417, ti: "2246834", all: "1259682988432"},
		{name: "N=3037000493", size: 3037000493, t: 3037000493, ti: "6074000986", all: "9223371991445242556"},
		{name: "N=3037000507", size: 3037000507, t: 3037000507, ti: "6074001014", all: "9223372076481256542", allOver: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := pitchclass.CountTranspositions(tc.size)
			require.NoError(t, err)
			assert.Equal(t, tc.t, got, "CountTranspositions")
			assert.Equal(t, tc.t, pitchclass.MustCountTranspositions(tc.size), "MustCountTranspositions")

			wantTI := bigFromString(t, tc.ti)
			assertCount(t, wantTI, tc.tiOverInt, tc.size, "CountTranspositionsInversions",
				pitchclass.CountTranspositionsInversions, pitchclass.MustCountTranspositionsInversions)
			assertBigCount(t, wantTI, tc.size, "CountTranspositionsInversionsBig",
				pitchclass.CountTranspositionsInversionsBig, pitchclass.MustCountTranspositionsInversionsBig)

			wantAll := bigFromString(t, tc.all)
			assertCount(t, wantAll, tc.allOver, tc.size, "CountOperators", pitchclass.CountOperators, pitchclass.MustCountOperators)
			assertBigCount(t, wantAll, tc.size, "CountOperatorsBig", pitchclass.CountOperatorsBig, pitchclass.MustCountOperatorsBig)
		})
	}

	t.Run("the table of MAX", func(t *testing.T) {
		t.Parallel()

		product := new(big.Int).Mul(bigInt(maxInt), bigInt(7713001620195508224))
		assert.Equal(t, "71139883463926830590344547103944736768", product.String(), "MAX * phi(MAX)")
		assert.Equal(t, "18446744073709551614", new(big.Int).Mul(bigInt(maxInt), big.NewInt(2)).String(), "2 * MAX")
		require.True(t, bigInt(3037000507).ProbablyPrime(20))
		require.True(t, bigInt(3037000493).ProbablyPrime(20))

		// MAX/2 = 3 * 715827883 * 2147483647, phi = 2 * 715827882 * 2147483646.
		half := new(big.Int).Mul(big.NewInt(3), big.NewInt(715827883))
		half.Mul(half, big.NewInt(2147483647))
		assert.Equal(t, bigInt(maxInt/2).String(), half.String(), "factors of MAX/2")
		for _, p := range []int64{3, 715827883, 2147483647} {
			require.True(t, big.NewInt(p).ProbablyPrime(20), "%d is prime", p)
		}
		phiHalf := new(big.Int).Mul(big.NewInt(2*715827882), big.NewInt(2147483646))
		assert.Equal(t, "14178431928629715148775868154390424232", new(big.Int).Mul(half, phiHalf).String(), "MAX/2 * phi(MAX/2)")
		assert.Equal(t, "10633823966279326983230456482242756608", new(big.Int).Lsh(big.NewInt(1), 123).String(), "2^62 * 2^61")
	})

	for _, size := range []int{0, -1, minInt} {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			for _, f := range []struct {
				name string
				call func(int) (int, error)
				must func(int) int
			}{
				{name: "CountTranspositions", call: pitchclass.CountTranspositions, must: pitchclass.MustCountTranspositions},
				{name: "CountTranspositionsInversions", call: pitchclass.CountTranspositionsInversions, must: pitchclass.MustCountTranspositionsInversions},
				{name: "CountOperators", call: pitchclass.CountOperators, must: pitchclass.MustCountOperators},
			} {
				got, err := f.call(size)
				assertOnlyError(t, err, pitchclass.ErrInvalidModulus, f.name)
				assert.Zero(t, got)
				assertPanicsWithErrorIs(t, pitchclass.ErrInvalidModulus, func() { f.must(size) }, "Must"+f.name)
			}

			for _, f := range []struct {
				name string
				call func(int) (*big.Int, error)
				must func(int) *big.Int
			}{
				{name: "CountTranspositionsInversionsBig", call: pitchclass.CountTranspositionsInversionsBig, must: pitchclass.MustCountTranspositionsInversionsBig},
				{name: "CountOperatorsBig", call: pitchclass.CountOperatorsBig, must: pitchclass.MustCountOperatorsBig},
			} {
				got, err := f.call(size)
				assertOnlyError(t, err, pitchclass.ErrInvalidModulus, f.name)
				assert.Nil(t, got)
				assertPanicsWithErrorIs(t, pitchclass.ErrInvalidModulus, func() { f.must(size) }, "Must"+f.name)
			}
		})
	}
}

// assertCount checks an int count function: the exact value, or
// ErrOverflow with 0 when the value does not fit into int.
func assertCount(t *testing.T, want *big.Int, overflow bool, size int, name string,
	call func(int) (int, error), must func(int) int,
) {
	t.Helper()

	require.Equal(t, !want.IsInt64(), overflow, "test table: %s(%d) overflow", name, size)

	got, err := call(size)
	if overflow {
		assertOnlyError(t, err, pitchclass.ErrOverflow, name)
		assert.Zero(t, got, name)
		assertPanicsWithErrorIs(t, pitchclass.ErrOverflow, func() { must(size) }, "Must"+name)

		return
	}

	require.NoError(t, err, name)
	assert.Equal(t, want.Int64(), int64(got), name)
	assert.Equal(t, want.Int64(), int64(must(size)), "Must"+name)
}

// assertBigCount checks an exact count function.
func assertBigCount(t *testing.T, want *big.Int, size int, name string,
	call func(int) (*big.Int, error), must func(int) *big.Int,
) {
	t.Helper()

	got, err := call(size)
	require.NoError(t, err, name)
	require.NotNil(t, got, name)
	assert.Equal(t, want.String(), got.String(), name)

	m := must(size)
	require.NotNil(t, m, "Must"+name)
	assert.Equal(t, want.String(), m.String(), "Must"+name)
}

// K-24: for N <= 200 the counts agree with the oracle and with the lengths
// of the groups for N <= 40.
func TestCountOperatorsOracle(t *testing.T) {
	t.Parallel()

	for size := 1; size <= 200; size++ {
		ti := 2 * size
		if size <= 2 {
			ti = size
		}
		all := size * totientBrute(size)

		assert.Equal(t, size, res(pitchclass.CountTranspositions(size)).get(t), "CountTranspositions(%d)", size)
		assert.Equal(t, ti, res(pitchclass.CountTranspositionsInversions(size)).get(t), "CountTranspositionsInversions(%d)", size)
		assert.Equal(t, all, res(pitchclass.CountOperators(size)).get(t), "CountOperators(%d)", size)
		assert.Equal(t, bigInt(ti).String(), res(pitchclass.CountTranspositionsInversionsBig(size)).get(t).String(), "CountTranspositionsInversionsBig(%d)", size)
		assert.Equal(t, bigInt(all).String(), res(pitchclass.CountOperatorsBig(size)).get(t).String(), "CountOperatorsBig(%d)", size)

		if size <= 40 {
			assert.Len(t, slices.Collect(res(pitchclass.Transpositions(size)).get(t)), size, "Transpositions(%d)", size)
			assert.Len(t, slices.Collect(res(pitchclass.TranspositionsInversions(size)).get(t)), ti, "TranspositionsInversions(%d)", size)
			assert.Len(t, slices.Collect(res(pitchclass.Operators(size)).get(t)), all, "Operators(%d)", size)
		}
	}
}
