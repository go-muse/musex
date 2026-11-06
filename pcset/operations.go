package pcset

import "math/bits"

// Complement returns the set-theoretic complement Z_N \ s of s in its universe Z_N.
// If s is Set{}, Complement returns Set{}.
func (s Set) Complement() Set {
	if !s.IsValid() {
		return Set{}
	}

	if s.n <= inlineModulus {
		return Set{
			n:  s.n,
			lo: (^s.lo) & universeMask64(s.n),
		}
	}

	byteLen := (s.n + 7) >> 3
	buf := make([]byte, byteLen)

	for i := range byteLen {
		buf[i] = ^s.byteAt(i)
	}

	if rem := s.n & 7; rem != 0 {
		buf[byteLen-1] &= (1 << rem) - 1
	}

	return makeLargeSetFromBytes(s.n, buf)
}

// Union returns the set union s ∪ other.
//
// Union checks errors in the following order:
//  1. !s.IsValid() || !other.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != other.Modulus(): [ErrModulusMismatch]
func (s Set) Union(other Set) (Set, error) {
	if err := checkSetPair(s, other); err != nil {
		return Set{}, err
	}

	if s.n <= inlineModulus {
		return Set{n: s.n, lo: s.lo | other.lo}, nil
	}

	maxLen := max(len(s.hi), len(other.hi))
	buf := make([]byte, maxLen)

	for i := range maxLen {
		buf[i] = s.byteAt(i) | other.byteAt(i)
	}

	return makeLargeSetFromBytes(s.n, buf), nil
}

// MustUnion is like [Set.Union] but panics on error.
func (s Set) MustUnion(other Set) Set {
	res, err := s.Union(other)
	if err != nil {
		panic(err)
	}

	return res
}

// Intersection returns the set intersection s ∩ other.
//
// Intersection checks errors in the following order:
//  1. !s.IsValid() || !other.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != other.Modulus(): [ErrModulusMismatch]
func (s Set) Intersection(other Set) (Set, error) {
	if err := checkSetPair(s, other); err != nil {
		return Set{}, err
	}

	if s.n <= inlineModulus {
		return Set{n: s.n, lo: s.lo & other.lo}, nil
	}

	minLen := min(len(s.hi), len(other.hi))
	buf := make([]byte, minLen)

	for i := range minLen {
		buf[i] = s.hi[i] & other.hi[i]
	}

	return makeLargeSetFromBytes(s.n, buf), nil
}

// MustIntersection is like [Set.Intersection] but panics on error.
func (s Set) MustIntersection(other Set) Set {
	res, err := s.Intersection(other)
	if err != nil {
		panic(err)
	}

	return res
}

// Difference returns the set difference s \ other.
//
// Difference checks errors in the following order:
//  1. !s.IsValid() || !other.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != other.Modulus(): [ErrModulusMismatch]
func (s Set) Difference(other Set) (Set, error) {
	if err := checkSetPair(s, other); err != nil {
		return Set{}, err
	}

	if s.n <= inlineModulus {
		return Set{n: s.n, lo: s.lo &^ other.lo}, nil
	}

	buf := make([]byte, len(s.hi))
	for i := range len(s.hi) {
		buf[i] = s.hi[i] &^ other.byteAt(i)
	}

	return makeLargeSetFromBytes(s.n, buf), nil
}

// MustDifference is like [Set.Difference] but panics on error.
func (s Set) MustDifference(other Set) Set {
	res, err := s.Difference(other)
	if err != nil {
		panic(err)
	}

	return res
}

// SymmetricDifference returns the symmetric difference (s \ other) ∪ (other \ s).
//
// SymmetricDifference checks errors in the following order:
//  1. !s.IsValid() || !other.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != other.Modulus(): [ErrModulusMismatch]
func (s Set) SymmetricDifference(other Set) (Set, error) {
	if err := checkSetPair(s, other); err != nil {
		return Set{}, err
	}

	if s.n <= inlineModulus {
		return Set{n: s.n, lo: s.lo ^ other.lo}, nil
	}

	maxLen := max(len(s.hi), len(other.hi))
	buf := make([]byte, maxLen)

	for i := range maxLen {
		buf[i] = s.byteAt(i) ^ other.byteAt(i)
	}

	return makeLargeSetFromBytes(s.n, buf), nil
}

// MustSymmetricDifference is like [Set.SymmetricDifference] but panics on error.
func (s Set) MustSymmetricDifference(other Set) Set {
	res, err := s.SymmetricDifference(other)
	if err != nil {
		panic(err)
	}

	return res
}

