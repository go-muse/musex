package fraction

// DenomMode specifies the denominator-selection rule for binary arithmetic
// operations ([Fraction.AddWith], [Fraction.SubWith], [Fraction.MulWith], and
// [Fraction.DivWith]).
type DenomMode uint8

// Supported denominator-selection modes (modes 1 through 6).
const (
	// DenomReduced fully reduces the result to canonical form (d > 0,
	// gcd(|n|, d) == 1, 0/1 for zero).
	DenomReduced DenomMode = iota

	// DenomUnreducedLCM uses lcm(|d1|, |d2|) for addition and subtraction
	// without post-reduction, |d1|*|d2| for multiplication, and |d1*n2| for
	// division.
	DenomUnreducedLCM

	// DenomUnreducedProduct uses |d1|*|d2| for addition, subtraction, and
	// multiplication without post-reduction, and |d1*n2| for division.
	DenomUnreducedProduct

	// DenomSameUnit requires |d1| == |d2| == D and expresses the exact result
	// with denominator D without reduction.
	DenomSameUnit

	// DenomLargerUnit expresses the exact result in the larger unit
	// D = min(|d1|, |d2|).
	DenomLargerUnit

	// DenomSmallerUnit expresses the exact result in the smaller unit
	// D = max(|d1|, |d2|).
	DenomSmallerUnit
)

// Neg returns -f with a positive denominator |d| without reducing.
func (f Fraction) Neg() (Fraction, error) {
	if !f.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if f.num.isZero() {
		return makeFraction(u256Zero, f.den, false, false), nil
	}
	resNeg := f.numNeg() == f.denNeg()
	return makeFraction(f.num, f.den, resNeg, false), nil
}

// MustNeg is like [Fraction.Neg] but panics on error.
func (f Fraction) MustNeg() Fraction {
	res, err := f.Neg()
	if err != nil {
		panic(err)
	}
	return res
}

// Abs returns |f| with a positive denominator |d| without reducing.
func (f Fraction) Abs() (Fraction, error) {
	if !f.IsValid() {
		return Fraction{}, ErrInvalid
	}
	return makeFraction(f.num, f.den, false, false), nil
}

// MustAbs is like [Fraction.Abs] but panics on error.
func (f Fraction) MustAbs() Fraction {
	res, err := f.Abs()
	if err != nil {
		panic(err)
	}
	return res
}

// Reciprocal returns 1/f with a positive denominator |n| without reducing. It
// returns [ErrInvalid] if f is invalid and [ErrDivisionByZero] if f is zero.
func (f Fraction) Reciprocal() (Fraction, error) {
	if !f.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if f.num.isZero() {
		return Fraction{}, ErrDivisionByZero
	}
	resNeg := f.numNeg() != f.denNeg()
	return makeFraction(f.den, f.num, resNeg, false), nil
}

// MustReciprocal is like [Fraction.Reciprocal] but panics on error.
func (f Fraction) MustReciprocal() Fraction {
	res, err := f.Reciprocal()
	if err != nil {
		panic(err)
	}
	return res
}

// Inc returns (f + 1) in canonical reduced form.
func (f Fraction) Inc() (Fraction, error) {
	return f.AddInt(1)
}

// MustInc is like [Fraction.Inc] but panics on error.
func (f Fraction) MustInc() Fraction {
	res, err := f.Inc()
	if err != nil {
		panic(err)
	}
	return res
}

// Dec returns (f - 1) in canonical reduced form.
func (f Fraction) Dec() (Fraction, error) {
	return f.SubInt(1)
}

// MustDec is like [Fraction.Dec] but panics on error.
func (f Fraction) MustDec() Fraction {
	res, err := f.Dec()
	if err != nil {
		panic(err)
	}
	return res
}

