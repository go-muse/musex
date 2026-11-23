package hindemith_test

import (
	"testing"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/problemsolving/hindemith"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dyadOf returns a constructor of the dyad measured from the note from to
// the note to in the traditional system.
func dyadOf(from, to note.Note) func(t *testing.T) hindemith.Dyad {
	return func(t *testing.T) hindemith.Dyad {
		t.Helper()
		d, err := hindemith.AnalyzeDyad(trad(), from, to)
		require.NoError(t, err)
		return d
	}
}

// unsetDyad returns a constructor of the zero Dyad.
func unsetDyad() func(t *testing.T) hindemith.Dyad {
	return func(*testing.T) hindemith.Dyad { return hindemith.Dyad{} }
}

// chordOf returns a constructor of the chord of the notes in the
// traditional system.
func chordOf(ns ...note.Note) func(t *testing.T) hindemith.Chord {
	return func(t *testing.T) hindemith.Chord {
		t.Helper()
		c, err := hindemith.AnalyzeChord(trad(), ns)
		require.NoError(t, err)
		return c
	}
}

// unsetChord returns a constructor of the zero Chord.
func unsetChord() func(t *testing.T) hindemith.Chord {
	return func(*testing.T) hindemith.Chord { return hindemith.Chord{} }
}

// rootState is the expected state of a value before it is compared: unset,
// set with a root, or set without a root. It guards each case against
// testing another situation than the one its name states.
type rootState uint8

const (
	unset rootState = iota
	withRoot
	rootless
)

// checkDyadState requires d to be in the state want.
func checkDyadState(t *testing.T, d hindemith.Dyad, want rootState, which string) {
	t.Helper()
	switch want {
	case unset:
		require.False(t, d.IsValid(), "%s: want an unset dyad", which)
	case withRoot:
		require.True(t, d.IsValid(), "%s: want a set dyad", which)
		require.True(t, d.HasRoot(), "%s: want a dyad with a root", which)
	case rootless:
		require.True(t, d.IsValid(), "%s: want a set dyad", which)
		require.False(t, d.HasRoot(), "%s: want a dyad without a root", which)
		require.True(t, d.IsTritone(), "%s: want the tritone", which)
	}
}

// checkChordState requires c to be in the state want.
func checkChordState(t *testing.T, c hindemith.Chord, want rootState, which string) {
	t.Helper()
	switch want {
	case unset:
		require.False(t, c.IsValid(), "%s: want an unset chord", which)
	case withRoot:
		require.True(t, c.IsValid(), "%s: want a set chord", which)
		require.True(t, c.HasRoot(), "%s: want a chord with a root", which)
	case rootless:
		require.True(t, c.IsValid(), "%s: want a set chord", which)
		require.False(t, c.HasRoot(), "%s: want a chord without a root", which)
		require.Contains(t, []hindemith.Group{hindemith.GroupV, hindemith.GroupVI}, c.Group(),
			"%s: a chord without a root is of group V or VI", which)
	}
}

// TestDyadEqualWithoutRoot is the regression test of issue #151 for
// Dyad.Equal: two set dyads are equal when the ordered pair of notes, the
// Series II position and, when the interval has one, the root are equal.
// The tritone has no root and is equal to an identical measurement. An unset
// dyad equals no dyad, not even another unset one.
func TestDyadEqualWithoutRoot(t *testing.T) {
	t.Parallel()

	c4, fSharp4 := n(spelling.C(), 4), n(spelling.FSharp(), 4)

	cases := []struct {
		name   string
		a, b   func(t *testing.T) hindemith.Dyad
		sa, sb rootState
		want   bool
	}{
		// Equal values without a root.
		{"tritone C4-F#4 analyzed twice", dyadOf(c4, fSharp4), dyadOf(c4, fSharp4), rootless, rootless, true},
		{"tritone C4-Gb4 analyzed twice", dyadOf(c4, n(spelling.GFlat(), 4)), dyadOf(c4, n(spelling.GFlat(), 4)), rootless, rootless, true},
		{"descending tritone F#4-C4 analyzed twice", dyadOf(fSharp4, c4), dyadOf(fSharp4, c4), rootless, rootless, true},
		{"compound tritone C4-F#5 analyzed twice", dyadOf(c4, n(spelling.FSharp(), 5)), dyadOf(c4, n(spelling.FSharp(), 5)), rootless, rootless, true},
		{
			"tritone of the chord evidence and its own analysis",
			func(t *testing.T) hindemith.Dyad {
				t.Helper()
				ds := chordOf(n(spelling.B(), 3), n(spelling.D(), 4), n(spelling.F(), 4))(t).Dyads()
				require.Len(t, ds, 3)
				return ds[1] // the pair B3-F4 in input order
			},
			dyadOf(n(spelling.B(), 3), n(spelling.F(), 4)),
			rootless, rootless, true,
		},

		// Controls with a root.
		{"perfect fifth C4-G4 analyzed twice", dyadOf(c4, n(spelling.G(), 4)), dyadOf(c4, n(spelling.G(), 4)), withRoot, withRoot, true},
		{"perfect fourth C4-F4 analyzed twice", dyadOf(c4, n(spelling.F(), 4)), dyadOf(c4, n(spelling.F(), 4)), withRoot, withRoot, true},

		// Values without a root that differ only in their notes.
		{"tritone from another note", dyadOf(c4, fSharp4), dyadOf(n(spelling.D(), 4), n(spelling.GSharp(), 4)), rootless, rootless, false},
		{"tritone in another octave", dyadOf(c4, fSharp4), dyadOf(n(spelling.C(), 5), n(spelling.FSharp(), 5)), rootless, rootless, false},
		{"tritone with the notes in the other order", dyadOf(c4, fSharp4), dyadOf(fSharp4, c4), rootless, rootless, false},
		{"tritone respelled at the same positions", dyadOf(c4, fSharp4), dyadOf(c4, n(spelling.GFlat(), 4)), rootless, rootless, false},
		{"simple and compound tritone", dyadOf(c4, fSharp4), dyadOf(c4, n(spelling.FSharp(), 5)), rootless, rootless, false},

		// A value without a root against a value with a root.
		{"tritone and perfect fifth from one note", dyadOf(c4, fSharp4), dyadOf(c4, n(spelling.G(), 4)), rootless, withRoot, false},
		{"tritone and perfect fourth from one note", dyadOf(c4, fSharp4), dyadOf(c4, n(spelling.F(), 4)), rootless, withRoot, false},

		// Unset values.
		{"tritone and unset", dyadOf(c4, fSharp4), unsetDyad(), rootless, unset, false},
		{"perfect fifth and unset", dyadOf(c4, n(spelling.G(), 4)), unsetDyad(), withRoot, unset, false},
		{"two unset dyads", unsetDyad(), unsetDyad(), unset, unset, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a, b := tc.a(t), tc.b(t)
			checkDyadState(t, a, tc.sa, "a")
			checkDyadState(t, b, tc.sb, "b")

			assert.Equal(t, tc.want, a.Equal(b), "a.Equal(b)")
			assert.Equal(t, tc.want, b.Equal(a), "b.Equal(a): symmetry")
		})
	}
}

