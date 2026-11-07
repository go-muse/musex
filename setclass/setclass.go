package setclass

import (
	"cmp"
	"fmt"
	"math/big"
	"slices"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"
)

// SetClass represents an equivalence class (orbit) of pitch-class sets in Z_N
// under one of the three canonical transformation groups (Tn, TnIn, or TTO).
//
// SetClass is an immutable comparable value type and may be used as a map key.
// The zero value SetClass{} is invalid (IsValid() == false).
type SetClass struct {
	kind  Kind
	prime pcset.Set
}

func validateModulusAndKind(modulus int, kind Kind) error {
	return validateBoundedModulusKindAlg(modulus, MaxModulus, kind, Rahn)
}

// New constructs the SetClass of s under the equivalence group kind.
// It is a synonym for [FromSet].
func New(s pcset.Set, kind Kind) (SetClass, error) {
	return FromSet(s, kind)
}

// MustNew is like [New] but panics on error.
func MustNew(s pcset.Set, kind Kind) SetClass {
	sc, err := New(s, kind)
	if err != nil {
		panic(err)
	}
	return sc
}

// FromSet constructs the SetClass containing the pitch-class set s under the equivalence group kind.
func FromSet(s pcset.Set, kind Kind) (SetClass, error) {
	if !s.IsValid() {
		return SetClass{}, fmt.Errorf("%w: invalid pcset.Set", ErrInvalidValue)
	}
	if !kind.IsValid() {
		return SetClass{}, fmt.Errorf("%w: %v", ErrInvalidKind, kind)
	}
	n := s.Modulus()
	if kind == TTO && n > MaxTTOModulus {
		return SetClass{}, fmt.Errorf("%w: TTO modulus %d > %d", ErrModulusTooLarge, n, MaxTTOModulus)
	}
	return SetClass{
		kind:  kind,
		prime: computePrimeForm(s, kind, Rahn),
	}, nil
}

// MustFromSet is like [FromSet] but panics on error.
func MustFromSet(s pcset.Set, kind Kind) SetClass {
	sc, err := FromSet(s, kind)
	if err != nil {
		panic(err)
	}
	return sc
}

// NewTn constructs the transpositional (Tn) SetClass containing s.
func NewTn(s pcset.Set) (SetClass, error) {
	return FromSet(s, Tn)
}

// MustNewTn is like [NewTn] but panics on error.
func MustNewTn(s pcset.Set) SetClass {
	sc, err := NewTn(s)
	if err != nil {
		panic(err)
	}
	return sc
}

// NewTnIn constructs the transpositional/inversional (TnIn) SetClass containing s.
func NewTnIn(s pcset.Set) (SetClass, error) {
	return FromSet(s, TnIn)
}

// MustNewTnIn is like [NewTnIn] but panics on error.
func MustNewTnIn(s pcset.Set) SetClass {
	sc, err := NewTnIn(s)
	if err != nil {
		panic(err)
	}
	return sc
}

// NewTTO constructs the twelve-tone operator (TTO) SetClass containing s.
func NewTTO(s pcset.Set) (SetClass, error) {
	return FromSet(s, TTO)
}

// MustNewTTO is like [NewTTO] but panics on error.
func MustNewTTO(s pcset.Set) SetClass {
	sc, err := NewTTO(s)
	if err != nil {
		panic(err)
	}
	return sc
}

// FromClasses constructs the SetClass of modulus and kind from a slice of pitch classes.
// Duplicate classes are deduplicated. If any class is invalid, FromClasses returns [ErrInvalidValue].
// If any class has a modulus different from modulus, FromClasses returns an error wrapping both
// [ErrInvalidPitchClass] and [ErrModulusMismatch].
func FromClasses(modulus int, kind Kind, classes ...pitchclass.Class) (SetClass, error) {
	if err := validateModulusAndKind(modulus, kind); err != nil {
		return SetClass{}, err
	}
	for _, c := range classes {
		if !c.IsValid() {
			return SetClass{}, fmt.Errorf("%w: invalid pitchclass.Class", ErrInvalidValue)
		}
		if c.Modulus() != modulus {
			return SetClass{}, fmt.Errorf("%w (%w): class modulus %d != %d", ErrInvalidPitchClass, ErrModulusMismatch, c.Modulus(), modulus)
		}
	}
	s := pcset.MustFromClassesWithModulus(modulus, classes...)
	return FromSet(s, kind)
}

// MustFromClasses is like [FromClasses] but panics on error.
func MustFromClasses(modulus int, kind Kind, classes ...pitchclass.Class) SetClass {
	sc, err := FromClasses(modulus, kind, classes...)
	if err != nil {
		panic(err)
	}
	return sc
}

