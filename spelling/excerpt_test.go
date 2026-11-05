package spelling_test

import (
	jsonv2 "encoding/json/v2"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the bounded excerpt of the input in reading errors (architecture
// of task 3.7, KR-S15, properties T1-T4). Long inputs reach 1 MiB, so no
// assertion prints the text of an error of a long input: the length of the
// text is checked first with require, and substrings are checked with short
// messages.

// errTextLimit is the bound T1 on the length of the text of a reading error.
const errTextLimit = 512

// The lengths of long inputs: 1 KiB and 1 MiB.
var longSizes = []struct {
	name string
	n    int
}{
	{"1KiB", 1 << 10},
	{"1MiB", 1 << 20},
}

// repeatTo is R(c, n) of the architecture: c repeated to at most n bytes,
// with whole copies of c only.
func repeatTo(c string, n int) string {
	return strings.Repeat(c, n/len(c))
}

// excerptPath is a reading path that takes the quoted value x itself:
// P1 Parse, P2 MustParse, P3 UnmarshalText, P4 UnmarshalJSON of a valid JSON
// string with the content x. read returns the reading error and checks the
// value that the failed reading leaves.
type excerptPath struct {
	name string
	// validUTF8Only reports that the path takes only valid UTF-8: a valid
	// JSON string cannot hold invalid UTF-8.
	validUTF8Only bool
	read          func(t *testing.T, x string) error
}

func excerptPaths() []excerptPath {
	return []excerptPath{
		{
			name: "P1 Parse",
			read: func(t *testing.T, x string) error {
				t.Helper()

				got, err := spelling.Parse(x)
				assert.True(t, got == spelling.Spelling{}, "Parse must return the unset name")

				return err
			},
		},
		{
			name: "P2 MustParse",
			read: func(t *testing.T, x string) error {
				t.Helper()

				value, panicked := catchPanic(func() { _ = spelling.MustParse(x) })
				require.True(t, panicked, "MustParse must panic")

				err, ok := value.(error)
				require.True(t, ok, "the panic value must be an error")

				return err
			},
		},
		{
			name: "P3 UnmarshalText",
			read: func(t *testing.T, x string) error {
				t.Helper()

				receiver := spelling.DFlat()
				err := receiver.UnmarshalText([]byte(x))
				assert.Equal(t, spelling.DFlat(), receiver, "the receiver must keep its value")

				return err
			},
		},
		{
			name:          "P4 UnmarshalJSON string",
			validUTF8Only: true,
			read: func(t *testing.T, x string) error {
				t.Helper()

				data, err := jsonv2.Marshal(x)
				require.NoError(t, err, "encoding the JSON string")

				receiver := spelling.DFlat()
				err = receiver.UnmarshalJSON(data)
				assert.Equal(t, spelling.DFlat(), receiver, "the receiver must keep its value")

				return err
			},
		},
	}
}

// readJSON is P5 and P6: UnmarshalJSON of data that is not a valid JSON
// string or not a string at all; the quoted value is the whole data.
func readJSON(t *testing.T, data string) error {
	t.Helper()

	receiver := spelling.DFlat()
	err := receiver.UnmarshalJSON([]byte(data))
	assert.Equal(t, spelling.DFlat(), receiver, "the receiver must keep its value")

	return err
}

// requireErrorText checks T1 and returns the error text. It prints only
// lengths, never the text.
func requireErrorText(t *testing.T, err error, inputLen int) string {
	t.Helper()

	require.Error(t, err)

	msg := err.Error()
	require.LessOrEqualf(t, len(msg), errTextLimit,
		"T1: length of the error text for an input of %d bytes", inputLen)

	return msg
}

// assertLongInputError checks T1 and T3 for an input x longer than 32 bytes:
// the text quotes the excerpt x[:k] and gives the length of x, and does not
// quote the whole x. k is a literal of the architecture table.
func assertLongInputError(t *testing.T, err error, x string, k int) {
	t.Helper()

	msg := requireErrorText(t, err, len(x))

	excerpt := strconv.Quote(x[:k])
	assert.Truef(t, strings.Contains(msg, excerpt), "T3: the error text must contain the excerpt %s", excerpt)
	assert.Truef(t, strings.Contains(msg, strconv.Itoa(len(x))),
		"T3: the error text must contain the input length %d", len(x))

	if len(x) <= errTextLimit {
		assert.Falsef(t, strings.Contains(msg, strconv.Quote(x)),
			"T3: the error text must not quote the whole input of %d bytes", len(x))
	}
}

// hasLengthAnnotation reports whether the text msg of the error for the value
// x quoted as quoted annotates x with a number right after the quotation, as
// the text for a longer input does: "<excerpt>"... (<n> bytes). It inspects
// only the text from the first occurrence of quoted up to the next ": " that
// starts the secondary error or the nested JSON diagnostic, or up to the end
// of the text. That part is the own text of the package: the prefix before the
// quotation has no quote character, so the first occurrence of quoted is the
// quotation of the package, and the part after it holds no user data. Any
// decimal digit there is an annotation, whatever the number is. The digits of
// the quotation itself and the text after ": " are not inspected: the nested
// diagnostic quotes user data, and its format is not part of the contract.
func hasLengthAnnotation(msg, quoted string) bool {
	i := strings.Index(msg, quoted)
	if i < 0 {
		return false
	}

	tail := msg[i+len(quoted):]
	if j := strings.Index(tail, ": "); j >= 0 {
		tail = tail[:j]
	}

	return strings.ContainsAny(tail, "0123456789")
}

// assertShortInputError checks T1, T2 and T2a for an input x of at most 32
// bytes: the text quotes the whole x and, unlike an excerpt, does not
// annotate it with its length. Digits of the quotation itself (a digit in x,
// or an escape such as \x01) and of the text after the first ": " are not an
// annotation: see hasLengthAnnotation.
func assertShortInputError(t *testing.T, err error, x string) {
	t.Helper()

	require.LessOrEqual(t, len(x), 32, "a short input")

	msg := requireErrorText(t, err, len(x))
	quoted := strconv.Quote(x)
	require.Truef(t, strings.Contains(msg, quoted), "T2: the error text must quote the whole input %s", quoted)

	assert.Falsef(t, hasLengthAnnotation(msg, quoted),
		"T2a: the text right after the quotation of a short input of %d bytes must have no digits: %s",
		len(x), msg)
}

// The detector of the annotation tells the annotation of a mutant from the
// digits of a quotation or of a nested diagnostic. The texts are
// illustrations made up for the detector, not the format of the package.
func TestHasLengthAnnotation(t *testing.T) {
	t.Parallel()

	const prefix = "spelling: invalid machine form "

	tests := []struct {
		name string
		x    string
		tail string
		want bool
	}{
		{"annotation with bytes", "C#x", " (3 bytes)", true},
		{"annotation with an ellipsis", "ab", "... (2 bytes)", true},
		{"annotation in parentheses", "ab", " (2)", true},
		{"annotation in brackets", "ab", " [2 bytes]", true},
		{"annotation without parentheses", "ab", " 2 bytes", true},
		{"annotation as a key", "ab", " len=2", true},
		{"bare number", "ab", " 2", true},
		{"other number", "ab", " (99 bytes)", true},
		{"annotation before the secondary error", "ab", " (2 bytes): alteration: value out of range", true},
		{"annotation before a nested diagnostic", "ab", " (2 bytes): nested: invalid character", true},
		{"no annotation", "ab", "", false},
		{"own text without digits", "ab", " is not a string or null", false},
		{"digit in the quotation", "a2", "", false},
		{"digit equal to the length in the quotation", "a2", " is not a string or null", false},
		{"escape in the quotation", "\x01", "", false},
		{"escape in the quotation before own text", "\x01", " is not a string or null", false},
		{"nested diagnostic with an escape", "\"\x03\"", `: nested: invalid character '\x03' in string`, false},
		{"nested diagnostic with a digit character", "\"\\4\"", `: nested: invalid escape '4'`, false},
		{"nested diagnostic with a number in parentheses", "\"\\u(7)\"", `: nested: invalid escape (7)`, false},
		{"nested diagnostic with a number and a word in parentheses", "ab", `: nested: invalid surrogate (14 a)`, false},
		{"secondary error with a number", "ab", ": alteration: value 2 out of range", false},
		// The accepted limitation: an annotation after the first ": " is not
		// found. It is not a defect of the detector.
		{"annotation after the secondary error", "ab", ": alteration: value out of range (2 bytes)", false},
		{"annotation after a nested diagnostic", "ab", ": nested: invalid character (2 bytes)", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			quoted := strconv.Quote(tc.x)
			assert.Equal(t, tc.want, hasLengthAnnotation(prefix+quoted+tc.tail, quoted))
		})
	}
}

// T1, T3, T4 on P1-P4 for long uniform inputs of 1 KiB and 1 MiB.
func TestReadErrorExcerptLongInput(t *testing.T) {
	t.Parallel()

	families := []struct {
		name string
		c    string
		k    int
	}{
		{"zero bytes", "\x00", 32},
		{"x", "x", 32},
		{"e acute", "é", 32},
		{"euro sign", "€", 30},
		{"invalid byte 0xff", "\xff", 32},
	}

	for _, path := range excerptPaths() {
		for _, family := range families {
			if path.validUTF8Only && family.c == "\xff" {
				continue
			}

			for _, size := range longSizes {
				t.Run(path.name+"/"+family.name+"/"+size.name, func(t *testing.T) {
					t.Parallel()

					x := repeatTo(family.c, size.n)
					err := path.read(t, x)

					assertLongInputError(t, err, x, family.k)
					assertInvalidForm(t, formOnly, err)
				})
			}
		}
	}
}

// T1, T3 and the classes T4 of the table "Reading errors" on long inputs of
// the number shape and after a valid form, on P1-P4.
func TestReadErrorExcerptLongInputClasses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input func(n int) string
		class formErrorClass
	}{
		{"long numerator", func(n int) string { return "C+" + repeatTo("9", n) }, formOutOfRange},
		{"long numerator over zero", func(n int) string { return "C+" + repeatTo("9", n) + "/0" }, formZeroDenominator},
		{"long denominator", func(n int) string { return "C+1/" + repeatTo("9", n) }, formOutOfRange},
		{"form and a long tail", func(n int) string { return "C+1" + repeatTo("x", n) }, formOnly},
	}

	for _, path := range excerptPaths() {
		for _, tc := range cases {
			for _, size := range longSizes {
				t.Run(path.name+"/"+tc.name+"/"+size.name, func(t *testing.T) {
					t.Parallel()

					x := tc.input(size.n)
					err := path.read(t, x)

					assertLongInputError(t, err, x, 32)
					assertInvalidForm(t, tc.class, err)
				})
			}
		}
	}
}

