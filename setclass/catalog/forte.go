package catalog

import (
	"cmp"
	"fmt"
	"iter"
)

// ForteClass represents one of the 224 canonical Tn/In set classes in 12-tone equal temperament (cardinalities 0..12).
// The zero value ForteClass{} is invalid (IsValid() == false).
type ForteClass struct {
	index uint8
}

// ParseForteClass parses a canonical Forte set-class label (e.g. "0-1", "3-11", "4-Z15", "12-1").
// It rejects A/B suffixes (e.g. "3-11A") and returns ErrInvalidNotation, ErrInvalidCardinality,
// or ErrInvalidOrdinal on any syntax, Z-flag, or range violation.
func ParseForteClass(s string) (ForteClass, error) {
	fcIdx, _, err := parseForteComponents(s, false)
	if err != nil {
		return ForteClass{}, err
	}
	return ForteClass{index: fcIdx}, nil
}

// MustParseForteClass is like ParseForteClass but panics if s is invalid.
func MustParseForteClass(s string) ForteClass {
	fc, err := ParseForteClass(s)
	if err != nil {
		panic(err)
	}
	return fc
}

// ResolveForteClass resolves either a canonical ForteClass label (224 labels) or a canonical TnType label (352 labels,
// including "A"/"B" suffixes on chiral classes such as "3-11A" or "3-11B") to its parent ForteClass.
// It returns ErrInvalidNotation, ErrInvalidCardinality, ErrInvalidOrdinal, or ErrInvalidSuffix if s is neither
// a valid ForteClass label nor a valid TnType label.
func ResolveForteClass(s string) (ForteClass, error) {
	fcIdx, suffix, err := parseForteComponents(s, true)
	if err != nil {
		return ForteClass{}, err
	}
	if suffix != SuffixNone && forteTable[fcIdx].inSym > 0 {
		return ForteClass{}, fmt.Errorf("%w: %w: %q", ErrInvalidNotation, ErrInvalidSuffix, s)
	}
	return ForteClass{index: fcIdx}, nil
}

// MustResolveForteClass is like ResolveForteClass but panics if s is invalid.
func MustResolveForteClass(s string) ForteClass {
	fc, err := ResolveForteClass(s)
	if err != nil {
		panic(err)
	}
	return fc
}

// ForteClassByNumber returns the ForteClass with the given cardinality (0..12) and 1-based Forte ordinal (1..k_max(c)).
// The Z-status is determined automatically by the catalog.
// It returns ErrInvalidCardinality if cardinality is outside [0, 12], or ErrInvalidOrdinal if ordinal is out of range.
func ForteClassByNumber(cardinality, ordinal int) (ForteClass, error) {
	if cardinality < 0 || cardinality > 12 {
		return ForteClass{}, fmt.Errorf("%w: %d", ErrInvalidCardinality, cardinality)
	}
	if ordinal < 1 || ordinal > int(forteMaxOrdinal[cardinality]) {
		return ForteClass{}, fmt.Errorf("%w: %d for cardinality %d", ErrInvalidOrdinal, ordinal, cardinality)
	}
	return ForteClass{index: forteCardStart[cardinality] + uint8(ordinal-1)}, nil
}

// MustForteClassByNumber is like ForteClassByNumber but panics on error.
func MustForteClassByNumber(cardinality, ordinal int) ForteClass {
	fc, err := ForteClassByNumber(cardinality, ordinal)
	if err != nil {
		panic(err)
	}
	return fc
}

// ForteClassByPrimeMask returns the ForteClass whose canonical Rahn prime form bitmask equals mask.
// It returns ErrNotPrimeForm if mask is not one of the 224 canonical Rahn prime form masks.
func ForteClassByPrimeMask(mask uint64) (ForteClass, error) {
	if mask > 0xFFF {
		return ForteClass{}, fmt.Errorf("%w: mask %d exceeds 12 bits", ErrNotPrimeForm, mask)
	}
	idx := forteByRahnMask[mask]
	if idx == 0 {
		return ForteClass{}, fmt.Errorf("%w: mask %d is not a Rahn prime form", ErrNotPrimeForm, mask)
	}
	return ForteClass{index: idx}, nil
}

// MustForteClassByPrimeMask is like ForteClassByPrimeMask but panics on error.
func MustForteClassByPrimeMask(mask uint64) ForteClass {
	fc, err := ForteClassByPrimeMask(mask)
	if err != nil {
		panic(err)
	}
	return fc
}

