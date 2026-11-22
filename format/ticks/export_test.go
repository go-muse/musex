package ticks_test

import (
	"fmt"
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

// Expected tempos of gradual steps (КП-15 map, D = 480, step 1/4) were
// computed outside Go with 80 significant digits (Python decimal) from
// tempomap ИК-20: with speed s(p) = 30 − 5p whole notes per minute, the step
// [q₁, q₂) lasts T = 12·ln(s(q₁)/s(q₂)) seconds, and with w = 1/4 and
// Δp = 1/4 the "equal time" tempo is 10⁶·w·T/Δp = 10⁶·T microseconds:
//
//	step 0: 510715.373025551146…
//	step 1: 533421.150850006047…
//	step 2: 558240.187618714283…
//	step 3: 585481.970033184036…
//
// The fermata stretch of the last step (hold E = 1 s at position 1, L = 1/4)
// uses the time S of the base step, and multiplies its tempo by (S + E)/S.
// A quarter-long step has m = 10⁶·S, so its tempo increases by exactly 10⁶:
//
//	StepStart:     4000000/7 + 1000000 = 11000000/7 µs
//	StepEnd:       600000 + 1000000 = 1600000 µs
//	StepEqualTime: 10⁶·S·(S+1)/S = 10⁶·(S+1) = 1585481.970033184036…
//
// The gradual change from quarter = 4 to quarter = 3 over one whole note lasts
// T = 240·ln(4/3) s; one step of a whole note has the tempo
// 10⁶·(1/4)·T = 17260924.347106855646… > 16777215.

func exportOptions(t *testing.T, mode rounding.Mode) ticks.ExportOptions {
	t.Helper()
	o, err := ticks.NewExportOptions(mode)
	require.NoError(t, err)
	return o
}

// TestExportOptions checks О-18, КР-TK-4 and 4.10.
func TestExportOptions(t *testing.T) {
	t.Parallel()
	o := exportOptions(t, rounding.HalfUp)
	assert.True(t, o.IsSet())
	assert.True(t, o.Equal(ticks.MustNewExportOptions(rounding.HalfUp)))
	mode, err := o.TempoMode()
	require.NoError(t, err)
	assert.Equal(t, rounding.HalfUp, mode)
	assert.Equal(t, rounding.HalfUp, o.MustTempoMode())
	pm, ok, err := o.PositionMode()
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Zero(t, pm)
	step, method, ok, err := o.Steps()
	require.NoError(t, err)
	assert.False(t, ok)
	assert.False(t, step.IsSet())
	assert.Zero(t, method)
	clock, err := o.ClockSteps()
	require.NoError(t, err)
	assert.False(t, clock)
	holds, err := o.Holds()
	require.NoError(t, err)
	assert.Equal(t, ticks.HoldsRefuse, holds)
	length, ok, err := o.StretchLength()
	require.NoError(t, err)
	assert.False(t, ok)
	assert.False(t, length.IsSet())

	withPositions := o.MustWithPositionMode(rounding.Down)
	pm, ok = withPositions.MustPositionMode()
	assert.True(t, ok)
	assert.Equal(t, rounding.Down, pm)
	assert.False(t, withPositions.Equal(o))
	assert.True(t, withPositions.MustWithExactPositions() == o)
	assert.True(t, withPositions.MustWithExactPositions().Equal(o))

	withSteps, err := o.WithSteps(dur(1, 4), ticks.StepEqualTime)
	require.NoError(t, err)
	step, method, ok = withSteps.MustSteps()
	assert.True(t, ok)
	assertEqualValue(t, dur(1, 4), step)
	assert.Equal(t, ticks.StepEqualTime, method)
	assert.True(t, withSteps.MustWithoutSteps() == o)
	assert.True(t, withSteps.MustWithoutSteps().Equal(o))

	onClock := o.MustWithClockSteps(true)
	assert.True(t, onClock.MustClockSteps())
	assert.True(t, onClock.MustWithClockSteps(false) == o)

	skipped := o.MustWithHoldsSkipped()
	assert.Equal(t, ticks.HoldsSkip, skipped.MustHolds())
	stretched := o.MustWithHoldsStretched(dur(1, 4))
	assert.Equal(t, ticks.HoldsStretch, stretched.MustHolds())
	length, ok = stretched.MustStretchLength()
	assert.True(t, ok)
	assertEqualValue(t, dur(1, 4), length)
	assert.True(t, stretched.MustWithHoldsSkipped() == skipped)
	assert.True(t, stretched.MustWithHoldsRefused() == o)
	assert.True(t, stretched.MustWithHoldsRefused().Equal(o))
	_, ok = skipped.MustStretchLength()
	assert.False(t, ok)
	assert.False(t, stretched.Equal(o.MustWithHoldsStretched(dur(1, 2))))

	// Non-Must variants must reset exactly the same fields as their wrappers.
	reset, err := withPositions.WithExactPositions()
	require.NoError(t, err)
	assert.Equal(t, o, reset)
	reset, err = withSteps.WithoutSteps()
	require.NoError(t, err)
	assert.Equal(t, o, reset)
	reset, err = stretched.WithHoldsRefused()
	require.NoError(t, err)
	assert.Equal(t, o, reset)
	assert.True(t, ticks.MustParseExportOptions(o.String()).Equal(o))
	_, err = ticks.ParseExportOptions("invalid")
	requireMustPanic(t, err, ticks.ErrInvalidText, func() { ticks.MustParseExportOptions("invalid") })

	// Enumerations of 4.10.
	for _, tt := range []struct {
		m    ticks.StepMethod
		text string
	}{
		{m: ticks.StepStart, text: "start"},
		{m: ticks.StepEnd, text: "end"},
		{m: ticks.StepEqualTime, text: "equal-time"},
	} {
		assert.True(t, tt.m.IsValid())
		assert.True(t, tt.m.Equal(tt.m))
		assert.Equal(t, tt.text, tt.m.String())
	}
	assert.False(t, ticks.StepMethod(0).IsValid())
	assert.Equal(t, "StepMethod(invalid)", ticks.StepMethod(4).String())
	for _, tt := range []struct {
		h    ticks.HoldPolicy
		text string
	}{
		{h: ticks.HoldsRefuse, text: "refuse"},
		{h: ticks.HoldsSkip, text: "skip"},
		{h: ticks.HoldsStretch, text: "stretch"},
	} {
		assert.True(t, tt.h.IsValid())
		assert.Equal(t, tt.text, tt.h.String())
	}
	assert.False(t, ticks.HoldPolicy(0).IsValid())
	assert.Equal(t, "HoldPolicy(invalid)", ticks.HoldPolicy(9).String())
	assert.False(t, ticks.HoldsSkip.Equal(ticks.HoldsRefuse))
}

func TestExportOptionsErrors(t *testing.T) {
	t.Parallel()
	o := exportOptions(t, rounding.HalfUp)
	var unset ticks.ExportOptions
	for _, tt := range []struct {
		name string
		call func() error
		own  error
		deps []error
	}{
		{name: "unset tempo mode", call: func() error { _, err := ticks.NewExportOptions(rounding.Mode(0)); return err }, own: ticks.ErrInvalidArgument, deps: []error{rounding.ErrUnset}},
		{name: "unknown tempo mode", call: func() error { _, err := ticks.NewExportOptions(rounding.Mode(42)); return err }, own: ticks.ErrInvalidArgument, deps: []error{rounding.ErrUnknownMode}},
		{name: "unset position mode", call: func() error { _, err := o.WithPositionMode(rounding.Mode(0)); return err }, own: ticks.ErrInvalidArgument, deps: []error{rounding.ErrUnset}},
		{name: "unset step", call: func() error { _, err := o.WithSteps(duration.Duration{}, ticks.StepStart); return err }, own: ticks.ErrInvalidArgument, deps: []error{duration.ErrUnset}},
		{name: "invalid step method", call: func() error { _, err := o.WithSteps(dur(1, 4), ticks.StepMethod(0)); return err }, own: ticks.ErrInvalidArgument},
		{name: "unset stretch length", call: func() error { _, err := o.WithHoldsStretched(duration.Duration{}); return err }, own: ticks.ErrInvalidArgument, deps: []error{duration.ErrUnset}},
		{name: "unset receiver mode", call: func() error { _, err := unset.TempoMode(); return err }, own: ticks.ErrUnset},
		{name: "unset receiver steps", call: func() error { _, err := unset.WithSteps(dur(1, 4), ticks.StepStart); return err }, own: ticks.ErrUnset},
		{name: "unset receiver holds", call: func() error { _, err := unset.WithHoldsSkipped(); return err }, own: ticks.ErrUnset},
		{name: "unset receiver clock", call: func() error { _, err := unset.WithClockSteps(true); return err }, own: ticks.ErrUnset},
		{name: "unset receiver before argument", call: func() error { _, err := unset.WithSteps(duration.Duration{}, 0); return err }, own: ticks.ErrUnset},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, tt.call(), tt.own, tt.deps...)
		})
	}
	_, err := o.WithSteps(dur(1, 4), ticks.StepMethod(0))
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { o.MustWithSteps(dur(1, 4), ticks.StepMethod(0)) })
	_, err = ticks.NewExportOptions(rounding.Mode(0))
	requireMustPanic(t, err, ticks.ErrInvalidArgument, func() { ticks.MustNewExportOptions(rounding.Mode(0)) })
	assert.False(t, unset.IsSet())
	assert.Equal(t, "ExportOptions(unset)", unset.String())
	assert.True(t, unset.Equal(ticks.ExportOptions{}))
	assert.False(t, unset.Equal(o))
}

