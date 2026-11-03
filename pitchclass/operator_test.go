package pitchclass_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-18, C-21, C-22: operators created by the constructors, with m and n
// reduced by the circle, and their properties.
func TestOperatorCreation(t *testing.T) {
	t.Parallel()

	// The classes are created here, in the test goroutine, so the table
	// closures run in the subtests do not use the parent's t.
	c0, c1, c2, c3, c4 := newClass(t, 0, 12), newClass(t, 1, 12), newClass(t, 2, 12), newClass(t, 3, 12), newClass(t, 4, 12)
	last, beforeLast := newClass(t, maxInt-1, maxInt), newClass(t, maxInt-2, maxInt)

	tests := []struct {
		name string
		call func() (pitchclass.Operator, error)
		must func() pitchclass.Operator
		want opSpec
	}{
		{
			name: "NewOperator(1, 5, 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(1, 5, 12) },
			must: func() pitchclass.Operator { return pitchclass.MustNewOperator(1, 5, 12) },
			want: opSpec{m: 1, n: 5, size: 12},
		},
		{
			name: "NewOperator(11, 5, 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(11, 5, 12) },
			must: func() pitchclass.Operator { return pitchclass.MustNewOperator(11, 5, 12) },
			want: opSpec{m: 11, n: 5, size: 12},
		},
		{
			name: "NewOperator(17, -7, 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(17, -7, 12) },
			must: func() pitchclass.Operator { return pitchclass.MustNewOperator(17, -7, 12) },
			want: opSpec{m: 5, n: 5, size: 12},
		},
		{
			name: "NewOperator(MAX, MIN, 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(maxInt, minInt, 12) },
			must: func() pitchclass.Operator { return pitchclass.MustNewOperator(maxInt, minInt, 12) },
			want: opSpec{m: 7, n: 4, size: 12},
		},
		{
			name: "NewOperator(5, 3, 1)",
			call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(5, 3, 1) },
			must: func() pitchclass.Operator { return pitchclass.MustNewOperator(5, 3, 1) },
			want: opSpec{m: 0, n: 0, size: 1},
		},
		{
			name: "NewOperator(MIN, MIN, MAX)",
			call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(minInt, minInt, maxInt) },
			must: func() pitchclass.Operator { return pitchclass.MustNewOperator(minInt, minInt, maxInt) },
			want: opSpec{m: maxInt - 1, n: maxInt - 1, size: maxInt},
		},
		{
			name: "Identity(12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.Identity(12) },
			must: func() pitchclass.Operator { return pitchclass.MustIdentity(12) },
			want: opSpec{m: 1, n: 0, size: 12},
		},
		{
			name: "Identity(1)",
			call: func() (pitchclass.Operator, error) { return pitchclass.Identity(1) },
			must: func() pitchclass.Operator { return pitchclass.MustIdentity(1) },
			want: opSpec{m: 0, n: 0, size: 1},
		},
		{
			name: "Transposition(5, 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.Transposition(5, 12) },
			must: func() pitchclass.Operator { return pitchclass.MustTransposition(5, 12) },
			want: opSpec{m: 1, n: 5, size: 12},
		},
		{
			name: "Transposition(-1, 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.Transposition(-1, 12) },
			must: func() pitchclass.Operator { return pitchclass.MustTransposition(-1, 12) },
			want: opSpec{m: 1, n: 11, size: 12},
		},
		{
			name: "Transposition(MIN, MAX)",
			call: func() (pitchclass.Operator, error) { return pitchclass.Transposition(minInt, maxInt) },
			must: func() pitchclass.Operator { return pitchclass.MustTransposition(minInt, maxInt) },
			want: opSpec{m: 1, n: maxInt - 1, size: maxInt},
		},
		{
			name: "Inversion(5, 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.Inversion(5, 12) },
			must: func() pitchclass.Operator { return pitchclass.MustInversion(5, 12) },
			want: opSpec{m: 11, n: 5, size: 12},
		},
		{
			name: "Inversion(MAX, 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.Inversion(maxInt, 12) },
			must: func() pitchclass.Operator { return pitchclass.MustInversion(maxInt, 12) },
			want: opSpec{m: 11, n: 7, size: 12},
		},
		{
			name: "Inversion(0, 1)",
			call: func() (pitchclass.Operator, error) { return pitchclass.Inversion(0, 1) },
			must: func() pitchclass.Operator { return pitchclass.MustInversion(0, 1) },
			want: opSpec{m: 0, n: 0, size: 1},
		},
		{
			name: "InversionAround(2 of 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.InversionAround(c2) },
			must: func() pitchclass.Operator { return pitchclass.MustInversionAround(c2) },
			want: opSpec{m: 11, n: 4, size: 12},
		},
		{
			name: "InversionAround(MAX-1 of MAX)",
			call: func() (pitchclass.Operator, error) { return pitchclass.InversionAround(last) },
			must: func() pitchclass.Operator { return pitchclass.MustInversionAround(last) },
			want: opSpec{m: maxInt - 1, n: maxInt - 2, size: maxInt},
		},
		{
			name: "InversionSwapping(0 of 12, 4 of 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.InversionSwapping(c0, c4) },
			must: func() pitchclass.Operator { return pitchclass.MustInversionSwapping(c0, c4) },
			want: opSpec{m: 11, n: 4, size: 12},
		},
		{
			name: "InversionSwapping(1 of 12, 3 of 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.InversionSwapping(c1, c3) },
			must: func() pitchclass.Operator { return pitchclass.MustInversionSwapping(c1, c3) },
			want: opSpec{m: 11, n: 4, size: 12},
		},
		{
			name: "InversionSwapping(MAX-1 of MAX, MAX-2 of MAX)",
			call: func() (pitchclass.Operator, error) {
				return pitchclass.InversionSwapping(last, beforeLast)
			},
			must: func() pitchclass.Operator {
				return pitchclass.MustInversionSwapping(last, beforeLast)
			},
			want: opSpec{m: maxInt - 1, n: maxInt - 3, size: maxInt},
		},
		{
			name: "Multiplication(7, 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.Multiplication(7, 12) },
			must: func() pitchclass.Operator { return pitchclass.MustMultiplication(7, 12) },
			want: opSpec{m: 7, n: 0, size: 12},
		},
		{
			name: "Multiplication(-7, 12)",
			call: func() (pitchclass.Operator, error) { return pitchclass.Multiplication(-7, 12) },
			must: func() pitchclass.Operator { return pitchclass.MustMultiplication(-7, 12) },
			want: opSpec{m: 5, n: 0, size: 12},
		},
		{
			name: "Multiplication(2, 1)",
			call: func() (pitchclass.Operator, error) { return pitchclass.Multiplication(2, 1) },
			must: func() pitchclass.Operator { return pitchclass.MustMultiplication(2, 1) },
			want: opSpec{m: 0, n: 0, size: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := tc.call()
			require.NoError(t, err)
			assertOperator(t, tc.want, f, tc.name)
			assert.Equal(t, tc.want.isTransposition(), f.IsTransposition(), "IsTransposition")
			assert.Equal(t, tc.want.isInversion(), f.IsInversion(), "IsInversion")
			assert.True(t, f.Equal(tc.must()), "Must variant gives the same operator")
		})
	}
}

