package notation

import (
	"fmt"
	"slices"
	"strings"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/spelling"
)

// AlterationPosition selects the side of the letter on which a generated
// accidental is written.
type AlterationPosition uint8

const (
	// AlterationPositionUnset is the zero, unspecified position.
	AlterationPositionUnset AlterationPosition = iota
	// AlterationPrefix writes the accidental before the letter.
	AlterationPrefix
	// AlterationSuffix writes the accidental after the letter.
	AlterationSuffix
)

// AlterationForm defines the canonical text and accepted aliases of one
// alteration value in generated name rules.
type AlterationForm struct {
	// Value is the exact alteration represented by the form.
	Value alteration.Alteration
	// Text is the canonical accidental text, empty for an unmarked natural.
	Text string
	// Aliases contains additional accidental texts accepted during parsing.
	Aliases []string
}

// NameForm defines a whole-name canonical form and its accepted aliases.
// In generated rules it replaces both the generated form and its aliases.
type NameForm struct {
	// Name is the exact written name represented by the form.
	Name spelling.Spelling
	// Text is the canonical whole-name output and must be nonempty.
	Text string
	// Aliases contains additional whole-name inputs accepted during parsing.
	Aliases []string
}

// NameRulesDefinition defines a finite set of note-name forms. Rules can be
// generated from letters and alterations or listed explicitly in Names.
type NameRulesDefinition struct {
	// Letters holds letter strings indexed by Letter.Ordinal in C through B order.
	Letters [letter.Count]string
	// Alterations supplies values and accidental forms for generated rules.
	Alterations []AlterationForm
	// Names supplies whole-name exceptions or an explicit complete table.
	Names []NameForm
	// Omit removes generated letter-and-alteration pairs from the domain.
	Omit []spelling.Spelling
	// Position selects prefix or suffix accidental composition.
	Position AlterationPosition
}

type nameEntry struct {
	name    spelling.Spelling
	text    string
	aliases []string
}

type nameRulesData struct {
	entries       []nameEntry
	format        map[spelling.Spelling]string
	parse         map[string]spelling.Spelling
	chromaticUnit int
}

// NameRules is an immutable finite domain of note names with one canonical
// output form and any number of explicit aliases for each name.
//
// The zero value is undefined. NameRules is not comparable; use Equal. Copies
// share only immutable tables and are safe for concurrent use.
type NameRules struct {
	data *nameRulesData
	_    []struct{}
}

