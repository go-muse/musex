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

func sortedByPosition(op string, c opContext, notes []Note) ([]Note, error) {
	if len(notes) == 0 {
		if err := c.check(op); err != nil {
			return nil, err
		}

		return []Note{}, nil
	}

	sys, err := position(op, c, notes, false)
	if err != nil {
		return nil, err
	}

	out := slices.Clone(notes)
	slices.SortStableFunc(out, func(x, y Note) int {
		r, _ := sys.Compare(x.pitch(), y.pitch()) // the notes are checked: no error

		return r
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
	return sortedByPosition("SortedByPosition", ownContext(), notes)
}

// MustSortedByPosition is like SortedByPosition but panics with its error.
func MustSortedByPosition(notes []Note) []Note {
	return must(SortedByPosition(notes))
}

// SortedByPositionIn is like SortedByPosition in the settings of the explicit
// source src for all notes. A nil src reports ErrUnset.
func SortedByPositionIn(src settings.Source, notes []Note) ([]Note, error) {
	return sortedByPosition("SortedByPositionIn", sourceContext(src), notes)
}

// MustSortedByPositionIn is like SortedByPositionIn but panics with its error.
func MustSortedByPositionIn(src settings.Source, notes []Note) []Note {
	return must(SortedByPositionIn(src, notes))
}

// SortedByPositionInTuning is like SortedByPosition in the system of the
// explicit tuning t. A nil or unset t reports ErrUnset.
func SortedByPositionInTuning(t *tuning.Tuning, notes []Note) ([]Note, error) {
	return sortedByPosition("SortedByPositionInTuning", tuningContext(t), notes)
}

// MustSortedByPositionInTuning is like SortedByPositionInTuning but panics
// with its error.
func MustSortedByPositionInTuning(t *tuning.Tuning, notes []Note) []Note {
	return must(SortedByPositionInTuning(t, notes))
}

// soundKey is the key of a note in a sort by sound: the exact height or the
// frequency.
type soundKey struct {
	note   Note
	height cents.Cents
	freq   float64
}

func sortedBySound(op string, c opContext, notes []Note) ([]Note, error) {
	if len(notes) == 0 {
		if err := c.check(op); err != nil {
			return nil, err
		}

		return []Note{}, nil
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
			return a.height.Compare(b.height)
		}

		return cmp.Compare(a.freq, b.freq)
	})

	out := make([]Note, len(keys))
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
	return sortedBySound("SortedBySound", ownContext(), notes)
}

// MustSortedBySound is like SortedBySound but panics with its error.
func MustSortedBySound(notes []Note) []Note {
	return must(SortedBySound(notes))
}

// SortedBySoundIn is like SortedBySound in the settings of the explicit source
// src for all notes. A nil src reports ErrUnset.
func SortedBySoundIn(src settings.Source, notes []Note) ([]Note, error) {
	return sortedBySound("SortedBySoundIn", sourceContext(src), notes)
}

// MustSortedBySoundIn is like SortedBySoundIn but panics with its error.
func MustSortedBySoundIn(src settings.Source, notes []Note) []Note {
	return must(SortedBySoundIn(src, notes))
}

// SortedBySoundInTuning is like SortedBySound in the explicit tuning t for all
// notes. A nil or unset t reports ErrUnset.
func SortedBySoundInTuning(t *tuning.Tuning, notes []Note) ([]Note, error) {
	return sortedBySound("SortedBySoundInTuning", tuningContext(t), notes)
}

// MustSortedBySoundInTuning is like SortedBySoundInTuning but panics with its
// error.
func MustSortedBySoundInTuning(t *tuning.Tuning, notes []Note) []Note {
	return must(SortedBySoundInTuning(t, notes))
}
