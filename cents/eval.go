package cents

import (
	"math/big"
	"sync"
)

// Evaluation with a guaranteed error bound and growing precision.
//
// An irrational value is never equal to a rational number, in particular to
// a rounding boundary (a half of a decimal digit, a midpoint between two
// float64 values, a multiple of 1200, the range bound). Every decision on an
// irrational value is therefore made from an interval that contains the
// exact value: if both ends of the interval give the same answer, the answer
// is correct; otherwise the precision is doubled. Rational values are
// decided exactly, without intervals. The logarithms are computed by the
// package itself, so the results do not depend on math.Log2 or on the
// platform.
//
// Fixed-point numbers are integers X that stand for X/2^prec.

const (
	// guardBits are the extra bits of the internal computations of the
	// logarithms and powers of two.
	guardBits = 64
	// tablePrec is the precision of the table of J(p).
	tablePrec = 2048
	// minPrec is the smallest working precision.
	minPrec = 64
)

var big3 = big.NewInt(3)

// atanhFixed returns S with atanh(a/b)·2^w - (3N+2) <= S <= atanh(a/b)·2^w,
// where N is the number of terms of the series; 0 <= a/b <= 1/3.
func atanhFixed(a, b *big.Int, w uint) *big.Int {
	sum := new(big.Int)
	pw := new(big.Int).Lsh(a, w)
	pw.Quo(pw, b)
	a2 := new(big.Int).Mul(a, a)
	b2 := new(big.Int).Mul(b, b)
	var t, k big.Int
	for n := int64(1); pw.Sign() != 0; n += 2 {
		k.SetInt64(n)
		t.Quo(pw, &k)
		sum.Add(sum, &t)
		pw.Mul(pw, a2)
		pw.Quo(pw, b2)
	}
	return sum
}

// centsLog returns C with |C - 1200·log2(n/d)·2^prec| <= 2 for positive n
// and d. The ratio is reduced to 2^k·m with 1 <= m < 2, and
// log2(m) = atanh(x)/atanh(1/3) with x = (m-1)/(m+1) < 1/3.
func centsLog(n, d *big.Int, prec uint) *big.Int {
	k := n.BitLen() - d.BitLen()
	num := new(big.Int).Set(n)
	den := new(big.Int).Set(d)
	if k >= 0 {
		den.Lsh(den, uint(k))
	} else {
		num.Lsh(num, uint(-k))
	}
	if num.Cmp(den) < 0 {
		num.Lsh(num, 1)
		k--
	}
	w := prec + guardBits
	a := new(big.Int).Sub(num, den)
	b := new(big.Int).Add(num, den)
	frac := atanhFixed(a, b, w)
	frac.Lsh(frac, w)
	frac.Quo(frac, atanhFixed(bigOne, big3, w))
	l := new(big.Int).Lsh(big.NewInt(int64(k)), w)
	l.Add(l, frac)
	l.Mul(l, big1200)
	return l.Rsh(l, w-prec)
}

// jTable holds J(p) = 1200·log2(p) for the odd primes at tablePrec, each
// within 2 units of the last place. It is computed once and only read.
var jTable = sync.OnceValue(func() *[numPrimes]*big.Int {
	var t [numPrimes]*big.Int
	for i, p := range primes {
		t[i] = centsLog(big.NewInt(p), bigOne, tablePrec)
	}
	return &t
})

// jFixed returns J(p) for the odd primes at prec, each within 3 units of the
// last place.
func jFixed(prec uint) [numPrimes]*big.Int {
	var js [numPrimes]*big.Int
	if prec <= tablePrec {
		t := jTable()
		for i := range js {
			js[i] = new(big.Int).Rsh(t[i], tablePrec-prec)
		}
		return js
	}
	for i, p := range primes {
		js[i] = centsLog(big.NewInt(p), bigOne, prec)
	}
	return js
}

// bounds returns lo and hi with lo/2^prec <= v <= hi/2^prec for the value v
// of the record.
func (rc *record) bounds(prec uint) (lo, hi *big.Int) {
	v := new(big.Int).Lsh(rc.r.Num(), prec)
	v.Div(v, rc.r.Denom())
	lo = new(big.Int).Set(v)
	hi = v.Add(v, bigOne)
	if rc.isRational() {
		return lo, hi
	}
	js := jFixed(prec)
	var t, errBound big.Int
	for i := range rc.e {
		e := &rc.e[i]
		if e.Sign() == 0 {
			continue
		}
		// e·J(p) lies within [t - 3|e|, t + 1 + 3|e|].
		t.Mul(e.Num(), js[i])
		t.Div(&t, e.Denom())
		errBound.Abs(e.Num())
		errBound.Mul(&errBound, big3)
		ceilDiv(&errBound, &errBound, e.Denom())
		lo.Add(lo, &t)
		lo.Sub(lo, &errBound)
		hi.Add(hi, &t)
		hi.Add(hi, bigOne)
		hi.Add(hi, &errBound)
	}
	return lo, hi
}

