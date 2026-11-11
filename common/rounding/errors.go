package rounding

import "errors"

var (
	// ErrUnset reports an unset mode or a missing arbitrary-precision argument.
	ErrUnset = errors.New("rounding: unset value")
	// ErrUnknownMode reports a mode outside the nine defined modes.
	ErrUnknownMode = errors.New("rounding: unknown mode")
	// ErrInvalidArgument reports an argument outside an operation's domain.
	ErrInvalidArgument = errors.New("rounding: invalid argument")
	// ErrOutOfRange reports a result that cannot be represented by its result type.
	ErrOutOfRange = errors.New("rounding: result out of range")
)

func validateMode(m Mode) error {
	if m == 0 {
		return ErrUnset
	}
	if !m.IsValid() {
		return ErrUnknownMode
	}
	return nil
}
