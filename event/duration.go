package event

import (
	"fmt"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
)

// DurationKind identifies the form of an event's duration.
type DurationKind uint8

const (
	DurationUnset    DurationKind = iota // Unset duration.
	DurationWritten                      // Written note value, dots, and tuplets.
	DurationValue                        // Exact positive value in whole notes.
	DurationMeasures                     // Positive number of measures.
)

// IsValid reports whether k is one of the three duration forms.
func (k DurationKind) IsValid() bool {
	return k >= DurationWritten && k <= DurationMeasures
}

// Duration holds a written duration, a positive exact value, or a positive
// measure count. Its zero value is unset. Duration can be used as a map key.
// For set durations, == agrees with [Duration.Equal].
type Duration struct {
	kind     DurationKind
	written  duration.Written
	value    duration.Duration
	measures int
}

// WrittenDuration preserves w, including its dots and ordered tuplet chain.
// An unset w gives duration.ErrUnset.
func WrittenDuration(w duration.Written) (Duration, error) {
	if !w.IsSet() {
		return Duration{}, fmt.Errorf("WrittenDuration: %w", duration.ErrUnset)
	}
	return Duration{kind: DurationWritten, written: w}, nil
}

// MustWrittenDuration is like [WrittenDuration] but panics with the returned error.
func MustWrittenDuration(w duration.Written) Duration {
	return must(WrittenDuration(w))
}

// ValueDuration stores a positive fraction of a whole note as a reduced exact
// value. Errors come from duration.FromFraction.
func ValueDuration(value fraction.Fraction) (Duration, error) {
	d, err := duration.FromFraction(value)
	if err != nil {
		return Duration{}, err
	}
	return Duration{kind: DurationValue, value: d}, nil
}

// MustValueDuration is like [ValueDuration] but panics with the returned error.
func MustValueDuration(value fraction.Fraction) Duration {
	return must(ValueDuration(value))
}

// MeasureDuration stores a positive measure count for use with a rest.
// A nonpositive count gives [ErrMeasureCount].
func MeasureDuration(count int) (Duration, error) {
	if count <= 0 {
		return Duration{}, ErrMeasureCount
	}
	return Duration{kind: DurationMeasures, measures: count}, nil
}

// MustMeasureDuration is like [MeasureDuration] but panics with the returned error.
func MustMeasureDuration(count int) Duration {
	return must(MeasureDuration(count))
}

// IsSet reports whether d has one of the three duration forms.
func (d Duration) IsSet() bool {
	return d.kind.IsValid()
}

// Kind returns the duration form, or [DurationUnset] for an unset duration.
func (d Duration) Kind() DurationKind {
	return d.kind
}

// Written returns a copy of the original notation. An unset duration gives
// [ErrUnsetDuration]. Another form gives [ErrWrongDurationKind].
func (d Duration) Written() (duration.Written, error) {
	if !d.IsSet() {
		return duration.Written{}, ErrUnsetDuration
	}
	if d.kind != DurationWritten {
		return duration.Written{}, ErrWrongDurationKind
	}
	return d.written, nil
}

// MustWritten is like [Duration.Written] but panics with the returned error.
func (d Duration) MustWritten() duration.Written {
	return must(d.Written())
}

// Value returns the positive musical value of written or exact duration.
// A measure count gives [ErrNeedsMeasureContext]. An unset duration gives
// [ErrUnsetDuration]. For the occupied length of a grace note, use [Event.Length].
func (d Duration) Value() (duration.Duration, error) {
	switch d.kind {
	case DurationWritten:
		return d.written.Duration()
	case DurationValue:
		return d.value, nil
	case DurationMeasures:
		return duration.Duration{}, ErrNeedsMeasureContext
	default:
		return duration.Duration{}, ErrUnsetDuration
	}
}

// MustValue is like [Duration.Value] but panics with the returned error.
func (d Duration) MustValue() duration.Duration {
	return must(d.Value())
}

// MeasureCount returns the number of measures. An unset duration gives
// [ErrUnsetDuration]. Another form gives [ErrWrongDurationKind].
func (d Duration) MeasureCount() (int, error) {
	if !d.IsSet() {
		return 0, ErrUnsetDuration
	}
	if d.kind != DurationMeasures {
		return 0, ErrWrongDurationKind
	}
	return d.measures, nil
}

// MustMeasureCount is like [Duration.MeasureCount] but panics with the returned error.
func (d Duration) MustMeasureCount() int {
	return must(d.MeasureCount())
}

// Equal reports whether two set durations have the same form and record.
// Written notation, exact value, and measure count are compared within their forms.
func (d Duration) Equal(other Duration) bool {
	return d.IsSet() && d == other
}
