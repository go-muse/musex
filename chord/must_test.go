package chord_test

import (
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-7, Must variants: ОП-39; К-36, К-37; СЕ-8; architecture 5.11.

// mustCase pairs a changing method with its Must variant, with arguments that
// succeed and arguments that fail on the chord of mustBase.
type mustCase struct {
	name       string
	ok, bad    func(c *chord.Chord) error
	mustOK     func(c *chord.Chord) *chord.Chord
	mustBad    func(c *chord.Chord) *chord.Chord
	badErr     error
	ruleFailed bool
}

func mustBase() []note.Note { return list(e4, c4, g4, c4) }

func mustCases() []mustCase {
	isC4 := func(n note.Note) (bool, error) { return n.Equal(c4), nil }
	failCond := func(note.Note) (bool, error) { return false, errRule }
	failTr := func(n note.Note) (note.Note, error) { return n, errRule }
	failCmp := func(note.Note, note.Note) (int, error) { return 0, errRule }
	other := chord.MustNew(b4, c4)
	return []mustCase{
		{name: "Add", ok: func(c *chord.Chord) error { return c.Add(b4) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustAdd(b4) },
			bad: func(c *chord.Chord) error { return c.Add(note.Note{}) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustAdd(note.Note{}) }, badErr: chord.ErrUnset},
		{name: "AddMissing", ok: func(c *chord.Chord) error { return c.AddMissing(b4, c4) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustAddMissing(b4, c4) },
			bad: func(c *chord.Chord) error { return c.AddMissing(note.Note{}) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustAddMissing(note.Note{}) }, badErr: chord.ErrUnset},
		{name: "Insert", ok: func(c *chord.Chord) error { return c.Insert(1, b4) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustInsert(1, b4) },
			bad: func(c *chord.Chord) error { return c.Insert(5, b4) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustInsert(5, b4) }, badErr: chord.ErrIndex},
		{name: "Remove", ok: func(c *chord.Chord) error { return c.Remove(c4) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustRemove(c4) },
			bad: func(c *chord.Chord) error { return c.Remove(d4) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustRemove(d4) }, badErr: chord.ErrNotFound},
		{name: "RemoveAll", ok: func(c *chord.Chord) error { return c.RemoveAll(c4) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustRemoveAll(c4) },
			bad: func(c *chord.Chord) error { return c.RemoveAll(d4) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustRemoveAll(d4) }, badErr: chord.ErrNotFound},
		{name: "RemoveAt", ok: func(c *chord.Chord) error { return c.RemoveAt(1) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustRemoveAt(1) },
			bad: func(c *chord.Chord) error { return c.RemoveAt(4) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustRemoveAt(4) }, badErr: chord.ErrIndex},
		{name: "RemoveFunc", ok: func(c *chord.Chord) error { return c.RemoveFunc(isC4) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustRemoveFunc(isC4) },
			bad: func(c *chord.Chord) error { return c.RemoveFunc(failCond) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustRemoveFunc(failCond) }, badErr: chord.ErrRuleFailed, ruleFailed: true},
		{name: "KeepFunc", ok: func(c *chord.Chord) error { return c.KeepFunc(isC4) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustKeepFunc(isC4) },
			bad: func(c *chord.Chord) error { return c.KeepFunc(nil) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustKeepFunc(nil) }, badErr: chord.ErrUnset},
		{name: "Replace", ok: func(c *chord.Chord) error { return c.Replace(c4, d4) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustReplace(c4, d4) },
			bad: func(c *chord.Chord) error { return c.Replace(f4, d4) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustReplace(f4, d4) }, badErr: chord.ErrNotFound},
		{name: "ReplaceAll", ok: func(c *chord.Chord) error { return c.ReplaceAll(c4, d4) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustReplaceAll(c4, d4) },
			bad: func(c *chord.Chord) error { return c.ReplaceAll(f4, d4) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustReplaceAll(f4, d4) }, badErr: chord.ErrNotFound},
		{name: "ReplaceAt", ok: func(c *chord.Chord) error { return c.ReplaceAt(2, d4) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustReplaceAt(2, d4) },
			bad: func(c *chord.Chord) error { return c.ReplaceAt(2, note.Note{}) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustReplaceAt(2, note.Note{}) }, badErr: chord.ErrUnset},
		{name: "ReplaceFunc", ok: func(c *chord.Chord) error { return c.ReplaceFunc(up) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustReplaceFunc(up) },
			bad: func(c *chord.Chord) error { return c.ReplaceFunc(failTr) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustReplaceFunc(failTr) }, badErr: chord.ErrRuleFailed, ruleFailed: true},
		{name: "ReplaceAtFunc", ok: func(c *chord.Chord) error { return c.ReplaceAtFunc(3, up) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustReplaceAtFunc(3, up) },
			bad: func(c *chord.Chord) error { return c.ReplaceAtFunc(3, failTr) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustReplaceAtFunc(3, failTr) }, badErr: chord.ErrRuleFailed, ruleFailed: true},
		{name: "Clear", ok: func(c *chord.Chord) error { return c.Clear() }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustClear() }},
		{name: "RemoveRepeats", ok: func(c *chord.Chord) error { return c.RemoveRepeats() }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustRemoveRepeats() }},
		{name: "Union", ok: func(c *chord.Chord) error { return c.Union(other) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustUnion(other) },
			bad: func(c *chord.Chord) error { return c.Union(chord.Chord{}) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustUnion(chord.Chord{}) }, badErr: chord.ErrUnset},
		{name: "Intersect", ok: func(c *chord.Chord) error { return c.Intersect(other) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustIntersect(other) },
			bad: func(c *chord.Chord) error { return c.Intersect(chord.Chord{}) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustIntersect(chord.Chord{}) }, badErr: chord.ErrUnset},
		{name: "Subtract", ok: func(c *chord.Chord) error { return c.Subtract(other) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustSubtract(other) },
			bad: func(c *chord.Chord) error { return c.Subtract(chord.Chord{}) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustSubtract(chord.Chord{}) }, badErr: chord.ErrUnset},
		{name: "Sort", ok: func(c *chord.Chord) error { return c.Sort() }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustSort() }},
		{name: "SortFunc", ok: func(c *chord.Chord) error { return c.SortFunc(reverseByRecord) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustSortFunc(reverseByRecord) },
			bad: func(c *chord.Chord) error { return c.SortFunc(failCmp) }, mustBad: func(c *chord.Chord) *chord.Chord { return c.MustSortFunc(failCmp) }, badErr: chord.ErrRuleFailed, ruleFailed: true},
		{name: "Reverse", ok: func(c *chord.Chord) error { return c.Reverse() }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustReverse() }},
		{name: "Shift", ok: func(c *chord.Chord) error { return c.Shift(2) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustShift(2) }},
		{name: "ShiftDown", ok: func(c *chord.Chord) error { return c.ShiftDown(2) }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustShiftDown(2) }},
		{name: "Up", ok: func(c *chord.Chord) error { return c.Up() }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustUp() }},
		{name: "Down", ok: func(c *chord.Chord) error { return c.Down() }, mustOK: func(c *chord.Chord) *chord.Chord { return c.MustDown() }},
	}
}

