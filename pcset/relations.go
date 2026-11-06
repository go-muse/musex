package pcset

import "github.com/go-muse/musex/pitchclass"

// TranspositionIndicesTo returns the set of transposition indices
// {n in Z_N : T_n(s) == target} as a [Set] of modulus N.
//
// TranspositionIndicesTo checks errors in the following order:
//  1. !s.IsValid() || !target.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != target.Modulus(): [ErrModulusMismatch]
func (s Set) TranspositionIndicesTo(target Set) (Set, error) {
	if err := checkSetPair(s, target); err != nil {
		return Set{}, err
	}

	return s.transpositionIndicesToUnchecked(target), nil
}

// MustTranspositionIndicesTo is like [Set.TranspositionIndicesTo] but panics on error.
func (s Set) MustTranspositionIndicesTo(target Set) Set {
	res, err := s.TranspositionIndicesTo(target)
	if err != nil {
		panic(err)
	}

	return res
}

// InversionIndicesTo returns the set of inversion indices
// {n in Z_N : I_n(s) == target} as a [Set] of modulus N.
//
// InversionIndicesTo checks errors in the following order:
//  1. !s.IsValid() || !target.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != target.Modulus(): [ErrModulusMismatch]
func (s Set) InversionIndicesTo(target Set) (Set, error) {
	if err := checkSetPair(s, target); err != nil {
		return Set{}, err
	}

	return s.Invert(0).transpositionIndicesToUnchecked(target), nil
}

// MustInversionIndicesTo is like [Set.InversionIndicesTo] but panics on error.
func (s Set) MustInversionIndicesTo(target Set) Set {
	res, err := s.InversionIndicesTo(target)
	if err != nil {
		panic(err)
	}

	return res
}

// TranspositionEmbeddingsInto returns the set of transposition indices
// {n in Z_N : T_n(s) ⊆ superset} as a [Set] of modulus N.
//
// TranspositionEmbeddingsInto checks errors in the following order:
//  1. !s.IsValid() || !superset.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != superset.Modulus(): [ErrModulusMismatch]
func (s Set) TranspositionEmbeddingsInto(superset Set) (Set, error) {
	if err := checkSetPair(s, superset); err != nil {
		return Set{}, err
	}

	return s.transpositionEmbeddingsIntoUnchecked(superset), nil
}

// MustTranspositionEmbeddingsInto is like [Set.TranspositionEmbeddingsInto] but panics on error.
func (s Set) MustTranspositionEmbeddingsInto(superset Set) Set {
	res, err := s.TranspositionEmbeddingsInto(superset)
	if err != nil {
		panic(err)
	}

	return res
}

// CoveringTranspositionsOf returns the set of template transposition indices
// (tonics) {tau in Z_N : s ⊆ T_tau(template)} as a [Set] of modulus N.
// When s is empty, it returns [Universe](N).
//
// CoveringTranspositionsOf checks errors in the following order:
//  1. !s.IsValid() || !template.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != template.Modulus(): [ErrModulusMismatch]
func (s Set) CoveringTranspositionsOf(template Set) (Set, error) {
	if err := checkSetPair(s, template); err != nil {
		return Set{}, err
	}

	return s.transpositionEmbeddingsIntoUnchecked(template).Invert(0), nil
}

// MustCoveringTranspositionsOf is like [Set.CoveringTranspositionsOf] but panics on error.
func (s Set) MustCoveringTranspositionsOf(template Set) Set {
	res, err := s.CoveringTranspositionsOf(template)
	if err != nil {
		panic(err)
	}

	return res
}

// InversionEmbeddingsInto returns the set of inversion indices
// {n in Z_N : I_n(s) ⊆ superset} as a [Set] of modulus N.
//
// InversionEmbeddingsInto checks errors in the following order:
//  1. !s.IsValid() || !superset.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != superset.Modulus(): [ErrModulusMismatch]
func (s Set) InversionEmbeddingsInto(superset Set) (Set, error) {
	if err := checkSetPair(s, superset); err != nil {
		return Set{}, err
	}

	return s.Invert(0).transpositionEmbeddingsIntoUnchecked(superset), nil
}

