package fraction_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// benchOp is one operation of the T-2 list with the inputs of its benchmark.
type benchOp struct {
	name   string
	family benchFamily
	run    func(in benchInput) error
	// want is the oracle expectation of a Fraction result; nil when the
	// operation returns no Fraction.
	want func(in benchInput) *wideWant
	got  func(in benchInput) (fraction.Fraction, error)
}

func benchWantArith(op wideArith, mode fraction.DenomMode) func(in benchInput) *wideWant {
	return func(in benchInput) *wideWant {
		w := wantArith(op, in.x, in.y, mode)
		return &w
	}
}

func fracRun(got func(in benchInput) (fraction.Fraction, error)) func(in benchInput) error {
	return func(in benchInput) error {
		_, err := got(in)
		return err
	}
}

// benchOps lists the T-2 operations in the order of the benchmark list.
func benchOps() []benchOp {
	arith := func(name string, op wideArith, family benchFamily, mode fraction.DenomMode, call func(in benchInput) (fraction.Fraction, error)) benchOp {
		return benchOp{name: name, family: family, run: fracRun(call), want: benchWantArith(op, mode), got: call}
	}
	frac := func(name string, family benchFamily, want func(in benchInput) wideWant, call func(in benchInput) (fraction.Fraction, error)) benchOp {
		return benchOp{name: name, family: family, run: fracRun(call), got: call, want: func(in benchInput) *wideWant {
			w := want(in)
			return &w
		}}
	}
	plain := func(name string, family benchFamily, run func(in benchInput) error) benchOp {
		return benchOp{name: name, family: family, run: run}
	}
	return []benchOp{
		arith("Add", wideAdd, benchBinary, fraction.DenomReduced, func(in benchInput) (fraction.Fraction, error) { return in.x.Add(in.y) }),
		arith("Sub", wideSub, benchBinary, fraction.DenomReduced, func(in benchInput) (fraction.Fraction, error) { return in.x.Sub(in.y) }),
		arith("Mul", wideMul, benchBinary, fraction.DenomReduced, func(in benchInput) (fraction.Fraction, error) { return in.x.Mul(in.y) }),
		arith("Div", wideDiv, benchBinary, fraction.DenomReduced, func(in benchInput) (fraction.Fraction, error) { return in.x.Div(in.y) }),
		arith("AddWithUnreducedLCM", wideAdd, benchLCM, fraction.DenomUnreducedLCM, func(in benchInput) (fraction.Fraction, error) {
			return in.x.AddWith(in.y, fraction.DenomUnreducedLCM)
		}),
		arith("AddWithLargerUnit", wideAdd, benchUnit, fraction.DenomLargerUnit, func(in benchInput) (fraction.Fraction, error) {
			return in.x.AddWith(in.y, fraction.DenomLargerUnit)
		}),
		frac("AddInDenominator", benchUnit, func(in benchInput) wideWant { return wantInDenominator(wideAdd, in.x, in.y, in.den) },
			func(in benchInput) (fraction.Fraction, error) { return in.x.AddInDenominator(in.y, in.den) }),
		plain("Compare", benchBinary, func(in benchInput) error {
			_ = in.x.Compare(in.y)
			return nil
		}),
		plain("Equal", benchBinary, func(in benchInput) error {
			_ = in.x.Equal(in.y)
			return nil
		}),
		plain("QuoRem", benchDivMod, func(in benchInput) error {
			_, _, err := in.x.QuoRem(in.y)
			return err
		}),
		plain("DivModFloor", benchDivMod, func(in benchInput) error {
			_, _, err := in.x.DivModFloor(in.y)
			return err
		}),
		plain("IsMultipleOf", benchDivMod, func(in benchInput) error {
			_, err := in.x.IsMultipleOf(in.y)
			return err
		}),
		frac("Reduce", benchReduce, func(in benchInput) wideWant { return wFromRat(wRat(in.x)) },
			func(in benchInput) (fraction.Fraction, error) { return in.x.Reduce(), nil }),
		plain("IsReduced", benchReduce, func(in benchInput) error {
			_ = in.x.IsReduced()
			return nil
		}),
		plain("IsInteger", benchSplit, func(in benchInput) error {
			_ = in.x.IsInteger()
			return nil
		}),
		plain("Split", benchSplit, func(in benchInput) error {
			_, _, err := in.x.Split()
			return err
		}),
		plain("SplitFloor", benchSplit, func(in benchInput) error {
			_, _, err := in.x.SplitFloor()
			return err
		}),
		plain("RoundToInt", benchSplit, func(in benchInput) error {
			_, err := in.x.RoundToInt(fraction.RoundHalfToEven)
			return err
		}),
		frac("RoundToDenominator", benchRoundDen, func(in benchInput) wideWant {
			return wantRoundToBigDenominator(in.x, big.NewInt(int64(in.den)), fraction.RoundHalfToEven)
		}, func(in benchInput) (fraction.Fraction, error) {
			return in.x.RoundToDenominator(in.den, fraction.RoundHalfToEven)
		}),
		frac("WithDenominator", benchWithDen, func(in benchInput) wideWant {
			return wantWithBigDenominator(in.x, big.NewInt(int64(in.den)))
		}, func(in benchInput) (fraction.Fraction, error) { return in.x.WithDenominator(in.den) }),
	}
}

