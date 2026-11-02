package alteration_test

import (
	"math"
	"slices"
	"testing"

	"github.com/go-muse/musex/alteration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// zeroValues returns the zero value obtained by a declaration and by a
// composite literal.
func zeroValues() map[string]alteration.Alteration {
	var declared alteration.Alteration

	return map[string]alteration.Alteration{
		"declared": declared,
		"literal":  {},
	}
}

// K-15 (architecture "Invalid values"): an invalid value of alteration cannot
// be obtained; the zero value is the natural value 0 itself.
func TestZeroValue(t *testing.T) {
	t.Parallel()

	for name, zero := range zeroValues() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, zero.Equal(alteration.Natural()), "Equal(Natural())")
			assert.True(t, alteration.Natural().Equal(zero), "Natural().Equal(zero)")
			assertSame(t, alteration.Natural(), zero, "== Natural()")
			assert.Equal(t, frac{0, 1}, fractionOf(zero))
			assert.Equal(t, "0", zero.String())
			assert.True(t, zero.IsNatural())
			assert.True(t, zero.IsInteger())
			assert.Equal(t, 0, zero.Sign())
			assert.Equal(t, 0, zero.Compare(alteration.Natural()))
		})
	}
}

// valueCall is a fallible operation that takes a value in one position and
// its Must variant.
type valueCall struct {
	name string
	op   func(a alteration.Alteration) (alteration.Alteration, error)
	must func(a alteration.Alteration) alteration.Alteration
}

// valueQuery is an operation without an error that takes a value in one
// position.
type valueQuery struct {
	name string
	op   func(a alteration.Alteration) any
}

// K-15: every operation that takes a value as receiver or argument gives on
// the zero value the same result as on Natural(): the same value or the same
// sentinel error, the same panic of the Must variant.
func TestZeroValueOperations(t *testing.T) {
	t.Parallel()

	others := slices.Concat(setS(), []frac{edgeTop(), edgeBottom(), edgeInvQ(), edgeNearTop()})
	ints := []int{-3, -1, 0, 1, 5, bigM, bigM + 1, -bigM - 1, math.MaxInt, math.MinInt}

	build := func(t *testing.T) ([]valueCall, []valueQuery) {
		t.Helper()

		calls := []valueCall{
			{name: "Raise", op: alteration.Alteration.Raise, must: alteration.Alteration.MustRaise},
			{name: "Lower", op: alteration.Alteration.Lower, must: alteration.Alteration.MustLower},
		}
		queries := []valueQuery{
			{name: "Fraction", op: func(a alteration.Alteration) any { return fractionOf(a) }},
			{name: "Sign", op: func(a alteration.Alteration) any { return a.Sign() }},
			{name: "IsNatural", op: func(a alteration.Alteration) any { return a.IsNatural() }},
			{name: "IsInteger", op: func(a alteration.Alteration) any { return a.IsInteger() }},
			{name: "String", op: func(a alteration.Alteration) any { return a.String() }},
		}

		for _, n := range ints {
			calls = append(calls,
				valueCall{
					name: "RaiseBy",
					op:   func(a alteration.Alteration) (alteration.Alteration, error) { return a.RaiseBy(n) },
					must: func(a alteration.Alteration) alteration.Alteration { return a.MustRaiseBy(n) },
				},
				valueCall{
					name: "LowerBy",
					op:   func(a alteration.Alteration) (alteration.Alteration, error) { return a.LowerBy(n) },
					must: func(a alteration.Alteration) alteration.Alteration { return a.MustLowerBy(n) },
				},
				valueCall{
					name: "Mul",
					op:   func(a alteration.Alteration) (alteration.Alteration, error) { return a.Mul(n) },
					must: func(a alteration.Alteration) alteration.Alteration { return a.MustMul(n) },
				},
			)
		}

		for _, f := range others {
			x := value(t, f)
			calls = append(calls,
				valueCall{
					name: "receiver of Add" + f.String(),
					op:   func(a alteration.Alteration) (alteration.Alteration, error) { return a.Add(x) },
					must: func(a alteration.Alteration) alteration.Alteration { return a.MustAdd(x) },
				},
				valueCall{
					name: "argument of Add" + f.String(),
					op:   func(a alteration.Alteration) (alteration.Alteration, error) { return x.Add(a) },
					must: func(a alteration.Alteration) alteration.Alteration { return x.MustAdd(a) },
				},
				valueCall{
					name: "receiver of Sub" + f.String(),
					op:   func(a alteration.Alteration) (alteration.Alteration, error) { return a.Sub(x) },
					must: func(a alteration.Alteration) alteration.Alteration { return a.MustSub(x) },
				},
				valueCall{
					name: "argument of Sub" + f.String(),
					op:   func(a alteration.Alteration) (alteration.Alteration, error) { return x.Sub(a) },
					must: func(a alteration.Alteration) alteration.Alteration { return x.MustSub(a) },
				},
			)
			queries = append(queries,
				valueQuery{name: "receiver of Compare" + f.String(), op: func(a alteration.Alteration) any { return a.Compare(x) }},
				valueQuery{name: "argument of Compare" + f.String(), op: func(a alteration.Alteration) any { return x.Compare(a) }},
				valueQuery{name: "receiver of Equal" + f.String(), op: func(a alteration.Alteration) any { return a.Equal(x) }},
				valueQuery{name: "argument of Equal" + f.String(), op: func(a alteration.Alteration) any { return x.Equal(a) }},
			)
		}

		return calls, queries
	}

	for name, zero := range zeroValues() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			calls, queries := build(t)
			natural := alteration.Natural()

			for _, c := range calls {
				want := call(func() (alteration.Alteration, error) { return c.op(natural) })
				got := call(func() (alteration.Alteration, error) { return c.op(zero) })
				assertSameOutcome(t, want, got, c.name)

				wantMust := callMust(func() alteration.Alteration { return c.must(natural) })
				gotMust := callMust(func() alteration.Alteration { return c.must(zero) })
				assertSameOutcome(t, wantMust, gotMust, "Must variant of "+c.name)
			}

			for _, q := range queries {
				assert.Equal(t, q.op(natural), q.op(zero), q.name)
			}
		})
	}

	t.Run("results on the zero value", func(t *testing.T) {
		t.Parallel()

		var zero alteration.Alteration

		up, err := zero.Raise()
		require.NoError(t, err)
		assertSame(t, alteration.Sharp(), up, "0 up 1 is a sharp")

		down, err := zero.Lower()
		require.NoError(t, err)
		assertSame(t, alteration.Flat(), down, "0 down 1 is a flat")

		sum, err := zero.Add(value(t, frac{1, 2}))
		require.NoError(t, err)
		assert.Equal(t, frac{1, 2}, fractionOf(sum))

		product, err := zero.Mul(math.MaxInt)
		require.NoError(t, err)
		assert.True(t, product.IsNatural(), "0 * MaxInt is natural")

		assert.Equal(t, -1, zero.Compare(alteration.Sharp()))
		assert.Equal(t, 1, alteration.Sharp().Compare(zero))
	})
}
