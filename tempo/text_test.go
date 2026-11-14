package tempo_test

import (
	"encoding"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempo/term"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Interfaces of architecture 4.2.
var (
	_ fmt.Stringer             = tempo.Tempo{}
	_ encoding.TextMarshaler   = tempo.Tempo{}
	_ encoding.TextAppender    = tempo.Tempo{}
	_ encoding.TextUnmarshaler = (*tempo.Tempo)(nil)
	_ fmt.Stringer             = tempo.Approx{}
	_ encoding.TextMarshaler   = tempo.Approx{}
	_ encoding.TextAppender    = tempo.Approx{}
	_ encoding.TextUnmarshaler = (*tempo.Approx)(nil)
	_ fmt.Stringer             = tempo.Range{}
	_ encoding.TextMarshaler   = tempo.Range{}
	_ encoding.TextAppender    = tempo.Range{}
	_ encoding.TextUnmarshaler = (*tempo.Range)(nil)
	_ fmt.Stringer             = tempo.Seconds{}
	_ encoding.TextMarshaler   = tempo.Seconds{}
	_ encoding.TextAppender    = tempo.Seconds{}
	_ encoding.TextUnmarshaler = (*tempo.Seconds)(nil)
	_ fmt.Stringer             = tempo.Indication{}
	_ encoding.TextMarshaler   = tempo.Indication{}
	_ encoding.TextAppender    = tempo.Indication{}
	_ fmt.Stringer             = tempo.Kind(0)
)

// TestTextOutput covers С-27 and architecture 4.14, "Вывод": canonical
// text of every form.
func TestTextOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		x    fmt.Stringer
		want string
	}{
		{"♩ = 60", tq(60), "4=60"},
		{"♩. = 60", tm(dq, fi(60)), "4.=60"},
		{"♩ = 185/2", tm(q, fr(185, 2)), "4=185/2"},
		{"♩. = 200/3", tm(dq, fr(200, 3)), "4.=200/3"},
		{"triplet eighth", tm(e3, fi(360)), "8[3:2]=360"},
		{"chain", tm(duration.MustParseWritten("4.[5:4,3:2]"), fi(60)), "4.[5:4,3:2]=60"},
		{"breve", tm(breve, fr(1, 1000)), "breve=1/1000"},
		{"bpm 120/2", tm(q, fraction.MustNew(120, 2)), "4=60"},
		{"bpm max", tm(q, maxF()), "4=" + fraction.MaxComponentString},
		{"approx", ap(q, fi(108)), "4≈108"},
		{"approx dotted", ap(dq, fi(60)), "4.≈60"},
		{"approx fraction", ap(q, fr(185, 2)), "4≈185/2"},
		{"range", rg(e, fi(132), fi(144)), "8=132-144"},
		{"range fractions", rg(q, fr(1, 2), fr(3, 2)), "4=1/2-3/2"},
		{"seconds zero", tempo.ZeroSeconds(), "0/1"},
		{"seconds 1/3", sec(1, 3), "1/3"},
		{"seconds 2", sec(2, 1), "2/1"},
		{"seconds reduced", tempo.MustNewSeconds(4, 6), "2/3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.x.String())

			m, ok := tt.x.(encoding.TextMarshaler)
			require.True(t, ok)

			text, err := m.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(text))

			a, ok := tt.x.(encoding.TextAppender)
			require.True(t, ok)

			buf := make([]byte, 2, 64)
			copy(buf, "<<")
			got, err := a.AppendText(buf)
			require.NoError(t, err)
			assert.Equal(t, "<<"+tt.want, string(got))
			assert.Equal(t, "<<", string(buf), "AppendText keeps dst[:len(dst)]")
		})
	}
}

