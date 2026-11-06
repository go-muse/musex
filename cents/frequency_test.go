package cents_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-18, NF-3 p. 4: the frequency multiplier 2^(c/1200). The exact result is
// rounded once to 53 bits without a limit on the exponent (ties to even);
// a rounded result below 2^-1022 or at least 2^1024 is ErrOutOfRange.
func TestMultiplier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		c    cents.Cents
		want float64
		err  error
	}{
		{name: "0", c: cents.Zero(), want: 1},
		{name: "<1200>", c: cents.Octave(), want: 2},
		{name: "<-1200>", c: noErr(cents.FromInt(-1200))(t), want: 0.5},
		{name: "<2400>", c: noErr(cents.FromInt(2400))(t), want: 4},
		// 2^(7/12): Python decimal, 120 digits, nearest float64.
		{name: "<700>", c: noErr(cents.FromInt(700))(t), want: 1.4983070768766815},
		// 2^(1/12): Python decimal, 120 digits, nearest float64.
		{name: "<100>", c: cents.Semitone(), want: 1.0594630943592953},
		// 5^(1/4): Python decimal, 120 digits, nearest float64.
		{name: "Q", c: byName(t, setS(t), "Q").v, want: 1.4953487812212205},
		{name: "R(3:2)", c: noErr(cents.FromRatio(3, 2))(t), want: 1.5},
		{name: "R(5:4)", c: noErr(cents.FromRatio(5, 4))(t), want: 1.25},
		{name: "<-1226400>", c: noErr(cents.FromInt(-1226400))(t), want: 0x1p-1022},
		{name: "<1227600>", c: noErr(cents.FromInt(1227600))(t), want: 0x1p1023},
		// 3^34/2^53: exactly between two float64; the tie goes to the even
		// mantissa (Python fractions, float(Fraction) rounds correctly).
		{name: "midpoint down to even", c: noErr(cents.FromRatio(16677181699666569, 1<<53))(t), want: 0x1.d9fe779881944p+0},
		{name: "<1300000>", c: noErr(cents.FromInt(1300000))(t), err: cents.ErrOutOfRange},
		{name: "<-1300000>", c: noErr(cents.FromInt(-1300000))(t), err: cents.ErrOutOfRange},
		{name: "<1228800>", c: noErr(cents.FromInt(1228800))(t), err: cents.ErrOutOfRange},
		{name: "<-1226401>", c: noErr(cents.FromInt(-1226401))(t), err: cents.ErrOutOfRange},
		{name: "<9000000>", c: noErr(cents.FromInt(9000000))(t), err: cents.ErrOutOfRange},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.c.Multiplier()
			assertMust(t, got, err, tc.c.MustMultiplier, "MustMultiplier")

			if tc.err != nil {
				assertOnlyError(t, err, tc.err)
				assert.Zero(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got, "%x", got)
			assert.InEpsilon(t, tc.want, got, 1e-12)
		})
	}

	t.Run("largest finite", func(t *testing.T) {
		t.Parallel()

		got, err := noErr(cents.FromInt(1228799))(t).Multiplier()
		require.NoError(t, err)
		assert.False(t, math.IsInf(got, 0))
		assert.Greater(t, got, 0x1p1023)
	})

	// Multiples of 1200 from -1226400 to 1227600 give 2^k exactly.
	t.Run("powers of two", func(t *testing.T) {
		t.Parallel()

		for _, k := range []int{-1022, -1021, -500, -1, 0, 1, 2, 52, 53, 500, 1022, 1023} {
			c := noErr(cents.FromInt(1200*k))(t, "<1200*%d>", k)
			got, err := c.Multiplier()
			require.NoError(t, err, "k = %d", k)
			assertMust(t, got, err, c.MustMultiplier, "MustMultiplier: k = %d", k)
			assert.Equal(t, math.Ldexp(1, k), got, "k = %d", k)
		}
	})
}

