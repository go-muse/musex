package pitchclass_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-1, C-1, C-3, E-1: strict creation reads the number and N back; numbers
// outside 0...N-1 and N < 1 are ErrInvalidClass with the zero value.
func TestNew(t *testing.T) {
	t.Parallel()

	valid := []struct {
		name      string
		number, n int
	}{
		{name: "0 of 12", number: 0, n: 12},
		{name: "11 of 12", number: 11, n: 12},
		{name: "0 of 1", number: 0, n: 1},
		{name: "MAX-1 of MAX", number: maxInt - 1, n: maxInt},
		{name: "0 of MAX", number: 0, n: maxInt},
	}

	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := pitchclass.New(tc.number, tc.n)
			require.NoError(t, err)
			assertClass(t, tc.number, tc.n, c, "New")
			assertClass(t, tc.number, tc.n, pitchclass.MustNew(tc.number, tc.n), "MustNew")
		})
	}

	invalid := []struct {
		name      string
		number, n int
	}{
		{name: "12 of 12", number: 12, n: 12},
		{name: "-1 of 12", number: -1, n: 12},
		{name: "3 of 2", number: 3, n: 2},
		{name: "0 of 0", number: 0, n: 0},
		{name: "0 of -5", number: 0, n: -5},
		{name: "1 of 1", number: 1, n: 1},
		{name: "MAX of MAX", number: maxInt, n: maxInt},
		{name: "MIN of 12", number: minInt, n: 12},
		{name: "MAX of 12", number: maxInt, n: 12},
		{name: "0 of MIN", number: 0, n: minInt},
		{name: "MIN of MIN", number: minInt, n: minInt},
	}

	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := pitchclass.New(tc.number, tc.n)
			assertOnlyError(t, err, pitchclass.ErrInvalidClass, "New")
			assert.Equal(t, pitchclass.Class{}, c, "the zero Class on an error")
			assert.False(t, c.IsValid())
			assertPanicsWithErrorIs(t, pitchclass.ErrInvalidClass, func() { pitchclass.MustNew(tc.number, tc.n) }, "MustNew")
		})
	}
}

// K-1: for all N in {1, 2, 12, 19, 24, 31} and all k in 0...N-1 the class
// has number k and modulus N.
func TestNewAllNumbers(t *testing.T) {
	t.Parallel()

	for _, n := range []int{1, 2, 12, 19, 24, 31} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			t.Parallel()

			for k := range n {
				c, err := pitchclass.New(k, n)
				require.NoError(t, err, "New(%d, %d)", k, n)
				assertClass(t, k, n, c, fmt.Sprintf("New(%d, %d)", k, n))
			}
		})
	}
}

// K-2, C-2, G-4, G-6: creation by the circle gives x mod N for any int x;
// N < 1 is ErrInvalidClass.
func TestReduce(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		x, n   int
		number int
	}{
		{name: "13 of 12", x: 13, n: 12, number: 1},
		{name: "-1 of 12", x: -1, n: 12, number: 11},
		{name: "-13 of 12", x: -13, n: 12, number: 11},
		{name: "24 of 12", x: 24, n: 12, number: 0},
		{name: "5 of 1", x: 5, n: 1, number: 0},
		{name: "MIN of 12", x: minInt, n: 12, number: 4},
		{name: "MAX of 12", x: maxInt, n: 12, number: 7},
		{name: "-1 of MAX", x: -1, n: maxInt, number: maxInt - 1},
		{name: "MIN of MAX", x: minInt, n: maxInt, number: maxInt - 1},
		{name: "MAX of MAX", x: maxInt, n: maxInt, number: 0},
		{name: "MIN of 1", x: minInt, n: 1, number: 0},
		{name: "MIN of 2", x: minInt, n: 2, number: 0},
		{name: "MAX of 2", x: maxInt, n: 2, number: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, mod(tc.x, tc.n), tc.number, "test table disagrees with the oracle")

			c, err := pitchclass.Reduce(tc.x, tc.n)
			require.NoError(t, err)
			assertClass(t, tc.number, tc.n, c, "Reduce")
			assertClass(t, tc.number, tc.n, pitchclass.MustReduce(tc.x, tc.n), "MustReduce")
		})
	}

	for _, n := range []int{0, -1, minInt} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			t.Parallel()

			for _, x := range []int{0, 1, -1, minInt, maxInt} {
				c, err := pitchclass.Reduce(x, n)
				assertOnlyError(t, err, pitchclass.ErrInvalidClass, fmt.Sprintf("Reduce(%d, %d)", x, n))
				assert.Equal(t, pitchclass.Class{}, c)
				assertPanicsWithErrorIs(t, pitchclass.ErrInvalidClass, func() { pitchclass.MustReduce(x, n) }, "MustReduce")
			}
		})
	}
}