// TestParseTempo covers С-27, С-41 and architecture 4.14, "Разбор
// обозначений и времени" for Parse: accepted forms, numbers (integer, ratio,
// decimal), and rejected texts.
func TestParseTempo(t *testing.T) {
	t.Parallel()

	ok := []struct {
		text string
		want tempo.Tempo
	}{
		{"4=60", tq(60)},
		{"4.=60", tm(dq, fi(60))},
		{"4=185/2", tm(q, fr(185, 2))},
		{"С-27 4=92.5", tm(q, fr(185, 2))},
		{"4=120/2", tq(60)},
		{"4=60/1", tq(60)},
		{"4=60.0", tq(60)},
		{"4=60.000", tq(60)},
		{"4=0.5", tm(q, fr(1, 2))},
		{"4=1.50", tm(q, fr(3, 2))},
		{"4=0.001", tm(q, fr(1, 1000))},
		{"4=100000", tq(100000)},
		{"8[3:2]=360", tm(e3, fi(360))},
		{"breve=1/3", tm(breve, fr(1, 3))},
		{"1=60", tm(whole, fi(60))},
		{"4=" + fraction.MaxComponentString, tm(q, maxF())},
		{"4=1/" + fraction.MaxComponentString, tm(q, invMaxF())},
		{"4=60." + strings.Repeat("0", 100), tq(60)},
		{"4=0." + strings.Repeat("0", 75) + "1", tm(q, fbig(big.NewInt(1), pow10(76)))},
	}

	for _, tt := range ok {
		text := strings.TrimPrefix(tt.text, "С-27 ")

		got, err := tempo.Parse(text)
		if assert.NoError(t, err, "%q", text) {
			assert.Equal(t, tt.want, got, "%q", text)
		}
	}

	bad := []struct {
		text string
		want errClass
	}{
		{"", textClass},
		{"Tempo(unset)", textClass},
		{"4≈108", textClass},
		{"4=132-144", textClass},
		{"4=~108", textClass},
		{"4~108", textClass},
		{"4~8=60", textClass},
		{"4 = 60", textClass},
		{" 4=60", textClass},
		{"4=60 ", textClass},
		{"4=", textClass},
		{"=60", textClass},
		{"4==60", textClass},
		{"4=1/0", textClass},
		{"4=01", textClass},
		{"4=00", textClass},
		{"4=.5", textClass},
		{"4=5.", textClass},
		{"4=+5", textClass},
		{"4=-5", textClass},
		{"4=1/02", textClass},
		{"4=60e2", textClass},
		{"4=1/2/3", textClass},
		{"4=1.5/2", textClass},
		{"q=60", textClass},
		{"3=60", textClass},
		{"quarter=60", textClass},
		{"4=60x", textClass},
		{"♩=60", textClass},
		{"4=６０", textClass},
		{"\xff", textClass},
		{"4=0", noTempoClass},
		{"4=0/5", noTempoClass},
		{"4=0.0", noTempoClass},
		{"4=" + maxPlusOne(), fracRange},
		{"4=1" + strings.Repeat("0", 77), fracRange},
		{"4=1/" + maxPlusOne(), fracRange},
		{"4=0." + strings.Repeat("0", 76) + "1", fracRange},
		{"4[9223372036854775808:1]=60", unitTupletErr},
		{"8[" + strings.Repeat("9223372036854775807:1,", 4) + "9223372036854775807:1]=60", unitChainErr},
		{"4" + strings.Repeat(".", 300) + "=60", unitValueErr},
		{"4[9223372036854775808:1]=0", unitTupletErr},
		{"4[9223372036854775808:1]=6x", textClass},
		{"4=" + maxPlusOne() + "x", textClass},
	}

	for _, tt := range bad {
		got, err := tempo.Parse(tt.text)
		if !tt.want.check(t, err) {
			t.Logf("Parse(%.80q)", tt.text)
		}

		assert.Equal(t, tempo.Tempo{}, got)
	}
}

