package meter_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// maxIntText is the decimal text of math.MaxInt.
func maxIntText() string { return strconv.Itoa(math.MaxInt) }

// TestNew covers О-1, П-3, П-9, КП-1, КП-27, КП-28, КП-40, КП-41 and the order
// of checks of New (architecture 4.2, 4.4).
func TestNew(t *testing.T) {
	t.Parallel()
	defer guard(t)

	ok := []struct {
		name string
		n, d int
		size string
	}{
		{"3/4", 3, 4, "3/4"},
		{"6/8", 6, 8, "3/4"},
		{"12/8", 12, 8, "3/2"},
		{"КП-1 4/3", 4, 3, "4/3"},
		{"КП-1 3/1", 3, 1, "3"},
		{"КП-41 1/1", 1, 1, "1"},
		{"КП-41 3/2048", 3, 2048, "3/2048"},
		{"КП-40 1024/8", 1024, 8, "128"},
		{"6/1024", 6, 1024, "3/512"},
		{"denominator MaxInt", 1, math.MaxInt, "1/" + maxIntText()},
		{"1024/MaxInt", 1024, math.MaxInt, "1024/" + maxIntText()},
	}

	for _, c := range ok {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m, err := meter.New(c.n, c.d)
			require.NoError(t, err)

			assert.True(t, m.IsSet())
			assert.Equal(t, meter.FormSimple, m.MustForm())
			assert.Equal(t, c.n, m.MustNumerator())
			assert.Equal(t, c.d, m.MustDenominator())
			assert.Equal(t, c.n, m.MustUnitCount())
			assert.Equal(t, []int{c.n}, m.MustAddends())
			assert.Equal(t, 1, m.MustPartCount())
			assert.Equal(t, durT(c.size), m.MustDuration())
			assert.False(t, m.MustHasExplicitGrouping())
		})
	}

	fail := []struct {
		name string
		n, d int
		want error
	}{
		{"КП-27 numerator 0", 0, 4, meter.ErrInvalidArgument},
		{"КП-27 numerator -3", -3, 4, meter.ErrInvalidArgument},
		{"numerator MinInt", math.MinInt, 4, meter.ErrInvalidArgument},
		{"КП-27 denominator 0", 3, 0, meter.ErrInvalidArgument},
		{"КП-27 denominator -4", 3, -4, meter.ErrInvalidArgument},
		{"denominator MinInt", 3, math.MinInt, meter.ErrInvalidArgument},
		{"КП-28 1025/8", 1025, 8, meter.ErrTooMany},
		{"numerator MaxInt", math.MaxInt, 4, meter.ErrTooMany},
		{"order: E-2 before E-8", 1025, 0, meter.ErrInvalidArgument},
		{"order: both invalid", 0, 0, meter.ErrInvalidArgument},
	}

	for _, c := range fail {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m, err := meter.New(c.n, c.d)
			assertErr(t, err, c.want)
			assert.Equal(t, meter.Meter{}, m, "the value on error is Meter{}")
		})
	}
}

