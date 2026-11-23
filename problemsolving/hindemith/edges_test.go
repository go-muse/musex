package hindemith_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/problemsolving/hindemith"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnumStrings covers the text of every group, subgroup and direction.
func TestEnumStrings(t *testing.T) {
	t.Parallel()

	groups := []struct {
		g    hindemith.Group
		want string
	}{
		{hindemith.GroupI, "I"}, {hindemith.GroupII, "II"}, {hindemith.GroupIII, "III"},
		{hindemith.GroupIV, "IV"}, {hindemith.GroupV, "V"}, {hindemith.GroupVI, "VI"},
		{hindemith.Group(0), "Group(unset)"}, {hindemith.Group(99), "Group(unset)"},
	}
	for _, tc := range groups {
		assert.Equal(t, tc.want, tc.g.String())
	}

	subs := []struct {
		s    hindemith.Subgroup
		want string
	}{
		{hindemith.SubNone, ""}, {hindemith.Sub1, "1"}, {hindemith.Sub2, "2"},
		{hindemith.SubA, "a"}, {hindemith.SubB1, "b1"}, {hindemith.SubB2, "b2"},
		{hindemith.SubB3, "b3"}, {hindemith.Subgroup(99), ""},
	}
	for _, tc := range subs {
		assert.Equal(t, tc.want, tc.s.String())
	}

	assert.Equal(t, "level", hindemith.Level.String())
	assert.Equal(t, "rise", hindemith.Rise.String())
	assert.Equal(t, "fall", hindemith.Fall.String())
	assert.Equal(t, "level", hindemith.Direction(99).String())
}

// TestUnsetAccessors covers the zero values.
func TestUnsetAccessors(t *testing.T) {
	t.Parallel()

	var d hindemith.Dyad
	assert.False(t, d.IsValid())
	assert.Equal(t, 0, d.Series())
	assert.Empty(t, d.Name())
	assert.False(t, d.HasRoot())
	assert.False(t, d.IsTritone())
	assert.False(t, d.Root().IsValid())
	assert.False(t, d.From().IsValid())
	assert.False(t, d.To().IsValid())

	var c hindemith.Chord
	assert.False(t, c.IsValid())
	assert.Equal(t, 0, c.Value())
	assert.Empty(t, c.Notes())
	assert.Empty(t, c.Dyads())
	assert.False(t, c.Best().IsValid())
	assert.False(t, c.HasRoot())
	assert.False(t, c.Bass().IsValid())

	var p hindemith.Progression
	assert.Empty(t, p.Chords())
	assert.Empty(t, p.Steps())
}

// TestInvertedSubgroups covers group III and IV with the root above the
// bass.
func TestInvertedSubgroups(t *testing.T) {
	t.Parallel()

	c := hindemith.MustAnalyzeChord(trad(), []note.Note{
		n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.C(), 5), n(spelling.D(), 5),
	})
	assert.Equal(t, hindemith.GroupIII, c.Group())
	assert.Equal(t, hindemith.Sub2, c.Subgroup())
	assert.Equal(t, 8, c.Value())
	assert.True(t, c.Root().Equal(n(spelling.G(), 4)))

	c = hindemith.MustAnalyzeChord(trad(), []note.Note{
		n(spelling.DFlat(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.C(), 5),
	})
	assert.Equal(t, hindemith.GroupIV, c.Group())
	assert.Equal(t, hindemith.Sub2, c.Subgroup())
	assert.Equal(t, 10, c.Value())
	assert.True(t, c.Root().Equal(n(spelling.C(), 5)))
}

// TestDistanceOverflow covers the out-of-range path at the edges of int.
func TestDistanceOverflow(t *testing.T) {
	t.Parallel()

	hi := note.MustFromASPN(spelling.C(), math.MaxInt, cents.Zero(), nil)
	lo := note.MustFromASPN(spelling.C(), math.MinInt, cents.Zero(), nil)

	_, err := hindemith.AnalyzeDyad(trad(), hi, lo)
	assert.ErrorIs(t, err, hindemith.ErrOutOfRange)
	assert.ErrorIs(t, err, system.ErrOutOfRange)

	_, err = hindemith.AnalyzeChord(trad(), []note.Note{hi, lo})
	assert.ErrorIs(t, err, hindemith.ErrOutOfRange)
	assert.ErrorIs(t, err, system.ErrOutOfRange)

	assert.Panics(t, func() { hindemith.MustAnalyzeDyad(trad(), hi, lo) })
}

// edgeOctave is the ASPN octave of the C whose distance from G sharp 4 is
// exactly math.MinInt steps and from G4 math.MinInt+1 steps in the
// traditional system: 12 * (edgeOctave - 4) - 8 == math.MinInt.
const edgeOctave = 4 - 768614336404564650

// panicValue runs f and returns the value it panics with, or nil.
func panicValue(f func()) (v any) {
	defer func() { v = recover() }()
	f()
	return nil
}

