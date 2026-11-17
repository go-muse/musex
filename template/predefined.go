package template

// builtInCatalog is initialized once from immutable compiled-in data. Its
// entries and templates never expose mutable internal slices.
var builtInCatalog = makeBuiltInCatalog()

func makeBuiltInCatalog() Catalog {
	entries := make([]Entry, len(catalogRows))
	for i, row := range catalogRows {
		tmpl := MustParse(row.notation)
		entries[i] = MustNewEntry(row.code, row.names, tmpl)
	}
	return MustNewCatalog(entries...)
}

// BuiltIn returns the immutable built-in catalog of 81 entries in its accepted
// catalog order. Operations on the returned value never change this catalog.
func BuiltIn() Catalog {
	return builtInCatalog
}

func builtInTemplate(index int) Template {
	return builtInCatalog.entries[index].tmpl
}

func builtInEntry(index int) Entry {
	return builtInCatalog.entries[index]
}

// Ionian returns the built-in template with code "ionian".
func Ionian() Template { return builtInTemplate(0) }

// IonianEntry returns the built-in entry with code "ionian".
func IonianEntry() Entry { return builtInEntry(0) }

// Dorian returns the built-in template with code "dorian".
func Dorian() Template { return builtInTemplate(1) }

// DorianEntry returns the built-in entry with code "dorian".
func DorianEntry() Entry { return builtInEntry(1) }

// Phrygian returns the built-in template with code "phrygian".
func Phrygian() Template { return builtInTemplate(2) }

// PhrygianEntry returns the built-in entry with code "phrygian".
func PhrygianEntry() Entry { return builtInEntry(2) }

// Lydian returns the built-in template with code "lydian".
func Lydian() Template { return builtInTemplate(3) }

// LydianEntry returns the built-in entry with code "lydian".
func LydianEntry() Entry { return builtInEntry(3) }

// Mixolydian returns the built-in template with code "mixolydian".
func Mixolydian() Template { return builtInTemplate(4) }

// MixolydianEntry returns the built-in entry with code "mixolydian".
func MixolydianEntry() Entry { return builtInEntry(4) }

// Aeolian returns the built-in template with code "aeolian".
func Aeolian() Template { return builtInTemplate(5) }

// AeolianEntry returns the built-in entry with code "aeolian".
func AeolianEntry() Entry { return builtInEntry(5) }

// Locrian returns the built-in template with code "locrian".
func Locrian() Template { return builtInTemplate(6) }

// LocrianEntry returns the built-in entry with code "locrian".
func LocrianEntry() Entry { return builtInEntry(6) }

// HarmonicMajor returns the built-in template with code "harmonic-major".
func HarmonicMajor() Template { return builtInTemplate(7) }

// HarmonicMajorEntry returns the built-in entry with code "harmonic-major".
func HarmonicMajorEntry() Entry { return builtInEntry(7) }

// MelodicMajor returns the built-in template with code "melodic-major".
func MelodicMajor() Template { return builtInTemplate(8) }

// MelodicMajorEntry returns the built-in entry with code "melodic-major".
func MelodicMajorEntry() Entry { return builtInEntry(8) }

// HarmonicMinor returns the built-in template with code "harmonic-minor".
func HarmonicMinor() Template { return builtInTemplate(9) }

// HarmonicMinorEntry returns the built-in entry with code "harmonic-minor".
func HarmonicMinorEntry() Entry { return builtInEntry(9) }

// MelodicMinor returns the built-in template with code "melodic-minor".
func MelodicMinor() Template { return builtInTemplate(10) }

// MelodicMinorEntry returns the built-in entry with code "melodic-minor".
func MelodicMinorEntry() Entry { return builtInEntry(10) }

// DoubleHarmonicMajor returns the template with code "double-harmonic-major".
func DoubleHarmonicMajor() Template { return builtInTemplate(11) }