func requireExport(t *testing.T, m tempomap.Map, r ticks.Resolution, o ticks.ExportOptions, grid ...measure.Grid) (ticks.MIDITempoMap, []tempomap.Extension) {
	t.Helper()
	mm, skipped, err := ticks.MIDITempoMapOf(m, r, o, grid...)
	require.NoError(t, err)
	assert.True(t, mm.MustResolution().Equal(r))
	assert.Equal(t, 500000, mm.MustInitial().MustMicroseconds())
	changes := mm.MustChanges()
	require.NotEmpty(t, changes)
	assert.Equal(t, 0, changes[0].Tick)
	mustMap, mustSkipped := ticks.MustMIDITempoMapOf(m, r, o, grid...)
	assert.True(t, mustMap.Equal(mm))
	assert.Equal(t, len(skipped), len(mustSkipped))
	return mm, skipped
}

func changeTexts(ticksAndTempos ...int) []string {
	out := make([]string, 0, len(ticksAndTempos)/2)
	for i := 0; i < len(ticksAndTempos); i += 2 {
		out = append(out, fmt.Sprintf("%d:%d", ticksAndTempos[i], ticksAndTempos[i+1]))
	}
	return out
}

// TestMIDITempoMapOfConstant checks О-18, П-9, ТК-11, КП-37, КП-41 and 4.10.
func TestMIDITempoMapOfConstant(t *testing.T) {
	t.Parallel()
	r := res(480)
	o := exportOptions(t, rounding.HalfEven)
	mm, skipped := requireExport(t, constantMap(120), r, o)
	assert.Nil(t, skipped)
	assert.Equal(t, changeTexts(0, 500000), tempoChanges(t, mm))

	// a tempo and jumps are written by their computed tempo.
	jumps := constantMap(120).
		MustWithChange(tempomap.MustNewJump(at(1, 1), quarterTempo(60, 1))).
		MustWithChange(tempomap.MustNewATempo(at(2, 1)))
	mm, _ = requireExport(t, jumps, r, o)
	assert.Equal(t, changeTexts(0, 500000, 1920, 1000000, 3840, 500000), tempoChanges(t, mm))

	// The tempo is rounded once by the tempo mode.
	dotted := tempomap.MustNew(tempo.MustParse("4.=60"))
	mm, _ = requireExport(t, dotted, r, o)
	assert.Equal(t, changeTexts(0, 666667), tempoChanges(t, mm))
	mm, _ = requireExport(t, dotted, r, exportOptions(t, rounding.Down))
	assert.Equal(t, changeTexts(0, 666666), tempoChanges(t, mm))

	// The written quarter of the resolution: quarter = 120 is 1000000 per half.
	mm, _ = requireExport(t, constantMap(120), resQ(480, 16), o)
	assert.Equal(t, changeTexts(0, 1000000), tempoChanges(t, mm))

	// A gradual change with equal speeds is constant and needs no steps.
	flat := constantMap(120).
		MustWithChange(tempomap.MustNewGradual(at(0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarterTempo(120, 1)))
	mm, _ = requireExport(t, flat, r, o)
	assert.Equal(t, changeTexts(0, 500000, 1920, 500000), tempoChanges(t, mm))

	// КП-41: a change at position 1/7 with D = 96.
	seventh := constantMap(120).MustWithChange(tempomap.MustNewMain(at(1, 7), quarterTempo(60, 1)))
	_, _, err := ticks.MIDITempoMapOf(seventh, res(96), o)
	requireTicksError(t, err, ticks.ErrNotRepresentable)
	requireMustPanic(t, err, ticks.ErrNotRepresentable, func() { ticks.MustMIDITempoMapOf(seventh, res(96), o) })
	mm, _ = requireExport(t, seventh, res(96), o.MustWithPositionMode(rounding.HalfUp))
	assert.Equal(t, changeTexts(0, 500000, 55, 1000000), tempoChanges(t, mm))
	mm, _ = requireExport(t, seventh, res(96), o.MustWithPositionMode(rounding.Down))
	assert.Equal(t, changeTexts(0, 500000, 54, 1000000), tempoChanges(t, mm))
	// Positions that coincide after rounding give several changes on a tick, in position order.
	nearby := constantMap(120).
		MustWithChange(tempomap.MustNewMain(at(1, 400), quarterTempo(60, 1))).
		MustWithChange(tempomap.MustNewMain(at(1, 300), quarterTempo(90, 1)))
	mm, _ = requireExport(t, nearby, res(1), o.MustWithPositionMode(rounding.Down))
	assert.Equal(t, changeTexts(0, 500000, 0, 1000000, 0, 666667), tempoChanges(t, mm))
	assert.Equal(t, 666667, mm.MustTempoAt(0).MustMicroseconds())

	// КП-37: a measured map with its grid.
	grid := grid44()
	measured := tempomap.MustNewMeasured(grid, quarterTempo(120, 1)).
		MustWithChange(tempomap.MustNewMain(tempomap.MustAtBar(addr(3, 0, 1)), quarterTempo(90, 1)), grid)
	mm, _ = requireExport(t, measured, r, o, grid)
	assert.Equal(t, changeTexts(0, 500000, 3840, 666667), tempoChanges(t, mm))
}

// TestMIDITempoMapOfSteps checks П-10, ТК-11, ТК-13, КП-15, КП-47 and ИТ-11.
func TestMIDITempoMapOfSteps(t *testing.T) {
	t.Parallel()
	r := res(480)
	rit := ritardandoMap()
	for _, tt := range []struct {
		name   string
		method ticks.StepMethod
		mode   rounding.Mode
		want   []string
	}{
		{name: "KP-15 start", method: ticks.StepStart, mode: rounding.HalfUp, want: changeTexts(0, 500000, 480, 521739, 960, 545455, 1440, 571429, 1920, 600000)},
		{name: "KP-15 end", method: ticks.StepEnd, mode: rounding.HalfUp, want: changeTexts(0, 521739, 480, 545455, 960, 571429, 1440, 600000, 1920, 600000)},
		{name: "KP-15 equal time", method: ticks.StepEqualTime, mode: rounding.HalfUp, want: changeTexts(0, 510715, 480, 533421, 960, 558240, 1440, 585482, 1920, 600000)},
		{name: "start down", method: ticks.StepStart, mode: rounding.Down, want: changeTexts(0, 500000, 480, 521739, 960, 545454, 1440, 571428, 1920, 600000)},
		{name: "start up", method: ticks.StepStart, mode: rounding.Up, want: changeTexts(0, 500000, 480, 521740, 960, 545455, 1440, 571429, 1920, 600000)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o := exportOptions(t, tt.mode).MustWithSteps(dur(1, 4), tt.method)
			mm, skipped := requireExport(t, rit, r, o)
			assert.Nil(t, skipped)
			assert.Equal(t, tt.want, tempoChanges(t, mm))
		})
	}
	for _, mode := range allModes {
		t.Run("equal time "+mode.String(), func(t *testing.T) {
			t.Parallel()
			o := exportOptions(t, mode).MustWithSteps(dur(1, 4), ticks.StepEqualTime)
			mm, _ := requireExport(t, rit, r, o)
			want := changeTexts(
				0, int(irrationalIndex(mode, 510715, false)),
				480, int(irrationalIndex(mode, 533421, false)),
				960, int(irrationalIndex(mode, 558240, false)),
				1440, int(irrationalIndex(mode, 585481, true)),
				1920, 600000,
			)
			assert.Equal(t, want, tempoChanges(t, mm))
			// The total time agrees with the musex map within the rounding of microseconds.
			total := mm.MustSeconds(1920).MustFraction()
			exact := rit.MustStart(at(1, 1), 12, rounding.HalfUp).MustValue().MustFraction()
			diff := total.MustSub(exact).MustAbs()
			assert.True(t, diff.Less(frac(4, 1000000)), "difference %s", diff)
		})
	}

	// The last step is cut by the end of the change: step 3/8 gives steps at 0, 3/8, 3/4.
	o := exportOptions(t, rounding.HalfUp).MustWithSteps(dur(3, 8), ticks.StepStart)
	mm, _ := requireExport(t, rit, r, o)
	assert.Equal(t, changeTexts(0, 500000, 720, 533333, 1440, 571429, 1920, 600000), tempoChanges(t, mm))

	// КП-47: steps on MIDI clocks.
	clock := exportOptions(t, rounding.HalfUp).MustWithSteps(dur(1, 4), ticks.StepStart).MustWithClockSteps(true)
	mm, _ = requireExport(t, rit, r, clock)
	assert.Equal(t, changeTexts(0, 500000, 480, 521739, 960, 545455, 1440, 571429, 1920, 600000), tempoChanges(t, mm))
	offClock := constantMap(120).
		MustWithChange(tempomap.MustNewGradual(at(1, 192), tempomap.LinearSpeed)).
		MustWithChange(tempomap.MustNewMain(at(193, 192), quarterTempo(100, 1)))
	_, _, err := ticks.MIDITempoMapOf(offClock, r, clock)
	requireTicksError(t, err, ticks.ErrNotRepresentable)
	_, _, err = ticks.MIDITempoMapOf(rit, res(100), clock)
	requireTicksError(t, err, ticks.ErrNotRepresentable)
	// Without the clock option the same map is written.
	mm, _ = requireExport(t, offClock, r, exportOptions(t, rounding.HalfUp).MustWithSteps(dur(1, 4), ticks.StepStart))
	assert.Equal(t, 10, mm.MustChanges()[1].Tick)
}

