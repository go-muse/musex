package ticks_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"
	"github.com/go-muse/musex/tempo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kp10Map is the MIDI map of КП-10: D = 96, 500000 from tick 0, 1000000 from tick 192.
func kp10Map() ticks.MIDITempoMap {
	return ticks.MustNewMIDITempoMap(res(96), tchange(0, 500000), tchange(192, 1000000))
}

// TestMIDITempoMapCreation checks О-15, П-6, П-7, ТК-9.
func TestMIDITempoMapCreation(t *testing.T) {
	t.Parallel()
	empty, err := ticks.NewMIDITempoMap(res(96))
	require.NoError(t, err)
	assert.True(t, empty.IsSet())
	assert.True(t, empty.MustResolution().Equal(res(96)))
	r, err := empty.Resolution()
	require.NoError(t, err)
	assert.True(t, r.Equal(res(96)))
	initial, err := empty.Initial()
	require.NoError(t, err)
	assert.Equal(t, 500000, initial.MustMicroseconds())
	changes, err := empty.Changes()
	require.NoError(t, err)
	assert.Nil(t, changes)
	assert.Equal(t, 500000, empty.MustTempoAt(0).MustMicroseconds())
	assert.Equal(t, 500000, empty.MustTempoAt(1<<40).MustMicroseconds())

	m := kp10Map()
	assert.Equal(t, []string{"0:500000", "192:1000000"}, tempoChanges(t, m))
	assert.Equal(t, 500000, m.MustTempoAt(0).MustMicroseconds())
	assert.Equal(t, 500000, m.MustTempoAt(191).MustMicroseconds())
	assert.Equal(t, 1000000, m.MustTempoAt(192).MustMicroseconds())
	assert.Equal(t, 1000000, m.MustTempoAt(10000).MustMicroseconds())
	at, err := m.ChangesAt(192)
	require.NoError(t, err)
	require.Len(t, at, 1)
	assert.Equal(t, 1000000, at[0].MustMicroseconds())
	none, err := m.ChangesAt(100)
	require.NoError(t, err)
	assert.Nil(t, none)

	// П-7, КП-12: all events on a tick are kept; the last one is in effect.
	same := ticks.MustNewMIDITempoMap(res(96), tchange(0, 500000), tchange(0, 600000))
	assert.Equal(t, []string{"0:500000", "0:600000"}, tempoChanges(t, same))
	onZero := same.MustChangesAt(0)
	require.Len(t, onZero, 2)
	assert.Equal(t, 500000, onZero[0].MustMicroseconds())
	assert.Equal(t, 600000, onZero[1].MustMicroseconds())
	assert.Equal(t, 600000, same.MustTempoAt(0).MustMicroseconds())
	// Repeated identical changes are kept.
	repeated := ticks.MustNewMIDITempoMap(res(96), tchange(0, 500000), tchange(0, 500000), tchange(5, 500000))
	assert.Equal(t, []string{"0:500000", "0:500000", "5:500000"}, tempoChanges(t, repeated))
}

func TestMIDITempoMapCreationErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		r       ticks.Resolution
		changes []ticks.MIDITempoChange
		own     error
	}{
		{name: "unset resolution", r: ticks.Resolution{}, own: ticks.ErrUnset},
		{name: "unset resolution first", r: ticks.Resolution{}, changes: []ticks.MIDITempoChange{tchange(5, 1), tchange(1, 1)}, own: ticks.ErrUnset},
		{name: "ticks decrease", r: res(96), changes: []ticks.MIDITempoChange{tchange(10, 500000), tchange(5, 500000)}, own: ticks.ErrOrder},
		{name: "negative tick", r: res(96), changes: []ticks.MIDITempoChange{tchange(-1, 500000)}, own: ticks.ErrInvalidArgument},
		{name: "unset tempo", r: res(96), changes: []ticks.MIDITempoChange{{Tick: 0}}, own: ticks.ErrUnset},
		{name: "unset tempo before negative tick", r: res(96), changes: []ticks.MIDITempoChange{{Tick: -1}}, own: ticks.ErrUnset},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := ticks.NewMIDITempoMap(tt.r, tt.changes...)
			requireTicksError(t, err, tt.own)
			assert.False(t, m.IsSet())
			requireMustPanic(t, err, tt.own, func() { ticks.MustNewMIDITempoMap(tt.r, tt.changes...) })
		})
	}
}

