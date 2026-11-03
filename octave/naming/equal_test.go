package naming_test

import (
	"reflect"
	"testing"

	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pairsInOrder returns a map filled with the given pairs in the given order.
func pairsInOrder(numbers []int, names []string) map[int]string {
	pairs := make(map[int]string, len(numbers))
	for i, n := range numbers {
		pairs[n] = names[i]
	}

	return pairs
}

// swapNames returns pairs with the names of octaves a and b exchanged.
func swapNames(pairs map[int]string, a, b int) map[int]string {
	pairs[a], pairs[b] = pairs[b], pairs[a]

	return pairs
}

// equalityClass is a catalog maker with the class of catalogs it is equal to.
type equalityClass struct {
	namedCatalog
	class int
}

// K-13: two catalogs are equal exactly when they name the same octaves with
// the same names; the same numbers and names paired differently are not equal.
// Every pair of the set below, including a catalog with itself, is compared in
// both positions: catalogs of one class are equal, of different classes are
// not. This also checks reflexivity, symmetry and transitivity. Pointer
// identity does not matter: every maker returns a new catalog.
func TestEqual(t *testing.T) {
	t.Parallel()

	must := func(pairs map[int]string) func() *naming.Catalog {
		return func() *naming.Catalog { return naming.MustNew(pairs) }
	}
	set := []equalityClass{
		// Catalogs without names.
		{namedCatalog{"nil", func() *naming.Catalog { return nil }}, 0},
		{namedCatalog{"zero value", func() *naming.Catalog { return &naming.Catalog{} }}, 0},
		{namedCatalog{"MustNew(nil)", must(nil)}, 0},
		{namedCatalog{"empty map", must(map[int]string{})}, 0},
		{namedCatalog{"empty names only", must(map[int]string{1: "", 2: ""})}, 0},
		// An empty name is not a pair of the catalog.
		{namedCatalog{"{1 a, 2 empty}", must(map[int]string{1: "a", 2: ""})}, 1},
		{namedCatalog{"{1 a}", must(map[int]string{1: "a"})}, 1},
		// The same pairs filled in a different order.
		{namedCatalog{"{1 a, 2 b} ascending", must(pairsInOrder([]int{1, 2}, []string{"a", "b"}))}, 2},
		{namedCatalog{"{1 a, 2 b} descending", must(pairsInOrder([]int{2, 1}, []string{"b", "a"}))}, 2},
		// One name differs.
		{namedCatalog{"{1 a, 2 c}", must(map[int]string{1: "a", 2: "c"})}, 3},
		// One number differs.
		{namedCatalog{"{1 a, 3 b}", must(map[int]string{1: "a", 3: "b"})}, 4},
		// Names differ only in case or spaces.
		{namedCatalog{"{1 A}", must(map[int]string{1: "A"})}, 5},
		{namedCatalog{"{1 a with space}", must(map[int]string{1: "a "})}, 6},
		// Built-in catalogs and custom catalogs with the same pairs.
		{namedCatalog{"Russian()", naming.Russian}, 7},
		{namedCatalog{"MustNew(Russian oracle)", must(russianOracle())}, 7},
		{namedCatalog{"English()", naming.English}, 8},
		{namedCatalog{"MustNew(English oracle)", must(englishOracle())}, 8},
		{namedCatalog{"German()", naming.German}, 9},
		{namedCatalog{"MustNew(German oracle)", must(germanOracle())}, 9},
		// A built-in name in a custom catalog is not the built-in catalog.
		{namedCatalog{"{4 первая}", must(map[int]string{4: "первая"})}, 10},
		// The same numbers and the same names, paired differently.
		{namedCatalog{"{1 b, 2 a}", must(map[int]string{1: "b", 2: "a"})}, 11},
		{namedCatalog{"{1 a, 2 b, 3 c}", must(map[int]string{1: "a", 2: "b", 3: "c"})}, 12},
		{namedCatalog{"{1 b, 2 c, 3 a}", must(map[int]string{1: "b", 2: "c", 3: "a"})}, 13},
		{namedCatalog{"Russian oracle with ASPN 2 and 3 swapped", must(swapNames(russianOracle(), 2, 3))}, 14},
	}

	for _, a := range set {
		t.Run(a.name, func(t *testing.T) {
			t.Parallel()

			for _, b := range set {
				want := a.class == b.class
				assert.Equal(t, want, a.catalog().Equal(b.catalog()), "%s.Equal(%s)", a.name, b.name)
			}

			same := a.catalog()
			assert.True(t, same.Equal(same), "%s.Equal(itself)", a.name)
		})
	}
}

// K-13: Equal does not change the catalogs it compares.
func TestEqualKeepsCatalogs(t *testing.T) {
	t.Parallel()

	ru, custom := naming.Russian(), naming.MustNew(customPairs())
	require.NotNil(t, ru)
	require.NotNil(t, custom)

	assert.False(t, ru.Equal(custom))
	assert.False(t, custom.Equal(ru))

	for n, want := range russianOracle() {
		assert.Equal(t, want, ru.Name(n), "ASPN %d", n)
	}
	for n, want := range customPairs() {
		assert.Equal(t, want, custom.Name(n), "ASPN %d", n)
	}
	assert.True(t, ru.Equal(naming.Russian()))
	assert.True(t, custom.Equal(naming.MustNew(customPairs())))
}

// Catalog is a struct with unexported fields only and is
// not comparable, so == cannot be used on catalog values by mistake. This
// observes the public type through reflect; it does not read the
// implementation.
func TestCatalogType(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeFor[naming.Catalog]()
	require.Equal(t, reflect.Struct, typ.Kind())
	assert.False(t, typ.Comparable())

	for i := range typ.NumField() {
		assert.False(t, typ.Field(i).IsExported(), "exported field %s", typ.Field(i).Name)
	}
}
