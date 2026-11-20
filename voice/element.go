package voice

import (
	"fmt"
	"strings"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/note"
)

// Kind is the kind of an element: a sound or a rest.
type Kind uint8

// Kinds of an element. KindUnset is the kind of the unset Element.
const (
	// KindUnset is the zero value: no kind.
	KindUnset Kind = iota
	// KindSound is a sound: one note with an octave.
	KindSound
	// KindRest is a rest.
	KindRest
)

var kindNames = [...]string{
	KindSound: "sound",
	KindRest:  "rest",
}

// IsValid reports whether k is KindSound or KindRest.
func (k Kind) IsValid() bool {
	return k == KindSound || k == KindRest
}

// String returns "sound" or "rest"; "Kind(unset)" for KindUnset and
// "Kind(invalid)" for values outside the list.
func (k Kind) String() string {
	return enumString(k, kindNames[:], "Kind")
}

// Element is an element of a line: a sound with one note with an octave, or a
// rest; a duration; and a fermata sign. A tied duration is allowed only for a
// sound, which then is one sound under a tie. Grace notes, chords and arpeggios
// are never elements.
//
// Element is an immutable value: a sound keeps a copy of its note, which keeps
// the reference to its source of settings. The zero value Element{} is the
// unset element. Compare elements with Equal; == is not defined by the
// contract.
type Element struct {
	kind    Kind
	note    note.Note
	dur     Duration
	fermata bool
}

// NewSound returns a sound with a copy of n, the duration d and no fermata.
// The record, the correction and the source of n are kept as given.
//
// Errors, in this order: ErrUnset for an unset n (with note.ErrUnset), then
// for an unset d; ErrNoOctave when n has no octave.
func NewSound(n note.Note, d Duration) (Element, error) {
	if !n.IsValid() {
		return Element{}, unsetError("note", note.ErrUnset)
	}

	if !d.IsSet() {
		return Element{}, unsetError("duration", nil)
	}

	if !n.HasOctave() {
		return Element{}, fmt.Errorf("%w: %s", ErrNoOctave, n)
	}

	return Element{kind: KindSound, note: n, dur: d}, nil
}

// MustNewSound is like NewSound but panics with its error.
func MustNewSound(n note.Note, d Duration) Element {
	return must(NewSound(n, d))
}

// NewRest returns a rest with the duration d and no fermata.
//
// Errors, in this order: ErrUnset for an unset d; ErrTiedRest for a tied d.
func NewRest(d Duration) (Element, error) {
	if !d.IsSet() {
		return Element{}, unsetError("duration", nil)
	}

	if d.kind == DurationTied {
		return Element{}, fmt.Errorf("%w: %s", ErrTiedRest, d)
	}

	return Element{kind: KindRest, dur: d}, nil
}

// MustNewRest is like NewRest but panics with its error.
func MustNewRest(d Duration) Element {
	return must(NewRest(d))
}

// IsSet reports whether e is a set element.
func (e Element) IsSet() bool {
	return e.kind != KindUnset
}

// Kind returns the kind of e; KindUnset for the unset element.
func (e Element) Kind() Kind {
	return e.kind
}

// IsSound reports whether e is a sound; false for the unset element.
func (e Element) IsSound() bool {
	return e.kind == KindSound
}

// IsRest reports whether e is a rest; false for the unset element.
func (e Element) IsRest() bool {
	return e.kind == KindRest
}

// Note returns a copy of the note and true for a sound, and note.Note{} and
// false for a rest. The unset element reports ErrUnset.
func (e Element) Note() (note.Note, bool, error) {
	if !e.IsSet() {
		return note.Note{}, false, unsetError("element", nil)
	}

	if e.kind != KindSound {
		return note.Note{}, false, nil
	}

	return e.note, true, nil
}

// MustNote is like Note but panics with its error.
func (e Element) MustNote() (note.Note, bool) {
	return must2(e.Note())
}

// Duration returns the duration of e. The unset element reports ErrUnset.
func (e Element) Duration() (Duration, error) {
	if !e.IsSet() {
		return Duration{}, unsetError("element", nil)
	}

	return e.dur, nil
}

// MustDuration is like Duration but panics with its error.
func (e Element) MustDuration() Duration {
	return must(e.Duration())
}

