# Pitch classes

A pitch class identifies a position within a repeating octave. In a system with *N* positions, its number runs from 0 to *N* − 1. Counting continues around the circle, so in twelve positions the next class after 11 is 0. A class retains its number and *N*. The register and spelling are separate information.

In the familiar twelve-position system with C at 0, C♯ and D♭ both correspond to `1 mod 12`. Their [spellings](spelling.md) remain different. The musical system supplies this correspondence. The pitch-class package starts with numbered positions.

## Distances and transformations

The ordered interval from 10 to 2 modulo 12 is four upward steps. The return journey takes eight. Their interval class is the shorter distance, four. Comparing positions and composing transformations requires the same modulus throughout.

Transposition adds a fixed number of steps. Inversion with index *n* sends *x* to *n* − *x* around the circle. Multiplication sends *x* to *mx*. With twelve positions, multiplying by seven takes the chromatic sequence 0, 1, 2, … to 0, 7, 2, …, the circle of fifths.

An operator multiplies and then transposes. Its multiplier must be invertible modulo *N* so that every class has a unique image and preimage. For twelve positions there are four such multipliers (1, 5, 7, 11) and twelve transpositions, yielding 48 operators. Composition applies the inner operator first.

## Representing a class

The canonical forms include the modulus, for example `1 mod 12` for a class and `T5M7 mod 12` for an operator. Literary styles display just the class number and may use T/E, t/e, or A/B for ten and eleven. Sorting by modulus and number provides a stable listing of labels. Their order around the circle supplies no absolute pitch height.

First released in v0.7.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/pitchclass) · [Go examples](../../pitchclass/example_test.go) · [Roadmap](../../ROADMAP.md)
