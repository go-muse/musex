package hindemith_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/problemsolving/hindemith"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests of this file pin the doubling rule: a chord is classified as if
// it were reduced to the lowest note of every pitch class, so a repetition
// of a present tone in unison or one or more octaves above its lowest
// occurrence adds no interval to the classification. A pitch class is a
// position in the system modulo 12, so a doubling may be spelled
// differently; of two notes at one position the first input note stays.
// The doubled notes stay only in the evidence: Notes and Dyads.

// spelled formats notes for failure messages.
func spelled(ns []note.Note) string {
	std := settings.Standard()
	parts := make([]string, len(ns))
	for i, x := range ns {
		parts[i] = x.MustFormatIn(std)
	}
	return strings.Join(parts, " ")
}

// shifted returns the note with the same name the given number of octaves
// away; zero gives a unison copy.
func shifted(x note.Note, octaves int) note.Note {
	return note.MustFromASPN(x.Spelling(), x.MustOctave().MustASPN()+octaves, cents.Zero(), nil)
}

// withNote returns a copy of ns with x appended.
func withNote(ns []note.Note, x note.Note) []note.Note {
	return append(append([]note.Note(nil), ns...), x)
}

// analyzed analyzes ns in the traditional system and requires success.
func analyzed(t *testing.T, ns []note.Note) hindemith.Chord {
	t.Helper()
	c, err := hindemith.AnalyzeChord(trad(), ns)
	require.NoError(t, err, "chord %s", spelled(ns))
	require.True(t, c.IsValid(), "chord %s", spelled(ns))
	return c
}

// assertSameClass asserts that got has the group, subgroup, fluctuation
// value, bass and root of want.
func assertSameClass(t *testing.T, want, got hindemith.Chord, context string) {
	t.Helper()
	assert.Equal(t, want.Group(), got.Group(), "%s: group", context)
	assert.Equal(t, want.Subgroup(), got.Subgroup(), "%s: subgroup", context)
	assert.Equal(t, want.Value(), got.Value(), "%s: value", context)
	assert.True(t, got.Bass().Equal(want.Bass()), "%s: bass %v, want %v", context, got.Bass(), want.Bass())
	if assert.Equal(t, want.HasRoot(), got.HasRoot(), "%s: has root", context) && want.HasRoot() {
		assert.True(t, got.Root().Equal(want.Root()), "%s: root %v, want %v", context, got.Root(), want.Root())
	}
}

// assertEvidence asserts that the doubled notes stay in the evidence: every
// input note in input order and one dyad per unordered pair of input notes,
// and that the best dyad carries the chord root.
func assertEvidence(t *testing.T, c hindemith.Chord, ns []note.Note, context string) {
	t.Helper()
	got := c.Notes()
	if assert.Len(t, got, len(ns), "%s: notes", context) {
		for i := range ns {
			assert.True(t, got[i].Equal(ns[i]), "%s: note %d is %v, want %v", context, i, got[i], ns[i])
		}
	}
	assert.Len(t, c.Dyads(), len(ns)*(len(ns)-1)/2, "%s: dyads", context)
	if c.HasRoot() {
		if assert.True(t, c.Best().IsValid(), "%s: best", context) {
			assert.True(t, c.Best().Root().Equal(c.Root()), "%s: best root %v, chord root %v", context, c.Best().Root(), c.Root())
		}
	} else {
		assert.False(t, c.Best().IsValid(), "%s: best without a root", context)
		assert.False(t, c.Root().IsValid(), "%s: root without a root", context)
	}
}

// permutations calls visit with every ordering of ns.
func permutations(ns []note.Note, visit func([]note.Note)) {
	var rec func(rest, acc []note.Note)
	rec = func(rest, acc []note.Note) {
		if len(rest) == 0 {
			visit(acc)
			return
		}
		for i := range rest {
			next := append(append([]note.Note(nil), rest[:i]...), rest[i+1:]...)
			rec(next, withNote(acc, rest[i]))
		}
	}
	rec(ns, nil)
}

