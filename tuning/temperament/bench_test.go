package temperament_test

import (
	"testing"

	"github.com/go-muse/musex/cents"
	"github.com/go-muse/musex/tuning/temperament"
)

// Benchmarks cycle through varied inputs on every iteration (different presets, custom and large temperaments,
// steps across many octaves, varied cent values) so that the measured time and allocations describe the
// operation in general rather than one repeated input (NF-6).

// benchTemperaments returns all built-in presets followed by single-step, custom, and large temperaments.
func benchTemperaments() []temperament.Temperament {
	temps := append(temperament.Presets(), temperament.MustEDO(1), temperament.MustEDO(7))
	temps = append(temps, temperament.MustFromRatios("Tetrad", []temperament.Ratio{
		{Num: 1, Den: 1}, {Num: 5, Den: 4}, {Num: 3, Den: 2}, {Num: 15, Den: 8},
	}))
	return append(temps, largeTemperaments()...)
}

// benchTwelveToneTemperaments returns the 12-step temperaments of benchTemperaments.
func benchTwelveToneTemperaments() []temperament.Temperament {
	var out []temperament.Temperament
	for _, tm := range benchTemperaments() {
		if tm.StepsPerOctave() == 12 {
			out = append(out, tm)
		}
	}
	return out
}

// benchSteps lists signed cumulative steps spanning negative, zero, and positive octaves.
var benchSteps = []int{0, 1, 7, 11, 12, 13, 19, 25, 53, 99, 131, 1000, -1, -7, -12, -13, -19, -54, -133, -999}

// benchCents returns varied cent values, including exact ratios, comma fractions, and decimal values.
func benchCents() []cents.Cents {
	return []cents.Cents{
		cents.Zero(),
		cents.MustFromInt(50),
		cents.MustFromRatio(5, 4),
		cents.MustFromRatio(7, 4),
		cents.MustFromRatio(3, 2).MustSub(cents.SyntonicComma().MustDiv(4)),
		cents.MustFromFloat(1199.999999),
		cents.MustFromFloat(-10.5),
		cents.MustFromFloat(2450.123456),
		cents.MustFromInt(-1210),
		cents.PythagoreanComma().MustMul(37),
		cents.MustStep(5, 19),
		cents.MustStep(-200, 53),
	}
}

// circleFifths returns the 12 fifth sizes of tm along the circle C-G-D-A-E-B-F#-C#-G#-Eb-Bb-F-C.
func circleFifths(tm temperament.Temperament) []cents.Cents {
	out := make([]cents.Cents, 12)
	for i := range 12 {
		out[i] = tm.MustCyclicInterval((7*i)%12, (7*(i+1))%12)
	}
	return out
}

// circleDeviations returns the 12 fifth deviations from the pure fifth 3:2 of tm along the circle of fifths.
func circleDeviations(tm temperament.Temperament) []cents.Cents {
	pure := cents.MustFromRatio(3, 2)
	out := circleFifths(tm)
	for i := range out {
		out[i] = out[i].MustSub(pure)
	}
	return out
}

// BenchmarkPresetAccess measures retrieval of built-in presets, cycling through all preset functions.
func BenchmarkPresetAccess(b *testing.B) {
	groups := presetsWithAliases()
	var funcs []func() temperament.Temperament
	for _, group := range groups {
		funcs = append(funcs, group...)
	}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_ = funcs[i%len(funcs)]()
		i++
	}
}

// BenchmarkPresets measures building the slice of all built-in presets.
func BenchmarkPresets(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = temperament.Presets()
	}
}

// BenchmarkEDO measures construction of N-EDO temperaments for varied N in small and large size classes.
func BenchmarkEDO(b *testing.B) {
	for _, tc := range []struct {
		name string
		ns   []int
	}{
		{"small", []int{5, 7, 12, 17, 19, 22, 24, 31, 41, 53}},
		{"large", []int{72, 96, 171, 311, 665, 1200}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			i := 0
			for b.Loop() {
				_, _ = temperament.EDO(tc.ns[i%len(tc.ns)])
				i++
			}
		})
	}
}