// K-2: for all x in -50...50 and N in 1...31, Reduce(x, N) is the class of
// strict creation with the number x mod N.
func TestReduceMatchesNew(t *testing.T) {
	t.Parallel()

	for n := 1; n <= 31; n++ {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			t.Parallel()

			for x := -50; x <= 50; x++ {
				desc := fmt.Sprintf("Reduce(%d, %d)", x, n)
				c, err := pitchclass.Reduce(x, n)
				require.NoError(t, err, desc)
				assertClass(t, mod(x, n), n, c, desc)
				assert.True(t, c.Equal(newClass(t, mod(x, n), n)), "%s equals New", desc)
			}
		})
	}
}

// K-3, C-4, G-8: the same class obtained in different ways is equal; classes
// of different N or different numbers are not.
func TestClassEqual(t *testing.T) {
	t.Parallel()

	strict := newClass(t, 1, 12)
	circle := pitchclass.MustReduce(13, 12)
	shifted := newClass(t, 11, 12).Transpose(2)

	equal := []struct {
		name string
		a, b pitchclass.Class
	}{
		{name: "New and Reduce", a: strict, b: circle},
		{name: "New and Transpose", a: strict, b: shifted},
		{name: "Reduce and Transpose", a: circle, b: shifted},
	}

	for _, tc := range equal {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, tc.a.Equal(tc.b))
			assert.True(t, tc.b.Equal(tc.a))
			assert.True(t, tc.a == tc.b, "== agrees with Equal on valid classes")
		})
	}

	notEqual := []struct {
		name string
		a, b pitchclass.Class
	}{
		{name: "1 of 12 and 1 of 19", a: strict, b: newClass(t, 1, 19)},
		{name: "1 of 12 and 2 of 12", a: strict, b: newClass(t, 2, 12)},
		{name: "0 of 1 and 0 of 2", a: newClass(t, 0, 1), b: newClass(t, 0, 2)},
	}

	for _, tc := range notEqual {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.False(t, tc.a.Equal(tc.b))
			assert.False(t, tc.b.Equal(tc.a))
			assert.False(t, tc.a == tc.b)
		})
	}
}

// K-3: equality is reflexive, symmetric and transitive on all classes of
// N in {1, 2, 12, 19}, and holds exactly for the same number and N.
func TestClassEqualRelation(t *testing.T) {
	t.Parallel()

	all := classesOfAll(t, 1, 2, 12, 19)
	for _, a := range all {
		assert.True(t, a.Equal(a), "%v reflexive", a)
		for _, b := range all {
			want := a.Number() == b.Number() && a.Modulus() == b.Modulus()
			assert.Equal(t, want, a.Equal(b), "%v.Equal(%v)", a, b)
			assert.Equal(t, a.Equal(b), b.Equal(a), "symmetry of %v and %v", a, b)
			assert.Equal(t, a.Equal(b), a == b, "== and Equal for %v and %v", a, b)
		}
	}

	for _, a := range all {
		for _, b := range all {
			if !a.Equal(b) {
				continue
			}

			for _, c := range all {
				if b.Equal(c) {
					assert.True(t, a.Equal(c), "transitivity %v, %v, %v", a, b, c)
				}
			}
		}
	}
}

