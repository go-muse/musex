package fraction_test

import (
	"fmt"
	"math"
	"math/big"
	"slices"
	"sync"
	"testing"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of task 26.7 (architecture, "Library API contract" and "Tests"): GCD,
// LCM, CommonDenominator, CommonDenominatorBig, Sum, Product, their Must forms
// and BigRatTo. Every case is checked against the contract oracle of
// aggregate_oracle_test.go; the explicit expectations of the tables come from
// the examples of the contract and check the oracle itself.

// agN is a short form of MustNew for table data.
func agN(num, den int) fraction.Fraction {
	return fraction.MustNew(num, den)
}

// agQuoExact returns a / b and fails the test unless b divides a.
func agQuoExact(t *testing.T, a, b *big.Int) *big.Int {
	t.Helper()
	q, r := new(big.Int).QuoRem(a, b, new(big.Int))
	require.Zerof(t, r.Sign(), "%s must divide %s", b, a)
	return q
}

// agBoundaryValues returns the boundary values of the architecture: maxBy7 with
// LCM(maxBy7, 7) = 2^255 − 1 = MaxComponent and above3 with LCM(above3, 9) =
// 2^255 + 1, the smallest reachable value above the range.
func agBoundaryValues(t *testing.T) (maxBy7, above3 *big.Int) {
	t.Helper()
	maxC := fraction.MaxComponentBig()
	maxBy7 = agQuoExact(t, maxC, big.NewInt(7))
	require.Zero(t, agLCMInt(maxBy7, big.NewInt(7)).Cmp(maxC), "LCM(maxBy7, 7) must be MaxComponent")
	plus1 := wOff(wPow2(255), 1)
	above3 = agQuoExact(t, plus1, big.NewInt(3))
	require.Zero(t, agLCMInt(above3, big.NewInt(9)).Cmp(plus1), "LCM(above3, 9) must be 2^255 + 1")
	return maxBy7, above3
}

func TestFraction_GCDLCM(t *testing.T) {
	t.Parallel()

	maxBy7, above3 := agBoundaryValues(t)
	maxC := fraction.MaxComponentBig()
	one := big.NewInt(1)
	a, b := bp(200, 0), bp(200, 1)
	pow63 := wPow2(63)
	outOfRange := wErr(fraction.ErrOutOfRange)
	invalid := wErr(fraction.ErrInvalid)
	var zero fraction.Fraction

	tests := []struct {
		name     string
		x, y     fraction.Fraction
		gcd, lcm wideWant
	}{
		{"quarter and sixth", agN(1, 4), agN(1, 6), agWantInts(1, 12), agWantInts(1, 2)},
		{"opposite signs", agN(3, 4), agN(-1, 2), agWantInts(1, 4), agWantInts(3, 2)},
		{"zero argument", agN(6, 12), agN(0, 5), agWantInts(1, 2), agWantInts(0, 1)},
		{"both zero", agN(0, 3), agN(0, 7), agWantInts(0, 1), agWantInts(0, 1)},
		{"zero integer", agN(2, 3), agN(0, 1), agWantInts(2, 3), agWantInts(0, 1)},
		{"negative denominators", agN(1, -4), agN(-1, -6), agWantInts(1, 12), agWantInts(1, 2)},
		{"equal values of different pairs", agN(10, 4), agN(-15, -6), agWantInts(5, 2), agWantInts(5, 2)},
		{"integers", agN(12, 1), agN(-18, 1), agWantInts(6, 1), agWantInts(36, 1)},
		{"coprime fractions", agN(2, 3), agN(3, 4), agWantInts(1, 12), agWantInts(6, 1)},
		{"MinInt numerator", agN(math.MinInt, 3), agN(4, 1), agWantInts(4, 3), agWantPair(pow63, one)},
		{"MinInt denominator", agN(1, math.MinInt), agN(1, 6), agWantPair(one, wMul(big.NewInt(3), pow63)), agWantInts(1, 2)},
		{"LCM numerator MaxComponent", wFrac(maxBy7, one), agN(7, 1), agWantInts(1, 1), agWantPair(maxC, one)},
		{"LCM numerator 2^255+1", wFrac(above3, one), agN(9, 1), agWantInts(3, 1), outOfRange},
		{"GCD denominator MaxComponent", wFrac(one, maxBy7), agN(1, 7), agWantPair(one, maxC), agWantInts(1, 1)},
		{"GCD denominator 2^255+1", wFrac(one, above3), agN(-1, 9), outOfRange, agWantInts(1, 3)},
		{"MaxComponent integers", wFrac(maxC, one), wFrac(wNeg(maxC), one), agWantPair(maxC, one), agWantPair(maxC, one)},
		{"MaxComponent reciprocals", wFrac(one, maxC), wFrac(one, wNeg(maxC)), agWantPair(one, maxC), agWantPair(one, maxC)},
		{"MaxComponent pair", wFrac(maxC, maxC), agN(1, 2), agWantInts(1, 2), agWantInts(1, 1)},
		{"wide coprime denominators", wFrac(one, a), wFrac(one, b), outOfRange, agWantInts(1, 1)},
		{"wide coprime numerators", wFrac(a, one), wFrac(b, one), agWantInts(1, 1), outOfRange},
		{"wide common factor", wFrac(wMul(a, big.NewInt(6)), wMul(b, big.NewInt(5))), wFrac(wMul(a, big.NewInt(4)), wMul(b, big.NewInt(3))),
			agWantPair(wMul(a, big.NewInt(2)), wMul(b, big.NewInt(15))), agWantPair(wMul(a, big.NewInt(12)), b)},
		{"invalid receiver", zero, agN(1, 2), invalid, invalid},
		{"invalid argument", agN(1, 2), zero, invalid, invalid},
		{"both invalid", zero, zero, invalid, invalid},
		{"invalid receiver before out of range", zero, wFrac(one, above3), invalid, invalid},
		{"invalid argument before out of range", wFrac(above3, one), zero, invalid, invalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			agSameWant(t, "GCD", agWantGCD(tt.x, tt.y), tt.gcd)
			agSameWant(t, "LCM", agWantLCM(tt.x, tt.y), tt.lcm)
			agCheckGCDLCM(t, tt.name, tt.x, tt.y)
		})
	}
}

