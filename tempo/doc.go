// Package tempo provides tempo markings and the time they give at a constant
// tempo: a written unit and beats per minute, such as ♩. = 60, the seconds of
// a duration, the formula of bars, units, beats per minute, and time, the
// metronome scale, and tempo indications that combine a marking with a term
// of package tempo/term.
//
// The package converts musical time of package duration into seconds while
// the tempo does not change. Changes of tempo, gradual changes, and fermatas
// belong to the package of tempo maps, ticks of MIDI and MusicXML to the
// package of ticks; meters and measures are not known here: the length of a
// bar comes to the formula as a duration.
//
// # Markings
//
// A [Tempo] is an exact marking: a unit u, any written duration of one sign
// with dots and tuplets, and beats per minute N, any exact positive fraction.
// [NewQuarter] uses the unit of a quarter, which a number of beats per minute
// alone means. Beats per minute need not be whole or lie on the metronome
// scale: ♩ = 1/1000 and ♩ = 100000 are valid. The speed of a marking is u·N
// whole notes per minute.
//
// An [Approx] is an approximate marking, such as ♩ ≈ 108, and a [Range] is a
// tempo range, such as ♪ = 132–144. Both are kinds of their own: they never
// equal an exact marking and never turn into one implicitly. Approx.Tempo
// takes the named number, and Range.LowerTempo, Range.UpperTempo, and
// Range.MidTempo take a bound or the exact middle without rounding; a rounded
// middle comes from Tempo.Quantize. Both numbers of a range belong to it, a
// convention of the product: Range.Contains compares a number of beats, and
// Range.ContainsTempo compares speeds, so ♩ = 69 lies in ♪ = 132–144. A range
// always has both bounds, the lower less than the upper.
//
// # Two kinds of equality
//
// Equal compares the record: units equal as duration.Written.Equal compares
// them and equal numbers, so ♩ = 60 equals ♩ = 120/2. EqualSpeed and
// CompareSpeed compare speeds: ♩. = 60, ♪ = 180, and ♩ = 90 are different
// records of the same speed. Compare is a total order for sorting: by speed,
// and for equal speeds by unit as duration.Written.Compare orders them, so
// ♪ = 180 < ♩ = 90 < ♩. = 60; ranges are ordered by the speeds of the lower
// and then the upper bound. Tempo.WithUnit converts a marking to another unit
// of the same speed, Tempo.Mul, Double, and Halve change the speed, and
// Tempo.Modulate performs a metric modulation "new v = old w".
//
// # Time and system time
//
// [Seconds] is an exact time of at least zero seconds. Tempo.Seconds gives
// the time of a duration, 60·d/(u·N); Tempo.DurationIn, [FromSeconds],
// Tempo.UnitsIn, and Tempo.SecondsOfUnits answer the reverse questions. Every
// result is exact. Seconds.TimeDuration rounds a time to whole nanoseconds
// with one of the nine modes of package common/rounding; there is no default
// mode, and a whole number of nanoseconds is unchanged in every mode. Since
// times are not negative, TowardZero equals Down, AwayFromZero equals Up,
// HalfTowardZero equals HalfDown, and HalfAwayFromZero equals HalfUp. The
// largest system time is math.MaxInt64 nanoseconds, about 292 years; a
// longer time fails with [ErrTimeOverflow] and stays exact in Seconds.
// Rounding happens only there, in Tempo.Quantize, and on the metronome scale.
//
// # The formula
//
// n·r = N·t with r = L/u: n bars of length L, each holding r units of
// duration u, last t minutes at N beats per minute; t is the time in seconds
// divided by 60. The formula holds only for a constant tempo and equal full
// bars. The package does not check this condition, as it knows no bars. The
// methods of Tempo solve it for a known unit and beats per minute, and
// [SolveBars], [SolveBarUnits], [SolveBPM], and [SolveSeconds] find any of
// the four numbers from the other three; [BarUnits], [BarLength], and
// [UnitLength] relate a bar, a unit, and r. The number of bars may be
// fractional when the time does not end on a bar line.
//
// # The metronome scale
//
// The scale of Maelzel has 39 divisions from [ScaleMin] (40) to [ScaleMax]
// (208): by 2 to 60, by 3 to 72, by 4 to 120, by 6 to 144, and by 8 to 208,
// numbered from 0. A division is a number of beats per minute without a
// duration; a marking keeps its unit. [ScaleNeighbors] returns the two
// neighboring divisions of a number and which is nearer, and
// [NearestScaleDivision] and Tempo.SnapToScale choose one by a mode. These
// rules are conventions of the product: nearness is measured by difference,
// so 110 is equally near to 108 and 112; HalfEven breaks a tie by the parity
// of the index of the lower neighbor, so 110 gives 112 because 108 has index
// 23; and a number outside 40..208 fails with [ErrOffScale] instead of being
// moved to the nearest end of the scale. [NextScaleDivision] and
// [PrevScaleDivision] return the end of the scale for a number beyond it in
// the given direction.
//
// # Terms and indications
//
// Tempo.TermsIn and Approx.TermsIn find the terms of a list of package
// tempo/term whose guides contain the beats per minute of a marking; the unit
// is not taken into account. [RangeOf] turns the guide of a term with both
// bounds into a range with a unit chosen by the caller.
//
// An [Indication] is a term, a marking of any kind, or both, as written above
// a score: "Allegro, ♩ = 120". Whether the number fits the guide of the term is
// not checked. Seconds come only from an exact marking, read with
// Indication.Tempo.
//
// # Text
//
// String, MarshalText, and AppendText write a canonical text, and [Parse],
// [ParseApprox], [ParseRange], [ParseSeconds], and [ParseIndication] read it:
//
//	tempo      = unit "=" number                 ; such as 4=60 or 4.=60
//	approx     = unit "≈" number                 ; such as 4≈108
//	range      = unit "=" number "-" number      ; such as 8=132-144
//	seconds    = integer "/" count               ; such as 0/1 or 1/3
//	indication = name / mark / name ", " mark    ; such as Allegro, 4=120
//	mark       = tempo / approx / range
//	unit       = written                         ; the grammar of duration
//	number     = integer / ratio / decimal
//	integer    = "0" / count
//	ratio      = integer "/" count
//	decimal    = integer "." 1*digit
//	count      = nzdigit *digit
//
// Numbers print as integers or reduced fractions such as 185/2, and parsing
// also accepts a finite decimal such as 92.5. The approximate marking is
// written with "≈" (U+2248) only: the forms "4=~108" and "4~108" are rejected.
// Spaces, signs, leading zeros, and note signs such as ♩ are not part of the
// grammar. A name of an indication is written in Go quotation marks, as
// strconv.Quote writes it, only if it contains "=" or "≈" or starts with a
// quotation mark, so that the term "4=60" of a user list prints as "\"4=60\""
// and differs from the marking 4=60; parsing accepts quotation marks around any
// name. ParseIndication finds the name only in the list it is given, ignoring
// case; Indication does not implement encoding.TextUnmarshaler. For every set
// value x of Tempo, Approx, Range, and Seconds, parsing x.String() returns x,
// and ParseIndication(i.String(), l) returns i for an indication whose term
// is from l.
//
// Every error of a parse function, of its Must variant, and of UnmarshalText
// with a non-nil receiver quotes the rejected text: in full if it is at most
// 64 bytes long, otherwise its longest prefix of whole UTF-8 characters of at
// most 64 bytes followed by "...". The message consists only of the texts of
// the wrapped sentinels and this excerpt, is at most 512 bytes long, and for a
// given operation and error depends on the input only through its first 67
// bytes. Parsing takes time linear in the length of the text.
//
// # Unset values
//
// The zero values Tempo{}, Approx{}, Range{}, Seconds{}, and Indication{} are
// unset; Seconds{} is not zero, which [ZeroSeconds] returns. Only IsSet,
// Equal, Compare, and String accept them without an error: they equal only an
// unset value of their type, precede every set value, and print as
// "Tempo(unset)", "Approx(unset)", "Range(unset)", "Seconds(unset)", and the
// empty string for an indication, which parsing rejects. Every other operation
// whose receiver or argument is an unset value of the package returns an error
// wrapping [ErrUnset]; an unset value of another package, such as an unset
// fraction, duration, term, or list, gives [ErrInvalidArgument]. [Kind] has no
// unset state: its zero value NoMark means that an indication has no marking.
//
// # Errors
//
// Every error matches exactly one of eleven sentinels via [errors.Is]:
// [ErrUnset], [ErrInvalidArgument], [ErrNoTempo], [ErrNoSolution],
// [ErrNegativeTime], [ErrOffScale], [ErrNoRange], [ErrQuantizedToZero],
// [ErrTimeOverflow], [ErrInvalidText], and [ErrOutOfRange]. When the cause is a
// failure of package common/fraction, common/rounding, or duration, the error
// also wraps the sentinels of that package: a zero denominator also matches
// fraction.ErrZeroDenominator, an unset or unknown mode rounding.ErrUnset or
// rounding.ErrUnknownMode, a result or a number outside the range
// fraction.ErrOutOfRange, and a unit in text outside its range every sentinel
// of the error of duration.ParseWritten. The sentinels of package tempo/term
// are not wrapped. When several conditions fail, the first one in the
// documented order of the operation is reported. A failing call returns zero
// results, except that AppendText returns its argument unchanged and
// UnmarshalText leaves its receiver unchanged. Intermediate values are exact,
// and only a result outside the range of package common/fraction fails with
// [ErrOutOfRange].
//
// # Must variants
//
// Every fallible function and method X, except MarshalText, AppendText, and
// UnmarshalText, has a variant MustX that returns the same results without the
// error and panics with the same error value on failure. Must variants allow
// chains such as tempo.MustNewQuarter(f).MustDouble().MustSeconds(d).
//
// # Concurrency and comparison
//
// All values are immutable, and the package has no mutable state, caches, or
// goroutines. No operation modifies its receiver or arguments, except
// UnmarshalText for its receiver, and slices returned are new and belong to
// the caller. All functions and methods are safe for concurrent use without
// synchronization, except concurrent calls to UnmarshalText writing through a
// pointer while the value is read. Tempo, Approx, Range, Seconds, Kind, and
// Indication are comparable with ==, which matches their Equal methods, and
// can be used as map keys; structures that contain them stay comparable.
package tempo
