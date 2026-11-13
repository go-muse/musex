package tuning_test

import (
	"errors"
	"fmt"
	"math"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
	"github.com/go-muse/musex/tuning/temperament"
)

func ExampleStandard() {
	std := tuning.Standard()
	fmt.Printf("A4: %.1f Hz\n", std.MustFrequency(pitch.MustFromASPN(spelling.A(), 4)))
	fmt.Printf("A5: %.1f Hz\n", std.MustFrequency(pitch.MustFromASPN(spelling.A(), 5)))
	fmt.Printf("C5: %.4f Hz\n", std.MustFrequency(pitch.MustFromASPN(spelling.C(), 5)))
	// Output:
	// A4: 440.0 Hz
	// A5: 880.0 Hz
	// C5: 523.2511 Hz
}

func ExampleNew() {
	tune := tuning.MustNew(
		system.Traditional(),
		temperament.PythagoreanEFlatGSharp(),
		tuning.RootFromName(spelling.C()),
		tuning.A4At440(),
	)
	frequency := tune.MustFrequency(pitch.MustFromASPN(spelling.F(), 5))
	fmt.Printf("F5: %.4f Hz\n", frequency)
	// Output:
	// F5: 695.3086 Hz
}

func ExampleTuning_Nearest() {
	std := tuning.Standard()
	step, deviation := std.MustNearest(445)
	note := std.MustPitchOfStep(step, system.Sharps())
	fmt.Printf("step %d, %s, deviation %s cents\n", step, note, deviation.Decimal(6, 0))
	// Output:
	// step 0, A ASPN 4, deviation 19.562175 cents
}

func ExampleTuning_NearestBy() {
	std := tuning.Standard()
	note, deviation := std.MustNearestPitchBy(445, rounding.Up, system.Sharps())
	fmt.Printf("%s, deviation %s cents\n", note, deviation.Decimal(6, 0))
	// Output:
	// A+1 ASPN 4, deviation -80.437825 cents
}

func ExampleTuning_WithReferenceNote() {
	std := tuning.Standard()
	c5 := pitch.MustFromASPN(spelling.C(), 5)
	reexpressed := std.MustWithReferenceNote(c5)
	difference := std.MustLevelDifference(reexpressed)
	same := std.MustSoundsSame(reexpressed)
	fmt.Printf("new reference: %s = %.4f Hz\n", reexpressed.Reference().Pitch(), reexpressed.Reference().Frequency())
	fmt.Printf("level difference: %s cents\n", difference.Decimal(6, 0))
	fmt.Printf("strictly equal frequencies: %t\n", same)
	// Output:
	// new reference: C ASPN 5 = 523.2511 Hz
	// level difference: 0.000000 cents
	// strictly equal frequencies: false
}

func ExampleTuning_ShiftedFrequency() {
	std := tuning.Standard()
	c4 := pitch.MustFromASPN(spelling.C(), 4)
	cSharp4 := pitch.MustFromASPN(spelling.CSharp(), 4)
	shifted := std.MustShiftedFrequency(c4, cents.MustFromInt(100))
	plain := std.MustFrequency(cSharp4)
	fmt.Printf("bitwise equal: %t\n", math.Float64bits(shifted) == math.Float64bits(plain))
	fmt.Printf("A4 + 1200 cents: %.1f Hz\n", std.MustShiftedFrequency(pitch.MustFromASPN(spelling.A(), 4), cents.Octave()))
	// Output:
	// bitwise equal: true
	// A4 + 1200 cents: 880.0 Hz
}

func ExampleNew_classCountMismatch() {
	_, err := tuning.New(
		system.Positions19(),
		temperament.Equal12(),
		tuning.RootFromClass(pitchclass.MustNew(0, 19)),
		tuning.A4At440(),
	)
	fmt.Println(errors.Is(err, tuning.ErrClassCount))
	// Output:
	// true
}
