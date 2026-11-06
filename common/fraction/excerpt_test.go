package fraction_test

import (
	"errors"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Properties T1-T5 of decision KR-F9 (architecture, "Текст ошибок разбора"): the rejected text is quoted by an
// excerpt of at most 64 bytes without splitting a valid UTF-8 character, together with its length for long
// texts. The tests check properties of the error text, never its exact form.
//
// Failure messages of the tests carry case names and lengths only: inputs of 1 MiB, their quotes and the error
// texts must not reach the test log.

const (
	excerptLimit = 64
	errorLimit   = 512
	kib          = 1024
	mib          = 1024 * 1024
	digits       = "0123456789"
)

// sentinelErrors lists all six exported sentinel errors of the package.
var sentinelErrors = []struct {
	name string
	err  error
}{
	{"ErrInvalid", fraction.ErrInvalid},
	{"ErrZeroDenominator", fraction.ErrZeroDenominator},
	{"ErrDivisionByZero", fraction.ErrDivisionByZero},
	{"ErrNotRepresentable", fraction.ErrNotRepresentable},
	{"ErrOutOfRange", fraction.ErrOutOfRange},
	{"ErrInvalidText", fraction.ErrInvalidText},
}

// input is a rejected text prepared once: the string for Parse and MustParse, the slice for UnmarshalText.
type input struct {
	s string
	b []byte
}

func newInput(s string) input {
	return input{s: s, b: []byte(s)}
}

// excerptCase is one rejected text with the expected sentinel error of its path.
type excerptCase struct {
	name  string
	in    input
	class error
	// wantLen is the expected length of a short text (guards against a wrongly built input).
	wantLen int
	// wantK is the expected excerpt length of a long text by item 3 of KR-F9.
	wantK int
	// wantEscape, if not empty, must appear in the error text (an invalid byte shown as \xNN).
	wantEscape string
}

// parseCall is one of the three paths that quote the rejected text.
type parseCall struct {
	name string
	run  func(t *testing.T, in input) error
}

var parseCalls = []parseCall{
	{"Parse", callParse},
	{"MustParse", callMustParse},
	{"UnmarshalText", callUnmarshalText},
}

func callParse(t *testing.T, in input) error {
	t.Helper()

	got, err := fraction.Parse(in.s)
	require.Error(t, err, "Parse must fail, len(x)=%d", len(in.s))
	assert.Equal(t, fraction.Fraction{}, got, "Parse must return the zero Fraction on error, len(x)=%d", len(in.s))

	return err
}

func callMustParse(t *testing.T, in input) error {
	t.Helper()

	var recovered any
	func() {
		defer func() { recovered = recover() }()

		_ = fraction.MustParse(in.s)
	}()

	require.NotNil(t, recovered, "MustParse must panic, len(x)=%d", len(in.s))
	err, ok := recovered.(error)
	require.True(t, ok, "MustParse panic value must be an error, got %T, len(x)=%d", recovered, len(in.s))

	return err
}

func callUnmarshalText(t *testing.T, in input) error {
	t.Helper()

	orig := fraction.MustNew(6, 12)
	target := orig
	err := target.UnmarshalText(in.b)
	require.Error(t, err, "UnmarshalText must fail, len(x)=%d", len(in.b))
	assert.Equal(t, orig, target, "UnmarshalText must not change the receiver on error, len(x)=%d", len(in.b))

	return err
}

// assertClass checks that err matches exactly the sentinel error class and none of the other five (T4).
func assertClass(t *testing.T, err, class error, label string) {
	t.Helper()

	for _, s := range sentinelErrors {
		assert.Equal(t, s.err == class, errors.Is(err, s.err), "%s: errors.Is(err, %s)", label, s.name)
	}
}

// excerptLen is the oracle of item 3 of KR-F9: the longest prefix of x of whole decoding units not longer than
// 64 bytes. A valid encoding of a rune is one unit; any other byte is a unit of one byte.
func excerptLen(x string) int {
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

// repeat builds R(u, n): n bytes for a one-byte u, floor(n/len(u)) whole runes for a multi-byte u.
func repeat(u string, n int) string {
	return strings.Repeat(u, n/len(u))
}

// splitAtQuote returns the error text without the first occurrence of quote.
func splitAtQuote(msg, quote string) (rest string, found bool) {
	i := strings.Index(msg, quote)
	if i < 0 {
		return msg, false
	}

	return msg[:i] + msg[i+len(quote):], true
}

// checkShort verifies T1, T2, T2a and T4 for a text of at most 64 bytes.
func checkShort(t *testing.T, c excerptCase, err error, label string) {
	t.Helper()

	x := c.in.s
	require.Equal(t, c.wantLen, len(x), "%s: input built with a wrong length", label)
	require.LessOrEqual(t, len(x), excerptLimit, "%s: input is not short", label)

	msg := err.Error()
	assertClass(t, err, c.class, label)
	assert.LessOrEqual(t, len(msg), errorLimit, "T1 %s: len(x)=%d", label, len(x))

	rest, found := splitAtQuote(msg, strconv.Quote(x))
	assert.True(t, found, "T2 %s: error must quote the whole text, len(x)=%d, error=%q", label, len(x), msg)
	assert.False(t, strings.ContainsAny(rest, digits),
		"T2a %s: no length may be added to a short text, len(x)=%d, outside the quote=%q", label, len(x), rest)
}

// checkLong verifies T1, T3 and T4 for a text longer than 64 bytes.
func checkLong(t *testing.T, c excerptCase, err error, label string) {
	t.Helper()

	x := c.in.s
	require.Greater(t, len(x), excerptLimit, "%s: input is not long", label)
	assert.Equal(t, c.wantK, excerptLen(x), "%s: oracle of the excerpt length disagrees with the table", label)

	msg := err.Error()
	assertClass(t, err, c.class, label)
	assert.LessOrEqual(t, len(msg), errorLimit, "T1 %s: len(x)=%d", label, len(x))

	k := c.wantK
	rest, found := splitAtQuote(msg, strconv.Quote(x[:k]))
	assert.True(t, found, "T3 %s: error must quote the excerpt of k=%d bytes, len(x)=%d, len(err)=%d",
		label, k, len(x), len(msg))
	assert.True(t, strings.Contains(rest, strconv.Itoa(len(x))),
		"T3 %s: error must state the length outside the excerpt, len(x)=%d, len(err)=%d", label, len(x), len(msg))

	// A full quote is longer than the text itself, so it cannot fit into a shorter message.
	if len(x)+2 <= len(msg) {
		assert.False(t, strings.Contains(msg, strconv.Quote(x)),
			"T3 %s: error must not quote the whole text, len(x)=%d, len(err)=%d", label, len(x), len(msg))
	}

	if c.wantEscape != "" {
		assert.True(t, strings.Contains(msg, c.wantEscape),
			"T3 %s: excerpt must show the byte %s, len(x)=%d, len(err)=%d", label, c.wantEscape, len(x), len(msg))
	}
}

// runCases runs every case on every path in parallel subtests.
func runCases(t *testing.T, cases []excerptCase, check func(t *testing.T, c excerptCase, err error, label string)) {
	t.Helper()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			for _, call := range parseCalls {
				t.Run(call.name, func(t *testing.T) {
					t.Parallel()

					err := call.run(t, c.in)
					check(t, c, err, call.name)
				})
			}
		})
	}
}

