package pcset

import (
	"cmp"
	"fmt"

	"github.com/go-muse/musex/pitchclass"
)

// Algorithm specifies the packing convention used to select the normal form
// of a pitch-class set among rotations with minimal outer span.
type Algorithm uint8

const (
	// Rahn selects the rotation most tightly packed to the right (minimizing
	// d_{k-1}, d_{k-2}, ..., d_1, then the starting class r_0), following
	// Rahn (1980) and Straus (2016). Rahn is the zero value and default.
	Rahn Algorithm = iota

	// Forte selects the rotation most tightly packed to the left from the
	// first element among rotations with minimal outer span (minimizing
	// d_{k-1}, d_1, d_2, ..., d_{k-2}, then the starting class r_0),
	// following Forte (1973).
	Forte
)

// IsValid reports whether a is a defined normal-form algorithm ([Rahn] or [Forte]).
func (a Algorithm) IsValid() bool {
	return a == Rahn || a == Forte
}

// String returns the name of the algorithm ("Rahn" or "Forte"), or
// "Algorithm(<n>)" for an invalid value.
func (a Algorithm) String() string {
	switch a {
	case Rahn:
		return "Rahn"
	case Forte:
		return "Forte"
	default:
		return fmt.Sprintf("Algorithm(%d)", uint8(a))
	}
}

// Rotations returns all k cyclic rotations of the ascending pitch-class list
// of s as newly allocated slices. For an empty valid set it returns a non-nil
// empty slice [][][pitchclass.Class]{}; for Set{} it returns nil.
func (s Set) Rotations() [][]pitchclass.Class {
	if !s.IsValid() {
		return nil
	}

	classes := s.Classes()
	k := len(classes)

	if k == 0 {
		return [][]pitchclass.Class{}
	}

	res := make([][]pitchclass.Class, k)
	for i := range k {
		rot := make([]pitchclass.Class, k)
		copy(rot, classes[i:])
		copy(rot[k-i:], classes[:i])
		res[i] = rot
	}

	return res
}

// NormalForm returns the normal form of s using the default [Rahn] algorithm.
// For an empty valid set it returns a non-nil empty slice; for Set{} it
// returns nil.
func (s Set) NormalForm() []pitchclass.Class {
	if !s.IsValid() {
		return nil
	}

	return s.normalFormUnchecked(Rahn)
}

// NormalFormWith returns the normal form of s using the specified packing
// algorithm alg ([Rahn] or [Forte]).
//
// NormalFormWith checks errors in the following order:
//  1. !s.IsValid(): [ErrInvalidValue]
//  2. !alg.IsValid(): [ErrInvalidAlgorithm]
func (s Set) NormalFormWith(alg Algorithm) ([]pitchclass.Class, error) {
	if !s.IsValid() {
		return nil, errInvalidValue()
	}

	if !alg.IsValid() {
		return nil, errInvalidAlgorithm(alg)
	}

	return s.normalFormUnchecked(alg), nil
}

// MustNormalFormWith is like [Set.NormalFormWith] but panics on error.
func (s Set) MustNormalFormWith(alg Algorithm) []pitchclass.Class {
	nf, err := s.NormalFormWith(alg)
	if err != nil {
		panic(err)
	}

	return nf
}

func (s Set) normalFormUnchecked(alg Algorithm) []pitchclass.Class {
	nums := s.normalFormNumbersUnchecked(alg)
	res := make([]pitchclass.Class, len(nums))

	for i, x := range nums {
		res[i] = pitchclass.MustNew(x, s.n)
	}

	return res
}

func (s Set) normalFormNumbersUnchecked(alg Algorithm) []int {
	nums := s.Numbers()
	k := len(nums)

	if k <= 1 {
		return nums
	}

	bestI := 0
	for i := 1; i < k; i++ {
		if compareRotation(nums, s.n, i, bestI, alg) < 0 {
			bestI = i
		}
	}

	if bestI == 0 {
		return nums
	}

	rot := make([]int, k)
	copy(rot, nums[bestI:])
	copy(rot[k-bestI:], nums[:bestI])

	return rot
}

func rotationSpan(nums []int, n, start, j int) int {
	k := len(nums)
	idx := start + j

	if idx < k {
		return nums[idx] - nums[start]
	}

	return nums[idx-k] + n - nums[start]
}

func compareRotation(nums []int, n, a, b int, alg Algorithm) int {
	k := len(nums)
	if c := cmp.Compare(rotationSpan(nums, n, a, k-1), rotationSpan(nums, n, b, k-1)); c != 0 {
		return c
	}

	if alg == Rahn {
		for j := k - 2; j >= 1; j-- {
			if c := cmp.Compare(rotationSpan(nums, n, a, j), rotationSpan(nums, n, b, j)); c != 0 {
				return c
			}
		}
	} else {
		for j := 1; j <= k-2; j++ {
			if c := cmp.Compare(rotationSpan(nums, n, a, j), rotationSpan(nums, n, b, j)); c != 0 {
				return c
			}
		}
	}

	return cmp.Compare(nums[a], nums[b])
}
