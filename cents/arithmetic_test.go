package cents_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// binaryResult runs op on a and b and checks the result against the record
// oracle: the exact record, or ErrTooComplex / ErrOutOfRange by the record of
// the exact result (order: record limit, then range).
func binaryResult(t *testing.T, name string, want comp, got cents.Cents, err error) bool {
	t.Helper()

	return assertResult(t, want, want.wantErr(), got, err, name)
}

// checked runs the operations of fixed cases and checks the Must variant of
// each call on the same inputs (K-31, assertMust); an error of the
// operation fails the test.
type checked struct{ t *testing.T }

func (o checked) add(a, b cents.Cents) cents.Cents {
	o.t.Helper()

	got, err := a.Add(b)
	assertMust(o.t, got, err, func() cents.Cents { return a.MustAdd(b) }, "MustAdd: %s + %s", a, b)
	require.NoError(o.t, err, "%s + %s", a, b)

	return got
}

func (o checked) sub(a, b cents.Cents) cents.Cents {
	o.t.Helper()

	got, err := a.Sub(b)
	assertMust(o.t, got, err, func() cents.Cents { return a.MustSub(b) }, "MustSub: %s - %s", a, b)
	require.NoError(o.t, err, "%s - %s", a, b)

	return got
}

func (o checked) mul(a cents.Cents, k int) cents.Cents {
	o.t.Helper()

	got, err := a.Mul(k)
	assertMust(o.t, got, err, func() cents.Cents { return a.MustMul(k) }, "MustMul: %s * %d", a, k)
	require.NoError(o.t, err, "%s * %d", a, k)

	return got
}

func (o checked) div(a cents.Cents, n int) cents.Cents {
	o.t.Helper()

	got, err := a.Div(n)
	assertMust(o.t, got, err, func() cents.Cents { return a.MustDiv(n) }, "MustDiv: %s / %d", a, n)
	require.NoError(o.t, err, "%s / %d", a, n)

	return got
}

func (o checked) scale(a cents.Cents, num, den int) cents.Cents {
	o.t.Helper()

	got, err := a.Scale(num, den)
	assertMust(o.t, got, err, func() cents.Cents { return a.MustScale(num, den) }, "MustScale: %s * %d/%d", a, num, den)
	require.NoError(o.t, err, "%s * %d/%d", a, num, den)

	return got
}