// T2 and T3 at the 32-byte boundary: whole runes only, an invalid byte is a
// one-byte unit. k is a literal of the architecture table.
func TestReadErrorExcerptBoundary(t *testing.T) {
	t.Parallel()

	x := func(n int) string { return strings.Repeat("x", n) }

	cases := []struct {
		name  string
		input string
		// k is the excerpt length; 0 means a short input quoted whole.
		k           int
		invalidUTF8 bool
	}{
		{"32 bytes", x(32), 0, false},
		{"33 bytes", x(33), 32, false},
		{"3-byte rune ends at byte 32", x(29) + "€" + x(10), 32, false},
		{"2-byte rune across the boundary", x(31) + "é" + x(10), 31, false},
		{"3-byte rune across the boundary", x(30) + "€" + x(10), 30, false},
		{"4-byte rune across the boundary", x(29) + "😀" + x(10), 29, false},
		{"invalid byte at byte 32", x(31) + "\xff" + x(10), 32, true},
		{"invalid byte after the boundary", x(32) + "\xff" + x(10), 32, true},
		{"truncated sequence across the boundary", x(31) + "\xe2\x82" + x(10), 32, true},
		{"lone continuation byte at byte 32", x(31) + "\x80" + x(10), 32, true},
	}

	for _, path := range excerptPaths() {
		for _, tc := range cases {
			if path.validUTF8Only && tc.invalidUTF8 {
				continue
			}

			t.Run(path.name+"/"+tc.name, func(t *testing.T) {
				t.Parallel()

				err := path.read(t, tc.input)

				if tc.k == 0 {
					assertShortInputError(t, err, tc.input)
				} else {
					assertLongInputError(t, err, tc.input, tc.k)
				}

				assertInvalidForm(t, formOnly, err)
			})
		}
	}
}