// K-18: equal operators from different constructors and the kind of the
// operator.
func TestOperatorCreationEquivalences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b pitchclass.Operator
	}{
		{name: "(11, 5) is I5", a: res(pitchclass.NewOperator(11, 5, 12)).get(t), b: res(pitchclass.Inversion(5, 12)).get(t)},
		{name: "axis 2 is I4", a: res(pitchclass.InversionAround(newClass(t, 2, 12))).get(t), b: res(pitchclass.Inversion(4, 12)).get(t)},
		{name: "pair (0, 4) is I4", a: res(pitchclass.InversionSwapping(newClass(t, 0, 12), newClass(t, 4, 12))).get(t), b: res(pitchclass.Inversion(4, 12)).get(t)},
		{name: "(7, 0) is M7", a: res(pitchclass.NewOperator(7, 0, 12)).get(t), b: res(pitchclass.Multiplication(7, 12)).get(t)},
		{name: "(17, -7) is (5, 5)", a: res(pitchclass.NewOperator(17, -7, 12)).get(t), b: res(pitchclass.NewOperator(5, 5, 12)).get(t)},
		{name: "identity is T0", a: res(pitchclass.Identity(12)).get(t), b: res(pitchclass.Transposition(0, 12)).get(t)},
		{name: "(1, 5) is T5", a: res(pitchclass.NewOperator(1, 5, 12)).get(t), b: res(pitchclass.Transposition(5, 12)).get(t)},
		{name: "N=1: (0, 0) is (5, 3)", a: res(pitchclass.NewOperator(0, 0, 1)).get(t), b: res(pitchclass.NewOperator(5, 3, 1)).get(t)},
		{name: "N=1: identity is I0", a: res(pitchclass.Identity(1)).get(t), b: res(pitchclass.Inversion(0, 1)).get(t)},
		{name: "N=2: T1 is I1", a: res(pitchclass.Transposition(1, 2)).get(t), b: res(pitchclass.Inversion(1, 2)).get(t)},
		{name: "N=2: T0 is I0", a: res(pitchclass.Transposition(0, 2)).get(t), b: res(pitchclass.Inversion(0, 2)).get(t)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, tc.a.Equal(tc.b))
			assert.True(t, tc.b.Equal(tc.a))
			assert.True(t, tc.a == tc.b, "== agrees with Equal on valid operators")
		})
	}

	kinds := []struct {
		name                     string
		f                        pitchclass.Operator
		transposition, inversion bool
	}{
		{name: "T5 at 12", f: res(pitchclass.Transposition(5, 12)).get(t), transposition: true},
		{name: "I5 at 12", f: res(pitchclass.Inversion(5, 12)).get(t), inversion: true},
		{name: "M5 at 12", f: res(pitchclass.Multiplication(5, 12)).get(t)},
		{name: "T5M7 at 12", f: res(pitchclass.NewOperator(7, 5, 12)).get(t)},
		{name: "T1 at 2", f: res(pitchclass.Transposition(1, 2)).get(t), transposition: true, inversion: true},
		{name: "T0 at 2", f: res(pitchclass.Identity(2)).get(t), transposition: true, inversion: true},
		{name: "the operator at 1", f: res(pitchclass.NewOperator(0, 0, 1)).get(t), transposition: true, inversion: true},
		{name: "I0 at 3", f: res(pitchclass.Inversion(0, 3)).get(t), inversion: true},
		{name: "T0 at MAX", f: res(pitchclass.Identity(maxInt)).get(t), transposition: true},
		{name: "I0 at MAX", f: res(pitchclass.Inversion(0, maxInt)).get(t), inversion: true},
	}

	for _, tc := range kinds {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.transposition, tc.f.IsTransposition(), "IsTransposition")
			assert.Equal(t, tc.inversion, tc.f.IsInversion(), "IsInversion")
		})
	}
}

