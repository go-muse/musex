package cents_test

import (
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/go-muse/musex/cents"
)

// How much wider is the pure fifth 3:2 than the equal-tempered fifth, and
// how far apart are 440 Hz and 442 Hz?
func Example() {
	fifth := cents.MustFromRatio(3, 2)
	equal := cents.Semitone().MustMul(7)
	fmt.Println(fifth.MustSub(equal).Decimal(3, 0))
	fmt.Println(cents.MustBetween(440, 442).Decimal(2, 0))
	// Output:
	// 1.955
	// 7.85
}

func ExampleOctave() {
	fmt.Println(cents.Zero(), cents.Semitone(), cents.Octave())
	fmt.Println(cents.Octave() == cents.MustFromRatio(2, 1))
	fmt.Println(cents.Octave().Equal(cents.MustStep(12, 12)))
	// Output:
	// 0 100 1200
	// true
	// true
}

func ExampleFromFloat() {
	fmt.Println(cents.MustFromFloat(7.1))        // exact: up to 6 digits
	fmt.Println(cents.MustFromFloat(1.23456789)) // the grid of 10^-6 cent
	fmt.Println(cents.MustFromFloat(5e-7))       // half a step of the grid, away from zero
	// Output:
	// 7.1
	// 1.234568
	// 0.000001
}

func ExampleStep() {
	fmt.Println(cents.MustStep(7, 12))
	step := cents.MustStep(1, 19)
	fmt.Println(step)
	fmt.Println(step.Decimal(3, 0))
	// Output:
	// 700
	// 1200/19 cents
	// 63.158
}

func ExampleFromRatio() {
	fifth := cents.MustFromRatio(3, 2)
	fmt.Println(fifth)
	fmt.Println(fifth.Decimal(6, 0))
	fmt.Println(cents.MustFromRatio(6, 4) == fifth)
	// Output:
	// 3:2
	// 701.955001
	// true
}

func ExampleFromRatioFloat() {
	approx := cents.MustFromRatioFloat(1.5)
	fmt.Println(approx.Decimal(6, 0))
	fmt.Println(approx.Equal(cents.MustFromRatio(3, 2)))
	// Output:
	// 701.955001
	// false
}

func ExampleBetween() {
	fmt.Println(cents.MustBetween(440, 442))
	fmt.Println(cents.MustBetween(442, 440))
	// Output:
	// 7.851415
	// -7.851415
}

func Example_namedValues() {
	comma := cents.SyntonicComma()
	fmt.Println(comma)
	fmt.Println(comma.Decimal(6, 0))
	fmt.Println(cents.Schisma() == cents.PythagoreanComma().MustSub(comma))
	fmt.Println(len(cents.NamedValues()))
	// Output:
	// 81:80
	// 21.506290
	// true
	// 19
}

func ExampleCents_Add() {
	fmt.Println(cents.MustFromRatio(3, 2).MustAdd(cents.MustFromRatio(4, 3)))
	fmt.Println(cents.MustFromRatio(5, 4).MustAdd(cents.MustFromRatio(6, 5)))
	// Output:
	// 1200
	// 3:2
}

func ExampleCents_Mul() {
	comma := cents.MustFromRatio(3, 2).MustMul(12).MustSub(cents.Octave().MustMul(7))
	fmt.Println(comma)
	fmt.Println(comma == cents.PythagoreanComma())
	// Output:
	// 531441:524288
	// true
}

func ExampleCents_Div() {
	meantone := cents.MustFromRatio(3, 2).MustSub(cents.SyntonicComma().MustDiv(4))
	fmt.Println(meantone)
	fmt.Println(meantone.Decimal(6, 0))
	// Output:
	// 5^(1/4)
	// 696.578428
}

func ExampleCents_Compare() {
	values := []cents.Cents{
		cents.MustFromInt(700),
		cents.MustFromRatio(3, 2),
		cents.MustFromRatio(3, 2).MustSub(cents.SyntonicComma().MustDiv(4)),
		cents.MustFromFraction(-137, 10),
		cents.Zero(),
	}
	slices.SortFunc(values, cents.Cents.Compare)
	fmt.Println(values)
	// Output:
	// [-13.7 0 5^(1/4) 700 3:2]
}

func ExampleCents_SplitOctaves() {
	octaves, rest := cents.MustFromInt(1900).SplitOctaves()
	fmt.Println(octaves, rest)
	octaves, rest = cents.MustFromInt(-100).SplitOctaves()
	fmt.Println(octaves, rest)
	// Output:
	// 1 700
	// -1 1100
}

func ExampleCents_ApplyTo() {
	fmt.Println(cents.Octave().MustApplyTo(440))
	fmt.Println(cents.Octave().MustMultiplier())
	// Output:
	// 880
	// 2
}

