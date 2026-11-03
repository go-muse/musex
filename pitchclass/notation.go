package pitchclass

import (
	"fmt"
	"strconv"
	"strings"
)

// modSeparator separates the number of a class or an operator from N in the
// canonical forms.
const modSeparator = " mod "

// letterStyleLimit is the largest number of positions for which the letter
// styles are defined: they have symbols only for 10 and 11.
const letterStyleLimit = 12

// Style is a style of the literary notation of class numbers, used by
// [Class.Notation] and [ParseNotation]. The styles differ only in how they
// write the numbers 10 and 11:
//
//	Style    10  11  Source               Defined for
//	Decimal  10  11  integer notation     any N
//	TE       T   E   MT21, §33.3          N <= 12
//	LowerTE  t   e   Whittall; OMT        N <= 12
//	AB       A   B   Morris               N <= 12
//
// The numbers 0 to 9 are single digits in every style. For more than 12
// positions, the literature has no one-character symbols, so only Decimal
// applies.
//
// The valid styles are exactly the four constants. The zero value is
// [Decimal], which is defined for any N. Any other value, such as Style(9),
// can only be made by an explicit conversion; [Class.Notation] and
// [ParseNotation] return an error that wraps [ErrNotation] for it. New
// styles, if any, will be added after the existing constants, but the
// numeric values of the constants are not a recommended storage format.
type Style uint8

// The literary notation styles.
const (
	// Decimal writes every number in decimal digits: 10 is "10" and 11 is
	// "11". It is the zero value and is defined for any N.
	Decimal Style = iota
	// TE writes 10 as "T" and 11 as "E", as in MT21 (§33.3).
	TE
	// LowerTE writes 10 as "t" and 11 as "e", as in Whittall and Open Music
	// Theory (OMT).
	LowerTE
	// AB writes 10 as "A" and 11 as "B", as in Morris.
	AB
)

// String returns the name of the style constant: "Decimal", "TE",
// "LowerTE", or "AB". For any other value, it returns "Style(n)" with the
// decimal value n, such as "Style(9)". It is meant for output and
// debugging; it is not a literary notation.
func (s Style) String() string {
	switch s {
	case Decimal:
		return "Decimal"
	case TE:
		return "TE"
	case LowerTE:
		return "LowerTE"
	case AB:
		return "AB"
	default:
		return "Style(" + strconv.Itoa(int(s)) + ")"
	}
}

// letters returns the symbols of a letter style for 10 and 11 and reports
// whether s is a letter style.
func (s Style) letters() (ten, eleven byte, ok bool) {
	switch s {
	case TE:
		return 'T', 'E', true
	case LowerTE:
		return 't', 'e', true
	case AB:
		return 'A', 'B', true
	default:
		return 0, 0, false
	}
}

// checkStyle returns an error that wraps [ErrNotation] if s is not a valid
// style or is a letter style and n > letterStyleLimit, and nil otherwise.
func checkStyle(s Style, n int) error {
	if s == Decimal {
		return nil
	}

	if _, _, ok := s.letters(); !ok {
		return fmt.Errorf("%w: invalid style %v", ErrNotation, s)
	}

	if n > letterStyleLimit {
		return fmt.Errorf("%w: style %v is not defined for %d positions", ErrNotation, s, n)
	}

	return nil
}

// Notation returns the number of the class, without N, in the literary
// notation style:
//
//   - [Decimal] writes the number in decimal digits without a sign or
//     leading zeros, for any N;
//   - [TE], [LowerTE], and [AB], defined for N <= 12, write the numbers 0
//     to 9 as single digits and 10 and 11 as the letters of the style. For
//     N <= 10 they agree with Decimal.
//
// For example, 10 mod 12 is "10" in Decimal, "T" in TE, "t" in LowerTE,
// and "A" in AB. [ParseNotation] reverses Notation.
//
// If c is the zero value, Notation returns "" and an error that wraps
// [ErrInvalidValue]. Otherwise, if style is not a valid style, or is a
// letter style and N > 12, it returns "" and an error that wraps
// [ErrNotation]. For a valid class and Decimal, there is no error.
func (c Class) Notation(style Style) (string, error) {
	if !c.IsValid() {
		return "", errValue()
	}

	if err := checkStyle(style, c.n); err != nil {
		return "", err
	}

	ten, eleven, ok := style.letters()

	switch {
	case ok && c.k == 10:
		return string(ten), nil
	case ok && c.k == 11:
		return string(eleven), nil
	default:
		return strconv.Itoa(c.k), nil
	}
}

// MustNotation is like [Class.Notation] but panics with its error instead
// of returning it. It is intended for values known to be valid.
func (c Class) MustNotation(style Style) string {
	s, err := c.Notation(style)
	if err != nil {
		panic(err)
	}

	return s
}

// Parse returns the class written in the canonical form of [Class.String]:
// "<k> mod <N>", such as "1 mod 12". Parse(c.String()) equals c for every
// class c.
//
// The form is exact and case-sensitive: k and N are decimal ASCII digits
// without a sign or leading zeros (0 is "0"), with 1 <= N <= math.MaxInt and
// 0 <= k < N, separated by " mod " with single spaces and no other spaces.
// For any other text, including "", "01 mod 12", "12 mod 12", "0 mod 0",
// and "Class(invalid)", Parse returns the zero Class and an error that
// wraps [ErrNotation].
func Parse(s string) (Class, error) {
	k, n, ok := parseClass(s)
	if !ok {
		return Class{}, errNotation(s)
	}

	return Class{k: k, n: n}, nil
}

