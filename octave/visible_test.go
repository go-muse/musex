package octave_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-3, C-3: the numbers of the octave of middle C and of the octaves next to
// the skipped number of Franco-Belgian.
func TestNumberKnownValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		aspn int
		n    number.Numbering
		want int
	}{
		{name: "O(4) ASPN", aspn: 4, n: number.ASPN(), want: 4},
		{name: "O(4) FB", aspn: 4, n: fb(), want: 3},
		{name: "O(4) Mex", aspn: 4, n: mex(), want: 5},
		{name: "O(4) C3", aspn: 4, n: c3(), want: 3},
		{name: "O(4) C5", aspn: 4, n: c5(), want: 5},
		{name: "O(2) FB", aspn: 2, n: fb(), want: 1},
		{name: "O(1) FB", aspn: 1, n: fb(), want: -1},
		{name: "O(0) FB", aspn: 0, n: fb(), want: -2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := oct(tc.aspn).Number(tc.n)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.want, oct(tc.aspn).MustNumber(tc.n))
		})
	}
}

// K-3, K-31: for every a in -20...20 and every numbering of K-1, the number
// of O(a) is the number given by the formula of the numbering and by
// octave/number for ASPN a; it is never a skipped number; Number does not
// fail and does not panic.
func TestNumberMatchesOracle(t *testing.T) {
	t.Parallel()

	for _, nc := range criteriaNumberings() {
		t.Run(nc.name, func(t *testing.T) {
			t.Parallel()

			for _, a := range span(-20, 20) {
				want, ok := nc.number(a)
				require.True(t, ok, "oracle: %s ASPN %d", nc.name, a)

				var (
					got int
					err error
				)
				require.NotPanics(t, func() { got, err = oct(a).Number(nc.n) }, "O(%d).Number(%s)", a, nc.name)
				require.NoError(t, err, "O(%d).Number(%s)", a, nc.name)
				assert.Equal(t, want, got, "N(O(%d), %s)", a, nc.name)
				assert.Equal(t, nc.n.MustFromASPN(a), got, "N(O(%d), %s) vs octave/number", a, nc.name)
				assert.NotContains(t, nc.skipped, got, "N(O(%d), %s) is skipped", a, nc.name)
				assert.Equal(t, want, oct(a).MustNumber(nc.n), "O(%d).MustNumber(%s)", a, nc.name)
			}
		})
	}
}

// K-23, G-2, K-31: numbers at the limits of int. A number that does not fit
// in an int gives ErrOutOfRange that also matches number.ErrOutOfRange, with
// the value 0; the ASPN number of every octave is given without an error.
func TestNumberLimits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		aspn int
		n    number.Numbering
		want int
		err  bool
	}{
		{name: "O(MAX) C5", aspn: maxInt, n: c5(), err: true},
		{name: "O(MAX) Mex", aspn: maxInt, n: mex(), err: true},
		{name: "O(MIN) C3", aspn: minInt, n: c3(), err: true},
		{name: "O(MIN) FB", aspn: minInt, n: fb(), err: true},
		{name: "O(MIN+1) FB", aspn: minInt + 1, n: fb(), err: true},
		{name: "O(MIN) ASPN", aspn: minInt, n: number.ASPN(), want: minInt},
		{name: "O(MAX) ASPN", aspn: maxInt, n: number.ASPN(), want: maxInt},
		{name: "O(MAX) C3", aspn: maxInt, n: c3(), want: maxInt - 1},
		{name: "O(MAX) FB", aspn: maxInt, n: fb(), want: maxInt - 1},
		{name: "O(MIN+2) FB", aspn: minInt + 2, n: fb(), want: minInt},
		{name: "O(MIN) C5", aspn: minInt, n: c5(), want: minInt + 1},
		{name: "O(MAX-1) Mex", aspn: maxInt - 1, n: mex(), want: maxInt},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got int
				err error
			)
			require.NotPanics(t, func() { got, err = oct(tc.aspn).Number(tc.n) })

			if tc.err {
				assertCause(t, err, causeNumberOutOfRange, "Number")
				assert.Zero(t, got, "value on error")
				assertPanicsCause(t, causeNumberOutOfRange, func() { oct(tc.aspn).MustNumber(tc.n) }, "MustNumber")

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// C-3, K-23, K-31, B-9: the ASPN number of every octave, including the
// limits, without an error; ASPN is Number(number.ASPN()).
func TestASPN(t *testing.T) {
	t.Parallel()

	for _, a := range append(span(-20, 20), edgeASPN()...) {
		var (
			got int
			err error
		)
		require.NotPanics(t, func() { got, err = oct(a).ASPN() }, "O(%d).ASPN()", a)
		require.NoError(t, err, "O(%d).ASPN()", a)
		assert.Equal(t, a, got, "O(%d).ASPN()", a)
		assert.Equal(t, a, oct(a).MustASPN(), "O(%d).MustASPN()", a)

		viaNumber, err := oct(a).Number(number.ASPN())
		require.NoError(t, err, "O(%d).Number(ASPN)", a)
		assert.Equal(t, a, viaNumber, "O(%d).Number(ASPN)", a)
	}
}

// K-21, E-2: the number in an undefined numbering gives number.ErrUndefined
// with the value 0; the invalid value is checked before the numbering and
// gives ErrInvalid.
func TestNumberUndefined(t *testing.T) {
	t.Parallel()

	for _, a := range []int{4, 0, minInt, maxInt} {
		t.Run(fmt.Sprint(a), func(t *testing.T) {
			t.Parallel()

			got, err := oct(a).Number(number.Numbering{})
			assertCause(t, err, causeUndefined, "Number(undefined)")
			assert.Zero(t, got, "value on error")
			assertPanicsCause(t, causeUndefined, func() { oct(a).MustNumber(number.Numbering{}) }, "MustNumber(undefined)")
		})
	}

	t.Run("invalid value", func(t *testing.T) {
		t.Parallel()

		got, err := octave.Octave{}.Number(number.Numbering{})
		assertCause(t, err, causeInvalid, "Z.Number(undefined)")
		assert.Zero(t, got, "value on error")
		assertPanicsCause(t, causeInvalid, func() { octave.Octave{}.MustNumber(number.Numbering{}) }, "Z.MustNumber(undefined)")
	})
}

// K-4, C-4, G-5: names of the octave of middle C, of an octave named only in
// the Russian catalog, and of octaves without names.
func TestNameKnownValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		aspn       int
		ru, en, de string
	}{
		{aspn: 4, ru: "первая", en: "one-line", de: "eingestrichene"},
		{aspn: 8, ru: "пятая", en: "", de: ""},
		{aspn: -2, ru: "", en: "", de: ""},
		{aspn: 10, ru: "", en: "", de: ""},
		{aspn: 1000, ru: "", en: "", de: ""},
		{aspn: -1000, ru: "", en: "", de: ""},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.aspn), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.ru, oct(tc.aspn).Name(naming.Russian()), "Russian")
			assert.Equal(t, tc.en, oct(tc.aspn).Name(naming.English()), "English")
			assert.Equal(t, tc.de, oct(tc.aspn).Name(naming.German()), "German")
		})
	}
}