// K-18, E-2, E-3, E-4: creation errors and the zero Operator on an error.
func TestOperatorCreationErrors(t *testing.T) {
	t.Parallel()

	c12, c19 := newClass(t, 0, 12), newClass(t, 4, 19)

	tests := []struct {
		name string
		want error
		call func() (pitchclass.Operator, error)
		must func()
	}{
		{name: "NewOperator(2, 0, 12)", want: pitchclass.ErrNotInvertible, call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(2, 0, 12) }, must: func() { pitchclass.MustNewOperator(2, 0, 12) }},
		{name: "NewOperator(0, 0, 12)", want: pitchclass.ErrNotInvertible, call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(0, 0, 12) }, must: func() { pitchclass.MustNewOperator(0, 0, 12) }},
		{name: "NewOperator(6, 1, 12)", want: pitchclass.ErrNotInvertible, call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(6, 1, 12) }, must: func() { pitchclass.MustNewOperator(6, 1, 12) }},
		{name: "NewOperator(12, 0, 12)", want: pitchclass.ErrNotInvertible, call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(12, 0, 12) }, must: func() { pitchclass.MustNewOperator(12, 0, 12) }},
		{name: "NewOperator(MIN, MAX, 12)", want: pitchclass.ErrNotInvertible, call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(minInt, maxInt, 12) }, must: func() { pitchclass.MustNewOperator(minInt, maxInt, 12) }},
		{name: "NewOperator(7, 0, MAX)", want: pitchclass.ErrNotInvertible, call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(7, 0, maxInt) }, must: func() { pitchclass.MustNewOperator(7, 0, maxInt) }},
		{name: "NewOperator(1, 0, 0)", want: pitchclass.ErrInvalidModulus, call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(1, 0, 0) }, must: func() { pitchclass.MustNewOperator(1, 0, 0) }},
		{name: "NewOperator(1, 0, -1)", want: pitchclass.ErrInvalidModulus, call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(1, 0, -1) }, must: func() { pitchclass.MustNewOperator(1, 0, -1) }},
		{name: "NewOperator(2, 0, 0) checks N first", want: pitchclass.ErrInvalidModulus, call: func() (pitchclass.Operator, error) { return pitchclass.NewOperator(2, 0, 0) }, must: func() { pitchclass.MustNewOperator(2, 0, 0) }},
		{name: "Identity(0)", want: pitchclass.ErrInvalidModulus, call: func() (pitchclass.Operator, error) { return pitchclass.Identity(0) }, must: func() { pitchclass.MustIdentity(0) }},
		{name: "Identity(-1)", want: pitchclass.ErrInvalidModulus, call: func() (pitchclass.Operator, error) { return pitchclass.Identity(-1) }, must: func() { pitchclass.MustIdentity(-1) }},
		{name: "Transposition(1, 0)", want: pitchclass.ErrInvalidModulus, call: func() (pitchclass.Operator, error) { return pitchclass.Transposition(1, 0) }, must: func() { pitchclass.MustTransposition(1, 0) }},
		{name: "Transposition(1, -1)", want: pitchclass.ErrInvalidModulus, call: func() (pitchclass.Operator, error) { return pitchclass.Transposition(1, -1) }, must: func() { pitchclass.MustTransposition(1, -1) }},
		{name: "Inversion(1, 0)", want: pitchclass.ErrInvalidModulus, call: func() (pitchclass.Operator, error) { return pitchclass.Inversion(1, 0) }, must: func() { pitchclass.MustInversion(1, 0) }},
		{name: "Inversion(1, MIN)", want: pitchclass.ErrInvalidModulus, call: func() (pitchclass.Operator, error) { return pitchclass.Inversion(1, minInt) }, must: func() { pitchclass.MustInversion(1, minInt) }},
		{name: "Multiplication(1, 0)", want: pitchclass.ErrInvalidModulus, call: func() (pitchclass.Operator, error) { return pitchclass.Multiplication(1, 0) }, must: func() { pitchclass.MustMultiplication(1, 0) }},
		{name: "Multiplication(2, -1) checks N first", want: pitchclass.ErrInvalidModulus, call: func() (pitchclass.Operator, error) { return pitchclass.Multiplication(2, -1) }, must: func() { pitchclass.MustMultiplication(2, -1) }},
		{name: "Multiplication(2, 12)", want: pitchclass.ErrNotInvertible, call: func() (pitchclass.Operator, error) { return pitchclass.Multiplication(2, 12) }, must: func() { pitchclass.MustMultiplication(2, 12) }},
		{name: "Multiplication(0, 2)", want: pitchclass.ErrNotInvertible, call: func() (pitchclass.Operator, error) { return pitchclass.Multiplication(0, 2) }, must: func() { pitchclass.MustMultiplication(0, 2) }},
		{name: "InversionSwapping(0 of 12, 4 of 19)", want: pitchclass.ErrModulusMismatch, call: func() (pitchclass.Operator, error) { return pitchclass.InversionSwapping(c12, c19) }, must: func() { pitchclass.MustInversionSwapping(c12, c19) }},
		{name: "InversionSwapping(4 of 19, 0 of 12)", want: pitchclass.ErrModulusMismatch, call: func() (pitchclass.Operator, error) { return pitchclass.InversionSwapping(c19, c12) }, must: func() { pitchclass.MustInversionSwapping(c19, c12) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := tc.call()
			assertOnlyError(t, err, tc.want, tc.name)
			assert.Equal(t, pitchclass.Operator{}, f, "the zero Operator on an error")
			assertPanicsWithErrorIs(t, tc.want, tc.must, tc.name)
		})
	}
}

