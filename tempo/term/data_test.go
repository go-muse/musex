package term_test

import (
	"strings"

	"github.com/go-muse/musex/tempo/term"
)

// The expected contents of the eight built-in lists: business-spec 2.4 (names,
// places, groups, guides) and the data files of the domain base
// docs/domain-knowledge/data/tempo-markings.csv and
// tempo-metronome-labels.csv (meanings, verbatim), architecture 4.5. The
// tests do not read the data files: they are outside the repository.

// wantPlace is one expected place of a built-in list. An open bound of the
// guide is 0. english is the same for all words of the place; russian has one
// entry per word, "" meaning no meaning.
type wantPlace struct {
	names   []string
	group   term.Group
	lo, hi  int
	english string
	russian []string
}

// wantList is one expected built-in list.
type wantList struct {
	name   string
	get    func() term.List
	places []wantPlace
	words  int
}

// p1 is a place of list 1 (English Wikipedia).
func p1(group term.Group, lo, hi int, english string, words ...string) wantPlace {
	n := len(words) / 2
	p := wantPlace{group: group, lo: lo, hi: hi, english: english}

	for i := range n {
		p.names = append(p.names, words[2*i])
		p.russian = append(p.russian, words[2*i+1])
	}

	return p
}

// p2 is a place of list 2 (Russian Wikipedia): one Russian meaning shared by
// all words of the place.
func p2(group term.Group, lo, hi int, russian string, words ...string) wantPlace {
	p := wantPlace{names: words, group: group, lo: lo, hi: hi}
	for range words {
		p.russian = append(p.russian, russian)
	}

	return p
}

// metroGroups are the groups of the words of lists 3-8 (business-spec 2.4,
// the item under the table of lists 3-8), by lower-case name.
var metroGroups = map[string]term.Group{
	"grave": term.Slow, "largo": term.Slow, "larghetto": term.Slow, "lento": term.Slow,
	"adagio": term.Slow, "adagietto": term.Slow,
	"andante": term.Moderate, "andantino": term.Moderate, "moderato": term.Moderate,
	"allegretto": term.Fast, "allegro": term.Fast, "vivace": term.Fast, "presto": term.Fast,
	"prestissimo": term.Fast,
	"maestoso":    term.NoGroup, "animato": term.NoGroup, "assai": term.NoGroup,
}

// pm is a place of lists 3-8: one word, no meanings.
func pm(name string, lo, hi int) wantPlace {
	g, ok := metroGroups[strings.ToLower(name)]
	if !ok {
		panic("pm: no group for " + name)
	}

	return wantPlace{names: []string{name}, group: g, lo: lo, hi: hi, russian: []string{""}}
}

const (
	gSlow     = term.Slow
	gModerate = term.Moderate
	gFast     = term.Fast
	gNone     = term.NoGroup
)

// list1 is the expected English Wikipedia list (perechen 1).
var list1 = []wantPlace{
	p1(gSlow, 0, 24, "extremely slow, slowest type of tempo", "Larghissimo", ""),
	p1(gSlow, 24, 40, "very slow and solemn", "Adagissimo", "", "Grave", "тяжело"),
	p1(gSlow, 40, 66, "slow and broad", "Largo", "широко"),
	p1(gSlow, 44, 66, "rather slow and broad", "Larghetto", "широко, но не очень"),
	p1(gSlow, 40, 60, "slow", "Lento", "протяжно"),
	p1(gSlow, 44, 66, "slow with great expression", "Adagio", "медленно"),
	p1(gSlow, 46, 80, "slower than andante or slightly faster than adagio", "Adagietto", "не очень медленно"),
	p1(gModerate, 56, 108, "at a walking pace, moderately slow", "Andante", "не спеша"),
	p1(gModerate, 80, 108, "slightly faster than andante, but slower than moderato (although, in some cases, "+
		"it can be taken to mean slightly slower than andante)", "Andantino", "подвижнее, чем Andante"),
	p1(gModerate, 66, 80, "moderately, in the manner of a march", "Marcia moderato", ""),
	p1(gModerate, 80, 108, "between andante and moderato (at a moderate walking speed)", "Andante moderato", ""),
	p1(gModerate, 108, 120, "at a moderate speed", "Moderato", "умеренно"),
	p1(gFast, 112, 120, "by the mid-19th century, moderately fast", "Allegretto", "не очень скоро, оживленно"),
	p1(gFast, 116, 120, "close to, but not quite allegro", "Allegro moderato", "умеренно скоро"),
	p1(gFast, 120, 156, "fast and bright", "Allegro", "скоро"),
	p1(gFast, 124, 156, "at least slightly faster and livelier than allegro, but always at its range "+
		"(and no faster than vivace)", "Molto Allegro", "", "Allegro vivace", ""),
	p1(gFast, 156, 176, "lively and fast", "Vivace", "живо"),
	p1(gFast, 172, 176, "very fast, lively and bright", "Vivacissimo", "", "Allegrissimo", ""),
	p1(gFast, 168, 200, "very fast", "Presto", "быстро"),
	p1(gFast, 200, 0, "extremely fast", "Prestissimo", ""),
}

