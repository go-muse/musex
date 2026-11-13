package duration_test

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// Sizes and limits of the error text of parsing (architecture 4.12).
const (
	kib            = 1024
	mib            = 1024 * 1024
	excerptLimit   = 64
	dependentBytes = 67
	errorTextLimit = 512
)

// excerptLen returns k of architecture 4.12: the longest prefix of whole
// UTF-8 decoding units of x that is at most 64 bytes. A unit is a valid rune
// or a single byte that does not start one. The text x itself is decoded.
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

// errClass is the class of an error: the package sentinel and the wrapped
// dependency sentinels.
type errClass struct {
	want   error
	causes []error
}

var (
	textClass     = errClass{want: duration.ErrInvalidText}
	fracRange     = errClass{want: duration.ErrOutOfRange, causes: []error{fraction.ErrOutOfRange}}
	tupletRange   = errClass{want: duration.ErrOutOfRange, causes: []error{tuplet.ErrOutOfRange}}
	chainRange    = errClass{want: duration.ErrOutOfRange, causes: []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}}
	invalidEvery  = grammarClasses{textClass, textClass, textClass, textClass, textClass}
	grammarNames  = []string{"duration", "position", "written", "tied", "base"}
	excerptInputs = []excerptInput{
		{"fraction denominator", func(n int) string { return "1/" + rep("9", n) },
			grammarClasses{fracRange, fracRange, textClass, textClass, textClass}},
		{"fraction numerator", func(n int) string { return rep("9", n) + "/1" },
			grammarClasses{fracRange, fracRange, textClass, textClass, textClass}},
		{"fraction grammar at the end", func(n int) string { return "1/" + rep("9", n) + "x" }, invalidEvery},
		{"tuplet number", func(n int) string { return "4[3:" + rep("9", n) + "]" },
			grammarClasses{textClass, textClass, tupletRange, textClass, textClass}},
		{"chain coefficient", func(n int) string { return "4[" + rep("3:2,", n) + "3:2]" },
			grammarClasses{textClass, textClass, chainRange, textClass, textClass}},
		{"dots", func(n int) string { return "4" + rep(".", n) },
			grammarClasses{textClass, textClass, fracRange, textClass, textClass}},
		{"dots grammar at the end", func(n int) string { return "4" + rep(".", n) + "x" }, invalidEvery},
		{"tie part", func(n int) string { return rep("4~", n) + "4" + strings.Repeat(".", 300) },
			grammarClasses{textClass, textClass, textClass, fracRange, textClass}},
		{"tie tuplet number", func(n int) string { return rep("4~", n) + "4[" + rep("9", 30) + ":1]" },
			grammarClasses{textClass, textClass, textClass, tupletRange, textClass}},
		{"tie grammar at the end", func(n int) string { return rep("4~", n) + "3" }, invalidEvery},
		{"letters", func(n int) string { return rep("a", n) }, invalidEvery},
		{"name and spaces", func(n int) string { return "quarter note" + rep(" ", n) }, invalidEvery},
		{"rune at the boundary", func(n int) string { return rep("a", 63) + "ё" + rep("b", n) }, invalidEvery},
		{"two-byte runes", func(n int) string { return rep("ж", n) }, invalidEvery},
		{"three-byte runes", func(n int) string { return rep("€", n) }, invalidEvery},
		{"four-byte runes", func(n int) string { return rep("😀", n) }, invalidEvery},
		{"invalid bytes", func(n int) string { return rep("\xff", n) }, invalidEvery},
		{"truncated rune at the boundary", func(n int) string { return rep("a", 63) + "\xd0" + rep("b", n) }, invalidEvery},
	}
)

// grammarClasses are the classes of one input for the five grammars.
type grammarClasses [5]errClass

// excerptInput is a family of long inputs with the same first bytes.
type excerptInput struct {
	name    string
	make    func(n int) string
	classes grammarClasses
}

// parsePath is one path of architecture 4.12 that quotes the rejected input.
type parsePath struct {
	name    string
	grammar int // index in grammarNames
	run     func(t *testing.T, x string) error
	base    string // the non-Must path whose error a Must path panics with
}

