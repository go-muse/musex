package pitchclass_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-7, C-8, G-4, G-6: Tn(k) = (k + n) mod N for any int n, without overflow.
func TestClassTranspose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		k, n, size int
		want       int
	}{
		{name: "T2(11) at 12", k: 11, n: 2, size: 12, want: 1},
		{name: "T-1(0) at 12", k: 0, n: -1, size: 12, want: 11},
		{name: "T0(5) at 12", k: 5, n: 0, size: 12, want: 5},
		{name: "T12(5) at 12", k: 5, n: 12, size: 12, want: 5},
		{name: "TMAX(11) at 12", k: 11, n: maxInt, size: 12, want: 6},
		{name: "TMIN(0) at 12", k: 0, n: minInt, size: 12, want: 4},
		{name: "T(MAX-1)(MAX-1) at MAX", k: maxInt - 1, n: maxInt - 1, size: maxInt, want: maxInt - 2},
		{name: "T1(MAX-1) at MAX", k: maxInt - 1, n: 1, size: maxInt, want: 0},
		{name: "TMIN(0) at MAX", k: 0, n: minInt, size: maxInt, want: maxInt - 1},
		{name: "TMAX(MAX-1) at MAX", k: maxInt - 1, n: maxInt, size: maxInt, want: maxInt - 1},
		{name: "TMIN(MAX-1) at MAX", k: maxInt - 1, n: minInt, size: maxInt, want: maxInt - 2},
		{name: "TMAX(0) at 1", k: 0, n: maxInt, size: 1, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, addMod(tc.k, tc.n, tc.size), tc.want, "test table disagrees with the oracle")
			assertClass(t, tc.want, tc.size, newClass(t, tc.k, tc.size).Transpose(tc.n), "Transpose")
		})
	}
}

// K-7: for all k, n in -30...30 and N in {1, 2, 12, 19} Tn agrees with the
// oracle and T-n(Tn(x)) = x; edge shifts at N = 12 and MAX as well.
func TestClassTransposeOracle(t *testing.T) {
	t.Parallel()

	for _, size := range []int{1, 2, 12, 19} {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			for _, x := range classesOf(t, size) {
				for _, n := range append(ints(-30, 30), edgeInts()...) {
					desc := fmt.Sprintf("%v.Transpose(%d)", x, n)
					got := x.Transpose(n)
					assertClass(t, addMod(x.Number(), n, size), size, got, desc)
					if n != minInt {
						assert.True(t, got.Transpose(-n).Equal(x), "%s, then Transpose(%d)", desc, -n)
					}
				}
			}
		})
	}

	t.Run("N=MAX", func(t *testing.T) {
		t.Parallel()

		for _, k := range []int{0, 1, maxInt / 2, maxInt - 2, maxInt - 1} {
			x := newClass(t, k, maxInt)
			for _, n := range edgeInts() {
				assertClass(t, addMod(k, n, maxInt), maxInt, x.Transpose(n), fmt.Sprintf("%v.Transpose(%d)", x, n))
			}
		}
	})
}

