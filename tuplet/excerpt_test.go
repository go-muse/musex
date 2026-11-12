package tuplet_test

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// Sizes and limits of architecture 4.12 (K-11, G-7).
const (
	kib             = 1024
	mib             = 1024 * 1024
	excerptLimit    = 64
	errorTextLimit  = 512
	maxQuotedLength = 4096
)

// rep is R(u, n) of KR-F9: u repeated to n bytes for a one-byte u, or
// floor(n / len(u)) whole copies for a multi-byte u.
func rep(u string, n int) string {
	return strings.Repeat(u, n/len(u))
}

// excerptLen returns the excerpt boundary k of architecture 4.12, item 3:
// the longest prefix of whole decoding units of x that is at most 64 bytes.
// A unit is a valid UTF-8 rune or a single byte that does not start one. The
// text x itself is decoded, not x[:64].
func excerptLen(x string) int {
	if len(x) <= excerptLimit {
		return len(x)
	}

	k := 0
	for k < len(x) {
		_, size := utf8.DecodeRuneInString(x[k:])
		if k+size > excerptLimit {
			break
		}

		k += size
	}

	return k
}

// panicError runs f, expects a panic with an error value, and returns it.
func panicError(t *testing.T, f func()) error {
	t.Helper()

	value, panicked := capturePanic(f)
	if !assert.True(t, panicked, "expected a panic") {
		return nil
	}

	err, ok := value.(error)
	assert.True(t, ok, "panic value is %T, want error", value)

	return err
}

// textClasses are the expected sentinels of one rejected input for the three
// grammars: Parse (also MustParse and UnmarshalText), ParseShort with Binary,
// and ParseName.
type textClasses struct {
	parse, short, name error
}

// allInvalid is the class of an input rejected by every grammar as text.
var allInvalid = textClasses{parse: tuplet.ErrInvalidText, short: tuplet.ErrInvalidText, name: tuplet.ErrInvalidText}

// parsePath is one path of architecture 4.12 that quotes the rejected input.
type parsePath struct {
	name  string
	class func(textClasses) error
	run   func(t *testing.T, x string) error
	// base names the non-Must path whose error a Must path panics with.
	base string
}

func parsePaths() []parsePath {
	parseClass := func(c textClasses) error { return c.parse }
	shortClass := func(c textClasses) error { return c.short }
	nameClass := func(c textClasses) error { return c.name }

	return []parsePath{
		{
			name: "Parse", class: parseClass,
			run: func(t *testing.T, x string) error {
				got, err := tuplet.Parse(x)
				assert.Equal(t, tuplet.Tuplet{}, got)

				return err
			},
		},
		{
			name: "MustParse", class: parseClass, base: "Parse",
			run: func(t *testing.T, x string) error {
				return panicError(t, func() { tuplet.MustParse(x) })
			},
		},
		{
			name: "UnmarshalText", class: parseClass,
			run: func(t *testing.T, x string) error {
				target := tuplet.Quintuplet()
				err := target.UnmarshalText([]byte(x))
				assert.Equal(t, tuplet.Quintuplet(), target, "the receiver is unchanged")

				return err
			},
		},
		{
			name: "ParseShort", class: shortClass,
			run: func(t *testing.T, x string) error {
				got, err := tuplet.ParseShort(x, tuplet.Binary)
				assert.Equal(t, tuplet.Tuplet{}, got)

				return err
			},
		},
		{
			name: "MustParseShort", class: shortClass, base: "ParseShort",
			run: func(t *testing.T, x string) error {
				return panicError(t, func() { tuplet.MustParseShort(x, tuplet.Binary) })
			},
		},
		{
			name: "ParseName", class: nameClass,
			run: func(t *testing.T, x string) error {
				got, err := tuplet.ParseName(x)
				assert.Equal(t, tuplet.Tuplet{}, got)

				return err
			},
		},
		{
			name: "MustParseName", class: nameClass, base: "ParseName",
			run: func(t *testing.T, x string) error {
				return panicError(t, func() { tuplet.MustParseName(x) })
			},
		},
	}
}

// checkErrorText checks the guarantees of architecture 4.12 for the error of
// a rejected input x: the text quotes the excerpt x[:k], is at most 512
// bytes, and quotes a short input in full. Diagnostics name the case and
// lengths, not the input.
func checkErrorText(t *testing.T, err error, x, label string) {
	t.Helper()

	if err == nil {
		return
	}

	text := err.Error()
	k := excerptLen(x)

	assert.LessOrEqual(t, len(text), errorTextLimit, "%s: error text length, input length %d", label, len(x))
	assert.True(t, strings.Contains(text, strconv.Quote(x[:k])),
		"%s: error text does not quote the excerpt of %d bytes of an input of %d bytes", label, k, len(x))

	if len(x) > excerptLimit && len(x) <= maxQuotedLength {
		assert.False(t, strings.Contains(text, strconv.Quote(x)),
			"%s: an input of %d bytes is quoted in full", label, len(x))
	}
}

