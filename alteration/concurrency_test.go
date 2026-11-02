package alteration_test

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"testing"

	"github.com/go-muse/musex/alteration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-19: all operations called concurrently from many goroutines on shared
// values give the results of K-1...K-12. Data races are detected when the
// test runs under go test -race, as in make test and CI. The expected
// results are computed with math/big in the test goroutine before the
// workers start; every worker writes only its own slot; the comparison runs
// in the test goroutine after all workers finish.
func TestConcurrentUse(t *testing.T) {
	t.Parallel()

	const iterations = 3

	shared := values(t, setS())
	top := value(t, edgeTop())
	want := expectedLog(t)

	workers := runtime.GOMAXPROCS(0) * 4
	logs := make([][]string, workers)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			<-start
			for range iterations {
				logs[w] = append(logs[w], operationLog(shared, top)...)
			}
		})
	}
	close(start)
	wg.Wait()

	for w, log := range logs {
		require.Len(t, log, iterations*len(want), "worker %d", w)

		for i := range iterations {
			assert.Equal(t, want, log[i*len(want):(i+1)*len(want)], "worker %d, iteration %d", w, i)
		}
	}
}

// describe renders a value by its exact fraction, independently of String.
func describe(a alteration.Alteration) string {
	num, den := a.Fraction()

	return fmt.Sprintf("%d/%d", num, den)
}

// describeOutcome renders the result of a fallible call.
func describeOutcome(a alteration.Alteration, err error) string {
	switch {
	case err == nil:
		return describe(a)
	case errors.Is(err, alteration.ErrOutOfRange):
		return "out of range"
	case errors.Is(err, alteration.ErrZeroDenominator):
		return "zero denominator"
	default:
		return "unexpected error"
	}
}

// describeMust renders the result of a Must call, including a panic.
func describeMust(f func() alteration.Alteration) string {
	o := callMust(f)

	return "must " + describeOutcome(o.value, o.err)
}

// operationLog calls every public operation on the shared values and returns
// a description of each result. It runs in worker goroutines: it must not
// call testing functions.
func operationLog(shared []alteration.Alteration, top alteration.Alteration) []string {
	var log []string

	add := func(format string, args ...any) {
		log = append(log, fmt.Sprintf(format, args...))
	}

	add("signs %s %s %s %s %s", describe(alteration.Natural()), describe(alteration.Sharp()),
		describe(alteration.Flat()), describe(alteration.DoubleSharp()), describe(alteration.DoubleFlat()))

	for n := -10; n <= 10; n++ {
		add("FromInt %s", describeOutcome(alteration.FromInt(n)))
		add("MustFromInt %s", describeMust(func() alteration.Alteration { return alteration.MustFromInt(n) }))
	}

	for p := -6; p <= 6; p++ {
		for q := -6; q <= 6; q++ {
			add("FromFraction %s", describeOutcome(alteration.FromFraction(p, q)))
			add("MustFromFraction %s", describeMust(func() alteration.Alteration { return alteration.MustFromFraction(p, q) }))
		}
	}

	add("top Raise %s", describeOutcome(top.Raise()))
	add("top MustRaise %s", describeMust(top.MustRaise))

	for _, a := range shared {
		num, den := a.Fraction()
		add("Fraction %d %d", num, den)
		add("Sign %d IsNatural %t IsInteger %t String %s", a.Sign(), a.IsNatural(), a.IsInteger(), a.String())
		add("Raise %s", describeOutcome(a.Raise()))
		add("Lower %s", describeOutcome(a.Lower()))
		add("MustRaise %s", describeMust(a.MustRaise))
		add("MustLower %s", describeMust(a.MustLower))

		for n := -5; n <= 5; n++ {
			add("RaiseBy %s", describeOutcome(a.RaiseBy(n)))
			add("LowerBy %s", describeOutcome(a.LowerBy(n)))
			add("MustRaiseBy %s", describeMust(func() alteration.Alteration { return a.MustRaiseBy(n) }))
			add("MustLowerBy %s", describeMust(func() alteration.Alteration { return a.MustLowerBy(n) }))
		}

		for k := -4; k <= 4; k++ {
			add("Mul %s", describeOutcome(a.Mul(k)))
			add("MustMul %s", describeMust(func() alteration.Alteration { return a.MustMul(k) }))
		}

		for _, b := range shared {
			add("Add %s", describeOutcome(a.Add(b)))
			add("Sub %s", describeOutcome(a.Sub(b)))
			add("MustAdd %s", describeMust(func() alteration.Alteration { return a.MustAdd(b) }))
			add("MustSub %s", describeMust(func() alteration.Alteration { return a.MustSub(b) }))
			add("Equal %t Compare %d", a.Equal(b), a.Compare(b))
		}
	}

	return log
}

