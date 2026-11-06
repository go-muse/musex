package catalog

import (
	"cmp"
	"fmt"
	"iter"
)

// ZPair represents one of the 23 unordered pairs of Z-correspondent ForteClasses (First.Ordinal() < Second.Ordinal()).
// The zero value ZPair{} is invalid (IsValid() == false).
type ZPair struct {
	index uint8
}

// ZPairs returns a fresh slice of all 23 binary Z-correspondent pairs in canonical (Cardinality, First.Ordinal()) order.
func ZPairs() []ZPair {
	out := make([]ZPair, 23)
	for i := range uint8(23) {
		out[i] = ZPair{index: i + 1}
	}
	return out
}

// AllZPairs returns an iterator over all 23 binary ZPairs in canonical (Cardinality, First.Ordinal()) order.
// It supports early termination when yield returns false.
func AllZPairs() iter.Seq[ZPair] {
	return func(yield func(ZPair) bool) {
		for i := uint8(1); i <= 23; i++ {
			if !yield(ZPair{index: i}) {
				return
			}
		}
	}
}

// ZPairsOfCardinality returns a fresh slice of all binary ZPairs of the given cardinality (0..12).
// For cardinalities in {0, 1, 2, 3, 9, 10, 11, 12} where no Z-pairs exist, it returns a non-nil empty slice []ZPair{} and nil error.
// For cardinality < 0 or cardinality > 12, it returns nil and ErrInvalidCardinality.
func ZPairsOfCardinality(cardinality int) ([]ZPair, error) {
	if cardinality < 0 || cardinality > 12 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCardinality, cardinality)
	}
	start := zPairCardStart[cardinality]
	count := zPairCardCount[cardinality]
	out := make([]ZPair, count)
	for i := range count {
		out[i] = ZPair{index: start + i}
	}
	return out, nil
}

// MustZPairsOfCardinality is like ZPairsOfCardinality but panics on error.
func MustZPairsOfCardinality(cardinality int) []ZPair {
	list, err := ZPairsOfCardinality(cardinality)
	if err != nil {
		panic(err)
	}
	return list
}

// IsValid reports whether zp is one of the 23 valid binary Z-pairs in the catalog.
func (zp ZPair) IsValid() bool {
	return zp.index >= 1 && zp.index <= 23
}

// Cardinality returns the shared cardinality c in {4, 5, 6, 7, 8} of the Z-pair (or 0 if !zp.IsValid()).
func (zp ZPair) Cardinality() int {
	if !zp.IsValid() {
		return 0
	}
	return int(zPairTable[zp.index].cardinality)
}

// First returns the first ForteClass of the Z-pair (with the smaller Ordinal; or ForteClass{} if !zp.IsValid()).
func (zp ZPair) First() ForteClass {
	if !zp.IsValid() {
		return ForteClass{}
	}
	return ForteClass{index: zPairTable[zp.index].firstIdx}
}

// Second returns the second ForteClass of the Z-pair (with the larger Ordinal; or ForteClass{} if !zp.IsValid()).
func (zp ZPair) Second() ForteClass {
	if !zp.IsValid() {
		return ForteClass{}
	}
	return ForteClass{index: zPairTable[zp.index].secondIdx}
}

// IntervalVector returns the shared 6-element interval-class vector of First and Second ([6]int{} if !zp.IsValid()).
func (zp ZPair) IntervalVector() [6]int {
	return zp.First().IntervalVector()
}

// IntervalVectorSlice returns a fresh 6-element slice copy of IntervalVector() (or nil if !zp.IsValid()).
func (zp ZPair) IntervalVectorSlice() []int {
	return zp.First().IntervalVectorSlice()
}

// Equal reports whether zp and other are both valid and represent the same ZPair.
// If either zp or other is invalid, Equal returns false.
func (zp ZPair) Equal(other ZPair) bool {
	return zp.IsValid() && other.IsValid() && zp.index == other.index
}

// Compare orders ZPairs by (Cardinality, First.Ordinal(), Second.Ordinal()).
// An invalid ZPair is ordered before any valid ZPair; two invalid ZPairs return 0.
func (zp ZPair) Compare(other ZPair) int {
	zv, ov := zp.IsValid(), other.IsValid()
	if !zv && !ov {
		return 0
	}
	if !zv {
		return -1
	}
	if !ov {
		return 1
	}
	return cmp.Compare(zp.index, other.index)
}

// String returns the canonical representation "<First>/<Second>" (e.g. "4-Z15/4-Z29", "5-Z17/5-Z37"),
// or "ZPair(invalid)" if !zp.IsValid().
func (zp ZPair) String() string {
	if !zp.IsValid() {
		return "ZPair(invalid)"
	}
	return zPairTable[zp.index].label
}
