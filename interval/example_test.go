package interval_test

import (
	"errors"
	"fmt"
	"slices"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// Executable examples of the scenarios S-1...S-10 (business-spec 5.1; NF-6;
// architecture 4.14).

// S-1: a predefined interval or one made from size and quality; parts and
// properties.
func ExampleMajorThird() {
	maj3 := interval.MajorThird()
	fmt.Println(maj3, maj3.Size(), maj3.Quality())
	fmt.Println(maj3.IsSimple(), maj3.IsDiatonic(), maj3.IsNarrow())

	dd5 := interval.MustNewMultiple(2, interval.Diminished, 5)
	fmt.Println(dd5, dd5.Multiplicity(), dd5.Deviation())

	tritone := interval.AugmentedFourth()
	fmt.Println(tritone.IsTritone(), tritone.IsBasic(), interval.AugmentedSecond().IsChromatic())
	// Output:
	// M3 3 major
	// true true true
	// dd5 2 -2
	// true true true
}

// S-2: the interval between two names in a direction, and both of them.
func ExampleBothWays() {
	up, down := interval.MustBothWays(spelling.C(), spelling.E())
	fmt.Println(up, down)

	_, err := interval.Between(spelling.C(), spelling.E(), interval.DirectionUnset)
	fmt.Println(errors.Is(err, interval.ErrNoDirection))

	fmt.Println(interval.MustBetween(spelling.C(), spelling.CSharp(), interval.Up))
	fmt.Println(interval.MustBetween(spelling.C(), spelling.CSharp(), interval.Down))
	// Output:
	// M3 -m6
	// true
	// A1
	// -d8
}

// S-3: the interval between two written pitches in a system: direction by
// writing, size and quality.
func ExampleBetweenPitches() {
	sys := system.Traditional()
	c4 := pitch.MustFromASPN(spelling.C(), 4)
	bs3 := sys.MustFromWritten(spelling.BSharp(), octave.FromASPN(3)) // "B♯3", Octave 4

	fmt.Println(interval.MustBetweenPitches(sys, bs3, c4)) // up by writing, same position
	fmt.Println(interval.MustBetweenPitches(sys, c4, pitch.MustFromASPN(spelling.E(), 5)))
	fmt.Println(interval.MustBetweenPitches(sys, pitch.MustFromASPN(spelling.A(), 3), pitch.MustFromASPN(spelling.C(), 3)))
	// Output:
	// d2
	// M10
	// -M6
}

// S-4: transposing a name, a written pitch and a note up and down, with no
// enharmonic substitution.
func ExampleTranspose() {
	fmt.Println(interval.MustTranspose(spelling.E(), interval.MajorThird()))
	fmt.Println(interval.MustTranspose(spelling.BSharp(), interval.AugmentedSixth()))
	fmt.Println(interval.MustTranspose(spelling.F(), interval.AugmentedFourth().MustDown()))

	sys := system.Traditional()
	c4 := pitch.MustFromASPN(spelling.C(), 4)
	fmt.Println(interval.MustTransposePitch(sys, c4, interval.MajorTenth()))
	fmt.Println(interval.MustTransposePitch(sys, c4, interval.DiminishedSecond().MustDown()))
	// Output:
	// G+1
	// G+3
	// C-1
	// E ASPN 5
	// B+1 ASPN 4
}

// S-4: a note keeps its correction and source; the input note does not
// change.
func ExampleTransposeNote() {
	src := settings.Standard()
	x := note.MustFromPitch(pitch.MustFromASPN(spelling.C(), 4), cents.MustFromInt(30), src)

	y := interval.MustTransposeNote(x, interval.MajorThird())
	fmt.Println(y.Spelling(), y.SameCents(x), y.SameSource(x))
	fmt.Println(x.Spelling())
	fmt.Println(interval.MustBetweenNotes(x, y, interval.DirectionUnset))
	// Output:
	// E true true
	// C
	// M3
}

// S-5: inversion, simple part and octaves, adding octaves.
func ExampleInterval_Invert() {
	fmt.Println(interval.MajorThird().MustInvert())
	fmt.Println(interval.MajorThird().MustDown().MustInvert())

	m10 := interval.MajorTenth()
	fmt.Println(m10.MustSimple(), m10.Octaves())
	fmt.Println(interval.MajorThird().MustAddOctaves(2))

	_, err := interval.AugmentedOctave().Invert()
	fmt.Println(errors.Is(err, interval.ErrNoInversion))
	// Output:
	// m6
	// -m6
	// M3 1
	// M17
	// true
}

// S-6: sum and difference; the steps of a scale template add up to P8.
func ExampleInterval_Add() {
	fmt.Println(interval.MajorThird().MustAdd(interval.MinorThird()))

	sum := interval.PerfectUnison()
	for _, step := range []interval.Interval{
		interval.MajorSecond(), interval.MajorSecond(), interval.MinorSecond(), interval.MajorSecond(),
		interval.MajorSecond(), interval.MajorSecond(), interval.MinorSecond(),
	} {
		sum = sum.MustAdd(step)
	}

	fmt.Println(sum, sum.Equal(interval.PerfectOctave()))
	fmt.Println(interval.MajorNinth().MustSub(interval.PerfectOctave()))
	fmt.Println(interval.PerfectFifth().MustMul(12).MustSub(interval.PerfectOctave().MustMul(7)))
	// Output:
	// P5
	// P8 true
	// M2
	// -d2
}

// S-7: positional size, enharmonic equality, intervals of a size, quality
// by size, respellings.
func ExampleInterval_Steps() {
	sys := system.Traditional()
	a4, d5 := interval.AugmentedFourth(), interval.DiminishedFifth()

	fmt.Println(a4.MustSteps(sys, spelling.Spelling{}))
	fmt.Println(interval.MustEnharmonicallyEqual(sys, a4, spelling.Spelling{}, d5, spelling.Spelling{}))
	fmt.Println(interval.MustEnharmonicallyEqual(system.Positions19(), a4, spelling.Spelling{}, d5, spelling.Spelling{}))
	fmt.Println(interval.MustFromSteps(sys, 4, 6, spelling.Spelling{}))
	fmt.Println(slices.Collect(interval.MustAllFromSteps(sys, 6, 2, spelling.Spelling{})))

	c4, e4 := pitch.MustFromASPN(spelling.C(), 4), pitch.MustFromASPN(spelling.E(), 4)
	for r := range interval.MustRespellings(sys, c4, e4, alteration.DoubleSharp(), interval.ReplaceTo) {
		fmt.Println(r.To, r.Interval)
	}

	irregular := system.MustNew(system.Definition{
		Name: "irregular", Classes: 12, Boundary: letter.C,
		Offsets: [letter.Count]int{0, 3, 4, 5, 7, 9, 11}, Chromatic: 1, Alterations: system.Integers(),
	})
	_, err := interval.MajorSecond().Steps(irregular, spelling.Spelling{})
	fmt.Println(errors.Is(err, interval.ErrStartRequired))
	fmt.Println(interval.MajorSecond().MustSteps(irregular, spelling.C()), interval.MajorSecond().MustSteps(irregular, spelling.D()))
	// Output:
	// 6
	// true
	// false
	// A4
	// [AA3 A4 d5 dd6]
	// D+2 ASPN 4 AA2
	// E ASPN 4 M3
	// F-1 ASPN 4 d4
	// true
	// 3 1
}

// S-8: short notation and names, output and parsing.
func ExampleParse() {
	fmt.Println(interval.MustParse("aug4"))
	fmt.Println(interval.MustParse("-m6").IsDescending())

	_, err := interval.Parse("TT")
	fmt.Println(errors.Is(err, interval.ErrParse))

	dd5 := interval.DoublyDiminishedFifth()
	name, ok := dd5.MustName()
	fmt.Println(name, ok)

	ru, _ := dd5.MustRussianName()
	abbr, _ := dd5.MustRussianAbbreviation()
	fmt.Println(ru, abbr)

	_, ok = interval.MustNewMultiple(4, interval.Augmented, 4).MustName()
	fmt.Println(ok)
	fmt.Println(interval.MustParseName("perfect 8ve"), interval.MustParseName("ув.4"))
	// Output:
	// A4
	// true
	// true
	// doubly diminished fifth true
	// дважды уменьшенная квинта дв.ум.5
	// false
	// P8 A4
}

// S-9: the named inversion of a name and a pitch; places on the line of
// fifths.
func ExampleReflect() {
	fmt.Println(interval.MustReflect(spelling.D(), spelling.FSharp()))
	fmt.Println(interval.MustFifthsOf(spelling.FSharp()), interval.MustFifthsBetween(spelling.F(), spelling.B()))
	fmt.Println(interval.MustFromFifths(-6))

	_, err := interval.FifthsOf(spelling.MustNew(letter.E, alteration.MustFromFraction(-1, 2)))
	fmt.Println(errors.Is(err, interval.ErrNotOnFifths))

	sys := system.Traditional()
	fmt.Println(interval.MustReflectPitch(sys, pitch.MustFromASPN(spelling.D(), 4), pitch.MustFromASPN(spelling.E(), 5)))
	// Output:
	// B-1
	// 6 6
	// d5
	// true
	// C ASPN 3
}

// S-10: fractional alterations: C → E½♭ is a neutral third; transposing by
// it; its size in the 24-quarter-tone system.
func ExampleFromDeviation() {
	eHalfFlat := spelling.MustNew(letter.E, alteration.MustFromFraction(-1, 2))
	n3 := interval.MustBetween(spelling.C(), eHalfFlat, interval.Up)
	fmt.Println(n3, n3.Quality())
	fmt.Println(interval.MustTranspose(spelling.C(), n3))
	fmt.Println(n3.MustSteps(system.QuarterTones24(), spelling.Spelling{}))

	_, err := n3.Steps(system.Traditional(), spelling.Spelling{})
	fmt.Println(errors.Is(err, interval.ErrFractionalSteps))

	fmt.Println(interval.MustFromSteps(system.QuarterTones24(), 3, 7, spelling.Spelling{}))
	fmt.Println(interval.MustFromDeviation(3, alteration.MustFromFraction(1, 4)))
	// Output:
	// n3 neutral
	// E-1/2
	// 7
	// true
	// n3
	// {+1/4}3
}