// DoubleHarmonicMajorEntry returns the entry with code "double-harmonic-major".
func DoubleHarmonicMajorEntry() Entry { return builtInEntry(11) }

// DoubleHarmonicMinor returns the template with code "double-harmonic-minor".
func DoubleHarmonicMinor() Template { return builtInTemplate(12) }

// DoubleHarmonicMinorEntry returns the entry with code "double-harmonic-minor".
func DoubleHarmonicMinorEntry() Entry { return builtInEntry(12) }

// Acoustic returns the built-in template with code "acoustic".
func Acoustic() Template { return builtInTemplate(13) }

// AcousticEntry returns the built-in entry with code "acoustic".
func AcousticEntry() Entry { return builtInEntry(13) }

// DorianB2 returns the built-in template with code "dorian-b2".
func DorianB2() Template { return builtInTemplate(14) }

// DorianB2Entry returns the built-in entry with code "dorian-b2".
func DorianB2Entry() Entry { return builtInEntry(14) }

// LydianAugmented returns the built-in template with code "lydian-augmented".
func LydianAugmented() Template { return builtInTemplate(15) }

// LydianAugmentedEntry returns the built-in entry with code "lydian-augmented".
func LydianAugmentedEntry() Entry { return builtInEntry(15) }

// LocrianSharp2 returns the built-in template with code "locrian-sharp2".
func LocrianSharp2() Template { return builtInTemplate(16) }

// LocrianSharp2Entry returns the built-in entry with code "locrian-sharp2".
func LocrianSharp2Entry() Entry { return builtInEntry(16) }

// Altered returns the built-in template with code "altered".
func Altered() Template { return builtInTemplate(17) }

// AlteredEntry returns the built-in entry with code "altered".
func AlteredEntry() Entry { return builtInEntry(17) }

// PentatonicMajor returns the built-in template with code "pentatonic-major".
func PentatonicMajor() Template { return builtInTemplate(18) }

// PentatonicMajorEntry returns the built-in entry with code "pentatonic-major".
func PentatonicMajorEntry() Entry { return builtInEntry(18) }

// PentatonicMinor returns the built-in template with code "pentatonic-minor".
func PentatonicMinor() Template { return builtInTemplate(19) }

// PentatonicMinorEntry returns the built-in entry with code "pentatonic-minor".
func PentatonicMinorEntry() Entry { return builtInEntry(19) }

// Blues returns the built-in template with code "blues".
func Blues() Template { return builtInTemplate(20) }

// BluesEntry returns the built-in entry with code "blues".
func BluesEntry() Entry { return builtInEntry(20) }

// BluesB5 returns the built-in template with code "blues-b5".
func BluesB5() Template { return builtInTemplate(21) }

// BluesB5Entry returns the built-in entry with code "blues-b5".
func BluesB5Entry() Entry { return builtInEntry(21) }

// BluesMajor returns the built-in template with code "blues-major".
func BluesMajor() Template { return builtInTemplate(22) }

// BluesMajorEntry returns the built-in entry with code "blues-major".
func BluesMajorEntry() Entry { return builtInEntry(22) }

// WholeTone returns the built-in template with code "whole-tone".
func WholeTone() Template { return builtInTemplate(23) }

// WholeToneEntry returns the built-in entry with code "whole-tone".
func WholeToneEntry() Entry { return builtInEntry(23) }

// OctatonicWH returns the built-in template with code "octatonic-wh".
func OctatonicWH() Template { return builtInTemplate(24) }

// OctatonicWHEntry returns the built-in entry with code "octatonic-wh".
func OctatonicWHEntry() Entry { return builtInEntry(24) }

// OctatonicHW returns the built-in template with code "octatonic-hw".
func OctatonicHW() Template { return builtInTemplate(25) }

// OctatonicHWEntry returns the built-in entry with code "octatonic-hw".
func OctatonicHWEntry() Entry { return builtInEntry(25) }

// Hexatonic returns the built-in template with code "hexatonic".
func Hexatonic() Template { return builtInTemplate(26) }

