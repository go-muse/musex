package chord

import (
	"iter"
	"slices"
	"strings"

	"github.com/go-muse/musex/note"
)

// Chord is a combination of notes sounding together, without a duration:
// zero or more set notes in the order of record. A note can enter a chord
// several times.
//
// A Chord has three states: unset, empty and with notes. The zero Chord
// (var c chord.Chord, or chord.Chord{}) is the unset chord, its only form:
// its reading methods give zero results, while At and every changing method
// report ErrUnset and leave it unset. New and MustNew make set chords; New()
// without notes makes the empty chord, a working value that accepts notes.
// No operation turns a set chord into the unset one.
//
// A Chord is a value. Assignment copies it in O(1) without allocations, and
// the copy and the original are independent at once, both ways: a change of
// either never shows in the other. Only the sources of the notes are shared.
// There is no Clone: a copy is made by assignment.
//
// The changing methods have a pointer receiver and change the chord itself;
// the other methods have a value receiver and never change it. A chord can be
// changed only where it is addressable: a variable, an element of a slice,
// a field of an addressable struct, or through a *Chord.
//
// == does not compile for chords, and a Chord cannot be a key of a map; a
// struct with a Chord field is not comparable either. any(c) == any(d)
// compiles and panics at run time. reflect.DeepEqual compares the
// representation, not the equality of records. Use Equal, EqualInOrder or
// SameRecords.
type Chord struct {
	// notes are the notes in the order of record. The array behind notes is
	// never written after it is stored in a chord: copies share it, and every
	// change stores a new array of exact length (copy on write).
	notes []note.Note
	// set tells a set chord from the unset one; the unset chord has no notes.
	set bool
}

// Condition is a rule of the user that selects a note. It gets a copy of a
// note of the chord. Its error stops the operation, which then reports
// ErrRuleFailed together with this error.
type Condition func(n note.Note) (bool, error)

// Transform is a rule of the user that returns the note that replaces n, a
// copy of a note of the chord. Its error stops the operation, which then
// reports ErrRuleFailed together with this error; an unset returned note is
// reported as ErrUnset.
type Transform func(n note.Note) (note.Note, error)

// Comparison is a rule of the user that orders two notes: a negative number,
// 0 or a positive number when a comes before, together with, or after b. Its
// error stops the operation, which then reports ErrRuleFailed together with
// this error.
type Comparison func(a, b note.Note) (int, error)

// New returns a set chord with the notes in the given order. Without notes it
// returns the empty chord. Repeated notes and any notes are allowed: with and
// without an octave, with any corrections, bound to any sources or unbound.
// The sources are not read. The list is copied: a later change of it does not
// change the chord.
//
// Errors: ErrUnset for the first unset note of the list; the result is then
// the unset chord.
func New(notes ...note.Note) (Chord, error) {
	if err := checkNotes("New", notes); err != nil {
		return Chord{}, err
	}

	return Chord{notes: cloneNotes(notes), set: true}, nil
}

// IsValid reports whether c is set: false only for the zero Chord.
func (c Chord) IsValid() bool {
	return c.set
}

// IsEmpty reports whether c has no notes. The unset chord has no notes.
func (c Chord) IsEmpty() bool {
	return len(c.notes) == 0
}

// Len returns the number of the notes of c, repeats included; 0 for the unset
// chord.
func (c Chord) Len() int {
	return len(c.notes)
}

// Notes returns a new slice with copies of the notes of c in the order of
// record. It is never nil: the empty and the unset chord give an empty slice.
// Changing the slice or its notes does not change c.
func (c Chord) Notes() []note.Note {
	out := make([]note.Note, len(c.notes))
	copy(out, c.notes)

	return out
}

// All returns a sequence of the numbers and copies of the notes of c in the
// order of record. The sequence is a snapshot of c at the call of All: a later
// change of c, even inside the loop over the sequence, does not change it.
// The sequence can be iterated again and concurrently with the same notes.
// The unset chord gives an empty sequence.
func (c Chord) All() iter.Seq2[int, note.Note] {
	notes := c.notes

	return func(yield func(int, note.Note) bool) {
		for i, n := range notes {
			if !yield(i, n) {
				return
			}
		}
	}
}

// Values returns a sequence of copies of the notes of c in the order of
// record. Like All, it is a snapshot of c at the call of Values.
func (c Chord) Values() iter.Seq[note.Note] {
	notes := c.notes

	return func(yield func(note.Note) bool) {
		for _, n := range notes {
			if !yield(n) {
				return
			}
		}
	}
}

// At returns a copy of the note with number i, counted from 0 in the order of
// record.
//
// Errors, in this order: ErrUnset for the unset chord; ErrIndex for i < 0 or
// i >= Len(). The note is then the unset note.
func (c Chord) At(i int) (note.Note, error) {
	if !c.set {
		return note.Note{}, unsetError("At", "unset chord")
	}

	if i < 0 || i >= len(c.notes) {
		return note.Note{}, indexError("At", i, len(c.notes))
	}

	return c.notes[i], nil
}

