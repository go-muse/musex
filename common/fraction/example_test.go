package fraction_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/common/fraction"
)

func Example() {
	// Creation preserving the original pair (e.g., 6/12 of a circle) and reduced creation.
	pos := fraction.MustNew(6, 12)
	half := fraction.MustNewReduced(6, 12)

	fmt.Println(pos.String(), half.String())
	fmt.Println("equal value:", pos.Equal(half), "equal pair:", pos.EqualExact(half))

	// Adding triplet eighth notes (1/8 * 2/3 = 1/12):
	tripletEighth := fraction.MustNew(1, 8).MustMul(fraction.MustNew(2, 3))
	sumReduced := tripletEighth.MustAdd(tripletEighth)
	sumLCM := tripletEighth.MustAddWith(tripletEighth, fraction.DenomUnreducedLCM)
	fmt.Println(tripletEighth, sumReduced, sumLCM)
	// Output:
	// 6/12 1/2
	// equal value: true equal pair: false
	// 1/12 1/6 2/12
}

func Example_errors() {
	var invalid fraction.Fraction
	_, err := invalid.Add(fraction.One())
	fmt.Println("invalid:", errors.Is(err, fraction.ErrInvalid))

	_, err = fraction.New(1, 0)
	fmt.Println("zero den:", errors.Is(err, fraction.ErrZeroDenominator))

	_, err = fraction.One().Div(fraction.Zero())
	fmt.Println("div zero:", errors.Is(err, fraction.ErrDivisionByZero))

	_, err = fraction.MustNew(1, 4).AddWith(fraction.MustNew(1, 8), fraction.DenomLargerUnit)
	fmt.Println("not representable:", errors.Is(err, fraction.ErrNotRepresentable))
	// Output:
	// invalid: true
	// zero den: true
	// div zero: true
	// not representable: true
}

func ExampleNew() {
	raw1, _ := fraction.New(6, 12)
	raw2, _ := fraction.New(1, -2)
	raw3, _ := fraction.New(0, 3)

	red1, _ := fraction.NewReduced(6, 12)
	red2, _ := fraction.NewReduced(1, -2)
	red3, _ := fraction.NewReduced(0, 3)

	fmt.Println("New:", raw1, raw2, raw3)
	fmt.Println("NewReduced:", red1, red2, red3)
	// Output:
	// New: 6/12 1/-2 0/3
	// NewReduced: 1/2 -1/2 0/1
}

func ExampleFraction_Reduce() {
	a := fraction.MustNew(3, 2)
	b := fraction.MustNew(6, 4)

	byPair := map[fraction.Fraction]string{
		a: "3/2",
		b: "6/4",
	}
	byValue := map[fraction.Fraction]string{
		a.Reduce(): "first",
		b.Reduce(): "second",
	}

	fmt.Println("by pair len:", len(byPair))
	fmt.Println("by reduced value len:", len(byValue), byValue[fraction.MustNew(3, 2)])
	// Output:
	// by pair len: 2
	// by reduced value len: 1 second
}

func ExampleFraction_WithDenominator() {
	third := fraction.MustNew(1, 3)
	quarter := fraction.MustNew(1, 4)

	fmt.Println(third.MustWithDenominator(12))
	fmt.Println(quarter.MustWithDenominator(12))
	// Output:
	// 4/12
	// 3/12
}

func ExampleFraction_StepNumerator() {
	pos := fraction.MustNew(6, 12)

	stepped := pos.MustStepNumerator(1)
	added := pos.MustAddInt(1)

	fmt.Println("StepNumerator(1):", stepped)
	fmt.Println("AddInt(1):", added)
	// Output:
	// StepNumerator(1): 7/12
	// AddInt(1): 3/2
}

func ExampleFraction_AddWith() {
	sixth := fraction.MustNew(1, 6)
	fmt.Println("Reduced:", sixth.MustAddWith(sixth, fraction.DenomReduced))
	fmt.Println("LCM:", sixth.MustAddWith(sixth, fraction.DenomUnreducedLCM))
	fmt.Println("Product:", sixth.MustAddWith(sixth, fraction.DenomUnreducedProduct))
	fmt.Println("SameUnit:", sixth.MustAddWith(sixth, fraction.DenomSameUnit))

	q := fraction.MustNew(1, 4)
	fmt.Println("LargerUnit:", q.MustAddWith(fraction.MustNew(2, 8), fraction.DenomLargerUnit))
	fmt.Println("SmallerUnit:", q.MustAddWith(fraction.MustNew(1, 8), fraction.DenomSmallerUnit))
	fmt.Println("InDenominator(12):", fraction.MustNew(1, 3).MustAddInDenominator(q, 12))
	// Output:
	// Reduced: 1/3
	// LCM: 2/6
	// Product: 12/36
	// SameUnit: 2/6
	// LargerUnit: 2/4
	// SmallerUnit: 3/8
	// InDenominator(12): 7/12
}

func ExampleFraction_SplitFloor() {
	f := fraction.MustNew(-1, 12)

	w, r := f.MustSplit()
	fmt.Println("Split:", w, r)

	p, rFloor := f.MustSplitFloor()
	fmt.Println("SplitFloor:", p, rFloor)

	q, rem := f.MustDivModFloor(fraction.MustNew(1, 4))
	fmt.Println("DivModFloor by 1/4:", q, rem)
	// Output:
	// Split: 0 -1/12
	// SplitFloor: -1 11/12
	// DivModFloor by 1/4: -1 1/6
}

func ExampleFraction_RoundToInt() {
	negHalf := fraction.MustNew(-1, 2)
	fmt.Println("HalfAwayFromZero:", negHalf.MustRoundToInt(fraction.RoundHalfAwayFromZero))
	fmt.Println("HalfToEven (-1/2, 3/2):", negHalf.MustRoundToInt(fraction.RoundHalfToEven), fraction.MustNew(3, 2).MustRoundToInt(fraction.RoundHalfToEven))
	fmt.Println("Floor:", negHalf.MustRoundToInt(fraction.RoundFloor))
	fmt.Println("Ceil:", negHalf.MustRoundToInt(fraction.RoundCeil))
	fmt.Println("Trunc:", negHalf.MustRoundToInt(fraction.RoundTrunc))

	sixth := fraction.MustNew(1, 6)
	fmt.Println("1/6 to /10:", sixth.MustRoundToDenominator(10, fraction.RoundHalfAwayFromZero), sixth.MustRoundToDenominator(10, fraction.RoundTrunc))
	// Output:
	// HalfAwayFromZero: -1
	// HalfToEven (-1/2, 3/2): 0 2
	// Floor: -1
	// Ceil: 0
	// Trunc: 0
	// 1/6 to /10: 2/10 1/10
}

func ExampleParse() {
	a := fraction.MustParse("6/12")
	b := fraction.MustParse("5")
	c := fraction.MustParse("1 1/4")
	d := fraction.MustParse("-1 1/4")
	e := fraction.MustParse("7.1")
	f := fraction.MustParse("-0.25")

	fmt.Println(a.String(), a.ReducedString())
	fmt.Println(b.String(), b.CompactString())
	fmt.Println(c.String(), c.MixedString())
	fmt.Println(d.String(), d.MixedString())
	fmt.Println(e.String(), e.MustDecimal(1))
	fmt.Println(f.String(), f.MustDecimal(2))
	// Output:
	// 6/12 1/2
	// 5/1 5
	// 5/4 1 1/4
	// -5/4 -1 1/4
	// 71/10 7.1
	// -1/4 -0.25
}
