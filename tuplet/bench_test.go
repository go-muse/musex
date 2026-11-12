package tuplet_test

import (
	"math"
	"testing"

	"github.com/go-muse/musex/tuplet"
)

// benchTuplets are the inputs of the benchmarks: small numbers and numbers
// near math.MaxInt. Benchmarks report allocations and assert no thresholds.
var benchTuplets = []struct {
	name string
	tp   tuplet.Tuplet
}{
	{name: "small", tp: tuplet.MustNew(6, 4)},
	{name: "large", tp: tuplet.MustNew(math.MaxInt, math.MaxInt-1)},
}

func BenchmarkNew(b *testing.B) {
	for _, bt := range benchTuplets {
		actual, normal := bt.tp.MustPair()

		b.Run(bt.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				sinkTuplet, sinkErr = tuplet.New(actual, normal)
			}
		})
	}
}

func BenchmarkDefault(b *testing.B) {
	for _, bt := range benchTuplets {
		actual := bt.tp.MustActual()

		b.Run(bt.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				sinkTuplet, sinkErr = tuplet.Default(actual, tuplet.Binary)
			}
		})
	}
}

func BenchmarkCoefficient(b *testing.B) {
	for _, bt := range benchTuplets {
		tp := bt.tp

		b.Run(bt.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				sinkFraction, sinkErr = tp.Coefficient()
			}
		})
	}
}

func BenchmarkCompare(b *testing.B) {
	for _, bt := range benchTuplets {
		tp := bt.tp
		other := tuplet.Triplet()

		b.Run(bt.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				sinkInt = tp.Compare(other)
			}
		})
	}
}

func BenchmarkApply(b *testing.B) {
	d := frac(3, 16)

	for _, bt := range benchTuplets {
		tp := bt.tp

		b.Run(bt.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				sinkFraction, sinkErr = tp.Apply(d)
			}
		})
	}
}

func BenchmarkChainCoefficient(b *testing.B) {
	for _, bt := range benchTuplets {
		chain := []tuplet.Tuplet{bt.tp, tuplet.Triplet(), tuplet.Quintuplet()}

		b.Run(bt.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				sinkFraction, sinkErr = tuplet.ChainCoefficient(chain...)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	for _, bt := range benchTuplets {
		text := bt.tp.String()

		b.Run(bt.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				sinkTuplet, sinkErr = tuplet.Parse(text)
			}
		})
	}
}

func BenchmarkString(b *testing.B) {
	for _, bt := range benchTuplets {
		tp := bt.tp

		b.Run(bt.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				sinkString = tp.String()
			}
		})
	}
}
