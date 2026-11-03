package naming_test

import (
	"maps"
	"math"
	"testing"

	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-6: a name that is not in the catalog — another spelling, case,
// form, a name of another catalog or the empty string — gives exactly 0 and
// ErrUnknownName (not ErrDuplicateName); MustASPN panics with ErrUnknownName.
// The returned 0 is checked as a contractual value, not used as an answer.
func TestUnknownName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		catalog func() *naming.Catalog
		input   string
	}{
		{name: "Russian empty", catalog: naming.Russian, input: ""},
		{name: "Russian capitalized", catalog: naming.Russian, input: "Первая"},
		{name: "Russian leading space", catalog: naming.Russian, input: " первая"},
		{name: "Russian trailing space", catalog: naming.Russian, input: "первая "},
		{name: "Russian е instead of ё", catalog: naming.Russian, input: "четвертая"},
		{name: "Russian decomposed ё", catalog: naming.Russian, input: "четве\u0308ртая"},
		{name: "English empty", catalog: naming.English, input: ""},
		{name: "English capitalized", catalog: naming.English, input: "Great"},
		{name: "English one-lined", catalog: naming.English, input: "one-lined"},
		{name: "English sub-contra", catalog: naming.English, input: "sub-contra"},
		{name: "English subcontraoctave", catalog: naming.English, input: "subcontraoctave"},
		{name: "English five-line", catalog: naming.English, input: "five-line"},
		{name: "English Russian name", catalog: naming.English, input: "первая"},
		{name: "German empty", catalog: naming.German, input: ""},
		{name: "German capitalized", catalog: naming.German, input: "Gro\u00dfe"},
		{name: "German with noun", catalog: naming.German, input: "gro\u00dfe Oktave"},
		{name: "German hyphenated", catalog: naming.German, input: "Subkontra-Oktave"},
		{name: "German fünfgestrichene", catalog: naming.German, input: "f\u00fcnfgestrichene"},
		{name: "German sechsgestrichene", catalog: naming.German, input: "sechsgestrichene"},
		{name: "German ss instead of ß", catalog: naming.German, input: "grosse"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.catalog()
			require.NotNil(t, c)

			got, err := c.ASPN(tc.input)
			require.ErrorIs(t, err, naming.ErrUnknownName)
			assert.NotErrorIs(t, err, naming.ErrDuplicateName)
			assert.Equal(t, 0, got)

			assertPanicsWithErrorIs(t, naming.ErrUnknownName, naming.ErrDuplicateName,
				func() { c.MustASPN(tc.input) })
		})
	}
}

// K-7: a custom catalog gives its names, the empty name for other
// octaves and finds its names. A pair with the empty name means "no name"
// and is not found.
func TestNewCustomCatalog(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		name  string
		pairs func() map[int]string
	}{
		{name: "K-7 pairs", pairs: customPairs},
		{name: "K-7 pairs and an empty name", pairs: func() map[int]string {
			pairs := customPairs()
			pairs[0] = ""

			return pairs
		}},
	}

	for _, in := range inputs {
		t.Run(in.name, func(t *testing.T) {
			t.Parallel()

			c, err := naming.New(in.pairs())
			require.NoError(t, err)
			require.NotNil(t, c)

			names := []struct {
				aspn int
				want string
			}{
				{aspn: 4, want: "middle"},
				{aspn: -3, want: "deep"},
				{aspn: 12, want: "high"},
				{aspn: 0, want: ""},
				{aspn: 5, want: ""},
			}
			for _, tc := range names {
				assert.Equal(t, tc.want, c.Name(tc.aspn), "ASPN %d", tc.aspn)
			}

			lookups := []struct {
				input string
				want  int
			}{
				{input: "deep", want: -3},
				{input: "middle", want: 4},
				{input: "high", want: 12},
			}
			for _, tc := range lookups {
				got, err := c.ASPN(tc.input)
				if assert.NoError(t, err, "name %q", tc.input) {
					assert.Equal(t, tc.want, got, "name %q", tc.input)
				}
			}

			got, err := c.ASPN("")
			assert.ErrorIs(t, err, naming.ErrUnknownName)
			assert.Equal(t, 0, got)
		})
	}
}

