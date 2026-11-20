package ticks_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/format/ticks"
	"github.com/go-muse/musex/tempo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFrameRates checks О-2, КП-20 and 4.12.
func TestFrameRates(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		code int
		rate ticks.FrameRate
		text string
	}{
		{code: -24, rate: ticks.FrameRate24, text: "-24"},
		{code: -25, rate: ticks.FrameRate25, text: "-25"},
		{code: -29, rate: ticks.FrameRate29, text: "-29"},
		{code: -30, rate: ticks.FrameRate30, text: "-30"},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			got, err := ticks.FrameRateOf(tt.code)
			require.NoError(t, err)
			assert.Equal(t, tt.rate, got)
			assert.Equal(t, tt.rate, ticks.MustFrameRateOf(tt.code))
			assert.True(t, got.IsValid())
			assert.True(t, got.Equal(tt.rate))
			code, err := got.Code()
			require.NoError(t, err)
			assert.Equal(t, tt.code, code)
			assert.Equal(t, tt.code, got.MustCode())
			assert.Equal(t, tt.text, got.String())
		})
	}
	for _, code := range []int{-28, 0, 24, 25, 30, -31, -23} {
		got, err := ticks.FrameRateOf(code)
		requireTicksError(t, err, ticks.ErrInvalidArgument)
		assert.Zero(t, got)
		requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { ticks.MustFrameRateOf(code) })
	}
	for _, invalid := range []ticks.FrameRate{0, 5, 255} {
		assert.False(t, invalid.IsValid())
		assert.Equal(t, "FrameRate(invalid)", invalid.String())
		_, err := invalid.Code()
		requireTicksError(t, err, ticks.ErrInvalidArgument)
	}
	for _, tt := range []struct {
		r    ticks.Reading
		text string
	}{
		{r: ticks.ReadingDropFrame, text: "df"},
		{r: ticks.Reading30, text: "30"},
		{r: ticks.Reading29, text: "29"},
	} {
		assert.True(t, tt.r.IsValid())
		assert.True(t, tt.r.Equal(tt.r))
		assert.Equal(t, tt.text, tt.r.String())
	}
	assert.False(t, ticks.Reading(0).IsValid())
	assert.Equal(t, "Reading(invalid)", ticks.Reading(4).String())
	assert.False(t, ticks.ReadingDropFrame.Equal(ticks.Reading30))
}

