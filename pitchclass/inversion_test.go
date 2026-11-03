package pitchclass_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-12, C-14, G-4: In(x) = (n - x) mod N for any int n; I0 is the inversion
// without an index.
func TestClassInvert(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		x, index, size int
		want           int
	}{
		{name: "I8(2) at 12 (OMT)", x: 2, index: 8, size: 12, want: 6},
		{name: "I8(4) at 12 (OMT)", x: 4, index: 8, size: 12, want: 4},
		{name: "I8(7) at 12 (OMT)", x: 7, index: 8, size: 12, want: 1},
		{name: "I3(4) at 12 (MT21: T3I(E) = B)", x: 4, index: 3, size: 12, want: 11},
		{name: "I0(4) at 12", x: 4, index: 0, size: 12, want: 8},
		{name: "I0(0) at 12", x: 0, index: 0, size: 12, want: 0},
		{name: "I-1(0) at 12", x: 0, index: -1, size: 12, want: 11},
		{name: "IMAX(0) at 12", x: 0, index: maxInt, size: 12, want: 7},
		{name: "IMIN(0) at 12", x: 0, index: minInt, size: 12, want: 4},
		{name: "I0(1) at MAX", x: 1, index: 0, size: maxInt, want: maxInt - 1},
		{name: "IMIN(1) at MAX", x: 1, index: minInt, size: maxInt, want: maxInt - 2},
		{name: "IMAX(MAX-1) at MAX", x: maxInt - 1, index: maxInt, size: maxInt, want: 1},
		{name: "I5(0) at 1", x: 0, index: 5, size: 1, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, subMod(tc.index, tc.x, tc.size), tc.want, "test table disagrees with the oracle")
			assertClass(t, tc.want, tc.size, newClass(t, tc.x, tc.size).Invert(tc.index), "Invert")
		})
	}
}

// K-12: for all x, y, n in 0...N-1 at N in {1, 2, 12, 19}: In(In(x)) = x,
// ic(In x, In y) = ic(x, y), i<In x, In y> = i<y, x>, In(x) = Tn(I0(x));
// for n in -30...30 and edge values In agrees with the oracle.
func TestClassInvertRelation(t *testing.T) {
	t.Parallel()

	for _, size := range []int{1, 2, 12, 19} {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			cs := classesOf(t, size)
			for n := range size {
				for _, x := range cs {
					ix := x.Invert(n)
					assert.True(t, ix.Invert(n).Equal(x), "I%d(I%d(%v))", n, n, x)
					assert.True(t, ix.Equal(x.Invert(0).Transpose(n)), "I%d(%v) = T%d(I0(%v))", n, x, n, x)

					for _, y := range cs {
						iy := y.Invert(n)

						icInv, err := ix.IntervalClass(iy)
						require.NoError(t, err)
						icOrig, err := x.IntervalClass(y)
						require.NoError(t, err)
						assert.Equal(t, icOrig, icInv, "ic is invariant under I%d for %v, %v", n, x, y)

						ordInv, err := ix.IntervalTo(iy)
						require.NoError(t, err)
						ordBack, err := y.IntervalTo(x)
						require.NoError(t, err)
						assert.Equal(t, ordBack, ordInv, "I%d reverses the interval of %v, %v", n, x, y)
					}
				}
			}

			for _, x := range cs {
				for _, n := range append(ints(-30, 30), edgeInts()...) {
					assertClass(t, subMod(n, x.Number(), size), size, x.Invert(n), fmt.Sprintf("%v.Invert(%d)", x, n))
				}
			}
		})
	}
}