// T1, T3, T4 on P5 (an invalid JSON string) and P6 (not a string and not
// null) for long data of 1 KiB and 1 MiB; the whole data is quoted.
func TestUnmarshalJSONExcerptLongInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		data func(n int) string
	}{
		{"P5 no closing quote", func(n int) string { return `"` + repeatTo("x", n) }},
		{"P5 tail after a string", func(n int) string { return `"C+1"` + repeatTo("x", n) }},
		{"P5 invalid escape", func(n int) string { return `"\q` + repeatTo("x", n) + `"` }},
		{"P5 raw control bytes", func(n int) string { return `"` + repeatTo("\x00", n) + `"` }},
		{"P5 invalid UTF-8", func(n int) string { return `"` + repeatTo("\xff", n) + `"` }},
		{"P6 number", func(n int) string { return repeatTo("1", n) }},
		{"P6 array", func(n int) string { return "[" + repeatTo("1,", n) + "1]" }},
		{"P6 object", func(n int) string { return `{"a":` + repeatTo(" ", n) + "1}" }},
		{"P6 true and garbage", func(n int) string { return "true" + repeatTo(" ", n) + "x" }},
		{"P6 only spaces", func(n int) string { return repeatTo(" ", n) }},
		{"P6 zero bytes", func(n int) string { return repeatTo("\x00", n) }},
	}

	for _, tc := range cases {
		for _, size := range longSizes {
			t.Run(tc.name+"/"+size.name, func(t *testing.T) {
				t.Parallel()

				data := tc.data(size.n)
				err := readJSON(t, data)

				assertLongInputError(t, err, data, 32)
				assertInvalidForm(t, formOnly, err)
			})
		}
	}
}

