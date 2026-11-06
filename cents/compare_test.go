package cents_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-14: exact equality, however the values were obtained.
func TestEqual(t *testing.T) {
	t.Parallel()

	r32 := noErr(cents.FromRatio(3, 2))(t)
	q := byName(t, setS(t), "Q")

	t.Run("criteria", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "696.578428", q.v.Decimal(6, 0))
		assertCents(t, noErr(cents.FromRatio(5, 4))(t), noErr(noErr(q.v.Mul(4))(t).Sub(noErr(cents.FromInt(2400))(t)))(t), "Q*4 - <2400>")
		assertCents(t, noErr(cents.FromInt(700))(t), noErr(cents.Step(7, 12))(t), "E(7,12) = <700>")
		assertCents(t, noErr(cents.FromFloat(700.0))(t), noErr(cents.Step(7, 12))(t), "F(700.0) = E(7,12)")
		assertCents(t, noErr(cents.FromRatio(10, 1))(t), noErr(cents.Savart().Mul(1000))(t), "SAV*1000 = R(10:1)")

		// Werckmeister III: 8 pure fifths and 4 fifths narrowed by PYTH/4.
		narrow := noErr(r32.Sub(noErr(cents.PythagoreanComma().Div(4))(t)))(t)
		sum := noErr(noErr(r32.Mul(8))(t).Add(noErr(narrow.Mul(4))(t)))(t)
		assert.True(t, noErr(sum.Sub(noErr(cents.FromInt(8400))(t)))(t) == cents.Cents{}, "Werckmeister III")
	})

	t.Run("not equal", func(t *testing.T) {
		t.Parallel()

		pairs := []struct {
			name string
			a, b cents.Cents
		}{
			{"R(3:2) and Rf(1.5)", r32, noErr(cents.FromRatioFloat(1.5))(t)},
			{"R(3:2) and <701.955000865387>", r32, noErr(cents.FromFraction(701955000865387, 1000000000000))(t)},
			{"<1/2> and <1/3>", noErr(cents.FromFraction(1, 2))(t), noErr(cents.FromFraction(1, 3))(t)},
		}

		for _, p := range pairs {
			assert.False(t, p.a.Equal(p.b), p.name)
			assert.False(t, p.b.Equal(p.a), p.name)
			assert.False(t, p.a == p.b, p.name)
			assert.NotEqual(t, 0, p.a.Compare(p.b), p.name)
		}
	})

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		set := setS(t)
		for _, a := range set {
			assert.True(t, a.v.Equal(a.v), "reflexive %s", a.name)

			for _, b := range set {
				eq := a.v.Equal(b.v)
				assert.Equal(t, a.c.equal(b.c), eq, "%s = %s", a.name, b.name)
				assert.Equal(t, eq, b.v.Equal(a.v), "symmetric %s, %s", a.name, b.name)
				assert.Equal(t, eq, a.v.Compare(b.v) == 0, "Equal ⇔ Compare == 0: %s, %s", a.name, b.name)
				assert.Equal(t, eq, a.v == b.v, "Equal ⇔ ==: %s, %s", a.name, b.name)
			}
		}
	})
}