// K-10, G-1: sum and difference.
func TestAddSub(t *testing.T) {
	t.Parallel()

	t.Run("criteria", func(t *testing.T) {
		t.Parallel()

		op := checked{t}
		r32, r43 := noErr(cents.FromRatio(3, 2))(t), noErr(cents.FromRatio(4, 3))(t)
		r54, r65 := noErr(cents.FromRatio(5, 4))(t), noErr(cents.FromRatio(6, 5))(t)

		assertCents(t, cents.Octave(), op.add(r32, r43), "R(3:2) + R(4:3)")
		assertCents(t, r32, op.add(r54, r65), "R(5:4) + R(6:5)")
		assertCents(t, noErr(cents.FromFraction(14500, 19))(t),
			op.add(noErr(cents.FromInt(700))(t), noErr(cents.Step(1, 19))(t)), "<700> + E(1,19)")
		assertCents(t, noErr(cents.FromFraction(208, 10))(t),
			op.sub(noErr(cents.FromFraction(71, 10))(t), noErr(cents.FromFraction(-137, 10))(t)), "<7.1> - <-13.7>")
	})

	t.Run("edges", func(t *testing.T) {
		t.Parallel()

		top, bottom := noErr(cents.FromInt(9000000))(t), noErr(cents.FromInt(-9000000))(t)
		micro := noErr(cents.FromFraction(1, 1000000))(t)

		got, err := top.Add(micro)
		assertResult(t, comp{}, cents.ErrOutOfRange, got, err, "<9000000> + <1/1000000>")
		assertMust(t, got, err, func() cents.Cents { return top.MustAdd(micro) }, "MustAdd")

		got, err = bottom.Sub(cents.Cent())
		assertResult(t, comp{}, cents.ErrOutOfRange, got, err, "<-9000000> - <1>")
		assertMust(t, got, err, func() cents.Cents { return bottom.MustSub(cents.Cent()) }, "MustSub")

		got, err = top.Sub(bottom)
		assertResult(t, comp{}, cents.ErrOutOfRange, got, err, "<9000000> - <-9000000>")
		assertMust(t, got, err, func() cents.Cents { return top.MustSub(bottom) }, "MustSub")

		got, err = bottom.Sub(top)
		assertResult(t, comp{}, cents.ErrOutOfRange, got, err, "<-9000000> - <9000000>")
		assertMust(t, got, err, func() cents.Cents { return bottom.MustSub(top) }, "MustSub")

		op := checked{t}
		assertComp(t, ratC(8999999, 1), op.add(top, noErr(cents.FromInt(-1))(t)), "<9000000> + <-1>")
		assert.True(t, op.add(top, bottom) == cents.Cents{}, "<9000000> + <-9000000>")
	})

	set := setS(t)
	for _, a := range set {
		t.Run(a.name, func(t *testing.T) {
			t.Parallel()

			for _, b := range set {
				sum, errSum := a.v.Add(b.v)
				binaryResult(t, a.name+" + "+b.name, a.c.add(b.c), sum, errSum)
				assertMust(t, sum, errSum, func() cents.Cents { return a.v.MustAdd(b.v) }, "MustAdd: %s + %s", a.name, b.name)

				diff, errDiff := a.v.Sub(b.v)
				binaryResult(t, a.name+" - "+b.name, a.c.sub(b.c), diff, errDiff)
				assertMust(t, diff, errDiff, func() cents.Cents { return a.v.MustSub(b.v) }, "MustSub: %s - %s", a.name, b.name)

				if errSum == nil {
					assertCents(t, sum, noErr(b.v.Add(a.v))(t, "%s + %s", b.name, a.name), "%s + %s = %s + %s", a.name, b.name, b.name, a.name)
					assertCents(t, a.v, noErr(sum.Sub(b.v))(t, "(%s + %s) - %s", a.name, b.name, b.name), "(%s + %s) - %s", a.name, b.name, b.name)
				}

				if errDiff == nil {
					assertCents(t, diff, noErr(b.v.Sub(a.v))(t, "%s - %s", b.name, a.name).Neg(), "%s - %s = -(%s - %s)", a.name, b.name, b.name, a.name)
				}
			}

			op := checked{t}
			assert.True(t, op.sub(a.v, a.v) == cents.Cents{}, "a - a")
			assertCents(t, a.v, op.add(a.v, cents.Zero()), "a + 0")
		})
	}
}

// K-11: negation, absolute value and sign.
func TestNegAbsSign(t *testing.T) {
	t.Parallel()

	r32, r23 := noErr(cents.FromRatio(3, 2))(t), noErr(cents.FromRatio(2, 3))(t)
	assertCents(t, r23, r32.Neg(), "-R(3:2)")
	assertCents(t, r32, r23.Abs(), "|R(2:3)|")
	assert.True(t, cents.Zero().Neg() == cents.Cents{}, "-0")
	assert.True(t, cents.Zero().Abs() == cents.Cents{}, "|0|")
	assert.Equal(t, -1, r23.Sign())
	assert.Equal(t, 0, cents.Zero().Sign())
	assert.Equal(t, 1, cents.Schisma().Sign())
	assertComp(t, ratC(9000000, 1), noErr(cents.FromInt(-9000000))(t).Abs(), "|<-9000000>|")
	assertComp(t, ratC(9000000, 1), noErr(cents.FromInt(9000000))(t).Abs(), "|<9000000>|")

	for _, a := range setS(t) {
		assertComp(t, a.c.neg(), a.v.Neg(), "-%s", a.name)
		assertCents(t, a.v, a.v.Neg().Neg(), "-(-%s)", a.name)
		assertCents(t, a.v.Abs(), a.v.Neg().Abs(), "|%s| = |-%s|", a.name, a.name)
		assert.GreaterOrEqual(t, a.v.Abs().Compare(cents.Zero()), 0, "|%s| >= 0", a.name)
		assert.Equal(t, a.c.sign(), a.v.Sign(), "sign of %s", a.name)

		wantAbs := a.c
		if a.c.sign() < 0 {
			wantAbs = a.c.neg()
		}

		assertComp(t, wantAbs, a.v.Abs(), "|%s|", a.name)
	}
}

