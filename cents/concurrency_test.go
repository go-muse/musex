package cents_test

import (
	"fmt"
	"math"
	"math/big"
	"runtime"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
)

// recorder collects the results of calls as text. It never fails a test: it
// is used in worker goroutines.
type recorder struct {
	out []string
}

func (r *recorder) add(format string, args ...any) {
	r.out = append(r.out, fmt.Sprintf(format, args...))
}

// must records the result of a Must variant, or its panic as an error kind,
// so that a panic in a worker does not stop the test binary.
func (r *recorder) must(name string, f func() string) {
	v, panicked := panicValue(func() { r.add("%s %s", name, f()) })
	if !panicked {
		return
	}

	err, ok := v.(error)
	if !ok {
		r.add("%s panic %v", name, v)

		return
	}

	r.add("%s panic %s", name, errName(err))
}

// quotient records every method of a Quotient.
func (r *recorder) quotient(name string, q cents.Quotient) {
	exact, ok := q.Exact()
	s := "<nil>"
	if exact != nil {
		s = exact.RatString()
	}

	r.add("%s exact %s %v float %x decimal %s string %s", name, s, ok, math.Float64bits(q.Float64()), q.Decimal(12), q.String())
}

// creations calls every function that creates a value, on valid and failing
// fixed inputs, and returns the created values. It is the first use of the
// package in TestConcurrentUse, so lazy package state is first built by
// several goroutines at once.
func creations(r *recorder) []cents.Cents {
	var vs []cents.Cents

	keep := func(name string, c cents.Cents, err error) {
		r.add("%s %s %s", name, c, errName(err))

		if err == nil {
			vs = append(vs, c)
		}
	}

	for _, f := range []func() cents.Cents{
		cents.Zero, cents.Cent, cents.Semitone, cents.Octave, cents.MIDITuningUnit, cents.Jot,
		cents.Millioctave, cents.Centitone, cents.Savart, cents.Eptameride, cents.Meride,
		cents.Schisma, cents.SeptimalKleisma, cents.Kleisma, cents.PtolemysComma, cents.Diaschisma,
		cents.SyntonicComma, cents.HoldrianComma, cents.PythagoreanComma, cents.SeptimalComma,
		cents.LesserDiesis, cents.SeptimalQuarterTone, cents.UndecimalComma, cents.GreaterDiesis,
		cents.TridecimalComma, cents.MinorChroma, cents.Limma, cents.MajorChroma, cents.Apotome,
		cents.LargeLimma,
	} {
		keep("function", f(), nil)
	}

	for _, nv := range cents.NamedValues() {
		r.add("named %s %s", nv.Name, nv.Value)
	}

	for _, nv := range cents.NamedUnits() {
		r.add("unit %s %s", nv.Name, nv.Value)
	}

	for _, n := range []int{700, -1200, 9000001} {
		c, err := cents.FromInt(n)
		keep("FromInt", c, err)
		r.must("MustFromInt", func() string { return cents.MustFromInt(n).String() })
	}

	for _, p := range [][2]int{{1200, 19}, {-137, 10}, {1, 0}} {
		c, err := cents.FromFraction(p[0], p[1])
		keep("FromFraction", c, err)
		r.must("MustFromFraction", func() string { return cents.MustFromFraction(p[0], p[1]).String() })
	}

	for _, x := range []float64{701.955001, 0.1234565, math.NaN()} {
		c, err := cents.FromFloat(x)
		keep("FromFloat", c, err)
		r.must("MustFromFloat", func() string { return cents.MustFromFloat(x).String() })
	}

	for _, p := range [][2]int{{1, 53}, {31, 53}, {1, 0}} {
		c, err := cents.Step(p[0], p[1])
		keep("Step", c, err)
		r.must("MustStep", func() string { return cents.MustStep(p[0], p[1]).String() })
	}

	for _, p := range [][2]int{{3, 2}, {442, 440}, {31, 16}, {37, 32}} {
		c, err := cents.FromRatio(p[0], p[1])
		keep("FromRatio", c, err)
		r.must("MustFromRatio", func() string { return cents.MustFromRatio(p[0], p[1]).String() })
	}

	for _, x := range []float64{1.5, 1.1, 0} {
		c, err := cents.FromRatioFloat(x)
		keep("FromRatioFloat", c, err)
		r.must("MustFromRatioFloat", func() string { return cents.MustFromRatioFloat(x).String() })
	}

	for _, p := range [][2]float64{{440, 442}, {1e-300, 1e300}, {440, -440}} {
		c, err := cents.Between(p[0], p[1])
		keep("Between", c, err)
		r.must("MustBetween", func() string { return cents.MustBetween(p[0], p[1]).String() })
	}

	for _, ps := range [][]cents.Power{
		{{Base: 5, Num: big.NewInt(1), Den: big.NewInt(4)}},
		{{Base: 3, Num: big.NewInt(53)}, {Base: 2, Num: big.NewInt(-84)}},
		{{Base: 37, Num: big.NewInt(1)}},
	} {
		c, err := cents.FromComposition(big.NewInt(6), big.NewInt(5), ps...)
		keep("FromComposition", c, err)
		r.must("MustFromComposition", func() string { return cents.MustFromComposition(big.NewInt(6), big.NewInt(5), ps...).String() })
	}

	for _, s := range []string{"5^(1/4)", "2^(1/1000)*5^(1/1000)", "-1200/19 cents", "−13.7 ¢", "3/2"} {
		c, err := cents.Parse(s)
		keep("Parse", c, err)
		r.must("MustParse", func() string { return cents.MustParse(s).String() })
	}

	for _, s := range []string{" 5/4   E", "408.", "37/32 x"} {
		c, rest, err := cents.ParseScala(s)
		keep("ParseScala "+rest, c, err)
		r.must("MustParseScala", func() string { c, rest := cents.MustParseScala(s); return c.String() + rest })

		c, rest, err = cents.ParseScalaApprox(s)
		keep("ParseScalaApprox "+rest, c, err)
		r.must("MustParseScalaApprox", func() string { c, rest := cents.MustParseScalaApprox(s); return c.String() + rest })
	}

	return vs
}