// MustInversionEmbeddingsInto is like [Set.InversionEmbeddingsInto] but panics on error.
func (s Set) MustInversionEmbeddingsInto(superset Set) Set {
	res, err := s.InversionEmbeddingsInto(superset)
	if err != nil {
		panic(err)
	}

	return res
}

// EquivalenceOperatorsTI returns all operators f in the T/I group of Z_N such
// that f(s) == target, ordered by [pitchclass.Operator.Compare] without
// duplicates. If no operators match, it returns a non-nil empty slice.
//
// EquivalenceOperatorsTI checks errors in the following order:
//  1. !s.IsValid() || !target.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != target.Modulus(): [ErrModulusMismatch]
func (s Set) EquivalenceOperatorsTI(target Set) ([]pitchclass.Operator, error) {
	if err := checkSetPair(s, target); err != nil {
		return nil, err
	}

	tSet := s.transpositionIndicesToUnchecked(target)
	if s.n <= 2 {
		ops := make([]pitchclass.Operator, 0, tSet.Cardinality())
		tSet.forEachNumber(func(n int) bool {
			ops = append(ops, pitchclass.MustTransposition(n, s.n))
			return true
		})

		return ops, nil
	}

	iSet := s.Invert(0).transpositionIndicesToUnchecked(target)
	ops := make([]pitchclass.Operator, 0, tSet.Cardinality()+iSet.Cardinality())

	tSet.forEachNumber(func(n int) bool {
		ops = append(ops, pitchclass.MustTransposition(n, s.n))
		return true
	})
	iSet.forEachNumber(func(n int) bool {
		ops = append(ops, pitchclass.MustInversion(n, s.n))
		return true
	})

	return ops, nil
}

// MustEquivalenceOperatorsTI is like [Set.EquivalenceOperatorsTI] but panics on error.
func (s Set) MustEquivalenceOperatorsTI(target Set) []pitchclass.Operator {
	ops, err := s.EquivalenceOperatorsTI(target)
	if err != nil {
		panic(err)
	}

	return ops
}

// EquivalenceOperatorsTTO returns all operators f in the full affine TTO group
// of Z_N such that f(s) == target, ordered by [pitchclass.Operator.Compare]
// without duplicates. If no operators match, it returns a non-nil empty slice.
//
// EquivalenceOperatorsTTO checks errors in the following order:
//  1. !s.IsValid() || !target.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != target.Modulus(): [ErrModulusMismatch]
func (s Set) EquivalenceOperatorsTTO(target Set) ([]pitchclass.Operator, error) {
	if err := checkSetPair(s, target); err != nil {
		return nil, err
	}

	ops := []pitchclass.Operator{}
	if s.Cardinality() != target.Cardinality() {
		return ops, nil
	}

	for m := range pitchclass.MustInvertibleMultipliers(s.n) {
		ms := s.Multiply(m)
		tSet := ms.transpositionIndicesToUnchecked(target)

		tSet.forEachNumber(func(n int) bool {
			ops = append(ops, pitchclass.MustNewOperator(m, n, s.n))
			return true
		})
	}

	return ops, nil
}

// MustEquivalenceOperatorsTTO is like [Set.EquivalenceOperatorsTTO] but panics on error.
func (s Set) MustEquivalenceOperatorsTTO(target Set) []pitchclass.Operator {
	ops, err := s.EquivalenceOperatorsTTO(target)
	if err != nil {
		panic(err)
	}

	return ops
}

