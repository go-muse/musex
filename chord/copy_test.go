package chord_test

import (
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-7, copies and independence: П-7, П-8; К-34, К-35, К-38, К-39; КП-2, КП-3,
// КП-5; architecture 5.2, 5.6 (sequences), 6; plan, "Independence of copies".

// copyBase is the chord of the copy tests: repeats, a source, an order that
// Sort, Reverse and RemoveRepeats all change.
func copyBase() []note.Note {
	return list(e4, c4, pn(g4.Spelling(), 4, 0, std()), c4)
}

// successOps returns every changing method on the success path for the chord
// of copyBase, keyed by its name. Each changes the chord.
func successOps() map[string]func(c *chord.Chord) error {
	isC4 := func(n note.Note) (bool, error) { return n.Equal(c4), nil }
	return map[string]func(c *chord.Chord) error{
		"Add":           func(c *chord.Chord) error { return c.Add(b4) },
		"AddMissing":    func(c *chord.Chord) error { return c.AddMissing(b4) },
		"Insert":        func(c *chord.Chord) error { return c.Insert(1, b4) },
		"Remove":        func(c *chord.Chord) error { return c.Remove(c4) },
		"RemoveAll":     func(c *chord.Chord) error { return c.RemoveAll(c4) },
		"RemoveAt":      func(c *chord.Chord) error { return c.RemoveAt(0) },
		"RemoveFunc":    func(c *chord.Chord) error { return c.RemoveFunc(isC4) },
		"KeepFunc":      func(c *chord.Chord) error { return c.KeepFunc(isC4) },
		"Replace":       func(c *chord.Chord) error { return c.Replace(c4, d4) },
		"ReplaceAll":    func(c *chord.Chord) error { return c.ReplaceAll(c4, d4) },
		"ReplaceAt":     func(c *chord.Chord) error { return c.ReplaceAt(0, d4) },
		"ReplaceFunc":   func(c *chord.Chord) error { return c.ReplaceFunc(up) },
		"ReplaceAtFunc": func(c *chord.Chord) error { return c.ReplaceAtFunc(0, up) },
		"Clear":         func(c *chord.Chord) error { return c.Clear() },
		"RemoveRepeats": func(c *chord.Chord) error { return c.RemoveRepeats() },
		"Union":         func(c *chord.Chord) error { return c.Union(chord.MustNew(b4, c4, c4, c4)) },
		"Intersect":     func(c *chord.Chord) error { return c.Intersect(chord.MustNew(c4)) },
		"Subtract":      func(c *chord.Chord) error { return c.Subtract(chord.MustNew(c4, e4)) },
		"Sort":          func(c *chord.Chord) error { return c.Sort() },
		"SortFunc":      func(c *chord.Chord) error { return c.SortFunc(reverseByRecord) },
		"Reverse":       func(c *chord.Chord) error { return c.Reverse() },
		"Shift":         func(c *chord.Chord) error { return c.Shift(1) },
		"ShiftDown":     func(c *chord.Chord) error { return c.ShiftDown(1) },
		"Up":            func(c *chord.Chord) error { return c.Up() },
		"Down":          func(c *chord.Chord) error { return c.Down() },
	}
}

func TestCopyIndependentBothWays(t *testing.T) {
	t.Parallel()
	// КП-3, G-8: for each changing method, a change of the copy does not
	// change the original, and a later change of the original does not
	// change the copy.
	for name, op := range successOps() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, copyBase()...)
			before := stateOf(c)

			d := c
			require.NoError(t, op(&d))
			require.False(t, d.EqualInOrder(mk(t, copyBase()...)), "precondition: %s changes the chord", name)
			assertState(t, before, c, "the original after a change of the copy")

			afterD := stateOf(d)
			require.NoError(t, op(&c))
			assertState(t, afterD, d, "the copy after a change of the original")

			// The same change gives the same result on both.
			assertNotes(t, afterD.notes, c.Notes(), "the original after the same change")
		})
	}
}

func TestCopyK34(t *testing.T) {
	t.Parallel()
	c := mk(t, c4, e4)
	d := c
	require.NoError(t, d.Add(g4))
	assertChord(t, list(c4, e4, g4), d)
	assertChord(t, list(c4, e4), c)
	require.NoError(t, c.Remove(c4))
	assertChord(t, list(e4), c)
	assertChord(t, list(c4, e4, g4), d)
}

