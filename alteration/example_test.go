package alteration_test

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/go-muse/musex/alteration"
)

// A value of alteration is an exact number of chromatic units, whole or
// fractional: values are raised, lowered and subtracted exactly.
func Example() {
	half := alteration.MustFromFraction(1, 2)
	fmt.Println(half.MustRaiseBy(1))
	fmt.Println(alteration.Sharp().MustSub(alteration.Flat()))
	// Output:
	// +3/2
	// +2
}

// The five traditional signs are ready values; the zero value is natural.
func Example_signs() {
	signs := []struct {
		name  string
		value alteration.Alteration
	}{
		{"natural", alteration.Natural()},
		{"sharp", alteration.Sharp()},
		{"flat", alteration.Flat()},
		{"double sharp", alteration.DoubleSharp()},
		{"double flat", alteration.DoubleFlat()},
	}
	for _, s := range signs {
		fmt.Println(s.name, s.value)
	}

	var zero alteration.Alteration
	fmt.Println(zero.IsNatural(), zero == alteration.Natural())
	// Output:
	// natural 0
	// sharp +1
	// flat -1
	// double sharp +2
	// double flat -2
	// true true
}

// FromInt creates a value from any whole number of chromatic units in the
// guaranteed range.
func ExampleFromInt() {
	tripleSharp, err := alteration.FromInt(3)
	fmt.Println(tripleSharp, tripleSharp.Sign(), err == nil)

	minusFour, err := alteration.FromInt(-4)
	fmt.Println(minusFour, minusFour.Sign(), err == nil)

	_, err = alteration.FromInt(alteration.MaxMagnitude + 1)
	fmt.Println(errors.Is(err, alteration.ErrOutOfRange))
	// Output:
	// +3 1 true
	// -4 -1 true
	// true
}

// FromFraction normalizes the sign and reduces the fraction; a zero
// denominator is an error.
func ExampleFromFraction() {
	for _, f := range [][2]int{{2, 4}, {1, -2}, {4, 2}} {
		a, err := alteration.FromFraction(f[0], f[1])
		fmt.Println(a, err == nil)
	}

	_, err := alteration.FromFraction(1, 0)
	fmt.Println(errors.Is(err, alteration.ErrZeroDenominator))
	// Output:
	// +1/2 true
	// -1/2 true
	// +2 true
	// true
}

// Raise and Lower move a value by one chromatic unit.
func ExampleAlteration_Raise() {
	a := alteration.Flat()
	for range 3 {
		a = a.MustRaise()
		fmt.Println(a)
	}
	fmt.Println(a.MustLower())

	up, err := alteration.MustFromFraction(1, 2).Raise()
	fmt.Println(up, err == nil)
	// Output:
	// 0
	// +1
	// +2
	// +1
	// +3/2 true
}

// RaiseBy and LowerBy move a value by a whole number of chromatic units.
func ExampleAlteration_RaiseBy() {
	fmt.Println(alteration.Natural().MustRaiseBy(3))
	fmt.Println(alteration.Sharp().MustRaiseBy(-2))
	fmt.Println(alteration.DoubleSharp().MustLowerBy(4))

	r, err := alteration.MustFromFraction(1, 2).RaiseBy(2)
	fmt.Println(r, err == nil)
	// Output:
	// +3
	// -1
	// -2
	// +5/2 true
}

// Add raises a value by another value, whole or fractional.
func ExampleAlteration_Add() {
	half := alteration.MustFromFraction(1, 2)
	fmt.Println(alteration.Natural().MustAdd(half))

	sum, err := half.Add(half)
	fmt.Println(sum, sum.Equal(alteration.Sharp()), err == nil)

	fmt.Println(half.MustAdd(alteration.MustFromFraction(1, 3)))
	// Output:
	// +1/2
	// +1 true true
	// +5/6
}

// Sub lowers a value by another value; it is also the difference: how much
// higher one value is than another.
func ExampleAlteration_Sub() {
	fmt.Println(alteration.Sharp().MustSub(alteration.Flat()))

	d, err := alteration.MustFromFraction(1, 2).Sub(alteration.MustFromFraction(1, 3))
	fmt.Println(d, err == nil)

	fmt.Println(alteration.Sharp().MustSub(alteration.MustFromFraction(1, 2)))
	// Output:
	// +2
	// +1/6 true
	// +1/2
}