var excerptSizes = []struct {
	name string
	n    int
}{
	{"1KiB", kib},
	{"1MiB", mib},
}

// excerptFillings are the long fillings L of architecture; k is the expected excerpt length of a uniform text.
var excerptFillings = []struct {
	name string
	unit string
	k    int
}{
	{"x", "x", 64},
	{"NUL", "\x00", 64},
	{"SOH", "\x01", 64},
	{"e_acute", "é", 64},
	{"euro", "€", 63},
	{"emoji", "😀", 64},
	{"invalid_ff", "\xff", 64},
	{"continuation_80", "\x80", 64},
}

// TestParseErrorExcerptSentinelTextPrecondition guards the format on which T2a relies: the texts of the three
// sentinel errors of parsing contain neither a double quote nor a decimal digit.
func TestParseErrorExcerptSentinelTextPrecondition(t *testing.T) {
	t.Parallel()

	for _, s := range []struct {
		name string
		err  error
	}{
		{"ErrInvalidText", fraction.ErrInvalidText},
		{"ErrZeroDenominator", fraction.ErrZeroDenominator},
		{"ErrOutOfRange", fraction.ErrOutOfRange},
	} {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()

			text := s.err.Error()
			assert.NotContains(t, text, `"`)
			assert.False(t, strings.ContainsAny(text, digits), "sentinel text must not contain digits: %q", text)
		})
	}
}

