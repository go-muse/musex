package octave_test

import (
	"errors"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-25, NF-7: the two sentinel errors of the package exist and are distinct
// from each other and from the sentinel errors of octave/number and
// octave/naming; none of them matches another with errors.Is.
func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	require.Error(t, octave.ErrInvalid)
	require.Error(t, octave.ErrOutOfRange)

	all := allSentinels()
	for i, a := range all {
		for j, b := range all {
			assert.Equal(t, i == j, errors.Is(a, b), "errors.Is(%s, %s)", sentinelName(a), sentinelName(b))
		}
	}
}

// errorCase is one row of the table "operation x input -> cause".
type errorCase struct {
	name string
	call func() error
	must func()
	want cause
}

// errorCases returns the cases E-1...E-7 of the test set, one or more per
// operation with an error.
func errorCases() []errorCase {
	var z octave.Octave
	ignoreOctave := func(_ octave.Octave, err error) error { return err }
	ignoreInt := func(_ int, err error) error { return err }

	return []errorCase{
		// E-1: skipped number.
		{
			name: "FromNumber(0, FB)",
			call: func() error { return ignoreOctave(octave.FromNumber(0, fb())) },
			must: func() { octave.MustFromNumber(0, fb()) },
			want: causeSkipped,
		},
		{
			name: "FromNumber(6, New(4, 4, 6))",
			call: func() error { return ignoreOctave(octave.FromNumber(6, number.MustNew(4, 4, 6))) },
			must: func() { octave.MustFromNumber(6, number.MustNew(4, 4, 6)) },
			want: causeSkipped,
		},
		// E-2: undefined numbering.
		{
			name: "FromNumber(4, undefined)",
			call: func() error { return ignoreOctave(octave.FromNumber(4, number.Numbering{})) },
			must: func() { octave.MustFromNumber(4, number.Numbering{}) },
			want: causeUndefined,
		},
		{
			name: "FromNumber(MAX, undefined)",
			call: func() error { return ignoreOctave(octave.FromNumber(maxInt, number.Numbering{})) },
			must: func() { octave.MustFromNumber(maxInt, number.Numbering{}) },
			want: causeUndefined,
		},
		{
			name: "O(4).Number(undefined)",
			call: func() error { return ignoreInt(oct(4).Number(number.Numbering{})) },
			must: func() { oct(4).MustNumber(number.Numbering{}) },
			want: causeUndefined,
		},
		// E-3: unknown name.
		{
			name: "FromName(Первая, Ru)",
			call: func() error { return ignoreOctave(octave.FromName("Первая", naming.Russian())) },
			must: func() { octave.MustFromName("Первая", naming.Russian()) },
			want: causeUnknownName,
		},
		{
			name: "FromName(\"\", nil)",
			call: func() error { return ignoreOctave(octave.FromName("", nil)) },
			must: func() { octave.MustFromName("", nil) },
			want: causeUnknownName,
		},
		// E-4 found by octave/number.
		{
			name: "FromNumber(MAX, C3)",
			call: func() error { return ignoreOctave(octave.FromNumber(maxInt, c3())) },
			must: func() { octave.MustFromNumber(maxInt, c3()) },
			want: causeNumberOutOfRange,
		},
		{
			name: "FromNumber(MIN, C5)",
			call: func() error { return ignoreOctave(octave.FromNumber(minInt, c5())) },
			must: func() { octave.MustFromNumber(minInt, c5()) },
			want: causeNumberOutOfRange,
		},
		{
			name: "O(MAX).Number(Mex)",
			call: func() error { return ignoreInt(oct(maxInt).Number(mex())) },
			must: func() { oct(maxInt).MustNumber(mex()) },
			want: causeNumberOutOfRange,
		},
		{
			name: "O(MIN+1).Number(FB)",
			call: func() error { return ignoreInt(oct(minInt + 1).Number(fb())) },
			must: func() { oct(minInt + 1).MustNumber(fb()) },
			want: causeNumberOutOfRange,
		},
		// E-4 found by the arithmetic of octave.
		{
			name: "O(MAX).Shift(1)",
			call: func() error { return ignoreOctave(oct(maxInt).Shift(1)) },
			must: func() { oct(maxInt).MustShift(1) },
			want: causeOutOfRange,
		},
		{
			name: "O(MIN).ShiftDown(1)",
			call: func() error { return ignoreOctave(oct(minInt).ShiftDown(1)) },
			must: func() { oct(minInt).MustShiftDown(1) },
			want: causeOutOfRange,
		},
		{
			name: "O(MAX).Up()",
			call: func() error { return ignoreOctave(oct(maxInt).Up()) },
			must: func() { oct(maxInt).MustUp() },
			want: causeOutOfRange,
		},
		{
			name: "O(MIN).Down()",
			call: func() error { return ignoreOctave(oct(minInt).Down()) },
			must: func() { oct(minInt).MustDown() },
			want: causeOutOfRange,
		},
		{
			name: "O(0).Sub(O(MIN))",
			call: func() error { return ignoreInt(oct(0).Sub(oct(minInt))) },
			must: func() { oct(0).MustSub(oct(minInt)) },
			want: causeOutOfRange,
		},
		{
			name: "O(MIN).Distance(O(0))",
			call: func() error { return ignoreInt(oct(minInt).Distance(oct(0))) },
			must: func() { oct(minInt).MustDistance(oct(0)) },
			want: causeOutOfRange,
		},
		// E-5: the invalid value.
		{
			name: "Z.Number(ASPN)",
			call: func() error { return ignoreInt(z.Number(number.ASPN())) },
			must: func() { z.MustNumber(number.ASPN()) },
			want: causeInvalid,
		},
		{
			name: "Z.Number(undefined)",
			call: func() error { return ignoreInt(z.Number(number.Numbering{})) },
			must: func() { z.MustNumber(number.Numbering{}) },
			want: causeInvalid,
		},
		{
			name: "Z.ASPN()",
			call: func() error { return ignoreInt(z.ASPN()) },
			must: func() { z.MustASPN() },
			want: causeInvalid,
		},
		{
			name: "Z.Shift(MAX)",
			call: func() error { return ignoreOctave(z.Shift(maxInt)) },
			must: func() { z.MustShift(maxInt) },
			want: causeInvalid,
		},
		{
			name: "Z.ShiftDown(MIN)",
			call: func() error { return ignoreOctave(z.ShiftDown(minInt)) },
			must: func() { z.MustShiftDown(minInt) },
			want: causeInvalid,
		},
		{
			name: "Z.Up()",
			call: func() error { return ignoreOctave(z.Up()) },
			must: func() { z.MustUp() },
			want: causeInvalid,
		},
		{
			name: "Z.Down()",
			call: func() error { return ignoreOctave(z.Down()) },
			must: func() { z.MustDown() },
			want: causeInvalid,
		},
		{
			name: "O(MIN).Sub(Z)",
			call: func() error { return ignoreInt(oct(minInt).Sub(z)) },
			must: func() { oct(minInt).MustSub(z) },
			want: causeInvalid,
		},
		{
			name: "Z.Distance(O(MAX))",
			call: func() error { return ignoreInt(z.Distance(oct(maxInt))) },
			must: func() { z.MustDistance(oct(maxInt)) },
			want: causeInvalid,
		},
		{
			name: "Range(Z, O(4))",
			call: func() error { _, err := octave.Range(z, oct(4)); return err },
			must: func() { octave.MustRange(z, oct(4)) },
			want: causeInvalid,
		},
		{
			name: "NewNumbering(Z, 0, 0)",
			call: func() error { _, err := octave.NewNumbering(z, 0, 0); return err },
			must: func() { octave.MustNewNumbering(z, 0, 0) },
			want: causeInvalid,
		},
		{
			name: "NewCatalog({Z: a, O(1): a})",
			call: func() error {
				_, err := octave.NewCatalog(map[octave.Octave]string{z: "a", oct(1): "a"})
				return err
			},
			must: func() { octave.MustNewCatalog(map[octave.Octave]string{z: "a", oct(1): "a"}) },
			want: causeInvalid,
		},
		// E-6: invalid definition of an own numbering.
		{
			name: "NewNumbering(O(4), 0, 0)",
			call: func() error { _, err := octave.NewNumbering(oct(4), 0, 0); return err },
			must: func() { octave.MustNewNumbering(oct(4), 0, 0) },
			want: causeInvalidDefinition,
		},
		{
			name: "NewNumbering(O(4), 3, 0, 0)",
			call: func() error { _, err := octave.NewNumbering(oct(4), 3, 0, 0); return err },
			must: func() { octave.MustNewNumbering(oct(4), 3, 0, 0) },
			want: causeInvalidDefinition,
		},
		// E-7: duplicate name in an own catalog.
		{
			name: "NewCatalog({O(1): a, O(2): a})",
			call: func() error {
				_, err := octave.NewCatalog(map[octave.Octave]string{oct(1): "a", oct(2): "a"})
				return err
			},
			must: func() { octave.MustNewCatalog(map[octave.Octave]string{oct(1): "a", oct(2): "a"}) },
			want: causeDuplicateName,
		},
	}
}

