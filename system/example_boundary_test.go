package system_test

import (
	"fmt"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// A user system with five classes where B is on offset N = 5: its natural name has
// carry 1, so B written in octave 4 lies in Octave 5, on the position of C5.
func ExampleNew_boundaryLetter() {
	s5 := system.MustNew(system.Definition{
		Name:        "five",
		Classes:     5,
		Boundary:    letter.C,
		Offsets:     [letter.Count]int{0, 1, 2, 2, 3, 4, 5},
		Chromatic:   1,
		Alterations: system.Integers(),
	})

	fmt.Println(s5.MustClass(spelling.B()), s5.MustCarry(spelling.B()))

	b4 := s5.MustFromWritten(spelling.B(), octave.FromASPN(4))
	c5 := pitch.MustFromASPN(spelling.C(), 5)
	fmt.Println(b4)
	fmt.Println(s5.MustEquivalent(b4, c5), s5.MustDistance(b4, c5))
	// Output:
	// 0 mod 5 1
	// B ASPN 5
	// true 0
}
