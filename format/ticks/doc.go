// Package ticks converts the exact musical time of musex to the integer units
// of exchange formats and back: MusicXML divisions, MEI ppq and Standard MIDI
// File (SMF) ticks for durations, positions and measure addresses; MIDI tempo
// and tempo maps; MIDI time signatures and their changes; SMPTE time and MIDI
// clocks. The package does not read or write files or bytes.
//
// # Resolution and the MIDI quarter
//
// A Resolution is D ticks per MIDI quarter together with the written length w
// of that quarter, so one tick is w/D of a whole note. MusicXML divisions and
// MEI ppq always count written quarters (w = 1/4). An SMF counts MIDI quarters,
// which the field bb of its Time Signature event writes as bb thirty-second
// notes: usually bb = 8 and w = 1/4, but with bb = 16 the MIDI quarter is
// written as a half note. MIDIQuarter holds bb; Resolution.WithQuarter and the
// optional MIDIQuarter arguments of the tempo functions select it, and the
// standard quarter is the default.
//
// Ticks, PositionTicks and the measure address operations convert exactly: a
// value that is not a whole number of ticks is ErrNotRepresentable, never
// rounded silently. Units returns the exact number of units as a fraction, and
// Quantize rounds once with an explicit mode of package common/rounding and
// reports the exact deviation. MinResolution, AtLeast, AtMost, Represents and
// Unrepresentable choose a resolution for a set of values. Rescale and
// QuantizeRescale convert ticks between resolutions; Deltas and Absolute
// convert absolute ticks and SMF delta-times. A Resolution has no upper limit:
// FitsSMF, WithinMusicXML and Word report the SMF limit 32767 and the MusicXML
// recommendation 16383.
//
// # SMPTE time
//
// SMPTE is an SMPTE resolution of an SMF division word: a frame rate code and
// ticks per frame, with ticks as fractions of a second. Code -29 is read
// differently by different sources, so its seconds need an explicit Reading:
// 30000/1001 (drop-frame), 30 or 29 frames per second. DecodeDivision decodes
// both kinds of division words. SMPTE time code (hours, minutes, seconds,
// frames) is not part of the package.
//
// # MIDI tempo and tempo maps
//
// MIDITempo is a Set Tempo value, 1 through 16777215 microseconds per MIDI
// quarter. MIDITempoOf rounds the exact value of a tempo marking once with an
// explicit mode; MIDITempo.Tempo returns the exact marking with a chosen unit.
// QuartersPerMinute and FormatQuartersPerMinute give the MusicXML sound tempo
// and the MEI midi.bpm.
//
// MIDITempoMap is an immutable list of tempo changes by tick with an initial
// tempo of 500000 microseconds (120 MIDI quarters per minute) before the first
// change; on one tick the last change is in effect. It converts ticks and
// positions to exact seconds and back, and to SMPTE ticks. TempoMap converts
// it to a musex tempo map of new main tempos; MIDITempoMapOf converts a musex
// tempo map to MIDI with ExportOptions: the rounding mode of tempos, an
// optional rounding of positions, steps for gradual changes (the speed at the
// start or the end of each step, or steps that last as long as that part of
// the change), steps on MIDI clocks, and the handling of fermata holds
// (refuse, skip and report, or stretch the segment before the hold). Measure
// maps of package tempomap take one measure.Grid and are converted to
// positions first.
//
// # Correct rounding
//
// Every integer result is rounded once from the exact value with an explicit
// mode, without double rounding, also when the exact value is irrational on a
// gradual tempo change. The package uses the step methods of package tempomap,
// which round to any step without a precision limit, and exact comparisons
// built on them. The only exception is a stretched equal-time step whose step
// and stretched segment are both irrational: there the rounding is decided by
// intervals of up to 76 decimal places and, extremely close to a rounding
// boundary, fails with ErrOutOfRange instead of returning a wrong integer.
//
// # MIDI time signatures
//
// MIDIMeter holds the four fields nn, dd, cc and bb of an SMF Time Signature
// event. MIDIMeterOf converts a meter of package meter with a Click (explicit
// or on the beat) either strictly or with a reported loss of symbols, additive
// numerators, sums and explicit grouping. MIDIMeterMapOf converts a measure
// grid, including pickups and bars of explicit length, to events by tick, and
// MIDIMeterMap.Grid converts events back to an open grid. Events with
// different bb in one conversion are not supported.
//
// # Values, text and errors
//
// All types are immutable values with value receivers; the zero value of each
// struct type is unset, and enumerations treat zero as invalid. Every
// fallible operation other than text marshaling has a Must variant that
// panics with the same error. Types with a text form implement fmt.Stringer,
// encoding.TextMarshaler, encoding.TextAppender and encoding.TextUnmarshaler
// and have a Parse function that reads the canonical text back. Parse errors
// quote at most 64 bytes of input and stay below 512 bytes.
//
// Errors match exactly one of ten package sentinels with errors.Is: ErrUnset,
// ErrInvalidArgument, ErrZero, ErrNotRepresentable, ErrFormatLimit, ErrOrder,
// ErrNotConvertible, ErrOutOfRange, ErrInvalidText and ErrMeterNotMIDI. Causes
// from dependencies remain visible. Failing calls return zero results and do
// not change their inputs.
//
// # Concurrency
//
// The package has no global mutable state, caches or goroutines. All functions
// and methods are safe for concurrent use; as usual, UnmarshalText must not
// write through a pointer while another goroutine reads the same value.
package ticks
