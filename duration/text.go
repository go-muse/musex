package duration

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"
)

// Strings of unset values. The parse functions reject them.
const (
	unsetDurationString = "Duration(unset)"
	unsetWrittenString  = "Written(unset)"
	unsetTiedString     = "Tied(unset)"
	unsetPositionString = "Position(unset)"
)

// maxComponentDigits is the number of decimal digits of
// fraction.MaxComponent; a count without leading zeros that is longer is
// outside the range.
const maxComponentDigits = len(fraction.MaxComponentString)

// maxIntDigits is the number of decimal digits of math.MaxInt; a tuplet count
// without leading zeros that is longer is greater than math.MaxInt.
const maxIntDigits = len("9223372036854775807")

// appendFraction appends the reduced value of a valid fraction as "n/d".
func appendFraction(b []byte, f fraction.Fraction) []byte {
	num, den, _ := f.ReducedBigPair()
	b = num.Append(b, 10)
	b = append(b, '/')
	return den.Append(b, 10)
}

// String returns the reduced value of d as "n/d", such as "3/8", "1/1", or
// "2/1". For the unset duration it returns "Duration(unset)", which Parse
// rejects.
func (d Duration) String() string {
	if !d.IsSet() {
		return unsetDurationString
	}
	return string(appendFraction(nil, d.v))
}

// MarshalText implements encoding.TextMarshaler with the text of String. It
// returns nil and an error wrapping ErrUnset if d is unset, so an unset struct
// field is not encoded; use the omitzero tag option for an optional field.
func (d Duration) MarshalText() ([]byte, error) {
	if !d.IsSet() {
		return nil, unsetError("Duration.MarshalText", "receiver")
	}
	return appendFraction(nil, d.v), nil
}

// AppendText implements encoding.TextAppender: it appends the text of String
// to dst without modifying dst[:len(dst)]. It returns dst unchanged and an
// error wrapping ErrUnset if d is unset.
func (d Duration) AppendText(dst []byte) ([]byte, error) {
	if !d.IsSet() {
		return dst, unsetError("Duration.AppendText", "receiver")
	}
	return appendFraction(dst, d.v), nil
}

// UnmarshalText implements encoding.TextUnmarshaler with the grammar of Parse.
// On error the receiver is not modified. It returns an error wrapping
// ErrInvalidArgument if d is nil, and otherwise the errors of Parse. The text
// is not retained.
func (d *Duration) UnmarshalText(text []byte) error {
	if d == nil {
		return invalidError("Duration.UnmarshalText", "nil receiver")
	}
	v, err := Parse(string(text))
	if err != nil {
		return err
	}
	*d = v
	return nil
}

// String returns the reduced value of p as "n/d", such as "0/1" or "3/8".
// For the unset position it returns "Position(unset)", which ParsePosition
// rejects.
func (p Position) String() string {
	if !p.IsSet() {
		return unsetPositionString
	}
	return string(appendFraction(nil, p.v))
}

// MarshalText implements encoding.TextMarshaler with the text of String. It
// returns nil and an error wrapping ErrUnset if p is unset, so an unset struct
// field is not encoded; use the omitzero tag option for an optional field.
func (p Position) MarshalText() ([]byte, error) {
	if !p.IsSet() {
		return nil, unsetError("Position.MarshalText", "receiver")
	}
	return appendFraction(nil, p.v), nil
}

// AppendText implements encoding.TextAppender: it appends the text of String
// to dst without modifying dst[:len(dst)]. It returns dst unchanged and an
// error wrapping ErrUnset if p is unset.
func (p Position) AppendText(dst []byte) ([]byte, error) {
	if !p.IsSet() {
		return dst, unsetError("Position.AppendText", "receiver")
	}
	return appendFraction(dst, p.v), nil
}

// UnmarshalText implements encoding.TextUnmarshaler with the grammar of
// ParsePosition. On error the receiver is not modified. It returns an error
// wrapping ErrInvalidArgument if p is nil, and otherwise the errors of
// ParsePosition. The text is not retained.
func (p *Position) UnmarshalText(text []byte) error {
	if p == nil {
		return invalidError("Position.UnmarshalText", "nil receiver")
	}
	v, err := ParsePosition(string(text))
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// appendWritten appends the canonical text of a set written duration.
func appendWritten(b []byte, w Written) []byte {
	b = append(b, w.base.entry(textTable)...)
	for range w.dots {
		b = append(b, '.')
	}
	if w.chain == "" {
		return b
	}
	b = append(b, '[')
	for i, t := range unpackChain(w.chain) {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, t.String()...)
	}
	return append(b, ']')
}

