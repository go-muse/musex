package hindemith

import (
	"errors"
	"fmt"
)

// Sentinel errors of the package. Every returned error matches at most one
// of them with errors.Is; a failure caused by a dependency also wraps the
// error of that dependency.
var (
	// ErrUnset indicates an unset input: a nil or unset system or an unset
	// note.
	ErrUnset = errors.New("hindemith: unset value")

	// ErrNoOctave indicates a note without an octave.
	ErrNoOctave = errors.New("hindemith: note without octave")

	// ErrUnsupportedSystem indicates a system that is not twelve classes
	// with chromatic unit 1: the model is stated for twelve-tone equal
	// temperament.
	ErrUnsupportedSystem = errors.New("hindemith: system is not 12-tone equal-tempered")

	// ErrTooFewNotes indicates a chord of fewer than two notes.
	ErrTooFewNotes = errors.New("hindemith: chord needs at least two notes")

	// ErrInadmissible indicates a note name that the system does not admit,
	// wrapping system.ErrInadmissible.
	ErrInadmissible = errors.New("hindemith: inadmissible note name")

	// ErrOutOfRange indicates a distance that does not fit in int, or a
	// distance of exactly math.MinInt steps whose magnitude does not fit,
	// wrapping system.ErrOutOfRange.
	ErrOutOfRange = errors.New("hindemith: result out of range")
)

func opError(op string, sentinel error) error {
	return fmt.Errorf("hindemith: %s: %w", op, sentinel)
}

func rangeError(op string, cause error) error {
	return fmt.Errorf("hindemith: %s: %w: %w", op, ErrOutOfRange, cause)
}

func inadmissibleError(op string, cause error) error {
	return fmt.Errorf("hindemith: %s: %w: %w", op, ErrInadmissible, cause)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
