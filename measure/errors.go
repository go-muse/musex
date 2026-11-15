package measure

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"
)

// Sentinel errors returned by package operations. Every error returned by the
// package wraps exactly one of these sentinels, so they are pairwise
// distinguishable with [errors.Is]; no sentinel wraps another. When the cause
// of a failure is an error of package meter, duration, or common/fraction,
// the returned error additionally wraps the sentinel of that package, so both
// checks with [errors.Is] hold. The message text is not part of the API,
// except for the limits on parse error messages described in the package
// documentation.
var (
	// ErrUnset indicates that the receiver or an argument is the unset grid
	// Grid{} or the unset address Address{}.
	ErrUnset = errors.New("measure: unset value")

	// ErrInvalidArgument indicates an argument outside the domain of an
	// operation: an unset meter, duration, position, or fraction; a negative
	// bar number, or a bar number less than 1 where a numbered bar is needed
	// (less than 2 for WithoutMark); a range whose first bar follows its last
	// bar; a slice of the pickup alone; deleting all numbered bars of a finite
	// grid; a number of bars or of added or inserted bars less than 1; an
	// ordinal less than 1; an empty series of meters; an invalid alignment or
	// level; a nil receiver of UnmarshalText. For an unset meter it also wraps
	// meter.ErrUnset, for an unset duration or position duration.ErrUnset, and
	// for an unset fraction fraction.ErrInvalid.
	ErrInvalidArgument = errors.New("measure: invalid argument")

	// ErrLongerThanMeter indicates a pickup longer than the size of the meter
	// of bar 1, or an explicit length aligned to the end that is longer than
	// the size of the meter of its bar, whether set directly, produced by a
	// change of the grid, given in text, or given to Complement.
	ErrLongerThanMeter = errors.New("measure: longer than the meter")

	// ErrOutsideGrid indicates a bar that the grid does not have (bar 0
	// without a pickup, a bar after the last bar of a finite grid), a position
	// or a span that reaches the end of a finite grid or goes beyond it, a
	// record after the last bar, or an ordinal greater than the number of
	// bars.
	ErrOutsideGrid = errors.New("measure: outside the grid")

	// ErrOutsideBar indicates an offset outside its bar or its meter, or a
	// count or a tstamp less than 1 or giving an offset outside the bar.
	ErrOutsideBar = errors.New("measure: outside the bar")

	// ErrUndefined indicates an operation that is not defined: a place in the
	// meter beyond the size of the meter in an extended bar, an MEI tstamp in
	// a bar that is not full or for a meter without a common denominator,
	// adding bars to an open grid, or appending to an open grid. For a meter
	// without a common denominator it also wraps meter.ErrNoCommonDenominator.
	ErrUndefined = errors.New("measure: undefined")

	// ErrGroupingRequired indicates a split by groups or by beats when the
	// meter of a spanned bar has no grouping. It always also wraps
	// meter.ErrGroupingRequired.
	ErrGroupingRequired = errors.New("measure: grouping required")

	// ErrTooMany indicates more than MaxParts parts of a split, or a series of
	// two or more meters over more than MaxAlternation bars.
	ErrTooMany = errors.New("measure: too many")

	// ErrInvalidText indicates text that does not match the grammar of the
	// package or its canonical order, a nested text that the parser of a
	// dependency rejects for a reason other than the range, or the string of
	// an unset value.
	ErrInvalidText = errors.New("measure: invalid text")

	// ErrOutOfRange indicates a result or a part of it outside the range of
	// package common/fraction, a position at or after the end of bar
	// math.MaxInt of an open grid, a bar number or a number of bars greater
	// than math.MaxInt, or a number in text greater than math.MaxInt. For a
	// fraction outside the range it also wraps fraction.ErrOutOfRange; in
	// parsing it wraps those of meter.ErrOutOfRange, duration.ErrOutOfRange,
	// and fraction.ErrOutOfRange that the failure of the nested parser wraps.
	ErrOutOfRange = errors.New("measure: value out of range")
)

// errUnset returns an error that wraps ErrUnset for an unset operand.
func errUnset(op, operand string) error {
	return fmt.Errorf("%w: %s: unset %s", ErrUnset, op, operand)
}

// errUnsetGrid returns an error that wraps ErrUnset for an unset grid
// receiver.
func errUnsetGrid(op string) error {
	return errUnset(op, "grid")
}

