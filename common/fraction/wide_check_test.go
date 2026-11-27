package fraction_test

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"

	"github.com/stretchr/testify/assert"
)

// This file runs every operation of the wide-value test list on given operands
// and compares the observable result with the contract oracle: the original
// pair of the result, the value, the error and its sentinel class, the text,
// and the panic value of the Must form.

// wCheckErr checks that err wraps want and no other sentinel of the package.
func wCheckErr(t *testing.T, label string, err, want error) bool {
	t.Helper()
	return assert.Truef(t, wSentinelOnly(err, want), "%s: error %v, want only %v", label, err, want)
}

// wCheckFrac checks a (Fraction, error) result against the expectation.
func wCheckFrac(t *testing.T, label string, got fraction.Fraction, err error, want wideWant) {
	t.Helper()
	if want.err != nil {
		wCheckErr(t, label, err, want.err)
		assert.Equalf(t, fraction.Fraction{}, got, "%s: result on error must be Fraction{}", label)
		return
	}
	if !assert.NoErrorf(t, err, "%s", label) {
		return
	}
	n, d, ok := got.BigPair()
	if !assert.Truef(t, ok, "%s: result must be valid", label) {
		return
	}
	assert.Truef(t, n.Cmp(want.num) == 0 && d.Cmp(want.den) == 0,
		"%s: got %s/%s, want %s/%s", label, n, d, want.num, want.den)
}

// wCheckMust checks the Must form: the same result, or a panic with the same error.
func wCheckMust(t *testing.T, label string, fn func() fraction.Fraction, want wideWant) {
	t.Helper()
	var got fraction.Fraction
	rec := wCatch(func() { got = fn() })
	if want.err != nil {
		err, ok := rec.(error)
		assert.Truef(t, ok && wSentinelOnly(err, want.err), "%s: Must must panic with %v, got %v", label, want.err, rec)
		return
	}
	if !assert.Nilf(t, rec, "%s: Must must not panic", label) {
		return
	}
	wCheckFrac(t, label, got, nil, want)
}

// wCheckMustErr checks that a Must form panics with want, or does not panic when want is nil.
func wCheckMustErr(t *testing.T, label string, fn func(), want error) {
	t.Helper()
	rec := wCatch(fn)
	if want == nil {
		assert.Nilf(t, rec, "%s: Must must not panic", label)
		return
	}
	err, ok := rec.(error)
	assert.Truef(t, ok && wSentinelOnly(err, want), "%s: Must must panic with %v, got %v", label, want, rec)
}

// wCheckBigInt compares a *big.Int result; want == nil means a nil result.
func wCheckBigInt(t *testing.T, label string, got, want *big.Int) {
	t.Helper()
	if want == nil {
		assert.Nilf(t, got, "%s: result must be nil", label)
		return
	}
	if assert.NotNilf(t, got, "%s: result must not be nil", label) {
		assert.Truef(t, got.Cmp(want) == 0, "%s: got %s, want %s", label, got, want)
	}
}

// wCheckDivResult checks the (quotient, remainder, error) results of the narrow forms.
func wCheckDivResult(t *testing.T, label string, q int, rem fraction.Fraction, err error, want wideDivResult) {
	t.Helper()
	if want.rem.err != nil {
		assert.Equalf(t, 0, q, "%s: quotient on error", label)
	} else {
		assert.Truef(t, want.q.IsInt64() && int64(q) == want.q.Int64(), "%s: quotient %d, want %s", label, q, want.q)
	}
	wCheckFrac(t, label, rem, err, want.rem)
}

// wCheckBigDivResult checks the (quotient, remainder, error) results of the *Big forms.
func wCheckBigDivResult(t *testing.T, label string, q *big.Int, rem fraction.Fraction, err error, want wideDivResult) {
	t.Helper()
	wCheckBigInt(t, label, q, want.q)
	wCheckFrac(t, label, rem, err, want.rem)
}

