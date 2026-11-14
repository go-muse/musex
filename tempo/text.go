package tempo

import (
	"errors"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo/term"
)

// Signs of the text grammar.
const (
	exactSign  = "="
	approxSign = "≈" // U+2248, three bytes in UTF-8
	rangeSign  = "-"
	nameSep    = ", "
)

// Strings of unset values. The parse functions reject them.
const (
	unsetTempoString   = "Tempo(unset)"
	unsetApproxString  = "Approx(unset)"
	unsetRangeString   = "Range(unset)"
	unsetSecondsString = "Seconds(unset)"
)

// maxComponentDigits is the number of decimal digits of
// fraction.MaxComponent; a count without leading zeros that is longer is
// outside the range.
const maxComponentDigits = len(fraction.MaxComponentString)

// maxFractionDigits is the largest number of significant digits after the
// decimal point that a number in range can have: a last digit that is not
// zero leaves a factor 2^k or 5^k in the reduced denominator, which is greater
// than fraction.MaxComponent for k > 255.
const maxFractionDigits = fraction.MaxComponentBits

// appendNumber appends the reduced value of the valid fraction f: an integer
// as its digits and any other value as "n/d".
func appendNumber(b []byte, f fraction.Fraction) []byte {
	num, den, _ := f.ReducedBigPair()
	b = num.Append(b, 10)
	if den.IsInt64() && den.Int64() == 1 {
		return b
	}
	b = append(b, '/')
	return den.Append(b, 10)
}

// appendRatio appends the reduced value of the valid fraction f as "n/d".
func appendRatio(b []byte, f fraction.Fraction) []byte {
	num, den, _ := f.ReducedBigPair()
	b = num.Append(b, 10)
	b = append(b, '/')
	return den.Append(b, 10)
}

// appendTempo appends the text of the set tempo t.
func appendTempo(b []byte, t Tempo) []byte {
	return appendMark(b, t.m, exactSign)
}

// String returns the canonical text of t: the unit as duration.Written.String
// writes it, "=", and beats per minute as an integer or a reduced fraction
// "n/d", such as "4=60", "4.=60", "4=185/2", or "8[3:2]=360". For the unset
// tempo it returns "Tempo(unset)", which Parse rejects.
func (t Tempo) String() string {
	if !t.IsSet() {
		return unsetTempoString
	}
	return string(appendTempo(nil, t))
}

// MarshalText implements encoding.TextMarshaler with the text of String. It
// returns nil and an error wrapping ErrUnset if t is unset, so an unset struct
// field is not encoded; use the omitzero tag option for an optional field.
func (t Tempo) MarshalText() ([]byte, error) {
	if !t.IsSet() {
		return nil, t.unset("Tempo.MarshalText")
	}
	return appendTempo(nil, t), nil
}

// AppendText implements encoding.TextAppender: it appends the text of String
// to dst without modifying dst[:len(dst)]. It returns dst unchanged and an
// error wrapping ErrUnset if t is unset.
func (t Tempo) AppendText(dst []byte) ([]byte, error) {
	if !t.IsSet() {
		return dst, t.unset("Tempo.AppendText")
	}
	return appendTempo(dst, t), nil
}

// UnmarshalText implements encoding.TextUnmarshaler with the grammar of Parse.
// On error the receiver is not modified. It returns an error wrapping
// ErrInvalidArgument if t is nil, and otherwise the errors of Parse. The text
// is not retained.
func (t *Tempo) UnmarshalText(text []byte) error {
	if t == nil {
		return invalidError("Tempo.UnmarshalText", "nil receiver")
	}
	v, err := Parse(string(text))
	if err != nil {
		return err
	}
	*t = v
	return nil
}

// String returns the canonical text of a: the unit, "≈", and the named number
// as an integer or a reduced fraction "n/d", such as "4≈108" or "4.≈60". For
// the unset marking it returns "Approx(unset)", which ParseApprox rejects.
func (a Approx) String() string {
	if !a.IsSet() {
		return unsetApproxString
	}
	return string(appendMark(nil, a.m, approxSign))
}