// startPrec returns the first working precision for the record: enough bits
// for its largest prime exponent and a margin.
func (rc *record) startPrec() uint {
	prec := uint(minPrec)
	for i := range rc.e {
		e := &rc.e[i]
		if e.Sign() == 0 {
			continue
		}
		if b := uint(max(e.Num().BitLen()-e.Denom().BitLen(), 0)); b+minPrec > prec {
			prec = b + minPrec
		}
	}
	return prec
}

// sign returns the sign of the value.
func (rc *record) sign() int {
	if rc.isRational() {
		return rc.r.Sign()
	}
	for prec := rc.startPrec(); ; prec *= 2 {
		lo, hi := rc.bounds(prec)
		switch {
		case lo.Sign() > 0:
			return 1
		case hi.Sign() < 0:
			return -1
		}
	}
}

// cmpRat compares the value with the rational q.
func (rc *record) cmpRat(q *big.Rat) int {
	if rc.isRational() {
		return rc.r.Cmp(q)
	}
	var qs, t big.Int
	for prec := rc.startPrec(); ; prec *= 2 {
		lo, hi := rc.bounds(prec)
		qs.Lsh(q.Num(), prec)
		if t.Mul(hi, q.Denom()).Cmp(&qs) < 0 {
			return -1
		}
		if t.Mul(lo, q.Denom()).Cmp(&qs) > 0 {
			return 1
		}
	}
}

// ceilDiv sets z = ceil(x/y) for y > 0 and returns z.
func ceilDiv(z, x, y *big.Int) *big.Int {
	var m big.Int
	z.DivMod(x, y, &m)
	if m.Sign() != 0 {
		z.Add(z, bigOne)
	}
	return z
}

// roundHalfAway returns the integer nearest to num/den for den > 0; an
// exact half is rounded away from zero.
func roundHalfAway(num, den *big.Int) *big.Int {
	t := new(big.Int).Abs(num)
	t.Lsh(t, 1)
	t.Add(t, den)
	d2 := new(big.Int).Lsh(den, 1)
	t.Quo(t, d2)
	if num.Sign() < 0 {
		t.Neg(t)
	}
	return t
}

// roundRat returns the integer nearest to x·mul; an exact half is rounded
// away from zero.
func roundRat(x *big.Rat, mul *big.Int) *big.Int {
	n := new(big.Int).Mul(x.Num(), mul)
	return roundHalfAway(n, x.Denom())
}

// roundValue returns the integer nearest to v·mul/div for the value v of
// the record and positive mul and div; an exact half, which only a rational
// value can have, is rounded away from zero.
func (rc *record) roundValue(mul, div *big.Int) *big.Int {
	if rc.isRational() {
		n := new(big.Int).Mul(rc.r.Num(), mul)
		d := new(big.Int).Mul(rc.r.Denom(), div)
		return roundHalfAway(n, d)
	}
	var n, d big.Int
	for prec := rc.startPrec() + uint(mul.BitLen()); ; prec *= 2 {
		lo, hi := rc.bounds(prec)
		d.Lsh(div, prec)
		a := roundHalfAway(n.Mul(lo, mul), &d)
		b := roundHalfAway(n.Mul(hi, mul), &d)
		if a.Cmp(b) == 0 {
			return a
		}
	}
}

// floorValue returns floor(v/div) for the value v of the record and a
// positive div.
func (rc *record) floorValue(div *big.Int) *big.Int {
	if rc.isRational() {
		d := new(big.Int).Mul(rc.r.Denom(), div)
		return new(big.Int).Div(rc.r.Num(), d)
	}
	var d big.Int
	for prec := rc.startPrec(); ; prec *= 2 {
		lo, hi := rc.bounds(prec)
		d.Lsh(div, prec)
		a := lo.Div(lo, &d)
		b := hi.Div(hi, &d)
		if a.Cmp(b) == 0 {
			return a
		}
	}
}

// round53 returns m·2^exp rounded to a 53-bit mantissa, to nearest with
// ties to even, without a limit on the exponent.
func round53(m *big.Int, exp int) *big.Float {
	z := new(big.Float).SetPrec(53).SetMode(big.ToNearestEven)
	z.SetInt(m)
	return z.SetMantExp(z, exp)
}

// exp2Frac returns E with |E - 2^(f/2^prec)·2^prec| <= 2 for
// 0 <= f < 2^prec. It sums the Taylor series of exp(u) for
// u = (f/2^prec)·ln 2 < 0.7.
func exp2Frac(f *big.Int, prec uint) *big.Int {
	w := prec + guardBits
	ln2 := atanhFixed(bigOne, big3, w)
	ln2.Lsh(ln2, 1)
	u := new(big.Int).Mul(f, ln2)
	u.Rsh(u, prec)
	sum := new(big.Int).Lsh(bigOne, w)
	t := new(big.Int).Set(sum)
	var k big.Int
	for i := int64(1); ; i++ {
		t.Mul(t, u)
		t.Rsh(t, w)
		k.SetInt64(i)
		t.Quo(t, &k)
		if t.Sign() == 0 {
			break
		}
		sum.Add(sum, t)
	}
	return sum.Rsh(sum, w-prec)
}