// K-4, C-5: Compare is a total order of labels: by N, then by number; 0
// exactly for equal classes; antisymmetric and transitive; classes of
// different N are compared without an error and are never "equal".
func TestClassCompare(t *testing.T) {
	t.Parallel()

	t.Run("ascending at 12", func(t *testing.T) {
		t.Parallel()

		cs := classesOf(t, 12)
		for i := range cs {
			for j := range cs {
				assert.Equal(t, sign(i-j), cs[i].Compare(cs[j]), "%v.Compare(%v)", cs[i], cs[j])
			}
		}
	})

	t.Run("order of labels", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			a, b pitchclass.Class
			want int
		}{
			{name: "5 of 12 before 0 of 19", a: newClass(t, 5, 12), b: newClass(t, 0, 19), want: -1},
			{name: "0 of 19 after 11 of 12", a: newClass(t, 0, 19), b: newClass(t, 11, 12), want: 1},
			{name: "0 of 1 before 0 of 2", a: newClass(t, 0, 1), b: newClass(t, 0, 2), want: -1},
			{name: "1 of 12 before 1 of 19", a: newClass(t, 1, 12), b: newClass(t, 1, 19), want: -1},
			{name: "MAX-1 of MAX after 11 of 12", a: newClass(t, maxInt-1, maxInt), b: newClass(t, 11, 12), want: 1},
			{name: "0 of MAX before MAX-1 of MAX", a: newClass(t, 0, maxInt), b: newClass(t, maxInt-1, maxInt), want: -1},
			{name: "equal", a: newClass(t, 7, 12), b: pitchclass.MustReduce(-5, 12), want: 0},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.want, tc.a.Compare(tc.b))
				assert.Equal(t, -tc.want, tc.b.Compare(tc.a))
			})
		}
	})

	t.Run("relation", func(t *testing.T) {
		t.Parallel()

		all := classesOfAll(t, 1, 2, 12, 19)
		for _, a := range all {
			for _, b := range all {
				got := a.Compare(b)
				want := compareClassSpec(a.Number(), a.Modulus(), b.Number(), b.Modulus())
				assert.Equal(t, want, got, "%v.Compare(%v)", a, b)
				assert.Equal(t, a.Equal(b), got == 0, "Compare == 0 iff Equal for %v, %v", a, b)
				assert.Equal(t, -got, b.Compare(a), "antisymmetry %v, %v", a, b)
				if a.Modulus() != b.Modulus() {
					assert.NotZero(t, got, "different N are never equal: %v, %v", a, b)
				}
			}
		}

		for _, a := range all {
			for _, b := range all {
				if a.Compare(b) > 0 {
					continue
				}

				for _, c := range all {
					if b.Compare(c) <= 0 {
						assert.LessOrEqual(t, a.Compare(c), 0, "transitivity %v <= %v <= %v", a, b, c)
					}
				}
			}
		}
	})

	t.Run("sorting permutations", func(t *testing.T) {
		t.Parallel()

		want := classesOfAll(t, 2, 12, 19)
		for _, perm := range permutations(len(want)) {
			xs := make([]pitchclass.Class, len(want))
			for i, p := range perm {
				xs[i] = want[p]
			}

			slices.SortFunc(xs, pitchclass.Class.Compare)
			assert.Equal(t, want, xs, "sorted permutation %v", perm)
		}
	})
}

// permutations returns deterministic permutations of 0...n-1: the reverse,
// rotations and strided orders with strides coprime to n.
func permutations(n int) [][]int {
	var out [][]int

	reverse := make([]int, n)
	for i := range n {
		reverse[i] = n - 1 - i
	}
	out = append(out, reverse)

	for _, shift := range []int{1, 7} {
		p := make([]int, n)
		for i := range n {
			p[i] = (i + shift) % n
		}
		out = append(out, p)
	}

	for stride := 2; stride < n && len(out) < 8; stride++ {
		if gcdInt(stride, n) != 1 {
			continue
		}

		p := make([]int, n)
		for i := range n {
			p[i] = (i * stride) % n
		}
		out = append(out, p)
	}

	return out
}

