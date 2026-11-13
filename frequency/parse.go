package frequency

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/go-muse/musex/cents"
)

// Parse reads a criterion in the canonical functional grammar: unset, strict,
// cents(...), hertz(...), relative(...), all(...), and any(...). Lists use a
// comma followed by exactly one space. The whole input is checked for grammar
// before the first semantic error is returned. Parsing is iterative and
// supports arbitrarily deep nesting without recursion.
func Parse(text string) (Criterion, error) {
	p := criterionParser{text: text}
	criterion, err := p.parse()
	if err != nil {
		return Criterion{}, err
	}
	if p.semantic != nil {
		return Criterion{}, p.semantic
	}
	return criterion, nil
}

// MustParse calls Parse and panics with its error, if any.
func MustParse(text string) Criterion {
	c, err := Parse(text)
	if err != nil {
		panic(err)
	}
	return c
}

type parseFrame struct {
	kind  Kind
	parts []Criterion
}

type criterionParser struct {
	text     string
	pos      int
	stack    []parseFrame
	semantic error
}

func (p *criterionParser) parse() (Criterion, error) {
	if len(p.text) == 0 {
		return Criterion{}, invalidCriterionText(p.text, 0, nil)
	}
	for {
		criterion, opened, err := p.nextCriterion()
		if err != nil {
			return Criterion{}, err
		}
		if opened {
			continue
		}

		for {
			if len(p.stack) == 0 {
				if p.pos != len(p.text) {
					return Criterion{}, invalidCriterionText(p.text, p.pos, nil)
				}
				return criterion, nil
			}

			parentIndex := len(p.stack) - 1
			parent := &p.stack[parentIndex]
			parent.parts = append(parent.parts, criterion)
			if strings.HasPrefix(p.text[p.pos:], ", ") {
				p.pos += 2
				break
			}
			if p.pos < len(p.text) && p.text[p.pos] == ')' {
				p.pos++
				criterion = Criterion{kind: parent.kind, parts: parent.parts}
				p.stack = p.stack[:parentIndex]
				continue
			}
			return Criterion{}, invalidCriterionText(p.text, p.pos, nil)
		}
	}
}

func (p *criterionParser) nextCriterion() (Criterion, bool, error) {
	start := p.pos
	switch {
	case strings.HasPrefix(p.text[start:], "unset"):
		p.pos += len("unset")
		if len(p.stack) > 0 {
			p.recordSemantic(ErrUnset, start, nil)
		}
		return Criterion{}, false, nil
	case strings.HasPrefix(p.text[start:], "strict"):
		p.pos += len("strict")
		return Strict(), false, nil
	case strings.HasPrefix(p.text[start:], "cents("):
		return p.parseCents(start)
	case strings.HasPrefix(p.text[start:], "hertz("):
		return p.parseNumber(start, KindHertz, len("hertz("))
	case strings.HasPrefix(p.text[start:], "relative("):
		return p.parseNumber(start, KindRelative, len("relative("))
	case strings.HasPrefix(p.text[start:], "all("):
		return p.openComposite(start, KindAllOf)
	case strings.HasPrefix(p.text[start:], "any("):
		return p.openComposite(start, KindAnyOf)
	default:
		return Criterion{}, false, invalidCriterionText(p.text, start, nil)
	}
}

func (p *criterionParser) parseCents(start int) (Criterion, bool, error) {
	contentStart := start + len("cents(")
	depth := 0
	close := -1
	for i := contentStart; i < len(p.text); i++ {
		switch p.text[i] {
		case '(':
			depth++
		case ')':
			if depth == 0 {
				close = i
			} else {
				depth--
			}
		}
		if close >= 0 {
			break
		}
	}
	if close < 0 {
		return Criterion{}, false, invalidCriterionText(p.text, start, nil)
	}
	p.pos = close + 1
	value, err := cents.Parse(p.text[contentStart:close])
	if err != nil {
		if errors.Is(err, cents.ErrInvalidText) {
			return Criterion{}, false, invalidCriterionText(p.text, contentStart, err)
		}
		p.recordSemantic(ErrInvalidTolerance, contentStart, err)
		return Criterion{kind: KindCents}, false, nil
	}
	if value.Sign() < 0 {
		p.recordSemantic(ErrInvalidTolerance, contentStart, nil)
	}
	return Criterion{kind: KindCents, tolerance: value}, false, nil
}

func (p *criterionParser) parseNumber(start int, kind Kind, prefixLength int) (Criterion, bool, error) {
	contentStart := start + prefixLength
	closeOffset := strings.IndexByte(p.text[contentStart:], ')')
	if closeOffset < 0 {
		return Criterion{}, false, invalidCriterionText(p.text, start, nil)
	}
	close := contentStart + closeOffset
	number := p.text[contentStart:close]
	if position, ok := validNumber(number); !ok {
		return Criterion{}, false, invalidCriterionText(p.text, contentStart+position, nil)
	}
	p.pos = close + 1
	value, err := strconv.ParseFloat(number, 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return Criterion{}, false, invalidCriterionText(p.text, contentStart, nil)
	}
	nonzero := nonzeroMantissa(number)
	negative := strings.HasPrefix(number, "-") && nonzero
	if negative || math.IsInf(value, 0) || (nonzero && value == 0) {
		p.recordSemantic(ErrInvalidTolerance, contentStart, nil)
	}
	if value == 0 {
		value = 0 // Normalize negative zero.
	}
	return Criterion{kind: kind, number: value}, false, nil
}

func (p *criterionParser) openComposite(start int, kind Kind) (Criterion, bool, error) {
	p.pos = start + len(kind.String()) + 1
	if p.pos < len(p.text) && p.text[p.pos] == ')' {
		p.pos++
		p.recordSemantic(ErrNoComponents, start, nil)
		return Criterion{kind: kind}, false, nil
	}
	p.stack = append(p.stack, parseFrame{kind: kind})
	return Criterion{}, true, nil
}

func (p *criterionParser) recordSemantic(kind error, position int, cause error) {
	if p.semantic == nil {
		p.semantic = semanticError(kind, p.text, position, cause)
	}
}

func validNumber(text string) (int, bool) {
	if text == "" {
		return 0, false
	}
	position := 0
	if text[position] == '-' {
		position++
		if position == len(text) {
			return position, false
		}
	}
	position, ok := scanDigits(text, position)
	if !ok {
		return position, false
	}
	if position < len(text) && text[position] == '.' {
		position++
		position, ok = scanDigits(text, position)
		if !ok {
			return position, false
		}
	}
	if position < len(text) && text[position] == 'e' {
		position++
		if position < len(text) && (text[position] == '+' || text[position] == '-') {
			position++
		}
		position, ok = scanDigits(text, position)
		if !ok {
			return position, false
		}
	}
	return position, position == len(text)
}

func scanDigits(text string, position int) (int, bool) {
	start := position
	for position < len(text) && text[position] >= '0' && text[position] <= '9' {
		position++
	}
	return position, position > start
}

func nonzeroMantissa(text string) bool {
	for i := 0; i < len(text) && text[i] != 'e'; i++ {
		if text[i] >= '1' && text[i] <= '9' {
			return true
		}
	}
	return false
}