// agLCMUpTo returns LCM(1, …, n) as the product of the largest prime powers
// not above n, a computation independent of the oracle.
func agLCMUpTo(n int64) *big.Int {
	l := big.NewInt(1)
	for p := int64(2); p <= n; p++ {
		if !big.NewInt(p).ProbablyPrime(0) {
			continue
		}
		pk := p
		for pk*p <= n {
			pk *= p
		}
		l.Mul(l, big.NewInt(pk))
	}
	return l
}

func TestCommonDenominator(t *testing.T) {
	t.Parallel()

	var zero fraction.Fraction
	maxInt := big.NewInt(math.MaxInt)
	pow63 := wPow2(63)
	one := big.NewInt(1)
	const maxIntFactor = 3577 // 7² · 73; MaxInt = 3577 · 2578437596736553
	require.Zero(t, math.MaxInt%maxIntFactor)
	require.Equal(t, int64(1), wGCD(big.NewInt(maxIntFactor), big.NewInt(math.MaxInt/maxIntFactor)).Int64())

	cycle := []int{2, 3, 4, 6, 12}
	small64 := make([]fraction.Fraction, 64)
	upTo64 := make([]fraction.Fraction, 64)
	for i := range 64 {
		k := i%3 + 1
		small64[i] = agN(k, k*cycle[i%len(cycle)])
		upTo64[i] = agN(1, i+1)
	}
	wide := make([]fraction.Fraction, 4)
	wideD := big.NewInt(1)
	for i := range wide {
		p := bp(250, i)
		wide[i] = wFrac(big.NewInt(int64(2*i+1)), p)
		wideD.Mul(wideD, p)
	}
	require.Greater(t, wideD.BitLen(), 512)

	tests := []struct {
		name string
		fs   []fraction.Fraction
		want *big.Int // nil when the error is ErrInvalid
	}{
		{"example", []fraction.Fraction{agN(1, 4), agN(1, 6), agN(5, 12)}, big.NewInt(12)},
		{"reduced denominators", []fraction.Fraction{agN(6, 12), agN(0, 7)}, big.NewInt(2)},
		{"empty nil set", nil, one},
		{"empty set", []fraction.Fraction{}, one},
		{"single unreduced", []fraction.Fraction{agN(3, 9)}, big.NewInt(3)},
		{"single zero", []fraction.Fraction{agN(0, -5)}, one},
		{"single integer", []fraction.Fraction{agN(-10, 2)}, one},
		{"negative denominators", []fraction.Fraction{agN(1, -4), agN(-1, -6)}, big.NewInt(12)},
		{"MaxInt single", []fraction.Fraction{agN(1, math.MaxInt)}, maxInt},
		{"MaxInt by LCM", []fraction.Fraction{agN(1, maxIntFactor), agN(-5, math.MaxInt/maxIntFactor)}, maxInt},
		{"MaxInt+1 single", []fraction.Fraction{wFrac(one, pow63)}, pow63},
		{"MaxInt+1 with powers of two", []fraction.Fraction{agN(1, 1<<62), wFrac(big.NewInt(3), pow63), agN(1, 2)}, pow63},
		{"MinInt denominator", []fraction.Fraction{agN(1, math.MinInt)}, pow63},
		{"2·MaxInt by LCM", []fraction.Fraction{agN(1, math.MaxInt), agN(1, 2)}, wMul(maxInt, big.NewInt(2))},
		{"64 elements with a small LCM", small64, big.NewInt(12)},
		{"64 elements up to 64", upTo64, agLCMUpTo(64)},
		{"beyond 2^512", wide, wideD},
		{"MaxComponent denominators", []fraction.Fraction{wFrac(one, fraction.MaxComponentBig()), agN(1, 2)},
			wMul(fraction.MaxComponentBig(), big.NewInt(2))},
		{"invalid only", []fraction.Fraction{zero}, nil},
		{"invalid first before not representable", []fraction.Fraction{zero, wFrac(one, pow63)}, nil},
		{"invalid middle before not representable", []fraction.Fraction{wFrac(one, pow63), zero, agN(1, 3)}, nil},
		{"invalid last before not representable", []fraction.Fraction{agN(1, math.MaxInt), agN(1, 2), zero}, nil},
		{"invalid among wide", append(slices.Clone(wide), zero), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := agWantCommonDenominatorBig(tt.fs)
			if tt.want == nil {
				assert.ErrorIs(t, err, fraction.ErrInvalid, "oracle")
			} else if assert.NoError(t, err, "oracle") {
				assert.Zero(t, got.Cmp(tt.want), "oracle %s, table %s", got, tt.want)
			}
			agCheckCommonDenominator(t, tt.name, tt.fs)
		})
	}
}

