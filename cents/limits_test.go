package cents_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// k30Values returns the edge values of K-30 (minimum NF-5 (a), (b)).
func k30Values(t testing.TB) []named {
	t.Helper()

	pow := func(b, e int64) *big.Int { return new(big.Int).Exp(big.NewInt(b), big.NewInt(e), nil) }
	r := noErr(cents.FromRatio(1<<62, int(pow(3, 39).Int64())))(t)
	rc, _ := ratioComp(pow(2, 62), pow(3, 39))
	micro := noErr(cents.FromFraction(1, 1000000))(t)
	sav := ratC(6, 5).withExp(5, rat(1, 1000))

	return []named{
		{"<0.000001> + E(1,999983)", noErr(micro.Add(noErr(cents.Step(1, 999983))(t)))(t), ratC(1, 1000000).add(stepC(1, 999983))},
		{"<1/1000000> / 1000000", noErr(micro.Div(1000000))(t), ratC(1, 1000000000000)},
		{"R(2^62:3^39) * 1000", noErr(r.Mul(1000))(t), rc.mulInt(1000)},
		{"R(2^62:3^39) / 999999", noErr(r.Div(999999))(t), rc.scale(rat(1, 999999))},
		{"SAV(1/999983)", noErr(cents.Savart().ScaleRat(big.NewRat(1, 999983)))(t), sav.scale(rat(1, 999983))},
	}
}

// K-29, E-1, G-2, NF-4: the range |c| <= 9 000 000, ends included.
func TestRange(t *testing.T) {
	t.Parallel()

	top, bottom := noErr(cents.FromInt(9000000))(t), noErr(cents.FromInt(-9000000))(t)
	micro, minusOne := noErr(cents.FromFraction(1, 1000000))(t), noErr(cents.FromInt(-1))(t)
	milliStep, fifth := noErr(cents.Step(1, 1000))(t), noErr(cents.FromRatio(3, 2))(t)

	// c returns a call and its Must variant (K-31).
	c := func(plain func() (cents.Cents, error), must func() cents.Cents) ctor {
		return ctor{plain: plain, must: must}
	}

	outOfRange := []struct {
		name string
		call ctor
	}{
		{"<9000001>", c(func() (cents.Cents, error) { return cents.FromInt(9000001) },
			func() cents.Cents { return cents.MustFromInt(9000001) })},
		{"<9000000> + <1/1000000>", c(func() (cents.Cents, error) { return top.Add(micro) },
			func() cents.Cents { return top.MustAdd(micro) })},
		{"<-9000000> - <1>", c(func() (cents.Cents, error) { return bottom.Sub(cents.Cent()) },
			func() cents.Cents { return bottom.MustSub(cents.Cent()) })},
		{"<9000000> - <-9000000>", c(func() (cents.Cents, error) { return top.Sub(bottom) },
			func() cents.Cents { return top.MustSub(bottom) })},
		{"<1200> * 7501", c(func() (cents.Cents, error) { return cents.Octave().Mul(7501) },
			func() cents.Cents { return cents.Octave().MustMul(7501) })},
		{"<1> * MaxInt", c(func() (cents.Cents, error) { return cents.Cent().Mul(maxInt) },
			func() cents.Cents { return cents.Cent().MustMul(maxInt) })},
		{"E(90001, 12)", c(func() (cents.Cents, error) { return cents.Step(90001, 12) },
			func() cents.Cents { return cents.MustStep(90001, 12) })},
		{"7500001 units of 1/1000 octave", c(func() (cents.Cents, error) { return milliStep.Scale(7500001, 1) },
			func() cents.Cents { return milliStep.MustScale(7500001, 1) })},
		{"Parse 9000000.000001", c(func() (cents.Cents, error) { return cents.Parse("9000000.000001") },
			func() cents.Cents { return cents.MustParse("9000000.000001") })},
		{"R(3:2) * 12822", c(func() (cents.Cents, error) { return fifth.Mul(12822) },
			func() cents.Cents { return fifth.MustMul(12822) })},
	}

	for _, tc := range outOfRange {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.call.plain()
			assertResult(t, comp{}, cents.ErrOutOfRange, got, err)
			assertMust(t, got, err, tc.call.must, "the Must variant")
		})
	}

	ok := []struct {
		name string
		call ctor
		want comp
	}{
		{"0 * MaxInt", c(func() (cents.Cents, error) { return cents.Zero().Mul(maxInt) },
			func() cents.Cents { return cents.Zero().MustMul(maxInt) }), zeroComp()},
		{"E(MaxInt, MaxInt)", c(func() (cents.Cents, error) { return cents.Step(maxInt, maxInt) },
			func() cents.Cents { return cents.MustStep(maxInt, maxInt) }), ratC(1200, 1)},
		{"<9000000> + <-1>", c(func() (cents.Cents, error) { return top.Add(minusOne) },
			func() cents.Cents { return top.MustAdd(minusOne) }), ratC(8999999, 1)},
		{"<-9000000> / 2", c(func() (cents.Cents, error) { return bottom.Div(2) },
			func() cents.Cents { return bottom.MustDiv(2) }), ratC(-4500000, 1)},
		// 12821 fifths are 8999765.066095 cents, 12822 are 9000467.021096 (Python
		// decimal, 120 digits).
		{"R(3:2) * 12821", c(func() (cents.Cents, error) { return fifth.Mul(12821) },
			func() cents.Cents { return fifth.MustMul(12821) }), ratio(3, 2).mulInt(12821)},
		{"Parse 9000000", c(func() (cents.Cents, error) { return cents.Parse("9000000") },
			func() cents.Cents { return cents.MustParse("9000000") }), ratC(9000000, 1)},
		{"Parse -9000000 cents", c(func() (cents.Cents, error) { return cents.Parse("-9000000 cents") },
			func() cents.Cents { return cents.MustParse("-9000000 cents") }), ratC(-9000000, 1)},
	}

	for _, tc := range ok {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.call.plain()
			assertResult(t, tc.want, nil, got, err)
			assertMust(t, got, err, tc.call.must, "the Must variant")
		})
	}

	t.Run("oracle range check", func(t *testing.T) {
		t.Parallel()

		assert.False(t, ratio(3, 2).mulInt(12821).outOfRange())
		assert.True(t, ratio(3, 2).mulInt(12822).outOfRange())
	})
}

