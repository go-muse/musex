package tempomap_test

import (
	"fmt"
	"testing"

	"github.com/go-muse/musex/common/rounding"
	"github.com/go-muse/musex/duration"
	"github.com/go-muse/musex/tempo"
	"github.com/go-muse/musex/tempomap"
)

var (
	sinkMap   tempomap.Map
	sinkTempo tempo.Tempo
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

func BenchmarkMapBuildWithChange(b *testing.B) {
	for _, count := range []int{1, 16, 64} {
		b.Run(fmt.Sprintf("changes=%d", count), func(b *testing.B) {
			changes := make([]tempomap.Change, 0, count)
			for i := 1; i <= count; i++ {
				changes = append(changes, tempomap.MustNewMain(at(i, 1), quarter(60+i%120)))
			}
			initial := quarter(60)
			b.ReportAllocs()
			for b.Loop() {
				m, err := tempomap.New(initial)
				if err != nil {
					b.Fatal(err)
				}
				for _, change := range changes {
					if m, err = m.WithChange(change); err != nil {
						b.Fatal(err)
					}
				}
				sinkMap = m
			}
		})
	}
}

func BenchmarkMapBuildWithExtension(b *testing.B) {
	for _, count := range []int{1, 16, 64} {
		b.Run(fmt.Sprintf("extensions=%d", count), func(b *testing.B) {
			extensions := make([]tempomap.Extension, 0, count)
			for i := 1; i <= count; i++ {
				extensions = append(extensions, tempomap.MustNewExtension(at(i, 1), sec(1, 1000)))
			}
			initial := quarter(60)
			b.ReportAllocs()
			for b.Loop() {
				m, err := tempomap.New(initial)
				if err != nil {
					b.Fatal(err)
				}
				for _, extension := range extensions {
					if m, err = m.WithExtension(extension); err != nil {
						b.Fatal(err)
					}
				}
				sinkMap = m
			}
		})
	}
}

func BenchmarkMapTempoAt(b *testing.B) {
	unit := duration.Quarter.MustWritten()
	steady := tempomap.MustNew(quarter(60))
	for i := 1; i < 64; i++ {
		steady = steady.MustWithChange(tempomap.MustNewMain(at(i, 1), quarter(60+i%120)))
	}
	cases := []struct {
		name    string
		m       tempomap.Map
		address tempomap.Address
	}{
		{name: "steady/changes=64", m: steady, address: at(63, 1)},
		{name: "gradual", m: newGradualMap(pos(1, 1), quarter(120)), address: at(1, 3)},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				got, err := tc.m.TempoAt(tc.address, unit)
				if err != nil {
					b.Fatal(err)
				}
				sinkTempo = got
			}
		})
	}
}
