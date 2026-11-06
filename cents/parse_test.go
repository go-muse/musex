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

// parseCase is a text with the expected record or error.
type parseCase struct {
	name string
	text string
	want comp
	err  error
}

func runParseCases(t *testing.T, cases []parseCase) {
	t.Helper()

	for _, tc := range cases {
		name := tc.name
		if name == "" {
			name = strconv.Quote(tc.text)
		}

		t.Run(name+" "+errName(tc.err), func(t *testing.T) {
			t.Parallel()

			got, err := cents.Parse(tc.text)
			assertResult(t, tc.want, tc.err, got, err)
			assertMust(t, got, err, func() cents.Cents { return cents.MustParse(tc.text) }, "MustParse")
		})
	}
}

// repeatFactors returns first followed by n copies of "*" + factor.
func repeatFactors(first, factor string, n int) string {
	return first + strings.Repeat("*"+factor, n)
}

// K-24, G-9: accepted texts.
func TestParseAccepted(t *testing.T) {
	t.Parallel()

	c30 := ratC(30, 1)
	cm30 := ratC(-30, 1)
	e119 := stepC(1, 19)
	q := zeroComp().withExp(5, rat(1, 4))
	sav := ratC(6, 5).withExp(5, rat(1, 1000))

	runParseCases(t, []parseCase{
		{text: "30", want: c30},
		{text: "+30", want: c30},
		{text: "-30", want: cm30},
		{text: "−30", want: cm30},
		{text: "701.955", want: ratC(701955, 1000)},
		{text: "0.1", want: ratC(1, 10)},
		{text: "+30 ¢", want: c30},
		{text: "+30¢", want: c30},
		{text: "30 cents", want: c30},
		{text: "30 Cents", want: c30},
		{text: "30 CENTS", want: c30},
		{text: "30 cent", want: c30},
		{text: "1 cent", want: ratC(1, 1)},
		{text: "1 cents", want: ratC(1, 1)},
		{text: "2 cent", want: ratC(2, 1)},
		{text: "30 cents", want: c30},
		{text: "30 ¢", want: c30},
		{text: "-30 cents", want: cm30},
		{text: "−30 ¢", want: cm30},
		{text: "-0", want: zeroComp()},
		{text: "+0", want: zeroComp()},
		{text: "-0.0", want: zeroComp()},
		{text: "0007", want: ratC(7, 1)},
		{text: "1200/19 ¢", want: e119},
		{text: "-1200/19 cents", want: e119.neg()},
		{text: "2400/38 ¢", want: e119},
		{text: "700/1 cents", want: ratC(700, 1)},
		{text: "1/1 cents", want: ratC(1, 1)},
		{text: "0/1 cents", want: zeroComp()},
		{text: "37/32 cents", want: ratC(37, 32)},
		{text: "3:2", want: ratio(3, 2)},
		{text: "6:4", want: ratio(3, 2)},
		{text: "111:74", want: ratio(3, 2)},
		{text: "2:3", want: ratio(2, 3)},
		{text: "1:1", want: zeroComp()},
		{text: "5^(1/4)", want: q},
		{text: "10^(1/1000)", want: sav},
		{text: "2^(1/1000)*5^(1/1000)", want: sav},
		{text: "2^0", want: zeroComp()},
		{text: "2^1", want: ratC(1200, 1)},
		{text: "3^1", want: ratio(3, 1)},
		{text: "2^-1*3", want: ratio(3, 2)},
		{text: "2^-84*3^53", want: ratC(-100800, 1).withExp(3, rat(53, 1))},
		{text: "2^(-1199/1200)*3", want: ratio(3, 2).add(ratC(1, 1))},
		{text: "2^(2/4)", want: ratC(600, 1)},
		{text: "4^(1/2)", want: ratC(1200, 1)},
		{text: "2*3", want: ratio(6, 1)},
		{text: "3*2^-1", want: ratio(3, 2)},
		{text: "3^2*3^-2", want: zeroComp()},
		{text: "2^0*2^0", want: zeroComp()},
		{text: "31^(1/2)", want: zeroComp().withExp(31, rat(1, 2))},
		{text: "9000000", want: ratC(9000000, 1)},
		{text: "-9000000 cents", want: ratC(-9000000, 1)},
		{text: "2^7500", want: ratC(9000000, 1)},
		{name: "64 factors", text: repeatFactors("2", "2", 63), want: ratC(76800, 1)},
		{name: "64 factors with zero exponents", text: repeatFactors("2^0", "2^0", 63), want: zeroComp()},
		{name: "100 digits with leading zeros", text: strings.Repeat("0", 99) + "7", want: ratC(7, 1)},
		{name: "100-digit denominator", text: "1200/" + strings.Repeat("0", 98) + "19 cents", want: e119},
		{name: "30 digits after the point", text: "0." + strings.Repeat("0", 29) + "1", want: bigRatC(new(big.Rat).SetFrac(big.NewInt(1), pow10(30)))},
	})
}