// MarshalText implements encoding.TextMarshaler with the text of String. It
// returns nil and an error wrapping ErrUnset if a is unset.
func (a Approx) MarshalText() ([]byte, error) {
	if !a.IsSet() {
		return nil, a.unset("Approx.MarshalText")
	}
	return appendMark(nil, a.m, approxSign), nil
}

// AppendText implements encoding.TextAppender: it appends the text of String
// to dst without modifying dst[:len(dst)]. It returns dst unchanged and an
// error wrapping ErrUnset if a is unset.
func (a Approx) AppendText(dst []byte) ([]byte, error) {
	if !a.IsSet() {
		return dst, a.unset("Approx.AppendText")
	}
	return appendMark(dst, a.m, approxSign), nil
}

// UnmarshalText implements encoding.TextUnmarshaler with the grammar of
// ParseApprox. On error the receiver is not modified. It returns an error
// wrapping ErrInvalidArgument if a is nil, and otherwise the errors of
// ParseApprox. The text is not retained.
func (a *Approx) UnmarshalText(text []byte) error {
	if a == nil {
		return invalidError("Approx.UnmarshalText", "nil receiver")
	}
	v, err := ParseApprox(string(text))
	if err != nil {
		return err
	}
	*a = v
	return nil
}

// appendRange appends the text of the set range r.
func appendRange(b []byte, r Range) []byte {
	b = appendMark(b, mark{unit: r.unit, bpm: r.lower}, exactSign)
	b = append(b, rangeSign...)
	return appendNumber(b, r.upper)
}

// String returns the canonical text of r: the unit, "=", the lower number,
// "-", and the upper number, such as "8=132-144". For the unset range it
// returns "Range(unset)", which ParseRange rejects.
func (r Range) String() string {
	if !r.IsSet() {
		return unsetRangeString
	}
	return string(appendRange(nil, r))
}

// MarshalText implements encoding.TextMarshaler with the text of String. It
// returns nil and an error wrapping ErrUnset if r is unset.
func (r Range) MarshalText() ([]byte, error) {
	if !r.IsSet() {
		return nil, r.unset("Range.MarshalText")
	}
	return appendRange(nil, r), nil
}

// AppendText implements encoding.TextAppender: it appends the text of String
// to dst without modifying dst[:len(dst)]. It returns dst unchanged and an
// error wrapping ErrUnset if r is unset.
func (r Range) AppendText(dst []byte) ([]byte, error) {
	if !r.IsSet() {
		return dst, r.unset("Range.AppendText")
	}
	return appendRange(dst, r), nil
}

// UnmarshalText implements encoding.TextUnmarshaler with the grammar of
// ParseRange. On error the receiver is not modified. It returns an error
// wrapping ErrInvalidArgument if r is nil, and otherwise the errors of
// ParseRange. The text is not retained.
func (r *Range) UnmarshalText(text []byte) error {
	if r == nil {
		return invalidError("Range.UnmarshalText", "nil receiver")
	}
	v, err := ParseRange(string(text))
	if err != nil {
		return err
	}
	*r = v
	return nil
}

// String returns the reduced number of seconds as "n/d", such as "0/1", "1/3",
// or "2/1". For the unset time it returns "Seconds(unset)", which
// ParseSeconds rejects.
func (s Seconds) String() string {
	if !s.IsSet() {
		return unsetSecondsString
	}
	return string(appendRatio(nil, s.v))
}

// MarshalText implements encoding.TextMarshaler with the text of String. It
// returns nil and an error wrapping ErrUnset if s is unset.
func (s Seconds) MarshalText() ([]byte, error) {
	if !s.IsSet() {
		return nil, s.unset("Seconds.MarshalText")
	}
	return appendRatio(nil, s.v), nil
}

