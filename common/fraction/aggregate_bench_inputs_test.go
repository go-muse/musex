package fraction_test

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/go-muse/musex/common/fraction"
)

// Benchmark inputs of task 26.7 (architecture, "Tests", "Benchmarks"). The
// width classes are those of task 26.6 (w31, alt, w63, w128, w255); the values
// are products of the probable primes bp of bench_inputs_test.go, so every
// common factor is known by construction. TestAggregateBenchmarkInputs checks
// the required properties.

// agClasses are the width classes of task 26.6.
var agClasses = []string{"w31", "alt", "w63", "w128", "w255"}

// agSizes are the set sizes of the set benchmarks.
var agSizes = []int{2, 8, 64}

// agSetInput is one set of a set benchmark.
type agSetInput struct {
	class    string
	n        int
	overflow bool
	fs       []fraction.Fraction
}

// name returns the sub-benchmark name <class>/<n> or <class>/<n>-overflow.
func (in agSetInput) name() string {
	if in.overflow {
		return fmt.Sprintf("%s/%d-overflow", in.class, in.n)
	}
	return fmt.Sprintf("%s/%d", in.class, in.n)
}

// agPairInput is one operand pair of the GCD and LCM benchmarks.
type agPairInput struct {
	class string
	x, y  fraction.Fraction
}

// agInputSet holds every input family of the task 26.7 benchmarks.
type agInputSet struct {
	sum, product, commonDen     []agSetInput
	sumOverflow, prodOverflow   []agSetInput
	pairs                       []agPairInput
	skippedProduct              []string
	skippedSumOF, skippedProdOF []string
}

var agInputsOnce = sync.OnceValue(buildAgInputs)

// agInputs returns the inputs; the slices must not be modified.
func agInputs() *agInputSet {
	return agInputsOnce()
}

// agSumSet returns n elements (±g·aᵢ, g·q) with q from qs: the original
// components are in the class, the reduced values ±aᵢ/q are small (at most
// 31 bits), every denominator divides D = ∏qs ≤ 2^62, so every reduced prefix
// sum is in range and the fast path never continues exactly. Every second
// numerator is negative; every third element has a negative original
// denominator.
func agSumSet(n int, g *big.Int, aBits uint, qs []*big.Int) []fraction.Fraction {
	fs := make([]fraction.Fraction, n)
	for i := range n {
		num := benchMul(g, bp(aBits, i))
		den := benchMul(g, qs[i%len(qs)])
		if i%2 == 1 {
			num.Neg(num)
		}
		if i%3 == 2 {
			num.Neg(num)
			den.Neg(den)
		}
		fs[i] = benchFrac(num, den)
	}
	return fs
}

// agProductChain returns n elements cᵢ/cᵢ₊₁ with cⱼ = bp(bits, j): the reduced
// partial product is ±c₀/cᵢ₊₁, always in range, and every step of the fast
// path reduces by cᵢ. Every third numerator is negative.
func agProductChain(n int, bits uint) []fraction.Fraction {
	fs := make([]fraction.Fraction, n)
	for i := range n {
		num := bp(bits, i)
		if i%3 == 2 {
			num = benchNeg(num)
		}
		fs[i] = benchFrac(num, bp(bits, i+1))
	}
	return fs
}

// agProductAlt returns n elements (Bᵢ₊₁·tᵢ, Bᵢ) of class alt with Bⱼ =
// bp(15, j) and tᵢ = bp(23, i): numerators near 2^40, denominators up to 2^16;
// every step reduces by Bᵢ and the reduced partial product is ∏tⱼ·Bᵢ₊₁/B₀.
func agProductAlt(n int) []fraction.Fraction {
	fs := make([]fraction.Fraction, n)
	for i := range n {
		fs[i] = benchFrac(benchMul(bp(15, i+1), bp(23, i)), bp(15, i))
	}
	return fs
}

// agCommonDenSet returns n elements (±g·aᵢ, g·q) with q from qs: the original
// components are in the class, every element reduces by g, and the reduced
// denominators are the divisors q of D = ∏qs ≤ math.MaxInt.
func agCommonDenSet(n int, g *big.Int, aBits uint, qs []*big.Int) []fraction.Fraction {
	fs := make([]fraction.Fraction, n)
	for i := range n {
		num := benchMul(g, bp(aBits, i))
		if i%2 == 1 {
			num.Neg(num)
		}
		fs[i] = benchFrac(num, benchMul(g, qs[i%len(qs)]))
	}
	return fs
}

