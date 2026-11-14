package term_test

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// checkErrorText checks the guarantees of architecture 4.14 (the rule of
// duration 4.12) for the error text of a rejected input x: it quotes the
// excerpt x[:k] by strconv.Quote, is at most 512 bytes, and quotes an input
// longer than 64 bytes only by its excerpt.
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

// TestParseErrorText covers architecture 4.14, "Текст ошибок разбора", for
// term.List.Parse and its Must variant: short names are quoted in full, long
// names by an excerpt; the text is bounded and depends only on the first 67
// bytes; a Must variant panics with the same text; an unset list (Е-1) has
// no quotation requirement.
func TestParseErrorText(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()

	for _, x := range []string{
		"", "Fast", "List(unset)", "ё", "\xff", "\"quoted\"", "tab\there", rep("a", 64), rep("ж", 32), " Allegro",
	} {
		_, err := en.Parse(x)
		if !assertErr(t, err, term.ErrInvalidText) {
			continue
		}

		assert.Contains(t, err.Error(), strconv.Quote(x), "%q is quoted in full", x)
		checkErrorText(t, err, x, "Parse")
	}

	families := map[string]func(n int) string{
		"letters":              func(n int) string { return rep("a", n) },
		"name prefix":          func(n int) string { return "Allegro" + rep(" ", n) },
		"rune at the boundary": func(n int) string { return rep("a", 63) + "ё" + rep("b", n) },
		"two-byte runes":       func(n int) string { return rep("ж", n) },
		"four-byte runes":      func(n int) string { return rep("😀", n) },
		"invalid bytes":        func(n int) string { return rep("\xff", n) },
		"truncated rune":       func(n int) string { return rep("a", 63) + "\xd0" + rep("b", n) },
	}

	for name, gen := range families {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			small, large := gen(kib), gen(mib)
			require.Equal(t, small[:dependentBytes], large[:dependentBytes])

			var texts []string

			for _, x := range []string{small, large} {
				gotErr := fast(t, func() error { _, e := en.Parse(x); return e })

				if !assertErr(t, gotErr, term.ErrInvalidText) {
					continue
				}

				checkErrorText(t, gotErr, x, "Parse")
				texts = append(texts, gotErr.Error())

				mustErr := panicError(t, func() { en.MustParse(x) })
				if assert.Error(t, mustErr) {
					assert.Equal(t, gotErr.Error(), mustErr.Error(), "MustParse panics with the error of Parse")
				}
			}

			if len(texts) == 2 {
				assert.Equal(t, texts[0], texts[1], "the text depends only on the first 67 bytes")
			}
		})
	}
}

// FuzzListParse covers architecture 4.4, 4.14: List.Parse never panics; a
// found word has a name equal to the input ignoring case (or the synonym Vivo
// of list 1), and parsing its name gives it back; a failure is exactly
// term.ErrInvalidText with a bounded text.
func FuzzListParse(f *testing.F) {
	for _, s := range []string{
		"Allegro", "allegro", "Vivo", "vivo", "", " ", "Fast", "\xff", "Скоро", "СКОРО", "4=60", "\"x\"",
		"Molto Allegro", "prestissimo", "Ё", "ё",
	} {
		f.Add(s)
	}

	mine := term.MustNewList(
		term.Row{Entries: []term.Entry{{Name: "Скоро"}, {Name: "4=60"}, {Name: "\"x\""}}},
		term.Row{Entries: []term.Entry{{Name: " "}, {Name: "ё"}}},
	)
	lists := append(term.BuiltIn(), mine)

	f.Fuzz(func(t *testing.T, x string) {
		for i, l := range lists {
			w, err := l.Parse(x)
			if err != nil {
				assertErr(t, err, term.ErrInvalidText)
				assert.Equal(t, term.Term{}, w)
				assert.LessOrEqual(t, len(err.Error()), errorTextLimit)

				continue
			}

			require.True(t, utf8.ValidString(x))

			name := w.MustName()
			assert.True(t, strings.EqualFold(name, x) || (i == 0 && strings.EqualFold(x, "Vivo") && name == "Vivace"),
				"Parse(%q) gave %q", x, name)
			assert.Equal(t, w, l.MustParse(name))
		}
	})
}
