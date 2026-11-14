package note_test

import (
	"testing"

	"github.com/go-muse/musex/frequency"
	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/tuning"
)

// Benchmarks of the hot path and the costs of NF-3 and R-N-6, without
// thresholds (plan, Sh1 "Benchmarks").

var (
	benchFloat float64
	benchInt   int
	benchBool  bool
	benchNotes []note.Note
)

func BenchmarkFrequency(b *testing.B) {
	s := std()
	n := note.MustFromPitch(pt(spelling.CSharp(), 4), ct(30), s)
	tu := tuning.Standard()

	b.Run("own source", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchFloat, _ = n.Frequency()
		}
	})
	b.Run("explicit tuning", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchFloat, _ = n.FrequencyInTuning(tu)
		}
	})
}

func BenchmarkEqual(b *testing.B) {
	x := note.MustFromPitch(pt(spelling.CSharp(), 4), ct(30), std())
	y := note.MustFromPitch(pt(spelling.CSharp(), 4), ct(30), nil)
	b.ReportAllocs()
	for b.Loop() {
		benchBool = x.Equal(y)
	}
}

func BenchmarkReadParts(b *testing.B) {
	n := note.MustFromPitch(pt(spelling.CSharp(), 4), ct(30), std())
	b.ReportAllocs()
	for b.Loop() {
		benchBool = n.IsValid() && n.HasOctave() && n.IsNatural()
		_ = n.Spelling()
		_ = n.Cents()
		benchInt, _ = n.ASPN()
	}
}

func BenchmarkCompareSound(b *testing.B) {
	s := std()
	x := note.MustFromPitch(pt(spelling.C(), 4), ct(150), s)
	y := note.MustFromPitch(pt(spelling.CSharp(), 4), ct(0), s)
	z := note.MustFromPitch(pt(spelling.CSharp(), 4), ct(0), set19())

	b.Run("exact path", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchInt, _ = x.CompareSound(y)
		}
	})
	b.Run("frequency path", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchInt, _ = x.CompareSound(z)
		}
	})
	b.Run("criterion", func(b *testing.B) {
		c := frequency.MustWithinCents(ct(5))
		b.ReportAllocs()
		for b.Loop() {
			benchInt, _ = x.CompareSoundBy(y, c)
		}
	})
}

func BenchmarkSortedBySound(b *testing.B) {
	s := std()
	seq := note.MustFromPitch(pt(spelling.C(), 0), ct(0), s).
		MustWalk(note.MustFromPitch(pt(spelling.C(), 9), ct(0), s), system.Sharps())
	var notes []note.Note
	for len(notes) < 1000 {
		for n := range seq {
			notes = append(notes, n)
		}
	}
	notes = notes[:1000]
	for i, j := 0, len(notes)-1; i < j; i, j = i+1, j-1 {
		notes[i], notes[j] = notes[j], notes[i]
	}
	b.ReportAllocs()
	for b.Loop() {
		benchNotes, _ = note.SortedBySound(notes)
	}
}
