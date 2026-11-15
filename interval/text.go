package interval

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/go-muse/musex/alteration"
)

// maxLetterSigns is the largest multiplicity written by repeating A or d in
// the short notation; larger multiplicities use the deviation in braces.
const maxLetterSigns = 3

// String returns the short notation of i, the same text as MarshalText:
// "M3", "-m6", "AA4", "n3", "{+1/4}3". The unset interval returns
// "Interval(unset)", which is a debugging mark and not a notation.
func (i Interval) String() string {
	if !i.IsValid() {
		return "Interval(unset)"
	}
	var buf [64]byte
	return string(i.appendText(buf[:0]))
}

// MarshalText implements encoding.TextMarshaler. It returns the short
// notation of i, the machine form for storage and exchange:
//
//	text      = [ "-" ] quality size
//	quality   = "P" / "M" / "m" / "n" / 1*3"A" / 1*3"d" / "{" deviation "}"
//	deviation = sign number [ "/" denominator ]   ; d itself, lowest terms
//	size      = nonzero-digit *digit              ; no leading zeros
//
// "-" marks a descending interval. P is perfect, M major, m minor, n neutral;
// A and d repeated k times are augmented and diminished of multiplicity
// k <= 3. Other qualities, fractional ones and multiplicities k >= 4, are
// written as the deviation d from the basic kind in braces, with the sign:
// "{+1/4}3" is a third 1/4 wider than major, "{+4}4" an augmented fourth of
// multiplicity 4. P1 is always "P1".
//
// It returns ErrUnset for the unset interval. In JSON, an interval is a
// string with its short notation; for an optional field, use the omitzero
// tag so that the unset interval is not encoded.
func (i Interval) MarshalText() ([]byte, error) {
	if !i.IsValid() {
		return nil, opError("MarshalText", ErrUnset)
	}
	return i.appendText(nil), nil
}

// AppendText implements encoding.TextAppender. It appends the short notation
// of i to b and returns the extended buffer; it never changes b[:len(b)].
// It returns b and ErrUnset for the unset interval.
func (i Interval) AppendText(b []byte) ([]byte, error) {
	if !i.IsValid() {
		return b, opError("AppendText", ErrUnset)
	}
	return i.appendText(b), nil
}

// appendText appends the short notation of the set interval i to b.
func (i Interval) appendText(b []byte) []byte {
	if i.down {
		b = append(b, '-')
	}
	q, k := qualityOf(perfectSize(i.size), i.dev)
	switch {
	case q == Perfect:
		b = append(b, 'P')
	case q == Major:
		b = append(b, 'M')
	case q == Minor:
		b = append(b, 'm')
	case q == Neutral:
		b = append(b, 'n')
	case q == Augmented && k <= maxLetterSigns:
		for range k {
			b = append(b, 'A')
		}
	case q == Diminished && k <= maxLetterSigns:
		for range k {
			b = append(b, 'd')
		}
	default:
		num, den := i.dev.Fraction()
		b = append(b, '{')
		if num > 0 {
			b = append(b, '+')
		}
		b = strconv.AppendInt(b, int64(num), 10)
		if den != 1 {
			b = append(b, '/')
			b = strconv.AppendInt(b, int64(den), 10)
		}
		b = append(b, '}')
	}
	return strconv.AppendInt(b, int64(i.size), 10)
}

