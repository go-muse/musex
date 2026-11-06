package pcset

import (
	"fmt"
	"iter"
	"math/big"
	"math/bits"
)

func emptySetSeq() iter.Seq[Set] {
	return func(func(Set) bool) {}
}

// All returns a reusable lazy iterator yielding all 2^modulus subsets of
// Z_modulus in ascending order of their binary characteristic code sum_{x in X} 2^x
// (∅, {0}, {1}, {0, 1}, {2}, {0, 2}, {1, 2}, {0, 1, 2}, ...).
//
// The returned iterator is never nil; on error it returns an empty iterator.
// All checks errors in the following order:
//  1. modulus < 1: [ErrInvalidModulus]
//  2. modulus > [MaxModulus]: [ErrModulusTooLarge]
func All(modulus int) (iter.Seq[Set], error) {
	if err := checkModulus(modulus); err != nil {
		return emptySetSeq(), err
	}

	return func(yield func(Set) bool) {
		if modulus <= inlineModulus {
			maxMask := universeMask64(modulus)
			for mask := uint64(0); ; mask++ {
				if !yield(Set{n: modulus, lo: mask}) {
					return
				}

				if mask == maxMask {
					return
				}
			}
		}

		buf := make([]byte, (modulus+7)>>3)
		for {
			if !yield(makeLargeSetFromBytes(modulus, buf)) {
				return
			}

			for bit := 0; ; bit++ {
				byteIdx := bit >> 3
				bitMask := byte(1 << (bit & 7))

				if (buf[byteIdx] & bitMask) == 0 {
					buf[byteIdx] |= bitMask
					break
				}

				buf[byteIdx] &^= bitMask
			}
		}
	}, nil
}

// MustAll is like [All] but panics on error.
func MustAll(modulus int) iter.Seq[Set] {
	seq, err := All(modulus)
	if err != nil {
		panic(err)
	}

	return seq
}

// OfCardinality returns a reusable lazy iterator yielding all C(modulus, k)
// subsets of Z_modulus of cardinality k in lexicographical order of their
// ascending element lists (matching [Set.Compare]).
//
// The returned iterator is never nil; on error it returns an empty iterator.
// OfCardinality checks errors in the following order:
//  1. modulus < 1: [ErrInvalidModulus]
//  2. modulus > [MaxModulus]: [ErrModulusTooLarge]
//  3. k < 0 || k > modulus: [ErrInvalidCardinality]
func OfCardinality(modulus, k int) (iter.Seq[Set], error) {
	if err := checkModulus(modulus); err != nil {
		return emptySetSeq(), err
	}

	if k < 0 || k > modulus {
		return emptySetSeq(), errInvalidCardinality(k, modulus)
	}

	return func(yield func(Set) bool) {
		comb := make([]int, k)
		for i := range k {
			comb[i] = i
		}

		for {
			if !yield(MustNew(modulus, comb...)) {
				return
			}

			i := k - 1
			for i >= 0 && comb[i] == modulus-k+i {
				i--
			}

			if i < 0 {
				return
			}

			comb[i]++
			for j := i + 1; j < k; j++ {
				comb[j] = comb[j-1] + 1
			}
		}
	}, nil
}

// MustOfCardinality is like [OfCardinality] but panics on error.
func MustOfCardinality(modulus, k int) iter.Seq[Set] {
	seq, err := OfCardinality(modulus, k)
	if err != nil {
		panic(err)
	}

	return seq
}

// Count returns the total number of subsets 2^modulus of Z_modulus as an int.
//
// Count checks errors in the following order:
//  1. modulus < 1: [ErrInvalidModulus]
//  2. modulus > [MaxModulus]: [ErrModulusTooLarge]
//  3. 2^modulus does not fit in int (modulus >= 63 on 64-bit platforms): [ErrOverflow]
func Count(modulus int) (int, error) {
	if err := checkModulus(modulus); err != nil {
		return 0, err
	}

	if modulus >= bits.UintSize-1 {
		return 0, fmt.Errorf("%w: 2^%d overflows int", ErrOverflow, modulus)
	}

	return 1 << modulus, nil
}

// MustCount is like [Count] but panics on error.
func MustCount(modulus int) int {
	c, err := Count(modulus)
	if err != nil {
		panic(err)
	}

	return c
}

// CountBig returns a newly allocated *big.Int equal to the total number of
// subsets 2^modulus of Z_modulus.
//
// CountBig checks errors in the following order:
//  1. modulus < 1: [ErrInvalidModulus]
//  2. modulus > [MaxModulus]: [ErrModulusTooLarge]
func CountBig(modulus int) (*big.Int, error) {
	if err := checkModulus(modulus); err != nil {
		return nil, err
	}

	return new(big.Int).Lsh(big.NewInt(1), uint(modulus)), nil
}

// MustCountBig is like [CountBig] but panics on error.
func MustCountBig(modulus int) *big.Int {
	c, err := CountBig(modulus)
	if err != nil {
		panic(err)
	}

	return c
}

// CountOfCardinality returns the number of k-element subsets C(modulus, k) of
// Z_modulus as an int.
//
// CountOfCardinality checks errors in the following order:
//  1. modulus < 1: [ErrInvalidModulus]
//  2. modulus > [MaxModulus]: [ErrModulusTooLarge]
//  3. k < 0 || k > modulus: [ErrInvalidCardinality]
//  4. C(modulus, k) does not fit in int: [ErrOverflow]
func CountOfCardinality(modulus, k int) (int, error) {
	b, err := CountOfCardinalityBig(modulus, k)
	if err != nil {
		return 0, err
	}

	if !b.IsInt64() {
		return 0, fmt.Errorf("%w: C(%d, %d) overflows int", ErrOverflow, modulus, k)
	}

	return int(b.Int64()), nil
}

// MustCountOfCardinality is like [CountOfCardinality] but panics on error.
func MustCountOfCardinality(modulus, k int) int {
	c, err := CountOfCardinality(modulus, k)
	if err != nil {
		panic(err)
	}

	return c
}

// CountOfCardinalityBig returns a newly allocated *big.Int equal to the
// binomial coefficient C(modulus, k).
//
// CountOfCardinalityBig checks errors in the following order:
//  1. modulus < 1: [ErrInvalidModulus]
//  2. modulus > [MaxModulus]: [ErrModulusTooLarge]
//  3. k < 0 || k > modulus: [ErrInvalidCardinality]
func CountOfCardinalityBig(modulus, k int) (*big.Int, error) {
	if err := checkModulus(modulus); err != nil {
		return nil, err
	}

	if k < 0 || k > modulus {
		return nil, errInvalidCardinality(k, modulus)
	}

	return new(big.Int).Binomial(int64(modulus), int64(k)), nil
}

// MustCountOfCardinalityBig is like [CountOfCardinalityBig] but panics on error.
func MustCountOfCardinalityBig(modulus, k int) *big.Int {
	c, err := CountOfCardinalityBig(modulus, k)
	if err != nil {
		panic(err)
	}

	return c
}