// T2 and T3 on P5 and P6 at the boundary of the data.
func TestUnmarshalJSONExcerptBoundary(t *testing.T) {
	t.Parallel()

	x := func(n int) string { return strings.Repeat("x", n) }

	cases := []struct {
		name string
		data string
		// k is the excerpt length; 0 means short data quoted whole.
		k int
	}{
		{"P5 32 bytes", `"` + x(31), 0},
		{"P5 33 bytes", `"` + x(32), 32},
		{"P6 32 bytes", x(32), 0},
		{"P6 33 bytes", x(33), 32},
		{"P5 3-byte rune across the boundary", `"` + x(29) + "€" + x(10), 30},
		{"P5 invalid UTF-8 at byte 32", `"` + x(30) + "\xff" + x(10) + `"`, 32},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := readJSON(t, tc.data)

			if tc.k == 0 {
				assertShortInputError(t, err, tc.data)
			} else {
				assertLongInputError(t, err, tc.data, tc.k)
			}

			assertInvalidForm(t, formOnly, err)
		})
	}
}

// ownQuotation returns the part of the error text that the package writes
// about the data: from the quotation to the separator ": " that starts the
// nested JSON diagnostic, or to the end of the text. This is the same part of
// the text that T2a inspects (architecture, KR-S15: the text before the
// quotation has no '"', and the quotation is separated from the nested
// diagnostic by ": "). The nested diagnostic is not part of it: its content
// may change and the tests do not check it, so it may mention any text or
// number. The data must not contain ": ", or the separator could not be told
// from the data.
func ownQuotation(t *testing.T, msg, data string) string {
	t.Helper()

	require.NotContains(t, data, ": ", "the data must not contain the separator")

	i := strings.Index(msg, `"`)
	require.GreaterOrEqual(t, i, 0, "the error text must have a quotation")

	own := msg[i:]
	if j := strings.Index(own, ": "); j >= 0 {
		own = own[:j]
	}

	return own
}

