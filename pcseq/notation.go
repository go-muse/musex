package pcseq

import (
	"strconv"
	"strings"

	"github.com/go-muse/musex/pitchclass"
)

func parseDecimal(text string) (int, bool) {
	if text == "" || len(text) > 1 && text[0] == '0' {
		return 0, false
	}
	for i := range len(text) {
		if text[i] < '0' || text[i] > '9' {
			return 0, false
		}
	}
	value, err := strconv.ParseInt(text, 10, 0)
	if err != nil {
		return 0, false
	}
	return int(value), true
}

// Parse parses canonical sequence notation.
func Parse(s string) (Sequence, error) {
	body, modulusText, ok := strings.Cut(s, "> mod ")
	if !ok || !strings.HasPrefix(body, "<") {
		return Sequence{}, invalidNotation()
	}
	modulus, ok := parseDecimal(modulusText)
	if !ok {
		return Sequence{}, invalidNotation()
	}
	content := strings.TrimPrefix(body, "<")
	var numbers []int
	if content != "" {
		tokens := strings.Split(content, ", ")
		numbers = make([]int, len(tokens))
		for i, token := range tokens {
			number, valid := parseDecimal(token)
			if !valid {
				return Sequence{}, invalidNotation()
			}
			numbers[i] = number
		}
	}
	if err := checkModulus(modulus); err != nil {
		return Sequence{}, err
	}
	for _, number := range numbers {
		if number >= modulus {
			return Sequence{}, invalidNotation()
		}
	}
	return newSequenceUnchecked(modulus, encodeNumbers(numbers)), nil
}

func checkNotationStyle(style pitchclass.Style, modulus int) error {
	if _, err := pitchclass.MustNew(0, modulus).Notation(style); err != nil {
		return invalidNotation()
	}
	return nil
}

func formatNotation(modulus int, elements string, style pitchclass.Style, compact bool) (string, error) {
	if err := checkNotationStyle(style, modulus); err != nil {
		return "", err
	}
	if compact && style == pitchclass.Decimal && modulus > 10 {
		return "", invalidNotation()
	}
	separator := ", "
	if compact {
		separator = ""
	}
	parts := make([]string, len(elements)/2)
	for i := range len(parts) {
		parts[i] = classAt(modulus, elements, i).MustNotation(style)
	}
	return "⟨" + strings.Join(parts, separator) + "⟩", nil
}

// Notation returns delimited literary sequence notation.
func (s Sequence) Notation(style pitchclass.Style) (string, error) {
	if !s.IsValid() {
		return "", invalidValue()
	}
	return formatNotation(s.modulus, s.elements, style, false)
}

// CompactNotation returns compact literary sequence notation.
func (s Sequence) CompactNotation(style pitchclass.Style) (string, error) {
	if !s.IsValid() {
		return "", invalidValue()
	}
	return formatNotation(s.modulus, s.elements, style, true)
}

func literaryContent(text string) (string, bool) {
	switch {
	case strings.HasPrefix(text, "⟨") && strings.HasSuffix(text, "⟩"):
		return strings.TrimSuffix(strings.TrimPrefix(text, "⟨"), "⟩"), true
	case strings.HasPrefix(text, "<") && strings.HasSuffix(text, ">"):
		return text[1 : len(text)-1], true
	default:
		return "", false
	}
}

// ParseNotation parses literary sequence notation in an explicit style and modulus.
func ParseNotation(s string, style pitchclass.Style, modulus int) (Sequence, error) {
	if err := checkModulus(modulus); err != nil {
		return Sequence{}, err
	}
	if err := checkNotationStyle(style, modulus); err != nil {
		return Sequence{}, err
	}
	content, ok := literaryContent(s)
	if !ok {
		return Sequence{}, invalidNotation()
	}
	if content == "" {
		return Empty(modulus)
	}
	if strings.Contains(content, ", ") {
		tokens := strings.Split(content, ", ")
		classes := make([]pitchclass.Class, len(tokens))
		for i, token := range tokens {
			class, err := pitchclass.ParseNotation(token, style, modulus)
			if err != nil {
				return Sequence{}, invalidNotation()
			}
			classes[i] = class
		}
		return FromClassesWithModulus(modulus, classes...)
	}
	if style == pitchclass.Decimal && modulus > 10 {
		class, err := pitchclass.ParseNotation(content, style, modulus)
		if err != nil {
			return Sequence{}, invalidNotation()
		}
		return FromClassesWithModulus(modulus, class)
	}
	classes := make([]pitchclass.Class, len(content))
	for i := range len(content) {
		class, err := pitchclass.ParseNotation(content[i:i+1], style, modulus)
		if err != nil {
			return Sequence{}, invalidNotation()
		}
		classes[i] = class
	}
	return FromClassesWithModulus(modulus, classes...)
}

// ParseRow parses canonical notation and validates the row invariant.
func ParseRow(s string) (Row, error) {
	sequence, err := Parse(s)
	if err != nil {
		return Row{}, err
	}
	return RowFromSequence(sequence)
}

// Notation returns delimited literary row notation.
func (r Row) Notation(style pitchclass.Style) (string, error) {
	if !r.IsValid() {
		return "", invalidValue()
	}
	return formatNotation(r.modulus, r.elements, style, false)
}

// CompactNotation returns compact literary row notation.
func (r Row) CompactNotation(style pitchclass.Style) (string, error) {
	if !r.IsValid() {
		return "", invalidValue()
	}
	return formatNotation(r.modulus, r.elements, style, true)
}

// ParseRowNotation parses literary notation and validates the row invariant.
func ParseRowNotation(s string, style pitchclass.Style, modulus int) (Row, error) {
	sequence, err := ParseNotation(s, style, modulus)
	if err != nil {
		return Row{}, err
	}
	return RowFromSequence(sequence)
}
