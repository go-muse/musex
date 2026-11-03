package pitchclass

import (
	"cmp"
	"fmt"
	"iter"
	"strconv"
)

// Class is a pitch class: a position number k within the octave of a system
// with N positions, written "k mod N". For example, 1 mod 12 is the class of
// C♯ and D♭ in the traditional 12-position system with C = 0.
//
// A class holds only its number and N. It does not remember a system, an
// octave, a note name, or a pitch: classes of different systems with the
// same N are equal, and classes of different N are never equal.
//
// Create a class with [New], [Reduce], [Parse], [ParseNotation], or [All],
// or as the result of an operation on classes or operators. A valid class
// has 1 <= N <= math.MaxInt and 0 <= k < N.
//
// The zero value is not a class: it means that no class is given, and it is
// never treated as 0 mod 12. [Class.IsValid] reports false for it, and the
// operations recognize it as described in the package documentation.
//
// Class is an immutable value. It can be copied by assignment and passed by
// value at any time; copies share no data. No method changes its receiver or
// its arguments.
//
// Class is comparable: == can be used and a Class can be a map key. For
// valid classes, == and [Class.Equal] agree. The only difference is the zero
// value: Class{} == Class{} is true, but [Class.Equal] reports false for it.
//
// Class has no exported fields and no serialization of its own:
// encoding/json writes {} for it. To store a class, use [Class.String] and
// [Parse], or [Class.Number] and [Class.Modulus] with [New].
type Class struct {
	// k is the number of the class, in the range 0 to n-1.
	k int
	// n is the number of positions; it is 0 only for the zero value.
	n int
}

// New returns the class number mod modulus.
//
// The modulus must be at least 1, and number must be in the range 0 to
// modulus-1. New does not wrap number around the circle; use [Reduce] for
// that. Otherwise New returns the zero Class and an error that wraps
// [ErrInvalidClass].
func New(number, modulus int) (Class, error) {
	if modulus < 1 {
		return Class{}, fmt.Errorf("%w: modulus %d is less than 1", ErrInvalidClass, modulus)
	}

	if number < 0 || number >= modulus {
		return Class{}, fmt.Errorf("%w: number %d is outside the range 0 to %d", ErrInvalidClass, number, modulus-1)
	}

	return Class{k: number, n: modulus}, nil
}

// MustNew is like [New] but panics with the error of [New] instead of
// returning it. It is intended for literals and values known to be valid.
func MustNew(number, modulus int) Class {
	c, err := New(number, modulus)
	if err != nil {
		panic(err)
	}

	return c
}

// Reduce returns the class of x modulo modulus: x mod modulus, where mod is
// the Euclidean remainder in the range 0 to modulus-1. Any int x is
// accepted, including negative values and math.MinInt: Reduce(-1, 12) is
// 11 mod 12, and Reduce(13, 12) is 1 mod 12.
//
// For x in the range 0 to modulus-1, Reduce equals [New]. If modulus is less
// than 1, Reduce returns the zero Class and an error that wraps
// [ErrInvalidClass].
func Reduce(x, modulus int) (Class, error) {
	if modulus < 1 {
		return Class{}, fmt.Errorf("%w: modulus %d is less than 1", ErrInvalidClass, modulus)
	}

	return Class{k: mod(x, modulus), n: modulus}, nil
}

// MustReduce is like [Reduce] but panics with the error of [Reduce] instead
// of returning it. It is intended for literals and values known to be valid.
func MustReduce(x, modulus int) Class {
	c, err := Reduce(x, modulus)
	if err != nil {
		panic(err)
	}

	return c
}

// All returns an iterator over the modulus classes of modulus positions in
// ascending order of their numbers: 0 mod N, 1 mod N, ..., N-1 mod N.
//
// The iterator is lazy: it computes each class when it is reached, and the
// first k classes take O(k) time and constant memory for any N. Each range
// over it starts from the first class, and break stops it. The iterator may
// be used by several goroutines at once. Use [slices.Collect] to get a slice.
//
// If modulus is less than 1, All returns an error that wraps
// [ErrInvalidModulus] and an iterator that yields nothing; the iterator is
// never nil.
func All(modulus int) (iter.Seq[Class], error) {
	if modulus < 1 {
		return emptySeq[Class](), errModulus(modulus)
	}

	return func(yield func(Class) bool) {
		for k := range modulus {
			if !yield(Class{k: k, n: modulus}) {
				return
			}
		}
	}, nil
}

// MustAll is like [All] but panics with the error of [All] instead of
// returning it. It is intended for values known to be valid.
func MustAll(modulus int) iter.Seq[Class] {
	seq, err := All(modulus)
	if err != nil {
		panic(err)
	}

	return seq
}

// Number returns the number k of the class, in the range 0 to N-1. For the
// zero value, it returns -1, which no class has.
func (c Class) Number() int {
	if c.n == 0 {
		return -1
	}

	return c.k
}

// Modulus returns the number of positions N of the class, at least 1. For
// the zero value, it returns 0, which no class has.
func (c Class) Modulus() int {
	return c.n
}

