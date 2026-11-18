package scale_test

import (
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/scale"
	"github.com/go-muse/musex/spelling"
)

func ExampleScale_Next() {
	g := note.MustFromASPN(spelling.G(), 4, cents.Zero(), nil)
	a := note.MustFromASPN(spelling.A(), 4, cents.Zero(), nil)
	b := note.MustFromASPN(spelling.B(), 4, cents.Zero(), nil)
	s := scale.MustNew(g, a, b)
	next := s.MustNext(2)
	fmt.Println(next.Index, next.Note.Spelling(), next.Note.MustASPN())
	// Output: 0 G 4
}

func ExampleScale_Cycle() {
	g := note.MustFromSpelling(spelling.G(), cents.Zero(), nil)
	a := note.MustFromSpelling(spelling.A(), cents.Zero(), nil)
	s := scale.MustNew(g, a)
	snapshot := s.MustCycle(0, scale.Forward)
	s.MustClear()
	count := 0
	for n := range snapshot {
		fmt.Println(n.Spelling())
		count++
		if count == 3 {
			break
		}
	}
	// Output:
	// G
	// A
	// G
}
