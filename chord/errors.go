package chord

import (
	"errors"
	"fmt"
)

// Errors reported by the package. Test them with errors.Is; the texts of
// errors are not part of the API.
//
// An error of the package matches at most one of these sentinels, and the
// sentinels never wrap each other. The one exception is ErrRuleFailed: it
// keeps the error of the rule of the user, which may itself match anything,
// even a sentinel of chord.
//
// Shift, ShiftDown, Up and Down report the error of package note for the
// first note, in the order of record, that cannot be moved: note.ErrNoOctave,
// or note.ErrOutOfRange together with octave.ErrOutOfRange. Such an error
// matches no sentinel of chord.
//
// When an operation has several problems, it reports the first one in this
// order. Every kind of problem is checked over all the inputs before the
// next kind:
//
//  1. ErrUnset for an input: the receiver chord (at At and the changing
//     methods), the chord argument, the note arguments in the order of the
//     parameters and of the list, then the rule;
//  2. ErrIndex;
//  3. ErrNotFound;
//  4. the failures of the execution, note by note in the order of record:
//     ErrRuleFailed; ErrUnset for an unset note returned by a Transform
//     (the error of the rule is checked first); the error of package note.
var (
	// ErrUnset reports an unset value: an unset chord (the zero Chord) at At
	// or at a changing method, a nil *Chord at a changing method, an unset
	// chord argument, an unset note given to create or change a chord, a nil
	// rule, or an unset note returned by a Transform. It does not match
	// note.ErrUnset: the chord finds the unset note itself.
	ErrUnset = errors.New("chord: unset value")

	// ErrNotFound reports a note that the chord does not contain: Remove,
	// RemoveAll, Replace and ReplaceAll.
	ErrNotFound = errors.New("chord: note not found")

	// ErrIndex reports a number of a note outside the chord: i < 0 or
	// i >= Len(); for Insert, i < 0 or i > Len().
	ErrIndex = errors.New("chord: index out of range")

	// ErrRuleFailed reports a rule of the user that returned an error.
	// The error of the rule is kept: errors.Is matches both ErrRuleFailed and
	// the error of the rule.
	ErrRuleFailed = errors.New("chord: rule failed")
)

// unsetError returns an error that matches ErrUnset.
func unsetError(op, what string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnset, op, what)
}

// unsetNoteError returns an error that matches ErrUnset for the unset note
// with number i in the list of the notes given to op.
func unsetNoteError(op string, i int) error {
	return fmt.Errorf("%w: %s: note %d of the list", ErrUnset, op, i)
}

// indexError returns an error that matches ErrIndex.
func indexError(op string, i, n int) error {
	return fmt.Errorf("%w: %s: index %d, %d notes", ErrIndex, op, i, n)
}

// notFoundError returns an error that matches ErrNotFound.
func notFoundError(op string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, op)
}

// ruleError returns an error that matches ErrRuleFailed and the error of the
// rule. A negative i means that the failure belongs to no single note.
func ruleError(op string, i int, err error) error {
	if i < 0 {
		return fmt.Errorf("%w: %s: %w", ErrRuleFailed, op, err)
	}

	return fmt.Errorf("%w: %s: note %d: %w", ErrRuleFailed, op, i, err)
}

// noteError returns the error of package note for the note with number i,
// with the operation added and without a sentinel of chord.
func noteError(op string, i int, err error) error {
	return fmt.Errorf("chord: %s: note %d: %w", op, i, err)
}
