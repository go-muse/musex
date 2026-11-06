package pcset

import (
	"math/bits"

	"github.com/go-muse/musex/pitchclass"
)

// Transpose returns the transposed set T_n(s) = {(x + n) mod N : x in s} for any
// integer n. If s is Set{}, Transpose returns Set{}.
func (s Set) Transpose(n int) Set {
	if !s.IsValid() {
		return Set{}
	}

	shift := modPositive(n, s.n)
	if shift == 0 || s.IsEmpty() || s.IsUniverse() {
		return s
	}

	if s.n <= inlineModulus {
		return Set{
			n:  s.n,
			lo: rotateLeft64(s.lo, shift, s.n),
		}
	}

	buf := make([]byte, (s.n+7)>>3)
	s.forEachNumber(func(x int) bool {
		y := x + shift
		if y >= s.n {
			y -= s.n
		}

		buf[y>>3] |= 1 << (y & 7)

		return true
	})

	return makeLargeSetFromBytes(s.n, buf)
}

// Invert returns the inverted set I_index(s) = {(index - x) mod N : x in s} for
// any integer index. If s is Set{}, Invert returns Set{}.
func (s Set) Invert(index int) Set {
	if !s.IsValid() {
		return Set{}
	}

	idx := modPositive(index, s.n)
	if s.IsEmpty() || s.IsUniverse() {
		return s
	}

	if s.n <= inlineModulus {
		var mask uint64

		for w := s.lo; w != 0; w &= w - 1 {
			x := bits.TrailingZeros64(w)
			y := idx - x
			if y < 0 {
				y += s.n
			}

			mask |= uint64(1) << y
		}

		return Set{n: s.n, lo: mask}
	}

	buf := make([]byte, (s.n+7)>>3)
	s.forEachNumber(func(x int) bool {
		y := idx - x
		if y < 0 {
			y += s.n
		}

		buf[y>>3] |= 1 << (y & 7)

		return true
	})

	return makeLargeSetFromBytes(s.n, buf)
}

// InvertAround returns the set inverted around axis, I_{2*axis}(s).
//
// InvertAround checks errors in the following order:
//  1. !s.IsValid() || !axis.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != axis.Modulus(): [ErrModulusMismatch]
func (s Set) InvertAround(axis pitchclass.Class) (Set, error) {
	if !s.IsValid() || !axis.IsValid() {
		return Set{}, errInvalidValue()
	}

	if s.n != axis.Modulus() {
		return Set{}, errModulusMismatch(s.n, axis.Modulus())
	}

	return s.Invert(2 * axis.Number()), nil
}

// MustInvertAround is like [Set.InvertAround] but panics on error.
func (s Set) MustInvertAround(axis pitchclass.Class) Set {
	res, err := s.InvertAround(axis)
	if err != nil {
		panic(err)
	}

	return res
}

// InvertSwapping returns the set inverted by the inversion I_{x+y} that swaps
// pitch classes x and y.
//
// InvertSwapping checks errors in the following order:
//  1. !s.IsValid() || !x.IsValid() || !y.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != x.Modulus() || s.Modulus() != y.Modulus(): [ErrModulusMismatch]
func (s Set) InvertSwapping(x, y pitchclass.Class) (Set, error) {
	if !s.IsValid() || !x.IsValid() || !y.IsValid() {
		return Set{}, errInvalidValue()
	}

	if s.n != x.Modulus() {
		return Set{}, errModulusMismatch(s.n, x.Modulus())
	}

	if s.n != y.Modulus() {
		return Set{}, errModulusMismatch(s.n, y.Modulus())
	}

	return s.Invert(x.Number() + y.Number()), nil
}

// MustInvertSwapping is like [Set.InvertSwapping] but panics on error.
func (s Set) MustInvertSwapping(x, y pitchclass.Class) Set {
	res, err := s.InvertSwapping(x, y)
	if err != nil {
		panic(err)
	}

	return res
}

