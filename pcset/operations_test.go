package pcset_test

import (
	"testing"

	"github.com/go-muse/musex/pcset"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetOperationsAndLiteralInclusion(t *testing.T) {
	t.Parallel()

	a := pcset.MustNew(12, 0, 4, 7)
	b := pcset.MustNew(12, 4, 7, 11)

	t.Run("k14_canonical_examples_mod_12", func(t *testing.T) {
		t.Parallel()

		compA := a.Complement()
		assertSet(t, compA, 12, []int{1, 2, 3, 5, 6, 8, 9, 10, 11})
		assert.Equal(t, a, compA.Complement())

		u, err := a.Union(b)
		require.NoError(t, err)
		assertSet(t, u, 12, []int{0, 4, 7, 11})
		assert.Equal(t, u, a.MustUnion(b))

		inter, err := a.Intersection(b)
		require.NoError(t, err)
		assertSet(t, inter, 12, []int{4, 7})
		assert.Equal(t, inter, a.MustIntersection(b))

		diff, err := a.Difference(b)
		require.NoError(t, err)
		assertSet(t, diff, 12, []int{0})
		assert.Equal(t, diff, a.MustDifference(b))

		sym, err := a.SymmetricDifference(b)
		require.NoError(t, err)
		assertSet(t, sym, 12, []int{0, 11})
		assert.Equal(t, sym, a.MustSymmetricDifference(b))

		disjAB, err := a.IsDisjoint(b)
		require.NoError(t, err)
		assert.False(t, disjAB)
		assert.False(t, a.MustIsDisjoint(b))

		disjAComp, err := a.IsDisjoint(compA)
		require.NoError(t, err)
		assert.True(t, disjAComp)
		assert.True(t, a.MustIsDisjoint(compA))

		assert.True(t, res(inter.IsSubsetOf(a)).get(t))
		assert.True(t, inter.MustIsSubsetOf(a))
		assert.True(t, res(inter.IsProperSubsetOf(a)).get(t))
		assert.True(t, inter.MustIsProperSubsetOf(a))

		assert.True(t, res(a.IsSubsetOf(a)).get(t))
		assert.True(t, a.MustIsSubsetOf(a))
		assert.False(t, res(a.IsProperSubsetOf(a)).get(t))
		assert.False(t, a.MustIsProperSubsetOf(a))

		assert.True(t, res(a.IsSupersetOf(inter)).get(t))
		assert.True(t, a.MustIsSupersetOf(inter))
		assert.True(t, res(a.IsProperSupersetOf(inter)).get(t))
		assert.True(t, a.MustIsProperSupersetOf(inter))

		assert.True(t, res(a.IsSupersetOf(a)).get(t))
		assert.True(t, a.MustIsSupersetOf(a))
		assert.False(t, res(a.IsProperSupersetOf(a)).get(t))
		assert.False(t, a.MustIsProperSupersetOf(a))
	})

	t.Run("boolean_algebra_identities_across_moduli", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			modulus int
			xNums   []int
			yNums   []int
		}{
			{name: "mod_1", modulus: 1, xNums: []int{}, yNums: []int{0}},
			{name: "mod_12", modulus: 12, xNums: []int{0, 2, 4, 7, 9}, yNums: []int{2, 5, 7, 11}},
			{name: "mod_64", modulus: 64, xNums: []int{0, 1, 31, 63}, yNums: []int{1, 32, 63}},
			{name: "mod_65", modulus: 65, xNums: []int{0, 32, 64}, yNums: []int{1, 32, 64}},
			{name: "mod_65536", modulus: 65_536, xNums: []int{0, 1024, 65_535}, yNums: []int{1024, 32_768, 65_535}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				x := pcset.MustNew(tc.modulus, tc.xNums...)
				y := pcset.MustNew(tc.modulus, tc.yNums...)
				empty := pcset.MustEmpty(tc.modulus)
				univ := pcset.MustUniverse(tc.modulus)

				assert.Equal(t, univ, empty.Complement())
				assert.Equal(t, empty, univ.Complement())
				assert.Equal(t, x, x.Complement().Complement())
				assert.Equal(t, univ, x.MustUnion(x.Complement()))
				assert.Equal(t, empty, x.MustIntersection(x.Complement()))
				assert.True(t, x.MustIsDisjoint(x.Complement()))

				// De Morgan's laws: (X u Y)^c == X^c n Y^c and (X n Y)^c == X^c u Y^c.
				assert.Equal(t, x.MustUnion(y).Complement(), x.Complement().MustIntersection(y.Complement()))
				assert.Equal(t, x.MustIntersection(y).Complement(), x.Complement().MustUnion(y.Complement()))

				// Symmetric difference identities.
				sym := x.MustSymmetricDifference(y)
				assert.Equal(t, x.MustDifference(y).MustUnion(y.MustDifference(x)), sym)
				assert.Equal(t, x.MustUnion(y).MustDifference(x.MustIntersection(y)), sym)
				assert.Equal(t, empty, x.MustSymmetricDifference(x))
				assert.Equal(t, x.Complement(), x.MustSymmetricDifference(univ))

				if tc.modulus <= 65 {
					ox := newOracleSet(tc.modulus, tc.xNums...)
					oy := newOracleSet(tc.modulus, tc.yNums...)
					assertSet(t, x.Complement(), tc.modulus, complementOracle(ox).elems)
					assertSet(t, x.MustUnion(y), tc.modulus, unionOracle(ox, oy).elems)
					assertSet(t, x.MustIntersection(y), tc.modulus, intersectionOracle(ox, oy).elems)
					assertSet(t, x.MustDifference(y), tc.modulus, differenceOracle(ox, oy).elems)
					assertSet(t, sym, tc.modulus, symDiffOracle(ox, oy).elems)
				}
			})
		}
	})

	t.Run("errors_and_must_panics", func(t *testing.T) {
		t.Parallel()

		var zero pcset.Set
		c19 := pcset.MustNew(19, 0, 4, 7)

		assert.False(t, zero.Complement().IsValid())

		pairs := []struct {
			name    string
			lhs     pcset.Set
			rhs     pcset.Set
			wantErr error
		}{
			{name: "invalid_lhs", lhs: zero, rhs: a, wantErr: pcset.ErrInvalidValue},
			{name: "invalid_rhs", lhs: a, rhs: zero, wantErr: pcset.ErrInvalidValue},
			{name: "both_invalid", lhs: zero, rhs: zero, wantErr: pcset.ErrInvalidValue},
			{name: "modulus_mismatch", lhs: a, rhs: c19, wantErr: pcset.ErrModulusMismatch},
		}

		for _, tc := range pairs {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				_, err := tc.lhs.Union(tc.rhs)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustUnion(tc.rhs) }, tc.wantErr)

				_, err = tc.lhs.Intersection(tc.rhs)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustIntersection(tc.rhs) }, tc.wantErr)

				_, err = tc.lhs.Difference(tc.rhs)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustDifference(tc.rhs) }, tc.wantErr)

				_, err = tc.lhs.SymmetricDifference(tc.rhs)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustSymmetricDifference(tc.rhs) }, tc.wantErr)

				_, err = tc.lhs.IsDisjoint(tc.rhs)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustIsDisjoint(tc.rhs) }, tc.wantErr)

				_, err = tc.lhs.IsSubsetOf(tc.rhs)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustIsSubsetOf(tc.rhs) }, tc.wantErr)

				_, err = tc.lhs.IsProperSubsetOf(tc.rhs)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustIsProperSubsetOf(tc.rhs) }, tc.wantErr)

				_, err = tc.lhs.IsSupersetOf(tc.rhs)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustIsSupersetOf(tc.rhs) }, tc.wantErr)

				_, err = tc.lhs.IsProperSupersetOf(tc.rhs)
				requireSentinelError(t, err, tc.wantErr)
				requirePanicsWithSentinel(t, func() { _ = tc.lhs.MustIsProperSupersetOf(tc.rhs) }, tc.wantErr)
			})
		}
	})
}
