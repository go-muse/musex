package pitch_test

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
)

func Example() {
	// Create F# in octave 3 of the Franco-Belgian numbering (which is ASPN 4).
	oct := octave.MustFromNumber(3, number.FrancoBelgian())
	fSharp := pitch.MustNew(spelling.FSharp(), oct)

	fmt.Println(fSharp)
	fmt.Println(fSharp.Octave().Name(naming.Russian()))
	fmt.Println(fSharp.MustUp())

	for p := range pitch.MustRange(spelling.FSharp(), octave.FromASPN(2), octave.FromASPN(4)) {
		fmt.Println(p)
	}
	// Output:
	// F+1 ASPN 4
	// первая
	// F+1 ASPN 5
	// F+1 ASPN 2
	// F+1 ASPN 3
	// F+1 ASPN 4
}

func ExampleNew() {
	pASPN, err := pitch.New(spelling.CSharp(), octave.FromASPN(4))
	fmt.Println(pASPN, err == nil)

	pFB := pitch.MustNew(spelling.CSharp(), octave.MustFromNumber(3, number.FrancoBelgian()))
	pRu := pitch.MustNew(spelling.CSharp(), octave.MustFromName("первая", naming.Russian()))
	fmt.Println(pASPN.Equal(pFB), pASPN == pRu)

	_, err = pitch.New(spelling.Spelling{}, octave.FromASPN(4))
	fmt.Println(errors.Is(err, pitch.ErrInvalid))
	// Output:
	// C+1 ASPN 4 true
	// true true
	// true
}

func ExampleFromASPN() {
	p, err := pitch.FromASPN(spelling.BFlat(), 3)
	fmt.Println(p, err == nil)

	s, o := p.Parts()
	fmt.Println(p.Spelling() == s, p.Octave() == o, p.MustASPN())
	// Output:
	// B-1 ASPN 3 true
	// true true 3
}

func ExamplePitch_IsValid() {
	valid := pitch.MustFromASPN(spelling.C(), 4)
	var zero pitch.Pitch

	fmt.Println(valid.IsValid(), zero.IsValid())
	// Output:
	// true false
}

func ExamplePitch_predicates() {
	c4 := pitch.MustFromASPN(spelling.C(), 4)
	cSharp4 := pitch.MustFromASPN(spelling.CSharp(), 4)
	dFlat4 := pitch.MustFromASPN(spelling.DFlat(), 4)

	fmt.Println(c4.IsNatural(), c4.IsInteger(), c4.Direction())
	fmt.Println(cSharp4.IsNatural(), cSharp4.IsInteger(), cSharp4.Direction())
	fmt.Println(cSharp4.SameSpelling(pitch.MustFromASPN(spelling.CSharp(), 5)))
	fmt.Println(cSharp4.SameOctave(dFlat4))
	fmt.Println(cSharp4.SameLetter(c4))
	fmt.Println(cSharp4.SameAlteration(pitch.MustFromASPN(spelling.FSharp(), 2)))
	// Output:
	// true true natural
	// false true raised
	// true
	// true
	// true
	// true
}

func ExamplePitch_Equal() {
	cSharp4 := pitch.MustFromASPN(spelling.CSharp(), 4)
	dFlat4 := pitch.MustFromASPN(spelling.DFlat(), 4)
	bSharp4 := pitch.MustFromASPN(spelling.BSharp(), 4)
	c4 := pitch.MustFromASPN(spelling.C(), 4)

	// Written pitches preserve both spelling and positional octave.
	fmt.Println(cSharp4.Equal(dFlat4), bSharp4.Equal(c4))

	counts := map[pitch.Pitch]int{
		cSharp4: 1,
		dFlat4:  2,
	}
	counts[pitch.MustNew(spelling.CSharp(), octave.FromASPN(4))]++
	fmt.Println(counts[cSharp4], counts[dFlat4])
	// Output:
	// false false
	// 2 2
}

func ExamplePitch_Compare() {
	pitches := []pitch.Pitch{
		pitch.MustFromASPN(spelling.BSharp(), 4),
		pitch.MustFromASPN(spelling.DFlat(), 4),
		pitch.MustFromASPN(spelling.C(), 4),
		pitch.MustFromASPN(spelling.B(), 3),
		pitch.MustFromASPN(spelling.CSharp(), 4),
	}
	slices.SortFunc(pitches, pitch.Pitch.Compare)

	for _, p := range pitches {
		fmt.Println(p)
	}
	// Output:
	// B ASPN 3
	// C ASPN 4
	// C+1 ASPN 4
	// D-1 ASPN 4
	// B+1 ASPN 4
}

