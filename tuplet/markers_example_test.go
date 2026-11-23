package tuplet_test

import (
	"fmt"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"
)

// UnapplyChain recovers the written duration from a sounding one.
func ExampleUnapplyChain() {
	d, err := tuplet.UnapplyChain(fraction.MustNewReduced(1, 15), tuplet.Quintuplet(), tuplet.Triplet())
	fmt.Println(d, err)
	// Output:
	// 1/8 <nil>
}

// ABC markers read the short form with the defaults of the meter context.
func ExampleParseABCMarker() {
	simple := tuplet.MustParseABCMarker("(5", tuplet.Binary)
	compound := tuplet.MustParseABCMarker("(5", tuplet.Ternary)
	scoped := tuplet.MustParseABCMarker("(3:2:2", tuplet.Binary)
	fmt.Println(simple.Tuplet, simple.Events)
	fmt.Println(compound.Tuplet, compound.Events)
	fmt.Println(scoped.Tuplet, scoped.Events)
	// Output:
	// 5:2 5
	// 5:3 5
	// 3:2 2
}

// LilyPond group markers and duration multipliers are separate operations.
func ExampleParseLilyPondTuplet() {
	t := tuplet.MustParseLilyPondTuplet(`\tuplet 3/2`)
	fmt.Println(t)
	// Output:
	// 3:2
}

func ExampleParseLilyPondMultiplier() {
	c := tuplet.MustParseLilyPondMultiplier("*2/3")
	fmt.Println(c)
	// Output:
	// 2/3
}
