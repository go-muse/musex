package notation

import (
	"embed"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"
)

const (
	// EnglishUnicodeID identifies English names with Unicode accidentals.
	EnglishUnicodeID = "english-unicode"
	// EnglishASCIIMT21ID identifies English ASCII names in the MT21 domain.
	EnglishASCIIMT21ID = "english-ascii-mt21"
	// GermanAMID identifies German names with AM octave marks.
	GermanAMID = "german-am"
	// RussianAMID identifies Russian names with AM octave marks.
	RussianAMID = "russian-am"
	// RomanceItalianID identifies Italian syllabic names without octaves.
	RomanceItalianID = "romance-italian"
	// ABCID identifies ABC names and octave marks.
	ABCID = "abc"
	// LilyPondNederlandsID identifies Dutch LilyPond names.
	LilyPondNederlandsID = "lilypond-nederlands"
	// LilyPondEnglishID identifies English LilyPond names.
	LilyPondEnglishID = "lilypond-english"
	// UnicodeSteinID identifies Stein-Zimmermann Unicode accidentals.
	UnicodeSteinID = "unicode-stein"
	// UnicodeGouldID identifies Gould Unicode accidentals.
	UnicodeGouldID = "unicode-gould"
	// LilyPondItalianoID identifies Italian LilyPond names.
	LilyPondItalianoID = "lilypond-italiano"
	// UnicodeSimsID identifies Sims Unicode accidentals.
	UnicodeSimsID = "unicode-sims"
	// UnicodeWyschnegradskyID identifies Wyschnegradsky Unicode accidentals.
	UnicodeWyschnegradskyID = "unicode-wyschnegradsky"
	// UnicodeAEUID identifies Arel-Ezgi-Uzdilek Unicode accidentals.
	UnicodeAEUID = "unicode-aeu"
	// UnicodePersianID identifies Persian koron and sori forms.
	UnicodePersianID = "unicode-persian"
	// UnicodeArabicID identifies Arabic Unicode accidentals.
	UnicodeArabicID = "unicode-arabic"
	// UnicodeStockhausenID identifies Stockhausen Unicode accidentals.
	UnicodeStockhausenID = "unicode-stockhausen"
	// ASCIIUpDownID identifies the parameterized ups-and-downs family.
	ASCIIUpDownID = "ascii-updown"
	// ASCIIFokkerID identifies Fokker's ASCII accidental forms.
	ASCIIFokkerID = "ascii-fokker"
)

var builtInOrder = [...]string{
	EnglishUnicodeID,
	EnglishASCIIMT21ID,
	GermanAMID,
	RussianAMID,
	RomanceItalianID,
	ABCID,
	LilyPondNederlandsID,
	LilyPondEnglishID,
	UnicodeSteinID,
	UnicodeGouldID,
	LilyPondItalianoID,
	UnicodeSimsID,
	UnicodeWyschnegradskyID,
	UnicodeAEUID,
	UnicodePersianID,
	UnicodeArabicID,
	UnicodeStockhausenID,
	ASCIIUpDownID,
	ASCIIFokkerID,
}

//go:embed data/*.csv
var builtInCSV embed.FS

var builtIns map[string]Notation

func init() {
	content, err := builtInCSV.ReadFile("data/notation-built-in-rules.csv")
	if err != nil {
		panic(err)
	}
	reader := csv.NewReader(strings.NewReader(string(content)))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}
	builtIns = make(map[string]Notation, len(rows)-2)
	for _, row := range rows[1:] {
		if row[0] == ASCIIUpDownID {
			continue
		}
		n, err := builtInFromRow(row)
		if err != nil {
			panic(fmt.Errorf("notation: built-in %s: %w", row[0], err))
		}
		builtIns[row[0]] = n
	}
}

