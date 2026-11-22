package tempomap

import (
	"math/big"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/internal/timecalc"
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

func makeRoundingAttempt(p, max uint, lower, upper *big.Rat, places int, mode rounding.Mode) roundingAttempt {
	return localAttempt(timecalc.MakeRoundingAttempt(p, max, lower, upper, places, mode))
}
func makeStepAttempt(p, max uint, lower, upper, step *big.Rat, mode rounding.Mode) roundingAttempt {
	return localAttempt(timecalc.MakeStepAttempt(p, max, lower, upper, step, mode))
}
func localAttempt(a timecalc.Attempt) roundingAttempt {
	return roundingAttempt{lowerK: a.LowerK, upperK: a.UpperK, decided: a.Decided, limit: a.Limit}
}
func roundWithLimit(places int, mode rounding.Mode, initial, max uint, eval func(uint) (*big.Rat, *big.Rat)) (*big.Int, uint, uint, bool) {
	return timecalc.RoundWithLimit(places, mode, initial, max, eval)
}
func roundStepWithLimit(step *big.Rat, mode rounding.Mode, initial, max uint, eval func(uint) (*big.Rat, *big.Rat)) (*big.Int, uint, uint, bool) {
	return timecalc.RoundStepWithLimit(step, mode, initial, max, eval)
}
func roundStepWithBound(step *big.Rat, mode rounding.Mode, initial uint, max *big.Int, eval func(uint) (*big.Rat, *big.Rat)) (*big.Int, uint, uint, bool) {
	return timecalc.RoundStepWithBound(step, mode, initial, max, eval)
}
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

func coreLogForm(f logForm) timecalc.LogForm {
	terms := make([]timecalc.LogTerm, len(f.terms))
	for i, t := range f.terms {
		terms[i] = timecalc.LogTerm{Coefficient: t.coefficient, Base: t.base}
	}
	return timecalc.LogForm{Constant: f.constant, Terms: terms}
}
func logFormIsZero(f logForm) bool { return timecalc.IsZero(coreLogForm(f)) }
