package cents

import "sync"

// NamedValue is a value with its English name, as listed by [NamedValues]
// and [NamedUnits].
type NamedValue struct {
	Name  string
	Value Cents
}

// Basic values and units with a rational record have constant encodings.
var (
	zeroValue      = Cents{}
	centValue      = Cents{rec: "1"}
	semitoneValue  = Cents{rec: "100"}
	octaveValue    = Cents{rec: "1200"}
	midiUnitValue  = Cents{rec: "25/4096"}
	jotValue       = Cents{rec: "1200/30103"}
	milliOctValue  = Cents{rec: "6/5"}
	centitoneValue = Cents{rec: "2"}
	savartValue    = Cents{rec: "6/5;5=1/1000"}
	eptamerValue   = Cents{rec: "1200/301"}
	merideValue    = Cents{rec: "1200/43"}
	holdrianValue  = Cents{rec: "1200/53"}
)

// Zero returns 0 cents, the same as the zero value of Cents.
func Zero() Cents { return zeroValue }

// Cent returns 1 cent: 1/1200 of an octave.
func Cent() Cents { return centValue }

// Semitone returns 100 cents: the equal-tempered semitone, 1/12 of an
// octave, also a named unit.
func Semitone() Cents { return semitoneValue }

// Octave returns 1200 cents: the ratio 2:1.
func Octave() Cents { return octaveValue }

// MIDITuningUnit returns the unit of the MIDI Tuning Standard: 1/196608 of
// an octave (1/16384 of a semitone), 25/4096 = 0.006103515625 cents exactly.
func MIDITuningUnit() Cents { return midiUnitValue }

// Jot returns the jot of Augustus De Morgan: 1/30103 of an octave,
// 1200/30103 ≈ 0.039863 cents.
func Jot() Cents { return jotValue }

// Millioctave returns the millioctave: 1/1000 of an octave, 6/5 = 1.2 cents
// exactly.
func Millioctave() Cents { return milliOctValue }

// Centitone returns the centitone (the Iring unit): 1/600 of an octave,
// 2 cents exactly.
func Centitone() Cents { return centitoneValue }

// Savart returns the decimal savart: 1/1000 of a decade, so that 1000
// savarts are the ratio 10:1; 1200·log2(10^(1/1000)) ≈ 3.986314 cents, about
// 301.029996 savarts in an octave. Its exact record is 6/5 cents plus 1/1000
// of the size of the prime 5. Other meanings of the word "savart" are
// described in the package documentation.
func Savart() Cents { return savartValue }

// Eptameride returns the eptaméride (heptaméride) of Joseph Sauveur: 1/7 of
// a méride, 1/301 of an octave, 1200/301 ≈ 3.986711 cents.
func Eptameride() Cents { return eptamerValue }

// Meride returns the méride of Joseph Sauveur: 1/43 of an octave,
// 1200/43 ≈ 27.906977 cents.
func Meride() Cents { return merideValue }

// units lists the named units in increasing order of size.
var units = sync.OnceValue(func() *[8]NamedValue {
	return &[8]NamedValue{
		{"MIDI tuning unit", midiUnitValue},
		{"jot", jotValue},
		{"millioctave", milliOctValue},
		{"centitone", centitoneValue},
		{"savart", savartValue},
		{"eptameride", eptamerValue},
		{"meride", merideValue},
		{"semitone", semitoneValue},
	}
})

// NamedUnits returns the 8 named units in increasing order of size: "MIDI
// tuning unit", "jot", "millioctave", "centitone", "savart", "eptameride",
// "meride", and "semitone". The Value of each is the result of the function
// of the same unit, such as [Jot]. The slice is new on every call and
// belongs to the caller; changing it changes neither later calls nor the
// units.
func NamedUnits() []NamedValue {
	return cloneNamed(units()[:])
}

// named lists the named small intervals in increasing order.
var named = sync.OnceValue(func() *[19]NamedValue {
	ratio := func(name string, n, d int) NamedValue {
		return NamedValue{name, MustFromRatio(n, d)}
	}
	return &[19]NamedValue{
		ratio("schisma", 32805, 32768),
		ratio("septimal kleisma", 225, 224),
		ratio("kleisma", 15625, 15552),
		ratio("Ptolemy's comma", 100, 99),
		ratio("diaschisma", 2048, 2025),
		ratio("syntonic comma", 81, 80),
		{"Holdrian comma", holdrianValue},
		ratio("Pythagorean comma", 531441, 524288),
		ratio("septimal comma", 64, 63),
		ratio("lesser diesis", 128, 125),
		ratio("septimal quarter tone", 36, 35),
		ratio("undecimal comma", 33, 32),
		ratio("greater diesis", 648, 625),
		ratio("tridecimal comma", 27, 26),
		ratio("minor chroma", 25, 24),
		ratio("limma", 256, 243),
		ratio("major chroma", 135, 128),
		ratio("apotome", 2187, 2048),
		ratio("large limma", 27, 25),
	}
})

