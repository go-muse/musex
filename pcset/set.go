package pcset

import (
	"cmp"
	"fmt"
	"iter"
	"math/big"
	"math/bits"
	"slices"

	"github.com/go-muse/musex/pitchclass"
)

const inlineModulus = 64

// Set is an immutable mathematical subset of Z_N for 1 <= N <= [MaxModulus].
//
// Set is a value type that supports the built-in == operator and can be used
// as a map key. Its zero value Set{} is invalid (IsValid() == false) and is
// distinct from an empty set [Empty](N) of any modulus N.
type Set struct {
	n  int
	lo uint64
	hi string
}

// Empty returns the empty pitch-class set ∅ in Z_modulus.
//
// It returns [ErrInvalidModulus] if modulus < 1 and [ErrModulusTooLarge] if
// modulus > [MaxModulus].
func Empty(modulus int) (Set, error) {
	if err := checkModulus(modulus); err != nil {
		return Set{}, err
	}

	return Set{n: modulus}, nil
}

// MustEmpty is like [Empty] but panics if modulus is out of range.
func MustEmpty(modulus int) Set {
	s, err := Empty(modulus)
	if err != nil {
		panic(err)
	}

	return s
}

// Universe returns the universal pitch-class set Z_modulus = {0, 1, ..., modulus-1}.
//
// It returns [ErrInvalidModulus] if modulus < 1 and [ErrModulusTooLarge] if
// modulus > [MaxModulus].
func Universe(modulus int) (Set, error) {
	if err := checkModulus(modulus); err != nil {
		return Set{}, err
	}

	return universeSet(modulus), nil
}

// MustUniverse is like [Universe] but panics if modulus is out of range.
func MustUniverse(modulus int) Set {
	s, err := Universe(modulus)
	if err != nil {
		panic(err)
	}

	return s
}

// New creates a pitch-class set in Z_modulus from strict class numbers in
// 0..modulus-1. Duplicate numbers are ignored and element order does not
// matter. If no numbers are provided, New returns [Empty](modulus).
//
// New checks errors in the following order:
//  1. modulus < 1: [ErrInvalidModulus]
//  2. modulus > [MaxModulus]: [ErrModulusTooLarge]
//  3. any number < 0 or >= modulus: [ErrInvalidClass] (no modular wraparound)
func New(modulus int, numbers ...int) (Set, error) {
	if err := checkModulus(modulus); err != nil {
		return Set{}, err
	}

	if modulus <= inlineModulus {
		var mask uint64

		for _, k := range numbers {
			if k < 0 || k >= modulus {
				return Set{}, fmt.Errorf("%w: class %d out of range [0, %d)", ErrInvalidClass, k, modulus)
			}

			mask |= uint64(1) << k
		}

		return Set{n: modulus, lo: mask}, nil
	}

	maxK := -1

	for _, k := range numbers {
		if k < 0 || k >= modulus {
			return Set{}, fmt.Errorf("%w: class %d out of range [0, %d)", ErrInvalidClass, k, modulus)
		}

		if k > maxK {
			maxK = k
		}
	}

	if maxK < 0 {
		return Set{n: modulus}, nil
	}

	buf := make([]byte, (maxK>>3)+1)
	for _, k := range numbers {
		buf[k>>3] |= 1 << (k & 7)
	}

	return Set{n: modulus, hi: string(buf)}, nil
}

// MustNew is like [New] but panics if modulus or any class number is invalid.
func MustNew(modulus int, numbers ...int) Set {
	s, err := New(modulus, numbers...)
	if err != nil {
		panic(err)
	}

	return s
}

// Reduce creates a pitch-class set in Z_modulus by reducing arbitrary integers
// xs modulo modulus into 0..modulus-1. Duplicate residues are merged. If no
// integers are provided, Reduce returns [Empty](modulus).
//
// It returns [ErrInvalidModulus] if modulus < 1 and [ErrModulusTooLarge] if
// modulus > [MaxModulus].
func Reduce(modulus int, xs ...int) (Set, error) {
	if err := checkModulus(modulus); err != nil {
		return Set{}, err
	}

	if modulus <= inlineModulus {
		var mask uint64

		for _, x := range xs {
			k := modPositive(x, modulus)
			mask |= uint64(1) << k
		}

		return Set{n: modulus, lo: mask}, nil
	}

	buf := make([]byte, (modulus+7)>>3)
	for _, x := range xs {
		k := modPositive(x, modulus)
		buf[k>>3] |= 1 << (k & 7)
	}

	return makeLargeSetFromBytes(modulus, buf), nil
}

// MustReduce is like [Reduce] but panics if modulus is out of range.
func MustReduce(modulus int, xs ...int) Set {
	s, err := Reduce(modulus, xs...)
	if err != nil {
		panic(err)
	}

	return s
}

