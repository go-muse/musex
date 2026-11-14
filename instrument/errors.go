package instrument

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
)

// Errors reported by the package. Test them with errors.Is; the texts of
// errors are not part of the API.
//
// Every error with a cause of this package matches exactly one of these
// values. Two of them also match a cause of a dependency: an error that matches
// ErrInadmissible also matches system.ErrInadmissible, and an error that
// matches ErrIncompatible also matches settings.ErrIncompatible. The values
// never wrap each other, and ErrUnset matches neither settings.ErrUnset nor
// system.ErrInvalid.
var (
	// ErrUnset reports an unset input: an instrument, a preset, a set of
	// settings, a system, a transformation (RetuneFunc) or a preset given to
	// a catalog. In the Retune methods for one part of the set it reports
	// only an unset instrument; an unset part is reported by package settings.
	ErrUnset = errors.New("instrument: unset value")

	// ErrEmptyName reports an empty name of a preset or of an instrument.
	ErrEmptyName = errors.New("instrument: empty name")

	// ErrUnsetPitch reports an unset pitch: a bound given to NewRange,
	// a pitch to check or an element given to RangeOf.
	ErrUnsetPitch = errors.New("instrument: unset pitch")

	// ErrInadmissible reports a pitch whose name is not admissible in the
	// system: a bound of a range or a pitch to check. The error names the
	// bound or the pitch. An error that matches it also matches
	// system.ErrInadmissible.
	ErrInadmissible = errors.New("instrument: pitch not admissible")

	// ErrInverted reports a range whose lower bound is above its upper bound
	// by position in the system.
	ErrInverted = errors.New("instrument: inverted range")

	// ErrIncompatible reports a next set of settings that cannot replace the
	// current set of an instrument by the rules of package settings. The
	// error is an *IncompatibleError with the report of the check; it also
	// matches settings.ErrIncompatible.
	ErrIncompatible = errors.New("instrument: incompatible settings")

	// ErrNoRange reports a missing range: the unset Range given to an
	// operation of a range, or an operation of the range of an instrument
	// that has no range.
	ErrNoRange = errors.New("instrument: no range")

	// ErrDuplicateName reports a name that occurs twice in a catalog.
	ErrDuplicateName = errors.New("instrument: duplicate preset name")

	// ErrNotFound reports a name that no preset of a catalog has.
	ErrNotFound = errors.New("instrument: preset not found")

	// ErrNoPitches reports RangeOf called without pitches.
	ErrNoPitches = errors.New("instrument: no pitches")
)

// IncompatibleError is the error of a refused retuning: the next set of
// settings is incompatible with the current one (ErrIncompatible).
//
// Report is the report of settings.Settings.CheckReplacement with the violated
// rules and their counterexamples. The package creates IncompatibleError only
// with an incompatible report. Get it with
//
//	ie, ok := errors.AsType[*instrument.IncompatibleError](err)
//
// The nil pointer and the zero value do not panic: Error returns a text
// without a report and Unwrap returns only ErrIncompatible.
type IncompatibleError struct {
	// Report is the settings compatibility report; never compatible.
	Report settings.Report
}

// Error returns the text of the error with the report, when there is one.
func (e *IncompatibleError) Error() string {
	if e == nil || e.Report == (settings.Report{}) {
		return ErrIncompatible.Error()
	}

	return ErrIncompatible.Error() + ": " + e.Report.String()
}

// Unwrap returns ErrIncompatible and the error of the report (Report.Err),
// which matches settings.ErrIncompatible. For the nil pointer and for the
// zero value, which checked no rule, it returns only ErrIncompatible.
func (e *IncompatibleError) Unwrap() []error {
	if e == nil || e.Report == (settings.Report{}) {
		return []error{ErrIncompatible}
	}

	if err := e.Report.Err(); err != nil {
		return []error{ErrIncompatible, err}
	}

	return []error{ErrIncompatible}
}

// unsetError returns an error that matches ErrUnset.
func unsetError(op, what string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnset, op, what)
}

// emptyNameError returns an error that matches ErrEmptyName.
func emptyNameError(op string) error {
	return fmt.Errorf("%w: %s", ErrEmptyName, op)
}

// unsetPitchError returns an error that matches ErrUnsetPitch.
func unsetPitchError(op, what string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnsetPitch, op, what)
}

// inadmissibleError returns an error that matches both ErrInadmissible and
// system.ErrInadmissible: the name of p, the bound or the pitch named by
// what, is not admissible in sys.
func inadmissibleError(op, what string, p pitch.Pitch, sys *system.System) error {
	return fmt.Errorf("%w: %s: %s %v in system %q: %w",
		ErrInadmissible, op, what, p, sys.Name(), system.ErrInadmissible)
}

// invertedError returns an error that matches ErrInverted.
func invertedError(op string, r Range, sys *system.System) error {
	return fmt.Errorf("%w: %s: %v in system %q", ErrInverted, op, r, sys.Name())
}

// noRangeError returns an error that matches ErrNoRange.
func noRangeError(op string) error {
	return fmt.Errorf("%w: %s", ErrNoRange, op)
}

// duplicateNameError returns an error that matches ErrDuplicateName.
func duplicateNameError(op, name string) error {
	return fmt.Errorf("%w: %s: %q", ErrDuplicateName, op, name)
}

// notFoundError returns an error that matches ErrNotFound.
func notFoundError(op, name string) error {
	return fmt.Errorf("%w: %s: %q", ErrNotFound, op, name)
}

// noPitchesError returns an error that matches ErrNoPitches.
func noPitchesError(op string) error {
	return fmt.Errorf("%w: %s", ErrNoPitches, op)
}

// wrapError adds the operation to an error of a dependency or of a
// transformation, keeping its cause available through errors.Is.
func wrapError(op string, err error) error {
	return fmt.Errorf("instrument: %s: %w", op, err)
}
