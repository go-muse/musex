package notation_test

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/built-in-rules.csv
var builtInRulesFixture string

//go:embed testdata/microtonal-symbols.csv
var microtonalSymbolsFixture string

//go:embed testdata/octave-marks.csv
var octaveMarksFixture string

type testHelper interface {
	require.TestingT
	Helper()
	Fatalf(format string, args ...any)
}

type expectedNotation struct {
	id            string
	style         notation.OctaveStyle
	forms         map[spelling.Spelling]string
	aliases       map[string]spelling.Spelling
	chromaticUnit int
}

func (e expectedNotation) names() []spelling.Spelling {
	names := make([]spelling.Spelling, 0, len(e.forms))
	for name := range e.forms {
		names = append(names, name)
	}
	slices.SortFunc(names, spelling.Spelling.Compare)
	return names
}

func readFixture(t testHelper, text string) [][]string {
	t.Helper()
	r := csv.NewReader(strings.NewReader(text))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	require.NoError(t, err)
	require.NotEmpty(t, rows)
	return rows[1:]
}

func splitEntries(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(text, ";")
}

func assignment(t testHelper, text string) (string, string) {
	t.Helper()
	left, right, ok := strings.Cut(text, "=")
	require.True(t, ok, text)
	return left, right
}

func alterationValue(t testHelper, text string) alteration.Alteration {
	t.Helper()
	numText, denText, found := strings.Cut(text, "/")
	num, err := strconv.Atoi(numText)
	require.NoError(t, err, text)
	den := 1
	if found {
		den, err = strconv.Atoi(denText)
		require.NoError(t, err, text)
	}
	return alteration.MustFromFraction(num, den)
}

func symbolText(t testHelper, text string) string {
	t.Helper()
	if hex, ok := strings.CutPrefix(text, "U+"); ok {
		value, err := strconv.ParseInt(hex, 16, 32)
		require.NoError(t, err, text)
		return string(rune(value))
	}
	return text
}

func letterByName(t testHelper, name string) letter.Letter {
	t.Helper()
	for _, l := range letter.All() {
		if l.String() == name {
			return l
		}
	}
	t.Fatalf("unknown letter %q", name)
	return 0
}

func isNameKey(text string) bool {
	return len(text) > 0 && text[0] >= 'A' && text[0] <= 'G'
}

func nameByParts(t testHelper, text string) spelling.Spelling {
	t.Helper()
	require.NotEmpty(t, text)
	return spelling.MustNew(letterByName(t, text[:1]), alterationValue(t, text[1:]))
}

func nameWithAlteration(l letter.Letter, num, den int) spelling.Spelling {
	return spelling.MustNew(l, alteration.MustFromFraction(num, den))
}

func composeName(letterText, alterationText string, position notation.AlterationPosition) string {
	if position == notation.AlterationPrefix {
		return alterationText + letterText
	}
	return letterText + alterationText
}

