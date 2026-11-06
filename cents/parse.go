package cents

import (
	"fmt"
	"math/big"
	"strings"
)

// Parse reads a value from text in one of the exact forms, strictly:
//
//	text      = cents | fraction | ratio | power
//	cents     = [ sign ] digits [ "." digits ] [ unitpart ]   ; KindDecimal
//	fraction  = [ sign ] digits "/" digits unitpart           ; KindFraction
//	ratio     = digits ":" digits                             ; KindRatio
//	power     = factor { "*" factor }                         ; KindPower
//	factor    = digits [ "^" exponent ]
//	exponent  = [ "-" ] digits | "(" [ "-" ] digits "/" digits ")"
//	digits    = digit { digit }                               ; "0"..."9"
//	sign      = "+" | "-" | "−" (U+2212)
//	unitpart  = [ " " | U+00A0 ] unit                          ; at most one separator
//	unit      = "¢" | "cent" | "cents"                         ; cent and cents in any case
//
// The form is chosen by a marker: a text with ":" is a ratio; otherwise a
// text with "^" or "*" is a product of powers; otherwise a text with "/" is a
// fraction; otherwise it is decimal cents. The whole text must then match the
// grammar of that form. A bare number such as "5" is 5 cents. A ratio a:b is
// 1200·log2(a/b) cents; a product of powers is 1200·log2 of the product, and
// its bases are any integers of at least 2, split into primes, so
// "10^(1/1000)" is [Savart]. Non-canonical texts such as "6:4",
// "2400/38 ¢", "0007", "-0", and "30 CENTS" are accepted. Parse reads back
// [Cents.String], [Cents.Text], and every form of [Cents.TextAs] to an equal
// value. The text is not retained.
//
// The errors are checked in this order. [ErrInvalidText]: the text does not
// match the grammar, including an empty text, spaces around the parts of a
// form, a zero term of a ratio, a zero denominator of a fraction or of an
// exponent, and a base 0 or 1. A bare fraction of digits such as "3/2" is
// invalid, and the message of the error contains the two valid spellings
// "3:2" and "3/2 ¢" (with the letters a:b and a/b ¢ for numbers longer than
// MaxDigits digits); this hint is the only part of the error text that is
// part of the API. [ErrTooComplex]: a run of more than MaxDigits digits, or a
// product of more than MaxFactors factors; these limits are checked before
// any arithmetic. [ErrPrimeLimit]: a prime greater than 31 in a ratio or in a
// base. [ErrTooComplex]: the record of the result is beyond the record
// limits; exponents are added before the check, so "3^(1/p)*3^(-1/p)" is 0.
// [ErrOutOfRange]: the value is outside the range.
func Parse(text string) (Cents, error) {
	rc, err := parseText(text)
	if err != nil {
		return Cents{}, err
	}
	return fromRecord(rc), nil
}

// MustParse is like [Parse] but panics with the error. It is meant for
// literals and values known to be valid.
func MustParse(text string) Cents {
	return must(Parse(text))
}

// parseText parses text in the general mode.
func parseText(text string) (*record, error) {
	switch {
	case text == "":
		return nil, textError(text, "empty text")
	case strings.IndexByte(text, ':') >= 0:
		return parseRatio(text)
	case strings.IndexByte(text, '^') >= 0 || strings.IndexByte(text, '*') >= 0:
		return parsePower(text)
	case strings.IndexByte(text, '/') >= 0:
		return parseFraction(text)
	}
	return parseDecimal(text)
}

// digitsEnd returns the end of the run of ASCII digits of s starting at i.
func digitsEnd(s string, i int) int {
	for i < len(s) && '0' <= s[i] && s[i] <= '9' {
		i++
	}
	return i
}

// isZeroDigits reports whether the digits are all zeros.
func isZeroDigits(digits string) bool {
	return strings.TrimLeft(digits, "0") == ""
}

// signAt reads an optional sign of cents at the start of s.
func signAt(s string) (negative bool, i int) {
	switch {
	case strings.HasPrefix(s, "+"):
		return false, 1
	case strings.HasPrefix(s, "-"):
		return true, 1
	case strings.HasPrefix(s, typoMinus):
		return true, len(typoMinus)
	}
	return false, 0
}