// TestNewAdditive covers О-1, ИР-4, КП-1, КП-27, КП-28 and the order of
// checks of NewAdditive.
func TestNewAdditive(t *testing.T) {
	t.Parallel()
	defer guard(t)

	ok := []struct {
		name    string
		addends []int
		d       int
		size    string
	}{
		{"КП-1 (3+2+3)/8", []int{3, 2, 3}, 8, "1"},
		{"(2+3)/8", []int{2, 3}, 8, "5/8"},
		{"(3+3)/8", []int{3, 3}, 8, "3/4"},
		{"(1+1)/1", []int{1, 1}, 1, "2"},
		{"(4+4)/8", []int{4, 4}, 8, "1"},
		{"1024 units", []int{1000, 24}, 8, "128"},
		{"(2+2)/3", []int{2, 2}, 3, "4/3"},
	}

	for _, c := range ok {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m, err := meter.NewAdditive(c.addends, c.d)
			require.NoError(t, err)

			assert.Equal(t, meter.FormAdditive, m.MustForm())
			assert.Equal(t, c.addends, m.MustAddends())
			assert.Equal(t, c.d, m.MustDenominator())
			assert.Equal(t, 1, m.MustPartCount())
			assert.Equal(t, durT(c.size), m.MustDuration())
			assert.Equal(t, c.addends, m.MustGrouping(), "ИР-4: the groups are the addends")
			assert.False(t, m.MustHasExplicitGrouping())
		})
	}

	fail := []struct {
		name    string
		addends []int
		d       int
		want    error
	}{
		{"КП-27 nil", nil, 8, meter.ErrInvalidArgument},
		{"КП-27 empty", []int{}, 8, meter.ErrInvalidArgument},
		{"КП-27 one addend", []int{5}, 8, meter.ErrInvalidArgument},
		{"КП-27 addend 0", []int{0, 3}, 8, meter.ErrInvalidArgument},
		{"negative addend", []int{3, -1}, 8, meter.ErrInvalidArgument},
		{"denominator 0", []int{2, 3}, 0, meter.ErrInvalidArgument},
		{"denominator -8", []int{2, 3}, -8, meter.ErrInvalidArgument},
		{"КП-28 (1000+25)/8", []int{1000, 25}, 8, meter.ErrTooMany},
		{"sum overflows int", []int{math.MaxInt, math.MaxInt}, 8, meter.ErrTooMany},
		{"sum wraps to a small number", []int{math.MaxInt, math.MaxInt, 7}, 8, meter.ErrTooMany},
		{"order: E-2 denominator before E-8", []int{1000, 25}, 0, meter.ErrInvalidArgument},
		{"order: E-2 addend before E-8", []int{1025, 0}, 8, meter.ErrInvalidArgument},
	}

	for _, c := range fail {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m, err := meter.NewAdditive(c.addends, c.d)
			assertErr(t, err, c.want)
			assert.Equal(t, meter.Meter{}, m)
		})
	}
}

// TestNewSum covers О-1, П-5, П-10, ИР-1, КП-1, КП-27, КП-28, КП-39 and the
// order of checks of NewSum.
func TestNewSum(t *testing.T) {
	t.Parallel()
	defer guard(t)

	ok := []struct {
		name    string
		parts   []string
		size    string
		addends []int
	}{
		{"КП-1 2/4+3/8", []string{"2/4", "3/8"}, "7/8", []int{2, 3}},
		{"2/8+3/8", []string{"2/8", "3/8"}, "5/8", []int{2, 3}},
		{"(2+3)/8+2/4", []string{"(2+3)/8", "2/4"}, "9/8", []int{2, 3, 2}},
		{"three parts", []string{"2/8", "2/8", "2/8"}, "3/4", []int{2, 2, 2}},
		{"1024 units", []string{"512/8", "512/8"}, "128", []int{512, 512}},
		{"irrational 2/4+2/3", []string{"2/4", "2/3"}, "7/6", []int{2, 2}},
	}

	for _, c := range ok {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			parts := mts(c.parts...)

			m, err := meter.NewSum(parts...)
			require.NoError(t, err)

			assert.Equal(t, meter.FormSum, m.MustForm())
			assert.Equal(t, durT(c.size), m.MustDuration())
			assert.Equal(t, c.addends, m.MustAddends())
			assert.Equal(t, len(parts), m.MustPartCount())
			assert.Equal(t, parts, m.MustParts())
		})
	}

	t.Run("КП-39 four parts are in range", func(t *testing.T) {
		t.Parallel()
		defer guard(t)

		_, err := meter.NewSum(unitParts(kp39Denominators()[:4])...)
		assert.NoError(t, err)
	})

	var unset meter.Meter

	fail := []struct {
		name   string
		parts  []meter.Meter
		want   error
		causes []error
	}{
		{"no parts", nil, meter.ErrInvalidArgument, nil},
		{"КП-27 one part", mts("2/4"), meter.ErrInvalidArgument, nil},
		{"КП-27 unset part is E-2, not E-1", []meter.Meter{mt("2/4"), unset}, meter.ErrInvalidArgument, nil},
		{"unset first part", []meter.Meter{unset, mt("2/4")}, meter.ErrInvalidArgument, nil},
		{"two unset parts", []meter.Meter{unset, unset}, meter.ErrInvalidArgument, nil},
		{"symbol part", mts("2/4", "C"), meter.ErrInvalidArgument, nil},
		{"cut symbol part", mts("C|", "2/4"), meter.ErrInvalidArgument, nil},
		{"sum part", mts("2/4+3/8", "2/4"), meter.ErrInvalidArgument, nil},
		{"part with explicit grouping", mts("2/4", "5/8[2+3]"), meter.ErrInvalidArgument, nil},
		{"КП-28 1024/8+1/4", mts("1024/8", "1/4"), meter.ErrTooMany, nil},
		{"513+512 units", mts("513/8", "512/8"), meter.ErrTooMany, nil},
		{"КП-39 five parts", unitParts(kp39Denominators()), meter.ErrOutOfRange, []error{fraction.ErrOutOfRange}},
		{"order: E-2 part before E-8", []meter.Meter{mt("1024/8"), mt("1/4"), unset}, meter.ErrInvalidArgument, nil},
		{"one unset part", []meter.Meter{unset}, meter.ErrInvalidArgument, nil},
		{"order: E-8 before E-10", kp39TooMany(), meter.ErrTooMany, nil},
	}

	for _, c := range fail {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m, err := meter.NewSum(c.parts...)
			assertErr(t, err, c.want, c.causes...)
			assert.Equal(t, meter.Meter{}, m)
		})
	}
}

