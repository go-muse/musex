package hindemith_test

import (
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/problemsolving/hindemith"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

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

		// Values without a root whose notes differ, and with them the group
		// or the bass. Differences in the classification alone, for the same
		// notes, are covered by TestEqualComparesTheClassification.
		{"other notes of group V and of group VI", chordOf(augmented...),
			chordOf(n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4)), rootless, rootless, false},
		{"augmented triad inverted to other notes and another bass", chordOf(augmented...),
			chordOf(n(spelling.E(), 4), n(spelling.GSharp(), 4), n(spelling.C(), 5)), rootless, rootless, false},
		{"diminished triad inverted to other notes and another bass", chordOf(diminished...),
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

// twelveTone returns a system of twelve classes with the boundary letter C,
// chromatic unit 1 and the integer alteration values, with the offsets of
// the letters C D E F G A B. Every such system is admitted by the package
// (ХН-П1), so the same notes may be classified differently in two of them.
func twelveTone(name string, offsets [letter.Count]int) *system.System {
	return system.MustNew(system.Definition{
		Name:        name,
		Classes:     12,
		Boundary:    letter.C,
		Offsets:     offsets,
		Chromatic:   1,
		Alterations: system.Integers(),
	})
}

// Twelve-tone systems that differ from the traditional one (C 0 D 2 E 4 F 5
// G 7 A 9 B 11) in the offsets of one or more letters.
var (
	sysD1    = twelveTone("d1", [letter.Count]int{0, 1, 4, 5, 7, 9, 11})
	sysB9    = twelveTone("b9", [letter.Count]int{0, 2, 4, 5, 7, 8, 9})
	sysE5    = twelveTone("e5", [letter.Count]int{0, 2, 5, 5, 7, 9, 11})
	sysG6    = twelveTone("g6", [letter.Count]int{0, 2, 4, 5, 6, 9, 11})
	sysMinor = twelveTone("minor", [letter.Count]int{0, 2, 3, 5, 7, 8, 10})
)

// sameNotes reports whether a and b are the same notes in the same order.
func sameNotes(a, b []note.Note) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}
	return true
}

// sameRoot reports whether two optional roots agree: both absent, or both
// present and equal notes.
func sameRoot(hasA bool, a note.Note, hasB bool, b note.Note) bool {
	if hasA != hasB {
		return false
	}
	return !hasA || a.Equal(b)
}

// dyadDiff lists the criteria of Dyad.Equal in which the set dyads a and b
// differ: "notes" (the ordered pair), "series" and "root".
func dyadDiff(a, b hindemith.Dyad) []string {
	var diff []string
	if !a.From().Equal(b.From()) || !a.To().Equal(b.To()) {
		diff = append(diff, "notes")
	}
	if a.Series() != b.Series() {
		diff = append(diff, "series")
	}
	if !sameRoot(a.HasRoot(), a.Root(), b.HasRoot(), b.Root()) {
		diff = append(diff, "root")
	}
	return diff
}

// chordDiff lists the criteria of Chord.Equal in which the set chords a and
// b differ: "notes" (in order), "group", "subgroup", "bass" and "root".
func chordDiff(a, b hindemith.Chord) []string {
	var diff []string
	if !sameNotes(a.Notes(), b.Notes()) {
		diff = append(diff, "notes")
	}
	if a.Group() != b.Group() {
		diff = append(diff, "group")
	}
	if a.Subgroup() != b.Subgroup() {
		diff = append(diff, "subgroup")
	}
	if !a.Bass().Equal(b.Bass()) {
		diff = append(diff, "bass")
	}
	if !sameRoot(a.HasRoot(), a.Root(), b.HasRoot(), b.Root()) {
		diff = append(diff, "root")
	}
	return diff
}

