package notation

import (
	"fmt"
	"strconv"

	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/spelling"
)

// OctaveStyle selects how a written octave is shown next to a note name.
type OctaveStyle uint8

const (
	// OctaveStyleUnset is the zero, unspecified style.
	OctaveStyleUnset OctaveStyle = iota
	// OctaveStyleNone defines a name-only notation.
	OctaveStyleNone
	// OctaveStyleNumber appends a visible number in an explicit numbering.
	OctaveStyleNumber
	// OctaveStyleName appends a visible name from an explicit catalog.
	OctaveStyleName
	// OctaveStyleAM uses the AM case and superscript/subscript convention.
	OctaveStyleAM
	// OctaveStyleHelmholtz uses Helmholtz case and prime marks.
	OctaveStyleHelmholtz
	// OctaveStyleHelmholtzEnglish repeats the first cased letter for ASPN 0 and 1.
	OctaveStyleHelmholtzEnglish
	// OctaveStyleHelmholtzNumbered uses the numbered Helmholtz variant.
	OctaveStyleHelmholtzNumbered
	// OctaveStyleLilyPond uses LilyPond commas and apostrophes.
	OctaveStyleLilyPond
	// OctaveStyleABC uses ABC case, commas, and apostrophes.
	OctaveStyleABC
)

// IsValid reports whether s is one of the nine defined styles. It returns
// false for OctaveStyleUnset and every unlisted value.
func (s OctaveStyle) IsValid() bool {
	return s >= OctaveStyleNone && s <= OctaveStyleABC
}

// String returns the debugging identifier of s, not a notation or serialized
// form. Unknown values use their unsigned decimal representation.
func (s OctaveStyle) String() string {
	switch s {
	case OctaveStyleNone:
		return "none"
	case OctaveStyleNumber:
		return "number"
	case OctaveStyleName:
		return "name"
	case OctaveStyleAM:
		return "am"
	case OctaveStyleHelmholtz:
		return "helmholtz"
	case OctaveStyleHelmholtzEnglish:
		return "helmholtz-english"
	case OctaveStyleHelmholtzNumbered:
		return "helmholtz-numbered"
	case OctaveStyleLilyPond:
		return "lilypond"
	case OctaveStyleABC:
		return "abc"
	case OctaveStyleUnset:
		return "OctaveStyle(unset)"
	default:
		return "OctaveStyle(" + strconv.Itoa(int(uint8(s))) + ")"
	}
}

// Definition defines a notation from name rules, an octave style, and the
// separator used by number and name styles. Name is only a debugging label.
type Definition struct {
	// Name is an optional debugging label, ignored by Equal.
	Name string
	// Rules supplies the finite name domain and its canonical forms.
	Rules NameRules
	// OctaveStyle selects the written-octave grammar.
	OctaveStyle OctaveStyle
	// OctaveSeparator joins the name to a visible number or octave name.
	OctaveSeparator string
}

// Notation is an immutable rule for formatting and parsing names and written
// pitches. The zero value is undefined and has no implicit built-in behavior.
//
// Notation is not comparable; use Equal. Copies share only immutable data and
// are safe for concurrent use.
type Notation struct {
	name      string
	rules     NameRules
	style     OctaveStyle
	separator string
	_         []struct{}
}

// New validates and creates a notation. It rejects structurally invalid rules
// and any ambiguity between name forms, aliases, separators, and octave marks.
func New(def Definition) (Notation, error) {
	if !def.Rules.IsDefined() || def.OctaveStyle == OctaveStyleUnset {
		return Notation{}, fmt.Errorf("%w: new notation", ErrUndefined)
	}
	if !def.OctaveStyle.IsValid() {
		return Notation{}, fmt.Errorf("%w: octave style %d", ErrInvalidDefinition, def.OctaveStyle)
	}
	if def.OctaveStyle == OctaveStyleNone || def.OctaveStyle == OctaveStyleNumber || def.OctaveStyle == OctaveStyleName {
		if def.OctaveStyle == OctaveStyleNone && def.OctaveSeparator != "" {
			return Notation{}, fmt.Errorf("%w: separator for name-only notation", ErrInvalidDefinition)
		}
	} else if def.OctaveSeparator != "" {
		return Notation{}, fmt.Errorf("%w: separator for %s", ErrInvalidDefinition, def.OctaveStyle)
	}
	if err := validateNotation(def.Rules, def.OctaveStyle, def.OctaveSeparator); err != nil {
		return Notation{}, err
	}
	return Notation{name: def.Name, rules: def.Rules, style: def.OctaveStyle, separator: def.OctaveSeparator}, nil
}

