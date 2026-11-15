package chord_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-6. Order, transforms, octave shifts and rules of the user: ОП-35…ОП-37;
// П-9, П-10а; К-26…К-31, К-44; architecture 5.8, 5.10.

func TestOrderExamples(t *testing.T) {
	t.Parallel()
	st := std()
	s19 := set19()
	c4st := pn(c4.Spelling(), 4, 0, st)
	c419 := pn(c4.Spelling(), 4, 0, s19)
	failing := func(note.Note, note.Note) (int, error) { return 0, errRule }
	allEqual := func(note.Note, note.Note) (int, error) { return 0, nil }
	// byLetter compares the names only (the agreed order of notes without an
	// octave): C5, C4 and C are equal by it.
	byLetter := func(a, b note.Note) (int, error) {
		return nn(a.Spelling(), nil).Compare(nn(b.Spelling(), nil)), nil
	}
	base := list(e4, c4, g4) // K-26
	runChanges(t, []change{
		{name: "K-26 reverse", base: base, op: func(c *chord.Chord) error { return c.Reverse() }, want: list(g4, c4, e4)},
		{name: "K-26 sort", base: base, op: func(c *chord.Chord) error { return c.Sort() }, want: list(c4, e4, g4)},
		{name: "K-26 sort by the reverse agreed order", base: base, op: func(c *chord.Chord) error { return c.SortFunc(reverseByRecord) }, want: list(g4, e4, c4)},
		{name: "K-26 sort by a failing rule", base: base, op: func(c *chord.Chord) error { return c.SortFunc(failing) }, err: chord.ErrRuleFailed},
		{name: "K-27 sort keeps the order of equal notes", base: list(c4st, e4, c419), op: func(c *chord.Chord) error { return c.Sort() }, want: list(c4st, c419, e4)},
		{name: "K-27 sort by a rule that finds all notes equal", base: list(c4st, e4, c419), op: func(c *chord.Chord) error { return c.SortFunc(allEqual) }, want: list(c4st, e4, c419)},
		{name: "sort by a rule is stable", base: list(c5, e4, c4, cNo), op: func(c *chord.Chord) error { return c.SortFunc(byLetter) }, want: list(c5, c4, cNo, e4)},
		{name: "K-15 sort: no octave first, then the record", base: list(bs4, c4p10, csNo, c4), op: func(c *chord.Chord) error { return c.Sort() }, want: list(csNo, c4, c4p10, bs4)},
		{name: "sort with repeats", base: list(g4, c4, g4, c4), op: func(c *chord.Chord) error { return c.Sort() }, want: list(c4, c4, g4, g4)},
		{name: "reverse with repeats and sources", base: list(c4st, e4, c419, g4), op: func(c *chord.Chord) error { return c.Reverse() }, want: list(g4, c419, e4, c4st)},
		{name: "reverse of one note", base: list(c4), op: func(c *chord.Chord) error { return c.Reverse() }, want: list(c4)},
		{name: "reverse of the empty chord", base: list(), op: func(c *chord.Chord) error { return c.Reverse() }, want: list()},
		{name: "sort of the empty chord", base: list(), op: func(c *chord.Chord) error { return c.Sort() }, want: list()},
		{name: "sort by a rule of the empty chord", base: list(), op: func(c *chord.Chord) error { return c.SortFunc(failing) }, want: list()},
		{name: "sort by a nil rule", base: base, op: func(c *chord.Chord) error { return c.SortFunc(nil) }, err: chord.ErrUnset},
		{name: "sort by a nil rule, the empty chord", base: list(), op: func(c *chord.Chord) error { return c.SortFunc(nil) }, err: chord.ErrUnset},
	})
}

func TestOrderByPosition(t *testing.T) {
	t.Parallel()
	// Architecture 5.15, 5.16: the order by position through package note.
	st := std()
	g4st := pn(g4.Spelling(), 4, 0, st)
	c4st := pn(c4.Spelling(), 4, 0, st)
	c := mk(t, g4st, c4st)
	require.NoError(t, c.SortFunc(func(a, b note.Note) (int, error) { return a.ComparePosition(b) }))
	assertChord(t, list(c4st, g4st), c)

	// A comparison by position fails for notes without an octave (note Е-4).
	d := mk(t, g4st, nn(c4.Spelling(), st))
	before := stateOf(d)
	err := d.SortFunc(func(a, b note.Note) (int, error) { return a.ComparePosition(b) })
	assertOnly(t, err, chord.ErrRuleFailed)
	assert.ErrorIs(t, err, note.ErrNoOctave, "the cause of the rule is kept (П-10а)")
	assertState(t, before, d)
}