// TestMIDITempoMapEditing checks О-15, КП-43 and 4.9.
func TestMIDITempoMapEditing(t *testing.T) {
	t.Parallel()
	m := kp10Map()
	before := m.String()

	withInitial, err := m.WithInitial(mtempo(1000000))
	require.NoError(t, err)
	assert.Equal(t, 1000000, withInitial.MustInitial().MustMicroseconds())
	assert.Equal(t, before, m.String(), "receiver changed")

	added, err := m.WithChange(tchange(192, 700000))
	require.NoError(t, err)
	assert.Equal(t, []string{"0:500000", "192:1000000", "192:700000"}, tempoChanges(t, added))
	assert.Equal(t, 700000, added.MustTempoAt(192).MustMicroseconds())
	inserted := m.MustWithChange(tchange(96, 800000))
	assert.Equal(t, []string{"0:500000", "96:800000", "192:1000000"}, tempoChanges(t, inserted))
	assert.Equal(t, []string{"0:500000", "0:400000", "192:1000000"}, tempoChanges(t, m.MustWithChange(tchange(0, 400000))))

	same := ticks.MustNewMIDITempoMap(res(96), tchange(0, 500000), tchange(0, 600000))
	replaced, err := same.ReplaceChange(0, 1, mtempo(700000))
	require.NoError(t, err)
	assert.Equal(t, []string{"0:500000", "0:700000"}, tempoChanges(t, replaced))
	assert.Equal(t, []string{"0:450000", "0:600000"}, tempoChanges(t, same.MustReplaceChange(0, 0, mtempo(450000))))

	// КП-43: removing the second change on tick 0.
	removed, err := same.WithoutChange(0, 1)
	require.NoError(t, err)
	assert.True(t, removed.Equal(ticks.MustNewMIDITempoMap(res(96), tchange(0, 500000))))
	assert.Equal(t, []string{"0:600000"}, tempoChanges(t, same.MustWithoutChange(0, 0)))
	// A missing change: the same map without an error.
	missing, err := same.WithoutChange(0, 2)
	require.NoError(t, err)
	assert.True(t, missing.Equal(same))
	assert.True(t, same.MustWithoutChange(5, 0).Equal(same))

	for _, tt := range []struct {
		name string
		call func() error
		own  error
	}{
		{name: "replace a missing change", call: func() error { _, err := same.ReplaceChange(0, 2, mtempo(1)); return err }, own: ticks.ErrInvalidArgument},
		{name: "replace on an empty tick", call: func() error { _, err := same.ReplaceChange(7, 0, mtempo(1)); return err }, own: ticks.ErrInvalidArgument},
		{name: "replace with a negative tick", call: func() error { _, err := same.ReplaceChange(-1, 0, mtempo(1)); return err }, own: ticks.ErrInvalidArgument},
		{name: "replace with a negative index", call: func() error { _, err := same.ReplaceChange(0, -1, mtempo(1)); return err }, own: ticks.ErrInvalidArgument},
		{name: "replace with an unset tempo", call: func() error { _, err := same.ReplaceChange(0, 0, ticks.MIDITempo{}); return err }, own: ticks.ErrUnset},
		{name: "remove with a negative tick", call: func() error { _, err := same.WithoutChange(-1, 0); return err }, own: ticks.ErrInvalidArgument},
		{name: "remove with a negative index", call: func() error { _, err := same.WithoutChange(0, -1); return err }, own: ticks.ErrInvalidArgument},
		{name: "add with a negative tick", call: func() error { _, err := same.WithChange(tchange(-1, 1)); return err }, own: ticks.ErrInvalidArgument},
		{name: "add with an unset tempo", call: func() error { _, err := same.WithChange(ticks.MIDITempoChange{Tick: 1}); return err }, own: ticks.ErrUnset},
		{name: "unset initial tempo", call: func() error { _, err := same.WithInitial(ticks.MIDITempo{}); return err }, own: ticks.ErrUnset},
		{name: "unset map", call: func() error { _, err := ticks.MIDITempoMap{}.WithChange(tchange(0, 1)); return err }, own: ticks.ErrUnset},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, tt.call(), tt.own)
		})
	}
	_, err = same.ReplaceChange(0, 2, mtempo(1))
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { same.MustReplaceChange(0, 2, mtempo(1)) })
	assert.Equal(t, []string{"0:500000", "0:600000"}, tempoChanges(t, same), "receiver changed by a failed edit")
}