// AppendText implements encoding.TextAppender: it appends the text of String
// to dst without modifying dst[:len(dst)]. It returns dst unchanged and an
// error wrapping ErrUnset if s is unset.
func (s Seconds) AppendText(dst []byte) ([]byte, error) {
	if !s.IsSet() {
		return dst, s.unset("Seconds.AppendText")
	}
	return appendRatio(dst, s.v), nil
}

// UnmarshalText implements encoding.TextUnmarshaler with the grammar of
// ParseSeconds. On error the receiver is not modified. It returns an error
// wrapping ErrInvalidArgument if s is nil, and otherwise the errors of
// ParseSeconds. The text is not retained.
func (s *Seconds) UnmarshalText(text []byte) error {
	if s == nil {
		return invalidError("Seconds.UnmarshalText", "nil receiver")
	}
	v, err := ParseSeconds(string(text))
	if err != nil {
		return err
	}
	*s = v
	return nil
}

// needsQuotes reports whether a name is written in quotation marks in the text
// of an indication: when it contains "=" or "≈" or starts with a quotation
// mark.
func needsQuotes(name string) bool {
	return strings.ContainsAny(name, exactSign+approxSign) || strings.HasPrefix(name, `"`)
}

// appendIndication appends the text of the set indication i.
func appendIndication(b []byte, i Indication) []byte {
	if i.term.IsSet() {
		name := i.term.String()
		if needsQuotes(name) {
			b = strconv.AppendQuote(b, name)
		} else {
			b = append(b, name...)
		}
		if i.kind != NoMark {
			b = append(b, nameSep...)
		}
	}
	switch i.kind {
	case ExactMark:
		b = appendMark(b, mark{unit: i.unit, bpm: i.lower}, exactSign)
	case ApproxMark:
		b = appendMark(b, mark{unit: i.unit, bpm: i.lower}, approxSign)
	case RangeMark:
		b = appendRange(b, Range{unit: i.unit, lower: i.lower, upper: i.upper})
	}
	return b
}

// String returns the text of i: the name of its term, its marking, or both
// separated by ", ", such as "Allegro", "4=60", "Allegro, 4=120", or
// "Moderato, 4≈108". A name that contains "=" or "≈" or starts with a
// quotation mark is written as strconv.Quote writes it, so the term "4=60" of
// a user list prints as "\"4=60\"" and differs from the marking 4=60. For the
// unset indication it returns the empty string.
func (i Indication) String() string {
	if !i.IsSet() {
		return ""
	}
	return string(appendIndication(nil, i))
}

// MarshalText implements encoding.TextMarshaler with the text of String. It
// returns nil and an error wrapping ErrUnset if i is unset. Indication does
// not implement encoding.TextUnmarshaler: an indication is parsed only with a
// named list, by ParseIndication.
func (i Indication) MarshalText() ([]byte, error) {
	if !i.IsSet() {
		return nil, i.unset("Indication.MarshalText")
	}
	return appendIndication(nil, i), nil
}

// AppendText implements encoding.TextAppender: it appends the text of String
// to dst without modifying dst[:len(dst)]. It returns dst unchanged and an
// error wrapping ErrUnset if i is unset.
func (i Indication) AppendText(dst []byte) ([]byte, error) {
	if !i.IsSet() {
		return dst, i.unset("Indication.AppendText")
	}
	return appendIndication(dst, i), nil
}

// isCount reports whether s matches count = nzdigit *digit.
func isCount(s string) bool {
	if s == "" || s[0] < '1' || s[0] > '9' {
		return false
	}
	return isDigits(s[1:])
}

