package interval

import (
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/system"
)

// checkNotes returns ErrUnset for the first unset note.
func checkNotes(op string, ns ...note.Note) error {
	for _, n := range ns {
		if !n.IsValid() {
			return detailError(op, ErrUnset, "note")
		}
	}
	return nil
}

// checkOctaves returns ErrNoOctave when one of the valid notes x and y has an
// octave and the other has none.
func checkOctaves(op string, x, y note.Note) error {
	if x.HasOctave() != y.HasOctave() {
		return opError(op, ErrNoOctave)
	}
	return nil
}

// noteSystem returns the system of the settings of the valid note n with an
// octave. It reports ErrNoSystem for a note that is not bound and passes on
// the errors of note.Note.Settings.
func noteSystem(op string, n note.Note) (*system.System, error) {
	if !n.IsBound() {
		return nil, opError(op, ErrNoSystem)
	}
	set, err := n.Settings()
	if err != nil {
		return nil, causeError(op, err)
	}
	return set.System(), nil
}

// BetweenNotes returns the interval from the note x to the note y. The cent
// corrections of the notes do not take part: C4 +30 cents to E4 -20 cents is
// M3.
//
// Notes without octaves are names: the interval is Between of their names in
// the direction d, and the sources of the notes are not read. Notes with
// octaves are written pitches: the interval is BetweenPitches in the system of
// their settings, and d is not used. Each source is read once, and once in
// all when the notes have the same source; the systems must be equal by
// system.System.Equal.
//
// Errors, in this order: ErrUnset when x or y is unset; ErrNoOctave when one
// note has an octave and the other has none; ErrNoDirection when the notes
// have no octaves and d is neither Up nor Down; ErrNoSystem when a note with
// an octave, x first, is not bound to a source; the errors of
// note.Note.Settings, such as note.ErrUnsetSettings; ErrDifferentSystems when
// the systems differ; note.ErrInadmissible, together with
// system.ErrInadmissible, when the name of x, then of y, is not admissible in
// the system; ErrOutOfRange as for BetweenPitches.
func BetweenNotes(x, y note.Note, d Direction) (Interval, error) {
	const op = "BetweenNotes"
	if err := checkNotes(op, x, y); err != nil {
		return Interval{}, err
	}
	if err := checkOctaves(op, x, y); err != nil {
		return Interval{}, err
	}
	if !x.HasOctave() {
		return notesByNames(op, x, y, d)
	}
	if !x.IsBound() || !y.IsBound() {
		return Interval{}, opError(op, ErrNoSystem)
	}
	sx, err := noteSystem(op, x)
	if err != nil {
		return Interval{}, err
	}
	sy := sx
	if !x.SameSource(y) {
		if sy, err = noteSystem(op, y); err != nil {
			return Interval{}, err
		}
	}
	if !sx.Equal(sy) {
		return Interval{}, opError(op, ErrDifferentSystems)
	}
	return notesByPitches(op, sx, x, y, noteInadmissibleError)
}

// MustBetweenNotes is like BetweenNotes but panics if BetweenNotes returns an
// error. The panic value is the error returned by BetweenNotes.
func MustBetweenNotes(x, y note.Note, d Direction) Interval {
	return must(BetweenNotes(x, y, d))
}

// notesByNames returns the interval between the names of the valid notes x
// and y without octaves in the direction d.
func notesByNames(op string, x, y note.Note, d Direction) (Interval, error) {
	if !d.isSet() {
		return Interval{}, opError(op, ErrNoDirection)
	}
	i, ok := nameInterval(x.Spelling(), y.Spelling(), d == Down)
	if !ok {
		return Interval{}, alterationRangeError(op)
	}
	return i, nil
}

// notesByPitches returns the interval between the written pitches of the valid
// notes x and y with octaves in sys, checking their names first.
func notesByPitches(op string, sys *system.System, x, y note.Note, inadm inadmissibleFunc) (Interval, error) {
	px, _ := x.Pitch()
	py, _ := y.Pitch()
	if err := checkAdmitted(op, sys, inadm, "name", px.Spelling(), py.Spelling()); err != nil {
		return Interval{}, err
	}
	return pitchInterval(op, sys, px, py)
}

// BetweenNotesIn is like BetweenNotes but works in the explicit system sys:
// the sources of the notes are not read. The system is the context of the
// operation for all notes, so the names of notes without octaves are checked
// in it too.
//
// Errors, in this order: ErrUnset when sys is nil or unset or when x or y is
// unset; ErrNoOctave; ErrNoDirection when the notes have no octaves and d is
// neither Up nor Down; system.ErrInadmissible when the name of x, then of y,
// is not admissible in sys; ErrOutOfRange as for BetweenPitches.
func BetweenNotesIn(sys *system.System, x, y note.Note, d Direction) (Interval, error) {
	const op = "BetweenNotesIn"
	if err := checkSystem(op, sys); err != nil {
		return Interval{}, err
	}
	if err := checkNotes(op, x, y); err != nil {
		return Interval{}, err
	}
	if err := checkOctaves(op, x, y); err != nil {
		return Interval{}, err
	}
	if x.HasOctave() {
		return notesByPitches(op, sys, x, y, inadmissibleError)
	}
	if !d.isSet() {
		return Interval{}, opError(op, ErrNoDirection)
	}
	if err := checkAdmitted(op, sys, inadmissibleError, "name", x.Spelling(), y.Spelling()); err != nil {
		return Interval{}, err
	}
	return notesByNames(op, x, y, d)
}