// K-13, C-15, C-16: inversion around an axis class a is I(2a); the
// inversion of a pair x, y is I(x+y) and swaps x and y.
func TestClassInvertAroundAndSwapping(t *testing.T) {
	t.Parallel()

	t.Run("axis 2 at 12", func(t *testing.T) {
		t.Parallel()

		axis := newClass(t, 2, 12)
		for _, tc := range []struct{ x, want int }{{0, 4}, {1, 3}, {2, 2}, {8, 8}} {
			got, err := newClass(t, tc.x, 12).InvertAround(axis)
			require.NoError(t, err)
			assertClass(t, tc.want, 12, got, fmt.Sprintf("%d around 2", tc.x))
			assertClass(t, tc.want, 12, newClass(t, tc.x, 12).MustInvertAround(axis), "MustInvertAround")
		}

		for _, x := range classesOf(t, 12) {
			got, err := x.InvertAround(axis)
			require.NoError(t, err)
			assert.True(t, got.Equal(x.Invert(4)), "%v around 2 = I4", x)

			got, err = x.InvertAround(newClass(t, 0, 12))
			require.NoError(t, err)
			assert.True(t, got.Equal(x.Invert(0)), "%v around 0 = I0", x)
		}
	})

	t.Run("pair (0, 4) at 12", func(t *testing.T) {
		t.Parallel()

		p, q := newClass(t, 0, 12), newClass(t, 4, 12)
		for _, tc := range []struct{ x, want int }{{0, 4}, {4, 0}, {7, 9}} {
			got, err := newClass(t, tc.x, 12).InvertSwapping(p, q)
			require.NoError(t, err)
			assertClass(t, tc.want, 12, got, fmt.Sprintf("%d swapping 0 and 4", tc.x))
			assertClass(t, tc.want, 12, newClass(t, tc.x, 12).MustInvertSwapping(p, q), "MustInvertSwapping")
		}
	})

	t.Run("pairs of I8 at 12", func(t *testing.T) {
		t.Parallel()

		for _, pair := range [][2]int{{3, 5}, {2, 6}, {4, 4}, {5, 3}, {0, 8}} {
			p, q := newClass(t, pair[0], 12), newClass(t, pair[1], 12)
			for _, x := range classesOf(t, 12) {
				got, err := x.InvertSwapping(p, q)
				require.NoError(t, err)
				assert.True(t, got.Equal(x.Invert(8)), "%v swapping %v and %v = I8", x, p, q)
			}
		}
	})

	for _, size := range []int{1, 2, 12, 19} {
		t.Run(fmt.Sprintf("relation N=%d", size), func(t *testing.T) {
			t.Parallel()

			cs := classesOf(t, size)
			for _, a := range cs {
				fixed, err := a.InvertAround(a)
				require.NoError(t, err)
				assert.True(t, fixed.Equal(a), "the axis %v stays in place", a)

				for _, x := range cs {
					around, err := x.InvertAround(a)
					require.NoError(t, err)
					assertClass(t, subMod(2*a.Number(), x.Number(), size), size, around, fmt.Sprintf("%v around %v", x, a))
					assert.True(t, around.Equal(x.Invert(2*a.Number())), "%v around %v = I(2a)", x, a)

					same, err := x.InvertSwapping(a, a)
					require.NoError(t, err)
					assert.True(t, same.Equal(around), "pair (%v, %v) is the axis %v", a, a, a)

					for _, b := range cs {
						swapped, err := x.InvertSwapping(a, b)
						require.NoError(t, err)
						assertClass(t, subMod(a.Number()+b.Number(), x.Number(), size), size, swapped,
							fmt.Sprintf("%v swapping %v and %v", x, a, b))
					}

					got, err := a.InvertSwapping(a, x)
					require.NoError(t, err)
					assert.True(t, got.Equal(x), "%v swapped with %v gives %v", a, x, x)
				}
			}
		})
	}

	t.Run("MAX", func(t *testing.T) {
		t.Parallel()

		last := newClass(t, maxInt-1, maxInt)
		one := newClass(t, 1, maxInt)
		require.Equal(t, subMod(addMod(maxInt-1, maxInt-1, maxInt), 1, maxInt), maxInt-3, "the oracle")

		got, err := one.InvertAround(last)
		require.NoError(t, err)
		assertClass(t, maxInt-3, maxInt, got, "1 around MAX-1")

		got, err = one.InvertSwapping(last, last)
		require.NoError(t, err)
		assertClass(t, maxInt-3, maxInt, got, "1 swapping MAX-1 and MAX-1")
	})

	t.Run("different N", func(t *testing.T) {
		t.Parallel()

		c12, c19 := newClass(t, 1, 12), newClass(t, 1, 19)
		tests := []struct {
			name string
			call func() (pitchclass.Class, error)
			must func()
		}{
			{
				name: "axis of 19 for a class of 12",
				call: func() (pitchclass.Class, error) { return c12.InvertAround(c19) },
				must: func() { c12.MustInvertAround(c19) },
			},
			{
				name: "axis of 12 for a class of 19",
				call: func() (pitchclass.Class, error) { return c19.InvertAround(c12) },
				must: func() { c19.MustInvertAround(c12) },
			},
			{
				name: "pair of 19 for a class of 12",
				call: func() (pitchclass.Class, error) { return c12.InvertSwapping(c19, c19) },
				must: func() { c12.MustInvertSwapping(c19, c19) },
			},
			{
				name: "first of the pair from 19",
				call: func() (pitchclass.Class, error) { return c12.InvertSwapping(c19, c12) },
				must: func() { c12.MustInvertSwapping(c19, c12) },
			},
			{
				name: "second of the pair from 19",
				call: func() (pitchclass.Class, error) { return c12.InvertSwapping(c12, c19) },
				must: func() { c12.MustInvertSwapping(c12, c19) },
			},
			{
				name: "pair of 12 for a class of 19",
				call: func() (pitchclass.Class, error) { return c19.InvertSwapping(c12, c12) },
				must: func() { c19.MustInvertSwapping(c12, c12) },
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := tc.call()
				assertOnlyError(t, err, pitchclass.ErrModulusMismatch, tc.name)
				assert.Equal(t, pitchclass.Class{}, got, "the zero Class on an error")
				assertPanicsWithErrorIs(t, pitchclass.ErrModulusMismatch, tc.must, tc.name)
			})
		}
	})
}

