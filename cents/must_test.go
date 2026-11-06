package cents_test

import (
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustCase is one call of an operation with an error and of its Must
// variant on the same inputs. want is nil for a successful call, otherwise
// the only sentinel error both report. Results are normalized by results.
type mustCase struct {
	op   string // the Must variant, e.g. "MustAdd"
	name string
	want error
	call func() ([]any, error)
	must func() []any
}

// results normalizes results for comparison: a Quotient (not comparable)
// becomes its exact value and decimal view.
func results(vs ...any) []any {
	out := make([]any, 0, len(vs))
	for _, v := range vs {
		if q, ok := v.(cents.Quotient); ok {
			exact, isExact := q.Exact()
			s := "irrational"
			if isExact && exact != nil {
				s = exact.RatString()
			}

			out = append(out, "quotient "+s+" "+q.Decimal(30))

			continue
		}

		out = append(out, v)
	}

	return out
}

// mustOperations returns the 23 Must variants of the architecture.
func mustOperations() []string {
	return []string{
		"MustFromInt", "MustFromFraction", "MustFromFloat", "MustStep", "MustFromRatio",
		"MustFromRatioFloat", "MustBetween", "MustFromComposition", "MustParse", "MustParseScala",
		"MustParseScalaApprox", "MustAdd", "MustSub", "MustMul", "MustDiv", "MustScale",
		"MustScaleRat", "MustRoundTo", "MustNearestStep", "MustNear", "MustQuo",
		"MustMultiplier", "MustApplyTo",
	}
}

// mustCases returns, for each of the 23 operations with an error, calls on
// valid inputs and on the failing inputs of K-2...K-30 and K-40...K-42.
func mustCases(t testing.TB) []mustCase {
	t.Helper()

	var (
		errRange   = cents.ErrOutOfRange
		errComplex = cents.ErrTooComplex
		errPrime   = cents.ErrPrimeLimit
		errArg     = cents.ErrInvalidArgument
		errText    = cents.ErrInvalidText
	)

	top := noErr(cents.FromInt(9000000))(t)
	r32 := noErr(cents.FromRatio(3, 2))(t)
	tiny := noErr(cents.Parse("0." + strings.Repeat("0", 29) + "1"))(t) // denominator 10^30
	third := noErr(cents.FromFraction(1, 3))(t)
	micro := noErr(cents.FromFraction(1, 1000000))(t)
	smallStep := noErr(micro.Add(noErr(cents.Step(1, 999983))(t)))(t)
	bigExp := fromComp(t, bigRatC(new(big.Rat).SetInt(bigInt("-190195500086538741774448"))).withExp(3, new(big.Rat).SetInt(pow10(20))))
	digits101 := "1" + strings.Repeat("0", 100)
	milliStep := noErr(cents.Step(1, 1000))(t)

	c := func(name string, f func() (cents.Cents, error), m func() cents.Cents, want error, op string) mustCase {
		return mustCase{
			op: op, name: name, want: want,
			call: func() ([]any, error) { v, err := f(); return results(v), err },
			must: func() []any { return results(m()) },
		}
	}

	var cases []mustCase

	for _, tc := range []struct {
		n    int
		want error
	}{{5, nil}, {9000000, nil}, {9000001, errRange}, {minInt, errRange}, {maxInt, errRange}} {
		cases = append(cases, c("FromInt", func() (cents.Cents, error) { return cents.FromInt(tc.n) },
			func() cents.Cents { return cents.MustFromInt(tc.n) }, tc.want, "MustFromInt"))
	}

	for _, tc := range []struct {
		num, den int
		want     error
	}{{1200, 19, nil}, {1, 0, errArg}, {0, 0, errArg}, {minInt, 0, errArg}, {maxInt, 1, errRange}} {
		cases = append(cases, c("FromFraction", func() (cents.Cents, error) { return cents.FromFraction(tc.num, tc.den) },
			func() cents.Cents { return cents.MustFromFraction(tc.num, tc.den) }, tc.want, "MustFromFraction"))
	}

	for _, tc := range []struct {
		x    float64
		want error
	}{{7.1, nil}, {5e-7, nil}, {math.NaN(), errArg}, {math.Inf(1), errArg}, {math.Inf(-1), errArg}, {9000001, errRange}} {
		cases = append(cases, c("FromFloat", func() (cents.Cents, error) { return cents.FromFloat(tc.x) },
			func() cents.Cents { return cents.MustFromFloat(tc.x) }, tc.want, "MustFromFloat"))
	}

	for _, tc := range []struct {
		k, n int
		want error
	}{{7, 12, nil}, {maxInt, maxInt, nil}, {1, 0, errArg}, {1, -12, errArg}, {90001, 12, errRange}} {
		cases = append(cases, c("Step", func() (cents.Cents, error) { return cents.Step(tc.k, tc.n) },
			func() cents.Cents { return cents.MustStep(tc.k, tc.n) }, tc.want, "MustStep"))
	}

	for _, tc := range []struct {
		a, b int
		want error
	}{{3, 2, nil}, {111, 74, nil}, {37, 32, errPrime}, {415, 440, errPrime}, {maxInt, 1, errPrime},
		{0, 1, errArg}, {-3, 2, errArg}, {3, 0, errArg}} {
		cases = append(cases, c("FromRatio", func() (cents.Cents, error) { return cents.FromRatio(tc.a, tc.b) },
			func() cents.Cents { return cents.MustFromRatio(tc.a, tc.b) }, tc.want, "MustFromRatio"))
	}

	for _, tc := range []struct {
		r    float64
		want error
	}{{1.5, nil}, {0, errArg}, {-1, errArg}, {math.NaN(), errArg}, {math.Inf(1), errArg}, {math.Inf(-1), errArg}} {
		cases = append(cases, c("FromRatioFloat", func() (cents.Cents, error) { return cents.FromRatioFloat(tc.r) },
			func() cents.Cents { return cents.MustFromRatioFloat(tc.r) }, tc.want, "MustFromRatioFloat"))
	}

	for _, tc := range []struct {
		f1, f2 float64
		want   error
	}{{440, 442, nil}, {1e-300, 1e300, nil}, {0, 440, errArg}, {-440, 440, errArg}, {math.NaN(), 440, errArg},
		{440, math.Inf(1), errArg}, {440, 0, errArg}, {440, math.Inf(-1), errArg}} {
		cases = append(cases, c("Between", func() (cents.Cents, error) { return cents.Between(tc.f1, tc.f2) },
			func() cents.Cents { return cents.MustBetween(tc.f1, tc.f2) }, tc.want, "MustBetween"))
	}

	for _, tc := range []struct {
		name     string
		num, den *big.Int
		ps       []cents.Power
		want     error
	}{
		{"ok", big.NewInt(-1200), nil, []cents.Power{pw(3, 1, 1)}, nil},
		{"base 37", nil, nil, []cents.Power{pw(37, 1, 1)}, errPrime},
		{"base 37 exponent 0", nil, nil, []cents.Power{{Base: 37}}, errPrime},
		{"base 4", nil, nil, []cents.Power{pw(4, 1, 1)}, errArg},
		{"base 0", nil, nil, []cents.Power{pw(0, 1, 1)}, errArg},
		{"den 0", big.NewInt(1), big.NewInt(0), nil, errArg},
		{"exponent den 0", nil, nil, []cents.Power{pw(3, 1, 0)}, errArg},
		{"(0; 3 → 6000)", nil, nil, []cents.Power{pw(3, 6000, 1)}, errRange},
		{"(9000001; —)", big.NewInt(9000001), nil, nil, errRange},
		{"exponent den 10^30+1", nil, nil, []cents.Power{{Base: 3, Num: big.NewInt(1), Den: new(big.Int).Add(pow10(30), big.NewInt(1))}}, errComplex},
		{"(1/0; 37 → 1)", big.NewInt(1), big.NewInt(0), []cents.Power{pw(37, 1, 1)}, errArg},
		{"(0; 37 → 1, 3 → 6000)", nil, nil, []cents.Power{pw(37, 1, 1), pw(3, 6000, 1)}, errPrime},
	} {
		cases = append(cases, c("FromComposition "+tc.name,
			func() (cents.Cents, error) { return cents.FromComposition(tc.num, tc.den, tc.ps...) },
			func() cents.Cents { return cents.MustFromComposition(tc.num, tc.den, tc.ps...) }, tc.want, "MustFromComposition"))
	}

	parseInputs := []struct {
		text string
		want error
	}{
		{"3:2", nil}, {"+30 ¢", nil}, {"1200/19 cents", nil}, {repeatFactors("2", "2", 63), nil},
		{"", errText}, {"+", errText}, {"¢", errText}, {"+-30", errText}, {"1,5", errText}, {"30 ct", errText},
		{"1e3", errText}, {"408.", errText}, {".5", errText}, {" 30", errText}, {"30  cents", errText},
		{"+3:2", errText}, {"3:2 ¢", errText}, {"3 : 2", errText}, {"3:2^2", errText}, {"3/2", errText},
		{"0:1", errText}, {"3:0", errText}, {"1/0 ¢", errText}, {"37:32", errPrime}, {"83:88", errPrime},
		{"9000000.000001", errRange}, {"-9000001 cents", errRange}, {digits101, errComplex},
		{"0." + strings.Repeat("0", 30) + "1", errComplex}, {repeatFactors("2", "2", 64), errComplex},
		{repeatFactors("2^0", "2^0", 64), errComplex}, {"37" + strings.Repeat("0", 100) + ":1", errComplex},
	}
	for _, tc := range parseInputs {
		cases = append(cases, c("Parse "+tc.text, func() (cents.Cents, error) { return cents.Parse(tc.text) },
			func() cents.Cents { return cents.MustParse(tc.text) }, tc.want, "MustParse"))
	}

	scalaInputs := []struct {
		line   string
		want   error
		approx error
	}{
		{" 5/4   E", nil, nil}, {"  100.0 cents", nil, nil}, {"37/32", errPrime, nil},
		{"-3/2", errText, errText}, {"0/1", errText, errText}, {"3/0", errText, errText}, {"+5.0", errText, errText},
		{"−5.0", errText, errText}, {"5/4E", errText, errText}, {"408.0.5", errText, errText}, {"", errText, errText},
		{"   ", errText, errText}, {digits101 + "/3", errComplex, errComplex},
		{"0." + strings.Repeat("0", 30) + "1", errComplex, errComplex}, {"9000000.000001", errRange, errRange},
	}
	for _, tc := range scalaInputs {
		cases = append(cases,
			mustCase{
				op: "MustParseScala", name: "ParseScala " + tc.line, want: tc.want,
				call: func() ([]any, error) { v, rest, err := cents.ParseScala(tc.line); return results(v, rest), err },
				must: func() []any { v, rest := cents.MustParseScala(tc.line); return results(v, rest) },
			},
			mustCase{
				op: "MustParseScalaApprox", name: "ParseScalaApprox " + tc.line, want: tc.approx,
				call: func() ([]any, error) { v, rest, err := cents.ParseScalaApprox(tc.line); return results(v, rest), err },
				must: func() []any { v, rest := cents.MustParseScalaApprox(tc.line); return results(v, rest) },
			})
	}

	binary := []struct {
		name string
		a, b cents.Cents
		add  error
		sub  error
	}{
		{"R(3:2), R(4:3)", r32, noErr(cents.FromRatio(4, 3))(t), nil, nil},
		{"<9000000>, <1/1000000>", top, micro, errRange, nil},
		{"<-9000000>, <1>", top.Neg(), cents.Cent(), nil, errRange},
		{"<9000000>, <-9000000>", top, top.Neg(), nil, errRange},
		{"<10^-30>, <1/3>", tiny, third, errComplex, errComplex},
	}
	for _, tc := range binary {
		cases = append(cases,
			c("Add "+tc.name, func() (cents.Cents, error) { return tc.a.Add(tc.b) },
				func() cents.Cents { return tc.a.MustAdd(tc.b) }, tc.add, "MustAdd"),
			c("Sub "+tc.name, func() (cents.Cents, error) { return tc.a.Sub(tc.b) },
				func() cents.Cents { return tc.a.MustSub(tc.b) }, tc.sub, "MustSub"))
	}

	for _, tc := range []struct {
		name string
		a    cents.Cents
		k    int
		want error
	}{
		{"R(3:2) * 12", r32, 12, nil}, {"0 * MaxInt", cents.Zero(), maxInt, nil}, {"<1200> * 7501", cents.Octave(), 7501, errRange},
		{"<1> * MaxInt", cents.Cent(), maxInt, errRange}, {"e_3 10^20 * 10^5", bigExp, 100000, errComplex},
	} {
		cases = append(cases,
			c("Mul "+tc.name, func() (cents.Cents, error) { return tc.a.Mul(tc.k) },
				func() cents.Cents { return tc.a.MustMul(tc.k) }, tc.want, "MustMul"),
			c("Scale "+tc.name, func() (cents.Cents, error) { return tc.a.Scale(tc.k, 1) },
				func() cents.Cents { return tc.a.MustScale(tc.k, 1) }, tc.want, "MustScale"),
			c("ScaleRat "+tc.name, func() (cents.Cents, error) { return tc.a.ScaleRat(big.NewRat(int64(tc.k), 1)) },
				func() cents.Cents { return tc.a.MustScaleRat(big.NewRat(int64(tc.k), 1)) }, tc.want, "MustScaleRat"))
	}

	for _, tc := range []struct {
		name string
		a    cents.Cents
		n    int
		want error
	}{
		{"SYNT / 4", cents.SyntonicComma(), 4, nil}, {"<-9000000> / 2", top.Neg(), 2, nil}, {"R(3:2) / 0", r32, 0, errArg},
		{"0 / 0", cents.Zero(), 0, errArg}, {"<10^-30> / 3", tiny, 3, errComplex},
	} {
		cases = append(cases,
			c("Div "+tc.name, func() (cents.Cents, error) { return tc.a.Div(tc.n) },
				func() cents.Cents { return tc.a.MustDiv(tc.n) }, tc.want, "MustDiv"),
			c("Scale 1/n "+tc.name, func() (cents.Cents, error) { return tc.a.Scale(1, tc.n) },
				func() cents.Cents { return tc.a.MustScale(1, tc.n) }, tc.want, "MustScale"))
	}

	cases = append(cases,
		c("Scale 7500001 units", func() (cents.Cents, error) { return milliStep.Scale(7500001, 1) },
			func() cents.Cents { return milliStep.MustScale(7500001, 1) }, errRange, "MustScale"),
		c("ScaleRat 1/10^30 millioctave", func() (cents.Cents, error) {
			return cents.Millioctave().ScaleRat(new(big.Rat).SetFrac(big.NewInt(1), pow10(30)))
		}, func() cents.Cents {
			return cents.Millioctave().MustScaleRat(new(big.Rat).SetFrac(big.NewInt(1), pow10(30)))
		}, errComplex, "MustScaleRat"),
		c("ScaleRat nil", func() (cents.Cents, error) { return r32.ScaleRat(nil) },
			func() cents.Cents { return r32.MustScaleRat(nil) }, nil, "MustScaleRat"),
	)

	for _, tc := range []struct {
		d    int
		want error
	}{{1000, nil}, {1, nil}, {0, errArg}, {-1, errArg}, {minInt, errArg}} {
		cases = append(cases, c("RoundTo", func() (cents.Cents, error) { return r32.RoundTo(tc.d) },
			func() cents.Cents { return r32.MustRoundTo(tc.d) }, tc.want, "MustRoundTo"))
	}

	for _, tc := range []struct {
		name string
		a    cents.Cents
		n    int
		want error
	}{
		{"R(3:2) in 53", r32, 53, nil}, {"<1200> in MaxInt", cents.Octave(), maxInt, nil}, {"N = 0", r32, 0, errArg},
		{"N = -12", r32, -12, errArg}, {"<9000000> in MaxInt", top, maxInt, errRange},
		{"<0.000001> + E(1,999983) in MaxInt", smallStep, maxInt, errComplex},
	} {
		cases = append(cases, mustCase{
			op: "MustNearestStep", name: "NearestStep " + tc.name, want: tc.want,
			call: func() ([]any, error) { k, d, err := tc.a.NearestStep(tc.n); return results(k, d), err },
			must: func() []any { k, d := tc.a.MustNearestStep(tc.n); return results(k, d) },
		})
	}

	for _, tc := range []struct {
		name string
		tol  cents.Cents
		want error
	}{
		{"t = 2", noErr(cents.FromInt(2))(t), nil}, {"t = 1.955", noErr(cents.FromFraction(1955, 1000))(t), nil},
		{"t < 0", noErr(cents.FromInt(-1))(t), errArg},
	} {
		c700 := noErr(cents.FromInt(700))(t)
		cases = append(cases, mustCase{
			op: "MustNear", name: "Near " + tc.name, want: tc.want,
			call: func() ([]any, error) { ok, err := r32.Near(c700, tc.tol); return results(ok), err },
			must: func() []any { return results(r32.MustNear(c700, tc.tol)) },
		})
	}

	for _, tc := range []struct {
		name string
		a, b cents.Cents
		want error
	}{
		{"<1200> / HOLDER", cents.Octave(), cents.HoldrianComma(), nil}, {"<1200> / PYTH", cents.Octave(), cents.PythagoreanComma(), nil},
		{"R(3:2) / 0", r32, cents.Zero(), errArg}, {"0 / zero value", cents.Zero(), cents.Cents{}, errArg},
	} {
		cases = append(cases, mustCase{
			op: "MustQuo", name: "Quo " + tc.name, want: tc.want,
			call: func() ([]any, error) { q, err := tc.a.Quo(tc.b); return results(q), err },
			must: func() []any { return results(tc.a.MustQuo(tc.b)) },
		})
	}

	for _, tc := range []struct {
		n    int
		want error
	}{{1200, nil}, {700, nil}, {-1226400, nil}, {1300000, errRange}, {-1300000, errRange}, {-1226401, errRange}} {
		v := noErr(cents.FromInt(tc.n))(t)
		cases = append(cases, mustCase{
			op: "MustMultiplier", name: "Multiplier", want: tc.want,
			call: func() ([]any, error) { m, err := v.Multiplier(); return results(m), err },
			must: func() []any { return results(v.MustMultiplier()) },
		})
	}

	for _, tc := range []struct {
		name string
		c    cents.Cents
		freq float64
		want error
	}{
		{"440 by <1200>", cents.Octave(), 440, nil}, {"2^-1074 by <62400>", noErr(cents.FromInt(62400))(t), 0x1p-1074, nil},
		{"1e300 by <120000>", noErr(cents.FromInt(120000))(t), 1e300, errRange},
		{"2^-1022 by <-1200>", noErr(cents.FromInt(-1200))(t), 0x1p-1022, errRange},
		{"2^-1074 by 0", cents.Zero(), 0x1p-1074, errRange},
		{"freq 0", cents.Zero(), 0, errArg}, {"freq -440", cents.Zero(), -440, errArg},
		{"freq NaN", cents.Zero(), math.NaN(), errArg}, {"freq +Inf", cents.Zero(), math.Inf(1), errArg},
	} {
		cases = append(cases, mustCase{
			op: "MustApplyTo", name: "ApplyTo " + tc.name, want: tc.want,
			call: func() ([]any, error) { f, err := tc.c.ApplyTo(tc.freq); return results(f), err },
			must: func() []any { return results(tc.c.MustApplyTo(tc.freq)) },
		})
	}

	return cases
}

// TestMustCasesCoverAllOperations checks that every one of the 23 Must
// variants has valid and failing calls in the table.
func TestMustCasesCoverAllOperations(t *testing.T) {
	t.Parallel()

	ops := mustOperations()
	require.Len(t, ops, 23)

	valid, failing := map[string]bool{}, map[string]bool{}
	for _, c := range mustCases(t) {
		require.Contains(t, ops, c.op, "unknown Must variant in the cases")

		if c.want == nil {
			valid[c.op] = true
		} else {
			failing[c.op] = true
		}
	}

	for _, op := range ops {
		assert.True(t, valid[op], "%s has a valid call", op)
		assert.True(t, failing[op], "%s has a failing call", op)
	}
}

// K-31, NF-7, E-6: a Must variant panics exactly when the operation returns
// an error, with that error as the panic value (errors.Is works on the
// recovered error), and otherwise returns the same results.
func TestMustVariants(t *testing.T) {
	t.Parallel()

	for _, c := range mustCases(t) {
		t.Run(c.name+" "+errName(c.want), func(t *testing.T) {
			t.Parallel()

			var (
				got []any
				err error
			)
			require.NotPanics(t, func() { got, err = c.call() }, "the variant with an error never panics")

			if c.want != nil {
				assertOnlyError(t, err, c.want, c.name)
				assertZeroResults(t, got, c.name)
				assertPanicsWithErrorIs(t, c.want, func() { c.must() }, c.op)

				return
			}

			require.NoError(t, err, c.name)

			var fromMust []any
			require.NotPanics(t, func() { fromMust = c.must() }, c.op)
			assert.Equal(t, got, fromMust, c.op)
		})
	}
}

// assertZeroResults checks that every result of a failed call is the zero
// value of its type ("0" for a Quotient).
func assertZeroResults(t testing.TB, vs []any, desc string) {
	t.Helper()

	for i, v := range vs {
		switch x := v.(type) {
		case cents.Cents:
			assert.True(t, x == cents.Cents{}, "%s: result %d", desc, i)
		case string:
			if strings.HasPrefix(x, "quotient ") {
				assert.Equal(t, results(cents.Quotient{})[0], x, "%s: result %d", desc, i)
			} else {
				assert.Empty(t, x, "%s: result %d", desc, i)
			}
		default:
			assert.Zero(t, x, "%s: result %d", desc, i)
		}
	}
}
