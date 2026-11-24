package timecalc

import (
	"math/big"

	"github.com/go-muse/musex/common/rounding"
)

type interval struct {
	lo *big.Float
	hi *big.Float
}

func directedFloat(precision uint, mode big.RoundingMode) *big.Float {
	return new(big.Float).SetPrec(precision).SetMode(mode)
}

func intervalFromRat(value *big.Rat, precision uint) interval {
	return interval{
		lo: directedFloat(precision, big.ToNegativeInf).SetRat(value),
		hi: directedFloat(precision, big.ToPositiveInf).SetRat(value),
	}
}

func intervalAdd(a, b interval, precision uint) interval {
	return interval{
		lo: directedFloat(precision, big.ToNegativeInf).Add(a.lo, b.lo),
		hi: directedFloat(precision, big.ToPositiveInf).Add(a.hi, b.hi),
	}
}

func minMaxFloat(values ...*big.Float) (lo, hi *big.Float) {
	lo, hi = values[0], values[0]
	for _, value := range values[1:] {
		if value.Cmp(lo) < 0 {
			lo = value
		}
		if value.Cmp(hi) > 0 {
			hi = value
		}
	}
	return lo, hi
}

func intervalMul(a, b interval, precision uint) interval {
	products := []*big.Float{
		directedFloat(precision, big.ToNegativeInf).Mul(a.lo, b.lo),
		directedFloat(precision, big.ToNegativeInf).Mul(a.lo, b.hi),
		directedFloat(precision, big.ToNegativeInf).Mul(a.hi, b.lo),
		directedFloat(precision, big.ToNegativeInf).Mul(a.hi, b.hi),
	}
	lo, _ := minMaxFloat(products...)
	products = []*big.Float{
		directedFloat(precision, big.ToPositiveInf).Mul(a.lo, b.lo),
		directedFloat(precision, big.ToPositiveInf).Mul(a.lo, b.hi),
		directedFloat(precision, big.ToPositiveInf).Mul(a.hi, b.lo),
		directedFloat(precision, big.ToPositiveInf).Mul(a.hi, b.hi),
	}
	_, hi := minMaxFloat(products...)
	return interval{lo: lo, hi: hi}
}

func intervalQuo(a, b interval, precision uint) interval {
	values := []*big.Float{
		directedFloat(precision, big.ToNegativeInf).Quo(a.lo, b.lo),
		directedFloat(precision, big.ToNegativeInf).Quo(a.lo, b.hi),
		directedFloat(precision, big.ToNegativeInf).Quo(a.hi, b.lo),
		directedFloat(precision, big.ToNegativeInf).Quo(a.hi, b.hi),
	}
	lo, _ := minMaxFloat(values...)
	values = []*big.Float{
		directedFloat(precision, big.ToPositiveInf).Quo(a.lo, b.lo),
		directedFloat(precision, big.ToPositiveInf).Quo(a.lo, b.hi),
		directedFloat(precision, big.ToPositiveInf).Quo(a.hi, b.lo),
		directedFloat(precision, big.ToPositiveInf).Quo(a.hi, b.hi),
	}
	_, hi := minMaxFloat(values...)
	return interval{lo: lo, hi: hi}
}

func intervalScale(a interval, factor *big.Rat, precision uint) interval {
	return intervalMul(a, intervalFromRat(factor, precision), precision)
}

// intervalWidenFrac widens a on both sides by the error bound num/den, which
// is rounded up to precision first. It requires num >= 0 and den > 0: both
// callers pass a nonnegative power over a positive denominator. As with
// big.Rat, a zero den panics with "division by zero".
//
// The fraction is deliberately not reduced. The series remainders passed here
// have operands of hundreds of thousands of bits, and reducing them with
// big.Rat costs a quadratic GCD while the rounded bound does not change.
// Under the precondition above, the bound is bit-identical to
// big.Float.SetRat of the reduced big.Rat num/den: big.Float.Quo rounds the
// exact quotient correctly, so a nonzero result depends only on the value of
// the fraction, and a zero num gives +0 in both. The precondition matters for
// zero only: 0 divided by a negative den would give -0, while SetRat gives +0.
func intervalWidenFrac(a interval, num, den *big.Int, precision uint) interval {
	if den.Sign() == 0 {
		panic("division by zero")
	}
	bound := directedFloat(precision, big.ToPositiveInf).Quo(new(big.Float).SetInt(num), new(big.Float).SetInt(den))
	return interval{
		lo: directedFloat(precision, big.ToNegativeInf).Sub(a.lo, bound),
		hi: directedFloat(precision, big.ToPositiveInf).Add(a.hi, bound),
	}
}