// String returns the canonical text of w: the base, one "." per dot, and the
// chain in brackets if it is not empty, such as "4", "4.", "8[3:2]",
// "4.[5:4,3:2]", or "breve..". For the unset written duration it returns
// "Written(unset)", which ParseWritten rejects.
func (w Written) String() string {
	if !w.IsSet() {
		return unsetWrittenString
	}
	return string(appendWritten(nil, w))
}

// MarshalText implements encoding.TextMarshaler with the text of String. It
// returns nil and an error wrapping ErrUnset if w is unset, so an unset struct
// field is not encoded; use the omitzero tag option for an optional field.
func (w Written) MarshalText() ([]byte, error) {
	if !w.IsSet() {
		return nil, w.unset("Written.MarshalText")
	}
	return appendWritten(nil, w), nil
}

// AppendText implements encoding.TextAppender: it appends the text of String
// to dst without modifying dst[:len(dst)]. It returns dst unchanged and an
// error wrapping ErrUnset if w is unset.
func (w Written) AppendText(dst []byte) ([]byte, error) {
	if !w.IsSet() {
		return dst, w.unset("Written.AppendText")
	}
	return appendWritten(dst, w), nil
}

// UnmarshalText implements encoding.TextUnmarshaler with the grammar of
// ParseWritten. On error the receiver is not modified. It returns an error
// wrapping ErrInvalidArgument if w is nil, and otherwise the errors of
// ParseWritten. The text is not retained.
func (w *Written) UnmarshalText(text []byte) error {
	if w == nil {
		return invalidError("Written.UnmarshalText", "nil receiver")
	}
	v, err := ParseWritten(string(text))
	if err != nil {
		return err
	}
	*w = v
	return nil
}

// appendTied appends the canonical text of a set tied duration.
func appendTied(b []byte, t Tied) []byte {
	for i, w := range unpackParts(t.parts, t.n) {
		if i > 0 {
			b = append(b, '~')
		}
		b = appendWritten(b, w)
	}
	return b
}

// String returns the canonical text of t: the parts as Written.String gives
// them, joined by "~", such as "2~4" or "4[3:2]~8[3:2]". For the unset tied
// duration it returns "Tied(unset)", which ParseTied rejects.
func (t Tied) String() string {
	if !t.IsSet() {
		return unsetTiedString
	}
	return string(appendTied(nil, t))
}

// MarshalText implements encoding.TextMarshaler with the text of String. It
// returns nil and an error wrapping ErrUnset if t is unset, so an unset struct
// field is not encoded; use the omitzero tag option for an optional field.
func (t Tied) MarshalText() ([]byte, error) {
	if !t.IsSet() {
		return nil, t.unset("Tied.MarshalText")
	}
	return appendTied(nil, t), nil
}

// AppendText implements encoding.TextAppender: it appends the text of String
// to dst without modifying dst[:len(dst)]. It returns dst unchanged and an
// error wrapping ErrUnset if t is unset.
func (t Tied) AppendText(dst []byte) ([]byte, error) {
	if !t.IsSet() {
		return dst, t.unset("Tied.AppendText")
	}
	return appendTied(dst, t), nil
}