// StepNumerator returns a new Fraction with stored numerator n + delta and the
// same stored denominator d (including its sign) without reduction.
func (f Fraction) StepNumerator(delta int) (Fraction, error) {
	if !f.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if delta == 0 {
		return f, nil
	}
	dMag, dNeg := u256FromInt(delta)
	if f.num.isZero() {
		return makeFraction(dMag, f.den, dNeg, f.denNeg()), nil
	}
	if f.numNeg() == dNeg {
		sumMag, carry := u256Add(f.num, dMag)
		if carry != 0 || sumMag.exceedsMax() {
			return Fraction{}, ErrOutOfRange
		}
		return makeFraction(sumMag, f.den, f.numNeg(), f.denNeg()), nil
	}
	c := u256Cmp(f.num, dMag)
	if c == 0 {
		return makeFraction(u256Zero, f.den, false, f.denNeg()), nil
	}
	if c > 0 {
		diff := u256Sub(f.num, dMag)
		return makeFraction(diff, f.den, f.numNeg(), f.denNeg()), nil
	}
	diff := u256Sub(dMag, f.num)
	return makeFraction(diff, f.den, dNeg, f.denNeg()), nil
}

// MustStepNumerator is like [Fraction.StepNumerator] but panics on error.
func (f Fraction) MustStepNumerator(delta int) Fraction {
	res, err := f.StepNumerator(delta)
	if err != nil {
		panic(err)
	}
	return res
}

func powU256(base u256, exp uint64) (u256, bool) {
	if exp == 0 || base.isOne() {
		return u256One, true
	}
	if exp >= MaxComponentBits {
		return u256Zero, false
	}
	res := u256One
	cur := base
	e := exp
	for e > 0 {
		if (e & 1) != 0 {
			prod := u256Mul(res, cur)
			if !prod.fitsU256Max() {
				return u256Zero, false
			}
			res = prod.low256()
		}
		e >>= 1
		if e > 0 {
			sq := u256Mul(cur, cur)
			if !sq.fitsU256Max() {
				return u256Zero, false
			}
			cur = sq.low256()
		}
	}
	return res, true
}

// Pow returns f raised to the integer power exp in canonical reduced form. It
// returns [ErrInvalid] if f is invalid, [ErrDivisionByZero] if f is zero and
// exp < 0, and [ErrOutOfRange] if the reduced result exceeds [MaxComponent].
func (f Fraction) Pow(exp int) (Fraction, error) {
	if !f.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if exp == 0 {
		return One(), nil
	}
	if f.IsZero() {
		if exp < 0 {
			return Fraction{}, ErrDivisionByZero
		}
		return Zero(), nil
	}
	red := f.Reduce()
	p := red.num
	q := red.den
	neg := red.numNeg()
	var expMag uint64
	if exp < 0 {
		p, q = q, p
		expMag = uint64(-(exp + 1)) + 1
	} else {
		expMag = uint64(exp)
	}
	resNeg := neg && (expMag&1 == 1)
	if p.isOne() && q.isOne() {
		return makeFraction(u256One, u256One, resNeg, false), nil
	}
	pPow, okP := powU256(p, expMag)
	if !okP {
		return Fraction{}, ErrOutOfRange
	}
	qPow, okQ := powU256(q, expMag)
	if !okQ {
		return Fraction{}, ErrOutOfRange
	}
	return makeFraction(pPow, qPow, resNeg, false), nil
}

// MustPow is like [Fraction.Pow] but panics on error.
func (f Fraction) MustPow(exp int) Fraction {
	res, err := f.Pow(exp)
	if err != nil {
		panic(err)
	}
	return res
}

func signedAddU512(a u512, aNeg bool, b u512, bNeg bool) (u512, bool) {
	if a.isZero() {
		if b.isZero() {
			return u512{}, false
		}
		return b, bNeg
	}
	if b.isZero() {
		return a, aNeg
	}
	if aNeg == bNeg {
		return u512Add(a, b), aNeg
	}
	c := u512Cmp(a, b)
	if c == 0 {
		return u512{}, false
	}
	if c > 0 {
		return u512Sub(a, b), aNeg
	}
	return u512Sub(b, a), bNeg
}

