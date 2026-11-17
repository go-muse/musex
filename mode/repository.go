package mode

import (
	"math/big"
	"slices"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"
)

// Repository is an immutable catalog snapshot. The zero value is unset.
type Repository struct {
	catalog template.Catalog
	valid   bool
}

// Rotation identifies the smallest starting degree of a catalog entry.
type Rotation struct {
	Entry  template.Entry
	Degree int
}

// Relation identifies a catalog entry, its rotation and its differing degrees.
type Relation struct {
	Entry   template.Entry
	Degree  int
	Degrees []int
}

// Count returns the number of differing degrees.
func (r Relation) Count() int { return len(r.Degrees) }

// NewRepository snapshots catalog. A pointer to an empty catalog is valid.
func NewRepository(catalog *template.Catalog) (Repository, error) {
	if catalog == nil {
		return Repository{}, ErrUnset
	}
	return Repository{catalog: *catalog, valid: true}, nil
}

// BuiltInRepository snapshots the built-in catalog.
func BuiltInRepository() Repository {
	return Repository{catalog: template.BuiltIn(), valid: true}
}

// IsValid reports whether r holds a catalog.
func (r Repository) IsValid() bool { return r.valid }

// Catalog returns the immutable catalog snapshot, if set.
func (r Repository) Catalog() (template.Catalog, bool) { return r.catalog, r.valid }

// Identify returns all catalog entries that exactly match m's structure.
func (r Repository) Identify(m Mode) ([]template.Entry, error) {
	if !r.valid {
		return nil, ErrUnset
	}
	entries, err := r.catalog.Identify(m.structure)
	return entries, templateError(err)
}

// IdentifyRotations returns entries that give m after a rotation.
func (r Repository) IdentifyRotations(m Mode) ([]Rotation, error) {
	if !r.valid {
		return nil, ErrUnset
	}
	matches, err := r.catalog.IdentifyRotations(m.structure)
	if err != nil {
		return nil, templateError(err)
	}
	result := make([]Rotation, len(matches))
	for i, match := range matches {
		result[i] = Rotation{Entry: match.Entry, Degree: match.Place}
	}
	return result, nil
}

// Related finds entries within maxDifferences, ordered by count and catalog
// order. With rotations, each entry contributes its best rotation.
func (r Repository) Related(m Mode, maxDifferences int, measure Measure, rotations bool) ([]Relation, error) {
	if !r.valid {
		return nil, ErrUnset
	}
	matches, err := r.catalog.Related(m.structure, maxDifferences, measure, rotations)
	if err != nil {
		return nil, templateError(err)
	}
	result := make([]Relation, len(matches))
	for i, match := range matches {
		result[i] = Relation{Entry: match.Entry, Degree: match.Place, Degrees: match.Places}
	}
	return result, nil
}

// IdentifyInSystem finds entries with m's kind and adjacent sizes.
func (r Repository) IdentifyInSystem(m Mode, sys *system.System) ([]template.Entry, error) {
	return entriesOf(r.identifyInSystem(m, sys, spelling.Spelling{}, false, false))
}

// IdentifyInSystemFrom looks for numerical matches from the given spelling.
func (r Repository) IdentifyInSystemFrom(m Mode, sys *system.System, start spelling.Spelling) ([]template.Entry, error) {
	return entriesOf(r.identifyInSystem(m, sys, start, true, false))
}

// IdentifyRotationsInSystem matches rotations of the entries' numerical sizes.
func (r Repository) IdentifyRotationsInSystem(m Mode, sys *system.System) ([]Rotation, error) {
	return r.identifyInSystem(m, sys, spelling.Spelling{}, false, true)
}

// IdentifyRotationsInSystemFrom computes each entry's numbers from start
// before looking for a rotation of those numbers.
func (r Repository) IdentifyRotationsInSystemFrom(m Mode, sys *system.System, start spelling.Spelling) ([]Rotation, error) {
	return r.identifyInSystem(m, sys, start, true, true)
}