// FromClasses creates a pitch-class set from one or more [pitchclass.Class]
// values, inferring the modulus N from classes[0].Modulus().
//
// FromClasses checks errors in the following order:
//  1. len(classes) == 0: [ErrInvalidClass] (use [FromClassesWithModulus] for potentially empty slices)
//  2. any !c.IsValid(): [ErrInvalidValue]
//  3. any c.Modulus() > [MaxModulus]: [ErrModulusTooLarge]
//  4. any c.Modulus() != classes[0].Modulus(): [ErrModulusMismatch]
func FromClasses(classes ...pitchclass.Class) (Set, error) {
	if len(classes) == 0 {
		return Set{}, fmt.Errorf("%w: empty class list without explicit modulus", ErrInvalidClass)
	}

	for _, c := range classes {
		if !c.IsValid() {
			return Set{}, errInvalidValue()
		}
	}

	for _, c := range classes {
		if c.Modulus() > MaxModulus {
			return Set{}, errModulusTooLarge(c.Modulus())
		}
	}

	modulus := classes[0].Modulus()
	for _, c := range classes[1:] {
		if c.Modulus() != modulus {
			return Set{}, errModulusMismatch(modulus, c.Modulus())
		}
	}

	return buildSetFromValidatedClasses(modulus, classes), nil
}

// MustFromClasses is like [FromClasses] but panics on error.
func MustFromClasses(classes ...pitchclass.Class) Set {
	s, err := FromClasses(classes...)
	if err != nil {
		panic(err)
	}

	return s
}

// FromClassesWithModulus creates a pitch-class set in Z_modulus from zero or
// more [pitchclass.Class] values. When classes is empty, it returns
// [Empty](modulus).
//
// FromClassesWithModulus checks errors in the following order:
//  1. modulus < 1: [ErrInvalidModulus]
//  2. modulus > [MaxModulus]: [ErrModulusTooLarge]
//  3. any !c.IsValid(): [ErrInvalidValue]
//  4. any c.Modulus() > [MaxModulus]: [ErrModulusTooLarge]
//  5. any c.Modulus() != modulus: [ErrModulusMismatch]
func FromClassesWithModulus(modulus int, classes ...pitchclass.Class) (Set, error) {
	if err := checkModulus(modulus); err != nil {
		return Set{}, err
	}

	for _, c := range classes {
		if !c.IsValid() {
			return Set{}, errInvalidValue()
		}
	}

	for _, c := range classes {
		if c.Modulus() > MaxModulus {
			return Set{}, errModulusTooLarge(c.Modulus())
		}
	}

	for _, c := range classes {
		if c.Modulus() != modulus {
			return Set{}, errModulusMismatch(modulus, c.Modulus())
		}
	}

	return buildSetFromValidatedClasses(modulus, classes), nil
}

// MustFromClassesWithModulus is like [FromClassesWithModulus] but panics on error.
func MustFromClassesWithModulus(modulus int, classes ...pitchclass.Class) Set {
	s, err := FromClassesWithModulus(modulus, classes...)
	if err != nil {
		panic(err)
	}

	return s
}

// FromBitmask creates a pitch-class set in Z_modulus from a 64-bit
// characteristic code sum_{x in X} 2^x for 1 <= modulus <= 64.
//
// FromBitmask checks errors in the following order:
//  1. modulus < 1: [ErrInvalidModulus]
//  2. modulus > [MaxModulus]: [ErrModulusTooLarge]
//  3. modulus > 64: [ErrOverflow]
//  4. mask has any bit set at position >= modulus: [ErrInvalidClass]
func FromBitmask(mask uint64, modulus int) (Set, error) {
	if err := checkModulus(modulus); err != nil {
		return Set{}, err
	}

	if modulus > inlineModulus {
		return Set{}, fmt.Errorf("%w: modulus %d exceeds 64-bit bitmask capacity", ErrOverflow, modulus)
	}

	if modulus < inlineModulus && (mask>>modulus) != 0 {
		return Set{}, fmt.Errorf("%w: bitmask 0x%x has bits >= modulus %d", ErrInvalidClass, mask, modulus)
	}

	return Set{n: modulus, lo: mask}, nil
}

// MustFromBitmask is like [FromBitmask] but panics on error.
func MustFromBitmask(mask uint64, modulus int) Set {
	s, err := FromBitmask(mask, modulus)
	if err != nil {
		panic(err)
	}

	return s
}

