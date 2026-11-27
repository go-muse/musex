package fraction_test

import (
	"math/big"
	"strings"
	"sync"

	"github.com/go-muse/musex/common/fraction"
)

// Benchmark inputs by width class (architecture of task 26.6, "Width classes"
// and "Requirements for benchmark inputs"). The inputs are deterministic
// products of probable primes, so every common factor is known by
// construction; TestBenchmarkInputs checks the required properties.
//
// Classes: w31 (components up to 31 bits), alt (numerator near 2^40,
// denominator up to 2^16), w63 (33–63 bits, targets near 2^62), w128 (65–128
// bits), w255 (193–255 bits). A "-short" suffix marks the short-quotient
// inputs of the division operations.

// The benchmark files (bench_test.go and this file) are self-contained: the
// baseline benchmark binary of task 26.6 is built from them and the product
// code only, so they use no helpers of the other test files.

func benchPow2(e uint) *big.Int {
	return new(big.Int).Lsh(big.NewInt(1), e)
}

// benchMul returns the product of the factors as a new value.
func benchMul(factors ...*big.Int) *big.Int {
	p := big.NewInt(1)
	for _, f := range factors {
		p.Mul(p, f)
	}
	return p
}

func benchNeg(x *big.Int) *big.Int {
	return new(big.Int).Neg(x)
}

// benchFrac builds a fraction with the original pair (num, den); it panics on
// invalid table data.
func benchFrac(num, den *big.Int) fraction.Fraction {
	return fraction.MustFromBigInt(num, den)
}

// benchRat returns the exact value of a valid fraction as a big.Rat.
func benchRat(f fraction.Fraction) *big.Rat {
	num, den, ok := f.BigPair()
	if !ok {
		panic("benchRat: invalid fraction")
	}
	return new(big.Rat).SetFrac(num, den)
}

// benchInput is one operand set of a benchmark: x is the receiver, y the
// argument of binary operations, den the int denominator where one is taken.
type benchInput struct {
	class string
	x, y  fraction.Fraction
	den   int
}

// baseClass returns the class without the "-short" suffix.
func (in benchInput) baseClass() string {
	return strings.TrimSuffix(in.class, "-short")
}

func (in benchInput) short() bool {
	return strings.HasSuffix(in.class, "-short")
}

// benchFamily groups operations that share inputs.
type benchFamily uint8

const (
	// Add, Sub, Mul, Div, Compare, Equal and the big.Rat references Add, Sub, Mul, Quo, Cmp.
	benchBinary benchFamily = iota
	// AddWith(DenomUnreducedLCM).
	benchLCM
	// AddWith(DenomLargerUnit), AddInDenominator (den = g).
	benchUnit
	// QuoRem, DivModFloor, IsMultipleOf.
	benchDivMod
	// Reduce, IsReduced and the big.Rat reference SetFrac.
	benchReduce
	// WithDenominator.
	benchWithDen
	// IsInteger, Split, SplitFloor, RoundToInt.
	benchSplit
	// RoundToDenominator.
	benchRoundDen
)

// bp returns the smallest probable prime not below 2^bits + idx·2^(bits/2+1) + 1.
// Different idx values give different primes of the same size.
func bp(bits uint, idx int) *big.Int {
	p := benchPow2(bits)
	p.Add(p, new(big.Int).Lsh(big.NewInt(int64(idx)), bits/2+1))
	p.Add(p, big.NewInt(1))
	for !p.ProbablyPrime(0) {
		p.Add(p, big.NewInt(2))
	}
	return p
}

var benchInputsOnce = sync.OnceValue(buildBenchInputs)

// benchInputs returns the inputs of a family; the slice must not be modified.
func benchInputs(family benchFamily) []benchInput {
	return benchInputsOnce()[family]
}

func benchPair(class string, n1, d1, n2, d2 *big.Int) benchInput {
	return benchInput{class: class, x: benchFrac(n1, d1), y: benchFrac(n2, d2)}
}

func benchOne(class string, n, d *big.Int, den int) benchInput {
	return benchInput{class: class, x: benchFrac(n, d), den: den}
}

// benchCross returns x = (c·p1, g·x1), y = (p2, g·c·y1): the denominators share g,
// the numerator of x shares c with the denominator of y, so Add, Sub, Div reduce
// by at least g and Mul by c. Both operands are positive, so Compare and Equal
// cannot decide by the signs alone.
func benchCross(class string, g, c, p1, x1, p2, y1 *big.Int) benchInput {
	return benchPair(class, benchMul(c, p1), benchMul(g, x1), p2, benchMul(g, c, y1))
}

