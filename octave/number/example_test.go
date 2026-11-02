package number_test

import (
	"errors"
	"fmt"
	"math"

	"github.com/go-muse/musex/octave/number"
)

// An octave is given by its ASPN number; a numbering gives it a visible
// number. Numbers convert between any two numberings.
func Example() {
	fmt.Println(number.FrancoBelgian().MustFromASPN(4))

	x, err := number.Convert(3, number.MiddleC3(), number.FrancoBelgian())
	fmt.Println(x, err == nil)
	// Output:
	// 3
	// 3 true
}

// Builtins lists the built-in numberings in a fixed order. The second column
// is the number of the octave of middle C (ASPN 4).
func ExampleBuiltins() {
	for _, n := range number.Builtins() {
		fmt.Printf("%s: %d\n", n.Label(), n.MustFromASPN(4))
	}
	// Output:
	// ASPN: 4
	// Franco-Belgian: 3
	// Mexican: 5
	// Middle C = C3: 3
	// Middle C = C5: 5
}

// Built-in numberings have labels; own numberings have none, even when they
// equal a built-in one.
func ExampleNumbering_Label() {
	fmt.Println(number.FrancoBelgian().Label())

	own := number.MustNew(4, 4)
	fmt.Printf("%q %t\n", own.Label(), own.Equal(number.ASPN()))
	// Output:
	// Franco-Belgian
	// "" true
}

// New defines an own numbering by one pair: the octave with ASPN number 4
// has number 5.
func ExampleNew() {
	n, err := number.New(4, 5)
	fmt.Println(n, err == nil)
	fmt.Println(n.MustFromASPN(4), n.MustFromASPN(0))
	fmt.Println(number.MustNew(4, 3).Equal(number.MiddleC3()))
	// Output:
	// ASPN 4 -> 5 true
	// 5 1
	// true
}

// Skipped numbers are not used: the octaves around them get the next used
// numbers.
func ExampleNew_skipped() {
	n := number.MustNew(4, 4, 0, 1)
	fmt.Println(n)
	fmt.Println(n.MustFromASPN(2), n.MustFromASPN(1))
	fmt.Println(number.MustNew(4, 3, 0).Equal(number.FrancoBelgian()))
	// Output:
	// ASPN 4 -> 4, skip [0 1]
	// 2 -1
	// true
}

// An invalid definition is an error: the anchor number is skipped.
func ExampleNew_invalid() {
	n, err := number.New(4, 3, 3)
	fmt.Println(errors.Is(err, number.ErrInvalidDefinition), n.IsDefined())
	// Output:
	// true false
}

// FromASPN gives the number of an octave; Franco-Belgian has no number 0.
func ExampleNumbering_FromASPN() {
	fb := number.FrancoBelgian()
	for _, aspn := range []int{2, 1, 0} {
		x, err := fb.FromASPN(aspn)
		fmt.Println(aspn, x, err == nil)
	}
	// Output:
	// 2 1 true
	// 1 -1 true
	// 0 -2 true
}

// ToASPN gives the ASPN number of the octave with a given number.
func ExampleNumbering_ToASPN() {
	fb := number.FrancoBelgian()
	for _, x := range []int{3, -1} {
		aspn, err := fb.ToASPN(x)
		fmt.Println(x, aspn, err == nil)
	}

	aspn, err := number.Mexican().ToASPN(5)
	fmt.Println(5, aspn, err == nil)
	// Output:
	// 3 4 true
	// -1 1 true
	// 5 4 true
}

// A skipped number has no octave: no nearest number is substituted.
func ExampleNumbering_ToASPN_skipped() {
	_, err := number.FrancoBelgian().ToASPN(0)
	fmt.Println(errors.Is(err, number.ErrSkipped))
	// Output:
	// true
}

// Convert gives the number of the same octave in another numbering.
func ExampleConvert() {
	x, err := number.Convert(3, number.MiddleC3(), number.FrancoBelgian())
	fmt.Println(x, err == nil)

	x, err = number.Convert(1, number.FrancoBelgian(), number.MiddleC5())
	fmt.Println(x, err == nil)
	// Output:
	// 3 true
	// 3 true
}

// MustConvert is for values known to be valid; on an error it panics with
// that error.
func ExampleMustConvert() {
	fmt.Println(number.MustConvert(5, number.Mexican(), number.MiddleC3()))

	defer func() {
		err, _ := recover().(error)
		fmt.Println(errors.Is(err, number.ErrSkipped))
	}()
	number.MustConvert(0, number.FrancoBelgian(), number.ASPN())
	// Output:
	// 3
	// true
}

// Numberings are equal when they give the same numbers to all octaves;
// labels do not matter.
func ExampleNumbering_Equal() {
	mex, c5 := number.Mexican(), number.MiddleC5()
	fmt.Println(mex.Equal(c5), mex.Label(), "/", c5.Label())
	fmt.Println(number.FrancoBelgian().Equal(number.MiddleC3()))
	// Output:
	// true Mexican / Middle C = C5
	// false
}

// String prints the label of a built-in numbering and the definition of an
// own one. It is not equality: equal numberings may print differently.
func ExampleNumbering_String() {
	fmt.Println(number.FrancoBelgian())
	fmt.Println(number.MustNew(4, 3, 0))
	fmt.Println(number.MustNew(-1, 0))
	// Output:
	// Franco-Belgian
	// ASPN 4 -> 3, skip [0]
	// ASPN -1 -> 0
}

// A number that does not fit into int is an error, not a wrapped value.
func Example_outOfRange() {
	_, err := number.MiddleC5().FromASPN(math.MaxInt)
	fmt.Println(errors.Is(err, number.ErrOutOfRange))
	// Output:
	// true
}

// The zero value is an undefined numbering, not ASPN.
func Example_zeroValue() {
	var zero number.Numbering
	fmt.Println(zero.IsDefined(), zero)

	_, err := zero.FromASPN(4)
	fmt.Println(errors.Is(err, number.ErrUndefined), zero.Equal(zero))
	// Output:
	// false Numbering(undefined)
	// true false
}
