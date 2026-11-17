package mode

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/degree"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/template"
)

// Errors identify the categories returned by mode operations. Use errors.Is
// to match a category or an underlying package error.
var (
	ErrUnset                = errors.New("mode: unset value")
	ErrInvalidArgument      = errors.New("mode: invalid argument")
	ErrEmptyInput           = errors.New("mode: empty input")
	ErrInvalidStructure     = errors.New("mode: invalid structure")
	ErrSystemOctaveRequired = errors.New("mode: system octave required")
	ErrWrongDegreeCount     = errors.New("mode: wrong degree count")
	ErrIncomparable         = errors.New("mode: incomparable modes")
	ErrNonPositiveNumber    = errors.New("mode: non-positive number")
	ErrNoEntry              = errors.New("mode: catalog entry not found")
	ErrAmbiguousName        = errors.New("mode: ambiguous name")
	ErrOutOfRange           = errors.New("mode: result out of range")
	ErrParse                = errors.New("mode: cannot parse")
	ErrNoSystem             = errors.New("mode: no system for a note")
	ErrCannotBuildNote      = errors.New("mode: cannot build note")
	ErrOpenMode             = errors.New("mode: operation undefined for an open mode")
)

func templateError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, template.ErrParse):
		return errors.Join(ErrParse, err)
	case errors.Is(err, template.ErrUnset):
		return errors.Join(ErrUnset, err)
	case errors.Is(err, template.ErrEmptyInput):
		return errors.Join(ErrEmptyInput, err)
	case errors.Is(err, template.ErrInvalidArgument):
		return errors.Join(ErrInvalidArgument, err)
	case errors.Is(err, template.ErrOpenRotation):
		return errors.Join(ErrOpenMode, err)
	case errors.Is(err, template.ErrInvalidInterval):
		return errors.Join(ErrInvalidStructure, err)
	case errors.Is(err, template.ErrIncomparable):
		return errors.Join(ErrIncomparable, err)
	case errors.Is(err, template.ErrOutOfRange):
		return errors.Join(ErrOutOfRange, err)
	default:
		return err
	}
}

func degreeError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, degree.ErrUnset):
		return errors.Join(ErrUnset, err)
	case errors.Is(err, degree.ErrInvalidNumber):
		return errors.Join(ErrInvalidArgument, err)
	case errors.Is(err, degree.ErrOutOfRange):
		return errors.Join(ErrOutOfRange, err)
	default:
		return err
	}
}

func buildNoteError(op string, err error) error {
	category := ErrCannotBuildNote
	if errors.Is(err, interval.ErrNoSystem) {
		category = ErrNoSystem
	}
	return fmt.Errorf("%s: %w: %w", op, category, err)
}

func positionError(op string, err error) error {
	if errors.Is(err, interval.ErrOutOfRange) {
		return fmt.Errorf("%s: %w: %w", op, ErrOutOfRange, err)
	}
	return fmt.Errorf("%s: %w", op, err)
}
