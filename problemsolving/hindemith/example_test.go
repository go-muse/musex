package hindemith_test

import (
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/problemsolving/hindemith"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// The first inversion of the C major triad is group I, subgroup 2: the root
// C lies above the bass E.
func ExampleAnalyzeChord() {
	sys := system.Traditional()
	std := settings.Standard()
	c := hindemith.MustAnalyzeChord(sys, []note.Note{
		note.MustFromASPN(spelling.E(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.G(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.C(), 5, cents.Zero(), nil),
	})
	fmt.Println(c.Group(), c.Subgroup(), c.Root().MustFormatIn(std))
	// Output:
	// I 2 C5
}

// Two voicings of the C major triad: in root position the root C is the bass
// (subgroup 1), in the first inversion C lies above the bass E (subgroup 2).
func ExampleAnalyzeChord_voicings() {
	sys := system.Traditional()
	std := settings.Standard()
	voicings := [][]note.Note{
		{
			note.MustFromASPN(spelling.C(), 4, cents.Zero(), nil),
			note.MustFromASPN(spelling.E(), 4, cents.Zero(), nil),
			note.MustFromASPN(spelling.G(), 4, cents.Zero(), nil),
		},
		{
			note.MustFromASPN(spelling.E(), 4, cents.Zero(), nil),
			note.MustFromASPN(spelling.G(), 4, cents.Zero(), nil),
			note.MustFromASPN(spelling.C(), 5, cents.Zero(), nil),
		},
	}
	for _, ns := range voicings {
		c := hindemith.MustAnalyzeChord(sys, ns)
		fmt.Println(c.Group(), c.Subgroup(), c.Value(), c.Root().Spelling(), c.Bass().MustFormatIn(std))
	}
	// Output:
	// I 1 1 C C4
	// I 2 2 C E4
}

// The tritone has no root of its own.
func ExampleAnalyzeDyad() {
	sys := system.Traditional()
	d := hindemith.MustAnalyzeDyad(sys,
		note.MustFromASPN(spelling.C(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.FSharp(), 4, cents.Zero(), nil))
	fmt.Println(d.Series(), d.Name(), d.HasRoot())
	// Output:
	// 13 tritone false
}

// Fluctuation rises from the root-position triad to its inversion and to
// the dominant seventh, then falls back.
func ExampleAnalyzeProgression() {
	sys := system.Traditional()
	triad := []note.Note{
		note.MustFromASPN(spelling.C(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.E(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.G(), 4, cents.Zero(), nil),
	}
	inversion := []note.Note{triad[1], triad[2], note.MustFromASPN(spelling.C(), 5, cents.Zero(), nil)}
	seventh := append(append([]note.Note(nil), triad...), note.MustFromASPN(spelling.BFlat(), 4, cents.Zero(), nil))
	p := hindemith.MustAnalyzeProgression(sys, [][]note.Note{triad, inversion, seventh, triad})
	for _, s := range p.Steps() {
		fmt.Println(s.Direction())
	}
	// Output:
	// rise
	// rise
	// fall
}
