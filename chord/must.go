package chord

import (
	"github.com/go-muse/musex/note"
)

// The Must variants call their operation and panic with its error, the same
// error value: recover gives the error, and errors.Is works with it. Otherwise
// they return the results of the operation without the error. The Must
// variant of a changing method returns c itself, not a copy, so changes are
// chained with Must variants:
//
//	c.MustAdd(e4).MustRemove(c4).MustSort()
//
// A chain is not atomic: when a step fails, its Must variant panics, the
// earlier steps stay done and the later ones are not done.

// must panics with err when it is not nil.
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustNew is like New but panics on an error.
func MustNew(notes ...note.Note) Chord {
	c, err := New(notes...)
	must(err)

	return c
}

// MustAt is like At but panics on an error.
func (c Chord) MustAt(i int) note.Note {
	n, err := c.At(i)
	must(err)

	return n
}

// MustContainsFunc is like ContainsFunc but panics on an error.
func (c Chord) MustContainsFunc(cond Condition) bool {
	ok, err := c.ContainsFunc(cond)
	must(err)

	return ok
}

// MustCountFunc is like CountFunc but panics on an error.
func (c Chord) MustCountFunc(cond Condition) int {
	count, err := c.CountFunc(cond)
	must(err)

	return count
}

// MustIndicesFunc is like IndicesFunc but panics on an error.
func (c Chord) MustIndicesFunc(cond Condition) []int {
	indices, err := c.IndicesFunc(cond)
	must(err)

	return indices
}

// MustNotesFunc is like NotesFunc but panics on an error.
func (c Chord) MustNotesFunc(cond Condition) []note.Note {
	notes, err := c.NotesFunc(cond)
	must(err)

	return notes
}

// MustAdd is like Add but panics on an error; it returns c.
func (c *Chord) MustAdd(notes ...note.Note) *Chord {
	must(c.Add(notes...))

	return c
}

// MustAddMissing is like AddMissing but panics on an error; it returns c.
func (c *Chord) MustAddMissing(notes ...note.Note) *Chord {
	must(c.AddMissing(notes...))

	return c
}

// MustInsert is like Insert but panics on an error; it returns c.
func (c *Chord) MustInsert(i int, notes ...note.Note) *Chord {
	must(c.Insert(i, notes...))

	return c
}

// MustRemove is like Remove but panics on an error; it returns c.
func (c *Chord) MustRemove(n note.Note) *Chord {
	must(c.Remove(n))

	return c
}

// MustRemoveAll is like RemoveAll but panics on an error; it returns c.
func (c *Chord) MustRemoveAll(n note.Note) *Chord {
	must(c.RemoveAll(n))

	return c
}

// MustRemoveAt is like RemoveAt but panics on an error; it returns c.
func (c *Chord) MustRemoveAt(i int) *Chord {
	must(c.RemoveAt(i))

	return c
}

// MustRemoveFunc is like RemoveFunc but panics on an error; it returns c.
func (c *Chord) MustRemoveFunc(cond Condition) *Chord {
	must(c.RemoveFunc(cond))

	return c
}

// MustKeepFunc is like KeepFunc but panics on an error; it returns c.
func (c *Chord) MustKeepFunc(cond Condition) *Chord {
	must(c.KeepFunc(cond))

	return c
}

// MustReplace is like Replace but panics on an error; it returns c.
func (c *Chord) MustReplace(x, y note.Note) *Chord {
	must(c.Replace(x, y))

	return c
}

// MustReplaceAll is like ReplaceAll but panics on an error; it returns c.
func (c *Chord) MustReplaceAll(x, y note.Note) *Chord {
	must(c.ReplaceAll(x, y))

	return c
}

// MustReplaceAt is like ReplaceAt but panics on an error; it returns c.
func (c *Chord) MustReplaceAt(i int, y note.Note) *Chord {
	must(c.ReplaceAt(i, y))

	return c
}

// MustReplaceFunc is like ReplaceFunc but panics on an error; it returns c.
func (c *Chord) MustReplaceFunc(t Transform) *Chord {
	must(c.ReplaceFunc(t))

	return c
}

// MustReplaceAtFunc is like ReplaceAtFunc but panics on an error; it returns c.
func (c *Chord) MustReplaceAtFunc(i int, t Transform) *Chord {
	must(c.ReplaceAtFunc(i, t))

	return c
}

// MustClear is like Clear but panics on an error; it returns c.
func (c *Chord) MustClear() *Chord {
	must(c.Clear())

	return c
}

// MustRemoveRepeats is like RemoveRepeats but panics on an error; it returns c.
func (c *Chord) MustRemoveRepeats() *Chord {
	must(c.RemoveRepeats())

	return c
}

// MustUnion is like Union but panics on an error; it returns c.
func (c *Chord) MustUnion(other Chord) *Chord {
	must(c.Union(other))

	return c
}

// MustIntersect is like Intersect but panics on an error; it returns c.
func (c *Chord) MustIntersect(other Chord) *Chord {
	must(c.Intersect(other))

	return c
}

// MustSubtract is like Subtract but panics on an error; it returns c.
func (c *Chord) MustSubtract(other Chord) *Chord {
	must(c.Subtract(other))

	return c
}

// MustSort is like Sort but panics on an error; it returns c.
func (c *Chord) MustSort() *Chord {
	must(c.Sort())

	return c
}

// MustSortFunc is like SortFunc but panics on an error; it returns c.
func (c *Chord) MustSortFunc(order Comparison) *Chord {
	must(c.SortFunc(order))

	return c
}

// MustReverse is like Reverse but panics on an error; it returns c.
func (c *Chord) MustReverse() *Chord {
	must(c.Reverse())

	return c
}

// MustShift is like Shift but panics on an error; it returns c.
func (c *Chord) MustShift(k int) *Chord {
	must(c.Shift(k))

	return c
}

// MustShiftDown is like ShiftDown but panics on an error; it returns c.
func (c *Chord) MustShiftDown(k int) *Chord {
	must(c.ShiftDown(k))

	return c
}

// MustUp is like Up but panics on an error; it returns c.
func (c *Chord) MustUp() *Chord {
	must(c.Up())

	return c
}

// MustDown is like Down but panics on an error; it returns c.
func (c *Chord) MustDown() *Chord {
	must(c.Down())

	return c
}
