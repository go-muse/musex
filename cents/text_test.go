package cents_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// K-23: decimal views, the canonical exact view and explicit kinds.
func TestDecimal(t *testing.T) {
	t.Parallel()

	r32, r23 := noErr(cents.FromRatio(3, 2))(t), noErr(cents.FromRatio(2, 3))(t)

	cases := []struct {
		name   string
		c      cents.Cents
		digits int
		opts   cents.Options
		want   string
	}{
		// Python decimal, 120 digits, half away from zero.
		{"R(3:2) 3", r32, 3, 0, "701.955"},
		{"R(3:2) 6", r32, 6, 0, "701.955001"},
		{"R(3:2) 0", r32, 0, 0, "702"},
		{"R(3:2) 15", r32, 15, 0, "701.955000865387418"},
		{"R(2:3) ASCII", r23, 3, 0, "-701.955"},
		{"R(2:3) typographic", r23, 3, cents.Typographic, "−701.955"},
		{"R(2:3) typographic with unit", r23, 3, cents.Typographic | cents.WithUnit, "−701.955 ¢"},
		{"E(1,19) 3", noErr(cents.Step(1, 19))(t), 3, 0, "63.158"},
		{"R(442:440) 2", noErr(cents.FromRatio(442, 440))(t), 2, 0, "7.85"},
		{"<-0.0004> 3", noErr(cents.FromFraction(-4, 10000))(t), 3, 0, "0.000"},
		{"<-0.0004> 3 typographic", noErr(cents.FromFraction(-4, 10000))(t), 3, cents.Typographic | cents.WithPlus, "0.000"},
		{"R(3:2)*5000 6", noErr(r32.Mul(5000))(t), 6, 0, "3509775.004327"},
		{"<7.1> 3 with plus", noErr(cents.FromFraction(71, 10))(t), 3, cents.WithPlus, "+7.100"},
		{"<7.1> 1 with unit", noErr(cents.FromFraction(71, 10))(t), 1, cents.WithUnit, "7.1 cents"},
		{"<700> 0", noErr(cents.FromInt(700))(t), 0, 0, "700"},
		{"<700> 2", noErr(cents.FromInt(700))(t), 2, 0, "700.00"},
		{"negative digits as 0", r32, -1, 0, "702"},
		{"MinInt digits as 0", r32, minInt, 0, "702"},
		// Exact halves of rational values: away from zero.
		{"1/8 2", noErr(cents.FromFraction(1, 8))(t), 2, 0, "0.13"},
		{"-1/8 2", noErr(cents.FromFraction(-1, 8))(t), 2, 0, "-0.13"},
		{"5/2 0", noErr(cents.FromFraction(5, 2))(t), 0, 0, "3"},
		{"-5/2 0", noErr(cents.FromFraction(-5, 2))(t), 0, 0, "-3"},
		{"<9000000> 2", noErr(cents.FromInt(9000000))(t), 2, 0, "9000000.00"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.c.Decimal(tc.digits, tc.opts))
		})
	}

	t.Run("many digits", func(t *testing.T) {
		t.Parallel()

		got := r32.Decimal(60, 0)
		assert.Equal(t, decimalOracle(ratio(3, 2), 60, 0), got)
		assert.True(t, strings.HasPrefix(got, "701.955000865387417744486732737379810511777289230977272546903"[:40]))
	})

	// K-23, NF-11: hundreds of digits are still correctly rounded.
	t.Run("hundreds of digits", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			file   string
			c      cents.Cents
			digits int
		}{
			{"r32-900", cents.MustFromRatio(3, 2), 900},
			{"r23-1000", cents.MustFromRatio(2, 3), 1000},
			{"q-900", cents.MustParse("5^(1/4)"), 900},
		}

		for _, tc := range cases {
			want := decimalLiteral(t, tc.file)
			got := tc.c.Decimal(tc.digits, 0)
			assertDecimalShape(t, got, tc.digits, tc.file)
			assert.Equal(t, want, got, tc.file)
		}
	})

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			for _, digits := range []int{-1, 0, 1, 2, 3, 6, 9, 15, 30} {
				for _, opts := range optionSets() {
					assert.Equal(t, decimalOracle(a.c, digits, opts), a.v.Decimal(digits, opts),
						"%s.Decimal(%d, %d)", a.name, digits, opts)
				}
			}
		}
	})
}

// decimalLiteral returns the expected decimal view in
// testdata/decimal/<name>.txt: computed with Python decimal at 1200 and
// 1400 significant digits (the same views), quantize with ROUND_HALF_UP.
func decimalLiteral(t testing.TB, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", "decimal", name+".txt"))
	require.NoError(t, err)

	return strings.TrimSuffix(string(data), "\n")
}

