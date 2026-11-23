package note_test

import (
	"slices"
	"testing"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Exact method-expression types prevent an unnoticed return type change.
var (
	_ func(note.Notes) (note.Notes, error)                = note.Notes.UniqueBySpelling
	_ func(note.Notes) note.Notes                         = note.Notes.MustUniqueBySpelling
	_ func(note.Notes) ([]spelling.Spelling, error)       = note.Notes.Spellings
	_ func(note.Notes) []spelling.Spelling                = note.Notes.MustSpellings
	_ func(note.Notes) (note.Notes, error)                = note.Notes.SortedByPositionDescending
	_ func(note.Notes) note.Notes                         = note.Notes.MustSortedByPositionDescending
	_ func(note.Notes) (note.Notes, error)                = note.Notes.SortedBySoundDescending
	_ func(note.Notes) note.Notes                         = note.Notes.MustSortedBySoundDescending
	_ func(note.Notes, settings.Source) ([]string, error) = note.Notes.FormatIn
	_ func(note.Notes, settings.Source) []string          = note.Notes.MustFormatIn
	_ func(note.Notes, settings.Source) ([]string, error) = note.Notes.FormatNamesIn
	_ func(note.Notes, settings.Source) []string          = note.Notes.MustFormatNamesIn
	_ func([]note.Note) ([]note.Note, error)              = note.SortedByPositionDescending
	_ func([]note.Note) []note.Note                       = note.MustSortedByPositionDescending
	_ func([]note.Note) ([]note.Note, error)              = note.SortedBySoundDescending
	_ func([]note.Note) []note.Note                       = note.MustSortedBySoundDescending
)

// UniqueBySpelling keeps the first complete note of each name; the octave,
// the correction and the source do not take part in the key.
func TestNotesUniqueBySpelling(t *testing.T) {
	t.Parallel()

	t.Run("keeps the first note of each spelling", func(t *testing.T) {
		t.Parallel()
		src := newCounting(std())
		c5 := mk(t, spelling.C(), 5, ct(17), src)
		c4 := mk(t, spelling.C(), 4, ct(-3), src)
		cs4 := mk(t, spelling.CSharp(), 4, ct(0), src)
		df4 := mk(t, spelling.DFlat(), 4, ct(0), src)
		ns := note.Notes{c5, c4, cs4, df4}
		before := slices.Clone([]note.Note(ns))
		reads := src.reads()

		got, err := ns.UniqueBySpelling()
		require.NoError(t, err)
		assertNotes(t, []want{
			wp(spelling.C(), 5, ct(17)).boundTo(src),
			wp(spelling.CSharp(), 4, ct(0)).boundTo(src),
			wp(spelling.DFlat(), 4, ct(0)).boundTo(src),
		}, got)
		assert.Equal(t, reads, src.reads(), "UniqueBySpelling reads no source")
		assertNotesPreserved(t, before, ns)

		got[0].MustSetCents(ct(99))
		assertNotesPreserved(t, before, ns)
	})

	t.Run("names without octaves and mixed forms", func(t *testing.T) {
		t.Parallel()
		c := mkName(t, spelling.C(), ct(5), nil)
		c4 := mk(t, spelling.C(), 4, ct(0), nil)
		bs3 := mk(t, spelling.BSharp(), 4, ct(0), nil) // written B♯3, Octave 4
		ns := note.Notes{c, c4, bs3, c}
		got, err := ns.UniqueBySpelling()
		require.NoError(t, err)
		assertNotes(t, []want{
			wn(spelling.C(), ct(5)).free(),
			wp(spelling.BSharp(), 4, ct(0)).free(),
		}, got)
	})

	t.Run("empty and nil", func(t *testing.T) {
		t.Parallel()
		for _, ns := range []note.Notes{nil, {}, make(note.Notes, 0)} {
			got, err := ns.UniqueBySpelling()
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Empty(t, got)
		}
	})

	t.Run("unset note", func(t *testing.T) {
		t.Parallel()
		a := mkName(t, spelling.C(), ct(0), nil)
		ns := note.Notes{a, {}}
		got, err := ns.UniqueBySpelling()
		assert.Nil(t, got)
		assertErr(t, err, note.ErrUnset)
		assertPanicErr(t, func() { ns.MustUniqueBySpelling() }, note.ErrUnset)
	})
}

// Spellings projects the names in order, including repetitions, without a
// tuning or a source lookup.
func TestNotesSpellings(t *testing.T) {
	t.Parallel()

	t.Run("ordered names with repetitions", func(t *testing.T) {
		t.Parallel()
		src := newCounting(std())
		c4 := mk(t, spelling.C(), 4, ct(0), src)
		ef4 := mk(t, spelling.EFlat(), 4, ct(0), src)
		c5 := mk(t, spelling.C(), 5, ct(0), src)
		bs3 := mk(t, spelling.BSharp(), 4, ct(0), src) // written B♯3, Octave 4
		ns := note.Notes{c4, ef4, c5, bs3}
		reads := src.reads()

		got, err := ns.Spellings()
		require.NoError(t, err)
		require.Len(t, got, 4)
		assert.True(t, got[0].Equal(spelling.C()))
		assert.True(t, got[1].Equal(spelling.EFlat()))
		assert.True(t, got[2].Equal(spelling.C()))
		assert.True(t, got[3].Equal(spelling.BSharp()), "the name of the written pitch")
		assert.Equal(t, reads, src.reads(), "Spellings reads no source")
	})

	t.Run("empty and nil", func(t *testing.T) {
		t.Parallel()
		for _, ns := range []note.Notes{nil, {}, make(note.Notes, 0)} {
			got, err := ns.Spellings()
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Empty(t, got)
		}
	})

	t.Run("unset note", func(t *testing.T) {
		t.Parallel()
		ns := note.Notes{{}}
		got, err := ns.Spellings()
		assert.Nil(t, got)
		assertErr(t, err, note.ErrUnset)
		assertPanicErr(t, func() { ns.MustSpellings() }, note.ErrUnset)
	})
}

// The descending sorts use the comparison, validation and snapshot rules of
// the ascending ones; equal keys keep their input order.
func TestNotesSortedDescending(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"position", "sound"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			sort := func(ns note.Notes) (note.Notes, error) {
				if kind == "position" {
					return ns.SortedByPositionDescending()
				}
				return ns.SortedBySoundDescending()
			}
			mustSort := func(ns note.Notes) note.Notes {
				if kind == "position" {
					return ns.MustSortedByPositionDescending()
				}
				return ns.MustSortedBySoundDescending()
			}

			t.Run("stable ties keep input order", func(t *testing.T) {
				t.Parallel()
				src := newCounting(std())
				cs := mk(t, spelling.CSharp(), 4, ct(0), src)
				df := mk(t, spelling.DFlat(), 4, ct(0), src)
				g := mk(t, spelling.G(), 4, ct(0), src)
				ns := note.Notes{cs, df, g}
				before := slices.Clone([]note.Note(ns))

				wantOrder := []want{
					wp(spelling.G(), 4, ct(0)).boundTo(src),
					wp(spelling.CSharp(), 4, ct(0)).boundTo(src),
					wp(spelling.DFlat(), 4, ct(0)).boundTo(src),
				}
				got, err := sort(ns)
				require.NoError(t, err)
				assertNotes(t, wantOrder, got)
				assertNotesPreserved(t, before, ns)
				assertNotes(t, wantOrder, mustSort(ns))

				got[1].MustSetCents(ct(5))
				assertNotesPreserved(t, before, ns)
			})

			t.Run("empty and nil", func(t *testing.T) {
				t.Parallel()
				for _, ns := range []note.Notes{nil, {}, make(note.Notes, 0)} {
					got, err := sort(ns)
					require.NoError(t, err)
					assert.NotNil(t, got)
					assert.Empty(t, got)
				}
			})

			t.Run("unset note", func(t *testing.T) {
				t.Parallel()
				a := mk(t, spelling.C(), 4, ct(0), std())
				ns := note.Notes{a, {}}
				got, err := sort(ns)
				assert.Nil(t, got)
				assertErr(t, err, note.ErrUnset)
				assertPanicErr(t, func() { mustSort(ns) }, note.ErrUnset)
			})

			t.Run("unbound note", func(t *testing.T) {
				t.Parallel()
				ns := note.Notes{mk(t, spelling.C(), 4, ct(0), nil)}
				got, err := sort(ns)
				assert.Nil(t, got)
				assertErr(t, err, note.ErrNoSource)
			})

			t.Run("note without octave", func(t *testing.T) {
				t.Parallel()
				ns := note.Notes{mkName(t, spelling.C(), ct(0), std())}
				got, err := sort(ns)
				assert.Nil(t, got)
				assertErr(t, err, note.ErrNoOctave)
			})
		})
	}

	t.Run("position rejects different systems", func(t *testing.T) {
		t.Parallel()
		ns := note.Notes{
			mk(t, spelling.C(), 4, ct(0), std()),
			mk(t, spelling.C(), 4, ct(0), set19()),
		}
		got, err := ns.SortedByPositionDescending()
		assert.Nil(t, got)
		assertErr(t, err, note.ErrDifferentSystems)
	})

	t.Run("package functions match the methods", func(t *testing.T) {
		t.Parallel()
		src := std()
		ns := note.Notes{
			mk(t, spelling.C(), 4, ct(0), src),
			mk(t, spelling.G(), 4, ct(0), src),
		}
		byPosition, err := note.SortedByPositionDescending(ns)
		require.NoError(t, err)
		assertNotes(t, []want{
			wp(spelling.G(), 4, ct(0)).boundTo(src),
			wp(spelling.C(), 4, ct(0)).boundTo(src),
		}, byPosition)
		bySound, err := note.SortedBySoundDescending(ns)
		require.NoError(t, err)
		assertNotes(t, []want{
			wp(spelling.G(), 4, ct(0)).boundTo(src),
			wp(spelling.C(), 4, ct(0)).boundTo(src),
		}, bySound)
		assertPanicErr(t, func() { note.MustSortedByPositionDescending([]note.Note{{}}) }, note.ErrUnset)
		assertPanicErr(t, func() { note.MustSortedBySoundDescending([]note.Note{{}}) }, note.ErrUnset)
	})
}

