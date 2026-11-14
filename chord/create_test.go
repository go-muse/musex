package chord_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-1. Creation, reading and the string form: ОП-1…ОП-7, ОП-38; П-1, П-5;
// К-1…К-4, К-15, К-41 (business-spec); architecture 5.5, 5.6, 5.12.

func TestCreateNew(t *testing.T) {
	t.Parallel()
	s19 := set19()
	st := std()
	tests := []struct {
		name  string
		notes []note.Note
	}{
		{name: "K-1 C4 E4 G4", notes: list(c4, e4, g4)},
		{name: "K-2 no notes", notes: nil},
		{name: "empty slice", notes: []note.Note{}},
		{name: "one note", notes: list(c4)},
		{name: "P-3 repeats are kept", notes: list(c4, e4, c4, c4)},
		{name: "P-2 any notes in any combination", notes: list(cNo, c4, c4p5, pn(c4.Spelling(), 4, 0, st), pn(c4.Spelling(), 4, 0, s19), bs4, eNo)},
		{name: "P-4 order of the list", notes: list(g4, c4, e4)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, err := chord.New(tt.notes...)
			require.NoError(t, err)
			assertChord(t, tt.notes, c)

			m := chord.MustNew(tt.notes...)
			assertChord(t, tt.notes, m)
		})
	}
}

func TestCreateK1(t *testing.T) {
	t.Parallel()
	c, err := chord.New(c4, e4, g4)
	require.NoError(t, err)
	assert.True(t, c.IsValid())
	assert.False(t, c.IsEmpty())
	assert.Equal(t, 3, c.Len())
	assertNotes(t, list(c4, e4, g4), c.Notes())
	n, err := c.At(1)
	require.NoError(t, err)
	assert.True(t, sameNote(e4, n), "At(1) = %v, want E4", n)
}

func TestCreateK2Empty(t *testing.T) {
	t.Parallel()
	c, err := chord.New()
	require.NoError(t, err)
	assert.True(t, c.IsValid())
	assert.True(t, c.IsEmpty())
	assert.Zero(t, c.Len())
	notes := c.Notes()
	assert.NotNil(t, notes, "Notes of the empty chord is not nil")
	assert.Empty(t, notes)

	other := chord.MustNew()
	assert.True(t, c.Equal(other), "two empty chords are equal by record")
	assert.True(t, c.EqualInOrder(other), "two empty chords are equal in order")

	var unset chord.Chord
	assert.False(t, c.Equal(unset), "the empty chord is not equal to the unset chord")
	assert.False(t, unset.Equal(c))
	assert.False(t, c.EqualInOrder(unset))
}

func TestCreateUnsetNote(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		notes []note.Note
	}{
		{name: "K-3 C4 and an unset note", notes: list(c4, note.Note{})},
		{name: "unset note first", notes: list(note.Note{}, c4)},
		{name: "unset note in the middle", notes: list(c4, note.Note{}, e4)},
		{name: "only an unset note", notes: list(note.Note{})},
		{name: "two unset notes", notes: list(note.Note{}, note.Note{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			before := stateNotes(tt.notes)
			c, err := chord.New(tt.notes...)
			assertOnly(t, err, chord.ErrUnset)
			assertUnsetChord(t, c)
			assertNotes(t, before, tt.notes, "the list is not changed")

			assertPanicsWith(t, chord.ErrUnset, func() { chord.MustNew(tt.notes...) })
		})
	}
}

// stateNotes returns a copy of a list of notes.
func stateNotes(notes []note.Note) []note.Note {
	return append([]note.Note{}, notes...)
}

func TestReadUnset(t *testing.T) {
	t.Parallel()
	var c chord.Chord // K-4: the zero value is the unset chord

	assert.False(t, c.IsValid())
	assert.True(t, c.IsEmpty())
	assert.Zero(t, c.Len())
	for name, got := range map[string][]note.Note{
		"Notes":         c.Notes(),
		"SortedNotes":   c.SortedNotes(),
		"DistinctNotes": c.DistinctNotes(),
	} {
		assert.NotNil(t, got, name)
		assert.Empty(t, got, name)
	}
	assert.False(t, c.HasRepeats())
	assert.False(t, c.HasNotesWithOctave())
	assert.False(t, c.HasNotesWithoutOctave())
	for range c.All() {
		assert.Fail(t, "All of the unset chord yields nothing")
	}
	for range c.Values() {
		assert.Fail(t, "Values of the unset chord yields nothing")
	}

	for _, i := range []int{0, -1, 1, math.MaxInt} {
		n, err := c.At(i)
		assertOnly(t, err, chord.ErrUnset, "At(%d): E-1 comes before E-3", i)
		assert.False(t, n.IsValid(), "At(%d) gives the unset note on error", i)
	}
}