// TestParseErrorExcerptShortTexts verifies T1, T2, T2a and T4 on texts of at most 64 bytes: the whole text is
// quoted and no length is added. The text of exactly 64 bytes is the boundary of the rule.
func TestParseErrorExcerptShortTexts(t *testing.T) {
	t.Parallel()

	cases := []excerptCase{
		// ErrInvalidText: any content and length are reachable.
		shortCase("invalid_empty", "", 0, fraction.ErrInvalidText),
		shortCase("invalid_1_byte", "a", 1, fraction.ErrInvalidText),
		shortCase("invalid_63_x", repeat("x", 63), 63, fraction.ErrInvalidText),
		shortCase("invalid_64_x", repeat("x", 64), 64, fraction.ErrInvalidText),
		shortCase("invalid_17_NUL_issue", repeat("\x00", 17), 17, fraction.ErrInvalidText),
		shortCase("invalid_64_NUL", repeat("\x00", 64), 64, fraction.ErrInvalidText),
		shortCase("invalid_64_continuation_80", repeat("\x80", 64), 64, fraction.ErrInvalidText),
		shortCase("invalid_32_e_acute_64_bytes", repeat("é", 64), 64, fraction.ErrInvalidText),
		shortCase("invalid_digits_inside_quote_64", repeat("9", 63)+"x", 64, fraction.ErrInvalidText),
		shortCase("invalid_x_then_63_ff_64", "x"+repeat("\xff", 63), 64, fraction.ErrInvalidText),
		shortCase("invalid_62_x_e_acute_ends_at_64", repeat("x", 62)+"é", 64, fraction.ErrInvalidText),
		shortCase("invalid_63_x_continuation_80_ends_at_64", repeat("x", 63)+"\x80", 64, fraction.ErrInvalidText),
		// ErrZeroDenominator. ErrOutOfRange is unreachable for short texts: the shortest such text,
		// MaxComponent+1, has 77 bytes.
		shortCase("zero_denominator_1/0", "1/0", 3, fraction.ErrZeroDenominator),
		shortCase("zero_denominator_0/0", "0/0", 3, fraction.ErrZeroDenominator),
		shortCase("zero_denominator_mixed_1_1/0", "1 1/0", 5, fraction.ErrZeroDenominator),
		shortCase("zero_denominator_63", "1/"+repeat("0", 61), 63, fraction.ErrZeroDenominator),
		shortCase("zero_denominator_64", "1/"+repeat("0", 62), 64, fraction.ErrZeroDenominator),
		shortCase("zero_denominator_digits_64", repeat("9", 31)+"/"+repeat("0", 32), 64, fraction.ErrZeroDenominator),
	}

	runCases(t, cases, checkShort)
}

func shortCase(name, x string, wantLen int, class error) excerptCase {
	return excerptCase{name: name, in: newInput(x), class: class, wantLen: wantLen}
}

// TestParseErrorExcerptLongTexts verifies T1, T3 and T4 on long texts of all three classes, 1 KiB and 1 MiB,
// on Parse, MustParse and UnmarshalText.
func TestParseErrorExcerptLongTexts(t *testing.T) {
	t.Parallel()

	maxPlusOne := new(big.Int).Add(fraction.MaxComponentBig(), big.NewInt(1)).String()
	require.Len(t, maxPlusOne, 77, "MaxComponent+1 must have 77 decimal digits")

	var cases []excerptCase
	add := func(name, x string, class error) {
		cases = append(cases, excerptCase{name: name, in: newInput(x), class: class, wantK: excerptLimit})
	}

	// Every long input is built here once and only read by the parallel subtests.
	for _, size := range excerptSizes {
		for _, f := range excerptFillings {
			cases = append(cases, excerptCase{
				name:  "invalid_L_" + f.name + "_" + size.name,
				in:    newInput(repeat(f.unit, size.n)),
				class: fraction.ErrInvalidText,
				wantK: f.k,
			})
		}

		digitsText := repeat("9", size.n)
		add("invalid_D_slash_x_"+size.name, digitsText+"/x", fraction.ErrInvalidText)
		add("invalid_x_slash_zeros_"+size.name, "x/"+repeat("0", size.n), fraction.ErrInvalidText)
		add("range_D_"+size.name, digitsText, fraction.ErrOutOfRange)
		add("range_1_slash_D_"+size.name, "1/"+digitsText, fraction.ErrOutOfRange)
		add("range_D_slash_0_"+size.name, digitsText+"/0", fraction.ErrOutOfRange)
	}

	add("invalid_65_x", repeat("x", 65), fraction.ErrInvalidText)
	add("range_MaxComponent_plus_1_77", maxPlusOne, fraction.ErrOutOfRange)
	add("range_101_nines", repeat("9", 101), fraction.ErrOutOfRange)
	add("range_mixed_83", repeat("9", 40)+" 1/"+repeat("9", 40), fraction.ErrOutOfRange)
	add("range_decimal_102", repeat("9", 100)+".5", fraction.ErrOutOfRange)
	add("range_1_slash_101_zeros_103", "1/"+repeat("0", 101), fraction.ErrOutOfRange)
	add("zero_denominator_1_slash_63_zeros_65", "1/"+repeat("0", 63), fraction.ErrZeroDenominator)
	add("zero_denominator_100_nines_slash_0_102", repeat("9", 100)+"/0", fraction.ErrZeroDenominator)
	add("zero_denominator_mixed_100_zeros_104", "1 1/"+repeat("0", 100), fraction.ErrZeroDenominator)
	add("zero_denominator_max_path_303",
		"-"+repeat("9", 100)+" "+repeat("9", 100)+"/"+repeat("0", 100), fraction.ErrZeroDenominator)

	runCases(t, cases, checkLong)
}