// errInvalid returns an error that wraps ErrInvalidArgument only.
func errInvalid(op, detail string) error {
	return fmt.Errorf("%w: %s: %s", ErrInvalidArgument, op, detail)
}

// errUnsetMeter returns an error that wraps ErrInvalidArgument and
// meter.ErrUnset for an unset meter argument.
func errUnsetMeter(op, operand string) error {
	return fmt.Errorf("%w: %s: unset %s: %w", ErrInvalidArgument, op, operand, meter.ErrUnset)
}

// errUnsetTime returns an error that wraps ErrInvalidArgument and
// duration.ErrUnset for an unset duration or position argument.
func errUnsetTime(op, operand string) error {
	return fmt.Errorf("%w: %s: unset %s: %w", ErrInvalidArgument, op, operand, duration.ErrUnset)
}

// errUnsetFraction returns an error that wraps ErrInvalidArgument and
// fraction.ErrInvalid for an unset fraction argument.
func errUnsetFraction(op, operand string) error {
	return fmt.Errorf("%w: %s: unset %s: %w", ErrInvalidArgument, op, operand, fraction.ErrInvalid)
}

// errSimple returns an error that wraps sentinel only.
func errSimple(sentinel error, op, detail string) error {
	return fmt.Errorf("%w: %s: %s", sentinel, op, detail)
}

// errCaused returns an error that wraps sentinel and cause: a failure of a
// dependency whose sentinels must stay visible to errors.Is.
func errCaused(sentinel error, op string, cause error) error {
	return fmt.Errorf("%w: %s: %w", sentinel, op, cause)
}

// errNoCommonDenominator returns an error that wraps ErrUndefined and
// meter.ErrNoCommonDenominator.
func errNoCommonDenominator(op string) error {
	return fmt.Errorf("%w: %s: MEI tstamp needs a common denominator: %w",
		ErrUndefined, op, meter.ErrNoCommonDenominator)
}

// errGrouping returns an error that wraps ErrGroupingRequired and
// meter.ErrGroupingRequired.
func errGrouping(op string) error {
	return fmt.Errorf("%w: %s: the meter of a spanned bar has no grouping: %w",
		ErrGroupingRequired, op, meter.ErrGroupingRequired)
}

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
// with the set of wrapped dependency sentinels of the range.
type textClass struct {
	sentinel error
	// meterRange, durationRange, and fractionRange select the wrapped range
	// sentinels of the dependencies; they are used with ErrOutOfRange only.
	meterRange, durationRange, fractionRange bool
}

// Classes of parse errors without dependency sentinels.
var (
	classInvalidText = textClass{sentinel: ErrInvalidText}
	classNumberRange = textClass{sentinel: ErrOutOfRange}
	classLonger      = textClass{sentinel: ErrLongerThanMeter}
	classOutsideGrid = textClass{sentinel: ErrOutsideGrid}
)

// nestedClass classifies a failure of the parser of a dependency: a failure
// because of the range is ErrOutOfRange together with the range sentinels of
// the dependencies that the failure wraps; any other failure is
// ErrInvalidText.
func nestedClass(err error) textClass {
	c := textClass{
		sentinel:      ErrOutOfRange,
		meterRange:    errors.Is(err, meter.ErrOutOfRange),
		durationRange: errors.Is(err, duration.ErrOutOfRange),
		fractionRange: errors.Is(err, fraction.ErrOutOfRange),
	}
	if !c.meterRange && !c.durationRange && !c.fractionRange {
		return classInvalidText
	}
	return c
}

// parseError returns an error of class c that quotes an excerpt of the
// rejected text. The message consists only of constant sentinel texts and the
// excerpt, and the error wraps only sentinels.
func parseError(c textClass, text string) error {
	format := "%w"
	args := []any{c.sentinel}
	if c.meterRange {
		format += ": %w"
		args = append(args, meter.ErrOutOfRange)
	}
	if c.durationRange {
		format += ": %w"
		args = append(args, duration.ErrOutOfRange)
	}
	if c.fractionRange {
		format += ": %w"
		args = append(args, fraction.ErrOutOfRange)
	}
	args = append(args, quoteExcerpt(text))
	return fmt.Errorf(format+": %s", args...)
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
