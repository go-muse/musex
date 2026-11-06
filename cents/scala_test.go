package cents_test

import (
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-26: the Scala pitch value, ASCII only: n/d for ratios with n, d up to
// 2^31-1, otherwise cents with a point and at least one digit after it.
func TestScala(t *testing.T) {
	t.Parallel()

	set := setS(t)

	cases := []struct {
		name   string
		c      cents.Cents
		digits int
		want   string
	}{
		{"R(3:2)", noErr(cents.FromRatio(3, 2))(t), 5, "3/2"},
		{"<1200>", cents.Octave(), 5, "2/1"},
		{"SYNT", cents.SyntonicComma(), 5, "81/80"},
		{"R(2:3)", noErr(cents.FromRatio(2, 3))(t), 5, "2/3"},
		{"0", cents.Zero(), 5, "1/1"},
		{"<700> 1", noErr(cents.FromInt(700))(t), 1, "700.0"},
		{"<700> 5", noErr(cents.FromInt(700))(t), 5, "700.00000"},
		{"<700> 0 as 1", noErr(cents.FromInt(700))(t), 0, "700.0"},
		{"<700> -3 as 1", noErr(cents.FromInt(700))(t), -3, "700.0"},
		{"E(1,19) 5", noErr(cents.Step(1, 19))(t), 5, "63.15789"},
		{"<-5> 1", noErr(cents.FromInt(-5))(t), 1, "-5.0"},
		// 1200*log2(3^53/2^84), 5 digits: Python decimal, 120 digits.
		{"MK 5", byName(t, set, "MK").v, 5, "3.61505"},
		// 300*log2(5), 5 digits: Python decimal, 120 digits.
		{"Q 5", byName(t, set, "Q").v, 5, "696.57843"},
		{"1/8 at 2", noErr(cents.FromFraction(1, 8))(t), 2, "0.13"},
		{"-1/8 at 2", noErr(cents.FromFraction(-1, 8))(t), 2, "-0.13"},
		{"-1/3000 at 1", noErr(cents.FromFraction(-1, 3000))(t), 1, "0.0"},
		// 3^19 = 1162261467 <= 2^31-1 < 3^20 = 3486784401.
		{"3^19:1", noErr(cents.FromRatio(1162261467, 1))(t), 1, "1162261467/1"},
		{"1:3^19", noErr(cents.FromRatio(1, 1162261467))(t), 1, "1/1162261467"},
		{"2^30:1", noErr(cents.FromRatio(1<<30, 1))(t), 1, "1073741824/1"},
		{"2^31:1", noErr(cents.FromRatio(1<<31, 1))(t), 1, "37200.0"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.c.Scala(tc.digits))
		})
	}

	t.Run("3^20:1", func(t *testing.T) {
		t.Parallel()

		c := noErr(cents.FromRatio(3486784401, 1))(t)
		got := c.Scala(3)
		assert.Equal(t, scalaOracle(ratio(3486784401, 1), 3), got)
		assert.NotContains(t, got, "/")
	})

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			assert.Equal(t, a.v.Scala(1), a.v.Scala(0), a.name)

			for _, digits := range []int{-1, 0, 1, 3, 5, 10} {
				got := a.v.Scala(digits)
				assert.Equal(t, scalaOracle(a.c, digits), got, "%s.Scala(%d)", a.name, digits)
				assert.True(t, isASCII(got), a.name)
				assert.NotContains(t, got, "+")
				assert.NotContains(t, got, " ")

				if !strings.Contains(got, "/") {
					point := strings.Index(got, ".")
					require.GreaterOrEqual(t, point, 0, "cents with a point: %q", got)
					assert.Less(t, point, len(got)-1, "a digit after the point: %q", got)
				}
			}
		}
	})
}

// scalaCase is a line with the expected value, rest or error.
type scalaCase struct {
	name string
	line string
	want comp
	rest string
	err  error
}

func runScalaCases(t *testing.T, parse func(string) (cents.Cents, string, error), must func(string) (cents.Cents, string), cases []scalaCase) {
	t.Helper()

	for _, tc := range cases {
		name := tc.name
		if name == "" {
			name = strconv.Quote(tc.line)
		}

		t.Run(name+" "+errName(tc.err), func(t *testing.T) {
			t.Parallel()

			got, rest, err := parse(tc.line)
			assertResult(t, tc.want, tc.err, got, err)
			assert.Equal(t, tc.rest, rest)
			assertMust(t, mustScala{got, rest}, err, func() mustScala {
				c, rest := must(tc.line)

				return mustScala{c, rest}
			}, "the Must variant")
		})
	}
}