func TestOrderK44Inconsistent(t *testing.T) {
	t.Parallel()
	// K-44: a rule that answers "the first comes before" for any two notes
	// keeps the composition; the order is not defined.
	before := list(c4, e4, g4, c4, pn(e4.Spelling(), 4, 0, std()))
	c := mk(t, before...)
	require.NoError(t, c.SortFunc(func(note.Note, note.Note) (int, error) { return -1, nil }))
	assert.True(t, c.Equal(mk(t, before...)), "the chord is equal by record to the chord before: %v", c)
	assert.Equal(t, len(before), c.Len())
	for _, n := range before { // the notes with their sources stay: each note of before is in c as often
		k := 0
		for _, x := range c.Notes() {
			if sameNote(n, x) {
				k++
			}
		}
		want := 0
		for _, x := range before {
			if sameNote(n, x) {
				want++
			}
		}
		assert.Equal(t, want, k, "note %v with its source", n)
	}
}

func TestTransformReplaceFunc(t *testing.T) {
	t.Parallel()
	st := std()
	s19 := set19()
	c4st := pn(c4.Spelling(), 4, 0, st)
	e419 := pn(e4.Spelling(), 4, 0, s19)
	toUnset := func(note.Note) (note.Note, error) { return note.Note{}, nil }
	toE419 := func(note.Note) (note.Note, error) { return e419, nil }
	base := list(c4p30, e4) // K-28, K-29
	runChanges(t, []change{
		{name: "K-28 up for all notes", base: base, op: func(c *chord.Chord) error { return c.ReplaceFunc(up) }, want: list(c5p30, e5)},
		{name: "K-28 up for the note at 1", base: base, op: func(c *chord.Chord) error { return c.ReplaceAtFunc(1, up) }, want: list(c4p30, e5)},
		{name: "K-28 up for the note at 2", base: base, op: func(c *chord.Chord) error { return c.ReplaceAtFunc(2, up) }, err: chord.ErrIndex},
		{name: "up for the note at -1", base: base, op: func(c *chord.Chord) error { return c.ReplaceAtFunc(-1, up) }, err: chord.ErrIndex},
		{name: "K-29 a transform that returns an unset note", base: base, op: func(c *chord.Chord) error { return c.ReplaceFunc(toUnset) }, err: chord.ErrUnset},
		{name: "K-29 a nil transform", base: base, op: func(c *chord.Chord) error { return c.ReplaceFunc(nil) }, err: chord.ErrUnset},
		{name: "a transform at 0 returns an unset note", base: base, op: func(c *chord.Chord) error { return c.ReplaceAtFunc(0, toUnset) }, err: chord.ErrUnset},
		{name: "a nil transform at 0", base: base, op: func(c *chord.Chord) error { return c.ReplaceAtFunc(0, nil) }, err: chord.ErrUnset},
		{name: "the source of the returned note", base: list(c4st, e4), op: func(c *chord.Chord) error { return c.ReplaceAtFunc(1, toE419) }, want: list(c4st, e419)},
		{name: "the transform keeps the source of the copy", base: list(c4st, e4), op: func(c *chord.Chord) error { return c.ReplaceFunc(up) }, want: list(pn(c4.Spelling(), 5, 0, st), e5)},
		{name: "the number of notes is the same", base: list(c4, e4, c4), op: func(c *chord.Chord) error { return c.ReplaceFunc(toE419) }, want: list(e419, e419, e419)},
		{name: "the empty chord", base: list(), op: func(c *chord.Chord) error { return c.ReplaceFunc(toUnset) }, want: list()},
	})
}

func TestTransformK29Failing(t *testing.T) {
	t.Parallel()
	c := mk(t, c4p30, e4)
	before := stateOf(c)
	var r recorder
	err := c.ReplaceFunc(r.transformFailAt(2, up))
	assertRuleFailed(t, err, errRule)
	assertState(t, before, c)
	assert.Equal(t, 2, r.calls(), "K-29: the transform is called twice")
}

func TestTransformCauseOfNote(t *testing.T) {
	t.Parallel()
	// A rule that fails with an error of note: E-4 keeps the cause (П-10а).
	c := mk(t, c4, eNo)
	err := c.ReplaceFunc(up)
	assertOnly(t, err, chord.ErrRuleFailed)
	assert.ErrorIs(t, err, note.ErrNoOctave)
}

