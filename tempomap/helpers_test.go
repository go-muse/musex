package tempomap_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allRoundingModes = []rounding.Mode{
	rounding.Down,
	rounding.Up,
	rounding.TowardZero,
	rounding.AwayFromZero,
	rounding.HalfUp,
	rounding.HalfDown,
	rounding.HalfAwayFromZero,
	rounding.HalfTowardZero,
	rounding.HalfEven,
}

var allSentinels = []error{
	tempomap.ErrUnset,
	tempomap.ErrInvalidArgument,
	tempomap.ErrNoInitialTempo,
	tempomap.ErrDuplicateAddress,
	tempomap.ErrInvalidAddress,
	tempomap.ErrNoDeviation,
	tempomap.ErrInvalidSequence,
	tempomap.ErrUnfinished,
	tempomap.ErrNegativeTime,
	tempomap.ErrOutOfRange,
	tempomap.ErrInvalidText,
}

func pos(num, den int) duration.Position {
	return duration.MustNewPosition(num, den)
}

func at(num, den int) tempomap.Address {
	return tempomap.MustAt(pos(num, den))
}

func bar(number, offsetNum, offsetDen int) tempomap.Address {
	return tempomap.MustAtBar(measure.MustNewAddress(number, pos(offsetNum, offsetDen)))
}

func sec(num, den int) tempo.Seconds {
	return tempo.MustNewSeconds(num, den)
}

func quarter(bpm int) tempo.Tempo {
	return tempo.MustNewQuarter(fraction.FromInt(bpm))
}

func dottedQuarter(bpm int) tempo.Tempo {
	return tempo.MustNew(duration.Quarter.MustWritten().MustAddDot(), fraction.FromInt(bpm))
}

func eighth(bpm int) tempo.Tempo {
	return tempo.MustNew(duration.Eighth.MustWritten(), fraction.FromInt(bpm))
}

func fractionRat(text string) fraction.Fraction {
	r, ok := new(big.Rat).SetString(text)
	if !ok {
		panic("invalid test rational: " + text)
	}
	return fraction.MustFromBigRat(r)
}

func secondsRat(text string) tempo.Seconds {
	return tempo.MustSecondsFromFraction(fractionRat(text))
}

func grid34Pickup() measure.Grid {
	return measure.MustNew(meter.ThreeFour()).MustWithPickup(duration.MustNew(1, 4))
}

func grid44Finite(bars int) measure.Grid {
	return measure.MustNew(meter.FourFour()).MustWithBarCount(bars)
}

func newJumpMap() tempomap.Map {
	return tempomap.MustNew(quarter(60)).MustWithChange(
		tempomap.MustNewJump(at(2, 1), quarter(120)),
	)
}

func newGradualMap(endPosition duration.Position, final tempo.Tempo) tempomap.Map {
	return tempomap.MustNew(quarter(60)).
		MustWithChange(tempomap.MustNewGradual(tempomap.MustAt(duration.Origin()), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(tempomap.MustAt(endPosition), final))
}

func newHoldMap() tempomap.Map {
	return tempomap.MustNew(quarter(60)).MustWithExtension(
		tempomap.MustNewExtension(at(1, 2), sec(1, 1)),
	)
}

func requireTempomapError(t *testing.T, err error, own error, dependencies ...error) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, own)
	for _, sentinel := range allSentinels {
		if sentinel != own {
			assert.NotErrorIs(t, err, sentinel, "unexpected additional tempomap sentinel")
		}
	}
	for _, dependency := range dependencies {
		assert.ErrorIs(t, err, dependency)
	}
}

func requirePanicError(t *testing.T, expected error, own error, fn func()) {
	t.Helper()
	require.NotNil(t, fn)
	defer func() {
		recovered := recover()
		require.NotNil(t, recovered, "Must variant did not panic")
		actual, ok := recovered.(error)
		require.True(t, ok, "panic value has type %T, want error", recovered)
		assert.EqualError(t, actual, expected.Error())
		requireTempomapError(t, actual, own)
	}()
	fn()
}

func checkSecondsResult(t *testing.T, got tempomap.SecondsResult, value string, approximate bool, places int, bound string) {
	t.Helper()
	require.True(t, got.IsSet())
	actualValue, err := got.Value()
	require.NoError(t, err)
	assert.Equal(t, tempo.MustParseSeconds(value), actualValue)
	actualApproximate, err := got.Approximate()
	require.NoError(t, err)
	assert.Equal(t, approximate, actualApproximate)
	actualPlaces, hasPlaces, err := got.Places()
	require.NoError(t, err)
	assert.Equal(t, approximate, hasPlaces)
	if approximate {
		assert.Equal(t, places, actualPlaces)
	} else {
		assert.Zero(t, actualPlaces)
	}
	actualBound, err := got.ErrorBound()
	require.NoError(t, err)
	assert.Equal(t, tempo.MustParseSeconds(bound), actualBound)
}

func checkPositionResult(t *testing.T, got tempomap.PositionResult, value string, approximate bool, places int, bound string) {
	t.Helper()
	require.True(t, got.IsSet())
	actualValue, err := got.Value()
	require.NoError(t, err)
	assert.Equal(t, duration.MustParsePosition(value), actualValue)
	actualApproximate, err := got.Approximate()
	require.NoError(t, err)
	assert.Equal(t, approximate, actualApproximate)
	actualPlaces, hasPlaces, err := got.Places()
	require.NoError(t, err)
	assert.Equal(t, approximate, hasPlaces)
	if approximate {
		assert.Equal(t, places, actualPlaces)
	} else {
		assert.Zero(t, actualPlaces)
	}
	actualBound, err := got.ErrorBound()
	require.NoError(t, err)
	assert.Equal(t, duration.MustParsePosition(bound), actualBound)
}

func checkState(t *testing.T, got tempomap.State, current, main tempo.Tempo, deviating bool) {
	t.Helper()
	actualCurrent, err := got.Current()
	require.NoError(t, err)
	actualMain, err := got.Main()
	require.NoError(t, err)
	actualDeviating, err := got.Deviating()
	require.NoError(t, err)
	assert.True(t, current.Equal(actualCurrent), "current tempo: got %s, want %s", actualCurrent, current)
	assert.True(t, main.Equal(actualMain), "main tempo: got %s, want %s", actualMain, main)
	assert.Equal(t, deviating, actualDeviating)
}
