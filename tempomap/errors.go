package tempomap

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempo"
)

// Sentinel errors returned by package operations. Each error wraps exactly one
// tempomap sentinel and preserves the relevant causes from dependencies.
var (
	ErrUnset            = errors.New("tempomap: unset value")
	ErrInvalidArgument  = errors.New("tempomap: invalid argument")
	ErrNoInitialTempo   = errors.New("tempomap: no initial main tempo")
	ErrDuplicateAddress = errors.New("tempomap: duplicate address")
	ErrInvalidAddress   = errors.New("tempomap: invalid address")
	ErrNoDeviation      = errors.New("tempomap: no current deviation")
	ErrInvalidSequence  = errors.New("tempomap: invalid change sequence")
	ErrUnfinished       = errors.New("tempomap: unfinished gradual change")
	ErrNegativeTime     = errors.New("tempomap: negative time")
	ErrOutOfRange       = errors.New("tempomap: value out of range")
	ErrInvalidText      = errors.New("tempomap: invalid text")
)

func simpleError(sentinel error, op, detail string) error {
	return fmt.Errorf("%w: %s: %s", sentinel, op, detail)
}

func causedError(sentinel error, op, detail string, cause error) error {
	return fmt.Errorf("%w: %s: %s: %w", sentinel, op, detail, cause)
}

func rangeError(op string) error {
	return fmt.Errorf("%w: %s: %w", ErrOutOfRange, op, fraction.ErrOutOfRange)
}

func modeError(op string, mode rounding.Mode) error {
	cause := rounding.ErrUnknownMode
	if mode == 0 {
		cause = rounding.ErrUnset
	}
	return causedError(ErrInvalidArgument, op, "rounding mode", cause)
}

func checkPrecision(op string, places int, mode rounding.Mode) error {
	if places < 0 || places > MaxPlaces {
		return simpleError(ErrInvalidArgument, op, "decimal places outside 0..1000")
	}
	if !mode.IsValid() {
		return modeError(op, mode)
	}
	return nil
}

func dependencyArgumentError(op string, err error) error {
	return causedError(ErrInvalidArgument, op, "dependency argument", err)
}

func measureOperationError(op string, err error) error {
	switch {
	case errors.Is(err, measure.ErrOutsideGrid), errors.Is(err, measure.ErrOutsideBar):
		return causedError(ErrInvalidAddress, op, "address outside grid", err)
	case errors.Is(err, measure.ErrOutOfRange):
		return fmt.Errorf("%w: %s: %w: %w", ErrOutOfRange, op, err, fraction.ErrOutOfRange)
	case errors.Is(err, measure.ErrUnset), errors.Is(err, duration.ErrUnset):
		return dependencyArgumentError(op, err)
	default:
		return causedError(ErrInvalidArgument, op, "measure operation", err)
	}
}

func tempoRangeError(op string, err error) error {
	return fmt.Errorf("%w: %s: %w: %w", ErrOutOfRange, op, err, fraction.ErrOutOfRange)
}

const maxErrorQuoteBytes = 64

func quoteExcerpt(text string) string {
	if len(text) <= maxErrorQuoteBytes {
		return strconv.Quote(text)
	}
	k := 0
	for k < maxErrorQuoteBytes {
		_, width := utf8.DecodeRuneInString(text[k:])
		if k+width > maxErrorQuoteBytes {
			break
		}
		k += width
	}
	return strconv.Quote(text[:k]) + "..."
}

func invalidTextError(text string) error {
	return fmt.Errorf("%w: %s", ErrInvalidText, quoteExcerpt(text))
}

func rangeTextError(text string, cause error) error {
	return fmt.Errorf("%w: %w: %s", ErrOutOfRange, cause, quoteExcerpt(text))
}

func nestedTextError(text string, err error) error {
	if !errors.Is(err, fraction.ErrOutOfRange) &&
		!errors.Is(err, duration.ErrOutOfRange) &&
		!errors.Is(err, measure.ErrOutOfRange) &&
		!errors.Is(err, tempo.ErrOutOfRange) {
		return invalidTextError(text)
	}
	return rangeTextError(text, err)
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func must2[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}
	return a, b
}

func must3[A, B, C any](a A, b B, c C, err error) (A, B, C) {
	if err != nil {
		panic(err)
	}
	return a, b, c
}
