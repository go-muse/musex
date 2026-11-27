package fraction_test

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of F-1 (task 26.6, architecture "API changes (F-1)" and "Tests of F-1
// and of the DivInDenominator error order"): the nine rounding modes of
// common/rounding in every rounding method, the mode constants, mode errors
// and their priority, RoundToMultiple, and the error order of DivInDenominator.
// The expected values come from rounding.Mode.RatInt and RatMultiple on the
// exact value being rounded, and from the contract examples.

// f1Modes lists the nine named rounding modes of the package.
var f1Modes = []struct {
	name string
	mode fraction.RoundingMode
}{
	{"RoundFloor", fraction.RoundFloor},
	{"RoundCeil", fraction.RoundCeil},
	{"RoundTrunc", fraction.RoundTrunc},
	{"RoundAwayFromZero", fraction.RoundAwayFromZero},
	{"RoundHalfUp", fraction.RoundHalfUp},
	{"RoundHalfDown", fraction.RoundHalfDown},
	{"RoundHalfAwayFromZero", fraction.RoundHalfAwayFromZero},
	{"RoundHalfTowardZero", fraction.RoundHalfTowardZero},
	{"RoundHalfToEven", fraction.RoundHalfToEven},
}

// f1BadModes lists the unset mode and unknown modes with the rounding sentinel
// their error must wrap besides ErrNotRepresentable.
var f1BadModes = []struct {
	name     string
	mode     fraction.RoundingMode
	sentinel error
}{
	{"unset", fraction.RoundingMode(0), rounding.ErrUnset},
	{"unknown_10", fraction.RoundingMode(10), rounding.ErrUnknownMode},
	{"unknown_99", fraction.RoundingMode(99), rounding.ErrUnknownMode},
	{"unknown_255", fraction.RoundingMode(255), rounding.ErrUnknownMode},
}

func TestRoundingModeConstants(t *testing.T) {
	t.Parallel()

	// RoundingMode and rounding.Mode are one type: values pass in either
	// direction without conversion, which the compiler checks here.
	asRounding := func(m rounding.Mode) rounding.Mode { return m }
	asFraction := func(m fraction.RoundingMode) fraction.RoundingMode { return m }
	assert.Equal(t, rounding.HalfUp, asRounding(fraction.RoundHalfUp))
	assert.Equal(t, rounding.Down, asRounding(f1Modes[0].mode))
	assert.Equal(t, fraction.RoundHalfToEven, asFraction(rounding.HalfEven))
	up := rounding.Up
	assert.Equal(t, fraction.RoundCeil, asFraction(up))

	want := map[string]rounding.Mode{
		"RoundFloor":            rounding.Down,
		"RoundCeil":             rounding.Up,
		"RoundTrunc":            rounding.TowardZero,
		"RoundAwayFromZero":     rounding.AwayFromZero,
		"RoundHalfUp":           rounding.HalfUp,
		"RoundHalfDown":         rounding.HalfDown,
		"RoundHalfAwayFromZero": rounding.HalfAwayFromZero,
		"RoundHalfTowardZero":   rounding.HalfTowardZero,
		"RoundHalfToEven":       rounding.HalfEven,
	}
	seen := make(map[fraction.RoundingMode]bool)
	for _, m := range f1Modes {
		assert.Equal(t, want[m.name], m.mode, m.name)
		assert.True(t, m.mode.IsValid(), m.name)
		assert.False(t, seen[m.mode], "%s must be distinct", m.name)
		seen[m.mode] = true
	}
	assert.ElementsMatch(t, rounding.Modes(), func() []rounding.Mode {
		var all []rounding.Mode
		for _, m := range f1Modes {
			all = append(all, m.mode)
		}
		return all
	}())

	var zero fraction.RoundingMode
	assert.False(t, zero.IsValid(), "the zero value is the unset mode")
}

// TestRoundToIntNineModesContractValues checks the values of criterion K-22.
func TestRoundToIntNineModesContractValues(t *testing.T) {
	t.Parallel()

	inputs := [][2]int{{1, 2}, {-1, 2}, {3, 2}, {-3, 2}, {5, 2}, {-5, 2}, {7, 3}, {-7, 3}}
	cases := []struct {
		name string
		mode fraction.RoundingMode
		want []int
	}{
		{"RoundAwayFromZero", fraction.RoundAwayFromZero, []int{1, -1, 2, -2, 3, -3, 3, -3}},
		{"RoundHalfUp", fraction.RoundHalfUp, []int{1, 0, 2, -1, 3, -2, 2, -2}},
		{"RoundHalfDown", fraction.RoundHalfDown, []int{0, -1, 1, -2, 2, -3, 2, -2}},
		{"RoundHalfTowardZero", fraction.RoundHalfTowardZero, []int{0, 0, 1, -1, 2, -2, 2, -2}},
		{"RoundHalfAwayFromZero", fraction.RoundHalfAwayFromZero, []int{1, -1, 2, -2, 3, -3, 2, -2}},
		{"RoundHalfToEven", fraction.RoundHalfToEven, []int{0, 0, 2, -2, 2, -2, 2, -2}},
		{"RoundFloor", fraction.RoundFloor, []int{0, -1, 1, -2, 2, -3, 2, -3}},
		{"RoundCeil", fraction.RoundCeil, []int{1, 0, 2, -1, 3, -2, 3, -2}},
		{"RoundTrunc", fraction.RoundTrunc, []int{0, 0, 1, -1, 2, -2, 2, -2}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			for i, in := range inputs {
				got, err := fraction.MustNew(in[0], in[1]).RoundToInt(c.mode)
				require.NoError(t, err)
				assert.Equal(t, c.want[i], got, "%d/%d", in[0], in[1])
			}

			// (1, 3) to D = 10: (3, 10), except RoundCeil and RoundAwayFromZero: (4, 10).
			wantNum := 3
			if c.mode == fraction.RoundCeil || c.mode == fraction.RoundAwayFromZero {
				wantNum = 4
			}
			got, err := fraction.MustNew(1, 3).RoundToDenominator(10, c.mode)
			require.NoError(t, err)
			assertExactPairInt(t, got, wantNum, 10)
		})
	}
}

