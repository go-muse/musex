package ticks_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// allModes lists the nine rounding modes in the order of rounding.Modes.
var allModes = []rounding.Mode{
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

// ticksSentinels lists every sentinel error of the package (E-1…E-10).
var ticksSentinels = []error{
	ticks.ErrUnset,
	ticks.ErrInvalidArgument,
	ticks.ErrZero,
	ticks.ErrNotRepresentable,
	ticks.ErrFormatLimit,
	ticks.ErrOrder,
	ticks.ErrNotConvertible,
	ticks.ErrOutOfRange,
	ticks.ErrInvalidText,
	ticks.ErrMeterNotMIDI,
}

// tieIndex returns, for every rounding mode, the expected index for a value that
// lies exactly between lower and lower+1, where lowerEven reports whether
// lower is even. The expectations follow the definitions of the modes for
// non-negative values.
func tieIndex(mode rounding.Mode, lower int64, lowerEven bool) int64 {
	switch mode {
	case rounding.Down, rounding.TowardZero, rounding.HalfDown, rounding.HalfTowardZero:
		return lower
	case rounding.Up, rounding.AwayFromZero, rounding.HalfUp, rounding.HalfAwayFromZero:
		return lower + 1
	case rounding.HalfEven:
		if lowerEven {
			return lower
		}
		return lower + 1
	}
	panic("unknown mode")
}

// irrationalIndex returns the expected index for a value strictly between
// lower and lower+1 that is not a midpoint; nearestUp reports whether the
// value is above lower+1/2.
func irrationalIndex(mode rounding.Mode, lower int64, nearestUp bool) int64 {
	switch mode {
	case rounding.Down, rounding.TowardZero:
		return lower
	case rounding.Up, rounding.AwayFromZero:
		return lower + 1
	}
	if nearestUp {
		return lower + 1
	}
	return lower
}

func dur(num, den int) duration.Duration {
	return duration.MustNew(num, den)
}

func pos(num, den int) duration.Position {
	return duration.MustNewPosition(num, den)
}

func frac(num, den int) fraction.Fraction {
	return fraction.MustNew(num, den)
}

func ratFrac(text string) fraction.Fraction {
	r, ok := new(big.Rat).SetString(text)
	if !ok {
		panic("invalid test rational: " + text)
	}
	return fraction.MustFromBigRat(r)
}

func sec(num, den int) tempo.Seconds {
	return tempo.MustNewSeconds(num, den)
}

func quarterTempo(num, den int) tempo.Tempo {
	return tempo.MustNewQuarter(fraction.MustNew(num, den))
}

func writtenQuarter() duration.Written {
	return duration.Quarter.MustWritten()
}

func writtenDottedQuarter() duration.Written {
	return duration.Quarter.MustWritten().MustAddDot()
}

func writtenHalf() duration.Written {
	return duration.Half.MustWritten()
}

func writtenEighth() duration.Written {
	return duration.Eighth.MustWritten()
}

func writtenWhole() duration.Written {
	return duration.Whole.MustWritten()
}

func at(num, den int) tempomap.Address {
	return tempomap.MustAt(pos(num, den))
}

func res(perQuarter int) ticks.Resolution {
	return ticks.MustNewResolution(perQuarter)
}

func resQ(perQuarter, thirtySeconds int) ticks.Resolution {
	return ticks.MustNewResolution(perQuarter).MustWithQuarter(ticks.MustNewMIDIQuarter(thirtySeconds))
}

func mtempo(microseconds int) ticks.MIDITempo {
	return ticks.MustNewMIDITempo(microseconds)
}

func tchange(tick, microseconds int) ticks.MIDITempoChange {
	return ticks.MIDITempoChange{Tick: tick, Tempo: mtempo(microseconds)}
}

func mmeter(nn, dd, cc, bb int) ticks.MIDIMeter {
	return ticks.MustNewMIDIMeter(nn, dd, cc, bb)
}

func mchange(tick, nn, dd, cc, bb int) ticks.MIDIMeterChange {
	return ticks.MIDIMeterChange{Tick: tick, Meter: mmeter(nn, dd, cc, bb)}
}

func smpte(code, ticksPerFrame int) ticks.SMPTE {
	return ticks.MustNewSMPTE(ticks.MustFrameRateOf(code), ticksPerFrame)
}

// ritardandoMap is the musex map of KP-15: quarter = 120 at 0 with a gradual
// change starting at 0, and quarter = 100 as a new main tempo at 1.
func ritardandoMap() tempomap.Map {
	return tempomap.MustNew(quarterTempo(120, 1)).
		MustWithChange(tempomap.MustNewGradual(at(0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarterTempo(100, 1)))
}

// constantMap is a musex position map with one constant quarter tempo.
func constantMap(quarterBPM int) tempomap.Map {
	return tempomap.MustNew(quarterTempo(quarterBPM, 1))
}

// holdMap is quarter = 60 with a fermata hold of the given seconds at position p.
func holdMap(p duration.Position, seconds tempo.Seconds) tempomap.Map {
	return constantMap(60).MustWithExtension(tempomap.MustNewExtension(tempomap.MustAt(p), seconds))
}

// unfinishedMap ends with an unfinished gradual change at position 1.
func unfinishedMap() tempomap.Map {
	return constantMap(120).MustWithChange(tempomap.MustNewGradual(at(1, 1), tempomap.LinearSpeed))
}

func grid44() measure.Grid {
	return measure.MustNew(meter.FourFour())
}

func addr(bar, num, den int) measure.Address {
	return measure.MustNewAddress(bar, pos(num, den))
}

// requireTicksError checks that err wraps own, no other sentinel of the
// package, and every listed dependency sentinel.
func requireTicksError(t *testing.T, err error, own error, dependencies ...error) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, own)
	for _, sentinel := range ticksSentinels {
		if sentinel != own {
			assert.NotErrorIs(t, err, sentinel, "unexpected additional ticks sentinel")
		}
	}
	for _, dependency := range dependencies {
		assert.ErrorIs(t, err, dependency)
	}
}