func Example_units() {
	fmt.Println(cents.MustFromInt(700).MustQuo(cents.Millioctave()))
	fmt.Println(cents.MustFromRatio(3, 2).MustQuo(cents.Savart()).Decimal(9))
	fmt.Println(cents.Millioctave().MustScale(1750, 3))
	// Output:
	// 1750/3
	// 176.091259056
	// 700
}

func ExampleCents_Quo() {
	fmt.Println(cents.Octave().MustQuo(cents.PythagoreanComma()).Decimal(6))
	fmt.Println(cents.Octave().MustQuo(cents.HoldrianComma()))
	// Output:
	// 51.150872
	// 53
}

func ExampleCents_NearestStep() {
	fifth := cents.MustFromRatio(3, 2)
	k, deviation := fifth.MustNearestStep(12)
	fmt.Println(k, deviation.Decimal(3, cents.WithPlus))
	k, deviation = fifth.MustNearestStep(53)
	fmt.Println(k, deviation.Decimal(3, cents.WithPlus))
	// Output:
	// 7 +1.955
	// 31 +0.068
}

func ExampleCents_Decimal() {
	fourth := cents.MustFromRatio(2, 3)
	fmt.Println(fourth.Decimal(3, 0))
	fmt.Println(fourth.Decimal(3, cents.Typographic|cents.WithUnit))
	// Output:
	// -701.955
	// −701.955 ¢
}

func ExampleCents_TextAs() {
	for _, c := range []cents.Cents{
		cents.MustFromInt(700),
		cents.MustStep(1, 19),
		cents.MustFromRatio(3, 2),
		cents.MustFromRatio(3, 2).MustSub(cents.SyntonicComma().MustDiv(4)),
	} {
		fmt.Println(c.Kind(), c)
	}

	fmt.Println(cents.Octave().TextAs(cents.KindRatio, 0))
	_, ok := cents.MustFromInt(700).TextAs(cents.KindRatio, 0)
	fmt.Println(ok)
	// Output:
	// decimal 700
	// fraction 1200/19 cents
	// ratio 3:2
	// power 5^(1/4)
	// 2:1 true
	// false
}

func ExampleParse() {
	for _, text := range []string{"+30 ¢", "3:2", "1200/19 cents", "5^(1/4)"} {
		fmt.Println(cents.MustParse(text))
	}
	// Output:
	// 30
	// 3:2
	// 1200/19 cents
	// 5^(1/4)
}

func ExampleParseScala() {
	c, rest := cents.MustParseScala("  100.0 cents")
	fmt.Printf("%s %q\n", c, rest)
	c, _ = cents.MustParseScala("3/2")
	fmt.Println(c)
	// Output:
	// 100 " cents"
	// 3:2
}

func ExampleCents_Scala() {
	fmt.Println(cents.MustFromRatio(3, 2).Scala(5))
	fmt.Println(cents.MustFromInt(700).Scala(5))
	// Output:
	// 3/2
	// 700.00000
}

func ExampleFromComposition() {
	// The monzo [-4 4 -1> of 81/80.
	comma := cents.MustFromComposition(nil, nil,
		cents.Power{Base: 2, Num: big.NewInt(-4)},
		cents.Power{Base: 3, Num: big.NewInt(4)},
		cents.Power{Base: 5, Num: big.NewInt(-1)},
	)
	fmt.Println(comma)

	meantone := cents.MustFromRatio(3, 2).MustSub(cents.SyntonicComma().MustDiv(4))
	num, den, powers := meantone.Composition()
	fmt.Printf("%s/%s\n", num, den)
	for _, p := range powers {
		fmt.Printf("%d → %s/%s\n", p.Base, p.Num, p.Den)
	}
	// Output:
	// 81:80
	// 0/1
	// 5 → 1/4
}

func Example_outOfRange() {
	_, err := cents.Octave().Mul(7501)
	fmt.Println(errors.Is(err, cents.ErrOutOfRange))
	// Output:
	// true
}

func Example_primeLimit() {
	_, err := cents.FromRatio(37, 32)
	fmt.Println(errors.Is(err, cents.ErrPrimeLimit))
	approx, _ := cents.MustParseScalaApprox("37/32")
	fmt.Println(approx)
	// Output:
	// true
	// 251.344039
}

func Example_invalidText() {
	_, err := cents.Parse("3/2")
	fmt.Println(errors.Is(err, cents.ErrInvalidText))
	_, err = cents.Parse("3:2")
	fmt.Println(err == nil)
	_, err = cents.Parse("3/2 ¢")
	fmt.Println(err == nil)
	// Output:
	// true
	// true
	// true
}