func ExamplePitch_WithSpelling() {
	p := pitch.MustFromASPN(spelling.C(), 4)
	fmt.Println(p.MustWithSpelling(spelling.DFlat()))

	// Apply a quarter-tone (+1/2) shift via spelling.Spelling.
	half := alteration.MustFromFraction(1, 2)
	quarterSharp := p.MustWithSpelling(p.Spelling().MustRaiseByAlteration(half))
	fmt.Println(quarterSharp)
	fmt.Println(quarterSharp.MustWithOctave(octave.FromASPN(5)))
	fmt.Println(quarterSharp.MustWithASPN(2))
	// Output:
	// D-1 ASPN 4
	// C+1/2 ASPN 4
	// C+1/2 ASPN 5
	// C+1/2 ASPN 2
}

func ExamplePitch_Natural() {
	fmt.Println(pitch.MustFromASPN(spelling.CSharp(), 4).Natural())
	fmt.Println(pitch.MustFromASPN(spelling.BSharp(), 4).Natural())

	var zero pitch.Pitch
	fmt.Println(zero.Natural())
	// Output:
	// C ASPN 4
	// B ASPN 4
	// Pitch(invalid)
}

func ExamplePitch_Raise() {
	b4 := pitch.MustFromASPN(spelling.B(), 4)
	// Raising B in ASPN 4 yields B# in ASPN 4 (letter and positional octave are preserved).
	fmt.Println(b4.MustRaise())
	fmt.Println(b4.MustRaiseBy(2))
	fmt.Println(b4.MustLower())
	fmt.Println(b4.MustLowerBy(2))

	_, err := b4.RaiseBy(alteration.MaxMagnitude + 1)
	fmt.Println(errors.Is(err, alteration.ErrOutOfRange))
	// Output:
	// B+1 ASPN 4
	// B+2 ASPN 4
	// B-1 ASPN 4
	// B-2 ASPN 4
	// true
}

func ExamplePitch_Shift() {
	fSharp4 := pitch.MustFromASPN(spelling.FSharp(), 4)
	fmt.Println(fSharp4.MustUp())
	fmt.Println(fSharp4.MustDown())
	fmt.Println(fSharp4.MustShift(2))
	fmt.Println(fSharp4.MustShiftDown(3))

	_, err := pitch.MustFromASPN(spelling.C(), math.MaxInt).Up()
	fmt.Println(errors.Is(err, octave.ErrOutOfRange))
	// Output:
	// F+1 ASPN 5
	// F+1 ASPN 3
	// F+1 ASPN 6
	// F+1 ASPN 1
	// true
}

func ExamplePitch_OctaveSub() {
	c5 := pitch.MustFromASPN(spelling.C(), 5)
	gSharp2 := pitch.MustFromASPN(spelling.GSharp(), 2)
	fmt.Println(c5.MustOctaveSub(gSharp2))
	fmt.Println(gSharp2.MustOctaveSub(c5))
	fmt.Println(c5.MustOctaveDistance(gSharp2))

	// Any two pitches in the same positional octave have octave difference and distance 0.
	c4 := pitch.MustFromASPN(spelling.C(), 4)
	bSharp4 := pitch.MustFromASPN(spelling.BSharp(), 4)
	fmt.Println(c4.MustOctaveSub(bSharp4), c4.MustOctaveDistance(bSharp4))
	// Output:
	// 3
	// -3
	// 3
	// 0 0
}

func ExampleRange() {
	for p := range pitch.MustRange(spelling.FSharp(), octave.FromASPN(4), octave.FromASPN(2)) {
		fmt.Println(p)
	}

	// Lazy iteration over the full int range with early break.
	count := 0
	for p := range pitch.MustRange(spelling.C(), octave.FromASPN(math.MinInt), octave.FromASPN(math.MaxInt)) {
		fmt.Println(p.MustASPN() == math.MinInt+count)
		count++
		if count == 2 {
			break
		}
	}
	// Output:
	// F+1 ASPN 4
	// F+1 ASPN 3
	// F+1 ASPN 2
	// true
	// true
}

func ExamplePitch_String() {
	p := pitch.MustFromASPN(spelling.CSharp(), 4)
	var zero pitch.Pitch

	fmt.Printf("%v | %s | %#v\n", p, p, p)
	fmt.Printf("%v | %#v\n", zero, zero)
	// Output:
	// C+1 ASPN 4 | C+1 ASPN 4 | pitch.MustFromASPN(spelling.MustParse("C+1"), 4)
	// Pitch(invalid) | pitch.Pitch{}
}
