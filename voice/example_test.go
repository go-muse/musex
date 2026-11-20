package voice_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/event"
	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/interval"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
	"github.com/go-muse/musex/voice"
)

// exampleNote returns a note with the standard settings.
func exampleNote(name spelling.Spelling, aspn int) note.Note {
	return note.MustFromASPN(name, aspn, cents.Zero(), settings.Standard())
}

func exampleDuration(base duration.Base) voice.Duration {
	return voice.MustWrittenDuration(duration.MustNewWritten(base, 0))
}

// A line without measures: starts of elements, its length and the element at
// a moment (КП-1, КП-2).
func ExampleNew() {
	q, e, h := exampleDuration(duration.Quarter), exampleDuration(duration.Eighth), exampleDuration(duration.Half)
	line := voice.MustNew(
		voice.MustNewSound(exampleNote(spelling.C(), 4), q),
		voice.MustNewSound(exampleNote(spelling.D(), 4), e),
		voice.MustNewSound(exampleNote(spelling.E(), 4), e),
		voice.MustNewSound(exampleNote(spelling.F(), 4), h),
	).MustWithName("Melody").MustWithPlacement(voice.Soprano)

	starts, end := line.MustStarts()
	fmt.Println(starts, end)
	k, ok := line.MustAt(duration.MustNewPosition(3, 8))
	fmt.Println(k, ok)
	_, ok = line.MustAt(end)
	fmt.Println(ok)
	fmt.Println(line.MustName(), line.MustPlacement())
	// Output:
	// [0/1 1/4 3/8 1/2] 1/1
	// 2 true
	// false
	// Melody soprano
}

// Melodic intervals of a line (КП-5) and an interval through a rest (КП-4).
func ExampleLine_Intervals() {
	q := exampleDuration(duration.Quarter)
	melody := voice.MustNew(
		voice.MustNewSound(exampleNote(spelling.C(), 4), q),
		voice.MustNewSound(exampleNote(spelling.D(), 4), q),
		voice.MustNewSound(exampleNote(spelling.G(), 4), q),
		voice.MustNewRest(q),
		voice.MustNewSound(exampleNote(spelling.E(), 4), q),
	)
	for _, m := range melody.MustIntervals() {
		fmt.Println(m.From, m.To, m.Interval, m.ThroughRest)
	}
	// Output:
	// 0 1 M2 false
	// 1 2 P4 false
	// 2 4 -m3 true
}

// The tessitura relative to the soprano voice (КП-16). The kind is a rule of
// the product: the mean position of the sounds, not whether the line fits.
func ExampleLine_Tessitura() {
	q, h := exampleDuration(duration.Quarter), exampleDuration(duration.Half)
	line := voice.MustNew(
		voice.MustNewSound(exampleNote(spelling.C(), 4), q),
		voice.MustNewSound(exampleNote(spelling.E(), 4), q),
		voice.MustNewSound(exampleNote(spelling.G(), 4), h),
	)
	soprano := voice.PresetReference(instrument.SopranoVoice())
	t := line.MustTessitura(soprano)
	fmt.Println(t.Number, t.Kind)
	byCount := line.MustTessituraBy(soprano, voice.ByCount, voice.DefaultZones())
	fmt.Println(byCount.Number, byCount.Kind)
	// Output:
	// 9/38 low
	// 11/57 low
}

// An exact imitation a fifth above after a half rest (КП-11).
func ExampleLine_ImitateAfter() {
	q := exampleDuration(duration.Quarter)
	subject := voice.MustNew(
		voice.MustNewSound(exampleNote(spelling.C(), 4), q),
		voice.MustNewSound(exampleNote(spelling.D(), 4), q),
		voice.MustNewSound(exampleNote(spelling.E(), 4), q),
	)
	answer := subject.MustImitateAfter(interval.PerfectFifth(), exampleDuration(duration.Half))
	fmt.Println(answer.MustThematic(), answer.MustLength())
	for _, s := range answer.MustSounds() {
		fmt.Println(s.Index, s.Start, s.End)
	}
	// Output:
	// imitating 5/4
	// 1 1/2 3/4
	// 2 3/4 1/1
	// 3 1/1 5/4
}

