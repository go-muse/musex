package timecalc_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/internal/timecalc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var modes = []rounding.Mode{
	rounding.Down, rounding.Up, rounding.TowardZero, rounding.AwayFromZero,
	rounding.HalfUp, rounding.HalfDown, rounding.HalfAwayFromZero, rounding.HalfTowardZero, rounding.HalfEven,
}

func requireInt(t *testing.T, want string, got *big.Int) {
	t.Helper()
	require.NotNil(t, got)
	assert.Equal(t, want, got.String())
}

func TestRoundLogStepsRational(t *testing.T) {
	t.Parallel()
	assert.Equal(t, modes, rounding.Modes())
	for _, tt := range []struct {
		name, value, step string
		want              [9]string
	}{
		{name: "zero", value: "0", step: "1/7", want: [9]string{"0", "0", "0", "0", "0", "0", "0", "0", "0"}},
		{name: "even lower midpoint", value: "1/6", step: "1/3", want: [9]string{"0", "1", "0", "1", "1", "0", "1", "0", "0"}},
		{name: "odd lower midpoint", value: "1/2", step: "1/3", want: [9]string{"1", "2", "1", "2", "2", "1", "2", "1", "2"}},
		{name: "exact multiple", value: "2/3", step: "1/3", want: [9]string{"2", "2", "2", "2", "2", "2", "2", "2", "2"}},
		{name: "below midpoint", value: "7/9", step: "1/3", want: [9]string{"2", "3", "2", "3", "2", "2", "2", "2", "2"}},
		{name: "above midpoint", value: "8/9", step: "1/3", want: [9]string{"2", "3", "2", "3", "3", "3", "3", "3", "3"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for i, mode := range modes {
				t.Run(mode.String(), func(t *testing.T) {
					t.Parallel()
					requireInt(t, tt.want[i], timecalc.RoundLogSteps(timecalc.Constant(rat(tt.value)), rat(tt.step), mode))
				})
			}
		})
	}
}

func TestRoundLogStepsUnbounded(t *testing.T) {
	t.Parallel()
	large := new(big.Int).Lsh(big.NewInt(1), 300)
	largeRat := new(big.Rat).SetInt(large)
	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			requireInt(t, large.String(), timecalc.RoundLogSteps(timecalc.Constant(largeRat), rat("1"), mode))
			requireInt(t, "1", timecalc.RoundLogSteps(timecalc.Constant(largeRat), largeRat, mode))
			// 3/2 +/- 2^-300 distinguishes a true tie from a tiny offset.
			epsilon := new(big.Rat).Inv(largeRat)
			below := new(big.Rat).Sub(rat("3/2"), epsilon)
			above := new(big.Rat).Add(rat("3/2"), epsilon)
			low, high := "1", "2"
			if mode == rounding.Up || mode == rounding.AwayFromZero {
				low = "2"
			}
			if mode == rounding.Down || mode == rounding.TowardZero {
				high = "1"
			}
			requireInt(t, low, timecalc.RoundLogSteps(timecalc.Constant(below), rat("1"), mode))
			requireInt(t, high, timecalc.RoundLogSteps(timecalc.Constant(above), rat("1"), mode))
			wantSmall := "0"
			if mode == rounding.Up || mode == rounding.AwayFromZero {
				wantSmall = "1"
			}
			requireInt(t, wantSmall, timecalc.RoundLogSteps(logForm("0", term("1", "2")), largeRat, mode))
		})
	}
	wideBase := timecalc.LogForm{Constant: rat("0"), Terms: []timecalc.LogTerm{{Coefficient: rat("1"), Base: largeRat}}}
	// ln(2^300)=300*ln2=207.944154167983592825... .
	requireInt(t, "208", timecalc.RoundLogSteps(wideBase, rat("1"), rounding.HalfUp))
}

// Independent Decimal(110): ln(2)*2^80 =
// 837963523372001241319907.70153832132080841330395133033749...
// At 300 bits, independent Decimal(140) gives the second index below plus
// .72963746481872144686... . The step also lies outside fraction's range.
// Both indices exceed int64; every nearest mode must round upward.
func TestRoundLogStepsIrrational(t *testing.T) {
	t.Parallel()
	form := logForm("0", term("1", "2"))
	for _, tt := range []struct {
		bits         uint
		lower, upper string
	}{
		{bits: 80, lower: "837963523372001241319907", upper: "837963523372001241319908"},
		{bits: 300, lower: "1411965743695424507508857192728570223734551239724995589061361292466257174357546532574725854", upper: "1411965743695424507508857192728570223734551239724995589061361292466257174357546532574725855"},
	} {
		t.Run(big.NewInt(int64(tt.bits)).String(), func(t *testing.T) {
			t.Parallel()
			step := new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), tt.bits))
			for _, mode := range modes {
				t.Run(mode.String(), func(t *testing.T) {
					t.Parallel()
					want := tt.upper
					if mode == rounding.Down || mode == rounding.TowardZero {
						want = tt.lower
					}
					requireInt(t, want, timecalc.RoundLogSteps(form, step, mode))
				})
			}
		})
	}
}

func TestDecimalIntervalRational(t *testing.T) {
	t.Parallel()
	for _, places := range []int{0, 3, 1000} {
		t.Run(big.NewInt(int64(places)).String(), func(t *testing.T) {
			t.Parallel()
			form := timecalc.Constant(rat("1/3"))
			lo, hi, exact := timecalc.DecimalInterval(form, places)
			require.True(t, exact)
			requireRat(t, rat("1/3"), lo)
			requireRat(t, rat("1/3"), hi)
			lo.SetInt64(99)
			requireRat(t, rat("1/3"), hi)
			requireRational(t, form, "1/3")
			hi.SetInt64(88)
			requireRational(t, form, "1/3")
		})
	}
}

// Centers were computed independently using Python Decimal with 110 digits.
// The decimal bracket is also independent of RoundLogSteps; assertions
// check the exact endpoints as well as inclusion of the true logarithm.
func TestDecimalIntervalIrrational(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		places int
		center string
	}{
		{places: 0, center: "1"},
		{places: 3, center: ".693"},
		{places: 20, center: ".69314718055994530942"},
		{places: 80, center: ".69314718055994530941723212145817656807550013436025525412068000949339362196969472"},
	} {
		t.Run(big.NewInt(int64(tt.places)).String(), func(t *testing.T) {
			t.Parallel()
			lo, hi, exact := timecalc.DecimalInterval(logForm("0", term("1", "2")), tt.places)
			assert.False(t, exact)
			half := new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Mul(big.NewInt(2), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tt.places)), nil)))
			requireRat(t, new(big.Rat).Sub(rat(tt.center), half), lo)
			requireRat(t, new(big.Rat).Add(rat(tt.center), half), hi)
			lower := rat(".6931471805599453094172321214581765680755001343602552541206800094933936219696947156058633269964186875")
			upper := rat(".6931471805599453094172321214581765680755001343602552541206800094933936219696947156058633269964186876")
			assert.Negative(t, lo.Cmp(lower))
			assert.Positive(t, hi.Cmp(upper))
		})
	}
	// Core intervals are not clamped by the positive-time public layer.
	lo, hi, exact := timecalc.DecimalInterval(logForm("0", term("1", "4/3")), 0)
	assert.False(t, exact)
	requireRat(t, rat("-1/2"), lo)
	requireRat(t, rat("1/2"), hi)
}