func expectedBuiltIns(t testHelper) []expectedNotation {
	t.Helper()
	rows := readFixture(t, builtInRulesFixture)
	result := make([]expectedNotation, 0, len(rows)-1)
	for _, row := range rows {
		id := row[0]
		if id == notation.ASCIIUpDownID {
			continue
		}
		letters := strings.Fields(row[1])
		require.Len(t, letters, letter.Count, id)
		position := notation.AlterationSuffix
		if row[2] == "префикс" {
			position = notation.AlterationPrefix
		}
		e := expectedNotation{
			id:      id,
			forms:   make(map[spelling.Spelling]string),
			aliases: make(map[string]spelling.Spelling),
		}
		switch row[6] {
		case "aspn":
			e.style = notation.OctaveStyleNumber
		case "am":
			e.style = notation.OctaveStyleAM
		case "abc":
			e.style = notation.OctaveStyleABC
		case "lilypond":
			e.style = notation.OctaveStyleLilyPond
		case "нет":
			e.style = notation.OctaveStyleNone
		default:
			t.Fatalf("unknown octave style %q for %s", row[6], id)
		}

		type altForm struct {
			value alteration.Alteration
			text  string
		}
		altForms := make([]altForm, 0)
		for _, entry := range splitEntries(row[3]) {
			valueText, formText := assignment(t, entry)
			altForms = append(altForms, altForm{
				value: alterationValue(t, valueText),
				text:  symbolText(t, formText),
			})
		}
		canonicalExceptions := make(map[spelling.Spelling]string)
		aliasExceptions := make(map[spelling.Spelling][]string)
		omitted := make(map[spelling.Spelling]bool)
		for _, entry := range splitEntries(row[4]) {
			nameText, formText := assignment(t, entry)
			name := nameByParts(t, nameText)
			if formText == "!" {
				omitted[name] = true
				continue
			}
			canonicalExceptions[name] = symbolText(t, formText)
		}
		for _, entry := range splitEntries(row[5]) {
			nameText, formText := assignment(t, entry)
			for _, alias := range strings.Split(formText, "|") {
				if isNameKey(nameText) {
					name := nameByParts(t, nameText)
					aliasExceptions[name] = append(aliasExceptions[name], symbolText(t, alias))
				}
			}
		}

		for i, letterText := range letters {
			l := letter.MustFromOrdinal(i)
			for _, form := range altForms {
				name := spelling.MustNew(l, form.value)
				if omitted[name] {
					continue
				}
				text := composeName(letterText, form.text, position)
				if exception, ok := canonicalExceptions[name]; ok {
					text = exception
				}
				e.forms[name] = text
				e.aliases[text] = name
			}
		}
		for _, entry := range splitEntries(row[5]) {
			valueText, formsText := assignment(t, entry)
			if isNameKey(valueText) {
				continue
			}
			value := alterationValue(t, valueText)
			for _, alias := range strings.Split(formsText, "|") {
				for i, letterText := range letters {
					name := spelling.MustNew(letter.MustFromOrdinal(i), value)
					if _, ok := e.forms[name]; ok {
						e.aliases[composeName(letterText, symbolText(t, alias), position)] = name
					}
				}
			}
		}
		for name, aliases := range aliasExceptions {
			if _, ok := e.forms[name]; !ok {
				continue
			}
			for _, alias := range aliases {
				e.aliases[alias] = name
			}
		}
		result = append(result, e)
	}
	return result
}

func expectedUpDown(t testHelper, c int) expectedNotation {
	t.Helper()
	e := expectedNotation{
		id:            notation.ASCIIUpDownID,
		style:         notation.OctaveStyleNone,
		forms:         make(map[spelling.Spelling]string),
		aliases:       make(map[string]spelling.Spelling),
		chromaticUnit: c,
	}
	ordinary := []struct {
		value int
		text  string
	}{
		{-2, "𝄫"},
		{-1, "♭"},
		{0, ""},
		{1, "♯"},
		{2, "𝄪"},
	}
	arrows := []struct {
		value int
		text  string
	}{
		{-1, "v"},
		{0, ""},
		{1, "^"},
	}
	for _, l := range letter.All() {
		letterText := l.String()
		for _, base := range ordinary {
			for _, arrow := range arrows {
				step := alteration.MustFromFraction(arrow.value, c)
				value := alteration.MustFromInt(base.value).MustAdd(step)
				name := spelling.MustNew(l, value)
				text := arrow.text + letterText + base.text
				e.forms[name] = text
				e.aliases[text] = name
			}
		}
		natural := spelling.MustNew(l, alteration.Natural())
		e.aliases[letterText+"♮"] = natural
	}
	return e
}

func explicitNameRules(t testHelper, forms ...notation.NameForm) notation.NameRules {
	t.Helper()
	rules, err := notation.NewNameRules(notation.NameRulesDefinition{Names: forms})
	require.NoError(t, err)
	return rules
}

