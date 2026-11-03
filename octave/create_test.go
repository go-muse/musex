package octave_test

import (
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// C-1, G-1, G-2, B-9: FromASPN accepts every int, including the limits, and
// gives a valid octave whose ASPN number is the argument; it is the same
// octave as MustFromNumber(a, number.ASPN()).
func TestFromASPN(t *testing.T) {
	t.Parallel()

	for _, a := range append(span(-20, 20), edgeASPN()...) {
		t.Run(fmt.Sprint(a), func(t *testing.T) {
			t.Parallel()

			var o octave.Octave
			require.NotPanics(t, func() { o = octave.FromASPN(a) }, "FromASPN(%d)", a)

			assert.True(t, o.IsValid(), "FromASPN(%d).IsValid()", a)

			got, err := o.ASPN()
			require.NoError(t, err, "FromASPN(%d).ASPN()", a)
			assert.Equal(t, a, got, "FromASPN(%d).ASPN()", a)
			assert.Equal(t, a, o.MustASPN(), "FromASPN(%d).MustASPN()", a)

			viaNumber, err := octave.FromNumber(a, number.ASPN())
			require.NoError(t, err, "FromNumber(%d, ASPN)", a)
			assert.True(t, o.Equal(viaNumber), "FromASPN(%d).Equal(FromNumber(%d, ASPN))", a, a)
			assert.True(t, o == viaNumber, "FromASPN(%d) == FromNumber(%d, ASPN)", a, a)
			assert.True(t, o.Equal(octave.MustFromNumber(a, number.ASPN())), "MustFromNumber(%d, ASPN)", a)
		})
	}
}

// K-1: the octave of middle C by its number in each of the five built-in
// numberings is one octave: pairwise Equal and ==.
func TestFromNumberMiddleC(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		num  int
		n    number.Numbering
	}{
		{name: "Franco-Belgian 3", num: 3, n: fb()},
		{name: "ASPN 4", num: 4, n: number.ASPN()},
		{name: "Mexican 5", num: 5, n: mex()},
		{name: "Middle C = C3 3", num: 3, n: c3()},
		{name: "Middle C = C5 5", num: 5, n: c5()},
	}

	got := make([]octave.Octave, 0, len(cases))
	for _, tc := range cases {
		o, err := octave.FromNumber(tc.num, tc.n)
		require.NoError(t, err, "FromNumber: %s", tc.name)
		require.True(t, o.IsValid(), "FromNumber: %s", tc.name)
		got = append(got, o)
	}

	for i, a := range got {
		for j, b := range got {
			assert.True(t, a.Equal(b), "%s Equal %s", cases[i].name, cases[j].name)
			assert.True(t, a == b, "%s == %s", cases[i].name, cases[j].name)
		}
		assert.Equal(t, 4, a.MustASPN(), "ASPN number of %s", cases[i].name)
	}
}

// K-1, C-1: for every a in -20...20 and every numbering of K-1 (five built-in,
// New(4, 5), New(4, 4, 0, 1)), the octave by the number of O(a) is O(a), and
// the number of the octave by number x is x. The numbers come from the
// formulas of the numberings, not from the package under test.
func TestFromNumberRoundTrip(t *testing.T) {
	t.Parallel()

	for _, nc := range criteriaNumberings() {
		t.Run(nc.name, func(t *testing.T) {
			t.Parallel()

			for _, a := range span(-20, 20) {
				x, ok := nc.number(a)
				require.True(t, ok, "oracle: %s ASPN %d", nc.name, a)

				got, err := octave.FromNumber(x, nc.n)
				require.NoError(t, err, "FromNumber(%d, %s)", x, nc.name)
				assert.True(t, got.Equal(oct(a)), "FromNumber(%d, %s) Equal O(%d)", x, nc.name, a)
				assert.True(t, got == oct(a), "FromNumber(%d, %s) == O(%d)", x, nc.name, a)
				assert.Equal(t, got, octave.MustFromNumber(x, nc.n), "MustFromNumber(%d, %s)", x, nc.name)

				back, err := got.Number(nc.n)
				require.NoError(t, err, "FromNumber(%d, %s).Number", x, nc.name)
				assert.Equal(t, x, back, "N(O(%d, %s), %s)", x, nc.name, nc.name)

				again, err := octave.FromNumber(oct(a).MustNumber(nc.n), nc.n)
				require.NoError(t, err, "FromNumber(O(%d).MustNumber(%s))", a, nc.name)
				assert.True(t, again.Equal(oct(a)), "O(N(O(%d), %s), %s) = O(%d)", a, nc.name, nc.name, a)
			}
		})
	}
}