// doublingBases are undoubled chords of every kind the doubling rule must
// hold for: group VI (diminished triad, diminished seventh, bare tritone),
// group V (augmented and perfect quartal triads), triads and seventh chords
// in every inversion, and chords whose best interval has its root on top
// (seconds, fourths, sixths).
func doublingBases() []struct {
	name  string
	notes []note.Note
} {
	return []struct {
		name  string
		notes []note.Note
	}{
		{"major triad root position", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)}},
		{"major triad first inversion", []note.Note{n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.C(), 5)}},
		{"major triad second inversion", []note.Note{n(spelling.G(), 4), n(spelling.C(), 5), n(spelling.E(), 5)}},
		{"minor triad root position", []note.Note{n(spelling.A(), 3), n(spelling.C(), 4), n(spelling.E(), 4)}},
		{"minor triad first inversion", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.A(), 4)}},
		{"minor triad second inversion", []note.Note{n(spelling.E(), 4), n(spelling.A(), 4), n(spelling.C(), 5)}},
		{"dominant seventh root position", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BFlat(), 4)}},
		{"dominant seventh first inversion", []note.Note{n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BFlat(), 4), n(spelling.C(), 5)}},
		{"dominant seventh second inversion", []note.Note{n(spelling.G(), 4), n(spelling.BFlat(), 4), n(spelling.C(), 5), n(spelling.E(), 5)}},
		{"dominant seventh third inversion", []note.Note{n(spelling.BFlat(), 4), n(spelling.C(), 5), n(spelling.E(), 5), n(spelling.G(), 5)}},
		{"dominant seventh without fifth", []note.Note{n(spelling.C(), 3), n(spelling.E(), 3), n(spelling.BFlat(), 3)}},
		{"major seventh root position", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.B(), 4)}},
		{"major seventh first inversion", []note.Note{n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.B(), 4), n(spelling.C(), 5)}},
		{"major seventh second inversion", []note.Note{n(spelling.G(), 4), n(spelling.B(), 4), n(spelling.C(), 5), n(spelling.E(), 5)}},
		{"major seventh third inversion", []note.Note{n(spelling.B(), 4), n(spelling.C(), 5), n(spelling.E(), 5), n(spelling.G(), 5)}},
		{"minor seventh root position", []note.Note{n(spelling.A(), 3), n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)}},
		{"minor seventh first inversion", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.A(), 4)}},
		{"minor seventh second inversion", []note.Note{n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.A(), 4), n(spelling.C(), 5)}},
		{"minor seventh third inversion", []note.Note{n(spelling.G(), 4), n(spelling.A(), 4), n(spelling.C(), 5), n(spelling.E(), 5)}},
		{"half-diminished seventh root position", []note.Note{n(spelling.B(), 3), n(spelling.D(), 4), n(spelling.F(), 4), n(spelling.A(), 4)}},
		{"half-diminished seventh third inversion", []note.Note{n(spelling.A(), 3), n(spelling.B(), 3), n(spelling.D(), 4), n(spelling.F(), 4)}},
		{"diminished triad root position", []note.Note{n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4)}},
		{"diminished triad first inversion", []note.Note{n(spelling.EFlat(), 4), n(spelling.GFlat(), 4), n(spelling.C(), 5)}},
		{"diminished triad second inversion", []note.Note{n(spelling.GFlat(), 4), n(spelling.C(), 5), n(spelling.EFlat(), 5)}},
		{"diminished seventh root position", []note.Note{n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4), n(spelling.A(), 4)}},
		{"diminished seventh first inversion", []note.Note{n(spelling.EFlat(), 4), n(spelling.GFlat(), 4), n(spelling.A(), 4), n(spelling.C(), 5)}},
		{"diminished seventh second inversion", []note.Note{n(spelling.GFlat(), 4), n(spelling.A(), 4), n(spelling.C(), 5), n(spelling.EFlat(), 5)}},
		{"diminished seventh third inversion", []note.Note{n(spelling.A(), 4), n(spelling.C(), 5), n(spelling.EFlat(), 5), n(spelling.GFlat(), 5)}},
		{"tritone C-F sharp", []note.Note{n(spelling.C(), 4), n(spelling.FSharp(), 4)}},
		{"tritone F sharp-C", []note.Note{n(spelling.FSharp(), 4), n(spelling.C(), 5)}},
		{"augmented triad root position", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.GSharp(), 4)}},
		{"augmented triad first inversion", []note.Note{n(spelling.E(), 4), n(spelling.GSharp(), 4), n(spelling.C(), 5)}},
		{"augmented triad second inversion", []note.Note{n(spelling.GSharp(), 4), n(spelling.C(), 5), n(spelling.E(), 5)}},
		{"quartal triad root position", []note.Note{n(spelling.C(), 4), n(spelling.F(), 4), n(spelling.BFlat(), 4)}},
		{"quartal triad first inversion", []note.Note{n(spelling.F(), 4), n(spelling.BFlat(), 4), n(spelling.C(), 5)}},
		{"quartal triad second inversion", []note.Note{n(spelling.BFlat(), 4), n(spelling.C(), 5), n(spelling.F(), 5)}},
		{"major second root on top", []note.Note{n(spelling.C(), 4), n(spelling.D(), 4)}},
		{"minor second root on top", []note.Note{n(spelling.C(), 4), n(spelling.DFlat(), 4)}},
		{"minor seventh root at the bottom", []note.Note{n(spelling.D(), 3), n(spelling.C(), 4)}},
		{"perfect fourth root on top", []note.Note{n(spelling.C(), 4), n(spelling.F(), 4)}},
		{"major sixth root on top", []note.Note{n(spelling.C(), 4), n(spelling.A(), 4)}},
		{"minor sixth root on top", []note.Note{n(spelling.E(), 4), n(spelling.C(), 5)}},
		{"second under a fourth", []note.Note{n(spelling.C(), 4), n(spelling.D(), 4), n(spelling.F(), 4)}},
		{"minor second under a fourth", []note.Note{n(spelling.B(), 3), n(spelling.C(), 4), n(spelling.E(), 4)}},
		{"inverted triad with an added second", []note.Note{n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.C(), 5), n(spelling.D(), 5)}},
	}
}

