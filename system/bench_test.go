package system_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/octave"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/pitchclass"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// Benchmarks: hot paths are expected to report 0 allocs/op (contract 6.2);
// the worst case of SubsetOf checks the estimate of contract 6.3 (risk R-7).

// Typed sinks avoid allocations from conversion to interfaces.
var (
	sinkBool   bool
	sinkInt    int
	sinkClass  pitchclass.Class
	sinkStep   system.Step
	sinkOctave octave.Octave
	sinkPitch  pitch.Pitch
	sinkErr    error
)

func BenchmarkHotPaths(b *testing.B) {
	s := system.Traditional()
	name := spelling.BSharp()
	x, y := pt(spelling.BSharp(), 4), pt(spelling.EFlat(), 5)
	w := o(3)
	r := system.Integers()
	v := alteration.Sharp()

	b.Run("Class", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkClass, sinkErr = s.Class(name)
		}
	})
	b.Run("Admits", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBool, sinkErr = s.Admits(name)
		}
	})
	b.Run("Carry", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkInt, sinkErr = s.Carry(name)
		}
	})
	b.Run("PitchClass", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkClass, sinkErr = s.PitchClass(x)
		}
	})
	b.Run("Compare", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkInt, sinkErr = s.Compare(x, y)
		}
	})
	b.Run("Distance", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkStep, sinkErr = s.Distance(x, y)
		}
	})
	b.Run("Equivalent", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBool, sinkErr = s.Equivalent(x, y)
		}
	})
	b.Run("WrittenOctave", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkOctave, sinkErr = s.WrittenOctave(x)
		}
	})
	b.Run("FromWritten", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkPitch, sinkErr = s.FromWritten(name, w)
		}
	})
	b.Run("Rule.Contains", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBool, sinkErr = r.Contains(v)
		}
	})
}

// worstRules returns two equal rules of MaxProgressions progressions with steps
// 1/d for d near MaxDenominator, in different orders.
func worstRules(b *testing.B) (*system.Rule, *system.Rule) {
	b.Helper()
	ps := make([]system.Progression, 0, system.MaxProgressions)
	for i := range system.MaxProgressions {
		p, err := system.NewProgression(a(0), q(1, system.MaxDenominator-i))
		if err != nil {
			b.Fatal(err)
		}
		ps = append(ps, p)
	}
	x, err := system.NewRule(ps...)
	if err != nil {
		b.Fatal(err)
	}
	rev := make([]system.Progression, len(ps))
	for i := range ps {
		rev[i] = ps[len(ps)-1-i]
	}
	y, err := system.NewRule(rev...)
	if err != nil {
		b.Fatal(err)
	}
	return x, y
}

func BenchmarkSubsetOfWorstCase(b *testing.B) {
	x, y := worstRules(b)
	b.ReportAllocs()
	for b.Loop() {
		sinkBool, sinkErr = x.SubsetOf(y)
	}
	if !sinkBool || sinkErr != nil {
		b.Fatalf("SubsetOf = %v, %v; want true, nil", sinkBool, sinkErr)
	}
}

func BenchmarkSubsetOfTypical(b *testing.B) {
	x := system.Integers()
	y := system.MustUnion(system.MustMultiples(2), system.Single(q(1, 4)))
	b.ReportAllocs()
	for b.Loop() {
		sinkBool, sinkErr = x.SubsetOf(y)
	}
}
