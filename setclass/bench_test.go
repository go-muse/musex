package setclass_test

import (
	"testing"

	"github.com/go-muse/musex/pcset"
	"github.com/go-muse/musex/setclass"
)

func BenchmarkNewMod12(b *testing.B) {
	s := pcset.MustNew(12, 0, 1, 3, 7, 8)
	for _, kind := range setclass.Kinds() {
		b.Run(kind.String(), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, _ = setclass.New(s, kind)
			}
		})
	}
}

func BenchmarkCollinsMaskMod12(b *testing.B) {
	sc := setclass.MustFromForteNotation("6-Z29", setclass.TnIn)
	b.ReportAllocs()
	for b.Loop() {
		_, _ = sc.CollinsMask()
	}
}

func BenchmarkCatalogBridgeMod12(b *testing.B) {
	sc := setclass.MustFromForteNotation("4-Z15", setclass.TnIn)
	b.ReportAllocs()
	for b.Loop() {
		_, _ = sc.ForteClass()
		_, _, _ = sc.ZPartner()
		_, _ = sc.CatalogName()
	}
}

func BenchmarkEmbeddingCountMod12(b *testing.B) {
	sub := setclass.MustFromForteNotation("3-11", setclass.TnIn)
	super := setclass.MustFromForteNotation("7-35", setclass.TnIn)
	b.ReportAllocs()
	for b.Loop() {
		_, _ = sub.EmbeddingCount(super)
	}
}

func BenchmarkSubclassProfileMod12(b *testing.B) {
	sc := setclass.MustFromForteNotation("7-35", setclass.TnIn)
	b.ReportAllocs()
	for b.Loop() {
		_, _ = sc.SubclassProfile()
	}
}

func BenchmarkBurnsideCount(b *testing.B) {
	b.Run("Mod24_TnIn", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = setclass.Count(24, setclass.TnIn)
		}
	})

	b.Run("Mod24_TTO", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = setclass.Count(24, setclass.TTO)
		}
	})

	b.Run("Mod1024_TnIn_Big", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = setclass.CountBig(1_024, setclass.TnIn)
		}
	})
}

func BenchmarkAllMod12(b *testing.B) {
	for _, kind := range setclass.Kinds() {
		b.Run(kind.String(), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, _ = setclass.All(12, kind)
			}
		})
	}
}