// assertWholeDataCited checks that the error of data cites the whole data,
// whitespace included, and not the value that the libraries would hand over
// after cutting the whitespace: the part of the text that the package writes
// about the data has neither the quotation nor the length of that value. The
// nested diagnostic is not inspected; see ownQuotation.
func assertWholeDataCited(t *testing.T, err error, data string) {
	t.Helper()

	stripped := strings.Trim(data, " \t\r\n")
	require.NotEqual(t, data, stripped, "the data must have whitespace around the value")

	own := ownQuotation(t, requireErrorText(t, err, len(data)), data)
	assert.Falsef(t, strings.Contains(own, strconv.Quote(stripped)),
		"the quotation must not be of the data without its whitespace: %s", strconv.Quote(stripped))

	// Only a long data has a length in the text; the digits of a short one
	// belong to its quotation (T2a covers a short data). The check is sound
	// only if the two numbers do not nest and the data has no digits.
	if len(data) > 32 {
		require.NotContains(t, strconv.Itoa(len(data)), strconv.Itoa(len(stripped)))
		require.False(t, strings.ContainsAny(data, "0123456789"), "the data must have no digits")
		assert.Falsef(t, strings.Contains(own, strconv.Itoa(len(stripped))),
			"the quotation must not give the length %d of the data without its whitespace", len(stripped))
	}
}

// P5 and P6 with whitespace around the data: the quoted value is the whole
// data, with the whitespace, and not the value cut out of it. A short data is
// quoted whole without a length.
func TestUnmarshalJSONExcerptSurroundingSpaceShort(t *testing.T) {
	t.Parallel()

	bases := []struct {
		name string
		json string
	}{
		{"P5 invalid escape", `"C\x2B1"`},
		{"P5 no closing quote", `"C+1`},
		{"P5 two values", `"C+1" "D"`},
		{"P6 number", `1`},
		{"P6 array", `[]`},
	}

	for _, b := range bases {
		for _, sp := range spaceVariants() {
			t.Run(b.name+"/"+sp.name, func(t *testing.T) {
				t.Parallel()

				data := sp.lead + b.json + sp.trail

				err := readJSON(t, data)

				assertShortInputError(t, err, data)
				assertInvalidForm(t, formOnly, err)
				assertWholeDataCited(t, err, data)
			})
		}
	}
}