// BenchmarkNew measures construction from cent tables of varied length and content.
func BenchmarkNew(b *testing.B) {
	var tables [][]cents.Cents
	for _, tm := range benchTemperaments() {
		tables = append(tables, tm.Offsets())
	}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temperament.New("Custom", tables[i%len(tables)])
		i++
	}
}

// BenchmarkFromFloats measures construction from float64 cent tables of varied length and content.
func BenchmarkFromFloats(b *testing.B) {
	var tables [][]float64
	for _, tm := range benchTemperaments() {
		offsets := tm.Offsets()
		floats := make([]float64, len(offsets))
		for m, c := range offsets {
			floats[m] = c.MustRoundTo(1_000_000).Float64()
		}
		tables = append(tables, floats)
	}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temperament.FromFloats("Custom", tables[i%len(tables)])
		i++
	}
}

// BenchmarkFromRatios measures construction from varied tables of 31-limit frequency ratios.
func BenchmarkFromRatios(b *testing.B) {
	tables := [][]temperament.Ratio{
		{{1, 1}, {16, 15}, {9, 8}, {6, 5}, {5, 4}, {4, 3}, {45, 32}, {3, 2}, {8, 5}, {5, 3}, {9, 5}, {15, 8}},
		{{1, 1}, {15, 14}, {8, 7}, {6, 5}, {5, 4}, {4, 3}, {7, 5}, {3, 2}, {8, 5}, {5, 3}, {7, 4}, {15, 8}},
		{{1, 1}, {256, 243}, {9, 8}, {32, 27}, {81, 64}, {4, 3}, {729, 512}, {3, 2}, {128, 81}, {27, 16}, {16, 9}, {243, 128}},
		{{1, 1}, {9, 8}, {5, 4}, {11, 8}, {3, 2}, {13, 8}, {7, 4}, {15, 8}},
		{{1, 1}, {31, 30}, {11, 9}, {13, 9}, {3, 2}, {62, 33}},
		{{1, 1}, {5, 4}, {3, 2}},
	}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temperament.FromRatios("Custom", tables[i%len(tables)])
		i++
	}
}

// BenchmarkFromFifths measures construction from 12 and 11 fifth sizes of varied 12-step temperaments.
func BenchmarkFromFifths(b *testing.B) {
	var fifths [][]cents.Cents
	for _, tm := range benchTwelveToneTemperaments() {
		all := circleFifths(tm)
		fifths = append(fifths, all, all[:11])
	}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temperament.FromFifths("Custom", fifths[i%len(fifths)])
		i++
	}
}

// BenchmarkFromFifthsWithStart measures construction from 11 fifth sizes with varied chain starts.
func BenchmarkFromFifthsWithStart(b *testing.B) {
	var fifths [][]cents.Cents
	for _, tm := range benchTwelveToneTemperaments() {
		fifths = append(fifths, circleFifths(tm)[:11])
	}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temperament.FromFifthsWithStart("Custom", fifths[i%len(fifths)], -(i % 12))
		i++
	}
}

// BenchmarkFromFifthDeviations measures construction from 12 and 11 fifth deviations of varied 12-step temperaments.
func BenchmarkFromFifthDeviations(b *testing.B) {
	var deviations [][]cents.Cents
	for _, tm := range benchTwelveToneTemperaments() {
		all := circleDeviations(tm)
		deviations = append(deviations, all, all[:11])
	}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temperament.FromFifthDeviations("Custom", deviations[i%len(deviations)])
		i++
	}
}

// BenchmarkFromFifthDeviationsWithStart measures construction from 11 fifth deviations with varied chain starts.
func BenchmarkFromFifthDeviationsWithStart(b *testing.B) {
	var deviations [][]cents.Cents
	for _, tm := range benchTwelveToneTemperaments() {
		deviations = append(deviations, circleDeviations(tm)[:11])
	}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temperament.FromFifthDeviationsWithStart("Custom", deviations[i%len(deviations)], -(i % 12))
		i++
	}
}

