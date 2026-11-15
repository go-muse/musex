package interval

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// maxNamedSize is the largest size that has a name: the fifteenth.
const maxNamedSize = 15

// englishSizes holds the English words of the sizes 1 to 15.
var englishSizes = [maxNamedSize + 1]string{
	"", "unison", "second", "third", "fourth", "fifth", "sixth", "seventh",
	"octave", "ninth", "tenth", "eleventh", "twelfth", "thirteenth",
	"fourteenth", "fifteenth",
}

// englishOrdinals holds the English sizes written with digits, accepted by
// ParseName.
var englishOrdinals = [maxNamedSize + 1][]string{
	2: {"2nd"}, 3: {"3rd"}, 4: {"4th"}, 5: {"5th"}, 6: {"6th"}, 7: {"7th"},
	8: {"8ve", "8va"}, 9: {"9th"}, 10: {"10th"}, 11: {"11th"}, 12: {"12th"},
	13: {"13th"}, 14: {"14th"}, 15: {"15th"},
}

// russianSizes holds the Russian words of the sizes 1 to 15.
var russianSizes = [maxNamedSize + 1]string{
	"", "прима", "секунда", "терция", "кварта", "квинта", "секста", "септима",
	"октава", "нона", "децима", "ундецима", "дуодецима", "терцдецима",
	"квартдецима", "квинтдецима",
}

// englishNeutral reports whether the neutral interval of size n has an
// English name: the second, third, sixth, seventh, and ninth.
func englishNeutral(n int) bool {
	switch n {
	case 2, 3, 6, 7, 9:
		return true
	default:
		return false
	}
}

// russianNeutral reports whether the neutral interval of size n has a
// Russian name: only the third.
func russianNeutral(n int) bool {
	return n == 3
}

// nameForm holds the words of one form of names.
type nameForm struct {
	perfect, major, minor, augmented, diminished, neutral string
	doubly, triply                                        string
	neutralNamed                                          func(int) bool
}

var (
	englishForm = nameForm{
		perfect: "perfect", major: "major", minor: "minor",
		augmented: "augmented", diminished: "diminished", neutral: "neutral",
		doubly: "doubly ", triply: "triply ", neutralNamed: englishNeutral,
	}
	russianForm = nameForm{
		perfect: "чистая", major: "большая", minor: "малая",
		augmented: "увеличенная", diminished: "уменьшенная", neutral: "нейтральная",
		doubly: "дважды ", triply: "трижды ", neutralNamed: russianNeutral,
	}
	abbreviationForm = nameForm{
		perfect: "ч.", major: "б.", minor: "м.",
		augmented: "ув.", diminished: "ум.",
		doubly: "дв.", neutralNamed: func(int) bool { return false },
	}
)

// qualityWords returns the words of the quality of the set interval i in the
// form f and reports whether the form has them.
func (f *nameForm) qualityWords(i Interval) (string, bool) {
	q, k := qualityOf(perfectSize(i.size), i.dev)
	var word string
	switch q {
	case Perfect:
		return f.perfect, true
	case Major:
		return f.major, true
	case Minor:
		return f.minor, true
	case Neutral:
		return f.neutral, f.neutralNamed(i.size)
	case Augmented:
		word = f.augmented
	case Diminished:
		word = f.diminished
	default:
		return "", false
	}
	switch {
	case k == 1:
		return word, true
	case k == 2 && f.doubly != "":
		return f.doubly + word, true
	case k == 3 && f.triply != "":
		return f.triply + word, true
	default:
		return "", false
	}
}

// name returns the name of i in the form f, with the size given by size.
func (i Interval) name(op string, f *nameForm, size func(int) string) (string, bool, error) {
	if !i.IsValid() {
		return "", false, opError(op, ErrUnset)
	}
	if i.size > maxNamedSize {
		return "", false, nil
	}
	q, ok := f.qualityWords(i)
	if !ok {
		return "", false, nil
	}
	return q + size(i.size), true, nil
}

