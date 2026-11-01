package letter_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/letter"
)

// The package works with the seven-letter set C D E F G A B: letters are
// shifted around the circle and distances are counted in letter steps.
func Example() {
	fmt.Println(letter.C.Shift(9))
	fmt.Println(letter.B.DistanceUp(letter.D))
	// Output:
	// E
	// 2
}

// All lists the seven letters in canonical order; Count is their number.
func ExampleAll() {
	all := letter.All()
	// The %v verb prints each element with its String method.
	fmt.Printf("%v\n", all)
	fmt.Println(len(all), letter.Count)
	// Output:
	// [C D E F G A B]
	// 7 7
}

// Cycle walks the circle upward starting from the letter.
func ExampleLetter_Cycle() {
	fmt.Printf("%v\n", letter.A.Cycle())
	// Output:
	// [A B C D E F G]
}

// DistanceUp counts letter steps upward around the circle.
func ExampleLetter_DistanceUp() {
	fmt.Println(letter.C.DistanceUp(letter.E))
	fmt.Println(letter.B.DistanceUp(letter.D))
	fmt.Println(letter.C.DistanceUp(letter.B))
	// Output:
	// 2
	// 2
	// 6
}

// DistanceDown counts letter steps downward around the circle.
func ExampleLetter_DistanceDown() {
	fmt.Println(letter.C.DistanceDown(letter.E))
	fmt.Println(letter.C.DistanceDown(letter.B))
	// Output:
	// 5
	// 1
}

// Shift moves a letter by any number of steps: up for positive, down for
// negative, wrapping around the circle.
func ExampleLetter_Shift() {
	fmt.Println(letter.C.Shift(9))
	fmt.Println(letter.C.Shift(-1))
	fmt.Println(letter.B.Shift(1))
	// Output:
	// E
	// B
	// C
}

// FromOrdinal creates a letter from its place 0...6 and reports ErrInvalid
// for any other number.
func ExampleFromOrdinal() {
	l, err := letter.FromOrdinal(2)
	fmt.Println(l, err == nil)

	invalid, err := letter.FromOrdinal(7)
	fmt.Println(errors.Is(err, letter.ErrInvalid), invalid.IsValid())
	// Output:
	// E true
	// true false
}

// MustFromOrdinal suits ordinals known to be valid, for example in a chain.
// Place 2 is E; three steps up from E is A.
func ExampleMustFromOrdinal() {
	fmt.Println(letter.MustFromOrdinal(2))
	fmt.Println(letter.MustFromOrdinal(2).Shift(3))
	// Output:
	// E
	// A
}

// The zero Letter is not a letter: it means "no letter set" and is not C.
func ExampleLetter_IsValid() {
	var zero letter.Letter
	fmt.Println(zero.IsValid())
	fmt.Println(letter.C.IsValid())
	fmt.Println(letter.Letter(9).IsValid())
	// Output:
	// false
	// true
	// false
}

// Equal reports whether two values are the same valid letter, however they
// were obtained. An invalid value is never equal to anything.
func ExampleLetter_Equal() {
	var zero letter.Letter
	fmt.Println(letter.C.Shift(9).Equal(letter.E))
	fmt.Println(zero.Equal(letter.C))
	fmt.Println(zero.Equal(zero))
	// Output:
	// true
	// false
	// false
}

// Ordinal gives the place 0...6 of a letter, suitable as a table index.
func ExampleLetter_Ordinal() {
	var syllables [letter.Count]string
	syllables[letter.G.Ordinal()] = "sol"

	fmt.Println(letter.C.Ordinal(), letter.B.Ordinal())
	fmt.Println(syllables[4])
	// Output:
	// 0 6
	// sol
}

// String gives the fixed letter symbol; invalid values are shown with their
// number and never look like a letter.
func ExampleLetter_String() {
	var zero letter.Letter
	fmt.Println(letter.A.String())
	fmt.Println(zero.String())
	// Output:
	// A
	// Letter(0)
}
