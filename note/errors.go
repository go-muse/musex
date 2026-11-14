package note

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
)

// Errors reported by the package. Test them with errors.Is; the texts of
// errors are not part of the API.
//
// An error of the package matches at most one of these sentinels; the
// sentinels never wrap each other. When a failure is caused by a dependency,
// the cause of the dependency is kept: errors.Is matches both the sentinel of
// note and the error of the dependency. A failure that has no meaning of its
// own for a note, such as a parse error or a spelling rule that gives no name,
// is returned as the error of the dependency with the operation added, without
// a sentinel of note.
//
// When an operation has several problems, it reports the first one in this
// order: an unset input (ErrUnset), a note without an octave (ErrNoOctave),
// no source (ErrNoSource), unset settings (ErrUnsetSettings), an inadmissible
// input name (ErrInadmissible), different systems (ErrDifferentSystems), lost
// expressibility (ErrInexpressible), then the errors of the computation in the
// order of the contracts of the dependencies. The name of a result is checked
// after it is formed: an alteration of the result out of range comes before
// its inadmissibility. The inputs are checked in this order: the receiver, the
// note arguments, then the other arguments in the order of the parameters.
var (
	// ErrUnset reports an unset value: an unset note (the zero Note), a nil
	// *Note at a changing method, an unset name, pitch, octave, spelling
	// rule, rounding mode or criterion, a nil explicit source or an unset
	// explicit tuning. An error for an unset frequency criterion also matches
	// frequency.ErrUnset.
	ErrUnset = errors.New("note: unset value")

	// ErrNoSource reports an unbound note in an operation that needs the
	// settings of the note when no explicit source is given.
	ErrNoSource = errors.New("note: no settings source")

	// ErrUnsetSettings reports a source that returned unset settings. An error
	// that matches it also matches settings.ErrUnset.
	ErrUnsetSettings = errors.New("note: source returned unset settings")

	// ErrNoOctave reports a note without an octave in an operation that needs
	// a written pitch: a frequency, a step, a written octave, a position.
	ErrNoOctave = errors.New("note: note has no octave")

	// ErrInadmissible reports a name that is not admissible in the system of
	// the settings or of the explicit tuning. An error that matches it also
	// matches system.ErrInadmissible.
	ErrInadmissible = errors.New("note: name not admissible")

	// ErrDifferentSystems reports notes whose settings have different systems
	// in an operation that compares positions.
	ErrDifferentSystems = errors.New("note: different systems")

	// ErrInexpressible reports a change of the source that would lose
	// expressibility: the note is expressible by its settings and would not be
	// expressible by the new ones.
	ErrInexpressible = errors.New("note: expressibility lost")

	// ErrOutOfRange reports a result out of the range of a dependency: an
	// alteration, an octave, a written octave, a visible number, a step, cents
	// or a frequency. The error of the dependency is kept.
	ErrOutOfRange = errors.New("note: result out of range")
)

// unsetError returns an error that matches ErrUnset.
func unsetError(op, what string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnset, op, what)
}

// unsetCriterionError returns an error that matches ErrUnset and
// frequency.ErrUnset.
func unsetCriterionError(op string) error {
	return fmt.Errorf("%w: %s: criterion: %w", ErrUnset, op, frequency.ErrUnset)
}

// noSourceError returns an error that matches ErrNoSource.
func noSourceError(op string) error {
	return fmt.Errorf("%w: %s", ErrNoSource, op)
}

// unsetSettingsError returns an error that matches ErrUnsetSettings and
// settings.ErrUnset.
func unsetSettingsError(op string) error {
	return fmt.Errorf("%w: %s: %w", ErrUnsetSettings, op, settings.ErrUnset)
}

// noOctaveError returns an error that matches ErrNoOctave.
func noOctaveError(op string, name spelling.Spelling) error {
	return fmt.Errorf("%w: %s: %v", ErrNoOctave, op, name)
}

// inadmissibleError returns an error that matches ErrInadmissible and
// system.ErrInadmissible.
func inadmissibleError(op string, name spelling.Spelling, sys *system.System) error {
	return fmt.Errorf("%w: %s: %v in system %q: %w",
		ErrInadmissible, op, name, sys.Name(), system.ErrInadmissible)
}

// differentSystemsError returns an error that matches ErrDifferentSystems.
func differentSystemsError(op string, a, b *system.System) error {
	return fmt.Errorf("%w: %s: %q and %q", ErrDifferentSystems, op, a.Name(), b.Name())
}

// inexpressibleError returns an error that matches ErrInexpressible.
func inexpressibleError(op string) error {
	return fmt.Errorf("%w: %s: expressible by the previous settings only", ErrInexpressible, op)
}

// invalidFrequencyError returns an error that matches
// frequency.ErrInvalidFrequency and cents.ErrInvalidArgument.
func invalidFrequencyError(op string, f float64) error {
	return fmt.Errorf("note: %s: %v: %w: %w", op, f, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
}

// outOfRangeCauses are the errors of the dependencies that mean a result out
// of range or too complex to be represented.
var outOfRangeCauses = [...]error{
	alteration.ErrOutOfRange,
	octave.ErrOutOfRange,
	system.ErrOutOfRange,
	tuning.ErrOutOfRange,
	tuning.ErrTooComplex,
	cents.ErrOutOfRange,
	cents.ErrTooComplex,
}

// isOutOfRange reports whether err of a dependency means a result out of
// range.
func isOutOfRange(err error) bool {
	for _, cause := range outOfRangeCauses {
		if errors.Is(err, cause) {
			return true
		}
	}

	return false
}

// depError classifies an error of a dependency of a computation: a result
// out of range matches ErrOutOfRange together with the cause; every other
// error is returned with the operation added and without a sentinel of note.
func depError(op string, err error) error {
	if isOutOfRange(err) {
		return fmt.Errorf("%w: %s: %w", ErrOutOfRange, op, err)
	}

	return fmt.Errorf("note: %s: %w", op, err)
}

// formatError classifies an error of settings.Format or settings.FormatPitch:
// an inadmissible name matches ErrInadmissible, a result out of range matches
// ErrOutOfRange; every other error is returned without a sentinel of note.
func formatError(op string, err error) error {
	if errors.Is(err, settings.ErrInadmissible) {
		return fmt.Errorf("%w: %s: %w", ErrInadmissible, op, err)
	}

	return depError(op, err)
}

// parseError classifies an error of settings.Parse or settings.ParsePitch:
// only an inadmissible name gets a sentinel of note (ErrInadmissible); an
// error out of range stays a cause of settings.
func parseError(op string, err error) error {
	if errors.Is(err, settings.ErrInadmissible) {
		return fmt.Errorf("%w: %s: %w", ErrInadmissible, op, err)
	}

	return fmt.Errorf("note: %s: %w", op, err)
}
