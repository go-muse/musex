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

## Matching observations

An observation pairs an ASPN octave number with the displayed number seen for that octave, for example in pitch data imported from another tool. `MatchBuiltins` returns every built-in numbering consistent with all the observations, in the fixed built-in order and with labels preserved: ASPN 4 displayed as 3 matches Franco-Belgian and Middle C = C3; also observing ASPN 1 displayed as 0 leaves only Middle C = C3. ASPN 4 displayed as 5 retains both Mexican and Middle C = C5, which give the same numbers but have distinct labels. Contradictory observations give an empty result; without observations all five built-ins are returned. `Matches` applies the same check to any numbering, including a custom one; an undefined numbering matches nothing.

First released in v0.4.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/octave/number) · [Go examples](../../../octave/number/example_test.go) · [Roadmap](../../../ROADMAP.md)