// TestMIDITempoMapTime checks О-16, ИТ-8, ТК-9, КП-10, КП-11, КП-44.
func TestMIDITempoMapTime(t *testing.T) {
	t.Parallel()
	m := kp10Map()
	for _, tt := range []struct {
		tick    int
		seconds tempo.Seconds
	}{
		{tick: 0, seconds: tempo.ZeroSeconds()},
		{tick: 48, seconds: sec(1, 4)},
		{tick: 96, seconds: sec(1, 2)},
		{tick: 192, seconds: sec(1, 1)},
		{tick: 288, seconds: sec(2, 1)},
		{tick: 384, seconds: sec(3, 1)},
		{tick: 1, seconds: sec(1, 192)},
		{tick: 193, seconds: sec(97, 96)},
	} {
		s, err := m.Seconds(tt.tick)
		require.NoError(t, err)
		assertSeconds(t, tt.seconds, s)
		assertSeconds(t, tt.seconds, m.MustSeconds(tt.tick))
		in, err := m.TicksAt(tt.seconds)
		require.NoError(t, err)
		assertFraction(t, fraction.FromInt(tt.tick), in)
		assertFraction(t, fraction.FromInt(tt.tick), m.MustTicksAt(tt.seconds))
		for _, mode := range allModes {
			assert.Equal(t, tt.tick, m.MustTickAt(tt.seconds, mode))
		}
		p := res(96).MustPosition(tt.tick)
		ps, err := m.PositionSeconds(p)
		require.NoError(t, err)
		assertSeconds(t, tt.seconds, ps)
		assertSeconds(t, tt.seconds, m.MustPositionSeconds(p))
		back, err := m.PositionAt(tt.seconds)
		require.NoError(t, err)
		assertEqualValue(t, p, back)
		assertEqualValue(t, p, m.MustPositionAt(tt.seconds))
	}
	assertFraction(t, frac(5, 2), m.MustTicksAt(sec(5, 384)))
	assertEqualValue(t, pos(5, 768), m.MustPositionAt(sec(5, 384)))
	assertSeconds(t, sec(1, 768), m.MustPositionSeconds(pos(1, 1536)))

	// КП-44.
	between, err := m.SecondsBetween(96, 288)
	require.NoError(t, err)
	assertSeconds(t, sec(3, 2), between)
	assertSeconds(t, sec(3, 2), m.MustSecondsBetween(96, 288))
	assertSeconds(t, tempo.ZeroSeconds(), m.MustSecondsBetween(5, 5))
	_, err = m.SecondsBetween(288, 96)
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { m.MustSecondsBetween(288, 96) })

	// КП-11: the default initial tempo is 500000 and can be replaced.
	late := ticks.MustNewMIDITempoMap(res(96), tchange(96, 1000000))
	assertSeconds(t, sec(3, 2), late.MustSeconds(192))
	assertSeconds(t, sec(2, 1), late.MustWithInitial(mtempo(1000000)).MustSeconds(192))
	assert.Equal(t, 500000, late.MustTempoAt(95).MustMicroseconds())

	// The last change on a tick is in effect for time as well.
	same := ticks.MustNewMIDITempoMap(res(96), tchange(0, 500000), tchange(0, 600000))
	assertSeconds(t, sec(3, 5), same.MustSeconds(96))

	// The written quarter: position p is p·D/w ticks.
	half := ticks.MustNewMIDITempoMap(resQ(96, 16), tchange(0, 500000))
	assertSeconds(t, sec(1, 2), half.MustPositionSeconds(pos(1, 2)))
	assertEqualValue(t, pos(1, 2), half.MustPositionAt(sec(1, 2)))
}

// TestMIDITempoMapTickAtRounding checks all nine modes on ties (ИТ-11).
func TestMIDITempoMapTickAtRounding(t *testing.T) {
	t.Parallel()
	m := kp10Map()
	for _, mode := range allModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			got, err := m.TickAt(sec(5, 384), mode)
			require.NoError(t, err)
			assert.Equal(t, int(tieIndex(mode, 2, true)), got)
			assert.Equal(t, int(tieIndex(mode, 192, true)), m.MustTickAt(sec(193, 192), mode))
			assert.Equal(t, int(tieIndex(mode, 193, false)), m.MustTickAt(sec(65, 64), mode))
		})
	}
}

