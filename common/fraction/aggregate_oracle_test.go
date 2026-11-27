package fraction_test

import (
	"math/big"
	"slices"
	"testing"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/assert"
)

// Contract oracle of task 26.7 (architecture, "Library API contract"): GCD and
// LCM of two fractions, the common denominator of a set, and the exact sum and
// product of a set. Every expectation is computed on math/big values from the
// reduced forms p/q and r/s of the operands; the oracle never calls the
// operation it predicts.

// agReduced returns the reduced canonical pair (p, q) of a valid fraction,
// q > 0 and NOD(|p|, q) = 1, as fresh values.
func agReduced(f fraction.Fraction) (p, q *big.Int) {
	r := wRat(f)
	return new(big.Int).Set(r.Num()), new(big.Int).Set(r.Denom())
}

// agLCMInt returns LCM(|a|, |b|) for nonzero a, b as a new value.
func agLCMInt(a, b *big.Int) *big.Int {
	g := wGCD(a, b)
	l := new(big.Int).Quo(new(big.Int).Abs(a), g)
	return l.Mul(l, new(big.Int).Abs(b))
}

// agAllValid reports whether every element of the set is valid.
func agAllValid(fs []fraction.Fraction) bool {
	return !slices.ContainsFunc(fs, func(f fraction.Fraction) bool { return !f.IsValid() })
}

// agWantGCD returns NOD(|p|, |r|) / LCM(q, s) in reduced form, 0/1 when both
// operands are zero; ErrInvalid comes first (step 1), ErrOutOfRange (step 4)
// when the denominator LCM(q, s) exceeds MaxComponent.
func agWantGCD(x, y fraction.Fraction) wideWant {
	if !x.IsValid() || !y.IsValid() {
		return wErr(fraction.ErrInvalid)
	}
	p, q := agReduced(x)
	r, s := agReduced(y)
	num := wGCD(p, r)
	if num.Sign() == 0 {
		return wRange(big.NewInt(0), big.NewInt(1))
	}
	return wRange(num, agLCMInt(q, s))
}

// agWantLCM returns LCM(|p|, |r|) / NOD(q, s) in reduced form, 0/1 when an
// operand is zero; ErrOutOfRange when the numerator exceeds MaxComponent.
func agWantLCM(x, y fraction.Fraction) wideWant {
	if !x.IsValid() || !y.IsValid() {
		return wErr(fraction.ErrInvalid)
	}
	p, q := agReduced(x)
	r, s := agReduced(y)
	if p.Sign() == 0 || r.Sign() == 0 {
		return wRange(big.NewInt(0), big.NewInt(1))
	}
	return wRange(agLCMInt(p, r), wGCD(q, s))
}

// agWantCommonDenominatorBig returns the LCM of the reduced denominators, 1 for
// the empty set, or ErrInvalid when any element is invalid. There is no limit.
func agWantCommonDenominatorBig(fs []fraction.Fraction) (*big.Int, error) {
	if !agAllValid(fs) {
		return nil, fraction.ErrInvalid
	}
	d := big.NewInt(1)
	for _, f := range fs {
		_, q := agReduced(f)
		d = agLCMInt(d, q)
	}
	return d, nil
}

// agWantCommonDenominator is the narrow form: ErrNotRepresentable (step 3) when
// the common denominator exceeds math.MaxInt, after the validity check.
func agWantCommonDenominator(fs []fraction.Fraction) (int, error) {
	d, err := agWantCommonDenominatorBig(fs)
	if err != nil {
		return 0, err
	}
	if !wFitsInt(d) {
		return 0, fraction.ErrNotRepresentable
	}
	return int(d.Int64()), nil
}

// agExactSum returns the exact sum of a valid set; 0 for the empty set.
func agExactSum(fs []fraction.Fraction) *big.Rat {
	acc := new(big.Rat)
	for _, f := range fs {
		acc.Add(acc, wRat(f))
	}
	return acc
}

// agExactProduct returns the exact product of a valid set; 1 for the empty set.
func agExactProduct(fs []fraction.Fraction) *big.Rat {
	acc := big.NewRat(1, 1)
	for _, f := range fs {
		acc.Mul(acc, wRat(f))
	}
	return acc
}

// agWantSum returns the reduced canonical exact sum; only the result is range-checked.
func agWantSum(fs []fraction.Fraction) wideWant {
	if !agAllValid(fs) {
		return wErr(fraction.ErrInvalid)
	}
	return wFromRat(agExactSum(fs))
}

