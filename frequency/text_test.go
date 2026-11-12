package frequency_test

import (
	jsonv1 "encoding/json"
	jsonv2 "encoding/json/v2"
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func canonicalTextCases() []struct {
	name string
	c    frequency.Criterion
	text string
} {
	return []struct {
		name string
		c    frequency.Criterion
		text string
	}{
		{"unset", frequency.Criterion{}, "unset"},
		{"strict", frequency.Strict(), "strict"},
		{"zero cents", frequency.MustWithinCents(cents.Zero()), "cents(0)"},
		{"five cents", centsCriterion(5), "cents(5)"},
		{"decimal cents", frequency.MustWithinCents(cents.MustParse("3.930158")), "cents(3.930158)"},
		{"fraction cents", frequency.MustWithinCents(cents.MustFromFraction(1, 3)), "cents(1/3 cents)"},
		{"step cents", frequency.MustWithinCents(cents.MustStep(1, 19)), "cents(1200/19 cents)"},
		{"ratio cents", frequency.MustWithinCents(cents.MustFromRatio(81, 80)), "cents(81:80)"},
		{"power cents", frequency.MustWithinCents(cents.MustParse("5^(1/4)")), "cents(5^(1/4))"},
		{"zero hertz", frequency.MustWithinHertz(0), "hertz(0)"},
		{"negative zero hertz", frequency.MustWithinHertz(math.Copysign(0, -1)), "hertz(0)"},
		{"hertz", frequency.MustWithinHertz(1.5), "hertz(1.5)"},
		{"decimal hertz", frequency.MustWithinHertz(0.1), "hertz(0.1)"},
		{"subnormal hertz", frequency.MustWithinHertz(math.SmallestNonzeroFloat64), "hertz(5e-324)"},
		{"maximum hertz", frequency.MustWithinHertz(math.MaxFloat64), "hertz(1.7976931348623157e+308)"},
		{"exponent hertz", frequency.MustWithinHertz(1e6), "hertz(1e+06)"},
		{"relative", frequency.MustWithinRelative(0.01), "relative(0.01)"},
		{"subnormal relative", frequency.MustWithinRelative(math.SmallestNonzeroFloat64), "relative(5e-324)"},
		{"all", frequency.MustAllOf(centsCriterion(1), frequency.MustWithinHertz(1)), "all(cents(1), hertz(1))"},
		{"all reversed", frequency.MustAllOf(frequency.MustWithinHertz(1), centsCriterion(1)), "all(hertz(1), cents(1))"},
		{"nested", frequency.MustAnyOf(centsCriterion(1), frequency.MustAllOf(frequency.MustWithinHertz(1))), "any(cents(1), all(hertz(1)))"},
		{"deeply nested", frequency.MustAnyOf(frequency.MustAllOf(frequency.Strict(), centsCriterion(2)), frequency.MustWithinHertz(0.5)), "any(all(strict, cents(2)), hertz(0.5))"},
	}
}

func TestString_CanonicalAndRoundTrip(t *testing.T) {
	for _, tt := range canonicalTextCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.text, tt.c.String())
			parsed, err := frequency.Parse(tt.text)
			require.NoError(t, err)
			requireCriterionEqual(t, tt.c, parsed)
			assert.Equal(t, tt.text, parsed.String())
			assert.Equal(t, tt.text, frequency.MustParse(tt.text).String())
		})
	}
}

func TestString_ControlComparisons(t *testing.T) {
	zero := frequency.MustWithinHertz(0)
	negativeZero := frequency.MustWithinHertz(math.Copysign(0, -1))
	assert.Equal(t, zero.String(), negativeZero.String())
	assert.NotEqual(t, centsCriterion(1).String(), frequency.MustWithinHertz(1).String())
	assert.NotEqual(t, frequency.Strict().String(), zero.String())
	assert.NotEqual(t,
		frequency.MustAllOf(centsCriterion(1), frequency.MustWithinHertz(1)).String(),
		frequency.MustAllOf(frequency.MustWithinHertz(1), centsCriterion(1)).String())
}

func TestParse_NonCanonicalSuccess(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"cents(+5)", "cents(5)"},
		{"cents(5 cents)", "cents(5)"},
		{"cents(5¢)", "cents(5)"},
		{"cents(6:4)", "cents(3:2)"},
		{"cents(-0)", "cents(0)"},
		{"cents(0.10)", "cents(0.1)"},
		{"cents(9000000)", "cents(9000000)"},
		{"cents(1200/19 cents)", "cents(1200/19 cents)"},
		{"hertz(00.50)", "hertz(0.5)"},
		{"hertz(1e5)", "hertz(100000)"},
		{"hertz(1e6)", "hertz(1e+06)"},
		{"hertz(3e-324)", "hertz(5e-324)"},
		{"hertz(1.7976931348623157e308)", "hertz(1.7976931348623157e+308)"},
		{"hertz(-0e5)", "hertz(0)"},
		{"hertz(5e-324)", "hertz(5e-324)"},
		{"hertz(-0)", "hertz(0)"},
		{"hertz(0.1)", "hertz(0.1)"},
		{"relative(1e0)", "relative(1)"},
		{"unset", "unset"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			criterion, err := frequency.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, criterion.String())
			roundTrip, err := frequency.Parse(criterion.String())
			require.NoError(t, err)
			requireCriterionEqual(t, criterion, roundTrip)
		})
	}
}

func TestParse_NumericToleranceValues(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"hertz(5e-324)", math.SmallestNonzeroFloat64},
		{"hertz(-0)", 0},
		{"hertz(0.1)", 0.1},
	}
	for _, tt := range tests {
		criterion, err := frequency.Parse(tt.input)
		require.NoError(t, err)
		tolerance, ok := criterion.Tolerance()
		require.True(t, ok)
		assert.Equal(t, tt.want, tolerance)
		assert.False(t, math.Signbit(tolerance))
	}
}

