package search

import (
	"errors"
	"fmt"
	"iter"
	"slices"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// PositionMatch is one tonic family found by pitch class in one effective
// system. Its zero value is unset.
type PositionMatch struct {
	valid      bool
	mode       mode.Mode
	tonic      spelling.Spelling
	family     TonicFamily
	inputs     []pitchclass.Class
	degrees    map[pitchclass.Class][]int
	collection pcset.Set
}

// IsValid reports whether m is a result produced by a successful search.
func (m PositionMatch) IsValid() bool {
	return m.valid
}

// Mode returns an independent copy of the matched mode, including its origin.
func (m PositionMatch) Mode() mode.Mode {
	return m.mode
}

// Tonic returns the deterministic representative of the tonic family.
func (m PositionMatch) Tonic() spelling.Spelling {
	return m.tonic
}

// Family returns the complete tonic-family identity and its effective system.
func (m PositionMatch) Family() TonicFamily {
	return m.family
}

// InputClasses returns the distinct input classes in increasing number order.
func (m PositionMatch) InputClasses() []pitchclass.Class {
	return slices.Clone(m.inputs)
}

// DegreesForClass returns every degree producing class, in increasing order.
// It returns nil for a class outside the input, another modulus, or an unset match.
func (m PositionMatch) DegreesForClass(class pitchclass.Class) []int {
	if !m.valid {
		return nil
	}

	return slices.Clone(m.degrees[class])
}

// TonicFamily identifies all compatible tonic spellings of a position match.
// A regular-by-classes system unites all letters of one class; otherwise the
// family also has a fixed letter. The zero value is unset.
type TonicFamily struct {
	valid  bool
	sys    *system.System
	class  pitchclass.Class
	letter letter.Letter
	fixed  bool
	mode   mode.Mode
}

// IsValid reports whether f belongs to a position match.
func (f TonicFamily) IsValid() bool {
	return f.valid
}

// System returns the immutable effective system captured by the search.
func (f TonicFamily) System() *system.System {
	return f.sys
}

// Class returns the tonic class in the effective system.
func (f TonicFamily) Class() pitchclass.Class {
	return f.class
}

// Letter returns the fixed letter and true for a nonregular family. It returns
// false when a regular-by-classes family unites all letters, or f is unset.
func (f TonicFamily) Letter() (letter.Letter, bool) {
	if !f.valid || !f.fixed {
		return 0, false
	}

	return f.letter, true
}

// Spellings returns all compatible names of f with absolute alteration at most
// limit, without duplicates, in spelling.Compare order. The sequence is lazy,
// restartable, interruptible, and safe for independent concurrent iterations.
func (f TonicFamily) Spellings(limit alteration.Alteration) (iter.Seq[spelling.Spelling], error) {
	if !f.IsValid() {
		return nil, fmt.Errorf("search: spellings of an unset family: %w", ErrUnset)
	}

	if limit.Sign() < 0 {
		return nil, fmt.Errorf("search: negative spelling limit: %w", errors.Join(ErrInvalidArgument, system.ErrLimit))
	}

	solver := f.solver(nil)
	return solver.spellings(ratOfAlteration(limit)), nil
}

// MustSpellings is like Spellings but panics with its error.
func (f TonicFamily) MustSpellings(limit alteration.Alteration) iter.Seq[spelling.Spelling] {
	return must(f.Spellings(limit))
}

func (f TonicFamily) solver(inputNames map[spelling.Spelling]struct{}) familySolver {
	letters := letter.All()
	if f.fixed {
		letters = []letter.Letter{f.letter}
	}

	intervals := f.mode.DegreeIntervals()
	data := make(map[letter.Letter]degreeData, len(letters))
	for _, current := range letters {
		if currentData, ok := makeDegreeData(f.sys, intervals, current); ok {
			data[current] = currentData
		}
	}

	return familySolver{
		sys:        f.sys,
		class:      f.class.Number(),
		letters:    letters,
		data:       data,
		inputNames: inputNames,
	}
}

// ByPosition finds tonic families by class in sys. Every input name is
// interpreted in sys; sources of the input notes are never read.
func (s Searcher) ByPosition(sys *system.System, notes []note.Note, condition Condition) ([]PositionMatch, error) {
	if err := s.checkInput(notes, condition); err != nil {
		return nil, err
	}

	if !sys.IsValid() {
		return nil, fmt.Errorf("search: explicit system: %w", errors.Join(ErrUnset, system.ErrInvalid))
	}

	input, err := readPositionInput(sys, notes)
	if err != nil {
		return nil, err
	}

	return s.byPosition(sys, input, condition)
}

// MustByPosition is like ByPosition but panics with its error.
func (s Searcher) MustByPosition(sys *system.System, notes []note.Note, condition Condition) []PositionMatch {
	return must(s.ByPosition(sys, notes, condition))
}

// ByPositionFromContext obtains one effective system from snapshots of the
// input sources. Each distinct source is read once, and all systems must be equal.
func (s Searcher) ByPositionFromContext(notes []note.Note, condition Condition) ([]PositionMatch, error) {
	if err := s.checkInput(notes, condition); err != nil {
		return nil, err
	}

	var (
		sys     *system.System
		read    note.Notes
		systems []*system.System
	)
	for _, n := range notes {
		current := (*system.System)(nil)
		found := false
		for i, previous := range read {
			if n.SameSource(previous) {
				current, found = systems[i], true
				break
			}
		}

		if !found {
			set, err := n.Settings()
			if errors.Is(err, note.ErrNoSource) {
				return nil, fmt.Errorf("search: context has no system: %w", errors.Join(ErrNoSystem, note.ErrNoSource))
			}
			if err != nil {
				return nil, fmt.Errorf("search: read context: %w", errors.Join(ErrContext, err))
			}

			current = set.System()
			read = append(read, n)
			systems = append(systems, current)
		}

		if sys == nil {
			sys = current
		} else if !sys.Equal(current) {
			return nil, fmt.Errorf("search: context systems differ: %w", errors.Join(ErrDifferentSystems, note.ErrDifferentSystems))
		}
	}

	input, err := readPositionInput(sys, notes)
	if err != nil {
		return nil, err
	}

	return s.byPosition(sys, input, condition)
}

// MustByPositionFromContext is like ByPositionFromContext but panics with its error.
func (s Searcher) MustByPositionFromContext(notes []note.Note, condition Condition) []PositionMatch {
	return must(s.ByPositionFromContext(notes, condition))
}

type positionInput struct {
	classes []pitchclass.Class
	set     pcset.Set
	names   map[spelling.Spelling]struct{}
}

func readPositionInput(sys *system.System, notes note.Notes) (positionInput, error) {
	classes := make(map[pitchclass.Class]struct{}, len(notes))
	names := make(map[spelling.Spelling]struct{}, len(notes))
	for i, n := range notes {
		name := n.Spelling()
		names[name] = struct{}{}
		class, err := sys.Class(name)
		if err != nil {
			return positionInput{}, fmt.Errorf("search: input name %d: %w", i+1, errors.Join(ErrUninterpretableName, err))
		}
		classes[class] = struct{}{}
	}

	ordered := make([]pitchclass.Class, 0, len(classes))
	numbers := make([]int, 0, len(classes))
	for class := range classes {
		ordered = append(ordered, class)
	}
	slices.SortFunc(ordered, pitchclass.Class.Compare)
	for _, class := range ordered {
		numbers = append(numbers, class.Number())
	}

	set, err := pcset.New(sys.Classes(), numbers...)
	if err != nil {
		return positionInput{}, fmt.Errorf("search: input classes: %w", err)
	}

	return positionInput{classes: ordered, set: set, names: names}, nil
}

func (s Searcher) byPosition(sys *system.System, input positionInput, condition Condition) ([]PositionMatch, error) {
	regularity, err := sys.Regularity()
	if err != nil {
		return nil, fmt.Errorf("search: system regularity: %w", err)
	}

	result := make([]PositionMatch, 0)
	for entry := range s.catalog.All() {
		m, err := mode.FromEntry(entry)
		if err != nil {
			continue
		}

		intervals := m.DegreeIntervals()
		data := make(map[letter.Letter]degreeData, letter.Count)
		for _, current := range letter.All() {
			if currentData, ok := makeDegreeData(sys, intervals, current); ok {
				data[current] = currentData
			}
		}
		if len(data) == 0 {
			continue
		}

		entryResult := make([]PositionMatch, 0)
		if regularity.RegularByClasses {
			currentData, ok := data[letter.C]
			if !ok {
				continue
			}
			entryResult = s.positionCandidates(m, sys, input, condition, currentData, data, 0)
		} else {
			for _, current := range letter.All() {
				currentData, ok := data[current]
				if !ok {
					continue
				}
				entryResult = append(entryResult,
					s.positionCandidates(m, sys, input, condition, currentData, data, current)...)
			}
		}

		slices.SortStableFunc(entryResult, func(a, b PositionMatch) int {
			if a.family.class.Number() != b.family.class.Number() {
				return a.family.class.Number() - b.family.class.Number()
			}
			return a.family.letter.Ordinal() - b.family.letter.Ordinal()
		})
		result = append(result, entryResult...)
	}

	return result, nil
}

func (s Searcher) positionCandidates(m mode.Mode, sys *system.System, input positionInput,
	condition Condition, data degreeData, allData map[letter.Letter]degreeData, fixed letter.Letter,
) []PositionMatch {
	collection, err := pcset.New(sys.Classes(), data.offsets...)
	if err != nil {
		return nil
	}
	tonics, err := input.set.CoveringTranspositionsOf(collection)
	if err != nil {
		return nil
	}

	family := TonicFamily{sys: sys, mode: m, fixed: fixed.IsValid(), letter: fixed}
	letters := letter.All()
	if fixed.IsValid() {
		letters = []letter.Letter{fixed}
	}

	result := make([]PositionMatch, 0)
	for _, number := range tonics.Numbers() {
		class, err := pitchclass.New(number, sys.Classes())
		if err != nil {
			continue
		}

		family.valid, family.class = true, class
		solver := familySolver{
			sys:        sys,
			class:      number,
			letters:    letters,
			data:       allData,
			inputNames: input.names,
		}
		tonic, ok := solver.representative()
		if !ok {
			continue
		}

		match, ok := positionCandidate(family, tonic, input, condition)
		if ok {
			result = append(result, match)
		}
	}

	return result
}

func positionCandidate(family TonicFamily, tonic spelling.Spelling, input positionInput,
	condition Condition,
) (PositionMatch, bool) {
	first, err := note.FromSpelling(tonic, cents.Zero(), nil)
	if err != nil {
		return PositionMatch{}, false
	}

	notes, err := family.mode.NotesFromIn(family.sys, first)
	if err != nil {
		return PositionMatch{}, false
	}

	classes := make([]pitchclass.Class, 0, len(notes))
	degrees := make(map[pitchclass.Class][]int, len(input.classes))
	inputClasses := make(map[pitchclass.Class]struct{}, len(input.classes))
	for _, class := range input.classes {
		inputClasses[class] = struct{}{}
	}

	for i, current := range notes {
		class, err := family.sys.Class(current.Spelling())
		if err != nil {
			return PositionMatch{}, false
		}
		classes = append(classes, class)
		if _, wanted := inputClasses[class]; wanted {
			degrees[class] = append(degrees[class], i+1)
		}
	}

	collection, err := pcset.FromClassesWithModulus(family.sys.Classes(), classes...)
	if err != nil {
		return PositionMatch{}, false
	}

	if !matchesCondition(collection.Cardinality(), input.set.Cardinality(), condition,
		func(class pitchclass.Class) bool { return collection.Contains(class) }, input.classes...) {
		return PositionMatch{}, false
	}

	return PositionMatch{
		valid:      true,
		mode:       family.mode,
		tonic:      tonic,
		family:     family,
		inputs:     slices.Clone(input.classes),
		degrees:    degrees,
		collection: collection,
	}, true
}
