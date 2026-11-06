package pcset_test

import (
	"testing"

	"github.com/go-muse/musex/pcset"
)

func BenchmarkSetOperationsMod12(b *testing.B) {
	cMaj := pcset.MustNew(12, 0, 4, 7)
	majorScale := pcset.MustNew(12, 0, 2, 4, 5, 7, 9, 11)

	b.Run("New", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = pcset.New(12, 0, 2, 4, 5, 7, 9, 11)
		}
	})

	b.Run("Transpose", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = majorScale.Transpose(5)
		}
	})

	b.Run("Union", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = cMaj.Union(majorScale)
		}
	})

	b.Run("CoveringTranspositionsOf", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = cMaj.CoveringTranspositionsOf(majorScale)
		}
	})

	b.Run("NormalForm_Rahn", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = majorScale.NormalForm()
		}
	})

	b.Run("NormalForm_Forte", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = majorScale.NormalFormWith(pcset.Forte)
		}
	})

	b.Run("IntervalVector", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = majorScale.IntervalVector()
		}
	})

	b.Run("InvarianceVector", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = majorScale.InvarianceVector()
		}
	})
}

func BenchmarkIntervalVectorMaxModulusUniverse(b *testing.B) {
	univ := pcset.MustUniverse(pcset.MaxModulus)

	b.ReportAllocs()
	for b.Loop() {
		_ = univ.IntervalVector()
	}
}
