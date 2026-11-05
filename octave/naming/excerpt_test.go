package naming_test

import (
	"fmt"
	"maps"
	"slices"
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
// consist of letters only. The ASPN numbers of the creation paths are four
// digits long, so that they are neither a length of a short name nor a
// substring of a length of a long one.

// maxErrorText is the upper bound of len(err.Error()) for names up to 1 MiB.
const maxErrorText = 512

// excerptBytes is the longest name quoted in full.
const excerptBytes = 32

// The two ASPN numbers at which the creation paths repeat a name.
const (
	repeatedLow  = 4107
	repeatedHigh = 5208
)

// errorPath is one of the four ways a name reaches an error text: lookup of
// an unknown name and creation with a repeated name, each directly and as the
// panic of the Must variant.
type errorPath struct {
	name   string
	target error
	other  error
	// numbers are the ASPN numbers the text of the path contains: the two
	// numbers of the repeated name on the creation paths, none on lookup.
	numbers []string
	err     func(name string) error
}

// errorPaths returns the four paths. The creation paths repeat the name at
// the ASPN numbers repeatedLow and repeatedHigh.
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
			numbers: repeatedNumbers(),
			err: func(name string) error {
				c, err := naming.New(repeatedPairs(name))
				if c != nil {
					return fmt.Errorf("New returned a catalog")
				}

				return err
			},
		},
		{
			name: "MustNew", target: naming.ErrDuplicateName, other: naming.ErrUnknownName,
			numbers: repeatedNumbers(),
			err: func(name string) error {
				return panicError(func() { naming.MustNew(repeatedPairs(name)) })
			},
		},
	}
}

// repeatedPairs returns the pairs that give name to both repeated ASPN numbers.
func repeatedPairs(name string) map[int]string {
	return map[int]string{repeatedLow: name, repeatedHigh: name}
}