// NamedValues returns the 19 named small intervals, in increasing order:
// the same values and order as the functions [Schisma] to [LargeLimma]. The
// Name of each is its main English name: "schisma", "septimal kleisma",
// "kleisma", "Ptolemy's comma", "diaschisma", "syntonic comma", "Holdrian
// comma", "Pythagorean comma", "septimal comma", "lesser diesis", "septimal
// quarter tone", "undecimal comma", "greater diesis", "tridecimal comma",
// "minor chroma", "limma", "major chroma", "apotome", and "large limma".
// Synonyms are given in the documentation of the functions. The slice is new
// on every call and belongs to the caller; changing it changes neither later
// calls nor the values.
func NamedValues() []NamedValue {
	return cloneNamed(named()[:])
}

// cloneNamed returns a copy of s with a capacity equal to its length.
func cloneNamed(s []NamedValue) []NamedValue {
	c := make([]NamedValue, len(s))
	copy(c, s)
	return c
}

// Schisma returns the schisma 32805:32768 ≈ 1.953721 cents: the
// Pythagorean comma minus the syntonic comma. It is not the "Grad" or
// "schisma" of Werckmeister (≈ 1.955 cents, a twelfth of the Pythagorean
// comma).
func Schisma() Cents { return named()[0].Value }

// SeptimalKleisma returns the septimal kleisma 225:224 ≈ 7.711523 cents.
func SeptimalKleisma() Cents { return named()[1].Value }

// Kleisma returns the kleisma 15625:15552 ≈ 8.107279 cents.
func Kleisma() Cents { return named()[2].Value }

// PtolemysComma returns Ptolemy's comma 100:99 ≈ 17.399484 cents: the
// difference of the ratios 10:9 and 11:10. Wikipedia also calls it the
// "small undecimal comma", a name that the Huygens-Fokker list of intervals
// gives to 99:98.
func PtolemysComma() Cents { return named()[3].Value }

// Diaschisma returns the diaschisma 2048:2025 ≈ 19.552569 cents: two
// syntonic commas minus the Pythagorean comma.
func Diaschisma() Cents { return named()[4].Value }

// SyntonicComma returns the syntonic comma 81:80 ≈ 21.506290 cents, also
// called Didymus' comma; "comma" without a qualifier usually means it.
func SyntonicComma() Cents { return named()[5].Value }

// HoldrianComma returns the Holdrian comma (53-TET comma, Arabic comma):
// 1/53 of an octave, 1200/53 ≈ 22.641509 cents, the same as Step(1, 53). It
// is not the cents of a ratio of integers.
func HoldrianComma() Cents { return named()[6].Value }

// PythagoreanComma returns the Pythagorean (ditonic) comma
// 531441:524288 = 3^12:2^19 ≈ 23.460010 cents: twelve fifths 3:2 minus seven
// octaves.
func PythagoreanComma() Cents { return named()[7].Value }

// SeptimalComma returns the septimal comma (Archytas' comma) 64:63 ≈
// 27.264092 cents.
func SeptimalComma() Cents { return named()[8].Value }

// LesserDiesis returns the lesser (minor, enharmonic) diesis 128:125 ≈
// 41.058858 cents: three syntonic commas minus the Pythagorean comma;
// "diesis" without a qualifier usually means it.
func LesserDiesis() Cents { return named()[9].Value }

// SeptimalQuarterTone returns the septimal quarter tone 36:35 ≈ 48.770381
// cents. It is also called the "septimal diesis", but that name is used for
// 49:48 (≈ 35.696812 cents, not a named value of the package) as well.
func SeptimalQuarterTone() Cents { return named()[10].Value }

// UndecimalComma returns the undecimal comma 33:32 ≈ 53.272943 cents.
func UndecimalComma() Cents { return named()[11].Value }

// GreaterDiesis returns the greater (major) diesis 648:625 ≈ 62.565148
// cents: four syntonic commas minus the Pythagorean comma.
func GreaterDiesis() Cents { return named()[12].Value }

// TridecimalComma returns the tridecimal comma 27:26 ≈ 65.337341 cents.
func TridecimalComma() Cents { return named()[13].Value }

// MinorChroma returns the minor chroma (classic chromatic semitone) 25:24 ≈
// 70.672427 cents.
func MinorChroma() Cents { return named()[14].Value }

// Limma returns the limma (Pythagorean minor second) 256:243 ≈ 90.224996
// cents.
func Limma() Cents { return named()[15].Value }

// MajorChroma returns the major chroma (major limma) 135:128 ≈ 92.178716
// cents.
func MajorChroma() Cents { return named()[16].Value }

// Apotome returns the apotome 2187:2048 ≈ 113.685006 cents: the Pythagorean
// chromatic semitone; the apotome minus the limma is the Pythagorean comma.
func Apotome() Cents { return named()[17].Value }

// LargeLimma returns the large limma (Zarlino semitone) 27:25 ≈ 133.237575
// cents.
func LargeLimma() Cents { return named()[18].Value }