// benchGBits is the approximate exponent of the common factor g of target pairs.
var benchGBits = map[string]int{"w63": 16, "w128": 32, "w255": 60}

// benchComponentBits checks the bit lengths of a component against its class.
func benchComponentBits(t *testing.T, class, what string, v *big.Int, isDen bool) {
	t.Helper()
	bits := v.BitLen()
	var lo, hi int
	switch class {
	case "w31":
		lo, hi = 2, 31
	case "alt":
		lo, hi = 39, 42
		if isDen {
			lo, hi = 2, 16
		}
	case "w63":
		lo, hi = 33, 63
	case "w128":
		lo, hi = 65, 128
	default:
		lo, hi = 193, 255
	}
	assert.Truef(t, bits >= lo && bits <= hi, "%s has %d bits, class %s wants %d..%d", what, bits, class, lo, hi)
}

// benchNontrivialGCD checks that NOD(a, b) is above 1 and differs from the smaller value.
func benchNontrivialGCD(t *testing.T, what string, a, b *big.Int) *big.Int {
	t.Helper()
	g := wGCD(a, b)
	smaller := a
	if new(big.Int).Abs(b).CmpAbs(a) < 0 {
		smaller = b
	}
	assert.Truef(t, g.Cmp(big.NewInt(1)) > 0, "%s: NOD must be above 1", what)
	assert.Truef(t, g.CmpAbs(smaller) != 0, "%s: NOD must differ from the smaller value", what)
	return g
}

// benchQuotient returns trunc(a / b) and a mod b for positive a, b.
func benchQuotient(a, b *big.Int) (q, r *big.Int) {
	return new(big.Int).QuoRem(new(big.Int).Abs(a), new(big.Int).Abs(b), new(big.Int))
}

