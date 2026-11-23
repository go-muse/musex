package note

import (
	"cmp"
	"slices"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/tuning"
)

// The sorting functions return a new slice and never change the passed slice
// or its notes. The sort is stable: notes that compare as equal keep their
// order. An empty or nil slice gives an empty, non-nil slice without an error
// and without reading any source; the In and InTuning forms then check only
// that the explicit context is set.
//
// sign is +1 for an ascending order and -1 for a descending one; it reverses
// the comparison, not the result, so equal keys keep their input order in
// both directions.

func sortedByPosition(op string, c opContext, notes Notes, sign int) (Notes, error) {
	if len(notes) == 0 {
		if err := c.check(op); err != nil {
			return nil, err
		}

		return Notes{}, nil
	}

	sys, err := position(op, c, notes, false)
	if err != nil {
		return nil, err
	}

	out := notes.Clone()
	slices.SortStableFunc(out, func(x, y Note) int {
		r, _ := sys.Compare(x.pitch(), y.pitch()) // the notes are checked: no error

		return sign * r
	})

	return out, nil
}

// SortedByPosition returns a new slice of notes sorted by position from the
// lowest (system.System.Compare); the correction does not take part. Each note
// is in its own settings; all the systems must be equal.
//
// Errors, in this order: ErrUnset for a note; ErrNoOctave; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; ErrDifferentSystems.
func SortedByPosition(notes []Note) ([]Note, error) {
	return Notes(notes).SortedByPosition()
}

// MustSortedByPosition is like SortedByPosition but panics with its error.
func MustSortedByPosition(notes []Note) []Note {
	return Notes(notes).MustSortedByPosition()
}

// SortedByPositionIn is like SortedByPosition in the settings of the explicit
// source src for all notes. A nil src reports ErrUnset.
func SortedByPositionIn(src settings.Source, notes []Note) ([]Note, error) {
	return Notes(notes).SortedByPositionIn(src)
}

// MustSortedByPositionIn is like SortedByPositionIn but panics with its error.
func MustSortedByPositionIn(src settings.Source, notes []Note) []Note {
	return Notes(notes).MustSortedByPositionIn(src)
}

// SortedByPositionInTuning is like SortedByPosition in the system of the
// explicit tuning t. A nil or unset t reports ErrUnset.
func SortedByPositionInTuning(t *tuning.Tuning, notes []Note) ([]Note, error) {
	return Notes(notes).SortedByPositionInTuning(t)
}

// MustSortedByPositionInTuning is like SortedByPositionInTuning but panics
// with its error.
func MustSortedByPositionInTuning(t *tuning.Tuning, notes []Note) []Note {
	return Notes(notes).MustSortedByPositionInTuning(t)
}

// SortedByPositionDescending returns a new slice of notes sorted by position
// from the highest (system.System.Compare); the correction does not take
// part. It has the comparison, validation and source snapshot rules of
// SortedByPosition; notes of equal position keep their input order.
//
// Errors, in this order: ErrUnset for a note; ErrNoOctave; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; ErrDifferentSystems.
func SortedByPositionDescending(notes []Note) ([]Note, error) {
	return Notes(notes).SortedByPositionDescending()
}

// MustSortedByPositionDescending is like SortedByPositionDescending but
// panics with its error.
func MustSortedByPositionDescending(notes []Note) []Note {
	return Notes(notes).MustSortedByPositionDescending()
}

// soundKey is the key of a note in a sort by sound: the exact height or the
// frequency.
type soundKey struct {
	note   Note
	height cents.Cents
	freq   float64
}

func sortedBySound(op string, c opContext, notes Notes, sign int) (Notes, error) {
	if len(notes) == 0 {
		if err := c.check(op); err != nil {
			return nil, err
		}

		return Notes{}, nil
	}

	fs, err := sounding(op, c, nil, notes)
	if err != nil {
		return nil, err
	}

	// Notes with one snapshot share a frame: each distinct tuning is copied
	// and compared with the first one once.
	ts := make([]*tuning.Tuning, len(fs))
	exact := true
	compared := map[*frame]bool{fs[0]: true}

	for i, f := range fs {
		ts[i] = f.tuning()

		if !compared[f] {
			compared[f] = true
			exact = exact && ts[i].Equal(ts[0])
		}
	}

	keys := make([]soundKey, len(notes))

	for i, x := range notes {
		keys[i].note = x

		if exact {
			keys[i].height, err = x.height(op, ts[i])
		} else {
			keys[i].freq, err = x.sound(op, ts[i])
		}

		if err != nil {
			return nil, err
		}
	}

	slices.SortStableFunc(keys, func(a, b soundKey) int {
		if exact {
			return sign * a.height.Compare(b.height)
		}

		return sign * cmp.Compare(a.freq, b.freq)
	})

	out := make(Notes, len(keys))
	for i, k := range keys {
		out[i] = k.note
	}

	return out, nil
}

// SortedBySound returns a new slice of notes sorted by sound from the lowest,
// strictly. Each note is in its own settings. When the tunings of all notes
// are equal, the key is the exact height h + c; otherwise it is the frequency
// of each note.
//
// Errors, in this order: ErrUnset for a note; ErrNoOctave; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; ErrOutOfRange for a key that does not fit
// (also the error of tuning or cents).
func SortedBySound(notes []Note) ([]Note, error) {
	return Notes(notes).SortedBySound()
}

// MustSortedBySound is like SortedBySound but panics with its error.
func MustSortedBySound(notes []Note) []Note {
	return Notes(notes).MustSortedBySound()
}

// SortedBySoundIn is like SortedBySound in the settings of the explicit source
// src for all notes. A nil src reports ErrUnset.
func SortedBySoundIn(src settings.Source, notes []Note) ([]Note, error) {
	return Notes(notes).SortedBySoundIn(src)
}

// MustSortedBySoundIn is like SortedBySoundIn but panics with its error.
func MustSortedBySoundIn(src settings.Source, notes []Note) []Note {
	return Notes(notes).MustSortedBySoundIn(src)
}

// SortedBySoundInTuning is like SortedBySound in the explicit tuning t for all
// notes. A nil or unset t reports ErrUnset.
func SortedBySoundInTuning(t *tuning.Tuning, notes []Note) ([]Note, error) {
	return Notes(notes).SortedBySoundInTuning(t)
}

// MustSortedBySoundInTuning is like SortedBySoundInTuning but panics with its
// error.
func MustSortedBySoundInTuning(t *tuning.Tuning, notes []Note) []Note {
	return Notes(notes).MustSortedBySoundInTuning(t)
}

// SortedBySoundDescending returns a new slice of notes sorted by sound from
// the highest, strictly. It has the comparison, validation and source
// snapshot rules of SortedBySound; notes that sound the same keep their input
// order.
//
// Errors, in this order: ErrUnset for a note; ErrNoOctave; ErrNoSource;
// ErrUnsetSettings; ErrInadmissible; ErrOutOfRange for a key that does not fit
// (also the error of tuning or cents).
func SortedBySoundDescending(notes []Note) ([]Note, error) {
	return Notes(notes).SortedBySoundDescending()
}

// MustSortedBySoundDescending is like SortedBySoundDescending but panics with
// its error.
func MustSortedBySoundDescending(notes []Note) []Note {
	return Notes(notes).MustSortedBySoundDescending()
}
