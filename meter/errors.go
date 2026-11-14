package meter

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
)

// Sentinel errors returned by package operations. Every error returned by the
// package wraps exactly one of these sentinels, so they are pairwise
// distinguishable with [errors.Is]; no sentinel wraps another. When the cause
// of a failure is an error of package duration or common/fraction, the
// returned error additionally wraps the sentinel of that package, so both
// checks with [errors.Is] hold. The message text is not part of the API,
// except for the limits on parse error messages described in the package
// documentation.
var (
	// ErrUnset indicates that the receiver or a meter argument is the unset
	// value Meter{}. An unset part of NewSum is ErrInvalidArgument instead.
	ErrUnset = errors.New("meter: unset value")

	// ErrInvalidArgument indicates an argument outside the domain of an
	// operation: a numerator, an addend, a denominator, or a group less than
	// 1; a compound numerator of fewer than two addends; a sum of fewer than
	// two parts; a part of a sum that is unset, is not a simple signature or a
	// compound numerator, or has an explicit grouping; an explicit grouping of
	// a compound numerator or a sum; an empty list of groups; groups whose sum
	// differs from the numerator; an invalid symbol; an unset duration or
	// position argument or an unset field of a Span; a size that is not a
	// multiple of the unit; a number of units less than 1; an invalid value of
	// an enumeration; a nil receiver of UnmarshalText; a group in text greater
	// than math.MaxInt or with a sum that differs from the numerator. For an
	// unset duration or position it also wraps duration.ErrUnset.
	ErrInvalidArgument = errors.New("meter: invalid argument")

	// ErrGroupingRequired indicates that an operation needs the grouping of a
	// meter that has none, such as 5/8: its groups, beats, strong beats,
	// metric weights, a group or a beat by offset or by number, the record as
	// a compound numerator or a sum, or a kind that cannot be determined
	// without a grouping.
	ErrGroupingRequired = errors.New("meter: grouping required")

	// ErrNoCommonDenominator indicates that an operation needs the common
	// denominator of a sum whose parts have different denominators, such as
	// 2/4+3/8.
	ErrNoCommonDenominator = errors.New("meter: no common denominator")

	// ErrUndefined indicates an operation that is not defined for the meter:
	// the beat division of an asymmetrical or undetermined kind, a record
	// form that cannot keep the structure, doubling the unit of an odd
	// denominator, or the power of two of a denominator that is not a power
	// of two.
	ErrUndefined = errors.New("meter: undefined")

	// ErrNotNotatable indicates that a unit, a beat, or a group is not
	// written as a single sign without tuplets. It always also wraps
	// duration.ErrNotNotatable.
	ErrNotNotatable = errors.New("meter: not notatable as a single sign")

	// ErrOutsideMeter indicates a position not less than the size of the
	// meter, or a number of a unit, a group, or a beat less than 1 or greater
	// than their count.
	ErrOutsideMeter = errors.New("meter: outside the meter")

	// ErrTooMany indicates more than MaxUnits units, including a numerator
	// or an addend in text greater than math.MaxInt, or more than
	// MaxGroupingVariants grouping variants.
	ErrTooMany = errors.New("meter: too many")

	// ErrInvalidText indicates text that does not match the grammar of the
	// package or the string of the unset value.
	ErrInvalidText = errors.New("meter: invalid text")

	// ErrOutOfRange indicates a size, a start, an offset, the end of a Span,
	// or a remainder outside the range of package common/fraction, a quotient
	// greater than math.MaxInt, or a denominator greater than math.MaxInt. It
	// also wraps fraction.ErrOutOfRange for a fraction outside the range and
	// fraction.ErrNotRepresentable for a quotient; when the cause is a failure
	// of package duration, it wraps the error of that package with its
	// duration.ErrOutOfRange. A denominator greater than math.MaxInt wraps no
	// other sentinel.
	ErrOutOfRange = errors.New("meter: value out of range")
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
	// classDenominator wraps ErrOutOfRange only: a denominator greater than
	// math.MaxInt.
	classDenominator
	// classTooMany wraps ErrTooMany only.
	classTooMany
	// classInvalidArgument wraps ErrInvalidArgument only.
	classInvalidArgument
	// classFractionRange wraps ErrOutOfRange and fraction.ErrOutOfRange.
	classFractionRange
)

// parseError returns an error of class c that quotes an excerpt of the
// rejected text. The message consists only of constant sentinel texts and the
// excerpt.
func parseError(c textClass, text string) error {
	q := quoteExcerpt(text)
	switch c {
	case classDenominator:
		return fmt.Errorf("%w: %s", ErrOutOfRange, q)
	case classTooMany:
		return fmt.Errorf("%w: %s", ErrTooMany, q)
	case classInvalidArgument:
		return fmt.Errorf("%w: %s", ErrInvalidArgument, q)
	case classFractionRange:
		return fmt.Errorf("%w: %w: %s", ErrOutOfRange, fraction.ErrOutOfRange, q)
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

// unsetInputError returns an error that wraps ErrInvalidArgument and
// duration.ErrUnset for an unset duration or position argument.
func unsetInputError(op, operand string) error {
	return fmt.Errorf("%w: %s: unset %s: %w", ErrInvalidArgument, op, operand, duration.ErrUnset)
}

// simpleError returns an error that wraps sentinel only.
func simpleError(sentinel error, op, detail string) error {
	return fmt.Errorf("%w: %s: %s", sentinel, op, detail)
}

// notNotatableError returns an error that wraps ErrNotNotatable and the
// failure of duration.Duration.ToWritten, which wraps
// duration.ErrNotNotatable.
func notNotatableError(op string, cause error) error {
	return fmt.Errorf("%w: %s: %w", ErrNotNotatable, op, cause)
}

// causedRangeError returns an error that wraps ErrOutOfRange and cause: a
// failure of common/fraction or of duration whose sentinels must stay
// visible to errors.Is.
func causedRangeError(op string, cause error) error {
	return fmt.Errorf("%w: %s: %w", ErrOutOfRange, op, cause)
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
