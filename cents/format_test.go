package cents_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
)

// C-23: fmt.Formatter (architecture, "fmt and machine form").
func TestFormat(t *testing.T) {
	t.Parallel()

	r32, r23 := noErr(cents.FromRatio(3, 2))(t), noErr(cents.FromRatio(2, 3))(t)
	c700, c71 := noErr(cents.FromInt(700))(t), noErr(cents.FromFraction(71, 10))(t)
	eighth := noErr(cents.FromFraction(1, 8))(t)
	e119 := noErr(cents.Step(1, 19))(t)

	cases := []struct {
		name   string
		format string
		c      cents.Cents
		want   string
	}{
		{"%v", "%v", c700, "700"},
		{"%s", "%s", c700, "700"},
		{"%v ratio", "%v", r32, "3:2"},
		{"%v fraction", "%v", e119, "1200/19 cents"},
		{"%+v", "%+v", c71, "+7.1"},
		{"%+v ratio", "%+v", r32, "3:2"},
		{"%+v fraction", "%+v", e119, "+1200/19 cents"},
		{"%+v zero", "%+v", cents.Zero(), "0"},
		{"%+s", "%+s", c71, "+7.1"},
		{"%q", "%q", e119, strconv.Quote("1200/19 cents")},
		{"%q ratio", "%q", r32, `"3:2"`},
		{"%f", "%f", r32, "701.955001"},
		{"%F", "%F", r32, "701.955001"},
		{"%.3f", "%.3f", r32, "701.955"},
		{"%.0f", "%.0f", r32, "702"},
		{"%+.1f", "%+.1f", c71, "+7.1"},
		{"%+.1f negative", "%+.1f", r23, "-702.0"},
		{"%.3f negative", "%.3f", r23, "-701.955"},
		{"%.2f half", "%.2f", eighth, "0.13"},
		{"%.2f negative half", "%.2f", eighth.Neg(), "-0.13"},
		{"%f zero", "%f", cents.Zero(), "0.000000"},
		{"%+f zero", "%+f", cents.Zero(), "0.000000"},
		{"%.3f rounded to zero", "%.3f", noErr(cents.FromFraction(-4, 10000))(t), "0.000"},
		{"%12.3f", "%12.3f", r32, "     701.955"},
		{"%-12.3f|", "%-12.3f|", r32, "701.955     |"},
		{"%10.3f negative", "%10.3f", r23, "  -701.955"},
		{"%-12v|", "%-12v|", r32, "3:2         |"},
		{"%-6v|", "%-6v|", c700, "700   |"},
		{"%6v", "%6v", c700, "   700"},
		{"%2v shorter than the text", "%2v", c700, "700"},
		{"%8q", "%8q", r32, `   "3:2"`},
		{"%16s", "%16s", e119, "   1200/19 cents"},
		// Ignored flags and precision (N-6).
		{"%09.3f", "%09.3f", r32, "  701.955"},
		{"%#v", "%#v", r32, "3:2"},
		{"% v", "% v", c71, "7.1"},
		{"%+q", "%+q", r32, `"3:2"`},
		{"%.3v", "%.3v", e119, "1200/19 cents"},
		{"%.3s", "%.3s", e119, "1200/19 cents"},
		{"%.1q", "%.1q", r32, `"3:2"`},
		{"%#f", "%#f", c700, "700.000000"},
		{"% f", "% f", c71, "7.100000"},
		// Other verbs: exactly %!verb(cents.Cents=String()), without width,
		// precision and flags (N-9).
		{"%d", "%d", c700, "%!d(cents.Cents=700)"},
		{"%5d", "%5d", c700, "%!d(cents.Cents=700)"},
		{"%-5.2d", "%-5.2d", c700, "%!d(cents.Cents=700)"},
		{"%e", "%e", c700, "%!e(cents.Cents=700)"},
		{"%g", "%g", r32, "%!g(cents.Cents=3:2)"},
		{"%x", "%x", c700, "%!x(cents.Cents=700)"},
		{"%t", "%t", e119, "%!t(cents.Cents=1200/19 cents)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, fmt.Sprintf(tc.format, tc.c))
		})
	}

	t.Run("for all S", func(t *testing.T) {
		t.Parallel()

		for _, a := range setS(t) {
			assert.Equal(t, a.v.String(), fmt.Sprint(a.v), a.name)
			assert.Equal(t, a.v.String(), fmt.Sprintf("%v", a.v), a.name)
			assert.Equal(t, a.v.String(), fmt.Sprintf("%s", a.v), a.name)
			assert.Equal(t, a.v.Text(cents.WithPlus), fmt.Sprintf("%+v", a.v), a.name)
			assert.Equal(t, strconv.Quote(a.v.String()), fmt.Sprintf("%q", a.v), a.name)
			assert.Equal(t, a.v.Decimal(6, 0), fmt.Sprintf("%f", a.v), a.name)
			assert.Equal(t, a.v.Decimal(3, cents.WithPlus), fmt.Sprintf("%+.3f", a.v), a.name)
			assert.Equal(t, a.v.Decimal(0, 0), fmt.Sprintf("%.0f", a.v), a.name)
		}
	})

	t.Run("in containers", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "[700 3:2]", fmt.Sprint([]cents.Cents{c700, r32}))
		assert.Equal(t, "{3:2}", fmt.Sprintf("%v", struct{ C cents.Cents }{r32}))
		assert.Equal(t, "3:2", fmt.Sprintf("%v", &r32), "methods of *Cents")
		assert.Equal(t, "map[700:a]", fmt.Sprint(map[cents.Cents]string{c700: "a"}))
	})
}
