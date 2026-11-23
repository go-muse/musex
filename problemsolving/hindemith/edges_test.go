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
