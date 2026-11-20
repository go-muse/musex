package ticks_test

import (
	"testing"

	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMIDIMeterFields checks О-22, О-24, П-17, ТК-16, КП-33, КП-45 and 4.13.
func TestMIDIMeterFields(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name        string
		nn, dd      int
		cc, bb      int
		denominator duration.Duration
		size        duration.Duration
		click       duration.Duration
		meter       meter.Meter
	}{
		{name: "KP-33 6/8", nn: 6, dd: 3, cc: 36, bb: 8, denominator: dur(1, 8), size: dur(3, 4), click: dur(3, 8), meter: meter.MustNew(6, 8)},
		{name: "KP-33 4/4", nn: 4, dd: 2, cc: 24, bb: 8, denominator: dur(1, 4), size: dur(1, 1), click: dur(1, 4), meter: meter.MustNew(4, 4)},
		{name: "5/8 click eighth", nn: 5, dd: 3, cc: 12, bb: 8, denominator: dur(1, 8), size: dur(5, 8), click: dur(1, 8), meter: meter.MustNew(5, 8)},
		{name: "KP-45 written half quarter", nn: 4, dd: 2, cc: 12, bb: 16, denominator: dur(1, 4), size: dur(1, 1), click: dur(1, 4), meter: meter.MustNew(4, 4)},
		{name: "click of bb 10", nn: 4, dd: 2, cc: 24, bb: 10, denominator: dur(1, 4), size: dur(1, 1), click: dur(5, 16), meter: meter.MustNew(4, 4)},
		{name: "dd 0", nn: 2, dd: 0, cc: 96, bb: 8, denominator: dur(1, 1), size: dur(2, 1), click: dur(1, 1), meter: meter.MustNew(2, 1)},
		{name: "largest bytes", nn: 255, dd: 5, cc: 255, bb: 255, denominator: dur(1, 32), size: dur(255, 32), click: dur(65025, 768), meter: meter.MustNew(255, 32)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := ticks.NewMIDIMeter(tt.nn, tt.dd, tt.cc, tt.bb)
			require.NoError(t, err)
			assert.True(t, m.IsSet())
			assert.True(t, m.Equal(ticks.MustNewMIDIMeter(tt.nn, tt.dd, tt.cc, tt.bb)))
			nn, err := m.Numerator()
			require.NoError(t, err)
			assert.Equal(t, tt.nn, nn)
			assert.Equal(t, tt.nn, m.MustNumerator())
			dd, err := m.DenominatorExponent()
			require.NoError(t, err)
			assert.Equal(t, tt.dd, dd)
			assert.Equal(t, tt.dd, m.MustDenominatorExponent())
			cc, err := m.ClocksPerClick()
			require.NoError(t, err)
			assert.Equal(t, tt.cc, cc)
			assert.Equal(t, tt.cc, m.MustClocksPerClick())
			bb, err := m.ThirtySeconds()
			require.NoError(t, err)
			assert.Equal(t, tt.bb, bb)
			assert.Equal(t, tt.bb, m.MustThirtySeconds())
			q, err := m.Quarter()
			require.NoError(t, err)
			assert.True(t, q.Equal(ticks.MustNewMIDIQuarter(tt.bb)))
			assert.True(t, m.MustQuarter().Equal(ticks.MustNewMIDIQuarter(tt.bb)))
			d, err := m.Denominator()
			require.NoError(t, err)
			assertEqualValue(t, tt.denominator, d)
			assertEqualValue(t, tt.denominator, m.MustDenominator())
			size, err := m.Size()
			require.NoError(t, err)
			assertEqualValue(t, tt.size, size)
			assertEqualValue(t, tt.size, m.MustSize())
			click, err := m.Click()
			require.NoError(t, err)
			assertEqualValue(t, tt.click, click)
			assertEqualValue(t, tt.click, m.MustClick())
			mt, err := m.Meter()
			require.NoError(t, err)
			assertEqualValue(t, tt.meter, mt)
			assertEqualValue(t, tt.meter, m.MustMeter())
		})
	}
	assert.True(t, ticks.DefaultMIDIMeter().Equal(mmeter(4, 2, 24, 8)))
	assert.True(t, ticks.DefaultMIDIMeter() == mmeter(4, 2, 24, 8))
	assert.False(t, mmeter(4, 2, 24, 8).Equal(mmeter(4, 2, 24, 16)))
	assert.False(t, mmeter(4, 2, 24, 8).Equal(mmeter(4, 2, 12, 8)))

	// dd up to 254 is a value; 2^dd above int has no meter.
	far := mmeter(1, 254, 24, 8)
	d, err := far.Denominator()
	require.NoError(t, err)
	assertEqualValue(t, duration.MustFromFraction(fractionOfPowerOfTwo(254)), d)
	_, err = far.Meter()
	requireTicksError(t, err, ticks.ErrOutOfRange)
	_, err = mmeter(1, 63, 24, 8).Meter()
	requireTicksError(t, err, ticks.ErrOutOfRange)
	requireMustPanic(t, err, ticks.ErrOutOfRange, func() { mmeter(1, 63, 24, 8).MustMeter() })
	assertEqualValue(t, meter.MustNew(1, 1<<62), mmeter(1, 62, 24, 8).MustMeter())
}

