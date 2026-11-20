package ticks_test

import (
	"sync"
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/format/ticks"

	"github.com/stretchr/testify/assert"
)

type unsetValue interface {
	IsSet() bool
	String() string
	MarshalText() ([]byte, error)
	AppendText([]byte) ([]byte, error)
}

// TestUnsetValues checks 4.2: for zero values only IsSet, Equal, Compare and
// String work; every other operation fails with ErrUnset.
func TestUnsetValues(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		value     unsetValue
		text      string
		equalZero bool
		equalSet  bool
	}{
		{name: "MIDIQuarter", value: ticks.MIDIQuarter{}, text: "MIDIQuarter(unset)", equalZero: ticks.MIDIQuarter{}.Equal(ticks.MIDIQuarter{}), equalSet: ticks.MIDIQuarter{}.Equal(ticks.StandardQuarter())},
		{name: "Resolution", value: ticks.Resolution{}, text: "Resolution(unset)", equalZero: ticks.Resolution{}.Equal(ticks.Resolution{}), equalSet: ticks.Resolution{}.Equal(res(1))},
		{name: "Quantized", value: ticks.Quantized{}, text: "Quantized(unset)", equalZero: ticks.Quantized{}.Equal(ticks.Quantized{}), equalSet: ticks.Quantized{}.Equal(ticks.MustParseQuantized("0+0/1"))},
		{name: "SMPTE", value: ticks.SMPTE{}, text: "SMPTE(unset)", equalZero: ticks.SMPTE{}.Equal(ticks.SMPTE{}), equalSet: ticks.SMPTE{}.Equal(smpte(-25, 40))},
		{name: "MIDITempo", value: ticks.MIDITempo{}, text: "MIDITempo(unset)", equalZero: ticks.MIDITempo{}.Equal(ticks.MIDITempo{}), equalSet: ticks.MIDITempo{}.Equal(mtempo(1))},
		{name: "MIDITempoMap", value: ticks.MIDITempoMap{}, text: "MIDITempoMap(unset)", equalZero: ticks.MIDITempoMap{}.Equal(ticks.MIDITempoMap{}), equalSet: ticks.MIDITempoMap{}.Equal(kp10Map())},
		{name: "Click", value: ticks.Click{}, text: "Click(unset)", equalZero: ticks.Click{}.Equal(ticks.Click{}), equalSet: ticks.Click{}.Equal(ticks.ClickOnBeat())},
		{name: "MIDIMeter", value: ticks.MIDIMeter{}, text: "MIDIMeter(unset)", equalZero: ticks.MIDIMeter{}.Equal(ticks.MIDIMeter{}), equalSet: ticks.MIDIMeter{}.Equal(ticks.DefaultMIDIMeter())},
		{name: "MIDIMeterMap", value: ticks.MIDIMeterMap{}, text: "MIDIMeterMap(unset)", equalZero: ticks.MIDIMeterMap{}.Equal(ticks.MIDIMeterMap{}), equalSet: ticks.MIDIMeterMap{}.Equal(ticks.MustNewMIDIMeterMap(res(1)))},
		{name: "ExportOptions", value: ticks.ExportOptions{}, text: "ExportOptions(unset)", equalZero: ticks.ExportOptions{}.Equal(ticks.ExportOptions{}), equalSet: ticks.ExportOptions{}.Equal(ticks.MustNewExportOptions(rounding.Down))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.False(t, tt.value.IsSet())
			assert.Equal(t, tt.text, tt.value.String())
			assert.True(t, tt.equalZero)
			assert.False(t, tt.equalSet)
			out, err := tt.value.MarshalText()
			requireTicksError(t, err, ticks.ErrUnset)
			assert.Nil(t, out)
			dst := []byte("keep")
			appended, err := tt.value.AppendText(dst)
			requireTicksError(t, err, ticks.ErrUnset)
			assert.Equal(t, []byte("keep"), appended)
		})
	}
	_, err := ticks.Quantized{}.Ticks()
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = ticks.Quantized{}.Deviation()
	requireTicksError(t, err, ticks.ErrUnset)
	requireMustPanic(t, err, ticks.ErrUnset, func() { ticks.Quantized{}.MustDeviation() })
	_, err = ticks.MIDITempo{}.Microseconds()
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = ticks.MIDITempoMap{}.Changes()
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = ticks.MIDITempoMap{}.Initial()
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = ticks.MIDITempoMap{}.EqualTimeline(kp10Map())
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = ticks.MIDIMeterMap{}.Changes()
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = ticks.MIDIMeterMap{}.MeterAt(0)
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = ticks.ExportOptions{}.Holds()
	requireTicksError(t, err, ticks.ErrUnset)
	_, err = ticks.Click{}.MarshalText()
	requireTicksError(t, err, ticks.ErrUnset)
}

