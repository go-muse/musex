package pitchclass_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/pitchclass"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time checks: the types implement fmt.Stringer (C-29, C-33).
var (
	_ fmt.Stringer = pitchclass.Class{}
	_ fmt.Stringer = pitchclass.Operator{}
	_ fmt.Stringer = pitchclass.Decimal
)

// K-25, C-29, C-30: the canonical form "<k> mod <N>" of every class is
// distinct and parses back to an equal class.
func TestClassStringAndParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		c    pitchclass.Class
		want string
	}{
		{name: "1 of 12", c: newClass(t, 1, 12), want: "1 mod 12"},
		{name: "0 of 1", c: newClass(t, 0, 1), want: "0 mod 1"},
		{name: "1 of 19", c: newClass(t, 1, 19), want: "1 mod 19"},
		{name: "11 of 12", c: newClass(t, 11, 12), want: "11 mod 12"},
		{name: "MAX-1 of MAX", c: newClass(t, maxInt-1, maxInt), want: "9223372036854775806 mod 9223372036854775807"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.c.String())
			assert.Equal(t, tc.want, fmt.Sprint(tc.c), "fmt uses String")

			got, err := pitchclass.Parse(tc.want)
			require.NoError(t, err)
			assert.True(t, got.Equal(tc.c))
			assert.True(t, pitchclass.MustParse(tc.want).Equal(tc.c))
		})
	}

	t.Run("all classes", func(t *testing.T) {
		t.Parallel()

		seen := map[string]pitchclass.Class{}
		all := append(classesOfAll(t, 1, 2, 12, 19, 31), newClass(t, maxInt-1, maxInt), newClass(t, 0, maxInt))
		for _, c := range all {
			s := c.String()
			assert.Equal(t, fmt.Sprintf("%d mod %d", c.Number(), c.Modulus()), s)

			prev, dup := seen[s]
			assert.False(t, dup, "%q is the form of %v and %v", s, prev, c)
			seen[s] = c

			back, err := pitchclass.Parse(s)
			require.NoError(t, err, "Parse(%q)", s)
			assertClass(t, c.Number(), c.Modulus(), back, fmt.Sprintf("Parse(%q)", s))
		}
	})
}

// K-25, E-6: Parse accepts only canonical forms of valid classes.
func TestParseErrors(t *testing.T) {
	t.Parallel()

	inputs := []string{
		"",
		"1",
		"12",
		"mod 12",
		"1 mod",
		"1 mod ",
		"1mod12",
		"1 mod12",
		"1mod 12",
		"0 mod 0",
		"0 mod -1",
		"12 mod 12",
		"13 mod 12",
		"1 mod 1",
		"-1 mod 12",
		"+1 mod 12",
		"-0 mod 12",
		"1 mod +12",
		"1 mod -12",
		"01 mod 12",
		"00 mod 12",
		"1 mod 012",
		"1  mod 12",
		"1 mod  12",
		" 1 mod 12",
		"1 mod 12 ",
		"1 mod 12\n",
		"1\tmod 12",
		"1 MOD 12",
		"1 Mod 12",
		"1 of 12",
		"1/12",
		"1 mod 12 mod 12",
		"0x1 mod 12",
		"1 mod 0xC",
		"1_0 mod 12",
		"1.0 mod 12",
		"\u0661 mod 12",
		"\uff11 mod 12",
		"1 mod \uff11\uff12",
		"1\u00a0mod 12",
		"Class(invalid)",
		"9223372036854775807 mod 9223372036854775807",
		"0 mod 9223372036854775808",
		"9223372036854775808 mod 12",
		"99999999999999999999 mod 12",
		"T5M7 mod 12",
	}

	for _, in := range inputs {
		t.Run(fmt.Sprintf("%q", in), func(t *testing.T) {
			t.Parallel()

			c, err := pitchclass.Parse(in)
			assertOnlyError(t, err, pitchclass.ErrNotation, "Parse")
			assert.Equal(t, pitchclass.Class{}, c, "the zero Class on an error")
			assertPanicsWithErrorIs(t, pitchclass.ErrNotation, func() { pitchclass.MustParse(in) }, "MustParse")
		})
	}
}