// TestMIDIMeterErrors checks П-3, КП-39 and the order of 4.16.
func TestMIDIMeterErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name           string
		nn, dd, cc, bb int
		own            error
	}{
		{name: "KP-39 nn zero", nn: 0, dd: 2, cc: 24, bb: 8, own: ticks.ErrZero},
		{name: "KP-39 cc zero", nn: 4, dd: 2, cc: 0, bb: 8, own: ticks.ErrZero},
		{name: "KP-39 bb zero", nn: 4, dd: 2, cc: 24, bb: 0, own: ticks.ErrZero},
		{name: "KP-39 dd 255", nn: 4, dd: 255, cc: 24, bb: 8, own: ticks.ErrOutOfRange},
		{name: "dd above a byte", nn: 4, dd: 256, cc: 24, bb: 8, own: ticks.ErrFormatLimit},
		{name: "nn above a byte", nn: 256, dd: 2, cc: 24, bb: 8, own: ticks.ErrFormatLimit},
		{name: "cc above a byte", nn: 4, dd: 2, cc: 256, bb: 8, own: ticks.ErrFormatLimit},
		{name: "bb above a byte", nn: 4, dd: 2, cc: 24, bb: 256, own: ticks.ErrFormatLimit},
		{name: "negative nn", nn: -1, dd: 2, cc: 24, bb: 8, own: ticks.ErrInvalidArgument},
		{name: "negative dd", nn: 4, dd: -1, cc: 24, bb: 8, own: ticks.ErrInvalidArgument},
		{name: "negative cc", nn: 4, dd: 2, cc: -1, bb: 8, own: ticks.ErrInvalidArgument},
		{name: "negative bb", nn: 4, dd: 2, cc: 24, bb: -1, own: ticks.ErrInvalidArgument},
		{name: "fields in order: nn first", nn: 0, dd: 255, cc: 0, bb: 0, own: ticks.ErrZero},
		{name: "fields in order: dd before cc", nn: 4, dd: 255, cc: 0, bb: 8, own: ticks.ErrOutOfRange},
		{name: "fields in order: cc before bb", nn: 4, dd: 2, cc: 256, bb: 0, own: ticks.ErrFormatLimit},
		{name: "fields in order: negative dd before zero cc", nn: 4, dd: -3, cc: 0, bb: 0, own: ticks.ErrInvalidArgument},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := ticks.NewMIDIMeter(tt.nn, tt.dd, tt.cc, tt.bb)
			requireTicksError(t, err, tt.own)
			assert.Equal(t, ticks.MIDIMeter{}, m)
			requireMustPanic(t, err, tt.own, func() { ticks.MustNewMIDIMeter(tt.nn, tt.dd, tt.cc, tt.bb) })
		})
	}
	var unset ticks.MIDIMeter
	assert.False(t, unset.IsSet())
	assert.Equal(t, "MIDIMeter(unset)", unset.String())
	_, err := unset.Numerator()
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = unset.Meter()
	requireTicksError(t, err, ticks.ErrUnset)
}

