package template

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// Entry is an immutable catalog record: an exact code, at least one name in a
// fixed order, and a set Template. The zero Entry is unset. Copies may share
// immutable internal data; returned name slices are independent copies.
type Entry struct {
	code       string
	names      []string
	normalized []string
	tmpl       Template
}

var (
	nameFolder = cases.Fold()
	nameSigns  = strings.NewReplacer("ё", "е", "♭", "b", "♯", "#", "𝄫", "bb", "𝄪", "##")
)

// NewEntry creates an entry. Codes are compared exactly and contain no
// whitespace. Names retain their original spelling, must be nonempty and
// trimmed, and must be distinct after name normalization. The names slice is
// copied.
func NewEntry(code string, names []string, tmpl Template) (Entry, error) {
	if !tmpl.IsValid() {
		return Entry{}, opError("NewEntry", ErrUnset)
	}
	if code == "" || strings.IndexFunc(code, unicode.IsSpace) >= 0 {
		return Entry{}, opError("NewEntry", ErrInvalidArgument)
	}
	if len(names) == 0 {
		return Entry{}, opError("NewEntry", ErrInvalidArgument)
	}
	original := append([]string{}, names...)
	normalized := make([]string, len(names))
	seen := make(map[string]struct{}, len(names))
	for i, name := range names {
		if name == "" || strings.TrimSpace(name) != name {
			return Entry{}, opError("NewEntry", ErrInvalidArgument)
		}
		normalized[i] = normalizeName(name)
		if _, exists := seen[normalized[i]]; exists {
			return Entry{}, opError("NewEntry", ErrInvalidArgument)
		}
		seen[normalized[i]] = struct{}{}
	}
	return Entry{code: code, names: original, normalized: normalized, tmpl: tmpl}, nil
}

// MustNewEntry is NewEntry without the error result. It panics with the same
// error.
func MustNewEntry(code string, names []string, tmpl Template) Entry {
	return must(NewEntry(code, names, tmpl))
}

// IsValid reports whether e is a set entry.
func (e Entry) IsValid() bool {
	return e.code != "" && len(e.names) > 0 && e.tmpl.IsValid()
}

// Code returns the exact entry code, or an empty string for an unset entry.
func (e Entry) Code() string {
	return e.code
}

// Template returns the entry template, or the zero Template for an unset entry.
func (e Entry) Template() Template {
	return e.tmpl
}

// Names returns a new slice of all original names in their stored order. It
// returns nil for an unset entry.
func (e Entry) Names() []string {
	if !e.IsValid() {
		return nil
	}
	return append([]string{}, e.names...)
}

// RussianNames returns a new slice of names containing a Cyrillic letter, in
// their original order. Other names are classified as English by this package.
func (e Entry) RussianNames() []string {
	return e.namesByLanguage(true)
}

// EnglishNames returns a new slice of names without a Cyrillic letter, in their
// original order.
func (e Entry) EnglishNames() []string {
	return e.namesByLanguage(false)
}

func (e Entry) namesByLanguage(russian bool) []string {
	if !e.IsValid() {
		return nil
	}
	result := make([]string, 0)
	for _, name := range e.names {
		isRussian := strings.IndexFunc(name, func(r rune) bool {
			return unicode.Is(unicode.Cyrillic, r) && unicode.IsLetter(r)
		}) >= 0
		if isRussian == russian {
			result = append(result, name)
		}
	}
	return result
}

// RussianName returns the first name containing a Cyrillic letter.
func (e Entry) RussianName() (string, bool) {
	return e.firstName(true)
}

// EnglishName returns the first name without a Cyrillic letter.
func (e Entry) EnglishName() (string, bool) {
	return e.firstName(false)
}

func (e Entry) firstName(russian bool) (string, bool) {
	for _, name := range e.namesByLanguage(russian) {
		return name, true
	}
	return "", false
}

// Equal reports whether both entries are set and have equal codes, original
// names in the same order, and templates. An unset entry is equal to no entry.
func (e Entry) Equal(other Entry) bool {
	if !e.IsValid() || !other.IsValid() || e.code != other.code || len(e.names) != len(other.names) || !e.tmpl.Equal(other.tmpl) {
		return false
	}
	for i, name := range e.names {
		if name != other.names[i] {
			return false
		}
	}
	return true
}

// String returns a debug representation. An unset entry returns
// "Entry(unset)".
func (e Entry) String() string {
	if !e.IsValid() {
		return "Entry(unset)"
	}
	return e.code + ": " + e.tmpl.String()
}

// normalizeName computes the comparison key without changing the stored name.
func normalizeName(name string) string {
	folded := nameFolder.String(name)
	folded = norm.NFC.String(folded)
	folded = nameSigns.Replace(folded)
	folded = strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Pd, r) || r == '\u2212' {
			return '-'
		}
		return r
	}, folded)
	decomposed := norm.NFD.String(folded)
	var result strings.Builder
	latinStarter := false
	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) {
			if latinStarter {
				continue
			}
		} else {
			latinStarter = unicode.Is(unicode.Latin, r)
		}
		result.WriteRune(r)
	}
	return strings.Join(strings.Fields(result.String()), " ")
}
