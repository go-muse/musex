package search

import (
	"slices"
	"strings"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
)

// SpellingMatch is one mode and tonic found by exact spelling. Its zero value
// is unset. Copies share only immutable internal state.
type SpellingMatch struct {
	valid      bool
	mode       mode.Mode
	tonic      spelling.Spelling
	inputs     []spelling.Spelling
	degrees    map[spelling.Spelling][]int
	collection string
}

// IsValid reports whether m is a result produced by a successful search.
func (m SpellingMatch) IsValid() bool {
	return m.valid
}

// Mode returns an independent copy of the matched mode, including its origin.
func (m SpellingMatch) Mode() mode.Mode {
	return m.mode
}

// Tonic returns the exact tonic spelling, without an octave or correction.
func (m SpellingMatch) Tonic() spelling.Spelling {
	return m.tonic
}

// InputNames returns the distinct input names in spelling.Compare order.
func (m SpellingMatch) InputNames() []spelling.Spelling {
	return slices.Clone(m.inputs)
}

// DegreesFor returns every degree whose note has name, in increasing order.
// It returns nil for a name outside the input or for an unset match.
func (m SpellingMatch) DegreesFor(name spelling.Spelling) []int {
	if !m.valid {
		return nil
	}

	return slices.Clone(m.degrees[name])
}

// BySpelling finds all catalog entries and tonics whose actual degree notes
// satisfy condition by exact name. Input order, duplicates, octaves,
// corrections, and sources do not affect the result.
func (s Searcher) BySpelling(notes []note.Note, condition Condition) ([]SpellingMatch, error) {
	if err := s.checkInput(notes, condition); err != nil {
		return nil, err
	}

	inputSet := make(map[spelling.Spelling]struct{}, len(notes))
	for _, n := range notes {
		inputSet[n.Spelling()] = struct{}{}
	}

	inputs := make([]spelling.Spelling, 0, len(inputSet))
	for name := range inputSet {
		inputs = append(inputs, name)
	}
	slices.SortFunc(inputs, spelling.Spelling.Compare)

	result := make([]SpellingMatch, 0)
	for entry := range s.catalog.All() {
		m, err := mode.FromEntry(entry)
		if err != nil {
			continue
		}

		intervals := m.DegreeIntervals()
		candidateSet := make(map[spelling.Spelling]struct{}, len(intervals))
		for _, current := range intervals {
			opposite, err := current.Opposite()
			if err != nil {
				continue
			}

			candidate, err := interval.Transpose(inputs[0], opposite)
			if err == nil {
				candidateSet[candidate] = struct{}{}
			}
		}

		candidates := make([]spelling.Spelling, 0, len(candidateSet))
		for candidate := range candidateSet {
			candidates = append(candidates, candidate)
		}
		slices.SortFunc(candidates, spelling.Spelling.Compare)

		for _, tonic := range candidates {
			match, ok := spellingCandidate(m, tonic, inputs, inputSet, condition)
			if ok {
				result = append(result, match)
			}
		}
	}

	return result, nil
}

// MustBySpelling is like BySpelling but panics with its error.
func (s Searcher) MustBySpelling(notes []note.Note, condition Condition) []SpellingMatch {
	return must(s.BySpelling(notes, condition))
}

func spellingCandidate(m mode.Mode, tonic spelling.Spelling, inputs []spelling.Spelling,
	inputSet map[spelling.Spelling]struct{}, condition Condition,
) (SpellingMatch, bool) {
	first, err := note.FromSpelling(tonic, cents.Zero(), nil)
	if err != nil {
		return SpellingMatch{}, false
	}

	notes, err := m.NotesFrom(first)
	if err != nil {
		return SpellingMatch{}, false
	}

	collection := make(map[spelling.Spelling]struct{}, len(notes))
	degrees := make(map[spelling.Spelling][]int, len(inputs))
	for i, n := range notes {
		name := n.Spelling()
		collection[name] = struct{}{}
		if _, wanted := inputSet[name]; wanted {
			degrees[name] = append(degrees[name], i+1)
		}
	}

	if !matchesCondition(len(collection), len(inputSet), condition, func(name spelling.Spelling) bool {
		_, ok := collection[name]
		return ok
	}, inputs...) {
		return SpellingMatch{}, false
	}

	names := make([]spelling.Spelling, 0, len(collection))
	for name := range collection {
		names = append(names, name)
	}
	slices.SortFunc(names, spelling.Spelling.Compare)

	var key strings.Builder
	for _, name := range names {
		key.WriteString(name.String())
		key.WriteByte('|')
	}

	return SpellingMatch{
		valid:      true,
		mode:       m,
		tonic:      tonic,
		inputs:     slices.Clone(inputs),
		degrees:    degrees,
		collection: key.String(),
	}, true
}

func matchesCondition[T comparable](collectionSize, inputSize int, condition Condition, contains func(T) bool,
	inputs ...T,
) bool {
	if collectionSize < inputSize {
		return false
	}

	for _, input := range inputs {
		if !contains(input) {
			return false
		}
	}

	return condition == Containment || collectionSize == inputSize
}