// NewNameRules validates and creates name rules from def. It copies every
// input slice and rejects duplicate or ambiguous forms.
func NewNameRules(def NameRulesDefinition) (NameRules, error) {
	hasLetters := false
	for _, text := range def.Letters {
		if text != "" {
			hasLetters = true
			break
		}
	}

	generated := hasLetters || len(def.Alterations) > 0 || def.Position != AlterationPositionUnset
	if !generated {
		if len(def.Names) == 0 || len(def.Omit) != 0 {
			return NameRules{}, fmt.Errorf("%w: explicit rules require names and no omissions", ErrInvalidDefinition)
		}

		entries := make([]nameEntry, len(def.Names))
		for i, form := range def.Names {
			entries[i] = nameEntry{name: form.Name, text: form.Text, aliases: slices.Clone(form.Aliases)}
		}

		return newNameRules(entries, 0)
	}

	for _, text := range def.Letters {
		if text == "" {
			return NameRules{}, fmt.Errorf("%w: all seven letters are required", ErrInvalidDefinition)
		}
	}
	if len(def.Alterations) == 0 || (def.Position != AlterationPrefix && def.Position != AlterationSuffix) {
		return NameRules{}, fmt.Errorf("%w: generated rules require alterations and a valid position", ErrInvalidDefinition)
	}

	alterations := make([]AlterationForm, len(def.Alterations))
	seenValues := make(map[alteration.Alteration]struct{}, len(def.Alterations))
	for i, form := range def.Alterations {
		if _, exists := seenValues[form.Value]; exists {
			return NameRules{}, fmt.Errorf("%w: repeated alteration value", ErrInvalidDefinition)
		}
		seenValues[form.Value] = struct{}{}
		if err := checkAliases(form.Text, form.Aliases, false); err != nil {
			return NameRules{}, err
		}
		alterations[i] = AlterationForm{Value: form.Value, Text: form.Text, Aliases: slices.Clone(form.Aliases)}
	}

	exceptions := make(map[spelling.Spelling]NameForm, len(def.Names))
	for _, form := range def.Names {
		if !form.Name.IsValid() {
			return NameRules{}, fmt.Errorf("%w: unset whole-name exception", ErrInvalidDefinition)
		}
		if _, exists := exceptions[form.Name]; exists {
			return NameRules{}, fmt.Errorf("%w: repeated whole-name exception", ErrInvalidDefinition)
		}
		if err := checkAliases(form.Text, form.Aliases, true); err != nil {
			return NameRules{}, err
		}
		exceptions[form.Name] = NameForm{Name: form.Name, Text: form.Text, Aliases: slices.Clone(form.Aliases)}
	}

	omitted := make(map[spelling.Spelling]struct{}, len(def.Omit))
	for _, name := range def.Omit {
		if !name.IsValid() {
			return NameRules{}, fmt.Errorf("%w: unset omitted name", ErrInvalidDefinition)
		}
		if _, exists := omitted[name]; exists {
			return NameRules{}, fmt.Errorf("%w: repeated omitted name", ErrInvalidDefinition)
		}
		if _, exists := exceptions[name]; exists {
			return NameRules{}, fmt.Errorf("%w: a name cannot be omitted and replaced", ErrInvalidDefinition)
		}
		if _, exists := seenValues[name.Alteration()]; !exists {
			return NameRules{}, fmt.Errorf("%w: omitted name is outside the generated domain", ErrInvalidDefinition)
		}
		omitted[name] = struct{}{}
	}

	entries := make([]nameEntry, 0, letter.Count*len(alterations))
	for _, l := range letter.All() {
		for _, form := range alterations {
			name := spelling.MustNew(l, form.Value)
			if _, skip := omitted[name]; skip {
				continue
			}

			text := def.Letters[l.Ordinal()] + form.Text
			if def.Position == AlterationPrefix {
				text = form.Text + def.Letters[l.Ordinal()]
			}
			aliases := make([]string, len(form.Aliases))
			for i, alias := range form.Aliases {
				aliases[i] = def.Letters[l.Ordinal()] + alias
				if def.Position == AlterationPrefix {
					aliases[i] = alias + def.Letters[l.Ordinal()]
				}
			}
			if exception, exists := exceptions[name]; exists {
				text = exception.Text
				aliases = slices.Clone(exception.Aliases)
			}
			entries = append(entries, nameEntry{name: name, text: text, aliases: aliases})
		}
	}

	return newNameRules(entries, 0)
}

// MustNewNameRules is like NewNameRules but panics with its error on failure.
func MustNewNameRules(def NameRulesDefinition) NameRules {
	return must(NewNameRules(def))
}

