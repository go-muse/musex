package timecalc_test

import (
	"math/big"
	"sync"
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/internal/timecalc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func copyForm(form timecalc.LogForm) timecalc.LogForm {
	out := timecalc.LogForm{Constant: new(big.Rat).Set(form.Constant)}
	if form.Terms != nil {
		out.Terms = make([]timecalc.LogTerm, len(form.Terms))
		for i, item := range form.Terms {
			out.Terms[i] = timecalc.LogTerm{Coefficient: new(big.Rat).Set(item.Coefficient), Base: new(big.Rat).Set(item.Base)}
		}
	}
	return out
}

func mutateForm(form timecalc.LogForm) {
	form.Constant.SetInt64(99)
	for i := range form.Terms {
		form.Terms[i].Coefficient.SetInt64(77)
		form.Terms[i].Base.SetInt64(5)
	}
	if len(form.Terms) > 0 {
		form.Terms[0] = term("88", "7")
	}
}

func TestConstructorsOwnTheirResults(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"constant", "add", "scale"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			input := logForm("3", term("2", "2"))
			other := logForm("4", term("-1", "2"))
			factor := rat("1/2")
			inputBefore, otherBefore := copyForm(input), copyForm(other)
			var got, want timecalc.LogForm
			switch operation {
			case "constant":
				got, want = timecalc.Constant(input.Constant), logForm("3")
			case "add":
				got, want = timecalc.Add(input, other), logForm("7", term("1", "2"))
			case "scale":
				got, want = timecalc.Scale(input, factor), logForm("3/2", term("1", "2"))
			}
			assert.Equal(t, inputBefore, input)
			assert.Equal(t, otherBefore, other)
			requireRat(t, rat("1/2"), factor)
			mutateForm(input)
			mutateForm(other)
			factor.SetInt64(9)
			requireRational(t, timecalc.Add(got, timecalc.Scale(want, rat("-1"))), "0")
			// In the opposite direction, changing the result cannot write back.
			inputAfter, otherAfter := copyForm(input), copyForm(other)
			mutateForm(got)
			assert.Equal(t, inputAfter, input)
			assert.Equal(t, otherAfter, other)
			requireRat(t, rat("9"), factor)
		})
	}
}

func TestLinearSpanDoesNotRetainInputs(t *testing.T) {
	t.Parallel()
	inputs := []*big.Rat{rat("0"), rat("1"), rat("60"), rat("120"), rat("0"), rat("1")}
	before := []string{"0", "1", "60", "120", "0", "1"}
	got := timecalc.LinearSpan(inputs[0], inputs[1], inputs[2], inputs[3], inputs[4], inputs[5])
	for i, value := range inputs {
		requireRat(t, rat(before[i]), value)
		value.SetInt64(99)
	}
	requireRational(t, timecalc.Add(got, logForm("0", term("-1", "2"))), "0")
	mutateForm(got)
	for _, value := range inputs {
		requireRat(t, rat("99"), value)
	}
}

func TestQueriesDoNotMutateOrAlias(t *testing.T) {
	t.Parallel()
	form := logForm("1/3", term("1", "4"), term("-2", "2"))
	before := copyForm(form)
	step, threshold := rat("1/10"), rat("1/2")
	rational, exact := timecalc.Rational(form)
	require.True(t, exact)
	requireRat(t, rat("1/3"), rational)
	rational.SetInt64(99)
	first := timecalc.RoundLogSteps(form, step, rounding.HalfUp)
	requireInt(t, "3", first)
	first.SetInt64(99)
	requireInt(t, "3", timecalc.RoundLogSteps(form, step, rounding.HalfUp))
	assert.Equal(t, -1, timecalc.CompareLog(form, threshold))
	lo, hi, intervalExact := timecalc.DecimalInterval(form, 3)
	require.True(t, intervalExact)
	require.NotNil(t, lo)
	require.NotNil(t, hi)
	lo.SetInt64(77)
	hi.SetInt64(88)
	assert.Equal(t, before, form)
	requireRat(t, rat("1/10"), step)
	requireRat(t, rat("1/2"), threshold)
	// Irrational queries also leave all original coefficient/base pointers alone.
	irrational := logForm("0", term("1", "2"))
	irrationalBefore := copyForm(irrational)
	value, irrationalExact := timecalc.Rational(irrational)
	assert.False(t, irrationalExact)
	assert.Nil(t, value)
	requireInt(t, "7", timecalc.RoundLogSteps(irrational, step, rounding.HalfUp))
	assert.Equal(t, 1, timecalc.CompareLog(irrational, threshold))
	lo, hi, intervalExact = timecalc.DecimalInterval(irrational, 3)
	assert.False(t, intervalExact)
	require.NotNil(t, lo)
	require.NotNil(t, hi)
	lo.SetInt64(77)
	hi.SetInt64(88)
	assert.Equal(t, irrationalBefore, irrational)
	requireRat(t, rat("1/10"), step)
	requireRat(t, rat("1/2"), threshold)
}

func TestConcurrentQueries(t *testing.T) {
	t.Parallel()
	form := logForm("0", term("1", "2"))
	before := copyForm(form)
	step, threshold := rat("1/1000"), rat("7/10")
	start := make(chan struct{})
	results := make(chan bool, 16)
	var workers sync.WaitGroup
	for range 16 {
		workers.Go(func() {
			<-start
			valid := true
			for range 8 {
				index := timecalc.RoundLogSteps(form, step, rounding.HalfEven)
				lo, hi, exact := timecalc.DecimalInterval(form, 3)
				value, rational := timecalc.Rational(form)
				cancelled, cancelledExact := timecalc.Rational(timecalc.Add(timecalc.Scale(form, rat("2")), timecalc.Scale(form, rat("-2"))))
				valid = valid && index != nil && index.String() == "693" &&
					lo != nil && lo.Cmp(rat(".6925")) == 0 && hi != nil && hi.Cmp(rat(".6935")) == 0 &&
					!exact && !rational && value == nil && timecalc.CompareLog(form, threshold) == -1 &&
					cancelledExact && cancelled != nil && cancelled.Sign() == 0
			}
			results <- valid
		})
	}
	close(start)
	workers.Wait()
	close(results)
	for valid := range results {
		assert.True(t, valid)
	}
	assert.Equal(t, before, form)
	requireRat(t, rat("1/1000"), step)
	requireRat(t, rat("7/10"), threshold)
}