// MustBetweenNotesIn is like BetweenNotesIn but panics if BetweenNotesIn
// returns an error. The panic value is the error returned by BetweenNotesIn.
func MustBetweenNotesIn(sys *system.System, x, y note.Note, d Direction) Interval {
	return must(BetweenNotesIn(sys, x, y, d))
}

// TransposeNote returns a new note: x moved by the interval i, with the cent
// correction and the source of x. The note x is not changed. A note without
// an octave is moved as a name by Transpose and stays without an octave;
// a note with an octave is moved as a written pitch by TransposePitch in the
// system of its settings. A bound note gets the new name only when it is
// admissible in the system of its settings, as package note requires.
//
// The source of a bound note with an octave is read twice: for the system and
// when the record of the new note is replaced. If the settings of the source
// change in between, the two reads can see different settings.
//
// Errors, in this order: ErrUnset when x or i is unset. For a note without an
// octave: ErrOutOfRange, together with alteration.ErrOutOfRange, for the
// alteration value of the result; then the errors of package note when the
// record is replaced, such as note.ErrInadmissible. For a note with an
// octave: ErrNoSystem when it is not bound; the errors of note.Note.Settings;
// note.ErrInadmissible, together with system.ErrInadmissible, when the name of
// x is not admissible in the system; ErrOutOfRange for the alteration value of
// the result; note.ErrInadmissible for the name of the result; ErrOutOfRange
// for the written octave and the octave; then the errors of package note when
// the record is replaced.
func TransposeNote(x note.Note, i Interval) (note.Note, error) {
	const op = "TransposeNote"
	if err := checkNotes(op, x); err != nil {
		return note.Note{}, err
	}
	if !i.IsValid() {
		return note.Note{}, detailError(op, ErrUnset, "interval")
	}
	if !x.HasOctave() {
		return transposeNoteName(op, nil, x, i)
	}
	sys, err := noteSystem(op, x)
	if err != nil {
		return note.Note{}, err
	}
	return transposeNotePitch(op, sys, x, i, noteInadmissibleError)
}

// MustTransposeNote is like TransposeNote but panics if TransposeNote returns
// an error. The panic value is the error returned by TransposeNote.
func MustTransposeNote(x note.Note, i Interval) note.Note {
	return must(TransposeNote(x, i))
}

// transposeNoteName moves the valid note x without an octave by the valid
// interval i. A non-nil sys is an explicit system in which the names of x and
// of the result are checked.
func transposeNoteName(op string, sys *system.System, x note.Note, i Interval) (note.Note, error) {
	if sys != nil {
		if err := checkAdmitted(op, sys, inadmissibleError, "name", x.Spelling()); err != nil {
			return note.Note{}, err
		}
	}
	name, ok := transposeName(x.Spelling(), i)
	if !ok {
		return note.Note{}, alterationRangeError(op)
	}
	if sys != nil {
		if err := checkAdmitted(op, sys, inadmissibleError, "name of the result", name); err != nil {
			return note.Note{}, err
		}
	}
	m := x
	if err := m.ReplaceWithSpelling(name); err != nil {
		return note.Note{}, causeError(op, err)
	}
	return m, nil
}

// transposeNotePitch moves the valid note x with an octave by the valid
// interval i in sys.
func transposeNotePitch(op string, sys *system.System, x note.Note, i Interval, inadm inadmissibleFunc) (note.Note, error) {
	px, _ := x.Pitch()
	if err := checkAdmitted(op, sys, inadm, "name", px.Spelling()); err != nil {
		return note.Note{}, err
	}
	p, err := transposePitch(op, sys, px, i, inadm)
	if err != nil {
		return note.Note{}, err
	}
	m := x
	if err := m.ReplaceWithPitch(p); err != nil {
		return note.Note{}, causeError(op, err)
	}
	return m, nil
}

// TransposeNoteIn is like TransposeNote but works in the explicit system sys
// for notes with and without octaves: the names of x and of the result are
// checked in sys, and a bound note then also checks the name of the result in
// the system of its settings, as package note requires.
//
// Errors, in this order: ErrUnset when sys is nil or unset or when x or i is
// unset; system.ErrInadmissible when the name of x is not admissible in sys;
// ErrOutOfRange for the alteration value of the result; system.ErrInadmissible
// for the name of the result; ErrOutOfRange for the written octave and the
// octave; then the errors of package note when the record is replaced.
func TransposeNoteIn(sys *system.System, x note.Note, i Interval) (note.Note, error) {
	const op = "TransposeNoteIn"
	if err := checkSystem(op, sys); err != nil {
		return note.Note{}, err
	}
	if err := checkNotes(op, x); err != nil {
		return note.Note{}, err
	}
	if !i.IsValid() {
		return note.Note{}, detailError(op, ErrUnset, "interval")
	}
	if !x.HasOctave() {
		return transposeNoteName(op, sys, x, i)
	}
	return transposeNotePitch(op, sys, x, i, inadmissibleError)
}

// MustTransposeNoteIn is like TransposeNoteIn but panics if TransposeNoteIn
// returns an error. The panic value is the error returned by TransposeNoteIn.
func MustTransposeNoteIn(sys *system.System, x note.Note, i Interval) note.Note {
	return must(TransposeNoteIn(sys, x, i))
}