// f1Inputs returns the receivers of the nine-mode tests: ties, values next to
// integers, negative values, zeros, integers, negative original denominators
// and values of every width class.
func f1Inputs() []fraction.Fraction {
	p := wPow2
	maxC := fraction.MaxComponentBig()
	ints := [][2]int{
		{1, 2}, {-1, 2}, {3, 2}, {-3, 2}, {5, 2}, {-5, 2}, {5, -2}, {-10, 4}, {15, -6},
		{1999, 1000}, {-2001, 1000}, {1, 1000}, {-1, 1000}, {999, -1000},
		{0, 1}, {0, -7}, {6, 3}, {-4, 2}, {7, -3}, {-7, 3},
		{math.MaxInt, 2}, {math.MinInt, 3}, {math.MinInt, -1},
	}
	var fs []fraction.Fraction
	for _, in := range ints {
		fs = append(fs, fraction.MustNew(in[0], in[1]))
	}
	fs = append(fs,
		wFrac(wOff(p(200), 1), big.NewInt(2)),        // tie beyond int
		wFrac(wNeg(wOff(p(64), -1)), big.NewInt(-2)), // tie next to 2^63
		wFrac(maxC, big.NewInt(1)),                   // MaxComponent
		wFrac(wNeg(maxC), big.NewInt(3)),             // −MaxComponent/3
		wFrac(maxC, big.NewInt(2)),                   // MaxComponent/2, a tie
		wFrac(wOff(p(127), 1), wNeg(wOff(p(64), 1))), // two-word values
	)
	g := wOff(p(64), 13)
	fs = append(fs, wFrac(wMul(g, big.NewInt(7)), wMul(g, big.NewInt(2)))) // reducible tie 7/2
	for _, family := range []benchFamily{benchSplit, benchRoundDen, benchReduce} {
		for _, in := range benchInputs(family) {
			fs = append(fs, in.x)
		}
	}
	return fs
}

// f1Steps returns the steps of the RoundToMultiple tests, including a zero and
// negative values, which are errors.
func f1Steps() []fraction.Fraction {
	return []fraction.Fraction{
		fraction.MustNew(1, 4), fraction.MustNew(-1, -4), fraction.MustNew(1, 3), fraction.MustNew(3, 2),
		fraction.MustNew(7, 1), fraction.MustNew(2, 1), fraction.MustNew(5, 10),
		wFrac(big.NewInt(1), wOff(wPow2(64), 1)), wFrac(big.NewInt(1), wPow2(200)),
		wFrac(wPow2(100), big.NewInt(3)), wFrac(fraction.MaxComponentBig(), big.NewInt(1)),
		fraction.MustNew(0, 5), fraction.MustNew(-1, 4), fraction.MustNew(1, -3),
	}
}

