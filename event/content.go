package event

import (
	"fmt"

	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/note"
)

// Kind identifies the form of an event's content.
type Kind uint8

const (
	KindUnset Kind = iota // Unset content or event.
	KindNote              // A single note.
	KindChord             // A nonempty chord, possibly of one note.
	KindRest              // A rest without notes.
)

// IsValid reports whether k is a note, chord, or rest kind.
func (k Kind) IsValid() bool {
	return k >= KindNote && k <= KindRest
}

// Content holds a note, a nonempty chord, or a rest. Its zero value is unset.
// Copies can be read independently. Use [Content.Equal] to compare records.
type Content struct {
	kind  Kind
	note  note.Note
	chord chord.Chord
}

// NoteContent creates content for a copy of n. An unset note gives note.ErrUnset.
func NoteContent(n note.Note) (Content, error) {
	if !n.IsValid() {
		return Content{}, fmt.Errorf("NoteContent: %w", note.ErrUnset)
	}
	return Content{kind: KindNote, note: n}, nil
}

// MustNoteContent is like [NoteContent] but panics with the returned error.
func MustNoteContent(n note.Note) Content {
	return must(NoteContent(n))
}

// ChordContent creates content for an independent copy of c. An unset chord
// gives chord.ErrUnset and a set, empty chord gives [ErrEmptyChord].
func ChordContent(c chord.Chord) (Content, error) {
	if !c.IsValid() {
		return Content{}, fmt.Errorf("ChordContent: %w", chord.ErrUnset)
	}
	if c.IsEmpty() {
		return Content{}, ErrEmptyChord
	}
	return Content{kind: KindChord, chord: c}, nil
}

// MustChordContent is like [ChordContent] but panics with the returned error.
func MustChordContent(c chord.Chord) Content {
	return must(ChordContent(c))
}

// RestContent creates an explicit rest with no notes.
func RestContent() Content {
	return Content{kind: KindRest}
}

// IsSet reports whether c was created as a note, chord, or rest.
func (c Content) IsSet() bool {
	return c.kind.IsValid()
}

// Kind returns the content kind, or [KindUnset] for unset content.
func (c Content) Kind() Kind {
	return c.kind
}

// Note returns a copy of the note. Unset content gives [ErrUnsetContent].
// Content of a different kind gives [ErrWrongKind].
func (c Content) Note() (note.Note, error) {
	if !c.IsSet() {
		return note.Note{}, ErrUnsetContent
	}
	if c.kind != KindNote {
		return note.Note{}, ErrWrongKind
	}
	return c.note, nil
}

// MustNote is like [Content.Note] but panics with the returned error.
func (c Content) MustNote() note.Note {
	return must(c.Note())
}

// Chord returns an independent copy of the chord. Unset content gives
// [ErrUnsetContent]. Content of a different kind gives [ErrWrongKind].
func (c Content) Chord() (chord.Chord, error) {
	if !c.IsSet() {
		return chord.Chord{}, ErrUnsetContent
	}
	if c.kind != KindChord {
		return chord.Chord{}, ErrWrongKind
	}
	return c.chord, nil
}

// MustChord is like [Content.Chord] but panics with the returned error.
func (c Content) MustChord() chord.Chord {
	return must(c.Chord())
}

// Len returns the number of note occurrences, counting repeats. A rest has
// zero occurrences. Unset content gives [ErrUnsetContent].
func (c Content) Len() (int, error) {
	if !c.IsSet() {
		return 0, ErrUnsetContent
	}
	return c.noteCount(), nil
}

func (c Content) noteCount() int {
	if c.kind == KindNote {
		return 1
	}
	return c.chord.Len()
}

// MustLen is like [Content.Len] but panics with the returned error.
func (c Content) MustLen() int {
	return must(c.Len())
}

// Notes returns a new slice of note copies in written order, keeping repeats.
// A rest gives a non-nil empty slice. Unset content gives [ErrUnsetContent].
func (c Content) Notes() ([]note.Note, error) {
	if !c.IsSet() {
		return nil, ErrUnsetContent
	}
	if c.kind == KindNote {
		return []note.Note{c.note}, nil
	}
	return c.chord.Notes(), nil
}

// MustNotes is like [Content.Notes] but panics with the returned error.
func (c Content) MustNotes() []note.Note {
	return must(c.Notes())
}

// At returns a copy of the note at the given zero-based index in written
// order. Unset content gives [ErrUnsetContent]. An invalid index gives [ErrIndex].
func (c Content) At(index int) (note.Note, error) {
	if !c.IsSet() {
		return note.Note{}, ErrUnsetContent
	}
	if err := checkIndex(index, c.noteCount()); err != nil {
		return note.Note{}, err
	}
	if c.kind == KindNote {
		return c.note, nil
	}
	return c.chord.At(index)
}

func checkIndex(index, count int) error {
	if index < 0 || index >= count {
		return fmt.Errorf("%w: index %d, %d notes", ErrIndex, index, count)
	}
	return nil
}

// MustAt is like [Content.At] but panics with the returned error.
func (c Content) MustAt(index int) note.Note {
	return must(c.At(index))
}

// Contains reports whether c has a note equal to n by written record. Unset
// content gives [ErrUnsetContent]. An unset note gives note.ErrUnset.
func (c Content) Contains(n note.Note) (bool, error) {
	if !c.IsSet() {
		return false, ErrUnsetContent
	}
	if !n.IsValid() {
		return false, fmt.Errorf("Contains: %w", note.ErrUnset)
	}
	if c.kind == KindNote {
		return c.note.Equal(n), nil
	}
	return c.chord.Contains(n), nil
}

// MustContains is like [Content.Contains] but panics with the returned error.
func (c Content) MustContains(n note.Note) bool {
	return must(c.Contains(n))
}

// Equal reports whether both values are set and have the same kind and the
// same written notes in the same order. Note settings sources are ignored.
func (c Content) Equal(other Content) bool {
	if c.kind != other.kind {
		return false
	}
	switch c.kind {
	case KindNote:
		return c.note.Equal(other.note)
	case KindChord:
		return c.chord.EqualInOrder(other.chord)
	case KindRest:
		return true
	default:
		return false
	}
}
