package naming_test

import (
	"math"
	"sync"
	"testing"

	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// aliasCatalog returns the Russian catalog extended with the aliases of the
// acceptance example: ASPN 4 also accepts "первая октава" and "FirstOctave".
func aliasCatalog() *naming.Catalog {
	return naming.Russian().MustWithAliases(map[int][]string{
		4: {"первая октава", "FirstOctave"},
	})
}

// WithAliases accepts additional exact input names for a named octave: all
// the names of the octave resolve to its ASPN number, while Name keeps
// returning the canonical name.
func TestWithAliasesLookup(t *testing.T) {
	t.Parallel()

	c := aliasCatalog()

	for _, name := range []string{"первая", "первая октава", "FirstOctave"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			aspn, err := c.ASPN(name)
			require.NoError(t, err)
			assert.Equal(t, 4, aspn)
			assert.Equal(t, 4, c.MustASPN(name))
		})
	}

	assert.Equal(t, "первая", c.Name(4))
	assert.Equal(t, "первая", c.Name(c.MustASPN("FirstOctave")))
	assert.Equal(t, []string{"FirstOctave", "первая октава"}, c.Aliases(4))
}

// An alias that already resolves to the same octave is a no-op: a repeated
// alias in the list, an alias the receiver already has, and the octave's own
// canonical name are accepted and not stored twice.
func TestWithAliasesIdempotent(t *testing.T) {
	t.Parallel()

	t.Run("repeated in the list", func(t *testing.T) {
		t.Parallel()

		c := naming.Russian().MustWithAliases(map[int][]string{4: {"x", "x"}})
		assert.Equal(t, []string{"x"}, c.Aliases(4))
		assert.Equal(t, 4, c.MustASPN("x"))
	})

	t.Run("own canonical name", func(t *testing.T) {
		t.Parallel()

		c := naming.Russian().MustWithAliases(map[int][]string{4: {"первая"}})
		assert.Empty(t, c.Aliases(4))
		assert.True(t, c.Equal(naming.Russian()))
		assert.Equal(t, 4, c.MustASPN("первая"))
	})

	t.Run("already in the receiver", func(t *testing.T) {
		t.Parallel()

		once := naming.Russian().MustWithAliases(map[int][]string{4: {"x"}})
		twice := once.MustWithAliases(map[int][]string{4: {"x"}})
		assert.Equal(t, []string{"x"}, twice.Aliases(4))
		assert.True(t, once.Equal(twice))
	})

	t.Run("same input twice", func(t *testing.T) {
		t.Parallel()

		input := map[int][]string{4: {"x", "y"}, 3: {"z"}}
		once := naming.Russian().MustWithAliases(input)
		twice := once.MustWithAliases(input)
		assert.True(t, once.Equal(twice))
	})
}

// Successive calls accumulate: aliases already in the receiver are kept.
func TestWithAliasesAccumulates(t *testing.T) {
	t.Parallel()

	c := naming.Russian().
		MustWithAliases(map[int][]string{4: {"x"}}).
		MustWithAliases(map[int][]string{4: {"y"}, 3: {"z"}})

	assert.Equal(t, []string{"x", "y"}, c.Aliases(4))
	assert.Equal(t, []string{"z"}, c.Aliases(3))
	assert.Equal(t, 4, c.MustASPN("x"))
	assert.Equal(t, 4, c.MustASPN("y"))
	assert.Equal(t, 3, c.MustASPN("z"))
}

// Empty alias strings are ignored, as in New: they name nothing and do not
// fail: a list of only empty strings leaves the catalog equal to the receiver.
func TestWithAliasesIgnoresEmpty(t *testing.T) {
	t.Parallel()

	c := naming.Russian().MustWithAliases(map[int][]string{4: {"", "x", ""}, 10: {"", ""}})
	assert.Equal(t, []string{"x"}, c.Aliases(4))
	assert.Empty(t, c.Aliases(10))

	empty := naming.Russian().MustWithAliases(map[int][]string{4: {""}, 10: {""}})
	assert.True(t, empty.Equal(naming.Russian()))
}