// TestDoublingAboveKeepsTheClass covers every tone of every base chord
// doubled in unison, one octave and two octaves above: the doubling never
// changes the group, subgroup, fluctuation value, bass or root.
func TestDoublingAboveKeepsTheClass(t *testing.T) {
	t.Parallel()

	for _, base := range doublingBases() {
		for tone := range base.notes {
			for _, oct := range []int{0, 1, 2} {
				doubled := withNote(base.notes, shifted(base.notes[tone], oct))
				name := fmt.Sprintf("%s as %s", base.name, spelled(doubled))
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					want := analyzed(t, base.notes)
					got := analyzed(t, doubled)
					assertSameClass(t, want, got, spelled(doubled))
					assertEvidence(t, got, doubled, spelled(doubled))
				})
			}
		}
	}
}

// TestDoublingBelowFollowsTheLowestOccurrence covers every tone of every
// base chord doubled one and two octaves below. Such a doubling moves the
// lowest occurrence of its class, so the class is not claimed to stay;
// the chord is classified as its reduction, the base chord with the doubled
// tone moved down to the doubling.
func TestDoublingBelowFollowsTheLowestOccurrence(t *testing.T) {
	t.Parallel()

	for _, base := range doublingBases() {
		for tone := range base.notes {
			for _, oct := range []int{-1, -2} {
				lower := shifted(base.notes[tone], oct)
				doubled := withNote(base.notes, lower)
				reduced := append([]note.Note(nil), base.notes...)
				reduced[tone] = lower
				name := fmt.Sprintf("%s as %s", base.name, spelled(doubled))
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					want := analyzed(t, reduced)
					got := analyzed(t, doubled)
					context := fmt.Sprintf("%s as %s", spelled(doubled), spelled(reduced))
					assertSameClass(t, want, got, context)
					assertEvidence(t, got, doubled, context)
				})
			}
		}
	}
}