// TestParseApproxRangeSeconds covers С-27, С-41 and architecture 4.14 for
// ParseApprox, ParseRange, ParseSeconds and the order of checks of 4.3.
func TestParseApproxRangeSeconds(t *testing.T) {
	t.Parallel()

	t.Run("ParseApprox", func(t *testing.T) {
		t.Parallel()

		for text, want := range map[string]tempo.Approx{
			"4≈108": ap(q, fi(108)), "4.≈60": ap(dq, fi(60)), "4≈92.5": ap(q, fr(185, 2)), "4≈216/2": ap(q, fi(108)),
		} {
			got, err := tempo.ParseApprox(text)
			require.NoError(t, err, text)
			assert.Equal(t, want, got, text)
		}

		for text, want := range map[string]errClass{
			"4=108": textClass, "4=~108": textClass, "4~108": textClass, "4≈132-144": textClass, "Approx(unset)": textClass,
			"4 ≈ 108": textClass, "4≈": textClass, "≈108": textClass, "4≈≈108": textClass, "4≈0": noTempoClass,
			"4≈" + maxPlusOne(): fracRange, "4[9223372036854775808:1]≈0": unitTupletErr,
		} {
			got, err := tempo.ParseApprox(text)
			want.check(t, err)
			assert.Equal(t, tempo.Approx{}, got, text)
		}
	})

	t.Run("ParseRange", func(t *testing.T) {
		t.Parallel()

		for text, want := range map[string]tempo.Range{
			"8=132-144":   rg(e, fi(132), fi(144)),
			"4=120-131":   rg(q, fi(120), fi(131)),
			"4=120.5-131": rg(q, fr(241, 2), fi(131)),
			"4=1/2-3/2":   rg(q, fr(1, 2), fr(3, 2)),
			"4=2/4-1":     rg(q, fr(1, 2), fi(1)),
		} {
			got, err := tempo.ParseRange(text)
			require.NoError(t, err, text)
			assert.Equal(t, want, got, text)
		}

		for text, want := range map[string]errClass{
			"4=60": textClass, "4≈132-144": textClass, "4=132 - 144": textClass, "4=132-": textClass,
			"4=-132": textClass, "4=132-144-150": textClass, "Range(unset)": textClass, "4=132–144": textClass,
			"С-38 4=144-132": argClass, "4=120-120": argClass, "4=0-132": noTempoClass, "4=132-0": noTempoClass,
			"order 4=5-0": noTempoClass, "order 4=0-" + maxPlusOne(): fracRange, "order 4=" + maxPlusOne() + "-0": fracRange,
			"order 4=0-1x": textClass, "order 4[9223372036854775808:1]=5-0": unitTupletErr,
		} {
			text = strings.TrimPrefix(strings.TrimPrefix(text, "С-38 "), "order ")

			got, err := tempo.ParseRange(text)
			if !want.check(t, err) {
				t.Logf("ParseRange(%.80q)", text)
			}

			assert.Equal(t, tempo.Range{}, got, text)
		}
	})

	t.Run("ParseSeconds", func(t *testing.T) {
		t.Parallel()

		for text, want := range map[string]tempo.Seconds{
			"0/1": tempo.ZeroSeconds(), "1/3": sec(1, 3), "2/1": sec(2, 1), "4/2": sec(2, 1), "0/7": tempo.ZeroSeconds(),
			fraction.MaxComponentString + "/1": secF(maxF()),
		} {
			got, err := tempo.ParseSeconds(text)
			require.NoError(t, err, text)
			assert.Equal(t, want, got, text)
		}

		for text, want := range map[string]errClass{
			"2": textClass, "-1/2": textClass, "1/0": textClass, "0/0": textClass, "Seconds(unset)": textClass,
			"1.5/1": textClass, "1/2s": textClass, " 1/2": textClass, "01/2": textClass, "1/-2": textClass,
			maxPlusOne() + "/1": fracRange, "1/" + maxPlusOne(): fracRange, maxPlusOne() + "/" + maxPlusOne(): fracRange,
		} {
			got, err := tempo.ParseSeconds(text)
			want.check(t, err)
			assert.Equal(t, tempo.Seconds{}, got, text)
		}
	})
}

