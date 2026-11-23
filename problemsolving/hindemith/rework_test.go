package hindemith_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/problemsolving/hindemith"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// restricted admits only the alterations -1, 0 and +1.
func restricted() *system.System {
	return system.MustNew(system.Definition{
		Name:        "restricted",
		Classes:     12,
		Boundary:    letter.C,
		Offsets:     [letter.Count]int{0, 2, 4, 5, 7, 9, 11},
		Chromatic:   1,
		Alterations: system.MustIntegerRange(-1, 1),
	})
}

// classOf returns the group, subgroup and the spellings of root and bass.
type classOf struct {
	group hindemith.Group
	sub   hindemith.Subgroup
	root  spelling.Spelling
	bass  spelling.Spelling
}

func classifyOf(t *testing.T, ns []note.Note) classOf {
	t.Helper()
	c, err := hindemith.AnalyzeChord(trad(), ns)
	require.NoError(t, err)
	out := classOf{group: c.Group(), sub: c.Subgroup(), bass: c.Bass().Spelling()}
	if c.HasRoot() {
		out.root = c.Root().Spelling()
	}
	return out
}

// TestDoublingsKeepTheClass covers triads and seventh chords with each tone
// doubled, in every inversion: a doubling carries no root information.
func TestDoublingsKeepTheClass(t *testing.T) {
	t.Parallel()

	triadInversions := [][]note.Note{
		{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)},
		{n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.C(), 5)},
		{n(spelling.G(), 4), n(spelling.C(), 5), n(spelling.E(), 5)},
	}
	for inv, ns := range triadInversions {
		want := classifyOf(t, ns)
		for tone := range ns {
			for _, oct := range []int{-1, 1} {
				doubled := append(append([]note.Note(nil), ns...),
					note.MustFromASPN(ns[tone].Spelling(), ns[tone].MustOctave().MustASPN()+oct, cents.Zero(), nil))
				got := classifyOf(t, doubled)
				// The group and the root name survive any doubling. A
				// doubling above the lowest occurrence of its class changes
				// nothing; one below moves that occurrence, so the bass and
				// hence the subgroup follow the new lowest voicing.
				assert.Equal(t, want.group, got.group, "inversion %d tone %d octave %+d group", inv, tone, oct)
				assert.Equal(t, want.root, got.root, "inversion %d tone %d octave %+d root", inv, tone, oct)
			}
		}
	}

	seventhInversions := [][]note.Note{
		{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BFlat(), 4)},
		{n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BFlat(), 4), n(spelling.C(), 5)},
		{n(spelling.G(), 4), n(spelling.BFlat(), 4), n(spelling.C(), 5), n(spelling.E(), 5)},
		{n(spelling.BFlat(), 4), n(spelling.C(), 5), n(spelling.E(), 5), n(spelling.G(), 5)},
	}
	for inv, ns := range seventhInversions {
		want := classifyOf(t, ns)
		for tone := range ns {
			// A doubling never decides the root and adds no interval to
			// the classification: a second or seventh between the doubled
			// note and its neighbours is not counted. A doubling below the
			// lowest occurrence of its class moves the bass, so only the
			// group and the root name are claimed here.
			for _, oct := range []int{-1, 1} {
				doubled := append(append([]note.Note(nil), ns...),
					note.MustFromASPN(ns[tone].Spelling(), ns[tone].MustOctave().MustASPN()+oct, cents.Zero(), nil))
				got := classifyOf(t, doubled)
				assert.Equal(t, want.group, got.group, "inversion %d tone %d octave %+d group", inv, tone, oct)
				assert.Equal(t, want.root, got.root, "inversion %d tone %d octave %+d root", inv, tone, oct)
			}
		}
	}

	t.Run("root doubled above the seventh adds no major second", func(t *testing.T) {
		t.Parallel()
		// The major second B flat 4 - C5 comes from the doubling and is
		// not counted, so the chord stays II.a like the undoubled seventh.
		got := classifyOf(t, []note.Note{
			n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BFlat(), 4), n(spelling.C(), 5),
		})
		assert.Equal(t, classOf{hindemith.GroupII, hindemith.SubA, spelling.C(), spelling.C()}, got)
	})

	t.Run("reviewer voicings", func(t *testing.T) {
		t.Parallel()
		got := classifyOf(t, []note.Note{n(spelling.E(), 3), n(spelling.G(), 3), n(spelling.C(), 4), n(spelling.E(), 4)})
		assert.Equal(t, classOf{hindemith.GroupI, hindemith.Sub2, spelling.C(), spelling.E()}, got)
		got = classifyOf(t, []note.Note{n(spelling.G(), 3), n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)})
		assert.Equal(t, classOf{hindemith.GroupI, hindemith.Sub2, spelling.C(), spelling.G()}, got)
		got = classifyOf(t, []note.Note{n(spelling.C(), 3), n(spelling.E(), 3), n(spelling.BFlat(), 3), n(spelling.E(), 4)})
		assert.Equal(t, classOf{hindemith.GroupII, hindemith.SubA, spelling.C(), spelling.C()}, got)
	})
}

