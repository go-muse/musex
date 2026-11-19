package chord

import (
	"slices"

	"github.com/go-muse/musex/note"
)

// The changing methods change c itself and return only an error; each Must
// variant returns c. Every changing method is atomic: it is done whole, or on
// an error c does not change at all, with the same notes, order, corrections
// and sources. A nil *Chord and the unset chord report ErrUnset and stay as
// they are. The notes, lists, chords and rules given as arguments never
// change, and the lists are not kept. The sources of the notes are not read.

// Add appends the notes of the list to c in the order of the list. Repeats
// are allowed. Without notes c does not change.
//
// Errors: ErrUnset for the chord, then for the first unset note of the list.
func (c *Chord) Add(notes ...note.Note) error {
	if err := c.check("Add"); err != nil {
		return err
	}

	if err := checkNotes("Add", notes); err != nil {
		return err
	}

	if len(notes) == 0 {
		return nil
	}

	out := make(note.Notes, len(c.notes)+len(notes))
	copy(out, c.notes)
	copy(out[len(c.notes):], notes)
	c.put(out)

	return nil
}

// AddMissing appends the notes of the list in its order, each one only when c
// does not yet contain an occurrence of it at that moment: of repeated notes
// in the list, only the first is added.
//
// Errors: ErrUnset for the chord, then for the first unset note of the list.
func (c *Chord) AddMissing(notes ...note.Note) error {
	if err := c.check("AddMissing"); err != nil {
		return err
	}

	if err := checkNotes("AddMissing", notes); err != nil {
		return err
	}

	ids, inC, _ := multiplicities(c.notes, notes, note.Note.Equal)
	present := make([]bool, len(inC))

	for id, count := range inC {
		present[id] = count > 0
	}

	var added note.Notes

	for k, n := range notes {
		id := ids[len(c.notes)+k]
		if !present[id] {
			present[id] = true
			added = append(added, n)
		}
	}

	if len(added) == 0 {
		return nil
	}

	out := make(note.Notes, len(c.notes)+len(added))
	copy(out, c.notes)
	copy(out[len(c.notes):], added)
	c.put(out)

	return nil
}

// Insert puts the notes of the list into c starting at number i, in the order
// of the list; the notes with numbers from i move after them. i == Len()
// appends. Without notes c does not change, but i is checked.
//
// Errors, in this order: ErrUnset for the chord, then for the first unset note
// of the list; ErrIndex for i < 0 or i > Len().
func (c *Chord) Insert(i int, notes ...note.Note) error {
	if err := c.check("Insert"); err != nil {
		return err
	}

	if err := checkNotes("Insert", notes); err != nil {
		return err
	}

	if i < 0 || i > len(c.notes) {
		return indexError("Insert", i, len(c.notes))
	}

	if len(notes) == 0 {
		return nil
	}

	out := make(note.Notes, len(c.notes)+len(notes))
	copy(out, c.notes[:i])
	copy(out[i:], notes)
	copy(out[i+len(notes):], c.notes[i:])
	c.put(out)

	return nil
}

// Remove removes the first occurrence of n from c.
//
// Errors, in this order: ErrUnset for the chord, then for n; ErrNotFound.
func (c *Chord) Remove(n note.Note) error {
	if err := c.checkWith("Remove", n); err != nil {
		return err
	}

	i, _, ok := c.Find(n)
	if !ok {
		return notFoundError("Remove")
	}

	c.put(without(c.notes, i))

	return nil
}

// RemoveAll removes all the occurrences of n from c.
//
// Errors, in this order: ErrUnset for the chord, then for n; ErrNotFound.
func (c *Chord) RemoveAll(n note.Note) error {
	if err := c.checkWith("RemoveAll", n); err != nil {
		return err
	}

	out := make(note.Notes, 0, len(c.notes))

	for _, x := range c.notes {
		if !x.Equal(n) {
			out = append(out, x)
		}
	}

	if len(out) == len(c.notes) {
		return notFoundError("RemoveAll")
	}

	c.put(out)

	return nil
}

// RemoveAt removes the note with number i from c.
//
// Errors, in this order: ErrUnset for the chord; ErrIndex for i < 0 or
// i >= Len().
func (c *Chord) RemoveAt(i int) error {
	if err := c.check("RemoveAt"); err != nil {
		return err
	}

	if i < 0 || i >= len(c.notes) {
		return indexError("RemoveAt", i, len(c.notes))
	}

	c.put(without(c.notes, i))

	return nil
}

