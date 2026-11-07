# Set classes

A set class groups [pitch-class sets](pcset.md) of the same modulus *N* (1 ≤ *N* ≤ 65,536) that map onto one another under a chosen symmetry group: transposition *T*_*n*, transposition and inversion *T*_*n*/*I*_*n*, or the full affine group *TTO* = { *T*_*n* *M*_*m* : gcd(*m*, *N*) = 1 }. Where a pitch-class set retains its exact positions around the circle, a set class captures the underlying interval structure shared across its orbit. For twelve positions, live set classes connect directly to the reference entries in [`setclass/catalog`](setclass-catalog.md) (`ForteClass`, `TnType`, and `CarterClass`).

## Prime forms, orbits, and hierarchy

Each equivalence class is represented by its prime form—the most tightly left-packed normal form starting on 0 across all group images of a member set. Ties in outer span are broken either from the right by the Rahn convention (the default) or from the left by the Forte convention; every class also carries its maximal orbit bitmask (`CollinsMask`).

Generating the orbit enumerates all distinct member sets and records the symmetry degrees (*DS*_*T*, *DS*_*I*, *DS*_*T*/*I*, and *DS*_*TTO*), while linking queries find every operator in the group that maps a member set to the prime form or sends one equivalent set onto another. Moving up or down the group hierarchy projects a class to a coarser group (`Tn` → `TnIn` → `TTO`) or splits a class into its constituent orbits in a finer group.

## Structural relations, embeddings, and enumeration

Taking the set-theoretic complement of any representative induces the complement class of cardinality *N* − *k* and identifies self-complementary classes when *N* is even and *k* = *N*/2. Classes of the same modulus, group, and cardinality that share the same interval-class vector without being equivalent form *Z*-partners—generalizing twelve-tone *Z*-pairs to arbitrary moduli, including *Z*-triples and *Z*-quadruples at *N* = 16.

Lewin's embedding number `EMB(X, Y)` counts how many member sets of class *X* lie inside a concrete set or representative of class *Y*, building complete abstract subclass and superclass profiles across all cardinalities. Burnside–Pólya cycle indices compute exact class counts for any modulus and cardinality without full enumeration, while canonical, literary, and Morris normal-form notations format and parse set classes (`Tn(0, 4, 7) mod 12`, `(047)`, `3-11[037]`).

First released in v0.13.0.

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/setclass) · [Go examples](../../setclass/example_test.go) · [Roadmap](../../ROADMAP.md)
