package octave

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-muse/musex/octave/naming"
	"github.com/go-muse/musex/octave/number"
)

// Errors of the package. Together with the errors of packages
// [github.com/go-muse/musex/octave/number] and
// [github.com/go-muse/musex/octave/naming] that the package passes on, each
// returned error matches exactly one cause with [errors.Is]; it may be the
// error itself or wrap it with context. The text of the errors is not part of
// the API.
var (
	// ErrInvalid reports that an octave given to an operation, as the
	// receiver or as an argument, is the invalid value: the zero Octave. It is
	// returned by [Octave.Number], [Octave.ASPN], [Octave.Shift],
	// [Octave.ShiftDown], [Octave.Up], [Octave.Down], [Octave.Sub],
	// [Octave.Distance], [Range], [NewNumbering], and [NewCatalog].
	ErrInvalid = errors.New("octave: invalid octave")

	// ErrOutOfRange reports that the octave, number, difference, or distance
	// that an operation would return does not fit in an int. Results never
	// wrap around; the limits of int are not musical limits. It is returned
	// by [FromNumber], [Octave.Number], [Octave.Shift], [Octave.ShiftDown],
	// [Octave.Up], [Octave.Down], [Octave.Sub], and [Octave.Distance].
	//
	// When [FromNumber] or [Octave.Number] finds the cause in package
	// [github.com/go-muse/musex/octave/number], the error matches both
	// ErrOutOfRange and [number.ErrOutOfRange]: it is one cause, and a single
	// check for ErrOutOfRange covers every case in this package.
	ErrOutOfRange = errors.New("octave: octave out of range")
)

// Octave is an octave: an element of the unbounded ordered axis of octaves.
// The identity of an octave is its place on the axis; it is separate from its
// number and its name. The visible number and name of an octave are computed
// from the numbering or the catalog passed to [Octave.Number] and
// [Octave.Name]; they are not stored in the value. The internal
// representation of the axis is not visible and is not part of the API.
//
// Create an octave with [FromASPN], [FromNumber], [MustFromNumber],
// [FromName], or [MustFromName], or get one from [Octave.Shift],
// [Octave.ShiftDown], [Octave.Up], [Octave.Down], and [Range].
//
// The zero value, such as var o Octave or Octave{}, is not an octave: it is
// the invalid value. It is not the octave with ASPN number 0 or 4, and it does
// not mean that a note has no octave; that is a concern of the note, not of
// this package. [Octave.IsValid] reports false for it. Operations that return
// an octave return the zero value together with an error.
//
// Octave is an immutable value with no reference data inside. It can be
// copied by assignment and passed by value at any time; a copy is equal to
// the original and behaves the same way. No function or method changes an
// octave, its arguments, or the data of the package.
//
// The type is comparable: ==, a map key, and a switch work. Each octave has
// exactly one representation, so for octaves a == b if and only if
// a.Equal(b), however a and b were obtained, and two values of one octave are
// one map key. For the invalid value they differ: Octave{} == Octave{} is
// true, while Octave{}.Equal(Octave{}) is false, and all invalid values are
// one map key. Use [Octave.Equal] to compare octaves in domain logic.
//
// Octave has no serialization of its own, and encoders based on reflection
// do not see its fields: encoding/json writes {} for it, and
// encoding/json/v2 reports an error. Store an octave by its ASPN number, with
// [Octave.ASPN] and [FromASPN].
type Octave struct {
	// aspn is the ASPN number of the octave; it is 0 for the invalid value.
	aspn int
	// valid is false only for the invalid value.
	valid bool
}

// of returns the octave with ASPN number aspn.
func of(aspn int) Octave {
	return Octave{aspn: aspn, valid: true}
}

// FromASPN returns the octave with ASPN number aspn. Every int, including
// math.MinInt and math.MaxInt, is the ASPN number of an octave, so FromASPN
// never fails or panics, and the result is always valid.
//
// FromASPN(aspn) is the same octave as MustFromNumber(aspn, number.ASPN()).
func FromASPN(aspn int) Octave {
	return of(aspn)
}

