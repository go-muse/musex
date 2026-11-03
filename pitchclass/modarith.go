package pitchclass

import (
	"iter"
	"math/bits"
)

// The helpers below implement exact arithmetic modulo n for 1 <= n <=
// math.MaxInt without overflow and without allocation. Unless stated
// otherwise, their operands are residues in the range 0 to n-1.

// mod returns the Euclidean remainder of x modulo n: the r in 0 to n-1 with
// r ≡ x (mod n). It is exact for any int x, including math.MinInt. n must
// be at least 1.
func mod(x, n int) int {
	r := x % n
	if r < 0 {
		r += n
	}

	return r
}

// addMod returns (a + b) mod n for residues a and b without computing the
// possibly overflowing sum a + b.
func addMod(a, b, n int) int {
	if a >= n-b {
		return a - (n - b)
	}

	return a + b
}

// subMod returns (a - b) mod n for residues a and b.
func subMod(a, b, n int) int {
	if a >= b {
		return a - b
	}

	return a - b + n
}

// negMod returns (-a) mod n for a residue a.
func negMod(a, n int) int {
	if a == 0 {
		return 0
	}

	return n - a
}

// mulMod returns (a * b) mod n for residues a and b, using the exact 128-bit
// product.
func mulMod(a, b, n int) int {
	return int(mulMod64(uint64(a), uint64(b), uint64(n)))
}

// mulMod64 returns (a * b) mod n for any a and b and n >= 1, using the exact
// 128-bit product.
func mulMod64(a, b, n uint64) uint64 {
	hi, lo := bits.Mul64(a, b)

	return bits.Rem64(hi, lo, n)
}

// gcd returns the greatest common divisor of a and b; gcd(0, b) is b.
func gcd(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}

	return a
}

// invertible reports whether the residue m is invertible modulo n, that is,
// gcd(m, n) = 1. Every residue is invertible modulo 1.
func invertible(m, n int) bool {
	return gcd(uint64(m), uint64(n)) == 1
}

// inverseMod returns the inverse of the invertible residue m modulo n: the r
// in 0 to n-1 with (m * r) mod n = 1 mod n. The inverse modulo 1 is 0.
//
// It uses the extended Euclidean algorithm and keeps the coefficient of m as
// a residue modulo n, so no intermediate value overflows.
func inverseMod(m, n int) int {
	if n == 1 {
		return 0
	}

	// Invariant: oldS*m ≡ oldR and s*m ≡ r (mod n).
	oldR, r := n, m
	oldS, s := 0, 1

	for r != 0 {
		q := oldR / r
		oldR, r = r, oldR-q*r
		oldS, s = s, subMod(oldS, mulMod(q%n, s, n), n)
	}

	// oldR is gcd(m, n) = 1 here.
	return oldS
}

// emptySeq returns an iterator that yields nothing. Operations return it
// with an error, so that the returned iterator is never nil.
func emptySeq[T any]() iter.Seq[T] {
	return func(func(T) bool) {}
}
