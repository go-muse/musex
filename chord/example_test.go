package chord_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
)

// unbound returns the note with the name s in the octave with the ASPN number
// aspn, without a correction and without a source of settings.
func unbound(s spelling.Spelling, aspn int) note.Note {
	return note.MustFromPitch(pitch.MustFromASPN(s, aspn), cents.Zero(), nil)
}

// A chord is made from notes; it is changed in place, and a copy is made by
// assignment.
func ExampleNew() {
	c4, e4, g4 := unbound(spelling.C(), 4), unbound(spelling.E(), 4), unbound(spelling.G(), 4)

	c := chord.MustNew(c4, e4, g4)
	fmt.Println(c.Len(), c.Contains(e4))

	c.MustAdd(e4).MustRemove(c4) // c is now [E4, G4, E4]
	fmt.Println(c.Count(e4))

	d := c        // an independent copy
	d.MustClear() // c does not change
	fmt.Println(c.Len(), d.IsEmpty(), d.IsValid())
	// Output:
	// 3 true
	// 2
	// 3 true true
}

// The zero value is the unset chord: it takes no notes. New without notes
// gives the empty chord to work with.
func Example_unset() {
	c4 := unbound(spelling.C(), 4)

	var u chord.Chord
	fmt.Println(u.IsValid(), u.IsEmpty())
	fmt.Println(errors.Is(u.Add(c4), chord.ErrUnset))
	fmt.Println(u)

	empty := chord.MustNew()
	fmt.Println(empty.IsValid(), empty.IsEmpty(), empty)
	empty.MustAdd(c4)
	fmt.Println(empty)
	// Output:
	// false true
	// true
	// Chord(unset)
	// true true []
	// [C ASPN 4, 0 cents, unbound]
}

// A note may enter a chord several times; the chord keeps the record order.
func Example_repeatsAndOrder() {
	c4, e4 := unbound(spelling.C(), 4), unbound(spelling.E(), 4)

	c := chord.MustNew(c4, e4, c4)
	fmt.Println(c.Len(), c.Count(c4), c.Indices(c4), c.HasRepeats())
	i, n, ok := c.Find(e4)
	fmt.Println(i, n, ok)
	fmt.Println(len(c.DistinctNotes()))

	c.MustRemoveRepeats()
	fmt.Println(c)
	// Output:
	// 3 2 [0 2] true
	// 1 E ASPN 4, 0 cents, unbound true
	// 2
	// [C ASPN 4, 0 cents, unbound; E ASPN 4, 0 cents, unbound]
}

// Three equalities: by record without the order, in order, and without the
// corrections. The source of a note is never compared; the correction is.
func ExampleChord_Equal() {
	c4, e4, g4 := unbound(spelling.C(), 4), unbound(spelling.E(), 4), unbound(spelling.G(), 4)

	a := chord.MustNew(c4, e4, g4)
	b := chord.MustNew(g4, c4, e4)
	fmt.Println(a.Equal(b), a.EqualInOrder(b), a.Compare(b))

	bound := chord.MustNew(note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.Zero(), settings.Standard()), e4, g4)
	fmt.Println(a.Equal(bound))

	corrected := chord.MustNew(note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.MustFromInt(5), nil), e4, g4)
	fmt.Println(a.Equal(corrected), a.SameRecords(corrected))
	// Output:
	// true false 0
	// true
	// false true
}

// A changing method changes the chord itself; its Must variant returns the
// same chord, so changes are chained. A failed operation leaves the chord
// as it was.
func ExampleChord_MustAdd() {
	c4, d4, e4, g4 := unbound(spelling.C(), 4), unbound(spelling.D(), 4), unbound(spelling.E(), 4), unbound(spelling.G(), 4)

	c := chord.MustNew(c4)
	p := c.MustAdd(e4).MustAdd(g4)
	fmt.Println(p == &c, c.Len())

	err := c.Remove(d4)
	fmt.Println(errors.Is(err, chord.ErrNotFound), c.Len())
	// Output:
	// true 3
	// true 3
}

// A copy made by assignment and the original change independently.
func Example_copy() {
	c4, e4, g4 := unbound(spelling.C(), 4), unbound(spelling.E(), 4), unbound(spelling.G(), 4)

	c := chord.MustNew(c4, e4)
	d := c
	d.MustAdd(g4)
	c.MustRemove(c4)
	fmt.Println(c)
	fmt.Println(d)
	// Output:
	// [E ASPN 4, 0 cents, unbound]
	// [C ASPN 4, 0 cents, unbound; E ASPN 4, 0 cents, unbound; G ASPN 4, 0 cents, unbound]
}

