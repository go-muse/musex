package template

import "github.com/go-muse/musex/interval"

// MaxPlaces is the largest number of places in a template.
const MaxPlaces = 1 << 16

// Template is an immutable mode template: an ordered sequence of ascending
// intervals between neighboring places. A closed template has one interval per
// place, the last of which closes the cycle. An open template has one fewer
// interval than places and no cycle. The zero Template is unset.
//
// Template is not comparable because it contains a slice. Use Equal for domain
// equality and Compare for sorting. Copies may share immutable internal data.
// All slices accepted or returned by the API are copied as described by the
// individual operations, and concurrent reads are safe without synchronization.
type Template struct {
	intervals []interval.Interval
	open      bool
}

// IsValid reports whether t is a set template.
func (t Template) IsValid() bool {
	return len(t.intervals) > 0
}

// Kind returns Closed, Open, or KindUnset for the zero Template.
func (t Template) Kind() Kind {
	if !t.IsValid() {
		return KindUnset
	}
	if t.open {
		return Open
	}
	return Closed
}

// IsClosed reports whether t is a set closed template.
func (t Template) IsClosed() bool {
	return t.IsValid() && !t.open
}

// IsOpen reports whether t is a set open template.
func (t Template) IsOpen() bool {
	return t.IsValid() && t.open
}

// Len returns the number of places, or zero for an unset template.
func (t Template) Len() int {
	if !t.IsValid() {
		return 0
	}
	if t.open {
		return len(t.intervals) + 1
	}
	return len(t.intervals)
}

// AdjacentIntervals returns a new slice containing all adjacent intervals in
// order. For a closed template the last interval is the closing interval. It
// returns nil for an unset template.
func (t Template) AdjacentIntervals() []interval.Interval {
	if !t.IsValid() {
		return nil
	}
	return append([]interval.Interval{}, t.intervals...)
}

// Equal reports whether both templates are set and have the same kind and
// adjacent intervals. An unset template is equal to no template, including
// another unset template.
func (t Template) Equal(other Template) bool {
	if !t.IsValid() || !other.IsValid() || t.open != other.open || len(t.intervals) != len(other.intervals) {
		return false
	}
	for i, current := range t.intervals {
		if !current.Equal(other.intervals[i]) {
			return false
		}
	}
	return true
}

// Compare compares templates in a total order suitable for sorting: unset
// before set, closed before open, then by place count and adjacent intervals
// from left to right. It returns only -1, 0, or 1. For set templates, zero is
// equivalent to Equal.
func (t Template) Compare(other Template) int {
	if !t.IsValid() {
		if !other.IsValid() {
			return 0
		}
		return -1
	}
	if !other.IsValid() {
		return 1
	}
	if t.open != other.open {
		if t.open {
			return 1
		}
		return -1
	}
	if t.Len() < other.Len() {
		return -1
	}
	if t.Len() > other.Len() {
		return 1
	}
	for i, current := range t.intervals {
		if result := current.Compare(other.intervals[i]); result != 0 {
			return result
		}
	}
	return 0
}

// Rotate returns the closed template starting at place, or the same open
// template when place is 1. An open template has no other rotations.
func (t Template) Rotate(place int) (Template, error) {
	if !t.IsValid() {
		return Template{}, opError("Rotate", ErrUnset)
	}
	if place < 1 || place > t.Len() {
		return Template{}, opError("Rotate", ErrInvalidArgument)
	}
	if t.open {
		if place > 1 {
			return Template{}, opError("Rotate", ErrOpenRotation)
		}
		return t, nil
	}
	if place == 1 {
		return t, nil
	}
	start := place - 1
	rotated := make([]interval.Interval, 0, len(t.intervals))
	rotated = append(rotated, t.intervals[start:]...)
	rotated = append(rotated, t.intervals[:start]...)
	return Template{intervals: rotated}, nil
}

// MustRotate is Rotate without the error result. It panics with the same error.
func (t Template) MustRotate(place int) Template {
	return must(t.Rotate(place))
}

// RotationOf reports whether t is a rotation of other and returns the smallest
// place of other that produces t. For open templates only place 1 is possible.
// It returns (0, false) if either template is unset or their kinds differ.
func (t Template) RotationOf(other Template) (place int, ok bool) {
	if !t.IsValid() || !other.IsValid() || t.open != other.open || t.Len() != other.Len() {
		return 0, false
	}
	if t.open {
		if t.Equal(other) {
			return 1, true
		}
		return 0, false
	}
	k := other.Len()
	for start := 0; start < k; start++ {
		matches := true
		for i := 0; i < k; i++ {
			if !other.intervals[(start+i)%k].Equal(t.intervals[i]) {
				matches = false
				break
			}
		}
		if matches {
			return start + 1, true
		}
	}
	return 0, false
}