// RemoveFunc removes all the notes of c for which cond gives true. cond gets
// copies of all the notes in the order of record, once for each note, until
// its first error. When nothing is to be removed, c does not change; this is
// not an error.
//
// Errors, in this order: ErrUnset for the chord, then for a nil cond;
// ErrRuleFailed with the error of cond.
func (c *Chord) RemoveFunc(cond Condition) error {
	return c.filter("RemoveFunc", cond, false)
}

// KeepFunc keeps in c only the notes for which cond gives true: it removes
// all the notes for which cond gives false. cond is called as by RemoveFunc.
//
// Errors, in this order: ErrUnset for the chord, then for a nil cond;
// ErrRuleFailed with the error of cond.
func (c *Chord) KeepFunc(cond Condition) error {
	return c.filter("KeepFunc", cond, true)
}

// Replace puts y in the place of the first occurrence of x in c.
//
// Errors, in this order: ErrUnset for the chord, then for x, then for y;
// ErrNotFound for x.
func (c *Chord) Replace(x, y note.Note) error {
	if err := c.checkWith("Replace", x, y); err != nil {
		return err
	}

	i, _, ok := c.Find(x)
	if !ok {
		return notFoundError("Replace")
	}

	out := c.Notes()
	out[i] = y
	c.put(out)

	return nil
}

// ReplaceAll puts y in the place of every occurrence of x in c, as
// strings.ReplaceAll does.
//
// Errors, in this order: ErrUnset for the chord, then for x, then for y;
// ErrNotFound for x.
func (c *Chord) ReplaceAll(x, y note.Note) error {
	if err := c.checkWith("ReplaceAll", x, y); err != nil {
		return err
	}

	out := c.Notes()
	found := false

	for i := range out {
		if out[i].Equal(x) {
			out[i] = y
			found = true
		}
	}

	if !found {
		return notFoundError("ReplaceAll")
	}

	c.put(out)

	return nil
}

// ReplaceAt puts y in the place of the note with number i.
//
// Errors, in this order: ErrUnset for the chord, then for y; ErrIndex for
// i < 0 or i >= Len().
func (c *Chord) ReplaceAt(i int, y note.Note) error {
	if err := c.checkWith("ReplaceAt", y); err != nil {
		return err
	}

	if i < 0 || i >= len(c.notes) {
		return indexError("ReplaceAt", i, len(c.notes))
	}

	out := c.Notes()
	out[i] = y
	c.put(out)

	return nil
}

// ReplaceFunc puts in the place of every note of c the note that t returns
// for a copy of it. t is called in the order of record, once for each note,
// until its first error. Either every note is replaced, or none.
//
// Errors, in this order: ErrUnset for the chord, then for a nil t; then, note
// by note in the order of record: ErrRuleFailed with the error of t, or
// ErrUnset for an unset note returned by t.
func (c *Chord) ReplaceFunc(t Transform) error {
	if err := c.check("ReplaceFunc"); err != nil {
		return err
	}

	if t == nil {
		return unsetError("ReplaceFunc", "nil transform")
	}

	if len(c.notes) == 0 {
		return nil
	}

	out := make(note.Notes, len(c.notes))

	for i, n := range c.notes {
		r, err := transform("ReplaceFunc", i, n, t)
		if err != nil {
			return err
		}

		out[i] = r
	}

	c.put(out)

	return nil
}

// ReplaceAtFunc puts in the place of the note with number i the note that t
// returns for a copy of it. t is called once, after i is checked.
//
// Errors, in this order: ErrUnset for the chord, then for a nil t; ErrIndex
// for i < 0 or i >= Len(); ErrRuleFailed with the error of t, or ErrUnset for
// an unset note returned by t.
func (c *Chord) ReplaceAtFunc(i int, t Transform) error {
	if err := c.check("ReplaceAtFunc"); err != nil {
		return err
	}

	if t == nil {
		return unsetError("ReplaceAtFunc", "nil transform")
	}

	if i < 0 || i >= len(c.notes) {
		return indexError("ReplaceAtFunc", i, len(c.notes))
	}

	r, err := transform("ReplaceAtFunc", i, c.notes[i], t)
	if err != nil {
		return err
	}

	out := c.Notes()
	out[i] = r
	c.put(out)

	return nil
}

// Clear removes all the notes from c: it becomes the empty chord, still set.
//
// Errors: ErrUnset for the chord.
func (c *Chord) Clear() error {
	if err := c.check("Clear"); err != nil {
		return err
	}

	c.notes = nil

	return nil
}