// K-14, C-17, G-3, E-2: the fixed classes of In in ascending order: 0, 1 or
// 2 classes; an empty result is an empty non-nil slice, not an error.
func TestInversionFixedClasses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		index, size int
		want        []int
	}{
		{name: "I0 at 12", index: 0, size: 12, want: []int{0, 6}},
		{name: "I1 at 12", index: 1, size: 12, want: []int{}},
		{name: "I4 at 12", index: 4, size: 12, want: []int{2, 8}},
		{name: "I-2 at 12", index: -2, size: 12, want: []int{5, 11}},
		{name: "I0 at 19", index: 0, size: 19, want: []int{0}},
		{name: "I1 at 19", index: 1, size: 19, want: []int{10}},
		{name: "I3 at 7", index: 3, size: 7, want: []int{5}},
		{name: "I0 at 1", index: 0, size: 1, want: []int{0}},
		{name: "I0 at 2", index: 0, size: 2, want: []int{0, 1}},
		{name: "I1 at 2", index: 1, size: 2, want: []int{}},
		{name: "IMIN at 12", index: minInt, size: 12, want: []int{2, 8}},
		{name: "IMAX at 12", index: maxInt, size: 12, want: []int{}},
		{name: "I0 at MAX", index: 0, size: maxInt, want: []int{0}},
		{name: "I(MAX-1) at MAX", index: maxInt - 1, size: maxInt, want: []int{(maxInt - 1) / 2}},
		{name: "I1 at MAX", index: 1, size: maxInt, want: []int{4611686018427387904}},
		{name: "I(MAX-2) at MAX", index: maxInt - 2, size: maxInt, want: []int{maxInt - 1}},
		{name: "IMIN at MAX", index: minInt, size: maxInt, want: fixedClassesFormula(minInt, maxInt)},
		{name: "I0 at MAX-1", index: 0, size: maxInt - 1, want: []int{0, (maxInt - 1) / 2}},
		{name: "I1 at MAX-1", index: 1, size: maxInt - 1, want: []int{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, fixedClassesFormula(tc.index, tc.size), tc.want, "test table disagrees with the oracle")

			got, err := pitchclass.InversionFixedClasses(tc.index, tc.size)
			require.NoError(t, err)
			assertFixedClasses(t, tc.index, tc.size, tc.want, got)
			assertFixedClasses(t, tc.index, tc.size, tc.want, pitchclass.MustInversionFixedClasses(tc.index, tc.size))
		})
	}

	for _, size := range []int{0, -1, minInt} {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			got, err := pitchclass.InversionFixedClasses(0, size)
			assertOnlyError(t, err, pitchclass.ErrInvalidModulus, "InversionFixedClasses")
			assert.Nil(t, got, "nil slice on an error")
			assertPanicsWithErrorIs(t, pitchclass.ErrInvalidModulus, func() { pitchclass.MustInversionFixedClasses(0, size) }, "MustInversionFixedClasses")
		})
	}
}

// assertFixedClasses checks that got holds exactly the classes with numbers
// want at modulus size, in order, as a non-nil slice with len = cap, and
// that each of them is fixed by I(index).
func assertFixedClasses(t *testing.T, index, size int, want []int, got []pitchclass.Class) {
	t.Helper()

	if !assert.NotNil(t, got, "an empty result is an empty slice, not nil") {
		return
	}

	assert.Equal(t, want, numbersOf(got))
	assert.Equal(t, len(got), cap(got), "len = cap")
	for _, c := range got {
		assert.Equal(t, size, c.Modulus())
		assert.True(t, c.Invert(index).Equal(c), "%v is fixed by I%d", c, index)
	}
}

// K-14: for all N in 1...48 and n in 0...N-1 the fixed classes are exactly
// the classes z with In(z) = z, their number follows P-12.
func TestInversionFixedClassesOracle(t *testing.T) {
	t.Parallel()

	for size := 1; size <= 48; size++ {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			for n := range size {
				want := fixedClassesBrute(n, size)
				require.Equal(t, fixedClassesFormula(n, size), want, "the oracles disagree for I%d at %d", n, size)

				count := 1
				if size%2 == 0 {
					count = 2
					if n%2 == 1 {
						count = 0
					}
				}
				require.Len(t, want, count, "P-12 for I%d at %d", n, size)

				got, err := pitchclass.InversionFixedClasses(n, size)
				require.NoError(t, err)
				assertFixedClasses(t, n, size, want, got)
			}
		})
	}
}
