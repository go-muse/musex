package octave_test

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
)

// The octave that a French score calls the third one is the "first" octave
// of a Russian textbook, the one-line octave in English, and octave 4 in ASPN.
func Example() {
	o := octave.MustFromNumber(3, number.FrancoBelgian())

	fmt.Println(o.Name(naming.Russian()))
	fmt.Println(o.MustNumber(number.ASPN()))
	fmt.Println(o.Name(naming.English()))
	// Output:
	// первая
	// 4
	// one-line
}

func ExampleFromASPN() {
	o := octave.FromASPN(4)

	fmt.Println(o)
	fmt.Println(o.MustNumber(number.FrancoBelgian()))
	// Output:
	// ASPN 4
	// 3
}

func ExampleFromNumber() {
	fb, err := octave.FromNumber(3, number.FrancoBelgian())
	if err != nil {
		fmt.Println(err)
		return
	}
	mex := octave.MustFromNumber(5, number.Mexican())
	c3 := octave.MustFromNumber(3, number.MiddleC3())

	fmt.Println(fb.Equal(mex), fb.Equal(c3))
	fmt.Println(fb)
	// Output:
	// true true
	// ASPN 4
}

func ExampleFromNumber_skipped() {
	_, err := octave.FromNumber(0, number.FrancoBelgian())

	fmt.Println(errors.Is(err, number.ErrSkipped))
	// Output:
	// true
}

func ExampleFromNumber_undefined() {
	var n number.Numbering // the zero Numbering is not a numbering
	_, err := octave.FromNumber(4, n)

	fmt.Println(errors.Is(err, number.ErrUndefined))
	// Output:
	// true
}

func ExampleFromName() {
	small, err := octave.FromName("малая", naming.Russian())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(small)
	fmt.Println(small.MustNumber(number.FrancoBelgian()))

	great := octave.MustFromName("great", naming.English())
	fmt.Println(great.Name(naming.Russian()))
	// Output:
	// ASPN 3
	// 2
	// большая
}

func ExampleFromName_unknown() {
	_, err := octave.FromName("Первая", naming.Russian()) // names are exact

	fmt.Println(errors.Is(err, naming.ErrUnknownName))
	// Output:
	// true
}

func ExampleOctave_Number() {
	o := octave.FromASPN(4)
	for _, n := range number.Builtins() {
		fmt.Printf("%s: %d\n", n.Label(), o.MustNumber(n))
	}

	// Number 0 is skipped in the Franco-Belgian numbering.
	fmt.Println(octave.FromASPN(1).MustNumber(number.FrancoBelgian()))
	// Output:
	// ASPN: 4
	// Franco-Belgian: 3
	// Mexican: 5
	// Middle C = C3: 3
	// Middle C = C5: 5
	// -1
}

func ExampleOctave_ASPN() {
	o := octave.MustFromNumber(3, number.FrancoBelgian())

	aspn, err := o.ASPN()
	fmt.Println(aspn, err)
	fmt.Println(o.MustASPN())
	// Output:
	// 4 <nil>
	// 4
}

func ExampleOctave_Name() {
	o := octave.FromASPN(4)
	fmt.Println(o.Name(naming.Russian()))
	fmt.Println(o.Name(naming.English()))
	fmt.Println(o.Name(naming.German()))

	// ASPN 8 has a name only in the Russian catalog.
	fmt.Printf("%q\n", octave.FromASPN(8).Name(naming.English()))
	// Output:
	// первая
	// one-line
	// eingestrichene
	// ""
}

func ExampleOctave_Equal() {
	a := octave.MustFromNumber(3, number.FrancoBelgian())
	b := octave.FromASPN(4)
	c := octave.MustFromName("первая", naming.Russian())

	fmt.Println(a.Equal(b), b.Equal(c))
	fmt.Println(a == c)
	fmt.Println(b.Equal(octave.MustFromNumber(4, number.FrancoBelgian())))
	// Output:
	// true true
	// true
	// false
}

