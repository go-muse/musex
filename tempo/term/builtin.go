package term

import (
	"sync"

	"github.com/go-muse/musex/common/fraction"
)

// The eight built-in lists are immutable tables of Go values. Each is packed
// once, on first use, and every call returns an equal value that shares the
// packed data. The texts of names and meanings are copied verbatim from the
// sources of the lists; see the package documentation for the sources and the
// conventions of the product.

// builtinRow is a row of a built-in list: a group, a guide whose bound 0 means
// an open bound, and the terms of the place.
type builtinRow struct {
	group   Group
	lower   int
	upper   int
	entries []Entry
}

// builtinList packs rows with the given synonyms of parsing.
func builtinList(rows []builtinRow, synonyms ...synonym) List {
	out := make([]Row, len(rows))
	for i, r := range rows {
		var g Guide
		if r.lower > 0 {
			g.lower = fraction.FromInt(r.lower)
		}
		if r.upper > 0 {
			g.upper = fraction.FromInt(r.upper)
		}
		out[i] = Row{Entries: r.entries, Group: r.group, Guide: &g}
	}
	return must(newList("builtin", out, synonyms))
}

// vivaceIndex is the index of the term Vivace in the list of the English
// Wikipedia: the 19th term, at place 17.
const vivaceIndex = 18

var (
	englishList = sync.OnceValue(func() List {
		return builtinList(englishRows, synonym{alias: "Vivo", index: vivaceIndex})
	})
	russianList    = sync.OnceValue(func() List { return builtinList(russianRows) })
	japaneseList   = sync.OnceValue(func() List { return builtinList(japaneseRows) })
	germanList     = sync.OnceValue(func() List { return builtinList(germanRows) })
	frenchList     = sync.OnceValue(func() List { return builtinList(frenchRows) })
	maelzelList    = sync.OnceValue(func() List { return builtinList(maelzelRows) })
	metronome1950  = sync.OnceValue(func() List { return builtinList(metronome1950Rows) })
	electronicList = sync.OnceValue(func() List { return builtinList(electronicRows) })
)

// Default returns the default list, the list of the English Wikipedia; it is
// equal to [EnglishWikipedia]. Operations never use a list implicitly: Default
// exists to pass the list explicitly.
func Default() List {
	return englishList()
}

// BuiltIn returns the eight built-in lists in the order EnglishWikipedia,
// RussianWikipedia, JapaneseMetronome, GermanMetronome, FrenchMetronome,
// MaelzelMetronome, Metronome1950, and ElectronicMetronome, as a new slice
// that belongs to the caller.
func BuiltIn() []List {
	return []List{
		englishList(), russianList(), japaneseList(), germanList(),
		frenchList(), maelzelList(), metronome1950(), electronicList(),
	}
}

// EnglishWikipedia returns the list of the article "Tempo" of the English
// Wikipedia, section "Approximately from the slowest to the fastest": 20
// places with 23 terms from Larghissimo to Prestissimo, with English meanings
// from the article and Russian meanings from the textbook "Элементарная
// теория музыки" by B. Alekseev and A. Myasoedov where it gives them. "Vivo" is
// parsed as Vivace.
func EnglishWikipedia() List {
	return englishList()
}

// RussianWikipedia returns the list of the article "Темп (музыка)" of the
// Russian Wikipedia, table "Темповые обозначения": 37 places with 39 terms in
// lower case from grave to prestissimo, in the order of the table, with
// Russian meanings from the article. vivo and vivace are different terms.
func RussianWikipedia() List {
	return russianList()
}

// JapaneseMetronome returns the list of a Japanese metronome of the 21st
// century from the article "Tempo (Musik)" of the German Wikipedia: 17 places
// from Grave to Prestissimo, without meanings.
func JapaneseMetronome() List {
	return japaneseList()
}

// GermanMetronome returns the list of a German metronome of the 20th century
// from the article "Tempo (Musik)" of the German Wikipedia: 8 places from
// Largo to Prestissimo, without meanings.
func GermanMetronome() List {
	return germanList()
}

// FrenchMetronome returns the list of a French metronome of the 19th century
// from the article "Tempo (Musik)" of the German Wikipedia: 6 places from
// Largo to Presto, without meanings.
func FrenchMetronome() List {
	return frenchList()
}