// K-1, K-26: numbers next to the skipped number of Franco-Belgian and of
// Mexican below 1: O(-3, FB) = O(-1), O(0, Mex) = O(-1), O(1, FB) = O(2),
// O(-1, FB) = O(1).
func TestFromNumberKnownValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		num  int
		n    number.Numbering
		aspn int
	}{
		{name: "FB -3", num: -3, n: fb(), aspn: -1},
		{name: "Mex 0", num: 0, n: mex(), aspn: -1},
		{name: "FB 1", num: 1, n: fb(), aspn: 2},
		{name: "FB -1", num: -1, n: fb(), aspn: 1},
		{name: "FB -2", num: -2, n: fb(), aspn: 0},
		{name: "C3 1", num: 1, n: c3(), aspn: 2},
		{name: "New(4, 4, 0, 1) -1", num: -1, n: number.MustNew(4, 4, 0, 1), aspn: 1},
		{name: "New(4, 4, 0, 1) 2", num: 2, n: number.MustNew(4, 4, 0, 1), aspn: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := octave.FromNumber(tc.num, tc.n)
			require.NoError(t, err)
			assert.Equal(t, oct(tc.aspn), got)
			assert.Equal(t, tc.aspn, got.MustASPN())
		})
	}
}

// K-2, C-2: octaves by name in the built-in catalogs and in an own catalog of
// octave/naming.
func TestFromName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		catalog func() *naming.Catalog
		s       string
		aspn    int
	}{
		{name: "Russian первая", catalog: naming.Russian, s: "первая", aspn: 4},
		{name: "English one-line", catalog: naming.English, s: "one-line", aspn: 4},
		{name: "German eingestrichene", catalog: naming.German, s: "eingestrichene", aspn: 4},
		{name: "Russian субсубконтроктава", catalog: naming.Russian, s: "субсубконтроктава", aspn: -1},
		{name: "Russian шестая", catalog: naming.Russian, s: "шестая", aspn: 9},
		{name: "English subcontra", catalog: naming.English, s: "subcontra", aspn: 0},
		{name: "Russian малая", catalog: naming.Russian, s: "малая", aspn: 3},
		{name: "English great", catalog: naming.English, s: "great", aspn: 2},
		{
			name:    "own catalog middle",
			catalog: func() *naming.Catalog { return naming.MustNew(map[int]string{4: "middle"}) },
			s:       "middle",
			aspn:    4,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.catalog()
			got, err := octave.FromName(tc.s, c)
			require.NoError(t, err)
			assert.True(t, got.Equal(oct(tc.aspn)), "FromName(%q) Equal O(%d)", tc.s, tc.aspn)
			assert.True(t, got == oct(tc.aspn), "FromName(%q) == O(%d)", tc.s, tc.aspn)
			assert.Equal(t, got, octave.MustFromName(tc.s, c))
			assert.Equal(t, tc.s, got.Name(c))
		})
	}
}

// K-2: for every name s of the built-in catalogs (octave/naming K-1...K-3),
// the name of the octave found by s is s.
func TestFromNameRoundTrip(t *testing.T) {
	t.Parallel()

	for _, cc := range builtinCatalogs() {
		t.Run(cc.name, func(t *testing.T) {
			t.Parallel()

			for a, s := range cc.names {
				got, err := octave.FromName(s, cc.c)
				require.NoError(t, err, "FromName(%q, %s)", s, cc.name)
				assert.Equal(t, oct(a), got, "FromName(%q, %s)", s, cc.name)
				assert.Equal(t, s, got.Name(cc.c), "Name(FromName(%q, %s))", s, cc.name)
			}
		})
	}
}

