package degree

import (
	"math"

	"github.com/go-muse/musex/note"
)

const (
	// MaxNumber is the greatest degree number supported by the package.
	MaxNumber = math.MaxInt
	// MaxRomanNumber is the greatest number with a canonical Roman form.
	MaxRomanNumber = 3999
)

// Degree is an ordinal degree with an optional note.
//
// The zero Degree is unset. A set degree has a number of at least one and
// either has no note or contains any set note. Assignment makes an independent
// copy of the whole degree; only the source of settings of its note is shared.
//
// Degree is intentionally not comparable: == and use as a map key do not
// compile. Use Equal, SameNumber, SameNote, and Compare instead.
type Degree struct {
	number int
	n      note.Note
	_      [0]func()
}

// Transform returns the note that replaces the given independent copy. Its
// error stops the operation and is preserved inside ErrTransformFailed.
type Transform func(n note.Note) (note.Note, error)

// New returns a set degree without a note. number must be at least one;
// otherwise New reports ErrInvalidNumber and returns the unset degree.
func New(number int) (Degree, error) {
	if number < 1 {
		return Degree{}, invalidNumberError("New", "number is below one")
	}

	return Degree{number: number}, nil
}

// NewWithNote returns a set degree containing an independent copy of n. Any
// set note is allowed. It reports ErrUnset for an unset note before checking
// number, and ErrInvalidNumber when number is below one. On an error it
// returns the unset degree.
func NewWithNote(number int, n note.Note) (Degree, error) {
	if !n.IsValid() {
		return Degree{}, unsetError("NewWithNote", "unset note")
	}

	if number < 1 {
		return Degree{}, invalidNumberError("NewWithNote", "number is below one")
	}

	return Degree{number: number, n: n}, nil
}

// First returns a new independent degree with number 1 and no note.
func First() Degree { return Degree{number: 1} }

// Second returns a new independent degree with number 2 and no note.
func Second() Degree { return Degree{number: 2} }

// Third returns a new independent degree with number 3 and no note.
func Third() Degree { return Degree{number: 3} }

// Fourth returns a new independent degree with number 4 and no note.
func Fourth() Degree { return Degree{number: 4} }

// Fifth returns a new independent degree with number 5 and no note.
func Fifth() Degree { return Degree{number: 5} }

// Sixth returns a new independent degree with number 6 and no note.
func Sixth() Degree { return Degree{number: 6} }

// Seventh returns a new independent degree with number 7 and no note.
// These seven predefined degrees are conveniences, not a seven-degree model.
func Seventh() Degree { return Degree{number: 7} }

// IsValid reports whether d is set. It is false only for the zero Degree.
func (d Degree) IsValid() bool {
	return d.number > 0
}

// Number returns the ordinal number of d, or 0 for the unset degree.
func (d Degree) Number() int {
	return d.number
}

// HasNote reports whether d is set and has a note.
func (d Degree) HasNote() bool {
	return d.IsValid() && d.n.IsValid()
}

// Note returns a new independent copy of the note of d. A degree without a
// note and the unset degree return the unset note.
func (d Degree) Note() note.Note {
	return d.n
}

// check reports ErrUnset for a nil or unset receiver of a changing method.
func (d *Degree) check(op string) error {
	if d == nil {
		return unsetError(op, "nil *Degree")
	}

	if !d.IsValid() {
		return unsetError(op, "unset degree")
	}

	return nil
}
