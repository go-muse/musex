// Package spelling provides note names: a letter of the seven-letter set
// together with an alteration value, such as C, C sharp, or B with -7/2.
//
// A Spelling is the written name of a note without an octave. It is made of
// two parts: a letter from package letter (C D E F G A B) and an alteration
// value from package alteration. Any alteration value is allowed, including
// fractional values such as +1/2 and values beyond the traditional accidentals,
// such as a triple sharp (+3) or -7/2. Thirty-five ready-made names are
// provided as functions: for each letter X, X() is the natural name, and
// XSharp, XFlat, XDoubleSharp, and XDoubleFlat add the traditional
// accidentals, for example C, CSharp, and BDoubleFlat. Predefined lists all
// 35 of them.
//
// The package defines the following operations:
//
//   - New and MustNew create a name from a letter and an alteration value;
//     Parse and MustParse read it from its machine form.
//   - Spelling.Letter, Spelling.Alteration, and Spelling.Parts return the
//     parts of a name.
//   - Spelling.Raise, Spelling.Lower, Spelling.RaiseBy, Spelling.LowerBy,
//     Spelling.RaiseByAlteration, and Spelling.LowerByAlteration raise or
//     lower the alteration value of a name by one chromatic unit, by a whole
//     number of units, or by another alteration value.
//   - Spelling.Natural and Spelling.WithAlteration replace the alteration
//     value and keep the letter.
//   - Spelling.Equal reports whether two names are the same; Spelling.Compare
//     orders names for sorting.
//   - Spelling.LetterDistanceUp, Spelling.LetterDistanceDown, and
//     Spelling.AlterationDifference compare the parts of two names.
//   - Spelling.IsValid, Spelling.IsNatural, Spelling.IsInteger,
//     Spelling.Direction, Spelling.SameLetter, and Spelling.SameAlteration
//     report properties of names.
//   - Spelling.String, Spelling.MarshalText, Spelling.AppendText,
//     Spelling.UnmarshalText, Spelling.MarshalJSON, and
//     Spelling.UnmarshalJSON write and read the machine form.
//
// # Names are not pitches
//
// Names are compared by how they are written, not by how they sound: C sharp
// is not equal to D flat, and B sharp is not equal to C. A name has no pitch.
// Which names sound the same (enharmonic equivalence), and which names a
// musical system allows, are questions for the musical system that uses the
// names, not for this package.
//
// Spelling.Compare defines an agreed order for sorting: by letter in the
// canonical order C D E F G A B, then by alteration value. It is not an order
// by pitch: B sharp sorts after C, and C flat sorts before B, although in the
// traditional system they sound the same.
//
// The letter distance and the alteration difference of two names are not an
// interval: they have no interval quality or direction. Intervals are the
// subject of a future package.
//
// Raising, lowering, and shifting a name by any value always keep its letter,
// with no enharmonic substitution: C sharp raised by one chromatic unit is C
// double sharp (C with +2), never D. All arithmetic on values is done by
// package alteration, with the same results and the same errors.
//
// A name is not a string of a notation. Writing and reading names in
// notations, such as "C♯", "Cis", or "до-диез", is the subject of a future
// notation package.
//
// # Machine form
//
// Every name has exactly one text form, the machine form, used by
// Spelling.String, Spelling.MarshalText, Spelling.AppendText, and JSON.
// It is meant for storage and exchange and is part of the stable API of the
// package: a name is always written the same way, and the form does not depend
// on any settings, notation, or locale. It is defined by this package and does
// not follow future changes of [alteration.Alteration.String].
//
// The form is ASCII text without spaces:
//
//	form        = name | "unset"
//	name        = letter [ sign number [ "/" denominator ] ]
//	letter      = "A" | "B" | "C" | "D" | "E" | "F" | "G"
//	sign        = "+" | "-"
//	number      = nonzero-digit { digit }    ; no leading zeros
//	denominator = nonzero-digit { digit }    ; at least 2, coprime with number
//
// A natural name is a single letter, such as "C" or "B". Otherwise the letter
// is followed by the sign of the alteration value, the absolute value of its
// numerator in lowest terms (see [alteration.Alteration.Fraction]), and, for a
// value that is not whole, a slash and the denominator: "C+1", "D-1", "E+3",
// "C+1/2", "B-7/2", "C+16777216". The unset name is written as "unset".
//
// Reading accepts only the canonical form: each accepted string is the form of
// exactly one name. Leading zeros, a zero value written with a sign ("C+0"),
// a fraction that is not in lowest terms ("C+2/4", "C+1/1"), spaces, lower-case
// letters, accidental symbols ("C#", "C♯"), and notation names ("Cis") are
// errors. Reading errors match ErrInvalidForm with errors.Is. If the input has
// the shape of a number, that is, a letter, a sign, and decimal digits without
// leading zeros, optionally followed by a slash and either such digits or the
// single digit 0, then the error also matches:
//
//   - [alteration.ErrZeroDenominator] when the denominator is exactly "0",
//     however large the numerator of that shape is, as in "C+1/0" or
//     "C+99999999999999999999/0";
//   - [alteration.ErrOutOfRange] when the denominator is not "0" and the
//     numerator or the denominator does not fit in an int, or the exact value
//     is outside the guaranteed range of package alteration, as in
//     "C+16777217" or "C+1/65537".
//
// An input with the denominator "0" never matches [alteration.ErrOutOfRange].
// Other reading errors match only ErrInvalidForm. This includes inputs such as
// "C+0/0" and "C+01/0", whose numerator does not have the shape of a number.
//
// In JSON, a name is a JSON string with its machine form: a value, a struct
// field, an array element, and a map key alike, for example
// {"tonic":"C+1","notes":["D-1","unset"]}. This works with both encoding/json
// and encoding/json/v2: values use Spelling.MarshalJSON and
// Spelling.UnmarshalJSON, and map keys use the text methods or JSON strings.
// A JSON string is read by its decoded content, so escaped characters are
// allowed: "C\u002B1" reads as C sharp. JSON null reads as the unset name.
// The order of map keys in the output is chosen by the JSON package, not by
// this package.
//
// # Unset names
//
// The zero value of Spelling, var s spelling.Spelling, is the unset name: it
// means that no name is given, and it is never treated as C or as a natural
// name. It is the only value that is not a valid name. Spelling.IsValid
// reports false for it and true for every other value that the package
// produces.
//
// The parts of the unset name are the zero letter, which is not a valid
// letter, and the natural alteration value. They make sense only together
// with Spelling.IsValid: a check of the alteration part with package
// alteration, such as s.Alteration().IsNatural(), does not show that a name is
// unset. The properties defined by this package take it into account:
// Spelling.IsNatural, Spelling.IsInteger, Spelling.SameLetter,
// Spelling.SameAlteration, and Spelling.Equal report false for it,
// Spelling.Direction returns DirectionUnset, the letter distances are -1,
// and Spelling.Compare sorts it before every valid name. Spelling.Natural and
// Spelling.WithAlteration return the unset name without an error. Raising and
// lowering it with Spelling.Raise, Spelling.Lower, Spelling.RaiseBy,
// Spelling.LowerBy, Spelling.RaiseByAlteration, and
// Spelling.LowerByAlteration, and Spelling.AlterationDifference with the unset
// name as either operand, return an error that matches [letter.ErrInvalid],
// and their Must variants panic with it. No other operation returns an error
// for the unset name: it is written as "unset" with a nil error, and reading
// "unset" or JSON null gives it back.
//
// # Errors
//
// Errors are checked with errors.Is; their text is not part of the API.
//
//   - New with an invalid letter returns an error that matches
//     [letter.ErrInvalid]. So do Spelling.Raise, Spelling.Lower,
//     Spelling.RaiseBy, Spelling.LowerBy, Spelling.RaiseByAlteration,
//     Spelling.LowerByAlteration, and Spelling.AlterationDifference when a
//     name they use is unset. Other operations on the unset name return no
//     error, as described in the Unset names section; in particular, writing
//     it gives "unset" with a nil error.
//   - Operations whose exact alteration value would be outside the guaranteed
//     range of package alteration return errors that match
//     [alteration.ErrOutOfRange].
//   - Parse, Spelling.UnmarshalText, and Spelling.UnmarshalJSON return errors
//     that match ErrInvalidForm for input that is not the machine form of a
//     name, as described in the Machine form section.
//
// Operations that return a name return the unset name together with an error,
// and Spelling.AlterationDifference returns the natural. A large argument is
// not an error by itself: only the exact value of the result is checked.
//
// The functions and methods with an error result have Must variants, except
// the methods of the encoding interfaces (Spelling.MarshalText,
// Spelling.AppendText, Spelling.UnmarshalText, Spelling.MarshalJSON, and
// Spelling.UnmarshalJSON), which encoders and decoders call. The Must variants
// are MustNew, MustParse, Spelling.MustRaise, Spelling.MustLower,
// Spelling.MustRaiseBy, Spelling.MustLowerBy, Spelling.MustRaiseByAlteration,
// Spelling.MustLowerByAlteration, and Spelling.MustAlterationDifference. Each
// returns the same result as its variant with an error result, but instead of
// returning an error it panics with that error; to read the machine form
// directly with a panic on error, use MustParse. Must variants are intended
// for literals and values known to be valid and in range, for example
// spelling.C().MustRaise().MustRaiseBy(2); use the variants with an error
// result for values taken from data.
//
// Calling a method through a nil *Spelling is a caller error, as with any Go
// pointer, and may panic; in particular, Spelling.UnmarshalText and
// Spelling.UnmarshalJSON need a non-nil receiver. Apart from such calls and
// the Must variants, no function or method panics.
//
// # Values and concurrency
//
// Spelling is a small comparable value type: it can be copied freely, a copy
// shares no data with the original, and it can be used as a map key or in a
// switch. All methods have value receivers and return new values, so the usual
// style is s = s.MustRaise(). The only exceptions are Spelling.UnmarshalText
// and Spelling.UnmarshalJSON, which have pointer receivers and write the name
// they read into the receiver.
//
// The package has no mutable state, and the ready-made names cannot be
// changed. All functions and methods are safe for concurrent use by multiple
// goroutines; as with any Go value, a caller that writes a variable with
// Spelling.UnmarshalText or Spelling.UnmarshalJSON while another goroutine
// uses it must synchronize them.
package spelling
