package event

import "slices"

// Equal reports whether both events are set and have the same content,
// duration, and marks by record. Note settings sources are ignored.
func (e *Event) Equal(other *Event) bool {
	return e.IsSet() && other.IsSet() &&
		e.content.Equal(other.content) && e.duration.Equal(other.duration) &&
		e.marks.Fermata == other.marks.Fermata && e.marks.Arpeggio == other.marks.Arpeggio &&
		e.marks.Grace == other.marks.Grace && slices.Equal(e.marks.Ties, other.marks.Ties)
}

// EqualLength reports whether two events occupy the same known length. It
// returns the same errors as [Event.CompareLength].
func (e *Event) EqualLength(other *Event) (bool, error) {
	order, err := e.CompareLength(other)
	if err != nil {
		return false, err
	}
	return order == 0, nil
}

// MustEqualLength is like [Event.EqualLength] but panics with the returned error.
func (e *Event) MustEqualLength(other *Event) bool {
	return must(e.EqualLength(other))
}

// CompareLength returns -1, 0, or 1 as the length of e is less than, equal to,
// or greater than the length of other. Unset events give [ErrUnset] before
// length is checked. A rest in measures gives [ErrNeedsMeasureContext].
func (e *Event) CompareLength(other *Event) (int, error) {
	if !e.IsSet() || !other.IsSet() {
		return 0, ErrUnset
	}
	left, err := e.Length()
	if err != nil {
		return 0, err
	}
	right, err := other.Length()
	if err != nil {
		return 0, err
	}
	return left.Compare(right), nil
}

// MustCompareLength is like [Event.CompareLength] but panics with the returned error.
func (e *Event) MustCompareLength(other *Event) int {
	return must(e.CompareLength(other))
}
