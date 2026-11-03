package octave_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/octave"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-10, C-10, G-6: shift by k, including k = 0 at the limits and k = MIN.
func TestShift(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		o    octave.Octave
		k    int
		want octave.Octave
	}{
		{name: "O(4) by 1", o: oct(4), k: 1, want: oct(5)},
		{name: "O(4) by -5", o: oct(4), k: -5, want: oct(-1)},
		{name: "O(4) by 0", o: oct(4), k: 0, want: oct(4)},
		{name: "O(MIN) by 0", o: oct(minInt), k: 0, want: oct(minInt)},
		{name: "O(MAX) by 0", o: oct(maxInt), k: 0, want: oct(maxInt)},
		{name: "O(0) by MIN", o: oct(0), k: minInt, want: oct(minInt)},
		{name: "O(MAX) by MIN", o: oct(maxInt), k: minInt, want: oct(-1)},
		{name: "O(MIN) by MAX", o: oct(minInt), k: maxInt, want: oct(-1)},
		{name: "O(0) by MAX", o: oct(0), k: maxInt, want: oct(maxInt)},
		{name: "O(-1000) by 1", o: oct(-1000), k: 1, want: oct(-999)},
		{name: "O(1000) by -1", o: oct(1000), k: -1, want: oct(999)},
		{name: "O(3, FB) by 1", o: octave.MustFromNumber(3, fb()), k: 1, want: oct(5)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got octave.Octave
				err error
			)
			require.NotPanics(t, func() { got, err = tc.o.Shift(tc.k) })
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.True(t, got.Equal(tc.want), "Equal")
			assert.Equal(t, tc.want, tc.o.MustShift(tc.k), "MustShift")
		})
	}

	t.Run("FB 1 by -1 has FB number -1", func(t *testing.T) {
		t.Parallel()

		got := octave.MustFromNumber(1, fb()).MustShift(-1)
		assert.Equal(t, -1, got.MustNumber(fb()))
	})
}

// K-10: for all a, k1, k2 in -10...10, shifting by k1 and then by k2 is
// shifting by k1 + k2.
func TestShiftComposition(t *testing.T) {
	t.Parallel()

	for _, a := range span(-10, 10) {
		for _, k1 := range span(-10, 10) {
			first, err := oct(a).Shift(k1)
			require.NoError(t, err, "a=%d k1=%d", a, k1)
			for _, k2 := range span(-10, 10) {
				twice, err := first.Shift(k2)
				require.NoError(t, err, "a=%d k1=%d k2=%d", a, k1, k2)
				once, err := oct(a).Shift(k1 + k2)
				require.NoError(t, err, "a=%d k1+k2=%d", a, k1+k2)
				assert.Equal(t, once, twice, "a=%d k1=%d k2=%d", a, k1, k2)
				assert.Equal(t, oct(a+k1+k2), twice, "a=%d k1=%d k2=%d", a, k1, k2)
			}
		}
	}
}

// K-11, C-11, G-6: shift down by k, including k = MIN.
func TestShiftDown(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		o    octave.Octave
		k    int
		want octave.Octave
	}{
		{name: "O(4) by 1", o: oct(4), k: 1, want: oct(3)},
		{name: "O(4) by -2", o: oct(4), k: -2, want: oct(6)},
		{name: "O(4) by 0", o: oct(4), k: 0, want: oct(4)},
		{name: "O(MIN) by 0", o: oct(minInt), k: 0, want: oct(minInt)},
		{name: "O(MAX) by 0", o: oct(maxInt), k: 0, want: oct(maxInt)},
		{name: "O(-1) by MIN", o: oct(-1), k: minInt, want: oct(maxInt)},
		{name: "O(MIN) by MIN", o: oct(minInt), k: minInt, want: oct(0)},
		{name: "O(MAX) by MAX", o: oct(maxInt), k: maxInt, want: oct(0)},
		{name: "O(5) by 1", o: oct(5), k: 1, want: oct(4)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got octave.Octave
				err error
			)
			require.NotPanics(t, func() { got, err = tc.o.ShiftDown(tc.k) })
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.want, tc.o.MustShiftDown(tc.k), "MustShiftDown")
		})
	}

	t.Run("O(0) shifted by MIN and back", func(t *testing.T) {
		t.Parallel()

		low := oct(0).MustShift(minInt)
		require.Equal(t, oct(minInt), low)
		back, err := low.ShiftDown(minInt)
		require.NoError(t, err)
		assert.Equal(t, oct(0), back)
	})

	t.Run("O(0) by MIN is out of range", func(t *testing.T) {
		t.Parallel()

		got, err := oct(0).ShiftDown(minInt)
		assertCause(t, err, causeOutOfRange, "O(0).ShiftDown(MIN)")
		assert.Equal(t, octave.Octave{}, got, "value on error")
		assertPanicsCause(t, causeOutOfRange, func() { oct(0).MustShiftDown(minInt) }, "MustShiftDown")
	})
}

