package timecalc_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/internal/timecalc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rat(text string) *big.Rat {
	r, ok := new(big.Rat).SetString(text)
	if !ok {
		panic("invalid test rational: " + text)
	}
	return r
}

func logForm(constant string, terms ...timecalc.LogTerm) timecalc.LogForm {
	return timecalc.LogForm{Constant: rat(constant), Terms: terms}
}

func term(coefficient, base string) timecalc.LogTerm {
	return timecalc.LogTerm{Coefficient: rat(coefficient), Base: rat(base)}
}

func requireRat(t *testing.T, want, got *big.Rat) {
	t.Helper()
	require.NotNil(t, got)
	assert.Zero(t, want.Cmp(got), "want %s, got %s", want, got)
}

func requireRational(t *testing.T, form timecalc.LogForm, want string) {
	t.Helper()
	got, exact := timecalc.Rational(form)
	require.True(t, exact)
	requireRat(t, rat(want), got)
}

// These identities exercise the complete multiplicative relation, including
// rational coefficients and rational bases (architecture 15.2 and old 7.1).
func TestRationalAndExactZero(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		form  timecalc.LogForm
		value string
	}{
		{name: "constant zero", form: timecalc.Constant(rat("0")), value: "0"},
		{name: "negative constant", form: timecalc.Constant(rat("-7/3")), value: "-7/3"},
		{name: "empty terms", form: timecalc.LogForm{Constant: rat("2"), Terms: []timecalc.LogTerm{}}, value: "2"},
		{name: "discard neutral terms", form: logForm("3", term("0", "5"), term("9", "1")), value: "3"},
		{name: "ln4 minus twice ln2", form: logForm("0", term("1", "4"), term("-2", "2")), value: "0"},
		{name: "mixed prime bases", form: logForm("7/3", term("1", "6"), term("-1", "2"), term("-1", "3")), value: "7/3"},
		{name: "fractional coefficients and bases", form: logForm("0", term("1/2", "9/4"), term("1", "2/3")), value: "0"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireRational(t, tt.form, tt.value)
			assert.Zero(t, timecalc.CompareLog(tt.form, rat(tt.value)))
		})
	}
	for _, tt := range []struct {
		name string
		form timecalc.LogForm
	}{
		{name: "nonzero algebraic right side", form: logForm("1", term("1/2", "2"))},
		{name: "nearly canceled logarithms", form: logForm("0", term("1", "4"), term("-2", "2"), term("1", "1000001/1000000"))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			value, exact := timecalc.Rational(tt.form)
			assert.False(t, exact)
			assert.Nil(t, value)
			assert.Equal(t, 1, timecalc.CompareLog(tt.form, rat("0")))
		})
	}
}

func TestAddAndScale(t *testing.T) {
	t.Parallel()
	a := logForm("3/2", term("2", "2"))
	b := logForm("-1/2", term("-1", "4"))
	requireRational(t, timecalc.Add(a, b), "1")
	requireRational(t, timecalc.Scale(a, rat("0")), "0")
	requireRational(t, timecalc.Scale(timecalc.Constant(rat("3/2")), rat("-2/3")), "-1")
	// 3*a/2 + 3*b/2 = 3/2, including both constant fields.
	requireRational(t, timecalc.Add(timecalc.Scale(a, rat("3/2")), timecalc.Scale(b, rat("3/2"))), "3/2")
	// An irrational result must retain the exact sign and logarithmic weight.
	weighted := timecalc.Add(timecalc.Scale(a, rat("-1/2")), timecalc.Constant(rat("3/4")))
	difference := timecalc.Add(weighted, logForm("0", term("1", "2")))
	requireRational(t, difference, "0")
	assert.Equal(t, -1, timecalc.CompareLog(weighted, rat("0")))
	large := new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 300))
	logarithm := logForm("0", term("1", "2"))
	scaled := timecalc.Scale(timecalc.Scale(logarithm, large), new(big.Rat).Inv(large))
	requireRational(t, timecalc.Add(scaled, logForm("0", term("-1", "2"))), "0")
	requireRational(t, timecalc.Add(timecalc.Constant(large), timecalc.Constant(new(big.Rat).Neg(large))), "0")
}

func TestCompareLogSigned(t *testing.T) {
	t.Parallel()
	// ln2-1 is between -1/2 and -1/4. Its negative is positive.
	form := logForm("-1", term("1", "2"))
	for _, tt := range []struct {
		value string
		want  int
	}{
		{value: "0", want: -1},
		{value: "-1", want: 1},
		{value: "-1/4", want: -1},
		{value: "-1/2", want: 1},
	} {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, timecalc.CompareLog(form, rat(tt.value)))
			assert.Equal(t, -tt.want, timecalc.CompareLog(timecalc.Scale(form, rat("-1")), new(big.Rat).Neg(rat(tt.value))))
		})
	}
}

func TestLinearSpan(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name               string
		a, b, s0, s1, x, y string
		want               timecalc.LogForm
	}{
		{name: "constant subspan", a: "2", b: "6", s0: "30", s1: "30", x: "3", y: "9/2", want: timecalc.Constant(rat("3"))},
		{name: "zero span", a: "0", b: "1", s0: "60", s1: "120", x: "1/2", y: "1/2", want: timecalc.Constant(rat("0"))},
		{name: "accelerating full span", a: "0", b: "1", s0: "60", s1: "120", x: "0", y: "1", want: logForm("0", term("1", "2"))},
		{name: "decelerating full span", a: "0", b: "1", s0: "120", s1: "60", x: "0", y: "1", want: logForm("0", term("1", "2"))},
		{name: "interpolated endpoints", a: "2", b: "4", s0: "30", s1: "90", x: "5/2", y: "7/2", want: logForm("0", term("2", "5/3"))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := timecalc.LinearSpan(rat(tt.a), rat(tt.b), rat(tt.s0), rat(tt.s1), rat(tt.x), rat(tt.y))
			requireRational(t, timecalc.Add(got, timecalc.Scale(tt.want, rat("-1"))), "0")
		})
	}
	first := timecalc.LinearSpan(rat("0"), rat("1"), rat("60"), rat("120"), rat("0"), rat("1/2"))
	last := timecalc.LinearSpan(rat("0"), rat("1"), rat("60"), rat("120"), rat("1/2"), rat("1"))
	requireRational(t, timecalc.Add(timecalc.Add(first, last), logForm("0", term("-1", "2"))), "0")
}

// Every coordinate and speed may exceed fraction's 255-bit components.
// Translation and a common speed/span scale cancel exactly in 60*dx/s.
func TestLinearSpanBeyondFraction(t *testing.T) {
	t.Parallel()
	large := new(big.Int).Lsh(big.NewInt(1), 300)
	a := new(big.Rat).SetFrac(big.NewInt(1), large)
	b := new(big.Rat).Add(a, rat("1"))
	x := new(big.Rat).Add(a, rat("1/2"))
	got := timecalc.LinearSpan(a, b, rat("60"), rat("120"), x, b)
	requireRational(t, timecalc.Add(got, logForm("0", term("-1", "4/3"))), "0")
	span := new(big.Rat).SetInt(large)
	speed := new(big.Rat).Mul(rat("60"), span)
	requireRational(t, timecalc.LinearSpan(rat("0"), span, speed, speed, rat("0"), span), "1")
}
