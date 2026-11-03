# Octave numbering

An octave numbering assigns a visible integer to each [octave](../octave.md) in ascending order. ASPN (also called SPN or IPN) gives the octave containing middle C the number 4. It provides the reference by which the library converts between numberings.

| Numbering | Octave containing middle C | Skipped numbers |
| --- | ---: | --- |
| ASPN | 4 | None |
| Franco-Belgian | 3 | 0 |
| Mexican | 5 | None |
| Middle C = C3 | 3 | None |
| Middle C = C5 | 5 | None |

## What defines a numbering

One anchor pair establishes the correspondence between an ASPN number and a visible number. Away from that anchor, numbers advance by one per octave, passing over any declared skips. In Franco-Belgian numbering the visible sequence around zero is −2, −1, 1, 2. A lookup for 0 has no answer.

Numberings are equal when they assign the same numbers to every octave. Mexican and Middle C = C5 are equal by this rule. Franco-Belgian and Middle C = C3 agree at middle C but differ lower down because only the former skips 0. Labels identify conventions for readers and have no effect on equality.

Custom numberings follow the same anchor-and-skips rule. Conversion preserves octave identity and reports an error when a number is skipped or a result exceeds the range of an `int`. The zero value denotes an undefined numbering.

First released in v0.4.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/octave/number) · [Go examples](../../../octave/number/example_test.go) · [Roadmap](../../../ROADMAP.md)