// K-7: names are stored and found exactly as given: Cyrillic, spaces,
// spaces only, letters ü and ß, and even invalid UTF-8. No normalization,
// trimming or case folding.
func TestNewExactStrings(t *testing.T) {
	t.Parallel()

	pairs := func() map[int]string {
		return map[int]string{
			3:  "малая октава",
			5:  " x ",
			6:  "\u00fc",
			7:  "\u00df",
			8:  "\xff",
			9:  "   ",
			10: "Первая",
		}
	}

	t.Run("names and lookup", func(t *testing.T) {
		t.Parallel()

		c, err := naming.New(pairs())
		require.NoError(t, err)
		require.NotNil(t, c)

		for aspn, name := range pairs() {
			assert.Equal(t, name, c.Name(aspn), "ASPN %d", aspn)

			got, err := c.ASPN(name)
			if assert.NoError(t, err, "name %q", name) {
				assert.Equal(t, aspn, got, "name %q", name)
			}
		}
	})

	cases := []struct {
		name  string
		input string
	}{
		{name: "first word only", input: "малая"},
		{name: "double space", input: "малая  октава"},
		{name: "trimmed", input: "x"},
		{name: "leading space only", input: " x"},
		{name: "trailing space only", input: "x "},
		{name: "decomposed ü", input: "u\u0308"},
		{name: "ss instead of ß", input: "ss"},
		{name: "capital ẞ", input: "\u1e9e"},
		{name: "other invalid byte", input: "\xfe"},
		{name: "one space", input: " "},
		{name: "four spaces", input: "    "},
		{name: "lowercase", input: "первая"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := naming.New(pairs())
			require.NoError(t, err)
			require.NotNil(t, c)

			got, err := c.ASPN(tc.input)
			assert.ErrorIs(t, err, naming.ErrUnknownName)
			assert.Equal(t, 0, got)
		})
	}
}

// Names at the extreme int numbers are given and found exactly; neighboring
// numbers have no name.
func TestNewExtremeNumbers(t *testing.T) {
	t.Parallel()

	c, err := naming.New(map[int]string{math.MinInt: "lo", math.MaxInt: "hi"})
	require.NoError(t, err)
	require.NotNil(t, c)

	assert.Equal(t, "lo", c.Name(math.MinInt))
	assert.Equal(t, "hi", c.Name(math.MaxInt))
	assert.Empty(t, c.Name(math.MinInt+1))
	assert.Empty(t, c.Name(math.MaxInt-1))
	assert.Empty(t, c.Name(0))

	lo, err := c.ASPN("lo")
	require.NoError(t, err)
	assert.Equal(t, math.MinInt, lo)

	hi, err := c.ASPN("hi")
	require.NoError(t, err)
	assert.Equal(t, math.MaxInt, hi)
}

// K-8: the same non-empty name for two different octaves is a
// creation error: New returns nil and ErrDuplicateName (not ErrUnknownName),
// MustNew panics with it, and the input map is not changed (K-12).
func TestNewDuplicateName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		pairs func() map[int]string
	}{
		{name: "two equal names", pairs: func() map[int]string { return map[int]string{1: "a", 2: "a"} }},
		{name: "repeat among others", pairs: func() map[int]string { return map[int]string{1: "a", 2: "b", 3: "a"} }},
		{name: "Cyrillic name", pairs: func() map[int]string { return map[int]string{4: "первая", 5: "первая"} }},
		{name: "spaces only", pairs: func() map[int]string { return map[int]string{1: " ", 2: " "} }},
		{name: "extreme numbers", pairs: func() map[int]string { return map[int]string{math.MinInt: "x", math.MaxInt: "x"} }},
		{name: "with empty names", pairs: func() map[int]string { return map[int]string{1: "", 2: "a", 3: "", 4: "a"} }},
		{name: "zero and a greater number", pairs: func() map[int]string { return map[int]string{0: "x", 1: "x"} }},
		{name: "a smaller number and zero", pairs: func() map[int]string { return map[int]string{-1: "x", 0: "x"} }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			in := tc.pairs()
			c, err := naming.New(in)
			require.ErrorIs(t, err, naming.ErrDuplicateName)
			assert.NotErrorIs(t, err, naming.ErrUnknownName)
			assert.Nil(t, c)
			assert.Equal(t, tc.pairs(), in, "the input map is not changed")

			must := tc.pairs()
			assertPanicsWithErrorIs(t, naming.ErrDuplicateName, naming.ErrUnknownName,
				func() { naming.MustNew(must) })
			assert.Equal(t, tc.pairs(), must, "the input map is not changed")
		})
	}
}

