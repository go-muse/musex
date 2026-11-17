package event

import "fmt"

// WithContent creates an event with the new content and the original
// duration and marks. If a tied occurrence disappears or changes its note
// by written record, WithContent gives [ErrTiedNoteChanged]. The result
// must satisfy the rules of [New]. An unset receiver gives [ErrUnset].
func (e *Event) WithContent(content Content) (*Event, error) {
	if !e.IsSet() {
		return nil, ErrUnset
	}
	if !content.IsSet() {
		return nil, ErrUnsetContent
	}
	for _, index := range e.marks.Ties {
		old, err := e.content.At(index)
		if err != nil {
			return nil, err
		}
		replacement, err := content.At(index)
		if err != nil || !old.Equal(replacement) {
			return nil, fmt.Errorf("%w: index %d", ErrTiedNoteChanged, index)
		}
	}
	return New(content, e.duration, e.marks)
}

// MustWithContent is like [Event.WithContent] but panics with the returned error.
func (e *Event) MustWithContent(content Content) *Event {
	return must(e.WithContent(content))
}

// WithContentAndMarks creates an event with new content and marks and the
// original duration. [New] validates ties against the new content. An unset
// receiver gives [ErrUnset].
func (e *Event) WithContentAndMarks(content Content, marks Marks) (*Event, error) {
	if !e.IsSet() {
		return nil, ErrUnset
	}
	return New(content, e.duration, marks)
}

// MustWithContentAndMarks is like [Event.WithContentAndMarks] but panics with
// the returned error.
func (e *Event) MustWithContentAndMarks(content Content, marks Marks) *Event {
	return must(e.WithContentAndMarks(content, marks))
}

// WithDuration creates an event with the new duration and the original
// content and marks, validated as in [New]. An unset receiver gives [ErrUnset].
func (e *Event) WithDuration(d Duration) (*Event, error) {
	if !e.IsSet() {
		return nil, ErrUnset
	}
	return New(e.content, d, e.marks)
}

// MustWithDuration is like [Event.WithDuration] but panics with the returned error.
func (e *Event) MustWithDuration(d Duration) *Event {
	return must(e.WithDuration(d))
}

// WithMarks creates an event with all marks replaced and with the original
// content and duration. [New] validates the result. An unset receiver gives
// [ErrUnset].
func (e *Event) WithMarks(marks Marks) (*Event, error) {
	if !e.IsSet() {
		return nil, ErrUnset
	}
	return New(e.content, e.duration, marks)
}

// MustWithMarks is like [Event.WithMarks] but panics with the returned error.
func (e *Event) MustWithMarks(marks Marks) *Event {
	return must(e.WithMarks(marks))
}
