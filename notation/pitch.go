package notation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// FormatPitch formats p with an explicit musical system and octave context.
// The system alone converts between the written octave and the positional
// octave stored in p. No traditional system, numbering, or catalog is implied.
func (n Notation) FormatPitch(p pitch.Pitch, s *system.System, ctx OctaveContext) (string, error) {
	if !p.IsValid() {
		return "", fmt.Errorf("%w: format pitch", ErrUndefined)
	}
	if err := n.checkPitchOperation(s, ctx); err != nil {
		return "", err
	}
	name, err := n.rules.Format(p.Spelling())
	if err != nil {
		return "", err
	}
	written, err := s.WrittenOctave(p)
	if err != nil {
		return "", fmt.Errorf("notation: written octave: %w", err)
	}
	return n.formatWritten(name, written, ctx)
}

// MustFormatPitch is like Notation.FormatPitch but panics with its error on
// failure.
func (n Notation) MustFormatPitch(p pitch.Pitch, s *system.System, ctx OctaveContext) string {
	return must(n.FormatPitch(p, s, ctx))
}

// ParsePitch parses text and reconstructs the pitch through the explicit
// musical system. Dependency errors remain available through errors.Is.
func (n Notation) ParsePitch(text string, s *system.System, ctx OctaveContext) (pitch.Pitch, error) {
	if err := n.checkPitchOperation(s, ctx); err != nil {
		return pitch.Pitch{}, err
	}
	name, written, err := n.parseWritten(text, ctx)
	if err != nil {
		return pitch.Pitch{}, err
	}
	p, err := s.FromWritten(name, written)
	if err != nil {
		return pitch.Pitch{}, fmt.Errorf("notation: pitch from written octave: %w", err)
	}
	return p, nil
}

// MustParsePitch is like Notation.ParsePitch but panics with its error on
// failure.
func (n Notation) MustParsePitch(text string, s *system.System, ctx OctaveContext) pitch.Pitch {
	return must(n.ParsePitch(text, s, ctx))
}

// ConvertPitch parses text with from and formats the same pitch with to. One
// explicit musical system is used for both operations.
func ConvertPitch(text string, from, to Notation, s *system.System, fromCtx, toCtx OctaveContext) (string, error) {
	p, err := from.ParsePitch(text, s, fromCtx)
	if err != nil {
		return "", err
	}
	return to.FormatPitch(p, s, toCtx)
}

// MustConvertPitch is like ConvertPitch but panics with its error on failure.
func MustConvertPitch(text string, from, to Notation, s *system.System, fromCtx, toCtx OctaveContext) string {
	return must(ConvertPitch(text, from, to, s, fromCtx, toCtx))
}

func (n Notation) checkPitchOperation(s *system.System, ctx OctaveContext) error {
	if !n.IsDefined() || !s.IsValid() {
		return fmt.Errorf("%w: pitch operation", ErrUndefined)
	}
	if c, ok := n.ChromaticUnit(); ok && c != s.Chromatic() {
		return fmt.Errorf("%w: notation c=%d, system c=%d", ErrIncompatibleSystem, c, s.Chromatic())
	}
	if n.style == OctaveStyleNone {
		return fmt.Errorf("%w: name-only notation", ErrNoOctave)
	}
	if n.style == OctaveStyleNumber && !ctx.Numbering.IsDefined() {
		return fmt.Errorf("%w: octave numbering", ErrUndefined)
	}
	return nil
}