// EmbeddingOperatorsTI returns all operators f in the T/I group of Z_N such
// that f(s) ⊆ superset, ordered by [pitchclass.Operator.Compare] without
// duplicates. If no operators match, it returns a non-nil empty slice.
//
// EmbeddingOperatorsTI checks errors in the following order:
//  1. !s.IsValid() || !superset.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != superset.Modulus(): [ErrModulusMismatch]
func (s Set) EmbeddingOperatorsTI(superset Set) ([]pitchclass.Operator, error) {
	if err := checkSetPair(s, superset); err != nil {
		return nil, err
	}

	tSet := s.transpositionEmbeddingsIntoUnchecked(superset)
	if s.n <= 2 {
		ops := make([]pitchclass.Operator, 0, tSet.Cardinality())
		tSet.forEachNumber(func(n int) bool {
			ops = append(ops, pitchclass.MustTransposition(n, s.n))
			return true
		})

		return ops, nil
	}

	iSet := s.Invert(0).transpositionEmbeddingsIntoUnchecked(superset)
	ops := make([]pitchclass.Operator, 0, tSet.Cardinality()+iSet.Cardinality())

	tSet.forEachNumber(func(n int) bool {
		ops = append(ops, pitchclass.MustTransposition(n, s.n))
		return true
	})
	iSet.forEachNumber(func(n int) bool {
		ops = append(ops, pitchclass.MustInversion(n, s.n))
		return true
	})

	return ops, nil
}

// MustEmbeddingOperatorsTI is like [Set.EmbeddingOperatorsTI] but panics on error.
func (s Set) MustEmbeddingOperatorsTI(superset Set) []pitchclass.Operator {
	ops, err := s.EmbeddingOperatorsTI(superset)
	if err != nil {
		panic(err)
	}

	return ops
}

// EmbeddingOperatorsTTO returns all operators f in the full affine TTO group
// of Z_N such that f(s) ⊆ superset, ordered by [pitchclass.Operator.Compare]
// without duplicates. If no operators match, it returns a non-nil empty slice.
//
// EmbeddingOperatorsTTO checks errors in the following order:
//  1. !s.IsValid() || !superset.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != superset.Modulus(): [ErrModulusMismatch]
func (s Set) EmbeddingOperatorsTTO(superset Set) ([]pitchclass.Operator, error) {
	if err := checkSetPair(s, superset); err != nil {
		return nil, err
	}

	ops := []pitchclass.Operator{}
	if s.Cardinality() > superset.Cardinality() {
		return ops, nil
	}

	for m := range pitchclass.MustInvertibleMultipliers(s.n) {
		ms := s.Multiply(m)
		tSet := ms.transpositionEmbeddingsIntoUnchecked(superset)

		tSet.forEachNumber(func(n int) bool {
			ops = append(ops, pitchclass.MustNewOperator(m, n, s.n))
			return true
		})
	}

	return ops, nil
}

// MustEmbeddingOperatorsTTO is like [Set.EmbeddingOperatorsTTO] but panics on error.
func (s Set) MustEmbeddingOperatorsTTO(superset Set) []pitchclass.Operator {
	ops, err := s.EmbeddingOperatorsTTO(superset)
	if err != nil {
		panic(err)
	}

	return ops
}

// CommonToneCountTranspose returns the number of common pitch classes
// |T_n(s) ∩ s| between s and its transposition by n.
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) CommonToneCountTranspose(n int) (int, error) {
	if !s.IsValid() {
		return 0, errInvalidValue()
	}

	return s.Transpose(n).intersectionCardinalityUnchecked(s), nil
}

// MustCommonToneCountTranspose is like [Set.CommonToneCountTranspose] but panics on error.
func (s Set) MustCommonToneCountTranspose(n int) int {
	c, err := s.CommonToneCountTranspose(n)
	if err != nil {
		panic(err)
	}

	return c
}

// CommonToneCountInvert returns the number of common pitch classes
// |I_index(s) ∩ s| between s and its inversion by index.
//
// It returns [ErrInvalidValue] if !s.IsValid().
func (s Set) CommonToneCountInvert(index int) (int, error) {
	if !s.IsValid() {
		return 0, errInvalidValue()
	}

	return s.Invert(index).intersectionCardinalityUnchecked(s), nil
}

// MustCommonToneCountInvert is like [Set.CommonToneCountInvert] but panics on error.
func (s Set) MustCommonToneCountInvert(index int) int {
	c, err := s.CommonToneCountInvert(index)
	if err != nil {
		panic(err)
	}

	return c
}