// TestSMPTE checks О-2, О-20, П-5, ИТ-9, ТК-12, КП-13, КП-44.
func TestSMPTE(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name            string
		s               ticks.SMPTE
		framesPerSecond fraction.Fraction
		ticksPerSecond  fraction.Fraction
		ticks           int
		seconds         tempo.Seconds
	}{
		{name: "KP-13 -25x40", s: smpte(-25, 40), framesPerSecond: frac(25, 1), ticksPerSecond: frac(1000, 1), ticks: 1000, seconds: sec(1, 1)},
		{name: "KP-13 -30x80", s: smpte(-30, 80), framesPerSecond: frac(30, 1), ticksPerSecond: frac(2400, 1), ticks: 2400, seconds: sec(1, 1)},
		{name: "-24x4", s: smpte(-24, 4), framesPerSecond: frac(24, 1), ticksPerSecond: frac(96, 1), ticks: 48, seconds: sec(1, 2)},
		{name: "KP-13 -29x80 read as 30", s: smpte(-29, 80).MustWithReading(ticks.Reading30), framesPerSecond: frac(30, 1), ticksPerSecond: frac(2400, 1), ticks: 2400, seconds: sec(1, 1)},
		{name: "KP-13 -29x80 read as 29", s: smpte(-29, 80).MustWithReading(ticks.Reading29), framesPerSecond: frac(29, 1), ticksPerSecond: frac(2320, 1), ticks: 2400, seconds: sec(30, 29)},
		{name: "KP-13 -29x80 read as drop-frame", s: smpte(-29, 80).MustWithReading(ticks.ReadingDropFrame), framesPerSecond: frac(30000, 1001), ticksPerSecond: frac(2400000, 1001), ticks: 2400, seconds: sec(1001, 1000)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fps, err := tt.s.FramesPerSecond()
			require.NoError(t, err)
			assertFraction(t, tt.framesPerSecond, fps)
			assertFraction(t, tt.framesPerSecond, tt.s.MustFramesPerSecond())
			tps, err := tt.s.TicksPerSecond()
			require.NoError(t, err)
			assertFraction(t, tt.ticksPerSecond, tps)
			assertFraction(t, tt.ticksPerSecond, tt.s.MustTicksPerSecond())
			s, err := tt.s.Seconds(tt.ticks)
			require.NoError(t, err)
			assertSeconds(t, tt.seconds, s)
			assertSeconds(t, tt.seconds, tt.s.MustSeconds(tt.ticks))
			in, err := tt.s.TicksIn(tt.seconds)
			require.NoError(t, err)
			assertFraction(t, fraction.FromInt(tt.ticks), in)
			assertFraction(t, fraction.FromInt(tt.ticks), tt.s.MustTicksIn(tt.seconds))
			for _, mode := range allModes {
				n, err := tt.s.TickAt(tt.seconds, mode)
				require.NoError(t, err)
				assert.Equal(t, tt.ticks, n)
			}
			zero, err := tt.s.Seconds(0)
			require.NoError(t, err)
			assertSeconds(t, tempo.ZeroSeconds(), zero)
		})
	}
}

func TestSMPTEFields(t *testing.T) {
	t.Parallel()
	s, err := ticks.NewSMPTE(ticks.FrameRate25, 40)
	require.NoError(t, err)
	assert.True(t, s.IsSet())
	rate, err := s.Rate()
	require.NoError(t, err)
	assert.Equal(t, ticks.FrameRate25, rate)
	assert.Equal(t, ticks.FrameRate25, s.MustRate())
	tpf, err := s.TicksPerFrame()
	require.NoError(t, err)
	assert.Equal(t, 40, tpf)
	assert.Equal(t, 40, s.MustTicksPerFrame())
	r, ok, err := s.Reading()
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Zero(t, r)

	df := ticks.MustNewSMPTE(ticks.FrameRate29, 80).MustWithReading(ticks.ReadingDropFrame)
	r, ok = df.MustReading()
	assert.True(t, ok)
	assert.Equal(t, ticks.ReadingDropFrame, r)
	assert.False(t, df.Equal(smpte(-29, 80)))
	assert.False(t, df == smpte(-29, 80))
	assert.True(t, df.Equal(smpte(-29, 80).MustWithReading(ticks.ReadingDropFrame)))
	assert.False(t, df.Equal(smpte(-29, 80).MustWithReading(ticks.Reading30)))
	assert.True(t, df.MustWithReading(ticks.Reading29).Equal(smpte(-29, 80).MustWithReading(ticks.Reading29)))
	assert.False(t, smpte(-25, 40).Equal(smpte(-24, 40)))
	assert.False(t, smpte(-25, 40).Equal(smpte(-25, 41)))
	assert.Equal(t, uint16(0xE350), df.MustWord())
	assert.Equal(t, 255, smpte(-30, 255).MustTicksPerFrame())
}

func TestSMPTEErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		rate ticks.FrameRate
		tpf  int
		own  error
	}{
		{name: "invalid rate", rate: ticks.FrameRate(0), tpf: 40, own: ticks.ErrInvalidArgument},
		{name: "rate outside the list", rate: ticks.FrameRate(9), tpf: 40, own: ticks.ErrInvalidArgument},
		{name: "KP-19 zero ticks per frame", rate: ticks.FrameRate25, tpf: 0, own: ticks.ErrZero},
		{name: "negative ticks per frame", rate: ticks.FrameRate25, tpf: -1, own: ticks.ErrInvalidArgument},
		{name: "ticks per frame above a byte", rate: ticks.FrameRate25, tpf: 256, own: ticks.ErrFormatLimit},
		{name: "rate is checked first", rate: ticks.FrameRate(0), tpf: 0, own: ticks.ErrInvalidArgument},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, err := ticks.NewSMPTE(tt.rate, tt.tpf)
			requireTicksError(t, err, tt.own)
			assert.Equal(t, ticks.SMPTE{}, s)
			requireMustPanic(t, err, tt.own, func() { ticks.MustNewSMPTE(tt.rate, tt.tpf) })
		})
	}
	noReading := smpte(-29, 80)
	for name, call := range map[string]func() error{
		"KP-13 seconds of -29 need a reading": func() error { _, err := noReading.Seconds(2400); return err },
		"frames per second of -29":            func() error { _, err := noReading.FramesPerSecond(); return err },
		"ticks per second of -29":             func() error { _, err := noReading.TicksPerSecond(); return err },
		"ticks in seconds of -29":             func() error { _, err := noReading.TicksIn(sec(1, 1)); return err },
		"tick at seconds of -29":              func() error { _, err := noReading.TickAt(sec(1, 1), rounding.Down); return err },
		"reading for -25":                     func() error { _, err := smpte(-25, 40).WithReading(ticks.Reading30); return err },
		"invalid reading":                     func() error { _, err := noReading.WithReading(ticks.Reading(0)); return err },
		"negative ticks":                      func() error { _, err := smpte(-25, 40).Seconds(-1); return err },
		"unset seconds":                       func() error { _, err := smpte(-25, 40).TicksIn(tempo.Seconds{}); return err },
		"unset mode":                          func() error { _, err := smpte(-25, 40).TickAt(sec(1, 1), rounding.Mode(0)); return err },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, call(), ticks.ErrInvalidArgument)
		})
	}
	_, err := smpte(-25, 40).TicksIn(tempo.Seconds{})
	requireTicksError(t, err, ticks.ErrInvalidArgument, tempo.ErrUnset)
	_, err = smpte(-25, 40).TickAt(sec(1, 1), rounding.Mode(0))
	requireTicksError(t, err, ticks.ErrInvalidArgument, rounding.ErrUnset)
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { smpte(-25, 40).MustTickAt(sec(1, 1), rounding.Mode(0)) })
	_, err = noReading.Seconds(2400)
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { noReading.MustSeconds(2400) })

	var unset ticks.SMPTE
	assert.False(t, unset.IsSet())
	assert.Equal(t, "SMPTE(unset)", unset.String())
	assert.True(t, unset.Equal(ticks.SMPTE{}))
	_, err = unset.Rate()
	requireTicksError(t, err, ticks.ErrUnset)
	_, _, err = unset.Reading()
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = unset.WithReading(ticks.Reading30)
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = unset.Word()
	requireTicksError(t, err, ticks.ErrUnset)
}

// TestSMPTETickAtRounding checks КП-44 and all nine modes on a tie (ИТ-11).
func TestSMPTETickAtRounding(t *testing.T) {
	t.Parallel()
	s := smpte(-30, 80)
	in := s.MustTicksIn(sec(4, 10000))
	assertFraction(t, frac(24, 25), in)
	n, err := s.TickAt(sec(4, 10000), rounding.HalfUp)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, 0, s.MustTickAt(sec(4, 10000), rounding.Down))

	// 1/2000 s at 1000 ticks per second is half a tick; 3/2000 s is 1.5 ticks.
	thousand := smpte(-25, 40)
	for _, mode := range allModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, int(tieIndex(mode, 0, true)), thousand.MustTickAt(sec(1, 2000), mode))
			assert.Equal(t, int(tieIndex(mode, 1, false)), thousand.MustTickAt(sec(3, 2000), mode))
		})
	}
}