// K-8: a repeated name is a creation error also when one of its octaves is
// ASPN 0. The order in which New examines the pairs is not observable, so a
// fault that depends on the order (for example, treating number 0 as "no
// octave yet") is exposed in two ways: the partner of 0 lies on either side of
// it, which covers examination in ascending and in descending order, and every
// input is checked repeatedly with a new map, which covers randomized map
// iteration. A correct implementation passes every attempt. This black-box test
// cannot cover every conceivable fixed order.
func TestNewDuplicateNameAtZero(t *testing.T) {
	t.Parallel()

	const attempts = 100

	cases := []struct {
		name  string
		pairs func() map[int]string
	}{
		{name: "zero and one", pairs: func() map[int]string { return map[int]string{0: "x", 1: "x"} }},
		{name: "minus one and zero", pairs: func() map[int]string { return map[int]string{-1: "x", 0: "x"} }},
		{name: "zero and MinInt", pairs: func() map[int]string { return map[int]string{0: "x", math.MinInt: "x"} }},
		{name: "zero and MaxInt", pairs: func() map[int]string { return map[int]string{0: "x", math.MaxInt: "x"} }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for attempt := range attempts {
				c, err := naming.New(tc.pairs())
				if !assert.ErrorIs(t, err, naming.ErrDuplicateName, "New, attempt %d", attempt) ||
					!assert.Nil(t, c, "New, attempt %d", attempt) {
					break
				}
			}

			for attempt := range attempts {
				if !assertPanicsWithErrorIs(t, naming.ErrDuplicateName, naming.ErrUnknownName,
					func() { naming.MustNew(tc.pairs()) }, "MustNew, attempt %d", attempt) {
					break
				}
			}
		})
	}
}

// K-8: names that differ in any byte are not a repeat; a name
// of a built-in catalog may be used in a custom catalog.
func TestNewDistinctNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		pairs func() map[int]string
	}{
		{name: "case differs", pairs: func() map[int]string { return map[int]string{1: "a", 2: "A"} }},
		{name: "leading space differs", pairs: func() map[int]string { return map[int]string{1: " a", 2: "a"} }},
		{name: "trailing space differs", pairs: func() map[int]string { return map[int]string{1: "a", 2: "a "} }},
		{name: "built-in name", pairs: func() map[int]string { return map[int]string{4: "первая"} }},
		{name: "е and ё", pairs: func() map[int]string { return map[int]string{6: "четвертая", 7: "четв\u0451ртая"} }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := naming.New(tc.pairs())
			require.NoError(t, err)
			require.NotNil(t, c)

			for aspn, name := range tc.pairs() {
				assert.Equal(t, name, c.Name(aspn), "ASPN %d", aspn)

				got, err := c.ASPN(name)
				if assert.NoError(t, err, "name %q", name) {
					assert.Equal(t, aspn, got, "name %q", name)
				}
			}
		})
	}
}

// K-10: a catalog without names is created without an
// error from a nil map, an empty map and pairs with empty names only (two
// empty names are not a repeat). It gives the empty name for every octave,
// and every lookup, including the empty string, gives 0 and ErrUnknownName.
func TestNewWithoutNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		pairs func() map[int]string
	}{
		{name: "nil map", pairs: func() map[int]string { return nil }},
		{name: "empty map", pairs: func() map[int]string { return map[int]string{} }},
		{name: "two empty names", pairs: func() map[int]string { return map[int]string{1: "", 2: ""} }},
		{name: "one empty name", pairs: func() map[int]string { return map[int]string{0: ""} }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := naming.New(tc.pairs())
			require.NoError(t, err)
			require.NotNil(t, c)

			var must *naming.Catalog
			require.NotPanics(t, func() { must = naming.MustNew(tc.pairs()) })
			require.NotNil(t, must)

			for _, n := range append([]int{-1, 0, 1, 2, 4, 100}, edgeNumbers()...) {
				assert.Empty(t, c.Name(n), "ASPN %d", n)
				assert.Empty(t, must.Name(n), "ASPN %d", n)
			}

			for _, s := range []string{"", " ", "a", "первая", "subcontra"} {
				got, err := c.ASPN(s)
				assert.ErrorIs(t, err, naming.ErrUnknownName, "name %q", s)
				assert.NotErrorIs(t, err, naming.ErrDuplicateName, "name %q", s)
				assert.Equal(t, 0, got, "name %q", s)
			}
		})
	}
}

