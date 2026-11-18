package scale

import (
	"errors"
	"fmt"
	"slices"
	"unsafe"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
)

// Scale owns an ordered list. Its zero value is an empty scale. Mutations
// replace the backing list, so copies made by assignment remain independent.
type Scale struct{ notes []note.Note }

// Direction selects a direction through a list.
type Direction int8

const (
	DirectionUnset Direction = 0
	Forward        Direction = 1
	Backward       Direction = -1
)

func (d Direction) IsValid() bool { return d == Forward || d == Backward }
func (d Direction) String() string {
	switch d {
	case Forward:
		return "forward"
	case Backward:
		return "backward"
	case DirectionUnset:
		return "unset"
	}
	return fmt.Sprintf("Direction(%d)", d)
}

// Element identifies one occurrence in a scale.
type Element struct {
	Index int
	Note  note.Note
}

// Selection optionally disambiguates a note with an absolute list index.
type Selection struct {
	Note  note.Note
	Index *int
}

// Signature describes standard key signs along the line of fifths.
type Signature struct {
	Index int
	Signs []spelling.Spelling
}

var (
	ErrUnset           = errors.New("scale: unset value")
	ErrInvalidArgument = errors.New("scale: invalid argument")
	ErrEmpty           = errors.New("scale: empty scale")
	ErrIndex           = errors.New("scale: index out of range")
	ErrNotFound        = errors.New("scale: note not found")
	ErrAmbiguous       = errors.New("scale: ambiguous occurrence")
	ErrNotEquivalent   = errors.New("scale: spelling is not equivalent")
	ErrOutOfRange      = errors.New("scale: result out of range")
)

func validNotes(ns []note.Note) error {
	for _, n := range ns {
		if !n.IsValid() {
			return errors.Join(ErrUnset, note.ErrUnset)
		}
	}
	return nil
}
func owned(ns []note.Note) []note.Note { return append(make([]note.Note, 0, len(ns)), ns...) }

// New copies notes, preserving order and repetitions.
func New(notes ...note.Note) (*Scale, error) {
	if err := validNotes(notes); err != nil {
		return nil, err
	}
	return &Scale{notes: owned(notes)}, nil
}

// IsValid reports whether the receiver is present.
func (s *Scale) IsValid() bool { return s != nil }

// Len returns the number of notes, or zero for a nil receiver.
func (s *Scale) Len() int {
	if s == nil {
		return 0
	}
	return len(s.notes)
}

// IsEmpty reports whether the receiver is a present, empty scale.
func (s *Scale) IsEmpty() bool { return s != nil && len(s.notes) == 0 }

// Notes returns an independent slice, or nil for an absent receiver.
func (s *Scale) Notes() []note.Note {
	if s == nil {
		return nil
	}
	return owned(s.notes)
}

// Clone returns an independently editable scale, or nil for an absent receiver.
func (s *Scale) Clone() *Scale {
	if s == nil {
		return nil
	}
	return &Scale{notes: owned(s.notes)}
}
func (s *Scale) check() error {
	if s == nil {
		return ErrUnset
	}
	return nil
}
func (s *Scale) index(i int) error {
	if s == nil {
		return ErrUnset
	}
	if i < 0 || i >= len(s.notes) {
		return ErrIndex
	}
	return nil
}
func (s *Scale) elementIndex(i int) error {
	if s == nil {
		return ErrUnset
	}
	if len(s.notes) == 0 {
		return ErrEmpty
	}
	return s.index(i)
}

// At reads an existing zero-based index without wrapping.
func (s *Scale) At(index int) (note.Note, error) {
	if err := s.index(index); err != nil {
		return note.Note{}, err
	}
	return s.notes[index], nil
}

// First returns the first note, or ErrEmpty for an empty scale.
func (s *Scale) First() (note.Note, error) {
	if err := s.elementIndex(0); err != nil {
		return note.Note{}, err
	}
	return s.notes[0], nil
}

// Last returns the last note, or ErrEmpty for an empty scale.
func (s *Scale) Last() (note.Note, error) {
	if err := s.elementIndex(s.Len() - 1); err != nil {
		return note.Note{}, err
	}
	return s.notes[len(s.notes)-1], nil
}

// Replace atomically replaces the complete note list with copies of notes.
func (s *Scale) Replace(notes ...note.Note) error {
	if err := s.check(); err != nil {
		return err
	}
	if err := validNotes(notes); err != nil {
		return err
	}
	s.notes = owned(notes)
	return nil
}

// Set atomically replaces the note at an existing index.
func (s *Scale) Set(index int, n note.Note) error {
	if err := s.check(); err != nil {
		return err
	}
	if err := validNotes([]note.Note{n}); err != nil {
		return err
	}
	if err := s.index(index); err != nil {
		return err
	}
	ns := owned(s.notes)
	ns[index] = n
	s.notes = ns
	return nil
}

const maxNotes = int(^uint(0)>>1) / int(unsafe.Sizeof(note.Note{}))

func listSize(n, extra int) error {
	if extra > maxNotes-n {
		return ErrOutOfRange
	}
	return nil
}

// Insert inserts copies before index, which may equal Len.
func (s *Scale) Insert(index int, notes ...note.Note) error {
	if err := s.check(); err != nil {
		return err
	}
	if err := validNotes(notes); err != nil {
		return err
	}
	if index < 0 || index > len(s.notes) {
		return ErrIndex
	}
	if err := listSize(len(s.notes), len(notes)); err != nil {
		return err
	}
	ns := make([]note.Note, 0, len(s.notes)+len(notes))
	ns = append(ns, s.notes[:index]...)
	ns = append(ns, notes...)
	ns = append(ns, s.notes[index:]...)
	s.notes = ns
	return nil
}

// Append appends complete copies of the supplied notes.
func (s *Scale) Append(notes ...note.Note) error { return s.Insert(s.Len(), notes...) }

// Remove removes the note at an existing index.
func (s *Scale) Remove(index int) error {
	if err := s.index(index); err != nil {
		return err
	}
	return s.Delete(index, index+1)
}
func (s *Scale) bounds(begin, end int) error {
	if s == nil {
		return ErrUnset
	}
	if begin < 0 || end < begin || end > len(s.notes) {
		return ErrIndex
	}
	return nil
}

// Delete removes the half-open range [begin, end).
func (s *Scale) Delete(begin, end int) error {
	if err := s.bounds(begin, end); err != nil {
		return err
	}
	ns := make([]note.Note, 0, len(s.notes)-(end-begin))
	ns = append(ns, s.notes[:begin]...)
	s.notes = append(ns, s.notes[end:]...)
	return nil
}

// Clear replaces the list with a present, empty list.
func (s *Scale) Clear() error { return s.Replace() }

// Equal compares complete notes in order, ignoring their sources. Absent scales are unequal.
func (s *Scale) Equal(other *Scale) bool {
	return s != nil && other != nil && slices.EqualFunc(s.notes, other.notes, note.Note.Equal)
}

// Compare orders scales lexicographically by note records, with nil first.
func (s *Scale) Compare(other *Scale) int {
	if s == nil {
		if other == nil {
			return 0
		}
		return -1
	}
	if other == nil {
		return 1
	}
	return slices.CompareFunc(s.notes, other.notes, note.Note.Compare)
}