// HexatonicEntry returns the built-in entry with code "hexatonic".
func HexatonicEntry() Entry { return builtInEntry(26) }

// Messiaen3 returns the built-in template with code "messiaen-3".
func Messiaen3() Template { return builtInTemplate(27) }

// Messiaen3Entry returns the built-in entry with code "messiaen-3".
func Messiaen3Entry() Entry { return builtInEntry(27) }

// Messiaen3Commons returns the template with code "messiaen-3-commons".
func Messiaen3Commons() Template { return builtInTemplate(28) }

// Messiaen3CommonsEntry returns the entry with code "messiaen-3-commons".
func Messiaen3CommonsEntry() Entry { return builtInEntry(28) }

// Messiaen4Papadopoulos returns the template with code "messiaen-4-papadopoulos".
func Messiaen4Papadopoulos() Template { return builtInTemplate(29) }

// Messiaen4PapadopoulosEntry returns the entry with code "messiaen-4-papadopoulos".
func Messiaen4PapadopoulosEntry() Entry { return builtInEntry(29) }

// Messiaen4Degraaf returns the template with code "messiaen-4-degraaf".
func Messiaen4Degraaf() Template { return builtInTemplate(30) }

// Messiaen4DegraafEntry returns the entry with code "messiaen-4-degraaf".
func Messiaen4DegraafEntry() Entry { return builtInEntry(30) }

// Messiaen4Commons returns the template with code "messiaen-4-commons".
func Messiaen4Commons() Template { return builtInTemplate(31) }

// Messiaen4CommonsEntry returns the entry with code "messiaen-4-commons".
func Messiaen4CommonsEntry() Entry { return builtInEntry(31) }

// Messiaen5Papadopoulos returns the template with code "messiaen-5-papadopoulos".
func Messiaen5Papadopoulos() Template { return builtInTemplate(32) }

// Messiaen5PapadopoulosEntry returns the entry with code "messiaen-5-papadopoulos".
func Messiaen5PapadopoulosEntry() Entry { return builtInEntry(32) }

// Messiaen5Degraaf returns the template with code "messiaen-5-degraaf".
func Messiaen5Degraaf() Template { return builtInTemplate(33) }

// Messiaen5DegraafEntry returns the entry with code "messiaen-5-degraaf".
func Messiaen5DegraafEntry() Entry { return builtInEntry(33) }

// Messiaen5Commons returns the template with code "messiaen-5-commons".
func Messiaen5Commons() Template { return builtInTemplate(34) }

// Messiaen5CommonsEntry returns the entry with code "messiaen-5-commons".
func Messiaen5CommonsEntry() Entry { return builtInEntry(34) }

// Messiaen6 returns the built-in template with code "messiaen-6".
func Messiaen6() Template { return builtInTemplate(35) }

// Messiaen6Entry returns the built-in entry with code "messiaen-6".
func Messiaen6Entry() Entry { return builtInEntry(35) }

// Messiaen7Degraaf returns the template with code "messiaen-7-degraaf".
func Messiaen7Degraaf() Template { return builtInTemplate(36) }

// Messiaen7DegraafEntry returns the entry with code "messiaen-7-degraaf".
func Messiaen7DegraafEntry() Entry { return builtInEntry(36) }

// Messiaen7Commons returns the template with code "messiaen-7-commons".
func Messiaen7Commons() Template { return builtInTemplate(37) }

// Messiaen7CommonsEntry returns the entry with code "messiaen-7-commons".
func Messiaen7CommonsEntry() Entry { return builtInEntry(37) }

// BebopDominant returns the built-in template with code "bebop-dominant".
func BebopDominant() Template { return builtInTemplate(38) }

// BebopDominantEntry returns the built-in entry with code "bebop-dominant".
func BebopDominantEntry() Entry { return builtInEntry(38) }

// BebopMajor returns the built-in template with code "bebop-major".
func BebopMajor() Template { return builtInTemplate(39) }

