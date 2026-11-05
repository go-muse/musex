package fraction

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"
)

// Sentinel errors returned by package operations. Each returned error wraps or
// equals at most one of these sentinels, so they are pairwise distinguishable
// with [errors.Is].
var (
	// ErrInvalid indicates that a fraction operand is the uninitialized zero
	// value Fraction{} (0/0), or that a required pointer receiver or input
	// pointer (such as num in FromBigInt, r in FromBigRat, or the receiver of
	// UnmarshalText) is nil.
	ErrInvalid = errors.New("fraction: invalid zero-value fraction or nil argument")

	// ErrZeroDenominator indicates that a denominator is zero (in constructors
	// or parsed text), or that a target denominator or floor step is not
	// strictly positive (<= 0).
	ErrZeroDenominator = errors.New("fraction: denominator or step must be non-zero / positive")

	// ErrDivisionByZero indicates division by a zero fraction or zero integer,
	// taking the reciprocal of zero, raising zero to a negative power, or
	// dividing by a zero step.
	ErrDivisionByZero = errors.New("fraction: division by zero")

	// ErrNotRepresentable indicates that an exact rational value cannot be
	// represented without remainder in the requested denominator or narrow int
	// type, or that an unknown DenomMode, unknown RoundingMode, or out-of-range
	// decimal digit count was supplied.
	ErrNotRepresentable = errors.New("fraction: value is not representable in requested form")

	// ErrOutOfRange indicates that a requested numerator or denominator magnitude
	// exceeds MaxComponent (2^255 - 1), or that a parsed digit group exceeds
	// MaxDigits (100 digits).
	ErrOutOfRange = errors.New("fraction: component exceeds guaranteed range")

	// ErrInvalidText indicates that an input string does not conform to any of
	// the supported fraction syntax forms.
	ErrInvalidText = errors.New("fraction: invalid fraction text")
)

const maxErrorQuoteBytes = 64

// parseError formats a text parsing error wrapping exactly one sentinel baseErr
// while bounding the quoted excerpt of text to at most 64 bytes plus the full
// byte length.
func parseError(baseErr error, text string) error {
	truncated := false
	limit := len(text)
	if limit > maxErrorQuoteBytes {
		limit = maxErrorQuoteBytes
		for limit > 0 && !utf8.RuneStart(text[limit]) {
			limit--
		}
		truncated = true
	}
	quoted := strconv.Quote(text[:limit])
	for len(quoted) > maxErrorQuoteBytes && limit > 0 {
		limit--
		for limit > 0 && !utf8.RuneStart(text[limit]) {
			limit--
		}
		quoted = strconv.Quote(text[:limit])
		truncated = true
	}
	if !truncated {
		return fmt.Errorf("%w: %s (len=%d)", baseErr, quoted, len(text))
	}
	return fmt.Errorf("%w: %s... (len=%d)", baseErr, quoted, len(text))
}
