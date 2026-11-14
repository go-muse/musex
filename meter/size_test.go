package meter_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDuration covers О-4, ИР-1, Т-2, КП-1, КП-40, КП-41.
func TestDuration(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct{ text, size string }{
		{"3/4", "3/4"},
		{"6/8", "3/4"},
		{"12/8", "3/2"},
		{"(3+2+3)/8", "1"},
		{"C", "1"},
		{"C|", "1"},
		{"C[3+1]", "1"},
		{"2/4+3/8", "7/8"},
		{"4/3", "4/3"},
		{"3/1", "3"},
		{"1024/8", "128"},
		{"3/2048", "3/2048"},
		{"5/8[2+3]", "5/8"},
		{"2/4+2/3", "7/6"},
		{"1/1+1/2+1/4", "7/4"},
	} {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got, err := mt(c.text).Duration()
			require.NoError(t, err)
			assert.Equal(t, durT(c.size), got)
		})
	}
}

// TestQuoRem covers О-16, Т-2, КП-19, КП-38, КП-52 and architecture 4.6:
// the result equals Duration.QuoRem of package duration.
func TestQuoRem(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct {
		name, text, v string
		quo           int
		rem           string
	}{
		{"КП-19 6/8 by 1/8", "6/8", "1/8", 6, "0"},
		{"КП-19 6/8 by 3/8", "6/8", "3/8", 2, "0"},
		{"КП-19 6/8 by 1/2", "6/8", "1/2", 1, "1/4"},
		{"КП-19 3/4 by 1/4", "3/4", "1/4", 3, "0"},
		{"longer than the measure", "3/4", "1", 0, "3/4"},
		{"sum by 1/4", "2/4+3/8", "1/4", 3, "1/8"},
		{"4/3 by 1/2", "4/3", "1/2", 2, "1/3"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m := mt(c.text)

			quo, rem, err := m.QuoRem(durT(c.v))
			require.NoError(t, err)
			assert.Equal(t, c.quo, quo)
			assert.True(t, rem.Equal(fracRat(ratText(c.rem))), "rem = %v, want %s", rem, c.rem)
			assert.True(t, rem.IsReduced(), "rem is reduced")

			wantQuo, wantRem, err := m.MustDuration().QuoRem(durT(c.v))
			require.NoError(t, err)
			assert.Equal(t, wantQuo, quo)
			assert.Equal(t, wantRem, rem, "the same result as duration")
		})
	}

	t.Run("КП-38 unset v", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		quo, rem, err := meter.ThreeFour().QuoRem(duration.Duration{})
		assertErr(t, err, meter.ErrInvalidArgument, duration.ErrUnset)
		assert.Zero(t, quo)
		assert.Equal(t, fraction.Fraction{}, rem)
	})

	t.Run("КП-52 quotient 2^200", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		v := durRat(new(big.Rat).SetFrac(big.NewInt(1), pow2(200)))

		quo, rem, err := meter.FourFour().QuoRem(v)
		assertErr(t, err, meter.ErrOutOfRange, fraction.ErrNotRepresentable, duration.ErrOutOfRange)
		assert.Zero(t, quo)
		assert.Equal(t, fraction.Fraction{}, rem)
	})

	t.Run("quotient of 1024/1 by 1/(2^255-1)", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		v := durRat(new(big.Rat).SetFrac(big.NewInt(1), maxComponent()))

		_, _, err := meter.MustNew(1024, 1).QuoRem(v)
		assertErr(t, err, meter.ErrOutOfRange, fraction.ErrNotRepresentable, duration.ErrOutOfRange)
	})

	t.Run("order: quotient before remainder", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		// v = 1/(2^127−1) gives a quotient of about 2^81, above math.MaxInt,
		// and a remainder with a denominator of about 2^347, outside the
		// range. The quotient is checked first (architecture 4.2).
		m := meter.MustNewSum(unitParts(kp39Denominators()[:4])...)
		v := durRat(new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Sub(pow2(127), big.NewInt(1))))

		_, _, err := m.QuoRem(v)
		assertErr(t, err, meter.ErrOutOfRange, fraction.ErrNotRepresentable, duration.ErrOutOfRange)
		assert.NotErrorIs(t, err, fraction.ErrOutOfRange)
	})

	t.Run("remainder out of range", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		// The size has a denominator of about 2^220 and v = 1/(2^103−1) is
		// coprime with it, so the quotient fits in int while the remainder
		// needs a denominator above 2^255.
		m := meter.MustNewSum(unitParts(kp39Denominators()[:4])...)
		v := durRat(new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Sub(pow2(103), big.NewInt(1))))

		_, _, refErr := m.MustDuration().QuoRem(v)
		require.ErrorIs(t, refErr, fraction.ErrOutOfRange, "the input gives a remainder out of range in duration")

		_, _, err := m.QuoRem(v)
		assertErr(t, err, meter.ErrOutOfRange, fraction.ErrOutOfRange, duration.ErrOutOfRange)
		assert.NotErrorIs(t, err, fraction.ErrNotRepresentable)
	})
}
