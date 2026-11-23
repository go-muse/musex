package modal

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/mode"
)

var (
	// ErrUnset indicates an unset mode; the error also wraps mode.ErrUnset.
	ErrUnset = errors.New("modal: unset mode")
	// ErrInvalidDegree indicates a number outside the actual degrees 1..Len.
	ErrInvalidDegree = errors.New("modal: invalid degree")
	// ErrUnsupportedStructure indicates a mode structure outside the domain of
	// a query. It has three sources: an absolute query on a mode without a
	// closed perfect-octave cycle, a full comparison (ComparePositions) with
	// such a mode, and a relation (RelationTo) of an open mode to a target
	// below its source degree, since an open mode has no next cycle. Each
	// source adds its own context and wraps this error once.
	ErrUnsupportedStructure = errors.New("modal: unsupported mode structure")
	// ErrUnsupportedQuality indicates a neutral or other fractional interval quality.
	ErrUnsupportedQuality = errors.New("modal: unsupported interval quality")
	// ErrIncomparable indicates compared modes with different numbers of
	// degrees. It is distinct from mode.ErrIncomparable.
	ErrIncomparable = errors.New("modal: incomparable modes")
)

func checkDegrees(m mode.Mode, numbers ...int) error {
	if !m.IsValid() {
		return fmt.Errorf("%w: %w", ErrUnset, mode.ErrUnset)
	}
	for _, number := range numbers {
		if number < 1 || number > m.Len() {
			return fmt.Errorf("%w: %d outside 1..%d", ErrInvalidDegree, number, m.Len())
		}
	}
	return nil
}

func classify(q interval.Quality) (Category, error) {
	switch q {
	case interval.Major, interval.Augmented:
		return High, nil
	case interval.Perfect:
		return Neutral, nil
	case interval.Minor, interval.Diminished:
		return Low, nil
	default:
		return CategoryUnset, fmt.Errorf("%w: %s", ErrUnsupportedQuality, q)
	}
}

// checkOctaveCycle rejects a mode outside the domain of absolute analysis.
func checkOctaveCycle(m mode.Mode, context string) error {
	if !m.IsClosed() || !m.IsOctave() {
		return fmt.Errorf("%s requires a closed perfect-octave cycle: %w", context, ErrUnsupportedStructure)
	}
	return nil
}
