package ticks_test

import (
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"
	"github.com/go-muse/musex/tempo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMIDITempoValue checks О-12, П-4, ТК-8, КП-19, КП-28.
func TestMIDITempoValue(t *testing.T) {
	t.Parallel()
	for _, us := range []int{1, 500000, 600000, 16777215} {
		m, err := ticks.NewMIDITempo(us)
		require.NoError(t, err)
		assert.True(t, m.IsSet())
		got, err := m.Microseconds()
		require.NoError(t, err)
		assert.Equal(t, us, got)
		assert.Equal(t, us, m.MustMicroseconds())
		assert.True(t, m.Equal(ticks.MustNewMIDITempo(us)))
	}
	assert.Equal(t, 500000, ticks.DefaultMIDITempo().MustMicroseconds())
	for _, tt := range []struct {
		name string
		us   int
		own  error
	}{
		{name: "KP-19 zero", us: 0, own: ticks.ErrZero},
		{name: "negative", us: -1, own: ticks.ErrInvalidArgument},
		{name: "KP-28 above Set Tempo", us: 16777216, own: ticks.ErrFormatLimit},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := ticks.NewMIDITempo(tt.us)
			requireTicksError(t, err, tt.own)
			assert.Equal(t, ticks.MIDITempo{}, m)
			requireMustPanic(t, err, tt.own, func() { ticks.MustNewMIDITempo(tt.us) })
		})
	}
	assert.Equal(t, -1, mtempo(500000).Compare(mtempo(600000)))
	assert.Equal(t, 1, mtempo(600000).Compare(mtempo(500000)))
	assert.Equal(t, 0, mtempo(600000).Compare(mtempo(600000)))
	assert.Equal(t, -1, ticks.MIDITempo{}.Compare(mtempo(1)))
	assert.Equal(t, 1, mtempo(1).Compare(ticks.MIDITempo{}))
	assert.Equal(t, 0, ticks.MIDITempo{}.Compare(ticks.MIDITempo{}))
}

// TestMicrosecondsPerQuarter checks ИТ-6, О-12, КП-8, КП-21, КП-28, КП-45.
func TestMicrosecondsPerQuarter(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		t     tempo.Tempo
		q     []ticks.MIDIQuarter
		exact fraction.Fraction
		near  int
		down  int
	}{
		{name: "KP-8 quarter 120", t: tempo.MustParse("4=120"), exact: frac(500000, 1), near: 500000, down: 500000},
		{name: "KP-8 half 60", t: tempo.MustParse("2=60"), exact: frac(500000, 1), near: 500000, down: 500000},
		{name: "KP-8 quarter 100", t: tempo.MustParse("4=100"), exact: frac(600000, 1), near: 600000, down: 600000},
		{name: "KP-8 dotted quarter 60", t: tempo.MustParse("4.=60"), exact: frac(2000000, 3), near: 666667, down: 666666},
		{name: "KP-21 quarter 4", t: tempo.MustParse("4=4"), exact: frac(15000000, 1), near: 15000000, down: 15000000},
		{name: "KP-21 quarter 3.6", t: quarterTempo(18, 5), exact: frac(50000000, 3), near: 16666667, down: 16666666},
		{name: "KP-28 largest", t: quarterTempo(60000000, 16777215), exact: frac(16777215, 1), near: 16777215, down: 16777215},
		{name: "KP-28 smallest", t: quarterTempo(60000000, 1), exact: frac(1, 1), near: 1, down: 1},
		{name: "KP-45 written half quarter", t: tempo.MustParse("2=120"), q: []ticks.MIDIQuarter{ticks.MustNewMIDIQuarter(16)}, exact: frac(500000, 1), near: 500000, down: 500000},
		{name: "KP-45 quarter 240 with written half quarter", t: tempo.MustParse("4=240"), q: []ticks.MIDIQuarter{ticks.MustNewMIDIQuarter(16)}, exact: frac(500000, 1), near: 500000, down: 500000},
		{name: "explicit standard quarter", t: tempo.MustParse("4=120"), q: []ticks.MIDIQuarter{ticks.StandardQuarter()}, exact: frac(500000, 1), near: 500000, down: 500000},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			exact, err := ticks.MicrosecondsPerQuarter(tt.t, tt.q...)
			require.NoError(t, err)
			assertFraction(t, tt.exact, exact)
			assertFraction(t, tt.exact, ticks.MustMicrosecondsPerQuarter(tt.t, tt.q...))
			near, err := ticks.MIDITempoOf(tt.t, rounding.HalfEven, tt.q...)
			require.NoError(t, err)
			assert.Equal(t, tt.near, near.MustMicroseconds())
			assert.Equal(t, tt.down, ticks.MustMIDITempoOf(tt.t, rounding.Down, tt.q...).MustMicroseconds())
		})
	}
}

