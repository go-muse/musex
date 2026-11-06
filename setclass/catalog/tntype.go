package catalog

import (
	"cmp"
	"fmt"
	"iter"
)

// TnType represents one of the 352 canonical directed Tn-types in 12-tone equal temperament (cardinalities 0..12).
// The zero value TnType{} is invalid (IsValid() == false).
type TnType struct {
	index uint16
}

// ParseTnType parses a canonical directed Tn-type label (e.g. "0-1", "3-1", "3-11A", "3-11B", "4-Z15A", "6-Z29").
// Chiral set classes require an "A" or "B" suffix (e.g. "3-11" without suffix is rejected with ErrInvalidNotation);
// inversionally symmetric set classes forbid an "A" or "B" suffix (e.g. "3-1A" is rejected with ErrInvalidNotation).
func ParseTnType(s string) (TnType, error) {
	fcIdx, suffix, err := parseForteComponents(s, true)
	if err != nil {
		return TnType{}, err
	}
	rec := &forteTable[fcIdx]
	if rec.inSym > 0 {
		if suffix != SuffixNone {
			return TnType{}, fmt.Errorf("%w: %w: %q", ErrInvalidNotation, ErrInvalidSuffix, s)
		}
		return TnType{index: rec.tnTypeAIdx}, nil
	}
	if suffix == SuffixA {
		return TnType{index: rec.tnTypeAIdx}, nil
	}
	if suffix == SuffixB {
		return TnType{index: rec.tnTypeBIdx}, nil
	}
	return TnType{}, fmt.Errorf("%w: %w: %q", ErrInvalidNotation, ErrInvalidSuffix, s)
}

// MustParseTnType is like ParseTnType but panics if s is invalid.
func MustParseTnType(s string) TnType {
	t, err := ParseTnType(s)
	if err != nil {
		panic(err)
	}
	return t
}

// TnTypeByNumber returns the TnType with the given cardinality (0..12), 1-based Forte ordinal, and orientation suffix.
// It returns ErrInvalidCardinality, ErrInvalidOrdinal, or ErrInvalidSuffix on invalid or incompatible arguments.
func TnTypeByNumber(cardinality, ordinal int, suffix Suffix) (TnType, error) {
	fc, err := ForteClassByNumber(cardinality, ordinal)
	if err != nil {
		return TnType{}, err
	}
	rec := &forteTable[fc.index]
	if rec.inSym > 0 {
		if suffix != SuffixNone {
			return TnType{}, fmt.Errorf("%w: symmetric class %s does not accept suffix %v", ErrInvalidSuffix, rec.label, suffix)
		}
		return TnType{index: rec.tnTypeAIdx}, nil
	}
	switch suffix {
	case SuffixA:
		return TnType{index: rec.tnTypeAIdx}, nil
	case SuffixB:
		return TnType{index: rec.tnTypeBIdx}, nil
	default:
		return TnType{}, fmt.Errorf("%w: chiral class %s requires SuffixA or SuffixB, got %v", ErrInvalidSuffix, rec.label, suffix)
	}
}

// MustTnTypeByNumber is like TnTypeByNumber but panics on error.
func MustTnTypeByNumber(cardinality, ordinal int, suffix Suffix) TnType {
	t, err := TnTypeByNumber(cardinality, ordinal, suffix)
	if err != nil {
		panic(err)
	}
	return t
}

// TnTypeByPrimeMask returns the TnType whose canonical Rahn Tn-prime form bitmask equals mask.
// It returns ErrNotPrimeForm if mask is not one of the 352 canonical Rahn Tn-prime form masks.
func TnTypeByPrimeMask(mask uint64) (TnType, error) {
	if mask > 0xFFF {
		return TnType{}, fmt.Errorf("%w: mask %d exceeds 12 bits", ErrNotPrimeForm, mask)
	}
	idx := tnByRahnMask[mask]
	if idx == 0 {
		return TnType{}, fmt.Errorf("%w: mask %d is not a Rahn Tn-prime form", ErrNotPrimeForm, mask)
	}
	return TnType{index: idx}, nil
}