// ForteClassByPrimeMaskWith returns the ForteClass whose canonical prime form bitmask under alg equals mask.
// It returns ErrInvalidAlgorithm if alg is invalid, or ErrNotPrimeForm if mask is not a canonical prime form under alg.
func ForteClassByPrimeMaskWith(mask uint64, alg Algorithm) (ForteClass, error) {
	if !alg.IsValid() {
		return ForteClass{}, fmt.Errorf("%w: %d", ErrInvalidAlgorithm, alg)
	}
	if alg == Rahn {
		return ForteClassByPrimeMask(mask)
	}
	if mask > 0xFFF {
		return ForteClass{}, fmt.Errorf("%w: mask %d exceeds 12 bits", ErrNotPrimeForm, mask)
	}
	idx := forteByForteMask[mask]
	if idx == 0 {
		return ForteClass{}, fmt.Errorf("%w: mask %d is not a Forte prime form", ErrNotPrimeForm, mask)
	}
	return ForteClass{index: idx}, nil
}

// MustForteClassByPrimeMaskWith is like ForteClassByPrimeMaskWith but panics on error.
func MustForteClassByPrimeMaskWith(mask uint64, alg Algorithm) ForteClass {
	fc, err := ForteClassByPrimeMaskWith(mask, alg)
	if err != nil {
		panic(err)
	}
	return fc
}

// ForteClassByAnyPrimeMask returns the ForteClass whose canonical Rahn OR Forte prime form bitmask equals mask
// (accepting all 230 distinct Rahn/Forte prime form masks across the 224 set classes).
// It returns ErrNotPrimeForm if mask is neither a canonical Rahn nor a canonical Forte prime form mask.
func ForteClassByAnyPrimeMask(mask uint64) (ForteClass, error) {
	if mask > 0xFFF {
		return ForteClass{}, fmt.Errorf("%w: mask %d exceeds 12 bits", ErrNotPrimeForm, mask)
	}
	idx := forteByAnyMask[mask]
	if idx == 0 {
		return ForteClass{}, fmt.Errorf("%w: mask %d is neither a Rahn nor a Forte prime form", ErrNotPrimeForm, mask)
	}
	return ForteClass{index: idx}, nil
}

// MustForteClassByAnyPrimeMask is like ForteClassByAnyPrimeMask but panics on error.
func MustForteClassByAnyPrimeMask(mask uint64) ForteClass {
	fc, err := ForteClassByAnyPrimeMask(mask)
	if err != nil {
		panic(err)
	}
	return fc
}

// ForteClassByPrimeForm returns the ForteClass whose canonical Rahn prime form equals numbers.
// Any zero-length slice len(numbers) == 0 (both []int{} and nil) represents the empty prime form (mask 0) and resolves to "0-1".
// The input slice is not mutated or retained. It returns ErrNotPrimeForm if numbers is not a canonical Rahn prime form.
func ForteClassByPrimeForm(numbers []int) (ForteClass, error) {
	mask, ok := sliceToPrimeMask(numbers)
	if !ok {
		return ForteClass{}, fmt.Errorf("%w: %v", ErrNotPrimeForm, numbers)
	}
	return ForteClassByPrimeMask(mask)
}

// MustForteClassByPrimeForm is like ForteClassByPrimeForm but panics on error.
func MustForteClassByPrimeForm(numbers []int) ForteClass {
	fc, err := ForteClassByPrimeForm(numbers)
	if err != nil {
		panic(err)
	}
	return fc
}

// ForteClassByPrimeFormWith returns the ForteClass whose canonical prime form under alg equals numbers.
// Any zero-length slice len(numbers) == 0 (both []int{} and nil) represents the empty prime form (mask 0) and resolves to "0-1".
// It returns ErrInvalidAlgorithm if alg is invalid, or ErrNotPrimeForm if numbers is not a canonical prime form under alg.
func ForteClassByPrimeFormWith(numbers []int, alg Algorithm) (ForteClass, error) {
	if !alg.IsValid() {
		return ForteClass{}, fmt.Errorf("%w: %d", ErrInvalidAlgorithm, alg)
	}
	mask, ok := sliceToPrimeMask(numbers)
	if !ok {
		return ForteClass{}, fmt.Errorf("%w: %v", ErrNotPrimeForm, numbers)
	}
	return ForteClassByPrimeMaskWith(mask, alg)
}

