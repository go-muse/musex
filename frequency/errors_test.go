package frequency_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseErrorCases() []struct {
	name          string
	input         string
	wantFrequency error
	wantCents     error
} {
	cases := []struct {
		name          string
		input         string
		wantFrequency error
		wantCents     error
	}{
		{"K-33 unknown", "abc", frequency.ErrInvalidText, nil},
		{"K-33 trailing", "hertz(1)x", frequency.ErrInvalidText, nil},
		{"K-33 negative hertz", "hertz(-1)", frequency.ErrInvalidTolerance, nil},
		{"K-33 empty all", "all()", frequency.ErrNoComponents, nil},
		{"K-33 negative nested cents", "any(cents(1), cents(-1))", frequency.ErrInvalidTolerance, nil},
		{"K-34 prime limit", "cents(37:32)", frequency.ErrInvalidTolerance, cents.ErrPrimeLimit},
		{"K-34 out of range", "cents(9000001)", frequency.ErrInvalidTolerance, cents.ErrOutOfRange},
		{"K-34 too complex", "cents(1" + strings.Repeat("0", 100) + ")", frequency.ErrInvalidTolerance, cents.ErrTooComplex},
		{"K-35 overflow", "hertz(1e400)", frequency.ErrInvalidTolerance, nil},
		{"K-35 underflow", "hertz(1e-400)", frequency.ErrInvalidTolerance, nil},
		{"K-36 tolerance before unset", "any(hertz(-1), unset)", frequency.ErrInvalidTolerance, nil},
		{"K-36 unset before prime limit", "all(unset, cents(37:32))", frequency.ErrUnset, nil},
		{"K-36 grammar before tolerance", "all(hertz(-1), strict)x", frequency.ErrInvalidText, nil},
		{"E-3 all unset", "all(unset)", frequency.ErrUnset, nil},
		{"E-3 any trailing unset", "any(strict, unset)", frequency.ErrUnset, nil},
		{"E-3 nested unset", "any(all(unset))", frequency.ErrUnset, nil},
		{"E-4 any empty", "any()", frequency.ErrNoComponents, nil},
		{"E-4 nested empty", "any(strict, all())", frequency.ErrNoComponents, nil},
		{"order empty then tolerance", "all(all(), hertz(-1))", frequency.ErrNoComponents, nil},
		{"order tolerance then empty", "all(hertz(-1), all())", frequency.ErrInvalidTolerance, nil},
		{"order unset then empty", "any(unset, all())", frequency.ErrUnset, nil},
		{"order empty then unset", "any(all(), unset)", frequency.ErrNoComponents, nil},
		{"E-2 negative cents", "cents(-1)", frequency.ErrInvalidTolerance, nil},
		{"E-2 negative ratio", "cents(2:3)", frequency.ErrInvalidTolerance, nil},
		{"E-2 negative fraction", "cents(-1/2 cents)", frequency.ErrInvalidTolerance, nil},
		{"E-2 typographic minus", "cents(−5)", frequency.ErrInvalidTolerance, nil},
		{"E-2 negative exponent hertz", "hertz(-0.5e-3)", frequency.ErrInvalidTolerance, nil},
		{"E-2 negative relative", "relative(-1)", frequency.ErrInvalidTolerance, nil},
		{"E-2 hertz overflow 309", "hertz(1e309)", frequency.ErrInvalidTolerance, nil},
		{"E-2 relative overflow", "relative(1e400)", frequency.ErrInvalidTolerance, nil},
		{"E-2 hertz underflow", "hertz(2e-324)", frequency.ErrInvalidTolerance, nil},
		{"E-2 negative underflow", "hertz(-1e-400)", frequency.ErrInvalidTolerance, nil},
		{"E-2 cents negative out of range", "cents(-9000001)", frequency.ErrInvalidTolerance, cents.ErrOutOfRange},
		{"E-2 cents fractional out of range", "cents(9000000.000001)", frequency.ErrInvalidTolerance, cents.ErrOutOfRange},
		{"E-5 cents empty", "cents()", frequency.ErrInvalidText, cents.ErrInvalidText},
		{"E-5 cents unknown", "cents(abc)", frequency.ErrInvalidText, cents.ErrInvalidText},
		{"E-5 cents bare fraction", "cents(3/2)", frequency.ErrInvalidText, cents.ErrInvalidText},
		{"E-5 cents leading space", "cents( 5)", frequency.ErrInvalidText, cents.ErrInvalidText},
		{"E-5 cents trailing space", "cents(5 )", frequency.ErrInvalidText, cents.ErrInvalidText},
		{"E-5 cents nested number", "cents((5))", frequency.ErrInvalidText, cents.ErrInvalidText},
		{"E-5 cents grammar before tolerance", "all(hertz(-1), cents(abc))", frequency.ErrInvalidText, cents.ErrInvalidText},
	}
	invalidText := []string{
		"", " strict", "strict ", "Strict", "STRICT", "strict()", "unset()",
		"cents", "cents(5", "all(strict", "strict)", "all(strict,strict)",
		"all(strict,  strict)", "all(strict ,strict)", "all(strict, )", "all(, strict)",
		"all((strict))", "hertz()", "hertz(+1)", "hertz(.5)", "hertz(5.)",
		"hertz(1E5)", "hertz(1e)", "hertz(1e+)", "hertz(inf)", "hertz(Inf)",
		"hertz(NaN)", "hertz(0x1p-2)", "hertz(1_000)", "hertz( 1)", "hertz(1 )",
		"relative(--1)", "hertz(１)", "strict\x00",
	}
	for _, input := range invalidText {
		cases = append(cases, struct {
			name          string
			input         string
			wantFrequency error
			wantCents     error
		}{"E-5 " + input, input, frequency.ErrInvalidText, nil})
	}
	return cases
}

func TestParse_ErrorsAndChains(t *testing.T) {
	for _, tt := range parseErrorCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			criterion, err := frequency.Parse(tt.input)
			assert.False(t, criterion.IsValid())
			requireErrorChain(t, err, tt.wantFrequency, tt.wantCents)

			input := []byte(tt.input)
			originalInput := slices.Clone(input)
			receiver := centsCriterion(5)
			err = receiver.UnmarshalText(input)
			requireErrorChain(t, err, tt.wantFrequency, tt.wantCents)
			requireCriterionEqual(t, centsCriterion(5), receiver)
			assert.Equal(t, originalInput, input)
		})
	}
}

func TestFrequencySentinels_AreDistinct(t *testing.T) {
	sentinels := []error{
		frequency.ErrInvalidFrequency,
		frequency.ErrInvalidTolerance,
		frequency.ErrUnset,
		frequency.ErrNoComponents,
		frequency.ErrInvalidText,
		frequency.ErrInvalidParameter,
		frequency.ErrOutOfRange,
	}
	for _, outer := range sentinels {
		for _, inner := range sentinels {
			assert.Equal(t, outer == inner, errors.Is(outer, inner))
		}
	}
}

func TestParse_ValidCentsStep(t *testing.T) {
	criterion, err := frequency.Parse("cents(1200/19 cents)")
	require.NoError(t, err)
	requireCriterionEqual(t, frequency.MustWithinCents(cents.MustStep(1, 19)), criterion)
	assert.Equal(t, "cents(1200/19 cents)", criterion.String())
}
