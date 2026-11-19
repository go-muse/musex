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
	// ErrUnsupportedStructure indicates an absolute query without a closed P8 cycle.
	ErrUnsupportedStructure = errors.New("modal: absolute analysis requires a closed perfect-octave cycle")
	// ErrUnsupportedQuality indicates a neutral or other fractional interval quality.
	ErrUnsupportedQuality = errors.New("modal: unsupported interval quality")
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
