package naming_test

import (
	"fmt"
	"maps"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/octave/naming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The error texts are not part of the API; these tests check only the
// properties the contract promises: the length bound, the quoted excerpt of a
// name, the decimal length of a long name and the numbers of a repeated name.
// Numbers and lengths in "contains" and "does not contain" checks are unique
// multi-digit values chosen so that no checked number is a substring of
// another one or of a name length in the same text; the long names below
// consist of letters only.

// maxErrorText is the upper bound of len(err.Error()) for names up to 1 MiB.
const maxErrorText = 512

// excerptBytes is the longest name quoted in full.
const excerptBytes = 32

// errorPath is one of the four ways a name reaches an error text: lookup of
// an unknown name and creation with a repeated name, each directly and as the
// panic of the Must variant.
type errorPath struct {
	name   string
	target error
	other  error
	err    func(name string) error
}

// errorPaths returns the four paths. The creation paths repeat the name at
// ASPN numbers 1 and 2.
func errorPaths() []errorPath {
	return []errorPath{
		{
			name: "ASPN", target: naming.ErrUnknownName, other: naming.ErrDuplicateName,
			err: func(name string) error {
				aspn, err := naming.English().ASPN(name)
				if aspn != 0 {
					return fmt.Errorf("ASPN returned %d, want 0", aspn)
				}

				return err
			},
		},
		{
			name: "MustASPN", target: naming.ErrUnknownName, other: naming.ErrDuplicateName,
			err: func(name string) error {
				return panicError(func() { naming.English().MustASPN(name) })
			},
		},
		{
			name: "New", target: naming.ErrDuplicateName, other: naming.ErrUnknownName,
			err: func(name string) error {
				c, err := naming.New(map[int]string{1: name, 2: name})
				if c != nil {
					return fmt.Errorf("New returned a catalog")
				}

				return err
			},
		},
		{
			name: "MustNew", target: naming.ErrDuplicateName, other: naming.ErrUnknownName,
			err: func(name string) error {
				return panicError(func() { naming.MustNew(map[int]string{1: name, 2: name}) })
			},
		},
	}
}

// panicError calls f and returns the error it panicked with. A missing panic
// or a panic value that is not an error is returned as a separate error, which
// fails the classification checks of the caller.
func panicError(f func()) error {
	recovered, panicked := recoverPanic(f)
	if !panicked {
		return fmt.Errorf("no panic")
	}

	err, ok := recovered.(error)
	if !ok {
		return fmt.Errorf("panic value %T (%v) is not an error", recovered, recovered)
	}

	return err
}

// errorText returns the text of the error of path p for name after checking
// its classification.
func errorText(t *testing.T, p errorPath, name string) string {
	t.Helper()

	err := p.err(name)
	require.ErrorIs(t, err, p.target)
	require.NotErrorIs(t, err, p.other)

	return err.Error()
}

// The text of every error stays within maxErrorText bytes for names of 1 KiB
// and 1 MiB, including the worst expansions of quoting ("\x00", invalid bytes)
// and multi-byte runes, on all four paths.
func TestErrorTextLengthBound(t *testing.T) {
	t.Parallel()

	contents := []struct {
		name string
		unit string
	}{
		{name: "ASCII", unit: "a"},
		{name: "NUL", unit: "\x00"},
		{name: "Cyrillic", unit: "ё"},
		{name: "invalid byte", unit: "\xff"},
	}
	sizes := []struct {
		name  string
		bytes int
	}{
		{name: "1KiB", bytes: 1 << 10},
		{name: "1MiB", bytes: 1 << 20},
	}

	for _, content := range contents {
		for _, size := range sizes {
			for _, p := range errorPaths() {
				t.Run(content.name+"/"+size.name+"/"+p.name, func(t *testing.T) {
					t.Parallel()

					name := strings.Repeat(content.unit, size.bytes/len(content.unit))
					require.Len(t, name, size.bytes)

					text := errorText(t, p, name)
					assert.LessOrEqual(t, len(text), maxErrorText, "len(err.Error())")
				})
			}
		}
	}
}

// A name of up to excerptBytes bytes is quoted in full on all four paths.
func TestErrorTextShortName(t *testing.T) {
	t.Parallel()

	names := []struct {
		name  string
		value string
	}{
		{name: "1 byte", value: "q"},
		{name: "31 bytes", value: strings.Repeat("q", 31)},
		{name: "32 bytes", value: strings.Repeat("q", excerptBytes)},
		{name: "31 bytes Cyrillic", value: "q" + strings.Repeat("ё", 15)},
		{name: "32 bytes Cyrillic", value: strings.Repeat("ё", 16)},
	}

	for _, n := range names {
		for _, p := range errorPaths() {
			t.Run(n.name+"/"+p.name, func(t *testing.T) {
				t.Parallel()

				require.LessOrEqual(t, len(n.value), excerptBytes)

				text := errorText(t, p, n.value)
				assert.Contains(t, text, strconv.Quote(n.value))
			})
		}
	}
}

// A longer name is quoted as its prefix up to the last rune that fits in
// excerptBytes bytes, followed by its length in bytes. A multi-byte rune
// across the boundary is not split, so its escaped first byte (`\xd1` for
// "ё") does not appear; a rune that ends at the boundary is kept. Invalid
// bytes count as runes of length 1.
func TestErrorTextLongNameExcerpt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		value    string
		excerpt  int
		absent   string
		wantSize int
	}{
		{
			name:     "rune across the boundary",
			value:    strings.Repeat("a", 31) + "ё" + strings.Repeat("b", 10),
			excerpt:  31,
			absent:   `\xd1`,
			wantSize: 43,
		},
		{
			name:     "rune ending at the boundary",
			value:    strings.Repeat("a", 30) + "ё" + strings.Repeat("b", 11),
			excerpt:  32,
			wantSize: 43,
		},
		{
			name:     "invalid bytes only",
			value:    strings.Repeat("\xff", 40),
			excerpt:  32,
			wantSize: 40,
		},
		{
			name:     "lone leading byte at the boundary",
			value:    strings.Repeat("a", 31) + "\xd1" + strings.Repeat("a", 10),
			excerpt:  32,
			wantSize: 42,
		},
	}

	for _, tc := range cases {
		for _, p := range errorPaths() {
			t.Run(tc.name+"/"+p.name, func(t *testing.T) {
				t.Parallel()

				require.Len(t, tc.value, tc.wantSize)

				text := errorText(t, p, tc.value)
				assert.Contains(t, text, strconv.Quote(tc.value[:tc.excerpt]), "excerpt")
				assert.Contains(t, text, strconv.Itoa(tc.wantSize), "length of the name")
				if tc.absent != "" {
					assert.NotContains(t, text, tc.absent, "split rune")
				}
			})
		}
	}
}

