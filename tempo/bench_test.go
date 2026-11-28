package tempo_test

import (
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
)

// Sinks keep benchmark results alive. Benchmarks report allocations and
// assert no thresholds.
var (
	sinkTempo    tempo.Tempo
	sinkSeconds  tempo.Seconds
	sinkFraction fraction.Fraction
	sinkDuration duration.Duration
	sinkErr      error
)

func BenchmarkDurationIn(b *testing.B) {
	mark := tm(dq, fr(185, 2))
	s := sec(90, 1)
	if _, err := mark.DurationIn(s); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		sinkDuration, sinkErr = mark.DurationIn(s)
	}
}

func BenchmarkMul(b *testing.B) {
	mark := tq(120)
	k := fr(3, 2)
	if _, err := mark.Mul(k); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		sinkTempo, sinkErr = mark.Mul(k)
	}
}

func BenchmarkSpeed(b *testing.B) {
	mark := tm(dq, fr(185, 2))
	if _, err := mark.Speed(); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		sinkFraction, sinkErr = mark.Speed()
	}
}

func BenchmarkQuartersPerMinute(b *testing.B) {
	mark := tm(dq, fr(185, 2))
	if _, err := mark.QuartersPerMinute(); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		sinkFraction, sinkErr = mark.QuartersPerMinute()
	}
}

func BenchmarkBarLength(b *testing.B) {
	unit := dur(3, 8)
	r := fr(7, 4)
	if _, err := tempo.BarLength(unit, r); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		sinkDuration, sinkErr = tempo.BarLength(unit, r)
	}
}

func BenchmarkSecondsFromMinutes(b *testing.B) {
	f := fr(5, 2)
	if _, err := tempo.SecondsFromMinutes(f); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		sinkSeconds, sinkErr = tempo.SecondsFromMinutes(f)
	}
}

func BenchmarkSecondsAdd(b *testing.B) {
	x, y := sec(5, 7), sec(3, 11)
	if _, err := x.Add(y); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		sinkSeconds, sinkErr = x.Add(y)
	}
}

func BenchmarkSecondsSub(b *testing.B) {
	x, y := sec(5, 7), sec(3, 11)
	if _, err := x.Sub(y); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		sinkSeconds, sinkErr = x.Sub(y)
	}
}
