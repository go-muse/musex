package pitchclass_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests check the properties T1-T5 of the decision KR-PC17: the error of
// Parse, ParseNotation, and ParseOperator, and the panic of their Must
// variants, quote the rejected text in full up to 64 bytes and by a bounded
// excerpt with the length in bytes beyond that.
//
// Failure messages never contain a tested text or the text of an error: with
// inputs of 1 MiB an error may be several MiB long. They carry only names and
// lengths, and the checks use assert.True on a computed condition instead of
// assert.Contains, assert.Equal, or assert.ErrorIs.

const (
	// excerptLimit is the longest text quoted in full (KR-PC17).
	excerptLimit = 64
	// excerptErrorLimit is the bound of len(err.Error()) (T1).
	excerptErrorLimit = 512

	kiB = 1 << 10
	miB = 1 << 20
)

// excerptEnd is the oracle of the cut point k of a text longer than
// excerptLimit bytes: the text is divided into units by decoding it from the
// start, a valid encoding of a rune is one unit of 1-4 bytes and every other
// byte is a unit of one byte, and k is the longest prefix of whole units that
// is not longer than excerptLimit. The text itself is decoded, not its
// excerptLimit-byte prefix.
func excerptEnd(s string) int {
	k := 0
	for k < len(s) {
		_, size := utf8.DecodeRuneInString(s[k:])
		if k+size > excerptLimit {
			break
		}

		k += size
	}

	return k
}

// repeatBytes is R(u, n): n bytes for a one-byte unit, otherwise the whole
// units that fit into n bytes.
func repeatBytes(unit string, n int) string {
	return strings.Repeat(unit, n/len(unit))
}

// excerptUnit is a unit of the long filler L with the cut points that follow
// from the definition: kPlain for the text L and kPrefixed for "T" + L.
type excerptUnit struct {
	name      string
	unit      string
	kPlain    int
	kPrefixed int
}

func excerptUnits() []excerptUnit {
	return []excerptUnit{
		{name: "x", unit: "x", kPlain: 64, kPrefixed: 64},
		{name: "nul", unit: "\x00", kPlain: 64, kPrefixed: 64},
		{name: "soh", unit: "\x01", kPlain: 64, kPrefixed: 64},
		{name: "e-acute", unit: "é", kPlain: 64, kPrefixed: 63},
		{name: "euro", unit: "€", kPlain: 63, kPrefixed: 64},
		{name: "emoji", unit: "😀", kPlain: 64, kPrefixed: 61},
		{name: "xff", unit: "\xff", kPlain: 64, kPrefixed: 64},
		{name: "x80", unit: "\x80", kPlain: 64, kPrefixed: 64},
	}
}

// excerptFiller is a built filler of a given size. Fillers are built once per
// test function and only read by the subtests.
type excerptFiller struct {
	name string
	text string
}

var fillerSizes = []struct {
	name string
	n    int
}{
	{name: "1KiB", n: kiB},
	{name: "1MiB", n: miB},
}

// letterFillers is L: every unit at every size.
func letterFillers() []excerptFiller {
	var out []excerptFiller

	for _, u := range excerptUnits() {
		for _, size := range fillerSizes {
			out = append(out, excerptFiller{
				name: u.name + "/" + size.name,
				text: repeatBytes(u.unit, size.n),
			})
		}
	}

	return out
}

// digitFillers is D: the digit 9 at every size.
func digitFillers() []excerptFiller {
	var out []excerptFiller

	for _, size := range fillerSizes {
		out = append(out, excerptFiller{
			name: "digits/" + size.name,
			text: strings.Repeat("9", size.n),
		})
	}

	return out
}

// rejectionPath is a path of the table of KR-PC17 on which the rejected text
// is quoted: a call with a text built from a filler, and the same call through
// the Must variant.
type rejectionPath struct {
	name string
	// digits marks a path built from the digit filler D; the others use L.
	digits bool
	// text builds the whole text s passed to the function from a filler.
	text func(filler string) string
	// parse calls the function and reports whether its result is zero.
	parse func(s string) (zero bool, err error)
	// must calls the Must variant.
	must func(s string)
	// valid, if not nil, reports a short text that the function accepts, and
	// the short tests skip it.
	valid func(s string) bool
}

func identity(s string) string { return s }

func singleDigit(s string) bool {
	return len(s) == 1 && s[0] >= '0' && s[0] <= '9'
}

func parseClass(s string) (bool, error) {
	c, err := pitchclass.Parse(s)

	return c == pitchclass.Class{}, err
}