// MustTnTypeByPrimeMask is like TnTypeByPrimeMask but panics on error.
func MustTnTypeByPrimeMask(mask uint64) TnType {
	t, err := TnTypeByPrimeMask(mask)
	if err != nil {
		panic(err)
	}
	return t
}

// TnTypeByPrimeMaskWith returns the TnType whose canonical Tn-prime form bitmask under alg equals mask.
// It returns ErrInvalidAlgorithm if alg is invalid, or ErrNotPrimeForm if mask is not a canonical Tn-prime form under alg.
func TnTypeByPrimeMaskWith(mask uint64, alg Algorithm) (TnType, error) {
	if !alg.IsValid() {
		return TnType{}, fmt.Errorf("%w: %d", ErrInvalidAlgorithm, alg)
	}
	if alg == Rahn {
		return TnTypeByPrimeMask(mask)
	}
	if mask > 0xFFF {
		return TnType{}, fmt.Errorf("%w: mask %d exceeds 12 bits", ErrNotPrimeForm, mask)
	}
	idx := tnByForteMask[mask]
	if idx == 0 {
		return TnType{}, fmt.Errorf("%w: mask %d is not a Forte Tn-prime form", ErrNotPrimeForm, mask)
	}
	return TnType{index: idx}, nil
}

// MustTnTypeByPrimeMaskWith is like TnTypeByPrimeMaskWith but panics on error.
func MustTnTypeByPrimeMaskWith(mask uint64, alg Algorithm) TnType {
	t, err := TnTypeByPrimeMaskWith(mask, alg)
	if err != nil {
		panic(err)
	}
	return t
}

// TnTypeByAnyPrimeMask returns the TnType whose canonical Rahn OR Forte Tn-prime form bitmask equals mask
// (accepting all 369 distinct Rahn/Forte Tn-prime form masks across the 352 Tn-types).
// It returns ErrNotPrimeForm if mask is neither a canonical Rahn nor a canonical Forte Tn-prime form mask.
func TnTypeByAnyPrimeMask(mask uint64) (TnType, error) {
	if mask > 0xFFF {
		return TnType{}, fmt.Errorf("%w: mask %d exceeds 12 bits", ErrNotPrimeForm, mask)
	}
	idx := tnByAnyMask[mask]
	if idx == 0 {
		return TnType{}, fmt.Errorf("%w: mask %d is neither a Rahn nor a Forte Tn-prime form", ErrNotPrimeForm, mask)
	}
	return TnType{index: idx}, nil
}

// MustTnTypeByAnyPrimeMask is like TnTypeByAnyPrimeMask but panics on error.
func MustTnTypeByAnyPrimeMask(mask uint64) TnType {
	t, err := TnTypeByAnyPrimeMask(mask)
	if err != nil {
		panic(err)
	}
	return t
}

// TnTypeByPrimeForm returns the TnType whose canonical Rahn Tn-prime form equals numbers.
// Any zero-length slice len(numbers) == 0 (both []int{} and nil) represents the empty prime form (mask 0) and resolves to "0-1".
// The input slice is not mutated or retained. It returns ErrNotPrimeForm if numbers is not a canonical Rahn Tn-prime form.
func TnTypeByPrimeForm(numbers []int) (TnType, error) {
	mask, ok := sliceToPrimeMask(numbers)
	if !ok {
		return TnType{}, fmt.Errorf("%w: %v", ErrNotPrimeForm, numbers)
	}
	return TnTypeByPrimeMask(mask)
}

// MustTnTypeByPrimeForm is like TnTypeByPrimeForm but panics on error.
func MustTnTypeByPrimeForm(numbers []int) TnType {
	t, err := TnTypeByPrimeForm(numbers)
	if err != nil {
		panic(err)
	}
	return t
}

