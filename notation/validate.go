package notation

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/spelling"
)

type fixedOctaveForm struct {
	text string
	name spelling.Spelling
	w    int
}

type suffixLanguage struct {
	base  string
	mark  string
	minN  int
	baseW int
	dir   int
	name  spelling.Spelling
}

func validateNotation(rules NameRules, style OctaveStyle, separator string) error {
	forms := allNameForms(rules)
	if usesLetterCase(style) {
		for _, form := range forms {
			if _, ok := changeCase(form.text, true); !ok {
				return fmt.Errorf("%w: %q has no cased letter", ErrInvalidDefinition, form.text)
			}
			if _, ok := changeCase(form.text, false); !ok {
				return fmt.Errorf("%w: %q has no cased letter", ErrInvalidDefinition, form.text)
			}
		}
	}

	switch style {
	case OctaveStyleNone:
		return nil
	case OctaveStyleNumber:
		return validateNumberStyle(forms, separator)
	case OctaveStyleName:
		return validateNameStyle(forms, separator)
	case OctaveStyleAM:
		return validateAMStyle(forms)
	case OctaveStyleHelmholtz, OctaveStyleLilyPond, OctaveStyleHelmholtzEnglish:
		languages, fixed := suffixForms(forms, style)
		return validateSuffixForms(languages, fixed)
	case OctaveStyleABC:
		return validateABCStyle(forms)
	case OctaveStyleHelmholtzNumbered:
		return validateNumberedStyle(forms)
	default:
		return fmt.Errorf("%w: octave style %d", ErrInvalidDefinition, style)
	}
}

func usesLetterCase(style OctaveStyle) bool {
	return style >= OctaveStyleAM && style <= OctaveStyleABC
}

func validateNumberStyle(forms []nameEntry, separator string) error {
	for i, a := range forms {
		for j, b := range forms {
			if i == j {
				continue
			}
			x, y := a.text+separator, b.text+separator
			if x == y {
				return ambiguousForms(a.text, b.text)
			}
			if d, ok := properPrefixRemainder(x, y); ok {
				for _, suffix := range []string{"0", "1"} {
					if strictSignedNumber(d+suffix) && strictSignedNumber(suffix) {
						return ambiguousForms(a.text, b.text)
					}
				}
			}
		}
	}
	return nil
}

func validateNameStyle(forms []nameEntry, separator string) error {
	for i, a := range forms {
		for j, b := range forms {
			if i == j {
				continue
			}
			x, y := a.text+separator, b.text+separator
			if x == y {
				return ambiguousForms(a.text, b.text)
			}
			if _, ok := properPrefixRemainder(x, y); ok {
				return ambiguousForms(a.text, b.text)
			}
		}
	}
	return nil
}

func validateAMStyle(forms []nameEntry) error {
	type result struct {
		name spelling.Spelling
		w    int
	}
	seen := make(map[string]result)
	for _, form := range forms {
		for w := range 9 {
			text := formatAM(form.text, w)
			previous, exists := seen[text]
			if exists && (previous.name != form.name || previous.w != w) {
				return ambiguousForms(form.text, text)
			}
			seen[text] = result{name: form.name, w: w}
		}
	}
	return nil
}

func suffixForms(forms []nameEntry, style OctaveStyle) ([]suffixLanguage, []fixedOctaveForm) {
	languages := make([]suffixLanguage, 0, len(forms)*2)
	var fixed []fixedOctaveForm
	for _, form := range forms {
		upper, _ := changeCase(form.text, true)
		lower, _ := changeCase(form.text, false)
		switch style {
		case OctaveStyleHelmholtz:
			languages = append(languages,
				suffixLanguage{base: upper, mark: "͵", baseW: 2, dir: -1, name: form.name},
				suffixLanguage{base: lower, mark: "′", baseW: 3, dir: 1, name: form.name})
		case OctaveStyleLilyPond:
			languages = append(languages,
				suffixLanguage{base: lower, mark: ",", baseW: 3, dir: -1, name: form.name},
				suffixLanguage{base: lower, mark: "'", baseW: 3, dir: 1, name: form.name})
		case OctaveStyleHelmholtzEnglish:
			languages = append(languages,
				suffixLanguage{base: upper, mark: "͵", minN: 3, baseW: 2, dir: -1, name: form.name},
				suffixLanguage{base: lower, mark: "′", baseW: 3, dir: 1, name: form.name})
			triple, _ := repeatFirstCased(upper, 3)
			double, _ := repeatFirstCased(upper, 2)
			fixed = append(fixed,
				fixedOctaveForm{text: triple, name: form.name, w: 0},
				fixedOctaveForm{text: double, name: form.name, w: 1},
				fixedOctaveForm{text: upper, name: form.name, w: 2})
		}
	}
	return languages, fixed
}