// TestPermutationsGiveOneResult covers the independence of the listing
// order: ties between equal best intervals are broken by pitch.
func TestPermutationsGiveOneResult(t *testing.T) {
	t.Parallel()

	permute := func(ns []note.Note, visit func([]note.Note)) {
		var rec func([]note.Note, []note.Note)
		rec = func(rest, acc []note.Note) {
			if len(rest) == 0 {
				visit(acc)
				return
			}
			for i := range rest {
				next := append(append([]note.Note(nil), rest[:i]...), rest[i+1:]...)
				rec(next, append(append([]note.Note(nil), acc...), rest[i]))
			}
		}
		rec(ns, nil)
	}

	t.Run("fifths tie by the lower pair", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.D(), 5)}
		want := classOf{hindemith.GroupIII, hindemith.Sub1, spelling.C(), spelling.C()}
		permute(ns, func(p []note.Note) {
			assert.Equal(t, want, classifyOf(t, p), "order %v", p)
		})
	})

	t.Run("quartal triad is V in every order", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 3), n(spelling.G(), 3), n(spelling.D(), 4)}
		permute(ns, func(p []note.Note) {
			c, err := hindemith.AnalyzeChord(trad(), p)
			require.NoError(t, err)
			assert.Equal(t, hindemith.GroupV, c.Group(), "order %v", p)
			assert.False(t, c.HasRoot())
		})
	})

	t.Run("seventh chord", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BFlat(), 4)}
		want := classOf{hindemith.GroupII, hindemith.SubA, spelling.C(), spelling.C()}
		permute(ns, func(p []note.Note) {
			assert.Equal(t, want, classifyOf(t, p), "order %v", p)
		})
	})
}

// TestGroupVShapes covers the augmented triad and the perfect quartal
// triad, and the shapes that are not group V.
func TestGroupVShapes(t *testing.T) {
	t.Parallel()

	t.Run("quartal triad in every inversion and with doublings", func(t *testing.T) {
		t.Parallel()
		voicings := [][]note.Note{
			{n(spelling.C(), 4), n(spelling.F(), 4), n(spelling.BFlat(), 4)},
			{n(spelling.F(), 4), n(spelling.BFlat(), 4), n(spelling.C(), 5)},
			{n(spelling.BFlat(), 4), n(spelling.C(), 5), n(spelling.F(), 5)},
			{n(spelling.C(), 4), n(spelling.F(), 4), n(spelling.BFlat(), 4), n(spelling.C(), 5)},
		}
		for _, ns := range voicings {
			c, err := hindemith.AnalyzeChord(trad(), ns)
			require.NoError(t, err)
			assert.Equal(t, hindemith.GroupV, c.Group(), "voicing %v", ns)
			assert.Equal(t, hindemith.SubNone, c.Subgroup())
			assert.False(t, c.HasRoot())
		}
	})

	t.Run("a single third is not indeterminate", func(t *testing.T) {
		t.Parallel()
		got := classifyOf(t, []note.Note{n(spelling.C(), 4), n(spelling.E(), 4)})
		assert.Equal(t, classOf{hindemith.GroupI, hindemith.Sub1, spelling.C(), spelling.C()}, got)
		got = classifyOf(t, []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.C(), 5)})
		assert.Equal(t, classOf{hindemith.GroupI, hindemith.Sub1, spelling.C(), spelling.C()}, got)
		got = classifyOf(t, []note.Note{n(spelling.E(), 4), n(spelling.C(), 5)})
		assert.Equal(t, classOf{hindemith.GroupI, hindemith.Sub2, spelling.C(), spelling.E()}, got)
		d, err := hindemith.AnalyzeDyad(trad(), n(spelling.C(), 4), n(spelling.E(), 4))
		require.NoError(t, err)
		assert.True(t, d.Root().Equal(n(spelling.C(), 4)))
	})

	t.Run("augmented triad with doublings stays V", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.GSharp(), 4), n(spelling.E(), 5)}
		c := hindemith.MustAnalyzeChord(trad(), ns)
		assert.Equal(t, hindemith.GroupV, c.Group())
		assert.False(t, c.HasRoot())
	})
}