func (n Notation) formatWritten(name string, written octave.Octave, ctx OctaveContext) (string, error) {
	w, err := written.ASPN()
	if err != nil {
		return "", fmt.Errorf("notation: written octave: %w", err)
	}

	switch n.style {
	case OctaveStyleNumber:
		visible, err := written.Number(ctx.Numbering)
		if err != nil {
			return "", fmt.Errorf("notation: visible octave number: %w", err)
		}
		return joinStrings(name, n.separator, strconv.Itoa(visible))
	case OctaveStyleName:
		visible := written.Name(ctx.Catalog)
		if visible == "" {
			return "", fmt.Errorf("%w: empty octave name", ErrNoOctave)
		}
		return joinStrings(name, n.separator, visible)
	case OctaveStyleAM:
		if w < 0 || w > 8 {
			return "", fmt.Errorf("%w: AM octave %d", ErrNoOctave, w)
		}
		return formatAM(name, w), nil
	case OctaveStyleHelmholtz:
		if w <= 2 {
			count, ok := checkedSub(2, w)
			if !ok {
				return "", octaveRangeError()
			}
			upper, _ := changeCase(name, true)
			return joinRepeated(upper, "͵", count)
		}
		count, ok := checkedSub(w, 3)
		if !ok {
			return "", octaveRangeError()
		}
		lower, _ := changeCase(name, false)
		return joinRepeated(lower, "′", count)
	case OctaveStyleHelmholtzEnglish:
		switch {
		case w == 0:
			upper, _ := changeCase(name, true)
			text, _ := repeatFirstCased(upper, 3)
			return text, nil
		case w == 1:
			upper, _ := changeCase(name, true)
			text, _ := repeatFirstCased(upper, 2)
			return text, nil
		case w == 2:
			upper, _ := changeCase(name, true)
			return upper, nil
		case w < 0:
			count, ok := checkedSub(2, w)
			if !ok {
				return "", octaveRangeError()
			}
			upper, _ := changeCase(name, true)
			return joinRepeated(upper, "͵", count)
		default:
			count, ok := checkedSub(w, 3)
			if !ok {
				return "", octaveRangeError()
			}
			lower, _ := changeCase(name, false)
			return joinRepeated(lower, "′", count)
		}
	case OctaveStyleHelmholtzNumbered:
		var count int
		var ok bool
		var text string
		if w <= 2 {
			count, ok = checkedSub(2, w)
			text, _ = changeCase(name, true)
		} else {
			count, ok = checkedSub(w, 3)
			text, _ = changeCase(name, false)
		}
		if !ok {
			return "", octaveRangeError()
		}
		if count == 0 {
			return text, nil
		}
		return joinStrings(text, strconv.Itoa(count))
	case OctaveStyleLilyPond:
		lower, _ := changeCase(name, false)
		if w < 3 {
			count, ok := checkedSub(3, w)
			if !ok {
				return "", octaveRangeError()
			}
			return joinRepeated(lower, ",", count)
		}
		count, ok := checkedSub(w, 3)
		if !ok {
			return "", octaveRangeError()
		}
		return joinRepeated(lower, "'", count)
	case OctaveStyleABC:
		switch {
		case w == 4:
			upper, _ := changeCase(name, true)
			return upper, nil
		case w == 5:
			lower, _ := changeCase(name, false)
			return lower, nil
		case w < 4:
			count, ok := checkedSub(4, w)
			if !ok {
				return "", octaveRangeError()
			}
			upper, _ := changeCase(name, true)
			return joinRepeated(upper, ",", count)
		default:
			count, ok := checkedSub(w, 5)
			if !ok {
				return "", octaveRangeError()
			}
			lower, _ := changeCase(name, false)
			return joinRepeated(lower, "'", count)
		}
	default:
		return "", fmt.Errorf("%w: name-only notation", ErrNoOctave)
	}
}

