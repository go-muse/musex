package octave_test

import (
	"fmt"
	"maps"
	"testing"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-16, C-16: an own numbering defined by an octave equals number.New with
// the ASPN number of that octave; it has no label.
func TestNewNumbering(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		o       octave.Octave
		num     int
		skipped []int
		want    number.Numbering
	}{
		{name: "первая is 1", o: octave.MustFromName("первая", naming.Russian()), num: 1, want: number.MustNew(4, 1)},
		{name: "O(4) is 3, skip 0", o: oct(4), num: 3, skipped: []int{0}, want: fb()},
		{name: "O(4) is 3", o: oct(4), num: 3, want: c3()},
		{name: "O(3, FB) is 4", o: octave.MustFromNumber(3, fb()), num: 4, want: number.ASPN()},
		{name: "O(4) is 4, skip 0 1", o: oct(4), num: 4, skipped: []int{0, 1}, want: number.MustNew(4, 4, 0, 1)},
		{name: "O(4) is 3, empty skip", o: oct(4), num: 3, skipped: []int{}, want: c3()},
		{name: "O(MIN) is MAX", o: oct(minInt), num: maxInt, want: number.MustNew(minInt, maxInt)},
		{name: "O(MAX) is MIN", o: oct(maxInt), num: minInt, want: number.MustNew(maxInt, minInt)},
		{name: "O(-1000) is 0, skip MIN MAX", o: oct(-1000), num: 0, skipped: []int{minInt, maxInt}, want: number.MustNew(-1000, 0, minInt, maxInt)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got number.Numbering
				err error
			)
			require.NotPanics(t, func() { got, err = octave.NewNumbering(tc.o, tc.num, tc.skipped...) })
			require.NoError(t, err)
			assert.True(t, got.IsDefined(), "IsDefined")
			assert.True(t, got.Equal(tc.want), "Equal: got %v, want %v", got, tc.want)
			assert.Empty(t, got.Label(), "Label")
			assert.Equal(t, tc.num, tc.o.MustNumber(got), "number of the defining octave")

			must := octave.MustNewNumbering(tc.o, tc.num, tc.skipped...)
			assert.True(t, must.Equal(tc.want), "MustNewNumbering: got %v, want %v", must, tc.want)
			assert.Empty(t, must.Label(), "MustNewNumbering Label")
		})
	}

	t.Run("number of O(3) from первая is 1", func(t *testing.T) {
		t.Parallel()

		n := octave.MustNewNumbering(octave.MustFromName("первая", naming.Russian()), 1)
		got, err := oct(3).Number(n)
		require.NoError(t, err)
		assert.Equal(t, 0, got)
	})
}

// K-16, E-6: an invalid definition gives number.ErrInvalidDefinition and the
// undefined numbering; the Must variant panics with the same cause.
func TestNewNumberingInvalidDefinition(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		o       octave.Octave
		num     int
		skipped []int
	}{
		{name: "O(4) is 0, skip 0", o: oct(4), num: 0, skipped: []int{0}},
		{name: "O(4) is 3, skip 0 0", o: oct(4), num: 3, skipped: []int{0, 0}},
		{name: "O(MIN) is MAX, skip MAX", o: oct(minInt), num: maxInt, skipped: []int{maxInt}},
		{name: "O(4) is 3, skip 1 2 1", o: oct(4), num: 3, skipped: []int{1, 2, 1}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				got number.Numbering
				err error
			)
			require.NotPanics(t, func() { got, err = octave.NewNumbering(tc.o, tc.num, tc.skipped...) })
			assertCause(t, err, causeInvalidDefinition, "NewNumbering")
			assert.False(t, got.IsDefined(), "value on error is the zero Numbering")

			assertPanicsCause(t, causeInvalidDefinition, func() {
				octave.MustNewNumbering(tc.o, tc.num, tc.skipped...)
			}, "MustNewNumbering")
		})
	}
}