// Names with the same first excerptBytes bytes are different names: they are
// stored, named and found separately, and a third name with that prefix or
// the prefix itself is not found.
func TestLongNamesWithCommonPrefix(t *testing.T) {
	t.Parallel()

	prefix := strings.Repeat("p", excerptBytes)
	pairs := map[int]string{1: prefix + "x", 2: prefix + "y"}

	constructors := []struct {
		name string
		make func(map[int]string) (*naming.Catalog, error)
	}{
		{name: "New", make: naming.New},
		{name: "MustNew", make: func(m map[int]string) (*naming.Catalog, error) { return naming.MustNew(m), nil }},
	}

	for _, ctor := range constructors {
		t.Run(ctor.name, func(t *testing.T) {
			t.Parallel()

			var (
				c   *naming.Catalog
				err error
			)
			require.NotPanics(t, func() { c, err = ctor.make(pairs) })
			require.NoError(t, err)
			require.NotNil(t, c)

			for aspn, name := range pairs {
				assert.Equal(t, name, c.Name(aspn), "ASPN %d", aspn)

				got, err := c.ASPN(name)
				if assert.NoError(t, err, "name %q", name) {
					assert.Equal(t, aspn, got, "name %q", name)
				}
			}

			for _, unknown := range []string{prefix + "z", prefix} {
				got, err := c.ASPN(unknown)
				assert.ErrorIs(t, err, naming.ErrUnknownName, "name %q", unknown)
				assert.Equal(t, 0, got, "name %q", unknown)
			}
		})
	}
}