// K-20, E-1: a number skipped in the numbering gives number.ErrSkipped and
// the invalid value; no nearby octave is substituted. The Must variant panics
// with the same cause.
func TestFromNumberSkipped(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		num  int
		n    number.Numbering
	}{
		{name: "FB 0", num: 0, n: fb()},
		{name: "New(4, 4, 0, 1) 0", num: 0, n: number.MustNew(4, 4, 0, 1)},
		{name: "New(4, 4, 0, 1) 1", num: 1, n: number.MustNew(4, 4, 0, 1)},
		{name: "New(4, 4, 6) 6", num: 6, n: number.MustNew(4, 4, 6)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got octave.Octave
				err error
			)
			require.NotPanics(t, func() { got, err = octave.FromNumber(tc.num, tc.n) })
			assertCause(t, err, causeSkipped, "FromNumber")
			assert.Equal(t, octave.Octave{}, got, "value on error")
			assert.False(t, got.IsValid(), "value on error")

			assertPanicsCause(t, causeSkipped, func() { octave.MustFromNumber(tc.num, tc.n) }, "MustFromNumber")
		})
	}
}

// K-21, E-2: an undefined numbering gives number.ErrUndefined; ASPN is not
// used instead.
func TestFromNumberUndefined(t *testing.T) {
	t.Parallel()

	for _, num := range []int{4, 0, -1, minInt, maxInt} {
		t.Run(fmt.Sprint(num), func(t *testing.T) {
			t.Parallel()

			var (
				got octave.Octave
				err error
			)
			require.NotPanics(t, func() { got, err = octave.FromNumber(num, number.Numbering{}) })
			assertCause(t, err, causeUndefined, "FromNumber(undefined)")
			assert.Equal(t, octave.Octave{}, got, "value on error")

			assertPanicsCause(t, causeUndefined, func() { octave.MustFromNumber(num, number.Numbering{}) }, "MustFromNumber(undefined)")
		})
	}
}

// K-22, E-3, G-5: a name that is not in the catalog gives
// naming.ErrUnknownName and the invalid value; the Must variant panics.
func TestFromNameUnknown(t *testing.T) {
	t.Parallel()

	type entry struct {
		catalog string
		c       *naming.Catalog
		s       string
	}

	entries := []entry{
		{catalog: "Russian", c: naming.Russian(), s: ""},
		{catalog: "Russian", c: naming.Russian(), s: "Первая"},
		{catalog: "Russian", c: naming.Russian(), s: " первая"},
		{catalog: "Russian", c: naming.Russian(), s: "первая "},
		{catalog: "Russian", c: naming.Russian(), s: "четвертая"},
		{catalog: "Russian", c: naming.Russian(), s: "one-line"},
		{catalog: "English", c: naming.English(), s: ""},
		{catalog: "English", c: naming.English(), s: "первая"},
		{catalog: "English", c: naming.English(), s: "Great"},
		{catalog: "German", c: naming.German(), s: ""},
		{catalog: "German", c: naming.German(), s: "große Oktave"},
		{catalog: "own", c: naming.MustNew(map[int]string{4: "middle"}), s: "first"},
	}
	for _, ec := range emptyCatalogs() {
		for _, s := range []string{"", "первая", "one-line", "middle"} {
			entries = append(entries, entry{catalog: ec.name, c: ec.c, s: s})
		}
	}

	for _, e := range entries {
		t.Run(fmt.Sprintf("%s %q", e.catalog, e.s), func(t *testing.T) {
			t.Parallel()

			var (
				got octave.Octave
				err error
			)
			require.NotPanics(t, func() { got, err = octave.FromName(e.s, e.c) })
			assertCause(t, err, causeUnknownName, "FromName")
			assert.Equal(t, octave.Octave{}, got, "value on error")
			assert.False(t, got.IsValid(), "value on error")

			assertPanicsCause(t, causeUnknownName, func() { octave.MustFromName(e.s, e.c) }, "MustFromName")
		})
	}
}

