package settings_test

import (
	"testing"

	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"
)

// sinkString keeps formatted results alive.
var sinkString string

// BenchmarkCheckReplacement covers NF-3: the standard set and the set with
// the largest built-in name domain (unicode-wyschnegradsky, 175 names).
func BenchmarkCheckReplacement(b *testing.B) {
	std := settings.Standard()
	wy := setIn(b, system.ManeriSims72(), builtin(notation.UnicodeWyschnegradskyID))
	for _, bc := range []struct {
		name       string
		prev, next *settings.Settings
	}{
		{name: "standard", prev: std, next: settings.Standard()},
		{name: "standard to quarter tones", prev: std, next: std.MustWithTuning(tuningFor(b, system.QuarterTones24()))},
		{name: "quarter tones to standard", prev: std.MustWithTuning(tuningFor(b, system.QuarterTones24())), next: std},
		{name: "wyschnegradsky", prev: wy, next: setIn(b, system.ManeriSims72(), builtin(notation.UnicodeWyschnegradskyID))},
		{name: "wyschnegradsky to traditional", prev: wy, next: setIn(b, system.Traditional(), builtin(notation.UnicodeWyschnegradskyID))},
	} {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkBool = bc.prev.MustCheckReplacement(bc.next).Compatible()
			}
		})
	}
}

// BenchmarkStandard covers NF-3: getting the standard settings.
func BenchmarkStandard(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkSettings = settings.Standard()
	}
}

// BenchmarkEqual covers NF-3: comparing equal and unequal sets.
func BenchmarkEqual(b *testing.B) {
	std := settings.Standard()
	for _, bc := range []struct {
		name  string
		other *settings.Settings
	}{
		{name: "equal", other: settings.Standard()},
		{name: "other notation", other: std.MustWithNotation(builtin(notation.GermanAMID))},
	} {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkBool = std.Equal(bc.other)
			}
		})
	}
}

// BenchmarkExpressesPitch covers NF-3 for several octave styles.
func BenchmarkExpressesPitch(b *testing.B) {
	std := settings.Standard()
	p := at(spelling.CSharp(), 4)
	for _, bc := range []struct {
		name string
		set  *settings.Settings
	}{
		{name: "visible number", set: std},
		{name: "AM", set: std.MustWithNotation(builtin(notation.GermanAMID))},
		{name: "LilyPond", set: std.MustWithNotation(builtin(notation.LilyPondEnglishID))},
		{name: "name", set: std.MustWithNotation(englishNamed(b))},
	} {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkBool = bc.set.MustExpressesPitch(p)
			}
		})
	}
}

// BenchmarkFormatPitch covers NF-3: formatting a pitch.
func BenchmarkFormatPitch(b *testing.B) {
	std := settings.Standard()
	p := at(spelling.CSharp(), 4)
	for _, bc := range []struct {
		name string
		set  *settings.Settings
	}{
		{name: "visible number", set: std},
		{name: "AM", set: std.MustWithNotation(builtin(notation.GermanAMID))},
	} {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkString = bc.set.MustFormatPitch(p)
			}
		})
	}
}
