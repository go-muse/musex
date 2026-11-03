package pitchclass_test

import (
	"errors"
	"fmt"
	"iter"
	"runtime"
	"sync"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
)

// shared holds the values used by all workers. It is created in the test
// goroutine before the workers start and only read by them; the expected
// results come from the oracle, not from the package.
type shared struct {
	classes   []pitchclass.Class // the classes of 12, ascending
	others    []pitchclass.Class // the classes of 19, ascending
	ops       []pitchclass.Operator
	specs     []opSpec
	all       iter.Seq[pitchclass.Class]
	cycle     iter.Seq[pitchclass.Class] // from 10 of 12
	operators iter.Seq[pitchclass.Operator]
	cycleWant []int

	transpositions iter.Seq[pitchclass.Operator]
	ti             iter.Seq[pitchclass.Operator]
	multipliers    iter.Seq[int]
	tSpecs         []opSpec // the oracle of transpositions
	tiSpecs        []opSpec // the oracle of ti
	multWant       []int    // the oracle of multipliers
}

func newShared(t *testing.T) shared {
	t.Helper()

	ops, specs := operatorsOf(t, 12)

	return shared{
		classes:   classesOf(t, 12),
		others:    classesOf(t, 19),
		ops:       ops,
		specs:     specs,
		all:       pitchclass.MustAll(12),
		cycle:     newClass(t, 10, 12).Cycle(),
		operators: pitchclass.MustOperators(12),
		cycleWant: []int{10, 11, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9},

		transpositions: pitchclass.MustTranspositions(12),
		ti:             pitchclass.MustTranspositionsInversions(12),
		multipliers:    pitchclass.MustInvertibleMultipliers(12),
		tSpecs:         transpositionSpecs(12),
		tiSpecs:        tiSpecs(12),
		multWant:       invertibleList(12),
	}
}

// K-32, NF-3: all operations of the architecture summary (the Must variants
// are wrappers and are not repeated) called concurrently from many goroutines
// on the same values, including ranging over the same iterators at the same
// time, give the results of K-1...K-27. Data races are detected under go test
// -race (make test, CI). Workers report mismatches through their own slot;
// assertions run only in the test goroutine.
func TestConcurrentUse(t *testing.T) {
	t.Parallel()

	const iterations = 2

	s := newShared(t)
	workers := runtime.GOMAXPROCS(0) * 4
	failures := make([][]string, workers)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			// A panic in a worker is recorded in its slot instead of
			// stopping the whole test binary.
			defer func() {
				if r := recover(); r != nil {
					failures[w] = append(failures[w], fmt.Sprintf("panic: %v", r))
				}
			}()

			<-start
			for range iterations {
				failures[w] = append(failures[w], checkConcurrent(s, w)...)
			}
		})
	}
	close(start)
	wg.Wait()

	for w, f := range failures {
		assert.Empty(t, f, "worker %d", w)
	}
}