// isUnitPart reports whether s is a unit with at most one separator before
// it.
func isUnitPart(s string) bool {
	if strings.HasPrefix(s, " ") {
		s = s[1:]
	} else if strings.HasPrefix(s, nbsp) {
		s = s[len(nbsp):]
	}
	return s == centSign || foldsTo(s, "cent") || foldsTo(s, "cents")
}

// foldsTo reports whether s equals the lowercase ASCII word in any case of
// ASCII letters.
func foldsTo(s, word string) bool {
	if len(s) != len(word) {
		return false
	}
	for i := range len(s) {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != word[i] {
			return false
		}
	}
	return true
}

// checkDigits returns ErrTooComplex if a run of digits is longer than
// MaxDigits.
func checkDigits(text string, runs ...string) error {
	for _, d := range runs {
		if len(d) > MaxDigits {
			return errorf(ErrTooComplex, "%s: number longer than %d digits", quoted(text), MaxDigits)
		}
	}
	return nil
}

// atoi returns the value of a run of ASCII digits.
func atoi(digits string) *big.Int {
	n, _ := new(big.Int).SetString(digits, 10)
	return n
}

// parseDecimal parses decimal cents.
func parseDecimal(text string) (*record, error) {
	negative, i := signAt(text)
	j := digitsEnd(text, i)
	if j == i {
		return nil, textError(text, "expected digits")
	}
	intPart, fracPart := text[i:j], ""
	i = j
	if i < len(text) && text[i] == '.' {
		i++
		j = digitsEnd(text, i)
		if j == i {
			return nil, textError(text, "expected digits after the point")
		}
		fracPart = text[i:j]
		i = j
	}
	if i < len(text) && !isUnitPart(text[i:]) {
		return nil, textError(text, "unexpected characters after the number")
	}
	if err := checkDigits(text, intPart, fracPart); err != nil {
		return nil, err
	}
	rc := new(record)
	rc.r.SetFrac(atoi(intPart+fracPart), pow10(len(fracPart)))
	if negative {
		rc.r.Neg(&rc.r)
	}
	return rc, rc.check("Parse")
}

// parseFraction parses a fraction of cents with a unit.
func parseFraction(text string) (*record, error) {
	if a, b, ok := bareFraction(text); ok {
		if len(a) > MaxDigits || len(b) > MaxDigits {
			a, b = "a", "b"
		}
		return nil, textError(text, fmt.Sprintf(
			"a fraction of cents needs a unit: write %s:%s for a frequency ratio or %s/%s ¢ for cents", a, b, a, b))
	}
	negative, i := signAt(text)
	j := digitsEnd(text, i)
	if j == i {
		return nil, textError(text, "expected digits")
	}
	num := text[i:j]
	i = j
	if i == len(text) || text[i] != '/' {
		return nil, textError(text, "expected /")
	}
	i++
	j = digitsEnd(text, i)
	if j == i {
		return nil, textError(text, "expected digits after /")
	}
	den := text[i:j]
	i = j
	if i == len(text) || !isUnitPart(text[i:]) {
		return nil, textError(text, "a fraction of cents needs a unit")
	}
	if isZeroDigits(den) {
		return nil, textError(text, "zero denominator")
	}
	if err := checkDigits(text, num, den); err != nil {
		return nil, err
	}
	rc := new(record)
	rc.r.SetFrac(atoi(num), atoi(den))
	if negative {
		rc.r.Neg(&rc.r)
	}
	return rc, rc.check("Parse")
}

// bareFraction returns a and b if text is exactly digits "/" digits.
func bareFraction(text string) (a, b string, ok bool) {
	i := digitsEnd(text, 0)
	if i == 0 || i == len(text) || text[i] != '/' {
		return "", "", false
	}
	j := digitsEnd(text, i+1)
	if j == i+1 || j != len(text) {
		return "", "", false
	}
	return text[:i], text[i+1:], true
}