func ExampleOctave_Compare() {
	fmt.Println(octave.FromASPN(3).Compare(octave.FromASPN(4)))

	xs := []octave.Octave{
		octave.FromASPN(5),
		octave.FromASPN(-1),
		octave.FromASPN(4),
		octave.FromASPN(0),
	}
	slices.SortFunc(xs, octave.Octave.Compare)
	for _, x := range xs {
		fmt.Println(x)
	}
	// Output:
	// -1
	// ASPN -1
	// ASPN 0
	// ASPN 4
	// ASPN 5
}

func ExampleOctave_Above() {
	four, five := octave.FromASPN(4), octave.FromASPN(5)

	fmt.Println(five.Above(four), four.Above(four))
	fmt.Println(four.Below(five), five.Below(four))
	// Output:
	// true false
	// true false
}

func ExampleOctave_Between() {
	lo, hi := octave.FromASPN(2), octave.FromASPN(6)

	fmt.Println(octave.FromASPN(4).Between(lo, hi))
	fmt.Println(lo.Between(lo, hi), hi.Between(hi, lo))
	fmt.Println(octave.FromASPN(7).Between(lo, hi))
	// Output:
	// true
	// true true
	// false
}

func ExampleOctave_Shift() {
	o := octave.FromASPN(4)

	up, err := o.Shift(1)
	fmt.Println(up, err)
	fmt.Println(o.MustShift(-5))
	fmt.Println(o.MustShift(0))
	// Output:
	// ASPN 5 <nil>
	// ASPN -1
	// ASPN 4
}

func ExampleOctave_ShiftDown() {
	o := octave.FromASPN(4)
	fmt.Println(o.MustShiftDown(1))
	fmt.Println(o.MustShiftDown(-2))

	// Shifting down by k undoes shifting by k for every k, even math.MinInt.
	low := octave.FromASPN(0).MustShift(math.MinInt)
	fmt.Println(low.MustShiftDown(math.MinInt))
	// Output:
	// ASPN 3
	// ASPN 6
	// ASPN 0
}

func ExampleOctave_Up() {
	o := octave.FromASPN(4)
	fmt.Println(o.MustUp(), o.MustDown())

	// The octaves on both sides of the skipped Franco-Belgian number 0 are
	// neighbors.
	fb := number.FrancoBelgian()
	fmt.Println(octave.MustFromNumber(1, fb).MustDown().MustNumber(fb))
	fmt.Println(octave.MustFromNumber(-1, fb).MustUp().MustNumber(fb))
	// Output:
	// ASPN 5 ASPN 3
	// -1
	// 1
}

func ExampleOctave_Sub() {
	fmt.Println(octave.FromASPN(5).MustSub(octave.FromASPN(4)))

	// The difference of octaves, not of numbers: the Franco-Belgian numbers 3
	// and -1 differ by 4, but the octaves differ by 3.
	fb := number.FrancoBelgian()
	d, err := octave.MustFromNumber(3, fb).Sub(octave.MustFromNumber(-1, fb))
	fmt.Println(d, err)
	// Output:
	// 1
	// 3 <nil>
}

func ExampleOctave_Distance() {
	a, b := octave.FromASPN(2), octave.FromASPN(5)

	fmt.Println(a.MustDistance(b), b.MustDistance(a))
	// Output:
	// 3 3
}

func ExampleRange() {
	up, err := octave.Range(octave.FromASPN(2), octave.FromASPN(5))
	if err != nil {
		fmt.Println(err)
		return
	}
	for o := range up {
		fmt.Println(o, o.Name(naming.English()))
	}

	for o := range octave.MustRange(octave.FromASPN(5), octave.FromASPN(2)) {
		fmt.Println(o)
	}
	// Output:
	// ASPN 2 great
	// ASPN 3 small
	// ASPN 4 one-line
	// ASPN 5 two-line
	// ASPN 5
	// ASPN 4
	// ASPN 3
	// ASPN 2
}