// wIntDens returns int denominators for the mode-7, WithDenominator and
// RoundToDenominator calls: fixed boundary values, invalid values, and
// multiples of the reduced denominator q of the relevant value when they fit.
func wIntDens(q *big.Int) []int {
	dens := []int{1, 2, 6, 1 << 62, math.MaxInt, 0, -3, math.MinInt}
	if q == nil {
		return dens
	}
	for _, k := range []int64{1, 2, 3, 1 << 20} {
		m := new(big.Int).Mul(q, big.NewInt(k))
		if m.IsInt64() {
			dens = append(dens, int(m.Int64()))
		}
	}
	return dens
}

// wIntArgs returns int operands for AddInt, SubInt, MulInt and DivInt.
var wIntArgs = []int{0, 1, -1, 3, 1 << 40, math.MaxInt, math.MinInt}

type wideBinaryMethods struct {
	plain    func(x, y fraction.Fraction) (fraction.Fraction, error)
	must     func(x, y fraction.Fraction) fraction.Fraction
	with     func(x, y fraction.Fraction, m fraction.DenomMode) (fraction.Fraction, error)
	mustWith func(x, y fraction.Fraction, m fraction.DenomMode) fraction.Fraction
	inDen    func(x, y fraction.Fraction, den int) (fraction.Fraction, error)
	mustDen  func(x, y fraction.Fraction, den int) fraction.Fraction
	intOp    func(x fraction.Fraction, n int) (fraction.Fraction, error)
	mustInt  func(x fraction.Fraction, n int) fraction.Fraction
}

var wideBinary = map[wideArith]wideBinaryMethods{
	wideAdd: {
		fraction.Fraction.Add, fraction.Fraction.MustAdd,
		fraction.Fraction.AddWith, fraction.Fraction.MustAddWith,
		fraction.Fraction.AddInDenominator, fraction.Fraction.MustAddInDenominator,
		fraction.Fraction.AddInt, fraction.Fraction.MustAddInt,
	},
	wideSub: {
		fraction.Fraction.Sub, fraction.Fraction.MustSub,
		fraction.Fraction.SubWith, fraction.Fraction.MustSubWith,
		fraction.Fraction.SubInDenominator, fraction.Fraction.MustSubInDenominator,
		fraction.Fraction.SubInt, fraction.Fraction.MustSubInt,
	},
	wideMul: {
		fraction.Fraction.Mul, fraction.Fraction.MustMul,
		fraction.Fraction.MulWith, fraction.Fraction.MustMulWith,
		fraction.Fraction.MulInDenominator, fraction.Fraction.MustMulInDenominator,
		fraction.Fraction.MulInt, fraction.Fraction.MustMulInt,
	},
	wideDiv: {
		fraction.Fraction.Div, fraction.Fraction.MustDiv,
		fraction.Fraction.DivWith, fraction.Fraction.MustDivWith,
		fraction.Fraction.DivInDenominator, fraction.Fraction.MustDivInDenominator,
		fraction.Fraction.DivInt, fraction.Fraction.MustDivInt,
	},
}

var wideArithOps = []wideArith{wideAdd, wideSub, wideMul, wideDiv}