// repeatedNumbers returns the decimal forms of the repeated ASPN numbers.
func repeatedNumbers() []string {
	return []string{strconv.Itoa(repeatedLow), strconv.Itoa(repeatedHigh)}
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

// unexpectedNumber is the oracle of "no length annotation" for the text msg of
// the error for a name quoted as quoted. It cuts the first occurrence of
// quoted out of msg and returns the first decimal number (a run of ASCII
// digits) of the rest that is not one of allowed, or "" if there is none.
// A quoted name may contain digits, as a name of digits or an escape such as
// \x82 does, and these digits are not inspected. The ASPN numbers of the path
// are allowed, wherever they stand. The length of the name, with a right or
// a wrong value and in any wording, is a number outside both, so it is found.
// The position of the annotation and the words around it are not assumed.
//
// Assumed about the text of a short name: apart from the quotation and the
// ASPN numbers it has no digits, as with the sentinel text of the package.
// Accepted limits: an annotation equal to an allowed ASPN number and an
// annotation without digits (for example, only an ellipsis) are not found;
// a regression of that kind is left to review-code.
func unexpectedNumber(msg, quoted string, allowed []string) string {
	i := strings.Index(msg, quoted)
	if i < 0 {
		return "no quotation of the name"
	}

	rest := msg[:i] + " " + msg[i+len(quoted):]
	isNotDigit := func(r rune) bool { return r < '0' || r > '9' }

	for _, number := range strings.FieldsFunc(rest, isNotDigit) {
		if !slices.Contains(allowed, number) {
			return "number " + number + " outside the quotation"
		}
	}

	return ""
}

// The oracle accepts the texts of a correct short name and finds a length
// annotation of a mutant, whatever the number in it. The texts are made up
// for the oracle and are not the format of the package.
func TestUnexpectedNumber(t *testing.T) {
	t.Parallel()

	const (
		unknown   = "naming: unknown octave name: "
		duplicate = "naming: duplicate octave name: "
	)

	numbers := repeatedNumbers()
	digits := strings.Repeat("32", 16)

	tests := []struct {
		name      string
		nameValue string
		msg       func(quoted string) string
		allowed   []string
		want      bool
	}{
		{"lookup, plain name", "q", func(q string) string { return unknown + q }, nil, false},
		{"lookup, name of a digit", "7", func(q string) string { return unknown + q }, nil, false},
		{"lookup, name of digits", digits, func(q string) string { return unknown + q }, nil, false},
		{"lookup, escapes with digits", "\xe2\x82", func(q string) string { return unknown + q }, nil, false},
		{"lookup, NUL", "\x00", func(q string) string { return unknown + q }, nil, false},
		{
			"creation, numbers after the quotation", "q",
			func(q string) string { return duplicate + q + " at 4107 and 5208" }, numbers, false,
		},
		{
			"creation, numbers before the quotation", "q",
			func(q string) string { return duplicate + "4107, 5208: " + q }, numbers, false,
		},
		{
			"creation, numbers around the quotation", digits,
			func(q string) string { return duplicate + "4107 " + q + " and 5208" }, numbers, false,
		},
		{
			"creation, negative numbers", "q",
			func(q string) string { return duplicate + q + " at -4107 and -5208" }, numbers, false,
		},
		{
			"creation, escapes with digits", "\xe2\x82",
			func(q string) string { return duplicate + q + " at 4107 and 5208" }, numbers, false,
		},

		{"annotation of 1 byte", "q", func(q string) string { return unknown + q + "... (1 bytes)" }, nil, true},
		{"annotation with a wrong number", digits, func(q string) string { return unknown + q + "... (0 bytes)" }, nil, true},
		{"annotation with the right number", digits, func(q string) string { return unknown + q + "... (32 bytes)" }, nil, true},
		{"annotation in parentheses", "ab", func(q string) string { return unknown + q + " (2)" }, nil, true},
		{"annotation in brackets", "ab", func(q string) string { return unknown + q + " [2 bytes]" }, nil, true},
		{"annotation as a key", "ab", func(q string) string { return unknown + q + " len=2" }, nil, true},
		{"bare number", "ab", func(q string) string { return unknown + q + " 2" }, nil, true},
		{"annotation before the quotation", "q", func(q string) string { return "naming (1): unknown octave name: " + q }, nil, true},
		{"annotation equal to the digit of the name", "7", func(q string) string { return unknown + q + "... (7 bytes)" }, nil, true},
		{
			"annotation with creation numbers", "q",
			func(q string) string { return duplicate + q + "... (1 bytes) at 4107 and 5208" }, numbers, true,
		},
		{
			"annotation between creation numbers", "q",
			func(q string) string { return duplicate + "4107 " + q + "... (0 bytes) and 5208" }, numbers, true,
		},
		{"creation number in a lookup", "q", func(q string) string { return unknown + q + " 4107" }, nil, true},
		{"number glued to a creation number", "q", func(q string) string { return duplicate + q + " 41075208" }, numbers, true},
		{"no quotation", "q", func(string) string { return unknown + `"r"` }, nil, true},

		// The accepted limits: these annotations are not found.
		{
			"limit, annotation equal to a creation number", "q",
			func(q string) string { return duplicate + q + "... (4107 bytes) at 4107 and 5208" }, numbers, false,
		},
		{"limit, annotation without digits", "q", func(q string) string { return unknown + q + "..." }, nil, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			quoted := strconv.Quote(tc.nameValue)
			problem := unexpectedNumber(tc.msg(quoted), quoted, tc.allowed)
			assert.Equal(t, tc.want, problem != "", "problem: %q", problem)
		})
	}
}

// A name of up to excerptBytes bytes is quoted in full on all four paths and
// is not followed by its length, including a name of 1 byte and a name of
// exactly excerptBytes bytes: it is the longest short name, not the shortest
// long one. Besides the quotation and, on the creation paths, the ASPN numbers
// of the repeated name, the text has no numbers: see unexpectedNumber. The
// names may contain digits themselves.
func TestErrorTextShortName(t *testing.T) {
	t.Parallel()

	names := []struct {
		name  string
		value string
	}{
		{name: "1 byte", value: "q"},
		{name: "1 byte digit", value: "7"},
		{name: "10 bytes", value: strings.Repeat("q", 10)},
		{name: "31 bytes", value: strings.Repeat("q", 31)},
		{name: "32 bytes", value: strings.Repeat("q", excerptBytes)},
		{name: "32 bytes of digits", value: strings.Repeat("32", 16)},
		{name: "31 bytes Cyrillic", value: "q" + strings.Repeat("ё", 15)},
		{name: "32 bytes Cyrillic", value: strings.Repeat("ё", 16)},
		{name: "32 bytes with a 3-byte rune ending at the boundary", value: strings.Repeat("q", 29) + "♯"},
		{name: "32 bytes with a 4-byte rune ending at the boundary", value: strings.Repeat("q", 28) + "𝄞"},
		{name: "32 bytes of 3-byte runes and letters", value: strings.Repeat("♯", 10) + "qq"},
		{name: "32 bytes of 4-byte runes", value: strings.Repeat("𝄞", 8)},
		{name: "32 invalid bytes", value: strings.Repeat("\xff", excerptBytes)},
		{name: "31 bytes ending with an incomplete sequence", value: strings.Repeat("q", 29) + "\xe2\x82"},
	}

	for _, n := range names {
		for _, p := range errorPaths() {
			t.Run(n.name+"/"+p.name, func(t *testing.T) {
				t.Parallel()

				require.LessOrEqual(t, len(n.value), excerptBytes)

				text := errorText(t, p, n.value)
				quoted := strconv.Quote(n.value)
				require.Contains(t, text, quoted)

				assert.Empty(t, unexpectedNumber(text, quoted, p.numbers), "text: %s", text)
			})
		}
	}
}

