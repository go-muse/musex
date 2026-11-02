package naming_test

import (
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The oracle tables themselves: valid UTF-8, no combining marks
// (U+0300...U+036F, which a decomposed Cyrillic io would contain), no spaces
// at the ends, and distinct names within each catalog. This guards the
// literals typed from the requirements; it does not call the package.
func TestOracleTables(t *testing.T) {
	t.Parallel()

	for _, b := range builtins() {
		t.Run(b.name, func(t *testing.T) {
			t.Parallel()

			oracle := b.oracle()
			assert.Len(t, inverse(oracle), len(oracle), "names are distinct")

			for n, name := range oracle {
				assert.True(t, utf8.ValidString(name), "ASPN %d", n)
				assert.Equal(t, strings.TrimSpace(name), name, "ASPN %d", n)
				assert.NotEmpty(t, name, "ASPN %d", n)

				for _, r := range name {
					assert.False(t, r >= 0x0300 && r <= 0x036F, "ASPN %d: combining mark %U", n, r)
				}
			}
		})
	}
}

// K-1, K-2, K-3: every named octave of a built-in catalog has
// exactly the name of the requirements.
func TestBuiltinNames(t *testing.T) {
	t.Parallel()

	for _, b := range builtins() {
		t.Run(b.name, func(t *testing.T) {
			t.Parallel()

			for n, want := range b.oracle() {
				t.Run(strconv.Itoa(n), func(t *testing.T) {
					t.Parallel()

					c := b.catalog()
					require.NotNil(t, c)
					assert.Equal(t, want, c.Name(n))
				})
			}
		})
	}
}

// K-1, K-3: the exact bytes of the names with Cyrillic io (U+0451) and
// sharp s (U+00DF).
// The expected values are the UTF-8 encodings of the NFC strings.
func TestBuiltinNameBytes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		catalog func() *naming.Catalog
		aspn    int
		wantHex string
	}{
		{name: "Russian четвёртая", catalog: naming.Russian, aspn: 7, wantHex: "d187d0b5d182d0b2d191d180d182d0b0d18f"},
		{name: "German große", catalog: naming.German, aspn: 2, wantHex: "67726fc39f65"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := tc.catalog()
			require.NotNil(t, c)
			assert.Equal(t, tc.wantHex, hex.EncodeToString([]byte(c.Name(tc.aspn))))
		})
	}
}

// K-4: the octaves named in the criterion have no name, without an
// error or a panic.
func TestBuiltinUnnamedOctaves(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		catalog func() *naming.Catalog
		numbers []int
	}{
		{name: "Russian", catalog: naming.Russian, numbers: append([]int{-2, 10}, edgeNumbers()...)},
		{name: "English", catalog: naming.English, numbers: append([]int{-1, 8, 9}, edgeNumbers()...)},
		{name: "German", catalog: naming.German, numbers: append([]int{-1, 8, 9, 10}, edgeNumbers()...)},
	}

	for _, tc := range cases {
		for _, n := range tc.numbers {
			t.Run(tc.name+"/"+strconv.Itoa(n), func(t *testing.T) {
				t.Parallel()

				c := tc.catalog()
				require.NotNil(t, c)

				var got string
				require.NotPanics(t, func() { got = c.Name(n) })
				assert.Empty(t, got)
			})
		}
	}
}

// K-4: apart from the octaves of K-1...K-3, no number in -20...20 and none of
// the extreme numbers has a name; named numbers give the oracle name.
func TestBuiltinNameProbe(t *testing.T) {
	t.Parallel()

	for _, b := range builtins() {
		t.Run(b.name, func(t *testing.T) {
			t.Parallel()

			c := b.catalog()
			require.NotNil(t, c)

			oracle := b.oracle()
			for _, n := range probeNumbers() {
				assert.Equal(t, oracle[n], c.Name(n), "ASPN %d", n)
			}
		})
	}
}

