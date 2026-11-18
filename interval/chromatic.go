package interval

import (
	"errors"
	"strconv"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// Chromatic is an immutable signed shift in one musical system. Its zero
// value is unset; a constructed zero shift is valid. Use SameShift for musical
// equality: Go equality compares the system pointers. Copies share only the
// immutable system and can be read concurrently.
type Chromatic struct {
	sys   *system.System
	steps system.Step
}

// NewChromatic records steps in sys. Any int is accepted; an unset system
// returns ErrUnset.
func NewChromatic(sys *system.System, steps int) (Chromatic, error) {
	if err := checkSystem("NewChromatic", sys); err != nil {
		return Chromatic{}, err
	}
	return Chromatic{sys, system.Step(steps)}, nil
}

// MustNewChromatic panics with the error returned by NewChromatic.
func MustNewChromatic(sys *system.System, steps int) Chromatic { return must(NewChromatic(sys, steps)) }

// IsValid reports whether a system was supplied.
func (c Chromatic) IsValid() bool { return c.sys.IsValid() }

// Steps returns the signed shift, or zero for an unset value.
func (c Chromatic) Steps() int { return int(c.steps) }

// System returns the immutable system, or nil for an unset value.
func (c Chromatic) System() *system.System { return c.sys }

// String returns a signed decimal shift, "0", or "Chromatic(unset)".
func (c Chromatic) String() string {
	if !c.IsValid() {
		return "Chromatic(unset)"
	}
	if c.steps > 0 {
		return "+" + strconv.Itoa(int(c.steps))
	}
	return strconv.Itoa(int(c.steps))
}

// SameShift compares signed shifts in musically equal systems. It returns
// ErrUnset for either unset operand and ErrDifferentSystems for unequal scales.
func (c Chromatic) SameShift(other Chromatic) (bool, error) {
	if !c.IsValid() || !other.IsValid() {
		return false, ErrUnset
	}
	if !c.sys.Equal(other.sys) {
		return false, ErrDifferentSystems
	}
	return c.steps == other.steps, nil
}

// MustSameShift panics with the error returned by SameShift.
func (c Chromatic) MustSameShift(other Chromatic) bool { return must(c.SameShift(other)) }

func chromaticNotesCheck(from, to note.Note) error {
	if err := checkNotes("ChromaticBetweenNotes", from, to); err != nil {
		return err
	}
	if !from.HasOctave() || !to.HasOctave() {
		return ErrNoOctave
	}
	return nil
}

// ChromaticBetweenNotes measures registered notes, ignoring cents. Each
// distinct settings source is read once; the first note supplies the retained
// system. Unbound notes return ErrNoSystem; unequal systems return
// ErrDifferentSystems. Both notes must have octaves.
func ChromaticBetweenNotes(from, to note.Note) (Chromatic, error) {
	if err := chromaticNotesCheck(from, to); err != nil {
		return Chromatic{}, err
	}
	if !from.IsBound() || !to.IsBound() {
		return Chromatic{}, ErrNoSystem
	}
	sys, err := noteSystem("ChromaticBetweenNotes", from)
	if err != nil {
		return Chromatic{}, err
	}
	other := sys
	if !from.SameSource(to) {
		other, err = noteSystem("ChromaticBetweenNotes", to)
		if err != nil {
			return Chromatic{}, err
		}
	}
	if !sys.Equal(other) {
		return Chromatic{}, ErrDifferentSystems
	}
	return chromaticNotes(sys, from, to)
}
func chromaticNotes(sys *system.System, from, to note.Note) (Chromatic, error) {
	a, _ := from.Pitch()
	b, _ := to.Pitch()
	s, err := sys.Distance(a, b)
	if err != nil {
		if errors.Is(err, system.ErrOutOfRange) {
			return Chromatic{}, rangeError("ChromaticBetweenNotes", err)
		}
		return Chromatic{}, err
	}
	return Chromatic{sys, s}, nil
}

// MustChromaticBetweenNotes panics with the measurement error.
func MustChromaticBetweenNotes(from, to note.Note) Chromatic {
	return must(ChromaticBetweenNotes(from, to))
}

// ChromaticBetweenNotesIn measures in an explicit system without reading
// either source. The system is checked before notes and their octaves.
func ChromaticBetweenNotesIn(sys *system.System, from, to note.Note) (Chromatic, error) {
	if err := checkSystem("ChromaticBetweenNotesIn", sys); err != nil {
		return Chromatic{}, err
	}
	if err := chromaticNotesCheck(from, to); err != nil {
		return Chromatic{}, err
	}
	return chromaticNotes(sys, from, to)
}

// MustChromaticBetweenNotesIn panics with the measurement error.
func MustChromaticBetweenNotesIn(sys *system.System, from, to note.Note) Chromatic {
	return must(ChromaticBetweenNotesIn(sys, from, to))
}

// Chromatic measures i using SignedSteps, retaining sys. start is optional
// only in a regular system; all errors of SignedSteps are preserved.
func (i Interval) Chromatic(sys *system.System, start spelling.Spelling) (Chromatic, error) {
	s, err := i.SignedSteps(sys, start)
	if err != nil {
		return Chromatic{}, err
	}
	return Chromatic{sys, s}, nil
}

// MustChromatic panics with the error returned by Chromatic.
func (i Interval) MustChromatic(sys *system.System, start spelling.Spelling) Chromatic {
	return must(i.Chromatic(sys, start))
}

// TransposeNote moves a registered note using System.Move and rule, even for
// a zero shift. It preserves cents and Source without changing from. A bound
// note must have an equal system; replacing its pitch rechecks its settings.
func (c Chromatic) TransposeNote(from note.Note, rule system.SpellingRule) (note.Note, error) {
	if !c.IsValid() || !from.IsValid() {
		return note.Note{}, ErrUnset
	}
	if !from.HasOctave() {
		return note.Note{}, ErrNoOctave
	}
	if from.IsBound() {
		sys, err := noteSystem("Chromatic.TransposeNote", from)
		if err != nil {
			return note.Note{}, err
		}
		if !c.sys.Equal(sys) {
			return note.Note{}, ErrDifferentSystems
		}
	}
	p, _ := from.Pitch()
	target, err := c.sys.Move(p, c.steps, rule)
	if err != nil {
		return note.Note{}, err
	}
	result := from
	if err = result.ReplaceWithPitch(target); err != nil {
		return note.Note{}, err
	}
	return result, nil
}

// MustTransposeNote panics with the error returned by TransposeNote.
func (c Chromatic) MustTransposeNote(from note.Note, rule system.SpellingRule) note.Note {
	return must(c.TransposeNote(from, rule))
}
