package ticks_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTicksPerClockAndBeat checks О-11, ИТ-10, ТК-7, КП-7.
func TestTicksPerClockAndBeat(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		r        ticks.Resolution
		perClock fraction.Fraction
		perBeat  fraction.Fraction
	}{
		{name: "KP-7 96", r: res(96), perClock: frac(4, 1), perBeat: frac(24, 1)},
		{name: "KP-7 480", r: res(480), perClock: frac(20, 1), perBeat: frac(120, 1)},
		{name: "KP-7 100 is not integer", r: res(100), perClock: frac(25, 6), perBeat: frac(25, 1)},
		{name: "one", r: res(1), perClock: frac(1, 24), perBeat: frac(1, 4)},
		{name: "written half quarter does not change ticks per quarter", r: resQ(480, 16), perClock: frac(20, 1), perBeat: frac(120, 1)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, err := tt.r.TicksPerClock()
			require.NoError(t, err)
			assertFraction(t, tt.perClock, c)
			assertFraction(t, tt.perClock, tt.r.MustTicksPerClock())
			b, err := tt.r.TicksPerBeat()
			require.NoError(t, err)
			assertFraction(t, tt.perBeat, b)
			assertFraction(t, tt.perBeat, tt.r.MustTicksPerBeat())
		})
	}
	assert.False(t, res(100).MustTicksPerClock().IsInteger())
	assert.True(t, res(96).MustTicksPerClock().IsInteger())
}

// TestClocksAndBeats checks О-11, ИТ-10, КП-7, КП-43, КП-45.
func TestClocksAndBeats(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		r      ticks.Resolution
		p      duration.Position
		clocks fraction.Fraction
		beats  fraction.Fraction
	}{
		{name: "KP-7 eighth is 12 clocks", r: res(480), p: pos(1, 8), clocks: frac(12, 1), beats: frac(2, 1)},
		{name: "KP-43 dotted quarter is 36 clocks", r: res(96), p: pos(3, 8), clocks: frac(36, 1), beats: frac(6, 1)},
		{name: "origin", r: res(96), p: duration.Origin(), clocks: fraction.Zero(), beats: fraction.Zero()},
		{name: "triplet thirty-second", r: res(96), p: pos(1, 48), clocks: frac(2, 1), beats: frac(1, 3)},
		{name: "not integer clocks", r: res(96), p: pos(1, 192), clocks: frac(1, 2), beats: frac(1, 12)},
		{name: "written half quarter", r: resQ(480, 16), p: pos(1, 2), clocks: frac(24, 1), beats: frac(4, 1)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, err := tt.r.Clocks(tt.p)
			require.NoError(t, err)
			assertFraction(t, tt.clocks, c)
			assertFraction(t, tt.clocks, tt.r.MustClocks(tt.p))
			b, err := tt.r.Beats(tt.p)
			require.NoError(t, err)
			assertFraction(t, tt.beats, b)
			assertFraction(t, tt.beats, tt.r.MustBeats(tt.p))
			if n, ok := tt.clocks.Int(); ok {
				back, err := tt.r.ClockPosition(n)
				require.NoError(t, err)
				assertEqualValue(t, tt.p, back)
				assertEqualValue(t, tt.p, tt.r.MustClockPosition(n))
			}
			if n, ok := tt.beats.Int(); ok {
				back, err := tt.r.BeatPosition(n)
				require.NoError(t, err)
				assertEqualValue(t, tt.p, back)
				assertEqualValue(t, tt.p, tt.r.MustBeatPosition(n))
			}
		})
	}
	assertEqualValue(t, pos(3, 8), res(96).MustClockPosition(36))
	assertEqualValue(t, pos(1, 4), res(96).MustBeatPosition(4))
	assertEqualValue(t, pos(1, 2), resQ(96, 16).MustBeatPosition(4))

	for name, call := range map[string]func() error{
		"negative clocks":    func() error { _, err := res(96).ClockPosition(-1); return err },
		"negative beats":     func() error { _, err := res(96).BeatPosition(-1); return err },
		"unset position":     func() error { _, err := res(96).Clocks(duration.Position{}); return err },
		"unset beatposition": func() error { _, err := res(96).Beats(duration.Position{}); return err },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, call(), ticks.ErrInvalidArgument)
		})
	}
	_, err := res(96).Clocks(duration.Position{})
	requireTicksError(t, err, ticks.ErrInvalidArgument, duration.ErrUnset)
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { res(96).MustClocks(duration.Position{}) })
	_, err = ticks.Resolution{}.TicksPerBeat()
	requireTicksError(t, err, ticks.ErrUnset)
}