// minimumA returns values of the minimum NF-5 (a): integers, decimals with
// up to 6 digits, steps with N <= 10^6, ratios of 64-bit integers, the
// named values and MK, units and numbers of units p/q with q <= 10^6.
func minimumA(t testing.TB) []named {
	t.Helper()

	pow := func(b, e int64) *big.Int { return new(big.Int).Exp(big.NewInt(b), big.NewInt(e), nil) }
	r1, _ := ratioComp(pow(31, 12), pow(29, 12))
	sav := ratC(6, 5).withExp(5, rat(1, 1000))

	out := []named{
		{"<8999999>", noErr(cents.FromInt(8999999))(t), ratC(8999999, 1)},
		{"<0.000001>", noErr(cents.FromFloat(0.000001))(t), ratC(1, 1000000)},
		{"<-1234.567891>", noErr(cents.Parse("-1234.567891"))(t), ratC(-1234567891, 1000000)},
		{"E(1,999983)", noErr(cents.Step(1, 999983))(t), stepC(1, 999983)},
		{"E(999999,1000000)", noErr(cents.Step(999999, 1000000))(t), stepC(999999, 1000000)},
		{"R(31^12:29^12)", noErr(cents.FromRatio(int(pow(31, 12).Int64()), int(pow(29, 12).Int64())))(t), r1},
		{"MK", byName(t, setS(t), "MK").v, ratC(-100800, 1).withExp(3, rat(53, 1))},
		{"999999/1000000 MIDI units", noErr(cents.MIDITuningUnit().Scale(999999, 1000000))(t), ratC(25, 4096).scale(rat(999999, 1000000))},
		{"1/999983 jot", noErr(cents.Jot().Scale(1, 999983))(t), ratC(1200, 30103).scale(rat(1, 999983))},
		{"999999/999983 of 1/999983 octave", noErr(noErr(cents.Step(1, 999983))(t).Scale(999999, 999983))(t), stepC(1, 999983).scale(rat(999999, 999983))},
		{"-1/999983 savart", noErr(cents.Savart().Scale(-1, 999983))(t), sav.scale(rat(-1, 999983))},
		{"2258 savarts", noErr(cents.Savart().Scale(2258, 1))(t), sav.mulInt(2258)},
	}

	for _, s := range namedSpecs() {
		out = append(out, named{s.name, s.get(), s.def})
	}

	return out
}