func ratOfFloat(value *big.Float) *big.Rat {
	r, _ := value.Rat(nil)
	return r
}

func intervalMaxAbs(a interval) *big.Rat {
	lo := new(big.Rat).Abs(ratOfFloat(a.lo))
	hi := new(big.Rat).Abs(ratOfFloat(a.hi))
	if lo.Cmp(hi) >= 0 {
		return lo
	}
	return hi
}

func intervalLogForm(form logForm, precision uint) interval {
	result := intervalFromRat(form.constant, precision)
	for _, term := range form.terms {
		logarithm := intervalLogRat(term.base, precision)
		result = intervalAdd(result, intervalScale(logarithm, term.coefficient, precision), precision)
	}
	return result
}

func pow2Rat(exponent int) *big.Rat {
	if exponent >= 0 {
		return new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(exponent)))
	}
	return new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), uint(-exponent)))
}

func decomposeLogBase(base *big.Rat) (int, *big.Rat) {
	k := base.Num().BitLen() - base.Denom().BitLen()
	u := new(big.Rat).Mul(base, pow2Rat(-k))
	for {
		twiceSquare := new(big.Rat).Mul(u, u)
		twiceSquare.Mul(twiceSquare, big.NewRat(2, 1))
		if twiceSquare.Cmp(big.NewRat(1, 1)) >= 0 {
			break
		}
		u.Mul(u, big.NewRat(2, 1))
		k--
	}
	for new(big.Rat).Mul(u, u).Cmp(big.NewRat(2, 1)) >= 0 {
		u.Quo(u, big.NewRat(2, 1))
		k++
	}
	return k, u
}

func intervalAtanhLog(base *big.Rat, precision uint) interval {
	z := new(big.Rat).Sub(base, big.NewRat(1, 1))
	z.Quo(z, new(big.Rat).Add(base, big.NewRat(1, 1)))
	terms := 2*int(precision) + 32
	zInterval := intervalFromRat(z, precision)
	zSquared := intervalMul(zInterval, zInterval, precision)
	power := zInterval
	sum := zInterval
	for n := 1; n < terms; n++ {
		power = intervalMul(power, zSquared, precision)
		term := intervalQuo(power, intervalFromRat(big.NewRat(int64(2*n+1), 1), precision), precision)
		sum = intervalAdd(sum, term, precision)
	}
	// The remainder of the doubled series is bounded by
	// 2|z|^(2T+1) / ((2T+1)(1-z^2)) for T terms. With z = p/q, multiplying
	// the numerator and the denominator by q^(2T+1) keeps everything in
	// integers: 2|p|^(2T+1) / ((2T+1) q^(2T-1) (q^2-p^2)).
	p, q := z.Num(), z.Denom()
	exponent := big.NewInt(int64(2*terms + 1))
	remainderNum := new(big.Int).Exp(new(big.Int).Abs(p), exponent, nil)
	remainderNum.Lsh(remainderNum, 1)
	remainderDen := new(big.Int).Exp(q, big.NewInt(int64(2*terms-1)), nil)
	remainderDen.Mul(remainderDen, new(big.Int).Sub(new(big.Int).Mul(q, q), new(big.Int).Mul(p, p)))
	remainderDen.Mul(remainderDen, exponent)
	sum = intervalScale(sum, big.NewRat(2, 1), precision)
	return intervalWidenFrac(sum, remainderNum, remainderDen, precision)
}

func intervalLogRat(base *big.Rat, precision uint) interval {
	k, u := decomposeLogBase(base)
	logU := intervalAtanhLog(u, precision)
	if k == 0 {
		return logU
	}
	log2 := intervalAtanhLog(big.NewRat(2, 1), precision)
	log2 = intervalScale(log2, new(big.Rat).SetInt64(int64(k)), precision)
	return intervalAdd(log2, logU, precision)
}