func TestReadEmpty(t *testing.T) {
	t.Parallel()
	c := chord.MustNew()
	assert.True(t, c.IsValid())
	assert.True(t, c.IsEmpty())
	assert.Zero(t, c.Len())
	for name, got := range map[string][]note.Note{
		"Notes":         c.Notes(),
		"SortedNotes":   c.SortedNotes(),
		"DistinctNotes": c.DistinctNotes(),
	} {
		assert.NotNil(t, got, name)
		assert.Empty(t, got, name)
	}
	assert.False(t, c.HasRepeats())
	assert.False(t, c.HasNotesWithOctave())
	assert.False(t, c.HasNotesWithoutOctave())
	for range c.All() {
		assert.Fail(t, "All of the empty chord yields nothing")
	}
	for range c.Values() {
		assert.Fail(t, "Values of the empty chord yields nothing")
	}
	n, err := c.At(0)
	assertOnly(t, err, chord.ErrIndex)
	assert.False(t, n.IsValid())
}

func TestReadNotesAllValues(t *testing.T) {
	t.Parallel()
	st := std()
	tests := []struct {
		name  string
		notes []note.Note
	}{
		{name: "K-1", notes: list(c4, e4, g4)},
		{name: "K-7 repeats", notes: list(c4, e4, c4)},
		{name: "sources and notes without an octave", notes: list(pn(c4.Spelling(), 4, 0, st), cNo, c4p30, pn(e4.Spelling(), 4, 0, set19()))},
		{name: "one note", notes: list(g4)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, tt.notes...)
			assertNotes(t, tt.notes, c.Notes(), "Notes")

			var idx []int
			var all []note.Note
			for i, n := range c.All() {
				idx = append(idx, i)
				all = append(all, n)
			}
			wantIdx := make([]int, len(tt.notes))
			for i := range wantIdx {
				wantIdx[i] = i
			}
			assert.Equal(t, wantIdx, idx, "All yields the numbers in order")
			assertNotes(t, tt.notes, all, "All")

			var vals []note.Note
			for n := range c.Values() {
				vals = append(vals, n)
			}
			assertNotes(t, tt.notes, vals, "Values")

			for i, want := range tt.notes {
				got, err := c.At(i)
				require.NoError(t, err)
				assert.True(t, sameNote(want, got), "At(%d) = %v, want %v", i, got, want)
				assert.True(t, sameNote(want, c.MustAt(i)), "MustAt(%d)", i)
			}
		})
	}
}

func TestReadSequenceBreak(t *testing.T) {
	t.Parallel()
	c := mk(t, c4, e4, g4, b4)

	var got []note.Note
	for i, n := range c.All() {
		got = append(got, n)
		if i == 1 {
			break
		}
	}
	assertNotes(t, list(c4, e4), got, "break stops All")

	got = nil
	for n := range c.Values() {
		got = append(got, n)
		if len(got) == 3 {
			break
		}
	}
	assertNotes(t, list(c4, e4, g4), got, "break stops Values")
}

func TestReadAtOutOfRange(t *testing.T) {
	t.Parallel()
	c := mk(t, c4, e4, g4)
	for _, i := range []int{-1, 3, 4, math.MaxInt, math.MinInt} {
		n, err := c.At(i)
		assertOnly(t, err, chord.ErrIndex, "At(%d)", i)
		assert.False(t, n.IsValid(), "At(%d) gives the unset note on error", i)
	}
}

func TestReadSortedNotes(t *testing.T) {
	t.Parallel()
	st := std()
	s19 := set19()
	c4st := pn(c4.Spelling(), 4, 0, st)
	c419 := pn(c4.Spelling(), 4, 0, s19)
	tests := []struct {
		name  string
		notes []note.Note
		want  []note.Note
	}{
		{name: "K-15 B♯4 C4+10 C♯ C4", notes: list(bs4, c4p10, csNo, c4), want: list(csNo, c4, c4p10, bs4)},
		{name: "K-15 equal notes keep the record order", notes: list(c4st, e4, c419), want: list(c4st, c419, e4)},
		{name: "K-26 E4 C4 G4", notes: list(e4, c4, g4), want: list(c4, e4, g4)},
		{name: "a note without an octave comes first", notes: list(e4, cNo), want: list(cNo, e4)},
		{name: "repeats", notes: list(g4, c4, g4, c4), want: list(c4, c4, g4, g4)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, tt.notes...)
			before := stateOf(c)
			assertNotes(t, tt.want, c.SortedNotes())
			assertState(t, before, c, "the chord is the same")
		})
	}
}

