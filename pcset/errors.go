package pcset

import (
	"errors"
	"fmt"
)

// MaxModulus is the maximum supported modulus N for pitch-class sets (65,536).
//
// Operations with N < 1 return [ErrInvalidModulus], and operations with
// N > MaxModulus return [ErrModulusTooLarge].
const MaxModulus = 65_536

// Sentinel errors returned by operations in package pcset. Each returned
// error matches exactly one of these errors with [errors.Is]. The errors do
// not wrap one another, and their message text is not part of the contract.
var (
	// ErrInvalidModulus reports that a modulus N is less than 1.
	ErrInvalidModulus = errors.New("pcset: invalid modulus")

	// ErrModulusTooLarge reports that a modulus N exceeds [MaxModulus]
	// (65,536), including when a valid [github.com/go-muse/musex/pitchclass.Class]
	// or [github.com/go-muse/musex/pitchclass.Operator] with Modulus() > MaxModulus
	// is passed to an operation or when [Parse] reads a canonical string
	// whose modulus exceeds MaxModulus.
	ErrModulusTooLarge = errors.New("pcset: modulus exceeds maximum")

	// ErrInvalidClass reports that a class number is outside the range
	// 0..N-1 in [New], a bitmask has bits outside 0..N-1 (or is negative)
	// in [FromBitmask] or [FromBitmaskBig], or [FromClasses] is called with
	// an empty class list.
	ErrInvalidClass = errors.New("pcset: invalid pitch class")

	// ErrInvalidCardinality reports that a requested subset cardinality k is
	// outside the range 0..N in [OfCardinality], [CountOfCardinality], or
	// [CountOfCardinalityBig].
	ErrInvalidCardinality = errors.New("pcset: invalid cardinality")

	// ErrInvalidIntervalClass reports that an interval class d is outside
	// the range 1..floor(N/2) in [IntervalVector.Count].
	ErrInvalidIntervalClass = errors.New("pcset: invalid interval class")

	// ErrModulusMismatch reports that operands of a two-place operation or
	// classes passed to [FromClasses] / [FromClassesWithModulus] have
	// different moduli N.
	ErrModulusMismatch = errors.New("pcset: different moduli")

	// ErrNotModulus12 reports that [Set.InvarianceVector] was called on a
	// valid set whose modulus N is not 12.
	ErrNotModulus12 = errors.New("pcset: operation requires modulus 12")

	// ErrInvalidValue reports that an operand is the zero [Set], zero
	// [IntervalVector], zero [InvarianceVector], zero
	// [github.com/go-muse/musex/pitchclass.Class], zero
	// [github.com/go-muse/musex/pitchclass.Operator], or a nil *big.Int.
	ErrInvalidValue = errors.New("pcset: invalid value")

	// ErrInvalidAlgorithm reports that an [Algorithm] value is neither
	// [Rahn] nor [Forte].
	ErrInvalidAlgorithm = errors.New("pcset: invalid normal form algorithm")

	// ErrNotation reports that a text, [github.com/go-muse/musex/pitchclass.Style],
	// or compact notation is invalid for the given modulus or value.
	ErrNotation = errors.New("pcset: invalid notation")

	// ErrOverflow reports that a bitmask does not fit in uint64 (when
	// N > 64 in [Set.Bitmask] or [FromBitmask]) or that a subset count does
	// not fit in int (in [Count] or [CountOfCardinality]).
	ErrOverflow = errors.New("pcset: result does not fit in integer type")
)

func checkModulus(modulus int) error {
	if modulus < 1 {
		return errInvalidModulus(modulus)
	}

	if modulus > MaxModulus {
		return errModulusTooLarge(modulus)
	}

	return nil
}

func checkSetPair(a, b Set) error {
	if !a.IsValid() || !b.IsValid() {
		return errInvalidValue()
	}

	if a.n != b.n {
		return errModulusMismatch(a.n, b.n)
	}

	return nil
}

func errInvalidModulus(modulus int) error {
	return fmt.Errorf("%w: modulus %d is less than 1", ErrInvalidModulus, modulus)
}

func errModulusTooLarge(modulus int) error {
	return fmt.Errorf("%w: modulus %d exceeds %d", ErrModulusTooLarge, modulus, MaxModulus)
}

func errModulusMismatch(a, b int) error {
	return fmt.Errorf("%w: %d and %d", ErrModulusMismatch, a, b)
}

func errInvalidValue() error {
	return fmt.Errorf("%w: zero or nil operand", ErrInvalidValue)
}

func errInvalidCardinality(k, modulus int) error {
	return fmt.Errorf("%w: cardinality %d for modulus %d", ErrInvalidCardinality, k, modulus)
}

func errInvalidAlgorithm(alg Algorithm) error {
	return fmt.Errorf("%w: %v", ErrInvalidAlgorithm, alg)
}

func errNotation(s string) error {
	return fmt.Errorf("%w: %q", ErrNotation, s)
}