// K-25, K-32: rejected texts; each error is its own kind, by the order of
// checks: syntax, text length, primes, record limit, range.
func TestParseRejected(t *testing.T) {
	t.Parallel()

	invalid := []string{
		"", "+", "-", "¢", "cents", "+-30", "--30", "1,5", "1,200", "30 цента", "30 ц.", "30 c", "30 ct",
		"1e3", "408.", ".5", " 30", "30 ", "30  cents", "+3:2", "-3:2", "3:2 ¢", "3 : 2", "3:2^2",
		"0:1", "3:0", "1/0 ¢", "2^(1/0)", "0^2", "1^2", "30\u00a0\u00a0cents", "30 \u00a0cents",
		"30\tcents", "30 cents ", "3:2 cents", "-2^2", "2^2 cents", "2^2¢", "2^1/2", "2^(2)", "2^(-1/-2)",
		"2^", "2^-", "*2", "2*", "2**3", "3:2:1", "1/2/3 cents", "1/2", "_30", "3_0", "0x10", "3·2", "3×2",
		"3∶2", "٣٠", "３０", "30.5.1", "1/-2 cents", "-1/-2 cents", "+-1/2 cents", "1 / 2 cents", "1/2  cents",
		"cent", "30 centss", "30centsx", "30 ct.", "3:", ":2", "3::2", "(3)", "2^(1/2", "2^1/2)",
		"2^(1/2)^2", "2 ^2", "2^ 2", "2 *3", "2* 3", "\n30", "30\n", "30\x00", "37:1*2^100000",
	}

	cases := make([]parseCase, 0, len(invalid)+32)
	for _, s := range invalid {
		cases = append(cases, parseCase{text: s, err: cents.ErrInvalidText})
	}

	digits101 := "1" + strings.Repeat("0", 100)
	p := new(big.Int).Add(new(big.Int).Exp(big.NewInt(10), big.NewInt(99), nil), big.NewInt(7)).String()

	cases = append(cases,
		parseCase{text: "37:32", err: cents.ErrPrimeLimit},
		parseCase{text: "83:88", err: cents.ErrPrimeLimit},
		parseCase{text: "74:1", err: cents.ErrPrimeLimit},
		parseCase{text: "37^(1/2)", err: cents.ErrPrimeLimit},
		parseCase{text: "2*37^0", err: cents.ErrPrimeLimit},
		// Architecture "Errors" and "Parsing": the prime limit is checked on
		// every written base, whatever its exponent and before exponents are
		// summed (confirmed by the architect).
		parseCase{text: "37^0", err: cents.ErrPrimeLimit},
		parseCase{text: "2^1*37^0", err: cents.ErrPrimeLimit},
		parseCase{text: "74^0", err: cents.ErrPrimeLimit},
		parseCase{text: "37*37^-1", err: cents.ErrPrimeLimit},
		parseCase{text: "37^(1/2)*37^(-1/2)", err: cents.ErrPrimeLimit},
		parseCase{text: "9000000.000001", err: cents.ErrOutOfRange},
		parseCase{text: "-9000001 cents", err: cents.ErrOutOfRange},
		parseCase{text: "9000001/1 cents", err: cents.ErrOutOfRange},
		parseCase{text: "2^7501", err: cents.ErrOutOfRange},
		parseCase{text: "2^100000", err: cents.ErrOutOfRange},
		// The record limit: 31 digits after the point (denominator 10^31).
		parseCase{name: "31 digits after the point", text: "0." + strings.Repeat("0", 30) + "1", err: cents.ErrTooComplex},
		// The length of numbers: 101 digits, including leading zeros.
		parseCase{name: "101 digits", text: digits101, err: cents.ErrTooComplex},
		parseCase{name: "101 digits with leading zeros", text: strings.Repeat("0", 100) + "7", err: cents.ErrTooComplex},
		parseCase{name: "101 digits after the point", text: "0." + strings.Repeat("0", 100) + "1", err: cents.ErrTooComplex},
		parseCase{name: "101-digit ratio part", text: "3:" + digits101, err: cents.ErrTooComplex},
		parseCase{name: "101-digit exponent", text: "2^" + digits101, err: cents.ErrTooComplex},
		// Length before primes and range (V-2, K-32).
		parseCase{name: "37 and 100 zeros : 1", text: "37" + strings.Repeat("0", 100) + ":1", err: cents.ErrTooComplex},
		parseCase{name: "101-digit cents out of range", text: "9" + strings.Repeat("0", 100), err: cents.ErrTooComplex},
		// The length of the numbers of a fraction of cents with unit
		// (K-25, K-32: syntax, then length, then range).
		parseCase{name: "fraction: 101-digit denominator", text: "1/" + digits101 + " cents", err: cents.ErrTooComplex},
		parseCase{name: "fraction: 101-digit denominator with leading zeros", text: "1/" + strings.Repeat("0", 100) + "7 cents", err: cents.ErrTooComplex},
		parseCase{name: "fraction: 101-digit numerator out of range", text: digits101 + "/19 cents", err: cents.ErrTooComplex},
		parseCase{name: "fraction: 101-digit numerator with leading zeros", text: strings.Repeat("0", 100) + "7/1 cents", err: cents.ErrTooComplex},
		parseCase{name: "fraction: 100-digit numerator", text: strings.Repeat("0", 98) + "12/1 cents", want: ratC(12, 1)},
		parseCase{name: "fraction: sign and ¢ with 101 digits", text: "-1/" + digits101 + "¢", err: cents.ErrTooComplex},
		parseCase{name: "fraction: cent with 101 digits", text: "1/" + digits101 + "cent", err: cents.ErrTooComplex},
		parseCase{name: "fraction: two spaces with 101 digits", text: "1/" + digits101 + "  cents", err: cents.ErrInvalidText},
		parseCase{name: "fraction: 101-digit zero denominator", text: "1/" + strings.Repeat("0", 101) + " cents", err: cents.ErrInvalidText},
		// Syntax before length: a bare a/b with long parts is invalid text.
		parseCase{name: "bare a/b with 101 digits", text: digits101 + "/2", err: cents.ErrInvalidText},
		parseCase{name: "syntax error with 101 digits", text: digits101 + "x", err: cents.ErrInvalidText},
		// Primes before the record limit and the range.
		parseCase{name: "prime before range in a power", text: "2^100000*37", err: cents.ErrPrimeLimit},
		parseCase{name: "prime before record limit", text: "37^(1/" + p + ")", err: cents.ErrPrimeLimit},
		// The record limit before the range.
		parseCase{name: "record limit before range", text: "2^100000*3^(1/" + p + ")", err: cents.ErrTooComplex},
		// Exponents are summed before the record limit is checked.
		parseCase{name: "3^(1/p)*3^(-1/p)", text: "3^(1/" + p + ")*3^(-1/" + p + ")", want: zeroComp()},
		parseCase{name: "3^(1/p)", text: "3^(1/" + p + ")", err: cents.ErrTooComplex},
	)

	runParseCases(t, cases)
}

