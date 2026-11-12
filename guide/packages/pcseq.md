# Pitch-class sequences and rows

A Sequence is a finite ordered list of [pitch classes](pitchclass.md) in one modulus *N*, where 1 ≤ *N* ≤ 65,536. Position and repetition matter: repeating a class or exchanging two positions produces a different sequence. Every modulus also has a valid empty sequence, distinct from an invalid zero value. A Row is the stricter special case: it contains every class of *Z*_*N* exactly once and therefore has length *N*. Unlike a [pitch-class set](pcset.md), these values preserve order; unlike a row class, an individual row retains its particular succession of classes.

## Transformations and segments

Transposition *T*_*n*, inversion *I*_*n*, multiplication *M*_*m*, and invertible affine operators act elementwise around the circle. Retrograde reverses the order, retrograde inversion combines inversion with reversal, and rotation moves positions cyclically. Multiplication need not be invertible: applying it to a Row yields a Sequence, because several classes may collapse onto one another and the row invariant may be lost. Ordered intervals connect each class to its successor, while content discards order and repetition to form a set.

Segments use zero-based positions. A single segment may be empty at any boundary from 0 through the sequence length; overlapping segments cover every window of a positive length; discrete partitioning requires a positive length that divides the whole sequence. Partitioning an empty sequence yields no parts rather than one fictitious empty part. Set-based comparison returns every transposition or inversion index relating two segments, including all indices when both contents are empty.

## Forms and the row matrix

For a source row *p* with first class *p*₀, the labelled forms are *P*_*n* = *T*_(n−p₀)(*p*), *I*_*n* = *I*_(n+p₀)(*p*), *R*_*n* = *R*(*P*_*n*), and *RI*_*n* = *R*(*I*_*n*). Thus *P* and *I* are labelled by the first result class, while *R* and *RI* are labelled by the last. *P*₀ always normalizes the first class to zero: fixed zero and moveable zero are two cases of the same rule, and there is no separate *IR* label.

The matrix is determined entirely by its source row, with cell *m*_*i*,*j* = (*p*_*j* − *p*_*i*) mod *N*. Its row *i* is the *P*-form indexed by *p*₀ − *p*_*i*, and column *j* is the *I*-form indexed by *p*_*j* − *p*₀. Reading a row backwards gives the corresponding *R*-form; reading a column upwards gives the corresponding *RI*-form. No table is cached in the matrix value; the full square is created only when requested.

## Row classes

Three equivalence relations group rows into classes: transposition, inversion, and retrograde (*T*/*I*/*R*); the same operations together with all circular rotations; or all invertible affine operators together with retrograde (*TTO* + *R*). Each class has a lexicographically minimal canonical representative, so equivalent inputs produce the same value. Row classes report their size, test membership, and materialize unique forms in deterministic order without retaining hidden lists or tables.

## Derivation, combinatoriality, and interval structure

A derived row consists of equal-length discrete segments related to the first segment by *T*_*n*, *I*_*n*, *T*_*n**R*, or *I*_*n**R*. Decomposition returns every valid operator for every segment—including the generator itself—in family order and then ascending index. Distinct operator families are preserved even when they have the same action. An absent connection makes the row non-derived rather than silently selecting one arbitrary decomposition.

For an even modulus, two forms are combinatorial when the content of the first half of one equals the complement of the first half of the other. The analysis covers individual pairs, *P*-, *I*-, and *RI*-forms, and rotations. At *N* = 12, a six-class first half also has separate transposition, inversion, and retrograde-inversion combinatoriality lists. An all-interval row contains every nonzero adjacent interval exactly once; the one-class row is not all-interval, while either two-class row is.

## Text and parsing

Canonical text includes the modulus, as in `<0, 1, 2> mod 12`, with the empty sequence written `<> mod 12`. Literary notation uses `⟨0, 1, 2⟩` or compact forms such as `⟨04TE⟩`; decimal, T/E, t/e, and A/B element styles are supported. Parsing accepts both Unicode and ASCII angle brackets but never guesses modulus, style, case, bracket shape, or spacing. Labels such as `P0`, `I3`, `R6`, and `RI7` round-trip through the same strict rules. Leading zeros, signs, extra text, unsupported compact notation, and invalid row content are rejected rather than normalized silently.

## Limits and ownership

Basic sequences, rows, transformations, rotations, and segments support *N* up to 65,536. Matrices and analysis or materialization that could grow quadratically are limited to 1,024; rotation classes are limited to 1,024, and *TTO* + *R* classes and *TTO* content classification to 256. The global *T*/*I*/*R* class count and special historical counts are provided only within their established limits, with the special all-interval totals restricted to *N* = 12. Machine-sized counts reject overflow; exact counts use independent arbitrary-precision integers.

Values are immutable and safe to copy and use concurrently. Inputs are not retained as mutable buffers. Returned slices, matrix cells, forms, analysis lists, and exact integers are newly owned by the caller, and changing them cannot alter later results. Basic transformations allocate a bounded result buffer rather than per-element state, and values contain no hidden reusable caches.

First released in v0.16.0

[Package reference](https://pkg.go.dev/github.com/go-muse/musex/pcseq) · [Go examples](../../pcseq/example_test.go) · [Roadmap](../../ROADMAP.md)