// TestParseErrorExcerptCheckPriorityOnLongTexts verifies T4: the order of checks (syntax, token length, b == 0,
// range) does not change on long texts. Each case separates one pair of adjacent priorities.
func TestParseErrorExcerptCheckPriorityOnLongTexts(t *testing.T) {
	t.Parallel()

	maxPlusOne := new(big.Int).Add(fraction.MaxComponentBig(), big.NewInt(1)).String()
	require.Len(t, maxPlusOne, 77, "MaxComponent+1 must have 77 decimal digits")

	digitsText := repeat("9", kib)
	zeros := repeat("0", kib)

	build := func(name, x string, class error) excerptCase {
		return excerptCase{name: name, in: newInput(x), class: class, wantK: excerptLimit}
	}

	cases := []excerptCase{
		build("syntax_before_token_length_D_slash_x", digitsText+"/x", fraction.ErrInvalidText),
		build("syntax_before_zero_denominator_x_slash_zeros", "x/"+zeros, fraction.ErrInvalidText),
		build("token_length_before_zero_denominator_D_slash_0", digitsText+"/0", fraction.ErrOutOfRange),
		build("token_length_before_zero_denominator_1_slash_101_zeros", "1/"+repeat("0", 101), fraction.ErrOutOfRange),
		build("zero_denominator_before_range_100_nines_slash_0", repeat("9", 100)+"/0", fraction.ErrZeroDenominator),
		build("zero_denominator_before_range_mixed_303",
			"-"+repeat("9", 100)+" "+repeat("9", 100)+"/"+repeat("0", 100), fraction.ErrZeroDenominator),
		build("range_MaxComponent_plus_1", maxPlusOne, fraction.ErrOutOfRange),
		build("range_mixed_overflow", repeat("9", 40)+" 1/"+repeat("9", 40), fraction.ErrOutOfRange),
		build("range_decimal_overflow", repeat("9", 100)+".5", fraction.ErrOutOfRange),
	}

	runCases(t, cases, checkLong)
}

// TestParseErrorExcerptBoundary verifies T3 on the boundary of the excerpt: a rune ending exactly at the 64th
// byte is quoted whole, a rune crossing the boundary is not, and an invalid byte is a unit of one byte.
func TestParseErrorExcerptBoundary(t *testing.T) {
	t.Parallel()

	tail := repeat("x", 10)

	build := func(name, x string, k int, escape string) excerptCase {
		return excerptCase{name: name, in: newInput(x), class: fraction.ErrInvalidText, wantK: k, wantEscape: escape}
	}

	cases := []excerptCase{
		build("e_acute_ends_at_64", repeat("x", 62)+"é"+tail, 64, ""),
		build("e_acute_crosses_with_tail", repeat("x", 63)+"é"+tail, 63, ""),
		build("e_acute_crosses_65_bytes", repeat("x", 63)+"é", 63, ""),
		build("euro_ends_at_64", repeat("x", 61)+"€"+tail, 64, ""),
		build("euro_crosses_with_tail", repeat("x", 62)+"€"+tail, 62, ""),
		build("euro_crosses_66_bytes", repeat("x", 63)+"€", 63, ""),
		build("emoji_ends_at_64", repeat("x", 60)+"😀"+tail, 64, ""),
		build("emoji_crosses_boundary", repeat("x", 61)+"😀"+tail, 61, ""),
		build("invalid_ff_at_64th_byte", repeat("x", 63)+"\xff"+tail, 64, `\xff`),
		build("invalid_ff_right_after_boundary", repeat("x", 64)+"\xff"+tail, 64, ""),
		build("invalid_ff_right_after_boundary_65_bytes", repeat("x", 64)+"\xff", 64, ""),
		build("invalid_leading_byte_without_continuation", repeat("x", 63)+"\xd1"+tail, 64, `\xd1`),
		build("invalid_two_e0_issue", repeat("x", 63)+"\xe0\xe0"+tail, 64, `\xe0`),
		build("truncated_3_byte_sequence_at_boundary", repeat("x", 63)+"\xe2\x82"+tail, 64, `\xe2`),
		build("truncated_3_byte_sequence_65_bytes", repeat("x", 63)+"\xe2\x82", 64, `\xe2`),
		build("lone_continuation_80_at_64th_byte", repeat("x", 63)+"\x80"+tail, 64, `\x80`),
		build("100_continuation_bytes_issue", repeat("\x80", 100), 64, `\x80`),
		build("65_x", repeat("x", 65), 64, ""),
	}

	zero := build("zero_denominator_65_bytes", "1/"+repeat("0", 63), 64, "")
	zero.class = fraction.ErrZeroDenominator
	cases = append(cases, zero)

	runCases(t, cases, checkLong)
}