// benchCancel returns x = (a·u, g·u), y = (s·b·v, g·v) with s = ±1: the operands
// reduce to a/g and s·b/g, so results of every arithmetic operation stay small
// while the intermediate values have the full width of the class.
func benchCancel(class string, g, a, u, b, v *big.Int, negY bool) benchInput {
	nb := benchMul(b, v)
	if negY {
		nb.Neg(nb)
	}
	return benchPair(class, benchMul(a, u), benchMul(g, u), nb, benchMul(g, v))
}

func buildBenchInputs() map[benchFamily][]benchInput {
	in := make(map[benchFamily][]benchInput)
	three := big.NewInt(3)

	// Binary arithmetic and comparison.
	in[benchBinary] = []benchInput{
		benchCross("w31", bp(8, 0), bp(8, 1), bp(22, 0), bp(22, 1), bp(30, 0), bp(14, 0)),
		benchCross("alt", bp(4, 0), bp(4, 1), bp(36, 0), bp(11, 0), bp(40, 0), bp(7, 0)),
		benchCross("w63", bp(16, 0), bp(16, 1), bp(46, 0), bp(46, 1), bp(62, 0), bp(30, 0)),
		benchCross("w128", bp(32, 0), bp(32, 1), bp(94, 0), bp(94, 1), bp(126, 0), bp(62, 0)),
		// w255: with denominators near 2^250 a result is in range only when the
		// reduction removes at least 2^245, so both operands reduce to a/g, b/g.
		benchCancel("w255", bp(60, 0), bp(60, 1), bp(190, 0), bp(60, 2), bp(190, 1), false),
	}

	// AddWith(DenomUnreducedLCM): the common factor of the denominators is g.
	in[benchLCM] = []benchInput{
		in[benchBinary][0], in[benchBinary][1], in[benchBinary][2], in[benchBinary][3],
		// w255: LCM(g·x1, g·y1) with both denominators above 2^192 and g near 2^60
		// exceeds MaxComponent, so y = (≈2^60, 3·g) is outside the class
		// (decision of the architect, task 26.6).
		benchPair("w255", bp(248, 0), benchMul(bp(60, 0), bp(188, 0)), bp(60, 1), benchMul(three, bp(60, 0))),
	}

	// AddWith(DenomLargerUnit) and AddInDenominator(den = g): x = (a·u, g·u),
	// y = (−b·v, g·v); the reduced sum (a − b)/g fits both target denominators.
	unit := func(class string, gBits, abBits, uvBits uint) benchInput {
		g := bp(gBits, 0)
		c := benchCancel(class, g, bp(abBits, 1), bp(uvBits, 0), bp(abBits, 2), bp(uvBits, 1), true)
		c.den = int(g.Int64())
		return c
	}
	in[benchUnit] = []benchInput{
		unit("w31", 8, 8, 22),
		unit("alt", 4, 29, 11),
		unit("w63", 16, 16, 46),
		unit("w128", 32, 32, 94),
		unit("w255", 60, 60, 190),
	}

	// QuoRem, DivModFloor, IsMultipleOf: f = (−n1, g·x1), step = (n2, g·y1); the
	// quotient is about n1·y1 / (x1·n2).
	divmod := func(class string, g, n1, x1, n2, y1 *big.Int) benchInput {
		return benchPair(class, benchNeg(n1), benchMul(g, x1), n2, benchMul(g, y1))
	}
	binW31, binAlt := in[benchBinary][0], in[benchBinary][1]
	in[benchDivMod] = []benchInput{
		// Controls: the binary operands with a negative dividend.
		{class: "w31", x: binW31.x.MustNeg(), y: binW31.y},
		{class: "alt", x: binAlt.x.MustNeg(), y: binAlt.y},
		divmod("w63", bp(16, 0), bp(62, 0), bp(24, 0), bp(35, 0), bp(46, 0)),          // quotient ≈ 2^49
		divmod("w63-short", bp(16, 0), bp(62, 0), bp(44, 0), bp(60, 0), bp(46, 0)),    // quotient ≈ 2^4
		divmod("w128", bp(32, 0), bp(126, 0), bp(69, 0), bp(101, 0), bp(94, 0)),       // quotient ≈ 2^50
		divmod("w128-short", bp(32, 0), bp(126, 0), bp(92, 0), bp(124, 0), bp(94, 0)), // quotient ≈ 2^4
		// w255: f = (−a·u, g·u), step = (b·v, g·v); dividend ≈ 2^500, divisor ≈ 2^450,
		// quotient −a/b ≈ −2^50; the remainder (a mod b)/g-like value stays in range.
		benchDivModCancel("w255", bp(60, 0), bp(110, 0), bp(140, 0), bp(60, 1), bp(190, 0)),
		benchDivModCancel("w255-short", bp(60, 0), bp(64, 0), bp(186, 0), bp(60, 1), bp(190, 0)),
	}

	// Reduce and IsReduced: unreduced pairs (g·a, g·b) with positive
	// denominators, so IsReduced has to compute the NOD.
	in[benchReduce] = []benchInput{
		benchOne("w31", benchMul(bp(8, 0), bp(22, 0)), benchMul(bp(8, 0), bp(22, 1)), 0),
		benchOne("alt", benchMul(bp(4, 0), bp(36, 0)), benchMul(bp(4, 0), bp(11, 0)), 0),
		benchOne("w63", benchMul(bp(16, 0), bp(46, 0)), benchMul(bp(16, 0), bp(46, 1)), 0),
		benchOne("w128", benchNeg(benchMul(bp(32, 0), bp(94, 0))), benchMul(bp(32, 0), bp(94, 1)), 0),
		benchOne("w255", benchNeg(benchMul(bp(60, 0), bp(190, 0))), benchMul(bp(60, 0), bp(190, 1)), 0),
	}

	// WithDenominator: (g·a, g·b) with den = m·b, a multiple of the reduced
	// denominator b not above 2^62.
	withDen := func(class string, g, a, b, m *big.Int) benchInput {
		return benchOne(class, benchMul(g, a), benchMul(g, b), int(benchMul(b, m).Int64()))
	}
	in[benchWithDen] = []benchInput{
		withDen("w31", bp(8, 0), bp(22, 0), bp(22, 1), bp(8, 1)),
		withDen("alt", bp(4, 0), bp(36, 0), bp(11, 0), bp(4, 1)),
		withDen("w63", bp(16, 0), bp(46, 0), bp(44, 0), bp(16, 1)),
		withDen("w128", bp(32, 0), bp(94, 0), bp(60, 0), three),
		// w255: den is an int, so the reduced denominator b is at most 2^62 and,
		// with g near 2^60, the denominator g·b ≈ 2^120 is outside the class
		// (decision of the architect, task 26.6).
		withDen("w255", bp(60, 0), bp(190, 0), bp(60, 1), three),
	}

	// IsInteger, Split, SplitFloor, RoundToInt: |n| ÷ D.
	in[benchSplit] = []benchInput{
		benchOne("w31", benchNeg(bp(30, 0)), bp(20, 0), 0),
		benchOne("alt", bp(40, 0), bp(15, 0), 0),
		benchOne("w63", benchNeg(bp(62, 0)), bp(40, 0), 0),          // control, quotient ≈ 2^22
		benchOne("w63-short", benchNeg(bp(62, 0)), bp(56, 0), 0),    // quotient ≈ 2^6
		benchOne("w128", benchNeg(bp(126, 0)), bp(76, 0), 0),        // quotient ≈ 2^50
		benchOne("w128-short", benchNeg(bp(126, 0)), bp(120, 0), 0), // quotient ≈ 2^6
		benchOne("w255", benchNeg(bp(250, 0)), bp(200, 0), 0),       // quotient ≈ 2^50
		benchOne("w255-short", benchNeg(bp(250, 0)), bp(245, 0), 0), // quotient ≈ 2^5
	}

	// RoundToDenominator: |n|·den ÷ D.
	den := func(p *big.Int) int { return int(p.Int64()) }
	in[benchRoundDen] = []benchInput{
		benchOne("w31", bp(30, 0), bp(29, 0), den(bp(30, 1))),
		benchOne("alt", bp(40, 0), bp(15, 0), den(bp(16, 0))),
		benchOne("w63", benchNeg(bp(50, 0)), bp(60, 0), den(bp(62, 0))),         // quotient ≈ 2^52
		benchOne("w63-short", benchNeg(bp(39, 0)), bp(62, 0), den(bp(30, 0))),   // quotient ≈ 2^7
		benchOne("w128", benchNeg(bp(124, 0)), bp(126, 0), den(bp(62, 0))),      // quotient ≈ 2^60
		benchOne("w128-short", benchNeg(bp(70, 0)), bp(127, 0), den(bp(62, 0))), // quotient ≈ 2^5
		benchOne("w255", benchNeg(bp(250, 0)), bp(204, 0), den(bp(4, 0))),       // quotient ≈ 2^50
		benchOne("w255-short", benchNeg(bp(250, 0)), bp(250, 1), den(bp(4, 0))), // quotient ≈ 2^4
	}
	return in
}

// benchDivModCancel returns f = (−a·u, g·u), step = (b·v, g·v).
func benchDivModCancel(class string, g, a, u, b, v *big.Int) benchInput {
	return benchPair(class, benchNeg(benchMul(a, u)), benchMul(g, u), benchMul(b, v), benchMul(g, v))
}