// parseRatio parses a frequency ratio.
func parseRatio(text string) (*record, error) {
	i := digitsEnd(text, 0)
	if i == 0 || i == len(text) || text[i] != ':' {
		return nil, textError(text, "expected a ratio of digits a:b")
	}
	j := digitsEnd(text, i+1)
	if j == i+1 || j != len(text) {
		return nil, textError(text, "expected a ratio of digits a:b")
	}
	a, b := text[:i], text[i+1:]
	if isZeroDigits(a) || isZeroDigits(b) {
		return nil, textError(text, "zero term of a ratio")
	}
	if err := checkDigits(text, a, b); err != nil {
		return nil, err
	}
	rc, err := ratioRecord(atoi(a), atoi(b), "Parse")
	if err != nil {
		return nil, err
	}
	return rc, rc.check("Parse")
}

// powerFactor is a factor of a product of powers in a text.
type powerFactor struct {
	base, num, den string // den is empty for an integer exponent
	negative       bool
}

// parsePower parses a product of powers.
func parsePower(text string) (*record, error) {
	var factors [MaxFactors]powerFactor
	nFactors := 0
	for i := 0; ; i++ {
		j := digitsEnd(text, i)
		if j == i {
			return nil, textError(text, "expected a base")
		}
		f := powerFactor{base: text[i:j], num: "1"}
		if isZeroDigits(f.base) || strings.TrimLeft(f.base, "0") == "1" {
			return nil, textError(text, "base 0 or 1")
		}
		i = j
		if i < len(text) && text[i] == '^' {
			var err error
			if i, err = parseExponent(text, i+1, &f); err != nil {
				return nil, err
			}
		}
		if nFactors < MaxFactors {
			factors[nFactors] = f
		}
		nFactors++
		if i == len(text) {
			break
		}
		if text[i] != '*' {
			return nil, textError(text, "expected * between factors")
		}
	}
	if nFactors > MaxFactors {
		return nil, errorf(ErrTooComplex, "%s: more than %d factors", quoted(text), MaxFactors)
	}
	fs := factors[:nFactors]
	for _, f := range fs {
		if err := checkDigits(text, f.base, f.num, f.den); err != nil {
			return nil, err
		}
	}
	var exps [MaxFactors][numPrimes + 1]int64
	for k, f := range fs {
		n := atoi(f.base)
		exps[k] = factor(n)
		if n.Cmp(bigOne) != 0 {
			return nil, errorf(ErrPrimeLimit, "%s: base %s has a prime factor greater than %d", quoted(text), f.base, MaxPrime)
		}
	}
	rc := new(record)
	var q, t big.Rat
	for k, f := range fs {
		den := bigOne
		if f.den != "" {
			den = atoi(f.den)
		}
		q.SetFrac(atoi(f.num), den)
		if f.negative {
			q.Neg(&q)
		}
		for i, a := range exps[k] {
			if a == 0 {
				continue
			}
			t.Mul(&q, new(big.Rat).SetInt64(a))
			if i == 0 {
				rc.r.Add(&rc.r, t.Mul(&t, rat1200))
			} else {
				rc.e[i-1].Add(&rc.e[i-1], &t)
			}
		}
	}
	return rc, rc.check("Parse")
}

// parseExponent parses the exponent of f that starts at i and returns the
// position after it.
func parseExponent(text string, i int, f *powerFactor) (int, error) {
	paren := i < len(text) && text[i] == '('
	if paren {
		i++
	}
	if i < len(text) && text[i] == '-' {
		f.negative = true
		i++
	}
	j := digitsEnd(text, i)
	if j == i {
		return 0, textError(text, "expected an exponent")
	}
	f.num = text[i:j]
	i = j
	if !paren {
		return i, nil
	}
	if i == len(text) || text[i] != '/' {
		return 0, textError(text, "expected / in the exponent")
	}
	i++
	j = digitsEnd(text, i)
	if j == i {
		return 0, textError(text, "expected digits after / in the exponent")
	}
	f.den = text[i:j]
	i = j
	if i == len(text) || text[i] != ')' {
		return 0, textError(text, "expected ) after the exponent")
	}
	if isZeroDigits(f.den) {
		return 0, textError(text, "zero denominator of an exponent")
	}
	return i + 1, nil
}