// f1RatInt rounds the exact value x by mode with the common/rounding oracle.
func f1RatInt(t *testing.T, mode fraction.RoundingMode, x *big.Rat) *big.Int {
	t.Helper()
	m, err := mode.RatInt(x)
	require.NoError(t, err)
	return m
}

// f1FormatInt formats the integer m scaled by 10^-digits per section 8 of the contract.
func f1FormatInt(m *big.Int, digits int) string {
	if digits == 0 {
		return m.String()
	}
	sign := ""
	if m.Sign() < 0 {
		sign = "-"
	}
	s := new(big.Int).Abs(m).String()
	if len(s) <= digits {
		s = strings.Repeat("0", digits+1-len(s)) + s
	}
	return sign + s[:len(s)-digits] + "." + s[len(s)-digits:]
}

// f1CheckMustSame checks that fn panics with exactly the value err, or does not panic when err is nil.
func f1CheckMustSame(t *testing.T, label string, fn func(), err error) {
	t.Helper()
	rec := wCatch(fn)
	if err == nil {
		assert.Nilf(t, rec, "%s: Must must not panic", label)
		return
	}
	assert.Equalf(t, err, rec, "%s: Must must panic with the error of the method", label)
}

func TestRoundingNineModes(t *testing.T) {
	t.Parallel()

	maxC := fraction.MaxComponentBig()
	for _, m := range f1Modes {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			mode := m.mode
			for i, f := range f1Inputs() {
				x := wRat(f)
				label := fmt.Sprintf("#%d %s", i, f)

				// RoundToBigInt and RoundToInt: x = f.
				want := f1RatInt(t, mode, x)
				gotBig, err := f.RoundToBigInt(mode)
				if assert.NoErrorf(t, err, "%s RoundToBigInt", label) {
					wCheckBigInt(t, label+" RoundToBigInt", gotBig, want)
				}
				gotInt, err := f.RoundToInt(mode)
				if want.IsInt64() {
					if assert.NoErrorf(t, err, "%s RoundToInt", label) {
						assert.Equalf(t, want.Int64(), int64(gotInt), "%s RoundToInt", label)
					}
				} else {
					wCheckErr(t, label+" RoundToInt", err, fraction.ErrNotRepresentable)
					assert.Zerof(t, gotInt, "%s RoundToInt on error", label)
				}
				f1CheckMustSame(t, label+" MustRoundToInt", func() { f.MustRoundToInt(mode) }, err)

				// RoundToDenominator and RoundToBigDenominator: x = f·den, result (m, den).
				dens := []*big.Int{
					big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(10),
					big.NewInt(1 << 62), big.NewInt(math.MaxInt),
					wOff(wPow2(64), 1), wPow2(200), maxC, wOff(maxC, 1),
				}
				for _, den := range dens {
					var wantD wideWant
					if !wInRange(den) {
						wantD = wErr(fraction.ErrOutOfRange)
					} else {
						wantD = wRange(f1RatInt(t, mode, new(big.Rat).Mul(x, new(big.Rat).SetInt(den))), den)
					}
					dl := fmt.Sprintf("%s den %s", label, den)
					got, err := f.RoundToBigDenominator(den, mode)
					wCheckFrac(t, dl+" RoundToBigDenominator", got, err, wantD)
					f1CheckMustSame(t, dl+" MustRoundToBigDenominator", func() { f.MustRoundToBigDenominator(den, mode) }, err)
					if den.IsInt64() {
						got, err := f.RoundToDenominator(int(den.Int64()), mode)
						wCheckFrac(t, dl+" RoundToDenominator", got, err, wantD)
						f1CheckMustSame(t, dl+" MustRoundToDenominator", func() { f.MustRoundToDenominator(int(den.Int64()), mode) }, err)
					}
				}

				// FormatDecimal: x = f·10^digits, formatted per section 8.
				for _, digits := range []int{0, 1, 2, 5, 40, fraction.MaxDigits} {
					scale := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil))
					wantS := f1FormatInt(f1RatInt(t, mode, new(big.Rat).Mul(x, scale)), digits)
					got, err := f.FormatDecimal(digits, mode)
					if assert.NoErrorf(t, err, "%s FormatDecimal(%d)", label, digits) {
						assert.Equalf(t, wantS, got, "%s FormatDecimal(%d)", label, digits)
					}
					assert.Equalf(t, wantS, f.MustFormatDecimal(digits, mode), "%s MustFormatDecimal(%d)", label, digits)
				}

				// RoundToMultiple and RoundToMultipleBig: k = round(f/step), v = k·step reduced.
				for _, step := range f1Steps() {
					sl := fmt.Sprintf("%s step %s", label, step)
					checkRoundToMultiple(t, sl, f, step, mode)
				}
			}
		})
	}
}

