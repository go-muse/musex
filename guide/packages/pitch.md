# Written pitches

A written pitch combines a [note spelling](spelling.md) with a positional [octave](octave.md) on the register axis. F♯ in the octave containing middle C pairs the spelling F♯ with the octave whose ASPN number is 4. That octave retains its identity whether it was found from ASPN 4, from 3 in Franco-Belgian numbering, or from *первая* in the Russian catalog. Both the spelling and the octave are required; a note without an octave belongs to the note model.

## Positional octave and identity on the page

The octave of a pitch identifies the positional octave section in a musical system rather than the register digit printed next to a letter in score notation. In the traditional system, where each octave section runs from C through B, notated B♯3 and C4 share the same positional octave ASPN 4: their written pitches are (B♯, ASPN 4) and (C, ASPN 4), with written register digits 3 and 4. Likewise, notated C♭4 and B3 share positional octave ASPN 3 as (C♭, ASPN 3) and (B, ASPN 3). Translating between a positional octave and a written register digit is the role of the musical system and notation packages.

Two pitches are equal when both their spelling and their positional octave agree. Enharmonic writings remain distinct: C♯ and D♭ in ASPN 4 are different pitches, as are B♯ and C in ASPN 4. Raising or lowering an accidental, or returning a pitch to its natural form, changes only the alteration while keeping the letter and the positional octave (raising B in ASPN 4 gives B♯ in ASPN 4, and stripping the sharp returns B in ASPN 4).

## Register movement, order, and debug form

Shifting a pitch by octaves, stepping to the next octave up or down, and iterating across a range of octaves preserve the spelling while moving along the octave axis. Octave difference and octave distance compare only the octaves of two pitches—any two pitches in the same octave are zero octaves apart—whereas musical intervals between pitches belong to the interval package. For deterministic sorting, pitches are ordered first by octave and then by spelling (C through B, then alteration value); this conventional order of written records does not represent sounding pitch height or system position.

For logs, debugging, and tests, a pitch formats its spelling and ASPN octave as `F+1 ASPN 4`. This debug form is not score notation and is not parsed back. An uninitialized pitch is invalid (`Pitch(invalid)`), never a default such as C4 or A4.

First released in v0.10.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/pitch) · [Go examples](../../pitch/example_test.go) · [Roadmap](../../ROADMAP.md)
