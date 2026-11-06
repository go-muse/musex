package cents_test

import (
	"errors"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertExactlyOneError checks that err is exactly one of the five errors.
func assertExactlyOneError(t testing.TB, err error, msgAndArgs ...any) bool {
	t.Helper()

	n := 0
	for _, e := range allErrors() {
		if errors.Is(err, e) {
			n++
		}
	}

	return assert.Equal(t, 1, n, msgAndArgs...)
}

// textSeeds are texts of the criteria and the architecture.
func textSeeds() []string {
	return []string{
		"30", "+30", "-30", "−30", "701.955", "0.1", "+30 ¢", "+30¢", "30 cents", "30 Cents", "30 cents",
		"1 cent", "-0", "1200/19 ¢", "-1200/19 cents", "3:2", "6:4", "2:3", "5^(1/4)", "10^(1/1000)",
		"2^-84*3^53", "2^(-1199/1200)*3", "2^0", "9000000", "-9000000 cents", "9000000.000001", "37:32",
		"83:88", "0:1", "3:0", "1/0 ¢", "3/2", "408.", ".5", " 30", "30 ", "30  cents", "1e3", "2^(1/0)",
		"0." + strings.Repeat("0", 30) + "1", strings.Repeat("9", 101), "37" + strings.Repeat("0", 100) + ":1",
		repeatFactors("2", "2", 63), repeatFactors("2", "2", 64), "", "¢", "cents", "3:2^2",
		"19383245667680019896796723:19342813113834066795298816",
		"2^(1/2)*3^(1/2)*5^(1/2)*7^(1/2)*11^(1/2)*13^(1/2)*17^(1/2)*19^(1/2)*23^(1/2)*29^(1/2)*31^(1/2)",
	}
}

// FuzzParse: no panic; an error is exactly one of the five and gives the
// zero value; a parsed value reads back from its canonical text, which is a
// fixed point, and from every available explicit view.
func FuzzParse(f *testing.F) {
	for _, s := range textSeeds() {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, text string) {
		c, err := cents.Parse(text)
		if err != nil {
			assertExactlyOneError(t, err, "%q", text)
			assert.True(t, c == cents.Cents{}, "%q", text)

			return
		}

		s := c.String()
		back, err := cents.Parse(s)
		require.NoError(t, err, "Parse(String()) of %q", text)
		assert.True(t, back == c, "Parse(String()) of %q", text)
		assert.Equal(t, s, back.String(), "String() is a fixed point")

		for _, opts := range optionSets() {
			for _, k := range kinds() {
				view, ok := c.TextAs(k, opts)
				if !ok {
					continue
				}

				v, err := cents.Parse(view)
				require.NoError(t, err, "%q", view)
				assert.True(t, v == c, "TextAs(%s, %d) = %q of %q", k, opts, view, text)
			}
		}
	})
}

// FuzzParseScala: no panic; the rest is a suffix of the line starting with a
// space or a tab, or empty; a ratio with n, d up to 2^31-1 reads back from
// Scala(5); ErrPrimeLimit has an approximate value.
func FuzzParseScala(f *testing.F) {
	for _, s := range []string{
		"408.", "  100.0 cents", "-5.0", "3/2", "5", "10/20", " 5/4   E", "\t81/80", "-3/2", "0/1", "3/0",
		"+5.0", "−5.0", "5/4E", "408.0.5", "", "   ", "37/32", "5/4\r", "5/4 " + strings.Repeat("9", 100),
		strings.Repeat(" ", 100) + "5/4", "37" + strings.Repeat("0", 100) + "/1", "0." + strings.Repeat("0", 30) + "1",
		"9000000.000001", "1162261467/1", "3486784401/1",
	} {
		f.Add(s)
	}

	const limit = 1<<31 - 1

	f.Fuzz(func(t *testing.T, line string) {
		c, rest, err := cents.ParseScala(line)
		if err != nil {
			assertExactlyOneError(t, err, "%q", line)
			assert.True(t, c == cents.Cents{}, "%q", line)
			assert.Empty(t, rest, "%q", line)

			if errors.Is(err, cents.ErrPrimeLimit) {
				_, _, errApprox := cents.ParseScalaApprox(line)
				assert.NoError(t, errApprox, "approximate value of %q", line)
			}

			return
		}

		assert.True(t, strings.HasSuffix(line, rest), "rest %q of %q", rest, line)
		if rest != "" {
			assert.Contains(t, []byte{' ', '\t'}, rest[0], "rest %q of %q", rest, line)
		}

		approx, approxRest, err := cents.ParseScalaApprox(line)
		require.NoError(t, err, "%q", line)
		assert.True(t, approx == c, "%q", line)
		assert.Equal(t, rest, approxRest)

		if n, d, ok := c.Ratio(); ok && n <= limit && d <= limit {
			back, backRest, err := cents.ParseScala(c.Scala(5))
			require.NoError(t, err, "%q", c.Scala(5))
			assert.True(t, back == c, "Scala(5) of %q", line)
			assert.Empty(t, backRest)
		}
	})
}

// small reduces a fuzzed number to |x| < m (m > 0), keeping 0.
func small(x, m int64) int64 {
	return x % m
}