// Sound and position descending orders can differ: equal positions with
// different tunings are a tie by position and distinct by sound.
func TestNotesDescendingSoundVsPosition(t *testing.T) {
	t.Parallel()

	loud := newCounting(std())
	soft := newCounting(stdAt(415))
	c4soft := mk(t, spelling.C(), 4, ct(0), soft)
	bs3loud := mk(t, spelling.BSharp(), 4, ct(0), loud) // written B♯3, Octave 4, one position with C4
	ns := note.Notes{c4soft, bs3loud}

	bySound, err := ns.SortedBySoundDescending()
	require.NoError(t, err)
	assertNotes(t, []want{
		wp(spelling.BSharp(), 4, ct(0)).boundTo(loud),
		wp(spelling.C(), 4, ct(0)).boundTo(soft),
	}, bySound)

	byPosition, err := ns.SortedByPositionDescending()
	require.NoError(t, err)
	assertNotes(t, []want{
		wp(spelling.C(), 4, ct(0)).boundTo(soft),
		wp(spelling.BSharp(), 4, ct(0)).boundTo(loud),
	}, byPosition)
}

// FormatIn and FormatNamesIn render every note in the notation of one
// explicit source, read once, and propagate the formatting errors.
func TestNotesFormatIn(t *testing.T) {
	t.Parallel()

	t.Run("renders in order with octaves where present", func(t *testing.T) {
		t.Parallel()
		src := newCounting(std())
		ns := note.Notes{
			mk(t, spelling.C(), 4, ct(0), nil),
			mk(t, spelling.EFlat(), 4, ct(0), nil),
			mk(t, spelling.G(), 4, ct(0), nil),
			mkName(t, spelling.G(), ct(0), nil),
			mk(t, spelling.C(), 4, ct(0), nil),
		}
		reads := src.reads()
		got, err := ns.FormatIn(src)
		require.NoError(t, err)
		assert.Equal(t, []string{"C4", "E♭4", "G4", "G", "C4"}, got)
		assert.Equal(t, reads+1, src.reads(), "the explicit source is read once")
	})

	t.Run("names only", func(t *testing.T) {
		t.Parallel()
		src := newCounting(std())
		ns := note.Notes{
			mk(t, spelling.C(), 4, ct(0), nil),
			mk(t, spelling.EFlat(), 4, ct(0), nil),
			mk(t, spelling.G(), 4, ct(0), nil),
		}
		got, err := ns.FormatNamesIn(src)
		require.NoError(t, err)
		assert.Equal(t, []string{"C", "E♭", "G"}, got)
	})

	t.Run("the own sources are not read", func(t *testing.T) {
		t.Parallel()
		own := newCounting(std())
		explicit := newCounting(std())
		ns := note.Notes{mk(t, spelling.C(), 4, ct(0), own)}
		ownReads, explicitReads := own.reads(), explicit.reads()
		_, err := ns.FormatIn(explicit)
		require.NoError(t, err)
		assert.Equal(t, ownReads, own.reads())
		assert.Equal(t, explicitReads+1, explicit.reads())
	})

	t.Run("empty and nil", func(t *testing.T) {
		t.Parallel()
		src := newCounting(std())
		reads := src.reads()
		for _, ns := range []note.Notes{nil, {}, make(note.Notes, 0)} {
			got, err := ns.FormatIn(src)
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Empty(t, got)
			names, err := ns.FormatNamesIn(src)
			require.NoError(t, err)
			assert.NotNil(t, names)
			assert.Empty(t, names)
		}
		assert.Equal(t, reads, src.reads(), "an empty list reads no source")
	})

	t.Run("nil source", func(t *testing.T) {
		t.Parallel()
		ns := note.Notes{mk(t, spelling.C(), 4, ct(0), nil)}
		got, err := ns.FormatIn(nil)
		assert.Nil(t, got)
		assertErr(t, err, note.ErrUnset)
		assertPanicErr(t, func() { ns.MustFormatIn(nil) }, note.ErrUnset)
		names, err := ns.FormatNamesIn(nil)
		assert.Nil(t, names)
		assertErr(t, err, note.ErrUnset)
		assertPanicErr(t, func() { ns.MustFormatNamesIn(nil) }, note.ErrUnset)
		var empty note.Notes
		_, err = empty.FormatIn(nil)
		assertErr(t, err, note.ErrUnset)
	})

	t.Run("unset note", func(t *testing.T) {
		t.Parallel()
		ns := note.Notes{mk(t, spelling.C(), 4, ct(0), nil), {}}
		got, err := ns.FormatIn(std())
		assert.Nil(t, got)
		assertErr(t, err, note.ErrUnset)
	})

	t.Run("unset settings", func(t *testing.T) {
		t.Parallel()
		ns := note.Notes{mk(t, spelling.C(), 4, ct(0), nil)}
		got, err := ns.FormatIn(&unsetSource{})
		assert.Nil(t, got)
		assertErr(t, err, note.ErrUnsetSettings, settings.ErrUnset)
	})

	t.Run("inadmissible name in the explicit system", func(t *testing.T) {
		t.Parallel()
		ns := note.Notes{mk(t, spelling.CDoubleSharp(), 4, ct(0), nil)}
		got, err := ns.FormatIn(setK25())
		assert.Nil(t, got)
		assertErr(t, err, note.ErrInadmissible, settings.ErrInadmissible)
		assertPanicErr(t, func() { ns.MustFormatIn(setK25()) }, note.ErrInadmissible, settings.ErrInadmissible)
	})

	t.Run("octave repeat budget propagates", func(t *testing.T) {
		t.Parallel()
		s := repeatNoteSet(t, notation.OctaveStyleLilyPond)
		n := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 65540), ct(0), s)
		ns := note.Notes{n}
		got, err := ns.FormatIn(s)
		assert.Nil(t, got)
		repeatNoteError(t, err)
		assertPanicErr(t, func() { ns.MustFormatIn(s) }, nil, notation.ErrOctaveRepeatLimit)
		names, err := ns.FormatNamesIn(s)
		require.NoError(t, err)
		assert.Equal(t, []string{"C"}, names, "name bytes are outside the budget")
	})

	t.Run("Must variants return the texts", func(t *testing.T) {
		t.Parallel()
		src := std()
		ns := note.Notes{mk(t, spelling.C(), 4, ct(0), nil)}
		assert.Equal(t, []string{"C4"}, ns.MustFormatIn(src))
		assert.Equal(t, []string{"C"}, ns.MustFormatNamesIn(src))
	})
}
