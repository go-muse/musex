# Alterations

An alteration measures how far an accidental raises or lowers a letter. Its unit is the change made by one sharp. The five traditional values are double flat (−2), flat (−1), natural (0), sharp (+1), and double sharp (+2). Zero denotes the natural even when it is the zero value of the Go type.

## Exact chromatic changes

A value can extend beyond the traditional accidentals or fall between them. The fraction +1/2 is half the change made by a sharp. Adding two such changes gives exactly +1. The representation reduces fractions to lowest terms, so +2/4 and +1/2 compare equal and have the same stored numerator and denominator.

Raising, lowering, and multiplication produce new values. Subtracting a flat from a sharp gives +2 chromatic units. That difference measures accidentals alone. To determine the positions or pitches of two written notes, the musical system and their letters also matter.

The package guarantees exact arithmetic for magnitudes through 2²⁴ chromatic units and denominators through 2¹⁶. It reports a result outside that range as an error. A [spelling](spelling.md) combines one of these values with a [letter](letter.md).

First released in v0.2.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/alteration) · [Go examples](../../alteration/example_test.go) · [Roadmap](../../ROADMAP.md)