// TnTypeByPrimeFormWith returns the TnType whose canonical Tn-prime form under alg equals numbers.
// Any zero-length slice len(numbers) == 0 (both []int{} and nil) represents the empty prime form (mask 0) and resolves to "0-1".
// It returns ErrInvalidAlgorithm if alg is invalid, or ErrNotPrimeForm if numbers is not a canonical Tn-prime form under alg.
func TnTypeByPrimeFormWith(numbers []int, alg Algorithm) (TnType, error) {
	if !alg.IsValid() {
		return TnType{}, fmt.Errorf("%w: %d", ErrInvalidAlgorithm, alg)
	}
	mask, ok := sliceToPrimeMask(numbers)
	if !ok {
		return TnType{}, fmt.Errorf("%w: %v", ErrNotPrimeForm, numbers)
	}
	return TnTypeByPrimeMaskWith(mask, alg)
}

// MustTnTypeByPrimeFormWith is like TnTypeByPrimeFormWith but panics on error.
func MustTnTypeByPrimeFormWith(numbers []int, alg Algorithm) TnType {
	t, err := TnTypeByPrimeFormWith(numbers, alg)
	if err != nil {
		panic(err)
	}
	return t
}

// TnTypeByAnyPrimeForm returns the TnType whose canonical Rahn OR Forte Tn-prime form equals numbers
// (accepting all 369 distinct Rahn/Forte Tn-prime forms; any zero-length slice len(numbers) == 0 resolves to "0-1").
// It returns ErrNotPrimeForm if numbers is neither a canonical Rahn nor a canonical Forte Tn-prime form.
func TnTypeByAnyPrimeForm(numbers []int) (TnType, error) {
	mask, ok := sliceToPrimeMask(numbers)
	if !ok {
		return TnType{}, fmt.Errorf("%w: %v", ErrNotPrimeForm, numbers)
	}
	return TnTypeByAnyPrimeMask(mask)
}

// MustTnTypeByAnyPrimeForm is like TnTypeByAnyPrimeForm but panics on error.
func MustTnTypeByAnyPrimeForm(numbers []int) TnType {
	t, err := TnTypeByAnyPrimeForm(numbers)
	if err != nil {
		panic(err)
	}
	return t
}

// TnTypes returns a fresh slice of all 352 TnTypes in canonical (Cardinality, Ordinal, Suffix) order.
func TnTypes() []TnType {
	out := make([]TnType, 352)
	for i := range uint16(352) {
		out[i] = TnType{index: i + 1}
	}
	return out
}

// AllTnTypes returns an iterator over all 352 TnTypes in canonical (Cardinality, Ordinal, Suffix) order.
// It supports early termination when yield returns false.
func AllTnTypes() iter.Seq[TnType] {
	return func(yield func(TnType) bool) {
		for i := uint16(1); i <= 352; i++ {
			if !yield(TnType{index: i}) {
				return
			}
		}
	}
}

// TnTypesOfCardinality returns a fresh slice of all TnTypes of the given cardinality (0..12)
// in canonical (Ordinal, Suffix) order, or ErrInvalidCardinality if cardinality < 0 or cardinality > 12.
func TnTypesOfCardinality(cardinality int) ([]TnType, error) {
	if cardinality < 0 || cardinality > 12 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCardinality, cardinality)
	}
	start := tnCardStart[cardinality]
	count := uint16(tnCardCount[cardinality])
	out := make([]TnType, count)
	for i := range count {
		out[i] = TnType{index: start + i}
	}
	return out, nil
}

// MustTnTypesOfCardinality is like TnTypesOfCardinality but panics on error.
func MustTnTypesOfCardinality(cardinality int) []TnType {
	list, err := TnTypesOfCardinality(cardinality)
	if err != nil {
		panic(err)
	}
	return list
}

