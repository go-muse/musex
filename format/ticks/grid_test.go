package ticks_test

import (
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pickupGrid(m meter.Meter, length duration.Duration) measure.Grid {
	return measure.MustNew(m).MustWithPickup(length)
}

// TestAddressTicks checks О-25, О-26, П-20, ТК-17, КП-34.
func TestAddressTicks(t *testing.T) {
	t.Parallel()
	r := res(480)
	g := grid44()
	n, err := r.AddressTicks(g, addr(5, 1, 4))
	require.NoError(t, err)
	assert.Equal(t, 8160, n)
	assert.Equal(t, 8160, r.MustAddressTicks(g, addr(5, 1, 4)))
	a, err := r.Address(g, 8160)
	require.NoError(t, err)
	assert.True(t, a.Equal(addr(5, 1, 4)), "address %s", a)
	assert.True(t, r.MustAddress(g, 8160).Equal(addr(5, 1, 4)))
	start, length, err := r.BarTicks(g, 3)
	require.NoError(t, err)
	assert.Equal(t, 3840, start)
	assert.Equal(t, 1920, length)
	start, length = r.MustBarTicks(g, 3)
	assert.Equal(t, 3840, start)
	assert.Equal(t, 1920, length)
	// A bar line belongs to the next bar.
	assert.True(t, r.MustAddress(g, 1920).Equal(addr(2, 0, 1)))
	assert.True(t, r.MustAddress(g, 0).Equal(addr(1, 0, 1)))

	pickup := pickupGrid(meter.FourFour(), dur(1, 4))
	assert.Equal(t, 0, r.MustAddressTicks(pickup, addr(0, 0, 1)))
	assert.Equal(t, 480, r.MustAddressTicks(pickup, addr(1, 0, 1)))
	assert.True(t, r.MustAddress(pickup, 480).Equal(addr(1, 0, 1)))
	start, length = r.MustBarTicks(pickup, 0)
	assert.Equal(t, 0, start)
	assert.Equal(t, 480, length)

	// The written quarter: bar 2 of 4/4 starts at position 1, 960 ticks with w = 1/2.
	assert.Equal(t, 960, resQ(480, 16).MustAddressTicks(g, addr(2, 0, 1)))

	q, err := res(96).QuantizeAddress(g, addr(1, 1, 7), rounding.HalfUp)
	require.NoError(t, err)
	assertQuantized(t, q, 55, frac(1, 2688))
	assertQuantized(t, res(96).MustQuantizeAddress(g, addr(1, 1, 7), rounding.Down), 54, frac(-1, 448))
	assertQuantized(t, res(96).MustQuantizeAddress(g, addr(2, 0, 1), rounding.Down), 384, frac(0, 1))
}

func TestAddressTicksErrors(t *testing.T) {
	t.Parallel()
	r := res(480)
	g := grid44()
	finite := measure.MustNew(meter.FourFour()).MustWithBarCount(2)
	third := pickupGrid(meter.FourFour(), dur(1, 3))
	for _, tt := range []struct {
		name string
		call func() error
		own  error
		deps []error
	}{
		{name: "address after the grid", call: func() error { _, err := r.AddressTicks(finite, addr(5, 0, 1)); return err }, own: ticks.ErrInvalidArgument, deps: []error{measure.ErrOutsideGrid}},
		{name: "offset outside the bar", call: func() error { _, err := r.AddressTicks(g, addr(1, 5, 4)); return err }, own: ticks.ErrInvalidArgument, deps: []error{measure.ErrOutsideBar}},
		{name: "tick at the end of the grid", call: func() error { _, err := r.Address(finite, 3840); return err }, own: ticks.ErrInvalidArgument, deps: []error{measure.ErrOutsideGrid}},
		{name: "bar after the grid", call: func() error { _, _, err := r.BarTicks(finite, 5); return err }, own: ticks.ErrInvalidArgument},
		{name: "negative bar", call: func() error { _, _, err := r.BarTicks(g, -1); return err }, own: ticks.ErrInvalidArgument},
		{name: "negative tick", call: func() error { _, err := r.Address(g, -1); return err }, own: ticks.ErrInvalidArgument},
		{name: "unset grid", call: func() error { _, err := r.AddressTicks(measure.Grid{}, addr(1, 0, 1)); return err }, own: ticks.ErrInvalidArgument, deps: []error{measure.ErrUnset}},
		{name: "unset address", call: func() error { _, err := r.AddressTicks(g, measure.Address{}); return err }, own: ticks.ErrInvalidArgument, deps: []error{measure.ErrUnset}},
		{name: "unset grid of Address", call: func() error { _, err := r.Address(measure.Grid{}, 0); return err }, own: ticks.ErrInvalidArgument, deps: []error{measure.ErrUnset}},
		{name: "unset mode", call: func() error { _, err := r.QuantizeAddress(g, addr(1, 0, 1), rounding.Mode(0)); return err }, own: ticks.ErrInvalidArgument, deps: []error{rounding.ErrUnset}},
		{name: "address not representable", call: func() error { _, err := res(1).AddressTicks(third, addr(1, 0, 1)); return err }, own: ticks.ErrNotRepresentable},
		{name: "bar not representable", call: func() error { _, _, err := res(1).BarTicks(third, 1); return err }, own: ticks.ErrNotRepresentable},
		{name: "unset receiver", call: func() error { _, err := ticks.Resolution{}.AddressTicks(measure.Grid{}, addr(1, 0, 1)); return err }, own: ticks.ErrUnset},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, tt.call(), tt.own, tt.deps...)
		})
	}
	_, err := r.AddressTicks(finite, addr(5, 0, 1))
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { r.MustAddressTicks(finite, addr(5, 0, 1)) })
	_, _, err = res(1).BarTicks(third, 1)
	requireMustPanic(t, err, ticks.ErrNotRepresentable, func() { res(1).MustBarTicks(third, 1) })
}