// BebopMajorEntry returns the built-in entry with code "bebop-major".
func BebopMajorEntry() Entry { return builtInEntry(39) }

// LocrianNatural6 returns the built-in template with code "locrian-natural6".
func LocrianNatural6() Template { return builtInTemplate(40) }

// LocrianNatural6Entry returns the built-in entry with code "locrian-natural6".
func LocrianNatural6Entry() Entry { return builtInEntry(40) }

// IonianSharp5 returns the built-in template with code "ionian-sharp5".
func IonianSharp5() Template { return builtInTemplate(41) }

// IonianSharp5Entry returns the built-in entry with code "ionian-sharp5".
func IonianSharp5Entry() Entry { return builtInEntry(41) }

// UkrainianDorian returns the built-in template with code "ukrainian-dorian".
func UkrainianDorian() Template { return builtInTemplate(42) }

// UkrainianDorianEntry returns the built-in entry with code "ukrainian-dorian".
func UkrainianDorianEntry() Entry { return builtInEntry(42) }

// PhrygianDominant returns the built-in template with code "phrygian-dominant".
func PhrygianDominant() Template { return builtInTemplate(43) }

// PhrygianDominantEntry returns the built-in entry with code "phrygian-dominant".
func PhrygianDominantEntry() Entry { return builtInEntry(43) }

// LydianSharp2 returns the built-in template with code "lydian-sharp2".
func LydianSharp2() Template { return builtInTemplate(44) }

// LydianSharp2Entry returns the built-in entry with code "lydian-sharp2".
func LydianSharp2Entry() Entry { return builtInEntry(44) }

// Ultralocrian returns the built-in template with code "ultralocrian".
func Ultralocrian() Template { return builtInTemplate(45) }

// UltralocrianEntry returns the built-in entry with code "ultralocrian".
func UltralocrianEntry() Entry { return builtInEntry(45) }

// DorianB5 returns the built-in template with code "dorian-b5".
func DorianB5() Template { return builtInTemplate(46) }

// DorianB5Entry returns the built-in entry with code "dorian-b5".
func DorianB5Entry() Entry { return builtInEntry(46) }

// PhrygianB4 returns the built-in template with code "phrygian-b4".
func PhrygianB4() Template { return builtInTemplate(47) }

// PhrygianB4Entry returns the built-in entry with code "phrygian-b4".
func PhrygianB4Entry() Entry { return builtInEntry(47) }

// LydianB3 returns the built-in template with code "lydian-b3".
func LydianB3() Template { return builtInTemplate(48) }

// LydianB3Entry returns the built-in entry with code "lydian-b3".
func LydianB3Entry() Entry { return builtInEntry(48) }

// MixolydianB2 returns the built-in template with code "mixolydian-b2".
func MixolydianB2() Template { return builtInTemplate(49) }

// MixolydianB2Entry returns the built-in entry with code "mixolydian-b2".
func MixolydianB2Entry() Entry { return builtInEntry(49) }

// LydianSharp2Sharp5 returns the template with code "lydian-sharp2-sharp5".
func LydianSharp2Sharp5() Template { return builtInTemplate(50) }

// LydianSharp2Sharp5Entry returns the entry with code "lydian-sharp2-sharp5".
func LydianSharp2Sharp5Entry() Entry { return builtInEntry(50) }

// LocrianBB7 returns the built-in template with code "locrian-bb7".
func LocrianBB7() Template { return builtInTemplate(51) }

// LocrianBB7Entry returns the built-in entry with code "locrian-bb7".
func LocrianBB7Entry() Entry { return builtInEntry(51) }

// LydianSharp2Sharp6 returns the template with code "lydian-sharp2-sharp6".
func LydianSharp2Sharp6() Template { return builtInTemplate(52) }

// LydianSharp2Sharp6Entry returns the entry with code "lydian-sharp2-sharp6".
func LydianSharp2Sharp6Entry() Entry { return builtInEntry(52) }

