package measure_test

import (
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/go-muse/musex/common/fraction"
	"github.com/go-muse/musex/measure"
)

// BenchmarkBars measures g.Bars(first, last) without iterating the sequence
// on the classes of architecture 6.2 of task 26.8. Before the loop each case
// is checked once against the model of the criterion and against the
// expected class: a sequence without an error, or Е-10 with
// fraction.ErrOutOfRange.
func BenchmarkBars(b *testing.B) {
	thirds := []*big.Rat{big.NewRat(1, 3), big.NewRat(2, 5), big.NewRat(3, 7)}

	// many-runs: 64 runs of two bars each; neighboring runs have lengths
	// with different denominators.
	var alternating []lcmLength

	for run := range 64 {
		bar := 1 + 2*run
		alternating = append(alternating, lcmRun(bar, bar+1, thirds[run%len(thirds)])...)
	}

	// wide: explicit lengths with denominators wider than 64 bits.
	// The runs are 2⁻¹⁰⁰, (3·2¹⁰⁰)⁻¹, and (5·2⁹⁰)⁻¹; their D are 2¹⁰⁰,
	// 3·2¹⁰⁰, and 15·2⁹⁶.
	wide := append(lcmRun(1, 4, lcmInv(pow2(100))), lcmRun(5, 8, lcmInv(new(big.Int).Mul(big.NewInt(3), pow2(100))))...)
	wide = append(wide, lcmRun(9, 12, lcmInv(new(big.Int).Mul(big.NewInt(5), pow2(90))))...)

	cases := []struct {
		name        string
		spec        lcmSpec
		first, last int
		refused     bool
	}{
		{"one-run", lcmSpec{first: lcmSimple(4, 4)}, 1, math.MaxInt, false},
		{"many-runs", lcmSpec{first: lcmSimple(4, 4), lengths: alternating}, 1, 128, false},
		{"wide", lcmSpec{first: lcmSimple(4, 4), lengths: wide}, 1, 12, false},
		{"out-of-range", lcmSpec{first: lcmSimple(1, 9), pickup: lcmInv(lcmP9())}, 1, 9, true},
	}

	for _, bc := range cases {
		b.Run(bc.name, func(b *testing.B) {
			g, model := lcmBuild(b, bc.spec, false)

			if got := model.refuses(bc.first, bc.last) != lcmAccepted; got != bc.refused {
				b.Fatalf("the model refuses Bars(%d, %d): %t, want %t", bc.first, bc.last, got, bc.refused)
			}

			if bc.name == "many-runs" {
				if n := len(model.runs(bc.first, bc.last)); n < 64 {
					b.Fatalf("Bars(%d, %d) has %d runs, want at least 64", bc.first, bc.last, n)
				}
			}

			seq, err := g.Bars(bc.first, bc.last)
			refused := errors.Is(err, measure.ErrOutOfRange) && errors.Is(err, fraction.ErrOutOfRange)

			switch {
			case bc.refused && !refused:
				b.Fatalf("Bars(%d, %d): %v, want Е-10 with fraction.ErrOutOfRange", bc.first, bc.last, err)
			case !bc.refused && (err != nil || seq == nil):
				b.Fatalf("Bars(%d, %d): %v, want a sequence", bc.first, bc.last, err)
			}

			b.ReportAllocs()

			for b.Loop() {
				_, _ = g.Bars(bc.first, bc.last)
			}
		})
	}
}
