package pitchclass

import (
	"cmp"
	"fmt"
	"strconv"
)

// Operator is a transformation of the classes of N positions of the form
// x ↦ m·x + n (mod N) with a multiplier m that is invertible modulo N and an
// index n, both in the range 0 to N-1. It is written TnMm: multiply by m,
// then transpose by n. For N = 12, the 48 operators are the twelve-tone
// operators (TTOs) of Morris.
//
// Transpositions Tn have m = 1 mod N, and inversions In have m = (N-1) mod
// N. For N <= 2 these coincide: each transposition is also an inversion, and
// for N = 1 the single operator T0M0 is both.
//
// Create an operator with [NewOperator], [Identity], [Transposition],
// [Inversion], [InversionAround], [InversionSwapping], [Multiplication], or
// [ParseOperator], from the operator groups such as [Operators], or with
// [Operator.Compose] and [Operator.Inverse].
//
// The zero value is not an operator: it is never treated as the identity
// T0. [Operator.IsValid] reports false for it, and the operations recognize
// it as described in the package documentation.
//
// Operator is an immutable value. It can be copied by assignment and passed
// by value at any time; copies share no data. No method changes its receiver
// or its arguments.
//
// Operator is comparable: == can be used and an Operator can be a map key.
// For valid operators, == and [Operator.Equal] agree: two operators are
// equal exactly when they have the same N, m, and n, that is, when they map
// every class in the same way. The only difference is the zero value:
// Operator{} == Operator{} is true, but [Operator.Equal] reports false for
// it.
//
// Operator has no exported fields and no serialization of its own:
// encoding/json writes {} for it. To store an operator, use
// [Operator.String] and [ParseOperator].
type Operator struct {
	// n is the number of positions; it is 0 only for the zero value.
	n int
	// m is the multiplier, an invertible residue modulo n.
	m int
	// t is the index (the transposition), a residue modulo n.
	t int
}

// NewOperator returns the operator x ↦ m·x + n (mod N) with m = multiplier
// mod N, n = index mod N, and N = modulus. Any int multiplier and index are
// accepted and reduced around the circle: NewOperator(17, -7, 12) equals
// NewOperator(5, 5, 12).
//
// If modulus is less than 1, NewOperator returns the zero Operator and an
// error that wraps [ErrInvalidModulus]. Otherwise, if the multiplier is not
// invertible modulo N, it returns the zero Operator and an error that wraps
// [ErrNotInvertible]; for example, the multipliers 0, 2, and 6 are not
// invertible modulo 12.
func NewOperator(multiplier, index, modulus int) (Operator, error) {
	if modulus < 1 {
		return Operator{}, errModulus(modulus)
	}

	m := mod(multiplier, modulus)
	if !invertible(m, modulus) {
		return Operator{}, errNotInvertible(multiplier, modulus)
	}

	return Operator{n: modulus, m: m, t: mod(index, modulus)}, nil
}

// MustNewOperator is like [NewOperator] but panics with its error instead of
// returning it. It is intended for literals and values known to be valid.
func MustNewOperator(multiplier, index, modulus int) Operator {
	f, err := NewOperator(multiplier, index, modulus)
	if err != nil {
		panic(err)
	}

	return f
}

// Identity returns the identity operator T0 of modulus positions, which maps
// every class to itself. It equals Transposition(0, modulus).
//
// If modulus is less than 1, Identity returns the zero Operator and an error
// that wraps [ErrInvalidModulus].
func Identity(modulus int) (Operator, error) {
	return Transposition(0, modulus)
}

// MustIdentity is like [Identity] but panics with its error instead of
// returning it. It is intended for values known to be valid.
func MustIdentity(modulus int) Operator {
	f, err := Identity(modulus)
	if err != nil {
		panic(err)
	}

	return f
}

// Transposition returns the transposition Tn with n = index mod N of
// modulus positions: x ↦ x + n. Any int index is accepted. Its
// [Operator.Apply] equals [Class.Transpose].
//
// If modulus is less than 1, Transposition returns the zero Operator and an
// error that wraps [ErrInvalidModulus].
func Transposition(index, modulus int) (Operator, error) {
	if modulus < 1 {
		return Operator{}, errModulus(modulus)
	}

	return Operator{n: modulus, m: 1 % modulus, t: mod(index, modulus)}, nil
}

// MustTransposition is like [Transposition] but panics with its error
// instead of returning it. It is intended for literals and values known to
// be valid.
func MustTransposition(index, modulus int) Operator {
	f, err := Transposition(index, modulus)
	if err != nil {
		panic(err)
	}

	return f
}