// TestClick checks П-15 and 4.13.
func TestClick(t *testing.T) {
	t.Parallel()
	every, err := ticks.ClickEvery(dur(3, 8))
	require.NoError(t, err)
	assert.True(t, every.IsSet())
	onBeat, err := every.OnBeat()
	require.NoError(t, err)
	assert.False(t, onBeat)
	d, ok, err := every.Every()
	require.NoError(t, err)
	assert.True(t, ok)
	assertEqualValue(t, dur(3, 8), d)
	d, ok = every.MustEvery()
	assert.True(t, ok)
	assertEqualValue(t, dur(3, 8), d)
	assert.True(t, every.Equal(ticks.MustClickEvery(dur(3, 8))))
	assert.True(t, every == ticks.MustClickEvery(dur(6, 16)))

	beat := ticks.ClickOnBeat()
	assert.True(t, beat.IsSet())
	assert.True(t, beat.MustOnBeat())
	d, ok, err = beat.Every()
	require.NoError(t, err)
	assert.False(t, ok)
	assert.False(t, d.IsSet())
	assert.False(t, beat.Equal(every))
	assert.True(t, beat.Equal(ticks.ClickOnBeat()))

	_, err = ticks.ClickEvery(duration.Duration{})
	requireTicksError(t, err, ticks.ErrInvalidArgument, duration.ErrUnset)
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { ticks.MustClickEvery(duration.Duration{}) })
	var unset ticks.Click
	assert.False(t, unset.IsSet())
	assert.Equal(t, "Click(unset)", unset.String())
	assert.True(t, unset.Equal(ticks.Click{}))
	assert.False(t, unset.Equal(beat))
	_, err = unset.OnBeat()
	requireTicksError(t, err, ticks.ErrUnset)
	_, _, err = unset.Every()
	requireTicksError(t, err, ticks.ErrUnset)
}

// TestMIDIMeterOf checks О-23, П-14, П-15, П-16, ТК-16, ТК-19, КП-31, КП-32, КП-45.
func TestMIDIMeterOf(t *testing.T) {
	t.Parallel()
	additive := meter.MustNewAdditive([]int{3, 2, 3}, 8)
	sum := meter.MustNewSum(meter.TwoFour(), meter.ThreeEight())
	grouped := meter.MustNew(7, 8).MustWithGrouping(2, 2, 3)
	for _, tt := range []struct {
		name   string
		meter  meter.Meter
		click  ticks.Click
		q      ticks.MIDIQuarter
		f      ticks.Fidelity
		want   ticks.MIDIMeter
		losses []ticks.Loss
	}{
		{name: "KP-31 6/8 click 3/8", meter: meter.SixEight(), click: ticks.MustClickEvery(dur(3, 8)), q: ticks.StandardQuarter(), f: ticks.Strict, want: mmeter(6, 3, 36, 8)},
		{name: "KP-31 4/4 click 1/4", meter: meter.FourFour(), click: ticks.MustClickEvery(dur(1, 4)), q: ticks.StandardQuarter(), f: ticks.Strict, want: mmeter(4, 2, 24, 8)},
		{name: "KP-31 3/4 on beat", meter: meter.ThreeFour(), click: ticks.ClickOnBeat(), q: ticks.StandardQuarter(), f: ticks.Strict, want: mmeter(3, 2, 24, 8)},
		{name: "KP-31 6/8 on beat", meter: meter.SixEight(), click: ticks.ClickOnBeat(), q: ticks.StandardQuarter(), f: ticks.Strict, want: mmeter(6, 3, 36, 8)},
		{name: "KP-31 5/8 click 1/8", meter: meter.MustNew(5, 8), click: ticks.MustClickEvery(dur(1, 8)), q: ticks.StandardQuarter(), f: ticks.Strict, want: mmeter(5, 3, 12, 8)},
		{name: "lossless meter in lossy mode", meter: meter.SixEight(), click: ticks.ClickOnBeat(), q: ticks.StandardQuarter(), f: ticks.Lossy, want: mmeter(6, 3, 36, 8)},
		{name: "KP-32 common time", meter: meter.CommonTime(), click: ticks.ClickOnBeat(), q: ticks.StandardQuarter(), f: ticks.Lossy, want: mmeter(4, 2, 24, 8), losses: []ticks.Loss{ticks.LostSymbol}},
		{name: "cut time", meter: meter.CutTime(), click: ticks.ClickOnBeat(), q: ticks.StandardQuarter(), f: ticks.Lossy, want: mmeter(2, 1, 48, 8), losses: []ticks.Loss{ticks.LostSymbol}},
		{name: "KP-32 additive numerator", meter: additive, click: ticks.MustClickEvery(dur(1, 8)), q: ticks.StandardQuarter(), f: ticks.Lossy, want: mmeter(8, 3, 12, 8), losses: []ticks.Loss{ticks.LostGrouping}},
		{name: "KP-32 sum of meters", meter: sum, click: ticks.MustClickEvery(dur(1, 8)), q: ticks.StandardQuarter(), f: ticks.Lossy, want: mmeter(7, 3, 12, 8), losses: []ticks.Loss{ticks.LostParts}},
		{name: "explicit grouping", meter: grouped, click: ticks.MustClickEvery(dur(1, 8)), q: ticks.StandardQuarter(), f: ticks.Lossy, want: mmeter(7, 3, 12, 8), losses: []ticks.Loss{ticks.LostGrouping}},
		{name: "KP-45 written half quarter", meter: meter.FourFour(), click: ticks.MustClickEvery(dur(1, 4)), q: ticks.MustNewMIDIQuarter(16), f: ticks.Strict, want: mmeter(4, 2, 12, 16)},
		{name: "on beat with written half quarter", meter: meter.ThreeFour(), click: ticks.ClickOnBeat(), q: ticks.MustNewMIDIQuarter(16), f: ticks.Strict, want: mmeter(3, 2, 12, 16)},
		{name: "smallest click", meter: meter.FourFour(), click: ticks.MustClickEvery(dur(1, 96)), q: ticks.StandardQuarter(), f: ticks.Strict, want: mmeter(4, 2, 1, 8)},
		{name: "largest click", meter: meter.FourFour(), click: ticks.MustClickEvery(dur(85, 32)), q: ticks.StandardQuarter(), f: ticks.Strict, want: mmeter(4, 2, 255, 8)},
		{name: "largest numerator", meter: meter.MustNew(255, 4), click: ticks.MustClickEvery(dur(1, 4)), q: ticks.StandardQuarter(), f: ticks.Strict, want: mmeter(255, 2, 24, 8)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, losses, err := ticks.MIDIMeterOf(tt.meter, tt.click, tt.q, tt.f)
			require.NoError(t, err)
			assertEqualValue(t, tt.want, got)
			assert.Equal(t, tt.losses, losses)
			mustGot, mustLosses := ticks.MustMIDIMeterOf(tt.meter, tt.click, tt.q, tt.f)
			assertEqualValue(t, tt.want, mustGot)
			assert.Equal(t, tt.losses, mustLosses)
		})
	}
}

