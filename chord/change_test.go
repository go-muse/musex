package chord_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-4. Changing the composition: ОП-21…ОП-31; П-4, П-5, Р-9, Р-11; К-5,
// К-16…К-22, К-32 (removing and keeping), К-33, К-43; architecture 5.10.

// change is one changing operation on a chord and its expected outcome: the
// notes of the chord after success, or the sentinel of the failure (then the
// chord is the same).
type change struct {
	name string
	base []note.Note
	op   func(c *chord.Chord) error
	want []note.Note
	err  error
}

// runChanges runs each change on a new chord with its base notes. A failure
// is checked with assertOnly, or with assertRuleFailed for ErrRuleFailed.
func runChanges(t *testing.T, tests []change) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, tt.base...)
			before := stateOf(c)
			err := tt.op(&c)
			switch {
			case tt.err == nil:
				require.NoError(t, err)
				assertChord(t, tt.want, c)
			case tt.err == chord.ErrRuleFailed:
				assertRuleFailed(t, err, errRule)
				assertState(t, before, c, "the chord is the same after a failure")
			default:
				assertOnly(t, err, tt.err)
				assertState(t, before, c, "the chord is the same after a failure")
			}
		})
	}
}

func TestChangeAdd(t *testing.T) {
	t.Parallel()
	st := std()
	g4st := pn(g4.Spelling(), 4, 0, st)
	base := list(c4, e4)
	runChanges(t, []change{
		{name: "K-16 add G4", base: base, op: func(c *chord.Chord) error { return c.Add(g4) }, want: list(c4, e4, g4)},
		{name: "K-16 add E4: repeats are allowed", base: base, op: func(c *chord.Chord) error { return c.Add(e4) }, want: list(c4, e4, e4)},
		{name: "K-16 add G4 and B4", base: base, op: func(c *chord.Chord) error { return c.Add(g4, b4) }, want: list(c4, e4, g4, b4)},
		{name: "K-16 add an empty list", base: base, op: func(c *chord.Chord) error { return c.Add() }, want: base},
		{name: "add a nil slice", base: base, op: func(c *chord.Chord) error { return c.Add([]note.Note(nil)...) }, want: base},
		{name: "the note keeps its source", base: base, op: func(c *chord.Chord) error { return c.Add(g4st) }, want: list(c4, e4, g4st)},
		{name: "add to the empty chord", base: list(), op: func(c *chord.Chord) error { return c.Add(c4) }, want: list(c4)},
		{name: "K-17 add G4 and an unset note", base: base, op: func(c *chord.Chord) error { return c.Add(g4, note.Note{}) }, err: chord.ErrUnset},
		{name: "an unset note first", base: base, op: func(c *chord.Chord) error { return c.Add(note.Note{}, g4) }, err: chord.ErrUnset},
	})
}

func TestChangeAddMissing(t *testing.T) {
	t.Parallel()
	st := std()
	s19 := set19()
	c4st := pn(c4.Spelling(), 4, 0, st)
	g419 := pn(g4.Spelling(), 4, 0, s19)
	base := list(c4, e4)
	runChanges(t, []change{
		{name: "K-16 add missing E4 G4 G4", base: base, op: func(c *chord.Chord) error { return c.AddMissing(e4, g4, g4) }, want: list(c4, e4, g4)},
		{name: "add missing: empty list", base: base, op: func(c *chord.Chord) error { return c.AddMissing() }, want: base},
		{name: "a correction makes another note", base: base, op: func(c *chord.Chord) error { return c.AddMissing(c4p5) }, want: list(c4, e4, c4p5)},
		{name: "a note with another source is present", base: list(c4st), op: func(c *chord.Chord) error { return c.AddMissing(pn(c4.Spelling(), 4, 0, s19)) }, want: list(c4st)},
		{name: "the first of the repeats in the list with its source", base: base, op: func(c *chord.Chord) error { return c.AddMissing(g419, g4) }, want: list(c4, e4, g419)},
		{name: "order of the list", base: list(), op: func(c *chord.Chord) error { return c.AddMissing(g4, c4, g4, e4) }, want: list(g4, c4, e4)},
		{name: "repeats of the chord stay", base: list(c4, c4), op: func(c *chord.Chord) error { return c.AddMissing(c4, e4) }, want: list(c4, c4, e4)},
		{name: "an unset note", base: base, op: func(c *chord.Chord) error { return c.AddMissing(g4, note.Note{}) }, err: chord.ErrUnset},
	})
}