// list2 is the expected Russian Wikipedia list (perechen 2).
var list2 = []wantPlace{
	p2(gSlow, 40, 48, "очень медленно, значительно, торжественно, тяжело", "grave"),
	p2(gSlow, 44, 52, "широко, очень медленно", "largo"),
	p2(gSlow, 46, 54, "протяжно", "largamente"),
	p2(gSlow, 48, 56, "медленно, спокойно", "adagio"),
	p2(gSlow, 50, 58, "медленно, слабо, тихо, скорее, чем largo", "lento"),
	p2(gSlow, 52, 60, "медленно, слабо, тихо, скорее, чем lento", "lentamente"),
	p2(gSlow, 54, 63, "довольно широко", "larghetto"),
	p2(gModerate, 56, 66, "очень спокойным шагом", "andante assai"),
	p2(gSlow, 58, 72, "довольно медленно, но подвижнее, чем adagio", "adagietto"),
	p2(gModerate, 58, 72, "умеренный темп, в характере шага (букв. «идя»)", "andante"),
	p2(gModerate, 60, 69, "торжественным шагом", "andante maestoso"),
	p2(gModerate, 63, 76, "оживлённым шагом", "andante mosso"),
	p2(gNone, 63, 80, "удобно, непринуждённо, не спеша", "comodo", "comodamente"),
	p2(gModerate, 66, 80, "небыстрым шагом", "andante non troppo"),
	p2(gModerate, 69, 84, "непринуждённо, не спеша", "andante con moto"),
	p2(gModerate, 72, 88, "скорее, чем andante, но медленнее, чем allegretto", "andantino"),
	p2(gModerate, 76, 92, "очень умеренно", "moderato assai"),
	p2(gModerate, 80, 96, "умеренно, сдержанно, средний темп между andante и allegro", "moderato"),
	p2(gNone, 84, 100, "с движением", "con moto"),
	p2(gFast, 88, 104, "умеренно оживлённо", "allegretto moderato"),
	p2(gFast, 92, 108, "медленнее, чем allegro, но скорее, чем andante", "allegretto"),
	p2(gFast, 96, 112, "быстрее, чем allegretto", "allegretto mosso"),
	p2(gNone, 100, 116, "оживлённо", "animato"),
	p2(gNone, 104, 120, "очень оживлённо", "animato assai"),
	p2(gFast, 108, 126, "умеренно быстро", "allegro moderato"),
	p2(gNone, 112, 126, "в темпе марша", "tempo di marcia"),
	p2(gFast, 116, 132, "скоро, но не слишком", "allegro ma non troppo"),
	p2(gFast, 116, 132, "скоро, но спокойно", "allegro tranquillo"),
	p2(gFast, 120, 144, "скорый темп (буквально: «весело»)", "allegro"),
	p2(gFast, 138, 160, "весьма скоро", "allegro molto"),
	p2(gFast, 144, 168, "весьма скоро", "allegro assai"),
	p2(gFast, 152, 176, "весьма скоро, взволнованно", "allegro agitato", "allegro animato"),
	p2(gFast, 160, 170, "значительно скоро, почти vivo", "allegro vivace"),
	p2(gFast, 160, 176, "живо", "vivo"),
	p2(gFast, 176, 192, "очень живо", "vivace"),
	p2(gFast, 184, 200, "быстро", "presto"),
	p2(gFast, 192, 0, "очень быстро (предельно)", "prestissimo"),
}

