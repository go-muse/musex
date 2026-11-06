// Package catalog provides immutable reference catalogs of pitch-class set
// classes in twelve-tone equal temperament (N = 12): Allen Forte's catalog of
// 224 undirected Tn/In set classes and 352 directed Tn-types (with "A" and "B"
// orientation suffixes), Elliott Carter's Harmony Book catalog of 135 chord
// classes in cardinalities 2, 3, 4, 6, and 7, and the 23 binary Z-correspondent
// pairs.
//
// The package is a self-contained reference data layer with zero dependencies
// outside the Go standard library. Dynamic normalization of arbitrary pitch-class
// sets to normal or prime form and set-class operations for arbitrary moduli N
// belong to the parent package setclass, which imports pcset and this package.
//
// The package defines six immutable comparable value types:
//
//   - [ForteClass] is one of the 224 canonical Tn/In set classes for
//     cardinalities 0 through 12 (from "0-1" to "12-1").
//   - [TnType] is one of the 352 canonical directed Tn-types for cardinalities
//     0 through 12 (96 inversionally symmetric types without a suffix and 256
//     chiral types in 128 "A"/"B" pairs).
//   - [CarterClass] is one of the 135 entries in Elliott Carter's catalog for
//     cardinalities 2, 3, 4, 6, and 7.
//   - [ZPair] is one of the 23 binary pairs of Z-correspondent Forte set
//     classes sharing an interval-class vector.
//   - [Algorithm] selects the prime form packing convention ([Rahn] by default,
//     or [Forte]).
//   - [Suffix] represents the orientation suffix of a directed Tn-type
//     ([SuffixNone], [SuffixA], or [SuffixB]).
//
// The package defines the following operations:
//
//   - [ParseForteClass], [ParseTnType], and [ResolveForteClass] parse canonical
//     Forte and Tn-type labels; [ForteClassByNumber], [TnTypeByNumber], and
//     [CarterByNumber] look up entries by numeric coordinates.
//   - [ForteClassByPrimeMask], [ForteClassByPrimeMaskWith],
//     [ForteClassByAnyPrimeMask], [ForteClassByPrimeForm],
//     [ForteClassByPrimeFormWith], and [ForteClassByAnyPrimeForm] look up a
//     [ForteClass] by its canonical 12-bit prime form mask or pitch-class slice;
//     [TnTypeByPrimeMask], [TnTypeByPrimeMaskWith], [TnTypeByAnyPrimeMask],
//     [TnTypeByPrimeForm], [TnTypeByPrimeFormWith], and [TnTypeByAnyPrimeForm]
//     perform the corresponding lookups for [TnType].
//   - [ForteClasses], [AllForteClasses], and [ForteClassesOfCardinality] list
//     Forte set classes; [TnTypes], [AllTnTypes], and [TnTypesOfCardinality]
//     list directed Tn-types; [CarterClasses], [AllCarterClasses],
//     [CarterClassesOfCardinality], [CarterCardinalities], and
//     [HasCarterCardinality] query Elliott Carter's catalog; [ZPairs],
//     [AllZPairs], and [ZPairsOfCardinality] list binary Z-pairs.
//   - Methods on [ForteClass], [TnType], [CarterClass], and [ZPair] read
//     identifiers, Rahn and Forte prime forms and masks, interval-class vectors,
//     Morris invariance vectors, symmetry orders, orbit sizes, minimum spans,
//     complements, inversions, Z-partners, M5-partners, and Carter numbers.
//
// Functions and methods whose names start with Must are wrappers for literals
// and values known to be valid: they return the same result as the corresponding
// operation without Must, or panic with its error.
//
// # Forte set classes and directed Tn-types
//
// Under the 24 operations of the twelve-tone transposition and inversion group
// Tn/In, the 2^12 = 4096 subsets of Z_12 partition into 224 undirected set
// classes ([ForteClass]): Allen Forte's (1973) classical 208 classes in
// cardinalities 3 through 9, together with the 16 boundary classes in
// cardinalities 0, 1, 2, 10, 11, and 12 ("0-1", "1-1", "2-1".."2-6",
// "10-1".."10-6", "11-1", "12-1"). The counts by cardinality 0..12 are
// 1, 1, 6, 12, 29, 38, 50, 38, 29, 12, 6, 1, 1.
//
// Under the 12 transpositions Tn alone, the 4096 subsets partition into 352
// directed Tn-types ([TnType]). Exactly 96 of the 224 Forte classes are
// inversionally symmetric (achiral, InSymmetry > 0) and correspond to a single
// directed Tn-type without an orientation suffix (such as "0-1", "3-1",
// "5-Z12", "6-Z29"). The remaining 128 Forte classes are chiral
// (InSymmetry == 0) and split into two inversionally related directed Tn-types
// with suffixes "A" ([SuffixA], containing the parent class's prime form) and
// "B" ([SuffixB], its inversion), such as "3-11A" (minor triad, [0, 3, 7]) and
// "3-11B" (major triad, [0, 4, 7]), yielding 96 + 2*128 = 352 Tn-types. The
// counts by cardinality 0..12 are 1, 1, 6, 19, 43, 66, 80, 66, 43, 19, 6, 1, 1.
//
// # Prime forms: Rahn and Forte
//
// Each [ForteClass] and [TnType] provides both standard canonical prime form
// conventions:
//
//   - [Rahn] (John Rahn, 1980; default in musex) packs intervals most tightly
//     toward the left by comparing outer spans from right to left.
//   - [Forte] (Allen Forte, 1973) packs intervals from left to right among
//     forms of minimal outer span.
//
// Prime forms are exposed both as a fresh ascending slice of pitch-class
// numbers in 0..11 ([ForteClass.PrimeForm], [ForteClass.FortePrimeForm],
// [ForteClass.PrimeFormWith]) and as a 12-bit characteristic bitmask in
// 0..4095 ([ForteClass.PrimeMask], [ForteClass.FortePrimeMask],
// [ForteClass.PrimeMaskWith]), where bit p is 1 iff pitch class p belongs to
// the prime form.
//
// On the 224 Tn/In classes, Rahn and Forte prime forms differ on exactly 6
// classes ([ForteClass.RahnForteDiffer]): "5-20", "6-Z29", "6-31", "7-Z18",
// "7-20", and "8-26" (giving 224 + 6 = 230 distinct canonical prime forms).
// On the 352 directed Tn-types, Rahn and Forte Tn-prime forms differ on
// exactly 17 Tn-types ([TnType.RahnForteDiffer]): "5-20A", "5-20B", "5-32B",
// "6-Z29", "6-31A", "6-31B", "6-Z44B", "7-Z18A", "7-Z18B", "7-20A", "7-20B",
// "8-22B", "8-26", "8-27B", "9-7B", "9-8B", and "9-11B" (giving 352 + 17 = 369
// distinct canonical Tn-prime forms; note that for 7 chiral classes—"5-32",
// "6-Z44", "8-22", "8-27", "9-7", "9-8", and "9-11"—the Rahn and Forte forms
// agree on the parent class and the "A" branch but diverge on the "B" branch).
//
// # Z-correspondents and Z-pairs
//
// Forty-six Forte classes carry the "Z" label ([ForteClass.IsZ]) and form 23
// binary [ZPair] entries (1 in cardinality 4, 3 in cardinality 5, 15 in
// cardinality 6, 3 in cardinality 7, and 1 in cardinality 8). Two Z-related
// classes share the same cardinality and 6-element interval-class vector without
// being related by transposition or inversion.
//
// In two Z-pairs ("5-Z12"/"5-Z36" and "7-Z12"/"7-Z36"), one partner is
// inversionally symmetric while the other is chiral. In four historical Z-pairs
// ("6-Z28"/"6-Z49", "7-Z12"/"7-Z36", "7-Z17"/"7-Z37", and "7-Z18"/"7-Z38"),
// Forte's catalog assigns the smaller ordinal to the partner with the larger
// minimum span or less left-packed prime form; the catalog preserves Forte's
// exact historical numbering. For all 15 hexachordal Z-pairs (30 Z-hexachords),
// the Z-partner coincides with the set-theoretic complement.
//
// # Complements, inversions, and M5 partners
//
// Complement maps each class of cardinality c to its complement of cardinality
// 12 - c ([ForteClass.Complement], [TnType.Complement]). On the undirected
// Tn/In level, [ForteClass.IsSelfComplementary] is true for the 20 non-Z
// hexachords. On the directed Tn level without inversion, complement swaps
// suffixes "A" <-> "B" for 115 chiral classes, but preserves suffixes "A" <-> "A"
// and "B" <-> "B" for 13 chiral classes ("4-12", "4-14", "5-11", "5-26",
// "5-28", "6-Z10", "6-14", "6-Z39", "7-11", "7-26", "7-28", "8-12", "8-14").
// Consequently, [TnType.IsSelfComplementary] is true for exactly 8 directed
// hexachordal Tn-types: the 6 symmetric hexachords ("6-1", "6-7", "6-8",
// "6-20", "6-32", "6-35") and the chiral pair "6-14A" and "6-14B".
//
// Multiplication by 5 modulo 12 ([ForteClass.M5Partner], equivalent to M7 under
// Tn/In) is an involution on the 224 Forte classes with 92 fixed points
// ([ForteClass.IsM5Invariant]) and 66 mutual pairs (yielding 92 + 66 = 158
// equivalence classes under the full 48-element TTO group).
//
// # Structural invariants
//
// Every [ForteClass] and [TnType] provides its precomputed 6-element
// interval-class vector ([ForteClass.IntervalVector], counting interval classes
// 1..6), 8-element Morris invariance vector ([ForteClass.InvarianceVector],
// [h(Tn), h(TnI), h(TnM5), h(TnM7), c(Tn), c(TnI), c(TnM5), c(TnM7)]),
// transpositional symmetry order s_Tn in {1, 2, 3, 4, 6, 12}, inversional
// symmetry order s_In in {0, 1, 2, 3, 4, 6, 12}, total degree of symmetry,
// orbit size (summing to 4096 across all 224 ForteClasses and across all 352
// TnTypes), and minimum chromatic span ([ForteClass.MinSpan]).
//
// # Elliott Carter catalog
//
// Elliott Carter's Harmony Book (2002) assigns chord numbers within each
// covered cardinality. The package includes all 135 verified Carter entries
// ([CarterClass]) across cardinalities 2 (1..6), 3 (1..12), 4 (1..29),
// 6 (1..50), and 7 (1..38), with bidirectional lookup via [CarterByNumber],
// [ForteClass.Carter], and [ForteClass.CarterNumber]. For cardinalities
// 0, 1, 5, 8, 9, 10, 11, and 12, where Carter's numbering is not defined in
// verified open sources, [ForteClass.HasCarter] reports false and
// [CarterByNumber] / [CarterClassesOfCardinality] return [ErrNoCarterCatalog].
//
// # Errors, invalid values, and concurrency
//
// Operations returning an error match one or more of the exported sentinel
// errors ([ErrInvalidNotation], [ErrInvalidCardinality], [ErrInvalidOrdinal],
// [ErrInvalidSuffix], [ErrNoCarterCatalog], [ErrInvalidCarterNumber],
// [ErrNotPrimeForm], [ErrInvalidValue], [ErrInvalidAlgorithm]) via [errors.Is].
//
// The zero values ForteClass{}, TnType{}, CarterClass{}, and ZPair{} are
// invalid (IsValid reports false, Equal reports false even for two zero values,
// Compare orders invalid values before all valid values and returns 0 for two
// invalid values, and String returns "<Type>(invalid)"). All functions and
// methods are immutable, return fresh slice copies when slices are produced,
// and are safe for concurrent use by multiple goroutines.
package catalog
