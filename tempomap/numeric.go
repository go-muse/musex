package tempomap

import (
	"math/big"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/internal/timecalc"
)

// Private domain adapters retain the old form and call one shared kernel.
type numericExpression interface {
	calculation() timecalc.Expression
	inputBits() int
}
type logExpression struct{ form logForm }

func (e logExpression) calculation() timecalc.Expression {
	return timecalc.Logarithmic(coreLogForm(e.form))
}

type expForm struct {
	a, b           *big.Rat
	x              logForm
	minExp, maxExp *big.Rat
}

func (e expForm) calculation() timecalc.Expression {
	return timecalc.Exponential(e.a, e.b, coreLogForm(e.x), e.minExp, e.maxExp)
}
func stepInitialPrecision(step *big.Rat, bits int) uint {
	return timecalc.StepInitialPrecision(step, bits)
}
func roundExpression(e numericExpression, places int, mode rounding.Mode) *big.Rat {
	return timecalc.RoundExpression(e.calculation(), places, mode)
}
func roundExpressionSteps(e numericExpression, step *big.Rat, mode rounding.Mode, initial uint) *big.Int {
	return timecalc.RoundExpressionSteps(e.calculation(), step, mode, initial)
}
func compareLogForm(f logForm, y *big.Rat) int { return timecalc.CompareLog(coreLogForm(f), y) }

func (e logExpression) inputBits() int { return e.calculation().InputBits() }
func (e expForm) inputBits() int       { return e.calculation().InputBits() }
