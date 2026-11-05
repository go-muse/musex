package fraction

import (
	"math"
	"math/big"
)

// RoundingMode specifies the rounding rule for explicit rounding operations.
type RoundingMode uint8

// Supported explicit rounding rules.
const (
	// RoundHalfAwayFromZero rounds ties (exact halves) away from zero
	// (±1/2 -> ±1).
	RoundHalfAwayFromZero RoundingMode = iota

	// RoundHalfToEven rounds ties (exact halves) to the nearest even integer
	// (±1/2 -> 0, ±3/2 -> ±2).
	RoundHalfToEven

	// RoundFloor rounds toward negative infinity (-∞).
	RoundFloor

	// RoundCeil rounds toward positive infinity (+∞).
	RoundCeil

	// RoundTrunc rounds toward zero (truncates the fractional part).
	RoundTrunc
)

func u512ToInt(u u512, neg bool) (int, bool) {
	if !u.fitsU256Max() {
		return 0, false
	}
	return u256ToInt(u.low256(), neg)
}

func (u u512) toBigInt(neg bool) *big.Int {
	if u.isZero() {
		return new(big.Int)
	}
	if u.fitsU256Max() {
		return u.low256().toBigInt(neg)
	}
	var buf [64]byte
	for i := range 8 {
		putUint64BE(buf[(7-i)*8:(8-i)*8], u[i])
	}
	z := new(big.Int).SetBytes(buf[:])
	if neg {
		z.Neg(z)
	}
	return z
}

// Split decomposes f toward zero into an integer part whole = trunc(f) and a
// remainder rem = f - whole with stored denominator |d|. It returns
// [ErrInvalid] if f is invalid and [ErrNotRepresentable] if whole does not fit
// in int.
func (f Fraction) Split() (whole int, rem Fraction, err error) {
	if !f.IsValid() {
		return 0, Fraction{}, ErrInvalid
	}
	neg := f.Sign() < 0
	qMag, rMag := u256DivMod(f.num, f.den)
	wInt, ok := u256ToInt(qMag, neg && !qMag.isZero())
	if !ok {
		return 0, Fraction{}, ErrNotRepresentable
	}
	rem = makeFraction(rMag, f.den, neg && !rMag.isZero(), false)
	return wInt, rem, nil
}

// MustSplit is like [Fraction.Split] but panics on error.
func (f Fraction) MustSplit() (whole int, rem Fraction) {
	w, r, err := f.Split()
	if err != nil {
		panic(err)
	}
	return w, r
}

// SplitBig decomposes f toward zero into a *big.Int whole part and a remainder
// rem with stored denominator |d|. It returns [ErrInvalid] if f is invalid.
func (f Fraction) SplitBig() (whole *big.Int, rem Fraction, err error) {
	if !f.IsValid() {
		return nil, Fraction{}, ErrInvalid
	}
	neg := f.Sign() < 0
	qMag, rMag := u256DivMod(f.num, f.den)
	whole = qMag.toBigInt(neg && !qMag.isZero())
	rem = makeFraction(rMag, f.den, neg && !rMag.isZero(), false)
	return whole, rem, nil
}

// MustSplitBig is like [Fraction.SplitBig] but panics on error.
func (f Fraction) MustSplitBig() (whole *big.Int, rem Fraction) {
	w, r, err := f.SplitBig()
	if err != nil {
		panic(err)
	}
	return w, r
}

// SplitFloor decomposes f toward negative infinity into periods = floor(f) and
// a non-negative remainder rem = f - periods (0 <= rem < 1) with stored
// denominator |d|. It returns [ErrInvalid] if f is invalid and
// [ErrNotRepresentable] if periods does not fit in int.
func (f Fraction) SplitFloor() (periods int, rem Fraction, err error) {
	if !f.IsValid() {
		return 0, Fraction{}, ErrInvalid
	}
	neg := f.Sign() < 0
	qMag, rMag := u256DivMod(f.num, f.den)
	if !neg || rMag.isZero() {
		pInt, ok := u256ToInt(qMag, neg && !qMag.isZero())
		if !ok {
			return 0, Fraction{}, ErrNotRepresentable
		}
		return pInt, makeFraction(rMag, f.den, false, false), nil
	}
	qFloorMag, _ := u256Add(qMag, u256One)
	pInt, ok := u256ToInt(qFloorMag, true)
	if !ok {
		return 0, Fraction{}, ErrNotRepresentable
	}
	remMag := u256Sub(f.den, rMag)
	return pInt, makeFraction(remMag, f.den, false, false), nil
}

