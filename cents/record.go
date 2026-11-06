package cents

import (
	"math/big"
	"strconv"
	"strings"
)

// An exact value v is the record
//
//	v = r + e3·J(3) + e5·J(5) + ... + e31·J(31),
//
// where r and the exponents ep are rational numbers and J(p) = 1200·log2(p)
// is the size of the prime p in cents. The prime 2 is part of r, because
// J(2) = 1200. The numbers 1, log2(3), ..., log2(31) are linearly
// independent over the rationals, so every value has exactly one record, and
// two values are equal if and only if their records are equal.
//
// A Cents value keeps its record in a canonical string encoding: the
// reduced r, followed by ";p=ep" for each nonzero exponent in increasing
// order of p. The rationals are written as "n" or "n/d" in base 10. The
// record of 0 is the empty string, so that the zero Cents is 0.

// primes are the odd primes of the record, in increasing order.
var primes = [numPrimes]int64{3, 5, 7, 11, 13, 17, 19, 23, 29, 31}

// numPrimes is the number of odd primes of the record.
const numPrimes = 10

// allPrimes are the primes of the prime limit, starting with 2.
var allPrimes = [numPrimes + 1]int64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31}

// record is a decoded exact value. Its fields are never shared between
// records: every operation that returns a record allocates new numbers.
type record struct {
	r big.Rat
	e [numPrimes]big.Rat
}

var (
	bigOne     = big.NewInt(1)
	big1200    = big.NewInt(1200)
	rat1200    = big.NewRat(1200, 1)
	maxDen     = pow10(30) // largest reduced denominator of a record
	maxExp     = pow10(24) // largest magnitude of a prime exponent
	magnitude  = big.NewRat(MaxMagnitude, 1)
	nMagnitude = big.NewRat(-MaxMagnitude, 1)
)