// TestTextRoundTrip covers С-27 and architecture 4.14, "Обратимость":
// Parse…(x.String()) == x and UnmarshalText(MarshalText(x)) == x.
func TestTextRoundTrip(t *testing.T) {
	t.Parallel()

	units := []duration.Written{q, dq, e, h, dh, e3, whole, breve, duration.MustParseWritten("4.[5:4,3:2]"),
		duration.MustParseWritten("1024..")}
	numbers := []fraction.Fraction{fi(1), fi(60), fr(185, 2), fr(200, 3), fr(1, 1000), fi(100000), maxF(), invMaxF()}

	for _, u := range units {
		for _, n := range numbers {
			x := tm(u, n)
			got, err := tempo.Parse(x.String())
			if assert.NoError(t, err, x.String()) {
				assert.Equal(t, x, got)
			}

			var y tempo.Tempo
			text, err := x.MarshalText()
			require.NoError(t, err)
			require.NoError(t, y.UnmarshalText(text))
			assert.Equal(t, x, y)

			a := ap(u, n)
			gotA, err := tempo.ParseApprox(a.String())
			if assert.NoError(t, err, a.String()) {
				assert.Equal(t, a, gotA)
			}

			var ya tempo.Approx
			require.NoError(t, ya.UnmarshalText([]byte(a.String())))
			assert.Equal(t, a, ya)
		}

		r := rg(u, fr(185, 2), fi(1000))
		gotR, err := tempo.ParseRange(r.String())
		if assert.NoError(t, err, r.String()) {
			assert.Equal(t, r, gotR)
		}

		var yr tempo.Range
		require.NoError(t, yr.UnmarshalText([]byte(r.String())))
		assert.Equal(t, r, yr)
	}

	for _, s := range []tempo.Seconds{tempo.ZeroSeconds(), sec(1, 3), sec(24, 37), secF(maxF()), secF(invMaxF())} {
		got, err := tempo.ParseSeconds(s.String())
		if assert.NoError(t, err, s.String()) {
			assert.Equal(t, s, got)
		}

		var y tempo.Seconds
		require.NoError(t, y.UnmarshalText([]byte(s.String())))
		assert.Equal(t, s, y)
	}
}

// TestUnmarshalText covers architecture 4.14: the grammar of the matching
// parse; on an error the receiver is unchanged; a nil receiver is Е-2.
func TestUnmarshalText(t *testing.T) {
	t.Parallel()

	x := tq(60)
	require.NoError(t, x.UnmarshalText([]byte("4.=60")))
	assert.Equal(t, tm(dq, fi(60)), x)

	x = tq(60)
	textClass.check(t, x.UnmarshalText([]byte("4≈60")))
	assert.Equal(t, tq(60), x)
	noTempoClass.check(t, x.UnmarshalText([]byte("4=0")))
	assert.Equal(t, tq(60), x)

	a := ap(q, fi(1))
	textClass.check(t, a.UnmarshalText([]byte("4=60")))
	assert.Equal(t, ap(q, fi(1)), a)

	r := rg(q, fi(1), fi(2))
	argClass.check(t, r.UnmarshalText([]byte("4=3-2")))
	assert.Equal(t, rg(q, fi(1), fi(2)), r)

	s := sec(1, 3)
	textClass.check(t, s.UnmarshalText([]byte("1/0")))
	assert.Equal(t, sec(1, 3), s)

	var unset tempo.Tempo
	require.NoError(t, unset.UnmarshalText([]byte("4=60")))
	assert.Equal(t, tq(60), unset)

	argClass.check(t, (*tempo.Tempo)(nil).UnmarshalText([]byte("4=60")))
	argClass.check(t, (*tempo.Tempo)(nil).UnmarshalText([]byte("bad")))
	argClass.check(t, (*tempo.Approx)(nil).UnmarshalText([]byte("4≈60")))
	argClass.check(t, (*tempo.Range)(nil).UnmarshalText([]byte("4=1-2")))
	argClass.check(t, (*tempo.Seconds)(nil).UnmarshalText([]byte("1/2")))

	text := []byte("4=60")
	var y tempo.Tempo
	require.NoError(t, y.UnmarshalText(text))
	copy(text, "8=90")
	assert.Equal(t, tq(60), y, "UnmarshalText does not retain text")
	assert.Equal(t, "8=90", string(text), "UnmarshalText does not change text")
}

