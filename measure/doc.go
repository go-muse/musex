// Package measure provides measure grids: a [Grid] is the sequence of bars
// (measures) from an origin with their time signatures and lengths — a
// pickup, meter changes, bars of an explicit length, and an optional number
// of bars. It gives bar numbers, lengths, and starts; converts between a
// position and an address "bar and offset"; places a moment in its meter;
// counts units and the MEI tstamp; splits a duration at bar lines and inside
// bars; and changes a grid by explicit operations that check its rules.
//
// The package does not know notes, events, voices, the tempo, seconds, or
// notation formats. The structure of one full bar is the job of package
// meter, lengths and positions are values of package duration, and exact
// results are fractions of package common/fraction.
//
// The name of the package means the measure grid, and the name of package
// meter means the time signature. The names of the packages do not follow
// the topics of the same names in the reference base.
//
// # Bars and their numbers
//
// The pickup (anacrusis) is bar 0, as in MusicXML, and the first full bar is
// bar 1; in music theory the pickup has no number. The ordinal place of a bar
// counts from 1 with the pickup first: [Grid.Ordinal] and
// [Grid.BarAtOrdinal]. A grid without a number of bars is open: its last
// meter continues, and it has bars 1 to math.MaxInt; positions at or after
// the end of bar math.MaxInt are [ErrOutOfRange]. A grid with a number of
// bars N is finite: it has bars 1 to N, and a position at or after its end is
// [ErrOutsideGrid]. N counts numbered bars; the pickup is not counted.
//
// # The record
//
// A grid keeps its record as given. A mark ([Mark]) is a time signature
// written at a bar k ≥ 1; it applies from bar k until the next mark, and bar
// 1 always has one. The pickup has the meter of bar 1. A repeated mark, with
// the meter that already applies, is kept, and a meter change is only a mark
// whose meter differs by record (meter.Meter.Equal) from the previous bar: C
// after 4/4 is a change, 4/4 after 4/4 is not. An explicit length
// ([ExplicitLength]) sets the length of one bar apart from its meter, and a
// length equal to the size of the meter is kept: after a later meter change it
// keeps its length. [Grid.Simplify] removes repeated marks and such lengths.
//
// The length of a bar is the pickup length, its explicit length, or the size
// of its meter. A bar is full ([KindFull]) when its length equals the size of
// its meter, incomplete ([KindIncomplete]) when it is shorter, and extended
// ([KindExtended]) when it is longer. The start of a bar is the sum of the
// lengths of all previous bars; the first bar starts at 0.
//
// # Alignment and the place in the meter
//
// An incomplete bar lies in a full bar of its meter in one of two ways: from
// the start ([AlignStart]), as a bar that begins on the downbeat, or to the
// end ([AlignEnd]), as a pickup, which always ends where a full bar would
// end. The shift of a bar (the field Shift of [Bar]) is the size of the meter
// minus the length for the pickup and for a bar aligned to the end, and 0
// otherwise. The place in the meter of a moment is the shift plus its offset
// in the bar
// ([Grid.MetricPlace]); package meter gives the unit, the beat, the group,
// and the metric weight at that place. In an extended bar there is no place
// in the meter beyond the size of the meter ([ErrUndefined]). A pickup must
// not be longer than the meter of bar 1, and a bar aligned to the end must not
// be longer than its meter ([ErrLongerThanMeter]); every change of a grid
// checks both rules for the result.
//
// # Addresses and positions
//
// An [Address] is a bar number and an offset from the start of the bar. An
// address is valid in a grid when its bar exists and its offset is less than
// the length of the bar, so every position has exactly one valid address: a
// position on a bar line belongs to the next bar. [Grid.Locate] and
// [Grid.PositionOf] convert between positions and valid addresses and are
// inverse to each other; [Grid.PositionOf] rejects an offset not less than the
// length of the bar ([ErrOutsideBar]), [Grid.Contains] reports validity, and
// [Grid.Normalize] carries any offset over into the next bars.
//
// # Unit count, MEI tstamp, and beats
//
// The unit count ([Grid.CountAt]) is a rule of the product for any bar and
// any meter. For a place u in a meter with the units e₁…eₙ starting at
// s₁…sₙ it is i + (u − sᵢ)/eᵢ inside unit i, n + 1 at the size of the meter
// (the right bar line), and n + 1 + (u − size)/eₙ beyond it in an extended
// bar: in 4/4 the eighth after the start is 3/2, and in a bar of 4/4 with the
// length 5/4 its end is 6. In 2/4+3/8 the count goes by the units in order.
// It is counted in two ways: from the start of the bar (u is the offset) and
// by the place in the meter ([Grid.MetricCountAt], u is the shift plus the
// offset), which differ only for the pickup and bars aligned to the end: in
// the pickup 1/4 of 4/4 the first moment counts 1 from the start of the bar
// and 4 by the place in the meter. Counts less than 1 are not moments of a bar.
//
// The MEI tstamp ([Grid.TstampAt]) is 1 + t·d, and only where it is
// established: in a full bar of a meter with a common denominator d. In
// other bars and for meters such as 2/4+3/8 it is [ErrUndefined]; there the
// unit count of the product is used instead. In a full bar of a meter with a
// common denominator the two are equal. Neither is a beat of package meter,
// which is the beat of the English-language tradition: 6/8 has six units and
// two beats. [MeterCountAt] and [MeterTstampAt] count in a full bar of one
// meter without a grid, and [Complement] gives the length of the closing bar
// that completes a pickup.
//
// # Ranges and splitting
//
// A range first…last of bars has a length, a start and an end, meter
// changes inside it (a change at bar first itself is not counted), runs
// without a change ([Grid.MeterRuns]), and full bars. [Grid.BarsSpanned]
// gives the bars that a span of positions occupies; a span ending on a bar
// line does not occupy the next bar.
//
// [Grid.Split] divides a duration at bar lines, and [Grid.SplitBy] also at
// the starts of groups, beats, or units of the meter of every bar ([Level]).
// The parts follow each other, their lengths sum to the duration, and each
// part lies in one bar. Splitting is mechanical: it gives the lengths of the
// parts and does not claim a rule of notation. At most [MaxParts] (1024) parts
// are returned; a longer split is [ErrTooMany] without building the parts.
//
// # Changing a grid
//
// A grid is an immutable value, and every change returns a new grid: the
// pickup, marks, explicit lengths, and the number of bars are set and removed
// one by one; [Grid.WithMetersIn] sets the meters of a range of bars, in a
// cycle for a series of meters, and keeps the meters after the range;
// [Grid.InsertBars] and [Grid.DeleteBars] insert and delete bars and move the
// records of the next bars; [Grid.Slice] cuts a finite grid out of a range of
// bars; and [Grid.Append] joins a finite grid with the bars of another grid,
// turning its pickup into a bar aligned to the end. On an error a change
// returns Grid{} and the receiver stays as it was.
//
// # Three kinds of equality
//
// [Grid.Equal] compares records: the pickup, the marks, the explicit lengths,
// and the number of bars. [Grid.EqualMarkup] compares markups: the pickup,
// the number of bars, and the meter, length, and shift of every bar, so a
// repeated mark does not matter. [Grid.EqualBarLengths] compares the pickup,
// the number of bars, and the length of every bar, so 3/4 and 6/8 are equal.
// Equal records have equal markups, and equal markups have equal bar
// lengths. Grids are not comparable with ==, which does not compile; they have
// no order. The other types of the package are comparable with ==, which is
// the same as their Equal.
//
// # Text
//
// Grid.String, Grid.MarshalText, and Grid.AppendText write the canonical text
// without spaces: the meter of bar 1, the pickup, the marks of bars k ≥ 2 and
// the explicit lengths in the order of bars, and the number of bars, such as
// "4/4", "3/4;pickup:1/4", "4/4;3:3/4;4:length:1/2;bars:8", and
// "4/4;5:length:1/4:end". An address is written "2:1/8". [Parse] and
// [ParseAddress] accept exactly this grammar and the non-canonical forms of
// the nested texts that package meter and package duration accept. Texts of
// grids that differ by record differ, and parsing the text of a set grid or
// address returns an equal value.
//
// The unset grid Grid{} and the unset address Address{} are not values: only
// IsSet, Equal, String, and for an address Compare accept them without an
// error, and their strings "Grid(unset)" and "Address(unset)" are rejected by
// parsing. MarshalText fails for them, so an optional field of a struct takes
// the omitzero tag in JSON.
//
// A parse error quotes the rejected text: in full up to 64 bytes, and
// otherwise its longest prefix of whole UTF-8 characters up to 64 bytes
// followed by "...". The message consists only of the texts of the sentinels
// and this excerpt, is at most 512 bytes long, and for one class of error
// depends on the input only through its first 67 bytes.
//
// # Errors
//
// Every error wraps exactly one of ten sentinels: [ErrUnset],
// [ErrInvalidArgument], [ErrLongerThanMeter], [ErrOutsideGrid],
// [ErrOutsideBar], [ErrUndefined], [ErrGroupingRequired], [ErrTooMany],
// [ErrInvalidText], and [ErrOutOfRange]. When the cause is a failure of a
// dependency, the error also wraps its sentinel: meter.ErrUnset,
// duration.ErrUnset, or fraction.ErrInvalid with ErrInvalidArgument for an
// unset meter, duration or position, or fraction;
// meter.ErrNoCommonDenominator with ErrUndefined; meter.ErrGroupingRequired
// with ErrGroupingRequired; and fraction.ErrOutOfRange with ErrOutOfRange for
// a value outside the range. "No pickup", "an open grid", "no incomplete
// closing bar", "an invalid address", and "not on a bar line" are results,
// not errors. Each operation documents its errors in the order of its
// checks. Every function and method that returns an error, except
// MarshalText, AppendText, and UnmarshalText, has a Must variant that panics
// with the same error.
//
// # Cost, values, and concurrency
//
// The cost of finding a bar by number and a bar by position does not depend
// on the bar number: it depends on the number of marks and explicit lengths,
// so the start of bar 1<<62 is found at once. Only iteration over bars and
// splitting depend on the number of bars they produce. Results are exact.
//
// [Grid.Bars] reports all its failures before iteration, without visiting
// bars, by a sufficient condition over the runs of adjacent bars of equal
// length: the start S and the end F of every run must be in the range of
// common/fraction, and for a run of n ≥ 2 bars of the length L the least
// common multiple D of the denominators of S and L and the integer
// (F − L)·D must not exceed the largest component 2²⁵⁵ − 1; the shifts of
// the bars must be in the range too. After a successful call every bar of
// the range is representable, so iteration never stops early. Near the limit
// of the range the condition can refuse a range whose bars are all
// representable; such bars are read one by one with [Grid.Bar], and for a
// single bar the condition is exact.
//
// Values are immutable: operations return new values and new slices, and
// the slices passed to the package are only read during the call. The
// package has no mutable state, and all its functions, methods, and the
// sequences returned by [Grid.Bars] are safe for concurrent use, except that
// UnmarshalText must not write a value that is read at the same time.
package measure