func TestMIDITempoMapTimeErrors(t *testing.T) {
	t.Parallel()
	m := kp10Map()
	for _, tt := range []struct {
		name string
		call func() error
		own  error
		deps []error
	}{
		{name: "negative tick seconds", call: func() error { _, err := m.Seconds(-1); return err }, own: ticks.ErrInvalidArgument},
		{name: "negative tick tempo", call: func() error { _, err := m.TempoAt(-1); return err }, own: ticks.ErrInvalidArgument},
		{name: "negative tick changes", call: func() error { _, err := m.ChangesAt(-1); return err }, own: ticks.ErrInvalidArgument},
		{name: "negative from", call: func() error { _, err := m.SecondsBetween(-1, 5); return err }, own: ticks.ErrInvalidArgument},
		{name: "unset seconds", call: func() error { _, err := m.TicksAt(tempo.Seconds{}); return err }, own: ticks.ErrInvalidArgument, deps: []error{tempo.ErrUnset}},
		{name: "unset seconds tick", call: func() error { _, err := m.TickAt(tempo.Seconds{}, rounding.Down); return err }, own: ticks.ErrInvalidArgument, deps: []error{tempo.ErrUnset}},
		{name: "unset mode", call: func() error { _, err := m.TickAt(sec(1, 1), rounding.Mode(0)); return err }, own: ticks.ErrInvalidArgument, deps: []error{rounding.ErrUnset}},
		{name: "unset position", call: func() error { _, err := m.PositionSeconds(duration.Position{}); return err }, own: ticks.ErrInvalidArgument, deps: []error{duration.ErrUnset}},
		{name: "unset seconds position", call: func() error { _, err := m.PositionAt(tempo.Seconds{}); return err }, own: ticks.ErrInvalidArgument, deps: []error{tempo.ErrUnset}},
		{name: "unset map", call: func() error { _, err := ticks.MIDITempoMap{}.Seconds(1); return err }, own: ticks.ErrUnset},
		{name: "unset map before argument", call: func() error { _, err := ticks.MIDITempoMap{}.Seconds(-1); return err }, own: ticks.ErrUnset},
		{name: "ticks outside int", call: func() error {
			_, err := m.TickAt(tempo.MustSecondsFromFraction(hugeFraction(70)), rounding.Down)
			return err
		}, own: ticks.ErrOutOfRange},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, tt.call(), tt.own, tt.deps...)
		})
	}
}

// TestMIDITempoMapSMPTE checks О-21, ИТ-8, ИТ-9.
func TestMIDITempoMapSMPTE(t *testing.T) {
	t.Parallel()
	m := kp10Map()
	s := smpte(-25, 40)
	n, err := m.SMPTETick(s, 384, rounding.HalfUp)
	require.NoError(t, err)
	assert.Equal(t, 3000, n)
	assert.Equal(t, 3000, m.MustSMPTETick(s, 384, rounding.Down))
	back, err := m.TickFromSMPTE(s, 3000, rounding.HalfUp)
	require.NoError(t, err)
	assert.Equal(t, 384, back)
	assert.Equal(t, 384, m.MustTickFromSMPTE(s, 3000, rounding.Up))
	// Tick 1 is 1/192 s = 125/24 SMPTE ticks ≈ 5.21.
	assert.Equal(t, 5, m.MustSMPTETick(s, 1, rounding.HalfUp))
	assert.Equal(t, 6, m.MustSMPTETick(s, 1, rounding.Up))
	// 5 SMPTE ticks are 1/200 s = 0.96 MIDI ticks.
	assert.Equal(t, 1, m.MustTickFromSMPTE(s, 5, rounding.HalfUp))
	assert.Equal(t, 0, m.MustTickFromSMPTE(s, 5, rounding.Down))
	// MIDI tick 3 is 3/192 s = 15.625 SMPTE ticks: the nearest is 16.
	for _, mode := range allModes {
		assert.Equal(t, int(irrationalIndex(mode, 15, true)), m.MustSMPTETick(s, 3, mode))
	}

	df := smpte(-29, 80)
	for name, call := range map[string]func() error{
		"KP-13 -29 without reading": func() error { _, err := m.SMPTETick(df, 1, rounding.Down); return err },
		"from -29 without reading":  func() error { _, err := m.TickFromSMPTE(df, 1, rounding.Down); return err },
		"negative tick":             func() error { _, err := m.SMPTETick(s, -1, rounding.Down); return err },
		"negative SMPTE tick":       func() error { _, err := m.TickFromSMPTE(s, -1, rounding.Down); return err },
		"unset mode":                func() error { _, err := m.SMPTETick(s, 1, rounding.Mode(0)); return err },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, call(), ticks.ErrInvalidArgument)
		})
	}
	_, err = m.SMPTETick(ticks.SMPTE{}, 1, rounding.Down)
	requireTicksError(t, err, ticks.ErrUnset)
	requireMustPanic(t, err, ticks.ErrUnset, func() { m.MustSMPTETick(ticks.SMPTE{}, 1, rounding.Down) })
	assert.Equal(t, 2400, m.MustSMPTETick(df.MustWithReading(ticks.Reading30), 192, rounding.Down))
}

