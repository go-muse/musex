package ticks_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/format/ticks"
	"github.com/go-muse/musex/measure"
	"github.com/go-muse/musex/meter"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
)

// Durations and positions in ticks at 480 ticks per quarter.
func ExampleResolution() {
	r := ticks.MustNewResolution(480)
	fmt.Println(r.MustTicks(duration.MustNew(3, 8)))
	fmt.Println(r.MustPosition(1440))
	_, err := ticks.MustNewResolution(2).Ticks(duration.MustNew(1, 12))
	fmt.Println(errors.Is(err, ticks.ErrNotRepresentable))
	fmt.Println(r.MustQuantize(duration.MustNew(1, 28), rounding.HalfEven))
	// Output:
	// 720
	// 3/4
	// true
	// 69+1/4480
}

// A MIDI quarter written as a half note (bb = 16).
func ExampleResolution_WithQuarter() {
	r := ticks.MustNewResolution(480).MustWithQuarter(ticks.MustNewMIDIQuarter(16))
	fmt.Println(r)
	fmt.Println(r.MustTicks(duration.MustNew(1, 2)))
	// Output:
	// 480@1/2
	// 480
}

// The smallest resolution for a set of durations, raised to at least 480.
func ExampleMinResolution() {
	ds := []duration.Duration{duration.MustNew(1, 4), duration.MustNew(1, 8), duration.MustNew(1, 12)}
	fmt.Println(ticks.MustMinResolution(ds, nil))
	fmt.Println(ticks.MustMinResolution(ds, nil).MustAtLeast(480).MustAtMost(ticks.MaxSMFResolution))
	fmt.Println(ticks.MustMinResolution([]duration.Duration{duration.MustNew(1, 28)}, nil).MustAtLeast(480))
	// Output:
	// 6
	// 480
	// 483
}

// MIDI tempo and quarters per minute.
func ExampleMIDITempoOf() {
	fmt.Println(ticks.MustMIDITempoOf(tempo.MustParse("4.=60"), rounding.HalfEven))
	fmt.Println(ticks.MustNewMIDITempo(600000).MustQuarterTempo())
	fmt.Println(ticks.MustFormatQuartersPerMinute(tempo.MustParse("4=600/7"), 2, rounding.HalfUp))
	// Output:
	// 666667
	// 4=100
	// 85.71
}

// Seconds by a MIDI tempo map and back.
func ExampleMIDITempoMap() {
	m := ticks.MustNewMIDITempoMap(ticks.MustNewResolution(96),
		ticks.MIDITempoChange{Tick: 0, Tempo: ticks.MustNewMIDITempo(500000)},
		ticks.MIDITempoChange{Tick: 192, Tempo: ticks.MustNewMIDITempo(1000000)})
	fmt.Println(m.MustSeconds(384))
	fmt.Println(m.MustTickAt(tempo.MustNewSeconds(2, 1), rounding.Down))
	fmt.Println(m)
	// Output:
	// 3/1
	// 288
	// miditempo[96;500000;0:500000;192:1000000]
}

// A ritardando of a musex tempo map written as MIDI tempo steps of a quarter
// that last as long as the parts of the change.
func ExampleMIDITempoMapOf() {
	quarter := func(bpm int64) tempo.Tempo { return tempo.MustParse(fmt.Sprintf("4=%d", bpm)) }
	at := func(n, d int) tempomap.Address { return tempomap.MustAt(duration.MustNewPosition(n, d)) }
	rit := tempomap.MustNew(quarter(120)).
		MustWithChange(tempomap.MustNewGradual(at(0, 1), tempomap.LinearSpeed).MustWithSlot(tempomap.SlotSecond)).
		MustWithChange(tempomap.MustNewMain(at(1, 1), quarter(100)))
	o := ticks.MustNewExportOptions(rounding.HalfUp).
		MustWithSteps(duration.MustNew(1, 4), ticks.StepEqualTime)
	m, skipped := ticks.MustMIDITempoMapOf(rit, ticks.MustNewResolution(480), o)
	fmt.Println(m)
	fmt.Println(len(skipped))
	// Output:
	// miditempo[480;500000;0:510715;480:533421;960:558240;1440:585482;1920:600000]
	// 0
}

// Time signatures for MIDI, strictly and with a loss of the record.
func ExampleMIDIMeterOf() {
	m, losses := ticks.MustMIDIMeterOf(meter.SixEight(), ticks.ClickOnBeat(), ticks.StandardQuarter(), ticks.Strict)
	fmt.Println(m, len(losses))
	m, losses = ticks.MustMIDIMeterOf(meter.CommonTime(), ticks.ClickOnBeat(), ticks.StandardQuarter(), ticks.Lossy)
	fmt.Println(m, len(losses))
	for _, loss := range losses {
		fmt.Println(loss)
	}
	_, _, err := ticks.MIDIMeterOf(meter.CommonTime(), ticks.ClickOnBeat(), ticks.StandardQuarter(), ticks.Strict)
	fmt.Println(errors.Is(err, ticks.ErrMeterNotMIDI))
	// Output:
	// 6,3,36,8 0
	// 4,2,24,8 1
	// symbol
	// true
}

// Time signature changes of a measure grid and back.
func ExampleMIDIMeterMapOf() {
	g := measure.MustNew(meter.FourFour()).MustWithMark(3, meter.ThreeFour())
	m, losses := ticks.MustMIDIMeterMapOf(g, ticks.MustNewResolution(480), ticks.ClickOnBeat(), ticks.Strict, ticks.IrregularEvent)
	fmt.Println(m)
	fmt.Println(len(losses))
	fmt.Println(m.MustGrid())
	// Output:
	// midimeter[480;0:4,2,24,8;3840:3,2,24,8]
	// 0
	// 4/4;3:3/4
}

// SMPTE resolution: the division word and seconds.
func ExampleSMPTE() {
	s := ticks.MustNewSMPTE(ticks.FrameRate25, 40)
	fmt.Println(s)
	fmt.Printf("%#04x\n", s.MustWord())
	fmt.Println(s.MustSeconds(1000))
	_, err := ticks.MustNewSMPTE(ticks.FrameRate29, 80).Seconds(2400)
	fmt.Println(errors.Is(err, ticks.ErrInvalidArgument))
	fmt.Println(ticks.MustNewSMPTE(ticks.FrameRate29, 80).MustWithReading(ticks.ReadingDropFrame).MustSeconds(2400))
	// Output:
	// -25x40
	// 0xe728
	// 1/1
	// true
	// 1001/1000
}