// TestMIDITempoMapOfHolds checks П-11, ТК-11, КП-16, КП-25, КП-46 and 5.2.
func TestMIDITempoMapOfHolds(t *testing.T) {
	t.Parallel()
	r := res(480)
	o := exportOptions(t, rounding.HalfUp)
	hold := holdMap(pos(1, 1), sec(1, 1))

	// КП-16.
	mm, skipped := requireExport(t, hold, r, o.MustWithHoldsStretched(dur(1, 4)))
	assert.Nil(t, skipped)
	assert.Equal(t, changeTexts(0, 1000000, 1440, 2000000, 1920, 1000000), tempoChanges(t, mm))
	assertSeconds(t, sec(5, 1), mm.MustSeconds(1920))

	// КП-25.
	mm, skipped = requireExport(t, hold, r, o.MustWithHoldsSkipped())
	assert.Equal(t, changeTexts(0, 1000000), tempoChanges(t, mm))
	require.Len(t, skipped, 1)
	assert.True(t, skipped[0].Equal(tempomap.MustNewExtension(at(1, 1), sec(1, 1))), "skipped %s", skipped[0])
	_, _, err := ticks.MIDITempoMapOf(hold, r, o)
	requireTicksError(t, err, ticks.ErrNotConvertible)
	atZero := holdMap(duration.Origin(), sec(1, 1))
	_, _, err = ticks.MIDITempoMapOf(atZero, r, o.MustWithHoldsStretched(dur(1, 4)))
	requireTicksError(t, err, ticks.ErrNotConvertible)
	requireMustPanic(t, err, ticks.ErrNotConvertible, func() { ticks.MustMIDITempoMapOf(atZero, r, o.MustWithHoldsStretched(dur(1, 4))) })

	// КП-46.
	twoHolds := hold.MustWithExtension(tempomap.MustNewExtension(at(5, 4), sec(1, 1)))
	_, _, err = ticks.MIDITempoMapOf(twoHolds, r, o.MustWithHoldsStretched(dur(1, 2)))
	requireTicksError(t, err, ticks.ErrNotConvertible)
	mm, _ = requireExport(t, twoHolds, r, o.MustWithHoldsStretched(dur(1, 4)))
	assert.Equal(t, changeTexts(0, 1000000, 1440, 2000000, 1920, 2000000, 2400, 1000000), tempoChanges(t, mm))
	early := holdMap(pos(1, 8), sec(1, 1))
	_, _, err = ticks.MIDITempoMapOf(early, r, o.MustWithHoldsStretched(dur(1, 4)))
	requireTicksError(t, err, ticks.ErrNotConvertible)

	// Skipped holds keep the order of the map.
	mm, skipped = requireExport(t, twoHolds, r, o.MustWithHoldsSkipped())
	assert.Equal(t, changeTexts(0, 1000000), tempoChanges(t, mm))
	require.Len(t, skipped, 2)
	assert.True(t, skipped[0].Equal(tempomap.MustNewExtension(at(1, 1), sec(1, 1))))
	assert.True(t, skipped[1].Equal(tempomap.MustNewExtension(at(5, 4), sec(1, 1))))

	// КР-TK-6: zero holds are neither an error nor reported.
	zero := holdMap(pos(1, 1), tempo.ZeroSeconds())
	for _, opts := range []ticks.ExportOptions{o, o.MustWithHoldsSkipped(), o.MustWithHoldsStretched(dur(1, 4))} {
		mm, skipped = requireExport(t, zero, r, opts)
		assert.Nil(t, skipped)
		assert.Equal(t, changeTexts(0, 1000000), tempoChanges(t, mm))
	}

	// A measured map reports holds with a positional address.
	grid := grid44()
	measured := tempomap.MustNewMeasured(grid, quarterTempo(60, 1)).
		MustWithExtension(tempomap.MustNewExtension(tempomap.MustAtBar(addr(2, 0, 1)), sec(1, 1)), grid)
	_, skipped = requireExport(t, measured, r, o.MustWithHoldsSkipped(), grid)
	require.Len(t, skipped, 1)
	assert.True(t, skipped[0].Equal(tempomap.MustNewExtension(at(1, 1), sec(1, 1))))
}

