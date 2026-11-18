package scale

import (
	"errors"

	"github.com/go-muse/musex/frequency"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
)

type comparisonKind uint8

const (
	compareNote comparisonKind = iota
	compareRecord
	compareName
	comparePosition
	comparePositionIn
	compareSound
)

// Comparison is an immutable strict equality rule. The zero value is ByNote.
type Comparison struct {
	kind   comparisonKind
	sys    *system.System
	tuning *tuning.Tuning
	_      [0]func()
}

// ByNote compares spelling, octave presence and value, and cents.
func ByNote() Comparison { return Comparison{} }

// ByRecord compares records while ignoring cents and sources.
func ByRecord() Comparison { return Comparison{kind: compareRecord} }

// ByName compares spellings while ignoring octaves, cents, and sources.
func ByName() Comparison { return Comparison{kind: compareName} }

// ByPosition compares positions or classes in compatible systems from note sources.
func ByPosition() Comparison { return Comparison{kind: comparePosition} }

// ByPositionIn compares positions or classes in one explicit system.
func ByPositionIn(sys *system.System) (Comparison, error) {
	if !sys.IsValid() {
		return Comparison{}, errors.Join(ErrUnset, system.ErrInvalid)
	}
	return Comparison{kind: comparePositionIn, sys: sys}, nil
}

// BySoundIn compares exact sound in one explicit tuning. All notes need octaves.
func BySoundIn(t *tuning.Tuning) (Comparison, error) {
	if !t.IsValid() {
		return Comparison{}, errors.Join(ErrUnset, tuning.ErrUnset)
	}
	return Comparison{kind: compareSound, tuning: t}, nil
}

type prepared struct {
	comparison Comparison
	sys        *system.System
}

func prepare(by []Comparison, groups ...[]note.Note) (prepared, error) {
	p := prepared{}
	for _, ns := range groups {
		if err := validNotes(ns); err != nil {
			return p, err
		}
	}
	if len(by) > 1 {
		return p, ErrInvalidArgument
	}
	if len(by) == 1 {
		p.comparison = by[0]
	}
	c := p.comparison
	if c.kind < comparePosition {
		return p, nil
	}
	var all []note.Note
	for _, ns := range groups {
		all = append(all, ns...)
	}
	if len(all) == 0 {
		return p, nil
	}
	hasOctave := all[0].HasOctave()
	for _, n := range all {
		if n.HasOctave() != hasOctave || c.kind == compareSound && !n.HasOctave() {
			return p, note.ErrNoOctave
		}
	}
	p.sys = c.sys
	if c.kind == compareSound {
		p.sys = c.tuning.System()
	}
	var sources []settings.Source
	var sets []*settings.Settings
	systems := make([]*system.System, len(all))
	if c.kind == comparePosition {
		for _, n := range all {
			if !n.IsBound() {
				return p, note.ErrNoSource
			}
		}
		for i, n := range all {
			var set *settings.Settings
			for j, source := range sources {
				if settings.SameSource(source, n.Source()) {
					set = sets[j]
					break
				}
			}
			if set == nil {
				set = n.Source().Settings()
				if !set.IsValid() {
					return p, errors.Join(note.ErrUnsetSettings, settings.ErrUnset)
				}
				sources = append(sources, n.Source())
				sets = append(sets, set)
			}
			systems[i] = set.System()
		}
		p.sys = systems[0]
	} else {
		for i := range systems {
			systems[i] = p.sys
		}
	}
	for i, n := range all {
		_, err := systems[i].Class(n.Spelling())
		if err != nil {
			if errors.Is(err, system.ErrInadmissible) {
				return p, errors.Join(note.ErrInadmissible, err)
			}
			return p, err
		}
	}
	for _, sys := range systems {
		if !p.sys.Equal(sys) {
			return p, note.ErrDifferentSystems
		}
	}
	return p, nil
}
func (p prepared) equal(a, b note.Note) (bool, error) {
	switch p.comparison.kind {
	case compareNote:
		return a.Equal(b), nil
	case compareRecord:
		return a.SameRecord(b), nil
	case compareName:
		return a.SameSpelling(b), nil
	case comparePosition, comparePositionIn:
		if a.HasOctave() {
			return p.sys.Equivalent(a.MustPitch(), b.MustPitch())
		}
		x, err := p.sys.Class(a.Spelling())
		if err != nil {
			return false, err
		}
		y, err := p.sys.Class(b.Spelling())
		if err != nil {
			return false, err
		}
		return x.Equal(y), nil
	case compareSound:
		return a.SoundsSameInTuning(p.comparison.tuning, b)
	}
	return false, ErrInvalidArgument
}