// P5 with whitespace around a long data: the excerpt is taken from the start
// of the whole data and the length is the length of the whole data. The
// leading whitespace shifts the excerpt, and the whitespace on either side
// changes the length.
func TestUnmarshalJSONExcerptSurroundingSpaceLong(t *testing.T) {
	t.Parallel()

	x := func(n int) string { return strings.Repeat("x", n) }

	// The value is `"\q` + x(n) + `"`: a string with an invalid escape.
	invalid := func(n int) string { return `"\q` + x(n) + `"` }

	cases := []struct {
		name string
		data string
		// k is the excerpt length by the rule of the boundary.
		k int
	}{
		// The data is longer than 32 bytes, the value is not (29 and 30 bytes).
		{"short value, both sides", " \t\r\n" + invalid(25) + "\r\n\t ", 32},
		{"short value, leading", " \t\r\n" + invalid(26), 32},
		{"short value, trailing", invalid(26) + "\r\n\t ", 32},
		{"short value, leading space only", "    " + invalid(26), 32},
		{"short value, leading tab only", "\t\t\t\t" + invalid(26), 32},
		{"short value, leading CR only", "\r\r\r\r" + invalid(26), 32},
		{"short value, leading LF only", "\n\n\n\n" + invalid(26), 32},
		{"short value, trailing space only", invalid(26) + "    ", 32},
		{"short value, trailing tab only", invalid(26) + "\t\t\t\t", 32},
		{"short value, trailing CR only", invalid(26) + "\r\r\r\r", 32},
		{"short value, trailing LF only", invalid(26) + "\n\n\n\n", 32},
		// Both are longer than 32 bytes; the lengths and the excerpts differ.
		{"long value, both sides", " \t\r\n" + invalid(60) + "\r\n\t ", 32},
		{"long value, leading", " \t\r\n" + invalid(60), 32},
		{"long value, trailing", invalid(60) + "\r\n\t ", 32},
		{"long value, leading space", strings.Repeat(" ", 5) + invalid(60) + strings.Repeat(" ", 5), 32},
		{"long value, leading tab", strings.Repeat("\t", 5) + invalid(60) + strings.Repeat("\t", 5), 32},
		{"long value, leading CR", strings.Repeat("\r", 5) + invalid(60) + strings.Repeat("\r", 5), 32},
		{"long value, leading LF", strings.Repeat("\n", 5) + invalid(60) + strings.Repeat("\n", 5), 32},
		// The leading whitespace moves a 3-byte rune across the boundary of
		// the data (bytes 30-32, so k is 30), while in the value it ends at
		// byte 30 and the excerpt of the value is 32 bytes.
		{"rune across the boundary of the data", "  \"" + x(27) + "€" + x(10), 30},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Greater(t, len(tc.data), 32, "a long data")

			err := readJSON(t, tc.data)

			assertLongInputError(t, err, tc.data, tc.k)
			assertInvalidForm(t, formOnly, err)
			assertWholeDataCited(t, err, tc.data)
		})
	}
}

// The whitespace inside a JSON string is content, not whitespace around the
// value: it is not cut, so the content is not a machine form, and the error
// quotes the decoded content.
func TestUnmarshalJSONSpaceInsideString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		data string
		// content is the decoded content of the JSON string.
		content string
	}{
		{"leading space", ` " C+1" `, " C+1"},
		{"trailing space", "\t\"C+1 \"\n", "C+1 "},
		{"escaped tab", ` "\tC+1" `, "\tC+1"},
		{"space in the unset mark", "\r\n\"unset \" ", "unset "},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := readJSON(t, tc.data)

			assertShortInputError(t, err, tc.content)
			assertInvalidForm(t, formOnly, err)
		})
	}
}