// pow10 returns 10^n.
func pow10(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

// decode returns the record of an encoding. The encoding is produced by
// encode, so it is always valid.
func decode(s string) *record {
	rc := new(record)
	if s == "" {
		return rc
	}
	head, tail, _ := strings.Cut(s, ";")
	setRat(&rc.r, head)
	for tail != "" {
		var part string
		part, tail, _ = strings.Cut(tail, ";")
		ps, es, _ := strings.Cut(part, "=")
		p, _ := strconv.ParseInt(ps, 10, 64)
		setRat(&rc.e[primeIndex(p)], es)
	}
	return rc
}

// setRat sets z to the rational "n" or "n/d" written in base 10.
func setRat(z *big.Rat, s string) {
	ns, ds, frac := strings.Cut(s, "/")
	n, _ := new(big.Int).SetString(ns, 10)
	if !frac {
		z.SetInt(n)
		return
	}
	d, _ := new(big.Int).SetString(ds, 10)
	z.SetFrac(n, d)
}

// primeIndex returns the index of the odd prime p in primes; p is one of
// primes.
func primeIndex(p int64) int {
	i := 0
	for primes[i] != p {
		i++
	}
	return i
}

// encode returns the canonical encoding of the record.
func (rc *record) encode() string {
	if rc.isZero() {
		return ""
	}
	var b strings.Builder
	b.WriteString(rc.r.RatString())
	for i := range rc.e {
		if rc.e[i].Sign() != 0 {
			b.WriteByte(';')
			b.WriteString(strconv.FormatInt(primes[i], 10))
			b.WriteByte('=')
			b.WriteString(rc.e[i].RatString())
		}
	}
	return b.String()
}

// isZero reports whether the record is 0.
func (rc *record) isZero() bool {
	return rc.r.Sign() == 0 && rc.isRational()
}

// isRational reports whether all prime exponents are 0.
func (rc *record) isRational() bool {
	for i := range rc.e {
		if rc.e[i].Sign() != 0 {
			return false
		}
	}
	return true
}

// add returns a + b.
func add(a, b *record) *record {
	z := new(record)
	z.r.Add(&a.r, &b.r)
	for i := range z.e {
		z.e[i].Add(&a.e[i], &b.e[i])
	}
	return z
}

// sub returns a - b.
func sub(a, b *record) *record {
	z := new(record)
	z.r.Sub(&a.r, &b.r)
	for i := range z.e {
		z.e[i].Sub(&a.e[i], &b.e[i])
	}
	return z
}

// neg returns -a.
func neg(a *record) *record {
	z := new(record)
	z.r.Neg(&a.r)
	for i := range z.e {
		z.e[i].Neg(&a.e[i])
	}
	return z
}

// scale returns a·k.
func scale(a *record, k *big.Rat) *record {
	z := new(record)
	z.r.Mul(&a.r, k)
	for i := range z.e {
		z.e[i].Mul(&a.e[i], k)
	}
	return z
}

// withinLimits reports whether the record is within the record limits: every
// reduced denominator is at most 10^30, and every prime exponent is at most
// 10^24 in magnitude.
func (rc *record) withinLimits() bool {
	if rc.r.Denom().Cmp(maxDen) > 0 {
		return false
	}
	var t big.Int
	for i := range rc.e {
		e := &rc.e[i]
		if e.Sign() == 0 {
			continue
		}
		if e.Denom().Cmp(maxDen) > 0 {
			return false
		}
		t.Mul(maxExp, e.Denom())
		if t.CmpAbs(e.Num()) < 0 {
			return false
		}
	}
	return true
}

// inRange reports whether the exact value is within -MaxMagnitude to
// MaxMagnitude, bounds included.
func (rc *record) inRange() bool {
	return rc.cmpRat(magnitude) <= 0 && rc.cmpRat(nMagnitude) >= 0
}

// check returns ErrTooComplex if the record is beyond the record limits and
// ErrOutOfRange if its value is out of range, in this order; op names the
// operation in the error message.
func (rc *record) check(op string) error {
	if !rc.withinLimits() {
		return errorf(ErrTooComplex, "%s: exact value beyond the record limits", op)
	}
	if !rc.inRange() {
		return errorf(ErrOutOfRange, "%s: value outside ±%d cents", op, MaxMagnitude)
	}
	return nil
}

// fromExponents returns the record of the cents of 2^a2 · Π p^e[i].
func fromExponents(a2 int64, e *[numPrimes]int64) *record {
	rc := new(record)
	rc.r.SetInt64(a2)
	rc.r.Mul(&rc.r, rat1200)
	for i := range rc.e {
		rc.e[i].SetInt64(e[i])
	}
	return rc
}

// proportional returns λ with c = λ·d if the records are proportional; d is
// not 0.
func proportional(c, d *record) (*big.Rat, bool) {
	var lambda *big.Rat
	switch {
	case d.r.Sign() != 0:
		lambda = new(big.Rat).Quo(&c.r, &d.r)
	default:
		for i := range d.e {
			if d.e[i].Sign() != 0 {
				lambda = new(big.Rat).Quo(&c.e[i], &d.e[i])
				break
			}
		}
	}
	var t big.Rat
	if t.Mul(lambda, &d.r).Cmp(&c.r) != 0 {
		return nil, false
	}
	for i := range d.e {
		if t.Mul(lambda, &d.e[i]).Cmp(&c.e[i]) != 0 {
			return nil, false
		}
	}
	return lambda, true
}

// ratioParts returns the reduced n and d with v = 1200·log2(n/d) if the value
// is the cents of a ratio of integers and both numbers have at most maxBits
// bits.
func (rc *record) ratioParts(maxBits int) (n, d *big.Int, ok bool) {
	if !rc.r.IsInt() {
		return nil, nil, false
	}
	var a2, rem big.Int
	a2.QuoRem(rc.r.Num(), big1200, &rem)
	// Each exponent is at most maxBits in magnitude if both terms have at
	// most maxBits bits; check this before computing the powers.
	if rem.Sign() != 0 || !a2.IsInt64() || absInt64(a2.Int64()) > int64(maxBits) {
		return nil, nil, false
	}
	for i := range rc.e {
		e := &rc.e[i]
		if !e.IsInt() || e.Num().BitLen() > 32 || absInt64(e.Num().Int64()) > int64(maxBits) {
			return nil, nil, false
		}
	}
	n, d = big.NewInt(1), big.NewInt(1)
	mulPower(n, d, 2, a2.Int64())
	for i := range rc.e {
		mulPower(n, d, primes[i], rc.e[i].Num().Int64())
	}
	if n.BitLen() > maxBits || d.BitLen() > maxBits {
		return nil, nil, false
	}
	return n, d, true
}

// mulPower multiplies n by p^k if k > 0 and d by p^-k if k < 0.
func mulPower(n, d *big.Int, p, k int64) {
	switch {
	case k > 0:
		n.Mul(n, new(big.Int).Exp(big.NewInt(p), big.NewInt(k), nil))
	case k < 0:
		d.Mul(d, new(big.Int).Exp(big.NewInt(p), big.NewInt(-k), nil))
	}
}

// absInt64 returns |x| for x > math.MinInt64.
func absInt64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// factor removes the primes up to 31 from n and returns their exponents in
// the order of allPrimes; n keeps the rest, which is 1 if n has no larger
// prime factor. n is positive.
func factor(n *big.Int) [numPrimes + 1]int64 {
	var exps [numPrimes + 1]int64
	if n.IsUint64() {
		u := n.Uint64()
		for i, p := range allPrimes {
			for u%uint64(p) == 0 {
				u /= uint64(p)
				exps[i]++
			}
		}
		n.SetUint64(u)
		return exps
	}
	var q, m, bp big.Int
	for i, p := range allPrimes {
		bp.SetInt64(p)
		for {
			q.QuoRem(n, &bp, &m)
			if m.Sign() != 0 {
				break
			}
			n.Set(&q)
			exps[i]++
		}
	}
	return exps
}

// ratioRecord returns the record of 1200·log2(n/d) for positive n and d, or
// ErrPrimeLimit if the reduced ratio has a prime factor greater than 31. n
// and d are not modified.
func ratioRecord(n, d *big.Int, op string) (*record, error) {
	var g big.Int
	g.GCD(nil, nil, n, d)
	rn := new(big.Int).Quo(n, &g)
	rd := new(big.Int).Quo(d, &g)
	en := factor(rn)
	ed := factor(rd)
	if rn.Cmp(bigOne) != 0 || rd.Cmp(bigOne) != 0 {
		return nil, errorf(ErrPrimeLimit, "%s: ratio has a prime factor greater than %d", op, MaxPrime)
	}
	var e [numPrimes]int64
	for i := range e {
		e[i] = en[i+1] - ed[i+1]
	}
	return fromExponents(en[0]-ed[0], &e), nil
}