func ExampleNewNumbering() {
	// "The first octave is my number 1."
	first := octave.MustFromName("первая", naming.Russian())
	mine, err := octave.NewNumbering(first, 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(octave.FromASPN(3).MustNumber(mine))

	// The octave of middle C is 3 and number 0 is skipped: Franco-Belgian.
	fb := octave.MustNewNumbering(octave.FromASPN(4), 3, 0)
	fmt.Println(fb.Equal(number.FrancoBelgian()))
	// Output:
	// 0
	// true
}

func ExampleNewCatalog() {
	c, err := octave.NewCatalog(map[octave.Octave]string{
		octave.FromASPN(4):  "middle",
		octave.FromASPN(-3): "deep",
		octave.FromASPN(12): "high",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(octave.FromASPN(-3).Name(c))
	fmt.Println(octave.MustFromName("high", c))
	fmt.Printf("%q\n", octave.FromASPN(5).Name(c))
	// Output:
	// deep
	// ASPN 12
	// ""
}

func ExampleOctave_IsValid() {
	var z octave.Octave // the zero value is not an octave

	fmt.Println(octave.FromASPN(0).IsValid())
	fmt.Println(z.IsValid())
	// Output:
	// true
	// false
}

func ExampleOctave_String() {
	var z octave.Octave

	fmt.Println(octave.FromASPN(4).String())
	fmt.Println(octave.FromASPN(-1).String())
	fmt.Println(z.String())
	fmt.Printf("%v %#v\n", octave.FromASPN(4), octave.FromASPN(4))
	// Output:
	// ASPN 4
	// ASPN -1
	// Octave(invalid)
	// ASPN 4 octave.FromASPN(4)
}

func Example_outOfRange() {
	_, err := octave.FromASPN(math.MaxInt).Shift(1)
	fmt.Println(errors.Is(err, octave.ErrOutOfRange))

	// The ASPN number of this octave would be math.MaxInt + 1.
	_, err = octave.FromNumber(math.MaxInt, number.MiddleC3())
	fmt.Println(errors.Is(err, octave.ErrOutOfRange), errors.Is(err, number.ErrOutOfRange))

	// octave/number still converts the number: it is a number, not an octave.
	x, err := number.Convert(math.MaxInt, number.MiddleC3(), number.FrancoBelgian())
	fmt.Println(x == math.MaxInt, err)
	// Output:
	// true
	// true true
	// true <nil>
}

func Example_zeroValue() {
	var z octave.Octave // invalid: not ASPN 0 or ASPN 4

	_, err := z.Number(number.ASPN())
	fmt.Println(errors.Is(err, octave.ErrInvalid))
	fmt.Printf("%q\n", z.Name(naming.Russian()))
	fmt.Println(z.Equal(z), z == octave.Octave{})
	fmt.Println(z.Compare(octave.Octave{}), z.Compare(octave.FromASPN(math.MinInt)))
	fmt.Println(z)
	// Output:
	// true
	// ""
	// false true
	// 0 -1
	// Octave(invalid)
}

func ExampleNewNumbering_invalid() {
	_, err := octave.NewNumbering(octave.FromASPN(4), 0, 0) // number 0 is both used and skipped

	fmt.Println(errors.Is(err, number.ErrInvalidDefinition))
	// Output:
	// true
}

func ExampleNewCatalog_duplicate() {
	_, err := octave.NewCatalog(map[octave.Octave]string{
		octave.FromASPN(1): "low",
		octave.FromASPN(2): "low",
	})

	fmt.Println(errors.Is(err, naming.ErrDuplicateName))
	// Output:
	// true
}

func ExampleMustFromNumber() {
	n := octave.MustFromName("первая", naming.Russian()).MustShift(-1).MustNumber(number.FrancoBelgian())
	fmt.Println(n)

	defer func() {
		err, ok := recover().(error)
		fmt.Println(ok && errors.Is(err, number.ErrSkipped))
	}()
	octave.MustFromNumber(0, number.FrancoBelgian())
	// Output:
	// 2
	// true
}

func ExampleOctave_GoString() {
	var z octave.Octave

	fmt.Println(octave.FromASPN(4).GoString())
	fmt.Printf("%#v\n", octave.FromASPN(-1))
	fmt.Println(z.GoString())
	// Output:
	// octave.FromASPN(4)
	// octave.FromASPN(-1)
	// octave.Octave{}
}
