package instrument

import (
	"fmt"

	"github.com/go-muse/musex/settings"
)

// builtinEntry is a row of the table of built-in presets: the name and the
// bounds of the range, written by the standard settings (settings.Standard).
type builtinEntry struct {
	name  string
	lower string
	upper string
}

// builtinTable lists the built-in presets in the order of the built-in
// catalog. The ranges are sounding ranges; every bound is written by the
// standard settings, and for every bound the written octave equals the
// Octave. The functions below refer to the rows by index.
var builtinTable = [...]builtinEntry{
	// Bowed strings.
	{"Violin", "G3", "E7"},      // 0
	{"Cello", "C2", "G6"},       // 1
	{"Double Bass", "E1", "D5"}, // 2
	// Plucked strings.
	{"Classical Guitar", "E2", "B5"}, // 3
	{"Bass Guitar", "E1", "G4"},      // 4
	{"Mandolin", "G3", "F♯6"},        // 5
	// Woodwinds.
	{"Flute", "B3", "C7"},               // 6
	{"Oboe", "B♭3", "G6"},               // 7
	{"English Horn", "E3", "C6"},        // 8
	{"Clarinet in B-flat", "D3", "B♭6"}, // 9
	{"Clarinet in A", "C♯3", "A6"},      // 10
	{"Bass Clarinet", "B♭1", "D6"},      // 11
	{"Bassoon", "B♭1", "E5"},            // 12
	{"Contrabassoon", "B♭0", "D4"},      // 13
	{"Soprano Saxophone", "A♭3", "E6"},  // 14
	{"Alto Saxophone", "D♭3", "A♭5"},    // 15
	{"Tenor Saxophone", "A♭2", "E5"},    // 16
	// Brass.
	{"Horn", "B♭1", "F5"},              // 17
	{"Trumpet in B-flat", "E3", "B♭5"}, // 18
	{"Trombone", "E2", "F5"},           // 19
	{"Bass Trombone", "B♭0", "B♭4"},    // 20
	{"Bass Tuba", "A1", "A♭4"},         // 21
	{"Contrabass Tuba", "E1", "F4"},    // 22
	// Keyboards.
	{"Piano", "A0", "C8"},           // 23
	{"Piano (85 keys)", "A0", "A7"}, // 24
	{"Piano (92 keys)", "F0", "C8"}, // 25
	{"Organ", "C0", "C9"},           // 26
	{"Harpsichord", "F1", "F6"},     // 27
	{"Celesta", "C4", "C8"},         // 28
	// Pitched percussion.
	{"Timpani", "D2", "A3"},    // 29
	{"Xylophone", "F4", "C8"},  // 30
	{"Marimba", "C2", "C7"},    // 31
	{"Vibraphone", "F3", "F6"}, // 32
	// Voices.
	{"Soprano Voice", "C4", "G5"},       // 33
	{"Mezzo-Soprano Voice", "A3", "A5"}, // 34
	{"Alto Voice", "G3", "D5"},          // 35
	{"Tenor Voice", "C3", "G4"},         // 36
	{"Baritone Voice", "G2", "A♭4"},     // 37
	{"Bass Voice", "F2", "D4"},          // 38
}

// builtinRanges holds the parsed ranges of builtinTable, in the same order.
// It is built when the package is initialized and only read afterward.
var builtinRanges = parseBuiltins()

// parseBuiltins parses the bounds of builtinTable by the standard settings
// and checks every range in the traditional system. An error is an error of
// the data of the package: it panics while the package is initialized, as
// regexp.MustCompile does for a constant pattern.
func parseBuiltins() [len(builtinTable)]Range {
	std := settings.Standard()
	sys := std.System()

	var ranges [len(builtinTable)]Range
	for i, e := range builtinTable {
		lower, err := std.ParsePitch(e.lower)
		if err != nil {
			panic(fmt.Sprintf("instrument: built-in preset %q: lower bound %q: %v", e.name, e.lower, err))
		}

		upper, err := std.ParsePitch(e.upper)
		if err != nil {
			panic(fmt.Sprintf("instrument: built-in preset %q: upper bound %q: %v", e.name, e.upper, err))
		}

		r := Range{lower: lower, upper: upper}
		if err := r.Check(sys); err != nil {
			panic(fmt.Sprintf("instrument: built-in preset %q: %v", e.name, err))
		}

		ranges[i] = r
	}

	return ranges
}

// builtin returns a new built-in preset of row i: the name, new standard
// settings and the range.
func builtin(i int) *Preset {
	return &Preset{name: builtinTable[i].name, set: settings.Standard(), rng: builtinRanges[i]}
}

// Builtins returns a new catalog of all built-in presets, in the order of the
// functions below. Every call builds new presets and a new catalog, so the
// built-in catalog cannot be changed. Find a built-in preset by its exact name
// with Builtins().Lookup(name).
func Builtins() *Catalog {
	presets := make([]Preset, len(builtinTable))
	for i, e := range builtinTable {
		presets[i] = Preset{name: e.name, set: settings.Standard(), rng: builtinRanges[i]}
	}

	return &Catalog{presets: presets}
}

// Violin returns a new built-in preset "Violin": G3–E7.
func Violin() *Preset { return builtin(0) }

// Cello returns a new built-in preset "Cello": C2–G6.
func Cello() *Preset { return builtin(1) }

