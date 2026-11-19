package scale

import (
	"errors"
	"iter"
	"slices"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// Rotate copies a rotation from one existing index. With no index it returns a copy.
func (s *Scale) Rotate(first ...int) (*Scale, error) {
	if s == nil {
		return nil, ErrUnset
	}
	if len(first) > 1 {
		return nil, ErrInvalidArgument
	}
	if len(first) == 0 {
		return s.Clone(), nil
	}
	if err := s.elementIndex(first[0]); err != nil {
		return nil, err
	}
	return s.OneRound(first[0], Forward)
}

// Reverse returns an independent reversed list.
func (s *Scale) Reverse() (*Scale, error) {
	if s == nil {
		return nil, ErrUnset
	}
	result := s.Clone()
	slices.Reverse(result.notes)
	return result, nil
}

// Slice copies the finite half-open range [begin, end).
func (s *Scale) Slice(begin, end int) (*Scale, error) {
	if err := s.bounds(begin, end); err != nil {
		return nil, err
	}
	return New(s.notes[begin:end]...)
}

// Repeat copies times repetitions without changing note octaves.
func (s *Scale) Repeat(times int) (*Scale, error) {
	if s == nil {
		return nil, ErrUnset
	}
	if times < 0 {
		return nil, ErrInvalidArgument
	}
	if s.Len() != 0 && times > maxNotes/s.Len() {
		return nil, ErrOutOfRange
	}
	if s.Len() == 0 || times == 0 {
		return New()
	}
	ns := make(note.Notes, 0, s.Len()*times)
	for range times {
		ns = append(ns, s.notes...)
	}
	return &Scale{notes: ns}, nil
}
func (s *Scale) mapNotes(change func(note.Note) (note.Note, error)) (*Scale, error) {
	if s == nil {
		return nil, ErrUnset
	}
	ns := make(note.Notes, len(s.notes))
	for i, n := range s.notes {
		value, err := change(n)
		if err != nil {
			return nil, err
		}
		ns[i] = value
	}
	return &Scale{notes: ns}, nil
}

// WithoutOctaves copies the notes with their octave information removed.
func (s *Scale) WithoutOctaves() (*Scale, error) {
	return s.mapNotes(func(n note.Note) (note.Note, error) { err := n.RemoveOctave(); return n, err })
}

// ShiftOctaves copies the notes after shifting every octave by count.
func (s *Scale) ShiftOctaves(count int) (*Scale, error) {
	return s.mapNotes(func(n note.Note) (note.Note, error) { err := n.Shift(count); return n, err })
}

// Transpose transposes each complete note by a written interval in its own context.
func (s *Scale) Transpose(i interval.Interval) (*Scale, error) {
	if s == nil {
		return nil, ErrUnset
	}
	if !i.IsValid() {
		return nil, errors.Join(ErrUnset, interval.ErrUnset)
	}
	return s.mapNotes(func(n note.Note) (note.Note, error) { return interval.TransposeNote(n, i) })
}

// TransposeIn transposes each note using an explicit musical system.
func (s *Scale) TransposeIn(sys *system.System, i interval.Interval) (*Scale, error) {
	if s == nil {
		return nil, ErrUnset
	}
	if !sys.IsValid() {
		return nil, errors.Join(ErrUnset, system.ErrInvalid)
	}
	if !i.IsValid() {
		return nil, errors.Join(ErrUnset, interval.ErrUnset)
	}
	return s.mapNotes(func(n note.Note) (note.Note, error) { return interval.TransposeNoteIn(sys, n, i) })
}

// TransposeTo transposes every note by the interval from the first record to first.
func (s *Scale) TransposeTo(first note.Note, direction interval.Direction) (*Scale, error) {
	if s == nil {
		return nil, ErrUnset
	}
	if !first.IsValid() {
		return nil, errors.Join(ErrUnset, note.ErrUnset)
	}
	old, err := s.First()
	if err != nil {
		return nil, err
	}
	i, err := interval.BetweenNotes(old, first, direction)
	if err != nil {
		return nil, err
	}
	return s.Transpose(i)
}

// TransposeToIn transposes to a first record using an explicit system.
func (s *Scale) TransposeToIn(sys *system.System, first note.Note, direction interval.Direction) (*Scale, error) {
	if s == nil {
		return nil, ErrUnset
	}
	if !sys.IsValid() {
		return nil, errors.Join(ErrUnset, system.ErrInvalid)
	}
	if !first.IsValid() {
		return nil, errors.Join(ErrUnset, note.ErrUnset)
	}
	old, err := s.First()
	if err != nil {
		return nil, err
	}
	i, err := interval.BetweenNotesIn(sys, old, first, direction)
	if err != nil {
		return nil, err
	}
	return s.TransposeIn(sys, i)
}

// Enharmonics returns repeatable alternatives for one note, preserving its other parts.
func (s *Scale) Enharmonics(index int, limit alteration.Alteration) (iter.Seq[note.Note], error) {
	if err := s.elementIndex(index); err != nil {
		return nil, err
	}
	return s.notes[index].Enharmonics(limit)
}

// EnharmonicsIn returns alternatives for one note using an explicit settings source.
func (s *Scale) EnharmonicsIn(src settings.Source, index int, limit alteration.Alteration) (iter.Seq[note.Note], error) {
	if s == nil {
		return nil, ErrUnset
	}
	if src == nil {
		return nil, ErrUnset
	}
	if err := s.elementIndex(index); err != nil {
		return nil, err
	}
	return s.notes[index].EnharmonicsIn(src, limit)
}

// WithEnharmonic copies the scale with one position-equivalent spelling replacement.
func (s *Scale) WithEnharmonic(index int, name spelling.Spelling) (*Scale, error) {
	return s.respell(nil, index, name, false)
}

// WithEnharmonicIn validates a spelling replacement in an explicit settings context.
func (s *Scale) WithEnharmonicIn(src settings.Source, index int, name spelling.Spelling) (*Scale, error) {
	return s.respell(src, index, name, true)
}
func (s *Scale) respell(src settings.Source, index int, name spelling.Spelling, explicit bool) (*Scale, error) {
	if s == nil {
		return nil, ErrUnset
	}
	if explicit && src == nil {
		return nil, ErrUnset
	}
	if !name.IsValid() {
		return nil, errors.Join(ErrUnset, letter.ErrInvalid)
	}
	if err := s.elementIndex(index); err != nil {
		return nil, err
	}
	old := s.notes[index]
	if !explicit {
		src = old.Source()
		if src == nil {
			return nil, note.ErrNoSource
		}
	}
	set := src.Settings()
	if !set.IsValid() {
		return nil, errors.Join(note.ErrUnsetSettings, settings.ErrUnset)
	}
	sys := set.System()
	if _, err := sys.Class(old.Spelling()); err != nil {
		return nil, errors.Join(note.ErrInadmissible, err)
	}
	if _, err := sys.Class(name); err != nil {
		return nil, errors.Join(note.ErrInadmissible, err)
	}
	next := old
	if err := next.Respell(name); err != nil {
		return nil, err
	}
	p := prepared{sys: sys, comparison: Comparison{kind: comparePositionIn, sys: sys}}
	same, err := p.equal(old, next)
	if err != nil {
		return nil, err
	}
	if !same {
		return nil, ErrNotEquivalent
	}
	result := s.Clone()
	result.notes[index] = next
	return result, nil
}