// Inversion returns the inversion In with n = index mod N of modulus
// positions: x ↦ n - x, also written TnI. Any int index is accepted. Its
// [Operator.Apply] equals [Class.Invert].
//
// If modulus is less than 1, Inversion returns the zero Operator and an
// error that wraps [ErrInvalidModulus].
func Inversion(index, modulus int) (Operator, error) {
	if modulus < 1 {
		return Operator{}, errModulus(modulus)
	}

	return Operator{n: modulus, m: (modulus - 1) % modulus, t: mod(index, modulus)}, nil
}

// MustInversion is like [Inversion] but panics with its error instead of
// returning it. It is intended for literals and values known to be valid.
func MustInversion(index, modulus int) Operator {
	f, err := Inversion(index, modulus)
	if err != nil {
		panic(err)
	}

	return f
}

// InversionAround returns the inversion around the axis class a, I(2a), for
// the number of positions of the axis. The axis stays in place:
// InversionAround of 2 mod 12 equals Inversion(4, 12).
//
// If axis is the zero value, InversionAround returns the zero Operator and
// an error that wraps [ErrInvalidValue].
func InversionAround(axis Class) (Operator, error) {
	if !axis.IsValid() {
		return Operator{}, errValue()
	}

	return Inversion(addMod(axis.k, axis.k, axis.n), axis.n)
}

// MustInversionAround is like [InversionAround] but panics with its error
// instead of returning it. It is intended for values known to be valid.
func MustInversionAround(axis Class) Operator {
	f, err := InversionAround(axis)
	if err != nil {
		panic(err)
	}

	return f
}

// InversionSwapping returns the inversion that exchanges the classes x and
// y, I(x+y), for their number of positions. The pair (x, x) gives the
// inversion around x. For example, the pair 0 mod 12 and 4 mod 12 gives
// Inversion(4, 12).
//
// If x or y is the zero value, InversionSwapping returns the zero Operator
// and an error that wraps [ErrInvalidValue]. Otherwise, if they have
// different N, it returns the zero Operator and an error that wraps
// [ErrModulusMismatch].
func InversionSwapping(x, y Class) (Operator, error) {
	if err := checkPair(x, y); err != nil {
		return Operator{}, err
	}

	return Inversion(addMod(x.k, y.k, x.n), x.n)
}

// MustInversionSwapping is like [InversionSwapping] but panics with its
// error instead of returning it. It is intended for values known to be
// valid.
func MustInversionSwapping(x, y Class) Operator {
	f, err := InversionSwapping(x, y)
	if err != nil {
		panic(err)
	}

	return f
}

// Multiplication returns the multiplication Mm with m = multiplier mod N of
// modulus positions: x ↦ m·x. Any int multiplier is accepted. Its
// [Operator.Apply] equals [Class.Multiply]. For N = 12, M7 maps the
// chromatic circle onto the circle of fifths.
//
// If modulus is less than 1, Multiplication returns the zero Operator and an
// error that wraps [ErrInvalidModulus]. Otherwise, if the multiplier is not
// invertible modulo N, it returns the zero Operator and an error that wraps
// [ErrNotInvertible]; [Class.Multiply] accepts such multipliers.
func Multiplication(multiplier, modulus int) (Operator, error) {
	return NewOperator(multiplier, 0, modulus)
}

// MustMultiplication is like [Multiplication] but panics with its error
// instead of returning it. It is intended for literals and values known to
// be valid.
func MustMultiplication(multiplier, modulus int) Operator {
	f, err := Multiplication(multiplier, modulus)
	if err != nil {
		panic(err)
	}

	return f
}

// Modulus returns the number of positions N of the operator, at least 1.
// For the zero value, it returns 0.
func (f Operator) Modulus() int {
	return f.n
}

// Multiplier returns the multiplier m of the operator, in the range 0 to
// N-1. For the zero value, it returns -1.
func (f Operator) Multiplier() int {
	if f.n == 0 {
		return -1
	}

	return f.m
}

// Index returns the index n of the operator x ↦ m·x + n, in the range 0 to
// N-1: the n of Tn and In, and of TnMm. For the zero value, it returns -1.
func (f Operator) Index() int {
	if f.n == 0 {
		return -1
	}

	return f.t
}

// IsValid reports whether f is an operator. It is true for every operator
// returned without an error and false only for the zero value.
func (f Operator) IsValid() bool {
	return f.n != 0
}

// IsTransposition reports whether f is a transposition Tn: its multiplier is
// 1 mod N. For N = 1 the single operator, and for N = 2 every operator, is
// both a transposition and an inversion. It is false for the zero value.
func (f Operator) IsTransposition() bool {
	return f.n != 0 && f.m == 1%f.n
}

// IsInversion reports whether f is an inversion In: its multiplier is
// (N-1) mod N. For N <= 2 inversions coincide with transpositions. It is
// false for the zero value.
func (f Operator) IsInversion() bool {
	return f.n != 0 && f.m == (f.n-1)%f.n
}