func TestTransformShift(t *testing.T) {
	t.Parallel()
	st := std()
	c4p30st := pn(c4.Spelling(), 4, 30, st)
	base := list(c4p30st, e4) // K-30
	runChanges(t, []change{
		{name: "K-30 shift 2 up", base: base, op: func(c *chord.Chord) error { return c.Shift(2) }, want: list(pn(c4.Spelling(), 6, 30, st), e6)},
		{name: "K-30 shift 1 down", base: base, op: func(c *chord.Chord) error { return c.Shift(-1) }, want: list(pn(c4.Spelling(), 3, 30, st), e3)},
		{name: "K-30 shift 0", base: base, op: func(c *chord.Chord) error { return c.Shift(0) }, want: base},
		{name: "shift down 1", base: base, op: func(c *chord.Chord) error { return c.ShiftDown(1) }, want: list(pn(c4.Spelling(), 3, 30, st), e3)},
		{name: "shift down -1", base: base, op: func(c *chord.Chord) error { return c.ShiftDown(-1) }, want: list(pn(c4.Spelling(), 5, 30, st), e5)},
		{name: "shift down 0", base: base, op: func(c *chord.Chord) error { return c.ShiftDown(0) }, want: base},
		{name: "up", base: base, op: func(c *chord.Chord) error { return c.Up() }, want: list(pn(c4.Spelling(), 5, 30, st), e5)},
		{name: "down", base: base, op: func(c *chord.Chord) error { return c.Down() }, want: list(pn(c4.Spelling(), 3, 30, st), e3)},
		{name: "the order and the repeats stay", base: list(g4, c4, g4), op: func(c *chord.Chord) error { return c.Up() }, want: list(g5, c5, g5)},
		{name: "K-31 shift of the empty chord", base: list(), op: func(c *chord.Chord) error { return c.Shift(1) }, want: list()},
		{name: "shift down of the empty chord", base: list(), op: func(c *chord.Chord) error { return c.ShiftDown(1) }, want: list()},
		{name: "up of the empty chord", base: list(), op: func(c *chord.Chord) error { return c.Up() }, want: list()},
		{name: "down of the empty chord", base: list(), op: func(c *chord.Chord) error { return c.Down() }, want: list()},
		{name: "shift down by math.MinInt is exact", base: list(cMin), op: func(c *chord.Chord) error { return c.ShiftDown(math.MinInt) }, want: list(on(c4.Spelling(), 0))},
		{name: "shift by math.MaxInt", base: list(cMin), op: func(c *chord.Chord) error { return c.Shift(math.MaxInt) }, want: list(on(c4.Spelling(), -1))},
	})
}

func TestTransformShiftFailures(t *testing.T) {
	t.Parallel()
	ops := map[string]func(c *chord.Chord) error{
		"Shift(1)":     func(c *chord.Chord) error { return c.Shift(1) },
		"Shift(0)":     func(c *chord.Chord) error { return c.Shift(0) },
		"Shift(-1)":    func(c *chord.Chord) error { return c.Shift(-1) },
		"ShiftDown(1)": func(c *chord.Chord) error { return c.ShiftDown(1) },
		"ShiftDown(0)": func(c *chord.Chord) error { return c.ShiftDown(0) },
		"Up":           func(c *chord.Chord) error { return c.Up() },
		"Down":         func(c *chord.Chord) error { return c.Down() },
	}
	for name, op := range ops {
		t.Run("K-31 C4 E: "+name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, c4, eNo)
			before := stateOf(c)
			err := op(&c)
			assertNoChordSentinel(t, err)
			assert.ErrorIs(t, err, note.ErrNoOctave, "the cause of note: no octave (note Е-4)")
			assertState(t, before, c)
		})
	}

	upOps := map[string]func(c *chord.Chord) error{
		"Shift(1)":      func(c *chord.Chord) error { return c.Shift(1) },
		"ShiftDown(-1)": func(c *chord.Chord) error { return c.ShiftDown(-1) },
		"Up":            func(c *chord.Chord) error { return c.Up() },
	}
	for name, op := range upOps {
		t.Run("K-31 E4 X: "+name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, e4, cMax)
			before := stateOf(c)
			err := op(&c)
			assertNoChordSentinel(t, err)
			assert.ErrorIs(t, err, note.ErrOutOfRange, "the cause of note: out of range (note Е-8)")
			assert.ErrorIs(t, err, octave.ErrOutOfRange, "the cause of octave is kept")
			assertState(t, before, c, "E4 is not shifted")
		})
	}

	downOps := map[string]func(c *chord.Chord) error{
		"Shift(-1)":    func(c *chord.Chord) error { return c.Shift(-1) },
		"ShiftDown(1)": func(c *chord.Chord) error { return c.ShiftDown(1) },
		"Down":         func(c *chord.Chord) error { return c.Down() },
	}
	for name, op := range downOps {
		t.Run("lowest octave: "+name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, e4, cMin)
			before := stateOf(c)
			err := op(&c)
			assertNoChordSentinel(t, err)
			assert.ErrorIs(t, err, note.ErrOutOfRange)
			assert.ErrorIs(t, err, octave.ErrOutOfRange)
			assertState(t, before, c)
		})
	}
}