// FromNumber returns the octave that has number num in numbering n, built-in
// or own: the octave with ASPN number n.ToASPN(num).
//
// On error, FromNumber returns the zero Octave and an error. The conditions
// are checked in this order:
//
//  1. n is undefined: the error matches [number.ErrUndefined]; ASPN is not
//     used instead;
//  2. num is skipped in n: the error matches [number.ErrSkipped]; no nearby
//     octave is used instead;
//  3. the ASPN number of the octave does not fit in an int: the error matches
//     both [ErrOutOfRange] and [number.ErrOutOfRange].
//
// With n equal to number.ASPN() there is no error; use [FromASPN] instead.
// A number that [number.Convert] converts without an error to another
// numbering may still be out of range here: for example,
// number.Convert(math.MaxInt, number.MiddleC3(), number.FrancoBelgian())
// returns math.MaxInt, while FromNumber(math.MaxInt, number.MiddleC3())
// returns an error that matches ErrOutOfRange.
func FromNumber(num int, n number.Numbering) (Octave, error) {
	aspn, err := n.ToASPN(num)
	if err != nil {
		return Octave{}, numberError(fmt.Sprintf("octave with number %d", num), err)
	}

	return of(aspn), nil
}

// MustFromNumber is like [FromNumber] but panics if FromNumber returns an
// error. The panic value is that error.
//
// It is intended for literals and values known to be valid.
func MustFromNumber(num int, n number.Numbering) Octave {
	return must(FromNumber(num, n))
}

// FromName returns the octave that has the given name in catalog c: the
// octave with ASPN number c.ASPN(name). Names are compared exactly, by the
// rules of package [github.com/go-muse/musex/octave/naming].
//
// If c has no octave with this name, including for the empty name, a name
// of another catalog, and a nil or empty catalog, FromName returns the zero
// Octave and an error that matches [naming.ErrUnknownName]. There are no
// other errors.
func FromName(name string, c *naming.Catalog) (Octave, error) {
	aspn, err := c.ASPN(name)
	if err != nil {
		return Octave{}, fmt.Errorf("octave: octave by name: %w", err)
	}

	return of(aspn), nil
}

// MustFromName is like [FromName] but panics if FromName returns an error.
// The panic value is that error.
//
// It is intended for literals and values known to be valid.
func MustFromName(name string, c *naming.Catalog) Octave {
	return must(FromName(name, c))
}

// Number returns the number of o in numbering n, built-in or own:
// n.FromASPN applied to the ASPN number of o. The result is never a skipped
// number of n, and FromNumber(o.MustNumber(n), n) is o.
//
// On error, Number returns 0 and an error; ignore the 0. The conditions are
// checked in this order:
//
//  1. o is invalid: the error matches [ErrInvalid], even if n is undefined;
//  2. n is undefined: the error matches [number.ErrUndefined];
//  3. the number does not fit in an int: the error matches both
//     [ErrOutOfRange] and [number.ErrOutOfRange].
func (o Octave) Number(n number.Numbering) (int, error) {
	if !o.valid {
		return 0, ErrInvalid
	}

	num, err := n.FromASPN(o.aspn)
	if err != nil {
		return 0, numberError(fmt.Sprintf("number of %v", o), err)
	}

	return num, nil
}

// MustNumber is like [Octave.Number] but panics if Number returns an error.
// The panic value is that error.
//
// It is intended for literals and values known to be valid.
func (o Octave) MustNumber(n number.Numbering) int {
	return must(o.Number(n))
}

// ASPN returns the ASPN number of o, the same as o.Number(number.ASPN()).
// The ASPN number is a visible number, like a number in any other numbering.
// Every octave has one, so the only error is for the invalid value: ASPN
// returns 0 and an error that matches [ErrInvalid]; the 0 is not an answer.
func (o Octave) ASPN() (int, error) {
	if !o.valid {
		return 0, ErrInvalid
	}

	return o.aspn, nil
}

// MustASPN is like [Octave.ASPN] but panics if ASPN returns an error. The
// panic value is that error, which matches [ErrInvalid].
//
// It is intended for values known to be valid.
func (o Octave) MustASPN() int {
	return must(o.ASPN())
}

// Name returns the name of o in catalog c, or "" if c has no name for o.
// Name never fails or panics, for any octave and any catalog, including a nil
// or zero catalog.
//
// For the invalid value, Name returns "": it is never treated as an octave.
// Use [Octave.IsValid] to tell it from an octave without a name.
func (o Octave) Name(c *naming.Catalog) string {
	if !o.valid {
		return ""
	}

	return c.Name(o.aspn)
}