// K-24, K-25, K-32 v5 (B-1, V-5): the limit MaxFactors of factors of a
// product of powers, checked with the length of numbers: after the syntax,
// before primes, the record limit and the range.
func TestParseFactorLimit(t *testing.T) {
	t.Parallel()

	p := func(i int64) string {
		return new(big.Int).Add(new(big.Int).Exp(big.NewInt(10), big.NewInt(99), nil), big.NewInt(i)).String()
	}
	b332 := new(big.Int).Lsh(big.NewInt(1), 332).String() // 2^332 has 100 digits (Python)
	require.Len(t, b332, 100)

	var b strings.Builder
	for i := int64(1); i <= 64; i++ {
		if i > 1 {
			b.WriteString("*")
		}

		b.WriteString(b332 + "^(1/" + p(i) + ")")
	}

	worst := b.String()

	runParseCases(t, []parseCase{
		{name: "64 factors", text: repeatFactors("2", "2", 63), want: ratC(76800, 1)},
		{name: "65 factors", text: repeatFactors("2", "2", 64), err: cents.ErrTooComplex},
		{name: "65 zero exponents", text: repeatFactors("2^0", "2^0", 64), err: cents.ErrTooComplex},
		{name: "65 factors with 37", text: repeatFactors("37", "2", 64), err: cents.ErrTooComplex},
		{name: "65 factors with 37 last", text: repeatFactors("2", "2", 63) + "*37", err: cents.ErrTooComplex},
		{name: "65 factors and a trailing star", text: repeatFactors("2", "2", 64) + "*", err: cents.ErrInvalidText},
		{name: "65 factors out of range", text: repeatFactors("2^100000", "2", 64), err: cents.ErrTooComplex},
		{name: "65 factors with a 101-digit number", text: repeatFactors("2", "2", 63) + "*" + "1" + strings.Repeat("0", 100), err: cents.ErrTooComplex},
		{name: "64 factors B^(1/p_i)", text: worst, err: cents.ErrTooComplex},
		{name: "64 factors with 37", text: repeatFactors("37", "2", 63), err: cents.ErrPrimeLimit},
		{name: "64 factors out of range", text: repeatFactors("2^100000", "2", 63), err: cents.ErrOutOfRange},
	})
}

