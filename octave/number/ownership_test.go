package number_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertK8 checks the K-8 results of New(4, 4, 0, 1) on n.
func assertK8(t *testing.T, n number.Numbering, desc string) {
	t.Helper()

	spec := newSpec(4, 4, 0, 1)
	for a := -20; a <= 20; a++ {
		got, err := n.FromASPN(a)
		assertOutcome(t, spec.fromASPN(a), got, err, fmt.Sprintf("%s: FromASPN(%d)", desc, a))
	}
	for _, x := range []int{-2, -1, 0, 1, 2, 3} {
		got, err := n.ToASPN(x)
		assertOutcome(t, spec.toASPN(x), got, err, fmt.Sprintf("%s: ToASPN(%d)", desc, x))
	}
	assert.Equal(t, "ASPN 4 -> 4, skip [0 1]", n.String(), desc)
}

// assertK1 checks the K-1 results on a fresh Builtins and the built-in
// functions.
func assertK1(t *testing.T, desc string) {
	t.Helper()

	got := number.Builtins()
	require.Len(t, got, builtinCount, desc)
	for i, b := range builtinTable() {
		assert.Equal(t, b.label, got[i].Label(), "%s: label %d", desc, i)
		assert.True(t, got[i].Equal(b.get()), "%s: element %d", desc, i)
		assert.Equal(t, b.label, b.get().Label(), "%s: function %s", desc, b.label)

		for a := -5; a <= 5; a++ {
			x, err := b.get().FromASPN(a)
			assertOutcome(t, b.spec.fromASPN(a), x, err, fmt.Sprintf("%s: %s.FromASPN(%d)", desc, b.label, a))
		}
	}
}

// K-16, G-6: New copies the skipped numbers. Changing the slice passed to New
// after the call does not change the numbering, and New does not change the
// slice either (no sorting in place).
func TestNew_CopiesSkipped(t *testing.T) {
	t.Parallel()

	t.Run("MustNew", func(t *testing.T) {
		t.Parallel()

		skips := []int{0, 1}
		n := number.MustNew(4, 4, skips...)
		skips[0], skips[1] = 6, 7
		assertK8(t, n, "after changing the slice")
	})

	t.Run("New", func(t *testing.T) {
		t.Parallel()

		skips := []int{1, 0}
		n, err := number.New(4, 4, skips...)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 0}, skips, "New does not change the argument")

		skips[0], skips[1] = 4, 5
		assertK8(t, n, "after changing the slice")
	})

	t.Run("New with an error", func(t *testing.T) {
		t.Parallel()

		skips := []int{5, 3}
		n, err := number.New(4, 3, skips...)
		assertOnlyError(t, err, number.ErrInvalidDefinition, "New(4, 3, 5, 3)")
		assert.Equal(t, []int{5, 3}, skips, "New does not change the argument")

		skips[1] = 6
		assert.False(t, n.IsDefined(), "the result stays undefined")
		_, err = n.FromASPN(4)
		assertOnlyError(t, err, number.ErrUndefined, "FromASPN of the failed result")
	})
}

// assertSingleSkipped checks n against the oracle spec over a window of ASPN
// numbers and octave numbers, and its equality with ref.
func assertSingleSkipped(t *testing.T, n, ref number.Numbering, s spec, desc string) {
	t.Helper()

	for a := -30; a <= 30; a++ {
		got, err := n.FromASPN(a)
		assertOutcome(t, s.fromASPN(a), got, err, fmt.Sprintf("%s: FromASPN(%d)", desc, a))
	}
	for x := -30; x <= 30; x++ {
		got, err := n.ToASPN(x)
		assertOutcome(t, s.toASPN(x), got, err, fmt.Sprintf("%s: ToASPN(%d)", desc, x))
	}
	assert.True(t, n.Equal(ref), "%s: Equal(independent)", desc)
	assert.True(t, ref.Equal(n), "%s: independent.Equal", desc)
}

// K-16, G-6: New and MustNew copy a single skipped number as well. Changing
// the only element of the slice passed to the call does not change the
// numbering: the original number stays skipped and the new value stays used.
func TestNew_CopiesSingleSkipped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		aspn    int
		num     int
		skipped int
		changed int // the value written to the caller's slice after the call
	}{
		{name: "Franco-Belgian, number moved up", aspn: 4, num: 3, skipped: 0, changed: 1},
		{name: "Franco-Belgian, number moved down", aspn: 4, num: 3, skipped: 0, changed: -5},
		{name: "below the anchor, moved to the anchor side", aspn: 4, num: 3, skipped: 1, changed: 2},
		{name: "above the anchor", aspn: 4, num: 4, skipped: 6, changed: 5},
		{name: "above the anchor, moved far", aspn: 4, num: 4, skipped: 6, changed: 12},
		{name: "above the anchor, moved below it", aspn: 4, num: 4, skipped: 6, changed: -3},
		{name: "negative anchor", aspn: 1, num: -1, skipped: 0, changed: -3},
		{name: "negative anchor, negative skipped", aspn: -2, num: -5, skipped: -8, changed: -7},
		{name: "zero anchor, skipped below", aspn: 0, num: 0, skipped: -10, changed: 10},
	}

	for _, tt := range tests {
		t.Run("New/"+tt.name, func(t *testing.T) {
			t.Parallel()

			s := newSpec(tt.aspn, tt.num, tt.skipped)
			ref, err := number.New(tt.aspn, tt.num, tt.skipped)
			require.NoError(t, err)

			skips := []int{tt.skipped}
			n, err := number.New(tt.aspn, tt.num, skips...)
			require.NoError(t, err)
			assert.Equal(t, []int{tt.skipped}, skips, "New does not change the argument")
			assertSingleSkipped(t, n, ref, s, "before changing the slice")

			skips[0] = tt.changed
			assertSingleSkipped(t, n, ref, s, "after changing the slice")
		})

		t.Run("MustNew/"+tt.name, func(t *testing.T) {
			t.Parallel()

			s := newSpec(tt.aspn, tt.num, tt.skipped)
			ref := number.MustNew(tt.aspn, tt.num, tt.skipped)

			skips := []int{tt.skipped}
			n := number.MustNew(tt.aspn, tt.num, skips...)
			assert.Equal(t, []int{tt.skipped}, skips, "MustNew does not change the argument")
			assertSingleSkipped(t, n, ref, s, "before changing the slice")

			skips[0] = tt.changed
			assertSingleSkipped(t, n, ref, s, "after changing the slice")
		})
	}

	t.Run("Franco-Belgian", func(t *testing.T) {
		t.Parallel()

		skips := []int{0}
		n, err := number.New(4, 3, skips...)
		require.NoError(t, err)

		x, err := n.FromASPN(2)
		require.NoError(t, err)
		assert.Equal(t, 1, x)

		skips[0] = 1
		x, err = n.FromASPN(2)
		require.NoError(t, err)
		assert.Equal(t, 1, x, "FromASPN is unchanged")

		_, err = n.ToASPN(0)
		assertOnlyError(t, err, number.ErrSkipped, "ToASPN(0) after changing the slice")
		assert.True(t, n.Equal(number.FrancoBelgian()), "equals the built-in numbering")
	})
}

