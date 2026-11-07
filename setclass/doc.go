// Package setclass provides pitch-class set classes modulo N and their
// algebraic, transformational, and Inclusion/Embedding relations.
//
// A [SetClass] represents an equivalence class (orbit) of pitch-class sets
// ([pcset.Set]) in the cyclic universe Z_N (1 <= N <= [MaxModulus]) under one
// of three canonical transformation groups ([Kind]):
//   - [Tn]: transpositional equivalence (T_n), also known as Tn-types or prime
//     forms under transposition alone (352 classes at N = 12, including the
//     empty set and the aggregate).
//   - [TnIn]: transpositional and inversional equivalence (T_n / I_n), the
//     classical Forte set classes (224 classes at N = 12).
//   - [TTO]: affine equivalence under the group of twelve-tone operators
//     T_n M_m with gcd(m, N) = 1 (1 <= N <= [MaxTTOModulus]; 158 classes at
//     N = 12, where M_5 and M_7 pair non-invariant Forte classes).
//
// # Prime Forms and Packing Algorithms
//
// Every [SetClass] stores its canonical representative in John Rahn's (1980)
// prime form ([Rahn]), which packs intervals toward the left from the right
// (outermost) span inward. Allen Forte's (1973) historical packing convention
// ([Forte]), which tie-breaks from the left after minimizing the outer span, is
// supported alongside Rahn's via [SetClass.FortePrimeForm],
// [SetClass.PrimeFormWith], and related methods. At N = 12, [Rahn] and [Forte]
// differ on 6 [TnIn] classes (5-20, 6-Z29, 6-31, 7-Z18, 7-20, 8-26), 17 [Tn]
// types (10 across both branches of 5-20, 6-Z29, 6-31, 7-Z18, 7-20 and 7 in B
// branches: 5-32B, 6-Z44B, 8-22B, 8-27B, 9-7B, 9-8B, 9-11B), and 2 [TTO]
// classes (6-Z24/Z39, 8-Z15/Z29); [SetClass.RahnForteDiffer] reports whether
// the two conventions produce distinct prime forms for a given class. For
// 1 <= N <= [MaxBitmaskModulus], [SetClass.CollinsMask] computes the maximum
// binary bitmask representation across the orbit.
//
// # Orbit Operations, Symmetry, and the Z-Relation
//
// For 1 <= N <= [MaxOrbitModulus] (and 1 <= N <= [MaxTTOModulus] for [TTO]),
// the package provides membership and equivalence tests ([SetClass.Contains],
// [AreEquivalent], [AreEquivalentTn], [AreEquivalentTnIn],
// [AreEquivalentTTO]), connecting transposition and inversion
// indices and operators ([SetClass.TranspositionIndicesTo],
// [SetClass.InversionIndicesTo], [SetClass.OperatorsTo],
// [SetClass.OperatorsFrom], [FindSetOperators]), orbit materialization
// ([SetClass.Members]), stabilizer degrees ([SetClass.TnSymmetry],
// [SetClass.InSymmetry], [SetClass.SymmetryDegreeTI],
// [SetClass.SymmetryDegreeTTO], [SetClass.OrbitSize]), chirality and
// inversional/transpositional symmetry predicates, group hierarchy projections
// ([SetClass.ToTnIn], [SetClass.ToTTO], [SetClass.TnClasses],
// [SetClass.TnInClasses]), interval-class and invariance vectors, abstract
// complementation, inversion, multiplication ([SetClass.Multiply],
// [SetClass.M5Partner]), and Z-relation queries ([SetClass.IsZRelated],
// [SetClass.HasZPartner], [SetClass.ZPartners]).
//
// # Mod-12 Catalog Bridge
//
// At N = 12, [SetClass] converts in O(1) time to and from the static tables in
// [github.com/go-muse/musex/setclass/catalog]: [catalog.ForteClass],
// [catalog.TnType], and [catalog.CarterClass], including composite Morris
// labels such as "6-1/32" and "4-Z15/Z29" for [TTO] classes.
//
// # Abstract Inclusion, Lewin's EMB, and Enumeration
//
// The package implements abstract subset/superset inclusion
// ([SetClass.IsSubclassOf], [SetClass.Includes]), David Lewin's embedding
// number EMB(X, Y) ([SetClass.EmbeddingCount], [CountEmbedding],
// [CountSetEmbedding]) together
// with the full sub-class and super-class profiles ([SetClass.Subclasses],
// [SetClass.SubclassProfile], [SetClass.Superclasses]), exact Burnside-Polya
// orbit counting ([Count], [CountBig], [CountOfCardinality],
// [CountOfCardinalityBig]), and deterministic canonical enumeration ([All],
// [OfCardinality], [ByIntervalVector]).
//
// # Notation
//
// Set classes can be formatted and parsed in three notations:
//   - Canonical unambiguous form ([SetClass.String], [SetClass.Canonical],
//     [Parse]), such as "TnIn(0, 3, 7) mod 12" or "Tn(0, 4, 7) mod 12".
//   - Literary and compact parenthesized prime-form notation
//     ([SetClass.Notation], [SetClass.CompactNotation], [ParseNotation]), such
//     as "(0, 3, 7)" or "(037)".
//   - Robert Morris's (1987) compositional notation at N = 12
//     ([SetClass.FormatMorris], [ParseMorris], [ParseMorrisWithKind]), such as
//     "3-11[037]", "3-11B[047]", or "6-1/32[012345]".
//
// # Value Semantics
//
// [SetClass], [Kind], [Algorithm], and [SubsetProfileEntry] are immutable
// comparable value types. The zero value of each type is invalid
// (IsValid() == false) and never panics when inspected through non-Must
// methods.
package setclass