// Find returns every matching zero-based index in ascending order.
func (s *Scale) Find(n note.Note, by ...Comparison) ([]int, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	p, err := prepare(by, s.notes, []note.Note{n})
	if err != nil {
		return nil, err
	}
	result := make([]int, 0)
	for i, value := range s.notes {
		same, err := p.equal(value, n)
		if err != nil {
			return nil, err
		}
		if same {
			result = append(result, i)
		}
	}
	return result, nil
}

// Contains reports whether at least one note matches under the comparison.
func (s *Scale) Contains(n note.Note, by ...Comparison) (bool, error) {
	indices, err := s.Find(n, by...)
	return len(indices) > 0, err
}

// Count returns the number of matching occurrences.
func (s *Scale) Count(n note.Note, by ...Comparison) (int, error) {
	indices, err := s.Find(n, by...)
	return len(indices), err
}

// FindFirst returns the first matching index and whether one exists.
func (s *Scale) FindFirst(n note.Note, by ...Comparison) (int, bool, error) {
	indices, err := s.Find(n, by...)
	if err != nil {
		return 0, false, err
	}
	if len(indices) == 0 {
		return 0, false, nil
	}
	return indices[0], true, nil
}

// FindLast returns the last matching index and whether one exists.
func (s *Scale) FindLast(n note.Note, by ...Comparison) (int, bool, error) {
	indices, err := s.Find(n, by...)
	if err != nil {
		return 0, false, err
	}
	if len(indices) == 0 {
		return 0, false, nil
	}
	return indices[len(indices)-1], true, nil
}

// EqualBy compares lists in order after validating all notes for the comparison.
func (s *Scale) EqualBy(other *Scale, by Comparison) (bool, error) {
	if s == nil || other == nil {
		return false, ErrUnset
	}
	p, err := prepare([]Comparison{by}, s.notes, other.notes)
	if err != nil {
		return false, err
	}
	if len(s.notes) != len(other.notes) {
		return false, nil
	}
	for i, n := range s.notes {
		same, err := p.equal(n, other.notes[i])
		if err != nil {
			return false, err
		}
		if !same {
			return false, nil
		}
	}
	return true, nil
}

// FindSound performs pairwise sound search in the notes' own settings.
func (s *Scale) FindSound(n note.Note, c frequency.Criterion) ([]int, error) {
	return s.findSound(nil, n, c, false)
}

// FindSoundInTuning performs pairwise sound search in one explicit tuning.
func (s *Scale) FindSoundInTuning(t *tuning.Tuning, n note.Note, c frequency.Criterion) ([]int, error) {
	return s.findSound(t, n, c, true)
}
func (s *Scale) findSound(t *tuning.Tuning, n note.Note, c frequency.Criterion, explicit bool) ([]int, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	ns := append(s.Notes(), n)
	if err := validNotes(ns); err != nil {
		return nil, err
	}
	if explicit && !t.IsValid() {
		return nil, errors.Join(ErrUnset, tuning.ErrUnset)
	}
	if !c.IsValid() {
		return nil, errors.Join(ErrUnset, frequency.ErrUnset)
	}
	for _, value := range ns {
		if !value.HasOctave() {
			return nil, note.ErrNoOctave
		}
	}
	// Freeze each distinct source while retaining the pairwise sounding rule.
	var sources []settings.Source
	var sets []*settings.Settings
	if !explicit {
		for _, value := range ns {
			if !value.IsBound() {
				return nil, note.ErrNoSource
			}
		}
		for i, value := range ns {
			var set *settings.Settings
			for j, source := range sources {
				if settings.SameSource(source, value.Source()) {
					set = sets[j]
					break
				}
			}
			if set == nil {
				set = value.Source().Settings()
				if !set.IsValid() {
					return nil, errors.Join(note.ErrUnsetSettings, settings.ErrUnset)
				}
				sources = append(sources, value.Source())
				sets = append(sets, set)
			}
			if _, err := set.System().Class(value.Spelling()); err != nil {
				return nil, errors.Join(note.ErrInadmissible, err)
			}
			frozen, err := note.FromPitch(value.MustPitch(), value.Cents(), set)
			if err != nil {
				return nil, err
			}
			ns[i] = frozen
		}
	} else {
		for _, value := range ns {
			if _, err := value.SoundsSameByInTuning(t, value, c); err != nil {
				return nil, err
			}
		}
	}
	query := ns[len(ns)-1]
	result := make([]int, 0)
	for i, value := range ns[:len(ns)-1] {
		var same bool
		var err error
		if explicit {
			same, err = value.SoundsSameByInTuning(t, query, c)
		} else {
			same, err = value.SoundsSameBy(query, c)
		}
		if err != nil {
			return nil, err
		}
		if same {
			result = append(result, i)
		}
	}
	return result, nil
}
