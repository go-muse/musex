package temperament

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/cents"
)

// ErrInvalidArgument reports an invalid argument: an uninitialized zero-value
// [Temperament] receiver, a step count outside [1, MaxStepsPerOctave], a
// violated table invariant (T(0) != 0, non-increasing offsets, or an offset of
// 1200 cents or more), a NaN or infinite float offset, an invalid ratio, an
// out-of-bounds step index or parameter, a meantone fifth outside
// (4800/7, 720) cents, an unclosed circle of fifths, or a 12-tone analysis
// called on a temperament with N != 12.
var ErrInvalidArgument = errors.New("temperament: invalid argument")

// ErrPrimeLimit reports a frequency ratio with a prime factor > 31 in its
// reduced form.
var ErrPrimeLimit = errors.New("temperament: prime factor exceeds 31-limit")

// ErrTooComplex reports that an exact cents representation exceeds the
// complexity limit of package cents.
var ErrTooComplex = errors.New("temperament: exact representation too complex")

// ErrOutOfRange reports that a computed cents offset or interval exceeds the
// guaranteed range of package cents.
var ErrOutOfRange = errors.New("temperament: value out of range")

// errorf returns an error that wraps sentinel and appends the formatted
// details: "<sentinel text>: <details>".
func errorf(sentinel error, format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{sentinel}, args...)...)
}

// wrapCentsError translates an error from package cents into the corresponding
// package temperament sentinel error while preserving the underlying cents error
// in the error chain: "<sentinel text>: <op>: <cents error>".
func wrapCentsError(op string, err error) error {
	var sentinel error
	switch {
	case errors.Is(err, cents.ErrPrimeLimit):
		sentinel = ErrPrimeLimit
	case errors.Is(err, cents.ErrTooComplex):
		sentinel = ErrTooComplex
	case errors.Is(err, cents.ErrOutOfRange):
		sentinel = ErrOutOfRange
	default:
		sentinel = ErrInvalidArgument
	}
	return fmt.Errorf("%w: %s: %w", sentinel, op, err)
}

// must returns v if err is nil and panics with err otherwise.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// must2 returns (v1, v2) if err is nil and panics with err otherwise.
func must2[T1, T2 any](v1 T1, v2 T2, err error) (T1, T2) {
	if err != nil {
		panic(err)
	}
	return v1, v2
}