// operations calls every operation of a value on the values of vs: text,
// numbers, arithmetic, comparison, quotients and frequencies, with valid and
// failing arguments and the Must variants. Pairs (Add, Sub, Compare, Equal)
// are all pairs; Near and Quo with their Must variants take a quarter of
// them.
func operations(r *recorder, vs []cents.Cents) {
	for i, v := range vs {
		r.add("text %s %s %s %s %s %v", v, v.Text(cents.Typographic|cents.WithUnit|cents.WithPlus), v.Decimal(9, 0), v.Scala(5), v.Kind(), v.Sign())

		for _, k := range kinds() {
			s, ok := v.TextAs(k, cents.WithUnit)
			r.add("as %s %q %v", k, s, ok)
		}

		r.add("fmt %v %+v %.3f %q", v, v, v, v)
		r.add("float %x", math.Float64bits(v.Float64()))

		num, den, ok := v.Fraction()
		r.add("fraction %d %d %v", num, den, ok)

		num, den, ok = v.Ratio()
		r.add("ratio %d %d %v", num, den, ok)
		r.add("composition %s", compOf(v))
		r.add("neg abs %s %s", v.Neg(), v.Abs())

		octaves, rest := v.SplitOctaves()
		r.add("split %d %s", octaves, rest)

		for _, n := range []int{12, 53, 0} {
			k, dev, err := v.NearestStep(n)
			r.add("step %d %d %s %s", n, k, dev, errName(err))
			r.must("MustNearestStep", func() string { k, dev := v.MustNearestStep(n); return fmt.Sprint(k, " ", dev) })
		}

		m, err := v.Multiplier()
		r.add("multiplier %x %s", math.Float64bits(m), errName(err))
		r.must("MustMultiplier", func() string { return fmt.Sprintf("%x", math.Float64bits(v.MustMultiplier())) })

		for _, f := range []float64{440, -440} {
			got, err := v.ApplyTo(f)
			r.add("apply %x %s", math.Float64bits(got), errName(err))
			r.must("MustApplyTo", func() string { return fmt.Sprintf("%x", math.Float64bits(v.MustApplyTo(f))) })
		}

		for _, d := range []int{1000, 0} {
			got, err := v.RoundTo(d)
			r.add("round %d %s %s", d, got, errName(err))
			r.must("MustRoundTo", func() string { return v.MustRoundTo(d).String() })
		}

		for _, k := range []int{3, -2, 7501} {
			got, err := v.Mul(k)
			r.add("mul %d %s %s", k, got, errName(err))
			r.must("MustMul", func() string { return v.MustMul(k).String() })
		}

		for _, n := range []int{7, -4, 0} {
			got, err := v.Div(n)
			r.add("div %d %s %s", n, got, errName(err))
			r.must("MustDiv", func() string { return v.MustDiv(n).String() })
		}

		for _, p := range [][2]int{{5, 3}, {1, 0}} {
			got, err := v.Scale(p[0], p[1])
			r.add("scale %v %s %s", p, got, errName(err))
			r.must("MustScale", func() string { return v.MustScale(p[0], p[1]).String() })
		}

		for _, x := range []*big.Rat{big.NewRat(-2, 9), nil} {
			got, err := v.ScaleRat(x)
			r.add("scale rat %s %s %s", x, got, errName(err))
			r.must("MustScaleRat", func() string { return v.MustScaleRat(x).String() })
		}

		parsed, err := cents.Parse(v.String())
		r.add("parse %s %s", parsed, errName(err))

		scala, scalaRest, err := cents.ParseScala(v.Scala(5) + " x")
		r.add("scala %s %q %s", scala, scalaRest, errName(err))

		approx, approxRest, err := cents.ParseScalaApprox(v.Scala(5) + " x")
		r.add("scala approx %s %q %s", approx, approxRest, errName(err))

		text, err := v.MarshalText()
		r.add("marshal %s %s", text, errName(err))

		// AppendText only with buffers of this goroutine (K-33, KR-C16): a
		// buffer with spare capacity, made here with an explicit capacity,
		// and nil. A buffer with spare capacity shared by goroutines is
		// data of the caller that the caller synchronizes.
		buf := make([]byte, len("prefix "), 64)
		copy(buf, "prefix ")
		appended, err := v.AppendText(buf)
		r.add("append %s %s", appended, errName(err))

		appended, err = v.AppendText(nil)
		r.add("append nil %s %s", appended, errName(err))

		var own cents.Cents
		err = own.UnmarshalText(text)
		r.add("unmarshal %s %s", own, errName(err))

		for j, b := range vs {
			sum, errSum := v.Add(b)
			diff, errDiff := v.Sub(b)
			r.add("pair %d %d %s %s %s %s %d %v %v", i, j, sum, errName(errSum), diff, errName(errDiff), v.Compare(b), v.Equal(b), v == b)

			if j%4 != i%4 {
				continue
			}

			r.must("MustAdd", func() string { return v.MustAdd(b).String() })
			r.must("MustSub", func() string { return v.MustSub(b).String() })

			near, err := v.Near(b, cents.Semitone())
			r.add("near %d %d %v %s", i, j, near, errName(err))
			r.must("MustNear", func() string { return fmt.Sprint(v.MustNear(b, cents.Semitone())) })

			q, errQ := v.Quo(b)
			r.add("quo %d %d %s", i, j, errName(errQ))
			r.quotient("quotient", q)
			r.must("MustQuo", func() string { q := v.MustQuo(b); return q.Decimal(12) })
		}
	}

	keys := map[cents.Cents]int{}
	for i, v := range vs {
		if _, ok := keys[v]; !ok {
			keys[v] = i
		}
	}

	for i, v := range vs {
		r.add("key %d %d", i, keys[v])
	}
}