// A non-empty alias for an octave without a canonical name gives
// ErrUnnamedOctave, including on a nil or zero receiver.
func TestWithAliasesUnnamed(t *testing.T) {
	t.Parallel()

	cases := []namedCatalog{
		{name: "Russian", catalog: naming.Russian},
		{name: "nil", catalog: func() *naming.Catalog { return nil }},
		{name: "zero value", catalog: func() *naming.Catalog { return &naming.Catalog{} }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.catalog()
			got, err := c.WithAliases(map[int][]string{10: {"x"}})
			assert.Nil(t, got)
			assert.ErrorIs(t, err, naming.ErrUnnamedOctave)
			assert.NotErrorIs(t, err, naming.ErrDuplicateName)

			assertPanicsWithErrorIs(t, naming.ErrUnnamedOctave, naming.ErrDuplicateName,
				func() { c.MustWithAliases(map[int][]string{10: {"x"}}) })
		})
	}
}

// An alias that would resolve to two different octaves gives
// ErrDuplicateName: another octave's canonical name, another octave's alias
// already in the receiver, and the same new alias given to two octaves.
func TestWithAliasesDuplicate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		base  func() *naming.Catalog
		input map[int][]string
	}{
		{
			name:  "another octave's canonical name",
			base:  naming.Russian,
			input: map[int][]string{4: {"малая"}},
		},
		{
			name: "another octave's existing alias",
			base: func() *naming.Catalog {
				return naming.Russian().MustWithAliases(map[int][]string{3: {"x"}})
			},
			input: map[int][]string{4: {"x"}},
		},
		{
			name:  "same new alias for two octaves",
			base:  naming.Russian,
			input: map[int][]string{4: {"x"}, 3: {"x"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.base()
			got, err := c.WithAliases(tc.input)
			assert.Nil(t, got)
			assert.ErrorIs(t, err, naming.ErrDuplicateName)
			assert.NotErrorIs(t, err, naming.ErrUnnamedOctave)

			assertPanicsWithErrorIs(t, naming.ErrDuplicateName, naming.ErrUnnamedOctave,
				func() { c.MustWithAliases(tc.input) })
		})
	}
}

// The conditions are checked in order: an alias of an unnamed octave is
// reported before an ambiguous alias, whatever the map iteration order.
func TestWithAliasesErrorOrder(t *testing.T) {
	t.Parallel()

	input := map[int][]string{10: {"u"}, 4: {"малая"}, 3: {"малая"}}

	got, err := naming.Russian().WithAliases(input)
	require.Nil(t, got)
	require.ErrorIs(t, err, naming.ErrUnnamedOctave)

	for range 20 {
		again, errAgain := naming.Russian().WithAliases(input)
		assert.Nil(t, again)
		assert.Equal(t, err.Error(), errAgain.Error(), "the error must not depend on map iteration order")
	}
}

// The error message of an ambiguous alias is stable across calls with
// several collisions.
func TestWithAliasesDuplicateErrorStable(t *testing.T) {
	t.Parallel()

	input := map[int][]string{4: {"малая", "большая"}, 3: {"малая"}}

	var first error
	for range 20 {
		got, err := naming.Russian().WithAliases(input)
		require.Nil(t, got)
		require.ErrorIs(t, err, naming.ErrDuplicateName)

		if first == nil {
			first = err
		}

		assert.Equal(t, first.Error(), err.Error())
	}
}

// With several collisions the message reports the smallest alias and the
// two smallest octaves it would name, whatever the input order.
func TestWithAliasesDuplicateErrorChoosesSmallestAlias(t *testing.T) {
	t.Parallel()

	base := naming.MustNew(map[int]string{1: "a", 2: "b"})
	input := map[int][]string{1: {"b"}, 2: {"a"}}

	got, err := base.WithAliases(input)
	require.Nil(t, got)
	require.ErrorIs(t, err, naming.ErrDuplicateName)
	assert.Contains(t, err.Error(), `"a"`)
	assert.Contains(t, err.Error(), "1 and 2")
}

