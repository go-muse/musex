package meter_test

import (
	"encoding/json/v2"
	"math"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// textCases are the canonical texts of architecture 4.15 for КП-26.
var textCases = []struct{ in, text string }{
	{"3/4", "3/4"},
	{"6/8", "6/8"},
	{"C", "C"},
	{"C|", "C|"},
	{"C[3+1]", "C[3+1]"},
	{"C|[1+1]", "C|[1+1]"},
	{"(3+2+3)/8", "(3+2+3)/8"},
	{"2/4 + 3/8", "2/4+3/8"},
	{"5/8 [2+3]", "5/8[2+3]"},
	{"(2+3)/8 + 2/4", "(2+3)/8+2/4"},
	{"5/8", "5/8"},
	{"6/8 [3+3]", "6/8"},
	{"6/8 [2+2+2]", "6/8[2+2+2]"},
	{"4/3", "4/3"},
	{"1024/8", "1024/8"},
	{"2/8+(1+1)/8+3/8", "2/8+(1+1)/8+3/8"},
}

// TestString covers О-19, Т-11, КП-26, КП-47 and the output of architecture
// 4.15: the canonical text without spaces, with C| for ¢.
func TestString(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range textCases {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m := mt(c.in)
			assert.Equal(t, c.text, m.String())

			data, err := m.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, c.text, string(data))

			appended, err := m.AppendText([]byte("x:"))
			require.NoError(t, err)
			assert.Equal(t, "x:"+c.text, string(appended))
		})
	}

	t.Run("КП-47 unset", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		var u meter.Meter
		assert.Equal(t, "Meter(unset)", u.String())

		data, err := u.MarshalText()
		assertErr(t, err, meter.ErrUnset)
		assert.Nil(t, data)

		dst := []byte("x:")
		appended, err := u.AppendText(dst)
		assertErr(t, err, meter.ErrUnset)
		assert.Equal(t, "x:", string(appended), "dst without changes")
	})

	t.Run("large denominator", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		assert.Equal(t, "1/"+maxIntText(), meter.MustNew(1, math.MaxInt).String())
	})
}

// TestParse covers О-19, Т-11, КП-26 and the grammar of architecture 4.15:
// Parse reverses String, accepts ¢, and accepts an explicit grouping equal
// to the standard one.
func TestParse(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range textCases {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := meter.Parse(c.text)
			require.NoError(t, err)
			assert.Equal(t, mt(c.in), got)
		})
	}

	for _, c := range []struct{ text, want string }{
		{"¢", "C|"},
		{"¢[1+1]", "C|[1+1]"},
		{"¢[2]", "C|"},
		{"6/8[3+3]", "6/8"},
		{"C[2+2]", "C"},
		{"C|[2]", "C|"},
		{"4/4[2+2]", "4/4"},
		{"3/4[3]", "3/4"},
		{"1/" + maxIntText(), "1/" + maxIntText()},
		{"2/4+(2+3)/8", "2/4+(2+3)/8"},
		{"5/8[5]", "5/8[5]"},
		{"1024/8[1024]", "1024/8[1024]"},
	} {
		t.Run("accepts "+c.text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := meter.Parse(c.text)
			require.NoError(t, err)
			assert.Equal(t, mt(c.want), got)
		})
	}
}

