package setclass

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"
)

func errNotation(s string) error {
	return fmt.Errorf("%w: %q", ErrNotation, s)
}

func parseCanonicalInt(s string) (int, bool) {
	v, err := strconv.Atoi(s)
	return v, err == nil && strconv.Itoa(v) == s
}

func parseCanonicalNumbers(modStr, inner string) (int, []int, bool) {
	var nums []int
	if inner != "" {
		parts := strings.Split(inner, ", ")
		nums = make([]int, len(parts))
		for i, part := range parts {
			v, ok := parseCanonicalInt(part)
			if !ok || v < 0 {
				return 0, nil, false
			}
			nums[i] = v
		}
	}
	modulus, ok := parseCanonicalInt(modStr)
	return modulus, nums, ok
}

// String returns the canonical self-contained representation
// "<Kind>(<p_0>, <p_1>, ..., <p_{k-1}>) mod <N>" using the Rahn prime form
// (for example, "TnIn(0, 3, 7) mod 12", "Tn(0, 4, 7) mod 12", "TnIn() mod 12"),
// or "SetClass(invalid)" if sc is invalid.
func (sc SetClass) String() string {
	if !sc.IsValid() {
		return "SetClass(invalid)"
	}

	nums := sc.prime.Numbers()

	var b strings.Builder
	b.WriteString(sc.kind.String())
	b.WriteByte('(')

	for i, x := range nums {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(strconv.Itoa(x))
	}

	b.WriteString(") mod ")
	b.WriteString(strconv.Itoa(sc.prime.Modulus()))

	return b.String()
}

// Canonical returns the canonical representation of sc, identical to [SetClass.String].
func (sc SetClass) Canonical() string {
	return sc.String()
}

// Parse parses a canonical set-class representation
// "<Kind>(<p_0>, <p_1>, ..., <p_{k-1}>) mod <N>" and verifies that the tuple
// is the canonical Rahn prime form for (N, Kind).
func Parse(s string) (SetClass, error) {
	kindStr, rest, ok := strings.Cut(s, "(")
	if !ok {
		return SetClass{}, errNotation(s)
	}

	inner, modStr, ok := strings.Cut(rest, ") mod ")
	if !ok {
		return SetClass{}, errNotation(s)
	}

	kind, kindErr := ParseKind(kindStr)
	if kindErr != nil {
		return SetClass{}, fmt.Errorf("%w: %w", errNotation(s), ErrInvalidKind)
	}

	modulus, nums, ok := parseCanonicalNumbers(modStr, inner)
	if !ok {
		return SetClass{}, errNotation(s)
	}

	if err := validateModulusAndKind(modulus, kind); err != nil {
		return SetClass{}, fmt.Errorf("%w: %w", errNotation(s), err)
	}

	prev := -1
	for _, x := range nums {
		if x >= modulus {
			return SetClass{}, fmt.Errorf("%w: %w", errNotation(s), ErrInvalidPitchClass)
		}

		if x <= prev {
			return SetClass{}, errNotation(s)
		}

		prev = x
	}

	sc := MustFromNumbers(modulus, kind, nums...)
	if !slices.Equal(nums, sc.prime.Numbers()) {
		return SetClass{}, errNotation(s)
	}

	return sc, nil
}

// MustParse is like [Parse] but panics on error.
func MustParse(s string) SetClass {
	sc, err := Parse(s)
	if err != nil {
		panic(err)
	}

	return sc
}

// Notation returns the delimited literary representation "(p_0, p_1, ..., p_{k-1})"
// of the Rahn prime form of sc in the given pitch-class style.
func (sc SetClass) Notation(style pitchclass.Style) (string, error) {
	return sc.NotationWith(style, Rahn)
}

// MustNotation is like [SetClass.Notation] but panics on error.
func (sc SetClass) MustNotation(style pitchclass.Style) string {
	str, err := sc.Notation(style)
	if err != nil {
		panic(err)
	}

	return str
}