// NewUpsAndDownsNameRules creates the 105 standard ups-and-downs forms for a
// chromatic unit c in the range 3..65536. An up arrow adds 1/c and a down
// arrow subtracts 1/c.
func NewUpsAndDownsNameRules(c int) (NameRules, error) {
	if c < 3 || c > 65536 {
		return NameRules{}, fmt.Errorf("%w: chromatic unit %d", ErrInvalidChromaticUnit, c)
	}

	step, err := alteration.FromFraction(1, c)
	if err != nil {
		return NameRules{}, fmt.Errorf("%w: %v", ErrInvalidChromaticUnit, err)
	}

	accidentals := [...]string{"𝄫", "♭", "", "♯", "𝄪"}
	entries := make([]nameEntry, 0, letter.Count*15)
	for _, l := range letter.All() {
		for base := -2; base <= 2; base++ {
			baseValue := alteration.MustFromInt(base)
			for arrow := -1; arrow <= 1; arrow++ {
				value := baseValue
				if arrow != 0 {
					delta := step
					if arrow < 0 {
						delta = alteration.MustFromFraction(-1, c)
					}
					value, err = baseValue.Add(delta)
					if err != nil {
						return NameRules{}, fmt.Errorf("%w: %v", ErrInvalidChromaticUnit, err)
					}
				}

				text := l.String() + accidentals[base+2]
				if arrow > 0 {
					text = "^" + text
				} else if arrow < 0 {
					text = "v" + text
				}

				var aliases []string
				if base == 0 && arrow == 0 {
					aliases = []string{l.String() + "♮"}
				}
				entries = append(entries, nameEntry{
					name:    spelling.MustNew(l, value),
					text:    text,
					aliases: aliases,
				})
			}
		}
	}

	return newNameRules(entries, c)
}

// MustNewUpsAndDownsNameRules is like NewUpsAndDownsNameRules but panics with
// its error on failure.
func MustNewUpsAndDownsNameRules(c int) NameRules {
	return must(NewUpsAndDownsNameRules(c))
}

// IsDefined reports whether r is defined. It is false only for the zero value.
func (r NameRules) IsDefined() bool {
	return r.data != nil
}

// Equal reports whether both rule sets are defined and have the same domain,
// canonical forms, aliases as sets, and chromatic unit. Input order is ignored.
func (r NameRules) Equal(other NameRules) bool {
	if r.data == nil || other.data == nil {
		return false
	}
	if r.data == other.data {
		return true
	}
	if r.data.chromaticUnit != other.data.chromaticUnit || len(r.data.entries) != len(other.data.entries) {
		return false
	}
	for i, entry := range r.data.entries {
		otherEntry := other.data.entries[i]
		if !entry.name.Equal(otherEntry.name) || entry.text != otherEntry.text ||
			!slices.Equal(entry.aliases, otherEntry.aliases) {
			return false
		}
	}
	return true
}

// Names returns the domain in spelling.Compare order. The caller owns the new
// slice. Undefined rules return nil.
func (r NameRules) Names() []spelling.Spelling {
	if r.data == nil {
		return nil
	}
	names := make([]spelling.Spelling, len(r.data.entries))
	for i, entry := range r.data.entries {
		names[i] = entry.name
	}
	return names
}

// ChromaticUnit returns the fixed chromatic unit of ups-and-downs rules. For
// other defined rules and undefined rules it returns (0, false).
func (r NameRules) ChromaticUnit() (int, bool) {
	if r.data == nil || r.data.chromaticUnit == 0 {
		return 0, false
	}
	return r.data.chromaticUnit, true
}

// Format returns the canonical text of name. A valid name outside the domain
// gives ErrUnwritableName; an unset name or rules give ErrUndefined.
func (r NameRules) Format(name spelling.Spelling) (string, error) {
	if r.data == nil || !name.IsValid() {
		return "", fmt.Errorf("%w: format name", ErrUndefined)
	}
	text, ok := r.data.format[name]
	if !ok {
		return "", fmt.Errorf("%w: %v", ErrUnwritableName, name)
	}
	return text, nil
}

// MustFormat is like NameRules.Format but panics with its error on failure.
func (r NameRules) MustFormat(name spelling.Spelling) string {
	return must(r.Format(name))
}

// Parse returns the name represented exactly by a canonical form or alias.
// No whitespace, case, Unicode, or accidental normalization is performed.
func (r NameRules) Parse(text string) (spelling.Spelling, error) {
	if r.data == nil {
		return spelling.Spelling{}, fmt.Errorf("%w: parse name", ErrUndefined)
	}
	name, ok := r.data.parse[text]
	if text == "" || !ok {
		return spelling.Spelling{}, fmt.Errorf("%w: %q", ErrParse, text)
	}
	return name, nil
}

