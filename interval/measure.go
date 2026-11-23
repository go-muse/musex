package interval

import (
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/system"
)

// MeasureNotes measures the interval from the note from to the note to in
// one operation: the written interval and the signed chromatic displacement
// in the system of their settings. Both notes must have octaves; the cent
// corrections do not take part: C4 +30 cents to E4 -20 cents is M3 and +4.
// The generic size is the size of the written interval (Interval.Generic).
//
// Each distinct source is read once, and the written interval and the
// displacement use the same settings snapshot; with equal sources the system
// of from is retained. The systems must be equal by system.System.Equal.
//
// Errors, in this order: ErrUnset when from or to is unset; ErrNoOctave when
// either note has no octave; ErrNoSystem when a note, from first, is not
// bound to a source; the errors of note.Note.Settings, such as
// note.ErrUnsetSettings; ErrDifferentSystems when the systems differ;
// note.ErrInadmissible, together with system.ErrInadmissible, when the name
// of from, then of to, is not admissible in the system; ErrOutOfRange,
// together with system.ErrOutOfRange, when the displacement does not fit.
func MeasureNotes(from, to note.Note) (Interval, Chromatic, error) {
	const op = "MeasureNotes"

	if err := checkNotes(op, from, to); err != nil {
		return Interval{}, Chromatic{}, err
	}

	if !from.HasOctave() || !to.HasOctave() {
		return Interval{}, Chromatic{}, opError(op, ErrNoOctave)
	}

	if !from.IsBound() || !to.IsBound() {
		return Interval{}, Chromatic{}, opError(op, ErrNoSystem)
	}

	sys, err := noteSystem(op, from)
	if err != nil {
		return Interval{}, Chromatic{}, err
	}

	other := sys
	if !from.SameSource(to) {
		if other, err = noteSystem(op, to); err != nil {
			return Interval{}, Chromatic{}, err
		}
	}

	if !sys.Equal(other) {
		return Interval{}, Chromatic{}, opError(op, ErrDifferentSystems)
	}

	return measureNotes(op, sys, from, to, noteInadmissibleError)
}

// MustMeasureNotes is like MeasureNotes but panics if MeasureNotes returns
// an error. The panic value is the error returned by MeasureNotes.
func MustMeasureNotes(from, to note.Note) (Interval, Chromatic) {
	return must3(MeasureNotes(from, to))
}

// MeasureNotesIn is like MeasureNotes but works in the explicit system sys:
// the sources of the notes are not read. Both notes must have octaves.
//
// Errors, in this order: ErrUnset when sys is nil or unset or when from or
// to is unset; ErrNoOctave when either note has no octave;
// system.ErrInadmissible when the name of from, then of to, is not
// admissible in sys; ErrOutOfRange, together with system.ErrOutOfRange,
// when the displacement does not fit.
func MeasureNotesIn(sys *system.System, from, to note.Note) (Interval, Chromatic, error) {
	const op = "MeasureNotesIn"

	if err := checkSystem(op, sys); err != nil {
		return Interval{}, Chromatic{}, err
	}

	if err := checkNotes(op, from, to); err != nil {
		return Interval{}, Chromatic{}, err
	}

	if !from.HasOctave() || !to.HasOctave() {
		return Interval{}, Chromatic{}, opError(op, ErrNoOctave)
	}

	return measureNotes(op, sys, from, to, inadmissibleError)
}

// MustMeasureNotesIn is like MeasureNotesIn but panics if MeasureNotesIn
// returns an error. The panic value is the error returned by MeasureNotesIn.
func MustMeasureNotesIn(sys *system.System, from, to note.Note) (Interval, Chromatic) {
	return must3(MeasureNotesIn(sys, from, to))
}

// measureNotes returns the written interval and the chromatic displacement
// between the written pitches of the valid notes from and to with octaves in
// the one system sys, checking their names first.
func measureNotes(op string, sys *system.System, from, to note.Note, inadm inadmissibleFunc) (Interval, Chromatic, error) {
	i, err := notesByPitches(op, sys, from, to, inadm)
	if err != nil {
		return Interval{}, Chromatic{}, err
	}

	c, err := chromaticNotes(op, sys, from, to)
	if err != nil {
		return Interval{}, Chromatic{}, err
	}

	return i, c, nil
}

// must3 returns a and b or panics with err.
func must3[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}

	return a, b
}