// TestParseErrors covers КП-28, КП-37, КП-39, КП-52 and the order of checks
// of architecture 4.15: E-9 → E-10 (denominator) → construction (E-8, E-10,
// E-2). Parse errors wrap only the package sentinel and the named
// dependency sentinels.
func TestParseErrors(t *testing.T) {
	t.Parallel()
	defer guard(t)

	const huge = "99999999999999999999" // 20 digits, greater than math.MaxInt

	overMaxInt := "9223372036854775808" // math.MaxInt + 1

	for _, c := range []struct {
		name, text string
		want       error
		causes     []error
	}{
		{"empty", "", meter.ErrInvalidText, nil},
		{"leading space", " 3/4", meter.ErrInvalidText, nil},
		{"trailing space", "3/4 ", meter.ErrInvalidText, nil},
		{"space inside", "3 /4", meter.ErrInvalidText, nil},
		{"spaces around plus", "2/4 + 3/8", meter.ErrInvalidText, nil},
		{"leading zero", "03/4", meter.ErrInvalidText, nil},
		{"leading zero in denominator", "3/04", meter.ErrInvalidText, nil},
		{"zero numerator", "0/4", meter.ErrInvalidText, nil},
		{"zero denominator", "3/0", meter.ErrInvalidText, nil},
		{"sign", "+3/4", meter.ErrInvalidText, nil},
		{"negative", "-3/4", meter.ErrInvalidText, nil},
		{"lower case symbol", "c", meter.ErrInvalidText, nil},
		{"lower case cut", "c|", meter.ErrInvalidText, nil},
		{"double bar", "C||", meter.ErrInvalidText, nil},
		{"¢ with a bar", "¢|", meter.ErrInvalidText, nil},
		{"name", "common time", meter.ErrInvalidText, nil},
		{"one addend", "(3)/8", meter.ErrInvalidText, nil},
		{"empty parentheses", "()/8", meter.ErrInvalidText, nil},
		{"dangling plus in parentheses", "(3+)/8", meter.ErrInvalidText, nil},
		{"grouping of a compound numerator", "(2+3)/8[2+3]", meter.ErrInvalidText, nil},
		{"symbol in a sum", "2/4+C", meter.ErrInvalidText, nil},
		{"symbol first in a sum", "C+2/4", meter.ErrInvalidText, nil},
		{"grouping of a sum part", "5/8[2+3]+1/8", meter.ErrInvalidText, nil},
		{"grouping after a sum", "2/4+3/8[2+3]", meter.ErrInvalidText, nil},
		{"empty grouping", "5/8[]", meter.ErrInvalidText, nil},
		{"unclosed grouping", "5/8[2+3", meter.ErrInvalidText, nil},
		{"zero group", "5/8[0+5]", meter.ErrInvalidText, nil},
		{"leading zero in a group", "5/8[05]", meter.ErrInvalidText, nil},
		{"trailing plus", "2/4+", meter.ErrInvalidText, nil},
		{"leading plus", "+2/4", meter.ErrInvalidText, nil},
		{"double plus", "2/4++3/8", meter.ErrInvalidText, nil},
		{"no denominator", "3/", meter.ErrInvalidText, nil},
		{"no numerator", "/4", meter.ErrInvalidText, nil},
		{"number only", "3", meter.ErrInvalidText, nil},
		{"fraction numerator", "2.5/4", meter.ErrInvalidText, nil},
		{"unset string", "Meter(unset)", meter.ErrInvalidText, nil},
		{"full-width digit", "３/4", meter.ErrInvalidText, nil},
		{"invalid UTF-8", "3/4\xff", meter.ErrInvalidText, nil},
		{"КП-28 1025/8", "1025/8", meter.ErrTooMany, nil},
		{"КП-28 (1000+25)/8", "(1000+25)/8", meter.ErrTooMany, nil},
		{"КП-28 1024/8+1/4", "1024/8+1/4", meter.ErrTooMany, nil},
		{"КП-52 numerator of 20 digits", huge + "/4", meter.ErrTooMany, nil},
		{"numerator above MaxInt", overMaxInt + "/4", meter.ErrTooMany, nil},
		{"addend above MaxInt", "(" + overMaxInt + "+1)/4", meter.ErrTooMany, nil},
		{"numerator of 100 digits", strings.Repeat("9", 100) + "/4", meter.ErrTooMany, nil},
		{"КП-52 denominator of 20 digits", "3/" + huge, meter.ErrOutOfRange, nil},
		{"denominator above MaxInt", "1/" + overMaxInt, meter.ErrOutOfRange, nil},
		{"order: denominator before numerator", huge + "/" + huge, meter.ErrOutOfRange, nil},
		{"order: denominator before grouping", "3/" + huge + "[1+1]", meter.ErrOutOfRange, nil},
		{"order: first denominator from the left", "1025/8+3/" + huge, meter.ErrOutOfRange, nil},
		{"КП-39 sum out of range", kp39Text(), meter.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"unit start out of range, size in range", sumText(unitStartInput()), meter.ErrOutOfRange,
			[]error{fraction.ErrOutOfRange}},
		{"order: E-8 before E-10 of the sum", kp39TooManyText(), meter.ErrTooMany, nil},
		{"wrong grouping", "5/8[2+2]", meter.ErrInvalidArgument, nil},
		{"grouping of the symbol", "C[3+2]", meter.ErrInvalidArgument, nil},
		{"group above MaxInt", "5/8[" + huge + "+1]", meter.ErrInvalidArgument, nil},
		{"order: E-8 before grouping", "1025/8[2+3]", meter.ErrTooMany, nil},
		{"order: E-9 before E-8", "1025/8 ", meter.ErrInvalidText, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m, err := meter.Parse(c.text)
			assertErrExact(t, err, c.want, c.causes...)
			assert.Equal(t, meter.Meter{}, m)

			var got meter.Meter

			uerr := got.UnmarshalText([]byte(c.text))
			assertErrExact(t, uerr, c.want, c.causes...)
			assert.Equal(t, meter.Meter{}, got)

			if err != nil && uerr != nil {
				assert.Equal(t, err.Error(), uerr.Error(), "the same error text as Parse")
			}
		})
	}
}

