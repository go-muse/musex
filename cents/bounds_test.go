package cents_test

import (
	"math"
	"strconv"
	"sync"
	"testing"

	"github.com/go-muse/musex/cents"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	boundMinNormalFrequency = 0x1p-1022
	boundMaxFrequency       = math.MaxFloat64
)

type boundFrequencyCase struct {
	name   string
	value  float64
	valid  bool
	normal bool
}

var (
	boundValidPredicateSink  bool
	boundNormalPredicateSink bool
	boundValidFrequencySink  float64
	boundNormalFrequencySink float64
	boundValidFrequencyErr   error
	boundNormalFrequencyErr  error
)

func boundFrequencyCases() []boundFrequencyCase {
	return []boundFrequencyCase{
		{name: "440 Hz", value: 440, valid: true, normal: true},
		{name: "one", value: 1, valid: true, normal: true},
		{name: "small normal", value: 1e-300, valid: true, normal: true},
		{name: "large normal", value: 1e300, valid: true, normal: true},
		{name: "minimum normal", value: boundMinNormalFrequency, valid: true, normal: true},
		{name: "above minimum normal", value: math.Nextafter(boundMinNormalFrequency, 1), valid: true, normal: true},
		{name: "maximum finite", value: boundMaxFrequency, valid: true, normal: true},
		{name: "largest subnormal", value: math.Nextafter(boundMinNormalFrequency, 0), valid: true, normal: false},
		{name: "smallest subnormal", value: math.SmallestNonzeroFloat64, valid: true, normal: false},
		{name: "positive zero", value: 0, valid: false, normal: false},
		{name: "negative zero", value: math.Copysign(0, -1), valid: false, normal: false},
		{name: "negative 440 Hz", value: -440, valid: false, normal: false},
		{name: "negative subnormal", value: -math.SmallestNonzeroFloat64, valid: false, normal: false},
		{name: "negative minimum normal", value: -boundMinNormalFrequency, valid: false, normal: false},
		{name: "negative maximum finite", value: -boundMaxFrequency, valid: false, normal: false},
		{name: "NaN", value: math.NaN(), valid: false, normal: false},
		{name: "positive infinity", value: math.Inf(1), valid: false, normal: false},
		{name: "negative infinity", value: math.Inf(-1), valid: false, normal: false},
	}
}

func boundAssertOnlyFrequencyError(t *testing.T, err, want error) {
	t.Helper()
	require.Error(t, err)
	is := assert.New(t)
	is.ErrorIs(err, want)
	for _, sentinel := range []error{
		cents.ErrOutOfRange,
		cents.ErrTooComplex,
		cents.ErrPrimeLimit,
		cents.ErrInvalidArgument,
		cents.ErrInvalidText,
	} {
		if sentinel != want {
			is.NotErrorIs(err, sentinel)
		}
	}
}

func boundObserveMustFrequency(call func() float64) (value float64, panicValue any) {
	defer func() {
		panicValue = recover()
	}()
	value = call()
	return value, nil
}

func boundAssertFrequencyResult(
	t *testing.T,
	tc boundFrequencyCase,
	normalRule bool,
	got float64,
	err error,
	mustGot float64,
	panicValue any,
) {
	t.Helper()
	is := assert.New(t)
	wantOK := tc.valid
	wantError := cents.ErrInvalidArgument
	if normalRule {
		wantOK = tc.normal
		wantError = cents.ErrOutOfRange
	}

	if wantOK {
		is.NoError(err)
		is.Equal(math.Float64bits(tc.value), math.Float64bits(got))
		is.Nil(panicValue)
		is.Equal(math.Float64bits(tc.value), math.Float64bits(mustGot))
		return
	}

	is.Zero(got)
	boundAssertOnlyFrequencyError(t, err, wantError)
	require.NotNil(t, panicValue)
	panicErr, ok := panicValue.(error)
	require.True(t, ok, "panic value has type %T, want error", panicValue)
	boundAssertOnlyFrequencyError(t, panicErr, wantError)
}

func boundAssertZeroApplyTo(t *testing.T, tc boundFrequencyCase) {
	t.Helper()
	got, err := cents.Zero().ApplyTo(tc.value)
	is := assert.New(t)
	switch {
	case tc.normal:
		is.NoError(err)
		is.Equal(math.Float64bits(tc.value), math.Float64bits(got))
	case !tc.valid:
		is.Zero(got)
		boundAssertOnlyFrequencyError(t, err, cents.ErrInvalidArgument)
	default:
		is.Zero(got)
		boundAssertOnlyFrequencyError(t, err, cents.ErrOutOfRange)
	}
}

func TestFrequencyBoundConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		read     func() float64
		wantBits uint64
		wantText string
	}{
		{
			name:     "minimum normal frequency",
			read:     func() float64 { return cents.MinNormalFrequency },
			wantBits: 0x0010000000000000,
			wantText: "2.2250738585072014e-308",
		},
		{
			name:     "maximum frequency",
			read:     func() float64 { return cents.MaxFrequency },
			wantBits: 0x7fefffffffffffff,
			wantText: "1.7976931348623157e+308",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			first := tt.read()
			second := tt.read()
			is.Equal(tt.wantBits, math.Float64bits(first))
			is.Equal(tt.wantBits, math.Float64bits(second))
			is.Equal(tt.wantText, strconv.FormatFloat(first, 'g', -1, 64))
		})
	}
}

func TestFrequencyBoundPredicates(t *testing.T) {
	t.Parallel()

	for _, tt := range boundFrequencyCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			is.Equal(tt.valid, cents.IsValidFrequency(tt.value))
			is.Equal(tt.normal, cents.IsNormalFrequency(tt.value))
		})
	}
}

func TestFrequencyBoundValidFrequency(t *testing.T) {
	t.Parallel()

	for _, tt := range boundFrequencyCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := cents.ValidFrequency(tt.value)
			mustGot, panicValue := boundObserveMustFrequency(func() float64 {
				return cents.MustValidFrequency(tt.value)
			})
			boundAssertFrequencyResult(t, tt, false, got, err, mustGot, panicValue)
		})
	}
}

func TestFrequencyBoundNormalFrequency(t *testing.T) {
	t.Parallel()

	for _, tt := range boundFrequencyCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := cents.NormalFrequency(tt.value)
			mustGot, panicValue := boundObserveMustFrequency(func() float64 {
				return cents.MustNormalFrequency(tt.value)
			})
			boundAssertFrequencyResult(t, tt, true, got, err, mustGot, panicValue)
		})
	}
}

func TestFrequencyBoundApplyTo(t *testing.T) {
	t.Parallel()

	for _, tt := range boundFrequencyCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			boundAssertZeroApplyTo(t, tt)
		})
	}
}

func TestFrequencyBoundBetween(t *testing.T) {
	t.Parallel()

	for _, tt := range boundFrequencyCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, call := range []struct {
				name string
				f1   float64
				f2   float64
			}{
				{name: "value first", f1: tt.value, f2: 440},
				{name: "value second", f1: 440, f2: tt.value},
			} {
				t.Run(call.name, func(t *testing.T) {
					t.Parallel()
					got, err := cents.Between(call.f1, call.f2)
					if tt.valid {
						assert.NoError(t, err)
						return
					}
					assert.Equal(t, cents.Zero(), got)
					boundAssertOnlyFrequencyError(t, err, cents.ErrInvalidArgument)
				})
			}
		})
	}
}

