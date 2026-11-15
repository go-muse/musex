package chord_test

import (
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-5. Operations with another chord: ОП-32…ОП-34; П-10, Р-10, Р-11;
// К-23…К-25, К-42; architecture 5.10 (Union, Intersect, Subtract).

// setOp is one operation with another chord.
type setOp struct {
	name string
	op   func(c *chord.Chord, other chord.Chord) error
}

var setOps = []setOp{
	{name: "Union", op: (*chord.Chord).Union},
	{name: "Intersect", op: (*chord.Chord).Intersect},
	{name: "Subtract", op: (*chord.Chord).Subtract},
}

func TestSetOpsExamples(t *testing.T) {
	t.Parallel()
	st1 := stdSource() // two different sources with the standard settings
	st2 := stdSource()
	s19 := set19()
	c4st1 := pn(c4.Spelling(), 4, 0, st1)
	c4st2 := pn(c4.Spelling(), 4, 0, st2)
	c419 := pn(c4.Spelling(), 4, 0, s19)
	tests := []struct {
		name        string
		base, other []note.Note
		op          func(c *chord.Chord, other chord.Chord) error
		want        []note.Note
	}{
		// Union (K-23).
		{name: "K-23 union with E4 E4 G4", base: list(c4, e4), other: list(e4, e4, g4), op: (*chord.Chord).Union, want: list(c4, e4, e4, g4)},
		{name: "K-23 union with the empty chord", base: list(c4, e4), other: list(), op: (*chord.Chord).Union, want: list(c4, e4)},
		{name: "K-23 union of C4 (st.) with C4 (st.) C4 (19)", base: list(c4st1), other: list(c4st2, c419), op: (*chord.Chord).Union, want: list(c4st1, c419)},
		{name: "union adds the occurrences from p+1 to q in the order of the argument", base: list(c4), other: list(g4, e4, c4, c4), op: (*chord.Chord).Union, want: list(c4, g4, e4, c4)},
		{name: "union adds the third occurrence with its source", base: list(c4, e4, c4), other: list(c4, c4, c419), op: (*chord.Chord).Union, want: list(c4, e4, c4, c419)},
		{name: "union into the empty chord", base: list(), other: list(e4, c4, e4), op: (*chord.Chord).Union, want: list(e4, c4, e4)},
		{name: "union without repeats is the union of sets", base: list(c4, e4), other: list(g4, c4), op: (*chord.Chord).Union, want: list(c4, e4, g4)},
		// Intersect (K-24).
		{name: "K-24 intersect C4 E4 E4 G4 with E4 G4 G4 B4", base: list(c4, e4, e4, g4), other: list(e4, g4, g4, b4), op: (*chord.Chord).Intersect, want: list(e4, g4)},
		{name: "K-24 the note of the receiver keeps its source", base: list(c4st1, e4), other: list(c419), op: (*chord.Chord).Intersect, want: list(c4st1)},
		{name: "K-24 intersect with the empty chord", base: list(c4, e4), other: list(), op: (*chord.Chord).Intersect, want: list()},
		{name: "intersect removes the last occurrences", base: list(c4st1, e4, c419, c4st2), other: list(c4, c4), op: (*chord.Chord).Intersect, want: list(c4st1, c419)},
		{name: "intersect keeps the record order of the receiver", base: list(g4, c4, e4), other: list(e4, g4), op: (*chord.Chord).Intersect, want: list(g4, e4)},
		{name: "intersect of the empty chord", base: list(), other: list(c4), op: (*chord.Chord).Intersect, want: list()},
		// Subtract (K-25).
		{name: "K-25 C4 E4 C4 G4 C4 minus C4 C4 D4", base: list(c4, e4, c4, g4, c4), other: list(c4, c4, d4), op: (*chord.Chord).Subtract, want: list(e4, g4, c4)},
		{name: "K-25 minus the empty chord", base: list(c4, e4), other: list(), op: (*chord.Chord).Subtract, want: list(c4, e4)},
		{name: "subtract removes the first occurrences", base: list(c4st1, e4, c419), other: list(c4), op: (*chord.Chord).Subtract, want: list(e4, c419)},
		{name: "subtract a note that is absent", base: list(c4, e4), other: list(g4, c4p5), op: (*chord.Chord).Subtract, want: list(c4, e4)},
		{name: "subtract from the empty chord", base: list(), other: list(c4), op: (*chord.Chord).Subtract, want: list()},
		{name: "subtract everything", base: list(c4, e4), other: list(e4, c4, c4), op: (*chord.Chord).Subtract, want: list()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, tt.base...)
			other := mk(t, tt.other...)
			otherBefore := stateOf(other)
			require.NoError(t, tt.op(&c, other))
			assertChord(t, tt.want, c)
			assertState(t, otherBefore, other, "the argument chord is the same")
		})
	}
}