// Ultraphrygian returns the built-in template with code "ultraphrygian".
func Ultraphrygian() Template { return builtInTemplate(53) }

// UltraphrygianEntry returns the built-in entry with code "ultraphrygian".
func UltraphrygianEntry() Entry { return builtInEntry(53) }

// Oriental returns the built-in template with code "oriental".
func Oriental() Template { return builtInTemplate(54) }

// OrientalEntry returns the built-in entry with code "oriental".
func OrientalEntry() Entry { return builtInEntry(54) }

// IonianAugmentedSharp2 returns the template with code "ionian-augmented-sharp2".
func IonianAugmentedSharp2() Template { return builtInTemplate(55) }

// IonianAugmentedSharp2Entry returns the entry with code "ionian-augmented-sharp2".
func IonianAugmentedSharp2Entry() Entry { return builtInEntry(55) }

// LocrianBB3BB7 returns the built-in template with code "locrian-bb3-bb7".
func LocrianBB3BB7() Template { return builtInTemplate(56) }

// LocrianBB3BB7Entry returns the built-in entry with code "locrian-bb3-bb7".
func LocrianBB3BB7Entry() Entry { return builtInEntry(56) }

// HungarianMajor returns the built-in template with code "hungarian-major".
func HungarianMajor() Template { return builtInTemplate(57) }

// HungarianMajorEntry returns the built-in entry with code "hungarian-major".
func HungarianMajorEntry() Entry { return builtInEntry(57) }

// AeolianSharp4 returns the built-in template with code "aeolian-sharp4".
func AeolianSharp4() Template { return builtInTemplate(58) }

// AeolianSharp4Entry returns the built-in entry with code "aeolian-sharp4".
func AeolianSharp4Entry() Entry { return builtInEntry(58) }

// NeapolitanMinor returns the built-in template with code "neapolitan-minor".
func NeapolitanMinor() Template { return builtInTemplate(59) }

// NeapolitanMinorEntry returns the built-in entry with code "neapolitan-minor".
func NeapolitanMinorEntry() Entry { return builtInEntry(59) }

// NeapolitanMajor returns the built-in template with code "neapolitan-major".
func NeapolitanMajor() Template { return builtInTemplate(60) }

// NeapolitanMajorEntry returns the built-in entry with code "neapolitan-major".
func NeapolitanMajorEntry() Entry { return builtInEntry(60) }

// Persian returns the built-in template with code "persian".
func Persian() Template { return builtInTemplate(61) }

// PersianEntry returns the built-in entry with code "persian".
func PersianEntry() Entry { return builtInEntry(61) }

// Enigmatic returns the built-in template with code "enigmatic".
func Enigmatic() Template { return builtInTemplate(62) }

// EnigmaticEntry returns the built-in entry with code "enigmatic".
func EnigmaticEntry() Entry { return builtInEntry(62) }

// Prometheus returns the built-in template with code "prometheus".
func Prometheus() Template { return builtInTemplate(63) }

// PrometheusEntry returns the built-in entry with code "prometheus".
func PrometheusEntry() Entry { return builtInEntry(63) }

// Tritone returns the built-in template with code "tritone".
func Tritone() Template { return builtInTemplate(64) }

// TritoneEntry returns the built-in entry with code "tritone".
func TritoneEntry() Entry { return builtInEntry(64) }

// TwoSemitoneTritone returns the template with code "two-semitone-tritone".
func TwoSemitoneTritone() Template { return builtInTemplate(65) }

// TwoSemitoneTritoneEntry returns the entry with code "two-semitone-tritone".
func TwoSemitoneTritoneEntry() Entry { return builtInEntry(65) }

// Augmented returns the built-in template with code "augmented".
func Augmented() Template { return builtInTemplate(66) }

// AugmentedEntry returns the built-in entry with code "augmented".
func AugmentedEntry() Entry { return builtInEntry(66) }