// WithAliases does not modify the receiver, on success or on error.
func TestWithAliasesKeepsReceiver(t *testing.T) {
	t.Parallel()

	for _, input := range []map[int][]string{
		{4: {"x", "y"}},
		{4: {"малая"}},
		{10: {"x"}},
	} {
		ru := naming.Russian()
		before := naming.Russian()

		_, _ = ru.WithAliases(input)

		assert.True(t, ru.Equal(before), "input %v", input)
		assert.Empty(t, ru.Aliases(4))
		assert.Equal(t, "первая", ru.Name(4))
		assert.Equal(t, 4, ru.MustASPN("первая"))

		_, err := ru.ASPN("x")
		assert.ErrorIs(t, err, naming.ErrUnknownName)
	}
}

// WithAliases does not retain the input: later changes to the map or its
// slices do not affect the new catalog.
func TestWithAliasesDoesNotRetainInput(t *testing.T) {
	t.Parallel()

	list := []string{"x"}
	input := map[int][]string{4: list}

	c, err := naming.Russian().WithAliases(input)
	require.NoError(t, err)

	list[0] = "changed"
	input[4] = []string{"other"}
	input[3] = []string{"more"}

	assert.Equal(t, []string{"x"}, c.Aliases(4))
	assert.Empty(t, c.Aliases(3))
	assert.Equal(t, 4, c.MustASPN("x"))

	_, err = c.ASPN("changed")
	assert.ErrorIs(t, err, naming.ErrUnknownName)
}

// Aliases are compared exactly, byte for byte: case, whitespace and io
// variants of an alias are not found.
func TestWithAliasesExactMatch(t *testing.T) {
	t.Parallel()

	c := naming.Russian().MustWithAliases(map[int][]string{4: {"первая октава", "FirstOctave"}})

	for _, name := range []string{"Первая октава", "первая октава ", " firstoctave", "первая октава\u0301"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := c.ASPN(name)
			assert.ErrorIs(t, err, naming.ErrUnknownName)
		})
	}
}

// Any int is a valid octave number for aliases, including the extreme int
// values.
func TestWithAliasesExtremeNumbers(t *testing.T) {
	t.Parallel()

	base := naming.MustNew(map[int]string{math.MinInt: "low", math.MaxInt: "high"})
	c := base.MustWithAliases(map[int][]string{math.MinInt: {"l"}, math.MaxInt: {"h"}})

	assert.Equal(t, math.MinInt, c.MustASPN("l"))
	assert.Equal(t, math.MaxInt, c.MustASPN("h"))
	assert.Equal(t, []string{"l"}, c.Aliases(math.MinInt))
	assert.Equal(t, []string{"h"}, c.Aliases(math.MaxInt))
	assert.Equal(t, "low", c.Name(math.MinInt))
	assert.Equal(t, "high", c.Name(math.MaxInt))
}

// A nil or zero receiver with an empty alias map gives a new catalog without
// names; MustWithAliases does not panic on it.
func TestWithAliasesNilAndZero(t *testing.T) {
	t.Parallel()

	for _, tc := range zeroReceivers() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.catalog()

			for _, input := range []map[int][]string{nil, {}, {1: {""}}} {
				got, err := c.WithAliases(input)
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.True(t, got.Equal(nil))
				assert.Empty(t, got.Aliases(1))
			}

			assert.NotPanics(t, func() { c.MustWithAliases(nil) })
		})
	}
}

// MustWithAliases returns the same catalog as WithAliases on success.
func TestMustWithAliases(t *testing.T) {
	t.Parallel()

	input := map[int][]string{4: {"x"}}
	want, err := naming.Russian().WithAliases(input)
	require.NoError(t, err)

	assert.True(t, want.Equal(naming.Russian().MustWithAliases(input)))
}

// Aliases returns a new caller-owned non-nil slice in ascending byte order;
// it is empty for an octave without aliases and on a nil or zero catalog.
func TestAliasesAccessor(t *testing.T) {
	t.Parallel()

	c := naming.Russian().MustWithAliases(map[int][]string{4: {"z", "a", "m"}})
	assert.Equal(t, []string{"a", "m", "z"}, c.Aliases(4))

	got := c.Aliases(4)
	got[0] = "changed"
	assert.Equal(t, []string{"a", "m", "z"}, c.Aliases(4), "the caller owns the slice")

	assert.Empty(t, c.Aliases(3))
	assert.Empty(t, c.Aliases(10))
	assert.NotNil(t, c.Aliases(10))

	for _, tc := range zeroReceivers() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.catalog().Aliases(4)
			assert.NotNil(t, got)
			assert.Empty(t, got)
		})
	}
}

