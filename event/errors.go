package event

import "errors"

// Errors can be checked with errors.Is.
var (
	// ErrUnset reports an unset event, including a nil *Event.
	ErrUnset = errors.New("event: unset event")
	// ErrUnsetContent reports unset event content.
	ErrUnsetContent = errors.New("event: unset content")
	// ErrUnsetDuration reports an unset event duration.
	ErrUnsetDuration = errors.New("event: unset duration")
	// ErrEmptyChord reports a set chord with no notes.
	ErrEmptyChord = errors.New("event: empty chord")
	// ErrMeasureCount reports a nonpositive measure count.
	ErrMeasureCount = errors.New("event: measure count must be positive")
	// ErrWrongKind reports a request for a different content form.
	ErrWrongKind = errors.New("event: wrong content kind")
	// ErrWrongDurationKind reports a request for a different duration form.
	ErrWrongDurationKind = errors.New("event: wrong duration kind")
	// ErrMeasureContent reports a measure-based duration on sounding content.
	ErrMeasureContent = errors.New("event: measure-based duration requires a rest")
	// ErrIncompatibleMark reports a mark that does not apply to the event.
	ErrIncompatibleMark = errors.New("event: incompatible mark")
	// ErrIndex reports a note occurrence index outside the content.
	ErrIndex = errors.New("event: note occurrence index out of range")
	// ErrTiedNoteChanged reports that a preserved tie lost its written note.
	ErrTiedNoteChanged = errors.New("event: tied note changed")
	// ErrNeedsMeasureContext reports that a duration requires measure lengths.
	ErrNeedsMeasureContext = errors.New("event: measure context required")
)

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