// checkConcurrent calls the public operations on the shared values and on
// new values and returns a description of every mismatch with the oracle.
func checkConcurrent(s shared, worker int) []string {
	var failures []string
	fail := func(format string, args ...any) {
		failures = append(failures, fmt.Sprintf(format, args...))
	}

	const size = 12

	for i, x := range s.classes {
		if x.Number() != i || x.Modulus() != size || !x.IsValid() || x.String() != fmt.Sprintf("%d mod 12", i) {
			fail("class %d reads %d, %d, %t, %q", i, x.Number(), x.Modulus(), x.IsValid(), x.String())
		}

		for n := -3; n <= 3; n++ {
			if got := x.Transpose(n + worker); got.Number() != addMod(i, n+worker, size) {
				fail("T%d(%d) = %v", n+worker, i, got)
			}
			if got := x.Invert(n); got.Number() != subMod(n, i, size) {
				fail("I%d(%d) = %v", n, i, got)
			}
			if got := x.Multiply(n); got.Number() != mulMod(n, i, size) {
				fail("M%d(%d) = %v", n, i, got)
			}
		}

		for j, y := range s.classes {
			if got, err := x.IntervalTo(y); err != nil || got != subMod(j, i, size) {
				fail("i<%d, %d> = %d, %v", i, j, got, err)
			}
			if got, err := x.IntervalClass(y); err != nil || got != intervalClassOf(j-i, size) {
				fail("ic(%d, %d) = %d, %v", i, j, got, err)
			}
			if got, err := x.IndexSum(y); err != nil || got != addMod(i, j, size) {
				fail("%d + %d = %d, %v", i, j, got, err)
			}
			if got, err := x.InvertAround(y); err != nil || got.Number() != subMod(2*j, i, size) {
				fail("%d around %d = %v, %v", i, j, got, err)
			}
			if x.Equal(y) != (i == j) || x.Compare(y) != sign(i-j) {
				fail("Equal or Compare of %d and %d", i, j)
			}
		}

		if _, err := x.IntervalTo(s.others[i]); !errors.Is(err, pitchclass.ErrModulusMismatch) {
			fail("IntervalTo across N: %v", err)
		}

		if got, err := pitchclass.Parse(x.String()); err != nil || !got.Equal(x) {
			fail("Parse(%q) = %v, %v", x.String(), got, err)
		}

		if got, err := x.Notation(pitchclass.TE); err != nil || got != notationOf(i, pitchclass.TE) {
			fail("Notation(TE) of %d = %q, %v", i, got, err)
		}
	}

	for i, f := range s.ops {
		want := s.specs[i]
		if f.String() != want.String() || f.Multiplier() != want.m || f.Index() != want.n {
			fail("operator %v reads %q", want, f.String())
		}
		if inv := f.Inverse(); inv.Multiplier() != want.inverse().m || inv.Index() != want.inverse().n {
			fail("%v^-1 = %v", want, inv)
		}

		for _, x := range s.classes {
			if got, err := f.Apply(x); err != nil || got.Number() != want.apply(x.Number()) {
				fail("%v(%v) = %v, %v", want, x, got, err)
			}
		}

		for j := worker % len(s.ops); j < len(s.ops); j += 7 {
			g := s.ops[j]
			fg, err := f.Compose(g)
			c := want.compose(s.specs[j])
			if err != nil || fg.Multiplier() != c.m || fg.Index() != c.n {
				fail("%v∘%v = %v, %v", want, s.specs[j], fg, err)
			}
			if f.Equal(g) != (i == j) || f.Compare(g) != compareSpec(want, s.specs[j]) {
				fail("Equal or Compare of %v and %v", want, s.specs[j])
			}
		}

		if f.Modulus() != size || !f.IsValid() || f.IsTransposition() != want.isTransposition() || f.IsInversion() != want.isInversion() {
			fail("properties of %v: %d, %t, %t, %t", want, f.Modulus(), f.IsValid(), f.IsTransposition(), f.IsInversion())
		}
		if got, err := pitchclass.ParseOperator(want.String()); err != nil || !got.Equal(f) {
			fail("ParseOperator(%q) = %v, %v", want.String(), got, err)
		}
	}

	failures = append(failures, checkConcurrentMore(s, worker)...)

	k := 0
	for c := range s.all {
		if c.Number() != k {
			fail("All(12)[%d] = %v", k, c)
		}
		k++
	}
	if k != size {
		fail("All(12) yields %d classes", k)
	}

	k = 0
	for c := range s.cycle {
		if k >= len(s.cycleWant) || c.Number() != s.cycleWant[k] {
			fail("Cycle from 10, element %d = %v", k, c)
		}
		k++
	}
	if k != size {
		fail("Cycle yields %d classes", k)
	}

	k = 0
	for f := range s.operators {
		if k >= len(s.specs) || f.Multiplier() != s.specs[k].m || f.Index() != s.specs[k].n {
			fail("Operators(12)[%d] = %v", k, f)
		}
		k++
	}
	if k != len(s.specs) {
		fail("Operators(12) yields %d operators", k)
	}

	fixed, err := pitchclass.InversionFixedClasses(4, size)
	if err != nil || len(fixed) != 2 || fixed[0].Number() != 2 || fixed[1].Number() != 8 {
		fail("InversionFixedClasses(4, 12) = %v, %v", fixed, err)
	}
	if len(fixed) > 0 {
		fixed[0] = pitchclass.Class{}
	}

	if n, err := pitchclass.CountOperators(size); err != nil || n != 48 {
		fail("CountOperators(12) = %d, %v", n, err)
	}
	if b, err := pitchclass.CountOperatorsBig(size); err != nil || b.String() != "48" {
		fail("CountOperatorsBig(12) = %v, %v", b, err)
	} else {
		b.SetInt64(int64(worker))
	}
	if phi, err := pitchclass.Totient(maxInt); err != nil || phi != 7713001620195508224 {
		fail("Totient(MAX) = %d, %v", phi, err)
	}

	if f, err := pitchclass.NewOperator(7, worker, size); err != nil || f.Index() != mod(worker, size) {
		fail("NewOperator(7, %d, 12) = %v, %v", worker, f, err)
	}
	if c, err := pitchclass.Reduce(worker, size); err != nil || c.Number() != mod(worker, size) {
		fail("Reduce(%d, 12) = %v, %v", worker, c, err)
	}

	return failures
}

