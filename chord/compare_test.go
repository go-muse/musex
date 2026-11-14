package chord_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-3. Comparison of chords: ОП-15…ОП-20; П-6; К-6, К-12…К-14 and the
// properties of П-6; architecture 5.9.

// unsetMark marks the unset chord in the tables of this file.
var unsetMark = []note.Note{{}}

// chordOf returns the unset chord for unsetMark, otherwise a chord.
func chordOf(t *testing.T, notes []note.Note) chord.Chord {
	t.Helper()
	if len(notes) == 1 && !notes[0].IsValid() {
		return chord.Chord{}
	}
	return mk(t, notes...)
}

func TestCompareEqualities(t *testing.T) {
	t.Parallel()
	st := std()
	s19 := set19()
	c4st := pn(c4.Spelling(), 4, 0, st)
	c419 := pn(c4.Spelling(), 4, 0, s19)
	tests := []struct {
		name                     string
		a, b                     []note.Note
		equal, inOrder, sameRecs bool
	}{
		{name: "K-12 C4 E4 G4 and G4 C4 E4", a: list(c4, e4, g4), b: list(g4, c4, e4), equal: true, inOrder: false, sameRecs: true},
		{name: "K-12 C4 E4 and C4 E4 E4", a: list(c4, e4), b: list(c4, e4, e4)},
		{name: "K-12 sources differ", a: list(c4st, e4), b: list(c419, e4), equal: true, inOrder: true, sameRecs: true},
		{name: "K-12 C4 E4 and C4+5 E4", a: list(c4, e4), b: list(c4p5, e4), sameRecs: true},
		{name: "K-12 C4 C4+5 and C4 C4", a: list(c4, c4p5), b: list(c4, c4), sameRecs: true},
		{name: "K-12 B♯4 E4 and C4 E4", a: list(bs4, e4), b: list(c4, e4)},
		{name: "K-12 C E4 and C4 E4", a: list(cNo, e4), b: list(c4, e4)},
		{name: "K-2 two empty chords", a: list(), b: list(), equal: true, inOrder: true, sameRecs: true},
		{name: "same chord", a: list(c4, e4, c4), b: list(c4, e4, c4), equal: true, inOrder: true, sameRecs: true},
		{name: "repeats in another order", a: list(c4, e4, c4), b: list(e4, c4, c4), equal: true, sameRecs: true},
		{name: "C4 C4+5 and C4", a: list(c4, c4p5), b: list(c4)},
		{name: "empty and C4", a: list(), b: list(c4)},
		{name: "K-2 empty and unset", a: list(), b: unsetMark},
		{name: "K-6 unset and unset", a: unsetMark, b: unsetMark},
		{name: "K-6 unset and C4", a: unsetMark, b: list(c4)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := chordOf(t, tt.a), chordOf(t, tt.b)
			for _, dir := range []struct {
				name string
				x, y chord.Chord
			}{{name: "a, b", x: a, y: b}, {name: "b, a", x: b, y: a}} {
				assert.Equal(t, tt.equal, dir.x.Equal(dir.y), "Equal(%s)", dir.name)
				assert.Equal(t, tt.inOrder, dir.x.EqualInOrder(dir.y), "EqualInOrder(%s)", dir.name)
				assert.Equal(t, tt.sameRecs, dir.x.SameRecords(dir.y), "SameRecords(%s)", dir.name)
			}
		})
	}
}

func TestCompareUnsetChord(t *testing.T) {
	t.Parallel()
	var u chord.Chord // K-6
	empty := chord.MustNew()
	one := chord.MustNew(c4)
	assert.False(t, u.Equal(u), "the unset chord is not equal to itself")
	assert.False(t, u.Equal(chord.Chord{}))
	assert.False(t, u.Equal(empty))
	assert.False(t, u.EqualInOrder(empty))
	assert.False(t, u.SameRecords(empty))
	assert.False(t, u.Includes(one))
	assert.False(t, one.Includes(u))
	assert.False(t, u.Includes(empty))
	assert.False(t, empty.Includes(u))
	assert.False(t, u.HasCommonNotes(one))
	assert.False(t, one.HasCommonNotes(u))
	assert.Equal(t, -1, u.Compare(empty))
	assert.Equal(t, 1, empty.Compare(u))
	assert.Equal(t, -1, u.Compare(one))
	assert.Equal(t, 1, one.Compare(u))
	assert.Equal(t, 0, u.Compare(chord.Chord{}))
}

