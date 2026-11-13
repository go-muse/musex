package tuning

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning/temperament"
)

var (
	// ErrUnset reports an unset tuning, reference, root, pitch, system,
	// temperament, spelling rule, or rounding mode.
	ErrUnset = errors.New("tuning: unset value")
	// ErrInvalidArgument reports an invalid frequency, root class, or rounding mode.
	ErrInvalidArgument = errors.New("tuning: invalid argument")
	// ErrClassCount reports different class counts in a system and temperament.
	ErrClassCount = errors.New("tuning: class count mismatch")
	// ErrInadmissible reports a pitch or root name that the system does not admit.
	ErrInadmissible = errors.New("tuning: name not admissible")
	// ErrIncomparable reports references or tunings that cannot be compared.
	ErrIncomparable = errors.New("tuning: incomparable values")
	// ErrOutOfRange reports a result outside the range of cents, frequency, or int.
	ErrOutOfRange = errors.New("tuning: value out of range")
	// ErrTooComplex reports an exact cent value beyond the cents record limit.
	ErrTooComplex = errors.New("tuning: exact value too complex")
	// ErrSpelling reports a spelling rule that cannot produce the requested pitch.
	ErrSpelling = errors.New("tuning: spelling rule failed")
)

func invalidArgument() error {
	return ErrInvalidArgument
}

func classCountMismatch() error {
	return ErrClassCount
}

func inadmissible() error {
	return ErrInadmissible
}

func incomparable() error {
	return ErrIncomparable
}

func centsError(err error) error {
	switch {
	case errors.Is(err, cents.ErrTooComplex), errors.Is(err, temperament.ErrTooComplex):
		return fmt.Errorf("tuning: %w: %w", ErrTooComplex, cents.ErrTooComplex)
	case errors.Is(err, cents.ErrOutOfRange), errors.Is(err, temperament.ErrOutOfRange):
		return fmt.Errorf("tuning: %w: %w", ErrOutOfRange, cents.ErrOutOfRange)
	case errors.Is(err, cents.ErrInvalidArgument), errors.Is(err, temperament.ErrInvalidArgument):
		return ErrInvalidArgument
	default:
		return ErrInvalidArgument
	}
}

func centsOutOfRange() error {
	return fmt.Errorf("tuning: %w: %w", ErrOutOfRange, cents.ErrOutOfRange)
}

func systemDistanceError(err error) error {
	switch {
	case errors.Is(err, system.ErrInvalid):
		return ErrUnset
	case errors.Is(err, system.ErrInadmissible):
		return ErrInadmissible
	case errors.Is(err, system.ErrOutOfRange):
		return fmt.Errorf("tuning: %w: %w", ErrOutOfRange, system.ErrOutOfRange)
	default:
		return ErrInvalidArgument
	}
}

func systemMoveError(err error) error {
	switch {
	case errors.Is(err, system.ErrInvalid):
		return ErrUnset
	case errors.Is(err, system.ErrOutOfRange):
		return fmt.Errorf("tuning: %w: %w", ErrOutOfRange, system.ErrOutOfRange)
	case errors.Is(err, system.ErrNotApplicable):
		return fmt.Errorf("tuning: %w: %w", ErrSpelling, system.ErrNotApplicable)
	case errors.Is(err, system.ErrInadmissible):
		return fmt.Errorf("tuning: %w: %w", ErrSpelling, system.ErrInadmissible)
	case errors.Is(err, system.ErrNoSpelling):
		return fmt.Errorf("tuning: %w: %w", ErrSpelling, system.ErrNoSpelling)
	case errors.Is(err, system.ErrBadSpelling):
		return fmt.Errorf("tuning: %w: %w", ErrSpelling, system.ErrBadSpelling)
	default:
		return ErrSpelling
	}
}