// K-23, G-2, G-3: octaves by numbers at the limits of int. A number whose
// ASPN number does not fit in an int gives ErrOutOfRange that also matches
// number.ErrOutOfRange; a number whose ASPN number fits gives the octave
// (counterexample of G-3: MIN in Middle C = C3 is ASPN MIN + 1).
func TestFromNumberLimits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		num  int
		n    number.Numbering
		aspn int
		err  bool
	}{
		{name: "MIN in C5", num: minInt, n: c5(), err: true},
		{name: "MAX in C3", num: maxInt, n: c3(), err: true},
		{name: "MIN in Mex", num: minInt, n: mex(), err: true},
		{name: "MAX in FB", num: maxInt, n: fb(), err: true},
		{name: "MIN in C3", num: minInt, n: c3(), aspn: minInt + 1},
		{name: "MAX in C5", num: maxInt, n: c5(), aspn: maxInt - 1},
		{name: "MAX in Mex", num: maxInt, n: mex(), aspn: maxInt - 1},
		{name: "MIN in FB", num: minInt, n: fb(), aspn: minInt + 2},
		{name: "MIN in ASPN", num: minInt, n: number.ASPN(), aspn: minInt},
		{name: "MAX in ASPN", num: maxInt, n: number.ASPN(), aspn: maxInt},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got octave.Octave
				err error
			)
			require.NotPanics(t, func() { got, err = octave.FromNumber(tc.num, tc.n) })

			if tc.err {
				assertCause(t, err, causeNumberOutOfRange, "FromNumber")
				assert.Equal(t, octave.Octave{}, got, "value on error")
				assertPanicsCause(t, causeNumberOutOfRange, func() { octave.MustFromNumber(tc.num, tc.n) }, "MustFromNumber")

				return
			}

			require.NoError(t, err)
			assert.Equal(t, oct(tc.aspn), got)
			assert.Equal(t, tc.aspn, got.MustASPN())
			assert.Equal(t, tc.num, got.MustNumber(tc.n), "number back")
		})
	}
}

// K-23, G-3: the asymmetry with number.Convert, a property of octave/number
// next to the refusal of FromNumber: Convert translates MAX from Middle C = C3
// to Franco-Belgian and MIN from Middle C = C5 to Mexican, while the octaves
// with these numbers are not representable; Convert of MIN from C5 to C3 is
// itself out of range.
func TestConvertAsymmetry(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		x        int
		from, to number.Numbering
	}{
		{name: "MAX from C3 to FB", x: maxInt, from: c3(), to: fb()},
		{name: "MIN from C5 to Mex", x: minInt, from: c5(), to: mex()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := number.Convert(tc.x, tc.from, tc.to)
			require.NoError(t, err, "number.Convert")
			assert.Equal(t, tc.x, got, "number.Convert")

			_, err = octave.FromNumber(tc.x, tc.from)
			assertCause(t, err, causeNumberOutOfRange, "FromNumber")
		})
	}

	t.Run("MIN from C5 to C3", func(t *testing.T) {
		t.Parallel()

		got, err := number.Convert(minInt, c5(), c3())
		require.ErrorIs(t, err, number.ErrOutOfRange)
		assert.Zero(t, got)
	})
}

// K-31: the octaves created by number and by name in the usual range are
// valid and creation never panics, also on erroneous input.
func TestCreationNeverPanics(t *testing.T) {
	t.Parallel()

	nums := append(span(-20, 20), minInt, minInt+1, maxInt-1, maxInt)
	numberings := append(criteriaNumberings(), numberingCase{name: "undefined", n: number.Numbering{}})
	for _, nc := range numberings {
		for _, x := range nums {
			assert.NotPanics(t, func() { _, _ = octave.FromNumber(x, nc.n) }, "FromNumber(%d, %s)", x, nc.name)
		}
	}

	catalogs := append(builtinCatalogs(), emptyCatalogs()...)
	names := slices.Concat(
		[]string{"", "Первая", "x"},
		slices.Collect(maps.Values(russianNames())),
		slices.Collect(maps.Values(englishNames())),
		slices.Collect(maps.Values(germanNames())),
	)
	for _, cc := range catalogs {
		for _, s := range names {
			assert.NotPanics(t, func() { _, _ = octave.FromName(s, cc.c) }, "FromName(%q, %s)", s, cc.name)
		}
	}
}