// assertDecimalShape checks the form of a decimal view with digits > 0:
// an optional minus, integer digits, the point and exactly digits digits.
func assertDecimalShape(t testing.TB, s string, digits int, msgAndArgs ...any) bool {
	t.Helper()

	intPart, frac, ok := strings.Cut(strings.TrimPrefix(s, "-"), ".")
	ok = assert.True(t, ok, msgAndArgs...) && ok
	ok = assert.Len(t, frac, digits, msgAndArgs...) && ok
	ok = assert.NotEmpty(t, intPart, msgAndArgs...) && ok

	return assert.Empty(t, strings.Trim(intPart+frac, "0123456789"), msgAndArgs...) && ok
}

// K-23: canonical views (String, Text, Kind) and explicit views (TextAs).
func TestText(t *testing.T) {
	t.Parallel()

	set := setS(t)
	r32 := noErr(cents.FromRatio(3, 2))(t)

	canonical := []struct {
		name string
		c    cents.Cents
		kind cents.Kind
		want string
	}{
		{"<700>", noErr(cents.FromInt(700))(t), cents.KindDecimal, "700"},
		{"<7.1>", noErr(cents.FromFraction(71, 10))(t), cents.KindDecimal, "7.1"},
		{"<-13.7>", noErr(cents.FromFraction(-137, 10))(t), cents.KindDecimal, "-13.7"},
		{"MIDI unit", cents.MIDITuningUnit(), cents.KindDecimal, "0.006103515625"},
		{"0", cents.Zero(), cents.KindDecimal, "0"},
		{"<1200>", cents.Octave(), cents.KindDecimal, "1200"},
		{"E(1,19)", noErr(cents.Step(1, 19))(t), cents.KindFraction, "1200/19 cents"},
		{"-E(1,19)", noErr(cents.Step(-1, 19))(t), cents.KindFraction, "-1200/19 cents"},
		{"R(3:2)", r32, cents.KindRatio, "3:2"},
		{"R(2:3)", noErr(cents.FromRatio(2, 3))(t), cents.KindRatio, "2:3"},
		{"PYTH", cents.PythagoreanComma(), cents.KindRatio, "531441:524288"},
		{"MK", byName(t, set, "MK").v, cents.KindRatio, "19383245667680019896796723:19342813113834066795298816"},
		{"Q", byName(t, set, "Q").v, cents.KindPower, "5^(1/4)"},
		{"SAV", cents.Savart(), cents.KindPower, "2^(1/1000)*5^(1/1000)"},
		{"R(3:2) + <1>", noErr(r32.Add(cents.Cent()))(t), cents.KindPower, "2^(-1199/1200)*3"},
		// 3^5000:2^5000 has 2386 and 1506 digits (Python), beyond MaxDigits.
		{"R(3:2)*5000", noErr(r32.Mul(5000))(t), cents.KindPower, "2^-5000*3^5000"},
	}

	for _, tc := range canonical {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.kind, tc.c.Kind())
			assert.Equal(t, tc.want, tc.c.String())
			assert.Equal(t, tc.want, tc.c.Text(0))
			assert.Equal(t, tc.want, stringOracle(compOf(tc.c)), "oracle")

			s, ok := tc.c.TextAs(tc.kind, 0)
			assert.True(t, ok)
			assert.Equal(t, tc.want, s)

			back, err := cents.Parse(tc.c.String())
			require.NoError(t, err)
			assertCents(t, tc.c, back, "round trip")
		})
	}

	t.Run("explicit kinds", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name string
			c    cents.Cents
			kind cents.Kind
			want string
			ok   bool
		}{
			{"<1200> as ratio", cents.Octave(), cents.KindRatio, "2:1", true},
			{"<700> as ratio", noErr(cents.FromInt(700))(t), cents.KindRatio, "", false},
			{"<700> as fraction", noErr(cents.FromInt(700))(t), cents.KindFraction, "700/1 cents", true},
			{"<700> as power", noErr(cents.FromInt(700))(t), cents.KindPower, "2^(7/12)", true},
			{"0 as ratio", cents.Zero(), cents.KindRatio, "1:1", true},
			{"0 as fraction", cents.Zero(), cents.KindFraction, "0/1 cents", true},
			{"0 as power", cents.Zero(), cents.KindPower, "2^0", true},
			{"<1200> as power", cents.Octave(), cents.KindPower, "2^1", true},
			{"<-1200> as power", noErr(cents.FromInt(-1200))(t), cents.KindPower, "2^-1", true},
			{"R(3:1) as power", noErr(cents.FromRatio(3, 1))(t), cents.KindPower, "3^1", true},
			{"R(3:2) as power", r32, cents.KindPower, "2^-1*3", true},
			{"MK as power", byName(t, set, "MK").v, cents.KindPower, "2^-84*3^53", true},
			{"R(3:2) as decimal", r32, cents.KindDecimal, "", false},
			{"R(3:2) as fraction", r32, cents.KindFraction, "", false},
			{"E(1,19) as decimal", noErr(cents.Step(1, 19))(t), cents.KindDecimal, "", false},
			{"Q as ratio", byName(t, set, "Q").v, cents.KindRatio, "", false},
			{"R(3:2)*5000 as ratio", noErr(r32.Mul(5000))(t), cents.KindRatio, "", false},
			{"unknown kind 4", r32, cents.Kind(4), "", false},
			{"unknown kind 9", cents.Zero(), cents.Kind(9), "", false},
		}

		for _, tc := range cases {
			s, ok := tc.c.TextAs(tc.kind, 0)
			assert.Equal(t, tc.ok, ok, tc.name)
			assert.Equal(t, tc.want, s, tc.name)
		}
	})

	// The limit MaxDigits of KindRatio: 3^209 has 100 digits, 3^210 has 101
	// (Python).
	t.Run("ratio digits limit", func(t *testing.T) {
		t.Parallel()

		r31 := noErr(cents.FromRatio(3, 1))(t)
		at := noErr(r31.Mul(209))(t)
		assert.Equal(t, cents.KindRatio, at.Kind())
		require.True(t, strings.HasSuffix(at.String(), ":1"))
		assert.Len(t, strings.TrimSuffix(at.String(), ":1"), 100)
		assertCents(t, at, noErr(cents.Parse(at.String()))(t))

		above := noErr(r31.Mul(210))(t)
		assert.Equal(t, cents.KindPower, above.Kind())
		assert.Equal(t, "3^210", above.String())
		s, ok := above.TextAs(cents.KindRatio, 0)
		assert.False(t, ok)
		assert.Empty(t, s)

		big5000 := noErr(r32.Mul(5000))(t)
		_, _, ok = big5000.Ratio()
		assert.False(t, ok)
		assertCents(t, big5000, noErr(cents.Parse("2^-5000*3^5000"))(t))
	})

	t.Run("options", func(t *testing.T) {
		t.Parallel()

		c71, cm137 := noErr(cents.FromFraction(71, 10))(t), noErr(cents.FromFraction(-137, 10))(t)
		e119 := noErr(cents.Step(1, 19))(t)

		cases := []struct {
			name string
			c    cents.Cents
			opts cents.Options
			want string
		}{
			{"<7.1> unit", c71, cents.WithUnit, "7.1 cents"},
			{"<7.1> unit typographic", c71, cents.WithUnit | cents.Typographic, "7.1 ¢"},
			{"<7.1> plus", c71, cents.WithPlus, "+7.1"},
			{"<7.1> plus unit", c71, cents.WithPlus | cents.WithUnit, "+7.1 cents"},
			{"<7.1> all", c71, cents.WithPlus | cents.WithUnit | cents.Typographic, "+7.1 ¢"},
			{"<-13.7> typographic unit", cm137, cents.Typographic | cents.WithUnit, "−13.7 ¢"},
			{"<-13.7> plus", cm137, cents.WithPlus, "-13.7"},
			{"0 plus", cents.Zero(), cents.WithPlus, "0"},
			{"0 all", cents.Zero(), cents.WithPlus | cents.WithUnit | cents.Typographic, "0 ¢"},
			{"E(1,19) typographic", e119, cents.Typographic, "1200/19 ¢"},
			{"E(1,19) plus", e119, cents.WithPlus, "+1200/19 cents"},
			{"-E(1,19) typographic", e119.Neg(), cents.Typographic, "−1200/19 ¢"},
			{"R(3:2) all", r32, cents.WithPlus | cents.WithUnit | cents.Typographic, "3:2"},
			{"R(2:3) all", noErr(cents.FromRatio(2, 3))(t), cents.WithPlus | cents.WithUnit | cents.Typographic, "2:3"},
			{"Q all", byName(t, set, "Q").v, cents.WithPlus | cents.WithUnit | cents.Typographic, "5^(1/4)"},
			{"unknown bits ignored", c71, cents.WithUnit | 0x80, "7.1 cents"},
		}

		for _, tc := range cases {
			assert.Equal(t, tc.want, tc.c.Text(tc.opts), tc.name)
		}
	})

	t.Run("for all S, kinds and options", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			assert.Equal(t, kindOracle(a.c), a.v.Kind(), a.name)
			assert.Equal(t, stringOracle(a.c), a.v.String(), a.name)
			assert.True(t, isASCII(a.v.String()), a.name)

			for _, opts := range optionSets() {
				want, _ := textOracle(a.c, kindOracle(a.c), opts)
				assert.Equal(t, want, a.v.Text(opts), "%s.Text(%d)", a.name, opts)

				for _, k := range kinds() {
					wantS, wantOK := textOracle(a.c, k, opts)
					s, ok := a.v.TextAs(k, opts)
					assert.Equal(t, wantOK, ok, "%s.TextAs(%s, %d)", a.name, k, opts)
					assert.Equal(t, wantS, s, "%s.TextAs(%s, %d)", a.name, k, opts)

					if opts&cents.Typographic == 0 {
						assert.True(t, isASCII(s), "ASCII: %s.TextAs(%s, %d)", a.name, k, opts)
					}
				}
			}

			assert.True(t, isASCII(a.v.Decimal(6, cents.WithUnit|cents.WithPlus)), a.name)
		}
	})
}