func TestFrequencyBoundOperationEdges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		call         func() (float64, error)
		want         float64
		wantErr      error
		compareValue bool
		relative     bool
	}{
		{name: "multiplier zero", call: func() (float64, error) { return cents.Zero().Multiplier() }, want: 1, compareValue: true},
		{name: "multiplier one octave", call: func() (float64, error) { return cents.MustFromInt(1200).Multiplier() }, want: 2, compareValue: true},
		{name: "multiplier minus one octave", call: func() (float64, error) { return cents.MustFromInt(-1200).Multiplier() }, want: 0.5, compareValue: true},
		{name: "multiplier two octaves", call: func() (float64, error) { return cents.MustFromInt(2400).Multiplier() }, want: 4, compareValue: true},
		{name: "multiplier 700 cents", call: func() (float64, error) { return cents.MustFromInt(700).Multiplier() }, want: 1.4983070768766815, compareValue: true, relative: true},
		{name: "multiplier three to two", call: func() (float64, error) { return cents.MustFromRatio(3, 2).Multiplier() }, want: 1.5, compareValue: true, relative: true},
		{name: "multiplier minimum normal", call: func() (float64, error) { return cents.MustFromInt(-1226400).Multiplier() }, want: boundMinNormalFrequency, compareValue: true},
		{name: "multiplier below minimum normal", call: func() (float64, error) { return cents.MustFromInt(-1226401).Multiplier() }, wantErr: cents.ErrOutOfRange},
		{name: "multiplier below overflow", call: func() (float64, error) { return cents.MustFromInt(1228799).Multiplier() }},
		{name: "multiplier overflow", call: func() (float64, error) { return cents.MustFromInt(1228800).Multiplier() }, wantErr: cents.ErrOutOfRange},
		{name: "multiplier far above range", call: func() (float64, error) { return cents.MustFromInt(1300000).Multiplier() }, wantErr: cents.ErrOutOfRange},
		{name: "multiplier far below range", call: func() (float64, error) { return cents.MustFromInt(-1300000).Multiplier() }, wantErr: cents.ErrOutOfRange},
		{name: "apply zero", call: func() (float64, error) { return cents.Zero().ApplyTo(440) }, want: 440, compareValue: true},
		{name: "apply one octave", call: func() (float64, error) { return cents.MustFromInt(1200).ApplyTo(440) }, want: 880, compareValue: true},
		{name: "apply ratio 442 to 440", call: func() (float64, error) { return cents.MustFromRatio(442, 440).ApplyTo(440) }, want: 442, compareValue: true, relative: true},
		{name: "apply overflow", call: func() (float64, error) { return cents.MustFromInt(120000).ApplyTo(1e300) }, wantErr: cents.ErrOutOfRange},
		{name: "apply underflow", call: func() (float64, error) { return cents.MustFromInt(-120000).ApplyTo(1e-300) }, wantErr: cents.ErrOutOfRange},
		{name: "apply below minimum normal", call: func() (float64, error) { return cents.MustFromInt(-1200).ApplyTo(boundMinNormalFrequency) }, wantErr: cents.ErrOutOfRange},
		{name: "apply at maximum rounding band", call: func() (float64, error) { return cents.MustFromFraction(1, 100000000000000).ApplyTo(boundMaxFrequency) }, want: boundMaxFrequency, compareValue: true},
		{name: "apply beyond maximum rounding band", call: func() (float64, error) { return cents.MustFromFraction(1, 10000000000000).ApplyTo(boundMaxFrequency) }, wantErr: cents.ErrOutOfRange},
		{name: "apply at minimum rounding band", call: func() (float64, error) {
			return cents.MustFromFraction(-1, 100000000000000).ApplyTo(boundMinNormalFrequency)
		}, want: boundMinNormalFrequency, compareValue: true},
		{name: "apply beyond minimum rounding band", call: func() (float64, error) {
			return cents.MustFromFraction(-1, 10000000000000).ApplyTo(boundMinNormalFrequency)
		}, wantErr: cents.ErrOutOfRange},
		{name: "apply subnormal to minimum normal", call: func() (float64, error) { return cents.MustFromInt(62400).ApplyTo(math.SmallestNonzeroFloat64) }, want: boundMinNormalFrequency, compareValue: true},
		{name: "apply zero to subnormal", call: func() (float64, error) { return cents.Zero().ApplyTo(math.SmallestNonzeroFloat64) }, wantErr: cents.ErrOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			got, err := tt.call()
			if tt.wantErr != nil {
				is.Zero(got)
				boundAssertOnlyFrequencyError(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			is.True(cents.IsNormalFrequency(got))
			if !tt.compareValue {
				return
			}
			if tt.relative {
				is.InDelta(tt.want, got, math.Abs(tt.want)*1e-12)
				return
			}
			is.Equal(math.Float64bits(tt.want), math.Float64bits(got))
		})
	}
}

func TestFrequencyBoundBetweenValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		f1         float64
		f2         float64
		want       string
		checkValue bool
	}{
		{name: "440 to 442", f1: 440, f2: 442, want: "7.851415", checkValue: true},
		{name: "442 to 440", f1: 442, f2: 440, want: "-7.851415", checkValue: true},
		{name: "one octave", f1: 440, f2: 880, want: "1200.000000", checkValue: true},
		{name: "unison", f1: 440, f2: 440, want: "0.000000", checkValue: true},
		{name: "wide normal range", f1: 1e-300, f2: 1e300, want: "2391788.228319", checkValue: true},
		{name: "subnormal to maximum", f1: math.SmallestNonzeroFloat64, f2: boundMaxFrequency},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := cents.Between(tt.f1, tt.f2)
			require.NoError(t, err)
			if tt.checkValue {
				assert.Equal(t, tt.want, got.Decimal(6, 0))
			}
		})
	}
}

func FuzzFrequencyBoundRules(f *testing.F) {
	for _, tt := range boundFrequencyCases() {
		f.Add(math.Float64bits(tt.value))
	}

	f.Fuzz(func(t *testing.T, bits uint64) {
		value := math.Float64frombits(bits)
		sign := bits>>63 == 0
		exponent := (bits >> 52) & 0x7ff
		mantissa := bits & 0xfffffffffffff
		normal := sign && exponent >= 1 && exponent <= 2046
		valid := normal || (sign && exponent == 0 && mantissa != 0)
		tc := boundFrequencyCase{name: "fuzz", value: value, valid: valid, normal: normal}
		is := assert.New(t)

		is.Equal(valid, cents.IsValidFrequency(value))
		is.Equal(normal, cents.IsNormalFrequency(value))

		validGot, validErr := cents.ValidFrequency(value)
		if valid {
			is.NoError(validErr)
			is.Equal(bits, math.Float64bits(validGot))
		} else {
			is.Zero(validGot)
			boundAssertOnlyFrequencyError(t, validErr, cents.ErrInvalidArgument)
		}

		normalGot, normalErr := cents.NormalFrequency(value)
		if normal {
			is.NoError(normalErr)
			is.Equal(bits, math.Float64bits(normalGot))
		} else {
			is.Zero(normalGot)
			boundAssertOnlyFrequencyError(t, normalErr, cents.ErrOutOfRange)
		}

		boundAssertZeroApplyTo(t, tc)
	})
}

