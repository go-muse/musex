package settings

import (
	"errors"
	"fmt"
	"math"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// Byte lengths of the repeated octave marks of package notation.
const (
	helmholtzLowMark  = len("͵") // U+0375, below the octave of ASPN 2
	helmholtzHighMark = len("′") // U+2032, above the octave of ASPN 3
	asciiMark         = len(",") // the comma and the apostrophe of LilyPond and ABC
)

// lengthSafeCount is a number of repeated marks that never makes a result
// too long whatever the base name: a name longer than math.MaxInt/2 bytes
// cannot exist in memory. Above it the exact base length is checked.
const lengthSafeCount = math.MaxInt / 2

// Expresses reports whether name is expressible by s: admissible in the
// system of s and in the name domain of its notation.
//
// Errors, in this order: ErrUnset when s is unset; ErrUnsetName when name is
// unset. Every other case is an answer, not an error. Expresses reports true
// exactly when Format succeeds.
func (s *Settings) Expresses(name spelling.Spelling) (bool, error) {
	const op = "expresses"
	if !s.IsValid() {
		return false, unsetError(op, "settings")
	}
	if !name.IsValid() {
		return false, unsetNameError(op, "name")
	}
	return s.expresses(name), nil
}

// MustExpresses is like Expresses but panics with its error on failure.
func (s *Settings) MustExpresses(name spelling.Spelling) bool {
	return must(s.Expresses(name))
}

// ExpressesPitch reports whether the pitch p is expressible by s: its name
// is expressible (see Expresses), the notation of s shows octaves, and the
// written octave w of p in the system of s is writable by the octave style
// of the notation with the numbering and the catalog of s.
//
// Errors, in this order: ErrUnset when s is unset; ErrUnsetName when p is
// unset. Every other case is an answer: a written octave outside int, a
// visible number outside int and a result too long for an int length give
// false, not an error. ExpressesPitch reports true exactly when FormatPitch
// succeeds, except for octaves so far from the middle that FormatPitch would
// build a string that does not fit in memory (see the package documentation).
// ExpressesPitch never builds the formatted string.
func (s *Settings) ExpressesPitch(p pitch.Pitch) (bool, error) {
	const op = "expresses pitch"
	if !s.IsValid() {
		return false, unsetError(op, "settings")
	}
	if !p.IsValid() {
		return false, unsetNameError(op, "pitch")
	}
	return s.expressesPitch(p), nil
}

// MustExpressesPitch is like ExpressesPitch but panics with its error on
// failure.
func (s *Settings) MustExpressesPitch(p pitch.Pitch) bool {
	return must(s.ExpressesPitch(p))
}

// Format returns the text of name in the notation of s, without an octave.
//
// Errors, in this order: ErrUnset; ErrUnsetName; ErrInadmissible (also
// system.ErrInadmissible) when name is not admissible in the system of s;
// the error of notation.Notation.Format (notation.ErrUnwritableName for a
// name outside the domain of the notation).
func (s *Settings) Format(name spelling.Spelling) (string, error) {
	const op = "format"
	if !s.IsValid() {
		return "", unsetError(op, "settings")
	}
	if !name.IsValid() {
		return "", unsetNameError(op, "name")
	}
	sys := s.tuning.System()
	if !admits(sys, name) {
		return "", inadmissibleError(op, name, sys)
	}
	text, err := s.notation.Format(name)
	if err != nil {
		return "", wrapError(op, err)
	}
	return text, nil
}

// MustFormat is like Format but panics with its error on failure.
func (s *Settings) MustFormat(name spelling.Spelling) string {
	return must(s.Format(name))
}

// Parse parses the text of a name in the notation of s.
//
// Errors, in this order: ErrUnset; the error of notation.Notation.Parse
// (notation.ErrParse for text outside the grammar); ErrInadmissible (also
// system.ErrInadmissible) when the parsed name is not admissible in the
// system of s. A parsed name is expressible by s.
func (s *Settings) Parse(text string) (spelling.Spelling, error) {
	const op = "parse"
	if !s.IsValid() {
		return spelling.Spelling{}, unsetError(op, "settings")
	}
	name, err := s.notation.Parse(text)
	if err != nil {
		return spelling.Spelling{}, wrapError(op, err)
	}
	sys := s.tuning.System()
	if !admits(sys, name) {
		return spelling.Spelling{}, inadmissibleError(op, name, sys)
	}
	return name, nil
}

// MustParse is like Parse but panics with its error on failure.
func (s *Settings) MustParse(text string) spelling.Spelling {
	return must(s.Parse(text))
}

// FormatPitch returns the text of the pitch p by the notation, the system,
// the numbering and the catalog of s.
//
// Errors, in this order: ErrUnset; ErrUnsetName; ErrInadmissible (also
// system.ErrInadmissible) when the name of p is not admissible in the system
// of s; the errors of notation.Notation.FormatPitch and its dependencies in
// the order of that contract, for example notation.ErrNoOctave for a
// notation without octaves, an AM octave outside 0..8 or an unnamed octave,
// notation.ErrUnwritableName for a name outside the domain,
// system.ErrOutOfRange for a written octave outside int, and
// number.ErrOutOfRange for a visible number outside int.
//
// For octaves extremely far from the middle, a notation with repeated marks
// can need a result string that does not fit in memory; FormatPitch then
// panics or exhausts memory inside package notation (see the package
// documentation). Use ExpressesPitch to test such pitches.
func (s *Settings) FormatPitch(p pitch.Pitch) (string, error) {
	const op = "format pitch"
	if !s.IsValid() {
		return "", unsetError(op, "settings")
	}
	if !p.IsValid() {
		return "", unsetNameError(op, "pitch")
	}
	sys := s.tuning.System()
	if !admits(sys, p.Spelling()) {
		return "", inadmissibleError(op, p.Spelling(), sys)
	}
	text, err := s.notation.FormatPitch(p, sys, s.context())
	if err != nil {
		return "", wrapError(op, err)
	}
	return text, nil
}

// MustFormatPitch is like FormatPitch but panics with its error on failure.
func (s *Settings) MustFormatPitch(p pitch.Pitch) string {
	return must(s.FormatPitch(p))
}

// ParsePitch parses the text of a pitch by the notation, the system, the
// numbering and the catalog of s.
//
// Errors, in this order: ErrUnset; the errors of notation.Notation.ParsePitch
// and its dependencies in the order of that contract. The written octave is
// recovered before the system checks the name, so for text with both a wrong
// octave mark and an inadmissible name the error names the octave. When the
// system reports an inadmissible name, the error also matches
// ErrInadmissible. A parsed pitch is expressible by s.
func (s *Settings) ParsePitch(text string) (pitch.Pitch, error) {
	const op = "parse pitch"
	if !s.IsValid() {
		return pitch.Pitch{}, unsetError(op, "settings")
	}
	p, err := s.notation.ParsePitch(text, s.tuning.System(), s.context())
	if err != nil {
		if errors.Is(err, system.ErrInadmissible) {
			return pitch.Pitch{}, joinInadmissible(op, err)
		}
		return pitch.Pitch{}, wrapError(op, err)
	}
	return p, nil
}

// MustParsePitch is like ParsePitch but panics with its error on failure.
func (s *Settings) MustParsePitch(text string) pitch.Pitch {
	return must(s.ParsePitch(text))
}

// context returns the octave context of a set s for its own calls of package
// notation. The catalog is borrowed: notation does not keep it.
func (s *Settings) context() notation.OctaveContext {
	return notation.OctaveContext{Numbering: s.numbering, Catalog: s.catalog}
}

// expresses reports whether the valid name is expressible by the set s.
func (s *Settings) expresses(name spelling.Spelling) bool {
	if !admits(s.tuning.System(), name) {
		return false
	}
	_, err := s.notation.Format(name)
	return err == nil
}

// expressesPitch reports whether the valid pitch p is expressible by the set
// s, by the exact conditions of FormatPitch, without building its text.
func (s *Settings) expressesPitch(p pitch.Pitch) bool {
	name := p.Spelling()
	if !s.expresses(name) {
		return false
	}
	style := s.notation.OctaveStyle()
	if style == notation.OctaveStyleNone {
		return false
	}
	sys := s.tuning.System()
	written, err := sys.WrittenOctave(p)
	if err != nil {
		return false
	}
	w, err := written.ASPN()
	if err != nil {
		return false
	}
	switch style {
	case notation.OctaveStyleNumber:
		_, err := s.numbering.FromASPN(w)
		return err == nil
	case notation.OctaveStyleName:
		return s.catalog.Name(w) != ""
	case notation.OctaveStyleAM:
		return 0 <= w && w <= 8
	case notation.OctaveStyleHelmholtz:
		return s.helmholtzFits(name, w)
	case notation.OctaveStyleHelmholtzEnglish:
		if 0 <= w && w <= 2 {
			return true
		}
		return s.helmholtzFits(name, w)
	case notation.OctaveStyleHelmholtzNumbered:
		if w <= 2 {
			_, ok := checkedSub(2, w)
			return ok
		}
		return true
	case notation.OctaveStyleLilyPond:
		if w < 3 {
			count, ok := checkedSub(3, w)
			return ok && s.repeatFits(name, count, asciiMark, 3)
		}
		return s.repeatFits(name, w-3, asciiMark, 3)
	case notation.OctaveStyleABC:
		switch {
		case w == 4 || w == 5:
			return true
		case w < 4:
			count, ok := checkedSub(4, w)
			return ok && s.repeatFits(name, count, asciiMark, 4)
		default:
			return s.repeatFits(name, w-5, asciiMark, 5)
		}
	default:
		return false
	}
}

// helmholtzFits reports whether the Helmholtz form of name in the written
// octave w fits in an int length: below ASPN 3 the upper-case name with
// 2 − w low marks, from ASPN 3 the lower-case name with w − 3 high marks.
func (s *Settings) helmholtzFits(name spelling.Spelling, w int) bool {
	if w <= 2 {
		count, ok := checkedSub(2, w)
		return ok && s.repeatFits(name, count, helmholtzLowMark, 2)
	}
	return s.repeatFits(name, w-3, helmholtzHighMark, 3)
}

// repeatFits reports whether count marks of markLen bytes after the base form
// of name fit in an int length, as package notation checks before it builds
// the text. The base form is the text of name in the written octave base,
// where the style writes no marks; it is obtained from package notation
// itself, and only when count is large enough for its length to matter.
func (s *Settings) repeatFits(name spelling.Spelling, count, markLen, base int) bool {
	if count <= lengthSafeCount/markLen {
		return true
	}
	baseLen, ok := s.baseLength(name, base)
	return ok && count <= (math.MaxInt-baseLen)/markLen
}

// baseLength returns the byte length of the text of name in the written
// octave base by FormatPitch of the notation of s.
func (s *Settings) baseLength(name spelling.Spelling, base int) (int, bool) {
	sys := s.tuning.System()
	carry, err := sys.Carry(name)
	if err != nil {
		return 0, false
	}
	p, err := pitch.FromASPN(name, base+carry)
	if err != nil {
		return 0, false
	}
	text, err := s.notation.FormatPitch(p, sys, s.context())
	if err != nil {
		return 0, false
	}
	return len(text), true
}

// admits reports whether the valid name is admissible in the valid system.
func admits(sys *system.System, name spelling.Spelling) bool {
	ok, err := sys.Admits(name)
	return err == nil && ok
}

// checkedSub returns a − b and whether it fits in an int.
func checkedSub(a, b int) (int, bool) {
	d := a - b
	if (b > 0 && d > a) || (b < 0 && d < a) {
		return 0, false
	}
	return d, true
}

// joinInadmissible returns an error that matches ErrInadmissible and keeps
// err, which already matches system.ErrInadmissible.
func joinInadmissible(op string, err error) error {
	return fmt.Errorf("%w: %s: %w", ErrInadmissible, op, err)
}