// agWantProduct returns the reduced canonical exact product; only the result is range-checked.
func agWantProduct(fs []fraction.Fraction) wideWant {
	if !agAllValid(fs) {
		return wErr(fraction.ErrInvalid)
	}
	return wFromRat(agExactProduct(fs))
}

// agRatInRange reports whether both components of the reduced value r are in range.
func agRatInRange(r *big.Rat) bool {
	return wInRange(r.Num()) && wInRange(r.Denom())
}

// agPrefixesInRange reports whether every reduced prefix sum (product) of the
// set, by the order of the elements, is in range.
func agPrefixesInRange(fs []fraction.Fraction, product bool) bool {
	for i := 1; i <= len(fs); i++ {
		var r *big.Rat
		if product {
			r = agExactProduct(fs[:i])
		} else {
			r = agExactSum(fs[:i])
		}
		if !agRatInRange(r) {
			return false
		}
	}
	return true
}

// agWantPair is an expected reduced result num/den.
func agWantPair(num, den *big.Int) wideWant {
	return wideWant{num: num, den: den}
}

// agWantInts is an expected reduced result num/den of int64 components.
func agWantInts(num, den int64) wideWant {
	return agWantPair(big.NewInt(num), big.NewInt(den))
}

// agSameWant checks that the oracle agrees with an explicit expectation of a table.
func agSameWant(t *testing.T, label string, oracle, want wideWant) {
	t.Helper()
	if want.err != nil {
		assert.Truef(t, oracle.err == want.err, "%s: oracle error %v, table %v", label, oracle.err, want.err)
		return
	}
	if assert.NoErrorf(t, oracle.err, "%s: oracle", label) {
		assert.Truef(t, oracle.num.Cmp(want.num) == 0 && oracle.den.Cmp(want.den) == 0,
			"%s: oracle %s/%s, table %s/%s", label, oracle.num, oracle.den, want.num, want.den)
	}
}

// agCheckCommonDenominator checks CommonDenominator, CommonDenominatorBig and
// their Must forms against the oracle: the value, the error, the zero results
// on error (0 and nil), a fresh *big.Int per call and an unchanged input set.
func agCheckCommonDenominator(t *testing.T, label string, fs []fraction.Fraction) {
	t.Helper()
	before := slices.Clone(fs)
	wantBig, wantBigErr := agWantCommonDenominatorBig(fs)
	wantNarrow, wantNarrowErr := agWantCommonDenominator(fs)

	got, err := fraction.CommonDenominator(fs...)
	if wantNarrowErr != nil {
		wCheckErr(t, label+": CommonDenominator", err, wantNarrowErr)
		assert.Zerof(t, got, "%s: CommonDenominator result on error must be 0", label)
	} else if assert.NoErrorf(t, err, "%s: CommonDenominator", label) {
		assert.Equalf(t, wantNarrow, got, "%s: CommonDenominator", label)
	}
	var mustGot int
	wCheckMustErr(t, label+": MustCommonDenominator", func() { mustGot = fraction.MustCommonDenominator(fs...) }, wantNarrowErr)
	if wantNarrowErr == nil {
		assert.Equalf(t, wantNarrow, mustGot, "%s: MustCommonDenominator", label)
	}

	gotBig, err := fraction.CommonDenominatorBig(fs...)
	if wantBigErr != nil {
		wCheckErr(t, label+": CommonDenominatorBig", err, wantBigErr)
		assert.Nilf(t, gotBig, "%s: CommonDenominatorBig result on error must be nil", label)
	} else if assert.NoErrorf(t, err, "%s: CommonDenominatorBig", label) {
		wCheckBigInt(t, label+": CommonDenominatorBig", gotBig, wantBig)
		// The result is a new value: changing it affects neither the set nor a later call.
		if gotBig != nil {
			gotBig.SetInt64(-7)
			again, err := fraction.CommonDenominatorBig(fs...)
			if assert.NoErrorf(t, err, "%s: second CommonDenominatorBig", label) {
				assert.Falsef(t, again == gotBig, "%s: every call must return a new *big.Int", label)
				wCheckBigInt(t, label+": second CommonDenominatorBig", again, wantBig)
			}
		}
	}
	var mustBig *big.Int
	wCheckMustErr(t, label+": MustCommonDenominatorBig", func() { mustBig = fraction.MustCommonDenominatorBig(fs...) }, wantBigErr)
	if wantBigErr == nil {
		wCheckBigInt(t, label+": MustCommonDenominatorBig", mustBig, wantBig)
	}
	assert.Truef(t, slices.Equal(before, fs), "%s: the input set must not change", label)
}

