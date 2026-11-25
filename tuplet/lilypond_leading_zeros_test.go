package tuplet_test

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests in this file cover leading zeros in the counts of the LilyPond
// tuplet marker and duration multiplier (issue #162; specification tuplet,
// O-20, O-21, P-10, E-4, E-5, G-11, KP-21, KP-22, K-18; architecture 4.14,
// rule ly-count). A count is ASCII digits without a sign or spaces; leading
// zeros are accepted and do not change the value; a zero value and an empty
// count are ErrInvalidText; the range check is by value, not by the length of
// the text. The formatters keep writing the counts without leading zeros, and
// the canonical grammar of Parse keeps rejecting leading zeros.

// zeros returns n ASCII zeros.
func zeros(n int) string {
	return strings.Repeat("0", n)
}

// The tuplet markers of KP-21 for issue #162 and markers of the same kind:
// each reads as the same text without leading zeros, keeps the pair as
// written (no reduction), and is written back without leading zeros.
func TestParseLilyPondTuplet_LeadingZeros(t *testing.T) {
	t.Parallel()

	maxText := strconv.Itoa(math.MaxInt)

	tests := []struct {
		name   string
		text   string
		actual int
		normal int
		format string
	}{
		{"actual", `\tuplet 03/2`, 3, 2, `\tuplet 3/2`},
		{"normal", `\tuplet 3/02`, 3, 2, `\tuplet 3/2`},
		{"both", `\tuplet 003/002`, 3, 2, `\tuplet 3/2`},
		{"not reduced", `\tuplet 006/004`, 6, 4, `\tuplet 6/4`},
		{"thirty zeros before actual", `\tuplet ` + zeros(30) + `3/2`, 3, 2, `\tuplet 3/2`},
		{"thirty zeros before normal", `\tuplet 3/` + zeros(30) + `2`, 3, 2, `\tuplet 3/2`},
		{"largest int actual", `\tuplet 0` + maxText + `/1`, math.MaxInt, 1, `\tuplet ` + maxText + `/1`},
		{"largest int normal", `\tuplet 1/0` + maxText, 1, math.MaxInt, `\tuplet 1/` + maxText},
		{"inner zeros kept", `\tuplet 0100/0010`, 100, 10, `\tuplet 100/10`},
		{"one over one", `\tuplet 01/01`, 1, 1, `\tuplet 1/1`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseLilyPondTuplet(tt.text)
			require.NoError(t, err)
			assertTuplet(t, got, tt.actual, tt.normal)

			// The result equals the result of the text without leading zeros.
			ref, err := tuplet.ParseLilyPondTuplet(tt.format)
			require.NoError(t, err, "the form %q is the reference", tt.format)
			assert.Equal(t, ref, got)

			if assertNoPanic(t, func() { tuplet.MustParseLilyPondTuplet(tt.text) }) {
				assert.Equal(t, got, tuplet.MustParseLilyPondTuplet(tt.text))
			}

			formatted, err := tuplet.FormatLilyPondTuplet(got)
			require.NoError(t, err)
			assert.Equal(t, tt.format, formatted, "the counts are written without leading zeros")

			back, err := tuplet.ParseLilyPondTuplet(formatted)
			require.NoError(t, err)
			assert.Equal(t, got, back)
		})
	}
}

