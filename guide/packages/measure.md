# Measure grids

A **measure grid** is the sequence of bars of a piece from its start: which time signature each bar has, how long each bar is, and where each bar begins. It knows a pickup, the bars where the signature changes, single bars that are shorter or longer than their signature, and, if wanted, how many bars there are. On top of this sequence the library answers the questions that every editor, player, and analyzer asks: where bar 37 starts, which bar and beat a moment falls on, how a long note is split by the bar lines, and how the grid changes when bars are inserted or deleted. All lengths and positions are exact [durations and positions](duration.md), and the structure of one full bar — its units, groups, beats, and weights — comes from the [time signature](meter.md).

A measure grid holds no notes, rests, or voices, no tempo and no seconds, and no file formats. Whether the notes of a bar fill it, how the tempo turns bars into seconds, and how several parts line up in a score belong to the packages built on top of it. The name of the package means the measure grid, and the time signature is the package named meter; the names do not follow the topics of the same names in theory books.

## Bars and their numbers

Bars follow each other without gaps. The first full bar is bar 1. A **pickup** (anacrusis, upbeat bar) before it is bar 0, as in MusicXML; theory books do not number it at all. The first bar of the grid — the pickup or bar 1 — starts at zero, and each next bar starts where the previous one ends, so the start of a bar is the sum of the lengths of all bars before it.

Besides its number, every bar has an **ordinal place** counted from one with the pickup first: with a pickup, bar 3 is the fourth bar; without one, it is the third.

A grid without a number of bars is **open**: its last signature continues without end, and any later moment falls into some bar. A grid with a number of bars is **finite**: it ends after its last bar, and moments from that end on are outside it. The number of bars counts numbered bars only; a pickup is extra. A finite grid of `4/4` with eight bars and a pickup of a quarter has nine bars in all. In an open grid bar numbers go up to the largest integer of the platform, which no real score reaches.

The cost of finding where a bar starts, or which bar holds a moment, does not depend on the bar number — only on how many signatures and special bars the grid records — so the start of bar 4,611,686,018,427,387,904 in `4/4` is found at once.

## Time signatures and changes

A time signature is written at a bar and applies from that bar until the next one; bar 1 always has one, and the pickup takes the signature of bar 1. In "4/4, with 3/4 from bar 3", bars 1 and 2 are `4/4`, and bar 3 and all later bars are `3/4`.

The grid keeps what is written. Writing the signature that already applies — `4/4` again at bar 5 — is kept as a **repeated** signature, and removing one signature leaves the others alone. A **change** of signature is only a written signature that differs by record from the previous bar: `4/4` to `3/4` is a change, and so is `C` to `4/4`, since the two are different records of the same structure, while `4/4` to `4/4` is not. Checks such as "is there a change inside these bars" — needed for multi-bar rests — count changes, not written signatures. Repeated signatures can be cleaned up explicitly, and the cleaned grid has the same bars.

## Bar lengths and kinds of bars

A bar normally lasts the size of its signature. A pickup lasts its own length, from more than zero up to a whole bar of the first signature: a pickup of a whole bar stays bar 0 and is not turned into bar 1. Any single bar can also be given an **explicit length**, shorter or longer than its signature, without changing the signature. An explicit length equal to the size of the signature is kept as written, so a later change of signature does not change that bar.

Comparing its length with its signature, a bar is:

- **full** when they are equal;
- **incomplete** when the bar is shorter — a pickup, or a short bar inside a piece;
- **extended** when the bar is longer.

The theory sources do not say how an extended bar should be notated; the grid keeps its length and leaves notation alone.

The closing bar of a piece that starts with a pickup usually completes it: after a pickup of a quarter in `4/4` the last bar lasts three quarters. The library computes this length directly from the signature and the pickup; a pickup of a whole bar has no incomplete closing bar.

## Alignment of an incomplete bar

An incomplete bar can sit in a full bar of its signature in two ways. **From the start**, it begins on the downbeat, as a short bar at the start of a phrase. **To the end**, it ends where a full bar would end, as a pickup does: a pickup of a quarter in `3/4` is the third beat of a bar. A pickup is always aligned to the end; other bars are aligned from the start unless told otherwise, and only a bar not longer than its signature can be aligned to the end. Alignment changes where the beats of the bar are, not how long it is or where it starts.