// BuiltIn returns one of the 18 non-parameterized built-in notations, or an
// ups-and-downs notation when called with ASCIIUpDownID and exactly one
// chromatic unit in 3..65536. Non-parameterized lookup allocates no memory.
func BuiltIn(id string, chromaticUnit ...int) (Notation, error) {
	if len(chromaticUnit) > 1 {
		return Notation{}, fmt.Errorf("%w: more than one argument", ErrInvalidChromaticUnit)
	}
	if id == ASCIIUpDownID {
		if len(chromaticUnit) == 0 {
			return Notation{}, fmt.Errorf("%w: %s", ErrMissingChromaticUnit, id)
		}
		rules, err := NewUpsAndDownsNameRules(chromaticUnit[0])
		if err != nil {
			return Notation{}, err
		}
		return New(Definition{Name: id, Rules: rules, OctaveStyle: OctaveStyleNone})
	}
	if len(chromaticUnit) != 0 {
		return Notation{}, fmt.Errorf("%w: unexpected argument for %s", ErrInvalidChromaticUnit, id)
	}
	n, ok := builtIns[id]
	if !ok {
		return Notation{}, fmt.Errorf("%w: %q", ErrUnknownID, id)
	}
	return n, nil
}

// MustBuiltIn is like BuiltIn but panics with its error on failure.
func MustBuiltIn(id string, chromaticUnit ...int) Notation {
	return must(BuiltIn(id, chromaticUnit...))
}

// BuiltInIDs returns the 19 identifiers in the order of the built-in rules
// table. The caller owns the new slice and may change it freely.
func BuiltInIDs() []string {
	ids := make([]string, len(builtInOrder))
	copy(ids, builtInOrder[:])
	return ids
}

func builtInFromRow(row []string) (Notation, error) {
	if len(row) < 12 {
		return Notation{}, fmt.Errorf("%w: built-in row has %d fields", ErrInvalidDefinition, len(row))
	}
	letters := strings.Fields(row[1])
	if len(letters) != letter.Count {
		return Notation{}, fmt.Errorf("%w: built-in letters", ErrInvalidDefinition)
	}
	var def NameRulesDefinition
	copy(def.Letters[:], letters)
	switch row[2] {
	case "суффикс":
		def.Position = AlterationSuffix
	case "префикс":
		def.Position = AlterationPrefix
	default:
		return Notation{}, fmt.Errorf("%w: built-in alteration position", ErrInvalidDefinition)
	}

	alterationForms := make(map[alteration.Alteration]AlterationForm)
	for _, item := range splitList(row[3]) {
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			return Notation{}, fmt.Errorf("%w: built-in alteration %q", ErrInvalidDefinition, item)
		}
		a, err := alterationFromText(key)
		if err != nil {
			return Notation{}, fmt.Errorf("%w: built-in alteration %q", ErrInvalidDefinition, key)
		}
		text, err := decodeSymbol(value)
		if err != nil {
			return Notation{}, err
		}
		alterationForms[a] = AlterationForm{Value: a, Text: text}
		def.Alterations = append(def.Alterations, alterationForms[a])
	}

	wholeAliases := make(map[spelling.Spelling][]string)
	for _, item := range splitList(row[5]) {
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			return Notation{}, fmt.Errorf("%w: built-in alias %q", ErrInvalidDefinition, item)
		}
		aliases, err := decodeSymbols(strings.Split(value, "|"))
		if err != nil {
			return Notation{}, err
		}
		if name, ok, err := nameFromKey(key); err != nil {
			return Notation{}, err
		} else if ok {
			wholeAliases[name] = append(wholeAliases[name], aliases...)
			continue
		}
		a, err := alterationFromText(key)
		if err != nil {
			return Notation{}, fmt.Errorf("%w: built-in alias value %q", ErrInvalidDefinition, key)
		}
		form := alterationForms[a]
		form.Aliases = append(form.Aliases, aliases...)
		alterationForms[a] = form
	}
	for i := range def.Alterations {
		def.Alterations[i] = alterationForms[def.Alterations[i].Value]
	}

	canonical := make(map[spelling.Spelling]string)
	omitted := make(map[spelling.Spelling]struct{})
	for _, item := range splitList(row[4]) {
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			return Notation{}, fmt.Errorf("%w: built-in exception %q", ErrInvalidDefinition, item)
		}
		name, nameOK, err := nameFromKey(key)
		if err != nil || !nameOK {
			return Notation{}, fmt.Errorf("%w: built-in exception name %q", ErrInvalidDefinition, key)
		}
		if value == "!" {
			omitted[name] = struct{}{}
			def.Omit = append(def.Omit, name)
			continue
		}
		text, err := decodeSymbol(value)
		if err != nil {
			return Notation{}, err
		}
		canonical[name] = text
	}

	names := make(map[spelling.Spelling]struct{}, len(canonical)+len(wholeAliases))
	for name := range canonical {
		names[name] = struct{}{}
	}
	for name := range wholeAliases {
		names[name] = struct{}{}
	}
	for name := range names {
		a := name.Alteration()
		form := alterationForms[a]
		text := def.Letters[name.Letter().Ordinal()] + form.Text
		if def.Position == AlterationPrefix {
			text = form.Text + def.Letters[name.Letter().Ordinal()]
		}
		aliases := make([]string, len(form.Aliases))
		for i, alias := range form.Aliases {
			aliases[i] = def.Letters[name.Letter().Ordinal()] + alias
			if def.Position == AlterationPrefix {
				aliases[i] = alias + def.Letters[name.Letter().Ordinal()]
			}
		}
		if replacement, ok := canonical[name]; ok {
			text = replacement
			aliases = nil
		}
		seen := make(map[string]struct{}, len(aliases)+len(wholeAliases[name]))
		for _, alias := range aliases {
			seen[alias] = struct{}{}
		}
		for _, alias := range wholeAliases[name] {
			if alias == text {
				continue
			}
			if _, exists := seen[alias]; exists {
				continue
			}
			seen[alias] = struct{}{}
			aliases = append(aliases, alias)
		}
		def.Names = append(def.Names, NameForm{Name: name, Text: text, Aliases: aliases})
	}

	rules, err := NewNameRules(def)
	if err != nil {
		return Notation{}, err
	}
	style, err := builtInOctaveStyle(row[6])
	if err != nil {
		return Notation{}, err
	}
	return New(Definition{Name: row[0], Rules: rules, OctaveStyle: style})
}