// FromBitmaskBig creates a pitch-class set in Z_modulus from an arbitrary-precision
// characteristic code sum_{x in X} 2^x for 1 <= modulus <= [MaxModulus]. The input
// mask is neither modified nor retained.
//
// FromBitmaskBig checks errors in the following order:
//  1. modulus < 1: [ErrInvalidModulus]
//  2. modulus > [MaxModulus]: [ErrModulusTooLarge]
//  3. mask == nil: [ErrInvalidValue]
//  4. mask < 0 or mask >= 2^modulus: [ErrInvalidClass]
func FromBitmaskBig(mask *big.Int, modulus int) (Set, error) {
	if err := checkModulus(modulus); err != nil {
		return Set{}, err
	}

	if mask == nil {
		return Set{}, errInvalidValue()
	}

	if mask.Sign() < 0 || mask.BitLen() > modulus {
		return Set{}, fmt.Errorf("%w: bitmask out of range [0, 2^%d)", ErrInvalidClass, modulus)
	}

	if modulus <= inlineModulus {
		return Set{n: modulus, lo: mask.Uint64()}, nil
	}

	be := mask.Bytes()
	le := make([]byte, len(be))
	for i, b := range be {
		le[len(be)-1-i] = b
	}

	return Set{n: modulus, hi: string(le)}, nil
}

// MustFromBitmaskBig is like [FromBitmaskBig] but panics on error.
func MustFromBitmaskBig(mask *big.Int, modulus int) Set {
	s, err := FromBitmaskBig(mask, modulus)
	if err != nil {
		panic(err)
	}

	return s
}

// IsValid reports whether s is a valid pitch-class set (not the zero value Set{}).
func (s Set) IsValid() bool {
	return s.n >= 1
}

// Modulus returns the modulus N of s, or 0 if s is the zero value Set{}.
func (s Set) Modulus() int {
	return s.n
}

// Cardinality returns the number of elements |s| in 0..N, or 0 if s is Set{}.
func (s Set) Cardinality() int {
	if !s.IsValid() {
		return 0
	}

	if s.n <= inlineModulus {
		return bits.OnesCount64(s.lo)
	}

	count := 0
	for i := range len(s.hi) {
		count += bits.OnesCount8(s.hi[i])
	}

	return count
}

// IsEmpty reports whether s is a valid empty set ∅ (IsValid() && Cardinality() == 0).
// For the zero value Set{}, IsEmpty returns false.
func (s Set) IsEmpty() bool {
	return s.n >= 1 && s.lo == 0 && s.hi == ""
}

// IsUniverse reports whether s is the valid universal set Z_N (IsValid() && Cardinality() == N).
// For the zero value Set{}, IsUniverse returns false.
func (s Set) IsUniverse() bool {
	if !s.IsValid() {
		return false
	}

	if s.n <= inlineModulus {
		return s.lo == universeMask64(s.n)
	}

	return s.Cardinality() == s.n
}

// Contains reports whether c is a valid pitch class with the same modulus as s
// and belongs to s.
func (s Set) Contains(c pitchclass.Class) bool {
	return s.IsValid() && c.IsValid() && s.n == c.Modulus() && s.hasBit(c.Number())
}

// ContainsNumber reports whether number is in 0..s.Modulus()-1 and belongs to s.
// No modular reduction is performed.
func (s Set) ContainsNumber(number int) bool {
	return s.IsValid() && number >= 0 && number < s.n && s.hasBit(number)
}

// Classes returns a newly allocated slice of the pitch classes in s in strictly
// ascending order of class number. For an empty valid set it returns a non-nil
// empty slice; for Set{} it returns nil.
func (s Set) Classes() []pitchclass.Class {
	if !s.IsValid() {
		return nil
	}

	res := make([]pitchclass.Class, 0, s.Cardinality())
	s.forEachNumber(func(x int) bool {
		res = append(res, pitchclass.MustNew(x, s.n))
		return true
	})

	return res
}

// Numbers returns a newly allocated slice of the class numbers in s in strictly
// ascending order. For an empty valid set it returns a non-nil empty slice; for
// Set{} it returns nil.
func (s Set) Numbers() []int {
	if !s.IsValid() {
		return nil
	}

	res := make([]int, 0, s.Cardinality())
	s.forEachNumber(func(x int) bool {
		res = append(res, x)
		return true
	})

	return res
}

// All returns a reusable lazy iterator over the pitch classes of s in strictly
// ascending order of class number. The returned iterator is never nil; for
// [Empty](N) and Set{} it yields no elements.
func (s Set) All() iter.Seq[pitchclass.Class] {
	return func(yield func(pitchclass.Class) bool) {
		if !s.IsValid() {
			return
		}

		s.forEachNumber(func(x int) bool {
			return yield(pitchclass.MustNew(x, s.n))
		})
	}
}