// runParsePaths runs every path on x, checks the class and the error text,
// and returns the error texts by path name.
func runParsePaths(t *testing.T, x, label string, classes textClasses) map[string]string {
	t.Helper()

	texts := map[string]string{}

	for _, p := range parsePaths() {
		err := p.run(t, x)
		if !assertErr(t, err, p.class(classes)) {
			t.Logf("%s/%s: input length %d", label, p.name, len(x))

			continue
		}

		checkErrorText(t, err, x, label+"/"+p.name)
		texts[p.name] = err.Error()
	}

	for _, p := range parsePaths() {
		if p.base != "" {
			assert.Equal(t, texts[p.base], texts[p.name], "%s: %s panics with the error of %s", label, p.name, p.base)
		}
	}

	return texts
}

// TestParseErrorText_Long covers K-11 and G-7 on long inputs: the class is
// determined by the whole input, the text is bounded, quotes the excerpt, and
// depends only on the first 67 bytes, so inputs of 128 bytes, 1 KiB and 1 MiB
// of one kind give the same text on each path.
func TestParseErrorText_Long(t *testing.T) {
	t.Parallel()

	type family struct {
		name    string
		build   func(n int) string
		classes textClasses
	}

	families := []family{
		{name: "digits", build: func(n int) string { return rep("9", n) }, classes: textClasses{
			parse: tuplet.ErrInvalidText, short: tuplet.ErrOutOfRange, name: tuplet.ErrInvalidText,
		}},
		{name: "count then digits", build: func(n int) string { return "3:" + rep("9", n-2) }, classes: textClasses{
			parse: tuplet.ErrOutOfRange, short: tuplet.ErrInvalidText, name: tuplet.ErrInvalidText,
		}},
		{name: "digits then count", build: func(n int) string { return rep("9", n-2) + ":2" }, classes: textClasses{
			parse: tuplet.ErrOutOfRange, short: tuplet.ErrInvalidText, name: tuplet.ErrInvalidText,
		}},
	}

	for _, u := range []string{"x", "\x00", "\x01", "é", "€", "😀", "\xff", "\x80"} {
		families = append(families, family{
			name:    "fill " + strconv.QuoteToASCII(u),
			build:   func(n int) string { return rep(u, n) },
			classes: allInvalid,
		})
	}

	for _, f := range families {
		inputs := []string{f.build(128), f.build(kib), f.build(mib)}

		t.Run(f.name, func(t *testing.T) {
			t.Parallel()

			var reference map[string]string

			for i, x := range inputs {
				label := f.name + "/" + itoa(len(x))
				texts := runParsePaths(t, x, label, f.classes)

				if i == 0 {
					reference = texts

					continue
				}

				for path, text := range texts {
					assert.Equal(t, reference[path], text,
						"%s/%s: the text depends only on the first 67 bytes", label, path)
				}
			}
		})
	}
}

// TestParseErrorText_Boundary covers architecture 4.12, item 3, with the
// "Excerpt boundary" set of KR-F9 and the limit of 64 bytes.
func TestParseErrorText_Boundary(t *testing.T) {
	t.Parallel()

	tail := rep("x", 10)

	tests := []struct {
		name string
		x    string
		k    int
	}{
		{name: "two-byte rune ends at 64", x: rep("x", 62) + "é" + tail, k: 64},
		{name: "three-byte rune ends at 64", x: rep("x", 61) + "€" + tail, k: 64},
		{name: "four-byte rune ends at 64", x: rep("x", 60) + "😀" + tail, k: 64},
		{name: "two-byte rune crosses 64", x: rep("x", 63) + "é" + tail, k: 63},
		{name: "two-byte rune crosses 64 without tail", x: rep("x", 63) + "é", k: 63},
		{name: "three-byte rune crosses 64", x: rep("x", 62) + "€" + tail, k: 62},
		{name: "four-byte rune crosses 64", x: rep("x", 61) + "😀" + tail, k: 61},
		{name: "invalid byte at 64", x: rep("x", 63) + "\xff" + tail, k: 64},
		{name: "invalid byte after 64", x: rep("x", 64) + "\xff" + tail, k: 64},
		{name: "lead byte without continuation", x: rep("x", 63) + "\xd1" + tail, k: 64},
		{name: "two lead bytes", x: rep("x", 63) + "\xe0\xe0" + tail, k: 64},
		{name: "truncated three-byte sequence", x: rep("x", 63) + "\xe2\x82" + tail, k: 64},
		{name: "lone continuation byte", x: rep("x", 63) + "\x80" + tail, k: 64},
		{name: "continuation bytes only", x: rep("\x80", 100), k: 64},
		{name: "65 ASCII bytes", x: rep("x", 65), k: 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.k, excerptLen(tt.x), "the test oracle agrees with the expected boundary")
			assert.Positive(t, tt.k, "the excerpt is not empty")

			runParsePaths(t, tt.x, tt.name, allInvalid)
		})
	}
}