// FromNumbers constructs the SetClass of modulus and kind from strict pitch-class numbers in [0, modulus-1].
// Duplicate numbers are deduplicated. Any number outside [0, modulus-1] returns [ErrInvalidPitchClass].
func FromNumbers(modulus int, kind Kind, numbers ...int) (SetClass, error) {
	if err := validateModulusAndKind(modulus, kind); err != nil {
		return SetClass{}, err
	}
	for _, x := range numbers {
		if x < 0 || x >= modulus {
			return SetClass{}, fmt.Errorf("%w: pitch-class number %d out of range [0, %d)", ErrInvalidPitchClass, x, modulus)
		}
	}
	s := pcset.MustNew(modulus, numbers...)
	return FromSet(s, kind)
}

// MustFromNumbers is like [FromNumbers] but panics on error.
func MustFromNumbers(modulus int, kind Kind, numbers ...int) SetClass {
	sc, err := FromNumbers(modulus, kind, numbers...)
	if err != nil {
		panic(err)
	}
	return sc
}

// ReduceNumbers constructs the SetClass of modulus and kind by reducing arbitrary integers modulo modulus.
func ReduceNumbers(modulus int, kind Kind, numbers ...int) (SetClass, error) {
	if err := validateModulusAndKind(modulus, kind); err != nil {
		return SetClass{}, err
	}
	s := pcset.MustReduce(modulus, numbers...)
	return FromSet(s, kind)
}

// MustReduceNumbers is like [ReduceNumbers] but panics on error.
func MustReduceNumbers(modulus int, kind Kind, numbers ...int) SetClass {
	sc, err := ReduceNumbers(modulus, kind, numbers...)
	if err != nil {
		panic(err)
	}
	return sc
}

// FromBitmask constructs the SetClass of modulus and kind from a 64-bit characteristic bitmask (1 <= modulus <= 64).
// For modulus > 64, FromBitmask returns an error wrapping both [ErrOverflow] and [ErrModulusTooLarge].
// If mask has bits set at positions >= modulus, FromBitmask returns [ErrInvalidBitmask].
func FromBitmask(modulus int, kind Kind, mask uint64) (SetClass, error) {
	if modulus > MaxBitmaskModulus {
		return SetClass{}, fmt.Errorf("%w (%w): modulus %d > %d", ErrOverflow, ErrModulusTooLarge, modulus, MaxBitmaskModulus)
	}
	return FromBitmaskBig(modulus, kind, new(big.Int).SetUint64(mask))
}

// MustFromBitmask is like [FromBitmask] but panics on error.
func MustFromBitmask(modulus int, kind Kind, mask uint64) SetClass {
	sc, err := FromBitmask(modulus, kind, mask)
	if err != nil {
		panic(err)
	}
	return sc
}

// FromBitmaskBig constructs the SetClass of modulus and kind from an arbitrary-precision characteristic bitmask.
// If mask is nil, FromBitmaskBig returns [ErrInvalidValue].
// If mask < 0 or mask >= 2^modulus, FromBitmaskBig returns [ErrInvalidBitmask].
func FromBitmaskBig(modulus int, kind Kind, mask *big.Int) (SetClass, error) {
	if err := validateModulusAndKind(modulus, kind); err != nil {
		return SetClass{}, err
	}
	if mask == nil {
		return SetClass{}, fmt.Errorf("%w: nil *big.Int bitmask", ErrInvalidValue)
	}
	if mask.Sign() < 0 || mask.BitLen() > modulus {
		return SetClass{}, fmt.Errorf("%w: bitmask out of range [0, 2^%d)", ErrInvalidBitmask, modulus)
	}
	s := pcset.MustFromBitmaskBig(mask, modulus)
	return FromSet(s, kind)
}

// MustFromBitmaskBig is like [FromBitmaskBig] but panics on error.
func MustFromBitmaskBig(modulus int, kind Kind, mask *big.Int) SetClass {
	sc, err := FromBitmaskBig(modulus, kind, mask)
	if err != nil {
		panic(err)
	}
	return sc
}

// Empty returns the empty SetClass (cardinality 0) of modulus and kind.
func Empty(modulus int, kind Kind) (SetClass, error) {
	if err := validateModulusAndKind(modulus, kind); err != nil {
		return SetClass{}, err
	}
	return SetClass{
		kind:  kind,
		prime: pcset.MustEmpty(modulus),
	}, nil
}

// MustEmpty is like [Empty] but panics on error.
func MustEmpty(modulus int, kind Kind) SetClass {
	sc, err := Empty(modulus, kind)
	if err != nil {
		panic(err)
	}
	return sc
}

// Universe returns the aggregate SetClass (cardinality N) of modulus and kind.
func Universe(modulus int, kind Kind) (SetClass, error) {
	if err := validateModulusAndKind(modulus, kind); err != nil {
		return SetClass{}, err
	}
	return SetClass{
		kind:  kind,
		prime: pcset.MustUniverse(modulus),
	}, nil
}

