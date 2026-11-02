package naming_test

import (
	"testing"

	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// zeroReceivers returns the catalogs not created by New or a built-in
// function, which behave as a catalog without names: the address of a zero
// value variable, a pointer to an empty literal and the nil pointer.
func zeroReceivers() []namedCatalog {
	return []namedCatalog{
		{name: "zero value variable", catalog: func() *naming.Catalog {
			var z naming.Catalog

			return &z
		}},
		{name: "pointer to empty literal", catalog: func() *naming.Catalog { return &naming.Catalog{} }},
		{name: "nil pointer", catalog: func() *naming.Catalog { return nil }},
	}
}

// K-11: the zero value and the nil pointer are a catalog without
// names. Name gives "" for every number, ASPN gives exactly 0 and
// ErrUnknownName for every name, MustASPN panics with ErrUnknownName, and no
// other method panics. Such a catalog never gives a name it does not have.
func TestZeroAndNil(t *testing.T) {
	t.Parallel()

	for _, tc := range zeroReceivers() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.catalog()

			for _, n := range append(probeNumbers(), -1, 4, 9) {
				var got string
				if assert.NotPanics(t, func() { got = c.Name(n) }, "ASPN %d", n) {
					assert.Empty(t, got, "ASPN %d", n)
				}
			}

			for _, s := range []string{"", " ", "a", "первая", "subcontra", "Subkontraoktave"} {
				var (
					got int
					err error
				)
				require.NotPanics(t, func() { got, err = c.ASPN(s) }, "name %q", s)
				assert.ErrorIs(t, err, naming.ErrUnknownName, "name %q", s)
				assert.NotErrorIs(t, err, naming.ErrDuplicateName, "name %q", s)
				assert.Equal(t, 0, got, "name %q", s)

				assertPanicsWithErrorIs(t, naming.ErrUnknownName, naming.ErrDuplicateName,
					func() { c.MustASPN(s) }, "name %q", s)
			}
		})
	}
}

// K-11, K-13: the zero value and nil are equal to each other and to every
// catalog without names, in both positions; they are not equal to a catalog
// with names, in both positions. Equal does not panic on them.
func TestZeroAndNilEqual(t *testing.T) {
	t.Parallel()

	withoutNames := []namedCatalog{
		{name: "MustNew(nil)", catalog: func() *naming.Catalog { return naming.MustNew(nil) }},
		{name: "MustNew(empty map)", catalog: func() *naming.Catalog { return naming.MustNew(map[int]string{}) }},
		{name: "MustNew(empty name)", catalog: func() *naming.Catalog { return naming.MustNew(map[int]string{1: ""}) }},
	}
	withoutNames = append(withoutNames, zeroReceivers()...)

	withNames := []namedCatalog{
		{name: "Russian", catalog: naming.Russian},
		{name: "English", catalog: naming.English},
		{name: "German", catalog: naming.German},
		{name: "one name", catalog: func() *naming.Catalog { return naming.MustNew(map[int]string{1: "a"}) }},
	}

	for _, z := range zeroReceivers() {
		t.Run(z.name, func(t *testing.T) {
			t.Parallel()

			for _, other := range withoutNames {
				var forward, backward bool
				require.NotPanics(t, func() {
					forward = z.catalog().Equal(other.catalog())
					backward = other.catalog().Equal(z.catalog())
				}, other.name)
				assert.True(t, forward, "%s.Equal(%s)", z.name, other.name)
				assert.True(t, backward, "%s.Equal(%s)", other.name, z.name)
			}

			for _, other := range withNames {
				var forward, backward bool
				require.NotPanics(t, func() {
					forward = z.catalog().Equal(other.catalog())
					backward = other.catalog().Equal(z.catalog())
				}, other.name)
				assert.False(t, forward, "%s.Equal(%s)", z.name, other.name)
				assert.False(t, backward, "%s.Equal(%s)", other.name, z.name)
			}
		})
	}
}

// K-11: methods are called directly on an addressable zero value variable,
// as in "var c naming.Catalog; c.Name(4)", and on a nil *Catalog variable.
func TestZeroValueVariable(t *testing.T) {
	t.Parallel()

	t.Run("value", func(t *testing.T) {
		t.Parallel()

		var c naming.Catalog

		assert.Empty(t, c.Name(4))

		got, err := c.ASPN("первая")
		assert.ErrorIs(t, err, naming.ErrUnknownName)
		assert.Equal(t, 0, got)

		assert.True(t, c.Equal(nil))
		assert.True(t, c.Equal(&naming.Catalog{}))
		assert.False(t, c.Equal(naming.Russian()))
	})

	t.Run("nil pointer", func(t *testing.T) {
		t.Parallel()

		var c *naming.Catalog

		assert.Empty(t, c.Name(4))

		got, err := c.ASPN("x")
		assert.ErrorIs(t, err, naming.ErrUnknownName)
		assert.Equal(t, 0, got)

		assert.True(t, c.Equal(nil))
		assert.True(t, c.Equal(naming.MustNew(nil)))
		assert.False(t, c.Equal(naming.English()))
	})
}