// K-18, C-22: for N in {1, 2, 12, 19} and m, n in -30...30 the constructors
// agree with the oracle: m and n are reduced by the circle.
func TestOperatorCreationOracle(t *testing.T) {
	t.Parallel()

	for _, size := range []int{1, 2, 12, 19} {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			values := append(ints(-30, 30), edgeInts()...)
			for _, m := range values {
				for _, n := range values {
					desc := fmt.Sprintf("NewOperator(%d, %d, %d)", m, n, size)
					f, err := pitchclass.NewOperator(m, n, size)
					if !invertibleOracle(m, size) {
						assertOnlyError(t, err, pitchclass.ErrNotInvertible, desc)

						continue
					}

					require.NoError(t, err, desc)
					want := newOpSpec(m, n, size)
					assertOperator(t, want, f, desc)
					assert.Equal(t, want.isTransposition(), f.IsTransposition(), "%s: IsTransposition", desc)
					assert.Equal(t, want.isInversion(), f.IsInversion(), "%s: IsInversion", desc)
				}

				assertOperator(t, transpositionSpec(m, size), res(pitchclass.Transposition(m, size)).get(t), fmt.Sprintf("Transposition(%d, %d)", m, size))
				assertOperator(t, inversionSpec(m, size), res(pitchclass.Inversion(m, size)).get(t), fmt.Sprintf("Inversion(%d, %d)", m, size))
				if invertibleOracle(m, size) {
					assertOperator(t, newOpSpec(m, 0, size), res(pitchclass.Multiplication(m, size)).get(t), fmt.Sprintf("Multiplication(%d, %d)", m, size))
				}
			}

			assertOperator(t, identitySpec(size), res(pitchclass.Identity(size)).get(t), "Identity")

			cs := classesOf(t, size)
			for _, a := range cs {
				assertOperator(t, newOpSpec(size-1, 2*a.Number(), size), res(pitchclass.InversionAround(a)).get(t), fmt.Sprintf("InversionAround(%v)", a))
				for _, b := range cs {
					assertOperator(t, newOpSpec(size-1, a.Number()+b.Number(), size), res(pitchclass.InversionSwapping(a, b)).get(t),
						fmt.Sprintf("InversionSwapping(%v, %v)", a, b))
				}
			}
		})
	}
}

