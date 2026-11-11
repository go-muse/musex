// Package pcseq provides immutable ordered sequences of pitch classes and
// rows that contain every class of their modulus exactly once.
//
// Sequence preserves order, position, and repetition. Row is a permutation of
// Z_N and adds labelled P, I, R, and RI forms, matrices, equivalence classes,
// derivation, combinatoriality, and all-interval analysis. Canonical text and
// all literary pitch-class styles support lossless round trips.
//
// Values are comparable and contain no mutable shared buffers or hidden
// caches. Returned slices and big integers are owned by the caller, and
// independent calls are safe for concurrent use.
package pcseq
