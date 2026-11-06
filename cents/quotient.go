package cents

import (
	"math/big"
	"strings"
)

// Quotient is the number c/d of two values, the result of [Cents.Quo]: for
// example, the amount of a unit in a value. It is exact when the number is
// rational, and correctly rounded to any number of digits otherwise.
//
// A Quotient is immutable, safe for concurrent use, and copied by
// assignment. It is not comparable: the package does not decide the
// equality of irrational quotients, so == does not compile and a Quotient is
// not a map key; compare the results of [Quotient.Exact],
// [Quotient.Float64], or [Quotient.Decimal] instead. The zero value is the
// number 0.
type Quotient struct {
	num, den string
	_        [0]func()
}

// Quo returns the number c/d. With a unit as d, such as [Millioctave],
// [Savart], or Step(1, n) for "1/n of an octave", it is the amount of the
// unit in c: 700 cents are 1750/3 millioctaves. It returns
// [ErrInvalidArgument] if d is 0.
func (c Cents) Quo(d Cents) (Quotient, error) {
	if d.rec == "" {
		return Quotient{}, errorf(ErrInvalidArgument, "Quo: zero divisor")
	}
	return Quotient{num: c.rec, den: d.rec}, nil
}

// MustQuo is like [Cents.Quo] but panics with the error. It is meant for
// literals and values known to be valid.
func (c Cents) MustQuo(d Cents) Quotient {
	return must(c.Quo(d))
}

// exact returns the rational value of q, if it is rational.
func (q Quotient) exact() (*big.Rat, bool) {
	if q.den == "" {
		return new(big.Rat), true
	}
	return proportional(decode(q.num), decode(q.den))
}

// Exact returns the exact value of q if it is rational, which is the case if
// and only if the records of the two values are proportional; otherwise it
// returns (nil, false). The returned number is new and belongs to the
// caller.
func (q Quotient) Exact() (*big.Rat, bool) {
	return q.exact()
}

// Float64 returns the float64 nearest to q; a rational value exactly
// halfway is rounded to the even mantissa.
func (q Quotient) Float64() float64 {
	if r, ok := q.exact(); ok {
		f, _ := r.Float64()
		return f
	}
	num, den := decode(q.num), decode(q.den)
	for prec := max(num.startPrec(), den.startPrec()); ; prec *= 2 {
		lo, hi, ok := quoBounds(num, den, prec)
		if !ok {
			continue
		}
		a := round53(lo, -int(prec))
		if a.Cmp(round53(hi, -int(prec))) == 0 {
			f, _ := a.Float64()
			return f
		}
	}
}

// Decimal returns q in decimal with digits digits after the point, correctly
// rounded; a rational value exactly halfway is rounded away from zero. The
// text is ASCII, with the minus "-"; a result rounded to zero has no minus.
// A negative digits means 0, and with 0 digits there is no point.
func (q Quotient) Decimal(digits int) string {
	digits = max(digits, 0)
	mul := pow10(digits)
	var m *big.Int
	if r, ok := q.exact(); ok {
		m = roundRat(r, mul)
	} else {
		m = q.roundIrrational(mul)
	}
	return signedDecimal(m, digits)
}

// roundIrrational returns the integer nearest to q·mul for an irrational q.
func (q Quotient) roundIrrational(mul *big.Int) *big.Int {
	num, den := decode(q.num), decode(q.den)
	var t, s big.Int
	for prec := max(num.startPrec(), den.startPrec()) + uint(mul.BitLen()); ; prec *= 2 {
		lo, hi, ok := quoBounds(num, den, prec)
		if !ok {
			continue
		}
		s.Lsh(bigOne, prec)
		a := roundHalfAway(t.Mul(lo, mul), &s)
		b := roundHalfAway(t.Mul(hi, mul), &s)
		if a.Cmp(b) == 0 {
			return a
		}
	}
}

// String returns q as "p" or "p/q" (reduced, ASCII minus) if it is rational,
// such as "53", "1750/3", or "-1/3", and otherwise Decimal(15) followed by
// "...", such as "51.150872498556794...".
func (q Quotient) String() string {
	if r, ok := q.exact(); ok {
		return r.RatString()
	}
	return q.Decimal(15) + "..."
}

// quoBounds returns lo and hi with lo/2^prec <= num/den <= hi/2^prec, or
// ok == false if the interval of den at prec contains 0.
func quoBounds(num, den *record, prec uint) (lo, hi *big.Int, ok bool) {
	a1, a2 := num.bounds(prec)
	b1, b2 := den.bounds(prec)
	if b1.Sign() <= 0 && b2.Sign() >= 0 {
		return nil, nil, false
	}
	var t, m big.Int
	for i, a := range []*big.Int{a1, a2} {
		for j, b := range []*big.Int{b1, b2} {
			t.Lsh(a, prec)
			var f, c big.Int
			f.Div(&t, b)
			if b.Sign() < 0 {
				// Euclidean division rounds toward +∞ for a negative divisor
				// with a nonzero remainder; take the floor.
				m.Mod(&t, b)
				if m.Sign() != 0 {
					f.Sub(&f, bigOne)
				}
			}
			c.Set(&f)
			m.Mul(&f, b)
			if m.Cmp(&t) != 0 {
				c.Add(&c, bigOne)
			}
			if i == 0 && j == 0 {
				lo, hi = new(big.Int).Set(&f), new(big.Int).Set(&c)
				continue
			}
			if f.Cmp(lo) < 0 {
				lo.Set(&f)
			}
			if c.Cmp(hi) > 0 {
				hi.Set(&c)
			}
		}
	}
	return lo, hi, true
}

// signedDecimal returns the integer m/10^digits in decimal, with "-" for a
// negative m.
func signedDecimal(m *big.Int, digits int) string {
	var b strings.Builder
	if m.Sign() < 0 {
		b.WriteByte('-')
	}
	b.WriteString(scaledDecimal(new(big.Int).Abs(m), digits))
	return b.String()
}

// scaledDecimal returns the nonnegative integer m/10^digits in decimal with
// exactly digits digits after the point, and no point if digits is 0.
func scaledDecimal(m *big.Int, digits int) string {
	s := m.String()
	if digits == 0 {
		return s
	}
	if len(s) <= digits {
		s = strings.Repeat("0", digits-len(s)+1) + s
	}
	return s[:len(s)-digits] + "." + s[len(s)-digits:]
}