// K-17, C-17: an own catalog defined by octaves equals naming.New with the
// ASPN numbers of the octaves; names and lookups are exact.
func TestNewCatalog(t *testing.T) {
	t.Parallel()

	t.Run("criteria pairs", func(t *testing.T) {
		t.Parallel()

		pairs := map[octave.Octave]string{oct(4): "middle", oct(-3): "deep", oct(12): "high"}
		want := naming.MustNew(map[int]string{4: "middle", -3: "deep", 12: "high"})

		var (
			c   *naming.Catalog
			err error
		)
		require.NotPanics(t, func() { c, err = octave.NewCatalog(pairs) })
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.True(t, c.Equal(want), "Equal to naming.MustNew")
		assert.True(t, octave.MustNewCatalog(pairs).Equal(want), "MustNewCatalog Equal to naming.MustNew")

		assert.Equal(t, "deep", oct(-3).Name(c))
		assert.Equal(t, "middle", oct(4).Name(c))
		assert.Equal(t, "high", oct(12).Name(c))
		assert.Empty(t, oct(5).Name(c))

		got, err := octave.FromName("deep", c)
		require.NoError(t, err)
		assert.Equal(t, oct(-3), got)
		assert.Equal(t, oct(12), octave.MustFromName("high", c))
	})

	t.Run("keys obtained in other ways", func(t *testing.T) {
		t.Parallel()

		pairs := map[octave.Octave]string{
			octave.MustFromNumber(3, fb()):                 "middle",
			octave.MustFromName("малая", naming.Russian()): "small",
		}
		c, err := octave.NewCatalog(pairs)
		require.NoError(t, err)
		assert.True(t, c.Equal(naming.MustNew(map[int]string{4: "middle", 3: "small"})))
	})

	t.Run("limits", func(t *testing.T) {
		t.Parallel()

		c, err := octave.NewCatalog(map[octave.Octave]string{oct(minInt): "lo", oct(maxInt): "hi"})
		require.NoError(t, err)
		assert.True(t, c.Equal(naming.MustNew(map[int]string{minInt: "lo", maxInt: "hi"})))
		assert.Equal(t, "lo", oct(minInt).Name(c))
		assert.Equal(t, "hi", oct(maxInt).Name(c))
		assert.Equal(t, oct(minInt), octave.MustFromName("lo", c))
		assert.Equal(t, oct(maxInt), octave.MustFromName("hi", c))
		assert.Empty(t, oct(minInt+1).Name(c))
	})
}

// K-17, C-17, G-5: empty names mean "no name"; nil and empty maps give a
// non-nil catalog without names.
func TestNewCatalogWithoutNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		pairs map[octave.Octave]string
	}{
		{name: "nil", pairs: nil},
		{name: "empty map", pairs: map[octave.Octave]string{}},
		{name: "one empty name", pairs: map[octave.Octave]string{oct(1): ""}},
		{name: "two empty names", pairs: map[octave.Octave]string{oct(1): "", oct(2): ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := octave.NewCatalog(tc.pairs)
			require.NoError(t, err)
			require.NotNil(t, c)
			assert.True(t, c.Equal(naming.MustNew(nil)), "a catalog without names")
			assert.Empty(t, oct(1).Name(c))
			_, err = octave.FromName("", c)
			assertCause(t, err, causeUnknownName, "FromName(\"\")")

			must := octave.MustNewCatalog(tc.pairs)
			require.NotNil(t, must)
			assert.True(t, must.Equal(naming.MustNew(nil)), "MustNewCatalog: a catalog without names")
		})
	}

	t.Run("empty name next to a name", func(t *testing.T) {
		t.Parallel()

		c, err := octave.NewCatalog(map[octave.Octave]string{oct(1): "", oct(2): "b"})
		require.NoError(t, err)
		assert.True(t, c.Equal(naming.MustNew(map[int]string{2: "b"})))
		assert.Empty(t, oct(1).Name(c))
		assert.Equal(t, "b", oct(2).Name(c))
	})
}

// K-17, E-7: one non-empty name for two different octaves gives
// naming.ErrDuplicateName and nil; the Must variant panics.
func TestNewCatalogDuplicateName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		pairs map[octave.Octave]string
	}{
		{name: "O(1) a, O(2) a", pairs: map[octave.Octave]string{oct(1): "a", oct(2): "a"}},
		{name: "O(MIN) x, O(MAX) x", pairs: map[octave.Octave]string{oct(minInt): "x", oct(maxInt): "x"}},
		{name: "with other names", pairs: map[octave.Octave]string{oct(1): "a", oct(2): "b", oct(3): "a", oct(4): ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				c   *naming.Catalog
				err error
			)
			require.NotPanics(t, func() { c, err = octave.NewCatalog(tc.pairs) })
			assertCause(t, err, causeDuplicateName, "NewCatalog")
			assert.Nil(t, c, "value on error")

			assertPanicsCause(t, causeDuplicateName, func() { octave.MustNewCatalog(tc.pairs) }, "MustNewCatalog")
		})
	}
}