func parseNotationWith(style pitchclass.Style, modulus int) func(string) (bool, error) {
	return func(s string) (bool, error) {
		c, err := pitchclass.ParseNotation(s, style, modulus)

		return c == pitchclass.Class{}, err
	}
}

func mustParseNotationWith(style pitchclass.Style, modulus int) func(string) {
	return func(s string) { _ = pitchclass.MustParseNotation(s, style, modulus) }
}

func parseOperator(s string) (bool, error) {
	f, err := pitchclass.ParseOperator(s)

	return f == pitchclass.Operator{}, err
}

func mustParseClass(s string) { _ = pitchclass.MustParse(s) }

func mustParseOperator(s string) { _ = pitchclass.MustParseOperator(s) }

// rejectionPaths are the paths of KR-PC17: Parse, ParseNotation with a style
// defined at N = 12, and the four branches of ParseOperator that reject the
// text, and the branch that must reject the text before the multiplier is
// checked for invertibility.
func rejectionPaths() []rejectionPath {
	return []rejectionPath{
		{
			name: "Parse", text: identity,
			parse: parseClass, must: mustParseClass,
		},
		{
			name: "Parse/digits-mod", digits: true,
			text:  func(f string) string { return f + " mod 12" },
			parse: parseClass, must: mustParseClass,
		},
		{
			name: "ParseNotation/Decimal", text: identity,
			parse: parseNotationWith(pitchclass.Decimal, 12),
			must:  mustParseNotationWith(pitchclass.Decimal, 12),
			valid: singleDigit,
		},
		{
			name: "ParseNotation/Decimal/digits", digits: true, text: identity,
			parse: parseNotationWith(pitchclass.Decimal, 12),
			must:  mustParseNotationWith(pitchclass.Decimal, 12),
		},
		{
			name: "ParseNotation/TE", text: identity,
			parse: parseNotationWith(pitchclass.TE, 12),
			must:  mustParseNotationWith(pitchclass.TE, 12),
			valid: singleDigit,
		},
		{
			name: "ParseOperator/no-T", text: identity,
			parse: parseOperator, must: mustParseOperator,
		},
		{
			name:  "ParseOperator/no-mod",
			text:  func(f string) string { return "T" + f },
			parse: parseOperator, must: mustParseOperator,
		},
		{
			name:  "ParseOperator/no-M",
			text:  func(f string) string { return "T" + f + " mod 12" },
			parse: parseOperator, must: mustParseOperator,
		},
		{
			name: "ParseOperator/number-out-of-int", digits: true,
			text:  func(f string) string { return "T" + f + "M1 mod 12" },
			parse: parseOperator, must: mustParseOperator,
		},
		{
			// The text is rejected before the multiplier 2 is found
			// not invertible modulo 12.
			name: "ParseOperator/text-before-invertibility", digits: true,
			text:  func(f string) string { return "T" + f + "M2 mod 12" },
			parse: parseOperator, must: mustParseOperator,
		},
	}
}

func pathByName(t *testing.T, name string) rejectionPath {
	t.Helper()

	for _, p := range rejectionPaths() {
		if p.name == name {
			return p
		}
	}

	require.FailNow(t, "unknown rejection path", name)

	return rejectionPath{}
}

// outcome calls the path with the text s, directly or through the Must
// variant. For Must it requires a panic with an error value and returns it.
func (p rejectionPath) outcome(t *testing.T, s string, viaMust bool) (zero bool, err error) {
	t.Helper()

	if !viaMust {
		return p.parse(s)
	}

	v, panicked := panicValue(func() { p.must(s) })
	require.True(t, panicked, "%s: Must must panic (len(text)=%d)", p.name, len(s))

	e, ok := v.(error)
	require.True(t, ok, "%s: panic value of type %T is not an error", p.name, v)

	return true, e
}

func callName(viaMust bool) string {
	if viaMust {
		return "Must"
	}

	return "call"
}

var sentinelNames = []string{
	"ErrInvalidClass", "ErrInvalidModulus", "ErrModulusMismatch",
	"ErrNotInvertible", "ErrInvalidValue", "ErrNotation", "ErrOverflow",
}

// matchedSentinels names the sentinel errors that err matches by errors.Is.
func matchedSentinels(err error) []string {
	matched := []string{}

	for i, s := range allErrors() {
		if errors.Is(err, s) {
			matched = append(matched, sentinelNames[i])
		}
	}

	return matched
}

