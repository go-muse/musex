package pcset

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-muse/musex/pitchclass"
)

// String returns the canonical string representation "{c_0, c_1, ..., c_{k-1}} mod N"
// with class numbers in strictly ascending order, or "Set(invalid)" if s is Set{}.
func (s Set) String() string {
	if !s.IsValid() {
		return "Set(invalid)"
	}

	var b strings.Builder
	b.WriteByte('{')

	first := true
	s.forEachNumber(func(x int) bool {
		if !first {
			b.WriteString(", ")
		}

		first = false
		b.WriteString(strconv.Itoa(x))

		return true
	})

	b.WriteString("} mod ")
	b.WriteString(strconv.Itoa(s.n))

	return b.String()
}

// Parse parses a canonical pitch-class set representation "{c_0, c_1, ..., c_{k-1}} mod N"
// where 1 <= N <= [MaxModulus] and 0 <= c_0 < c_1 < ... < c_{k-1} < N are
// strictly ascending decimal integers without sign or leading zeros.
//
// Parse checks errors in the following order:
//  1. invalid syntax, non-decimal tokens, leading zeros, or modulus < 1: [ErrNotation]
//  2. modulus > [MaxModulus]: [ErrModulusTooLarge]
//  3. element >= modulus, duplicate element, or non-ascending order: [ErrNotation]
func Parse(s string) (Set, error) {
	if !strings.HasPrefix(s, "{") {
		return Set{}, errNotation(s)
	}

	inner, modStr, ok := strings.Cut(s[1:], "} mod ")
	if !ok {
		return Set{}, errNotation(s)
	}

	modulus, ok := parseStrictPositiveInt(modStr)
	if !ok {
		return Set{}, errNotation(s)
	}

	var nums []int
	if inner != "" {
		parts := strings.Split(inner, ", ")
		nums = make([]int, len(parts))

		for i, part := range parts {
			val, valid := parseStrictNonNegativeInt(part)
			if !valid {
				return Set{}, errNotation(s)
			}

			nums[i] = val
		}
	}

	if modulus > MaxModulus {
		return Set{}, fmt.Errorf("%w: %q", errModulusTooLarge(modulus), s)
	}

	prev := -1
	for _, x := range nums {
		if x <= prev || x >= modulus {
			return Set{}, errNotation(s)
		}

		prev = x
	}

	return MustNew(modulus, nums...), nil
}

// MustParse is like [Parse] but panics on error.
func MustParse(s string) Set {
	res, err := Parse(s)
	if err != nil {
		panic(err)
	}

	return res
}

// Notation returns the delimited literary representation "{elem_0, elem_1, ..., elem_{k-1}}"
// of s in the given pitch-class style, ordered by ascending class number.
//
// Notation checks errors in the following order:
//  1. !s.IsValid(): [ErrInvalidValue]
//  2. invalid style or letter style ([pitchclass.TE], [pitchclass.LowerTE], [pitchclass.AB]) with Modulus() > 12: [ErrNotation]
func (s Set) Notation(style pitchclass.Style) (string, error) {
	if !s.IsValid() {
		return "", errInvalidValue()
	}

	if err := checkLiteraryStyle(style, s.n, false); err != nil {
		return "", err
	}

	return formatNumbersLiterary(s.Numbers(), s.n, style, '{', '}', ", "), nil
}

// MustNotation is like [Set.Notation] but panics on error.
func (s Set) MustNotation(style pitchclass.Style) string {
	str, err := s.Notation(style)
	if err != nil {
		panic(err)
	}

	return str
}

// CompactNotation returns the compact literary representation "{elem_0elem_1...elem_{k-1}}"
// of s without separators in the given pitch-class style, ordered by ascending
// class number.
//
// CompactNotation checks errors in the following order:
//  1. !s.IsValid(): [ErrInvalidValue]
//  2. invalid style, [pitchclass.Decimal] with Modulus() > 10, or letter style with Modulus() > 12: [ErrNotation]
func (s Set) CompactNotation(style pitchclass.Style) (string, error) {
	if !s.IsValid() {
		return "", errInvalidValue()
	}

	if err := checkLiteraryStyle(style, s.n, true); err != nil {
		return "", err
	}

	return formatNumbersLiterary(s.Numbers(), s.n, style, '{', '}', ""), nil
}

// MustCompactNotation is like [Set.CompactNotation] but panics on error.
func (s Set) MustCompactNotation(style pitchclass.Style) string {
	str, err := s.CompactNotation(style)
	if err != nil {
		panic(err)
	}

	return str
}

// NormalFormNotation returns the delimited literary representation
// "[elem_0, elem_1, ..., elem_{k-1}]" of the normal form of s under alg in the
// given pitch-class style.
//
// NormalFormNotation checks errors in the following order:
//  1. !s.IsValid(): [ErrInvalidValue]
//  2. !alg.IsValid(): [ErrInvalidAlgorithm]
//  3. invalid style or letter style with Modulus() > 12: [ErrNotation]
func (s Set) NormalFormNotation(style pitchclass.Style, alg Algorithm) (string, error) {
	if !s.IsValid() {
		return "", errInvalidValue()
	}

	if !alg.IsValid() {
		return "", errInvalidAlgorithm(alg)
	}

	if err := checkLiteraryStyle(style, s.n, false); err != nil {
		return "", err
	}

	return formatNumbersLiterary(s.normalFormNumbersUnchecked(alg), s.n, style, '[', ']', ", "), nil
}

