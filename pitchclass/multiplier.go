package pitchclass

import (
	"fmt"
	"iter"
)

// Invertible reports whether the multiplier m is invertible modulo modulus:
// the greatest common divisor of m mod N and N is 1. Then [Class.Multiply]
// by m permutes the classes, and m can be the multiplier of an [Operator].
// Any int m is accepted. For N = 12 the invertible multipliers are 1, 5, 7,
// and 11; for N = 1 every multiplier is invertible.
//
// If modulus is less than 1, Invertible returns false and an error that
// wraps [ErrInvalidModulus].
func Invertible(m, modulus int) (bool, error) {
	if modulus < 1 {
		return false, errModulus(modulus)
	}

	return invertible(mod(m, modulus), modulus), nil
}

// MustInvertible is like [Invertible] but panics with its error instead of
// returning it. It is intended for values known to be valid.
func MustInvertible(m, modulus int) bool {
	ok, err := Invertible(m, modulus)
	if err != nil {
		panic(err)
	}

	return ok
}

// InverseMultiplier returns the inverse of the multiplier m modulo modulus:
// the r in the range 0 to N-1 with (m * r) mod N = 1 mod N. Any int m is
// accepted: modulo 12, the inverse of 5 is 5, of -1 is 11, and of 17 is 5;
// modulo 19, the inverse of 2 is 10. Modulo 1, the inverse is 0.
//
// If modulus is less than 1, InverseMultiplier returns 0 and an error that
// wraps [ErrInvalidModulus]. Otherwise, if m is not invertible modulo N, it
// returns 0 and an error that wraps [ErrNotInvertible].
func InverseMultiplier(m, modulus int) (int, error) {
	if modulus < 1 {
		return 0, errModulus(modulus)
	}

	r := mod(m, modulus)
	if !invertible(r, modulus) {
		return 0, errNotInvertible(m, modulus)
	}

	return inverseMod(r, modulus), nil
}

// MustInverseMultiplier is like [InverseMultiplier] but panics with its
// error instead of returning it. It is intended for values known to be
// valid.
func MustInverseMultiplier(m, modulus int) int {
	r, err := InverseMultiplier(m, modulus)
	if err != nil {
		panic(err)
	}

	return r
}

// InvertibleMultipliers returns an iterator over the invertible multipliers
// m in the range 0 to N-1 for N = modulus, in ascending order: 1, 5, 7, 11
// for N = 12, only 0 for N = 1, and only 1 for N = 2. Their number is
// [Totient] of N.
//
// The iterator is lazy, restartable, and may be stopped with break, as
// described for [All]. The time to reach the next multiplier is
// proportional to the gap to it times log N.
//
// If modulus is less than 1, InvertibleMultipliers returns an error that
// wraps [ErrInvalidModulus] and an iterator that yields nothing; the
// iterator is never nil.
func InvertibleMultipliers(modulus int) (iter.Seq[int], error) {
	if modulus < 1 {
		return emptySeq[int](), errModulus(modulus)
	}

	return func(yield func(int) bool) {
		for m := range modulus {
			if invertible(m, modulus) && !yield(m) {
				return
			}
		}
	}, nil
}

// MustInvertibleMultipliers is like [InvertibleMultipliers] but panics with
// its error instead of returning it. It is intended for values known to be
// valid.
func MustInvertibleMultipliers(modulus int) iter.Seq[int] {
	seq, err := InvertibleMultipliers(modulus)
	if err != nil {
		panic(err)
	}

	return seq
}

// Totient returns Euler's totient φ(N) of N = modulus: the number of
// invertible multipliers in the range 0 to N-1. φ(12) = 4, φ(19) = 18, and
// φ(1) = 1. The result always fits in an int, since φ(N) <= N.
//
// Totient factors N with Pollard's rho method in Brent's variant and tests
// primality with a deterministic Miller-Rabin test, without searching for
// divisors up to √N: in the worst case, a product of two primes near √N,
// its time grows about as the fourth root of N. For a prime N, it divides
// by small numbers and runs a single primality test. Totient is
// deterministic, does not allocate, and has no state.
//
// If modulus is less than 1, Totient returns 0 and an error that wraps
// [ErrInvalidModulus].
func Totient(modulus int) (int, error) {
	if modulus < 1 {
		return 0, errModulus(modulus)
	}

	return int(totient(uint64(modulus))), nil
}

// MustTotient is like [Totient] but panics with its error instead of
// returning it. It is intended for values known to be valid.
func MustTotient(modulus int) int {
	r, err := Totient(modulus)
	if err != nil {
		panic(err)
	}

	return r
}

// errNotInvertible returns the error for a multiplier m that is not
// invertible modulo n.
func errNotInvertible(m, n int) error {
	return fmt.Errorf("%w: multiplier %d modulo %d", ErrNotInvertible, m, n)
}
