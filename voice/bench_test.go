package voice_test

import (
	"testing"

	"github.com/go-muse/musex/note"
	"github.com/go-muse/musex/voice"
)

// Sinks keep benchmark results alive. Benchmarks report allocations and
// assert no thresholds.
var (
	sinkTessitura voice.Tessitura
	sinkVoiceErr  error
)

// BenchmarkTessitura measures the by-length tessitura over a line of 64
// sounds within the soprano range.
func BenchmarkTessitura(b *testing.B) {
	pitches := []note.Note{c4(), d4(), e4(), f4(), g4()}
	lengths := []voice.Duration{q(), e8(), h(), dq()}
	elements := make([]voice.Element, 64)
	for i := range elements {
		elements[i] = snd(pitches[i%len(pitches)], lengths[i%len(lengths)])
	}
	l := line(elements...)
	ref := voice.PresetReference(soprano())
	if _, err := l.TessituraBy(ref, voice.ByLength, voice.DefaultZones()); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		sinkTessitura, sinkVoiceErr = l.TessituraBy(ref, voice.ByLength, voice.DefaultZones())
	}
}