type boundConcurrentResult struct {
	frequency       boundFrequencyCase
	minBits         uint64
	maxBits         uint64
	isValid         bool
	isNormal        bool
	validValue      float64
	validErr        error
	mustValidValue  float64
	mustValidPanic  any
	normalValue     float64
	normalErr       error
	mustNormalValue float64
	mustNormalPanic any
}

func TestFrequencyBoundConcurrency(t *testing.T) {
	t.Parallel()

	const workers = 8
	cases := boundFrequencyCases()
	results := make([][]boundConcurrentResult, workers)
	var wg sync.WaitGroup
	for worker := range workers {
		wg.Go(func() {
			workerResults := make([]boundConcurrentResult, 0, len(cases))
			for _, tc := range cases {
				validValue, validErr := cents.ValidFrequency(tc.value)
				mustValidValue, mustValidPanic := boundObserveMustFrequency(func() float64 {
					return cents.MustValidFrequency(tc.value)
				})
				normalValue, normalErr := cents.NormalFrequency(tc.value)
				mustNormalValue, mustNormalPanic := boundObserveMustFrequency(func() float64 {
					return cents.MustNormalFrequency(tc.value)
				})
				workerResults = append(workerResults, boundConcurrentResult{
					frequency:       tc,
					minBits:         math.Float64bits(cents.MinNormalFrequency),
					maxBits:         math.Float64bits(cents.MaxFrequency),
					isValid:         cents.IsValidFrequency(tc.value),
					isNormal:        cents.IsNormalFrequency(tc.value),
					validValue:      validValue,
					validErr:        validErr,
					mustValidValue:  mustValidValue,
					mustValidPanic:  mustValidPanic,
					normalValue:     normalValue,
					normalErr:       normalErr,
					mustNormalValue: mustNormalValue,
					mustNormalPanic: mustNormalPanic,
				})
			}
			results[worker] = workerResults
		})
	}
	wg.Wait()

	for worker, workerResults := range results {
		for _, result := range workerResults {
			t.Run(result.frequency.name, func(t *testing.T) {
				is := assert.New(t)
				is.Equal(uint64(0x0010000000000000), result.minBits, "worker %d", worker)
				is.Equal(uint64(0x7fefffffffffffff), result.maxBits, "worker %d", worker)
				is.Equal(result.frequency.valid, result.isValid, "worker %d", worker)
				is.Equal(result.frequency.normal, result.isNormal, "worker %d", worker)
				boundAssertFrequencyResult(t, result.frequency, false, result.validValue, result.validErr, result.mustValidValue, result.mustValidPanic)
				boundAssertFrequencyResult(t, result.frequency, true, result.normalValue, result.normalErr, result.mustNormalValue, result.mustNormalPanic)
			})
		}
	}
}

func TestFrequencyBoundAllocations(t *testing.T) {
	// AllocsPerRun is process-wide, so this test and its subtests are sequential.
	for _, tt := range boundFrequencyCases() {
		t.Run(tt.name+"/predicates", func(t *testing.T) {
			allocs := testing.AllocsPerRun(100, func() {
				boundValidPredicateSink = cents.IsValidFrequency(tt.value)
				boundNormalPredicateSink = cents.IsNormalFrequency(tt.value)
			})
			assert.Zero(t, allocs)
		})

		if tt.valid {
			t.Run(tt.name+"/valid", func(t *testing.T) {
				allocs := testing.AllocsPerRun(100, func() {
					boundValidFrequencySink, boundValidFrequencyErr = cents.ValidFrequency(tt.value)
				})
				assert.Zero(t, allocs)
				assert.NoError(t, boundValidFrequencyErr)
			})
		}

		if tt.normal {
			t.Run(tt.name+"/normal", func(t *testing.T) {
				allocs := testing.AllocsPerRun(100, func() {
					boundNormalFrequencySink, boundNormalFrequencyErr = cents.NormalFrequency(tt.value)
				})
				assert.Zero(t, allocs)
				assert.NoError(t, boundNormalFrequencyErr)
			})
		}
	}
}
