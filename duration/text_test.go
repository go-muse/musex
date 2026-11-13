package duration_test

import (
	jsonv1 "encoding/json"
	jsonv2 "encoding/json/v2"
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// maxIntText is math.MaxInt in decimal; maxIntPlusText is one more.
var (
	maxIntText     = strconv.Itoa(math.MaxInt)
	maxIntPlusText = new(big.Int).Add(big.NewInt(math.MaxInt), big.NewInt(1)).String()
	maxCompText    = fraction.MaxComponentString
	maxCompPlus    = new(big.Int).Add(fraction.MaxComponentBig(), big.NewInt(1)).String()
)

// TestWrittenString covers architecture 4.12: the canonical text of written
// and tied durations.
func TestWrittenString(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		w    duration.Written
		want string
	}{
		{wr(duration.Quarter, 0), "4"},
		{wr(duration.Quarter, 1), "4."},
		{wr(duration.Quarter, 3), "4..."},
		{wr(duration.Eighth, 0, tuplet.Triplet()), "8[3:2]"},
		{wr(duration.Quarter, 1, tuplet.Quintuplet(), tuplet.Triplet()), "4.[5:4,3:2]"},
		{wr(duration.Quarter, 0, tp(6, 4)), "4[6:4]"},
		{wr(duration.Quarter, 0, tp(1, 1)), "4[1:1]"},
		{wr(duration.Breve, 2), "breve.."},
		{wr(duration.Longa, 0), "longa"},
		{wr(duration.Maxima, 1), "maxima."},
		{wr(duration.Whole, 0), "1"},
		{wr(duration.Sixteenth, 1), "16."},
		{wr(duration.ThousandTwentyFourth, 0, bigTuplet), "1024[" + maxIntText + ":1]"},
	} {
		assert.Equal(t, c.want, c.w.String())

		text, err := c.w.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, c.want, string(text))
	}

	for _, c := range []struct {
		t    duration.Tied
		want string
	}{
		{tie(wr(duration.Half, 0), wr(duration.Quarter, 0)), "2~4"},
		{tie(wr(duration.Quarter, 0, tuplet.Triplet()), wr(duration.Eighth, 0, tuplet.Triplet())), "4[3:2]~8[3:2]"},
		{tie(wr(duration.Breve, 0), wr(duration.Breve, 0), wr(duration.Whole, 0)), "breve~breve~1"},
		{tie(wr(duration.Quarter, 1), wr(duration.Sixteenth, 0)), "4.~16"},
	} {
		assert.Equal(t, c.want, c.t.String())

		text, err := c.t.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, c.want, string(text))
	}
}