// MustParse is like NameRules.Parse but panics with its error on failure.
func (r NameRules) MustParse(text string) spelling.Spelling {
	return must(r.Parse(text))
}

func newNameRules(entries []nameEntry, chromaticUnit int) (NameRules, error) {
	if len(entries) == 0 {
		return NameRules{}, fmt.Errorf("%w: empty name domain", ErrInvalidDefinition)
	}

	data := &nameRulesData{
		entries:       make([]nameEntry, len(entries)),
		format:        make(map[spelling.Spelling]string, len(entries)),
		parse:         make(map[string]spelling.Spelling, len(entries)),
		chromaticUnit: chromaticUnit,
	}
	copy(data.entries, entries)

	for i := range data.entries {
		entry := &data.entries[i]
		if !entry.name.IsValid() || entry.text == "" {
			return NameRules{}, fmt.Errorf("%w: invalid name or empty canonical form", ErrInvalidDefinition)
		}
		if _, exists := data.format[entry.name]; exists {
			return NameRules{}, fmt.Errorf("%w: repeated name", ErrInvalidDefinition)
		}
		if err := checkAliases(entry.text, entry.aliases, true); err != nil {
			return NameRules{}, err
		}
		slices.Sort(entry.aliases)
		data.format[entry.name] = entry.text
	}

	slices.SortFunc(data.entries, func(a, b nameEntry) int {
		return a.name.Compare(b.name)
	})

	for _, entry := range data.entries {
		texts := make([]string, 0, len(entry.aliases)+1)
		texts = append(texts, entry.text)
		texts = append(texts, entry.aliases...)
		for _, text := range texts {
			if previous, exists := data.parse[text]; exists {
				if previous == entry.name {
					return NameRules{}, fmt.Errorf("%w: repeated form for one name", ErrInvalidDefinition)
				}
				return NameRules{}, fmt.Errorf("%w: %q denotes two names", ErrAmbiguous, text)
			}
			data.parse[text] = entry.name
		}
	}

	return NameRules{data: data}, nil
}

func checkAliases(canonical string, aliases []string, requireCanonical bool) error {
	if requireCanonical && canonical == "" {
		return fmt.Errorf("%w: empty canonical form", ErrInvalidDefinition)
	}
	seen := make(map[string]struct{}, len(aliases)+1)
	if canonical != "" {
		seen[canonical] = struct{}{}
	}
	for _, alias := range aliases {
		if alias == "" {
			return fmt.Errorf("%w: empty alias", ErrInvalidDefinition)
		}
		if _, exists := seen[alias]; exists {
			return fmt.Errorf("%w: repeated alias", ErrInvalidDefinition)
		}
		seen[alias] = struct{}{}
	}
	return nil
}

func allNameForms(rules NameRules) []nameEntry {
	if rules.data == nil {
		return nil
	}
	forms := make([]nameEntry, 0, len(rules.data.parse))
	for _, entry := range rules.data.entries {
		forms = append(forms, nameEntry{name: entry.name, text: entry.text})
		for _, alias := range entry.aliases {
			forms = append(forms, nameEntry{name: entry.name, text: alias})
		}
	}
	return forms
}

func alterationFromText(text string) (alteration.Alteration, error) {
	if numerator, denominator, ok := strings.Cut(text, "/"); ok {
		var n, d int
		if _, err := fmt.Sscanf(numerator, "%d", &n); err != nil {
			return alteration.Alteration{}, err
		}
		if _, err := fmt.Sscanf(denominator, "%d", &d); err != nil {
			return alteration.Alteration{}, err
		}
		return alteration.FromFraction(n, d)
	}
	var n int
	if _, err := fmt.Sscanf(text, "%d", &n); err != nil {
		return alteration.Alteration{}, err
	}
	return alteration.FromInt(n)
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
