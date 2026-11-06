package cents_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// quoExact returns a/b when the records are proportional (the only case
// where the quotient is rational), otherwise ok is false.
func quoExact(a, b comp) (*big.Rat, bool) {
	parts := func(c comp) []*big.Rat { return append([]*big.Rat{c.r}, c.e[:]...) }

	pa, pb := parts(a), parts(b)
	var lambda *big.Rat
	for i := range pb {
		if pb[i].Sign() != 0 {
			lambda = new(big.Rat).Quo(pa[i], pb[i])

			break
		}
	}

	if lambda == nil {
		panic("division by the zero record")
	}

	if !b.scale(lambda).equal(a) {
		return nil, false
	}

	return lambda, true
}

// quoView returns the view A/B of a/b and a bound of its error: with
// |A - a| <= Ea and |B - b| <= Eb (errBound), |A/B - a/b| is at most
// (Ea + |A/B| * Eb) / (|B| - Eb); the bound below doubles it and uses |B|/2
// for |B| - Eb, valid while 2*Eb < |B| (checked below: otherwise the oracle
// panics instead of giving a wrong view).
func quoView(a, b comp) (v, bound *big.Float) {
	va, vb := a.float(), b.float()
	ea, eb := errBound(a), errBound(b)

	absB := newFloat().Abs(vb)
	if newFloat().Mul(eb, newFloat().SetInt64(2)).Cmp(absB) >= 0 {
		panic("oracle precision is not enough for the divisor " + b.String())
	}

	v = newFloat().Quo(va, vb)
	bound = newFloat().Mul(newFloat().Abs(v), eb)
	bound.Add(bound, ea)
	bound.Mul(bound, newFloat().SetInt64(4))
	bound.Quo(bound, absB)

	return v, bound
}

// quoFloatOracle is the expected Quotient.Float64 of an irrational a/b: the
// nearest float64.
func quoFloatOracle(a, b comp) float64 {
	v, bound := quoView(a, b)

	return nearestFloat64(v, bound, "the quotient Float64")
}

// quoKey is a comparable form of a quotient for assertMust: the exact value
// or "irrational", and the decimal view with 30 digits.
func quoKey(q cents.Quotient) string {
	exact, ok := q.Exact()
	s := "irrational"
	if ok && exact != nil {
		s = exact.RatString()
	}

	return s + " " + q.Decimal(30)
}

// assertMustQuo checks MustQuo on the inputs of a.Quo(b) with the results q
// and err (K-31).
func assertMustQuo(t testing.TB, a, b cents.Cents, q cents.Quotient, err error, msgAndArgs ...any) bool {
	t.Helper()

	return assertMust(t, quoKey(q), err, func() string { return quoKey(a.MustQuo(b)) }, msgAndArgs...)
}

// quoDecimalOracle is the expected Quotient.Decimal(digits): correctly
// rounded, half away from zero (only rational quotients have halves).
func quoDecimalOracle(a, b comp, digits int) string {
	digits = max(digits, 0)
	if x, ok := quoExact(a, b); ok {
		return decimalOracle(bigRatC(x), digits, 0)
	}

	v, bound := quoView(a, b)
	scale := newFloat().SetInt(pow10(digits))
	v.Mul(v, scale)
	bound.Mul(bound, scale)
	sign := v.Sign()
	v.Abs(v)

	n, _ := v.Int(nil)
	frac := new(big.Float).SetPrec(oraclePrec).Sub(v, new(big.Float).SetPrec(oraclePrec).SetInt(n))
	half := new(big.Float).SetPrec(oraclePrec).SetFloat64(0.5)
	assertMargin(newFloat().Sub(frac, half), bound, "quotient rounding")

	if frac.Cmp(half) > 0 {
		n.Add(n, big.NewInt(1))
	}

	return styled(unsignedDecimal(n, digits), sign, n.Sign() == 0, 0, false)
}