// scalaCommonCases are the cases where ParseScala and ParseScalaApprox
// agree: values within the prime limit and errors other than ErrPrimeLimit.
func scalaCommonCases() []scalaCase {
	digits101 := "1" + strings.Repeat("0", 100)

	cases := []scalaCase{
		{line: "408.", want: ratC(408, 1)},
		{line: "  100.0 cents", want: ratC(100, 1), rest: " cents"},
		{line: "-5.0", want: ratC(-5, 1)},
		{line: "3/2", want: ratio(3, 2)},
		{line: "5", want: ratio(5, 1)},
		{line: "10/20", want: ratC(-1200, 1)},
		{line: " 5/4   E", want: ratio(5, 4), rest: "   E"},
		{line: "\t81/80", want: ratio(81, 80)},
		{line: ".5", want: ratC(1, 2)},
		{line: "-.5", want: ratC(-1, 2)},
		{line: "5.", want: ratC(5, 1)},
		{line: "0.0", want: zeroComp()},
		{line: "-0.0", want: zeroComp()},
		{line: "1/1", want: zeroComp()},
		{line: "1", want: zeroComp()},
		{line: "0007/0004", want: ratio(7, 4)},
		{line: "111/74", want: ratio(3, 2)},
		{line: "5/4\tcomment", want: ratio(5, 4), rest: "\tcomment"},
		{line: "5 /4", want: ratio(5, 1), rest: " /4"},
		{line: "700.0 ! a comment", want: ratC(700, 1), rest: " ! a comment"},
		{line: "9000000.0", want: ratC(9000000, 1)},
		{line: "-9000000.0", want: ratC(-9000000, 1)},
		{name: "long rest", line: "5/4 " + strings.Repeat("9", 1000), want: ratio(5, 4), rest: " " + strings.Repeat("9", 1000)},
		{name: "long leading spaces", line: strings.Repeat(" ", 1000) + "5/4", want: ratio(5, 4)},
		{name: "long leading tabs and rest", line: strings.Repeat("\t ", 500) + "3/2 " + strings.Repeat("x", 1000), want: ratio(3, 2), rest: " " + strings.Repeat("x", 1000)},
		{name: "100-digit ratio part", line: "1/" + strings.Repeat("0", 97) + "125", want: ratio(1, 125)},
		{name: "30 digits after the point", line: "0." + strings.Repeat("0", 29) + "1", want: bigRatC(new(big.Rat).SetFrac(big.NewInt(1), pow10(30)))},
	}

	invalid := []string{
		"-3/2", "0/1", "3/0", "-0/1", "+5.0", "−5.0", "5/4E", "408.0.5", "", "   ", "\t", ".", "-", "-.",
		"5/4\r", "5/4\n", "\r", "5/ 4", "/4", "5//4", "5/4-", "1e3", "0x10", "3:2", "5,0", "5/4/3", "5.0/4",
		"5/4.0", "--5.0", "abc", "５/４",
		// "^" and "*" are not in the Scala grammar (architecture "Scala mode").
		"2^1*37^0", "2*3", "3^2",
	}
	for _, s := range invalid {
		cases = append(cases, scalaCase{line: s, err: cents.ErrInvalidText})
	}

	return append(cases,
		scalaCase{name: "101-digit ratio part", line: digits101 + "/3", err: cents.ErrTooComplex},
		scalaCase{name: "101 digits after the point", line: "0." + strings.Repeat("0", 100) + "1", err: cents.ErrTooComplex},
		scalaCase{name: "31 digits after the point", line: "0." + strings.Repeat("0", 30) + "1", err: cents.ErrTooComplex},
		scalaCase{line: "9000000.000001", err: cents.ErrOutOfRange},
		scalaCase{line: "-9000000.000001", err: cents.ErrOutOfRange},
		scalaCase{name: "101-digit cents out of range", line: digits101 + ".0", err: cents.ErrTooComplex},
		scalaCase{name: "syntax before length", line: digits101 + "/3E", err: cents.ErrInvalidText},
	)
}