// RemoveRepeats keeps in c only the first occurrence of every note, in the
// order of record.
//
// Errors: ErrUnset for the chord.
func (c *Chord) RemoveRepeats() error {
	if err := c.check("RemoveRepeats"); err != nil {
		return err
	}

	distinct := c.DistinctNotes()
	if len(distinct) == len(c.notes) {
		return nil
	}

	c.put(distinct)

	return nil
}

// Union makes the multiplicity of every note in c the greater of its
// multiplicities in c and other. The notes of c stay in their places; then,
// in the order of record of other, every occurrence that c lacks is appended
// with its own source: for a note that enters c p times and other q > p
// times, its occurrences in other from the (p+1)-th to the q-th. For chords
// without repeats it is the usual union. other does not change; it can be a
// copy of c, and then c does not change.
//
// Errors: ErrUnset for the chord, then for other.
func (c *Chord) Union(other Chord) error {
	if err := c.checkOther("Union", other); err != nil {
		return err
	}

	ids, inC, _ := multiplicities(c.notes, other.notes, note.Note.Equal)
	seen := make([]int, len(inC))

	var added note.Notes

	for k, n := range other.notes {
		id := ids[len(c.notes)+k]

		seen[id]++
		if seen[id] > inC[id] {
			added = append(added, n)
		}
	}

	if len(added) == 0 {
		return nil
	}

	out := make(note.Notes, len(c.notes)+len(added))
	copy(out, c.notes)
	copy(out[len(c.notes):], added)
	c.put(out)

	return nil
}

// Intersect makes the multiplicity of every note in c the smaller of its
// multiplicities in c and other: the first occurrences in the order of record
// of c stay, with the sources of the notes of c, and the last occurrences
// beyond the multiplicity in other are removed. other does not change; it can
// be a copy of c, and then c does not change.
//
// Errors: ErrUnset for the chord, then for other.
func (c *Chord) Intersect(other Chord) error {
	if err := c.checkOther("Intersect", other); err != nil {
		return err
	}

	ids, _, inOther := multiplicities(c.notes, other.notes, note.Note.Equal)
	kept := make([]int, len(inOther))
	out := make(note.Notes, 0, len(c.notes))

	for k, n := range c.notes {
		id := ids[k]
		if kept[id] < inOther[id] {
			kept[id]++
			out = append(out, n)
		}
	}

	if len(out) == len(c.notes) {
		return nil
	}

	c.put(out)

	return nil
}

// Subtract removes from c, for every note of other in its order of record,
// the first remaining occurrence of that note; a note that c lacks, or no
// longer has, is skipped. The multiplicity of every note becomes the
// difference of its multiplicities, but not less than 0. other does not
// change; subtracting a copy of c makes c empty.
//
// Errors: ErrUnset for the chord, then for other.
func (c *Chord) Subtract(other Chord) error {
	if err := c.checkOther("Subtract", other); err != nil {
		return err
	}

	ids, _, inOther := multiplicities(c.notes, other.notes, note.Note.Equal)
	removed := make([]int, len(inOther))
	out := make(note.Notes, 0, len(c.notes))

	for k, n := range c.notes {
		id := ids[k]
		if removed[id] < inOther[id] {
			removed[id]++

			continue
		}

		out = append(out, n)
	}

	if len(out) == len(c.notes) {
		return nil
	}

	c.put(out)

	return nil
}

// Sort puts the notes of c in the agreed order of notes (note.Note.Compare).
// The sort is stable: notes equal by Compare keep their order.
//
// Errors: ErrUnset for the chord.
func (c *Chord) Sort() error {
	if err := c.check("Sort"); err != nil {
		return err
	}

	if len(c.notes) < 2 {
		return nil
	}

	c.put(c.SortedNotes())

	return nil
}

// SortFunc puts the notes of c in the order of order. The sort is stable:
// notes that order finds equal keep their order. The number and order of the
// calls of order are not defined; after its first error it is not called
// again. When order is not consistent, that is, it gives no strict weak
// order, the resulting order is not defined, but the notes of c stay the
// same.
//
// Errors, in this order: ErrUnset for the chord, then for a nil order;
// ErrRuleFailed with the error of order.
func (c *Chord) SortFunc(order Comparison) error {
	if err := c.check("SortFunc"); err != nil {
		return err
	}

	if order == nil {
		return unsetError("SortFunc", "nil comparison")
	}

	out := c.Notes()

	var failure error

	slices.SortStableFunc(out, func(a, b note.Note) int {
		if failure != nil {
			return 0
		}

		r, err := order(a, b)
		if err != nil {
			failure = err

			return 0
		}

		return r
	})

	if failure != nil {
		return ruleError("SortFunc", -1, failure)
	}

	if len(out) < 2 {
		return nil
	}

	c.put(out)

	return nil
}