// DoubleBass returns a new built-in preset "Double Bass": E1–D5, sounding.
func DoubleBass() *Preset { return builtin(2) }

// ClassicalGuitar returns a new built-in preset "Classical Guitar": E2–B5,
// sounding.
func ClassicalGuitar() *Preset { return builtin(3) }

// BassGuitar returns a new built-in preset "Bass Guitar": E1–G4, sounding.
func BassGuitar() *Preset { return builtin(4) }

// Mandolin returns a new built-in preset "Mandolin": G3–F♯6.
func Mandolin() *Preset { return builtin(5) }

// Flute returns a new built-in preset "Flute": B3–C7.
func Flute() *Preset { return builtin(6) }

// Oboe returns a new built-in preset "Oboe": B♭3–G6.
func Oboe() *Preset { return builtin(7) }

// EnglishHorn returns a new built-in preset "English Horn": E3–C6, sounding.
func EnglishHorn() *Preset { return builtin(8) }

// ClarinetInBFlat returns a new built-in preset "Clarinet in B-flat":
// D3–B♭6, sounding.
func ClarinetInBFlat() *Preset { return builtin(9) }

// ClarinetInA returns a new built-in preset "Clarinet in A": C♯3–A6,
// sounding.
func ClarinetInA() *Preset { return builtin(10) }

// BassClarinet returns a new built-in preset "Bass Clarinet": B♭1–D6,
// sounding.
func BassClarinet() *Preset { return builtin(11) }

// Bassoon returns a new built-in preset "Bassoon": B♭1–E5.
func Bassoon() *Preset { return builtin(12) }

// Contrabassoon returns a new built-in preset "Contrabassoon": B♭0–D4,
// sounding.
func Contrabassoon() *Preset { return builtin(13) }

// SopranoSaxophone returns a new built-in preset "Soprano Saxophone":
// A♭3–E6, sounding.
func SopranoSaxophone() *Preset { return builtin(14) }

// AltoSaxophone returns a new built-in preset "Alto Saxophone": D♭3–A♭5,
// sounding.
func AltoSaxophone() *Preset { return builtin(15) }

// TenorSaxophone returns a new built-in preset "Tenor Saxophone": A♭2–E5,
// sounding.
func TenorSaxophone() *Preset { return builtin(16) }

// Horn returns a new built-in preset "Horn": B♭1–F5, sounding.
func Horn() *Preset { return builtin(17) }

// TrumpetInBFlat returns a new built-in preset "Trumpet in B-flat": E3–B♭5,
// sounding.
func TrumpetInBFlat() *Preset { return builtin(18) }

// Trombone returns a new built-in preset "Trombone": E2–F5.
func Trombone() *Preset { return builtin(19) }

// BassTrombone returns a new built-in preset "Bass Trombone": B♭0–B♭4.
func BassTrombone() *Preset { return builtin(20) }

// BassTuba returns a new built-in preset "Bass Tuba": A1–A♭4.
func BassTuba() *Preset { return builtin(21) }

// ContrabassTuba returns a new built-in preset "Contrabass Tuba": E1–F4.
func ContrabassTuba() *Preset { return builtin(22) }

// Piano returns a new built-in preset "Piano", 88 keys: A0–C8.
func Piano() *Preset { return builtin(23) }

// Piano85Keys returns a new built-in preset "Piano (85 keys)": A0–A7.
func Piano85Keys() *Preset { return builtin(24) }

// Piano92Keys returns a new built-in preset "Piano (92 keys)": F0–C8.
func Piano92Keys() *Preset { return builtin(25) }

// Organ returns a new built-in preset "Organ": C0–C9, the sounding range with
// stops, not the keyboard.
func Organ() *Preset { return builtin(26) }

// Harpsichord returns a new built-in preset "Harpsichord": F1–F6.
func Harpsichord() *Preset { return builtin(27) }

// Celesta returns a new built-in preset "Celesta": C4–C8, sounding.
func Celesta() *Preset { return builtin(28) }

// Timpani returns a new built-in preset "Timpani", a set of four: D2–A3.
func Timpani() *Preset { return builtin(29) }

// Xylophone returns a new built-in preset "Xylophone": F4–C8, sounding.
func Xylophone() *Preset { return builtin(30) }

// Marimba returns a new built-in preset "Marimba", five octaves: C2–C7.
func Marimba() *Preset { return builtin(31) }

// Vibraphone returns a new built-in preset "Vibraphone": F3–F6.
func Vibraphone() *Preset { return builtin(32) }

// SopranoVoice returns a new built-in preset "Soprano Voice": C4–G5.
func SopranoVoice() *Preset { return builtin(33) }

// MezzoSopranoVoice returns a new built-in preset "Mezzo-Soprano Voice":
// A3–A5.
func MezzoSopranoVoice() *Preset { return builtin(34) }

// AltoVoice returns a new built-in preset "Alto Voice": G3–D5.
func AltoVoice() *Preset { return builtin(35) }

// TenorVoice returns a new built-in preset "Tenor Voice": C3–G4.
func TenorVoice() *Preset { return builtin(36) }

// BaritoneVoice returns a new built-in preset "Baritone Voice": G2–A♭4.
func BaritoneVoice() *Preset { return builtin(37) }

// BassVoice returns a new built-in preset "Bass Voice": F2–D4.
func BassVoice() *Preset { return builtin(38) }