// TestMIDIMeterOfErrors checks П-14, П-15, ТК-16, КП-38 and the order of 4.16.
func TestMIDIMeterOfErrors(t *testing.T) {
	t.Parallel()
	std := ticks.StandardQuarter()
	additive := meter.MustNewAdditive([]int{3, 2, 3}, 8)
	sum := meter.MustNewSum(meter.TwoFour(), meter.ThreeEight())
	grouped := meter.MustNew(7, 8).MustWithGrouping(2, 2, 3)
	for _, tt := range []struct {
		name  string
		meter meter.Meter
		click ticks.Click
		q     ticks.MIDIQuarter
		f     ticks.Fidelity
		own   error
		deps  []error
		not   []error
	}{
		{name: "KP-38 4/3 strict", meter: meter.MustNew(4, 3), click: ticks.MustClickEvery(dur(1, 3)), q: std, f: ticks.Strict, own: ticks.ErrMeterNotMIDI},
		{name: "KP-38 4/3 lossy", meter: meter.MustNew(4, 3), click: ticks.MustClickEvery(dur(1, 3)), q: std, f: ticks.Lossy, own: ticks.ErrMeterNotMIDI},
		{name: "KP-38 common time strict", meter: meter.CommonTime(), click: ticks.ClickOnBeat(), q: std, f: ticks.Strict, own: ticks.ErrMeterNotMIDI},
		{name: "KP-38 additive strict", meter: additive, click: ticks.MustClickEvery(dur(1, 8)), q: std, f: ticks.Strict, own: ticks.ErrMeterNotMIDI},
		{name: "KP-38 sum strict", meter: sum, click: ticks.MustClickEvery(dur(1, 8)), q: std, f: ticks.Strict, own: ticks.ErrMeterNotMIDI},
		{name: "explicit grouping strict", meter: grouped, click: ticks.MustClickEvery(dur(1, 8)), q: std, f: ticks.Strict, own: ticks.ErrMeterNotMIDI},
		{name: "KP-38 5/8 on beat without grouping", meter: meter.MustNew(5, 8), click: ticks.ClickOnBeat(), q: std, f: ticks.Strict, own: ticks.ErrMeterNotMIDI, deps: []error{meter.ErrGroupingRequired}},
		{name: "KP-38 7/8 as 2+2+3 on beat", meter: meter.MustNewAdditive([]int{2, 2, 3}, 8), click: ticks.ClickOnBeat(), q: std, f: ticks.Lossy, own: ticks.ErrMeterNotMIDI},
		{name: "KP-38 click 1/64 gives cc 1.5", meter: meter.FourFour(), click: ticks.MustClickEvery(dur(1, 64)), q: std, f: ticks.Strict, own: ticks.ErrMeterNotMIDI},
		{name: "click above 255 clocks", meter: meter.FourFour(), click: ticks.MustClickEvery(dur(3, 1)), q: std, f: ticks.Strict, own: ticks.ErrMeterNotMIDI},
		{name: "on beat with bb 10 is not integer", meter: meter.FourFour(), click: ticks.ClickOnBeat(), q: ticks.MustNewMIDIQuarter(10), f: ticks.Strict, own: ticks.ErrMeterNotMIDI},
		{name: "numerator above 255", meter: meter.MustNew(256, 4), click: ticks.MustClickEvery(dur(1, 4)), q: std, f: ticks.Strict, own: ticks.ErrMeterNotMIDI},
		{name: "unset meter", meter: meter.Meter{}, click: ticks.ClickOnBeat(), q: std, f: ticks.Strict, own: ticks.ErrInvalidArgument, deps: []error{meter.ErrUnset}},
		{name: "unset meter before unset click", meter: meter.Meter{}, click: ticks.Click{}, q: ticks.MIDIQuarter{}, f: ticks.Fidelity(0), own: ticks.ErrInvalidArgument, deps: []error{meter.ErrUnset}},
		{name: "unset click", meter: meter.FourFour(), click: ticks.Click{}, q: std, f: ticks.Strict, own: ticks.ErrUnset},
		{name: "unset quarter", meter: meter.FourFour(), click: ticks.ClickOnBeat(), q: ticks.MIDIQuarter{}, f: ticks.Strict, own: ticks.ErrUnset},
		{name: "invalid fidelity", meter: meter.FourFour(), click: ticks.ClickOnBeat(), q: std, f: ticks.Fidelity(0), own: ticks.ErrInvalidArgument},
		{name: "unset click before invalid fidelity", meter: meter.FourFour(), click: ticks.Click{}, q: std, f: ticks.Fidelity(7), own: ticks.ErrUnset},
		{name: "denominators before the record", meter: meter.MustNew(4, 3), click: ticks.ClickOnBeat(), q: std, f: ticks.Strict, own: ticks.ErrMeterNotMIDI, not: []error{meter.ErrGroupingRequired}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, losses, err := ticks.MIDIMeterOf(tt.meter, tt.click, tt.q, tt.f)
			requireTicksError(t, err, tt.own, tt.deps...)
			for _, n := range tt.not {
				assert.NotErrorIs(t, err, n)
			}
			assert.Equal(t, ticks.MIDIMeter{}, got)
			assert.Nil(t, losses)
			requireMustPanic(t, err, tt.own, func() { ticks.MustMIDIMeterOf(tt.meter, tt.click, tt.q, tt.f) })
		})
	}
}