func TestChangeInsert(t *testing.T) {
	t.Parallel()
	base := list(c4, g4)
	runChanges(t, []change{
		{name: "K-18 insert E4 at 1", base: base, op: func(c *chord.Chord) error { return c.Insert(1, e4) }, want: list(c4, e4, g4)},
		{name: "K-18 insert E4 at 0", base: base, op: func(c *chord.Chord) error { return c.Insert(0, e4) }, want: list(e4, c4, g4)},
		{name: "K-18 insert E4 at 2: at the end", base: base, op: func(c *chord.Chord) error { return c.Insert(2, e4) }, want: list(c4, g4, e4)},
		{name: "K-18 insert D4 E4 at 1", base: base, op: func(c *chord.Chord) error { return c.Insert(1, d4, e4) }, want: list(c4, d4, e4, g4)},
		{name: "K-18 insert E4 at 3", base: base, op: func(c *chord.Chord) error { return c.Insert(3, e4) }, err: chord.ErrIndex},
		{name: "K-18 insert E4 at -1", base: base, op: func(c *chord.Chord) error { return c.Insert(-1, e4) }, err: chord.ErrIndex},
		{name: "insert a repeat", base: base, op: func(c *chord.Chord) error { return c.Insert(1, g4) }, want: list(c4, g4, g4)},
		{name: "insert no notes at 1", base: base, op: func(c *chord.Chord) error { return c.Insert(1) }, want: base},
		{name: "insert no notes at 2", base: base, op: func(c *chord.Chord) error { return c.Insert(2) }, want: base},
		{name: "insert no notes at 3: the number is always checked", base: base, op: func(c *chord.Chord) error { return c.Insert(3) }, err: chord.ErrIndex},
		{name: "insert no notes at -1", base: base, op: func(c *chord.Chord) error { return c.Insert(-1) }, err: chord.ErrIndex},
		{name: "insert at math.MaxInt", base: base, op: func(c *chord.Chord) error { return c.Insert(math.MaxInt, e4) }, err: chord.ErrIndex},
		{name: "insert at math.MinInt", base: base, op: func(c *chord.Chord) error { return c.Insert(math.MinInt, e4) }, err: chord.ErrIndex},
		{name: "insert into the empty chord at 0", base: list(), op: func(c *chord.Chord) error { return c.Insert(0, c4) }, want: list(c4)},
		{name: "insert into the empty chord at 1", base: list(), op: func(c *chord.Chord) error { return c.Insert(1, c4) }, err: chord.ErrIndex},
		{name: "insert an unset note", base: base, op: func(c *chord.Chord) error { return c.Insert(0, note.Note{}) }, err: chord.ErrUnset},
		{name: "K-43 insert an unset note at 9", base: list(c4, e4), op: func(c *chord.Chord) error { return c.Insert(9, note.Note{}) }, err: chord.ErrUnset},
	})
}