// TestBenchmarkInputs checks the properties required of the benchmark inputs:
// the classes present, component widths, nontrivial reductions, the size of the
// common factor in target pairs, long and short quotients, the denominators,
// and success and oracle agreement of every benchmarked operation.
func TestBenchmarkInputs(t *testing.T) {
	t.Parallel()

	five := []string{"w31", "alt", "w63", "w128", "w255"}
	eight := []string{"w31", "alt", "w63", "w63-short", "w128", "w128-short", "w255", "w255-short"}
	wantClasses := map[benchFamily][]string{
		benchBinary: five, benchLCM: five, benchUnit: five, benchReduce: five, benchWithDen: five,
		benchDivMod: eight, benchSplit: eight, benchRoundDen: eight,
	}
	for family, want := range wantClasses {
		var got []string
		for _, in := range benchInputs(family) {
			got = append(got, in.class)
		}
		assert.Equal(t, want, got, "classes of family %d", family)
	}

	two48, two62, two8 := wPow2(48), wPow2(62), wPow2(8)

	t.Run("binary_families", func(t *testing.T) {
		t.Parallel()
		for _, family := range []benchFamily{benchBinary, benchLCM, benchUnit, benchDivMod} {
			for _, in := range benchInputs(family) {
				t.Run(fmt.Sprintf("%d/%s", family, in.class), func(t *testing.T) {
					n1, d1 := wNorm(in.x)
					n2, d2 := wNorm(in.y)
					class := in.baseClass()
					benchComponentBits(t, class, "n1", n1, false)
					benchComponentBits(t, class, "d1", d1, true)
					// The second operand of AddWith(DenomUnreducedLCM)/w255 is outside
					// the class by the decision of the architect.
					if family != benchLCM || class != "w255" {
						benchComponentBits(t, class, "n2", n2, false)
						benchComponentBits(t, class, "d2", d2, true)
					}

					g := wGCD(d1, d2)
					require.Truef(t, g.Cmp(big.NewInt(1)) > 0, "denominators must share g >= 2")
					if k, ok := benchGBits[class]; ok {
						assert.InDeltaf(t, k+1, g.BitLen(), 2, "common factor of the denominators must be near 2^%d", k)
					}

					sum := new(big.Int).Abs(new(big.Int).Add(wMul(n1, d2), wMul(n2, d1)))
					switch family {
					case benchBinary:
						diff := new(big.Int).Abs(new(big.Int).Sub(wMul(n1, d2), wMul(n2, d1)))
						benchNontrivialGCD(t, "Add", sum, wMul(d1, d2))
						benchNontrivialGCD(t, "Sub", diff, wMul(d1, d2))
						benchNontrivialGCD(t, "Mul", wMul(n1, n2), wMul(d1, d2))
						benchNontrivialGCD(t, "Div", wMul(n1, d2), wMul(d1, n2))
					case benchLCM:
						benchNontrivialGCD(t, "AddWith(DenomUnreducedLCM)", d1, d2)
					case benchUnit:
						benchNontrivialGCD(t, "AddWith(DenomLargerUnit)", sum, wMul(d1, d2))
						assert.Equal(t, g.Int64(), int64(in.den), "AddInDenominator den must be g")
					case benchDivMod:
						require.Positive(t, n2.Sign(), "the step must be positive for DivModFloor")
						divisor := new(big.Int).Abs(wMul(d1, n2))
						q, r := benchQuotient(wMul(n1, d2), divisor)
						require.NotZero(t, r.Sign(), "the remainder must be nonzero")
						benchNontrivialGCD(t, "QuoRem remainder", r, wMul(d1, d2))
						benchNontrivialGCD(t, "DivModFloor remainder", new(big.Int).Sub(divisor, r), wMul(d1, d2))
						switch {
						case in.short():
							assert.LessOrEqual(t, q.Cmp(two8), 0, "short quotient must be at most 2^8")
						case class == "w63" || class == "w128" || class == "w255":
							assert.GreaterOrEqual(t, q.Cmp(two48), 0, "long quotient must be at least 2^48")
							assert.LessOrEqual(t, q.Cmp(two62), 0, "long int quotient must be at most 2^62")
						}
						isMul, err := in.x.IsMultipleOf(in.y)
						require.NoError(t, err)
						assert.False(t, isMul, "IsMultipleOf must be false")
					}
				})
			}
		}
	})

	t.Run("single_families", func(t *testing.T) {
		t.Parallel()
		for _, family := range []benchFamily{benchReduce, benchWithDen, benchSplit, benchRoundDen} {
			for _, in := range benchInputs(family) {
				t.Run(fmt.Sprintf("%d/%s", family, in.class), func(t *testing.T) {
					n, d := wNorm(in.x)
					class := in.baseClass()
					benchComponentBits(t, class, "n", n, false)
					// The denominator of WithDenominator/w255 is outside the class by
					// the decision of the architect.
					if family != benchWithDen || class != "w255" {
						benchComponentBits(t, class, "d", d, true)
					}
					switch family {
					case benchReduce, benchWithDen:
						g := benchNontrivialGCD(t, "NOD(|n|, D)", n, d)
						if k, ok := benchGBits[class]; ok {
							assert.InDeltaf(t, k+1, g.BitLen(), 2, "g must be near 2^%d", k)
						}
						if family == benchWithDen {
							assert.LessOrEqual(t, in.den, 1<<62, "den must not exceed 2^62")
							assert.Zero(t, new(big.Int).Rem(big.NewInt(int64(in.den)), wRat(in.x).Denom()).Sign(),
								"den must be a multiple of the reduced denominator")
						}
					case benchSplit, benchRoundDen:
						dividend := n
						if family == benchRoundDen {
							dividend = wMul(n, big.NewInt(int64(in.den)))
							wantDenBits := map[string]int{"w31": 31, "alt": 17, "w63": 63, "w128": 63, "w255": 5}[class]
							if in.class == "w63-short" {
								wantDenBits = 31
							}
							assert.Equal(t, wantDenBits, big.NewInt(int64(in.den)).BitLen(), "den width")
						}
						q, r := benchQuotient(dividend, d)
						require.NotZero(t, r.Sign(), "the remainder must be nonzero")
						long := class == "w128" || class == "w255" || family == benchRoundDen && class == "w63"
						switch {
						case in.short():
							assert.LessOrEqual(t, q.Cmp(two8), 0, "short quotient must be at most 2^8")
						case long:
							assert.GreaterOrEqual(t, q.Cmp(two48), 0, "long quotient must be at least 2^48")
							if family == benchSplit {
								assert.LessOrEqual(t, q.Cmp(two62), 0, "long int quotient must be at most 2^62")
							}
						}
					}
				})
			}
		}
	})

	t.Run("operations", func(t *testing.T) {
		t.Parallel()
		for _, op := range benchOps() {
			for _, in := range benchInputs(op.family) {
				t.Run(op.name+"/"+in.class, func(t *testing.T) {
					require.NoError(t, op.run(in), "the benchmarked operation must succeed")
					if op.want != nil {
						got, err := op.got(in)
						wCheckFrac(t, op.name, got, err, *op.want(in))
					}
				})
			}
		}
		for _, family := range []benchFamily{benchBinary, benchDivMod} {
			for _, in := range benchInputs(family) {
				t.Run(fmt.Sprintf("oracle/%d/%s", family, in.class), func(t *testing.T) {
					checkWideBinary(t, in.x, in.y)
				})
			}
		}
		for _, family := range []benchFamily{benchReduce, benchSplit, benchRoundDen} {
			for _, in := range benchInputs(family) {
				t.Run(fmt.Sprintf("oracle/%d/%s", family, in.class), func(t *testing.T) {
					checkWideUnary(t, in.x)
				})
			}
		}
	})
}

// TestZeroAllocsByClass checks T-2: every operation of the list runs without
// heap allocations on every benchmark input of every class, including the
// "-short" inputs. It is not parallel: AllocsPerRun measures process-wide
// allocation counters.
func TestZeroAllocsByClass(t *testing.T) {
	for _, op := range benchOps() {
		for _, in := range benchInputs(op.family) {
			t.Run(op.name+"/"+in.class, func(t *testing.T) {
				require.NoError(t, op.run(in))
				allocs := testing.AllocsPerRun(100, func() {
					_ = op.run(in)
				})
				assert.Zero(t, allocs, "allocations per call")
			})
		}
	}
}
