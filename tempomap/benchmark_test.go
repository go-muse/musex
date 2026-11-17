package tempomap_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/tempomap"
)

func BenchmarkMapStartByChangeCount(b *testing.B) {
	for _, count := range []int{1, 16, 64, 256} {
		b.Run(fmt.Sprintf("changes=%d", count), func(b *testing.B) {
			m := tempomap.MustNew(quarter(60))
			for i := 1; i < count; i++ {
				m = m.MustWithChange(tempomap.MustNewMain(at(i, 1), quarter(60+i%120)))
			}
			address := at(count-1, 1)
			b.ReportAllocs()
			for b.Loop() {
				_ = m.MustStart(address, 3, rounding.HalfUp)
			}
		})
	}
}

func BenchmarkMapPositionByExtensionCount(b *testing.B) {
	for _, count := range []int{0, 16, 64, 256} {
		b.Run(fmt.Sprintf("extensions=%d", count), func(b *testing.B) {
			m := tempomap.MustNew(quarter(60))
			for i := range count {
				m = m.MustWithExtension(tempomap.MustNewExtension(at(i+1, 1), sec(1, 1000)))
			}
			b.ReportAllocs()
			for b.Loop() {
				_ = m.MustPosition(sec(1, 1), 3, rounding.HalfUp)
			}
		})
	}
}

func BenchmarkMapStartByPlaces(b *testing.B) {
	m := newGradualMap(pos(1, 1), quarter(120))
	address := at(1, 1)
	for _, places := range []int{0, 3, 12, 40, 76} {
		b.Run(fmt.Sprintf("places=%d", places), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = m.MustStart(address, places, rounding.HalfUp)
			}
		})
	}
}