// TestParse covers O-16, KP-36, and the grammar of architecture 4.12 for
// Parse and ParsePosition.
func TestParse(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		text string
		want duration.Duration
	}{
		{"3/8", dur(3, 8)},
		{"6/16", dur(3, 8)},
		{"1/1", dur(1, 1)},
		{"8/1", dur(8, 1)},
		{"4/2", dur(2, 1)},
		{"1/1024", dur(1, 1024)},
		{maxCompText + "/1", durRat(ratInt(maxComponent()))},
		{"1/" + maxCompText, durRat(new(big.Rat).SetFrac(big.NewInt(1), maxComponent()))},
		{maxCompText + "/" + maxCompText, dur(1, 1)},
	} {
		got, err := duration.Parse(c.text)
		require.NoError(t, err, c.text)
		assert.Equal(t, c.want, got, c.text)
	}

	for _, text := range []string{
		"", "3", "/", "3/", "/8", "0/1", "00/1", "0/8", "03/8", "3/08", "3/0", "3//8", "3/8/1", "-3/8", "+3/8", "3/-8",
		" 3/8", "3/8 ", "3 /8", "3/ 8", "3/8\n", "1.5", "1 1/2", "3:8", "Duration(unset)", "３/8", "3/8x", "x3/8", "4", "4.",
		"quarter", "1e3/1", "0x10/1", "1_0/1",
	} {
		got, err := duration.Parse(text)
		assertErr(t, err, duration.ErrInvalidText)
		assert.Equal(t, duration.Duration{}, got, "%q", text)
	}

	for _, text := range []string{maxCompPlus + "/1", "1/" + maxCompPlus, strings.Repeat("9", 78) + "/1", "1/" + strings.Repeat("9", 200), maxCompPlus + "/" + maxCompPlus} {
		got, err := duration.Parse(text)
		assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, duration.Duration{}, got)
	}

	for _, c := range []struct {
		text string
		want duration.Position
	}{
		{"0/1", duration.Origin()},
		{"0/7", duration.Origin()},
		{"3/8", pos(3, 8)},
		{"6/16", pos(3, 8)},
		{"2/1", pos(2, 1)},
		{"0/" + maxCompText, duration.Origin()},
	} {
		got, err := duration.ParsePosition(c.text)
		require.NoError(t, err, c.text)
		assert.Equal(t, c.want, got, c.text)
	}

	for _, text := range []string{
		"", "0", "00/1", "01/2", "-0/1", "0/0", "1/0", "-1/4", "+1/4", "0/01", " 0/1", "0/1 ", "Position(unset)", "0.0", "3/8x",
	} {
		got, err := duration.ParsePosition(text)
		assertErr(t, err, duration.ErrInvalidText)
		assert.Equal(t, duration.Position{}, got, "%q", text)
	}

	for _, text := range []string{maxCompPlus + "/1", "0/" + maxCompPlus} {
		got, err := duration.ParsePosition(text)
		assertErr(t, err, duration.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, duration.Position{}, got)
	}
}

// TestParseWritten covers the grammar of architecture 4.12 for written and
// tied durations.
func TestParseWritten(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		text string
		want duration.Written
	}{
		{"4", wr(duration.Quarter, 0)},
		{"4.", wr(duration.Quarter, 1)},
		{"4..", wr(duration.Quarter, 2)},
		{"16.", wr(duration.Sixteenth, 1)},
		{"1", wr(duration.Whole, 0)},
		{"2", wr(duration.Half, 0)},
		{"128", wr(duration.HundredTwentyEighth, 0)},
		{"1024.", wr(duration.ThousandTwentyFourth, 1)},
		{"breve", wr(duration.Breve, 0)},
		{"breve.", wr(duration.Breve, 1)},
		{"longa", wr(duration.Longa, 0)},
		{"maxima...", wr(duration.Maxima, 3)},
		{"8[3:2]", wr(duration.Eighth, 0, tuplet.Triplet())},
		{"4.[5:4,3:2]", wr(duration.Quarter, 1, tuplet.Quintuplet(), tuplet.Triplet())},
		{"4[6:4]", wr(duration.Quarter, 0, tp(6, 4))},
		{"4[1:1]", wr(duration.Quarter, 0, tp(1, 1))},
		{"1[3:2,3:2,3:2]", wr(duration.Whole, 0, tuplet.Triplet(), tuplet.Triplet(), tuplet.Triplet())},
		{"4[" + maxIntText + ":1]", wr(duration.Quarter, 0, bigTuplet)},
		{"maxima" + strings.Repeat(".", 254), wr(duration.Maxima, 254)},
	} {
		got, err := duration.ParseWritten(c.text)
		require.NoError(t, err, c.text)
		assert.Equal(t, c.want, got, c.text)
	}

	for _, text := range []string{
		"", "3", "6", "0", "01", "04", "2048", "10", "Breve", "BREVE", `\breve`, "quarter", "4[]", "4[3:2,]", "4[,3:2]", "4[3:0]",
		"4[0:2]", "4[03:2]", "4[3:02]", "4[3:2]]", "4[3:2", "4 [3:2]", "4[3: 2]", "4[3:2] ", " 4", ".4", "4.[3:2].", "4[3:2][5:4]",
		"4,", "4~2", "Written(unset)", "4[3]", "4[3:2:1]", "4[+3:2]", "4[-3:2]", "4[3/2]", "4.5", "max", "maxima[", "4[3:2,,5:4]",
		"4[3:2;5:4]", "4[3:2, 5:4]", "maxima~", "4\n", "４", "8[triplet]", "8[3]", "16..x",
	} {
		got, err := duration.ParseWritten(text)
		assertErr(t, err, duration.ErrInvalidText)
		assert.Equal(t, duration.Written{}, got, "%q", text)
	}

	for _, c := range []struct {
		text string
		want duration.Tied
	}{
		{"2~4", tie(wr(duration.Half, 0), wr(duration.Quarter, 0))},
		{"4[3:2]~8[3:2]", tie(wr(duration.Quarter, 0, tuplet.Triplet()), wr(duration.Eighth, 0, tuplet.Triplet()))},
		{"breve~breve~1", tie(wr(duration.Breve, 0), wr(duration.Breve, 0), wr(duration.Whole, 0))},
		{"4.~16", tie(wr(duration.Quarter, 1), wr(duration.Sixteenth, 0))},
		{"4~4", tie(wr(duration.Quarter, 0), wr(duration.Quarter, 0))},
	} {
		got, err := duration.ParseTied(c.text)
		require.NoError(t, err, c.text)
		assert.Equal(t, c.want, got, c.text)
	}

	for _, text := range []string{
		"", "4", "4.", "8[3:2]", "~4", "4~", "4~~2", "4 ~ 2", "4 ~2", "Tied(unset)", "2~4~", "2~3", "2~4[]", "2-4", "2_4", "2~4 ",
	} {
		got, err := duration.ParseTied(text)
		assertErr(t, err, duration.ErrInvalidText)
		assert.Equal(t, duration.Tied{}, got, "%q", text)
	}
}

