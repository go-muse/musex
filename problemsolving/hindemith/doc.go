// Package hindemith analyzes dyads, voiced chords and chord progressions
// with Paul Hindemith's harmonic theory: Series II, interval roots, the
// table of chord groups and harmonic fluctuation.
//
// # Model and sources
//
// The model follows Hindemith, The Craft of Musical Composition, Book I
// (1937/1945), pp. 68-106: Series II ranks the intervals by diminishing
// harmonic value, every interval except the tritone has a root (the lower
// tone of unisons, octaves, fifths, thirds and sevenths; the upper tone of
// fourths, sixths and seconds), and the table of chord groups places every
// simultaneity into groups I-VI with subgroups. The chord root is the root
// of the strongest intra-chord interval; groups V and VI (augmented-triad
// and diminished-seventh-like content) have no root in isolation. Harmonic
// fluctuation is the rise and fall of harmonic tension along the group
// order of a progression.
//
// Sources: the chord-group table and Series II figure reproduced at
// https://digital.library.unt.edu/iiif/ark:/67531/metadc663498/m1/22/full/1200,/0/default.jpg
// and the formalization in Apostolopoulos et al., "The ontocomc Ontology: A
// Semantic Framework for Hindemith's Harmonic Theory" (CEUR Vol-4091,
// paper 16), https://ceur-ws.org/Vol-4091/paper16.pdf.
//
// # Tuning assumptions
//
// The theory is stated for twelve-tone equal temperament. Every operation
// takes an explicit system with twelve classes and chromatic unit 1 and
// measures intra-chord intervals by their signed distance in that system,
// reduced modulo 12: enharmonic spellings of one tempered interval (a
// diminished fifth and an augmented fourth, an augmented fifth and a minor
// sixth) are one Series II interval. Compound intervals keep their root
// side (lower or upper tone) but take the Series II position of their
// simple part; a distance of whole octaves is the octave, position 2.
//
// # Rules of this implementation
//
// Series II positions are 1 to 13: perfect unison, octave, perfect fifth,
// perfect fourth, major third, minor sixth, minor third, major sixth, major
// second, minor seventh, minor second, major seventh, tritone.
//
// A chord is classified by its intra-chord interval content: without a
// tritone, content of only unisons/octaves, major thirds and minor sixths
// with at least one third or sixth is group V; otherwise absence of seconds
// and sevenths gives group I and their presence group III. A chord of one
// doubled note is group I with that note as its root. With a tritone, content of only tritones, minor
// thirds and major sixths is group VI; otherwise absence of minor seconds
// and major sevenths gives group II and their presence group IV. Groups V
// and VI have no root; in the other groups the root is the root of the
// strongest interval, unison pairs ignored, ties broken by the first pair
// in input order. Subgroups 1 and 2 mean root equal to the bass and root
// above the bass; group II adds subgroup a (one tritone, no major second,
// root in the bass) and subgroups b1, b2, b3 (one tritone with root in the
// bass, one tritone with root above the bass, more than one tritone). A
// group-II chord with neither a major second nor a minor seventh takes b1
// or b2 by root position.
//
// Harmonic fluctuation orders the classes I.1, I.2, II.a, II.b1, II.b2,
// II.b3, III.1, III.2, IV.1, IV.2, V, VI by rising tension; a step to a
// higher value raises tension, a step to a lower value releases it.
//
// The result is an analysis under these stated rules, not a claim about
// the composer's intention or the heard center of the music.
package hindemith
