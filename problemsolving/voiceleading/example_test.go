package voiceleading_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/problemsolving/voiceleading"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/voice"
)

func ExampleAnalyze() {
	quarter := voice.MustValueDuration(duration.MustNew(1, 4))
	sound := func(name string, octave int) voice.Element {
		n := note.MustFromPitch(pitch.MustFromASPN(spelling.MustParse(name), octave), cents.Zero(), nil)
		return voice.MustNewSound(n, quarter)
	}
	bass := voice.MustNew(sound("C", 3), sound("D", 3)).MustWithName("Bass")
	middle := voice.MustNew(sound("G", 3), sound("A", 3))
	upper := voice.MustNew(sound("C", 4), sound("D", 4))
	a, err := voiceleading.Analyze([]*voice.Line{bass, middle, upper}, voiceleading.Options{System: system.Traditional(), Order: voiceleading.LowToHigh})
	if err != nil {
		panic(err)
	}
	pair := voiceleading.Pair{First: 0, Second: 1}
	v, ok := a.MustVerticalAt(pair, duration.Origin())
	fmt.Println(a.MustPartCount(), a.MustName(0), ok, v.Specific.Size(), v.Chromatic)
	fmt.Println(a.MustTransitions(pair)[0].Motion)
	fmt.Println(len(a.MustFindParallels(voiceleading.ParallelOptions{Perfect: voiceleading.PerfectBoth})))
	_, err = a.TransitionAt(pair, duration.Origin(), duration.Origin())
	fmt.Println(errors.Is(err, voiceleading.ErrTransition))
	xs := a.MustVerticals(pair)
	xs[0].Chromatic = 123
	fmt.Println(a.MustVerticals(pair)[0].Chromatic)
	assessment := a.MustCheckOMTFirstSpecies(pair)
	fmt.Println(assessment.Applicable, len(assessment.Violations))
	// Output:
	// 3 Bass true 5 7
	// parallel
	// 2
	// true
	// 7
	// true 1
}
