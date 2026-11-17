// Package midi converts between MIDI note numbers and the written musical model.
//
// A Number is an integer position from 0 through 127 in the ordinary MIDI map:
// C in positional octave -1 is 0, C4 is 60, A4 is 69, and G9 is 127. The zero
// Number is invalid, not MIDI 0. Min, Max, Count, and All expose the complete
// range; Shift, ShiftOctaves, their neighboring forms, and Sub never wrap,
// clamp, or saturate. Parts returns an ASPN positional octave, a class modulo
// 12, and the numeric C-layout position P = m-12. FromParts performs the
// inverse numeric assembly without consulting a musical system.
//
// Conversion from a written pitch requires an explicit system with the
// traditional C layout: twelve classes, offsets C0 D2 E4 F5 G7 A9 B11,
// chromatic unit 1, and an octave boundary immediately before C. The
// traditional alteration rule is not required: a custom system may restrict or
// extend admissible alterations. Other layouts, including the same offsets with
// a different boundary letter, are rejected. Conversion from a note reads one
// whole settings snapshot, either from the note itself or from an explicit
// source. FromNoteIn never reads or changes the note's own source.
//
// MIDI stores a position, not a spelling. B-sharp written as B♯3 and C4 both
// produce 60, while C-flat written as C♭4 and B3 both produce 59. Reverse
// conversion therefore requires an explicit system.Sharps, system.Flats, or
// system.CustomSpelling rule. system.ChromaticScale and unknown rule kinds are
// rejected with ErrNotApplicable before system.PitchAt is called. A correction
// in cents does not affect conversion to MIDI; Number.Note passes the supplied
// correction unchanged to the new note.
//
// Positional Octave is not the register digit in a written record. B♯3 has
// positional octave 4, and C♭4 has positional octave 3. This distinction also
// determines the MIDI number. Frequencies, temperament, reference pitch, MIDI
// messages and files, channels, velocity, and duration are outside this package.
//
// Number.Note preserves the original source by calling note.FromPitch with it.
// That constructor reads the source again, so the operation requires both reads
// to return equal settings. A caller that can replace a source's settings
// concurrently must synchronize that replacement with the entire operation. An
// immutable *settings.Settings needs no such external synchronization.
//
// All values and operations have no shared mutable state and are safe for
// concurrent use when their sources and custom spelling functions satisfy their
// own contracts. The package starts no goroutines. Every operation that returns
// an error has a Must variant which calls it and panics with the same error.
package midi
