package search_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-muse/musex/search"
	"github.com/go-muse/musex/settings"
	"github.com/go-muse/musex/system"
	"github.com/go-muse/musex/template"
)

func exampleSearcher(codes ...string) search.Searcher {
	catalog := template.BuiltIn()
	entries := make([]template.Entry, len(codes))
	for i, code := range codes {
		entry, ok := catalog.ByCode(code)
		if !ok {
			panic("unknown built-in entry: " + code)
		}
		entries[i] = entry
	}
	selected := template.MustNewCatalog(entries...)
	return search.MustNew(&selected)
}

func printSpellingMatches(matches []search.SpellingMatch) {
	for _, match := range matches {
		origin, _ := match.Mode().Origin()
		fmt.Printf("%s / %s\n", origin.Code(), spellingText(match.Tonic()))
	}
}

func spellingGroupText(group []search.SpellingMatch) string {
	parts := make([]string, len(group))
	for i, match := range group {
		origin, _ := match.Mode().Origin()
		parts[i] = origin.Code() + " / " + spellingText(match.Tonic())
	}
	return strings.Join(parts, "; ")
}

func ExampleSearcher_BySpelling() {
	searcher := exampleSearcher("ionian")
	matches := searcher.MustBySpelling(notes("D", "E", "F♯", "A", "B"), search.Containment)
	printSpellingMatches(matches)
	// Output:
	// ionian / D
	// ionian / G
	// ionian / A
}

func ExampleSearcher_BySpelling_fullMatch() {
	searcher := exampleSearcher("ionian", "dorian", "phrygian", "lydian", "mixolydian", "aeolian", "locrian")
	matches := searcher.MustBySpelling(notes("C", "D", "E", "F", "G", "A", "B"), search.FullMatch)
	printSpellingMatches(matches)
	// Output:
	// ionian / C
	// dorian / D
	// phrygian / E
	// lydian / F
	// mixolydian / G
	// aeolian / A
	// locrian / B
}

func ExampleSearcher_ByPosition() {
	searcher := exampleSearcher("ionian")
	matches := searcher.MustByPosition(system.Traditional(), notes("C♯", "F"), search.Containment)
	for _, match := range matches {
		origin, _ := match.Mode().Origin()
		fmt.Printf("%s / %s (class %d)\n", origin.Code(), spellingText(match.Tonic()), match.Family().Class().Number())
	}
	// Output:
	// ionian / C♯ (class 1)
	// ionian / F♯ (class 6)
	// ionian / G♯ (class 8)
}

func ExampleSearcher_ByPositionFromContext() {
	source := settings.Standard()
	input := notesIn(source, "D", "E", "F♯", "A", "B")
	matches := exampleSearcher("ionian").MustByPositionFromContext(input, search.Containment)
	for _, match := range matches {
		origin, _ := match.Mode().Origin()
		fmt.Printf("%s / %s\n", origin.Code(), spellingText(match.Tonic()))
	}
	// Output:
	// ionian / D
	// ionian / G
	// ionian / A
}

func ExampleTonicFamily_Spellings() {
	searcher := exampleSearcher("ionian")
	matches := searcher.MustByPosition(system.Traditional(), notes("C♯", "F"), search.Containment)
	for name := range matches[0].Family().MustSpellings(alt(1)) {
		fmt.Println(spellingText(name))
	}
	// Output:
	// C♯
	// D♭
}

func ExampleGroupSpellingByCollection() {
	searcher := exampleSearcher("ionian", "aeolian")
	matches := searcher.MustBySpelling(notes("D", "E", "F♯", "A", "B"), search.Containment)
	groups := search.MustGroupSpellingByCollection(matches)
	for _, group := range groups {
		fmt.Println(spellingGroupText(group))
	}
	// Output:
	// ionian / D; aeolian / B
	// ionian / G; aeolian / E
	// ionian / A; aeolian / F♯
}

func ExampleGroupSpellingByTonic() {
	searcher := exampleSearcher("hirajoshi-sachs", "iwato")
	matches := searcher.MustBySpelling(notes("C", "D♭", "F", "G♭", "B♭"), search.FullMatch)
	groups := search.MustGroupSpellingByTonic(matches)
	for _, group := range groups {
		fmt.Println(spellingGroupText(group))
	}
	// Output:
	// hirajoshi-sachs / C; iwato / C
}

func ExampleGroupSpellingByStructure() {
	searcher := exampleSearcher("ionian", "aeolian")
	matches := searcher.MustBySpelling(notes("D", "E", "F♯", "A", "B"), search.Containment)
	groups := search.MustGroupSpellingByStructure(matches)
	for _, group := range groups {
		fmt.Println(spellingGroupText(group))
	}
	// Output:
	// ionian / D; ionian / G; ionian / A
	// aeolian / E; aeolian / F♯; aeolian / B
}

func ExampleSearcher_BySpelling_emptyInput() {
	_, err := search.BuiltIn().BySpelling(nil, search.Containment)
	fmt.Println(errors.Is(err, search.ErrEmptyInput))
	// Output:
	// true
}