// K-27, G-9: ParseScala reads one pitch value and returns the rest.
func TestParseScala(t *testing.T) {
	t.Parallel()

	cases := append(scalaCommonCases(),
		scalaCase{line: "37/32", err: cents.ErrPrimeLimit},
		scalaCase{line: "74/1", err: cents.ErrPrimeLimit},
		scalaCase{line: "37", err: cents.ErrPrimeLimit},
		scalaCase{line: "37/32 x", err: cents.ErrPrimeLimit},
		// The length of numbers before primes (V-2, K-32).
		scalaCase{name: "37 and 100 zeros / 1", line: "37" + strings.Repeat("0", 100) + "/1", err: cents.ErrTooComplex},
	)

	runScalaCases(t, cents.ParseScala, cents.MustParseScala, cases)

	t.Run("5 is a ratio", func(t *testing.T) {
		t.Parallel()

		// 1200*log2(5), 6 digits: Python decimal, 120 digits.
		c, _, err := cents.ParseScala("5")
		require.NoError(t, err)
		assert.Equal(t, "2786.313714", c.Decimal(6, 0))
		assertComp(t, ratC(5, 1), noErr(cents.Parse("5"))(t), "5 cents in the general mode")
	})

	// For the values of K-26: the Scala output reads back as the value (a
	// ratio) or as the printed decimal number (cents).
	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		values := append(setS(t), named{"3^19:1", noErr(cents.FromRatio(1162261467, 1))(t), ratio(1162261467, 1)},
			named{"2^31:1", noErr(cents.FromRatio(1<<31, 1))(t), ratio(1<<31, 1)})

		for _, a := range values {
			for _, digits := range []int{1, 5, 9} {
				out := a.v.Scala(digits)
				got, rest, err := cents.ParseScala(out)
				require.NoError(t, err, "%s: %q", a.name, out)
				assert.Empty(t, rest)

				if strings.Contains(out, "/") {
					assertCents(t, a.v, got, "%s: %q", a.name, out)
				} else {
					assertCents(t, noErr(cents.Parse(out))(t), got, "%s: %q", a.name, out)
				}
			}
		}
	})

	t.Run("must", func(t *testing.T) {
		t.Parallel()

		c, rest := cents.MustParseScala(" 5/4   E")
		assertCents(t, noErr(cents.FromRatio(5, 4))(t), c)
		assert.Equal(t, "   E", rest)
		assertPanicsWithErrorIs(t, cents.ErrPrimeLimit, func() { cents.MustParseScala("37/32") })
	})
}

// K-27: the explicit approximate path for ratios with a prime above 31.
func TestParseScalaApprox(t *testing.T) {
	t.Parallel()

	cases := append(scalaCommonCases(),
		// 1200*log2(37/32) = 251.3440386...: Python decimal, 120 digits.
		scalaCase{line: "37/32", want: ratC(251344039, 1000000)},
		scalaCase{line: "37/32 x", want: ratC(251344039, 1000000), rest: " x"},
		scalaCase{line: "32/37", want: ratC(-251344039, 1000000)},
		scalaCase{name: "37 and 100 zeros / 1", line: "37" + strings.Repeat("0", 100) + "/1", err: cents.ErrTooComplex},
	)

	runScalaCases(t, cents.ParseScalaApprox, cents.MustParseScalaApprox, cases)

	t.Run("same as ParseScala within the limit", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			for _, digits := range []int{1, 5} {
				line := a.v.Scala(digits)
				want, wantRest, err := cents.ParseScala(line)
				require.NoError(t, err)

				got, rest, err := cents.ParseScalaApprox(line)
				require.NoError(t, err)
				assertCents(t, want, got, line)
				assert.Equal(t, wantRest, rest)
			}
		}
	})

	t.Run("never prime limit", func(t *testing.T) {
		t.Parallel()

		for _, line := range []string{"37", "41/40", "1000003/1000000", "74/1", "999999/1"} {
			c, rest, err := cents.ParseScalaApprox(line)
			require.NoError(t, err, line)
			assertMust(t, mustScala{c, rest}, err, func() mustScala {
				c, rest := cents.MustParseScalaApprox(line)

				return mustScala{c, rest}
			}, "MustParseScalaApprox(%q)", line)
		}
	})

	t.Run("must", func(t *testing.T) {
		t.Parallel()

		c, rest := cents.MustParseScalaApprox("37/32")
		assertComp(t, ratC(251344039, 1000000), c)
		assert.Empty(t, rest)
		assertPanicsWithErrorIs(t, cents.ErrInvalidText, func() { cents.MustParseScalaApprox("-3/2") })
	})
}
