package note_test

import (
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/scale"
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