// K-26, C-31, G-10: the literary notation of the class number.
func TestClassNotation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		k, n  int
		style pitchclass.Style
		want  string
	}{
		{name: "10 of 12 Decimal", k: 10, n: 12, style: pitchclass.Decimal, want: "10"},
		{name: "10 of 12 TE", k: 10, n: 12, style: pitchclass.TE, want: "T"},
		{name: "10 of 12 LowerTE", k: 10, n: 12, style: pitchclass.LowerTE, want: "t"},
		{name: "10 of 12 AB", k: 10, n: 12, style: pitchclass.AB, want: "A"},
		{name: "11 of 12 Decimal", k: 11, n: 12, style: pitchclass.Decimal, want: "11"},
		{name: "11 of 12 TE", k: 11, n: 12, style: pitchclass.TE, want: "E"},
		{name: "11 of 12 LowerTE", k: 11, n: 12, style: pitchclass.LowerTE, want: "e"},
		{name: "11 of 12 AB", k: 11, n: 12, style: pitchclass.AB, want: "B"},
		{name: "10 of 11 Decimal", k: 10, n: 11, style: pitchclass.Decimal, want: "10"},
		{name: "10 of 11 TE", k: 10, n: 11, style: pitchclass.TE, want: "T"},
		{name: "10 of 11 LowerTE", k: 10, n: 11, style: pitchclass.LowerTE, want: "t"},
		{name: "10 of 11 AB", k: 10, n: 11, style: pitchclass.AB, want: "A"},
		{name: "12 of 19 Decimal", k: 12, n: 19, style: pitchclass.Decimal, want: "12"},
		{name: "18 of 19 Decimal", k: 18, n: 19, style: pitchclass.Decimal, want: "18"},
		{name: "MAX-1 of MAX Decimal", k: maxInt - 1, n: maxInt, style: pitchclass.Decimal, want: "9223372036854775806"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := newClass(t, tc.k, tc.n)
			got, err := c.Notation(tc.style)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.want, c.MustNotation(tc.style))
		})
	}

	t.Run("digits 0...9 at 12 in all styles", func(t *testing.T) {
		t.Parallel()

		for k := range 10 {
			for _, st := range styles() {
				got, err := newClass(t, k, 12).Notation(st)
				require.NoError(t, err)
				assert.Equal(t, fmt.Sprint(k), got, "%d in %v", k, st)
			}
		}
	})

	t.Run("letter styles equal Decimal at N <= 10", func(t *testing.T) {
		t.Parallel()

		for n := 1; n <= 10; n++ {
			for _, c := range classesOf(t, n) {
				for _, st := range styles() {
					got, err := c.Notation(st)
					require.NoError(t, err, "%v in %v", c, st)
					assert.Equal(t, fmt.Sprint(c.Number()), got, "%v in %v", c, st)
				}
			}
		}
	})

	t.Run("Decimal at any N", func(t *testing.T) {
		t.Parallel()

		for _, c := range append(classesOfAll(t, 13, 19, 31), newClass(t, 0, maxInt), newClass(t, maxInt/2, maxInt)) {
			got, err := c.Notation(pitchclass.Decimal)
			require.NoError(t, err, "%v", c)
			assert.Equal(t, fmt.Sprint(c.Number()), got)
		}
	})

	t.Run("letter styles at N > 12", func(t *testing.T) {
		t.Parallel()

		for _, c := range append(classesOfAll(t, 13, 19), newClass(t, 0, maxInt)) {
			for _, st := range []pitchclass.Style{pitchclass.TE, pitchclass.LowerTE, pitchclass.AB} {
				got, err := c.Notation(st)
				assertOnlyError(t, err, pitchclass.ErrNotation, fmt.Sprintf("%v in %v", c, st))
				assert.Empty(t, got)
				assertPanicsWithErrorIs(t, pitchclass.ErrNotation, func() { c.MustNotation(st) }, "MustNotation")
			}
		}
	})

	t.Run("invalid style", func(t *testing.T) {
		t.Parallel()

		for _, st := range []pitchclass.Style{pitchclass.Style(4), pitchclass.Style(9), pitchclass.Style(255)} {
			for _, c := range []pitchclass.Class{newClass(t, 1, 12), newClass(t, 0, 1), newClass(t, 3, 19)} {
				got, err := c.Notation(st)
				assertOnlyError(t, err, pitchclass.ErrNotation, fmt.Sprintf("%v in %v", c, st))
				assert.Empty(t, got)
				assertPanicsWithErrorIs(t, pitchclass.ErrNotation, func() { c.MustNotation(st) }, "MustNotation")
			}
		}
	})
}