// K-25, E-1...E-7, NF-7: each error matches exactly its cause with errors.Is
// and none of the other six (an out-of-range error found by octave/number
// also matches number.ErrOutOfRange, an out-of-range error of the arithmetic
// does not); the Must variant panics with an error of the same cause.
func TestErrorCauses(t *testing.T) {
	t.Parallel()

	for _, tc := range errorCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var err error
			require.NotPanics(t, func() { err = tc.call() })
			assertCause(t, err, tc.want, tc.name)
			assertPanicsCause(t, tc.want, tc.must, "Must variant of "+tc.name)
		})
	}
}

// K-23, NF-7: a single check errors.Is(err, octave.ErrOutOfRange) is true for
// every out-of-range error of the package, whichever part found it.
func TestOutOfRangeSingleCheck(t *testing.T) {
	t.Parallel()

	for _, tc := range errorCases() {
		if tc.want.name != causeOutOfRange.name && tc.want.name != causeNumberOutOfRange.name {
			continue
		}
		assert.ErrorIs(t, tc.call(), octave.ErrOutOfRange, tc.name)
	}
}

// K-25, dev.md "Must": on inputs without an error, each of the 13 Must
// variants returns the same result as the operation it wraps and does not
// panic.
func TestMustMatchesOnSuccess(t *testing.T) {
	t.Parallel()

	t.Run("MustFromNumber", func(t *testing.T) {
		t.Parallel()

		for _, nc := range criteriaNumberings() {
			for _, a := range []int{-20, -1, 0, 1, 2, 4, 20} {
				x, ok := nc.number(a)
				require.True(t, ok, "oracle")
				want, err := octave.FromNumber(x, nc.n)
				require.NoError(t, err, "FromNumber(%d, %s)", x, nc.name)
				var got octave.Octave
				require.NotPanics(t, func() { got = octave.MustFromNumber(x, nc.n) }, "MustFromNumber(%d, %s)", x, nc.name)
				assert.Equal(t, want, got, "MustFromNumber(%d, %s)", x, nc.name)
			}
		}
	})

	t.Run("MustFromName", func(t *testing.T) {
		t.Parallel()

		for _, cc := range builtinCatalogs() {
			for _, s := range cc.names {
				want, err := octave.FromName(s, cc.c)
				require.NoError(t, err, "FromName(%q)", s)
				var got octave.Octave
				require.NotPanics(t, func() { got = octave.MustFromName(s, cc.c) }, "MustFromName(%q)", s)
				assert.Equal(t, want, got, "MustFromName(%q)", s)
			}
		}
	})

	t.Run("methods", func(t *testing.T) {
		t.Parallel()

		for _, a := range []int{minInt + 2, -5, 0, 4, 5, maxInt - 2} {
			o := oct(a)
			for _, nc := range builtinNumberings() {
				want, err := o.Number(nc.n)
				require.NoError(t, err, "O(%d).Number(%s)", a, nc.name)
				assert.Equal(t, want, o.MustNumber(nc.n), "O(%d).MustNumber(%s)", a, nc.name)
			}

			wantASPN, err := o.ASPN()
			require.NoError(t, err)
			assert.Equal(t, wantASPN, o.MustASPN(), "O(%d).MustASPN()", a)

			for _, k := range []int{-2, 0, 2} {
				up, err := o.Shift(k)
				require.NoError(t, err, "O(%d).Shift(%d)", a, k)
				assert.Equal(t, up, o.MustShift(k), "O(%d).MustShift(%d)", a, k)

				down, err := o.ShiftDown(k)
				require.NoError(t, err, "O(%d).ShiftDown(%d)", a, k)
				assert.Equal(t, down, o.MustShiftDown(k), "O(%d).MustShiftDown(%d)", a, k)
			}

			next, err := o.Up()
			require.NoError(t, err)
			assert.Equal(t, next, o.MustUp(), "O(%d).MustUp()", a)

			prev, err := o.Down()
			require.NoError(t, err)
			assert.Equal(t, prev, o.MustDown(), "O(%d).MustDown()", a)

			for _, b := range []int{-3, 0, 4} {
				sub, err := o.Sub(oct(b))
				if err == nil {
					assert.Equal(t, sub, o.MustSub(oct(b)), "O(%d).MustSub(O(%d))", a, b)
				}
				dist, err := o.Distance(oct(b))
				if err == nil {
					assert.Equal(t, dist, o.MustDistance(oct(b)), "O(%d).MustDistance(O(%d))", a, b)
				}
			}
		}
	})

	t.Run("MustRange", func(t *testing.T) {
		t.Parallel()

		for _, ends := range [][2]int{{2, 5}, {5, 2}, {4, 4}, {maxInt - 2, maxInt}, {minInt, maxInt}} {
			seq, err := octave.Range(oct(ends[0]), oct(ends[1]))
			require.NoError(t, err, "Range(O(%d), O(%d))", ends[0], ends[1])
			assert.Equal(t, collect(seq, 10), collect(octave.MustRange(oct(ends[0]), oct(ends[1])), 10),
				"MustRange(O(%d), O(%d))", ends[0], ends[1])
		}
	})

	t.Run("MustNewNumbering", func(t *testing.T) {
		t.Parallel()

		want, err := octave.NewNumbering(oct(4), 3, 0)
		require.NoError(t, err)
		got := octave.MustNewNumbering(oct(4), 3, 0)
		assert.True(t, got.Equal(want), "MustNewNumbering(O(4), 3, 0): %v, %v", got, want)
		assert.Equal(t, want.Label(), got.Label())
	})

	t.Run("MustNewCatalog", func(t *testing.T) {
		t.Parallel()

		pairs := map[octave.Octave]string{oct(4): "middle", oct(-3): "deep"}
		want, err := octave.NewCatalog(pairs)
		require.NoError(t, err)
		got := octave.MustNewCatalog(pairs)
		require.NotNil(t, got)
		assert.True(t, got.Equal(want))
	})
}