// checkRoundToMultiple compares both RoundToMultiple forms and their Must forms with the oracle.
func checkRoundToMultiple(t *testing.T, label string, f, step fraction.Fraction, mode fraction.RoundingMode) {
	t.Helper()
	fBefore, stepBefore := f, step

	var rawK, wantK *big.Int // rawK is the rounded index even when v is out of range
	var wantV wideWant
	rs := wRat(step)
	switch {
	case rs.Sign() == 0:
		wantV = wErr(fraction.ErrDivisionByZero)
	case rs.Sign() < 0:
		wantV = wErr(fraction.ErrZeroDenominator)
	default:
		k, v, err := mode.RatMultiple(wRat(f), rs)
		require.NoError(t, err)
		rawK = k
		wantV = wFromRat(v)
		if wantV.err == nil {
			wantK = k
		}
	}

	kb, vb, err := f.RoundToMultipleBig(step, mode)
	wCheckBigInt(t, label+" RoundToMultipleBig k", kb, wantK)
	wCheckFrac(t, label+" RoundToMultipleBig v", vb, err, wantV)
	f1CheckMustSame(t, label+" MustRoundToMultipleBig", func() { f.MustRoundToMultipleBig(step, mode) }, err)

	// Step 3 (k beyond int) precedes step 4 (v beyond the range) in the narrow form.
	narrowV := wantV
	if rawK != nil && !rawK.IsInt64() {
		narrowV = wErr(fraction.ErrNotRepresentable)
	}
	k, v, err := f.RoundToMultiple(step, mode)
	if narrowV.err == nil {
		assert.Equalf(t, wantK.Int64(), int64(k), "%s RoundToMultiple k", label)
	} else {
		assert.Zerof(t, k, "%s RoundToMultiple k on error", label)
	}
	wCheckFrac(t, label+" RoundToMultiple v", v, err, narrowV)
	f1CheckMustSame(t, label+" MustRoundToMultiple", func() { f.MustRoundToMultiple(step, mode) }, err)

	assert.Equalf(t, fBefore, f, "%s: receiver unchanged", label)
	assert.Equalf(t, stepBefore, step, "%s: step unchanged", label)
}