// IsValid reports whether t is one of the 352 valid catalog TnTypes.
func (t TnType) IsValid() bool {
	return t.index >= 1 && t.index <= 352
}

// ForteClass returns the parent ForteClass of t (or ForteClass{} if !t.IsValid()).
func (t TnType) ForteClass() ForteClass {
	if !t.IsValid() {
		return ForteClass{}
	}
	return ForteClass{index: tnTable[t.index].forteIdx}
}

// Cardinality returns the cardinality c in 0..12 (or 0 if !t.IsValid()).
func (t TnType) Cardinality() int {
	return t.ForteClass().Cardinality()
}

// Ordinal returns the 1-based Forte ordinal d in 1..k_max(c) (or 0 if !t.IsValid()).
func (t TnType) Ordinal() int {
	return t.ForteClass().Ordinal()
}

// IsZ reports whether the parent ForteClass carries the "Z" flag (false if !t.IsValid()).
func (t TnType) IsZ() bool {
	return t.ForteClass().IsZ()
}

// Suffix returns SuffixNone for symmetric Tn-types, SuffixA or SuffixB for chiral Tn-types (or SuffixNone if !t.IsValid()).
func (t TnType) Suffix() Suffix {
	if !t.IsValid() {
		return SuffixNone
	}
	return tnTable[t.index].suffix
}

// HasSuffix reports whether t has orientation suffix SuffixA or SuffixB (true iff t.IsChiral(); false if !t.IsValid()).
func (t TnType) HasSuffix() bool {
	if !t.IsValid() {
		return false
	}
	return tnTable[t.index].suffix != SuffixNone
}

// String returns the canonical Tn-type label (e.g. "0-1", "3-1", "3-11A", "3-11B", "4-Z15A", "6-Z29"),
// or "TnType(invalid)" if !t.IsValid().
func (t TnType) String() string {
	if !t.IsValid() {
		return "TnType(invalid)"
	}
	return tnTable[t.index].label
}

// Equal reports whether t and other are both valid and represent the same TnType.
// If either t or other is invalid, Equal returns false (whereas the built-in == operator returns true for two zero values).
func (t TnType) Equal(other TnType) bool {
	return t.IsValid() && other.IsValid() && t.index == other.index
}

// Compare orders TnTypes by (Cardinality, Ordinal, Suffix) where SuffixNone < SuffixA < SuffixB.
// An invalid TnType is ordered before any valid TnType; two invalid TnTypes return 0.
func (t TnType) Compare(other TnType) int {
	tv, ov := t.IsValid(), other.IsValid()
	if !tv && !ov {
		return 0
	}
	if !tv {
		return -1
	}
	if !ov {
		return 1
	}
	return cmp.Compare(t.index, other.index)
}

// PrimeForm returns a fresh copy of the canonical Rahn Tn-prime form as a slice of pitch-class numbers (0..11).
// For "0-1" it returns a non-nil empty slice []int{}; if !t.IsValid() it returns nil.
func (t TnType) PrimeForm() []int {
	if !t.IsValid() {
		return nil
	}
	rec := &tnTable[t.index]
	return maskToSlice(rec.rahnMask, forteTable[rec.forteIdx].cardinality)
}

// FortePrimeForm returns a fresh copy of the canonical Forte (1973) Tn-prime form as a slice of pitch-class numbers (0..11).
// For "0-1" it returns a non-nil empty slice []int{}; if !t.IsValid() it returns nil.
func (t TnType) FortePrimeForm() []int {
	if !t.IsValid() {
		return nil
	}
	rec := &tnTable[t.index]
	return maskToSlice(rec.forteMask, forteTable[rec.forteIdx].cardinality)
}

