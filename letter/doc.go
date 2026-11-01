// Package letter provides the seven-letter set C D E F G A B: the letters
// (basic degrees) used to name notes, with their circular order and
// letter distances.
//
// A [Letter] is a letter of the seven-letter set. The letters follow one
// another in a circle: each letter is followed by the next one, and the last
// letter B is followed by the first letter C. The canonical order of the set
// starts with C: C D E F G A B. The ordinal of a letter is its place in this
// order, numbered from 0 (C is 0, B is 6).
//
// The package defines the following operations:
//
//   - [All] lists the letters in canonical order; [Count] is their number.
//   - [Letter.Cycle] walks the circle upward from a letter.
//   - [Letter.DistanceUp] and [Letter.DistanceDown] count letter steps
//     between two letters in a direction.
//   - [Letter.Shift] moves a letter by any number of letter steps,
//     wrapping around the circle.
//   - [Letter.Equal] reports whether two values are the same letter.
//   - [Letter.Ordinal], [FromOrdinal] and [MustFromOrdinal] convert between
//     letters and their ordinals.
//
// The number 7, the range of distances 0 to 6, and the shift period of 7
// are properties of the seven-letter set, not of letters in general.
// Letter sets with a different number of letters are not provided by this
// package.
//
// A letter is a letter name only: it carries no accidental, octave, pitch,
// or notation symbol. [Letter.String] returns a fixed form ("C" to "B") for
// output and debugging; it is not a notation and has no parser.
//
// # Invalid values
//
// The zero value of [Letter] is not a letter: it means that no letter is
// given, and it is never treated as C. Any value other than the seven
// constants, such as Letter(0), Letter(42), or B + 1, is an invalid value.
// All methods accept invalid values without panicking. Methods that return a
// number, a letter, a slice, or a string then return a result that no valid
// letter can produce: -1 for numbers, the zero Letter for letters, nil for
// slices, and "Letter(n)" for [Letter.String]. [Letter.IsValid] returns
// false for an invalid value, and [Letter.Equal] returns false whenever
// either value is invalid, even if both hold the same invalid value.
// Only [MustFromOrdinal] panics.
//
// # Concurrency
//
// Letters are immutable values, and the package has no mutable state. All
// functions and methods are safe for concurrent use by multiple goroutines.
package letter
