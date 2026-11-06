// Package pcset provides pitch-class sets: unordered subsets of the cyclic
// space Z_N of N pitch classes (1 <= N <= [MaxModulus]), together with
// set-theoretic operations, transpositions, inversions, multiplications,
// operators x ↦ m·x + n, indices of equivalence and embedding, positional
// search, cyclic rotations, normal forms (Rahn and Forte), interval-class
// vectors, symmetry and orbit measures, the 8-element Morris invariance
// vector for N = 12, subset enumeration and counting, and canonical and
// literary text notations.
//
// A [Set] is an unordered subset of Z_N. An [Algorithm] selects the packing
// rule ([Rahn] by default, or [Forte]) for the normal form of a set. An
// [IntervalVector] holds the interval-class multiplicities of length
// floor(N/2). An [InvarianceVector] holds the 8-element Morris invariance
// vector <a, b, c, d, ā, b̄, c̄, d̄> for N = 12.
//
// Functions and methods whose names start with Must are wrappers for
// literals and values known to be valid: they return the same result as the
// operation without Must, or panic with its error. For data from users, use
// the operations that return errors.
//
// # Pitch-class sets and modulus bound
//
// A pitch-class set X ⊆ Z_N is determined solely by its modulus N and the
// set of distinct class numbers in 0..N-1 that belong to it. Order of
// insertion and duplicate elements are not retained. A set does not remember
// a musical system, note spellings, octaves, a tonic, temporal order (which
// belongs to the planned pcseq package), or equivalence-class labels (which
// belong to the planned setclass package).
//
// Whereas a single class or operator in [github.com/go-muse/musex/pitchclass]
// has O(1) size for any N, operations on sets such as [Universe],
// [Set.Complement], and [Set.IntervalVector] have size proportional to N.
// Package pcset therefore enforces the engineering upper bound
// [MaxModulus] = 65,536 (2^16), which covers all traditional, microtonal,
// and analytical octave divisions (including 1200-EDO) with more than a
// 50-fold margin while guaranteeing bounded time and memory for single-set
// operations. Passing N < 1 returns [ErrInvalidModulus]; passing
// N > [MaxModulus] (including a valid [pitchclass.Class] or
// [pitchclass.Operator] with Modulus() > [MaxModulus]) returns
// [ErrModulusTooLarge].
//
// # Creation and binary codes
//
//   - [Empty] and [Universe] create the empty set ∅ and the full aggregate
//     Z_N of modulus N.
//   - [New] creates a set strictly from class numbers in 0..N-1; [Reduce]
//     wraps arbitrary integers modulo N.
//   - [FromClasses] creates a set from a non-empty list of
//     [pitchclass.Class] values of the same N; [FromClassesWithModulus]
//     takes N explicitly and accepts an empty list.
//   - Every subset X ⊆ Z_N has a unique binary characteristic code
//     sum_{x ∈ X} 2^x in 0..2^N-1. [FromBitmask] and [Set.Bitmask] convert
//     between a set and a uint64 code for 1 <= N <= 64 (returning
//     [ErrOverflow] for N > 64); [FromBitmaskBig] and [Set.BitmaskBig]
//     work with [math/big.Int] across the entire range 1 <= N <= [MaxModulus].
//
// # Set-theoretic operations and inclusion
//
// [Set.Complement] returns the literal complement Z_N \ X. Two-place
// operations [Set.Union], [Set.Intersection], [Set.Difference],
// [Set.SymmetricDifference], [Set.IsDisjoint], [Set.IsSubsetOf],
// [Set.IsProperSubsetOf], [Set.IsSupersetOf], and [Set.IsProperSupersetOf]
// operate on sets of the same modulus N and return [ErrModulusMismatch] if
// the moduli differ.
//
// # Transformations and fixed classes
//
//   - [Set.Transpose] applies T_n(X) = {(x + n) mod N : x ∈ X}.
//   - [Set.Invert], [Set.InvertAround], and [Set.InvertSwapping] apply
//     inversion I_n(X) = {(n - x) mod N : x ∈ X} by index, axis class, or
//     swapped class pair.
//   - [Set.Multiply] applies M_m(X) = {(m·x) mod N : x ∈ X} for any integer
//     m; when gcd(m, N) > 1, distinct classes may collapse and cardinality
//     may decrease.
//   - [Set.Apply] applies an invertible [pitchclass.Operator] T_n M_m.
//   - [OperatorFixedClasses] and [MultiplicationFixedClasses] return the set
//     of classes x ∈ Z_N left fixed by an operator or by M_m.
//
// # Equivalence, embedding, and positional search
//
// Given two sets X and Y of modulus N:
//
//   - [Set.TranspositionIndicesTo] and [Set.InversionIndicesTo] return the
//     sets of indices n ∈ Z_N such that T_n(X) = Y and I_n(X) = Y.
//   - [Set.TranspositionEmbeddingsInto] and [Set.InversionEmbeddingsInto]
//     return the sets of indices n ∈ Z_N such that T_n(X) ⊆ Y and
//     I_n(X) ⊆ Y.
//   - [Set.CoveringTranspositionsOf] returns {τ ∈ Z_N : X ⊆ T_τ(template)},
//     which directly computes the admissible tonics of a scale or mode
//     template for a given set of pitch classes (positional search).
//   - [Set.EquivalenceOperatorsTI], [Set.EquivalenceOperatorsTTO],
//     [Set.EmbeddingOperatorsTI], and [Set.EmbeddingOperatorsTTO] return
//     the operators f in the T/I group or the full TTO group satisfying
//     f(X) = Y or f(X) ⊆ Y.
//
// # Rotations and normal form (Rahn and Forte)
//
// [Set.Rotations] returns the k = |X| cyclic rotations of the ascending
// class list of X. [Set.NormalForm] and [Set.NormalFormWith] choose the most
// compact rotation on the pitch-class circle:
//
//   - [Rahn] (the default) minimizes the outer span d_{k-1} = (r_{k-1} - r_0) mod N,
//     breaking ties from right to left (d_{k-2}, ..., d_1) and finally by the
//     smallest starting class number r_0.
//   - [Forte] minimizes the outer span d_{k-1}, breaking ties from left to
//     right from the first element (d_1, d_2, ..., d_{k-2}) and finally by
//     the smallest starting class number r_0.
//
// For N = 12, Rahn and Forte agree on most sets but differ on sets where the
// minimum outer span is achieved by multiple rotations with different
// internal interval distributions (such as {0, 1, 3, 7, 8} and
// {0, 1, 5, 6, 8} in Forte class 5-20).
//
// # Interval-class vector and common-tone theorems
//
// [Set.IntervalVector] returns the [IntervalVector] <m_1, ..., m_{floor(N/2)}>
// counting unordered pairs {x, y} ⊆ X of each interval class d = 1..floor(N/2).
// By the generalized complement theorem, m_d(Z_N \ X) - m_d(X) = N - 2|X|
// for d < N/2 and N/2 - |X| for even N and d = N/2. By the common-tone
// theorem for transposition, |T_n(X) ∩ X| equals m_{min(n, N-n)}(X) for
// n != N/2 and 2·m_{N/2}(X) for even N and n = N/2.
// [Set.CommonToneCountTranspose], [Set.CommonToneCountInvert],
// [Set.CommonToneCountTransposeWith], [Set.CommonToneCountInvertWith], and
// [Set.CommonToneCountWith] return common-tone counts under transformations.
//
// # Symmetry, orbits, and Morris invariance vector
//
// [Set.TranspositionSymmetryIndices], [Set.InversionSymmetryIndices],
// [Set.SymmetryOperatorsTI], [Set.SymmetryOperatorsTTO],
// [Set.DegreeOfTranspositionSymmetry], [Set.DegreeOfInversionSymmetry],
// [Set.DegreeOfSymmetryTI], [Set.DegreeOfSymmetryTTO],
// [Set.DistinctTranspositionCount], and [Set.OrbitSizeTI] describe the
// self-mappings and orbit sizes of a set for any valid modulus N.
//
// For N = 12, [Set.InvarianceVector] returns the 8-element Morris
// [InvarianceVector] <a, b, c, d, ā, b̄, c̄, d̄> counting the indices n ∈ Z_12
// under T_n, I_n, T_n M_5, and T_n M_7 that map X onto itself (a, b, c, d)
// and into its complement Z_12 \ X (ā, b̄, c̄, d̄). For N != 12, it returns
// [ErrNotModulus12].
//
// # Enumeration and counting
//
// [All] and [OfCardinality] return lazy, restartable iterators over all 2^N
// subsets of Z_N (in ascending bitmask order) and all C(N, k) subsets of
// cardinality k (in lexicographic order of ascending element lists).
// [Count], [CountBig], [CountOfCardinality], and [CountOfCardinalityBig]
// return their counts.
//
// # Canonical and literary notation
//
// [Set.String] and [Parse] format and parse the canonical representation
// "{c_0, c_1, ..., c_{k-1}} mod N" (such as "{0, 4, 7} mod 12" and
// "{} mod 12"). [Set.Notation], [Set.CompactNotation],
// [Set.NormalFormNotation], [Set.NormalFormCompactNotation], and
// [ParseNotation] format and parse literary notations without the modulus
// using [pitchclass.Style].
//
// # Errors and invalid values
//
// Every error returned by this package matches one of the 11 exported
// sentinel errors with [errors.Is]: [ErrInvalidModulus],
// [ErrModulusTooLarge], [ErrInvalidClass], [ErrInvalidCardinality],
// [ErrInvalidIntervalClass], [ErrModulusMismatch], [ErrNotModulus12],
// [ErrInvalidValue], [ErrInvalidAlgorithm], [ErrNotation], and
// [ErrOverflow].
//
// The zero values Set{}, IntervalVector{}, and InvarianceVector{} are
// invalid (IsValid reports false) and are never treated as an empty set or
// zero vector of modulus 12. For valid values, == and Equal agree; for two
// identical zero values, == is true in Go while Equal reports false.
//
// # Concurrency
//
// All types in this package are immutable values, and the package has no
// mutable global state. All functions, methods, and returned iterators are
// safe for concurrent use by multiple goroutines without external
// synchronization.
package pcset