// TestParseWrittenRange covers architecture 4.12: numbers of a tuplet above
// math.MaxInt, chain coefficients and values out of range, and sums of ties.
func TestParseWrittenRange(t *testing.T) {
	t.Parallel()

	big4 := strings.Repeat(maxIntText+":1,", 3) + maxIntText + ":1"
	big5 := big4 + "," + maxIntText + ":1"
	n2 := strconv.Itoa(math.MaxInt - 2)

	for _, c := range []struct {
		name   string
		text   string
		causes []error
	}{
		{"actual above max int", "4[" + maxIntPlusText + ":1]", []error{tuplet.ErrOutOfRange}},
		{"normal above max int", "4[1:" + maxIntPlusText + "]", []error{tuplet.ErrOutOfRange}},
		{"long tuplet number", "4[3:" + strings.Repeat("9", 100) + "]", []error{tuplet.ErrOutOfRange}},
		{"coefficient", "4[" + big5 + "]", []error{tuplet.ErrOutOfRange, fraction.ErrOutOfRange}},
		{"value under four tuplets", "1024[" + big4 + "]", []error{fraction.ErrOutOfRange}},
		{"KP-37 300 dots", "4" + strings.Repeat(".", 300), []error{fraction.ErrOutOfRange}},
		{"maxima with 255 dots", "maxima" + strings.Repeat(".", 255), []error{fraction.ErrOutOfRange}},
		{"many dots", "maxima" + strings.Repeat(".", 5000), []error{fraction.ErrOutOfRange}},
	} {
		got, err := duration.ParseWritten(c.text)
		assertErr(t, err, duration.ErrOutOfRange, c.causes...)
		assert.Equal(t, duration.Written{}, got, c.name)
	}

	for _, c := range []struct {
		name   string
		text   string
		causes []error
	}{
		{"part above max int", "4~4[" + maxIntPlusText + ":1]", []error{tuplet.ErrOutOfRange}},
		{"part value", "4~1024[" + big4 + "]", []error{fraction.ErrOutOfRange}},
		{"sum", "4[" + maxIntText + ":1," + maxIntText + ":1," + maxIntText + ":1]~4[" + n2 + ":1," + n2 + ":1," + n2 + ":1]",
			[]error{fraction.ErrOutOfRange}},
	} {
		got, err := duration.ParseTied(c.text)
		assertErr(t, err, duration.ErrOutOfRange, c.causes...)
		assert.Equal(t, duration.Tied{}, got, c.name)
	}

	// KR-D11: the sum of a tie is exact; partial sums do not matter.
	text := "1" + strings.Repeat(".", 199) + "~1[" + maxIntText + ":1," + maxIntText + ":1]~1024[" +
		"4611686018427387904:1,4611686018427387904:1,4611686018427387904:1,8:1]"
	got, err := duration.ParseTied(text)
	require.NoError(t, err)
	assert.Equal(t, 3, got.MustLen())
	assert.Equal(t, text, got.String())
}

