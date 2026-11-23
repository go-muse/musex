package note_test

import (
	"fmt"
	"strings"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/tuning"
)

func ExampleNotes_Clone() {
	c := note.MustFromSpelling(spelling.C(), cents.Zero(), nil)
	d := note.MustFromSpelling(spelling.D(), cents.Zero(), nil)
	ns := note.Notes{c, c, d}
	independent := ns.Clone()
	independent[0] = d
	fmt.Println(ns[0].Spelling().Equal(spelling.C()))
	fmt.Println(independent[0].Spelling().Equal(spelling.D()))
	var empty note.Notes
	fmt.Println(empty.Clone() == nil)
	// Output:
	// true
	// true
	// true
}

func ExampleNotes_SortedBySoundInTuning() {
	c := note.MustFromASPN(spelling.C(), 4, cents.MustFromInt(150), nil)
	cs := note.MustFromASPN(spelling.CSharp(), 4, cents.Zero(), nil)
	ns := note.Notes{c, cs, c}
	ordered, err := ns.SortedBySoundInTuning(tuning.Standard())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ordered[0].Spelling().Equal(spelling.CSharp()))
	fmt.Println(len(ordered))
	s := scale.MustNew(ordered...)
	ordered[0] = c
	fmt.Println(s.MustAt(0).Spelling().Equal(spelling.CSharp()))
	// Output:
	// true
	// 3
	// true
}

func ExampleNotes_SortedByPosition() {
	_, err := (note.Notes{{}}).SortedByPosition()
	fmt.Println(err != nil)
	var empty note.Notes
	ordered, err := empty.SortedByPosition()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ordered != nil, len(ordered))
	// Output:
	// true
	// true 0
}

// UniqueBySpelling keeps the first note of each written name: C5 and C4
// collapse, while the enharmonic spellings C♯4 and D♭4 stay distinct.
func ExampleNotes_UniqueBySpelling() {
	std := settings.Standard()
	ns := note.Notes{
		note.MustFromASPN(spelling.C(), 5, cents.Zero(), std),
		note.MustFromASPN(spelling.C(), 4, cents.Zero(), std),
		note.MustFromASPN(spelling.CSharp(), 4, cents.Zero(), std),
		note.MustFromASPN(spelling.DFlat(), 4, cents.Zero(), std),
	}
	for _, n := range ns.MustUniqueBySpelling() {
		fmt.Println(n.MustFormatIn(std))
	}
	// Output:
	// C5
	// C♯4
	// D♭4
}

// Spellings projects the names in order, including repetitions, without a
// tuning or a source lookup.
func ExampleNotes_Spellings() {
	ns := note.Notes{
		note.MustFromASPN(spelling.C(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.EFlat(), 4, cents.Zero(), nil),
		note.MustFromASPN(spelling.C(), 5, cents.Zero(), nil),
	}
	names := ns.MustSpellings()
	fmt.Println(len(names))
	fmt.Println(names[0].Equal(spelling.C()), names[1].Equal(spelling.EFlat()), names[2].Equal(spelling.C()))
	// Output:
	// 3
	// true true true
}

// The descending sort is stable: C♯4 and D♭4 sound the same in 12-tone equal
// temperament and keep their input order below G4.
func ExampleNotes_SortedBySoundDescending() {
	std := settings.Standard()
	ns := note.Notes{
		note.MustFromASPN(spelling.CSharp(), 4, cents.Zero(), std),
		note.MustFromASPN(spelling.DFlat(), 4, cents.Zero(), std),
		note.MustFromASPN(spelling.G(), 4, cents.Zero(), std),
	}
	for _, n := range ns.MustSortedBySoundDescending() {
		fmt.Println(n.MustFormatIn(std))
	}
	// Output:
	// G4
	// C♯4
	// D♭4
}

// FormatIn renders every note in the notation of one explicit source, with
// written octaves where present.
func ExampleNotes_FormatIn() {
	std := settings.Standard()
	ns := note.Notes{
		note.MustFromASPN(spelling.C(), 4, cents.Zero(), std),
		note.MustFromASPN(spelling.EFlat(), 4, cents.Zero(), std),
		note.MustFromASPN(spelling.G(), 4, cents.Zero(), std),
	}
	fmt.Println(strings.Join(ns.MustFormatIn(std), " "))
	// Output:
	// C4 E♭4 G4
}

// FormatNamesIn prints the names only, for notes with or without an octave.
func ExampleNotes_FormatNamesIn() {
	std := settings.Standard()
	ns := note.Notes{
		note.MustFromASPN(spelling.C(), 4, cents.Zero(), std),
		note.MustFromASPN(spelling.EFlat(), 4, cents.Zero(), std),
		note.MustFromASPN(spelling.G(), 4, cents.Zero(), std),
	}
	fmt.Println(strings.Join(ns.MustFormatNamesIn(std), " "))
	// Output:
	// C E♭ G
}