// MaelzelMetronome returns the list of a 19th-century Mälzel metronome from
// the table of tempo markings by Dolmetsch: 10 places in lower case from largo
// to prestissimo, in the order of the source, without meanings.
func MaelzelMetronome() List {
	return maelzelList()
}

// Metronome1950 returns the list of a metronome of 1950 from the table of
// tempo markings by Dolmetsch: 12 places in lower case from grave to
// prestissimo, in the order of the source, without meanings.
func Metronome1950() List {
	return metronome1950()
}

// ElectronicMetronome returns the list of a modern electronic metronome from
// the table of tempo markings by Dolmetsch: 8 places in lower case from largo
// to prestissimo, without meanings.
func ElectronicMetronome() List {
	return electronicList()
}

// englishRows are the rows of the list of the English Wikipedia.
var englishRows = []builtinRow{
	{Slow, 0, 24, []Entry{{"Larghissimo", "extremely slow, slowest type of tempo", ""}}},
	{Slow, 24, 40, []Entry{{"Adagissimo", "very slow and solemn", ""}, {"Grave", "very slow and solemn", "тяжело"}}},
	{Slow, 40, 66, []Entry{{"Largo", "slow and broad", "широко"}}},
	{Slow, 44, 66, []Entry{{"Larghetto", "rather slow and broad", "широко, но не очень"}}},
	{Slow, 40, 60, []Entry{{"Lento", "slow", "протяжно"}}},
	{Slow, 44, 66, []Entry{{"Adagio", "slow with great expression", "медленно"}}},
	{Slow, 46, 80, []Entry{{"Adagietto", "slower than andante or slightly faster than adagio", "не очень медленно"}}},
	{Moderate, 56, 108, []Entry{{"Andante", "at a walking pace, moderately slow", "не спеша"}}},
	{Moderate, 80, 108, []Entry{{"Andantino", "slightly faster than andante, but slower than moderato (although, in some cases, it can be taken to mean slightly slower than andante)", "подвижнее, чем Andante"}}},
	{Moderate, 66, 80, []Entry{{"Marcia moderato", "moderately, in the manner of a march", ""}}},
	{Moderate, 80, 108, []Entry{{"Andante moderato", "between andante and moderato (at a moderate walking speed)", ""}}},
	{Moderate, 108, 120, []Entry{{"Moderato", "at a moderate speed", "умеренно"}}},
	{Fast, 112, 120, []Entry{{"Allegretto", "by the mid-19th century, moderately fast", "не очень скоро, оживленно"}}},
	{Fast, 116, 120, []Entry{{"Allegro moderato", "close to, but not quite allegro", "умеренно скоро"}}},
	{Fast, 120, 156, []Entry{{"Allegro", "fast and bright", "скоро"}}},
	{Fast, 124, 156, []Entry{{"Molto Allegro", "at least slightly faster and livelier than allegro, but always at its range (and no faster than vivace)", ""}, {"Allegro vivace", "at least slightly faster and livelier than allegro, but always at its range (and no faster than vivace)", ""}}},
	{Fast, 156, 176, []Entry{{"Vivace", "lively and fast", "живо"}}},
	{Fast, 172, 176, []Entry{{"Vivacissimo", "very fast, lively and bright", ""}, {"Allegrissimo", "very fast, lively and bright", ""}}},
	{Fast, 168, 200, []Entry{{"Presto", "very fast", "быстро"}}},
	{Fast, 200, 0, []Entry{{"Prestissimo", "extremely fast", ""}}},
}