// T2 and the former classes on the short inputs of the existing tables: the
// whole input is quoted.
func TestReadErrorExcerptShortInput(t *testing.T) {
	t.Parallel()

	for _, path := range excerptPaths() {
		for _, tc := range invalidForms() {
			t.Run(path.name+"/"+strconv.Quote(tc.text), func(t *testing.T) {
				t.Parallel()

				err := path.read(t, tc.text)
				assertShortInputError(t, err, tc.text)
				assertInvalidForm(t, tc.class, err)
			})
		}
	}

	// Inputs whose quotation contains digits: the digit of the length is
	// told apart from the digits of the quoted input.
	quotedDigits := []struct {
		name  string
		input string
	}{
		{"one byte 0x01", "\x01"},
		{"one byte 0x00", "\x00"},
		{"one byte digit", "1"},
		{"digit equal to the length", "a3b"},
		{"digits of 32 bytes", strings.Repeat("3", 16) + strings.Repeat("2", 16)},
		{"escapes of 32 bytes", strings.Repeat("\x01", 32)},
	}

	for _, path := range excerptPaths() {
		for _, tc := range quotedDigits {
			t.Run(path.name+"/quoted digits/"+tc.name, func(t *testing.T) {
				t.Parallel()

				err := path.read(t, tc.input)
				assertShortInputError(t, err, tc.input)
				assertInvalidForm(t, formOnly, err)
			})
		}
	}

	direct := []struct {
		name string
		data string
		// quoted is the value the error quotes: the content of a valid JSON
		// string, otherwise the whole data.
		quoted string
	}{
		{"empty", "", ""},
		{"number", "1", "1"},
		{"one byte 0x01", "\x01", "\x01"},
		{"string of one byte 0x01", "\"\x01\"", "\"\x01\""},
		{"string of an escape 0x01", `"\u0001"`, "\x01"},
		{"true", "true", "true"},
		{"array", "[]", "[]"},
		{"object", "{}", "{}"},
		{"empty string", `""`, ""},
		{"notation", `"C#"`, "C#"},
		{"escaped notation", `"C\u0023"`, "C#"},
		{"unquoted form", "C+1", "C+1"},
		{"unterminated string", `"C+1`, `"C+1`},
		{"two values", `"C+1" "D"`, `"C+1" "D"`},
		{"invalid escape", `"C\x2B1"`, `"C\x2B1"`},
		{"invalid escape of a digit", `"\4"`, `"\4"`},
		{"digit and control byte", "\"1\x04\"", "\"1\x04\""},
		{"surrogate escape in parentheses", `"\u(7)"`, `"\u(7)"`},
		{"surrogate escape with a number in parentheses", `"\uD800(12)"`, `"\uD800(12)"`},
		{"surrogate escape with a number and a word in parentheses", `"\uD800(14 a)"`, `"\uD800(14 a)"`},
	}

	// A control byte inside a JSON string: the nested JSON diagnostic may
	// quote it, and its digits may equal the length of the data (3 bytes).
	// 0x01 is the case "string of one byte 0x01" above.
	for _, c := range []byte{0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x1f} {
		data := string([]byte{'"', c, '"'})
		direct = append(direct, struct {
			name   string
			data   string
			quoted string
		}{fmt.Sprintf("string of control byte 0x%02x", c), data, data})
	}

	for _, tc := range direct {
		t.Run("UnmarshalJSON/"+tc.name, func(t *testing.T) {
			t.Parallel()

			err := readJSON(t, tc.data)
			assertShortInputError(t, err, tc.quoted)
			assertInvalidForm(t, formOnly, err)
		})
	}
}

// P7: through encoding/json and encoding/json/v2, a field with a short name
// at the top level of the document: T1 and the class ErrInvalidForm.
func TestJSONLibraryReadErrorLength(t *testing.T) {
	t.Parallel()

	type document struct {
		F spelling.Spelling `json:"f"`
	}

	jsonString := func(t *testing.T, x string) string {
		t.Helper()

		data, err := jsonv2.Marshal(x)
		require.NoError(t, err)

		return string(data)
	}

	values := []struct {
		name  string
		value func(t *testing.T, n int) string
	}{
		{"zero bytes string", func(t *testing.T, n int) string { return jsonString(t, repeatTo("\x00", n)) }},
		{"x string", func(t *testing.T, n int) string { return jsonString(t, repeatTo("x", n)) }},
		{"number", func(_ *testing.T, n int) string { return repeatTo("1", n) }},
	}

	for _, codec := range jsonCodecs() {
		for _, v := range values {
			for _, size := range longSizes {
				t.Run(codec.name+"/"+v.name+"/"+size.name, func(t *testing.T) {
					t.Parallel()

					data := `{"f":` + v.value(t, size.n) + `}`
					doc := document{F: spelling.DFlat()}
					err := codec.unmarshal([]byte(data), &doc)

					requireErrorText(t, err, len(data))
					assert.ErrorIs(t, err, spelling.ErrInvalidForm)
					assert.Equal(t, spelling.DFlat(), doc.F, "the field must keep its value")
				})
			}
		}
	}
}
