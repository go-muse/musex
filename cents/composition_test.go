package cents_test

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pw returns a Power with the exponent num/den.
func pw(base int, num, den int64) cents.Power {
	return cents.Power{Base: base, Num: big.NewInt(num), Den: big.NewInt(den)}
}

// K-40: values from an exact composition (r; base → exponent).
func TestFromComposition(t *testing.T) {
	t.Parallel()

	b := big.NewInt
	set := setS(t)

	cases := []struct {
		name     string
		num, den *big.Int
		powers   []cents.Power
		want     comp
		err      error
	}{
		{name: "(0; —)", num: b(0), den: b(1), want: zeroComp()},
		{name: "nil, nil", want: zeroComp()},
		{name: "(71/10; —)", num: b(71), den: b(10), want: ratC(71, 10)},
		{name: "(1200/53; —)", num: b(1200), den: b(53), want: stepC(1, 53)},
		{name: "(-1200; 3 → 1)", num: b(-1200), powers: []cents.Power{pw(3, 1, 1)}, want: ratio(3, 2)},
		{name: "(0; 2 → -1, 3 → 1)", powers: []cents.Power{pw(2, -1, 1), pw(3, 1, 1)}, want: ratio(3, 2)},
		{name: "(0; 2 → 1)", powers: []cents.Power{pw(2, 1, 1)}, want: ratC(1200, 1)},
		{name: "(0; 5 → 1/4)", powers: []cents.Power{pw(5, 1, 4)}, want: byName(t, set, "Q").c},
		{name: "(-22800; 3 → 12)", num: b(-22800), powers: []cents.Power{pw(3, 12, 1)}, want: ratio(531441, 524288)},
		{name: "(6/5; 5 → 1/1000)", num: b(6), den: b(5), powers: []cents.Power{pw(5, 1, 1000)}, want: byName(t, set, "SAV").c},
		{name: "(-100800; 3 → 53)", num: b(-100800), powers: []cents.Power{pw(3, 53, 1)}, want: byName(t, set, "MK").c},
		{name: "(0; 3 → 0)", powers: []cents.Power{pw(3, 0, 1)}, want: zeroComp()},
		{name: "(9000000; —)", num: b(9000000), want: ratC(9000000, 1)},
		{name: "(-9000000; —)", num: b(-9000000), want: ratC(-9000000, 1)},
		{name: "(-2400; 3 → 1, 3 → 1)", num: b(-2400), powers: []cents.Power{pw(3, 1, 1), pw(3, 1, 1)}, want: ratio(9, 4)},
		{name: "(0; 3 → 1/2, 3 → -1/2)", powers: []cents.Power{pw(3, 1, 2), pw(3, -1, 2)}, want: zeroComp()},
		{name: "negative denominators", num: b(1200), den: b(-19), powers: []cents.Power{pw(5, -1, -4)}, want: stepC(-1, 19).withExp(5, rat(1, 4))},
		{name: "unreduced", num: b(2400), den: b(38), powers: []cents.Power{pw(5, 2, 8)}, want: stepC(1, 19).withExp(5, rat(1, 4))},
		{name: "nil exponent", powers: []cents.Power{{Base: 5}}, want: zeroComp()},
		{name: "nil exponent denominator", powers: []cents.Power{{Base: 5, Num: b(1)}}, want: ratio(5, 1)},
		{name: "31", powers: []cents.Power{pw(31, 1, 1)}, want: ratio(31, 1)},
		{name: "unsorted bases", num: b(-2400), powers: []cents.Power{pw(17, 1, 1), pw(13, 1, 1), pw(11, -1, 1), pw(5, -1, 1)}, want: ratio(442, 440)},
		// Errors.
		{name: "base 37", powers: []cents.Power{pw(37, 1, 1)}, err: cents.ErrPrimeLimit},
		{name: "base 41", powers: []cents.Power{pw(41, 1, 1)}, err: cents.ErrPrimeLimit},
		{name: "base 1000003", powers: []cents.Power{pw(1000003, 1, 1)}, err: cents.ErrPrimeLimit},
		{name: "base 37 exponent 0", powers: []cents.Power{{Base: 37}}, err: cents.ErrPrimeLimit},
		{name: "base 0", powers: []cents.Power{pw(0, 1, 1)}, err: cents.ErrInvalidArgument},
		{name: "base 1", powers: []cents.Power{pw(1, 1, 1)}, err: cents.ErrInvalidArgument},
		{name: "base -3", powers: []cents.Power{pw(-3, 1, 1)}, err: cents.ErrInvalidArgument},
		{name: "base 4", powers: []cents.Power{pw(4, 1, 1)}, err: cents.ErrInvalidArgument},
		{name: "base 9", powers: []cents.Power{pw(9, 1, 1)}, err: cents.ErrInvalidArgument},
		{name: "base 10", powers: []cents.Power{pw(10, 1, 1)}, err: cents.ErrInvalidArgument},
		{name: "base 4 exponent 0", powers: []cents.Power{{Base: 4}}, err: cents.ErrInvalidArgument},
		{name: "base MinInt", powers: []cents.Power{pw(minInt, 1, 1)}, err: cents.ErrInvalidArgument},
		{name: "base MaxInt", powers: []cents.Power{pw(maxInt, 1, 1)}, err: cents.ErrInvalidArgument},
		{name: "zero denominator of r", num: b(1), den: b(0), err: cents.ErrInvalidArgument},
		{name: "zero denominator of r, zero numerator", num: b(0), den: b(0), err: cents.ErrInvalidArgument},
		{name: "nil numerator, zero denominator", den: b(0), err: cents.ErrInvalidArgument},
		{name: "zero denominator of an exponent", powers: []cents.Power{pw(3, 1, 0)}, err: cents.ErrInvalidArgument},
		{name: "(0; 3 → 6000)", powers: []cents.Power{pw(3, 6000, 1)}, err: cents.ErrOutOfRange},
		{name: "(9000001; —)", num: b(9000001), err: cents.ErrOutOfRange},
		{name: "(9000000; 3 → 1/1000000)", num: b(9000000), powers: []cents.Power{pw(3, 1, 1000000)}, err: cents.ErrOutOfRange},
		// Exponent denominator 10^30 + 1 and |e_3| = 10^24 + 1 (record limit).
		{name: "exponent denominator 10^30+1", powers: []cents.Power{{Base: 3, Num: b(1), Den: new(big.Int).Add(pow10(30), b(1))}}, err: cents.ErrTooComplex},
		{name: "|e_3| = 10^24+1", powers: []cents.Power{{Base: 3, Num: new(big.Int).Add(pow10(24), b(1))}}, err: cents.ErrTooComplex},
		{name: "repeated exponents above 10^24", powers: []cents.Power{{Base: 3, Num: pow10(24)}, {Base: 3, Num: b(1)}}, err: cents.ErrTooComplex},
		{name: "r denominator 10^30+1", num: b(1), den: new(big.Int).Add(pow10(30), b(1)), err: cents.ErrTooComplex},
		// Order of checks: invalid argument, prime limit, too complex, range.
		{name: "(1/0; 37 → 1)", num: b(1), den: b(0), powers: []cents.Power{pw(37, 1, 1)}, err: cents.ErrInvalidArgument},
		{name: "(0; 37 → 1, 3 → 6000)", powers: []cents.Power{pw(37, 1, 1), pw(3, 6000, 1)}, err: cents.ErrPrimeLimit},
		{name: "(0; 37 → 1, 4 → 1)", powers: []cents.Power{pw(37, 1, 1), pw(4, 1, 1)}, err: cents.ErrInvalidArgument},
		{name: "(0; 37 → 1, 3 → 1/0)", powers: []cents.Power{pw(37, 1, 1), pw(3, 1, 0)}, err: cents.ErrInvalidArgument},
		{name: "(0; 37 → 1, 3 → 1/(10^30+1))", powers: []cents.Power{pw(37, 1, 1), {Base: 3, Num: b(1), Den: new(big.Int).Add(pow10(30), b(1))}}, err: cents.ErrPrimeLimit},
		{name: "too complex and out of range", num: b(9000000), powers: []cents.Power{{Base: 3, Num: new(big.Int).Add(pow10(24), b(1))}}, err: cents.ErrTooComplex},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := cents.FromComposition(tc.num, tc.den, tc.powers...)
			assertResult(t, tc.want, tc.err, got, err)
			assertMust(t, got, err, func() cents.Cents { return cents.MustFromComposition(tc.num, tc.den, tc.powers...) }, "MustFromComposition")
		})
	}

	// An exponent at the limit 10^24 with r = -floor(10^24 * J(3)) keeps the
	// value in range (floor by Python decimal, 120 digits).
	t.Run("exponent 10^24 in range", func(t *testing.T) {
		t.Parallel()

		r := bigInt("-1901955000865387417744486732")
		want := bigRatC(new(big.Rat).SetInt(r)).withExp(3, new(big.Rat).SetInt(pow10(24)))
		require.True(t, want.cmp(zeroComp()) >= 0 && want.cmp(ratC(1, 1)) < 0, "oracle: the value is in [0, 1)")

		got, err := cents.FromComposition(r, nil, cents.Power{Base: 3, Num: pow10(24)})
		require.NoError(t, err)
		assertComp(t, want, got)
		assertMust(t, got, err, func() cents.Cents { return cents.MustFromComposition(r, nil, cents.Power{Base: 3, Num: pow10(24)}) }, "MustFromComposition")
	})

	// For all a in S, the 19 named values and the 8 units: the composition
	// of a creates a again (round trip, NF-5 (g)).
	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		values := setS(t)
		for _, nv := range append(cents.NamedValues(), cents.NamedUnits()...) {
			values = append(values, named{nv.Name, nv.Value, compOf(nv.Value)})
		}

		for _, a := range values {
			num, den, ps := a.v.Composition()
			got, err := cents.FromComposition(num, den, ps...)
			require.NoError(t, err, a.name)
			assertCents(t, a.v, got, a.name)
			assertMust(t, got, err, func() cents.Cents { return cents.MustFromComposition(num, den, ps...) }, "MustFromComposition of %s", a.name)

			assertCents(t, a.v, fromComp(t, a.c), "from the oracle record of %s", a.name)
		}
	})

	t.Run("must", func(t *testing.T) {
		t.Parallel()

		assertCents(t, noErr(cents.FromRatio(3, 2))(t), cents.MustFromComposition(big.NewInt(-1200), nil, pw(3, 1, 1)))
		assertPanicsWithErrorIs(t, cents.ErrPrimeLimit, func() { cents.MustFromComposition(nil, nil, pw(37, 1, 1)) })
	})
}

