package note

import (
	"slices"

	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
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

// UniqueBySpelling returns an independent list keeping the first note of
// each distinct name (letter and alteration), in input order. The octave,
// the correction and the source of that first note are preserved and do not
// take part in the key: C5 and C4 are one spelling, C♯4 and D♭4 are two.
// No source is read. An empty or nil input gives a non-nil empty list.
//
// Errors: ErrUnset for the first unset note.
func (ns Notes) UniqueBySpelling() (Notes, error) {
	const op = "UniqueBySpelling"

	if err := checkNotes(op, ns); err != nil {
		return nil, err
	}

	out := make(Notes, 0, len(ns))
	seen := make(map[spelling.Spelling]bool, len(ns))

	for _, x := range ns {
		if seen[x.name] {
			continue
		}

		seen[x.name] = true
		out = append(out, x)
	}

	return out, nil
}

// MustUniqueBySpelling calls UniqueBySpelling and panics with its error.
func (ns Notes) MustUniqueBySpelling() Notes {
	return must(ns.UniqueBySpelling())
}

// Spellings returns the names of the notes in order, including repetitions,
// as a new slice. The name of a note with an octave is the name of its
// written pitch. Reading names needs no tuning and reads no source. An empty
// or nil input gives a non-nil empty slice.
//
// Errors: ErrUnset for the first unset note.
func (ns Notes) Spellings() ([]spelling.Spelling, error) {
	const op = "Spellings"

	if err := checkNotes(op, ns); err != nil {
		return nil, err
	}

	out := make([]spelling.Spelling, len(ns))
	for i, x := range ns {
		out[i] = x.name
	}

	return out, nil
}

// MustSpellings calls Spellings and panics with its error.
func (ns Notes) MustSpellings() []spelling.Spelling {
	return must(ns.Spellings())
}

// SortedByPosition returns a stable, independent list ordered by position.
// It has the errors and source snapshot rules of SortedByPosition. An empty
// input gives a non-nil empty list; an error gives nil.
func (ns Notes) SortedByPosition() (Notes, error) {
	return sortedByPosition("SortedByPosition", ownContext(), ns, 1)
}

// MustSortedByPosition calls SortedByPosition and panics with its error.
func (ns Notes) MustSortedByPosition() Notes {
	return must(ns.SortedByPosition())
}

// SortedByPositionIn is SortedByPosition in the explicit source src for every
// note. It follows the contract of SortedByPositionIn, including its errors.
// It never reads the notes' own sources or changes their bindings.
func (ns Notes) SortedByPositionIn(src settings.Source) (Notes, error) {
	return sortedByPosition("SortedByPositionIn", sourceContext(src), ns, 1)
}

// MustSortedByPositionIn calls SortedByPositionIn and panics with its error.
func (ns Notes) MustSortedByPositionIn(src settings.Source) Notes {
	return must(ns.SortedByPositionIn(src))
}

// SortedByPositionInTuning is SortedByPosition in the system of t for every
// note, with the errors of SortedByPositionInTuning. It reads no source.
func (ns Notes) SortedByPositionInTuning(t *tuning.Tuning) (Notes, error) {
	return sortedByPosition("SortedByPositionInTuning", tuningContext(t), ns, 1)
}

// MustSortedByPositionInTuning calls SortedByPositionInTuning and panics with
// its error.
func (ns Notes) MustSortedByPositionInTuning(t *tuning.Tuning) Notes {
	return must(ns.SortedByPositionInTuning(t))
}

// SortedByPositionDescending returns a stable, independent list ordered by
// position from the highest. It has the comparison, validation and source
// snapshot rules of SortedByPosition; notes of equal position keep their
// input order. An empty input gives a non-nil empty list; an error gives nil.
func (ns Notes) SortedByPositionDescending() (Notes, error) {
	return sortedByPosition("SortedByPositionDescending", ownContext(), ns, -1)
}

// MustSortedByPositionDescending calls SortedByPositionDescending and panics
// with its error.
func (ns Notes) MustSortedByPositionDescending() Notes {
	return must(ns.SortedByPositionDescending())
}

// SortedBySound returns a stable, independent list ordered strictly by sound.
// It has the errors and source snapshot rules of SortedBySound. An empty
// input gives a non-nil empty list; an error gives nil.
func (ns Notes) SortedBySound() (Notes, error) {
	return sortedBySound("SortedBySound", ownContext(), ns, 1)
}

// MustSortedBySound calls SortedBySound and panics with its error.
func (ns Notes) MustSortedBySound() Notes {
	return must(ns.SortedBySound())
}

// SortedBySoundIn is SortedBySound in the explicit source src for every note.
// It follows the contract of SortedBySoundIn, including its errors. It never
// reads the notes' own sources or changes their bindings.
func (ns Notes) SortedBySoundIn(src settings.Source) (Notes, error) {
	return sortedBySound("SortedBySoundIn", sourceContext(src), ns, 1)
}

// MustSortedBySoundIn calls SortedBySoundIn and panics with its error.
func (ns Notes) MustSortedBySoundIn(src settings.Source) Notes {
	return must(ns.SortedBySoundIn(src))
}

// SortedBySoundInTuning is SortedBySound in t for every note, with the errors
// of SortedBySoundInTuning. It reads no source.
func (ns Notes) SortedBySoundInTuning(t *tuning.Tuning) (Notes, error) {
	return sortedBySound("SortedBySoundInTuning", tuningContext(t), ns, 1)
}

// MustSortedBySoundInTuning calls SortedBySoundInTuning and panics with its error.
func (ns Notes) MustSortedBySoundInTuning(t *tuning.Tuning) Notes {
	return must(ns.SortedBySoundInTuning(t))
}

// SortedBySoundDescending returns a stable, independent list ordered strictly
// by sound from the highest. It has the comparison, validation and source
// snapshot rules of SortedBySound; notes that sound the same keep their input
// order. An empty input gives a non-nil empty list; an error gives nil.
func (ns Notes) SortedBySoundDescending() (Notes, error) {
	return sortedBySound("SortedBySoundDescending", ownContext(), ns, -1)
}

// MustSortedBySoundDescending calls SortedBySoundDescending and panics with
// its error.
func (ns Notes) MustSortedBySoundDescending() Notes {
	return must(ns.SortedBySoundDescending())
}

// FormatIn returns the text of each note in the notation of the explicit
// source src, in order, preserving repetitions: the written pitch for a note
// with an octave (settings.Settings.FormatPitch), the name for a note without
// one (settings.Settings.Format). The correction is not written. The source
// is read once; the own sources of the notes are not read and their bindings
// do not change. An empty or nil input gives a non-nil empty slice without
// reading the source.
//
// Errors, in this order: ErrUnset for the first unset note or a nil src;
// ErrUnsetSettings; then the errors of FormatIn for the first note that
// cannot be formatted.
func (ns Notes) FormatIn(src settings.Source) ([]string, error) {
	return ns.formatIn("FormatIn", src, false)
}

// MustFormatIn calls FormatIn and panics with its error.
func (ns Notes) MustFormatIn(src settings.Source) []string {
	return must(ns.FormatIn(src))
}

// FormatNamesIn returns the text of the name of each note without an octave
// in the notation of the explicit source src, in order, preserving
// repetitions, for notes with or without an octave
// (settings.Settings.Format). The errors are those of FormatIn.
func (ns Notes) FormatNamesIn(src settings.Source) ([]string, error) {
	return ns.formatIn("FormatNamesIn", src, true)
}

// MustFormatNamesIn calls FormatNamesIn and panics with its error.
func (ns Notes) MustFormatNamesIn(src settings.Source) []string {
	return must(ns.FormatNamesIn(src))
}

// formatIn formats every note in the settings of src, read once; nameOnly
// selects the name of the note instead of its written pitch.
func (ns Notes) formatIn(op string, src settings.Source, nameOnly bool) ([]string, error) {
	if err := checkNotes(op, ns); err != nil {
		return nil, err
	}

	if err := checkSource(op, src); err != nil {
		return nil, err
	}

	out := make([]string, 0, len(ns))

	if len(ns) == 0 {
		return out, nil
	}

	set, err := newReader(op).read(src)
	if err != nil {
		return nil, err
	}

	for _, x := range ns {
		var text string

		if x.oct.IsValid() && !nameOnly {
			text, err = set.FormatPitch(x.pitch())
		} else {
			text, err = set.Format(x.name)
		}

		if err != nil {
			return nil, formatError(op, err)
		}

		out = append(out, text)
	}

	return out, nil
}