// K-26, C-32, E-2, E-6: parsing the literary notation in a given style.
func TestParseNotation(t *testing.T) {
	t.Parallel()

	valid := []struct {
		name  string
		s     string
		style pitchclass.Style
		n     int
		want  int
	}{
		{name: "T in TE at 12", s: "T", style: pitchclass.TE, n: 12, want: 10},
		{name: "E in TE at 12", s: "E", style: pitchclass.TE, n: 12, want: 11},
		{name: "e in LowerTE at 12", s: "e", style: pitchclass.LowerTE, n: 12, want: 11},
		{name: "t in LowerTE at 12", s: "t", style: pitchclass.LowerTE, n: 12, want: 10},
		{name: "B in AB at 12", s: "B", style: pitchclass.AB, n: 12, want: 11},
		{name: "A in AB at 12", s: "A", style: pitchclass.AB, n: 12, want: 10},
		{name: "T in TE at 11", s: "T", style: pitchclass.TE, n: 11, want: 10},
		{name: "11 in Decimal at 19", s: "11", style: pitchclass.Decimal, n: 19, want: 11},
		{name: "10 in Decimal at 12", s: "10", style: pitchclass.Decimal, n: 12, want: 10},
		{name: "0 in Decimal at 1", s: "0", style: pitchclass.Decimal, n: 1, want: 0},
		{name: "7 in AB at 12", s: "7", style: pitchclass.AB, n: 12, want: 7},
		{name: "MAX-1 in Decimal at MAX", s: "9223372036854775806", style: pitchclass.Decimal, n: maxInt, want: maxInt - 1},
	}

	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := pitchclass.ParseNotation(tc.s, tc.style, tc.n)
			require.NoError(t, err)
			assertClass(t, tc.want, tc.n, c, "ParseNotation")
			assertClass(t, tc.want, tc.n, pitchclass.MustParseNotation(tc.s, tc.style, tc.n), "MustParseNotation")
		})
	}

	invalid := []struct {
		name  string
		s     string
		style pitchclass.Style
		n     int
		want  error
	}{
		{name: "T in Decimal at 12", s: "T", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "01 in Decimal at 12", s: "01", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "+1 in Decimal at 12", s: "+1", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "-0 (ASCII hyphen) in Decimal at 12", s: "-0", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "U+2212 minus sign and 0 in Decimal at 12", s: "\u22120", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "-1 in Decimal at 12", s: "-1", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "space 1 in Decimal at 12", s: " 1", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "1 space in Decimal at 12", s: "1 ", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "12 in Decimal at 12", s: "12", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "00 in Decimal at 12", s: "00", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "fullwidth 1 in Decimal at 12", s: "\uff11", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "1 mod 12 in Decimal at 12", s: "1 mod 12", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "MAX in Decimal at MAX", s: "9223372036854775807", style: pitchclass.Decimal, n: maxInt, want: pitchclass.ErrNotation},
		{name: "beyond int in Decimal at MAX", s: "9223372036854775808", style: pitchclass.Decimal, n: maxInt, want: pitchclass.ErrNotation},
		{name: "t in TE at 12", s: "t", style: pitchclass.TE, n: 12, want: pitchclass.ErrNotation},
		{name: "A in TE at 12", s: "A", style: pitchclass.TE, n: 12, want: pitchclass.ErrNotation},
		{name: "T in LowerTE at 12", s: "T", style: pitchclass.LowerTE, n: 12, want: pitchclass.ErrNotation},
		{name: "a in AB at 12", s: "a", style: pitchclass.AB, n: 12, want: pitchclass.ErrNotation},
		{name: "E in TE at 11", s: "E", style: pitchclass.TE, n: 11, want: pitchclass.ErrNotation},
		{name: "T in TE at 10", s: "T", style: pitchclass.TE, n: 10, want: pitchclass.ErrNotation},
		{name: "10 in TE at 12", s: "10", style: pitchclass.TE, n: 12, want: pitchclass.ErrNotation},
		{name: "11 in AB at 12", s: "11", style: pitchclass.AB, n: 12, want: pitchclass.ErrNotation},
		{name: "10 in LowerTE at 11", s: "10", style: pitchclass.LowerTE, n: 11, want: pitchclass.ErrNotation},
		{name: "TE twice", s: "TT", style: pitchclass.TE, n: 12, want: pitchclass.ErrNotation},
		{name: "empty in Decimal", s: "", style: pitchclass.Decimal, n: 12, want: pitchclass.ErrNotation},
		{name: "empty in TE", s: "", style: pitchclass.TE, n: 12, want: pitchclass.ErrNotation},
		{name: "empty in LowerTE", s: "", style: pitchclass.LowerTE, n: 12, want: pitchclass.ErrNotation},
		{name: "empty in AB", s: "", style: pitchclass.AB, n: 12, want: pitchclass.ErrNotation},
		{name: "0 in TE at 13", s: "0", style: pitchclass.TE, n: 13, want: pitchclass.ErrNotation},
		{name: "T in TE at 13", s: "T", style: pitchclass.TE, n: 13, want: pitchclass.ErrNotation},
		{name: "1 in LowerTE at 19", s: "1", style: pitchclass.LowerTE, n: 19, want: pitchclass.ErrNotation},
		{name: "B in AB at 19", s: "B", style: pitchclass.AB, n: 19, want: pitchclass.ErrNotation},
		{name: "12 in AB at 19", s: "12", style: pitchclass.AB, n: 19, want: pitchclass.ErrNotation},
		{name: "1 in invalid style 9", s: "1", style: pitchclass.Style(9), n: 12, want: pitchclass.ErrNotation},
		{name: "1 in invalid style 4", s: "1", style: pitchclass.Style(4), n: 12, want: pitchclass.ErrNotation},
		{name: "1 in Decimal at 0", s: "1", style: pitchclass.Decimal, n: 0, want: pitchclass.ErrInvalidModulus},
		{name: "0 in Decimal at -1", s: "0", style: pitchclass.Decimal, n: -1, want: pitchclass.ErrInvalidModulus},
		{name: "T in TE at MIN", s: "T", style: pitchclass.TE, n: minInt, want: pitchclass.ErrInvalidModulus},
		{name: "N checked before the style", s: "x", style: pitchclass.Style(9), n: 0, want: pitchclass.ErrInvalidModulus},
	}

	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := pitchclass.ParseNotation(tc.s, tc.style, tc.n)
			assertOnlyError(t, err, tc.want, "ParseNotation")
			assert.Equal(t, pitchclass.Class{}, c, "the zero Class on an error")
			assertPanicsWithErrorIs(t, tc.want, func() { pitchclass.MustParseNotation(tc.s, tc.style, tc.n) }, "MustParseNotation")
		})
	}
}