// K-12: multiplication by an integer.
func TestMul(t *testing.T) {
	t.Parallel()

	t.Run("criteria and edges", func(t *testing.T) {
		t.Parallel()

		op := checked{t}
		pyth := op.sub(op.mul(noErr(cents.FromRatio(3, 2))(t), 12), noErr(cents.FromInt(8400))(t))
		assertCents(t, cents.PythagoreanComma(), pyth, "R(3:2)*12 - <8400>")
		assertComp(t, ratC(9000000, 1), op.mul(cents.Octave(), 7500), "<1200>*7500")
		assertComp(t, ratC(-9000000, 1), op.mul(cents.Octave(), -7500), "<1200>*-7500")

		for _, tc := range []struct {
			name string
			c    cents.Cents
			k    int
		}{
			{"<1200>*7501", cents.Octave(), 7501},
			{"<1200>*-7501", cents.Octave(), -7501},
			{"<1>*MaxInt", cents.Cent(), maxInt},
			{"<1>*MinInt", cents.Cent(), minInt},
			{"R(3:2)*MaxInt", noErr(cents.FromRatio(3, 2))(t), maxInt},
		} {
			got, err := tc.c.Mul(tc.k)
			assertResult(t, comp{}, cents.ErrOutOfRange, got, err, tc.name)
			assertMust(t, got, err, func() cents.Cents { return tc.c.MustMul(tc.k) }, "MustMul: %s", tc.name)
		}

		assert.True(t, op.mul(cents.Zero(), maxInt) == cents.Cents{}, "0*MaxInt")
		assert.True(t, op.mul(cents.Zero(), minInt) == cents.Cents{}, "0*MinInt")
	})

	set := setS(t)
	for _, a := range set {
		t.Run(a.name, func(t *testing.T) {
			t.Parallel()

			for k := -4; k <= 4; k++ {
				got, err := a.v.Mul(k)
				binaryResult(t, a.name+" * k", a.c.mulInt(k), got, err)
				assertMust(t, got, err, func() cents.Cents { return a.v.MustMul(k) }, "MustMul: %s * %d", a.name, k)

				byScale, errScale := a.v.Scale(k, 1)
				assert.True(t, got == byScale, "Mul(%d) = Scale(%d, 1)", k, k)
				assert.Equal(t, errName(err), errName(errScale), "Mul(%d) = Scale(%d, 1)", k, k)
				assertMust(t, byScale, errScale, func() cents.Cents { return a.v.MustScale(k, 1) }, "MustScale: %s * %d/1", a.name, k)

				if err != nil {
					continue
				}

				if twice, err2 := a.v.Mul(2 * k); err2 == nil {
					assertCents(t, twice, noErr(got.Mul(2))(t, "(%s * %d) * 2", a.name, k), "(%s * %d) * 2", a.name, k)
				}
			}

			op := checked{t}
			assert.True(t, op.mul(a.v, 0) == cents.Cents{}, "a * 0")
			assertCents(t, a.v, op.mul(a.v, 1), "a * 1")
			assertCents(t, a.v.Neg(), op.mul(a.v, -1), "a * -1")

			if double, err := a.v.Mul(2); err == nil {
				assertCents(t, double, op.add(a.v, a.v), "a * 2 = a + a")
			}
		})
	}
}

