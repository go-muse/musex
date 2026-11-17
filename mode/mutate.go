package mode

import (
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/template"
)

func (m *Mode) checkReceiver() error {
	if m == nil || !m.IsValid() {
		return ErrUnset
	}
	return nil
}

func (m *Mode) rebuild(makeCandidate func() (Mode, error)) error {
	if err := m.checkReceiver(); err != nil {
		return err
	}
	candidate, err := makeCandidate()
	if err != nil {
		return err
	}
	*m = candidate
	return nil
}

// Rebuild replaces the structure and clears the origin.
func (m *Mode) Rebuild(t template.Template) error {
	return m.rebuild(func() (Mode, error) { return FromTemplate(t) })
}

// RebuildFromEntry replaces the structure and its origin.
func (m *Mode) RebuildFromEntry(e template.Entry) error {
	return m.rebuild(func() (Mode, error) { return FromEntry(e) })
}

// RebuildFromCode replaces m using an exact code in the built-in catalog.
func (m *Mode) RebuildFromCode(code string) error {
	return m.rebuild(func() (Mode, error) { return FromCode(code) })
}

// RebuildFromCodeIn replaces m using an exact code in catalog.
func (m *Mode) RebuildFromCodeIn(catalog *template.Catalog, code string) error {
	return m.rebuild(func() (Mode, error) { return FromCodeIn(catalog, code) })
}

// RebuildFromName replaces m using a normalized name in the built-in catalog.
func (m *Mode) RebuildFromName(name string) error {
	return m.rebuild(func() (Mode, error) { return FromName(name) })
}

// RebuildFromNameIn replaces m using a normalized name in catalog.
func (m *Mode) RebuildFromNameIn(catalog *template.Catalog, name string) error {
	return m.rebuild(func() (Mode, error) { return FromNameIn(catalog, name) })
}

// Rotate makes first the new first degree and preserves the origin.
func (m *Mode) Rotate(first int) error {
	return m.rebuild(func() (Mode, error) { return FromDegree(*m, first) })
}

func (m *Mode) replaceStructure(t template.Template, err error) error {
	if err != nil {
		return templateError(err)
	}
	candidate, err := buildMode(t, m.origin, m.name)
	if err != nil {
		return err
	}
	*m = candidate
	return nil
}

func (m *Mode) replaceAdjacent(adjacent []interval.Interval) error {
	if m.IsOpen() {
		return m.replaceStructure(template.OpenFromIntervals(adjacent))
	}
	return m.replaceStructure(template.ClosedFromIntervals(adjacent))
}

func (m *Mode) replacePrefixes(prefixes []interval.Interval) error {
	if m.IsOpen() {
		return m.replaceStructure(template.OpenFromPlaceIntervals(prefixes))
	}
	return m.replaceStructure(template.ClosedFromPlaceIntervals(prefixes, m.cycle))
}

// SetDegreeInterval changes the interval of one degree without moving the
// other degrees. The interval must remain strictly between its neighbors.
func (m *Mode) SetDegreeInterval(number int, value interval.Interval) error {
	if err := m.checkReceiver(); err != nil {
		return err
	}
	if !value.IsValid() {
		return ErrUnset
	}
	if number < 2 || number > m.Len() {
		return ErrInvalidArgument
	}
	prefixes := m.DegreeIntervals()
	prefixes[number-1] = value
	return m.replacePrefixes(prefixes)
}

// AddDegreeInterval inserts a new degree at its ordered position.
func (m *Mode) AddDegreeInterval(value interval.Interval) error {
	if err := m.checkReceiver(); err != nil {
		return err
	}
	if !value.IsValid() {
		return ErrUnset
	}
	if m.Len() == MaxDegrees {
		return ErrInvalidArgument
	}
	i, found := slices.BinarySearchFunc(m.prefixes, value, interval.Interval.Compare)
	if found {
		return ErrInvalidArgument
	}
	if i == 0 || (m.IsClosed() && value.Compare(m.cycle) >= 0) {
		return ErrInvalidStructure
	}
	prefixes := slices.Insert(m.DegreeIntervals(), i, value)
	return m.replacePrefixes(prefixes)
}

// RemoveDegree removes one degree, retaining the cycle of a closed mode.
func (m *Mode) RemoveDegree(number int) error {
	if err := m.checkReceiver(); err != nil {
		return err
	}
	if err := m.checkDegrees(number); err != nil {
		return err
	}
	if m.Len() == 1 || (m.IsOpen() && m.Len() == 2) {
		return ErrInvalidArgument
	}

	adjacent := m.AdjacentIntervals()
	if m.IsOpen() {
		switch number {
		case 1:
			return m.replaceAdjacent(adjacent[1:])
		case m.Len():
			return m.replaceAdjacent(adjacent[:len(adjacent)-1])
		}
	}
	if number == 1 {
		// The old ending and beginning now form the new closing interval.
		joined, err := adjacent[len(adjacent)-1].Add(adjacent[0])
		if err != nil {
			return positionError("RemoveDegree", err)
		}
		adjacent[len(adjacent)-1] = joined
		return m.replaceAdjacent(adjacent[1:])
	}
	left := number - 2
	joined, err := adjacent[left].Add(adjacent[left+1])
	if err != nil {
		return positionError("RemoveDegree", err)
	}
	adjacent[left] = joined
	return m.replaceAdjacent(slices.Delete(adjacent, left+1, left+2))
}

// Mirror reverses the order of the adjacent intervals.
func (m *Mode) Mirror() error {
	if err := m.checkReceiver(); err != nil {
		return err
	}
	adjacent := m.AdjacentIntervals()
	slices.Reverse(adjacent)
	return m.replaceAdjacent(adjacent)
}

// Repeat repeats the structure of a closed mode times times.
func (m *Mode) Repeat(times int) error {
	if err := m.checkReceiver(); err != nil {
		return err
	}
	if times < 1 || times > MaxDegrees/m.Len() {
		return ErrInvalidArgument
	}
	if err := m.checkClosed(); err != nil {
		return err
	}
	return m.replaceAdjacent(slices.Repeat(m.AdjacentIntervals(), times))
}

// ReduceToMinimalPart replaces the structure by its shortest repeating part.
func (m *Mode) ReduceToMinimalPart() error {
	if err := m.checkReceiver(); err != nil {
		return err
	}
	part, _, err := m.MinimalPart()
	if err != nil {
		return err
	}
	return m.replaceAdjacent(part)
}