// TestChordEqualWithoutRoot is the regression test of issue #151 for
// Chord.Equal: two set chords are equal when the notes in the same order,
// the group, the subgroup, the bass and, when the chord has one, the root
// are equal. Chords of groups V and VI have no root and are equal to an
// identical analysis. An unset chord equals no chord, not even another unset
// one.
func TestChordEqualWithoutRoot(t *testing.T) {
	t.Parallel()

	augmented := []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.GSharp(), 4)}
	quartal := []note.Note{n(spelling.C(), 3), n(spelling.G(), 3), n(spelling.D(), 4)}
	diminished := []note.Note{n(spelling.B(), 3), n(spelling.D(), 4), n(spelling.F(), 4)}
	major := []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)}

	cases := []struct {
		name   string
		a, b   func(t *testing.T) hindemith.Chord
		sa, sb rootState
		want   bool
	}{
		// Equal values without a root.
		{"group V augmented triad C4 E4 G#4 analyzed twice", chordOf(augmented...), chordOf(augmented...), rootless, rootless, true},
		{"group V quartal triad C3 G3 D4 analyzed twice", chordOf(quartal...), chordOf(quartal...), rootless, rootless, true},
		{"group V quartal triad C4 F4 Bb4 analyzed twice",
			chordOf(n(spelling.C(), 4), n(spelling.F(), 4), n(spelling.BFlat(), 4)),
			chordOf(n(spelling.C(), 4), n(spelling.F(), 4), n(spelling.BFlat(), 4)),
			rootless, rootless, true},
		{"group V doubled augmented triad analyzed twice",
			chordOf(n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.GSharp(), 4), n(spelling.E(), 5)),
			chordOf(n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.GSharp(), 4), n(spelling.E(), 5)),
			rootless, rootless, true},
		{"group VI diminished triad B3 D4 F4 analyzed twice", chordOf(diminished...), chordOf(diminished...), rootless, rootless, true},
		{"group VI diminished seventh analyzed twice",
			chordOf(n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4), n(spelling.A(), 4)),
			chordOf(n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4), n(spelling.A(), 4)),
			rootless, rootless, true},
		{"group VI tritone C4 F#4 analyzed twice",
			chordOf(n(spelling.C(), 4), n(spelling.FSharp(), 4)),
			chordOf(n(spelling.C(), 4), n(spelling.FSharp(), 4)),
			rootless, rootless, true},
		{
			"group V chord of a progression and its own analysis",
			func(t *testing.T) hindemith.Chord {
				t.Helper()
				p, err := hindemith.AnalyzeProgression(trad(), [][]note.Note{major, augmented})
				require.NoError(t, err)
				cs := p.Chords()
				require.Len(t, cs, 2)
				return cs[1]
			},
			chordOf(augmented...),
			rootless, rootless, true,
		},

		// Controls with a root.
		{"I.1 C4 E4 G4 analyzed twice", chordOf(major...), chordOf(major...), withRoot, withRoot, true},
		{"I.2 E4 G4 C5 analyzed twice",
			chordOf(n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.C(), 5)),
			chordOf(n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.C(), 5)),
			withRoot, withRoot, true},

		// Values without a root that differ only in their notes.
		{"augmented triad from another note", chordOf(augmented...),
			chordOf(n(spelling.D(), 4), n(spelling.FSharp(), 4), n(spelling.ASharp(), 4)), rootless, rootless, false},
		{"augmented triad in another octave", chordOf(augmented...),
			chordOf(n(spelling.C(), 5), n(spelling.E(), 5), n(spelling.GSharp(), 5)), rootless, rootless, false},
		{"augmented triad with the notes in another order", chordOf(augmented...),
			chordOf(n(spelling.C(), 4), n(spelling.GSharp(), 4), n(spelling.E(), 4)), rootless, rootless, false},
		{"augmented triad respelled at the same positions", chordOf(augmented...),
			chordOf(n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.AFlat(), 4)), rootless, rootless, false},
		{"augmented triad with an added doubling", chordOf(augmented...),
			chordOf(n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.GSharp(), 4), n(spelling.E(), 5)), rootless, rootless, false},
		{"quartal triad with the notes in another order", chordOf(quartal...),
			chordOf(n(spelling.G(), 3), n(spelling.C(), 3), n(spelling.D(), 4)), rootless, rootless, false},
		{"quartal triad from another note", chordOf(quartal...),
			chordOf(n(spelling.D(), 3), n(spelling.A(), 3), n(spelling.E(), 4)), rootless, rootless, false},
		{"diminished triad in another octave", chordOf(diminished...),
			chordOf(n(spelling.B(), 4), n(spelling.D(), 5), n(spelling.F(), 5)), rootless, rootless, false},
		{"diminished triad and diminished seventh", chordOf(diminished...),
			chordOf(n(spelling.B(), 3), n(spelling.D(), 4), n(spelling.F(), 4), n(spelling.AFlat(), 4)), rootless, rootless, false},

		// Values without a root that differ in the group or the bass.
		{"group V and group VI", chordOf(augmented...),
			chordOf(n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4)), rootless, rootless, false},
		{"augmented triad with another bass", chordOf(augmented...),
			chordOf(n(spelling.E(), 4), n(spelling.GSharp(), 4), n(spelling.C(), 5)), rootless, rootless, false},
		{"diminished triad with another bass", chordOf(diminished...),
			chordOf(n(spelling.D(), 4), n(spelling.F(), 4), n(spelling.B(), 4)), rootless, rootless, false},

		// A value without a root against a value with a root.
		{"group V and I.1 differing in one note", chordOf(augmented...), chordOf(major...), rootless, withRoot, false},
		{"group VI and I.2 differing in one note", chordOf(diminished...),
			chordOf(n(spelling.B(), 3), n(spelling.D(), 4), n(spelling.G(), 4)), rootless, withRoot, false},
		{"group VI tritone and I.1 fifth from one note",
			chordOf(n(spelling.C(), 4), n(spelling.FSharp(), 4)),
			chordOf(n(spelling.C(), 4), n(spelling.G(), 4)), rootless, withRoot, false},

		// Unset values.
		{"group V and unset", chordOf(augmented...), unsetChord(), rootless, unset, false},
		{"group VI and unset", chordOf(diminished...), unsetChord(), rootless, unset, false},
		{"I.1 and unset", chordOf(major...), unsetChord(), withRoot, unset, false},
		{"two unset chords", unsetChord(), unsetChord(), unset, unset, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a, b := tc.a(t), tc.b(t)
			checkChordState(t, a, tc.sa, "a")
			checkChordState(t, b, tc.sb, "b")

			assert.Equal(t, tc.want, a.Equal(b), "a.Equal(b)")
			assert.Equal(t, tc.want, b.Equal(a), "b.Equal(a): symmetry")
		})
	}
}