func TestCopyK35(t *testing.T) {
	t.Parallel()
	// K-35: different notes added to two copies after copying, for chords
	// that were built at once and note by note (a spare capacity must not be
	// shared).
	builds := map[string]func(t *testing.T) chord.Chord{
		"created": func(t *testing.T) chord.Chord { return mk(t, c4, e4) },
		"added one by one": func(t *testing.T) chord.Chord {
			c := chord.MustNew()
			require.NoError(t, c.Add(c4))
			require.NoError(t, c.Add(e4))
			return c
		},
		"after removing": func(t *testing.T) chord.Chord {
			c := mk(t, c4, e4, g4, b4)
			require.NoError(t, c.RemoveAt(3))
			require.NoError(t, c.RemoveAt(2))
			return c
		},
	}
	for name, build := range builds {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			c := build(t)
			d := c
			require.NoError(t, c.Add(g4))
			require.NoError(t, d.Add(b4))
			assertChord(t, list(c4, e4, g4), c)
			assertChord(t, list(c4, e4, b4), d)

			x := build(t)
			y := x
			require.NoError(t, x.Insert(2, g4))
			require.NoError(t, y.Insert(2, b4))
			assertChord(t, list(c4, e4, g4), x)
			assertChord(t, list(c4, e4, b4), y)
		})
	}
}

func TestCopyRemoveLastThenAdd(t *testing.T) {
	t.Parallel()
	// Plan: removing the last note of c, then adding: d is the same.
	removals := map[string]func(c *chord.Chord) error{
		"RemoveAt":  func(c *chord.Chord) error { return c.RemoveAt(2) },
		"Remove":    func(c *chord.Chord) error { return c.Remove(g4) },
		"RemoveAll": func(c *chord.Chord) error { return c.RemoveAll(g4) },
		"RemoveFunc": func(c *chord.Chord) error {
			return c.RemoveFunc(func(n note.Note) (bool, error) { return n.Equal(g4), nil })
		},
		"KeepFunc": func(c *chord.Chord) error {
			return c.KeepFunc(func(n note.Note) (bool, error) { return !n.Equal(g4), nil })
		},
		"Subtract":  func(c *chord.Chord) error { return c.Subtract(chord.MustNew(g4)) },
		"Intersect": func(c *chord.Chord) error { return c.Intersect(chord.MustNew(c4, e4)) },
	}
	for name, remove := range removals {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, c4, e4, g4)
			d := c
			require.NoError(t, remove(&c))
			require.NoError(t, c.Add(b4))
			assertChord(t, list(c4, e4, b4), c)
			assertChord(t, list(c4, e4, g4), d, "the copy is the same")

			// The other way round.
			x := mk(t, c4, e4, g4)
			y := x
			require.NoError(t, remove(&y))
			require.NoError(t, y.Add(b4))
			assertChord(t, list(c4, e4, g4), x, "the original is the same")
		})
	}
}

func TestCopyAtomicFailure(t *testing.T) {
	t.Parallel()
	// Plan: a copy taken before a failed operation equals the chord after the
	// failure (atomicity, П-7).
	failing := map[string]func(c *chord.Chord) error{
		"Add":           func(c *chord.Chord) error { return c.Add(b4, note.Note{}) },
		"AddMissing":    func(c *chord.Chord) error { return c.AddMissing(b4, note.Note{}) },
		"Insert":        func(c *chord.Chord) error { return c.Insert(9, b4) },
		"Remove":        func(c *chord.Chord) error { return c.Remove(f4) },
		"RemoveAll":     func(c *chord.Chord) error { return c.RemoveAll(f4) },
		"RemoveAt":      func(c *chord.Chord) error { return c.RemoveAt(9) },
		"RemoveFunc":    func(c *chord.Chord) error { var r recorder; return c.RemoveFunc(r.condFailAt(3, yes)) },
		"KeepFunc":      func(c *chord.Chord) error { var r recorder; return c.KeepFunc(r.condFailAt(3, no)) },
		"Replace":       func(c *chord.Chord) error { return c.Replace(f4, d4) },
		"ReplaceAll":    func(c *chord.Chord) error { return c.ReplaceAll(c4, note.Note{}) },
		"ReplaceAt":     func(c *chord.Chord) error { return c.ReplaceAt(9, d4) },
		"ReplaceFunc":   func(c *chord.Chord) error { var r recorder; return c.ReplaceFunc(r.transformFailAt(3, up)) },
		"ReplaceAtFunc": func(c *chord.Chord) error { var r recorder; return c.ReplaceAtFunc(0, r.transformFailAt(1, up)) },
		"Union":         func(c *chord.Chord) error { return c.Union(chord.Chord{}) },
		"Intersect":     func(c *chord.Chord) error { return c.Intersect(chord.Chord{}) },
		"Subtract":      func(c *chord.Chord) error { return c.Subtract(chord.Chord{}) },
		"SortFunc": func(c *chord.Chord) error {
			return c.SortFunc(func(note.Note, note.Note) (int, error) { return 0, errRule })
		},
		"Shift":     func(c *chord.Chord) error { return c.Shift(1) },
		"ShiftDown": func(c *chord.Chord) error { return c.ShiftDown(-1) },
		"Up":        func(c *chord.Chord) error { return c.Up() },
		"Down":      func(c *chord.Chord) error { return c.Down() },
	}
	for name, op := range failing {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// The last note has no octave: the octave shifts fail at it, after
			// the notes before it.
			c := mk(t, e4, c4, pn(g4.Spelling(), 4, 0, std()), c4, eNo)
			d := c
			require.Error(t, op(&c))
			assert.True(t, c.EqualInOrder(d), "the chord equals in order the copy taken before")
			assertNotes(t, d.Notes(), c.Notes(), "the same notes with their sources")
		})
	}
}

