package ticks

import (
	"errors"
	"math/big"
	"strconv"
	"unicode/utf8"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
)

// Sentinel errors returned by package operations. Every returned error wraps
// exactly one of them; no sentinel wraps another. When the cause is an error of
// a dependency, the returned error also wraps the sentinels of that error.
var (
	// ErrUnset reports an unset value of the package as a receiver or an
	// argument.
	ErrUnset = errors.New("ticks: unset value")
	// ErrInvalidArgument reports an argument outside the domain of an
	// operation: an unset value of a dependency, a negative tick, bar, index or
	// count, a duration that is not positive, a rounding mode, an invalid
	// enumeration, an extra variadic argument, a frame code, code -29 without a
	// reading, step parameters, a reversed range, a missing change, different
	// MIDI quarter notations, or an address, grid or position that a dependency
	// rejects.
	ErrInvalidArgument = errors.New("ticks: invalid argument")
	// ErrZero reports a zero resolution, ticks per frame, tempo, quarters per
	// minute, or a zero field of a MIDI time signature.
	ErrZero = errors.New("ticks: zero value")
	// ErrNotRepresentable reports a value that is not a whole number of units:
	// ticks, a rescaled tick, a MIDI quarter notation, or a step off the MIDI
	// clock.
	ErrNotRepresentable = errors.New("ticks: not representable")
	// ErrFormatLimit reports a value outside a limit of a format: an SMF
	// resolution, a Set Tempo value, a delta-time, or a byte field.
	ErrFormatLimit = errors.New("ticks: outside the format limit")
	// ErrOrder reports decreasing absolute ticks or changes of a MIDI map out
	// of tick order.
	ErrOrder = errors.New("ticks: wrong order")
	// ErrNotConvertible reports a musex tempo map that cannot be converted as
	// requested: a gradual change without steps, an unfinished gradual change,
	// a fermata hold that is refused or cannot be stretched, or a position
	// outside the grid.
	ErrNotConvertible = errors.New("ticks: not convertible")
	// ErrOutOfRange reports a result outside int or the range of
	// common/fraction, or a rounding that could not be decided.
	ErrOutOfRange = errors.New("ticks: value out of range")
	// ErrInvalidText reports text that does not match the grammar.
	ErrInvalidText = errors.New("ticks: invalid text")
	// ErrMeterNotMIDI reports a time signature that a MIDI Time Signature
	// event cannot express.
	ErrMeterNotMIDI = errors.New("ticks: meter not expressible in MIDI")
)

// opError is an error of an operation: a constant message and its causes.
type opError struct {
	message string
	causes  []error
}

func (e *opError) Error() string { return e.message }

func (e *opError) Unwrap() []error { return e.causes }

func newError(sentinel error, op, detail string, causes ...error) error {
	message := sentinel.Error() + ": " + op + ": " + detail
	for _, cause := range causes {
		message += ": " + cause.Error()
	}
	all := make([]error, 0, len(causes)+1)
	all = append(all, sentinel)
	all = append(all, causes...)
	return &opError{message: message, causes: all}
}

func unsetError(op, what string) error { return newError(ErrUnset, op, what+" is unset") }

func invalidError(op, detail string, causes ...error) error {
	return newError(ErrInvalidArgument, op, detail, causes...)
}

func rangeError(op, detail string) error {
	return newError(ErrOutOfRange, op, detail, fraction.ErrOutOfRange)
}

func intRangeError(op, detail string) error { return newError(ErrOutOfRange, op, detail) }

func modeError(op string, mode rounding.Mode) error {
	cause := rounding.ErrUnknownMode
	if mode == 0 {
		cause = rounding.ErrUnset
	}
	return invalidError(op, "rounding mode", cause)
}

func checkMode(op string, mode rounding.Mode) error {
	if !mode.IsValid() {
		return modeError(op, mode)
	}
	return nil
}

func isRangeFailure(err error) bool {
	for _, sentinel := range []error{
		fraction.ErrOutOfRange, duration.ErrOutOfRange, tempo.ErrOutOfRange,
		tempomap.ErrOutOfRange, meter.ErrOutOfRange, measure.ErrOutOfRange,
	} {
		if errors.Is(err, sentinel) {
			return true
		}
	}
	return false
}

// dependencyError maps a failure of a dependency: a range failure is
// ErrOutOfRange together with fraction.ErrOutOfRange, an unfinished gradual
// change is ErrNotConvertible, and other failures are ErrInvalidArgument. The
// chain of the dependency is preserved.
func dependencyError(op string, err error) error {
	switch {
	case isRangeFailure(err):
		if errors.Is(err, fraction.ErrOutOfRange) || errors.Is(err, fraction.ErrNotRepresentable) {
			return newError(ErrOutOfRange, op, "dependency result", err)
		}
		return newError(ErrOutOfRange, op, "dependency result", err, fraction.ErrOutOfRange)
	case errors.Is(err, tempomap.ErrUnfinished):
		return newError(ErrNotConvertible, op, "unfinished gradual change", err)
	default:
		return invalidError(op, "dependency argument", err)
	}
}

