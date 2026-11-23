// Package sonance scores the tonal consonance of pitch-class sets and of
// voiced note collections with explicit weight profiles: interval-class
// profiles for sets, profiles over the actual tempered interval for pairs
// of pitches.
//
// # Model and sources
//
// The score is Huron's aggregate dyadic consonance: the sum, over the
// interval classes present, of the number of pairs of that class times a
// consonance weight (Huron, "Interval-Class Content in Equally Tempered
// Pitch-Class Sets: Common Scales Exhibit Optimum Tonal Consonance", Music
// Perception 11(3), 1994, pp. 289-305, doi:10.2307/40285624). The published
// weights of the six non-zero interval classes are reproduced by
// HuronProfile as quoted in Parncutt, "Consonance and Dissonance of
// Simultaneous Trichords in Western Music" (2023),
// https://hrcak.srce.hr/file/433486: IC1 -1.428, IC2 -0.582, IC3 +0.594,
// IC4 +0.386, IC5 +1.240, IC6 -0.453. The published profile weights the
// six non-zero classes of pitch-class sets; unisons and octaves (IC0) carry
// weight zero in it.
//
// A higher score means a more consonant interval inventory under the
// chosen profile; a negative score means a dissonant one. The score is a
// property of the interval content under the stated weights, not a
// prediction of heard pleasantness: timbre, register and familiarity are
// outside the model.
//
// # Counting rules
//
// ScoreSet counts the interval-class vector of a twelve-class pitch-class
// set: one count per unordered pair of distinct pitch classes, so IC0 is
// always zero and repetitions of a pitch class add nothing.
//
// ScoreNotes counts actual pitch pairs: one count per unordered pair of
// note positions, including unisons and octaves (distance 0) and including
// pairs of equal pitch classes at different octaves. A pair is counted at
// its tempered distance modulo 12, 0 to 11 semitones, under a PairProfile
// of twelve weights, so a profile can distinguish a fourth from a fifth.
// HuronPairProfile spreads the published interval-class weights over both
// members of each class; NewPairProfile accepts any twelve weights. The
// distance assumes twelve-tone equal temperament: the explicit system
// must have twelve classes and chromatic unit 1. Cents do not take part.
//
// Every score exposes its contributions: per counted interval (interval
// class for a set score, tempered distance for a note-pair score), the
// count, the weight and their product, so a caller can see which intervals
// drive the value.
package sonance