// TestMIDIMeterMapOf checks О-27, П-18, ТК-17, КП-35, КП-38.
func TestMIDIMeterMapOf(t *testing.T) {
	t.Parallel()
	r := res(480)
	beat := ticks.ClickOnBeat()
	for _, tt := range []struct {
		name      string
		grid      measure.Grid
		r         ticks.Resolution
		click     ticks.Click
		f         ticks.Fidelity
		irregular ticks.IrregularBars
		want      []string
		losses    []ticks.BarLoss
	}{
		{name: "KP-35 change to 3/4 at bar 3", grid: grid44().MustWithMark(3, meter.ThreeFour()), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, want: []string{"0:4,2,24,8", "3840:3,2,24,8"}},
		{name: "KP-35 pickup as an event", grid: pickupGrid(meter.ThreeFour(), dur(1, 4)), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, want: []string{"0:1,2,24,8", "480:3,2,24,8"}},
		{name: "KP-35 pickup without an event", grid: pickupGrid(meter.ThreeFour(), dur(1, 4)), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularSkip, want: []string{"0:3,2,24,8"}},
		{name: "one meter", grid: measure.MustNew(meter.SixEight()), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, want: []string{"0:6,3,36,8"}},
		{name: "late change", grid: grid44().MustWithMark(5, meter.ThreeFour()), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, want: []string{"0:4,2,24,8", "7680:3,2,24,8"}},
		{name: "repeated mark gives no event", grid: grid44().MustWithMark(3, meter.ThreeFour()).MustWithMark(4, meter.ThreeFour()), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, want: []string{"0:4,2,24,8", "3840:3,2,24,8"}},
		{name: "explicit length as an event", grid: grid44().MustWithLength(2, dur(1, 2), measure.AlignStart), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, want: []string{"0:4,2,24,8", "1920:2,2,24,8", "2880:4,2,24,8"}},
		{name: "explicit length without an event", grid: grid44().MustWithLength(2, dur(1, 2), measure.AlignStart), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularSkip, want: []string{"0:4,2,24,8"}},
		{name: "pickup of an eighth uses the larger denominator", grid: pickupGrid(meter.FourFour(), dur(1, 8)), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, want: []string{"0:1,3,24,8", "240:4,2,24,8"}},
		{name: "explicit click", grid: grid44().MustWithMark(2, meter.SixEight()), r: r, click: ticks.MustClickEvery(dur(1, 8)), f: ticks.Strict, irregular: ticks.IrregularEvent, want: []string{"0:4,2,12,8", "1920:6,3,12,8"}},
		{name: "written half quarter", grid: grid44().MustWithMark(2, meter.ThreeFour()), r: resQ(480, 16), click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, want: []string{"0:4,2,12,16", "960:3,2,12,16"}},
		{
			name: "lossy common time", grid: measure.MustNew(meter.CommonTime()), r: r, click: beat, f: ticks.Lossy, irregular: ticks.IrregularEvent,
			want:   []string{"0:4,2,24,8"},
			losses: []ticks.BarLoss{{Bar: 1, Losses: []ticks.Loss{ticks.LostSymbol}}},
		},
		{
			name: "lossy record equal to the previous event gives no event and no loss", grid: grid44().MustWithMark(2, meter.CommonTime()).MustWithMark(3, meter.ThreeFour()), r: r, click: beat, f: ticks.Lossy, irregular: ticks.IrregularEvent,
			want: []string{"0:4,2,24,8", "3840:3,2,24,8"},
		},
		{
			name: "lossy additive meter", grid: grid44().MustWithMark(2, meter.MustNewAdditive([]int{3, 2, 3}, 8)), r: r, click: ticks.MustClickEvery(dur(1, 8)), f: ticks.Lossy, irregular: ticks.IrregularEvent,
			want:   []string{"0:4,2,12,8", "1920:8,3,12,8"},
			losses: []ticks.BarLoss{{Bar: 2, Losses: []ticks.Loss{ticks.LostGrouping}}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, losses, err := ticks.MIDIMeterMapOf(tt.grid, tt.r, tt.click, tt.f, tt.irregular)
			require.NoError(t, err)
			assert.Equal(t, tt.want, meterChanges(t, m))
			assert.True(t, m.MustResolution().Equal(tt.r))
			require.Len(t, losses, len(tt.losses))
			for i := range tt.losses {
				assertEqualValue(t, tt.losses[i], losses[i])
			}
			if tt.losses == nil {
				assert.Nil(t, losses)
			}
			mustMap, mustLosses := ticks.MustMIDIMeterMapOf(tt.grid, tt.r, tt.click, tt.f, tt.irregular)
			assert.True(t, mustMap.Equal(m))
			assert.Len(t, mustLosses, len(losses))
		})
	}
}