// kp39TooMany is a sum of five parts with the denominators of КП-39 and 205
// units each: 1025 units in all, and a size outside the range of
// common/fraction. E-8 is checked before E-10 (architecture 4.2).
func kp39TooMany() []meter.Meter {
	dens := kp39Denominators()
	out := make([]meter.Meter, len(dens))

	for i, d := range dens {
		out[i] = meter.MustNew(205, d)
	}

	return out
}

// TestNewSymbol covers О-1, П-2, КП-1, КП-27.
func TestNewSymbol(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct {
		name     string
		s        meter.Symbol
		n, d     int
		grouping []int
	}{
		{"C", meter.SymbolCommon, 4, 4, []int{2, 2}},
		{"C|", meter.SymbolCut, 2, 2, []int{2}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m, err := meter.NewSymbol(c.s)
			require.NoError(t, err)

			assert.Equal(t, meter.FormSymbol, m.MustForm())

			s, ok := m.MustSymbol()
			assert.True(t, ok)
			assert.Equal(t, c.s, s)
			assert.Equal(t, c.n, m.MustNumerator())
			assert.Equal(t, c.d, m.MustDenominator())
			assert.Equal(t, durT("1"), m.MustDuration())
			assert.Equal(t, c.grouping, m.MustGrouping())
			assert.False(t, m.MustHasExplicitGrouping())
		})
	}

	for _, s := range []meter.Symbol{0, 3, 255} {
		t.Run("КП-27 invalid "+strconv.Itoa(int(s)), func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m, err := meter.NewSymbol(s)
			assertErr(t, err, meter.ErrInvalidArgument)
			assert.Equal(t, meter.Meter{}, m)
		})
	}
}

// TestReadyMade covers О-2 and КП-25: ready-made meters equal the values
// created by О-1 and are the same on every call.
func TestReadyMade(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, c := range []struct {
		name string
		f    func() meter.Meter
		want meter.Meter
		text string
	}{
		{"CommonTime", meter.CommonTime, meter.MustNewSymbol(meter.SymbolCommon), "C"},
		{"CutTime", meter.CutTime, meter.MustNewSymbol(meter.SymbolCut), "C|"},
		{"TwoFour", meter.TwoFour, meter.MustNew(2, 4), "2/4"},
		{"ThreeFour", meter.ThreeFour, meter.MustNew(3, 4), "3/4"},
		{"FourFour", meter.FourFour, meter.MustNew(4, 4), "4/4"},
		{"TwoTwo", meter.TwoTwo, meter.MustNew(2, 2), "2/2"},
		{"ThreeEight", meter.ThreeEight, meter.MustNew(3, 8), "3/8"},
		{"SixEight", meter.SixEight, meter.MustNew(6, 8), "6/8"},
		{"NineEight", meter.NineEight, meter.MustNew(9, 8), "9/8"},
		{"TwelveEight", meter.TwelveEight, meter.MustNew(12, 8), "12/8"},
		{"SixFour", meter.SixFour, meter.MustNew(6, 4), "6/4"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			got := c.f()
			assert.True(t, got.IsSet())
			assert.True(t, got.Equal(c.want))
			assert.Equal(t, c.want, got)
			assert.Equal(t, got, c.f(), "the same value on every call")
			assert.Equal(t, c.text, got.String())
		})
	}
}