// K-13: shares (division by a non-zero integer).
func TestDiv(t *testing.T) {
	t.Parallel()

	t.Run("criteria", func(t *testing.T) {
		t.Parallel()

		op := checked{t}
		r32 := noErr(cents.FromRatio(3, 2))(t)
		twelfth := op.div(cents.PythagoreanComma(), 12)
		assertCents(t, op.sub(r32, noErr(cents.FromInt(700))(t)), twelfth, "PYTH / 12 = R(3:2) - <700>")
		assert.Equal(t, "1.955001", twelfth.Decimal(6, 0))
		assert.Equal(t, "5.376572", op.div(cents.SyntonicComma(), 4).Decimal(6, 0))
		assertComp(t, ratC(100, 3), op.div(cents.Semitone(), 3), "<100> / 3")

		mk := byName(t, setS(t), "MK")
		assertCents(t, mk.v, op.mul(op.div(mk.v, 53), 53), "MK / 53 * 53")

		assertComp(t, ratC(-4500000, 1), op.div(noErr(cents.FromInt(-9000000))(t), 2), "<-9000000> / 2")
		assertComp(t, bigRatC(new(big.Rat).SetFrac(big.NewInt(-1), new(big.Int).Lsh(big.NewInt(1), 63))),
			op.div(cents.Cent(), minInt), "<1> / MinInt")
		assertComp(t, ratC(1, maxInt), op.div(cents.Cent(), maxInt), "<1> / MaxInt")
	})

	t.Run("by zero", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			got, err := a.v.Div(0)
			assertResult(t, comp{}, cents.ErrInvalidArgument, got, err, "%s / 0", a.name)
			assertMust(t, got, err, func() cents.Cents { return a.v.MustDiv(0) }, "MustDiv: %s / 0", a.name)

			got, err = a.v.Scale(1, 0)
			assertResult(t, comp{}, cents.ErrInvalidArgument, got, err, "%s.Scale(1, 0)", a.name)
			assertMust(t, got, err, func() cents.Cents { return a.v.MustScale(1, 0) }, "MustScale: %s * 1/0", a.name)
		}
	})

	set := setS(t)
	for _, a := range set {
		t.Run(a.name, func(t *testing.T) {
			t.Parallel()

			for _, n := range []int{-12, -4, -1, 1, 2, 3, 4, 12} {
				got, err := a.v.Div(n)
				want := a.c.scale(rat(1, int64(n)))
				binaryResult(t, a.name+" / n", want, got, err)
				require.NoError(t, err, "%s / %d: no error", a.name, n)
				assertMust(t, got, err, func() cents.Cents { return a.v.MustDiv(n) }, "MustDiv: %s / %d", a.name, n)

				byScale, errScale := a.v.Scale(1, n)
				require.NoError(t, errScale)
				assert.True(t, got == byScale, "Div(%d) = Scale(1, %d)", n, n)

				assertCents(t, a.v, noErr(got.Mul(n))(t, "(%s / %d) * %d", a.name, n, n), "(%s / %d) * %d", a.name, n, n)
			}

			op := checked{t}
			assertCents(t, a.v, op.div(a.v, 1), "a / 1")
			assertCents(t, a.v.Neg(), op.div(a.v, -1), "a / -1")
		})
	}
}

// Scale and ScaleRat: c*num/den in one operation; Mul(k) = Scale(k, 1),
// Div(n) = Scale(1, n); ScaleRat(nil) is 0; ScaleRat = Scale for big.NewRat.
func TestScale(t *testing.T) {
	t.Parallel()

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		op := checked{t}
		assertComp(t, ratC(700, 1), op.scale(cents.Millioctave(), 1750, 3), "Millioctave * 1750/3")

		milli := noErr(cents.Step(1, 1000))(t)
		got, err := milli.Scale(7500001, 1)
		assertResult(t, comp{}, cents.ErrOutOfRange, got, err, "E(1,1000) * 7500001")
		assertMust(t, got, err, func() cents.Cents { return milli.MustScale(7500001, 1) }, "MustScale")

		top := noErr(cents.FromInt(9000000))(t)
		assertCents(t, top, op.scale(top, maxInt, maxInt), "no intermediate overflow")
		assertCents(t, top, op.scale(top, minInt, minInt), "no intermediate overflow")
		assertCents(t, top.Neg(), op.scale(top, maxInt, -maxInt), "sign of den")
		assertCents(t, top, op.scale(top, 2, 2), "2/2")

		got, err = cents.Cent().Scale(maxInt, 1)
		assertResult(t, comp{}, cents.ErrOutOfRange, got, err, "<1> * MaxInt/1")
		assertMust(t, got, err, func() cents.Cents { return cents.Cent().MustScale(maxInt, 1) }, "MustScale")

		got, err = cents.Zero().Scale(1, 0)
		assertResult(t, comp{}, cents.ErrInvalidArgument, got, err, "0.Scale(1, 0)")
		assertMust(t, got, err, func() cents.Cents { return cents.Zero().MustScale(1, 0) }, "MustScale")

		got, err = top.Scale(maxInt, 0)
		assertResult(t, comp{}, cents.ErrInvalidArgument, got, err, "den 0 before range")
		assertMust(t, got, err, func() cents.Cents { return top.MustScale(maxInt, 0) }, "MustScale")
	})

	t.Run("ScaleRat", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			assert.True(t, noErr(a.v.ScaleRat(nil))(t, "%s.ScaleRat(nil)", a.name) == cents.Cents{}, "%s.ScaleRat(nil)", a.name)
			assert.True(t, a.v.MustScaleRat(nil) == cents.Cents{}, "%s.MustScaleRat(nil)", a.name)

			for _, num := range []int{-4, -1, 0, 1, 3} {
				for _, den := range []int{1, 2, 3, -5, 12} {
					byScale, errScale := a.v.Scale(num, den)
					byRat, errRat := a.v.ScaleRat(big.NewRat(int64(num), int64(den)))
					assert.True(t, byScale == byRat, "%s * %d/%d", a.name, num, den)
					assert.Equal(t, errName(errScale), errName(errRat), "%s * %d/%d", a.name, num, den)
					binaryResult(t, a.name+" scaled", a.c.scale(rat(int64(num), int64(den))), byRat, errRat)
					assertMust(t, byScale, errScale, func() cents.Cents { return a.v.MustScale(num, den) }, "MustScale: %s * %d/%d", a.name, num, den)
					assertMust(t, byRat, errRat, func() cents.Cents { return a.v.MustScaleRat(big.NewRat(int64(num), int64(den))) },
						"MustScaleRat: %s * %d/%d", a.name, num, den)
				}
			}
		}
	})
}