// kp39TooManyText is a text of five parts with the denominators of КП-39 and
// 205 units each: 1025 units, E-8 before the range of the sum.
func kp39TooManyText() string {
	parts := make([]string, 0, 5)
	for _, d := range kp39Denominators() {
		parts = append(parts, "205/"+strconv.Itoa(d))
	}

	return strings.Join(parts, "+")
}

// TestUnmarshalText covers architecture 4.15: UnmarshalText reverses
// MarshalText, keeps the receiver on error, and a nil receiver gives E-2.
func TestUnmarshalText(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range textCases {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m := mt(c.in)
			data := noErr(m.MarshalText())

			var got meter.Meter
			require.NoError(t, got.UnmarshalText(data))
			assert.Equal(t, m, got)
		})
	}

	t.Run("the receiver is kept on error", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		got := meter.SixEight()
		err := got.UnmarshalText([]byte("6/8 "))
		assertErrExact(t, err, meter.ErrInvalidText)
		assert.Equal(t, meter.SixEight(), got)
	})

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		var p *meter.Meter

		err := p.UnmarshalText([]byte("3/4"))
		assertErr(t, err, meter.ErrInvalidArgument)

		err = p.UnmarshalText([]byte("x"))
		assertErr(t, err, meter.ErrInvalidArgument)
		assert.NotErrorIs(t, err, meter.ErrInvalidText, "order: E-2 before the grammar")
	})

	t.Run("replaces a set value", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		got := meter.SixEight()
		require.NoError(t, got.UnmarshalText([]byte("C|")))
		assert.Equal(t, meter.CutTime(), got)
	})
}

// TestJSON covers architecture 4.15: the meter is a JSON string, and an
// unset field is omitted with omitzero.
func TestJSON(t *testing.T) {
	t.Parallel()
	defer guard(t)

	type bar struct {
		Meter    meter.Meter `json:"meter"`
		Optional meter.Meter `json:"optional,omitzero"`
	}

	data, err := json.Marshal(bar{Meter: mt("5/8[2+3]")})
	require.NoError(t, err)
	assert.JSONEq(t, `{"meter":"5/8[2+3]"}`, string(data))

	var back bar
	require.NoError(t, json.Unmarshal([]byte(`{"meter":"2/4+3/8","optional":"¢"}`), &back))
	assert.Equal(t, mt("2/4+3/8"), back.Meter)
	assert.Equal(t, meter.CutTime(), back.Optional)

	_, err = json.Marshal(struct{ M meter.Meter }{})
	require.Error(t, err)
	assert.ErrorIs(t, err, meter.ErrUnset)
}

// excerpt is the expected quotation of a rejected input x by architecture
// 4.15: the whole x for up to 64 bytes, otherwise the longest prefix of
// whole UTF-8 decoding units within 64 bytes and "..." after the quotes.
func excerpt(x string) string {
	if len(x) <= 64 {
		return strconv.Quote(x)
	}

	k := 0

	for k < len(x) {
		_, size := utf8.DecodeRuneInString(x[k:])
		if k+size > 64 {
			break
		}

		k += size
	}

	return strconv.Quote(x[:k]) + "..."
}