// TestDoubledChordClasses covers chords with doublings against stated
// classifications, in the given order and in every other listing order.
func TestDoubledChordClasses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		notes   []note.Note
		group   hindemith.Group
		sub     hindemith.Subgroup
		bass    note.Note
		root    note.Note
		hasRoot bool
		value   int
	}{
		{
			"diminished triad with the bass doubled an octave above",
			[]note.Note{n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4), n(spelling.C(), 5)},
			hindemith.GroupVI, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 12,
		},
		{
			"diminished triad with the top doubled two octaves above",
			[]note.Note{n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4), n(spelling.GFlat(), 6)},
			hindemith.GroupVI, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 12,
		},
		{
			"diminished seventh with the bass doubled an octave above",
			[]note.Note{n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4), n(spelling.A(), 4), n(spelling.C(), 5)},
			hindemith.GroupVI, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 12,
		},
		{
			"diminished seventh with the third doubled in unison",
			[]note.Note{n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4), n(spelling.A(), 4), n(spelling.EFlat(), 4)},
			hindemith.GroupVI, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 12,
		},
		{
			"bare tritone",
			[]note.Note{n(spelling.C(), 4), n(spelling.FSharp(), 4)},
			hindemith.GroupVI, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 12,
		},
		{
			"tritone with the bass doubled an octave above",
			[]note.Note{n(spelling.C(), 4), n(spelling.FSharp(), 4), n(spelling.C(), 5)},
			hindemith.GroupVI, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 12,
		},
		{
			"tritone with the top doubled an octave above",
			[]note.Note{n(spelling.C(), 4), n(spelling.FSharp(), 4), n(spelling.FSharp(), 5)},
			hindemith.GroupVI, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 12,
		},
		{
			"minor seventh",
			[]note.Note{n(spelling.D(), 3), n(spelling.C(), 4)},
			hindemith.GroupIII, hindemith.Sub1, n(spelling.D(), 3), n(spelling.D(), 3), true, 7,
		},
		{
			"minor seventh with the bass doubled above the seventh",
			[]note.Note{n(spelling.D(), 3), n(spelling.C(), 4), n(spelling.D(), 4)},
			hindemith.GroupIII, hindemith.Sub1, n(spelling.D(), 3), n(spelling.D(), 3), true, 7,
		},
		{
			"dominant seventh without fifth",
			[]note.Note{n(spelling.C(), 3), n(spelling.E(), 3), n(spelling.BFlat(), 3)},
			hindemith.GroupII, hindemith.SubA, n(spelling.C(), 3), n(spelling.C(), 3), true, 3,
		},
		{
			"dominant seventh without fifth, root doubled above the seventh",
			[]note.Note{n(spelling.C(), 3), n(spelling.E(), 3), n(spelling.BFlat(), 3), n(spelling.C(), 4)},
			hindemith.GroupII, hindemith.SubA, n(spelling.C(), 3), n(spelling.C(), 3), true, 3,
		},
		{
			"dominant seventh, root doubled above the seventh",
			[]note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BFlat(), 4), n(spelling.C(), 5)},
			hindemith.GroupII, hindemith.SubA, n(spelling.C(), 4), n(spelling.C(), 4), true, 3,
		},
		{
			"major seventh, root doubled above the seventh",
			[]note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.B(), 4), n(spelling.C(), 5)},
			hindemith.GroupIII, hindemith.Sub1, n(spelling.C(), 4), n(spelling.C(), 4), true, 7,
		},
		{
			"major triad with the fifth doubled in unison",
			[]note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.G(), 4)},
			hindemith.GroupI, hindemith.Sub1, n(spelling.C(), 4), n(spelling.C(), 4), true, 1,
		},
		{
			"octave",
			[]note.Note{n(spelling.C(), 4), n(spelling.C(), 5)},
			hindemith.GroupI, hindemith.Sub1, n(spelling.C(), 4), n(spelling.C(), 4), true, 1,
		},
		{
			"major second with the bass doubled above",
			[]note.Note{n(spelling.C(), 4), n(spelling.D(), 4), n(spelling.C(), 5)},
			hindemith.GroupIII, hindemith.Sub2, n(spelling.C(), 4), n(spelling.D(), 4), true, 8,
		},
		{
			"minor second with the top doubled an octave above",
			[]note.Note{n(spelling.C(), 4), n(spelling.DFlat(), 4), n(spelling.DFlat(), 5)},
			hindemith.GroupIII, hindemith.Sub2, n(spelling.C(), 4), n(spelling.DFlat(), 4), true, 8,
		},
		{
			"perfect fourth with the top doubled two octaves above",
			[]note.Note{n(spelling.C(), 4), n(spelling.F(), 4), n(spelling.F(), 6)},
			hindemith.GroupI, hindemith.Sub2, n(spelling.C(), 4), n(spelling.F(), 4), true, 2,
		},
		{
			"major sixth with the bass doubled above",
			[]note.Note{n(spelling.C(), 4), n(spelling.A(), 4), n(spelling.C(), 5)},
			hindemith.GroupI, hindemith.Sub2, n(spelling.C(), 4), n(spelling.A(), 4), true, 2,
		},
		{
			"augmented triad with the bass doubled above",
			[]note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.GSharp(), 4), n(spelling.C(), 5)},
			hindemith.GroupV, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 11,
		},
		{
			"quartal triad with the middle doubled above",
			[]note.Note{n(spelling.C(), 4), n(spelling.F(), 4), n(spelling.BFlat(), 4), n(spelling.F(), 5)},
			hindemith.GroupV, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 11,
		},
		// Enharmonic doublings: the class of a tone is its position in the
		// system modulo 12, not its spelling.
		{
			"diminished triad with the top doubled as F sharp an octave above",
			[]note.Note{n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4), n(spelling.FSharp(), 5)},
			hindemith.GroupVI, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 12,
		},
		{
			// The octave argument of n counts positions from C, so B sharp
			// with 5 is at the position of C5 and is written B♯4.
			"major triad with the root doubled as B sharp",
			[]note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BSharp(), 5)},
			hindemith.GroupI, hindemith.Sub1, n(spelling.C(), 4), n(spelling.C(), 4), true, 1,
		},
		{
			"dominant seventh without fifth, seventh doubled as A sharp",
			[]note.Note{n(spelling.C(), 3), n(spelling.E(), 3), n(spelling.BFlat(), 3), n(spelling.ASharp(), 4)},
			hindemith.GroupII, hindemith.SubA, n(spelling.C(), 3), n(spelling.C(), 3), true, 3,
		},
		{
			"minor seventh with the bass doubled as C double sharp above the seventh",
			[]note.Note{n(spelling.D(), 3), n(spelling.C(), 4), n(spelling.CDoubleSharp(), 4)},
			hindemith.GroupIII, hindemith.Sub1, n(spelling.D(), 3), n(spelling.D(), 3), true, 7,
		},
		{
			"minor seventh with the bass doubled as E double flat above the seventh",
			[]note.Note{n(spelling.D(), 3), n(spelling.C(), 4), n(spelling.EDoubleFlat(), 4)},
			hindemith.GroupIII, hindemith.Sub1, n(spelling.D(), 3), n(spelling.D(), 3), true, 7,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			check := func(ns []note.Note) {
				context := spelled(ns)
				c, err := hindemith.AnalyzeChord(trad(), ns)
				if !assert.NoError(t, err, context) {
					return
				}
				assert.Equal(t, tc.group, c.Group(), "%s: group", context)
				assert.Equal(t, tc.sub, c.Subgroup(), "%s: subgroup", context)
				assert.Equal(t, tc.value, c.Value(), "%s: value", context)
				assert.True(t, c.Bass().Equal(tc.bass), "%s: bass %v", context, c.Bass())
				if assert.Equal(t, tc.hasRoot, c.HasRoot(), "%s: has root", context) && tc.hasRoot {
					assert.True(t, c.Root().Equal(tc.root), "%s: root %v", context, c.Root())
				}
				assertEvidence(t, c, ns, context)
			}
			check(tc.notes)
			permutations(tc.notes, check)
		})
	}
}