// assertQuotient checks a quotient: Exact equals want (nil when the
// quotient is irrational), Decimal(9) is the oracle view of a/b.
func assertQuotient(t testing.TB, a, b comp, q cents.Quotient, msgAndArgs ...any) {
	t.Helper()

	want, wantOK := quoExact(a, b)
	got, ok := q.Exact()
	if assert.Equal(t, wantOK, ok, msgAndArgs...) && wantOK && assert.NotNil(t, got, msgAndArgs...) {
		assert.Equal(t, want.RatString(), got.RatString(), msgAndArgs...)
	}

	if !wantOK {
		assert.Nil(t, got, msgAndArgs...)
	}

	assert.Equal(t, quoDecimalOracle(a, b, 9), q.Decimal(9), msgAndArgs...)
}

// unitList returns the units of C-21: the 8 named units, "1/N octave" for
// N in {1, 12, 19, 53, 1000, 1200, 196608} and the cent.
func unitList(t testing.TB) []named {
	t.Helper()

	out := []named{}
	for _, s := range unitSpecs() {
		out = append(out, named{s.name, s.get(), s.def})
	}

	for _, n := range []int64{1, 12, 19, 53, 1000, 1200, 196608} {
		out = append(out, named{"1/" + big.NewInt(n).String() + " octave", noErr(cents.Step(1, int(n)))(t), stepC(1, n)})
	}

	return append(out, named{"cent", cents.Cent(), ratC(1, 1)})
}

// K-21: values in numbers of units (C-21 is Quo with a unit).
func TestQuoUnits(t *testing.T) {
	t.Parallel()

	r32, c700 := noErr(cents.FromRatio(3, 2))(t), noErr(cents.FromInt(700))(t)

	exact := []struct {
		name string
		a, u cents.Cents
		want *big.Rat
	}{
		{"<700> in millioctaves", c700, cents.Millioctave(), big.NewRat(1750, 3)},
		{"<700> in 1/12 octave", c700, noErr(cents.Step(1, 12))(t), big.NewRat(7, 1)},
		{"<700> in semitones", c700, cents.Semitone(), big.NewRat(7, 1)},
		{"<700> in 1/1 octave", c700, noErr(cents.Step(1, 1))(t), big.NewRat(7, 12)},
		{"<700> in octaves", c700, cents.Octave(), big.NewRat(7, 12)},
		{"HOLDER in 1/53 octave", cents.HoldrianComma(), noErr(cents.Step(1, 53))(t), big.NewRat(1, 1)},
		{"<100> in MIDI units", cents.Semitone(), cents.MIDITuningUnit(), big.NewRat(16384, 1)},
		{"<1> in MIDI units", cents.Cent(), cents.MIDITuningUnit(), big.NewRat(4096, 25)},
		{"R(10:1) in savarts", noErr(cents.FromRatio(10, 1))(t), cents.Savart(), big.NewRat(1000, 1)},
		{"SAV in savarts", cents.Savart(), cents.Savart(), big.NewRat(1, 1)},
		{"0 in savarts", cents.Zero(), cents.Savart(), big.NewRat(0, 1)},
		{"0 in jots", cents.Zero(), cents.Jot(), big.NewRat(0, 1)},
		{"<1200> in jots", cents.Octave(), cents.Jot(), big.NewRat(30103, 1)},
	}

	for _, tc := range exact {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			q, err := tc.a.Quo(tc.u)
			require.NoError(t, err)
			assertMustQuo(t, tc.a, tc.u, q, err, "MustQuo")

			got, ok := q.Exact()
			require.True(t, ok)
			require.NotNil(t, got)
			assert.Equal(t, tc.want.RatString(), got.RatString())
		})
	}

	approx := []struct {
		name   string
		a, u   cents.Cents
		digits int
		want   string
	}{
		// Python decimal, 120 digits: 1200*log2(3/2) / (6/5).
		{"R(3:2) in millioctaves", r32, cents.Millioctave(), 9, "584.962500721"},
		// Python decimal, 120 digits: 1200*log2(3/2) / (1200*log2(10)/1000).
		{"R(3:2) in savarts", r32, cents.Savart(), 9, "176.091259056"},
		// Python decimal, 120 digits: 1000 / log10(2).
		{"<1200> in savarts", cents.Octave(), cents.Savart(), 6, "301.029996"},
	}

	for _, tc := range approx {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			q, err := tc.a.Quo(tc.u)
			require.NoError(t, err)
			assertMustQuo(t, tc.a, tc.u, q, err, "MustQuo")
			got, ok := q.Exact()
			assert.False(t, ok)
			assert.Nil(t, got)
			assert.Equal(t, tc.want, q.Decimal(tc.digits))
		})
	}

	t.Run("N < 1", func(t *testing.T) {
		t.Parallel()

		for _, n := range []int{0, -1} {
			got, err := cents.Step(1, n)
			assertOnlyError(t, err, cents.ErrInvalidArgument, "1/%d octave", n)
			assertMust(t, got, err, func() cents.Cents { return cents.MustStep(1, n) }, "MustStep(1, %d)", n)
		}
	})

	t.Run("for all S and units", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			for _, u := range unitList(t) {
				q, err := a.v.Quo(u.v)
				require.NoError(t, err, "%s in %s", a.name, u.name)
				assertQuotient(t, a.c, u.c, q, "%s in %s", a.name, u.name)
				assertMustQuo(t, a.v, u.v, q, err, "MustQuo: %s in %s", a.name, u.name)
			}
		}
	})
}