// agTelescoping returns 1/(1·2), 1/(2·3), …, 1/(n·(n+1)).
func agTelescoping(n int) []fraction.Fraction {
	fs := make([]fraction.Fraction, n)
	for i := range n {
		fs[i] = agN(1, (i+1)*(i+2))
	}
	return fs
}

// agRepeat returns n copies of f.
func agRepeat(f fraction.Fraction, n int) []fraction.Fraction {
	return slices.Repeat([]fraction.Fraction{f}, n)
}

func TestSum(t *testing.T) {
	t.Parallel()

	var zero fraction.Fraction
	maxC := fraction.MaxComponentBig()
	one := big.NewInt(1)
	a, b := bp(200, 0), bp(200, 1)
	invA, invB := wFrac(one, a), wFrac(one, b)
	restA := wFrac(wOff(a, -1), a)
	outOfRange := wErr(fraction.ErrOutOfRange)
	invalid := wErr(fraction.ErrInvalid)

	tests := []struct {
		name string
		fs   []fraction.Fraction
		want wideWant
	}{
		{"example thirds", []fraction.Fraction{agN(1, 3), agN(1, 6)}, agWantInts(1, 2)},
		{"example quarters", []fraction.Fraction{agN(1, 4), agN(1, 4), agN(2, 4)}, agWantInts(1, 1)},
		{"empty nil set", nil, agWantInts(0, 1)},
		{"empty set", []fraction.Fraction{}, agWantInts(0, 1)},
		{"single unreduced", []fraction.Fraction{agN(6, -12)}, agWantInts(-1, 2)},
		{"single zero", []fraction.Fraction{agN(0, -5)}, agWantInts(0, 1)},
		{"cancel to zero", []fraction.Fraction{agN(1, 2), agN(-2, 4)}, agWantInts(0, 1)},
		{"negative denominators", []fraction.Fraction{agN(1, -3), agN(1, 6)}, agWantInts(-1, 6)},
		{"intermediate out of range", []fraction.Fraction{invA, invB, restA}, agWantPair(wOff(b, 1), b)},
		{"result out of range", []fraction.Fraction{invA, invB}, outOfRange},
		{"result MaxComponent", []fraction.Fraction{wFrac(wOff(maxC, -1), one), agN(1, 1)}, agWantPair(maxC, one)},
		{"result above MaxComponent", []fraction.Fraction{wFrac(maxC, one), agN(1, 1)}, outOfRange},
		{"MaxComponent after out of range", []fraction.Fraction{wFrac(maxC, one), agN(1, 1), agN(-1, 1)}, agWantPair(maxC, one)},
		{"negative MaxComponent after out of range", []fraction.Fraction{wFrac(wNeg(maxC), one), agN(-1, 1), agN(2, 2)}, agWantPair(wNeg(maxC), one)},
		{"64 equal elements", agRepeat(agN(1, 64), 64), agWantInts(1, 1)},
		{"64 telescoping elements", agTelescoping(64), agWantInts(64, 65)},
		{"invalid only", []fraction.Fraction{zero}, invalid},
		{"invalid first before out of range", []fraction.Fraction{zero, invA, invB}, invalid},
		{"invalid middle before out of range", []fraction.Fraction{invA, zero, invB}, invalid},
		{"invalid last before out of range", []fraction.Fraction{invA, invB, zero}, invalid},
		{"invalid after out of range prefix", []fraction.Fraction{invA, invB, restA, zero}, invalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			agSameWant(t, "Sum", agWantSum(tt.fs), tt.want)
			agCheckSumProduct(t, tt.name, tt.fs)
		})
	}
}

