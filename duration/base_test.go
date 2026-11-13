package duration_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/duration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// invalidBases are values of Base outside the 14 constants (architecture 4.3).
var invalidBases = []duration.Base{0, 15, 16, 200, 255}

// TestBaseConstants covers P-1 and architecture 4.3: 14 constants from the
// longest, consecutive from 1, with the values of KP-4 and business-spec 2.2.
func TestBaseConstants(t *testing.T) {
	t.Parallel()

	require.Len(t, baseTable, 14)

	for i, row := range baseTable {
		t.Run(row.text, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, duration.Base(i+1), row.base, "constants are consecutive from 1")
			assert.True(t, row.base.IsValid())

			got, err := row.base.Duration()
			require.NoError(t, err)
			assert.Equal(t, dur(int(row.num), int(row.den)), got)
			assert.Equal(t, 0, ratOfDuration(got).Cmp(rat(row.num, row.den)))
		})
	}

	assert.Equal(t, 1, int(duration.Maxima))
	assert.Equal(t, 14, int(duration.ThousandTwentyFourth))
	assert.Equal(t, 1024, duration.MaxParts)
}

// TestBaseSeriesHalves covers note-value inv-1 for the series: each base is
// twice the next one.
func TestBaseSeriesHalves(t *testing.T) {
	t.Parallel()

	for i := 0; i+1 < len(baseTable); i++ {
		longer := baseTable[i].base.MustDuration()
		shorter := baseTable[i+1].base.MustDuration()
		assert.Equal(t, longer, shorter.MustMulInt(2), "%s = 2 × %s", baseTable[i].text, baseTable[i+1].text)
	}
}

// TestBases covers architecture 4.3 and 3.2: all 14 in declaration order, a
// new slice of length and capacity 14 that belongs to the caller.
func TestBases(t *testing.T) {
	t.Parallel()

	first := duration.Bases()
	require.Len(t, first, 14)
	assert.Equal(t, 14, cap(first))

	for i, row := range baseTable {
		assert.Equal(t, row.base, first[i])
	}

	first[0] = duration.Base(0)
	first[13] = duration.Quarter

	second := duration.Bases()
	assert.Equal(t, duration.Maxima, second[0], "changing a result does not change later results")
	assert.Equal(t, duration.ThousandTwentyFourth, second[13])
}

// TestBaseIsValid covers architecture 4.3: only the 14 constants are valid.
func TestBaseIsValid(t *testing.T) {
	t.Parallel()

	for v := range 256 {
		b := duration.Base(v)
		assert.Equal(t, v >= 1 && v <= 14, b.IsValid(), "Base(%d)", v)
	}
}

// TestBaseEqual covers architecture 4.3 and KR-D14: Equal matches == for all
// values, including invalid ones, without an error.
func TestBaseEqual(t *testing.T) {
	t.Parallel()

	values := []duration.Base{255}
	for v := range 16 {
		values = append(values, duration.Base(v))
	}

	for _, a := range values {
		for _, b := range values {
			assert.Equal(t, a == b, a.Equal(b), "Base(%d).Equal(Base(%d))", a, b)
		}
	}

	assert.True(t, duration.Base(0).Equal(duration.Base(0)))
	assert.False(t, duration.Base(0).Equal(duration.Base(15)))
}

// TestBaseInvalid covers architecture 4.2 and 4.3: every operation of an
// invalid base fails with E-2 and zero results; String is "Base(invalid)".
func TestBaseInvalid(t *testing.T) {
	t.Parallel()

	for _, b := range invalidBases {
		d, err := b.Duration()
		assertErr(t, err, duration.ErrInvalidArgument)
		assert.Equal(t, duration.Duration{}, d)

		w, err := b.Written()
		assertErr(t, err, duration.ErrInvalidArgument)
		assert.Equal(t, duration.Written{}, w)

		for _, f := range []func() (duration.Base, error){
			b.Longer, b.Shorter,
			func() (duration.Base, error) { return b.Shift(0) },
			func() (duration.Base, error) { return b.Shift(1) },
			func() (duration.Base, error) { return b.Shift(math.MaxInt) },
		} {
			got, err := f()
			assertErr(t, err, duration.ErrInvalidArgument)
			assert.Equal(t, duration.Base(0), got)
		}

		for _, f := range []func() (string, bool, error){b.Name, b.BritishName, b.RussianName} {
			name, ok, err := f()
			assertErr(t, err, duration.ErrInvalidArgument)
			assert.Empty(t, name)
			assert.False(t, ok)
		}

		assert.Equal(t, "Base(invalid)", b.String())

		text, err := b.MarshalText()
		assertErr(t, err, duration.ErrInvalidArgument)
		assert.Nil(t, text)

		dst := []byte("prefix")
		out, err := b.AppendText(dst)
		assertErr(t, err, duration.ErrInvalidArgument)
		assert.Equal(t, []byte("prefix"), out, "AppendText returns dst unchanged")
	}
}

// TestBaseWritten covers architecture 4.3: the basic written duration of a
// base, without dots and tuplets.
func TestBaseWritten(t *testing.T) {
	t.Parallel()

	for _, row := range baseTable {
		w, err := row.base.Written()
		require.NoError(t, err)

		assert.Equal(t, wr(row.base, 0), w)
		assert.Equal(t, row.base, w.MustBase())
		assert.Equal(t, 0, w.MustDots())
		assert.Nil(t, w.MustChain())
		assert.True(t, w.MustIsBasic())
		assert.Equal(t, row.base.MustDuration(), w.MustDuration())
	}
}