The **shift** of a bar is how far into a full bar it begins: the size of the signature minus the length for the pickup and for a bar aligned to the end, and zero otherwise. The pickup of a quarter in `3/4` has the shift of a half.

A grid enforces two rules at every change: a pickup is never longer than the first signature, and a bar aligned to the end is never longer than its signature. A change that would break either rule — say, replacing the first signature `4/4` by `2/4` under a pickup of three quarters — is refused, and the grid stays as it was. To make both the signature and the pickup smaller, make the pickup smaller first.

## Addresses and positions

A moment can be given as a **position**, its distance from the start of the grid, or as an **address**, a bar and an offset from the start of that bar. In `3/4` with a pickup of a quarter, the position `9/8` is bar 2 at `1/8`, and bar 1 at `1/2` is the position `3/4`.

Every moment has exactly one valid address: its bar exists and its offset is less than the length of the bar. A moment exactly on a bar line belongs to the next bar, so the position `1` in `4/4` is bar 2 at zero, not bar 1 at a whole note. An address such as "bar 2 at five quarters" is not valid in `4/4`, but it can be **normalized**: the extra length carries over into the next bars, giving bar 3 at a quarter. Whether an address is valid is a yes-or-no answer, not an error.

For any moment the grid also tells the bar that holds it, how much of the bar is left until the next bar line, and whether the moment is on a bar line — the start of a bar or the end of a finite grid.

## Ranges of bars

For a range of bars the grid gives its total length, where it starts and ends, whether every bar in it is full, whether the signature changes inside it, and the runs of bars without a change: in "4/4, with 3/4 from bar 3", bars 1 to 5 are bars 1–2 in `4/4` and bars 3–5 in `3/4`. A change at the first bar of the range itself is not counted as a change inside it. These answers serve whole-bar and multi-bar rests, which are written in their own packages.

For a span of time from a given moment, the grid tells which bars it occupies. A span that ends exactly on a bar line does not reach into the next bar: a whole note from the start of `4/4` occupies bar 1 only. The grid also tells what part of a bar a moment has passed and what part of a bar a length makes up: a quarter is a third of a bar of `3/4`, and a whole note four thirds.

## The place in the meter

For any moment the grid gives the signature of its bar and the **place in the meter**: the distance from the start of a full bar of that signature, which is the shift of the bar plus the offset. In the pickup of a quarter in `3/4`, its first moment is at a half — the third beat, a weak one — and the first moment of bar 1 is at zero, the strong beat. The [time signature](meter.md) then gives the unit, the beat, the group, and the weight at that place, so a metronome can follow the grid bar by bar. In an extended bar there is no place in the meter beyond the size of the signature.

## Counting units and the MEI tstamp

Musicians count bars by units: "one-and-two-and". The library gives this **unit count** as an exact number: the whole part is the number of the unit, and the fraction is how much of the unit has passed. In `4/4` the eighth after the start is 1½, and the right bar line is 5. In a signature whose parts have different units, such as `2/4+3/8`, the count goes through the units in order: two quarters, then three eighths. In an extended bar the count goes on past the bar line of the signature in steps of the last unit: in a bar of `4/4` that lasts five quarters, its end is 6. Counts below one are not moments of a bar.

The count can be taken in two ways, which differ only in a pickup and in bars aligned to the end. **From the start of the bar**, the first moment counts one. **By the place in the meter**, the bar counts as the end of a full bar: the pickup of a quarter in `4/4` starts at 4, and the pickup of a quarter in `3/4` at 3. Both directions are available — from a moment to its count and from a count back to its moment.

The **MEI tstamp** used by the MEI encoding counts the same way, but only where it is established: in a full bar of a signature with one denominator. There the tstamp equals the unit count. In a pickup, a short or extended bar, or a signature such as `2/4+3/8` the tstamp is undefined, and the library refuses it rather than guess how a file would count; the unit count is the library's own rule for those cases. Neither is the "beat" of the time signature: `6/8` has six units and two beats. Both the count and the tstamp are also available for a full bar of a single signature without a grid.