// assertOnlySentinel checks that err matches exactly one sentinel error,
// want, with errors.Is. Unlike assertOnlyError it never prints the error.
func assertOnlySentinel(t *testing.T, err, want error, desc string) bool {
	t.Helper()

	var name string

	for i, s := range allErrors() {
		if s == want {
			name = sentinelNames[i]
		}
	}

	return assert.Equal(t, []string{name}, matchedSentinels(err), "%s: matched sentinel errors", desc)
}

// withoutFirst returns msg without the first occurrence of quoted and reports
// whether it was found.
func withoutFirst(msg, quoted string) (rest string, found bool) {
	i := strings.Index(msg, quoted)
	if i < 0 {
		return "", false
	}

	return msg[:i] + msg[i+len(quoted):], true
}

// assertRejection checks the properties of KR-PC17 for the error err returned
// (or carried by the panic) for the rejected text: the classification and the
// zero result (T4), the size of the error (T1), and the quotation: T2 and T2a
// for a text of up to 64 bytes, T3 for a longer one.
func assertRejection(t *testing.T, name, text string, zero bool, err error) {
	t.Helper()

	desc := fmt.Sprintf("%s, len(text)=%d", name, len(text))

	require.Error(t, err, desc)
	assert.True(t, zero, "%s: the result must be zero", desc)
	assertOnlySentinel(t, err, pitchclass.ErrNotation, desc)

	msg := err.Error()
	assert.LessOrEqual(t, len(msg), excerptErrorLimit, "%s: T1, len(err.Error())", desc)

	if len(text) <= excerptLimit {
		// T2: the text is quoted in full.
		rest, found := withoutFirst(msg, strconv.Quote(text))
		if !assert.True(t, found, "%s: T2, the whole text must be quoted (len(err.Error())=%d)", desc, len(msg)) {
			return
		}

		// T2a: the length is not added to the quotation of a short text.
		assert.False(t, strings.ContainsAny(rest, "0123456789"),
			"%s: T2a, digits outside the quotation of a short text (len(rest)=%d)", desc, len(rest))

		return
	}

	// T3: the excerpt x[:k] is quoted, the length of the text follows it
	// outside the quotation, and the whole text is not quoted.
	k := excerptEnd(text)
	rest, found := withoutFirst(msg, strconv.Quote(text[:k]))

	if assert.True(t, found, "%s: T3, the excerpt of k=%d bytes must be quoted (len(err.Error())=%d)", desc, k, len(msg)) {
		assert.True(t, strings.Contains(rest, strconv.Itoa(len(text))),
			"%s: T3, the length of the text must follow the excerpt (k=%d)", desc, k)
	}

	// Quote(text) is longer than the text, so a shorter message cannot hold it.
	if len(msg) >= len(text)+2 {
		assert.False(t, strings.Contains(msg, strconv.Quote(text)),
			"%s: T3, the whole text must not be quoted (len(err.Error())=%d)", desc, len(msg))
	}
}

// TestExcerptOracle checks the oracle of the cut point k against the values
// that follow from the definition by hand: the fillers L and the table of the
// boundary.
func TestExcerptOracle(t *testing.T) {
	t.Parallel()

	for _, u := range excerptUnits() {
		t.Run(u.name, func(t *testing.T) {
			t.Parallel()

			text := repeatBytes(u.unit, kiB)
			assert.Equal(t, u.kPlain, excerptEnd(text), "s = L")
			assert.Equal(t, u.kPrefixed, excerptEnd("T"+text), `s = "T" + L`)
		})
	}

	for _, c := range boundaryCases() {
		t.Run("boundary/"+c.name, func(t *testing.T) {
			t.Parallel()

			assert.Greater(t, len(c.text), excerptLimit, "the text must be longer than the limit")
			assert.Equal(t, c.k, excerptEnd(c.text))
		})
	}
}

// TestRejectedTextLong checks T1, T3, and T4 for every path of KR-PC17,
// directly and through the Must variants, for the fillers L (and D for the
// paths with digits) of 1 KiB and 1 MiB.
func TestRejectedTextLong(t *testing.T) {
	t.Parallel()

	letters, digits := letterFillers(), digitFillers()

	for _, p := range rejectionPaths() {
		fillers := letters
		if p.digits {
			fillers = digits
		}

		for _, f := range fillers {
			for _, viaMust := range []bool{false, true} {
				t.Run(p.name+"/"+f.name+"/"+callName(viaMust), func(t *testing.T) {
					t.Parallel()

					s := p.text(f.text)
					require.Greater(t, len(s), excerptLimit, "the text must be longer than the limit")

					zero, err := p.outcome(t, s, viaMust)
					assertRejection(t, p.name, s, zero, err)
				})
			}
		}
	}
}