// checkWideBinary runs every binary operation of the list on (x, y).
func checkWideBinary(t *testing.T, x, y fraction.Fraction) {
	t.Helper()

	for _, op := range wideArithOps {
		m := wideBinary[op]
		want := wantArith(op, x, y, fraction.DenomReduced)
		got, err := m.plain(x, y)
		wCheckFrac(t, op.String(), got, err, want)
		wCheckMust(t, "Must"+op.String(), func() fraction.Fraction { return m.must(x, y) }, want)

		for _, mode := range wideModes {
			label := fmt.Sprintf("%sWith(mode %d)", op, mode)
			want := wantArith(op, x, y, mode)
			got, err := m.with(x, y, mode)
			wCheckFrac(t, label, got, err, want)
			wCheckMust(t, "Must"+label, func() fraction.Fraction { return m.mustWith(x, y, mode) }, want)
		}

		var q *big.Int
		if v, ok := wValue(op, x, y); ok {
			q = v.Denom()
		}
		for _, den := range wIntDens(q) {
			label := fmt.Sprintf("%sInDenominator(%d)", op, den)
			want := wantInDenominator(op, x, y, den)
			got, err := m.inDen(x, y, den)
			wCheckFrac(t, label, got, err, want)
			wCheckMust(t, "Must"+label, func() fraction.Fraction { return m.mustDen(x, y, den) }, want)
		}
	}

	wantCmp := wRat(x).Cmp(wRat(y))
	assert.Equalf(t, wantCmp, x.Compare(y), "Compare")
	assert.Equalf(t, -wantCmp, y.Compare(x), "Compare reversed")
	assert.Equalf(t, wantCmp == 0, x.Equal(y), "Equal")

	for _, floor := range []bool{false, true} {
		name := "QuoRem"
		narrow, wide := fraction.Fraction.QuoRem, fraction.Fraction.QuoRemBig
		mustNarrow := fraction.Fraction.MustQuoRem
		if floor {
			name = "DivModFloor"
			narrow, wide = fraction.Fraction.DivModFloor, fraction.Fraction.DivModFloorBig
			mustNarrow = fraction.Fraction.MustDivModFloor
		}
		want := wantQuoRem(x, y, floor, true)
		q, rem, err := narrow(x, y)
		wCheckDivResult(t, name, q, rem, err, want)
		wCheckMustErr(t, "Must"+name, func() { mustNarrow(x, y) }, want.rem.err)

		wantBig := wantQuoRem(x, y, floor, false)
		qb, remb, err := wide(x, y)
		wCheckBigDivResult(t, name+"Big", qb, remb, err, wantBig)
	}

	wantMul, wantMulErr := wantIsMultipleOf(x, y)
	gotMul, err := x.IsMultipleOf(y)
	if wantMulErr != nil {
		wCheckErr(t, "IsMultipleOf", err, wantMulErr)
		assert.False(t, gotMul, "IsMultipleOf on error")
	} else if assert.NoError(t, err, "IsMultipleOf") {
		assert.Equal(t, wantMul, gotMul, "IsMultipleOf")
	}
	wCheckMustErr(t, "MustIsMultipleOf", func() { x.MustIsMultipleOf(y) }, wantMulErr)

	starts := []fraction.Fraction{fraction.Zero(), y, x}
	for i, start := range starts {
		wantK, wantOK, wantErr := wantInProgression(x, start, y)
		gotK, gotOK, err := x.InProgression(start, y)
		label := fmt.Sprintf("InProgression(start %d)", i)
		if wantErr != nil {
			wCheckErr(t, label, err, wantErr)
		} else {
			assert.NoErrorf(t, err, "%s", label)
		}
		assert.Equalf(t, wantK, gotK, "%s: steps", label)
		assert.Equalf(t, wantOK, gotOK, "%s: ok", label)
	}

	// A point of the progression start + k·step with a long k, when it is in range.
	if ry := wRat(y); ry.Sign() > 0 {
		k := big.NewRat(1<<40+3, 1)
		point := new(big.Rat).Add(wRat(x), new(big.Rat).Mul(k, ry))
		if p, err := fraction.FromBigRat(point); err == nil {
			gotK, gotOK, err := p.InProgression(x, y)
			assert.NoError(t, err, "InProgression(point)")
			assert.True(t, gotOK, "InProgression(point): ok")
			assert.Equal(t, 1<<40+3, gotK, "InProgression(point): steps")
		}
	}
}