func TestChangeRemove(t *testing.T) {
	t.Parallel()
	st := std()
	s19 := set19()
	c4st := pn(c4.Spelling(), 4, 0, st)
	c419 := pn(c4.Spelling(), 4, 0, s19)
	base := list(c4, e4, c4, g4)
	runChanges(t, []change{
		{name: "K-19 remove C4", base: base, op: func(c *chord.Chord) error { return c.Remove(c4) }, want: list(e4, c4, g4)},
		{name: "K-19 remove all C4", base: base, op: func(c *chord.Chord) error { return c.RemoveAll(c4) }, want: list(e4, g4)},
		{name: "K-19 remove D4", base: base, op: func(c *chord.Chord) error { return c.Remove(d4) }, err: chord.ErrNotFound},
		{name: "K-19 remove all D4", base: base, op: func(c *chord.Chord) error { return c.RemoveAll(d4) }, err: chord.ErrNotFound},
		{name: "K-19 remove at 1", base: base, op: func(c *chord.Chord) error { return c.RemoveAt(1) }, want: list(c4, c4, g4)},
		{name: "K-19 remove at 4", base: base, op: func(c *chord.Chord) error { return c.RemoveAt(4) }, err: chord.ErrIndex},
		{name: "remove at -1", base: base, op: func(c *chord.Chord) error { return c.RemoveAt(-1) }, err: chord.ErrIndex},
		{name: "remove at 0", base: base, op: func(c *chord.Chord) error { return c.RemoveAt(0) }, want: list(e4, c4, g4)},
		{name: "remove at 3", base: base, op: func(c *chord.Chord) error { return c.RemoveAt(3) }, want: list(c4, e4, c4)},
		{name: "remove C4+5: a correction makes another note", base: base, op: func(c *chord.Chord) error { return c.Remove(c4p5) }, err: chord.ErrNotFound},
		{name: "R-11 remove removes the first occurrence", base: list(c4st, e4, c419), op: func(c *chord.Chord) error { return c.Remove(c4) }, want: list(e4, c419)},
		{name: "remove by a note with another source", base: list(c4st, e4), op: func(c *chord.Chord) error { return c.Remove(c419) }, want: list(e4)},
		{name: "K-43 remove an unset note", base: list(c4, e4), op: func(c *chord.Chord) error { return c.Remove(note.Note{}) }, err: chord.ErrUnset},
		{name: "remove all of an unset note", base: list(c4, e4), op: func(c *chord.Chord) error { return c.RemoveAll(note.Note{}) }, err: chord.ErrUnset},
		{name: "G-2 remove the last note", base: list(c4), op: func(c *chord.Chord) error { return c.Remove(c4) }, want: list()},
		{name: "remove all: the chord becomes empty", base: list(c4, c4), op: func(c *chord.Chord) error { return c.RemoveAll(c4) }, want: list()},
		{name: "remove at 0 of one note", base: list(c4), op: func(c *chord.Chord) error { return c.RemoveAt(0) }, want: list()},
		{name: "remove from the empty chord", base: list(), op: func(c *chord.Chord) error { return c.Remove(c4) }, err: chord.ErrNotFound},
		{name: "remove all from the empty chord", base: list(), op: func(c *chord.Chord) error { return c.RemoveAll(c4) }, err: chord.ErrNotFound},
		{name: "remove at 0 of the empty chord", base: list(), op: func(c *chord.Chord) error { return c.RemoveAt(0) }, err: chord.ErrIndex},
	})
}

func TestChangeK20RemoveThenAdd(t *testing.T) {
	t.Parallel()
	c := mk(t, c4)
	require.NoError(t, c.Remove(c4))
	assertChord(t, list(), c, "after removing: the empty set chord")
	require.NoError(t, c.Add(e4))
	assertChord(t, list(e4), c)
}

func TestChangeRemoveKeepFunc(t *testing.T) {
	t.Parallel()
	nameD := func(n note.Note) (bool, error) { return n.SameLetter(d4), nil }
	failing := func(note.Note) (bool, error) { return false, errRule }
	all := func(note.Note) (bool, error) { return true, nil }
	none := func(note.Note) (bool, error) { return false, nil }
	base := list(c4, eNo, g4)
	runChanges(t, []change{
		{name: "K-32 remove the notes with an octave", base: base, op: func(c *chord.Chord) error { return c.RemoveFunc(hasOctave) }, want: list(eNo)},
		{name: "K-32 keep the notes with an octave", base: base, op: func(c *chord.Chord) error { return c.KeepFunc(hasOctave) }, want: list(c4, g4)},
		{name: "K-32 remove: nil condition", base: base, op: func(c *chord.Chord) error { return c.RemoveFunc(nil) }, err: chord.ErrUnset},
		{name: "K-32 keep: nil condition", base: base, op: func(c *chord.Chord) error { return c.KeepFunc(nil) }, err: chord.ErrUnset},
		{name: "K-32 remove: failing condition", base: base, op: func(c *chord.Chord) error { return c.RemoveFunc(failing) }, err: chord.ErrRuleFailed},
		{name: "K-32 keep: failing condition", base: base, op: func(c *chord.Chord) error { return c.KeepFunc(failing) }, err: chord.ErrRuleFailed},
		{name: "K-33 nothing to remove is not a failure", base: list(c4, e4), op: func(c *chord.Chord) error { return c.RemoveFunc(nameD) }, want: list(c4, e4)},
		{name: "remove all notes", base: base, op: func(c *chord.Chord) error { return c.RemoveFunc(all) }, want: list()},
		{name: "keep all notes", base: base, op: func(c *chord.Chord) error { return c.KeepFunc(all) }, want: base},
		{name: "keep no notes", base: base, op: func(c *chord.Chord) error { return c.KeepFunc(none) }, want: list()},
		{name: "remove: the empty chord", base: list(), op: func(c *chord.Chord) error { return c.RemoveFunc(all) }, want: list()},
		{name: "keep: the empty chord", base: list(), op: func(c *chord.Chord) error { return c.KeepFunc(none) }, want: list()},
		{name: "remove keeps the order of the others", base: list(g4, c4, e4, c4), op: func(c *chord.Chord) error {
			return c.RemoveFunc(func(n note.Note) (bool, error) { return n.Equal(c4), nil })
		}, want: list(g4, e4)},
	})
}

