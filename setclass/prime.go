package setclass

import (
	"cmp"
	"fmt"
	"math/big"
	"math/bits"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/pitchclass"
)

func comparePackedNumbers(a, b []int, alg Algorithm) int {
	k := len(a)
	if c := cmp.Compare(a[k-1], b[k-1]); c != 0 {
		return c
	}
	if alg == Rahn {
		for j := k - 2; j >= 1; j-- {
			if c := cmp.Compare(a[j], b[j]); c != 0 {
				return c
			}
		}
		return 0
	}
	for j := 1; j <= k-2; j++ {
		if c := cmp.Compare(a[j], b[j]); c != 0 {
			return c
		}
	}
	return 0
}

func tnPrimeNumbers(s pcset.Set, alg Algorithm) []int {
	n := s.Modulus()
	nf := s.MustNormalFormWith(toPCSetAlgorithm(alg))
	base := nf[0].Number()
	out := make([]int, len(nf))
	for i, c := range nf {
		out[i] = (c.Number() - base + n) % n
	}
	return out
}

func computePrimeForm(s pcset.Set, kind Kind, alg Algorithm) pcset.Set {
	n := s.Modulus()
	k := s.Cardinality()
	if k == 0 || k == n {
		return s
	}
	switch kind {
	case Tn:
		return pcset.MustNew(n, tnPrimeNumbers(s, alg)...)
	case TnIn:
		best := tnPrimeNumbers(s, alg)
		if n > 2 {
			inv := tnPrimeNumbers(s.Invert(0), alg)
			if comparePackedNumbers(inv, best, alg) < 0 {
				best = inv
			}
		}
		return pcset.MustNew(n, best...)
	default: // TTO
		var best []int
		for m := range pitchclass.MustInvertibleMultipliers(n) {
			cand := tnPrimeNumbers(s.Multiply(m), alg)
			if best == nil || comparePackedNumbers(cand, best, alg) < 0 {
				best = cand
			}
		}
		return pcset.MustNew(n, best...)
	}
}

// MinSpan returns the minimal outer span p_{k-1} - p_0 of the prime form of sc,
// or 0 if !sc.IsValid() or sc.Cardinality() <= 1.
func (sc SetClass) MinSpan() int {
	if !sc.IsValid() || sc.prime.Cardinality() <= 1 {
		return 0
	}
	nums := sc.prime.Numbers()
	return nums[len(nums)-1]
}

// RahnForteDiffer reports whether Rahn's and Forte's prime form packing algorithms
// yield distinct prime forms for sc. For an invalid SetClass, it returns false.
func (sc SetClass) RahnForteDiffer() bool {
	if !sc.IsValid() {
		return false
	}
	return !sc.prime.Equal(sc.FortePrimeForm())
}

// PrimeForm returns the canonical Rahn prime form of sc as a [pcset.Set],
// or the zero [pcset.Set] if !sc.IsValid().
func (sc SetClass) PrimeForm() pcset.Set {
	if !sc.IsValid() {
		return pcset.Set{}
	}
	return sc.prime
}

// FortePrimeForm returns Allen Forte's prime form of sc as a [pcset.Set],
// or the zero [pcset.Set] if !sc.IsValid().
func (sc SetClass) FortePrimeForm() pcset.Set {
	if !sc.IsValid() {
		return pcset.Set{}
	}
	return computePrimeForm(sc.prime, sc.kind, Forte)
}

