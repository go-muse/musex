package system_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// The class and the carry of a name in the traditional system.
func ExampleSystem_Class() {
	sys := system.Traditional()

	fmt.Println(sys.MustClass(spelling.BSharp()))
	fmt.Println(sys.MustClass(spelling.CFlat()))
	fmt.Println(sys.MustCarry(spelling.BSharp()), sys.MustCarry(spelling.CFlat()))
	// Output:
	// 0 mod 12
	// 11 mod 12
	// 1 -1
}

// "B♯3" is written in octave 3 but lies in Octave 4, on the position of C4.
func ExampleSystem_FromWritten() {
	sys := system.Traditional()

	bs3 := sys.MustFromWritten(spelling.BSharp(), octave.FromASPN(3))
	fmt.Println(bs3)
	fmt.Println(sys.MustWrittenOctave(bs3).MustASPN())
	fmt.Println(sys.MustEquivalent(bs3, pitch.MustFromASPN(spelling.C(), 4)))
	// Output:
	// B+1 ASPN 4
	// 3
	// true
}

// RaiseWritten keeps the written register: B3 becomes B♯3. For comparison,
// pitch.Pitch.Raise keeps the Octave and gives B♯2.
func ExampleSystem_RaiseWritten() {
	sys := system.Traditional()
	b3 := sys.MustFromWritten(spelling.B(), octave.FromASPN(3))

	r := sys.MustRaiseWritten(b3)
	fmt.Println(r, "written", sys.MustWrittenOctave(r).MustASPN(), "steps", sys.MustDistance(b3, r))

	p := b3.MustRaise()
	fmt.Println(p, "written", sys.MustWrittenOctave(p).MustASPN(), "steps", sys.MustDistance(b3, p))
	// Output:
	// B+1 ASPN 4 written 3 steps 1
	// B+1 ASPN 3 written 2 steps -11
}

// A name for a class is chosen only by an explicit rule.
func ExampleSystem_Spell() {
	sys := system.Traditional()
	cs := pitchclass.MustNew(1, 12)
	ds := pitchclass.MustNew(3, 12)

	fmt.Println(sys.MustSpell(cs, system.Sharps()))
	fmt.Println(sys.MustSpell(cs, system.Flats()))

	up := system.MustChromaticScale(spelling.C(), system.Major, system.Ascending)
	down := system.MustChromaticScale(spelling.C(), system.Major, system.Descending)
	fmt.Println(sys.MustSpell(ds, up), sys.MustSpell(ds, down))
	// Output:
	// C+1
	// D-1
	// D+1 E-1
}

// The distance in steps between two written pitches.
func ExampleSystem_Distance() {
	sys := system.Traditional()
	c4 := pitch.MustFromASPN(spelling.C(), 4)
	e5 := pitch.MustFromASPN(spelling.E(), 5)

	fmt.Println(sys.MustDistance(c4, e5), sys.MustDistance(e5, c4))

	bs3 := sys.MustFromWritten(spelling.BSharp(), octave.FromASPN(3))
	fmt.Println(sys.MustDistance(bs3, c4))
	// Output:
	// 16 -16
	// 0
}

// A user system with 14 classes and equal steps between letters.
func ExampleNew() {
	s, err := system.New(system.Definition{
		Name:        "14 equal steps",
		Classes:     14,
		Boundary:    letter.C,
		Offsets:     [letter.Count]int{0, 2, 4, 6, 8, 10, 12},
		Chromatic:   1,
		Alterations: system.Integers(),
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(s.Classes(), s.Steps())
	fmt.Println(s.MustClass(spelling.ESharp()), "and", s.MustClass(spelling.FFlat()))
	fmt.Println(s.MustRegularity().Regular)
	// Output:
	// 14 [2 2 2 2 2 2 2]
	// 5 mod 14 and 5 mod 14
	// false
}

// The admissible alteration values of the quarter-tone rule up to one sharp.
func ExampleRule_Values() {
	for v := range system.MustMultiples(2).MustValues(alteration.Sharp()) {
		fmt.Println(v)
	}
	// Output:
	// -1
	// -1/2
	// 0
	// +1/2
	// +1
}

// Errors are matched with errors.Is; an invalid definition also matches its cause.
func ExampleNew_error() {
	_, err := system.New(system.Definition{
		Classes: 0, Boundary: letter.C, Chromatic: 1, Alterations: system.Integers(),
	})
	fmt.Println(errors.Is(err, system.ErrDefinition), errors.Is(err, system.ErrClassCount), errors.Is(err, system.ErrOffsets))

	dHalf := spelling.MustNew(letter.D, alteration.MustFromFraction(1, 2))
	_, err = system.Traditional().Class(dHalf)
	fmt.Println(errors.Is(err, system.ErrInadmissible))
	// Output:
	// true true false
	// true
}
