package pitch

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/spelling"
)

// ErrInvalid reports that a [Pitch], [spelling.Spelling], or [octave.Octave]
// supplied to an operation is invalid (its zero value).
var ErrInvalid = errors.New("pitch: invalid pitch")

// Pitch is an immutable written pitch: a note [spelling.Spelling] paired with
// a positional [octave.Octave].
//
// The octave component identifies the positional octave section in a musical
// system (such as C through B in ASPN 4 in the traditional system, where both
// notated "B♯3" and "C4" have positional octave ASPN 4), not the written
// register digit printed in score notation.
//
// Create a Pitch with [New], [MustNew], [FromASPN], or [MustFromASPN], or
// obtain one from [Range], [MustRange], and the transformation methods on
// Pitch.
//
// The zero value, Pitch{}, is the only invalid Pitch value: both its spelling
// and its octave are unset, and [Pitch.IsValid] reports false. No partially
// valid Pitch can be constructed, and no operation silently treats Pitch{} as
// C4, C0, A4, or a note without an octave.
//
// Pitch is a comparable value type suitable for ==, !=, switch statements, and
// map keys. For all valid pitches, p == q agrees with p.Equal(q). For the zero
// value, Pitch{} == Pitch{} is true in Go whereas (Pitch{}).Equal(Pitch{}) is
// false.
type Pitch struct {
	s spelling.Spelling
	o octave.Octave
}

// New returns the written pitch with spelling s and positional octave o.
//
// If s or o is invalid, New returns Pitch{} and an error matching
// [ErrInvalid].
func New(s spelling.Spelling, o octave.Octave) (Pitch, error) {
	if !s.IsValid() || !o.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: new: %w", ErrInvalid)
	}

	return Pitch{s: s, o: o}, nil
}

// MustNew is like [New] but panics with the returned error if s or o is
// invalid.
func MustNew(s spelling.Spelling, o octave.Octave) Pitch {
	return must(New(s, o))
}

// FromASPN returns the written pitch with spelling s and the positional octave
// whose ASPN number is aspn, equivalent to New(s, octave.FromASPN(aspn)).
//
// Every int is a valid ASPN octave number, so FromASPN fails only when s is
// invalid, returning Pitch{} and an error matching [ErrInvalid].
func FromASPN(s spelling.Spelling, aspn int) (Pitch, error) {
	if !s.IsValid() {
		return Pitch{}, fmt.Errorf("pitch: from ASPN: %w", ErrInvalid)
	}

	return Pitch{s: s, o: octave.FromASPN(aspn)}, nil
}

// MustFromASPN is like [FromASPN] but panics with the returned error if s is
// invalid.
func MustFromASPN(s spelling.Spelling, aspn int) Pitch {
	return must(FromASPN(s, aspn))
}

// IsValid reports whether p is a valid written pitch. It returns false only
// for the zero value, Pitch{}.
func (p Pitch) IsValid() bool {
	return p.s.IsValid() && p.o.IsValid()
}

// Spelling returns the note spelling of p, or spelling.Spelling{} if p is
// invalid.
func (p Pitch) Spelling() spelling.Spelling {
	return p.s
}

// Octave returns the positional octave of p, or octave.Octave{} if p is
// invalid.
func (p Pitch) Octave() octave.Octave {
	return p.o
}

// Parts returns the note spelling and the positional octave of p. For the zero
// value Pitch{}, Parts returns (spelling.Spelling{}, octave.Octave{}).
func (p Pitch) Parts() (spelling.Spelling, octave.Octave) {
	return p.s, p.o
}

// ASPN returns the ASPN number of p's positional octave.
//
// If p is invalid, ASPN returns 0 and an error matching [ErrInvalid].
func (p Pitch) ASPN() (int, error) {
	if !p.IsValid() {
		return 0, fmt.Errorf("pitch: ASPN: %w", ErrInvalid)
	}

	return p.o.MustASPN(), nil
}

// MustASPN is like [Pitch.ASPN] but panics with the returned error if p is
// invalid.
func (p Pitch) MustASPN() int {
	return must(p.ASPN())
}

// IsNatural reports whether p is valid and its spelling has alteration value 0
// (natural). For Pitch{}, IsNatural returns false.
func (p Pitch) IsNatural() bool {
	return p.s.IsNatural()
}

