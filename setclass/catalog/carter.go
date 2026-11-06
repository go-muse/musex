package catalog

import (
	"cmp"
	"fmt"
	"iter"
)

// CarterClass represents one of the 135 entries in Elliott Carter's Harmony Book catalog (cardinalities 2, 3, 4, 6, 7).
// The zero value CarterClass{} is invalid (IsValid() == false).
type CarterClass struct {
	index uint8
}

// CarterByNumber returns the CarterClass for the given cardinality (in {2, 3, 4, 6, 7}) and 1-based Carter number.
// It returns ErrInvalidCardinality if cardinality is outside [0, 12],
// ErrNoCarterCatalog if cardinality is in {0, 1, 5, 8, 9, 10, 11, 12},
// or ErrInvalidCarterNumber if number is outside [1, k_max(cardinality)].
func CarterByNumber(cardinality, number int) (CarterClass, error) {
	if cardinality < 0 || cardinality > 12 {
		return CarterClass{}, fmt.Errorf("%w: %d", ErrInvalidCardinality, cardinality)
	}
	start := carterCardStart[cardinality]
	if start == 0 {
		return CarterClass{}, fmt.Errorf("%w: %d", ErrNoCarterCatalog, cardinality)
	}
	if number < 1 || number > int(forteMaxOrdinal[cardinality]) {
		return CarterClass{}, fmt.Errorf("%w: %d for cardinality %d", ErrInvalidCarterNumber, number, cardinality)
	}
	return CarterClass{index: start + uint8(number-1)}, nil
}

// MustCarterByNumber is like CarterByNumber but panics on error.
func MustCarterByNumber(cardinality, number int) CarterClass {
	cc, err := CarterByNumber(cardinality, number)
	if err != nil {
		panic(err)
	}
	return cc
}

// CarterCardinalities returns a fresh slice []int{2, 3, 4, 6, 7} of cardinalities covered by Elliott Carter's catalog.
func CarterCardinalities() []int {
	return []int{2, 3, 4, 6, 7}
}

// HasCarterCardinality reports whether Elliott Carter's catalog is defined for cardinality (true iff c is 2, 3, 4, 6, or 7).
func HasCarterCardinality(cardinality int) bool {
	return cardinality >= 0 && cardinality <= 12 && carterCardStart[cardinality] != 0
}

// CarterClasses returns a fresh slice of all 135 CarterClasses in canonical (Cardinality, Number) order.
func CarterClasses() []CarterClass {
	out := make([]CarterClass, 135)
	for i := range uint8(135) {
		out[i] = CarterClass{index: i + 1}
	}
	return out
}

// AllCarterClasses returns an iterator over all 135 CarterClasses in canonical (Cardinality, Number) order.
// It supports early termination when yield returns false.
func AllCarterClasses() iter.Seq[CarterClass] {
	return func(yield func(CarterClass) bool) {
		for i := uint8(1); i <= 135; i++ {
			if !yield(CarterClass{index: i}) {
				return
			}
		}
	}
}

// CarterClassesOfCardinality returns a fresh slice of all CarterClasses of the given cardinality (in {2, 3, 4, 6, 7})
// in canonical Number order. It returns ErrInvalidCardinality if cardinality < 0 or cardinality > 12,
// or ErrNoCarterCatalog if cardinality is in {0, 1, 5, 8, 9, 10, 11, 12}.
func CarterClassesOfCardinality(cardinality int) ([]CarterClass, error) {
	if cardinality < 0 || cardinality > 12 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCardinality, cardinality)
	}
	start := carterCardStart[cardinality]
	if start == 0 {
		return nil, fmt.Errorf("%w: %d", ErrNoCarterCatalog, cardinality)
	}
	count := forteMaxOrdinal[cardinality]
	out := make([]CarterClass, count)
	for i := range count {
		out[i] = CarterClass{index: start + i}
	}
	return out, nil
}

// MustCarterClassesOfCardinality is like CarterClassesOfCardinality but panics on error.
func MustCarterClassesOfCardinality(cardinality int) []CarterClass {
	list, err := CarterClassesOfCardinality(cardinality)
	if err != nil {
		panic(err)
	}
	return list
}

// IsValid reports whether cc is one of the 135 valid Elliott Carter catalog entries.
func (cc CarterClass) IsValid() bool {
	return cc.index >= 1 && cc.index <= 135
}

// Cardinality returns the cardinality c in {2, 3, 4, 6, 7} (or 0 if !cc.IsValid()).
func (cc CarterClass) Cardinality() int {
	if !cc.IsValid() {
		return 0
	}
	return int(carterTable[cc.index].cardinality)
}

// Number returns the 1-based Carter catalog number in 1..k_max(c) (or 0 if !cc.IsValid()).
func (cc CarterClass) Number() int {
	if !cc.IsValid() {
		return 0
	}
	return int(carterTable[cc.index].number)
}

// ForteClass returns the corresponding ForteClass (or ForteClass{} if !cc.IsValid()).
func (cc CarterClass) ForteClass() ForteClass {
	if !cc.IsValid() {
		return ForteClass{}
	}
	return ForteClass{index: carterTable[cc.index].forteIdx}
}

// Equal reports whether cc and other are both valid and represent the same CarterClass.
// If either cc or other is invalid, Equal returns false.
func (cc CarterClass) Equal(other CarterClass) bool {
	return cc.IsValid() && other.IsValid() && cc.index == other.index
}

// Compare orders CarterClasses by (Cardinality, Number).
// An invalid CarterClass is ordered before any valid CarterClass; two invalid CarterClasses return 0.
func (cc CarterClass) Compare(other CarterClass) int {
	cv, ov := cc.IsValid(), other.IsValid()
	if !cv && !ov {
		return 0
	}
	if !cv {
		return -1
	}
	if !ov {
		return 1
	}
	return cmp.Compare(cc.index, other.index)
}

// String returns the canonical representation "Carter(<c>, <n>): <forte>" (e.g. "Carter(4, 18): 4-Z15"),
// or "CarterClass(invalid)" if !cc.IsValid().
func (cc CarterClass) String() string {
	if !cc.IsValid() {
		return "CarterClass(invalid)"
	}
	return carterTable[cc.index].label
}