// isDigits reports whether s consists of ASCII decimal digits only.
func isDigits(s string) bool {
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// isInteger reports whether s matches integer = "0" / count.
func isInteger(s string) bool {
	return s == "0" || isCount(s)
}

// isNumber reports whether s matches number = integer / ratio / decimal.
func isNumber(s string) bool {
	if num, den, found := strings.Cut(s, "/"); found {
		return isInteger(num) && isCount(den)
	}
	if whole, frac, found := strings.Cut(s, "."); found {
		return isInteger(whole) && frac != "" && isDigits(frac)
	}
	return isInteger(s)
}

// componentInt returns the integer s as a *big.Int and reports whether it is
// at most fraction.MaxComponent. Longer integers are rejected without
// conversion.
func componentInt(s string) (*big.Int, bool) {
	if len(s) > maxComponentDigits {
		return nil, false
	}
	n, ok := new(big.Int).SetString(s, 10)
	if !ok || n.Cmp(fraction.MaxComponentBig()) > 0 {
		return nil, false
	}
	return n, true
}

// numberValue returns the exact value of a number that matches the grammar
// and reports whether it is within the range: for an integer and a ratio both
// written components are at most fraction.MaxComponent, and for a decimal the
// reduced value is in range. The time is linear in the length of s.
func numberValue(s string) (*big.Rat, bool) {
	if num, den, found := strings.Cut(s, "/"); found {
		n, ok := componentInt(num)
		if !ok {
			return nil, false
		}
		d, ok := componentInt(den)
		if !ok {
			return nil, false
		}
		return new(big.Rat).SetFrac(n, d), true
	}
	whole, frac, _ := strings.Cut(s, ".")
	w, ok := componentInt(whole)
	if !ok {
		return nil, false
	}
	frac = strings.TrimRight(frac, "0")
	if frac == "" {
		return new(big.Rat).SetInt(w), true
	}
	if len(frac) > maxFractionDigits {
		return nil, false
	}
	f, _ := new(big.Int).SetString(frac, 10)
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(len(frac))), nil)
	num := new(big.Int).Mul(w, scale)
	num.Add(num, f)
	r := new(big.Rat).SetFrac(num, scale)
	if _, err := fraction.FromBigRat(r); err != nil {
		return nil, false
	}
	return r, true
}

// parseMarkText parses body as unit sign number, or unit sign number "-"
// number if isRange, in the order of the parse functions: the grammar of the
// whole text, the range of the unit, the range of the numbers from left to
// right, and numbers that are zero from left to right. Errors quote full, the
// whole text passed to the public function.
func parseMarkText(body, sign string, isRange bool, full string) (duration.Written, []*big.Rat, error) {
	unitText, rest, found := strings.Cut(body, sign)
	if !found {
		return duration.Written{}, nil, parseError(classInvalidText, full)
	}
	nums := []string{rest}
	if isRange {
		lo, hi, found := strings.Cut(rest, rangeSign)
		if !found {
			return duration.Written{}, nil, parseError(classInvalidText, full)
		}
		nums = []string{lo, hi}
	}
	for _, n := range nums {
		if !isNumber(n) {
			return duration.Written{}, nil, parseError(classInvalidText, full)
		}
	}
	unit, err := duration.ParseWritten(unitText)
	if err != nil {
		if errors.Is(err, duration.ErrInvalidText) {
			return duration.Written{}, nil, parseError(classInvalidText, full)
		}
		return duration.Written{}, nil, unitRange(full, err)
	}
	values := make([]*big.Rat, len(nums))
	for i, n := range nums {
		v, ok := numberValue(n)
		if !ok {
			return duration.Written{}, nil, parseError(classFractionRange, full)
		}
		values[i] = v
	}
	for _, v := range values {
		if v.Sign() == 0 {
			return duration.Written{}, nil, parseError(classNoTempo, full)
		}
	}
	return unit, values, nil
}

// fractionOf returns the in-range value r as a reduced fraction.
func fractionOf(r *big.Rat) fraction.Fraction {
	f, _ := fraction.FromBigRat(r)
	return f
}

// parseTempo parses body as an exact marking, quoting full in errors.
func parseTempo(body, full string) (Tempo, error) {
	unit, v, err := parseMarkText(body, exactSign, false, full)
	if err != nil {
		return Tempo{}, err
	}
	return Tempo{m: mark{unit: unit, bpm: fractionOf(v[0])}}, nil
}