// MustSplitFloor is like [Fraction.SplitFloor] but panics on error.
func (f Fraction) MustSplitFloor() (periods int, rem Fraction) {
	p, r, err := f.SplitFloor()
	if err != nil {
		panic(err)
	}
	return p, r
}

// SplitFloorBig decomposes f toward negative infinity into a *big.Int period
// count and a non-negative remainder rem (0 <= rem < 1) with stored
// denominator |d|. It returns [ErrInvalid] if f is invalid.
func (f Fraction) SplitFloorBig() (periods *big.Int, rem Fraction, err error) {
	if !f.IsValid() {
		return nil, Fraction{}, ErrInvalid
	}
	neg := f.Sign() < 0
	qMag, rMag := u256DivMod(f.num, f.den)
	if !neg || rMag.isZero() {
		return qMag.toBigInt(neg && !qMag.isZero()), makeFraction(rMag, f.den, false, false), nil
	}
	qFloorMag, _ := u256Add(qMag, u256One)
	remMag := u256Sub(f.den, rMag)
	return qFloorMag.toBigInt(true), makeFraction(remMag, f.den, false, false), nil
}

// MustSplitFloorBig is like [Fraction.SplitFloorBig] but panics on error.
func (f Fraction) MustSplitFloorBig() (periods *big.Int, rem Fraction) {
	p, r, err := f.SplitFloorBig()
	if err != nil {
		panic(err)
	}
	return p, r
}

func (f Fraction) quoRemRaw(step Fraction) (qMag u512, qNeg bool, remNum, remDen u512, remNeg bool) {
	num512 := u256Mul(f.num, step.den)
	den512 := u256Mul(f.den, step.num)
	ratioNeg := (f.Sign() * step.Sign()) < 0
	q512, r512 := u512DivMod(num512, den512)
	qNeg = ratioNeg && !q512.isZero()
	remDen = u256Mul(f.den, step.den)
	remNeg = (f.Sign() < 0) && !r512.isZero()
	return q512, qNeg, r512, remDen, remNeg
}

// QuoRem divides f by step toward zero, returning quo = trunc(f / step) and
// the canonical reduced remainder rem = (f - quo * step).Reduce().
func (f Fraction) QuoRem(step Fraction) (quo int, rem Fraction, err error) {
	if !f.IsValid() || !step.IsValid() {
		return 0, Fraction{}, ErrInvalid
	}
	if step.IsZero() {
		return 0, Fraction{}, ErrDivisionByZero
	}
	qMag, qNeg, remNum, remDen, remNeg := f.quoRemRaw(step)
	qInt, ok := u512ToInt(qMag, qNeg)
	if !ok {
		return 0, Fraction{}, ErrNotRepresentable
	}
	rem, err = reduceU512ToFraction(remNum, remDen, remNeg)
	if err != nil {
		return 0, Fraction{}, err
	}
	return qInt, rem, nil
}

// MustQuoRem is like [Fraction.QuoRem] but panics on error.
func (f Fraction) MustQuoRem(step Fraction) (quo int, rem Fraction) {
	q, r, err := f.QuoRem(step)
	if err != nil {
		panic(err)
	}
	return q, r
}

// QuoRemBig divides f by step toward zero, returning quo = trunc(f / step) as
// a *big.Int and the canonical reduced remainder rem.
func (f Fraction) QuoRemBig(step Fraction) (quo *big.Int, rem Fraction, err error) {
	if !f.IsValid() || !step.IsValid() {
		return nil, Fraction{}, ErrInvalid
	}
	if step.IsZero() {
		return nil, Fraction{}, ErrDivisionByZero
	}
	qMag, qNeg, remNum, remDen, remNeg := f.quoRemRaw(step)
	rem, err = reduceU512ToFraction(remNum, remDen, remNeg)
	if err != nil {
		return nil, Fraction{}, err
	}
	return qMag.toBigInt(qNeg), rem, nil
}

// MustQuoRemBig is like [Fraction.QuoRemBig] but panics on error.
func (f Fraction) MustQuoRemBig(step Fraction) (quo *big.Int, rem Fraction) {
	q, r, err := f.QuoRemBig(step)
	if err != nil {
		panic(err)
	}
	return q, r
}

func (f Fraction) divModFloorRaw(step Fraction) (qMag u512, qNeg bool, remNum, remDen u512) {
	num512 := u256Mul(f.num, step.den)
	den512 := u256Mul(f.den, step.num)
	fNeg := f.Sign() < 0
	q512, r512 := u512DivMod(num512, den512)
	remDen = u256Mul(f.den, step.den)
	if !fNeg || r512.isZero() {
		return q512, fNeg && !q512.isZero(), r512, remDen
	}
	qFloor := u512Add(q512, u512{1})
	remNum = u512Sub(den512, r512)
	return qFloor, true, remNum, remDen
}

