package scale

import (
	"iter"

	"github.com/go-muse/musex/note"
)

// wrap adds a signed distance without overflowing the intermediate sum.
func wrap(index, distance, n int) int {
	d := distance % n
	if d < 0 {
		d += n
	}
	if index >= n-d {
		return index - (n - d)
	}
	return index + d
}

// Move moves through the stored list by a signed distance, wrapping at its ends.
func (s *Scale) Move(index, distance int) (Element, error) {
	if err := s.elementIndex(index); err != nil {
		return Element{}, err
	}
	i := wrap(index, distance, len(s.notes))
	return Element{Index: i, Note: s.notes[i]}, nil
}

// Next returns the next stored occurrence, wrapping after the last.
func (s *Scale) Next(index int) (Element, error) { return s.Move(index, 1) }

// Previous returns the preceding stored occurrence, wrapping before the first.
func (s *Scale) Previous(index int) (Element, error) { return s.Move(index, -1) }

// DistanceForward returns the forward circular distance between existing indexes.
func (s *Scale) DistanceForward(from, to int) (int, error) {
	if err := s.elementIndex(from); err != nil {
		return 0, err
	}
	if err := s.index(to); err != nil {
		return 0, err
	}
	return wrap(to, -from, len(s.notes)), nil
}

// DistanceBackward returns the backward circular distance between existing indexes.
func (s *Scale) DistanceBackward(from, to int) (int, error) { return s.DistanceForward(to, from) }

// MirrorIndex reflects an existing index around an existing axis in the circle.
func (s *Scale) MirrorIndex(index, axis int) (Element, error) {
	if err := s.elementIndex(index); err != nil {
		return Element{}, err
	}
	if err := s.index(axis); err != nil {
		return Element{}, err
	}
	return s.Move(axis, axis-index)
}
func (s *Scale) walkCheck(start, count int, direction Direction) error {
	if err := s.check(); err != nil {
		return err
	}
	if count < 0 || !direction.IsValid() {
		return ErrInvalidArgument
	}
	if count == 0 {
		return nil
	}
	return s.elementIndex(start)
}

// Iter captures a snapshot and lazily yields count notes from start, inclusively.
func (s *Scale) Iter(start, count int, direction Direction) (iter.Seq[note.Note], error) {
	if err := s.walkCheck(start, count, direction); err != nil {
		return nil, err
	}
	snapshot := s.Notes()
	return func(yield func(note.Note) bool) {
		index := start
		for range count {
			if !yield(snapshot[index]) {
				return
			}
			index = wrap(index, int(direction), len(snapshot))
		}
	}, nil
}

// Round captures a snapshot and yields one complete traversal.
func (s *Scale) Round(start int, direction Direction) (iter.Seq[note.Note], error) {
	return s.Iter(start, s.Len(), direction)
}

// Cycle captures a nonempty snapshot and repeats it until the consumer stops.
func (s *Scale) Cycle(start int, direction Direction) (iter.Seq[note.Note], error) {
	if err := s.walkCheck(start, 1, direction); err != nil {
		return nil, err
	}
	snapshot := s.Notes()
	return func(yield func(note.Note) bool) {
		index := start
		for yield(snapshot[index]) {
			index = wrap(index, int(direction), len(snapshot))
		}
	}, nil
}

// Take copies count notes from a circular traversal, including start.
func (s *Scale) Take(start, count int, direction Direction) (*Scale, error) {
	seq, err := s.Iter(start, count, direction)
	if err != nil {
		return nil, err
	}
	if count > maxNotes {
		return nil, ErrOutOfRange
	}
	ns := make([]note.Note, 0, count)
	for n := range seq {
		ns = append(ns, n)
	}
	return &Scale{notes: ns}, nil
}

// OneRound copies one full traversal of the stored list.
func (s *Scale) OneRound(start int, direction Direction) (*Scale, error) {
	return s.Take(start, s.Len(), direction)
}
func (s *Scale) resolve(from Selection, p prepared) (int, error) {
	if from.Index != nil {
		index := *from.Index
		if err := s.index(index); err != nil {
			return 0, err
		}
		same, err := p.equal(s.notes[index], from.Note)
		if err != nil {
			return 0, err
		}
		if !same {
			return 0, ErrNotFound
		}
		return index, nil
	}
	result := -1
	for i, n := range s.notes {
		same, err := p.equal(n, from.Note)
		if err != nil {
			return 0, err
		}
		if same {
			if result >= 0 {
				return 0, ErrAmbiguous
			}
			result = i
		}
	}
	if result < 0 {
		return 0, ErrNotFound
	}
	return result, nil
}

// MoveFrom resolves an occurrence and moves by a signed circular distance.
func (s *Scale) MoveFrom(from Selection, distance int, by ...Comparison) (Element, error) {
	if err := s.check(); err != nil {
		return Element{}, err
	}
	p, err := prepare(by, s.notes, []note.Note{from.Note})
	if err != nil {
		return Element{}, err
	}
	index, err := s.resolve(from, p)
	if err != nil {
		return Element{}, err
	}
	return s.Move(index, distance)
}
func selections(ns []note.Note) []Selection {
	result := make([]Selection, len(ns))
	for i, n := range ns {
		result[i] = Selection{Note: n}
	}
	return result
}

// ShiftMelody selects notes at a circular distance from each melody occurrence.
func (s *Scale) ShiftMelody(notes []note.Note, distance int, by ...Comparison) (*Scale, error) {
	return s.ShiftSelected(selections(notes), distance, by...)
}

// MirrorMelody reflects each melody occurrence around an existing axis.
func (s *Scale) MirrorMelody(notes []note.Note, axis int, by ...Comparison) (*Scale, error) {
	return s.MirrorSelected(selections(notes), axis, by...)
}

// ShiftSelected shifts melody occurrences with explicit duplicate selection.
func (s *Scale) ShiftSelected(notes []Selection, distance int, by ...Comparison) (*Scale, error) {
	return s.melody(notes, distance, false, by)
}

// MirrorSelected reflects melody occurrences with explicit duplicate selection.
func (s *Scale) MirrorSelected(notes []Selection, axis int, by ...Comparison) (*Scale, error) {
	return s.melody(notes, axis, true, by)
}
func (s *Scale) melody(notes []Selection, amount int, mirror bool, by []Comparison) (*Scale, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	ns := make([]note.Note, len(notes))
	selected := make([]Selection, len(notes))
	for i, n := range notes {
		selected[i] = n
		ns[i] = n.Note
		if n.Index != nil {
			index := *n.Index
			selected[i].Index = &index
		}
	}
	if err := validNotes(ns); err != nil {
		return nil, err
	}
	if len(by) > 1 {
		return nil, ErrInvalidArgument
	}
	if mirror {
		if err := s.elementIndex(amount); err != nil {
			return nil, err
		}
	}
	for _, n := range selected {
		if n.Index != nil {
			if err := s.index(*n.Index); err != nil {
				return nil, err
			}
		}
	}
	if len(ns) == 0 {
		return New()
	}
	p, err := prepare(by, s.notes, ns)
	if err != nil {
		return nil, err
	}
	result := make([]note.Note, len(ns))
	for i, n := range selected {
		index, err := s.resolve(n, p)
		if err != nil {
			return nil, err
		}
		var e Element
		if mirror {
			e, err = s.MirrorIndex(index, amount)
		} else {
			e, err = s.Move(index, amount)
		}
		if err != nil {
			return nil, err
		}
		result[i] = e.Note
	}
	return New(result...)
}
