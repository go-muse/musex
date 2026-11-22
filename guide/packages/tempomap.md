# Tempo map

A **tempo map** puts [tempo markings](tempo.md) and fermata holds on a musical timeline. It answers not only how fast a constant tempo is, but which tempo applies at each moment, how long the music has lasted in seconds, and which position or measure corresponds to a given number of seconds. Maps are immutable: every edit produces a new map, and rejected edits leave the input untouched.

A map uses one addressing scheme throughout: positions measured in whole notes from the beginning, or a bar number with an offset inside that bar. A measure-addressed map does not own a [measure grid](measure.md); the caller supplies the grid whenever positions, seconds, or address validation are needed. The same bar addresses therefore retain their meaning if the grid later changes. Conversion between the two addressing schemes is explicit and preserves every change and hold.

## Main and current tempo

Every map begins with a **main tempo**. The **current tempo** is the marking in force at a particular moment. These can differ:

- a new main marking replaces both;
- a jump replaces only the current marking and starts a deviation;
- **a tempo** immediately returns to the last main marking and ends the deviation, so repeating it without another deviation is an error;
- **Tempo I** immediately restores the initial marking and makes it main again, affecting later returns as well;
- a metric modulation derives a new main marking from the current marking and an equation between a new unit and an old duration.

Changes act at their own addresses. Seconds already accumulated remain continuous when the speed jumps. Repeating an equal marking is allowed and changes the written record even though it does not change the timeline. Record equality and timeline equality are therefore separate comparisons.

## Gradual changes

A gradual deviation changes speed from its current value to the exact tempo given by the next change. The product supports one explicit law: speed in whole notes per minute is linear by musical position. This is a chosen computational law, not a claim that every accelerando or ritardando follows the same mathematical curve. The written note unit is not interpolated; an equivalent intermediate marking is produced only in a unit selected by the caller.

The next change must provide a computable endpoint: a main marking, a jump, a tempo, or Tempo I. Another gradual change or an immediate metric modulation cannot provide it. A map may temporarily end with an unfinished gradual change for later editing, but speed and time calculations on that tail or after it are refused; calculations strictly before its start remain available.

An initial pair can place the main marking and a gradual deviation at the same starting address. This allows a gradual segment to begin at zero while preserving the required initial marking. The pair has a first and a second slot for independent replacement and removal, and it creates no zero-length constant segment. No other address permits duplicate changes.

Speed inside a gradual segment is always an exact rational value. Seconds, however, involve a logarithm, and the reverse conversion involves an exponential. These results are approximate only when a gradual segment with unequal endpoint speeds participates; equal endpoints use the exact constant-speed formula.

## Fermata holds

A fermata hold is an explicit non-negative number of seconds at one address. It does not change the tempo, the position, or the measure grid. Zero holds are retained in the record although they do not affect time.

Each moment has two times:

- **start of moment** includes all holds strictly before the position, but not its own hold;
- **end of hold** adds the hold at the position itself.

Time inside a hold converts back to that same position. An interval between positions includes the hold at its first position and excludes one at its last, so adjacent intervals add correctly. For a sounding note, use the end of hold at both boundaries: the next note starts only after the hold of the previous one has ended.

## Approximation and guarantees

The caller always chooses the number of decimal places and one of the nine [rounding modes](common/rounding.md); there is no default. All participating segments, exact holds, and rational terms are combined before one final rounding. Intermediate segments are never rounded separately.

An approximate result carries its value, decimal precision, and a guaranteed error bound: one unit in the last decimal place for directed modes and half a unit for nearest modes. Exact results carry zero bound and do not change with the requested precision. Stored addresses, change records, and equivalent intermediate speeds remain exact even when seconds are approximate.

Calculations use exact rational arithmetic and directed arbitrary-precision intervals, with a precomputed limit on refinement. They do not use machine floating-point values, unbounded iteration, NaN, or infinity. Values and error bounds must fit the library's exact fraction range. Both bounds remain representable at 76 decimal places; at 77 an approximate operation fails because its required bound is outside that range, while an exact result still succeeds.

## Rounding to a step

Seconds and positions can also be rounded to a multiple of any positive step instead of to decimal places: a MIDI tick, a video frame, a millisecond, or a microsecond. The result is the number of whole steps, chosen once from the exact value by the caller's rounding mode. The guarantee is stronger than for decimal places: the choice is always correct, even when the exact value is not a fraction or lies exactly halfway between two steps, and there is no precision limit at which the calculation gives up. This is what allows exact export of tempo maps to MIDI ticks and microseconds.

The library can also round a whole weighted sum of time intervals at once. For example, assigning one third of a one-second interval and two thirds of another one-second interval gives exactly one second before the final rounding. Intervals may overlap or repeat. Their positive weights and the chosen step have unrestricted rational precision; only the resulting whole-step count must fit the machine integer range. The call leaves the input list and numbers unchanged. The caller must keep them unchanged while a concurrent calculation reads them.

## Text, unset values, and concurrency

Changes, holds, maps, laws, addresses, and both result types have canonical text with round-trip parsing. Exact and approximate results use different forms, so an approximation cannot silently be read as exact. Diagnostic placeholder strings for unset values are rejected by parsers.

Unset zero values are not empty maps or zero results. They support only presence checks, equality, ordering where defined, and diagnostic text; calculations reject them with a distinct error. Other failures distinguish invalid arguments, missing initial tempo, duplicate or invalid addresses, missing deviation, invalid sequences, unfinished gradual tails, negative time, out-of-range values, and malformed text, while retaining relevant causes from dependencies.

All values are immutable and safe for concurrent reading. Returned collections belong to the caller, and modifying them cannot change a map. The only mutating operation is text unmarshaling into a receiver, which follows the ordinary Go rule against concurrent writes and reads of that same value.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/tempomap) · [Go examples](../../tempomap/example_test.go) · [Roadmap](../../ROADMAP.md)