// MustUniverse is like [Universe] but panics on error.
func MustUniverse(modulus int, kind Kind) SetClass {
	sc, err := Universe(modulus, kind)
	if err != nil {
		panic(err)
	}
	return sc
}

// ChromaticCluster returns the SetClass of the contiguous chromatic cluster {0, 1, ..., cardinality-1}
// of the given modulus and kind.
func ChromaticCluster(modulus, cardinality int, kind Kind) (SetClass, error) {
	if err := validateEnumerationParams(modulus, MaxModulus, cardinality, 0, modulus, 0, kind, Rahn); err != nil {
		return SetClass{}, err
	}
	nums := make([]int, cardinality)
	for i := range nums {
		nums[i] = i
	}
	return FromNumbers(modulus, kind, nums...)
}

// MustChromaticCluster is like [ChromaticCluster] but panics on error.
func MustChromaticCluster(modulus, cardinality int, kind Kind) SetClass {
	sc, err := ChromaticCluster(modulus, cardinality, kind)
	if err != nil {
		panic(err)
	}
	return sc
}

// IsValid reports whether sc is a valid SetClass (not the zero value SetClass{}).
func (sc SetClass) IsValid() bool {
	return sc.kind.IsValid() && sc.prime.IsValid()
}

// Modulus returns the pitch-class universe size N of sc, or 0 if !sc.IsValid().
func (sc SetClass) Modulus() int {
	if !sc.IsValid() {
		return 0
	}
	return sc.prime.Modulus()
}

// Kind returns the equivalence group kind (Tn, TnIn, or TTO) of sc, or Kind(0) if !sc.IsValid().
func (sc SetClass) Kind() Kind {
	if !sc.IsValid() {
		return 0
	}
	return sc.kind
}

// Cardinality returns the number of pitch classes k in each member set of sc, or 0 if !sc.IsValid().
func (sc SetClass) Cardinality() int {
	if !sc.IsValid() {
		return 0
	}
	return sc.prime.Cardinality()
}

// IsEmpty reports whether sc is a valid empty set class (cardinality 0).
func (sc SetClass) IsEmpty() bool {
	return sc.IsValid() && sc.prime.IsEmpty()
}

// IsUniverse reports whether sc is a valid aggregate set class (cardinality N).
func (sc SetClass) IsUniverse() bool {
	return sc.IsValid() && sc.prime.IsUniverse()
}

// Equal reports whether sc and other are both valid and represent the exact same SetClass
// (same Modulus, same Kind, and same canonical prime form).
// If either sc or other is invalid, Equal returns false.
func (sc SetClass) Equal(other SetClass) bool {
	return sc.IsValid() && other.IsValid() && sc == other
}

// Compare orders SetClasses by (Modulus, Kind, Cardinality, PrimeNumbers) using Rahn's prime form.
// Any invalid SetClass is ordered before all valid SetClasses; two invalid SetClasses return 0.
func (sc SetClass) Compare(other SetClass) int {
	sv, ov := sc.IsValid(), other.IsValid()
	if !sv && !ov {
		return 0
	}
	if !sv {
		return -1
	}
	if !ov {
		return 1
	}
	if c := cmp.Compare(sc.prime.Modulus(), other.prime.Modulus()); c != 0 {
		return c
	}
	if c := sc.kind.Compare(other.kind); c != 0 {
		return c
	}
	if c := cmp.Compare(sc.prime.Cardinality(), other.prime.Cardinality()); c != 0 {
		return c
	}
	return slices.Compare(sc.prime.Numbers(), other.prime.Numbers())
}

// CompareWith orders sc and other by (Modulus, Kind, Cardinality, PrimeNumbersWith(alg)).
// If either sc or other is invalid, CompareWith returns [ErrInvalidValue].
// If alg is invalid, CompareWith returns [ErrInvalidAlgorithm].
func (sc SetClass) CompareWith(other SetClass, alg Algorithm) (int, error) {
	if !sc.IsValid() || !other.IsValid() {
		return 0, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	if !alg.IsValid() {
		return 0, fmt.Errorf("%w: %v", ErrInvalidAlgorithm, alg)
	}
	if alg == Rahn || sc.prime.Modulus() != other.prime.Modulus() || sc.kind != other.kind || sc.prime.Cardinality() != other.prime.Cardinality() {
		return sc.Compare(other), nil
	}
	p1 := computePrimeForm(sc.prime, sc.kind, Forte).Numbers()
	p2 := computePrimeForm(other.prime, other.kind, Forte).Numbers()
	return slices.Compare(p1, p2), nil
}

// MustCompareWith is like [SetClass.CompareWith] but panics on error.
func (sc SetClass) MustCompareWith(other SetClass, alg Algorithm) int {
	c, err := sc.CompareWith(other, alg)
	if err != nil {
		panic(err)
	}
	return c
}
