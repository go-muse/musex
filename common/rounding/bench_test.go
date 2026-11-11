package rounding_test

import (
	"testing"

	"github.com/go-muse/musex/common/rounding"
)

var (
	benchNeighbor rounding.Neighbor
	benchMode     rounding.Mode
	benchInt      int
	benchInt2     int
	benchFloat    float64
	benchString   string
	benchErr      error
)

func BenchmarkChoose(b *testing.B) {
	b.Run("half_even", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchNeighbor, benchErr = rounding.HalfEven.Choose(0, rounding.Positive, false)
		}
	})
}

func BenchmarkChooseIndex(b *testing.B) {
	b.Run("half_even", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchInt, benchErr = rounding.HalfEven.ChooseIndex(3, 0)
		}
	})
}

func BenchmarkQuoRem(b *testing.B) {
	b.Run("half_even", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchInt, benchInt2, benchErr = rounding.HalfEven.QuoRem(7, 2)
		}
	})
}

func BenchmarkFloat(b *testing.B) {
	b.Run("half_even", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchFloat, benchErr = rounding.HalfEven.Float(2.5)
		}
	})
}

func BenchmarkFloatInt(b *testing.B) {
	b.Run("half_even", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchInt, benchErr = rounding.HalfEven.FloatInt(2.5)
		}
	})
}

func BenchmarkFloatMultiple(b *testing.B) {
	b.Run("half_up", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchFloat, benchErr = rounding.HalfUp.FloatMultiple(0.15, 0.1)
		}
	})
}

func BenchmarkFloatDecimal(b *testing.B) {
	b.Run("half_up", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchString, benchErr = rounding.HalfUp.FloatDecimal(0.145, 2)
		}
	})
}

func BenchmarkRatDecimal(b *testing.B) {
	b.Run("half_even", func(b *testing.B) {
		b.ReportAllocs()
		x := bigRat(1, 8)
		for b.Loop() {
			benchString, benchErr = rounding.HalfEven.RatDecimal(x, 2)
		}
	})
}

func BenchmarkParse(b *testing.B) {
	b.Run("half_even", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchMode, benchErr = rounding.Parse("half-even")
		}
	})
}
