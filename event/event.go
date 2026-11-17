package event

import (
	"fmt"
	"slices"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/note"
)

// Marks records signs attached to an event. The zero value has no signs.
type Marks struct {
	Ties     []int // Zero-based indexes of notes tied to the next event.
	Fermata  bool  // A fermata sign, with its hold determined by the caller.
	Arpeggio bool  // An arpeggio sign on a chord.
	Grace    bool  // A grace note or chord with written duration.
}

// Event is an immutable voice element. A nil *Event and the zero Event are
// unset. Copies, reads, and With methods can be used concurrently. Use
// [Event.Equal] for equality by record and == on pointers for identity.
type Event struct {
	content  Content
	duration Duration
	marks    Marks
}

// New creates an event after validating content, duration, and marks in that
// order. A duration in measures requires a rest. Ties require notes and valid
// indexes, an arpeggio requires a chord, and a grace sign requires sounding
// content with a written duration. Fermatas are allowed on any set event.
// The new event stores its own copy of the tie indexes in ascending order,
// with duplicate indexes removed.
func New(content Content, d Duration, marks Marks) (*Event, error) {
	if !content.IsSet() {
		return nil, ErrUnsetContent
	}
	if !d.IsSet() {
		return nil, ErrUnsetDuration
	}
	if d.kind == DurationMeasures && content.kind != KindRest {
		return nil, ErrMeasureContent
	}
	if err := validateMarks(content, d, marks); err != nil {
		return nil, err
	}

	if len(marks.Ties) > 0 {
		ties := slices.Clone(marks.Ties)
		slices.Sort(ties)
		unique := slices.Compact(ties)
		if len(unique) < len(ties) {
			unique = slices.Clone(unique)
		}
		marks.Ties = unique
	} else {
		marks.Ties = nil
	}

	return &Event{content: content, duration: d, marks: marks}, nil
}

func validateMarks(content Content, d Duration, marks Marks) error {
	if content.kind == KindRest && len(marks.Ties) > 0 {
		return fmt.Errorf("%w: ties require notes", ErrIncompatibleMark)
	}
	count := content.noteCount()
	for _, index := range marks.Ties {
		if err := checkIndex(index, count); err != nil {
			return err
		}
	}
	if marks.Arpeggio && content.kind != KindChord {
		return fmt.Errorf("%w: arpeggio requires a chord", ErrIncompatibleMark)
	}
	if marks.Grace && (content.kind == KindRest || d.kind != DurationWritten) {
		return fmt.Errorf("%w: grace requires notes and a written duration", ErrIncompatibleMark)
	}
	return nil
}

// MustNew is like [New] but panics with the returned error.
func MustNew(content Content, d Duration, marks Marks) *Event {
	return must(New(content, d, marks))
}

// IsSet reports whether e is a constructed event. It accepts nil.
func (e *Event) IsSet() bool {
	return e != nil && e.content.IsSet()
}

// Kind returns the content kind, or [KindUnset] for an unset event.
func (e *Event) Kind() Kind {
	if !e.IsSet() {
		return KindUnset
	}
	return e.content.Kind()
}

// Content returns the event's content. An unset event gives [ErrUnset].
func (e *Event) Content() (Content, error) {
	if !e.IsSet() {
		return Content{}, ErrUnset
	}
	return e.content, nil
}

// MustContent is like [Event.Content] but panics with the returned error.
func (e *Event) MustContent() Content {
	return must(e.Content())
}

// Duration returns the event's duration. An unset event gives [ErrUnset].
func (e *Event) Duration() (Duration, error) {
	if !e.IsSet() {
		return Duration{}, ErrUnset
	}
	return e.duration, nil
}

// MustDuration is like [Event.Duration] but panics with the returned error.
func (e *Event) MustDuration() Duration {
	return must(e.Duration())
}

// Marks returns a copy of the marks. Ties is a new slice of unique indexes in
// ascending order, or a non-nil empty slice. An unset event gives [ErrUnset].
func (e *Event) Marks() (Marks, error) {
	if !e.IsSet() {
		return Marks{}, ErrUnset
	}
	marks := e.marks
	marks.Ties = append([]int{}, marks.Ties...)
	return marks, nil
}

// MustMarks is like [Event.Marks] but panics with the returned error.
func (e *Event) MustMarks() Marks {
	return must(e.Marks())
}

// Length returns the exact space occupied by e in whole notes. A grace event
// occupies zero. A rest in measures gives [ErrNeedsMeasureContext]. An unset
// event gives [ErrUnset].
func (e *Event) Length() (fraction.Fraction, error) {
	if !e.IsSet() {
		return fraction.Fraction{}, ErrUnset
	}
	if e.marks.Grace {
		return fraction.Zero(), nil
	}
	d, err := e.duration.Value()
	if err != nil {
		return fraction.Fraction{}, err
	}
	return d.Fraction()
}

// MustLength is like [Event.Length] but panics with the returned error.
func (e *Event) MustLength() fraction.Fraction {
	return must(e.Length())
}

// IsTied reports whether the note occurrence at index has a tie. An unset
// event gives [ErrUnset]. An invalid index gives [ErrIndex].
func (e *Event) IsTied(index int) (bool, error) {
	if !e.IsSet() {
		return false, ErrUnset
	}
	if err := checkIndex(index, e.content.noteCount()); err != nil {
		return false, err
	}
	_, tied := slices.BinarySearch(e.marks.Ties, index)
	return tied, nil
}

// MustIsTied is like [Event.IsTied] but panics with the returned error.
func (e *Event) MustIsTied(index int) bool {
	return must(e.IsTied(index))
}

// TiedNotes returns a new slice of tied note occurrences in written order.
// With no ties it returns a non-nil empty slice. An unset event gives [ErrUnset].
func (e *Event) TiedNotes() ([]note.Note, error) {
	if !e.IsSet() {
		return nil, ErrUnset
	}
	notes := make([]note.Note, len(e.marks.Ties))
	for i, index := range e.marks.Ties {
		n, err := e.content.At(index)
		if err != nil {
			return nil, err
		}
		notes[i] = n
	}
	return notes, nil
}

// MustTiedNotes is like [Event.TiedNotes] but panics with the returned error.
func (e *Event) MustTiedNotes() []note.Note {
	return must(e.TiedNotes())
}
