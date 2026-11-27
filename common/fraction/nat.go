package fraction

import "math/bits"

// Word-level primitives behind the GCD and division of u256 and u512 values.
//
// Magnitudes are little-endian slices of 64-bit words. A slice is normalized
// when its most significant word is non-zero; the empty slice is zero. All
// slices refer to fixed-size arrays owned by the caller, so the primitives do
// not allocate and do not use math/big.
//
// The sizes are bounded: dividends and GCD operands have at most 8 words
// (u512), so every loop runs a bounded number of iterations.

// maxWords is the word length of the longest operand (u512).
const maxWords = 8

// natLen returns the number of significant words of x.
func natLen(x []uint64) int {
	n := len(x)
	for n > 0 && x[n-1] == 0 {
		n--
	}
	return n
}

// natCmp compares normalized magnitudes a and b.
func natCmp(a, b []uint64) int {
	if len(a) != len(b) {
		if len(a) < len(b) {
			return -1
		}
		return 1
	}
	for i := len(a) - 1; i >= 0; i-- {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

// natDivWord sets q[i] for every i < len(a) to the words of a / d and returns
// a mod d. a is normalized and non-empty, d is non-zero, q is nil or has at
// least len(a) words.
func natDivWord(q, a []uint64, d uint64) uint64 {
	top := len(a) - 1
	r := a[top] % d
	if q != nil {
		q[top] = a[top] / d
	}
	for i := top - 1; i >= 0; i-- {
		var qi uint64
		qi, r = bits.Div64(r, a[i], d)
		if q != nil {
			q[i] = qi
		}
	}
	return r
}

// div128 divides the two-word value a1:a0 by the two-word value b1:b0 with
// b1 != 0. The quotient fits in one word.
func div128(a0, a1, b0, b1 uint64) (q, r0, r1 uint64) {
	s := uint(bits.LeadingZeros64(b1))
	bn1 := b1<<s | b0>>(64-s)
	an2 := a1 >> (64 - s)
	an1 := a1<<s | a0>>(64-s)
	// The estimate from the normalized leading words exceeds the quotient by
	// at most two (Knuth, TAOCP vol. 2, 4.3.1, theorem B).
	q, _ = bits.Div64(an2, an1, bn1)
	ph0, p0 := bits.Mul64(q, b0)
	ph1, p1 := bits.Mul64(q, b1)
	var c uint64
	p1, c = bits.Add64(p1, ph0, 0)
	p2 := ph1 + c
	for p2 != 0 || p1 > a1 || (p1 == a1 && p0 > a0) {
		q--
		var bw uint64
		p0, bw = bits.Sub64(p0, b0, 0)
		p1, bw = bits.Sub64(p1, b1, bw)
		p2 -= bw
	}
	var bw uint64
	r0, bw = bits.Sub64(a0, p0, 0)
	r1, _ = bits.Sub64(a1, p1, bw)
	return q, r0, r1
}

// divMod2 divides a1:a0 by b1:b0 != 0, both at most two words.
func divMod2(a0, a1, b0, b1 uint64) (q0, q1, r0, r1 uint64) {
	if b1 == 0 {
		if a1 == 0 {
			return a0 / b0, 0, a0 % b0, 0
		}
		q1 = a1 / b0
		q0, r0 = bits.Div64(a1%b0, a0, b0)
		return q0, q1, r0, 0
	}
	if a1 < b1 {
		return 0, 0, a0, a1
	}
	q0, r0, r1 = div128(a0, a1, b0, b1)
	return q0, 0, r0, r1
}

// gcd2 returns gcd(a1:a0, b1:b0), with gcd(x, 0) = x.
func gcd2(a0, a1, b0, b1 uint64) (g0, g1 uint64) {
	if a1 == 0 && b1 == 0 {
		return gcdUint64(a0, b0), 0
	}
	if a1 < b1 || (a1 == b1 && a0 < b0) {
		a0, a1, b0, b1 = b0, b1, a0, a1
	}
	if b0|b1 == 0 {
		return a0, a1
	}
	return gcd128(a0, a1, b0, b1)
}

// natShl sets dst[i] = (src << s)[i] for i < len(src) and returns the bits
// shifted out of the top word. 0 <= s < 64, len(dst) >= len(src).
func natShl(dst, src []uint64, s uint) uint64 {
	var carry uint64
	for i, w := range src {
		dst[i] = w<<s | carry
		carry = w >> (64 - s)
	}
	return carry
}

// natShr sets dst = src >> s for 0 <= s < 64 and len(dst) == len(src) >= 1.
func natShr(dst, src []uint64, s uint) {
	n := len(src)
	for i := range n - 1 {
		dst[i] = src[i]>>s | src[i+1]<<(64-s)
	}
	dst[n-1] = src[n-1] >> s
}

// natDivLong divides normalized a by normalized b with
// 2 <= len(b) <= len(a) <= maxWords (Knuth, TAOCP vol. 2, 4.3.1, algorithm D).
// It writes the quotient to q (nil or at least len(a)-len(b)+1 words) and the
// remainder to r (nil or exactly len(b) words).
func natDivLong(q, r, a, b []uint64) {
	m, n := len(a), len(b)
	s := uint(bits.LeadingZeros64(b[n-1]))
	var vn [maxWords]uint64
	var un [maxWords + 1]uint64
	natShl(vn[:n], b, s)
	un[m] = natShl(un[:m], a, s)
	vTop, vNext := vn[n-1], vn[n-2]
	for j := m - n; j >= 0; j-- {
		ujn := un[j+n]
		var qhat, rhat uint64
		correct := true
		if ujn >= vTop {
			// ujn == vTop: the estimate is the largest word.
			qhat = ^uint64(0)
			var c uint64
			rhat, c = bits.Add64(un[j+n-1], vTop, 0)
			correct = c == 0
		} else {
			qhat, rhat = bits.Div64(ujn, un[j+n-1], vTop)
		}
		for correct {
			ph, pl := bits.Mul64(qhat, vNext)
			if ph < rhat || (ph == rhat && pl <= un[j+n-2]) {
				break
			}
			qhat--
			var c uint64
			rhat, c = bits.Add64(rhat, vTop, 0)
			correct = c == 0
		}
		// un[j:j+n+1] -= qhat * vn[:n]
		var borrow, carry uint64
		for i := range n {
			ph, pl := bits.Mul64(qhat, vn[i])
			var c uint64
			pl, c = bits.Add64(pl, carry, 0)
			carry = ph + c
			un[j+i], borrow = bits.Sub64(un[j+i], pl, borrow)
		}
		un[j+n], borrow = bits.Sub64(un[j+n], carry, borrow)
		if borrow != 0 {
			// The estimate was one too large: add the divisor back.
			qhat--
			var c uint64
			for i := range n {
				un[j+i], c = bits.Add64(un[j+i], vn[i], c)
			}
			un[j+n] += c
		}
		if q != nil {
			q[j] = qhat
		}
	}
	if r != nil {
		natShr(r, un[:n], s)
	}
}

// natDivMod sets q = a / b and r = a mod b for b != 0. a and b may contain
// leading zero words; q has at least len(a) words and r at least len(b) words,
// and both are zero on entry.
func natDivMod(q, r, a, b []uint64) {
	m, n := natLen(a), natLen(b)
	a, b = a[:m], b[:n]
	if natCmp(a, b) < 0 {
		copy(r, a)
		return
	}
	// The callers (u256DivModWide and u512DivModWide) pass operands of which
	// at least one has three or more words, so a >= b here means len(a) >= 3;
	// natDivLong also covers len(a) == len(b) == 2 should that ever change.
	if n == 1 {
		r[0] = natDivWord(q, a, b[0])
		return
	}
	natDivLong(q[:m-n+1], r[:n], a, b)
}

// natGCD sets g = gcd(a, b), with gcd(x, 0) = x. a and b may contain leading
// zero words; len(g) >= min(len(a), len(b)) and g is zero on entry.
func natGCD(g, a, b []uint64) {
	m, n := natLen(a), natLen(b)
	if m == 0 {
		copy(g, b[:n])
		return
	}
	if n == 0 {
		copy(g, a[:m])
		return
	}
	var x, y [maxWords]uint64
	copy(x[:], a[:m])
	copy(y[:], b[:n])
	if natCmp(x[:m], y[:n]) < 0 {
		x, y = y, x
		m, n = n, m
	}
	// Invariant: x >= y > 0, m and n are their word lengths.
	for n > 2 {
		u0, u1, v0, v1, even := lehmerCosequence(lehmerLeading(x[:m], y[:n]))
		if v0 == 0 {
			// No quotient could be simulated: one Euclidean step.
			var r [maxWords]uint64
			natDivLong(nil, r[:n], x[:m], y[:n])
			x, m = y, n
			y, n = r, natLen(r[:n])
			if n == 0 {
				copy(g, x[:m])
				return
			}
			continue
		}
		// A Lehmer update never makes y zero, so no zero check follows it. Let
		// s be the shift of the leading words, X = ⌊x/2^s⌋ and Y = ⌊y/2^s⌋, so
		// that x = 2^s·(X + εx) and y = 2^s·(Y + εy) with εx, εy in [0, 1).
		// The new y is σ·(u1·x − v1·y) with σ = ±1, and its simulated value
		// r = σ·(u1·X − v1·Y) is the last divisor of lehmerCosequence, which
		// divides only when r >= v1. Once a step has been simulated (v0 != 0)
		// the cofactors satisfy 0 <= u1 <= v1 and v1 >= 1, so
		// σ·(u1·εx − v1·εy) > −v1 and the new y = 2^s·(r + σ·(u1·εx − v1·εy))
		// exceeds 2^s·(r − v1) >= 0.
		lehmerUpdate(x[:m], y[:m], u0, u1, v0, v1, even)
		m, n = natLen(x[:m]), natLen(y[:m])
	}
	if m > 2 {
		var r [2]uint64
		if n == 1 {
			r[0] = natDivWord(nil, x[:m], y[0])
		} else {
			natDivLong(nil, r[:], x[:m], y[:2])
		}
		x[0], x[1] = y[0], y[1]
		y[0], y[1] = r[0], r[1]
	}
	g[0], g[1] = gcd128(x[0], x[1], y[0], y[1])
}

// gcd128 returns gcd(a, b) for two-word values a = a1:a0 >= b = b1:b0 > 0
// (Lehmer's algorithm with Collins' condition; Knuth, TAOCP vol. 2, 4.5.2).
func gcd128(a0, a1, b0, b1 uint64) (g0, g1 uint64) {
	for b1 != 0 {
		h := uint(bits.LeadingZeros64(a1))
		x := a1<<h | a0>>(64-h)
		y := b1<<h | b0>>(64-h)
		u0, u1, v0, v1, even := lehmerCosequence(x, y)
		if v0 == 0 {
			// No quotient could be simulated: one Euclidean step.
			var r0, r1 uint64
			_, r0, r1 = div128(a0, a1, b0, b1)
			a0, a1, b0, b1 = b0, b1, r0, r1
			continue
		}
		// a' = ±(u0·a − v0·b), b' = ±(v1·b − u1·a); both lie in [0, b], so
		// arithmetic modulo 2^128 is exact.
		ua0, ua1 := mulLo128(u0, a0, a1)
		vb0, vb1 := mulLo128(v0, b0, b1)
		ub0, ub1 := mulLo128(u1, a0, a1)
		wb0, wb1 := mulLo128(v1, b0, b1)
		if even {
			a0, a1 = sub128(ua0, ua1, vb0, vb1)
			b0, b1 = sub128(wb0, wb1, ub0, ub1)
		} else {
			a0, a1 = sub128(vb0, vb1, ua0, ua1)
			b0, b1 = sub128(ub0, ub1, wb0, wb1)
		}
	}
	if b0 == 0 {
		return a0, a1
	}
	if a1 != 0 {
		var r uint64
		_, r = bits.Div64(a1%b0, a0, b0)
		a0 = r
	}
	return gcdUint64(b0, a0), 0
}

// lehmerCosequence simulates Euclidean steps on the leading words x >= y of
// two values and returns the cosequence: the values are replaced by
// u0·a + v0·b and u1·a + v1·b, where u0 and v1 are non-negative and u1 and v0
// non-positive when even is true, and the signs are reversed otherwise.
// v0 == 0 means that no step could be simulated.
func lehmerCosequence(x, y uint64) (u0, u1, v0, v1 uint64, even bool) {
	var u2, v2 uint64
	u1, v2 = 1, 1
	for y >= v2 && x-y >= v1+v2 {
		q, r := x/y, x%y
		x, y = y, r
		u0, u1, u2 = u1, u2, u1+q*u2
		v0, v1, v2 = v1, v2, v1+q*v2
		even = !even
	}
	return u0, u1, v0, v1, even
}

// lehmerLeading returns the leading 64 bits of x and the bits of y at the
// same positions, for normalized x >= y with len(y) >= 2.
func lehmerLeading(x, y []uint64) (uint64, uint64) {
	m, n := len(x), len(y)
	h := uint(bits.LeadingZeros64(x[m-1]))
	a := x[m-1]<<h | x[m-2]>>(64-h)
	var b uint64
	switch m - n {
	case 0:
		b = y[n-1]<<h | y[n-2]>>(64-h)
	case 1:
		b = y[n-1] >> (64 - h)
	}
	return a, b
}

// lehmerUpdate replaces x and y (equal lengths, y zero-extended) in place by
// u0·x + v0·y and u1·x + v1·y with the signs of the cosequence given by even
// (see lehmerCosequence). Both results lie in [0, x], so the arithmetic is
// exact modulo 2^(64·len(x)).
func lehmerUpdate(x, y []uint64, u0, u1, v0, v1 uint64, even bool) {
	y = y[:len(x)]
	var cux, cvy, cuy, cvx, bx, by uint64
	for i, xi := range x {
		yi := y[i]
		var c uint64
		h, ux := bits.Mul64(u0, xi)
		ux, c = bits.Add64(ux, cux, 0)
		cux = h + c
		h, vy := bits.Mul64(v0, yi)
		vy, c = bits.Add64(vy, cvy, 0)
		cvy = h + c
		h, uy := bits.Mul64(u1, xi)
		uy, c = bits.Add64(uy, cuy, 0)
		cuy = h + c
		h, vx := bits.Mul64(v1, yi)
		vx, c = bits.Add64(vx, cvx, 0)
		cvx = h + c
		if even {
			x[i], bx = bits.Sub64(ux, vy, bx)
			y[i], by = bits.Sub64(vx, uy, by)
		} else {
			x[i], bx = bits.Sub64(vy, ux, bx)
			y[i], by = bits.Sub64(uy, vx, by)
		}
	}
}

// mulLo128 returns w·(x1:x0) modulo 2^128.
func mulLo128(w, x0, x1 uint64) (lo, hi uint64) {
	h, l := bits.Mul64(w, x0)
	return l, h + w*x1
}

// sub128 returns (a1:a0) − (b1:b0) modulo 2^128.
func sub128(a0, a1, b0, b1 uint64) (lo, hi uint64) {
	var bw uint64
	lo, bw = bits.Sub64(a0, b0, 0)
	hi, _ = bits.Sub64(a1, b1, bw)
	return lo, hi
}