// UnmarshalText implements encoding.TextUnmarshaler. It reads text as Parse
// does and stores the interval in i. On error i is not changed.
//
// It returns ErrUnset for a nil receiver; otherwise the errors of Parse.
func (i *Interval) UnmarshalText(text []byte) error {
	if i == nil {
		return opError("UnmarshalText", ErrUnset)
	}
	v, err := Parse(string(text))
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// qualitySign is a parsed sign of quality: a kind with its multiplicity.
type qualitySign struct {
	kind  Quality
	count int
}

// singleSigns holds the quality signs that are not repeated. The one-letter
// signs are case-sensitive; the longer ones are matched ignoring ASCII case
// in parseQualitySign.
var singleSigns = map[string]Quality{
	"P": Perfect, "M": Major, "m": Minor, "n": Neutral,
	"o": Diminished, "+": Augmented, "°": Diminished,
}

// foldedSigns holds the quality signs read ignoring ASCII case.
var foldedSigns = map[string]Quality{
	"ma": Major, "maj": Major, "mi": Minor, "min": Minor,
	"perf": Perfect, "aug": Augmented, "dim": Diminished,
}

// parseQualitySign reads the sign of quality s, which is not a deviation in
// braces, and reports whether it is one of the accepted signs.
func parseQualitySign(s string) (qualitySign, bool) {
	if q, ok := singleSigns[s]; ok {
		return qualitySign{kind: q, count: 1}, true
	}
	if len(s) >= 2 && len(s) <= 4 {
		if q, ok := foldedSigns[strings.ToLower(s)]; ok {
			return qualitySign{kind: q, count: 1}, true
		}
	}
	if s == "" || (s[0] != 'A' && s[0] != 'd') {
		return qualitySign{}, false
	}
	for j := 1; j < len(s); j++ {
		if s[j] != s[0] {
			return qualitySign{}, false
		}
	}
	if s[0] == 'A' {
		return qualitySign{kind: Augmented, count: len(s)}, true
	}
	return qualitySign{kind: Diminished, count: len(s)}, true
}

// isDigits reports whether s is a non-empty string of ASCII digits without a
// leading zero.
func isDigits(s string) bool {
	if s == "" || s[0] == '0' {
		return false
	}
	for j := range len(s) {
		if s[j] < '0' || s[j] > '9' {
			return false
		}
	}
	return true
}

// parsedSize is a size read from digits: its value when it fits in an int,
// and its simple size, which is known from the digits for any value.
type parsedSize struct {
	value  int
	fits   bool
	simple int
}

// parseSize reads the digits of a size, which satisfy isDigits.
func parseSize(s string) parsedSize {
	v, err := strconv.Atoi(s)
	if err == nil {
		return parsedSize{value: v, fits: true, simple: simpleSizeOf(v)}
	}
	// The value exceeds MaxSize, so it is not 1: the simple size follows
	// from the value modulo 7.
	m := 0
	for j := range len(s) {
		m = (m*10 + int(s[j]-'0')) % 7
	}
	return parsedSize{simple: (m+5)%7 + 2}
}

// parseDeviation reads the deviation inside braces in canonical form and
// reports whether it matches the grammar. A value outside the range of package
// alteration is reported by inRange = false.
func parseDeviation(s string) (d alteration.Alteration, negative, inRange, ok bool) {
	if len(s) < 2 || (s[0] != '+' && s[0] != '-') {
		return alteration.Alteration{}, false, false, false
	}
	negative = s[0] == '-'
	numText, denText, hasDen := strings.Cut(s[1:], "/")
	if !isDigits(numText) || (hasDen && !isDigits(denText)) {
		return alteration.Alteration{}, false, false, false
	}
	if !hasDen {
		num, err := strconv.Atoi(numText)
		if err != nil {
			return alteration.Alteration{}, negative, false, true
		}
		if negative {
			num = -num
		}
		a, err := alteration.FromInt(num)
		return a, negative, err == nil, true
	}
	num, numErr := strconv.Atoi(numText)
	den, denErr := strconv.Atoi(denText)
	if numErr != nil || denErr != nil {
		// Large numbers: only the canonical form is checked; the value is
		// out of range because the denominator or the magnitude is too big.
		var n, m big.Int
		n.SetString(numText, 10)
		m.SetString(denText, 10)
		var g big.Int
		if m.Cmp(big.NewInt(1)) == 0 || g.GCD(nil, nil, &n, &m).Cmp(big.NewInt(1)) != 0 {
			return alteration.Alteration{}, false, false, false
		}
		return alteration.Alteration{}, negative, false, true
	}
	if den == 1 || gcd64(int64(num), int64(den)) != 1 {
		return alteration.Alteration{}, false, false, false
	}
	if negative {
		num = -num
	}
	a, err := alteration.FromFraction(num, den)
	return a, negative, err == nil, true
}

// Parse reads an interval from its short notation, as written by
// Interval.String and Interval.MarshalText, and from other signs of quality:
//
//	input = [ "-" ] sign size
//	sign  = "P" / "M" / "m" / "n" / "o" / "+" / "°" / 1*"A" / 1*"d"  ; case-sensitive
//	      / "ma" / "mi" / "perf" / "maj" / "min" / "aug" / "dim"     ; ASCII case-insensitive
//	      / "{" deviation "}"
//
// The sign is the whole text between the optional "-" and the first digit of
// the size, or the group in braces, and must match one of the variants: ma,
// maj, and M are major; mi, min, and m minor; perf and P perfect; aug, +, and
// A augmented; dim, ° (U+00B0), o, and d diminished; n neutral. Only A and d
// are repeated, any number of times. The deviation in braces is any nonzero
// value in canonical form ("{+1}4" is A4). There are no spaces. A "-" before
// a perfect unison gives the perfect unison, which has no direction.
// Parse(i.String()) equals i for every set interval i.
//
// Errors, in this order: ErrParse for text that does not match the grammar,
// including the empty string, "TT", "++4", "a4", "M 3", leading zeros, a size
// 0, and a deviation that is not in canonical form; ErrInvalidArgument when
// the sign does not fit the group of the size ("M5", "P3", "n4") or gives a
// diminished unison ("d1", "{-1}1"); ErrOutOfRange when the size is greater
// than MaxSize or, together with alteration.ErrOutOfRange, when the deviation
// is outside the range of package alteration. The text of the error quotes at
// most 64 bytes of the input.
func Parse(text string) (Interval, error) {
	s := text
	down := false
	if strings.HasPrefix(s, "-") {
		down, s = true, s[1:]
	}
	var (
		sign      qualitySign
		dev       alteration.Alteration
		braced    bool
		negative  bool
		inRange   bool
		sizeDigit string
	)
	if strings.HasPrefix(s, "{") {
		inner, rest, found := strings.Cut(s[1:], "}")
		if !found {
			return Interval{}, parseError(classParse, text)
		}
		var ok bool
		dev, negative, inRange, ok = parseDeviation(inner)
		if !ok {
			return Interval{}, parseError(classParse, text)
		}
		braced, sizeDigit = true, rest
	} else {
		k := strings.IndexFunc(s, func(r rune) bool { return r >= '0' && r <= '9' })
		if k < 0 {
			return Interval{}, parseError(classParse, text)
		}
		var ok bool
		if sign, ok = parseQualitySign(s[:k]); !ok {
			return Interval{}, parseError(classParse, text)
		}
		sizeDigit = s[k:]
	}
	if !isDigits(sizeDigit) {
		return Interval{}, parseError(classParse, text)
	}
	size := parseSize(sizeDigit)
	perfect := false
	switch size.simple {
	case 1, 4, 5, 8:
		perfect = true
	}
	if braced {
		if size.simple == 1 && size.fits && negative {
			return Interval{}, parseError(classInvalid, text)
		}
		if !size.fits {
			return Interval{}, parseError(classSizeRange, text)
		}
		if !inRange {
			return Interval{}, parseError(classDeviationRange, text)
		}
		return newInterval(size.value, dev, down), nil
	}
	d, class := signDeviation(sign, perfect, size.fits && size.value == 1)
	if class == classInvalid {
		return Interval{}, parseError(classInvalid, text)
	}
	if !size.fits {
		return Interval{}, parseError(classSizeRange, text)
	}
	if class == classDeviationRange {
		return Interval{}, parseError(classDeviationRange, text)
	}
	return newInterval(size.value, d, down), nil
}

// signDeviation returns the deviation of the quality sign in the group given
// by perfect and the class of the failure: classParse for success,
// classInvalid for a sign that does not fit, and classDeviationRange for a
// deviation out of range.
func signDeviation(sign qualitySign, perfect, unison bool) (alteration.Alteration, textClass) {
	switch sign.kind {
	case Perfect:
		if !perfect {
			return alteration.Alteration{}, classInvalid
		}
		return alteration.Natural(), classParse
	case Major, Minor, Neutral:
		if perfect {
			return alteration.Alteration{}, classInvalid
		}
		switch sign.kind {
		case Minor:
			return alteration.Flat(), classParse
		case Neutral:
			return neutralDeviation, classParse
		}
		return alteration.Natural(), classParse
	case Augmented:
		d, err := alteration.FromInt(sign.count)
		if err != nil {
			return alteration.Alteration{}, classDeviationRange
		}
		return d, classParse
	default:
		if unison {
			return alteration.Alteration{}, classInvalid
		}
		// -count - (offset - 1) never overflows: count >= 1.
		d, err := alteration.FromInt(-sign.count - (basicOffset(perfect) - 1))
		if err != nil {
			return alteration.Alteration{}, classDeviationRange
		}
		return d, classParse
	}
}

// MustParse is like Parse but panics if Parse returns an error. The panic
// value is the error returned by Parse.
func MustParse(text string) Interval {
	return must(Parse(text))
}
