package tempomap

import (
	"slices"

	"github.com/go-muse/musex/measure"
)

func (m Map) withRecords(op string, changes []Change, extensions []Extension, grid []measure.Grid) (Map, error) {
	candidate := Map{addressing: m.addressing, changes: changes, extensions: extensions}
	idx, err := candidate.buildIndex(grid)
	if err != nil {
		return Map{}, err
	}
	if candidate.addressing == PositionAddressing {
		candidate.index = idx
	}
	return candidate, nil
}

// WithChange adds or replaces a change selected by address and slot.
func (m Map) WithChange(change Change, grid ...measure.Grid) (Map, error) {
	const op = "Map.WithChange"
	if !m.IsSet() {
		return Map{}, simpleError(ErrUnset, op, "receiver")
	}
	if !change.IsSet() {
		return Map{}, simpleError(ErrUnset, op, "change")
	}
	if !change.slot.IsValid() || !change.kind.IsValid() || change.at.addressing != m.addressing {
		return Map{}, simpleError(ErrInvalidArgument, op, "invalid change or addressing mismatch")
	}
	if _, err := m.gridArgument(op, grid); err != nil {
		return Map{}, err
	}
	changes := slices.Clone(m.changes)
	initialAddress := changes[0].at
	found := -1
	for i, current := range changes {
		if current.at.Equal(change.at) && current.slot.Equal(change.slot) {
			found = i
			break
		}
	}

	switch change.slot {
	case SlotFirst:
		if change.kind != ChangeMain {
			return Map{}, simpleError(ErrNoInitialTempo, op, "initial slot requires a main tempo")
		}
		if !change.at.Equal(initialAddress) {
			return Map{}, simpleError(ErrInvalidAddress, op, "initial slot outside the initial address")
		}
		changes[0] = change
	case SlotSecond:
		if change.kind != ChangeGradual {
			return Map{}, simpleError(ErrDuplicateAddress, op, "second initial slot must be gradual")
		}
		if !change.at.Equal(initialAddress) {
			return Map{}, simpleError(ErrInvalidAddress, op, "second slot outside the initial address")
		}
		if found >= 0 {
			changes[found] = change
		} else {
			changes = slices.Insert(changes, 1, change)
		}
	case SlotSingle:
		if change.at.Equal(initialAddress) {
			return Map{}, simpleError(ErrDuplicateAddress, op, "single slot at the occupied initial address")
		}
		if found >= 0 {
			changes[found] = change
		} else {
			index := len(changes)
			for i, current := range changes {
				if change.at.Compare(current.at) < 0 {
					index = i
					break
				}
			}
			changes = slices.Insert(changes, index, change)
		}
	}
	return m.withRecords(op, changes, slices.Clone(m.extensions), grid)
}

// MustWithChange is like WithChange but panics with its error.
func (m Map) MustWithChange(change Change, grid ...measure.Grid) Map {
	return must(m.WithChange(change, grid...))
}

// WithoutChange removes a change selected by address and slot.
func (m Map) WithoutChange(at Address, slot Slot, grid ...measure.Grid) (Map, error) {
	const op = "Map.WithoutChange"
	if !m.IsSet() {
		return Map{}, simpleError(ErrUnset, op, "receiver")
	}
	if !at.IsSet() {
		return Map{}, simpleError(ErrUnset, op, "address")
	}
	if !slot.IsValid() || at.addressing != m.addressing {
		return Map{}, simpleError(ErrInvalidArgument, op, "invalid slot or addressing mismatch")
	}
	g, err := m.gridArgument(op, grid)
	if err != nil {
		return Map{}, err
	}
	if slot == SlotFirst {
		return Map{}, simpleError(ErrNoInitialTempo, op, "the initial change cannot be removed")
	}
	if _, err = addressPosition(op, at, m.addressing, g); err != nil {
		return Map{}, err
	}
	index := -1
	for i, current := range m.changes {
		if current.at.Equal(at) && current.slot.Equal(slot) {
			index = i
			break
		}
	}
	if index < 0 {
		return m, nil
	}
	changes := slices.Clone(m.changes)
	changes = slices.Delete(changes, index, index+1)
	return m.withRecords(op, changes, slices.Clone(m.extensions), grid)
}

// MustWithoutChange is like WithoutChange but panics with its error.
func (m Map) MustWithoutChange(at Address, slot Slot, grid ...measure.Grid) Map {
	return must(m.WithoutChange(at, slot, grid...))
}

// WithExtension adds or replaces the hold at its address.
func (m Map) WithExtension(extension Extension, grid ...measure.Grid) (Map, error) {
	const op = "Map.WithExtension"
	if !m.IsSet() {
		return Map{}, simpleError(ErrUnset, op, "receiver")
	}
	if !extension.IsSet() {
		return Map{}, simpleError(ErrUnset, op, "extension")
	}
	if extension.at.addressing != m.addressing {
		return Map{}, simpleError(ErrInvalidArgument, op, "addressing mismatch")
	}
	g, err := m.gridArgument(op, grid)
	if err != nil {
		return Map{}, err
	}
	if _, err = addressPosition(op, extension.at, m.addressing, g); err != nil {
		return Map{}, err
	}
	extensions := slices.Clone(m.extensions)
	found := -1
	for i, current := range extensions {
		if current.at.Equal(extension.at) {
			found = i
			break
		}
	}
	if found >= 0 {
		extensions[found] = extension
	} else {
		index := len(extensions)
		for i, current := range extensions {
			if extension.at.Compare(current.at) < 0 {
				index = i
				break
			}
		}
		extensions = slices.Insert(extensions, index, extension)
	}
	return m.withRecords(op, slices.Clone(m.changes), extensions, grid)
}

// MustWithExtension is like WithExtension but panics with its error.
func (m Map) MustWithExtension(extension Extension, grid ...measure.Grid) Map {
	return must(m.WithExtension(extension, grid...))
}

// WithoutExtension removes the hold at at; a missing hold is not an error.
func (m Map) WithoutExtension(at Address, grid ...measure.Grid) (Map, error) {
	const op = "Map.WithoutExtension"
	if !m.IsSet() {
		return Map{}, simpleError(ErrUnset, op, "receiver")
	}
	if !at.IsSet() {
		return Map{}, simpleError(ErrUnset, op, "address")
	}
	if at.addressing != m.addressing {
		return Map{}, simpleError(ErrInvalidArgument, op, "addressing mismatch")
	}
	g, err := m.gridArgument(op, grid)
	if err != nil {
		return Map{}, err
	}
	if _, err = addressPosition(op, at, m.addressing, g); err != nil {
		return Map{}, err
	}
	index := -1
	for i, current := range m.extensions {
		if current.at.Equal(at) {
			index = i
			break
		}
	}
	if index < 0 {
		return m, nil
	}
	extensions := slices.Clone(m.extensions)
	extensions = slices.Delete(extensions, index, index+1)
	return m.withRecords(op, slices.Clone(m.changes), extensions, grid)
}

// MustWithoutExtension is like WithoutExtension but panics with its error.
func (m Map) MustWithoutExtension(at Address, grid ...measure.Grid) Map {
	return must(m.WithoutExtension(at, grid...))
}
