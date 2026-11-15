package interval

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/system"
)

// Sentinel errors returned by package operations. Every error returned by the
// package matches at most one of these sentinels with [errors.Is]; no sentinel
// wraps another. When the cause of a failure is an error of a dependency, the
// returned error also wraps that error, so the checks for the sentinel of the
// dependency hold too. Failures that are causes of package system or note only,
// such as a name that is not admissible in a system, match no sentinel of this
// package. The message text is not part of the API, except for the limits on
// parse error messages described in the package documentation.
var (
	// ErrUnset indicates an unset input: the zero Interval, an unset name,
	// pitch, or note, a nil or unset system, the quality QualityUnset in New
	// and NewMultiple, a nil receiver of UnmarshalText, or a value of Replace
	// other than ReplaceFrom, ReplaceTo, and ReplaceBoth.
	ErrUnset = errors.New("interval: unset value")

	// ErrInvalidArgument indicates a combination of parts or a parameter
	// that has no interval: a size less than 1, a quality that does not fit
	// the group of the size, a multiplicity less than 1, a diminished unison
	// (also as the result of a change of quality), a negative number of
	// octaves, a negative limit of multiplicity or alteration, or a parsed
	// text whose parts do not form an interval.
	ErrInvalidArgument = errors.New("interval: no such interval or parameter")

	// ErrNoDirection indicates a direction that is not Up or Down where an
	// operation needs one.
	ErrNoDirection = errors.New("interval: direction not set")

	// ErrNoInversion indicates an interval whose simple part is wider than a
	// perfect octave (an augmented octave of any multiplicity, or an octave
	// with a positive fractional deviation): its inversion would be a
	// diminished unison.
	ErrNoInversion = errors.New("interval: inversion undefined")

	// ErrNoOctave indicates two notes of which one has an octave and the
	// other has none.
	ErrNoOctave = errors.New("interval: note without octave")

	// ErrNoSystem indicates a note with an octave that is not bound to a
	// source of settings when no explicit system is given.
	ErrNoSystem = errors.New("interval: no system for a note")

	// ErrDifferentSystems indicates notes whose settings have different
	// systems.
	ErrDifferentSystems = errors.New("interval: notes in different systems")

	// ErrStartRequired indicates a positional operation in a system that is
	// not regular when no start name is given.
	ErrStartRequired = errors.New("interval: start name required")

	// ErrFractionalSteps indicates a positional size that is not a whole
	// number of steps of the system.
	ErrFractionalSteps = errors.New("interval: non-integer number of steps")

	// ErrParse indicates text that is neither the short notation nor a name
	// of an interval.
	ErrParse = errors.New("interval: cannot parse")

	// ErrNotOnFifths indicates a name with a fractional alteration value or
	// an interval with a fractional quality, which have no place on the line
	// of fifths.
	ErrNotOnFifths = errors.New("interval: no place on the line of fifths")

	// ErrOutOfRange indicates a result that cannot be represented: a size
	// greater than MaxSize, a deviation or an alteration value outside the
	// range of package alteration, or a written octave, an octave, or a number
	// of steps that does not fit in an int. When a dependency reports the
	// cause, the error also wraps alteration.ErrOutOfRange,
	// octave.ErrOutOfRange, or system.ErrOutOfRange.
	ErrOutOfRange = errors.New("interval: result out of range")
)

// opError returns an error that wraps sentinel only and names the operation.
func opError(op string, sentinel error) error {
	return fmt.Errorf("interval: %s: %w", op, sentinel)
}

// detailError returns an error that wraps sentinel only and names the
// operation and the detail of the failure.
func detailError(op string, sentinel error, detail string) error {
	return fmt.Errorf("interval: %s: %s: %w", op, detail, sentinel)
}

// rangeError returns an error that wraps ErrOutOfRange and the error of the
// dependency that reported the cause.
func rangeError(op string, cause error) error {
	return fmt.Errorf("interval: %s: %w: %w", op, ErrOutOfRange, cause)
}

// alterationRangeError returns an error that wraps ErrOutOfRange and
// alteration.ErrOutOfRange for a deviation or an alteration value outside the
// range of package alteration.
func alterationRangeError(op string) error {
	return fmt.Errorf("interval: %s: %w: %w", op, ErrOutOfRange, alteration.ErrOutOfRange)
}

// causeError returns an error that wraps the error of a dependency, without a
// sentinel of this package.
func causeError(op string, cause error) error {
	return fmt.Errorf("interval: %s: %w", op, cause)
}

// inadmissibleError returns an error that wraps system.ErrInadmissible for a
// name that is not admissible in an explicit system.
func inadmissibleError(op, what string) error {
	return fmt.Errorf("interval: %s: %s: %w", op, what, system.ErrInadmissible)
}

// noteInadmissibleError returns an error that wraps note.ErrInadmissible and
// system.ErrInadmissible for a name that is not admissible in the system of
// the settings of a note.
func noteInadmissibleError(op, what string) error {
	return fmt.Errorf("interval: %s: %s: %w: %w", op, what, note.ErrInadmissible, system.ErrInadmissible)
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
// with the set of wrapped dependency sentinels.
type textClass uint8

const (
	// classParse wraps ErrParse only.
	classParse textClass = iota
	// classInvalid wraps ErrInvalidArgument only.
	classInvalid
	// classSizeRange wraps ErrOutOfRange only: the size exceeds MaxSize.
	classSizeRange
	// classDeviationRange wraps ErrOutOfRange and alteration.ErrOutOfRange.
	classDeviationRange
)

// parseError returns an error of class c that quotes an excerpt of the
// rejected text. The message consists only of constant sentinel texts and the
// excerpt, and the error keeps no reference to text.
func parseError(c textClass, text string) error {
	q := quoteExcerpt(text)
	switch c {
	case classInvalid:
		return fmt.Errorf("%w: %s", ErrInvalidArgument, q)
	case classSizeRange:
		return fmt.Errorf("%w: %s", ErrOutOfRange, q)
	case classDeviationRange:
		return fmt.Errorf("%w: %w: %s", ErrOutOfRange, alteration.ErrOutOfRange, q)
	default:
		return fmt.Errorf("%w: %s", ErrParse, q)
	}
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