// TestRoundToMultipleContract covers the examples and boundary cases of the contract.
func TestRoundToMultipleContract(t *testing.T) {
	t.Parallel()

	maxC := fraction.MaxComponentBig()
	quarter := fraction.MustNew(1, 4)

	t.Run("examples", func(t *testing.T) {
		t.Parallel()
		k, v, err := fraction.MustNew(7, 10).RoundToMultiple(quarter, fraction.RoundHalfToEven)
		require.NoError(t, err)
		assert.Equal(t, 3, k)
		assertExactPairInt(t, v, 3, 4)

		k, v, err = fraction.MustNew(5, 8).RoundToMultiple(quarter, fraction.RoundHalfToEven)
		require.NoError(t, err)
		assert.Equal(t, 2, k, "a tie goes to the even k")
		assertExactPairInt(t, v, 1, 2)

		// The value is in reduced canonical form with a positive denominator.
		k, v, err = fraction.MustNew(-9, 12).RoundToMultiple(fraction.MustNew(-2, -4), fraction.RoundFloor)
		require.NoError(t, err)
		assert.Equal(t, -2, k)
		assertExactPairInt(t, v, -1, 1)

		k, v, err = fraction.MustNew(0, -3).RoundToMultiple(fraction.MustNew(6, 4), fraction.RoundCeil)
		require.NoError(t, err)
		assert.Equal(t, 0, k)
		assertExactPairInt(t, v, 0, 1)
	})

	t.Run("step_zero_and_negative", func(t *testing.T) {
		t.Parallel()
		f := fraction.MustNew(1, 2)
		for _, step := range []fraction.Fraction{fraction.Zero(), fraction.MustNew(0, -5)} {
			k, v, err := f.RoundToMultiple(step, fraction.RoundHalfToEven)
			wCheckErr(t, "RoundToMultiple(0)", err, fraction.ErrDivisionByZero)
			assert.Zero(t, k)
			assert.Equal(t, fraction.Fraction{}, v)
			kb, vb, err := f.RoundToMultipleBig(step, fraction.RoundHalfToEven)
			wCheckErr(t, "RoundToMultipleBig(0)", err, fraction.ErrDivisionByZero)
			assert.Nil(t, kb)
			assert.Equal(t, fraction.Fraction{}, vb)
		}
		for _, step := range []fraction.Fraction{fraction.MustNew(-1, 4), fraction.MustNew(1, -4)} {
			_, _, err := f.RoundToMultiple(step, fraction.RoundHalfToEven)
			wCheckErr(t, "RoundToMultiple(<0)", err, fraction.ErrZeroDenominator)
			_, _, err = f.RoundToMultipleBig(step, fraction.RoundHalfToEven)
			wCheckErr(t, "RoundToMultipleBig(<0)", err, fraction.ErrZeroDenominator)
		}
	})

	t.Run("k_beyond_int", func(t *testing.T) {
		t.Parallel()
		f := wFrac(wOff(wPow2(100), 1), big.NewInt(1))
		k, v, err := f.RoundToMultiple(fraction.One(), fraction.RoundHalfToEven)
		wCheckErr(t, "RoundToMultiple", err, fraction.ErrNotRepresentable)
		assert.Zero(t, k)
		assert.Equal(t, fraction.Fraction{}, v)
		f1CheckMustSame(t, "MustRoundToMultiple", func() { f.MustRoundToMultiple(fraction.One(), fraction.RoundHalfToEven) }, err)

		kb, vb, err := f.RoundToMultipleBig(fraction.One(), fraction.RoundHalfToEven)
		require.NoError(t, err, "the wide form succeeds")
		wCheckBigInt(t, "RoundToMultipleBig k", kb, wOff(wPow2(100), 1))
		wCheckFrac(t, "RoundToMultipleBig v", vb, nil, wideWant{num: wOff(wPow2(100), 1), den: big.NewInt(1)})
	})

	t.Run("v_beyond_range", func(t *testing.T) {
		t.Parallel()
		// MaxComponent/2 = 2^254 − 1/2 to the step 2: rounding up gives v = 2^255.
		f := wFrac(maxC, big.NewInt(1))
		two := fraction.MustNew(2, 1)
		for _, mode := range []fraction.RoundingMode{fraction.RoundCeil, fraction.RoundHalfUp, fraction.RoundAwayFromZero} {
			_, v, err := f.RoundToMultiple(two, mode)
			wCheckErr(t, "RoundToMultiple", err, fraction.ErrNotRepresentable)
			assert.Equal(t, fraction.Fraction{}, v)
			kb, vb, err := f.RoundToMultipleBig(two, mode)
			wCheckErr(t, "RoundToMultipleBig", err, fraction.ErrOutOfRange)
			assert.Nil(t, kb)
			assert.Equal(t, fraction.Fraction{}, vb)
			f1CheckMustSame(t, "MustRoundToMultipleBig", func() { f.MustRoundToMultipleBig(two, mode) }, err)
		}
		// k = 3 fits int and v = MaxComponent is in range.
		step := wFrac(maxC, big.NewInt(3))
		_, _, err := wFrac(maxC, big.NewInt(1)).RoundToMultiple(step, fraction.RoundCeil)
		require.NoError(t, err)
		// MaxComponent / (2^253 + 1) ≈ 3.99 rounds up to k = 4, which fits int,
		// while v = 2^255 + 4 is beyond the range.
		bigStep := wFrac(wOff(wPow2(253), 1), big.NewInt(1))
		k, v, err := wFrac(maxC, big.NewInt(1)).RoundToMultiple(bigStep, fraction.RoundCeil)
		wCheckErr(t, "RoundToMultiple v beyond range", err, fraction.ErrOutOfRange)
		assert.Zero(t, k)
		assert.Equal(t, fraction.Fraction{}, v)
		f1CheckMustSame(t, "MustRoundToMultiple", func() {
			wFrac(maxC, big.NewInt(1)).MustRoundToMultiple(bigStep, fraction.RoundCeil)
		}, err)
	})

	t.Run("wide_operands", func(t *testing.T) {
		t.Parallel()
		for _, family := range []benchFamily{benchDivMod, benchBinary} {
			for _, in := range benchInputs(family) {
				for _, m := range f1Modes {
					checkRoundToMultiple(t, fmt.Sprintf("%d/%s %s", family, in.class, m.name), in.x, in.y, m.mode)
					checkRoundToMultiple(t, fmt.Sprintf("%d/%s %s reversed", family, in.class, m.name), in.y, in.x, m.mode)
				}
			}
		}
	})
}