func intervalExp(value interval, precision uint) interval {
	maxAbs := intervalMaxAbs(value)
	j := 0
	if maxAbs.Sign() > 0 {
		j = ceilLog2Rat(maxAbs)
		if j < 0 {
			j = 0
		}
	}
	y := intervalScale(value, pow2Rat(-j), precision)
	yAbs := new(big.Rat).Mul(maxAbs, pow2Rat(-j))
	terms := int(precision) + 32
	term := y
	sum := intervalAdd(intervalFromRat(big.NewRat(1, 1), precision), y, precision)
	for n := 2; n <= terms; n++ {
		term = intervalMul(term, y, precision)
		term = intervalQuo(term, intervalFromRat(new(big.Rat).SetInt64(int64(n)), precision), precision)
		sum = intervalAdd(sum, term, precision)
	}
	// The remainder after T terms is bounded by 2|y|^(T+1)/(T+1)!; with
	// |y| = n/d this is 2n^(T+1) / ((T+1)! d^(T+1)).
	exponent := big.NewInt(int64(terms + 1))
	remainderNum := new(big.Int).Exp(yAbs.Num(), exponent, nil)
	remainderNum.Lsh(remainderNum, 1)
	remainderDen := new(big.Int).Exp(yAbs.Denom(), exponent, nil)
	remainderDen.Mul(remainderDen, new(big.Int).MulRange(1, int64(terms+1)))
	result := intervalWidenFrac(sum, remainderNum, remainderDen, precision)
	for range j {
		result = intervalMul(result, result, precision)
	}
	return result
}

type expForm struct {
	a, b   *big.Rat
	x      logForm
	minExp *big.Rat
	maxExp *big.Rat
}

func (e expForm) interval(precision uint) interval {
	x := intervalLogForm(e.x, precision)
	exponential := intervalExp(x, precision)
	exponential = intervalAdd(exponential, intervalFromRat(big.NewRat(-1, 1), precision), precision)
	exponential = intervalScale(exponential, e.b, precision)
	return intervalAdd(intervalFromRat(e.a, precision), exponential, precision)
}

func (e expForm) upperBound(y *big.Rat) *big.Rat {
	u := new(big.Rat).Abs(e.a)
	u.Add(u, new(big.Rat).Abs(y))
	term := new(big.Rat).Add(e.maxExp, big.NewRat(1, 1))
	term.Mul(term, new(big.Rat).Abs(e.b))
	return u.Add(u, term)
}

func (e expForm) inputBits() int {
	return ratBitLength(e.a) + ratBitLength(e.b) + logFormBitLength(e.x) + 1
}

func (e expForm) exponentialJ() int {
	u := logFormUpper(e.x)
	if u.Sign() == 0 {
		return 0
	}
	j := ceilLog2Rat(u)
	if j < 0 {
		return 0
	}
	return j
}

func (e expForm) comparisonR(y *big.Rat) *big.Rat {
	r := new(big.Rat).Sub(y, e.a)
	r.Add(r, e.b)
	return r.Quo(r, e.b)
}

func (e expForm) delta(y *big.Rat) logForm {
	r := e.comparisonR(y)
	return logFormAdd(e.x, logForm{constant: new(big.Rat), terms: []logTerm{{coefficient: big.NewRat(-1, 1), base: r}}})
}

func (e expForm) isZeroAt(y *big.Rat) bool {
	r := e.comparisonR(y)
	return r.Sign() > 0 && logFormIsZero(e.delta(y))
}

func (e expForm) separationLog2(y *big.Rat) *big.Rat {
	r := e.comparisonR(y)
	if r.Sign() <= 0 {
		bound := new(big.Rat).Sub(e.minExp, r)
		bound.Mul(bound, new(big.Rat).Abs(e.b))
		return new(big.Rat).SetInt64(int64(ceilLog2Rat(new(big.Rat).Inv(bound))))
	}
	delta := e.delta(y)
	h := new(big.Rat).Mul(big.NewRat(2, 1), logSeparation(delta))
	factor := new(big.Rat).Mul(new(big.Rat).Abs(e.b), r)
	return h.Add(h, new(big.Rat).Sub(big.NewRat(1, 1), new(big.Rat).SetInt64(int64(floorLog2Rat(factor)))))
}

type numericExpression interface {
	interval(precision uint) interval
	upperBound(y *big.Rat) *big.Rat
	inputBits() int
	exponentialJ() int
	isZeroAt(y *big.Rat) bool
	separationLog2(y *big.Rat) *big.Rat
}

type logExpression struct{ form logForm }

func (e logExpression) interval(precision uint) interval { return intervalLogForm(e.form, precision) }

func (e logExpression) upperBound(y *big.Rat) *big.Rat {
	shifted := logFormSub(e.form, rationalLogForm(y))
	return logFormUpper(shifted)
}

func (e logExpression) inputBits() int    { return logFormBitLength(e.form) }
func (e logExpression) exponentialJ() int { return 0 }
func (e logExpression) isZeroAt(y *big.Rat) bool {
	return logFormIsZero(logFormSub(e.form, rationalLogForm(y)))
}
func (e logExpression) separationLog2(y *big.Rat) *big.Rat {
	return new(big.Rat).Mul(big.NewRat(2, 1), logSeparation(logFormSub(e.form, rationalLogForm(y))))
}