// Compare orders values by magnitude and works with the slices package.
func ExampleAlteration_Compare() {
	xs := []alteration.Alteration{
		alteration.MustFromInt(3),
		alteration.MustFromFraction(1, 2),
		alteration.Flat(),
		alteration.MustFromFraction(-7, 2),
		alteration.Natural(),
		alteration.DoubleSharp(),
		alteration.MustFromFraction(1, 3),
		alteration.DoubleFlat(),
		alteration.MustFromFraction(3, 2),
		alteration.MustFromFraction(-1, 2),
		alteration.Sharp(),
	}
	slices.SortFunc(xs, alteration.Alteration.Compare)
	// The %v verb prints each element with its String method.
	fmt.Printf("%v\n", xs)

	fmt.Println(alteration.Flat().Compare(alteration.MustFromFraction(1, 2)))
	// Output:
	// [-7/2 -2 -1 -1/2 0 +1/3 +1/2 +1 +3/2 +2 +3]
	// -1
}

// Mul multiplies a value by an integer.
func ExampleAlteration_Mul() {
	fmt.Println(alteration.MustFromFraction(1, 2).MustMul(2))
	fmt.Println(alteration.MustFromFraction(1, 3).MustMul(3))
	fmt.Println(alteration.MustFromFraction(-3, 2).MustMul(4))

	p, err := alteration.DoubleFlat().Mul(-1)
	fmt.Println(p, err == nil)
	// Output:
	// +1
	// +1
	// -6
	// +2 true
}

// A result outside the guaranteed range is an error; the input is unchanged
// and the Must variant panics with that error. Only the exact result matters:
// natural times the largest int is natural.
func Example_outOfRange() {
	top := alteration.MustFromInt(alteration.MaxMagnitude)

	r, err := top.Raise()
	fmt.Println(errors.Is(err, alteration.ErrOutOfRange), errors.Is(err, alteration.ErrZeroDenominator))
	fmt.Println(r.IsNatural(), top)

	_, err = alteration.FromFraction(1, alteration.MaxDenominator+1)
	fmt.Println(errors.Is(err, alteration.ErrOutOfRange))

	fmt.Println(alteration.Natural().MustMul(math.MaxInt))

	func() {
		defer func() {
			e, ok := recover().(error)
			fmt.Println(ok && errors.Is(e, alteration.ErrOutOfRange))
		}()
		_ = top.MustRaise()
	}()
	// Output:
	// true false
	// true +16777216
	// true
	// 0
	// true
}

// Fraction returns the exact value in lowest terms; FromFraction accepts it
// back.
func ExampleAlteration_Fraction() {
	num, den := alteration.MustFromFraction(6, -4).Fraction()
	fmt.Println(num, den)
	fmt.Println(alteration.MustFromFraction(num, den))
	// Output:
	// -3 2
	// -3/2
}

// Sign gives the direction: +1 raises, -1 lowers, 0 is natural.
func ExampleAlteration_Sign() {
	for _, a := range []alteration.Alteration{alteration.MustFromFraction(1, 3), alteration.Natural(), alteration.Flat()} {
		fmt.Println(a, a.Sign())
	}
	// Output:
	// +1/3 1
	// 0 0
	// -1 -1
}

// IsInteger reports a whole number of chromatic units; IsNatural reports 0.
func ExampleAlteration_IsInteger() {
	half := alteration.MustFromFraction(1, 2)
	fmt.Println(half.IsInteger(), half.MustMul(2).IsInteger())
	fmt.Println(alteration.MustFromFraction(4, 2).IsInteger(), alteration.MustFromFraction(0, 7).IsNatural())
	// Output:
	// false true
	// true true
}

// Equal compares exact numbers, however they were obtained; == gives the
// same answer, and values are valid map keys.
func ExampleAlteration_Equal() {
	a := alteration.MustFromFraction(2, 4)
	b := alteration.Natural().MustAdd(alteration.MustFromFraction(1, 2))
	fmt.Println(a.Equal(b), a == b)

	names := map[alteration.Alteration]string{alteration.Sharp(): "sharp"}
	fmt.Println(names[a.MustAdd(b)])
	// Output:
	// true true
	// sharp
}

// String gives a fixed form with an ASCII sign; it is not a notation.
func ExampleAlteration_String() {
	fmt.Println(alteration.Natural().String(), alteration.DoubleFlat().String())
	fmt.Println(alteration.MustFromFraction(-7, 2), alteration.MustFromFraction(5, 6))
	// Output:
	// 0 -2
	// -7/2 +5/6
}

// MustFromFraction is for literals and values known to be valid; it panics
// with the error of FromFraction.
func ExampleMustFromFraction() {
	fmt.Println(alteration.MustFromFraction(1, 2).MustRaiseBy(1))

	func() {
		defer func() {
			e, ok := recover().(error)
			fmt.Println(ok && errors.Is(e, alteration.ErrZeroDenominator))
		}()
		_ = alteration.MustFromFraction(1, 0)
	}()
	// Output:
	// +3/2
	// true
}