// f1RoundCall is one rounding method with a valid receiver; it returns the error only.
type f1RoundCall struct {
	name string
	call func(f fraction.Fraction, mode fraction.RoundingMode) error
	must func(f fraction.Fraction, mode fraction.RoundingMode)
}

var f1RoundCalls = []f1RoundCall{
	{"RoundToInt", func(f fraction.Fraction, m fraction.RoundingMode) error {
		_, err := f.RoundToInt(m)
		return err
	}, func(f fraction.Fraction, m fraction.RoundingMode) { f.MustRoundToInt(m) }},
	{"RoundToBigInt", func(f fraction.Fraction, m fraction.RoundingMode) error {
		_, err := f.RoundToBigInt(m)
		return err
	}, func(f fraction.Fraction, m fraction.RoundingMode) { f.MustRoundToBigInt(m) }},
	{"RoundToDenominator", func(f fraction.Fraction, m fraction.RoundingMode) error {
		_, err := f.RoundToDenominator(10, m)
		return err
	}, func(f fraction.Fraction, m fraction.RoundingMode) { f.MustRoundToDenominator(10, m) }},
	{"RoundToBigDenominator", func(f fraction.Fraction, m fraction.RoundingMode) error {
		_, err := f.RoundToBigDenominator(big.NewInt(10), m)
		return err
	}, func(f fraction.Fraction, m fraction.RoundingMode) { f.MustRoundToBigDenominator(big.NewInt(10), m) }},
	{"FormatDecimal", func(f fraction.Fraction, m fraction.RoundingMode) error {
		_, err := f.FormatDecimal(2, m)
		return err
	}, func(f fraction.Fraction, m fraction.RoundingMode) { f.MustFormatDecimal(2, m) }},
	{"RoundToMultiple", func(f fraction.Fraction, m fraction.RoundingMode) error {
		_, _, err := f.RoundToMultiple(fraction.MustNew(1, 4), m)
		return err
	}, func(f fraction.Fraction, m fraction.RoundingMode) { f.MustRoundToMultiple(fraction.MustNew(1, 4), m) }},
	{"RoundToMultipleBig", func(f fraction.Fraction, m fraction.RoundingMode) error {
		_, _, err := f.RoundToMultipleBig(fraction.MustNew(1, 4), m)
		return err
	}, func(f fraction.Fraction, m fraction.RoundingMode) {
		f.MustRoundToMultipleBig(fraction.MustNew(1, 4), m)
	}},
}

// f1CheckModeErr checks a mode error: ErrNotRepresentable and the rounding
// sentinel, and none of the other sentinels of the package.
func f1CheckModeErr(t *testing.T, label string, err, sentinel error) {
	t.Helper()
	assert.Truef(t, wSentinelOnly(err, fraction.ErrNotRepresentable), "%s: %v must match only ErrNotRepresentable among the package sentinels", label, err)
	assert.Truef(t, errors.Is(err, sentinel), "%s: %v must match %v", label, err, sentinel)
}

func TestRoundingModeErrors(t *testing.T) {
	t.Parallel()

	receivers := []fraction.Fraction{
		fraction.One(), fraction.MustNew(-5, 2), fraction.MustNew(0, 3),
		wFrac(fraction.MaxComponentBig(), big.NewInt(7)),
	}
	for _, bad := range f1BadModes {
		t.Run(bad.name, func(t *testing.T) {
			t.Parallel()
			for _, f := range receivers {
				for _, c := range f1RoundCalls {
					label := fmt.Sprintf("%s(%s)", c.name, f)
					err := c.call(f, bad.mode)
					f1CheckModeErr(t, label, err, bad.sentinel)
					f1CheckMustSame(t, "Must"+label, func() { c.must(f, bad.mode) }, err)
				}
			}

			// Zero values of the results on a mode error.
			f := fraction.MustNew(-5, 2)
			n, err := f.RoundToInt(bad.mode)
			f1CheckModeErr(t, "RoundToInt", err, bad.sentinel)
			assert.Zero(t, n)
			bi, err := f.RoundToBigInt(bad.mode)
			f1CheckModeErr(t, "RoundToBigInt", err, bad.sentinel)
			assert.Nil(t, bi)
			r, err := f.RoundToDenominator(4, bad.mode)
			f1CheckModeErr(t, "RoundToDenominator", err, bad.sentinel)
			assert.Equal(t, fraction.Fraction{}, r)
			s, err := f.FormatDecimal(3, bad.mode)
			f1CheckModeErr(t, "FormatDecimal", err, bad.sentinel)
			assert.Empty(t, s)
			k, v, err := f.RoundToMultiple(fraction.One(), bad.mode)
			f1CheckModeErr(t, "RoundToMultiple", err, bad.sentinel)
			assert.Zero(t, k)
			assert.Equal(t, fraction.Fraction{}, v)
			kb, vb, err := f.RoundToMultipleBig(fraction.One(), bad.mode)
			f1CheckModeErr(t, "RoundToMultipleBig", err, bad.sentinel)
			assert.Nil(t, kb)
			assert.Equal(t, fraction.Fraction{}, vb)

			// Decimal keeps RoundHalfAwayFromZero and is not affected.
			d, err := f.Decimal(0)
			require.NoError(t, err)
			assert.Equal(t, "-3", d)
		})
	}
}

