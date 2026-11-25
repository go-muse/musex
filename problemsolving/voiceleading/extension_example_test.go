package voiceleading_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/chord"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"
)

// The examples use natural notes, whose ASPN octave is also the written one,
// and the musical values of business-spec 33.2.

func exampleNote(name string, aspn int) note.Note {
	return note.MustFromPitch(pitch.MustFromASPN(spelling.MustParse(name), aspn), cents.Zero(), nil)
}

func exampleName(name string) note.Note {
	return note.MustFromSpelling(spelling.MustParse(name), cents.Zero(), nil)
}

func exampleNotes(notes []note.Note) string {
	names := make([]string, len(notes))
	for i, n := range notes {
		names[i] = fmt.Sprintf("%s%d", n.Spelling(), n.MustPitch().MustASPN())
	}
	return strings.Join(names, " ")
}

func ExampleNewMelody() {
	quarter := voice.MustValueDuration(duration.MustNew(1, 4))
	sound := func(name string, aspn int) voice.Element { return voice.MustNewSound(exampleNote(name, aspn), quarter) }
	line := voice.MustNew(sound("C", 4), sound("E", 4), sound("G", 4), sound("C", 5), sound("B", 4))
	m, err := voiceleading.NewMelody(line, system.Traditional())
	if err != nil {
		panic(err)
	}
	for _, mv := range m.MustMoves() {
		fmt.Println(mv.Interval, mv.Chromatic, mv.IsLeap(voiceleading.LeapFromThird), mv.IsLeap(voiceleading.LeapFromFourth))
	}
	fmt.Println(m.MustLeapRuns(voiceleading.LeapFromThird, 3))
	fmt.Println(len(m.MustLeapRuns(voiceleading.LeapFromFourth, 2)))
	a := m.MustCheck(voiceleading.OMTLeapRecovery)
	fmt.Println(a.Rule, len(a.Violations))
	other := voiceleading.MustNewMelody(voice.MustNew(sound("C", 4), sound("A", 4), sound("B", 4)), system.Traditional())
	v := other.MustCheck(voiceleading.OMTLeapRecovery).Violations[0]
	fmt.Println(v.Index, v.Move.Interval, v.Next.Interval, v.Reasons)
	_, err = m.LeapRuns(voiceleading.LeapFromThird, 0)
	fmt.Println(errors.Is(err, voiceleading.ErrOption))
	// Output:
	// M3 4 true false
	// m3 3 true false
	// P4 5 true true
	// -m2 -1 false false
	// [{0 2}]
	// 0
	// omt-leap-recovery 0
	// 0 M6 M2 [no-step-back]
	// true
}

func ExampleNewSonority() {
	sys := system.Traditional()
	s := voiceleading.MustNewSonority(chord.MustNew(exampleNote("C", 3), exampleNote("E", 4), exampleNote("G", 4), exampleNote("C", 5)), sys)
	fmt.Println(s.MustClassifySpacing(voiceleading.SpacingME).Class,
		s.MustClassifySpacing(voiceleading.SpacingAM).Class,
		s.MustClassifySpacing(voiceleading.SpacingOMT).Class)
	spacing := s.MustSpacing()
	fmt.Println(spacing.Span.Interval, spacing.Span.Chromatic, spacing.WithinOctave)
	d := s.MustDoublings(voiceleading.BySpelling, chord.MustNew(exampleName("C"), exampleName("E"), exampleName("G")))
	doubled := d.Doubled()[0]
	fmt.Println(d.BassDoubled, d.Complete, doubled.Name, doubled.Voices, doubled.Doubling)
	satb := voiceleading.MustNewSonority(chord.MustNew(exampleNote("C", 3), exampleNote("G", 3), exampleNote("C", 4), exampleNote("E", 5)), sys)
	a := satb.MustCheckDistances(voiceleading.DistanceNorm{Rule: voiceleading.DistanceMT21SATB})
	v := a.Violations[0]
	fmt.Println(a.Rule, v.Pair.Lower, v.Pair.Upper, v.Pair.Interval, v.Limit)
	// Output:
	// close mixed omt-open
	// P15 24 false
	// true true C [0 3] octave
	// mt21-satb 2 3 M10 P8
}

