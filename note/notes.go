package note

import (
	"slices"

	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/tuning"
)

// Notes is an ordered, mutable list of notes. It preserves repetitions and
// permits unset notes; each operation checks the elements it needs. Its zero
// value is a nil, empty list. Literal, make, indexing, range and append work
// as for []Note.
//
// Assignment, slicing and conversion from []Note share the underlying array.
// Use Clone for independently editable elements. Sources remain shared as
// part of each Note; copying a list neither reads nor clones its sources.
// Notes is not comparable and cannot be a map key. Its methods do not change
// the input. Concurrent readers are safe while nobody writes the same list
// or its elements, subject to the concurrency contract of settings.Source.
type Notes []Note

// Clone returns an independent copy of the list, preserving order, repetitions,
// complete note values and nil versus non-nil emptiness. It does not validate
// notes or read their settings sources; those sources remain shared.
func (ns Notes) Clone() Notes {
	return slices.Clone(ns)
}

// SortedByPosition returns a stable, independent list ordered by position.
// It has the errors and source snapshot rules of SortedByPosition. An empty
// input gives a non-nil empty list; an error gives nil.
func (ns Notes) SortedByPosition() (Notes, error) {
	return sortedByPosition("SortedByPosition", ownContext(), ns)
}

// MustSortedByPosition calls SortedByPosition and panics with its error.
func (ns Notes) MustSortedByPosition() Notes {
	return must(ns.SortedByPosition())
}

// SortedByPositionIn is SortedByPosition in the explicit source src for every
// note. It follows the contract of SortedByPositionIn, including its errors.
// It never reads the notes' own sources or changes their bindings.
func (ns Notes) SortedByPositionIn(src settings.Source) (Notes, error) {
	return sortedByPosition("SortedByPositionIn", sourceContext(src), ns)
}

// MustSortedByPositionIn calls SortedByPositionIn and panics with its error.
func (ns Notes) MustSortedByPositionIn(src settings.Source) Notes {
	return must(ns.SortedByPositionIn(src))
}

// SortedByPositionInTuning is SortedByPosition in the system of t for every
// note, with the errors of SortedByPositionInTuning. It reads no source.
func (ns Notes) SortedByPositionInTuning(t *tuning.Tuning) (Notes, error) {
	return sortedByPosition("SortedByPositionInTuning", tuningContext(t), ns)
}

// MustSortedByPositionInTuning calls SortedByPositionInTuning and panics with
// its error.
func (ns Notes) MustSortedByPositionInTuning(t *tuning.Tuning) Notes {
	return must(ns.SortedByPositionInTuning(t))
}

// SortedBySound returns a stable, independent list ordered strictly by sound.
// It has the errors and source snapshot rules of SortedBySound. An empty
// input gives a non-nil empty list; an error gives nil.
func (ns Notes) SortedBySound() (Notes, error) {
	return sortedBySound("SortedBySound", ownContext(), ns)
}

// MustSortedBySound calls SortedBySound and panics with its error.
func (ns Notes) MustSortedBySound() Notes {
	return must(ns.SortedBySound())
}

// SortedBySoundIn is SortedBySound in the explicit source src for every note.
// It follows the contract of SortedBySoundIn, including its errors. It never
// reads the notes' own sources or changes their bindings.
func (ns Notes) SortedBySoundIn(src settings.Source) (Notes, error) {
	return sortedBySound("SortedBySoundIn", sourceContext(src), ns)
}

// MustSortedBySoundIn calls SortedBySoundIn and panics with its error.
func (ns Notes) MustSortedBySoundIn(src settings.Source) Notes {
	return must(ns.SortedBySoundIn(src))
}

// SortedBySoundInTuning is SortedBySound in t for every note, with the errors
// of SortedBySoundInTuning. It reads no source.
func (ns Notes) SortedBySoundInTuning(t *tuning.Tuning) (Notes, error) {
	return sortedBySound("SortedBySoundInTuning", tuningContext(t), ns)
}

// MustSortedBySoundInTuning calls SortedBySoundInTuning and panics with its error.
func (ns Notes) MustSortedBySoundInTuning(t *tuning.Tuning) Notes {
	return must(ns.SortedBySoundInTuning(t))
}