// NotationWith returns the delimited literary representation "(p_0, p_1, ..., p_{k-1})"
// of the prime form of sc under alg in the given pitch-class style.
func (sc SetClass) NotationWith(style pitchclass.Style, alg Algorithm) (string, error) {
	nums, err := sc.PrimeNumbersWith(alg)
	if err != nil {
		return "", err
	}

	n := sc.prime.Modulus()
	if err := checkLiteraryStyle(style, n, false); err != nil {
		return "", err
	}

	return formatNumbersLiterary(nums, n, style, '(', ')', ", "), nil
}

// MustNotationWith is like [SetClass.NotationWith] but panics on error.
func (sc SetClass) MustNotationWith(style pitchclass.Style, alg Algorithm) string {
	str, err := sc.NotationWith(style, alg)
	if err != nil {
		panic(err)
	}

	return str
}

// CompactNotation returns the compact literary representation "(p_0p_1...p_{k-1})"
// without separators of the Rahn prime form of sc in the given pitch-class style.
func (sc SetClass) CompactNotation(style pitchclass.Style) (string, error) {
	return sc.CompactNotationWith(style, Rahn)
}

// MustCompactNotation is like [SetClass.CompactNotation] but panics on error.
func (sc SetClass) MustCompactNotation(style pitchclass.Style) string {
	str, err := sc.CompactNotation(style)
	if err != nil {
		panic(err)
	}

	return str
}

// CompactNotationWith returns the compact literary representation "(p_0p_1...p_{k-1})"
// without separators of the prime form of sc under alg in the given pitch-class style.
func (sc SetClass) CompactNotationWith(style pitchclass.Style, alg Algorithm) (string, error) {
	nums, err := sc.PrimeNumbersWith(alg)
	if err != nil {
		return "", err
	}

	n := sc.prime.Modulus()
	if err := checkLiteraryStyle(style, n, true); err != nil {
		return "", err
	}

	return formatNumbersLiterary(nums, n, style, '(', ')', ""), nil
}

// MustCompactNotationWith is like [SetClass.CompactNotationWith] but panics on error.
func (sc SetClass) MustCompactNotationWith(style pitchclass.Style, alg Algorithm) string {
	str, err := sc.CompactNotationWith(style, alg)
	if err != nil {
		panic(err)
	}

	return str
}

func splitNotationTokens(inner string, style pitchclass.Style, modulus int) []string {
	if inner == "" {
		return nil
	}

	if strings.Contains(inner, ", ") || (style == pitchclass.Decimal && modulus > 10) {
		return strings.Split(inner, ", ")
	}

	return strings.Split(inner, "")
}

func parseBracketedSet(inner string, style pitchclass.Style, modulus int, full string) (pcset.Set, error) {
	tokens := splitNotationTokens(inner, style, modulus)
	nums := make([]int, len(tokens))

	for i, tok := range tokens {
		c, err := pitchclass.ParseNotation(tok, style, modulus)
		if err != nil {
			return pcset.Set{}, errNotation(full)
		}

		nums[i] = c.Number()
	}

	return pcset.MustNew(modulus, nums...), nil
}

// ParseNotation parses a literary prime-form or pitch-class set string "(...)"
// (delimited by ", " or compact) in the given style and modulus and constructs
// the set class of the given kind.
func ParseNotation(s string, style pitchclass.Style, modulus int, kind Kind) (SetClass, error) {
	if err := validateModulusAndKind(modulus, kind); err != nil {
		return SetClass{}, fmt.Errorf("%w: %w", errNotation(s), err)
	}

	if err := checkLiteraryStyle(style, modulus, false); err != nil {
		return SetClass{}, err
	}

	if len(s) < 2 || s[0] != '(' || s[len(s)-1] != ')' {
		return SetClass{}, errNotation(s)
	}

	set, err := parseBracketedSet(s[1:len(s)-1], style, modulus, s)
	if err != nil {
		return SetClass{}, err
	}

	return MustFromSet(set, kind), nil
}

