package tempomap

import (
	"math/big"

	"github.com/go-muse/musex/common/rounding"
)

// logTerm is one coefficient times the natural logarithm of a positive base.
type logTerm struct {
	coefficient *big.Rat
	base        *big.Rat
}

// logForm represents constant + sum(coefficient_i*ln(base_i)).
type logForm struct {
	constant *big.Rat
	terms    []logTerm
}

type roundingAttempt struct {
	lowerK  *big.Int
	upperK  *big.Int
	decided bool
	limit   uint
}

func refinementLimit(precision, maxPrecision uint) uint {
	var limit uint
	for precision < maxPrecision {
		if precision > maxPrecision/2 {
			precision = maxPrecision
		} else {
			precision *= 2
		}
		limit++
	}
	return limit
}

func makeRoundingAttempt(
	precision, maxPrecision uint,
	lower, upper *big.Rat,
	places int,
	mode rounding.Mode,
) roundingAttempt {
	return makeStepAttempt(precision, maxPrecision, lower, upper, decimalStep(places), mode)
}

// makeStepAttempt selects the rounding candidates of lower and upper on the
// grid k*step. It is makeRoundingAttempt with an arbitrary positive step.
func makeStepAttempt(
	precision, maxPrecision uint,
	lower, upper, step *big.Rat,
	mode rounding.Mode,
) roundingAttempt {
	lowerK, _, err := mode.RatMultiple(lower, step)
	if err != nil {
		panic(err)
	}
	upperK, _, err := mode.RatMultiple(upper, step)
	if err != nil {
		panic(err)
	}
	return roundingAttempt{
		lowerK:  lowerK,
		upperK:  upperK,
		decided: lowerK.Cmp(upperK) == 0,
		limit:   refinementLimit(precision, maxPrecision),
	}
}

func roundWithLimit(
	places int,
	mode rounding.Mode,
	initialPrecision, maxPrecision uint,
	eval func(precision uint) (lower, upper *big.Rat),
) (k *big.Int, usedPrecision, refinements uint, ok bool) {
	return roundStepWithLimit(decimalStep(places), mode, initialPrecision, maxPrecision, eval)
}

// roundStepWithLimit is roundWithLimit on the grid k*step.
func roundStepWithLimit(
	step *big.Rat,
	mode rounding.Mode,
	initialPrecision, maxPrecision uint,
	eval func(precision uint) (lower, upper *big.Rat),
) (k *big.Int, usedPrecision, refinements uint, ok bool) {
	return roundStepWithBound(step, mode, initialPrecision, new(big.Int).SetUint64(uint64(maxPrecision)), eval)
}

// roundStepWithBound refines only as far as needed. The theoretical bound
// may exceed uint even when the actual working precision is small.
func roundStepWithBound(
	step *big.Rat,
	mode rounding.Mode,
	initialPrecision uint,
	maxPrecision *big.Int,
	eval func(precision uint) (lower, upper *big.Rat),
) (k *big.Int, usedPrecision, refinements uint, ok bool) {
	precision := initialPrecision
	for {
		lower, upper := eval(precision)
		attempt := makeStepAttempt(precision, precision, lower, upper, step, mode)
		if attempt.decided {
			return new(big.Int).Set(attempt.lowerK), precision, refinements, true
		}
		current := new(big.Int).SetUint64(uint64(precision))
		if current.Cmp(maxPrecision) >= 0 {
			return nil, precision, refinements, false
		}
		next := current.Lsh(current, 1)
		if next.Cmp(maxPrecision) > 0 {
			next.Set(maxPrecision)
		}
		// The evaluator's resource contract requires actual working precisions
		// to fit uint; no conversion of the theoretical bound is performed.
		precision = uint(next.Uint64())
		refinements++
	}
}

func roundingNode(k *big.Int, places int) *big.Rat {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(places)), nil)
	return new(big.Rat).SetFrac(new(big.Int).Set(k), scale)
}

// stepNode returns k*step as a new value.
func stepNode(k *big.Int, step *big.Rat) *big.Rat {
	return new(big.Rat).Mul(new(big.Rat).SetInt(k), step)
}

func zeroLogForm() logForm { return logForm{constant: new(big.Rat)} }

func rationalLogForm(value *big.Rat) logForm {
	return logForm{constant: new(big.Rat).Set(value)}
}

func cloneLogForm(f logForm) logForm {
	out := logForm{constant: new(big.Rat).Set(f.constant), terms: make([]logTerm, len(f.terms))}
	for i, term := range f.terms {
		out.terms[i] = logTerm{coefficient: new(big.Rat).Set(term.coefficient), base: new(big.Rat).Set(term.base)}
	}
	return out
}