// TestComparableValues checks 3.1: for comparable types a == b ⇔ a.Equal(b).
func TestComparableValues(t *testing.T) {
	t.Parallel()
	quarters := []ticks.MIDIQuarter{{}, ticks.StandardQuarter(), ticks.MustNewMIDIQuarter(16), ticks.MustMIDIQuarterOf(dur(1, 2))}
	for _, a := range quarters {
		for _, b := range quarters {
			assert.Equal(t, a == b, a.Equal(b), "%s %s", a, b)
		}
	}
	resolutions := []ticks.Resolution{{}, res(480), resQ(480, 8), resQ(480, 16), res(96), ticks.MustParseResolution("480@2/4")}
	for _, a := range resolutions {
		for _, b := range resolutions {
			assert.Equal(t, a == b, a.Equal(b), "%s %s", a, b)
		}
	}
	quantized := []ticks.Quantized{{}, ticks.MustParseQuantized("69+1/4480"), ticks.MustParseQuantized("69+2/8960"), ticks.MustParseQuantized("720-0/1"), ticks.MustParseQuantized("720+0/1"), res(480).MustQuantize(dur(3, 8), rounding.Down)}
	for _, a := range quantized {
		for _, b := range quantized {
			assert.Equal(t, a == b, a.Equal(b), "%s %s", a, b)
		}
	}
	smptes := []ticks.SMPTE{{}, smpte(-25, 40), smpte(-29, 80), smpte(-29, 80).MustWithReading(ticks.Reading30), ticks.MustParseSMPTE("-29x80:30")}
	for _, a := range smptes {
		for _, b := range smptes {
			assert.Equal(t, a == b, a.Equal(b), "%s %s", a, b)
		}
	}
	tempos := []ticks.MIDITempo{{}, mtempo(1), ticks.DefaultMIDITempo(), mtempo(500000)}
	for _, a := range tempos {
		for _, b := range tempos {
			assert.Equal(t, a == b, a.Equal(b), "%s %s", a, b)
		}
	}
	clicks := []ticks.Click{{}, ticks.ClickOnBeat(), ticks.MustClickEvery(dur(1, 4)), ticks.MustParseClick("2/8")}
	for _, a := range clicks {
		for _, b := range clicks {
			assert.Equal(t, a == b, a.Equal(b), "%s %s", a, b)
		}
	}
	meters := []ticks.MIDIMeter{{}, ticks.DefaultMIDIMeter(), mmeter(4, 2, 24, 8), mmeter(6, 3, 36, 8)}
	for _, a := range meters {
		for _, b := range meters {
			assert.Equal(t, a == b, a.Equal(b), "%s %s", a, b)
		}
	}
	options := []ticks.ExportOptions{
		{},
		ticks.MustNewExportOptions(rounding.HalfUp),
		ticks.MustNewExportOptions(rounding.HalfUp).MustWithHoldsStretched(dur(1, 4)).MustWithHoldsRefused(),
		ticks.MustNewExportOptions(rounding.HalfUp).MustWithSteps(dur(1, 4), ticks.StepEnd).MustWithoutSteps(),
		ticks.MustNewExportOptions(rounding.HalfUp).MustWithPositionMode(rounding.Up).MustWithExactPositions(),
		ticks.MustNewExportOptions(rounding.HalfUp).MustWithSteps(dur(1, 4), ticks.StepEnd),
		ticks.MustNewExportOptions(rounding.Down),
	}
	for _, a := range options {
		for _, b := range options {
			assert.Equal(t, a == b, a.Equal(b), "%s %s", a, b)
		}
	}
	changes := []ticks.MIDITempoChange{{}, tchange(0, 1), tchange(0, 2), ticks.MustParseMIDITempoChange("0:1")}
	for _, a := range changes {
		for _, b := range changes {
			assert.Equal(t, a == b, a.Equal(b), "%s %s", a, b)
		}
	}
	meterChanges := []ticks.MIDIMeterChange{{}, mchange(0, 4, 2, 24, 8), mchange(1, 4, 2, 24, 8)}
	for _, a := range meterChanges {
		for _, b := range meterChanges {
			assert.Equal(t, a == b, a.Equal(b), "%s %s", a, b)
		}
	}
}