// K-23 v5: the value with all 11 bases to the power 1/2 (P-12).
func TestAllBasesHalf(t *testing.T) {
	t.Parallel()

	ps := []cents.Power{}
	want := zeroComp()
	for _, p := range allBases() {
		ps = append(ps, cents.Power{Base: p, Num: big.NewInt(1), Den: big.NewInt(2)})
		want = want.withExp(p, rat(1, 2))
	}

	c, err := cents.FromComposition(nil, nil, ps...)
	require.NoError(t, err)

	assertComp(t, want, c)
	assertComp(t, ratC(600, 1).withExp(3, rat(1, 2)).withExp(5, rat(1, 2)).withExp(7, rat(1, 2)).
		withExp(11, rat(1, 2)).withExp(13, rat(1, 2)).withExp(17, rat(1, 2)).withExp(19, rat(1, 2)).
		withExp(23, rat(1, 2)).withExp(29, rat(1, 2)).withExp(31, rat(1, 2)), c)

	_, _, powers := c.Composition()
	assert.Len(t, powers, 10)

	const text = "2^(1/2)*3^(1/2)*5^(1/2)*7^(1/2)*11^(1/2)*13^(1/2)*17^(1/2)*19^(1/2)*23^(1/2)*29^(1/2)*31^(1/2)"
	assert.Equal(t, cents.KindPower, c.Kind())
	assert.Equal(t, text, c.String())
	assert.Len(t, text, 94)

	back, err := cents.Parse(c.String())
	require.NoError(t, err)
	assertCents(t, c, back)

	// 600 * log2(2*3*5*...*31) = 600 * log2(200560490130): Python decimal,
	// 120 digits.
	assert.Equal(t, "22527.147882", c.Decimal(6, 0))

	s, ok := c.TextAs(cents.KindRatio, 0)
	assert.False(t, ok)
	assert.Empty(t, s)

	s, ok = c.TextAs(cents.KindPower, 0)
	assert.True(t, ok)
	assertCents(t, c, noErr(cents.Parse(s))(t), "round trip of the power text")

	for i, p := range powers {
		assert.Equal(t, oddPrimes()[i], p.Base, "base %d", i)
		assert.Equal(t, "1/2", new(big.Rat).SetFrac(p.Num, p.Den).RatString(), "exponent of "+strconv.Itoa(p.Base))
	}
}