// MustParse is like [Parse] but panics with its error instead of returning
// it. It is intended for literals and values known to be valid.
func MustParse(s string) Class {
	c, err := Parse(s)
	if err != nil {
		panic(err)
	}

	return c
}

// ParseNotation returns the class of modulus positions whose number is
// written in the literary notation style, as produced by [Class.Notation]:
// ParseNotation(c.MustNotation(style), style, c.Modulus()) equals c. The
// text is read strictly in the given style, without guessing the style or
// the case:
//
//   - in [Decimal], the number is decimal ASCII digits without a sign or
//     leading zeros, such as "11";
//   - in [TE], [LowerTE], and [AB], the number is a single digit "0" to "9"
//     or the letter of the style for 10 or 11, such as "T" in TE; "10" and
//     "t" are not accepted in TE.
//
// The number must be less than modulus.
//
// If modulus is less than 1, ParseNotation returns the zero Class and an
// error that wraps [ErrInvalidModulus]. Otherwise, if style is not a valid
// style, or is a letter style and modulus > 12, or the text is not a number
// below modulus in the style, it returns the zero Class and an error that
// wraps [ErrNotation].
func ParseNotation(s string, style Style, modulus int) (Class, error) {
	if modulus < 1 {
		return Class{}, errModulus(modulus)
	}

	if err := checkStyle(style, modulus); err != nil {
		return Class{}, err
	}

	k, ok := parseStyled(s, style)
	if !ok || k >= modulus {
		return Class{}, errNotation(s)
	}

	return Class{k: k, n: modulus}, nil
}

// MustParseNotation is like [ParseNotation] but panics with its error
// instead of returning it. It is intended for literals and values known to
// be valid.
func MustParseNotation(s string, style Style, modulus int) Class {
	c, err := ParseNotation(s, style, modulus)
	if err != nil {
		panic(err)
	}

	return c
}

// ParseOperator returns the operator written in the canonical form of
// [Operator.String]: "T<n>M<m> mod <N>", such as "T5M7 mod 12".
// ParseOperator(f.String()) equals f for every operator f.
//
// The numbers follow the rules of [Parse]: decimal ASCII digits without a
// sign or leading zeros, with 1 <= N <= math.MaxInt and n, m < N.
//
// For any other text, ParseOperator returns the zero Operator and an error
// that wraps [ErrNotation]. If the text has the form but m is not
// invertible modulo N, as in "T0M2 mod 12", it returns the zero Operator
// and an error that wraps [ErrNotInvertible].
func ParseOperator(s string) (Operator, error) {
	rest, ok := strings.CutPrefix(s, "T")
	if !ok {
		return Operator{}, errNotation(s)
	}

	pair, modulus, ok := strings.Cut(rest, modSeparator)
	if !ok {
		return Operator{}, errNotation(s)
	}

	index, multiplier, ok := strings.Cut(pair, "M")
	if !ok {
		return Operator{}, errNotation(s)
	}

	n, okN := parseModulus(modulus)
	t, okT := parseNumber(index)
	m, okM := parseNumber(multiplier)

	if !okN || !okT || !okM || t >= n || m >= n {
		return Operator{}, errNotation(s)
	}

	if !invertible(m, n) {
		return Operator{}, errNotInvertible(m, n)
	}

	return Operator{n: n, m: m, t: t}, nil
}

// MustParseOperator is like [ParseOperator] but panics with its error
// instead of returning it. It is intended for literals and values known to
// be valid.
func MustParseOperator(s string) Operator {
	f, err := ParseOperator(s)
	if err != nil {
		panic(err)
	}

	return f
}

// parseClass parses the canonical form "<k> mod <N>" and reports whether s
// is the form of a class.
func parseClass(s string) (k, n int, ok bool) {
	number, modulus, ok := strings.Cut(s, modSeparator)
	if !ok {
		return 0, 0, false
	}

	n, okN := parseModulus(modulus)
	k, okK := parseNumber(number)

	if !okN || !okK || k >= n {
		return 0, 0, false
	}

	return k, n, true
}

// parseModulus parses N of a canonical form: a number of parseNumber that is
// at least 1.
func parseModulus(s string) (int, bool) {
	n, ok := parseNumber(s)
	if !ok || n < 1 {
		return 0, false
	}

	return n, true
}

// parseNumber parses a non-negative decimal int written in ASCII digits
// without a sign or leading zeros ("0" for zero) and reports whether s has
// this form and fits in an int.
func parseNumber(s string) (int, bool) {
	if s == "" || (s[0] == '0' && len(s) > 1) {
		return 0, false
	}

	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
	}

	v, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return 0, false
	}

	return int(v), true
}

// parseStyled parses a class number written in the valid style s and
// reports whether the text has the form of the style.
func parseStyled(text string, s Style) (int, bool) {
	ten, eleven, ok := s.letters()
	if !ok {
		return parseNumber(text)
	}

	if len(text) != 1 {
		return 0, false
	}

	switch ch := text[0]; {
	case ch >= '0' && ch <= '9':
		return int(ch - '0'), true
	case ch == ten:
		return 10, true
	case ch == eleven:
		return 11, true
	default:
		return 0, false
	}
}

// errNotation returns the error for a text that is not a valid notation.
func errNotation(s string) error {
	return fmt.Errorf("%w: %q", ErrNotation, s)
}