// K-19, C-23, E-3: f(x) = (m*x + n) mod N; every operator permutes the N
// classes; an operator and a class of different N are ErrModulusMismatch.
func TestOperatorApply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    pitchclass.Operator
		spec opSpec
		x    int
		want int
	}{
		{name: "T5M7(1) at 12", f: res(pitchclass.NewOperator(7, 5, 12)).get(t), spec: opSpec{m: 7, n: 5, size: 12}, x: 1, want: 0},
		{name: "I8(2) at 12", f: res(pitchclass.Inversion(8, 12)).get(t), spec: inversionSpec(8, 12), x: 2, want: 6},
		{name: "T3(11) at 12", f: res(pitchclass.Transposition(3, 12)).get(t), spec: transpositionSpec(3, 12), x: 11, want: 2},
		{name: "M7(2) at 12", f: res(pitchclass.Multiplication(7, 12)).get(t), spec: opSpec{m: 7, n: 0, size: 12}, x: 2, want: 2},
		{
			name: "T(MAX-1)M(MAX-1)(MAX-1) at MAX", f: res(pitchclass.NewOperator(maxInt-1, maxInt-1, maxInt)).get(t),
			spec: opSpec{m: maxInt - 1, n: maxInt - 1, size: maxInt}, x: maxInt - 1, want: 0,
		},
		{name: "T1(MAX-1) at MAX", f: res(pitchclass.Transposition(1, maxInt)).get(t), spec: transpositionSpec(1, maxInt), x: maxInt - 1, want: 0},
		{name: "T0M2(MAX-1) at MAX", f: res(pitchclass.Multiplication(2, maxInt)).get(t), spec: opSpec{m: 2, n: 0, size: maxInt}, x: maxInt - 1, want: maxInt - 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			size := tc.spec.size
			require.Equal(t, tc.spec.apply(tc.x), tc.want, "test table disagrees with the oracle")

			x := newClass(t, tc.x, size)
			got, err := tc.f.Apply(x)
			require.NoError(t, err)
			assertClass(t, tc.want, size, got, "Apply")
			assertClass(t, tc.want, size, tc.f.MustApply(x), "MustApply")
		})
	}

	t.Run("all operators of 12", func(t *testing.T) {
		t.Parallel()

		ops, specs := operatorsOf(t, 12)
		require.Len(t, ops, 48)
		for i, f := range ops {
			images := map[pitchclass.Class]bool{}
			for _, x := range classesOf(t, 12) {
				got, err := f.Apply(x)
				require.NoError(t, err)
				assertClass(t, specs[i].apply(x.Number()), 12, got, fmt.Sprintf("%v(%v)", specs[i], x))
				images[got] = true
			}
			assert.Len(t, images, 12, "%v permutes the classes", specs[i])
		}
	})

	t.Run("different N", func(t *testing.T) {
		t.Parallel()

		f12 := res(pitchclass.Transposition(1, 12)).get(t)
		f19 := res(pitchclass.Transposition(1, 19)).get(t)
		c12, c19 := newClass(t, 1, 12), newClass(t, 1, 19)

		for _, tc := range []struct {
			name string
			f    pitchclass.Operator
			x    pitchclass.Class
		}{
			{name: "operator of 12, class of 19", f: f12, x: c19},
			{name: "operator of 19, class of 12", f: f19, x: c12},
		} {
			got, err := tc.f.Apply(tc.x)
			assertOnlyError(t, err, pitchclass.ErrModulusMismatch, tc.name)
			assert.Equal(t, pitchclass.Class{}, got, "the zero Class on an error")
			assertPanicsWithErrorIs(t, pitchclass.ErrModulusMismatch, func() { tc.f.MustApply(tc.x) }, tc.name)
		}
	})
}