func reduceU512ToFraction(num, den u512, neg bool) (Fraction, error) {
	if num.isZero() {
		return Zero(), nil
	}
	g := u512GCD(num, den)
	if !g.isOne() {
		num, _ = u512DivMod(num, g)
		den, _ = u512DivMod(den, g)
	}
	if !num.fitsU256Max() || !den.fitsU256Max() {
		return Fraction{}, ErrOutOfRange
	}
	return makeFraction(num.low256(), den.low256(), neg, false), nil
}

func scaleU512ToTargetDen(num, den u512, neg bool, targetDen u256) (Fraction, error) {
	if num.isZero() {
		return makeFraction(u256Zero, targetDen, false, false), nil
	}
	g := u512GCD(num, den)
	p := num
	q := den
	if !g.isOne() {
		p, _ = u512DivMod(num, g)
		q, _ = u512DivMod(den, g)
	}
	target512 := u512FromU256(targetDen)
	k, rem := u512DivMod(target512, q)
	if !rem.isZero() {
		return Fraction{}, ErrNotRepresentable
	}
	if !p.fitsU256Max() || !k.fitsU256Max() {
		return Fraction{}, ErrOutOfRange
	}
	prod := u256Mul(p.low256(), k.low256())
	if !prod.fitsU256Max() {
		return Fraction{}, ErrOutOfRange
	}
	return makeFraction(prod.low256(), targetDen, neg, false), nil
}

func minU256(a, b u256) u256 {
	if u256Cmp(a, b) <= 0 {
		return a
	}
	return b
}

func maxU256(a, b u256) u256 {
	if u256Cmp(a, b) >= 0 {
		return a
	}
	return b
}

func addSubWith(f, other Fraction, sub bool, mode DenomMode) (Fraction, error) {
	if !f.IsValid() || !other.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if mode > DenomSmallerUnit {
		return Fraction{}, ErrNotRepresentable
	}
	neg1 := f.Sign() < 0
	neg2 := other.Sign() < 0
	if sub && !other.num.isZero() {
		neg2 = !neg2
	}

	switch mode {
	case DenomReduced:
		t1 := u256Mul(f.num, other.den)
		t2 := u256Mul(other.num, f.den)
		rawNum, rawNeg := signedAddU512(t1, neg1, t2, neg2)
		rawDen := u256Mul(f.den, other.den)
		return reduceU512ToFraction(rawNum, rawDen, rawNeg)

	case DenomUnreducedLCM:
		g := u256GCD(f.den, other.den)
		d1DivG, _ := u256DivMod(f.den, g)
		d2DivG, _ := u256DivMod(other.den, g)
		lcm512 := u256Mul(d1DivG, other.den)
		t1 := u256Mul(f.num, d2DivG)
		t2 := u256Mul(other.num, d1DivG)
		num512, numNeg := signedAddU512(t1, neg1, t2, neg2)
		if !num512.fitsU256Max() || !lcm512.fitsU256Max() {
			return Fraction{}, ErrOutOfRange
		}
		return makeFraction(num512.low256(), lcm512.low256(), numNeg, false), nil

	case DenomUnreducedProduct:
		t1 := u256Mul(f.num, other.den)
		t2 := u256Mul(other.num, f.den)
		rawNum, rawNeg := signedAddU512(t1, neg1, t2, neg2)
		rawDen := u256Mul(f.den, other.den)
		if !rawNum.fitsU256Max() || !rawDen.fitsU256Max() {
			return Fraction{}, ErrOutOfRange
		}
		return makeFraction(rawNum.low256(), rawDen.low256(), rawNeg, false), nil

	case DenomSameUnit:
		if f.den != other.den {
			return Fraction{}, ErrNotRepresentable
		}
		num512, numNeg := signedAddU512(u512FromU256(f.num), neg1, u512FromU256(other.num), neg2)
		if !num512.fitsU256Max() {
			return Fraction{}, ErrOutOfRange
		}
		return makeFraction(num512.low256(), f.den, numNeg, false), nil

	case DenomLargerUnit:
		t1 := u256Mul(f.num, other.den)
		t2 := u256Mul(other.num, f.den)
		rawNum, rawNeg := signedAddU512(t1, neg1, t2, neg2)
		rawDen := u256Mul(f.den, other.den)
		return scaleU512ToTargetDen(rawNum, rawDen, rawNeg, minU256(f.den, other.den))

	case DenomSmallerUnit:
		t1 := u256Mul(f.num, other.den)
		t2 := u256Mul(other.num, f.den)
		rawNum, rawNeg := signedAddU512(t1, neg1, t2, neg2)
		rawDen := u256Mul(f.den, other.den)
		return scaleU512ToTargetDen(rawNum, rawDen, rawNeg, maxU256(f.den, other.den))

	default:
		return Fraction{}, ErrNotRepresentable
	}
}

