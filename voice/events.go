package voice

import (
	"fmt"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/note"
)

// eventRecord is a checked event of FromEvents.
type eventRecord struct {
	kind     event.Kind
	note     note.Note
	dur      event.Duration
	fermata  bool
	tiedNext bool
}

// readEvent returns the record of e or the first failure of the checks 1-5
// of FromEvents.
func readEvent(e *event.Event) (eventRecord, error) {
	if !e.IsSet() {
		return eventRecord{}, unsetError("event", event.ErrUnset)
	}

	content, err := e.Content()
	if err != nil {
		return eventRecord{}, fmt.Errorf("voice: event content: %w", err)
	}

	d, err := e.Duration()
	if err != nil {
		return eventRecord{}, fmt.Errorf("voice: event duration: %w", err)
	}

	marks, err := e.Marks()
	if err != nil {
		return eventRecord{}, fmt.Errorf("voice: event marks: %w", err)
	}

	r := eventRecord{kind: content.Kind(), dur: d, fermata: marks.Fermata}

	switch {
	case r.kind == event.KindChord:
		return eventRecord{}, fmt.Errorf("%w: chord event", ErrChord)
	case d.Kind() == event.DurationMeasures:
		return eventRecord{}, fmt.Errorf("%w: rest of %d measures", ErrMeasureDuration, d.MustMeasureCount())
	case marks.Grace:
		return eventRecord{}, fmt.Errorf("%w: grace event", ErrGrace)
	}

	if r.kind != event.KindNote {
		return r, nil
	}

	if r.note, err = content.Note(); err != nil {
		return eventRecord{}, fmt.Errorf("voice: event note: %w", err)
	}

	if !r.note.HasOctave() {
		return eventRecord{}, fmt.Errorf("%w: event note %s", ErrNoOctave, r.note)
	}

	for _, t := range marks.Ties {
		if t == 0 {
			r.tiedNext = true
		}
	}

	return r, nil
}

// checkTie returns the failure of the checks 6-8 of FromEvents for the note
// event k with the record r; prevTied reports that the event k-1 is tied to
// it.
func checkTie(events []*event.Event, k int, r eventRecord, prevTied bool) error {
	if (r.tiedNext || prevTied) && r.dur.Kind() == event.DurationValue {
		return fmt.Errorf("%w: event in a tie with a value %s", ErrTiedValue, r.dur.MustValue())
	}

	if !r.tiedNext {
		return nil
	}

	if r.fermata {
		return fmt.Errorf("%w: fermata on an event tied to the next one", ErrFermataInTie)
	}

	if k+1 >= len(events) || !events[k+1].IsSet() || events[k+1].Kind() != event.KindNote {
		return fmt.Errorf("%w: no note after the tie", ErrBrokenTie)
	}

	next, err := events[k+1].MustContent().Note()
	if err != nil {
		return fmt.Errorf("voice: next event note: %w", err)
	}

	if !next.Equal(r.note) {
		return fmt.Errorf("%w: %s tied to %s", ErrBrokenTie, r.note, next)
	}

	return nil
}

// FromEvents returns a line of events of package event in order, without a
// name and roles; without arguments the empty line. A note event outside a
// tie becomes a sound with its note, duration and fermata; a series of note
// events k to m tied to each other becomes one sound with the note of the
// event k (its correction and source), the tied duration of their written
// durations and the fermata of the event m; a rest event becomes a rest.
// An event is tied to the next one when its Marks.Ties contain 0.
//
// The events are checked in order, each completely before the next one.
// Every failure of an event is an *ItemError with its index: in this order,
// ErrUnset for a nil or unset event (with event.ErrUnset); ErrChord for a
// chord, even of one note; ErrMeasureDuration for a duration in measures;
// ErrGrace for a grace event; ErrNoOctave for a note without an octave;
// ErrTiedValue for an event of a tied series with a value duration;
// ErrFermataInTie for a fermata on an event tied to the next one;
// ErrBrokenTie when the next event is missing, unset, not a note or a note of
// another record (C♯4 and D♭4 differ). After all the checks, a duration or a
// start out of the range of package common/fraction reports the error of
// package duration (with fraction.ErrOutOfRange) without an *ItemError.
func FromEvents(events ...*event.Event) (*Line, error) {
	records := make([]eventRecord, len(events))
	prevTied := false

	for k, e := range events {
		r, err := readEvent(e)
		if err == nil && r.kind == event.KindNote {
			err = checkTie(events, k, r, prevTied)
		}

		if err != nil {
			return nil, itemError(k, err)
		}

		records[k] = r
		prevTied = r.tiedNext
	}

	elems, err := elementsOf(records)
	if err != nil {
		return nil, err
	}

	return newLine(elems, nil)
}

