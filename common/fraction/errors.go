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

// quoteInputExcerpt quotes text in full when it is at most maxErrorQuoteBytes
// bytes long; otherwise it quotes at most the first maxErrorQuoteBytes bytes
// without splitting a valid UTF-8 character, followed by an ellipsis and the
// full length of text in bytes.
func quoteInputExcerpt(text string) string {
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
	return strconv.Quote(text[:k]) + "... (" + strconv.Itoa(len(text)) + " bytes)"
}

// parseError formats a text parsing error wrapping exactly one sentinel baseErr
// and quoting text in full up to maxErrorQuoteBytes bytes, or quoting at most
// its first maxErrorQuoteBytes bytes without splitting a valid UTF-8 character
// followed by its byte length.
func parseError(baseErr error, text string) error {
	return fmt.Errorf("%w: %s", baseErr, quoteInputExcerpt(text))
}