// MustForteClassByPrimeFormWith is like ForteClassByPrimeFormWith but panics on error.
func MustForteClassByPrimeFormWith(numbers []int, alg Algorithm) ForteClass {
	fc, err := ForteClassByPrimeFormWith(numbers, alg)
	if err != nil {
		panic(err)
	}
	return fc
}

// ForteClassByAnyPrimeForm returns the ForteClass whose canonical Rahn OR Forte prime form equals numbers
// (accepting all 230 distinct Rahn/Forte prime forms; any zero-length slice len(numbers) == 0 resolves to "0-1").
// It returns ErrNotPrimeForm if numbers is neither a canonical Rahn nor a canonical Forte prime form.
func ForteClassByAnyPrimeForm(numbers []int) (ForteClass, error) {
	mask, ok := sliceToPrimeMask(numbers)
	if !ok {
		return ForteClass{}, fmt.Errorf("%w: %v", ErrNotPrimeForm, numbers)
	}
	return ForteClassByAnyPrimeMask(mask)
}

// MustForteClassByAnyPrimeForm is like ForteClassByAnyPrimeForm but panics on error.
func MustForteClassByAnyPrimeForm(numbers []int) ForteClass {
	fc, err := ForteClassByAnyPrimeForm(numbers)
	if err != nil {
		panic(err)
	}
	return fc
}

// ForteClasses returns a fresh slice of all 224 ForteClasses in canonical (Cardinality, Ordinal) order.
func ForteClasses() []ForteClass {
	out := make([]ForteClass, 224)
	for i := range uint8(224) {
		out[i] = ForteClass{index: i + 1}
	}
	return out
}

// AllForteClasses returns an iterator over all 224 ForteClasses in canonical (Cardinality, Ordinal) order.
// It supports early termination when yield returns false.
func AllForteClasses() iter.Seq[ForteClass] {
	return func(yield func(ForteClass) bool) {
		for i := uint8(1); i <= 224; i++ {
			if !yield(ForteClass{index: i}) {
				return
			}
		}
	}
}

// ForteClassesOfCardinality returns a fresh slice of all ForteClasses of the given cardinality (0..12)
// in canonical Ordinal order, or ErrInvalidCardinality if cardinality < 0 or cardinality > 12.
func ForteClassesOfCardinality(cardinality int) ([]ForteClass, error) {
	if cardinality < 0 || cardinality > 12 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCardinality, cardinality)
	}
	start := forteCardStart[cardinality]
	count := forteMaxOrdinal[cardinality]
	out := make([]ForteClass, count)
	for i := range count {
		out[i] = ForteClass{index: start + i}
	}
	return out, nil
}

// MustForteClassesOfCardinality is like ForteClassesOfCardinality but panics on error.
func MustForteClassesOfCardinality(cardinality int) []ForteClass {
	list, err := ForteClassesOfCardinality(cardinality)
	if err != nil {
		panic(err)
	}
	return list
}

// IsValid reports whether fc is one of the 224 valid catalog ForteClasses.
func (fc ForteClass) IsValid() bool {
	return fc.index >= 1 && fc.index <= 224
}

// Cardinality returns the set class cardinality c in 0..12 (or 0 if !fc.IsValid()).
func (fc ForteClass) Cardinality() int {
	if !fc.IsValid() {
		return 0
	}
	return int(forteTable[fc.index].cardinality)
}

// Ordinal returns the 1-based Forte ordinal number d in 1..k_max(c) (or 0 if !fc.IsValid()).
func (fc ForteClass) Ordinal() int {
	if !fc.IsValid() {
		return 0
	}
	return int(forteTable[fc.index].ordinal)
}

// IsZ reports whether fc carries the "Z" flag in Forte's catalog (true for the 46 Z-classes; false if !fc.IsValid()).
func (fc ForteClass) IsZ() bool {
	if !fc.IsValid() {
		return false
	}
	return forteTable[fc.index].isZ
}

// String returns the canonical Forte label (e.g. "0-1", "3-11", "4-Z15", "12-1"),
// or "ForteClass(invalid)" if !fc.IsValid().
func (fc ForteClass) String() string {
	if !fc.IsValid() {
		return "ForteClass(invalid)"
	}
	return forteTable[fc.index].label
}

// Equal reports whether fc and other are both valid and represent the same ForteClass.
// If either fc or other is invalid, Equal returns false (whereas the built-in == operator returns true for two zero values).
func (fc ForteClass) Equal(other ForteClass) bool {
	return fc.IsValid() && other.IsValid() && fc.index == other.index
}