// The multipliers of KP-21 for issue #162 and multipliers of the same kind:
// each reads as the same text without leading zeros, keeps the source pair
// by value (no reduction, (a, 1) for the integer form), and is written back
// without leading zeros.
func TestParseLilyPondMultiplier_LeadingZeros(t *testing.T) {
	t.Parallel()

	maxText := strconv.Itoa(math.MaxInt)

	tests := []struct {
		name   string
		text   string
		num    int
		den    int
		format string
	}{
		{"numerator", "*02/3", 2, 3, "*2/3"},
		{"denominator", "*2/03", 2, 3, "*2/3"},
		{"both", "*002/003", 2, 3, "*2/3"},
		{"integer form", "*04", 4, 1, "*4"},
		{"denominator one", "*04/01", 4, 1, "*4"},
		{"thirty zeros before the largest int", "*" + zeros(30) + maxText, math.MaxInt, 1, "*" + maxText},
		{"largest int denominator", "*1/0" + maxText, 1, math.MaxInt, "*1/" + maxText},
		{"not reduced", "*04/06", 4, 6, "*4/6"},
		{"inner zeros kept", "*010/0100", 10, 100, "*10/100"},
		{"one", "*01", 1, 1, "*1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseLilyPondMultiplier(tt.text)
			require.NoError(t, err)
			assertPair(t, got, tt.num, tt.den)

			// The result equals the result of the text without leading zeros,
			// with the same pair.
			ref, err := tuplet.ParseLilyPondMultiplier(tt.format)
			require.NoError(t, err, "the form %q is the reference", tt.format)
			assert.True(t, ref.Equal(got), "%s and %s", ref, got)
			assertPair(t, ref, tt.num, tt.den)

			if assertNoPanic(t, func() { tuplet.MustParseLilyPondMultiplier(tt.text) }) {
				assertPair(t, tuplet.MustParseLilyPondMultiplier(tt.text), tt.num, tt.den)
			}

			formatted, err := tuplet.FormatLilyPondMultiplier(got)
			require.NoError(t, err)
			assert.Equal(t, tt.format, formatted, "the counts are written without leading zeros")

			back, err := tuplet.ParseLilyPondMultiplier(formatted)
			require.NoError(t, err)
			assertPair(t, back, tt.num, tt.den)
		})
	}
}

// G-11: a long run of zeros before a count does not change its value and does
// not give ErrOutOfRange, in every count position of both operations, up to
// a run of 1 MiB.
func TestLilyPondLongLeadingZeros(t *testing.T) {
	t.Parallel()

	for _, n := range []int{30, kib, mib} {
		z := zeros(n)

		t.Run(itoa(n), func(t *testing.T) {
			t.Parallel()

			for _, text := range []string{`\tuplet ` + z + `3/2`, `\tuplet 3/` + z + `2`, `\tuplet ` + z + `3/` + z + `2`} {
				got, err := tuplet.ParseLilyPondTuplet(text)
				if assert.NoError(t, err, "a run of %d zeros", n) {
					assert.Equal(t, tuplet.Triplet(), got)
				}
			}

			for _, text := range []string{"*" + z + "2/3", "*2/" + z + "3", "*" + z + "2/" + z + "3"} {
				got, err := tuplet.ParseLilyPondMultiplier(text)
				if assert.NoError(t, err, "a run of %d zeros", n) {
					assertPair(t, got, 2, 3)
				}
			}

			got, err := tuplet.ParseLilyPondMultiplier("*" + z + "4")
			if assert.NoError(t, err, "a run of %d zeros", n) {
				assertPair(t, got, 4, 1)
			}
		})
	}
}

// G-11: the largest int with leading zeros is accepted and the next value
// gives ErrOutOfRange alone (without a common/fraction cause), in every count
// position; the check is by value, not by the length of the text.
func TestLilyPondLeadingZerosRange(t *testing.T) {
	t.Parallel()

	maxText := strconv.Itoa(math.MaxInt)
	const beyond = "9223372036854775808"

	for _, n := range []int{1, 30, kib} {
		z := zeros(n)

		t.Run(itoa(n), func(t *testing.T) {
			t.Parallel()

			tp, err := tuplet.ParseLilyPondTuplet(`\tuplet ` + z + maxText + `/` + z + maxText)
			if assert.NoError(t, err) {
				assertTuplet(t, tp, math.MaxInt, math.MaxInt)
			}

			c, err := tuplet.ParseLilyPondMultiplier("*" + z + maxText + "/" + z + maxText)
			if assert.NoError(t, err) {
				assertPair(t, c, math.MaxInt, math.MaxInt)
			}

			c, err = tuplet.ParseLilyPondMultiplier("*" + z + maxText)
			if assert.NoError(t, err) {
				assertPair(t, c, math.MaxInt, 1)
			}

			for _, text := range []string{`\tuplet ` + z + beyond + `/2`, `\tuplet 3/` + z + beyond} {
				got, err := tuplet.ParseLilyPondTuplet(text)
				assertErr(t, err, tuplet.ErrOutOfRange)
				assert.Equal(t, tuplet.Tuplet{}, got)
			}

			for _, text := range []string{"*" + z + beyond, "*" + z + beyond + "/3", "*2/" + z + beyond} {
				got, err := tuplet.ParseLilyPondMultiplier(text)
				assertErr(t, err, tuplet.ErrOutOfRange)
				assert.Equal(t, fraction.Fraction{}, got)
			}
		})
	}
}