// expectedLog builds the log operationLog must produce, from math/big and the
// contract alone, in the same order.
func expectedLog(t *testing.T) []string {
	t.Helper()

	var log []string

	add := func(format string, args ...any) {
		log = append(log, fmt.Sprintf(format, args...))
	}

	exact := func(want frac) string {
		return fmt.Sprintf("%d/%d", want.num, want.den)
	}

	oracle := func(r *big.Rat) string {
		if r == nil {
			return "zero denominator"
		}
		if !inRange(r) {
			return "out of range"
		}

		return exact(expectedFrac(t, r))
	}

	add("signs 0/1 1/1 -1/1 2/1 -2/1")

	for n := -10; n <= 10; n++ {
		add("FromInt %s", oracle(ratInt(n)))
		add("MustFromInt must %s", oracle(ratInt(n)))
	}

	for p := -6; p <= 6; p++ {
		for q := -6; q <= 6; q++ {
			var r *big.Rat
			if q != 0 {
				r = ratOf(p, q)
			}
			add("FromFraction %s", oracle(r))
			add("MustFromFraction must %s", oracle(r))
		}
	}

	topPlusOne := ratAdd(edgeTop().rat(), ratInt(1))
	add("top Raise %s", oracle(topPlusOne))
	add("top MustRaise must %s", oracle(topPlusOne))

	for _, fa := range setS() {
		ra := fa.rat()
		f := expectedFrac(t, ra)
		add("Fraction %d %d", f.num, f.den)
		add("Sign %d IsNatural %t IsInteger %t String %s", ra.Sign(), ra.Sign() == 0, ra.IsInt(), contractString(ra))
		add("Raise %s", oracle(ratAdd(ra, ratInt(1))))
		add("Lower %s", oracle(ratSub(ra, ratInt(1))))
		add("MustRaise must %s", oracle(ratAdd(ra, ratInt(1))))
		add("MustLower must %s", oracle(ratSub(ra, ratInt(1))))

		for n := -5; n <= 5; n++ {
			add("RaiseBy %s", oracle(ratAdd(ra, ratInt(n))))
			add("LowerBy %s", oracle(ratSub(ra, ratInt(n))))
			add("MustRaiseBy must %s", oracle(ratAdd(ra, ratInt(n))))
			add("MustLowerBy must %s", oracle(ratSub(ra, ratInt(n))))
		}

		for k := -4; k <= 4; k++ {
			add("Mul %s", oracle(ratMul(ra, k)))
			add("MustMul must %s", oracle(ratMul(ra, k)))
		}

		for _, fb := range setS() {
			rb := fb.rat()
			add("Add %s", oracle(ratAdd(ra, rb)))
			add("Sub %s", oracle(ratSub(ra, rb)))
			add("MustAdd must %s", oracle(ratAdd(ra, rb)))
			add("MustSub must %s", oracle(ratSub(ra, rb)))
			add("Equal %t Compare %d", ra.Cmp(rb) == 0, ra.Cmp(rb))
		}
	}

	return log
}