// TestEnharmonicUnisonKeepsTheFirstNote covers a unison of two spellings of
// one position: it is one doubled tone, group I, and both the root and the
// bass are the first input note, as for equal positions in Bass.
func TestEnharmonicUnisonKeepsTheFirstNote(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		notes []note.Note
	}{
		{"C sharp then D flat", []note.Note{n(spelling.CSharp(), 4), n(spelling.DFlat(), 4)}},
		{"D flat then C sharp", []note.Note{n(spelling.DFlat(), 4), n(spelling.CSharp(), 4)}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := analyzed(t, tc.notes)
			first, second := tc.notes[0], tc.notes[1]
			assert.Equal(t, hindemith.GroupI, c.Group())
			assert.Equal(t, hindemith.Sub1, c.Subgroup())
			assert.Equal(t, 1, c.Value())
			assert.True(t, c.Bass().Equal(first), "bass %v, want %v", c.Bass(), first)
			assert.False(t, c.Bass().Equal(second), "bass %v is the second note", c.Bass())
			if assert.True(t, c.HasRoot()) {
				assert.True(t, c.Root().Equal(first), "root %v, want %v", c.Root(), first)
				assert.False(t, c.Root().Equal(second), "root %v is the second note", c.Root())
			}
			assertEvidence(t, c, tc.notes, spelled(tc.notes))
		})
	}
}

// TestDoublingsKeepTheFluctuation covers a progression step between a chord
// and the same chord with doublings above: the tension stays level.
func TestDoublingsKeepTheFluctuation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		plain   []note.Note
		doubled []note.Note
	}{
		{
			"dominant seventh",
			[]note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BFlat(), 4)},
			[]note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BFlat(), 4), n(spelling.C(), 5)},
		},
		{
			"diminished triad",
			[]note.Note{n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4)},
			[]note.Note{n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4), n(spelling.C(), 5)},
		},
		{
			"minor seventh dyad",
			[]note.Note{n(spelling.D(), 3), n(spelling.C(), 4)},
			[]note.Note{n(spelling.D(), 3), n(spelling.C(), 4), n(spelling.D(), 4)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p, err := hindemith.AnalyzeProgression(trad(), [][]note.Note{tc.plain, tc.doubled, tc.plain})
			require.NoError(t, err)
			steps := p.Steps()
			require.Len(t, steps, 2)
			for i, s := range steps {
				assert.Equal(t, hindemith.Level, s.Direction(), "step %d", i)
				assert.Equal(t, 0, s.Delta(), "step %d", i)
			}
		})
	}
}