// BenchmarkFromGenerator measures regular temperament construction for varied N, generators, and chain starts.
func BenchmarkFromGenerator(b *testing.B) {
	generators := []cents.Cents{
		cents.MustFromRatio(3, 2),
		cents.MustFromRatio(3, 2).MustSub(cents.SyntonicComma().MustDiv(4)),
		cents.MustFromRatio(3, 2).MustSub(cents.SyntonicComma().MustScale(1, 3)),
		cents.MustFromRatio(5, 4),
	}
	ns := []int{5, 7, 12, 19, 31, 53}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		n := ns[i%len(ns)]
		_, _ = temperament.FromGenerator("Custom", n, generators[i%len(generators)], -(i % n))
		i++
	}
}

// BenchmarkNewPythagorean measures Pythagorean construction for every chain start kMin in [-11, 0].
func BenchmarkNewPythagorean(b *testing.B) {
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temperament.NewPythagorean(-(i % 12))
		i++
	}
}

// BenchmarkNewMeantone measures meantone construction for varied comma fractions and chain starts.
func BenchmarkNewMeantone(b *testing.B) {
	fractions := [][2]int{{1, 4}, {1, 3}, {1, 5}, {1, 6}, {2, 7}, {2, 9}, {3, 10}, {1, 11}, {-1, 12}, {0, 1}}
	b.Run("NewMeantone", func(b *testing.B) {
		b.ReportAllocs()
		i := 0
		for b.Loop() {
			f := fractions[i%len(fractions)]
			_, _ = temperament.NewMeantone(f[0], f[1])
			i++
		}
	})
	b.Run("NewMeantoneWithRange", func(b *testing.B) {
		b.ReportAllocs()
		i := 0
		for b.Loop() {
			f := fractions[i%len(fractions)]
			_, _ = temperament.NewMeantoneWithRange(f[0], f[1], -(i % 12))
			i++
		}
	})
}

// BenchmarkNewMeantoneFromFifth measures meantone construction for varied explicit fifth sizes.
func BenchmarkNewMeantoneFromFifth(b *testing.B) {
	var fifths []cents.Cents
	for c := 690; c <= 715; c += 5 {
		fifths = append(fifths, cents.MustFromInt(c), cents.MustFromFloat(float64(c)+0.123456))
	}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temperament.NewMeantoneFromFifthWithRange(fifths[i%len(fifths)], -(i % 12))
		i++
	}
}

// BenchmarkOffset measures intra-octave offset lookup across varied temperaments and steps.
func BenchmarkOffset(b *testing.B) {
	temps := benchTemperaments()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		tm := temps[i%len(temps)]
		_, _ = tm.Offset((i * 7) % tm.StepsPerOctave())
		i++
	}
}

// BenchmarkSplitStep measures step decomposition across varied temperaments and signed steps.
func BenchmarkSplitStep(b *testing.B) {
	temps := benchTemperaments()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _, _ = temps[i%len(temps)].SplitStep(benchSteps[i%len(benchSteps)])
		i++
	}
}

// BenchmarkCumulativeOffset measures multi-octave offsets across varied temperaments and signed steps.
func BenchmarkCumulativeOffset(b *testing.B) {
	temps := benchTemperaments()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temps[i%len(temps)].CumulativeOffset(benchSteps[i%len(benchSteps)])
		i++
	}
}

// BenchmarkDirectedInterval measures directed intervals across varied temperaments and step pairs.
func BenchmarkDirectedInterval(b *testing.B) {
	temps := benchTemperaments()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		from := benchSteps[i%len(benchSteps)]
		to := benchSteps[(i*7+3)%len(benchSteps)]
		_, _ = temps[i%len(temps)].DirectedInterval(from, to)
		i++
	}
}

// BenchmarkCyclicInterval measures cyclic intervals across varied temperaments and step pairs.
func BenchmarkCyclicInterval(b *testing.B) {
	temps := benchTemperaments()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		tm := temps[i%len(temps)]
		n := tm.StepsPerOctave()
		_, _ = tm.CyclicInterval((i*5)%n, (i*3+1)%n)
		i++
	}
}

// BenchmarkStepSizes measures elementary step sizes across varied temperaments.
func BenchmarkStepSizes(b *testing.B) {
	temps := benchTemperaments()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temps[i%len(temps)].StepSizes()
		i++
	}
}