func TestMIDIMeterMapOfErrors(t *testing.T) {
	t.Parallel()
	r := res(480)
	beat := ticks.ClickOnBeat()
	for _, tt := range []struct {
		name      string
		grid      measure.Grid
		r         ticks.Resolution
		click     ticks.Click
		f         ticks.Fidelity
		irregular ticks.IrregularBars
		own       error
		deps      []error
	}{
		{name: "KP-38 pickup of 1/6 as an event", grid: pickupGrid(meter.ThreeFour(), dur(1, 6)), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, own: ticks.ErrMeterNotMIDI},
		{name: "strict symbol", grid: measure.MustNew(meter.CommonTime()), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, own: ticks.ErrMeterNotMIDI},
		{name: "click on beat of 5/8", grid: measure.MustNew(meter.MustNew(5, 8)), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, own: ticks.ErrMeterNotMIDI, deps: []error{meter.ErrGroupingRequired}},
		{name: "bar start not representable", grid: pickupGrid(meter.FourFour(), dur(1, 8)).MustWithMark(2, meter.ThreeFour()), r: res(1), click: beat, f: ticks.Strict, irregular: ticks.IrregularSkip, own: ticks.ErrNotRepresentable},
		{name: "unset grid", grid: measure.Grid{}, r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, own: ticks.ErrInvalidArgument, deps: []error{measure.ErrUnset}},
		{name: "unset grid before unset resolution", grid: measure.Grid{}, r: ticks.Resolution{}, click: ticks.Click{}, f: 0, irregular: 0, own: ticks.ErrInvalidArgument, deps: []error{measure.ErrUnset}},
		{name: "unset resolution", grid: grid44(), r: ticks.Resolution{}, click: beat, f: ticks.Strict, irregular: ticks.IrregularEvent, own: ticks.ErrUnset},
		{name: "unset click", grid: grid44(), r: r, click: ticks.Click{}, f: ticks.Strict, irregular: ticks.IrregularEvent, own: ticks.ErrUnset},
		{name: "invalid fidelity", grid: grid44(), r: r, click: beat, f: ticks.Fidelity(0), irregular: ticks.IrregularEvent, own: ticks.ErrInvalidArgument},
		{name: "invalid irregular bars", grid: grid44(), r: r, click: beat, f: ticks.Strict, irregular: ticks.IrregularBars(0), own: ticks.ErrInvalidArgument},
		{name: "meter error before representability in a bar", grid: pickupGrid(meter.FourFour(), dur(1, 8)).MustWithMark(2, meter.MustNew(4, 3)), r: res(1), click: beat, f: ticks.Strict, irregular: ticks.IrregularSkip, own: ticks.ErrMeterNotMIDI},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, losses, err := ticks.MIDIMeterMapOf(tt.grid, tt.r, tt.click, tt.f, tt.irregular)
			requireTicksError(t, err, tt.own, tt.deps...)
			assert.False(t, m.IsSet())
			assert.Nil(t, losses)
			requireMustPanic(t, err, tt.own, func() { ticks.MustMIDIMeterMapOf(tt.grid, tt.r, tt.click, tt.f, tt.irregular) })
		})
	}
}

