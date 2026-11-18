package search

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/mode"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/template"
)

// Errors reported by search operations. Use errors.Is to inspect both the
// search category and a preserved cause from a dependency.
var (
	ErrUnset               = errors.New("search: unset value")
	ErrEmptyInput          = errors.New("search: empty input")
	ErrInvalidArgument     = errors.New("search: invalid argument")
	ErrNoSystem            = errors.New("search: no system")
	ErrUninterpretableName = errors.New("search: name cannot be interpreted")
	ErrDifferentSystems    = errors.New("search: different systems")
	ErrContext             = errors.New("search: context error")
)

// Condition selects how the input set is related to the notes of a mode.
// The zero value is unset.
type Condition uint8

const (
	// ConditionUnset is the invalid zero condition.
	ConditionUnset Condition = iota
	// Containment requires every distinct input element to occur in the mode.
	Containment
	// FullMatch requires the input and the mode to be equal as sets.
	FullMatch
)

// IsValid reports whether c is Containment or FullMatch.
func (c Condition) IsValid() bool {
	return c == Containment || c == FullMatch
}

// String returns the name of c, or a numeric form for an unknown value.
func (c Condition) String() string {
	switch c {
	case ConditionUnset:
		return "unset"
	case Containment:
		return "containment"
	case FullMatch:
		return "full-match"
	default:
		return fmt.Sprintf("Condition(%d)", uint8(c))
	}
}

// Searcher is an immutable snapshot of the catalog used as a search scope.
// Its zero value is unset. A set Searcher is safe for concurrent searches.
type Searcher struct {
	catalog template.Catalog
	valid   bool
}

// BuiltIn returns a Searcher over the built-in template catalog.
func BuiltIn() Searcher {
	return Searcher{catalog: template.BuiltIn(), valid: true}
}

// New returns a Searcher holding a copy of catalog. A pointer to an empty
// catalog is a valid empty scope; a nil pointer is unset.
func New(catalog *template.Catalog) (Searcher, error) {
	if catalog == nil {
		return Searcher{}, fmt.Errorf("search: new scope: %w", ErrUnset)
	}

	return Searcher{catalog: *catalog, valid: true}, nil
}

// MustNew is like New but panics with its error.
func MustNew(catalog *template.Catalog) Searcher {
	return must(New(catalog))
}

// NewFromRepository returns a Searcher over the repository's catalog snapshot.
func NewFromRepository(repository mode.Repository) (Searcher, error) {
	catalog, ok := repository.Catalog()
	if !ok {
		return Searcher{}, fmt.Errorf("search: new scope from repository: %w", errors.Join(ErrUnset, mode.ErrUnset))
	}

	return Searcher{catalog: catalog, valid: true}, nil
}

// MustNewFromRepository is like NewFromRepository but panics with its error.
func MustNewFromRepository(repository mode.Repository) Searcher {
	return must(NewFromRepository(repository))
}

// IsValid reports whether s holds a catalog snapshot, including an empty one.
func (s Searcher) IsValid() bool {
	return s.valid
}

func (s Searcher) checkInput(notes []note.Note, condition Condition) error {
	if !s.IsValid() {
		return fmt.Errorf("search: unset searcher: %w", ErrUnset)
	}

	if len(notes) == 0 {
		return fmt.Errorf("search: no notes: %w", ErrEmptyInput)
	}

	for i, n := range notes {
		if !n.IsValid() {
			return fmt.Errorf("search: input note %d: %w", i+1, errors.Join(ErrUnset, note.ErrUnset))
		}
	}

	if !condition.IsValid() {
		return fmt.Errorf("search: condition %v: %w", condition, ErrInvalidArgument)
	}

	return nil
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}

	return value
}