// IsDisjoint reports whether s and other have no elements in common (s ∩ other == ∅).
//
// IsDisjoint checks errors in the following order:
//  1. !s.IsValid() || !other.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != other.Modulus(): [ErrModulusMismatch]
func (s Set) IsDisjoint(other Set) (bool, error) {
	if err := checkSetPair(s, other); err != nil {
		return false, err
	}

	return s.intersectionCardinalityUnchecked(other) == 0, nil
}

// MustIsDisjoint is like [Set.IsDisjoint] but panics on error.
func (s Set) MustIsDisjoint(other Set) bool {
	ok, err := s.IsDisjoint(other)
	if err != nil {
		panic(err)
	}

	return ok
}

// IsSubsetOf reports whether every element of s belongs to other (s ⊆ other).
//
// IsSubsetOf checks errors in the following order:
//  1. !s.IsValid() || !other.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != other.Modulus(): [ErrModulusMismatch]
func (s Set) IsSubsetOf(other Set) (bool, error) {
	if err := checkSetPair(s, other); err != nil {
		return false, err
	}

	return s.isSubsetOfUnchecked(other), nil
}

// MustIsSubsetOf is like [Set.IsSubsetOf] but panics on error.
func (s Set) MustIsSubsetOf(other Set) bool {
	ok, err := s.IsSubsetOf(other)
	if err != nil {
		panic(err)
	}

	return ok
}

// IsProperSubsetOf reports whether s is a strict subset of other (s ⊂ other).
//
// IsProperSubsetOf checks errors in the following order:
//  1. !s.IsValid() || !other.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != other.Modulus(): [ErrModulusMismatch]
func (s Set) IsProperSubsetOf(other Set) (bool, error) {
	if err := checkSetPair(s, other); err != nil {
		return false, err
	}

	return s != other && s.isSubsetOfUnchecked(other), nil
}

// MustIsProperSubsetOf is like [Set.IsProperSubsetOf] but panics on error.
func (s Set) MustIsProperSubsetOf(other Set) bool {
	ok, err := s.IsProperSubsetOf(other)
	if err != nil {
		panic(err)
	}

	return ok
}

// IsSupersetOf reports whether s contains every element of other (s ⊇ other).
//
// IsSupersetOf checks errors in the following order:
//  1. !s.IsValid() || !other.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != other.Modulus(): [ErrModulusMismatch]
func (s Set) IsSupersetOf(other Set) (bool, error) {
	if err := checkSetPair(s, other); err != nil {
		return false, err
	}

	return other.isSubsetOfUnchecked(s), nil
}

// MustIsSupersetOf is like [Set.IsSupersetOf] but panics on error.
func (s Set) MustIsSupersetOf(other Set) bool {
	ok, err := s.IsSupersetOf(other)
	if err != nil {
		panic(err)
	}

	return ok
}

// IsProperSupersetOf reports whether s is a strict superset of other (s ⊃ other).
//
// IsProperSupersetOf checks errors in the following order:
//  1. !s.IsValid() || !other.IsValid(): [ErrInvalidValue]
//  2. s.Modulus() != other.Modulus(): [ErrModulusMismatch]
func (s Set) IsProperSupersetOf(other Set) (bool, error) {
	if err := checkSetPair(s, other); err != nil {
		return false, err
	}

	return s != other && other.isSubsetOfUnchecked(s), nil
}

// MustIsProperSupersetOf is like [Set.IsProperSupersetOf] but panics on error.
func (s Set) MustIsProperSupersetOf(other Set) bool {
	ok, err := s.IsProperSupersetOf(other)
	if err != nil {
		panic(err)
	}

	return ok
}

func (s Set) isSubsetOfUnchecked(other Set) bool {
	if s.n <= inlineModulus {
		return (s.lo &^ other.lo) == 0
	}

	for i := range len(s.hi) {
		if (s.hi[i] &^ other.byteAt(i)) != 0 {
			return false
		}
	}

	return true
}

func (s Set) intersectionCardinalityUnchecked(other Set) int {
	if s.n <= inlineModulus {
		return bits.OnesCount64(s.lo & other.lo)
	}

	minLen := min(len(s.hi), len(other.hi))
	count := 0

	for i := range minLen {
		count += bits.OnesCount8(s.hi[i] & other.hi[i])
	}

	return count
}