// K-11, C-11: for all a, k in -10...10, shift down is reversible in both
// directions and equals Shift(-k).
func TestShiftDownReversible(t *testing.T) {
	t.Parallel()

	for _, a := range span(-10, 10) {
		for _, k := range span(-10, 10) {
			up, err := oct(a).Shift(k)
			require.NoError(t, err, "a=%d k=%d", a, k)
			back, err := up.ShiftDown(k)
			require.NoError(t, err, "a=%d k=%d", a, k)
			assert.Equal(t, oct(a), back, "ShiftDown(Shift(O(%d), %d), %d)", a, k, k)

			down, err := oct(a).ShiftDown(k)
			require.NoError(t, err, "a=%d k=%d", a, k)
			again, err := down.Shift(k)
			require.NoError(t, err, "a=%d k=%d", a, k)
			assert.Equal(t, oct(a), again, "Shift(ShiftDown(O(%d), %d), %d)", a, k, k)

			neg, err := oct(a).Shift(-k)
			require.NoError(t, err, "a=%d k=%d", a, k)
			assert.Equal(t, neg, down, "ShiftDown(O(%d), %d) = Shift(O(%d), %d)", a, k, a, -k)
		}
	}
}

// K-10, K-11, K-23, G-2, G-6: Shift and ShiftDown on a grid of octaves and
// shift amounts at the limits of int, against exact integer arithmetic: the
// result is the octave with ASPN number a + k (a - k) when it fits in an int,
// otherwise ErrOutOfRange without number.ErrOutOfRange and the invalid value.
// Where a shift succeeds, the opposite shift by the same k returns the
// original octave, for every k including MIN, MIN + 1 and MAX.
func TestShiftLimits(t *testing.T) {
	t.Parallel()

	for _, a := range edgeASPN() {
		for _, k := range edgeShifts() {
			var (
				up, down       octave.Octave
				errUp, errDown error
			)
			require.NotPanics(t, func() { up, errUp = oct(a).Shift(k) }, "O(%d).Shift(%d)", a, k)
			require.NotPanics(t, func() { down, errDown = oct(a).ShiftDown(k) }, "O(%d).ShiftDown(%d)", a, k)

			if s, ok := exactAdd(a, k); ok {
				require.NoError(t, errUp, "O(%d).Shift(%d)", a, k)
				assert.Equal(t, oct(s), up, "O(%d).Shift(%d)", a, k)
				back, err := up.ShiftDown(k)
				require.NoError(t, err, "O(%d).MustShift(%d).ShiftDown(%d)", a, k, k)
				assert.Equal(t, oct(a), back, "O(%d).MustShift(%d).ShiftDown(%d)", a, k, k)
			} else {
				assertCause(t, errUp, causeOutOfRange, fmt.Sprintf("O(%d).Shift(%d)", a, k))
				assert.Equal(t, octave.Octave{}, up, "O(%d).Shift(%d) value on error", a, k)
			}

			if d, ok := exactSub(a, k); ok {
				require.NoError(t, errDown, "O(%d).ShiftDown(%d)", a, k)
				assert.Equal(t, oct(d), down, "O(%d).ShiftDown(%d)", a, k)
				back, err := down.Shift(k)
				require.NoError(t, err, "O(%d).MustShiftDown(%d).Shift(%d)", a, k, k)
				assert.Equal(t, oct(a), back, "O(%d).MustShiftDown(%d).Shift(%d)", a, k, k)
			} else {
				assertCause(t, errDown, causeOutOfRange, fmt.Sprintf("O(%d).ShiftDown(%d)", a, k))
				assert.Equal(t, octave.Octave{}, down, "O(%d).ShiftDown(%d) value on error", a, k)
			}
		}
	}
}