// PrimeFormWith returns a fresh copy of the canonical Tn-prime form under alg.
// It returns ErrInvalidValue if !t.IsValid(), or ErrInvalidAlgorithm if !alg.IsValid().
func (t TnType) PrimeFormWith(alg Algorithm) ([]int, error) {
	if !t.IsValid() {
		return nil, fmt.Errorf("%w: TnType is invalid", ErrInvalidValue)
	}
	if !alg.IsValid() {
		return nil, fmt.Errorf("%w: %d", ErrInvalidAlgorithm, alg)
	}
	if alg == Rahn {
		return t.PrimeForm(), nil
	}
	return t.FortePrimeForm(), nil
}

// MustPrimeFormWith is like PrimeFormWith but panics on error.
func (t TnType) MustPrimeFormWith(alg Algorithm) []int {
	pf, err := t.PrimeFormWith(alg)
	if err != nil {
		panic(err)
	}
	return pf
}

// PrimeMask returns the 12-bit Rahn Tn-prime form bitmask (in 0..4095; 0 for both "0-1" and invalid TnType).
func (t TnType) PrimeMask() uint64 {
	if !t.IsValid() {
		return 0
	}
	return uint64(tnTable[t.index].rahnMask)
}

// FortePrimeMask returns the 12-bit Forte (1973) Tn-prime form bitmask (in 0..4095; 0 for both "0-1" and invalid TnType).
func (t TnType) FortePrimeMask() uint64 {
	if !t.IsValid() {
		return 0
	}
	return uint64(tnTable[t.index].forteMask)
}

// PrimeMaskWith returns the 12-bit Tn-prime form bitmask under alg.
// It returns ErrInvalidValue if !t.IsValid(), or ErrInvalidAlgorithm if !alg.IsValid().
func (t TnType) PrimeMaskWith(alg Algorithm) (uint64, error) {
	if !t.IsValid() {
		return 0, fmt.Errorf("%w: TnType is invalid", ErrInvalidValue)
	}
	if !alg.IsValid() {
		return 0, fmt.Errorf("%w: %d", ErrInvalidAlgorithm, alg)
	}
	if alg == Rahn {
		return t.PrimeMask(), nil
	}
	return t.FortePrimeMask(), nil
}

// MustPrimeMaskWith is like PrimeMaskWith but panics on error.
func (t TnType) MustPrimeMaskWith(alg Algorithm) uint64 {
	m, err := t.PrimeMaskWith(alg)
	if err != nil {
		panic(err)
	}
	return m
}

// RahnForteDiffer reports whether Rahn and Forte Tn-prime forms differ for t
// (true only for the 17 Tn-types where Rahn and Forte Tn-prime forms diverge; false if !t.IsValid()).
func (t TnType) RahnForteDiffer() bool {
	if !t.IsValid() {
		return false
	}
	rec := &tnTable[t.index]
	return rec.rahnMask != rec.forteMask
}

// IntervalVector returns the 6-element interval-class vector [ic1..ic6] of t ([6]int{} if !t.IsValid()).
func (t TnType) IntervalVector() [6]int {
	return t.ForteClass().IntervalVector()
}

// IntervalVectorSlice returns a fresh 6-element slice copy of IntervalVector() (or nil if !t.IsValid()).
func (t TnType) IntervalVectorSlice() []int {
	return t.ForteClass().IntervalVectorSlice()
}

// InvarianceVector returns the 8-element Morris invariance vector of the parent ForteClass ([8]int{} if !t.IsValid()).
func (t TnType) InvarianceVector() [8]int {
	return t.ForteClass().InvarianceVector()
}

// InvarianceVectorSlice returns a fresh 8-element slice copy of InvarianceVector() (or nil if !t.IsValid()).
func (t TnType) InvarianceVectorSlice() []int {
	return t.ForteClass().InvarianceVectorSlice()
}

// TnSymmetry returns the order of transpositional symmetry s_Tn in {1, 2, 3, 4, 6, 12} (or 0 if !t.IsValid()).
func (t TnType) TnSymmetry() int {
	return t.ForteClass().TnSymmetry()
}

