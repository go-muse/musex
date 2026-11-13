package frequency_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRatio_ControlExamples(t *testing.T) {
	for _, tt := range []struct {
		name   string
		f1, f2 float64
		want   string
	}{
		{"K-38 just fifth", 440, 660, "3/2"},
		{"K-38 measured deviation", 440, 442, "221/220"},
		{"K-38 decimals", 0.3, 0.1, "1/3"},
		{"reverse", 660, 440, "2/3"},
		{"unison", 440, 440, "1/1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Ratio(tt.f1, tt.f2)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.want, got.RatString())
			want := new(big.Rat).Quo(decimalRat(tt.f2), decimalRat(tt.f1))
			assert.Zero(t, got.Cmp(want))
			product := new(big.Rat).Mul(got, decimalRat(tt.f1))
			assert.Zero(t, product.Cmp(decimalRat(tt.f2)))
			assert.Equal(t, tt.want, frequency.MustRatio(tt.f1, tt.f2).RatString())
		})
	}
}

func TestRatio_DecimalInputDiffersFromBinaryValue(t *testing.T) {
	sum := binarySum()
	expectedDecimal, ok := new(big.Rat).SetString("0.30000000000000004")
	require.True(t, ok)
	assert.Zero(t, decimalRat(sum).Cmp(expectedDecimal))
	assert.NotEqual(t, decimalRat(sum).RatString(), rational(sum).RatString())

	got, err := frequency.Ratio(sum, 0.1)
	require.NoError(t, err)
	want := new(big.Rat).Quo(decimalRat(0.1), decimalRat(sum))
	assert.Zero(t, got.Cmp(want))
	assert.NotZero(t, got.Cmp(big.NewRat(1, 3)))
}

func TestRatio_ResultOwnership(t *testing.T) {
	first, err := frequency.Ratio(440, 660)
	require.NoError(t, err)
	first.SetInt64(100)
	second, err := frequency.Ratio(440, 660)
	require.NoError(t, err)
	assert.Equal(t, "3/2", second.RatString())
	assert.NotSame(t, first, second)

	mustFirst := frequency.MustRatio(440, 660)
	mustFirst.SetInt64(200)
	mustSecond := frequency.MustRatio(440, 660)
	assert.Equal(t, "3/2", mustSecond.RatString())
	assert.NotSame(t, mustFirst, mustSecond)
}

func TestRatio_InvalidFrequency(t *testing.T) {
	for _, invalid := range invalidFrequencyCases() {
		t.Run(invalid.name+"/first", func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Ratio(invalid.f, 440)
			assert.Nil(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
		t.Run(invalid.name+"/second", func(t *testing.T) {
			t.Parallel()
			got, err := frequency.Ratio(440, invalid.f)
			assert.Nil(t, got)
			requireErrorChain(t, err, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument)
		})
	}
}
