# Ticks

Exchange formats store musical time as whole numbers. A Standard MIDI File counts **ticks**, MusicXML counts **divisions**, and MEI counts pulses per quarter. The library keeps time exactly, as fractions of a whole note, so every exchange needs a translation between the two. The ticks package performs that translation in both directions: for durations and positions, for measure addresses, for tempo and tempo changes, for time signatures and their changes, for clock time in seconds and SMPTE frames, and for MIDI clock pulses. It does not read or write files; it only produces and interprets the numbers that such files contain.

## Resolution and the MIDI quarter

A **resolution** says how many ticks make up one quarter note. MusicXML and MEI always count written quarters, so at a resolution of 480 a quarter note is 480 ticks, an eighth is 240, a triplet eighth is 160, and a dotted quarter is 720. A Standard MIDI File counts **MIDI quarters** instead. Its time signature event records how many notated thirty-second notes make up one MIDI quarter: usually eight, which makes the MIDI quarter an ordinary quarter note, but sixteen makes it a half note, and then a half note is 480 ticks at the same resolution. The library keeps this notation together with the resolution and with every tempo calculation; by default the MIDI quarter is an ordinary quarter.

A conversion is exact. A value that is not a whole number of ticks is an error, never silently rounded: a triplet eighth needs a resolution divisible by three. When a whole number is not required, the exact number of units is available as a fraction. When the caller wants a whole number anyway, quantization rounds once with an explicitly chosen rounding mode and reports exactly how far the result moved. Ticks can also be rescaled from one resolution to another, exactly or with explicit rounding.

The library can choose a resolution for a piece: the smallest one that represents every duration and position exactly, the smallest multiple of it at or above a preferred value such as 480, and a check against an upper limit. It can also list the values that a given resolution cannot represent. A resolution itself has no upper limit; separate checks report whether it fits the limit of a MIDI file header (32767) and whether it stays within the value that MusicXML recommends for compatibility with MIDI (16383).

A MIDI file stores events with **delta-times**: the ticks since the previous event. The library converts between absolute ticks and delta-times, checks their order, and checks the largest delta-time a file can store. It also gives the ticks of a MIDI clock pulse (24 per MIDI quarter) and of a MIDI beat (a sixteenth), and converts positions to and from these pulses.

## SMPTE time

A MIDI file can also measure time in frames of a film or video standard instead of quarters. Such a resolution consists of a frame rate and a number of ticks per frame, and its ticks are fractions of a second that do not depend on the tempo. The library encodes and decodes the header value of both kinds of resolution. One of the four frame rate codes is read differently by different sources: as 29.97 frames per second, as 30, or as 29. The caller therefore chooses the reading before any seconds are calculated; the library never guesses. Time code labels in hours, minutes, seconds, and frames are outside the package.

## Tempo

MIDI stores tempo as microseconds per MIDI quarter, within a fixed range. The library converts a [tempo marking](../tempo.md) to this value exactly, and rounds it to a whole number only with an explicitly chosen mode; a tempo too slow or too fast for the format is an error. In the other direction, a MIDI tempo becomes an exact marking in a note unit chosen by the caller, so the same speed can be written as a quarter note or as a dotted quarter. MusicXML and MEI state tempo in quarter notes per minute; the library converts these numbers too, formats them as decimals with a chosen number of places, and reads them back.

A **MIDI tempo map** is a list of tempo changes by tick. Before the first change the standard MIDI tempo of 120 quarters per minute applies unless the caller sets another one. Several changes may share one tick; the last of them takes effect. The map can be edited by tick and by place on a tick, gives the exact time in seconds of any tick or position and the tick at any time, converts to and from SMPTE ticks, and can be rescaled. Two maps can be compared as records or by the exact seconds at every musical position. Changing only the resolution and rescaling the ticks preserves timeline equality; moving a tempo change to another musical position can change that timeline.

A MIDI tempo map becomes a library [tempo map](../tempomap.md) of new main tempos at the ticks where the tempo changes, by position or by bar. The opposite direction is richer, because a library tempo map knows more than MIDI can store:

- **Gradual changes** become a series of steps. The caller chooses the step length and how each step gets its tempo: the speed at its start, the speed at its end, or a tempo that makes the step last exactly as long as that part of the gradual change. Steps can be required to fall on MIDI clock pulses. Without these choices a gradual change is an error.
- **Fermata holds** are refused, skipped and reported, or stretched: a chosen length of music before the hold is slowed down so that it absorbs the extra time, and the original tempo resumes at the hold. Splitting a step at a hold boundary preserves that step's base tempo. The extra time is calculated from the complete exported staircase after positions have been converted to ticks, so all three step methods add exactly the requested seconds before final tempo rounding. If position rounding collapses the whole stretched interval to one tick, export fails explicitly.
- **Positions** that are not whole ticks are an error unless the caller allows rounding.

## Correct rounding

Every whole number the package produces is rounded only once, directly from the exact value, by the mode the caller chose. This holds even inside a gradual tempo change, where the exact time is not a fraction at all: the package shares the tempo map's exact time arithmetic and can round seconds and positions to any step, such as a tick, a frame, or a microsecond, with a guaranteed result. Calculated internal boundaries may use larger fractions than public position values; only the final public results have those limits. One rare case has no such guarantee: an equal-time step inside a gradual change that is also stretched for a fermata. When both the step time and the total time being stretched are irrational, the package uses certified intervals at four increasing precisions. It refuses with a distinct error if those intervals cannot prove the rounded result; it never substitutes an unproven number. A rational fragment stretched by an irrational total still uses exact comparisons and has no such precision refusal.

## Time signatures

A MIDI time signature event holds four numbers: the numerator, the denominator as a power of two, the MIDI clock pulses between metronome clicks, and the notation of the MIDI quarter. The library converts a [time signature](../meter.md) into such an event. The metronome click is either an explicit duration or the beat of the time signature; a time signature without equal beats needs an explicit click. A denominator that is not a power of two cannot be stored. Symbols such as common time, additive numerators, sums of time signatures, and explicit grouping cannot be stored either: in strict mode they are errors, while in lossy mode they are written by their total size and the loss is reported. An event converts back into a simple time signature and a click.

A [measure grid](../measure.md) becomes a list of time signature events by tick, and such a list becomes an open grid again. A pickup or a bar of unusual length can get its own event of its actual length, or none at all. An event that falls inside a bar ends that bar early, and the next bar starts with the new time signature. Measure addresses, bar starts, and bar lengths convert to ticks and back through the grid.

## Text, errors, and concurrency

Every value has canonical text with round-trip parsing; diagnostic text of unset values is rejected by the parsers. Unset values are not defaults: they support only presence checks, equality, and diagnostic text. Errors distinguish unset values, invalid arguments, zero values, values that are not whole ticks, values outside a format limit, wrong order, maps that cannot be converted as requested, values out of range, malformed text, and time signatures that MIDI cannot store, while keeping the causes reported by the underlying packages.

All values are immutable and safe for concurrent use. Returned lists belong to the caller. The only mutating operation is text unmarshaling into a receiver, which follows the ordinary Go rule against concurrent writes and reads of that same value.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/format/ticks) · [Go examples](../../../format/ticks/example_test.go) · [Roadmap](../../../ROADMAP.md)