func TestProduct(t *testing.T) {
	t.Parallel()

	var zero fraction.Fraction
	maxC := fraction.MaxComponentBig()
	one := big.NewInt(1)
	p200, p250, p254 := wPow2(200), wPow2(250), wPow2(254)
	big200 := wFrac(p200, one)
	big250 := wFrac(p250, one)
	outOfRange := wErr(fraction.ErrOutOfRange)
	invalid := wErr(fraction.ErrInvalid)
	ratios := make([]fraction.Fraction, 64)
	for i := range ratios {
		ratios[i] = agN(i+1, i+2)
	}

	tests := []struct {
		name string
		fs   []fraction.Fraction
		want wideWant
	}{
		{"example", []fraction.Fraction{agN(1, 8), agN(2, 3)}, agWantInts(1, 12)},
		{"example with intermediate out of range", []fraction.Fraction{big200, big200, wFrac(one, p200)}, agWantPair(p200, one)},
		{"empty nil set", nil, agWantInts(1, 1)},
		{"empty set", []fraction.Fraction{}, agWantInts(1, 1)},
		{"single unreduced", []fraction.Fraction{agN(6, -12)}, agWantInts(-1, 2)},
		{"two negatives", []fraction.Fraction{agN(-1, 2), agN(1, -3)}, agWantInts(1, 6)},
		{"negative result", []fraction.Fraction{agN(1, -2), agN(3, 1)}, agWantInts(-3, 2)},
		{"zero alone", []fraction.Fraction{agN(0, 5)}, agWantInts(0, 1)},
		{"zero after large", []fraction.Fraction{big250, big250, agN(0, 1)}, agWantInts(0, 1)},
		{"zero first", []fraction.Fraction{agN(0, -3), big250, big250}, agWantInts(0, 1)},
		{"zero between large", []fraction.Fraction{big250, agN(0, 7), wFrac(one, p250), big250}, agWantInts(0, 1)},
		{"result MaxComponent", []fraction.Fraction{wFrac(maxC, one), agN(-1, -1)}, agWantPair(maxC, one)},
		{"MaxComponent after out of range", []fraction.Fraction{wFrac(maxC, one), wFrac(maxC, one), wFrac(one, maxC)}, agWantPair(maxC, one)},
		{"result out of range", []fraction.Fraction{big200, big200}, outOfRange},
		{"result 2^255", []fraction.Fraction{wFrac(p254, one), agN(2, 1)}, outOfRange},
		{"2^254 after 2^255", []fraction.Fraction{wFrac(p254, one), agN(2, 1), agN(1, 2)}, agWantPair(p254, one)},
		{"64 doublings", agRepeat(agN(2, 1), 64), agWantPair(wPow2(64), one)},
		{"64 telescoping ratios", ratios, agWantInts(1, 65)},
		{"invalid only", []fraction.Fraction{zero}, invalid},
		{"invalid after zero", []fraction.Fraction{agN(0, 1), zero}, invalid},
		{"invalid before zero", []fraction.Fraction{zero, agN(0, 1)}, invalid},
		{"invalid after out of range prefix", []fraction.Fraction{big200, big200, zero}, invalid},
		{"invalid middle before out of range", []fraction.Fraction{big200, zero, big200}, invalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			agSameWant(t, "Product", agWantProduct(tt.fs), tt.want)
			agCheckSumProduct(t, tt.name, tt.fs)
		})
	}
}