func entriesOf(matches []Rotation, err error) ([]template.Entry, error) {
	if err != nil {
		return nil, err
	}
	result := make([]template.Entry, len(matches))
	for i, match := range matches {
		result[i] = match.Entry
	}
	return result, nil
}

func (r Repository) identifyInSystem(m Mode, sys *system.System, start spelling.Spelling, requireStart, rotations bool) ([]Rotation, error) {
	if !r.valid {
		return nil, ErrUnset
	}
	query, err := m.numbers(sys, start, requireStart)
	if err != nil {
		return nil, err
	}
	result := make([]Rotation, 0)
	for entry := range r.catalog.All() {
		t := entry.Template()
		if t.Kind() != m.Kind() {
			continue
		}
		candidate, err := numbersForTemplate(sys, start, t)
		if err != nil {
			continue
		}
		if rotations && m.IsClosed() {
			if first, ok := rotationOfNumbers(query.Adjacent, candidate.Adjacent); ok {
				result = append(result, Rotation{Entry: entry, Degree: first})
			}
		} else if slices.Equal(query.Adjacent, candidate.Adjacent) {
			result = append(result, Rotation{Entry: entry, Degree: 1})
		}
	}
	return result, nil
}

// rotationOfNumbers returns the first rotation of source that gives target.
func rotationOfNumbers(target, source []system.Step) (int, bool) {
	if len(target) != len(source) {
		return 0, false
	}
	for i := range source {
		rest := len(source) - i
		if slices.Equal(source[i:], target[:rest]) && slices.Equal(source[:i], target[rest:]) {
			return i + 1, true
		}
	}
	return 0, false
}

func numbersForTemplate(sys *system.System, start spelling.Spelling, t template.Template) (Positions, error) {
	prefixes, err := t.PlaceIntervals()
	if err == nil {
		cycle, _ := t.Cycle()
		return numbersFromIntervals(sys, start, prefixes, cycle)
	}

	// A catalog entry can contain a rotated template whose prefix is too
	// large for Interval. Its exact positions may still fit in system.Step.
	adjacent := t.AdjacentIntervals()
	position := coordinatesOf(interval.PerfectUnison())
	return collectPositions(t.Len(), t.IsClosed(), func(i int) (system.Step, error) {
		if i != 0 {
			position = position.add(coordinatesOf(adjacent[i-1]))
		}
		return exactSteps(sys, start, position)
	})
}

// exactSteps computes an ascending position after the query has already
// validated the system, regularity (when needed), and initial spelling.
func exactSteps(sys *system.System, start spelling.Spelling, position coordinates) (system.Step, error) {
	octaves, rem := splitLetters(position.letters)
	initial := start
	deviation := position.deviation()
	if start.IsValid() {
		var err error
		target, err := position.transposeName(start)
		if err != nil {
			return 0, err
		}
		if _, err := sys.Carry(target); err != nil {
			return 0, err
		}
		deviation = new(big.Rat).Sub(rational(target.Alteration()), rational(start.Alteration()))
	} else {
		// This natural letter only indexes the regular system's step
		// pattern. No spelling is tested in the unanchored calculation.
		initial = spelling.MustParse("C")
	}
	deviation.Mul(deviation, new(big.Rat).SetInt64(int64(sys.Chromatic())))
	if !deviation.IsInt() {
		return 0, interval.ErrFractionalSteps
	}
	total := new(big.Int).Mul(octaves, big.NewInt(int64(sys.Classes())))
	for j := range rem {
		step, err := sys.StepAfter(initial.Letter().Shift(j))
		if err != nil {
			return 0, err
		}
		total.Add(total, big.NewInt(int64(step)))
	}
	total.Add(total, deviation.Num())
	if !total.IsInt64() {
		return 0, interval.ErrOutOfRange
	}
	return system.Step(total.Int64()), nil
}
