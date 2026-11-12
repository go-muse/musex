package tuplet_test

import (
	"math"
	"sync"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/tuplet"

	"github.com/stretchr/testify/assert"
)

// TestImmutability covers K-15, IP-9 and architecture 3.2: input fractions,
// tuplets and chain slices keep their values after successful and failing
// calls, including Must panics.
func TestImmutability(t *testing.T) {
	t.Parallel()

	maxFrac := maxComponentFrac(t)
	maxInt := tuplet.MustNew(1, math.MaxInt)

	fractions := []fraction.Fraction{frac(2, 16), frac(-3, 4), frac(1, -8), frac(0, 5), {}, maxFrac}
	tuplets := []tuplet.Tuplet{tuplet.Triplet(), tuplet.Sextuplet(), {}, maxInt, tuplet.MustNew(math.MaxInt, 1)}

	for _, tp := range tuplets {
		for _, d := range fractions {
			tpBefore, dBefore := tp, d

			_, _ = tp.Apply(d)
			_, _ = tp.Unapply(d)
			_, _ = tp.Span(d)
			_, _ = tuplet.ApplyChain(d, tp, tp)
			_, _ = tuplet.FromCoefficient(d)
			_, _ = tuplet.FromReciprocalCoefficient(d)
			_, _ = tuplet.FromUnitRatio(3, 2, d)
			_, _ = capturePanic(func() { tp.MustApply(d) })
			_, _ = capturePanic(func() { tuplet.MustFromCoefficient(d) })

			assert.Equal(t, tpBefore, tp, "tuplet %s", tpBefore)
			assert.Equal(t, dBefore, d, "fraction %s", dBefore)
		}

		for _, other := range tuplets {
			tpBefore, otherBefore := tp, other

			_, _ = tp.Nest(other)
			_, _ = tp.EqualCoefficient(other)
			_, _ = tp.CompareCoefficient(other)
			_ = tp.Equal(other)
			_ = tp.Compare(other)
			_, _ = tp.Scale(2)
			_, _ = tp.Reduce()
			_, _ = capturePanic(func() { tp.MustNest(other) })

			assert.Equal(t, tpBefore, tp)
			assert.Equal(t, otherBefore, other)
		}
	}

	chains := [][]tuplet.Tuplet{
		{tuplet.Triplet(), tuplet.Duplet(), tuplet.Sextuplet()},
		{tuplet.Triplet(), {}, tuplet.Duplet()},
		{maxInt, maxInt, maxInt, maxInt, maxInt},
	}

	for _, chain := range chains {
		before := append([]tuplet.Tuplet(nil), chain...)

		_, _ = tuplet.Flatten(chain...)
		_, _ = tuplet.ChainCoefficient(chain...)
		_, _ = tuplet.ApplyChain(frac(1, 8), chain...)
		_, _ = capturePanic(func() { tuplet.MustFlatten(chain...) })
		_, _ = capturePanic(func() { tuplet.MustChainCoefficient(chain...) })

		assert.Equal(t, before, chain, "the chain slice is unchanged")
	}
}

// concurrentResult collects the results of a fixed set of operations.
type concurrentResult struct {
	coefficient fraction.Fraction
	applied     fraction.Fraction
	span        fraction.Fraction
	chain       fraction.Fraction
	flat        tuplet.Tuplet
	nested      tuplet.Tuplet
	reduced     tuplet.Tuplet
	scaled      tuplet.Tuplet
	defaulted   tuplet.Tuplet
	parsed      tuplet.Tuplet
	text        string
	name        string
	compare     int
	equal       bool
	errText     string
}

// computeConcurrent runs the operations on shared values.
func computeConcurrent(tp tuplet.Tuplet, d fraction.Fraction, chain []tuplet.Tuplet) concurrentResult {
	var r concurrentResult

	r.coefficient, _ = tp.Coefficient()
	r.applied, _ = tp.Apply(d)
	r.span, _ = tp.Span(d)
	r.chain, _ = tuplet.ChainCoefficient(chain...)
	r.flat, _ = tuplet.Flatten(chain...)
	r.nested, _ = tp.Nest(chain[0])
	r.reduced, _ = tp.Reduce()
	r.scaled, _ = tp.Scale(3)
	r.defaulted, _ = tuplet.Default(7, tuplet.Binary)
	r.parsed, _ = tuplet.Parse(tp.String())
	r.text = tp.String()
	r.name, _, _ = tp.Name()
	r.compare = tp.Compare(chain[0])
	r.equal, _ = tp.EqualCoefficient(chain[1])

	if _, err := tuplet.Parse(rep("x", 200)); err != nil {
		r.errText = err.Error()
	}

	return r
}

// TestConcurrency covers K-15 and architecture 3.2: concurrent calls on
// shared values give the same results as sequential calls; run with -race.
func TestConcurrency(t *testing.T) {
	t.Parallel()

	tp := tuplet.Sextuplet()
	d := frac(3, 16)
	chain := []tuplet.Tuplet{tuplet.Triplet(), tuplet.MustNew(9, 6), tuplet.Quintuplet()}
	want := computeConcurrent(tp, d, chain)

	const workers = 32

	results := make([]concurrentResult, workers)

	var wg sync.WaitGroup

	for i := range workers {
		wg.Go(func() {
			results[i] = computeConcurrent(tp, d, chain)
		})
	}

	wg.Wait()

	for i, got := range results {
		assert.Equal(t, want, got, "worker %d", i)
	}
}
