package pitchclass_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The only invalid values are the zero Class and the zero Operator
// (architecture, "Invalid value"). The tests below follow E-5 items 1-7 and
// K-29 (1)-(7) with the tables of the architecture.

// validClassesForZero are the valid classes named in K-29 (2).
func validClassesForZero(t testing.TB) []pitchclass.Class {
	t.Helper()

	return []pitchclass.Class{newClass(t, 0, 12), newClass(t, 5, 12), newClass(t, 0, 19), newClass(t, 0, 1)}
}

// validOperatorsForZero are the valid operators named in K-29 (2).
func validOperatorsForZero() []pitchclass.Operator {
	return []pitchclass.Operator{pitchclass.MustIdentity(12), pitchclass.MustInversion(5, 12), pitchclass.MustIdentity(1)}
}

// K-29 (1), (2), (7): the zero values are invalid, not equal to anything,
// including themselves, while == of two zero values is true.
func TestZeroValidityAndEquality(t *testing.T) {
	t.Parallel()

	var z pitchclass.Class
	var o pitchclass.Operator

	assert.False(t, z.IsValid(), "(1) the zero Class")
	assert.False(t, o.IsValid(), "(1) the zero Operator")
	assert.False(t, pitchclass.Class{}.IsValid())
	assert.False(t, pitchclass.Operator{}.IsValid())

	assert.False(t, z.Equal(z), "(2) the zero Class with itself")
	assert.False(t, z.Equal(pitchclass.Class{}), "(2) two zero Classes")
	assert.False(t, o.Equal(o), "(2) the zero Operator with itself")
	assert.False(t, o.Equal(pitchclass.Operator{}), "(2) two zero Operators")

	for _, c := range validClassesForZero(t) {
		assert.False(t, z.Equal(c), "(2) zero.Equal(%v)", c)
		assert.False(t, c.Equal(z), "(2) %v.Equal(zero)", c)
	}

	for _, f := range validOperatorsForZero() {
		assert.False(t, o.Equal(f), "(2) zero.Equal(%v)", f)
		assert.False(t, f.Equal(o), "(2) %v.Equal(zero)", f)
	}

	assert.True(t, z == pitchclass.Class{}, "(7) == of zero Classes")
	assert.True(t, o == pitchclass.Operator{}, "(7) == of zero Operators")

	classes := map[pitchclass.Class]int{}
	classes[z]++
	classes[pitchclass.Class{}]++
	assert.Len(t, classes, 1, "(7) one map key")

	operators := map[pitchclass.Operator]int{}
	operators[o]++
	operators[pitchclass.Operator{}]++
	assert.Len(t, operators, 1, "(7) one map key")
}

// K-29 (3): the zero value comes before all valid values, without an error;
// with itself it compares as 0; sorting with the zero value is stable.
func TestZeroCompare(t *testing.T) {
	t.Parallel()

	var z pitchclass.Class
	var o pitchclass.Operator

	assert.Zero(t, z.Compare(z), "the zero Class with itself")
	assert.Zero(t, z.Compare(pitchclass.Class{}))
	assert.Zero(t, o.Compare(o), "the zero Operator with itself")

	for _, c := range append(validClassesForZero(t), newClass(t, 0, maxInt), newClass(t, maxInt-1, maxInt)) {
		assert.Equal(t, -1, z.Compare(c), "zero.Compare(%v)", c)
		assert.Equal(t, 1, c.Compare(z), "%v.Compare(zero)", c)
	}

	for _, f := range append(validOperatorsForZero(), pitchclass.MustIdentity(maxInt)) {
		assert.Equal(t, -1, o.Compare(f), "zero.Compare(%v)", f)
		assert.Equal(t, 1, f.Compare(o), "%v.Compare(zero)", f)
	}

	t.Run("sorting classes", func(t *testing.T) {
		t.Parallel()

		want := append([]pitchclass.Class{z}, classesOfAll(t, 2, 12, 19)...)
		for _, perm := range permutations(len(want)) {
			xs := make([]pitchclass.Class, len(want))
			for i, p := range perm {
				xs[i] = want[p]
			}

			slices.SortFunc(xs, pitchclass.Class.Compare)
			assert.Equal(t, want, xs, "sorted permutation %v", perm)
		}
	})

	t.Run("sorting operators", func(t *testing.T) {
		t.Parallel()

		want := append([]pitchclass.Operator{o}, operatorsOfOnly(t, 12)...)
		for _, perm := range permutations(len(want)) {
			xs := make([]pitchclass.Operator, len(want))
			for i, p := range perm {
				xs[i] = want[p]
			}

			slices.SortFunc(xs, pitchclass.Operator.Compare)
			assert.Equal(t, want, xs, "sorted permutation")
		}
	})
}

