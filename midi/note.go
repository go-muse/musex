package midi

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
)

// FromNote returns the MIDI number of n using one snapshot of its own settings source.
// The note and its source are not changed.
func FromNote(n note.Note) (Number, error) {
	if !n.IsValid() {
		return Number{}, fmt.Errorf("midi: from note: %w: %w", ErrUnset, note.ErrUnset)
	}
	if !n.HasOctave() {
		return Number{}, fmt.Errorf("midi: from note: %w: %w", ErrNoOctave, note.ErrNoOctave)
	}
	if n.Source() == nil {
		return Number{}, fmt.Errorf("midi: from note: %w: %w", ErrNoSource, note.ErrNoSource)
	}

	p, err := n.Pitch()
	if err != nil {
		return Number{}, fmt.Errorf("midi: from note: %w: %w", ErrNoOctave, err)
	}

	snapshot, err := n.Settings()
	if err != nil {
		return Number{}, fromOwnSettingsError(err)
	}
	if snapshot == nil || !snapshot.IsValid() {
		return Number{}, fmt.Errorf("midi: from note: %w: %w: %w", ErrUnsetSettings, settings.ErrUnset, note.ErrUnsetSettings)
	}

	return FromPitch(p, snapshot.System())
}

// MustFromNote calls FromNote and panics with the same error if it fails.
func MustFromNote(n note.Note) Number {
	number, err := FromNote(n)
	if err != nil {
		panic(err)
	}

	return number
}

// FromNoteIn returns the MIDI number of n using one snapshot of the explicit
// source src. It never reads or changes n's own source or its binding.
func FromNoteIn(src settings.Source, n note.Note) (Number, error) {
	if !n.IsValid() {
		return Number{}, fmt.Errorf("midi: from note in: %w: %w", ErrUnset, note.ErrUnset)
	}
	if src == nil {
		return Number{}, fmt.Errorf("midi: from note in: source: %w", ErrUnset)
	}
	if !n.HasOctave() {
		return Number{}, fmt.Errorf("midi: from note in: %w: %w", ErrNoOctave, note.ErrNoOctave)
	}

	snapshot := src.Settings()
	if snapshot == nil || !snapshot.IsValid() {
		return Number{}, fmt.Errorf("midi: from note in: %w: %w", ErrUnsetSettings, settings.ErrUnset)
	}

	p, err := n.Pitch()
	if err != nil {
		return Number{}, fmt.Errorf("midi: from note in: %w: %w", ErrNoOctave, err)
	}

	return FromPitch(p, snapshot.System())
}

// MustFromNoteIn calls FromNoteIn and panics with the same error if it fails.
func MustFromNoteIn(src settings.Source, n note.Note) Number {
	number, err := FromNoteIn(src, n)
	if err != nil {
		panic(err)
	}

	return number
}

// Note returns a new note for n, bound to src, with spelling selected by rule
// and correction c. The zero cents value means no correction.
//
// The source is read once here and once more by note.FromPitch. The caller must
// synchronize any concurrent replacement of its settings with the whole Note
// operation so that both reads return equal snapshots. An immutable *settings.Settings
// satisfies this precondition without additional synchronization.
func (n Number) Note(src settings.Source, rule system.SpellingRule, c cents.Cents) (note.Note, error) {
	if !n.IsValid() {
		return note.Note{}, fmt.Errorf("midi: note: %w", ErrInvalidNumber)
	}
	if src == nil {
		return note.Note{}, fmt.Errorf("midi: note: source: %w", ErrUnset)
	}

	snapshot := src.Settings()
	if snapshot == nil || !snapshot.IsValid() {
		return note.Note{}, fmt.Errorf("midi: note: %w: %w", ErrUnsetSettings, settings.ErrUnset)
	}

	p, err := n.Pitch(snapshot.System(), rule)
	if err != nil {
		return note.Note{}, err
	}

	created, err := note.FromPitch(p, c, src)
	if err != nil {
		return note.Note{}, fmt.Errorf("midi: note: %w", err)
	}

	return created, nil
}

// MustNote calls Note and panics with the same error if it fails.
func (n Number) MustNote(src settings.Source, rule system.SpellingRule, c cents.Cents) note.Note {
	created, err := n.Note(src, rule, c)
	if err != nil {
		panic(err)
	}

	return created
}

func fromOwnSettingsError(err error) error {
	switch {
	case errors.Is(err, note.ErrNoSource):
		return fmt.Errorf("midi: from note: %w: %w", ErrNoSource, err)
	case errors.Is(err, note.ErrUnsetSettings):
		return fmt.Errorf("midi: from note: %w: %w: %w", ErrUnsetSettings, settings.ErrUnset, err)
	case errors.Is(err, note.ErrUnset):
		return fmt.Errorf("midi: from note: %w: %w", ErrUnset, err)
	default:
		return fmt.Errorf("midi: from note: settings: %w", err)
	}
}
