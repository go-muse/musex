package duration_test

import (
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tuplet"
)

// Sinks keep benchmark results alive. Benchmarks report allocations and
// assert no thresholds (architecture 6).
var (
	sinkDuration duration.Duration
	sinkPosition duration.Position
	sinkWritten  duration.Written
	sinkParts    []duration.Written
	sinkString   string
	sinkErr      error
)

func BenchmarkAdd(b *testing.B) {
	x, y := dur(3, 8), dur(1, 12)

	b.ReportAllocs()

	for b.Loop() {
		sinkDuration, sinkErr = x.Add(y)
	}
}

func BenchmarkSum16(b *testing.B) {
	values := make([]duration.Duration, 16)
	for i := range values {
		values[i] = dur(1, i+1)
	}

	b.ReportAllocs()

	for b.Loop() {
		sinkDuration, sinkErr = duration.Sum(values...)
	}
}

func BenchmarkNewWritten(b *testing.B) {
	triplet := tuplet.Triplet()

	b.ReportAllocs()

	for b.Loop() {
		sinkWritten, sinkErr = duration.NewWritten(duration.Quarter, 1, triplet)
	}
}

func BenchmarkNotate(b *testing.B) {
	for _, c := range []struct {
		name string
		v    duration.Duration
	}{
		{"typical", dur(11, 16)},
		{"1024 parts", dur(2048, 1)},
		{"too many parts", durRat(pow2(200))},
	} {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				sinkParts, sinkErr = c.v.Notate()
			}
		})
	}
}

func BenchmarkNotateWith(b *testing.B) {
	v := dur(11, 16)

	b.ReportAllocs()

	for b.Loop() {
		sinkParts, sinkErr = v.NotateWith(duration.Whole, 2, tuplet.Triplet())
	}
}

func BenchmarkToWritten(b *testing.B) {
	v := dur(1, 3)
	triplet := tuplet.Triplet()

	b.ReportAllocs()

	for b.Loop() {
		sinkWritten, sinkErr = v.ToWritten(triplet)
	}
}

func BenchmarkQuantize(b *testing.B) {
	p := pos(7, 24)
	step := frac(1, 8)

	b.ReportAllocs()

	for b.Loop() {
		sinkPosition, sinkErr = p.Quantize(step, rounding.HalfEven)
	}
}

func BenchmarkParseWritten(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		sinkWritten, sinkErr = duration.ParseWritten("4.[5:4,3:2]")
	}
}

func BenchmarkWrittenString(b *testing.B) {
	w := wr(duration.Quarter, 1, tuplet.Quintuplet(), tuplet.Triplet())

	b.ReportAllocs()

	for b.Loop() {
		sinkString = w.String()
	}
}
