// Package search finds mode-and-tonic hypotheses whose actual degree notes
// contain a given set of notes or equal it as a set. It does not infer the mode
// of a musical passage, estimate likelihood, or rank answers.
//
// # Conditions and matching
//
// Containment accepts a mode when every distinct input element occurs among its
// notes. FullMatch additionally requires the two sets to be equal. Input order,
// duplicates, octaves, and cent corrections never matter. Open modes and closed
// modes with a non-octave cycle participate equally: only the actual k degrees
// are compared, without a continuation or another cycle.
//
// Searcher.BySpelling compares exact spelling.Spelling values, so C sharp and D
// flat differ. Searcher.ByPosition compares pitch classes in one explicit
// system. Searcher.ByPositionFromContext instead reads one snapshot from each
// distinct input source and requires equal systems. The explicit form never
// reads note sources.
//
// # Families and representatives
//
// A positional answer retains a TonicFamily rather than reducing enharmonic
// spellings to one arbitrary name. TonicFamily.Spellings lazily enumerates all
// compatible names within an explicitly supplied absolute-alteration limit.
// PositionMatch.Tonic is only the deterministic display representative: the
// smallest absolute alteration, then the most exact matches to input names,
// then spelling.Compare order. The limit applies to enumeration, not to which
// families the search can find.
//
// Results preserve catalog order. Spelling results then use spelling order;
// positional results use class number and, for nonregular systems, letter order.
// This stable order is not musical ranking. Six grouping functions present
// existing results by collection, tonic, or structure without another search.
//
// # Values, errors, and concurrency
//
// Searcher and its results are immutable and safe for concurrent reading.
// Accessors return independent slices, and Mode returns an independent value.
// A spelling sequence can be stopped, traversed again, or traversed concurrently;
// it starts no goroutines and owns no resources requiring closure.
//
// Errors match the package sentinels with errors.Is and preserve dependency
// causes. Invalid input rejects the whole operation. In contrast, a candidate
// whose tonic or degree name is inadmissible or unrepresentable is skipped alone;
// an empty successful result is not an error. Every fallible operation has a Must
// wrapper that panics with the operation's error.
package search
