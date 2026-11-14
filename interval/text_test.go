package interval_test

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-7, short notation: String, MarshalText, AppendText, Parse,
// UnmarshalText, the error text of parsing (OP-40, OP-41, OP-44; P-21;
// KP-16; F-5; architecture 4.12, KR-5, KR-14).

// canonical is the grammar of the canonical form of 4.12.
var canonical = regexp.MustCompile(`^-?(P|M|m|n|A{1,3}|d{1,3}|\{[+-][1-9][0-9]*(/[1-9][0-9]*)?\})[1-9][0-9]*$`)

func TestTextString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		i    interval.Interval
		want string
	}{
		{"P1", iv("P1"), "P1"},
		{"P1 made descending", iv("↓P1"), "P1"},
		{"m2", iv("m2"), "m2"},
		{"M3", iv("M3"), "M3"},
		{"A4", iv("A4"), "A4"},
		{"d5", iv("d5"), "d5"},
		{"AA4", iv("AA4"), "AA4"},
		{"dd5", iv("dd5"), "dd5"},
		{"AAA3", iv("AAA3"), "AAA3"},
		{"ddd6", iv("ddd6"), "ddd6"},
		{"n3", iv("n3"), "n3"},
		{"P15", iv("P15"), "P15"},
		{"d8", iv("d8"), "d8"},
		{"dd8", iv("dd8"), "dd8"},
		{"descending m6", iv("↓m6"), "-m6"},
		{"descending A1", iv("↓A1"), "-A1"},
		{"descending n3", iv("↓n3"), "-n3"},
		{"descending P8", iv("↓P8"), "-P8"},
		{"fifth -1/2", iv("{-1/2}5"), "{-1/2}5"},
		{"descending second +1/3", iv("↓{+1/3}2"), "-{+1/3}2"},
		{"third +1/4", iv("{+1/4}3"), "{+1/4}3"},
		{"third -3/2", iv("{-3/2}3"), "{-3/2}3"},
		{"third +1/2", iv("{+1/2}3"), "{+1/2}3"},
		{"unison +1/2", iv("{+1/2}1"), "{+1/2}1"},
		{"descending unison +1/2", iv("↓{+1/2}1"), "-{+1/2}1"},
		{"augmented fourth of multiplicity 4", aug(4, 4), "{+4}4"},
		{"diminished third of multiplicity 4", dim(4, 3), "{-5}3"},
		{"diminished fifth of multiplicity 4", dim(4, 5), "{-4}5"},
		{"augmented unison of multiplicity 5", aug(5, 1), "{+5}1"},
		{"deviation M", aug(mag, 4), "{+16777216}4"},
		{"deviation -M of a fifth", dim(mag, 5), "{-16777216}5"},
		{"deviation -M of a third", interval.MustFromDeviation(3, ai(-mag)), "{-16777216}3"},
		{"major of MaxSize", basicSize(interval.MaxSize), "M9223372036854775807"},
		{"descending minor of MaxSize", interval.MustFromDeviation(interval.MaxSize, ai(-1)).MustDown(), "-m9223372036854775807"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.i.String())

			b, err := tt.i.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(b))

			prefix := []byte("x=")
			out, err := tt.i.AppendText(prefix)
			require.NoError(t, err)
			assert.Equal(t, "x="+tt.want, string(out))
			assert.Equal(t, "x=", string(prefix), "AppendText does not change b[:len(b)]")

			assert.Regexp(t, canonical, tt.want)
		})
	}
}

func TestTextParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text string
		want interval.Interval
	}{
		{"P1", iv("P1")},
		{"-P1", iv("P1")},
		{"-perf1", iv("P1")},
		{"M3", iv("M3")},
		{"m3", iv("m3")},
		{"n3", iv("n3")},
		{"A4", iv("A4")},
		{"d5", iv("d5")},
		{"AA4", iv("AA4")},
		{"dd5", iv("dd5")},
		{"AAA3", iv("AAA3")},
		{"ddd5", iv("ddd5")},
		{"AAAA4", aug(4, 4)},
		{"dddd5", dim(4, 5)},
		{"dddd3", dim(4, 3)},
		{"P15", iv("P15")},
		{"ma3", iv("M3")},
		{"MA3", iv("M3")},
		{"maj3", iv("M3")},
		{"MAJ3", iv("M3")},
		{"Maj3", iv("M3")},
		{"mi3", iv("m3")},
		{"MI3", iv("m3")},
		{"min3", iv("m3")},
		{"Min3", iv("m3")},
		{"+4", iv("A4")},
		{"aug4", iv("A4")},
		{"AUG4", iv("A4")},
		{"°5", iv("d5")},
		{"o5", iv("d5")},
		{"dim5", iv("d5")},
		{"Dim5", iv("d5")},
		{"perf5", iv("P5")},
		{"PERF5", iv("P5")},
		{"-m6", iv("↓m6")},
		{"-A1", iv("↓A1")},
		{"-aug1", iv("↓A1")},
		{"-n3", iv("↓n3")},
		{"{+1}4", iv("A4")},
		{"{-1}3", iv("m3")},
		{"{-1/2}3", iv("n3")},
		{"{-1/2}5", iv("{-1/2}5")},
		{"-{+1/3}2", iv("↓{+1/3}2")},
		{"{+1/4}3", iv("{+1/4}3")},
		{"{-5}3", dim(4, 3)},
		{"{+4}4", aug(4, 4)},
		{"{+1/2}1", iv("{+1/2}1")},
		{"-{+1/2}1", iv("↓{+1/2}1")},
		{"{+16777216}4", aug(mag, 4)},
		{"{-16777216}3", interval.MustFromDeviation(3, ai(-mag))},
		{"M9223372036854775807", basicSize(interval.MaxSize)},
		{"-m9223372036854775807", interval.MustFromDeviation(interval.MaxSize, ai(-1)).MustDown()},
		{strings.Repeat("A", 1000) + "4", aug(1000, 4)},
		{strings.Repeat("d", 1000) + "6", dim(1000, 6)},
	}

	for _, tt := range tests {
		t.Run(tt.text[:min(len(tt.text), 24)], func(t *testing.T) {
			t.Parallel()

			got, err := interval.Parse(tt.text)
			require.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "Parse(%q) = %v, want %v", tt.text, got, tt.want)

			var u interval.Interval
			require.NoError(t, u.UnmarshalText([]byte(tt.text)))
			assert.True(t, u.Equal(tt.want))
		})
	}
}