// K-29 (4): the canonical forms of the zero values differ from the forms of
// all valid values and are not parsed.
func TestZeroString(t *testing.T) {
	t.Parallel()

	var z pitchclass.Class
	var o pitchclass.Operator

	assert.Equal(t, "Class(invalid)", z.String())
	assert.Equal(t, "Operator(invalid)", o.String())
	assert.Equal(t, "Class(invalid)", fmt.Sprint(z))
	assert.Equal(t, "Operator(invalid)", fmt.Sprint(o))

	for _, c := range classesOfAll(t, 1, 2, 12, 19) {
		assert.NotEqual(t, c.String(), z.String(), "%v", c)
	}

	for _, size := range []int{1, 2, 12} {
		for _, f := range operatorsOfOnly(t, size) {
			assert.NotEqual(t, f.String(), o.String(), "%v", f)
		}
	}

	_, err := pitchclass.Parse(z.String())
	assertOnlyError(t, err, pitchclass.ErrNotation, "Parse of the zero Class form")

	for _, st := range styles() {
		_, err := pitchclass.ParseNotation(z.String(), st, 12)
		assertOnlyError(t, err, pitchclass.ErrNotation, fmt.Sprintf("ParseNotation of the zero Class form in %v", st))
	}

	_, err = pitchclass.ParseOperator(o.String())
	assertOnlyError(t, err, pitchclass.ErrNotation, "ParseOperator of the zero Operator form")
}

