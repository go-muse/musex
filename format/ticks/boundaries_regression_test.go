package ticks_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/tempomap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The valid inputs p=1/2 and L=1/M produce the internal split
// (M-2)/(2*M), whose reduced denominator exceeds fraction's M. That split
// must stay exact until tick quantization (4.10, 5.2). At D=480, Down maps
// it to 959 and p to 960. S=1/480 s, E=1/1000 s, so (S+E)/S=37/25.
func TestStretchIntermediatePositionOutsideFraction(t *testing.T) {
	t.Parallel()
	length := duration.MustFromFraction(fraction.MustFromBigInt(big.NewInt(1), fraction.MaxComponentBig()))
	m := holdMap(pos(1, 2), sec(1, 1000))
	o := exportOptions(t, rounding.HalfUp).MustWithPositionMode(rounding.Down).MustWithHoldsStretched(length)
	before := m.String()
	var got ticks.MIDITempoMap
	var skipped []tempomap.Extension
	var err error
	require.NotPanics(t, func() { got, skipped, err = ticks.MIDITempoMapOf(m, res(480), o) })
	require.NoError(t, err)
	assert.Nil(t, skipped)
	assert.Equal(t, changeTexts(0, 1000000, 959, 1480000, 960, 1000000), tempoChanges(t, got))
	assertSeconds(t, sec(2001, 1000), got.MustSeconds(960))
	var must ticks.MIDITempoMap
	require.NotPanics(t, func() { must, skipped = ticks.MustMIDITempoMapOf(m, res(480), o) })
	assert.Nil(t, skipped)
	assert.True(t, must.Equal(got))
	assert.Equal(t, before, m.String())
}

// At D=1, each tick is a quarter note. A 1/2^62 bar therefore gives
// 2^60 complete bars per tick. These cases cross the machine bar-number
// boundary through a single quotient or the sum of two valid counts.
func TestMIDIMeterMapGridBarNumberOverflow(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		changes []ticks.MIDIMeterChange
	}{
		{name: "count 2^63 exceeds int", changes: []ticks.MIDIMeterChange{mchange(0, 1, 62, 24, 8), mchange(8, 4, 2, 24, 8)}},
		{name: "count 2^64 must not wrap to zero", changes: []ticks.MIDIMeterChange{mchange(0, 1, 62, 24, 8), mchange(16, 4, 2, 24, 8)}},
		{name: "sum of two representable counts", changes: []ticks.MIDIMeterChange{mchange(0, 1, 62, 24, 8), mchange(4, 1, 62, 48, 8), mchange(8, 4, 2, 24, 8)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := ticks.MustNewMIDIMeterMap(res(1), tt.changes...)
			var got measure.Grid
			var err error
			require.NotPanics(t, func() { got, err = m.Grid() })
			requireTicksError(t, err, ticks.ErrOutOfRange)
			assert.Equal(t, measure.Grid{}, got)
			requireMustPanic(t, err, ticks.ErrOutOfRange, func() { m.MustGrid() })
		})
	}
}

// Seven ticks at 1/2^62, then one tick at each size 1/2^(k+2), k=59..1,
// contain 7*2^60 + sum(2^k,k=1..59) = 2^63-2 = MaxInt-1 bars.
// The final 4/4 event thus starts exactly at MaxInt, at tick 66 (position
// 33/2). This fixture uses only 61 MIDI events and no huge enumeration.
func maxBarBoundaryChanges() ([]ticks.MIDIMeterChange, int) {
	changes := []ticks.MIDIMeterChange{mchange(0, 1, 62, 24, 8)}
	tick := 7
	for power := 59; power >= 1; power-- {
		changes = append(changes, mchange(tick, 1, power+2, 24, 8))
		tick++
	}
	return append(changes, mchange(tick, 4, 2, 24, 8)), tick
}

func TestMIDIMeterMapGridPartialBarAtMaxInt(t *testing.T) {
	t.Parallel()
	changes, tick := maxBarBoundaryChanges()
	prefix := ticks.MustNewMIDIMeterMap(res(1), changes...)
	g, err := prefix.Grid()
	require.NoError(t, err, "reaching MaxInt itself is allowed")
	start, err := g.BarStart(math.MaxInt)
	require.NoError(t, err)
	assertEqualValue(t, pos(33, 2), start)
	assertEqualValue(t, dur(1, 1), g.MustBarLength(math.MaxInt))
	// One more tick ends a quarter-long partial bar at MaxInt. The new
	// 3/4 event would start at MaxInt+1 and must fail with the range error.
	m := ticks.MustNewMIDIMeterMap(res(1), append(changes, mchange(tick+1, 3, 2, 24, 8))...)
	var got measure.Grid
	require.NotPanics(t, func() { got, err = m.Grid() })
	requireTicksError(t, err, ticks.ErrOutOfRange)
	assert.Equal(t, measure.Grid{}, got)
	requireMustPanic(t, err, ticks.ErrOutOfRange, func() { m.MustGrid() })
}

// Section 4.14 requires the reduced length denominator itself to be a
// power of two; taking max(4,3)=4 cannot make an explicit 1/3 bar MIDI.
// Both resolutions represent the next bar's start exactly, so E-4 is not
// involved. Skipping irregular bars still exports the ordinary 4/4 meter.
func TestMIDIMeterMapOfNonBinaryExplicitLength(t *testing.T) {
	t.Parallel()
	for _, d := range []int{3, 480} {
		t.Run(res(d).String(), func(t *testing.T) {
			t.Parallel()
			g := grid44().MustWithLength(1, dur(1, 3), measure.AlignStart)
			click := ticks.MustClickEvery(dur(1, 4))
			t.Run("event rejects the denominator", func(t *testing.T) {
				t.Parallel()
				got, losses, err := ticks.MIDIMeterMapOf(g, res(d), click, ticks.Strict, ticks.IrregularEvent)
				requireTicksError(t, err, ticks.ErrMeterNotMIDI)
				assert.Equal(t, ticks.MIDIMeterMap{}, got)
				assert.Nil(t, losses)
				requireMustPanic(t, err, ticks.ErrMeterNotMIDI, func() {
					ticks.MustMIDIMeterMapOf(g, res(d), click, ticks.Strict, ticks.IrregularEvent)
				})
			})
			t.Run("skip exports the ordinary meter", func(t *testing.T) {
				t.Parallel()
				got, losses, err := ticks.MIDIMeterMapOf(g, res(d), click, ticks.Strict, ticks.IrregularSkip)
				require.NoError(t, err)
				assert.Nil(t, losses)
				assert.Equal(t, []string{"0:4,2,24,8"}, meterChanges(t, got))
				must, mustLosses := ticks.MustMIDIMeterMapOf(g, res(d), click, ticks.Strict, ticks.IrregularSkip)
				assert.True(t, must.Equal(got))
				assert.Nil(t, mustLosses)
			})
		})
	}
}