// MustNormalFormNotation is like [Set.NormalFormNotation] but panics on error.
func (s Set) MustNormalFormNotation(style pitchclass.Style, alg Algorithm) string {
	str, err := s.NormalFormNotation(style, alg)
	if err != nil {
		panic(err)
	}

	return str
}

// NormalFormCompactNotation returns the compact literary representation
// "[elem_0elem_1...elem_{k-1}]" without separators of the normal form of s
// under alg in the given pitch-class style.
//
// NormalFormCompactNotation checks errors in the following order:
//  1. !s.IsValid(): [ErrInvalidValue]
//  2. !alg.IsValid(): [ErrInvalidAlgorithm]
//  3. invalid style, [pitchclass.Decimal] with Modulus() > 10, or letter style with Modulus() > 12: [ErrNotation]
func (s Set) NormalFormCompactNotation(style pitchclass.Style, alg Algorithm) (string, error) {
	if !s.IsValid() {
		return "", errInvalidValue()
	}

	if !alg.IsValid() {
		return "", errInvalidAlgorithm(alg)
	}

	if err := checkLiteraryStyle(style, s.n, true); err != nil {
		return "", err
	}

	return formatNumbersLiterary(s.normalFormNumbersUnchecked(alg), s.n, style, '[', ']', ""), nil
}

// MustNormalFormCompactNotation is like [Set.NormalFormCompactNotation] but
// panics on error.
func (s Set) MustNormalFormCompactNotation(style pitchclass.Style, alg Algorithm) string {
	str, err := s.NormalFormCompactNotation(style, alg)
	if err != nil {
		panic(err)
	}

	return str
}

// ParseNotation parses a literary pitch-class set or normal-form string
// ("{...}" or "[...]", delimited by ", " or compact) in the given style and
// modulus. Element order and duplicates in the input are arbitrary.
//
// ParseNotation checks errors in the following order:
//  1. modulus < 1: [ErrInvalidModulus]
//  2. modulus > [MaxModulus]: [ErrModulusTooLarge]
//  3. invalid style or letter style with modulus > 12: [ErrNotation]
//  4. missing/mismatched brackets or invalid element tokens: [ErrNotation]
func ParseNotation(s string, style pitchclass.Style, modulus int) (Set, error) {
	if err := checkModulus(modulus); err != nil {
		return Set{}, err
	}

	if err := checkLiteraryStyle(style, modulus, false); err != nil {
		return Set{}, err
	}

	if len(s) < 2 {
		return Set{}, errNotation(s)
	}

	openCh, closeCh := s[0], s[len(s)-1]
	if (openCh != '{' || closeCh != '}') && (openCh != '[' || closeCh != ']') {
		return Set{}, errNotation(s)
	}

	inner := s[1 : len(s)-1]
	if inner == "" {
		return Set{n: modulus}, nil
	}

	if strings.Contains(inner, ", ") {
		parts := strings.Split(inner, ", ")
		nums := make([]int, len(parts))

		for i, tok := range parts {
			if tok == "" || strings.ContainsAny(tok, ", ") {
				return Set{}, errNotation(s)
			}

			c, err := pitchclass.ParseNotation(tok, style, modulus)
			if err != nil {
				return Set{}, errNotation(s)
			}

			nums[i] = c.Number()
		}

		return MustNew(modulus, nums...), nil
	}

	if strings.ContainsAny(inner, ", ") {
		return Set{}, errNotation(s)
	}

	if style == pitchclass.Decimal && modulus > 10 {
		c, err := pitchclass.ParseNotation(inner, pitchclass.Decimal, modulus)
		if err != nil {
			return Set{}, errNotation(s)
		}

		return MustNew(modulus, c.Number()), nil
	}

	nums := make([]int, len(inner))
	for i := range len(inner) {
		c, err := pitchclass.ParseNotation(inner[i:i+1], style, modulus)
		if err != nil {
			return Set{}, errNotation(s)
		}

		nums[i] = c.Number()
	}

	return MustNew(modulus, nums...), nil
}

// MustParseNotation is like [ParseNotation] but panics on error.
func MustParseNotation(s string, style pitchclass.Style, modulus int) Set {
	res, err := ParseNotation(s, style, modulus)
	if err != nil {
		panic(err)
	}

	return res
}

func checkLiteraryStyle(style pitchclass.Style, modulus int, compact bool) error {
	if style > pitchclass.AB {
		return errNotation(style.String())
	}

	if style != pitchclass.Decimal && modulus > 12 {
		return errNotation(style.String())
	}

	if compact && style == pitchclass.Decimal && modulus > 10 {
		return errNotation(style.String())
	}

	return nil
}

func formatNumbersLiterary(nums []int, modulus int, style pitchclass.Style, openCh, closeCh byte, sep string) string {
	var b strings.Builder
	b.WriteByte(openCh)

	for i, x := range nums {
		if i > 0 && sep != "" {
			b.WriteString(sep)
		}

		b.WriteString(pitchclass.MustNew(x, modulus).MustNotation(style))
	}

	b.WriteByte(closeCh)

	return b.String()
}

func parseStrictPositiveInt(s string) (int, bool) {
	if len(s) == 0 || s[0] < '1' || s[0] > '9' {
		return 0, false
	}

	for i := 1; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
	}

	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}

	return v, true
}

func parseStrictNonNegativeInt(s string) (int, bool) {
	if s == "0" {
		return 0, true
	}

	return parseStrictPositiveInt(s)
}
