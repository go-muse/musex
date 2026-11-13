// Package duration provides exact musical time: durations and positions, both
// measured in whole notes. A [Duration] is a length greater than zero, such as
// 3/8 for a dotted quarter, and a [Position] is an offset of at least zero from
// an origin chosen by the caller, such as 3/8 after the start of a voice. The
// name of the package covers both quantities, not durations alone.
//
// The package is the common unit of musical time for meters, measures,
// tempos, tick conversion, events, voices, tracks, and scores. It does not
// know notes, rests as elements, measures, meters, tempos, seconds, or the
// units of notation formats; those belong to the packages that use it. Rests
// have the same durations as notes, so the package does not distinguish them.
//
// # Five types
//
//   - [Duration] is an exact length without notation.
//   - [Base] is a note value without dots: one of 14 values from maxima to
//     1024th.
//   - [Written] is a written duration: a base, a number of dots, and a chain
//     of tuplets.
//   - [Tied] is a tied duration: two or more written durations under a tie.
//   - [Position] is an exact offset of at least zero.
//
// A written or tied duration has a duration, returned by Written.Duration and
// Tied.Duration. Adding durations gives a duration, not a notation: notation
// changes only by an explicit operation whose result the caller receives.
//
// # Two kinds of equality
//
// Durations and positions are values: Duration.Equal and Position.Equal
// compare values, and New(6, 16) equals New(3, 8) because both are stored in
// reduced form. Written and tied durations are notations: Written.Equal and
// Tied.Equal compare the notation, so a quarter in a triplet 3:2 differs from
// a quarter in a sextuplet 6:4, and the chains [3:2, 5:4] and [5:4, 3:2]
// differ, although their durations are equal. Written.EqualDuration,
// Written.CompareDuration, Tied.EqualDuration, and Tied.CompareDuration
// compare durations; a written and a tied duration, such as a dotted half and
// a half tied to a quarter, are compared through their Duration methods. This
// differs from package common/fraction, where Fraction.Equal compares values
// and the unset fraction is not equal even to itself; here every unset value
// is equal to the unset value of its type.
//
// Every Compare method returns −1, 0, or +1 and suits slices.SortFunc.
// Written.Compare orders by duration, then by the number of tuplets (fewer
// first), the base (longer first), the number of dots (fewer first), and the
// tuplets in order; Tied.Compare orders by duration, then by the number of
// parts and the parts in order. Compare returns 0 exactly for equal values.
//
// # Bases
//
// The series of bases runs from [Maxima] (8 whole notes) and [Longa] (4) through
// [Breve] (2), [Whole], [Half], [Quarter], and so on to [ThousandTwentyFourth]
// (1/1024); each base is twice as long as the next. The values of longa and
// maxima are a convention of the product that continues this doubling; mensural
// relations are not modeled. [Bases] lists the series, [BaseOf] finds the base
// of a duration, and Base.Longer, Base.Shorter, and Base.Shift move along the
// series.
//
// Base.Name, Base.BritishName, and Base.RussianName return names only from the
// reference table of the package: "quarter note", "crotchet", "четвертная".
// Where the table has no name, as for the 128th in English or the 256th in any
// language, they return ok == false without an error: "no name" is a result,
// not a failure. The identifiers of the constants, such as
// HundredTwentyEighth, are Go names, not names of the values.
//
// # Dots and tuplets
//
// A base b with k dots lasts b·(2 − 1/2^k): a dotted quarter lasts 3/8. The
// number of dots is not limited by a rule, only by the range of package
// common/fraction; from 512 dots on the duration is always outside that range.
//
// A chain lists tuplets of package tuplet from the outermost to the innermost
// and multiplies the duration by its coefficient: an eighth in a triplet 3:2
// lasts 1/12. The chain is stored as given, without reduction or reordering,
// and a tuplet 1:1 is kept although it does not change the duration.
// Written.Coefficient returns the coefficient with the unreduced pair of
// tuplet.ChainCoefficient.
//
// # Notation of a duration
//
// Duration.ToWritten and Tied.ToWritten write a duration as one sign under a
// chain chosen by the caller: 3/8 gives a dotted quarter, and 1/3 under a
// triplet gives a half in 3:2. Such a sign is unique when it exists. The
// package never chooses a tuplet itself.
//
// Duration.NotateWith writes a duration as signs to be tied. The allowed signs
// are the bases from a given longest base down to the 1024th with at most a
// given number of dots. Each next part is the longest allowed sign not longer
// than the remainder, so the parts go from long to short, and every part
// carries the chain as given. Duration.Notate uses the defaults: bases from
// the breve down, without dots, so 5/16 gives a quarter and a sixteenth and 5
// gives two breves and a whole. If the duration is one allowed sign, the result
// is that sign alone; two or more parts make a Tied with [NewTied]. The result
// has at most [MaxParts] (1024) parts; a longer result fails with
// [ErrTooManyParts] without being built. [ErrNotNotatable] means that the rule
// gives no notation, not that none exists: 1/512 with bases down from the
// 1024th and one dot fails although two tied 1024ths make it. Splitting
// durations at bar lines and beats is the job of the measure packages.
//
// # Normal subdivision
//
// Written.Subdivide divides a base without dots into 2 parts and a base with
// one dot into 3 parts, each the base one step shorter, under the same chain:
// a quarter gives 2 eighths and a dotted half 3 quarters. Two or more dots
// have no rule ([ErrNoSubdivision]), and parts shorter than the 1024th are
// outside the series ([ErrOutsideSeries]).
//
// # Positions
//
// The origin of positions is chosen by the consumer, such as the start of a
// voice or a measure; [Origin] returns zero, which is a valid position. A
// position plus a duration is a position. Position.OffsetTo returns the
// offset between two positions as a position, which may be zero, and
// Position.DurationTo returns the duration between them, which must be
// positive.
//
// # Quantization
//
// Duration.Quantize, Position.Quantize, and [QuantizeSpan] round only when
// called. The grid consists of the multiples k·step of a positive step,
// starting at zero; a grid shifted by a start position is obtained by
// subtracting the start first. The mode is one of the nine modes of package
// common/rounding and is always given explicitly: there is no default. A value
// on a node is unchanged in every mode. Since durations and positions are not
// negative, TowardZero equals Down, AwayFromZero equals Up, HalfTowardZero
// equals HalfDown, and HalfAwayFromZero equals HalfUp. QuantizeSpan quantizes
// the start and the end of a span with the same grid and mode, so the end does
// not accumulate the rounding of the length. A duration quantized to zero
// fails with [ErrQuantizedToZero]. The results are durations and positions,
// not notations.
//
// # Unset values
//
// The zero values Duration{}, Written{}, Tied{}, and Position{} are unset;
// they are neither zero nor the origin. Only IsSet, Equal, Compare, and String
// accept them without an error; String returns "Duration(unset)",
// "Written(unset)", "Tied(unset)", or "Position(unset)". Every other operation
// whose receiver, argument, or list element is unset returns an error wrapping
// [ErrUnset]. A [Base] has no unset state: Base(0) and values above
// ThousandTwentyFourth are not bases and give [ErrInvalidArgument].
//
// # Text
//
// String, MarshalText, and AppendText write a canonical text, and [Parse],
// [ParsePosition], [ParseWritten], and [ParseTied] read exactly that grammar:
//
//	duration = count "/" count                   ; such as 3/8
//	position = ( "0" / count ) "/" count         ; such as 0/1
//	written  = base *"." [ "[" tuplet *( "," tuplet ) "]" ]
//	tied     = written 1*( "~" written )
//	base     = "maxima" / "longa" / "breve" / numeral
//	numeral  = "1" / "2" / "4" / "8" / "16" / "32" / "64" / "128" / "256" / "512" / "1024"
//	tuplet   = count ":" count
//	count    = nzdigit *digit
//
// So a dotted quarter is "4.", an eighth in a triplet "8[3:2]", a dotted
// quarter in a quintuplet holding a triplet "4.[5:4,3:2]", and a half tied to
// a quarter "2~4". Durations and positions print in reduced form, but an
// unreduced pair such as "6/16" is accepted. Spaces, signs, leading zeros,
// other case, and names are rejected, and so are the strings of unset values.
// For every set value x, parsing x.String() returns x. [ParseBase] separately
// accepts the text of a base, the LilyPond notations \maxima, \longa, and
// \breve, and the English and Russian names, ignoring case.
//
// Every error of a parse function, of its Must variant, and of UnmarshalText
// with a non-nil receiver quotes the rejected text: in full if it is at most
// 64 bytes long, otherwise its longest prefix of whole UTF-8 characters of at
// most 64 bytes followed by "...". The message consists only of the texts of
// the wrapped sentinels and this excerpt, is at most 512 bytes long, and for a
// given operation and error depends on the input only through its first 67
// bytes, however long the input is. Errors of the dependencies with their own
// quotations are not included.
//
// The types implement encoding.TextMarshaler, encoding.TextAppender, and
// encoding.TextUnmarshaler, so struct fields are encoded in JSON as strings
// such as "4." or "3/8". Encoding an unset value fails with [ErrUnset], so an
// optional field needs the omitzero tag option.
//
// # Errors
//
// Every error matches exactly one of eleven sentinels via [errors.Is]:
// [ErrUnset], [ErrInvalidArgument], [ErrNotPositive], [ErrNegativePosition],
// [ErrNotNotatable], [ErrOutsideSeries], [ErrNoSubdivision],
// [ErrQuantizedToZero], [ErrTooManyParts], [ErrInvalidText], and
// [ErrOutOfRange]. When the cause is a failure of package common/fraction,
// tuplet, or common/rounding, the error also wraps the sentinel of that
// package: a zero denominator also matches fraction.ErrZeroDenominator, an
// unset tuplet tuplet.ErrUnset, an unset or unknown mode rounding.ErrUnset or
// rounding.ErrUnknownMode, and a result outside the range
// fraction.ErrOutOfRange. When several conditions fail, the first one in the
// documented order of the operation is reported. A failing call returns zero
// results, except that AppendText returns its argument unchanged and
// UnmarshalText leaves its receiver unchanged. Results are exact: they are
// never rounded or truncated, and a result outside the range of package
// common/fraction fails with [ErrOutOfRange].
//
// # Must variants
//
// Every fallible function and method X, except MarshalText, AppendText, and
// UnmarshalText, has a variant MustX that returns the same results without the
// error and panics with the same error value on failure. Must variants allow
// chains such as duration.Quarter.MustWritten().MustAddDot().MustDuration().
//
// # Concurrency and comparison
//
// All values are immutable, and the package has no mutable state, caches, or
// goroutines. No operation modifies its receiver or arguments, except
// UnmarshalText for its receiver; slices passed in are only read, and slices
// returned are new and belong to the caller. All functions and methods are
// safe for concurrent use without synchronization, except concurrent calls to
// UnmarshalText writing through a pointer while the value is read. All five
// types are comparable with ==, which matches their Equal methods, and can be
// used as map keys; structures that contain them stay comparable.
package duration
