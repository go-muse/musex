package pitchclass_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-32, NF-3, G-11: operations do not change their inputs, also after an
// error or a panic of a Must variant.
func TestInputsAreNotChanged(t *testing.T) {
	t.Parallel()

	a, b := newClass(t, 4, 12), newClass(t, 9, 12)
	other := newClass(t, 1, 19)
	f, g := pitchclass.MustNewOperator(7, 5, 12), pitchclass.MustInversion(3, 12)

	check := func(desc string) {
		t.Helper()

		assertClass(t, 4, 12, a, desc+": a")
		assertClass(t, 9, 12, b, desc+": b")
		assertClass(t, 1, 19, other, desc+": other")
		assertOperator(t, opSpec{m: 7, n: 5, size: 12}, f, desc+": f")
		assertOperator(t, inversionSpec(3, 12), g, desc+": g")
	}

	_ = a.Transpose(5)
	_ = a.Invert(3)
	_ = a.Multiply(7)
	_, _ = a.IntervalTo(b)
	_, _ = a.IntervalClass(b)
	_, _ = a.IndexSum(b)
	_, _ = a.InvertAround(b)
	_, _ = a.InvertSwapping(a, b)
	_, _ = a.Notation(pitchclass.TE)
	_ = a.Compare(b)
	_ = a.Equal(b)
	_ = slices.Collect(a.Cycle())
	_, _ = f.Apply(a)
	_, _ = f.Compose(g)
	_ = f.Inverse()
	_ = f.Compare(g)
	_, _ = pitchclass.InversionAround(a)
	_, _ = pitchclass.InversionSwapping(a, b)
	check("after operations")

	_, _ = a.IntervalTo(other)
	_, _ = a.InvertSwapping(b, other)
	_, _ = f.Apply(other)
	_, _ = f.Compose(pitchclass.MustIdentity(19))
	check("after errors")

	_, _ = panicValue(func() { a.MustIntervalTo(other) })
	_, _ = panicValue(func() { a.MustInvertSwapping(b, other) })
	_, _ = panicValue(func() { f.MustApply(other) })
	_, _ = panicValue(func() { f.MustCompose(pitchclass.MustIdentity(19)) })
	check("after panics of Must variants")
}

// K-32, G-11: changing a slice or a *big.Int obtained from the package does
// not change later results; every call returns a new slice or *big.Int.
func TestResultsAreOwnedByTheCaller(t *testing.T) {
	t.Parallel()

	t.Run("InversionFixedClasses", func(t *testing.T) {
		t.Parallel()

		first := res(pitchclass.InversionFixedClasses(4, 12)).get(t)
		require.Len(t, first, 2)
		first[0] = newClass(t, 5, 12)
		first[1] = pitchclass.Class{}
		first = append(first[:0], newClass(t, 7, 12))
		assert.Len(t, first, 1)

		again := res(pitchclass.InversionFixedClasses(4, 12)).get(t)
		assert.Equal(t, []int{2, 8}, numbersOf(again))

		empty := res(pitchclass.InversionFixedClasses(1, 12)).get(t)
		empty = append(empty, newClass(t, 3, 12))
		assert.Len(t, empty, 1)
		assert.Empty(t, res(pitchclass.InversionFixedClasses(1, 12)).get(t))
	})

	t.Run("collected All", func(t *testing.T) {
		t.Parallel()

		collected := slices.Collect(res(pitchclass.All(12)).get(t))
		require.Len(t, collected, 12)
		for i := range collected {
			collected[i] = pitchclass.Class{}
		}

		assert.Equal(t, ints(0, 11), numbersOf(slices.Collect(res(pitchclass.All(12)).get(t))))
	})

	t.Run("collected InvertibleMultipliers", func(t *testing.T) {
		t.Parallel()

		collected := slices.Collect(res(pitchclass.InvertibleMultipliers(12)).get(t))
		require.Len(t, collected, 4)
		clear(collected)

		assert.Equal(t, []int{1, 5, 7, 11}, slices.Collect(res(pitchclass.InvertibleMultipliers(12)).get(t)))
		assert.Equal(t, 4, res(pitchclass.Totient(12)).get(t))
	})

	t.Run("collected Operators", func(t *testing.T) {
		t.Parallel()

		seq := res(pitchclass.Operators(12)).get(t)
		collected := slices.Collect(seq)
		require.Len(t, collected, 48)
		for i := range collected {
			collected[i] = pitchclass.Operator{}
		}

		assert.Equal(t, operatorSpecs(12), specsOf(slices.Collect(seq)))
		assert.Equal(t, operatorSpecs(12), specsOf(slices.Collect(res(pitchclass.Operators(12)).get(t))))
		assert.Len(t, slices.Collect(res(pitchclass.Transpositions(12)).get(t)), 12)
		assert.Len(t, slices.Collect(res(pitchclass.TranspositionsInversions(12)).get(t)), 24)
	})

	t.Run("big counts", func(t *testing.T) {
		t.Parallel()

		n := res(pitchclass.CountOperatorsBig(12)).get(t)
		require.Equal(t, "48", n.String())
		n.SetInt64(0)

		again := res(pitchclass.CountOperatorsBig(12)).get(t)
		assert.Equal(t, "48", again.String())
		assert.NotSame(t, n, again, "a new *big.Int on every call")
		assert.Equal(t, 48, res(pitchclass.CountOperators(12)).get(t))

		ti := res(pitchclass.CountTranspositionsInversionsBig(12)).get(t)
		require.Equal(t, "24", ti.String())
		ti.SetInt64(-1)

		tiAgain := res(pitchclass.CountTranspositionsInversionsBig(12)).get(t)
		assert.Equal(t, "24", tiAgain.String())
		assert.NotSame(t, ti, tiAgain, "a new *big.Int on every call")
	})
}

// K-32: a copy of a value behaves like the original and shares nothing.
func TestCopies(t *testing.T) {
	t.Parallel()

	a := newClass(t, 4, 12)
	b := a
	b = b.Transpose(1)
	assertClass(t, 4, 12, a, "the original")
	assertClass(t, 5, 12, b, "the copy")

	c := a
	assert.True(t, c.Equal(a))
	assert.True(t, c == a)
	assert.Equal(t, a.String(), c.String())

	f := res(pitchclass.NewOperator(7, 5, 12)).get(t)
	g := f
	g = g.Inverse()
	assertOperator(t, opSpec{m: 7, n: 5, size: 12}, f, "the original")
	assertOperator(t, opSpec{m: 7, n: 1, size: 12}, g, "the copy")

	h := f
	assert.True(t, h.Equal(f))
	assert.True(t, h == f)

	p := &a
	assert.Equal(t, a.String(), p.String(), "methods are available on *Class")
	assertClass(t, 6, 12, p.Transpose(2), "Transpose on *Class")
	assertClass(t, 4, 12, a, "the original after a call through the pointer")
}
