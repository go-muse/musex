package pitchclass

import (
	"fmt"
	"iter"
	"math"
	"math/big"
	"math/bits"
)

// Transpositions returns an iterator over the N transpositions of modulus
// positions in the order T0, T1, ..., T(N-1). The transpositions form a
// group: it contains the identity and is closed under [Operator.Compose]
// and [Operator.Inverse].
//
// The iterator is lazy, restartable, and may be stopped with break, as
// described for [All]; for N = math.MaxInt the first operators are reached
// at once.
//
// If modulus is less than 1, Transpositions returns an error that wraps
// [ErrInvalidModulus] and an iterator that yields nothing; the iterator is
// never nil.
func Transpositions(modulus int) (iter.Seq[Operator], error) {
	if modulus < 1 {
		return emptySeq[Operator](), errModulus(modulus)
	}

	return func(yield func(Operator) bool) {
		yieldIndices(1%modulus, modulus, yield)
	}, nil
}

// MustTranspositions is like [Transpositions] but panics with its error
// instead of returning it. It is intended for values known to be valid.
func MustTranspositions(modulus int) iter.Seq[Operator] {
	seq, err := Transpositions(modulus)
	if err != nil {
		panic(err)
	}

	return seq
}

// TranspositionsInversions returns an iterator over the transpositions and
// inversions of modulus positions in the order T0, ..., T(N-1), then I0,
// ..., I(N-1): 2N operators, the T/I group, which is 24 operators for
// N = 12. For N <= 2 every inversion is a transposition, so only the N
// transpositions are yielded, without repetition.
//
// The iterator is lazy, restartable, and may be stopped with break, as
// described for [All].
//
// If modulus is less than 1, TranspositionsInversions returns an error that
// wraps [ErrInvalidModulus] and an iterator that yields nothing; the
// iterator is never nil.
func TranspositionsInversions(modulus int) (iter.Seq[Operator], error) {
	if modulus < 1 {
		return emptySeq[Operator](), errModulus(modulus)
	}

	return func(yield func(Operator) bool) {
		if !yieldIndices(1%modulus, modulus, yield) || modulus <= 2 {
			return
		}

		yieldIndices(modulus-1, modulus, yield)
	}, nil
}

// MustTranspositionsInversions is like [TranspositionsInversions] but panics
// with its error instead of returning it. It is intended for values known
// to be valid.
func MustTranspositionsInversions(modulus int) iter.Seq[Operator] {
	seq, err := TranspositionsInversions(modulus)
	if err != nil {
		panic(err)
	}

	return seq
}

// Operators returns an iterator over all N·φ(N) operators x ↦ m·x + n of
// modulus positions: for each invertible multiplier m in ascending order,
// as yielded by [InvertibleMultipliers], the operators with index n = 0,
// ..., N-1. This is the order of [Operator.Compare]. For N = 12, they are
// the 48 twelve-tone operators of Morris.
//
// The iterator is lazy, restartable, and may be stopped with break, as
// described for [All]; reaching the next multiplier takes time proportional
// to the gap to it times log N.
//
// If modulus is less than 1, Operators returns an error that wraps
// [ErrInvalidModulus] and an iterator that yields nothing; the iterator is
// never nil.
func Operators(modulus int) (iter.Seq[Operator], error) {
	multipliers, err := InvertibleMultipliers(modulus)
	if err != nil {
		return emptySeq[Operator](), err
	}

	return func(yield func(Operator) bool) {
		for m := range multipliers {
			if !yieldIndices(m, modulus, yield) {
				return
			}
		}
	}, nil
}

// MustOperators is like [Operators] but panics with its error instead of
// returning it. It is intended for values known to be valid.
func MustOperators(modulus int) iter.Seq[Operator] {
	seq, err := Operators(modulus)
	if err != nil {
		panic(err)
	}

	return seq
}

// yieldIndices yields the operators with the multiplier m and the indices 0
// to n-1 in ascending order. It reports whether yield accepted all of them.
func yieldIndices(m, n int, yield func(Operator) bool) bool {
	for t := range n {
		if !yield(Operator{n: n, m: m, t: t}) {
			return false
		}
	}

	return true
}

// CountTranspositions returns the number of transpositions of modulus
// positions, N.
//
// If modulus is less than 1, CountTranspositions returns 0 and an error
// that wraps [ErrInvalidModulus].
func CountTranspositions(modulus int) (int, error) {
	if modulus < 1 {
		return 0, errModulus(modulus)
	}

	return modulus, nil
}

