package alteration_test

import (
	"math"
	"testing"
)

// Fuzz targets compare the package with the math/big oracle for arbitrary
// int inputs (P-17: integer arguments of any size; only the exact result
// decides the error). A plain go test run executes only the seed corpus.
// The seeds are the edges of the guaranteed range and of int.

// inputInRange reports whether the pair is a value of the guaranteed range;
// other pairs are outside the domain of the operations on values.
func inputInRange(num, den int) bool {
	return den != 0 && inRange(ratOf(num, den))
}

func edgeSeeds() [][2]int {
	return [][2]int{
		{0, 1}, {1, 2}, {-7, 2}, {1, 3},
		{bigM, 1}, {-bigM, 1}, {1, bigQ}, {-1, bigQ},
		{bigM*bigQ - 1, bigQ}, {1 - bigM*bigQ, bigQ},
		{bigM*bigQ + 1, bigQ}, {1, bigQ + 1}, {1, bigQ - 1},
		{math.MinInt, -1}, {math.MinInt, math.MinInt}, {math.MaxInt, math.MaxInt - 1},
		{math.MinInt, -(1 << 62)}, {-(1 << 62), math.MinInt}, {1, 0}, {0, 0},
	}
}

func intSeeds() []int {
	return []int{0, 1, -1, 2, bigM, -bigM, 2 * bigM, bigM*bigQ + 1, math.MaxInt, math.MinInt}
}

func FuzzFromFraction(f *testing.F) {
	for _, s := range edgeSeeds() {
		f.Add(s[0], s[1])
	}

	f.Fuzz(func(t *testing.T, num, den int) {
		checkFallible(t, opFromFraction(num, den))
	})
}

func FuzzFromInt(f *testing.F) {
	for _, n := range intSeeds() {
		f.Add(n)
	}

	f.Fuzz(func(t *testing.T, n int) {
		checkFallible(t, opFromInt(n))
	})
}

func FuzzRaiseByLowerByMul(f *testing.F) {
	for _, s := range edgeSeeds() {
		for _, n := range intSeeds() {
			f.Add(s[0], s[1], n)
		}
	}

	f.Fuzz(func(t *testing.T, num, den, n int) {
		if !inputInRange(num, den) {
			return
		}

		a := frac{num, den}
		checkFallible(t, opRaiseBy(a, n))
		checkFallible(t, opLowerBy(a, n))
		checkFallible(t, opMul(a, n))
	})
}

func FuzzAddSub(f *testing.F) {
	seeds := edgeSeeds()
	for _, a := range seeds {
		for _, b := range seeds {
			f.Add(a[0], a[1], b[0], b[1])
		}
	}

	f.Fuzz(func(t *testing.T, an, ad, bn, bd int) {
		if !inputInRange(an, ad) || !inputInRange(bn, bd) {
			return
		}

		a, b := frac{an, ad}, frac{bn, bd}
		checkFallible(t, opAdd(a, b))
		checkFallible(t, opSub(a, b))
	})
}
