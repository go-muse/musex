// Package tuning assigns frequencies to written pitches.
//
// A Tuning combines four immutable parts: a musical system, a temperament, the
// temperament root within that system, and a Reference. A reference is a written
// pitch together with a normal positive frequency in hertz. Frequencies are
// defined only for written pitches, which include an octave, not for pitch names
// or classes without an octave. Step numbers are counted from the reference
// note: it is step zero, the next higher position is step one, and the previous
// position is step minus one. Enharmonic spellings of one position have the same
// step number, cent height, and frequency.
//
// Cent heights and distances are exact cents.Cents values. Frequencies use one
// correctly rounded application of a cent value to the reference frequency.
// ShiftedFrequency adds any signed cent shift before that single conversion;
// this is a general operation, not a note-correction property. Standard returns
// the traditional twelve-position system, 12-EDO rooted at C, and A4 at 440 Hz.
// The standard reference does not by itself imply a choice of temperament.
//
// Moving a reference to another note at the frequency already assigned to that
// note preserves the pitch level. Changing the reference frequency changes the
// level and scales every assigned frequency. LevelDifference and LevelRatio
// compare levels, including references expressed through different notes.
// SoundsSame is deliberately different: it compares computed frequencies as
// exact float64 values through frequency.Equal, with no tolerance. Consequently,
// reexpressing a reference can produce mathematically equivalent frequencies
// that differ in their last bits and make SoundsSame return false. Use level
// comparison to test domain-level equivalence instead. Tunings with different
// systems never sound the same; changing the root of an equal temperament does
// not change the sound.
//
// Nearest and NearestPitch find a pitch for a measured frequency and return its
// signed deviation in cents. NearestBy and NearestPitchBy select between the two
// neighboring steps with any common/rounding.Mode. NearestMode is HalfDown:
// the nearest step wins and an exact midpoint resolves to the lower step.
// Constructing a pitch from a step always requires an explicit spelling rule.
//
// Eight ready-made references are available. A4At440 is the current
// international reference standard. A4At435 is the former international
// standard in Vienna, preceding 440 Hz. A4At442 and A4At444 are the ends of the
// 442-444 Hz range reflecting a modern tendency toward a higher orchestra
// pitch; neither is a standard. A4At453 and A4At466 are historical tuning-fork
// frequencies reached at times in the Paris and Vienna opera houses; neither is
// a standard. C5At523 is an ordinary tuning fork in the second octave, used
// less often than A4 at 440 Hz; it is not a standard, and its source does not
// establish a temperament. C0At16 is the scientific pitch standard.
//
// Errors match exactly one sentinel of this package. Range and spelling errors
// also preserve the relevant cause from cents or system. Errors from
// temperament, common/rounding, and frequency are translated rather than
// exposed through the error chain. Every operation with an error result has a
// Must variant that calls it and panics with the same error value.
//
// Reference and Root are small comparable values. Tuning is used through an
// immutable pointer; its zero value and nil are unset. All values are safe for
// concurrent use, except that the safety of a custom spelling function is the
// caller's responsibility. Values contain no derived fields, mutable caches,
// or package state. Operations on one pitch use a constant number of dependency
// calls, nearest-step lookup is logarithmic in the number of classes, and
// SoundsSame and ChangedClasses are linear in that number. Constructors and
// With methods allocate one new Tuning; other allocations belong to the exact
// cent arithmetic and dependency operations. Standard returns a fresh Tuning
// pointer on every call. The shared system returned by System must not be
// overwritten through a dereferenced pointer, because that would affect the
// system's own built-in value throughout the program.
package tuning