// An exact inversion around C4 (КП-13): C4, B flat 3, A flat 3.
func ExampleLine_Invert() {
	q := exampleDuration(duration.Quarter)
	subject := voice.MustNew(
		voice.MustNewSound(exampleNote(spelling.C(), 4), q),
		voice.MustNewSound(exampleNote(spelling.D(), 4), q),
		voice.MustNewSound(exampleNote(spelling.E(), 4), q),
	)
	mirror := subject.MustInvert(exampleNote(spelling.C(), 4))
	fmt.Println(mirror.MustThematic())
	for _, m := range mirror.MustIntervals() {
		fmt.Println(m.Interval)
	}
	// Output:
	// inversion
	// -M2
	// -M2
}

// Seconds of the sounds with a hold of 2 s at the end of the fermata
// (КП-25). The fermata itself adds no time: the hold is in the map.
func ExampleLine_SoundSeconds() {
	q, h := exampleDuration(duration.Quarter), exampleDuration(duration.Half)
	m := tempomap.MustNew(tempo.MustNewQuarter(fraction.FromInt(60))).
		MustWithExtension(tempomap.MustNewExtension(
			tempomap.MustAt(duration.MustNewPosition(3, 4)), tempo.MustNewSeconds(2, 1)))
	held := voice.MustNew(
		voice.MustNewSound(exampleNote(spelling.C(), 4), q),
		voice.MustNewSound(exampleNote(spelling.D(), 4), h).MustWithFermata(true),
		voice.MustNewSound(exampleNote(spelling.E(), 4), q),
	)
	for _, s := range held.MustSoundSeconds(m, 3, rounding.HalfUp) {
		fmt.Println(s.Index, s.Start.MustValue(), s.End.MustValue())
	}
	for _, f := range held.MustFermatas() {
		fmt.Println(f.Index, f.End)
	}
	// Output:
	// 0 0/1 1/1
	// 1 1/1 5/1
	// 2 5/1 6/1
	// 1 3/4
}

// Events with a tie become one sound; a chord is a distinguishable failure
// with the index of the event (КП-27, КП-30).
func ExampleFromEvents() {
	c4 := exampleNote(spelling.C(), 4)
	e4 := exampleNote(spelling.E(), 4)
	quarter := duration.MustNewWritten(duration.Quarter, 0)
	eighth := duration.MustNewWritten(duration.Eighth, 0)
	tiedC := event.MustNew(event.MustNoteContent(c4), event.MustWrittenDuration(quarter), event.Marks{Ties: []int{0}})
	nextC := event.MustNew(event.MustNoteContent(c4), event.MustWrittenDuration(eighth), event.Marks{})
	line := voice.MustFromEvents(tiedC, nextC)
	fmt.Println(line.MustLen(), line.MustLength())

	chordEvent := event.MustNew(event.MustChordContent(chord.MustNew(c4, e4)), event.MustWrittenDuration(quarter), event.Marks{})
	_, err := voice.FromEvents(tiedC, nextC, chordEvent)
	if ie, ok := errors.AsType[*voice.ItemError](err); ok && errors.Is(err, voice.ErrChord) {
		fmt.Println(ie.Index)
	}
	fmt.Println(len(line.MustEvents()))
	// Output:
	// 1 3/8
	// 2
	// 2
}

// Roles are parsed by their English or Russian names.
func ExampleParsePlacement() {
	p := voice.MustParsePlacement("тенор")
	fmt.Println(p, p.RussianName(), p.IsInner())
	r, ok := p.Range()
	fmt.Println(ok, r.Equal(instrument.TenorVoice().Range()))
	_, err := voice.ParsePlacement("Tenor ")
	fmt.Println(errors.Is(err, voice.ErrUnknownName))
	// Output:
	// tenor тенор true
	// true true
	// true
}