// Add returns f + other in canonical reduced form ([DenomReduced]).
func (f Fraction) Add(other Fraction) (Fraction, error) {
	return addSubWith(f, other, false, DenomReduced)
}

// MustAdd is like [Fraction.Add] but panics on error.
func (f Fraction) MustAdd(other Fraction) Fraction {
	res, err := f.Add(other)
	if err != nil {
		panic(err)
	}
	return res
}

// Sub returns f - other in canonical reduced form ([DenomReduced]).
func (f Fraction) Sub(other Fraction) (Fraction, error) {
	return addSubWith(f, other, true, DenomReduced)
}

// MustSub is like [Fraction.Sub] but panics on error.
func (f Fraction) MustSub(other Fraction) Fraction {
	res, err := f.Sub(other)
	if err != nil {
		panic(err)
	}
	return res
}

// Mul returns f * other in canonical reduced form ([DenomReduced]).
func (f Fraction) Mul(other Fraction) (Fraction, error) {
	return f.MulWith(other, DenomReduced)
}

// MustMul is like [Fraction.Mul] but panics on error.
func (f Fraction) MustMul(other Fraction) Fraction {
	res, err := f.Mul(other)
	if err != nil {
		panic(err)
	}
	return res
}

// Div returns f / other in canonical reduced form ([DenomReduced]).
func (f Fraction) Div(other Fraction) (Fraction, error) {
	return f.DivWith(other, DenomReduced)
}

// MustDiv is like [Fraction.Div] but panics on error.
func (f Fraction) MustDiv(other Fraction) Fraction {
	res, err := f.Div(other)
	if err != nil {
		panic(err)
	}
	return res
}

// AddInt returns f + n in canonical reduced form.
func (f Fraction) AddInt(n int) (Fraction, error) {
	return f.Add(FromInt(n))
}

// MustAddInt is like [Fraction.AddInt] but panics on error.
func (f Fraction) MustAddInt(n int) Fraction {
	res, err := f.AddInt(n)
	if err != nil {
		panic(err)
	}
	return res
}

// SubInt returns f - n in canonical reduced form.
func (f Fraction) SubInt(n int) (Fraction, error) {
	return f.Sub(FromInt(n))
}

// MustSubInt is like [Fraction.SubInt] but panics on error.
func (f Fraction) MustSubInt(n int) Fraction {
	res, err := f.SubInt(n)
	if err != nil {
		panic(err)
	}
	return res
}

// MulInt returns f * n in canonical reduced form.
func (f Fraction) MulInt(n int) (Fraction, error) {
	return f.Mul(FromInt(n))
}

// MustMulInt is like [Fraction.MulInt] but panics on error.
func (f Fraction) MustMulInt(n int) Fraction {
	res, err := f.MulInt(n)
	if err != nil {
		panic(err)
	}
	return res
}

// DivInt returns f / n in canonical reduced form.
func (f Fraction) DivInt(n int) (Fraction, error) {
	return f.Div(FromInt(n))
}

// MustDivInt is like [Fraction.DivInt] but panics on error.
func (f Fraction) MustDivInt(n int) Fraction {
	res, err := f.DivInt(n)
	if err != nil {
		panic(err)
	}
	return res
}

// AddWith returns f + other using the denominator-selection rule mode.
func (f Fraction) AddWith(other Fraction, mode DenomMode) (Fraction, error) {
	return addSubWith(f, other, false, mode)
}