// K-29 (5): every two-operand operation with a zero operand in any position
// is ErrInvalidValue, checked before ErrModulusMismatch.
func TestZeroInTwoOperandOperations(t *testing.T) {
	t.Parallel()

	var z pitchclass.Class
	var o pitchclass.Operator
	c12, c19 := newClass(t, 1, 12), newClass(t, 1, 19)
	f12, f19 := pitchclass.MustTransposition(5, 12), pitchclass.MustTransposition(5, 19)

	type call struct {
		name string
		call func() (any, error)
		must func()
	}

	var calls []call
	for _, pos := range []struct {
		name string
		a, b pitchclass.Class
	}{
		{name: "(zero, 12)", a: z, b: c12},
		{name: "(12, zero)", a: c12, b: z},
		{name: "(zero, 19)", a: z, b: c19},
		{name: "(19, zero)", a: c19, b: z},
		{name: "(zero, zero)", a: z, b: z},
	} {
		a, b := pos.a, pos.b
		calls = append(calls,
			call{"IntervalTo" + pos.name, func() (any, error) { return a.IntervalTo(b) }, func() { a.MustIntervalTo(b) }},
			call{"IntervalClass" + pos.name, func() (any, error) { return a.IntervalClass(b) }, func() { a.MustIntervalClass(b) }},
			call{"IndexSum" + pos.name, func() (any, error) { return a.IndexSum(b) }, func() { a.MustIndexSum(b) }},
			call{"InvertAround" + pos.name, func() (any, error) { return a.InvertAround(b) }, func() { a.MustInvertAround(b) }},
			call{"InversionSwapping" + pos.name, func() (any, error) { return pitchclass.InversionSwapping(a, b) }, func() { pitchclass.MustInversionSwapping(a, b) }},
		)
	}

	for _, pos := range []struct {
		name    string
		c, x, y pitchclass.Class
	}{
		{name: "(zero; 12, 12)", c: z, x: c12, y: c12},
		{name: "(12; zero, 12)", c: c12, x: z, y: c12},
		{name: "(12; 12, zero)", c: c12, x: c12, y: z},
		{name: "(zero; 19, 12)", c: z, x: c19, y: c12},
		{name: "(12; zero, 19)", c: c12, x: z, y: c19},
		{name: "(19; 12, zero)", c: c19, x: c12, y: z},
		{name: "(zero; zero, zero)", c: z, x: z, y: z},
	} {
		c, x, y := pos.c, pos.x, pos.y
		calls = append(calls, call{
			"InvertSwapping" + pos.name,
			func() (any, error) { return c.InvertSwapping(x, y) },
			func() { c.MustInvertSwapping(x, y) },
		})
	}

	for _, pos := range []struct {
		name string
		f    pitchclass.Operator
		c    pitchclass.Class
	}{
		{name: "(zero, 12)", f: o, c: c12},
		{name: "(12, zero)", f: f12, c: z},
		{name: "(zero, 19)", f: o, c: c19},
		{name: "(19, zero)", f: f19, c: z},
		{name: "(zero, zero)", f: o, c: z},
	} {
		f, c := pos.f, pos.c
		calls = append(calls, call{"Apply" + pos.name, func() (any, error) { return f.Apply(c) }, func() { f.MustApply(c) }})
	}

	for _, pos := range []struct {
		name string
		f, g pitchclass.Operator
	}{
		{name: "(zero, 12)", f: o, g: f12},
		{name: "(12, zero)", f: f12, g: o},
		{name: "(zero, 19)", f: o, g: f19},
		{name: "(19, zero)", f: f19, g: o},
		{name: "(zero, zero)", f: o, g: o},
	} {
		f, g := pos.f, pos.g
		calls = append(calls, call{"Compose" + pos.name, func() (any, error) { return f.Compose(g) }, func() { f.MustCompose(g) }})
	}

	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got any
				err error
			)
			require.NotPanics(t, func() { got, err = tc.call() })
			assertOnlyError(t, err, pitchclass.ErrInvalidValue, tc.name)
			assertZeroResult(t, got, tc.name)
			assertPanicsWithErrorIs(t, pitchclass.ErrInvalidValue, tc.must, "Must variant of "+tc.name)
		})
	}
}

// K-29 (6): one-operand operations on the zero values never panic and give
// an error or a result that is visibly invalid (architecture tables).
func TestZeroInOneOperandOperations(t *testing.T) {
	t.Parallel()

	var z pitchclass.Class
	var o pitchclass.Operator

	t.Run("Class", func(t *testing.T) {
		t.Parallel()

		require.NotPanics(t, func() {
			assert.Equal(t, -1, z.Number(), "Number")
			assert.Zero(t, z.Modulus(), "Modulus")
			assertEmptySeq(t, z.Cycle(), "Cycle")

			for _, v := range []int{minInt, -1, 0, 1, maxInt} {
				assert.Equal(t, pitchclass.Class{}, z.Transpose(v), "Transpose(%d)", v)
				assert.Equal(t, pitchclass.Class{}, z.Invert(v), "Invert(%d)", v)
				assert.Equal(t, pitchclass.Class{}, z.Multiply(v), "Multiply(%d)", v)
				assert.False(t, z.Transpose(v).IsValid())
			}
		})
	})

	t.Run("Class notation", func(t *testing.T) {
		t.Parallel()

		for _, st := range append(styles(), pitchclass.Style(9)) {
			var (
				s   string
				err error
			)
			require.NotPanics(t, func() { s, err = z.Notation(st) })
			assertOnlyError(t, err, pitchclass.ErrInvalidValue, fmt.Sprintf("Notation(%v)", st))
			assert.Empty(t, s)
			assertPanicsWithErrorIs(t, pitchclass.ErrInvalidValue, func() { z.MustNotation(st) }, "MustNotation")
		}
	})

	t.Run("InversionAround", func(t *testing.T) {
		t.Parallel()

		var (
			f   pitchclass.Operator
			err error
		)
		require.NotPanics(t, func() { f, err = pitchclass.InversionAround(z) })
		assertOnlyError(t, err, pitchclass.ErrInvalidValue, "InversionAround")
		assert.Equal(t, pitchclass.Operator{}, f)
		assertPanicsWithErrorIs(t, pitchclass.ErrInvalidValue, func() { pitchclass.MustInversionAround(z) }, "MustInversionAround")
	})

	t.Run("Operator", func(t *testing.T) {
		t.Parallel()

		require.NotPanics(t, func() {
			assert.Zero(t, o.Modulus(), "Modulus")
			assert.Equal(t, -1, o.Multiplier(), "Multiplier")
			assert.Equal(t, -1, o.Index(), "Index")
			assert.False(t, o.IsTransposition(), "IsTransposition")
			assert.False(t, o.IsInversion(), "IsInversion")
			assert.Equal(t, pitchclass.Operator{}, o.Inverse(), "Inverse")
			assert.False(t, o.Inverse().IsValid())
		})
	})
}

