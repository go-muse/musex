# MIDI note numbers

The `midi` package connects the written musical model to the ordinary MIDI note-number map. A MIDI number is an integer from 0 through 127, not a frequency, a pitch class, a tuning step, or a stored property of a note. MIDI files and messages, channels, velocity, controllers, time, and duration are outside the package.

## Numbers, octaves, and positions

The map starts at C in positional octave −1 and ends at G in positional octave 9:

- C−1 = 0, C0 = 12, A0 = 21;
- C4 = 60, A4 = 69, C5 = 72;
- C8 = 108, G9 = 127.

For number `m`, `Parts` returns positional octave `m/12−1` in ASPN numbering, class `m mod 12`, and numeric C-layout position `P = m−12`. `FromParts` reverses this numeric assembly. It needs no musical system, but its class must have modulus 12 and its result must remain in the MIDI range.

`Min`, `Max`, `Count`, and `All` describe the complete range. `NumbersInOctave` enumerates the valid numbers of one positional octave in ascending order: a full octave gives twelve numbers, the top octave ASPN 9 gives 120 through 127, and a valid octave outside the MIDI range gives an empty sequence; the invalid octave is an error, not an empty result. `Shift` moves by positions, `ShiftOctaves` moves by twelve positions while preserving the class, and `Next`, `Previous`, `Up`, and `Down` are their neighboring forms. `Sub` gives a signed position difference. Results outside 0 through 127 are errors: there is no wrapping, clamping, saturation, or transition through an endpoint, including for the smallest and largest Go `int` arguments.

The zero `Number` is invalid and does not mean MIDI 0. Its `IsValid` is false, operations that can fail report `ErrInvalidNumber`, and `Equal` is false even against itself. `Compare` still defines a total sorting order: invalid values precede valid ones and compare equal to one another. `String` distinguishes `Number(invalid)` from every valid number.

## Traditional C layout is required

Converting a written pitch or a note uses the system from an explicit argument or settings snapshot. That system must have all four properties of the traditional C layout:

1. twelve position classes;
2. offsets C0 D2 E4 F5 G7 A9 B11;
3. chromatic unit 1;
4. an octave boundary immediately before C.

The system name and its set of admissible alterations may differ. A restricted custom system works for the names it admits, and refuses a missing spelling rather than escaping that set. Twelve classes alone are not enough, and the traditional offsets with another boundary letter are not accepted.

The octave in this calculation is the **positional octave**, not the register digit written beside the letter. B♯3 is B♯ in positional octave 4, so it maps to 60 just like C4. C♭4 is C♭ in positional octave 3, so it maps to 59 just like B3. These enharmonic records remain different pitches even though their MIDI numbers coincide.

`FromPitch` performs this conversion for a written pitch. `FromNote` reads one snapshot from the note's own source; `FromNoteIn` reads one snapshot from an explicit source instead, without reading or changing the note's own binding. Temperament, reference frequency, notation, and the note's correction do not participate in the number.

## Spelling must be chosen explicitly

A MIDI number identifies a position but cannot recover its original spelling. `Number.Pitch` therefore requires an explicit rule: `system.Sharps`, `system.Flats`, or `system.CustomSpelling`. No neutral or modal rule is selected silently.

`system.ChromaticScale` and any unknown rule kind are rejected with `midi.ErrNotApplicable` before `system.PitchAt` is called. A missing rule is `ErrNoSpellingRule`; a rule with no admissible name preserves `system.ErrNoSpelling`; and a custom rule returning an unset, inadmissible, or wrong-class name preserves `system.ErrBadSpelling`.

`Number.Note` applies the same spelling filter and creates a new note with the requested correction in cents. The zero correction is exactly zero; a supplied positive or negative correction is passed unchanged. The new note is bound to the original source. The operation reads that source once to choose the pitch and once more inside `note.FromPitch`. If the source's settings can be replaced concurrently, the caller must synchronize that replacement with the whole operation so both snapshots are equal. An immutable settings value already satisfies this requirement.

## Errors and concurrent use

Sentinel errors distinguish an invalid number, unset inputs, a note without an octave or source, unset settings, a nontraditional layout, an inadmissible input name, an invalid class, spelling failures, and an out-of-range result. Errors from dependencies remain discoverable with `errors.Is`. Every error-returning operation has a `Must` variant that calls the ordinary operation and panics with the same error; ordinary operations do not panic for domain failures.

Numbers and conversions keep no shared mutable state and start no goroutines. They are safe for concurrent use when the supplied source and custom spelling function satisfy their own contracts. Inputs, notes, systems, and sources are not changed.

First released in v0.32.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/midi) · [Go examples](../../midi/example_test.go) · [Roadmap](../../ROADMAP.md)