// TestDoubledTritoneMember counts different tritones, not pairs of notes.
func TestDoubledTritoneMember(t *testing.T) {
	t.Parallel()

	got := classifyOf(t, []note.Note{n(spelling.C(), 3), n(spelling.E(), 3), n(spelling.BFlat(), 3)})
	assert.Equal(t, classOf{hindemith.GroupII, hindemith.SubA, spelling.C(), spelling.C()}, got)
	got = classifyOf(t, []note.Note{n(spelling.C(), 3), n(spelling.E(), 3), n(spelling.BFlat(), 3), n(spelling.E(), 4)})
	assert.Equal(t, classOf{hindemith.GroupII, hindemith.SubA, spelling.C(), spelling.C()}, got)
}

// TestInadmissibleNames covers the distinguishable inadmissible failure.
func TestInadmissibleNames(t *testing.T) {
	t.Parallel()

	x := n(spelling.CDoubleSharp(), 4)
	_, err := hindemith.AnalyzeDyad(restricted(), n(spelling.C(), 4), x)
	assert.ErrorIs(t, err, hindemith.ErrInadmissible)
	assert.ErrorIs(t, err, system.ErrInadmissible)
	_, err = hindemith.AnalyzeChord(restricted(), []note.Note{n(spelling.C(), 4), x})
	assert.ErrorIs(t, err, hindemith.ErrInadmissible)
	assert.ErrorIs(t, err, system.ErrInadmissible)
	assert.Panics(t, func() { hindemith.MustAnalyzeChord(restricted(), []note.Note{n(spelling.C(), 4), x}) })
}

// TestEqual covers the Equal methods of Dyad and Chord.
func TestEqual(t *testing.T) {
	t.Parallel()

	a := hindemith.MustAnalyzeDyad(trad(), n(spelling.C(), 4), n(spelling.G(), 4))
	b := hindemith.MustAnalyzeDyad(trad(), n(spelling.C(), 4), n(spelling.G(), 4))
	c := hindemith.MustAnalyzeDyad(trad(), n(spelling.C(), 4), n(spelling.F(), 4))
	assert.True(t, a.Equal(b))
	assert.False(t, a.Equal(c))
	assert.False(t, hindemith.Dyad{}.Equal(hindemith.Dyad{}), "unset values are not equal")

	ns := []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)}
	ch1 := hindemith.MustAnalyzeChord(trad(), ns)
	ch2 := hindemith.MustAnalyzeChord(trad(), ns)
	ch3 := hindemith.MustAnalyzeChord(trad(), []note.Note{n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.C(), 5)})
	assert.True(t, ch1.Equal(ch2))
	assert.False(t, ch1.Equal(ch3))
	assert.False(t, hindemith.Chord{}.Equal(hindemith.Chord{}), "unset values are not equal")
}