// fromScratch creates the values and runs all operations on them.
func fromScratch() []string {
	var r recorder

	vs := creations(&r)
	operations(&r, vs)

	return r.out
}

// onShared runs all operations on the shared values of S.
func onShared(set []named) []string {
	vs := make([]cents.Cents, 0, len(set))
	for _, a := range set {
		vs = append(vs, a.v)
	}

	var r recorder

	operations(&r, vs)

	return r.out
}

// runConcurrently starts n workers at once (a closed channel releases them)
// and returns their results.
func runConcurrently(n int, work func() []string) [][]string {
	got := make([][]string, n)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			<-start
			got[i] = work()
		})
	}

	close(start)
	wg.Wait()

	return got
}

// assertSameLines reports the first line where the results of a worker
// differ from the expected results, or a different number of lines.
func assertSameLines(t *testing.T, want, got []string, msgAndArgs ...any) bool {
	t.Helper()

	for i := range min(len(want), len(got)) {
		if want[i] != got[i] {
			return assert.Fail(t, fmt.Sprintf("line %d differs:\nwant %q\ngot  %q", i, want[i], got[i]), msgAndArgs...)
		}
	}

	if len(want) != len(got) {
		return assert.Fail(t, fmt.Sprintf("%d lines, want %d", len(got), len(want)), msgAndArgs...)
	}

	return true
}

// K-33, NF-2: concurrent calls of all operations (creation, the named values
// and units, arithmetic, comparison, values, quotients, frequencies, text,
// parsing and the Must variants) give the same results as sequential calls;
// run under the race detector (make test, CI).
//
// The test does not call t.Parallel: it runs in the sequential phase of the
// package tests, before the other tests (they pause at t.Parallel, and the
// test of allocations is in a later file), so the first use of the package
// in the process happens in its workers at once, and lazy package state is
// built concurrently. The guarantee of the first use holds only while no
// test file initializes package-level variables with calls of the package,
// every test in a file sorted before concurrency_test.go calls t.Parallel
// before any call of the package, and only in the first iteration of
// -count; with -shuffle or -run the order may differ as well. Without the
// guarantee the test is still a check of concurrent use. AppendText is
// called only with buffers of each worker (K-33).
//
// A failure reports, for each worker, the first line that differs from the
// expected results.
func TestConcurrentUse(t *testing.T) {
	workers := min(runtime.GOMAXPROCS(0)*4, 32)

	// First use: every worker creates its own values from the same inputs.
	// The expected results are computed after the workers, in the test
	// goroutine.
	got := runConcurrently(workers, fromScratch)
	want := fromScratch()

	for i := range workers {
		assertSameLines(t, want, got[i], "first use, worker %d", i)
	}

	// Shared values: all workers use the same values of S.
	set := setS(t)
	got = runConcurrently(workers, func() []string { return onShared(set) })
	want = onShared(set)

	for i := range workers {
		assertSameLines(t, want, got[i], "shared values, worker %d", i)
	}
}
