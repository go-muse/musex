package frequency_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/frequency"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireMustPanic(t *testing.T, expectedFrequency, expectedCents error, call func()) {
	t.Helper()
	defer func() {
		recovered := recover()
		require.NotNil(t, recovered, "Must variant did not panic")
		err, ok := recovered.(error)
		require.True(t, ok, "panic value has type %T, want error", recovered)
		requireErrorChain(t, err, expectedFrequency, expectedCents)
	}()
	call()
}

func TestMustVariants_Success(t *testing.T) {
	assert.True(t, frequency.MustEqual(440, 440))
	assert.Equal(t, -1, frequency.MustCompare(440, 441))

	withinCents := frequency.MustWithinCents(cents.MustFromInt(5))
	expectedCents, err := frequency.WithinCents(cents.MustFromInt(5))
	require.NoError(t, err)
	requireCriterionEqual(t, expectedCents, withinCents)

	withinHertz := frequency.MustWithinHertz(0.5)
	expectedHertz, err := frequency.WithinHertz(0.5)
	require.NoError(t, err)
	requireCriterionEqual(t, expectedHertz, withinHertz)

	withinRelative := frequency.MustWithinRelative(0.01)
	expectedRelative, err := frequency.WithinRelative(0.01)
	require.NoError(t, err)
	requireCriterionEqual(t, expectedRelative, withinRelative)

	all := frequency.MustAllOf(withinCents, withinHertz)
	expectedAll, err := frequency.AllOf(withinCents, withinHertz)
	require.NoError(t, err)
	requireCriterionEqual(t, expectedAll, all)

	any := frequency.MustAnyOf(withinCents, withinHertz)
	expectedAny, err := frequency.AnyOf(withinCents, withinHertz)
	require.NoError(t, err)
	requireCriterionEqual(t, expectedAny, any)

	parsed := frequency.MustParse("any(cents(5), hertz(0.5))")
	requireCriterionEqual(t, any, parsed)
	assert.False(t, frequency.MustParse("unset").IsValid())

	assert.True(t, frequency.Strict().MustNear(440, 440))
	assert.Equal(t, 1, frequency.Strict().MustCompare(441, 440))
}

func TestMustVariants_PanicWithOriginalError(t *testing.T) {
	tests := []struct {
		name          string
		call          func()
		wantFrequency error
		wantCents     error
	}{
		{"MustEqual", func() { frequency.MustEqual(440, 0) }, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument},
		{"MustCompare", func() { frequency.MustCompare(0, 440) }, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument},
		{"MustWithinCents", func() { frequency.MustWithinCents(cents.MustFromInt(-1)) }, frequency.ErrInvalidTolerance, nil},
		{"MustWithinHertz", func() { frequency.MustWithinHertz(math.NaN()) }, frequency.ErrInvalidTolerance, nil},
		{"MustWithinRelative", func() { frequency.MustWithinRelative(-1) }, frequency.ErrInvalidTolerance, nil},
		{"MustAllOf", func() { frequency.MustAllOf() }, frequency.ErrNoComponents, nil},
		{"MustAnyOf", func() { frequency.MustAnyOf(frequency.Criterion{}) }, frequency.ErrUnset, nil},
		{"MustParse", func() { frequency.MustParse("abc") }, frequency.ErrInvalidText, nil},
		{"Criterion.MustNear", func() { frequency.Strict().MustNear(440, 0) }, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument},
		{"Criterion.MustCompare", func() { frequency.Strict().MustCompare(0, 440) }, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requireMustPanic(t, tt.wantFrequency, tt.wantCents, tt.call)
		})
	}
}

func TestNumericMustVariants_SuccessAndOwnership(t *testing.T) {
	transposed, err := cents.MustStep(7, 12).ApplyTo(440)
	require.NoError(t, err)
	assert.Equal(t, transposed, frequency.MustTranspose(440, 7, 12))
	assert.Equal(t, "3/2", frequency.MustRatio(440, 660).RatString())
	assert.True(t, frequency.MustCentsBetween(440, 442).Equal(cents.MustParse("7.851415")))
	assert.Equal(t, 220.0, frequency.MustPartial(110, 2))
	assert.Equal(t, 220.0, frequency.MustSubharmonic(440, 2))
	assert.Equal(t, 2.0, frequency.MustBeatFrequency(440, 442))
	assert.Equal(t, 880.0, frequency.MustShiftOctaves(440, 1))
	representative, k := frequency.MustNormalizeOctave(880, 110)
	assert.Equal(t, 110.0, representative)
	assert.Equal(t, -3, k)
	period, ok := modelPeriod(440)
	require.True(t, ok)
	assert.Equal(t, period, frequency.MustPeriod(440))

	partials := frequency.MustPartials(110, []int{2, 3, 4})
	require.Len(t, partials, 3)
	assert.Equal(t, 3, cap(partials))
	partials[0] = 0
	freshPartials := frequency.MustPartials(110, []int{2, 3, 4})
	assert.Equal(t, 220.0, freshPartials[0])
	assert.NotSame(t, &partials[0], &freshPartials[0])

	subharmonics := frequency.MustSubharmonics(440, []int{1, 2, 3, 4})
	require.Len(t, subharmonics, 4)
	assert.Equal(t, 4, cap(subharmonics))
	subharmonics[0] = 0
	freshSubharmonics := frequency.MustSubharmonics(440, []int{1, 2, 3, 4})
	assert.Equal(t, 440.0, freshSubharmonics[0])
	assert.NotSame(t, &subharmonics[0], &freshSubharmonics[0])
}

func TestNumericMustVariants_PanicWithOriginalError(t *testing.T) {
	tests := []struct {
		name          string
		call          func()
		wantFrequency error
		wantCents     error
	}{
		{"MustTranspose", func() { frequency.MustTranspose(440, 1, 0) }, frequency.ErrInvalidParameter, cents.ErrInvalidArgument},
		{"MustRatio", func() { frequency.MustRatio(0, 440) }, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument},
		{"MustCentsBetween", func() { frequency.MustCentsBetween(440, 0) }, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument},
		{"MustPartial", func() { frequency.MustPartial(math.MaxFloat64, 2) }, frequency.ErrOutOfRange, cents.ErrOutOfRange},
		{"MustPartials", func() { frequency.MustPartials(110, nil) }, frequency.ErrInvalidParameter, nil},
		{"MustSubharmonic", func() { frequency.MustSubharmonic(cents.MinNormalFrequency, 2) }, frequency.ErrOutOfRange, cents.ErrOutOfRange},
		{"MustSubharmonics", func() { frequency.MustSubharmonics(440, []int{0}) }, frequency.ErrInvalidParameter, nil},
		{"MustBeatFrequency", func() { frequency.MustBeatFrequency(0, 440) }, frequency.ErrInvalidFrequency, cents.ErrInvalidArgument},
		{"MustShiftOctaves", func() { frequency.MustShiftOctaves(math.MaxFloat64, 1) }, frequency.ErrOutOfRange, cents.ErrOutOfRange},
		{"MustNormalizeOctave", func() { frequency.MustNormalizeOctave(math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64) }, frequency.ErrOutOfRange, cents.ErrOutOfRange},
		{"MustPeriod", func() { frequency.MustPeriod(math.SmallestNonzeroFloat64) }, frequency.ErrOutOfRange, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requireMustPanic(t, tt.wantFrequency, tt.wantCents, tt.call)
		})
	}
}
