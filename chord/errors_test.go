package chord_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-7, errors and the order of the causes: Е-1…Е-4; П-10а, П-11; К-42, К-43;
// architecture 5.4.

func TestErrorsSentinels(t *testing.T) {
	t.Parallel()
	for i, a := range sentinels {
		require.Error(t, a)
		for j, b := range sentinels {
			if i != j {
				assert.NotErrorIs(t, a, b, "the sentinels never wrap each other: %v, %v", a, b)
			}
		}
	}
	assert.NotErrorIs(t, chord.ErrUnset, note.ErrUnset, "E-1 of chord is not the E-1 of note")
}

func TestErrorsUnsetNoteIsNotNoteError(t *testing.T) {
	t.Parallel()
	// Architecture 5.4: an unset note is found by the chord; the error matches
	// ErrUnset of chord only, not note.ErrUnset.
	c := mk(t, c4)
	errs := map[string]error{
		"Add":     c.Add(note.Note{}),
		"Replace": c.Replace(c4, note.Note{}),
	}
	_, errs["New"] = chord.New(note.Note{})
	errs["ReplaceFunc"] = c.ReplaceFunc(func(note.Note) (note.Note, error) { return note.Note{}, nil })
	for name, err := range errs {
		assertOnly(t, err, chord.ErrUnset, name)
		assert.NotErrorIs(t, err, note.ErrUnset, name)
	}
}

// TestErrorsOrder checks П-11 and К-43 on inputs with two violations: the
// first cause in the order Е-1 of the inputs → Е-3 → Е-2 → the execution.
func TestErrorsOrder(t *testing.T) {
	t.Parallel()
	failTr := func(n note.Note) (note.Note, error) { return n, errRule }
	tests := []struct {
		name   string
		base   []note.Note // nil: the unset chord
		op     func(c *chord.Chord) error
		target error
	}{
		{name: "K-43 replace at 5 with an unset note: E-1 before E-3", base: list(c4, e4), op: func(c *chord.Chord) error { return c.ReplaceAt(5, note.Note{}) }, target: chord.ErrUnset},
		{name: "K-43 insert an unset note at 9: E-1 before E-3", base: list(c4, e4), op: func(c *chord.Chord) error { return c.Insert(9, note.Note{}) }, target: chord.ErrUnset},
		{name: "K-43 remove an unset note: E-1 before E-2", base: list(c4, e4), op: func(c *chord.Chord) error { return c.Remove(note.Note{}) }, target: chord.ErrUnset},
		{name: "replace an absent note with an unset note: E-1 before E-2", base: list(c4, e4), op: func(c *chord.Chord) error { return c.Replace(f4, note.Note{}) }, target: chord.ErrUnset},
		{name: "replace all of an unset note: E-1 before E-2", base: list(c4, e4), op: func(c *chord.Chord) error { return c.ReplaceAll(note.Note{}, d4) }, target: chord.ErrUnset},
		{name: "a nil rule and a number out of the chord: E-1 before E-3", base: list(c4), op: func(c *chord.Chord) error { return c.ReplaceAtFunc(5, nil) }, target: chord.ErrUnset},
		{name: "a failing rule and a number out of the chord: E-3 before E-4", base: list(c4), op: func(c *chord.Chord) error { return c.ReplaceAtFunc(5, failTr) }, target: chord.ErrIndex},
		{name: "an unset chord and an unset note", base: nil, op: func(c *chord.Chord) error { return c.Add(note.Note{}) }, target: chord.ErrUnset},
		{name: "an unset chord and a number out of the chord", base: nil, op: func(c *chord.Chord) error { return c.RemoveAt(5) }, target: chord.ErrUnset},
		{name: "an unset chord and an absent note", base: nil, op: func(c *chord.Chord) error { return c.Remove(d4) }, target: chord.ErrUnset},
		{name: "an unset chord and a nil rule", base: nil, op: func(c *chord.Chord) error { return c.SortFunc(nil) }, target: chord.ErrUnset},
		{name: "an unset chord and an unset argument chord", base: nil, op: func(c *chord.Chord) error { return c.Union(chord.Chord{}) }, target: chord.ErrUnset},
		{name: "an unset note and a number out of the chord in Insert with an empty chord", base: list(), op: func(c *chord.Chord) error { return c.Insert(1, note.Note{}) }, target: chord.ErrUnset},
		{name: "a transform returns an unset note with an error: E-4", base: list(c4, e4), op: func(c *chord.Chord) error {
			return c.ReplaceFunc(func(note.Note) (note.Note, error) { return note.Note{}, errRule })
		}, target: chord.ErrRuleFailed},
		{name: "at a number: a transform returns an unset note with an error: E-4", base: list(c4, e4), op: func(c *chord.Chord) error {
			return c.ReplaceAtFunc(1, func(note.Note) (note.Note, error) { return note.Note{}, errRule })
		}, target: chord.ErrRuleFailed},
		{name: "the first note returns an unset note, the second fails: E-1", base: list(c4, e4), op: func(c *chord.Chord) error {
			var r recorder
			return c.ReplaceFunc(r.transformFailAt(2, func(note.Note) (note.Note, error) { return note.Note{}, nil }))
		}, target: chord.ErrUnset},
		{name: "the first note fails, the second returns an unset note: E-4", base: list(c4, e4), op: func(c *chord.Chord) error {
			var r recorder
			return c.ReplaceFunc(r.transformFailAt(1, func(note.Note) (note.Note, error) { return note.Note{}, nil }))
		}, target: chord.ErrRuleFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var c chord.Chord
			if tt.base != nil {
				c = mk(t, tt.base...)
			}
			before := stateOf(c)
			err := tt.op(&c)
			if tt.target == chord.ErrRuleFailed {
				assertRuleFailed(t, err, errRule)
			} else {
				assertOnly(t, err, tt.target)
			}
			assertState(t, before, c)
		})
	}
}