// TestParseBase covers O-15, P-6, KP-36, KR-D13, and architecture 4.12:
// texts, LilyPond words, and names in any letter case; nothing else.
func TestParseBase(t *testing.T) {
	t.Parallel()

	for _, row := range baseTable {
		forms := []string{row.text, strings.ToUpper(row.text)}
		for _, name := range []string{row.us, row.uk, row.ru, row.lilyPad} {
			if name != "" {
				forms = append(forms, name, strings.ToUpper(name), capitalize(name))
			}
		}

		for _, text := range forms {
			got, err := duration.ParseBase(text)
			require.NoError(t, err, text)
			assert.Equal(t, row.base, got, text)
		}

		got, err := duration.ParseBase(row.base.String())
		require.NoError(t, err)
		assert.Equal(t, row.base, got, "ParseBase(String) is the base")
	}

	assert.Equal(t, duration.Quarter, duration.MustParseBase("Quarter Note"))
	assert.Equal(t, duration.Quarter, duration.MustParseBase("ЧЕТВЕРТНАЯ"))
	assert.Equal(t, duration.Whole, duration.MustParseBase("SemiBreve"))
	assert.Equal(t, duration.Breve, duration.MustParseBase(`\BREVE`))
	assert.Equal(t, duration.SixtyFourth, duration.MustParseBase("Шестьдесятчетвёртая"))

	for _, text := range []string{
		"", " 4", "4 ", " breve", "breve ", "quarter", "quarter  note", "quarter-note", "четвёртая", "шестьдесятчетвертая",
		`\4`, `\maxima `, "0", "3", "01", "2048", "4.", "8[3:2]", "Base(invalid)", "maxima.", "whole", "double whole",
		"hundred twenty-eighth note", "128th", "quaver.", "\tbreve",
	} {
		got, err := duration.ParseBase(text)
		assertErr(t, err, duration.ErrInvalidText)
		assert.Equal(t, duration.Base(0), got, "%q", text)
	}
}

// capitalize returns s with its first letter in upper case.
func capitalize(s string) string {
	r, size := utf8.DecodeRuneInString(s)

	return string(unicode.ToUpper(r)) + s[size:]
}

// TestRoundTrip covers P-6 and architecture 4.12: Parse…(x.String()) == x
// for set values, also for unreduced input.
func TestRoundTrip(t *testing.T) {
	t.Parallel()

	for _, v := range testValues() {
		d := durRat(v)
		assert.Equal(t, d, duration.MustParse(d.String()))

		p := posRat(v)
		assert.Equal(t, p, duration.MustParsePosition(p.String()))
	}

	assert.Equal(t, duration.Origin(), duration.MustParsePosition(duration.Origin().String()))

	for _, row := range baseTable {
		for dots := range 4 {
			for _, chain := range testChains() {
				w := wr(row.base, dots, chain...)
				assert.Equal(t, w, duration.MustParseWritten(w.String()))

				tt := tie(w, wr(duration.Quarter, 0), w)
				assert.Equal(t, tt, duration.MustParseTied(tt.String()))
			}
		}
	}
}