// A successful New always returns a new non-nil catalog;
// two calls with the same map give different pointers equal by Equal.
func TestNewReturnsNewCatalog(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		pairs map[int]string
	}{
		{name: "K-7 pairs", pairs: customPairs()},
		{name: "nil map", pairs: nil},
		{name: "empty name only", pairs: map[int]string{1: ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			first, err := naming.New(tc.pairs)
			require.NoError(t, err)
			second, err := naming.New(tc.pairs)
			require.NoError(t, err)

			require.NotNil(t, first)
			require.NotNil(t, second)
			assert.NotSame(t, first, second)
			assert.True(t, first.Equal(second))

			third, fourth := naming.MustNew(tc.pairs), naming.MustNew(tc.pairs)
			require.NotNil(t, third)
			require.NotNil(t, fourth)
			assert.NotSame(t, third, fourth)
			assert.True(t, third.Equal(first))
		})
	}
}

// K-12: New and MustNew copy their input. Changing, deleting, adding or
// clearing entries of the map after the call does not change the catalog; a
// successful call of either does not change the map. Every change is made on
// an input with non-empty names only and on an input with an empty name, with
// a separate map for each constructor, and both directions of every pair are
// checked afterwards.
func TestNewCopiesInput(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		name  string
		pairs func() map[int]string
	}{
		{name: "non-empty names only", pairs: customPairs},
		{name: "with an empty name", pairs: func() map[int]string {
			pairs := customPairs()
			pairs[0] = ""

			return pairs
		}},
	}

	constructors := []struct {
		name string
		make func(map[int]string) (*naming.Catalog, error)
	}{
		{name: "New", make: naming.New},
		{name: "MustNew", make: func(m map[int]string) (*naming.Catalog, error) { return naming.MustNew(m), nil }},
	}

	changes := []struct {
		name   string
		change func(map[int]string)
	}{
		{name: "replace a name", change: func(m map[int]string) { m[4] = "changed" }},
		{name: "delete a pair", change: func(m map[int]string) { delete(m, -3) }},
		{name: "add a pair", change: func(m map[int]string) { m[1] = "added" }},
		{name: "name an unnamed octave", change: func(m map[int]string) { m[0] = "zero" }},
		{name: "add a repeated name", change: func(m map[int]string) { m[7] = "high" }},
		{name: "clear", change: func(m map[int]string) { clear(m) }},
	}

	for _, input := range inputs {
		for _, ctor := range constructors {
			for _, tc := range changes {
				t.Run(input.name+"/"+ctor.name+"/"+tc.name, func(t *testing.T) {
					t.Parallel()

					in := input.pairs()
					before := maps.Clone(in)

					var (
						c   *naming.Catalog
						err error
					)
					require.NotPanics(t, func() { c, err = ctor.make(in) })
					require.NoError(t, err)
					require.NotNil(t, c)
					assert.Equal(t, before, in, "a successful call does not change its input")

					tc.change(in)

					names := []struct {
						aspn int
						want string
					}{
						{aspn: 4, want: "middle"},
						{aspn: -3, want: "deep"},
						{aspn: 12, want: "high"},
					}
					for _, pair := range names {
						assert.Equal(t, pair.want, c.Name(pair.aspn), "ASPN %d", pair.aspn)

						aspn, err := c.ASPN(pair.want)
						if assert.NoError(t, err, "name %q", pair.want) {
							assert.Equal(t, pair.aspn, aspn, "name %q", pair.want)
						}
					}

					for _, n := range []int{0, 1, 7} {
						assert.Empty(t, c.Name(n), "ASPN %d", n)
					}

					for _, s := range []string{"changed", "added", "zero"} {
						_, err := c.ASPN(s)
						assert.ErrorIs(t, err, naming.ErrUnknownName, "name %q", s)
					}

					assert.True(t, c.Equal(naming.MustNew(customPairs())))
				})
			}
		}
	}
}

