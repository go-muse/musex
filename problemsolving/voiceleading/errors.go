package voiceleading

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/duration"
)

// Errors distinguish invalid inputs from a successful absence of findings.
var (
	ErrUnset            = errors.New("voiceleading: unset value")
	ErrPartCount        = errors.New("voiceleading: at least two parts required")
	ErrPart             = errors.New("voiceleading: invalid part or pair")
	ErrOption           = errors.New("voiceleading: invalid option")
	ErrNoOrder          = errors.New("voiceleading: nominal order required")
	ErrNoSystem         = errors.New("voiceleading: no system for a note")
	ErrDifferentSystems = errors.New("voiceleading: different systems")
	ErrTransition       = errors.New("voiceleading: positions are not adjacent starts")
)

// PartError identifies an input slot and original element. Element -1 denotes
// an error in the line itself. Use errors.Is and errors.AsType for inspection.
type PartError struct {
	Part, Element int
	Err           error
}

// Error returns a nonempty description, also for a nil receiver or cause.
func (e *PartError) Error() string {
	if e == nil {
		return "voiceleading: part error"
	}
	return fmt.Sprintf("voiceleading: part %d element %d: %v", e.Part, e.Element, e.Err)
}

// Unwrap returns the cause, or nil for a nil receiver.
func (e *PartError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// MeasurementError identifies a pair and the start of the failed vertical;
// for melodic movement At is the target vertical's start.
type MeasurementError struct {
	Pair Pair
	At   duration.Position
	Err  error
}

// Error returns a nonempty description, also for a nil receiver or cause.
func (e *MeasurementError) Error() string {
	if e == nil {
		return "voiceleading: measurement error"
	}
	return fmt.Sprintf("voiceleading: pair (%d,%d) at %s: %v", e.Pair.First, e.Pair.Second, e.At, e.Err)
}

// Unwrap returns the cause, or nil for a nil receiver.
func (e *MeasurementError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func unset(what string, cause error) error {
	return fmt.Errorf("voiceleading: %s: %w", what, errors.Join(ErrUnset, cause))
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func must2[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}
	return a, b
}
