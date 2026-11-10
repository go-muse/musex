package temperament_test

import (
	"errors"
	"fmt"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/tuning/temperament"
)

// Example demonstrates basic inspection of the built-in 12-EDO temperament.
func Example() {
	tm := temperament.Equal12()
	fmt.Println(tm)
	fmt.Println(tm.IsValid(), tm.StepsPerOctave(), tm.IsEqual(), tm.IsRational())
	// Output:
	// 12-EDO (12 steps/octave)
	// true 12 true false
}

// ExampleEDO demonstrates constructing an equal division of the octave with 19 steps.
func ExampleEDO() {
	tm, err := temperament.EDO(19)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s: step 1 = %.4f cents\n", tm.Name(), tm.MustOffset(1).Float64())
	// Output:
	// 19-EDO: step 1 = 63.1579 cents
}

// ExampleWerckmeisterIII demonstrates inspecting intervals in Werckmeister III (1691).
func ExampleWerckmeisterIII() {
	tm := temperament.WerckmeisterIII()
	fifthCG := tm.MustDirectedInterval(0, 7)
	thirdCE := tm.MustDirectedInterval(0, 4)
	fmt.Printf("%s: C-G = %.3f cents, C-E = %.3f cents\n", tm.Name(), fifthCG.Float64(), thirdCE.Float64())
	// Output:
	// Werckmeister III: C-G = 696.090 cents, C-E = 390.225 cents
}