// agPermutations returns every permutation of fs as new slices.
func agPermutations(fs []fraction.Fraction) [][]fraction.Fraction {
	if len(fs) <= 1 {
		return [][]fraction.Fraction{slices.Clone(fs)}
	}
	var out [][]fraction.Fraction
	for i := range fs {
		rest := slices.Concat(fs[:i], fs[i+1:])
		for _, p := range agPermutations(rest) {
			out = append(out, append([]fraction.Fraction{fs[i]}, p...))
		}
	}
	return out
}

// TestSumProduct_OrderIndependence checks that the value and the error of Sum
// and Product depend only on the exact result: every order of a set gives the
// same result, also when some prefix is out of range.
func TestSumProduct_OrderIndependence(t *testing.T) {
	t.Parallel()

	maxC := fraction.MaxComponentBig()
	one := big.NewInt(1)
	a, b := bp(200, 0), bp(200, 1)
	invA, invB := wFrac(one, a), wFrac(one, b)
	p200 := wPow2(200)
	big250 := wFrac(wPow2(250), one)

	tests := []struct {
		name    string
		product bool
		fs      []fraction.Fraction
		want    wideWant
	}{
		{"sum 1/A + 1/B + (A-1)/A", false, []fraction.Fraction{invA, invB, wFrac(wOff(a, -1), a)}, agWantPair(wOff(b, 1), b)},
		{"sum cancelling wide reciprocals", false, []fraction.Fraction{invA, invB, wFrac(wNeg(one), a), wFrac(one, wNeg(b))}, agWantInts(0, 1)},
		{"sum MaxComponent", false, []fraction.Fraction{wFrac(maxC, one), agN(1, 1), agN(-1, 1)}, agWantPair(maxC, one)},
		{"sum out of range", false, []fraction.Fraction{invA, invB, agN(1, 2)}, wErr(fraction.ErrOutOfRange)},
		{"product 2^200", true, []fraction.Fraction{wFrac(p200, one), wFrac(p200, one), wFrac(one, p200)}, agWantPair(p200, one)},
		{"product MaxComponent", true, []fraction.Fraction{wFrac(maxC, one), wFrac(maxC, one), wFrac(one, maxC)}, agWantPair(maxC, one)},
		{"product with zero", true, []fraction.Fraction{big250, big250, agN(0, 1), big250}, agWantInts(0, 1)},
		{"product out of range", true, []fraction.Fraction{wFrac(p200, one), wFrac(p200, one), agN(1, 2)}, wErr(fraction.ErrOutOfRange)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for i, perm := range agPermutations(tt.fs) {
				label := fmt.Sprintf("%s, order %d", tt.name, i)
				var got fraction.Fraction
				var err error
				if tt.product {
					agSameWant(t, label, agWantProduct(perm), tt.want)
					got, err = fraction.Product(perm...)
				} else {
					agSameWant(t, label, agWantSum(perm), tt.want)
					got, err = fraction.Sum(perm...)
				}
				wCheckFrac(t, label, got, err, tt.want)
			}
		})
	}
}

