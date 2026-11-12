package tuplet

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"
)

// Sentinel errors returned by package operations. Every error returned by the
// package wraps exactly one of these sentinels, so they are pairwise
// distinguishable with [errors.Is]; no sentinel wraps another. When the cause
// of a failure is an error of package common/fraction, the returned error
// additionally wraps that cause, so both the tuplet sentinel and the fraction
// sentinel match with [errors.Is]. The message text is not part of the API,
// except for the limits on parse error messages described in the package
// documentation.
var (
	// ErrInvalidCount indicates an actual or normal count less than 1 at
	// creation, or a zero or negative coefficient passed to FromCoefficient
	// or FromReciprocalCoefficient.
	ErrInvalidCount = errors.New("tuplet: counts must be at least 1")

	// ErrUnset indicates that a receiver, an argument, or a chain element is
	// the unset zero value Tuplet{}.
	ErrUnset = errors.New("tuplet: unset tuplet")

	// ErrNoDefault indicates that no default normal count exists for the
	// actual count and division, or that the division is unknown.
	ErrNoDefault = errors.New("tuplet: no default normal count")

	// ErrInvalidText indicates text that does not match the tuplet grammar or
	// is not a known tuplet name.
	ErrInvalidText = errors.New("tuplet: invalid tuplet text")

	// ErrOutOfRange indicates a result that cannot be represented: an integer
	// greater than math.MaxInt, or a fraction outside the range of package
	// common/fraction. For fraction results it also wraps
	// fraction.ErrOutOfRange; for the integer result of FromUnitRatio it also
	// wraps fraction.ErrNotRepresentable.
	ErrOutOfRange = errors.New("tuplet: value out of range")

	// ErrInvalidArgument indicates an unset fraction argument, a ratio that
	// is not positive, a scale factor less than 1, or a nil receiver of
	// UnmarshalText.
	ErrInvalidArgument = errors.New("tuplet: invalid argument")

	// ErrNotRepresentable indicates that normal·ratio in FromUnitRatio is not
	// an integer. It is distinct from fraction.ErrNotRepresentable, which
	// FromUnitRatio wraps together with ErrOutOfRange when normal·ratio is an
	// integer greater than math.MaxInt.
	ErrNotRepresentable = errors.New("tuplet: normal count times ratio is not an integer")
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

// parseError returns an error that wraps sentinel and quotes an excerpt of
// the rejected text.
func parseError(sentinel error, text string) error {
	return fmt.Errorf("%w: %s", sentinel, quoteExcerpt(text))
}

// unsetError returns an error that wraps ErrUnset and names the unset operand.
func unsetError(operand string) error {
	return fmt.Errorf("%w: %s", ErrUnset, operand)
}

// fractionRangeError returns an error that wraps ErrOutOfRange and the cause
// returned by package common/fraction.
func fractionRangeError(op string, cause error) error {
	return fmt.Errorf("%w: %s: %w", ErrOutOfRange, op, cause)
}

// must returns v and panics with err if err is not nil.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