// Reverse puts the notes of c in the reverse order of record.
//
// Errors: ErrUnset for the chord.
func (c *Chord) Reverse() error {
	if err := c.check("Reverse"); err != nil {
		return err
	}

	if len(c.notes) < 2 {
		return nil
	}

	out := c.Notes()
	slices.Reverse(out)
	c.put(out)

	return nil
}

// Shift moves every note of c by k octaves up (note.Note.Shift); a negative k
// moves them down. The corrections, the sources and the order of record stay.
// Every note is checked by note.Note.Shift for any k, 0 included, so a chord
// with a note without an octave fails even for Shift(0). The empty chord does
// not change and does not fail.
//
// Errors: ErrUnset for the chord; then the error of package note for the
// first note, in the order of record, that cannot be moved: note.ErrNoOctave,
// or note.ErrOutOfRange together with octave.ErrOutOfRange. It matches no
// sentinel of chord.
func (c *Chord) Shift(k int) error {
	return c.move("Shift", func(n *note.Note) error { return n.Shift(k) })
}

// ShiftDown moves every note of c by k octaves down (note.Note.ShiftDown); a
// negative k moves them up. ShiftDown(math.MinInt) is exact. See Shift.
func (c *Chord) ShiftDown(k int) error {
	return c.move("ShiftDown", func(n *note.Note) error { return n.ShiftDown(k) })
}

// Up moves every note of c one octave up (note.Note.Up). See Shift.
func (c *Chord) Up() error {
	return c.move("Up", (*note.Note).Up)
}

// Down moves every note of c one octave down (note.Note.Down). See Shift.
func (c *Chord) Down() error {
	return c.move("Down", (*note.Note).Down)
}

// checkWith reports ErrUnset for the receiver of op, then for the first unset
// note of notes.
func (c *Chord) checkWith(op string, notes ...note.Note) error {
	if err := c.check(op); err != nil {
		return err
	}

	for _, n := range notes {
		if !n.IsValid() {
			return unsetError(op, "unset note")
		}
	}

	return nil
}

// checkOther reports ErrUnset for the receiver of op, then for the chord
// argument other.
func (c *Chord) checkOther(op string, other Chord) error {
	if err := c.check(op); err != nil {
		return err
	}

	if !other.set {
		return unsetError(op, "unset chord argument")
	}

	return nil
}

// filter removes the notes of c for which cond gives the answer other than
// keep: with keep false, it removes the notes for which cond gives true.
func (c *Chord) filter(op string, cond Condition, keep bool) error {
	if err := c.check(op); err != nil {
		return err
	}

	selected, err := c.selectFunc(op, cond)
	if err != nil {
		return err
	}

	out := make(note.Notes, 0, len(c.notes))

	for i, n := range c.notes {
		if selected[i] == keep {
			out = append(out, n)
		}
	}

	if len(out) == len(c.notes) {
		return nil
	}

	c.put(out)

	return nil
}

// move applies the change of package note to a copy of every note of c, in
// the order of record, and stores the copies only when every change succeeds.
func (c *Chord) move(op string, change func(n *note.Note) error) error {
	if err := c.check(op); err != nil {
		return err
	}

	if len(c.notes) == 0 {
		return nil
	}

	out := c.Notes()
	for i := range out {
		if err := change(&out[i]); err != nil {
			return noteError(op, i, err)
		}
	}

	c.put(out)

	return nil
}

// transform calls t for n, the note with number i, and checks the result.
func transform(op string, i int, n note.Note, t Transform) (note.Note, error) {
	r, err := t(n)
	if err != nil {
		return note.Note{}, ruleError(op, i, err)
	}

	if !r.IsValid() {
		return note.Note{}, unsetError(op, "unset note returned by the transform")
	}

	return r, nil
}

// without returns a new array with the notes except the one with number i.
func without(notes note.Notes, i int) note.Notes {
	out := make(note.Notes, len(notes)-1)
	copy(out, notes[:i])
	copy(out[i:], notes[i+1:])

	return out
}