// Length returns the value of the duration of e, always positive. The unset
// element reports ErrUnset.
func (e Element) Length() (duration.Duration, error) {
	if !e.IsSet() {
		return duration.Duration{}, unsetError("element", nil)
	}

	return e.dur.length, nil
}

// MustLength is like Length but panics with its error.
func (e Element) MustLength() duration.Duration {
	return must(e.Length())
}

// Fermata reports whether e has a fermata sign. The unset element reports
// ErrUnset.
func (e Element) Fermata() (bool, error) {
	if !e.IsSet() {
		return false, unsetError("element", nil)
	}

	return e.fermata, nil
}

// MustFermata is like Fermata but panics with its error.
func (e Element) MustFermata() bool {
	return must(e.Fermata())
}

// WithNote returns a sound with a copy of n and the duration and the fermata
// of e: a rest becomes a sound.
//
// Errors, in this order: ErrUnset for an unset e, then for an unset n (with
// note.ErrUnset); ErrNoOctave when n has no octave.
func (e Element) WithNote(n note.Note) (Element, error) {
	if !e.IsSet() {
		return Element{}, unsetError("element", nil)
	}

	if !n.IsValid() {
		return Element{}, unsetError("note", note.ErrUnset)
	}

	if !n.HasOctave() {
		return Element{}, fmt.Errorf("%w: %s", ErrNoOctave, n)
	}

	return Element{kind: KindSound, note: n, dur: e.dur, fermata: e.fermata}, nil
}

// MustWithNote is like WithNote but panics with its error.
func (e Element) MustWithNote(n note.Note) Element {
	return must(e.WithNote(n))
}

// WithDuration returns e with the duration d.
//
// Errors, in this order: ErrUnset for an unset e, then for an unset d;
// ErrTiedRest for a rest and a tied d.
func (e Element) WithDuration(d Duration) (Element, error) {
	if !e.IsSet() {
		return Element{}, unsetError("element", nil)
	}

	if !d.IsSet() {
		return Element{}, unsetError("duration", nil)
	}

	if e.kind == KindRest && d.kind == DurationTied {
		return Element{}, fmt.Errorf("%w: %s", ErrTiedRest, d)
	}

	e.dur = d

	return e, nil
}

// MustWithDuration is like WithDuration but panics with its error.
func (e Element) MustWithDuration(d Duration) Element {
	return must(e.WithDuration(d))
}

// WithFermata returns e with the fermata sign on or off. A fermata of a tied
// sound belongs to the whole sound. The unset element reports ErrUnset.
func (e Element) WithFermata(on bool) (Element, error) {
	if !e.IsSet() {
		return Element{}, unsetError("element", nil)
	}

	e.fermata = on

	return e, nil
}

// MustWithFermata is like WithFermata but panics with its error.
func (e Element) MustWithFermata(on bool) Element {
	return must(e.WithFermata(on))
}

// Equal reports whether e and other have the same kind, notes equal by
// record and correction (note.Note.Equal; the source does not take part),
// durations equal by Duration.Equal and the same fermata sign. Two unset
// elements are equal; an unset element equals no set one.
func (e Element) Equal(other Element) bool {
	return e.equalBy(other, Duration.Equal)
}

// EqualDuration is like Equal but compares the durations by value (Length):
// the notation and the form of the durations do not take part.
func (e Element) EqualDuration(other Element) bool {
	return e.equalBy(other, func(a, b Duration) bool {
		return a.length.Equal(b.length)
	})
}

func (e Element) equalBy(other Element, sameDuration func(a, b Duration) bool) bool {
	if e.kind != other.kind {
		return false
	}

	if !e.IsSet() {
		return true
	}

	if e.fermata != other.fermata || !sameDuration(e.dur, other.dur) {
		return false
	}

	return e.kind != KindSound || e.note.Equal(other.note)
}

// String returns a debugging form, such as "Element(sound C ASPN 4, 0 cents,
// bound; Duration(written 4); fermata)", or "Element(unset)". It is not a
// notation and has no parser.
func (e Element) String() string {
	if !e.IsSet() {
		return "Element(unset)"
	}

	var b strings.Builder

	b.WriteString("Element(")
	b.WriteString(e.kind.String())

	if e.kind == KindSound {
		b.WriteByte(' ')
		b.WriteString(e.note.String())
	}

	b.WriteString("; ")
	b.WriteString(e.dur.String())

	if e.fermata {
		b.WriteString("; fermata")
	}

	b.WriteByte(')')

	return b.String()
}