// K-20, C-24, E-3: f∘g applies g first, then f.
func TestOperatorCompose(t *testing.T) {
	t.Parallel()

	t3 := res(pitchclass.Transposition(3, 12)).get(t)
	i8 := res(pitchclass.Inversion(8, 12)).get(t)

	tests := []struct {
		name string
		f, g pitchclass.Operator
		want pitchclass.Operator
	}{
		{name: "T3∘I8 = I11", f: t3, g: i8, want: res(pitchclass.Inversion(11, 12)).get(t)},
		{name: "I8∘T3 = I5", f: i8, g: t3, want: res(pitchclass.Inversion(5, 12)).get(t)},
		{name: "I8∘I3 = T5", f: i8, g: res(pitchclass.Inversion(3, 12)).get(t), want: res(pitchclass.Transposition(5, 12)).get(t)},
		{name: "T5∘T9 = T2", f: res(pitchclass.Transposition(5, 12)).get(t), g: res(pitchclass.Transposition(9, 12)).get(t), want: res(pitchclass.Transposition(2, 12)).get(t)},
		{name: "T5∘I0 = I5", f: res(pitchclass.Transposition(5, 12)).get(t), g: res(pitchclass.Inversion(0, 12)).get(t), want: res(pitchclass.Inversion(5, 12)).get(t)},
		{name: "M5∘M5 = T0", f: res(pitchclass.Multiplication(5, 12)).get(t), g: res(pitchclass.Multiplication(5, 12)).get(t), want: res(pitchclass.Identity(12)).get(t)},
		{name: "M7∘M7 = T0", f: res(pitchclass.Multiplication(7, 12)).get(t), g: res(pitchclass.Multiplication(7, 12)).get(t), want: res(pitchclass.Identity(12)).get(t)},
		{name: "M5∘M7 = I0", f: res(pitchclass.Multiplication(5, 12)).get(t), g: res(pitchclass.Multiplication(7, 12)).get(t), want: res(pitchclass.Inversion(0, 12)).get(t)},
		{name: "M7 = I0∘M5", f: res(pitchclass.Inversion(0, 12)).get(t), g: res(pitchclass.Multiplication(5, 12)).get(t), want: res(pitchclass.Multiplication(7, 12)).get(t)},
		{name: "T5M7∘T2M5 = I7", f: res(pitchclass.NewOperator(7, 5, 12)).get(t), g: res(pitchclass.NewOperator(5, 2, 12)).get(t), want: res(pitchclass.Inversion(7, 12)).get(t)},
		{name: "T1∘T(MAX-1) at MAX = T0", f: res(pitchclass.Transposition(1, maxInt)).get(t), g: res(pitchclass.Transposition(maxInt-1, maxInt)).get(t), want: res(pitchclass.Identity(maxInt)).get(t)},
		{name: "I(MAX-1)∘I(MAX-1) at MAX = T0", f: res(pitchclass.Inversion(maxInt-1, maxInt)).get(t), g: res(pitchclass.Inversion(maxInt-1, maxInt)).get(t), want: res(pitchclass.Identity(maxInt)).get(t)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.f.Compose(tc.g)
			require.NoError(t, err)
			assert.True(t, got.Equal(tc.want), "%v, want %v", got, tc.want)
			assert.True(t, tc.f.MustCompose(tc.g).Equal(tc.want), "MustCompose")
		})
	}

	t.Run("not commutative", func(t *testing.T) {
		t.Parallel()

		assert.False(t, res(t3.Compose(i8)).get(t).Equal(res(i8.Compose(t3)).get(t)), "T3∘I8 != I8∘T3")
	})

	t.Run("different N", func(t *testing.T) {
		t.Parallel()

		f12, f19 := res(pitchclass.Transposition(5, 12)).get(t), res(pitchclass.Transposition(5, 19)).get(t)
		for _, tc := range []struct {
			name string
			f, g pitchclass.Operator
		}{
			{name: "12∘19", f: f12, g: f19},
			{name: "19∘12", f: f19, g: f12},
		} {
			got, err := tc.f.Compose(tc.g)
			assertOnlyError(t, err, pitchclass.ErrModulusMismatch, tc.name)
			assert.Equal(t, pitchclass.Operator{}, got, "the zero Operator on an error")
			assertPanicsWithErrorIs(t, pitchclass.ErrModulusMismatch, func() { tc.f.MustCompose(tc.g) }, tc.name)
		}
	})
}

// K-20: for all pairs of operators of N in {1, 2, 7, 12} and all x,
// (f∘g)(x) = f(g(x)) and f∘g agrees with the oracle; composition is
// associative on all triples of N in {1, 2, 7, 12}.
func TestOperatorComposeRelation(t *testing.T) {
	t.Parallel()

	for _, size := range []int{1, 2, 7, 12} {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			ops, specs := operatorsOf(t, size)
			cs := classesOf(t, size)
			composed := make([][]pitchclass.Operator, len(ops))
			for i, f := range ops {
				composed[i] = make([]pitchclass.Operator, len(ops))
				for j, g := range ops {
					fg, err := f.Compose(g)
					require.NoError(t, err)
					assertOperator(t, specs[i].compose(specs[j]), fg, fmt.Sprintf("%v∘%v", specs[i], specs[j]))
					composed[i][j] = fg

					for _, x := range cs {
						gx, err1 := g.Apply(x)
						fgx, err2 := f.Apply(gx)
						got, err3 := fg.Apply(x)
						if err1 != nil || err2 != nil || err3 != nil || !got.Equal(fgx) {
							assert.Fail(t, "composition", "(%v∘%v)(%v) = %v, f(g(x)) = %v; errors %v, %v, %v",
								specs[i], specs[j], x, got, fgx, err1, err2, err3)
						}
					}
				}
			}

			for i := range ops {
				for j := range ops {
					for k, h := range ops {
						left, err1 := composed[i][j].Compose(h)
						right, err2 := ops[i].Compose(composed[j][k])
						if err1 != nil || err2 != nil || !left.Equal(right) {
							assert.Fail(t, "associativity", "(%v∘%v)∘%v = %v, %v∘(%v∘%v) = %v; errors %v, %v",
								specs[i], specs[j], specs[k], left, specs[i], specs[j], specs[k], right, err1, err2)
						}
					}
				}
			}
		})
	}
}