// FuzzComposition: FromComposition with small integers agrees with the
// record oracle; Composition and FromComposition round trip.
func FuzzComposition(f *testing.F) {
	f.Add(int64(0), int64(1), int64(0), int64(1), int64(0), int64(1))
	f.Add(int64(-1200), int64(1), int64(1), int64(1), int64(0), int64(1))
	f.Add(int64(6), int64(5), int64(0), int64(1), int64(1), int64(1000))
	f.Add(int64(-100800), int64(1), int64(53), int64(1), int64(0), int64(1))
	f.Add(int64(0), int64(1), int64(6000), int64(1), int64(0), int64(1))
	f.Add(int64(9000000), int64(1), int64(1), int64(1000000), int64(0), int64(1))
	f.Add(int64(1), int64(0), int64(1), int64(2), int64(1), int64(3))
	f.Add(int64(0), int64(1), int64(1), int64(0), int64(1), int64(1))
	f.Add(int64(math.MaxInt64), int64(math.MinInt64), int64(-7), int64(-3), int64(5), int64(-4))

	f.Fuzz(func(t *testing.T, r, s, e3, d3, e5, d5 int64) {
		e3, d3, e5, d5 = small(e3, 1<<20), small(d3, 1<<20), small(e5, 1<<20), small(d5, 1<<20)

		got, err := cents.FromComposition(big.NewInt(r), big.NewInt(s),
			cents.Power{Base: 3, Num: big.NewInt(e3), Den: big.NewInt(d3)},
			cents.Power{Base: 5, Num: big.NewInt(e5), Den: big.NewInt(d5)})

		if s == 0 || d3 == 0 || d5 == 0 {
			assertResult(t, comp{}, cents.ErrInvalidArgument, got, err)

			return
		}

		want := bigRatC(new(big.Rat).SetFrac(big.NewInt(r), big.NewInt(s))).
			withExp(3, new(big.Rat).SetFrac64(e3, d3)).withExp(5, new(big.Rat).SetFrac64(e5, d5))
		if !assertResult(t, want, want.wantErr(), got, err, want.String()) || err != nil {
			return
		}

		num, den, ps := got.Composition()
		back, err := cents.FromComposition(num, den, ps...)
		require.NoError(t, err)
		assert.True(t, back == got, want.String())
	})
}

// FuzzArithmetic: Add, Sub, Mul and Div of values made of small integers
// agree with the record oracle (value or error); a + b - b = a.
func FuzzArithmetic(f *testing.F) {
	f.Add(int64(-1200), int64(1), int64(1), int64(0), int64(1), int64(4), int64(12), int64(53))
	f.Add(int64(9000000), int64(1), int64(0), int64(1), int64(1000000), int64(0), int64(2), int64(3))
	f.Add(int64(-9000000), int64(1), int64(0), int64(1200), int64(1), int64(0), int64(7501), int64(-1))
	f.Add(int64(71), int64(10), int64(0), int64(-137), int64(10), int64(0), int64(-4), int64(1000000))
	f.Add(int64(0), int64(1), int64(53), int64(-100800), int64(1), int64(0), int64(math.MaxInt64), int64(0))

	f.Fuzz(func(t *testing.T, r1, s1, e1, r2, s2, e2, k, n int64) {
		if s1 == 0 {
			s1 = 1
		}

		if s2 == 0 {
			s2 = 1
		}

		ac := bigRatC(new(big.Rat).SetFrac64(r1, s1)).withExp(3, rat(small(e1, 1<<16), 7))
		bc := bigRatC(new(big.Rat).SetFrac64(r2, s2)).withExp(5, rat(small(e2, 1<<16), 12))
		if ac.wantErr() != nil || bc.wantErr() != nil {
			return
		}

		a, b := fromComp(t, ac), fromComp(t, bc)

		sum, err := a.Add(b)
		assertResult(t, ac.add(bc), ac.add(bc).wantErr(), sum, err, "a + b")

		if err == nil {
			assertCents(t, a, noErr(sum.Sub(b))(t), "a + b - b")
		}

		diff, err := a.Sub(b)
		assertResult(t, ac.sub(bc), ac.sub(bc).wantErr(), diff, err, "a - b")

		prod, err := a.Mul(int(k))
		assertResult(t, ac.mulInt(int(k)), ac.mulInt(int(k)).wantErr(), prod, err, "a * k")

		quo, err := a.Div(int(n))
		if n == 0 {
			assertResult(t, comp{}, cents.ErrInvalidArgument, quo, err, "a / 0")
		} else {
			want := ac.scale(new(big.Rat).SetFrac64(1, n))
			assertResult(t, want, want.wantErr(), quo, err, "a / n")
		}
	})
}

// FuzzFromFloat: no panic; NaN and infinities are invalid; finite values
// agree with the exact grid oracle (shortest decimal, grid 10^-6, half away
// from zero) or give ErrOutOfRange; the same input gives an equal value.
func FuzzFromFloat(f *testing.F) {
	for _, x := range []float64{
		7.1, 0.1, -13.7, 701.955001, 700, 0, 1e-7, 6e-7, 1.23456789, 9000000, 9000001, 5e-7, -5e-7, 1.5e-6,
		2.5e-6, 4.9999999e-7, 0.1234565, -0.1234565, 9000000.0000005, 8999999.9999995, math.MaxFloat64,
		math.SmallestNonzeroFloat64, 0x1p-1022, math.NaN(), math.Inf(1), math.Inf(-1), math.Copysign(0, -1),
	} {
		f.Add(x)
	}

	f.Fuzz(func(t *testing.T, x float64) {
		c, err := cents.FromFloat(x)
		if math.IsNaN(x) || math.IsInf(x, 0) {
			assertResult(t, comp{}, cents.ErrInvalidArgument, c, err)

			return
		}

		want := gridOracle(x)
		assertResult(t, want, want.wantErr(), c, err, "%v", x)

		if math.Abs(x) <= cents.MaxMagnitude {
			require.NoError(t, err, "%v", x)
		}

		again, err2 := cents.FromFloat(x)
		assert.True(t, again == c)
		assert.Equal(t, errName(err), errName(err2))
	})
}