// russianRows are the rows of the list of the Russian Wikipedia.
var russianRows = []builtinRow{
	{Slow, 40, 48, []Entry{{"grave", "", "очень медленно, значительно, торжественно, тяжело"}}},
	{Slow, 44, 52, []Entry{{"largo", "", "широко, очень медленно"}}},
	{Slow, 46, 54, []Entry{{"largamente", "", "протяжно"}}},
	{Slow, 48, 56, []Entry{{"adagio", "", "медленно, спокойно"}}},
	{Slow, 50, 58, []Entry{{"lento", "", "медленно, слабо, тихо, скорее, чем largo"}}},
	{Slow, 52, 60, []Entry{{"lentamente", "", "медленно, слабо, тихо, скорее, чем lento"}}},
	{Slow, 54, 63, []Entry{{"larghetto", "", "довольно широко"}}},
	{Moderate, 56, 66, []Entry{{"andante assai", "", "очень спокойным шагом"}}},
	{Slow, 58, 72, []Entry{{"adagietto", "", "довольно медленно, но подвижнее, чем adagio"}}},
	{Moderate, 58, 72, []Entry{{"andante", "", "умеренный темп, в характере шага (букв. «идя»)"}}},
	{Moderate, 60, 69, []Entry{{"andante maestoso", "", "торжественным шагом"}}},
	{Moderate, 63, 76, []Entry{{"andante mosso", "", "оживлённым шагом"}}},
	{NoGroup, 63, 80, []Entry{{"comodo", "", "удобно, непринуждённо, не спеша"}, {"comodamente", "", "удобно, непринуждённо, не спеша"}}},
	{Moderate, 66, 80, []Entry{{"andante non troppo", "", "небыстрым шагом"}}},
	{Moderate, 69, 84, []Entry{{"andante con moto", "", "непринуждённо, не спеша"}}},
	{Moderate, 72, 88, []Entry{{"andantino", "", "скорее, чем andante, но медленнее, чем allegretto"}}},
	{Moderate, 76, 92, []Entry{{"moderato assai", "", "очень умеренно"}}},
	{Moderate, 80, 96, []Entry{{"moderato", "", "умеренно, сдержанно, средний темп между andante и allegro"}}},
	{NoGroup, 84, 100, []Entry{{"con moto", "", "с движением"}}},
	{Fast, 88, 104, []Entry{{"allegretto moderato", "", "умеренно оживлённо"}}},
	{Fast, 92, 108, []Entry{{"allegretto", "", "медленнее, чем allegro, но скорее, чем andante"}}},
	{Fast, 96, 112, []Entry{{"allegretto mosso", "", "быстрее, чем allegretto"}}},
	{NoGroup, 100, 116, []Entry{{"animato", "", "оживлённо"}}},
	{NoGroup, 104, 120, []Entry{{"animato assai", "", "очень оживлённо"}}},
	{Fast, 108, 126, []Entry{{"allegro moderato", "", "умеренно быстро"}}},
	{NoGroup, 112, 126, []Entry{{"tempo di marcia", "", "в темпе марша"}}},
	{Fast, 116, 132, []Entry{{"allegro ma non troppo", "", "скоро, но не слишком"}}},
	{Fast, 116, 132, []Entry{{"allegro tranquillo", "", "скоро, но спокойно"}}},
	{Fast, 120, 144, []Entry{{"allegro", "", "скорый темп (буквально: «весело»)"}}},
	{Fast, 138, 160, []Entry{{"allegro molto", "", "весьма скоро"}}},
	{Fast, 144, 168, []Entry{{"allegro assai", "", "весьма скоро"}}},
	{Fast, 152, 176, []Entry{{"allegro agitato", "", "весьма скоро, взволнованно"}, {"allegro animato", "", "весьма скоро, взволнованно"}}},
	{Fast, 160, 170, []Entry{{"allegro vivace", "", "значительно скоро, почти vivo"}}},
	{Fast, 160, 176, []Entry{{"vivo", "", "живо"}}},
	{Fast, 176, 192, []Entry{{"vivace", "", "очень живо"}}},
	{Fast, 184, 200, []Entry{{"presto", "", "быстро"}}},
	{Fast, 192, 0, []Entry{{"prestissimo", "", "очень быстро (предельно)"}}},
}

// japaneseRows are the rows of the scale "21. Jahrhundert, japanisch".
var japaneseRows = []builtinRow{
	{Slow, 40, 44, []Entry{{Name: "Grave"}}},
	{Slow, 44, 48, []Entry{{Name: "Largo"}}},
	{Slow, 48, 54, []Entry{{Name: "Lento"}}},
	{Slow, 54, 58, []Entry{{Name: "Adagio"}}},
	{Slow, 58, 63, []Entry{{Name: "Larghetto"}}},
	{Slow, 63, 69, []Entry{{Name: "Adagietto"}}},
	{Moderate, 69, 76, []Entry{{Name: "Andante"}}},
	{Moderate, 76, 84, []Entry{{Name: "Andantino"}}},
	{NoGroup, 84, 92, []Entry{{Name: "Maestoso"}}},
	{Moderate, 92, 104, []Entry{{Name: "Moderato"}}},
	{Fast, 104, 116, []Entry{{Name: "Allegretto"}}},
	{NoGroup, 116, 126, []Entry{{Name: "Animato"}}},
	{Fast, 126, 138, []Entry{{Name: "Allegro"}}},
	{NoGroup, 138, 152, []Entry{{Name: "Assai"}}},
	{Fast, 152, 176, []Entry{{Name: "Vivace"}}},
	{Fast, 176, 200, []Entry{{Name: "Presto"}}},
	{Fast, 208, 0, []Entry{{Name: "Prestissimo"}}},
}

