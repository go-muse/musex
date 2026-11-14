package note

import (
	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
)

// Note is a note: a written record (a name, or a written pitch: a name and
// an Octave), a correction in cents and an optional reference to a source of
// settings.
//
// The zero Note is the unset note; it is the only unset note. Create other
// notes with the functions of the package, such as FromPitch and Parse.
//
// A Note is a value. Assignment copies it whole: the copy and the original
// are independent, and only the reference to the source is shared. The
// changing methods have a pointer receiver and change the note itself; the
// other methods have a value receiver and never change the note.
//
// == compiles for notes but is not the equality of records: it also compares
// the sources and panics when two sources have one incomparable dynamic type.
// Use Equal.
type Note struct {
	// name is the name of the note; the unset name means the unset note.
	name spelling.Spelling
	// oct is the Octave of the written pitch; the invalid value means a note
	// without an octave.
	oct octave.Octave
	// c is the correction in cents.
	c cents.Cents
	// src is the source of settings; nil means an unbound note.
	src settings.Source
}

// IsValid reports whether n is set: false only for the zero Note.
func (n Note) IsValid() bool {
	return n.name.IsValid()
}

// HasOctave reports whether n has an octave, that is, a written pitch.
// The unset note has no octave.
func (n Note) HasOctave() bool {
	return n.name.IsValid() && n.oct.IsValid()
}

// Spelling returns the name of n; for a note with an octave it is the name of
// its written pitch. The unset note returns the unset name.
func (n Note) Spelling() spelling.Spelling {
	return n.name
}

// Letter returns the letter of the name of n. The unset note returns the
// zero letter, which is not a valid letter.
func (n Note) Letter() letter.Letter {
	return n.name.Letter()
}

// Alteration returns the alteration of the name of n. The unset note returns
// alteration.Natural; the result has meaning only together with IsValid.
func (n Note) Alteration() alteration.Alteration {
	return n.name.Alteration()
}

// Cents returns the correction of n. The unset note returns zero cents.
func (n Note) Cents() cents.Cents {
	return n.c
}

// Pitch returns the written pitch of n. A note without an octave reports
// ErrNoOctave. The unset note returns the zero pitch without an error.
func (n Note) Pitch() (pitch.Pitch, error) {
	if !n.name.IsValid() {
		return pitch.Pitch{}, nil
	}

	if !n.oct.IsValid() {
		return pitch.Pitch{}, noOctaveError("Pitch", n.name)
	}

	return n.pitch(), nil
}

// MustPitch is like Pitch but panics with the error of Pitch.
func (n Note) MustPitch() pitch.Pitch {
	return must(n.Pitch())
}

// Octave returns the Octave of the written pitch of n. A note without an
// octave reports ErrNoOctave. The unset note returns the zero Octave without
// an error.
func (n Note) Octave() (octave.Octave, error) {
	if !n.name.IsValid() {
		return octave.Octave{}, nil
	}

	if !n.oct.IsValid() {
		return octave.Octave{}, noOctaveError("Octave", n.name)
	}

	return n.oct, nil
}

// MustOctave is like Octave but panics with the error of Octave.
func (n Note) MustOctave() octave.Octave {
	return must(n.Octave())
}

// ASPN returns the ASPN number of the Octave of n. A note without an octave
// reports ErrNoOctave. The unset note returns 0 without an error.
func (n Note) ASPN() (int, error) {
	if !n.name.IsValid() {
		return 0, nil
	}

	if !n.oct.IsValid() {
		return 0, noOctaveError("ASPN", n.name)
	}

	return n.oct.MustASPN(), nil
}

// MustASPN is like ASPN but panics with the error of ASPN.
func (n Note) MustASPN() int {
	return must(n.ASPN())
}

// IsNatural reports whether the name of n is natural. The unset note
// reports false.
func (n Note) IsNatural() bool {
	return n.name.IsValid() && n.name.IsNatural()
}

// IsInteger reports whether the alteration of the name of n is an integer.
// The unset note reports false.
func (n Note) IsInteger() bool {
	return n.name.IsValid() && n.name.IsInteger()
}

// Direction returns the direction of the sign of the name of n. The unset
// note returns spelling.DirectionUnset.
func (n Note) Direction() spelling.Direction {
	if !n.name.IsValid() {
		return spelling.DirectionUnset
	}

	return n.name.Direction()
}

// Source returns the source of settings of n; nil for an unbound note.
// Source never reads the source.
func (n Note) Source() settings.Source {
	return n.src
}

// IsBound reports whether n is bound to a source of settings. The unset note
// is not bound.
func (n Note) IsBound() bool {
	return n.src != nil
}

// String returns a debugging form of n: the record, the correction and the
// binding, such as "C+1 ASPN 4, +30 cents, bound", or "Note(unset)" for the
// unset note. It is not a notation, it is never parsed back and it does not
// read the source. Use Format for the text of a note in a notation.
func (n Note) String() string {
	if !n.name.IsValid() {
		return "Note(unset)"
	}

	record := n.name.String()
	if n.oct.IsValid() {
		record = n.pitch().String()
	}

	binding := "unbound"
	if n.src != nil {
		binding = "bound"
	}

	return record + ", " + n.c.Text(cents.WithUnit|cents.WithPlus) + ", " + binding
}

// pitch returns the written pitch of a note with an octave.
func (n Note) pitch() pitch.Pitch {
	return pitch.MustNew(n.name, n.oct)
}

// withPitch returns a copy of n with the record replaced by p.
func (n Note) withPitch(p pitch.Pitch) Note {
	n.name, n.oct = p.Parts()

	return n
}

// must returns v or panics with err.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

// must3 returns a and b or panics with err.
func must3[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}

	return a, b
}