func TestTextParseErrors(t *testing.T) {
	t.Parallel()

	// 10^30 ≡ 1 (mod 7): a size whose simple part is 8, above MaxSize.
	huge := "1" + strings.Repeat("0", 30)

	tests := []struct {
		text   string
		want   error
		causes []error
	}{
		{"", interval.ErrParse, nil},
		{"TT", interval.ErrParse, nil},
		{"TT4", interval.ErrParse, nil},
		{"++4", interval.ErrParse, nil},
		{"oo5", interval.ErrParse, nil},
		{"°°5", interval.ErrParse, nil},
		{"M 3", interval.ErrParse, nil},
		{"a4", interval.ErrParse, nil},
		{"D5", interval.ErrParse, nil},
		{"p5", interval.ErrParse, nil},
		{"N3", interval.ErrParse, nil},
		{"O5", interval.ErrParse, nil},
		{"M", interval.ErrParse, nil},
		{"3", interval.ErrParse, nil},
		{"-", interval.ErrParse, nil},
		{"--M3", interval.ErrParse, nil},
		{"+M3", interval.ErrParse, nil},
		{"↓M3", interval.ErrParse, nil},
		{"−m6", interval.ErrParse, nil},
		{"M03", interval.ErrParse, nil},
		{"M0", interval.ErrParse, nil},
		{"P01", interval.ErrParse, nil},
		{"M3 ", interval.ErrParse, nil},
		{" M3", interval.ErrParse, nil},
		{"M3x", interval.ErrParse, nil},
		{"M3-", interval.ErrParse, nil},
		{"Ad4", interval.ErrParse, nil},
		{"dA4", interval.ErrParse, nil},
		{"Mm3", interval.ErrParse, nil},
		{"PP5", interval.ErrParse, nil},
		{"nn3", interval.ErrParse, nil},
		{"majmaj3", interval.ErrParse, nil},
		{"major3", interval.ErrParse, nil},
		{"tritone", interval.ErrParse, nil},
		{"aug", interval.ErrParse, nil},
		{"{+2/4}3", interval.ErrParse, nil},
		{"{+1/1}3", interval.ErrParse, nil},
		{"{+0}3", interval.ErrParse, nil},
		{"{-0}3", interval.ErrParse, nil},
		{"{1/2}3", interval.ErrParse, nil},
		{"{+1/0}3", interval.ErrParse, nil},
		{"{+01/2}3", interval.ErrParse, nil},
		{"{+1/02}3", interval.ErrParse, nil},
		{"{+1/2}", interval.ErrParse, nil},
		{"{+1/2 }3", interval.ErrParse, nil},
		{"{+1/2}03", interval.ErrParse, nil},
		{"A{+1}4", interval.ErrParse, nil},
		{"{+2/4}1", interval.ErrParse, nil},
		{"M5 ", interval.ErrParse, nil},
		{"M5", interval.ErrInvalidArgument, nil},
		{"P3", interval.ErrInvalidArgument, nil},
		{"d1", interval.ErrInvalidArgument, nil},
		{"dim1", interval.ErrInvalidArgument, nil},
		{"-d1", interval.ErrInvalidArgument, nil},
		{"n4", interval.ErrInvalidArgument, nil},
		{"n1", interval.ErrInvalidArgument, nil},
		{"n15", interval.ErrInvalidArgument, nil},
		{"m1", interval.ErrInvalidArgument, nil},
		{"P2", interval.ErrInvalidArgument, nil},
		{"M8", interval.ErrInvalidArgument, nil},
		{"maj5", interval.ErrInvalidArgument, nil},
		{"perf3", interval.ErrInvalidArgument, nil},
		{"{-1}1", interval.ErrInvalidArgument, nil},
		{"{-1/2}1", interval.ErrInvalidArgument, nil},
		{"P9223372036854775807", interval.ErrInvalidArgument, nil},
		{"M9223372036854775808", interval.ErrInvalidArgument, nil},
		{"M" + huge, interval.ErrInvalidArgument, nil},
		{"P9223372036854775808", interval.ErrOutOfRange, nil},
		{"P" + huge, interval.ErrOutOfRange, nil},
		{"{+16777217}4", interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"{-16777217}5", interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"{+1/65537}3", interval.ErrOutOfRange, []error{alteration.ErrOutOfRange}},
		{"{+1/999999999999999999999}3", interval.ErrOutOfRange, nil},
	}

	for _, tt := range tests {
		t.Run(strconv.Quote(tt.text), func(t *testing.T) {
			t.Parallel()

			got, err := interval.Parse(tt.text)
			assertErr(t, err, tt.want, tt.causes...)
			assert.Equal(t, interval.Interval{}, got)

			target := iv("M3")
			err = target.UnmarshalText([]byte(tt.text))
			assertErr(t, err, tt.want, tt.causes...)
			assert.True(t, target.Equal(iv("M3")), "the receiver does not change on error")
		})
	}

	t.Run("A repeated M+1 times", func(t *testing.T) {
		t.Parallel()

		got, err := interval.Parse(strings.Repeat("A", mag+1) + "4")
		assertErr(t, err, interval.ErrOutOfRange, alteration.ErrOutOfRange)
		assert.Equal(t, interval.Interval{}, got)
	})
}

// roundTripSet is the sweep of F-5: the predefined intervals in both
// directions, deviations ±1/2 and ±1/4 at sizes 1...15, multiplicities 1...5,
// MaxSize and the deviations ±M.
func roundTripSet() []interval.Interval {
	set := predefinedBoth()

	for n := 1; n <= 15; n++ {
		for _, d := range []alteration.Alteration{alt(1, 2), alt(-1, 2), alt(1, 4), alt(-1, 4)} {
			if n == 1 && d.Sign() < 0 {
				continue
			}

			set = append(set, both([]interval.Interval{interval.MustFromDeviation(n, d)})...)
		}

		for k := 1; k <= 5; k++ {
			set = append(set, both([]interval.Interval{aug(k, n)})...)
			if n > 1 {
				set = append(set, both([]interval.Interval{dim(k, n)})...)
			}
		}
	}

	return append(set, both([]interval.Interval{
		basicSize(interval.MaxSize), interval.MustFromDeviation(interval.MaxSize, ai(-1)),
		aug(mag, 4), dim(mag, 5), aug(mag, 3), interval.MustFromDeviation(3, ai(-mag)),
		aug(mag, 1), interval.MustFromDeviation(2, alt(-2, 7)),
	})...)
}

