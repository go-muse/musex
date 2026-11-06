package pcset

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-muse/musex/pitchclass"
)

// TranspositionSymmetryIndices returns the set of transposition symmetry
// indices S_T(s) = {n in Z_N : T_n(s) == s} as a [Set] of modulus N.
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) TranspositionSymmetryIndices() (Set, error) {
	if !s.IsValid() {
		return Set{}, errInvalidValue()
	}

	return s.transpositionIndicesToUnchecked(s), nil
}

// MustTranspositionSymmetryIndices is like [Set.TranspositionSymmetryIndices]
// but panics on error.
func (s Set) MustTranspositionSymmetryIndices() Set {
	res, err := s.TranspositionSymmetryIndices()
	if err != nil {
		panic(err)
	}

	return res
}

// InversionSymmetryIndices returns the set of inversion symmetry indices
// S_I(s) = {n in Z_N : I_n(s) == s} as a [Set] of modulus N.
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) InversionSymmetryIndices() (Set, error) {
	if !s.IsValid() {
		return Set{}, errInvalidValue()
	}

	return s.Invert(0).transpositionIndicesToUnchecked(s), nil
}

// MustInversionSymmetryIndices is like [Set.InversionSymmetryIndices] but
// panics on error.
func (s Set) MustInversionSymmetryIndices() Set {
	res, err := s.InversionSymmetryIndices()
	if err != nil {
		panic(err)
	}

	return res
}

// SymmetryOperatorsTI returns all operators f in the T/I group of Z_N that map
// s to itself (f(s) == s), ordered by [pitchclass.Operator.Compare] without
// duplicates.
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) SymmetryOperatorsTI() ([]pitchclass.Operator, error) {
	return s.EquivalenceOperatorsTI(s)
}

// MustSymmetryOperatorsTI is like [Set.SymmetryOperatorsTI] but panics on error.
func (s Set) MustSymmetryOperatorsTI() []pitchclass.Operator {
	ops, err := s.SymmetryOperatorsTI()
	if err != nil {
		panic(err)
	}

	return ops
}

// SymmetryOperatorsTTO returns all operators f in the full affine TTO group of
// Z_N that map s to itself (f(s) == s), ordered by [pitchclass.Operator.Compare]
// without duplicates.
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) SymmetryOperatorsTTO() ([]pitchclass.Operator, error) {
	return s.EquivalenceOperatorsTTO(s)
}

// MustSymmetryOperatorsTTO is like [Set.SymmetryOperatorsTTO] but panics on error.
func (s Set) MustSymmetryOperatorsTTO() []pitchclass.Operator {
	ops, err := s.SymmetryOperatorsTTO()
	if err != nil {
		panic(err)
	}

	return ops
}

// DegreeOfTranspositionSymmetry returns the transpositional degree of symmetry
// DS_T(s) = |{n in Z_N : T_n(s) == s}| >= 1.
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) DegreeOfTranspositionSymmetry() (int, error) {
	if !s.IsValid() {
		return 0, errInvalidValue()
	}

	return s.transpositionIndicesToUnchecked(s).Cardinality(), nil
}

// MustDegreeOfTranspositionSymmetry is like [Set.DegreeOfTranspositionSymmetry]
// but panics on error.
func (s Set) MustDegreeOfTranspositionSymmetry() int {
	d, err := s.DegreeOfTranspositionSymmetry()
	if err != nil {
		panic(err)
	}

	return d
}

// DegreeOfInversionSymmetry returns the inversional degree of symmetry
// DS_I(s) = |{n in Z_N : I_n(s) == s}| in {0, DS_T(s)}.
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) DegreeOfInversionSymmetry() (int, error) {
	if !s.IsValid() {
		return 0, errInvalidValue()
	}

	return s.Invert(0).transpositionIndicesToUnchecked(s).Cardinality(), nil
}

// MustDegreeOfInversionSymmetry is like [Set.DegreeOfInversionSymmetry] but
// panics on error.
func (s Set) MustDegreeOfInversionSymmetry() int {
	d, err := s.DegreeOfInversionSymmetry()
	if err != nil {
		panic(err)
	}

	return d
}

// DegreeOfSymmetryTI returns the degree of symmetry of s in the T/I group:
// DS_T(s) + DS_I(s) for N >= 3, and DS_T(s) for N <= 2 (where I_n coincides
// with T_n).
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) DegreeOfSymmetryTI() (int, error) {
	if !s.IsValid() {
		return 0, errInvalidValue()
	}

	dsT := s.transpositionIndicesToUnchecked(s).Cardinality()
	if s.n <= 2 {
		return dsT, nil
	}

	dsI := s.Invert(0).transpositionIndicesToUnchecked(s).Cardinality()

	return dsT + dsI, nil
}

// MustDegreeOfSymmetryTI is like [Set.DegreeOfSymmetryTI] but panics on error.
func (s Set) MustDegreeOfSymmetryTI() int {
	d, err := s.DegreeOfSymmetryTI()
	if err != nil {
		panic(err)
	}

	return d
}

