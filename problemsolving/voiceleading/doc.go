// Package voiceleading analyzes written melodic lines without changing them.
//
// Analyze builds an immutable snapshot over every pair of voice.Line inputs.
// Each pair uses only its own element starts and ends at its shorter line.
// VerticalAt reads a half-open segment; TransitionAt requires two adjacent
// segment starts. Rests remain visible and interrupt classified motion.
//
// Specific intervals retain written direction, quality, and full generic size.
// Chromatic distances count signed positions in the chosen musical system,
// not frequencies or performance corrections. With no explicit system, only
// sounding elements occurring opposite another sound read settings, once per
// input element. Later source changes do not change the measured snapshot.
//
// Searches distinguish parallel motion, strict perfect parallels, consecutive
// perfect families, and direct arrivals. Filters must be selected explicitly.
// Crossings and overlaps require an explicit LowToHigh input order. Names and
// voice roles do not establish order or identity; input slots identify parts.
// CheckOMTFirstSpecies is a separate, explicit educational assessment of one
// consecutive-perfect-family rule, not a general judgment of musical quality.
//
// Nil and zero analyses are unset. Except IsSet and Equal, operations return
// ErrUnset on them. Successful empty collections are non-nil; errors return
// zero results. Every fallible operation has a Must wrapper that panics with
// the same error. Use errors.Is and errors.AsType to inspect contextual causes.
//
// Analyses may be copied by assignment and read concurrently. Returned slices,
// nested slices, and assessments belong to the caller and are independent of
// other calls. Immutable elements, positions, intervals, and systems may share
// storage; systems are borrowed and must not be overwritten by callers. Notes
// preserve their source binding and are returned by value through Element.
// Use Analysis.Equal for record and context equality, not pointer identity.
package voiceleading