// agBigRatToValues returns fractions of every width class and form: zero,
// integers, negative values, negative denominators of the original pair and
// unreduced pairs.
func agBigRatToValues() []struct {
	name string
	f    fraction.Fraction
} {
	maxC := fraction.MaxComponentBig()
	one := big.NewInt(1)
	g := bp(60, 0)
	return []struct {
		name string
		f    fraction.Fraction
	}{
		{"zero", agN(0, -7)},
		{"one", fraction.One()},
		{"minus three", agN(-6, 2)},
		{"w31 unreduced", agN(6*1_000_003, 6*999_983)},
		{"w31 negative denominator", agN(5, -7)},
		{"w63", wFrac(wMul(bp(16, 0), bp(46, 0)), wMul(bp(16, 0), bp(46, 1)))},
		{"w63 MinInt", agN(math.MinInt, 3)},
		{"w128 negative", wFrac(wNeg(wMul(bp(32, 0), bp(94, 0))), wMul(bp(32, 0), bp(94, 1)))},
		{"w255", wFrac(wMul(g, bp(190, 0)), wNeg(wMul(g, bp(190, 1))))},
		{"w255 integer", wFrac(wMul(g, bp(190, 0)), g)},
		{"MaxComponent", wFrac(maxC, one)},
		{"MaxComponent reciprocal", wFrac(wNeg(one), maxC)},
		{"MaxComponent pair", wFrac(maxC, wNeg(maxC))},
	}
}

func TestFraction_BigRatTo(t *testing.T) {
	t.Parallel()

	presets := []struct {
		name string
		z    func() *big.Rat
	}{
		{"nil", func() *big.Rat { return nil }},
		{"zero value", func() *big.Rat { return new(big.Rat) }},
		{"5/7", func() *big.Rat { return big.NewRat(5, 7) }},
		{"-3", func() *big.Rat { return big.NewRat(-3, 1) }},
		{"0", func() *big.Rat { return new(big.Rat).SetInt64(0) }},
		{"wide", func() *big.Rat { return new(big.Rat).SetFrac(wPow2(300), wOff(wPow2(299), 1)) }},
	}
	for _, v := range agBigRatToValues() {
		for _, p := range presets {
			t.Run(v.name+"/"+p.name, func(t *testing.T) {
				t.Parallel()
				agCheckBigRatTo(t, "BigRatTo", v.f, p.z())
			})
		}
	}

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		var zero fraction.Fraction
		for _, p := range presets {
			z := p.z()
			agCheckBigRatTo(t, "invalid into "+p.name, zero, z)
			got, ok := zero.BigRatTo(z)
			assert.False(t, ok, p.name)
			assert.True(t, got == z, "%s: the result must be z", p.name)
		}
		got, ok := zero.BigRatTo(nil)
		assert.Nil(t, got)
		assert.False(t, ok)
		z := big.NewRat(5, 7)
		_, _ = zero.BigRatTo(z)
		assert.Equal(t, "5/7", z.String(), "z must not change")
		assert.Equal(t, "5", z.Num().String(), "numerator must not change")
		assert.Equal(t, "7", z.Denom().String(), "denominator must not change")
	})
}