func ceilLog2Rat(value *big.Rat) int {
	if value.Sign() <= 0 {
		panic("ceilLog2Rat requires a positive value")
	}
	k := value.Num().BitLen() - value.Denom().BitLen()
	for pow2Rat(k).Cmp(value) < 0 {
		k++
	}
	for k > -int(^uint(0)>>1) && pow2Rat(k-1).Cmp(value) >= 0 {
		k--
	}
	return k
}

func floorLog2Rat(value *big.Rat) int {
	if value.Sign() <= 0 {
		panic("floorLog2Rat requires a positive value")
	}
	k := value.Num().BitLen() - value.Denom().BitLen()
	for pow2Rat(k).Cmp(value) > 0 {
		k--
	}
	for pow2Rat(k+1).Cmp(value) <= 0 {
		k++
	}
	return k
}

func ceilRat(value *big.Rat) int {
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(value.Num(), value.Denom(), r)
	if r.Sign() > 0 {
		q.Add(q, big.NewInt(1))
	}
	if !q.IsInt64() {
		panic("rational ceiling does not fit int64")
	}
	return int(q.Int64())
}

func ceilRatPrecision(value *big.Rat) *big.Int {
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(value.Num(), value.Denom(), r)
	if r.Sign() > 0 {
		q.Add(q, big.NewInt(1))
	}
	return q
}

func initialPrecision(places, bits int) uint {
	return uint(128 + 4*(places+4) + 2*bits)
}

func precisionLimit(expression numericExpression, y *big.Rat, places int) *big.Int {
	return precisionLimitFrom(expression, y, initialPrecision(places, expression.inputBits()))
}

// precisionLimitFrom is the computable Pmax of a comparison with y whose
// starting precision is initial.
func precisionLimitFrom(expression numericExpression, y *big.Rat, initial uint) *big.Int {
	u := expression.upperBound(y)
	onePlusU := new(big.Rat).Add(big.NewRat(1, 1), u)
	logU := ceilLog2Rat(onePlusU)
	l := expression.separationLog2(y)
	l2 := ceilRatPrecision(new(big.Rat).Mul(big.NewRat(2, 1), new(big.Rat).Add(big.NewRat(int64(2+logU), 1), l)))
	c := uint(expression.exponentialJ() + 80)
	limit := new(big.Int).SetUint64(uint64(max(initial, c, 4)))
	candidate := new(big.Int).Add(l2, big.NewInt(64))
	if candidate.Cmp(limit) > 0 {
		limit.Set(candidate)
	}
	return limit
}

// stepInitialPrecision is the starting precision for rounding to multiples of
// step: 128 + max(0, ceil(log2(10^4/step))) + 2B, where B adds the bit lengths
// of the numerator and the denominator of step to the input bits.
func stepInitialPrecision(step *big.Rat, inputBits int) uint {
	scale := new(big.Rat).Quo(big.NewRat(10000, 1), step)
	extra := max(0, ceilLog2Rat(scale))
	bits := inputBits + step.Num().BitLen() + step.Denom().BitLen()
	return uint(128 + extra + 2*bits)
}

func decimalStep(places int) *big.Rat {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	return new(big.Rat).SetFrac(big.NewInt(1), scale)
}

func roundExpression(expression numericExpression, places int, mode rounding.Mode) *big.Rat {
	initial := initialPrecision(places, expression.inputBits())
	k := roundExpressionSteps(expression, decimalStep(places), mode, initial)
	return roundingNode(k, places)
}

