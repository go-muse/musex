// Package number provides octave numberings: the rules by which octaves get
// their visible numbers, such as 4 for the octave of middle C in ASPN or 3 in
// the Franco-Belgian numbering. It provides built-in numberings, own
// numberings defined by the caller, and conversion of octave numbers between
// any two numberings.
//
// # ASPN numbers
//
// An octave is identified by its ASPN number, an int. The octave that
// contains middle C (c¹) has ASPN number 4, and neighboring octaves have
// neighboring numbers: the octave below it is 3, the octave above it is 5.
// Any int is an ASPN number; there is no musical limit, and the MIDI note
// range does not apply. The ASPN number is the number of an octave, not the
// number written in the name of a note.
//
// ASPN (American Standard Pitch Notation) is also called SPN (Scientific
// Pitch Notation) and IPN (International Pitch Notation). In this package it
// is a numbering of octaves, not a notation for note names such as "C♯4".
// A numbering is not tied to any notation: it defines only numbers.
//
// # Numberings
//
// A [Numbering] gives each octave a number. It is defined by an anchor pair
// (the octave with ASPN number a has number x) and a finite set of skipped
// numbers that no octave has. Octave numbers strictly increase with the
// ASPN number, take every other integer, and continue without limit in both
// directions.
//
// The package defines the following operations:
//
//   - [ASPN], [FrancoBelgian], [Mexican], [MiddleC3], and [MiddleC5] return
//     the built-in numberings; [Builtins] lists them in a fixed order.
//   - [New] and [MustNew] create an own numbering from an anchor pair and
//     optional skipped numbers.
//   - [Numbering.FromASPN] returns the number of an octave given its ASPN
//     number; [Numbering.ToASPN] returns the ASPN number of the octave with
//     a given number.
//   - [Convert] converts an octave number from one numbering to another.
//   - [Numbering.Equal] reports whether two numberings give the same number
//     to every octave.
//   - [Numbering.Label] returns the label of a built-in numbering, and
//     [Numbering.String] returns a fixed form for output.
//
// # Built-in numberings
//
// The package provides five built-in numberings, listed by [Builtins] in this
// order:
//
//	Label           Middle C  Skipped  Used in
//	ASPN            4         none     English-language literature, science
//	Franco-Belgian  3         0        France, Belgium, parts of Spain
//	Mexican         5         none     Mexico, alongside ASPN
//	Middle C = C3   3         none     Yamaha, Logic Pro (default), Cubase (reportedly)
//	Middle C = C5   5         none     FL Studio
//
// The "Middle C" column is the number of the octave of middle C (ASPN 4).
//
// ASPN is also called SPN and IPN. Labels on Roland and Korg instruments
// match ASPN.
//
// In the Franco-Belgian numbering, the number is the ASPN number minus 1 for
// ASPN 2 and above, and minus 2 for ASPN 1 and below: number 0 is skipped.
// It is used in France, Belgium, and parts of Spain; in Italy it is
// prescribed by a law of 1989, but ASPN is common there in practice. Sources
// confirm the numbers -2 to 7. Below -2, the numbers continue in sequence
// without confirmation by sources; above 7, the source sets no limit, and
// the numbers continue in sequence.
//
// The Mexican numbering is the numbering of the Spanish translations of
// Riemann; it is used in Mexico alongside ASPN. A single encyclopedia article
// confirms the numbers from 1 (the lowest audible C, ASPN 0). Below 1, the
// numbers continue in sequence without confirmation by sources. It gives the
// same numbers as Middle C = C5, so [Mexican] and [MiddleC5] are equal by
// [Numbering.Equal], but their labels differ.
//
// The numbering with middle C as C3 is used by Yamaha instruments and by
// default in Logic Pro, and reportedly in Cubase. The numbering with middle C
// as C5 is used in FL Studio.
//
// # Own numberings
//
// [New] creates an own numbering from an anchor pair and skipped numbers.
// New(4, 3) gives number 3 to the octave with ASPN number 4 and equals
// [MiddleC3]; New(4, 3, 0) also skips number 0 and equals [FrancoBelgian].
// An own numbering has no label.
//
// Two numberings are equal if they give the same number to every octave,
// however they were created and whatever their labels.
//
// # Errors
//
// Operations return errors that match one of the following with [errors.Is]:
//
//   - [ErrSkipped]: the number is skipped in the numbering, such as 0 in the
//     Franco-Belgian numbering; no nearby number is used instead.
//   - [ErrInvalidDefinition]: the anchor number of an own numbering is among
//     its skipped numbers, or a skipped number is repeated.
//   - [ErrOutOfRange]: the exact result does not fit in an int. Results never
//     wrap around; the limits of int are not musical limits.
//   - [ErrUndefined]: the numbering is the zero [Numbering].
//
// On error, operations return 0 or the zero Numbering. Only the functions and
// methods whose names start with Must panic, with the error as the panic
// value.
//
// # Zero value
//
// The zero value of [Numbering] is not a numbering: it means that no
// numbering is given, and it is never treated as ASPN or any other numbering.
// [Numbering.IsDefined] reports false for it, [Numbering.FromASPN],
// [Numbering.ToASPN], and [Convert] return [ErrUndefined], and
// [Numbering.Equal] returns false, even when both values are zero.
//
// # Concurrency
//
// Numberings are immutable values, and the package has no mutable state. All
// functions and methods are safe for concurrent use by multiple goroutines.
package number
