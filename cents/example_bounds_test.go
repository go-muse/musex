package cents_test

import (
	"errors"
	"fmt"
	"math"

	"github.com/go-muse/musex/cents"
)

func Example_frequencyBounds() {
	fmt.Println(cents.MinNormalFrequency)
	fmt.Println(cents.MaxFrequency)
	fmt.Println(cents.MustFromInt(-1226400).MustMultiplier() == cents.MinNormalFrequency)

	// Output:
	// 2.2250738585072014e-308
	// 1.7976931348623157e+308
	// true
}

func ExampleIsValidFrequency() {
	fmt.Println(
		cents.IsValidFrequency(440),
		cents.IsValidFrequency(math.SmallestNonzeroFloat64),
		cents.IsValidFrequency(0),
		cents.IsValidFrequency(-440),
		cents.IsValidFrequency(math.Inf(1)),
		cents.IsValidFrequency(math.NaN()),
	)

	// Output: true true false false false false
}

func ExampleIsNormalFrequency() {
	fmt.Println(
		cents.IsNormalFrequency(440),
		cents.IsNormalFrequency(math.SmallestNonzeroFloat64),
		cents.IsNormalFrequency(cents.MinNormalFrequency),
		cents.IsNormalFrequency(cents.MaxFrequency),
		cents.IsNormalFrequency(math.Inf(1)),
	)

	// Output: true false true true false
}

func ExampleValidFrequency() {
	frequency, err := cents.ValidFrequency(math.SmallestNonzeroFloat64)
	fmt.Println(frequency, err == nil)

	invalid, err := cents.ValidFrequency(0)
	fmt.Println(invalid, errors.Is(err, cents.ErrInvalidArgument))

	// Output:
	// 5e-324 true
	// 0 true
}

func ExampleNormalFrequency() {
	frequency, err := cents.NormalFrequency(1.5)
	fmt.Println(frequency, err == nil)

	_, err = cents.NormalFrequency(math.SmallestNonzeroFloat64)
	fmt.Println(errors.Is(err, cents.ErrOutOfRange))

	_, err = cents.NormalFrequency(math.Inf(1))
	fmt.Println(errors.Is(err, cents.ErrOutOfRange))

	fmt.Println(cents.MustNormalFrequency(442))

	// Output:
	// 1.5 true
	// true
	// true
	// 442
}