// BenchmarkNearestStep measures nearest step search across varied temperaments and cent values.
func BenchmarkNearestStep(b *testing.B) {
	temps := benchTemperaments()
	xs := benchCents()
	b.Run("NearestStep", func(b *testing.B) {
		b.ReportAllocs()
		i := 0
		for b.Loop() {
			_, _, _ = temps[i%len(temps)].NearestStep(xs[i%len(xs)])
			i++
		}
	})
	b.Run("NearestCumulativeStep", func(b *testing.B) {
		b.ReportAllocs()
		i := 0
		for b.Loop() {
			_, _, _ = temps[i%len(temps)].NearestCumulativeStep(xs[i%len(xs)])
			i++
		}
	})
}

// BenchmarkRotate measures modal rotation across varied temperaments and signed rotation steps.
func BenchmarkRotate(b *testing.B) {
	temps := benchTemperaments()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temps[i%len(temps)].Rotate(benchSteps[i%len(benchSteps)])
		i++
	}
}

// BenchmarkEqual measures exact comparison across varied pairs, including equal and unequal tables.
func BenchmarkEqual(b *testing.B) {
	temps := benchTemperaments()
	copies := make([]temperament.Temperament, len(temps))
	for k, tm := range temps {
		copies[k] = temperament.MustNew("Copy", tm.Offsets())
	}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		k := i % len(temps)
		if i%2 == 0 {
			_ = temps[k].Equal(copies[k])
		} else {
			_ = temps[k].Equal(temps[(k+1)%len(temps)])
		}
		i++
	}
}

// BenchmarkNear measures approximate comparison across varied pairs and tolerances.
func BenchmarkNear(b *testing.B) {
	temps := benchTemperaments()
	tolerances := []cents.Cents{cents.Zero(), cents.MustFromFraction(1, 1000), cents.MustFromInt(2), cents.MustFromInt(30)}
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		k := i % len(temps)
		_ = temps[k].Near(temps[(k+i/len(temps))%len(temps)], tolerances[i%len(tolerances)])
		i++
	}
}

// BenchmarkIsEqual measures the equal-temperament predicate across varied temperaments.
func BenchmarkIsEqual(b *testing.B) {
	temps := benchTemperaments()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_ = temps[i%len(temps)].IsEqual()
		i++
	}
}

// BenchmarkIsRational measures the 31-limit rational predicate across varied temperaments.
func BenchmarkIsRational(b *testing.B) {
	temps := benchTemperaments()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_ = temps[i%len(temps)].IsRational()
		i++
	}
}

// BenchmarkAnalyzeEDO measures deviation analysis from N-EDO across varied temperaments.
func BenchmarkAnalyzeEDO(b *testing.B) {
	temps := benchTemperaments()
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		_, _ = temps[i%len(temps)].AnalyzeEDO()
		i++
	}
}

// BenchmarkAnalyzeIntervals measures cyclic interval analysis across varied temperaments, step sizes, and references.
func BenchmarkAnalyzeIntervals(b *testing.B) {
	var temps []temperament.Temperament
	for _, tm := range benchTemperaments() {
		if tm.StepsPerOctave() >= 2 {
			temps = append(temps, tm)
		}
	}
	threshold := cents.MustFromInt(15)
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		tm := temps[i%len(temps)]
		n := tm.StepsPerOctave()
		s := 1 + (i*7)%(n-1)
		_, _ = tm.AnalyzeIntervals(s, cents.MustStep(s, n), threshold)
		i++
	}
}

// BenchmarkAnalyzeFifths measures the 12-step fifth and third analyses across varied 12-step temperaments.
func BenchmarkAnalyzeFifths(b *testing.B) {
	temps := benchTwelveToneTemperaments()
	b.Run("AnalyzeFifths", func(b *testing.B) {
		b.ReportAllocs()
		i := 0
		for b.Loop() {
			_, _ = temps[i%len(temps)].AnalyzeFifths()
			i++
		}
	})
	b.Run("AnalyzeMajorThirds", func(b *testing.B) {
		b.ReportAllocs()
		i := 0
		for b.Loop() {
			_, _ = temps[i%len(temps)].AnalyzeMajorThirds()
			i++
		}
	})
	b.Run("AnalyzeMinorThirds", func(b *testing.B) {
		b.ReportAllocs()
		i := 0
		for b.Loop() {
			_, _ = temps[i%len(temps)].AnalyzeMinorThirds()
			i++
		}
	})
}
