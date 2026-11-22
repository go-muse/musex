// Package timecalc implements the shared exact time arithmetic of the module.
// It has no domain objects or limits on the size of intermediate rationals.
package timecalc

import (
	"math/big"

	"github.com/go-muse/musex/common/rounding"
)

// LogTerm is Coefficient times the natural logarithm of positive Base.
type LogTerm struct{ Coefficient, Base *big.Rat }

// LogForm is Constant plus the sum of Terms. Its zero value is not valid.
// All operations copy inputs; callers must prevent concurrent input mutation.
type LogForm struct {
	Constant *big.Rat
	Terms    []LogTerm
}

func privateForm(f LogForm) logForm {
	out := logForm{constant: new(big.Rat).Set(f.Constant), terms: make([]logTerm, len(f.Terms))}
	for i, t := range f.Terms {
		out.terms[i] = logTerm{coefficient: new(big.Rat).Set(t.Coefficient), base: new(big.Rat).Set(t.Base)}
	}
	return out
}
func publicForm(f logForm) LogForm {
	out := LogForm{Constant: new(big.Rat).Set(f.constant), Terms: make([]LogTerm, len(f.terms))}
	for i, t := range f.terms {
		out.Terms[i] = LogTerm{Coefficient: new(big.Rat).Set(t.coefficient), Base: new(big.Rat).Set(t.base)}
	}
	return out
}

// Constant creates an independent rational form.
func Constant(value *big.Rat) LogForm { return publicForm(rationalLogForm(value)) }

// Add creates the exact sum without retaining either input.
func Add(a, b LogForm) LogForm { return publicForm(logFormAdd(privateForm(a), privateForm(b))) }

// Scale multiplies the entire form by factor, including its constant.
func Scale(f LogForm, factor *big.Rat) LogForm {
	return publicForm(logFormScale(privateForm(f), factor))
}

// IsZero proves exact cancellation with multiplicative logarithmic identities.
func IsZero(f LogForm) bool { return logFormIsZero(privateForm(f)) }

// Rational returns a new exact rational when all logarithmic terms cancel.
func Rational(f LogForm) (*big.Rat, bool) {
	p := privateForm(f)
	c := new(big.Rat).Set(p.constant)
	p.constant = new(big.Rat)
	if logFormIsZero(p) {
		return c, true
	}
	return nil, false
}

// LinearSpan integrates 60/s(x) between from and to on a linear-speed
// segment [a,b]. Inputs must satisfy 0<=a<b, a<=from<=to<=b and positive speeds.
func LinearSpan(a, b, startSpeed, endSpeed, from, to *big.Rat) LogForm {
	if from.Cmp(to) == 0 {
		return Constant(new(big.Rat))
	}
	if startSpeed.Cmp(endSpeed) == 0 {
		return Constant(new(big.Rat).Quo(new(big.Rat).Mul(big.NewRat(60, 1), new(big.Rat).Sub(to, from)), startSpeed))
	}
	k := new(big.Rat).Quo(new(big.Rat).Sub(endSpeed, startSpeed), new(big.Rat).Sub(b, a))
	speed := func(x *big.Rat) *big.Rat {
		return new(big.Rat).Add(startSpeed, new(big.Rat).Mul(k, new(big.Rat).Sub(x, a)))
	}
	return LogForm{Constant: new(big.Rat), Terms: []LogTerm{{Coefficient: new(big.Rat).Quo(big.NewRat(60, 1), k), Base: new(big.Rat).Quo(speed(to), speed(from))}}}
}

// RoundLogSteps rounds a nonnegative form once to a positive rational step.
// The returned index has no machine-integer limit.
func RoundLogSteps(f LogForm, step *big.Rat, mode rounding.Mode) *big.Int {
	if r, ok := Rational(f); ok {
		k, _, err := mode.RatMultiple(r, step)
		if err != nil {
			panic(err)
		}
		return k
	}
	e := logExpression{form: privateForm(f)}
	return roundExpressionSteps(e, step, mode, stepInitialPrecision(step, e.inputBits()))
}

