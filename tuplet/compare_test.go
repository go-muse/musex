package tuplet_test

import (
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEqualAndCoefficient_KP4 covers KP-4: 3:2 and 6:4 differ by notation
// and are equal by coefficient; 3:2 comes first in the total order.
func TestEqualAndCoefficient_KP4(t *testing.T) {
	t.Parallel()

	a := tuplet.MustNew(3, 2)
	b := tuplet.MustNew(6, 4)

	assert.False(t, a.Equal(b))
	assert.False(t, b.Equal(a))

	equal, err := a.EqualCoefficient(b)
	require.NoError(t, err)
	assert.True(t, equal)

	cmp, err := a.CompareCoefficient(b)
	require.NoError(t, err)
	assert.Equal(t, 0, cmp)

	assert.Negative(t, a.Compare(b))
	assert.Positive(t, b.Compare(a))
}

// equalityValues returns distinct values for equality checks, including
// the unset tuplet and pairs equal by coefficient.
func equalityValues() []tuplet.Tuplet {
	return append([]tuplet.Tuplet{{}}, sampleTuplets()...)
}

// TestEqual covers K-6 and G-6: Equal is equality of notation, agrees with
// ==, and the unset tuplet equals only the unset tuplet.
func TestEqual(t *testing.T) {
	t.Parallel()

	values := equalityValues()

	for i, a := range values {
		for j, b := range values {
			assert.Equal(t, i == j, a.Equal(b), "%s.Equal(%s)", a, b)
			assert.Equal(t, a == b, a.Equal(b), "== and Equal for %s and %s", a, b)
		}

		same := a
		assert.True(t, a.Equal(same), "%s equals its copy", a)
	}

	assert.True(t, tuplet.Tuplet{}.Equal(tuplet.Tuplet{}))
	assert.False(t, tuplet.Tuplet{}.Equal(tuplet.MustNew(1, 1)))
	assert.False(t, tuplet.MustNew(1, 1).Equal(tuplet.Tuplet{}))
	assert.True(t, tuplet.MustNew(6, 4).Equal(tuplet.Sextuplet()))
}

// TestEqual_MapKey covers K-6 and NF-2: the tuplet is usable as a map key,
// and 3:2 and 6:4 are different keys.
func TestEqual_MapKey(t *testing.T) {
	t.Parallel()

	m := map[tuplet.Tuplet]string{}
	m[tuplet.MustNew(3, 2)] = "triplet"
	m[tuplet.MustNew(6, 4)] = "sextuplet"
	m[tuplet.Tuplet{}] = "unset"

	assert.Len(t, m, 3)
	assert.Equal(t, "triplet", m[tuplet.Triplet()])
	assert.Equal(t, "sextuplet", m[tuplet.Sextuplet()])
	assert.Equal(t, "unset", m[tuplet.Tuplet{}])
}

// TestCompare_Properties covers K-6 and architecture 4.7: Compare is zero
// exactly for equal values, antisymmetric, transitive, and agrees with the
// order by coefficient, then actual, then normal, with unset first.
func TestCompare_Properties(t *testing.T) {
	t.Parallel()

	values := equalityValues()

	pairOf := func(tp tuplet.Tuplet) [2]int {
		if !tp.IsSet() {
			return [2]int{}
		}

		a, n := tp.MustPair()

		return [2]int{a, n}
	}

	for _, a := range values {
		for _, b := range values {
			ab := a.Compare(b)
			ba := b.Compare(a)

			assert.Equal(t, a == b, ab == 0, "Compare(%s, %s) = %d", a, b, ab)
			assert.Equal(t, sign(ab), -sign(ba), "antisymmetry for %s and %s", a, b)
			assert.Equal(t, oracleCompare(pairOf(a), pairOf(b), a.IsSet(), b.IsSet()), sign(ab),
				"order of %s and %s", a, b)

			for _, c := range values {
				if sign(ab) <= 0 && sign(b.Compare(c)) <= 0 {
					assert.LessOrEqual(t, sign(a.Compare(c)), 0, "transitivity for %s, %s, %s", a, b, c)
				}
			}
		}
	}
}

// TestCompare_Sort covers architecture 4.7: slices.SortFunc with
// tuplet.Tuplet.Compare gives the documented order.
func TestCompare_Sort(t *testing.T) {
	t.Parallel()

	want := []tuplet.Tuplet{
		{},
		tuplet.MustNew(7, 4),
		tuplet.MustNew(3, 2),
		tuplet.MustNew(6, 4),
		tuplet.MustNew(4, 3),
		tuplet.MustNew(5, 4),
		tuplet.MustNew(1, 1),
		tuplet.MustNew(5, 5),
		tuplet.MustNew(2, 3),
	}

	got := []tuplet.Tuplet{
		tuplet.MustNew(5, 5),
		tuplet.MustNew(2, 3),
		tuplet.MustNew(6, 4),
		{},
		tuplet.MustNew(5, 4),
		tuplet.MustNew(1, 1),
		tuplet.MustNew(3, 2),
		tuplet.MustNew(4, 3),
		tuplet.MustNew(7, 4),
	}

	slices.SortFunc(got, tuplet.Tuplet.Compare)
	assert.Equal(t, want, got)
}

// TestCompareCoefficient covers O-9 and architecture 4.7 for pairs with
// numbers near math.MaxInt, where 64-bit cross products overflow.
func TestCompareCoefficient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b [2]int
	}{
		{a: [2]int{3, 2}, b: [2]int{6, 4}},
		{a: [2]int{3, 2}, b: [2]int{2, 3}},
		{a: [2]int{5, 4}, b: [2]int{4, 3}},
		{a: [2]int{1, 1}, b: [2]int{math.MaxInt, math.MaxInt}},
		{a: [2]int{math.MaxInt, math.MaxInt - 1}, b: [2]int{math.MaxInt - 1, math.MaxInt - 2}},
		{a: [2]int{math.MaxInt - 1, math.MaxInt}, b: [2]int{math.MaxInt - 2, math.MaxInt - 1}},
		{a: [2]int{math.MaxInt, 1}, b: [2]int{math.MaxInt - 1, 1}},
		{a: [2]int{1, math.MaxInt}, b: [2]int{1, math.MaxInt - 1}},
		{a: [2]int{math.MaxInt, math.MaxInt - 2}, b: [2]int{math.MaxInt - 1, math.MaxInt - 3}},
		{a: [2]int{1 << 62, 1 << 61}, b: [2]int{2, 1}},
		{a: [2]int{3 << 61, 1 << 62}, b: [2]int{3, 2}},
		{a: [2]int{math.MaxInt, 1 << 62}, b: [2]int{1 << 62, 3 << 60}},
	}

	for _, tt := range tests {
		for _, swap := range []bool{false, true} {
			a, b := tt.a, tt.b
			if swap {
				a, b = b, a
			}

			t.Run(itoa(a[0])+":"+itoa(a[1])+" vs "+itoa(b[0])+":"+itoa(b[1]), func(t *testing.T) {
				t.Parallel()

				ta := tuplet.MustNew(a[0], a[1])
				tb := tuplet.MustNew(b[0], b[1])
				want := ratInts(a[1], a[0]).Cmp(ratInts(b[1], b[0]))

				cmp, err := ta.CompareCoefficient(tb)
				require.NoError(t, err)
				assert.Equal(t, want, cmp)
				assert.Equal(t, want, ta.MustCompareCoefficient(tb))

				equal, err := ta.EqualCoefficient(tb)
				require.NoError(t, err)
				assert.Equal(t, want == 0, equal)
				assert.Equal(t, want == 0, ta.MustEqualCoefficient(tb))

				assert.Equal(t, oracleCompare(a, b, true, true), sign(ta.Compare(tb)))
			})
		}
	}
}

// TestCompareCoefficient_Oracle checks CompareCoefficient and EqualCoefficient
// on all pairs of sample tuplets.
func TestCompareCoefficient_Oracle(t *testing.T) {
	t.Parallel()

	values := sampleTuplets()

	for _, a := range values {
		for _, b := range values {
			aActual, aNormal := a.MustPair()
			bActual, bNormal := b.MustPair()
			want := ratInts(aNormal, aActual).Cmp(ratInts(bNormal, bActual))

			cmp, err := a.CompareCoefficient(b)
			if assert.NoError(t, err) {
				assert.Equal(t, want, cmp, "%s vs %s", a, b)
			}

			equal, err := a.EqualCoefficient(b)
			if assert.NoError(t, err) {
				assert.Equal(t, want == 0, equal, "%s vs %s", a, b)
			}
		}
	}
}