func generatedNameRules(t testHelper, def notation.NameRulesDefinition) notation.NameRules {
	t.Helper()
	rules, err := notation.NewNameRules(def)
	require.NoError(t, err)
	return rules
}

func customNotation(t testHelper, rules notation.NameRules, style notation.OctaveStyle, separator string) notation.Notation {
	t.Helper()
	n, err := notation.New(notation.Definition{
		Rules:           rules,
		OctaveStyle:     style,
		OctaveSeparator: separator,
	})
	require.NoError(t, err)
	return n
}

func requirePanicError(t testHelper, target error, fn func()) {
	t.Helper()
	defer func() {
		recovered := recover()
		require.NotNil(t, recovered)
		err, ok := recovered.(error)
		require.True(t, ok, recovered)
		require.ErrorIs(t, err, target)
	}()
	fn()
}

func readProductFile(t testHelper, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(name)
	require.NoError(t, err)
	return data
}

func changeFirstCase(text string, upper bool) string {
	for index := 0; index < len(text); {
		r, size := utf8.DecodeRuneInString(text[index:])
		changed := unicode.ToLower(r)
		if upper {
			changed = unicode.ToUpper(r)
		}
		if changed != r {
			return text[:index] + string(changed) + text[index+size:]
		}
		index += size
	}
	return text
}

func repeatFirstRegisterLetter(text string, count int) string {
	upper := changeFirstCase(text, true)
	for index := 0; index < len(upper); {
		r, size := utf8.DecodeRuneInString(upper[index:])
		if unicode.ToUpper(r) != unicode.ToLower(r) {
			return upper[:index] + strings.Repeat(string(r), count) + upper[index+size:]
		}
		index += size
	}
	return upper
}

func expectedSpecialOctave(style notation.OctaveStyle, name string, w int) string {
	switch style {
	case notation.OctaveStyleAM:
		switch w {
		case 0:
			return changeFirstCase(name, true) + "₂"
		case 1:
			return changeFirstCase(name, true) + "₁"
		case 2:
			return changeFirstCase(name, true)
		case 3:
			return changeFirstCase(name, false)
		default:
			return changeFirstCase(name, false) + []string{"¹", "²", "³", "⁴", "⁵"}[w-4]
		}
	case notation.OctaveStyleHelmholtz:
		if w <= 2 {
			return changeFirstCase(name, true) + strings.Repeat("͵", 2-w)
		}
		return changeFirstCase(name, false) + strings.Repeat("′", w-3)
	case notation.OctaveStyleHelmholtzEnglish:
		switch {
		case w == 0:
			return repeatFirstRegisterLetter(name, 3)
		case w == 1:
			return repeatFirstRegisterLetter(name, 2)
		case w == 2:
			return changeFirstCase(name, true)
		case w < 0:
			return changeFirstCase(name, true) + strings.Repeat("͵", 2-w)
		default:
			return changeFirstCase(name, false) + strings.Repeat("′", w-3)
		}
	case notation.OctaveStyleHelmholtzNumbered:
		switch {
		case w < 2:
			return fmt.Sprintf("%s%d", changeFirstCase(name, true), 2-w)
		case w == 2:
			return changeFirstCase(name, true)
		case w == 3:
			return changeFirstCase(name, false)
		default:
			return fmt.Sprintf("%s%d", changeFirstCase(name, false), w-3)
		}
	case notation.OctaveStyleLilyPond:
		name = changeFirstCase(name, false)
		if w < 3 {
			return name + strings.Repeat(",", 3-w)
		}
		return name + strings.Repeat("'", w-3)
	case notation.OctaveStyleABC:
		switch {
		case w < 4:
			return changeFirstCase(name, true) + strings.Repeat(",", 4-w)
		case w == 4:
			return changeFirstCase(name, true)
		default:
			return changeFirstCase(name, false) + strings.Repeat("'", w-5)
		}
	default:
		panic("not a special octave style")
	}
}