// Name returns the English name of i: the quality, then the size, such as
// "major third", "doubly diminished fifth", "perfect fifteenth", or "neutral
// third". The direction is not part of the name.
//
// Only the names of the domain base are given: sizes 1 to 15, multiplicities
// up to 3 ("doubly", "triply"), and the neutral second, third, sixth,
// seventh, and ninth. For other intervals Name returns "", false, and no
// error. It returns ErrUnset for the unset interval.
func (i Interval) Name() (name string, ok bool, err error) {
	return i.name("Name", &englishForm, func(n int) string { return " " + englishSizes[n] })
}

// MustName is like Name but panics if Name returns an error. The panic value
// is the error returned by Name.
func (i Interval) MustName() (name string, ok bool) {
	return must2(i.Name())
}

// RussianName returns the Russian name of i: the quality, then the size,
// such as "большая терция", "дважды уменьшенная квинта", or "нейтральная
// терция". Only the names of the domain base are given: sizes 1 to 15,
// multiplicities up to 3 ("дважды", "трижды"), and the neutral third. For
// other intervals RussianName returns "", false, and no error. It returns
// ErrUnset for the unset interval.
func (i Interval) RussianName() (name string, ok bool, err error) {
	return i.name("RussianName", &russianForm, func(n int) string { return " " + russianSizes[n] })
}

// MustRussianName is like RussianName but panics if RussianName returns an
// error. The panic value is the error returned by RussianName.
func (i Interval) MustRussianName() (name string, ok bool) {
	return must2(i.RussianName())
}

// RussianAbbreviation returns the Russian abbreviation of i without spaces,
// such as "б.3", "ч.15", "ув.4", or "дв.ум.5". There are no abbreviations for
// triply augmented and diminished intervals, for neutral intervals, and for
// sizes above 15: RussianAbbreviation returns "", false, and no error. It
// returns ErrUnset for the unset interval.
func (i Interval) RussianAbbreviation() (name string, ok bool, err error) {
	return i.name("RussianAbbreviation", &abbreviationForm, strconv.Itoa)
}

// MustRussianAbbreviation is like RussianAbbreviation but panics if
// RussianAbbreviation returns an error. The panic value is the error returned
// by RussianAbbreviation.
func (i Interval) MustRussianAbbreviation() (name string, ok bool) {
	return must2(i.RussianAbbreviation())
}

// lexicon holds the words accepted by ParseName for the full names of one
// language.
type lexicon struct {
	mults        map[string]int
	qualities    map[string]Quality
	size         func(string) (int, bool)
	neutralNamed func(int) bool
	// medium is a quality word allowed only in "medium second".
	medium string
}

// foldLookup returns the value of the key of m equal to w under Unicode case
// folding.
func foldLookup[V any](m map[string]V, w string) (V, bool) {
	for k, v := range m {
		if strings.EqualFold(k, w) {
			return v, true
		}
	}
	var zero V
	return zero, false
}

var (
	englishLexicon = lexicon{
		mults: map[string]int{"doubly": 2, "double": 2, "triply": 3},
		qualities: map[string]Quality{
			"perfect": Perfect, "major": Major, "minor": Minor,
			"augmented": Augmented, "diminished": Diminished,
			"neutral": Neutral, "medium": Neutral,
		},
		size:         englishSize,
		neutralNamed: englishNeutral,
		medium:       "medium",
	}
	russianLexicon = lexicon{
		mults: map[string]int{"дважды": 2, "трижды": 3},
		qualities: map[string]Quality{
			"чистая": Perfect, "большая": Major, "малая": Minor,
			"увеличенная": Augmented, "уменьшенная": Diminished,
			"уменьшённая": Diminished, "нейтральная": Neutral,
		},
		size:         russianSize,
		neutralNamed: russianNeutral,
	}
)

// englishSize returns the size of an English size word or ordinal.
func englishSize(w string) (int, bool) {
	for n := 1; n <= maxNamedSize; n++ {
		if strings.EqualFold(englishSizes[n], w) {
			return n, true
		}
		for _, o := range englishOrdinals[n] {
			if strings.EqualFold(o, w) {
				return n, true
			}
		}
	}
	return 0, false
}