// checkConcurrentMore continues checkConcurrent with the remaining operations:
// creation, notation in all styles, inversion of a pair, intervals by size,
// multipliers, operator constructors, the other shared iterators, counts and
// Style.String.
func checkConcurrentMore(s shared, worker int) []string {
	var failures []string
	fail := func(format string, args ...any) {
		failures = append(failures, fmt.Sprintf(format, args...))
	}

	const size = 12

	for i, x := range s.classes {
		if c, err := pitchclass.New(i, size); err != nil || !c.Equal(x) {
			fail("New(%d, 12) = %v, %v", i, c, err)
		}

		for _, st := range styles() {
			want := notationOf(i, st)
			if got, err := x.Notation(st); err != nil || got != want {
				fail("Notation(%v) of %d = %q, %v", st, i, got, err)
			}
			if got, err := pitchclass.ParseNotation(want, st, size); err != nil || !got.Equal(x) {
				fail("ParseNotation(%q, %v, 12) = %v, %v", want, st, got, err)
			}
		}

		if f, err := pitchclass.InversionAround(x); err != nil || f.Modulus() != size || f.Multiplier() != 11 || f.Index() != mod(2*i, size) {
			fail("InversionAround(%d) = %v, %v", i, f, err)
		}

		for j, y := range s.classes {
			if f, err := pitchclass.InversionSwapping(x, y); err != nil || f.Multiplier() != 11 || f.Index() != addMod(i, j, size) {
				fail("InversionSwapping(%d, %d) = %v, %v", i, j, f, err)
			}

			for k := worker % size; k < size; k += 5 {
				if got, err := x.InvertSwapping(y, s.classes[k]); err != nil || got.Number() != subMod(j+k, i, size) {
					fail("%d swapping %d and %d = %v, %v", i, j, k, got, err)
				}
			}
		}
	}

	for n := worker - 3; n <= worker+3; n++ {
		if got, err := pitchclass.ReduceInterval(n, size); err != nil || got != mod(n, size) {
			fail("ReduceInterval(%d, 12) = %d, %v", n, got, err)
		}
		if got, err := pitchclass.ComplementInterval(n, size); err != nil || got != negMod(n, size) {
			fail("ComplementInterval(%d, 12) = %d, %v", n, got, err)
		}
		if got, err := pitchclass.IntervalClass(n, size); err != nil || got != intervalClassOf(n, size) {
			fail("IntervalClass(%d, 12) = %d, %v", n, got, err)
		}
	}

	for m := range size {
		invertible := invertibleOracle(m, size)
		if got, err := pitchclass.Invertible(m, size); err != nil || got != invertible {
			fail("Invertible(%d, 12) = %t, %v", m, got, err)
		}

		inv, err := pitchclass.InverseMultiplier(m, size)
		switch {
		case invertible && (err != nil || inv != inverseOracle(m, size)):
			fail("InverseMultiplier(%d, 12) = %d, %v", m, inv, err)
		case !invertible && !errors.Is(err, pitchclass.ErrNotInvertible):
			fail("InverseMultiplier(%d, 12) error %v", m, err)
		}

		n := m + worker
		if f, err := pitchclass.Transposition(n, size); err != nil || f.Multiplier() != 1 || f.Index() != mod(n, size) {
			fail("Transposition(%d, 12) = %v, %v", n, f, err)
		}
		if f, err := pitchclass.Inversion(n, size); err != nil || f.Multiplier() != 11 || f.Index() != mod(n, size) {
			fail("Inversion(%d, 12) = %v, %v", n, f, err)
		}

		f, err := pitchclass.Multiplication(m, size)
		switch {
		case invertible && (err != nil || f.Multiplier() != m || f.Index() != 0):
			fail("Multiplication(%d, 12) = %v, %v", m, f, err)
		case !invertible && !errors.Is(err, pitchclass.ErrNotInvertible):
			fail("Multiplication(%d, 12) error %v", m, err)
		}
	}

	if f, err := pitchclass.Identity(size); err != nil || f.Multiplier() != 1 || f.Index() != 0 || f.Modulus() != size {
		fail("Identity(12) = %v, %v", f, err)
	}

	k := 0
	for f := range s.transpositions {
		if k >= len(s.tSpecs) || f.Multiplier() != s.tSpecs[k].m || f.Index() != s.tSpecs[k].n {
			fail("Transpositions(12)[%d] = %v", k, f)
		}
		k++
	}
	if k != len(s.tSpecs) {
		fail("Transpositions(12) yields %d operators", k)
	}

	k = 0
	for f := range s.ti {
		if k >= len(s.tiSpecs) || f.Multiplier() != s.tiSpecs[k].m || f.Index() != s.tiSpecs[k].n {
			fail("TranspositionsInversions(12)[%d] = %v", k, f)
		}
		k++
	}
	if k != len(s.tiSpecs) {
		fail("TranspositionsInversions(12) yields %d operators", k)
	}

	k = 0
	for m := range s.multipliers {
		if k >= len(s.multWant) || m != s.multWant[k] {
			fail("InvertibleMultipliers(12)[%d] = %d", k, m)
		}
		k++
	}
	if k != len(s.multWant) {
		fail("InvertibleMultipliers(12) yields %d multipliers", k)
	}

	if n, err := pitchclass.CountTranspositions(size); err != nil || n != 12 {
		fail("CountTranspositions(12) = %d, %v", n, err)
	}
	if n, err := pitchclass.CountTranspositionsInversions(size); err != nil || n != 24 {
		fail("CountTranspositionsInversions(12) = %d, %v", n, err)
	}
	if b, err := pitchclass.CountTranspositionsInversionsBig(size); err != nil || b.String() != "24" {
		fail("CountTranspositionsInversionsBig(12) = %v, %v", b, err)
	} else {
		b.SetInt64(int64(-worker))
	}

	for st, name := range map[pitchclass.Style]string{
		pitchclass.Decimal: "Decimal", pitchclass.TE: "TE", pitchclass.LowerTE: "LowerTE", pitchclass.AB: "AB",
	} {
		if got := st.String(); got != name {
			fail("Style %d String = %q", st, got)
		}
	}

	return failures
}