// K-5: every name of a built-in catalog is found in that catalog and
// gives its ASPN number; MustASPN gives the same number.
func TestBuiltinASPN(t *testing.T) {
	t.Parallel()

	for _, b := range builtins() {
		t.Run(b.name, func(t *testing.T) {
			t.Parallel()

			for want, name := range b.oracle() {
				t.Run(name, func(t *testing.T) {
					t.Parallel()

					c := b.catalog()
					require.NotNil(t, c)

					got, err := c.ASPN(name)
					require.NoError(t, err)
					assert.Equal(t, want, got)

					var must int
					require.NotPanics(t, func() { must = c.MustASPN(name) })
					assert.Equal(t, want, must)
				})
			}
		})
	}
}

// K-5: lookup and naming agree: ASPN(Name(k)) == k for every named
// octave k, and Name(ASPN(n)) == n for every name n of the catalog.
func TestBuiltinRoundTrip(t *testing.T) {
	t.Parallel()

	for _, b := range builtins() {
		t.Run(b.name, func(t *testing.T) {
			t.Parallel()

			c := b.catalog()
			require.NotNil(t, c)

			oracle := b.oracle()
			for k := range oracle {
				got, err := c.ASPN(c.Name(k))
				if assert.NoError(t, err, "ASPN %d", k) {
					assert.Equal(t, k, got)
				}
			}

			for name := range inverse(oracle) {
				aspn, err := c.ASPN(name)
				if assert.NoError(t, err, "name %q", name) {
					assert.Equal(t, name, c.Name(aspn))
				}
			}
		})
	}
}

// K-6: a name of another built-in catalog is not found; ASPN returns
// exactly 0 with ErrUnknownName.
func TestBuiltinForeignNames(t *testing.T) {
	t.Parallel()

	for _, b := range builtins() {
		t.Run(b.name, func(t *testing.T) {
			t.Parallel()

			c := b.catalog()
			require.NotNil(t, c)

			own := inverse(b.oracle())
			for _, other := range builtins() {
				for _, name := range other.oracle() {
					if _, ok := own[name]; ok {
						continue
					}

					got, err := c.ASPN(name)
					assert.ErrorIs(t, err, naming.ErrUnknownName, "name %q of %s", name, other.name)
					assert.NotErrorIs(t, err, naming.ErrDuplicateName, "name %q of %s", name, other.name)
					assert.Equal(t, 0, got, "name %q of %s", name, other.name)
				}
			}
		})
	}
}

// Every call of a built-in function returns a new non-nil
// catalog; two calls give different pointers that are equal by Equal.
func TestBuiltinNewPointer(t *testing.T) {
	t.Parallel()

	for _, b := range builtins() {
		t.Run(b.name, func(t *testing.T) {
			t.Parallel()

			first, second := b.catalog(), b.catalog()
			require.NotNil(t, first)
			require.NotNil(t, second)
			assert.NotSame(t, first, second)
			assert.True(t, first.Equal(second))
			assert.True(t, second.Equal(first))
		})
	}
}

// K-16: Name returns without an error or a panic for every number of
// K-1...K-4 in the three built-in, a custom and an empty catalog.
func TestNameNeverPanics(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		catalog func() *naming.Catalog
		oracle  map[int]string
	}{
		{name: "Russian", catalog: naming.Russian, oracle: russianOracle()},
		{name: "English", catalog: naming.English, oracle: englishOracle()},
		{name: "German", catalog: naming.German, oracle: germanOracle()},
		{name: "custom", catalog: func() *naming.Catalog { return naming.MustNew(customPairs()) }, oracle: customPairs()},
		{name: "empty", catalog: func() *naming.Catalog { return naming.MustNew(nil) }, oracle: map[int]string{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var c *naming.Catalog
			require.NotPanics(t, func() { c = tc.catalog() })
			require.NotNil(t, c)

			for _, n := range probeNumbers() {
				var got string
				if assert.NotPanics(t, func() { got = c.Name(n) }, "ASPN %d", n) {
					assert.Equal(t, tc.oracle[n], got, "ASPN %d", n)
				}
			}
		})
	}
}
