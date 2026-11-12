package notation_test

import (
	"strings"
	"testing"

	"github.com/go-muse/musex/letter"
	"github.com/go-muse/musex/notation"
	"github.com/go-muse/musex/octave/number"
	"github.com/go-muse/musex/pitch"
	"github.com/go-muse/musex/spelling"
	"github.com/go-muse/musex/system"

	"github.com/stretchr/testify/require"
)

var (
	benchmarkNotation notation.Notation
	benchmarkString   string
	benchmarkStrings  []string
	benchmarkName     spelling.Spelling
	benchmarkNames    []spelling.Spelling
	benchmarkPitch    pitch.Pitch
	benchmarkBool     bool
	benchmarkErr      error
)

func requireZeroAllocs(b *testing.B, fn func()) {
	b.Helper()
	if got := testing.AllocsPerRun(100, fn); got != 0 {
		b.Fatalf("got %v allocations per operation, want 0", got)
	}
}

func BenchmarkNotationIsDefined(b *testing.B) {
	n := notation.MustBuiltIn(notation.EnglishUnicodeID)
	requireZeroAllocs(b, func() { benchmarkBool = n.IsDefined() })
	for b.Loop() {
		benchmarkBool = n.IsDefined()
	}
}

func BenchmarkNotationEqual(b *testing.B) {
	first := notation.MustBuiltIn(notation.UnicodeWyschnegradskyID)
	second := notation.MustBuiltIn(notation.UnicodeWyschnegradskyID)
	customFirst := customNotation(b, first.NameRules(), notation.OctaveStyleNone, "")
	customSecond := customNotation(b, second.NameRules(), notation.OctaveStyleNone, "")
	for _, tt := range []struct {
		name string
		a    notation.Notation
		b    notation.Notation
	}{
		{"same built-in table", first, second},
		{"different tables with equal rules", customFirst, customSecond},
	} {
		b.Run(tt.name, func(b *testing.B) {
			requireZeroAllocs(b, func() { benchmarkBool = tt.a.Equal(tt.b) })
			for b.Loop() {
				benchmarkBool = tt.a.Equal(tt.b)
			}
		})
	}
}

func BenchmarkBuiltIn(b *testing.B) {
	for _, id := range builtInIDOrder {
		if id == notation.ASCIIUpDownID {
			continue
		}
		b.Run(id, func(b *testing.B) {
			requireZeroAllocs(b, func() {
				benchmarkNotation, benchmarkErr = notation.BuiltIn(id)
				if benchmarkErr != nil {
					panic(benchmarkErr)
				}
			})
			for b.Loop() {
				benchmarkNotation, benchmarkErr = notation.BuiltIn(id)
				if benchmarkErr != nil {
					b.Fatal(benchmarkErr)
				}
			}
		})
	}
}

func BenchmarkBuiltInIDs(b *testing.B) {
	for b.Loop() {
		benchmarkStrings = notation.BuiltInIDs()
	}
}

func BenchmarkNameRules_Names(b *testing.B) {
	rules := notation.MustBuiltIn(notation.UnicodeWyschnegradskyID).NameRules()
	for b.Loop() {
		benchmarkNames = rules.Names()
	}
}

func BenchmarkNotation_Format(b *testing.B) {
	for _, tt := range []struct {
		name         string
		id           string
		spellingName spelling.Spelling
	}{
		{"short", notation.EnglishUnicodeID, spelling.CSharp()},
		{"unicode", notation.UnicodeWyschnegradskyID, nameWithAlteration(letter.C, 11, 6)},
	} {
		b.Run(tt.name, func(b *testing.B) {
			n := notation.MustBuiltIn(tt.id)
			for b.Loop() {
				benchmarkString, benchmarkErr = n.Format(tt.spellingName)
				if benchmarkErr != nil {
					b.Fatal(benchmarkErr)
				}
			}
		})
	}
}