// TestDistanceEdges covers the distances math.MinInt, math.MinInt+1 and
// math.MaxInt between two notes. Exactly math.MinInt steps (and its
// negation) is out of range: an error matching ErrOutOfRange and
// system.ErrOutOfRange, never a runtime panic; the Must variants panic with
// that error. The neighbouring distances fit and give the perfect fifth
// C-G with the root on the lower C.
func TestDistanceEdges(t *testing.T) {
	t.Parallel()

	low := note.MustFromASPN(spelling.C(), edgeOctave, cents.Zero(), nil)
	gSharp := n(spelling.GSharp(), 4)
	g := n(spelling.G(), 4)

	outOfRange := []struct {
		name     string
		from, to note.Note
	}{
		{"distance math.MinInt", gSharp, low},
		{"distance -math.MinInt", low, gSharp},
	}
	for _, tc := range outOfRange {
		t.Run("dyad "+tc.name, func(t *testing.T) {
			t.Parallel()
			var err error
			if !assert.NotPanics(t, func() { _, err = hindemith.AnalyzeDyad(trad(), tc.from, tc.to) }) {
				return
			}
			assert.ErrorIs(t, err, hindemith.ErrOutOfRange)
			assert.ErrorIs(t, err, system.ErrOutOfRange)
			v := panicValue(func() { hindemith.MustAnalyzeDyad(trad(), tc.from, tc.to) })
			perr, ok := v.(error)
			if assert.True(t, ok, "MustAnalyzeDyad panics with %v, want an error", v) {
				assert.ErrorIs(t, perr, hindemith.ErrOutOfRange)
				assert.ErrorIs(t, perr, system.ErrOutOfRange)
			}
		})
		t.Run("chord "+tc.name, func(t *testing.T) {
			t.Parallel()
			ns := []note.Note{tc.from, tc.to}
			var err error
			if !assert.NotPanics(t, func() { _, err = hindemith.AnalyzeChord(trad(), ns) }) {
				return
			}
			assert.ErrorIs(t, err, hindemith.ErrOutOfRange)
			assert.ErrorIs(t, err, system.ErrOutOfRange)
			v := panicValue(func() { hindemith.MustAnalyzeChord(trad(), ns) })
			perr, ok := v.(error)
			if assert.True(t, ok, "MustAnalyzeChord panics with %v, want an error", v) {
				assert.ErrorIs(t, perr, hindemith.ErrOutOfRange)
				assert.ErrorIs(t, perr, system.ErrOutOfRange)
			}
		})
	}

	inRange := []struct {
		name     string
		from, to note.Note
	}{
		{"distance math.MinInt+1", g, low},
		{"distance math.MaxInt", low, g},
	}
	for _, tc := range inRange {
		t.Run("dyad "+tc.name, func(t *testing.T) {
			t.Parallel()
			var (
				d   hindemith.Dyad
				err error
			)
			if !assert.NotPanics(t, func() { d, err = hindemith.AnalyzeDyad(trad(), tc.from, tc.to) }) {
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 3, d.Series())
			assert.Equal(t, "perfect fifth", d.Name())
			assert.True(t, d.Root().Equal(low), "root %v", d.Root())
			assert.True(t, d.From().Equal(tc.from))
			assert.True(t, d.To().Equal(tc.to))
		})
		t.Run("chord "+tc.name, func(t *testing.T) {
			t.Parallel()
			var (
				c   hindemith.Chord
				err error
			)
			if !assert.NotPanics(t, func() { c, err = hindemith.AnalyzeChord(trad(), []note.Note{tc.from, tc.to}) }) {
				return
			}
			require.NoError(t, err)
			assert.Equal(t, hindemith.GroupI, c.Group())
			assert.Equal(t, hindemith.Sub1, c.Subgroup())
			assert.True(t, c.Bass().Equal(low), "bass %v", c.Bass())
			assert.True(t, c.Root().Equal(low), "root %v", c.Root())
		})
	}

}

// TestChordEqualByNotes covers Chord.Equal over the notes and their order:
// equal classifications of different note lists are not equal chords.
func TestChordEqualByNotes(t *testing.T) {
	t.Parallel()

	base := []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)}
	cases := []struct {
		name  string
		other []note.Note
		want  bool
	}{
		{"same notes", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)}, true},
		{"one note in another octave", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 5)}, false},
		{"one note respelled at the same position", []note.Note{n(spelling.C(), 4), n(spelling.FFlat(), 4), n(spelling.G(), 4)}, false},
		{"same notes in another order", []note.Note{n(spelling.C(), 4), n(spelling.G(), 4), n(spelling.E(), 4)}, false},
		{"an added doubling", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.C(), 5)}, false},
		{"another class", []note.Note{n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.C(), 5)}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := hindemith.MustAnalyzeChord(trad(), base)
			b := hindemith.MustAnalyzeChord(trad(), tc.other)
			assert.Equal(t, tc.want, a.Equal(b))
			assert.Equal(t, tc.want, b.Equal(a), "symmetry")
			assert.True(t, a.Equal(a), "reflexivity")
			assert.False(t, a.Equal(hindemith.Chord{}), "set and unset")
			assert.False(t, hindemith.Chord{}.Equal(a), "unset and set")
		})
	}
}
