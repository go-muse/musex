package hindemith_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/problemsolving/hindemith"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// n returns the unbound note with the name and the ASPN octave.
func n(name spelling.Spelling, aspn int) note.Note {
	return note.MustFromASPN(name, aspn, cents.Zero(), nil)
}

func trad() *system.System { return system.Traditional() }

// TestAnalyzeDyad covers Series II positions and interval roots, including
// the rootless tritone and compound intervals.
func TestAnalyzeDyad(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		from     note.Note
		to       note.Note
		series   int
		dyadName string
		root     note.Note
		hasRoot  bool
	}{
		{"unison", n(spelling.C(), 4), n(spelling.C(), 4), 1, "perfect unison", n(spelling.C(), 4), true},
		{"octave", n(spelling.C(), 4), n(spelling.C(), 5), 2, "octave", n(spelling.C(), 4), true},
		{"fifth root low", n(spelling.C(), 4), n(spelling.G(), 4), 3, "perfect fifth", n(spelling.C(), 4), true},
		{"fourth root high", n(spelling.C(), 4), n(spelling.F(), 4), 4, "perfect fourth", n(spelling.F(), 4), true},
		{"major third root low", n(spelling.C(), 4), n(spelling.E(), 4), 5, "major third", n(spelling.C(), 4), true},
		{"minor sixth root high", n(spelling.C(), 4), n(spelling.AFlat(), 4), 6, "minor sixth", n(spelling.AFlat(), 4), true},
		{"minor third root low", n(spelling.C(), 4), n(spelling.EFlat(), 4), 7, "minor third", n(spelling.C(), 4), true},
		{"major sixth root high", n(spelling.C(), 4), n(spelling.A(), 4), 8, "major sixth", n(spelling.A(), 4), true},
		{"major second root high", n(spelling.C(), 4), n(spelling.D(), 4), 9, "major second", n(spelling.D(), 4), true},
		{"minor seventh root low", n(spelling.C(), 4), n(spelling.BFlat(), 4), 10, "minor seventh", n(spelling.C(), 4), true},
		{"minor second root high", n(spelling.C(), 4), n(spelling.DFlat(), 4), 11, "minor second", n(spelling.DFlat(), 4), true},
		{"major seventh root low", n(spelling.C(), 4), n(spelling.B(), 4), 12, "major seventh", n(spelling.C(), 4), true},
		{"tritone has no root", n(spelling.C(), 4), n(spelling.FSharp(), 4), 13, "tritone", note.Note{}, false},
		{"descending fifth keeps the lower root", n(spelling.G(), 4), n(spelling.C(), 4), 3, "perfect fifth", n(spelling.C(), 4), true},
		{"compound fifth reduces", n(spelling.C(), 4), n(spelling.G(), 5), 3, "perfect fifth", n(spelling.C(), 4), true},
		{"enharmonic tritone", n(spelling.C(), 4), n(spelling.GFlat(), 4), 13, "tritone", note.Note{}, false},
		{"augmented fifth is a minor sixth", n(spelling.C(), 4), n(spelling.GSharp(), 4), 6, "minor sixth", n(spelling.GSharp(), 4), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d, err := hindemith.AnalyzeDyad(trad(), tc.from, tc.to)
			require.NoError(t, err)
			assert.True(t, d.IsValid())
			assert.Equal(t, tc.series, d.Series())
			assert.Equal(t, tc.dyadName, d.Name())
			assert.Equal(t, tc.hasRoot, d.HasRoot())
			assert.Equal(t, tc.series == 13, d.IsTritone())
			if tc.hasRoot {
				assert.True(t, d.Root().Equal(tc.root), "root: %v", d.Root())
			} else {
				assert.False(t, d.Root().IsValid())
			}
			assert.True(t, d.From().Equal(tc.from))
			assert.True(t, d.To().Equal(tc.to))
		})
	}

	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		_, err := hindemith.AnalyzeDyad(nil, n(spelling.C(), 4), n(spelling.G(), 4))
		assert.ErrorIs(t, err, hindemith.ErrUnset)
		_, err = hindemith.AnalyzeDyad(system.Positions19(), n(spelling.C(), 4), n(spelling.G(), 4))
		assert.ErrorIs(t, err, hindemith.ErrUnsupportedSystem)
		_, err = hindemith.AnalyzeDyad(trad(), note.Note{}, n(spelling.G(), 4))
		assert.ErrorIs(t, err, hindemith.ErrUnset)
		_, err = hindemith.AnalyzeDyad(trad(), note.MustFromSpelling(spelling.C(), cents.Zero(), nil), n(spelling.G(), 4))
		assert.ErrorIs(t, err, hindemith.ErrNoOctave)
		assert.Panics(t, func() { hindemith.MustAnalyzeDyad(nil, note.Note{}, note.Note{}) })
	})
}