func BenchmarkNotation_Parse(b *testing.B) {
	for _, tt := range []struct {
		name string
		id   string
		text string
	}{
		{"short", notation.EnglishUnicodeID, "C♯"},
		{"unicode", notation.UnicodeWyschnegradskyID, "C\ue42a"},
	} {
		b.Run(tt.name, func(b *testing.B) {
			n := notation.MustBuiltIn(tt.id)
			for b.Loop() {
				benchmarkName, benchmarkErr = n.Parse(tt.text)
				if benchmarkErr != nil {
					b.Fatal(benchmarkErr)
				}
			}
		})
	}
}

func BenchmarkPitch_Format(b *testing.B) {
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	lily := notation.MustBuiltIn(notation.LilyPondNederlandsID)
	for _, tt := range []struct {
		name string
		n    notation.Notation
		p    pitch.Pitch
		ctx  notation.OctaveContext
	}{
		{"short number", english, pitch.MustFromASPN(spelling.CSharp(), 4), notation.OctaveContext{Numbering: number.ASPN()}},
		{"long octave marks", lily, pitch.MustFromASPN(spelling.C(), 67), notation.OctaveContext{}},
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				benchmarkString, benchmarkErr = tt.n.FormatPitch(tt.p, system.Traditional(), tt.ctx)
				if benchmarkErr != nil {
					b.Fatal(benchmarkErr)
				}
			}
		})
	}
}

func BenchmarkPitch_Parse(b *testing.B) {
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	lily := notation.MustBuiltIn(notation.LilyPondNederlandsID)
	for _, tt := range []struct {
		name string
		n    notation.Notation
		text string
		ctx  notation.OctaveContext
	}{
		{"short number", english, "C♯4", notation.OctaveContext{Numbering: number.ASPN()}},
		{"long octave marks", lily, "c" + strings.Repeat("'", 64), notation.OctaveContext{}},
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				benchmarkPitch, benchmarkErr = tt.n.ParsePitch(tt.text, system.Traditional(), tt.ctx)
				if benchmarkErr != nil {
					b.Fatal(benchmarkErr)
				}
			}
		})
	}
}

func BenchmarkConvert(b *testing.B) {
	german := notation.MustBuiltIn(notation.GermanAMID)
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	wysch := notation.MustBuiltIn(notation.UnicodeWyschnegradskyID)
	for _, tt := range []struct {
		name string
		from notation.Notation
		to   notation.Notation
		text string
	}{
		{"short", german, english, "cis"},
		{"unicode", wysch, wysch, "C\ue42a"},
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				benchmarkString, benchmarkErr = notation.Convert(tt.text, tt.from, tt.to)
				if benchmarkErr != nil {
					b.Fatal(benchmarkErr)
				}
			}
		})
	}
}

func BenchmarkConvertPitch(b *testing.B) {
	english := notation.MustBuiltIn(notation.EnglishUnicodeID)
	lily := notation.MustBuiltIn(notation.LilyPondNederlandsID)
	for _, tt := range []struct {
		name string
		from notation.Notation
		text string
	}{
		{"short", lily, "cis'"},
		{"long octave marks", lily, "c" + strings.Repeat("'", 64)},
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				benchmarkString, benchmarkErr = notation.ConvertPitch(tt.text, tt.from, english, system.Traditional(), notation.OctaveContext{}, notation.OctaveContext{Numbering: number.ASPN()})
				if benchmarkErr != nil {
					b.Fatal(benchmarkErr)
				}
			}
		})
	}
}

func TestBenchmarkRepresentativeInputs(t *testing.T) {
	require.Equal(t, "C\ue42a", notation.MustBuiltIn(notation.UnicodeWyschnegradskyID).MustFormat(nameWithAlteration(letter.C, 11, 6)))
	require.Equal(t, "C67", notation.MustBuiltIn(notation.EnglishUnicodeID).MustFormatPitch(pitch.MustFromASPN(spelling.C(), 67), system.Traditional(), notation.OctaveContext{Numbering: number.ASPN()}))
}