// MustAddWith is like [Fraction.AddWith] but panics on error.
func (f Fraction) MustAddWith(other Fraction, mode DenomMode) Fraction {
	res, err := f.AddWith(other, mode)
	if err != nil {
		panic(err)
	}
	return res
}

// SubWith returns f - other using the denominator-selection rule mode.
func (f Fraction) SubWith(other Fraction, mode DenomMode) (Fraction, error) {
	return addSubWith(f, other, true, mode)
}

// MustSubWith is like [Fraction.SubWith] but panics on error.
func (f Fraction) MustSubWith(other Fraction, mode DenomMode) Fraction {
	res, err := f.SubWith(other, mode)
	if err != nil {
		panic(err)
	}
	return res
}

// MulWith returns f * other using the denominator-selection rule mode.
func (f Fraction) MulWith(other Fraction, mode DenomMode) (Fraction, error) {
	if !f.IsValid() || !other.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if mode > DenomSmallerUnit {
		return Fraction{}, ErrNotRepresentable
	}
	rawNum := u256Mul(f.num, other.num)
	rawDen := u256Mul(f.den, other.den)
	rawNeg := (f.Sign() * other.Sign()) < 0

	switch mode {
	case DenomReduced:
		return reduceU512ToFraction(rawNum, rawDen, rawNeg)
	case DenomUnreducedLCM, DenomUnreducedProduct:
		if !rawNum.fitsU256Max() || !rawDen.fitsU256Max() {
			return Fraction{}, ErrOutOfRange
		}
		return makeFraction(rawNum.low256(), rawDen.low256(), rawNeg, false), nil
	case DenomSameUnit:
		if f.den != other.den {
			return Fraction{}, ErrNotRepresentable
		}
		return scaleU512ToTargetDen(rawNum, rawDen, rawNeg, f.den)
	case DenomLargerUnit:
		return scaleU512ToTargetDen(rawNum, rawDen, rawNeg, minU256(f.den, other.den))
	case DenomSmallerUnit:
		return scaleU512ToTargetDen(rawNum, rawDen, rawNeg, maxU256(f.den, other.den))
	default:
		return Fraction{}, ErrNotRepresentable
	}
}

// MustMulWith is like [Fraction.MulWith] but panics on error.
func (f Fraction) MustMulWith(other Fraction, mode DenomMode) Fraction {
	res, err := f.MulWith(other, mode)
	if err != nil {
		panic(err)
	}
	return res
}

// DivWith returns f / other using the denominator-selection rule mode.
func (f Fraction) DivWith(other Fraction, mode DenomMode) (Fraction, error) {
	if !f.IsValid() || !other.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if other.IsZero() {
		return Fraction{}, ErrDivisionByZero
	}
	if mode > DenomSmallerUnit {
		return Fraction{}, ErrNotRepresentable
	}
	rawNum := u256Mul(f.num, other.den)
	rawDen := u256Mul(f.den, other.num)
	rawNeg := (f.Sign() * other.Sign()) < 0

	switch mode {
	case DenomReduced:
		return reduceU512ToFraction(rawNum, rawDen, rawNeg)
	case DenomUnreducedLCM, DenomUnreducedProduct:
		if !rawNum.fitsU256Max() || !rawDen.fitsU256Max() {
			return Fraction{}, ErrOutOfRange
		}
		return makeFraction(rawNum.low256(), rawDen.low256(), rawNeg, false), nil
	case DenomSameUnit:
		if f.den != other.den {
			return Fraction{}, ErrNotRepresentable
		}
		return scaleU512ToTargetDen(rawNum, rawDen, rawNeg, f.den)
	case DenomLargerUnit:
		return scaleU512ToTargetDen(rawNum, rawDen, rawNeg, minU256(f.den, other.den))
	case DenomSmallerUnit:
		return scaleU512ToTargetDen(rawNum, rawDen, rawNeg, maxU256(f.den, other.den))
	default:
		return Fraction{}, ErrNotRepresentable
	}
}