// IsInteger reports whether p is valid and its spelling has an integer
// alteration value. For Pitch{}, IsInteger returns false.
func (p Pitch) IsInteger() bool {
	return p.s.IsInteger()
}

// Direction returns the alteration direction of p's spelling:
// [spelling.DirectionLowered], [spelling.DirectionNatural], or
// [spelling.DirectionRaised] for a valid pitch, and [spelling.DirectionUnset]
// for Pitch{}.
func (p Pitch) Direction() spelling.Direction {
	return p.s.Direction()
}

// SameSpelling reports whether p and other are both valid and have the same
// note spelling (both letter and alteration value).
func (p Pitch) SameSpelling(other Pitch) bool {
	return p.IsValid() && other.IsValid() && p.s.Equal(other.s)
}

// SameOctave reports whether p and other are both valid and have the same
// positional octave.
func (p Pitch) SameOctave(other Pitch) bool {
	return p.IsValid() && other.IsValid() && p.o.Equal(other.o)
}

// SameLetter reports whether p and other are both valid and their spellings
// have the same note letter.
func (p Pitch) SameLetter(other Pitch) bool {
	return p.IsValid() && other.IsValid() && p.s.SameLetter(other.s)
}

// SameAlteration reports whether p and other are both valid and their
// spellings have the same alteration value.
func (p Pitch) SameAlteration(other Pitch) bool {
	return p.IsValid() && other.IsValid() && p.s.SameAlteration(other.s)
}

// Equal reports whether p and other are both valid and represent the same
// written pitch (the same [spelling.Spelling] and the same [octave.Octave]).
//
// Enharmonically equivalent spellings are not equal: (C♯, ASPN 4) is not equal
// to (D♭, ASPN 4), and (B♯, ASPN 4) is not equal to (C, ASPN 4). If either
// operand is Pitch{}, Equal returns false (whereas Pitch{} == Pitch{} is
// true).
func (p Pitch) Equal(other Pitch) bool {
	return p.IsValid() && other.IsValid() && p.s.Equal(other.s) && p.o.Equal(other.o)
}

// Compare compares p and other in the package's deterministic conventional
// order and returns -1 if p < other, 0 if p == other, and +1 if p > other.
//
// Valid pitches are ordered first by [octave.Octave.Compare] and, within the
// same octave, by [spelling.Spelling.Compare] (letter C through B, then
// alteration value). The invalid value Pitch{} sorts before every valid pitch
// and compares equal (0) to itself.
//
// This order is a conventional order of written records suitable for
// slices.SortFunc(xs, pitch.Pitch.Compare); it does not reflect sounding pitch
// height or position in a musical system (for example, within ASPN 4, C sorts
// before B♯ and C♯ sorts before D♭).
func (p Pitch) Compare(other Pitch) int {
	switch {
	case !p.IsValid() && !other.IsValid():
		return 0
	case !p.IsValid():
		return -1
	case !other.IsValid():
		return +1
	}

	if c := p.o.Compare(other.o); c != 0 {
		return c
	}

	return p.s.Compare(other.s)
}

// String returns the fixed debug representation of p: "<Spelling> ASPN <aspn>"
// (such as "C ASPN 4", "C+1 ASPN 4", or "C-1/2 ASPN -1") for a valid pitch,
// and "Pitch(invalid)" for Pitch{}.
//
// The returned string is intended for logs, debugging, and tests; it is not
// score notation and has no parser in this package. String implements
// [fmt.Stringer].
func (p Pitch) String() string {
	if !p.IsValid() {
		return "Pitch(invalid)"
	}

	return p.s.String() + " " + p.o.String()
}

// GoString returns a Go expression that constructs p through the public API,
// such as `pitch.MustFromASPN(spelling.MustParse("C+1"), 4)` for a valid pitch
// and `"pitch.Pitch{}"` for Pitch{}. GoString implements [fmt.GoStringer] for
// the %#v format verb.
func (p Pitch) GoString() string {
	if !p.IsValid() {
		return "pitch.Pitch{}"
	}

	return "pitch.MustFromASPN(spelling.MustParse(" + strconv.Quote(p.s.String()) + "), " + strconv.Itoa(p.o.MustASPN()) + ")"
}

// must returns v if err is nil and panics with err otherwise.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