// A zero value is ErrInvalidText with or without leading zeros, including a
// long run of zeros, and so is an empty count (P-10, E-4, G-11, KP-22).
func TestLilyPondZeroAndEmptyCounts(t *testing.T) {
	t.Parallel()

	tupletTexts := []string{
		`\tuplet 0/2`, `\tuplet 00/2`, `\tuplet ` + zeros(30) + `/2`, `\tuplet ` + zeros(kib) + `/2`,
		`\tuplet 3/0`, `\tuplet 3/00`, `\tuplet 3/` + zeros(30), `\tuplet 00/00`,
		`\tuplet /2`, `\tuplet 3/`, `\tuplet /`, `\tuplet `,
	}

	multiplierTexts := []string{
		"*0", "*00", "*" + zeros(30), "*" + zeros(kib),
		"*0/3", "*00/3", "*2/0", "*2/00", "*2/" + zeros(30), "*0/0", "*00/00",
		"*", "*/3", "*2/", "*/",
	}

	for _, text := range tupletTexts {
		t.Run("tuplet/"+strconv.Quote(text), func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseLilyPondTuplet(text)
			assertErr(t, err, tuplet.ErrInvalidText)
			assert.Equal(t, tuplet.Tuplet{}, got)

			_ = assertPanicErr(t, func() { tuplet.MustParseLilyPondTuplet(text) }, tuplet.ErrInvalidText)
		})
	}

	for _, text := range multiplierTexts {
		t.Run("multiplier/"+strconv.Quote(text), func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseLilyPondMultiplier(text)
			assertErr(t, err, tuplet.ErrInvalidText)
			assert.Equal(t, fraction.Fraction{}, got)

			_ = assertPanicErr(t, func() { tuplet.MustParseLilyPondMultiplier(text) }, tuplet.ErrInvalidText)
		})
	}
}

// Leading zeros do not relax the rest of the grammar: signs, spaces, other
// separators and digits other than ASCII stay ErrInvalidText (P-10, E-4,
// KP-22).
func TestLilyPondLeadingZerosOtherTextRejected(t *testing.T) {
	t.Parallel()

	tupletTexts := []string{
		`\tuplet +03/2`, `\tuplet -03/2`, `\tuplet 03/+02`, `\tuplet 03/-02`,
		`\tuplet 0 3/2`, `\tuplet 03 /2`, `\tuplet 03/ 2`, `\tuplet  03/2`,
		`\tuplet 03/2 `, `\tuplet 03/02 4`, `\tuplet 03:2`, `\tuplet 03//2`,
		`\tuplet ٠3/2`, `\tuplet 0x3/2`, ` \tuplet 03/2`, `\Tuplet 03/2`, `\tuplet03/2`,
	}

	multiplierTexts := []string{
		"*+02/3", "*-02/3", "*02/-3", "*02/+3",
		"* 02/3", "*02 /3", "*02/ 3", "*0 2/3",
		"*02/3 ", " *02/3", "*02:3", "*02//3",
		"*٠2/3", "*0x2/3", "02/3", "**02/3", "*04/",
	}

	for _, text := range tupletTexts {
		t.Run("tuplet/"+strconv.Quote(text), func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseLilyPondTuplet(text)
			assertErr(t, err, tuplet.ErrInvalidText)
			assert.Equal(t, tuplet.Tuplet{}, got)

			if err != nil {
				assert.Contains(t, err.Error(), strconv.Quote(text), "the error quotes the short input in full")
			}
		})
	}

	for _, text := range multiplierTexts {
		t.Run("multiplier/"+strconv.Quote(text), func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseLilyPondMultiplier(text)
			assertErr(t, err, tuplet.ErrInvalidText)
			assert.Equal(t, fraction.Fraction{}, got)

			if err != nil {
				assert.Contains(t, err.Error(), strconv.Quote(text), "the error quotes the short input in full")
			}
		})
	}
}

