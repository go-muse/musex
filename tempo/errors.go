package tempo

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
)

// Sentinel errors returned by package operations. Every error returned by the
// package wraps exactly one of these sentinels, so they are pairwise
// distinguishable with [errors.Is]; no sentinel wraps another. When the cause
// of a failure is an error of package common/fraction, common/rounding, or
// duration, the returned error additionally wraps the sentinels of that
// package, so all checks with [errors.Is] hold. The sentinels of package
// tempo/term are never wrapped. The message text is not part of the API,
// except for the limits on parse error messages described in the package
// documentation.
var (
	// ErrUnset indicates that a receiver or an argument is an unset value of
	// the package: Tempo{}, Approx{}, Range{}, Seconds{}, or Indication{}.
	ErrUnset = errors.New("tempo: unset value")

	// ErrInvalidArgument indicates an argument outside the domain of an
	// operation: an unset value of another package (a fraction, a duration, a
	// written duration, a term, or a list), a zero denominator, a factor or a
	// grid step that is not positive, an unset or unknown rounding mode, a
	// lower number of a range not less than the upper one, an index of a
	// division outside 0..38, an indication that would have neither a term nor
	// a marking, or a nil receiver of UnmarshalText. For a zero denominator it
	// also wraps fraction.ErrZeroDenominator, and for a mode rounding.ErrUnset
	// or rounding.ErrUnknownMode.
	ErrInvalidArgument = errors.New("tempo: invalid argument")

	// ErrNoTempo indicates beats per minute or a number of a range that is
	// zero or negative when a marking is created or parsed, and a number of
	// beats that is zero or negative in an operation of the metronome scale or
	// in Range.Contains.
	ErrNoTempo = errors.New("tempo: no tempo")

	// ErrNoSolution indicates that the formula of bars, units, beats per
	// minute, and time has no solution: a number of bars, a number of units
	// per bar, beats per minute, or bars per minute that is not positive, a
	// time of zero where a positive one is needed, or a negative number of
	// units.
	ErrNoSolution = errors.New("tempo: no solution")

	// ErrNegativeTime indicates a time less than zero, at creation or as the
	// result of Seconds.Sub.
	ErrNegativeTime = errors.New("tempo: negative time")

	// ErrOffScale indicates a number outside the metronome scale from 40 to
	// 208 where a neighbor is needed, no next division in the given
	// direction, or a number that is not a division where one is needed.
	ErrOffScale = errors.New("tempo: off the metronome scale")

	// ErrNoRange indicates a term whose guide cannot be a range: it has no
	// guide, an open bound, or equal bounds.
	ErrNoRange = errors.New("tempo: no range")

	// ErrQuantizedToZero indicates that quantized beats per minute are zero.
	ErrQuantizedToZero = errors.New("tempo: quantized to zero")

	// ErrTimeOverflow indicates that a time rounded to nanoseconds does not
	// fit in time.Duration.
	ErrTimeOverflow = errors.New("tempo: time overflows time.Duration")

	// ErrInvalidText indicates text that does not match the grammar of the
	// operation, the string of an unset value, or a name that is not a name of
	// the given list.
	ErrInvalidText = errors.New("tempo: invalid text")

	// ErrOutOfRange indicates a result outside the range of package
	// common/fraction, or a number or a unit in text outside its range. It
	// also wraps fraction.ErrOutOfRange for a result or a number, and all
	// sentinels of the error of duration.ParseWritten for a unit.
	ErrOutOfRange = errors.New("tempo: value out of range")
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
	// classNoTempo wraps ErrNoTempo only.
	classNoTempo
	// classInvalidArgument wraps ErrInvalidArgument only.
	classInvalidArgument
)

// parseError returns an error of class c that quotes an excerpt of the
// rejected text. The message consists only of constant sentinel texts and the
// excerpt.
func parseError(c textClass, text string) error {
	q := quoteExcerpt(text)
	switch c {
	case classFractionRange:
		return fmt.Errorf("%w: %w: %s", ErrOutOfRange, fraction.ErrOutOfRange, q)
	case classNoTempo:
		return fmt.Errorf("%w: %s", ErrNoTempo, q)
	case classInvalidArgument:
		return fmt.Errorf("%w: %s", ErrInvalidArgument, q)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidText, q)
	}
}

// unitRangeError is the error of a unit in text that is outside its range.
// Its message is the text of ErrOutOfRange and an excerpt of the whole
// rejected text; it wraps ErrOutOfRange and the error of duration.ParseWritten
// without including the text of that error, so every sentinel of that error
// matches with errors.Is.
type unitRangeError struct {
	quote string
	cause error
}

func (e *unitRangeError) Error() string {
	return ErrOutOfRange.Error() + ": " + e.quote
}

func (e *unitRangeError) Unwrap() []error {
	return []error{ErrOutOfRange, e.cause}
}

// unitRange returns the error of a unit outside its range in text.
func unitRange(text string, cause error) error {
	return &unitRangeError{quote: quoteExcerpt(text), cause: cause}
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

// modeError returns an error that wraps ErrInvalidArgument and the sentinel
// of package common/rounding for an unset or unknown mode.
func modeError(op string, mode rounding.Mode) error {
	cause := rounding.ErrUnknownMode
	if _, err := mode.Name(); errors.Is(err, rounding.ErrUnset) {
		cause = rounding.ErrUnset
	}
	return fmt.Errorf("%w: %s: rounding mode: %w", ErrInvalidArgument, op, cause)
}

// checkMode returns the error of an unset or unknown mode, or nil.
func checkMode(op string, mode rounding.Mode) error {
	if !mode.IsValid() {
		return modeError(op, mode)
	}
	return nil
}

// rangeError returns an error that wraps ErrOutOfRange and
// fraction.ErrOutOfRange for a result outside the range.
func rangeError(op string) error {
	return fmt.Errorf("%w: %s: %w", ErrOutOfRange, op, fraction.ErrOutOfRange)
}

// simpleError returns an error that wraps sentinel only.
func simpleError(sentinel error, op, detail string) error {
	return fmt.Errorf("%w: %s: %s", sentinel, op, detail)
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

// must3 returns a, b, and c and panics with err if err is not nil.
func must3[A, B, C any](a A, b B, c C, err error) (A, B, C) {
	if err != nil {
		panic(err)
	}
	return a, b, c
}