// Bitmask returns the 64-bit characteristic code sum_{x in s} 2^x for a valid
// set with Modulus() <= 64.
//
// It returns [ErrInvalidValue] if !s.IsValid() and [ErrOverflow] if
// s.Modulus() > 64.
func (s Set) Bitmask() (uint64, error) {
	if !s.IsValid() {
		return 0, errInvalidValue()
	}

	if s.n > inlineModulus {
		return 0, fmt.Errorf("%w: modulus %d exceeds 64-bit bitmask capacity", ErrOverflow, s.n)
	}

	return s.lo, nil
}

// MustBitmask is like [Set.Bitmask] but panics on error.
func (s Set) MustBitmask() uint64 {
	m, err := s.Bitmask()
	if err != nil {
		panic(err)
	}

	return m
}

// BitmaskBig returns a newly allocated *big.Int containing the characteristic
// code sum_{x in s} 2^x.
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) BitmaskBig() (*big.Int, error) {
	if !s.IsValid() {
		return nil, errInvalidValue()
	}

	if s.n <= inlineModulus {
		return new(big.Int).SetUint64(s.lo), nil
	}

	be := make([]byte, len(s.hi))
	for i := range len(s.hi) {
		be[len(s.hi)-1-i] = s.hi[i]
	}

	return new(big.Int).SetBytes(be), nil
}

// MustBitmaskBig is like [Set.BitmaskBig] but panics on error.
func (s Set) MustBitmaskBig() *big.Int {
	b, err := s.BitmaskBig()
	if err != nil {
		panic(err)
	}

	return b
}

// Equal reports whether s and other are both valid pitch-class sets with the
// same modulus and the same elements. If either operand is Set{}, Equal returns
// false.
func (s Set) Equal(other Set) bool {
	return s.IsValid() && other.IsValid() && s == other
}

// Compare compares s and other and returns -1, 0, or +1:
//  1. by Modulus() ascending (so Set{} with Modulus() == 0 is less than any valid set);
//  2. by Cardinality() ascending;
//  3. lexicographically by their ascending element lists (c_0, c_1, ..., c_{k-1}).
func (s Set) Compare(other Set) int {
	if c := cmp.Compare(s.n, other.n); c != 0 {
		return c
	}

	if c := cmp.Compare(s.Cardinality(), other.Cardinality()); c != 0 {
		return c
	}

	return slices.Compare(s.Numbers(), other.Numbers())
}

func (s Set) hasBit(x int) bool {
	if s.n <= inlineModulus {
		return (s.lo & (uint64(1) << x)) != 0
	}

	idx := x >> 3

	return idx < len(s.hi) && (s.hi[idx]&(1<<(x&7))) != 0
}

func (s Set) byteAt(idx int) byte {
	if idx < len(s.hi) {
		return s.hi[idx]
	}

	return 0
}

func (s Set) forEachNumber(fn func(x int) bool) {
	if s.n <= inlineModulus {
		for w := s.lo; w != 0; w &= w - 1 {
			if !fn(bits.TrailingZeros64(w)) {
				return
			}
		}

		return
	}

	for i := range len(s.hi) {
		for b := s.hi[i]; b != 0; b &= b - 1 {
			if !fn((i << 3) + bits.TrailingZeros8(b)) {
				return
			}
		}
	}
}

func universeMask64(n int) uint64 {
	if n == inlineModulus {
		return ^uint64(0)
	}

	return (uint64(1) << n) - 1
}

func universeSet(n int) Set {
	if n <= inlineModulus {
		return Set{n: n, lo: universeMask64(n)}
	}

	byteLen := (n + 7) >> 3
	buf := make([]byte, byteLen)

	for i := range buf {
		buf[i] = 0xFF
	}

	if rem := n & 7; rem != 0 {
		buf[byteLen-1] = (1 << rem) - 1
	}

	return Set{n: n, hi: string(buf)}
}

func makeLargeSetFromBytes(n int, buf []byte) Set {
	end := len(buf)
	for end > 0 && buf[end-1] == 0 {
		end--
	}

	if end == 0 {
		return Set{n: n}
	}

	return Set{n: n, hi: string(buf[:end])}
}

func buildSetFromValidatedClasses(modulus int, classes []pitchclass.Class) Set {
	if modulus <= inlineModulus {
		var mask uint64

		for _, c := range classes {
			mask |= uint64(1) << c.Number()
		}

		return Set{n: modulus, lo: mask}
	}

	buf := make([]byte, (modulus+7)>>3)
	for _, c := range classes {
		k := c.Number()
		buf[k>>3] |= 1 << (k & 7)
	}

	return makeLargeSetFromBytes(modulus, buf)
}

func modPositive(x, n int) int {
	r := x % n
	if r < 0 {
		r += n
	}

	return r
}
