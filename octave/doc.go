// Package octave provides octaves: elements of the axis of octaves, with
// their visible numbers in any octave numbering and their names in any
// catalog of octave names.
//
// An [Octave] is an immutable, comparable value. The package works with the
// numberings of package [github.com/go-muse/musex/octave/number] and the
// catalogs of package [github.com/go-muse/musex/octave/naming].
//
// The package defines the following operations:
//
//   - [FromASPN] returns the octave with a given ASPN number; [FromNumber]
//     and [MustFromNumber] return the octave with a given number in any
//     [number.Numbering]; [FromName] and [MustFromName] return the octave
//     with a given name in any [naming.Catalog].
//   - [Octave.Number] and [Octave.MustNumber] return the number of an octave
//     in a numbering; [Octave.ASPN] and [Octave.MustASPN] return its ASPN
//     number; [Octave.Name] returns its name in a catalog.
//   - [Octave.IsValid] reports whether a value is an octave.
//   - [Octave.Equal] reports whether two values are the same octave;
//     [Octave.Compare] orders octaves from lower to higher; [Octave.Above],
//     [Octave.Below], and [Octave.Between] compare positions on the axis.
//   - [Octave.Shift], [Octave.ShiftDown], [Octave.Up], and [Octave.Down]
//     move along the axis; [Octave.Sub] returns how many octaves one octave
//     is higher than another, and [Octave.Distance] returns how many octaves
//     lie between them.
//   - [Range] and [MustRange] return the octaves from one to another, one at
//     a time.
//   - [NewNumbering] and [MustNewNumbering] define an own numbering by an
//     octave and its number; [NewCatalog] and [MustNewCatalog] define a
//     catalog of names by octaves.
//   - [Octave.String] and [Octave.GoString] return fixed forms for output.
//
// Every operation that returns an error has a Must variant.
//
// # Octaves
//
// An octave is an element of the ordered sequence of octaves, which has no
// end in either direction; neighboring octaves differ by exactly one octave.
// The identity of an octave is its place in this sequence, and it is separate
// from its number and its name: the same octave is number 4 in ASPN and
// number 3 in the Franco-Belgian numbering, "первая" in the Russian catalog
// and "one-line" in the English one. The number and the name are
// computed from the numbering or the catalog passed to an operation; they are
// not stored in the value. The internal representation of the axis is not
// visible and is not part of the API.
//
// The ASPN number of an octave is a visible number that identifies the octave
// uniquely, and it is the common reference of the packages
// [github.com/go-muse/musex/octave/number] and
// [github.com/go-muse/musex/octave/naming]. That is why their documentation
// says that octaves are identified by their ASPN numbers: it agrees with this
// package, in which an octave has an ASPN number like a number in any other
// numbering.
//
// The octave of a note is a section of its position in a musical system.
// Which octave of the axis a note selects, and how it relates to the octave
// number written in the name of the note, such as the 3 in "B♯3", is decided
// by the system, not by this package. In the traditional system, an octave
// is the range of pitches from C up to the next B.
//
// # Representable octaves
//
// The axis has no musical limits: octaves below ASPN 0 and above ASPN 9 are
// ordinary octaves, and the MIDI range does not apply. The limits are
// technical: the representable octaves are exactly the octaves whose ASPN
// number fits in an int, from math.MinInt to math.MaxInt. An operation whose
// octave, number, difference, or distance does not fit in an int returns an
// error that matches [ErrOutOfRange]; results never wrap around and are never
// replaced with the nearest value.
//
// A number that [number.Convert] converts to another numbering without an
// error may not give a representable octave. This happens with numbers whose
// ASPN number does not fit in an int: number.Convert(math.MaxInt,
// number.MiddleC3(), number.FrancoBelgian()) returns math.MaxInt without an
// error, while FromNumber(math.MaxInt, number.MiddleC3()) returns an error
// that matches ErrOutOfRange, because that octave would have ASPN number
// math.MaxInt + 1. Not every number at the edge of int is like this:
// FromNumber(math.MinInt, number.MiddleC3()) is the octave with ASPN number
// math.MinInt + 1.
//
// # Invalid value
//
// The zero value of [Octave] is the invalid value: it is not an octave. It is
// not the octave with ASPN number 0 or 4, and it does not mean that a note
// has no octave; whether a note has an octave is a concern of the note.
// No operation treats it as any octave:
//
//   - [Octave.IsValid] reports false;
//   - [Octave.Number] and [Octave.ASPN] return an error that matches
//     [ErrInvalid], even for an undefined numbering;
//   - [Octave.Name] returns "" without an error;
//   - [Octave.Equal] returns false, even for two invalid values;
//   - [Octave.Compare] puts it below every octave and returns 0 for two
//     invalid values; [Octave.Above], [Octave.Below], and [Octave.Between]
//     return false if any value is invalid;
//   - [Octave.Shift], [Octave.ShiftDown], [Octave.Up], [Octave.Down],
//     [Octave.Sub], [Octave.Distance], [Range], [NewNumbering], and
//     [NewCatalog] return an error that matches [ErrInvalid] if an octave
//     given to them is invalid;
//   - [Octave.String] returns "Octave(invalid)", and [Octave.GoString]
//     returns "octave.Octave{}".
//
// Because the type is comparable, Octave{} == Octave{} is true, while
// Octave{}.Equal(Octave{}) is false and Octave{}.Compare(Octave{}) is 0, like
// NaN for floating-point numbers. For octaves, == and [Octave.Equal] always
// agree.
//
// # Errors
//
// The package has two errors of its own: [ErrInvalid] for the invalid value
// and [ErrOutOfRange] for a result that does not fit in an int. It passes on
// the errors of the numberings and catalogs, wrapped with context:
// [number.ErrUndefined] for an undefined numbering, [number.ErrSkipped] for a
// skipped number, [number.ErrInvalidDefinition] for an invalid own numbering,
// [naming.ErrUnknownName] for a name that is not in a catalog, and
// [naming.ErrDuplicateName] for the same name given to two octaves.
//
// Check errors with [errors.Is]. Each returned error matches exactly one of
// these seven causes. The only double match is one cause: when [FromNumber]
// or [Octave.Number] gets [number.ErrOutOfRange] from package number, the
// error matches both [ErrOutOfRange] and [number.ErrOutOfRange], so a single
// check for ErrOutOfRange covers every case of this package.
// When the package wraps an error of package number or naming, it does not
// repeat the name or the numbering: the wrapped error describes them in a
// bounded form, so the text stays short however long the name is or however
// many numbers are skipped. The text of the errors is not part of the API.
//
// The conditions are checked in this order: first the octaves given to an
// operation, then the numbering, catalog, or definition, and last the range
// of the result. On error, operations return the zero Octave, 0, a nil
// sequence or catalog, or the zero Numbering; ignore these values. Only the
// functions and methods whose names start with Must panic, in the same cases
// in which their variants return an error, with that error as the panic
// value.
//
// # String forms
//
// [Octave.String] returns "ASPN 4" for the octave with ASPN number 4 and
// "Octave(invalid)" for the invalid value; [Octave.GoString] returns
// "octave.FromASPN(4)" and "octave.Octave{}". They are for logs, debugging,
// and tests: they are not a notation such as "C4", contain no names, are not
// localized, and have no parser.
//
// Package fmt uses String for %v, %+v, %s, %q, %x, %X, and output without a
// verb, such as [fmt.Println], and GoString for %#v. Other verbs, such as %d,
// are applied by fmt to the fields of the value; their output is not part of
// the API.
//
// # Concurrency
//
// Octaves are immutable values, and the package has no mutable state. All
// functions and methods, and the sequences returned by [Range], are safe for
// concurrent use by multiple goroutines, on the same values or on copies,
// under the usual Go rules for writes: do not modify the map passed to
// [NewCatalog] or the slice passed to [NewNumbering] during the call.
package octave