// TestParseErrorText covers the limits of the text of parse errors in
// architecture 4.15: the text quotes the excerpt, is at most 512 bytes, and
// for one class of error depends only on the first 67 bytes of the input.
func TestParseErrorText(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct {
		name, text string
		k          int // the expected length of the quoted prefix, 0 for the whole input
	}{
		{"short", "3/4x", 0},
		{"64 bytes", "3/4" + strings.Repeat("x", 61), 0},
		{"65 bytes", "3/4" + strings.Repeat("x", 62), 64},
		{"rune across the limit", strings.Repeat("a", 63) + "¢" + strings.Repeat("z", 10), 63},
		{"invalid byte at the limit", strings.Repeat("a", 63) + "\xff" + strings.Repeat("z", 10), 64},
		{"truncated rune", strings.Repeat("a", 62) + "\xe2\x82" + strings.Repeat("z", 10), 64},
		{"two-byte runes", strings.Repeat("¢", 40), 64},
		{"three-byte runes", strings.Repeat("€", 30), 63},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			_, err := meter.Parse(c.text)
			require.Error(t, err)

			msg := err.Error()
			assert.Contains(t, msg, excerpt(c.text))
			assert.LessOrEqual(t, len(msg), 512)

			if c.k > 0 {
				assert.Contains(t, msg, strconv.Quote(c.text[:c.k])+"...")
				assert.NotContains(t, msg, strconv.Quote(c.text[:c.k+1]))
			}

			var m meter.Meter

			uerr := m.UnmarshalText([]byte(c.text))
			require.Error(t, uerr)
			assert.Contains(t, uerr.Error(), excerpt(c.text))
		})
	}

	prefix := "3/4" + strings.Repeat("y", 64) // 67 bytes

	for _, c := range []struct {
		name   string
		build  func(n int) string
		want   error
		causes []error
	}{
		{"E-9", func(n int) string { return prefix + strings.Repeat("z", n-len(prefix)) }, meter.ErrInvalidText, nil},
		{"E-8 long numerator", func(n int) string { return "1" + strings.Repeat("0", n-3) + "/4" }, meter.ErrTooMany, nil},
		{"E-10 long denominator", func(n int) string { return "1/1" + strings.Repeat("0", n-3) }, meter.ErrOutOfRange, nil},
	} {
		t.Run("class "+c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			small, large := c.build(1<<10), c.build(1<<20)
			require.Len(t, small, 1<<10)
			require.Len(t, large, 1<<20)
			require.Equal(t, small[:67], large[:67])

			rs := fast(t, func() result[meter.Meter] { m, err := meter.Parse(small); return result[meter.Meter]{m, err} })
			rl := fast(t, func() result[meter.Meter] { m, err := meter.Parse(large); return result[meter.Meter]{m, err} })

			assertErrExact(t, rs.err, c.want, c.causes...)
			assertErrExact(t, rl.err, c.want, c.causes...)

			if rs.err == nil || rl.err == nil {
				return
			}

			assert.Equal(t, rs.err.Error(), rl.err.Error(), "the text depends only on the first 67 bytes")
			assert.LessOrEqual(t, len(rl.err.Error()), 512)
			assert.Contains(t, rl.err.Error(), excerpt(large))
			assert.NotContains(t, rl.err.Error(), strconv.Itoa(len(large)), "the length is not printed")
		})
	}
}

// TestTextRoundTrip covers О-19, Т-11, КП-26 on the corpus: Parse reverses
// String and UnmarshalText reverses MarshalText; texts of meters different by
// record are different.
func TestTextRoundTrip(t *testing.T) {
	t.Parallel()
	defer guard(t)

	seen := map[string]meter.Meter{}

	for _, m := range corpus() {
		text := m.String()

		got, err := meter.Parse(text)
		require.NoError(t, err, text)
		assert.Equal(t, m, got, text)

		var u meter.Meter
		require.NoError(t, u.UnmarshalText(noErr(m.MarshalText())), text)
		assert.Equal(t, m, u, text)

		if prev, ok := seen[text]; ok {
			assert.Equal(t, prev, m, "one text for different records: %s", text)
		}

		seen[text] = m
	}
}