// K-12, C-12: the neighboring octaves.
func TestUpDown(t *testing.T) {
	t.Parallel()

	t.Run("O(4)", func(t *testing.T) {
		t.Parallel()

		up, err := oct(4).Up()
		require.NoError(t, err)
		assert.Equal(t, oct(5), up)
		down, err := oct(4).Down()
		require.NoError(t, err)
		assert.Equal(t, oct(3), down)
		assert.Equal(t, oct(4), oct(4).MustUp().MustDown(), "MustUp().MustDown()")
		assert.Equal(t, oct(4), oct(4).MustDown().MustUp(), "MustDown().MustUp()")
	})

	t.Run("limits", func(t *testing.T) {
		t.Parallel()

		up, err := oct(maxInt - 1).Up()
		require.NoError(t, err)
		assert.Equal(t, oct(maxInt), up)
		down, err := oct(minInt + 1).Down()
		require.NoError(t, err)
		assert.Equal(t, oct(minInt), down)
		assert.Equal(t, oct(maxInt-1), oct(maxInt).MustDown())
		assert.Equal(t, oct(minInt+1), oct(minInt).MustUp())
	})

	t.Run("agrees with Shift", func(t *testing.T) {
		t.Parallel()

		for _, a := range span(-10, 10) {
			up, err := oct(a).Up()
			require.NoError(t, err, "O(%d).Up()", a)
			assert.Equal(t, oct(a).MustShift(1), up, "O(%d).Up()", a)
			down, err := oct(a).Down()
			require.NoError(t, err, "O(%d).Down()", a)
			assert.Equal(t, oct(a).MustShift(-1), down, "O(%d).Down()", a)
			assert.Equal(t, oct(a), up.MustDown(), "O(%d).Up().Down()", a)
		}
	})

	t.Run("neighbors across the skipped Franco-Belgian number", func(t *testing.T) {
		t.Parallel()

		one := octave.MustFromNumber(1, fb())
		minusOne := octave.MustFromNumber(-1, fb())
		assert.Equal(t, minusOne, one.MustDown(), "FB 1 down")
		assert.Equal(t, one, minusOne.MustUp(), "FB -1 up")
	})
}

// K-13, C-13, G-4: difference of octaves.
func TestSub(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		a, b octave.Octave
		want int
	}{
		{name: "O(5) - O(4)", a: oct(5), b: oct(4), want: 1},
		{name: "O(4) - O(5)", a: oct(4), b: oct(5), want: -1},
		{name: "O(4) - O(4)", a: oct(4), b: oct(4), want: 0},
		{name: "O(MIN) - O(MIN)", a: oct(minInt), b: oct(minInt), want: 0},
		{name: "FB 3 - FB -1", a: octave.MustFromNumber(3, fb()), b: octave.MustFromNumber(-1, fb()), want: 3},
		{name: "FB 1 - FB -1", a: octave.MustFromNumber(1, fb()), b: octave.MustFromNumber(-1, fb()), want: 1},
		{name: "O(MIN) - O(0)", a: oct(minInt), b: oct(0), want: minInt},
		{name: "O(MAX) - O(0)", a: oct(maxInt), b: oct(0), want: maxInt},
		{name: "O(-1) - O(MAX)", a: oct(-1), b: oct(maxInt), want: minInt},
		{name: "O(MAX) - O(1)", a: oct(maxInt), b: oct(1), want: maxInt - 1},
		{name: "O(1000) - O(-1000)", a: oct(1000), b: oct(-1000), want: 2000},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got int
				err error
			)
			require.NotPanics(t, func() { got, err = tc.a.Sub(tc.b) })
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.want, tc.a.MustSub(tc.b), "MustSub")
		})
	}
}

// K-13: for all a, b, k in -10...10, shifting O(b) by Sub(O(a), O(b)) gives
// O(a), and Sub(Shift(O(a), k), O(a)) = k; Sub is antisymmetric.
func TestSubIdentities(t *testing.T) {
	t.Parallel()

	for _, a := range span(-10, 10) {
		for _, b := range span(-10, 10) {
			d, err := oct(a).Sub(oct(b))
			require.NoError(t, err, "Sub(O(%d), O(%d))", a, b)
			assert.Equal(t, a-b, d, "Sub(O(%d), O(%d))", a, b)
			assert.Equal(t, oct(a), oct(b).MustShift(d), "Shift(O(%d), Sub(O(%d), O(%d)))", b, a, b)
			assert.Equal(t, -d, oct(b).MustSub(oct(a)), "Sub(O(%d), O(%d))", b, a)
		}
		for _, k := range span(-10, 10) {
			d, err := oct(a).MustShift(k).Sub(oct(a))
			require.NoError(t, err, "Sub(Shift(O(%d), %d), O(%d))", a, k, a)
			assert.Equal(t, k, d, "Sub(Shift(O(%d), %d), O(%d))", a, k, a)
		}
	}
}