// K-23 (architecture, "Singular cent", B-2): the unit is "cent" exactly when
// the printed number without its sign is "1"; "¢" in the typographic style.
func TestUnitSingular(t *testing.T) {
	t.Parallel()

	c1, cm1 := cents.Cent(), noErr(cents.FromInt(-1))(t)

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"Cent().Text(WithUnit)", c1.Text(cents.WithUnit), "1 cent"},
		{"Cent().Decimal(0, WithUnit)", c1.Decimal(0, cents.WithUnit), "1 cent"},
		{"<-1>.Text(WithUnit)", cm1.Text(cents.WithUnit), "-1 cent"},
		{"Cent().Text(WithPlus|WithUnit)", c1.Text(cents.WithPlus | cents.WithUnit), "+1 cent"},
		{"<-1>.Text(Typographic|WithUnit)", cm1.Text(cents.Typographic | cents.WithUnit), "−1 ¢"},
		{"Cent().Decimal(3, WithUnit)", c1.Decimal(3, cents.WithUnit), "1.000 cents"},
		{"<2499/2500>.Decimal(0, WithUnit)", noErr(cents.FromFraction(2499, 2500))(t).Decimal(0, cents.WithUnit), "1 cent"},
		{"<2499/2500>.Decimal(3, WithUnit)", noErr(cents.FromFraction(2499, 2500))(t).Decimal(3, cents.WithUnit), "1.000 cents"},
		{"<-2499/2500>.Decimal(0, WithUnit)", noErr(cents.FromFraction(-2499, 2500))(t).Decimal(0, cents.WithUnit), "-1 cent"},
		{"<7/5>.Decimal(0, WithUnit)", noErr(cents.FromFraction(7, 5))(t).Decimal(0, cents.WithUnit), "1 cent"},
		{"Zero().Text(WithUnit)", cents.Zero().Text(cents.WithUnit), "0 cents"},
		{"<-1/2500>.Decimal(0, WithUnit|WithPlus)", noErr(cents.FromFraction(-1, 2500))(t).Decimal(0, cents.WithUnit|cents.WithPlus), "0 cents"},
		{"<2>.Text(WithUnit)", noErr(cents.FromInt(2))(t).Text(cents.WithUnit), "2 cents"},
		{"<10>.Text(WithUnit)", noErr(cents.FromInt(10))(t).Text(cents.WithUnit), "10 cents"},
		{"<7.1>.Text(WithUnit)", noErr(cents.FromFraction(71, 10))(t).Text(cents.WithUnit), "7.1 cents"},
		{"E(1,19).Text(WithUnit)", noErr(cents.Step(1, 19))(t).Text(cents.WithUnit), "1200/19 cents"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.got)
		})
	}

	t.Run("fraction", func(t *testing.T) {
		t.Parallel()

		s, ok := c1.TextAs(cents.KindFraction, 0)
		assert.True(t, ok)
		assert.Equal(t, "1/1 cents", s)

		s, ok = cm1.TextAs(cents.KindFraction, 0)
		assert.True(t, ok)
		assert.Equal(t, "-1/1 cents", s)
	})

	// For all a in S, digits 0...3 and options with WithUnit without
	// Typographic: "cent" ⇔ the printed number without its sign is "1".
	t.Run("rule for all S", func(t *testing.T) {
		t.Parallel()

		values := append(setS(t), named{"<1>", c1, ratC(1, 1)}, named{"<-1>", cm1, ratC(-1, 1)},
			named{"<0.9996>", noErr(cents.FromFraction(2499, 2500))(t), ratC(2499, 2500)})

		for _, a := range values {
			for digits := range 4 {
				for _, opts := range []cents.Options{cents.WithUnit, cents.WithUnit | cents.WithPlus} {
					got := a.v.Decimal(digits, opts)
					number, unit, found := strings.Cut(got, " ")
					require.True(t, found, got)

					wantUnit := "cents"
					if strings.TrimLeft(number, "+-") == "1" {
						wantUnit = "cent"
					}

					assert.Equal(t, wantUnit, unit, fmt.Sprintf("%s.Decimal(%d, %d) = %q", a.name, digits, opts, got))
				}
			}
		}
	})
}