// TestMIDITempoMapRescale checks О-8 for maps and 4.9.
func TestMIDITempoMapRescale(t *testing.T) {
	t.Parallel()
	m := ticks.MustNewMIDITempoMap(res(480), tchange(0, 500000), tchange(100, 600000), tchange(120, 700000))
	got, err := m.Rescale(res(96))
	require.NoError(t, err)
	assert.Equal(t, []string{"0:500000", "20:600000", "24:700000"}, tempoChanges(t, got))
	assert.True(t, got.MustResolution().Equal(res(96)))
	assert.Equal(t, 500000, got.MustInitial().MustMicroseconds())
	assert.True(t, got.Equal(m.MustRescale(res(96))))
	ok, err := got.EqualTimeline(m)
	require.NoError(t, err)
	assert.True(t, ok)

	fine := ticks.MustNewMIDITempoMap(res(480), tchange(1, 600000), tchange(3, 700000), tchange(7, 800000))
	_, err = fine.Rescale(res(96))
	requireTicksError(t, err, ticks.ErrNotRepresentable)
	requireMustPanic(t, err, ticks.ErrNotRepresentable, func() { fine.MustRescale(res(96)) })
	_, err = m.Rescale(resQ(96, 16))
	requireTicksError(t, err, ticks.ErrInvalidArgument)
	_, err = m.Rescale(ticks.Resolution{})
	requireTicksError(t, err, ticks.ErrUnset)

	quantized, err := fine.QuantizeRescale(res(96), rounding.Down)
	require.NoError(t, err)
	assert.Equal(t, []string{"0:600000", "0:700000", "1:800000"}, tempoChanges(t, quantized))
	assert.Equal(t, 700000, quantized.MustTempoAt(0).MustMicroseconds())
	assert.Equal(t, []string{"0:600000", "1:700000", "1:800000"}, tempoChanges(t, fine.MustQuantizeRescale(res(96), rounding.HalfUp)))
	_, err = fine.QuantizeRescale(res(96), rounding.Mode(0))
	requireTicksError(t, err, ticks.ErrInvalidArgument, rounding.ErrUnset)
	_, err = fine.QuantizeRescale(resQ(96, 16), rounding.Down)
	requireTicksError(t, err, ticks.ErrInvalidArgument)
}

// TestMIDITempoMapEquality checks Equal and EqualTimeline (4.9, КП-12).
func TestMIDITempoMapEquality(t *testing.T) {
	t.Parallel()
	same := ticks.MustNewMIDITempoMap(res(96), tchange(0, 500000), tchange(0, 600000))
	single := ticks.MustNewMIDITempoMap(res(96), tchange(0, 600000))
	assert.False(t, same.Equal(single))
	ok, err := same.EqualTimeline(single)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, same.MustEqualTimeline(single))

	empty := ticks.MustNewMIDITempoMap(res(96))
	explicit := ticks.MustNewMIDITempoMap(res(96), tchange(0, 500000))
	assert.False(t, empty.Equal(explicit))
	assert.True(t, empty.MustEqualTimeline(explicit))
	assert.True(t, empty.Equal(ticks.MustNewMIDITempoMap(res(96))))
	assert.False(t, empty.Equal(empty.MustWithInitial(mtempo(600000))))
	assert.False(t, empty.Equal(ticks.MustNewMIDITempoMap(res(97))))

	coarse := ticks.MustNewMIDITempoMap(res(96), tchange(0, 500000), tchange(96, 1000000))
	finer := ticks.MustNewMIDITempoMap(res(192), tchange(0, 500000), tchange(192, 1000000))
	assert.True(t, coarse.MustEqualTimeline(finer))
	assert.False(t, coarse.Equal(finer))
	assert.False(t, coarse.MustEqualTimeline(ticks.MustNewMIDITempoMap(res(96), tchange(0, 500000), tchange(97, 1000000))))
	assert.False(t, single.MustEqualTimeline(explicit))

	// A written half quarter halves the ticks of a position: 1000000 per such
	// quarter is the same speed as 500000 per standard quarter.
	halfQuarter := ticks.MustNewMIDITempoMap(resQ(96, 16), tchange(0, 1000000))
	assert.True(t, explicit.MustEqualTimeline(halfQuarter))

	_, err = same.EqualTimeline(ticks.MIDITempoMap{})
	requireTicksError(t, err, ticks.ErrUnset)
	assert.True(t, ticks.MIDITempoMap{}.Equal(ticks.MIDITempoMap{}))
	assert.False(t, ticks.MIDITempoMap{}.Equal(empty))
	assert.False(t, empty.Equal(ticks.MIDITempoMap{}))
}

