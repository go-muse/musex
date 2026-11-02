// Package alteration provides alteration values: the amount, in chromatic
// units, by which an accidental raises or lowers a letter.
//
// An [Alteration] is an exact rational number of chromatic units. It can be
// a whole number, such as +1 for a sharp or -2 for a double flat, or a
// fraction, such as +1/2 for a quarter tone up. Five traditional values are
// provided: [Natural] (0), [Sharp] (+1), [Flat] (-1), [DoubleSharp] (+2), and
// [DoubleFlat] (-2). Apart from the guaranteed range described below, the
// number of accidentals is not limited: +3 and -7/2 are alteration values
// too.
//
// The package defines the following operations:
//
//   - [FromInt] and [FromFraction] create a value from a whole number or from
//     a fraction; the fraction is stored in lowest terms, so 2/4 and 1/2 are
//     the same value.
//   - [Alteration.Raise] and [Alteration.Lower] change a value by one
//     chromatic unit; [Alteration.RaiseBy] and [Alteration.LowerBy] change it
//     by any whole number of units.
//   - [Alteration.Add] and [Alteration.Sub] raise or lower a value by another
//     alteration value, whole or fractional. a.Sub(b) is also the difference
//     between a and b: how much higher a is than b.
//   - [Alteration.Mul] multiplies a value by a whole number. There is no
//     division; use [FromFraction] for a value such as n/c.
//   - [Alteration.Equal] and [Alteration.Compare] compare values by size;
//     [Alteration.Sign], [Alteration.IsNatural], and [Alteration.IsInteger]
//     report the direction of a value and whether it is natural or whole.
//   - [Alteration.Fraction] returns the exact value as a fraction in lowest
//     terms.
//
// Operations return new values; values are never modified in place, so the
// usual style is a = a.MustRaise().
//
// An alteration value carries no letter, octave, pitch, tuning, cent offset,
// or notation symbol, and it does not know which values a musical system
// allows: checking that is the job of the system that uses the value.
// [Alteration.String] returns a fixed form, such as "0", "+1", or "-3/2", for
// output and debugging; it is not a notation and has no parser.
//
// The zero value of [Alteration] is the natural, the value 0. It equals
// [Natural], and every operation accepts it.
//
// # Range and errors
//
// Every value of the type lies in the guaranteed range: a value p/q in lowest
// terms, with q > 0, is in range if |p/q| <= [MaxMagnitude] (2^24, that is,
// 16777216) and q <= [MaxDenominator] (2^16, that is, 65536). Within the
// range all results are exact, with no rounding and no overflow, for any int
// argument, including [math.MinInt] and [math.MaxInt].
//
// Operations whose exact result would fall outside the range return the zero
// value and an error that matches [ErrOutOfRange] with [errors.Is].
// [FromFraction] with a zero denominator returns the zero value and an error
// that matches [ErrZeroDenominator]. Each returned error matches exactly one
// of the two; the text of the errors is not part of the API. Comparisons and
// the other operations without an error result never fail.
//
// Each operation with an error result has a Must variant, such as
// [MustFromFraction] or [Alteration.MustRaise], which returns the same value
// but panics with the same error in the same cases. Must variants are
// intended for literals and values known to be in range; use the variants
// with an error result for values taken from data. No other function or
// method panics.
//
// # Concurrency
//
// Alteration values are immutable, and the package has no mutable state. All
// functions and methods are safe for concurrent use by multiple goroutines.
package alteration