// K-5, C-6, G-7, E-2: All yields the classes of N in ascending order; at
// N = MAX the first elements come without iterating the rest; N < 1 is
// ErrInvalidModulus with a non-nil empty iterator.
func TestAll(t *testing.T) {
	t.Parallel()

	for _, n := range []int{1, 2, 12, 19} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			t.Parallel()

			want := classesOf(t, n)

			seq, err := pitchclass.All(n)
			require.NoError(t, err)
			assertSeq(t, seq, want, "All")
			assertSeq(t, pitchclass.MustAll(n), want, "MustAll")
		})
	}

	t.Run("N=12 numbers", func(t *testing.T) {
		t.Parallel()

		got := slices.Collect(pitchclass.MustAll(12))
		assert.Equal(t, ints(0, 11), numbersOf(got))
		for _, c := range got {
			assert.Equal(t, 12, c.Modulus())
		}
	})

	t.Run("N=MAX first three", func(t *testing.T) {
		t.Parallel()

		seq, err := pitchclass.All(maxInt)
		require.NoError(t, err)

		got := take(seq, 3)
		require.Len(t, got, 3)
		for i, c := range got {
			assertClass(t, i, maxInt, c, fmt.Sprintf("element %d", i))
		}
		assert.Zero(t, callsAfterStop(seq, 3))
	})

	for _, n := range []int{0, -1, minInt} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			t.Parallel()

			seq, err := pitchclass.All(n)
			assertOnlyError(t, err, pitchclass.ErrInvalidModulus, "All")
			assertEmptySeq(t, seq, "All")
			assertPanicsWithErrorIs(t, pitchclass.ErrInvalidModulus, func() { pitchclass.MustAll(n) }, "MustAll")
		})
	}
}

// K-6, C-7: Cycle yields N classes upward from the class, element i = Ti(x),
// each class exactly once; at N = MAX the first elements come without
// iterating the rest.
func TestClassCycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		from    pitchclass.Class
		numbers []int
	}{
		{name: "from 10 of 12", from: newClass(t, 10, 12), numbers: []int{10, 11, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{name: "from 0 of 12", from: newClass(t, 0, 12), numbers: ints(0, 11)},
		{name: "from 0 of 1", from: newClass(t, 0, 1), numbers: []int{0}},
		{name: "from 1 of 2", from: newClass(t, 1, 2), numbers: []int{1, 0}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			n := tc.from.Modulus()
			want := make([]pitchclass.Class, 0, len(tc.numbers))
			for _, k := range tc.numbers {
				want = append(want, newClass(t, k, n))
			}

			assertSeq(t, tc.from.Cycle(), want, "Cycle")
		})
	}

	t.Run("from 0 of 12 as All", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, slices.Collect(pitchclass.MustAll(12)), slices.Collect(newClass(t, 0, 12).Cycle()))
	})

	for _, n := range []int{2, 12, 19} {
		t.Run(fmt.Sprintf("all classes of %d", n), func(t *testing.T) {
			t.Parallel()

			for _, x := range classesOf(t, n) {
				got := take(x.Cycle(), n+1)
				require.Len(t, got, n, "Cycle of %v", x)

				seen := map[pitchclass.Class]bool{}
				for i, c := range got {
					assertClass(t, addMod(x.Number(), i, n), n, c, fmt.Sprintf("Cycle of %v, element %d", x, i))
					assert.False(t, seen[c], "Cycle of %v repeats %v", x, c)
					seen[c] = true
				}
				assert.Len(t, seen, n)
			}
		})
	}

	t.Run("from MAX-1 of MAX first three", func(t *testing.T) {
		t.Parallel()

		seq := newClass(t, maxInt-1, maxInt).Cycle()
		got := take(seq, 3)
		require.Len(t, got, 3)
		assertClass(t, maxInt-1, maxInt, got[0], "element 0")
		assertClass(t, 0, maxInt, got[1], "element 1")
		assertClass(t, 1, maxInt, got[2], "element 2")
		assert.Zero(t, callsAfterStop(seq, 3))
	})
}

// K-31 (classes), NF-2: equal classes are one map key, different classes
// are different keys.
func TestClassMapKey(t *testing.T) {
	t.Parallel()

	m := map[pitchclass.Class]int{}
	m[newClass(t, 1, 12)]++
	m[pitchclass.MustReduce(13, 12)]++
	m[newClass(t, 0, 12).Transpose(1)]++
	m[newClass(t, 1, 19)]++

	assert.Len(t, m, 2)
	assert.Equal(t, 3, m[newClass(t, 1, 12)])
	assert.Equal(t, 1, m[newClass(t, 1, 19)])
}