// TestMeterEnumerations checks the enumerations of 4.13, 4.14 and 4.15.
func TestMeterEnumerations(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		f    ticks.Fidelity
		text string
	}{
		{f: ticks.Strict, text: "strict"},
		{f: ticks.Lossy, text: "lossy"},
	} {
		assert.True(t, tt.f.IsValid())
		assert.True(t, tt.f.Equal(tt.f))
		assert.Equal(t, tt.text, tt.f.String())
	}
	assert.Equal(t, "Fidelity(invalid)", ticks.Fidelity(0).String())
	assert.False(t, ticks.Fidelity(3).IsValid())
	for _, tt := range []struct {
		l    ticks.Loss
		text string
	}{
		{l: ticks.LostSymbol, text: "symbol"},
		{l: ticks.LostGrouping, text: "grouping"},
		{l: ticks.LostParts, text: "parts"},
	} {
		assert.True(t, tt.l.IsValid())
		assert.True(t, tt.l.Equal(tt.l))
		assert.Equal(t, tt.text, tt.l.String())
	}
	assert.Equal(t, "Loss(invalid)", ticks.Loss(0).String())
	assert.False(t, ticks.Loss(4).IsValid())
	for _, tt := range []struct {
		b    ticks.IrregularBars
		text string
	}{
		{b: ticks.IrregularEvent, text: "event"},
		{b: ticks.IrregularSkip, text: "skip"},
	} {
		assert.True(t, tt.b.IsValid())
		assert.True(t, tt.b.Equal(tt.b))
		assert.Equal(t, tt.text, tt.b.String())
	}
	assert.Equal(t, "IrregularBars(invalid)", ticks.IrregularBars(0).String())
	assert.False(t, ticks.IrregularBars(3).IsValid())
	assert.False(t, ticks.IrregularEvent.Equal(ticks.IrregularSkip))
}