// TestAnalyzeChord covers the acceptance chords and the group table.
func TestAnalyzeChord(t *testing.T) {
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
			"major triad root position", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)},
			hindemith.GroupI, hindemith.Sub1, n(spelling.C(), 4), n(spelling.C(), 4), true, 1,
		},
		{
			"major triad first inversion", []note.Note{n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.C(), 5)},
			hindemith.GroupI, hindemith.Sub2, n(spelling.E(), 4), n(spelling.C(), 5), true, 2,
		},
		{
			"dominant seventh root position", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BFlat(), 4)},
			hindemith.GroupII, hindemith.SubA, n(spelling.C(), 4), n(spelling.C(), 4), true, 3,
		},
		{
			"dominant seventh inversion", []note.Note{n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BFlat(), 4), n(spelling.C(), 5)},
			hindemith.GroupII, hindemith.SubB2, n(spelling.E(), 4), n(spelling.C(), 5), true, 5,
		},
		{
			"two tritones", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.GFlat(), 4), n(spelling.BFlat(), 4)},
			hindemith.GroupII, hindemith.SubB3, n(spelling.C(), 4), n(spelling.C(), 4), true, 6,
		},
		{
			"minor second with tritone", []note.Note{n(spelling.C(), 4), n(spelling.DFlat(), 4), n(spelling.E(), 4), n(spelling.G(), 4)},
			hindemith.GroupIV, hindemith.Sub1, n(spelling.C(), 4), n(spelling.C(), 4), true, 9,
		},
		{
			"seconds without tritone", []note.Note{n(spelling.C(), 4), n(spelling.D(), 4), n(spelling.E(), 4), n(spelling.G(), 4)},
			hindemith.GroupIII, hindemith.Sub1, n(spelling.C(), 4), n(spelling.C(), 4), true, 7,
		},
		{
			"augmented triad is indeterminate", []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.GSharp(), 4)},
			hindemith.GroupV, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 11,
		},
		{
			"diminished seventh is indeterminate", []note.Note{n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4), n(spelling.A(), 4)},
			hindemith.GroupVI, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 12,
		},
		{
			"diminished triad is indeterminate", []note.Note{n(spelling.C(), 4), n(spelling.EFlat(), 4), n(spelling.GFlat(), 4)},
			hindemith.GroupVI, hindemith.SubNone, n(spelling.C(), 4), note.Note{}, false, 12,
		},
		{
			"doubled single note", []note.Note{n(spelling.C(), 4), n(spelling.C(), 4)},
			hindemith.GroupI, hindemith.Sub1, n(spelling.C(), 4), n(spelling.C(), 4), true, 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := hindemith.AnalyzeChord(trad(), tc.notes)
			require.NoError(t, err)
			assert.True(t, c.IsValid())
			assert.Equal(t, tc.group, c.Group())
			assert.Equal(t, tc.sub, c.Subgroup())
			assert.True(t, c.Bass().Equal(tc.bass), "bass: %v", c.Bass())
			assert.Equal(t, tc.hasRoot, c.HasRoot())
			if tc.hasRoot {
				assert.True(t, c.Root().Equal(tc.root), "root: %v", c.Root())
				assert.True(t, c.Best().IsValid())
				assert.True(t, c.Best().Root().Equal(tc.root))
			} else {
				assert.False(t, c.Root().IsValid())
				assert.False(t, c.Best().IsValid())
			}
			assert.Equal(t, tc.value, c.Value())
			assert.Len(t, c.Dyads(), len(tc.notes)*(len(tc.notes)-1)/2)
			assert.Len(t, c.Notes(), len(tc.notes))
		})
	}

	t.Run("input is not changed", func(t *testing.T) {
		t.Parallel()
		ns := []note.Note{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)}
		c := hindemith.MustAnalyzeChord(trad(), ns)
		got := c.Notes()
		got[0] = n(spelling.D(), 4)
		assert.True(t, c.Notes()[0].Equal(ns[0]))
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		_, err := hindemith.AnalyzeChord(nil, nil)
		assert.ErrorIs(t, err, hindemith.ErrUnset)
		_, err = hindemith.AnalyzeChord(system.Positions19(), []note.Note{n(spelling.C(), 4), n(spelling.E(), 4)})
		assert.ErrorIs(t, err, hindemith.ErrUnsupportedSystem)
		_, err = hindemith.AnalyzeChord(trad(), []note.Note{n(spelling.C(), 4)})
		assert.ErrorIs(t, err, hindemith.ErrTooFewNotes)
		_, err = hindemith.AnalyzeChord(trad(), []note.Note{n(spelling.C(), 4), {}})
		assert.ErrorIs(t, err, hindemith.ErrUnset)
		_, err = hindemith.AnalyzeChord(trad(), []note.Note{n(spelling.C(), 4), note.MustFromSpelling(spelling.E(), cents.Zero(), nil)})
		assert.ErrorIs(t, err, hindemith.ErrNoOctave)
		assert.Panics(t, func() { hindemith.MustAnalyzeChord(trad(), nil) })
	})
}