// russianSize returns the size of a Russian size word.
func russianSize(w string) (int, bool) {
	for n := 1; n <= maxNamedSize; n++ {
		if strings.EqualFold(russianSizes[n], w) {
			return n, true
		}
	}
	return 0, false
}

// nameParse is the result of reading a name: the interval or the class of
// the failure, and whether the text has the structure of the form at all.
type nameParse struct {
	interval Interval
	class    textClass
	matched  bool
}

// failed returns a nameParse with the failure class c.
func failed(c textClass) nameParse {
	return nameParse{class: c, matched: true}
}

// buildNamed returns the ascending interval of the parsed parts, or the class
// of the failure: classInvalid for parts without an interval and classParse
// for an interval without a name in the form.
func buildNamed(kind Quality, mult, size int, neutralNamed func(int) bool) nameParse {
	if mult > 1 && kind != Augmented && kind != Diminished {
		return failed(classInvalid)
	}
	perfect := perfectSize(size)
	switch kind {
	case Perfect:
		if !perfect {
			return failed(classInvalid)
		}
	case Major, Minor, Neutral:
		if perfect {
			return failed(classInvalid)
		}
	case Diminished:
		if size == 1 {
			return failed(classInvalid)
		}
	}
	if kind == Neutral && !neutralNamed(size) {
		return failed(classParse)
	}
	var (
		i   Interval
		err error
	)
	if kind == Augmented || kind == Diminished {
		i, err = NewMultiple(mult, kind, size)
	} else {
		i, err = New(kind, size)
	}
	if err != nil {
		return failed(classInvalid)
	}
	return nameParse{interval: i, matched: true}
}

// parseFullName reads words as a full name of the lexicon.
func parseFullName(words []string, lex *lexicon) nameParse {
	mult := 1
	switch len(words) {
	case 2:
	case 3:
		m, ok := foldLookup(lex.mults, words[0])
		if !ok {
			return nameParse{}
		}
		mult = m
	default:
		return nameParse{}
	}
	qw, sw := words[len(words)-2], words[len(words)-1]
	kind, ok := foldLookup(lex.qualities, qw)
	if !ok {
		return nameParse{}
	}
	size, ok := lex.size(sw)
	if !ok {
		return nameParse{}
	}
	if lex.medium != "" && strings.EqualFold(qw, lex.medium) && (size != 2 || mult != 1) {
		return failed(classParse)
	}
	return buildNamed(kind, mult, size, lex.neutralNamed)
}

// cutPrefixFold returns s without the prefix p, compared under Unicode case
// folding rune by rune, and reports whether s starts with p.
func cutPrefixFold(s, p string) (string, bool) {
	k := 0
	for range utf8.RuneCountInString(p) {
		if k >= len(s) {
			return s, false
		}
		_, w := utf8.DecodeRuneInString(s[k:])
		k += w
	}
	if !strings.EqualFold(s[:k], p) {
		return s, false
	}
	return s[k:], true
}

// cutSeparator removes one separator of an abbreviation from s: ". ", ".",
// " ", or nothing.
func cutSeparator(s string) string {
	switch {
	case strings.HasPrefix(s, ". "):
		return s[2:]
	case strings.HasPrefix(s, "."), strings.HasPrefix(s, " "):
		return s[1:]
	default:
		return s
	}
}

// abbreviationSigns holds the quality signs of Russian abbreviations, longer
// signs first.
var abbreviationSigns = [...]struct {
	sign string
	kind Quality
}{
	{"ув", Augmented}, {"ум", Diminished}, {"ч", Perfect}, {"б", Major}, {"м", Minor},
}