// IsValid reports whether c is a class. It is true for every class returned
// without an error and false only for the zero value.
func (c Class) IsValid() bool {
	return c.n != 0
}

// Equal reports whether c and other are both valid and are the same class:
// they have equal numbers and equal N. Classes of different N are not equal.
//
// The zero value is not equal to any value, including itself:
// Class{}.Equal(Class{}) is false, although Class{} == Class{} is true.
func (c Class) Equal(other Class) bool {
	return c.IsValid() && c == other
}

// Compare returns -1, 0, or +1 depending on whether c is ordered before,
// equal to, or after other. Classes are ordered by N first and then by
// number, so 5 mod 12 comes before 0 mod 19. This is an order of labels for
// sorting and stable output, not an order of pitch: the circle of classes
// has no beginning.
//
// For valid classes, Compare returns 0 exactly when [Class.Equal] reports
// true. The zero value comes before every valid class, and Compare of the
// zero value with itself is 0, although [Class.Equal] reports false for it.
//
// The method expression Class.Compare can be passed to [slices.SortFunc].
func (c Class) Compare(other Class) int {
	if r := cmp.Compare(c.n, other.n); r != 0 {
		return r
	}

	return cmp.Compare(c.k, other.k)
}

// Cycle returns an iterator over the N classes of c's number of positions,
// starting at c and moving upward around the circle: T0(c), T1(c), ...,
// T(N-1)(c). Each class appears exactly once; from 10 mod 12 the order is
// 10, 11, 0, 1, ..., 9. To move downward, use c.Transpose(-i).
//
// The iterator is lazy, restartable, and may be stopped with break, as
// described for [All]. For the zero value, it yields nothing.
func (c Class) Cycle() iter.Seq[Class] {
	return func(yield func(Class) bool) {
		k := c.k
		for range c.n {
			if !yield(Class{k: k, n: c.n}) {
				return
			}

			k++
			if k == c.n {
				k = 0
			}
		}
	}
}

// Transpose returns Tn(c), the class n positions above c around the circle:
// (k + n) mod N. Any int n is accepted; a negative n moves downward. For
// example, 11 mod 12 transposed by 2 is 1 mod 12.
//
// For the zero value, Transpose returns the zero value.
func (c Class) Transpose(n int) Class {
	if c.n == 0 {
		return Class{}
	}

	return Class{k: addMod(c.k, mod(n, c.n), c.n), n: c.n}
}

// IntervalTo returns the ordered pitch-class interval i<c, to> = (to - c)
// mod N: the distance from c upward around the circle to to, in the range 0
// to N-1. to equals c.Transpose(i). For example, from 10 mod 12 to 2 mod 12
// the interval is 4, and from 2 mod 12 to 10 mod 12 it is 8.
//
// The complementary interval of the pair is to.IntervalTo(c).
//
// If c or to is the zero value, IntervalTo returns 0 and an error that wraps
// [ErrInvalidValue]. Otherwise, if the classes have different N, it returns
// 0 and an error that wraps [ErrModulusMismatch].
func (c Class) IntervalTo(to Class) (int, error) {
	if err := checkPair(c, to); err != nil {
		return 0, err
	}

	return subMod(to.k, c.k, c.n), nil
}

// MustIntervalTo is like [Class.IntervalTo] but panics with its error
// instead of returning it. It is intended for values known to be valid.
func (c Class) MustIntervalTo(to Class) int {
	i, err := c.IntervalTo(to)
	if err != nil {
		panic(err)
	}

	return i
}

// IntervalClass returns the interval class ic(c, other): the smaller of the
// ordered intervals between c and other in the two directions, in the range
// 0 to N/2 (rounded down). It is symmetric and is 0 exactly when the classes
// are equal. For example, ic(0 mod 12, 7 mod 12) is 5.
//
// Errors are as for [Class.IntervalTo]; on error, IntervalClass returns 0.
// For an interval given by its size, use the function [IntervalClass].
func (c Class) IntervalClass(other Class) (int, error) {
	i, err := c.IntervalTo(other)
	if err != nil {
		return 0, err
	}

	return min(i, negMod(i, c.n)), nil
}

// MustIntervalClass is like [Class.IntervalClass] but panics with its error
// instead of returning it. It is intended for values known to be valid.
func (c Class) MustIntervalClass(other Class) int {
	i, err := c.IntervalClass(other)
	if err != nil {
		panic(err)
	}

	return i
}

// IndexSum returns the index sum (c + other) mod N, in the range 0 to N-1:
// the index s of the inversion that exchanges c and other, so that
// c.Invert(s) equals other and other.Invert(s) equals c. For example, the
// index sum of 4 mod 12 and 9 mod 12 is 1.
//
// Errors are as for [Class.IntervalTo]; on error, IndexSum returns 0.
func (c Class) IndexSum(other Class) (int, error) {
	if err := checkPair(c, other); err != nil {
		return 0, err
	}

	return addMod(c.k, other.k, c.n), nil
}

// MustIndexSum is like [Class.IndexSum] but panics with its error instead of
// returning it. It is intended for values known to be valid.
func (c Class) MustIndexSum(other Class) int {
	s, err := c.IndexSum(other)
	if err != nil {
		panic(err)
	}

	return s
}