func (n Notation) parseWritten(text string, ctx OctaveContext) (spelling.Spelling, octave.Octave, error) {
	forms := allNameForms(n.rules)
	switch n.style {
	case OctaveStyleNumber:
		for _, form := range forms {
			rest, ok := strings.CutPrefix(text, form.text+n.separator)
			if !ok || !strictSignedNumber(rest) {
				continue
			}
			visible, err := strconv.Atoi(rest)
			if err != nil {
				return spelling.Spelling{}, octave.Octave{}, fmt.Errorf("%w: visible octave number", number.ErrOutOfRange)
			}
			written, err := octave.FromNumber(visible, ctx.Numbering)
			if err != nil {
				return spelling.Spelling{}, octave.Octave{}, fmt.Errorf("notation: numbered octave: %w", err)
			}
			return form.name, written, nil
		}
	case OctaveStyleName:
		for _, form := range forms {
			rest, ok := strings.CutPrefix(text, form.text+n.separator)
			if !ok {
				continue
			}
			written, err := octave.FromName(rest, ctx.Catalog)
			if err != nil {
				return spelling.Spelling{}, octave.Octave{}, fmt.Errorf("notation: named octave: %w", err)
			}
			return form.name, written, nil
		}
	case OctaveStyleAM:
		for _, form := range forms {
			for w := range 9 {
				if text == formatAM(form.text, w) {
					return form.name, octave.FromASPN(w), nil
				}
			}
			upper, _ := changeCase(form.text, true)
			lower, _ := changeCase(form.text, false)
			if rest, ok := strings.CutPrefix(text, upper); ok && onlyRunes(rest, "₀₁₂₃₄₅₆₇₈₉") {
				return spelling.Spelling{}, octave.Octave{}, fmt.Errorf("%w: AM octave", ErrNoOctave)
			}
			if rest, ok := strings.CutPrefix(text, lower); ok && onlyRunes(rest, "⁰¹²³⁴⁵⁶⁷⁸⁹") {
				return spelling.Spelling{}, octave.Octave{}, fmt.Errorf("%w: AM octave", ErrNoOctave)
			}
		}
	case OctaveStyleHelmholtz, OctaveStyleLilyPond, OctaveStyleHelmholtzEnglish:
		languages, fixed := suffixForms(forms, n.style)
		for _, candidate := range fixed {
			if text == candidate.text {
				return candidate.name, octave.FromASPN(candidate.w), nil
			}
		}
		for _, language := range languages {
			rest, ok := strings.CutPrefix(text, language.base)
			if !ok {
				continue
			}
			count, ok := repeatedMarkCount(rest, language.mark)
			if !ok || count < language.minN {
				continue
			}
			w, ok := checkedAdd(language.baseW, language.dir*count)
			if !ok {
				return spelling.Spelling{}, octave.Octave{}, octaveRangeError()
			}
			return language.name, octave.FromASPN(w), nil
		}
	case OctaveStyleABC:
		for _, form := range forms {
			for _, upper := range []bool{true, false} {
				base, _ := changeCase(form.text, upper)
				rest, ok := strings.CutPrefix(text, base)
				if !ok || !onlyABCMarks(rest) {
					continue
				}
				baseW := 4
				if !upper {
					baseW = 5
				}
				w, ok := checkedAdd(baseW, abcMarkValue(rest))
				if !ok {
					return spelling.Spelling{}, octave.Octave{}, octaveRangeError()
				}
				return form.name, octave.FromASPN(w), nil
			}
		}
	case OctaveStyleHelmholtzNumbered:
		for _, form := range forms {
			upper, _ := changeCase(form.text, true)
			if text == "2"+upper {
				return form.name, octave.FromASPN(0), nil
			}
			if text == "1"+upper {
				return form.name, octave.FromASPN(1), nil
			}
		}
		for _, form := range forms {
			for _, upper := range []bool{true, false} {
				base, _ := changeCase(form.text, upper)
				rest, ok := strings.CutPrefix(text, base)
				if !ok {
					continue
				}
				visible := 0
				if rest != "" {
					if !positiveDigits(rest) {
						continue
					}
					var err error
					visible, err = strconv.Atoi(rest)
					if err != nil {
						return spelling.Spelling{}, octave.Octave{}, octaveRangeError()
					}
				}
				baseW, dir := 2, -1
				if !upper {
					baseW, dir = 3, 1
				}
				w, ok := checkedAdd(baseW, dir*visible)
				if !ok {
					return spelling.Spelling{}, octave.Octave{}, octaveRangeError()
				}
				return form.name, octave.FromASPN(w), nil
			}
		}
	}
	return spelling.Spelling{}, octave.Octave{}, fmt.Errorf("%w: %q", ErrParse, text)
}

func onlyRunes(text, allowed string) bool {
	if text == "" {
		return false
	}
	for _, r := range text {
		if !strings.ContainsRune(allowed, r) {
			return false
		}
	}
	return true
}

func joinRepeated(base, mark string, count int) (string, error) {
	if count < 0 || mark == "" || count > (int(^uint(0)>>1)-len(base))/len(mark) {
		return "", octaveRangeError()
	}
	return joinStrings(base, strings.Repeat(mark, count))
}

func joinStrings(parts ...string) (string, error) {
	total := 0
	for _, part := range parts {
		if len(part) > int(^uint(0)>>1)-total {
			return "", octaveRangeError()
		}
		total += len(part)
	}
	var builder strings.Builder
	builder.Grow(total)
	for _, part := range parts {
		builder.WriteString(part)
	}
	return builder.String(), nil
}