// TestRoundingModeErrorPriority checks section 9: ErrInvalid and
// ErrZeroDenominator / ErrDivisionByZero (step 2) precede the mode error
// (step 3), which precedes the other checks of step 3 and step 4.
func TestRoundingModeErrorPriority(t *testing.T) {
	t.Parallel()

	maxC := fraction.MaxComponentBig()
	var invalid fraction.Fraction
	f := fraction.MustNew(1, 3)
	for _, bad := range f1BadModes {
		t.Run(bad.name, func(t *testing.T) {
			t.Parallel()
			mode := bad.mode

			// Step 1 first.
			for _, c := range f1RoundCalls {
				err := c.call(invalid, mode)
				wCheckErr(t, c.name+" on an invalid receiver", err, fraction.ErrInvalid)
			}
			_, _, err := f.RoundToMultiple(invalid, mode)
			wCheckErr(t, "RoundToMultiple with an invalid step", err, fraction.ErrInvalid)
			_, _, err = f.RoundToMultipleBig(invalid, mode)
			wCheckErr(t, "RoundToMultipleBig with an invalid step", err, fraction.ErrInvalid)
			_, err = invalid.RoundToDenominator(0, mode)
			wCheckErr(t, "RoundToDenominator invalid and den 0", err, fraction.ErrInvalid)

			// Step 2 before the mode error.
			for _, den := range []int{0, -1, math.MinInt} {
				_, err := f.RoundToDenominator(den, mode)
				wCheckErr(t, fmt.Sprintf("RoundToDenominator(%d)", den), err, fraction.ErrZeroDenominator)
				f1CheckMustSame(t, "MustRoundToDenominator", func() { f.MustRoundToDenominator(den, mode) }, err)
			}
			for _, den := range []*big.Int{nil, big.NewInt(0), big.NewInt(-7)} {
				_, err := f.RoundToBigDenominator(den, mode)
				wCheckErr(t, fmt.Sprintf("RoundToBigDenominator(%v)", den), err, fraction.ErrZeroDenominator)
			}
			_, _, err = f.RoundToMultiple(fraction.Zero(), mode)
			wCheckErr(t, "RoundToMultiple step 0", err, fraction.ErrDivisionByZero)
			_, _, err = f.RoundToMultipleBig(fraction.Zero(), mode)
			wCheckErr(t, "RoundToMultipleBig step 0", err, fraction.ErrDivisionByZero)
			_, _, err = f.RoundToMultiple(fraction.MustNew(-1, 2), mode)
			wCheckErr(t, "RoundToMultiple step < 0", err, fraction.ErrZeroDenominator)
			_, _, err = f.RoundToMultipleBig(fraction.MustNew(1, -2), mode)
			wCheckErr(t, "RoundToMultipleBig step < 0", err, fraction.ErrZeroDenominator)

			// The mode error before the other checks of steps 3 and 4.
			for _, digits := range []int{-1, fraction.MaxDigits + 1, math.MinInt, math.MaxInt} {
				_, err := f.FormatDecimal(digits, mode)
				f1CheckModeErr(t, fmt.Sprintf("FormatDecimal(%d)", digits), err, bad.sentinel)
			}
			_, err = f.RoundToBigDenominator(wOff(maxC, 1), mode)
			f1CheckModeErr(t, "RoundToBigDenominator(den > MaxComponent)", err, bad.sentinel)
			_, err = wFrac(maxC, big.NewInt(1)).RoundToDenominator(math.MaxInt, mode)
			f1CheckModeErr(t, "RoundToDenominator with m beyond the range", err, bad.sentinel)
			_, err = wFrac(wPow2(100), big.NewInt(1)).RoundToInt(mode)
			f1CheckModeErr(t, "RoundToInt beyond int", err, bad.sentinel)
			_, _, err = wFrac(wPow2(100), big.NewInt(1)).RoundToMultiple(fraction.One(), mode)
			f1CheckModeErr(t, "RoundToMultiple with k beyond int", err, bad.sentinel)
			_, _, err = wFrac(maxC, big.NewInt(1)).RoundToMultipleBig(fraction.MustNew(2, 1), mode)
			f1CheckModeErr(t, "RoundToMultipleBig with v beyond the range", err, bad.sentinel)
		})
	}

	// With a valid mode the digits check of FormatDecimal stays ErrNotRepresentable.
	_, err := f.FormatDecimal(-1, fraction.RoundHalfUp)
	wCheckErr(t, "FormatDecimal(-1)", err, fraction.ErrNotRepresentable)
	assert.False(t, errors.Is(err, rounding.ErrUnset) || errors.Is(err, rounding.ErrUnknownMode),
		"a digits error is not a mode error")
}