// K-26: writing and parsing are inverse for all styles and all classes of
// N in {1, 2, 11, 12}.
func TestNotationRoundTrip(t *testing.T) {
	t.Parallel()

	for _, n := range []int{1, 2, 11, 12} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			t.Parallel()

			for _, st := range styles() {
				seen := map[string]bool{}
				for _, c := range classesOf(t, n) {
					s, err := c.Notation(st)
					require.NoError(t, err, "%v in %v", c, st)
					assert.Equal(t, notationOf(c.Number(), st), s, "%v in %v", c, st)
					assert.False(t, seen[s], "%q repeats in %v at %d", s, st, n)
					seen[s] = true

					back, err := pitchclass.ParseNotation(s, st, n)
					require.NoError(t, err, "ParseNotation(%q, %v, %d)", s, st, n)
					assert.True(t, back.Equal(c), "ParseNotation(%q, %v, %d)", s, st, n)
				}
			}
		})
	}
}

// K-26: Style.String names the constants; other values are "Style(<n>)".
func TestStyleString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		style pitchclass.Style
		want  string
	}{
		{style: pitchclass.Decimal, want: "Decimal"},
		{style: pitchclass.TE, want: "TE"},
		{style: pitchclass.LowerTE, want: "LowerTE"},
		{style: pitchclass.AB, want: "AB"},
		{style: pitchclass.Style(4), want: "Style(4)"},
		{style: pitchclass.Style(9), want: "Style(9)"},
		{style: pitchclass.Style(255), want: "Style(255)"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.style.String())
			assert.Equal(t, tc.want, fmt.Sprint(tc.style))
		})
	}
}