// TestAnalyzeProgression covers harmonic fluctuation with the individual
// classifications.
func TestAnalyzeProgression(t *testing.T) {
	t.Parallel()

	chords := [][]note.Note{
		{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)},
		{n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.C(), 5)},
		{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4), n(spelling.BFlat(), 4)},
		{n(spelling.C(), 4), n(spelling.E(), 4), n(spelling.G(), 4)},
	}
	p, err := hindemith.AnalyzeProgression(trad(), chords)
	require.NoError(t, err)
	require.Len(t, p.Chords(), 4)
	assert.Equal(t, hindemith.GroupI, p.Chords()[0].Group())
	assert.Equal(t, hindemith.GroupII, p.Chords()[2].Group())

	steps := p.Steps()
	require.Len(t, steps, 3)
	assert.Equal(t, 1, steps[0].From())
	assert.Equal(t, 2, steps[0].To())
	assert.Equal(t, 1, steps[0].Delta())
	assert.Equal(t, hindemith.Rise, steps[0].Direction())
	assert.Equal(t, hindemith.Rise, steps[1].Direction())
	assert.Equal(t, hindemith.Fall, steps[2].Direction())
	assert.Equal(t, -2, steps[2].Delta())

	t.Run("empty and single", func(t *testing.T) {
		t.Parallel()
		p, err := hindemith.AnalyzeProgression(trad(), nil)
		require.NoError(t, err)
		assert.Empty(t, p.Chords())
		assert.Empty(t, p.Steps())
		p, err = hindemith.AnalyzeProgression(trad(), chords[:1])
		require.NoError(t, err)
		assert.Len(t, p.Chords(), 1)
		assert.Empty(t, p.Steps())
	})

	t.Run("level step", func(t *testing.T) {
		t.Parallel()
		p := hindemith.MustAnalyzeProgression(trad(), [][]note.Note{chords[0], chords[0]})
		require.Len(t, p.Steps(), 1)
		assert.Equal(t, hindemith.Level, p.Steps()[0].Direction())
		assert.Equal(t, 0, p.Steps()[0].Delta())
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		_, err := hindemith.AnalyzeProgression(nil, nil)
		assert.ErrorIs(t, err, hindemith.ErrUnset)
		_, err = hindemith.AnalyzeProgression(trad(), [][]note.Note{nil})
		assert.ErrorIs(t, err, hindemith.ErrTooFewNotes)
		assert.Panics(t, func() { hindemith.MustAnalyzeProgression(trad(), [][]note.Note{nil}) })
	})
}