// K-14, C-14: distance, the absolute value of the difference.
func TestDistance(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		a, b octave.Octave
		want int
	}{
		{name: "O(2) O(5)", a: oct(2), b: oct(5), want: 3},
		{name: "O(5) O(2)", a: oct(5), b: oct(2), want: 3},
		{name: "O(4) O(4)", a: oct(4), b: oct(4), want: 0},
		{name: "O(MIN+1) O(0)", a: oct(minInt + 1), b: oct(0), want: maxInt},
		{name: "O(0) O(MIN+1)", a: oct(0), b: oct(minInt + 1), want: maxInt},
		{name: "O(MAX) O(0)", a: oct(maxInt), b: oct(0), want: maxInt},
		{name: "O(0) O(MAX)", a: oct(0), b: oct(maxInt), want: maxInt},
		{name: "O(MAX) O(MAX)", a: oct(maxInt), b: oct(maxInt), want: 0},
		{name: "FB 1 FB -1", a: octave.MustFromNumber(1, fb()), b: octave.MustFromNumber(-1, fb()), want: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got int
				err error
			)
			require.NotPanics(t, func() { got, err = tc.a.Distance(tc.b) })
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.want, tc.a.MustDistance(tc.b), "MustDistance")
		})
	}

	t.Run("equals the absolute difference", func(t *testing.T) {
		t.Parallel()

		for _, a := range span(-10, 10) {
			for _, b := range span(-10, 10) {
				d, err := oct(a).Distance(oct(b))
				require.NoError(t, err, "Distance(O(%d), O(%d))", a, b)
				sub := oct(a).MustSub(oct(b))
				assert.Equal(t, max(sub, -sub), d, "Distance(O(%d), O(%d)) = |Sub|", a, b)
				assert.Equal(t, d, oct(b).MustDistance(oct(a)), "symmetric: O(%d), O(%d)", a, b)
			}
		}
	})
}

// K-13, K-14, K-23: Sub and Distance on a grid of octaves at the limits of
// int, against exact integer arithmetic. Sub fails exactly when a - b does
// not fit in an int; Distance fails exactly when Sub fails or gives MIN
// (architecture, KR-OC9). Errors are ErrOutOfRange without
// number.ErrOutOfRange, with the value 0.
func TestSubDistanceLimits(t *testing.T) {
	t.Parallel()

	for _, a := range edgeASPN() {
		for _, b := range edgeASPN() {
			var (
				sub, dist       int
				errSub, errDist error
			)
			require.NotPanics(t, func() { sub, errSub = oct(a).Sub(oct(b)) }, "Sub(O(%d), O(%d))", a, b)
			require.NotPanics(t, func() { dist, errDist = oct(a).Distance(oct(b)) }, "Distance(O(%d), O(%d))", a, b)

			d, ok := exactSub(a, b)
			if ok {
				require.NoError(t, errSub, "Sub(O(%d), O(%d))", a, b)
				assert.Equal(t, d, sub, "Sub(O(%d), O(%d))", a, b)
			} else {
				assertCause(t, errSub, causeOutOfRange, fmt.Sprintf("Sub(O(%d), O(%d))", a, b))
				assert.Zero(t, sub, "Sub(O(%d), O(%d)) value on error", a, b)
			}

			if ok && d != minInt {
				require.NoError(t, errDist, "Distance(O(%d), O(%d))", a, b)
				assert.Equal(t, max(d, -d), dist, "Distance(O(%d), O(%d))", a, b)
			} else {
				assertCause(t, errDist, causeOutOfRange, fmt.Sprintf("Distance(O(%d), O(%d))", a, b))
				assert.Zero(t, dist, "Distance(O(%d), O(%d)) value on error", a, b)
			}
		}
	}
}

