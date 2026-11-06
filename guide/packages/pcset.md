# Pitch-class sets

A pitch-class set is an unordered collection of distinct [pitch classes](pitchclass.md) in a system with *N* positions, where 1 ≤ *N* ≤ 65,536. Repeating an element or changing the input order produces the same set, while sets of different moduli are never equal. Unlike a sequence (`pcseq`), a set records no temporal order; unlike a set class (`setclass`), it retains its specific positions rather than collapsing transposition or inversion orbits; unlike a mode (`mode`), it selects no tonic or scale degrees. Every modulus has an empty set, a full universe of all *N* positions, and a unique binary characteristic code ∑ 2^*x*.

## Operations, transformations, and positional search

Sets of the same modulus combine through complement, union, intersection, difference, symmetric difference, disjointness, and subset or superset comparison. Transposition *T*_*n*, inversion *I*_*n* (around an index, an axis class, or a swapped pair), multiplication *M*_*m*, and general operators *T*_*n* *M*_*m* act elementwise around the circle, while fixed-class queries find every class left in place by a transformation.

Comparing two sets yields all transposition or inversion indices and all *T*/*I* or *TTO* operators that map one set onto or into the other. Positional search finds every tonic *τ* for which a chord or passage lies inside the transposed template *T*_*τ*(*S*): for example, the triad `{0, 4, 7}` fits inside the twelve-position major template `{0, 2, 4, 5, 7, 9, 11}` at tonics 0, 5, and 7 (C, F, and G major).

## Normal form, interval content, and symmetry

Cycling through the ascending elements of a set produces its rotations. Normal form chooses the most compact rotation on the circle, breaking ties either from the right by the Rahn convention (the default) or from the left by the Forte convention.

The interval-class vector counts unordered pairs of each interval class from 1 to ⌊*N*/2⌋; its entries determine the common tones preserved under any transposition *T*_*n*, while the complement theorem relates the vector of a set to that of its complement. Symmetry queries measure how many transpositions, inversions, or *TTO* operators map a set onto itself and determine the size of its *T*/*I* orbit; for *N* = 12, the eight-entry Morris invariance vector summarizes invariance into the set and its complement under *T*_*n*, *I*_*n*, *T*_*n* *M*_5, and *T*_*n* *M*_7. Canonical strings include the modulus (`{0, 4, 7} mod 12`), whereas literary notations format sets (`{0, 4, 7}`, `{04TE}`) and normal forms (`[11, 2, 3, 7]`, `[E237]`) in decimal or lettered styles.

First released in v0.12.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/pcset) · [Go examples](../../pcset/example_test.go) · [Roadmap](../../ROADMAP.md)
