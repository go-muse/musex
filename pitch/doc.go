// Package pitch provides written pitches: a note spelling paired with a
// positional octave, such as C sharp in ASPN 4 or B with -7/2 in ASPN -1.
//
// A [Pitch] is an immutable, comparable value that combines a note name from
// package [github.com/go-muse/musex/spelling] with an octave from package
// [github.com/go-muse/musex/octave]. Both parts are required: every valid
// pitch has a valid spelling and a valid octave, and a note without an octave
// belongs to a future note package rather than to this package. A pitch is a
// written specification, independent of any musical system, tuning, frequency
// in hertz, cent offset, or notation style.
//
// The package defines the following operations:
//
//   - [New] and [MustNew] create a pitch from a [spelling.Spelling] and an
//     [octave.Octave]; [FromASPN] and [MustFromASPN] create a pitch from a
//     spelling and an ASPN octave number.
//   - [Pitch.Spelling], [Pitch.Octave], and [Pitch.Parts] return the parts of
//     a pitch; [Pitch.ASPN] and [Pitch.MustASPN] return the ASPN number of its
//     octave.
//   - [Pitch.IsValid] reports whether a value is a valid pitch;
//     [Pitch.IsNatural], [Pitch.IsInteger], and [Pitch.Direction] report the
//     accidental properties of its spelling; [Pitch.SameSpelling],
//     [Pitch.SameOctave], [Pitch.SameLetter], and [Pitch.SameAlteration]
//     compare individual parts of two pitches.
//   - [Pitch.Equal] reports whether two pitches have the same spelling and the
//     same octave; [Pitch.Compare] defines a deterministic conventional order
//     for sorting.
//   - [Pitch.WithSpelling], [Pitch.WithOctave], and [Pitch.WithASPN] replace
//     one part while keeping the other; [Pitch.Natural] strips the alteration
//     in the same octave.
//   - [Pitch.Raise], [Pitch.Lower], [Pitch.RaiseBy], and [Pitch.LowerBy] raise
//     or lower the alteration by whole chromatic units while keeping the
//     letter and the octave.
//   - [Pitch.Shift], [Pitch.ShiftDown], [Pitch.Up], and [Pitch.Down] move a
//     pitch along the octave axis while keeping its spelling;
//     [Pitch.OctaveSub] and [Pitch.OctaveDistance] return the signed
//     difference and unsigned distance between the octaves of two pitches;
//     [Range] and [MustRange] yield pitches with the same spelling across a
//     range of octaves.
//   - [Pitch.String] and [Pitch.GoString] return fixed debug representations.
//
// Every operation that can return an error has a Must variant.
//
// # Written pitch and positional octave
//
// The octave stored in a [Pitch] is a positional [octave.Octave]: the octave
// section of a note's position in a musical system, running from one reference
// degree with zero alteration to the next. It is not the written register
// number w printed next to a note letter in score notation. In the traditional
// system, where the octave section starts on C and ends on B, the notated
// pitches "B♯3" and "C4" occupy the same positional octave ASPN 4, so their
// written pitches are (B♯, ASPN 4) and (C, ASPN 4) with written register
// numbers w = 3 and w = 4. Likewise, the notated pitches "C♭4" and "B3" occupy
// the same positional octave ASPN 3, so their written pitches are (C♭, ASPN 3)
// and (B, ASPN 3) with w = 4 and w = 3, whereas (B♯, ASPN 3) corresponds to
// notated "B♯2" and (C♭, ASPN 4) corresponds to notated "C♭5". Converting
// between a positional octave and a written register number w is the
// responsibility of a future system package, and formatting or parsing notated
// pitch strings belongs to a future notation package.
//
// Pitches are compared by how they are written, component by component: two
// pitches are equal with [Pitch.Equal] and == if and only if they have the
// same [spelling.Spelling] and the same [octave.Octave], regardless of how the
// octave was constructed (for example via [octave.FromASPN],
// [octave.FromNumber] in [github.com/go-muse/musex/octave/number], or
// [octave.FromName] in [github.com/go-muse/musex/octave/naming]). Enharmonic
// spellings are never merged: (C♯, ASPN 4) is not equal to (D♭, ASPN 4), and
// (B♯, ASPN 4) is not equal to (C, ASPN 4) or (C, ASPN 3). System positions,
// pitch classes, enharmonic equivalence, frequencies in hertz, and cent
// offsets belong to future system, tuning, and cents packages.
//
// # Working with spelling and octave parts
//
// [Pitch.Raise], [Pitch.Lower], [Pitch.RaiseBy], [Pitch.LowerBy], and
// [Pitch.Natural] change only the alteration value of the spelling while
// preserving both its letter and the positional [octave.Octave]: raising
// (B, ASPN 4) by 1 yields (B♯, ASPN 4), and stripping the accidental of
// (B♯, ASPN 4) with [Pitch.Natural] yields (B, ASPN 4), never C.
//
// To shift a pitch by a fractional or arbitrary
// [github.com/go-muse/musex/alteration.Alteration] value, or to inspect its
// [github.com/go-muse/musex/letter.Letter], alteration value, letter distance,
// or alteration difference, use [Pitch.Spelling] and [Pitch.WithSpelling], for
// example p.MustWithSpelling(p.Spelling().MustRaiseByAlteration(delta)). To
// convert the octave to a visible number in a
// [github.com/go-muse/musex/octave/number.Numbering] or a name in a
// [github.com/go-muse/musex/octave/naming.Catalog], or to compare octave
// registers with [octave.Octave.Above], [octave.Octave.Below], or
// [octave.Octave.Between], use [Pitch.Octave] and [Pitch.WithOctave].
//
// [Pitch.OctaveSub] and [Pitch.OctaveDistance] measure the signed difference
// and unsigned distance between the octaves of two pitches alone, ignoring
// their spellings. For any two pitches in the same octave—such as (C, ASPN 4)
// and (B♯, ASPN 4), or (C𝄫, ASPN 4) and (B𝄪, ASPN 4)—both methods return 0.
// Musical intervals between pitches belong to a future interval package.
//
// # Conventional order
//
// [Pitch.Compare] defines a deterministic total order on written pitches for
// sorting and stable output: it compares octaves first with
// [octave.Octave.Compare] and, within the same octave, compares spellings with
// [spelling.Spelling.Compare] (letter order C through B, then alteration
// value). This is a conventional order of written records, not an order by
// sounding pitch or by position in a musical system: within ASPN 4, (C, ASPN 4)
// sorts before (B♯, ASPN 4), and (C♯, ASPN 4) sorts before (D♭, ASPN 4).
//
// # Representable pitches
//
// Any valid [spelling.Spelling] (with alteration numerator magnitude up to
// [github.com/go-muse/musex/alteration.MaxMagnitude], 2^24, and denominator up
// to [github.com/go-muse/musex/alteration.MaxDenominator], 2^16) can be paired
// with any valid [octave.Octave] (whose ASPN number fits in an int, from
// math.MinInt to math.MaxInt). These bounds are technical limits of the Go
// representation, not musical limits. Operations whose exact result exceeds
// them return an error without truncation, saturation, or wraparound.
//
// # Invalid value
//
// The zero value of [Pitch], var p pitch.Pitch or pitch.Pitch{}, is the only
// invalid pitch value: it means that no pitch is set. Partial states with only
// a spelling or only an octave cannot be constructed. The zero value is not
// C4, C0, or A4, and it does not represent a note without an octave (which
// belongs to a future note package). No operation silently treats it as a
// valid pitch:
//
//   - [Pitch.IsValid], [Pitch.IsNatural], and [Pitch.IsInteger] report false;
//   - [Pitch.Spelling] returns spelling.Spelling{}, [Pitch.Octave] returns
//     octave.Octave{}, and [Pitch.Parts] returns both zero parts;
//   - [Pitch.Direction] returns [spelling.DirectionUnset];
//   - [Pitch.SameSpelling], [Pitch.SameOctave], [Pitch.SameLetter],
//     [Pitch.SameAlteration], and [Pitch.Equal] report false whenever either
//     operand is the zero value, including for two zero values;
//   - [Pitch.Compare] places the zero value before every valid pitch (-1) and
//     returns 0 for two zero values;
//   - [Pitch.Natural] returns pitch.Pitch{} without an error;
//   - [Pitch.ASPN], [Pitch.WithSpelling], [Pitch.WithOctave],
//     [Pitch.WithASPN], [Pitch.Raise], [Pitch.Lower], [Pitch.RaiseBy],
//     [Pitch.LowerBy], [Pitch.Shift], [Pitch.ShiftDown], [Pitch.Up],
//     [Pitch.Down], [Pitch.OctaveSub], and [Pitch.OctaveDistance] return their
//     zero result and an error matching [ErrInvalid] when the receiver or
//     argument is invalid;
//   - [Pitch.String] returns "Pitch(invalid)", and [Pitch.GoString] returns
//     "pitch.Pitch{}".
//
// Because [Pitch] is a comparable struct, pitch.Pitch{} == pitch.Pitch{} is
// true in Go, whereas (pitch.Pitch{}).Equal(pitch.Pitch{}) is false, matching
// the convention of [spelling.Spelling] and [octave.Octave]. For all valid
// pitches, == and [Pitch.Equal] agree.
//
// # Errors
//
// Check returned errors with [errors.Is]; error strings are not part of the
// API contract. Every error returned by this package matches exactly one of
// three mutually exclusive causes:
//
//   - [ErrInvalid] when a pitch receiver or argument is the zero [Pitch], or
//     when a supplied [spelling.Spelling] or [octave.Octave] is invalid;
//   - [github.com/go-muse/musex/alteration.ErrOutOfRange] when raising or
//     lowering a valid pitch produces an alteration numerator magnitude above
//     [github.com/go-muse/musex/alteration.MaxMagnitude];
//   - [github.com/go-muse/musex/octave.ErrOutOfRange] when shifting a valid
//     pitch or computing the octave difference or distance between two valid
//     pitches overflows int.
//
// Validity of the receiver and domain arguments is always checked before any
// range computation, so calling [Pitch.RaiseBy] or [Pitch.Shift] on the zero
// [Pitch] with an extreme integer argument always returns an error matching
// [ErrInvalid], never an out-of-range error and never
// [github.com/go-muse/musex/letter.ErrInvalid] or [octave.ErrInvalid].
//
// Every function or method that returns an error has a corresponding Must
// variant that returns the value directly and panics with the same error when
// the non-Must variant fails. Outside the Must variants (and method calls on a
// nil *Pitch pointer), no function or method in this package panics.
//
// # String forms
//
// [Pitch.String] returns "<Spelling> ASPN <aspn>", such as "C ASPN 4",
// "C+1 ASPN 4", "B+1 ASPN 4", or "C-1/2 ASPN -1" for a valid pitch, and
// "Pitch(invalid)" for the zero value. [Pitch.GoString] returns a valid Go
// expression using the public API, such as
// pitch.MustFromASPN(spelling.MustParse("C+1"), 4), and "pitch.Pitch{}" for
// the zero value.
//
// Package fmt uses [Pitch.String] for %v, %+v, %s, %q, %x, %X, and unformatted
// printing such as [fmt.Println], and uses [Pitch.GoString] for %#v without
// exposing unexported fields. These forms are for debugging, logs, and tests:
// they are not musical score notation, have no parser in this package, and do
// not define text or JSON serialization ([Pitch] does not implement
// encoding.TextMarshaler or json.Marshaler, so reflection-based JSON encoders
// write "{}"). Store or exchange a pitch by its parts ([Pitch.Spelling] and
// [Pitch.ASPN] or [Pitch.Octave]), or use the future notation package for
// score text.
//
// # Concurrency
//
// [Pitch] values are immutable and contain no pointers, slices, or maps, and
// the package has no mutable global state, goroutines, or locks. All functions,
// methods, and [iter.Seq] sequences returned by [Range] and [MustRange] are
// safe for concurrent use by any number of goroutines without synchronization.
package pitch