// K-8, C-9, C-10, G-5: the ordered interval i<x, y> = (y - x) mod N; the
// complementary interval of the pair is y.IntervalTo(x).
func TestClassIntervalTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		x, y, size int
		want       int
	}{
		{name: "<10, 2> at 12", x: 10, y: 2, size: 12, want: 4},
		{name: "<2, 10> at 12", x: 2, y: 10, size: 12, want: 8},
		{name: "<0, 4> at 12", x: 0, y: 4, size: 12, want: 4},
		{name: "<4, 0> at 12", x: 4, y: 0, size: 12, want: 8},
		{name: "<0, 6> at 12", x: 0, y: 6, size: 12, want: 6},
		{name: "<6, 0> at 12", x: 6, y: 0, size: 12, want: 6},
		{name: "<5, 5> at 12", x: 5, y: 5, size: 12, want: 0},
		{name: "<0, 0> at 1", x: 0, y: 0, size: 1, want: 0},
		{name: "<MAX-1, 0> at MAX", x: maxInt - 1, y: 0, size: maxInt, want: 1},
		{name: "<0, MAX-1> at MAX", x: 0, y: maxInt - 1, size: maxInt, want: maxInt - 1},
		{name: "<MAX-1, MAX-1> at MAX", x: maxInt - 1, y: maxInt - 1, size: maxInt, want: 0},
		{name: "<1, MAX-2> at MAX", x: 1, y: maxInt - 2, size: maxInt, want: maxInt - 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, subMod(tc.y, tc.x, tc.size), tc.want, "test table disagrees with the oracle")

			x, y := newClass(t, tc.x, tc.size), newClass(t, tc.y, tc.size)
			got, err := x.IntervalTo(y)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.want, x.MustIntervalTo(y))
		})
	}

	complements := []struct {
		name string
		x, y int
		want int
	}{
		{name: "(0, 4)", x: 0, y: 4, want: 8},
		{name: "(0, 6)", x: 0, y: 6, want: 6},
		{name: "(3, 3)", x: 3, y: 3, want: 0},
		{name: "(10, 2)", x: 10, y: 2, want: 8},
	}

	for _, tc := range complements {
		t.Run("complement "+tc.name, func(t *testing.T) {
			t.Parallel()

			x, y := newClass(t, tc.x, 12), newClass(t, tc.y, 12)
			got, err := y.IntervalTo(x)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// K-8: for all x, y at N in {1, 2, 12, 19}: i<x, y> in 0...N-1,
// T_i<x, y>(x) = y, and the complement i<y, x> = (N - i<x, y>) mod N.
func TestClassIntervalToRelation(t *testing.T) {
	t.Parallel()

	for _, size := range []int{1, 2, 12, 19} {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			cs := classesOf(t, size)
			for _, x := range cs {
				for _, y := range cs {
					desc := fmt.Sprintf("%v.IntervalTo(%v)", x, y)
					i, err := x.IntervalTo(y)
					require.NoError(t, err, desc)
					assert.Equal(t, subMod(y.Number(), x.Number(), size), i, desc)
					assert.True(t, i >= 0 && i < size, "%s = %d in 0...N-1", desc, i)
					assert.True(t, x.Transpose(i).Equal(y), "%s: Transpose(%d) gives %v", desc, i, y)

					back, err := y.IntervalTo(x)
					require.NoError(t, err)
					assert.Equal(t, mod(size-i, size), back, "complement of %s", desc)
				}
			}
		})
	}
}

// K-9, C-11, G-1...G-3: the interval class of a pair.
func TestClassIntervalClass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		x, y, size int
		want       int
	}{
		{name: "ic(10, 2) at 12", x: 10, y: 2, size: 12, want: 4},
		{name: "ic(2, 10) at 12", x: 2, y: 10, size: 12, want: 4},
		{name: "ic(0, 6) at 12", x: 0, y: 6, size: 12, want: 6},
		{name: "ic(0, 7) at 12", x: 0, y: 7, size: 12, want: 5},
		{name: "ic(4, 4) at 12", x: 4, y: 4, size: 12, want: 0},
		{name: "ic(0, 10) at 19", x: 0, y: 10, size: 19, want: 9},
		{name: "ic(0, 1) at 2", x: 0, y: 1, size: 2, want: 1},
		{name: "ic(0, 0) at 1", x: 0, y: 0, size: 1, want: 0},
		{name: "ic(0, MAX-1) at MAX", x: 0, y: maxInt - 1, size: maxInt, want: 1},
		{name: "ic(0, MAX/2) at MAX", x: 0, y: maxInt / 2, size: maxInt, want: maxInt / 2},
		{name: "ic(0, MAX/2+1) at MAX", x: 0, y: maxInt/2 + 1, size: maxInt, want: maxInt / 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, intervalClassOf(tc.y-tc.x, tc.size), tc.want, "test table disagrees with the oracle")

			x, y := newClass(t, tc.x, tc.size), newClass(t, tc.y, tc.size)
			got, err := x.IntervalClass(y)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.want, x.MustIntervalClass(y))
		})
	}
}