// TestIndicationText covers С-28 and architecture 4.14: the output of an
// indication with quotes only where needed, and the round trip
// ParseIndication(i.String(), l) == i for words of built-in and user lists.
func TestIndicationText(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()
	mine := customList()
	word := func(name string) term.Term { return mine.MustParse(name) }

	tests := []struct {
		name string
		i    tempo.Indication
		want string
	}{
		{"word", tempo.MustIndicationOf(en.MustParse("Allegro")), "Allegro"},
		{"marking", tempo.MustIndicationOfTempo(tq(60)), "4=60"},
		{"word and marking", tempo.MustIndicationOf(en.MustParse("Allegro")).MustWithTempo(tq(120)), "Allegro, 4=120"},
		{"word with a space", tempo.MustIndicationOf(en.MustParse("Molto Allegro")), "Molto Allegro"},
		{"С-28 name 4=60", tempo.MustIndicationOf(word("4=60")), `"4=60"`},
		{"С-28 name 4=60 with ♩ = 120", tempo.MustIndicationOf(word("4=60")).MustWithTempo(tq(120)), `"4=60", 4=120`},
		{"С-28 name with a comma", tempo.MustIndicationOf(word("Скоро, но не слишком")), "Скоро, но не слишком"},
		{"С-28 name with a comma and ♩ = 120", tempo.MustIndicationOf(word("Скоро, но не слишком")).MustWithTempo(tq(120)),
			"Скоро, но не слишком, 4=120"},
		{"С-28 name starting with a quote", tempo.MustIndicationOf(word(`"Presto"`)), `"\"Presto\""`},
		{"name with ≈", tempo.MustIndicationOf(word("a≈b")), `"a≈b"`},
		{"name with = and control characters", tempo.MustIndicationOf(word("x=\t\\y")), strconv.Quote("x=\t\\y")},
		{"quote inside, no quotes needed", tempo.MustIndicationOf(word("q\"uote")), "q\"uote"},
		{"backslash, no quotes needed", tempo.MustIndicationOf(word("back\\slash")), "back\\slash"},
		{"spaces kept", tempo.MustIndicationOf(word(" spaced ")).MustWithApprox(ap(q, fi(108))), " spaced , 4≈108"},
		{"line break kept", tempo.MustIndicationOf(word("line\nbreak")), "line\nbreak"},
		{"name ≈", tempo.MustIndicationOf(word("≈")), `"≈"`},
		{"name =", tempo.MustIndicationOf(word("=")).MustWithRange(rg(e, fi(132), fi(144))), `"=", 8=132-144`},
		{"name like a marking", tempo.MustIndicationOf(word("x, 4=120")), `"x, 4=120"`},
		{"name ending with a comma", tempo.MustIndicationOf(word("x, ")).MustWithTempo(tq(60)), "x, , 4=60"},
		{"name like a special string", tempo.MustIndicationOf(word("Tempo(unset)")), "Tempo(unset)"},
		{"zero byte", tempo.MustIndicationOf(word("\x00")), "\x00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.i.String())

			text, err := tt.i.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(text))

			got, err := tt.i.AppendText([]byte("["))
			require.NoError(t, err)
			assert.Equal(t, "["+tt.want, string(got))

			l := mine
			if w, ok := tt.i.MustTerm(); ok {
				l = w.MustList()
			}

			back, err := tempo.ParseIndication(tt.i.String(), l)
			if assert.NoError(t, err) {
				assert.Equal(t, tt.i, back)
			}
		})
	}

	t.Run("С-28 a word 4=60 is not the marking ♩ = 60", func(t *testing.T) {
		t.Parallel()

		w := tempo.MustIndicationOf(word("4=60"))
		m := tempo.MustIndicationOfTempo(tq(60))
		assert.NotEqual(t, w.String(), m.String())
		assert.Equal(t, m, tempo.MustParseIndication("4=60", mine), "an unquoted 4=60 is the marking even in a list with that word")
		assert.Equal(t, w, tempo.MustParseIndication(`"4=60"`, mine))
	})

	t.Run("every word of every list round-trips", func(t *testing.T) {
		t.Parallel()

		for _, l := range append(term.BuiltIn(), mine) {
			for _, w := range l.MustTerms() {
				for _, i := range []tempo.Indication{
					tempo.MustIndicationOf(w),
					tempo.MustIndicationOf(w).MustWithTempo(tm(dq, fr(185, 2))),
					tempo.MustIndicationOf(w).MustWithApprox(ap(q, fi(108))),
					tempo.MustIndicationOf(w).MustWithRange(rg(e, fi(132), fi(144))),
				} {
					got, err := tempo.ParseIndication(i.String(), l)
					if assert.NoError(t, err, "%q", i.String()) {
						assert.Equal(t, i, got, "%q", i.String())
					}
				}
			}
		}
	})
}