func TestCriterionFormat(t *testing.T) {
	for _, tt := range canonicalTextCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.text, fmt.Sprintf("%v", tt.c))
			assert.Equal(t, tt.text, fmt.Sprintf("%s", tt.c))
			assert.Equal(t, strconv.Quote(tt.text), fmt.Sprintf("%q", tt.c))
			assert.Equal(t, tt.text, fmt.Sprintf("%#v", tt.c))
			assert.Equal(t, tt.text, fmt.Sprintf("%+v", tt.c))
			assert.Equal(t, tt.text, fmt.Sprintf("%.3v", tt.c))
			width := len(tt.text) + 2
			assert.Equal(t, "  "+tt.text, fmt.Sprintf("%*v", width, tt.c))
			assert.Equal(t, tt.text+"  ", fmt.Sprintf("%-*v", width, tt.c))
			assert.Equal(t, "%!d(frequency.Criterion="+tt.text+")", fmt.Sprintf("%d", tt.c))
			assert.Equal(t, "%!x(frequency.Criterion="+tt.text+")", fmt.Sprintf("%x", tt.c))
		})
	}
	assert.Equal(t, "  strict|", fmt.Sprintf("%8v|", frequency.Strict()))
	assert.Equal(t, "strict  |", fmt.Sprintf("%-8v|", frequency.Strict()))
	assert.Equal(t, "\"strict\"", fmt.Sprintf("%q", frequency.Strict()))
	assert.Equal(t, "%!d(frequency.Criterion=strict)", fmt.Sprintf("%d", frequency.Strict()))
}

func TestCriterionMarshalText(t *testing.T) {
	for _, tt := range canonicalTextCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			first, err := tt.c.MarshalText()
			require.NoError(t, err)
			second, err := tt.c.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.text, string(first))
			assert.Equal(t, tt.text, string(second))
			if len(first) > 0 {
				first[0] = 'X'
				assert.Equal(t, tt.text, string(second))
				assert.Equal(t, tt.text, tt.c.String())
			}
		})
	}
}

func TestCriterionAppendText(t *testing.T) {
	criterion := frequency.MustAnyOf(centsCriterion(1), frequency.MustWithinHertz(0.5))
	text := criterion.String()

	t.Run("sufficient capacity", func(t *testing.T) {
		backing := make([]byte, 3, len(text)+8)
		copy(backing, "pre")
		for i := 3; i < cap(backing); i++ {
			backing = backing[:i+1]
			backing[i] = 'X'
		}
		backing = backing[:3:cap(backing)]
		got, err := criterion.AppendText(backing)
		require.NoError(t, err)
		assert.Equal(t, "pre"+text, string(got))
		assert.Equal(t, "pre", string(backing))
		assert.True(t, &got[0] == &backing[0])
		for i := len(got); i < cap(backing); i++ {
			assert.Equal(t, byte('X'), backing[:cap(backing)][i])
		}
	})

	t.Run("insufficient capacity", func(t *testing.T) {
		backing := []byte("pre")
		b := backing[:len(backing):len(backing)]
		got, err := criterion.AppendText(b)
		require.NoError(t, err)
		assert.Equal(t, "pre"+text, string(got))
		assert.Equal(t, "pre", string(backing))
		assert.False(t, &got[0] == &b[0])
	})

	t.Run("nil", func(t *testing.T) {
		got, err := criterion.AppendText(nil)
		require.NoError(t, err)
		assert.Equal(t, text, string(got))
	})
}

func TestCriterionUnmarshalText(t *testing.T) {
	for _, tt := range canonicalTextCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := []byte(tt.text)
			originalInput := append([]byte(nil), input...)
			criterion := frequency.Strict()
			err := criterion.UnmarshalText(input)
			require.NoError(t, err)
			requireCriterionEqual(t, tt.c, criterion)
			assert.Equal(t, originalInput, input)
		})
	}

	t.Run("failure preserves receiver and input", func(t *testing.T) {
		criterion := centsCriterion(5)
		input := []byte("hertz(-1)")
		originalInput := append([]byte(nil), input...)
		err := criterion.UnmarshalText(input)
		requireErrorChain(t, err, frequency.ErrInvalidTolerance, nil)
		requireCriterionEqual(t, centsCriterion(5), criterion)
		assert.Equal(t, originalInput, input)
	})

	for _, input := range [][]byte{nil, {}} {
		criterion := frequency.Strict()
		err := criterion.UnmarshalText(input)
		requireErrorChain(t, err, frequency.ErrInvalidText, nil)
		requireCriterionEqual(t, frequency.Strict(), criterion)
	}
}

func TestCriterionJSON(t *testing.T) {
	type payload struct {
		Match frequency.Criterion `json:"match"`
	}
	versions := []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"encoding_json", jsonv1.Marshal, jsonv1.Unmarshal},
		{
			"encoding_json_v2",
			func(value any) ([]byte, error) { return jsonv2.Marshal(value) },
			func(data []byte, value any) error { return jsonv2.Unmarshal(data, value) },
		},
	}
	for _, version := range versions {
		for _, criterion := range []frequency.Criterion{centsCriterion(5), frequency.Strict(), {}} {
			t.Run(version.name+"/"+criterion.String(), func(t *testing.T) {
				t.Parallel()
				data, err := version.marshal(payload{Match: criterion})
				require.NoError(t, err)
				assert.JSONEq(t, `{"match":`+strconv.Quote(criterion.String())+`}`, string(data))
				var decoded payload
				require.NoError(t, version.unmarshal(data, &decoded))
				requireCriterionEqual(t, criterion, decoded.Match)
			})
		}
	}
}
