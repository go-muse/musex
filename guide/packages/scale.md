# Scale

> Draft API for review. The `scale` package is planned for
> `github.com/go-muse/musex/scale`.

A scale is a finite, ordered list of [notes](note.md). Build one directly
with `scale.New` or derive its notes from a [mode](mode.md) and a reference
note. A scale keeps each note's spelling, optional octave, cents correction,
and settings source. It also keeps the order and any repeated notes. Once
built, the scale holds only the notes, so it can be edited independently of
the mode.

## Build notes from a mode

`FromMode` gives one note per degree, in mode order. For a seven-degree
octave-based mode, the result has seven notes. To include the first degree
in the next cycle, request a longer result with one of the expansion
functions.

| Function | What it builds |
| --- | --- |
| `FromModeCount` | Exactly the requested number of notes, including the selected start. A closed mode continues through its cycle intervals. An open mode reports `ErrInsufficientNotes` if it runs out of notes. |
| `FromModeTo` | The notes from the selected start up to an inclusive pitch boundary in the chosen direction. The endpoint appears only if a generated note reaches its position. |
| `FromModeOctaves` | The same bounded expansion, with the end set a positive number of octaves above or below the selected start. A cycle of the mode can span a different interval. |

Set `BuildOptions.Direction` to `Forward` or `Backward`. `BuildOptions.Start`
defaults to the first degree of cycle zero. Its `Index` is zero-based,
and its `Cycle` can be negative for a closed mode. The reference note
always anchors the first degree of cycle zero. A different start is
computed from that anchor.

Pitch and octave bounds require an octave on the reference note. The `In`
variants take an explicit [musical system](system.md). With a pitch or
octave bound, a closed mode must advance strictly in the chosen direction
throughout its repeating structure. An open mode contributes every
in-range note from its remaining degrees in that direction.
`StripOctaves` removes octave information from the finished list and
preserves any repeated names.

## Walk the stored notes

Indexes start at zero. `At` reads an existing index. `Next`, `Previous`,
and `Move` wrap around the list and return an `Element` with the result's
index and note. In `[G4, A4, B4]`, `Next(2)` returns index zero and `G4`.
Wrapping preserves the complete stored note, including its octave.

`Take` starts with the selected note and returns the requested number of
notes. `OneRound` returns one full traversal. `Iter` and `Round` give
lazy sequences with the same respective limits. `Cycle` repeats a
nonempty list until the caller stops, for example with `break`. Each
sequence captures the notes when the method is called. If you create a
cycle at index one in `[G4, A4, B4]` and then replace the scale's notes,
the cycle still starts `A4, B4, G4, A4`. Starting another traversal of
that sequence starts over from `A4`.

## Compare notes and select occurrences

By default, search matches spelling, octave presence and value, and cents
correction. Use `ByRecord` to ignore cents or `ByName` to compare spellings
alone. `ByPosition` uses the notes' compatible systems.
`ByPositionIn` uses one supplied system and compares positions when all
notes have octaves, or classes when all notes lack them. `BySoundIn` uses
one supplied tuning and compares sound exactly, with octaves required.
Contextual comparisons validate all participating notes.

`Find` returns every matching index in ascending order. When selecting a
starting note for `MoveFrom`, pass its absolute list index in `Selection`
if it appears more than once under the chosen comparison. An unspecified
index with multiple matches results in `ErrAmbiguous`. Melody shifts and
reflections use the same selection rules and return the complete notes at
the resulting scale indexes.

`Equal` checks notes in order. `SameSet` ignores order and repeated
occurrences, while `SameMultiset` counts occurrences. `RotationStarts`
reports every start index whose rotation matches the other scale. Set
operations use the chosen strict comparison and keep the first matching
representative. To look for sound within a tolerance, use `FindSound` or
`FindSoundInTuning` with a frequency criterion.

## Edit and transform a scale

The zero value of `Scale` is an empty scale ready for use. `Replace`,
`Set`, `Insert`, `Append`, `Remove`, `Delete`, and `Clear` edit the scale
in place. `Delete` uses a half-open range, and `Insert` accepts the index
just past the last note. A failed edit leaves the scale unchanged.

`Notes` returns a copy of the list. `Clone` and copying the `Scale` value
give independently editable scales. Copying a pointer gives another
reference to the same scale. Note settings sources keep their original
owners. Synchronize edits with other access to the same scale.

`Rotate`, `Reverse`, `Slice`, and `Repeat` build new lists from the stored
notes. Use `Transpose` to move every note by a written interval,
`ShiftOctaves` for an octave shift, and `WithEnharmonic` to change one
note's spelling while preserving its position or class. A scale can also
report its [pitch-class set](pcset.md) in a supplied system and its
standard key signature when its spellings determine one.

[Roadmap](../../ROADMAP.md)
