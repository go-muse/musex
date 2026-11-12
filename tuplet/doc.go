// Package tuplet provides exact tuplets: groups of n equal notes played in the
// time of m notes of the same written value, such as duplets, triplets,
// quintuplets, and sextuplets.
//
// The package describes a tuplet by its two counts and computes its action on
// exact durations of package common/fraction without rounding. It is the
// basis of note durations, and through them of chords and tracks, and serves
// score editors, notation format exchange, rhythm generators, and
// quantization.
//
// # Actual and normal counts
//
// A [Tuplet] n:m stores the actual count n (how many notes sound, MusicXML
// actual-notes, MEI num) and the normal count m (how many notes of the same
// written value they replace, MusicXML normal-notes, MEI numbase). Both counts
// are at least 1 and at most math.MaxInt. A triplet 3:2 plays three notes in
// the time of two, and a duplet 2:3 plays two notes in the time of three.
//
// [New] creates any pair. [Duplet], [Triplet], [Quadruplet], [Quintuplet],
// [Sextuplet], [Septuplet], [Octuplet], [Nonuplet], and [Decuplet] return the
// named tuplets 2:3, 3:2, 4:3, 5:4, 6:4, 7:4, 8:6, 9:8, and 10:8.
// [FromCoefficient] and [FromReciprocalCoefficient] create a tuplet from a
// coefficient, and [FromUnitRatio] converts a notation with different units
// of the actual and normal parts, such as three eighths in the time of one
// quarter.
//
// The tuplet does not store the written note value of its parts, the meter,
// the inner grouping of its notes (2+2+2 or 3+3 in a sextuplet), or any
// graphics: these belong to the packages that use it.
//
// # Coefficients
//
// The duration coefficient of n:m is normal/actual, m/n: a written duration
// multiplied by the coefficient is the sounding duration of each note under
// the tuplet. The coefficient of a triplet 3:2 is 2/3 and of a duplet 2:3 is
// 3/2. The reciprocal coefficient n/m is the "speed" factor written by
// LilyPond as \tuplet 3/2.
//
// [Tuplet.Coefficient] and [Tuplet.ReciprocalCoefficient] keep the pair of the
// notation unreduced: the coefficient of 6:4 has the stored pair (4, 6).
// [Tuplet.ReducedCoefficient] returns the reduced form 2/3. [Tuplet.Apply]
// multiplies a duration by the coefficient, [Tuplet.Unapply] undoes it, and
// [Tuplet.Span] returns the total duration m·u of a tuplet written with parts
// of duration u; their results are fully reduced. Any set fraction is
// accepted, including zero and negative values such as offsets. To express a
// result in a particular denominator, use the denominator modes of package
// common/fraction.
//
// # Notation, equality, and order
//
// The pair is stored as given: 3:2 and 6:4 are different notations with equal
// coefficients. Two equality relations are provided:
//
//   - [Tuplet.Equal] compares the notation; it matches the built-in ==
//     operator, so a Tuplet can be used as a map key with 3:2 and 6:4 as
//     different keys.
//   - [Tuplet.EqualCoefficient] compares coefficient values: 3:2 and 6:4 are
//     equal.
//
// This is the opposite of package common/fraction, where Fraction.Equal
// compares values and Fraction.EqualExact compares the stored pair, and where
// the invalid fraction is not equal even to itself. Here the unset tuplet is
// equal to the unset tuplet.
//
// [Tuplet.CompareCoefficient] compares coefficient values. [Tuplet.Compare]
// is a total order for sorting: by coefficient value, then by actual count,
// then by normal count, so 3:2 precedes 6:4. [Tuplet.Reduce] and
// [Tuplet.Scale] change the notation without changing the coefficient value.
// Coefficient comparisons are exact for all counts up to math.MaxInt.
//
// # Kinds and neutral tuplets
//
// A tuplet is [Compressing] if n > m (notes shorter than written, as in a
// triplet), [Expanding] if n < m (notes longer than written, as in a duplet),
// and [Neutral] if n == m. Pairs with n == m, including 1:1, are accepted as a
// product decision: nesting a duplet 2:3 inside a triplet 3:2 gives 6:6, and
// an empty chain gives 1:1. Musical sources describe only tuplets with n ≠ m,
// and the package does not claim that neutral tuplets occur in practice.
//
// # Divisions and default normal counts
//
// A [Division] is the kind of normal division of the replaced value: [Binary]
// for a plain value divided into 2, 4, 8, … parts and [Ternary] for a dotted
// value divided into 3, 6, 12, … parts. The package does not model meters;
// the division is passed explicitly.
//
// [Default] returns the tuplet n:m whose normal count m is the largest member
// of the division's series that is less than n, except that the ternary
// default for n = 2 is 3. There is no default for n = 1 or for an n that
// itself belongs to the series, since such a division is normal rather than a
// tuplet. Binary defaults are 3:2, 5:4, 6:4, 7:4, 9:8; ternary defaults are
// 2:3, 4:3, 5:3, 7:6. The named tuplets use the binary default when it exists
// and the ternary default otherwise.
//
// This rule is a product choice. It follows the "nearest smaller normal
// value" rule of the English-language literature for binary division and
// generalizes it to ternary division. It differs from part of the sources:
// some Russian sources write a septuplet as 7:8, abc notation 2.1 uses other
// defaults for 5 through 9 (such as 5:2 and 6:2), and no source confirms the
// defaults 8:6, 9:8, and 10:8. Use [New] for any other pair.
//
// # Names
//
// [Tuplet.Name] and [Tuplet.RussianName] name a tuplet by its actual count
// alone, for counts 2 through 10 only: duplet (дуоль), triplet (триоль),
// quadruplet (квартоль), quintuplet (квинтоль), sextuplet (секстоль),
// septuplet (септоль), octuplet (октоль), nonuplet (нонемоль), and decuplet
// (децимоль). So 5:3 is also a quintuplet. Duplets and triplets in Russian
// usage are the same concept under the names on "-оль"; the Russian forms
// "дуплет" and "триплет" are not used.
//
// # Nesting
//
// A chain lists nested tuplets from the outermost to the innermost.
// [Tuplet.Nest] and [Flatten] return the flattened notation
// (n1·n2·…):(m1·m2·…), [ChainCoefficient] returns the total coefficient with
// the unreduced stored pair (m1·m2·…, n1·n2·…), and [ApplyChain] applies it
// to a duration. The flattened notation matches the chain only when all its
// tuplets count parts of the same unit. An empty chain gives 1:1 and 1/1.
// Results and failures do not depend on the order of the chain.
//
// Flattened counts are limited to math.MaxInt; the chain coefficient is
// limited to the range of package common/fraction (components up to
// 2^255 - 1), so any chain of four tuplets has a coefficient. Exceeding a
// limit returns an error wrapping [ErrOutOfRange].
//
// # Text
//
// [Tuplet.String] returns the canonical notation "n:m", and
// [Tuplet.ShortString] returns "n". [Parse] accepts only the canonical
// notation: ASCII decimal digits without signs, spaces, leading zeros, or zero
// values, separated by one colon. [ParseShort] accepts "n" and a division and
// returns the default tuplet. [ParseName] accepts the English, alternative
// English, and Russian names in lower case. The unset tuplet prints as
// "Tuplet(unset)", which Parse rejects. Parse(t.String()) returns t for every
// set tuplet. The French "pour" notation, LilyPond n/m, abc (p:q:r, MusicXML,
// and MEI are not parsed; the pair maps directly to their numbers.
//
// An error of Parse, ParseShort, ParseName, their Must variants, or
// UnmarshalText with a non-nil receiver quotes the rejected text: in full if
// it is at most 64 bytes long, otherwise by its longest prefix of whole UTF-8
// characters (or single invalid bytes) of at most 64 bytes followed by "...".
// The length of the text is not included, so the message is at most 512 bytes
// long however long the text is.
//
// Tuplet implements encoding.TextMarshaler, encoding.TextAppender, and
// encoding.TextUnmarshaler with the canonical notation, so a Tuplet struct
// field is encoded in JSON as a string such as "3:2". Encoding an unset
// tuplet fails with [ErrUnset]; use the omitzero tag option for an optional
// field.
//
// # Unset tuplet
//
// The zero value Tuplet{} is the unset tuplet; it is not 1:1. Only
// [Tuplet.IsSet], [Tuplet.Equal], [Tuplet.Compare], and [Tuplet.String]
// accept it without an error. Every other operation whose receiver, argument,
// or chain element is unset returns an error wrapping [ErrUnset].
//
// # Must variants
//
// Every fallible function and method X, except the standard interface methods
// MarshalText, AppendText, and UnmarshalText, has a variant MustX that
// returns the same results without the error and panics with the same error
// value on failure. Must variants allow chains such as
// tuplet.MustNew(6, 4).MustReduce().MustCoefficient().
//
// # Errors
//
// Every error matches exactly one of seven sentinels via [errors.Is]:
// [ErrInvalidCount], [ErrUnset], [ErrNoDefault], [ErrInvalidText],
// [ErrOutOfRange], [ErrInvalidArgument], and [ErrNotRepresentable]. When the
// cause is an error of package common/fraction, the error also wraps that
// cause: errors.Is(err, tuplet.ErrOutOfRange) and
// errors.Is(err, fraction.ErrOutOfRange) both hold when an applied duration
// leaves the fraction range. When several conditions fail, the first one in
// the documented order of the operation is reported. Error values return zero
// results.
//
// [ErrNotRepresentable] of this package and fraction.ErrNotRepresentable are
// different: [FromUnitRatio] returns ErrNotRepresentable when normal·ratio is
// not an integer, and ErrOutOfRange wrapping fraction.ErrNotRepresentable
// (but not tuplet.ErrNotRepresentable) when normal·ratio is an integer
// greater than math.MaxInt.
//
// # Concurrency
//
// Tuplet values are immutable, and the package has no mutable global state,
// caches, or goroutines. No operation modifies its receiver or arguments,
// except [Tuplet.UnmarshalText] for its receiver. All functions and methods
// are safe for concurrent use without synchronization, except concurrent
// calls to UnmarshalText writing through the same *Tuplet while it is read.
package tuplet