// K-38, C-35: IsValid is true for every class and operator obtained from
// constructors and operations without an error, false only for the zero
// values, without an error or a panic.
func TestIsValid(t *testing.T) {
	t.Parallel()

	for _, size := range []int{1, 2, 12, 19} {
		t.Run(fmt.Sprintf("classes N=%d", size), func(t *testing.T) {
			t.Parallel()

			f := res(pitchclass.NewOperator(1, 3, size)).get(t)
			for _, c := range classesOf(t, size) {
				assert.True(t, c.IsValid(), "New: %v", c)
				for _, n := range []int{-30, -1, 0, 1, 7, 30, minInt, maxInt} {
					assert.True(t, c.Transpose(n).IsValid(), "Transpose")
					assert.True(t, c.Invert(n).IsValid(), "Invert")
					assert.True(t, c.Multiply(n).IsValid(), "Multiply")
				}

				got, err := f.Apply(c)
				require.NoError(t, err)
				assert.True(t, got.IsValid(), "Apply")
			}

			for x := -30; x <= 30; x++ {
				assert.True(t, res(pitchclass.Reduce(x, size)).get(t).IsValid(), "Reduce(%d, %d)", x, size)
			}
		})
	}

	for _, size := range []int{1, 2, 12} {
		t.Run(fmt.Sprintf("operators N=%d", size), func(t *testing.T) {
			t.Parallel()

			ops := operatorsOfOnly(t, size)
			for _, f := range ops {
				assert.True(t, f.IsValid(), "NewOperator %v", f)
				assert.True(t, f.Inverse().IsValid(), "Inverse of %v", f)
				assert.True(t, res(pitchclass.ParseOperator(f.String())).get(t).IsValid(), "ParseOperator %v", f)
				for _, g := range ops {
					assert.True(t, res(f.Compose(g)).get(t).IsValid(), "Compose")
				}
			}

			created := []pitchclass.Operator{
				res(pitchclass.Identity(size)).get(t),
				res(pitchclass.Transposition(1, size)).get(t),
				res(pitchclass.Inversion(1, size)).get(t),
				res(pitchclass.Multiplication(1, size)).get(t),
				res(pitchclass.InversionAround(newClass(t, 0, size))).get(t),
				res(pitchclass.InversionSwapping(newClass(t, 0, size), newClass(t, size-1, size))).get(t),
			}
			for _, f := range created {
				assert.True(t, f.IsValid(), "%v", f)
			}

			for _, g := range groups() {
				for f := range res(g.list(size)).get(t) {
					assert.True(t, f.IsValid(), "%s: %v", g.name, f)
				}
			}
		})
	}

	t.Run("zero values", func(t *testing.T) {
		t.Parallel()

		var cv, ov bool
		require.NotPanics(t, func() {
			cv = pitchclass.Class{}.IsValid()
			ov = pitchclass.Operator{}.IsValid()
		})
		assert.False(t, cv)
		assert.False(t, ov)
	})
}
