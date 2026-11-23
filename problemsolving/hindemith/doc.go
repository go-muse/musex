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
// the summary of the groups at
// https://ux1.eiu.edu/~pdhesterman/old/analysis/hindemith.html (group V:
// augmented triad and perfect quartal triad; group VI: diminished triad
// and diminished seventh), and the formalization in Apostolopoulos et al.,
// "The ontocomc Ontology: A Semantic Framework for Hindemith's Harmonic
// Theory" (CEUR Vol-4091, paper 16), https://ceur-ws.org/Vol-4091/paper16.pdf.
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
// Doublings never change the classification. A chord is first reduced to
// the lowest note of every pitch class. The pitch class is the position in
// the system modulo 12, not the spelling, so enharmonic doublings such as
// G♭4 and F♯5, or C4 and B♯4, are one class. Of two notes of one class at
// the same position (a unison, also in different spellings such as C♯4
// and D♭4) the first in input order stays, as for the bass. The group,
// subgroup, tritone count, best interval, root, the comparison of root and
// bass, and the fluctuation value are computed on the reduced chord only;
// higher repetitions of a tone add no interval. A doubling below the
// lowest occurrence of a tone replaces it in the reduced chord and so may
// move the bass and the root: the result is that of the reduced chord. The
// evidence still shows the whole input: Chord.Notes returns every input
// note and Chord.Dyads every input pair.
//
// The reduced chord is classified by its interval content: without a
// tritone, three different tones a major third apart (the augmented triad)
// or a perfect fourth apart (the perfect quartal triad), in any voicing,
// are group V; otherwise absence of seconds and sevenths gives group I and
// their presence group III. A chord of one doubled tone (unisons and
// octaves only) is group I, subgroup 1, with its lowest note, the first of
// equal positions, as bass and root. With a tritone, content of only
// tritones, minor thirds and major sixths is group VI (diminished triad,
// diminished seventh and the like); otherwise absence of minor seconds and
// major sevenths gives group II and their presence group IV. Groups V and
// VI have no root; in the other groups the root is the root of the
// strongest interval of the reduced chord, and among equal best intervals
// the pair with the lower bottom note decides, so the result does not
// depend on the listing order. Each tritone of the reduced chord is one
// pair of pitch classes, so a doubled tritone member adds no tritone.
//
// Subgroups 1 and 2 mean root equal to the bass and root above the bass;
// group II adds subgroup a (one tritone, no major second, root in the bass)
// and subgroups b1, b2, b3 (one tritone with root in the bass, one tritone
// with root above the bass, more than one tritone). Every group-II chord
// contains a major second or a minor seventh; subgroup a is the chord with
// a minor seventh and no major second.
//
// Harmonic fluctuation orders the classes I.1, I.2, II.a, II.b1, II.b2,
// II.b3, III.1, III.2, IV.1, IV.2, V, VI by rising tension; a step to a
// higher value raises tension, a step to a lower value releases it.
//
// The result is an analysis under these stated rules, not a claim about
// the composer's intention or the heard center of the music.
package hindemith
