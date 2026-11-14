// Package meter provides time signatures: a [Meter] is a time signature as a
// record (the signature, its form, and its grouping) and as the structure of
// one full measure (its size, units, groups, beats, kind, and strong beats).
// The package does not know the sequence of measures, pickups, or bar
// numbers (package measure), the tempo (package tempo), notes, or the units
// of notation formats.
//
// The name of the package means the time signature, not meter as the topic of
// the reference base; the measure grid is package measure. The names of the
// packages do not follow the topics of the same names.
//
// # Units and beats
//
// The numerator counts units: the unit of 6/8 is the eighth, and 6/8 has six
// units. A beat is the beat of the English-language tradition: 6/8 has two
// beats of a dotted quarter. Both are available under different names:
// Meter.Units and the methods named Unit count units, and the methods named
// Beat count beats. The beats are the units for the standard simple
// structure (two, three, or four equal units in their standard grouping, as
// in 2/4, 3/4, 4/4, 2/2, 3/8, C, and ¢) and the groups otherwise: 5/8[2+3]
// has two beats, a quarter and a dotted quarter, and five units. A meter
// without a grouping, such as 5/8, has units but no beats.
//
// A group is several consecutive units that form one metric cell; the
// grouping is the sizes of the groups in units, such as 2+3. Every meter
// without an explicit grouping has the standard grouping of its numerator, if
// there is one: one group for 1, 2, and 3 units, 2+2 for 4 units, and groups
// of three for multiples of three from 6 (6/8 is 3+3, 15/8 is 3+3+3+3+3).
// Other numerators, such as 5, 7, 8, and 10, have no standard grouping. An
// explicit grouping is set by Meter.WithGrouping on a simple signature or a
// symbol; a grouping equal to the standard one gives the same value.
//
// # Four forms
//
//   - A simple signature, such as 3/4 ([New]).
//   - A symbol: C with the structure of 4/4 and ¢ (alla breve) with the
//     structure of 2/2 ([NewSymbol]). 4/2 is written with digits.
//   - A compound numerator, such as (3+2+3)/8 ([NewAdditive]); its groups are
//     its addends.
//   - A sum of signatures, such as 2/4+3/8 ([NewSum]); its units and groups
//     are those of its parts in order. A sum with different denominators has
//     no common denominator and no single unit.
//
// Any integer denominator of at least 1 is accepted, including 1 and numbers
// that are not powers of two, such as 4/3. Fractional numerators, such as
// 2½/4, are not supported: the same measure is 5/8 or (2+2+1)/8. A meter has
// at most [MaxUnits] (1024) units.
//
// A unit, a beat, or a group has a written duration only when its size is
// written as a single sign without tuplets: the unit of 4/3 has none, while
// the beats of 6/6 are two halves.
//
// # Three kinds of equality and the order
//
// Meter.Equal compares records and is the same as ==: C and 4/4 differ, and
// so do (2+3)/8, 5/8[2+3], and 2/8+3/8. Meter.EqualStructure compares the
// units and the grouping: all three of these are equal by structure, and so
// are C and 4/4. Meter.EqualDuration compares sizes: 3/4 and 6/8 are equal by
// size and differ by record and by structure. Meter.Compare orders by size
// and, at an equal size, by the canonical text; it suits slices.SortFunc.
//
// # Kinds in two traditions
//
// Meter.Kind classifies a meter in the English-language tradition: simple,
// compound, asymmetrical, or undetermined. Meter.RussianKind classifies it in
// the Russian tradition: simple, compound uniform, compound mixed, or
// undetermined. The two traditions are independent and use different rules;
// their words "simple" and "compound" mean different things. Undetermined is
// a result, not an error: 1/4 is undetermined in both. The beat divides into
// two and subdivides into four in a simple kind, and into three and six in a
// compound kind (Meter.BeatDivision); other kinds have no beat division.
//
// # Strong beats
//
// The first unit of the first group is strong, the first units of the other
// groups are relatively strong, and the other units are weak. This rule of
// the product applies to any grouping; a deeper hierarchy is not modeled. A
// position that is not the start of a unit has the weight off-unit. In these
// names a beat is a unit, as in the Russian tradition.
//
// # Positions
//
// Numbers of units, groups, and beats count from 1; offsets from the start of
// the measure count from 0. A position at the start of an element belongs to
// it; a position equal to the size of the meter is outside the meter.
//
// # Changing a meter
//
// Meter.ToNumeric, Meter.ToSymbol, Meter.ToSimple, Meter.ToAdditive, and
// Meter.ToSum change the record and keep the structure. Meter.Shorter,
// Meter.Longer, and Meter.WithDenominator change the unit. [FromDuration]
// builds a simple signature of a given size. [GroupingVariants] lists the
// ways to split a number of units into groups of two and three.
//
// # Names
//
// The name methods return names only from the reference table of the
// package, such as "simple triple", "сложный двухдольный", "cut time", and
// "alla breve". Where the table has no name, they return ok == false without
// an error: "no name" is a result, not a failure. The identifiers of the
// constants are Go names, not names of the values: RussianCompoundUniform has
// no English name.
//
// # Text
//
// Meter.String, Meter.MarshalText, and Meter.AppendText write the canonical
// text without spaces: "3/4", "5/8[2+3]", "C", "C|", "C[3+1]", "(3+2+3)/8",
// "2/4+3/8". The symbol ¢ is written as "C|", as in ABC notation; [Parse]
// also accepts "¢". Texts of meters that differ by record differ, and parsing
// the text of a set meter returns an equal meter. The unset meter has the
// string "Meter(unset)", which Parse rejects; MarshalText fails for it, so an
// optional meter field of a struct takes the omitzero tag in JSON. A parse
// error quotes at most 64 bytes of the rejected text, and its message is at
// most 512 bytes long.
//
// # Errors
//
// Every error wraps exactly one of ten sentinels: [ErrUnset],
// [ErrInvalidArgument], [ErrGroupingRequired], [ErrNoCommonDenominator],
// [ErrUndefined], [ErrNotNotatable], [ErrOutsideMeter], [ErrTooMany],
// [ErrInvalidText], and [ErrOutOfRange]. When the cause is a failure of a
// dependency, the error also wraps its sentinel: duration.ErrNotNotatable
// with ErrNotNotatable, duration.ErrUnset with ErrInvalidArgument for an
// unset duration or position, fraction.ErrOutOfRange or
// fraction.ErrNotRepresentable with ErrOutOfRange, and also
// duration.ErrOutOfRange with ErrOutOfRange when the failure comes from
// package duration, as in Meter.QuoRem and Span.End. Each operation documents
// its errors in the order of its checks. Every operation that returns an
// error, except MarshalText, AppendText, and UnmarshalText, has a Must
// variant that panics with the same error.
//
// # Values and concurrency
//
// All types are comparable with == and usable as map keys. Values are
// immutable: operations return new values and new slices, and the slices
// passed to the package are only read during the call. The package has no
// mutable state, and all its functions and methods are safe for concurrent
// use, except that UnmarshalText must not write a meter that is read at the
// same time.
package meter