func TestChangeReplace(t *testing.T) {
	t.Parallel()
	st := std()
	d4st := pn(d4.Spelling(), 4, 0, st)
	base := list(c4, e4, c4)
	runChanges(t, []change{
		{name: "K-21 replace C4 with D4", base: base, op: func(c *chord.Chord) error { return c.Replace(c4, d4) }, want: list(d4, e4, c4)},
		{name: "K-21 replace all C4 with D4", base: base, op: func(c *chord.Chord) error { return c.ReplaceAll(c4, d4) }, want: list(d4, e4, d4)},
		{name: "K-21 replace F4 with D4", base: base, op: func(c *chord.Chord) error { return c.Replace(f4, d4) }, err: chord.ErrNotFound},
		{name: "K-21 replace all F4 with D4", base: base, op: func(c *chord.Chord) error { return c.ReplaceAll(f4, d4) }, err: chord.ErrNotFound},
		{name: "K-21 replace at 1 with F4", base: base, op: func(c *chord.Chord) error { return c.ReplaceAt(1, f4) }, want: list(c4, f4, c4)},
		{name: "K-21 replace C4 with an unset note", base: base, op: func(c *chord.Chord) error { return c.Replace(c4, note.Note{}) }, err: chord.ErrUnset},
		{name: "K-21 replace at 3 with F4", base: base, op: func(c *chord.Chord) error { return c.ReplaceAt(3, f4) }, err: chord.ErrIndex},
		{name: "replace at -1", base: base, op: func(c *chord.Chord) error { return c.ReplaceAt(-1, f4) }, err: chord.ErrIndex},
		{name: "K-43 replace at 5 with an unset note", base: list(c4, e4), op: func(c *chord.Chord) error { return c.ReplaceAt(5, note.Note{}) }, err: chord.ErrUnset},
		{name: "replace an unset note", base: base, op: func(c *chord.Chord) error { return c.Replace(note.Note{}, d4) }, err: chord.ErrUnset},
		{name: "replace all with an unset note", base: base, op: func(c *chord.Chord) error { return c.ReplaceAll(c4, note.Note{}) }, err: chord.ErrUnset},
		{name: "replace at 0 with an unset note", base: base, op: func(c *chord.Chord) error { return c.ReplaceAt(0, note.Note{}) }, err: chord.ErrUnset},
		{name: "the replacement keeps its source", base: base, op: func(c *chord.Chord) error { return c.Replace(c4, d4st) }, want: list(d4st, e4, c4)},
		{name: "replace at keeps the source of y", base: base, op: func(c *chord.Chord) error { return c.ReplaceAt(2, d4st) }, want: list(c4, e4, d4st)},
		{name: "the search ignores sources", base: list(pn(c4.Spelling(), 4, 0, st), e4), op: func(c *chord.Chord) error {
			return c.Replace(pn(c4.Spelling(), 4, 0, set19()), d4)
		}, want: list(d4, e4)},
		{name: "replace with the same note", base: base, op: func(c *chord.Chord) error { return c.Replace(e4, e4) }, want: base},
		{name: "replace C4+5: a correction makes another note", base: base, op: func(c *chord.Chord) error { return c.Replace(c4p5, d4) }, err: chord.ErrNotFound},
		{name: "replace in the empty chord", base: list(), op: func(c *chord.Chord) error { return c.Replace(c4, d4) }, err: chord.ErrNotFound},
		{name: "replace at 0 in the empty chord", base: list(), op: func(c *chord.Chord) error { return c.ReplaceAt(0, d4) }, err: chord.ErrIndex},
	})
}

func TestChangeClearAndRemoveRepeats(t *testing.T) {
	t.Parallel()
	st := std()
	s19 := set19()
	c4st := pn(c4.Spelling(), 4, 0, st)
	c419 := pn(c4.Spelling(), 4, 0, s19)
	base := list(c4st, e4, c419, e4) // K-22
	runChanges(t, []change{
		{name: "K-22 remove repeats", base: base, op: func(c *chord.Chord) error { return c.RemoveRepeats() }, want: list(c4st, e4)},
		{name: "K-22 clear", base: base, op: func(c *chord.Chord) error { return c.Clear() }, want: list()},
		{name: "remove repeats without repeats", base: list(g4, c4, c4p5, cNo), op: func(c *chord.Chord) error { return c.RemoveRepeats() }, want: list(g4, c4, c4p5, cNo)},
		{name: "remove repeats keeps the order of the first occurrences", base: list(g4, c4, g4, e4, c4), op: func(c *chord.Chord) error { return c.RemoveRepeats() }, want: list(g4, c4, e4)},
		{name: "remove repeats of the empty chord", base: list(), op: func(c *chord.Chord) error { return c.RemoveRepeats() }, want: list()},
		{name: "clear the empty chord", base: list(), op: func(c *chord.Chord) error { return c.Clear() }, want: list()},
	})
}