// list3 to list8 are the expected metronome lists.
var (
	list3 = []wantPlace{
		pm("Grave", 40, 44), pm("Largo", 44, 48), pm("Lento", 48, 54), pm("Adagio", 54, 58),
		pm("Larghetto", 58, 63), pm("Adagietto", 63, 69), pm("Andante", 69, 76), pm("Andantino", 76, 84),
		pm("Maestoso", 84, 92), pm("Moderato", 92, 104), pm("Allegretto", 104, 116), pm("Animato", 116, 126),
		pm("Allegro", 126, 138), pm("Assai", 138, 152), pm("Vivace", 152, 176), pm("Presto", 176, 200),
		pm("Prestissimo", 208, 0),
	}
	list4 = []wantPlace{
		pm("Largo", 40, 60), pm("Larghetto", 60, 66), pm("Adagio", 66, 76), pm("Andante", 76, 108),
		pm("Moderato", 108, 120), pm("Allegro", 120, 168), pm("Presto", 168, 200), pm("Prestissimo", 200, 208),
	}
	list5 = []wantPlace{
		pm("Largo", 42, 70), pm("Larghetto", 70, 98), pm("Adagio", 98, 124), pm("Andante", 124, 152),
		pm("Allegro", 152, 180), pm("Presto", 180, 208),
	}
	list6 = []wantPlace{
		pm("largo", 40, 40), pm("adagio", 60, 60), pm("lento", 52, 52), pm("andante", 69, 69),
		pm("moderato", 84, 84), pm("allegretto", 100, 100), pm("vivace", 144, 144), pm("allegro", 120, 120),
		pm("presto", 160, 160), pm("prestissimo", 184, 240),
	}
	list7 = []wantPlace{
		pm("grave", 44, 44), pm("largo", 46, 46), pm("larghetto", 50, 50), pm("adagio", 54, 54),
		pm("lento", 52, 52), pm("andantino", 66, 66), pm("moderato", 80, 80), pm("allegretto", 100, 100),
		pm("vivace", 126, 126), pm("allegro", 116, 116), pm("presto", 144, 144), pm("prestissimo", 184, 184),
	}
	list8 = []wantPlace{
		pm("largo", 50, 50), pm("larghetto", 60, 60), pm("adagio", 70, 70), pm("andante", 80, 100),
		pm("moderato", 110, 110), pm("allegro", 120, 160), pm("presto", 180, 180), pm("prestissimo", 200, 200),
	}
)

// builtIns are the eight built-in lists in the order 1-8 of architecture 4.5.
func builtIns() []wantList {
	return []wantList{
		{"EnglishWikipedia", term.EnglishWikipedia, list1, 23},
		{"RussianWikipedia", term.RussianWikipedia, list2, 39},
		{"JapaneseMetronome", term.JapaneseMetronome, list3, 17},
		{"GermanMetronome", term.GermanMetronome, list4, 8},
		{"FrenchMetronome", term.FrenchMetronome, list5, 6},
		{"MaelzelMetronome", term.MaelzelMetronome, list6, 10},
		{"Metronome1950", term.Metronome1950, list7, 12},
		{"ElectronicMetronome", term.ElectronicMetronome, list8, 8},
	}
}

// rowsOf returns the rows of a user list with the same contents as an
// expected built-in list.
func rowsOf(places []wantPlace) []term.Row {
	rows := make([]term.Row, 0, len(places))

	for _, p := range places {
		row := term.Row{Group: p.group, Guide: guidePtr(guideOf(p.lo, p.hi))}
		for i, n := range p.names {
			row.Entries = append(row.Entries, term.Entry{Name: n, English: p.english, Russian: p.russian[i]})
		}

		rows = append(rows, row)
	}

	return rows
}

// stringOf returns the debug form of architecture 4.4 for the expected places.
func stringOf(places []wantPlace) string {
	parts := make([]string, 0, len(places))
	for _, p := range places {
		parts = append(parts, strings.Join(p.names, ", "))
	}

	return "List[" + strings.Join(parts, "; ") + "]"
}