// CommonToneCountTransposeWith returns the number of common pitch classes
// |T_n(s) ∩ other| between T_n(s) and other.
//
// CommonToneCountTransposeWith checks errors in the following order:
//  1. !s.IsValid() || !other.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != other.Modulus(): [ErrModulusMismatch]
func (s Set) CommonToneCountTransposeWith(other Set, n int) (int, error) {
	if err := checkSetPair(s, other); err != nil {
		return 0, err
	}

	return s.Transpose(n).intersectionCardinalityUnchecked(other), nil
}

// MustCommonToneCountTransposeWith is like [Set.CommonToneCountTransposeWith] but panics on error.
func (s Set) MustCommonToneCountTransposeWith(other Set, n int) int {
	c, err := s.CommonToneCountTransposeWith(other, n)
	if err != nil {
		panic(err)
	}

	return c
}

// CommonToneCountInvertWith returns the number of common pitch classes
// |I_index(s) ∩ other| between I_index(s) and other.
//
// CommonToneCountInvertWith checks errors in the following order:
//  1. !s.IsValid() || !other.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != other.Modulus(): [ErrModulusMismatch]
func (s Set) CommonToneCountInvertWith(other Set, index int) (int, error) {
	if err := checkSetPair(s, other); err != nil {
		return 0, err
	}

	return s.Invert(index).intersectionCardinalityUnchecked(other), nil
}

// MustCommonToneCountInvertWith is like [Set.CommonToneCountInvertWith] but panics on error.
func (s Set) MustCommonToneCountInvertWith(other Set, index int) int {
	c, err := s.CommonToneCountInvertWith(other, index)
	if err != nil {
		panic(err)
	}

	return c
}

// CommonToneCountWith returns the number of common pitch classes
// |op(s) ∩ other| between op(s) and other.
//
// CommonToneCountWith checks errors in the following order:
//  1. !s.IsValid() || !other.IsValid() || !op.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != other.Modulus() || s.Modulus() != op.Modulus(): [ErrModulusMismatch]
func (s Set) CommonToneCountWith(other Set, op pitchclass.Operator) (int, error) {
	if !s.IsValid() || !other.IsValid() || !op.IsValid() {
		return 0, errInvalidValue()
	}

	if s.n != other.n {
		return 0, errModulusMismatch(s.n, other.n)
	}

	if s.n != op.Modulus() {
		return 0, errModulusMismatch(s.n, op.Modulus())
	}

	transformed := s.Multiply(op.Multiplier()).Transpose(op.Index())

	return transformed.intersectionCardinalityUnchecked(other), nil
}

// MustCommonToneCountWith is like [Set.CommonToneCountWith] but panics on error.
func (s Set) MustCommonToneCountWith(other Set, op pitchclass.Operator) int {
	c, err := s.CommonToneCountWith(other, op)
	if err != nil {
		panic(err)
	}

	return c
}

func (s Set) transpositionIndicesToUnchecked(target Set) Set {
	sCard := s.Cardinality()
	if sCard != target.Cardinality() {
		return Set{n: s.n}
	}

	if sCard > s.n/2 {
		s = s.Complement()
		target = target.Complement()
	}

	return s.transpositionEmbeddingsIntoUnchecked(target)
}

func (s Set) transpositionEmbeddingsIntoUnchecked(superset Set) Set {
	if s.IsEmpty() || superset.IsUniverse() {
		return universeSet(s.n)
	}

	if s.Cardinality() > superset.Cardinality() {
		return Set{n: s.n}
	}

	if s.n <= inlineModulus {
		var mask uint64

		for n := range s.n {
			if (rotateLeft64(s.lo, n, s.n) &^ superset.lo) == 0 {
				mask |= uint64(1) << n
			}
		}

		return Set{n: s.n, lo: mask}
	}

	elems := s.Numbers()
	x0 := elems[0]
	rest := elems[1:]
	buf := make([]byte, (s.n+7)>>3)

	superset.forEachNumber(func(y int) bool {
		n := y - x0
		if n < 0 {
			n += s.n
		}

		for _, x := range rest {
			z := x + n
			if z >= s.n {
				z -= s.n
			}

			if !superset.hasBit(z) {
				return true
			}
		}

		buf[n>>3] |= 1 << (n & 7)

		return true
	})

	return makeLargeSetFromBytes(s.n, buf)
}