// Compare orders ForteClasses by (Cardinality, Ordinal).
// An invalid ForteClass is ordered before any valid ForteClass; two invalid ForteClasses return 0.
func (fc ForteClass) Compare(other ForteClass) int {
	fv, ov := fc.IsValid(), other.IsValid()
	if !fv && !ov {
		return 0
	}
	if !fv {
		return -1
	}
	if !ov {
		return 1
	}
	return cmp.Compare(fc.index, other.index)
}

// PrimeForm returns a fresh copy of the canonical Rahn prime form as a slice of pitch-class numbers (0..11).
// For "0-1" it returns a non-nil empty slice []int{}; if !fc.IsValid() it returns nil.
func (fc ForteClass) PrimeForm() []int {
	if !fc.IsValid() {
		return nil
	}
	rec := &forteTable[fc.index]
	return maskToSlice(rec.rahnMask, rec.cardinality)
}

// FortePrimeForm returns a fresh copy of the canonical Forte (1973) prime form as a slice of pitch-class numbers (0..11).
// For "0-1" it returns a non-nil empty slice []int{}; if !fc.IsValid() it returns nil.
func (fc ForteClass) FortePrimeForm() []int {
	if !fc.IsValid() {
		return nil
	}
	rec := &forteTable[fc.index]
	return maskToSlice(rec.forteMask, rec.cardinality)
}

// PrimeFormWith returns a fresh copy of the canonical prime form under alg.
// It returns ErrInvalidValue if !fc.IsValid(), or ErrInvalidAlgorithm if !alg.IsValid().
func (fc ForteClass) PrimeFormWith(alg Algorithm) ([]int, error) {
	if !fc.IsValid() {
		return nil, fmt.Errorf("%w: ForteClass is invalid", ErrInvalidValue)
	}
	if !alg.IsValid() {
		return nil, fmt.Errorf("%w: %d", ErrInvalidAlgorithm, alg)
	}
	if alg == Rahn {
		return fc.PrimeForm(), nil
	}
	return fc.FortePrimeForm(), nil
}

// MustPrimeFormWith is like PrimeFormWith but panics on error.
func (fc ForteClass) MustPrimeFormWith(alg Algorithm) []int {
	pf, err := fc.PrimeFormWith(alg)
	if err != nil {
		panic(err)
	}
	return pf
}

// PrimeMask returns the 12-bit Rahn prime form bitmask (in 0..4095; 0 for both "0-1" and invalid ForteClass).
func (fc ForteClass) PrimeMask() uint64 {
	if !fc.IsValid() {
		return 0
	}
	return uint64(forteTable[fc.index].rahnMask)
}

// FortePrimeMask returns the 12-bit Forte (1973) prime form bitmask (in 0..4095; 0 for both "0-1" and invalid ForteClass).
func (fc ForteClass) FortePrimeMask() uint64 {
	if !fc.IsValid() {
		return 0
	}
	return uint64(forteTable[fc.index].forteMask)
}

// PrimeMaskWith returns the 12-bit prime form bitmask under alg.
// It returns ErrInvalidValue if !fc.IsValid(), or ErrInvalidAlgorithm if !alg.IsValid().
func (fc ForteClass) PrimeMaskWith(alg Algorithm) (uint64, error) {
	if !fc.IsValid() {
		return 0, fmt.Errorf("%w: ForteClass is invalid", ErrInvalidValue)
	}
	if !alg.IsValid() {
		return 0, fmt.Errorf("%w: %d", ErrInvalidAlgorithm, alg)
	}
	if alg == Rahn {
		return fc.PrimeMask(), nil
	}
	return fc.FortePrimeMask(), nil
}

// MustPrimeMaskWith is like PrimeMaskWith but panics on error.
func (fc ForteClass) MustPrimeMaskWith(alg Algorithm) uint64 {
	m, err := fc.PrimeMaskWith(alg)
	if err != nil {
		panic(err)
	}
	return m
}

// RahnForteDiffer reports whether Rahn and Forte prime forms differ for fc
// (true only for the 6 classes: "5-20", "6-Z29", "6-31", "7-Z18", "7-20", "8-26"; false if !fc.IsValid()).
func (fc ForteClass) RahnForteDiffer() bool {
	if !fc.IsValid() {
		return false
	}
	rec := &forteTable[fc.index]
	return rec.rahnMask != rec.forteMask
}