// TestParseErrorExcerptOracle checks the oracle of the excerpt length against the table of architecture, so
// that an error of the oracle cannot hide an error of the implementation.
func TestParseErrorExcerptOracle(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		x    string
		want int
	}{
		{"e_acute_ends_at_64", repeat("x", 62) + "é" + repeat("x", 10), 64},
		{"e_acute_crosses_boundary", repeat("x", 63) + "é" + repeat("x", 10), 63},
		{"euro_ends_at_64", repeat("x", 61) + "€" + repeat("x", 10), 64},
		{"euro_crosses_boundary", repeat("x", 62) + "€" + repeat("x", 10), 62},
		{"emoji_ends_at_64", repeat("x", 60) + "😀" + repeat("x", 10), 64},
		{"emoji_crosses_boundary", repeat("x", 61) + "😀" + repeat("x", 10), 61},
		{"invalid_ff_at_64th_byte", repeat("x", 63) + "\xff" + repeat("x", 10), 64},
		{"invalid_ff_after_boundary", repeat("x", 64) + "\xff" + repeat("x", 10), 64},
		{"leading_byte_without_continuation", repeat("x", 63) + "\xd1" + repeat("x", 10), 64},
		{"two_e0", repeat("x", 63) + "\xe0\xe0" + repeat("x", 10), 64},
		{"truncated_3_byte_sequence", repeat("x", 63) + "\xe2\x82" + repeat("x", 10), 64},
		{"lone_continuation", repeat("x", 63) + "\x80" + repeat("x", 10), 64},
		{"100_continuation_bytes", repeat("\x80", 100), 64},
		{"short_text_is_its_own_excerpt", repeat("é", 64), 64},
		{"empty_text", "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, excerptLen(tc.x))
		})
	}

	for _, f := range excerptFillings {
		for _, size := range excerptSizes {
			t.Run("uniform_"+f.name+"_"+size.name, func(t *testing.T) {
				t.Parallel()

				k := excerptLen(repeat(f.unit, size.n))
				assert.Equal(t, f.k, k)
				assert.GreaterOrEqual(t, k, 61, "an excerpt of a long text is never shorter than 61 bytes")
				assert.LessOrEqual(t, k, excerptLimit)
			})
		}
	}
}

// TestParseErrorExcerptNilReceiverCitesNothing verifies T5: UnmarshalText on a nil receiver returns ErrInvalid
// without looking at the text, so the answer does not depend on the text.
func TestParseErrorExcerptNilReceiverCitesNothing(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		name string
		b    []byte
	}{
		{"short_1/2", []byte("1/2")},
		{"empty", []byte("")},
		{"long_1MiB_x", []byte(repeat("x", mib))},
	}

	var nilReceiver *fraction.Fraction

	messages := make([]string, len(inputs))
	for i, in := range inputs {
		err := nilReceiver.UnmarshalText(in.b)
		require.Error(t, err, "%s: UnmarshalText on a nil receiver must fail", in.name)
		assertClass(t, err, fraction.ErrInvalid, in.name)

		messages[i] = err.Error()
	}

	for i := range inputs {
		for j := i + 1; j < len(inputs); j++ {
			assert.True(t, messages[i] == messages[j],
				"nil receiver: answers for %s and %s differ, len=%d and %d",
				inputs[i].name, inputs[j].name, len(messages[i]), len(messages[j]))
		}
	}
}
