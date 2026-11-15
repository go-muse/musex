// Package interval provides named intervals: the distance between two note
// names, two written pitches, or two notes, given by a size and a quality,
// with a direction, such as a major third up or an augmented fourth down.
//
// An Interval has three parts:
//
//   - the direction: Up or Down;
//   - the size n >= 1: the number of letters the interval spans, counting both
//     ends (1 for a unison, 3 for a third, 8 for an octave, 10 for a tenth);
//   - the quality, stored as the deviation d from the basic kind of the size,
//     an alteration value of package alteration.
//
// An interval is computed from spelling: from the letters and the alteration
// values of the names, without a musical system. It is not a number of
// semitones: B sharp to C up is a diminished second, C to D double flat is a
// diminished second too, and both differ from the unison and from each other
// as intervals, although in the traditional system they all have the size of
// zero steps.
//
// # Quality and groups
//
// The simple sizes 1, 4, 5, and 8 form the perfect group, whose basic kind is
// perfect; 2, 3, 6, and 7 form the major-minor group, whose basic kind is
// major. A compound size belongs to the group of its simple part. The
// deviation d says by how many chromatic units the interval is wider than the
// basic kind. Whole deviations give the familiar qualities: perfect (0 in the
// perfect group), major (0) and minor (-1) in the major-minor group, augmented
// of multiplicity k (d = k), and diminished of multiplicity k (d = -k in the
// perfect group, -1 - k in the major-minor group). Interval.Quality reports
// the kind and Interval.Multiplicity the multiplicity.
//
// Fractional deviations are qualities too, for systems with fractional
// alteration values: the neutral quality is exactly d = -1/2 in the
// major-minor group (C to E with -1/2 is the neutral third); any other
// fractional deviation is reported as Fractional. FromDeviation creates an
// interval with any deviation in the range of package alteration. There is no
// other limit on the multiplicity.
//
// # The unison and the octave
//
// The perfect unison has no direction: it is always ascending, the opposite
// of P1 is P1, and Down or WithDirection keep it ascending. There is no
// diminished unison: lowering a name on the same letter gives an augmented
// unison down (C to C flat in one register is the descending A1). A unison
// with a positive deviation does have a direction.
//
// The simple part of an interval keeps its quality and direction and has the
// size 1 for a unison and 2 to 8 otherwise: P15 is P8 plus one octave, M10 is
// M3 plus one octave. Interval.Invert inverts the simple part: the size
// becomes 9 - n and the quality changes major to minor, augmented to
// diminished of the same multiplicity, perfect to perfect, and neutral to
// neutral. An augmented octave has no inversion, because it would be a
// diminished unison: Invert reports ErrNoInversion.
//
// # Sum and the line of fifths
//
// The sum of two intervals is defined by moving names: moving any name or
// pitch by i.Add(j) gives the same result as moving it by i and then by j. So
// M3 + m3 is P5, P5 + P4 is P8, and m2 plus a descending A1 is d2. Addition
// is commutative and associative, P1 is its neutral element, and the sum of an
// interval and its opposite is P1. Interval.Sub and Interval.Mul are defined
// through the sum.
//
// Names with whole alteration values have places on the line of fifths:
// F -1, C 0, G 1, D 2, A 3, E 4, B 5, plus 7 for each sharp. FifthsOf,
// FifthsBetween, FromFifths, and Interval.Fifths relate these places to
// intervals; fractional alteration values and qualities have no place and
// give ErrNotOnFifths.
//
// # Written pitches and systems
//
// Names need no system: Between, BothWays, Transpose, and Reflect work on
// names alone, and the octaves of an interval do not change a name. Written
// pitches and notes with octaves need a musical system to give the written
// octave, which tells where the octave of the writing begins: B sharp in the
// positional octave 4 is written B sharp 3 in the traditional system. The
// system is explicit for pitches (BetweenPitches, TransposePitch,
// ReflectPitch); for notes it comes from their settings or is given
// explicitly by the functions with the suffix In. No system is ever chosen by
// default. The direction of an interval between pitches follows the writing,
// not the position: in the traditional system B sharp 3 to C4 is d2 up, at
// the same position.
//
// # Positional size
//
// Interval.Steps gives the positional size of an interval in a system: the
// number of steps of the system between its lower and upper notes, which can
// be zero or negative (d2 and dd2 in the traditional system). In a regular
// system the size does not depend on the names, and the start name can be
// omitted by passing the zero spelling.Spelling{}; in other systems the size
// depends on the start name and ErrStartRequired is returned without it.
// Without a start name, a size that is not a whole number of steps, such as
// the neutral third in the traditional system, gives ErrFractionalSteps; with
// a start name, both names must be admissible in the system. FromSteps,
// AllFromSteps, Interval.Enharmonics, EnharmonicallyEqual, and Respellings
// work in the other direction or compare intervals by their sizes.
//
// # Notes
//
// BetweenNotes and TransposeNote work with notes of package note. The cent
// correction of a note does not take part in intervals. A transposed note is
// a new note with the correction and the source of the original one, and the
// original note is never changed. Notes without octaves are names; notes with
// octaves need a system from their settings or an explicit one. TransposeNote
// reads the source of a bound note with an octave twice, once for the system
// and once when the record of the new note is replaced; if the settings of
// the source change in between, the two reads see different settings.
//
// # Short notation and names
//
// Interval.String, Interval.MarshalText, and Interval.AppendText write the
// short notation, the machine form for storage and exchange: an optional "-"
// for a descending interval, a sign of quality, and the size, such as "M3",
// "-m6", "AA4", "n3", or "{+1/4}3" (a third 1/4 wider than major). Parse and
// Interval.UnmarshalText read it back, and Parse also accepts other signs of
// quality such as "maj3", "+4", "°5", and "perf5".
//
// Interval.Name, Interval.RussianName, and Interval.RussianAbbreviation give
// the English name, the Russian name, and the Russian abbreviation, such as
// "major third", "большая терция", and "б.3". Only names from the domain base
// are given: sizes up to the fifteenth, multiplicities up to 3, and a few
// neutral intervals; for the others the methods report that there is no name
// by returning false, which is not an error. ParseName reads names and
// abbreviations in many accepted forms.
//
// # Range
//
// MaxSize, equal to math.MaxInt, is the largest size, and the deviation is
// limited only by the range of package alteration. All operations are exact:
// a result that can be represented is returned even when an intermediate
// value would not fit, and a result that cannot be represented gives
// ErrOutOfRange, never a wrapped value.
//
// # Errors
//
// Errors are checked with errors.Is; their text is not part of the API. Each
// error matches at most one sentinel of this package, and an error caused by
// a dependency also wraps the error of the dependency: ErrOutOfRange together
// with alteration.ErrOutOfRange, octave.ErrOutOfRange, or
// system.ErrOutOfRange. A name that is not admissible in an explicit system
// gives an error that matches system.ErrInadmissible and no sentinel of this
// package; a name that is not admissible in the system of the settings of a
// note also matches note.ErrInadmissible, and the errors of reading the
// settings of a note are those of package note.
//
// When an operation has several problems, the first is reported in this
// order of categories: unset inputs (ErrUnset); impossible combinations and
// parameters (ErrInvalidArgument); a note without an octave (ErrNoOctave); a
// missing direction (ErrNoDirection); a note without a system (ErrNoSystem);
// the errors of reading the settings of notes; different systems
// (ErrDifferentSystems); a missing start name (ErrStartRequired); names that
// are not admissible; no place on the line of fifths (ErrNotOnFifths); no
// inversion (ErrNoInversion); a fractional number of steps
// (ErrFractionalSteps); and results out of range. Each category is checked
// for all inputs, in the order of the parameters, before the next one. Parse
// and ParseName report ErrParse before the other errors, and their error
// texts quote at most 64 bytes of the input, followed by "..." when it is
// longer.
//
// Every function and method with an error result has a Must variant, except
// MarshalText, AppendText, and UnmarshalText: MustNew, Interval.MustDown, and
// so on. A Must variant returns the same results without the error and panics
// with the same error value in the same cases, so that calls can be chained:
// interval.MajorThird().MustDown().MustInvert(). Must variants are intended
// for values known to be valid; use the variants with an error result for
// values taken from data. No other function or method panics.
//
// # Values and concurrency
//
// The zero Interval is the unset interval. It is not the perfect unison: it
// is equal to no interval under Interval.Equal, its methods without an error
// result return zero values and false, its String is "Interval(unset)", and
// operations with an error result report ErrUnset. For set intervals == and
// Equal agree, and Interval can be used as a map key; Interval.Compare gives
// an agreed order for sorting.
//
// Intervals are immutable values, and the package has no mutable state: the
// ready-made intervals and the tables of names cannot be changed. All
// functions and methods are safe for concurrent use, except that a variable
// written by Interval.UnmarshalText must not be used by another goroutine at
// the same time. The sequences returned by AllFromSteps, Interval.Enharmonics,
// and Respellings can be iterated many times and concurrently. Operations
// with notes read their sources as package note describes.
package interval
