package interval

import (
	"math"
	"math/bits"

	"github.com/go-muse/musex/alteration"
)

// i128 is a signed 128-bit integer in two's complement, used for exact
// intermediate results that may not fit in an int. Only the operations needed
// by the package are defined; callers keep the values well inside the range.
type i128 struct {
	hi int64
	lo uint64
}

// i128Of returns x as an i128.
func i128Of(x int64) i128 {
	return i128{hi: x >> 63, lo: uint64(x)}
}

// add returns a + b.
func (a i128) add(b i128) i128 {
	lo, carry := bits.Add64(a.lo, b.lo, 0)
	return i128{hi: a.hi + b.hi + int64(carry), lo: lo}
}

// neg returns -a.
func (a i128) neg() i128 {
	lo := ^a.lo + 1
	hi := ^a.hi
	if lo == 0 {
		hi++
	}
	return i128{hi: hi, lo: lo}
}

// sub returns a - b.
func (a i128) sub(b i128) i128 {
	return a.add(b.neg())
}

// sign returns -1, 0, or +1 for a negative, zero, or positive a.
func (a i128) sign() int {
	switch {
	case a.hi < 0:
		return -1
	case a.hi == 0 && a.lo == 0:
		return 0
	default:
		return 1
	}
}

// cmp compares a and b: -1, 0, or +1.
func (a i128) cmp(b i128) int {
	switch {
	case a.hi < b.hi:
		return -1
	case a.hi > b.hi:
		return 1
	case a.lo < b.lo:
		return -1
	case a.lo > b.lo:
		return 1
	default:
		return 0
	}
}

// abs returns the magnitude of a as an unsigned 128-bit number.
func (a i128) abs() (hi, lo uint64) {
	if a.hi < 0 {
		n := a.neg()
		return uint64(n.hi), n.lo
	}
	return uint64(a.hi), a.lo
}

// int64 returns a as an int64 and reports whether it fits.
func (a i128) int64() (int64, bool) {
	if a.hi != int64(a.lo)>>63 {
		return 0, false
	}
	return int64(a.lo), true
}

// mul returns a * b. The exact product must fit in 128 bits.
func (a i128) mul(b int64) i128 {
	negative := (a.hi < 0) != (b < 0)
	ah, al := a.abs()
	ub := uint64(b)
	if b < 0 {
		ub = -ub
	}
	ph, pl := bits.Mul64(al, ub)
	r := i128{hi: int64(ph + ah*ub), lo: pl}
	if negative {
		r = r.neg()
	}
	return r
}

// mul64 returns the exact product a * b.
func mul64(a, b int64) i128 {
	return i128Of(a).mul(b)
}

// floorDivMod returns the floor quotient and the non-negative remainder of a
// divided by d > 0.
func (a i128) floorDivMod(d int64) (i128, int64) {
	ud := uint64(d)
	h, l := a.abs()
	qh := h / ud
	ql, r := bits.Div64(h%ud, l, ud)
	q := i128{hi: int64(qh), lo: ql}
	if a.hi >= 0 {
		return q, int64(r)
	}
	q = q.neg()
	if r == 0 {
		return q, 0
	}
	return q.sub(i128Of(1)), d - int64(r)
}

// intValue returns a as an int and reports whether it fits.
func (a i128) intValue() (int, bool) {
	v, ok := a.int64()
	return int(v), ok
}

// fractionOf returns the exact value of a as a fraction in lowest terms.
func fractionOf(a alteration.Alteration) (num, den int64) {
	p, q := a.Fraction()
	return int64(p), int64(q)
}

// alterationOf returns the alteration value num/den, den > 0, and reports
// whether it is in range. Any num and den are accepted; the check is exact.
func alterationOf(num i128, den int64) (alteration.Alteration, bool) {
	n, ok := num.int64()
	if !ok {
		// |num| >= 2^63 and den <= 2^32 give a magnitude far beyond the range.
		return alteration.Alteration{}, false
	}
	a, err := alteration.FromFraction(int(n), int(den))
	if err != nil {
		return alteration.Alteration{}, false
	}
	return a, true
}

// combine returns k + sa·a + sb·b exactly and reports whether the result is in
// range. k is small, and sa and sb are -1, 0, or +1, so the intermediate
// values fit in an int64.
func combine(k, sa int64, a alteration.Alteration, sb int64, b alteration.Alteration) (alteration.Alteration, bool) {
	pa, qa := fractionOf(a)
	pb, qb := fractionOf(b)
	num := k*qa*qb + sa*pa*qb + sb*pb*qa
	return alterationOf(i128Of(num), qa*qb)
}

// fitsSize reports whether nm1 = n - 1 gives a size n not greater than
// MaxSize.
func fitsSize(nm1 i128) (int, bool) {
	v, ok := nm1.int64()
	if !ok || v < 0 || v > math.MaxInt-1 {
		return 0, false
	}
	return int(v) + 1, true
}