// TestEqualComparesTheClassification covers the criteria of Dyad.Equal and
// Chord.Equal (ХН-П13) other than the notes. The same notes are analyzed in
// two admitted twelve-tone systems in which exactly one criterion differs:
// for a dyad the Series II position or the root, for a chord the group, the
// subgroup, the bass (with and without a root) or the root. Each case first
// requires that exactly the named criterion differs and every other one
// agrees, then expects inequality in both directions. The positive controls
// classify the same notes identically in two systems and are equal.
//
// The expected classifications follow from the class (p(l) + a) mod 12 of
// each name and the position Octave·12 + class (package system), Series II
// and the interval roots, and the group table (package documentation).
func TestEqualComparesTheClassification(t *testing.T) {
	t.Parallel()

	t.Run("dyads", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name     string
			from, to note.Note
			sa, sb   *system.System
			series   int // the Series II position of a
			differs  []string
			want     bool
		}{
			// C4 0, D4 2: major second (9), root D4; in d1 D4 is 1: minor
			// second (11), root D4.
			{"only the Series II position", n(spelling.C(), 4), n(spelling.D(), 4), trad(), sysD1, 9, []string{"series"}, false},
			// Traditional: C𝄫3 46, B3 47, an ascending minor second with
			// the root on the upper B3. b9: C𝄫3 46, B3 45, a descending
			// minor second with the root on the upper C𝄫3.
			{"only the root", n(spelling.CDoubleFlat(), 3), n(spelling.B(), 3), trad(), sysB9, 11, []string{"root"}, false},
			// D sits at 2 in both: the same major second with the root D4.
			{"same classification of a second", n(spelling.C(), 4), n(spelling.D(), 4), trad(), sysB9, 9, nil, true},
			// F♯ sits at 6 in both: the same tritone without a root.
			{"same classification of a tritone", n(spelling.C(), 4), n(spelling.FSharp(), 4), trad(), sysB9, 13, nil, true},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				a, err := hindemith.AnalyzeDyad(tc.sa, tc.from, tc.to)
				require.NoError(t, err)
				b, err := hindemith.AnalyzeDyad(tc.sb, tc.from, tc.to)
				require.NoError(t, err)
				require.Equal(t, tc.series, a.Series(), "Series II position of a")
				require.Equal(t, tc.differs, dyadDiff(a, b),
					"differing criteria: a is %s %d root %v, b is %s %d root %v",
					a.Name(), a.Series(), a.Root(), b.Name(), b.Series(), b.Root())

				assert.Equal(t, tc.want, a.Equal(b), "a.Equal(b)")
				assert.Equal(t, tc.want, b.Equal(a), "b.Equal(a): symmetry")
			})
		}
	})

	t.Run("chords", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name    string
			notes   []note.Note
			sa, sb  *system.System
			group   hindemith.Group // the group of a
			differs []string
			want    bool
		}{
			// Traditional: C3 36, D3 38, A3 45; the best interval is the
			// fifth D3-A3, root D3 above the bass: III.2. d1: D3 37; the
			// best interval is the minor sixth D3-A3, root A3: III.2.
			{"only the root", []note.Note{n(spelling.C(), 3), n(spelling.D(), 3), n(spelling.A(), 3)},
				trad(), sysD1, hindemith.GroupIII, []string{"root"}, false},
			// e5: E3 and F3 share position 41, the first input note F3 is
			// the bass; F3-G3 is a major second, root G3: III.2. g6: E3 40,
			// F3 41, G3 42; the best interval is the major second E3-G3,
			// root G3 above the bass E3: III.2.
			{"only the bass", []note.Note{n(spelling.F(), 3), n(spelling.E(), 3), n(spelling.G(), 3)},
				sysE5, sysG6, hindemith.GroupIII, []string{"bass"}, false},
			// d1: D𝄫3 47, E♯3 41; e5: D𝄫3 36, E♯3 42. Both a lone
			// tritone, group VI without a root, with the bass on the other
			// note.
			{"only the bass, without a root", []note.Note{n(spelling.DDoubleFlat(), 3), n(spelling.ESharp(), 3)},
				sysD1, sysE5, hindemith.GroupVI, []string{"bass"}, false},
			// b9: D3 38, B3 45, a fifth with the root D3: I.1. d1: D3 37,
			// B3 47, a minor seventh with the root D3: III.1.
			{"only the group", []note.Note{n(spelling.D(), 3), n(spelling.B(), 3)},
				sysB9, sysD1, hindemith.GroupI, []string{"group"}, false},
			// Traditional: E𝄫4 50, A♭3 44, B♯4 48; one tritone A♭3-E𝄫4,
			// a major second B♯4-E𝄫4, the best interval the major third
			// A♭3-B♯4 with the root A♭3 in the bass: II.b1. minor: E𝄫4 49,
			// A♭3 43, B♯4 59; one tritone, a minor seventh E𝄫4-B♯4 and no
			// major second, the root A♭3 in the bass: II.a.
			{"only the subgroup", []note.Note{n(spelling.EDoubleFlat(), 4), n(spelling.AFlat(), 3), n(spelling.BSharp(), 4)},
				trad(), sysMinor, hindemith.GroupII, []string{"subgroup"}, false},
			// C, E and G sit at 0, 4 and 7 in both: I.1 with the root C4.
			{"same classification of a major triad", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)},
				trad(), sysB9, hindemith.GroupI, nil, true},
			// F♯ sits at 6 in both: group VI with the bass C4.
			{"same classification of a lone tritone", []note.Note{n(spelling.C(), 4), n(spelling.FSharp(), 4)},
				trad(), sysB9, hindemith.GroupVI, nil, true},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				a, err := hindemith.AnalyzeChord(tc.sa, tc.notes)
				require.NoError(t, err)
				b, err := hindemith.AnalyzeChord(tc.sb, tc.notes)
				require.NoError(t, err)
				require.Equal(t, tc.group, a.Group(), "group of a")
				require.Equal(t, tc.differs, chordDiff(a, b),
					"differing criteria: a is %v.%v bass %v root %v, b is %v.%v bass %v root %v",
					a.Group(), a.Subgroup(), a.Bass(), a.Root(), b.Group(), b.Subgroup(), b.Bass(), b.Root())

				assert.Equal(t, tc.want, a.Equal(b), "a.Equal(b)")
				assert.Equal(t, tc.want, b.Equal(a), "b.Equal(a): symmetry")
			})
		}
	})
}