// TestRejectedTextShort checks T2, T2a, and T4 for texts of 0, 1, 63, and 64
// bytes, with digits, NUL, and 0xff, on every path of KR-PC17, directly and
// through the Must variants. A length that a path cannot reach (the text
// template is longer) or a text that the function accepts is not tested.
func TestRejectedTextShort(t *testing.T) {
	t.Parallel()

	fillers := []excerptUnit{
		{name: "x", unit: "x"},
		{name: "nul", unit: "\x00"},
		{name: "xff", unit: "\xff"},
		{name: "digits", unit: "9"},
	}

	for _, p := range rejectionPaths() {
		overhead := len(p.text(""))

		// Two digits are the least that cannot be read as a valid number
		// below 12 on the paths built from digits.
		minFiller := 0
		if p.digits {
			minFiller = 2
		}

		for _, u := range fillers {
			if p.digits && u.name != "digits" {
				continue
			}

			for _, n := range []int{0, 1, excerptLimit - 1, excerptLimit} {
				if n < overhead+minFiller {
					continue
				}

				s := p.text(strings.Repeat(u.unit, n-overhead))
				if p.valid != nil && p.valid(s) {
					continue
				}

				require.Len(t, s, n)

				for _, viaMust := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/len%d/%s", p.name, u.name, n, callName(viaMust)), func(t *testing.T) {
						t.Parallel()

						zero, err := p.outcome(t, s, viaMust)
						assertRejection(t, p.name, s, zero, err)
					})
				}
			}
		}
	}
}

type boundaryCase struct {
	name string
	text string
	// k is the expected cut point by the definition, set by hand.
	k int
}

// boundaryCases is the table of the boundary of the excerpt: the text s has
// no leading "T" and no " mod " and is rejected by all three parsers.
func boundaryCases() []boundaryCase {
	x := func(n int) string { return strings.Repeat("x", n) }

	return []boundaryCase{
		{name: "e-acute-ends-at-64", text: x(62) + "é" + x(10), k: 64},
		{name: "e-acute-across-64", text: x(63) + "é" + x(10), k: 63},
		{name: "euro-ends-at-64", text: x(61) + "€" + x(10), k: 64},
		{name: "euro-across-64", text: x(62) + "€" + x(10), k: 62},
		{name: "emoji-ends-at-64", text: x(60) + "😀" + x(10), k: 64},
		{name: "emoji-across-64", text: x(61) + "😀" + x(10), k: 61},
		{name: "invalid-byte-at-64", text: x(63) + "\xff" + x(10), k: 64},
		{name: "invalid-byte-after-64", text: x(64) + "\xff" + x(10), k: 64},
		{name: "lead-byte-without-continuation", text: x(63) + "\xd1" + x(10), k: 64},
		{name: "truncated-three-byte-sequence", text: x(63) + "\xe2\x82" + x(10), k: 64},
		{name: "lone-continuation-byte", text: x(63) + "\x80" + x(10), k: 64},
	}
}

// TestRejectedTextBoundary checks T3 on the table of the boundary of the
// excerpt on Parse, ParseNotation, and ParseOperator, and directly checks
// the excerpt of the expected length k set by hand.
func TestRejectedTextBoundary(t *testing.T) {
	t.Parallel()

	paths := []rejectionPath{
		pathByName(t, "Parse"),
		pathByName(t, "ParseNotation/TE"),
		pathByName(t, "ParseOperator/no-T"),
	}

	for _, c := range boundaryCases() {
		for _, p := range paths {
			t.Run(c.name+"/"+p.name, func(t *testing.T) {
				t.Parallel()

				zero, err := p.outcome(t, c.text, false)
				assertRejection(t, p.name, c.text, zero, err)

				// The expected excerpt is set by hand, independent of the oracle.
				rest, found := withoutFirst(err.Error(), strconv.Quote(c.text[:c.k]))
				if assert.True(t, found, "%s: the excerpt of k=%d bytes must be quoted", c.name, c.k) {
					assert.True(t, strings.Contains(rest, strconv.Itoa(len(c.text))),
						"%s: the length %d must follow the excerpt", c.name, len(c.text))
				}
			})
		}
	}
}