// TestParseErrorText_Short covers architecture 4.12, item 1: an input of at
// most 64 bytes is quoted in full on every path.
func TestParseErrorText_Short(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		x       string
		classes textClasses
	}{
		{name: "empty", x: "", classes: allInvalid},
		{name: "one letter", x: "a", classes: allInvalid},
		{name: "63 letters", x: rep("x", 63), classes: allInvalid},
		{name: "64 letters", x: rep("x", 64), classes: allInvalid},
		{name: "17 zero bytes", x: rep("\x00", 17), classes: allInvalid},
		{name: "64 zero bytes", x: rep("\x00", 64), classes: allInvalid},
		{name: "64 continuation bytes", x: rep("\x80", 64), classes: allInvalid},
		{name: "32 two-byte runes", x: rep("é", 64), classes: allInvalid},
		{name: "digits then letter", x: rep("9", 63) + "x", classes: allInvalid},
		{name: "letter then invalid bytes", x: "x" + rep("\xff", 63), classes: allInvalid},
		{name: "three counts", x: "3:2:1", classes: allInvalid},
		{name: "zero count", x: "0:2", classes: allInvalid},
		{name: "capitalized name", x: "Triplet", classes: allInvalid},
		{name: "russian duplet", x: "дуплет", classes: allInvalid},
		{name: "unset string", x: "Tuplet(unset)", classes: allInvalid},
		{name: "pair overflow", x: "9223372036854775808:2", classes: textClasses{
			parse: tuplet.ErrOutOfRange, short: tuplet.ErrInvalidText, name: tuplet.ErrInvalidText,
		}},
		{name: "count overflow", x: "9223372036854775808", classes: textClasses{
			parse: tuplet.ErrInvalidText, short: tuplet.ErrOutOfRange, name: tuplet.ErrInvalidText,
		}},
		{name: "64 digits", x: rep("9", 64), classes: textClasses{
			parse: tuplet.ErrInvalidText, short: tuplet.ErrOutOfRange, name: tuplet.ErrInvalidText,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runParsePaths(t, tt.x, tt.name, tt.classes)
		})
	}
}

// TestParseErrorText_NoDefault covers architecture 4.12: E-3 of ParseShort
// and MustParseShort quotes the input.
func TestParseErrorText_NoDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		x        string
		division tuplet.Division
	}{
		{name: "one", x: "1", division: tuplet.Binary},
		{name: "binary member", x: "4", division: tuplet.Binary},
		{name: "ternary member", x: "12", division: tuplet.Ternary},
		{name: "unknown division", x: "3", division: tuplet.Division(0)},
		{name: "large binary member", x: "4611686018427387904", division: tuplet.Binary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tuplet.ParseShort(tt.x, tt.division)
			assertErr(t, err, tuplet.ErrNoDefault)
			checkErrorText(t, err, tt.x, tt.name)

			panicked := panicError(t, func() { tuplet.MustParseShort(tt.x, tt.division) })
			assertErr(t, panicked, tuplet.ErrNoDefault)
			checkErrorText(t, panicked, tt.x, tt.name+"/Must")

			if err != nil && panicked != nil {
				assert.Equal(t, err.Error(), panicked.Error())
			}
		})
	}
}

// TestParseErrorText_NilReceiver covers architecture 4.12: E-6 of
// UnmarshalText with a nil receiver does not quote the input, so its text is
// the same for any input.
func TestParseErrorText_NilReceiver(t *testing.T) {
	t.Parallel()

	var target *tuplet.Tuplet

	inputs := []string{"3:2", "", rep("x", mib)}
	texts := make([]string, 0, len(inputs))

	for _, x := range inputs {
		err := target.UnmarshalText([]byte(x))
		if assertErr(t, err, tuplet.ErrInvalidArgument) {
			texts = append(texts, err.Error())
			assert.LessOrEqual(t, len(err.Error()), errorTextLimit)
		}
	}

	for _, text := range texts {
		assert.Equal(t, texts[0], text, "the text does not depend on the input")
	}

	if len(texts) > 0 {
		assert.False(t, strings.Contains(texts[0], strconv.Quote("3:2")), "the input is not quoted")
	}
}