// CompareLog returns the sign of f-value, proving equality symbolically.
func CompareLog(f LogForm, value *big.Rat) int {
	if r, ok := Rational(f); ok {
		return r.Cmp(value)
	}
	return compareLogForm(privateForm(f), value)
}

// DecimalInterval encloses a nonnegative form by the HalfUp decimal result
// plus or minus half a decimal unit, or two independent exact rational copies.
func DecimalInterval(f LogForm, places int) (lo, hi *big.Rat, exact bool) {
	if r, ok := Rational(f); ok {
		return r, new(big.Rat).Set(r), true
	}
	h := decimalStep(places)
	v := stepNode(RoundLogSteps(f, h, rounding.HalfUp), h)
	half := new(big.Rat).Quo(h, big.NewRat(2, 1))
	return new(big.Rat).Sub(v, half), new(big.Rat).Add(v, half), false
}

// Expression wraps the shared logarithmic or exponential evaluation. It is
// initialized by Logarithmic or Exponential; its zero value is not valid.
type Expression struct{ value numericExpression }

// Logarithmic creates an independent expression for a log form.
func Logarithmic(f LogForm) Expression { return Expression{value: logExpression{form: privateForm(f)}} }

// Exponential creates a+b*(exp(x)-1), with positive bounds on exp(x).
func Exponential(a, b *big.Rat, x LogForm, minExp, maxExp *big.Rat) Expression {
	return Expression{value: expForm{a: new(big.Rat).Set(a), b: new(big.Rat).Set(b), x: privateForm(x), minExp: new(big.Rat).Set(minExp), maxExp: new(big.Rat).Set(maxExp)}}
}

// StepInitialPrecision preserves the initial-precision policy of tempomap.
func StepInitialPrecision(step *big.Rat, bits int) uint { return stepInitialPrecision(step, bits) }

// RoundExpression preserves decimal rounding and its initial precision.
func RoundExpression(e Expression, places int, mode rounding.Mode) *big.Rat {
	return roundExpression(e.value, places, mode)
}

// RoundExpressionSteps rounds from the supplied initial working precision.
func RoundExpressionSteps(e Expression, step *big.Rat, mode rounding.Mode, initial uint) *big.Int {
	return roundExpressionSteps(e.value, step, mode, initial)
}

// Attempt is the observable state of a legacy rounding-seam attempt.
type Attempt struct {
	LowerK, UpperK *big.Int
	Decided        bool
	Limit          uint
}

func publicAttempt(a roundingAttempt) Attempt {
	return Attempt{LowerK: a.lowerK, UpperK: a.upperK, Decided: a.decided, Limit: a.limit}
}

// MakeRoundingAttempt preserves the decimal rounding seam.
func MakeRoundingAttempt(p, max uint, lo, hi *big.Rat, places int, mode rounding.Mode) Attempt {
	return publicAttempt(makeRoundingAttempt(p, max, lo, hi, places, mode))
}

// MakeStepAttempt preserves the step rounding seam.
func MakeStepAttempt(p, max uint, lo, hi, step *big.Rat, mode rounding.Mode) Attempt {
	return publicAttempt(makeStepAttempt(p, max, lo, hi, step, mode))
}

// RoundWithLimit preserves the decimal refinement seam and its counters.
func RoundWithLimit(places int, mode rounding.Mode, initial, max uint, eval func(uint) (*big.Rat, *big.Rat)) (*big.Int, uint, uint, bool) {
	return roundWithLimit(places, mode, initial, max, eval)
}

// RoundStepWithLimit preserves the step refinement seam and its counters.
func RoundStepWithLimit(step *big.Rat, mode rounding.Mode, initial, max uint, eval func(uint) (*big.Rat, *big.Rat)) (*big.Int, uint, uint, bool) {
	return roundStepWithLimit(step, mode, initial, max, eval)
}

// RoundStepWithBound preserves a theoretical bound wider than uint.
func RoundStepWithBound(step *big.Rat, mode rounding.Mode, initial uint, max *big.Int, eval func(uint) (*big.Rat, *big.Rat)) (*big.Int, uint, uint, bool) {
	return roundStepWithBound(step, mode, initial, max, eval)
}

// InputBits is the expression size used by the existing precision policy.
func (e Expression) InputBits() int { return e.value.inputBits() }