// TestMIDITempoOfTies checks all nine modes on exact ties (ИТ-11).
func TestMIDITempoOfTies(t *testing.T) {
	t.Parallel()
	// 15·10⁶/N = 500000.5 for N = 30000000/1000001; 500001.5 for N = 30000000/1000003.
	even := quarterTempo(30000000, 1000001)
	odd := quarterTempo(30000000, 1000003)
	for _, mode := range allModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, int(tieIndex(mode, 500000, true)), ticks.MustMIDITempoOf(even, mode).MustMicroseconds())
			assert.Equal(t, int(tieIndex(mode, 500001, false)), ticks.MustMIDITempoOf(odd, mode).MustMicroseconds())
		})
	}
}

func TestMIDITempoOfErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		t    tempo.Tempo
		mode rounding.Mode
		q    []ticks.MIDIQuarter
		own  error
		deps []error
	}{
		{name: "KP-21 quarter 3.5", t: quarterTempo(7, 2), mode: rounding.HalfEven, own: ticks.ErrFormatLimit},
		{name: "KP-21 whole 0.5", t: tempo.MustNew(writtenWhole(), frac(1, 2)), mode: rounding.HalfEven, own: ticks.ErrFormatLimit},
		{name: "KP-28 just above the limit after rounding up", t: quarterTempo(120000000, 33554431), mode: rounding.Up, own: ticks.ErrFormatLimit},
		{name: "zero after rounding", t: quarterTempo(120000000, 1), mode: rounding.Down, own: ticks.ErrFormatLimit},
		{name: "unset tempo", t: tempo.Tempo{}, mode: rounding.HalfEven, own: ticks.ErrInvalidArgument, deps: []error{tempo.ErrUnset}},
		{name: "unset mode", t: quarterTempo(120, 1), mode: rounding.Mode(0), own: ticks.ErrInvalidArgument, deps: []error{rounding.ErrUnset}},
		{name: "unknown mode", t: quarterTempo(120, 1), mode: rounding.Mode(99), own: ticks.ErrInvalidArgument, deps: []error{rounding.ErrUnknownMode}},
		{name: "two written quarters", t: quarterTempo(120, 1), mode: rounding.HalfEven, q: []ticks.MIDIQuarter{ticks.StandardQuarter(), ticks.StandardQuarter()}, own: ticks.ErrInvalidArgument},
		{name: "unset written quarter", t: quarterTempo(120, 1), mode: rounding.HalfEven, q: []ticks.MIDIQuarter{{}}, own: ticks.ErrUnset},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := ticks.MIDITempoOf(tt.t, tt.mode, tt.q...)
			requireTicksError(t, err, tt.own, tt.deps...)
			assert.Equal(t, ticks.MIDITempo{}, m)
			requireMustPanic(t, err, tt.own, func() { ticks.MustMIDITempoOf(tt.t, tt.mode, tt.q...) })
		})
	}
	// Exact values are not limited.
	exact, err := ticks.MicrosecondsPerQuarter(quarterTempo(7, 2))
	require.NoError(t, err)
	assertFraction(t, frac(120000000, 7), exact)
	assert.Equal(t, 16777215, ticks.MustMIDITempoOf(quarterTempo(120000000, 33554431), rounding.Down).MustMicroseconds())
	_, err = ticks.MicrosecondsPerQuarter(tempo.Tempo{})
	requireTicksError(t, err, ticks.ErrInvalidArgument, tempo.ErrUnset)
}

