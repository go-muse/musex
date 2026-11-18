package search_test

import (
	"testing"

	"github.com/go-muse/musex/alteration"
	"github.com/go-muse/musex/search"
	"github.com/go-muse/musex/system"
)

var (
	benchmarkSpellingMatches []search.SpellingMatch
	benchmarkPositionMatches []search.PositionMatch
)

func BenchmarkBuiltInCatalog(b *testing.B) {
	searcher := search.BuiltIn()
	input := notes("D", "E", "F♯", "A", "B")
	b.Run("spelling", func(b *testing.B) {
		var err error
		for b.Loop() {
			benchmarkSpellingMatches, err = searcher.BySpelling(input, search.Containment)
		}
		if err != nil {
			b.Fatal(err)
		}
	})
	b.Run("position", func(b *testing.B) {
		var err error
		for b.Loop() {
			benchmarkPositionMatches, err = searcher.ByPosition(system.Traditional(), input, search.Containment)
		}
		if err != nil {
			b.Fatal(err)
		}
	})
}

func BenchmarkNonRegularSystem(b *testing.B) {
	searcher := builtInSearcher(b, "ionian")
	input := notes("E", "F♯")
	sys := nonRegularSystem()
	var err error
	for b.Loop() {
		benchmarkPositionMatches, err = searcher.ByPosition(sys, input, search.Containment)
	}
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkComplexCustomRule(b *testing.B) {
	first := system.MustNewProgression(alt(-4), fractional(1, 2)).MustWithUpper(alt(4))
	second := system.MustNewProgression(fractional(-7, 2), alt(1)).MustWithUpper(fractional(7, 2))
	rule := system.MustNewRule(first, second, system.MustNewProgression(alteration.Natural(), alteration.Natural()))
	sys := customSystem("complex-rule", 24, 2, naturalOffsets(), rule)
	searcher := builtInSearcher(b, "ionian")
	input := notes("C", "E", "G")
	var err error
	for b.Loop() {
		benchmarkPositionMatches, err = searcher.ByPosition(sys, input, search.Containment)
	}
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkLargeFractionalRule(b *testing.B) {
	sys := onePlaceSystem("large-fractional-benchmark", 1, alteration.MaxDenominator, system.MustMultiples(alteration.MaxDenominator))
	searcher := searcherFor(b, onePlaceEntry("one"))
	input := notes("C")
	var err error
	for b.Loop() {
		benchmarkPositionMatches, err = searcher.ByPosition(sys, input, search.FullMatch)
	}
	if err != nil {
		b.Fatal(err)
	}
}