// MustDivWith is like [Fraction.DivWith] but panics on error.
func (f Fraction) MustDivWith(other Fraction, mode DenomMode) Fraction {
	res, err := f.DivWith(other, mode)
	if err != nil {
		panic(err)
	}
	return res
}

// AddInDenominator returns f + other expressed with positive denominator den.
func (f Fraction) AddInDenominator(other Fraction, den int) (Fraction, error) {
	if !f.IsValid() || !other.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if den <= 0 {
		return Fraction{}, ErrZeroDenominator
	}
	targetDen, _ := u256FromInt(den)
	t1 := u256Mul(f.num, other.den)
	t2 := u256Mul(other.num, f.den)
	rawNum, rawNeg := signedAddU512(t1, f.Sign() < 0, t2, other.Sign() < 0)
	rawDen := u256Mul(f.den, other.den)
	return scaleU512ToTargetDen(rawNum, rawDen, rawNeg, targetDen)
}

// MustAddInDenominator is like [Fraction.AddInDenominator] but panics on error.
func (f Fraction) MustAddInDenominator(other Fraction, den int) Fraction {
	res, err := f.AddInDenominator(other, den)
	if err != nil {
		panic(err)
	}
	return res
}

// SubInDenominator returns f - other expressed with positive denominator den.
func (f Fraction) SubInDenominator(other Fraction, den int) (Fraction, error) {
	if !f.IsValid() || !other.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if den <= 0 {
		return Fraction{}, ErrZeroDenominator
	}
	targetDen, _ := u256FromInt(den)
	t1 := u256Mul(f.num, other.den)
	t2 := u256Mul(other.num, f.den)
	neg2 := other.Sign() < 0
	if !other.num.isZero() {
		neg2 = !neg2
	}
	rawNum, rawNeg := signedAddU512(t1, f.Sign() < 0, t2, neg2)
	rawDen := u256Mul(f.den, other.den)
	return scaleU512ToTargetDen(rawNum, rawDen, rawNeg, targetDen)
}

// MustSubInDenominator is like [Fraction.SubInDenominator] but panics on error.
func (f Fraction) MustSubInDenominator(other Fraction, den int) Fraction {
	res, err := f.SubInDenominator(other, den)
	if err != nil {
		panic(err)
	}
	return res
}

// MulInDenominator returns f * other expressed with positive denominator den.
func (f Fraction) MulInDenominator(other Fraction, den int) (Fraction, error) {
	if !f.IsValid() || !other.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if den <= 0 {
		return Fraction{}, ErrZeroDenominator
	}
	targetDen, _ := u256FromInt(den)
	rawNum := u256Mul(f.num, other.num)
	rawDen := u256Mul(f.den, other.den)
	rawNeg := (f.Sign() * other.Sign()) < 0
	return scaleU512ToTargetDen(rawNum, rawDen, rawNeg, targetDen)
}

// MustMulInDenominator is like [Fraction.MulInDenominator] but panics on error.
func (f Fraction) MustMulInDenominator(other Fraction, den int) Fraction {
	res, err := f.MulInDenominator(other, den)
	if err != nil {
		panic(err)
	}
	return res
}

// DivInDenominator returns f / other expressed with positive denominator den.
func (f Fraction) DivInDenominator(other Fraction, den int) (Fraction, error) {
	if !f.IsValid() || !other.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if other.IsZero() {
		return Fraction{}, ErrDivisionByZero
	}
	if den <= 0 {
		return Fraction{}, ErrZeroDenominator
	}
	targetDen, _ := u256FromInt(den)
	rawNum := u256Mul(f.num, other.den)
	rawDen := u256Mul(f.den, other.num)
	rawNeg := (f.Sign() * other.Sign()) < 0
	return scaleU512ToTargetDen(rawNum, rawDen, rawNeg, targetDen)
}

// MustDivInDenominator is like [Fraction.DivInDenominator] but panics on error.
func (f Fraction) MustDivInDenominator(other Fraction, den int) Fraction {
	res, err := f.DivInDenominator(other, den)
	if err != nil {
		panic(err)
	}
	return res
}