func TestCopyInputSlices(t *testing.T) {
	t.Parallel()
	// K-39, G-9: a list given to an operation is copied, not kept, and not
	// changed.
	ops := map[string]func(c *chord.Chord, in []note.Note) error{
		"New": func(c *chord.Chord, in []note.Note) error {
			x, err := chord.New(in...)
			*c = x
			return err
		},
		"Add":        func(c *chord.Chord, in []note.Note) error { return c.Add(in...) },
		"AddMissing": func(c *chord.Chord, in []note.Note) error { return c.AddMissing(in...) },
		"Insert":     func(c *chord.Chord, in []note.Note) error { return c.Insert(0, in...) },
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			in := list(c4, pn(e4.Spelling(), 4, 0, std()))
			inBefore := stateNotes(in)
			c := chord.MustNew()
			require.NoError(t, op(&c, in))
			assertNotes(t, inBefore, in, "the list is not changed by the operation")
			want := stateNotes(in)

			in[0] = g4
			require.NoError(t, in[1].Up())
			assertChord(t, want, c, "a change of the list after the operation does not change the chord")
		})
	}
}

func TestCopyInputSlicesOfQueries(t *testing.T) {
	t.Parallel()
	c := mk(t, c4, e4, g4)
	in := list(g4, c4)
	inBefore := stateNotes(in)
	assert.True(t, c.ContainsAll(in...))
	assert.True(t, c.ContainsAny(in...))
	assertNotes(t, inBefore, in, "the list is not changed")
}

func TestCopyK38Outputs(t *testing.T) {
	t.Parallel()
	// K-38, КП-5, П-8: what reading gives is new; changing it does not
	// change the chord.
	st := std()
	base := list(c4, pn(e4.Spelling(), 4, 0, st), c4)
	outputs := map[string]func(c chord.Chord) []note.Note{
		"Notes":         chord.Chord.Notes,
		"SortedNotes":   chord.Chord.SortedNotes,
		"DistinctNotes": chord.Chord.DistinctNotes,
		"NotesFunc":     func(c chord.Chord) []note.Note { return c.MustNotesFunc(hasOctave) },
		"All": func(c chord.Chord) []note.Note {
			var out []note.Note
			for _, n := range c.All() {
				out = append(out, n)
			}
			return out
		},
		"Values": func(c chord.Chord) []note.Note {
			var out []note.Note
			for n := range c.Values() {
				out = append(out, n)
			}
			return out
		},
	}
	for name, get := range outputs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, base...)
			before := stateOf(c)
			got := get(c)
			require.NotEmpty(t, got)
			require.NoError(t, got[0].Up()) // changes the note in the slice
			got[len(got)-1] = g4            // replaces an element of the slice
			assertState(t, before, c)
			again := get(c)
			assert.False(t, sameNote(got[0], again[0]), "the next call gives a new slice")
		})
	}

	t.Run("K-38 At", func(t *testing.T) {
		t.Parallel()
		c := mk(t, c4, e4)
		before := stateOf(c)
		n, err := c.At(0)
		require.NoError(t, err)
		require.NoError(t, n.Up())
		m := c.MustAt(1)
		m.MustUp()
		assertState(t, before, c)
	})

	t.Run("K-38 Notes", func(t *testing.T) {
		t.Parallel()
		c := mk(t, c4, e4)
		before := stateOf(c)
		notes := c.Notes()
		notes[0] = g4
		c.Notes()[0].MustShift(1) // changes only the returned copy (architecture 5.11)
		assertState(t, before, c)
	})

	t.Run("Find", func(t *testing.T) {
		t.Parallel()
		c := mk(t, c4, e4)
		before := stateOf(c)
		_, n, ok := c.Find(e4)
		require.True(t, ok)
		n.MustUp()
		assertState(t, before, c)
	})

	t.Run("Indices", func(t *testing.T) {
		t.Parallel()
		c := mk(t, c4, e4, c4)
		idx := c.Indices(c4)
		idx[0] = 1
		assert.Equal(t, []int{0, 2}, c.Indices(c4))
		fidx := c.MustIndicesFunc(hasOctave)
		fidx[0] = 9
		assert.Equal(t, []int{0, 1, 2}, c.MustIndicesFunc(hasOctave))
	})
}