// TestMIDITempoToTempo checks О-13, ИТ-6, КП-8, КП-40, КП-45.
func TestMIDITempoToTempo(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		us   int
		unit duration.Written
		q    []ticks.MIDIQuarter
		want tempo.Tempo
	}{
		{name: "KP-8 600000", us: 600000, unit: writtenQuarter(), want: tempo.MustParse("4=100")},
		{name: "KP-8 700000", us: 700000, unit: writtenQuarter(), want: quarterTempo(600, 7)},
		{name: "KP-8 666667 dotted quarter", us: 666667, unit: writtenDottedQuarter(), want: tempo.MustNew(writtenDottedQuarter(), frac(40000000, 666667))},
		{name: "KP-40 dotted quarter 40", us: 1000000, unit: writtenDottedQuarter(), want: tempo.MustParse("4.=40")},
		{name: "KP-40 quarter 60", us: 1000000, unit: writtenQuarter(), want: tempo.MustParse("4=60")},
		{name: "eighth", us: 500000, unit: writtenEighth(), want: tempo.MustParse("8=240")},
		{name: "KP-45 half 120", us: 500000, unit: writtenHalf(), q: []ticks.MIDIQuarter{ticks.MustNewMIDIQuarter(16)}, want: tempo.MustParse("2=120")},
		{name: "KP-45 quarter 240", us: 500000, unit: writtenQuarter(), q: []ticks.MIDIQuarter{ticks.MustNewMIDIQuarter(16)}, want: tempo.MustParse("4=240")},
		{name: "slowest", us: 16777215, unit: writtenQuarter(), want: quarterTempo(60000000, 16777215)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := mtempo(tt.us).Tempo(tt.unit, tt.q...)
			require.NoError(t, err)
			assertEqualValue(t, tt.want, got)
			assertEqualValue(t, tt.want, mtempo(tt.us).MustTempo(tt.unit, tt.q...))
			if tt.unit.Equal(writtenQuarter()) {
				qt, err := mtempo(tt.us).QuarterTempo(tt.q...)
				require.NoError(t, err)
				assertEqualValue(t, tt.want, qt)
				assertEqualValue(t, tt.want, mtempo(tt.us).MustQuarterTempo(tt.q...))
			}
		})
	}
	_, err := mtempo(500000).Tempo(duration.Written{})
	requireTicksError(t, err, ticks.ErrInvalidArgument, duration.ErrUnset)
	_, err = mtempo(500000).QuarterTempo(ticks.MIDIQuarter{})
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = ticks.MIDITempo{}.Tempo(writtenQuarter())
	requireTicksError(t, err, ticks.ErrUnset)
	requireMustPanic(t, err, ticks.ErrUnset, func() { ticks.MIDITempo{}.MustTempo(writtenQuarter()) })
}

// TestMIDITempoSeconds checks ИТ-8 for a constant tempo, КП-10.
func TestMIDITempoSeconds(t *testing.T) {
	t.Parallel()
	s, err := mtempo(500000).Seconds(res(96), 48)
	require.NoError(t, err)
	assertSeconds(t, sec(1, 4), s)
	assertSeconds(t, sec(1, 4), mtempo(500000).MustSeconds(res(96), 48))
	assertSeconds(t, tempo.ZeroSeconds(), mtempo(500000).MustSeconds(res(96), 0))
	assertSeconds(t, sec(1, 3), mtempo(1000000).MustSeconds(res(3), 1))
	// The written quarter does not change ticks per MIDI quarter.
	assertSeconds(t, sec(1, 4), mtempo(500000).MustSeconds(resQ(96, 16), 48))
	_, err = mtempo(500000).Seconds(res(96), -1)
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	_, err = mtempo(500000).Seconds(ticks.Resolution{}, 1)
	requireTicksError(t, err, ticks.ErrUnset)
}

