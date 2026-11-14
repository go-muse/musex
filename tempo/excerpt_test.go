package tempo_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempo/term"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// grammar indexes of the parse paths.
const (
	gTempo = iota
	gApprox
	gRange
	gSeconds
	gIndication
)

// parsePath is one path of architecture 4.14 that quotes the rejected input.
type parsePath struct {
	name    string
	grammar int
	run     func(t *testing.T, x string) error
	base    string // the non-Must path whose error a Must path panics with
}

func parsePaths() []parsePath {
	en := term.EnglishWikipedia()

	return []parsePath{
		{name: "Parse", grammar: gTempo, run: func(t *testing.T, x string) error {
			got, err := tempo.Parse(x)
			assert.Equal(t, tempo.Tempo{}, got)

			return err
		}},
		{name: "MustParse", grammar: gTempo, base: "Parse", run: func(t *testing.T, x string) error {
			return panicError(t, func() { tempo.MustParse(x) })
		}},
		{name: "Tempo.UnmarshalText", grammar: gTempo, run: func(t *testing.T, x string) error {
			target := tq(7)
			err := target.UnmarshalText([]byte(x))
			assert.Equal(t, tq(7), target)

			return err
		}},
		{name: "ParseApprox", grammar: gApprox, run: func(t *testing.T, x string) error {
			got, err := tempo.ParseApprox(x)
			assert.Equal(t, tempo.Approx{}, got)

			return err
		}},
		{name: "MustParseApprox", grammar: gApprox, base: "ParseApprox", run: func(t *testing.T, x string) error {
			return panicError(t, func() { tempo.MustParseApprox(x) })
		}},
		{name: "Approx.UnmarshalText", grammar: gApprox, run: func(t *testing.T, x string) error {
			target := ap(q, fi(7))
			err := target.UnmarshalText([]byte(x))
			assert.Equal(t, ap(q, fi(7)), target)

			return err
		}},
		{name: "ParseRange", grammar: gRange, run: func(t *testing.T, x string) error {
			got, err := tempo.ParseRange(x)
			assert.Equal(t, tempo.Range{}, got)

			return err
		}},
		{name: "MustParseRange", grammar: gRange, base: "ParseRange", run: func(t *testing.T, x string) error {
			return panicError(t, func() { tempo.MustParseRange(x) })
		}},
		{name: "Range.UnmarshalText", grammar: gRange, run: func(t *testing.T, x string) error {
			target := rg(q, fi(1), fi(7))
			err := target.UnmarshalText([]byte(x))
			assert.Equal(t, rg(q, fi(1), fi(7)), target)

			return err
		}},
		{name: "ParseSeconds", grammar: gSeconds, run: func(t *testing.T, x string) error {
			got, err := tempo.ParseSeconds(x)
			assert.Equal(t, tempo.Seconds{}, got)

			return err
		}},
		{name: "MustParseSeconds", grammar: gSeconds, base: "ParseSeconds", run: func(t *testing.T, x string) error {
			return panicError(t, func() { tempo.MustParseSeconds(x) })
		}},
		{name: "Seconds.UnmarshalText", grammar: gSeconds, run: func(t *testing.T, x string) error {
			target := sec(1, 7)
			err := target.UnmarshalText([]byte(x))
			assert.Equal(t, sec(1, 7), target)

			return err
		}},
		{name: "ParseIndication", grammar: gIndication, run: func(t *testing.T, x string) error {
			got, err := tempo.ParseIndication(x, en)
			assert.Equal(t, tempo.Indication{}, got)

			return err
		}},
		{name: "MustParseIndication", grammar: gIndication, base: "ParseIndication", run: func(t *testing.T, x string) error {
			return panicError(t, func() { tempo.MustParseIndication(x, en) })
		}},
	}
}

// grammarClasses are the classes of one input for the five grammars.
type grammarClasses [5]errClass

var invalidEvery = grammarClasses{textClass, textClass, textClass, textClass, textClass}