func builtInOctaveStyle(text string) (OctaveStyle, error) {
	switch text {
	case "aspn":
		return OctaveStyleNumber, nil
	case "am":
		return OctaveStyleAM, nil
	case "abc":
		return OctaveStyleABC, nil
	case "lilypond":
		return OctaveStyleLilyPond, nil
	case "нет":
		return OctaveStyleNone, nil
	default:
		return OctaveStyleUnset, fmt.Errorf("%w: built-in octave style %q", ErrInvalidDefinition, text)
	}
}

func splitList(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(text, ";")
}

func nameFromKey(key string) (spelling.Spelling, bool, error) {
	if len(key) < 2 {
		return spelling.Spelling{}, false, nil
	}
	for _, l := range letter.All() {
		if key[0] == l.String()[0] {
			a, err := alterationFromText(key[1:])
			if err != nil {
				return spelling.Spelling{}, false, err
			}
			return spelling.MustNew(l, a), true, nil
		}
	}
	return spelling.Spelling{}, false, nil
}

func decodeSymbols(values []string) ([]string, error) {
	result := make([]string, len(values))
	for i, value := range values {
		decoded, err := decodeSymbol(value)
		if err != nil {
			return nil, err
		}
		result[i] = decoded
	}
	return result, nil
}

func decodeSymbol(value string) (string, error) {
	if !strings.HasPrefix(value, "U+") {
		return value, nil
	}
	code, err := strconv.ParseUint(strings.TrimPrefix(value, "U+"), 16, 32)
	if err != nil {
		return "", fmt.Errorf("%w: Unicode code %q", ErrInvalidDefinition, value)
	}
	return string(rune(code)), nil
}