func TestErrorsAtOrder(t *testing.T) {
	t.Parallel()
	var u chord.Chord
	_, err := u.At(5)
	assertOnly(t, err, chord.ErrUnset, "At: E-1 before E-3")
}

func TestErrorsRuleErrorMatchingSentinel(t *testing.T) {
	t.Parallel()
	// Architecture 5.4: the error of a rule may itself match a sentinel of
	// chord; E-4 keeps it.
	c := mk(t, c4, e4)
	before := stateOf(c)
	ruleErr := fmt.Errorf("rule: %w", chord.ErrNotFound)
	err := c.RemoveFunc(func(note.Note) (bool, error) { return false, ruleErr })
	assert.ErrorIs(t, err, chord.ErrRuleFailed)
	assert.ErrorIs(t, err, ruleErr)
	assert.ErrorIs(t, err, chord.ErrNotFound, "the chain of the error of the rule is kept")
	assertState(t, before, c)
}

func TestErrorsRuleChainKept(t *testing.T) {
	t.Parallel()
	// П-10а: the cause of the rule is available through errors.Is and
	// errors.As.
	cause := ruleError{err: errRule}
	c := mk(t, c4, e4)
	_, err := c.CountFunc(func(note.Note) (bool, error) { return false, cause })
	assertRuleFailed(t, err, errRule)
	assert.ErrorIs(t, err, cause, "the error of the rule itself is in the chain")
	var got ruleError
	require.True(t, errors.As(err, &got), "errors.As finds the error of the rule")
	assert.Equal(t, cause, got)
}

// ruleError is an error type of a rule that wraps its cause.
type ruleError struct{ err error }

func (e ruleError) Error() string { return "rule error: " + e.err.Error() }

func (e ruleError) Unwrap() error { return e.err }

func TestErrorsResultsOnFailure(t *testing.T) {
	t.Parallel()
	// Architecture 5.4, "the result on an error".
	c, err := chord.New(c4, note.Note{})
	require.Error(t, err)
	assertUnsetChord(t, c)

	d := mk(t, c4)
	n, err := d.At(1)
	require.Error(t, err)
	assert.False(t, n.IsValid())

	ok, err := d.ContainsFunc(nil)
	require.Error(t, err)
	assert.False(t, ok)
	k, err := d.CountFunc(nil)
	require.Error(t, err)
	assert.Zero(t, k)
	idx, err := d.IndicesFunc(nil)
	require.Error(t, err)
	assert.Nil(t, idx)
	notes, err := d.NotesFunc(nil)
	require.Error(t, err)
	assert.Nil(t, notes)
}

func TestErrorsReadingThroughNilPointer(t *testing.T) {
	t.Parallel()
	// Architecture 5.2: a reading method through a nil *Chord panics by the
	// rules of Go.
	var p *chord.Chord
	assert.Panics(t, func() { _ = p.Len() })
	assert.Panics(t, func() { _ = p.IsValid() })
	assert.Panics(t, func() { _ = p.Contains(c4) })
}

func TestErrorsNoPanics(t *testing.T) {
	t.Parallel()
	// Architecture 5.4: no panics outside Must variants and rules. Every
	// method with valid or invalid inputs on the unset, empty and filled
	// chord returns.
	for name, base := range map[string]*chord.Chord{
		"unset":  {},
		"empty":  new(chord.MustNew()),
		"filled": new(chord.MustNew(c4, eNo, cMax, c4)),
	} {
		for opName, op := range changingOps() {
			c := *base
			assert.NotPanics(t, func() { _ = op(&c) }, "%s: %s", name, opName)
		}
		c := *base
		assert.NotPanics(t, func() {
			_, _ = c.At(-1)
			_, _ = c.At(100)
			_ = c.Notes()
			_ = c.SortedNotes()
			_ = c.DistinctNotes()
			_ = c.HasRepeats()
			_, _, _ = c.Find(note.Note{})
			_ = c.Indices(note.Note{})
			_ = c.ContainsAll(note.Note{})
			_ = c.ContainsAny(note.Note{})
			_, _ = c.ContainsFunc(nil)
			_ = c.Equal(chord.Chord{})
			_ = c.Compare(chord.Chord{})
			_ = c.String()
			_ = c.Insert(-5, note.Note{})
			_ = c.ReplaceAt(-5, note.Note{})
			_ = c.Shift(1)
			_ = c.Shift(-1)
		}, name)
	}
}
