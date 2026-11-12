package frequency_test

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/go-muse/musex/frequency"
)

var (
	ratioSink       *big.Rat
	frequenciesSink []float64
)

func BenchmarkFrequency(b *testing.B) {
	benchmark := func(name string, call func()) {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				call()
			}
		})
	}

	strict := frequency.Strict()
	withinCents := centsCriterion(5)
	withinHertz := frequency.MustWithinHertz(1)
	withinRelative := frequency.MustWithinRelative(0.01)
	composite := frequency.MustAnyOf(
		frequency.Strict(),
		withinCents,
		withinHertz,
		withinRelative,
	)
	text := "any(cents(1200/19 cents), hertz(0.5))"

	benchmark("Equal", func() { boolSink = frequency.MustEqual(440, 441) })
	benchmark("Strict.Near", func() { boolSink = strict.MustNear(440, 441) })
	benchmark("Cents.Near", func() { boolSink = withinCents.MustNear(440, 441) })
	benchmark("Hertz.Near", func() { boolSink = withinHertz.MustNear(440, 441) })
	benchmark("Relative.Near", func() { boolSink = withinRelative.MustNear(440, 441) })
	benchmark("Composite.Near", func() { boolSink = composite.MustNear(440, 441) })
	benchmark("Parse", func() { criterionSink = frequency.MustParse(text) })
	benchmark("String", func() { stringSink = composite.String() })

	benchmark("Transpose", func() { floatSink = frequency.MustTranspose(440, 7, 12) })
	benchmark("Ratio", func() { ratioSink = frequency.MustRatio(440, 660) })
	benchmark("CentsBetween", func() { centsSink = frequency.MustCentsBetween(440, 442) })
	benchmark("Partial", func() { floatSink = frequency.MustPartial(110, 2) })
	benchmark("Subharmonic", func() { floatSink = frequency.MustSubharmonic(440, 2) })
	benchmark("BeatFrequency", func() { floatSink = frequency.MustBeatFrequency(440, 442) })
	benchmark("ShiftOctaves", func() { floatSink = frequency.MustShiftOctaves(440, 1) })
	benchmark("NormalizeOctave", func() { floatSink, intSink = frequency.MustNormalizeOctave(880, 110) })
	benchmark("Period", func() { floatSink = frequency.MustPeriod(440) })

	for _, size := range []int{8, 1024} {
		indices := make([]int, size)
		for i := range indices {
			indices[i] = i%32 + 1
		}
		benchmark("Partials_m"+strconv.Itoa(size), func() {
			frequenciesSink = frequency.MustPartials(110, indices)
		})
		benchmark("Subharmonics_m"+strconv.Itoa(size), func() {
			frequenciesSink = frequency.MustSubharmonics(440, indices)
		})
	}
}