// K-41: the quotient of two values.
func TestQuo(t *testing.T) {
	t.Parallel()

	set := setS(t)
	pyth, synt := cents.PythagoreanComma(), cents.SyntonicComma()
	r32, r54 := noErr(cents.FromRatio(3, 2))(t), noErr(cents.FromRatio(5, 4))(t)

	exact := []struct {
		name string
		a, b cents.Cents
		want *big.Rat
		str  string
	}{
		{"<700> / <100>", noErr(cents.FromInt(700))(t), cents.Semitone(), big.NewRat(7, 1), "7"},
		{"<1200> / HOLDER", cents.Octave(), cents.HoldrianComma(), big.NewRat(53, 1), "53"},
		{"<1200> / E(1,19)", cents.Octave(), noErr(cents.Step(1, 19))(t), big.NewRat(19, 1), "19"},
		{"R(3:2) / R(9:4)", r32, noErr(cents.FromRatio(9, 4))(t), big.NewRat(1, 2), "1/2"},
		{"Q / R(5:1)", byName(t, set, "Q").v, noErr(cents.FromRatio(5, 1))(t), big.NewRat(1, 4), "1/4"},
		{"<100> / <-300>", cents.Semitone(), noErr(cents.FromInt(-300))(t), big.NewRat(-1, 3), "-1/3"},
		{"0 / PYTH", cents.Zero(), pyth, big.NewRat(0, 1), "0"},
		{"<700> / millioctave", noErr(cents.FromInt(700))(t), cents.Millioctave(), big.NewRat(1750, 3), "1750/3"},
	}

	for _, tc := range exact {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			q, err := tc.a.Quo(tc.b)
			require.NoError(t, err)
			assertMustQuo(t, tc.a, tc.b, q, err, "MustQuo")
			got, ok := q.Exact()
			require.True(t, ok)
			require.NotNil(t, got)
			assert.Equal(t, tc.want.RatString(), got.RatString())
			assert.Equal(t, tc.str, q.String())

			f, _ := tc.want.Float64()
			assert.Equal(t, f, q.Float64())
		})
	}

	approx := []struct {
		name string
		a, b cents.Cents
		want string
	}{
		// Python decimal, 120 digits, 6 digits after the point.
		{"<1200> / PYTH", cents.Octave(), pyth, "51.150872"},
		{"PYTH / SYNT", pyth, synt, "1.090844"},
		{"R(5:4) / R(3:2)", r54, r32, "0.550340"},
		{"<1200> / SAV", cents.Octave(), cents.Savart(), "301.029996"},
	}

	for _, tc := range approx {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			q, err := tc.a.Quo(tc.b)
			require.NoError(t, err)
			assertMustQuo(t, tc.a, tc.b, q, err, "MustQuo")
			got, ok := q.Exact()
			assert.False(t, ok)
			assert.Nil(t, got)
			assert.Equal(t, tc.want, q.Decimal(6))
		})
	}

	t.Run("String of an irrational quotient", func(t *testing.T) {
		t.Parallel()

		// Python decimal, 120 digits: 1200 / (1200*log2(3^12/2^19)) with 15
		// digits after the point.
		q := noErr(cents.Octave().Quo(pyth))(t)
		assert.Equal(t, "51.150872498556794...", q.String())
		assert.Equal(t, q.Decimal(15)+"...", q.String())

		assert.Equal(t, quoFloatOracle(ratC(1200, 1), ratio(531441, 524288)), q.Float64())
	})

	// K-41: the guarantees of Quotient do not depend on the size of the
	// divisor. d is the irrational value nearest to zero of the form
	// (num/10^30; 3 → 1) within the record limit: num = -floor(J(3)*10^30),
	// d = J(3) + num/10^30 ≈ 8.105e-31 cents, d > 0.
	t.Run("divisor near zero", func(t *testing.T) {
		t.Parallel()

		num := bigInt("-1901955000865387417744486732737379")
		dc := bigRatC(new(big.Rat).SetFrac(num, pow10(30))).withExp(3, rat(1, 1))
		d := noErr(cents.FromComposition(num, pow10(30), cents.Power{Base: 3, Num: big.NewInt(1)}))(t)
		require.Equal(t, "0.00000000000000000000000000000081051", d.Decimal(35, 0), "the divisor")

		cent, octave := cents.Cent(), cents.Octave()
		r32 := noErr(cents.FromRatio(3, 2))(t)

		// Python decimal, 300 and 400 digits (the same views): Float64 is
		// the nearest float64, Decimal is correctly rounded (irrational
		// quotients have no halves), String is Decimal(15) + "...".
		irrational := []struct {
			name             string
			a, b             cents.Cents
			ac, bc           comp
			float            float64
			d0, d5, d15, d40 string
		}{
			{
				"Cent / d", cent, d, ratC(1, 1), dc, 1.2337883643646922e+30,
				"1233788364364692242624011718053", "1233788364364692242624011718052.67723",
				"1233788364364692242624011718052.677232451076736", "",
			},
			{
				"Octave / d", octave, d, ratC(1200, 1), dc, 1.4805460372376307e+33,
				"1480546037237630691148814061663213", "1480546037237630691148814061663212.67894",
				"1480546037237630691148814061663212.678941292082814", "",
			},
			{
				"R(3:2) / d", r32, d, ratio(3, 2), dc, 8.660639123753225e+32,
				"866063912375322469851893079665263", "866063912375322469851893079665263.46379",
				"866063912375322469851893079665263.463792308179995", "",
			},
			{
				"Cent / -d", cent, d.Neg(), ratC(1, 1), dc.neg(), -1.2337883643646922e+30,
				"-1233788364364692242624011718053", "-1233788364364692242624011718052.67723",
				"-1233788364364692242624011718052.677232451076736", "",
			},
			{
				"-Cent / d", cent.Neg(), d, ratC(-1, 1), dc, -1.2337883643646922e+30,
				"-1233788364364692242624011718053", "-1233788364364692242624011718052.67723",
				"-1233788364364692242624011718052.677232451076736", "",
			},
			{
				"d / Cent", d, cent, dc, ratC(1, 1), 8.10511777289231e-31,
				"0", "0.00000", "0.000000000000000",
				"0.0000000000000000000000000000008105117773",
			},
		}

		for _, tc := range irrational {
			q, err := tc.a.Quo(tc.b)
			require.NoError(t, err, tc.name)
			assertMustQuo(t, tc.a, tc.b, q, err, "MustQuo: %s", tc.name)

			exact, ok := q.Exact()
			assert.False(t, ok, tc.name)
			assert.Nil(t, exact, tc.name)
			assert.Equal(t, tc.float, q.Float64(), tc.name)
			assert.Equal(t, tc.d0, q.Decimal(0), tc.name)
			assert.Equal(t, tc.d5, q.Decimal(5), tc.name)
			assert.Equal(t, tc.d15, q.Decimal(15), tc.name)
			assert.Equal(t, tc.d15+"...", q.String(), tc.name)

			if tc.d40 != "" {
				assert.Equal(t, tc.d40, q.Decimal(40), tc.name)
			}

			// The Go oracle of the tests agrees (512 bits).
			assertQuotient(t, tc.ac, tc.bc, q, tc.name)
			assert.Equal(t, quoFloatOracle(tc.ac, tc.bc), q.Float64(), tc.name)
		}

		exact := []struct {
			name string
			a    cents.Cents
			want string
		}{
			{"d / d", d, "1"},
			{"2d / d", d.MustMul(2), "2"},
		}

		for _, tc := range exact {
			q, err := tc.a.Quo(d)
			require.NoError(t, err, tc.name)
			assertMustQuo(t, tc.a, d, q, err, "MustQuo: %s", tc.name)

			got, ok := q.Exact()
			require.True(t, ok, tc.name)
			require.NotNil(t, got, tc.name)
			assert.Equal(t, tc.want, got.RatString(), tc.name)
			assert.Equal(t, tc.want, q.String(), tc.name)
			assert.Equal(t, tc.want, q.Decimal(0), tc.name)
			assert.Equal(t, tc.want+".00000", q.Decimal(5), tc.name)

			f, _ := got.Float64()
			assert.Equal(t, f, q.Float64(), tc.name)
		}
	})

	t.Run("divisor zero", func(t *testing.T) {
		t.Parallel()

		for _, a := range set {
			for _, zero := range []cents.Cents{cents.Zero(), {}} {
				q, err := a.v.Quo(zero)
				assertOnlyError(t, err, cents.ErrInvalidArgument, a.name)
				assertMustQuo(t, a.v, zero, q, err, "MustQuo: %s / 0", a.name)

				exact, ok := q.Exact()
				assert.True(t, ok, "the result is the zero Quotient")
				require.NotNil(t, exact)
				assert.Equal(t, 0, exact.Sign())
				assert.Equal(t, "0", q.String())
			}
		}
	})

	t.Run("a / a = 1", func(t *testing.T) {
		t.Parallel()

		for _, a := range set {
			if a.c.sign() == 0 {
				continue
			}

			q := noErr(a.v.Quo(a.v))(t)
			got, ok := q.Exact()
			require.True(t, ok, a.name)
			assert.Equal(t, "1", got.RatString(), a.name)
		}
	})

	// For all a in S: a / <100>, a / <1200>, a / HOLDER, a / E(1,19) equal
	// the numbers of units "1/N octave" for N = 12, 1, 53, 19.
	t.Run("same as units", func(t *testing.T) {
		t.Parallel()

		units := []struct {
			u cents.Cents
			n int
		}{
			{cents.Semitone(), 12},
			{cents.Octave(), 1},
			{cents.HoldrianComma(), 53},
			{noErr(cents.Step(1, 19))(t), 19},
		}

		for _, a := range set {
			for _, u := range units {
				q1 := noErr(a.v.Quo(u.u))(t)
				q2 := noErr(a.v.Quo(noErr(cents.Step(1, u.n))(t)))(t)

				e1, ok1 := q1.Exact()
				e2, ok2 := q2.Exact()
				assert.Equal(t, ok2, ok1, "%s / unit %d", a.name, u.n)

				if ok1 && ok2 {
					assert.Equal(t, e2.RatString(), e1.RatString(), "%s / unit %d", a.name, u.n)
				}

				assert.Equal(t, q2.Decimal(12), q1.Decimal(12), "%s / unit %d", a.name, u.n)
			}
		}
	})

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		for _, a := range set {
			for _, b := range set {
				if b.c.sign() == 0 {
					continue
				}

				q, err := a.v.Quo(b.v)
				require.NoError(t, err, "%s / %s", a.name, b.name)
				assertQuotient(t, a.c, b.c, q, "%s / %s", a.name, b.name)
				assertMustQuo(t, a.v, b.v, q, err, "MustQuo: %s / %s", a.name, b.name)
			}
		}
	})
}