// TestMIDITempoMapOfStretchOnGradualSteps checks 5.2 (the fractional-linear
// target) and 5.3 through the public API: the fermata stretch of a step of a
// gradual change gives the correctly rounded tempo without a failure.
func TestMIDITempoMapOfStretchOnGradualSteps(t *testing.T) {
	t.Parallel()
	r := res(480)
	rit := ritardandoMap().MustWithExtension(tempomap.MustNewExtension(at(1, 1), sec(1, 1)))
	for _, mode := range allModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			base := exportOptions(t, mode).MustWithHoldsStretched(dur(1, 4))
			mm, skipped := requireExport(t, rit, r, base.MustWithSteps(dur(1, 4), ticks.StepStart))
			assert.Nil(t, skipped)
			assert.Equal(t, changeTexts(
				0, 500000,
				480, int(irrationalIndex(mode, 521739, false)),
				960, int(irrationalIndex(mode, 545454, true)),
				1440, int(irrationalIndex(mode, 1571428, true)),
				1920, 600000,
			), tempoChanges(t, mm))

			mm, _ = requireExport(t, rit, r, base.MustWithSteps(dur(1, 4), ticks.StepEnd))
			assert.Equal(t, changeTexts(
				0, int(irrationalIndex(mode, 521739, false)),
				480, int(irrationalIndex(mode, 545454, true)),
				960, int(irrationalIndex(mode, 571428, true)),
				1440, 1600000,
				1920, 600000,
			), tempoChanges(t, mm))

			mm, _ = requireExport(t, rit, r, base.MustWithSteps(dur(1, 4), ticks.StepEqualTime))
			assert.Equal(t, changeTexts(
				0, int(irrationalIndex(mode, 510715, false)),
				480, int(irrationalIndex(mode, 533421, false)),
				960, int(irrationalIndex(mode, 558240, false)),
				1440, int(irrationalIndex(mode, 1585481, true)),
				1920, 600000,
			), tempoChanges(t, mm))
		})
	}
}