func TestSetOpsWithItself(t *testing.T) {
	t.Parallel()
	// G-5, P-10: K-23, K-24, K-25 with a copy of the chord itself.
	base := list(c4, e4, c4, pn(g4.Spelling(), 4, 0, std()))
	tests := []struct {
		name string
		op   func(c *chord.Chord, other chord.Chord) error
		want []note.Note
	}{
		{name: "K-23 union", op: (*chord.Chord).Union, want: base},
		{name: "K-24 intersect", op: (*chord.Chord).Intersect, want: base},
		{name: "K-25 subtract", op: (*chord.Chord).Subtract, want: list()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, base...)
			require.NoError(t, tt.op(&c, c))
			assertChord(t, tt.want, c)
		})
	}
}

func TestSetOpsUnset(t *testing.T) {
	t.Parallel()
	for _, so := range setOps {
		t.Run(so.name, func(t *testing.T) {
			t.Parallel()
			// K-42: an unset argument chord.
			c := mk(t, c4, e4)
			before := stateOf(c)
			assertOnly(t, so.op(&c, chord.Chord{}), chord.ErrUnset)
			assertState(t, before, c)

			// An unset receiver, with a set and an unset argument.
			var u chord.Chord
			assertOnly(t, so.op(&u, mk(t, c4)), chord.ErrUnset)
			assertUnsetChord(t, u)
			assertOnly(t, so.op(&u, chord.Chord{}), chord.ErrUnset)
			assertUnsetChord(t, u)

			// A nil receiver.
			var p *chord.Chord
			assertOnly(t, so.op(p, mk(t, c4)), chord.ErrUnset)
		})
	}
}

// TestSetOpsProperties checks the three operations on every pair of chords of
// up to 3 notes of the pool against the definitions of P-10 (oracles in
// helpers_test.go): the exact notes with their sources and the multiplicities
// of R-10 (the greatest, the least, the difference not below 0).
func TestSetOpsProperties(t *testing.T) {
	t.Parallel()
	seqs := sequences(pool()[:5], 3) // C4, C4+5, C4 (st.), C, E4
	for _, a := range seqs {
		for _, b := range seqs {
			other := mk(t, b...)
			otherBefore := stateOf(other)

			u := mk(t, a...)
			require.NoError(t, u.Union(other))
			in := mk(t, a...)
			require.NoError(t, in.Intersect(other))
			sub := mk(t, a...)
			require.NoError(t, sub.Subtract(other))

			ok := assertChord(t, unionOf(a, b), u, "Union(%v, %v)", a, b)
			ok = assertChord(t, intersectOf(a, b), in, "Intersect(%v, %v)", a, b) && ok
			ok = assertChord(t, subtractOf(a, b), sub, "Subtract(%v, %v)", a, b) && ok
			for _, x := range append(append([]note.Note{}, a...), b...) {
				ca, cb := count(a, x, equalNotes), count(b, x, equalNotes)
				ok = assert.Equal(t, max(ca, cb), u.Count(x), "Union: multiplicity of %v in %v, %v", x, a, b) && ok
				ok = assert.Equal(t, min(ca, cb), in.Count(x), "Intersect: multiplicity of %v in %v, %v", x, a, b) && ok
				ok = assert.Equal(t, max(ca-cb, 0), sub.Count(x), "Subtract: multiplicity of %v in %v, %v", x, a, b) && ok
			}
			ok = assertState(t, otherBefore, other, "the argument chord is the same") && ok
			// R-10: c includes d when d minus c is empty.
			d := mk(t, b...)
			require.NoError(t, d.Subtract(mk(t, a...)))
			ok = assert.Equal(t, mk(t, a...).Includes(other), d.IsEmpty(), "Includes(%v, %v) and Subtract", a, b) && ok
			if !ok {
				return
			}
		}
	}
}