// A longer name is quoted as its prefix up to the last unit that fits in
// excerptBytes bytes, followed by its length in bytes, and never in full. The
// units come from decoding the name from its start: a valid rune of 1 to 4
// bytes, or an invalid byte of 1 byte. A multi-byte rune across the boundary
// is not split, so its escaped first byte (`\xd1` for "ё") does not appear;
// a rune that ends at the boundary is kept. An incomplete sequence is a run
// of invalid bytes, so it does not move the boundary back to its first byte.
// The letters a and b are used for the filling: no letter is a digit of the
// length of a name.
func TestErrorTextLongNameExcerpt(t *testing.T) {
	t.Parallel()

	a := func(n int) string { return strings.Repeat("a", n) }
	b := func(n int) string { return strings.Repeat("b", n) }

	cases := []struct {
		name     string
		value    string
		excerpt  int
		absent   string
		wantSize int
	}{
		{
			name:     "33 bytes",
			value:    a(33),
			excerpt:  32,
			wantSize: 33,
		},
		{
			name:     "ASCII far over the boundary",
			value:    a(20) + b(60),
			excerpt:  32,
			wantSize: 80,
		},
		{
			name:     "rune across the boundary",
			value:    a(31) + "ё" + b(10),
			excerpt:  31,
			absent:   `\xd1`,
			wantSize: 43,
		},
		{
			name:     "rune ending at the boundary",
			value:    a(30) + "ё" + b(11),
			excerpt:  32,
			wantSize: 43,
		},
		{
			name:     "3-byte rune ending at the boundary",
			value:    a(29) + "♯" + b(10),
			excerpt:  32,
			wantSize: 42,
		},
		{
			name:     "3-byte rune from byte 30 across the boundary",
			value:    a(30) + "♯" + b(10),
			excerpt:  30,
			absent:   `\xe2`,
			wantSize: 43,
		},
		{
			name:     "3-byte rune from byte 31 across the boundary",
			value:    a(31) + "♯" + b(10),
			excerpt:  31,
			absent:   `\xe2`,
			wantSize: 44,
		},
		{
			name:     "4-byte rune ending at the boundary",
			value:    a(28) + "𝄞" + b(10),
			excerpt:  32,
			wantSize: 42,
		},
		{
			name:     "4-byte rune from byte 29 across the boundary",
			value:    a(29) + "𝄞" + b(10),
			excerpt:  29,
			absent:   `\xf0`,
			wantSize: 43,
		},
		{
			name:     "4-byte rune from byte 30 across the boundary",
			value:    a(30) + "𝄞" + b(10),
			excerpt:  30,
			absent:   `\xf0`,
			wantSize: 44,
		},
		{
			name:     "4-byte rune from byte 31 across the boundary",
			value:    a(31) + "𝄞" + b(10),
			excerpt:  31,
			absent:   `\xf0`,
			wantSize: 45,
		},
		{
			name:     "invalid bytes only",
			value:    strings.Repeat("\xff", 40),
			excerpt:  32,
			wantSize: 40,
		},
		{
			name:     "lone leading byte at the boundary",
			value:    a(31) + "\xd1" + a(10),
			excerpt:  32,
			wantSize: 42,
		},
		{
			name:     "incomplete 3-byte sequence across the boundary",
			value:    a(31) + "\xe2\x82" + b(10),
			excerpt:  32,
			wantSize: 43,
		},
		{
			name:     "incomplete 4-byte sequence across the boundary",
			value:    a(30) + "\xf0\x9d\x84" + b(10),
			excerpt:  32,
			wantSize: 43,
		},
		{
			name:     "lone continuation byte at the boundary",
			value:    a(32) + "\x82" + b(10),
			excerpt:  32,
			wantSize: 43,
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
				assert.NotContains(t, text, strconv.Quote(tc.value), "the name in full")
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