func parsePaths() []parsePath {
	return []parsePath{
		{name: "Parse", grammar: 0, run: func(t *testing.T, x string) error {
			got, err := duration.Parse(x)
			assert.Equal(t, duration.Duration{}, got)

			return err
		}},
		{name: "MustParse", grammar: 0, base: "Parse", run: func(t *testing.T, x string) error {
			return panicError(t, func() { duration.MustParse(x) })
		}},
		{name: "Duration.UnmarshalText", grammar: 0, run: func(t *testing.T, x string) error {
			target := dur(5, 7)
			err := target.UnmarshalText([]byte(x))
			assert.Equal(t, dur(5, 7), target)

			return err
		}},
		{name: "ParsePosition", grammar: 1, run: func(t *testing.T, x string) error {
			got, err := duration.ParsePosition(x)
			assert.Equal(t, duration.Position{}, got)

			return err
		}},
		{name: "MustParsePosition", grammar: 1, base: "ParsePosition", run: func(t *testing.T, x string) error {
			return panicError(t, func() { duration.MustParsePosition(x) })
		}},
		{name: "Position.UnmarshalText", grammar: 1, run: func(t *testing.T, x string) error {
			target := pos(5, 7)
			err := target.UnmarshalText([]byte(x))
			assert.Equal(t, pos(5, 7), target)

			return err
		}},
		{name: "ParseWritten", grammar: 2, run: func(t *testing.T, x string) error {
			got, err := duration.ParseWritten(x)
			assert.Equal(t, duration.Written{}, got)

			return err
		}},
		{name: "MustParseWritten", grammar: 2, base: "ParseWritten", run: func(t *testing.T, x string) error {
			return panicError(t, func() { duration.MustParseWritten(x) })
		}},
		{name: "Written.UnmarshalText", grammar: 2, run: func(t *testing.T, x string) error {
			target := wr(duration.Half, 1)
			err := target.UnmarshalText([]byte(x))
			assert.Equal(t, wr(duration.Half, 1), target)

			return err
		}},
		{name: "ParseTied", grammar: 3, run: func(t *testing.T, x string) error {
			got, err := duration.ParseTied(x)
			assert.Equal(t, duration.Tied{}, got)

			return err
		}},
		{name: "MustParseTied", grammar: 3, base: "ParseTied", run: func(t *testing.T, x string) error {
			return panicError(t, func() { duration.MustParseTied(x) })
		}},
		{name: "Tied.UnmarshalText", grammar: 3, run: func(t *testing.T, x string) error {
			target := tie(wr(duration.Half, 0), wr(duration.Half, 0))
			err := target.UnmarshalText([]byte(x))
			assert.Equal(t, tie(wr(duration.Half, 0), wr(duration.Half, 0)), target)

			return err
		}},
		{name: "ParseBase", grammar: 4, run: func(t *testing.T, x string) error {
			got, err := duration.ParseBase(x)
			assert.Equal(t, duration.Base(0), got)

			return err
		}},
		{name: "MustParseBase", grammar: 4, base: "ParseBase", run: func(t *testing.T, x string) error {
			return panicError(t, func() { duration.MustParseBase(x) })
		}},
		{name: "Base.UnmarshalText", grammar: 4, run: func(t *testing.T, x string) error {
			target := duration.Longa
			err := target.UnmarshalText([]byte(x))
			assert.Equal(t, duration.Longa, target)

			return err
		}},
	}
}

