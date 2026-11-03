# Note spellings

A spelling combines a [letter](letter.md) and an [alteration](alteration.md). F♯ has letter F and alteration +1. It can occur in any octave. The seven letters and five traditional accidentals give 35 ready-made spellings, and any supported alteration, including a fraction, can accompany any letter.

## Identity on the page

Two spellings are equal when both parts agree. C♯ and D♭ remain distinct spellings even in a musical system that assigns them the same position. Raising C♯ by one chromatic unit gives C double sharp. The operation keeps the letter C, preserving its role in the written music.

For sorting, spellings follow letter order from C through B, then alteration order within each letter. Letter distance and the difference in alterations describe two separate aspects of the writing. Neither comparison alone determines a sounding interval.

## Stored form

Every spelling has one machine-readable form. A natural uses the letter alone (`F`), and an altered name adds a signed value (`F+1` or `B-7/2`). These forms also serve as JSON strings and map keys. The parser expects that exact form. Display conventions such as F♯ and Fis belong to musical notation. The zero value is an unset spelling, written `unset`.

First released in v0.5.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/spelling) · [Go examples](../../spelling/example_test.go) · [Roadmap](../../ROADMAP.md)