// agPair returns x = (g·h·a, g·k·b), y = (g·h·c, g·k·d): each operand reduces
// by g, the reduced numerators share h and the reduced denominators share k,
// so GCD = h/(k·b·d) and LCM = h·a·c/k. All values are positive, so the
// big.Int reference of LCM computes the same value.
func agPair(class string, g, h, k, a, b, c, d *big.Int) agPairInput {
	return agPairInput{
		class: class,
		x:     benchFrac(benchMul(g, h, a), benchMul(g, k, b)),
		y:     benchFrac(benchMul(g, h, c), benchMul(g, k, d)),
	}
}

// agPrimes returns bp(bits, from), …, bp(bits, from+count−1).
func agPrimes(bits uint, from, count int) []*big.Int {
	ps := make([]*big.Int, count)
	for i := range ps {
		ps[i] = bp(bits, from+i)
	}
	return ps
}

// agOverflowElement returns the i-th element (cᵢ, dᵢ) of the overflow sets of a class.
func agOverflowElement(class string, i int) fraction.Fraction {
	switch class {
	case "w31":
		return benchFrac(bp(30, 2*i), bp(30, 2*i+1))
	case "alt":
		return benchFrac(bp(39, i), bp(15, i))
	case "w63":
		return benchFrac(bp(62, 2*i), bp(62, 2*i+1))
	case "w128":
		return benchFrac(bp(126, 2*i), bp(126, 2*i+1))
	default:
		return benchFrac(bp(250, 2*i), bp(250, 2*i+1))
	}
}

// agOverflowSet returns a set of n elements whose result is in range while a
// prefix is out of range, so Sum (Product) continues exactly: e₁…eₖ up to the
// first prefix out of range, then the inverses −eⱼ (1/eⱼ) of the first
// elements until the value is in range again, then pairs eⱼ, −eⱼ (eⱼ, 1/eⱼ)
// and, for an odd remainder, 0/d₁ (c₁/c₁). The second result is false when n
// elements cannot hold the prefix and the cancellation. These sets are for the
// report only: the reciprocals of alt elements leave the widths of the class.
func agOverflowSet(class string, n int, product bool) ([]fraction.Fraction, bool) {
	exact := agExactSum
	inverse := func(f fraction.Fraction) fraction.Fraction { return f.MustNeg() }
	if product {
		exact = agExactProduct
		inverse = func(f fraction.Fraction) fraction.Fraction { return f.MustReciprocal() }
	}
	var fs, base []fraction.Fraction
	for agRatInRange(exact(fs)) {
		e := agOverflowElement(class, len(base))
		base = append(base, e)
		fs = append(fs, e)
	}
	for j := 0; !agRatInRange(exact(fs)); j++ {
		fs = append(fs, inverse(base[j]))
	}
	for j := 0; len(fs)+2 <= n; j++ {
		e := base[j%len(base)]
		fs = append(fs, e, inverse(e))
	}
	if len(fs) < n {
		num, den, _ := base[0].BigPair()
		if product {
			fs = append(fs, benchFrac(num, num))
		} else {
			fs = append(fs, benchFrac(big.NewInt(0), den))
		}
	}
	return fs, len(fs) == n
}

