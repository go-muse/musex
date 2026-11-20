package voice

import (
	"errors"
	"fmt"
)

// Sentinel errors of the package. Every error returned by the package matches
// at most one of them with errors.Is, and the sentinels never wrap each
// other. Causes of dependencies are kept: an error for an unset value of a
// dependency also matches the sentinel of that dependency, such as
// note.ErrUnset, and failures without a meaning of their own for a line are
// returned with the cause of the dependency and no sentinel of this package.
// The message text is not part of the API.
var (
	// ErrUnset reports an unset value: a nil or zero *Line receiver, an unset
	// Element, Duration, Reference or Zones, an unset note, position,
	// interval, axis, explicit system, tempo map, preset, instrument, event,
	// zone bound or duration of a dependency, a value of an enumeration
	// outside its list, WeightingUnset, RetrogradeUnset, or an unset line in
	// Join. For an unset value of a dependency it also matches the sentinel
	// of that dependency.
	ErrUnset = errors.New("voice: unset value")

	// ErrIndex reports an element index outside its bounds, i > j for a
	// half-open range of indexes, or a > b for Span.
	ErrIndex = errors.New("voice: index out of range")

	// ErrNoOctave reports a note without an octave for a sound, for the axis
	// of an inversion, or in an event.
	ErrNoOctave = errors.New("voice: note without octave")

	// ErrChord reports an event with a chord, even a chord of one note: a line
	// sounds at most one pitch at a time.
	ErrChord = errors.New("voice: chord in a line")

	// ErrMeasureDuration reports an event with a duration in measures: a line
	// has no measures.
	ErrMeasureDuration = errors.New("voice: duration in measures")

	// ErrTiedRest reports a tied duration on a rest.
	ErrTiedRest = errors.New("voice: tied duration on a rest")

	// ErrGrace reports a grace event: a grace note takes no time and is not an
	// element of a line.
	ErrGrace = errors.New("voice: grace note in a line")

	// ErrNoSounds reports a range, extremes, ambitus or tessitura of a line
	// without sounds.
	ErrNoSounds = errors.New("voice: line has no sounds")

	// ErrNoRange reports a missing reference range: the unset instrument
	// range, a preset or an instrument without a range, or a placement
	// without a reference range.
	ErrNoRange = errors.New("voice: no reference range")

	// ErrZeroWidth reports a reference range of width zero in a tessitura.
	ErrZeroWidth = errors.New("voice: reference range has zero width")

	// ErrZones reports tessitura zones that do not satisfy
	// 0 <= low <= high <= 1.
	ErrZones = errors.New("voice: invalid tessitura zones")

	// ErrUnknownName reports a text that is not a name of a role of the
	// requested kind.
	ErrUnknownName = errors.New("voice: unknown role name")

	// ErrBrokenTie reports an event tied to the next one when the next event
	// is missing, unset, not a note, or a note with another written record.
	ErrBrokenTie = errors.New("voice: tie without continuation")

	// ErrTiedValue reports an event of a tied series with a duration that has
	// no notation.
	ErrTiedValue = errors.New("voice: tie on a duration without notation")

	// ErrFermataInTie reports a fermata on an event tied to the next one.
	ErrFermataInTie = errors.New("voice: fermata inside a tie")

	// ErrNoSystem reports an unbound note where the system of an operation is
	// taken from the settings of the notes. It does not match
	// interval.ErrNoSystem.
	ErrNoSystem = errors.New("voice: no system for a note")

	// ErrDifferentSystems reports sounds with different systems, or a bound
	// sound whose system differs from the system of the preset or the
	// instrument of the reference range. It does not match
	// interval.ErrDifferentSystems.
	ErrDifferentSystems = errors.New("voice: different systems")

	// ErrMeasureMap reports a tempo map addressed by measures in an operation
	// of seconds: a line has no measures.
	ErrMeasureMap = errors.New("voice: tempo map addressed by measures")
)

// ItemError is a failure caused by one item: an element of the input list of
// New, Append or Insert, a pair of FromPairs, an event of FromEvents, a line
// of Join, or an element of the receiver line in operations of pitch,
// building and seconds. Index is the zero-based index of that item and Err is
// the whole cause.
//
// ItemError unwraps to Err, so errors.Is sees through it. Get the index with
// errors.AsType[*voice.ItemError](err).
type ItemError struct {
	Index int   // zero-based index of the input item or of the line element
	Err   error // the cause
}

// Error returns a text with the index and the cause. It works for a nil
// pointer and for a nil cause.
func (e *ItemError) Error() string {
	if e == nil {
		return "voice: item error"
	}

	if e.Err == nil {
		return fmt.Sprintf("voice: item %d", e.Index)
	}

	return fmt.Sprintf("voice: item %d: %v", e.Index, e.Err)
}

// Unwrap returns the cause; nil for a nil pointer.
func (e *ItemError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

// itemError wraps err in an *ItemError with the index i.
func itemError(i int, err error) error {
	return &ItemError{Index: i, Err: err}
}

// unsetError returns an error matching ErrUnset for the value what and, when
// cause is not nil, the sentinel of the dependency that owns the value.
func unsetError(what string, cause error) error {
	if cause == nil {
		return fmt.Errorf("%w: %s", ErrUnset, what)
	}

	return fmt.Errorf("%w: %s: %w", ErrUnset, what, cause)
}

// errUnsetLine is the failure of an operation on an unset line.
func errUnsetLine() error {
	return unsetError("line", nil)
}

// indexError returns an error matching ErrIndex with a formatted detail.
func indexError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrIndex, fmt.Sprintf(format, args...))
}

// must returns v and panics with err when err is not nil.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

// must2 returns a and b and panics with err when err is not nil.
func must2[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}

	return a, b
}