// Quotient methods: the zero value, String, Decimal and its rounding.
func TestQuotientMethods(t *testing.T) {
	t.Parallel()

	t.Run("zero value", func(t *testing.T) {
		t.Parallel()

		var q cents.Quotient

		exact, ok := q.Exact()
		require.True(t, ok)
		require.NotNil(t, exact)
		assert.Equal(t, 0, exact.Sign())
		assert.Zero(t, q.Float64())
		assert.Equal(t, "0", q.String())
		assert.Equal(t, "0", q.Decimal(0))
		assert.Equal(t, "0.00", q.Decimal(2))
	})

	t.Run("decimal", func(t *testing.T) {
		t.Parallel()

		eighth := noErr(cents.Cent().Quo(noErr(cents.FromInt(8))(t)))(t)
		assert.Equal(t, "0.13", eighth.Decimal(2), "half away from zero")
		assert.Equal(t, "-0.13", noErr(noErr(cents.FromInt(-1))(t).Quo(noErr(cents.FromInt(8))(t)))(t).Decimal(2))
		assert.Equal(t, "0.13", noErr(cents.Cent().Quo(noErr(cents.FromInt(8))(t)))(t).Decimal(2))
		assert.Equal(t, "-0.13", noErr(cents.Cent().Quo(noErr(cents.FromInt(-8))(t)))(t).Decimal(2))
		assert.Equal(t, "0", eighth.Decimal(0))
		assert.Equal(t, eighth.Decimal(0), eighth.Decimal(-1), "negative digits as 0")
		assert.Equal(t, eighth.Decimal(0), eighth.Decimal(minInt), "negative digits as 0")

		small := noErr(noErr(cents.FromInt(-1))(t).Quo(noErr(cents.FromInt(1000))(t)))(t)
		assert.Equal(t, "0.00", small.Decimal(2), "rounded to zero without a minus")
		assert.Equal(t, "-0.001", small.Decimal(3))
		assert.Equal(t, "-1/1000", small.String())

		third := noErr(cents.Cent().Quo(noErr(cents.FromInt(3))(t)))(t)
		assert.Equal(t, "0.333333", third.Decimal(6))
		assert.Equal(t, "1/3", third.String())

		pyth := noErr(cents.Octave().Quo(cents.PythagoreanComma()))(t)
		assert.Equal(t, "51", pyth.Decimal(0), "no point when digits is 0")
		assert.Equal(t, "51.2", pyth.Decimal(1))
	})

	// K-41, NF-11: hundreds of digits are still correctly rounded; the
	// literals are in testdata (decimalLiteral).
	t.Run("hundreds of digits", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			file string
			a, b cents.Cents
		}{
			{"octave-pyth-900", cents.Octave(), cents.PythagoreanComma()},
			{"r32-sav-900", cents.MustFromRatio(3, 2), cents.Savart()},
		}

		for _, tc := range cases {
			want := decimalLiteral(t, tc.file)
			got := tc.a.MustQuo(tc.b).Decimal(900)
			assertDecimalShape(t, got, 900, tc.file)
			assert.Equal(t, want, got, tc.file)
		}
	})

	t.Run("float64 midpoint", func(t *testing.T) {
		t.Parallel()

		// (2^53+1)/2^30 is exactly between two float64; ties to even
		// (Python fractions).
		q := noErr(noErr(cents.FromFraction(1<<53+1, 1<<30))(t).Quo(cents.Cent()))(t)
		assert.Equal(t, 8388608.0, q.Float64())
	})

	t.Run("Exact returns a new value", func(t *testing.T) {
		t.Parallel()

		q := noErr(cents.Octave().Quo(cents.HoldrianComma()))(t)
		e1, ok := q.Exact()
		require.True(t, ok)
		e1.SetInt64(7)

		e2, ok := q.Exact()
		require.True(t, ok)
		assert.Equal(t, "53", e2.RatString())
		assert.Equal(t, "53", q.String())
	})

	t.Run("copy", func(t *testing.T) {
		t.Parallel()

		q := noErr(cents.Octave().Quo(cents.HoldrianComma()))(t)
		c := q
		assert.Equal(t, q.String(), c.String())
	})
}