// checkErrorText checks the guarantees of architecture 4.12 for the error
// text of a rejected input x: it quotes the excerpt x[:k], is at most 512
// bytes, and quotes an input longer than 64 bytes only by its excerpt.
// Diagnostics name lengths, not the input.
func checkErrorText(t *testing.T, err error, x, label string) {
	t.Helper()

	if err == nil {
		return
	}

	text := err.Error()
	k := excerptLen(x)

	assert.LessOrEqual(t, len(text), errorTextLimit, "%s: error text length, input length %d", label, len(x))
	assert.True(t, strings.Contains(text, strconv.Quote(x[:k])),
		"%s: the error text does not quote the excerpt of %d bytes of an input of %d bytes", label, k, len(x))

	if len(x) > excerptLimit {
		assert.False(t, strings.Contains(text, strconv.Quote(x)), "%s: an input of %d bytes is quoted in full", label, len(x))
		assert.False(t, strings.Contains(text, strconv.Itoa(len(x))), "%s: the length of the input is not in the text", label)
	}
}

// TestParseErrorExcerpt covers architecture 4.12: for long inputs of 1 KiB
// and 1 MiB, every parse path reports the class determined by the whole
// input, a bounded text with the excerpt, the same text for both sizes, and
// a Must variant panics with the text of its base path.
func TestParseErrorExcerpt(t *testing.T) {
	t.Parallel()

	for _, in := range excerptInputs {
		t.Run(in.name, func(t *testing.T) {
			t.Parallel()

			small, large := in.make(kib), in.make(mib)
			if !assert.Equal(t, small[:dependentBytes], large[:dependentBytes], "the first 67 bytes are the same") {
				return
			}

			texts := map[string]string{}

			for _, p := range parsePaths() {
				class := in.classes[p.grammar]

				var bySize []string

				for _, x := range []string{small, large} {
					err := p.run(t, x)
					if !assertErr(t, err, class.want, class.causes...) {
						t.Logf("%s, %s grammar, input of %d bytes", p.name, grammarNames[p.grammar], len(x))

						continue
					}

					checkErrorText(t, err, x, p.name)
					bySize = append(bySize, err.Error())
				}

				if len(bySize) == 2 {
					assert.Equal(t, bySize[0], bySize[1], "%s: the text depends only on the first 67 bytes", p.name)
					texts[p.name] = bySize[0]
				}
			}

			for _, p := range parsePaths() {
				if p.base != "" {
					assert.Equal(t, texts[p.base], texts[p.name], "%s panics with the error of %s", p.name, p.base)
				}
			}
		})
	}
}

// TestParseErrorShortInput covers architecture 4.12: an input of at most 64
// bytes is quoted in full, special strings included.
func TestParseErrorShortInput(t *testing.T) {
	t.Parallel()

	for _, x := range []string{
		"", "3", "4[]", "Duration(unset)", "Position(unset)", "Written(unset)", "Tied(unset)", "Base(invalid)",
		"ё", "\xff", "\"quoted\"", "tab\there", rep("a", 64), rep("ж", 32),
	} {
		for _, p := range parsePaths() {
			err := p.run(t, x)
			if !assertErr(t, err, duration.ErrInvalidText) {
				continue
			}

			assert.True(t, strings.Contains(err.Error(), strconv.Quote(x)), "%s: %q is quoted in full", p.name, x)
		}
	}
}

// TestParseErrorNoToken covers architecture 4.12: the error text does not
// quote a token apart from the excerpt, nor include texts of errors of the
// dependencies' parsers.
func TestParseErrorNoToken(t *testing.T) {
	t.Parallel()

	token := rep("9", 100)

	_, err := duration.Parse("1/" + token)
	assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
	assert.NotContains(t, err.Error(), strconv.Quote(token))
	assert.NotContains(t, err.Error(), "100")

	_, err = duration.ParseWritten("4[3:" + token + "]")
	assertErr(t, err, duration.ErrOutOfRange, tuplet.ErrOutOfRange)
	assert.NotContains(t, err.Error(), strconv.Quote(token))
	assert.NotContains(t, err.Error(), strconv.Quote("3:"+token))

	// The excerpt of a short input is the whole input; the token alone is
	// not quoted again.
	_, err = duration.ParseWritten("4[3:" + rep("9", 30) + "]")
	assertErr(t, err, duration.ErrOutOfRange, tuplet.ErrOutOfRange)
	assert.Equal(t, 1, strings.Count(err.Error(), rep("9", 30)), "the token appears only inside the excerpt")
}