// TestQuartersPerMinute checks О-14, ИТ-7, П-3, КП-9, КП-19.
func TestQuartersPerMinute(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		t    tempo.Tempo
		want fraction.Fraction
	}{
		{name: "KP-9 dotted quarter 60", t: tempo.MustParse("4.=60"), want: frac(90, 1)},
		{name: "KP-9 eighth 180", t: tempo.MustParse("8=180"), want: frac(90, 1)},
		{name: "KP-9 half 50", t: tempo.MustParse("2=50"), want: frac(100, 1)},
		{name: "KP-9 quarter 600/7", t: quarterTempo(600, 7), want: frac(600, 7)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ticks.QuartersPerMinute(tt.t)
			require.NoError(t, err)
			assertFraction(t, tt.want, got)
			assertFraction(t, tt.want, ticks.MustQuartersPerMinute(tt.t))
		})
	}
	text, err := ticks.FormatQuartersPerMinute(quarterTempo(600, 7), 2, rounding.HalfUp)
	require.NoError(t, err)
	assert.Equal(t, "85.71", text)
	assert.Equal(t, "85.71", ticks.MustFormatQuartersPerMinute(quarterTempo(600, 7), 2, rounding.HalfUp))
	assert.Equal(t, "85.72", ticks.MustFormatQuartersPerMinute(quarterTempo(600, 7), 2, rounding.Up))
	assert.Equal(t, "90.000", ticks.MustFormatQuartersPerMinute(tempo.MustParse("4.=60"), 3, rounding.Down))
	for _, places := range []int{0, 1, 5, rounding.MaxPlaces} {
		want, err := rounding.HalfEven.RatDecimal(big.NewRat(600, 7), places)
		require.NoError(t, err)
		assert.Equal(t, want, ticks.MustFormatQuartersPerMinute(quarterTempo(600, 7), places, rounding.HalfEven), "places %d", places)
	}
	for name, call := range map[string]func() error{
		"negative places": func() error {
			_, err := ticks.FormatQuartersPerMinute(quarterTempo(60, 1), -1, rounding.HalfUp)
			return err
		},
		"too many places": func() error {
			_, err := ticks.FormatQuartersPerMinute(quarterTempo(60, 1), rounding.MaxPlaces+1, rounding.HalfUp)
			return err
		},
		"unset mode": func() error {
			_, err := ticks.FormatQuartersPerMinute(quarterTempo(60, 1), 2, rounding.Mode(0))
			return err
		},
		"unset tempo":     func() error { _, err := ticks.FormatQuartersPerMinute(tempo.Tempo{}, 2, rounding.HalfUp); return err },
		"unset tempo qpm": func() error { _, err := ticks.QuartersPerMinute(tempo.Tempo{}); return err },
		"unset unit":      func() error { _, err := ticks.TempoFromQuartersPerMinute(frac(90, 1), duration.Written{}); return err },
		"negative qpm":    func() error { _, err := ticks.TempoFromQuartersPerMinute(frac(-90, 1), writtenQuarter()); return err },
		"unset qpm": func() error {
			_, err := ticks.TempoFromQuartersPerMinute(fraction.Fraction{}, writtenQuarter())
			return err
		},
		"qpm checked first": func() error { _, err := ticks.TempoFromQuartersPerMinute(frac(-1, 1), duration.Written{}); return err },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, call(), ticks.ErrInvalidArgument)
		})
	}
	_, err = ticks.TempoFromQuartersPerMinute(frac(-1, 1), duration.Written{})
	assert.NotErrorIs(t, err, duration.ErrUnset)
	_, err = ticks.TempoFromQuartersPerMinute(fraction.Fraction{}, writtenQuarter())
	requireTicksError(t, err, ticks.ErrInvalidArgument, fraction.ErrInvalid)

	got, err := ticks.TempoFromQuartersPerMinute(frac(90, 1), writtenDottedQuarter())
	require.NoError(t, err)
	assertEqualValue(t, tempo.MustParse("4.=60"), got)
	assertEqualValue(t, tempo.MustParse("4=90"), ticks.MustTempoFromQuartersPerMinute(frac(90, 1), writtenQuarter()))
	assertEqualValue(t, quarterTempo(8571, 100), ticks.MustTempoFromQuartersPerMinute(frac(8571, 100), writtenQuarter()))
	_, err = ticks.TempoFromQuartersPerMinute(fraction.Zero(), writtenQuarter())
	requireTicksError(t, err, ticks.ErrZero)
	requireMustPanic(t, err, ticks.ErrZero, func() { ticks.MustTempoFromQuartersPerMinute(fraction.Zero(), writtenQuarter()) })
}

// TestParseQuartersPerMinute checks the qpm grammar of 4.15 and КП-19.
func TestParseQuartersPerMinute(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text string
		want fraction.Fraction
	}{
		{text: "90", want: frac(90, 1)},
		{text: "85.71", want: frac(8571, 100)},
		{text: "600/7", want: frac(600, 7)},
		{text: "1200/14", want: frac(600, 7)},
		{text: "0.5", want: frac(1, 2)},
		{text: "120.000", want: frac(120, 1)},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			got, err := ticks.ParseQuartersPerMinute(tt.text)
			require.NoError(t, err)
			assertFraction(t, tt.want, got)
			assertFraction(t, tt.want, ticks.MustParseQuartersPerMinute(tt.text))
		})
	}
	for _, tt := range []struct {
		text string
		own  error
	}{
		{text: "0", own: ticks.ErrZero},
		{text: "0.0", own: ticks.ErrZero},
		{text: "0/5", own: ticks.ErrZero},
		{text: "", own: ticks.ErrInvalidText},
		{text: "-90", own: ticks.ErrInvalidText},
		{text: "090", own: ticks.ErrInvalidText},
		{text: "90.", own: ticks.ErrInvalidText},
		{text: ".5", own: ticks.ErrInvalidText},
		{text: "1/0", own: ticks.ErrInvalidText},
		{text: " 90", own: ticks.ErrInvalidText},
		{text: "ninety", own: ticks.ErrInvalidText},
		{text: "4=90", own: ticks.ErrInvalidText},
	} {
		t.Run("invalid "+tt.text, func(t *testing.T) {
			t.Parallel()
			got, err := ticks.ParseQuartersPerMinute(tt.text)
			requireTicksError(t, err, tt.own)
			assert.False(t, got.IsValid())
			requireMustPanic(t, err, tt.own, func() { ticks.MustParseQuartersPerMinute(tt.text) })
		})
	}
}
