package chord_test

import (
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-7, the sources of settings are not read: КП-4, Р-7, П-2; architecture
// 5.5…5.13.

func TestSourceNotRead(t *testing.T) {
	t.Parallel()
	src := newCounting()
	// The notes bound to src; creating them may read it, so the counter is
	// taken after.
	bc4 := pn(c4.Spelling(), 4, 0, src)
	be4 := pn(e4.Spelling(), 4, 0, src)
	bg4 := pn(g4.Spelling(), 4, 0, src)
	bd4 := pn(d4.Spelling(), 4, 0, src)
	bb4 := pn(b4.Spelling(), 4, 0, src)
	bc := nn(cNo.Spelling(), src)

	// Precondition: the counter works, the note reads its source when it
	// needs settings.
	probe := pn(c4.Spelling(), 4, 0, src)
	start := src.reads()
	_, err := probe.Frequency()
	require.NoError(t, err)
	require.Greater(t, src.reads(), start, "precondition: a note reads its source for a frequency")

	readsBefore := src.reads()

	// Every operation of the chord, on the success and failure paths, with
	// rules that do not read sources.
	c, err := chord.New(bc4, be4, bg4, bc4)
	require.NoError(t, err)
	other := chord.MustNew(be4, bd4, bc)
	_, _ = chord.New(bc4, note.Note{})

	_ = c.IsValid()
	_ = c.IsEmpty()
	_ = c.Len()
	_ = c.Notes()
	for range c.All() {
	}
	for range c.Values() {
	}
	_, _ = c.At(0)
	_, _ = c.At(9)
	_ = c.SortedNotes()
	_ = c.DistinctNotes()
	_ = c.HasRepeats()
	_ = c.HasNotesWithOctave()
	_ = c.HasNotesWithoutOctave()
	_ = c.Contains(bd4)
	_ = c.Count(bc4)
	_, _, _ = c.Find(bg4)
	_ = c.Indices(bc4)
	_ = c.ContainsAll(bc4, bc4, bd4)
	_ = c.ContainsAny(bd4, bc)
	_ = c.ContainsRecord(bc4)
	_ = c.ContainsSpelling(bc)
	_, _ = c.ContainsFunc(hasOctave)
	_, _ = c.CountFunc(hasOctave)
	_, _ = c.IndicesFunc(hasOctave)
	_, _ = c.NotesFunc(hasOctave)
	_ = c.Equal(other)
	_ = c.EqualInOrder(other)
	_ = c.SameRecords(other)
	_ = c.Includes(other)
	_ = c.HasCommonNotes(other)
	_ = c.Compare(other)
	_ = c.String()

	d := c
	require.NoError(t, d.Add(bd4, bc))
	require.NoError(t, d.AddMissing(bd4, bb4))
	require.NoError(t, d.Insert(1, bc4))
	require.NoError(t, d.Remove(bc))
	require.NoError(t, d.RemoveAll(bd4))
	require.NoError(t, d.RemoveAt(0))
	require.NoError(t, d.Replace(bc4, bd4))
	require.NoError(t, d.ReplaceAll(bd4, bc4))
	require.NoError(t, d.ReplaceAt(0, bg4))
	require.NoError(t, d.ReplaceFunc(up))
	require.NoError(t, d.ReplaceAtFunc(0, up))
	require.NoError(t, d.RemoveFunc(func(n note.Note) (bool, error) { return !n.HasOctave(), nil }))
	require.NoError(t, d.KeepFunc(hasOctave))
	require.NoError(t, d.RemoveRepeats())
	require.NoError(t, d.Union(other))
	require.NoError(t, d.Intersect(c))
	require.NoError(t, d.Subtract(other))
	require.NoError(t, d.Sort())
	require.NoError(t, d.SortFunc(byRecord))
	require.NoError(t, d.Reverse())
	require.NoError(t, d.Shift(1))
	require.NoError(t, d.ShiftDown(1))
	require.NoError(t, d.Up())
	require.NoError(t, d.Down())
	require.NoError(t, d.Clear())

	e := c
	_ = e.Add(note.Note{})
	_ = e.Remove(bd4)
	_ = e.ReplaceAt(9, bd4)
	_ = e.Union(chord.Chord{})
	_ = e.Add(bc)
	_ = e.Shift(1) // fails at the note without an octave
	_ = e.Up()

	assert.Equal(t, readsBefore, src.reads(), "no operation of the chord reads the source (КП-4)")
}

func TestSourceRuleMayRead(t *testing.T) {
	t.Parallel()
	// Architecture 5.8: a rule may read the sources; these reads are the
	// rule's, not the chord's.
	src := newCounting()
	c := chord.MustNew(pn(g4.Spelling(), 4, 0, src), pn(c4.Spelling(), 4, 0, src))
	before := src.reads()
	require.NoError(t, c.SortFunc(func(a, b note.Note) (int, error) { return a.ComparePosition(b) }))
	assert.Greater(t, src.reads(), before, "the comparison by position reads the sources")
	assertChord(t, list(pn(c4.Spelling(), 4, 0, src), pn(g4.Spelling(), 4, 0, src)), c)
}

func TestSourceK40Frequencies(t *testing.T) {
	t.Parallel()
	// K-40, С-12: the frequencies through package note; the chord is the same.
	st := std()
	c := mk(t, pn(c4.Spelling(), 4, 0, st), pn(e4.Spelling(), 4, 0, st))
	before := stateOf(c)
	var got []float64
	for n := range c.Values() {
		f, err := n.Frequency()
		require.NoError(t, err)
		got = append(got, f)
	}
	require.Len(t, got, 2)
	assert.InDelta(t, 261.6256, got[0], 5e-5)
	assert.InDelta(t, 329.6276, got[1], 5e-5)
	assertState(t, before, c)
}

func TestSourceSharedOnlyBetweenCopies(t *testing.T) {
	t.Parallel()
	// П-7: copies share only the sources of the notes.
	src := newCounting()
	c := mk(t, pn(c4.Spelling(), 4, 0, src))
	d := c
	require.NoError(t, d.Add(e4))
	cn, dn := c.MustAt(0), d.MustAt(0)
	assert.True(t, cn.SameSource(dn), "the notes of the copies share their source")
	assert.Equal(t, 1, c.Len())
}