// K-21, C-25: the inverse operator.
func TestOperatorInverse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		f, want pitchclass.Operator
	}{
		{name: "(T5M7)^-1 = T1M7", f: res(pitchclass.NewOperator(7, 5, 12)).get(t), want: res(pitchclass.NewOperator(7, 1, 12)).get(t)},
		{name: "M5^-1 = M5", f: res(pitchclass.Multiplication(5, 12)).get(t), want: res(pitchclass.Multiplication(5, 12)).get(t)},
		{name: "M7^-1 = M7", f: res(pitchclass.Multiplication(7, 12)).get(t), want: res(pitchclass.Multiplication(7, 12)).get(t)},
		{name: "T0^-1 = T0", f: res(pitchclass.Identity(12)).get(t), want: res(pitchclass.Identity(12)).get(t)},
		{name: "T1^-1 = T(MAX-1) at MAX", f: res(pitchclass.Transposition(1, maxInt)).get(t), want: res(pitchclass.Transposition(maxInt-1, maxInt)).get(t)},
		{name: "M2^-1 at MAX", f: res(pitchclass.Multiplication(2, maxInt)).get(t), want: res(pitchclass.Multiplication(4611686018427387904, maxInt)).get(t)},
		{name: "the operator at 1", f: res(pitchclass.Identity(1)).get(t), want: res(pitchclass.Identity(1)).get(t)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, tc.f.Inverse().Equal(tc.want), "%v, want %v", tc.f.Inverse(), tc.want)
		})
	}

	t.Run("Tn and In at 12", func(t *testing.T) {
		t.Parallel()

		for n := range 12 {
			assert.True(t, res(pitchclass.Transposition(n, 12)).get(t).Inverse().Equal(res(pitchclass.Transposition(-n, 12)).get(t)), "T%d^-1 = T-%d", n, n)
			assert.True(t, res(pitchclass.Inversion(n, 12)).get(t).Inverse().Equal(res(pitchclass.Inversion(n, 12)).get(t)), "I%d^-1 = I%d", n, n)
		}
	})

	for _, size := range []int{1, 2, 7, 12, 24} {
		t.Run(fmt.Sprintf("all operators N=%d", size), func(t *testing.T) {
			t.Parallel()

			id := res(pitchclass.Identity(size)).get(t)
			ops, specs := operatorsOf(t, size)
			for i, f := range ops {
				inv := f.Inverse()
				assertOperator(t, specs[i].inverse(), inv, fmt.Sprintf("%v^-1", specs[i]))
				assert.True(t, res(f.Compose(inv)).get(t).Equal(id), "%v∘%v^-1", specs[i], specs[i])
				assert.True(t, res(inv.Compose(f)).get(t).Equal(id), "%v^-1∘%v", specs[i], specs[i])
			}
		})
	}
}