// MustFromEvents is like FromEvents but panics with its error.
func MustFromEvents(events ...*event.Event) *Line {
	return must(FromEvents(events...))
}

// elementsOf returns the elements of checked event records.
func elementsOf(records []eventRecord) ([]Element, error) {
	elems := make([]Element, 0, len(records))

	for k := 0; k < len(records); k++ {
		r := records[k]

		if r.kind == event.KindRest {
			d, err := durationOf(r.dur)
			if err != nil {
				return nil, err
			}

			elems = append(elems, Element{kind: KindRest, dur: d, fermata: r.fermata})

			continue
		}

		if !r.tiedNext {
			d, err := durationOf(r.dur)
			if err != nil {
				return nil, err
			}

			elems = append(elems, Element{kind: KindSound, note: r.note, dur: d, fermata: r.fermata})

			continue
		}

		parts := []duration.Written{r.dur.MustWritten()}
		m := k

		for records[m].tiedNext {
			m++
			parts = append(parts, records[m].dur.MustWritten())
		}

		t, err := duration.NewTied(parts...)
		if err != nil {
			return nil, fmt.Errorf("voice: tied duration of events %d to %d: %w", k, m, err)
		}

		d, err := TiedDuration(t)
		if err != nil {
			return nil, err
		}

		elems = append(elems, Element{kind: KindSound, note: r.note, dur: d, fermata: records[m].fermata})
		k = m
	}

	return elems, nil
}

// durationOf returns the duration of an event written or given by value.
func durationOf(d event.Duration) (Duration, error) {
	if d.Kind() == event.DurationWritten {
		return WrittenDuration(d.MustWritten())
	}

	return ValueDuration(d.MustValue())
}

// Events returns a new slice of new events of package event for the elements
// in order: a sound with a written or a value duration gives one note event;
// a tied sound of the parts w₁…wₘ gives m note events with its note and these
// durations, the first m − 1 tied to the next one and only the last with the
// fermata; a rest gives a rest event. FromEvents of the result is equal to l
// by EqualElements. An unset line reports ErrUnset.
func (l *Line) Events() ([]*event.Event, error) {
	if !l.IsSet() {
		return nil, errUnsetLine()
	}

	out := make([]*event.Event, 0, len(l.elems))

	for _, e := range l.elems {
		content := event.RestContent()
		if e.kind == KindSound {
			content = event.MustNoteContent(e.note)
		}

		switch e.dur.kind {
		case DurationTied:
			parts := e.dur.tied.MustParts()
			for k, w := range parts {
				marks := event.Marks{Ties: []int{0}}
				if k == len(parts)-1 {
					marks = event.Marks{Fermata: e.fermata}
				}

				out = append(out, event.MustNew(content, event.MustWrittenDuration(w), marks))
			}
		case DurationWritten:
			out = append(out, event.MustNew(content, event.MustWrittenDuration(e.dur.written), event.Marks{Fermata: e.fermata}))
		default:
			value := event.MustValueDuration(e.dur.value.MustFraction())
			out = append(out, event.MustNew(content, value, event.Marks{Fermata: e.fermata}))
		}
	}

	return out, nil
}

// MustEvents is like Events but panics with its error.
func (l *Line) MustEvents() []*event.Event {
	return must(l.Events())
}