// Invert returns In(c) = (n - k) mod N for n = index: the inversion with
// index n. In and TnI are two notations of the same operation; the
// inversion written I without an index is I0, that is, c.Invert(0). Any int
// index is accepted. For example, I8 maps 2 mod 12 to 6 mod 12 and
// 4 mod 12 to itself.
//
// For the zero value, Invert returns the zero value.
func (c Class) Invert(index int) Class {
	if c.n == 0 {
		return Class{}
	}

	return Class{k: subMod(mod(index, c.n), c.k, c.n), n: c.n}
}

// InvertAround returns the inversion of c around the axis class a: I(2a)(c).
// The axis class stays in place. An axis between two classes is given by an
// index ([Class.Invert]) or by a pair ([Class.InvertSwapping]) instead.
// For example, around 2 mod 12, 0 mod 12 maps to 4 mod 12.
//
// If c or axis is the zero value, InvertAround returns the zero Class and an
// error that wraps [ErrInvalidValue]. Otherwise, if they have different N,
// it returns the zero Class and an error that wraps [ErrModulusMismatch].
func (c Class) InvertAround(axis Class) (Class, error) {
	if err := checkPair(c, axis); err != nil {
		return Class{}, err
	}

	return c.Invert(addMod(axis.k, axis.k, c.n)), nil
}

// MustInvertAround is like [Class.InvertAround] but panics with its error
// instead of returning it. It is intended for values known to be valid.
func (c Class) MustInvertAround(axis Class) Class {
	r, err := c.InvertAround(axis)
	if err != nil {
		panic(err)
	}

	return r
}

// InvertSwapping returns the inversion of c that exchanges the classes x and
// y: I(x+y)(c). The pair (x, x) gives the inversion around x. For example,
// the pair 0 mod 12 and 4 mod 12 gives I4, which maps 7 mod 12 to
// 9 mod 12; the pairs (3, 5), (2, 6), and (4, 4) mod 12 all give I8.
//
// If c, x, or y is the zero value, InvertSwapping returns the zero Class and
// an error that wraps [ErrInvalidValue]. Otherwise, if the three classes do
// not all have the same N, it returns the zero Class and an error that wraps
// [ErrModulusMismatch].
func (c Class) InvertSwapping(x, y Class) (Class, error) {
	if !c.IsValid() || !x.IsValid() || !y.IsValid() {
		return Class{}, errValue()
	}

	if c.n != x.n || c.n != y.n {
		return Class{}, fmt.Errorf("%w: classes have %d, %d, and %d positions", ErrModulusMismatch, c.n, x.n, y.n)
	}

	return c.Invert(addMod(x.k, y.k, c.n)), nil
}

// MustInvertSwapping is like [Class.InvertSwapping] but panics with its
// error instead of returning it. It is intended for values known to be
// valid.
func (c Class) MustInvertSwapping(x, y Class) Class {
	r, err := c.InvertSwapping(x, y)
	if err != nil {
		panic(err)
	}

	return r
}

// Multiply returns Mm(c) = (m * k) mod N. Any int m is accepted. The result
// is exact for all inputs; the product is computed in 128 bits. For example,
// M7 maps 1 mod 12 to 7 mod 12 and 2 mod 12 to 2 mod 12.
//
// A multiplier that is not invertible modulo N is not an error: Mm then maps
// different classes to the same class, and M0 maps every class to 0. See
// [Invertible].
//
// For the zero value, Multiply returns the zero value.
func (c Class) Multiply(m int) Class {
	if c.n == 0 {
		return Class{}
	}

	return Class{k: mulMod(mod(m, c.n), c.k, c.n), n: c.n}
}

// String returns the canonical form of the class, "<k> mod <N>", such as
// "1 mod 12" or "0 mod 1". Both numbers are written in decimal ASCII digits
// without a sign or leading zeros, separated from "mod" by single spaces.
// Different classes have different forms, and [Parse] reverses String.
//
// For the zero value, String returns "Class(invalid)", which is not the form
// of any class and is not accepted by [Parse].
func (c Class) String() string {
	if c.n == 0 {
		return "Class(invalid)"
	}

	return strconv.Itoa(c.k) + modSeparator + strconv.Itoa(c.n)
}

// checkPair returns the error of a two-place operation on the classes a and
// b: an error that wraps [ErrInvalidValue] if either is the zero value, an
// error that wraps [ErrModulusMismatch] if they have different N, or nil.
func checkPair(a, b Class) error {
	if !a.IsValid() || !b.IsValid() {
		return errValue()
	}

	if a.n != b.n {
		return fmt.Errorf("%w: classes have %d and %d positions", ErrModulusMismatch, a.n, b.n)
	}

	return nil
}

// errValue returns the error for a zero Class or Operator operand.
func errValue() error {
	return fmt.Errorf("%w: zero value operand", ErrInvalidValue)
}

// errModulus returns the error for a number of positions less than 1.
func errModulus(modulus int) error {
	return fmt.Errorf("%w: modulus %d is less than 1", ErrInvalidModulus, modulus)
}