// ExampleFromRatios demonstrates constructing a custom rational temperament from frequency ratios.
func ExampleFromRatios() {
	tm, err := temperament.FromRatios("Tetrad", []temperament.Ratio{
		{Num: 1, Den: 1},
		{Num: 5, Den: 4},
		{Num: 3, Den: 2},
		{Num: 15, Den: 8},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s (%d steps): third = %.3f cents, fifth = %.3f cents\n",
		tm.Name(), tm.StepsPerOctave(), tm.MustOffset(1).Float64(), tm.MustOffset(2).Float64())
	// Output:
	// Tetrad (4 steps): third = 386.314 cents, fifth = 701.955 cents
}

// ExampleFromFifthDeviations demonstrates constructing a temperament from fifth deviations in the circle of fifths.
func ExampleFromFifthDeviations() {
	quarterPC := cents.PythagoreanComma().MustDiv(4).Neg()
	z := cents.Zero()
	tm := temperament.MustFromFifthDeviations("Custom W3", []cents.Cents{
		quarterPC, quarterPC, quarterPC, z,
		z, quarterPC, z, z,
		z, z, z, z,
	})
	fmt.Println(tm.Equal(temperament.WerckmeisterIII()))
	// Output:
	// true
}

// ExampleTemperament_CumulativeOffset demonstrates computing cumulative cent offsets across positive and negative octaves.
func ExampleTemperament_CumulativeOffset() {
	tm := temperament.Equal12()
	fmt.Printf("k=-12: %.1f, k=-1: %.1f, k=0: %.1f, k=7: %.1f, k=19: %.1f\n",
		tm.MustCumulativeOffset(-12).Float64(),
		tm.MustCumulativeOffset(-1).Float64(),
		tm.MustCumulativeOffset(0).Float64(),
		tm.MustCumulativeOffset(7).Float64(),
		tm.MustCumulativeOffset(19).Float64(),
	)
	// Output:
	// k=-12: -1200.0, k=-1: -100.0, k=0: 0.0, k=7: 700.0, k=19: 1900.0
}

// ExampleTemperament_DirectedInterval demonstrates computing directed and cyclic intervals.
func ExampleTemperament_DirectedInterval() {
	tm := temperament.Just5Limit()
	dir := tm.MustDirectedInterval(2, 9)
	cyc := tm.MustCyclicInterval(9, 2)
	fmt.Printf("D->A = %.3f cents, A->D = %.3f cents\n", dir.Float64(), cyc.Float64())
	// Output:
	// D->A = 680.449 cents, A->D = 519.551 cents
}

// ExampleTemperament_Rotate demonstrates rotating a temperament to a new modal tonic.
func ExampleTemperament_Rotate() {
	tm := temperament.Just5Limit()
	dorian := tm.MustRotate(2)
	fmt.Printf("Name=%q Steps=%d Step7=%.3f cents\n",
		dorian.Name(), dorian.StepsPerOctave(), dorian.MustOffset(7).Float64())
	// Output:
	// Name="5-Limit Just Intonation" Steps=12 Step7=680.449 cents
}

// ExampleTemperament_NearestStep demonstrates finding the closest scale step on the octave circle.
func ExampleTemperament_NearestStep() {
	tm := temperament.Just5Limit()
	step, delta, err := tm.NearestStep(cents.MustFromInt(400))
	if err != nil {
		panic(err)
	}
	fmt.Printf("step=%d delta=%.3f cents\n", step, delta.Float64())
	// Output:
	// step=4 delta=13.686 cents
}

// ExampleTemperament_AnalyzeEDO demonstrates computing deviations from the corresponding N-EDO grid.
func ExampleTemperament_AnalyzeEDO() {
	tm := temperament.WerckmeisterIII()
	dev := tm.MustAnalyzeEDO()
	fmt.Printf("Max=%.3f (step %d), Mean=%.3f, RMS=%.3f\n",
		dev.MaxAbsDeviation.Float64(), dev.MaxPosition,
		dev.MeanAbsDeviation.Float64(), dev.RMSDeviation)
	// Output:
	// Max=11.730 (step 6), Mean=6.843, RMS=7.738
}

// ExampleTemperament_AnalyzeFifths demonstrates analyzing fifths and detecting wolf intervals.
func ExampleTemperament_AnalyzeFifths() {
	tm := temperament.MeantoneQuarterComma()
	fifths := tm.MustAnalyzeFifths()
	fmt.Printf("Pure=%d Min=%.3f Max=%.3f Spread=%.3f Wolves=%v\n",
		fifths.PureCount, fifths.Min.Float64(), fifths.Max.Float64(),
		fifths.Spread.Float64(), fifths.WolfPositions)
	// Output:
	// Pure=0 Min=696.578 Max=737.637 Spread=41.059 Wolves=[8]
}

// Example_quickStart reproduces the quick start of the package guide: preset names, major thirds,
// a frequency from A4 = 440 Hz, the 1/4-comma meantone wolf fifth, and quantizing a pure major third.
func Example_quickStart() {
	eq := temperament.Equal12()
	mt := temperament.MeantoneQuarterComma()
	v := temperament.Vallotti()
	fmt.Println(eq)
	fmt.Println(mt)
	fmt.Println(v)

	eqThird := eq.MustDirectedInterval(0, 4)
	mtThird := mt.MustDirectedInterval(0, 4)
	fmt.Printf("12-EDO C->E: %.4f cents\n", eqThird.Float64())
	fmt.Printf("1/4-Comma C->E: %.4f cents, pure 5:4: %t\n", mtThird.Float64(), mtThird.Equal(cents.MustFromRatio(5, 4)))

	a4ToE4 := mt.MustDirectedInterval(9, 4)
	fmt.Printf("1/4-Comma E4 (from A4 = 440 Hz): %.4f Hz\n", a4ToE4.MustApplyTo(440.0))

	fifths := mt.MustAnalyzeFifths()
	fmt.Printf("Pure fifths: %d, wolf fifths: %v\n", fifths.PureCount, fifths.WolfPositions)

	step, dev := eq.MustNearestStep(cents.MustFromRatio(5, 4))
	fmt.Printf("Nearest 12-EDO step to 5:4: %d (deviation %.4f cents)\n", step, dev.Float64())
	// Output:
	// 12-EDO (12 steps/octave)
	// 1/4-Comma Meantone (12 steps/octave)
	// Vallotti (12 steps/octave)
	// 12-EDO C->E: 400.0000 cents
	// 1/4-Comma C->E: 386.3137 cents, pure 5:4: true
	// 1/4-Comma E4 (from A4 = 440 Hz): 328.9767 Hz
	// Pure fifths: 0, wolf fifths: [8]
	// Nearest 12-EDO step to 5:4: 4 (deviation -13.6863 cents)
}

// ExampleVallotti computes the frequency of C4 in Vallotti's temperament from the concert pitch A4 = 440 Hz.
func ExampleVallotti() {
	v := temperament.Vallotti()
	a4ToC4 := v.MustDirectedInterval(9, 0)
	c4Hz, err := a4ToC4.ApplyTo(440.0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("A = %.6f cents above C, C4 = %.4f Hz\n", v.MustOffset(9).Float64(), c4Hz)
	// Output:
	// A = 894.134997 cents above C, C4 = 262.5134 Hz
}

// ExampleTemperament_NearestStep_harmonicSeventh quantizes the harmonic seventh 7:4 to 31-EDO.
func ExampleTemperament_NearestStep_harmonicSeventh() {
	edo31 := temperament.Equal31()
	seventh := cents.MustFromRatio(7, 4)
	step, dev := edo31.MustNearestStep(seventh)
	fmt.Printf("7:4 = %.4f cents -> step %d (%.4f cents), deviation %+.4f cents\n",
		seventh.Float64(), step, edo31.MustOffset(step).Float64(), dev.Float64())
	// Output:
	// 7:4 = 968.8259 cents -> step 25 (967.7419 cents), deviation +1.0840 cents
}

// ExampleTemperament_MustWithName rotates Werckmeister III to F# and renames the rotated temperament.
func ExampleTemperament_MustWithName() {
	w3 := temperament.WerckmeisterIII()
	onFSharp := w3.MustRotate(6).MustWithName("Werckmeister III (on F#)")
	third := onFSharp.MustOffset(4) // F# -> A#
	fmt.Println(onFSharp)
	fmt.Printf("F#->A# = %.4f cents, Pythagorean ditone 81:64: %t\n", third.Float64(), third.Equal(cents.MustFromRatio(81, 64)))
	fmt.Println(w3.Name())
	// Output:
	// Werckmeister III (on F#) (12 steps/octave)
	// F#->A# = 407.8200 cents, Pythagorean ditone 81:64: true
	// Werckmeister III
}

// ExampleTemperament_Near compares Vallotti's temperament with Vallotti's original scheme
// (six fifths narrowed by 1/6 syntonic comma and the fifth Bb-F narrowed by one schisma).
func ExampleTemperament_Near() {
	sixthSc := cents.SyntonicComma().MustScale(-1, 6)
	z := cents.Zero()
	original := temperament.MustFromFifthDeviations("Vallotti Original", []cents.Cents{
		sixthSc, sixthSc, sixthSc, sixthSc, sixthSc, // C-G, G-D, D-A, A-E, E-B
		z, z, z, z, z, // B-F#, F#-C#, C#-G#, G#-Eb, Eb-Bb
		cents.Schisma().Neg(), // Bb-F
		sixthSc,               // F-C
	})
	v := temperament.Vallotti()
	maxDiff := cents.Schisma().MustScale(5, 6)
	fmt.Printf("Bb: %.6f vs %.6f cents\n", original.MustOffset(10).Float64(), v.MustOffset(10).Float64())
	fmt.Println(original.Equal(v), original.Near(v, cents.MustFromInt(1)), original.Near(v, maxDiff))
	fmt.Printf("max difference = 5/6 schisma = %.6f cents\n", maxDiff.Float64())
	// Output:
	// Bb: 1001.628101 vs 1000.000000 cents
	// false false true
	// max difference = 5/6 schisma = 1.628101 cents
}

// ExampleNew shows the error classes of invalid input, checked with errors.Is.
func ExampleNew() {
	_, err := temperament.New("Invalid", []cents.Cents{cents.Zero(), cents.MustFromInt(500), cents.MustFromInt(400)})
	fmt.Println(errors.Is(err, temperament.ErrInvalidArgument))

	_, err = temperament.FromRatios("139-Limit", []temperament.Ratio{{Num: 1, Den: 1}, {Num: 139, Den: 128}})
	fmt.Println(errors.Is(err, temperament.ErrPrimeLimit))

	var zero temperament.Temperament
	_, err = zero.Offset(0)
	fmt.Println(errors.Is(err, temperament.ErrInvalidArgument), zero)
	// Output:
	// true
	// true
	// true Temperament(<invalid>)
}
