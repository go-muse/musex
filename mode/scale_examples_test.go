package mode_test

import (
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
)

func ExampleMode_FromTonic() {
	ionianMode := mode.MustFromCode("ionian")
	a := note.MustFromSpelling(spelling.A(), cents.Zero(), nil)
	s := ionianMode.MustFromTonic(a)
	for _, n := range s.Notes() {
		fmt.Println(n.Spelling())
	}
	// Output:
	// A
	// B
	// C+1
	// D
	// E
	// F+1
	// G+1
}