func TestTransformShiftFirstFailingNote(t *testing.T) {
	t.Parallel()
	// П-10а: the cause of the first note in the record order that is not
	// shifted.
	c := mk(t, eNo, cMax)
	err := c.Shift(1)
	assert.ErrorIs(t, err, note.ErrNoOctave)
	assert.NotErrorIs(t, err, note.ErrOutOfRange)

	d := mk(t, cMax, eNo)
	err = d.Shift(1)
	assert.ErrorIs(t, err, note.ErrOutOfRange)
	assert.NotErrorIs(t, err, note.ErrNoOctave)
}

// TestRulesProtocol checks П-9 and architecture 5.8 for the conditions and
// transforms: copies of the notes in the record order, at most one call for
// each note.
func TestRulesProtocol(t *testing.T) {
	t.Parallel()
	notes := list(pn(c4.Spelling(), 4, 0, std()), eNo, c4, g4, c4p30)
	tests := []struct {
		name string
		run  func(c *chord.Chord, r *recorder) error
	}{
		{name: "ContainsFunc", run: func(c *chord.Chord, r *recorder) error { _, err := c.ContainsFunc(r.cond(no)); return err }},
		{name: "CountFunc", run: func(c *chord.Chord, r *recorder) error { _, err := c.CountFunc(r.cond(yes)); return err }},
		{name: "IndicesFunc", run: func(c *chord.Chord, r *recorder) error { _, err := c.IndicesFunc(r.cond(yes)); return err }},
		{name: "NotesFunc", run: func(c *chord.Chord, r *recorder) error { _, err := c.NotesFunc(r.cond(yes)); return err }},
		{name: "RemoveFunc", run: func(c *chord.Chord, r *recorder) error { return c.RemoveFunc(r.cond(no)) }},
		{name: "KeepFunc", run: func(c *chord.Chord, r *recorder) error { return c.KeepFunc(r.cond(yes)) }},
		{name: "ReplaceFunc", run: func(c *chord.Chord, r *recorder) error {
			return c.ReplaceFunc(r.transform(func(n note.Note) (note.Note, error) { return n, nil }))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, notes...)
			before := stateOf(c)
			var r recorder
			require.NoError(t, tt.run(&c, &r))
			assertNotes(t, notes, r.got(), "the rule gets the notes in the record order, once each, with their sources")
			assertState(t, before, c, "a rule that changes its copy does not change the chord")
		})
	}
}

func TestRulesReplaceAtFuncCalledOnce(t *testing.T) {
	t.Parallel()
	c := mk(t, c4, e4, g4)
	var r recorder
	require.NoError(t, c.ReplaceAtFunc(1, r.transform(up)))
	assertNotes(t, list(e4), r.got(), "the transform is called once, for the note at 1")
	assertChord(t, list(c4, e5, g4), c)

	var r2 recorder
	assertOnly(t, c.ReplaceAtFunc(3, r2.transform(up)), chord.ErrIndex)
	assert.Zero(t, r2.calls(), "the transform is not called when the number is out of the chord")
}

func TestRulesContainsFuncStops(t *testing.T) {
	t.Parallel()
	c := mk(t, eNo, c4, cNo, g4)
	var r recorder
	ok, err := c.ContainsFunc(r.cond(func(n note.Note) bool { return n.HasOctave() }))
	require.NoError(t, err)
	assert.True(t, ok)
	assertNotes(t, list(eNo, c4), r.got(), "the search stops at the first matching note")
}

