package pitchclass

import "math/bits"

// trialLimit is the bound of trial division in totient: all prime factors
// below it are removed before Pollard's rho method is used. Rho with
// x² + c does not find the factors of powers of two, and small factors
// are faster to divide out.
const trialLimit = 1000

// maxFactors bounds the number of prime factors, counted with multiplicity,
// of a number below 2⁶⁴.
const maxFactors = 64

// totient returns Euler's totient of n >= 1.
//
// It removes prime factors below trialLimit by trial division, then factors
// the rest with Pollard's rho method (Brent's variant) into primes checked
// by a deterministic Miller-Rabin test. φ(n) = n·Π(1 - 1/p) over the
// distinct primes p of n is computed as phi -= phi/p, which never
// overflows.
func totient(n uint64) uint64 {
	phi := n

	for p := uint64(2); p < trialLimit && p*p <= n; p++ {
		if n%p != 0 {
			continue
		}

		phi -= phi / p
		for n%p == 0 {
			n /= p
		}
	}

	if n == 1 {
		return phi
	}

	// n > 1 has no prime factor below trialLimit, or it is a prime: the
	// loop stops early only when p*p > n.
	var f factors
	f.factor(n)

	for i, p := range f.primes[:f.count] {
		if !seenBefore(f.primes[:i], p) {
			phi -= phi / p
		}
	}

	return phi
}

// seenBefore reports whether p is among primes.
func seenBefore(primes []uint64, p uint64) bool {
	for _, q := range primes {
		if q == p {
			return true
		}
	}

	return false
}

// factors collects prime factors with multiplicity without allocation.
type factors struct {
	primes [maxFactors]uint64
	count  int
}

// factor adds the prime factors of n > 1 to f. n must be a prime or have no
// prime factor below trialLimit; rho then splits it into two factors that
// are greater than 1 and have the same property.
func (f *factors) factor(n uint64) {
	if isPrime(n) {
		f.primes[f.count] = n
		f.count++

		return
	}

	d := rho(n)
	f.factor(d)
	f.factor(n / d)
}

// isPrime reports whether n > 1 is a prime. The Miller-Rabin test with the
// prime bases 2 to 37 is exact for every n below 3.3·10²⁴, and therefore
// for every uint64.
func isPrime(n uint64) bool {
	bases := [...]uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37}

	for _, p := range bases {
		if n%p == 0 {
			return n == p
		}
	}

	d := n - 1
	s := bits.TrailingZeros64(d)
	d >>= uint(s)

	for _, a := range bases {
		if !millerRabinRound(n, a, d, s) {
			return false
		}
	}

	return true
}

// millerRabinRound reports whether the odd n > 37 with n - 1 = d·2^s passes
// the strong probable-prime test to base a.
func millerRabinRound(n, a, d uint64, s int) bool {
	x := powMod64(a, d, n)
	if x == 1 || x == n-1 {
		return true
	}

	for range s - 1 {
		x = mulMod64(x, x, n)
		if x == n-1 {
			return true
		}
	}

	return false
}

// powMod64 returns a^e mod n for n >= 1.
func powMod64(a, e, n uint64) uint64 {
	result := uint64(1) % n
	a %= n

	for e > 0 {
		if e&1 == 1 {
			result = mulMod64(result, a, n)
		}

		a = mulMod64(a, a, n)
		e >>= 1
	}

	return result
}

// rho returns a divisor d of the composite odd n with 1 < d < n, using
// Pollard's rho method in Brent's variant with f(x) = x² + c mod n. When an
// attempt fails (the gcd it finds is n), the next constant c is tried, so a
// divisor is always found. The constants are fixed: the result does not
// depend on any state.
func rho(n uint64) uint64 {
	for c := uint64(1); ; c++ {
		if d := brent(n, c); d != n {
			return d
		}
	}
}

// brent runs one attempt of Brent's variant of Pollard's rho method on n
// with f(x) = x² + c mod n and the start value 2. It returns a divisor
// 1 < d <= n of n; d = n means that the attempt failed.
func brent(n, c uint64) uint64 {
	// batch is the number of steps whose differences are multiplied before
	// one gcd is computed.
	const batch = 128

	f := func(x uint64) uint64 {
		// x² mod n + c < 2n < 2⁶⁴ because n < 2⁶³ and c < n.
		return (mulMod64(x, x, n) + c) % n
	}

	y, x, ys := uint64(2), uint64(2), uint64(2)
	g, q := uint64(1), uint64(1)

	for r := uint64(1); g == 1; r *= 2 {
		x = y
		for range r {
			y = f(y)
		}

		for k := uint64(0); k < r && g == 1; k += batch {
			ys = y
			for range min(batch, r-k) {
				y = f(y)
				q = mulMod64(q, absDiff(x, y), n)
			}

			g = gcd(q, n)
		}
	}

	if g == n {
		// The batch overshot or q became 0: repeat its steps one by one.
		for {
			ys = f(ys)

			g = gcd(absDiff(x, ys), n)
			if g > 1 {
				break
			}
		}
	}

	return g
}

// absDiff returns |a - b|.
func absDiff(a, b uint64) uint64 {
	if a > b {
		return a - b
	}

	return b - a
}