func TestMustChanging(t *testing.T) {
	t.Parallel()
	for _, mc := range mustCases() {
		t.Run(mc.name, func(t *testing.T) {
			t.Parallel()
			// Success: the same result as the variant with an error; the
			// receiver itself is returned.
			want := mk(t, mustBase()...)
			require.NoError(t, mc.ok(&want))
			c := mk(t, mustBase()...)
			p := mc.mustOK(&c)
			assert.Same(t, &c, p, "the Must variant returns the receiver")
			assertNotes(t, want.Notes(), c.Notes(), "the same result as the variant with an error")

			if mc.mustBad == nil {
				return
			}
			// Failure: a panic with the same error; the chord is the same.
			d := mk(t, mustBase()...)
			before := stateOf(d)
			err := mc.bad(&d)
			require.Error(t, err)
			v, panicked := recovered(func() { mc.mustBad(&d) })
			require.True(t, panicked, "the Must variant panics")
			perr, isErr := v.(error)
			require.True(t, isErr, "the value of the panic is an error: %#v", v)
			if mc.ruleFailed {
				assertRuleFailed(t, perr, errRule)
			} else {
				assertOnly(t, perr, mc.badErr)
			}
			assertState(t, before, d)
		})
	}
}

func TestMustShiftFailures(t *testing.T) {
	t.Parallel()
	shifts := map[string]func(c *chord.Chord) *chord.Chord{
		"MustShift":     func(c *chord.Chord) *chord.Chord { return c.MustShift(1) },
		"MustShiftDown": func(c *chord.Chord) *chord.Chord { return c.MustShiftDown(1) },
		"MustUp":        (*chord.Chord).MustUp,
		"MustDown":      (*chord.Chord).MustDown,
	}
	for name, f := range shifts {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, c4, eNo)
			before := stateOf(c)
			v, panicked := recovered(func() { f(&c) })
			require.True(t, panicked)
			err, isErr := v.(error)
			require.True(t, isErr)
			assert.ErrorIs(t, err, note.ErrNoOctave)
			assertNoChordSentinel(t, err)
			assertState(t, before, c)
		})
	}
}

