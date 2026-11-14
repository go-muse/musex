package instrument_test

import (
	"testing"

	"github.com/go-muse/musex/instrument"
	"github.com/go-muse/musex/settings"
)

// А 8: the hot paths, a retuning and the built-in catalog.

func BenchmarkInstrument_Settings(b *testing.B) {
	g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
	var src settings.Source = g
	b.ReportAllocs()
	for b.Loop() {
		_ = src.Settings()
	}
}

func BenchmarkInstrument_Contains(b *testing.B) {
	g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
	x := p("E4")
	b.ReportAllocs()
	for b.Loop() {
		if _, err := g.Contains(x); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRange_Contains(b *testing.B) {
	r := rng("C4", "G5")
	sys := trad()
	x := p("E4")
	b.ReportAllocs()
	for b.Loop() {
		if _, err := r.Contains(sys, x); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInstrument_Retune(b *testing.B) {
	g := instrument.MustFromParts("Lead", h0(), rng("C4", "G5"))
	sets := [2]*settings.Settings{h415(), h0()}
	i := 0
	b.ReportAllocs()
	for b.Loop() {
		if err := g.Retune(sets[i%2]); err != nil {
			b.Fatal(err)
		}
		i++
	}
}

func BenchmarkBuiltins(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if instrument.Builtins().Len() == 0 {
			b.Fatal("empty built-in catalog")
		}
	}
}
