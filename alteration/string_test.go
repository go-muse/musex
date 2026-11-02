package alteration_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/alteration"

	"github.com/stretchr/testify/assert"
)

// String implements fmt.Stringer (architecture, KR-A8).
var _ fmt.Stringer = alteration.Alteration{}

// KR-A8: the fixed String form: "0" for natural; a sign and the number for an
// integer; a sign, the magnitude of the numerator, "/" and the denominator for
// a fraction. ASCII signs, no spaces, no leading zeros.
func TestString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   frac
		want string
	}{
		{in: frac{0, 1}, want: "0"},
		{in: frac{1, 1}, want: "+1"},
		{in: frac{-1, 1}, want: "-1"},
		{in: frac{2, 1}, want: "+2"},
		{in: frac{-2, 1}, want: "-2"},
		{in: frac{3, 1}, want: "+3"},
		{in: frac{1, 2}, want: "+1/2"},
		{in: frac{-1, 2}, want: "-1/2"},
		{in: frac{1, 3}, want: "+1/3"},
		{in: frac{3, 2}, want: "+3/2"},
		{in: frac{-3, 2}, want: "-3/2"},
		{in: frac{-7, 2}, want: "-7/2"},
		{in: frac{5, 6}, want: "+5/6"},
		{in: frac{10, 4}, want: "+5/2"},
		{in: frac{100, 1}, want: "+100"},
		{in: frac{bigM, 1}, want: "+16777216"},
		{in: frac{-bigM, 1}, want: "-16777216"},
		{in: frac{1, bigQ}, want: "+1/65536"},
		{in: frac{-1, bigQ}, want: "-1/65536"},
		{in: frac{bigM*bigQ - 1, bigQ}, want: "+1099511627775/65536"},
	}

	for _, tc := range cases {
		t.Run(tc.in.String(), func(t *testing.T) {
			t.Parallel()

			a := value(t, tc.in)

			assert.Equal(t, tc.want, a.String())
			assert.Equal(t, tc.want, contractString(tc.in.rat()), "oracle of the format")
			assert.Equal(t, a.String(), fmt.Sprint(a), "fmt uses String")
			assert.Equal(t, "["+a.String()+"]", fmt.Sprintf("[%v]", a), "the v verb uses String")
		})
	}

	t.Run("ready values", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "0", alteration.Natural().String())
		assert.Equal(t, "+1", alteration.Sharp().String())
		assert.Equal(t, "-1", alteration.Flat().String())
		assert.Equal(t, "+2", alteration.DoubleSharp().String())
		assert.Equal(t, "-2", alteration.DoubleFlat().String())
	})

	t.Run("all p and q in -12...12", func(t *testing.T) {
		t.Parallel()

		for p := -12; p <= 12; p++ {
			for q := -12; q <= 12; q++ {
				if q == 0 {
					continue
				}

				f := frac{p, q}
				assert.Equal(t, contractString(f.rat()), value(t, f).String(), "String of %s", f)
			}
		}
	})
}