func validateSuffixForms(languages []suffixLanguage, fixed []fixedOctaveForm) error {
	for i, a := range languages {
		for j, b := range languages {
			if i != j && suffixLanguagesConflict(a, b) {
				return ambiguousForms(a.base, b.base)
			}
		}
		for _, f := range fixed {
			if w, ok := suffixLanguageMember(a, f.text); ok && (a.name != f.name || w != f.w) {
				return ambiguousForms(a.base, f.text)
			}
		}
	}
	for i, a := range fixed {
		for j, b := range fixed {
			if i != j && a.text == b.text && (a.name != b.name || a.w != b.w) {
				return ambiguousForms(a.text, b.text)
			}
		}
	}
	return nil
}

func suffixLanguagesConflict(a, b suffixLanguage) bool {
	if a.base == b.base {
		if a.mark != b.mark {
			if a.minN != 0 || b.minN != 0 {
				return false
			}
			return a.name != b.name || a.baseW != b.baseW
		}
		if a.name != b.name {
			return true
		}
		return a.baseW != b.baseW || a.dir != b.dir
	}

	d, ok := properPrefixRemainder(a.base, b.base)
	if !ok {
		return false
	}
	repeats, ok := repeatedMarkCount(d, a.mark)
	if !ok || repeats == 0 {
		return false
	}
	if a.mark != b.mark {
		if b.minN != 0 || repeats < a.minN {
			return false
		}
		wa, okA := checkedAdd(a.baseW, a.dir*repeats)
		return okA && (a.name != b.name || wa != b.baseW)
	}

	// For equal marks, b can use any permitted count and a uses that count
	// plus the repeats in d. If the written-octave slopes differ, only one
	// such string can have the same octave; all the others are ambiguous.
	if a.name != b.name || a.dir != b.dir {
		return true
	}
	shifted, ok := checkedAdd(a.baseW, a.dir*repeats)
	return !ok || shifted != b.baseW
}

func suffixLanguageMember(language suffixLanguage, text string) (int, bool) {
	d, ok := strings.CutPrefix(text, language.base)
	if !ok {
		return 0, false
	}
	n, ok := repeatedMarkCount(d, language.mark)
	if !ok || n < language.minN {
		return 0, false
	}
	w, ok := checkedAdd(language.baseW, language.dir*n)
	return w, ok
}

func validateABCStyle(forms []nameEntry) error {
	for i, a := range forms {
		for j, b := range forms {
			if i == j {
				continue
			}
			for _, upperA := range []bool{true, false} {
				for _, upperB := range []bool{true, false} {
					baseA, _ := changeCase(a.text, upperA)
					baseB, _ := changeCase(b.text, upperB)
					wA, wB := 4, 4
					if !upperA {
						wA = 5
					}
					if !upperB {
						wB = 5
					}
					if baseA == baseB {
						return ambiguousForms(a.text, b.text)
					}
					d, ok := properPrefixRemainder(baseA, baseB)
					if !ok || !onlyABCMarks(d) {
						continue
					}
					if a.name != b.name || wA+abcMarkValue(d) != wB {
						return ambiguousForms(a.text, b.text)
					}
				}
			}
		}
	}
	return nil
}

func validateNumberedStyle(forms []nameEntry) error {
	type numbered struct {
		base  string
		baseW int
		dir   int
		name  spelling.Spelling
	}
	languages := make([]numbered, 0, len(forms)*2)
	var aliases []fixedOctaveForm
	for _, form := range forms {
		upper, _ := changeCase(form.text, true)
		lower, _ := changeCase(form.text, false)
		languages = append(languages,
			numbered{base: upper, baseW: 2, dir: -1, name: form.name},
			numbered{base: lower, baseW: 3, dir: 1, name: form.name})
		aliases = append(aliases,
			fixedOctaveForm{text: "2" + upper, name: form.name, w: 0},
			fixedOctaveForm{text: "1" + upper, name: form.name, w: 1})
	}

	for i, a := range languages {
		for j, b := range languages {
			if i == j {
				continue
			}
			if a.base == b.base {
				if a.name != b.name || a.baseW != b.baseW || a.dir != b.dir {
					return ambiguousForms(a.base, b.base)
				}
				continue
			}
			d, ok := properPrefixRemainder(a.base, b.base)
			if !ok {
				continue
			}
			for kB := range 2 {
				candidate := d
				if kB != 0 {
					candidate += strconv.Itoa(kB)
				}
				kA, err := parsePositiveDecimal(candidate)
				if err != nil {
					continue
				}
				wA, okA := checkedAdd(a.baseW, a.dir*kA)
				wB, okB := checkedAdd(b.baseW, b.dir*kB)
				if okA && okB && (a.name != b.name || wA != wB) {
					return ambiguousForms(a.base, b.base)
				}
			}
		}

		for _, alias := range aliases {
			d, ok := strings.CutPrefix(alias.text, a.base)
			if !ok {
				continue
			}
			k := 0
			var err error
			if d != "" {
				k, err = parsePositiveDecimal(d)
				if err != nil {
					continue
				}
			}
			w, ok := checkedAdd(a.baseW, a.dir*k)
			if ok && (a.name != alias.name || w != alias.w) {
				return ambiguousForms(a.base, alias.text)
			}
		}
	}
	for i, a := range aliases {
		for j, b := range aliases {
			if i != j && a.text == b.text && (a.name != b.name || a.w != b.w) {
				return ambiguousForms(a.text, b.text)
			}
		}
	}
	return nil
}

