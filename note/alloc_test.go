package note_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"

	"github.com/stretchr/testify/require"
)

// Allocations (architecture 6.13, NF-3): the rows with "0 per call". The
// tests do not call t.Parallel: testing.AllocsPerRun counts the allocations of
// the whole process, so they run in the sequential phase of the package
// tests. Other operations promise no number of allocations.

// Sinks keep the results of the measured calls; concrete types, not
// interfaces.
var (
	sinkBool      bool
	sinkInt       int
	sinkSpelling  spelling.Spelling
	sinkLetter    letter.Letter
	sinkAlt       alteration.Alteration
	sinkCents     cents.Cents
	sinkPitch     pitch.Pitch
	sinkOctave    octave.Octave
	sinkDirection spelling.Direction
	sinkSource    settings.Source
	sinkErr       error
)

// TestReadsDoNotAllocate checks the readers of parts and signs, Equal and the
// Same... comparisons, OctaveSub, OctaveDistance (6.13, row 1).
func TestReadsDoNotAllocate(t *testing.T) {
	s := std()
	notes := []note.Note{
		mk(t, spelling.CSharp(), 4, ct(30), s),
		mk(t, spelling.C(), 4, ct(30), nil),
		mk(t, sp(letter.D, 1, 2), -2, cents.MustFromFraction(1, 3), nil),
		mkName(t, spelling.EFlat(), ct(-5), s),
		{},
	}
	reads := map[string]func(n note.Note){
		"IsValid":    func(n note.Note) { sinkBool = n.IsValid() },
		"HasOctave":  func(n note.Note) { sinkBool = n.HasOctave() },
		"Spelling":   func(n note.Note) { sinkSpelling = n.Spelling() },
		"Letter":     func(n note.Note) { sinkLetter = n.Letter() },
		"Alteration": func(n note.Note) { sinkAlt = n.Alteration() },
		"Cents":      func(n note.Note) { sinkCents = n.Cents() },
		"Pitch":      func(n note.Note) { sinkPitch, sinkErr = n.Pitch() },
		"Octave":     func(n note.Note) { sinkOctave, sinkErr = n.Octave() },
		"ASPN":       func(n note.Note) { sinkInt, sinkErr = n.ASPN() },
		"IsNatural":  func(n note.Note) { sinkBool = n.IsNatural() },
		"IsInteger":  func(n note.Note) { sinkBool = n.IsInteger() },
		"Direction":  func(n note.Note) { sinkDirection = n.Direction() },
	}
	for _, n := range notes {
		if !n.HasOctave() {
			continue // the error of a note without octave is not a part
		}
		for name, read := range reads {
			allocs := testing.AllocsPerRun(100, func() { read(n) })
			require.Zero(t, allocs, "%s of %v", name, n.Spelling())
		}
	}
	for _, n := range notes {
		for _, name := range []string{"IsValid", "HasOctave", "Spelling", "Letter", "Alteration", "Cents", "IsNatural", "IsInteger", "Direction"} {
			read := reads[name]
			allocs := testing.AllocsPerRun(100, func() { read(n) })
			require.Zero(t, allocs, "%s of %v", name, n.Spelling())
		}
	}

	pairs := map[string]func(a, b note.Note){
		"Equal":          func(a, b note.Note) { sinkBool = a.Equal(b) },
		"SameRecord":     func(a, b note.Note) { sinkBool = a.SameRecord(b) },
		"SameSpelling":   func(a, b note.Note) { sinkBool = a.SameSpelling(b) },
		"SameLetter":     func(a, b note.Note) { sinkBool = a.SameLetter(b) },
		"SameAlteration": func(a, b note.Note) { sinkBool = a.SameAlteration(b) },
		"SameOctave":     func(a, b note.Note) { sinkBool = a.SameOctave(b) },
		"SameCents":      func(a, b note.Note) { sinkBool = a.SameCents(b) },
	}
	for _, a := range notes {
		for _, b := range notes {
			for name, f := range pairs {
				allocs := testing.AllocsPerRun(100, func() { f(a, b) })
				require.Zero(t, allocs, "%s(%v, %v)", name, a.Spelling(), b.Spelling())
			}
		}
	}

	x, y := notes[0], notes[1]
	allocs := testing.AllocsPerRun(100, func() { sinkInt, sinkErr = x.OctaveSub(y) })
	require.Zero(t, allocs, "OctaveSub")
	allocs = testing.AllocsPerRun(100, func() { sinkInt, sinkErr = x.OctaveDistance(y) })
	require.Zero(t, allocs, "OctaveDistance")
}

// TestCompareDoesNotAllocateForDifferentRecords checks Compare on pairs with
// different records only (6.13, row 2: equal records compare cents, which
// may allocate).
func TestCompareDoesNotAllocateForDifferentRecords(t *testing.T) {
	notes := []note.Note{
		mk(t, spelling.C(), 4, ct(30), nil),
		mk(t, spelling.CSharp(), 4, ct(30), nil),
		mk(t, spelling.C(), 5, ct(0), nil),
		mkName(t, spelling.C(), ct(0), nil),
		mkName(t, spelling.D(), ct(7), nil),
		{},
	}
	for i, a := range notes {
		for j, b := range notes {
			if i == j {
				continue
			}
			allocs := testing.AllocsPerRun(100, func() { sinkInt = a.Compare(b) })
			require.Zero(t, allocs, "Compare(%d, %d)", i, j)
		}
	}
}

// TestSourceReadersDoNotAllocate checks SameSource, IsBound and Source for
// *settings.Settings and a pointer implementation (6.13, row 3).
func TestSourceReadersDoNotAllocate(t *testing.T) {
	user := newCounting(std())
	notes := []note.Note{
		mk(t, spelling.C(), 4, ct(0), std()),
		mk(t, spelling.C(), 4, ct(0), std()),
		mk(t, spelling.C(), 4, ct(0), user),
		mk(t, spelling.C(), 4, ct(0), nil),
	}
	for i, a := range notes {
		allocs := testing.AllocsPerRun(100, func() { sinkBool = a.IsBound() })
		require.Zero(t, allocs, "IsBound %d", i)
		allocs = testing.AllocsPerRun(100, func() { sinkSource = a.Source() })
		require.Zero(t, allocs, "Source %d", i)
		for j, b := range notes {
			allocs = testing.AllocsPerRun(100, func() { sinkBool = a.SameSource(b) })
			require.Zero(t, allocs, "SameSource(%d, %d)", i, j)
		}
	}
}