// K-9: for all x, y at N in 1...24: ic in 0...floor(N/2), symmetric, 0 only
// for equal classes; the value floor(N/2) is reached.
func TestClassIntervalClassRelation(t *testing.T) {
	t.Parallel()

	for size := 1; size <= 24; size++ {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			cs := classesOf(t, size)
			reached := false
			for _, x := range cs {
				for _, y := range cs {
					desc := fmt.Sprintf("%v.IntervalClass(%v)", x, y)
					ic, err := x.IntervalClass(y)
					require.NoError(t, err, desc)
					assert.Equal(t, intervalClassOf(y.Number()-x.Number(), size), ic, desc)
					assert.True(t, ic >= 0 && ic <= size/2, "%s = %d in 0...N/2", desc, ic)

					back, err := y.IntervalClass(x)
					require.NoError(t, err)
					assert.Equal(t, ic, back, "symmetry of %s", desc)
					assert.Equal(t, x.Equal(y), ic == 0, "%s is 0 iff the classes are equal", desc)
					reached = reached || ic == size/2
				}
			}
			assert.True(t, reached, "floor(N/2) = %d is reached", size/2)
		})
	}
}

// K-10, C-12, G-4, E-2: intervals given by size: i mod N, (-i) mod N and
// ic for any int i; N < 1 is ErrInvalidModulus.
func TestIntervalBySize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                         string
		i, size                      int
		ordered, complementary, icls int
	}{
		{name: "14 at 12", i: 14, size: 12, ordered: 2, complementary: 10, icls: 2},
		{name: "-1 at 12", i: -1, size: 12, ordered: 11, complementary: 1, icls: 1},
		{name: "0 at 12", i: 0, size: 12, ordered: 0, complementary: 0, icls: 0},
		{name: "4 at 12", i: 4, size: 12, ordered: 4, complementary: 8, icls: 4},
		{name: "6 at 12", i: 6, size: 12, ordered: 6, complementary: 6, icls: 6},
		{name: "7 at 12", i: 7, size: 12, ordered: 7, complementary: 5, icls: 5},
		{name: "19 at 12", i: 19, size: 12, ordered: 7, complementary: 5, icls: 5},
		{name: "-3 at 12", i: -3, size: 12, ordered: 9, complementary: 3, icls: 3},
		{name: "MIN at 12", i: minInt, size: 12, ordered: 4, complementary: 8, icls: 4},
		{name: "MAX at 12", i: maxInt, size: 12, ordered: 7, complementary: 5, icls: 5},
		{name: "MIN at 1", i: minInt, size: 1},
		{name: "-1 at 1", i: -1, size: 1},
		{name: "5 at 1", i: 5, size: 1},
		{name: "MAX at 1", i: maxInt, size: 1},
		{name: "MIN at MAX", i: minInt, size: maxInt, ordered: maxInt - 1, complementary: 1, icls: 1},
		{name: "MAX at MAX", i: maxInt, size: maxInt},
		{name: "-1 at MAX", i: -1, size: maxInt, ordered: maxInt - 1, complementary: 1, icls: 1},
		{name: "MAX/2 at MAX", i: maxInt / 2, size: maxInt, ordered: maxInt / 2, complementary: maxInt/2 + 1, icls: maxInt / 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, mod(tc.i, tc.size), tc.ordered, "test table disagrees with the oracle")
			require.Equal(t, negMod(tc.i, tc.size), tc.complementary, "test table disagrees with the oracle")
			require.Equal(t, intervalClassOf(tc.i, tc.size), tc.icls, "test table disagrees with the oracle")

			assertIntervalBySize(t, tc.i, tc.size, tc.ordered, tc.complementary, tc.icls)
		})
	}

	for _, size := range []int{0, -1, minInt} {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			for _, i := range []int{0, 1, minInt, maxInt} {
				desc := fmt.Sprintf("(%d, %d)", i, size)

				got, err := pitchclass.ReduceInterval(i, size)
				assertOnlyError(t, err, pitchclass.ErrInvalidModulus, "ReduceInterval"+desc)
				assert.Zero(t, got)

				got, err = pitchclass.ComplementInterval(i, size)
				assertOnlyError(t, err, pitchclass.ErrInvalidModulus, "ComplementInterval"+desc)
				assert.Zero(t, got)

				got, err = pitchclass.IntervalClass(i, size)
				assertOnlyError(t, err, pitchclass.ErrInvalidModulus, "IntervalClass"+desc)
				assert.Zero(t, got)

				assertPanicsWithErrorIs(t, pitchclass.ErrInvalidModulus, func() { pitchclass.MustReduceInterval(i, size) }, "MustReduceInterval"+desc)
				assertPanicsWithErrorIs(t, pitchclass.ErrInvalidModulus, func() { pitchclass.MustComplementInterval(i, size) }, "MustComplementInterval"+desc)
				assertPanicsWithErrorIs(t, pitchclass.ErrInvalidModulus, func() { pitchclass.MustIntervalClass(i, size) }, "MustIntervalClass"+desc)
			}
		})
	}
}