func TestTextRoundTrip(t *testing.T) {
	t.Parallel()

	for _, i := range roundTripSet() {
		s := i.String()
		assert.Regexp(t, canonical, s)
		assert.Equal(t, i.IsDescending(), strings.HasPrefix(s, "-"), "%s: the sign of a descending interval", s)

		// "{…}" is used for fractional qualities other than neutral and for
		// multiplicities of 4 and more; letters otherwise.
		braced := i.Quality() == interval.Fractional || i.Multiplicity() >= 4
		assert.Equal(t, braced, strings.Contains(s, "{"), "%s", s)

		got, err := interval.Parse(s)
		require.NoError(t, err, "%s", s)
		assert.True(t, got.Equal(i), "Parse(%q) = %v", s, got)

		b, err := i.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, s, string(b))

		var u interval.Interval
		require.NoError(t, u.UnmarshalText(b))
		assert.True(t, u.Equal(i), "%s", s)
	}
}

func TestTextUnset(t *testing.T) {
	t.Parallel()

	var u interval.Interval

	assert.Equal(t, "Interval(unset)", u.String())

	b, err := u.MarshalText()
	assertErr(t, err, interval.ErrUnset)
	assert.Nil(t, b)

	prefix := []byte("abc")
	out, err := u.AppendText(prefix)
	assertErr(t, err, interval.ErrUnset)
	assert.Equal(t, []byte("abc"), out, "AppendText returns b on error")

	var nilTarget *interval.Interval
	assertErr(t, nilTarget.UnmarshalText([]byte("M3")), interval.ErrUnset)

	_, err = interval.Parse("Interval(unset)")
	assertErr(t, err, interval.ErrParse)
}

func TestTextAppendText(t *testing.T) {
	t.Parallel()

	buf := make([]byte, 3, 64)
	copy(buf, "abc")

	out, err := iv("↓AA4").AppendText(buf)
	require.NoError(t, err)
	assert.Equal(t, "abc-AA4", string(out))
	assert.Equal(t, "abc", string(buf))

	out, err = iv("M3").AppendText(nil)
	require.NoError(t, err)
	assert.Equal(t, "M3", string(out))
}

func TestTextJSON(t *testing.T) {
	t.Parallel()

	type record struct {
		I interval.Interval `json:"i"`
		J interval.Interval `json:"j,omitzero"`
	}

	b, err := json.Marshal(record{I: iv("↓m6")})
	require.NoError(t, err)
	assert.JSONEq(t, `{"i":"-m6"}`, string(b))

	b, err = json.Marshal(record{I: iv("n3"), J: iv("{+1/4}3")})
	require.NoError(t, err)
	assert.JSONEq(t, `{"i":"n3","j":"{+1/4}3"}`, string(b))

	var r record
	require.NoError(t, json.Unmarshal([]byte(`{"i":"aug4","j":"-P1"}`), &r))
	assert.True(t, r.I.Equal(iv("A4")))
	assert.True(t, r.J.Equal(iv("P1")))

	_, err = json.Marshal(record{})
	assert.True(t, errors.Is(err, interval.ErrUnset), "an unset interval without omitzero: %v", err)

	err = json.Unmarshal([]byte(`{"i":"TT"}`), &r)
	assert.True(t, errors.Is(err, interval.ErrParse), "%v", err)
}

// ---------------------------------------------------------------------------
// The error text of parsing (architecture 4.12, rule of duration 4.12).

// Sizes and limits of the error text.
const (
	kib            = 1024
	mib            = 1024 * 1024
	excerptLimit   = 64
	dependentBytes = 67
	errorTextLimit = 512
)

// excerptLen returns the length of the longest prefix of whole UTF-8 units
// of x that is at most 64 bytes; a unit is a valid rune or a single byte
// that does not start one.
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

// parsePath is one path that quotes the rejected input.
type parsePath struct {
	name  string
	names bool   // the grammar of names (ParseName) rather than the short notation
	base  string // the non-Must path whose error a Must path panics with
	run   func(t *testing.T, x string) error
}