// germanRows are the rows of the scale "20. Jahrhundert, deutsch".
var germanRows = []builtinRow{
	{Slow, 40, 60, []Entry{{Name: "Largo"}}},
	{Slow, 60, 66, []Entry{{Name: "Larghetto"}}},
	{Slow, 66, 76, []Entry{{Name: "Adagio"}}},
	{Moderate, 76, 108, []Entry{{Name: "Andante"}}},
	{Moderate, 108, 120, []Entry{{Name: "Moderato"}}},
	{Fast, 120, 168, []Entry{{Name: "Allegro"}}},
	{Fast, 168, 200, []Entry{{Name: "Presto"}}},
	{Fast, 200, 208, []Entry{{Name: "Prestissimo"}}},
}

// frenchRows are the rows of the scale "19. Jahrhundert, französisch".
var frenchRows = []builtinRow{
	{Slow, 42, 70, []Entry{{Name: "Largo"}}},
	{Slow, 70, 98, []Entry{{Name: "Larghetto"}}},
	{Slow, 98, 124, []Entry{{Name: "Adagio"}}},
	{Moderate, 124, 152, []Entry{{Name: "Andante"}}},
	{Fast, 152, 180, []Entry{{Name: "Allegro"}}},
	{Fast, 180, 208, []Entry{{Name: "Presto"}}},
}

// maelzelRows are the rows of the scale "a nineteenth-century Mälzel (or Maelzel) metronome".
var maelzelRows = []builtinRow{
	{Slow, 40, 40, []Entry{{Name: "largo"}}},
	{Slow, 60, 60, []Entry{{Name: "adagio"}}},
	{Slow, 52, 52, []Entry{{Name: "lento"}}},
	{Moderate, 69, 69, []Entry{{Name: "andante"}}},
	{Moderate, 84, 84, []Entry{{Name: "moderato"}}},
	{Fast, 100, 100, []Entry{{Name: "allegretto"}}},
	{Fast, 144, 144, []Entry{{Name: "vivace"}}},
	{Fast, 120, 120, []Entry{{Name: "allegro"}}},
	{Fast, 160, 160, []Entry{{Name: "presto"}}},
	{Fast, 184, 240, []Entry{{Name: "prestissimo"}}},
}

// metronome1950Rows are the rows of the scale "a 1950 metronome".
var metronome1950Rows = []builtinRow{
	{Slow, 44, 44, []Entry{{Name: "grave"}}},
	{Slow, 46, 46, []Entry{{Name: "largo"}}},
	{Slow, 50, 50, []Entry{{Name: "larghetto"}}},
	{Slow, 54, 54, []Entry{{Name: "adagio"}}},
	{Slow, 52, 52, []Entry{{Name: "lento"}}},
	{Moderate, 66, 66, []Entry{{Name: "andantino"}}},
	{Moderate, 80, 80, []Entry{{Name: "moderato"}}},
	{Fast, 100, 100, []Entry{{Name: "allegretto"}}},
	{Fast, 126, 126, []Entry{{Name: "vivace"}}},
	{Fast, 116, 116, []Entry{{Name: "allegro"}}},
	{Fast, 144, 144, []Entry{{Name: "presto"}}},
	{Fast, 184, 184, []Entry{{Name: "prestissimo"}}},
}

// electronicRows are the rows of the scale "a modern electronic metronome".
var electronicRows = []builtinRow{
	{Slow, 50, 50, []Entry{{Name: "largo"}}},
	{Slow, 60, 60, []Entry{{Name: "larghetto"}}},
	{Slow, 70, 70, []Entry{{Name: "adagio"}}},
	{Moderate, 80, 100, []Entry{{Name: "andante"}}},
	{Moderate, 110, 110, []Entry{{Name: "moderato"}}},
	{Fast, 120, 160, []Entry{{Name: "allegro"}}},
	{Fast, 180, 180, []Entry{{Name: "presto"}}},
	{Fast, 200, 200, []Entry{{Name: "prestissimo"}}},
}