// TestParseIndication covers С-27, С-28, С-41 and architecture 4.14, "Разбор
// указания", steps 1-6 and their errors.
func TestParseIndication(t *testing.T) {
	t.Parallel()

	en := term.EnglishWikipedia()
	allegro := tempo.MustIndicationOf(en.MustParse("Allegro"))

	ok := []struct {
		text string
		want tempo.Indication
	}{
		{"Allegro, 4=120", allegro.MustWithTempo(tq(120))},
		{"allegro, 4=120", allegro.MustWithTempo(tq(120))},
		{`"Allegro", 4=120`, allegro.MustWithTempo(tq(120))},
		{`"allegro"`, allegro},
		{"ALLEGRO", allegro},
		{"Vivo, 4≈170", tempo.MustIndicationOf(en.MustParse("Vivace")).MustWithApprox(ap(q, fi(170)))},
		{"Moderato, 4≈108", tempo.MustIndicationOf(en.MustParse("Moderato")).MustWithApprox(ap(q, fi(108)))},
		{"Allegro, 8=132-144", allegro.MustWithRange(rg(e, fi(132), fi(144)))},
		{"Allegro, 4=92.5", allegro.MustWithTempo(tm(q, fr(185, 2)))},
		{"4=60", tempo.MustIndicationOfTempo(tq(60))},
		{"4≈108", tempo.MustIndicationOfApprox(ap(q, fi(108)))},
		{"8=132-144", tempo.MustIndicationOfRange(rg(e, fi(132), fi(144)))},
		{"Molto Allegro, 4=130", tempo.MustIndicationOf(en.MustParse("Molto Allegro")).MustWithTempo(tq(130))},
	}

	for _, tt := range ok {
		got, err := tempo.ParseIndication(tt.text, en)
		if assert.NoError(t, err, "%q", tt.text) {
			assert.Equal(t, tt.want, got, "%q", tt.text)
		}
	}

	bad := []struct {
		name string
		text string
		l    term.List
		want errClass
	}{
		{"step 1 unset list", "Allegro", term.List{}, argClass},
		{"step 1 before the text", "\xff", term.List{}, argClass},
		{"step 2 not UTF-8", "Allegro\xff, 4=120", en, textClass},
		{"step 3 no literal", `"Allegro, 4=120`, en, textClass},
		{"step 3 junk after the literal", `"Allegro"x`, en, textClass},
		{"step 3 no space", `"Allegro",4=120`, en, textClass},
		{"step 3 empty marking", `"Allegro", `, en, textClass},
		{"step 3 not UTF-8 after Unquote", `"\xff"`, en, textClass},
		{"step 3 empty quoted name", `""`, en, textClass},
		{"step 4 = in the name", "Allegro=x, 4=120", en, textClass},
		{"step 4 two markings", "Allegro, 4=120, 4=60", en, textClass},
		{"step 4 no space", "Allegro,4=120", en, textClass},
		{"step 5 unknown word", "Fast", en, textClass},
		{"С-41 unknown word with a marking", "Fast, 4=120", en, textClass},
		{"С-41 word of another list", "Larghissimo", term.RussianWikipedia(), textClass},
		{"step 5 empty", "", en, textClass},
		{"step 5 trailing comma", "Allegro, ", en, textClass},
		{"step 6 special string", "Tempo(unset)", en, textClass},
		{"step 6 old form", "Allegro, 4=~108", en, textClass},
		{"step 6 bad marking", "Allegro, 4=60x", en, textClass},
		{"step 6 bpm 0", "Allegro, 4=0", en, noTempoClass},
		{"step 6 range order", "Allegro, 4=132-120", en, argClass},
		{"step 6 number out of range", "Allegro, 4=" + maxPlusOne(), en, fracRange},
		{"step 6 unit out of range", "Allegro, 4[9223372036854775808:1]=60", en, unitTupletErr},
		{"order: name before marking", "Fast, 4=0", en, textClass},
	}

	for _, tt := range bad {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tempo.ParseIndication(tt.text, tt.l)
			tt.want.check(t, err)
			assert.Equal(t, tempo.Indication{}, got)
		})
	}
}