// K-18: applying cents to a frequency, f * 2^(c/1200), rounded once.
func TestApplyTo(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		c    cents.Cents
		freq float64
		want float64
		err  error
	}{
		{name: "440 by 0", c: cents.Zero(), freq: 440, want: 440},
		{name: "440 by <1200>", c: cents.Octave(), freq: 440, want: 880},
		{name: "440 by R(442:440)", c: noErr(cents.FromRatio(442, 440))(t), freq: 440, want: 442},
		// 440 * 2^(7/12): Python decimal, 120 digits, nearest float64.
		{name: "440 by <700>", c: noErr(cents.FromInt(700))(t), freq: 440, want: 659.2551138257398},
		{name: "1e300 by <120000>", c: noErr(cents.FromInt(120000))(t), freq: 1e300, err: cents.ErrOutOfRange},
		{name: "1e-300 by <-120000>", c: noErr(cents.FromInt(-120000))(t), freq: 1e-300, err: cents.ErrOutOfRange},
		{name: "2^-1022 by 0", c: cents.Zero(), freq: 0x1p-1022, want: 0x1p-1022},
		{name: "2^-1022 by <-1200>", c: noErr(cents.FromInt(-1200))(t), freq: 0x1p-1022, err: cents.ErrOutOfRange},
		{name: "2^-1074 by <62400>", c: noErr(cents.FromInt(62400))(t), freq: 0x1p-1074, want: 0x1p-1022},
		{name: "2^-1074 by 0", c: cents.Zero(), freq: 0x1p-1074, err: cents.ErrOutOfRange},
		{name: "MaxFloat64 by <1200>", c: cents.Octave(), freq: math.MaxFloat64, err: cents.ErrOutOfRange},
		{name: "MaxFloat64 by <-1200>", c: noErr(cents.FromInt(-1200))(t), freq: math.MaxFloat64, want: math.MaxFloat64 / 2},
		// 3 * (1 + 2^-52) = 3 + 1.5*2^-51 is a tie; to the even mantissa
		// (Python fractions).
		{name: "midpoint up to even", c: noErr(cents.FromRatio(3, 1))(t), freq: 1 + 0x1p-52, want: 0x1.8000000000002p+1},
		// The rounded result decides (Python decimal, 80 digits: the bands
		// near 2^1024 and 2^-1022 are 9.61e-14 cent wide).
		{name: "MaxFloat64 by 1e-13", c: noErr(cents.FromFraction(1, 10_000_000_000_000))(t), freq: math.MaxFloat64, err: cents.ErrOutOfRange},
		{name: "MaxFloat64 by 1e-14", c: noErr(cents.FromFraction(1, 100_000_000_000_000))(t), freq: math.MaxFloat64, want: math.MaxFloat64},
		{name: "2^-1022 by -1e-13", c: noErr(cents.FromFraction(-1, 10_000_000_000_000))(t), freq: 0x1p-1022, err: cents.ErrOutOfRange},
		{name: "2^-1022 by -1e-14", c: noErr(cents.FromFraction(-1, 100_000_000_000_000))(t), freq: 0x1p-1022, want: 0x1p-1022},
		{name: "freq 0", c: cents.Zero(), freq: 0, err: cents.ErrInvalidArgument},
		{name: "freq -0", c: cents.Zero(), freq: math.Copysign(0, -1), err: cents.ErrInvalidArgument},
		{name: "freq -440", c: cents.Octave(), freq: -440, err: cents.ErrInvalidArgument},
		{name: "freq NaN", c: cents.Octave(), freq: math.NaN(), err: cents.ErrInvalidArgument},
		{name: "freq +Inf", c: cents.Octave(), freq: math.Inf(1), err: cents.ErrInvalidArgument},
		{name: "freq -Inf", c: cents.Octave(), freq: math.Inf(-1), err: cents.ErrInvalidArgument},
		{name: "invalid before range", c: noErr(cents.FromInt(9000000))(t), freq: -1, err: cents.ErrInvalidArgument},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.c.ApplyTo(tc.freq)
			assertMust(t, got, err, func() float64 { return tc.c.MustApplyTo(tc.freq) }, "MustApplyTo")

			if tc.err != nil {
				assertOnlyError(t, err, tc.err)
				assert.Zero(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got, "%x", got)
		})
	}

	// The input frequency is the exact binary value of the float64 (not its
	// shortest decimal): by 0 the frequency is returned unchanged, by
	// multiples of 1200 it is scaled by a power of two exactly.
	t.Run("binary frequency", func(t *testing.T) {
		t.Parallel()

		freqs := []float64{0.1, 1.0 / 3.0, 440, 261.6255653005986, 1e-300, 0x1p-1022, math.MaxFloat64, 0x1.fffffffffffffp-2}
		for _, f := range freqs {
			got, err := cents.Zero().ApplyTo(f)
			require.NoError(t, err)
			assertMust(t, got, err, func() float64 { return cents.Zero().MustApplyTo(f) }, "MustApplyTo(%x)", f)
			assert.Equal(t, f, got, strconv.FormatFloat(f, 'x', -1, 64))

			if f < 0x1p1000 {
				got, err = noErr(cents.FromInt(2400))(t).ApplyTo(f)
				require.NoError(t, err)
				assert.Equal(t, f*4, got, strconv.FormatFloat(f, 'x', -1, 64))
			}
		}
	})
}