// TestMIDIMeterMap checks О-28, П-16, П-19, ТК-17, КП-36, КП-39 and 4.14.
func TestMIDIMeterMap(t *testing.T) {
	t.Parallel()
	r := res(480)
	m, err := ticks.NewMIDIMeterMap(r, mchange(0, 4, 2, 24, 8), mchange(2880, 3, 2, 24, 8))
	require.NoError(t, err)
	assert.True(t, m.IsSet())
	assert.True(t, m.MustResolution().Equal(r))
	assert.Equal(t, []string{"0:4,2,24,8", "2880:3,2,24,8"}, meterChanges(t, m))
	assertEqualValue(t, mmeter(4, 2, 24, 8), m.MustMeterAt(2879))
	assertEqualValue(t, mmeter(3, 2, 24, 8), m.MustMeterAt(2880))
	at, err := m.ChangesAt(2880)
	require.NoError(t, err)
	require.Len(t, at, 1)
	assertEqualValue(t, mmeter(3, 2, 24, 8), at[0])
	assert.Nil(t, m.MustChangesAt(5))

	// КП-36.
	g, err := m.Grid()
	require.NoError(t, err)
	want := grid44().MustWithLength(2, dur(1, 2), measure.AlignStart).MustWithMark(3, meter.ThreeFour())
	assertEqualValue(t, want, g)
	assertEqualValue(t, want, m.MustGrid())
	assertEqualValue(t, grid44(), ticks.MustNewMIDIMeterMap(r).MustGrid())

	// П-19: before the first event 4/4.
	late := ticks.MustNewMIDIMeterMap(r, mchange(1920, 3, 2, 24, 8))
	assertEqualValue(t, grid44().MustWithMark(2, meter.ThreeFour()), late.MustGrid())
	assertEqualValue(t, mmeter(4, 2, 24, 8), late.MustMeterAt(0))
	// The last event on a tick is in effect.
	same := ticks.MustNewMIDIMeterMap(r, mchange(0, 3, 2, 24, 8), mchange(0, 2, 2, 24, 8))
	assertEqualValue(t, measure.MustNew(meter.TwoFour()), same.MustGrid())
	assertEqualValue(t, mmeter(2, 2, 24, 8), same.MustMeterAt(0))
	// An event inside a 3/4 bar ends it.
	inside := ticks.MustNewMIDIMeterMap(r, mchange(0, 3, 2, 24, 8), mchange(960, 2, 2, 24, 8))
	assertEqualValue(t, measure.MustNew(meter.ThreeFour()).MustWithLength(1, dur(1, 2), measure.AlignStart).MustWithMark(2, meter.TwoFour()), inside.MustGrid())
	// Round trip of КП-35.
	kp35 := grid44().MustWithMark(3, meter.ThreeFour())
	exported, _ := ticks.MustMIDIMeterMapOf(kp35, r, ticks.ClickOnBeat(), ticks.Strict, ticks.IrregularEvent)
	assertEqualValue(t, kp35, exported.MustGrid())

	// П-19: the default meter uses the written quarter of the map.
	for _, tt := range []struct {
		name string
		r    ticks.Resolution
		want ticks.MIDIMeter
	}{
		{name: "standard", r: r, want: mmeter(4, 2, 24, 8)},
		{name: "half", r: resQ(480, 16), want: mmeter(4, 2, 24, 16)},
		{name: "bb 10", r: resQ(480, 10), want: mmeter(4, 2, 24, 10)},
		{name: "bb 255", r: resQ(480, 255), want: mmeter(4, 2, 24, 255)},
	} {
		empty := ticks.MustNewMIDIMeterMap(tt.r)
		got, err := empty.MeterAt(100)
		require.NoError(t, err, tt.name)
		assertEqualValue(t, tt.want, got)
	}
	half := ticks.MustNewMIDIMeterMap(resQ(480, 16), mchange(960, 3, 2, 12, 16))
	assertEqualValue(t, grid44().MustWithMark(2, meter.ThreeFour()), half.MustGrid())

	// КП-39: different bb in one map.
	mixed := ticks.MustNewMIDIMeterMap(r, mchange(0, 4, 2, 24, 8), mchange(1920, 4, 2, 24, 16))
	_, err = mixed.Grid()
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { mixed.MustGrid() })
	_, err = ticks.MustNewMIDIMeterMap(resQ(480, 16), mchange(0, 4, 2, 24, 8)).Grid()
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	// bb is checked before the range of the meter.
	_, err = ticks.MustNewMIDIMeterMap(r, mchange(0, 1, 63, 24, 8), mchange(5, 4, 2, 24, 16)).Grid()
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	_, err = ticks.MustNewMIDIMeterMap(r, mchange(0, 1, 63, 24, 8)).Grid()
	requireTicksError(t, err, ticks.ErrOutOfRange)
	_, err = ticks.MIDIMeterMap{}.Grid()
	requireTicksError(t, err, ticks.ErrUnset)
}