// Apply returns f(c) = (m·k + n) mod N for the class c = k mod N. For valid
// operands, f permutes the N classes. For example, T5M7 maps 1 mod 12 to
// 0 mod 12.
//
// If f or c is the zero value, Apply returns the zero Class and an error
// that wraps [ErrInvalidValue]. Otherwise, if f and c have different N, it
// returns the zero Class and an error that wraps [ErrModulusMismatch].
func (f Operator) Apply(c Class) (Class, error) {
	if !f.IsValid() || !c.IsValid() {
		return Class{}, errValue()
	}

	if f.n != c.n {
		return Class{}, fmt.Errorf("%w: operator has %d positions, class has %d", ErrModulusMismatch, f.n, c.n)
	}

	return Class{k: addMod(mulMod(f.m, c.k, f.n), f.t, f.n), n: f.n}, nil
}

// MustApply is like [Operator.Apply] but panics with its error instead of
// returning it. It is intended for values known to be valid.
func (f Operator) MustApply(c Class) Class {
	r, err := f.Apply(c)
	if err != nil {
		panic(err)
	}

	return r
}

// Compose returns the composition f∘g, which applies g first and then f:
// f.Compose(g).Apply(x) equals f.Apply(g.Apply(x)). Composition is
// associative but in general not commutative: modulo 12, T3∘I8 is I11, and
// I8∘T3 is I5.
//
// If f or g is the zero value, Compose returns the zero Operator and an
// error that wraps [ErrInvalidValue]. Otherwise, if they have different N,
// it returns the zero Operator and an error that wraps [ErrModulusMismatch].
func (f Operator) Compose(g Operator) (Operator, error) {
	if !f.IsValid() || !g.IsValid() {
		return Operator{}, errValue()
	}

	if f.n != g.n {
		return Operator{}, fmt.Errorf("%w: operators have %d and %d positions", ErrModulusMismatch, f.n, g.n)
	}

	return Operator{
		n: f.n,
		m: mulMod(f.m, g.m, f.n),
		t: addMod(mulMod(f.m, g.t, f.n), f.t, f.n),
	}, nil
}

// MustCompose is like [Operator.Compose] but panics with its error instead
// of returning it. It is intended for values known to be valid.
func (f Operator) MustCompose(g Operator) Operator {
	r, err := f.Compose(g)
	if err != nil {
		panic(err)
	}

	return r
}

// Inverse returns the inverse operator f⁻¹: f.Compose(f.Inverse()) and
// f.Inverse().Compose(f) equal the identity of N. The inverse of TnMm is
// T(-m⁻¹·n)M(m⁻¹); the inverse of Tn is T(-n), and every inversion In is its
// own inverse.
//
// For the zero value, Inverse returns the zero value.
func (f Operator) Inverse() Operator {
	if f.n == 0 {
		return Operator{}
	}

	mi := inverseMod(f.m, f.n)

	return Operator{n: f.n, m: mi, t: negMod(mulMod(mi, f.t, f.n), f.n)}
}

// Equal reports whether f and g are both valid and are the same operator:
// they have equal N, multipliers, and indices, so they map every class in
// the same way. Operators of different N are not equal.
//
// The zero value is not equal to any value, including itself.
func (f Operator) Equal(g Operator) bool {
	return f.IsValid() && f == g
}

// Compare returns -1, 0, or +1 depending on whether f is ordered before,
// equal to, or after g. Operators are ordered by N, then by multiplier, then
// by index; for one N this is the order of [Operators]. Operators of
// different N are compared without an error.
//
// For valid operators, Compare returns 0 exactly when [Operator.Equal]
// reports true. The zero value comes before every valid operator, and
// Compare of the zero value with itself is 0.
func (f Operator) Compare(g Operator) int {
	if r := cmp.Compare(f.n, g.n); r != 0 {
		return r
	}

	if r := cmp.Compare(f.m, g.m); r != 0 {
		return r
	}

	return cmp.Compare(f.t, g.t)
}

// String returns the canonical form of the operator, "T<n>M<m> mod <N>":
// multiply by m, then transpose by n. The numbers are written as in
// [Class.String]. For example, T5M7 modulo 12 is "T5M7 mod 12", the
// identity modulo 12 is "T0M1 mod 12", I5 modulo 12 is "T5M11 mod 12", and
// the single operator modulo 1 is "T0M0 mod 1". Different operators have
// different forms, and [ParseOperator] reverses String.
//
// For the zero value, String returns "Operator(invalid)", which is not the
// form of any operator and is not accepted by [ParseOperator].
func (f Operator) String() string {
	if f.n == 0 {
		return "Operator(invalid)"
	}

	return "T" + strconv.Itoa(f.t) + "M" + strconv.Itoa(f.m) + modSeparator + strconv.Itoa(f.n)
}
