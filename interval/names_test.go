package interval_test

import (
	"strconv"
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-7, names: Name, RussianName, RussianAbbreviation, ParseName (OP-42,
// OP-43; P-22; table 2.7; KP-16; F-6; architecture 4.13).

// Words of table 2.7 of business-spec.md (domain base, topic interval,
// version 2, section 6).
var (
	sizeEnglish = [...]string{1: "unison", "second", "third", "fourth", "fifth", "sixth", "seventh", "octave",
		"ninth", "tenth", "eleventh", "twelfth", "thirteenth", "fourteenth", "fifteenth"}
	sizeRussian = [...]string{1: "прима", "секунда", "терция", "кварта", "квинта", "секста", "септима", "октава",
		"нона", "децима", "ундецима", "дуодецима", "терцдецима", "квартдецима", "квинтдецима"}
)

// names is the expected result of the three name methods for one interval.
type names struct {
	en, ru, abbr       string
	enOK, ruOK, abbrOK bool
}

// expectedNames builds the names of P-22 from table 2.7.
func expectedNames(i interval.Interval) names {
	n := i.Size()
	if n > 15 {
		return names{}
	}

	k := i.Multiplicity()

	var en, ru, ab string

	switch i.Quality() {
	case interval.Perfect:
		en, ru, ab = "perfect", "чистая", "ч."
	case interval.Major:
		en, ru, ab = "major", "большая", "б."
	case interval.Minor:
		en, ru, ab = "minor", "малая", "м."
	case interval.Augmented:
		en, ru, ab = "augmented", "увеличенная", "ув."
	case interval.Diminished:
		en, ru, ab = "diminished", "уменьшенная", "ум."
	case interval.Neutral:
		var out names

		switch n {
		case 2, 3, 6, 7, 9:
			out.en, out.enOK = "neutral "+sizeEnglish[n], true
		}

		if n == 3 {
			out.ru, out.ruOK = "нейтральная терция", true
		}

		return out
	default:
		return names{}
	}

	switch {
	case k > 3:
		return names{}
	case k == 3:
		return names{en: "triply " + en + " " + sizeEnglish[n], enOK: true, ru: "трижды " + ru + " " + sizeRussian[n], ruOK: true}
	case k == 2:
		en, ru, ab = "doubly "+en, "дважды "+ru, "дв."+ab
	}

	return names{
		en: en + " " + sizeEnglish[n], enOK: true,
		ru: ru + " " + sizeRussian[n], ruOK: true,
		abbr: ab + strconv.Itoa(n), abbrOK: true,
	}
}

// readNames calls the three methods.
func readNames(t *testing.T, i interval.Interval) names {
	t.Helper()

	var (
		out names
		err error
	)

	out.en, out.enOK, err = i.Name()
	require.NoError(t, err)
	out.ru, out.ruOK, err = i.RussianName()
	require.NoError(t, err)
	out.abbr, out.abbrOK, err = i.RussianAbbreviation()
	require.NoError(t, err)

	return out
}

func TestNamesExamples(t *testing.T) {
	t.Parallel()

	tests := []struct {
		i    string
		want names
	}{
		{"M3", names{"major third", "большая терция", "б.3", true, true, true}},
		{"↓M3", names{"major third", "большая терция", "б.3", true, true, true}},
		{"P5", names{"perfect fifth", "чистая квинта", "ч.5", true, true, true}},
		{"A4", names{"augmented fourth", "увеличенная кварта", "ув.4", true, true, true}},
		{"dd5", names{"doubly diminished fifth", "дважды уменьшенная квинта", "дв.ум.5", true, true, true}},
		{"P1", names{"perfect unison", "чистая прима", "ч.1", true, true, true}},
		{"P15", names{"perfect fifteenth", "чистая квинтдецима", "ч.15", true, true, true}},
		{"AAA4", names{"triply augmented fourth", "трижды увеличенная кварта", "", true, true, false}},
		{"ddd5", names{"triply diminished fifth", "трижды уменьшенная квинта", "", true, true, false}},
		{"n3", names{"neutral third", "нейтральная терция", "", true, true, false}},
		{"↓n3", names{"neutral third", "нейтральная терция", "", true, true, false}},
		{"n2", names{"neutral second", "", "", true, false, false}},
		{"n9", names{"neutral ninth", "", "", true, false, false}},
		{"n10", names{}},
		{"n13", names{}},
		{"M16", names{}},
		{"P22", names{}},
		{"AAAA4", names{}},
		{"{+1/2}3", names{}},
		{"{-1/2}5", names{}},
		{"{+1/2}1", names{}},
		{"d3", names{"diminished third", "уменьшенная терция", "ум.3", true, true, true}},
		{"m13", names{"minor thirteenth", "малая терцдецима", "м.13", true, true, true}},
		{"A11", names{"augmented eleventh", "увеличенная ундецима", "ув.11", true, true, true}},
		{"dd8", names{"doubly diminished octave", "дважды уменьшенная октава", "дв.ум.8", true, true, true}},
		{"AA1", names{"doubly augmented unison", "дважды увеличенная прима", "дв.ув.1", true, true, true}},
		{"M14", names{"major fourteenth", "большая квартдецима", "б.14", true, true, true}},
		{"d12", names{"diminished twelfth", "уменьшенная дуодецима", "ум.12", true, true, true}},
		{"m6", names{"minor sixth", "малая секста", "м.6", true, true, true}},
		{"P8", names{"perfect octave", "чистая октава", "ч.8", true, true, true}},
	}

	for _, tt := range tests {
		t.Run(tt.i, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, readNames(t, iv(tt.i)))
			assert.Equal(t, tt.want, expectedNames(iv(tt.i)), "the generator of the sweep agrees with KP-16")
		})
	}

	t.Run("major of MaxSize", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, names{}, readNames(t, basicSize(interval.MaxSize)))
	})
}

