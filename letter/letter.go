package letter

import (
	"errors"
	"fmt"
	"strconv"
)

// Letter is a letter of the seven-letter set C D E F G A B.
//
// The valid values are exactly the seven constants [C], [D], [E], [F], [G],
// [A], and [B]. The zero value and any other value of the type are invalid:
// they are not letters, and the zero value is not C.
//
// The numeric values of the constants (C = 1 through B = 7) are stable
// because reflection-based encoders such as encoding/json and fmt with %d
// output them. They are an internal representation, not ordinals (C has
// ordinal 0), and they are not a recommended storage format: use
// [Letter.Ordinal] with [FromOrdinal], or [Letter.String], for storage and
// exchange.
//
// A conversion such as Letter(n) is not checked, and arithmetic on letters
// does not wrap around the circle: B + 1 is an invalid value. Use
// [Letter.Shift] to move along the circle and [FromOrdinal] to create a
// letter from a number taken from data.
//
// Letter is a comparable value type: it can be copied freely and used as a
// map key or in a switch. For valid letters, == and [Letter.Equal] agree.
type Letter uint8

// The letters of the seven-letter set in canonical order.
//
// Their numeric values are stable but are not ordinals; see [Letter].
const (
	C Letter = 1 + iota
	D
	E
	F
	G
	A
	B
)

// Count is the number of letters in the seven-letter set.
//
// It is an untyped constant, so it can be used in arithmetic with any integer
// type and as an array length, for example [Count]string for a table indexed
// by [Letter.Ordinal]. It equals len(All()) and len(l.Cycle()) for any valid
// letter l.
const Count = 7

// ErrInvalid reports that a value is not a letter of the seven-letter set.
//
// [FromOrdinal] returns errors that wrap ErrInvalid. Check for it with
// [errors.Is]; the text of the error is not part of the API.
var ErrInvalid = errors.New("letter: invalid letter")

// symbols holds the fixed string forms of the letters, indexed by ordinal.
// It is read-only.
var symbols = [Count]string{"C", "D", "E", "F", "G", "A", "B"}

// All returns the Count letters of the seven-letter set in canonical order:
// C D E F G A B.
//
// Each call returns a new slice with length and capacity Count. The caller
// owns it and may modify or retain it; this does not affect later calls.
func All() []Letter {
	return []Letter{C, D, E, F, G, A, B}
}

// FromOrdinal returns the letter with the given ordinal in the canonical order
// of the seven-letter set (0 for C through 6 for B).
//
// If ordinal is outside the range 0 to 6, FromOrdinal returns the zero
// Letter and an error that wraps [ErrInvalid]. It does not wrap the ordinal
// around the circle; use C.Shift(n) for that.
func FromOrdinal(ordinal int) (Letter, error) {
	if ordinal < 0 || ordinal >= Count {
		return 0, fmt.Errorf("%w: ordinal %d out of range [0, %d]", ErrInvalid, ordinal, Count-1)
	}

	return C + Letter(ordinal), nil
}

// MustFromOrdinal is like [FromOrdinal] but panics if ordinal is outside the
// range 0 to 6. The panic value is the error returned by FromOrdinal, which
// wraps [ErrInvalid].
//
// It is intended for literals and values known to be correct.
func MustFromOrdinal(ordinal int) Letter {
	l, err := FromOrdinal(ordinal)
	if err != nil {
		panic(err)
	}

	return l
}

// IsValid reports whether l is one of the seven letters C D E F G A B.
// It returns false for the zero value and for any other value.
func (l Letter) IsValid() bool {
	return l >= C && l <= B
}

// Equal reports whether l and other are both valid and the same letter,
// regardless of how they were obtained.
//
// For valid letters, l.Equal(other) equals l == other. If l or other is
// invalid, Equal returns false, even when both hold the same invalid value.
func (l Letter) Equal(other Letter) bool {
	return l.IsValid() && l == other
}

// Ordinal returns the place of l in the canonical order of the seven-letter
// set, from 0 for C to 6 for B. It equals C.DistanceUp(l).
//
// For an invalid l, Ordinal returns -1.
func (l Letter) Ordinal() int {
	if !l.IsValid() {
		return -1
	}

	return int(l - C)
}

// DistanceUp returns the number of letter steps upward around the circle of
// the seven-letter set from l to to, in the range 0 to 6. The distance from a
// letter to itself is 0.
//
// The result is a count of letter steps, not the size of an interval.
// If l or to is invalid, DistanceUp returns -1.
func (l Letter) DistanceUp(to Letter) int {
	if !l.IsValid() || !to.IsValid() {
		return -1
	}

	return wrap(to.Ordinal() - l.Ordinal())
}

// DistanceDown returns the number of letter steps downward around the circle
// of the seven-letter set from l to to, in the range 0 to 6. It equals
// to.DistanceUp(l); for different letters, DistanceUp and DistanceDown add
// up to Count.
//
// If l or to is invalid, DistanceDown returns -1.
func (l Letter) DistanceDown(to Letter) int {
	if !l.IsValid() || !to.IsValid() {
		return -1
	}

	return wrap(l.Ordinal() - to.Ordinal())
}

// Shift returns the letter reached from l by moving the given number of
// letter steps around the circle of the seven-letter set: upward for positive
// steps and downward for negative steps. The shift has period 7, so
// l.Shift(n+7) equals l.Shift(n).
//
// Any int value of steps is accepted, including [math.MinInt] and
// [math.MaxInt], without overflow.
//
// For an invalid l, Shift returns the zero Letter, which is invalid.
func (l Letter) Shift(steps int) Letter {
	if !l.IsValid() {
		return 0
	}

	// Reduce steps first so that the sum stays within -6 to 12.
	return C + Letter(wrap(l.Ordinal()+steps%Count))
}

// Cycle returns the Count letters of the seven-letter set going upward around
// the circle, starting with l. The element at index i equals l.Shift(i), and
// C.Cycle() equals [All]().
//
// Each call returns a new slice with length and capacity Count, owned by the
// caller. For an invalid l, Cycle returns nil.
func (l Letter) Cycle() []Letter {
	if !l.IsValid() {
		return nil
	}

	letters := make([]Letter, Count)
	for i := range Count {
		letters[i] = l.Shift(i)
	}

	return letters
}

// String returns the fixed form of l: "C", "D", "E", "F", "G", "A", or "B".
// It is meant for formatting, logs, and messages; it is not a notation,
// does not depend on any settings, and has no parser.
//
// For an invalid l, String returns "Letter(n)", where n is the decimal
// numeric value of l, for example "Letter(0)".
func (l Letter) String() string {
	if !l.IsValid() {
		return "Letter(" + strconv.Itoa(int(l)) + ")"
	}

	return symbols[l.Ordinal()]
}

// wrap reduces n, which must be in the range -Count to 2*Count-1, to the
// range 0 to Count-1.
func wrap(n int) int {
	switch {
	case n < 0:
		return n + Count
	case n >= Count:
		return n - Count
	default:
		return n
	}
}