// K-22, C-26: equality and order of operators.
func TestOperatorEqualAndCompare(t *testing.T) {
	t.Parallel()

	t.Run("equal", func(t *testing.T) {
		t.Parallel()

		same := []pitchclass.Operator{
			res(pitchclass.InversionAround(newClass(t, 2, 12))).get(t),
			res(pitchclass.InversionSwapping(newClass(t, 1, 12), newClass(t, 3, 12))).get(t),
			res(pitchclass.NewOperator(11, 4, 12)).get(t),
		}
		for _, a := range same {
			for _, b := range same {
				assert.True(t, a.Equal(b), "%v and %v", a, b)
				assert.Zero(t, a.Compare(b), "%v and %v", a, b)
			}
		}
	})

	t.Run("not equal", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name string
			a, b pitchclass.Operator
		}{
			{name: "T5 at 12 and 19", a: res(pitchclass.Transposition(5, 12)).get(t), b: res(pitchclass.Transposition(5, 19)).get(t)},
			{name: "T5 and I5 at 12", a: res(pitchclass.Transposition(5, 12)).get(t), b: res(pitchclass.Inversion(5, 12)).get(t)},
		} {
			assert.False(t, tc.a.Equal(tc.b), tc.name)
			assert.False(t, tc.b.Equal(tc.a), tc.name)
			assert.False(t, tc.a == tc.b, tc.name)
			assert.NotZero(t, tc.a.Compare(tc.b), tc.name)
		}
	})

	t.Run("relation", func(t *testing.T) {
		t.Parallel()

		var ops []pitchclass.Operator
		var specs []opSpec
		for _, size := range []int{1, 2, 12, 19} {
			o, s := operatorsOf(t, size)
			ops = append(ops, o...)
			specs = append(specs, s...)
		}
		require.Len(t, ops, 1+2+48+342)

		for i, a := range ops {
			for j, b := range ops {
				got, back, eq := a.Compare(b), b.Compare(a), a.Equal(b)
				if got != compareSpec(specs[i], specs[j]) || eq != (specs[i] == specs[j]) || eq != (got == 0) || back != -got || (a == b) != eq {
					assert.Fail(t, "operator order", "%v and %v: Compare %d, reverse %d, Equal %t, == %t",
						specs[i], specs[j], got, back, eq, a == b)
				}
			}
		}

		// ops are in ascending order, so each ordered index triple is an
		// instance of a <= b <= c (checked) and transitivity requires a <= c.
		checked := 0
		for _, tr := range triples(len(ops), 3000) {
			a, b, c := ops[tr[0]], ops[tr[1]], ops[tr[2]]
			if a.Compare(b) <= 0 && b.Compare(c) <= 0 {
				checked++
				assert.LessOrEqual(t, a.Compare(c), 0, "transitivity %v, %v, %v", specs[tr[0]], specs[tr[1]], specs[tr[2]])
			}
		}
		assert.GreaterOrEqual(t, checked, 1000, "transitivity is checked on at least 1000 triples")

		sorted := slices.Clone(ops)
		for _, perm := range permutations(len(ops)) {
			xs := make([]pitchclass.Operator, len(ops))
			for i, p := range perm {
				xs[i] = ops[p]
			}

			slices.SortFunc(xs, pitchclass.Operator.Compare)
			assert.Equal(t, sorted, xs, "sorted permutation")
		}
	})

	t.Run("transitivity at 1 and 2", func(t *testing.T) {
		t.Parallel()

		ops := append(operatorsOfOnly(t, 1), operatorsOfOnly(t, 2)...)
		for _, a := range ops {
			for _, b := range ops {
				for _, c := range ops {
					if a.Compare(b) <= 0 && b.Compare(c) <= 0 {
						assert.LessOrEqual(t, a.Compare(c), 0)
					}
				}
			}
		}
	})

	t.Run("order of Operators at 12", func(t *testing.T) {
		t.Parallel()

		listed := slices.Collect(res(pitchclass.Operators(12)).get(t))
		require.Len(t, listed, 48)
		for i := 1; i < len(listed); i++ {
			assert.Equal(t, -1, listed[i-1].Compare(listed[i]), "%v before %v", listed[i-1], listed[i])
		}
	})
}

// operatorsOfOnly returns the operators of operatorsOf without the specs.
func operatorsOfOnly(t testing.TB, size int) []pitchclass.Operator {
	t.Helper()

	ops, _ := operatorsOf(t, size)

	return ops
}

// triples returns count deterministic index triples i <= j <= k in 0...n-1
// from a linear congruential sequence; it uses no global random state. Over a
// slice in ascending order every triple satisfies the premise of
// transitivity.
func triples(n, count int) [][3]int {
	out := make([][3]int, 0, count)
	state := uint64(0x9E3779B97F4A7C15)
	next := func() int {
		state = state*6364136223846793005 + 1442695040888963407

		return int((state >> 33) % uint64(n))
	}

	for range count {
		tr := [3]int{next(), next(), next()}
		slices.Sort(tr[:])
		out = append(out, tr)
	}

	return out
}

// K-31 (operators), NF-2: equal operators are one map key.
func TestOperatorMapKey(t *testing.T) {
	t.Parallel()

	m := map[pitchclass.Operator]int{}
	m[res(pitchclass.InversionAround(newClass(t, 2, 12))).get(t)]++
	m[res(pitchclass.InversionSwapping(newClass(t, 0, 12), newClass(t, 4, 12))).get(t)]++
	m[res(pitchclass.NewOperator(11, 4, 12)).get(t)]++
	m[res(pitchclass.Transposition(5, 12)).get(t)]++
	m[res(pitchclass.Transposition(5, 19)).get(t)]++

	assert.Len(t, m, 3)
	assert.Equal(t, 3, m[res(pitchclass.Inversion(4, 12)).get(t)])
	assert.Equal(t, 1, m[res(pitchclass.Transposition(5, 12)).get(t)])
	assert.Equal(t, 1, m[res(pitchclass.Transposition(5, 19)).get(t)])
}