// MustNew is like New but panics with its error on failure.
func MustNew(def Definition) Notation {
	return must(New(def))
}

// IsDefined reports whether n is defined. It is false only for the zero value.
func (n Notation) IsDefined() bool {
	return n.rules.IsDefined() && n.style.IsValid()
}

// Equal reports whether both notations are defined and have equal name rules,
// octave style, separator, and chromatic unit. The debugging name is ignored.
func (n Notation) Equal(other Notation) bool {
	return n.IsDefined() && other.IsDefined() && n.rules.Equal(other.rules) &&
		n.style == other.style && n.separator == other.separator
}

// Name returns the optional debugging name. It does not affect behavior or
// equality and is not a machine-readable serialization.
func (n Notation) Name() string {
	if !n.IsDefined() {
		return ""
	}
	return n.name
}

// NameRules returns the immutable name rules, or zero rules for an undefined
// notation.
func (n Notation) NameRules() NameRules {
	if !n.IsDefined() {
		return NameRules{}
	}
	return n.rules
}

// OctaveStyle returns the octave style, or OctaveStyleUnset for an undefined
// notation.
func (n Notation) OctaveStyle() OctaveStyle {
	if !n.IsDefined() {
		return OctaveStyleUnset
	}
	return n.style
}

// OctaveSeparator returns the separator used by number and name styles.
func (n Notation) OctaveSeparator() string {
	if !n.IsDefined() {
		return ""
	}
	return n.separator
}

// ChromaticUnit returns the fixed chromatic unit of an ups-and-downs notation.
func (n Notation) ChromaticUnit() (int, bool) {
	if !n.IsDefined() {
		return 0, false
	}
	return n.rules.ChromaticUnit()
}

// Format returns the canonical text of name without an octave. No musical
// system, numbering, or naming catalog is needed.
func (n Notation) Format(name spelling.Spelling) (string, error) {
	if !n.IsDefined() {
		return "", fmt.Errorf("%w: format name", ErrUndefined)
	}
	return n.rules.Format(name)
}

// MustFormat is like Notation.Format but panics with its error on failure.
func (n Notation) MustFormat(name spelling.Spelling) string {
	return must(n.Format(name))
}

// Parse parses a canonical name or one of its explicit aliases.
func (n Notation) Parse(text string) (spelling.Spelling, error) {
	if !n.IsDefined() {
		return spelling.Spelling{}, fmt.Errorf("%w: parse name", ErrUndefined)
	}
	return n.rules.Parse(text)
}

// MustParse is like Notation.Parse but panics with its error on failure.
func (n Notation) MustParse(text string) spelling.Spelling {
	return must(n.Parse(text))
}

// Convert parses text with from and returns its canonical form in to.
func Convert(text string, from, to Notation) (string, error) {
	name, err := from.Parse(text)
	if err != nil {
		return "", err
	}
	return to.Format(name)
}

// MustConvert is like Convert but panics with its error on failure.
func MustConvert(text string, from, to Notation) string {
	return must(Convert(text, from, to))
}

// OctaveContext supplies the independently selected numbering or naming
// catalog required by a pitch operation. A nil catalog is a valid empty catalog.
type OctaveContext struct {
	// Numbering is required for OctaveStyleNumber and is otherwise ignored.
	Numbering number.Numbering
	// Catalog is borrowed for OctaveStyleName. Nil is a valid empty catalog.
	Catalog *naming.Catalog
}