// TestRejectedTextAtLimit checks the texts at the limit of 64 bytes and just
// above it: the nearly valid records with the largest numbers are quoted in
// full up to 64 bytes, and the one-byte-longer ones are cut with the length.
func TestRejectedTextAtLimit(t *testing.T) {
	t.Parallel()

	const (
		maxInt64     = "9223372036854775807"
		maxInt64Prev = "9223372036854775806"
	)

	cases := []struct {
		name string
		text string
		path string
		// length is the expected len(text); the test checks the table itself.
		length int
	}{
		{
			name: "class-number-not-below-N", path: "Parse",
			text: maxInt64 + " mod " + maxInt64, length: 43,
		},
		{
			name: "operator-m-not-below-N", path: "ParseOperator/no-T",
			text: "T" + maxInt64Prev + "M" + maxInt64 + " mod " + maxInt64, length: 64,
		},
		{
			name: "operator-N-out-of-int", path: "ParseOperator/no-T",
			text: "T" + maxInt64Prev + "M" + maxInt64Prev + " mod " + maxInt64 + "0", length: 65,
		},
		{
			name: "x-64", path: "Parse",
			text: strings.Repeat("x", 64), length: 64,
		},
		{
			name: "x-65", path: "Parse",
			text: strings.Repeat("x", 65), length: 65,
		},
	}

	for _, c := range cases {
		p := pathByName(t, c.path)

		for _, viaMust := range []bool{false, true} {
			t.Run(c.name+"/"+callName(viaMust), func(t *testing.T) {
				t.Parallel()

				require.Len(t, c.text, c.length)

				zero, err := p.outcome(t, c.text, viaMust)
				assertRejection(t, c.name, c.text, zero, err)
			})
		}
	}
}

// TestNotationErrorsIgnoreText checks the order of the checks of
// ParseNotation on a long text (T4) and that the errors of the modulus and of
// the style do not depend on the text (T5): the answer for a long wrong text
// equals the answers of the same call for the valid number "1" and for the
// empty text, so the modulus and the style are checked before the text.
func TestNotationErrorsIgnoreText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		style   pitchclass.Style
		modulus int
		want    error
	}{
		{name: "modulus-0", style: pitchclass.Decimal, modulus: 0, want: pitchclass.ErrInvalidModulus},
		{name: "modulus-negative", style: pitchclass.Decimal, modulus: -1, want: pitchclass.ErrInvalidModulus},
		{name: "style-9", style: pitchclass.Style(9), modulus: 12, want: pitchclass.ErrNotation},
		{name: "TE-13", style: pitchclass.TE, modulus: 13, want: pitchclass.ErrNotation},
		{name: "LowerTE-13", style: pitchclass.LowerTE, modulus: 13, want: pitchclass.ErrNotation},
		{name: "AB-13", style: pitchclass.AB, modulus: 13, want: pitchclass.ErrNotation},
	}

	// The fillers are built once and only read by the subtests.
	fillers := letterFillers()

	for _, c := range cases {
		for _, f := range fillers {
			t.Run(c.name+"/"+f.name, func(t *testing.T) {
				t.Parallel()

				c0, err := pitchclass.ParseNotation(f.text, c.style, c.modulus)
				require.Error(t, err)
				assert.True(t, c0 == pitchclass.Class{}, "the result must be zero")
				assertOnlySentinel(t, err, c.want, "long text")
				assert.LessOrEqual(t, len(err.Error()), excerptErrorLimit, "T1, len(err.Error())")

				for _, ref := range []string{"1", ""} {
					c1, errRef := pitchclass.ParseNotation(ref, c.style, c.modulus)
					require.Error(t, errRef)
					assert.True(t, c1 == pitchclass.Class{}, "the result must be zero for %q", ref)
					assertOnlySentinel(t, errRef, c.want, fmt.Sprintf("reference text %q", ref))

					assert.True(t, err.Error() == errRef.Error(),
						"T5: the error must not depend on the text: reference %q, len(err)=%d, len(reference err)=%d",
						ref, len(err.Error()), len(errRef.Error()))
				}
			})
		}
	}
}

// TestParseOperatorNotInvertible checks that a text of the form with a
// multiplier that is not invertible is still reported as ErrNotInvertible
// only, not as ErrNotation, and that the text of this error stays bounded.
func TestParseOperatorNotInvertible(t *testing.T) {
	t.Parallel()

	const s = "T0M2 mod 12"

	f, err := pitchclass.ParseOperator(s)
	assert.True(t, f == pitchclass.Operator{}, "the result must be zero")
	assertOnlySentinel(t, err, pitchclass.ErrNotInvertible, s)
	assert.LessOrEqual(t, len(err.Error()), excerptErrorLimit, "T1, len(err.Error())")

	assertPanicsWithErrorIs(t, pitchclass.ErrNotInvertible, func() { pitchclass.MustParseOperator(s) }, s)
}