## Splitting a duration

A long note or rest that crosses bar lines is split into parts, one per bar: a duration of two whole notes and a quarter from the start of `4/4` is a whole note in bar 1, a whole note in bar 2, and a quarter in bar 3. On request the split also stops at the starts of the groups, beats, or units of each bar: three eighths from the fourth eighth of `4/4`, split by groups, are an eighth up to the half of the bar and a quarter after it. In a pickup and in bars aligned to the end the groups and units are those of the full bar the bar belongs to. In an extended bar the bar line of the signature is a boundary too, and splitting by units continues past it in steps of the last unit.

Every part has its bar and offset, its position, and its length; the parts follow each other without gaps, add up to the whole duration, and each lies in one bar. Splitting by groups or beats needs a grouping in every bar it crosses, so `5/8` without a grouping is refused. The split is mechanical: it gives lengths and does not claim that notation must follow it, and writing the parts as note values with ties is done separately. At most 1024 parts are returned — the same limit as the number of tied notes of a duration; a longer split is refused without being built.

## Changing a grid

A grid is an immutable value: each change gives a new grid. Besides setting and removing the pickup, a signature, an explicit length, and the number of bars, a grid can be changed by ranges of bars:

- **signatures over a range**: bars 1 to 4 of `4/4` become `3/4`, and bar 5 goes back to `4/4`. A series of signatures repeats in a cycle — `6/8`, `3/4`, `6/8`, `3/4` — for at most 1024 bars, and the signature after the range is restored. A single irregular bar is a signature over a range of one bar, not an explicit length;
- **inserting bars**: new full bars take the signature of the bar before them, and the records of later bars move with their bars;
- **deleting bars**: the records of deleted bars go, later bars move up, and a signature set inside the deleted bars still applies to the bar that follows them. All numbered bars of a finite grid cannot be deleted;
- **adding bars** at the end of a finite grid, and **cutting** a grid to a number of bars, which drops the records after the cut;
- **slicing** a range of bars into a finite grid of its own, every bar keeping its signature, length, and alignment;
- **joining** a finite grid with another grid: its pickup becomes an ordinary bar aligned to the end, so its beats stay where they were. An open grid has no end to join to.

Notes written in the bars are not checked by the grid: whether they fit the new bars is the business of the part that holds them.

## Three kinds of equality

Two grids can be compared in three ways:

- **by record**, the main equality: the same pickup, the same signatures at the same bars, the same explicit lengths, and the same number of bars. A repeated signature makes two grids different;
- **by markup**: every bar has the same signature, length, and shift, so a repeated signature or an explicit length equal to the signature does not matter, while a bar aligned differently does;
- **by bar lengths**: every bar has the same length, whatever its signature, so `3/4` and `6/8` are equal. This is the question two parts of a score ask when their bar lines must coincide.

Equal records have equal markups, and equal markups have equal bar lengths. An open grid and a finite one are different in every way, and grids have no order for sorting.

## Text

Every grid has a short canonical text without spaces: the signature of bar 1, the pickup, the signatures from later bars and the explicit lengths in the order of bars, and the number of bars — `4/4`, `3/4;pickup:1/4`, `4/4;3:3/4;4:length:1/2;bars:8`, `4/4;5:length:1/4:end`. An address is written as its bar and offset, `2:1/8`. Reading accepts exactly this grammar and order, and the signatures and lengths inside it as the time signature and the duration read them. Texts of different records differ, and reading the text of any grid gives the same grid back. MusicXML, MEI, and MIDI files are not read; exchange with those formats happens outside the library.

When a long text is rejected, the error quotes at most its first 64 bytes, without cutting a character in half, so error messages stay short however long the input is.

## The unset value

An empty grid or address means that nothing is set; it is not a grid of no bars. It can be compared, an empty address sorts before every set one, and both print as recognizable placeholders that reading refuses, but every calculation with them reports an error. An optional grid in a structured document is therefore left out rather than written as a placeholder.

Grids and addresses are immutable values that are safe to copy and to use from many goroutines at once.

First released in v0.29.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/measure) · [Go examples](../../measure/example_test.go) · [Roadmap](../../ROADMAP.md)