func ExampleAnalysis_Sonorities() {
	sound := func(name string, aspn int, d duration.Duration) voice.Element {
		return voice.MustNewSound(exampleNote(name, aspn), voice.MustValueDuration(d))
	}
	half, quarter := duration.MustNew(1, 2), duration.MustNew(1, 4)
	bass := voice.MustNew(sound("C", 3, half))
	tenor := voice.MustNew(sound("G", 3, quarter), sound("A", 3, quarter))
	soprano := voice.MustNew(sound("C", 4, half))
	a := voiceleading.MustAnalyze([]*voice.Line{bass, tenor, soprano}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh})
	for _, ps := range a.MustSonorities() {
		fmt.Println(ps.Start, ps.End, exampleNotes(ps.Voices.MustNotes()), ps.Voices.MustSpacing().WithinOctave)
	}
	c := a.MustConnections()[0].Connection
	fmt.Println(c.Moves[1].Interval, c.Kind, c.Total)
	// Output:
	// 0/1 1/4 C3 G3 C4 true
	// 1/4 1/2 C3 A3 C4 true
	// M2 harmonic 2
}

func ExampleConnect() {
	sys := system.Traditional()
	from := voiceleading.MustNewSonority(chord.MustNew(exampleNote("C", 3), exampleNote("E", 4), exampleNote("G", 4), exampleNote("C", 5)), sys)
	to := voiceleading.MustNewSonority(chord.MustNew(exampleNote("F", 3), exampleNote("F", 4), exampleNote("A", 4), exampleNote("C", 5)), sys)
	c := voiceleading.MustConnect(from, to)
	moves := make([]string, len(c.Moves))
	for i, mv := range c.Moves {
		moves[i] = mv.Interval.String()
	}
	fmt.Println(strings.Join(moves, " "))
	fmt.Println(c.Kind, c.Total, c.Common)
	// Output:
	// P4 m2 M2 P1
	// harmonic 8 [C]
}

func ExampleCheckMT21Connection() {
	sys := system.Traditional()
	from := voiceleading.MustNewSonority(chord.MustNew(exampleNote("C", 3), exampleNote("E", 4), exampleNote("G", 4), exampleNote("C", 5)), sys)
	to := voiceleading.MustNewSonority(chord.MustNew(exampleNote("F", 3), exampleNote("F", 4), exampleNote("A", 4), exampleNote("C", 5)), sys)
	a := voiceleading.MustCheckMT21Connection(from, to)
	fmt.Println(a.Rule, a.Applicable, a.Conforms)
	for _, v := range a.Variants {
		fmt.Println(v.Variant, v.Satisfied)
	}
	// Output:
	// mt21-root-position-triads true true
	// 26.7.2-common-tone true
	// 26.7.2-same-direction false
}

func ExampleChooseConnections() {
	sys := system.Traditional()
	from := voiceleading.MustNewSonority(chord.MustNew(exampleNote("C", 3), exampleNote("E", 4), exampleNote("G", 4), exampleNote("C", 5)), sys)
	target := chord.MustNew(exampleName("F"), exampleName("A"), exampleName("C"))
	choices, err := voiceleading.ChooseConnections(from, target, exampleNote("F", 3), voiceleading.ConnectionConditions{
		Complete: true, CommonTones: true, NoCrossings: true,
		Parallels: voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(len(choices))
	for _, c := range choices[:3] {
		fmt.Println(exampleNotes(c.To.MustNotes()), c.Total)
	}
	// Output:
	// 11
	// F3 F4 A4 C5 8
	// F3 C4 A4 C5 11
	// F3 A4 A4 C5 12
}

func ExampleDecompose() {
	quarter := voice.MustValueDuration(duration.MustNew(1, 4))
	dec := voiceleading.MustDecompose([]voiceleading.Simultaneity{
		{Notes: chord.MustNew(exampleNote("G", 4), exampleNote("C", 3), exampleNote("C", 5), exampleNote("E", 4)), Duration: quarter},
		{Notes: chord.MustNew(exampleNote("G", 2), exampleNote("D", 4), exampleNote("G", 4), exampleNote("B", 4)), Duration: quarter},
	}, system.Traditional())
	for _, l := range dec.Lines {
		var notes []note.Note
		for _, e := range l.MustElements() {
			n, _ := e.MustNote()
			notes = append(notes, n)
		}
		fmt.Println(exampleNotes(notes))
	}
	a := voiceleading.MustAnalyze(dec.Lines, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh})
	fmt.Println(len(dec.Ambiguities), len(a.MustFindParallels(voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth})))
	// Output:
	// C3 G2
	// E4 D4
	// G4 G4
	// C5 B4
	// 0 0
}