// TestLongNumbers covers architecture 4.14 and 7: the result depends on the
// exact value, not the length; numbers of a megabyte are parsed in linear
// time (Е-11 or a value), and 0 followed by a megabyte of zeros is Е-3.
func TestLongNumbers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want errClass // zero means success
		bpm  fraction.Fraction
	}{
		{"integer of 1 MiB", "4=" + rep("1", mib), fracRange, fraction.Fraction{}},
		{"denominator of 1 MiB", "4=1/" + rep("1", mib), fracRange, fraction.Fraction{}},
		{"numerator of 1 MiB over a long denominator", "4=" + rep("9", mib) + "/" + rep("9", mib), fracRange, fraction.Fraction{}},
		{"decimal digits of 1 MiB", "4=1." + rep("1", mib), fracRange, fraction.Fraction{}},
		{"trailing zeros of 1 MiB", "4=1." + rep("0", mib), errClass{}, fi(1)},
		{"zero with zeros of 1 MiB", "4=0." + rep("0", mib), noTempoClass, fraction.Fraction{}},
		{"half with trailing zeros of 1 MiB", "4=0.5" + rep("0", mib), errClass{}, fr(1, 2)},
		{"256 significant decimal digits", "4=0." + rep("0", 255) + "1", fracRange, fraction.Fraction{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			type out struct {
				v   tempo.Tempo
				err error
			}

			got := fast(t, func() out { v, err := tempo.Parse(tt.text); return out{v, err} })

			if tt.want.want == nil {
				require.NoError(t, got.err)
				assert.Equal(t, tm(q, tt.bpm), got.v)

				return
			}

			tt.want.check(t, got.err)
			assert.LessOrEqual(t, len(got.err.Error()), errorTextLimit)
		})
	}

	t.Run("range and seconds", func(t *testing.T) {
		t.Parallel()

		err := fast(t, func() error { _, err := tempo.ParseRange("4=1-" + rep("1", mib)); return err })
		fracRange.check(t, err)

		err = fast(t, func() error { _, err := tempo.ParseSeconds(rep("1", mib) + "/1"); return err })
		fracRange.check(t, err)

		err = fast(t, func() error { _, err := tempo.ParseApprox("4≈" + rep("1", mib)); return err })
		fracRange.check(t, err)

		err = fast(t, func() error {
			_, err := tempo.ParseIndication("Allegro, 4="+rep("1", mib), term.EnglishWikipedia())
			return err
		})
		fracRange.check(t, err)
	})
}
