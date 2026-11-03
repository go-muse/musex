package pitchclass

import "errors"

// Errors returned by the package. Each returned error matches exactly one of
// them with [errors.Is]; it may be the error itself or wrap it with context.
// The errors do not wrap each other, and their text is not part of the API.
//
// Other packages may wrap these errors, so that a single [errors.Is] check
// works for errors from related packages.
var (
	// ErrInvalidClass reports that a class cannot be created: the number of
	// positions is less than 1, or [New] is given a number outside the range
	// 0 to N-1. It is returned by [New] and [Reduce].
	ErrInvalidClass = errors.New("pitchclass: invalid class")

	// ErrInvalidModulus reports that a number of positions N is less than 1
	// in an operation that takes N as an argument, such as [All], [Totient],
	// [NewOperator], or [ParseNotation].
	ErrInvalidModulus = errors.New("pitchclass: invalid modulus")

	// ErrModulusMismatch reports that the classes or operators of a
	// two-place operation have different numbers of positions, such as
	// 1 mod 12 and 1 mod 19 in [Class.IntervalTo]. Classes of different N
	// are never combined.
	ErrModulusMismatch = errors.New("pitchclass: different moduli")

	// ErrNotInvertible reports a multiplier m that is not invertible modulo
	// N, that is, m and N have a common divisor greater than 1. It is
	// returned by [InverseMultiplier], [NewOperator], [Multiplication], and
	// [ParseOperator]. [Class.Multiply] accepts any multiplier and never
	// returns it.
	ErrNotInvertible = errors.New("pitchclass: multiplier is not invertible")

	// ErrInvalidValue reports that an operand is the zero [Class] or the zero
	// [Operator], which is not a class or an operator. It is returned by
	// two-place operations, [InversionAround], and [Class.Notation], and it
	// is checked before [ErrModulusMismatch].
	ErrInvalidValue = errors.New("pitchclass: invalid class or operator value")

	// ErrNotation reports a text that is not a valid notation, or a notation
	// style that cannot write or read classes of the given number of
	// positions. It is returned by [Parse], [ParseNotation],
	// [ParseOperator], and [Class.Notation].
	ErrNotation = errors.New("pitchclass: invalid notation")

	// ErrOverflow reports that the exact result of a count does not fit in
	// an int. It is returned by [CountTranspositionsInversions] and
	// [CountOperators]; [CountTranspositionsInversionsBig] and
	// [CountOperatorsBig] return the exact value. Results are never
	// truncated.
	ErrOverflow = errors.New("pitchclass: result does not fit in int")
)