// K-9: the relations P-15 between the named values are exact equalities.
func TestNamedRelations(t *testing.T) {
	t.Parallel()

	op := checked{t}
	synt, pyth := cents.SyntonicComma(), cents.PythagoreanComma()
	r32 := noErr(cents.FromRatio(3, 2))(t)

	cases := []struct {
		name      string
		want, got cents.Cents
	}{
		{"schisma = PYTH - SYNT", cents.Schisma(), op.sub(pyth, synt)},
		{"diaschisma = SYNT*2 - PYTH", cents.Diaschisma(), op.sub(op.mul(synt, 2), pyth)},
		{"lesser diesis = SYNT*3 - PYTH", cents.LesserDiesis(), op.sub(op.mul(synt, 3), pyth)},
		{"greater diesis = SYNT*4 - PYTH", cents.GreaterDiesis(), op.sub(op.mul(synt, 4), pyth)},
		{"PYTH = apotome - limma", pyth, op.sub(cents.Apotome(), cents.Limma())},
		{"PYTH = R(3:2)*12 - <8400>", pyth, op.sub(op.mul(r32, 12), noErr(cents.FromInt(8400))(t))},
		{"Ptolemy = R(10:9) - R(11:10)", cents.PtolemysComma(), op.sub(noErr(cents.FromRatio(10, 9))(t), noErr(cents.FromRatio(11, 10))(t))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assertCents(t, tc.want, tc.got)
		})
	}
}

// K-17, G-4: c = 1200*q + rest, q = floor(c/1200), 0 <= rest < 1200.
func TestSplitOctaves(t *testing.T) {
	t.Parallel()

	r31, r32 := noErr(cents.FromRatio(3, 1))(t), noErr(cents.FromRatio(3, 2))(t)
	octaveMinusSchisma := noErr(cents.Octave().Sub(cents.Schisma()))(t)

	cases := []struct {
		name    string
		c       cents.Cents
		octaves int
		rest    cents.Cents
	}{
		{"<1900>", noErr(cents.FromInt(1900))(t), 1, noErr(cents.FromInt(700))(t)},
		{"<-100>", noErr(cents.FromInt(-100))(t), -1, noErr(cents.FromInt(1100))(t)},
		{"<1200>", cents.Octave(), 1, cents.Zero()},
		{"0", cents.Zero(), 0, cents.Zero()},
		{"<-1200>", noErr(cents.FromInt(-1200))(t), -1, cents.Zero()},
		{"R(3:1)", r31, 1, r32},
		{"R(2:3)", noErr(cents.FromRatio(2, 3))(t), -1, noErr(cents.FromRatio(4, 3))(t)},
		{"<1200> - schisma", octaveMinusSchisma, 0, octaveMinusSchisma},
		{"<9000000>", noErr(cents.FromInt(9000000))(t), 7500, cents.Zero()},
		{"<-9000000>", noErr(cents.FromInt(-9000000))(t), -7500, cents.Zero()},
		{"<-1/1000000>", noErr(cents.FromFraction(-1, 1000000))(t), -1, noErr(cents.FromFraction(1199999999, 1000000))(t)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			octaves, rest := tc.c.SplitOctaves()
			assert.Equal(t, tc.octaves, octaves)
			assertCents(t, tc.rest, rest)
		})
	}

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			q := floorScaled(a.c, rat(1, 1200))
			require.True(t, q.IsInt64(), a.name)

			octaves, rest := a.v.SplitOctaves()
			assert.Equal(t, int(q.Int64()), octaves, a.name)
			assertComp(t, a.c.sub(ratC(1200*q.Int64(), 1)), rest, "rest of %s", a.name)

			assert.GreaterOrEqual(t, rest.Compare(cents.Zero()), 0, "rest of %s >= 0", a.name)
			assert.Equal(t, -1, rest.Compare(cents.Octave()), "rest of %s < 1200", a.name)
			assertCents(t, a.v, noErr(noErr(cents.Octave().Mul(octaves))(t).Add(rest))(t), "1200*q + rest = %s", a.name)
		}
	})
}