// TestEqualIsEquivalenceOnSetValues checks that Dyad.Equal and Chord.Equal
// are reflexive and transitive on set values with and without a root (issue
// #151): a value equals itself, and three independent analyses of the same
// notes are pairwise equal. Unset values stay outside: an unset value does
// not equal itself.
func TestEqualIsEquivalenceOnSetValues(t *testing.T) {
	t.Parallel()

	dyads := []struct {
		name     string
		from, to note.Note
	}{
		{"tritone C4-F#4", n(spelling.C(), 4), n(spelling.FSharp(), 4)},
		{"descending tritone F#4-C4", n(spelling.FSharp(), 4), n(spelling.C(), 4)},
		{"perfect fifth C4-G4", n(spelling.C(), 4), n(spelling.G(), 4)},
	}
	for _, tc := range dyads {
		t.Run("dyad "+tc.name, func(t *testing.T) {
			t.Parallel()
			d1 := hindemith.MustAnalyzeDyad(trad(), tc.from, tc.to)
			d2 := hindemith.MustAnalyzeDyad(trad(), tc.from, tc.to)
			d3 := hindemith.MustAnalyzeDyad(trad(), tc.from, tc.to)
			assert.True(t, d1.Equal(d1), "reflexivity")
			assert.True(t, d1.Equal(d2), "d1 and d2")
			assert.True(t, d2.Equal(d3), "d2 and d3")
			assert.True(t, d1.Equal(d3), "d1 and d3")
		})
	}

	chords := []struct {
		name  string
		notes []note.Note
	}{
		{"group V C4 E4 G#4", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.GSharp(), 4)}},
		{"group V C3 G3 D4", []note.Note{n(spelling.C(), 3), n(spelling.G(), 3), n(spelling.D(), 4)}},
		{"group VI B3 D4 F4", []note.Note{n(spelling.B(), 3), n(spelling.D(), 4), n(spelling.F(), 4)}},
		{"I.1 C4 E4 G4", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)}},
	}
	for _, tc := range chords {
		t.Run("chord "+tc.name, func(t *testing.T) {
			t.Parallel()
			c1 := hindemith.MustAnalyzeChord(trad(), tc.notes)
			c2 := hindemith.MustAnalyzeChord(trad(), tc.notes)
			c3 := hindemith.MustAnalyzeChord(trad(), tc.notes)
			assert.True(t, c1.Equal(c1), "reflexivity")
			assert.True(t, c1.Equal(c2), "c1 and c2")
			assert.True(t, c2.Equal(c3), "c2 and c3")
			assert.True(t, c1.Equal(c3), "c1 and c3")
		})
	}

	t.Run("unset values", func(t *testing.T) {
		t.Parallel()
		var d hindemith.Dyad
		var c hindemith.Chord
		assert.False(t, d.Equal(d), "an unset dyad does not equal itself")
		assert.False(t, c.Equal(c), "an unset chord does not equal itself")
	})
}