func TestRulesStopAtFirstError(t *testing.T) {
	t.Parallel()
	notes := list(c4, e4, g4, b4)
	tests := []struct {
		name string
		run  func(c *chord.Chord, r *recorder) error
	}{
		{name: "ContainsFunc", run: func(c *chord.Chord, r *recorder) error {
			_, err := c.ContainsFunc(r.condFailAt(2, no))
			return err
		}},
		{name: "CountFunc", run: func(c *chord.Chord, r *recorder) error { _, err := c.CountFunc(r.condFailAt(2, yes)); return err }},
		{name: "IndicesFunc", run: func(c *chord.Chord, r *recorder) error { _, err := c.IndicesFunc(r.condFailAt(2, yes)); return err }},
		{name: "NotesFunc", run: func(c *chord.Chord, r *recorder) error { _, err := c.NotesFunc(r.condFailAt(2, yes)); return err }},
		{name: "RemoveFunc", run: func(c *chord.Chord, r *recorder) error { return c.RemoveFunc(r.condFailAt(2, yes)) }},
		{name: "KeepFunc", run: func(c *chord.Chord, r *recorder) error { return c.KeepFunc(r.condFailAt(2, no)) }},
		{name: "ReplaceFunc", run: func(c *chord.Chord, r *recorder) error { return c.ReplaceFunc(r.transformFailAt(2, up)) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, notes...)
			before := stateOf(c)
			var r recorder
			assertRuleFailed(t, tt.run(&c, &r), errRule)
			assert.Equal(t, 2, r.calls(), "the rule is not called after its first error")
			assertState(t, before, c)
		})
	}
}

func TestRulesSortFuncStopsAtFirstError(t *testing.T) {
	t.Parallel()
	c := mk(t, g4, e4, c4, b4, d4, f4)
	before := stateOf(c)
	calls := 0
	err := c.SortFunc(func(note.Note, note.Note) (int, error) {
		calls++
		return 0, errRule
	})
	assertRuleFailed(t, err, errRule)
	assert.Equal(t, 1, calls, "the comparison is not called after its first error")
	assertState(t, before, c)
}

func TestRulesPanicPasses(t *testing.T) {
	t.Parallel()
	type boom struct{ at int }
	panicCond := func(k int) chord.Condition {
		calls := 0
		return func(note.Note) (bool, error) {
			calls++
			if calls == k {
				panic(boom{at: k})
			}
			return true, nil
		}
	}
	panicTransform := func(k int) chord.Transform {
		calls := 0
		return func(n note.Note) (note.Note, error) {
			calls++
			if calls == k {
				panic(boom{at: k})
			}
			err := n.Up()
			return n, err
		}
	}
	tests := []struct {
		name string
		run  func(c *chord.Chord)
	}{
		{name: "ContainsFunc", run: func(c *chord.Chord) {
			_, _ = c.ContainsFunc(func(note.Note) (bool, error) { panic(boom{at: 1}) })
		}},
		{name: "CountFunc", run: func(c *chord.Chord) { _, _ = c.CountFunc(panicCond(2)) }},
		{name: "IndicesFunc", run: func(c *chord.Chord) { _, _ = c.IndicesFunc(panicCond(2)) }},
		{name: "NotesFunc", run: func(c *chord.Chord) { _, _ = c.NotesFunc(panicCond(2)) }},
		{name: "RemoveFunc", run: func(c *chord.Chord) { _ = c.RemoveFunc(panicCond(2)) }},
		{name: "KeepFunc", run: func(c *chord.Chord) { _ = c.KeepFunc(panicCond(2)) }},
		{name: "ReplaceFunc", run: func(c *chord.Chord) { _ = c.ReplaceFunc(panicTransform(2)) }},
		{name: "ReplaceAtFunc", run: func(c *chord.Chord) { _ = c.ReplaceAtFunc(1, panicTransform(1)) }},
		{name: "SortFunc", run: func(c *chord.Chord) {
			_ = c.SortFunc(func(note.Note, note.Note) (int, error) { panic(boom{at: 1}) })
		}},
		{name: "MustRemoveFunc", run: func(c *chord.Chord) { c.MustRemoveFunc(panicCond(2)) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, c4, e4, g4)
			before := stateOf(c)
			v, panicked := recovered(func() { tt.run(&c) })
			require.True(t, panicked, "the panic of the rule passes to the caller")
			assert.IsType(t, boom{}, v, "the value of the panic is the value of the rule")
			assertState(t, before, c, "the chord is the same after the panic of the rule")
		})
	}
}