func TestReadDistinctNotes(t *testing.T) {
	t.Parallel()
	st := std()
	s19 := set19()
	c4st := pn(c4.Spelling(), 4, 0, st)
	c419 := pn(c4.Spelling(), 4, 0, s19)
	tests := []struct {
		name  string
		notes []note.Note
		want  []note.Note
	}{
		{name: "K-7", notes: list(c4, e4, c4), want: list(c4, e4)},
		{name: "K-8 corrections differ", notes: list(c4, c4p5), want: list(c4, c4p5)},
		{name: "K-9 with and without an octave", notes: list(cNo, c4), want: list(cNo, c4)},
		{name: "R-11 the first occurrence with its source", notes: list(c419, e4, c4st, e4), want: list(c419, e4)},
		{name: "no repeats", notes: list(g4, e4, c4), want: list(g4, e4, c4)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, tt.notes...)
			before := stateOf(c)
			assertNotes(t, tt.want, c.DistinctNotes())
			assertState(t, before, c, "the chord is the same")
		})
	}
}

func TestReadFlags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                         string
		notes                        []note.Note
		repeats, withOct, withoutOct bool
	}{
		{name: "K-7 C4 E4 C4", notes: list(c4, e4, c4), repeats: true, withOct: true},
		{name: "K-8 C4 C4+5", notes: list(c4, c4p5), withOct: true},
		{name: "K-9 C C4", notes: list(cNo, c4), withOct: true, withoutOct: true},
		{name: "P-3 one note with two sources", notes: list(pn(c4.Spelling(), 4, 0, std()), pn(c4.Spelling(), 4, 0, set19())), repeats: true, withOct: true},
		{name: "B♯4 and C4 are different", notes: list(bs4, c4), withOct: true},
		{name: "only without an octave", notes: list(cNo, eNo, cNo), repeats: true, withoutOct: true},
		{name: "empty", notes: list()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := mk(t, tt.notes...)
			assert.Equal(t, tt.repeats, c.HasRepeats(), "HasRepeats")
			assert.Equal(t, tt.withOct, c.HasNotesWithOctave(), "HasNotesWithOctave")
			assert.Equal(t, tt.withoutOct, c.HasNotesWithoutOctave(), "HasNotesWithoutOctave")
		})
	}
}

func TestReadRepeatedCalls(t *testing.T) {
	t.Parallel()
	c := mk(t, c4, e4, c4, pn(g4.Spelling(), 4, 0, std()))
	before := stateOf(c)
	for range 3 { // G-7: repeated reading gives the same results
		assertNotes(t, before.notes, c.Notes())
		assertNotes(t, list(c4, c4, e4, before.notes[3]), c.SortedNotes())
		assertNotes(t, list(c4, e4, before.notes[3]), c.DistinctNotes())
		assert.True(t, c.HasRepeats())
		assert.Equal(t, 2, c.Count(c4))
	}
	assertState(t, before, c)
}

func TestStringForms(t *testing.T) {
	t.Parallel()
	var unset chord.Chord
	empty := chord.MustNew()
	withNotes := chord.MustNew(c4p30, e4)

	// K-41; architecture 5.12.
	assert.Equal(t, "[C ASPN 4, +30 cents, unbound; E ASPN 4, 0 cents, unbound]", withNotes.String())
	assert.Equal(t, "["+c4p30.String()+"; "+e4.String()+"]", withNotes.String(), "the strings of the notes in the record order")
	assert.Equal(t, "[]", empty.String())
	assert.Equal(t, "Chord(unset)", unset.String())
	assert.NotEqual(t, empty.String(), unset.String())

	one := chord.MustNew(c4)
	assert.Equal(t, "["+c4.String()+"]", one.String())
	assert.Equal(t, "["+cNo.String()+"; "+c4.String()+"]", chord.MustNew(cNo, c4).String())

	bound := chord.MustNew(pn(c4.Spelling(), 4, 0, std()), e4)
	assert.Equal(t, "[C ASPN 4, 0 cents, bound; E ASPN 4, 0 cents, unbound]", bound.String())

	reversed := chord.MustNew(e4, c4p30)
	assert.NotEqual(t, withNotes.String(), reversed.String(), "the string follows the record order")
}

func TestStringStringer(t *testing.T) {
	t.Parallel()
	var s fmt.Stringer = chord.MustNew(c4, e4) // implemented by the value
	assert.Equal(t, "["+c4.String()+"; "+e4.String()+"]", s.String())
	assert.Equal(t, "Chord(unset)", fmt.Sprint(chord.Chord{}))
	assert.Equal(t, "[]", fmt.Sprint(chord.MustNew()))
}