func TestCompareIncludesAndCommon(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		a, b             []note.Note
		includes, common bool
	}{
		{name: "K-13 C4 E4 G4 and G4 C4", a: list(c4, e4, g4), b: list(g4, c4), includes: true, common: true},
		{name: "K-13 C4 E4 G4 and C4 C4", a: list(c4, e4, g4), b: list(c4, c4), includes: false, common: true},
		{name: "K-13 C4 C4 E4 and C4 C4", a: list(c4, c4, e4), b: list(c4, c4), includes: true, common: true},
		{name: "K-13 C4 and empty", a: list(c4), b: list(), includes: true, common: false},
		{name: "K-13 empty and C4", a: list(), b: list(c4), includes: false, common: false},
		{name: "K-13 common: C4 E4 and E4 G4", a: list(c4, e4), b: list(e4, g4), includes: false, common: true},
		{name: "K-13 common: C4 and C4+5", a: list(c4), b: list(c4p5), includes: false, common: false},
		{name: "itself", a: list(c4, e4, c4), b: list(c4, e4, c4), includes: true, common: true},
		{name: "empty and empty", a: list(), b: list(), includes: true, common: false},
		{name: "sources are not compared", a: list(pn(c4.Spelling(), 4, 0, std())), b: list(pn(c4.Spelling(), 4, 0, set19())), includes: true, common: true},
		{name: "C and C4 differ", a: list(cNo), b: list(c4), includes: false, common: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := mk(t, tt.a...), mk(t, tt.b...)
			assert.Equal(t, tt.includes, a.Includes(b), "Includes")
			assert.Equal(t, tt.common, a.HasCommonNotes(b), "HasCommonNotes")
			assert.Equal(t, tt.common, b.HasCommonNotes(a), "HasCommonNotes is symmetric")
		})
	}
}

func TestCompareOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b []note.Note
		want int
	}{
		{name: "K-14 empty and C4", a: list(), b: list(c4), want: -1},
		{name: "K-14 C4 and C4 E4", a: list(c4), b: list(c4, e4), want: -1},
		{name: "K-14 C4 E4 and E4 C4", a: list(c4, e4), b: list(e4, c4), want: 0},
		{name: "K-14 C4 E4 and C4 G4", a: list(c4, e4), b: list(c4, g4), want: -1},
		{name: "K-14 C G4 and C4", a: list(cNo, g4), b: list(c4), want: -1},
		{name: "K-14 C4 and C4+5", a: list(c4), b: list(c4p5), want: -1},
		{name: "K-14 C4 (st.) and C4 (19)", a: list(pn(c4.Spelling(), 4, 0, std())), b: list(pn(c4.Spelling(), 4, 0, set19())), want: 0},
		{name: "K-6 unset and empty", a: unsetMark, b: list(), want: -1},
		{name: "K-6 unset and unset", a: unsetMark, b: unsetMark, want: 0},
		{name: "by the agreed order, not the record order", a: list(g4, c4), b: list(e4, c4), want: 1},
		{name: "repeats", a: list(c4, c4), b: list(c4), want: 1},
		{name: "empty and empty", a: list(), b: list(), want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := chordOf(t, tt.a), chordOf(t, tt.b)
			assert.Equal(t, tt.want, a.Compare(b), "a.Compare(b)")
			assert.Equal(t, -tt.want, b.Compare(a), "the reverse order of the arguments changes the sign")
		})
	}
}

func TestCompareSortFunc(t *testing.T) {
	t.Parallel()
	// Architecture 5.9: Compare is suitable for slices.SortFunc.
	xs := []chord.Chord{
		mk(t, c4, g4),
		mk(t, c4, e4),
		{},
		mk(t),
		mk(t, cNo),
		mk(t, c4),
	}
	slices.SortFunc(xs, chord.Chord.Compare)
	require.Len(t, xs, 6)
	assert.False(t, xs[0].IsValid(), "the unset chord comes first")
	wants := [][]note.Note{list(), list(cNo), list(c4), list(c4, e4), list(c4, g4)}
	for i, want := range wants {
		assertChord(t, want, xs[i+1], "position %d", i+1)
	}
}

