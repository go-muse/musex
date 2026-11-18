package search

import (
	"fmt"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
)

// GroupSpellingByCollection groups matches with equal sets of degree names.
// Groups and their elements retain first-occurrence and input order.
func GroupSpellingByCollection(matches []SpellingMatch) ([][]SpellingMatch, error) {
	if err := checkSpellingMatches(matches); err != nil {
		return nil, err
	}

	return groupBy(matches, func(m SpellingMatch) string { return m.collection }), nil
}

// MustGroupSpellingByCollection is like GroupSpellingByCollection but panics with its error.
func MustGroupSpellingByCollection(matches []SpellingMatch) [][]SpellingMatch {
	return must(GroupSpellingByCollection(matches))
}

// GroupSpellingByTonic groups matches with equal exact tonic names.
func GroupSpellingByTonic(matches []SpellingMatch) ([][]SpellingMatch, error) {
	if err := checkSpellingMatches(matches); err != nil {
		return nil, err
	}

	return groupBy(matches, func(m SpellingMatch) spelling.Spelling { return m.tonic }), nil
}

// MustGroupSpellingByTonic is like GroupSpellingByTonic but panics with its error.
func MustGroupSpellingByTonic(matches []SpellingMatch) [][]SpellingMatch {
	return must(GroupSpellingByTonic(matches))
}

// GroupSpellingByStructure groups structurally equal modes. Distinct catalog
// entries remain distinct elements of their group.
func GroupSpellingByStructure(matches []SpellingMatch) ([][]SpellingMatch, error) {
	if err := checkSpellingMatches(matches); err != nil {
		return nil, err
	}

	return groupBy(matches, func(m SpellingMatch) string { return m.mode.Template().String() }), nil
}

// MustGroupSpellingByStructure is like GroupSpellingByStructure but panics with its error.
func MustGroupSpellingByStructure(matches []SpellingMatch) [][]SpellingMatch {
	return must(GroupSpellingByStructure(matches))
}

// GroupPositionByCollection groups matches with equal sets of degree classes.
func GroupPositionByCollection(matches []PositionMatch) ([][]PositionMatch, error) {
	if err := checkPositionMatches(matches); err != nil {
		return nil, err
	}

	return groupBy(matches, func(m PositionMatch) string { return m.collection.String() }), nil
}

// MustGroupPositionByCollection is like GroupPositionByCollection but panics with its error.
func MustGroupPositionByCollection(matches []PositionMatch) [][]PositionMatch {
	return must(GroupPositionByCollection(matches))
}

// GroupPositionByTonic groups matches belonging to the same tonic family,
// independently of the family representative chosen for display.
func GroupPositionByTonic(matches []PositionMatch) ([][]PositionMatch, error) {
	if err := checkPositionMatches(matches); err != nil {
		return nil, err
	}

	return groupBy(matches, func(m PositionMatch) familyKey {
		return familyKey{class: m.family.class, letter: m.family.letter, fixed: m.family.fixed}
	}), nil
}

// MustGroupPositionByTonic is like GroupPositionByTonic but panics with its error.
func MustGroupPositionByTonic(matches []PositionMatch) [][]PositionMatch {
	return must(GroupPositionByTonic(matches))
}

// GroupPositionByStructure groups structurally equal modes without merging
// their distinct catalog entries.
func GroupPositionByStructure(matches []PositionMatch) ([][]PositionMatch, error) {
	if err := checkPositionMatches(matches); err != nil {
		return nil, err
	}

	return groupBy(matches, func(m PositionMatch) string { return m.mode.Template().String() }), nil
}

// MustGroupPositionByStructure is like GroupPositionByStructure but panics with its error.
func MustGroupPositionByStructure(matches []PositionMatch) [][]PositionMatch {
	return must(GroupPositionByStructure(matches))
}

type familyKey struct {
	class  pitchclass.Class
	letter letter.Letter
	fixed  bool
}

func checkSpellingMatches(matches []SpellingMatch) error {
	if matches == nil {
		return fmt.Errorf("search: group spelling matches: %w", ErrUnset)
	}

	for i, match := range matches {
		if !match.IsValid() {
			return fmt.Errorf("search: spelling match %d: %w", i+1, ErrUnset)
		}
	}

	return nil
}

func checkPositionMatches(matches []PositionMatch) error {
	if matches == nil {
		return fmt.Errorf("search: group position matches: %w", ErrUnset)
	}

	for i, match := range matches {
		if !match.IsValid() {
			return fmt.Errorf("search: position match %d: %w", i+1, ErrUnset)
		}
	}

	if len(matches) < 2 {
		return nil
	}

	for i, match := range matches[1:] {
		if !matches[0].family.sys.Equal(match.family.sys) {
			return fmt.Errorf("search: position match %d: %w", i+2, ErrDifferentSystems)
		}
	}

	return nil
}

func groupBy[T any, K comparable](values []T, key func(T) K) [][]T {
	groups := make([][]T, 0)
	indices := make(map[K]int, len(values))
	for _, value := range values {
		k := key(value)
		index, exists := indices[k]
		if !exists {
			index = len(groups)
			indices[k] = index
			groups = append(groups, make([]T, 0, 1))
		}

		groups[index] = append(groups[index], value)
	}

	return groups
}