// IsValid reports whether o is an octave. It returns false only for the
// invalid value, the zero Octave, which operations also return together with
// an error.
func (o Octave) IsValid() bool {
	return o.valid
}

// Equal reports whether o and other are both valid and are the same octave,
// however they were obtained: FromNumber(3, number.FrancoBelgian()),
// FromASPN(4), FromName("первая", naming.Russian()), and FromASPN(3) shifted
// up by 1 are equal.
//
// The invalid value is not equal to anything, including itself, although
// Octave{} == Octave{} is true. For octaves, Equal is reflexive, symmetric,
// and transitive, and it gives the same answer as ==.
func (o Octave) Equal(other Octave) bool {
	return o.valid && other.valid && o.aspn == other.aspn
}

// Compare returns -1 if o is lower than other, 0 if they are the same octave,
// and +1 if o is higher. It never fails, including for the lowest and the
// highest representable octaves, and gives a total order suitable for
// slices.SortFunc(xs, Octave.Compare) and slices.BinarySearchFunc. The order
// agrees with the order of numbers in every numbering.
//
// The invalid value is lower than every octave, and two invalid values
// compare as 0, although they are not equal by [Octave.Equal].
func (o Octave) Compare(other Octave) int {
	switch {
	case !o.valid && !other.valid:
		return 0
	case !o.valid:
		return -1
	case !other.valid:
		return +1
	case o.aspn < other.aspn:
		return -1
	case o.aspn > other.aspn:
		return +1
	default:
		return 0
	}
}

// Above reports whether o is higher than other. It returns false if either
// of them is the invalid value.
func (o Octave) Above(other Octave) bool {
	return o.valid && other.valid && o.aspn > other.aspn
}

// Below reports whether o is lower than other. It returns false if either of
// them is the invalid value.
func (o Octave) Below(other Octave) bool {
	return o.valid && other.valid && o.aspn < other.aspn
}

// Between reports whether o lies between a and b, inclusive: o is not lower
// than the lower of a and b and not higher than the higher of them. The order
// of a and b does not matter, and a equal to b is a range of one octave.
// Between returns false if any of o, a, and b is the invalid value.
func (o Octave) Between(a, b Octave) bool {
	if !o.valid || !a.valid || !b.valid {
		return false
	}

	lo, hi := min(a.aspn, b.aspn), max(a.aspn, b.aspn)

	return lo <= o.aspn && o.aspn <= hi
}

// String returns a fixed form of o for logs, debugging, and tests: "ASPN"
// and the ASPN number of the octave, such as "ASPN 4" or "ASPN -1", and
// "Octave(invalid)" for the invalid value. The number is decimal, with an
// ASCII minus sign.
//
// The form is not a notation such as "C4", has no names, is not localized,
// and has no parser. One octave always has one form, however it was
// obtained. String implements [fmt.Stringer], which fmt uses for %v, %+v, %s,
// %q, %x, %X, and output without a verb, such as [fmt.Println].
func (o Octave) String() string {
	if !o.valid {
		return "Octave(invalid)"
	}

	return "ASPN " + strconv.Itoa(o.aspn)
}

// GoString returns a Go expression that creates a value equal to o through
// the public API: "octave.FromASPN(4)" for an octave, and "octave.Octave{}"
// for the invalid value. It implements [fmt.GoStringer], which fmt uses for
// %#v.
func (o Octave) GoString() string {
	if !o.valid {
		return "octave.Octave{}"
	}

	return "octave.FromASPN(" + strconv.Itoa(o.aspn) + ")"
}

// numberError wraps an error of package number with context. An error that
// matches number.ErrOutOfRange also matches ErrOutOfRange, so that a single
// check for ErrOutOfRange covers every case of this package.
func numberError(context string, err error) error {
	if errors.Is(err, number.ErrOutOfRange) {
		return fmt.Errorf("%w: %s: %w", ErrOutOfRange, context, err)
	}

	return fmt.Errorf("octave: %s: %w", context, err)
}

// outOfRange returns an error that matches ErrOutOfRange with context.
func outOfRange(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrOutOfRange, fmt.Sprintf(format, args...))
}

// must returns v if err is nil and panics with err otherwise.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