// K-30, E-2, NF-5: the minimum is within the record limits; beyond them the
// error is ErrTooComplex.
func TestRecordLimit(t *testing.T) {
	t.Parallel()

	t.Run("K-30 values", func(t *testing.T) {
		t.Parallel()

		for _, a := range k30Values(t) {
			assertComp(t, a.c, a.v, a.name)
			assert.False(t, a.c.tooComplex(), "oracle: %s within the limits", a.name)

			num, den, ps := a.v.Composition()
			assertCents(t, a.v, noErr(cents.FromComposition(num, den, ps...))(t, "FromComposition of %s", a.name), "from the composition of %s", a.name)
			assertCents(t, a.v, cents.MustFromComposition(num, den, ps...), "MustFromComposition of %s", a.name)
		}
	})

	// NF-5 (b): the sum and difference of two values of (a); a value of (a)
	// times any integer; a share with |n| <= 10^6 — never ErrTooComplex.
	t.Run("minimum (b)", func(t *testing.T) {
		t.Parallel()

		values := minimumA(t)
		for _, a := range values {
			assertComp(t, a.c, a.v, a.name)

			for _, b := range values {
				sum, err := a.v.Add(b.v)
				binaryResult(t, a.name+" + "+b.name, a.c.add(b.c), sum, err)
				assert.NotErrorIs(t, err, cents.ErrTooComplex)
				assertMust(t, sum, err, func() cents.Cents { return a.v.MustAdd(b.v) }, "MustAdd: %s + %s", a.name, b.name)

				diff, err := a.v.Sub(b.v)
				binaryResult(t, a.name+" - "+b.name, a.c.sub(b.c), diff, err)
				assert.NotErrorIs(t, err, cents.ErrTooComplex)
				assertMust(t, diff, err, func() cents.Cents { return a.v.MustSub(b.v) }, "MustSub: %s - %s", a.name, b.name)
			}

			for _, k := range []int{maxInt, minInt, 999999937, -3} {
				got, err := a.v.Mul(k)
				binaryResult(t, a.name+" * k", a.c.mulInt(k), got, err)
				assert.NotErrorIs(t, err, cents.ErrTooComplex, "%s * %d", a.name, k)
				assertMust(t, got, err, func() cents.Cents { return a.v.MustMul(k) }, "MustMul: %s * %d", a.name, k)
			}

			for _, n := range []int{1000000, -1000000, 999983} {
				got, err := a.v.Div(n)
				require.NoError(t, err, "%s / %d", a.name, n)
				assertComp(t, a.c.scale(rat(1, int64(n))), got, "%s / %d", a.name, n)
				assertMust(t, got, err, func() cents.Cents { return a.v.MustDiv(n) }, "MustDiv: %s / %d", a.name, n)
			}
		}
	})

	t.Run("<1> divided by 10^6 repeatedly", func(t *testing.T) {
		t.Parallel()

		c := cents.Cent()
		for i := 1; i <= 5; i++ {
			var err error
			prev := c
			c, err = c.Div(1000000)
			require.NoError(t, err, "division %d", i)
			assertMust(t, c, err, func() cents.Cents { return prev.MustDiv(1000000) }, "MustDiv: division %d", i)
		}

		assertComp(t, bigRatC(new(big.Rat).SetFrac(big.NewInt(1), pow10(30))), c, "10^-30 is within the limit")

		got, err := c.Div(1000000)
		assertResult(t, comp{}, cents.ErrTooComplex, got, err, "the sixth division")
		assertMust(t, got, err, func() cents.Cents { return c.MustDiv(1000000) }, "MustDiv: the sixth division")
	})

	// Architecture D-6: E(1,999983) * 1/MaxInt is within the limit
	// (denominator about 9.2e24), divided once more by MaxInt it is not.
	t.Run("share of a share", func(t *testing.T) {
		t.Parallel()

		first := noErr(noErr(cents.Step(1, 999983))(t).Scale(1, maxInt))(t)
		assertComp(t, stepC(1, 999983).scale(rat(1, maxInt)), first)

		got, err := first.Div(maxInt)
		assertResult(t, comp{}, cents.ErrTooComplex, got, err)
		assertMust(t, got, err, func() cents.Cents { return first.MustDiv(maxInt) }, "MustDiv")
	})

	// A value in range with |e_3| near the limit: times 10^5 the exponent is
	// above 10^24 while the value stays in range (r = -floor(10^20 * J(3)),
	// Python decimal, 120 digits).
	t.Run("multiple beyond the exponent limit", func(t *testing.T) {
		t.Parallel()

		rec := bigRatC(new(big.Rat).SetInt(bigInt("-190195500086538741774448"))).withExp(3, new(big.Rat).SetInt(pow10(20)))
		require.True(t, rec.cmp(zeroComp()) >= 0 && rec.cmp(ratC(1, 1)) < 0, "oracle: the value is in [0, 1)")

		c := fromComp(t, rec)
		got, err := c.Mul(100000)
		assertResult(t, comp{}, cents.ErrTooComplex, got, err)
		assertMust(t, got, err, func() cents.Cents { return c.MustMul(100000) }, "MustMul")
		assert.False(t, rec.mulInt(100000).outOfRange(), "oracle: the multiple is in range")

		got, err = c.Mul(10000)
		require.NoError(t, err, "10^24 is within the limit")
		assertComp(t, rec.mulInt(10000), got)
		assertMust(t, got, err, func() cents.Cents { return c.MustMul(10000) }, "MustMul")
	})

	t.Run("composition limits", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name string
			p    cents.Power
			want comp
			err  error
		}{
			{"exponent denominator 10^30 + 1", cents.Power{Base: 3, Num: big.NewInt(1), Den: new(big.Int).Add(pow10(30), big.NewInt(1))}, comp{}, cents.ErrTooComplex},
			{"exponent denominator 10^30", cents.Power{Base: 3, Num: big.NewInt(1), Den: pow10(30)}, zeroComp().withExp(3, new(big.Rat).SetFrac(big.NewInt(1), pow10(30))), nil},
			{"|e_3| = 10^24 + 1", cents.Power{Base: 3, Num: new(big.Int).Add(pow10(24), big.NewInt(1))}, comp{}, cents.ErrTooComplex},
		} {
			got, err := cents.FromComposition(nil, nil, tc.p)
			assertResult(t, tc.want, tc.err, got, err, tc.name)
			assertMust(t, got, err, func() cents.Cents { return cents.MustFromComposition(nil, nil, tc.p) }, "MustFromComposition: %s", tc.name)
		}
	})
}