// DegreeOfSymmetryTTO returns the degree of symmetry of s in the full affine
// TTO group: DS_TTO(s) = |{f in TTO_N : f(s) == s}|.
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) DegreeOfSymmetryTTO() (int, error) {
	if !s.IsValid() {
		return 0, errInvalidValue()
	}

	count := 0
	for m := range pitchclass.MustInvertibleMultipliers(s.n) {
		ms := s.Multiply(m)
		count += ms.transpositionIndicesToUnchecked(s).Cardinality()
	}

	return count, nil
}

// MustDegreeOfSymmetryTTO is like [Set.DegreeOfSymmetryTTO] but panics on error.
func (s Set) MustDegreeOfSymmetryTTO() int {
	d, err := s.DegreeOfSymmetryTTO()
	if err != nil {
		panic(err)
	}

	return d
}

// DistinctTranspositionCount returns the number of distinct transpositions of s
// in Z_N, equal to N / DS_T(s).
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) DistinctTranspositionCount() (int, error) {
	if !s.IsValid() {
		return 0, errInvalidValue()
	}

	dsT := s.transpositionIndicesToUnchecked(s).Cardinality()

	return s.n / dsT, nil
}

// MustDistinctTranspositionCount is like [Set.DistinctTranspositionCount] but
// panics on error.
func (s Set) MustDistinctTranspositionCount() int {
	c, err := s.DistinctTranspositionCount()
	if err != nil {
		panic(err)
	}

	return c
}

// OrbitSizeTI returns the number of distinct sets in the T/I orbit of s:
// 2N / (DS_T(s) + DS_I(s)) for N >= 3, and N / DS_T(s) for N <= 2.
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) OrbitSizeTI() (int, error) {
	if !s.IsValid() {
		return 0, errInvalidValue()
	}

	dsT := s.transpositionIndicesToUnchecked(s).Cardinality()
	if s.n <= 2 {
		return s.n / dsT, nil
	}

	dsI := s.Invert(0).transpositionIndicesToUnchecked(s).Cardinality()

	return (2 * s.n) / (dsT + dsI), nil
}

// MustOrbitSizeTI is like [Set.OrbitSizeTI] but panics on error.
func (s Set) MustOrbitSizeTI() int {
	c, err := s.OrbitSizeTI()
	if err != nil {
		panic(err)
	}

	return c
}

// InvarianceVector is the 8-element Morris invariance vector
// <a, b, c, d, ā, b̄, c̄, d̄> for a pitch-class set in Z_12.
//
// InvarianceVector is a value type that supports the built-in == operator and
// can be used as a map key. Its zero value InvarianceVector{} is invalid
// (IsValid() == false).
type InvarianceVector struct {
	valid   bool
	entries [8]uint8
}

// InvarianceVector computes the 8-element Morris invariance vector
// <a, b, c, d, ā, b̄, c̄, d̄> for a set s in Z_12.
//
// InvarianceVector checks errors in the following order:
//  1. !s.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != 12: [ErrNotModulus12]
func (s Set) InvarianceVector() (InvarianceVector, error) {
	if !s.IsValid() {
		return InvarianceVector{}, errInvalidValue()
	}

	if s.n != 12 {
		return InvarianceVector{}, fmt.Errorf("%w: modulus is %d", ErrNotModulus12, s.n)
	}

	inv0 := s.Invert(0)
	s5 := s.Multiply(5)
	s7 := s.Multiply(7)
	comp := s.Complement()

	return InvarianceVector{
		valid: true,
		entries: [8]uint8{
			uint8(s.transpositionIndicesToUnchecked(s).Cardinality()),
			uint8(inv0.transpositionIndicesToUnchecked(s).Cardinality()),
			uint8(s5.transpositionIndicesToUnchecked(s).Cardinality()),
			uint8(s7.transpositionIndicesToUnchecked(s).Cardinality()),
			uint8(s.transpositionEmbeddingsIntoUnchecked(comp).Cardinality()),
			uint8(inv0.transpositionEmbeddingsIntoUnchecked(comp).Cardinality()),
			uint8(s5.transpositionEmbeddingsIntoUnchecked(comp).Cardinality()),
			uint8(s7.transpositionEmbeddingsIntoUnchecked(comp).Cardinality()),
		},
	}, nil
}

// MustInvarianceVector is like [Set.InvarianceVector] but panics on error.
func (s Set) MustInvarianceVector() InvarianceVector {
	iv, err := s.InvarianceVector()
	if err != nil {
		panic(err)
	}

	return iv
}

// IsValid reports whether v is a valid Morris invariance vector (not the zero
// value InvarianceVector{}).
func (v InvarianceVector) IsValid() bool {
	return v.valid
}