// parseApprox parses body as an approximate marking, quoting full in errors.
func parseApprox(body, full string) (Approx, error) {
	unit, v, err := parseMarkText(body, approxSign, false, full)
	if err != nil {
		return Approx{}, err
	}
	return Approx{m: mark{unit: unit, bpm: fractionOf(v[0])}}, nil
}

// parseRange parses body as a range, quoting full in errors.
func parseRange(body, full string) (Range, error) {
	unit, v, err := parseMarkText(body, exactSign, true, full)
	if err != nil {
		return Range{}, err
	}
	if v[0].Cmp(v[1]) >= 0 {
		return Range{}, parseError(classInvalidArgument, full)
	}
	return Range{unit: unit, lower: fractionOf(v[0]), upper: fractionOf(v[1])}, nil
}

// Parse returns the exact marking in the canonical text of Tempo.String:
//
//	tempo   = unit "=" number
//	unit    = written                 ; the grammar of duration.ParseWritten
//	number  = integer / ratio / decimal
//	integer = "0" / count
//	ratio   = integer "/" count
//	decimal = integer "." 1*digit
//	count   = nzdigit *digit
//
// The unit is the text before the first "=". A number is reduced, so "120/2"
// gives 60 and "92.5" gives 185/2. Spaces, signs, leading zeros, note signs,
// and other forms are rejected, including "4≈108" and the strings of unset
// values. In order of checking, it returns an error wrapping ErrInvalidText if
// the text does not match the grammar, including a grammar error of the unit;
// ErrOutOfRange together with every sentinel of the error of
// duration.ParseWritten if the unit is outside its range; ErrOutOfRange
// together with fraction.ErrOutOfRange if the number is outside the range of
// package common/fraction; and ErrNoTempo if the number is zero.
// Parse(t.String()) returns t for every set tempo.
func Parse(text string) (Tempo, error) {
	return parseTempo(text, text)
}

// MustParse is like Parse but panics with the returned error.
func MustParse(text string) Tempo {
	return must(Parse(text))
}

// ParseApprox returns the approximate marking in the canonical text of
// Approx.String: unit "≈" number, with the unit and the number of Parse, such
// as "4≈108". The sign is "≈" (U+2248) only: "4=~108", "4~108", and "4=108"
// are rejected. The order of checking is that of Parse.
// ParseApprox(a.String()) returns a for every set marking.
func ParseApprox(text string) (Approx, error) {
	return parseApprox(text, text)
}

// MustParseApprox is like ParseApprox but panics with the returned error.
func MustParseApprox(text string) Approx {
	return must(ParseApprox(text))
}

// ParseRange returns the range in the canonical text of Range.String: unit
// "=" number "-" number, with the unit and the numbers of Parse, such as
// "8=132-144". The order of checking is that of Parse with the numbers from
// left to right, followed by an error wrapping ErrInvalidArgument if the lower
// number is not less than the upper one. ParseRange(r.String()) returns r for
// every set range.
func ParseRange(text string) (Range, error) {
	return parseRange(text, text)
}

// MustParseRange is like ParseRange but panics with the returned error.
func MustParseRange(text string) Range {
	return must(ParseRange(text))
}

// ParseSeconds returns the time in the canonical text of Seconds.String:
// integer "/" count, such as "0/1", "1/3", or "4/2", which gives 2/1. In order
// of checking, it returns an error wrapping ErrInvalidText if the text does
// not match this grammar, including the string of the unset time, and
// ErrOutOfRange together with fraction.ErrOutOfRange if the numerator or then
// the denominator is greater than fraction.MaxComponent.
// ParseSeconds(s.String()) returns s for every set time.
func ParseSeconds(text string) (Seconds, error) {
	num, den, found := strings.Cut(text, "/")
	if !found || !isInteger(num) || !isCount(den) {
		return Seconds{}, parseError(classInvalidText, text)
	}
	n, ok := componentInt(num)
	if !ok {
		return Seconds{}, parseError(classFractionRange, text)
	}
	d, ok := componentInt(den)
	if !ok {
		return Seconds{}, parseError(classFractionRange, text)
	}
	return Seconds{v: fractionOf(new(big.Rat).SetFrac(n, d))}, nil
}