// K-27, C-33: the canonical form "T<n>M<m> mod <N>" of operators.
func TestOperatorStringAndParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    pitchclass.Operator
		want string
	}{
		{name: "T5M7 at 12", f: pitchclass.MustNewOperator(7, 5, 12), want: "T5M7 mod 12"},
		{name: "identity at 12", f: pitchclass.MustIdentity(12), want: "T0M1 mod 12"},
		{name: "I5 at 12", f: pitchclass.MustInversion(5, 12), want: "T5M11 mod 12"},
		{name: "T3∘I8 at 12", f: pitchclass.MustTransposition(3, 12).MustCompose(pitchclass.MustInversion(8, 12)), want: "T11M11 mod 12"},
		{name: "the operator at 1", f: pitchclass.MustIdentity(1), want: "T0M0 mod 1"},
		{name: "T5 at 12", f: pitchclass.MustTransposition(5, 12), want: "T5M1 mod 12"},
		{name: "T5 at 19", f: pitchclass.MustTransposition(5, 19), want: "T5M1 mod 19"},
		{name: "T1 at 2", f: pitchclass.MustTransposition(1, 2), want: "T1M1 mod 2"},
		{name: "NewOperator(MAX, MIN, 12)", f: pitchclass.MustNewOperator(maxInt, minInt, 12), want: "T4M7 mod 12"},
		{
			name: "T(MAX-1)M(MAX-1) at MAX", f: pitchclass.MustNewOperator(maxInt-1, maxInt-1, maxInt),
			want: "T9223372036854775806M9223372036854775806 mod 9223372036854775807",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.f.String())
			assert.Equal(t, tc.want, fmt.Sprint(tc.f), "fmt uses String")

			got, err := pitchclass.ParseOperator(tc.want)
			require.NoError(t, err)
			assert.True(t, got.Equal(tc.f))
			assert.True(t, pitchclass.MustParseOperator(tc.want).Equal(tc.f))
		})
	}

	t.Run("all operators", func(t *testing.T) {
		t.Parallel()

		seen := map[string]bool{}
		for _, size := range []int{1, 2, 12, 19} {
			ops, specs := operatorsOf(t, size)
			for i, f := range ops {
				s := f.String()
				assert.Equal(t, specs[i].String(), s)
				assert.False(t, seen[s], "%q repeats", s)
				seen[s] = true

				back, err := pitchclass.ParseOperator(s)
				require.NoError(t, err, "ParseOperator(%q)", s)
				assertOperator(t, specs[i], back, fmt.Sprintf("ParseOperator(%q)", s))
			}
		}
		assert.Len(t, seen, 1+2+48+342)
	})
}