func TestMIDIMeterMapEditing(t *testing.T) {
	t.Parallel()
	r := res(480)
	m := ticks.MustNewMIDIMeterMap(r, mchange(0, 4, 2, 24, 8), mchange(0, 3, 2, 24, 8))
	assert.Equal(t, []string{"0:4,2,24,8", "0:3,2,24,8", "1440:2,2,24,8"}, meterChanges(t, m.MustWithChange(mchange(1440, 2, 2, 24, 8))))
	assert.Equal(t, []string{"0:4,2,24,8", "0:3,2,24,8", "0:5,2,24,8"}, meterChanges(t, m.MustWithChange(mchange(0, 5, 2, 24, 8))))
	replaced, err := m.ReplaceChange(0, 1, mmeter(6, 3, 36, 8))
	require.NoError(t, err)
	assert.Equal(t, []string{"0:4,2,24,8", "0:6,3,36,8"}, meterChanges(t, replaced))
	removed, err := m.WithoutChange(0, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{"0:3,2,24,8"}, meterChanges(t, removed))
	assert.True(t, m.MustWithoutChange(9, 0).Equal(m))
	assert.Equal(t, []string{"0:4,2,24,8", "0:3,2,24,8"}, meterChanges(t, m), "receiver changed")

	for _, tt := range []struct {
		name string
		call func() error
		own  error
	}{
		{name: "creation with decreasing ticks", call: func() error {
			_, err := ticks.NewMIDIMeterMap(r, mchange(10, 4, 2, 24, 8), mchange(5, 4, 2, 24, 8))
			return err
		}, own: ticks.ErrOrder},
		{name: "creation with a negative tick", call: func() error { _, err := ticks.NewMIDIMeterMap(r, mchange(-1, 4, 2, 24, 8)); return err }, own: ticks.ErrInvalidArgument},
		{name: "creation with an unset meter", call: func() error { _, err := ticks.NewMIDIMeterMap(r, ticks.MIDIMeterChange{Tick: 1}); return err }, own: ticks.ErrUnset},
		{name: "creation with an unset resolution", call: func() error { _, err := ticks.NewMIDIMeterMap(ticks.Resolution{}); return err }, own: ticks.ErrUnset},
		{name: "replace a missing change", call: func() error { _, err := m.ReplaceChange(0, 2, mmeter(2, 2, 24, 8)); return err }, own: ticks.ErrInvalidArgument},
		{name: "replace with an unset meter", call: func() error { _, err := m.ReplaceChange(0, 0, ticks.MIDIMeter{}); return err }, own: ticks.ErrUnset},
		{name: "remove with a negative index", call: func() error { _, err := m.WithoutChange(0, -1); return err }, own: ticks.ErrInvalidArgument},
		{name: "meter at a negative tick", call: func() error { _, err := m.MeterAt(-1); return err }, own: ticks.ErrInvalidArgument},
		{name: "add with a negative tick", call: func() error { _, err := m.WithChange(mchange(-1, 4, 2, 24, 8)); return err }, own: ticks.ErrInvalidArgument},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, tt.call(), tt.own)
		})
	}

	// Rescaling.
	two := ticks.MustNewMIDIMeterMap(r, mchange(0, 4, 2, 24, 8), mchange(3840, 3, 2, 24, 8))
	rescaled, err := two.Rescale(res(96))
	require.NoError(t, err)
	assert.Equal(t, []string{"0:4,2,24,8", "768:3,2,24,8"}, meterChanges(t, rescaled))
	assert.True(t, rescaled.Equal(two.MustRescale(res(96))))
	odd := ticks.MustNewMIDIMeterMap(r, mchange(1, 4, 2, 24, 8), mchange(3, 3, 2, 24, 8))
	_, err = odd.Rescale(res(96))
	requireTicksError(t, err, ticks.ErrNotRepresentable)
	_, err = two.Rescale(resQ(96, 16))
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	quantized, err := odd.QuantizeRescale(res(96), rounding.Down)
	require.NoError(t, err)
	assert.Equal(t, []string{"0:4,2,24,8", "0:3,2,24,8"}, meterChanges(t, quantized))
	assert.Equal(t, []string{"0:4,2,24,8", "1:3,2,24,8"}, meterChanges(t, odd.MustQuantizeRescale(res(96), rounding.HalfUp)))

	// Equality and returned slices.
	assert.True(t, two.Equal(ticks.MustNewMIDIMeterMap(r, mchange(0, 4, 2, 24, 8), mchange(3840, 3, 2, 24, 8))))
	assert.False(t, two.Equal(rescaled))
	assert.False(t, ticks.MustNewMIDIMeterMap(r).Equal(ticks.MustNewMIDIMeterMap(res(96))))
	changes := two.MustChanges()
	changes[0] = mchange(9, 9, 2, 24, 8)
	assert.Equal(t, []string{"0:4,2,24,8", "3840:3,2,24,8"}, meterChanges(t, two))
	assert.Nil(t, ticks.MustNewMIDIMeterMap(r).MustChanges())
	assert.True(t, ticks.MIDIMeterMap{}.Equal(ticks.MIDIMeterMap{}))
	assert.Equal(t, "MIDIMeterMap(unset)", ticks.MIDIMeterMap{}.String())
}
