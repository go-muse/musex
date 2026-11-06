package cents_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// floatCase is a float64 input with the expected value or error.
type floatCase struct {
	name string
	x    float64
	want comp
	err  error
}

// K-3, G-6, G-12: values from float64 cents. The input is its shortest
// decimal (strconv.FormatFloat 'g' -1) as an exact rational; the value is
// the nearest multiple of 10^-6, half away from zero.
func TestFromFloat(t *testing.T) {
	t.Parallel()

	cases := []floatCase{
		{name: "7.1", x: 7.1, want: ratC(71, 10)},
		{name: "0.1", x: 0.1, want: ratC(1, 10)},
		{name: "-13.7", x: -13.7, want: ratC(-137, 10)},
		{name: "701.955001", x: 701.955001, want: ratC(701955001, 1000000)},
		{name: "700.0", x: 700.0, want: ratC(700, 1)},
		{name: "-0.0", x: math.Copysign(0, -1), want: zeroComp()},
		{name: "1e-7", x: 1e-7, want: zeroComp()},
		{name: "6e-7", x: 6e-7, want: ratC(1, 1000000)},
		{name: "1.23456789", x: 1.23456789, want: ratC(1234568, 1000000)},
		{name: "9000000.0", x: 9000000.0, want: ratC(9000000, 1)},
		{name: "-9000000.0", x: -9000000.0, want: ratC(-9000000, 1)},
		{name: "5e-7", x: 5e-7, want: ratC(1, 1000000)},
		{name: "-5e-7", x: -5e-7, want: ratC(-1, 1000000)},
		{name: "1.5e-6", x: 1.5e-6, want: ratC(2, 1000000)},
		{name: "2.5e-6", x: 2.5e-6, want: ratC(3, 1000000)},
		{name: "4.9999999e-7", x: 4.9999999e-7, want: zeroComp()},
		{name: "0.1234565", x: 0.1234565, want: ratC(123457, 1000000)},
		{name: "-0.1234565", x: -0.1234565, want: ratC(-123457, 1000000)},
		{name: "smallest subnormal", x: math.SmallestNonzeroFloat64, want: zeroComp()},
		{name: "8999999.9999995", x: 8999999.9999995, want: ratC(9000000, 1)},
		{name: "9000001.0", x: 9000001.0, err: cents.ErrOutOfRange},
		{name: "9000000.0000005", x: 9000000.0000005, err: cents.ErrOutOfRange},
		{name: "-9000000.0000005", x: -9000000.0000005, err: cents.ErrOutOfRange},
		{name: "MaxFloat64", x: math.MaxFloat64, err: cents.ErrOutOfRange},
		{name: "NaN", x: math.NaN(), err: cents.ErrInvalidArgument},
		{name: "+Inf", x: math.Inf(1), err: cents.ErrInvalidArgument},
		{name: "-Inf", x: math.Inf(-1), err: cents.ErrInvalidArgument},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := cents.FromFloat(tc.x)
			assertResult(t, tc.want, tc.err, got, err)
			assertMust(t, got, err, func() cents.Cents { return cents.MustFromFloat(tc.x) }, "MustFromFloat")

			again, err2 := cents.FromFloat(tc.x)
			assert.True(t, got == again, "the same input gives an equal value")
			assert.Equal(t, errName(err), errName(err2))

			if tc.err == nil && !math.IsNaN(tc.x) {
				// The oracle of the rule agrees with the literal of the case.
				assert.Equal(t, tc.want.String(), gridOracle(tc.x).String())
			}
		})
	}

	t.Run("grid oracle", func(t *testing.T) {
		t.Parallel()

		inputs := []float64{
			0.000001, 0.0000005, 0.0000015, 0.00000049999999, 1.0000005, -1.0000005,
			123.456789, 123.4567895, -123.4567895, 0.3333333333333333, 2.0 / 3.0,
			701.9550008653874, 1e-300, -1e-300, 4500000.123456789, 8999999.999999,
			8999999.9999994, -8999999.9999996, 0x1p-1022, 1e6 + 0.5e-6,
		}

		for _, x := range inputs {
			want := gridOracle(x)
			got, err := cents.FromFloat(x)
			assertResult(t, want, want.wantErr(), got, err, strconv.FormatFloat(x, 'g', -1, 64))
			assertMust(t, got, err, func() cents.Cents { return cents.MustFromFloat(x) }, "MustFromFloat(%v)", x)
		}
	})
}

