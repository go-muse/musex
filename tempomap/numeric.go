package tempomap

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

func intervalWiden(a interval, errorBound *big.Rat, precision uint) interval {
	bound := intervalFromRat(errorBound, precision)
	return interval{
		lo: directedFloat(precision, big.ToNegativeInf).Sub(a.lo, bound.hi),
		hi: directedFloat(precision, big.ToPositiveInf).Add(a.hi, bound.hi),
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

func integerPowerRat(base *big.Rat, exponent int) *big.Rat {
	n := new(big.Int).Exp(base.Num(), big.NewInt(int64(exponent)), nil)
	d := new(big.Int).Exp(base.Denom(), big.NewInt(int64(exponent)), nil)
	return new(big.Rat).SetFrac(n, d)
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
	zAbs := new(big.Rat).Abs(z)
	remainderPower := integerPowerRat(zAbs, 2*terms+1)
	denominator := new(big.Rat).Mul(big.NewRat(int64(2*terms+1), 1), new(big.Rat).Sub(big.NewRat(1, 1), new(big.Rat).Mul(z, z)))
	remainder := new(big.Rat).Quo(remainderPower, denominator)
	remainder.Mul(remainder, big.NewRat(2, 1))
	sum = intervalScale(sum, big.NewRat(2, 1), precision)
	return intervalWiden(sum, remainder, precision)
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
	factorial := new(big.Int).MulRange(1, int64(terms+1))
	remainder := new(big.Rat).SetFrac(new(big.Int).Mul(big.NewInt(2), integerPowerRat(yAbs, terms+1).Num()), new(big.Int).Mul(factorial, integerPowerRat(yAbs, terms+1).Denom()))
	result := intervalWiden(sum, remainder, precision)
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
	h := logSeparation(delta)
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
	return logSeparation(logFormSub(e.form, rationalLogForm(y)))
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

func ceilRatPrecision(value *big.Rat) uint {
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(value.Num(), value.Denom(), r)
	if r.Sign() > 0 {
		q.Add(q, big.NewInt(1))
	}
	if !q.IsUint64() {
		panic("precision bound does not fit uint")
	}
	return uint(q.Uint64())
}

func initialPrecision(places, bits int) uint {
	return uint(128 + 4*(places+4) + 2*bits)
}

func precisionLimit(expression numericExpression, y *big.Rat, places int) uint {
	u := expression.upperBound(y)
	onePlusU := new(big.Rat).Add(big.NewRat(1, 1), u)
	logU := ceilLog2Rat(onePlusU)
	l := expression.separationLog2(y)
	l2 := ceilRatPrecision(new(big.Rat).Mul(big.NewRat(2, 1), new(big.Rat).Add(big.NewRat(int64(2+logU), 1), l)))
	c := uint(expression.exponentialJ() + 80)
	limit := initialPrecision(places, expression.inputBits())
	for _, candidate := range []uint{c, 4, l2 + 64} {
		if candidate > limit {
			limit = candidate
		}
	}
	return limit
}

func decimalStep(places int) *big.Rat {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	return new(big.Rat).SetFrac(big.NewInt(1), scale)
}

func roundExpression(expression numericExpression, places int, mode rounding.Mode) *big.Rat {
	step := decimalStep(places)
	precision := initialPrecision(places, expression.inputBits())
	var limit uint
	for {
		value := expression.interval(precision)
		_, lower, err := mode.RatMultiple(ratOfFloat(value.lo), step)
		if err != nil {
			panic(err)
		}
		_, upper, err := mode.RatMultiple(ratOfFloat(value.hi), step)
		if err != nil {
			panic(err)
		}
		if lower.Cmp(upper) == 0 {
			return lower
		}
		if expression.isZeroAt(lower) {
			return lower
		}
		if expression.isZeroAt(upper) {
			return upper
		}
		if limit == 0 {
			lowerLimit := precisionLimit(expression, lower, places)
			upperLimit := precisionLimit(expression, upper, places)
			limit = lowerLimit
			if upperLimit > limit {
				limit = upperLimit
			}
		}
		if precision >= limit {
			panic("tempomap: certified precision limit did not select a rounding node")
		}
		precision *= 2
	}
}

func compareLogForm(form logForm, y *big.Rat) int {
	if logFormIsZero(logFormSub(form, rationalLogForm(y))) {
		return 0
	}
	expression := logExpression{form: form}
	precision := initialPrecision(0, expression.inputBits())
	var limit uint
	for {
		value := expression.interval(precision)
		if ratOfFloat(value.hi).Cmp(y) < 0 {
			return -1
		}
		if ratOfFloat(value.lo).Cmp(y) > 0 {
			return 1
		}
		if limit == 0 {
			limit = precisionLimit(expression, y, 0)
		}
		if precision >= limit {
			panic("tempomap: certified precision limit did not determine a comparison")
		}
		precision *= 2
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