// TestUnmarshalText covers architecture 4.12: the grammar of the matching
// parse, an unchanged receiver on failure, E-2 for a nil receiver, and the
// round trip with MarshalText.
func TestUnmarshalText(t *testing.T) {
	t.Parallel()

	t.Run("Base", func(t *testing.T) {
		t.Parallel()

		b := duration.Half
		require.NoError(t, b.UnmarshalText([]byte("crotchet")))
		assert.Equal(t, duration.Quarter, b)

		assertErr(t, b.UnmarshalText([]byte("3")), duration.ErrInvalidText)
		assert.Equal(t, duration.Quarter, b, "unchanged on failure")

		var nilBase *duration.Base
		assertErr(t, nilBase.UnmarshalText([]byte("4")), duration.ErrInvalidArgument)
	})

	t.Run("Duration", func(t *testing.T) {
		t.Parallel()

		d := dur(1, 2)
		require.NoError(t, d.UnmarshalText([]byte("6/16")))
		assert.Equal(t, dur(3, 8), d)

		assertErr(t, d.UnmarshalText([]byte("0/1")), duration.ErrInvalidText)
		assert.Equal(t, dur(3, 8), d)

		assertErr(t, d.UnmarshalText([]byte(maxCompPlus+"/1")), duration.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, dur(3, 8), d)

		var nilDuration *duration.Duration
		assertErr(t, nilDuration.UnmarshalText([]byte("1/4")), duration.ErrInvalidArgument)

		text, err := d.MarshalText()
		require.NoError(t, err)

		var back duration.Duration
		require.NoError(t, back.UnmarshalText(text))
		assert.Equal(t, d, back)
	})

	t.Run("Position", func(t *testing.T) {
		t.Parallel()

		p := pos(1, 2)
		require.NoError(t, p.UnmarshalText([]byte("0/5")))
		assert.Equal(t, duration.Origin(), p)

		assertErr(t, p.UnmarshalText([]byte("Position(unset)")), duration.ErrInvalidText)
		assert.Equal(t, duration.Origin(), p)

		var nilPosition *duration.Position
		assertErr(t, nilPosition.UnmarshalText([]byte("0/1")), duration.ErrInvalidArgument)
	})

	t.Run("Written", func(t *testing.T) {
		t.Parallel()

		w := wr(duration.Half, 0)
		require.NoError(t, w.UnmarshalText([]byte("4.[3:2]")))
		assert.Equal(t, wr(duration.Quarter, 1, tuplet.Triplet()), w)

		assertErr(t, w.UnmarshalText([]byte("Written(unset)")), duration.ErrInvalidText)
		assertErr(t, w.UnmarshalText([]byte("4"+strings.Repeat(".", 300))), duration.ErrOutOfRange, fraction.ErrOutOfRange)
		assert.Equal(t, wr(duration.Quarter, 1, tuplet.Triplet()), w)

		var nilWritten *duration.Written
		assertErr(t, nilWritten.UnmarshalText([]byte("4")), duration.ErrInvalidArgument)
	})

	t.Run("Tied", func(t *testing.T) {
		t.Parallel()

		tt := tie(wr(duration.Half, 0), wr(duration.Half, 0))
		require.NoError(t, tt.UnmarshalText([]byte("2~4")))
		assert.Equal(t, tie(wr(duration.Half, 0), wr(duration.Quarter, 0)), tt)

		assertErr(t, tt.UnmarshalText([]byte("4")), duration.ErrInvalidText)
		assert.Equal(t, tie(wr(duration.Half, 0), wr(duration.Quarter, 0)), tt)

		var nilTied *duration.Tied
		assertErr(t, nilTied.UnmarshalText([]byte("2~4")), duration.ErrInvalidArgument)
	})

	// A nil receiver is checked before the text.
	var nilWritten *duration.Written
	assertErr(t, nilWritten.UnmarshalText([]byte("3")), duration.ErrInvalidArgument)
}