// SortedNotes returns a new slice with copies of the notes of c in the agreed
// order of notes (note.Note.Compare). The sort is stable: notes equal by
// Compare keep their order of record. c does not change.
func (c Chord) SortedNotes() []note.Note {
	out := c.Notes()
	slices.SortStableFunc(out, note.Note.Compare)

	return out
}

// DistinctNotes returns a new slice with the first occurrence of every note
// of c in the order of record: the chord without its repeats.
func (c Chord) DistinctNotes() []note.Note {
	ids, count := classes(c.notes, nil, note.Note.Equal)
	seen := make([]bool, count)
	out := make([]note.Note, 0, count)

	for i, n := range c.notes {
		if !seen[ids[i]] {
			seen[ids[i]] = true
			out = append(out, n)
		}
	}

	return out
}

// HasRepeats reports whether a note enters c more than once.
func (c Chord) HasRepeats() bool {
	_, count := classes(c.notes, nil, note.Note.Equal)

	return count < len(c.notes)
}

// HasNotesWithOctave reports whether c has a note with an octave
// (note.Note.HasOctave).
func (c Chord) HasNotesWithOctave() bool {
	for _, n := range c.notes {
		if n.HasOctave() {
			return true
		}
	}

	return false
}

// HasNotesWithoutOctave reports whether c has a note without an octave.
func (c Chord) HasNotesWithoutOctave() bool {
	for _, n := range c.notes {
		if !n.HasOctave() {
			return true
		}
	}

	return false
}

// String returns a debugging form of c: the debugging forms of its notes
// (note.Note.String) in the order of record, separated by "; ", in square
// brackets, such as "[C ASPN 4, +30 cents, unbound; E ASPN 4, 0 cents,
// unbound]"; "[]" for the empty chord and "Chord(unset)" for the unset one.
// It is not a notation, it is never parsed back, it does not take part in
// equality and it does not read the sources.
func (c Chord) String() string {
	if !c.set {
		return "Chord(unset)"
	}

	var b strings.Builder

	b.WriteByte('[')

	for i, n := range c.notes {
		if i > 0 {
			b.WriteString("; ")
		}

		b.WriteString(n.String())
	}

	b.WriteByte(']')

	return b.String()
}

// check reports ErrUnset for a nil or unset receiver of the changing method
// op.
func (c *Chord) check(op string) error {
	if c == nil {
		return unsetError(op, "nil *Chord")
	}

	if !c.set {
		return unsetError(op, "unset chord")
	}

	return nil
}

// put stores notes as the new notes of c with one assignment. The array
// behind notes must be new or already shared by chords, and it is never
// written afterwards; the capacity is cut so that no later append can write
// into the array.
func (c *Chord) put(notes []note.Note) {
	c.notes = notes[:len(notes):len(notes)]
}

// checkNotes reports ErrUnset for the first unset note of notes given to op.
func checkNotes(op string, notes []note.Note) error {
	for i, n := range notes {
		if !n.IsValid() {
			return unsetNoteError(op, i)
		}
	}

	return nil
}

// cloneNotes returns a new array with the notes, or nil for no notes.
func cloneNotes(notes []note.Note) []note.Note {
	if len(notes) == 0 {
		return nil
	}

	out := make([]note.Note, len(notes))
	copy(out, notes)

	return out
}

// classes divides the notes of a followed by the notes of b into classes of
// equal notes: ids[k] is the class of a[k] for k < len(a) and of b[k-len(a)]
// otherwise, and count is the number of the classes, numbered from 0.
//
// The notes are sorted stably by note.Note.Compare, and neighbors for which
// same holds form one class. same must hold only for notes that stand
// together in that order: note.Note.Equal (Compare == 0 for set notes) or
// note.Note.SameRecord (the correction is the last key of Compare).
func classes(a, b []note.Note, same func(x, y note.Note) bool) (ids []int, count int) {
	total := len(a) + len(b)
	if total == 0 {
		return nil, 0
	}

	at := func(k int) note.Note {
		if k < len(a) {
			return a[k]
		}

		return b[k-len(a)]
	}

	order := make([]int, total)
	for k := range order {
		order[k] = k
	}

	slices.SortStableFunc(order, func(x, y int) int {
		return at(x).Compare(at(y))
	})

	ids = make([]int, total)
	ids[order[0]] = 0

	for k := 1; k < total; k++ {
		if !same(at(order[k-1]), at(order[k])) {
			count++
		}

		ids[order[k]] = count
	}

	return ids, count + 1
}

// multiplicities returns, for the classes of classes(a, b, same), the number
// of the notes of a and of b in every class.
func multiplicities(a, b []note.Note, same func(x, y note.Note) bool) (ids, inA, inB []int) {
	ids, count := classes(a, b, same)
	inA = make([]int, count)
	inB = make([]int, count)

	for k, id := range ids {
		if k < len(a) {
			inA[id]++
		} else {
			inB[id]++
		}
	}

	return ids, inA, inB
}