// PrimeFormWith returns the canonical prime form of sc under alg ([Rahn] or [Forte]).
func (sc SetClass) PrimeFormWith(alg Algorithm) (pcset.Set, error) {
	if !sc.IsValid() {
		return pcset.Set{}, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	if !alg.IsValid() {
		return pcset.Set{}, fmt.Errorf("%w: %v", ErrInvalidAlgorithm, alg)
	}
	if alg == Rahn {
		return sc.prime, nil
	}
	return sc.FortePrimeForm(), nil
}

// MustPrimeFormWith is like [SetClass.PrimeFormWith] but panics on error.
func (sc SetClass) MustPrimeFormWith(alg Algorithm) pcset.Set {
	pf, err := sc.PrimeFormWith(alg)
	if err != nil {
		panic(err)
	}
	return pf
}

// PrimeClasses returns a fresh slice of pitch classes in the Rahn prime form of sc,
// or nil if !sc.IsValid().
func (sc SetClass) PrimeClasses() []pitchclass.Class {
	if !sc.IsValid() {
		return nil
	}
	return sc.prime.Classes()
}

// PrimeClassesWith returns a fresh slice of pitch classes in the prime form of sc under alg.
func (sc SetClass) PrimeClassesWith(alg Algorithm) ([]pitchclass.Class, error) {
	pf, err := sc.PrimeFormWith(alg)
	if err != nil {
		return nil, err
	}
	return pf.Classes(), nil
}

// MustPrimeClassesWith is like [SetClass.PrimeClassesWith] but panics on error.
func (sc SetClass) MustPrimeClassesWith(alg Algorithm) []pitchclass.Class {
	classes, err := sc.PrimeClassesWith(alg)
	if err != nil {
		panic(err)
	}
	return classes
}

// PrimeNumbers returns a fresh ascending slice of pitch-class numbers in the Rahn prime form of sc,
// or nil if !sc.IsValid().
func (sc SetClass) PrimeNumbers() []int {
	if !sc.IsValid() {
		return nil
	}
	return sc.prime.Numbers()
}

// PrimeNumbersWith returns a fresh ascending slice of pitch-class numbers in the prime form of sc under alg.
func (sc SetClass) PrimeNumbersWith(alg Algorithm) ([]int, error) {
	pf, err := sc.PrimeFormWith(alg)
	if err != nil {
		return nil, err
	}
	return pf.Numbers(), nil
}

// MustPrimeNumbersWith is like [SetClass.PrimeNumbersWith] but panics on error.
func (sc SetClass) MustPrimeNumbersWith(alg Algorithm) []int {
	nums, err := sc.PrimeNumbersWith(alg)
	if err != nil {
		panic(err)
	}
	return nums
}

// PrimeMaskBig returns a fresh [*big.Int] characteristic bitmask of the Rahn prime form of sc,
// or nil if !sc.IsValid().
func (sc SetClass) PrimeMaskBig() *big.Int {
	if !sc.IsValid() {
		return nil
	}
	return sc.prime.MustBitmaskBig()
}

// PrimeMaskBigWith returns a fresh [*big.Int] characteristic bitmask of the prime form of sc under alg.
func (sc SetClass) PrimeMaskBigWith(alg Algorithm) (*big.Int, error) {
	pf, err := sc.PrimeFormWith(alg)
	if err != nil {
		return nil, err
	}
	return pf.MustBitmaskBig(), nil
}

// MustPrimeMaskBigWith is like [SetClass.PrimeMaskBigWith] but panics on error.
func (sc SetClass) MustPrimeMaskBigWith(alg Algorithm) *big.Int {
	mask, err := sc.PrimeMaskBigWith(alg)
	if err != nil {
		panic(err)
	}
	return mask
}

// PrimeMask returns the 64-bit characteristic bitmask of the Rahn prime form of sc (1 <= N <= 64).
// For N > 64, PrimeMask returns an error wrapping both [ErrOverflow] and [ErrModulusTooLarge].
func (sc SetClass) PrimeMask() (uint64, error) {
	return sc.PrimeMaskWith(Rahn)
}

// MustPrimeMask is like [SetClass.PrimeMask] but panics on error.
func (sc SetClass) MustPrimeMask() uint64 {
	m, err := sc.PrimeMask()
	if err != nil {
		panic(err)
	}
	return m
}

// FortePrimeMask returns the 64-bit characteristic bitmask of Allen Forte's prime form of sc (1 <= N <= 64).
// For N > 64, FortePrimeMask returns an error wrapping both [ErrOverflow] and [ErrModulusTooLarge].
func (sc SetClass) FortePrimeMask() (uint64, error) {
	return sc.PrimeMaskWith(Forte)
}

// MustFortePrimeMask is like [SetClass.FortePrimeMask] but panics on error.
func (sc SetClass) MustFortePrimeMask() uint64 {
	m, err := sc.FortePrimeMask()
	if err != nil {
		panic(err)
	}
	return m
}

// PrimeMaskWith returns the 64-bit characteristic bitmask of the prime form of sc under alg (1 <= N <= 64).
// For N > 64, PrimeMaskWith returns an error wrapping both [ErrOverflow] and [ErrModulusTooLarge].
func (sc SetClass) PrimeMaskWith(alg Algorithm) (uint64, error) {
	pf, err := sc.PrimeFormWith(alg)
	if err != nil {
		return 0, err
	}
	if pf.Modulus() > MaxBitmaskModulus {
		return 0, fmt.Errorf("%w (%w): modulus %d > %d", ErrOverflow, ErrModulusTooLarge, pf.Modulus(), MaxBitmaskModulus)
	}
	return pf.MustBitmask(), nil
}

// MustPrimeMaskWith is like [SetClass.PrimeMaskWith] but panics on error.
func (sc SetClass) MustPrimeMaskWith(alg Algorithm) uint64 {
	m, err := sc.PrimeMaskWith(alg)
	if err != nil {
		panic(err)
	}
	return m
}

func maxTnMask(s pcset.Set) uint64 {
	n := s.Modulus()
	mask := s.MustBitmask()
	fullMask := ^uint64(0)
	if n < 64 {
		fullMask = (uint64(1) << uint(n)) - 1
	}
	var best uint64
	for rem := mask; rem != 0; rem &= rem - 1 {
		y := bits.TrailingZeros64(rem)
		shift := n - 1 - y
		rot := ((mask << uint(shift)) | (mask >> uint(n-shift))) & fullMask
		if rot > best {
			best = rot
		}
	}
	return best
}

// CollinsMask returns the canonical maximum 64-bit integer representation
// max_{A in sc} sum_{x in A} 2^x across the orbit of sc in its equivalence group (1 <= N <= 64).
// For N > 64, CollinsMask returns an error wrapping both [ErrOverflow] and [ErrModulusTooLarge].
func (sc SetClass) CollinsMask() (uint64, error) {
	if !sc.IsValid() {
		return 0, fmt.Errorf("%w: invalid SetClass", ErrInvalidValue)
	}
	n := sc.prime.Modulus()
	if n > MaxBitmaskModulus {
		return 0, fmt.Errorf("%w (%w): modulus %d > %d", ErrOverflow, ErrModulusTooLarge, n, MaxBitmaskModulus)
	}
	if sc.prime.IsEmpty() {
		return 0, nil
	}
	switch sc.kind {
	case Tn:
		return maxTnMask(sc.prime), nil
	case TnIn:
		best := maxTnMask(sc.prime)
		if n > 2 {
			if inv := maxTnMask(sc.prime.Invert(0)); inv > best {
				best = inv
			}
		}
		return best, nil
	default: // TTO
		var best uint64
		for m := range pitchclass.MustInvertibleMultipliers(n) {
			if cand := maxTnMask(sc.prime.Multiply(m)); cand > best {
				best = cand
			}
		}
		return best, nil
	}
}

// MustCollinsMask is like [SetClass.CollinsMask] but panics on error.
func (sc SetClass) MustCollinsMask() uint64 {
	m, err := sc.CollinsMask()
	if err != nil {
		panic(err)
	}
	return m
}