// Equal takes aliases into account: the same canonical names with different
// alias-to-octave mappings are not equal; the order of aliases does not
// matter; catalogs without aliases behave as before.
func TestEqualWithAliases(t *testing.T) {
	t.Parallel()

	base := naming.MustNew(map[int]string{4: "middle", 3: "small"})

	cases := []struct {
		name        string
		left, right *naming.Catalog
		want        bool
	}{
		{
			name:  "same aliases",
			left:  base.MustWithAliases(map[int][]string{4: {"x", "y"}}),
			right: base.MustWithAliases(map[int][]string{4: {"y", "x"}}),
			want:  true,
		},
		{
			name:  "alias vs no alias",
			left:  base.MustWithAliases(map[int][]string{4: {"x"}}),
			right: base,
			want:  false,
		},
		{
			name:  "different alias",
			left:  base.MustWithAliases(map[int][]string{4: {"x"}}),
			right: base.MustWithAliases(map[int][]string{4: {"y"}}),
			want:  false,
		},
		{
			name:  "same alias, different octave",
			left:  base.MustWithAliases(map[int][]string{4: {"x"}}),
			right: base.MustWithAliases(map[int][]string{3: {"x"}}),
			want:  false,
		},
		{
			name:  "different canonical names, same aliases",
			left:  base.MustWithAliases(map[int][]string{4: {"x"}}),
			right: naming.MustNew(map[int]string{4: "other", 3: "small"}).MustWithAliases(map[int][]string{4: {"x"}}),
			want:  false,
		},
		{
			name:  "no aliases, both",
			left:  base,
			right: naming.MustNew(map[int]string{3: "small", 4: "middle"}),
			want:  true,
		},
		{
			name:  "nil vs extended",
			left:  nil,
			right: base.MustWithAliases(map[int][]string{4: {"x"}}),
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.left.Equal(tc.right))
			assert.Equal(t, tc.want, tc.right.Equal(tc.left), "symmetry")
		})
	}
}

// All keeps returning the canonical pairs only; aliases are not listed.
func TestAllWithAliases(t *testing.T) {
	t.Parallel()

	c := naming.MustNew(map[int]string{4: "middle", 3: "small"}).
		MustWithAliases(map[int][]string{4: {"x", "y"}})

	var pairs [][2]any
	for aspn, name := range c.All() {
		pairs = append(pairs, [2]any{aspn, name})
	}

	assert.Equal(t, [][2]any{{3, "small"}, {4, "middle"}}, pairs)
}

// The built-in catalogs have no aliases.
func TestBuiltinsHaveNoAliases(t *testing.T) {
	t.Parallel()

	for _, b := range builtins() {
		t.Run(b.name, func(t *testing.T) {
			t.Parallel()

			c := b.catalog()
			for aspn := range b.oracle() {
				assert.Empty(t, c.Aliases(aspn))
			}
		})
	}
}

// A copy of an extended catalog behaves like the original.
func TestWithAliasesCopy(t *testing.T) {
	t.Parallel()

	c := aliasCatalog()
	cp := *c

	assert.True(t, cp.Equal(c))
	assert.Equal(t, 4, cp.MustASPN("FirstOctave"))
	assert.Equal(t, []string{"FirstOctave", "первая октава"}, cp.Aliases(4))
}

// WithAliases on a shared catalog and reads of it are safe for concurrent
// use: the receiver is immutable.
func TestWithAliasesConcurrent(t *testing.T) {
	t.Parallel()

	ru := naming.Russian()

	const workers = 8
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range 20 {
				c, err := ru.WithAliases(map[int][]string{4: {"x"}})
				if err != nil || c.MustASPN("x") != 4 {
					t.Error(err)

					return
				}

				if ru.MustASPN("первая") != 4 || len(ru.Aliases(4)) != 0 {
					t.Error("the shared receiver changed")

					return
				}
			}
		}()
	}

	wg.Wait()
}