// The notes a chord gives are copies: changing them does not change the
// chord. A note in the chord is changed only by an operation of the chord.
func ExampleChord_Notes() {
	c4, e4 := unbound(spelling.C(), 4), unbound(spelling.E(), 4)

	c := chord.MustNew(c4, e4)
	notes := c.Notes()
	notes[0].MustShift(1)
	c.Notes()[1].MustShift(1) // changes only the returned copy
	fmt.Println(notes[0])
	fmt.Println(c)

	c.MustReplaceAt(1, unbound(spelling.E(), 5))
	fmt.Println(c)
	// Output:
	// C ASPN 5, 0 cents, unbound
	// [C ASPN 4, 0 cents, unbound; E ASPN 4, 0 cents, unbound]
	// [C ASPN 4, 0 cents, unbound; E ASPN 5, 0 cents, unbound]
}

// Rules of the user select and transform notes. A rule may fail: then the
// operation reports ErrRuleFailed with the error of the rule, and the chord
// does not change.
func ExampleChord_ReplaceFunc() {
	c4, g4 := unbound(spelling.C(), 4), unbound(spelling.G(), 4)
	e := note.MustFromSpelling(spelling.E(), cents.Zero(), nil) // without an octave

	c := chord.MustNew(c4, e, g4)
	hasOctave := func(n note.Note) (bool, error) { return n.HasOctave(), nil }
	fmt.Println(c.MustCountFunc(hasOctave))

	up := func(n note.Note) (note.Note, error) { // n is a copy given to the rule
		err := n.Up()
		return n, err
	}
	err := c.ReplaceFunc(up) // E has no octave: no note is changed
	fmt.Println(errors.Is(err, chord.ErrRuleFailed), errors.Is(err, note.ErrNoOctave))

	c.MustKeepFunc(hasOctave).MustReplaceFunc(up)
	fmt.Println(c)
	// Output:
	// 2
	// true true
	// [C ASPN 5, 0 cents, unbound; G ASPN 5, 0 cents, unbound]
}

// Shifting by octaves moves every note, or none: a note without an octave
// cannot be shifted.
func ExampleChord_Shift() {
	c4, e4 := unbound(spelling.C(), 4), unbound(spelling.E(), 4)

	c := chord.MustNew(c4, e4)
	c.MustShift(1)
	fmt.Println(c)

	mixed := chord.MustNew(c4, note.MustFromSpelling(spelling.E(), cents.Zero(), nil))
	err := mixed.Shift(1)
	fmt.Println(errors.Is(err, note.ErrNoOctave))
	fmt.Println(mixed)
	// Output:
	// [C ASPN 5, 0 cents, unbound; E ASPN 5, 0 cents, unbound]
	// true
	// [C ASPN 4, 0 cents, unbound; E, 0 cents, unbound]
}

// A chord reads no settings. The order by position and the frequencies come
// from the operations of package note on the notes of the chord.
func ExampleChord_SortFunc() {
	std := settings.Standard()
	bound := chord.MustNew(
		note.MustFromPitch(pitch.MustFromASPN(spelling.G(), 4), cents.Zero(), std),
		note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.Zero(), std),
	)
	bound.MustSortFunc(func(a, b note.Note) (int, error) { return a.ComparePosition(b) })
	for x := range bound.Values() {
		fmt.Printf("%.4f\n", x.MustFrequency())
	}
	// Output:
	// 261.6256
	// 391.9954
}

// A chain can start at creation through new. An element of a slice is
// changed in place; a range variable and an element of a map are copies.
func Example_newChain() {
	c4, e4, g4 := unbound(spelling.C(), 4), unbound(spelling.E(), 4), unbound(spelling.G(), 4)

	p := new(chord.MustNew(c4)).MustAdd(e4) // p is *chord.Chord

	chords := []chord.Chord{chord.MustNew(c4), chord.MustNew(e4)}
	chords[1].MustAdd(g4)
	for _, x := range chords {
		x.MustClear() // changes a copy only
	}

	m := map[string]chord.Chord{"a": chord.MustNew(c4)}
	v := m["a"]
	v.MustAdd(g4)
	m["a"] = v

	fmt.Println(p.Len(), chords[0].Len(), chords[1].Len(), m["a"].Len())
	// Output:
	// 2 1 2 2
}