// TestNamesSweep covers F-6: names are given only by table 2.7, every given
// name parses back into the ascending interval.
func TestNamesSweep(t *testing.T) {
	t.Parallel()

	var set []interval.Interval

	for n := 1; n <= 17; n++ {
		for d := -6; d <= 5; d++ {
			if n == 1 && d < 0 {
				continue
			}

			set = append(set, interval.MustFromDeviation(n, ai(d)))
		}

		for _, d := range []alteration.Alteration{alt(-1, 2), alt(1, 2), alt(-1, 4), alt(-3, 2)} {
			if n == 1 && d.Sign() < 0 {
				continue
			}

			set = append(set, interval.MustFromDeviation(n, d))
		}
	}

	for _, i := range both(set) {
		want := expectedNames(i)
		got := readNames(t, i)
		assert.Equal(t, want, got, "%v", i)

		for _, s := range []struct {
			text string
			ok   bool
		}{{got.en, got.enOK}, {got.ru, got.ruOK}, {got.abbr, got.abbrOK}} {
			if !s.ok {
				assert.Empty(t, s.text, "%v", i)

				continue
			}

			parsed, err := interval.ParseName(s.text)
			if assert.NoError(t, err, "%q", s.text) {
				assert.True(t, parsed.Equal(i.MustUp()), "ParseName(%q) = %v, want %v", s.text, parsed, i.MustUp())
			}
		}
	}
}

func TestNamesUnset(t *testing.T) {
	t.Parallel()

	var u interval.Interval

	for _, f := range []func() (string, bool, error){u.Name, u.RussianName, u.RussianAbbreviation} {
		name, ok, err := f()
		assertErr(t, err, interval.ErrUnset)
		assert.Empty(t, name)
		assert.False(t, ok)
	}
}

func TestNamesParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text string
		want string
	}{
		{"unison", "P1"},
		{"Unison", "P1"},
		{"prime", "P1"},
		{"perfect prime", "P1"},
		{"perfect unison", "P1"},
		{"PERFECT UNISON", "P1"},
		{"double octave", "P15"},
		{"tritave", "P12"},
		{"Major Third", "M3"},
		{"major third", "M3"},
		{"major 3rd", "M3"},
		{"perfect 8ve", "P8"},
		{"perfect 8va", "P8"},
		{"perfect octave", "P8"},
		{"medium second", "n2"},
		{"neutral second", "n2"},
		{"neutral third", "n3"},
		{"neutral ninth", "n9"},
		{"double augmented fourth", "AA4"},
		{"doubly augmented fourth", "AA4"},
		{"doubly diminished fifth", "dd5"},
		{"triply diminished fifth", "ddd5"},
		{"triply augmented unison", "AAA1"},
		{"augmented 2nd", "A2"},
		{"minor 7th", "m7"},
		{"perfect 4th", "P4"},
		{"perfect 5th", "P5"},
		{"minor 6th", "m6"},
		{"major 9th", "M9"},
		{"minor 10th", "m10"},
		{"perfect 11th", "P11"},
		{"perfect 12th", "P12"},
		{"major 13th", "M13"},
		{"major 14th", "M14"},
		{"perfect 15th", "P15"},
		{"perfect fifteenth", "P15"},
		{"diminished fifteenth", "d15"},
		{"уменьшённая квинта", "d5"},
		{"уменьшенная квинта", "d5"},
		{"ум. 5", "d5"},
		{"ум5", "d5"},
		{"ум.5", "d5"},
		{"ум 5", "d5"},
		{"дв. ум. 5", "dd5"},
		{"дв.ум.5", "dd5"},
		{"дв ум 5", "dd5"},
		{"двум5", "dd5"},
		{"дв.ув.4", "AA4"},
		{"б3", "M3"},
		{"б.3", "M3"},
		{"б. 3", "M3"},
		{"Б.3", "M3"},
		{"м.13", "m13"},
		{"ч.15", "P15"},
		{"ч.1", "P1"},
		{"ЧИСТАЯ КВИНТА", "P5"},
		{"чистая прима", "P1"},
		{"большая терция", "M3"},
		{"малая секста", "m6"},
		{"увеличенная кварта", "A4"},
		{"дважды увеличенная кварта", "AA4"},
		{"трижды увеличенная кварта", "AAA4"},
		{"трижды уменьшённая квинта", "ddd5"},
		{"нейтральная терция", "n3"},
		{"чистая квинтдецима", "P15"},
		{"большая квартдецима", "M14"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()

			got, err := interval.ParseName(tt.text)
			require.NoError(t, err)
			assert.True(t, got.Equal(iv(tt.want)), "ParseName(%q) = %v, want %s", tt.text, got, tt.want)
		})
	}
}

func TestNamesParseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		text string
		want error
	}{
		{"diminished unison", interval.ErrInvalidArgument},
		{"major fifth", interval.ErrInvalidArgument},
		{"perfect third", interval.ErrInvalidArgument},
		{"neutral fourth", interval.ErrInvalidArgument},
		{"doubly major third", interval.ErrInvalidArgument},
		{"triply perfect fifth", interval.ErrInvalidArgument},
		{"уменьшенная прима", interval.ErrInvalidArgument},
		{"чистая терция", interval.ErrInvalidArgument},
		{"большая квинта", interval.ErrInvalidArgument},
		{"ум.1", interval.ErrInvalidArgument},
		{"ч.3", interval.ErrInvalidArgument},
		{"б.5", interval.ErrInvalidArgument},
		{"tritone", interval.ErrParse},
		{"Tritone", interval.ErrParse},
		{"octave", interval.ErrParse},
		{"major", interval.ErrParse},
		{"third", interval.ErrParse},
		{"1st", interval.ErrParse},
		{"perfect 1st", interval.ErrParse},
		{"perfect 8th", interval.ErrParse},
		{"major 16th", interval.ErrParse},
		{"16th", interval.ErrParse},
		{"major sixteenth", interval.ErrParse},
		{"sixteenth", interval.ErrParse},
		{"neutral tenth", interval.ErrParse},
		{"medium third", interval.ErrParse},
		{"quadruply augmented fourth", interval.ErrParse},
		{"нейтральная секунда", interval.ErrParse},
		{"трижды ув.4", interval.ErrParse},
		{"major  third", interval.ErrParse},
		{" major third", interval.ErrParse},
		{"major third ", interval.ErrParse},
		{"major\tthird", interval.ErrParse},
		{"чистая  квинта", interval.ErrParse},
		{"большая 3", interval.ErrParse},
		{"ч.0", interval.ErrParse},
		{"ч.05", interval.ErrParse},
		{"ч.16", interval.ErrParse},
		{"б.16", interval.ErrParse},
		{"ч.99999999999999999999", interval.ErrParse},
		{"ч", interval.ErrParse},
		{"M3", interval.ErrParse},
		{"", interval.ErrParse},
	}

	for _, tt := range tests {
		t.Run(strconv.Quote(tt.text), func(t *testing.T) {
			t.Parallel()

			got, err := interval.ParseName(tt.text)
			assertErr(t, err, tt.want)
			assert.Equal(t, interval.Interval{}, got)
		})
	}
}