// The order of the parse checks is the grammar of the whole text (E-4), then
// the range by value (E-5) (K-18, architecture 4.2).
func TestLilyPondLeadingZerosCheckOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(string) error
		text string
		want error
	}{
		{"tuplet zero normal after a big actual", parseTupletErr, `\tuplet 099999999999999999999/00`, tuplet.ErrInvalidText},
		{"tuplet empty normal after a big actual", parseTupletErr, `\tuplet 099999999999999999999/`, tuplet.ErrInvalidText},
		{"tuplet span after a big actual", parseTupletErr, `\tuplet 099999999999999999999/2 4`, tuplet.ErrInvalidText},
		{"tuplet big actual", parseTupletErr, `\tuplet 099999999999999999999/2`, tuplet.ErrOutOfRange},
		{"tuplet big actual without zeros", parseTupletErr, `\tuplet 99999999999999999999/2`, tuplet.ErrOutOfRange},
		{"multiplier zero denominator after a big numerator", parseMultiplierErr, "*099999999999999999999/0", tuplet.ErrInvalidText},
		{"multiplier empty denominator after a big numerator", parseMultiplierErr, "*099999999999999999999/", tuplet.ErrInvalidText},
		{"multiplier sign after a big numerator", parseMultiplierErr, "*099999999999999999999/-3", tuplet.ErrInvalidText},
		{"multiplier big integer form", parseMultiplierErr, "*09223372036854775808", tuplet.ErrOutOfRange},
		{"multiplier big integer form without zeros", parseMultiplierErr, "*9223372036854775808", tuplet.ErrOutOfRange},
		{"multiplier big denominator", parseMultiplierErr, "*2/09223372036854775808", tuplet.ErrOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.run(tt.text)
			assertErr(t, err, tt.want)

			if err != nil {
				assert.Contains(t, err.Error(), strconv.Quote(tt.text), "the error quotes the short input in full")
			}
		})
	}
}

// parseTupletErr runs ParseLilyPondTuplet and returns its error.
func parseTupletErr(text string) error {
	_, err := tuplet.ParseLilyPondTuplet(text)

	return err
}

// parseMultiplierErr runs ParseLilyPondMultiplier and returns its error.
func parseMultiplierErr(text string) error {
	_, err := tuplet.ParseLilyPondMultiplier(text)

	return err
}

// The error text of a rejected count with a long run of leading zeros obeys
// architecture 4.12: it contains the quoted excerpt of the first 64 bytes
// followed by "...", is at most 512 bytes long, and depends on the input only
// through its first 67 bytes, so inputs of 128 bytes, 1 KiB and 1 MiB with
// the same prefix and the same error class give the same text.
func TestLilyPondLeadingZerosErrorText(t *testing.T) {
	t.Parallel()

	const beyond = "9223372036854775808"

	families := []struct {
		name  string
		build func(n int) string
		run   func(string) error
		want  error
	}{
		{"tuplet beyond MaxInt", func(n int) string { return `\tuplet ` + zeros(n-len(`\tuplet `)-len(beyond)-2) + beyond + "/2" }, parseTupletErr, tuplet.ErrOutOfRange},
		{"tuplet zero", func(n int) string { return `\tuplet ` + zeros(n-len(`\tuplet `)-2) + "/2" }, parseTupletErr, tuplet.ErrInvalidText},
		{"multiplier beyond MaxInt", func(n int) string { return "*" + zeros(n-1-len(beyond)) + beyond }, parseMultiplierErr, tuplet.ErrOutOfRange},
		{"multiplier zero", func(n int) string { return "*" + zeros(n-1) }, parseMultiplierErr, tuplet.ErrInvalidText},
	}

	for _, f := range families {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()

			var reference string

			for i, n := range []int{128, kib, mib} {
				x := f.build(n)
				require.Len(t, x, n)

				err := f.run(x)
				if !assertErr(t, err, f.want) {
					continue
				}

				text := err.Error()
				assert.Contains(t, text, strconv.Quote(x[:excerptLimit])+"...", "the error quotes the excerpt")
				assert.LessOrEqual(t, len(text), errorTextLimit)

				if i == 0 {
					reference = text

					continue
				}

				assert.Equal(t, reference, text, "the text depends only on the first 67 bytes")
			}
		})
	}
}