// excerptInputs are families of long inputs with the same first bytes.
var excerptInputs = []struct {
	name    string
	make    func(n int) string
	classes grammarClasses
}{
	{"bpm digits", func(n int) string { return "4=" + rep("9", n) },
		grammarClasses{fracRange, textClass, textClass, textClass, fracRange}},
	{"approx digits", func(n int) string { return "4≈" + rep("9", n) },
		grammarClasses{textClass, fracRange, textClass, textClass, fracRange}},
	{"range digits", func(n int) string { return "4=1-" + rep("9", n) },
		grammarClasses{textClass, textClass, fracRange, textClass, fracRange}},
	{"seconds digits", func(n int) string { return rep("9", n) + "/1" },
		grammarClasses{textClass, textClass, textClass, fracRange, textClass}},
	{"decimal digits", func(n int) string { return "4=1." + rep("1", n) },
		grammarClasses{fracRange, textClass, textClass, textClass, fracRange}},
	{"grammar at the end", func(n int) string { return "4=" + rep("9", n) + "x" }, invalidEvery},
	{"unit dots", func(n int) string { return "4" + rep(".", n) + "=60" },
		grammarClasses{unitValueErr, textClass, textClass, textClass, unitValueErr}},
	{"letters", func(n int) string { return rep("a", n) }, invalidEvery},
	{"rune at the boundary", func(n int) string { return rep("a", 63) + "ё" + rep("b", n) }, invalidEvery},
	{"two-byte runes", func(n int) string { return rep("ж", n) }, invalidEvery},
	{"four-byte runes", func(n int) string { return rep("😀", n) }, invalidEvery},
	{"invalid bytes", func(n int) string { return rep("\xff", n) }, invalidEvery},
	{"truncated rune at the boundary", func(n int) string { return rep("a", 63) + "\xd0" + rep("b", n) }, invalidEvery},
	{"word and long marking", func(n int) string { return "Allegro, 4=" + rep("9", n) },
		grammarClasses{textClass, textClass, textClass, textClass, fracRange}},
}

// checkErrorText checks the guarantees of architecture 4.14 for the error
// text of a rejected input x: it quotes the excerpt x[:k] by strconv.Quote,
// is at most 512 bytes, and quotes an input longer than 64 bytes only by its
// excerpt.
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
	}
}

// TestParseErrorExcerpt covers architecture 4.14, "Текст ошибок разбора":
// for inputs of 1 KiB and 1 MiB every parse path reports the class of the
// whole input, a bounded text with the excerpt, the same text for both sizes,
// and a Must variant panics with the text of its base path.
func TestParseErrorExcerpt(t *testing.T) {
	t.Parallel()

	for _, in := range excerptInputs {
		t.Run(in.name, func(t *testing.T) {
			t.Parallel()

			small, large := in.make(kib), in.make(mib)
			require.Equal(t, small[:dependentBytes], large[:dependentBytes], "the first 67 bytes are the same")

			texts := map[string]string{}

			for _, p := range parsePaths() {
				class := in.classes[p.grammar]

				var bySize []string

				for _, x := range []string{small, large} {
					err := fast(t, func() error { return p.run(t, x) })
					if !class.check(t, err) {
						t.Logf("%s, input of %d bytes", p.name, len(x))

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

// TestParseErrorShortInput covers architecture 4.14: an input of at most 64
// bytes is quoted in full, special strings included.
func TestParseErrorShortInput(t *testing.T) {
	t.Parallel()

	for _, x := range []string{
		"", "4", "4=", "Tempo(unset)", "Approx(unset)", "Range(unset)", "Seconds(unset)", "4=~108", "4~108", "ё", "\xff",
		"\"quoted", "tab\there", rep("a", 64), rep("ж", 32),
	} {
		for _, p := range parsePaths() {
			err := p.run(t, x)
			if !textClass.check(t, err) {
				t.Logf("%s(%q)", p.name, x)

				continue
			}

			assert.Contains(t, err.Error(), strconv.Quote(x), "%s: %q is quoted in full", p.name, x)
		}
	}

	_, err := term.EnglishWikipedia().Parse("Fast")
	require.Error(t, err)

	_, err = tempo.ParseIndication("Fast, 4=60", term.EnglishWikipedia())
	textClass.check(t, err)
	assert.Contains(t, err.Error(), strconv.Quote("Fast, 4=60"), "the whole text of the indication is quoted")
}

// TestParseErrorNoDependencyText covers architecture 4.3, 4.14: the error
// text contains only the texts of the sentinels of tempo and the excerpt,
// not the texts of the errors of duration.ParseWritten or fraction.
func TestParseErrorNoDependencyText(t *testing.T) {
	t.Parallel()

	in := "4[9223372036854775808:1]=60"
	_, err := tempo.Parse(in)
	unitTupletErr.check(t, err)
	assert.Equal(t, 1, strings.Count(err.Error(), "9223372036854775808"), "the token appears only inside the excerpt")
	assert.NotContains(t, err.Error(), duration.ErrOutOfRange.Error())
	assert.NotContains(t, err.Error(), tuplet.ErrOutOfRange.Error())

	_, err = tempo.Parse("3=60")
	textClass.check(t, err)
	assert.NotContains(t, err.Error(), duration.ErrInvalidText.Error())

	_, err = tempo.Parse("4=" + maxPlusOne())
	fracRange.check(t, err)
	assert.NotContains(t, err.Error(), fraction.ErrOutOfRange.Error())
}