// TestMIDITempoMapOfTempoLimit checks П-4 and 5.2 for exported tempos.
func TestMIDITempoMapOfTempoLimit(t *testing.T) {
	t.Parallel()
	r := res(480)
	// A stretch that makes the tempo 17000000.
	huge := holdMap(pos(1, 1), sec(16, 1))
	_, _, err := ticks.MIDITempoMapOf(huge, r, exportOptions(t, rounding.HalfUp).MustWithHoldsStretched(dur(1, 4)))
	requireTicksError(t, err, ticks.ErrFormatLimit)

	// The exact tempo of the piece is 16777215 and the stretch adds exactly 1/2:
	// S = 16.777215 s and E = 1/2000000 s.
	edge := tempomap.MustNew(quarterTempo(60000000, 16777215)).
		MustWithExtension(tempomap.MustNewExtension(at(1, 2), sec(1, 2000000)))
	for _, mode := range allModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			o := exportOptions(t, mode).MustWithHoldsStretched(dur(1, 4))
			stretched := tieIndex(mode, 16777215, false)
			mm, _, err := ticks.MIDITempoMapOf(edge, r, o)
			if stretched > 16777215 {
				requireTicksError(t, err, ticks.ErrFormatLimit)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, changeTexts(0, 16777215, 480, 16777215, 960, 16777215), tempoChanges(t, mm))
		})
	}

	// An exact tempo slower than Set Tempo.
	_, _, err = ticks.MIDITempoMapOf(constantMap(3), r, exportOptions(t, rounding.HalfUp))
	requireTicksError(t, err, ticks.ErrFormatLimit)

	// An irrational step tempo above the limit: quarter = 4 slowing to quarter = 3.
	slow := tempomap.MustNew(quarterTempo(4, 1)).
		MustWithChange(tempomap.MustNewGradual(at(0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarterTempo(3, 1)))
	for _, mode := range allModes {
		_, _, err = ticks.MIDITempoMapOf(slow, r, exportOptions(t, mode).MustWithSteps(dur(1, 1), ticks.StepEqualTime))
		requireTicksError(t, err, ticks.ErrFormatLimit)
	}

	// The last StepStart piece has m = 4000000/7 µs and spans one
	// quarter. A hold of 20 s adds 20000000 µs, exceeding the format limit.
	ritHold := ritardandoMap().MustWithExtension(tempomap.MustNewExtension(at(1, 1), sec(20, 1)))
	for _, mode := range allModes {
		o := exportOptions(t, mode).MustWithSteps(dur(1, 4), ticks.StepStart).MustWithHoldsStretched(dur(1, 4))
		_, _, err = ticks.MIDITempoMapOf(ritHold, r, o)
		requireTicksError(t, err, ticks.ErrFormatLimit)
	}

	// 4.16, E-5: the index of the step call for a tempo does not fit int.
	// Quarter = 10⁻¹³ slowing to quarter = 5·10⁻¹⁴ over one whole note:
	// speeds 2.5·10⁻¹⁴ and 1.25·10⁻¹⁴ whole notes per minute, the step of
	// a whole note lasts T = 60·ln2/(1.25·10⁻¹⁴) ≈ 3.327·10¹⁵ s, and its
	// equal-time tempo 10⁶·(1/4)·T ≈ 8.32·10²⁰ µs exceeds math.MaxInt.
	glacial := tempomap.MustNew(quarterTempo(1, 10000000000000)).
		MustWithChange(tempomap.MustNewGradual(at(0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarterTempo(1, 20000000000000)))
	for _, mode := range allModes {
		_, _, err = ticks.MIDITempoMapOf(glacial, r, exportOptions(t, mode).MustWithSteps(dur(1, 1), ticks.StepEqualTime))
		requireTicksError(t, err, ticks.ErrFormatLimit, tempomap.ErrOutOfRange, fraction.ErrNotRepresentable)
	}
}

// TestMIDITempoMapOfErrors checks КП-24 and the order of 4.16.
func TestMIDITempoMapOfErrors(t *testing.T) {
	t.Parallel()
	r := res(480)
	o := exportOptions(t, rounding.HalfUp)
	grid := grid44()
	measured := tempomap.MustNewMeasured(grid, quarterTempo(120, 1))
	for _, tt := range []struct {
		name string
		m    tempomap.Map
		r    ticks.Resolution
		o    ticks.ExportOptions
		grid []measure.Grid
		own  error
		deps []error
		not  []error
	}{
		{name: "KP-24 gradual change without steps", m: ritardandoMap(), r: r, o: o, own: ticks.ErrNotConvertible},
		{name: "KP-24 unfinished gradual change", m: unfinishedMap(), r: r, o: o.MustWithSteps(dur(1, 4), ticks.StepStart), own: ticks.ErrNotConvertible, deps: []error{tempomap.ErrUnfinished}},
		{name: "unset map", m: tempomap.Map{}, r: r, o: o, own: ticks.ErrInvalidArgument, deps: []error{tempomap.ErrUnset}},
		{name: "unset map before unset resolution", m: tempomap.Map{}, r: ticks.Resolution{}, o: ticks.ExportOptions{}, own: ticks.ErrInvalidArgument, deps: []error{tempomap.ErrUnset}},
		{name: "unset resolution", m: constantMap(120), r: ticks.Resolution{}, o: o, own: ticks.ErrUnset},
		{name: "unset options", m: constantMap(120), r: r, o: ticks.ExportOptions{}, own: ticks.ErrUnset},
		{name: "measured map without a grid", m: measured, r: r, o: o, own: ticks.ErrInvalidArgument},
		{name: "positional map with a grid", m: constantMap(120), r: r, o: o, grid: []measure.Grid{grid}, own: ticks.ErrInvalidArgument},
		{name: "unset grid", m: measured, r: r, o: o, grid: []measure.Grid{{}}, own: ticks.ErrInvalidArgument, deps: []error{measure.ErrUnset}},
		{name: "grid is checked before the unfinished tail", m: unfinishedMap(), r: r, o: o, grid: []measure.Grid{grid}, own: ticks.ErrInvalidArgument, not: []error{tempomap.ErrUnfinished}},
		{name: "unfinished tail before missing steps", m: ritardandoMap().MustWithChange(tempomap.MustNewGradual(at(2, 1), tempomap.LinearSpeed)), r: r, o: o, own: ticks.ErrNotConvertible, deps: []error{tempomap.ErrUnfinished}},
		{name: "missing steps before an unrepresentable position", m: constantMap(120).MustWithChange(tempomap.MustNewMain(at(1, 7), quarterTempo(60, 1))).MustWithChange(tempomap.MustNewGradual(at(1, 1), tempomap.LinearSpeed)).MustWithChange(tempomap.MustNewMain(at(2, 1), quarterTempo(90, 1))), r: res(96), o: o, own: ticks.ErrNotConvertible},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mm, skipped, err := ticks.MIDITempoMapOf(tt.m, tt.r, tt.o, tt.grid...)
			requireTicksError(t, err, tt.own, tt.deps...)
			for _, n := range tt.not {
				assert.NotErrorIs(t, err, n)
			}
			assert.False(t, mm.IsSet())
			assert.Nil(t, skipped)
		})
	}
}