// K-25: the bare a/b is invalid text with a hint naming a:b and a/b ¢.
func TestParseHint(t *testing.T) {
	t.Parallel()

	cases := []struct {
		text string
		a, b string
	}{
		{"3/2", "3", "2"},
		{"1200/19", "1200", "19"},
		{"0/5", "0", "5"},
		{"1/0", "1", "0"},
	}

	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()

			got, err := cents.Parse(tc.text)
			requireOnlyError(t, err, cents.ErrInvalidText)
			assertMust(t, got, err, func() cents.Cents { return cents.MustParse(tc.text) }, "MustParse")
			assert.True(t, got == cents.Cents{})
			assert.Contains(t, err.Error(), tc.a+":"+tc.b)
			assert.Contains(t, err.Error(), tc.a+"/"+tc.b+" ¢")
		})
	}
}

// P-17: round trip of every printed exact view for all S, all options.
func TestParseRoundTrip(t *testing.T) {
	t.Parallel()

	values := append(setS(t), named{"R(3:2)*5000", noErr(noErr(cents.FromRatio(3, 2))(t).Mul(5000))(t), ratio(3, 2).mulInt(5000)})
	for _, nv := range append(cents.NamedValues(), cents.NamedUnits()...) {
		values = append(values, named{nv.Name, nv.Value, compOf(nv.Value)})
	}

	for _, a := range values {
		t.Run(a.name, func(t *testing.T) {
			t.Parallel()

			assertCents(t, a.v, noErr(cents.Parse(a.v.String()))(t, "Parse(%q)", a.v.String()), "String")
			assertCents(t, a.v, cents.MustParse(a.v.String()), "MustParse of String")

			for _, opts := range optionSets() {
				text := a.v.Text(opts)
				assertCents(t, a.v, noErr(cents.Parse(text))(t, "Parse(%q)", text), "Text(%d) = %q", opts, text)

				for _, k := range kinds() {
					s, ok := a.v.TextAs(k, opts)
					if !ok {
						continue
					}

					assertCents(t, a.v, noErr(cents.Parse(s))(t, "Parse(%q)", s), "TextAs(%s, %d) = %q", k, opts, s)
				}
			}

			s, ok := a.v.TextAs(cents.KindPower, 0)
			require.True(t, ok)
			assertCents(t, a.v, noErr(cents.Parse(s))(t), "power view")
		})
	}
}

// Must: MustParse is Parse without the error.
func TestMustParse(t *testing.T) {
	t.Parallel()

	assertCents(t, noErr(cents.FromRatio(3, 2))(t), cents.MustParse("3:2"))
	assertPanicsWithErrorIs(t, cents.ErrInvalidText, func() { cents.MustParse("3/2") })
	assertPanicsWithErrorIs(t, cents.ErrPrimeLimit, func() { cents.MustParse("37:32") })
}