// MustCountTranspositions is like [CountTranspositions] but panics with its
// error instead of returning it. It is intended for values known to be
// valid.
func MustCountTranspositions(modulus int) int {
	r, err := CountTranspositions(modulus)
	if err != nil {
		panic(err)
	}

	return r
}

// CountTranspositionsInversions returns the number of operators yielded by
// [TranspositionsInversions]: 2N, or N for N <= 2. For N = 12 it is 24.
//
// If modulus is less than 1, CountTranspositionsInversions returns 0 and an
// error that wraps [ErrInvalidModulus]. If 2N does not fit in an int, it
// returns 0 and an error that wraps [ErrOverflow]; use
// [CountTranspositionsInversionsBig] for the exact value.
func CountTranspositionsInversions(modulus int) (int, error) {
	if modulus < 1 {
		return 0, errModulus(modulus)
	}

	if modulus <= 2 {
		return modulus, nil
	}

	if modulus > math.MaxInt/2 {
		return 0, fmt.Errorf("%w: 2 * %d", ErrOverflow, modulus)
	}

	return 2 * modulus, nil
}

// MustCountTranspositionsInversions is like [CountTranspositionsInversions]
// but panics with its error instead of returning it. It is intended for
// values known to be valid.
func MustCountTranspositionsInversions(modulus int) int {
	r, err := CountTranspositionsInversions(modulus)
	if err != nil {
		panic(err)
	}

	return r
}

// CountOperators returns the number of operators yielded by [Operators]:
// N·φ(N), where φ is [Totient]. For N = 12 it is 48. It takes the time of
// [Totient].
//
// If modulus is less than 1, CountOperators returns 0 and an error that
// wraps [ErrInvalidModulus]. If N·φ(N) does not fit in an int, it returns 0
// and an error that wraps [ErrOverflow]; use [CountOperatorsBig] for the
// exact value.
func CountOperators(modulus int) (int, error) {
	if modulus < 1 {
		return 0, errModulus(modulus)
	}

	hi, lo := bits.Mul64(uint64(modulus), totient(uint64(modulus)))
	if hi != 0 || lo > math.MaxInt {
		return 0, fmt.Errorf("%w: %d * φ(%d)", ErrOverflow, modulus, modulus)
	}

	return int(lo), nil
}

// MustCountOperators is like [CountOperators] but panics with its error
// instead of returning it. It is intended for values known to be valid.
func MustCountOperators(modulus int) int {
	r, err := CountOperators(modulus)
	if err != nil {
		panic(err)
	}

	return r
}

// CountTranspositionsInversionsBig returns the exact number of operators
// yielded by [TranspositionsInversions] for any modulus: 2N, or N for
// N <= 2. Each call returns a new [big.Int] that the caller owns and may
// modify.
//
// If modulus is less than 1, CountTranspositionsInversionsBig returns nil
// and an error that wraps [ErrInvalidModulus].
func CountTranspositionsInversionsBig(modulus int) (*big.Int, error) {
	if modulus < 1 {
		return nil, errModulus(modulus)
	}

	r := big.NewInt(int64(modulus))
	if modulus > 2 {
		r.Lsh(r, 1)
	}

	return r, nil
}

// MustCountTranspositionsInversionsBig is like
// [CountTranspositionsInversionsBig] but panics with its error instead of
// returning it. It is intended for values known to be valid.
func MustCountTranspositionsInversionsBig(modulus int) *big.Int {
	r, err := CountTranspositionsInversionsBig(modulus)
	if err != nil {
		panic(err)
	}

	return r
}

// CountOperatorsBig returns the exact number of operators yielded by
// [Operators] for any modulus: N·φ(N). Each call returns a new [big.Int]
// that the caller owns and may modify. It takes the time of [Totient].
//
// If modulus is less than 1, CountOperatorsBig returns nil and an error that
// wraps [ErrInvalidModulus].
func CountOperatorsBig(modulus int) (*big.Int, error) {
	if modulus < 1 {
		return nil, errModulus(modulus)
	}

	phi := new(big.Int).SetUint64(totient(uint64(modulus)))

	return phi.Mul(phi, big.NewInt(int64(modulus))), nil
}

// MustCountOperatorsBig is like [CountOperatorsBig] but panics with its
// error instead of returning it. It is intended for values known to be
// valid.
func MustCountOperatorsBig(modulus int) *big.Int {
	r, err := CountOperatorsBig(modulus)
	if err != nil {
		panic(err)
	}

	return r
}