// DivModFloor divides f by positive step toward negative infinity, returning
// periods = floor(f / step) and the non-negative canonical reduced remainder
// rem = (f - periods * step).Reduce() with 0 <= rem < step.
func (f Fraction) DivModFloor(step Fraction) (periods int, rem Fraction, err error) {
	if !f.IsValid() || !step.IsValid() {
		return 0, Fraction{}, ErrInvalid
	}
	if step.IsZero() {
		return 0, Fraction{}, ErrDivisionByZero
	}
	if step.Sign() < 0 {
		return 0, Fraction{}, ErrZeroDenominator
	}
	qMag, qNeg, remNum, remDen := f.divModFloorRaw(step)
	pInt, ok := u512ToInt(qMag, qNeg)
	if !ok {
		return 0, Fraction{}, ErrNotRepresentable
	}
	rem, err = reduceU512ToFraction(remNum, remDen, false)
	if err != nil {
		return 0, Fraction{}, err
	}
	return pInt, rem, nil
}

// MustDivModFloor is like [Fraction.DivModFloor] but panics on error.
func (f Fraction) MustDivModFloor(step Fraction) (periods int, rem Fraction) {
	p, r, err := f.DivModFloor(step)
	if err != nil {
		panic(err)
	}
	return p, r
}

// DivModFloorBig divides f by positive step toward negative infinity,
// returning periods = floor(f / step) as a *big.Int and the non-negative
// canonical reduced remainder rem.
func (f Fraction) DivModFloorBig(step Fraction) (periods *big.Int, rem Fraction, err error) {
	if !f.IsValid() || !step.IsValid() {
		return nil, Fraction{}, ErrInvalid
	}
	if step.IsZero() {
		return nil, Fraction{}, ErrDivisionByZero
	}
	if step.Sign() < 0 {
		return nil, Fraction{}, ErrZeroDenominator
	}
	qMag, qNeg, remNum, remDen := f.divModFloorRaw(step)
	rem, err = reduceU512ToFraction(remNum, remDen, false)
	if err != nil {
		return nil, Fraction{}, err
	}
	return qMag.toBigInt(qNeg), rem, nil
}

// MustDivModFloorBig is like [Fraction.DivModFloorBig] but panics on error.
func (f Fraction) MustDivModFloorBig(step Fraction) (periods *big.Int, rem Fraction) {
	p, r, err := f.DivModFloorBig(step)
	if err != nil {
		panic(err)
	}
	return p, r
}

// ScaleIntExact checks whether f * c is an exact integer and returns its int
// value if so. It returns [ErrInvalid] if f is invalid and
// [ErrNotRepresentable] if f * c is an exact integer that does not fit in int.
func (f Fraction) ScaleIntExact(c int) (product int, exact bool, err error) {
	if !f.IsValid() {
		return 0, false, ErrInvalid
	}
	if c == 0 || f.IsZero() {
		return 0, true, nil
	}
	red := f.Reduce()
	cMag, cNeg := u256FromInt(c)
	if !red.den.isUint64() {
		return 0, false, nil
	}
	if cMag[0]%red.den[0] != 0 {
		return 0, false, nil
	}
	k := u256{cMag[0] / red.den[0], 0, 0, 0}
	prod512 := u256Mul(red.num, k)
	prodNeg := red.numNeg() != cNeg
	prodInt, ok := u512ToInt(prod512, prodNeg)
	if !ok {
		return 0, false, ErrNotRepresentable
	}
	return prodInt, true, nil
}

// MustScaleIntExact is like [Fraction.ScaleIntExact] but panics on error.
func (f Fraction) MustScaleIntExact(c int) (product int, exact bool) {
	p, ok, err := f.ScaleIntExact(c)
	if err != nil {
		panic(err)
	}
	return p, ok
}

// IsMultipleOf reports whether f / step is an exact integer. It returns
// [ErrInvalid] if f or step is invalid and [ErrDivisionByZero] if step is
// zero.
func (f Fraction) IsMultipleOf(step Fraction) (bool, error) {
	if !f.IsValid() || !step.IsValid() {
		return false, ErrInvalid
	}
	if step.IsZero() {
		return false, ErrDivisionByZero
	}
	if f.IsZero() {
		return true, nil
	}
	num512 := u256Mul(f.num, step.den)
	den512 := u256Mul(f.den, step.num)
	_, rem512 := u512DivMod(num512, den512)
	return rem512.isZero(), nil
}