// The repeated name in the error of New and MustNew is the smallest repeated
// name by its full bytes (not by its excerpt), with its two smallest numbers.
// The two names of each case share their first excerptBytes bytes.
func TestDuplicateErrorChoosesByFullName(t *testing.T) {
	t.Parallel()

	prefix := strings.Repeat("p", excerptBytes)

	cases := []struct {
		name    string
		pairs   func() map[int]string
		present []string
		absent  []string
	}{
		{
			// A = prefix+"a" (33 bytes) is smaller than B = prefix+"bb" (34 bytes).
			name: "shorter name is smaller",
			pairs: func() map[int]string {
				return map[int]string{101: prefix + "a", 202: prefix + "a", -7: prefix + "bb", -8: prefix + "bb"}
			},
			present: []string{"101", "202", "33"},
			absent:  []string{"-7", "-8", "34"},
		},
		{
			// A = prefix+"aa" (34 bytes) is smaller than B = prefix+"b" (33 bytes).
			name: "longer name is smaller",
			pairs: func() map[int]string {
				return map[int]string{4001: prefix + "aa", 4002: prefix + "aa", 6001: prefix + "b", 6002: prefix + "b"}
			},
			present: []string{"4001", "4002", "34"},
			absent:  []string{"6001", "6002", "33"},
		},
		{
			// One name of 33 bytes at three numbers: the two smallest are given.
			name: "three numbers",
			pairs: func() map[int]string {
				return map[int]string{5003: prefix + "q", 3007: prefix + "q", 9001: prefix + "q"}
			},
			present: []string{"3007", "5003", "33"},
			absent:  []string{"9001"},
		},
	}

	paths := []struct {
		name string
		err  func(map[int]string) error
	}{
		{name: "New", err: func(m map[int]string) error {
			_, err := naming.New(m)

			return err
		}},
		{name: "MustNew", err: func(m map[int]string) error {
			return panicError(func() { naming.MustNew(m) })
		}},
	}

	for _, tc := range cases {
		for _, p := range paths {
			t.Run(tc.name+"/"+p.name, func(t *testing.T) {
				t.Parallel()

				err := p.err(tc.pairs())
				require.ErrorIs(t, err, naming.ErrDuplicateName)
				require.NotErrorIs(t, err, naming.ErrUnknownName)

				text := err.Error()
				assert.Contains(t, text, strconv.Quote(prefix), "excerpt")
				for _, s := range tc.present {
					assert.Contains(t, text, s)
				}
				for _, s := range tc.absent {
					assert.NotContains(t, text, s)
				}
			})
		}
	}
}

// The error of New for the same input has the same text on every call,
// whatever the iteration order of the map.
func TestDuplicateErrorIsReproducible(t *testing.T) {
	t.Parallel()

	const calls = 50

	prefix := strings.Repeat("p", excerptBytes)
	pairs := map[int]string{
		1001: prefix + "x", 1002: prefix + "x", 1003: prefix + "x",
		2001: prefix + "w", 2002: prefix + "w",
		3001: "y", 3002: "y",
	}

	_, first := naming.New(pairs)
	require.ErrorIs(t, first, naming.ErrDuplicateName)

	for call := range calls {
		_, err := naming.New(pairs)
		require.ErrorIs(t, err, naming.ErrDuplicateName)
		if !assert.Equal(t, first.Error(), err.Error(), "same map, call %d", call) {
			break
		}

		_, err = naming.New(maps.Clone(pairs))
		require.ErrorIs(t, err, naming.ErrDuplicateName)
		if !assert.Equal(t, first.Error(), err.Error(), "copy of the map, call %d", call) {
			break
		}
	}
}