// K-23, K-14: the named out-of-range cases of movement: ErrOutOfRange without
// number.ErrOutOfRange, the invalid value or 0, and Must variants panic with
// the same cause.
func TestMoveOutOfRange(t *testing.T) {
	t.Parallel()

	octaveCases := []struct {
		name string
		f    func() (octave.Octave, error)
		must func()
	}{
		{
			name: "O(MAX).Shift(1)",
			f:    func() (octave.Octave, error) { return oct(maxInt).Shift(1) },
			must: func() { oct(maxInt).MustShift(1) },
		},
		{
			name: "O(MIN).Shift(-1)",
			f:    func() (octave.Octave, error) { return oct(minInt).Shift(-1) },
			must: func() { oct(minInt).MustShift(-1) },
		},
		{
			name: "O(-1).Shift(MIN)",
			f:    func() (octave.Octave, error) { return oct(-1).Shift(minInt) },
			must: func() { oct(-1).MustShift(minInt) },
		},
		{
			name: "O(MIN).ShiftDown(1)",
			f:    func() (octave.Octave, error) { return oct(minInt).ShiftDown(1) },
			must: func() { oct(minInt).MustShiftDown(1) },
		},
		{
			name: "O(MAX).ShiftDown(-1)",
			f:    func() (octave.Octave, error) { return oct(maxInt).ShiftDown(-1) },
			must: func() { oct(maxInt).MustShiftDown(-1) },
		},
		{
			name: "O(0).ShiftDown(MIN)",
			f:    func() (octave.Octave, error) { return oct(0).ShiftDown(minInt) },
			must: func() { oct(0).MustShiftDown(minInt) },
		},
		{
			name: "O(MAX).Up()",
			f:    func() (octave.Octave, error) { return oct(maxInt).Up() },
			must: func() { oct(maxInt).MustUp() },
		},
		{
			name: "O(MIN).Down()",
			f:    func() (octave.Octave, error) { return oct(minInt).Down() },
			must: func() { oct(minInt).MustDown() },
		},
	}

	for _, tc := range octaveCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got octave.Octave
				err error
			)
			require.NotPanics(t, func() { got, err = tc.f() })
			assertCause(t, err, causeOutOfRange, tc.name)
			assert.Equal(t, octave.Octave{}, got, "value on error")
			assertPanicsCause(t, causeOutOfRange, tc.must, "Must variant of "+tc.name)
		})
	}

	intCases := []struct {
		name string
		f    func() (int, error)
		must func()
	}{
		{
			name: "O(MAX).Sub(O(-1))",
			f:    func() (int, error) { return oct(maxInt).Sub(oct(-1)) },
			must: func() { oct(maxInt).MustSub(oct(-1)) },
		},
		{
			name: "O(0).Sub(O(MIN))",
			f:    func() (int, error) { return oct(0).Sub(oct(minInt)) },
			must: func() { oct(0).MustSub(oct(minInt)) },
		},
		{
			name: "O(MAX).Sub(O(MIN))",
			f:    func() (int, error) { return oct(maxInt).Sub(oct(minInt)) },
			must: func() { oct(maxInt).MustSub(oct(minInt)) },
		},
		{
			name: "O(MIN).Sub(O(1))",
			f:    func() (int, error) { return oct(minInt).Sub(oct(1)) },
			must: func() { oct(minInt).MustSub(oct(1)) },
		},
		{
			name: "O(MIN).Distance(O(MAX))",
			f:    func() (int, error) { return oct(minInt).Distance(oct(maxInt)) },
			must: func() { oct(minInt).MustDistance(oct(maxInt)) },
		},
		{
			name: "O(MIN).Distance(O(0))",
			f:    func() (int, error) { return oct(minInt).Distance(oct(0)) },
			must: func() { oct(minInt).MustDistance(oct(0)) },
		},
		{
			name: "O(0).Distance(O(MIN))",
			f:    func() (int, error) { return oct(0).Distance(oct(minInt)) },
			must: func() { oct(0).MustDistance(oct(minInt)) },
		},
	}

	for _, tc := range intCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got int
				err error
			)
			require.NotPanics(t, func() { got, err = tc.f() })
			assertCause(t, err, causeOutOfRange, tc.name)
			assert.Zero(t, got, "value on error")
			assertPanicsCause(t, causeOutOfRange, tc.must, "Must variant of "+tc.name)
		})
	}
}

// K-26, G-1: far octaves move without errors.
func TestFarOctavesMove(t *testing.T) {
	t.Parallel()

	for _, a := range []int{-1000, 1000} {
		for _, k := range []int{-1, 1} {
			got, err := oct(a).Shift(k)
			require.NoError(t, err, "O(%d).Shift(%d)", a, k)
			assert.Equal(t, oct(a+k), got, "O(%d).Shift(%d)", a, k)
		}
	}

	d, err := oct(1000).Sub(oct(-1000))
	require.NoError(t, err)
	assert.Equal(t, 2000, d)
}