func TestRulesUnsetChordDoesNotCall(t *testing.T) {
	t.Parallel()
	calls := 0
	cond := func(note.Note) (bool, error) {
		calls++
		return true, nil
	}
	tr := func(n note.Note) (note.Note, error) {
		calls++
		return n, nil
	}
	cmp := func(note.Note, note.Note) (int, error) {
		calls++
		return 0, nil
	}
	var c chord.Chord
	assertOnly(t, c.RemoveFunc(cond), chord.ErrUnset)
	assertOnly(t, c.KeepFunc(cond), chord.ErrUnset)
	assertOnly(t, c.ReplaceFunc(tr), chord.ErrUnset)
	assertOnly(t, c.ReplaceAtFunc(0, tr), chord.ErrUnset)
	assertOnly(t, c.SortFunc(cmp), chord.ErrUnset)
	_, _ = c.ContainsFunc(cond)
	_, _ = c.CountFunc(cond)
	_, _ = c.IndicesFunc(cond)
	_, _ = c.NotesFunc(cond)
	assert.Zero(t, calls, "no rule is called for the unset chord")
	assertUnsetChord(t, c)
}

func TestRulesNil(t *testing.T) {
	t.Parallel()
	// П-9: an unset rule is E-1, also for reading methods and for the empty
	// chord.
	for name, c := range map[string]chord.Chord{"notes": mk(t, c4, e4), "empty": chord.MustNew()} {
		_, err := c.ContainsFunc(nil)
		assertOnly(t, err, chord.ErrUnset, name)
		_, err = c.CountFunc(nil)
		assertOnly(t, err, chord.ErrUnset, name)
		_, err = c.IndicesFunc(nil)
		assertOnly(t, err, chord.ErrUnset, name)
		_, err = c.NotesFunc(nil)
		assertOnly(t, err, chord.ErrUnset, name)
		before := stateOf(c)
		assertOnly(t, c.RemoveFunc(nil), chord.ErrUnset, name)
		assertOnly(t, c.KeepFunc(nil), chord.ErrUnset, name)
		assertOnly(t, c.ReplaceFunc(nil), chord.ErrUnset, name)
		assertOnly(t, c.SortFunc(nil), chord.ErrUnset, name)
		assertState(t, before, c)
	}
	c := mk(t, c4)
	assertOnly(t, c.ReplaceAtFunc(0, nil), chord.ErrUnset)
	assertOnly(t, c.ReplaceAtFunc(5, nil), chord.ErrUnset, "E-1 comes before E-3")
}

func TestRulesSortFuncGetsNotesOfChord(t *testing.T) {
	t.Parallel()
	// The comparison is called only with the notes of the chord.
	notes := list(g4, pn(e4.Spelling(), 4, 0, std()), c4)
	c := mk(t, notes...)
	var r recorder
	require.NoError(t, c.SortFunc(func(a, b note.Note) (int, error) {
		r.add(a)
		r.add(b)
		return a.Compare(b), nil
	}))
	for _, got := range r.got() {
		found := false
		for _, n := range notes {
			if sameNote(n, got) {
				found = true
			}
		}
		assert.True(t, found, "the comparison gets the note %v of the chord", got)
	}
	assertChord(t, list(c4, notes[1], g4), c)
}

func TestOrderStableLong(t *testing.T) {
	t.Parallel()
	// ОП-35, К-27: Sort and SortFunc are stable, also for 20 notes.
	t.Run("Sort", func(t *testing.T) {
		t.Parallel()
		notes := longBound()
		c := mk(t, notes...)
		require.NoError(t, c.Sort())
		assertChord(t, stableSorted(notes, note.Note.Compare), c)
	})
	t.Run("SortFunc by the names only", func(t *testing.T) {
		t.Parallel()
		notes := longNames()
		c := mk(t, notes...)
		require.NoError(t, c.SortFunc(byName))
		assertChord(t, stableSorted(notes, func(a, b note.Note) int { r, _ := byName(a, b); return r }), c)
	})
	t.Run("SortFunc by the agreed order, bound notes", func(t *testing.T) {
		t.Parallel()
		notes := longBound()
		c := mk(t, notes...)
		require.NoError(t, c.SortFunc(byRecord))
		assertChord(t, stableSorted(notes, note.Note.Compare), c)
	})
	t.Run("SortFunc that finds all notes equal", func(t *testing.T) {
		t.Parallel()
		notes := longBound()
		c := mk(t, notes...)
		require.NoError(t, c.SortFunc(func(note.Note, note.Note) (int, error) { return 0, nil }))
		assertChord(t, notes, c)
	})
}
