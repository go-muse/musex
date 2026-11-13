package settings

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// Errors reported by the package. Test them with errors.Is; the texts of
// errors are not part of the API.
var (
	// ErrUnset reports an unset set of settings (the receiver or the next
	// set), or an unset part: a tuning, a notation, a numbering or, for
	// CheckSystems and WithSystem, a system.
	ErrUnset = errors.New("settings: unset settings or part")

	// ErrInconsistent reports parts that do not fit together: the chromatic
	// unit of an ups-and-downs notation differs from the chromatic unit of
	// the system. An error that matches it also matches
	// notation.ErrIncompatibleSystem.
	ErrInconsistent = errors.New("settings: inconsistent parts")

	// ErrUnsetName reports an unset note name or pitch.
	ErrUnsetName = errors.New("settings: unset name or pitch")

	// ErrInadmissible reports a name that is not admissible in the system of
	// the set when a name or a pitch is formatted or parsed. An error that
	// matches it also matches system.ErrInadmissible.
	ErrInadmissible = errors.New("settings: name not admissible")

	// ErrIncompatible is reported only by Report.Err for a report with at
	// least one violated rule.
	ErrIncompatible = errors.New("settings: incompatible replacement")
)

// unsetError returns an error that matches ErrUnset, with the operation and
// the unset value in its text.
func unsetError(op, what string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnset, op, what)
}

// unsetNameError returns an error that matches ErrUnsetName.
func unsetNameError(op, what string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnsetName, op, what)
}

// inconsistentError returns an error that matches both ErrInconsistent and
// notation.ErrIncompatibleSystem: the chromatic unit cu of an ups-and-downs
// notation differs from the chromatic unit of sys.
func inconsistentError(op string, cu int, sys *system.System) error {
	return fmt.Errorf("%w: %s: notation c=%d, system %q c=%d: %w",
		ErrInconsistent, op, cu, sys.Name(), sys.Chromatic(), notation.ErrIncompatibleSystem)
}

// inadmissibleError returns an error that matches both ErrInadmissible and
// system.ErrInadmissible.
func inadmissibleError(op string, name spelling.Spelling, sys *system.System) error {
	return fmt.Errorf("%w: %s: %v in system %q: %w",
		ErrInadmissible, op, name, sys.Name(), system.ErrInadmissible)
}

// wrapError adds the operation to an error of a dependency, keeping the
// cause available through errors.Is.
func wrapError(op string, err error) error {
	return fmt.Errorf("settings: %s: %w", op, err)
}
