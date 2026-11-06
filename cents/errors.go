package cents

import (
	"errors"
	"fmt"
)

// Errors returned by the package. Each returned error matches exactly one of
// them with [errors.Is]; it is either the error itself or wraps it with
// context. The errors do not wrap each other, and their text is not part of
// the API, except for the hint described at [Parse].
//
// When several conditions apply at once, operations on values, numbers and
// compositions report the first one in the order [ErrInvalidArgument],
// [ErrPrimeLimit], [ErrTooComplex], [ErrOutOfRange]. Texts are checked in the
// order [ErrInvalidText], [ErrTooComplex] for the length limits,
// [ErrPrimeLimit], [ErrTooComplex] for the exact record, [ErrOutOfRange].
var (
	// ErrOutOfRange reports that the exact result is outside the range
	// -MaxMagnitude to MaxMagnitude cents, that the correctly rounded result
	// of [Cents.Multiplier] or [Cents.ApplyTo] is not a normal positive finite
	// float64 (overflow, underflow to zero, or the subnormal range), or that
	// the step number of [Cents.NearestStep] does not fit in an int.
	ErrOutOfRange = errors.New("cents: value out of range")

	// ErrTooComplex reports that the exact record of the result exceeds the
	// record limits (a reduced denominator greater than 10^30, or a prime
	// exponent greater than 10^24 in magnitude), that a text contains a run
	// of more than MaxDigits digits, or that a product of powers in a text
	// has more than MaxFactors factors.
	ErrTooComplex = errors.New("cents: exact value too complex")

	// ErrPrimeLimit reports that an exact path, such as a ratio of integers,
	// a text ratio, a base of a product of powers, a Scala ratio, or a base
	// of a composition, involves a prime greater than MaxPrime.
	ErrPrimeLimit = errors.New("cents: prime factor greater than 31")

	// ErrInvalidArgument reports an invalid argument: NaN or an infinity, a
	// ratio or a frequency that is not positive, a zero denominator, a number
	// of steps or a rounding denominator less than 1, a zero divisor, a zero
	// value as the divisor of [Cents.Quo], a negative tolerance, or a base of
	// a composition that is not a prime.
	ErrInvalidArgument = errors.New("cents: invalid argument")

	// ErrInvalidText reports a text that does not match the grammar of
	// [Parse] or of [ParseScala], including a zero in a ratio, a zero
	// denominator of a fraction or of an exponent, and a base 0 or 1 of a
	// product of powers.
	ErrInvalidText = errors.New("cents: invalid text")
)

// maxQuoted is the largest number of characters of an input quoted in an
// error message.
const maxQuoted = 64

// errorf returns an error that wraps sentinel with context. The message
// starts with "cents: ", as the messages of the sentinel errors do.
func errorf(sentinel error, format string, args ...any) error {
	return fmt.Errorf("cents: "+format+": %w", append(args, sentinel)...)
}

// quoted returns text quoted for an error message, cut to maxQuoted
// characters and followed by "..." if it is longer.
func quoted(text string) string {
	n := 0
	for i := range text {
		if n == maxQuoted {
			return fmt.Sprintf("%q...", text[:i])
		}
		n++
	}
	return fmt.Sprintf("%q", text)
}

// textError returns an ErrInvalidText error for text with the given reason.
func textError(text, reason string) error {
	return errorf(ErrInvalidText, "%s: %s", quoted(text), reason)
}