// checkWideUnary runs every operation of the list over a single fraction x.
func checkWideUnary(t *testing.T, x fraction.Fraction) {
	t.Helper()

	r := wRat(x)
	nRed, dRed := r.Num(), r.Denom()
	origN, origD, _ := x.BigPair()

	// Reduced forms and readers.
	wCheckFrac(t, "Reduce", x.Reduce(), nil, wideWant{num: nRed, den: dRed})
	wantReduced := origD.Sign() > 0 && wGCD(origN, origD).Cmp(big.NewInt(1)) == 0
	assert.Equal(t, wantReduced, x.IsReduced(), "IsReduced")
	assert.Equal(t, new(big.Int).Rem(origN, origD).Sign() == 0, x.IsInteger(), "IsInteger")

	rn, ok := x.ReducedNumerator()
	assert.Equal(t, nRed.IsInt64(), ok, "ReducedNumerator ok")
	if ok {
		assert.Equal(t, nRed.Int64(), int64(rn), "ReducedNumerator")
	}
	rd, ok := x.ReducedDenominator()
	assert.Equal(t, dRed.IsInt64(), ok, "ReducedDenominator ok")
	if ok {
		assert.Equal(t, dRed.Int64(), int64(rd), "ReducedDenominator")
	}
	pn, pd, ok := x.ReducedPair()
	assert.Equal(t, nRed.IsInt64() && dRed.IsInt64(), ok, "ReducedPair ok")
	if ok {
		assert.Equal(t, [2]int64{nRed.Int64(), dRed.Int64()}, [2]int64{int64(pn), int64(pd)}, "ReducedPair")
	} else {
		assert.Equal(t, [2]int{0, 0}, [2]int{pn, pd}, "ReducedPair on failure")
	}
	bn, ok := x.ReducedBigNumerator()
	assert.True(t, ok, "ReducedBigNumerator ok")
	wCheckBigInt(t, "ReducedBigNumerator", bn, nRed)
	bd, ok := x.ReducedBigDenominator()
	assert.True(t, ok, "ReducedBigDenominator ok")
	wCheckBigInt(t, "ReducedBigDenominator", bd, dRed)
	bpn, bpd, ok := x.ReducedBigPair()
	assert.True(t, ok, "ReducedBigPair ok")
	wCheckBigInt(t, "ReducedBigPair num", bpn, nRed)
	wCheckBigInt(t, "ReducedBigPair den", bpd, dRed)

	isInt := r.IsInt()
	iv, ok := x.Int()
	assert.Equal(t, isInt && nRed.IsInt64(), ok, "Int ok")
	if ok {
		assert.Equal(t, nRed.Int64(), int64(iv), "Int")
	}
	bi, ok := x.BigInt()
	assert.Equal(t, isInt, ok, "BigInt ok")
	if isInt {
		wCheckBigInt(t, "BigInt", bi, nRed)
	} else {
		assert.Nil(t, bi, "BigInt on failure")
	}
	br, ok := x.BigRat()
	if assert.True(t, ok, "BigRat ok") {
		assert.Equal(t, 0, br.Cmp(r), "BigRat")
	}
	wantF, wantExact := r.Float64()
	gotF, gotExact := x.Float64()
	assert.Equal(t, wantF, gotF, "Float64")
	assert.Equal(t, wantExact, gotExact, "Float64 exact")

	// Text forms.
	assert.Equal(t, nRed.String()+"/"+dRed.String(), x.ReducedString(), "ReducedString")
	assert.Equal(t, wantCompactString(x), x.CompactString(), "CompactString")
	assert.Equal(t, wantMixedString(x), x.MixedString(), "MixedString")
	for _, digits := range []int{0, 1, 5, 40, fraction.MaxDigits, fraction.MaxDigits + 1, -1} {
		for _, mode := range wideRoundingModes {
			label := fmt.Sprintf("FormatDecimal(%d, %d)", digits, mode)
			want, wantErr := wantFormatDecimal(x, digits, mode)
			got, err := x.FormatDecimal(digits, mode)
			if wantErr != nil {
				wCheckErr(t, label, err, wantErr)
			} else {
				assert.NoErrorf(t, err, "%s", label)
			}
			assert.Equalf(t, want, got, "%s", label)
			wCheckMustErr(t, "Must"+label, func() { x.MustFormatDecimal(digits, mode) }, wantErr)
		}
		want, wantErr := wantFormatDecimal(x, digits, fraction.RoundHalfAwayFromZero)
		got, err := x.Decimal(digits)
		if wantErr != nil {
			wCheckErr(t, "Decimal", err, wantErr)
		} else {
			assert.NoError(t, err, "Decimal")
		}
		assert.Equal(t, want, got, "Decimal(%d)", digits)
	}

	// Whole part, periods and rounding.
	for _, floor := range []bool{false, true} {
		name := "Split"
		narrow, wide, mustNarrow := fraction.Fraction.Split, fraction.Fraction.SplitBig, fraction.Fraction.MustSplit
		if floor {
			name = "SplitFloor"
			narrow, wide, mustNarrow = fraction.Fraction.SplitFloor, fraction.Fraction.SplitFloorBig, fraction.Fraction.MustSplitFloor
		}
		want := wantSplit(x, floor, true)
		q, rem, err := narrow(x)
		wCheckDivResult(t, name, q, rem, err, want)
		wCheckMustErr(t, "Must"+name, func() { mustNarrow(x) }, want.rem.err)
		qb, remb, err := wide(x)
		wCheckBigDivResult(t, name+"Big", qb, remb, err, wantSplit(x, floor, false))
	}

	for _, mode := range wideRoundingModes {
		want, wantErr := wantRoundToBigInt(x, mode)
		gotBig, err := x.RoundToBigInt(mode)
		label := fmt.Sprintf("RoundToBigInt(%d)", mode)
		if wantErr != nil {
			wCheckErr(t, label, err, wantErr)
			assert.Nilf(t, gotBig, "%s on error", label)
		} else if assert.NoErrorf(t, err, "%s", label) {
			wCheckBigInt(t, label, gotBig, want)
		}

		label = fmt.Sprintf("RoundToInt(%d)", mode)
		if wantErr == nil && !wFitsInt(want) {
			wantErr = fraction.ErrNotRepresentable
		}
		got, err := x.RoundToInt(mode)
		if wantErr != nil {
			wCheckErr(t, label, err, wantErr)
			assert.Zerof(t, got, "%s on error", label)
		} else if assert.NoErrorf(t, err, "%s", label) {
			assert.Equalf(t, want.Int64(), int64(got), "%s", label)
		}
		wCheckMustErr(t, "Must"+label, func() { x.MustRoundToInt(mode) }, wantErr)

		for _, den := range []int{1, 3, 1 << 30, 1<<62 + 1, math.MaxInt, 0, -1} {
			label := fmt.Sprintf("RoundToDenominator(%d, %d)", den, mode)
			want := wantRoundToBigDenominator(x, big.NewInt(int64(den)), mode)
			got, err := x.RoundToDenominator(den, mode)
			wCheckFrac(t, label, got, err, want)
			wCheckMust(t, "Must"+label, func() fraction.Fraction { return x.MustRoundToDenominator(den, mode) }, want)
		}
		bigDens := []*big.Int{
			nil, big.NewInt(0), big.NewInt(-5), wOff(wPow2(64), 1), wPow2(200),
			fraction.MaxComponentBig(), wOff(fraction.MaxComponentBig(), 1),
		}
		for i, den := range bigDens {
			label := fmt.Sprintf("RoundToBigDenominator(#%d, %d)", i, mode)
			want := wantRoundToBigDenominator(x, den, mode)
			got, err := x.RoundToBigDenominator(den, mode)
			wCheckFrac(t, label, got, err, want)
		}
	}

	// Conversion to a denominator.
	for _, den := range wIntDens(dRed) {
		label := fmt.Sprintf("WithDenominator(%d)", den)
		want := wantWithBigDenominator(x, big.NewInt(int64(den)))
		got, err := x.WithDenominator(den)
		wCheckFrac(t, label, got, err, want)
		wCheckMust(t, "Must"+label, func() fraction.Fraction { return x.MustWithDenominator(den) }, want)
	}
	bigDens := []*big.Int{
		nil, big.NewInt(0), big.NewInt(-2),
		wMul(dRed, wOff(wPow2(64), 1)), wMul(dRed, wPow2(200)), wOff(dRed, 1),
		fraction.MaxComponentBig(), wOff(fraction.MaxComponentBig(), 1), wMul(dRed, wPow2(300)),
	}
	for i, den := range bigDens {
		label := fmt.Sprintf("WithBigDenominator(#%d)", i)
		want := wantWithBigDenominator(x, den)
		got, err := x.WithBigDenominator(den)
		wCheckFrac(t, label, got, err, want)
		wCheckMust(t, "Must"+label, func() fraction.Fraction { return x.MustWithBigDenominator(den) }, want)
	}

	// Scaling by an integer.
	cs := []int{0, 1, -1, 12, math.MaxInt, math.MinInt}
	if dRed.IsInt64() {
		q := dRed.Int64()
		for _, k := range []int64{1, -3, 1 << 20} {
			if c := new(big.Int).Mul(big.NewInt(q), big.NewInt(k)); c.IsInt64() {
				cs = append(cs, int(c.Int64()))
			}
		}
	}
	for _, c := range cs {
		wantP, wantOK, wantErr := wantScaleIntExact(x, c)
		gotP, gotOK, err := x.ScaleIntExact(c)
		label := fmt.Sprintf("ScaleIntExact(%d)", c)
		if wantErr != nil {
			wCheckErr(t, label, err, wantErr)
		} else {
			assert.NoErrorf(t, err, "%s", label)
		}
		assert.Equalf(t, wantP, gotP, "%s: product", label)
		assert.Equalf(t, wantOK, gotOK, "%s: exact", label)
	}

	// Powers.
	for _, exp := range []int{0, 1, -1, 2, -2, 3, 255, math.MaxInt, math.MinInt} {
		label := fmt.Sprintf("Pow(%d)", exp)
		want := wantPow(x, exp)
		got, err := x.Pow(exp)
		wCheckFrac(t, label, got, err, want)
		wCheckMust(t, "Must"+label, func() fraction.Fraction { return x.MustPow(exp) }, want)
	}

	// Integer operands and steps by one.
	for _, op := range wideArithOps {
		m := wideBinary[op]
		for _, k := range wIntArgs {
			label := fmt.Sprintf("%sInt(%d)", op, k)
			want := wantArith(op, x, fraction.FromInt(k), fraction.DenomReduced)
			got, err := m.intOp(x, k)
			wCheckFrac(t, label, got, err, want)
			wCheckMust(t, "Must"+label, func() fraction.Fraction { return m.mustInt(x, k) }, want)
		}
	}
	one := fraction.FromInt(1)
	wantInc := wantArith(wideAdd, x, one, fraction.DenomReduced)
	got, err := x.Inc()
	wCheckFrac(t, "Inc", got, err, wantInc)
	wCheckMust(t, "MustInc", x.MustInc, wantInc)
	wantDec := wantArith(wideSub, x, one, fraction.DenomReduced)
	got, err = x.Dec()
	wCheckFrac(t, "Dec", got, err, wantDec)
	wCheckMust(t, "MustDec", x.MustDec, wantDec)

	// Reducing constructors from the same components.
	got, err = fraction.FromBigIntReduced(origN, origD)
	wCheckFrac(t, "FromBigIntReduced", got, err, wantFromBigIntReduced(origN, origD))
	got, err = fraction.FromBigRat(r)
	wCheckFrac(t, "FromBigRat", got, err, wFromRat(r))
	if origN.IsInt64() && origD.IsInt64() {
		got, err = fraction.NewReduced(int(origN.Int64()), int(origD.Int64()))
		wCheckFrac(t, "NewReduced", got, err, wFromRat(r))
	}
}