// K-15, G-3: exact order, also for almost equal values.
func TestCompare(t *testing.T) {
	t.Parallel()

	r32 := noErr(cents.FromRatio(3, 2))(t)

	t.Run("almost equal", func(t *testing.T) {
		t.Parallel()

		below := noErr(cents.FromFraction(701955000865387, 1000000000000))(t)
		above := noErr(cents.FromFraction(701955000865388, 1000000000000))(t)
		parsed := noErr(cents.Parse("701.955000865387"))(t)

		assertCents(t, below, parsed)
		assert.Equal(t, 1, r32.Compare(below), "R(3:2) > <701.955000865387>")
		assert.Equal(t, -1, below.Compare(r32))
		assert.Equal(t, -1, r32.Compare(above), "R(3:2) < <701.955000865388>")
		assert.Equal(t, 1, above.Compare(r32))
		assert.Equal(t, 1, r32.Compare(parsed))
	})

	t.Run("chains", func(t *testing.T) {
		t.Parallel()

		set := setS(t)
		q := byName(t, set, "Q").v

		chains := [][]cents.Cents{
			{q, noErr(cents.FromInt(700))(t), r32},
			{cents.SyntonicComma(), cents.HoldrianComma(), cents.PythagoreanComma()},
			{cents.Schisma(), cents.SeptimalKleisma(), cents.Kleisma(), cents.PtolemysComma(), cents.Diaschisma(), cents.SyntonicComma()},
		}

		for _, chain := range chains {
			for i := 1; i < len(chain); i++ {
				assert.Equal(t, -1, chain[i-1].Compare(chain[i]), "%s < %s", chain[i-1], chain[i])
				assert.Equal(t, 1, chain[i].Compare(chain[i-1]), "%s > %s", chain[i], chain[i-1])
			}
		}
	})

	t.Run("S in increasing order", func(t *testing.T) {
		t.Parallel()

		set := setS(t)
		names := sAscending()
		require.Len(t, names, len(set))

		for i := 1; i < len(names); i++ {
			a, b := byName(t, set, names[i-1]), byName(t, set, names[i])
			assert.Equal(t, -1, a.v.Compare(b.v), "%s < %s", a.name, b.name)
			assert.Equal(t, -1, a.c.cmp(b.c), "oracle: %s < %s", a.name, b.name)
		}

		shuffled := make([]cents.Cents, 0, len(set))
		for i := range set {
			shuffled = append(shuffled, set[(i*7+3)%len(set)].v) // 7 is coprime with 20
		}

		slices.SortFunc(shuffled, cents.Cents.Compare)

		for i, name := range names {
			assertCents(t, byName(t, set, name).v, shuffled[i], "position %d", i)
		}
	})

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		set := setS(t)
		for _, a := range set {
			for _, b := range set {
				ab, ba := a.v.Compare(b.v), b.v.Compare(a.v)
				assert.Contains(t, []int{-1, 0, 1}, ab, "one outcome")
				assert.Equal(t, a.c.cmp(b.c), ab, "%s vs %s", a.name, b.name)
				assert.Equal(t, -ab, ba, "antisymmetric %s, %s", a.name, b.name)

				if diff, err := b.v.Sub(a.v); err == nil {
					assert.Equal(t, ab, -diff.Sign(), "Compare agrees with the sign of %s - %s", b.name, a.name)
				}

				for _, c := range set {
					if ab <= 0 && b.v.Compare(c.v) <= 0 {
						assert.LessOrEqual(t, a.v.Compare(c.v), 0, "transitive %s <= %s <= %s", a.name, b.name, c.name)
					}
				}
			}
		}
	})
}