// scaleBy returns u.Scale(num, den) and u.MustScale(num, den) (K-22: the
// value of a number of units).
func scaleBy(u cents.Cents, num, den int) ctor {
	return ctor{
		plain: func() (cents.Cents, error) { return u.Scale(num, den) },
		must:  func() cents.Cents { return u.MustScale(num, den) },
	}
}

// K-22: values from a number of units, exact for every unit.
func TestFromUnits(t *testing.T) {
	t.Parallel()

	step1000, step301, step300 := noErr(cents.Step(1, 1000))(t), noErr(cents.Step(1, 301))(t), noErr(cents.Step(1, 300))(t)

	cases := []struct {
		name string
		got  ctor
		want comp
	}{
		{"1750/3 of 1/1000 octave", scaleBy(step1000, 1750, 3), ratC(700, 1)},
		{"301 of 1/301 octave", scaleBy(step301, 301, 1), ratC(1200, 1)},
		{"300 of 1/300 octave", scaleBy(step300, 300, 1), ratC(1200, 1)},
		{"1 centitone", scaleBy(cents.Centitone(), 1, 1), ratC(2, 1)},
		{"1 jot", scaleBy(cents.Jot(), 1, 1), ratC(1200, 30103)},
		{"1 meride", scaleBy(cents.Meride(), 1, 1), ratC(1200, 43)},
		{"1 eptameride", scaleBy(cents.Eptameride(), 1, 1), ratC(1200, 301)},
		{"1 MIDI unit", scaleBy(cents.MIDITuningUnit(), 1, 1), ratC(25, 4096)},
		{"1 millioctave", scaleBy(cents.Millioctave(), 1, 1), ratC(6, 5)},
		{"1 savart", scaleBy(cents.Savart(), 1, 1), ratC(6, 5).withExp(5, rat(1, 1000))},
		{"1000 savarts", scaleBy(cents.Savart(), 1000, 1), ratio(10, 1)},
		{"1/999983 savart", ctor{
			plain: func() (cents.Cents, error) { return cents.Savart().ScaleRat(big.NewRat(1, 999983)) },
			must:  func() cents.Cents { return cents.Savart().MustScaleRat(big.NewRat(1, 999983)) },
		}, ratC(6, 5).withExp(5, rat(1, 1000)).scale(rat(1, 999983))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.got.plain()
			require.NoError(t, err)
			assertComp(t, tc.want, got)
			assertMust(t, got, err, tc.got.must, "the Must variant")
		})
	}

	t.Run("savart decimal and 1000 savarts", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "3.986314", cents.Savart().Decimal(6, 0))
		assertCents(t, noErr(cents.FromRatio(10, 1))(t), cents.Savart().MustScale(1000, 1))
	})

	// For all a in S whose number of units is exact: from the number back to
	// the same value.
	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			for _, u := range unitList(t) {
				q := noErr(a.v.Quo(u.v))(t)
				exact, ok := q.Exact()
				if !ok {
					continue
				}

				back, err := u.v.ScaleRat(exact)
				require.NoError(t, err, "%s in %s", a.name, u.name)
				assertCents(t, a.v, back, "%s in %s and back", a.name, u.name)
				assertMust(t, back, err, func() cents.Cents { return u.v.MustScaleRat(exact) }, "MustScaleRat: %s in %s", a.name, u.name)

				if exact.Num().IsInt64() && exact.Denom().IsInt64() {
					num, den := int(exact.Num().Int64()), int(exact.Denom().Int64())
					back, err = u.v.Scale(num, den)
					require.NoError(t, err, "%s in %s by Scale", a.name, u.name)
					assertMust(t, back, err, func() cents.Cents { return u.v.MustScale(num, den) }, "MustScale: %s in %s", a.name, u.name)
					assertCents(t, a.v, back, "%s in %s and back by Scale", a.name, u.name)
				}
			}
		}
	})

	// A number with a denominator beyond the record limit (architecture D-6):
	// 6/5 * 1/10^30 has the denominator 2.5*10^30 > 10^30 (Python fractions).
	t.Run("too complex", func(t *testing.T) {
		t.Parallel()

		r := new(big.Rat).SetFrac(big.NewInt(1), pow10(30))
		got, err := cents.Millioctave().ScaleRat(r)
		assertResult(t, comp{}, cents.ErrTooComplex, got, err)
		assertMust(t, got, err, func() cents.Cents { return cents.Millioctave().MustScaleRat(r) }, "MustScaleRat")
	})
}
