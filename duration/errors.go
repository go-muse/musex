package duration

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/tuplet"
)

// Sentinel errors returned by package operations. Every error returned by the
// package wraps exactly one of these sentinels, so they are pairwise
// distinguishable with [errors.Is]; no sentinel wraps another. When the cause
// of a failure is an error of package common/fraction, tuplet, or
// common/rounding, the returned error additionally wraps the sentinel of that
// package, so both checks with [errors.Is] hold. The message text is not part
// of the API, except for the limits on parse error messages described in the
// package documentation.
var (
	// ErrUnset indicates that a receiver, an argument, or a list element is
	// an unset value of the package: Duration{}, Written{}, Tied{}, or
	// Position{}.
	ErrUnset = errors.New("duration: unset value")

	// ErrInvalidArgument indicates an argument outside the domain of an
	// operation: an unset fraction, a zero denominator, a negative number of
	// dots, a base outside the series, a factor, divisor, or grid step that
	// is unset or not positive, an unset or unknown rounding mode, a negative
	// maximum number of dots, an empty list, a tied duration of fewer than two
	// parts, removing a dot or a tuplet that is not there, an unset tuplet, or
	// a nil receiver of UnmarshalText. For a zero denominator it also wraps
	// fraction.ErrZeroDenominator, for an unset tuplet tuplet.ErrUnset, and
	// for a mode rounding.ErrUnset or rounding.ErrUnknownMode.
	ErrInvalidArgument = errors.New("duration: invalid argument")

	// ErrNotPositive indicates a duration that is zero or negative, at
	// creation or as the result of a subtraction or of DurationTo.
	ErrNotPositive = errors.New("duration: duration is not positive")

	// ErrNegativePosition indicates a position less than zero, at creation or
	// as the result of Position.Sub or OffsetTo.
	ErrNegativePosition = errors.New("duration: position is negative")

	// ErrNotNotatable indicates that the notation rule of ToWritten, Notate,
	// or NotateWith gives no signs for the duration. It does not mean that no
	// notation of the duration exists at all.
	ErrNotNotatable = errors.New("duration: notation rule gives no signs")

	// ErrOutsideSeries indicates a result that requires a base beyond either
	// end of the series from maxima to 1024th.
	ErrOutsideSeries = errors.New("duration: base outside the series")

	// ErrNoSubdivision indicates a normal subdivision of a written duration
	// with two or more dots, for which no rule exists.
	ErrNoSubdivision = errors.New("duration: no normal subdivision")

	// ErrQuantizedToZero indicates that a quantized duration or a quantized
	// span is zero.
	ErrQuantizedToZero = errors.New("duration: quantized to zero")

	// ErrTooManyParts indicates that notation with ties would need more than
	// MaxParts parts.
	ErrTooManyParts = errors.New("duration: too many tied parts")

	// ErrInvalidText indicates text that does not match the grammar of the
	// operation, the string of an unset value, or an unknown base name.
	ErrInvalidText = errors.New("duration: invalid text")

	// ErrOutOfRange indicates a result or an intermediate result outside the
	// range of package common/fraction, a tuplet chain coefficient outside
	// that range, a quotient of QuoRem greater than math.MaxInt, or a number
	// in text outside its range. It also wraps fraction.ErrOutOfRange for a
	// fraction outside the range, tuplet.ErrOutOfRange together with
	// fraction.ErrOutOfRange for a chain coefficient,
	// fraction.ErrNotRepresentable for a quotient of QuoRem, and
	// tuplet.ErrOutOfRange for a tuplet count in text greater than
	// math.MaxInt.
	ErrOutOfRange = errors.New("duration: value out of range")
)

// maxErrorQuoteBytes is the maximum number of input bytes quoted in a parse
// error message.
const maxErrorQuoteBytes = 64

// quoteExcerpt quotes text in full when it is at most maxErrorQuoteBytes bytes
// long. Otherwise it quotes the longest prefix of text that consists of whole
// decoding units (a valid UTF-8 rune or a single byte that does not start a
// valid rune) and fits in maxErrorQuoteBytes bytes, followed by an ellipsis.
// The length of text is not included, so the result depends only on the first
// maxErrorQuoteBytes+3 bytes of text.
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