// meterError maps a failure of package meter while a time signature is
// translated: missing grouping, undefined operations and missing common
// denominators are ErrMeterNotMIDI.
func meterError(op string, err error) error {
	if errors.Is(err, meter.ErrGroupingRequired) || errors.Is(err, meter.ErrUndefined) ||
		errors.Is(err, meter.ErrNoCommonDenominator) || errors.Is(err, measure.ErrGroupingRequired) ||
		errors.Is(err, measure.ErrUndefined) {
		return newError(ErrMeterNotMIDI, op, "time signature", err)
	}
	return dependencyError(op, err)
}

// ---- text errors ----

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

// textError is a parse error: the constant texts of its sentinels and a
// bounded excerpt of the input, without references to the input.
type textError struct {
	message string
	causes  []error
}

func (e *textError) Error() string { return e.message }

func (e *textError) Unwrap() []error { return e.causes }

func grammarError(text string) error {
	return &textError{message: ErrInvalidText.Error() + ": " + quoteExcerpt(text), causes: []error{ErrInvalidText}}
}

// sentinelsOf returns the package sentinel of err and the dependency sentinels
// that it wraps, in a fixed order.
func sentinelsOf(err error) []error {
	var out []error
	for _, s := range []error{
		ErrUnset, ErrInvalidArgument, ErrZero, ErrNotRepresentable, ErrFormatLimit,
		ErrOrder, ErrNotConvertible, ErrOutOfRange, ErrInvalidText, ErrMeterNotMIDI,
	} {
		if errors.Is(err, s) {
			out = append(out, s)
			break
		}
	}
	for _, s := range []error{
		fraction.ErrOutOfRange, fraction.ErrInvalid, fraction.ErrNotRepresentable,
		rounding.ErrUnset, rounding.ErrUnknownMode,
		duration.ErrUnset, duration.ErrOutOfRange, tempo.ErrUnset, tempo.ErrOutOfRange,
		meter.ErrUnset, meter.ErrOutOfRange, measure.ErrUnset, measure.ErrOutOfRange,
	} {
		if errors.Is(err, s) {
			out = append(out, s)
		}
	}
	return out
}

// valueTextError reports a failed value check of parsed text: the sentinels of
// err and the excerpt of text.
func valueTextError(text string, err error) error {
	causes := sentinelsOf(err)
	if len(causes) == 0 {
		return grammarError(text)
	}
	return &textError{message: causes[0].Error() + ": " + quoteExcerpt(text), causes: causes}
}

func numberRangeTextError(text string) error {
	return &textError{
		message: ErrOutOfRange.Error() + ": " + quoteExcerpt(text),
		causes:  []error{ErrOutOfRange, fraction.ErrOutOfRange},
	}
}

// ---- Must helpers ----

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

// ---- exact arithmetic helpers ----

func ratOfFraction(f fraction.Fraction) *big.Rat {
	r, _ := f.BigRat()
	return r
}

func ratOfDuration(d duration.Duration) *big.Rat { return ratOfFraction(d.MustFraction()) }

func ratOfPosition(p duration.Position) *big.Rat { return ratOfFraction(p.MustFraction()) }

func ratOfSeconds(s tempo.Seconds) *big.Rat { return ratOfFraction(s.MustFraction()) }

func ratInt(n int) *big.Rat { return new(big.Rat).SetInt64(int64(n)) }

func fractionOf(op string, value *big.Rat) (fraction.Fraction, error) {
	f, err := fraction.FromBigRat(value)
	if err != nil {
		return fraction.Fraction{}, rangeError(op, "result outside common/fraction")
	}
	return f, nil
}

func durationOf(op string, value *big.Rat) (duration.Duration, error) {
	f, err := fractionOf(op, value)
	if err != nil {
		return duration.Duration{}, err
	}
	d, err := duration.FromFraction(f)
	if err != nil {
		return duration.Duration{}, dependencyError(op, err)
	}
	return d, nil
}

func positionOf(op string, value *big.Rat) (duration.Position, error) {
	f, err := fractionOf(op, value)
	if err != nil {
		return duration.Position{}, err
	}
	p, err := duration.PositionFromFraction(f)
	if err != nil {
		return duration.Position{}, dependencyError(op, err)
	}
	return p, nil
}

func secondsOf(op string, value *big.Rat) (tempo.Seconds, error) {
	f, err := fractionOf(op, value)
	if err != nil {
		return tempo.Seconds{}, err
	}
	s, err := tempo.SecondsFromFraction(f)
	if err != nil {
		return tempo.Seconds{}, dependencyError(op, err)
	}
	return s, nil
}

// intOf returns the integer value of an integral rational or reports that it
// is outside int.
func intOf(op string, value *big.Int) (int, error) {
	if !value.IsInt64() {
		return 0, intRangeError(op, "result outside int")
	}
	return int(value.Int64()), nil
}

// roundInt rounds an exact value with mode and returns it as int.
func roundInt(op string, value *big.Rat, mode rounding.Mode) (*big.Int, error) {
	n, err := mode.RatInt(value)
	if err != nil {
		return nil, invalidError(op, "rounding", err)
	}
	return n, nil
}