// IntervalVector returns the 6-element interval-class vector [ic1, ic2, ic3, ic4, ic5, ic6] ([6]int{} if !fc.IsValid()).
func (fc ForteClass) IntervalVector() [6]int {
	if !fc.IsValid() {
		return [6]int{}
	}
	return forteTable[fc.index].icv
}

// IntervalVectorSlice returns a fresh 6-element slice copy of IntervalVector() (or nil if !fc.IsValid()).
func (fc ForteClass) IntervalVectorSlice() []int {
	if !fc.IsValid() {
		return nil
	}
	icv := forteTable[fc.index].icv
	out := make([]int, 6)
	copy(out, icv[:])
	return out
}

// InvarianceVector returns the 8-element Morris invariance vector
// [h(Tn), h(TnI), h(TnM5), h(TnM7), c(Tn), c(TnI), c(TnM5), c(TnM7)] ([8]int{} if !fc.IsValid()).
func (fc ForteClass) InvarianceVector() [8]int {
	if !fc.IsValid() {
		return [8]int{}
	}
	return forteTable[fc.index].morris
}

// InvarianceVectorSlice returns a fresh 8-element slice copy of InvarianceVector() (or nil if !fc.IsValid()).
func (fc ForteClass) InvarianceVectorSlice() []int {
	if !fc.IsValid() {
		return nil
	}
	iv := forteTable[fc.index].morris
	out := make([]int, 8)
	copy(out, iv[:])
	return out
}

// TnSymmetry returns the order of transpositional symmetry s_Tn in {1, 2, 3, 4, 6, 12} (or 0 if !fc.IsValid()).
func (fc ForteClass) TnSymmetry() int {
	if !fc.IsValid() {
		return 0
	}
	return int(forteTable[fc.index].tnSym)
}

// InSymmetry returns the number of inversional symmetry axes s_In in {0, 1, 2, 3, 4, 6, 12} (or 0 if !fc.IsValid()).
func (fc ForteClass) InSymmetry() int {
	if !fc.IsValid() {
		return 0
	}
	return int(forteTable[fc.index].inSym)
}

// DegreeOfSymmetry returns the total Tn/In symmetry order d_sym = s_Tn + s_In in 1..24 (or 0 if !fc.IsValid()).
func (fc ForteClass) DegreeOfSymmetry() int {
	if !fc.IsValid() {
		return 0
	}
	rec := &forteTable[fc.index]
	return int(rec.tnSym + rec.inSym)
}

// OrbitSize returns the number of distinct pitch-class sets in the Tn/In orbit (24 / d_sym, or 1 for "0-1"; 0 if !fc.IsValid()).
func (fc ForteClass) OrbitSize() int {
	if !fc.IsValid() {
		return 0
	}
	rec := &forteTable[fc.index]
	return 24 / int(rec.tnSym+rec.inSym)
}

// IsInversionallySymmetric reports whether fc is invariant under inversion (s_In > 0; false if !fc.IsValid()).
func (fc ForteClass) IsInversionallySymmetric() bool {
	if !fc.IsValid() {
		return false
	}
	return forteTable[fc.index].inSym > 0
}

// IsChiral reports whether fc splits into two distinct directed Tn-types "A" and "B" (s_In == 0; false if !fc.IsValid()).
func (fc ForteClass) IsChiral() bool {
	if !fc.IsValid() {
		return false
	}
	return forteTable[fc.index].inSym == 0
}

// TnTypeCount returns 1 if fc is inversionally symmetric, 2 if fc is chiral, and 0 if !fc.IsValid().
func (fc ForteClass) TnTypeCount() int {
	if !fc.IsValid() {
		return 0
	}
	if forteTable[fc.index].inSym > 0 {
		return 1
	}
	return 2
}

// MinSpan returns the minimum chromatic span of the prime form in 0..11 (0 for "0-1", "1-1", or !fc.IsValid()).
func (fc ForteClass) MinSpan() int {
	if !fc.IsValid() {
		return 0
	}
	return int(forteTable[fc.index].minSpan)
}

// Complement returns the complementary ForteClass of cardinality 12 - c (or ForteClass{} if !fc.IsValid()).
func (fc ForteClass) Complement() ForteClass {
	if !fc.IsValid() {
		return ForteClass{}
	}
	return ForteClass{index: forteTable[fc.index].complementIdx}
}