// MustIsMultipleOf is like [Fraction.IsMultipleOf] but panics on error.
func (f Fraction) MustIsMultipleOf(step Fraction) bool {
	ok, err := f.IsMultipleOf(step)
	if err != nil {
		panic(err)
	}
	return ok
}

// InProgression checks whether (f - start) / step is an exact integer k for
// positive step > 0, and returns k, true, nil when it is. It returns
// [ErrInvalid] if any operand is invalid, [ErrDivisionByZero] if step is zero,
// [ErrZeroDenominator] if step < 0, and [ErrNotRepresentable] if the integer
// step count k does not fit in int.
func (f Fraction) InProgression(start, step Fraction) (steps int, ok bool, err error) {
	if !f.IsValid() || !start.IsValid() || !step.IsValid() {
		return 0, false, ErrInvalid
	}
	if step.IsZero() {
		return 0, false, ErrDivisionByZero
	}
	if step.Sign() < 0 {
		return 0, false, ErrZeroDenominator
	}
	t1 := u256Mul(f.num, start.den)
	t2 := u256Mul(start.num, f.den)
	neg2 := start.Sign() < 0
	if !start.num.isZero() {
		neg2 = !neg2
	}
	diffNum, diffNeg := signedAddU512(t1, f.Sign() < 0, t2, neg2)
	if diffNum.isZero() {
		return 0, true, nil
	}
	diffDen := u256Mul(f.den, start.den)
	g := u512GCD(diffNum, diffDen)
	if !g.isOne() {
		diffNum, _ = u512DivMod(diffNum, g)
		diffDen, _ = u512DivMod(diffDen, g)
	}
	if diffNum.fitsU256Max() && diffDen.fitsU256Max() {
		num512 := u256Mul(diffNum.low256(), step.den)
		den512 := u256Mul(diffDen.low256(), step.num)
		q512, rem512 := u512DivMod(num512, den512)
		if !rem512.isZero() {
			return 0, false, nil
		}
		kInt, fits := u512ToInt(q512, diffNeg)
		if !fits {
			return 0, false, ErrNotRepresentable
		}
		return kInt, true, nil
	}
	// Extended precision fallback when reduced (f - start) exceeds 255 bits.
	bigNum := new(big.Int).Mul(diffNum.toBigInt(false), step.den.toBigInt(false))
	bigDen := new(big.Int).Mul(diffDen.toBigInt(false), step.num.toBigInt(false))
	quo, rem := new(big.Int).QuoRem(bigNum, bigDen, new(big.Int))
	if rem.Sign() != 0 {
		return 0, false, nil
	}
	if diffNeg {
		quo.Neg(quo)
	}
	if !quo.IsInt64() {
		return 0, false, ErrNotRepresentable
	}
	return int(quo.Int64()), true, nil
}

// MustInProgression is like [Fraction.InProgression] but panics on error.
func (f Fraction) MustInProgression(start, step Fraction) (steps int, ok bool) {
	s, ok, err := f.InProgression(start, step)
	if err != nil {
		panic(err)
	}
	return s, ok
}

func roundRatioU512(num, den u512, neg bool, mode RoundingMode) (qMag u512, qNeg bool) {
	if num.isZero() {
		return u512{}, false
	}
	q, rem := u512DivMod(num, den)
	if rem.isZero() {
		return q, neg
	}
	var inc bool
	switch mode {
	case RoundTrunc:
		inc = false
	case RoundFloor:
		inc = neg
	case RoundCeil:
		inc = !neg
	case RoundHalfAwayFromZero:
		inc = u512Cmp(rem, u512Sub(den, rem)) >= 0
	case RoundHalfToEven:
		diff := u512Sub(den, rem)
		c := u512Cmp(rem, diff)
		if c > 0 {
			inc = true
		} else if c == 0 {
			inc = (q[0] & 1) != 0
		}
	}
	if inc {
		q = u512Add(q, u512{1})
	}
	if q.isZero() {
		return u512{}, false
	}
	return q, neg
}

// RoundToInt rounds f to an int according to mode. It returns [ErrInvalid] if
// f is invalid and [ErrNotRepresentable] if mode is unknown or the rounded
// integer does not fit in int.
func (f Fraction) RoundToInt(mode RoundingMode) (int, error) {
	if !f.IsValid() {
		return 0, ErrInvalid
	}
	if mode > RoundTrunc {
		return 0, ErrNotRepresentable
	}
	qMag, qNeg := roundRatioU512(u512FromU256(f.num), u512FromU256(f.den), f.Sign() < 0, mode)
	res, ok := u512ToInt(qMag, qNeg)
	if !ok {
		return 0, ErrNotRepresentable
	}
	return res, nil
}