// K-6: values of a ratio given as float64, on the grid 10^-6.
func TestFromRatioFloat(t *testing.T) {
	t.Parallel()

	cases := []floatCase{
		// 1200*log2(1.5) = 701.95500086538...: Python decimal, 120 digits.
		{name: "1.5", x: 1.5, want: ratC(701955001, 1000000)},
		// 1200*log2(1.1) = 165.0042284999...: Python decimal, 120 digits.
		{name: "1.1", x: 1.1, want: ratC(165004228, 1000000)},
		{name: "2.0", x: 2.0, want: ratC(1200, 1)},
		{name: "0.5", x: 0.5, want: ratC(-1200, 1)},
		{name: "1.0", x: 1.0, want: zeroComp()},
		{name: "4.0", x: 4.0, want: ratC(2400, 1)},
		// 1200*log2(5e-324) = -1288779.329578...: Python decimal, 120 digits.
		{name: "smallest subnormal", x: math.SmallestNonzeroFloat64, want: ratC(-1288779329578, 1000000)},
		// 1200*log2(1.7976931348623157e308) = 1228800 - 1.922e-13: Python
		// decimal, 100 digits.
		{name: "MaxFloat64", x: math.MaxFloat64, want: ratC(1228800, 1)},
		// 1200*log2(1e300) = 1195894.114159...: Python decimal, 120 digits.
		{name: "1e300", x: 1e300, want: ratC(1195894114159, 1000000)},
		{name: "0", x: 0, err: cents.ErrInvalidArgument},
		{name: "-0", x: math.Copysign(0, -1), err: cents.ErrInvalidArgument},
		{name: "-1", x: -1, err: cents.ErrInvalidArgument},
		{name: "NaN", x: math.NaN(), err: cents.ErrInvalidArgument},
		{name: "+Inf", x: math.Inf(1), err: cents.ErrInvalidArgument},
		{name: "-Inf", x: math.Inf(-1), err: cents.ErrInvalidArgument},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := cents.FromRatioFloat(tc.x)
			assertResult(t, tc.want, tc.err, got, err)
			assertMust(t, got, err, func() cents.Cents { return cents.MustFromRatioFloat(tc.x) }, "MustFromRatioFloat")

			again, _ := cents.FromRatioFloat(tc.x)
			assert.True(t, got == again, "the same input gives an equal value")
		})
	}

	t.Run("not the exact ratio", func(t *testing.T) {
		t.Parallel()

		assert.False(t, noErr(cents.FromRatioFloat(1.5))(t).Equal(noErr(cents.FromRatio(3, 2))(t)))
		assert.Equal(t, "701.955001", noErr(cents.FromRatioFloat(1.5))(t).String())
	})
}

// K-7, G-6: the distance between two frequencies, on the grid 10^-6.
func TestBetween(t *testing.T) {
	t.Parallel()

	type pair struct {
		name   string
		f1, f2 float64
		want   comp
		err    error
	}

	cases := []pair{
		// 1200*log2(442/440) = 7.8514...: Python decimal, 120 digits.
		{name: "440 to 442", f1: 440, f2: 442, want: ratC(7851415, 1000000)},
		{name: "442 to 440", f1: 442, f2: 440, want: ratC(-7851415, 1000000)},
		{name: "440 to 880", f1: 440, f2: 880, want: ratC(1200, 1)},
		{name: "440 to 440", f1: 440, f2: 440, want: zeroComp()},
		// 1200*log2(1e600) = 2391788.2283189...: Python decimal, 120 digits.
		{name: "1e-300 to 1e300", f1: 1e-300, f2: 1e300, want: ratC(2391788228319, 1000000)},
		// 1200*log2(1.7976931348623157e308 / 5e-324) = 2517579.3295778...:
		// Python decimal, 120 digits.
		{name: "smallest to largest", f1: math.SmallestNonzeroFloat64, f2: math.MaxFloat64, want: ratC(2517579329578, 1000000)},
		{name: "largest to smallest", f1: math.MaxFloat64, f2: math.SmallestNonzeroFloat64, want: ratC(-2517579329578, 1000000)},
	}

	invalid := []float64{0, math.Copysign(0, -1), -440, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, x := range invalid {
		cases = append(cases,
			pair{name: "f1 " + strconv.FormatFloat(x, 'g', -1, 64), f1: x, f2: 440, err: cents.ErrInvalidArgument},
			pair{name: "f2 " + strconv.FormatFloat(x, 'g', -1, 64), f1: 440, f2: x, err: cents.ErrInvalidArgument},
		)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := cents.Between(tc.f1, tc.f2)
			assertResult(t, tc.want, tc.err, got, err)
			assertMust(t, got, err, func() cents.Cents { return cents.MustBetween(tc.f1, tc.f2) }, "MustBetween")
		})
	}

	// Subnormal inputs are valid (G-6): no error, and the distance back is
	// the negated distance.
	t.Run("subnormal inputs", func(t *testing.T) {
		t.Parallel()

		for _, f := range []float64{0x1p-1074, 0x1p-1050, 0x1p-1023} {
			got, err := cents.Between(f, 0x1p-1022)
			require.NoError(t, err)
			assertMust(t, got, err, func() cents.Cents { return cents.MustBetween(f, 0x1p-1022) }, "MustBetween(%x)", f)

			back, err := cents.Between(0x1p-1022, f)
			require.NoError(t, err)
			assertCents(t, got.Neg(), back)
			assert.Equal(t, -1, back.Sign())
		}
	})

	t.Run("decimal view", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "2391788.228319", noErr(cents.Between(1e-300, 1e300))(t).Decimal(6, 0))
		assert.Equal(t, "7.85", noErr(cents.Between(440, 442))(t).Decimal(2, 0))
	})
}