func TestCopySequenceSnapshot(t *testing.T) {
	t.Parallel()
	// Architecture 5.6: All and Values give the notes at the moment of their
	// call; a change of the chord, also inside the loop, does not change them.
	t.Run("All, changed inside the loop", func(t *testing.T) {
		t.Parallel()
		c := mk(t, c4, e4)
		var got []note.Note
		for _, n := range c.All() {
			got = append(got, n)
			require.NoError(t, c.Add(g4))
			require.NoError(t, c.ReplaceAt(1, b4))
		}
		assertNotes(t, list(c4, e4), got)
		assertChord(t, list(c4, b4, g4, g4), c)
	})
	t.Run("Values, changed inside the loop", func(t *testing.T) {
		t.Parallel()
		c := mk(t, c4, e4)
		var got []note.Note
		for n := range c.Values() {
			got = append(got, n)
			require.NoError(t, c.Clear())
		}
		assertNotes(t, list(c4, e4), got)
		assertChord(t, list(), c)
	})
	t.Run("repeated walk of one sequence", func(t *testing.T) {
		t.Parallel()
		c := mk(t, c4, e4)
		all := c.All()
		vals := c.Values()
		require.NoError(t, c.Sort())
		require.NoError(t, c.Reverse())
		require.NoError(t, c.Add(g4))
		for range 2 {
			var got []note.Note
			for _, n := range all {
				got = append(got, n)
			}
			assertNotes(t, list(c4, e4), got, "All")
			got = nil
			for n := range vals {
				got = append(got, n)
			}
			assertNotes(t, list(c4, e4), got, "Values")
		}
	})
	t.Run("changing a yielded note", func(t *testing.T) {
		t.Parallel()
		c := mk(t, c4, e4)
		before := stateOf(c)
		for _, n := range c.All() {
			n.MustUp()
		}
		for n := range c.Values() {
			n.MustDown()
		}
		assertState(t, before, c)
	})
}

func TestCopyUnsetAndEmpty(t *testing.T) {
	t.Parallel()
	var u chord.Chord
	v := u
	assertOnly(t, v.Add(c4), chord.ErrUnset)
	assertUnsetChord(t, v)
	assertUnsetChord(t, u)

	e := chord.MustNew()
	f := e
	require.NoError(t, f.Add(c4))
	assertChord(t, list(), e)
	assertChord(t, list(c4), f)
}

func TestCopyArgumentChord(t *testing.T) {
	t.Parallel()
	// П-7: an argument chord is not changed, also when it is a copy of the
	// receiver.
	c := mk(t, c4, e4, c4)
	d := c
	require.NoError(t, c.Subtract(d))
	assertChord(t, list(), c)
	assertChord(t, list(c4, e4, c4), d)

	x := mk(t, c4, e4)
	y := mk(t, e4, g4, g4)
	require.NoError(t, x.Union(y))
	require.NoError(t, y.Add(b4)) // a later change of the argument does not change x
	assertChord(t, list(c4, e4, g4, g4), x)
}

func TestCopyInCollections(t *testing.T) {
	t.Parallel()
	// Architecture 5.15: an element of a slice is changed in place; a range
	// variable is a copy; an element of a map is changed through a variable.
	chords := []chord.Chord{mk(t, c4), mk(t, e4)}
	chords[0].MustAdd(g4)
	for _, x := range chords {
		x.MustAdd(b4) // changes a copy only
	}
	assertChord(t, list(c4, g4), chords[0])
	assertChord(t, list(e4), chords[1])

	m := map[string]chord.Chord{"k": mk(t, c4)}
	v := m["k"]
	v.MustAdd(e4)
	assertChord(t, list(c4), m["k"], "the map keeps its copy until the assignment")
	m["k"] = v
	assertChord(t, list(c4, e4), m["k"])
	assert.Equal(t, 2, m["k"].Len(), "reading methods work on an element of a map")

	type holder struct{ c chord.Chord }
	h := holder{c: mk(t, c4)}
	h.c.MustAdd(e4)
	assertChord(t, list(c4, e4), h.c)
}

func TestCopyMustReturnsReceiver(t *testing.T) {
	t.Parallel()
	c := mk(t, c4)
	p := c.MustAdd(e4)
	assert.Same(t, &c, p, "p points to c")
	p.MustAdd(g4)
	assertChord(t, list(c4, e4, g4), c, "a change through p changes c")
}