// IsSelfComplementary reports whether fc.Complement() == fc (true for the 20 non-Z hexachords; false if !fc.IsValid()).
func (fc ForteClass) IsSelfComplementary() bool {
	if !fc.IsValid() {
		return false
	}
	return forteTable[fc.index].complementIdx == fc.index
}

// ZPartner returns the Z-correspondent ForteClass and true if fc has a Z-partner (the 46 Z-classes in 23 binary Z-pairs),
// or (ForteClass{}, false) otherwise (including all 178 non-Z classes and invalid ForteClass).
func (fc ForteClass) ZPartner() (ForteClass, bool) {
	if !fc.IsValid() {
		return ForteClass{}, false
	}
	zp := forteTable[fc.index].zPartnerIdx
	if zp == 0 {
		return ForteClass{}, false
	}
	return ForteClass{index: zp}, true
}

// ZPair returns the binary ZPair containing fc and true if fc is one of the 46 Z-classes,
// or (ZPair{}, false) otherwise.
func (fc ForteClass) ZPair() (ZPair, bool) {
	if !fc.IsValid() {
		return ZPair{}, false
	}
	zp := forteTable[fc.index].zPairIdx
	if zp == 0 {
		return ZPair{}, false
	}
	return ZPair{index: zp}, true
}

// M5Partner returns the ForteClass obtained by applying the M5 (circle-of-fifths, *5 mod 12) operation
// (or ForteClass{} if !fc.IsValid()).
func (fc ForteClass) M5Partner() ForteClass {
	if !fc.IsValid() {
		return ForteClass{}
	}
	return ForteClass{index: forteTable[fc.index].m5PartnerIdx}
}

// IsM5Invariant reports whether fc.M5Partner() == fc (true for 92 classes, false for 132 classes in 66 pairs and for invalid fc).
func (fc ForteClass) IsM5Invariant() bool {
	if !fc.IsValid() {
		return false
	}
	return forteTable[fc.index].m5PartnerIdx == fc.index
}

// TnTypes returns a fresh slice of the directed Tn-types belonging to fc:
// length 1 ([TnTypeA]) for inversionally symmetric classes, length 2 ([TnTypeA, TnTypeB]) for chiral classes,
// or nil if !fc.IsValid().
func (fc ForteClass) TnTypes() []TnType {
	if !fc.IsValid() {
		return nil
	}
	rec := &forteTable[fc.index]
	if rec.tnTypeBIdx == 0 {
		return []TnType{{index: rec.tnTypeAIdx}}
	}
	return []TnType{{index: rec.tnTypeAIdx}, {index: rec.tnTypeBIdx}}
}

// TnTypeA returns the primary directed TnType of fc (without suffix if symmetric, with SuffixA if chiral;
// or TnType{} if !fc.IsValid()).
func (fc ForteClass) TnTypeA() TnType {
	if !fc.IsValid() {
		return TnType{}
	}
	return TnType{index: forteTable[fc.index].tnTypeAIdx}
}

// TnTypeB returns the secondary directed TnType (with SuffixB) and true if fc is chiral,
// or (TnType{}, false) if fc is inversionally symmetric or !fc.IsValid().
func (fc ForteClass) TnTypeB() (TnType, bool) {
	if !fc.IsValid() {
		return TnType{}, false
	}
	b := forteTable[fc.index].tnTypeBIdx
	if b == 0 {
		return TnType{}, false
	}
	return TnType{index: b}, true
}

// HasCarter reports whether fc has a number in Elliott Carter's catalog (true iff c in {2, 3, 4, 6, 7} and fc.IsValid()).
func (fc ForteClass) HasCarter() bool {
	if !fc.IsValid() {
		return false
	}
	return forteTable[fc.index].carterIdx != 0
}

// CarterNumber returns the 1-based Carter catalog number and true if fc.HasCarter(), or (0, false) otherwise.
func (fc ForteClass) CarterNumber() (int, bool) {
	if !fc.IsValid() {
		return 0, false
	}
	num := forteTable[fc.index].carterNum
	if num == 0 {
		return 0, false
	}
	return int(num), true
}

// Carter returns the corresponding CarterClass and true if fc.HasCarter(), or (CarterClass{}, false) otherwise.
func (fc ForteClass) Carter() (CarterClass, bool) {
	if !fc.IsValid() {
		return CarterClass{}, false
	}
	idx := forteTable[fc.index].carterIdx
	if idx == 0 {
		return CarterClass{}, false
	}
	return CarterClass{index: idx}, true
}
