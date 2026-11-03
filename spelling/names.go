package spelling

import (
	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
)

// The ready-made names. Each function returns a new value of a name with one
// of the seven letters and one of the five traditional accidentals: the
// natural (no suffix), Sharp (+1), Flat (-1), DoubleSharp (+2), and
// DoubleFlat (-2). For example, CSharp() equals
// MustNew(letter.C, alteration.Sharp()). The names are functions, not
// variables, so they cannot be changed; Predefined lists all 35 of them.

// C returns the natural name C: the letter C with the natural (0).
func C() Spelling {
	return Spelling{l: letter.C}
}

// CSharp returns the name C sharp: the letter C with a sharp (+1).
func CSharp() Spelling {
	return Spelling{l: letter.C, a: alteration.Sharp()}
}

// CFlat returns the name C flat: the letter C with a flat (-1).
func CFlat() Spelling {
	return Spelling{l: letter.C, a: alteration.Flat()}
}

// CDoubleSharp returns the name C double sharp: the letter C with a double sharp (+2).
func CDoubleSharp() Spelling {
	return Spelling{l: letter.C, a: alteration.DoubleSharp()}
}

// CDoubleFlat returns the name C double flat: the letter C with a double flat (-2).
func CDoubleFlat() Spelling {
	return Spelling{l: letter.C, a: alteration.DoubleFlat()}
}

// D returns the natural name D: the letter D with the natural (0).
func D() Spelling {
	return Spelling{l: letter.D}
}

// DSharp returns the name D sharp: the letter D with a sharp (+1).
func DSharp() Spelling {
	return Spelling{l: letter.D, a: alteration.Sharp()}
}

// DFlat returns the name D flat: the letter D with a flat (-1).
func DFlat() Spelling {
	return Spelling{l: letter.D, a: alteration.Flat()}
}

// DDoubleSharp returns the name D double sharp: the letter D with a double sharp (+2).
func DDoubleSharp() Spelling {
	return Spelling{l: letter.D, a: alteration.DoubleSharp()}
}

// DDoubleFlat returns the name D double flat: the letter D with a double flat (-2).
func DDoubleFlat() Spelling {
	return Spelling{l: letter.D, a: alteration.DoubleFlat()}
}

// E returns the natural name E: the letter E with the natural (0).
func E() Spelling {
	return Spelling{l: letter.E}
}

// ESharp returns the name E sharp: the letter E with a sharp (+1).
func ESharp() Spelling {
	return Spelling{l: letter.E, a: alteration.Sharp()}
}

// EFlat returns the name E flat: the letter E with a flat (-1).
func EFlat() Spelling {
	return Spelling{l: letter.E, a: alteration.Flat()}
}

// EDoubleSharp returns the name E double sharp: the letter E with a double sharp (+2).
func EDoubleSharp() Spelling {
	return Spelling{l: letter.E, a: alteration.DoubleSharp()}
}

// EDoubleFlat returns the name E double flat: the letter E with a double flat (-2).
func EDoubleFlat() Spelling {
	return Spelling{l: letter.E, a: alteration.DoubleFlat()}
}

// F returns the natural name F: the letter F with the natural (0).
func F() Spelling {
	return Spelling{l: letter.F}
}

// FSharp returns the name F sharp: the letter F with a sharp (+1).
func FSharp() Spelling {
	return Spelling{l: letter.F, a: alteration.Sharp()}
}

// FFlat returns the name F flat: the letter F with a flat (-1).
func FFlat() Spelling {
	return Spelling{l: letter.F, a: alteration.Flat()}
}

// FDoubleSharp returns the name F double sharp: the letter F with a double sharp (+2).
func FDoubleSharp() Spelling {
	return Spelling{l: letter.F, a: alteration.DoubleSharp()}
}

// FDoubleFlat returns the name F double flat: the letter F with a double flat (-2).
func FDoubleFlat() Spelling {
	return Spelling{l: letter.F, a: alteration.DoubleFlat()}
}

// G returns the natural name G: the letter G with the natural (0).
func G() Spelling {
	return Spelling{l: letter.G}
}

// GSharp returns the name G sharp: the letter G with a sharp (+1).
func GSharp() Spelling {
	return Spelling{l: letter.G, a: alteration.Sharp()}
}

// GFlat returns the name G flat: the letter G with a flat (-1).
func GFlat() Spelling {
	return Spelling{l: letter.G, a: alteration.Flat()}
}

// GDoubleSharp returns the name G double sharp: the letter G with a double sharp (+2).
func GDoubleSharp() Spelling {
	return Spelling{l: letter.G, a: alteration.DoubleSharp()}
}

// GDoubleFlat returns the name G double flat: the letter G with a double flat (-2).
func GDoubleFlat() Spelling {
	return Spelling{l: letter.G, a: alteration.DoubleFlat()}
}

// A returns the natural name A: the letter A with the natural (0).
func A() Spelling {
	return Spelling{l: letter.A}
}

// ASharp returns the name A sharp: the letter A with a sharp (+1).
func ASharp() Spelling {
	return Spelling{l: letter.A, a: alteration.Sharp()}
}

// AFlat returns the name A flat: the letter A with a flat (-1).
func AFlat() Spelling {
	return Spelling{l: letter.A, a: alteration.Flat()}
}

// ADoubleSharp returns the name A double sharp: the letter A with a double sharp (+2).
func ADoubleSharp() Spelling {
	return Spelling{l: letter.A, a: alteration.DoubleSharp()}
}

// ADoubleFlat returns the name A double flat: the letter A with a double flat (-2).
func ADoubleFlat() Spelling {
	return Spelling{l: letter.A, a: alteration.DoubleFlat()}
}

// B returns the natural name B: the letter B with the natural (0).
func B() Spelling {
	return Spelling{l: letter.B}
}

// BSharp returns the name B sharp: the letter B with a sharp (+1).
func BSharp() Spelling {
	return Spelling{l: letter.B, a: alteration.Sharp()}
}

// BFlat returns the name B flat: the letter B with a flat (-1).
func BFlat() Spelling {
	return Spelling{l: letter.B, a: alteration.Flat()}
}

// BDoubleSharp returns the name B double sharp: the letter B with a double sharp (+2).
func BDoubleSharp() Spelling {
	return Spelling{l: letter.B, a: alteration.DoubleSharp()}
}

// BDoubleFlat returns the name B double flat: the letter B with a double flat (-2).
func BDoubleFlat() Spelling {
	return Spelling{l: letter.B, a: alteration.DoubleFlat()}
}

// Predefined returns the 35 ready-made names in the order of Spelling.Compare:
// for each letter in the canonical order C D E F G A B, the double flat, flat,
// natural, sharp, and double sharp. The first names are C double flat, C flat,
// C, C sharp, and C double sharp; the last one is B double sharp.
//
// Each call returns a new slice with length and capacity 35. The caller owns
// it and may modify or retain it; this does not affect the ready-made names or
// later calls.
func Predefined() []Spelling {
	accidentals := [...]alteration.Alteration{
		alteration.DoubleFlat(),
		alteration.Flat(),
		alteration.Natural(),
		alteration.Sharp(),
		alteration.DoubleSharp(),
	}

	names := make([]Spelling, 0, letter.Count*len(accidentals))
	for _, l := range letter.All() {
		for _, a := range accidentals {
			names = append(names, Spelling{l: l, a: a})
		}
	}

	return names
}