// textClass is the class of a parse error: the package sentinel together
// with the set of wrapped dependency sentinels.
type textClass uint8

const (
	// classInvalidText wraps ErrInvalidText only.
	classInvalidText textClass = iota
	// classFractionRange wraps ErrOutOfRange and fraction.ErrOutOfRange.
	classFractionRange
	// classTupletCount wraps ErrOutOfRange and tuplet.ErrOutOfRange.
	classTupletCount
	// classChainRange wraps ErrOutOfRange, tuplet.ErrOutOfRange, and
	// fraction.ErrOutOfRange.
	classChainRange
)

// parseError returns an error of class c that quotes an excerpt of the
// rejected text. The message consists only of constant sentinel texts and the
// excerpt.
func parseError(c textClass, text string) error {
	q := quoteExcerpt(text)
	switch c {
	case classFractionRange:
		return fmt.Errorf("%w: %w: %s", ErrOutOfRange, fraction.ErrOutOfRange, q)
	case classTupletCount:
		return fmt.Errorf("%w: %w: %s", ErrOutOfRange, tuplet.ErrOutOfRange, q)
	case classChainRange:
		return fmt.Errorf("%w: %w: %w: %s", ErrOutOfRange, tuplet.ErrOutOfRange, fraction.ErrOutOfRange, q)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidText, q)
	}
}

// unsetError returns an error that wraps ErrUnset and names the operation and
// the unset operand.
func unsetError(op, operand string) error {
	return fmt.Errorf("%w: %s: unset %s", ErrUnset, op, operand)
}

// invalidError returns an error that wraps ErrInvalidArgument only.
func invalidError(op, detail string) error {
	return fmt.Errorf("%w: %s: %s", ErrInvalidArgument, op, detail)
}

// zeroDenominatorError returns an error that wraps ErrInvalidArgument and
// fraction.ErrZeroDenominator.
func zeroDenominatorError(op string) error {
	return fmt.Errorf("%w: %s: %w", ErrInvalidArgument, op, fraction.ErrZeroDenominator)
}

// unsetTupletError returns an error that wraps ErrInvalidArgument and
// tuplet.ErrUnset for the unset tuplet at index i of a chain.
func unsetTupletError(op string, i int) error {
	return fmt.Errorf("%w: %s: chain element %d: %w", ErrInvalidArgument, op, i, tuplet.ErrUnset)
}

// modeError returns an error that wraps ErrInvalidArgument and the sentinel
// of package common/rounding for an unset or unknown mode.
func modeError(op string, mode rounding.Mode) error {
	cause := rounding.ErrUnknownMode
	if _, err := mode.Name(); errors.Is(err, rounding.ErrUnset) {
		cause = rounding.ErrUnset
	}
	return fmt.Errorf("%w: %s: rounding mode: %w", ErrInvalidArgument, op, cause)
}

// rangeError returns an error that wraps ErrOutOfRange and
// fraction.ErrOutOfRange for a fraction outside the range.
func rangeError(op string) error {
	return fmt.Errorf("%w: %s: %w", ErrOutOfRange, op, fraction.ErrOutOfRange)
}

// chainRangeError returns an error that wraps ErrOutOfRange,
// tuplet.ErrOutOfRange, and fraction.ErrOutOfRange for a chain coefficient
// outside the range.
func chainRangeError(op string) error {
	return fmt.Errorf("%w: %s: chain coefficient: %w: %w", ErrOutOfRange, op, tuplet.ErrOutOfRange, fraction.ErrOutOfRange)
}

// wrapNotRepresentable returns an error that wraps ErrOutOfRange and
// fraction.ErrNotRepresentable for an integer result greater than
// math.MaxInt.
func wrapNotRepresentable(op string) error {
	return fmt.Errorf("%w: %s: quotient exceeds math.MaxInt: %w", ErrOutOfRange, op, fraction.ErrNotRepresentable)
}

// simpleError returns an error that wraps sentinel only.
func simpleError(sentinel error, op string) error {
	return fmt.Errorf("%w: %s", sentinel, op)
}

// must returns v and panics with err if err is not nil.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// must2 returns a and b and panics with err if err is not nil.
func must2[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		panic(err)
	}
	return a, b
}