// K-12: overwriting the object returned by a
// built-in function changes only that object; later calls and other users
// get the full catalog of K-1...K-3.
func TestBuiltinOverwrite(t *testing.T) {
	t.Parallel()

	for _, b := range builtins() {
		t.Run(b.name, func(t *testing.T) {
			t.Parallel()

			kept := b.catalog()
			r := b.catalog()
			require.NotNil(t, kept)
			require.NotNil(t, r)

			*r = naming.Catalog{}

			assert.True(t, r.Equal(nil), "the overwritten object is a catalog without names")
			for n := range b.oracle() {
				assert.Empty(t, r.Name(n), "overwritten, ASPN %d", n)
			}

			fresh := b.catalog()
			require.NotNil(t, fresh)
			for n, want := range b.oracle() {
				assert.Equal(t, want, fresh.Name(n), "new call, ASPN %d", n)
				assert.Equal(t, want, kept.Name(n), "earlier call, ASPN %d", n)
			}
			assert.True(t, fresh.Equal(naming.MustNew(b.oracle())))
		})
	}
}

// K-12: a copy of a catalog value is equal to
// the original and behaves the same; overwriting the original does not change
// the copy.
func TestCopyValue(t *testing.T) {
	t.Parallel()

	c := naming.MustNew(customPairs())
	require.NotNil(t, c)

	cp := *c
	assert.True(t, cp.Equal(c))
	assert.True(t, c.Equal(&cp))
	assert.Equal(t, "middle", cp.Name(4))

	*c = naming.Catalog{}

	assert.True(t, c.Equal(nil))
	assert.Empty(t, c.Name(4))

	assert.Equal(t, "middle", cp.Name(4))
	assert.Equal(t, "deep", cp.Name(-3))
	aspn, err := cp.ASPN("high")
	require.NoError(t, err)
	assert.Equal(t, 12, aspn)
	assert.True(t, cp.Equal(naming.MustNew(customPairs())))
}

// Consistency for arbitrary input: a catalog from two pairs either reports a
// repeated non-empty name, or names each named octave, finds each name and
// leaves the input unchanged. The expectation is computed from the input
// pairs, not from the package. go test runs only the seed corpus.
func FuzzNewTwoPairs(f *testing.F) {
	f.Add(1, "a", 2, "b")
	f.Add(1, "a", 2, "a")
	f.Add(1, "a", 1, "b")
	f.Add(1, "", 2, "")
	f.Add(4, "первая", 5, "")
	f.Add(math.MinInt, "lo", math.MaxInt, "hi")
	f.Add(0, " x ", -1, "x")
	f.Add(3, "\xff", 4, "\xfe")
	f.Add(0, "x", 1, "x")
	f.Add(-1, "x", 0, "x")

	f.Fuzz(func(t *testing.T, k1 int, n1 string, k2 int, n2 string) {
		in := map[int]string{}
		in[k1] = n1
		in[k2] = n2
		before := maps.Clone(in)

		named := map[int]string{}
		for k, n := range in {
			if n != "" {
				named[k] = n
			}
		}
		repeated := len(named) == 2 && n1 == n2

		c, err := naming.New(in)
		require.Equal(t, before, in, "the input map is not changed")

		if repeated {
			require.ErrorIs(t, err, naming.ErrDuplicateName)
			require.Nil(t, c)

			return
		}

		require.NoError(t, err)
		require.NotNil(t, c)

		for k, n := range in {
			require.Equal(t, n, c.Name(k), "ASPN %d", k)
		}
		for k, n := range named {
			got, err := c.ASPN(n)
			require.NoError(t, err, "name %q", n)
			require.Equal(t, k, got, "name %q", n)
		}

		_, err = c.ASPN("")
		require.ErrorIs(t, err, naming.ErrUnknownName)
	})
}