// Multiply returns the affine-multiplied set M_m(s) = {(m * x) mod N : x in s}
// for any integer m. When gcd(m mod N, N) > 1, the cardinality of the result
// may be smaller than s.Cardinality(). If s is Set{}, Multiply returns Set{}.
func (s Set) Multiply(m int) Set {
	if !s.IsValid() {
		return Set{}
	}

	mult := modPositive(m, s.n)
	if s.IsEmpty() || mult == 1 {
		return s
	}

	if s.n <= inlineModulus {
		var mask uint64

		for w := s.lo; w != 0; w &= w - 1 {
			x := bits.TrailingZeros64(w)
			y := (mult * x) % s.n
			mask |= uint64(1) << y
		}

		return Set{n: s.n, lo: mask}
	}

	buf := make([]byte, (s.n+7)>>3)
	s.forEachNumber(func(x int) bool {
		y := (mult * x) % s.n
		buf[y>>3] |= 1 << (y & 7)

		return true
	})

	return makeLargeSetFromBytes(s.n, buf)
}

// Apply applies the affine operator op = T_n M_m to each element of s,
// returning {op(x) : x in s}.
//
// Apply checks errors in the following order:
//  1. !s.IsValid() || !op.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != op.Modulus(): [ErrModulusMismatch]
func (s Set) Apply(op pitchclass.Operator) (Set, error) {
	if !s.IsValid() || !op.IsValid() {
		return Set{}, errInvalidValue()
	}

	if s.n != op.Modulus() {
		return Set{}, errModulusMismatch(s.n, op.Modulus())
	}

	return s.Multiply(op.Multiplier()).Transpose(op.Index()), nil
}

// MustApply is like [Set.Apply] but panics on error.
func (s Set) MustApply(op pitchclass.Operator) Set {
	res, err := s.Apply(op)
	if err != nil {
		panic(err)
	}

	return res
}

// OperatorFixedClasses returns the set of pitch classes in Z_N that are mapped
// to themselves by op: {x in Z_N : op(x) == x}.
//
// OperatorFixedClasses checks errors in the following order:
//  1. !op.IsValid(): [ErrInvalidValue]
//  2. op.Modulus() > [MaxModulus]: [ErrModulusTooLarge]
func OperatorFixedClasses(op pitchclass.Operator) (Set, error) {
	if !op.IsValid() {
		return Set{}, errInvalidValue()
	}

	n := op.Modulus()
	if n > MaxModulus {
		return Set{}, errModulusTooLarge(n)
	}

	return affineFixedClasses(op.Multiplier(), op.Index(), n), nil
}

// MustOperatorFixedClasses is like [OperatorFixedClasses] but panics on error.
func MustOperatorFixedClasses(op pitchclass.Operator) Set {
	res, err := OperatorFixedClasses(op)
	if err != nil {
		panic(err)
	}

	return res
}

// MultiplicationFixedClasses returns the set of pitch classes in Z_modulus that
// are mapped to themselves by multiplication M_m: {x in Z_modulus : (m*x) mod modulus == x}.
//
// It returns [ErrInvalidModulus] if modulus < 1 and [ErrModulusTooLarge] if
// modulus > [MaxModulus].
func MultiplicationFixedClasses(m, modulus int) (Set, error) {
	if err := checkModulus(modulus); err != nil {
		return Set{}, err
	}

	return affineFixedClasses(modPositive(m, modulus), 0, modulus), nil
}

// MustMultiplicationFixedClasses is like [MultiplicationFixedClasses] but panics on error.
func MustMultiplicationFixedClasses(m, modulus int) Set {
	res, err := MultiplicationFixedClasses(m, modulus)
	if err != nil {
		panic(err)
	}

	return res
}

func rotateLeft64(lo uint64, shift, n int) uint64 {
	if shift == 0 {
		return lo
	}

	return ((lo << shift) | (lo >> (n - shift))) & universeMask64(n)
}

func affineFixedClasses(mult, trans, n int) Set {
	if n <= inlineModulus {
		var mask uint64

		for x := range n {
			if (mult*x+trans)%n == x {
				mask |= uint64(1) << x
			}
		}

		return Set{n: n, lo: mask}
	}

	buf := make([]byte, (n+7)>>3)
	for x := range n {
		if (mult*x+trans)%n == x {
			buf[x>>3] |= 1 << (x & 7)
		}
	}

	return makeLargeSetFromBytes(n, buf)
}