// K-17, E-7, KR-OC12: "one octave twice" cannot be passed: the input is a map
// keyed by Octave, and records of one octave obtained in different ways are
// one key before the call. The later assignment wins in the caller's map.
func TestNewCatalogOneOctaveIsOneKey(t *testing.T) {
	t.Parallel()

	pairs := map[octave.Octave]string{}
	pairs[oct(2)] = "a"
	pairs[octave.MustFromNumber(1, fb())] = "b"
	require.Len(t, pairs, 1, "O(2) and O(1, FB) are one key")

	c, err := octave.NewCatalog(pairs)
	require.NoError(t, err)
	assert.Equal(t, "b", oct(2).Name(c))
	assert.True(t, c.Equal(naming.MustNew(map[int]string{2: "b"})))
}

// K-27, G-10: the skipped numbers passed to NewNumbering are not retained:
// changing the slice after the call does not change the numbering, and the
// call does not change the slice.
func TestNewNumberingOwnership(t *testing.T) {
	t.Parallel()

	t.Run("NewNumbering", func(t *testing.T) {
		t.Parallel()

		skips := []int{0, 7}
		n, err := octave.NewNumbering(oct(4), 3, skips...)
		require.NoError(t, err)
		assert.Equal(t, []int{0, 7}, skips, "the call does not change the slice")

		skips[0] = 1
		skips[1] = 3
		skips = append(skips[:1], 5, 6)
		assert.Equal(t, []int{1, 5, 6}, skips)

		assert.True(t, n.Equal(number.MustNew(4, 3, 0, 7)), "numbering after changes: %v", n)
		assert.Equal(t, -1, oct(1).MustNumber(n), "number of O(1) after changes: 0 is still skipped")
		assert.Equal(t, 8, oct(8).MustNumber(n), "number of O(8) after changes: 7 is still skipped")
	})

	t.Run("MustNewNumbering", func(t *testing.T) {
		t.Parallel()

		skips := []int{0}
		n := octave.MustNewNumbering(oct(4), 3, skips...)
		skips[0] = 2
		assert.True(t, n.Equal(fb()), "numbering after changes: %v", n)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		skips := []int{3, 0}
		_, err := octave.NewNumbering(oct(4), 3, skips...)
		assertCause(t, err, causeInvalidDefinition, "NewNumbering")
		assert.Equal(t, []int{3, 0}, skips, "the call does not change the slice on error")
	})
}

// K-27, G-10: the map passed to NewCatalog is neither changed nor retained:
// replacing, deleting and adding pairs after the call does not change the
// catalog; the map is unchanged after a successful call and after an error.
func TestNewCatalogOwnership(t *testing.T) {
	t.Parallel()

	want := naming.MustNew(map[int]string{4: "middle", -3: "deep", 12: "high"})

	for _, must := range []bool{false, true} {
		t.Run(fmt.Sprintf("must=%t", must), func(t *testing.T) {
			t.Parallel()

			names := map[octave.Octave]string{oct(4): "middle", oct(-3): "deep", oct(12): "high"}
			orig := maps.Clone(names)

			var c *naming.Catalog
			if must {
				c = octave.MustNewCatalog(names)
			} else {
				var err error
				c, err = octave.NewCatalog(names)
				require.NoError(t, err)
			}
			assert.True(t, maps.Equal(orig, names), "the call does not change the map")

			names[oct(4)] = "changed"
			delete(names, oct(-3))
			names[oct(0)] = "added"

			assert.True(t, c.Equal(want), "catalog after changes")
			assert.Equal(t, "middle", oct(4).Name(c))
			assert.Equal(t, "deep", oct(-3).Name(c))
			assert.Empty(t, oct(0).Name(c))
			_, err := octave.FromName("changed", c)
			assertCause(t, err, causeUnknownName, "FromName(changed)")
		})
	}

	errorCases := []struct {
		name  string
		pairs map[octave.Octave]string
		want  cause
	}{
		{name: "duplicate name", pairs: map[octave.Octave]string{oct(1): "a", oct(2): "a"}, want: causeDuplicateName},
		{name: "invalid key", pairs: map[octave.Octave]string{{}: "z", oct(1): "a"}, want: causeInvalid},
	}
	for _, tc := range errorCases {
		t.Run("error "+tc.name, func(t *testing.T) {
			t.Parallel()

			orig := maps.Clone(tc.pairs)
			_, err := octave.NewCatalog(tc.pairs)
			assertCause(t, err, tc.want, "NewCatalog")
			assert.True(t, maps.Equal(orig, tc.pairs), "the call does not change the map on error")
		})
	}
}