// TestDivInDenominatorErrorOrder checks that a zero divisor is reported before
// a non-positive target denominator.
func TestDivInDenominatorErrorOrder(t *testing.T) {
	t.Parallel()

	receivers := []fraction.Fraction{fraction.One(), fraction.MustNew(-3, 4), fraction.MustNew(0, -2)}
	zeros := []fraction.Fraction{fraction.Zero(), fraction.MustNew(0, -3)}
	for _, f := range receivers {
		for _, z := range zeros {
			for _, den := range []int{0, -3, math.MinInt} {
				label := fmt.Sprintf("%s.DivInDenominator(%s, %d)", f, z, den)
				got, err := f.DivInDenominator(z, den)
				wCheckErr(t, label, err, fraction.ErrDivisionByZero)
				assert.Equalf(t, fraction.Fraction{}, got, "%s", label)
				f1CheckMustSame(t, "Must"+label, func() { f.MustDivInDenominator(z, den) }, err)
			}
		}
	}

	// Step 1 still comes first.
	var invalid fraction.Fraction
	_, err := invalid.DivInDenominator(fraction.Zero(), 0)
	wCheckErr(t, "invalid receiver", err, fraction.ErrInvalid)
	_, err = fraction.One().DivInDenominator(invalid, -3)
	wCheckErr(t, "invalid divisor", err, fraction.ErrInvalid)
}

// TestZeroAllocsRounding checks T-2 for F-1: the narrow RoundToMultiple in the
// classes w31, alt, w63, and the rounding methods of the list in every mode,
// run without heap allocations on the benchmark inputs. It is not parallel:
// AllocsPerRun measures process-wide allocation counters.
func TestZeroAllocsRounding(t *testing.T) {
	for _, m := range f1Modes {
		mode := m.mode
		for _, in := range benchInputs(benchSplit) {
			t.Run("RoundToInt/"+m.name+"/"+in.class, func(t *testing.T) {
				_, err := in.x.RoundToInt(mode)
				require.NoError(t, err)
				assert.Zero(t, testing.AllocsPerRun(100, func() { _, _ = in.x.RoundToInt(mode) }))
			})
		}
		for _, in := range benchInputs(benchRoundDen) {
			t.Run("RoundToDenominator/"+m.name+"/"+in.class, func(t *testing.T) {
				_, err := in.x.RoundToDenominator(in.den, mode)
				require.NoError(t, err)
				assert.Zero(t, testing.AllocsPerRun(100, func() { _, _ = in.x.RoundToDenominator(in.den, mode) }))
			})
		}
		for _, in := range benchInputs(benchDivMod) {
			if c := in.baseClass(); c != "w31" && c != "alt" && c != "w63" {
				continue
			}
			t.Run("RoundToMultiple/"+m.name+"/"+in.class, func(t *testing.T) {
				_, _, err := in.x.RoundToMultiple(in.y, mode)
				require.NoError(t, err)
				assert.Zero(t, testing.AllocsPerRun(100, func() { _, _, _ = in.x.RoundToMultiple(in.y, mode) }))
			})
		}
	}
}