// TestChangeRemoveRepeatsProperty checks RemoveRepeats on every chord of up to
// 3 notes of the pool: no repeats remain, and the notes are DistinctNotes of
// the chord before (plan, "Properties").
func TestChangeRemoveRepeatsProperty(t *testing.T) {
	t.Parallel()
	for _, s := range sequences(pool(), 3) {
		c := mk(t, s...)
		distinct := c.DistinctNotes()
		if !assertNotes(t, distinctOf(s), distinct, "DistinctNotes of %v", s) {
			return
		}
		require.NoError(t, c.RemoveRepeats())
		if !assert.False(t, c.HasRepeats(), "no repeats after RemoveRepeats of %v", s) ||
			!assertChord(t, distinct, c, "RemoveRepeats of %v", s) {
			return
		}
	}
}

func TestChangeUnsetChord(t *testing.T) {
	t.Parallel()
	// K-5, СЕ-2: every changing method of the unset chord reports E-1; the
	// chord stays unset.
	for name, op := range changingOps() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var c chord.Chord
			assertOnly(t, op(&c), chord.ErrUnset)
			assertUnsetChord(t, c)
		})
	}
}

func TestChangeNilPointer(t *testing.T) {
	t.Parallel()
	// Architecture 5.2: a changing method of a nil *Chord reports E-1.
	for name, op := range changingOps() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var p *chord.Chord
			assertOnly(t, op(p), chord.ErrUnset)
		})
	}
}

// changingOps returns every changing method with valid arguments, keyed by
// its name (architecture 5.1).
func changingOps() map[string]func(c *chord.Chord) error {
	other := chord.MustNew(c4, e4)
	return map[string]func(c *chord.Chord) error{
		"Add":           func(c *chord.Chord) error { return c.Add(c4) },
		"AddMissing":    func(c *chord.Chord) error { return c.AddMissing(c4) },
		"Insert":        func(c *chord.Chord) error { return c.Insert(0, c4) },
		"Remove":        func(c *chord.Chord) error { return c.Remove(c4) },
		"RemoveAll":     func(c *chord.Chord) error { return c.RemoveAll(c4) },
		"RemoveAt":      func(c *chord.Chord) error { return c.RemoveAt(0) },
		"RemoveFunc":    func(c *chord.Chord) error { return c.RemoveFunc(hasOctave) },
		"KeepFunc":      func(c *chord.Chord) error { return c.KeepFunc(hasOctave) },
		"Replace":       func(c *chord.Chord) error { return c.Replace(c4, d4) },
		"ReplaceAll":    func(c *chord.Chord) error { return c.ReplaceAll(c4, d4) },
		"ReplaceAt":     func(c *chord.Chord) error { return c.ReplaceAt(0, d4) },
		"ReplaceFunc":   func(c *chord.Chord) error { return c.ReplaceFunc(up) },
		"ReplaceAtFunc": func(c *chord.Chord) error { return c.ReplaceAtFunc(0, up) },
		"Clear":         func(c *chord.Chord) error { return c.Clear() },
		"RemoveRepeats": func(c *chord.Chord) error { return c.RemoveRepeats() },
		"Union":         func(c *chord.Chord) error { return c.Union(other) },
		"Intersect":     func(c *chord.Chord) error { return c.Intersect(other) },
		"Subtract":      func(c *chord.Chord) error { return c.Subtract(other) },
		"Sort":          func(c *chord.Chord) error { return c.Sort() },
		"SortFunc":      func(c *chord.Chord) error { return c.SortFunc(byRecord) },
		"Reverse":       func(c *chord.Chord) error { return c.Reverse() },
		"Shift":         func(c *chord.Chord) error { return c.Shift(1) },
		"ShiftDown":     func(c *chord.Chord) error { return c.ShiftDown(1) },
		"Up":            func(c *chord.Chord) error { return c.Up() },
		"Down":          func(c *chord.Chord) error { return c.Down() },
	}
}