// TestMIDITempoMapToMusex checks О-17, О-29, П-8, ТК-10, ТК-18, КП-12, КП-37, КП-40.
func TestMIDITempoMapToMusex(t *testing.T) {
	t.Parallel()
	same := ticks.MustNewMIDITempoMap(res(96), tchange(0, 500000), tchange(0, 600000))
	got, err := same.TempoMap(writtenQuarter())
	require.NoError(t, err)
	assertEqualValue(t, constantMap(100), got)
	assertEqualValue(t, constantMap(100), same.MustTempoMap(writtenQuarter()))

	second := ticks.MustNewMIDITempoMap(res(96), tchange(0, 1000000))
	assertEqualValue(t, tempomap.MustNew(tempo.MustParse("4.=40")), second.MustTempoMap(writtenDottedQuarter()))
	assertEqualValue(t, constantMap(60), second.MustTempoMap(writtenQuarter()))

	// П-6: before the first change the initial tempo is in effect.
	late := ticks.MustNewMIDITempoMap(res(96), tchange(96, 1000000), tchange(192, 600000), tchange(192, 700000))
	want := constantMap(120).
		MustWithChange(tempomap.MustNewMain(at(1, 4), quarterTempo(60, 1))).
		MustWithChange(tempomap.MustNewMain(at(1, 2), quarterTempo(600, 7)))
	assertEqualValue(t, want, late.MustTempoMap(writtenQuarter()))
	assertEqualValue(t, constantMap(60), ticks.MustNewMIDITempoMap(res(96)).MustWithInitial(mtempo(1000000)).MustTempoMap(writtenQuarter()))
	// Repeated equal changes stay new main tempos.
	repeated := ticks.MustNewMIDITempoMap(res(96), tchange(0, 500000), tchange(96, 500000))
	assertEqualValue(t, constantMap(120).MustWithChange(tempomap.MustNewMain(at(1, 4), quarterTempo(120, 1))), repeated.MustTempoMap(writtenQuarter()))

	// The written quarter of the map.
	half := ticks.MustNewMIDITempoMap(resQ(96, 16), tchange(0, 500000), tchange(96, 1000000))
	assertEqualValue(t, constantMap(240).MustWithChange(tempomap.MustNewMain(at(1, 2), quarterTempo(120, 1))), half.MustTempoMap(writtenQuarter()))

	// КП-37.
	grid := grid44()
	measured := ticks.MustNewMIDITempoMap(res(480), tchange(0, 500000), tchange(3840, 666667))
	m, err := measured.MeasuredTempoMap(grid, writtenQuarter())
	require.NoError(t, err)
	wantMeasured := tempomap.MustNewMeasured(grid, quarterTempo(120, 1)).
		MustWithChange(tempomap.MustNewMain(tempomap.MustAtBar(addr(3, 0, 1)), tempo.MustNewQuarter(frac(60000000, 666667))), grid)
	assertEqualValue(t, wantMeasured, m)
	assertEqualValue(t, wantMeasured, measured.MustMeasuredTempoMap(grid, writtenQuarter()))
	// Round trip of КП-37 through the musex map.
	back, _ := ticks.MustMIDITempoMapOf(m, res(480), exportOptions(t, rounding.HalfUp), grid)
	assert.True(t, back.Equal(measured))

	finite := measure.MustNew(meter.FourFour()).MustWithBarCount(2)
	beyond := ticks.MustNewMIDITempoMap(res(480), tchange(0, 500000), tchange(3840, 600000))
	_, err = beyond.MeasuredTempoMap(finite, writtenQuarter())
	requireTicksError(t, err, ticks.ErrNotConvertible, tempomap.ErrInvalidAddress, measure.ErrOutsideGrid)
	requireMustPanic(t, err, ticks.ErrNotConvertible, func() { beyond.MustMeasuredTempoMap(finite, writtenQuarter()) })

	for _, tt := range []struct {
		name string
		call func() error
		own  error
		deps []error
	}{
		{name: "unset unit", call: func() error { _, err := same.TempoMap(duration.Written{}); return err }, own: ticks.ErrInvalidArgument, deps: []error{duration.ErrUnset}},
		{name: "unset grid", call: func() error { _, err := same.MeasuredTempoMap(measure.Grid{}, writtenQuarter()); return err }, own: ticks.ErrInvalidArgument, deps: []error{measure.ErrUnset}},
		{name: "unset measured unit", call: func() error { _, err := same.MeasuredTempoMap(grid, duration.Written{}); return err }, own: ticks.ErrInvalidArgument, deps: []error{duration.ErrUnset}},
		{name: "unset map", call: func() error { _, err := ticks.MIDITempoMap{}.TempoMap(writtenQuarter()); return err }, own: ticks.ErrUnset},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireTicksError(t, tt.call(), tt.own, tt.deps...)
		})
	}
}

