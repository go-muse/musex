package tempomap

import (
	"errors"
	"fmt"
	"math/big"
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

type textParseError struct {
	message string
	causes  []error
}

func (e *textParseError) Error() string { return e.message }

func (e *textParseError) Unwrap() []error { return e.causes }

func quotedTextError(text string, causes ...error) error {
	return &textParseError{
		message: causes[0].Error() + ": " + quoteExcerpt(text),
		causes:  causes,
	}
}

func rangeTextError(text string, cause error) error {
	causes := []error{ErrOutOfRange, fraction.ErrOutOfRange}
	for _, dependency := range []error{duration.ErrOutOfRange, measure.ErrOutOfRange, tempo.ErrOutOfRange} {
		if errors.Is(cause, dependency) {
			causes = append(causes, dependency)
		}
	}
	return quotedTextError(text, causes...)
}

func normalizeTextError(text string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrOutOfRange) ||
		errors.Is(err, fraction.ErrOutOfRange) ||
		errors.Is(err, duration.ErrOutOfRange) ||
		errors.Is(err, measure.ErrOutOfRange) ||
		errors.Is(err, tempo.ErrOutOfRange) {
		return rangeTextError(text, err)
	}
	var sentinel error
	switch {
	case errors.Is(err, ErrInvalidArgument):
		sentinel = ErrInvalidArgument
	case errors.Is(err, ErrNoInitialTempo):
		sentinel = ErrNoInitialTempo
	case errors.Is(err, ErrDuplicateAddress):
		sentinel = ErrDuplicateAddress
	case errors.Is(err, ErrInvalidAddress):
		sentinel = ErrInvalidAddress
	case errors.Is(err, ErrNoDeviation):
		sentinel = ErrNoDeviation
	case errors.Is(err, ErrInvalidSequence):
		sentinel = ErrInvalidSequence
	case errors.Is(err, ErrNegativeTime):
		sentinel = ErrNegativeTime
	default:
		return invalidTextError(text)
	}
	causes := []error{sentinel}
	for _, dependency := range []error{
		duration.ErrUnset,
		measure.ErrUnset,
		measure.ErrOutsideGrid,
		measure.ErrOutsideBar,
		tempo.ErrUnset,
		tempo.ErrNegativeTime,
	} {
		if errors.Is(err, dependency) {
			causes = append(causes, dependency)
		}
	}
	return quotedTextError(text, causes...)
}

func numericTextOutOfRange(text string) bool {
	if text == "" {
		return false
	}
	digits := text
	if digits[0] == '-' {
		digits = digits[1:]
		if digits == "" {
			return false
		}
	}
	for i := range len(digits) {
		if digits[i] < '0' || digits[i] > '9' {
			return false
		}
	}
	component, ok := new(big.Int).SetString(digits, 10)
	return ok && component.Cmp(fraction.MaxComponentBig()) > 0
}

func nestedTextError(text, value string, err error) error {
	if errors.Is(err, fraction.ErrOutOfRange) ||
		errors.Is(err, duration.ErrOutOfRange) ||
		errors.Is(err, measure.ErrOutOfRange) ||
		errors.Is(err, tempo.ErrOutOfRange) ||
		numericTextOutOfRange(value) {
		return rangeTextError(text, err)
	}
	return invalidTextError(text)
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
