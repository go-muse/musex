package chord

import (
	"github.com/go-muse/musex/note"
)

// Contains reports whether c contains an occurrence of n: a note equal to n
// by record (note.Note.Equal: the record and the correction, without the
// source). An unset n and the unset chord give false.
func (c Chord) Contains(n note.Note) bool {
	for _, x := range c.notes {
		if x.Equal(n) {
			return true
		}
	}

	return false
}

// Count returns the multiplicity of n in c: the number of its occurrences.
func (c Chord) Count(n note.Note) int {
	count := 0

	for _, x := range c.notes {
		if x.Equal(n) {
			count++
		}
	}

	return count
}

// Find returns the first occurrence of n in c: its number, a copy of the note
// of the chord with its own source, and true. Without occurrences it returns
// -1, the unset note and false.
func (c Chord) Find(n note.Note) (int, note.Note, bool) {
	for i, x := range c.notes {
		if x.Equal(n) {
			return i, x, true
		}
	}

	return -1, note.Note{}, false
}

// Indices returns the numbers of all the occurrences of n in c in increasing
// order; without occurrences, an empty slice, never nil.
func (c Chord) Indices(n note.Note) []int {
	out := []int{}

	for i, x := range c.notes {
		if x.Equal(n) {
			out = append(out, i)
		}
	}

	return out
}

// ContainsAll reports whether c is set and contains all the notes of the list,
// counting multiplicity: the list is a multiset, so (C4, C4) needs two
// occurrences of C4. The empty list gives true for a set chord; an unset note
// in the list gives false.
func (c Chord) ContainsAll(notes ...note.Note) bool {
	if !c.set {
		return false
	}

	for _, n := range notes {
		if !n.IsValid() {
			return false
		}
	}

	_, inC, inList := multiplicities(c.notes, notes, note.Note.Equal)
	for id, need := range inList {
		if need > inC[id] {
			return false
		}
	}

	return true
}

// ContainsAny reports whether c is set and contains at least one note of the
// list. The empty list gives false; unset notes of the list are skipped.
func (c Chord) ContainsAny(notes ...note.Note) bool {
	if !c.set {
		return false
	}

	_, inC, inList := multiplicities(c.notes, notes, note.Note.Equal)
	for id, given := range inList {
		if given > 0 && inC[id] > 0 {
			return true
		}
	}

	return false
}

// ContainsRecord reports whether c has a note with the record of n,
// regardless of the correction (note.Note.SameRecord): one written pitch, or
// one name for notes without an octave. A note with an octave and a note
// without one never match.
func (c Chord) ContainsRecord(n note.Note) bool {
	for _, x := range c.notes {
		if x.SameRecord(n) {
			return true
		}
	}

	return false
}

// ContainsSpelling reports whether c has a note with the name of n,
// regardless of the octave and the correction (note.Note.SameSpelling).
func (c Chord) ContainsSpelling(n note.Note) bool {
	for _, x := range c.notes {
		if x.SameSpelling(n) {
			return true
		}
	}

	return false
}

// ContainsFunc reports whether c has a note for which cond gives true. cond
// gets copies of the notes in the order of record, at most once for each
// note, and the search stops at the first such note.
//
// Errors, in this order: ErrUnset for a nil cond; ErrRuleFailed with the
// error of cond. The result is then false. The unset chord gives false without
// calling cond.
func (c Chord) ContainsFunc(cond Condition) (bool, error) {
	if cond == nil {
		return false, unsetError("ContainsFunc", "nil condition")
	}

	for i, n := range c.notes {
		ok, err := cond(n)
		if err != nil {
			return false, ruleError("ContainsFunc", i, err)
		}

		if ok {
			return true, nil
		}
	}

	return false, nil
}

// CountFunc returns the number of the notes of c for which cond gives true.
// cond gets copies of all the notes in the order of record, once for each
// note, until its first error.
//
// Errors, in this order: ErrUnset for a nil cond; ErrRuleFailed with the
// error of cond. The result is then 0. The unset chord gives 0 without calling
// cond.
func (c Chord) CountFunc(cond Condition) (int, error) {
	selected, err := c.selectFunc("CountFunc", cond)
	if err != nil {
		return 0, err
	}

	count := 0

	for _, ok := range selected {
		if ok {
			count++
		}
	}

	return count, nil
}

// IndicesFunc returns the numbers of the notes of c for which cond gives true
// in increasing order; without such notes, an empty slice. cond is called as
// by CountFunc.
//
// Errors, in this order: ErrUnset for a nil cond; ErrRuleFailed with the
// error of cond. The result is then nil. The unset chord gives an empty slice
// without calling cond.
func (c Chord) IndicesFunc(cond Condition) ([]int, error) {
	selected, err := c.selectFunc("IndicesFunc", cond)
	if err != nil {
		return nil, err
	}

	out := []int{}

	for i, ok := range selected {
		if ok {
			out = append(out, i)
		}
	}

	return out, nil
}

// NotesFunc returns a new slice with copies of the notes of c for which cond
// gives true, in the order of record; without such notes, an empty slice.
// cond is called as by CountFunc.
//
// Errors, in this order: ErrUnset for a nil cond; ErrRuleFailed with the
// error of cond. The result is then nil. The unset chord gives an empty slice
// without calling cond.
func (c Chord) NotesFunc(cond Condition) ([]note.Note, error) {
	selected, err := c.selectFunc("NotesFunc", cond)
	if err != nil {
		return nil, err
	}

	out := []note.Note{}

	for i, ok := range selected {
		if ok {
			out = append(out, c.notes[i])
		}
	}

	return out, nil
}

// selectFunc calls cond for a copy of every note of c in the order of record
// until its first error and returns its answers. op names the operation in
// the errors.
func (c Chord) selectFunc(op string, cond Condition) ([]bool, error) {
	if cond == nil {
		return nil, unsetError(op, "nil condition")
	}

	selected := make([]bool, len(c.notes))

	for i, n := range c.notes {
		ok, err := cond(n)
		if err != nil {
			return nil, ruleError(op, i, err)
		}

		selected[i] = ok
	}

	return selected, nil
}