// TestAppendText covers architecture 3.2 and 4.12: the text is appended,
// the existing bytes are kept, and dst is returned unchanged on failure.
func TestAppendText(t *testing.T) {
	t.Parallel()

	type appender interface {
		AppendText([]byte) ([]byte, error)
	}

	for _, c := range []struct {
		name string
		v    appender
		text string
	}{
		{"Base", duration.Quarter, "4"},
		{"Duration", dur(3, 8), "3/8"},
		{"Position", duration.Origin(), "0/1"},
		{"Written", wr(duration.Eighth, 0, tuplet.Triplet()), "8[3:2]"},
		{"Tied", tie(wr(duration.Half, 0), wr(duration.Quarter, 0)), "2~4"},
	} {
		dst := make([]byte, 3, 64)
		copy(dst, "ab:")

		out, err := c.v.AppendText(dst)
		require.NoError(t, err, c.name)
		assert.Equal(t, "ab:"+c.text, string(out), c.name)
		assert.Equal(t, "ab:", string(dst), "%s: dst[:len(dst)] is unchanged", c.name)

		out, err = c.v.AppendText(nil)
		require.NoError(t, err)
		assert.Equal(t, c.text, string(out))
	}

	for _, c := range []struct {
		name string
		v    appender
		want error
	}{
		{"Base", duration.Base(0), duration.ErrInvalidArgument},
		{"Duration", duration.Duration{}, duration.ErrUnset},
		{"Position", duration.Position{}, duration.ErrUnset},
		{"Written", duration.Written{}, duration.ErrUnset},
		{"Tied", duration.Tied{}, duration.ErrUnset},
	} {
		dst := []byte("ab:")
		out, err := c.v.AppendText(dst)
		assertErr(t, err, c.want)
		assert.Equal(t, []byte("ab:"), out, c.name)
	}
}

// jsonHolder has required and optional fields of the five types.
type jsonHolder struct {
	Base     duration.Base     `json:"base"`
	Duration duration.Duration `json:"duration"`
	Position duration.Position `json:"position"`
	Written  duration.Written  `json:"written"`
	Tied     duration.Tied     `json:"tied"`
	Optional duration.Written  `json:"optional,omitzero"`
}

// TestJSON covers architecture 4.12: fields are strings of the canonical
// text; an unset field is omitted with omitzero and an encoding error
// otherwise.
func TestJSON(t *testing.T) {
	t.Parallel()

	codecs := []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"v1", jsonv1.Marshal, jsonv1.Unmarshal},
		{"v2", func(v any) ([]byte, error) { return jsonv2.Marshal(v) }, func(data []byte, v any) error { return jsonv2.Unmarshal(data, v) }},
	}

	full := jsonHolder{
		Base:     duration.Breve,
		Duration: dur(3, 8),
		Position: duration.Origin(),
		Written:  wr(duration.Quarter, 1, tuplet.Triplet()),
		Tied:     tie(wr(duration.Half, 0), wr(duration.Quarter, 0)),
		Optional: wr(duration.Eighth, 0),
	}

	for _, codec := range codecs {
		t.Run(codec.name, func(t *testing.T) {
			t.Parallel()

			data, err := codec.marshal(full)
			require.NoError(t, err)
			assert.JSONEq(t, `{"base":"breve","duration":"3/8","position":"0/1","written":"4.[3:2]","tied":"2~4","optional":"8"}`, string(data))

			var decoded jsonHolder
			require.NoError(t, codec.unmarshal(data, &decoded))
			assert.Equal(t, full, decoded)

			withoutOptional := full
			withoutOptional.Optional = duration.Written{}
			data, err = codec.marshal(withoutOptional)
			require.NoError(t, err)
			assert.JSONEq(t, `{"base":"breve","duration":"3/8","position":"0/1","written":"4.[3:2]","tied":"2~4"}`, string(data))

			unsetRequired := full
			unsetRequired.Duration = duration.Duration{}
			_, err = codec.marshal(unsetRequired)
			require.Error(t, err)
			assert.ErrorIs(t, err, duration.ErrUnset)

			err = codec.unmarshal([]byte(`{"written":"4[]"}`), &decoded)
			require.Error(t, err)
			assert.ErrorIs(t, err, duration.ErrInvalidText)

			err = codec.unmarshal([]byte(`{"base":"quarter note"}`), &decoded)
			require.NoError(t, err)
			assert.Equal(t, duration.Quarter, decoded.Base)
		})
	}
}
