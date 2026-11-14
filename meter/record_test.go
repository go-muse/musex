package meter_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRecord covers О-3, Т-1, КП-24, КП-45 and the table of architecture 4.5.
func TestRecord(t *testing.T) {
	t.Parallel()
	defer guard(t)

	const e4, e5 = -4, -5 // markers of E-4 and E-5 in the table

	for _, c := range []struct {
		text        string
		form        meter.Form
		symbol      meter.Symbol
		parts       []string
		addends     []int
		numerator   int
		denominator int // e4 for E-4
		common      bool
		powerOfTwo  bool
		exponent    int // e4 for E-4, e5 for E-5
	}{
		{"5/8[2+3]", meter.FormSimple, 0, []string{"5/8"}, []int{5}, 5, 8, true, true, 3},
		{"C", meter.FormSymbol, meter.SymbolCommon, []string{"4/4"}, []int{4}, 4, 4, true, true, 2},
		{"C|", meter.FormSymbol, meter.SymbolCut, []string{"2/2"}, []int{2}, 2, 2, true, true, 1},
		{"C[3+1]", meter.FormSymbol, meter.SymbolCommon, []string{"4/4"}, []int{4}, 4, 4, true, true, 2},
		{"КП-24 (3+2+3)/8", meter.FormAdditive, 0, []string{"(3+2+3)/8"}, []int{3, 2, 3}, 8, 8, true, true, 3},
		{"КП-24 2/4+3/8", meter.FormSum, 0, []string{"2/4", "3/8"}, []int{2, 3}, 5, e4, false, true, e4},
		{"КП-24 6/8", meter.FormSimple, 0, []string{"6/8"}, []int{6}, 6, 8, true, true, 3},
		{"КП-24 4/3", meter.FormSimple, 0, []string{"4/3"}, []int{4}, 4, 3, true, false, e5},
		{"КП-24 2/4+2/3", meter.FormSum, 0, []string{"2/4", "2/3"}, []int{2, 2}, 4, e4, false, false, e4},
		{"3/1", meter.FormSimple, 0, []string{"3/1"}, []int{3}, 3, 1, true, true, 0},
		{"3/2048", meter.FormSimple, 0, []string{"3/2048"}, []int{3}, 3, 2048, true, true, 11},
		{"6/6", meter.FormSimple, 0, []string{"6/6"}, []int{6}, 6, 6, true, false, e5},
		{"КП-45 2/8+3/8", meter.FormSum, 0, []string{"2/8", "3/8"}, []int{2, 3}, 5, 8, true, true, 3},
		{"(2+3)/8+2/4", meter.FormSum, 0, []string{"(2+3)/8", "2/4"}, []int{2, 3, 2}, 7, e4, false, true, e4},
		{"three parts", meter.FormSum, 0, []string{"2/8", "(1+1)/8", "3/8"}, []int{2, 1, 1, 3}, 7, 8, true, true, 3},
	} {
		t.Run(c.text, func(t *testing.T) {
			t.Parallel()
			defer guard(t)

			m := mt(c.text)

			assert.Equal(t, c.form, noErr(m.Form()))

			s, ok, err := m.Symbol()
			require.NoError(t, err)
			assert.Equal(t, c.symbol, s)
			assert.Equal(t, c.symbol != 0, ok)

			assert.Equal(t, len(c.parts), noErr(m.PartCount()))
			assert.Equal(t, mts(c.parts...), noErr(m.Parts()))
			assert.Equal(t, c.addends, noErr(m.Addends()))
			assert.Equal(t, c.numerator, noErr(m.Numerator()))
			assert.Equal(t, c.numerator, noErr(m.UnitCount()), "Numerator equals UnitCount")
			assert.Equal(t, c.common, noErr(m.HasCommonDenominator()))
			assert.Equal(t, c.powerOfTwo, noErr(m.HasPowerOfTwoDenominators()))

			d, err := m.Denominator()
			if c.denominator == e4 {
				assertErr(t, err, meter.ErrNoCommonDenominator)
				assert.Zero(t, d)
			} else {
				require.NoError(t, err)
				assert.Equal(t, c.denominator, d)
			}

			x, err := m.DenominatorExponent()

			switch c.exponent {
			case e4:
				assertErr(t, err, meter.ErrNoCommonDenominator)
				assert.Zero(t, x)
			case e5:
				assertErr(t, err, meter.ErrUndefined)
				assert.Zero(t, x)
			default:
				require.NoError(t, err)
				assert.Equal(t, c.exponent, x)
			}
		})
	}
}

// TestRecordLargeDenominators covers П-3: any denominator up to math.MaxInt
// and the exponent of a power of two.
func TestRecordLargeDenominators(t *testing.T) {
	t.Parallel()
	defer guard(t)

	m := meter.MustNew(1, 1<<62)
	assert.Equal(t, 62, m.MustDenominatorExponent())
	assert.True(t, m.MustHasPowerOfTwoDenominators())

	big := meter.MustNew(1, math.MaxInt)
	assert.Equal(t, math.MaxInt, big.MustDenominator())
	assert.False(t, big.MustHasPowerOfTwoDenominators())

	_, err := big.DenominatorExponent()
	assertErr(t, err, meter.ErrUndefined)
}

// TestPartsRestoreSum covers architecture 4.5: NewSum(m.MustParts()...)
// equals m for every sum, and the parts have no explicit grouping.
func TestPartsRestoreSum(t *testing.T) {
	t.Parallel()
	defer guard(t)

	for _, m := range corpus() {
		parts := m.MustParts()
		require.NotEmpty(t, parts)

		for _, p := range parts {
			form := p.MustForm()
			assert.True(t, form == meter.FormSimple || form == meter.FormAdditive, "%v: part %v", m, p)
			assert.False(t, p.MustHasExplicitGrouping(), "%v: part %v", m, p)
		}

		if m.MustForm() == meter.FormSum {
			assert.Equal(t, m, meter.MustNewSum(parts...), "%v", m)
		} else {
			assert.Len(t, parts, 1, "%v", m)
		}

		sum := 0
		for _, a := range m.MustAddends() {
			sum += a
		}

		assert.Equal(t, m.MustNumerator(), sum, "%v: Numerator is the sum of Addends", m)
	}
}