// K-27, E-4, E-6: ParseOperator accepts only canonical forms; a canonical
// form with a non-invertible m is ErrNotInvertible, checked after the text.
func TestParseOperatorErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		s    string
		want error
	}{
		{s: "T0M2 mod 12", want: pitchclass.ErrNotInvertible},
		{s: "T0M0 mod 12", want: pitchclass.ErrNotInvertible},
		{s: "T5M6 mod 12", want: pitchclass.ErrNotInvertible},
		{s: "T0M7 mod 9223372036854775807", want: pitchclass.ErrNotInvertible},
		{s: "T0M12 mod 12", want: pitchclass.ErrNotation},
		{s: "T12M2 mod 12", want: pitchclass.ErrNotation},
		{s: "T12M1 mod 12", want: pitchclass.ErrNotation},
		{s: "T0M13 mod 12", want: pitchclass.ErrNotation},
		{s: "T0M1 mod 0", want: pitchclass.ErrNotation},
		{s: "T0M0 mod 0", want: pitchclass.ErrNotation},
		{s: "T0M1 mod -1", want: pitchclass.ErrNotation},
		{s: "", want: pitchclass.ErrNotation},
		{s: "T5M7", want: pitchclass.ErrNotation},
		{s: "T5M7 mod", want: pitchclass.ErrNotation},
		{s: "T5M7 mod ", want: pitchclass.ErrNotation},
		{s: "T5 mod 12", want: pitchclass.ErrNotation},
		{s: "M7 mod 12", want: pitchclass.ErrNotation},
		{s: "M7T5 mod 12", want: pitchclass.ErrNotation},
		{s: "t5m7 mod 12", want: pitchclass.ErrNotation},
		{s: "T5m7 mod 12", want: pitchclass.ErrNotation},
		{s: "T5M7 MOD 12", want: pitchclass.ErrNotation},
		{s: "T5M7  mod 12", want: pitchclass.ErrNotation},
		{s: " T5M7 mod 12", want: pitchclass.ErrNotation},
		{s: "T5M7 mod 12 ", want: pitchclass.ErrNotation},
		{s: "T5 M7 mod 12", want: pitchclass.ErrNotation},
		{s: "T05M7 mod 12", want: pitchclass.ErrNotation},
		{s: "T5M07 mod 12", want: pitchclass.ErrNotation},
		{s: "T5M7 mod 012", want: pitchclass.ErrNotation},
		{s: "T-1M1 mod 12", want: pitchclass.ErrNotation},
		{s: "T+1M1 mod 12", want: pitchclass.ErrNotation},
		{s: "T1M-1 mod 12", want: pitchclass.ErrNotation},
		{s: "TM1 mod 12", want: pitchclass.ErrNotation},
		{s: "T1M mod 12", want: pitchclass.ErrNotation},
		{s: "T5I mod 12", want: pitchclass.ErrNotation},
		{s: "I5 mod 12", want: pitchclass.ErrNotation},
		{s: "T\uff15M7 mod 12", want: pitchclass.ErrNotation},
		{s: "T0M1 mod 9223372036854775808", want: pitchclass.ErrNotation},
		{s: "T9223372036854775808M1 mod 12", want: pitchclass.ErrNotation},
		{s: "Operator(invalid)", want: pitchclass.ErrNotation},
		{s: "5 mod 12", want: pitchclass.ErrNotation},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%q", tc.s), func(t *testing.T) {
			t.Parallel()

			f, err := pitchclass.ParseOperator(tc.s)
			assertOnlyError(t, err, tc.want, "ParseOperator")
			assert.Equal(t, pitchclass.Operator{}, f, "the zero Operator on an error")
			assertPanicsWithErrorIs(t, tc.want, func() { pitchclass.MustParseOperator(tc.s) }, "MustParseOperator")
		})
	}
}