func TestMustUnsetAndNil(t *testing.T) {
	t.Parallel()
	musts := map[string]func(c *chord.Chord) *chord.Chord{
		"MustAdd":           func(c *chord.Chord) *chord.Chord { return c.MustAdd(c4) },
		"MustAddMissing":    func(c *chord.Chord) *chord.Chord { return c.MustAddMissing(c4) },
		"MustInsert":        func(c *chord.Chord) *chord.Chord { return c.MustInsert(0, c4) },
		"MustRemove":        func(c *chord.Chord) *chord.Chord { return c.MustRemove(c4) },
		"MustRemoveAll":     func(c *chord.Chord) *chord.Chord { return c.MustRemoveAll(c4) },
		"MustRemoveAt":      func(c *chord.Chord) *chord.Chord { return c.MustRemoveAt(0) },
		"MustRemoveFunc":    func(c *chord.Chord) *chord.Chord { return c.MustRemoveFunc(hasOctave) },
		"MustKeepFunc":      func(c *chord.Chord) *chord.Chord { return c.MustKeepFunc(hasOctave) },
		"MustReplace":       func(c *chord.Chord) *chord.Chord { return c.MustReplace(c4, d4) },
		"MustReplaceAll":    func(c *chord.Chord) *chord.Chord { return c.MustReplaceAll(c4, d4) },
		"MustReplaceAt":     func(c *chord.Chord) *chord.Chord { return c.MustReplaceAt(0, d4) },
		"MustReplaceFunc":   func(c *chord.Chord) *chord.Chord { return c.MustReplaceFunc(up) },
		"MustReplaceAtFunc": func(c *chord.Chord) *chord.Chord { return c.MustReplaceAtFunc(0, up) },
		"MustClear":         (*chord.Chord).MustClear,
		"MustRemoveRepeats": (*chord.Chord).MustRemoveRepeats,
		"MustUnion":         func(c *chord.Chord) *chord.Chord { return c.MustUnion(chord.MustNew(c4)) },
		"MustIntersect":     func(c *chord.Chord) *chord.Chord { return c.MustIntersect(chord.MustNew(c4)) },
		"MustSubtract":      func(c *chord.Chord) *chord.Chord { return c.MustSubtract(chord.MustNew(c4)) },
		"MustSort":          (*chord.Chord).MustSort,
		"MustSortFunc":      func(c *chord.Chord) *chord.Chord { return c.MustSortFunc(byRecord) },
		"MustReverse":       (*chord.Chord).MustReverse,
		"MustShift":         func(c *chord.Chord) *chord.Chord { return c.MustShift(1) },
		"MustShiftDown":     func(c *chord.Chord) *chord.Chord { return c.MustShiftDown(1) },
		"MustUp":            (*chord.Chord).MustUp,
		"MustDown":          (*chord.Chord).MustDown,
	}
	require.Len(t, musts, 25)
	for name, f := range musts {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var c chord.Chord
			assertPanicsWith(t, chord.ErrUnset, func() { f(&c) }, "the unset chord")
			assertUnsetChord(t, c)

			var p *chord.Chord
			assertPanicsWith(t, chord.ErrUnset, func() { f(p) }, "a nil *Chord")
		})
	}
}