// The canonical grammar does not change (P-7, K-11, architecture 4.12):
// Parse, ParseShort and UnmarshalText keep rejecting leading zeros, and the
// LilyPond rule ly-count does not apply to them.
func TestCanonicalGrammarKeepsRejectingLeadingZeros(t *testing.T) {
	t.Parallel()

	for _, text := range []string{"03:2", "3:02", "003:002", "00:2", "3:00", "09223372036854775807:1"} {
		t.Run(strconv.Quote(text), func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.Parse(text)
			assertErr(t, err, tuplet.ErrInvalidText)
			assert.Equal(t, tuplet.Tuplet{}, got)

			target := tuplet.Quintuplet()
			err = target.UnmarshalText([]byte(text))
			assertErr(t, err, tuplet.ErrInvalidText)
			assert.Equal(t, tuplet.Quintuplet(), target, "the receiver is unchanged")
		})
	}

	for _, text := range []string{"03", "003", "00"} {
		t.Run("short/"+strconv.Quote(text), func(t *testing.T) {
			t.Parallel()

			got, err := tuplet.ParseShort(text, tuplet.Binary)
			assertErr(t, err, tuplet.ErrInvalidText)
			assert.Equal(t, tuplet.Tuplet{}, got)
		})
	}
}

// FuzzLilyPondLeadingZeros covers K-18 and G-11 on arbitrary counts and runs
// of leading zeros: a tuplet marker or multiplier with leading zeros reads
// exactly as the same text without them, and is written back without them.
func FuzzLilyPondLeadingZeros(f *testing.F) {
	f.Add(3, 2, uint8(1), uint8(0))
	f.Add(6, 4, uint8(2), uint8(2))
	f.Add(1, 1, uint8(30), uint8(30))
	f.Add(math.MaxInt, 1, uint8(1), uint8(5))
	f.Add(10, 100, uint8(0), uint8(1))

	f.Fuzz(func(t *testing.T, a, b int, za, zb uint8) {
		if a < 1 || b < 1 {
			t.Skip("outside the positive counts of the grammar")
		}

		aText, bText := strconv.Itoa(a), strconv.Itoa(b)
		zeroA, zeroB := zeros(int(za)), zeros(int(zb))

		plainTuplet := `\tuplet ` + aText + "/" + bText
		got, err := tuplet.ParseLilyPondTuplet(`\tuplet ` + zeroA + aText + "/" + zeroB + bText)
		require.NoError(t, err)
		assertTuplet(t, got, a, b)
		assert.Equal(t, plainTuplet, tuplet.MustFormatLilyPondTuplet(got))

		c, err := tuplet.ParseLilyPondMultiplier("*" + zeroA + aText + "/" + zeroB + bText)
		require.NoError(t, err)
		assertPair(t, c, a, b)

		want := "*" + aText + "/" + bText
		if b == 1 {
			want = "*" + aText
		}

		assert.Equal(t, want, tuplet.MustFormatLilyPondMultiplier(c))

		whole, err := tuplet.ParseLilyPondMultiplier("*" + zeroA + aText)
		require.NoError(t, err)
		assertPair(t, whole, a, 1)
	})
}