// TestBaseLongerShorter covers O-4, KP-34, and KP-50: neighbors twice as
// long and twice as short; E-6 beyond the ends of the series.
func TestBaseLongerShorter(t *testing.T) {
	t.Parallel()

	for i, row := range baseTable {
		longer, err := row.base.Longer()
		if i == 0 {
			assertErr(t, err, duration.ErrOutsideSeries)
			assert.Equal(t, duration.Base(0), longer)
		} else {
			require.NoError(t, err)
			assert.Equal(t, baseTable[i-1].base, longer)
			assert.Equal(t, row.base.MustDuration().MustMulInt(2), longer.MustDuration())
		}

		shorter, err := row.base.Shorter()
		if i == len(baseTable)-1 {
			assertErr(t, err, duration.ErrOutsideSeries)
			assert.Equal(t, duration.Base(0), shorter)
		} else {
			require.NoError(t, err)
			assert.Equal(t, baseTable[i+1].base, shorter)
			assert.Equal(t, row.base.MustDuration().MustDivInt(2), shorter.MustDuration())
		}
	}

	assert.Equal(t, duration.Breve, duration.Whole.MustLonger(), "KP-50")
	assert.Equal(t, duration.Half, duration.Whole.MustShorter(), "KP-50")
}

// TestBaseShift covers O-6 and architecture 4.3: steps > 0 are longer, 0 is
// the same base, E-6 outside the series for any int without overflow.
func TestBaseShift(t *testing.T) {
	t.Parallel()

	for i, row := range baseTable {
		for steps := -20; steps <= 20; steps++ {
			got, err := row.base.Shift(steps)

			j := i - steps
			if j < 0 || j >= len(baseTable) {
				assertErr(t, err, duration.ErrOutsideSeries)
				assert.Equal(t, duration.Base(0), got)

				continue
			}

			require.NoError(t, err)
			assert.Equal(t, baseTable[j].base, got, "%s shifted by %d", row.text, steps)
		}

		for _, steps := range []int{math.MaxInt, math.MinInt, math.MaxInt - 1, math.MinInt + 1, 1 << 40, -(1 << 40)} {
			got, err := row.base.Shift(steps)
			assertErr(t, err, duration.ErrOutsideSeries)
			assert.Equal(t, duration.Base(0), got)
		}
	}

	assert.Equal(t, duration.Maxima, duration.Quarter.MustShift(5))
	assert.Equal(t, duration.ThousandTwentyFourth, duration.Quarter.MustShift(-8))
}

// TestBaseOf covers O-4, KP-45, and architecture 4.3: the base of a value,
// E-2 for a value that is not a base, E-1 for an unset value.
func TestBaseOf(t *testing.T) {
	t.Parallel()

	for _, row := range baseTable {
		got, err := duration.BaseOf(dur(int(row.num), int(row.den)))
		require.NoError(t, err)
		assert.Equal(t, row.base, got)
	}

	for _, d := range []duration.Duration{dur(3, 8), dur(16, 1), dur(1, 2048), dur(3, 1), dur(1, 3), dur(3, 2), dur(32, 1)} {
		got, err := duration.BaseOf(d)
		assertErr(t, err, duration.ErrInvalidArgument)
		assert.Equal(t, duration.Base(0), got, "BaseOf(%s)", d)
	}

	got, err := duration.BaseOf(duration.Duration{})
	assertErr(t, err, duration.ErrUnset)
	assert.Equal(t, duration.Base(0), got)
}

// TestBaseNames covers O-15, P-6, KP-22, and architecture 4.11: names only
// from the table; "no name" is ok == false without an error.
func TestBaseNames(t *testing.T) {
	t.Parallel()

	for _, row := range baseTable {
		t.Run(row.text, func(t *testing.T) {
			t.Parallel()

			for _, c := range []struct {
				kind string
				want string
				get  func() (string, bool, error)
				must func() (string, bool)
			}{
				{"Name", row.us, row.base.Name, row.base.MustName},
				{"BritishName", row.uk, row.base.BritishName, row.base.MustBritishName},
				{"RussianName", row.ru, row.base.RussianName, row.base.MustRussianName},
			} {
				name, ok, err := c.get()
				require.NoError(t, err, c.kind)
				assert.Equal(t, c.want, name, c.kind)
				assert.Equal(t, c.want != "", ok, c.kind)

				mustName, mustOK := c.must()
				assert.Equal(t, name, mustName, c.kind)
				assert.Equal(t, ok, mustOK, c.kind)
			}
		})
	}

	// KP-22.
	name, ok := duration.HundredTwentyEighth.MustRussianName()
	assert.Equal(t, "стодвадцатьвосьмая", name)
	assert.True(t, ok)

	_, ok = duration.HundredTwentyEighth.MustName()
	assert.False(t, ok, "the Go identifier is not a name")

	for _, f := range []func() (string, bool){
		duration.TwoHundredFiftySixth.MustName,
		duration.TwoHundredFiftySixth.MustBritishName,
		duration.TwoHundredFiftySixth.MustRussianName,
	} {
		name, ok := f()
		assert.Empty(t, name)
		assert.False(t, ok)
	}
}

// TestBaseString covers architecture 4.12: the text of each base.
func TestBaseString(t *testing.T) {
	t.Parallel()

	for _, row := range baseTable {
		assert.Equal(t, row.text, row.base.String())

		text, err := row.base.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, []byte(row.text), text)
	}
}
