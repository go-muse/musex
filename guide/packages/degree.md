# Degree

A degree is an ordinal number with an optional [note](note.md). First, second, third, and so on: the number says which place it is, independently of any mode, tonic, letter, or musical system.

The note is what stands at that place. Questions about its name, octave, correction, settings, or frequency belong to the note itself. A degree does not duplicate those operations and does not check that its note fits any particular mode. Any set note is allowed, and a degree can also exist without one.

Seven ready-made degrees are provided for convenience. They are not a statement that every mode has seven places and do not make major or minor a reference: a degree with any positive number can be created.

## Set, unset, and changing values

An unset degree is the technical zero value and has no number. It differs from a set degree without a note, which has a positive number and can accept a note later.

A degree is a value in a variable. Its note and number can be changed in place: put or replace a note, remove it, transform it with an operation of the note, or assign another number. Changing the note preserves the number; changing the number preserves the note. Each individual change is all or nothing, so a refused operation leaves the whole degree as it was. A chain of changes is not atomic as a whole.

Assignment makes an independent copy of a degree in both directions. Notes are copied when they enter a degree and whenever they are read, so changing a note outside cannot change the degree that supplied or received it. Only a note's source of settings remains shared, as it is for copies of the note itself.

Reading one degree is safe concurrently. Changing it requires exclusive access; different copies can be changed concurrently.

## Equality and number notation

Degrees can be compared as a whole, only by number, or only by note. Two set degrees without notes are equal when their numbers match, but they do not thereby have “the same note”: there is no note to compare. The agreed order sorts by number, puts a degree without a note before one with a note at the same number, and then uses the agreed order of notes. An unset degree equals nothing, even another unset degree.

The number has two canonical written forms:

- decimal ASCII digits without a sign, spaces, or leading zeros, available for every supported number;
- uppercase Roman numerals with the usual subtractive notation, from 1 through 3999.

Parsing accepts exactly these forms and produces a degree without a note. The note never appears in either number form. A separate debugging string shows the number and the note or its absence; it is not a serialization.

## Circular counting

Circular operations work on numbers only. The caller supplies the circle size for each operation; it is not stored in the degree and is not a property of a mode. After the last number comes the first, and before the first comes the last.

The package supports membership, next and previous numbers, shifts with a signed cycle count, distances upward and downward, and the signed shortest distance. Positive shortest distance means upward and negative means downward; equal distances choose upward. Results of circular operations contain no note, even when the input degree had one.

A continuous number can also be normalized to a circle position and a nonnegative cycle count, then denormalized back. For example, number 8 in a circle of 7 is the first number of the next cycle. Arithmetic remains exact at the limits of the integer type, and an unrepresentable denormalized result is reported explicitly.

Degrees of a circle can be enumerated lazily from first to last. No slice sized for the whole circle is allocated, early termination is safe, and repeated or concurrent traversals are independent. A very large circle still requires the caller to limit how much of it is traversed.

The [mode package](mode.md) uses degrees to describe a mode's structure, count transitions in its cycle, and derive notes from a caller-supplied starting note. Accidentals relative to a tonic, degree names, solfège, and mode formulas are also outside this package.

First released in v0.30.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/degree) · [Go examples](../../degree/example_test.go) · [Roadmap](../../ROADMAP.md)