// MustParseSeconds is like ParseSeconds but panics with the returned error.
func MustParseSeconds(text string) Seconds {
	return must(ParseSeconds(text))
}

// ParseIndication returns the indication in the text of Indication.String,
// looking up the name in the list l:
//
//	indication = name / mark / name ", " mark
//	mark       = tempo / approx / range
//	name       = plain / quoted
//
// A plain name contains neither "=" nor "≈" and does not start with a
// quotation mark; a quoted name is a Go interpreted string literal, which is
// also accepted where String writes none, so "\"Allegro\", 4=120" equals
// "Allegro, 4=120". The steps of parsing are: a text that starts with a
// quotation mark begins with a quoted name, followed by nothing or by ", " and
// a marking; otherwise a text with "=" or "≈" is a marking, preceded by a
// plain name and ", " if it contains ", ", where the name ends at the last
// ", "; otherwise the whole text is a name. The name is found with l.Parse,
// ignoring case. A marking with "≈" is approximate, otherwise one with "-" is a
// range, otherwise it is exact. In order of checking, it returns an error
// wrapping ErrInvalidArgument if l is unset; ErrInvalidText if the text is not
// valid UTF-8, does not have this structure, has a name that is not a name of
// l, or has a marking that does not match its grammar; and then the errors of
// Parse, ParseApprox, or ParseRange for the marking. For an indication i whose
// term is from l, or which has no term, ParseIndication(i.String(), l) returns
// i.
func ParseIndication(text string, l term.List) (Indication, error) {
	if !l.IsSet() {
		return Indication{}, invalidError("ParseIndication", "unset list")
	}
	if !utf8.ValidString(text) {
		return Indication{}, parseError(classInvalidText, text)
	}
	var (
		name, markText   string
		hasName, hasMark bool
	)
	switch {
	case strings.HasPrefix(text, `"`):
		q, err := strconv.QuotedPrefix(text)
		if err != nil {
			return Indication{}, parseError(classInvalidText, text)
		}
		name, err = strconv.Unquote(q)
		if err != nil {
			return Indication{}, parseError(classInvalidText, text)
		}
		hasName = true
		rest := text[len(q):]
		switch {
		case rest == "":
		case strings.HasPrefix(rest, nameSep):
			markText, hasMark = rest[len(nameSep):], true
		default:
			return Indication{}, parseError(classInvalidText, text)
		}
	case strings.ContainsAny(text, exactSign+approxSign):
		markText, hasMark = text, true
		if k := strings.LastIndex(text, nameSep); k >= 0 {
			name, hasName = text[:k], true
			if strings.ContainsAny(name, exactSign+approxSign) {
				return Indication{}, parseError(classInvalidText, text)
			}
			markText = text[k+len(nameSep):]
		}
	default:
		name, hasName = text, true
	}
	var ind Indication
	if hasName {
		t, err := l.Parse(name)
		if err != nil {
			return Indication{}, parseError(classInvalidText, text)
		}
		ind.term = t
	}
	if !hasMark {
		return ind, nil
	}
	switch {
	case strings.Contains(markText, approxSign):
		a, err := parseApprox(markText, text)
		if err != nil {
			return Indication{}, err
		}
		return ind.withMark(ApproxMark, a.m.unit, a.m.bpm, fraction.Fraction{}), nil
	case strings.Contains(markText, rangeSign):
		r, err := parseRange(markText, text)
		if err != nil {
			return Indication{}, err
		}
		return ind.withMark(RangeMark, r.unit, r.lower, r.upper), nil
	default:
		t, err := parseTempo(markText, text)
		if err != nil {
			return Indication{}, err
		}
		return ind.withMark(ExactMark, t.m.unit, t.m.bpm, fraction.Fraction{}), nil
	}
}

// MustParseIndication is like ParseIndication but panics with the returned
// error.
func MustParseIndication(text string, l term.List) Indication {
	return must(ParseIndication(text, l))
}
