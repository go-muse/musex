package system

import (
	"math"
	"math/bits"
)

// floorDiv returns ⌊a/b⌋ for b > 0.
func floorDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && a < 0 {
		q--
	}

	return q
}

// ceilDiv returns ⌈a/b⌉ for b > 0.
func ceilDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && a > 0 {
		q++
	}

	return q
}

// floorMod returns a mod b in [0, b) for b > 0.
func floorMod(a, b int64) int64 {
	r := a % b
	if r < 0 {
		r += b
	}

	return r
}

// gcd returns the greatest common divisor of |a| and |b|; gcd(0, 0) = 0.
func gcd(a, b int64) int64 {
	if a < 0 {
		a = -a
	}

	if b < 0 {
		b = -b
	}

	for b != 0 {
		a, b = b, a%b
	}

	return a
}

// lcm returns the least common multiple of a > 0 and b > 0. The caller keeps
// the result within int64.
func lcm(a, b int64) int64 {
	return a / gcd(a, b) * b
}

// modInverse returns x in [0, m) with a·x ≡ 1 (mod m) for m >= 1 and
// gcd(a, m) = 1. For m = 1 it returns 0.
func modInverse(a, m int64) int64 {
	if m == 1 {
		return 0
	}

	a = floorMod(a, m)

	oldR, r := a, m
	oldS, s := int64(1), int64(0)

	for r != 0 {
		q := oldR / r
		oldR, r = r, oldR-q*r
		oldS, s = s, oldS-q*s
	}

	return floorMod(oldS, m)
}

// solveCongruence solves a·x ≡ b (mod m) for a >= 1 and m >= 1. When a
// solution exists, it returns x0 in [0, e) and the period e of the solutions
// x ≡ x0 (mod e). The caller keeps m <= 2^32 and e <= 2^16 so that every
// product fits in int64.
func solveCongruence(a, b, m int64) (x0, e int64, ok bool) {
	g := gcd(a, m)
	b = floorMod(b, m)

	if b%g != 0 {
		return 0, 0, false
	}

	e = m / g
	inv := modInverse((a/g)%e, e)
	x0 = floorMod((b/g)%e*inv, e)

	return x0, e, true
}

// mulAdd returns a·b + c when the exact result fits in int64.
func mulAdd(a, b, c int64) (int64, bool) {
	hi, lo := bits.Mul64(uint64(a), uint64(b))
	// Signed correction of the high word of the unsigned product.
	if a < 0 {
		hi -= uint64(b)
	}

	if b < 0 {
		hi -= uint64(a)
	}

	var ext uint64
	if c < 0 {
		ext = math.MaxUint64
	}

	lo, carry := bits.Add64(lo, uint64(c), 0)
	hi, _ = bits.Add64(hi, ext, carry)

	switch {
	case hi == 0 && lo <= math.MaxInt64:
		return int64(lo), true
	case hi == math.MaxUint64 && lo > math.MaxInt64:
		return int64(lo), true
	default:
		return 0, false
	}
}

// sumInts returns the exact sum of xs when it fits in int; partial sums may
// leave int.
func sumInts(xs []int) (int, bool) {
	var hi, lo uint64

	for _, x := range xs {
		var carry uint64

		lo, carry = bits.Add64(lo, uint64(x), 0)
		hi += carry

		if x < 0 {
			hi += math.MaxUint64 // sign extension
		}
	}

	switch {
	case hi == 0 && lo <= math.MaxInt64:
		return int(lo), true
	case hi == math.MaxUint64 && lo > math.MaxInt64:
		return int(lo), true
	default:
		return 0, false
	}
}

// addInt returns a + b when it fits in int.
func addInt(a, b int) (int, bool) {
	s := a + b
	if (b > 0 && s < a) || (b < 0 && s > a) {
		return 0, false
	}

	return s, true
}

// subInt returns a − b when it fits in int.
func subInt(a, b int) (int, bool) {
	s := a - b
	if (b > 0 && s > a) || (b < 0 && s < a) {
		return 0, false
	}

	return s, true
}