// TestMIDITempoMapStretchInheritsBaseStep checks KP-50: cutting a step
// must preserve its base tempo before stretching, for all three methods.
func TestMIDITempoMapStretchInheritsBaseStep(t *testing.T) {
	t.Parallel()
	m := ritardandoMap().MustWithExtension(tempomap.MustNewExtension(at(1, 1), sec(1, 1)))
	for _, tt := range []struct {
		name                     string
		method                   ticks.StepMethod
		first, second, stretched int
		seconds                  tempo.Seconds
	}{
		{name: "start", method: ticks.StepStart, first: 500000, second: 545455, stretched: 1545455, seconds: sec(309091, 100000)},
		{name: "end", method: ticks.StepEnd, first: 545455, second: 600000, stretched: 1600000, seconds: sec(329091, 100000)},
		{name: "equal time", method: ticks.StepEqualTime, first: 522068, second: 571861, stretched: 1571861, seconds: sec(1593929, 500000)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o := exportOptions(t, rounding.HalfUp).MustWithSteps(dur(1, 2), tt.method)
			base, _ := requireExport(t, ritardandoMap(), res(480), o)
			got, skipped := requireExport(t, m, res(480), o.MustWithHoldsStretched(dur(1, 4)))
			assert.Nil(t, skipped)
			assert.Equal(t, changeTexts(0, tt.first, 960, tt.second, 1440, tt.stretched, 1920, 600000), tempoChanges(t, got))
			assertSeconds(t, tt.seconds, got.MustSeconds(1920))
			assertSeconds(t, sec(1, 1), got.MustSeconds(1920).MustSub(base.MustSeconds(1920)))
			assertSeconds(t, sec(1, 1), got.MustSeconds(2400).MustSub(base.MustSeconds(2400)))
		})
	}
}

// TestMIDITempoMapStretchUsesFinalTickGrid checks KP-51 and the order of
// failures in 4.16: positions, then collapsed stretches, then tempo limits.
func TestMIDITempoMapStretchUsesFinalTickGrid(t *testing.T) {
	t.Parallel()
	m := holdMap(pos(5, 16), sec(1, 1))
	o := exportOptions(t, rounding.HalfUp).MustWithPositionMode(rounding.Down).MustWithHoldsStretched(dur(1, 8))
	got, skipped := requireExport(t, m, res(1), o)
	assert.Nil(t, skipped)
	// The split at 3/16 shares tick zero with the initial event.
	assert.Equal(t, changeTexts(0, 2000000, 0, 2000000, 1, 1000000), tempoChanges(t, got))
	assertSeconds(t, sec(2, 1), got.MustSeconds(1))
	assertSeconds(t, sec(3, 1), got.MustSeconds(2))

	for _, tt := range []struct {
		name string
		m    tempomap.Map
		o    ticks.ExportOptions
		own  error
	}{
		{name: "collapsed stretch", m: holdMap(pos(1, 8), sec(1, 1)), o: o.MustWithHoldsStretched(dur(1, 16)), own: ticks.ErrNotConvertible},
		{name: "positions precede collapse", m: holdMap(pos(1, 8), sec(1, 1)), o: o.MustWithExactPositions().MustWithHoldsStretched(dur(1, 16)), own: ticks.ErrNotRepresentable},
		{name: "collapse precedes overflowing tempo", m: tempomap.MustNew(quarterTempo(3, 1)).MustWithExtension(tempomap.MustNewExtension(at(1, 8), sec(1, 1))), o: o.MustWithHoldsStretched(dur(1, 16)), own: ticks.ErrNotConvertible},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, skipped, err := ticks.MIDITempoMapOf(tt.m, res(1), tt.o)
			requireTicksError(t, err, tt.own)
			assert.False(t, got.IsSet())
			assert.Nil(t, skipped)
			requireMustPanic(t, err, tt.own, func() { ticks.MustMIDITempoMapOf(tt.m, res(1), tt.o) })
		})
	}
}

// TestMIDITempoMapStretchIgnoresZeroTickFragments checks 4.10: all
// coincident changes survive, but only effective positive tick lengths
// contribute to S. Every change at a tick inside [u,v) is stretched.
func TestMIDITempoMapStretchIgnoresZeroTickFragments(t *testing.T) {
	t.Parallel()
	m := constantMap(60).
		MustWithChange(tempomap.MustNewMain(at(1, 8), quarterTempo(30, 1))).
		MustWithChange(tempomap.MustNewMain(at(3, 8), quarterTempo(20, 1))).
		MustWithExtension(tempomap.MustNewExtension(at(9, 16), sec(1, 1)))
	o := exportOptions(t, rounding.HalfEven).MustWithPositionMode(rounding.Down).MustWithHoldsStretched(dur(1, 2))
	got, skipped := requireExport(t, m, res(1), o)
	assert.Nil(t, skipped)
	// Positions 0,1/16,1/8 share tick zero. The effective base fragments
	// have tempos 2000000 and 3000000, each for one tick: S=5 s.
	// The multiplier 6/5 also applies to both zero-length 1000000 changes.
	assert.Equal(t, changeTexts(0, 1200000, 0, 1200000, 0, 2400000, 1, 3600000, 2, 3000000), tempoChanges(t, got))
	assertSeconds(t, sec(12, 5), got.MustSeconds(1))
	assertSeconds(t, sec(6, 1), got.MustSeconds(2))
	assertSeconds(t, sec(9, 1), got.MustSeconds(3))
}

