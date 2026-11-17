# Event

An event is one written element of a voice. It combines a [note](note.md), a nonempty [chord](chord.md), or an explicit rest with its duration and marks. A chord of one note keeps its chord form. Chords preserve the written order of their note occurrences, including repeats.

## Duration and occupied length

- A **written duration** keeps the note value, dots, and ordered tuplets. Two notations with the same value can represent different records.
- An **exact duration** stores a positive fraction of a whole note without choosing a written value.
- A **duration in measures** gives the number of measures occupied by a whole-measure or multimeasure rest. The caller uses the rest's position and the measure grid to turn that count into an exact length.

An ordinary note, chord, or rest occupies its duration in whole notes when that value is available. A grace note or chord keeps its written duration and occupies zero length. A rest in measures gives `ErrNeedsMeasureContext` when asked for its length.

## Marks and changes

A tie marks a particular zero-based note occurrence for continuation in the next event. Repeating its index gives one tie. Two identical notes at different indexes can be tied independently. A fermata can mark any event, and an arpeggio can mark any chord. The caller determines how the signs affect performance and matches ties to the next event.

Events are immutable. The `With` methods create new events and validate all their parts together. `WithContent` preserves each tie only when the written note stays at its original index. Use `WithContentAndMarks` to replace the content and specify new ties together. Notes, chords, and tie lists returned for inspection can be edited independently. Note copies retain their settings sources.

Record equality includes the content form, notes in order, duration form and record, and marks. Settings sources do not affect note equality. A separate length comparison uses only the exact occupied lengths.

Included in v0.32.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/event) · [Usage](../../README.md#build-a-voice-event) · [Roadmap](../../ROADMAP.md)