// assertIntervalBySize checks the three interval functions and their Must
// variants for interval i at modulus size.
func assertIntervalBySize(t *testing.T, i, size, ordered, complementary, icls int) {
	t.Helper()

	desc := fmt.Sprintf("(%d, %d)", i, size)

	got, err := pitchclass.ReduceInterval(i, size)
	require.NoError(t, err, "ReduceInterval"+desc)
	assert.Equal(t, ordered, got, "ReduceInterval"+desc)
	assert.Equal(t, ordered, pitchclass.MustReduceInterval(i, size), "MustReduceInterval"+desc)

	got, err = pitchclass.ComplementInterval(i, size)
	require.NoError(t, err, "ComplementInterval"+desc)
	assert.Equal(t, complementary, got, "ComplementInterval"+desc)
	assert.Equal(t, complementary, pitchclass.MustComplementInterval(i, size), "MustComplementInterval"+desc)

	got, err = pitchclass.IntervalClass(i, size)
	require.NoError(t, err, "IntervalClass"+desc)
	assert.Equal(t, icls, got, "IntervalClass"+desc)
	assert.Equal(t, icls, pitchclass.MustIntervalClass(i, size), "MustIntervalClass"+desc)
}

// K-10: for all i in -30...30 (and edge values) at N in {1, 2, 12, 19} the
// functions agree with the oracle, and ic by size equals ic of the pair
// (0, i mod N).
func TestIntervalBySizeOracle(t *testing.T) {
	t.Parallel()

	for _, size := range []int{1, 2, 12, 19, maxInt} {
		t.Run(fmt.Sprintf("N=%d", size), func(t *testing.T) {
			t.Parallel()

			zero := newClass(t, 0, size)
			for _, i := range append(ints(-30, 30), edgeInts()...) {
				assertIntervalBySize(t, i, size, mod(i, size), negMod(i, size), intervalClassOf(i, size))

				pair, err := zero.IntervalClass(pitchclass.MustReduce(i, size))
				require.NoError(t, err)
				assert.Equal(t, pair, pitchclass.MustIntervalClass(i, size), "ic by size %d equals ic of the pair", i)
			}
		})
	}
}

// K-11, C-13, G-5: the index sum (x + y) mod N is the index of the inversion
// that swaps x and y.
func TestClassIndexSum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		x, y, size int
		want       int
	}{
		{name: "4 + 9 at 12", x: 4, y: 9, size: 12, want: 1},
		{name: "0 + 0 at 12", x: 0, y: 0, size: 12, want: 0},
		{name: "6 + 6 at 12", x: 6, y: 6, size: 12, want: 0},
		{name: "3 + 3 at 12", x: 3, y: 3, size: 12, want: 6},
		{name: "(MAX-1) + (MAX-1) at MAX", x: maxInt - 1, y: maxInt - 1, size: maxInt, want: maxInt - 2},
		{name: "(MAX-1) + 1 at MAX", x: maxInt - 1, y: 1, size: maxInt, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, addMod(tc.x, tc.y, tc.size), tc.want, "test table disagrees with the oracle")

			x, y := newClass(t, tc.x, tc.size), newClass(t, tc.y, tc.size)
			got, err := x.IndexSum(y)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.want, x.MustIndexSum(y))
		})
	}

	for _, size := range []int{1, 2, 12, 19} {
		t.Run(fmt.Sprintf("relation N=%d", size), func(t *testing.T) {
			t.Parallel()

			cs := classesOf(t, size)
			for _, x := range cs {
				for _, y := range cs {
					desc := fmt.Sprintf("%v.IndexSum(%v)", x, y)
					s, err := x.IndexSum(y)
					require.NoError(t, err, desc)
					assert.Equal(t, addMod(x.Number(), y.Number(), size), s, desc)
					assert.True(t, x.Invert(s).Equal(y), "%s: I%d(x) = y", desc, s)
					assert.True(t, y.Invert(s).Equal(x), "%s: I%d(y) = x", desc, s)
				}
			}
		})
	}
}