// TestMIDITempoMapStretchAcrossBaseSteps checks a sum of effective base
// fragments, including the two-form EqualTime case. At step 1/2, cutting
// at 1/4 leaves one quarter of the first step and half of the second:
// S=(m0+2*m1)/10^6, and the final tempos are mi*(S+1)/S.
// Independent exact/85-digit Decimal expectations are:
// start: m0=500000, final=(5700000/7,68400000/77);
// end: m0=6000000/11, final=(9437500/11,943750);
// equal time: m0=6000000*ln(12/11), m1=6000000*ln(11/10),
// final=(835473.991613169695...,915158.213988253610...).
func TestMIDITempoMapStretchAcrossBaseSteps(t *testing.T) {
	t.Parallel()
	m := ritardandoMap().MustWithExtension(tempomap.MustNewExtension(at(1, 1), sec(1, 1)))
	for _, tt := range []struct {
		name                                   string
		method                                 ticks.StepMethod
		base, first, second                    int64
		baseUp, firstUp, secondUp, secondExact bool
	}{
		{name: "start", method: ticks.StepStart, base: 500000, first: 814285, second: 888311, firstUp: true, secondUp: true},
		{name: "end", method: ticks.StepEnd, base: 545454, first: 857954, second: 943750, baseUp: true, firstUp: true, secondExact: true},
		{name: "equal time", method: ticks.StepEqualTime, base: 522068, first: 835473, second: 915158, firstUp: true},
	} {
		for _, mode := range allModes {
			t.Run(tt.name+"/"+mode.String(), func(t *testing.T) {
				t.Parallel()
				base := tt.base
				if tt.method != ticks.StepStart {
					base = irrationalIndex(mode, tt.base, tt.baseUp)
				}
				second := tt.second
				if !tt.secondExact {
					second = irrationalIndex(mode, tt.second, tt.secondUp)
				}
				o := exportOptions(t, mode).MustWithSteps(dur(1, 2), tt.method).MustWithHoldsStretched(dur(3, 4))
				got, skipped := requireExport(t, m, res(480), o)
				assert.Nil(t, skipped)
				assert.Equal(t, changeTexts(0, int(base), 480, int(irrationalIndex(mode, tt.first, tt.firstUp)), 960, int(second), 1920, 600000), tempoChanges(t, got))
			})
		}
	}
}

// TestMIDITempoMapStretchRationalFragmentWithWeightedLogTime checks 5.2:
// a rational constant fragment is stretched using an irrational S from
// multiple EqualTime steps with unequal weights. The source step tempos
// are A=6000000*ln(12/11) and B=6000000*ln(11/10). The stretch [3/8,5/4)
// retains 1/4 of the first source interval, all of the second, and a
// quarter-long constant tail: S=(A/2+2*B+600000)/10^6.
// Independent 100-digit Decimal evaluation gives S=2.004756288620787619...
// and stretched tempos A'=782483.088867662403...,
// B'=857113.247417161054..., C'=899288.249352634308... microseconds.
// The rational tail must use exact fractional-linear comparisons; an
// ErrOutOfRange from a finite interval fallback is not an allowed result.
func TestMIDITempoMapStretchRationalFragmentWithWeightedLogTime(t *testing.T) {
	t.Parallel()
	m := ritardandoMap().MustWithExtension(tempomap.MustNewExtension(at(5, 4), sec(1, 1)))
	for _, mode := range allModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			o := exportOptions(t, mode).MustWithSteps(dur(1, 2), ticks.StepEqualTime).MustWithHoldsStretched(dur(7, 8))
			got, skipped := requireExport(t, m, res(480), o)
			assert.Nil(t, skipped)
			assert.Equal(t, changeTexts(
				0, int(irrationalIndex(mode, 522068, false)),
				720, int(irrationalIndex(mode, 782483, false)),
				960, int(irrationalIndex(mode, 857113, false)),
				1920, int(irrationalIndex(mode, 899288, false)),
				2400, 600000,
			), tempoChanges(t, got))
		})
	}
}

// TestMIDITempoMapStretchWeightedLogTimeAtFormatLimit checks the same
// fractional-linear path just above the Set Tempo limit M=16777215.
// For this slower ramp A=684170.176404998524... and
// B=2936915.488846336666...; S=(A/2+2*B+M)/10^6
// =22.993131065895172595... s (independent 100-digit Decimal calculation).
// E=10^-7 s gives f(S)=M+0.072966204349980843..., so 0<M*E/S<1/2.
// Thus f(S) is strictly between M and M+1/2: only Up/AwayFromZero
// select M+1 (E-5); every other mode selects M, without an E-8 fallback.
func TestMIDITempoMapStretchWeightedLogTimeAtFormatLimit(t *testing.T) {
	t.Parallel()
	m := tempomap.MustNew(quarterTempo(120, 1)).
		MustWithChange(tempomap.MustNewGradual(at(0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarterTempo(60000000, 16777215))).
		MustWithExtension(tempomap.MustNewExtension(at(5, 4), sec(1, 10000000)))
	for _, mode := range allModes {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()
			o := exportOptions(t, mode).MustWithSteps(dur(1, 2), ticks.StepEqualTime).MustWithHoldsStretched(dur(7, 8))
			got, skipped, err := ticks.MIDITempoMapOf(m, res(480), o)
			if mode == rounding.Up || mode == rounding.AwayFromZero {
				requireTicksError(t, err, ticks.ErrFormatLimit)
				assert.False(t, got.IsSet())
				assert.Nil(t, skipped)
				requireMustPanic(t, err, ticks.ErrFormatLimit, func() { ticks.MustMIDITempoMapOf(m, res(480), o) })
				return
			}
			require.NoError(t, err)
			assert.Nil(t, skipped)
			assert.Equal(t, 16777215, got.MustTempoAt(1920).MustMicroseconds())
			assert.Equal(t, 16777215, got.MustTempoAt(2400).MustMicroseconds())
			must, _ := ticks.MustMIDITempoMapOf(m, res(480), o)
			assert.True(t, got.Equal(must))
		})
	}
}
