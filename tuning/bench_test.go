package tuning_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"
)

var (
	benchStep    system.Step
	benchCents   cents.Cents
	benchFloat   float64
	benchPitch   pitch.Pitch
	benchClasses []pitchclass.Class
	benchTuning  *tuning.Tuning
	benchBool    bool
	benchError   error
)

func benchmarkFixture(classes int) (*tuning.Tuning, *tuning.Tuning) {
	if classes == 12 {
		std := tuning.Standard()
		return std, std.MustWithReferenceFrequency(435)
	}
	tune := tuning.MustNew(system.ExtendedPythagorean53(), temperament.Equal53(), tuning.RootFromClass(pitchclass.MustNew(0, 53)), tuning.A4At440())
	return tune, tune.MustWithReferenceFrequency(435)
}

func BenchmarkFrequencyOfStep(b *testing.B) {
	for _, classes := range []int{12, 53} {
		b.Run(fmt.Sprintf("N=%d", classes), func(b *testing.B) {
			tune, _ := benchmarkFixture(classes)
			b.ReportAllocs()
			for b.Loop() {
				benchFloat, benchError = tune.FrequencyOfStep(3)
			}
		})
	}
}

func BenchmarkFrequency(b *testing.B) {
	for _, classes := range []int{12, 53} {
		b.Run(fmt.Sprintf("N=%d", classes), func(b *testing.B) {
			tune, _ := benchmarkFixture(classes)
			note := aspn(spelling.C(), 5)
			b.ReportAllocs()
			for b.Loop() {
				benchFloat, benchError = tune.Frequency(note)
			}
		})
	}
}

func BenchmarkShiftedFrequency(b *testing.B) {
	for _, classes := range []int{12, 53} {
		b.Run(fmt.Sprintf("N=%d", classes), func(b *testing.B) {
			tune, _ := benchmarkFixture(classes)
			note := aspn(spelling.C(), 5)
			shift := cents.MustFromInt(7)
			b.ReportAllocs()
			for b.Loop() {
				benchFloat, benchError = tune.ShiftedFrequency(note, shift)
			}
		})
	}
}

func BenchmarkNearest(b *testing.B) {
	for _, classes := range []int{12, 53} {
		b.Run(fmt.Sprintf("N=%d", classes), func(b *testing.B) {
			tune, _ := benchmarkFixture(classes)
			b.ReportAllocs()
			for b.Loop() {
				benchStep, benchCents, benchError = tune.Nearest(445)
			}
		})
	}
}

func BenchmarkSoundsSame(b *testing.B) {
	for _, classes := range []int{12, 53} {
		b.Run(fmt.Sprintf("N=%d", classes), func(b *testing.B) {
			tune, other := benchmarkFixture(classes)
			b.ReportAllocs()
			for b.Loop() {
				benchBool, benchError = tune.SoundsSame(other)
			}
		})
	}
}

func BenchmarkChangedClasses(b *testing.B) {
	for _, classes := range []int{12, 53} {
		b.Run(fmt.Sprintf("N=%d", classes), func(b *testing.B) {
			tune, _ := benchmarkFixture(classes)
			root := tuning.RootFromClass(pitchclass.MustNew(1, classes))
			b.ReportAllocs()
			for b.Loop() {
				benchClasses, benchError = tune.ChangedClasses(root)
			}
		})
	}
}

func BenchmarkNew(b *testing.B) {
	for _, classes := range []int{12, 53} {
		b.Run(fmt.Sprintf("N=%d", classes), func(b *testing.B) {
			tune, _ := benchmarkFixture(classes)
			root := tuning.RootFromClass(pitchclass.MustNew(0, classes))
			b.ReportAllocs()
			for b.Loop() {
				benchTuning, benchError = tuning.New(tune.System(), tune.Temperament(), root, tuning.A4At440())
			}
		})
	}
}

func BenchmarkPitchOfStep(b *testing.B) {
	for _, classes := range []int{12, 53} {
		b.Run(fmt.Sprintf("N=%d", classes), func(b *testing.B) {
			tune, _ := benchmarkFixture(classes)
			b.ReportAllocs()
			for b.Loop() {
				benchPitch, benchError = tune.PitchOfStep(3, system.Sharps())
			}
		})
	}
}