// MustRoundToInt is like [Fraction.RoundToInt] but panics on error.
func (f Fraction) MustRoundToInt(mode RoundingMode) int {
	v, err := f.RoundToInt(mode)
	if err != nil {
		panic(err)
	}
	return v
}

// RoundToBigInt rounds f to a *big.Int according to mode. It returns
// [ErrInvalid] if f is invalid and [ErrNotRepresentable] if mode is unknown.
func (f Fraction) RoundToBigInt(mode RoundingMode) (*big.Int, error) {
	if !f.IsValid() {
		return nil, ErrInvalid
	}
	if mode > RoundTrunc {
		return nil, ErrNotRepresentable
	}
	qMag, qNeg := roundRatioU512(u512FromU256(f.num), u512FromU256(f.den), f.Sign() < 0, mode)
	return qMag.toBigInt(qNeg), nil
}

// MustRoundToBigInt is like [Fraction.RoundToBigInt] but panics on error.
func (f Fraction) MustRoundToBigInt(mode RoundingMode) *big.Int {
	v, err := f.RoundToBigInt(mode)
	if err != nil {
		panic(err)
	}
	return v
}

func (f Fraction) roundToTargetDenU256(targetDen u256, mode RoundingMode) (Fraction, error) {
	num512 := u256Mul(f.num, targetDen)
	den512 := u512FromU256(f.den)
	mMag, mNeg := roundRatioU512(num512, den512, f.Sign() < 0, mode)
	if !mMag.fitsU256Max() {
		return Fraction{}, ErrOutOfRange
	}
	return makeFraction(mMag.low256(), targetDen, mNeg, false), nil
}

// RoundToDenominator rounds f to the nearest multiple of 1/den (den > 0)
// according to mode, returning a Fraction with stored denominator den. It
// returns [ErrInvalid] if f is invalid, [ErrZeroDenominator] if den <= 0,
// [ErrNotRepresentable] if mode is unknown, and [ErrOutOfRange] if the rounded
// numerator exceeds [MaxComponent].
func (f Fraction) RoundToDenominator(den int, mode RoundingMode) (Fraction, error) {
	if !f.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if den <= 0 {
		return Fraction{}, ErrZeroDenominator
	}
	if mode > RoundTrunc {
		return Fraction{}, ErrNotRepresentable
	}
	targetDen, _ := u256FromInt(den)
	return f.roundToTargetDenU256(targetDen, mode)
}

// MustRoundToDenominator is like [Fraction.RoundToDenominator] but panics on
// error.
func (f Fraction) MustRoundToDenominator(den int, mode RoundingMode) Fraction {
	res, err := f.RoundToDenominator(den, mode)
	if err != nil {
		panic(err)
	}
	return res
}

// RoundToBigDenominator rounds f to the nearest multiple of 1/den (den > 0)
// according to mode, returning a Fraction with stored denominator den. It
// returns [ErrInvalid] if f is invalid, [ErrZeroDenominator] if den == nil or
// den.Sign() <= 0, [ErrNotRepresentable] if mode is unknown, and
// [ErrOutOfRange] if den or the rounded numerator exceeds [MaxComponent].
func (f Fraction) RoundToBigDenominator(den *big.Int, mode RoundingMode) (Fraction, error) {
	if !f.IsValid() {
		return Fraction{}, ErrInvalid
	}
	if den == nil || den.Sign() <= 0 {
		return Fraction{}, ErrZeroDenominator
	}
	if mode > RoundTrunc {
		return Fraction{}, ErrNotRepresentable
	}
	targetDen, _, ok := u256FromBigInt(den)
	if !ok {
		return Fraction{}, ErrOutOfRange
	}
	return f.roundToTargetDenU256(targetDen, mode)
}

// MustRoundToBigDenominator is like [Fraction.RoundToBigDenominator] but
// panics on error.
func (f Fraction) MustRoundToBigDenominator(den *big.Int, mode RoundingMode) Fraction {
	res, err := f.RoundToBigDenominator(den, mode)
	if err != nil {
		panic(err)
	}
	return res
}

// Float64 converts f to the nearest IEEE-754 float64 value and reports whether
// the conversion is exact. For an invalid Fraction{}, it returns (NaN, false).
func (f Fraction) Float64() (val float64, exact bool) {
	if !f.IsValid() {
		return math.NaN(), false
	}
	r, _ := f.BigRat()
	return r.Float64()
}
