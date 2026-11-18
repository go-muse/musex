// Package scale represents a finite, ordered list of complete notes.
//
// New copies its inputs. The zero Scale is a present empty list, while a nil
// *Scale is absent. Notes, Clone, and value assignment preserve independence
// from later list mutations. Copying a pointer preserves object identity.
// Sources attached to notes remain borrowed from their owners.
//
// Move, Take, and the iterators wrap around stored notes without modifying
// their octaves. Iterators capture their note snapshot when created, can be
// traversed repeatedly or concurrently, and stop immediately on consumer exit.
//
// Comparisons distinguish complete notes, records, names, positions, and exact
// sound in one tuning. Context-dependent operations validate all inputs before
// producing an answer. FindSound additionally supports pairwise tolerance.
//
// Mutations are atomic on error. Reading concurrently is safe subject to note
// source guarantees. Mutating the same Scale requires caller synchronization.
// Independent copies can be edited concurrently. Must operations call their
// error-returning counterparts and panic with the original error on failure.
package scale