// K-16: closeness with a tolerance |c - d| <= t.
func TestNear(t *testing.T) {
	t.Parallel()

	r32, c700 := noErr(cents.FromRatio(3, 2))(t), noErr(cents.FromInt(700))(t)

	cases := []struct {
		name string
		tol  cents.Cents
		want bool
	}{
		{"2", noErr(cents.FromInt(2))(t), true},
		{"1.955", noErr(cents.FromFraction(1955, 1000))(t), false},
		{"1.956", noErr(cents.FromFraction(1956, 1000))(t), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := r32.Near(c700, tc.tol)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assertMust(t, got, err, func() bool { return r32.MustNear(c700, tc.tol) }, "MustNear")

			got, err = c700.Near(r32, tc.tol)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got, "symmetric")
			assertMust(t, got, err, func() bool { return c700.MustNear(r32, tc.tol) }, "MustNear symmetric")
		})
	}

	t.Run("range ends", func(t *testing.T) {
		t.Parallel()

		top, bottom := noErr(cents.FromInt(9000000))(t), noErr(cents.FromInt(-9000000))(t)

		got, err := top.Near(bottom, cents.Cent())
		require.NoError(t, err, "the difference is not limited by the range")
		assert.False(t, got)
		assertMust(t, got, err, func() bool { return top.MustNear(bottom, cents.Cent()) }, "MustNear")

		got, err = top.Near(bottom, top)
		require.NoError(t, err)
		assert.False(t, got)
		assertMust(t, got, err, func() bool { return top.MustNear(bottom, top) }, "MustNear")
	})

	t.Run("negative tolerance", func(t *testing.T) {
		t.Parallel()

		for _, tol := range []cents.Cents{noErr(cents.FromInt(-1))(t), noErr(cents.FromFraction(-1, 1000000))(t), noErr(cents.FromRatio(2, 3))(t)} {
			got, err := r32.Near(r32, tol)
			assertOnlyError(t, err, cents.ErrInvalidArgument, "tolerance %s", tol)
			assert.False(t, got)
			assertMust(t, got, err, func() bool { return r32.MustNear(r32, tol) }, "MustNear: tolerance %s", tol)
		}
	})

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		set := setS(t)
		for _, a := range set {
			got, err := a.v.Near(a.v, cents.Zero())
			require.NoError(t, err)
			assert.True(t, got, "(a, a, 0) for %s", a.name)

			for _, b := range set {
				got, err := a.v.Near(b.v, cents.Zero())
				require.NoError(t, err, "(%s, %s, 0)", a.name, b.name)
				assert.Equal(t, a.v.Equal(b.v), got, "(%s, %s, 0)", a.name, b.name)
				assertMust(t, got, err, func() bool { return a.v.MustNear(b.v, cents.Zero()) }, "MustNear(%s, %s, 0)", a.name, b.name)

				got, err = a.v.Near(b.v, cents.Cents{})
				require.NoError(t, err)
				assert.Equal(t, a.v.Equal(b.v), got, "(%s, %s, zero value)", a.name, b.name)

				diff := a.c.sub(b.c)
				for _, tol := range []comp{ratC(1, 1), ratC(25, 1), ratC(1200, 1)} {
					want := diff.cmp(tol) <= 0 && diff.neg().cmp(tol) <= 0
					tolV := fromComp(t, tol)
					got, err := a.v.Near(b.v, tolV)
					require.NoError(t, err, "(%s, %s, %s)", a.name, b.name, tol)
					assert.Equal(t, want, got, "(%s, %s, %s)", a.name, b.name, tol)
					assertMust(t, got, err, func() bool { return a.v.MustNear(b.v, tolV) }, "MustNear(%s, %s, %s)", a.name, b.name, tol)
				}
			}
		}
	})
}

// K-28, NF-9: values as map keys.
func TestMapKey(t *testing.T) {
	t.Parallel()

	r32 := noErr(cents.FromRatio(3, 2))(t)

	m := map[cents.Cents]string{}
	m[noErr(cents.FromRatio(6, 4))(t)] = "6:4"
	m[r32] = "3:2"
	assert.Len(t, m, 1, "R(6:4) and R(3:2) are one key")
	assert.Equal(t, "3:2", m[noErr(cents.FromRatio(6, 4))(t)])

	pyth := noErr(noErr(r32.Mul(12))(t).Sub(noErr(cents.FromInt(8400))(t)))(t)
	m[cents.PythagoreanComma()] = "PYTH"
	m[pyth] = "12 fifths - 7 octaves"
	assert.Len(t, m, 2, "PYTH and R(3:2)*12 - <8400> are one key")

	m[noErr(cents.FromFloat(700.0))(t)] = "F(700.0)"
	m[noErr(cents.Step(7, 12))(t)] = "E(7,12)"
	assert.Len(t, m, 3, "F(700.0) and E(7,12) are one key")

	m[noErr(cents.FromRatioFloat(1.5))(t)] = "Rf(1.5)"
	assert.Len(t, m, 4, "R(3:2) and Rf(1.5) are different keys")

	named := map[cents.Cents]string{}
	for _, nv := range cents.NamedValues() {
		named[nv.Value] = nv.Name
	}

	assert.Len(t, named, 19, "19 named values are 19 keys")

	set := setS(t)
	keys := map[cents.Cents]int{}
	for i, a := range set {
		keys[a.v] = i
	}

	for _, a := range set {
		for _, b := range set {
			assert.Equal(t, a.v.Equal(b.v), keys[a.v] == keys[b.v], "%s, %s", a.name, b.name)
		}
	}
}