// TestCompareProperties checks the properties of П-6 and Р-17 on every pair of
// chords of up to 3 notes from a small pool, against oracles built from the
// definitions of architecture 5.9 with note.Note.Equal, SameRecord and Compare.
func TestCompareProperties(t *testing.T) {
	t.Parallel()
	p := pool()[:5] // C4, C4+5, C4 (st.), C, E4
	seqs := sequences(p, 3)
	chords := make([]chord.Chord, 0, len(seqs)+1)
	for _, s := range seqs {
		chords = append(chords, mk(t, s...))
	}
	chords = append(chords, chord.Chord{}) // the unset chord takes part where it may

	for i, a := range chords {
		for j, b := range chords {
			if !a.IsValid() || !b.IsValid() {
				assert.Equal(t, a.Compare(b), -b.Compare(a), "Compare is antisymmetric (%d, %d)", i, j)
				continue
			}
			x, y := seqs[i], seqs[j]

			eq := a.Equal(b)
			inOrder := a.EqualInOrder(b)
			same := a.SameRecords(b)
			inc := a.Includes(b)
			cmp := a.Compare(b)

			if !assert.Equal(t, sameComposition(x, y, equalNotes), eq, "Equal(%v, %v)", x, y) ||
				!assert.Equal(t, sameComposition(x, y, sameRecordNotes), same, "SameRecords(%v, %v)", x, y) ||
				!assert.Equal(t, includes(x, y), inc, "Includes(%v, %v)", x, y) ||
				!assert.Equal(t, common(x, y), a.HasCommonNotes(b), "HasCommonNotes(%v, %v)", x, y) ||
				!assert.Equal(t, compareNotes(x, y), cmp, "Compare(%v, %v)", x, y) {
				return
			}
			wantInOrder := len(x) == len(y)
			for k := range x {
				if wantInOrder && !x[k].Equal(y[k]) {
					wantInOrder = false
				}
			}
			if !assert.Equal(t, wantInOrder, inOrder, "EqualInOrder(%v, %v)", x, y) {
				return
			}

			// П-6.
			ok := assert.Equal(t, eq, b.Equal(a), "Equal is symmetric")
			ok = assert.Equal(t, inOrder, b.EqualInOrder(a), "EqualInOrder is symmetric") && ok
			ok = assert.Equal(t, same, b.SameRecords(a), "SameRecords is symmetric") && ok
			ok = assert.False(t, inOrder && !eq, "EqualInOrder implies Equal") && ok
			ok = assert.False(t, eq && !same, "Equal implies SameRecords") && ok
			ok = assert.Equal(t, eq, inc && b.Includes(a), "mutual Includes is Equal") && ok
			ok = assert.Equal(t, cmp, -b.Compare(a), "Compare is antisymmetric") && ok
			ok = assert.Contains(t, []int{-1, 0, 1}, cmp) && ok
			ok = assert.Equal(t, eq, cmp == 0, "Compare == 0 is Equal") && ok
			if !ok {
				t.Logf("a = %v, b = %v", x, y)
				return
			}
		}
		if a.IsValid() {
			assert.True(t, a.Equal(a), "Equal is reflexive for a set chord")
			assert.True(t, a.Includes(a), "a set chord includes itself")
			assert.True(t, a.Includes(chord.MustNew()), "a set chord includes the empty chord")
		}
	}
}

// TestCompareTransitivity checks the transitivity of Equal and of the order of
// Compare on the chords of up to 2 notes from the pool (П-6).
func TestCompareTransitivity(t *testing.T) {
	t.Parallel()
	seqs := sequences(pool()[:5], 2)
	chords := make([]chord.Chord, 0, len(seqs)+1)
	for _, s := range seqs {
		chords = append(chords, mk(t, s...))
	}
	chords = append(chords, chord.Chord{})
	for _, a := range chords {
		for _, b := range chords {
			ab := a.Compare(b)
			eqAB := a.Equal(b)
			for _, c := range chords {
				bc := b.Compare(c)
				if ab <= 0 && bc <= 0 && !assert.LessOrEqual(t, a.Compare(c), 0, "Compare is transitive: %v, %v, %v", a, b, c) {
					return
				}
				if eqAB && b.Equal(c) && !assert.True(t, a.Equal(c), "Equal is transitive: %v, %v, %v", a, b, c) {
					return
				}
			}
		}
	}
}

func TestCompareDoesNotChange(t *testing.T) {
	t.Parallel()
	a := mk(t, c4, e4, c4)
	b := mk(t, pn(e4.Spelling(), 4, 0, std()), g4)
	sa, sb := stateOf(a), stateOf(b)
	_ = a.Equal(b)
	_ = a.EqualInOrder(b)
	_ = a.SameRecords(b)
	_ = a.Includes(b)
	_ = a.HasCommonNotes(b)
	_ = a.Compare(b)
	assertState(t, sa, a)
	assertState(t, sb, b)
}