// K-16, G-6: Builtins returns a new slice owned by the caller; changing it
// changes neither later results nor the built-in functions nor other slices.
func TestBuiltins_Ownership(t *testing.T) {
	t.Parallel()

	t.Run("replace an element", func(t *testing.T) {
		t.Parallel()

		b := number.Builtins()
		b[0] = number.MiddleC3()
		b[1] = number.Numbering{}
		assertK1(t, "after replacing elements")
	})

	t.Run("reverse", func(t *testing.T) {
		t.Parallel()

		b := number.Builtins()
		slices.Reverse(b)
		assertK1(t, "after slices.Reverse")
	})

	t.Run("clear", func(t *testing.T) {
		t.Parallel()

		b := number.Builtins()
		clear(b)
		assertK1(t, "after clear")
	})

	t.Run("two slices are independent", func(t *testing.T) {
		t.Parallel()

		first := number.Builtins()
		second := number.Builtins()
		clear(first)

		require.Len(t, second, builtinCount)
		for i, b := range builtinTable() {
			assert.Equal(t, b.label, second[i].Label(), "element %d", i)
			assert.True(t, second[i].Equal(b.get()), "element %d", i)
		}
	})
}

// K-16: a copy of a value behaves like the original and is equal to it.
func TestCopy(t *testing.T) {
	t.Parallel()

	for _, n := range criteriaSet(t) {
		t.Run(n.name, func(t *testing.T) {
			t.Parallel()

			m := n.n
			assert.True(t, m.Equal(n.n))
			assert.True(t, n.n.Equal(m))
			assert.Equal(t, n.n.Label(), m.Label())
			assert.Equal(t, n.n.String(), m.String())

			for a := -10; a <= 10; a++ {
				x, err := m.FromASPN(a)
				assertOutcome(t, n.spec.fromASPN(a), x, err, fmt.Sprintf("copy FromASPN(%d)", a))
			}
		})
	}
}

// K-16: calls that fail and Must panics do not change the arguments.
func TestInputsUnchangedAfterErrors(t *testing.T) {
	t.Parallel()

	fb := number.FrancoBelgian()
	c3 := number.MiddleC3()
	gapped := number.MustNew(4, 4, 0, 1)

	_, err := fb.ToASPN(0)
	assertOnlyError(t, err, number.ErrSkipped, "ToASPN(0)")
	_, err = number.Convert(0, fb, c3)
	assertOnlyError(t, err, number.ErrSkipped, "Convert(0, Franco-Belgian, Middle C = C3)")
	_, err = c3.FromASPN(minInt)
	assertOnlyError(t, err, number.ErrOutOfRange, "Middle C = C3 FromASPN(MinInt)")
	_, err = gapped.ToASPN(1)
	assertOnlyError(t, err, number.ErrSkipped, "gapped ToASPN(1)")

	assertPanicsWithErrorIs(t, number.ErrSkipped, func() { fb.MustToASPN(0) }, "MustToASPN(0)")
	assertPanicsWithErrorIs(t, number.ErrSkipped, func() { number.MustConvert(0, fb, c3) }, "MustConvert")
	assertPanicsWithErrorIs(t, number.ErrOutOfRange, func() { c3.MustFromASPN(minInt) }, "MustFromASPN")
	assertPanicsWithErrorIs(t, number.ErrSkipped, func() { gapped.MustToASPN(0) }, "gapped MustToASPN(0)")

	tbl := builtinTable()
	assert.Equal(t, "Franco-Belgian", fb.Label())
	assert.Equal(t, "Middle C = C3", c3.Label())
	assert.True(t, fb.Equal(number.FrancoBelgian()))
	assert.True(t, c3.Equal(number.MiddleC3()))
	for a := -10; a <= 10; a++ {
		x, err := fb.FromASPN(a)
		assertOutcome(t, tbl[idxFrancoBelgian].spec.fromASPN(a), x, err, fmt.Sprintf("Franco-Belgian FromASPN(%d)", a))
		x, err = c3.FromASPN(a)
		assertOutcome(t, tbl[idxMiddleC3].spec.fromASPN(a), x, err, fmt.Sprintf("Middle C = C3 FromASPN(%d)", a))
	}
	assertK8(t, gapped, "after errors")
}
