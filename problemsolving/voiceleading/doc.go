// Package voiceleading analyzes written melodic lines and chords without
// changing them.
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
// NewMelody and Analysis.Melody measure one voice: every move between two
// consecutive sounds, with its written interval, chromatic distance, and a
// mark for a move through a rest. Repeats, steps, and leaps are read from a
// move; a leap needs an explicit LeapDefinition. LeapRuns finds runs of
// leaps, and Melody.Check applies one explicitly selected melodic norm:
// OMTLeapRecovery, OMTMelodicIntervals, FuxAllowedMotions, or
// MT21ChoraleMelody. Moves through a rest are never checked.
//
// NewSonority measures simultaneous notes named by voice from the bass
// upward; notes are never sorted. Analysis.Sonorities and SonorityAt give the
// sonorities of all parts at every element start before the joint end of all
// parts. A sonority describes its adjacent intervals and span, classifies its
// spacing by an explicit SpacingDefinition, checks a DistanceNorm, and lists
// its tone classes, doublings, and missing or extra classes of an expected
// chord, by spelling or by sound. Spacing, distances, and connections require
// LowToHigh; a part sonority with a silent part is outside every spacing and
// distance definition, which is a result rather than an error.
//
// Decompose splits a sequence of chords into lines from the bass upward by
// position, assuming that voices do not cross, and marks equal positions;
// DecomposeInOrder keeps the record order of every chord. The lines are ready
// for Analyze, so the pairwise searches apply to chord sequences too.
//
// Connect describes the connection of two sonorities with an equal number of
// sounding voices: every voice move, common classes, retained common tones,
// and the total chromatic movement. CheckMT21Connection applies the explicit
// MT21 norm for root-position triads with a doubled bass, as asserted by the
// caller; no root is detected. ChooseConnections lists every choice of upper
// voices toward a target chord and bass that passes the selected conditions,
// ordered by total movement. It returns the same set as a full enumeration;
// the number of combinations grows exponentially with the number of voices.
// None of these results is a judgment of musical quality or a harmonization.
//
// Analysis.Equal is unchanged by these operations: it compares pairwise
// verticals and transitions and does not see the elements of a part after all
// pairwise intersections. The new methods of an analysis measure such tails in
// the system of the analysis and never read sources of settings; NewMelody
// with a nil system checks every note's system by its source instead.
//
// Nil and zero analyses, melodies, and sonorities are unset. Except IsSet and
// Equal, operations return ErrUnset on them. Successful empty collections are
// non-nil; errors return zero results. Every fallible operation has a Must
// wrapper that panics with the same error. Use errors.Is and errors.AsType to inspect contextual causes
// such as PartError, ChordError, and MeasurementError.
//
// Analyses, melodies, and sonorities may be copied by assignment and read
// concurrently. Returned slices, nested slices, and assessments belong to the
// caller and are independent of other calls. Immutable elements, positions,
// intervals, and systems may share storage; systems are borrowed and must not
// be overwritten by callers. Notes preserve their source binding and are
// returned by value. Use the Equal methods for record and context equality,
// not pointer identity.
package voiceleading