// K-31, NF-5: for all a, b in -20...20, the operations of C-3, C-4 and
// C-6...C-15 succeed without errors and panics (numbers, names, equality,
// order, shifts by b, neighbors, difference, distance, range).
func TestUsualRangeHasNoErrors(t *testing.T) {
	t.Parallel()

	numberings := criteriaNumberings()
	catalogs := builtinCatalogs()

	for _, a := range span(-20, 20) {
		o := oct(a)
		for _, nc := range numberings {
			assert.NotPanics(t, func() {
				_, err := o.Number(nc.n)
				assert.NoError(t, err, "O(%d).Number(%s)", a, nc.name)
			})
		}
		for _, cc := range catalogs {
			assert.NotPanics(t, func() { _ = o.Name(cc.c) }, "O(%d).Name(%s)", a, cc.name)
		}

		assert.NotPanics(t, func() {
			_, err := o.Up()
			assert.NoError(t, err, "O(%d).Up()", a)
			_, err = o.Down()
			assert.NoError(t, err, "O(%d).Down()", a)
		})

		for _, b := range span(-20, 20) {
			p := oct(b)
			assert.NotPanics(t, func() {
				_ = o.Equal(p)
				_ = o.Compare(p)
				_ = o.Above(p)
				_ = o.Below(p)
				_ = o.Between(p, o)

				_, err := o.Shift(b)
				assert.NoError(t, err, "O(%d).Shift(%d)", a, b)
				_, err = o.ShiftDown(b)
				assert.NoError(t, err, "O(%d).ShiftDown(%d)", a, b)
				_, err = o.Sub(p)
				assert.NoError(t, err, "O(%d).Sub(O(%d))", a, b)
				_, err = o.Distance(p)
				assert.NoError(t, err, "O(%d).Distance(O(%d))", a, b)

				seq, err := octave.Range(o, p)
				if assert.NoError(t, err, "Range(O(%d), O(%d))", a, b) {
					assert.Len(t, collect(seq, rangeLimit), max(a-b, b-a)+1, "Range(O(%d), O(%d))", a, b)
				}
			}, "a=%d b=%d", a, b)
		}
	}
}