// requireMustPanic checks that fn panics with an error equal in text and
// sentinel to expected.
func requireMustPanic(t *testing.T, expected error, own error, fn func()) {
	t.Helper()
	require.Error(t, expected)
	defer func() {
		t.Helper()
		recovered := recover()
		require.NotNil(t, recovered, "Must variant did not panic")
		actual, ok := recovered.(error)
		require.True(t, ok, "panic value has type %T, want error", recovered)
		assert.EqualError(t, actual, expected.Error())
		requireTicksError(t, actual, own)
	}()
	fn()
}

// equaler is a value with an Equal method of its own type.
type equaler[T any] interface {
	Equal(T) bool
	String() string
}

func assertEqualValue[T equaler[T]](t *testing.T, want, got T) {
	t.Helper()
	assert.True(t, want.Equal(got), "got %s, want %s", got.String(), want.String())
}

func assertFraction(t *testing.T, want, got fraction.Fraction) {
	t.Helper()
	assert.True(t, want.Equal(got), "got %s, want %s", got, want)
}

func assertSeconds(t *testing.T, want, got tempo.Seconds) {
	t.Helper()
	assert.True(t, want.Equal(got), "got %s, want %s", got, want)
}

func assertQuantized(t *testing.T, got ticks.Quantized, wantTicks int, wantDeviation fraction.Fraction) {
	t.Helper()
	n, err := got.Ticks()
	require.NoError(t, err)
	assert.Equal(t, wantTicks, n)
	d, err := got.Deviation()
	require.NoError(t, err)
	assertFraction(t, wantDeviation, d)
}

func tempoChanges(t *testing.T, m ticks.MIDITempoMap) []string {
	t.Helper()
	changes, err := m.Changes()
	require.NoError(t, err)
	out := make([]string, 0, len(changes))
	for _, c := range changes {
		out = append(out, c.String())
	}
	return out
}

func meterChanges(t *testing.T, m ticks.MIDIMeterMap) []string {
	t.Helper()
	changes, err := m.Changes()
	require.NoError(t, err)
	out := make([]string, 0, len(changes))
	for _, c := range changes {
		out = append(out, c.String())
	}
	return out
}

// errNotTicks reports whether err wraps none of the package sentinels.
func errNotTicks(err error) bool {
	for _, sentinel := range ticksSentinels {
		if errors.Is(err, sentinel) {
			return false
		}
	}
	return true
}

// hugeFraction returns 2^bits as a fraction.
func hugeFraction(bits uint) fraction.Fraction {
	return fraction.MustFromBigInt(new(big.Int).Lsh(big.NewInt(1), bits), big.NewInt(1))
}

// fractionOfPowerOfTwo returns 1/2^bits as a fraction.
func fractionOfPowerOfTwo(bits uint) fraction.Fraction {
	return fraction.MustFromBigInt(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), bits))
}