// K-17, NF-3: an empty name means "the octave has no name in the catalog"; it
// does not allow NewCatalog or MustNewCatalog to delete the pair from the map
// of the caller. The whole map (keys, values including the empty ones, and
// length) is the same after the call, after a successful one and after an
// error.
func TestNewCatalogOwnershipEmptyNames(t *testing.T) {
	t.Parallel()

	successCases := []struct {
		name    string
		pairs   func() map[octave.Octave]string
		want    map[int]string
		unnamed []int
	}{
		{
			name:    "empty beside named",
			pairs:   func() map[octave.Octave]string { return map[octave.Octave]string{oct(1): "", oct(2): "named"} },
			want:    map[int]string{2: "named"},
			unnamed: []int{1},
		},
		{
			name:    "only empty",
			pairs:   func() map[octave.Octave]string { return map[octave.Octave]string{oct(1): "", oct(2): ""} },
			want:    map[int]string{},
			unnamed: []int{1, 2},
		},
		{
			name: "empty at extreme numbers",
			pairs: func() map[octave.Octave]string {
				return map[octave.Octave]string{oct(minInt): "", oct(0): "named", oct(maxInt): ""}
			},
			want:    map[int]string{0: "named"},
			unnamed: []int{minInt, maxInt},
		},
	}

	for _, tc := range successCases {
		for _, must := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/must=%t", tc.name, must), func(t *testing.T) {
				t.Parallel()

				names := tc.pairs()
				orig := maps.Clone(names)
				require.Equal(t, len(tc.want)+len(tc.unnamed), len(orig), "test data: every pair is either named or empty")

				var c *naming.Catalog
				if must {
					c = octave.MustNewCatalog(names)
				} else {
					var err error
					c, err = octave.NewCatalog(names)
					require.NoError(t, err)
				}

				assert.Equal(t, orig, names, "the call does not change the map")
				assert.Len(t, names, len(orig), "the call does not delete pairs from the map")
				for _, a := range tc.unnamed {
					v, ok := names[oct(a)]
					assert.True(t, ok, "the pair of O(%d) with the empty name is still in the map", a)
					assert.Empty(t, v, "the name of O(%d) in the map", a)
					assert.Empty(t, oct(a).Name(c), "O(%d) has no name in the catalog", a)
				}

				wantCatalog := naming.MustNew(tc.want)
				assert.True(t, c.Equal(wantCatalog), "catalog: got %v, want %v", c, wantCatalog)
			})
		}
	}

	t.Run("error duplicate name beside empty", func(t *testing.T) {
		t.Parallel()

		pairs := func() map[octave.Octave]string {
			return map[octave.Octave]string{oct(1): "", oct(2): "duplicate", oct(3): "duplicate"}
		}

		t.Run("NewCatalog", func(t *testing.T) {
			t.Parallel()

			names := pairs()
			orig := maps.Clone(names)

			c, err := octave.NewCatalog(names)
			assertCause(t, err, causeDuplicateName, "NewCatalog")
			assert.Nil(t, c, "no catalog on error")
			assert.Equal(t, orig, names, "the call does not change the map on error")
			v, ok := names[oct(1)]
			assert.True(t, ok, "the pair with the empty name is still in the map")
			assert.Empty(t, v, "the name of O(1) in the map")
		})

		t.Run("MustNewCatalog", func(t *testing.T) {
			t.Parallel()

			names := pairs()
			orig := maps.Clone(names)

			assertPanicsCause(t, causeDuplicateName, func() { octave.MustNewCatalog(names) }, "MustNewCatalog")
			assert.Equal(t, orig, names, "the call does not change the map on panic")
			v, ok := names[oct(1)]
			assert.True(t, ok, "the pair with the empty name is still in the map")
			assert.Empty(t, v, "the name of O(1) in the map")
		})
	})
}

// K-16, K-17: own numberings and catalogs created by octaves work for every
// octave in the usual range like their octave/number and octave/naming
// counterparts.
func TestOwnDefinitionsMatchSubpackages(t *testing.T) {
	t.Parallel()

	n := octave.MustNewNumbering(oct(2), -5, -3, 0)
	wantN := number.MustNew(2, -5, -3, 0)

	pairs := map[octave.Octave]string{oct(-7): "low", oct(0): "zero", oct(9): "nine"}
	c := octave.MustNewCatalog(pairs)
	wantC := naming.MustNew(map[int]string{-7: "low", 0: "zero", 9: "nine"})

	for _, a := range span(-20, 20) {
		assert.Equal(t, wantN.MustFromASPN(a), oct(a).MustNumber(n), "O(%d) in own numbering", a)
		assert.Equal(t, wantC.Name(a), oct(a).Name(c), "O(%d) in own catalog", a)
	}
}