// UnmarshalText implements encoding.TextUnmarshaler with the grammar of
// ParseTied. On error the receiver is not modified. It returns an error
// wrapping ErrInvalidArgument if t is nil, and otherwise the errors of
// ParseTied. The text is not retained.
func (t *Tied) UnmarshalText(text []byte) error {
	if t == nil {
		return invalidError("Tied.UnmarshalText", "nil receiver")
	}
	v, err := ParseTied(string(text))
	if err != nil {
		return err
	}
	*t = v
	return nil
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

// componentInt returns the count s as a *big.Int and reports whether it is at
// most fraction.MaxComponent. Longer counts are rejected without conversion.
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

// parseRatio parses num "/" den, where den is a count and num is a count or,
// if zeroAllowed, "0", and returns the exact value. Grammar errors precede
// range errors; the numerator is checked before the denominator.
func parseRatio(text string, zeroAllowed bool) (*big.Rat, error) {
	num, den, found := strings.Cut(text, "/")
	numOK := isCount(num) || zeroAllowed && num == "0"
	if !found || !numOK || !isCount(den) {
		return nil, parseError(classInvalidText, text)
	}
	n, ok := componentInt(num)
	if !ok {
		return nil, parseError(classFractionRange, text)
	}
	d, ok := componentInt(den)
	if !ok {
		return nil, parseError(classFractionRange, text)
	}
	return new(big.Rat).SetFrac(n, d), nil
}

// Parse returns the duration written as "n/d" with counts n and d of ASCII
// decimal digits without signs, spaces, or leading zeros, such as "3/8". An
// unreduced pair is accepted: "6/16" gives 3/8. In order of checking, it
// returns an error wrapping ErrInvalidText if text does not match this
// grammar, including the string of the unset duration, and ErrOutOfRange
// together with fraction.ErrOutOfRange if the numerator or then the
// denominator is greater than fraction.MaxComponent. Parse(d.String()) returns
// d for every set duration.
func Parse(text string) (Duration, error) {
	r, err := parseRatio(text, false)
	if err != nil {
		return Duration{}, err
	}
	f, _ := fromRat(r)
	return Duration{v: f}, nil
}

// MustParse is like Parse but panics with the returned error.
func MustParse(text string) Duration {
	return must(Parse(text))
}

// ParsePosition returns the position written as "n/d" with a numerator that
// is "0" or a count and a count denominator, such as "0/1" or "3/8". An
// unreduced pair is accepted: "0/7" gives the origin. In order of checking, it
// returns an error wrapping ErrInvalidText if text does not match this
// grammar, including the string of the unset position, and ErrOutOfRange
// together with fraction.ErrOutOfRange if the numerator or then the
// denominator is greater than fraction.MaxComponent.
// ParsePosition(p.String()) returns p for every set position.
func ParsePosition(text string) (Position, error) {
	r, err := parseRatio(text, true)
	if err != nil {
		return Position{}, err
	}
	f, _ := fromRat(r)
	return Position{v: f}, nil
}

// MustParsePosition is like ParsePosition but panics with the returned error.
func MustParsePosition(text string) Position {
	return must(ParsePosition(text))
}

// writtenText is a written duration that matches the grammar of ParseWritten,
// with its tuplet counts not yet converted.
type writtenText struct {
	base   Base
	dots   int
	counts []string // actual and normal count of each tuplet, in order
}

// scanBase reads the base at the start of s: a word or the longest run of
// digits. It returns the base and the rest of s.
func scanBase(s string) (Base, string, bool) {
	for _, b := range [...]Base{Maxima, Longa, Breve} {
		if word := b.entry(textTable); strings.HasPrefix(s, word) {
			return b, s[len(word):], true
		}
	}
	n := 0
	for n < len(s) && s[n] >= '0' && s[n] <= '9' {
		n++
	}
	for b := Whole; b <= ThousandTwentyFourth; b++ {
		if s[:n] == b.entry(textTable) {
			return b, s[n:], true
		}
	}
	return 0, s, false
}

// scanWritten checks s against the grammar
// written = base *"." [ "[" tuplet *( "," tuplet ) "]" ].
func scanWritten(s string) (writtenText, bool) {
	base, rest, ok := scanBase(s)
	if !ok {
		return writtenText{}, false
	}
	wt := writtenText{base: base}
	for rest != "" && rest[0] == '.' {
		wt.dots++
		rest = rest[1:]
	}
	if rest == "" {
		return wt, true
	}
	if rest[0] != '[' || rest[len(rest)-1] != ']' {
		return writtenText{}, false
	}
	for item := range strings.SplitSeq(rest[1:len(rest)-1], ",") {
		actual, normal, found := strings.Cut(item, ":")
		if !found || !isCount(actual) || !isCount(normal) {
			return writtenText{}, false
		}
		wt.counts = append(wt.counts, actual, normal)
	}
	return wt, true
}

// tupletCount converts a tuplet count and reports whether it is at most
// math.MaxInt.
func tupletCount(s string) (int, bool) {
	if len(s) > maxIntDigits {
		return 0, false
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil || n > math.MaxInt {
		return 0, false
	}
	return int(n), true
}

// build converts a scanned written duration in the order of ParseWritten:
// the tuplet counts from left to right, the chain coefficient, and the
// duration. On failure it returns the class of the error.
func (wt writtenText) build() (Written, textClass, bool) {
	chain := make([]tuplet.Tuplet, len(wt.counts)/2)
	for i := range chain {
		actual, ok := tupletCount(wt.counts[2*i])
		if !ok {
			return Written{}, classTupletCount, false
		}
		normal, ok := tupletCount(wt.counts[2*i+1])
		if !ok {
			return Written{}, classTupletCount, false
		}
		chain[i] = tuplet.MustNew(actual, normal)
	}
	coef, err := chainCoefficient("parse", chain)
	if err != nil {
		return Written{}, classChainRange, false
	}
	v, err := writtenValue("parse", wt.base, wt.dots, coef)
	if err != nil {
		return Written{}, classFractionRange, false
	}
	return Written{base: wt.base, dots: wt.dots, chain: packChain(chain), value: v}, 0, true
}

// ParseWritten returns the written duration in the canonical text of
// Written.String:
//
//	written  = base *"." [ "[" tuplet *( "," tuplet ) "]" ]
//	base     = "maxima" / "longa" / "breve" / numeral
//	numeral  = "1" / "2" / "4" / "8" / "16" / "32" / "64" / "128" / "256" / "512" / "1024"
//	tuplet   = count ":" count
//	count    = nzdigit *digit
//
// The digits of a base are read as the longest run, so "16." is a dotted
// sixteenth, and "3" is not a base. Spaces, signs, leading zeros, other case,
// and names are not accepted. In order of checking, it returns an error
// wrapping ErrInvalidText if text does not match the grammar, including the
// string of the unset written duration; ErrOutOfRange together with
// tuplet.ErrOutOfRange if a tuplet count is greater than math.MaxInt;
// ErrOutOfRange together with tuplet.ErrOutOfRange and fraction.ErrOutOfRange
// if the chain coefficient is outside the range of package common/fraction;
// and ErrOutOfRange together with fraction.ErrOutOfRange if the duration,
// including through the number of dots, is outside that range.
// ParseWritten(w.String()) returns w for every set written duration.
func ParseWritten(text string) (Written, error) {
	wt, ok := scanWritten(text)
	if !ok {
		return Written{}, parseError(classInvalidText, text)
	}
	w, class, ok := wt.build()
	if !ok {
		return Written{}, parseError(class, text)
	}
	return w, nil
}

// MustParseWritten is like ParseWritten but panics with the returned error.
func MustParseWritten(text string) Written {
	return must(ParseWritten(text))
}

// ParseTied returns the tied duration in the canonical text of Tied.String:
// two or more written durations in the grammar of ParseWritten joined by "~",
// such as "2~4". A single written duration, such as "4", is rejected. In
// order of checking, it returns an error wrapping ErrInvalidText if text does
// not match the grammar, including the string of the unset tied duration;
// then the errors of ParseWritten for each part from left to right; and
// ErrOutOfRange together with fraction.ErrOutOfRange if the sum of the
// durations is outside the range of package common/fraction.
// ParseTied(t.String()) returns t for every set tied duration.
func ParseTied(text string) (Tied, error) {
	var scanned []writtenText
	for item := range strings.SplitSeq(text, "~") {
		wt, ok := scanWritten(item)
		if !ok {
			return Tied{}, parseError(classInvalidText, text)
		}
		scanned = append(scanned, wt)
	}
	if len(scanned) < 2 {
		return Tied{}, parseError(classInvalidText, text)
	}
	parts := make([]Written, len(scanned))
	for i, wt := range scanned {
		w, class, ok := wt.build()
		if !ok {
			return Tied{}, parseError(class, text)
		}
		parts[i] = w
	}
	t, err := makeTied("ParseTied", parts)
	if err != nil {
		return Tied{}, parseError(classFractionRange, text)
	}
	return t, nil
}

// MustParseTied is like ParseTied but panics with the returned error.
func MustParseTied(text string) Tied {
	return must(ParseTied(text))
}
