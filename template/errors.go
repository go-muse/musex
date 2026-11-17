package template

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/interval"
)

// Sentinel errors returned by package operations. Use [errors.Is] to classify
// errors; their messages are not part of the API. An error caused by a
// dependency also wraps the corresponding dependency error.
var (
	// ErrUnset indicates an unset template, interval, spelling, or entry, a nil
	// predicate, or a nil receiver of UnmarshalText.
	ErrUnset = errors.New("template: unset value")

	// ErrInvalidArgument indicates an invalid place, measure, entry code or
	// name, a negative difference limit, or more than MaxPlaces places.
	ErrInvalidArgument = errors.New("template: invalid argument")

	// ErrEmptyInput indicates an empty sequence, or a single place or spelling
	// where an open template needs at least two places.
	ErrEmptyInput = errors.New("template: no intervals or names")

	// ErrInvalidInterval indicates a descending or perfect-unison adjacent
	// interval, non-increasing place intervals, or an invalid cycle.
	ErrInvalidInterval = errors.New("template: invalid interval sequence")

	// ErrOpenRotation indicates a rotation of an open template from a place
	// other than its first place.
	ErrOpenRotation = errors.New("template: open template has no such rotation")

	// ErrDuplicateCode indicates an entry code that already occurs in a catalog
	// or occurs more than once among entries being added.
	ErrDuplicateCode = errors.New("template: duplicate catalog code")

	// ErrUnknownCode indicates removal of a code that is not in the catalog,
	// including a second removal of the same code in one call.
	ErrUnknownCode = errors.New("template: unknown catalog code")

	// ErrOutOfRange indicates a result that cannot be represented by package
	// interval. The error also wraps the dependency error that gives the cause.
	ErrOutOfRange = errors.New("template: result out of range")

	// ErrParse indicates text that does not follow template notation or
	// contains a part that package interval cannot parse.
	ErrParse = errors.New("template: cannot parse")

	// ErrIncomparable indicates templates of different kinds or lengths in an
	// operation that compares corresponding places.
	ErrIncomparable = errors.New("template: incomparable templates")
)

// operationError is a comparable error value. Besides preserving the operation
// and sentinel, its value form lets a Must wrapper panic with an error equal to
// the independently returned operation error, as required by the Must contract.
type operationError struct {
	op       string
	sentinel error
}

func (e operationError) Error() string {
	return fmt.Sprintf("template: %s: %v", e.op, e.sentinel)
}

func (e operationError) Unwrap() error {
	return e.sentinel
}

// opError returns an error that wraps sentinel and names the operation.
func opError(op string, sentinel error) error {
	return operationError{op: op, sentinel: sentinel}
}

// dependencyError is a comparable error with a package sentinel and one direct
// dependency cause. Its value form preserves equality across a Must call.
type dependencyError struct {
	op       string
	sentinel error
	cause    error
}

func (e dependencyError) Error() string {
	return fmt.Sprintf("template: %s: %v: %v", e.op, e.sentinel, e.cause)
}

func (e dependencyError) Unwrap() []error {
	return []error{e.sentinel, e.cause}
}

// unsetIntervalError returns ErrUnset together with the interval cause.
func unsetIntervalError(op string) error {
	return dependencyError{op: op, sentinel: ErrUnset, cause: interval.ErrUnset}
}

// rangeError returns an error that wraps ErrOutOfRange and its causes.
func rangeError(op string, cause error) error {
	if errors.Is(cause, interval.ErrOutOfRange) {
		return fmt.Errorf("template: %s: %w: %w", op, ErrOutOfRange, cause)
	}
	return fmt.Errorf("template: %s: %w: %w: %w", op, ErrOutOfRange, interval.ErrOutOfRange, cause)
}

// must returns v and panics with err if err is not nil.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