func ambiguousForms(a, b string) error {
	return fmt.Errorf("%w: %q and %q", ErrAmbiguous, a, b)
}

func properPrefixRemainder(short, long string) (string, bool) {
	d, ok := strings.CutPrefix(long, short)
	return d, ok && d != ""
}

func strictSignedNumber(text string) bool {
	if text == "0" {
		return true
	}
	if text == "" || text == "-0" || text[0] == '+' {
		return false
	}
	digits := text
	if text[0] == '-' {
		digits = text[1:]
	}
	return positiveDigits(digits)
}

func positiveDigits(text string) bool {
	if text == "" || text[0] < '1' || text[0] > '9' {
		return false
	}
	for i := 1; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

func parsePositiveDecimal(text string) (int, error) {
	if !positiveDigits(text) {
		return 0, fmt.Errorf("invalid positive decimal")
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func repeatedMarkCount(text, mark string) (int, bool) {
	if text == "" {
		return 0, true
	}
	if mark == "" {
		return 0, false
	}
	count := 0
	for text != "" {
		var ok bool
		text, ok = strings.CutPrefix(text, mark)
		if !ok {
			return 0, false
		}
		count++
	}
	return count, true
}

func onlyABCMarks(text string) bool {
	for i := range len(text) {
		if text[i] != '\'' && text[i] != ',' {
			return false
		}
	}
	return true
}

func abcMarkValue(text string) int {
	value := 0
	for i := range len(text) {
		if text[i] == '\'' {
			value++
		} else {
			value--
		}
	}
	return value
}

func changeCase(text string, upper bool) (string, bool) {
	for offset, r := range text {
		lower, upperRune := unicode.ToLower(r), unicode.ToUpper(r)
		if lower == upperRune {
			continue
		}
		replacement := lower
		if upper {
			replacement = upperRune
		}
		size := utf8.RuneLen(r)
		return text[:offset] + string(replacement) + text[offset+size:], true
	}
	return "", false
}

func repeatFirstCased(text string, count int) (string, bool) {
	for offset, r := range text {
		if unicode.ToLower(r) == unicode.ToUpper(r) {
			continue
		}
		size := utf8.RuneLen(r)
		letter := text[offset : offset+size]
		return text[:offset] + strings.Repeat(letter, count) + text[offset+size:], true
	}
	return "", false
}

func formatAM(text string, w int) string {
	switch w {
	case 0:
		upper, _ := changeCase(text, true)
		return upper + "₂"
	case 1:
		upper, _ := changeCase(text, true)
		return upper + "₁"
	case 2:
		upper, _ := changeCase(text, true)
		return upper
	case 3:
		lower, _ := changeCase(text, false)
		return lower
	default:
		lower, _ := changeCase(text, false)
		return lower + [...]string{"", "¹", "²", "³", "⁴", "⁵"}[w-3]
	}
}

func checkedAdd(a, b int) (int, bool) {
	if b > 0 && a > int(^uint(0)>>1)-b {
		return 0, false
	}
	minInt := -int(^uint(0)>>1) - 1
	if b < 0 && a < minInt-b {
		return 0, false
	}
	return a + b, true
}

func checkedSub(a, b int) (int, bool) {
	if b == -int(^uint(0)>>1)-1 {
		if a >= 0 {
			return 0, false
		}
		return a - b, true
	}
	return checkedAdd(a, -b)
}

func octaveRangeError() error {
	return fmt.Errorf("%w: octave mark arithmetic", octave.ErrOutOfRange)
}