func logFormAdd(a, b logForm) logForm {
	out := zeroLogForm()
	out.constant.Add(a.constant, b.constant)
	out.terms = make([]logTerm, 0, len(a.terms)+len(b.terms))
	out.terms = append(out.terms, a.terms...)
	out.terms = append(out.terms, b.terms...)
	return out
}

func logFormScale(f logForm, factor *big.Rat) logForm {
	out := zeroLogForm()
	out.constant.Mul(f.constant, factor)
	out.terms = make([]logTerm, 0, len(f.terms))
	for _, term := range f.terms {
		out.terms = append(out.terms, logTerm{coefficient: new(big.Rat).Mul(term.coefficient, factor), base: term.base})
	}
	return out
}

func logFormSub(a, b logForm) logForm {
	return logFormAdd(a, logFormScale(b, big.NewRat(-1, 1)))
}

func logFormAddRat(f logForm, value *big.Rat) logForm {
	out := cloneLogForm(f)
	out.constant.Add(out.constant, value)
	return out
}

// logFormIsZero reports exact multiplicative cancellation without evaluating
// any logarithm. Non-nil pointers and positive bases are programming invariants.
func logFormIsZero(f logForm) bool {
	if f.constant.Sign() != 0 {
		return false
	}
	denominator := big.NewInt(1)
	active := make([]logTerm, 0, len(f.terms))
	for _, term := range f.terms {
		if term.coefficient.Sign() == 0 || term.base.Cmp(big.NewRat(1, 1)) == 0 {
			continue
		}
		active = append(active, term)
		gcd := new(big.Int).GCD(nil, nil, denominator, term.coefficient.Denom())
		denominator.Mul(denominator, new(big.Int).Quo(term.coefficient.Denom(), gcd))
	}
	if len(active) == 0 {
		return true
	}
	exponents := make(map[string]*big.Int)
	for _, term := range active {
		exponent := new(big.Rat).Mul(term.coefficient, new(big.Rat).SetInt(denominator)).Num()
		factorRational(new(big.Int).Set(term.base.Num()), exponent, exponents)
		factorRational(new(big.Int).Set(term.base.Denom()), new(big.Int).Neg(exponent), exponents)
	}
	for _, exponent := range exponents {
		if exponent.Sign() != 0 {
			return false
		}
	}
	return true
}

func factorRational(n, exponent *big.Int, totals map[string]*big.Int) {
	if n.Sign() == 0 || exponent.Sign() == 0 {
		return
	}
	n.Abs(n)
	if n.IsUint64() {
		factorUint64(n.Uint64(), exponent, totals)
		return
	}
	for p := big.NewInt(2); new(big.Int).Mul(p, p).Cmp(n) <= 0; p.Add(p, big.NewInt(1)) {
		for {
			q, r := new(big.Int), new(big.Int)
			q.QuoRem(n, p, r)
			if r.Sign() != 0 {
				break
			}
			addPrimeExponent(totals, p.String(), exponent)
			n = q
		}
	}
	if n.Cmp(big.NewInt(1)) > 0 {
		addPrimeExponent(totals, n.String(), exponent)
	}
}

func factorUint64(n uint64, exponent *big.Int, totals map[string]*big.Int) {
	for p := uint64(2); p <= n/p; p++ {
		for n%p == 0 {
			addPrimeExponent(totals, new(big.Int).SetUint64(p).String(), exponent)
			n /= p
		}
	}
	if n > 1 {
		addPrimeExponent(totals, new(big.Int).SetUint64(n).String(), exponent)
	}
}

func addPrimeExponent(totals map[string]*big.Int, prime string, exponent *big.Int) {
	if current, ok := totals[prime]; ok {
		current.Add(current, exponent)
		return
	}
	totals[prime] = new(big.Int).Set(exponent)
}

func logarithmUpperBound(base *big.Rat) *big.Rat {
	a := new(big.Rat).Sub(base, big.NewRat(1, 1))
	a.Abs(a)
	b := new(big.Rat).Sub(big.NewRat(1, 1), new(big.Rat).Inv(base))
	b.Abs(b)
	if a.Cmp(b) >= 0 {
		return a
	}
	return b
}

func logFormUpper(f logForm) *big.Rat {
	u := new(big.Rat).Abs(f.constant)
	for _, term := range f.terms {
		coefficient := new(big.Rat).Abs(term.coefficient)
		coefficient.Mul(coefficient, logarithmUpperBound(term.base))
		u.Add(u, coefficient)
	}
	return u
}

func ratBitLength(r *big.Rat) int {
	n := new(big.Int).Abs(r.Num()).BitLen()
	d := r.Denom().BitLen()
	if n > d {
		return n
	}
	return d
}

func logFormBitLength(f logForm) int {
	bits := ratBitLength(f.constant) + 1
	for _, term := range f.terms {
		bits += ratBitLength(term.coefficient) + ratBitLength(term.base) + 1
	}
	return bits
}
