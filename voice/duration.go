package voice

import (
	"fmt"

	"github.com/go-muse/musex/duration"
)

// DurationKind is the form of the duration of an element.
type DurationKind uint8

// Forms of a duration. DurationUnset is the form of the unset Duration.
const (
	// DurationUnset is the zero value: no form.
	DurationUnset DurationKind = iota
	// DurationWritten is a written duration (duration.Written).
	DurationWritten
	// DurationValue is a value without notation (duration.Duration).
	DurationValue
	// DurationTied is a tied duration of two or more written parts
	// (duration.Tied); only a sound can have it.
	DurationTied
)

var durationKindNames = [...]string{
	DurationWritten: "written",
	DurationValue:   "value",
	DurationTied:    "tied",
}

// IsValid reports whether k is one of the three forms.
func (k DurationKind) IsValid() bool {
	return k >= DurationWritten && k <= DurationTied
}

// String returns "written", "value" or "tied"; "DurationKind(unset)" for
// DurationUnset and "DurationKind(invalid)" for values outside the list.
func (k DurationKind) String() string {
	return enumString(k, durationKindNames[:], "DurationKind")
}

// Duration is the duration of an element in one of three forms: a written
// duration, a value without notation, or a tied duration. The notation is
// kept as given: one form is never replaced by another.
//
// Duration is an immutable value. The zero value Duration{} is the unset
// duration. Compare durations with Equal; == is not defined by the contract.
type Duration struct {
	kind    DurationKind
	written duration.Written
	value   duration.Duration
	tied    duration.Tied
	length  duration.Duration
}

// WrittenDuration returns the written duration w, kept as given.
// An unset w reports ErrUnset together with duration.ErrUnset.
func WrittenDuration(w duration.Written) (Duration, error) {
	if !w.IsSet() {
		return Duration{}, unsetError("written duration", duration.ErrUnset)
	}

	length, err := w.Duration()
	if err != nil {
		return Duration{}, fmt.Errorf("voice: written duration: %w", err)
	}

	return Duration{kind: DurationWritten, written: w, length: length}, nil
}

// MustWrittenDuration is like WrittenDuration but panics with its error.
func MustWrittenDuration(w duration.Written) Duration {
	return must(WrittenDuration(w))
}

// ValueDuration returns the duration with the value d and no notation.
// An unset d reports ErrUnset together with duration.ErrUnset.
func ValueDuration(d duration.Duration) (Duration, error) {
	if !d.IsSet() {
		return Duration{}, unsetError("duration value", duration.ErrUnset)
	}

	return Duration{kind: DurationValue, value: d, length: d}, nil
}

// MustValueDuration is like ValueDuration but panics with its error.
func MustValueDuration(d duration.Duration) Duration {
	return must(ValueDuration(d))
}

// TiedDuration returns the tied duration t of two or more written parts.
// An unset t reports ErrUnset together with duration.ErrUnset.
func TiedDuration(t duration.Tied) (Duration, error) {
	if !t.IsSet() {
		return Duration{}, unsetError("tied duration", duration.ErrUnset)
	}

	length, err := t.Duration()
	if err != nil {
		return Duration{}, fmt.Errorf("voice: tied duration: %w", err)
	}

	return Duration{kind: DurationTied, tied: t, length: length}, nil
}

// MustTiedDuration is like TiedDuration but panics with its error.
func MustTiedDuration(t duration.Tied) Duration {
	return must(TiedDuration(t))
}

// IsSet reports whether d is a set duration.
func (d Duration) IsSet() bool {
	return d.kind != DurationUnset
}

// Kind returns the form of d; DurationUnset for the unset duration.
func (d Duration) Kind() DurationKind {
	return d.kind
}

// Written returns the written duration and true for the written form, and
// duration.Written{} and false for the other forms. The unset duration
// reports ErrUnset.
func (d Duration) Written() (duration.Written, bool, error) {
	if !d.IsSet() {
		return duration.Written{}, false, unsetError("duration", nil)
	}

	if d.kind != DurationWritten {
		return duration.Written{}, false, nil
	}

	return d.written, true, nil
}

// MustWritten is like Written but panics with its error.
func (d Duration) MustWritten() (duration.Written, bool) {
	return must2(d.Written())
}

// Tied returns the tied duration and true for the tied form, and
// duration.Tied{} and false for the other forms. The unset duration reports
// ErrUnset.
func (d Duration) Tied() (duration.Tied, bool, error) {
	if !d.IsSet() {
		return duration.Tied{}, false, unsetError("duration", nil)
	}

	if d.kind != DurationTied {
		return duration.Tied{}, false, nil
	}

	return d.tied, true, nil
}

// MustTied is like Tied but panics with its error.
func (d Duration) MustTied() (duration.Tied, bool) {
	return must2(d.Tied())
}

// Length returns the value of d, always positive: the value of a written or
// a tied duration, or the value itself. The unset duration reports ErrUnset.
func (d Duration) Length() (duration.Duration, error) {
	if !d.IsSet() {
		return duration.Duration{}, unsetError("duration", nil)
	}

	return d.length, nil
}

// MustLength is like Length but panics with its error.
func (d Duration) MustLength() duration.Duration {
	return must(d.Length())
}

// Equal reports whether d and other have the same form and the same
// notation: a dotted quarter and a tied quarter and eighth are not equal.
// Two unset durations are equal; an unset duration equals no set one.
func (d Duration) Equal(other Duration) bool {
	if d.kind != other.kind {
		return false
	}

	switch d.kind {
	case DurationWritten:
		return d.written.Equal(other.written)
	case DurationValue:
		return d.value.Equal(other.value)
	case DurationTied:
		return d.tied.Equal(other.tied)
	default:
		return true
	}
}

// String returns a debugging form, such as "Duration(written 4.)", or
// "Duration(unset)". It is not a notation and has no parser.
func (d Duration) String() string {
	switch d.kind {
	case DurationWritten:
		return "Duration(written " + d.written.String() + ")"
	case DurationValue:
		return "Duration(value " + d.value.String() + ")"
	case DurationTied:
		return "Duration(tied " + d.tied.String() + ")"
	default:
		return "Duration(unset)"
	}
}