// MustParseNotation is like [ParseNotation] but panics on error.
func MustParseNotation(s string, style pitchclass.Style, modulus int, kind Kind) SetClass {
	sc, err := ParseNotation(s, style, modulus, kind)
	if err != nil {
		panic(err)
	}

	return sc
}

// FormatMorris returns the Robert Morris compositional notation
// "<CatalogName>[<CompactPrimeFormTE>]" for N = 12 using the Rahn prime form
// (for example, "3-11[037]", "3-11B[047]", "6-1/32[012345]", "0-1[]").
func (sc SetClass) FormatMorris() (string, error) {
	return sc.FormatMorrisWith(Rahn)
}

// MustFormatMorris is like [SetClass.FormatMorris] but panics on error.
func (sc SetClass) MustFormatMorris() string {
	str, err := sc.FormatMorris()
	if err != nil {
		panic(err)
	}

	return str
}

// FormatMorrisWith returns the Robert Morris compositional notation
// "<CatalogName>[<CompactPrimeFormTE>]" for N = 12 using the prime form under alg.
func (sc SetClass) FormatMorrisWith(alg Algorithm) (string, error) {
	nums, err := sc.PrimeNumbersWith(alg)
	if err != nil {
		return "", err
	}

	name, err := sc.CatalogName()
	if err != nil {
		return "", err
	}

	return name + formatNumbersLiterary(nums, 12, pitchclass.TE, '[', ']', ""), nil
}

// MustFormatMorrisWith is like [SetClass.FormatMorrisWith] but panics on error.
func (sc SetClass) MustFormatMorrisWith(alg Algorithm) string {
	str, err := sc.FormatMorrisWith(alg)
	if err != nil {
		panic(err)
	}

	return str
}

// ParseMorris parses a Robert Morris compositional notation string
// "<CatalogLabel>[<Elements>]" at N = 12, inferring the equivalence kind from
// the catalog label ("A"/"B" suffix -> [Tn], composite "c-d1/d2" -> [TTO],
// single Forte label "c-d" -> [TnIn]), and verifies that the set [<Elements>]
// belongs to the named set class.
func ParseMorris(s string) (SetClass, error) {
	label, _, _ := strings.Cut(s, "[")

	kind := TnIn
	if strings.Contains(label, "/") {
		kind = TTO
	} else if strings.HasSuffix(label, "A") || strings.HasSuffix(label, "B") {
		kind = Tn
	}

	return ParseMorrisWithKind(s, kind)
}

// MustParseMorris is like [ParseMorris] but panics on error.
func MustParseMorris(s string) SetClass {
	sc, err := ParseMorris(s)
	if err != nil {
		panic(err)
	}

	return sc
}

// ParseMorrisWithKind parses a Robert Morris compositional notation string
// "<CatalogLabel>[<Elements>]" at N = 12 into a set class of the specified kind
// and verifies that the set [<Elements>] belongs to the named set class.
func ParseMorrisWithKind(s string, kind Kind) (SetClass, error) {
	label, rest, ok := strings.Cut(s, "[")
	elemStr := rest
	if ok && strings.HasSuffix(rest, "]") {
		elemStr = "(" + rest[:len(rest)-1] + ")"
	}

	elemSC, elemErr := ParseNotation(elemStr, pitchclass.TE, 12, kind)
	targetSC, catErr := FromCatalogName(label, kind)
	if elemErr != nil || catErr != nil || !targetSC.Equal(elemSC) {
		return SetClass{}, fmt.Errorf("%w: %w", errNotation(s), cmp.Or(elemErr, catErr, ErrNotation))
	}

	return targetSC, nil
}

// MustParseMorrisWithKind is like [ParseMorrisWithKind] but panics on error.
func MustParseMorrisWithKind(s string, kind Kind) SetClass {
	sc, err := ParseMorrisWithKind(s, kind)
	if err != nil {
		panic(err)
	}

	return sc
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