// agCheckSumProduct checks Sum, Product and their Must forms against the
// oracle, and that the input set does not change.
func agCheckSumProduct(t *testing.T, label string, fs []fraction.Fraction) {
	t.Helper()
	before := slices.Clone(fs)
	sum := agWantSum(fs)
	got, err := fraction.Sum(fs...)
	wCheckFrac(t, label+": Sum", got, err, sum)
	wCheckMust(t, label+": MustSum", func() fraction.Fraction { return fraction.MustSum(fs...) }, sum)

	product := agWantProduct(fs)
	got, err = fraction.Product(fs...)
	wCheckFrac(t, label+": Product", got, err, product)
	wCheckMust(t, label+": MustProduct", func() fraction.Fraction { return fraction.MustProduct(fs...) }, product)
	assert.Truef(t, slices.Equal(before, fs), "%s: the input set must not change", label)
}

// agCheckGCDLCM checks GCD, LCM and their Must forms of x and y in both orders
// against the oracle, and that the operands do not change.
func agCheckGCDLCM(t *testing.T, label string, x, y fraction.Fraction) {
	t.Helper()
	x0, y0 := x, y
	for _, o := range []struct {
		name string
		a, b fraction.Fraction
	}{{"x,y", x, y}, {"y,x", y, x}} {
		l := label + " (" + o.name + ")"
		gcd := agWantGCD(o.a, o.b)
		got, err := o.a.GCD(o.b)
		wCheckFrac(t, l+": GCD", got, err, gcd)
		wCheckMust(t, l+": MustGCD", func() fraction.Fraction { return o.a.MustGCD(o.b) }, gcd)

		lcm := agWantLCM(o.a, o.b)
		got, err = o.a.LCM(o.b)
		wCheckFrac(t, l+": LCM", got, err, lcm)
		wCheckMust(t, l+": MustLCM", func() fraction.Fraction { return o.a.MustLCM(o.b) }, lcm)
	}
	assert.Truef(t, x == x0 && y == y0, "%s: the operands must not change", label)
}

// agCheckBigRatTo checks BigRatTo of f into z against BigRat: the same pointer
// (a new value for nil z), the value, the reduced canonical form (positive
// denominator, 0/1 for zero), a usable big.Rat afterwards, and for an invalid
// f the result (z, false) with z unchanged.
func agCheckBigRatTo(t *testing.T, label string, f fraction.Fraction, z *big.Rat) {
	t.Helper()
	var before *big.Rat
	if z != nil {
		before = new(big.Rat).Set(z)
	}
	got, ok := f.BigRatTo(z)
	if !f.IsValid() {
		assert.Falsef(t, ok, "%s: invalid fraction must report false", label)
		assert.Truef(t, got == z, "%s: invalid fraction must return z", label)
		if z != nil {
			assert.Equalf(t, before.String(), z.String(), "%s: z must not change", label)
		}
		return
	}
	if !assert.Truef(t, ok, "%s: valid fraction must report true", label) || !assert.NotNilf(t, got, "%s: result", label) {
		return
	}
	if z != nil {
		assert.Truef(t, got == z, "%s: the result must be z", label)
	}
	want, wantOK := f.BigRat()
	if !assert.Truef(t, wantOK, "%s: BigRat", label) {
		return
	}
	p, q := agReduced(f)
	assert.Zerof(t, got.Cmp(want), "%s: value %s, want %s", label, got, want)
	assert.Equalf(t, want.String(), got.String(), "%s: text", label)
	assert.Truef(t, got.Num().Cmp(p) == 0 && got.Denom().Cmp(q) == 0,
		"%s: components %s/%s, want reduced %s/%s", label, got.Num(), got.Denom(), p, q)
	assert.Equalf(t, q.Cmp(big.NewInt(1)) == 0, got.IsInt(), "%s: IsInt", label)
	// The written value takes part in further big.Rat arithmetic as a normalized value.
	third := big.NewRat(1, 3)
	sum := new(big.Rat).Add(got, third)
	assert.Zerof(t, sum.Cmp(new(big.Rat).Add(want, third)), "%s: arithmetic on the result", label)
	sq := new(big.Rat).Mul(got, got)
	assert.Equalf(t, new(big.Rat).Mul(want, want).String(), sq.String(), "%s: square of the result", label)
}