func TestMustReading(t *testing.T) {
	t.Parallel()
	c := mk(t, c4, eNo, g4)

	assertChord(t, list(c4, e4), chord.MustNew(c4, e4))
	assertPanicsWith(t, chord.ErrUnset, func() { chord.MustNew(c4, note.Note{}) })

	assert.True(t, sameNote(eNo, c.MustAt(1)))
	assertPanicsWith(t, chord.ErrIndex, func() { c.MustAt(3) })
	assertPanicsWith(t, chord.ErrUnset, func() { chord.Chord{}.MustAt(0) })

	assert.True(t, c.MustContainsFunc(hasOctave))
	assert.Equal(t, 2, c.MustCountFunc(hasOctave))
	assert.Equal(t, []int{0, 2}, c.MustIndicesFunc(hasOctave))
	assertNotes(t, list(c4, g4), c.MustNotesFunc(hasOctave))

	failing := func(note.Note) (bool, error) { return false, errRule }
	for name, f := range map[string]func(){
		"MustContainsFunc": func() { c.MustContainsFunc(failing) },
		"MustCountFunc":    func() { c.MustCountFunc(failing) },
		"MustIndicesFunc":  func() { c.MustIndicesFunc(failing) },
		"MustNotesFunc":    func() { c.MustNotesFunc(failing) },
	} {
		v, panicked := recovered(f)
		require.True(t, panicked, name)
		err, isErr := v.(error)
		require.True(t, isErr, name)
		assertRuleFailed(t, err, errRule, name)
	}
}

func TestMustK36Chain(t *testing.T) {
	t.Parallel()
	c := mk(t, c4)
	p1 := c.MustAdd(e4)
	p2 := p1.MustAdd(g4)
	p3 := p2.MustRemove(c4)
	assert.Same(t, &c, p1, "each step returns the same chord c")
	assert.Same(t, &c, p2)
	assert.Same(t, &c, p3)
	assertChord(t, list(e4, g4), c)

	d := mk(t, c4)
	d.MustAdd(e4).MustAdd(g4).MustRemove(c4)
	assertChord(t, list(e4, g4), d)
	assert.Equal(t, 2, d.MustAdd(c4).MustRemove(c4).Len(), "reading methods on *Chord")
}

func TestMustK37ChainFails(t *testing.T) {
	t.Parallel()
	c := mk(t, c4)
	v, panicked := recovered(func() {
		c.MustAdd(e4).MustRemove(d4).MustAdd(g4)
	})
	require.True(t, panicked)
	err, isErr := v.(error)
	require.True(t, isErr)
	assertOnly(t, err, chord.ErrNotFound, "E-2 at the second step")
	assertChord(t, list(c4, e4), c, "the first step is done, the third is not")
}

func TestMustNewChain(t *testing.T) {
	t.Parallel()
	// Architecture 5.11: a chain from creation through new(...).
	p := new(chord.MustNew(c4)).MustAdd(e4)
	assertChord(t, list(c4, e4), *p)
}
