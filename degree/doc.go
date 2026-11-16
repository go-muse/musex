// Package degree represents an ordinal degree and counts degree numbers in a
// caller-defined circle.
//
// A Degree contains a positive ordinal number and an optional note.Note. It
// knows nothing about a mode, tonic, template, letters, or musical system. The
// convenience functions First through Seventh are only seven ready-made
// number-only degrees; they neither limit numbers nor define a seven-degree
// reference mode.
//
// # Values and copies
//
// The zero Degree is unset: IsValid is false, Number is 0, HasNote is false,
// and Note returns the unset note. A set degree without a note is different
// from the unset degree and can accept a note later. Any set note is allowed,
// with or without an octave, with any correction, and bound or unbound.
//
// Notes are stored by value. Creating or changing a degree copies the supplied
// note, every call to Note returns another independent copy, and assignment of
// a Degree copies both its number and note independently in both directions.
// Only a note's source of settings remains shared, under the contract of note.
// There is no Clone method because assignment is sufficient.
//
// # Changing values
//
// SetNote, RemoveNote, TransformNote, and SetNumber change an addressable
// degree in place. Each operation is atomic: an error leaves the receiver
// wholly unchanged. A chain of Must variants is not atomic as a whole; each
// step succeeds or panics separately and returns the same pointer.
//
// Changing methods require an addressable value, such as a variable, a slice
// element, or a pointer. This does not compile:
//
//	degree.MustNew(4).MustSetNote(n)
//
// Use a variable instead:
//
//	d := degree.MustNew(4)
//	d.MustSetNote(n).MustSetNumber(5)
//
// Go 1.26 and later also permits new(degree.MustNew(4)).MustSetNote(n).
// Reading methods work on function results and map elements. To change a map
// element, copy it to a variable, change that variable, and put it back.
// Remember that ranging over a slice by value yields copies; use indices or
// pointers when changes must affect the slice.
//
// A nil *Degree reports ErrUnset from a changing method and panics with that
// error from its Must variant. A reading method invoked through a nil pointer
// panics under the usual Go rules before entering the method.
//
// # Equality and ordering
//
// Degree is deliberately not comparable with == and cannot be a map key.
// Comparing interface values containing degrees compiles but panics at run
// time. reflect.DeepEqual compares representation, not domain equality. Use
// Equal for equality of number and note, SameNumber to ignore notes, SameNote
// to compare only two present notes, and Compare for the agreed total order.
// The unset degree equals nothing, including another unset degree.
//
// # Number text and debugging
//
// Decimal and Roman format only the number, never the note. Decimal accepts
// every supported number. Roman uses canonical uppercase subtractive notation
// through 3999. Parse accepts exactly those two canonical forms and returns a
// degree without a note. String is a separate debugging form that includes
// the number and note or their absence; it is not parsed back.
//
// # Circular counting
//
// Every circular operation receives the circle size k explicitly. Next,
// Previous, Shift, DistanceUp, DistanceDown, ShortestDistance, Normalize, and
// Denormalize return or consume numbers independently of notes. Every degree
// returned by these operations and by All has no note; the input is unchanged.
// ShortestDistance is positive upward and negative downward, and a tie goes
// upward. Normalize maps a positive continuous number to its circle position
// and nonnegative cycle count; Denormalize performs the inverse operation and
// detects overflow.
//
// Shift is exact for all int inputs, including math.MinInt and math.MaxInt,
// and uses no arbitrary-precision arithmetic. All is lazy: it does not allocate
// a slice of k values, supports early termination, and permits repeated and
// concurrent iterations. Even All(MaxNumber) starts safely; callers remain
// responsible for limiting an impractically long traversal.
//
// # Concurrency and failures
//
// There is no global mutable state. Pure functions, reading, comparison,
// formatting, circular arithmetic, and sequence iteration may run
// concurrently. One degree may be read concurrently, but a changing method
// needs exclusive access; different copies may be changed concurrently. A
// TransformNote call is as safe as the supplied transform and anything it
// captures.
//
// Operations with failures return one of seven sentinel errors, inspectable
// with errors.Is: ErrUnset, ErrInvalidNumber, ErrNoNote, ErrParse,
// ErrRomanOutOfRange, ErrOutOfRange, and ErrTransformFailed. Input errors are
// checked by category: unset inputs first, then invalid numbers, a missing
// note, Roman range, and result range. Parse checks syntax, zero, and decimal
// overflow in that order. A transform failure preserves both
// ErrTransformFailed and the original cause. Every failing operation has a
// Must variant that panics with the same error value; no other operation
// panics except the documented Go nil-pointer and user-transform cases.
package degree