// TestFraction_BigRatTo_Reuse writes the values of every class into one z in
// both directions, so the length of the components grows and shrinks.
func TestFraction_BigRatTo_Reuse(t *testing.T) {
	t.Parallel()

	values := agBigRatToValues()
	backward := slices.Clone(values)
	slices.Reverse(backward)
	order := slices.Concat(values, backward)
	z := big.NewRat(5, 7)
	for i, v := range order {
		agCheckBigRatTo(t, fmt.Sprintf("step %d: %s", i, v.name), v.f, z)
	}

	t.Run("no shared memory", func(t *testing.T) {
		t.Parallel()
		g := bp(60, 0)
		f := wFrac(wMul(g, bp(190, 0)), wNeg(wMul(g, bp(190, 1))))
		z1, ok1 := f.BigRatTo(nil)
		z2, ok2 := f.BigRatTo(new(big.Rat))
		require.True(t, ok1)
		require.True(t, ok2)
		require.NotNil(t, z1)
		require.NotNil(t, z2)
		require.False(t, z1 == z2)
		want, _ := f.BigRat()
		f0 := f
		z1.SetInt64(42)
		z1.Num().SetInt64(43)
		assert.Zero(t, z2.Cmp(want), "another result must not change")
		again, _ := f.BigRat()
		assert.Zero(t, again.Cmp(want), "the fraction must not change")
		assert.True(t, f == f0, "the fraction must not change")
	})
}

// TestAggregate_Concurrent calls the new functions at the same time on one
// shared set, including a set with an out-of-range prefix, and compares the
// results with the oracle in the test goroutine (run under -race).
func TestAggregate_Concurrent(t *testing.T) {
	t.Parallel()

	one := big.NewInt(1)
	a, b := bp(200, 0), bp(200, 1)
	sets := [][]fraction.Fraction{
		{agN(1, 4), agN(1, 6), agN(5, 12), agN(-7, 18)},
		{wFrac(one, a), wFrac(one, b), wFrac(wOff(a, -1), a)},
		{wFrac(wPow2(200), one), wFrac(wPow2(200), one), wFrac(one, wPow2(200))},
	}
	type result struct {
		set              int
		sum, product     fraction.Fraction
		sumErr, prodErr  error
		den              int
		denErr           error
		denBig           *big.Int
		denBigErr        error
		gcd, lcm         fraction.Fraction
		gcdErr, lcmErr   error
		rat              *big.Rat
		ratOK, ownRatsOK bool
	}
	const workers = 8
	results := make(chan result, workers*len(sets))
	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			own := new(big.Rat)
			for i := range sets {
				fs := sets[(i+w)%len(sets)]
				r := result{set: (i + w) % len(sets)}
				r.sum, r.sumErr = fraction.Sum(fs...)
				r.product, r.prodErr = fraction.Product(fs...)
				r.den, r.denErr = fraction.CommonDenominator(fs...)
				r.denBig, r.denBigErr = fraction.CommonDenominatorBig(fs...)
				r.gcd, r.gcdErr = fs[0].GCD(fs[1])
				r.lcm, r.lcmErr = fs[0].LCM(fs[1])
				r.rat, r.ratOK = fs[2].BigRatTo(nil)
				_, r.ownRatsOK = fs[1].BigRatTo(own)
				results <- r
			}
		})
	}
	wg.Wait()
	close(results)

	for r := range results {
		fs := sets[r.set]
		label := fmt.Sprintf("set %d", r.set)
		wCheckFrac(t, label+": Sum", r.sum, r.sumErr, agWantSum(fs))
		wCheckFrac(t, label+": Product", r.product, r.prodErr, agWantProduct(fs))
		wantDen, wantDenErr := agWantCommonDenominator(fs)
		if wantDenErr != nil {
			assert.ErrorIs(t, r.denErr, wantDenErr, label)
		} else {
			assert.Equal(t, wantDen, r.den, label)
		}
		wantBig, _ := agWantCommonDenominatorBig(fs)
		if assert.NoError(t, r.denBigErr, label) {
			wCheckBigInt(t, label+": CommonDenominatorBig", r.denBig, wantBig)
		}
		wCheckFrac(t, label+": GCD", r.gcd, r.gcdErr, agWantGCD(fs[0], fs[1]))
		wCheckFrac(t, label+": LCM", r.lcm, r.lcmErr, agWantLCM(fs[0], fs[1]))
		want, _ := fs[2].BigRat()
		if assert.True(t, r.ratOK, label) && assert.NotNil(t, r.rat, label) {
			assert.Zero(t, r.rat.Cmp(want), label)
		}
		assert.True(t, r.ownRatsOK, label)
	}
}
