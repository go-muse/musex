package mode

import (
	"fmt"
	"iter"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/template"
)

// FromTemplate creates a mode without catalog provenance.
func FromTemplate(t template.Template) (Mode, error) {
	return buildMode(t, template.Entry{}, "")
}

// FromEntry creates a mode with e as its origin.
func FromEntry(e template.Entry) (Mode, error) {
	if !e.IsValid() {
		return Mode{}, ErrUnset
	}
	return buildMode(e.Template(), e, "")
}

// FromCode finds an exact code in the built-in catalog.
func FromCode(code string) (Mode, error) {
	catalog := template.BuiltIn()
	return FromCodeIn(&catalog, code)
}

// FromCodeIn finds an exact code in the supplied catalog.
func FromCodeIn(catalog *template.Catalog, code string) (Mode, error) {
	if catalog == nil {
		return Mode{}, ErrUnset
	}
	snapshot := *catalog
	entry, ok := snapshot.ByCode(code)
	if !ok {
		return Mode{}, fmt.Errorf("%w: code %q", ErrNoEntry, code)
	}
	return FromEntry(entry)
}

// FromName finds a normalized name in the built-in catalog.
func FromName(name string) (Mode, error) {
	catalog := template.BuiltIn()
	return FromNameIn(&catalog, name)
}

// FromNameIn finds a normalized name and preserves the chosen catalog spelling.
// Multiple matches are allowed when they all have the same structure.
func FromNameIn(catalog *template.Catalog, name string) (Mode, error) {
	if catalog == nil {
		return Mode{}, ErrUnset
	}
	snapshot := *catalog
	entries := snapshot.ByName(name)
	if len(entries) == 0 {
		return Mode{}, fmt.Errorf("%w: name %q", ErrNoEntry, name)
	}
	entry := entries[0]
	structure := entry.Template()
	for _, other := range entries[1:] {
		if !structure.Equal(other.Template()) {
			return Mode{}, fmt.Errorf("%w: %q", ErrAmbiguousName, name)
		}
	}
	for _, original := range entry.Names() {
		// Use the same normalization as the full catalog for a single alias.
		single, err := template.NewEntry(entry.Code(), []string{original}, structure)
		if err != nil {
			return Mode{}, templateError(err)
		}
		alias, err := template.NewCatalog(single)
		if err != nil {
			return Mode{}, templateError(err)
		}
		if len(alias.ByName(name)) != 0 {
			return buildMode(structure, entry, original)
		}
	}
	return Mode{}, fmt.Errorf("%w: name %q", ErrNoEntry, name)
}

// ClosedFromIntervals creates a cyclic mode from its adjacent intervals.
func ClosedFromIntervals(adjacent []interval.Interval) (Mode, error) {
	return fromConstruction(template.ClosedFromIntervals(adjacent))
}

// OpenFromIntervals creates an open mode from its adjacent intervals.
func OpenFromIntervals(adjacent []interval.Interval) (Mode, error) {
	return fromConstruction(template.OpenFromIntervals(adjacent))
}

// ClosedFromDegreeIntervals creates a cyclic mode from its degree intervals and cycle.
func ClosedFromDegreeIntervals(degrees []interval.Interval, cycle interval.Interval) (Mode, error) {
	return fromConstruction(template.ClosedFromPlaceIntervals(degrees, cycle))
}

// OpenFromDegreeIntervals creates an open mode from its degree intervals.
func OpenFromDegreeIntervals(degrees []interval.Interval) (Mode, error) {
	return fromConstruction(template.OpenFromPlaceIntervals(degrees))
}

func fromConstruction(t template.Template, err error) (Mode, error) {
	if err != nil {
		return Mode{}, templateError(err)
	}
	return FromTemplate(t)
}

// ClosedFromNotes reads only the spellings and uses the supplied cycle.
func ClosedFromNotes(notes []note.Note, cycle interval.Interval) (Mode, error) {
	names, err := noteSpellings(notes)
	if err != nil {
		return Mode{}, err
	}
	return fromConstruction(template.ClosedFromSpellings(names, cycle))
}

// FromNotesInOctave reads the spellings in a perfect-octave cycle.
func FromNotesInOctave(notes []note.Note) (Mode, error) {
	return ClosedFromNotes(notes, interval.PerfectOctave())
}

// OpenFromNotes reads an open sequence of spellings from notes.
func OpenFromNotes(notes []note.Note) (Mode, error) {
	names, err := noteSpellings(notes)
	if err != nil {
		return Mode{}, err
	}
	return fromConstruction(template.OpenFromSpellings(names))
}

func noteSpellings(notes []note.Note) ([]spelling.Spelling, error) {
	for i, n := range notes {
		if !n.IsValid() {
			return nil, fmt.Errorf("%w: note %d", ErrUnset, i+1)
		}
	}
	names := make([]spelling.Spelling, len(notes))
	for i, n := range notes {
		names[i] = n.Spelling()
	}
	return names, nil
}

// FromDegree creates a mode starting at the selected degree of source.
func FromDegree(source Mode, first int) (Mode, error) {
	if err := source.checkDegrees(first); err != nil {
		return Mode{}, err
	}
	t, err := source.structure.Rotate(first)
	if err != nil {
		return Mode{}, templateError(err)
	}
	return buildMode(t, source.origin, source.name)
}

// Rotations validates every rotation before returning a repeatable iterator.
// The iterator reads a snapshot of source and yields rotations in degree order.
func Rotations(source Mode) (iter.Seq[Mode], error) {
	if err := source.checkClosed(); err != nil {
		return nil, err
	}
	for first := 2; first <= source.Len(); first++ {
		t := source.structure.MustRotate(first)
		if _, err := t.PlaceIntervals(); err != nil {
			return nil, templateError(err)
		}
	}
	return func(yield func(Mode) bool) {
		for first := 1; first <= source.Len(); first++ {
			t := source.structure.MustRotate(first)
			// Each immutable rotation was checked above.
			candidate := modeFromPrefixes(t, t.MustPlaceIntervals(), source.origin, source.name)
			if !yield(candidate) {
				return
			}
		}
	}, nil
}

// Parse reads the template interval syntax and creates a mode without origin.
func Parse(text string) (Mode, error) {
	return fromConstruction(template.Parse(text))
}