// TestFailedOperationsDoNotChangeValues checks 3.2: no operation changes its
// receiver or arguments, also on failure and Must panics.
func TestFailedOperationsDoNotChangeValues(t *testing.T) {
	t.Parallel()
	m := kp10Map()
	copied := m
	r := res(480)
	rit := ritardandoMap()
	ritText := rit.String()
	_, _ = m.WithChange(tchange(-1, 1))
	_, _ = m.ReplaceChange(0, 5, mtempo(1))
	_, _ = m.WithInitial(mtempo(1))
	_, _ = m.MustWithChange(tchange(5, 1)).Rescale(res(7))
	assert.Panics(t, func() { m.MustReplaceChange(0, 5, mtempo(1)) })
	assert.True(t, m.Equal(copied))
	assert.Equal(t, []string{"0:500000", "192:1000000"}, tempoChanges(t, m))

	_, _, _ = ticks.MIDITempoMapOf(rit, r, ticks.MustNewExportOptions(rounding.HalfUp))
	_, _, _ = ticks.MIDITempoMapOf(rit, r, ticks.MustNewExportOptions(rounding.HalfUp).MustWithSteps(dur(1, 4), ticks.StepEqualTime))
	_, _ = r.TickAt(rit, sec(1, 1), rounding.Down)
	assert.Equal(t, ritText, rit.String())
	assert.True(t, r.Equal(res(480)))

	o := ticks.MustNewExportOptions(rounding.HalfUp)
	_, _ = o.WithSteps(dur(1, 4), ticks.StepMethod(0))
	_ = o.MustWithSteps(dur(1, 4), ticks.StepStart)
	assert.True(t, o.Equal(ticks.MustNewExportOptions(rounding.HalfUp)))

	mm := ticks.MustNewMIDIMeterMap(r, mchange(0, 4, 2, 24, 8))
	_, _ = mm.WithChange(mchange(-1, 4, 2, 24, 8))
	_ = mm.MustWithChange(mchange(5, 3, 2, 24, 8))
	assert.Equal(t, []string{"0:4,2,24,8"}, meterChanges(t, mm))

	durations := []int{0, 96, 192}
	_, _ = ticks.Deltas(durations)
	assert.Equal(t, []int{0, 96, 192}, durations)
}

// TestConcurrentUse checks 3.2: values are safe for concurrent use; run with -race.
func TestConcurrentUse(t *testing.T) {
	t.Parallel()
	r := res(480)
	m := kp10Map()
	rit := ritardandoMap()
	o := ticks.MustNewExportOptions(rounding.HalfUp).MustWithSteps(dur(1, 4), ticks.StepEqualTime)
	meterMap := ticks.MustNewMIDIMeterMap(r, mchange(0, 4, 2, 24, 8), mchange(2880, 3, 2, 24, 8))
	s := smpte(-25, 40)
	wantExport, _ := ticks.MustMIDITempoMapOf(rit, r, o)
	wantGrid := meterMap.MustGrid()
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			assert.Equal(t, 720, r.MustTicks(dur(3, 8)))
			assertSeconds(t, sec(3, 1), m.MustSeconds(384))
			assert.Equal(t, 288, m.MustTickAt(sec(2, 1), rounding.Down))
			_ = m.MustWithChange(tchange(1000, 1)).String()
			_ = m.MustChanges()
			got, _ := ticks.MustMIDITempoMapOf(rit, r, o)
			assert.True(t, got.Equal(wantExport))
			assert.Equal(t, 921, r.MustTickAt(rit, sec(1, 1), rounding.Down))
			assert.Equal(t, 2187, s.MustStartTick(rit, pos(1, 1), rounding.Down))
			assert.True(t, meterMap.MustGrid().Equal(wantGrid))
			assert.Equal(t, "miditempo[96;500000;0:500000;192:1000000]", m.String())
		})
	}
	wg.Wait()
}

// TestIndependentUnmarshalReceivers checks concurrent UnmarshalText into different receivers.
func TestIndependentUnmarshalReceivers(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	results := make([]ticks.MIDITempoMap, 8)
	for i := range results {
		wg.Go(func() {
			assert.NoError(t, results[i].UnmarshalText([]byte("miditempo[96;500000;0:500000;192:1000000]")))
		})
	}
	wg.Wait()
	for _, got := range results {
		assert.True(t, got.Equal(kp10Map()))
	}
}