// K-4, K-31: for every a in -20...20 and each built-in catalog, the name of
// O(a) is the name of octave/naming for ASPN a (the literal tables of the
// catalogs), or "" when there is none; Name never panics.
func TestNameMatchesOracle(t *testing.T) {
	t.Parallel()

	for _, cc := range builtinCatalogs() {
		t.Run(cc.name, func(t *testing.T) {
			t.Parallel()

			for _, a := range append(span(-20, 20), edgeASPN()...) {
				var got string
				require.NotPanics(t, func() { got = oct(a).Name(cc.c) }, "O(%d).Name(%s)", a, cc.name)
				assert.Equal(t, cc.names[a], got, "H(O(%d), %s)", a, cc.name)
				assert.Equal(t, cc.c.Name(a), got, "H(O(%d), %s) vs octave/naming", a, cc.name)
			}
		})
	}
}

// K-4, G-5: catalogs without names, including the zero Catalog and the nil
// *Catalog, give "" without a panic.
func TestNameInEmptyCatalogs(t *testing.T) {
	t.Parallel()

	for _, cc := range emptyCatalogs() {
		t.Run(cc.name, func(t *testing.T) {
			t.Parallel()

			for _, a := range []int{-1, 4, 100, minInt, maxInt} {
				var got string
				require.NotPanics(t, func() { got = oct(a).Name(cc.c) }, "O(%d).Name(%s)", a, cc.name)
				assert.Empty(t, got, "O(%d).Name(%s)", a, cc.name)
			}
		})
	}
}

// K-5, C-5: translation between conventions, the scenario of README Use
// Cases.
func TestTranslateConventions(t *testing.T) {
	t.Parallel()

	t.Run("Franco-Belgian 3", func(t *testing.T) {
		t.Parallel()

		o := octave.MustFromNumber(3, fb())
		assert.Equal(t, "первая", o.Name(naming.Russian()))
		assert.Equal(t, 4, o.MustNumber(number.ASPN()))
		assert.Equal(t, "one-line", o.Name(naming.English()))
	})

	t.Run("Russian малая", func(t *testing.T) {
		t.Parallel()

		o := octave.MustFromName("малая", naming.Russian())
		assert.Equal(t, 2, o.MustNumber(fb()))
		assert.Equal(t, 3, o.MustNumber(number.ASPN()))
	})

	t.Run("English great", func(t *testing.T) {
		t.Parallel()

		o := octave.MustFromName("great", naming.English())
		assert.Equal(t, "большая", o.Name(naming.Russian()))
		assert.Equal(t, 1, o.MustNumber(c3()))
	})
}

// K-26, G-1, NF-6: octaves far outside the usual range are ordinary: numbers
// in the five built-in numberings by their formulas and empty names in the
// three catalogs, without errors.
func TestFarOctavesNumbersAndNames(t *testing.T) {
	t.Parallel()

	for _, a := range []int{-1000, 1000} {
		t.Run(fmt.Sprint(a), func(t *testing.T) {
			t.Parallel()

			for _, nc := range builtinNumberings() {
				want, ok := nc.number(a)
				require.True(t, ok, "oracle")

				got, err := oct(a).Number(nc.n)
				require.NoError(t, err, "Number(%s)", nc.name)
				assert.Equal(t, want, got, "Number(%s)", nc.name)
				assert.False(t, slices.Contains(nc.skipped, got), "Number(%s) is skipped", nc.name)
			}
			for _, cc := range builtinCatalogs() {
				assert.Empty(t, oct(a).Name(cc.c), "Name(%s)", cc.name)
			}
		})
	}
}