// PentatonicSuspended returns the template with code "pentatonic-suspended".
func PentatonicSuspended() Template { return builtInTemplate(67) }

// PentatonicSuspendedEntry returns the entry with code "pentatonic-suspended".
func PentatonicSuspendedEntry() Entry { return builtInEntry(67) }

// PentatonicBluesMinor returns the template with code "pentatonic-blues-minor".
func PentatonicBluesMinor() Template { return builtInTemplate(68) }

// PentatonicBluesMinorEntry returns the entry with code "pentatonic-blues-minor".
func PentatonicBluesMinorEntry() Entry { return builtInEntry(68) }

// Yo returns the built-in template with code "yo".
func Yo() Template { return builtInTemplate(69) }

// YoEntry returns the built-in entry with code "yo".
func YoEntry() Entry { return builtInEntry(69) }

// HirajoshiBurrows returns the built-in template with code "hirajoshi-burrows".
func HirajoshiBurrows() Template { return builtInTemplate(70) }

// HirajoshiBurrowsEntry returns the built-in entry with code "hirajoshi-burrows".
func HirajoshiBurrowsEntry() Entry { return builtInEntry(70) }

// HirajoshiSachs returns the built-in template with code "hirajoshi-sachs".
func HirajoshiSachs() Template { return builtInTemplate(71) }

// HirajoshiSachsEntry returns the built-in entry with code "hirajoshi-sachs".
func HirajoshiSachsEntry() Entry { return builtInEntry(71) }

// HirajoshiKostka returns the built-in template with code "hirajoshi-kostka".
func HirajoshiKostka() Template { return builtInTemplate(72) }

// HirajoshiKostkaEntry returns the built-in entry with code "hirajoshi-kostka".
func HirajoshiKostkaEntry() Entry { return builtInEntry(72) }

// Iwato returns the built-in template with code "iwato".
func Iwato() Template { return builtInTemplate(73) }

// IwatoEntry returns the built-in entry with code "iwato".
func IwatoEntry() Entry { return builtInEntry(73) }

// In returns the built-in template with code "in".
func In() Template { return builtInTemplate(74) }

// InEntry returns the built-in entry with code "in".
func InEntry() Entry { return builtInEntry(74) }

// Insen returns the built-in template with code "insen".
func Insen() Template { return builtInTemplate(75) }

// InsenEntry returns the built-in entry with code "insen".
func InsenEntry() Entry { return builtInEntry(75) }

// BebopMelodicMinor returns the template with code "bebop-melodic-minor".
func BebopMelodicMinor() Template { return builtInTemplate(76) }

// BebopMelodicMinorEntry returns the entry with code "bebop-melodic-minor".
func BebopMelodicMinorEntry() Entry { return builtInEntry(76) }

// BebopHarmonicMinor returns the template with code "bebop-harmonic-minor".
func BebopHarmonicMinor() Template { return builtInTemplate(77) }

// BebopHarmonicMinorEntry returns the entry with code "bebop-harmonic-minor".
func BebopHarmonicMinorEntry() Entry { return builtInEntry(77) }

// SeventhFlat5Diminished returns the template with code "seventh-flat5-diminished".
func SeventhFlat5Diminished() Template { return builtInTemplate(78) }

// SeventhFlat5DiminishedEntry returns the entry with code "seventh-flat5-diminished".
func SeventhFlat5DiminishedEntry() Entry { return builtInEntry(78) }

// BebopDorian returns the built-in template with code "bebop-dorian".
func BebopDorian() Template { return builtInTemplate(79) }

// BebopDorianEntry returns the built-in entry with code "bebop-dorian".
func BebopDorianEntry() Entry { return builtInEntry(79) }

// Kumoi returns the built-in template with code "kumoi".
func Kumoi() Template { return builtInTemplate(80) }

// KumoiEntry returns the built-in entry with code "kumoi".
func KumoiEntry() Entry { return builtInEntry(80) }