// Modulus returns 12 for a valid invariance vector, or 0 for InvarianceVector{}.
func (v InvarianceVector) Modulus() int {
	if !v.valid {
		return 0
	}

	return 12
}

// Len returns 8 for a valid invariance vector, or 0 for InvarianceVector{}.
func (v InvarianceVector) Len() int {
	if !v.valid {
		return 0
	}

	return 8
}

// T returns component a = |{n in Z_12 : T_n(X) == X}|, or 0 for InvarianceVector{}.
func (v InvarianceVector) T() int {
	return int(v.entries[0])
}

// I returns component b = |{n in Z_12 : I_n(X) == X}|, or 0 for InvarianceVector{}.
func (v InvarianceVector) I() int {
	return int(v.entries[1])
}

// M5 returns component c = |{n in Z_12 : T_n M_5(X) == X}|, or 0 for InvarianceVector{}.
func (v InvarianceVector) M5() int {
	return int(v.entries[2])
}

// M7 returns component d = |{n in Z_12 : T_n M_7(X) == X}|, or 0 for InvarianceVector{}.
func (v InvarianceVector) M7() int {
	return int(v.entries[3])
}

// ComplementT returns component ā = |{n in Z_12 : T_n(X) ⊆ X̄}|, or 0 for InvarianceVector{}.
func (v InvarianceVector) ComplementT() int {
	return int(v.entries[4])
}

// ComplementI returns component b̄ = |{n in Z_12 : I_n(X) ⊆ X̄}|, or 0 for InvarianceVector{}.
func (v InvarianceVector) ComplementI() int {
	return int(v.entries[5])
}

// ComplementM5 returns component c̄ = |{n in Z_12 : T_n M_5(X) ⊆ X̄}|, or 0 for InvarianceVector{}.
func (v InvarianceVector) ComplementM5() int {
	return int(v.entries[6])
}

// ComplementM7 returns component d̄ = |{n in Z_12 : T_n M_7(X) ⊆ X̄}|, or 0 for InvarianceVector{}.
func (v InvarianceVector) ComplementM7() int {
	return int(v.entries[7])
}

// SymmetryTotal returns a + b + c + d = DS_TTO(X), or 0 for InvarianceVector{}.
func (v InvarianceVector) SymmetryTotal() int {
	return int(v.entries[0]) + int(v.entries[1]) + int(v.entries[2]) + int(v.entries[3])
}

// ComplementTotal returns ā + b̄ + c̄ + d̄, or 0 for InvarianceVector{}.
func (v InvarianceVector) ComplementTotal() int {
	return int(v.entries[4]) + int(v.entries[5]) + int(v.entries[6]) + int(v.entries[7])
}

// Slice returns a newly allocated 8-element slice [a, b, c, d, ā, b̄, c̄, d̄]
// for a valid vector, or nil for InvarianceVector{}.
func (v InvarianceVector) Slice() []int {
	if !v.valid {
		return nil
	}

	res := make([]int, 8)
	for i, e := range v.entries {
		res[i] = int(e)
	}

	return res
}

// Equal reports whether v and other are both valid Morris invariance vectors
// with identical components. If either operand is InvarianceVector{}, Equal
// returns false.
func (v InvarianceVector) Equal(other InvarianceVector) bool {
	return v.valid && other.valid && v == other
}

// Compare compares v and other and returns -1, 0, or +1:
// InvarianceVector{} is less than any valid vector, and two valid vectors are
// compared lexicographically by their 8 components.
func (v InvarianceVector) Compare(other InvarianceVector) int {
	if !v.valid && !other.valid {
		return 0
	}

	if !v.valid {
		return -1
	}

	if !other.valid {
		return 1
	}

	for i := range 8 {
		if c := cmp.Compare(v.entries[i], other.entries[i]); c != 0 {
			return c
		}
	}

	return 0
}

// String returns "<a, b, c, d, ā, b̄, c̄, d̄> mod 12" for a valid vector, or
// "InvarianceVector(invalid)" for InvarianceVector{}.
func (v InvarianceVector) String() string {
	if !v.valid {
		return "InvarianceVector(invalid)"
	}

	return v.formatDelimited() + " mod 12"
}

// Notation returns the delimited literary representation
// "<a, b, c, d, ā, b̄, c̄, d̄>" without the modulus suffix.
//
// It returns [ErrInvalidValue] if !v.IsValid().
func (v InvarianceVector) Notation() (string, error) {
	if !v.valid {
		return "", errInvalidValue()
	}

	return v.formatDelimited(), nil
}

// MustNotation is like [InvarianceVector.Notation] but panics on error.
func (v InvarianceVector) MustNotation() string {
	s, err := v.Notation()
	if err != nil {
		panic(err)
	}

	return s
}

func (v InvarianceVector) formatDelimited() string {
	var b strings.Builder
	b.Grow(32)
	b.WriteByte('<')

	for i, e := range v.entries {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(strconv.Itoa(int(e)))
	}

	b.WriteByte('>')

	return b.String()
}