func parsePaths() []parsePath {
	return []parsePath{
		{name: "Parse", run: func(t *testing.T, x string) error {
			got, err := interval.Parse(x)
			assert.Equal(t, interval.Interval{}, got)

			return err
		}},
		{name: "MustParse", base: "Parse", run: func(t *testing.T, x string) error {
			return panicError(t, func() { interval.MustParse(x) })
		}},
		{name: "UnmarshalText", run: func(t *testing.T, x string) error {
			target := iv("↓m6")
			err := target.UnmarshalText([]byte(x))
			assert.True(t, target.Equal(iv("↓m6")))

			return err
		}},
		{name: "ParseName", names: true, run: func(t *testing.T, x string) error {
			got, err := interval.ParseName(x)
			assert.Equal(t, interval.Interval{}, got)

			return err
		}},
		{name: "MustParseName", names: true, base: "ParseName", run: func(t *testing.T, x string) error {
			return panicError(t, func() { interval.MustParseName(x) })
		}},
	}
}

// checkErrorText checks the guarantees of 4.12 for the error text of x.
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
		assert.True(t, strings.Contains(text, strconv.Quote(x[:k])+"..."), "%s: the excerpt is followed by ...", label)
	}
}

func TestTextParseErrorExcerpt(t *testing.T) {
	t.Parallel()

	rep := strings.Repeat

	inputs := []struct {
		name         string
		make         func(n int) string
		short, name2 error // classes for the short notation and for names
	}{
		{"letters", func(n int) string { return rep("x", n) }, interval.ErrParse, interval.ErrParse},
		{"a name and spaces", func(n int) string { return "major third" + rep(" ", n) }, interval.ErrParse, interval.ErrParse},
		{"a notation and digits after a letter", func(n int) string { return "M3x" + rep("3", n) }, interval.ErrParse, interval.ErrParse},
		{"a long denominator", func(n int) string { return "{+1/" + rep("9", n) + "}3" }, interval.ErrOutOfRange, interval.ErrParse},
		{"a rune at the boundary", func(n int) string { return rep("a", 63) + "ё" + rep("b", n) }, interval.ErrParse, interval.ErrParse},
		{"two-byte runes", func(n int) string { return rep("ж", n) }, interval.ErrParse, interval.ErrParse},
		{"four-byte runes", func(n int) string { return rep("😀", n) }, interval.ErrParse, interval.ErrParse},
		{"invalid bytes", func(n int) string { return rep("\xff", n) }, interval.ErrParse, interval.ErrParse},
		{"a truncated rune at the boundary", func(n int) string { return rep("a", 63) + "\xd0" + rep("b", n) }, interval.ErrParse, interval.ErrParse},
	}

	for _, in := range inputs {
		t.Run(in.name, func(t *testing.T) {
			t.Parallel()

			small, large := in.make(kib), in.make(mib)
			require.Equal(t, small[:dependentBytes], large[:dependentBytes], "the first 67 bytes are the same")

			texts := map[string]string{}

			for _, p := range parsePaths() {
				want := in.short
				if p.names {
					want = in.name2
				}

				var bySize []string

				for _, x := range []string{small, large} {
					err := p.run(t, x)
					if !assertErr(t, err, want) {
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

func TestTextParseErrorShortInput(t *testing.T) {
	t.Parallel()

	for _, x := range []string{
		"", "TT", "tritone", "Interval(unset)", "ё", "\xff", "\"quoted\"", "tab\there",
		strings.Repeat("a", 64), strings.Repeat("ж", 32),
	} {
		for _, p := range parsePaths() {
			err := p.run(t, x)
			if !assertErr(t, err, interval.ErrParse) {
				continue
			}

			assert.True(t, strings.Contains(err.Error(), strconv.Quote(x)), "%s: %q is quoted in full", p.name, x)
		}
	}
}

// TestTextParseErrorWraps checks that parse errors wrap only sentinels: the
// cause alteration.ErrOutOfRange at E-12 of the deviation, nothing else of
// the dependencies.
func TestTextParseErrorWraps(t *testing.T) {
	t.Parallel()

	_, err := interval.Parse("{+1/65537}3")
	assertErr(t, err, interval.ErrOutOfRange, alteration.ErrOutOfRange)

	_, err = interval.Parse("M5")
	assertErr(t, err, interval.ErrInvalidArgument)
	assert.NotErrorIs(t, err, alteration.ErrOutOfRange)
}