// roundExpressionSteps returns the index k of the multiple k*step that mode
// selects from the exact value of expression, starting at precision initial.
func roundExpressionSteps(expression numericExpression, step *big.Rat, mode rounding.Mode, initial uint) *big.Int {
	precision := initial
	eval := func(p uint) (*big.Rat, *big.Rat) {
		value := expression.interval(p)
		return ratOfFloat(value.lo), ratOfFloat(value.hi)
	}
	for {
		lower, upper := eval(precision)
		attempt := makeStepAttempt(precision, precision, lower, upper, step, mode)
		if attempt.decided {
			return attempt.lowerK
		}
		lowerNode := stepNode(attempt.lowerK, step)
		upperNode := stepNode(attempt.upperK, step)
		if expression.isZeroAt(lowerNode) {
			return attempt.lowerK
		}
		if expression.isZeroAt(upperNode) {
			return attempt.upperK
		}

		var midpoint *big.Rat
		if !mode.IsDirected() && new(big.Rat).Sub(upperNode, lowerNode).Cmp(step) == 0 {
			midpoint = new(big.Rat).Add(lowerNode, upperNode)
			midpoint.Quo(midpoint, big.NewRat(2, 1))
			if expression.isZeroAt(midpoint) {
				midpointAttempt := makeStepAttempt(precision, precision, midpoint, midpoint, step, mode)
				return midpointAttempt.lowerK
			}
		}

		maxPrecision := precisionLimitFrom(expression, lowerNode, initial)
		if upperLimit := precisionLimitFrom(expression, upperNode, initial); upperLimit.Cmp(maxPrecision) > 0 {
			maxPrecision = upperLimit
		}
		if midpoint != nil {
			if midpointLimit := precisionLimitFrom(expression, midpoint, initial); midpointLimit.Cmp(maxPrecision) > 0 {
				maxPrecision = midpointLimit
			}
		}
		if new(big.Int).SetUint64(uint64(precision)).Cmp(maxPrecision) >= 0 {
			panic("tempomap: certified precision limit did not select a rounding node")
		}
		k, usedPrecision, _, ok := roundStepWithBound(step, mode, precision, maxPrecision, eval)
		if ok {
			return k
		}
		precision = usedPrecision
	}
}

func compareLogForm(form logForm, y *big.Rat) int {
	if logFormIsZero(logFormSub(form, rationalLogForm(y))) {
		return 0
	}
	expression := logExpression{form: form}
	precision := initialPrecision(0, expression.inputBits())
	var limit *big.Int
	for {
		value := expression.interval(precision)
		if ratOfFloat(value.hi).Cmp(y) < 0 {
			return -1
		}
		if ratOfFloat(value.lo).Cmp(y) > 0 {
			return 1
		}
		if limit == nil {
			limit = precisionLimit(expression, y, 0)
		}
		if new(big.Int).SetUint64(uint64(precision)).Cmp(limit) >= 0 {
			panic("tempomap: certified precision limit did not determine a comparison")
		}
		next := new(big.Int).Lsh(new(big.Int).SetUint64(uint64(precision)), 1)
		if next.Cmp(limit) > 0 {
			next.Set(limit)
		}
		precision = uint(next.Uint64())
	}
}

func logSeparation(form logForm) *big.Rat {
	active := make([]logTerm, 0, len(form.terms))
	for _, term := range form.terms {
		if term.coefficient.Sign() != 0 && term.base.Cmp(big.NewRat(1, 1)) != 0 {
			active = append(active, term)
		}
	}
	if len(active) == 0 {
		absolute := new(big.Rat).Abs(form.constant)
		return new(big.Rat).SetInt64(int64(ceilLog2Rat(new(big.Rat).Inv(absolute))))
	}
	type parameter struct {
		v int
	}
	parameters := make([]parameter, len(active))
	for i, term := range active {
		u := logarithmUpperBound(term.base)
		ceilU := 1
		if u.Cmp(big.NewRat(1, 1)) > 0 {
			ceilU = ceilRat(u)
		}
		component := term.base.Num().BitLen()
		if term.base.Denom().BitLen() > component {
			component = term.base.Denom().BitLen()
		}
		v := max(1, component, ceilU)
		parameters[i] = parameter{v: v}
	}
	for i := 1; i < len(parameters); i++ {
		for j := i; j > 0 && parameters[j].v < parameters[j-1].v; j-- {
			parameters[j], parameters[j-1] = parameters[j-1], parameters[j]
		}
	}
	w := 1
	for _, term := range active {
		w = max(w, ratBitLength(term.coefficient))
	}
	w = max(w, ratBitLength(form.constant))
	n := len(active)
	h := new(big.Int).Lsh(big.NewInt(1), uint(6*n+51))
	nPower := new(big.Int).Exp(big.NewInt(int64(n)), big.NewInt(int64(2*n)), nil)
	h.Mul(h, nPower)
	productV := big.NewInt(1)
	for _, parameter := range parameters {
		productV.Mul(productV, big.NewInt(int64(parameter.v)))
	}
	h.Mul(h, productV)
	last := parameters[n-1].v
	firstLog := new(big.Int).Add(big.NewInt(int64(w)), big.NewInt(int64(big.NewInt(int64(2*last)).BitLen())))
	h.Mul(h, firstLog)
	previous := 1
	if n > 1 {
		previous = parameters[n-2].v
	}
	h.Mul(h, big.NewInt(int64(big.NewInt(int64(2*previous)).BitLen())))
	h.Mul(h, new(big.Int).Lsh(big.NewInt(1), uint(n+1)))
	return new(big.Rat).SetInt(h)
}