func buildAgInputs() *agInputSet {
	in := &agInputSet{}

	// Sum: g, the width of aᵢ and the divisors q of D (architecture, "Inputs of the sets").
	sumParams := map[string]struct {
		g     *big.Int
		aBits uint
		qs    []*big.Int
	}{
		"w31":  {bp(8, 0), 20, agPrimes(19, 100, 2)},
		"alt":  {bp(9, 0), 29, agPrimes(5, 0, 2)},
		"w63":  {bp(32, 0), 28, agPrimes(28, 100, 2)},
		"w128": {bp(96, 0), 28, agPrimes(28, 100, 2)},
		"w255": {bp(220, 0), 28, agPrimes(28, 100, 2)},
	}
	// CommonDenominator, CommonDenominatorBig: D = ∏qs ≤ math.MaxInt.
	cdParams := map[string]struct {
		g     *big.Int
		aBits uint
		qs    []*big.Int
	}{
		"w31":  {bp(8, 0), 20, agPrimes(19, 100, 3)},
		"alt":  {bp(3, 0), 35, agPrimes(10, 0, 5)},
		"w63":  {bp(16, 0), 44, agPrimes(19, 100, 3)},
		"w128": {bp(100, 0), 24, agPrimes(19, 100, 3)},
		"w255": {bp(225, 0), 24, agPrimes(19, 100, 3)},
	}
	productBits := map[string]uint{"w31": 30, "w63": 62, "w128": 126, "w255": 250}

	for _, class := range agClasses {
		sp, cp := sumParams[class], cdParams[class]
		for _, n := range agSizes {
			in.sum = append(in.sum, agSetInput{class: class, n: n, fs: agSumSet(n, sp.g, sp.aBits, sp.qs)})
			in.commonDen = append(in.commonDen, agSetInput{class: class, n: n, fs: agCommonDenSet(n, cp.g, cp.aBits, cp.qs)})
			switch {
			case class != "alt":
				in.product = append(in.product, agSetInput{class: class, n: n, fs: agProductChain(n, productBits[class])})
			case n <= 8:
				in.product = append(in.product, agSetInput{class: class, n: n, fs: agProductAlt(n)})
			default:
				// alt/64: every value of the class has |num| near 2^40 and den
				// at most 2^16, so |value| ≥ 2^23 and a product of 64 values
				// exceeds MaxComponent; no successful operation exists.
				in.skippedProduct = append(in.skippedProduct, fmt.Sprintf("%s/%d", class, n))
			}
			if n == 2 {
				// A set of two has no intermediate value: its only partial
				// result is the result, so the exact continuation is unreachable.
				in.skippedSumOF = append(in.skippedSumOF, fmt.Sprintf("%s/%d", class, n))
				in.skippedProdOF = append(in.skippedProdOF, fmt.Sprintf("%s/%d", class, n))
				continue
			}
			if fs, ok := agOverflowSet(class, n, false); ok {
				in.sumOverflow = append(in.sumOverflow, agSetInput{class: class, n: n, overflow: true, fs: fs})
			} else {
				in.skippedSumOF = append(in.skippedSumOF, fmt.Sprintf("%s/%d", class, n))
			}
			if fs, ok := agOverflowSet(class, n, true); ok {
				in.prodOverflow = append(in.prodOverflow, agSetInput{class: class, n: n, overflow: true, fs: fs})
			} else {
				in.skippedProdOF = append(in.skippedProdOF, fmt.Sprintf("%s/%d", class, n))
			}
		}
	}

	// GCD, LCM: the common factor g of each operand is about a quarter of the
	// component width (task 26.6, "Requirements for benchmark inputs").
	in.pairs = []agPairInput{
		agPair("w31", bp(7, 0), bp(7, 1), bp(7, 2), bp(13, 0), bp(13, 1), bp(13, 2), bp(13, 3)),
		agPair("alt", bp(3, 0), bp(11, 0), bp(3, 1), bp(24, 0), bp(5, 0), bp(24, 1), bp(5, 1)),
		agPair("w63", bp(15, 0), bp(15, 1), bp(15, 2), bp(29, 0), bp(29, 1), bp(29, 2), bp(29, 3)),
		agPair("w128", bp(31, 0), bp(31, 1), bp(31, 2), bp(61, 0), bp(61, 1), bp(61, 2), bp(61, 3)),
		// w255: LCM = h·a·c/k is in range only with a, c well below the
		// component width, so h and k take the rest of it.
		agPair("w255", bp(60, 0), bp(128, 0), bp(128, 1), bp(59, 0), bp(59, 1), bp(59, 2), bp(59, 3)),
	}
	return in
}

// agBigRatToInputs returns the unreduced pairs (g·a)/(g·b) of the Reduce
// inputs of task 26.6 in the classes w31, w63, w128 and w255.
func agBigRatToInputs() []benchInput {
	var out []benchInput
	for _, in := range benchInputs(benchReduce) {
		if in.class != "alt" {
			out = append(out, in)
		}
	}
	return out
}