// parseAbbreviation reads text as a Russian abbreviation:
// [дв sep] (ч|б|м|ув|ум) sep digits, with sep one of "", ".", " ", ". ".
func parseAbbreviation(text string) nameParse {
	s := text
	mult := 1
	if rest, ok := cutPrefixFold(s, "дв"); ok {
		mult, s = 2, cutSeparator(rest)
	}
	var kind Quality
	for _, a := range abbreviationSigns {
		if rest, ok := cutPrefixFold(s, a.sign); ok {
			kind, s = a.kind, cutSeparator(rest)
			break
		}
	}
	if kind == QualityUnset || (mult == 2 && kind != Augmented && kind != Diminished) {
		return nameParse{}
	}
	if !isDigits(s) {
		return nameParse{}
	}
	size, err := strconv.Atoi(s)
	if err != nil || size > maxNamedSize {
		return nameParse{}
	}
	return buildNamed(kind, mult, size, func(int) bool { return false })
}

// specialNames holds the names accepted by ParseName beyond the regular
// forms.
var specialNames = []struct {
	words    []string
	interval Interval
}{
	{[]string{"unison"}, PerfectUnison()},
	{[]string{"prime"}, PerfectUnison()},
	{[]string{"perfect", "prime"}, PerfectUnison()},
	{[]string{"tritave"}, PerfectTwelfth()},
	{[]string{"double", "octave"}, PerfectFifteenth()},
}

// matchWords reports whether words equal want under Unicode case folding.
func matchWords(words, want []string) bool {
	if len(words) != len(want) {
		return false
	}
	for k := range words {
		if !strings.EqualFold(words[k], want[k]) {
			return false
		}
	}
	return true
}

// ParseName reads the name of an interval and returns the ascending interval.
// It accepts every name given by Interval.Name, Interval.RussianName, and
// Interval.RussianAbbreviation, ignoring case, with these additional forms:
//
//   - English: "[doubly|double|triply] quality size", with the size as a word
//     or with digits: 2nd, 3rd, 4th to 7th, 8ve, 8va, 9th to 15th; "medium
//     second" for the neutral second; "unison", "prime", and "perfect prime"
//     for P1, "tritave" for P12, and "double octave" for P15.
//   - Russian: "[дважды|трижды] quality size" with the size as a word only;
//     "уменьшённая" is accepted with ё.
//   - Russian abbreviations: "[дв sep] (ч|б|м|ув|ум) sep digits" with sep
//     empty, ".", " ", or ". ", such as "ч.5", "ч. 5", "ч5", "дв. ум. 5"; дв
//     goes only with ув and ум, and the digits give a size from 1 to 15.
//
// Words are separated by exactly one ASCII space, with no spaces at the ends.
//
// Errors: ErrParse for text that is not a name, including unknown words,
// sizes above 15, intervals that have no name in the form ("neutral tenth",
// "нейтральная секунда", "трижды ув.4"), and extra spaces; ErrInvalidArgument
// for words that form no interval: a quality that does not fit the size
// ("major fifth", "perfect third", "neutral fourth"), a diminished unison, or
// a multiplicity with a quality other than augmented and diminished. The text
// of the error quotes at most 64 bytes of the input.
func ParseName(text string) (Interval, error) {
	words := strings.Split(text, " ")
	empty := false
	for _, w := range words {
		if w == "" {
			empty = true
		}
	}
	if !empty {
		for _, sn := range specialNames {
			if matchWords(words, sn.words) {
				return sn.interval, nil
			}
		}
		for _, lex := range []*lexicon{&englishLexicon, &russianLexicon} {
			if r := parseFullName(words, lex); r.matched {
				return r.result(text)
			}
		}
	}
	if r := parseAbbreviation(text); r.matched {
		return r.result(text)
	}
	return Interval{}, parseError(classParse, text)
}

// result returns the interval of r or the error of its class for text.
func (r nameParse) result(text string) (Interval, error) {
	if r.interval.IsValid() {
		return r.interval, nil
	}
	return Interval{}, parseError(r.class, text)
}

// MustParseName is like ParseName but panics if ParseName returns an error.
// The panic value is the error returned by ParseName.
func MustParseName(text string) Interval {
	return must(ParseName(text))
}
