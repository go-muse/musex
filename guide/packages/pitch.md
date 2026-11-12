# Written pitches

A written pitch combines a [note spelling](spelling.md) with a positional [octave](octave.md) on the register axis. F♯ in the octave containing middle C pairs the spelling F♯ with the octave whose ASPN number is 4. That octave retains its identity whether it was found from ASPN 4, from 3 in Franco-Belgian numbering, or from *первая* in the Russian catalog. Both the spelling and the octave are required; a note without an octave belongs to the note model.

## Positional octave and identity on the page

The octave of a pitch identifies the positional octave section in a musical system rather than the register digit printed next to a letter in score notation. In the traditional system, where each octave section runs from C through B, notated B♯3 and C4 share the same positional octave ASPN 4: their written pitches are (B♯, ASPN 4) and (C, ASPN 4), with written register digits 3 and 4. Likewise, notated C♭4 and B3 share positional octave ASPN 3 as (C♭, ASPN 3) and (B, ASPN 3). Translating between a positional octave and a written register digit is the role of the musical system and notation packages.

Two pitches are equal when both their spelling and their positional octave agree. Enharmonic writings remain distinct: C♯ and D♭ in ASPN 4 are different pitches, as are B♯ and C in ASPN 4. Raising or lowering an accidental, or returning a pitch to its natural form, changes only the alteration while keeping the letter and the positional octave (raising B in ASPN 4 gives B♯ in ASPN 4, and stripping the sharp returns B in ASPN 4).

## Changing names: positional octave versus written register

`WithSpelling`, `Natural`, `Raise`, `Lower`, `RaiseBy`, `LowerBy`, and the `Must*` variants of the operations that return errors preserve the positional octave, not the written register. They do not guarantee preservation of the system position or sounding pitch. If the new spelling crosses a positional-octave boundary, the written register and the system position (and therefore the sounding pitch at a fixed tuning) can change even though the positional octave stays the same.

The following examples use the traditional system. The record forms below are score notation, not `Pitch.String` debug values; positions are positions in the musical system.

| Operation | Input pitch, record, and position | Resulting pitch | Resulting record and position |
|---|---|---|---|
| `Raise` | (B, Octave 3), record B3, position B3 | (B♯, Octave 3) | record B♯2, position C3 |
| `Lower` | (C, Octave 4), record C4, position C4 | (C♭, Octave 4) | record C♭5, position B4 |
| `Natural` | (B♯, Octave 4), record B♯3, position C4 | (B, Octave 4) | record B4, position B4 |

In the first two examples, the written register changes opposite to the intuitive expectation from raising or lowering. `Natural` does not turn B♯ into C: it produces B with record B4.

To preserve the written register instead, explicitly choose a [musical system](system.md#octave-and-written-octave) and call `System.RaiseWritten`, `System.LowerWritten`, `System.RaiseWrittenBy`, `System.LowerWrittenBy`, `System.RaiseWrittenByAlteration`, `System.LowerWrittenByAlteration`, `System.RespellWritten`, or `System.NaturalWritten`, or the corresponding `Must*` method such as `System.MustRaiseWritten`. These operations preserve the written register (the written octave in `system` terminology), while the positional octave of the result can change. The `pitch` package neither calls `system` nor selects a musical system for the user.

## Register movement, order, and debug form

Shifting a pitch by octaves, stepping to the next octave up or down, and iterating across a range of octaves preserve the spelling while moving along the octave axis. Octave difference and octave distance compare only the octaves of two pitches—any two pitches in the same octave are zero octaves apart—whereas musical intervals between pitches belong to the interval package. For deterministic sorting, pitches are ordered first by octave and then by spelling (C through B, then alteration value); this conventional order of written records does not represent sounding pitch height or system position.

For logs, debugging, and tests, a pitch formats its spelling and ASPN octave as `F+1 ASPN 4`. This debug form is not score notation and is not parsed back. An uninitialized pitch is invalid (`Pitch(invalid)`), never a default such as C4 or A4.

First released in v0.10.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/pitch) · [Go examples](../../pitch/example_test.go) · [Roadmap](../../ROADMAP.md)