// TestMIDITempoMapReturnedSlicesAreIndependent checks 3.2.
func TestMIDITempoMapReturnedSlicesAreIndependent(t *testing.T) {
	t.Parallel()
	m := kp10Map()
	changes := m.MustChanges()
	changes[0] = tchange(7, 1)
	at := m.MustChangesAt(192)
	at[0] = mtempo(2)
	assert.Equal(t, []string{"0:500000", "192:1000000"}, tempoChanges(t, m))
	assert.Equal(t, 1000000, m.MustChangesAt(192)[0].MustMicroseconds())

	input := []ticks.MIDITempoChange{tchange(0, 500000), tchange(5, 600000)}
	built := ticks.MustNewMIDITempoMap(res(96), input...)
	input[1] = tchange(6, 1)
	assert.Equal(t, []string{"0:500000", "5:600000"}, tempoChanges(t, built))

	copied := built
	edited := copied.MustWithChange(tchange(9, 9))
	assert.Equal(t, []string{"0:500000", "5:600000"}, tempoChanges(t, built))
	assert.Equal(t, []string{"0:500000", "5:600000"}, tempoChanges(t, copied))
	assert.Len(t, edited.MustChanges(), 3)
}

// TestMIDITempoMapMusicalTimeline checks KP-49 and IT-12 at musical
// positions, including positions whose tick coordinate is fractional.
func TestMIDITempoMapMusicalTimeline(t *testing.T) {
	t.Parallel()
	a := ticks.MustNewMIDITempoMap(res(480), tchange(480, 600000))
	b := ticks.MustNewMIDITempoMap(res(960), tchange(960, 600000))
	c := ticks.MustNewMIDITempoMap(res(960), tchange(1920, 600000))
	for _, p := range []duration.Position{pos(1, 1921), pos(1, 4), pos(1, 2), pos(2, 1)} {
		assertSeconds(t, a.MustPositionSeconds(p), b.MustPositionSeconds(p))
	}
	assertSeconds(t, sec(1, 2), a.MustPositionSeconds(pos(1, 4)))
	assertSeconds(t, sec(11, 10), b.MustPositionSeconds(pos(1, 2)))
	assertSeconds(t, sec(1, 1), c.MustPositionSeconds(pos(1, 2)))

	for _, tt := range []struct {
		name string
		a, b ticks.MIDITempoMap
		want bool
	}{
		{name: "KP-49 rescaled ticks", a: a, b: b, want: true},
		{name: "KP-49 moved change", a: a, b: c},
		{name: "KP-49 constant tempos", a: ticks.MustNewMIDITempoMap(res(480)), b: ticks.MustNewMIDITempoMap(res(960)).MustWithInitial(mtempo(600000))},
		{name: "redundant changes with a different initial record", a: a, b: b.MustWithInitial(mtempo(1)).MustWithChange(tchange(0, 500000)).MustWithChange(tchange(1920, 600000)), want: true},
		// Both give 1 s at 1/2 and thereafter, but differ inside [0,1/2).
		{name: "equal endpoint is insufficient", a: ticks.MustNewMIDITempoMap(res(480)), b: ticks.MustNewMIDITempoMap(res(960), tchange(0, 250000), tchange(960, 750000), tchange(1920, 500000))},
		// Same seconds per tick, different seconds per musical position.
		{name: "tick slope is insufficient", a: ticks.MustNewMIDITempoMap(res(480)), b: ticks.MustNewMIDITempoMap(res(960)).MustWithInitial(mtempo(1000000))},
		{name: "written quarter and resolution both change", a: a, b: ticks.MustNewMIDITempoMap(resQ(1440, 16), tchange(720, 1200000)).MustWithInitial(mtempo(1000000)), want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.a.EqualTimeline(tt.b)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want, tt.b.MustEqualTimeline(tt.a), "symmetry")
			assert.False(t, tt.a.Equal(tt.b), "record equality remains separate")
		})
	}
}