// K-20, G-7: rounding to a multiple of 1/d, half away from zero.
func TestRoundTo(t *testing.T) {
	t.Parallel()

	r32 := noErr(cents.FromRatio(3, 2))(t)
	e119 := noErr(cents.Step(1, 19))(t)

	cases := []struct {
		name string
		c    cents.Cents
		d    int
		want comp
	}{
		{"R(3:2) to 1/1000", r32, 1000, ratC(701955, 1000)},
		{"R(3:2) to 1", r32, 1, ratC(702, 1)},
		{"E(1,19) to 1/1000", e119, 1000, ratC(63158, 1000)},
		{"1/2", noErr(cents.FromFraction(1, 2))(t), 1, ratC(1, 1)},
		{"-1/2", noErr(cents.FromFraction(-1, 2))(t), 1, ratC(-1, 1)},
		{"5/2", noErr(cents.FromFraction(5, 2))(t), 1, ratC(3, 1)},
		{"-5/2", noErr(cents.FromFraction(-5, 2))(t), 1, ratC(-3, 1)},
		{"<9000000>", noErr(cents.FromInt(9000000))(t), 1, ratC(9000000, 1)},
		{"<-9000000> to 1/MaxInt", noErr(cents.FromInt(-9000000))(t), maxInt, ratC(-9000000, 1)},
		{"1/8 to 1/4", noErr(cents.FromFraction(1, 8))(t), 4, ratC(1, 4)},
		{"-1/8 to 1/4", noErr(cents.FromFraction(-1, 8))(t), 4, ratC(-1, 4)},
		{"-1/3 to 1", noErr(cents.FromFraction(-1, 3))(t), 1, zeroComp()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.c.RoundTo(tc.d)
			require.NoError(t, err)
			assertComp(t, tc.want, got)
			assertMust(t, got, err, func() cents.Cents { return tc.c.MustRoundTo(tc.d) }, "MustRoundTo")
		})
	}

	t.Run("invalid d", func(t *testing.T) {
		t.Parallel()

		for _, d := range []int{0, -1, minInt} {
			got, err := r32.RoundTo(d)
			assertResult(t, comp{}, cents.ErrInvalidArgument, got, err, "d = %d", d)
			assertMust(t, got, err, func() cents.Cents { return r32.MustRoundTo(d) }, "MustRoundTo: d = %d", d)
		}
	})

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			for _, d := range []int{1, 2, 3, 7, 10, 1000, 1024, 1000000} {
				q := nearestScaled(a.c, rat(int64(d), 1))
				want := bigRatC(new(big.Rat).SetFrac(q, big.NewInt(int64(d))))

				got, err := a.v.RoundTo(d)
				require.NoError(t, err, "%s to 1/%d", a.name, d)
				assertComp(t, want, got, "%s to 1/%d", a.name, d)
				assertMust(t, got, err, func() cents.Cents { return a.v.MustRoundTo(d) }, "MustRoundTo: %s to 1/%d", a.name, d)

				num, den, ok := got.Fraction()
				require.True(t, ok, "%s to 1/%d has a fraction", a.name, d)
				assert.Zero(t, d%den, "%s to 1/%d: denominator %d divides d", a.name, d, den)
				assertComp(t, ratC(int64(num), int64(den)), got)
			}
		}
	})
}