// InSymmetry returns the number of inversional symmetry axes s_In in {0, 1, 2, 3, 4, 6, 12} (or 0 if !t.IsValid()).
func (t TnType) InSymmetry() int {
	return t.ForteClass().InSymmetry()
}

// OrbitSize returns the number of distinct pitch-class sets in the directed Tn orbit:
// 12 / s_Tn for c in 1..12, 1 for "0-1", or 0 if !t.IsValid().
func (t TnType) OrbitSize() int {
	if !t.IsValid() {
		return 0
	}
	return 12 / int(forteTable[tnTable[t.index].forteIdx].tnSym)
}

// IsInversionallySymmetric reports whether t is invariant under inversion (s_In > 0; false if !t.IsValid()).
func (t TnType) IsInversionallySymmetric() bool {
	return t.ForteClass().IsInversionallySymmetric()
}

// IsChiral reports whether t is one of the 256 chiral Tn-types with SuffixA or SuffixB (s_In == 0; false if !t.IsValid()).
func (t TnType) IsChiral() bool {
	return t.ForteClass().IsChiral()
}

// MinSpan returns the minimum chromatic span of the prime form in 0..11 (0 for "0-1", "1-1", or !t.IsValid()).
func (t TnType) MinSpan() int {
	return t.ForteClass().MinSpan()
}

// Inversion returns the inversionally related TnType:
// t itself if t is inversionally symmetric, the opposite suffix ("A" <-> "B") of the same ForteClass if t is chiral,
// or TnType{} if !t.IsValid().
func (t TnType) Inversion() TnType {
	if !t.IsValid() {
		return TnType{}
	}
	return TnType{index: tnTable[t.index].inversionIdx}
}

// Complement returns the directed Tn-type of the set-theoretic complement (12 - c)
// as defined in forte-tn-types.csv (or TnType{} if !t.IsValid()).
func (t TnType) Complement() TnType {
	if !t.IsValid() {
		return TnType{}
	}
	return TnType{index: tnTable[t.index].complementIdx}
}

// IsSelfComplementary reports whether t.Complement() == t under directed Tn-equivalence without inversion.
// True only for the 8 Tn-types: "6-1", "6-7", "6-8", "6-14A", "6-14B", "6-20", "6-32", "6-35" (false if !t.IsValid()).
func (t TnType) IsSelfComplementary() bool {
	if !t.IsValid() {
		return false
	}
	return tnTable[t.index].complementIdx == t.index
}

// ZPartner returns the Z-correspondent ForteClass of the parent ForteClass and true if it exists,
// or (ForteClass{}, false) otherwise.
func (t TnType) ZPartner() (ForteClass, bool) {
	return t.ForteClass().ZPartner()
}

// M5Partner returns the M5-partner ForteClass of the parent ForteClass (or ForteClass{} if !t.IsValid()).
func (t TnType) M5Partner() ForteClass {
	return t.ForteClass().M5Partner()
}

// IsM5Invariant reports whether t.ForteClass().IsM5Invariant() (false if !t.IsValid()).
func (t TnType) IsM5Invariant() bool {
	return t.ForteClass().IsM5Invariant()
}

// HasCarter reports whether the parent ForteClass has a number in Elliott Carter's catalog (false if !t.IsValid()).
func (t TnType) HasCarter() bool {
	return t.ForteClass().HasCarter()
}

// CarterNumber returns the 1-based Carter catalog number of the parent ForteClass and true if t.HasCarter(),
// or (0, false) otherwise.
func (t TnType) CarterNumber() (int, bool) {
	return t.ForteClass().CarterNumber()
}

// Carter returns the CarterClass of the parent ForteClass and true if t.HasCarter(),
// or (CarterClass{}, false) otherwise.
func (t TnType) Carter() (CarterClass, bool) {
	return t.ForteClass().Carter()
}
